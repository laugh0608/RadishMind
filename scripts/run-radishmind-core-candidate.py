#!/usr/bin/env python3
from __future__ import annotations

import argparse
import copy
import json
import os
import sys
from pathlib import Path
from typing import Any

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent
COPILOT_REQUEST_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-request.schema.json"
COPILOT_RESPONSE_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-response.schema.json"
DEFAULT_MANIFEST_PATH = REPO_ROOT / "scripts/checks/fixtures/radishmind-core-candidate-dry-run-manifest.json"
DEFAULT_OUTPUT_DIR = REPO_ROOT / "tmp/radishmind-core-candidate-dry-run"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--manifest", default=str(DEFAULT_MANIFEST_PATH.relative_to(REPO_ROOT)))
    parser.add_argument("--output-dir", default=str(DEFAULT_OUTPUT_DIR.relative_to(REPO_ROOT)))
    parser.add_argument("--summary-output")
    parser.add_argument("--check-summary")
    parser.add_argument("--provider", choices=["golden_fixture", "echo", "local_transformers"])
    parser.add_argument("--model-dir")
    parser.add_argument("--max-new-tokens", type=int, default=1200)
    parser.add_argument("--sample-id")
    parser.add_argument(
        "--allow-invalid-output",
        action="store_true",
        help="Write invalid local model outputs for inspection instead of failing the whole run.",
    )
    return parser.parse_args()


def load_json(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json document '{repo_rel(path)}': {exc}") from exc


def write_json(path: Path, document: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(document, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def normalize_repo_path(path_value: str | None) -> Path | None:
    if not path_value:
        return None
    path = Path(path_value)
    return path if path.is_absolute() else REPO_ROOT / path


def repo_rel(path: Path) -> str:
    try:
        return path.relative_to(REPO_ROOT).as_posix()
    except ValueError:
        return path.as_posix()


def load_schema(path: Path) -> dict[str, Any]:
    schema = load_json(path)
    require(isinstance(schema, dict), f"{repo_rel(path)} must be a JSON object")
    jsonschema.Draft202012Validator.check_schema(schema)
    return schema


def validate_manifest(manifest: dict[str, Any]) -> None:
    require(manifest.get("schema_version") == 1, "candidate dry-run manifest schema_version must be 1")
    require(
        manifest.get("kind") == "radishmind_core_candidate_dry_run_manifest",
        "candidate dry-run manifest kind mismatch",
    )
    require(manifest.get("phase") == "M4-preparation", "candidate dry-run manifest phase mismatch")
    source_eval_manifest = str(manifest.get("source_eval_manifest") or "").strip()
    require(source_eval_manifest, "candidate dry-run manifest must include source_eval_manifest")

    provider = manifest.get("provider")
    require(isinstance(provider, dict), "candidate dry-run manifest provider must be an object")
    require(provider.get("provider_id") in {"golden_fixture", "echo"}, "committed candidate dry-run provider must be deterministic")
    require(provider.get("does_not_run_models") is True, "committed candidate dry-run must not run models")
    require(provider.get("provider_access") == "none", "committed candidate dry-run must not access providers")
    require(provider.get("model_artifacts_downloaded") is False, "committed candidate dry-run must not download model artifacts")

    output_policy = manifest.get("output_policy")
    require(isinstance(output_policy, dict), "candidate dry-run manifest output_policy must be an object")
    require(output_policy.get("commit_candidate_outputs") is False, "candidate outputs must remain local generated artifacts")


def validate_source_eval_manifest(source_eval_manifest: dict[str, Any]) -> None:
    require(source_eval_manifest.get("schema_version") == 1, "source eval manifest schema_version must be 1")
    require(
        source_eval_manifest.get("kind") == "radishmind_core_offline_eval_fixture_run_manifest",
        "source eval manifest kind mismatch",
    )
    require(source_eval_manifest.get("phase") == "M4-preparation", "source eval manifest phase mismatch")
    sample_selection = source_eval_manifest.get("sample_selection")
    require(isinstance(sample_selection, list) and sample_selection, "source eval manifest sample_selection must be non-empty")


def build_prompt_document(sample: dict[str, Any], *, model_id: str) -> dict[str, Any]:
    input_request = sample["input_request"]
    project = str(sample["project"])
    task = str(sample["task"])
    return {
        "schema_version": 1,
        "kind": "radishmind_core_candidate_prompt",
        "sample_id": sample["sample_id"],
        "project": project,
        "task": task,
        "model_id": model_id,
        "messages": [
            {
                "role": "system",
                "content": (
                    "你是 RadishMind-Core 候选模型。只输出一个符合 contracts/copilot-response.schema.json "
                    "的 JSON 对象；保持 advisory-only，不得声称已经写回业务真相源。"
                ),
            },
            {
                "role": "user",
                "content": (
                    f"请基于这个 {project}/{task} CopilotRequest 生成 CopilotResponse。"
                    "高风险或需要修改业务状态的候选动作必须保留 requires_confirmation=true。"
                ),
            },
            {
                "role": "user",
                "content_json": input_request,
            },
        ],
        "output_contract": "contracts/copilot-response.schema.json",
        "safety": {
            "advisory_only": True,
            "must_preserve_requires_confirmation": True,
            "must_not_download_models": True,
        },
    }


def build_echo_response(sample: dict[str, Any]) -> dict[str, Any]:
    request = sample["input_request"]
    request_id = str(request.get("request_id") or sample["sample_id"])
    summary = f"Dry-run echo provider received {sample['project']}/{sample['task']} request {request_id}."
    return {
        "schema_version": 1,
        "status": "partial",
        "project": sample["project"],
        "task": sample["task"],
        "summary": summary,
        "answers": [
            {
                "kind": "dry_run_echo",
                "text": summary,
            }
        ],
        "issues": [],
        "proposed_actions": [],
        "citations": [],
        "confidence": 0.0,
        "risk_level": "low",
        "requires_confirmation": False,
    }


def render_prompt_text(prompt_document: dict[str, Any]) -> str:
    lines: list[str] = []
    for message in prompt_document["messages"]:
        role = message["role"]
        if "content_json" in message:
            content = json.dumps(message["content_json"], ensure_ascii=False, indent=2)
        else:
            content = str(message["content"])
        lines.append(f"{role}:\n{content}")
    lines.append("assistant:\n")
    return "\n\n".join(lines)


def extract_json_object(text: str) -> dict[str, Any]:
    start = text.find("{")
    require(start >= 0, "local_transformers output did not contain a JSON object")
    depth = 0
    in_string = False
    escaped = False
    for index in range(start, len(text)):
        char = text[index]
        if in_string:
            if escaped:
                escaped = False
            elif char == "\\":
                escaped = True
            elif char == '"':
                in_string = False
            continue
        if char == '"':
            in_string = True
        elif char == "{":
            depth += 1
        elif char == "}":
            depth -= 1
            if depth == 0:
                document = json.loads(text[start : index + 1])
                require(isinstance(document, dict), "local_transformers output JSON must be an object")
                return document
    raise SystemExit("local_transformers output JSON object was not closed")


def run_local_transformers(prompt_document: dict[str, Any], *, model_dir: str | None, max_new_tokens: int) -> dict[str, Any]:
    resolved_model_dir = model_dir or os.environ.get("RADISHMIND_MODEL_DIR")
    require(
        bool(resolved_model_dir),
        "local_transformers provider requires --model-dir or RADISHMIND_MODEL_DIR; this script never downloads models",
    )
    model_path = Path(str(resolved_model_dir)).expanduser()
    require(model_path.is_dir(), f"local model directory does not exist: {model_path}")
    try:
        import torch
        from transformers import AutoModelForCausalLM, AutoTokenizer
    except Exception as exc:
        raise SystemExit(
            "local_transformers provider requires locally installed torch and transformers; "
            "dependency installation is not performed by this script"
        ) from exc

    tokenizer = AutoTokenizer.from_pretrained(model_path, local_files_only=True)
    model = AutoModelForCausalLM.from_pretrained(model_path, local_files_only=True)
    device = "cuda" if torch.cuda.is_available() else "cpu"
    model.to(device)
    prompt_text = render_prompt_text(prompt_document)
    inputs = tokenizer(prompt_text, return_tensors="pt").to(device)
    output_ids = model.generate(
        **inputs,
        max_new_tokens=max_new_tokens,
        do_sample=False,
        pad_token_id=tokenizer.eos_token_id,
    )
    generated_ids = output_ids[0][inputs["input_ids"].shape[-1] :]
    generated_text = tokenizer.decode(generated_ids, skip_special_tokens=True)
    return extract_json_object(generated_text)


def build_candidate_response(
    sample: dict[str, Any],
    *,
    provider_id: str,
    prompt_document: dict[str, Any],
    model_dir: str | None,
    max_new_tokens: int,
) -> tuple[dict[str, Any], str]:
    if provider_id == "golden_fixture":
        response = sample.get("golden_response")
        require(isinstance(response, dict), f"{sample['sample_id']} is missing golden_response")
        return copy.deepcopy(response), "golden_response"
    if provider_id == "echo":
        return build_echo_response(sample), "echo"
    if provider_id == "local_transformers":
        return run_local_transformers(prompt_document, model_dir=model_dir, max_new_tokens=max_new_tokens), "local_transformers"
    raise SystemExit(f"unsupported provider: {provider_id}")


def validate_candidate_response(
    response: dict[str, Any],
    *,
    sample: dict[str, Any],
    response_schema: dict[str, Any],
) -> None:
    jsonschema.validate(response, response_schema)
    require(response.get("project") == sample.get("project"), f"{sample['sample_id']} response project mismatch")
    require(response.get("task") == sample.get("task"), f"{sample['sample_id']} response task mismatch")
    for action_index, action in enumerate(response.get("proposed_actions") or []):
        if not isinstance(action, dict):
            continue
        risk_level = action.get("risk_level")
        if risk_level == "high":
            require(
                action.get("requires_confirmation") is True,
                f"{sample['sample_id']} proposed_actions[{action_index}] high risk action must require confirmation",
            )


def iter_selected_samples(
    source_eval_manifest: dict[str, Any],
    *,
    sample_id_filter: str | None,
) -> list[tuple[dict[str, Any], dict[str, Any]]]:
    selected: list[tuple[dict[str, Any], dict[str, Any]]] = []
    for group in source_eval_manifest["sample_selection"]:
        project = str(group["project"])
        task = str(group["task"])
        for selected_sample in group["selected_samples"]:
            sample_path_value = str(selected_sample.get("path") or "").strip()
            sample_id = str(selected_sample.get("sample_id") or "").strip()
            if sample_id_filter and sample_id != sample_id_filter:
                continue
            require(sample_path_value and sample_id, f"{project}/{task} selected sample must include sample_id and path")
            sample_path = normalize_repo_path(sample_path_value)
            require(sample_path is not None and sample_path.is_file(), f"candidate sample path is missing: {sample_path_value}")
            sample = load_json(sample_path)
            require(isinstance(sample, dict), f"{sample_path_value} must be a JSON object")
            require(sample.get("sample_id") == sample_id, f"{sample_path_value} sample_id mismatch")
            require(sample.get("project") == project, f"{sample_path_value} project mismatch")
            require(sample.get("task") == task, f"{sample_path_value} task mismatch")
            selected.append((sample, selected_sample))
    if sample_id_filter:
        require(selected, f"sample-id was not selected by source eval manifest: {sample_id_filter}")
    return selected


def summarize_validation_error(exc: Exception) -> str:
    if isinstance(exc, jsonschema.ValidationError):
        location = ".".join(str(part) for part in exc.absolute_path)
        return f"{location or '$'}: {exc.message}"
    return str(exc)


def build_candidate_run(args: argparse.Namespace, manifest: dict[str, Any]) -> dict[str, Any]:
    source_eval_manifest_path = normalize_repo_path(str(manifest["source_eval_manifest"]))
    require(source_eval_manifest_path is not None, "source_eval_manifest path is required")
    source_eval_manifest = load_json(source_eval_manifest_path)
    require(isinstance(source_eval_manifest, dict), "source eval manifest must be a JSON object")
    validate_source_eval_manifest(source_eval_manifest)

    request_schema = load_schema(COPILOT_REQUEST_SCHEMA_PATH)
    response_schema = load_schema(COPILOT_RESPONSE_SCHEMA_PATH)
    provider = dict(manifest["provider"])
    provider_id = args.provider or str(provider["provider_id"])
    if args.provider:
        provider.update(
            {
                "provider_id": provider_id,
                "does_not_run_models": provider_id != "local_transformers",
                "provider_access": "local_only" if provider_id == "local_transformers" else "none",
                "model_artifacts_downloaded": False,
            }
        )

    output_dir = normalize_repo_path(args.output_dir)
    require(output_dir is not None, "output-dir is required")
    prompt_dir = output_dir / "prompts"
    response_dir = output_dir / "responses"

    task_counts: dict[str, int] = {}
    outputs: list[dict[str, Any]] = []
    valid_response_count = 0
    invalid_response_count = 0
    for sample, selected_sample in iter_selected_samples(source_eval_manifest, sample_id_filter=args.sample_id):
        jsonschema.validate(sample["input_request"], request_schema)
        task_key = f"{sample['project']}/{sample['task']}"
        task_counts[task_key] = task_counts.get(task_key, 0) + 1
        prompt_document = build_prompt_document(sample, model_id=str(provider.get("model_id") or provider_id))
        sample_id = str(sample["sample_id"])
        prompt_relpath = Path("prompts") / f"{sample_id}.prompt.json"
        response_relpath = Path("responses") / f"{sample_id}.candidate-response.json"
        invalid_response_relpath = Path("invalid-responses") / f"{sample_id}.candidate-response.invalid.json"
        write_json(prompt_dir / prompt_relpath.name, prompt_document)

        candidate_response, response_source = build_candidate_response(
            sample,
            provider_id=provider_id,
            prompt_document=prompt_document,
            model_dir=args.model_dir,
            max_new_tokens=args.max_new_tokens,
        )
        response_valid = True
        validation_error: str | None = None
        try:
            validate_candidate_response(candidate_response, sample=sample, response_schema=response_schema)
        except Exception as exc:
            if not args.allow_invalid_output:
                raise
            response_valid = False
            validation_error = summarize_validation_error(exc)

        if response_valid:
            valid_response_count += 1
            write_json(response_dir / response_relpath.name, candidate_response)
        else:
            invalid_response_count += 1
            write_json(output_dir / invalid_response_relpath, candidate_response)
        outputs.append(
            {
                "sample_id": sample_id,
                "project": sample["project"],
                "task": sample["task"],
                "coverage_tags": selected_sample.get("coverage_tags") or [],
                "prompt_file": prompt_relpath.as_posix(),
                "candidate_response_file": response_relpath.as_posix() if response_valid else invalid_response_relpath.as_posix(),
                "response_source": response_source,
                "copilot_request_schema_validated": True,
                "copilot_response_schema_validated": response_valid,
                "project_task_matched": response_valid,
                **({"validation_error": validation_error} if validation_error else {}),
            }
        )

    sample_count = len(outputs)
    all_valid = invalid_response_count == 0
    return {
        "schema_version": 1,
        "kind": "radishmind_core_candidate_run_summary",
        "phase": manifest["phase"],
        "run_id": manifest["run_id"],
        "source_eval_manifest": repo_rel(source_eval_manifest_path),
        "provider": provider,
        "sample_count": sample_count,
        "task_counts": dict(sorted(task_counts.items())),
        "output_policy": {
            "output_dir": "provided at runtime",
            "prompt_files_written": sample_count,
            "candidate_response_files_written": valid_response_count,
            "invalid_candidate_response_files_written": invalid_response_count,
            "commit_candidate_outputs": False,
        },
        "quality_gates": {
            "copilot_request_schema_validated": sample_count,
            "copilot_response_schema_validated": valid_response_count,
            "project_task_matched": valid_response_count,
            "high_risk_actions_require_confirmation": all_valid,
            "does_not_download_model_artifacts": provider.get("model_artifacts_downloaded") is False,
            "does_not_start_training": True,
        },
        "outputs": outputs,
        "next_step": (
            "After an approved local model exists, rerun with --provider local_transformers and "
            "--model-dir or RADISHMIND_MODEL_DIR; this script uses local_files_only and never downloads weights."
        ),
    }


def main() -> int:
    args = parse_args()
    manifest_path = normalize_repo_path(args.manifest)
    require(manifest_path is not None and manifest_path.is_file(), f"manifest is missing: {args.manifest}")
    manifest = load_json(manifest_path)
    require(isinstance(manifest, dict), "candidate dry-run manifest must be a JSON object")
    validate_manifest(manifest)

    summary = build_candidate_run(args, manifest)
    if args.summary_output:
        summary_output = normalize_repo_path(args.summary_output)
        require(summary_output is not None, "summary-output is required")
        write_json(summary_output, summary)
    if args.check_summary:
        check_summary_path = normalize_repo_path(args.check_summary)
        require(check_summary_path is not None and check_summary_path.is_file(), f"check summary is missing: {args.check_summary}")
        expected = load_json(check_summary_path)
        if expected != summary:
            raise SystemExit(f"candidate run summary does not match {repo_rel(check_summary_path)}")

    print("radishmind core candidate wrapper passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

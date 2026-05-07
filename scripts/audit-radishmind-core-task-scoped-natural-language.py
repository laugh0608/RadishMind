#!/usr/bin/env python3
from __future__ import annotations

import argparse
import importlib.util
import json
import sys
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
DEFAULT_MANIFEST = "scripts/checks/fixtures/radishmind-core-holdout-probe-v2-candidate-eval-manifest.json"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Audit task-scoped builder natural-language quality without running models.",
    )
    parser.add_argument("--manifest", default=DEFAULT_MANIFEST)
    parser.add_argument("--candidate-summary", required=True)
    parser.add_argument("--candidate-output-dir", required=True)
    parser.add_argument("--output", default="")
    parser.add_argument("--check-summary", default="")
    return parser.parse_args()


def repo_path(path_value: str) -> Path:
    path = Path(path_value)
    return path if path.is_absolute() else REPO_ROOT / path


def repo_rel(path: Path) -> str:
    try:
        return path.relative_to(REPO_ROOT).as_posix()
    except ValueError:
        return path.as_posix()


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


def load_candidate_runner() -> Any:
    module_path = REPO_ROOT / "scripts/run-radishmind-core-candidate.py"
    spec = importlib.util.spec_from_file_location("radishmind_core_candidate_runner", module_path)
    require(spec is not None and spec.loader is not None, "candidate runner module spec must be loadable")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def sample_records(manifest: dict[str, Any]) -> dict[str, dict[str, Any]]:
    records: dict[str, dict[str, Any]] = {}
    for group in manifest.get("sample_selection") or []:
        for selected in group.get("selected_samples") or []:
            if not isinstance(selected, dict):
                continue
            sample_id = str(selected.get("sample_id") or "").strip()
            sample_path = str(selected.get("path") or "").strip()
            if sample_id and sample_path:
                sample = load_json(repo_path(sample_path))
                require(isinstance(sample, dict), f"{sample_path} must be a JSON object")
                records[sample_id] = {"selection": selected, "sample": sample}
    return records


def text_value(value: Any) -> str:
    return str(value).strip() if isinstance(value, str) else ""


def contains_any(text: str, terms: tuple[str, ...]) -> list[str]:
    return [term for term in terms if term and term in text]


def natural_language_fields(response: dict[str, Any]) -> list[tuple[str, str]]:
    fields: list[tuple[str, str]] = []
    summary = text_value(response.get("summary"))
    if summary:
        fields.append(("$.summary", summary))
    answers = response.get("answers") if isinstance(response.get("answers"), list) else []
    for index, answer in enumerate(answers):
        if isinstance(answer, dict):
            text = text_value(answer.get("text"))
            if text:
                fields.append((f"$.answers[{index}].text", text))
    issues = response.get("issues") if isinstance(response.get("issues"), list) else []
    for index, issue in enumerate(issues):
        if isinstance(issue, dict):
            message = text_value(issue.get("message"))
            if message:
                fields.append((f"$.issues[{index}].message", message))
    actions = response.get("proposed_actions") if isinstance(response.get("proposed_actions"), list) else []
    for index, action in enumerate(actions):
        if not isinstance(action, dict):
            continue
        for field_name in ("title", "rationale"):
            text = text_value(action.get(field_name))
            if text:
                fields.append((f"$.proposed_actions[{index}].{field_name}", text))
    return fields


def coverage_tags(sample: dict[str, Any], selection: dict[str, Any]) -> set[str]:
    tags = set(str(tag) for tag in selection.get("coverage_tags") or [] if str(tag).strip())
    tags.update(str(tag) for tag in sample.get("coverage_tags") or [] if str(tag).strip())
    return tags


def audit_output(
    *,
    output: dict[str, Any],
    response: dict[str, Any],
    sample: dict[str, Any],
    selection: dict[str, Any],
    generic_placeholders: tuple[str, ...],
    ghost_mistranslation_terms: tuple[str, ...],
) -> dict[str, Any]:
    sample_id = str(output.get("sample_id") or sample.get("sample_id") or "")
    project = str(output.get("project") or sample.get("project") or "")
    task = str(output.get("task") or sample.get("task") or "")
    paths = set(
        str(path)
        for path in (((output.get("postprocess") or {}).get("task_scoped_builder_paths")) or [])
        if str(path).strip()
    )
    natural_fields = natural_language_fields(response)
    merged = [path for path, _text in natural_fields if path in paths]
    fallback = [path for path, _text in natural_fields if path not in paths]
    violations: list[str] = []
    warning_notes: list[str] = []

    for path, text in natural_fields:
        placeholder_hits = contains_any(text, generic_placeholders)
        if placeholder_hits:
            violations.append(f"{path} contains generic placeholder text: {', '.join(placeholder_hits)}")
        if project == "radishflow" and task == "suggest_ghost_completion":
            mistranslation_hits = contains_any(text, ghost_mistranslation_terms)
            if mistranslation_hits:
                violations.append(f"{path} contains ghost legal/regulation mistranslation: {', '.join(mistranslation_hits)}")
        if len(text) < 8:
            warning_notes.append(f"{path} is very short")

    tags = coverage_tags(sample, selection)
    joined_text = "\n".join(text for _path, text in natural_fields)
    if task == "answer_docs_question" and "source_conflict" in tags:
        missing_terms = [term for term in ("docs", "FAQ", "forum") if term not in joined_text]
        if missing_terms:
            violations.append(f"source-conflict answer text is missing source context terms: {', '.join(missing_terms)}")

    if task == "answer_docs_question" and "evidence_gap" in tags:
        if not any(term in joined_text for term in ("证据不足", "partial", "不足")):
            warning_notes.append("evidence-gap response does not explicitly mention evidence insufficiency")

    return {
        "sample_id": sample_id,
        "project": project,
        "task": task,
        "natural_field_count": len(natural_fields),
        "merged_natural_field_count": len(merged),
        "fallback_natural_field_count": len(fallback),
        "merged_paths": merged,
        "fallback_paths": fallback,
        "violations": violations,
        "warnings": warning_notes,
        "status": "pass" if not violations else "fail",
    }


def build_summary(
    *,
    manifest_path: Path,
    manifest: dict[str, Any],
    candidate_summary: dict[str, Any],
    candidate_output_dir: Path,
) -> dict[str, Any]:
    runner = load_candidate_runner()
    postprocess_policy = candidate_summary.get("postprocess_policy")
    require(
        isinstance(postprocess_policy, dict) and postprocess_policy.get("build_task_scoped_response") is True,
        "task-scoped natural-language audit requires a --build-task-scoped-response candidate summary",
    )
    generic_placeholders = tuple(str(item) for item in runner.GENERIC_NATURAL_LANGUAGE_PLACEHOLDERS)
    ghost_mistranslation_terms = tuple(str(item) for item in runner.LEGAL_CANDIDATE_MISTRANSLATION_TERMS)
    records = sample_records(manifest)
    cases: list[dict[str, Any]] = []

    for output in candidate_summary.get("outputs") or []:
        if not isinstance(output, dict):
            continue
        sample_id = str(output.get("sample_id") or "")
        record = records.get(sample_id)
        require(record is not None, f"candidate output sample not found in manifest: {sample_id}")
        response_file = str(output.get("candidate_response_file") or "").strip()
        require(response_file, f"{sample_id} is missing candidate_response_file")
        response = load_json(candidate_output_dir / response_file)
        require(isinstance(response, dict), f"{response_file} must be a JSON object")
        cases.append(
            audit_output(
                output=output,
                response=response,
                sample=record["sample"],
                selection=record["selection"],
                generic_placeholders=generic_placeholders,
                ghost_mistranslation_terms=ghost_mistranslation_terms,
            )
        )

    require(cases, "task-scoped natural-language audit found no candidate outputs")
    violation_count = sum(len(case["violations"]) for case in cases)
    warning_count = sum(len(case["warnings"]) for case in cases)
    natural_field_count = sum(int(case["natural_field_count"]) for case in cases)
    fallback_field_count = sum(int(case["fallback_natural_field_count"]) for case in cases)
    merged_field_count = sum(int(case["merged_natural_field_count"]) for case in cases)
    return {
        "schema_version": 1,
        "kind": "radishmind_core_task_scoped_natural_language_audit",
        "phase": "M4-preparation",
        "audit_manifest": repo_rel(manifest_path),
        "source_eval_manifest": str(manifest.get("source_eval_manifest") or ""),
        "candidate_summary_kind": candidate_summary.get("kind"),
        "sample_count": len(cases),
        "audit_policy": {
            "does_not_run_models": True,
            "checks_generic_placeholders": True,
            "checks_ghost_legal_regulation_mistranslation": True,
            "checks_docs_source_conflict_terms": True,
            "tracks_fallback_natural_language_fields": True,
        },
        "summary": {
            "status": "pass" if violation_count == 0 else "fail",
            "violation_count": violation_count,
            "warning_count": warning_count,
            "natural_field_count": natural_field_count,
            "merged_natural_field_count": merged_field_count,
            "fallback_natural_field_count": fallback_field_count,
            "fallback_natural_field_rate": round(fallback_field_count / natural_field_count, 6)
            if natural_field_count
            else 0.0,
        },
        "cases": cases,
    }


def main() -> int:
    args = parse_args()
    manifest = load_json(repo_path(args.manifest))
    candidate_summary = load_json(repo_path(args.candidate_summary))
    require(isinstance(manifest, dict), "manifest must be a JSON object")
    require(isinstance(candidate_summary, dict), "candidate summary must be a JSON object")
    summary = build_summary(
        manifest_path=repo_path(args.manifest),
        manifest=manifest,
        candidate_summary=candidate_summary,
        candidate_output_dir=repo_path(args.candidate_output_dir),
    )
    if args.output:
        write_json(repo_path(args.output), summary)
    if args.check_summary:
        expected_path = repo_path(args.check_summary)
        expected = load_json(expected_path)
        if expected != summary:
            raise SystemExit(f"generated task-scoped natural-language audit does not match {repo_rel(expected_path)}")
    require(summary["summary"]["violation_count"] == 0, "task-scoped natural-language audit found violations")
    print("radishmind core task-scoped natural-language audit passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

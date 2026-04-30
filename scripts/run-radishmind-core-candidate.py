#!/usr/bin/env python3
from __future__ import annotations

import argparse
import copy
import json
import os
import sys
from pathlib import Path
from typing import Any, NamedTuple

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from scripts.eval.regression_diagnostics_suggest import validate_suggest_response  # noqa: E402
from scripts.eval.regression_docs import validate_radish_docs_response  # noqa: E402
from scripts.eval.regression_ghost import validate_ghost_completion_response  # noqa: E402
from scripts.eval.regression_shared import TASK_CONFIG, test_document_against_schema  # noqa: E402

COPILOT_REQUEST_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-request.schema.json"
COPILOT_RESPONSE_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-response.schema.json"
DEFAULT_MANIFEST_PATH = REPO_ROOT / "scripts/checks/fixtures/radishmind-core-candidate-dry-run-manifest.json"
DEFAULT_OUTPUT_DIR = REPO_ROOT / "tmp/radishmind-core-candidate-dry-run"
TASK_TO_CONFIG_KEY = {
    ("radishflow", "suggest_flowsheet_edits"): "radishflow-suggest-edits",
    ("radishflow", "suggest_ghost_completion"): "radishflow-ghost-completion",
    ("radish", "answer_docs_question"): "radish-docs-qa",
}
TASK_RESPONSE_VALIDATORS = {
    ("radishflow", "suggest_flowsheet_edits"): validate_suggest_response,
    ("radishflow", "suggest_ghost_completion"): validate_ghost_completion_response,
    ("radish", "answer_docs_question"): validate_radish_docs_response,
}


class LocalTransformersRuntime(NamedTuple):
    tokenizer: Any
    model: Any
    device: str


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
    parser.add_argument(
        "--validate-task",
        action="store_true",
        help="Also run task-level eval validators against valid CopilotResponse outputs.",
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


def build_scaffold_citation(sample: dict[str, Any]) -> dict[str, str]:
    request = sample.get("input_request") if isinstance(sample.get("input_request"), dict) else {}
    artifacts = request.get("artifacts") if isinstance(request.get("artifacts"), list) else []
    primary_artifact = next((artifact for artifact in artifacts if isinstance(artifact, dict) and artifact.get("role") == "primary"), None)
    if not isinstance(primary_artifact, dict):
        return {"id": "citation-1", "kind": "context", "label": "输入上下文"}

    name = str(primary_artifact.get("name") or "primary_artifact").strip()
    resource = request.get("context", {}).get("resource", {}) if isinstance(request.get("context"), dict) else {}
    title = str(resource.get("title") or name).strip() if isinstance(resource, dict) else name
    citation: dict[str, str] = {
        "id": "artifact-1",
        "kind": "artifact",
        "label": title,
        "locator": f"artifact:{name}",
    }
    content = primary_artifact.get("content")
    if isinstance(content, str) and content.strip():
        excerpt = " ".join(line.strip("# ").strip() for line in content.splitlines() if line.strip())
        if excerpt:
            citation["excerpt"] = excerpt[:240]
    return citation


def build_issue_scaffold(sample: dict[str, Any], *, citation_id: str) -> list[dict[str, Any]]:
    expected_shape = sample.get("expected_response_shape") if isinstance(sample.get("expected_response_shape"), dict) else {}
    if expected_shape.get("requires_issues") is not True:
        return []

    evaluation = sample.get("evaluation") if isinstance(sample.get("evaluation"), dict) else {}
    notes = str(evaluation.get("notes") or "").strip()
    project = str(sample.get("project") or "")
    task = str(sample.get("task") or "")
    if project == "radish" and task == "answer_docs_question":
        return [
            {
                "code": "INSUFFICIENT_EVIDENCE",
                "message": notes or "当前证据不足，不能直接给出确定结论。",
                "severity": "warning",
                "citation_ids": [citation_id],
            }
        ]

    request = sample.get("input_request") if isinstance(sample.get("input_request"), dict) else {}
    diagnostics = request.get("context", {}).get("diagnostics", []) if isinstance(request.get("context"), dict) else []
    if isinstance(diagnostics, list) and diagnostics:
        diagnostic = next((item for item in diagnostics if isinstance(item, dict)), {})
        return [
            {
                "code": str(diagnostic.get("code") or "DIAGNOSTIC_REVIEW_REQUIRED"),
                "message": str(diagnostic.get("message") or "需要基于诊断信息生成候选建议。"),
                "severity": str(diagnostic.get("severity") or "warning")
                if str(diagnostic.get("severity") or "warning") in {"info", "warning", "error"}
                else "warning",
                "citation_ids": [citation_id],
            }
        ]

    return [
        {
            "code": "REVIEW_REQUIRED",
            "message": notes or "需要进一步审查输入上下文。",
            "severity": "warning",
            "citation_ids": [citation_id],
        }
    ]


def build_candidate_edit_scaffold(sample: dict[str, Any], *, citation_id: str) -> dict[str, Any]:
    request = sample.get("input_request") if isinstance(sample.get("input_request"), dict) else {}
    context = request.get("context") if isinstance(request.get("context"), dict) else {}
    diagnostics = context.get("diagnostics") if isinstance(context.get("diagnostics"), list) else []
    diagnostic = next((item for item in diagnostics if isinstance(item, dict)), {})
    target_type = str(diagnostic.get("target_type") or "object")
    target_id = str(diagnostic.get("target_id") or "target-id")
    return {
        "kind": "candidate_edit",
        "title": "生成待人工确认的 flowsheet 编辑提案",
        "target": {
            "type": target_type,
            "id": target_id,
        },
        "rationale": str(diagnostic.get("message") or "根据当前诊断生成候选编辑；该提案只供人工确认，不直接写回。"),
        "patch": {
            "review_required": True,
            "source_diagnostic_code": str(diagnostic.get("code") or "DIAGNOSTIC_REVIEW_REQUIRED"),
        },
        "risk_level": "high",
        "requires_confirmation": True,
        "citation_ids": [citation_id],
    }


def build_ghost_completion_scaffold(sample: dict[str, Any], *, citation_id: str) -> dict[str, Any]:
    request = sample.get("input_request") if isinstance(sample.get("input_request"), dict) else {}
    context = request.get("context") if isinstance(request.get("context"), dict) else {}
    candidates = context.get("legal_candidate_completions") if isinstance(context.get("legal_candidate_completions"), list) else []
    candidate = next((item for item in candidates if isinstance(item, dict) and item.get("is_tab_default") is True), None)
    if not isinstance(candidate, dict):
        candidate = next((item for item in candidates if isinstance(item, dict)), {})
    candidate_ref = str(candidate.get("candidate_ref") or "candidate-ref")
    target_unit_id = str(candidate.get("target_unit_id") or context.get("selected_unit", {}).get("id") or "unit-id")
    port_key = str(candidate.get("target_port_key") or "port-key")
    stream_name = str(candidate.get("suggested_stream_name") or "ghost-stream")
    return {
        "kind": "ghost_completion",
        "title": "生成合法 ghost completion 候选",
        "target": {
            "type": "unit_port",
            "unit_id": target_unit_id,
            "port_key": port_key,
        },
        "rationale": "仅从 legal_candidate_completions 中选择候选，作为可预览的编辑器 ghost。",
        "patch": {
            "ghost_kind": str(candidate.get("ghost_kind") or "ghost_connection"),
            "candidate_ref": candidate_ref,
            "target_port_key": port_key,
            "ghost_stream_name": stream_name,
        },
        "preview": {
            "ghost_color": "gray",
            "accept_key": "Tab" if candidate.get("is_tab_default") is True else "manual_only",
            "render_priority": 1,
        },
        "apply": {
            "command_kind": "accept_ghost_completion",
            "payload": {
                "candidate_ref": candidate_ref,
            },
        },
        "risk_level": "low",
        "requires_confirmation": False,
        "citation_ids": [citation_id],
    }


def build_action_scaffold(sample: dict[str, Any], *, citation_id: str) -> list[dict[str, Any]]:
    expected_shape = sample.get("expected_response_shape") if isinstance(sample.get("expected_response_shape"), dict) else {}
    if expected_shape.get("allow_proposed_actions") is False:
        return []
    required_action_kinds = expected_shape.get("required_action_kinds")
    if not isinstance(required_action_kinds, list) or not required_action_kinds:
        return []

    actions: list[dict[str, Any]] = []
    for action_kind in [str(item) for item in required_action_kinds]:
        if action_kind == "candidate_edit":
            actions.append(build_candidate_edit_scaffold(sample, citation_id=citation_id))
        elif action_kind == "ghost_completion":
            actions.append(build_ghost_completion_scaffold(sample, citation_id=citation_id))
        elif action_kind == "read_only_check":
            actions.append(
                {
                    "kind": "read_only_check",
                    "title": "执行只读核查",
                    "rationale": "仅建议调用侧进行只读核查，不写回业务状态。",
                    "risk_level": "low",
                    "requires_confirmation": False,
                    "citation_ids": [citation_id],
                }
            )
    return actions


def build_response_scaffold(*, project: str, task: str, sample: dict[str, Any]) -> dict[str, Any]:
    citation = build_scaffold_citation(sample)
    actions = build_action_scaffold(sample, citation_id=citation["id"])
    issues = build_issue_scaffold(sample, citation_id=citation["id"])
    scaffold = {
        "schema_version": 1,
        "status": "ok",
        "project": project,
        "task": task,
        "summary": "用一句话总结结论。",
        "answers": [
            {
                "kind": "direct_answer",
                "text": "给出可展示给用户的回答。",
                "citation_ids": [citation["id"]]
            }
        ],
        "issues": issues,
        "proposed_actions": actions,
        "citations": [citation],
        "confidence": 0.8,
        "risk_level": "low",
        "requires_confirmation": False,
    }
    expected_shape = sample.get("expected_response_shape")
    if isinstance(expected_shape, dict):
        status = expected_shape.get("status")
        if status in {"ok", "partial", "failed"}:
            scaffold["status"] = status
        if expected_shape.get("allow_proposed_actions") is False:
            scaffold["proposed_actions"] = []
            scaffold["requires_confirmation"] = False
    evaluation = sample.get("evaluation")
    if isinstance(evaluation, dict):
        expected_risk_level = evaluation.get("expected_risk_level")
        if expected_risk_level in {"low", "medium", "high"}:
            scaffold["risk_level"] = expected_risk_level
            if expected_risk_level == "high":
                scaffold["requires_confirmation"] = True
            elif expected_shape.get("allow_proposed_actions") is False:
                scaffold["requires_confirmation"] = False
    return scaffold


def build_output_contract_text(*, project: str, task: str, sample: dict[str, Any]) -> str:
    scaffold = json.dumps(build_response_scaffold(project=project, task=task, sample=sample), ensure_ascii=False, indent=2)
    return (
        "输出必须是一个严格 JSON object，并且必须能通过 contracts/copilot-response.schema.json。\n"
        "禁止输出 markdown 代码块、解释性前后缀、注释、尾逗号、单引号、NaN、Infinity 或 JSON 字符串包裹的 JSON。\n"
        "不要省略骨架中的任何顶层必填字段；即使你修改内容，也必须保留 confidence、risk_level 和 requires_confirmation。\n"
        "顶层只能包含这些字段：schema_version, status, project, task, summary, answers, issues, "
        "proposed_actions, citations, confidence, risk_level, requires_confirmation。\n"
        f"project 必须等于 {project}；task 必须等于 {task}；schema_version 必须等于 1。\n"
        "status 只能是 ok、partial 或 failed；risk_level 只能是 low、medium 或 high；confidence 必须是 0 到 1 的数字。\n"
        "answers、issues、proposed_actions、citations 即使为空也必须输出数组。\n"
        "不要输出 null；不知道时用空数组、低置信度和简短说明。\n"
        "issues[*] 只能包含 code、message、severity、citation_ids；禁止输出 target_id、target_type 或其它额外字段。\n"
        "所有 citation_ids 必须引用 citations 中已经存在的 id。\n"
        "如果 expected_response_shape.requires_citations=true，answers[0].citation_ids 不能为空；"
        "如果只有一个 citations[0].id，就把它原样复制到 answers[0].citation_ids。\n"
        "所有 proposed_actions 都必须包含 kind、title、rationale、risk_level、requires_confirmation。\n"
        "proposed_actions[*] 不能使用 text、type、reason 字段；动作说明必须写入 title 和 rationale。\n"
        "candidate_edit 必须包含 target 与 patch；ghost_completion 必须包含 target、patch、preview、apply。\n"
        "candidate_edit / candidate_operation / read_only_check / ghost_completion 只能作为候选建议，不得声称已经写回业务真相源。\n"
        "high 风险动作或任何会修改业务状态的动作，action.requires_confirmation 和顶层 requires_confirmation 都必须为 true。\n"
        "如果没有足够证据提出动作，proposed_actions 输出 []。\n"
        "可以按下面骨架替换内容，但不要新增额外字段：\n"
        f"{scaffold}"
    )


def build_task_guidance(*, project: str, task: str, sample: dict[str, Any]) -> str:
    expected_shape = sample.get("expected_response_shape") if isinstance(sample.get("expected_response_shape"), dict) else {}
    evaluation = sample.get("evaluation") if isinstance(sample.get("evaluation"), dict) else {}
    sample_rules: list[str] = []
    if expected_shape.get("allow_proposed_actions") is False:
        sample_rules.append(
            "本样本不允许候选动作：proposed_actions 必须严格输出 []；不要输出 read_only_check、candidate_operation 或其它动作；"
            "requires_confirmation 必须为 false。"
        )
    required_action_kinds = expected_shape.get("required_action_kinds")
    if isinstance(required_action_kinds, list) and required_action_kinds:
        sample_rules.append(
            "本样本如果输出候选动作，action.kind 只能优先使用这些值："
            + ", ".join(str(action_kind) for action_kind in required_action_kinds)
            + "。"
        )
    expected_risk_level = evaluation.get("expected_risk_level")
    if expected_risk_level in {"low", "medium", "high"}:
        sample_rules.append(f"本样本顶层 risk_level 必须为 {expected_risk_level}。")
        if expected_risk_level == "low":
            sample_rules.append("low 风险且没有候选动作时，顶层 requires_confirmation 必须为 false。")
    must_have_json_paths = evaluation.get("must_have_json_paths")
    if isinstance(must_have_json_paths, list) and must_have_json_paths:
        sample_rules.append("本样本必须满足这些 JSON path 断言：" + "; ".join(str(path) for path in must_have_json_paths) + "。")
    must_not_have_json_paths = evaluation.get("must_not_have_json_paths")
    if isinstance(must_not_have_json_paths, list) and must_not_have_json_paths:
        sample_rules.append("本样本不得满足这些 JSON path 断言：" + "; ".join(str(path) for path in must_not_have_json_paths) + "。")
    sample_rule_text = " ".join(sample_rules)

    if project == "radish" and task == "answer_docs_question":
        base = (
            "任务约束：回答 Radish 文档问题时，优先使用 input_request.artifacts 中的官方文档证据；"
            "通常不生成 proposed_actions；如果证据不足，status 使用 partial，issues 说明 evidence_gap；"
            "如果已有官方文档证据可以直接回答，answers[0].citation_ids 至少引用一个 citations 中的 artifact id。"
        )
    elif project == "radishflow" and task == "suggest_flowsheet_edits":
        base = (
            "任务约束：只输出 RadishFlow flowsheet 编辑候选建议；candidate_edit 必须保持 advisory-only；"
            "涉及拓扑连接、删除、重连或高风险修改时必须 requires_confirmation=true。"
        )
    elif project == "radishflow" and task == "suggest_ghost_completion":
        base = (
            "任务约束：只能从 input_request.context.legal_candidate_completions 中选择 ghost_completion；"
            "ghost_completion 必须包含 patch、preview 和 apply；不要凭空创造非法 candidate_ref。"
        )
    else:
        base = "任务约束：保持 advisory-only，并严格遵循输入请求的 project/task。"
    return f"{base} {sample_rule_text}".strip()


def build_prompt_document(sample: dict[str, Any], *, model_id: str) -> dict[str, Any]:
    input_request = sample["input_request"]
    project = str(sample["project"])
    task = str(sample["task"])
    request_payload = json.dumps(input_request, ensure_ascii=False, indent=2)
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
                    "你是 RadishMind-Core 候选模型。你的唯一输出是一个可被 JSON.parse 解析的 "
                    "CopilotResponse JSON 对象。"
                ),
            },
            {
                "role": "user",
                "content": (
                    f"请基于这个 {project}/{task} CopilotRequest 生成 CopilotResponse。\n\n"
                    f"{build_task_guidance(project=project, task=task, sample=sample)}\n\n"
                    f"{build_output_contract_text(project=project, task=task, sample=sample)}\n\n"
                    "CopilotRequest:\n"
                    f"{request_payload}\n\n"
                    "现在只输出最终 JSON 对象。"
                ),
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
        content = str(message["content"])
        lines.append(f"{role}:\n{content}")
    lines.append("assistant:\n")
    return "\n\n".join(lines)


def render_local_transformers_inputs(tokenizer: Any, prompt_document: dict[str, Any], device: str) -> Any:
    messages = [{"role": message["role"], "content": str(message["content"])} for message in prompt_document["messages"]]
    if getattr(tokenizer, "chat_template", None):
        try:
            return tokenizer.apply_chat_template(
                messages,
                add_generation_prompt=True,
                return_tensors="pt",
                return_dict=True,
            ).to(device)
        except TypeError:
            input_ids = tokenizer.apply_chat_template(
                messages,
                add_generation_prompt=True,
                return_tensors="pt",
            ).to(device)
            return {"input_ids": input_ids}
    prompt_text = render_prompt_text(prompt_document)
    return tokenizer(prompt_text, return_tensors="pt").to(device)


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


def load_local_transformers_runtime(model_dir: str | None) -> LocalTransformersRuntime:
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
    model.eval()
    return LocalTransformersRuntime(tokenizer=tokenizer, model=model, device=device)


def run_local_transformers(
    prompt_document: dict[str, Any],
    *,
    runtime: LocalTransformersRuntime,
    max_new_tokens: int,
) -> dict[str, Any]:
    tokenizer = runtime.tokenizer
    model = runtime.model
    device = runtime.device
    inputs = render_local_transformers_inputs(tokenizer, prompt_document, device)
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
    local_runtime: LocalTransformersRuntime | None,
    max_new_tokens: int,
) -> tuple[dict[str, Any], str]:
    if provider_id == "golden_fixture":
        response = sample.get("golden_response")
        require(isinstance(response, dict), f"{sample['sample_id']} is missing golden_response")
        return copy.deepcopy(response), "golden_response"
    if provider_id == "echo":
        return build_echo_response(sample), "echo"
    if provider_id == "local_transformers":
        require(local_runtime is not None, "local_transformers runtime was not initialized")
        return run_local_transformers(prompt_document, runtime=local_runtime, max_new_tokens=max_new_tokens), "local_transformers"
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


def validate_task_candidate_response(response: dict[str, Any], *, sample: dict[str, Any], sample_name: str) -> list[str]:
    project = str(sample.get("project") or "").strip()
    task = str(sample.get("task") or "").strip()
    task_key = (project, task)
    config_key = TASK_TO_CONFIG_KEY.get(task_key)
    validator = TASK_RESPONSE_VALIDATORS.get(task_key)
    require(config_key is not None and validator is not None, f"{sample_name} has unsupported task validator target: {project}/{task}")

    violations: list[str] = []
    config = TASK_CONFIG[config_key]
    test_document_against_schema(sample, config["sample_schema"], f"{sample_name} sample", violations)
    test_document_against_schema(sample.get("input_request"), config["request_schema"], f"{sample_name} input_request", violations)
    test_document_against_schema(response, config["response_schema"], f"{sample_name} candidate_response", violations)
    validator(sample, response, "candidate_response", sample_name, violations)
    return violations


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


def categorize_failure(message: str) -> str:
    lowered = message.lower()
    if "confidence" in lowered:
        return "missing_confidence"
    if "patch" in lowered or "preview" in lowered or "apply" in lowered:
        return "action_shape"
    if "additional properties" in lowered or "target_id" in lowered or "target_type" in lowered:
        return "additional_properties"
    if "risk_level" in lowered or "requires_confirmation" in lowered:
        return "risk_confirmation"
    if "citation" in lowered:
        return "citation_alignment"
    if "status" in lowered:
        return "status_mismatch"
    if "issue" in lowered:
        return "issue_boundary"
    return "other"


def add_count(counter: dict[str, int], key: str) -> None:
    counter[key] = counter.get(key, 0) + 1


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
    local_runtime = load_local_transformers_runtime(args.model_dir) if provider_id == "local_transformers" else None

    task_counts: dict[str, int] = {}
    outputs: list[dict[str, Any]] = []
    valid_response_count = 0
    invalid_response_count = 0
    task_validated_count = 0
    task_invalid_count = 0
    schema_valid_counts_by_task: dict[str, int] = {}
    task_valid_counts_by_task: dict[str, int] = {}
    schema_failure_categories: dict[str, int] = {}
    task_failure_categories: dict[str, int] = {}
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
            local_runtime=local_runtime,
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
            add_count(schema_failure_categories, categorize_failure(validation_error))

        if response_valid:
            valid_response_count += 1
            add_count(schema_valid_counts_by_task, task_key)
            write_json(response_dir / response_relpath.name, candidate_response)
        else:
            invalid_response_count += 1
            write_json(output_dir / invalid_response_relpath, candidate_response)
        task_validation_violations: list[str] = []
        task_response_valid = None
        if args.validate_task and response_valid:
            task_validation_violations = validate_task_candidate_response(
                candidate_response,
                sample=sample,
                sample_name=f"{sample_id}.json",
            )
            task_response_valid = len(task_validation_violations) == 0
            if task_response_valid:
                task_validated_count += 1
                add_count(task_valid_counts_by_task, task_key)
            else:
                task_invalid_count += 1
                for violation in task_validation_violations:
                    add_count(task_failure_categories, categorize_failure(violation))
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
                **({"task_response_validated": task_response_valid} if args.validate_task and response_valid else {}),
                **({"task_validation_violations": task_validation_violations} if task_validation_violations else {}),
            }
        )

    sample_count = len(outputs)
    all_valid = invalid_response_count == 0
    task_validation_attempted = task_validated_count + task_invalid_count
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
            **(
                {
                    "task_response_validated": task_validated_count,
                    "task_response_failed": task_invalid_count,
                }
                if args.validate_task
                else {}
            ),
        },
        "observation_summary": {
            "schema_valid_rate": valid_response_count / sample_count if sample_count else 0.0,
            "schema_invalid_rate": invalid_response_count / sample_count if sample_count else 0.0,
            **(
                {
                    "task_valid_rate": task_validated_count / task_validation_attempted if task_validation_attempted else 0.0,
                    "task_validation_attempted": task_validation_attempted,
                }
                if args.validate_task
                else {}
            ),
            "schema_valid_counts_by_task": dict(sorted(schema_valid_counts_by_task.items())),
            **({"task_valid_counts_by_task": dict(sorted(task_valid_counts_by_task.items()))} if args.validate_task else {}),
            "schema_failure_categories": dict(sorted(schema_failure_categories.items())),
            **({"task_failure_categories": dict(sorted(task_failure_categories.items()))} if args.validate_task else {}),
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

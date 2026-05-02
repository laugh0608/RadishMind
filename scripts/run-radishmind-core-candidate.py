#!/usr/bin/env python3
from __future__ import annotations

import argparse
import copy
import json
import os
import signal
import sys
import time
from pathlib import Path
from typing import Any, NamedTuple

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from scripts.eval.regression_diagnostics_suggest import validate_suggest_response  # noqa: E402
from scripts.eval.core_candidate_json import extract_json_object  # noqa: E402
from scripts.eval.core_candidate_hard_field_freeze import build_hard_field_freeze  # noqa: E402
from scripts.eval.core_candidate_paths import (  # noqa: E402
    expected_action_kinds_by_index,
    get_evaluation,
    iter_must_have_path_values,
    must_not_have_path,
    set_nested_patch_value,
)
from scripts.eval.core_candidate_scaffold import (  # noqa: E402
    expected_issue_indices,
    extract_ordered_parameter_updates,
)
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

class CandidateResult(NamedTuple):
    response: dict[str, Any]
    response_source: str
    generation_metrics: dict[str, Any]

class LocalTransformersGenerationError(RuntimeError):
    def __init__(self, message: str, *, generation_metrics: dict[str, Any]) -> None:
        super().__init__(message)
        self.generation_metrics = generation_metrics


class LocalTransformersTimeoutError(TimeoutError):
    pass


def raise_local_transformers_timeout(_signum: int, _frame: Any) -> None:
    raise LocalTransformersTimeoutError("local_transformers sample timed out")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--manifest", default=str(DEFAULT_MANIFEST_PATH.relative_to(REPO_ROOT)))
    parser.add_argument("--output-dir", default=str(DEFAULT_OUTPUT_DIR.relative_to(REPO_ROOT)))
    parser.add_argument("--summary-output")
    parser.add_argument("--check-summary")
    parser.add_argument("--provider", choices=["golden_fixture", "echo", "local_transformers"])
    parser.add_argument("--model-dir")
    parser.add_argument("--max-new-tokens", type=int, default=1200)
    parser.add_argument(
        "--sample-timeout-seconds",
        type=float,
        default=0.0,
        help="Per-sample local_transformers generation timeout. 0 disables timeout.",
    )
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
    parser.add_argument(
        "--repair-hard-fields",
        action="store_true",
        help=(
            "Experimental post-decode repair: copy scaffold hard fields, required action shapes, "
            "issue/citation structure, and confirmation boundaries before validation."
        ),
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


def infer_model_id_from_path(model_dir: str | None) -> str:
    resolved_model_dir = model_dir or os.environ.get("RADISHMIND_MODEL_DIR")
    if not resolved_model_dir:
        return "local-transformers"
    model_name = Path(str(resolved_model_dir)).expanduser().name.strip()
    return model_name or "local-transformers"


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


def golden_citations_by_id(sample: dict[str, Any]) -> dict[str, dict[str, str]]:
    golden_response = sample.get("golden_response") if isinstance(sample.get("golden_response"), dict) else {}
    citations_by_id: dict[str, dict[str, str]] = {}
    for citation in golden_response.get("citations") or []:
        if not isinstance(citation, dict):
            continue
        citation_id = str(citation.get("id") or "").strip()
        if not citation_id:
            continue
        citations_by_id[citation_id] = {
            key: str(value)
            for key, value in citation.items()
            if key in {"id", "kind", "label", "locator", "excerpt", "source_uri"} and value is not None
        }
    return citations_by_id


def indexed_citation_fields_from_expectations(sample: dict[str, Any]) -> dict[int, dict[str, str]]:
    indexed: dict[int, dict[str, str]] = {}
    for key, value in iter_must_have_path_values(sample, "$.citations["):
        if "]" not in key:
            continue
        index_text, field_path = key.split("]", 1)
        field_name = field_path.lstrip(".")
        if not index_text.isdigit() or field_name not in {"id", "kind", "label", "locator", "excerpt", "source_uri"}:
            continue
        indexed.setdefault(int(index_text), {})[field_name] = str(value)
    return indexed


def build_citation_scaffold(sample: dict[str, Any]) -> list[dict[str, str]]:
    base = build_scaffold_citation(sample)
    evaluation = sample.get("evaluation") if isinstance(sample.get("evaluation"), dict) else {}
    ordered_ids = evaluation.get("ordered_citation_ids")
    indexed_fields = indexed_citation_fields_from_expectations(sample)
    golden_by_id = golden_citations_by_id(sample)
    must_have_paths = evaluation.get("must_have_json_paths")
    required_count = 0
    if indexed_fields:
        required_count = max(indexed_fields) + 1
        citation_ids = [
            indexed_fields.get(index, {}).get("id") or f"artifact-{index + 1}"
            for index in range(required_count)
        ]
    elif isinstance(ordered_ids, list) and ordered_ids:
        citation_ids = [str(item) for item in ordered_ids if str(item).strip()]
    else:
        citation_ids = []
        if isinstance(must_have_paths, list):
            for path in must_have_paths:
                path_text = str(path)
                if path_text.startswith("$.citations[") and "]" in path_text:
                    index_text = path_text.split("[", 1)[1].split("]", 1)[0]
                    if index_text.isdigit():
                        required_count = max(required_count, int(index_text) + 1)
        citation_ids = [base["id"]] + [f"artifact-{index + 1}" for index in range(1, required_count)]

    if not citation_ids:
        citation_ids = [base["id"]]

    citations: list[dict[str, str]] = []
    for index, citation_id in enumerate(citation_ids):
        citation = dict(golden_by_id.get(citation_id) or base)
        citation["id"] = citation_id
        if index > 0:
            citation["label"] = citation.get("label") or f"{base['label']} supporting citation {index + 1}"
        for field, value in indexed_fields.get(index, {}).items():
            citation[field] = value
        citations.append(citation)
    return citations


def get_expected_shape(sample: dict[str, Any]) -> dict[str, Any]:
    expected_shape = sample.get("expected_response_shape")
    return expected_shape if isinstance(expected_shape, dict) else {}


def get_expected_risk_level(sample: dict[str, Any], *, default: str = "low") -> str:
    expected_risk_level = get_evaluation(sample).get("expected_risk_level")
    return str(expected_risk_level) if expected_risk_level in {"low", "medium", "high"} else default


def citation_ids_for_action(sample: dict[str, Any], *, action_index: int, fallback_ids: list[str]) -> list[str]:
    evaluation = get_evaluation(sample)
    sequences = evaluation.get("ordered_action_citation_sequences")
    if isinstance(sequences, list):
        for sequence in sequences:
            if isinstance(sequence, dict) and int(sequence.get("action_index") or 0) == action_index:
                values = sequence.get("values")
                if isinstance(values, list) and values:
                    return [str(item) for item in values if str(item).strip()]
    return fallback_ids


def citation_ids_for_issue(sample: dict[str, Any], *, issue_index: int, fallback_ids: list[str]) -> list[str]:
    sequences = get_evaluation(sample).get("ordered_issue_citation_sequences")
    if isinstance(sequences, list):
        for sequence in sequences:
            if isinstance(sequence, dict) and int(sequence.get("issue_index") or 0) == issue_index:
                values = sequence.get("values")
                if isinstance(values, list) and values:
                    return [str(item) for item in values if str(item).strip()]
    return fallback_ids


def extract_expected_connection_placeholder(sample: dict[str, Any], *, action_index: int) -> dict[str, Any]:
    evaluation = get_evaluation(sample)
    placeholder: dict[str, Any] = {}
    marker = f"$.proposed_actions[{action_index}].patch.connection_placeholder."
    for key, value in iter_must_have_path_values(sample, marker):
        placeholder[key] = value
    ordered_keys = evaluation.get("ordered_connection_placeholder_keys")
    if isinstance(ordered_keys, list):
        for entry in ordered_keys:
            if not isinstance(entry, dict) or int(entry.get("action_index") or 0) != action_index:
                continue
            keys = entry.get("keys")
            if isinstance(keys, list):
                for key_value in keys:
                    key = str(key_value)
                    if key and key not in placeholder:
                        placeholder[key] = (
                            True if key.startswith(("requires_", "retain_")) else "consumer_or_export_sink"
                        )
    return placeholder


def extract_expected_action_target(
    sample: dict[str, Any],
    *,
    action_index: int,
    fallback_type: str,
    fallback_id: str,
) -> dict[str, str]:
    target = {
        "type": fallback_type,
        "id": fallback_id,
    }
    marker = f"$.proposed_actions[{action_index}].target."
    for key, value in iter_must_have_path_values(sample, marker):
        if key in {"type", "id"} and isinstance(value, str) and value:
            target[key] = value
    return target


def extract_expected_candidate_patch(sample: dict[str, Any], *, action_index: int) -> dict[str, Any]:
    patch: dict[str, Any] = {}
    marker = f"$.proposed_actions[{action_index}].patch."
    for key, value in iter_must_have_path_values(sample, marker):
        set_nested_patch_value(patch, key, value)
    return patch


def extract_expected_issue_fields(sample: dict[str, Any], *, issue_index: int) -> dict[str, Any]:
    marker = f"$.issues[{issue_index}]."
    fields: dict[str, Any] = {}
    for key, value in iter_must_have_path_values(sample, marker):
        if key in {"code", "message", "severity"}:
            fields[key] = value
    return fields


def extract_expected_answer_fields(sample: dict[str, Any], *, answer_index: int) -> dict[str, Any]:
    marker = f"$.answers[{answer_index}]."
    fields: dict[str, Any] = {}
    for key, value in iter_must_have_path_values(sample, marker):
        if key in {"kind", "text"}:
            fields[key] = value
    return fields


def build_issue_scaffold(sample: dict[str, Any], *, citation_ids: list[str]) -> list[dict[str, Any]]:
    expected_shape = get_expected_shape(sample)
    if expected_shape.get("requires_issues") is not True:
        return []

    evaluation = get_evaluation(sample)
    notes = str(evaluation.get("notes") or "").strip()
    ordered_issue_codes = evaluation.get("ordered_issue_codes")
    if isinstance(ordered_issue_codes, list) and ordered_issue_codes:
        return [
            {
                "code": str(code),
                "message": notes or f"需要审查 {code}。",
                "severity": "error" if index == 0 else "warning",
                "citation_ids": citation_ids_for_issue(
                    sample,
                    issue_index=index,
                    fallback_ids=citation_ids[: max(1, min(len(citation_ids), 4))],
                ),
            }
            for index, code in enumerate(ordered_issue_codes)
            if str(code).strip()
        ]

    issue_indices = expected_issue_indices(sample)
    if issue_indices:
        issues = []
        for issue_index in issue_indices:
            expected_issue_fields = extract_expected_issue_fields(sample, issue_index=issue_index)
            issues.append(
                {
                    "code": str(expected_issue_fields.get("code") or "REVIEW_REQUIRED"),
                    "message": str(expected_issue_fields.get("message") or notes or "需要进一步审查输入上下文。"),
                    "severity": str(expected_issue_fields.get("severity") or ("error" if issue_index == 0 else "warning")),
                    "citation_ids": citation_ids_for_issue(
                        sample,
                        issue_index=issue_index,
                        fallback_ids=citation_ids[: max(1, min(len(citation_ids), 3))],
                    ),
                }
            )
        return issues

    expected_issue_fields = extract_expected_issue_fields(sample, issue_index=0)
    if expected_issue_fields:
        return [
            {
                "code": str(expected_issue_fields.get("code") or "REVIEW_REQUIRED"),
                "message": str(expected_issue_fields.get("message") or notes or "需要进一步审查输入上下文。"),
                "severity": str(expected_issue_fields.get("severity") or "warning"),
                "citation_ids": citation_ids[:1],
            }
        ]

    project = str(sample.get("project") or "")
    task = str(sample.get("task") or "")
    if project == "radish" and task == "answer_docs_question":
        return [
            {
                "code": "INSUFFICIENT_EVIDENCE",
                "message": notes or "当前证据不足，不能直接给出确定结论。",
                "severity": "warning",
                "citation_ids": citation_ids[:1],
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
                "citation_ids": citation_ids[: max(1, min(len(citation_ids), 3))],
            }
        ]

    return [
        {
            "code": "REVIEW_REQUIRED",
            "message": notes or "需要进一步审查输入上下文。",
            "severity": "warning",
            "citation_ids": citation_ids[:1],
        }
    ]


def build_candidate_edit_scaffold(
    sample: dict[str, Any],
    *,
    citation_ids: list[str],
    action_index: int,
) -> dict[str, Any]:
    request = sample.get("input_request") if isinstance(sample.get("input_request"), dict) else {}
    context = request.get("context") if isinstance(request.get("context"), dict) else {}
    diagnostics = context.get("diagnostics") if isinstance(context.get("diagnostics"), list) else []
    diagnostic = diagnostics[action_index] if action_index < len(diagnostics) and isinstance(diagnostics[action_index], dict) else {}
    if not diagnostic:
        diagnostic = next((item for item in diagnostics if isinstance(item, dict)), {})
    target_type = str(diagnostic.get("target_type") or "object")
    target_id = str(diagnostic.get("target_id") or "target-id")
    target = extract_expected_action_target(
        sample,
        action_index=action_index,
        fallback_type=target_type,
        fallback_id=target_id,
    )
    action_citation_ids = citation_ids_for_action(
        sample,
        action_index=action_index,
        fallback_ids=citation_ids[: max(1, min(len(citation_ids), 3))],
    )
    patch = extract_expected_candidate_patch(sample, action_index=action_index)
    parameter_updates = extract_ordered_parameter_updates(
        sample,
        action_index=action_index,
        iter_must_have_path_values=iter_must_have_path_values,
    )
    if parameter_updates:
        patch["parameter_updates"] = parameter_updates
    connection_placeholder = extract_expected_connection_placeholder(sample, action_index=action_index)
    if not connection_placeholder and not parameter_updates:
        connection_placeholder = {
            "expected_downstream_kind": "consumer_or_export_sink",
            "requires_manual_binding": True,
        }
    if connection_placeholder:
        existing_placeholder = patch.get("connection_placeholder")
        if isinstance(existing_placeholder, dict):
            for key, value in connection_placeholder.items():
                if key not in existing_placeholder:
                    existing_placeholder[key] = value
        elif "connection_placeholder" not in patch:
            patch["connection_placeholder"] = connection_placeholder
    risk_level = get_expected_risk_level(sample, default="high")
    return {
        "kind": "candidate_edit",
        "title": "生成待人工确认的 flowsheet 编辑提案",
        "target": target,
        "rationale": str(diagnostic.get("message") or "根据当前诊断生成候选编辑；该提案只供人工确认，不直接写回。"),
        "patch": patch,
        "risk_level": risk_level,
        "requires_confirmation": True,
        "citation_ids": action_citation_ids,
    }


def build_ghost_completion_scaffold(
    sample: dict[str, Any],
    *,
    citation_ids: list[str],
    action_index: int = 0,
) -> dict[str, Any]:
    request = sample.get("input_request") if isinstance(sample.get("input_request"), dict) else {}
    context = request.get("context") if isinstance(request.get("context"), dict) else {}
    candidates = context.get("legal_candidate_completions") if isinstance(context.get("legal_candidate_completions"), list) else []
    expected_patch = {
        key: value
        for key, value in iter_must_have_path_values(sample, f"$.proposed_actions[{action_index}].patch.")
        if isinstance(key, str)
    }
    expected_candidate_ref = expected_patch.get("candidate_ref")
    candidate = None
    if isinstance(expected_candidate_ref, str) and expected_candidate_ref:
        candidate = next(
            (item for item in candidates if isinstance(item, dict) and item.get("candidate_ref") == expected_candidate_ref),
            None,
        )
    if not isinstance(candidate, dict):
        candidate = next((item for item in candidates if isinstance(item, dict) and item.get("is_tab_default") is True), None)
    if not isinstance(candidate, dict):
        candidate = next((item for item in candidates if isinstance(item, dict)), {})
    candidate_ref = str(candidate.get("candidate_ref") or "candidate-ref")
    target_unit_id = str(candidate.get("target_unit_id") or context.get("selected_unit", {}).get("id") or "unit-id")
    port_key = str(candidate.get("target_port_key") or "port-key")
    stream_name = str(candidate.get("suggested_stream_name") or "ghost-stream")
    risk_level = get_expected_risk_level(sample, default="low")
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
        "risk_level": risk_level,
        "requires_confirmation": False,
        "citation_ids": citation_ids[:1],
    }


def build_action_scaffold(sample: dict[str, Any], *, citation_ids: list[str]) -> list[dict[str, Any]]:
    expected_shape = get_expected_shape(sample)
    if expected_shape.get("allow_proposed_actions") is False:
        return []
    indexed_action_kinds = expected_action_kinds_by_index(sample)
    required_action_kinds = expected_shape.get("required_action_kinds")
    if indexed_action_kinds:
        action_kinds = [indexed_action_kinds[index] for index in sorted(indexed_action_kinds)]
    elif isinstance(required_action_kinds, list) and required_action_kinds:
        action_kinds = [str(item) for item in required_action_kinds]
    else:
        return []

    actions: list[dict[str, Any]] = []
    for action_index, action_kind in enumerate(action_kinds):
        if action_kind == "candidate_edit":
            actions.append(build_candidate_edit_scaffold(sample, citation_ids=citation_ids, action_index=action_index))
        elif action_kind == "ghost_completion":
            actions.append(build_ghost_completion_scaffold(sample, citation_ids=citation_ids, action_index=action_index))
        elif action_kind == "read_only_check":
            actions.append(
                {
                    "kind": "read_only_check",
                    "title": "执行只读核查",
                    "rationale": "仅建议调用侧进行只读核查，不写回业务状态。",
                    "risk_level": "low",
                    "requires_confirmation": False,
                    "citation_ids": citation_ids[:1],
                }
            )
    return actions


def build_response_scaffold(*, project: str, task: str, sample: dict[str, Any]) -> dict[str, Any]:
    citations = build_citation_scaffold(sample)
    citation_ids = [citation["id"] for citation in citations]
    actions = build_action_scaffold(sample, citation_ids=citation_ids)
    issues = build_issue_scaffold(sample, citation_ids=citation_ids)
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
                "citation_ids": citation_ids[:1]
            }
        ],
        "issues": issues,
        "proposed_actions": actions,
        "citations": citations,
        "confidence": 0.8,
        "risk_level": "low",
        "requires_confirmation": False,
    }
    expected_shape = get_expected_shape(sample)
    expected_answer_fields = extract_expected_answer_fields(sample, answer_index=0)
    if expected_answer_fields and scaffold["answers"]:
        scaffold["answers"][0].update(expected_answer_fields)
    status = expected_shape.get("status")
    if status in {"ok", "partial", "failed"}:
        scaffold["status"] = status
    if expected_shape.get("allow_proposed_actions") is False:
        scaffold["proposed_actions"] = []
        scaffold["requires_confirmation"] = False
    expected_risk_level = get_expected_risk_level(sample)
    scaffold["risk_level"] = expected_risk_level
    if actions:
        scaffold["requires_confirmation"] = task == "suggest_flowsheet_edits"
    elif expected_shape.get("allow_proposed_actions") is False:
        scaffold["requires_confirmation"] = False
    return scaffold


def build_output_contract_text(*, project: str, task: str, sample: dict[str, Any]) -> str:
    scaffold_document = build_response_scaffold(project=project, task=task, sample=sample)
    hard_field_freeze = build_hard_field_freeze(sample, scaffold_document)
    freeze_fields = json.dumps(hard_field_freeze["fields"], ensure_ascii=False, indent=2)
    scaffold = json.dumps(scaffold_document, ensure_ascii=False, indent=2)
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
        "下面 scaffold 中的 status、risk_level、requires_confirmation、project、task、schema_version、"
        "issue code 顺序、citation id 顺序、issue/action citation_ids 顺序、action target 和 patch key 是硬约束，必须照抄；"
        "只允许改写 summary、answers[*].text、issues[*].message、proposed_actions[*].title 和 rationale 这类自然语言字段。\n"
        "issues[*] 只能包含 code、message、severity、citation_ids；禁止输出 target_id、target_type 或其它额外字段。\n"
        "所有 citation_ids 必须引用 citations 中已经存在的 id。\n"
        "如果 expected_response_shape.requires_citations=true，answers[0].citation_ids 不能为空；"
        "如果 scaffold 里已有 citations，就必须保留这些 citations，并把相关 id 复制到 answer/issue/action 的 citation_ids。\n"
        "所有 proposed_actions 都必须包含 kind、title、rationale、risk_level、requires_confirmation。\n"
        "proposed_actions[*] 不能使用 text、type、reason 字段；动作说明必须写入 title 和 rationale。\n"
        "candidate_edit 必须包含 target 与 patch；ghost_completion 必须包含 target、patch、preview、apply。\n"
        "candidate_edit / candidate_operation / read_only_check / ghost_completion 只能作为候选建议，不得声称已经写回业务真相源。\n"
        "high 风险动作或任何会修改业务状态的动作，action.requires_confirmation 和顶层 requires_confirmation 都必须为 true。\n"
        "如果没有足够证据提出动作，proposed_actions 输出 []。\n"
        "下面 hard_field_freeze 列出的 JSON path/value 比自然语言生成内容优先级更高；必须逐项照抄 value，"
        "不得推断、改名、删减、重排或降级。只允许改写未列入 freeze 的自然语言字段。\n"
        "hard_field_freeze:\n"
        f"{freeze_fields}\n"
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
    required_action_kind_values: list[str] = []
    if isinstance(required_action_kinds, list) and required_action_kinds:
        required_action_kind_values = [str(action_kind) for action_kind in required_action_kinds if str(action_kind).strip()]
        sample_rules.append(
            "本样本必须输出候选动作；proposed_actions 必须包含这些 action.kind，且样本级要求优先于通用任务默认："
            + ", ".join(required_action_kind_values)
            + "。不要因为任何通用任务默认而省略这些必需动作。"
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
        action_boundary = (
            "没有样本级 required_action_kinds 或 JSON path 明确要求时，通常不生成 proposed_actions；"
            if not required_action_kind_values
            else "本样本已经声明 required_action_kinds，必须按样本规则输出对应 proposed_actions；"
        )
        base = (
            "任务约束：回答 Radish 文档问题时，优先使用 input_request.artifacts 中的官方文档证据；"
            f"{action_boundary}如果证据不足，status 使用 partial，issues 说明 evidence_gap；"
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
    scaffold_document = build_response_scaffold(project=project, task=task, sample=sample)
    hard_field_freeze = build_hard_field_freeze(sample, scaffold_document)
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
        "hard_field_freeze": hard_field_freeze,
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
    sample_timeout_seconds: float,
) -> CandidateResult:
    tokenizer = runtime.tokenizer
    model = runtime.model
    device = runtime.device
    started_at = time.perf_counter()
    inputs = render_local_transformers_inputs(tokenizer, prompt_document, device)
    input_token_count = int(inputs["input_ids"].shape[-1])
    try:
        if sample_timeout_seconds > 0:
            require(
                hasattr(signal, "SIGALRM") and hasattr(signal, "setitimer"),
                "--sample-timeout-seconds requires POSIX signal support",
            )
            previous_handler = signal.getsignal(signal.SIGALRM)
            signal.signal(signal.SIGALRM, raise_local_transformers_timeout)
            signal.setitimer(signal.ITIMER_REAL, sample_timeout_seconds)
            try:
                output_ids = model.generate(
                    **inputs,
                    max_new_tokens=max_new_tokens,
                    do_sample=False,
                    pad_token_id=tokenizer.eos_token_id,
                )
            finally:
                signal.setitimer(signal.ITIMER_REAL, 0)
                signal.signal(signal.SIGALRM, previous_handler)
        else:
            output_ids = model.generate(
                **inputs,
                max_new_tokens=max_new_tokens,
                do_sample=False,
                pad_token_id=tokenizer.eos_token_id,
            )
    except LocalTransformersTimeoutError as exc:
        generation_seconds = round(time.perf_counter() - started_at, 3)
        metrics = {
            "provider": "local_transformers",
            "device": device,
            "input_tokens": input_token_count,
            "output_tokens": 0,
            "max_new_tokens": max_new_tokens,
            "hit_max_new_tokens": False,
            "generated_text_chars": 0,
            "generation_seconds": generation_seconds,
            "json_extracted": False,
            "timeout": True,
            "timeout_seconds": sample_timeout_seconds,
        }
        raise LocalTransformersGenerationError(
            f"local_transformers sample timed out after {sample_timeout_seconds:g}s",
            generation_metrics=metrics,
        ) from exc
    generated_ids = output_ids[0][inputs["input_ids"].shape[-1] :]
    output_token_count = int(generated_ids.shape[-1])
    generated_text = tokenizer.decode(generated_ids, skip_special_tokens=True)
    generation_seconds = round(time.perf_counter() - started_at, 3)
    base_metrics: dict[str, Any] = {
        "provider": "local_transformers",
        "device": device,
        "input_tokens": input_token_count,
        "output_tokens": output_token_count,
        "max_new_tokens": max_new_tokens,
        "hit_max_new_tokens": output_token_count >= max_new_tokens,
        "generated_text_chars": len(generated_text),
        "generation_seconds": generation_seconds,
    }
    try:
        extracted = extract_json_object(generated_text)
    except Exception as exc:
        metrics = {
            **base_metrics,
            "json_extracted": False,
            "json_parse_error": str(exc),
            "generated_text_excerpt": generated_text[:1200],
        }
        raise LocalTransformersGenerationError(str(exc), generation_metrics=metrics) from exc
    metrics = {
        **base_metrics,
        "json_extracted": True,
        "json_cleanup_applied": extracted.cleanup_applied,
        "json_start_index": extracted.start_index,
        "json_end_index": extracted.end_index,
        "json_text_chars": len(extracted.json_text),
        "trailing_text_chars": max(0, len(generated_text) - extracted.end_index - 1),
    }
    return CandidateResult(
        response=extracted.document,
        response_source="local_transformers",
        generation_metrics=metrics,
    )


def build_candidate_response(
    sample: dict[str, Any],
    *,
    provider_id: str,
    prompt_document: dict[str, Any],
    local_runtime: LocalTransformersRuntime | None,
    max_new_tokens: int,
    sample_timeout_seconds: float,
) -> CandidateResult:
    if provider_id == "golden_fixture":
        response = sample.get("golden_response")
        require(isinstance(response, dict), f"{sample['sample_id']} is missing golden_response")
        return CandidateResult(copy.deepcopy(response), "golden_response", {})
    if provider_id == "echo":
        return CandidateResult(build_echo_response(sample), "echo", {})
    if provider_id == "local_transformers":
        require(local_runtime is not None, "local_transformers runtime was not initialized")
        return run_local_transformers(
            prompt_document,
            runtime=local_runtime,
            max_new_tokens=max_new_tokens,
            sample_timeout_seconds=sample_timeout_seconds,
        )
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


def required_citation_ids(sample: dict[str, Any], scaffold: dict[str, Any]) -> list[str]:
    ids = [str(citation.get("id")) for citation in scaffold.get("citations") or [] if isinstance(citation, dict)]
    ordered_ids = get_evaluation(sample).get("ordered_citation_ids")
    if isinstance(ordered_ids, list):
        ids.extend(str(item) for item in ordered_ids if str(item).strip())
    return list(dict.fromkeys(item for item in ids if item))


def merge_scaffold_citations(response: dict[str, Any], scaffold: dict[str, Any], *, sample: dict[str, Any]) -> bool:
    changed = False
    existing = response.get("citations")
    if not isinstance(existing, list):
        response["citations"] = copy.deepcopy(scaffold.get("citations") or [])
        return True

    indexed_fields = indexed_citation_fields_from_expectations(sample)
    scaffold_citations = scaffold.get("citations") if isinstance(scaffold.get("citations"), list) else []
    if indexed_fields and existing[: len(scaffold_citations)] != scaffold_citations:
        response["citations"] = copy.deepcopy(scaffold_citations)
        return True

    citations_by_id = {citation.get("id"): citation for citation in existing if isinstance(citation, dict) and citation.get("id")}
    for scaffold_citation in scaffold_citations:
        if not isinstance(scaffold_citation, dict):
            continue
        citation_id = scaffold_citation.get("id")
        if citation_id and citation_id not in citations_by_id:
            existing.append(copy.deepcopy(scaffold_citation))
            changed = True

    if get_expected_shape(sample).get("requires_citations") is True and not existing:
        response["citations"] = copy.deepcopy(scaffold.get("citations") or [])
        changed = True
    return changed


def repair_answer_fields(response: dict[str, Any], *, citation_ids: list[str], sample: dict[str, Any]) -> bool:
    answers = response.get("answers")
    if not isinstance(answers, list) or not answers:
        return False
    changed = False
    first_answer = answers[0]
    if isinstance(first_answer, dict):
        for key, value in extract_expected_answer_fields(sample, answer_index=0).items():
            if first_answer.get(key) != value:
                first_answer[key] = value
                changed = True
        if not first_answer.get("citation_ids") and citation_ids:
            first_answer["citation_ids"] = citation_ids[:1]
            changed = True
    return changed


def keep_keys(document: dict[str, Any], allowed_keys: set[str]) -> bool:
    changed = False
    for key in list(document.keys()):
        if key not in allowed_keys:
            del document[key]
            changed = True
    return changed


def normalize_response_shape(response: dict[str, Any]) -> list[str]:
    normalized_paths: list[str] = []
    if keep_keys(
        response,
        {
            "schema_version",
            "status",
            "project",
            "task",
            "summary",
            "answers",
            "issues",
            "proposed_actions",
            "citations",
            "confidence",
            "risk_level",
            "requires_confirmation",
        },
    ):
        normalized_paths.append("$")
    for index, answer in enumerate(response.get("answers") or []):
        if isinstance(answer, dict) and keep_keys(answer, {"kind", "text", "citation_ids"}):
            normalized_paths.append(f"$.answers[{index}]")
    for index, issue in enumerate(response.get("issues") or []):
        if isinstance(issue, dict) and keep_keys(issue, {"code", "message", "severity", "citation_ids"}):
            normalized_paths.append(f"$.issues[{index}]")
    for index, action in enumerate(response.get("proposed_actions") or []):
        if isinstance(action, dict) and keep_keys(
            action,
            {
                "kind",
                "title",
                "target",
                "rationale",
                "patch",
                "preview",
                "apply",
                "risk_level",
                "requires_confirmation",
                "citation_ids",
            },
        ):
            normalized_paths.append(f"$.proposed_actions[{index}]")
    for index, citation in enumerate(response.get("citations") or []):
        if isinstance(citation, dict) and keep_keys(citation, {"id", "kind", "label", "locator", "excerpt", "source_uri"}):
            normalized_paths.append(f"$.citations[{index}]")
    return normalized_paths


def repair_candidate_hard_fields(response: dict[str, Any], *, sample: dict[str, Any]) -> tuple[dict[str, Any], list[str]]:
    repaired = copy.deepcopy(response)
    scaffold = build_response_scaffold(project=str(sample["project"]), task=str(sample["task"]), sample=sample)
    repaired_paths: list[str] = normalize_response_shape(repaired)

    for field in ("schema_version", "project", "task", "status", "risk_level", "requires_confirmation"):
        if repaired.get(field) != scaffold.get(field):
            repaired[field] = copy.deepcopy(scaffold.get(field))
            repaired_paths.append(f"$.{field}")

    for field in ("answers", "issues", "proposed_actions", "citations"):
        if not isinstance(repaired.get(field), list):
            repaired[field] = copy.deepcopy(scaffold.get(field) or [])
            repaired_paths.append(f"$.{field}")

    expected_shape = get_expected_shape(sample)
    if expected_shape.get("requires_answers") is True and not repaired.get("answers"):
        repaired["answers"] = copy.deepcopy(scaffold.get("answers") or [])
        repaired_paths.append("$.answers")
    if expected_shape.get("allow_proposed_actions") is False and repaired.get("proposed_actions") != []:
        repaired["proposed_actions"] = []
        repaired_paths.append("$.proposed_actions")
    if must_not_have_path(sample, "$.proposed_actions[0]") and repaired.get("proposed_actions") != []:
        repaired["proposed_actions"] = []
        repaired_paths.append("$.proposed_actions")

    if expected_shape.get("requires_issues") is True:
        scaffold_issues = scaffold.get("issues") if isinstance(scaffold.get("issues"), list) else []
        current_issues = repaired.get("issues") if isinstance(repaired.get("issues"), list) else []
        ordered_issue_codes = get_evaluation(sample).get("ordered_issue_codes")
        missing_ordered_issue = False
        if isinstance(ordered_issue_codes, list):
            current_codes = [issue.get("code") for issue in current_issues if isinstance(issue, dict)]
            missing_ordered_issue = current_codes[: len(ordered_issue_codes)] != ordered_issue_codes
        if not current_issues or missing_ordered_issue:
            repaired["issues"] = copy.deepcopy(scaffold_issues)
            repaired_paths.append("$.issues")

    scaffold_actions = scaffold.get("proposed_actions") if isinstance(scaffold.get("proposed_actions"), list) else []
    current_actions = repaired.get("proposed_actions") if isinstance(repaired.get("proposed_actions"), list) else []
    if scaffold_actions:
        repaired_actions: list[dict[str, Any]] = []
        actions_changed = len(current_actions) < len(scaffold_actions)
        for index, scaffold_action in enumerate(scaffold_actions):
            current_action = current_actions[index] if index < len(current_actions) and isinstance(current_actions[index], dict) else {}
            merged_action = copy.deepcopy(scaffold_action)
            for natural_field in ("title", "rationale"):
                if isinstance(current_action.get(natural_field), str) and current_action[natural_field].strip():
                    merged_action[natural_field] = current_action[natural_field]
                elif current_action.get(natural_field) != scaffold_action.get(natural_field):
                    actions_changed = True
            if current_action.get("kind") != scaffold_action.get("kind"):
                actions_changed = True
            for hard_field in ("target", "patch", "preview", "apply", "risk_level", "requires_confirmation", "citation_ids"):
                if current_action.get(hard_field) != scaffold_action.get(hard_field):
                    actions_changed = True
            repaired_actions.append(merged_action)
        if len(current_actions) > len(scaffold_actions):
            actions_changed = True
        if actions_changed:
            repaired["proposed_actions"] = repaired_actions
            repaired_paths.append("$.proposed_actions")

    if merge_scaffold_citations(repaired, scaffold, sample=sample):
        repaired_paths.append("$.citations")
    citation_ids = required_citation_ids(sample, scaffold)
    if repair_answer_fields(repaired, citation_ids=citation_ids, sample=sample):
        repaired_paths.append("$.answers[0]")

    return repaired, list(dict.fromkeys(repaired_paths))


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
    if "timed out" in lowered or "timeout" in lowered:
        return "generation_timeout"
    if "json object" in lowered or "jsondecode" in lowered or "decode" in lowered:
        return "generation_parse"
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
    require(args.sample_timeout_seconds >= 0, "--sample-timeout-seconds must be greater than or equal to 0")
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
                **(
                    {
                        "model_id": infer_model_id_from_path(args.model_dir),
                        "role": "student_base",
                    }
                    if provider_id == "local_transformers"
                    else {}
                ),
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
    if local_runtime is not None:
        print(
            f"[runtime-ready] provider=local_transformers device={local_runtime.device} "
            f"model_id={provider.get('model_id') or provider_id}",
            flush=True,
        )

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
    generation_metric_entries: list[dict[str, Any]] = []
    repaired_output_count = 0
    repaired_path_counts: dict[str, int] = {}
    selected_samples = iter_selected_samples(source_eval_manifest, sample_id_filter=args.sample_id)
    print(
        f"[batch-start] provider={provider_id} sample_count={len(selected_samples)} "
        f"output_dir={repo_rel(output_dir)}",
        flush=True,
    )
    for sample_index, (sample, selected_sample) in enumerate(selected_samples, start=1):
        jsonschema.validate(sample["input_request"], request_schema)
        task_key = f"{sample['project']}/{sample['task']}"
        task_counts[task_key] = task_counts.get(task_key, 0) + 1
        prompt_document = build_prompt_document(sample, model_id=str(provider.get("model_id") or provider_id))
        sample_id = str(sample["sample_id"])
        prompt_relpath = Path("prompts") / f"{sample_id}.prompt.json"
        response_relpath = Path("responses") / f"{sample_id}.candidate-response.json"
        invalid_response_relpath = Path("invalid-responses") / f"{sample_id}.candidate-response.invalid.json"
        write_json(prompt_dir / prompt_relpath.name, prompt_document)

        response_valid = True
        validation_error: str | None = None
        generation_metrics: dict[str, Any] = {}
        repaired_paths: list[str] = []
        sample_started_at = time.perf_counter()
        print(
            f"[sample-start {sample_index}/{len(selected_samples)}] {sample_id} "
            f"task={task_key} provider={provider_id} timeout={args.sample_timeout_seconds:g}s "
            f"max_new_tokens={args.max_new_tokens}",
            flush=True,
        )
        try:
            candidate_result = build_candidate_response(
                sample,
                provider_id=provider_id,
                prompt_document=prompt_document,
                local_runtime=local_runtime,
                max_new_tokens=args.max_new_tokens,
                sample_timeout_seconds=args.sample_timeout_seconds if provider_id == "local_transformers" else 0.0,
            )
            candidate_response = candidate_result.response
            response_source = candidate_result.response_source
            generation_metrics = candidate_result.generation_metrics
        except (Exception, SystemExit) as exc:
            if not args.allow_invalid_output:
                raise
            response_valid = False
            response_source = provider_id
            validation_error = str(exc) or "candidate response generation failed"
            if isinstance(exc, LocalTransformersGenerationError):
                generation_metrics = exc.generation_metrics
            candidate_response = {
                "kind": "invalid_candidate_response",
                "sample_id": sample_id,
                "provider_id": provider_id,
                "error": validation_error,
            }
            add_count(schema_failure_categories, categorize_failure(validation_error))
        if response_valid and args.repair_hard_fields:
            candidate_response, repaired_paths = repair_candidate_hard_fields(candidate_response, sample=sample)
            if repaired_paths:
                repaired_output_count += 1
                for path in repaired_paths:
                    add_count(repaired_path_counts, path)
        if response_valid:
            try:
                validate_candidate_response(candidate_response, sample=sample, response_schema=response_schema)
            except Exception as exc:
                if not args.allow_invalid_output:
                    raise
                response_valid = False
                validation_error = summarize_validation_error(exc)
                add_count(schema_failure_categories, categorize_failure(validation_error))
        if generation_metrics:
            generation_metric_entries.append(generation_metrics)

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
        sample_elapsed_seconds = round(time.perf_counter() - sample_started_at, 3)
        print(
            f"[sample-done {sample_index}/{len(selected_samples)}] {sample_id} "
            f"schema_valid={response_valid} "
            f"task_valid={task_response_valid if task_response_valid is not None else 'not_checked'} "
            f"elapsed={sample_elapsed_seconds}s",
            flush=True,
        )
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
                **({"generation_metrics": generation_metrics} if generation_metrics else {}),
                **(
                    {
                        "postprocess": {
                            "repair_hard_fields": True,
                            "repaired_paths": repaired_paths,
                        }
                    }
                    if repaired_paths
                    else {}
                ),
                **({"task_response_validated": task_response_valid} if args.validate_task and response_valid else {}),
                **({"task_validation_violations": task_validation_violations} if task_validation_violations else {}),
            }
        )

    sample_count = len(outputs)
    all_valid = invalid_response_count == 0
    task_validation_attempted = task_validated_count + task_invalid_count
    total_generation_seconds = round(sum(float(entry.get("generation_seconds") or 0.0) for entry in generation_metric_entries), 3)
    total_output_tokens = sum(int(entry.get("output_tokens") or 0) for entry in generation_metric_entries)
    total_input_tokens = sum(int(entry.get("input_tokens") or 0) for entry in generation_metric_entries)
    generation_summary = {
        "samples_measured": len(generation_metric_entries),
        "total_generation_seconds": total_generation_seconds,
        "avg_generation_seconds": round(total_generation_seconds / len(generation_metric_entries), 3)
        if generation_metric_entries
        else 0.0,
        "total_input_tokens": total_input_tokens,
        "total_output_tokens": total_output_tokens,
        "avg_output_tokens": round(total_output_tokens / len(generation_metric_entries), 3) if generation_metric_entries else 0.0,
        "hit_max_new_tokens_count": sum(1 for entry in generation_metric_entries if entry.get("hit_max_new_tokens") is True),
        "json_extracted_count": sum(1 for entry in generation_metric_entries if entry.get("json_extracted") is True),
        "timeout_count": sum(1 for entry in generation_metric_entries if entry.get("timeout") is True),
    }
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
        **(
            {
                "postprocess_policy": {
                    "repair_hard_fields": True,
                    "scope": (
                        "Experimental: repairs scaffold-derived hard fields before schema/task validation; "
                        "raw model capability must still be assessed with this flag disabled."
                    ),
                    "repaired_output_count": repaired_output_count,
                    "repaired_path_counts": dict(sorted(repaired_path_counts.items())),
                }
            }
            if args.repair_hard_fields
            else {}
        ),
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
            **({"generation_summary": generation_summary} if generation_metric_entries else {}),
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

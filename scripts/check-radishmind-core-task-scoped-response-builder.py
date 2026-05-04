#!/usr/bin/env python3
from __future__ import annotations

import importlib.util
import json
import sys
from pathlib import Path
from typing import Any

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))


GHOST_AMBIGUOUS_SAMPLE = REPO_ROOT / "datasets/eval/radishflow/suggest-ghost-completion-valve-ambiguous-no-tab-001.json"
GHOST_VAPOR_SAMPLE = REPO_ROOT / "datasets/eval/radishflow/suggest-ghost-completion-flash-vapor-outlet-001.json"
DOCS_EVIDENCE_GAP_SAMPLE = REPO_ROOT / "datasets/eval/radish/answer-docs-question-evidence-gap-001.json"
DOCS_SOURCE_CONFLICT_SAMPLE = REPO_ROOT / "datasets/eval/radish/answer-docs-question-docs-faq-forum-conflict-001.json"
RESPONSE_SCHEMA = REPO_ROOT / "contracts/copilot-response.schema.json"


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must be a JSON object")
    return document


def load_candidate_runner() -> Any:
    module_path = REPO_ROOT / "scripts/run-radishmind-core-candidate.py"
    spec = importlib.util.spec_from_file_location("radishmind_core_candidate_runner", module_path)
    require(spec is not None and spec.loader is not None, "candidate runner module spec must be loadable")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def assert_valid_response(runner: Any, sample: dict[str, Any], response: dict[str, Any]) -> None:
    response_schema = load_json(RESPONSE_SCHEMA)
    jsonschema.validate(response, response_schema)
    runner.validate_candidate_response(response, sample=sample, response_schema=response_schema)
    violations = runner.validate_task_candidate_response(
        response,
        sample=sample,
        sample_name=f"{sample['sample_id']}.json",
    )
    require(not violations, "task-scoped builder output must pass task validation: " + "; ".join(violations))


def main() -> int:
    runner = load_candidate_runner()

    ghost_sample = load_json(GHOST_AMBIGUOUS_SAMPLE)
    raw_ghost_response = {
        "schema_version": 1,
        "status": "ok",
        "project": "radishflow",
        "task": "suggest_ghost_completion",
        "summary": "raw model ghost summary",
        "answers": [
            {
                "kind": "ghost_rationale",
                "text": "raw model ghost rationale",
                "citation_ids": [],
            }
        ],
        "issues": [],
        "proposed_actions": [
            {
                "kind": "ghost_completion",
                "title": "raw cooler title",
                "target": {"type": "unit_port", "unit_id": "wrong", "port_key": "outlet"},
                "rationale": "raw cooler rationale",
                "patch": {},
                "preview": {"accept_key": "Tab"},
                "apply": {"command_kind": "accept_ghost_completion", "payload": {}},
                "risk_level": "low",
                "requires_confirmation": False,
                "citation_ids": [],
            },
            {
                "kind": "ghost_completion",
                "title": "raw flash title",
                "target": {"type": "unit_port", "unit_id": "wrong", "port_key": "outlet"},
                "rationale": "raw flash rationale",
                "patch": {},
                "preview": {"accept_key": "Tab"},
                "apply": {"command_kind": "accept_ghost_completion", "payload": {}},
                "risk_level": "low",
                "requires_confirmation": False,
                "citation_ids": [],
            },
        ],
        "citations": [],
        "confidence": 0.37,
        "risk_level": "low",
        "requires_confirmation": False,
    }
    built_ghost, ghost_paths = runner.build_task_scoped_response(raw_ghost_response, sample=ghost_sample)
    assert_valid_response(runner, ghost_sample, built_ghost)
    require(built_ghost["summary"] == "raw model ghost summary", "builder must preserve ghost summary")
    require(built_ghost["answers"][0]["text"] == "raw model ghost rationale", "builder must preserve ghost answer text")
    require(
        [action.get("title") for action in built_ghost["proposed_actions"]] == ["raw cooler title", "raw flash title"],
        "builder must preserve ghost action titles",
    )
    require(
        [action.get("patch", {}).get("candidate_ref") for action in built_ghost["proposed_actions"]]
        == ["cand-valve-1-outlet-cooler", "cand-valve-1-outlet-flash"],
        "builder must rebuild ordered ghost candidate_ref values",
    )
    require(
        [action.get("patch", {}).get("ghost_kind") for action in built_ghost["proposed_actions"]]
        == ["ghost_connection", "ghost_connection"],
        "builder must rebuild ghost_kind values",
    )
    require(
        [action.get("preview", {}).get("accept_key") for action in built_ghost["proposed_actions"]]
        == ["manual_only", "manual_only"],
        "builder must preserve manual-only ambiguous ghost boundary",
    )
    require("$.proposed_actions" in ghost_paths, "builder must report rebuilt ghost action scaffold")

    ghost_vapor_sample = load_json(GHOST_VAPOR_SAMPLE)
    raw_ghost_vapor_response = {
        "schema_version": 1,
        "status": "ok",
        "project": "radishflow",
        "task": "suggest_ghost_completion",
        "summary": "根据法律候选完成体选择合法的 ghost 完成体，以供用户预览并应用。",
        "answers": [
            {
                "kind": "ghost_rationale",
                "text": "在 legal_candidate_completions 中选择一个符合法规要求的 ghost 完成体。",
                "citation_ids": [],
            }
        ],
        "issues": [],
        "proposed_actions": [
            {
                "kind": "ghost_completion",
                "title": "生成合法 ghost completion 候选",
                "target": {"type": "unit_port", "unit_id": "flash-2", "port_key": "vapor_outlet"},
                "rationale": "确保所选 ghost 完成体符合法规要求，避免潜在风险。",
                "patch": {},
                "preview": {"accept_key": "Tab"},
                "apply": {"command_kind": "accept_ghost_completion", "payload": {}},
                "risk_level": "low",
                "requires_confirmation": False,
                "citation_ids": [],
            }
        ],
        "citations": [],
        "confidence": 0.72,
        "risk_level": "low",
        "requires_confirmation": False,
    }
    built_ghost_vapor, ghost_vapor_paths = runner.build_task_scoped_response(
        raw_ghost_vapor_response,
        sample=ghost_vapor_sample,
    )
    assert_valid_response(runner, ghost_vapor_sample, built_ghost_vapor)
    guarded_text = json.dumps(built_ghost_vapor, ensure_ascii=False)
    for forbidden_term in ("法律", "法规", "法規", "法定", "合法", "合规", "合規"):
        require(forbidden_term not in guarded_text, f"builder must reject ghost mistranslation term: {forbidden_term}")
    require("$.summary" not in ghost_vapor_paths, "builder must not report rejected ghost summary as merged")
    require("$.answers[0].text" not in ghost_vapor_paths, "builder must not report rejected ghost answer as merged")
    require(
        "$.proposed_actions[0].rationale" not in ghost_vapor_paths,
        "builder must not report rejected ghost action rationale as merged",
    )
    require(
        built_ghost_vapor["proposed_actions"][0]["patch"]["candidate_ref"] == "cand-flash-2-vapor-stub",
        "builder must still rebuild the leading vapor candidate_ref",
    )

    docs_sample = load_json(DOCS_EVIDENCE_GAP_SAMPLE)
    raw_docs_response = {
        "schema_version": 1,
        "status": "ok",
        "project": "radish",
        "task": "answer_docs_question",
        "summary": "raw model docs summary",
        "answers": [
            {
                "kind": "direct_answer",
                "text": "raw model docs answer",
                "citation_ids": [],
            }
        ],
        "issues": [],
        "proposed_actions": [
            {
                "kind": "read_only_check",
                "title": "raw action should be removed",
                "rationale": "raw action rationale",
                "risk_level": "low",
                "requires_confirmation": False,
                "citation_ids": [],
            }
        ],
        "citations": [],
        "confidence": 0.61,
        "risk_level": "low",
        "requires_confirmation": False,
    }
    built_docs, docs_paths = runner.build_task_scoped_response(raw_docs_response, sample=docs_sample)
    assert_valid_response(runner, docs_sample, built_docs)
    require(built_docs["status"] == "partial", "builder must rebuild docs evidence-gap status")
    require(built_docs["risk_level"] == "medium", "builder must rebuild docs evidence-gap risk level")
    require(built_docs["proposed_actions"] == [], "builder must remove disallowed docs actions")
    require(built_docs["summary"] == "raw model docs summary", "builder must preserve docs summary")
    require(built_docs["answers"][0]["text"] == "raw model docs answer", "builder must preserve docs answer text")
    require(built_docs["issues"][0]["code"] == "INSUFFICIENT_EVIDENCE", "builder must rebuild docs evidence issue")
    require(built_docs["answers"][0]["citation_ids"] == ["doc-1"], "builder must rebuild docs answer citation")
    require(built_docs["issues"][0]["citation_ids"] == ["doc-1"], "builder must rebuild docs issue citation")
    require("$.status" in docs_paths and "$.risk_level" in docs_paths, "builder must report docs status/risk paths")

    docs_source_conflict_sample = load_json(DOCS_SOURCE_CONFLICT_SAMPLE)
    raw_docs_source_conflict_response = {
        "schema_version": 1,
        "status": "ok",
        "project": "radish",
        "task": "answer_docs_question",
        "summary": "正式 docs 仍要求优先保持 slug 稳定。",
        "answers": [
            {
                "kind": "direct_answer",
                "text": "给出可展示给用户的回答。",
                "citation_ids": [],
            }
        ],
        "issues": [],
        "proposed_actions": [],
        "citations": [],
        "confidence": 0.58,
        "risk_level": "low",
        "requires_confirmation": False,
    }
    built_docs_source_conflict, docs_source_conflict_paths = runner.build_task_scoped_response(
        raw_docs_source_conflict_response,
        sample=docs_source_conflict_sample,
    )
    assert_valid_response(runner, docs_source_conflict_sample, built_docs_source_conflict)
    require(
        built_docs_source_conflict["summary"] == "正式 docs 仍要求优先保持 slug 稳定。",
        "builder must still preserve acceptable docs summary",
    )
    require(
        "给出可展示给用户的回答" not in built_docs_source_conflict["answers"][0]["text"],
        "builder must reject generic docs answer placeholder",
    )
    require(
        all(term in built_docs_source_conflict["answers"][0]["text"] for term in ("docs", "FAQ", "forum")),
        "builder docs answer fallback must be task-aware and displayable",
    )
    require("$.summary" in docs_source_conflict_paths, "builder must report accepted docs summary as merged")
    require(
        "$.answers[0].text" not in docs_source_conflict_paths,
        "builder must not report rejected docs answer placeholder as merged",
    )

    print("radishmind core task-scoped response builder check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

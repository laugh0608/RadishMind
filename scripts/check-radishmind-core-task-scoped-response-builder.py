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
    require(
        built_ghost["summary"] == "valve-1 的 outlet 当前存在两个接近的可行下游，因此可以显示两个 ghost 备选，但不应默认把其中任何一个绑定到 Tab。",
        "builder must keep ambiguous ghost summary task-grounded instead of preserving misleading raw summary",
    )
    require(
        built_ghost["answers"][0]["text"] == "cooler-401 和 flash-401 都是合法下游，但距离与对齐评分非常接近，当前证据不足以让某一条候选显著领先，因此只返回可见 ghost 而不默认接受。",
        "builder must keep ambiguous ghost answer task-grounded instead of preserving misleading raw answer",
    )
    require(
        [action.get("title") for action in built_ghost["proposed_actions"]]
        == [
            "将 cooler-401 作为 valve-1 outlet 的可见 ghost 候选",
            "将 flash-401 作为 valve-1 outlet 的可见 ghost 候选",
        ],
        "builder must rewrite ambiguous ghost action titles to downstream-grounded wording",
    )
    require(
        [action.get("rationale") for action in built_ghost["proposed_actions"]]
        == [
            "通往 cooler-401 的路径更靠前，但领先幅度不足以成为默认 Tab 建议，因此只保留为手动可选 ghost。",
            "通往 flash-401 的候选同样合法，且与 cooler-401 差距不足以被直接排除，因此作为第二条可见 ghost 保留。",
        ],
        "builder must rewrite ambiguous ghost action rationales to no-Tab semantics",
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
    require(
        [action.get("preview", {}).get("render_priority") for action in built_ghost["proposed_actions"]] == [1, 2],
        "builder must keep ambiguous ghost render priorities ordered by action index",
    )
    require("$.proposed_actions" in ghost_paths, "builder must report rebuilt ghost action scaffold")
    require("$.summary" not in ghost_paths, "builder must reject misleading ambiguous ghost summary merge")
    require("$.answers[0].text" not in ghost_paths, "builder must reject misleading ambiguous ghost answer merge")
    require(
        "$.proposed_actions[0].title" not in ghost_paths and "$.proposed_actions[1].title" not in ghost_paths,
        "builder must reject generic ambiguous ghost action titles",
    )

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
    require(
        built_docs["summary"] == "当前文档只说明可以查看 Hangfire 任务状态，不能据此确认谁有权重试或删除任务。",
        "builder must keep docs evidence-gap summary grounded in the primary artifact",
    )
    require(
        built_docs["answers"][0]["kind"] == "evidence_limited_answer",
        "builder must keep docs evidence-gap answer kind evidence_limited_answer",
    )
    require(
        built_docs["answers"][0]["text"] == "现有文档只覆盖查看能力，没有给出重试或删除任务的授权口径，因此不能直接下权限结论。",
        "builder must keep docs evidence-gap answer grounded in the missing authorization evidence",
    )
    require(built_docs["issues"][0]["code"] == "INSUFFICIENT_EVIDENCE", "builder must rebuild docs evidence issue")
    require(
        built_docs["issues"][0]["message"] == "当前证据不足以确认是否允许重试或删除 Hangfire 任务。",
        "builder must keep docs evidence-gap issue grounded in the missing authorization evidence",
    )
    require(built_docs["answers"][0]["citation_ids"] == ["doc-1"], "builder must rebuild docs answer citation")
    require(built_docs["issues"][0]["citation_ids"] == ["doc-1"], "builder must rebuild docs issue citation")
    require("$.status" in docs_paths and "$.risk_level" in docs_paths, "builder must report docs status/risk paths")
    require("$.summary" not in docs_paths, "builder must reject instructional docs summary merge")
    require("$.answers[0].text" not in docs_paths, "builder must reject instructional docs answer merge")

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

    cross_object_sample = load_json(REPO_ROOT / "datasets/eval/radishflow/suggest-flowsheet-edits-cross-object-mixed-risk-reconnect-plus-pump-parameter-001.json")
    raw_cross_object_response = {
        "schema_version": 1,
        "status": "partial",
        "project": "radishflow",
        "task": "suggest_flowsheet_edits",
        "summary": "建议为 cooler-outlet-81 生成 advisory-only flowsheet 编辑候选，并保持人工确认边界。",
        "answers": [
            {
                "kind": "edit_rationale",
                "text": "这条样本把高风险断链 patch 和跨对象 pump parameter_updates 放进同一响应，用来冻结 reconnect、parameter_updates、warning-only unit context 与顶层 citations 的组合稳定性。",
                "citation_ids": [],
            }
        ],
        "issues": [
            {
                "code": "STREAM_DISCONNECTED",
                "message": "这条样本把高风险断链 patch 和跨对象 pump parameter_updates 放进同一响应，用来冻结 reconnect、parameter_updates、warning-only unit context 与顶层 citations 的组合稳定性。",
                "severity": "error",
                "citation_ids": [],
            },
            {
                "code": "PUMP_OUTLET_PRESSURE_TARGET_INVALID",
                "message": "这条样本把高风险断链 patch 和跨对象 pump parameter_updates 放进同一响应，用来冻结 reconnect、parameter_updates、warning-only unit context 与顶层 citations 的组合稳定性。",
                "severity": "warning",
                "citation_ids": [],
            },
        ],
        "proposed_actions": [
            {
                "kind": "candidate_edit",
                "title": "生成待人工确认的 flowsheet 编辑提案",
                "target": {"type": "stream", "id": "cooler-outlet-81"},
                "rationale": "Trim Cooler Outlet is not connected to any downstream consumer or export sink.",
                "patch": {},
                "risk_level": "high",
                "requires_confirmation": True,
                "citation_ids": [],
            },
            {
                "kind": "candidate_edit",
                "title": "生成待人工确认的 flowsheet 编辑提案",
                "target": {"type": "unit", "id": "pump-12"},
                "rationale": "Recycle Pump outlet pressure target is lower than the inlet pressure.",
                "patch": {},
                "risk_level": "high",
                "requires_confirmation": True,
                "citation_ids": [],
            },
        ],
        "citations": [],
        "confidence": 0.8,
        "risk_level": "high",
        "requires_confirmation": True,
    }
    built_cross_object, cross_object_paths = runner.build_task_scoped_response(
        raw_cross_object_response,
        sample=cross_object_sample,
    )
    assert_valid_response(runner, cross_object_sample, built_cross_object)
    require(
        built_cross_object["summary"] == "当前更合适的是把 cooler-outlet-81 的重连占位与 pump-12 的局部参数复核拆成两条 candidate_edit，并把 cooler-12 继续保留在 warning 与解释层。",
        "builder must keep cross-object summary task-grounded instead of preserving single-target fallback",
    )
    require(
        built_cross_object["answers"][0]["text"] == "这组提案同时混合了高风险连接修复与局部参数复核，因此顶层说明应先列直接诊断，再列两个可行动对象，最后再保留 supporting context 与 snapshot。",
        "builder must keep cross-object answer task-grounded instead of preserving eval-note text",
    )
    require(
        [issue.get("message") for issue in built_cross_object["issues"]]
        == [
            "cooler-outlet-81 当前没有任何 downstream consumer 或 export sink 绑定，需要先停留在待确认的重连占位层。",
            "pump-12 的 outlet_pressure_target_kpa 低于入口压力，当前参数设置与泵升压方向不一致。",
            "pump-12 的 efficiency_percent=108 超出了当前建议运行区间。",
            "cooler-12 的出口状态仍可能在恢复 downstream sink 后变化，因此当前不应给 cooler 侧额外生成第二条 unit patch。",
        ],
        "builder must rewrite cross-object issue messages from diagnostics instead of eval-note text",
    )
    require(
        [issue.get("severity") for issue in built_cross_object["issues"]] == ["error", "error", "warning", "warning"],
        "builder must preserve cross-object diagnostic severities",
    )
    require(
        [action.get("title") for action in built_cross_object["proposed_actions"]]
        == ["为 cooler-outlet-81 预留下游重连占位", "复核 pump-12 的目标压力与效率范围"],
        "builder must rewrite cross-object action titles to task-grounded wording",
    )
    require(
        [action.get("risk_level") for action in built_cross_object["proposed_actions"]] == ["high", "medium"],
        "builder must narrow the pump parameter patch risk back to medium",
    )
    require("$.summary" not in cross_object_paths, "builder must reject single-target cross-object summary merge")
    require("$.answers[0].text" not in cross_object_paths, "builder must reject eval-note cross-object answer merge")
    require(
        "$.issues[0].message" not in cross_object_paths and "$.issues[1].message" not in cross_object_paths,
        "builder must reject eval-note cross-object issue merges",
    )

    print("radishmind core task-scoped response builder check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

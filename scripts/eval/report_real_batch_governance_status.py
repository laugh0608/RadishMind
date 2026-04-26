#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import sys
from datetime import date
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parent.parent.parent
SCRIPT_DIR = Path(__file__).resolve().parent

if str(SCRIPT_DIR) not in sys.path:
    sys.path.insert(0, str(SCRIPT_DIR))

from report_suggest_edits_profile_coverage import build_report as build_suggest_edits_profile_coverage  # noqa: E402
from radishflow_batch_artifact_summary import (  # noqa: E402
    load_radishflow_batch_artifact_summary,
)
from radishflow_ghost_sample_groups import (  # noqa: E402
    HIGH_VALUE_PRIORITY_GROUPS,
    build_pending_group_summaries,
)

FIXTURE_ROOT = REPO_ROOT / "scripts/checks/fixtures"
SUGGEST_EDITS_BATCH_FIXTURE = FIXTURE_ROOT / "radishflow-suggest-edits-poc-batches.json"
GHOST_BATCH_FIXTURE = FIXTURE_ROOT / "radishflow-ghost-real-batches.json"
RADISH_DOCS_BATCH_FIXTURE = FIXTURE_ROOT / "radish-docs-real-batches.json"
RADISH_DOCS_REAL_DERIVED_FIXTURE = FIXTURE_ROOT / "radish-docs-real-derived-negatives.json"
RADISHFLOW_GHOST_REAL_DERIVED_FIXTURE = FIXTURE_ROOT / "radishflow-ghost-real-derived-negatives.json"
RADISHFLOW_SUGGEST_EDITS_REAL_DERIVED_FIXTURE = FIXTURE_ROOT / "radishflow-suggest-edits-real-derived-negatives.json"
MISSING_NEGATIVE_SAMPLES = "missing_negative_samples"
MISSING_REAL_DERIVED_NEGATIVE_SAMPLES = "missing_real_derived_negative_samples"
NEXT_SUGGEST_EDITS_HIGH_VALUE_GROUP = ""


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Summarize the current real-batch governance status across "
            "suggest_flowsheet_edits, suggest_ghost_completion, and Radish docs QA."
        ),
    )
    parser.add_argument(
        "--json-output",
        default="",
        help="Optional path to write the structured governance report JSON.",
    )
    return parser.parse_args()


def resolve_repo_relative(path_value: str) -> Path:
    path = Path(path_value)
    if not path.is_absolute():
        path = (REPO_ROOT / path).resolve()
    return path


def make_repo_relative(path: Path) -> str:
    resolved = path.resolve()
    if resolved.is_relative_to(REPO_ROOT):
        return str(resolved.relative_to(REPO_ROOT)).replace("\\", "/")
    return str(resolved)


def load_json_document(path: Path) -> dict[str, Any]:
    try:
        document = json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json '{make_repo_relative(path)}': {exc}") from exc
    if not isinstance(document, dict):
        raise SystemExit(f"expected json object: {make_repo_relative(path)}")
    return document


def load_json_list(path: Path) -> list[dict[str, Any]]:
    try:
        document = json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json '{make_repo_relative(path)}': {exc}") from exc
    if not isinstance(document, list):
        raise SystemExit(f"expected json array: {make_repo_relative(path)}")
    normalized: list[dict[str, Any]] = []
    for index, item in enumerate(document):
        if not isinstance(item, dict):
            raise SystemExit(f"expected object at index {index}: {make_repo_relative(path)}")
        normalized.append(item)
    return normalized


def load_sample_id(path: Path) -> str:
    document = load_json_document(path)
    sample_id = str(document.get("sample_id") or "").strip()
    if not sample_id:
        raise SystemExit(f"sample is missing sample_id: {make_repo_relative(path)}")
    return sample_id


def safe_bool(value: Any) -> bool:
    return bool(value)


def safe_ratio(numerator: int, denominator: int) -> float:
    if denominator <= 0:
        return 0.0
    return round(numerator / denominator, 4)


def collect_unique_sample_ids(batches: list[dict[str, Any]]) -> list[str]:
    return sorted({sample_id for batch in batches for sample_id in batch.get("sample_ids") or []})


def build_coverage_summary(
    *,
    total_eval_sample_count: int,
    real_captured_sample_count: int,
    latest_batch_matched_sample_count: int,
    latest_batch_record_count: int,
) -> dict[str, Any]:
    return {
        "total_eval_sample_count": total_eval_sample_count,
        "real_captured_sample_count": real_captured_sample_count,
        "real_captured_ratio": safe_ratio(real_captured_sample_count, total_eval_sample_count),
        "latest_batch_matched_sample_count": latest_batch_matched_sample_count,
        "latest_batch_matched_ratio": safe_ratio(latest_batch_matched_sample_count, total_eval_sample_count),
        "latest_batch_record_count": latest_batch_record_count,
    }


def load_optional_radishflow_artifact_summary(entry: dict[str, Any]) -> tuple[str, dict[str, Any] | None]:
    summary_path_value = str(entry.get("artifact_summary") or "").strip()
    if not summary_path_value:
        return "", None
    summary_path = resolve_repo_relative(summary_path_value)
    if not summary_path.is_file():
        raise SystemExit(f"artifact summary not found: {make_repo_relative(summary_path)}")
    return make_repo_relative(summary_path), load_radishflow_batch_artifact_summary(summary_path)


def find_latest_governance_entry(
    entries: list[dict[str, Any]],
    *,
    preferred_keys: tuple[str, ...] = (),
) -> dict[str, Any]:
    for entry in reversed(entries):
        if not str(entry.get("artifact_summary") or "").strip():
            continue
        if preferred_keys and not any(str(entry.get(key) or "").strip() for key in preferred_keys):
            continue
        return entry
    for entry in reversed(entries):
        if str(entry.get("artifact_summary") or "").strip():
            return entry
    return entries[-1]


def summarize_batch(entry: dict[str, Any]) -> dict[str, Any]:
    record_dir = resolve_repo_relative(str(entry.get("record_dir") or "").strip())
    manifest_path = resolve_repo_relative(str(entry.get("manifest") or "").strip())
    audit_path = resolve_repo_relative(str(entry.get("audit_report") or "").strip())
    if not manifest_path.is_file():
        raise SystemExit(f"manifest not found: {make_repo_relative(manifest_path)}")
    if not audit_path.is_file():
        raise SystemExit(f"audit report not found: {make_repo_relative(audit_path)}")
    manifest = load_json_document(manifest_path)
    audit = load_json_document(audit_path)
    records = manifest.get("records") or []
    if not isinstance(records, list):
        raise SystemExit(f"manifest records must be an array: {make_repo_relative(manifest_path)}")
    sample_ids: list[str] = []
    for record in records:
        if not isinstance(record, dict):
            raise SystemExit(f"manifest record entries must be objects: {make_repo_relative(manifest_path)}")
        sample_id = str(record.get("sample_id") or "").strip()
        if sample_id:
            sample_ids.append(sample_id)
    matched_sample_count = int(audit.get("matched_sample_count") or 0)
    passed_count = int(audit.get("passed_count") or 0)
    failed_count = int(audit.get("failed_count") or 0)
    violation_count = int(audit.get("violation_count") or 0)
    exit_code = int(audit.get("exit_code") or 0)
    audit_clean = exit_code == 0 and failed_count == 0 and violation_count == 0
    return {
        "record_dir": make_repo_relative(record_dir),
        "manifest_path": make_repo_relative(manifest_path),
        "audit_report_path": make_repo_relative(audit_path),
        "collection_batch": str(manifest.get("collection_batch") or "").strip(),
        "source": str(manifest.get("source") or "").strip(),
        "record_count": len(records),
        "sample_ids": sample_ids,
        "matched_sample_count": matched_sample_count,
        "passed_count": passed_count,
        "failed_count": failed_count,
        "violation_count": violation_count,
        "audit_exit_code": exit_code,
        "audit_clean": audit_clean,
    }


def load_real_batches(fixture_path: Path) -> list[dict[str, Any]]:
    batches: list[dict[str, Any]] = []
    for entry in load_json_list(fixture_path):
        summary = summarize_batch(entry)
        if summary["source"] != "captured_candidate_response":
            continue
        batches.append(summary)
    if not batches:
        raise SystemExit(f"no real batches found in fixture: {make_repo_relative(fixture_path)}")
    return batches


def count_eval_samples(pattern: str) -> int:
    return len(list(REPO_ROOT.glob(pattern)))


def make_governance_blockers(
    *,
    negative_replay_index_blocker: str = "",
    cross_sample_negative_replay_index_blocker: str = "",
    recommended_negative_replay_summary_blocker: str = "",
    cross_sample_recommended_negative_replay_summary_blocker: str = "",
    real_derived_negative_index_blocker: str = "",
) -> dict[str, str]:
    return {
        "negative_replay_index_blocker": negative_replay_index_blocker,
        "cross_sample_negative_replay_index_blocker": cross_sample_negative_replay_index_blocker,
        "recommended_negative_replay_summary_blocker": recommended_negative_replay_summary_blocker,
        "cross_sample_recommended_negative_replay_summary_blocker": cross_sample_recommended_negative_replay_summary_blocker,
        "real_derived_negative_index_blocker": real_derived_negative_index_blocker,
    }


def build_suggest_edits_chain() -> dict[str, Any]:
    fixture_entries = load_json_list(SUGGEST_EDITS_BATCH_FIXTURE)
    batches = load_real_batches(SUGGEST_EDITS_BATCH_FIXTURE)
    latest_batch = batches[-1]
    governance_entry = find_latest_governance_entry(
        fixture_entries,
        preferred_keys=(
            "negative_replay_index",
            "cross_sample_replay_index",
            "recommended_summary",
            "cross_sample_recommended_summary",
        ),
    )
    artifact_summary_path, artifact_summary = load_optional_radishflow_artifact_summary(governance_entry)
    real_derived_fixture = load_json_document(RADISHFLOW_SUGGEST_EDITS_REAL_DERIVED_FIXTURE)
    real_derived_index_path = resolve_repo_relative(str(real_derived_fixture.get("index") or "").strip())
    real_derived_index = load_json_document(real_derived_index_path) if real_derived_index_path.is_file() else None
    real_derived_summary: dict[str, Any] = {}
    if real_derived_index is not None:
        real_derived_summary = real_derived_index.get("summary") or {}
        if not isinstance(real_derived_summary, dict):
            raise SystemExit(f"real-derived summary must be an object: {make_repo_relative(real_derived_index_path)}")
    coverage_report = build_suggest_edits_profile_coverage()
    total_eval_sample_count = int(coverage_report.get("sample_count") or 0)
    real_captured_sample_ids = collect_unique_sample_ids(batches)
    teacher_candidates = list(coverage_report.get("teacher_comparison_candidates") or [])
    next_group = ""
    if teacher_candidates:
        next_group = str(teacher_candidates[0].get("group_name") or "").strip()
    if next_group:
        next_gap = (
            "same-sample / cross-sample replay 与首批 real-derived negative 已接通；"
            f"下一步应回到 suggest_flowsheet_edits 的 {next_group} default teacher capture。"
        )
    elif NEXT_SUGGEST_EDITS_HIGH_VALUE_GROUP:
        next_gap = (
            "当前四主 apiyi coverage 与 replay / real-derived 治理资产均已接通；"
            "既有 suggest_flowsheet_edits 高价值真实扩样入口也已跑通；"
            f"下一步应优先用 {NEXT_SUGGEST_EDITS_HIGH_VALUE_GROUP} 启动新的非重复高价值真实 capture。"
        )
    else:
        next_gap = (
            "当前四主 apiyi coverage 与 replay / real-derived 治理资产均已接通；"
            "既有 suggest_flowsheet_edits 高价值真实扩样入口也已跑通；"
            "下一步应优先定义新的非重复高价值样本组，或转向 ghost 链新增非重复高价值样本设计。"
        )
    governance: dict[str, Any] = {
        "level": (
            "artifact_summary_replay_and_real_derived"
            if real_derived_index is not None
            else
            "manifest_audit_profile_coverage_and_artifact_summary"
            if artifact_summary is not None
            else "manifest_audit_and_profile_coverage"
        ),
        "artifact_summary": artifact_summary is not None,
        "negative_replay_index": False,
        "cross_sample_negative_replay_index": False,
        "recommended_negative_replay_summary": False,
        "cross_sample_recommended_negative_replay_summary": False,
        "real_derived_negative_index": real_derived_index is not None,
        "real_derived_pattern_group_count": int(real_derived_summary.get("pattern_group_count") or 0),
        "real_derived_violation_group_count": int(real_derived_summary.get("violation_group_count") or 0),
        **make_governance_blockers(
            negative_replay_index_blocker=MISSING_NEGATIVE_SAMPLES,
            cross_sample_negative_replay_index_blocker=MISSING_NEGATIVE_SAMPLES,
            recommended_negative_replay_summary_blocker=MISSING_NEGATIVE_SAMPLES,
            cross_sample_recommended_negative_replay_summary_blocker=MISSING_NEGATIVE_SAMPLES,
            real_derived_negative_index_blocker=(
                "" if real_derived_index is not None else MISSING_REAL_DERIVED_NEGATIVE_SAMPLES
            ),
        ),
    }
    if artifact_summary is not None:
        artifacts = artifact_summary.get("artifacts") or {}
        if not isinstance(artifacts, dict):
            raise SystemExit(f"artifacts must be an object: {artifact_summary_path}")

        def artifact_exists(key: str) -> bool:
            value = artifacts.get(key)
            return isinstance(value, dict) and safe_bool(value.get("exists"))

        governance["negative_replay_index"] = artifact_exists("negative_replay_index")
        governance["cross_sample_negative_replay_index"] = artifact_exists("cross_sample_negative_replay_index")
        governance["recommended_negative_replay_summary"] = artifact_exists("recommended_negative_replay_summary")
        governance["cross_sample_recommended_negative_replay_summary"] = artifact_exists(
            "cross_sample_recommended_negative_replay_summary"
        )
        if governance["negative_replay_index"]:
            governance["negative_replay_index_blocker"] = ""
        if governance["cross_sample_negative_replay_index"]:
            governance["cross_sample_negative_replay_index_blocker"] = ""
        if governance["recommended_negative_replay_summary"] or artifact_exists("negative_replay_index"):
            governance["recommended_negative_replay_summary_blocker"] = ""
        if governance["cross_sample_recommended_negative_replay_summary"] or artifact_exists(
            "cross_sample_negative_replay_index"
        ):
            governance["cross_sample_recommended_negative_replay_summary_blocker"] = ""
    priority_category = "expand_real_capture_pool"
    if NEXT_SUGGEST_EDITS_HIGH_VALUE_GROUP:
        priority_recommendation = (
            f"继续扩高价值真实样本池，优先补 {NEXT_SUGGEST_EDITS_HIGH_VALUE_GROUP} 这组非重复高价值真实样本。"
        )
    else:
        priority_recommendation = (
            "既有 suggest_flowsheet_edits 高价值真实扩样入口已跑通；"
            "下一步先定义新的非重复高价值样本组，避免回到 remaining-horizontal-gaps 或重复 replay 扩样。"
        )
    if next_group:
        priority_category = "teacher_comparison_capture"
        priority_recommendation = f"补齐 suggest_flowsheet_edits 的 {next_group} default teacher capture。"
    elif str(governance.get("real_derived_negative_index_blocker") or "").strip():
        priority_category = "complete_real_derived_assets"
        priority_recommendation = "先把 suggest_flowsheet_edits 的 real-derived 负例资产补齐，再回到真实样本池扩样。"
    return {
        "chain_id": "radishflow-suggest-flowsheet-edits",
        "project": "radishflow",
        "task": "suggest_flowsheet_edits",
        "eval_task": "radishflow-suggest-edits",
        "formal_real_batch_count": len(batches),
        "formal_batches": batches,
        "latest_formal_batch": latest_batch,
        "coverage_summary": build_coverage_summary(
            total_eval_sample_count=total_eval_sample_count,
            real_captured_sample_count=len(real_captured_sample_ids),
            latest_batch_matched_sample_count=latest_batch["matched_sample_count"],
            latest_batch_record_count=latest_batch["record_count"],
        ),
        "coverage": {
            "kind": "profile_coverage",
            "sample_count": total_eval_sample_count,
            "fully_covered_count": int(coverage_report.get("fully_covered_count") or 0),
            "missing_count": int(coverage_report.get("missing_count") or 0),
            "real_captured_sample_ids": real_captured_sample_ids,
            "teacher_comparison_target": str(coverage_report.get("teacher_comparison_target") or "").strip(),
            "teacher_comparison_candidate_count": len(teacher_candidates),
            "teacher_comparison_candidates": teacher_candidates,
            "next_teacher_comparison_group": next_group,
            "next_high_value_capture_group": "" if next_group else NEXT_SUGGEST_EDITS_HIGH_VALUE_GROUP,
        },
        "governance": governance,
        "artifact_summary": (
            {
                "path": artifact_summary_path,
                "summary": artifact_summary.get("summary"),
            }
            if artifact_summary is not None
            else None
        ),
        "priority": {
            "rank": 1,
            "category": priority_category,
            "recommended_action": priority_recommendation,
        },
        "next_gap": next_gap,
    }


def build_ghost_chain() -> dict[str, Any]:
    fixture_entries = load_json_list(GHOST_BATCH_FIXTURE)
    batches = load_real_batches(GHOST_BATCH_FIXTURE)
    latest_batch = batches[-1]
    governance_entry = find_latest_governance_entry(
        fixture_entries,
        preferred_keys=(
            "negative_replay_index",
            "cross_sample_replay_index",
            "recommended_summary",
            "cross_sample_recommended_summary",
        ),
    )
    artifact_summary_path, artifact_summary = load_optional_radishflow_artifact_summary(governance_entry)
    real_derived_fixture = load_json_document(RADISHFLOW_GHOST_REAL_DERIVED_FIXTURE)
    real_derived_index_path = resolve_repo_relative(str(real_derived_fixture.get("index") or "").strip())
    real_derived_index = load_json_document(real_derived_index_path) if real_derived_index_path.is_file() else None
    real_derived_summary: dict[str, Any] = {}
    if real_derived_index is not None:
        real_derived_summary = real_derived_index.get("summary") or {}
        if not isinstance(real_derived_summary, dict):
            raise SystemExit(f"real-derived summary must be an object: {make_repo_relative(real_derived_index_path)}")
    all_sample_ids = collect_unique_sample_ids(batches)
    covered_sample_ids = set(all_sample_ids)
    total_eval_sample_count = count_eval_samples("datasets/eval/radishflow/suggest-ghost-completion-*.json")
    pending_groups = build_pending_group_summaries(
        REPO_ROOT,
        covered_sample_ids=covered_sample_ids,
        group_names=HIGH_VALUE_PRIORITY_GROUPS,
    )
    next_group = pending_groups[0]["group_name"] if pending_groups else ""
    governance: dict[str, Any] = {
        "level": (
            "artifact_summary_replay_and_real_derived"
            if real_derived_index is not None
            else "manifest_audit_and_artifact_summary"
            if artifact_summary is not None
            else "manifest_audit_only"
        ),
        "artifact_summary": artifact_summary is not None,
        "negative_replay_index": False,
        "cross_sample_negative_replay_index": False,
        "recommended_negative_replay_summary": False,
        "cross_sample_recommended_negative_replay_summary": False,
        "real_derived_negative_index": real_derived_index is not None,
        "real_derived_pattern_group_count": int(real_derived_summary.get("pattern_group_count") or 0),
        "real_derived_violation_group_count": int(real_derived_summary.get("violation_group_count") or 0),
        **make_governance_blockers(
            negative_replay_index_blocker=MISSING_NEGATIVE_SAMPLES,
            cross_sample_negative_replay_index_blocker=MISSING_NEGATIVE_SAMPLES,
            recommended_negative_replay_summary_blocker=MISSING_NEGATIVE_SAMPLES,
            cross_sample_recommended_negative_replay_summary_blocker=MISSING_NEGATIVE_SAMPLES,
            real_derived_negative_index_blocker=(
                "" if real_derived_index is not None else MISSING_REAL_DERIVED_NEGATIVE_SAMPLES
            ),
        ),
    }
    if artifact_summary is not None:
        artifacts = artifact_summary.get("artifacts") or {}
        if not isinstance(artifacts, dict):
            raise SystemExit(f"artifacts must be an object: {artifact_summary_path}")

        def artifact_exists(key: str) -> bool:
            value = artifacts.get(key)
            return isinstance(value, dict) and safe_bool(value.get("exists"))

        governance["negative_replay_index"] = artifact_exists("negative_replay_index")
        governance["cross_sample_negative_replay_index"] = artifact_exists("cross_sample_negative_replay_index")
        governance["recommended_negative_replay_summary"] = artifact_exists("recommended_negative_replay_summary")
        governance["cross_sample_recommended_negative_replay_summary"] = artifact_exists(
            "cross_sample_recommended_negative_replay_summary"
        )
        if governance["negative_replay_index"]:
            governance["negative_replay_index_blocker"] = ""
        if governance["cross_sample_negative_replay_index"]:
            governance["cross_sample_negative_replay_index_blocker"] = ""
        if governance["recommended_negative_replay_summary"] or artifact_exists("negative_replay_index"):
            governance["recommended_negative_replay_summary_blocker"] = ""
        if governance["cross_sample_recommended_negative_replay_summary"] or artifact_exists(
            "cross_sample_negative_replay_index"
        ):
            governance["cross_sample_recommended_negative_replay_summary_blocker"] = ""
    priority_category = "expand_real_capture_pool"
    priority_recommendation = "继续扩非重复高价值真实 capture 样本池，避免回到重复 replay 扩样。"
    next_gap = (
        "same-sample / cross-sample replay 与首批 real-derived negative 已接通；"
        "下一步应继续扩非重复高价值真实 capture 样本池。"
    )
    if next_group:
        priority_category = "grouped_real_capture"
        priority_recommendation = f"优先补 suggest_ghost_completion 的 {next_group} sample group。"
        next_gap = (
            "same-sample / cross-sample replay 与首批 real-derived negative 已接通；"
            f"下一步应先补 {next_group} 这组尚未真实化的高价值 ghost 样本。"
        )
    return {
        "chain_id": "radishflow-suggest-ghost-completion",
        "project": "radishflow",
        "task": "suggest_ghost_completion",
        "eval_task": "radishflow-ghost-completion",
        "formal_real_batch_count": len(batches),
        "formal_batches": batches,
        "latest_formal_batch": latest_batch,
        "coverage_summary": build_coverage_summary(
            total_eval_sample_count=total_eval_sample_count,
            real_captured_sample_count=len(all_sample_ids),
            latest_batch_matched_sample_count=latest_batch["matched_sample_count"],
            latest_batch_record_count=latest_batch["record_count"],
        ),
        "coverage": {
            "kind": "current_real_capture_scope",
            "total_eval_sample_count": total_eval_sample_count,
            "real_captured_sample_count": len(all_sample_ids),
            "real_captured_sample_ids": all_sample_ids,
            "uncovered_eval_sample_count": max(0, total_eval_sample_count - len(all_sample_ids)),
            "latest_batch_record_count": latest_batch["record_count"],
            "next_real_capture_group": next_group,
            "pending_sample_groups": pending_groups,
            "scope_note": (
                "当前正式真实 capture 已从固定 PoC trio 扩到十六批高价值链式样本。"
                if len(all_sample_ids) > 74
                else
                "当前正式真实 capture 已从固定 PoC trio 扩到十五批高价值链式样本。"
                if len(all_sample_ids) > 68
                else "当前正式真实 capture 已从固定 PoC trio 扩到十四批高价值链式样本。"
                if len(all_sample_ids) > 62
                else "当前正式真实 capture 已从固定 PoC trio 扩到十三批高价值链式样本。"
                if len(all_sample_ids) > 56
                else "当前正式真实 capture 已从固定 PoC trio 扩到十二批高价值链式样本。"
                if len(all_sample_ids) > 50
                else "当前正式真实 capture 已从固定 PoC trio 扩到十一批高价值链式样本。"
                if len(all_sample_ids) > 44
                else "当前正式真实 capture 已从固定 PoC trio 扩到十批高价值链式样本。"
                if len(all_sample_ids) > 38
                else "当前正式真实 capture 已从固定 PoC trio 扩到九批高价值链式样本。"
                if len(all_sample_ids) > 32
                else "当前正式真实 capture 已从固定 PoC trio 扩到八批高价值链式样本。"
                if len(all_sample_ids) > 26
                else "当前正式真实 capture 已从固定 PoC trio 扩到七批高价值链式样本。"
                if len(all_sample_ids) > 21
                else "当前正式真实 capture 已从固定 PoC trio 扩到六批高价值链式样本。"
                if len(all_sample_ids) > 18
                else "当前正式真实 capture 已从固定 PoC trio 扩到五批高价值链式样本。"
                if len(all_sample_ids) > 15
                else "当前正式真实 capture 已从固定 PoC trio 扩到四批高价值链式样本。"
                if len(all_sample_ids) > 12
                else "当前正式真实 capture 已从固定 PoC trio 扩到三批高价值链式样本。"
                if len(all_sample_ids) > 9
                else "当前正式真实 capture 已从固定 PoC trio 扩到两批高价值链式样本。"
                if len(all_sample_ids) > 6
                else "当前正式真实 capture 已从固定 PoC trio 扩到首批高价值链式样本。"
                if len(all_sample_ids) > 3
                else "当前正式真实 capture 仍集中在固定 PoC trio。"
            ),
        },
        "governance": governance,
        "artifact_summary": (
            {
                "path": artifact_summary_path,
                "summary": artifact_summary.get("summary"),
            }
            if artifact_summary is not None
            else None
        ),
        "priority": {
            "rank": 2,
            "category": priority_category,
            "recommended_action": priority_recommendation,
        },
        "next_gap": next_gap,
    }


def build_radish_docs_chain() -> dict[str, Any]:
    batches = load_real_batches(RADISH_DOCS_BATCH_FIXTURE)
    latest_batch = batches[-1]
    all_sample_ids = collect_unique_sample_ids(batches)
    fixture_entries = load_json_list(RADISH_DOCS_BATCH_FIXTURE)
    latest_fixture = fixture_entries[-1]
    artifact_summary_path = resolve_repo_relative(str(latest_fixture.get("artifact_summary") or "").strip())
    if not artifact_summary_path.is_file():
        raise SystemExit(f"artifact summary not found: {make_repo_relative(artifact_summary_path)}")
    artifact_summary = load_json_document(artifact_summary_path)
    real_derived_fixture = load_json_document(RADISH_DOCS_REAL_DERIVED_FIXTURE)
    real_derived_index_path = resolve_repo_relative(str(real_derived_fixture.get("index") or "").strip())
    if not real_derived_index_path.is_file():
        raise SystemExit(f"real-derived index not found: {make_repo_relative(real_derived_index_path)}")
    real_derived_index = load_json_document(real_derived_index_path)
    artifacts = artifact_summary.get("artifacts") or {}
    if not isinstance(artifacts, dict):
        raise SystemExit(f"artifacts must be an object: {make_repo_relative(artifact_summary_path)}")
    summary = artifact_summary.get("summary") or {}
    if not isinstance(summary, dict):
        raise SystemExit(f"summary must be an object: {make_repo_relative(artifact_summary_path)}")

    def artifact_exists(key: str) -> bool:
        value = artifacts.get(key)
        return isinstance(value, dict) and safe_bool(value.get("exists"))

    total_eval_sample_count = count_eval_samples("datasets/eval/radish/*.json")
    real_derived_summary = real_derived_index.get("summary") or {}
    if not isinstance(real_derived_summary, dict):
        raise SystemExit(f"real-derived summary must be an object: {make_repo_relative(real_derived_index_path)}")
    return {
        "chain_id": "radish-answer-docs-question",
        "project": "radish",
        "task": "answer_docs_question",
        "eval_task": "radish-docs-qa",
        "formal_real_batch_count": len(batches),
        "formal_batches": batches,
        "latest_formal_batch": {
            **latest_batch,
            "artifact_summary_path": make_repo_relative(artifact_summary_path),
        },
        "coverage_summary": build_coverage_summary(
            total_eval_sample_count=total_eval_sample_count,
            real_captured_sample_count=len(all_sample_ids),
            latest_batch_matched_sample_count=latest_batch["matched_sample_count"],
            latest_batch_record_count=latest_batch["record_count"],
        ),
        "coverage": {
            "kind": "latest_real_batch_eval_coverage",
            "total_eval_sample_count": total_eval_sample_count,
            "real_captured_sample_ids": all_sample_ids,
            "latest_batch_matched_sample_count": latest_batch["matched_sample_count"],
            "latest_batch_full_coverage": latest_batch["matched_sample_count"] == total_eval_sample_count,
            "latest_batch_failed_sample_count": latest_batch["failed_count"],
            "latest_batch_violation_count": latest_batch["violation_count"],
            "audit_interpretation": "最新 real batch 允许保留 fail，用于驱动 negative replay 与 repeated pattern 治理。",
        },
        "governance": {
            "level": "artifact_summary_replay_and_real_derived",
            "artifact_summary": artifact_summary_path.is_file(),
            "negative_replay_index": artifact_exists("negative_replay_index"),
            "cross_sample_negative_replay_index": artifact_exists("cross_sample_negative_replay_index"),
            "recommended_negative_replay_summary": artifact_exists("recommended_negative_replay_summary"),
            "cross_sample_recommended_negative_replay_summary": artifact_exists(
                "cross_sample_recommended_negative_replay_summary"
            ),
            "real_derived_negative_index": real_derived_index_path.is_file(),
            "same_sample_violation_group_count": int(summary.get("violation_group_count") or 0),
            "cross_sample_violation_group_count": int(summary.get("cross_sample_violation_group_count") or 0),
            "real_derived_pattern_group_count": int(real_derived_summary.get("pattern_group_count") or 0),
            "real_derived_violation_group_count": int(real_derived_summary.get("violation_group_count") or 0),
            **make_governance_blockers(),
        },
        "priority": {
            "rank": 3,
            "category": "expand_repeated_patterns",
            "recommended_action": "继续扩 captured negative 与跨 source 复合 drift，把 docs QA repeated pattern 做得更结构化。",
        },
        "next_gap": "把 docs QA 已完成的 artifact summary / replay / real-derived 能力上提为仓库级统一盘点基线。",
    }


def build_report() -> dict[str, Any]:
    chains = [
        build_suggest_edits_chain(),
        build_ghost_chain(),
        build_radish_docs_chain(),
    ]
    latest_clean_chain_count = sum(1 for chain in chains if chain["latest_formal_batch"]["audit_clean"])
    artifact_summary_connected_chain_count = sum(1 for chain in chains if chain["governance"]["artifact_summary"])
    replay_connected_chain_count = sum(1 for chain in chains if chain["governance"]["negative_replay_index"])
    cross_sample_replay_connected_chain_count = sum(
        1 for chain in chains if chain["governance"]["cross_sample_negative_replay_index"]
    )
    cross_sample_recommended_replay_connected_chain_count = sum(
        1 for chain in chains if chain["governance"]["cross_sample_recommended_negative_replay_summary"]
    )
    real_derived_connected_chain_count = sum(1 for chain in chains if chain["governance"]["real_derived_negative_index"])
    replay_asset_gap_chain_count = sum(
        1
        for chain in chains
        if chain["governance"].get("negative_replay_index_blocker") == MISSING_NEGATIVE_SAMPLES
    )
    recommended_replay_asset_gap_chain_count = sum(
        1
        for chain in chains
        if chain["governance"].get("recommended_negative_replay_summary_blocker") == MISSING_NEGATIVE_SAMPLES
    )
    real_derived_asset_gap_chain_count = sum(
        1
        for chain in chains
        if chain["governance"].get("real_derived_negative_index_blocker") == MISSING_REAL_DERIVED_NEGATIVE_SAMPLES
    )
    total_formal_batches = sum(int(chain["formal_real_batch_count"]) for chain in chains)
    full_governance_connected_chain_count = sum(
        1
        for chain in chains
        if (
            chain["governance"]["artifact_summary"]
            and chain["governance"]["negative_replay_index"]
            and chain["governance"]["cross_sample_negative_replay_index"]
            and chain["governance"]["recommended_negative_replay_summary"]
            and chain["governance"]["cross_sample_recommended_negative_replay_summary"]
            and chain["governance"]["real_derived_negative_index"]
        )
    )
    coverage_visible_chain_count = sum(1 for chain in chains if isinstance(chain.get("coverage_summary"), dict))
    priority_queue = sorted(
        [
            {
                "rank": int((chain.get("priority") or {}).get("rank") or 0),
                "chain_id": chain["chain_id"],
                "category": str(((chain.get("priority") or {}).get("category")) or "").strip(),
                "recommended_action": str(((chain.get("priority") or {}).get("recommended_action")) or "").strip(),
            }
            for chain in chains
        ],
        key=lambda item: (item["rank"], item["chain_id"]),
    )
    next_group = str(chains[0]["coverage"].get("next_teacher_comparison_group") or "").strip()
    ghost_next_group = str(chains[1]["coverage"].get("next_real_capture_group") or "").strip()
    if (
        replay_asset_gap_chain_count > 0
        or recommended_replay_asset_gap_chain_count > 0
        or real_derived_asset_gap_chain_count > 0
    ):
        next_mainline_focus = (
            "先把 suggest_flowsheet_edits 的 cross-sample replay / real-derived 缺口收口到 committed 负例资产，"
            f"再回到 suggest_flowsheet_edits 的 {next_group} default teacher capture。"
        )
    elif next_group:
        next_mainline_focus = (
            "suggest_flowsheet_edits 的 replay / real-derived 治理已接通；"
            f"下一步回到 suggest_flowsheet_edits 的 {next_group} default teacher capture。"
        )
    else:
        if ghost_next_group:
            next_mainline_focus = (
                "当前三条主治理链的 replay / real-derived 资产均已接通；"
                f"下一步优先补 suggest_ghost_completion 的 {ghost_next_group} 高价值真实样本池入口。"
            )
        else:
            next_mainline_focus = (
                "当前三条主治理链的 replay / real-derived 资产均已接通；"
                "下一步优先扩 suggest_flowsheet_edits 与 ghost 链的高价值真实样本池。"
            )
    return {
        "schema_version": 1,
        "report": "real_batch_governance_status",
        "generated_on": date.today().isoformat(),
        "mainline_stage": "M3",
        "summary": {
            "chain_count": len(chains),
            "total_formal_real_batch_count": total_formal_batches,
            "latest_clean_chain_count": latest_clean_chain_count,
            "artifact_summary_connected_chain_count": artifact_summary_connected_chain_count,
            "replay_connected_chain_count": replay_connected_chain_count,
            "cross_sample_replay_connected_chain_count": cross_sample_replay_connected_chain_count,
            "cross_sample_recommended_replay_connected_chain_count": cross_sample_recommended_replay_connected_chain_count,
            "real_derived_connected_chain_count": real_derived_connected_chain_count,
            "full_governance_connected_chain_count": full_governance_connected_chain_count,
            "coverage_visible_chain_count": coverage_visible_chain_count,
            "replay_asset_gap_chain_count": replay_asset_gap_chain_count,
            "recommended_replay_asset_gap_chain_count": recommended_replay_asset_gap_chain_count,
            "real_derived_asset_gap_chain_count": real_derived_asset_gap_chain_count,
        },
        "next_mainline_focus": next_mainline_focus,
        "priority_queue": priority_queue,
        "chains": chains,
    }


def print_report(report: dict[str, Any]) -> None:
    summary = report["summary"]
    print(
        "M3 real-batch governance overview:"
        f" chains={summary['chain_count']}"
        f" formal_batches={summary['total_formal_real_batch_count']}"
        f" latest_clean={summary['latest_clean_chain_count']}"
        f" artifact_summary_connected={summary['artifact_summary_connected_chain_count']}"
        f" replay_connected={summary['replay_connected_chain_count']}"
        f" cross_sample_replay_connected={summary['cross_sample_replay_connected_chain_count']}"
        f" cross_sample_recommended_connected={summary['cross_sample_recommended_replay_connected_chain_count']}"
        f" real_derived_connected={summary['real_derived_connected_chain_count']}"
        f" full_governance_connected={summary['full_governance_connected_chain_count']}"
        f" coverage_visible={summary['coverage_visible_chain_count']}"
        f" replay_asset_gap={summary['replay_asset_gap_chain_count']}"
        f" recommended_replay_asset_gap={summary['recommended_replay_asset_gap_chain_count']}"
        f" real_derived_asset_gap={summary['real_derived_asset_gap_chain_count']}"
    )
    for chain in report["chains"]:
        latest_batch = chain["latest_formal_batch"]
        coverage_summary = chain.get("coverage_summary") or {}
        coverage = chain["coverage"]
        governance = chain["governance"]
        print(f"{chain['chain_id']}:")
        print(
            f"  latest_batch={latest_batch['collection_batch']} "
            f"records={latest_batch['record_count']} "
            f"audit_clean={'yes' if latest_batch['audit_clean'] else 'no'} "
            f"(matched={latest_batch['matched_sample_count']} "
            f"passed={latest_batch['passed_count']} "
            f"failed={latest_batch['failed_count']} "
            f"violations={latest_batch['violation_count']})"
        )
        if isinstance(coverage_summary, dict):
            print(
                "  coverage_summary="
                f"real_captured={coverage_summary.get('real_captured_sample_count', 0)}/"
                f"{coverage_summary.get('total_eval_sample_count', 0)} "
                f"latest_batch_matched={coverage_summary.get('latest_batch_matched_sample_count', 0)}/"
                f"{coverage_summary.get('total_eval_sample_count', 0)}"
            )
        if coverage["kind"] == "profile_coverage":
            print(
                "  coverage="
                f"{coverage['fully_covered_count']}/{coverage['sample_count']} four-main-apiyi "
                f"teacher_candidates={coverage['teacher_comparison_candidate_count']} "
                f"next={coverage['next_teacher_comparison_group'] or '(none)'}"
            )
        elif coverage["kind"] == "current_real_capture_scope":
            print(
                "  coverage="
                f"real_captured={coverage['real_captured_sample_count']}/{coverage['total_eval_sample_count']} "
                f"scope={coverage['scope_note']}"
            )
        else:
            print(
                "  coverage="
                f"latest_batch_matched={coverage['latest_batch_matched_sample_count']}/"
                f"{coverage['total_eval_sample_count']} "
                f"full={'yes' if coverage['latest_batch_full_coverage'] else 'no'}"
            )
        print(
            "  governance="
            f"level={governance['level']} "
            f"artifact_summary={'yes' if governance['artifact_summary'] else 'no'} "
            f"replay={'yes' if governance['negative_replay_index'] else 'no'} "
            f"real_derived={'yes' if governance['real_derived_negative_index'] else 'no'}"
        )
        governance_blockers: list[str] = []
        if str(governance.get("negative_replay_index_blocker") or "").strip():
            governance_blockers.append(f"replay_index={governance['negative_replay_index_blocker']}")
        if str(governance.get("cross_sample_negative_replay_index_blocker") or "").strip():
            governance_blockers.append(
                f"cross_replay_index={governance['cross_sample_negative_replay_index_blocker']}"
            )
        if str(governance.get("recommended_negative_replay_summary_blocker") or "").strip():
            governance_blockers.append(
                f"recommended_summary={governance['recommended_negative_replay_summary_blocker']}"
            )
        if str(governance.get("cross_sample_recommended_negative_replay_summary_blocker") or "").strip():
            governance_blockers.append(
                "cross_recommended_summary="
                f"{governance['cross_sample_recommended_negative_replay_summary_blocker']}"
            )
        if str(governance.get("real_derived_negative_index_blocker") or "").strip():
            governance_blockers.append(f"real_derived={governance['real_derived_negative_index_blocker']}")
        if governance_blockers:
            print("  governance_blockers=" + " ".join(governance_blockers))
        priority = chain.get("priority") or {}
        if isinstance(priority, dict):
            print(
                "  priority="
                f"rank={priority.get('rank', 0)} "
                f"category={priority.get('category') or '(none)'} "
                f"action={priority.get('recommended_action') or '(none)'}"
            )
        print(f"  next_gap={chain['next_gap']}")
    if report.get("priority_queue"):
        print("priority_queue:")
        for item in report["priority_queue"]:
            print(
                f"  {item['rank']}. {item['chain_id']} "
                f"[{item['category'] or 'uncategorized'}] {item['recommended_action']}"
            )
    print(f"next_mainline_focus: {report['next_mainline_focus']}")


def write_report(path_value: str, report: dict[str, Any]) -> None:
    path = resolve_repo_relative(path_value)
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"wrote governance report to {make_repo_relative(path)}")


def main() -> int:
    args = parse_args()
    report = build_report()
    print_report(report)
    if args.json_output.strip():
        write_report(args.json_output.strip(), report)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import sys
from collections import defaultdict
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parent.parent.parent
SAMPLE_ROOT = REPO_ROOT / "datasets/eval/radishflow"
RECORD_ROOT = REPO_ROOT / "datasets/eval/candidate-records/radishflow"
TARGET_PROFILES = ["apiyi_cx", "apiyi_cc", "apiyi_ch", "apiyi_de"]
PROFILE_ALIASES = {
    "apiyi-cx": "apiyi_cx",
    "apiyi-cc": "apiyi_cc",
    "apiyi-ch": "apiyi_ch",
    "apiyi-de": "apiyi_de",
    "deepseek": "deepseek",
    "openrouter": "openrouter",
}
RECOMMENDED_GROUPS = {
    "default-early-trio": [
        "radishflow-suggest-flowsheet-edits-reconnect-outlet-001",
        "radishflow-suggest-flowsheet-edits-stream-spec-placeholder-001",
        "radishflow-suggest-flowsheet-edits-three-step-priority-chain-001",
    ],
    "default-selection-ordering": [
        "radishflow-suggest-flowsheet-edits-action-citation-ordering-diagnostic-artifact-snapshot-001",
        "radishflow-suggest-flowsheet-edits-issue-ordering-confirmed-before-unconfirmed-001",
        "radishflow-suggest-flowsheet-edits-same-risk-input-first-ordering-001",
        "radishflow-suggest-flowsheet-edits-selection-chronology-single-actionable-target-001",
        "radishflow-suggest-flowsheet-edits-selection-order-preserved-001",
    ],
    "heater-follow-up": [
        "radishflow-suggest-flowsheet-edits-heater-multi-action-001",
    ],
}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Summarize local real-batch profile coverage for RadishFlow "
            "suggest_flowsheet_edits samples."
        ),
    )
    parser.add_argument(
        "--json-output",
        default="",
        help="Optional path to write the structured coverage report JSON.",
    )
    return parser.parse_args()


def load_json(path: Path) -> dict[str, Any]:
    try:
        document = json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json '{path}': {exc}") from exc
    if not isinstance(document, dict):
        raise SystemExit(f"expected json object: {path}")
    return document


def collect_eval_sample_ids() -> list[str]:
    sample_ids: list[str] = []
    for sample_path in sorted(SAMPLE_ROOT.glob("suggest-flowsheet-edits-*.json")):
        sample = load_json(sample_path)
        sample_id = str(sample.get("sample_id") or "").strip()
        if not sample_id:
            raise SystemExit(f"sample is missing sample_id: {sample_path}")
        sample_ids.append(sample_id)
    return sample_ids


def infer_profile(collection_batch: str) -> str:
    for raw_name, normalized_name in PROFILE_ALIASES.items():
        if raw_name in collection_batch:
            return normalized_name
    return "default"


def collect_coverage() -> dict[str, set[str]]:
    coverage: dict[str, set[str]] = defaultdict(set)
    for record_path in sorted(RECORD_ROOT.glob("*suggest-edits*poc-real*/records/*.record.json")):
        record = load_json(record_path)
        sample_id = str(record.get("sample_id") or "").strip()
        if not sample_id:
            continue
        capture_metadata = record.get("capture_metadata")
        collection_batch = ""
        if isinstance(capture_metadata, dict):
            collection_batch = str(capture_metadata.get("collection_batch") or "").strip()
        coverage[sample_id].add(infer_profile(collection_batch))
    return coverage


def build_report() -> dict[str, Any]:
    sample_ids = collect_eval_sample_ids()
    coverage = collect_coverage()
    missing_samples: list[dict[str, Any]] = []
    for sample_id in sample_ids:
        covered_profiles = sorted(coverage.get(sample_id, set()))
        missing_profiles = [profile for profile in TARGET_PROFILES if profile not in coverage.get(sample_id, set())]
        if missing_profiles:
            missing_samples.append(
                {
                    "sample_id": sample_id,
                    "covered_profiles": covered_profiles,
                    "missing_profiles": missing_profiles,
                }
            )
    missing_sample_ids = {item["sample_id"] for item in missing_samples}
    recommended_groups: list[dict[str, Any]] = []
    for group_name, sample_ids_for_group in RECOMMENDED_GROUPS.items():
        remaining_sample_ids = [sample_id for sample_id in sample_ids_for_group if sample_id in missing_sample_ids]
        if not remaining_sample_ids:
            continue
        recommended_groups.append(
            {
                "group_name": group_name,
                "sample_ids": remaining_sample_ids,
            }
        )
    return {
        "schema_version": 1,
        "task": "radishflow-suggest-edits-profile-coverage",
        "target_profiles": TARGET_PROFILES,
        "sample_count": len(sample_ids),
        "fully_covered_count": len(sample_ids) - len(missing_samples),
        "missing_count": len(missing_samples),
        "missing_samples": missing_samples,
        "recommended_groups": recommended_groups,
    }


def print_report(report: dict[str, Any]) -> None:
    print(
        "RadishFlow suggest_flowsheet_edits local profile coverage:"
        f" total={report['sample_count']} fully_covered={report['fully_covered_count']}"
        f" missing={report['missing_count']}"
    )
    for item in report["missing_samples"]:
        covered_profiles = ",".join(item["covered_profiles"]) if item["covered_profiles"] else "(none)"
        print(item["sample_id"])
        print(f"  covered={covered_profiles}")
        print(f"  missing={','.join(item['missing_profiles'])}")
    if report["recommended_groups"]:
        print("recommended_groups:")
        for group in report["recommended_groups"]:
            print(f"  {group['group_name']}:")
            for sample_id in group["sample_ids"]:
                print(f"    - {sample_id}")


def write_report(path_value: str, report: dict[str, Any]) -> None:
    path = Path(path_value)
    if not path.is_absolute():
        path = (REPO_ROOT / path).resolve()
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"wrote coverage report to {path.relative_to(REPO_ROOT)}")


def main() -> int:
    args = parse_args()
    report = build_report()
    print_report(report)
    if args.json_output.strip():
        write_report(args.json_output.strip(), report)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

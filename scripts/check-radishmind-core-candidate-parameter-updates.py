#!/usr/bin/env python3
from __future__ import annotations

import json
import importlib.util
import sys
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from scripts.eval.core_candidate_scaffold import extract_ordered_parameter_updates  # noqa: E402
from scripts.eval.core_candidate_paths import iter_must_have_path_values  # noqa: E402


VALVE_HOLDOUT_SAMPLE = (
    REPO_ROOT / "datasets/eval/radishflow/suggest-flowsheet-edits-valve-local-fix-vs-global-balance-001.json"
)
CROSS_OBJECT_HOLDOUT_SAMPLE = (
    REPO_ROOT
    / "datasets/eval/radishflow/suggest-flowsheet-edits-cross-object-mixed-risk-reconnect-plus-pump-parameter-001.json"
)
COMPRESSOR_ORDERING_HOLDOUT_SAMPLE = (
    REPO_ROOT / "datasets/eval/radishflow/suggest-flowsheet-edits-compressor-parameter-update-ordering-001.json"
)


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


def main() -> int:
    sample = load_json(VALVE_HOLDOUT_SAMPLE)
    parameter_updates = extract_ordered_parameter_updates(
        sample,
        action_index=0,
        iter_must_have_path_values=iter_must_have_path_values,
    )
    require(
        list(parameter_updates.keys()) == ["pressure_drop_kpa", "opening_percent"],
        "detail-key-only ordered parameter updates must derive outer parameter order",
    )
    require(
        list(parameter_updates["pressure_drop_kpa"].keys()) == ["action", "minimum_value"],
        "pressure_drop_kpa detail keys must preserve declared order",
    )
    require(
        parameter_updates["pressure_drop_kpa"]["minimum_value"] == 10,
        "pressure_drop_kpa minimum_value must preserve the declared numeric threshold",
    )
    require(
        list(parameter_updates["opening_percent"].keys()) == ["action", "suggested_maximum"],
        "opening_percent detail keys must preserve declared order",
    )
    require(
        parameter_updates["opening_percent"]["suggested_maximum"] == 85,
        "opening_percent suggested_maximum must preserve the declared numeric threshold",
    )

    compressor_ordering_sample = load_json(COMPRESSOR_ORDERING_HOLDOUT_SAMPLE)
    compressor_ordering_parameter_updates = extract_ordered_parameter_updates(
        compressor_ordering_sample,
        action_index=0,
        iter_must_have_path_values=iter_must_have_path_values,
    )
    require(
        compressor_ordering_parameter_updates == {
            "outlet_pressure_target_kpa": {
                "action": "review_and_raise_margin",
                "minimum_delta_kpa": 90,
                "reference_stream_id": "feed-9",
            },
            "minimum_flow_bypass_percent": {
                "action": "raise_to_anti_surge_review_floor",
                "suggested_minimum": 8,
            },
            "efficiency_percent": {
                "action": "clamp_to_review_range",
                "suggested_range": [65, 85],
            },
        },
        "full-holdout compressor ordering sample must preserve numeric detail values, not boolean placeholders",
    )

    runner = load_candidate_runner()
    compressor_scaffold = runner.build_response_scaffold(
        project=str(compressor_ordering_sample["project"]),
        task=str(compressor_ordering_sample["task"]),
        sample=compressor_ordering_sample,
    )
    compressor_citations = compressor_scaffold.get("citations")
    require(isinstance(compressor_citations, list), "compressor scaffold citations must be a list")
    require(
        [citation.get("id") for citation in compressor_citations]
        == ["diag-1", "diag-2", "diag-3", "flowdoc-stream-1", "flowdoc-unit-1", "snapshot-1"],
        "compressor scaffold citation ids must use indexed diagnostics, suction stream, unit artifact, and snapshot",
    )
    require(
        [citation.get("locator") for citation in compressor_citations]
        == [
            "context:diagnostics[0]",
            "context:diagnostics[1]",
            "context:diagnostics[2]",
            "artifact:flowsheet_document.streams[0]",
            "artifact:flowsheet_document.units[0]",
            "context:latest_snapshot",
        ],
        "compressor scaffold citation locators must not fall back to broad artifact citation",
    )
    compressor_answers = compressor_scaffold.get("answers")
    require(isinstance(compressor_answers, list) and compressor_answers, "compressor scaffold must keep answers")
    require(
        compressor_answers[0].get("citation_ids")
        == ["diag-1", "diag-2", "diag-3", "flowdoc-stream-1", "flowdoc-unit-1", "snapshot-1"],
        "compressor answer citation ids must preserve declared evidence order",
    )
    compressor_issues = compressor_scaffold.get("issues")
    require(isinstance(compressor_issues, list) and len(compressor_issues) == 3, "compressor scaffold must keep three issues")
    require(
        [issue.get("citation_ids") for issue in compressor_issues]
        == [
            ["diag-1", "flowdoc-stream-1", "flowdoc-unit-1"],
            ["diag-2", "flowdoc-unit-1"],
            ["diag-3", "flowdoc-unit-1"],
        ],
        "compressor issue citation ids must stay diagnostic-first and unit-backed",
    )
    compressor_actions = compressor_scaffold.get("proposed_actions")
    require(
        isinstance(compressor_actions, list) and len(compressor_actions) == 1,
        "compressor scaffold must keep one candidate action",
    )
    require(
        compressor_actions[0].get("citation_ids")
        == ["diag-1", "diag-2", "diag-3", "flowdoc-stream-1", "flowdoc-unit-1", "snapshot-1"],
        "compressor action citation ids must preserve diagnostic/artifact/snapshot evidence order",
    )

    scaffold = runner.build_response_scaffold(
        project=str(sample["project"]),
        task=str(sample["task"]),
        sample=sample,
    )
    citations = scaffold.get("citations")
    require(isinstance(citations, list), "candidate scaffold citations must be a list")
    require(
        [citation.get("id") for citation in citations] == ["diag-1", "diag-2", "flowdoc-unit-1", "snapshot-1"],
        "indexed citation assertions must drive scaffold citation ids",
    )
    require(
        [citation.get("locator") for citation in citations]
        == [
            "context:diagnostics[0]",
            "context:diagnostics[1]",
            "artifact:flowsheet_document.units[0]",
            "context:latest_snapshot",
        ],
        "indexed citation assertions must drive scaffold citation locators",
    )

    cross_object_sample = load_json(CROSS_OBJECT_HOLDOUT_SAMPLE)
    cross_object_parameter_updates = extract_ordered_parameter_updates(
        cross_object_sample,
        action_index=1,
        iter_must_have_path_values=iter_must_have_path_values,
    )
    require(
        list(cross_object_parameter_updates.keys()) == ["outlet_pressure_target_kpa", "efficiency_percent"],
        "action-indexed parameter updates must preserve the second candidate_edit parameter order",
    )
    require(
        cross_object_parameter_updates["outlet_pressure_target_kpa"] == {
            "action": "review_and_raise_above_inlet",
            "minimum_reference_stream_id": "feed-81",
        },
        "second candidate_edit must preserve outlet pressure detail values",
    )
    require(
        cross_object_parameter_updates["efficiency_percent"] == {
            "action": "clamp_to_review_range",
            "suggested_range": [60, 85],
        },
        "second candidate_edit must preserve efficiency range detail values",
    )

    cross_object_scaffold = runner.build_response_scaffold(
        project=str(cross_object_sample["project"]),
        task=str(cross_object_sample["task"]),
        sample=cross_object_sample,
    )
    actions = cross_object_scaffold.get("proposed_actions")
    require(isinstance(actions, list) and len(actions) == 2, "cross-object scaffold must keep two candidate actions")
    require(
        actions[0].get("target") == {"type": "stream", "id": "cooler-outlet-81"},
        "first cross-object action must stay on the disconnected stream",
    )
    require(
        list(actions[0].get("patch", {}).get("connection_placeholder", {}).keys())
        == ["expected_downstream_kind", "requires_manual_binding", "retain_existing_source_binding"],
        "first cross-object action must preserve ordered connection placeholder keys",
    )
    require(
        actions[1].get("target") == {"type": "unit", "id": "pump-12"},
        "second cross-object action must stay on the pump unit",
    )
    require(
        list(actions[1].get("patch", {}).keys()) == ["parameter_updates", "preserve_topology"],
        "second cross-object action patch keys must not be polluted by the first action connection placeholder",
    )
    require(
        actions[1].get("patch", {}).get("parameter_updates") == cross_object_parameter_updates,
        "second cross-object action must use action-indexed parameter_updates",
    )
    require(
        actions[1].get("citation_ids") == ["diag-2", "diag-3", "flowdoc-unit-1"],
        "second cross-object action must use action-indexed citation ids",
    )

    print("radishmind core candidate parameter updates check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

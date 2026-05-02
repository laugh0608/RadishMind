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

    runner = load_candidate_runner()
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

    print("radishmind core candidate parameter updates check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

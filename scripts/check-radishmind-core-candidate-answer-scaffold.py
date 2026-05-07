#!/usr/bin/env python3
from __future__ import annotations

import importlib.util
import json
import sys
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))


EFFICIENCY_RANGE_SAMPLE = (
    REPO_ROOT / "datasets/eval/radishflow/suggest-flowsheet-edits-efficiency-range-ordering-001.json"
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
    runner = load_candidate_runner()
    sample = load_json(EFFICIENCY_RANGE_SAMPLE)
    scaffold = runner.build_response_scaffold(
        project=str(sample["project"]),
        task=str(sample["task"]),
        sample=sample,
    )
    answers = scaffold.get("answers")
    require(isinstance(answers, list) and len(answers) == 1, "efficiency-range scaffold must include one answer")
    answer = answers[0]
    require(isinstance(answer, dict), "efficiency-range scaffold answer must be an object")
    require(answer.get("kind") == "edit_rationale", "suggest edits answer scaffold must use edit_rationale")
    text = str(answer.get("text") or "")
    require(
        "suggested_range" in text and "candidate_edit" in text and "pump-3" in text,
        "suggest edits answer scaffold must describe the required patch shape, action kind and target",
    )
    require(
        "给出可展示给用户的回答" not in text,
        "suggest edits answer scaffold must not use the generic placeholder text",
    )

    raw_response = {
        "schema_version": 1,
        "status": "ok",
        "project": "radishflow",
        "task": "suggest_flowsheet_edits",
        "summary": "raw response with no answer",
        "answers": [],
        "issues": scaffold["issues"],
        "proposed_actions": scaffold["proposed_actions"],
        "citations": scaffold["citations"],
        "confidence": 0.8,
        "risk_level": "medium",
        "requires_confirmation": True,
    }
    repaired, repaired_paths = runner.repair_candidate_hard_fields(raw_response, sample=sample)
    repaired_answers = repaired.get("answers")
    require(isinstance(repaired_answers, list) and repaired_answers, "repair must restore the missing suggest edits answer")
    require(
        repaired_answers[0].get("kind") == "edit_rationale",
        "repair must restore the task-aware edit_rationale answer",
    )
    require(
        "给出可展示给用户的回答" not in str(repaired_answers[0].get("text") or ""),
        "repair must not restore the generic answer placeholder",
    )
    require("$.answers" in repaired_paths, "repair must report $.answers when restoring a missing answer")

    print("radishmind core candidate answer scaffold check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

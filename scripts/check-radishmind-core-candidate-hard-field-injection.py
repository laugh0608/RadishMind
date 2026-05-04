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


EVIDENCE_GAP_SAMPLE = REPO_ROOT / "datasets/eval/radish/answer-docs-question-evidence-gap-001.json"
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

    evidence_gap = load_json(EVIDENCE_GAP_SAMPLE)
    evidence_response = {
        "schema_version": 1,
        "status": "ok",
        "project": "radish",
        "task": "answer_docs_question",
        "summary": "raw summary",
        "answers": [
            {
                "kind": "direct_answer",
                "text": "raw answer",
                "citation_ids": ["doc-1"],
            }
        ],
        "issues": [],
        "proposed_actions": [
            {
                "kind": "read_only_check",
                "title": "raw action",
                "rationale": "raw rationale",
                "risk_level": "low",
                "requires_confirmation": False,
                "citation_ids": ["doc-1"],
            }
        ],
        "citations": [
            {
                "id": "doc-1",
                "kind": "artifact",
                "label": "Doc",
                "locator": "artifact:doc",
            }
        ],
        "confidence": 0.5,
        "risk_level": "low",
        "requires_confirmation": True,
    }
    injected, paths = runner.inject_candidate_hard_fields(evidence_response, sample=evidence_gap)
    require(injected["status"] == "partial", "evidence-gap status must be injected from freeze")
    require(injected["risk_level"] == "medium", "evidence-gap risk_level must be injected from freeze")
    require(injected["requires_confirmation"] is False, "evidence-gap confirmation must be injected from freeze")
    require(injected["proposed_actions"] == [], "evidence-gap no-action boundary must be injected from freeze")
    require(injected["answers"][0]["text"] == "raw answer", "injection must preserve non-frozen answer text")
    require("$.issues" not in paths, "injection must not rebuild missing issues unless issues are frozen")

    efficiency_range = load_json(EFFICIENCY_RANGE_SAMPLE)
    efficiency_response = {
        "schema_version": 1,
        "status": "ok",
        "project": "radishflow",
        "task": "suggest_flowsheet_edits",
        "summary": "raw summary",
        "answers": [],
        "issues": [],
        "proposed_actions": [
            {
                "kind": "candidate_edit",
                "title": "raw edit",
                "target": {"type": "unit", "id": "wrong"},
                "rationale": "raw rationale",
                "patch": {},
                "risk_level": "low",
                "requires_confirmation": False,
                "citation_ids": [],
            }
        ],
        "citations": [],
        "confidence": 0.5,
        "risk_level": "low",
        "requires_confirmation": False,
    }
    injected_efficiency, efficiency_paths = runner.inject_candidate_hard_fields(efficiency_response, sample=efficiency_range)
    require(injected_efficiency["status"] == "partial", "efficiency status must be injected from freeze")
    require(injected_efficiency["risk_level"] == "medium", "efficiency risk_level must be injected from freeze")
    require(injected_efficiency["requires_confirmation"] is True, "efficiency confirmation must be injected from freeze")
    require(
        injected_efficiency["issues"][0]["message"],
        "efficiency issue injection must preserve schema-required issue message",
    )
    require(
        injected_efficiency["issues"][0]["severity"],
        "efficiency issue injection must preserve schema-required issue severity",
    )
    require(
        injected_efficiency["proposed_actions"][0]["patch"]
        == {
            "parameter_updates": {
                "efficiency_percent": {
                    "action": "clamp_to_review_range",
                    "suggested_range": [65, 82],
                }
            },
            "preserve_topology": True,
        },
        "efficiency patch must be injected from freeze",
    )
    require(
        injected_efficiency["proposed_actions"][0]["target"] == {"type": "unit", "id": "pump-3"},
        "efficiency target must be injected from freeze",
    )
    require("$.issues[0].code" in efficiency_paths, "injection must preserve explicit issue code path")
    require("$.issues[0].message" in efficiency_paths, "injection must complete schema-required issue message")
    require("$.issues[0].severity" in efficiency_paths, "injection must complete schema-required issue severity")
    require("$.answers" not in efficiency_paths, "injection must not synthesize full answer scaffold")

    print("radishmind core candidate hard-field injection check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

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


DOCS_CONFLICT_SAMPLE = (
    REPO_ROOT / "datasets/eval/radish/answer-docs-question-docs-faq-forum-conflict-001.json"
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
    docs_conflict = load_json(DOCS_CONFLICT_SAMPLE)
    scaffold = runner.build_response_scaffold(
        project=str(docs_conflict["project"]),
        task=str(docs_conflict["task"]),
        sample=docs_conflict,
    )
    citations = scaffold.get("citations")
    require(isinstance(citations, list), "docs source-conflict scaffold citations must be a list")
    require(
        [citation.get("id") for citation in citations] == ["doc-1", "faq-1", "forum-1"],
        "docs source-conflict scaffold must preserve golden citation ids by index",
    )
    require(
        [citation.get("locator") for citation in citations]
        == ["artifact:docs_page", "artifact:faq_excerpt", "artifact:forum_excerpt"],
        "docs source-conflict scaffold must preserve golden citation locators by index",
    )

    raw_response = {
        "schema_version": 1,
        "status": "ok",
        "project": "radish",
        "task": "answer_docs_question",
        "summary": "raw response with only the primary citation",
        "answers": [
            {
                "kind": "direct_answer",
                "text": "根据正式文档应保持 slug 稳定。",
            }
        ],
        "issues": [],
        "proposed_actions": [
            {
                "kind": "read_only_check",
                "title": "执行只读核查",
                "rationale": "仅建议调用侧进行只读核查，不写回业务状态。",
                "risk_level": "low",
                "requires_confirmation": False,
                "citation_ids": ["doc-1"],
            }
        ],
        "citations": [citations[0]],
        "confidence": 0.8,
        "risk_level": "low",
        "requires_confirmation": False,
    }
    repaired, repaired_paths = runner.repair_candidate_hard_fields(raw_response, sample=docs_conflict)
    repaired_citations = repaired.get("citations")
    require(isinstance(repaired_citations, list), "repaired docs source-conflict citations must be a list")
    require(
        [citation.get("id") for citation in repaired_citations] == ["doc-1", "faq-1", "forum-1"],
        "repair must restore the golden docs/faq/forum citation ids instead of padding duplicate docs citations",
    )
    require(
        "$.citations" in repaired_paths,
        "repair must report $.citations when restoring missing source-context citations",
    )

    print("radishmind core candidate citation scaffold check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

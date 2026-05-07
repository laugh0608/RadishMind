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
EVIDENCE_GAP_SAMPLE = REPO_ROOT / "datasets/eval/radish/answer-docs-question-evidence-gap-001.json"


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
    docs_guidance = runner.build_task_guidance(
        project=str(docs_conflict["project"]),
        task=str(docs_conflict["task"]),
        sample=docs_conflict,
    )
    require(
        "本样本已经声明 required_action_kinds，必须按样本规则输出对应 proposed_actions" in docs_guidance,
        "docs QA required-action samples must say sample-level required_action_kinds override task defaults",
    )
    require(
        "read_only_check" in docs_guidance,
        "docs QA required-action prompt guidance must name read_only_check",
    )
    require(
        "通常不生成 proposed_actions" not in docs_guidance,
        "docs QA required-action prompt guidance must not keep the generic no-action wording",
    )

    evidence_gap = load_json(EVIDENCE_GAP_SAMPLE)
    evidence_guidance = runner.build_task_guidance(
        project=str(evidence_gap["project"]),
        task=str(evidence_gap["task"]),
        sample=evidence_gap,
    )
    require(
        "本样本不允许候选动作" in evidence_guidance,
        "docs QA no-action samples must keep the explicit no-action rule",
    )
    require(
        "没有样本级 required_action_kinds 或 JSON path 明确要求时，通常不生成 proposed_actions" in evidence_guidance,
        "docs QA no-action samples may keep the generic no-action wording",
    )

    print("radishmind core candidate prompt policy check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

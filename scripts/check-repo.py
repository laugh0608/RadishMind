#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import subprocess
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parent.parent
REQUIRED_FILES = [
    "AGENTS.md",
    "CLAUDE.md",
    "LICENSE",
    ".editorconfig",
    ".gitattributes",
    ".github/PULL_REQUEST_TEMPLATE.md",
    ".github/rulesets/README.md",
    ".github/rulesets/master-protection.json",
    ".github/workflows/pr-check.yml",
    ".github/workflows/release-check.yml",
    "README.md",
    "docs/README.md",
    "docs/task-cards/README.md",
    "docs/task-cards/radishflow-explain-diagnostics.md",
    "docs/task-cards/radishflow-suggest-flowsheet-edits.md",
    "docs/task-cards/radishflow-explain-control-plane-state.md",
    "docs/task-cards/radish-answer-docs-question.md",
    "docs/radishmind-product-scope.md",
    "docs/radishmind-architecture.md",
    "docs/radishmind-roadmap.md",
    "docs/radishmind-integration-contracts.md",
    "docs/adr/0001-branch-and-pr-governance.md",
    "docs/devlogs/README.md",
    "docs/devlogs/2026-W14.md",
    "contracts/README.md",
    "contracts/copilot-request.schema.json",
    "contracts/copilot-response.schema.json",
    "datasets/README.md",
    "datasets/eval/README.md",
    "datasets/eval/radishflow-task-sample.schema.json",
    "datasets/eval/radish-task-sample.schema.json",
    "datasets/eval/radishflow/explain-diagnostics-global-balance-gap-001.json",
    "datasets/eval/radishflow/explain-diagnostics-multi-object-feed-conflict-001.json",
    "datasets/eval/radishflow/explain-diagnostics-stream-spec-missing-001.json",
    "datasets/eval/radishflow/explain-diagnostics-unit-not-converged-001.json",
    "datasets/eval/radishflow/explain-control-plane-entitlement-expired-001.json",
    "datasets/eval/radishflow/explain-control-plane-package-sync-warning-001.json",
    "datasets/eval/radishflow/suggest-flowsheet-edits-stream-spec-placeholder-001.json",
    "datasets/eval/radishflow/suggest-flowsheet-edits-reconnect-outlet-001.json",
    "datasets/eval/radish/answer-docs-question-direct-answer-001.json",
    "datasets/eval/radish/answer-docs-question-evidence-gap-001.json",
    "datasets/eval/radish/answer-docs-question-navigation-001.json",
    "datasets/eval/radish/answer-docs-question-role-example-boundary-001.json",
    "scripts/check-radishflow-control-plane-eval.ps1",
    "scripts/check-radishflow-control-plane-eval.sh",
    "scripts/check-radishflow-diagnostics-eval.ps1",
    "scripts/check-radishflow-diagnostics-eval.sh",
    "scripts/check-radishflow-suggest-edits-eval.ps1",
    "scripts/check-radishflow-suggest-edits-eval.sh",
    "scripts/check-radish-docs-qa-eval.ps1",
    "scripts/check-radish-docs-qa-eval.sh",
    "scripts/check-text-files.py",
    "scripts/check-repo.py",
    "scripts/run-eval-regression.py",
    "scripts/run-radishflow-control-plane-regression.ps1",
    "scripts/run-radishflow-control-plane-regression.sh",
    "scripts/run-radishflow-diagnostics-regression.ps1",
    "scripts/run-radishflow-diagnostics-regression.sh",
    "scripts/run-radishflow-suggest-edits-regression.ps1",
    "scripts/run-radishflow-suggest-edits-regression.sh",
    "scripts/run-radish-docs-qa-regression.ps1",
    "scripts/run-radish-docs-qa-regression.sh",
    "scripts/check-text-files.ps1",
    "scripts/check-text-files.sh",
    "scripts/check-repo.ps1",
    "scripts/check-repo.sh",
]


def run_python_script(script_name: str, args: list[str]) -> None:
    result = subprocess.run([sys.executable, str(REPO_ROOT / "scripts" / script_name), *args], cwd=REPO_ROOT)
    if result.returncode != 0:
        raise SystemExit(result.returncode)


def check_required_files() -> None:
    for relative_path in REQUIRED_FILES:
        if not (REPO_ROOT / relative_path).is_file():
            raise SystemExit(f"missing required file: {relative_path}")


def check_content_baseline() -> None:
    agents_content = (REPO_ROOT / "AGENTS.md").read_text(encoding="utf-8")
    if "当前常态开发分支为 `dev`" not in agents_content:
        raise SystemExit("AGENTS.md does not mention dev as the default development branch")

    pr_template_content = (REPO_ROOT / ".github/PULL_REQUEST_TEMPLATE.md").read_text(encoding="utf-8")
    if "默认目标分支为 `dev`" not in pr_template_content:
        raise SystemExit("PULL_REQUEST_TEMPLATE.md does not mention dev as the default target branch")

    ruleset = json.loads((REPO_ROOT / ".github/rulesets/master-protection.json").read_text(encoding="utf-8"))
    include_refs = (((ruleset.get("conditions") or {}).get("ref_name") or {}).get("include") or [])
    if "refs/heads/master" not in include_refs:
        raise SystemExit("master-protection.json does not target refs/heads/master")

    required_check_rule = next((rule for rule in ruleset.get("rules", []) if rule.get("type") == "required_status_checks"), None)
    if required_check_rule is None:
        raise SystemExit("master-protection.json is missing required_status_checks")

    contexts = [
        item.get("context")
        for item in ((required_check_rule.get("parameters") or {}).get("required_status_checks") or [])
    ]
    if "Repo Hygiene" not in contexts:
        raise SystemExit("master-protection.json is missing Repo Hygiene required check")
    if "Planning Baseline" not in contexts:
        raise SystemExit("master-protection.json is missing Planning Baseline required check")

    pr_workflow = (REPO_ROOT / ".github/workflows/pr-check.yml").read_text(encoding="utf-8")
    for pattern in ("name: PR Checks", "- master", "name: Repo Hygiene", "name: Planning Baseline"):
        if pattern not in pr_workflow:
            raise SystemExit(f".github/workflows/pr-check.yml is missing expected content: {pattern}")

    release_workflow = (REPO_ROOT / ".github/workflows/release-check.yml").read_text(encoding="utf-8")
    for pattern in (
        "name: Release Checks",
        "v*-dev",
        "v*-test",
        "v*-release",
        "name: Release Repo Hygiene",
        "name: Release Planning Baseline",
    ):
        if pattern not in release_workflow:
            raise SystemExit(f".github/workflows/release-check.yml is missing expected content: {pattern}")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--skip-text-files", action="store_true")
    return parser.parse_args()


def main() -> int:
    args = parse_args()

    if not args.skip_text_files:
        run_python_script("check-text-files.py", [])

    run_python_script("run-eval-regression.py", ["radish-docs-qa", "--fail-on-violation"])
    run_python_script("run-eval-regression.py", ["radishflow-control-plane", "--fail-on-violation"])
    run_python_script("run-eval-regression.py", ["radishflow-diagnostics", "--fail-on-violation"])
    run_python_script("run-eval-regression.py", ["radishflow-suggest-edits", "--fail-on-violation"])

    check_required_files()
    check_content_baseline()

    print("repository baseline checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

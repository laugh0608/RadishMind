#!/usr/bin/env python3
from __future__ import annotations

import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parent.parent
DEFAULT_WARN_LINES = 500
DEFAULT_ERROR_LINES = 800
ENTRY_ERROR_LINES = 250
HISTORY_WARN_LINES = 350
HISTORY_ERROR_LINES = 600
ALLOWLIST_MARKER = "markdown-size-allow:"
ENTRY_DOCS = {
    "docs/README.md",
    "docs/radishmind-current-focus.md",
    "docs/radishmind-product-scope.md",
    "docs/radishmind-roadmap.md",
    "docs/radishmind-architecture.md",
    "docs/radishmind-code-standards.md",
    "docs/radishmind-capability-matrix.md",
}
HISTORY_PREFIXES = (
    "docs/devlogs/",
    "docs/task-cards/",
)


@dataclass(frozen=True)
class MarkdownBudget:
    warning_lines: int
    error_lines: int
    label: str


def iter_markdown_files() -> list[str]:
    result = subprocess.run(
        ["git", "-C", str(REPO_ROOT), "ls-files", "--cached", "--others", "--exclude-standard", "-z"],
        check=True,
        capture_output=True,
    )
    paths = [item.decode("utf-8") for item in result.stdout.split(b"\0") if item]
    return sorted(path for path in paths if path.endswith(".md"))


def budget_for(relative_path: str) -> MarkdownBudget:
    if relative_path in ENTRY_DOCS:
        return MarkdownBudget(ENTRY_ERROR_LINES, ENTRY_ERROR_LINES, "entry")
    if any(relative_path.startswith(prefix) for prefix in HISTORY_PREFIXES):
        return MarkdownBudget(HISTORY_WARN_LINES, HISTORY_ERROR_LINES, "history")
    return MarkdownBudget(DEFAULT_WARN_LINES, DEFAULT_ERROR_LINES, "default")


def count_lines(path: Path) -> int:
    text = path.read_text(encoding="utf-8")
    if not text:
        return 0
    return len(text.splitlines())


def has_allowlist_marker(path: Path) -> bool:
    head = "\n".join(path.read_text(encoding="utf-8").splitlines()[:20])
    return ALLOWLIST_MARKER in head


def main() -> int:
    warnings: list[str] = []
    errors: list[str] = []

    for relative_path in iter_markdown_files():
        path = REPO_ROOT / relative_path
        if not path.is_file():
            continue
        line_count = count_lines(path)
        budget = budget_for(relative_path)
        if line_count > budget.error_lines:
            if has_allowlist_marker(path):
                warnings.append(
                    f"{relative_path}: {line_count} lines exceeds {budget.label} error budget "
                    f"({budget.error_lines}) but is allowlisted"
                )
            else:
                errors.append(
                    f"{relative_path}: {line_count} lines exceeds {budget.label} error budget "
                    f"({budget.error_lines}); split the document or add a justified allowlist marker"
                )
        elif line_count > budget.warning_lines:
            warnings.append(
                f"{relative_path}: {line_count} lines exceeds {budget.label} warning budget "
                f"({budget.warning_lines})"
            )

    for warning in warnings:
        print(f"warning: {warning}", file=sys.stderr)
    if errors:
        for error in errors:
            print(error, file=sys.stderr)
        return 1

    print("markdown size checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

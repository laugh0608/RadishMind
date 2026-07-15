#!/usr/bin/env python3
from __future__ import annotations

import re
import subprocess
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[3]
PLATFORM_ROOT = REPO_ROOT / "services/platform"
PACKAGE_PREFIX = "radishmind.local/services/platform/"

CORE_COVERAGE_BUDGETS = {
    "internal/bridge": 75.0,
    "internal/config": 80.0,
    "internal/diagnostics": 85.0,
    "internal/httpapi": 70.0,
    "internal/secretbackend": 80.0,
    "internal/sqlitedev": 75.0,
}

COVERAGE_PATTERN = re.compile(
    rf"(?P<package>{re.escape(PACKAGE_PREFIX)}internal/[^\s]+).*"
    r"coverage:\s+(?P<coverage>\d+(?:\.\d+)?)%\s+of\s+statements"
)


def parse_core_coverage(output: str) -> dict[str, float]:
    coverage_by_package: dict[str, float] = {}
    for line in output.splitlines():
        match = COVERAGE_PATTERN.search(line)
        if match is None:
            continue
        package = match.group("package").removeprefix(PACKAGE_PREFIX)
        if package in CORE_COVERAGE_BUDGETS:
            coverage_by_package[package] = float(match.group("coverage"))
    return coverage_by_package


def run_go_coverage() -> tuple[int, str]:
    try:
        completed = subprocess.run(
            ["go", "test", "-count=1", "-cover", "./..."],
            cwd=PLATFORM_ROOT,
            check=False,
            capture_output=True,
            text=True,
        )
    except OSError as error:
        raise SystemExit(f"could not start Go coverage: {error}") from error

    output = completed.stdout
    if completed.stderr:
        output = f"{output}{completed.stderr}"
    return completed.returncode, output


def main() -> None:
    return_code, output = run_go_coverage()
    if output:
        print(output, end="" if output.endswith("\n") else "\n")
    if return_code != 0:
        raise SystemExit(return_code)

    coverage_by_package = parse_core_coverage(output)
    missing = sorted(set(CORE_COVERAGE_BUDGETS) - set(coverage_by_package))
    if missing:
        raise SystemExit(f"platform core coverage output missing packages: {missing}")

    failures: list[str] = []
    for package, budget in CORE_COVERAGE_BUDGETS.items():
        actual = coverage_by_package[package]
        print(f"platform core coverage: {package}={actual:.1f}% budget={budget:.1f}%")
        if actual < budget:
            failures.append(f"{package}: {actual:.1f}% < {budget:.1f}%")

    if failures:
        raise SystemExit("platform core coverage budget failed: " + "; ".join(failures))
    print("platform core coverage budgets passed.")


if __name__ == "__main__":
    try:
        main()
    except BrokenPipeError:
        sys.exit(1)

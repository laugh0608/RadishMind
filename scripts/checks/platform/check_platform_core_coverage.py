#!/usr/bin/env python3
from __future__ import annotations

import re
import subprocess
import sys
import tempfile
from dataclasses import dataclass, field
from pathlib import Path, PurePosixPath


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

POSTGRES_FILE_TOKEN = re.compile(r"(?:^|_)postgres(?:_|$)")


@dataclass
class CoverageSummary:
    covered_statements: int = 0
    total_statements: int = 0
    excluded_statements: int = 0
    excluded_files: set[str] = field(default_factory=set)

    @property
    def percentage(self) -> float:
        if self.total_statements == 0:
            return 0.0
        return self.covered_statements * 100.0 / self.total_statements


def package_for_profile_file(profile_file: str) -> str | None:
    if not profile_file.startswith(PACKAGE_PREFIX):
        return None
    relative_path = PurePosixPath(profile_file.removeprefix(PACKAGE_PREFIX))
    return relative_path.parent.as_posix()


def is_postgres_specific_httpapi_file(package: str, profile_file: str) -> bool:
    if package != "internal/httpapi":
        return False
    filename = PurePosixPath(profile_file).name
    if not filename.endswith(".go"):
        return False
    return POSTGRES_FILE_TOKEN.search(filename.removesuffix(".go")) is not None


def parse_cover_profile(profile_text: str) -> dict[str, CoverageSummary]:
    summaries: dict[str, CoverageSummary] = {
        package: CoverageSummary() for package in CORE_COVERAGE_BUDGETS
    }
    lines = profile_text.splitlines()
    if not lines or not lines[0].startswith("mode: "):
        raise ValueError("Go coverage profile is missing its mode header")

    for line_number, line in enumerate(lines[1:], start=2):
        if not line.strip():
            continue
        try:
            location, statement_count_text, execution_count_text = line.rsplit(maxsplit=2)
            profile_file, _ = location.rsplit(":", maxsplit=1)
            statement_count = int(statement_count_text)
            execution_count = int(execution_count_text)
        except (ValueError, TypeError) as error:
            raise ValueError(f"invalid Go coverage profile line {line_number}: {line}") from error

        package = package_for_profile_file(profile_file)
        if package not in summaries:
            continue
        summary = summaries[package]
        if is_postgres_specific_httpapi_file(package, profile_file):
            summary.excluded_statements += statement_count
            summary.excluded_files.add(profile_file)
            continue

        summary.total_statements += statement_count
        if execution_count > 0:
            summary.covered_statements += statement_count

    return summaries


def run_go_coverage(profile_path: Path) -> tuple[int, str]:
    try:
        completed = subprocess.run(
            [
                "go",
                "test",
                "-count=1",
                "-covermode=count",
                f"-coverprofile={profile_path}",
                "./...",
            ],
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
    with tempfile.TemporaryDirectory(prefix="radishmind-platform-coverage-") as temp_directory:
        profile_path = Path(temp_directory) / "platform.cover"
        return_code, output = run_go_coverage(profile_path)
        if output:
            print(output, end="" if output.endswith("\n") else "\n")
        if return_code != 0:
            raise SystemExit(return_code)
        if not profile_path.is_file():
            raise SystemExit("Go coverage profile was not created")
        try:
            summaries = parse_cover_profile(profile_path.read_text(encoding="utf-8"))
        except (OSError, UnicodeError, ValueError) as error:
            raise SystemExit(f"could not read Go coverage profile: {error}") from error

    missing = sorted(
        package
        for package, summary in summaries.items()
        if summary.total_statements == 0
    )
    if missing:
        raise SystemExit(f"platform core coverage profile missing packages: {missing}")

    failures: list[str] = []
    for package, budget in CORE_COVERAGE_BUDGETS.items():
        summary = summaries[package]
        detail = (
            f"platform core coverage: {package}={summary.percentage:.1f}% "
            f"budget={budget:.1f}% statements={summary.covered_statements}/{summary.total_statements}"
        )
        if summary.excluded_statements:
            detail = (
                f"{detail} excluded_postgres_statements={summary.excluded_statements} "
                f"excluded_postgres_files={len(summary.excluded_files)}"
            )
        print(detail)
        if summary.percentage < budget:
            failures.append(f"{package}: {summary.percentage:.1f}% < {budget:.1f}%")

    if failures:
        raise SystemExit("platform core coverage budget failed: " + "; ".join(failures))
    print("platform core coverage budgets passed.")


if __name__ == "__main__":
    try:
        main()
    except BrokenPipeError:
        sys.exit(1)

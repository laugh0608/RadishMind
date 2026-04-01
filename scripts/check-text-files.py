#!/usr/bin/env python3
from __future__ import annotations

import re
import subprocess
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parent.parent
TEXT_EXTENSIONS = {".md", ".txt", ".json", ".yml", ".yaml", ".ps1", ".sh", ".py"}
TEXT_NAMES = {".gitignore", ".gitattributes", ".editorconfig"}
TRAILING_WHITESPACE = re.compile(r"[ \t]+$")


def is_text_file(relative_path: str) -> bool:
    path = Path(relative_path)
    if path.name in TEXT_NAMES:
        return True
    return path.suffix.lower() in TEXT_EXTENSIONS


def iter_tracked_files() -> list[str]:
    result = subprocess.run(
        ["git", "-C", str(REPO_ROOT), "ls-files", "--cached", "--others", "--exclude-standard", "-z"],
        check=True,
        capture_output=True,
    )
    return [item.decode("utf-8") for item in result.stdout.split(b"\x00") if item]


def main() -> int:
    violations: list[str] = []

    for relative_path in iter_tracked_files():
        if not is_text_file(relative_path):
            continue

        full_path = REPO_ROOT / relative_path
        if not full_path.is_file():
            continue

        data = full_path.read_bytes()
        if data.startswith(b"\xef\xbb\xbf"):
            violations.append(f"{relative_path}: contains UTF-8 BOM")

        try:
            text = data.decode("utf-8")
        except UnicodeDecodeError:
            violations.append(f"{relative_path}: is not valid UTF-8")
            continue

        if "\r" in text:
            violations.append(f"{relative_path}: contains CRLF or carriage returns")

        if data and data[-1] != 0x0A:
            violations.append(f"{relative_path}: is missing a final newline")

        if full_path.suffix.lower() == ".md":
            continue

        lines = text.split("\n")
        for index, line in enumerate(lines, start=1):
            if index == len(lines) and line == "":
                continue
            if TRAILING_WHITESPACE.search(line):
                violations.append(f"{relative_path}:{index}: trailing whitespace")

    if violations:
        for violation in violations:
            print(violation, file=sys.stderr)
        return 1

    print("text file hygiene checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

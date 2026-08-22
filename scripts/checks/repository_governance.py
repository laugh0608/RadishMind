from __future__ import annotations

import re
import subprocess
from pathlib import Path
from urllib.parse import unquote, urlsplit


MARKDOWN_LINK = re.compile(
    r"!?\[[^\]\n]*\]\(\s*(?:<(?P<angle>[^>\n]+)>|(?P<plain>[^)\s]+))"
    r"(?:\s+(?:\"[^\"]*\"|'[^']*'|\([^)]*\)))?\s*\)"
)
INLINE_CODE = re.compile(r"`+[^`\n]*`+")
FENCE_START = re.compile(r"(?P<marker>`{3,}|~{3,})")


def repository_paths(repo_root: Path) -> list[Path]:
    result = subprocess.run(
        [
            "git",
            "-C",
            str(repo_root),
            "ls-files",
            "--cached",
            "--others",
            "--exclude-standard",
            "-z",
        ],
        check=True,
        capture_output=True,
    )
    return sorted(
        (Path(item.decode("utf-8")) for item in result.stdout.split(b"\0") if item),
        key=Path.as_posix,
    )


def markdown_prose(text: str) -> str:
    prose: list[str] = []
    fence_character = ""
    fence_length = 0

    for line in text.splitlines(keepends=True):
        stripped = line.lstrip()
        marker_match = FENCE_START.match(stripped)
        if marker_match:
            marker = marker_match.group("marker")
            if not fence_character:
                fence_character = marker[0]
                fence_length = len(marker)
                prose.append("\n")
                continue
            if marker[0] == fence_character and len(marker) >= fence_length:
                fence_character = ""
                fence_length = 0
                prose.append("\n")
                continue

        if fence_character:
            prose.append("\n")
            continue
        prose.append(INLINE_CODE.sub("", line))

    return "".join(prose)


def find_markdown_link_errors(repo_root: Path, paths: list[Path]) -> list[str]:
    resolved_root = repo_root.resolve()
    errors: list[str] = []

    for relative_path in paths:
        if relative_path.suffix.lower() != ".md":
            continue
        source_path = repo_root / relative_path
        if not source_path.is_file():
            continue

        for match in MARKDOWN_LINK.finditer(markdown_prose(source_path.read_text(encoding="utf-8"))):
            raw_target = match.group("angle") or match.group("plain") or ""
            target = unquote(raw_target.strip())
            if not target or target.startswith(("#", "/", "//")):
                continue

            parsed_target = urlsplit(target)
            if parsed_target.scheme:
                continue
            target_path = parsed_target.path
            if not target_path:
                continue

            resolved_target = (source_path.parent / target_path).resolve()
            try:
                resolved_target.relative_to(resolved_root)
            except ValueError:
                errors.append(
                    f"relative link escapes repository: {relative_path.as_posix()} -> {target_path}"
                )
                continue
            if not resolved_target.exists():
                errors.append(
                    f"broken relative link: {relative_path.as_posix()} -> {target_path}"
                )

    return errors


def check_markdown_links(repo_root: Path, paths: list[Path] | None = None) -> None:
    repository_file_paths = repository_paths(repo_root) if paths is None else paths
    errors = find_markdown_link_errors(repo_root, repository_file_paths)
    if errors:
        raise SystemExit("\n".join(errors))
    print("markdown relative link checks passed.")

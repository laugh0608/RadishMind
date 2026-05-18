#!/usr/bin/env python3
from __future__ import annotations

import json
import subprocess
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
CONSOLE_ROOT = REPO_ROOT / "apps/radishmind-console"
PACKAGE_JSON = CONSOLE_ROOT / "package.json"
README = CONSOLE_ROOT / "README.md"
GITIGNORE = REPO_ROOT / ".gitignore"
CURRENT_FOCUS = REPO_ROOT / "docs/radishmind-current-focus.md"
ROADMAP = REPO_ROOT / "docs/radishmind-roadmap.md"
PROJECT_GUIDE = REPO_ROOT / "docs/radishmind-project-guide.md"

ALLOWED_NPM_SCRIPTS = {"dev", "build", "preview"}
FORBIDDEN_NPM_SCRIPT_NAMES = {
    "deploy",
    "publish",
    "release",
    "package",
    "pack",
    "docker",
    "serve:prod",
    "start:prod",
}
FORBIDDEN_NPM_SCRIPT_TOKENS = (
    "npm publish",
    "vite preview --host 0.0.0.0",
    "wrangler",
    "vercel",
    "netlify",
    "docker build",
    "scp ",
    "rsync ",
)
FORBIDDEN_TRACKED_PATH_PARTS = (
    ("apps", "radishmind-console", "dist"),
    ("apps", "radishmind-console", "node_modules"),
)
REQUIRED_BOUNDARY_LITERALS = (
    "console production packaging",
    "不代表生产鉴权或公开部署已完成",
    "不生成、提交或发布 production package",
    "不添加 deploy / publish / release 脚本",
)


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_package() -> dict[str, Any]:
    document = json.loads(PACKAGE_JSON.read_text(encoding="utf-8"))
    require(isinstance(document, dict), "console package.json must be a JSON object")
    return document


def git_tracked_files() -> list[Path]:
    result = subprocess.run(
        ["git", "ls-files", "-z", "apps/radishmind-console"],
        cwd=REPO_ROOT,
        capture_output=True,
        text=False,
        check=True,
    )
    return [Path(raw.decode("utf-8")) for raw in result.stdout.split(b"\0") if raw]


def path_has_parts(path: Path, expected_parts: tuple[str, ...]) -> bool:
    return path.parts[: len(expected_parts)] == expected_parts


def check_package_boundary(package: dict[str, Any]) -> None:
    require(package.get("private") is True, "RadishMind console must remain a private package")
    require(package.get("name") == "@radishmind/console", "RadishMind console package name drifted")
    require(package.get("type") == "module", "RadishMind console package must stay ESM")
    require("files" not in package, "console package must not define npm package files before packaging is designed")
    require("publishConfig" not in package, "console package must not define publishConfig before packaging is designed")
    require("bin" not in package, "console package must not expose CLI binaries before packaging is designed")

    scripts = package.get("scripts") or {}
    require(isinstance(scripts, dict), "console package scripts must be an object")
    unexpected_scripts = sorted(set(scripts) - ALLOWED_NPM_SCRIPTS)
    if unexpected_scripts:
        raise SystemExit(f"console package has unexpected npm script: {unexpected_scripts[0]}")
    for script_name in FORBIDDEN_NPM_SCRIPT_NAMES:
        require(script_name not in scripts, f"console package must not define production script: {script_name}")
    for script_name, script_command in scripts.items():
        command = str(script_command)
        for token in FORBIDDEN_NPM_SCRIPT_TOKENS:
            require(
                token not in command,
                f"console npm script {script_name} contains forbidden production token: {token}",
            )
    require("vite --host 127.0.0.1" in str(scripts.get("dev", "")), "console dev script must bind localhost")
    require("tsc --noEmit" in str(scripts.get("build", "")), "console build script must typecheck")


def check_git_boundary() -> None:
    tracked_files = git_tracked_files()
    for tracked_file in tracked_files:
        for forbidden_parts in FORBIDDEN_TRACKED_PATH_PARTS:
            require(
                not path_has_parts(tracked_file, forbidden_parts),
                f"console production artifact must not be tracked: {tracked_file.as_posix()}",
            )

    gitignore = GITIGNORE.read_text(encoding="utf-8")
    require("node_modules/" in gitignore, ".gitignore must ignore frontend dependencies")
    require(
        "apps/radishmind-console/dist/" in gitignore,
        ".gitignore must ignore RadishMind console build output",
    )


def check_documented_boundary() -> None:
    readme = README.read_text(encoding="utf-8")
    for literal in REQUIRED_BOUNDARY_LITERALS:
        require(literal in readme, f"console README missing production boundary literal: {literal}")

    for doc_path in (CURRENT_FOCUS, ROADMAP, PROJECT_GUIDE):
        content = doc_path.read_text(encoding="utf-8")
        require(
            "console production packaging" in content,
            f"{doc_path.relative_to(REPO_ROOT).as_posix()} must keep console production packaging as a gap",
        )


def main() -> int:
    check_package_boundary(load_package())
    check_git_boundary()
    check_documented_boundary()
    print("platform local console production boundary check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

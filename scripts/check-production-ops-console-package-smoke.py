#!/usr/bin/env python3
from __future__ import annotations

import json
import subprocess
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-console-package-smoke.json"
PACKAGE_JSON = REPO_ROOT / "apps/radishmind-console/package.json"
PACKAGE_LOCK = REPO_ROOT / "apps/radishmind-console/package-lock.json"
GITIGNORE = REPO_ROOT / ".gitignore"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

REQUIRED_PACKAGE_ENTRIES = {
    "build": ("npm run build", "tsc --noEmit && vite build", "not_a_published_package"),
    "preview": ("npm run preview", "vite preview --host 127.0.0.1 --port 4000", "not_production_hosting"),
}

REQUIRED_ALLOWED_SCRIPTS = {"dev", "build", "preview"}
REQUIRED_FORBIDDEN_SCRIPTS = {
    "deploy",
    "publish",
    "release",
    "package",
    "pack",
    "docker",
    "serve:prod",
    "start:prod",
}

REQUIRED_FORBIDDEN_TRACKED_PATHS = {
    "apps/radishmind-console/dist/",
    "apps/radishmind-console/node_modules/",
}

REQUIRED_IGNORED_PATHS = {
    "node_modules/",
    "apps/radishmind-console/dist/",
}

REQUIRED_BLOCKED_CONDITIONS = {
    "console_production_packaging",
    "production_deploy_target",
    "production_auth_and_cors_policy",
    "production_static_asset_policy",
}

REQUIRED_FORBIDDEN_INTERPRETATIONS = {
    "npm run build is production package",
    "npm run preview is production hosting",
    "private package is publishable",
    "local dist output is committed artifact",
    "localhost preview is production deployment",
}

REQUIRED_DOC_REFERENCES = {
    "apps/radishmind-console/README.md": [
        "Production packaging 边界",
        "npm run build",
        "npm run preview",
        "不是 production package",
        "不生成、提交或发布 production package",
    ],
    "docs/task-cards/production-ops-hardening-v1-plan.md": [
        "console-production-package-smoke",
        "production-ops-console-package-smoke.json",
        "check-production-ops-console-package-smoke.py",
    ],
    "scripts/README.md": [
        "check-production-ops-console-package-smoke.py",
        "production-ops-console-package-smoke.json",
    ],
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT).as_posix()} must be a JSON object")
    return document


def load_fixture() -> dict[str, Any]:
    return load_json(FIXTURE_PATH)


def assert_slice(document: dict[str, Any]) -> None:
    slice_info = document.get("slice") or {}
    require(slice_info.get("id") == "console-production-package-smoke", "unexpected production ops slice id")
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected production ops track")
    require(
        slice_info.get("status") == "governance_boundary_satisfied",
        "console-production-package-smoke must only satisfy the governance boundary",
    )
    forbidden_claims = set(slice_info.get("does_not_claim") or [])
    required_forbidden_claims = {
        "production_ready",
        "console_production_package_ready",
        "production_deployment_ready",
        "production_auth_policy_ready",
        "public_hosting_ready",
    }
    missing = sorted(required_forbidden_claims - forbidden_claims)
    require(not missing, f"missing forbidden console packaging claims: {missing}")


def assert_package_smoke_entries(document: dict[str, Any]) -> None:
    entries = {
        str(item.get("id")): item
        for item in document.get("package_smoke_entries") or []
        if isinstance(item, dict)
    }
    missing = sorted(set(REQUIRED_PACKAGE_ENTRIES) - set(entries))
    require(not missing, f"missing package smoke entries: {missing}")
    for entry_id, (command, expected_script, production_use) in REQUIRED_PACKAGE_ENTRIES.items():
        entry = entries[entry_id]
        require(entry.get("command") == command, f"{entry_id} command drifted")
        require(entry.get("expected_script") == expected_script, f"{entry_id} expected script drifted")
        require(entry.get("production_use") == production_use, f"{entry_id} production use drifted")
        require(entry.get("artifact_policy"), f"{entry_id} must document artifact policy")


def assert_package_boundaries(document: dict[str, Any]) -> None:
    boundaries = document.get("package_boundaries") or {}
    require(boundaries.get("package_name") == "@radishmind/console", "fixture package name drifted")
    require(boundaries.get("private") is True, "fixture must keep private package boundary")
    allowed_scripts = set(boundaries.get("allowed_scripts") or [])
    forbidden_scripts = set(boundaries.get("forbidden_scripts") or [])
    forbidden_paths = set(boundaries.get("forbidden_tracked_paths") or [])
    ignored_paths = set(boundaries.get("required_ignored_paths") or [])
    require(REQUIRED_ALLOWED_SCRIPTS <= allowed_scripts, "fixture missing allowed scripts")
    require(REQUIRED_FORBIDDEN_SCRIPTS <= forbidden_scripts, "fixture missing forbidden scripts")
    require(REQUIRED_FORBIDDEN_TRACKED_PATHS <= forbidden_paths, "fixture missing forbidden tracked paths")
    require(REQUIRED_IGNORED_PATHS <= ignored_paths, "fixture missing required ignored paths")


def assert_package_json() -> None:
    package = load_json(PACKAGE_JSON)
    require(package.get("name") == "@radishmind/console", "console package name drifted")
    require(package.get("private") is True, "console package must remain private")
    require(package.get("type") == "module", "console package must stay ESM")
    require("files" not in package, "console package must not define package files before packaging is designed")
    require("publishConfig" not in package, "console package must not define publishConfig")
    require("bin" not in package, "console package must not expose CLI binaries")

    scripts = package.get("scripts") or {}
    require(isinstance(scripts, dict), "console package scripts must be an object")
    require(set(scripts) == REQUIRED_ALLOWED_SCRIPTS, f"console package scripts drifted: {sorted(scripts)}")
    require(scripts.get("build") == "tsc --noEmit && vite build", "console build script drifted")
    require(scripts.get("preview") == "vite preview --host 127.0.0.1 --port 4000", "console preview script drifted")
    for script_name in REQUIRED_FORBIDDEN_SCRIPTS:
        require(script_name not in scripts, f"console package must not define production script: {script_name}")


def assert_package_lock() -> None:
    lockfile = load_json(PACKAGE_LOCK)
    root_package = (lockfile.get("packages") or {}).get("") or {}
    require(root_package.get("name") == "@radishmind/console", "console lockfile package name drifted")
    require(root_package.get("version") == "0.1.0", "console lockfile version drifted")
    require(root_package.get("dependencies"), "console lockfile must preserve dependencies")
    require(root_package.get("devDependencies"), "console lockfile must preserve devDependencies")


def assert_git_boundaries() -> None:
    gitignore = GITIGNORE.read_text(encoding="utf-8")
    for ignored_path in REQUIRED_IGNORED_PATHS:
        require(ignored_path in gitignore, f".gitignore missing required ignored path: {ignored_path}")

    result = subprocess.run(
        ["git", "ls-files", "-z", "apps/radishmind-console"],
        cwd=REPO_ROOT,
        capture_output=True,
        text=False,
        check=True,
    )
    tracked_paths = [raw.decode("utf-8") for raw in result.stdout.split(b"\0") if raw]
    for tracked_path in tracked_paths:
        normalized = tracked_path.replace("\\", "/")
        for forbidden_path in REQUIRED_FORBIDDEN_TRACKED_PATHS:
            require(
                not normalized.startswith(forbidden_path),
                f"console production artifact must not be tracked: {normalized}",
            )


def assert_blocked_and_forbidden(document: dict[str, Any]) -> None:
    blocked = {str(item.get("id")): item for item in document.get("blocked_conditions") or [] if isinstance(item, dict)}
    missing_blocked = sorted(REQUIRED_BLOCKED_CONDITIONS - set(blocked))
    require(not missing_blocked, f"missing blocked console packaging conditions: {missing_blocked}")
    for condition_id in REQUIRED_BLOCKED_CONDITIONS:
        item = blocked[condition_id]
        require(item.get("status") == "not_satisfied", f"{condition_id} must remain not_satisfied")
        require(
            item.get("required_before_production_ready") is True,
            f"{condition_id} must gate production ready",
        )

    forbidden = set(document.get("forbidden_interpretations") or [])
    missing_forbidden = sorted(REQUIRED_FORBIDDEN_INTERPRETATIONS - forbidden)
    require(not missing_forbidden, f"missing forbidden interpretations: {missing_forbidden}")


def assert_evidence_and_consumers(document: dict[str, Any]) -> None:
    for evidence in document.get("evidence") or []:
        evidence_path = REPO_ROOT / str(evidence)
        require(evidence_path.exists(), f"missing console package smoke evidence path: {evidence}")

    consumers = set(document.get("required_consumers") or [])
    expected_consumers = {
        "scripts/check-production-ops-console-package-smoke.py",
        "scripts/check-repo.py",
        "scripts/check-p3-local-product-shell-short-close-checklist.py",
        "scripts/README.md",
    }
    missing_consumers = sorted(expected_consumers - consumers)
    require(not missing_consumers, f"missing console package smoke consumers: {missing_consumers}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("check-production-ops-console-package-smoke.py", [])' in check_repo,
        "check-repo.py must run production ops console package smoke check",
    )


def assert_doc_references() -> None:
    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        content = (REPO_ROOT / relative_path).read_text(encoding="utf-8")
        missing = [literal for literal in required_literals if literal not in content]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> int:
    document = load_fixture()
    require(document.get("schema_version") == 1, "unexpected schema_version")
    require(document.get("kind") == "production_ops_console_package_smoke", "unexpected fixture kind")
    assert_slice(document)
    assert_package_smoke_entries(document)
    assert_package_boundaries(document)
    assert_package_json()
    assert_package_lock()
    assert_git_boundaries()
    assert_blocked_and_forbidden(document)
    assert_evidence_and_consumers(document)
    assert_doc_references()
    print("production ops console package smoke checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

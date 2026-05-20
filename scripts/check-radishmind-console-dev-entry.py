#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parent.parent
DEV_ENTRY = REPO_ROOT / "scripts/run-radishmind-console-dev.ps1"
CHECK_REPO = REPO_ROOT / "scripts/check-repo.py"
SCRIPTS_README = REPO_ROOT / "scripts/README.md"
CONSOLE_README = REPO_ROOT / "apps/radishmind-console/README.md"
PLATFORM_README = REPO_ROOT / "services/platform/README.md"
CURRENT_FOCUS = REPO_ROOT / "docs/radishmind-current-focus.md"
P3_CHECKLIST = REPO_ROOT / "scripts/checks/fixtures/p3-local-product-shell-short-close-checklist.json"

REQUIRED_SCRIPT_LITERALS = (
    "http://127.0.0.1:7000",
    "http://127.0.0.1:4000",
    "/healthz",
    "/v1/platform/overview",
    "run-platform-service.ps1",
    "npm",
    "run",
    "dev",
    "Access-Control-Allow-Origin",
    "ERR_UNSAFE_PORT",
    "Port conflict",
    "CORS/preflight",
    "Browser unsafe port",
    "VerifyOnly",
    "dev-only launcher, not a production supervisor",
    "executor, durable store, confirmation, writeback, or replay",
    "RADISHMIND_PLATFORM_LISTEN_ADDR",
    "VITE_RADISHMIND_PLATFORM_BASE_URL",
    "Start-Process",
    "-WindowStyle Hidden",
)

FORBIDDEN_SCRIPT_LITERALS = (
    "deploy",
    "publish",
    "release",
    "docker",
    "restart",
    "supervisor",
    "durable_store_enabled = $true",
    "confirmation_flow_connected = $true",
    "business_truth_write_enabled = $true",
    "automatic_replay_enabled = $true",
)

REQUIRED_DOC_LITERALS = (
    "scripts/run-radishmind-console-dev.ps1",
    "http://127.0.0.1:7000/healthz",
    "http://127.0.0.1:7000/v1/platform/overview",
    "http://127.0.0.1:4000",
    "端口冲突",
    "CORS",
    "unsafe port",
    "不是 production supervisor",
    "不实现真实 executor、durable store、confirmation、业务写回或 replay",
)


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def require_all(content: str, literals: tuple[str, ...], path: Path) -> None:
    missing = [literal for literal in literals if literal not in content]
    require(not missing, f"{path.relative_to(REPO_ROOT).as_posix()} missing literals: {missing}")


def main() -> int:
    require(DEV_ENTRY.is_file(), "missing RadishMind console dev entry PowerShell script")
    script = DEV_ENTRY.read_text(encoding="utf-8")
    require_all(script, REQUIRED_SCRIPT_LITERALS, DEV_ENTRY)
    for literal in FORBIDDEN_SCRIPT_LITERALS:
        if literal == "supervisor":
            require(script.count(literal) == 1, "dev entry may only mention supervisor to reject production scope")
        else:
            require(literal not in script, f"dev entry contains forbidden production literal: {literal}")

    for doc_path in (SCRIPTS_README, CONSOLE_README, PLATFORM_README, CURRENT_FOCUS):
        require_all(doc_path.read_text(encoding="utf-8"), REQUIRED_DOC_LITERALS, doc_path)

    check_repo = CHECK_REPO.read_text(encoding="utf-8")
    require(
        'run_python_script("check-radishmind-console-dev-entry.py", [])' in check_repo,
        "check-repo.py must run the local console dev entry gate",
    )

    checklist = P3_CHECKLIST.read_text(encoding="utf-8")
    require("local_console_dev_entry" in checklist, "P3 checklist must include local console dev entry evidence")
    require(
        "scripts/check-radishmind-console-dev-entry.py" in checklist,
        "P3 checklist must include dev entry check evidence",
    )

    print("platform local console dev entry check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

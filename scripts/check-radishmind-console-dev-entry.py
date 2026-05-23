#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parent.parent
DEV_ENTRY_PS1 = REPO_ROOT / "scripts/run-radishmind-console-dev.ps1"
DEV_ENTRY_SH = REPO_ROOT / "scripts/run-radishmind-console-dev.sh"
CHECK_REPO = REPO_ROOT / "scripts/check-repo.py"
SCRIPTS_README = REPO_ROOT / "scripts/README.md"
CONSOLE_README = REPO_ROOT / "apps/radishmind-console/README.md"
PLATFORM_README = REPO_ROOT / "services/platform/README.md"
CURRENT_FOCUS = REPO_ROOT / "docs/radishmind-current-focus.md"
P3_CHECKLIST = REPO_ROOT / "scripts/checks/fixtures/p3-local-product-shell-short-close-checklist.json"

REQUIRED_COMMON_SCRIPT_LITERALS = (
    "http://127.0.0.1:7000",
    "http://127.0.0.1:4000",
    "/healthz",
    "/v1/platform/overview",
    "/v1/platform/local-smoke",
    "platform_local_smoke",
    "npm",
    "run",
    "dev",
    "Access-Control-Allow-Origin",
    "ERR_UNSAFE_PORT",
    "Port conflict",
    "CORS/preflight",
    "Browser unsafe port",
    "dev-only launcher, not a production supervisor",
    "executor, durable store, confirmation, writeback, or replay",
    "RADISHMIND_PLATFORM_LISTEN_ADDR",
    "VITE_RADISHMIND_PLATFORM_BASE_URL",
)

REQUIRED_PS1_LITERALS = (
    "run-platform-service.ps1",
    "Start-Process",
    "-WindowStyle Hidden",
    "VerifyOnly",
    "ExitAfterProbe",
)

REQUIRED_SH_LITERALS = (
    "run-platform-service.sh",
    "npm run dev",
    "--verify-only",
    "--exit-after-probe",
    "trap cleanup EXIT INT TERM",
    "Access-Control-Allow-Origin",
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
    "scripts/run-radishmind-console-dev.sh",
    "http://127.0.0.1:7000/healthz",
    "http://127.0.0.1:7000/v1/platform/overview",
    "http://127.0.0.1:7000/v1/platform/local-smoke",
    "http://127.0.0.1:4000",
    "端口冲突",
    "CORS",
    "unsafe port",
    "不是 production supervisor",
    "不实现真实 executor、durable store、confirmation、业务写回或 replay",
)

REQUIRED_CURRENT_FOCUS_LITERALS = (
    "local usable / read-only close",
    "不再默认继续补同类只读 console 小切片",
    "process supervisor",
    "production secret backend、process supervisor、部署环境隔离和 console production packaging 仍为 `not_satisfied`",
)


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def require_all(content: str, literals: tuple[str, ...], path: Path) -> None:
    missing = [literal for literal in literals if literal not in content]
    require(not missing, f"{path.relative_to(REPO_ROOT).as_posix()} missing literals: {missing}")


def main() -> int:
    require(DEV_ENTRY_PS1.is_file(), "missing RadishMind console dev entry PowerShell script")
    require(DEV_ENTRY_SH.is_file(), "missing RadishMind console dev entry shell script")

    ps1_script = DEV_ENTRY_PS1.read_text(encoding="utf-8")
    sh_script = DEV_ENTRY_SH.read_text(encoding="utf-8")
    for script_path, script, required_literals in (
        (DEV_ENTRY_PS1, ps1_script, REQUIRED_COMMON_SCRIPT_LITERALS + REQUIRED_PS1_LITERALS),
        (DEV_ENTRY_SH, sh_script, REQUIRED_COMMON_SCRIPT_LITERALS + REQUIRED_SH_LITERALS),
    ):
        require_all(script, required_literals, script_path)
        for literal in FORBIDDEN_SCRIPT_LITERALS:
            if literal == "supervisor":
                require(
                    script.count(literal) == 1,
                    f"{script_path.relative_to(REPO_ROOT).as_posix()} may only mention supervisor to reject production scope",
                )
            else:
                require(literal not in script, f"{script_path.relative_to(REPO_ROOT).as_posix()} contains forbidden production literal: {literal}")

    for doc_path in (SCRIPTS_README, CONSOLE_README, PLATFORM_README):
        require_all(doc_path.read_text(encoding="utf-8"), REQUIRED_DOC_LITERALS, doc_path)
    require_all(CURRENT_FOCUS.read_text(encoding="utf-8"), REQUIRED_CURRENT_FOCUS_LITERALS, CURRENT_FOCUS)

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
    require(
        "scripts/run-radishmind-console-dev.sh" in checklist,
        "P3 checklist must include shell dev entry evidence",
    )

    print("platform local console dev entry check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

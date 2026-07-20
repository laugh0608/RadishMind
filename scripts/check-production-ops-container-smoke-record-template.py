#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-container-smoke-record-template.json"
RUNBOOK_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-container-smoke-runbook.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
FORBIDDEN_LOCAL_REFERENCE = "D:\\Code\\Radish"

REQUIRED_FIELDS = {
    "schema_version",
    "kind",
    "run_id",
    "run_started_at",
    "run_finished_at",
    "operator",
    "git_commit",
    "environment",
    "compose_file",
    "provider",
    "commands",
    "probes",
    "containers",
    "cleanup",
    "result",
    "blocked_conditions_after_run",
}

REQUIRED_DOC_REFERENCES = {
    "docs/task-cards/production-ops-docker-deployment-v1-plan.md": [
        "container-smoke-record-template",
        "production-ops-container-smoke-record-template.json",
        "tmp/production-ops/container-smoke/",
        "container_smoke_ready",
        "production ready",
    ],
    "docs/task-cards/production-ops-hardening-v1-plan.md": [
        "container-smoke-record-template",
        "production-ops-container-smoke-record-template.json",
    ],
    "scripts/README.md": [
        "check-production-ops-container-smoke-record-template.py",
        "production-ops-container-smoke-record-template.json",
        "tmp/production-ops/container-smoke/",
    ],
    "docs/devlogs/2026-W21.md": [
        "container-smoke-record-template",
        "production-ops-container-smoke-record-template.json",
    ],
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def read(relative_path: str) -> str:
    return (REPO_ROOT / relative_path).read_text(encoding="utf-8")


def load_json(path: Path) -> dict[str, Any]:
    return json.loads(path.read_text(encoding="utf-8"))


def assert_fixture(fixture: dict[str, Any]) -> None:
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "container-smoke-record-template", "unexpected slice id")
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected slice track")
    require(slice_info.get("status") == "record_template_satisfied", "record template must be satisfied")
    does_not_claim = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "container_smoke_ready",
        "test_environment_smoke_ready",
        "production_preflight_ready",
        "production_ready",
        "image_publish_ready",
        "production_secret_backend_ready",
        "process_supervisor_ready",
        "production_auth_cors_policy_ready",
        "console_runtime_config_ready",
    }:
        require(claim in does_not_claim, f"record template must not claim {claim}")

    scope = fixture.get("record_scope") or {}
    require(scope.get("records_future_runtime_evidence_only") is True, "template must only describe future evidence")
    require(scope.get("committed_template_is_not_a_run_record") is True, "template must not be a run record")
    require(scope.get("fast_baseline_executes_docker") is False, "fast baseline must not execute Docker")
    require(scope.get("requires_explicit_human_run_window") is True, "template must require run window")
    require(scope.get("allowed_record_root") == "tmp/production-ops/container-smoke/", "unexpected record root")
    require(scope.get("committed_run_records_forbidden") is True, "runtime records must stay uncommitted")

    fields = set(fixture.get("required_record_fields") or [])
    missing_fields = sorted(REQUIRED_FIELDS - fields)
    require(not missing_fields, f"missing record fields: {missing_fields}")


def assert_record_shape(fixture: dict[str, Any]) -> None:
    record = fixture.get("local_container_smoke_record") or {}
    require(record.get("kind") == "production_ops_container_smoke_record", "unexpected record kind")
    require(record.get("environment") == "docker_local", "record template must target docker_local")
    require(record.get("compose_file") == "deploy/docker-compose.local.yaml", "unexpected compose path")
    require(record.get("provider") == "mock", "local record template must use mock provider")

    runbook = load_json(RUNBOOK_FIXTURE_PATH)
    runbook_commands = set((runbook.get("local_container_smoke") or {}).get("commands") or [])
    template_commands = {
        str(item.get("command"))
        for item in record.get("expected_commands") or []
        if item.get("required") is True
    }
    missing_commands = sorted(runbook_commands - template_commands)
    require(not missing_commands, f"record template missing runbook commands: {missing_commands}")

    runbook_probe_urls = set((runbook.get("local_container_smoke") or {}).get("probe_urls") or [])
    template_probe_urls = {str(item.get("url")) for item in record.get("expected_probes") or []}
    missing_probe_urls = sorted(runbook_probe_urls - template_probe_urls)
    require(not missing_probe_urls, f"record template missing runbook probes: {missing_probe_urls}")

    cleanup = record.get("required_cleanup") or {}
    require(
        cleanup.get("command") == "docker compose -f deploy/docker-compose.local.yaml down",
        "record template must require compose down cleanup",
    )
    require(cleanup.get("must_record_exit_code") is True, "cleanup must record exit code")
    require(
        cleanup.get("must_record_remaining_project_containers") is True,
        "cleanup must record remaining project containers",
    )


def assert_result_policy(fixture: dict[str, Any]) -> None:
    policy = fixture.get("result_policy") or {}
    may_satisfy = set(policy.get("passing_local_record_may_satisfy") or [])
    require(may_satisfy == {"container_smoke_ready"}, "local record may only satisfy container_smoke_ready")
    must_not = set(policy.get("passing_local_record_must_not_satisfy") or [])
    for condition_id in {
        "test_environment_smoke_ready",
        "production_preflight_ready",
        "production_ready",
        "image_publish_ready",
        "production_secret_backend_ready",
        "process_supervisor_ready",
        "production_auth_cors_policy_ready",
        "console_runtime_config_ready",
    }:
        require(condition_id in must_not, f"passing local record must not satisfy {condition_id}")

    blocked = {str(item.get("id")): item for item in fixture.get("blocked_conditions") or []}
    for condition_id in {"container_smoke_ready", "test_environment_smoke", "production_preflight_record"}:
        require(condition_id in blocked, f"missing blocked condition: {condition_id}")
        require(blocked[condition_id].get("status") == "not_satisfied", f"{condition_id} must stay not_satisfied")


def assert_consumers_and_docs(fixture: dict[str, Any]) -> None:
    required_consumers = set(fixture.get("required_consumers") or [])
    expected = {
        "scripts/check-production-ops-container-smoke-record-template.py",
        "scripts/check-production-ops-docker-deployment-mode.py",
        "scripts/check-repo.py",
        "scripts/README.md",
        "docs/task-cards/production-ops-hardening-v1-plan.md",
        "docs/task-cards/production-ops-docker-deployment-v1-plan.md",
        "docs/devlogs/2026-W21.md",
    }
    missing = sorted(expected - required_consumers)
    require(not missing, f"missing consumers: {missing}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("check-production-ops-container-smoke-record-template.py", [])' in check_repo,
        "check-repo.py must run container smoke record template check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing_literals = [literal for literal in required_literals if literal not in text]
        require(not missing_literals, f"{relative_path} missing literals: {missing_literals}")
        require(FORBIDDEN_LOCAL_REFERENCE not in text, f"{relative_path} must not commit local Radish path")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "production_ops_container_smoke_record_template", "unexpected fixture kind")
    assert_fixture(fixture)
    assert_record_shape(fixture)
    assert_result_policy(fixture)
    assert_consumers_and_docs(fixture)
    require(FORBIDDEN_LOCAL_REFERENCE not in FIXTURE_PATH.read_text(encoding="utf-8"), "fixture must not commit local path")
    print("production ops container smoke record template checks passed.")


if __name__ == "__main__":
    main()

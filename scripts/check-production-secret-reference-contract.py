#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any

try:
    import jsonschema
except ModuleNotFoundError:  # pragma: no cover - fallback for minimal local Python launchers.
    jsonschema = None


REPO_ROOT = Path(__file__).resolve().parent.parent
SCHEMA_PATH = REPO_ROOT / "contracts/production-secret-reference.schema.json"
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/production-secret-reference-basic.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
READINESS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
)

REQUIRED_REQUIRED_FIELDS = {
    "environment",
    "provider",
    "provider_profile",
    "secret_ref",
    "required_fields",
    "sanitized_fields",
}

REQUIRED_SANITIZED_FIELDS = {
    "credential_state",
    "secret_backend_configured",
    "secret_ref_present",
    "missing_secret_refs",
    "field_sources",
}

REQUIRED_FORBIDDEN_FIELDS = {
    "api_key",
    "token",
    "cookie",
    "authorization",
    "credential",
    "secret_value",
    "provider_base_url",
}

REQUIRED_DOC_REFERENCES = {
    "docs/task-cards/production-secret-backend-implementation-v1-plan.md": [
        "secret-ref-schema-and-fixtures",
        "contracts/production-secret-reference.schema.json",
        "production-secret-reference-basic.json",
        "check-production-secret-reference-contract.py",
        "不保存 secret value",
    ],
    "docs/radishmind-roadmap.md": [
        "secret-ref-schema-and-fixtures",
        "production-secret-reference.schema.json",
    ],
    "docs/radishmind-capability-matrix.md": [
        "secret ref schema",
        "production-secret-reference.schema.json",
    ],
    "services/platform/README.md": [
        "Production secret reference schema",
        "production-secret-reference.schema.json",
        "production-secret-reference-basic.json",
    ],
    "contracts/README.md": [
        "production-secret-reference.schema.json",
        "production-secret-reference-basic.json",
        "reference-only manifest",
    ],
    "docs/contracts/README.md": [
        "Production Secret Reference 契约",
        "production-secret-reference.md",
    ],
    "docs/contracts/production-secret-reference.md": [
        "Production Secret Reference 契约",
        "contracts/production-secret-reference.schema.json",
        "scripts/checks/fixtures/production-secret-reference-basic.json",
        "不保存 secret value",
        "不声明 production secret backend ready",
    ],
    "docs/radishmind-integration-contracts.md": [
        "Production Secret Reference",
        "production-secret-reference.schema.json",
        "reference-only manifest",
    ],
    "deploy/README.md": [
        "production-secret-reference.schema.json",
        "production-secret-reference-basic.json",
        "reference-only manifest",
    ],
    "scripts/README.md": [
        "check-production-secret-reference-contract.py",
        "production-secret-reference.schema.json",
    ],
    "docs/devlogs/2026-W22.md": [
        "secret-ref-schema-and-fixtures",
        "production-secret-reference.schema.json",
    ],
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def read(relative_path: str) -> str:
    return (REPO_ROOT / relative_path).read_text(encoding="utf-8")


def load_json(path: Path) -> Any:
    return json.loads(path.read_text(encoding="utf-8"))


def assert_manifest(manifest: dict[str, Any]) -> None:
    require(manifest.get("kind") == "production_secret_reference_manifest", "unexpected secret reference kind")
    require(manifest.get("scope") == "secret_reference_only", "secret reference scope must stay reference-only")

    bindings = manifest.get("bindings") or []
    require(isinstance(bindings, list) and len(bindings) >= 2, "secret reference fixture must include test and production bindings")
    environments = {str(item.get("environment")) for item in bindings if isinstance(item, dict)}
    require({"test", "production"}.issubset(environments), "fixture must cover test and production environments")

    for binding in bindings:
        required_fields = set(binding.get("required_fields") or [])
        missing_required = sorted(REQUIRED_REQUIRED_FIELDS - required_fields)
        require(not missing_required, f"binding missing required field declarations: {missing_required}")

        sanitized_fields = set(binding.get("sanitized_fields") or [])
        missing_sanitized = sorted(REQUIRED_SANITIZED_FIELDS - sanitized_fields)
        require(not missing_sanitized, f"binding missing sanitized fields: {missing_sanitized}")

        secret_ref = str(binding.get("secret_ref") or "")
        require(secret_ref.startswith("ref:"), "secret_ref must be a reference, not a value")
        require("://" not in secret_ref, "secret_ref must not contain a provider URL")
        require(binding.get("credential_requirement") == "required", "fixture profiles must require credential")
        require(binding.get("secret_ref_status") == "present", "fixture must declare reference presence only")

    policy = manifest.get("policy") or {}
    require(policy.get("stores_secret_values") is False, "secret reference fixture must not store values")
    require(policy.get("resolver_enabled") is False, "secret reference fixture must not enable resolver")
    require(policy.get("cloud_calls_allowed") is False, "secret reference fixture must not allow cloud calls")
    require(
        policy.get("production_secret_backend_ready") is False,
        "secret reference fixture must not claim production secret backend ready",
    )
    missing_forbidden = sorted(REQUIRED_FORBIDDEN_FIELDS - set(policy.get("forbidden_fields") or []))
    require(not missing_forbidden, f"policy missing forbidden fields: {missing_forbidden}")


def assert_readiness_alignment() -> None:
    readiness = load_json(READINESS_FIXTURE_PATH)
    planned = {
        str(item.get("id")): item
        for item in readiness.get("planned_slices") or []
        if isinstance(item, dict)
    }
    secret_ref_slice = planned.get("secret-ref-schema-and-fixtures") or {}
    require(secret_ref_slice.get("status") == "satisfied", "secret-ref-schema-and-fixtures must be marked satisfied")
    evidence = set(secret_ref_slice.get("evidence") or [])
    for path in {
        "contracts/production-secret-reference.schema.json",
        "scripts/checks/fixtures/production-secret-reference-basic.json",
        "scripts/check-production-secret-reference-contract.py",
    }:
        require(path in evidence, f"secret-ref-schema-and-fixtures missing evidence: {path}")
        require((REPO_ROOT / path).exists(), f"missing secret-ref evidence path: {path}")

    preconditions = {
        str(item.get("id")): item
        for item in readiness.get("required_preconditions") or []
        if isinstance(item, dict)
    }
    secret_ref_precondition = preconditions.get("secret-ref-schema") or {}
    require(secret_ref_precondition.get("status") == "satisfied", "secret-ref-schema precondition must be satisfied")
    require(
        "contracts/production-secret-reference.schema.json" in set(secret_ref_precondition.get("evidence") or []),
        "secret-ref-schema precondition must cite schema evidence",
    )


def assert_docs_and_check_repo() -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("check-production-secret-reference-contract.py", [])' in check_repo,
        "check-repo.py must run production secret reference contract check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def assert_no_secret_literals() -> None:
    text = "\n".join(
        [
            SCHEMA_PATH.read_text(encoding="utf-8"),
            FIXTURE_PATH.read_text(encoding="utf-8"),
        ]
    )
    forbidden_literals = [
        "Bearer ",
        "BEGIN PRIVATE KEY",
        "AKIA",
        "authorization:",
        "cookie:",
        "token:",
    ]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"secret reference schema or fixture contains forbidden secret-looking literals: {found}")
    require(
        re.search(r"sk-[A-Za-z0-9]{8,}", text) is None,
        "secret reference schema or fixture contains forbidden sk-like token",
    )
    require(
        re.search(r"https?://", FIXTURE_PATH.read_text(encoding="utf-8")) is None,
        "secret reference fixture must not contain provider raw URL",
    )


def main() -> None:
    schema = load_json(SCHEMA_PATH)
    manifest = load_json(FIXTURE_PATH)
    if jsonschema is not None:
        jsonschema.Draft202012Validator.check_schema(schema)
        jsonschema.validate(manifest, schema)
    require(isinstance(manifest, dict), "secret reference fixture must be an object")
    assert_manifest(manifest)
    assert_readiness_alignment()
    assert_docs_and_check_repo()
    assert_no_secret_literals()
    print("production secret reference contract checks passed.")


if __name__ == "__main__":
    main()

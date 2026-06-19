#!/usr/bin/env python3
from __future__ import annotations

import json
import re
import runpy
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-product-sample-consistency-v1.json"
RESPONSE_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-response-fixtures-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "control-plane-read-response-fixtures-v1": (
        "scripts/checks/fixtures/control-plane-read-response-fixtures-v1.json",
        "governance_boundary_satisfied",
    ),
    "control-plane-read-fake-store-handler-implementation-v1": (
        "scripts/checks/fixtures/control-plane-read-fake-store-handler-implementation-v1.json",
        "fake_store_handler_implemented",
    ),
    "control-plane-read-formal-ui-readiness-close-v1": (
        "scripts/checks/fixtures/control-plane-read-formal-ui-readiness-close-v1.json",
        "formal_ui_readiness_closed",
    ),
    "control-plane-read-dev-live-consumer-v1": (
        "scripts/checks/fixtures/control-plane-read-dev-live-consumer-v1.json",
        "dev_live_consumer_implemented",
    ),
}
EXPECTED_ROUTE_IDS = {
    "tenant-summary-route",
    "application-summary-list-route",
    "api-key-summary-list-route",
    "quota-summary-route",
    "workflow-definition-summary-list-route",
    "run-record-summary-list-route",
    "audit-summary-list-route",
}
EXPECTED_SUCCESS_ITEM_COUNTS = {
    "tenant-summary-route": 1,
    "application-summary-list-route": 2,
    "api-key-summary-list-route": 2,
    "quota-summary-route": 1,
    "workflow-definition-summary-list-route": 2,
    "run-record-summary-list-route": 2,
    "audit-summary-list-route": 2,
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "production_api_consumer_ready",
    "database_schema_ready",
    "database_query_ready",
    "repository_implementation_ready",
    "store_selector_ready",
    "radish_oidc_client_ready",
    "auth_middleware_ready",
    "api_key_lifecycle_ready",
    "quota_enforcement_ready",
    "workflow_executor_ready",
    "confirmation_flow_ready",
    "business_writeback_ready",
    "run_replay_ready",
    "run_resume_ready",
    "production_ready",
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must be a JSON object")
    return document


def read_text(relative_path: str) -> str:
    return (REPO_ROOT / relative_path).read_text(encoding="utf-8")


def success_examples_by_route() -> dict[str, dict[str, Any]]:
    fixture = load_json(RESPONSE_FIXTURE_PATH)
    examples = {
        str(example.get("route_id") or ""): example.get("success") or {}
        for example in fixture.get("response_examples") or []
        if isinstance(example, dict)
    }
    require(set(examples) == EXPECTED_ROUTE_IDS, "response fixture route ids drifted")
    return examples


def assert_fixture(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "control_plane_read_product_sample_consistency_v1", "unexpected kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "control-plane-read-product-sample-consistency-v1", "unexpected slice id")
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(slice_info.get("status") == "product_sample_consistency_guarded", "unexpected slice status")
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")

    source = fixture.get("canonical_source") or {}
    require(
        source.get("fixture") == "scripts/checks/fixtures/control-plane-read-response-fixtures-v1.json",
        "canonical source must remain response fixture",
    )
    require(set(source.get("matched_fields") or []) == {"success.items", "success.next_cursor"}, "matched fields drifted")
    require(
        set(source.get("consumer_specific_envelope_fields") or []) == {"request_id", "audit_ref"},
        "consumer-specific envelope fields drifted",
    )


def assert_dependencies(fixture: dict[str, Any]) -> None:
    declared = set(fixture.get("depends_on") or [])
    require(set(EXPECTED_DEPENDENCIES).issubset(declared), "fixture must declare sample consistency dependencies")
    for dependency_id, (relative_path, expected_status) in EXPECTED_DEPENDENCIES.items():
        dependency = load_json(REPO_ROOT / relative_path)
        slice_info = dependency.get("slice") or {}
        require(slice_info.get("id") == dependency_id, f"dependency {dependency_id} id drifted")
        require(slice_info.get("status") == expected_status, f"dependency {dependency_id} status drifted")


def assert_response_product_relationships(fixture: dict[str, Any], examples: dict[str, dict[str, Any]]) -> None:
    for route_id, expected_count in EXPECTED_SUCCESS_ITEM_COUNTS.items():
        items = examples[route_id].get("items") or []
        require(len(items) == expected_count, f"{route_id} success item count drifted")

    tenant_items = examples["tenant-summary-route"].get("items") or []
    quota_items = examples["quota-summary-route"].get("items") or []
    require(tenant_items[0].get("tenant_ref") == "tenant_demo", "tenant sample ref drifted")
    require(quota_items[0].get("tenant_ref") == "tenant_demo", "quota tenant ref drifted")
    require(
        tenant_items[0].get("quota_summary_ref") == quota_items[0].get("quota_id"),
        "tenant quota ref must match quota summary",
    )

    applications = {
        str(item.get("application_ref")): item
        for item in examples["application-summary-list-route"].get("items") or []
        if isinstance(item, dict)
    }
    workflow_definitions = {
        str(item.get("workflow_definition_id")): item
        for item in examples["workflow-definition-summary-list-route"].get("items") or []
        if isinstance(item, dict)
    }
    runs = {
        str(item.get("run_id")): item
        for item in examples["run-record-summary-list-route"].get("items") or []
        if isinstance(item, dict)
    }
    api_keys = [
        item
        for item in examples["api-key-summary-list-route"].get("items") or []
        if isinstance(item, dict)
    ]

    required_products = (fixture.get("canonical_source") or {}).get("required_product_refs") or []
    require(len(required_products) == 2, "required product refs must cover RadishFlow and Docs")
    for product in required_products:
        application_ref = str(product.get("application_ref") or "")
        workflow_definition_id = str(product.get("workflow_definition_id") or "")
        run_id = str(product.get("run_id") or "")
        owner_subject_ref = str(product.get("owner_subject_ref") or "")

        application = applications.get(application_ref)
        require(application is not None, f"missing application sample: {application_ref}")
        require(application.get("display_name") == product.get("display_name"), f"{application_ref} display name drifted")
        require(application.get("owner_subject_ref") == owner_subject_ref, f"{application_ref} owner drifted")
        require(
            application.get("latest_workflow_definition_ref") == workflow_definition_id,
            f"{application_ref} latest workflow definition drifted",
        )

        definition = workflow_definitions.get(workflow_definition_id)
        require(definition is not None, f"missing workflow definition sample: {workflow_definition_id}")
        require(definition.get("application_ref") == application_ref, f"{workflow_definition_id} application ref drifted")

        run = runs.get(run_id)
        require(run is not None, f"missing run sample: {run_id}")
        require(run.get("application_ref") == application_ref, f"{run_id} application ref drifted")
        require(run.get("workflow_definition_id") == workflow_definition_id, f"{run_id} definition ref drifted")

        require(
            any(api_key.get("owner_subject_ref") == owner_subject_ref for api_key in api_keys),
            f"{application_ref} owner has no read-side API key summary sample",
        )


def extract_typescript_return_object(source: str, function_name: str) -> dict[str, Any]:
    function_index = source.find(f"function {function_name}")
    require(function_index >= 0, f"missing frontend offline sample function: {function_name}")
    return_index = source.find("return", function_index)
    require(return_index >= 0, f"{function_name} missing return")
    object_start = source.find("{", return_index)
    require(object_start >= 0, f"{function_name} missing return object")
    object_end = matching_brace_index(source, object_start)
    object_literal = source[object_start : object_end + 1]
    return parse_typescript_object_literal(object_literal, function_name)


def matching_brace_index(source: str, object_start: int) -> int:
    depth = 0
    quote: str | None = None
    escaped = False
    for index in range(object_start, len(source)):
        char = source[index]
        if quote is not None:
            if escaped:
                escaped = False
            elif char == "\\":
                escaped = True
            elif char == quote:
                quote = None
            continue
        if char in {'"', "'", "`"}:
            quote = char
            continue
        if char == "{":
            depth += 1
        elif char == "}":
            depth -= 1
            if depth == 0:
                return index
    raise SystemExit("unterminated TypeScript object literal")


def parse_typescript_object_literal(object_literal: str, function_name: str) -> dict[str, Any]:
    json_like = re.sub(
        r"([{\[,]\s*)([A-Za-z_][A-Za-z0-9_]*)\s*:",
        lambda match: f'{match.group(1)}"{match.group(2)}":',
        object_literal,
    )
    json_like = re.sub(r",\s*([}\]])", r"\1", json_like)
    try:
        document = json.loads(json_like)
    except json.JSONDecodeError as exc:
        raise SystemExit(f"failed to parse {function_name} TypeScript sample: {exc}") from exc
    require(isinstance(document, dict), f"{function_name} must return an object")
    return document


def assert_frontend_offline_samples(fixture: dict[str, Any], examples: dict[str, dict[str, Any]]) -> None:
    samples = fixture.get("frontend_offline_samples") or []
    require(len(samples) == len(EXPECTED_ROUTE_IDS), "frontend sample matrix route count drifted")
    route_ids = {str(sample.get("route_id") or "") for sample in samples if isinstance(sample, dict)}
    require(route_ids == EXPECTED_ROUTE_IDS, "frontend sample matrix route ids drifted")

    for sample in samples:
        route_id = str(sample.get("route_id") or "")
        relative_path = str(sample.get("file") or "")
        function_name = str(sample.get("function") or "")
        source = read_text(relative_path)
        envelope = extract_typescript_return_object(source, function_name)
        expected = examples[route_id]

        require(envelope.get("items") == expected.get("items"), f"{relative_path} {function_name} items drifted")
        require(
            envelope.get("next_cursor") == expected.get("next_cursor"),
            f"{relative_path} {function_name} next_cursor drifted",
        )
        require(isinstance(envelope.get("request_id"), str) and envelope.get("request_id"), f"{function_name} request_id missing")
        require(isinstance(envelope.get("audit_ref"), str) and envelope.get("audit_ref"), f"{function_name} audit_ref missing")
        require(envelope.get("tenant_ref") == expected.get("tenant_ref"), f"{function_name} tenant_ref drifted")
        require(envelope.get("failure_code") is None, f"{function_name} must stay a success sample")


def assert_go_test_coverage(fixture: dict[str, Any]) -> None:
    go_consumer = next(
        (
            consumer
            for consumer in fixture.get("checked_consumers") or []
            if isinstance(consumer, dict) and consumer.get("consumer_id") == "go_fake_store"
        ),
        None,
    )
    require(go_consumer is not None, "missing Go fake store consumer entry")
    test_file = str(go_consumer.get("test_file") or "")
    source = read_text(test_file)
    for literal in (
        "TestControlPlaneReadFakeStoreMatchesResponseFixture",
        "control-plane-read-response-fixtures-v1.json",
        "TenantSummary(\"tenant_demo\")",
        "ApplicationSummaries(\"tenant_demo\", nil)",
        "APIKeySummaries(\"tenant_demo\", nil)",
        "QuotaSummary(\"tenant_demo\")",
        "WorkflowDefinitionSummaries(\"tenant_demo\", nil)",
        "RunRecordSummaries(\"tenant_demo\", nil)",
        "AuditSummaries(\"tenant_demo\", nil)",
    ):
        require(literal in source, f"Go product sample test missing literal: {literal}")
    for route_id in EXPECTED_ROUTE_IDS:
        require(route_id in source, f"Go product sample test missing route: {route_id}")


def assert_consumer_smoke_product_refs() -> None:
    module = runpy.run_path(str(REPO_ROOT / "scripts/run-control-plane-read-consumer-smoke.py"))
    document = module["build_consumer_view"]()
    invariants = document.get("invariants") or {}
    require(invariants.get("product_sample_refs_preserved") is True, "consumer smoke lost product sample refs invariant")
    product_samples = document.get("product_samples") or {}
    require(
        product_samples.get("canonical_source") == "scripts/checks/fixtures/control-plane-read-response-fixtures-v1.json",
        "consumer smoke canonical source drifted",
    )
    refs = product_samples.get("success_item_refs_by_route") or {}
    require(set(refs) == EXPECTED_ROUTE_IDS, "consumer smoke product sample route refs drifted")
    require(all(isinstance(items, list) and items for items in refs.values()), "consumer smoke product refs must be non-empty")


def assert_docs_and_baseline(fixture: dict[str, Any]) -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("checks/control_plane/check-control-plane-read-product-sample-consistency-v1.py", [])'
        in check_repo,
        "product sample consistency checker must be in check-repo.py",
    )

    doc_refs = fixture.get("required_doc_references") or {}
    require(isinstance(doc_refs, dict), "required_doc_references must be an object")
    for relative_path, required_literals in doc_refs.items():
        document = read_text(str(relative_path))
        for literal in required_literals:
            require(str(literal) in document, f"{relative_path} missing doc reference: {literal}")


def assert_testing_strategy(fixture: dict[str, Any]) -> None:
    commands = {str(item.get("command")) for item in fixture.get("testing_strategy") or [] if isinstance(item, dict)}
    expected_commands = {
        "./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-product-sample-consistency-v1.py",
        "go test ./...",
        "./scripts/run-python.sh scripts/run-control-plane-read-consumer-smoke.py --check",
        "npm run build",
        "./scripts/check-repo.sh --fast",
    }
    missing_commands = sorted(expected_commands - commands)
    require(not missing_commands, f"missing testing commands: {missing_commands}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    examples = success_examples_by_route()
    assert_fixture(fixture)
    assert_dependencies(fixture)
    assert_response_product_relationships(fixture, examples)
    assert_frontend_offline_samples(fixture, examples)
    assert_go_test_coverage(fixture)
    assert_consumer_smoke_product_refs()
    assert_testing_strategy(fixture)
    assert_docs_and_baseline(fixture)
    print("control-plane-read-product-sample-consistency-v1 check passed")


if __name__ == "__main__":
    main()

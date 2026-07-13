#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/doc-language-policy-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_CATEGORIES = [
    "technical_term_without_stable_chinese_equivalent",
    "code_identifier",
    "command",
    "path",
    "config_key",
    "type_name",
    "interface_name",
    "api_route",
    "protocol_field",
    "schema_field",
    "status_anchor",
    "fixture_id",
    "checker_id",
    "external_project_name",
    "quoted_source_text",
]
EXPECTED_REMEDIATION_ORDER = [
    "entry_documents",
    "feature_and_platform_topic_entries",
    "contracts",
    "task_cards",
    "devlogs",
]
EXPECTED_TERM_IDS = [
    "readiness",
    "entry_review",
    "runtime",
    "blocker_matrix",
    "evidence_rollup",
    "artifact_guard",
    "no_fallback",
    "no_side_effects",
    "smoke",
    "public_production_api",
    "surface",
    "backend",
    "future",
    "metadata-only",
    "fail-closed",
    "dev-only",
    "scope",
    "owner",
    "repository",
    "handoff",
    "candidate",
    "baseline",
    "drift",
    "offline",
    "production",
]
EXPECTED_PRIORITY_REMEDIATION_DOCS = [
    "docs/radishmind-current-focus.md",
    "docs/features/README.md",
    "docs/platform/README.md",
    "scripts/README.md",
    "docs/radishmind-code-standards.md",
]
EXPECTED_SECOND_BATCH_BOUNDARIES = [
    "new_or_touched_docs_use_chinese_prose",
    "preserve_machine_checked_identifiers",
    "prioritize_entry_and_topic_documents",
    "document_term_changes_with_fixture_checker_sync",
    "avoid_mass_translation",
]
EXPECTED_ACTIVE_PRODUCT_PATH_DOCS = [
    "docs/radishmind-current-focus.md",
    "docs/features/README.md",
    "docs/features/user-workspace.md",
    "docs/features/user-workspace/README.md",
    "docs/features/user-workspace/application-api-integration-invocation-v1.md",
    "docs/features/user-workspace/application-configuration-draft-review-v1.md",
    "docs/features/user-workspace/application-publish-governance-promotion-v1.md",
]
EXPECTED_FORBIDDEN_ACTIONS = [
    "mass_translate_repository",
    "rename_status_anchors_without_checker_update",
    "translate_code_protocol_or_config_identifiers",
    "turn_language_cleanup_into_runtime_scope",
    "add_task_card_for_plain_copyediting",
]


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def read(relative_path: str) -> str:
    path = REPO_ROOT / relative_path
    require(path.exists(), f"required file missing: {relative_path}")
    return path.read_text(encoding="utf-8")


def load_fixture() -> dict[str, Any]:
    document = json.loads(FIXTURE_PATH.read_text(encoding="utf-8"))
    require(isinstance(document, dict), "doc language policy fixture must contain an object")
    return document


def require_sequence(document: dict[str, Any], key: str, expected: list[str]) -> None:
    actual = document.get(key)
    require(isinstance(actual, list), f"{key} must be a list")
    require(actual == expected, f"{key} mismatch: expected {expected}, got {actual}")


def check_topic_document(document: dict[str, Any]) -> None:
    topic = document.get("topic_document")
    require(isinstance(topic, dict), "topic_document must be an object")
    path = topic.get("path")
    require(isinstance(path, str) and path, "topic_document.path must be a non-empty string")
    status_anchor = topic.get("status_anchor")
    require(isinstance(status_anchor, str) and status_anchor, "topic_document.status_anchor must be non-empty")
    required_snippets = topic.get("required_snippets")
    require(
        isinstance(required_snippets, list) and required_snippets,
        "topic_document.required_snippets must be a non-empty list",
    )

    text = read(path)
    require(status_anchor in text, f"{path} missing topic status anchor: {status_anchor}")
    for snippet in required_snippets:
        require(isinstance(snippet, str) and snippet, "topic document snippet must be non-empty")
        require(snippet in text, f"{path} missing topic snippet: {snippet}")


def check_policy_anchors(document: dict[str, Any]) -> None:
    anchors = document.get("required_policy_anchors")
    require(isinstance(anchors, list) and anchors, "required_policy_anchors must be a non-empty list")
    priority_documents = document.get("priority_documents")
    require(isinstance(priority_documents, list) and priority_documents, "priority_documents must be a non-empty list")
    priority_set = set(priority_documents)

    for anchor in anchors:
        require(isinstance(anchor, dict), "each required_policy_anchors entry must be an object")
        path = anchor.get("path")
        require(isinstance(path, str) and path in priority_set, f"anchor path is not in priority_documents: {path}")
        snippets = anchor.get("snippets")
        require(isinstance(snippets, list) and snippets, f"{path} snippets must be a non-empty list")
        text = read(path)
        for snippet in snippets:
            require(isinstance(snippet, str) and snippet, f"{path} snippet must be a non-empty string")
            require(snippet in text, f"{path} missing language policy snippet: {snippet}")


def check_scope(document: dict[str, Any]) -> None:
    require(document.get("schema_version") == 1, "schema_version must be 1")
    require(document.get("kind") == "doc_language_policy_v1", "kind mismatch")
    require(document.get("status") == "doc_language_policy_defined", "status mismatch")
    scope = document.get("scope")
    require(isinstance(scope, dict), "scope must be an object")
    require(scope.get("default_prose_language") == "zh-CN", "default prose language must be zh-CN")
    require(
        scope.get("legacy_remediation_policy") == "incremental_priority_documents_first",
        "legacy remediation policy mismatch",
    )
    require(
        scope.get("machine_checked_literals_policy")
        == "preserve_until_checker_fixture_and_docs_are_updated_together",
        "machine checked literals policy mismatch",
    )
    require_sequence(document, "preserve_original_categories", EXPECTED_CATEGORIES)
    require_sequence(document, "remediation_order", EXPECTED_REMEDIATION_ORDER)
    require_sequence(document, "second_batch_boundaries", EXPECTED_SECOND_BATCH_BOUNDARIES)
    require_sequence(document, "forbidden_document_language_actions", EXPECTED_FORBIDDEN_ACTIONS)

    does_not_require = document.get("does_not_require")
    require(isinstance(does_not_require, list), "does_not_require must be a list")
    for expected in (
        "mass_translate_legacy_docs",
        "rename_status_anchors",
        "rename_fixture_keys",
        "rewrite_checked_paths",
        "change_runtime_or_public_api",
    ):
        require(expected in does_not_require, f"does_not_require missing {expected}")


def check_preferred_terms(document: dict[str, Any]) -> None:
    terms = document.get("preferred_chinese_terms")
    require(isinstance(terms, list), "preferred_chinese_terms must be a list")
    actual_ids = []
    topic_path = document["topic_document"]["path"]
    topic_text = read(topic_path)

    for term in terms:
        require(isinstance(term, dict), "preferred_chinese_terms entries must be objects")
        term_id = term.get("term_id")
        preferred_zh = term.get("preferred_zh")
        usage_boundary = term.get("usage_boundary")
        require(isinstance(term_id, str) and term_id, "term_id must be non-empty")
        require(isinstance(preferred_zh, str) and preferred_zh, f"{term_id} preferred_zh must be non-empty")
        require(isinstance(usage_boundary, str) and usage_boundary, f"{term_id} usage_boundary must be non-empty")
        require(f"`{term_id}`" in topic_text, f"{topic_path} missing term id `{term_id}`")
        require(preferred_zh in topic_text, f"{topic_path} missing preferred Chinese term: {preferred_zh}")
        actual_ids.append(term_id)

    require(actual_ids == EXPECTED_TERM_IDS, f"preferred_chinese_terms ids mismatch: {actual_ids}")


def check_priority_remediation_batch(document: dict[str, Any]) -> None:
    batch = document.get("priority_remediation_batch")
    require(isinstance(batch, dict), "priority_remediation_batch must be an object")
    status = batch.get("status")
    require(
        status == "doc_language_priority_documents_remediation_v1_defined",
        "priority_remediation_batch.status mismatch",
    )
    target_documents = batch.get("target_documents")
    require(
        isinstance(target_documents, list),
        "priority_remediation_batch.target_documents must be a list",
    )
    require(
        target_documents == EXPECTED_PRIORITY_REMEDIATION_DOCS,
        f"priority remediation target docs mismatch: {target_documents}",
    )

    topic_text = read(document["topic_document"]["path"])
    require(status in topic_text, f"topic document missing priority remediation status: {status}")

    document_snippets = batch.get("document_snippets")
    require(isinstance(document_snippets, list), "priority remediation document_snippets must be a list")
    snippet_paths = []
    for entry in document_snippets:
        require(isinstance(entry, dict), "priority remediation snippet entry must be an object")
        path = entry.get("path")
        snippets = entry.get("snippets")
        require(isinstance(path, str) and path in target_documents, f"unknown priority remediation path: {path}")
        require(isinstance(snippets, list) and snippets, f"{path} priority snippets must be non-empty")
        text = read(path)
        for snippet in snippets:
            require(isinstance(snippet, str) and snippet, f"{path} priority snippet must be non-empty")
            require(snippet in text, f"{path} missing priority remediation snippet: {snippet}")
        snippet_paths.append(path)

    require(snippet_paths == EXPECTED_PRIORITY_REMEDIATION_DOCS, f"priority snippet paths mismatch: {snippet_paths}")


def check_active_product_path_remediation_batch(document: dict[str, Any]) -> None:
    batch = document.get("active_product_path_remediation_batch")
    require(isinstance(batch, dict), "active_product_path_remediation_batch must be an object")
    status = batch.get("status")
    require(
        status == "doc_language_active_product_path_remediation_v2_defined",
        "active_product_path_remediation_batch.status mismatch",
    )
    target_documents = batch.get("target_documents")
    require(
        target_documents == EXPECTED_ACTIVE_PRODUCT_PATH_DOCS,
        f"active product path target docs mismatch: {target_documents}",
    )
    topic_text = read(document["topic_document"]["path"])
    require(status in topic_text, f"topic document missing active product path status: {status}")

    document_snippets = batch.get("document_snippets")
    require(isinstance(document_snippets, list), "active product path document_snippets must be a list")
    snippet_paths = []
    for entry in document_snippets:
        require(isinstance(entry, dict), "active product path snippet entry must be an object")
        path = entry.get("path")
        snippets = entry.get("snippets")
        require(isinstance(path, str) and path in target_documents, f"unknown active product path: {path}")
        require(isinstance(snippets, list) and snippets, f"{path} active product path snippets must be non-empty")
        text = read(path)
        for snippet in snippets:
            require(isinstance(snippet, str) and snippet, f"{path} active product path snippet must be non-empty")
            require(snippet in text, f"{path} missing active product path snippet: {snippet}")
        snippet_paths.append(path)

    require(
        snippet_paths == EXPECTED_ACTIVE_PRODUCT_PATH_DOCS,
        f"active product path snippet paths mismatch: {snippet_paths}",
    )


def check_repo_registration() -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("check-doc-language-policy-v1.py", [])' in check_repo,
        "check-repo.py must register check-doc-language-policy-v1.py",
    )


def main() -> int:
    document = load_fixture()
    check_scope(document)
    check_topic_document(document)
    check_policy_anchors(document)
    check_preferred_terms(document)
    check_priority_remediation_batch(document)
    check_active_product_path_remediation_batch(document)
    check_repo_registration()
    print("doc language policy v1 checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

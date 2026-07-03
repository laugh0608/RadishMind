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


def check_repo_registration() -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("check-doc-language-policy-v1.py", [])' in check_repo,
        "check-repo.py must register check-doc-language-policy-v1.py",
    )


def main() -> int:
    document = load_fixture()
    check_scope(document)
    check_policy_anchors(document)
    check_repo_registration()
    print("doc language policy v1 checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

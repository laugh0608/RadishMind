#!/usr/bin/env python3
from __future__ import annotations

import json
import sys
from pathlib import Path

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent
REQUEST_SCHEMA_PATH = REPO_ROOT / "contracts" / "copilot-request.schema.json"


def load_schema() -> dict:
    return json.loads(REQUEST_SCHEMA_PATH.read_text(encoding="utf-8"))


def validate_request(document: dict, label: str, schema: dict) -> None:
    try:
        jsonschema.validate(document, schema)
    except jsonschema.ValidationError as exc:
        raise SystemExit(f"{label}: schema validation failed unexpectedly: {exc.message}") from exc


def expect_request_failure(document: dict, label: str, schema: dict, needle: str) -> None:
    try:
        jsonschema.validate(document, schema)
    except jsonschema.ValidationError as exc:
        message = exc.message
        if needle not in message:
            raise SystemExit(f"{label}: expected schema error containing {needle!r}, got: {message!r}") from exc
        return
    raise SystemExit(f"{label}: expected schema validation to fail")


def build_radish_docs_request() -> dict:
    return {
        "schema_version": 1,
        "request_id": "check-copilot-request-artifact-metadata-docs-001",
        "project": "radish",
        "task": "answer_docs_question",
        "locale": "zh-CN",
        "artifacts": [
            {
                "kind": "markdown",
                "role": "primary",
                "name": "docs_page",
                "mime_type": "text/markdown",
                "content": "# Sample\n\nDocs content.",
                "metadata": {
                    "source_type": "docs",
                    "page_slug": "guide/request-schema",
                    "fragment_id": "artifact-metadata",
                    "retrieval_rank": 1,
                    "is_official": True,
                },
            }
        ],
        "context": {
            "current_app": "document",
            "route": "/document/guide/request-schema",
            "resource": {
                "kind": "wiki_document",
                "slug": "guide/request-schema",
                "title": "Request Schema",
            },
            "search_scope": ["docs"],
        },
        "tool_hints": {
            "allow_retrieval": True,
            "allow_tool_calls": False,
            "allow_image_reasoning": False,
        },
        "safety": {
            "mode": "advisory",
            "requires_confirmation_for_actions": True,
        },
    }


def build_radishflow_control_plane_request() -> dict:
    return {
        "schema_version": 1,
        "request_id": "check-copilot-request-artifact-metadata-control-plane-001",
        "project": "radishflow",
        "task": "explain_control_plane_state",
        "locale": "zh-CN",
        "artifacts": [
            {
                "kind": "attachment_ref",
                "role": "supporting",
                "name": "private_registry_capture",
                "mime_type": "application/json",
                "uri": "capture://radishflow/control-plane/private-registry-429-check",
                "metadata": {
                    "sanitized_summary": "private package source returned repeated 429 responses; auth headers and payload bodies were stripped",
                    "source_scope": "control_plane",
                    "summary_variant": "sanitized_only",
                },
            },
            {
                "kind": "attachment_ref",
                "role": "supporting",
                "name": "lease_cache_capture",
                "mime_type": "application/json",
                "uri": "capture://radishflow/control-plane/lease-cache-stale-check",
                "metadata": {
                    "summary": "cached lease snapshot still marks the entitlement window as usable",
                    "redaction_summary": "token hash and refresh cookie were omitted",
                    "redactions": ["token_hash", "refresh_cookie"],
                    "source_scope": "control_plane",
                    "summary_variant": "summary_plus_redaction",
                },
            },
        ],
        "context": {
            "document_revision": 42,
            "control_plane_state": {
                "entitlement_status": "active",
                "lease_status": "stale",
                "sync_status": "partial_failed",
                "manifest_status": "current",
            },
        },
        "tool_hints": {
            "allow_retrieval": False,
            "allow_tool_calls": False,
            "allow_image_reasoning": False,
        },
        "safety": {
            "mode": "advisory",
            "requires_confirmation_for_actions": True,
        },
    }


def build_raw_metadata_request() -> dict:
    request = build_radishflow_control_plane_request()
    request["artifacts"][0]["metadata"]["headers"] = {"authorization": "redacted"}
    return request


def main() -> int:
    schema = load_schema()
    validate_request(build_radish_docs_request(), "docs request metadata regression", schema)
    validate_request(build_radishflow_control_plane_request(), "control-plane request metadata regression", schema)
    expect_request_failure(
        build_raw_metadata_request(),
        "raw metadata request regression",
        schema,
        "'headers' should not be valid",
    )
    print("copilot request artifact metadata regression passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

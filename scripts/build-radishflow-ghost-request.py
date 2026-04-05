#!/usr/bin/env python3
from __future__ import annotations

import argparse
import sys
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.candidate_records import (  # noqa: E402
    ensure_schema,
    load_json_document,
    make_repo_relative,
    resolve_relative_to_repo,
    write_json_document,
)


GHOST_CANDIDATE_SET_SCHEMA_PATH = REPO_ROOT / "contracts/radishflow-ghost-candidate-set.schema.json"
COPILOT_REQUEST_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-request.schema.json"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Build a minimal CopilotRequest for RadishFlow suggest_ghost_completion from a local ghost candidate set.",
    )
    parser.add_argument("--input", required=True, help="Path to a radishflow ghost candidate set json file.")
    parser.add_argument("--output", required=True, help="Output CopilotRequest json path.")
    parser.add_argument("--artifact-uri", required=True, help="Primary flowsheet artifact uri used in the generated request.")
    parser.add_argument("--request-id", required=True, help="request_id written into the generated CopilotRequest.")
    parser.add_argument("--locale", default="zh-CN", help="Locale written into the generated CopilotRequest.")
    parser.add_argument(
        "--assembly-profile",
        choices=["model-minimal", "debug-full"],
        default="model-minimal",
        help="Assembly profile. model-minimal trims local-only signals before sending to the model.",
    )
    parser.add_argument(
        "--check",
        action="store_true",
        help="Do not rewrite output; fail if the generated request does not match the existing output file.",
    )
    return parser.parse_args()


MODEL_REQUEST_CANDIDATE_KEYS = {
    "candidate_ref",
    "ghost_kind",
    "target_port_key",
    "target_unit_id",
    "source_stream_id",
    "target_node_id",
    "suggested_stream_name",
    "is_high_confidence",
    "is_tab_default",
}

MODEL_REQUEST_NEARBY_NODE_KEYS = {
    "type",
    "id",
    "direction",
}


def trim_candidate_for_model(candidate: dict[str, Any]) -> dict[str, Any]:
    return {key: value for key, value in candidate.items() if key in MODEL_REQUEST_CANDIDATE_KEYS}


def trim_nearby_node_for_model(node: dict[str, Any]) -> dict[str, Any]:
    return {key: value for key, value in node.items() if key in MODEL_REQUEST_NEARBY_NODE_KEYS}


def build_request(
    candidate_set: dict[str, Any],
    *,
    request_id: str,
    locale: str,
    artifact_uri: str,
    assembly_profile: str,
) -> dict[str, Any]:
    legal_candidate_completions = candidate_set["legal_candidate_completions"]
    nearby_nodes = candidate_set.get("nearby_nodes")

    if assembly_profile == "model-minimal":
        legal_candidate_completions = [
            trim_candidate_for_model(candidate)
            for candidate in legal_candidate_completions
        ]
        if nearby_nodes is not None:
            nearby_nodes = [
                trim_nearby_node_for_model(node)
                for node in nearby_nodes
            ]

    context: dict[str, Any] = {
        "document_revision": candidate_set["document_revision"],
        "selected_unit_ids": [candidate_set["selected_unit"]["id"]],
        "selected_unit": candidate_set["selected_unit"],
        "legal_candidate_completions": legal_candidate_completions,
    }

    for optional_key in (
        "unconnected_ports",
        "missing_canonical_ports",
        "cursor_context",
        "naming_hints",
        "topology_pattern_hints",
    ):
        if optional_key in candidate_set:
            context[optional_key] = candidate_set[optional_key]
    if nearby_nodes is not None:
        context["nearby_nodes"] = nearby_nodes

    request = {
        "schema_version": 1,
        "request_id": request_id,
        "project": "radishflow",
        "task": "suggest_ghost_completion",
        "locale": locale,
        "artifacts": [
            {
                "kind": "json",
                "role": "primary",
                "name": "flowsheet_document",
                "mime_type": "application/json",
                "uri": artifact_uri,
            }
        ],
        "context": context,
        "tool_hints": {
            "allow_retrieval": False,
            "allow_tool_calls": False,
            "allow_image_reasoning": False,
        },
        "safety": {
            "mode": "advisory",
            "requires_confirmation_for_actions": False,
        },
    }
    return request


def main() -> int:
    args = parse_args()
    input_path = resolve_relative_to_repo(args.input)
    output_path = resolve_relative_to_repo(args.output)

    if not input_path.is_file():
        raise SystemExit(f"input file not found: {args.input}")

    candidate_set = load_json_document(input_path)
    ensure_schema(candidate_set, GHOST_CANDIDATE_SET_SCHEMA_PATH, make_repo_relative(input_path))

    request = build_request(
        candidate_set,
        request_id=args.request_id.strip(),
        locale=args.locale.strip(),
        artifact_uri=args.artifact_uri.strip(),
        assembly_profile=args.assembly_profile.strip(),
    )
    ensure_schema(request, COPILOT_REQUEST_SCHEMA_PATH, "generated ghost CopilotRequest")

    if args.check:
        if not output_path.is_file():
            raise SystemExit(f"expected output file not found for --check: {args.output}")
        existing = load_json_document(output_path)
        if existing != request:
            raise SystemExit(f"generated request does not match expected output: {args.output}")
        print(f"ghost request assembly is up to date: {make_repo_relative(output_path)}")
        return 0

    write_json_document(output_path, request)
    print(f"wrote ghost CopilotRequest to {make_repo_relative(output_path)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

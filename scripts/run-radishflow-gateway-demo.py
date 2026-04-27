#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Any

import jsonschema

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from adapters.radishflow import build_adapter_snapshot_from_export_snapshot, build_request_from_snapshot  # noqa: E402
from services.gateway import GatewayOptions, handle_copilot_request, validate_gateway_envelope  # noqa: E402
from services.runtime.inference import validate_request_document, validate_response_document  # noqa: E402
from services.runtime.inference_support import load_schema  # noqa: E402


DEFAULT_EXPORT_PATH = (
    REPO_ROOT / "adapters/radishflow/exports/suggest-flowsheet-edits-reconnect-outlet-001.export.json"
)
EXPORT_SNAPSHOT_SCHEMA_PATH = REPO_ROOT / "contracts/radishflow-export-snapshot.schema.json"
ADAPTER_SNAPSHOT_SCHEMA_PATH = REPO_ROOT / "contracts/radishflow-adapter-snapshot.schema.json"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Run the RadishFlow export -> adapter/request -> Copilot Gateway demo with the mock provider."
        ),
    )
    parser.add_argument(
        "--input",
        default=str(DEFAULT_EXPORT_PATH.relative_to(REPO_ROOT)),
        help="Path to a RadishFlow export snapshot json file.",
    )
    parser.add_argument("--output", default="", help="Optional path for the generated gateway envelope json.")
    parser.add_argument(
        "--check",
        action="store_true",
        help="Validate the demo chain without printing the full gateway envelope.",
    )
    return parser.parse_args()


def resolve_repo_path(path_value: str) -> Path:
    path = Path(path_value)
    if not path.is_absolute():
        path = REPO_ROOT / path
    return path


def load_json_document(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to load json document: {path.relative_to(REPO_ROOT)}: {exc}") from exc


def write_json_document(path: Path, document: Any) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(document, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def assert_condition(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def build_demo_envelope(export_snapshot: dict[str, Any]) -> dict[str, Any]:
    jsonschema.validate(export_snapshot, load_schema(EXPORT_SNAPSHOT_SCHEMA_PATH))

    adapter_snapshot = build_adapter_snapshot_from_export_snapshot(export_snapshot)
    jsonschema.validate(adapter_snapshot, load_schema(ADAPTER_SNAPSHOT_SCHEMA_PATH))

    copilot_request = build_request_from_snapshot(adapter_snapshot)
    validate_request_document(copilot_request)

    envelope = handle_copilot_request(copilot_request, options=GatewayOptions(provider="mock"))
    validate_gateway_envelope(envelope)

    response = envelope.get("response")
    assert_condition(isinstance(response, dict), "gateway demo expected a response object")
    validate_response_document(response)

    metadata = envelope.get("metadata") or {}
    provider = metadata.get("provider") or {}
    assert_condition(metadata.get("route") == "radishflow/suggest_flowsheet_edits", "gateway route mismatch")
    assert_condition(metadata.get("advisory_only") is True, "gateway must remain advisory-only")
    assert_condition(metadata.get("request_validated") is True, "gateway request validation flag mismatch")
    assert_condition(metadata.get("response_validated") is True, "gateway response validation flag mismatch")
    assert_condition(provider.get("name") == "mock", "gateway demo must use the mock provider")
    assert_condition(envelope.get("status") in {"ok", "partial"}, "gateway demo expected ok or partial status")
    assert_condition(response.get("requires_confirmation") is True, "RadishFlow edit suggestions must require confirmation")
    return envelope


def main() -> int:
    args = parse_args()
    input_path = resolve_repo_path(args.input)
    if not input_path.is_file():
        raise SystemExit(f"input file not found: {args.input}")

    export_snapshot = load_json_document(input_path)
    if not isinstance(export_snapshot, dict):
        raise SystemExit(f"input must be a json object: {input_path.relative_to(REPO_ROOT)}")

    envelope = build_demo_envelope(export_snapshot)

    if args.output.strip():
        write_json_document(resolve_repo_path(args.output), envelope)

    if args.check:
        print("radishflow gateway demo smoke passed.")
    else:
        print(json.dumps(envelope, ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

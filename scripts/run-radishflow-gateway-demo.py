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
    parser.add_argument(
        "--manifest",
        default="",
        help="Optional manifest with multiple RadishFlow gateway demo fixtures.",
    )
    parser.add_argument("--output", default="", help="Optional path for the generated gateway envelope json.")
    parser.add_argument(
        "--check",
        action="store_true",
        help="Validate the demo chain without printing the full gateway envelope.",
    )
    parser.add_argument(
        "--summary-output",
        default="",
        help="Optional output path for a stable manifest summary json.",
    )
    parser.add_argument(
        "--check-summary",
        default="",
        help="Optional expected stable manifest summary json. Fails if regenerated summary differs.",
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


def assert_json_equal(expected: Any, actual: Any, *, label: str) -> None:
    if expected != actual:
        raise SystemExit(f"generated gateway demo summary does not match expected summary: {label}")


def resolve_expected_request(document: Any, label: str) -> dict[str, Any]:
    if isinstance(document, dict) and isinstance(document.get("input_request"), dict):
        return document["input_request"]
    if isinstance(document, dict):
        return document
    raise SystemExit(f"{label} must be a json object or an eval sample containing input_request")


def build_demo_envelope(export_snapshot: dict[str, Any], *, fixture: dict[str, Any] | None = None) -> dict[str, Any]:
    jsonschema.validate(export_snapshot, load_schema(EXPORT_SNAPSHOT_SCHEMA_PATH))

    adapter_snapshot = build_adapter_snapshot_from_export_snapshot(export_snapshot)
    jsonschema.validate(adapter_snapshot, load_schema(ADAPTER_SNAPSHOT_SCHEMA_PATH))

    copilot_request = build_request_from_snapshot(adapter_snapshot)
    validate_request_document(copilot_request)
    if fixture:
        expected_sample_path_value = str(fixture.get("sample") or "").strip()
        if expected_sample_path_value:
            expected_sample_path = resolve_repo_path(expected_sample_path_value)
            expected_request = resolve_expected_request(
                load_json_document(expected_sample_path),
                expected_sample_path.relative_to(REPO_ROOT).as_posix(),
            )
            assert_condition(
                copilot_request == expected_request,
                f"generated request does not match fixture sample: {expected_sample_path_value}",
            )

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


def load_export_snapshot(input_value: str) -> dict[str, Any]:
    input_path = resolve_repo_path(input_value)
    if not input_path.is_file():
        raise SystemExit(f"input file not found: {input_value}")

    export_snapshot = load_json_document(input_path)
    if not isinstance(export_snapshot, dict):
        raise SystemExit(f"input must be a json object: {input_path.relative_to(REPO_ROOT)}")
    return export_snapshot


def run_single_demo(args: argparse.Namespace) -> dict[str, Any]:
    return build_demo_envelope(load_export_snapshot(args.input))


def run_manifest_demo(manifest_path_value: str) -> dict[str, Any]:
    manifest_path = resolve_repo_path(manifest_path_value)
    if not manifest_path.is_file():
        raise SystemExit(f"manifest file not found: {manifest_path_value}")
    manifest = load_json_document(manifest_path)
    if not isinstance(manifest, dict):
        raise SystemExit(f"manifest must be a json object: {manifest_path.relative_to(REPO_ROOT)}")
    fixtures = manifest.get("fixtures")
    if not isinstance(fixtures, list) or not fixtures:
        raise SystemExit(f"manifest must contain a non-empty fixtures array: {manifest_path_value}")

    results: list[dict[str, Any]] = []
    for index, fixture in enumerate(fixtures, start=1):
        if not isinstance(fixture, dict):
            raise SystemExit(f"manifest fixture #{index} must be a json object")
        fixture_id = str(fixture.get("id") or "").strip()
        export_path_value = str(fixture.get("export_snapshot") or "").strip()
        if not fixture_id:
            raise SystemExit(f"manifest fixture #{index} is missing id")
        if not export_path_value:
            raise SystemExit(f"manifest fixture '{fixture_id}' is missing export_snapshot")
        envelope = build_demo_envelope(load_export_snapshot(export_path_value), fixture=fixture)
        metadata = envelope["metadata"]
        provider = metadata["provider"]
        response = envelope["response"]
        results.append(
            {
                "id": fixture_id,
                "request_id": envelope["request_id"],
                "status": envelope["status"],
                "route": metadata["route"],
                "provider": provider["name"],
                "advisory_only": metadata["advisory_only"],
                "request_validated": metadata["request_validated"],
                "response_validated": metadata["response_validated"],
                "requires_confirmation": response["requires_confirmation"],
            }
        )
    return {
        "schema_version": 1,
        "fixture_count": len(results),
        "results": results,
    }


def main() -> int:
    args = parse_args()
    if args.manifest.strip() and args.output.strip():
        raise SystemExit("--output is only supported for single --input mode")
    if not args.manifest.strip() and (args.summary_output.strip() or args.check_summary.strip()):
        raise SystemExit("--summary-output and --check-summary require --manifest")

    result = run_manifest_demo(args.manifest) if args.manifest.strip() else run_single_demo(args)

    if args.output.strip():
        write_json_document(resolve_repo_path(args.output), result)
    if args.summary_output.strip():
        write_json_document(resolve_repo_path(args.summary_output), result)
    if args.check_summary.strip():
        expected_summary_path = resolve_repo_path(args.check_summary)
        if not expected_summary_path.is_file():
            raise SystemExit(f"expected summary file not found: {args.check_summary}")
        assert_json_equal(
            load_json_document(expected_summary_path),
            result,
            label=expected_summary_path.relative_to(REPO_ROOT).as_posix(),
        )

    if args.check:
        if args.manifest.strip():
            print(f"radishflow gateway demo smoke passed: {result['fixture_count']} fixture(s).")
        else:
            print("radishflow gateway demo smoke passed.")
    else:
        print(json.dumps(result, ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

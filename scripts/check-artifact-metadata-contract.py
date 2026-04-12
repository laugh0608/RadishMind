#!/usr/bin/env python3
from __future__ import annotations

import json
import sys
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.artifact_metadata import (  # noqa: E402
    ARTIFACT_SUPPORTING_METADATA_KEYS,
    COPILOT_REQUEST_ARTIFACT_METADATA_KEYS,
    RAW_ARTIFACT_PAYLOAD_METADATA_KEYS,
)


def load_json(path: Path) -> dict[str, Any]:
    return json.loads(path.read_text(encoding="utf-8"))


def expect_equal(actual: Any, expected: Any, label: str) -> None:
    if actual != expected:
        raise SystemExit(f"{label}: expected {expected!r}, got {actual!r}")


def schema_metadata_contract(schema_path: Path, def_name: str) -> tuple[list[str], list[str]]:
    schema = load_json(schema_path)
    metadata_def = schema["$defs"][def_name]
    property_names = sorted(metadata_def.get("properties", {}).keys())
    banned_keys = list(metadata_def.get("propertyNames", {}).get("not", {}).get("enum", []))
    return property_names, banned_keys


def main() -> int:
    export_properties, export_banned = schema_metadata_contract(
        REPO_ROOT / "contracts" / "radishflow-export-snapshot.schema.json",
        "supporting_artifact_metadata",
    )
    adapter_properties, adapter_banned = schema_metadata_contract(
        REPO_ROOT / "contracts" / "radishflow-adapter-snapshot.schema.json",
        "supporting_artifact_metadata",
    )
    request_properties, request_banned = schema_metadata_contract(
        REPO_ROOT / "contracts" / "copilot-request.schema.json",
        "artifact_metadata",
    )

    expect_equal(
        export_properties,
        sorted(ARTIFACT_SUPPORTING_METADATA_KEYS),
        "radishflow export supporting artifact metadata properties",
    )
    expect_equal(
        adapter_properties,
        sorted(ARTIFACT_SUPPORTING_METADATA_KEYS),
        "radishflow adapter supporting artifact metadata properties",
    )
    expect_equal(
        request_properties,
        sorted(COPILOT_REQUEST_ARTIFACT_METADATA_KEYS),
        "copilot request artifact metadata properties",
    )
    expect_equal(
        export_banned,
        list(RAW_ARTIFACT_PAYLOAD_METADATA_KEYS),
        "radishflow export supporting artifact banned metadata keys",
    )
    expect_equal(
        adapter_banned,
        list(RAW_ARTIFACT_PAYLOAD_METADATA_KEYS),
        "radishflow adapter supporting artifact banned metadata keys",
    )
    expect_equal(
        request_banned,
        list(RAW_ARTIFACT_PAYLOAD_METADATA_KEYS),
        "copilot request artifact banned metadata keys",
    )

    print("artifact metadata contract alignment passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

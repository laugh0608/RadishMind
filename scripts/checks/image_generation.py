from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[2]
ALLOWED_RISK_LEVELS = {"low", "medium", "high"}


def load_json_document(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json document '{path.relative_to(REPO_ROOT)}': {exc}") from exc


def assert_condition(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def can_submit_to_backend(intent: dict[str, Any]) -> bool:
    safety = intent.get("safety") if isinstance(intent.get("safety"), dict) else {}
    return safety.get("requires_confirmation") is not True


def check_image_generation_intent(intent: dict[str, Any]) -> None:
    output = intent.get("output") if isinstance(intent.get("output"), dict) else {}
    prompt = intent.get("prompt") if isinstance(intent.get("prompt"), dict) else {}
    safety = intent.get("safety") if isinstance(intent.get("safety"), dict) else {}

    assert_condition(intent.get("schema_version") == 1, "image generation intent schema_version must be 1")
    assert_condition(intent.get("kind") == "image_generation", "image generation intent kind mismatch")
    assert_condition(int(output.get("width") or 0) > 0, "image generation output.width must be positive")
    assert_condition(int(output.get("height") or 0) > 0, "image generation output.height must be positive")
    assert_condition(int(output.get("count") or 0) >= 1, "image generation output.count must be at least 1")
    assert_condition(str(prompt.get("positive") or "").strip() != "", "image generation prompt.positive must be non-empty")
    assert_condition(safety.get("risk_level") in ALLOWED_RISK_LEVELS, "image generation safety.risk_level mismatch")

    if safety.get("requires_confirmation") is True:
        assert_condition(
            not can_submit_to_backend(intent),
            "image generation intent requiring confirmation must not be backend-submittable",
        )


def check_backend_request(intent: dict[str, Any], backend_request: dict[str, Any], *, expected_gate: str) -> None:
    intent_backend = intent.get("backend") if isinstance(intent.get("backend"), dict) else {}
    intent_output = intent.get("output") if isinstance(intent.get("output"), dict) else {}
    intent_style = intent.get("style") if isinstance(intent.get("style"), dict) else {}
    intent_constraints = intent.get("constraints") if isinstance(intent.get("constraints"), dict) else {}
    intent_safety = intent.get("safety") if isinstance(intent.get("safety"), dict) else {}
    backend_safety = backend_request.get("safety", {})

    assert_condition(
        backend_request.get("kind") == "image_generation_backend_request",
        "image generation backend request kind mismatch",
    )
    assert_condition(backend_request.get("intent_id") == intent.get("intent_id"), "backend request intent_id mismatch")
    assert_condition(
        backend_request.get("backend", {}).get("id") == intent_backend.get("preferred"),
        "backend request backend.id must come from intent backend.preferred",
    )
    assert_condition(
        backend_request.get("parameters", {}).get("seed") == intent_backend.get("seed"),
        "backend request seed must come from intent",
    )
    assert_condition(
        backend_request.get("parameters", {}).get("steps") == intent_backend.get("steps"),
        "backend request steps must come from intent",
    )
    assert_condition(
        backend_request.get("parameters", {}).get("guidance_scale") == intent_backend.get("guidance_scale"),
        "backend request guidance_scale must come from intent",
    )
    for key in ("width", "height", "count", "format"):
        assert_condition(
            backend_request.get("output", {}).get(key) == intent_output.get(key),
            f"backend request output.{key} must match intent",
        )
    assert_condition(
        backend_request.get("inputs", {}).get("reference_artifact_ids") == intent_style.get("reference_artifact_ids"),
        "backend request reference artifacts must come from intent style",
    )
    assert_condition(
        backend_request.get("inputs", {}).get("edit_artifact_id") == intent_constraints.get("edit_artifact_id"),
        "backend request edit artifact must come from intent constraints",
    )
    assert_condition(
        backend_request.get("inputs", {}).get("mask_artifact_id") == intent_constraints.get("mask_artifact_id"),
        "backend request mask artifact must come from intent constraints",
    )
    assert_condition(
        backend_request.get("constraints", {}).get("style_preset") == intent_style.get("preset"),
        "backend request style preset must come from intent style",
    )
    assert_condition(
        backend_request.get("constraints", {}).get("must_include") == intent_constraints.get("must_include"),
        "backend request must_include must come from intent constraints",
    )
    assert_condition(
        backend_request.get("constraints", {}).get("must_avoid") == intent_constraints.get("must_avoid"),
        "backend request must_avoid must come from intent constraints",
    )
    assert_condition(
        backend_safety.get("requires_confirmation") == intent_safety.get("requires_confirmation"),
        "backend request safety confirmation must match intent",
    )
    assert_condition(
        backend_safety.get("risk_level") == intent_safety.get("risk_level"),
        "backend request safety risk must match intent",
    )
    assert_condition(
        backend_safety.get("gate") == expected_gate,
        f"backend request safety gate must be {expected_gate}",
    )
    assert_condition(
        intent.get("source_request_id") in backend_request.get("trace", {}).get("trace_ids", []),
        "backend request trace must include source request id",
    )
    assert_condition(
        intent.get("intent_id") in backend_request.get("trace", {}).get("trace_ids", []),
        "backend request trace must include intent id",
    )


def check_artifact(intent: dict[str, Any], backend_request: dict[str, Any], artifact: dict[str, Any]) -> None:
    intent_metadata = intent.get("artifact_metadata") if isinstance(intent.get("artifact_metadata"), dict) else {}
    backend_output = backend_request.get("output") if isinstance(backend_request.get("output"), dict) else {}
    backend_parameters = backend_request.get("parameters") if isinstance(backend_request.get("parameters"), dict) else {}
    backend = backend_request.get("backend") if isinstance(backend_request.get("backend"), dict) else {}

    assert_condition(artifact.get("kind") == "image_generation_artifact", "image generation artifact kind mismatch")
    assert_condition(artifact.get("intent_id") == intent.get("intent_id"), "artifact intent_id mismatch")
    assert_condition(
        artifact.get("backend_request_id") == backend_request.get("request_id"),
        "artifact backend_request_id mismatch",
    )
    assert_condition(artifact.get("status") == "generated", "basic artifact must be generated")
    for key in ("width", "height", "format"):
        assert_condition(
            artifact.get("artifact", {}).get(key) == backend_output.get(key),
            f"artifact {key} must match backend request output",
        )
    assert_condition(
        artifact.get("artifact", {}).get("purpose") == intent_metadata.get("purpose"),
        "artifact purpose must come from intent metadata",
    )
    assert_condition(
        artifact.get("artifact", {}).get("title") == intent_metadata.get("proposed_title"),
        "artifact title must come from intent metadata",
    )
    assert_condition(
        artifact.get("generation", {}).get("backend_id") == backend.get("id"),
        "artifact backend_id must come from backend request",
    )
    assert_condition(
        artifact.get("generation", {}).get("model") == backend.get("model"),
        "artifact model must come from backend request",
    )
    for key in ("seed", "steps", "guidance_scale"):
        assert_condition(
            artifact.get("generation", {}).get(key) == backend_parameters.get(key),
            f"artifact generation.{key} must match backend request parameters",
        )
    assert_condition(
        artifact.get("safety", {}).get("requires_confirmation") == backend_request.get("safety", {}).get("requires_confirmation"),
        "artifact safety confirmation must match backend request",
    )
    assert_condition(
        artifact.get("safety", {}).get("risk_level") == backend_request.get("safety", {}).get("risk_level"),
        "artifact safety risk must match backend request",
    )
    provenance = artifact.get("provenance") if isinstance(artifact.get("provenance"), dict) else {}
    for expected_trace in (
        intent.get("source_request_id"),
        intent.get("intent_id"),
        backend_request.get("request_id"),
    ):
        assert_condition(expected_trace in provenance.get("trace_ids", []), f"artifact provenance missing {expected_trace}")


def with_confirmation_required(intent: dict[str, Any]) -> dict[str, Any]:
    return {
        **intent,
        "safety": {
            **intent["safety"],
            "requires_confirmation": True,
            "risk_level": "medium",
            "review_notes": ["requires manual review before backend submission"],
        },
    }

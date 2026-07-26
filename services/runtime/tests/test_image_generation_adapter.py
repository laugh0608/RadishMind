from __future__ import annotations

import copy
import json
import sys
import unittest
from collections.abc import Callable, Mapping
from dataclasses import replace
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.image_backend_profile_configuration import (  # noqa: E402
    ImageBackendProfile,
    compile_image_backend_profile,
)
from services.runtime.image_generation_adapter import (  # noqa: E402
    FAILURE_BACKEND_ARTIFACT_LINEAGE,
    FAILURE_BACKEND_PROFILE_INVALID,
    FAILURE_BACKEND_PROFILE_MISMATCH,
    FAILURE_BACKEND_RESPONSE_UNTRUSTED,
    FAILURE_BACKEND_SAFETY_BLOCKED,
    FAILURE_BACKEND_TIMEOUT,
    FAILURE_BACKEND_UNAVAILABLE,
    FAILURE_INTENT_BUDGET_EXCEEDED,
    FAILURE_INTENT_HIGH_RISK,
    FAILURE_INTENT_INVALID,
    FAILURE_INTENT_REQUIRES_CONFIRMATION,
    FAILURE_INTENT_SENSITIVE_MATERIAL,
    ImageBackendInvocationResult,
    adapter_side_effect_counters,
    compile_image_backend_request,
    invoke_image_generation,
)


INTENT_FIXTURE = REPO_ROOT / "scripts/checks/fixtures/image-generation-intent-basic.json"
PROFILE_FIXTURE = REPO_ROOT / "scripts/checks/fixtures/image-backend-profile-source-basic.json"
BACKEND_REQUEST_ID = "image-backend-request-runtime-001"
OBSERVED_SHA256 = "a" * 64


class RecordingClient:
    def __init__(
        self,
        result_factory: Callable[[Mapping[str, Any]], ImageBackendInvocationResult] | None = None,
        error: Exception | None = None,
    ) -> None:
        self.result_factory = result_factory or successful_backend_result
        self.error = error
        self.calls: list[tuple[dict[str, Any], float]] = []

    def invoke(
        self,
        request_document: Mapping[str, Any],
        timeout_seconds: float,
    ) -> ImageBackendInvocationResult:
        request_copy = copy.deepcopy(dict(request_document))
        self.calls.append((request_copy, timeout_seconds))
        if self.error is not None:
            raise self.error
        return self.result_factory(request_copy)


def load_intent() -> dict[str, Any]:
    return json.loads(INTENT_FIXTURE.read_text(encoding="utf-8"))


def load_profile(
    update: Callable[[dict[str, Any]], None] | None = None,
) -> ImageBackendProfile:
    source = json.loads(PROFILE_FIXTURE.read_text(encoding="utf-8"))
    if update is not None:
        update(source)
    result = compile_image_backend_profile(source)
    if not result.ok or result.profile is None:
        raise AssertionError(result.failure_message)
    return result.profile


PROFILE = load_profile()


def successful_backend_result(request_document: Mapping[str, Any]) -> ImageBackendInvocationResult:
    return ImageBackendInvocationResult(
        artifact_id="image-artifact-runtime-001",
        created_at="2026-07-25T00:00:00Z",
        observed_sha256=OBSERVED_SHA256,
        observed_mime_type="image/png",
        observed_width=request_document["output"]["width"],
        observed_height=request_document["output"]["height"],
    )


def invoke(
    intent: Mapping[str, Any],
    client: RecordingClient,
    *,
    profile: ImageBackendProfile | None = PROFILE,
    backend_request_id: str = BACKEND_REQUEST_ID,
):
    return invoke_image_generation(
        intent,
        profile=profile,
        client=client,
        backend_request_id=backend_request_id,
    )


class ImageGenerationAdapterTest(unittest.TestCase):
    def test_compiler_is_deterministic_and_preserves_canonical_lineage(self) -> None:
        intent = load_intent()

        first = compile_image_backend_request(
            intent,
            profile=PROFILE,
            backend_request_id=BACKEND_REQUEST_ID,
        )
        second = compile_image_backend_request(
            copy.deepcopy(intent),
            profile=PROFILE,
            backend_request_id=BACKEND_REQUEST_ID,
        )

        self.assertEqual(first, second)
        self.assertEqual(first["backend"]["id"], intent["backend"]["preferred"])
        self.assertEqual(first["prompt"]["transformed_from_intent"], False)
        self.assertEqual(
            first["trace"]["trace_ids"],
            [
                intent["source_request_id"],
                intent["intent_id"],
                BACKEND_REQUEST_ID,
            ],
        )
        self.assertNotIn("credential", json.dumps(first))
        self.assertNotIn("endpoint", json.dumps(first))

    def test_success_calls_backend_once_and_returns_existing_artifact_reference_shape(self) -> None:
        client = RecordingClient()

        result = invoke(load_intent(), client)

        self.assertTrue(result.ok)
        self.assertEqual(len(client.calls), 1)
        self.assertEqual(client.calls[0][1], 30)
        self.assertEqual(result.citation["kind"], "artifact")
        self.assertEqual(result.citation["locator"], "artifact://radishmind/generated/image-artifact-runtime-001.png")
        self.assertEqual(result.metadata_reference["sha256"], OBSERVED_SHA256)
        self.assertEqual(
            result.artifact_document["artifact"]["title"],
            load_intent()["artifact_metadata"]["proposed_title"],
        )
        self.assertEqual(
            result.artifact_document["provenance"]["trace_ids"],
            result.backend_request["trace"]["trace_ids"],
        )
        self.assertEqual(result.backend_call_count, 1)
        self.assertEqual(result.image_generation_count, 1)
        counters = adapter_side_effect_counters(result)
        self.assertEqual(counters["artifact_store_lookup_count"], 0)
        self.assertEqual(counters["artifact_binary_read_count"], 0)
        self.assertEqual(counters["artifact_upload_count"], 0)
        self.assertEqual(counters["public_url_resolution_count"], 0)
        self.assertEqual(counters["retry_count"], 0)
        self.assertEqual(counters["fallback_count"], 0)

    def test_confirmation_and_risk_gates_stop_before_backend_call(self) -> None:
        cases = (
            (
                "confirmation",
                {"requires_confirmation": True, "risk_level": "low"},
                FAILURE_INTENT_REQUIRES_CONFIRMATION,
            ),
            (
                "medium",
                {"requires_confirmation": False, "risk_level": "medium"},
                FAILURE_BACKEND_SAFETY_BLOCKED,
            ),
            (
                "high",
                {"requires_confirmation": False, "risk_level": "high"},
                FAILURE_INTENT_HIGH_RISK,
            ),
        )
        for name, safety_update, expected_failure in cases:
            with self.subTest(name=name):
                intent = load_intent()
                intent["safety"].update(safety_update)
                client = RecordingClient()

                result = invoke(intent, client)

                self.assertFalse(result.ok)
                self.assertEqual(result.failure_code, expected_failure)
                self.assertEqual(result.backend_call_count, 0)
                self.assertEqual(client.calls, [])

    def test_unknown_fields_and_noncanonical_identifiers_fail_before_call(self) -> None:
        cases = []
        unknown = load_intent()
        unknown["provider_config"] = {"token": "not-allowed"}
        cases.append(("unknown", unknown))
        missing_id = load_intent()
        del missing_id["intent_id"]
        cases.append(("missing-id", missing_id))
        invalid_id = load_intent()
        invalid_id["intent_id"] = "../escape"
        cases.append(("invalid-id", invalid_id))
        invalid_locale = load_intent()
        invalid_locale["prompt"]["locale"] = "ZH_cn"
        cases.append(("invalid-locale", invalid_locale))
        invalid_trace = load_intent()
        invalid_trace["artifact_metadata"]["trace_ids"] = ["../trace"]
        cases.append(("invalid-trace", invalid_trace))
        invalid_utf8 = load_intent()
        invalid_utf8["prompt"]["positive"] = "\ud800"
        cases.append(("invalid-utf8", invalid_utf8))

        for name, intent in cases:
            with self.subTest(name=name):
                client = RecordingClient()
                result = invoke(intent, client)
                self.assertEqual(result.failure_code, FAILURE_INTENT_INVALID)
                self.assertEqual(client.calls, [])

    def test_utf8_output_parameter_and_collection_budgets_are_enforced(self) -> None:
        cases = []
        prompt = load_intent()
        prompt["prompt"]["positive"] = "萝" * 4_001
        cases.append(("utf8-prompt", prompt))
        dimension = load_intent()
        dimension["output"]["width"] = 2_049
        cases.append(("dimension", dimension))
        pixels = load_intent()
        pixels["output"].update({"width": 2_048, "height": 2_048})
        cases.append(("pixel-boundary-valid", pixels))
        count = load_intent()
        count["output"]["count"] = 2
        cases.append(("count", count))
        steps = load_intent()
        steps["backend"]["steps"] = 101
        cases.append(("steps", steps))
        guidance = load_intent()
        guidance["backend"]["guidance_scale"] = 30.1
        cases.append(("guidance", guidance))
        refs = load_intent()
        refs["style"]["reference_artifact_ids"] = [f"artifact-{index}" for index in range(5)]
        cases.append(("references", refs))

        for name, intent in cases:
            with self.subTest(name=name):
                client = RecordingClient()
                result = invoke(intent, client)
                if name == "pixel-boundary-valid":
                    self.assertTrue(result.ok)
                    self.assertEqual(len(client.calls), 1)
                else:
                    self.assertEqual(result.failure_code, FAILURE_INTENT_BUDGET_EXCEEDED)
                    self.assertEqual(client.calls, [])

    def test_sensitive_material_is_rejected_before_call(self) -> None:
        samples = (
            "Authorization: Bearer secret-token-value",
            "api_key=secret-value",
            "token=secret-value",
            "headers: X-Provider-Raw",
            "endpoint=private-backend",
            "dsn=private-database",
            "model_dir=/private/model",
            "provider_config=raw",
            "postgresql://user:pass@db.example/database",
            "https://private.example/image",
            "-----BEGIN PRIVATE KEY-----",
        )
        for sample in samples:
            with self.subTest(sample=sample):
                intent = load_intent()
                intent["prompt"]["positive"] = sample
                client = RecordingClient()

                result = invoke(intent, client)

                self.assertEqual(result.failure_code, FAILURE_INTENT_SENSITIVE_MATERIAL)
                self.assertEqual(client.calls, [])
                self.assertNotIn(sample, result.failure_message)

    def test_profile_mismatch_does_not_fallback(self) -> None:
        client = RecordingClient()
        profile = load_profile(
            lambda source: source["backend"].update(
                {
                    "id": "other-backend",
                    "model": "other-model",
                    "adapter_profile": "other-profile",
                }
            )
        )

        result = invoke(load_intent(), client, profile=profile)

        self.assertEqual(result.failure_code, FAILURE_BACKEND_PROFILE_MISMATCH)
        self.assertEqual(client.calls, [])
        self.assertEqual(adapter_side_effect_counters(result)["fallback_count"], 0)

    def test_profile_digest_drift_is_rejected_before_backend_call(self) -> None:
        client = RecordingClient()
        drifted = replace(PROFILE, profile_digest="sha256:" + "f" * 64)

        result = invoke(load_intent(), client, profile=drifted)

        self.assertEqual(result.failure_code, FAILURE_BACKEND_PROFILE_INVALID)
        self.assertEqual(client.calls, [])

        wrong_type = invoke(load_intent(), client, profile={})  # type: ignore[arg-type]
        self.assertEqual(wrong_type.failure_code, FAILURE_BACKEND_PROFILE_INVALID)
        self.assertEqual(client.calls, [])

    def test_timeout_unavailable_and_untrusted_errors_call_once_without_retry(self) -> None:
        cases = (
            ("timeout", TimeoutError("timeout-provider-secret"), FAILURE_BACKEND_TIMEOUT),
            ("unavailable", ConnectionError("endpoint-provider-secret"), FAILURE_BACKEND_UNAVAILABLE),
            ("untrusted", RuntimeError("provider raw secret"), FAILURE_BACKEND_RESPONSE_UNTRUSTED),
        )
        for name, error, expected_failure in cases:
            with self.subTest(name=name):
                client = RecordingClient(error=error)

                result = invoke(load_intent(), client)

                self.assertEqual(result.failure_code, expected_failure)
                self.assertEqual(result.backend_call_count, 1)
                self.assertEqual(len(client.calls), 1)
                self.assertEqual(adapter_side_effect_counters(result)["retry_count"], 0)
                self.assertNotIn(str(error), result.failure_message)
                self.assertIsNone(result.backend_request)
                self.assertIsNone(result.artifact_document)

    def test_backend_outcome_identity_schema_and_lineage_drift_fail_closed(self) -> None:
        def observed(
            **updates: Any,
        ) -> Callable[[Mapping[str, Any]], ImageBackendInvocationResult]:
            def factory(request_document: Mapping[str, Any]) -> ImageBackendInvocationResult:
                result = successful_backend_result(request_document)
                values = {
                    "artifact_id": result.artifact_id,
                    "created_at": result.created_at,
                    "observed_sha256": result.observed_sha256,
                    "observed_mime_type": result.observed_mime_type,
                    "observed_width": result.observed_width,
                    "observed_height": result.observed_height,
                    **updates,
                }
                return ImageBackendInvocationResult(**values)

            return factory

        cases = (
            (
                "identity",
                observed(artifact_id="../artifact"),
                FAILURE_BACKEND_RESPONSE_UNTRUSTED,
            ),
            (
                "timestamp",
                observed(created_at="not-a-timestamp"),
                FAILURE_BACKEND_RESPONSE_UNTRUSTED,
            ),
            (
                "width",
                observed(observed_width=512),
                FAILURE_BACKEND_ARTIFACT_LINEAGE,
            ),
            (
                "mime",
                observed(observed_mime_type="image/jpeg"),
                FAILURE_BACKEND_ARTIFACT_LINEAGE,
            ),
        )
        for name, factory, expected_failure in cases:
            with self.subTest(name=name):
                client = RecordingClient(result_factory=factory)
                result = invoke(load_intent(), client)
                self.assertEqual(result.failure_code, expected_failure)
                self.assertEqual(len(client.calls), 1)
                self.assertFalse(result.ok)
                self.assertIsNone(result.backend_request)
                self.assertIsNone(result.artifact_document)

    def test_transport_observations_must_be_canonical(self) -> None:
        def observed(**updates: Any) -> Callable[[Mapping[str, Any]], ImageBackendInvocationResult]:
            def factory(request_document: Mapping[str, Any]) -> ImageBackendInvocationResult:
                result = successful_backend_result(request_document)
                values = {
                    "artifact_id": result.artifact_id,
                    "created_at": result.created_at,
                    "observed_sha256": result.observed_sha256,
                    "observed_mime_type": result.observed_mime_type,
                    "observed_width": result.observed_width,
                    "observed_height": result.observed_height,
                    **updates,
                }
                return ImageBackendInvocationResult(**values)

            return factory

        cases = (
            (
                "hash",
                observed(observed_sha256="not-a-digest"),
                FAILURE_BACKEND_RESPONSE_UNTRUSTED,
            ),
            (
                "mime",
                observed(observed_mime_type="application/octet-stream"),
                FAILURE_BACKEND_RESPONSE_UNTRUSTED,
            ),
            (
                "width",
                observed(observed_width=0),
                FAILURE_BACKEND_RESPONSE_UNTRUSTED,
            ),
        )
        for name, factory, expected_failure in cases:
            with self.subTest(name=name):
                client = RecordingClient(result_factory=factory)
                result = invoke(load_intent(), client)
                self.assertEqual(result.failure_code, expected_failure)
                self.assertEqual(len(client.calls), 1)

    def test_mask_requires_edit_artifact(self) -> None:
        intent = load_intent()
        intent["constraints"]["mask_artifact_id"] = "mask-001"
        client = RecordingClient()

        result = invoke(intent, client)

        self.assertEqual(result.failure_code, FAILURE_INTENT_INVALID)
        self.assertEqual(client.calls, [])


if __name__ == "__main__":
    unittest.main()

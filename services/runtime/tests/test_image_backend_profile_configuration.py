from __future__ import annotations

import copy
import json
import math
import sys
import unittest
from dataclasses import replace
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.image_backend_profile_configuration import (  # noqa: E402
    FAILURE_PROFILE_BINDING_INVALID,
    FAILURE_PROFILE_ENVIRONMENT_FORBIDDEN,
    FAILURE_PROFILE_SENSITIVE_MATERIAL,
    FAILURE_PROFILE_SOURCE_BUDGET,
    FAILURE_PROFILE_SOURCE_INVALID,
    MAX_PROFILE_SOURCE_BYTES,
    compile_image_backend_profile,
    compiled_profile_is_valid,
    profile_to_reference_document,
)


PROFILE_FIXTURE = REPO_ROOT / "scripts/checks/fixtures/image-backend-profile-source-basic.json"


def load_source() -> dict[str, Any]:
    return json.loads(PROFILE_FIXTURE.read_text(encoding="utf-8"))


def compile_profile(source: dict[str, Any]):
    result = compile_image_backend_profile(source)
    if not result.ok or result.profile is None:
        raise AssertionError(f"{result.failure_code}: {result.failure_message}")
    return result.profile


class ImageBackendProfileConfigurationTest(unittest.TestCase):
    def test_remote_profile_compilation_is_deterministic_and_canonical(self) -> None:
        source = load_source()
        reordered = {
            "limits": copy.deepcopy(source["limits"]),
            "credential": copy.deepcopy(source["credential"]),
            "runtime": copy.deepcopy(source["runtime"]),
            "backend": copy.deepcopy(source["backend"]),
            "enabled": source["enabled"],
            "environment": source["environment"],
            "profile_version": source["profile_version"],
            "profile_id": source["profile_id"],
            "kind": source["kind"],
            "schema_version": source["schema_version"],
        }

        first = compile_profile(source)
        second = compile_profile(reordered)

        self.assertEqual(first, second)
        self.assertTrue(first.profile_digest.startswith("sha256:"))
        self.assertEqual(len(first.profile_digest), 71)
        self.assertEqual(first.runtime_mode, "remote_https")
        self.assertEqual(first.timeout_seconds, 30.0)
        self.assertTrue(compiled_profile_is_valid(first))

        reference_document = profile_to_reference_document(first)
        self.assertEqual(reference_document["profile_digest"], first.profile_digest)
        self.assertNotIn("endpoint", reference_document["runtime"])
        self.assertNotIn("credential_value", json.dumps(reference_document))
        self.assertNotIn("model_dir", reference_document["runtime"])

    def test_local_model_profile_requires_model_dir_reference_without_credential(self) -> None:
        source = load_source()
        source["profile_id"] = "image.diagram.local"
        source["runtime"] = {
            "mode": "local_model",
            "endpoint_ref": None,
            "model_dir_ref": "ref:radishmind/test/image-backends/model-dirs/diagram-default",
        }
        source["credential"] = {
            "requirement": "not_required",
            "secret_ref": None,
        }

        profile = compile_profile(source)

        self.assertEqual(profile.runtime_mode, "local_model")
        self.assertIsNone(profile.endpoint_ref)
        self.assertIsNone(profile.secret_ref)
        self.assertTrue(compiled_profile_is_valid(profile))

    def test_contract_fixture_profile_is_test_only_and_has_no_runtime_reference(self) -> None:
        source = load_source()
        source["profile_id"] = "image.diagram.fixture"
        source["runtime"] = {
            "mode": "contract_fixture",
            "endpoint_ref": None,
            "model_dir_ref": None,
        }
        source["credential"] = {
            "requirement": "not_required",
            "secret_ref": None,
        }

        profile = compile_profile(source)

        self.assertEqual(profile.runtime_mode, "contract_fixture")
        self.assertTrue(compiled_profile_is_valid(profile))
        self.assertIsNone(profile.endpoint_ref)
        self.assertIsNone(profile.model_dir_ref)
        self.assertIsNone(profile.secret_ref)

        source["environment"] = "development"
        result = compile_image_backend_profile(source)
        self.assertEqual(result.failure_code, FAILURE_PROFILE_BINDING_INVALID)

    def test_strict_schema_rejects_unknown_missing_and_non_json_values(self) -> None:
        unknown = load_source()
        unknown["fallback_profile"] = "other"
        missing = load_source()
        del missing["backend"]["adapter_profile"]
        non_json = load_source()
        non_json["limits"]["timeout_seconds"] = object()
        invalid_utf8 = load_source()
        invalid_utf8["profile_id"] = "\ud800"

        for name, source in (
            ("unknown", unknown),
            ("missing", missing),
            ("non-json", non_json),
            ("invalid-utf8", invalid_utf8),
        ):
            with self.subTest(name=name):
                result = compile_image_backend_profile(source)
                self.assertEqual(result.failure_code, FAILURE_PROFILE_SOURCE_INVALID)
                self.assertIsNone(result.profile)

    def test_only_development_and_test_environments_are_allowed(self) -> None:
        for environment in ("production", "staging", "", None):
            with self.subTest(environment=environment):
                source = load_source()
                source["environment"] = environment
                result = compile_image_backend_profile(source)
                self.assertEqual(
                    result.failure_code,
                    FAILURE_PROFILE_ENVIRONMENT_FORBIDDEN,
                )

    def test_sensitive_and_concrete_runtime_material_is_rejected(self) -> None:
        cases = []
        for key, value in (
            ("api_key", "secret-value"),
            ("token", "secret-value"),
            ("authorization", "Bearer secret-token-value"),
            ("headers", {"X-Api-Key": "secret-value"}),
            ("endpoint_url", "https://private.example/image"),
            ("dsn", "postgresql://user:pass@db/database"),
            ("model_dir", "/private/models/image"),
            ("system_prompt", "unbounded provider instructions"),
            ("provider_config", {"raw": "value"}),
            ("runtime_config", {"raw": "value"}),
        ):
            source = load_source()
            source[key] = value
            cases.append((key, source))

        raw_endpoint = load_source()
        raw_endpoint["runtime"]["endpoint_ref"] = "https://private.example/image"
        cases.append(("raw-endpoint", raw_endpoint))
        raw_model_path = load_source()
        raw_model_path["runtime"]["model_dir_ref"] = "/private/models/image"
        cases.append(("raw-model-path", raw_model_path))
        raw_secret = load_source()
        raw_secret["credential"]["secret_ref"] = "sk-secret-secret-secret"
        cases.append(("raw-secret", raw_secret))
        marker_in_model = load_source()
        marker_in_model["backend"]["model"] = "token=secret-value"
        cases.append(("marker-in-model", marker_in_model))

        for name, source in cases:
            with self.subTest(name=name):
                result = compile_image_backend_profile(source)
                self.assertEqual(
                    result.failure_code,
                    FAILURE_PROFILE_SENSITIVE_MATERIAL,
                )
                self.assertNotIn("secret-value", result.failure_message)
                self.assertIsNone(result.profile)

    def test_runtime_and_credential_bindings_are_mutually_exclusive(self) -> None:
        cases = []
        missing_endpoint = load_source()
        missing_endpoint["runtime"]["endpoint_ref"] = None
        cases.append(("remote-missing-endpoint", missing_endpoint))
        remote_model_dir = load_source()
        remote_model_dir["runtime"]["model_dir_ref"] = (
            "ref:radishmind/test/image-backends/model-dirs/diagram-default"
        )
        cases.append(("remote-model-dir", remote_model_dir))
        remote_without_credential = load_source()
        remote_without_credential["credential"] = {
            "requirement": "not_required",
            "secret_ref": None,
        }
        cases.append(("remote-no-credential", remote_without_credential))

        local_with_endpoint = load_source()
        local_with_endpoint["runtime"]["mode"] = "local_model"
        local_with_endpoint["runtime"]["model_dir_ref"] = (
            "ref:radishmind/test/image-backends/model-dirs/diagram-default"
        )
        cases.append(("local-with-endpoint", local_with_endpoint))
        local_with_credential = load_source()
        local_with_credential["runtime"] = {
            "mode": "local_model",
            "endpoint_ref": None,
            "model_dir_ref": "ref:radishmind/test/image-backends/model-dirs/diagram-default",
        }
        cases.append(("local-with-credential", local_with_credential))

        for name, source in cases:
            with self.subTest(name=name):
                result = compile_image_backend_profile(source)
                self.assertEqual(result.failure_code, FAILURE_PROFILE_BINDING_INVALID)

    def test_references_are_environment_scoped_and_kind_scoped(self) -> None:
        cases = []
        cross_environment = load_source()
        cross_environment["runtime"]["endpoint_ref"] = (
            "ref:radishmind/development/image-backends/endpoints/diagram-default"
        )
        cases.append(("cross-environment", cross_environment))
        wrong_endpoint_kind = load_source()
        wrong_endpoint_kind["runtime"]["endpoint_ref"] = (
            "ref:radishmind/test/image-backends/credentials/diagram-default"
        )
        cases.append(("endpoint-kind", wrong_endpoint_kind))
        wrong_secret_kind = load_source()
        wrong_secret_kind["credential"]["secret_ref"] = (
            "ref:radishmind/test/image-backends/endpoints/diagram-default"
        )
        cases.append(("secret-kind", wrong_secret_kind))
        traversal = load_source()
        traversal["runtime"]["endpoint_ref"] = (
            "ref:radishmind/test/image-backends/endpoints/../credential"
        )
        cases.append(("traversal", traversal))
        empty_key = load_source()
        empty_key["runtime"]["endpoint_ref"] = (
            "ref:radishmind/test/image-backends/endpoints/"
        )
        cases.append(("empty-key", empty_key))

        for name, source in cases:
            with self.subTest(name=name):
                result = compile_image_backend_profile(source)
                self.assertEqual(result.failure_code, FAILURE_PROFILE_BINDING_INVALID)

    def test_timeout_and_source_byte_budgets_cover_boundaries(self) -> None:
        for timeout in (1, 300, 1.5):
            with self.subTest(valid_timeout=timeout):
                source = load_source()
                source["limits"]["timeout_seconds"] = timeout
                profile = compile_profile(source)
                self.assertEqual(profile.timeout_seconds, float(timeout))

        for timeout in (0, 300.1, math.nan, math.inf, True):
            with self.subTest(invalid_timeout=timeout):
                source = load_source()
                source["limits"]["timeout_seconds"] = timeout
                result = compile_image_backend_profile(source)
                self.assertEqual(result.failure_code, FAILURE_PROFILE_SOURCE_INVALID)

        oversized = load_source()
        oversized["padding"] = "x" * MAX_PROFILE_SOURCE_BYTES
        result = compile_image_backend_profile(oversized)
        self.assertEqual(result.failure_code, FAILURE_PROFILE_SOURCE_BUDGET)

    def test_digest_changes_with_policy_and_detects_manual_drift(self) -> None:
        first = compile_profile(load_source())
        changed_source = load_source()
        changed_source["limits"]["timeout_seconds"] = 31
        changed = compile_profile(changed_source)

        self.assertNotEqual(first.profile_digest, changed.profile_digest)
        self.assertFalse(
            compiled_profile_is_valid(
                replace(first, profile_digest="sha256:" + "f" * 64)
            )
        )
        self.assertFalse(
            compiled_profile_is_valid(
                replace(first, endpoint_ref="https://private.example/image")
            )
        )

    def test_compiler_does_not_mutate_source_or_enable_fallback(self) -> None:
        source = load_source()
        original = copy.deepcopy(source)

        profile = compile_profile(source)

        self.assertEqual(source, original)
        document = profile_to_reference_document(profile)
        serialized = json.dumps(document, sort_keys=True)
        self.assertNotIn("fallback", serialized)
        self.assertNotIn("retry", serialized)
        self.assertNotIn("https://", serialized)
        self.assertNotIn("/private/", serialized)


if __name__ == "__main__":
    unittest.main()

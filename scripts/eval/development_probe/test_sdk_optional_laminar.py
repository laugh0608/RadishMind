"""Run only in the authorized variant image, without network or SDK tools."""
import importlib.metadata as metadata
import importlib.util
import os
import signal
import sys
import unittest
from unittest.mock import patch


class OptionalLaminarTests(unittest.TestCase):
    def test_default_distribution_and_extra_metadata(self):
        from packaging.requirements import Requirement

        sdk = metadata.distribution("openhands-sdk")
        self.assertEqual(sdk.version, "1.59.0+radishmind.1")
        self.assertEqual(sdk.metadata["License-Expression"], "MIT")
        for name in ("lmnr", "lmnr-claude-code-proxy"):
            with self.assertRaises(metadata.PackageNotFoundError):
                metadata.distribution(name)
        self.assertIsNone(importlib.util.find_spec("lmnr"))
        requirements = [Requirement(value) for value in sdk.requires or []]
        laminar = [value for value in requirements if value.name == "lmnr"]
        self.assertEqual(len(laminar), 1)
        self.assertFalse(laminar[0].marker.evaluate({"extra": ""}))
        self.assertTrue(laminar[0].marker.evaluate({"extra": "laminar"}))
        files = {str(path) for path in sdk.files or []}
        self.assertTrue(any(path.endswith("/licenses/LICENSE") for path in files))
        self.assertTrue(any(path.endswith("/licenses/RADISHMIND-CHANGES.txt") for path in files))

    def test_sdk_import_and_plain_observation_need_no_laminar(self):
        from openhands.sdk import Agent, Conversation, LLM
        from openhands.sdk.observability.laminar import observe, should_enable_observability

        self.assertTrue(all(value is not None for value in (Agent, Conversation, LLM)))
        with patch.dict(os.environ, {}, clear=True):
            self.assertFalse(should_enable_observability())
            self.assertEqual(observe()(lambda value: value + 1)(2), 3)
        self.assertNotIn("lmnr", sys.modules)

    def test_requested_observation_fails_before_initialization(self):
        from openhands.sdk.observability.laminar import maybe_init_laminar, should_enable_observability

        for key in ("LMNR_PROJECT_API_KEY", "OTEL_ENDPOINT", "OTEL_EXPORTER_OTLP_ENDPOINT",
                    "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"):
            with self.subTest(key=key), patch.dict(os.environ, {key: "synthetic"}, clear=True):
                for operation in (should_enable_observability, maybe_init_laminar):
                    with self.assertRaisesRegex(RuntimeError, r"openhands-sdk\[laminar\]"):
                        operation()
            with patch.dict(os.environ, {}, clear=True):
                self.assertFalse(should_enable_observability())
        self.assertNotIn("lmnr", sys.modules)

    def test_direct_span_reports_missing_optional_dependency(self):
        from openhands.sdk.observability.laminar import RootSpan

        with self.assertRaisesRegex(RuntimeError, r"openhands-sdk\[laminar\]"):
            RootSpan("synthetic")


if __name__ == "__main__":
    signal.alarm(60)
    unittest.main()

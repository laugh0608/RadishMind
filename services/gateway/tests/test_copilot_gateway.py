from __future__ import annotations

import copy
import json
import sys
import unittest
from pathlib import Path

import jsonschema

REPO_ROOT = Path(__file__).resolve().parents[3]
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.gateway import GatewayOptions, handle_copilot_request, validate_gateway_envelope  # noqa: E402


REQUEST_FIXTURE = REPO_ROOT / "datasets/examples/radishflow-copilot-request-ghost-valve-ambiguous-001.json"


class GatewayTimingMetadataTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        cls.request_document = json.loads(REQUEST_FIXTURE.read_text(encoding="utf-8"))

    def test_mock_provider_reports_valid_timing_metadata(self) -> None:
        envelope = handle_copilot_request(
            copy.deepcopy(self.request_document),
            options=GatewayOptions(provider="mock"),
        )

        validate_gateway_envelope(envelope)
        metadata = envelope["metadata"]
        self.assertIsInstance(metadata["duration_ms"], int)
        self.assertIsInstance(metadata["provider_duration_ms"], int)
        self.assertGreaterEqual(metadata["duration_ms"], metadata["provider_duration_ms"])

    def test_provider_duration_is_required_by_gateway_contract(self) -> None:
        envelope = handle_copilot_request(
            copy.deepcopy(self.request_document),
            options=GatewayOptions(provider="mock"),
        )
        del envelope["metadata"]["provider_duration_ms"]

        with self.assertRaises(jsonschema.ValidationError):
            validate_gateway_envelope(envelope)


if __name__ == "__main__":
    unittest.main()

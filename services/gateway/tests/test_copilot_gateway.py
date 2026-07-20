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
from services.runtime.inference_support import make_mock_docs_qa_response  # noqa: E402


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

    def test_mock_provider_returns_strict_workflow_rag_answer_for_selected_evidence(self) -> None:
        evidence = [{
            "fragment_ref": "official_guide",
            "rank": 1,
            "source_type": "manual",
            "is_official": True,
            "excerpt": "Only the selected immutable fragment may support the answer.",
            "excerpt_truncated": False,
        }]
        request_document = {
            "schema_version": 1,
            "project": "radish",
            "task": "answer_docs_question",
            "locale": "zh-CN",
            "artifacts": [{
                "kind": "text",
                "role": "primary",
                "name": "northbound_prompt",
                "mime_type": "text/plain",
                "content": "Prepare answer\n\n用户问题：\nWhat is supported?\n\n仅可使用以下已召回证据回答：\n"
                    + json.dumps(evidence)
                    + "\n\n输出且只输出 workflow_rag_answer.v1 JSON",
            }],
            "context": {"northbound": {"protocol": "workflow-rag-retrieval-v1", "request_kind": "workflow-rag-retrieval-v1"}},
        }

        response = make_mock_docs_qa_response(request_document)
        answer = json.loads(response["summary"])

        self.assertEqual(answer["schema_version"], "workflow_rag_answer.v1")
        self.assertEqual(answer["citations"][0]["fragment_ref"], "official_guide")
        self.assertEqual(answer["confidence"], "high")
        self.assertNotIn("excerpt", response["summary"])

    def test_mock_provider_returns_application_rag_answer_for_selected_evidence(self) -> None:
        evidence = [{
            "fragment_ref": "promotion_governance",
            "rank": 1,
            "source_type": "manual",
            "is_official": True,
            "excerpt": "Only the active assignment snapshot may support the answer.",
            "excerpt_truncated": False,
        }]
        protocol = "workflow-rag-application-invocation-v1"
        request_document = {
            "schema_version": 1,
            "project": "radish",
            "task": "answer_docs_question",
            "locale": "zh-CN",
            "artifacts": [{
                "kind": "text",
                "role": "primary",
                "name": "northbound_prompt",
                "mime_type": "text/plain",
                "content": "Prepare answer\n\n用户问题：\nWhat is supported?\n\n仅可使用以下已召回证据回答：\n"
                    + json.dumps(evidence)
                    + "\n\n输出且只输出 workflow_rag_application_answer.v1 JSON",
            }],
            "context": {"northbound": {"protocol": protocol, "request_kind": protocol}},
        }

        response = make_mock_docs_qa_response(request_document)
        answer = json.loads(response["summary"])

        self.assertEqual(answer["schema_version"], "workflow_rag_application_answer.v1")
        self.assertEqual(answer["citations"][0]["fragment_ref"], "promotion_governance")
        self.assertEqual(answer["confidence"], "high")
        self.assertNotIn("excerpt", response["summary"])


if __name__ == "__main__":
    unittest.main()

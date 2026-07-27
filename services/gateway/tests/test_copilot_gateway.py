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
from services.gateway.copilot_gateway import sanitize_provider_usage  # noqa: E402
from services.runtime.inference_provider import normalize_provider_usage  # noqa: E402
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
        self.assertEqual(metadata["usage"]["availability"], "not_reported")
        self.assertEqual(metadata["usage"]["total_tokens"], 0)

    def test_provider_duration_is_required_by_gateway_contract(self) -> None:
        envelope = handle_copilot_request(
            copy.deepcopy(self.request_document),
            options=GatewayOptions(provider="mock"),
        )
        del envelope["metadata"]["provider_duration_ms"]

        with self.assertRaises(jsonschema.ValidationError):
            validate_gateway_envelope(envelope)

    def test_gateway_rejects_reported_usage_with_inconsistent_total(self) -> None:
        envelope = handle_copilot_request(
            copy.deepcopy(self.request_document),
            options=GatewayOptions(provider="mock"),
        )
        envelope["metadata"]["usage"] = {
            "availability": "reported",
            "source": "openai_compatible_usage",
            "input_tokens": 12,
            "output_tokens": 5,
            "total_tokens": 99,
        }

        with self.assertRaises(jsonschema.ValidationError):
            validate_gateway_envelope(envelope)


class ProviderUsageNormalizationTest(unittest.TestCase):
    def test_gateway_sanitizes_invalid_provider_usage_without_preserving_counts(self) -> None:
        self.assertEqual(
            sanitize_provider_usage(
                {
                    "availability": "reported",
                    "source": "openai_compatible_usage",
                    "input_tokens": 3,
                    "output_tokens": 2,
                    "total_tokens": 99,
                }
            ),
            {
                "availability": "not_reported",
                "source": "",
                "input_tokens": 0,
                "output_tokens": 0,
                "total_tokens": 0,
            },
        )

    def test_normalizes_supported_provider_shapes(self) -> None:
        cases = (
            (
                {"usage": {"prompt_tokens": 12, "completion_tokens": 5, "total_tokens": 17}},
                {"provider_id": "openai-compatible"},
                "openai_compatible_usage",
            ),
            (
                {
                    "usageMetadata": {
                        "promptTokenCount": 20,
                        "candidatesTokenCount": 8,
                        "thoughtsTokenCount": 3,
                        "totalTokenCount": 31,
                    }
                },
                {"provider_id": "openai-compatible", "api_style": "gemini-native"},
                "gemini_usage_metadata",
            ),
            (
                {
                    "usage": {
                        "input_tokens": 31,
                        "cache_creation_input_tokens": 4,
                        "cache_read_input_tokens": 6,
                        "output_tokens": 9,
                    }
                },
                {"provider_id": "openai-compatible", "api_style": "anthropic-messages"},
                "anthropic_usage",
            ),
            (
                {"prompt_eval_count": 7, "eval_count": 3},
                {"provider_id": "ollama"},
                "ollama_eval_counts",
            ),
        )

        for raw_response, options, expected_source in cases:
            with self.subTest(source=expected_source):
                usage = normalize_provider_usage(raw_response, **options)
                self.assertEqual(usage["availability"], "reported")
                self.assertEqual(usage["source"], expected_source)
                self.assertEqual(
                    usage["total_tokens"],
                    usage["input_tokens"] + usage["output_tokens"],
                )
        gemini_usage = normalize_provider_usage(
            {
                "usageMetadata": {
                    "promptTokenCount": 20,
                    "candidatesTokenCount": 8,
                    "thoughtsTokenCount": 3,
                    "totalTokenCount": 31,
                }
            },
            provider_id="openai-compatible",
            api_style="gemini-native",
        )
        self.assertEqual(gemini_usage["output_tokens"], 11)
        anthropic_usage = normalize_provider_usage(
            {
                "usage": {
                    "input_tokens": 31,
                    "cache_creation_input_tokens": 4,
                    "cache_read_input_tokens": 6,
                    "output_tokens": 9,
                }
            },
            provider_id="openai-compatible",
            api_style="anthropic-messages",
        )
        self.assertEqual(anthropic_usage["input_tokens"], 41)

    def test_uses_final_valid_stream_usage_chunk(self) -> None:
        usage = normalize_provider_usage(
            {
                "stream": True,
                "chunks": [
                    {"choices": [{"delta": {"content": "first"}}]},
                    {"usage": {"prompt_tokens": 15, "completion_tokens": 4, "total_tokens": 19}},
                ],
            },
            provider_id="huggingface",
        )

        self.assertEqual(usage["availability"], "reported")
        self.assertEqual(usage["source"], "huggingface_usage")
        self.assertEqual(usage["total_tokens"], 19)

    def test_invalid_or_partial_usage_remains_not_reported(self) -> None:
        cases = (
            {"usage": {"prompt_tokens": 2, "completion_tokens": 1}},
            {"usage": {"prompt_tokens": True, "completion_tokens": 1, "total_tokens": 2}},
            {"usage": {"prompt_tokens": -1, "completion_tokens": 1, "total_tokens": 0}},
            {"usage": {"prompt_tokens": 2, "completion_tokens": 1, "total_tokens": 4}},
        )

        for raw_response in cases:
            with self.subTest(raw_response=raw_response):
                usage = normalize_provider_usage(
                    raw_response,
                    provider_id="openai-compatible",
                )
                self.assertEqual(
                    usage,
                    {
                        "availability": "not_reported",
                        "source": "",
                        "input_tokens": 0,
                        "output_tokens": 0,
                        "total_tokens": 0,
                    },
                )

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

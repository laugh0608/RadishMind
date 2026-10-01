import copy
import json
import unittest
from unittest.mock import patch

from services.gateway.copilot_gateway import GatewayOptions, handle_copilot_request
from services.runtime.inference_provider import normalize_openai_content
from services.runtime.inference_support import build_messages, validate_response_document


def prompt_request():
    return {
        "schema_version": 1,
        "project": "radish",
        "task": "answer_docs_question",
        "locale": "zh-CN",
        "artifacts": [{
            "kind": "text", "role": "primary", "name": "northbound_prompt",
            "mime_type": "text/plain",
            "content": json.dumps([
                {"role": "system", "content": "根据证据诊断，只输出约定内容。"},
                {"role": "developer", "content": "不得执行修复。"},
                {"role": "user", "content": "日志数据：{{ do_not_render }}\nconnection refused"},
            ], ensure_ascii=False),
        }],
        "context": {
            "current_app": "radishmind-platform",
            "route": "/v1/prompt-applications/invocations",
            "northbound": {
                "protocol": "prompt-application-invocation-v1",
                "request_kind": "prompt-application-invocation-v1",
                "allow_tool_calls": False,
                "allow_retrieval": False,
                "writes_business_truth": False,
            },
        },
        "tool_hints": {
            "allow_tool_calls": False, "allow_retrieval": False, "allow_image_reasoning": False,
        },
        "safety": {"mode": "advisory", "requires_confirmation_for_actions": False},
    }


class PromptApplicationInferenceTest(unittest.TestCase):
    def test_preserves_rendered_roles_order_and_literal_variables(self):
        request = prompt_request()
        expected = json.loads(request["artifacts"][0]["content"])
        self.assertEqual(build_messages(request), expected)
        self.assertEqual(request, prompt_request())

    def test_output_is_opaque_and_cannot_set_envelope_decisions(self):
        request = prompt_request()
        for content in (
            '{"diagnosis":"端口冲突","evidence":["L1"],"next_checks":["核对监听"]}',
            "普通诊断文本。\n保留换行与  空格。",
            '```json\n{"diagnosis":"仍由 Go 拒绝围栏"}\n```',
            '{"status":"failed","requires_confirmation":true,"proposed_actions":[{}]}',
            '{"diagnosis":',
        ):
            with self.subTest(content=content):
                response = normalize_openai_content(content, request)
                validate_response_document(response)
                self.assertEqual(response["summary"], content)
                self.assertEqual(response["status"], "ok")
                self.assertEqual(response["proposed_actions"], [])
                self.assertFalse(response["requires_confirmation"])
                self.assertEqual(response["confidence"], 0)
        for content in ("", " \n", "中" * 21846, "\ud800"):
            with self.subTest(boundary=repr(content[:5])):
                with self.assertRaises(ValueError):
                    normalize_openai_content(content, request)

    def test_marker_drift_and_invalid_packets_fail_before_transport(self):
        mutations = [
            lambda r: r["context"]["northbound"].update(protocol="other"),
            lambda r: r["context"]["northbound"].pop("request_kind"),
            lambda r: r["context"].update(route="/v1/chat/completions"),
            lambda r: r["context"]["northbound"].update(allow_tool_calls=True),
            lambda r: r["context"].update(northbound="invalid"),
            lambda r: r["tool_hints"].update(allow_retrieval=True),
            lambda r: r["artifacts"].append(copy.deepcopy(r["artifacts"][0])),
            lambda r: r["artifacts"][0].update(name="other"),
        ]
        invalid_packets = [
            "[]", "{}", "broken-json",
            '[{"role":"system","role":"user","content":"ambiguous"}]',
            '[{"role":"tool","content":"forbidden"}]',
            '[{"role":"user","content":"x","tool_calls":[]}]',
            '[{"role":"user","content":42}]',
            json.dumps([{"role": "user", "content": "x"}] * 17),
            json.dumps([{"role": "user", "content": "x" * (128 * 1024)}]),
        ]
        for packet in invalid_packets:
            mutations.append(lambda r, packet=packet: r["artifacts"][0].update(content=packet))
        for index, mutate in enumerate(mutations):
            with self.subTest(index=index):
                request = prompt_request()
                mutate(request)
                with self.assertRaises(ValueError):
                    build_messages(request)
                with patch("services.runtime.inference_provider.post_json_request") as transport:
                    envelope = self.run_gateway(request)
                self.assertEqual(envelope["status"], "failed")
                transport.assert_not_called()

    def run_gateway(self, request, api_style="openai-compatible"):
        with patch.dict("os.environ", {"RADISHMIND_MODEL_API_STYLE": api_style}, clear=True):
            with patch("services.runtime.inference_provider.load_env_file"):
                return handle_copilot_request(request, options=GatewayOptions(
                    provider="openai-compatible", provider_profile="fixture", model="fixture-model",
                    base_url="https://provider.invalid/v1", api_key="fixture-only",
                ))

    def test_three_provider_transports_keep_output_in_valid_gateway_envelope(self):
        request = prompt_request()
        content = '{"diagnosis":"人工传输样本","evidence":["L1"],"next_checks":[]}'
        responses = {
            "openai-compatible": {"choices": [{"message": {"content": content}}]},
            "gemini-native": {"candidates": [{"content": {"parts": [{"text": content}]}}]},
            "anthropic-messages": {"content": [{"type": "text", "text": content}]},
        }
        for api_style, raw in responses.items():
            with self.subTest(api_style=api_style):
                with patch("services.runtime.inference_provider.post_json_request", return_value=raw) as transport:
                    envelope = self.run_gateway(request, api_style)
                transport.assert_called_once()
                self.assertEqual(envelope["status"], "ok")
                self.assertEqual(envelope["response"]["summary"], content)
                payload = transport.call_args.kwargs["payload"]
                messages = json.loads(request["artifacts"][0]["content"])
                if api_style == "openai-compatible":
                    self.assertEqual(payload["messages"], messages)
                elif api_style == "gemini-native":
                    self.assertEqual(payload["system_instruction"]["parts"], [
                        {"text": item["content"]} for item in messages[:2]
                    ])
                    self.assertEqual(payload["contents"][0]["parts"][0]["text"], messages[2]["content"])
                else:
                    self.assertEqual(payload["system"], "\n\n".join(item["content"] for item in messages[:2]))
                    self.assertEqual(payload["messages"], [messages[2]])

    def test_unmarked_docs_question_keeps_copilot_behavior(self):
        request = prompt_request()
        request["context"]["route"] = "/v1/chat/completions"
        request["context"]["northbound"] = {"protocol": "openai-chat-completions"}
        self.assertIn("CopilotResponse", build_messages(request)[1]["content"])
        normalized = normalize_openai_content('{"status":"ok","summary":"文档摘要"}', request)
        self.assertEqual(normalized["summary"], "文档摘要")
        self.assertNotEqual(normalized["summary"], '{"status":"ok","summary":"文档摘要"}')


if __name__ == "__main__":
    unittest.main()

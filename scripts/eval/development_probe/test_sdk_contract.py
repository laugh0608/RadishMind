"""Fixed-response SDK protocol checks; no Agent, file editing or command execution.

Run only in the authorized isolated candidate image. The endpoint is loopback-only.
"""
from __future__ import annotations

import json
import signal
import sys
import time
import unittest
from pathlib import Path
from typing import Literal

if __name__ == "__main__":
    signal.alarm(60)

from jsonschema import Draft202012Validator
from pydantic import ConfigDict
from openhands.sdk import LLM
from openhands.sdk.llm import Message, TextContent
from openhands.sdk.tool import Action, Observation, ToolDefinition

from fixed_endpoint import completion, fixed_endpoint


class ProbeAction(Action):
    model_config = ConfigDict(extra="forbid", strict=True)
    schema_version: Literal["development_task_probe_tool_request.v1"]
    probe_id: str
    sequence: int
    action: Literal["read_file", "replace_file", "run_tests"]
    path: str
    expected_digest: str
    content: str
    candidate_digest: str


class ProbeObservation(Observation):
    pass


class DevelopmentProbeTool(ToolDefinition[ProbeAction, ProbeObservation]):
    @classmethod
    def create(cls, *_args, **_kwargs):
        # No executor is attached: protocol acceptance cannot execute any action.
        return [cls(description="Validate a bounded synthetic task request; no execution.",
                    action_type=ProbeAction, observation_type=ProbeObservation)]


TOOL = DevelopmentProbeTool.create()[0]
SCHEMA = Draft202012Validator(json.loads(Path("/opt/probe/tool-request.schema.json").read_text()))
OBSERVED: dict = {}
REQUEST = {
    "schema_version": "development_task_probe_tool_request.v1", "probe_id": "probe_list_filter",
    "sequence": 1, "action": "read_file", "path": "list.go", "expected_digest": "",
    "content": "", "candidate_digest": "",
}


def llm(base_url: str, *, timeout: int = 2) -> LLM:
    return LLM(model="openai/gpt-4o", api_key="probe-fixture-only", base_url=base_url,
               api_mode="chat", num_retries=0, timeout=timeout, max_output_tokens=128,
               native_tool_calling=True, force_string_serializer=True, disable_vision=True,
               caching_prompt=False, prompt_cache_retention=None, reasoning_effort=None,
               drop_params=False, fallback_strategy=None, log_completions=False)


def tool_reply(arguments: str, *, name: str = TOOL.name) -> dict:
    return completion({"role": "assistant", "content": None, "tool_calls": [{
        "id": "call-probe-001", "type": "function", "function": {"name": name, "arguments": arguments},
    }]}, "tool_calls")


def unique_fields(pairs):
    result = {}
    for key, value in pairs:
        if key in result:
            raise ValueError("duplicate tool argument field")
        result[key] = value
    return result


def accepted_call(message: Message) -> dict:
    """Probe consumer validation only; no ordering, authority or execution grant."""
    if not message.tool_calls or len(message.tool_calls) != 1:
        raise ValueError("expected one native tool call")
    call = message.tool_calls[0]
    if call.name != TOOL.name or call.id != "call-probe-001" or call.origin != "completion":
        raise ValueError("unexpected tool identity")
    if len(call.arguments.encode("utf-8")) > 262144:
        raise ValueError("tool argument body too large")
    request = json.loads(call.arguments, object_pairs_hook=unique_fields)
    if not SCHEMA.is_valid(request):
        raise ValueError("tool request violates canonical schema")
    if request["probe_id"] != REQUEST["probe_id"] or request["sequence"] > 3:
        raise ValueError("tool request does not match the probe plan")
    if len(request["content"].encode("utf-8")) > 32768:
        raise ValueError("tool replacement exceeds UTF-8 byte limit")
    return request


class SDKContractTests(unittest.TestCase):
    def test_native_tool_and_observation_roundtrip(self):
        replies = [(200, tool_reply(json.dumps(REQUEST)), 0),
                   (200, completion({"role": "assistant", "content": "fixed response complete"}), 0)]
        with fixed_endpoint(replies) as (base, captured):
            client = llm(base)
            self.assertEqual(client.num_retries, 0)
            self.assertFalse(client.stream)
            self.assertIsNone(TOOL.executor)
            user = Message(role="user", content=[TextContent(text="Inspect the fixed synthetic task.")])
            first = client.completion(messages=[user], tools=[TOOL])
            self.assertEqual(accepted_call(first.message), REQUEST)
            self.assertEqual(first.raw_response.choices[0].finish_reason, "tool_calls")
            self.assertEqual(first.raw_response.usage.total_tokens, 18)
            observation = ProbeObservation(content=[TextContent(text='{"status":"fixture_observation_only"}')])
            observed = Message(role="tool", name=TOOL.name, tool_call_id=first.message.tool_calls[0].id,
                               content=observation.to_llm_content)
            second = client.completion(messages=[user, first.message, observed], tools=[TOOL])
            self.assertFalse(second.message.tool_calls)
            self.assertEqual(second.raw_response.choices[0].finish_reason, "stop")
            self.assertEqual(len(captured), 2)
            tools = captured[0]["body"]["tools"]
            self.assertEqual([tool["function"]["name"] for tool in tools], [TOOL.name])
            messages = captured[1]["body"]["messages"]
            self.assertEqual(messages[-2]["tool_calls"][0]["id"], "call-probe-001")
            self.assertEqual(messages[-1]["tool_call_id"], "call-probe-001")
            self.assertEqual(messages[-1]["name"], TOOL.name)
            self.assertEqual(messages[-1]["content"], '{"status":"fixture_observation_only"}')
            OBSERVED["first_request"] = captured[0]["body"]
            OBSERVED["observation_request"] = captured[1]["body"]
            OBSERVED["roundtrip_requests"] = len(captured)

    def test_plain_text_success_does_not_pass_tool_probe(self):
        with fixed_endpoint([(200, completion({"role": "assistant", "content": "plain text only"}), 0)]) as (base, captured):
            response = llm(base).completion(messages=[Message(role="user", content="probe")], tools=[TOOL])
            with self.assertRaises(ValueError):
                accepted_call(response.message)
            self.assertEqual(len(captured), 1)

    def test_consumer_rejects_invalid_tool_arguments(self):
        invalid = ["{", json.dumps(REQUEST)[:-1] + ',"path":"list.go"}',
                   json.dumps(dict(REQUEST, path="../list.go")),
                   json.dumps(dict(REQUEST, action="execute_shell")),
                   json.dumps(dict(REQUEST, probe_id="probe_other")),
                   json.dumps(dict(REQUEST, sequence=4)),
                   json.dumps(dict(REQUEST, content=None)),
                   json.dumps(dict(REQUEST, summary="SDK transport metadata")),
                   json.dumps(dict(REQUEST, command="go test"))]
        for arguments in invalid:
            with self.subTest(arguments=arguments), fixed_endpoint([(200, tool_reply(arguments), 0)]) as (base, captured):
                response = llm(base).completion(messages=[Message(role="user", content="probe")], tools=[TOOL])
                with self.assertRaises(ValueError):
                    accepted_call(response.message)
                self.assertEqual(len(captured), 1)

    def test_unknown_tool_is_not_authorized_by_sdk_response(self):
        with fixed_endpoint([(200, tool_reply(json.dumps(REQUEST), name="terminal"), 0)]) as (base, captured):
            response = llm(base).completion(messages=[Message(role="user", content="probe")], tools=[TOOL])
            with self.assertRaises(ValueError):
                accepted_call(response.message)
            self.assertEqual(len(captured), 1)

    def test_http_failures_have_one_attempt(self):
        counts = {}
        for status in (400, 429, 500):
            with self.subTest(status=status), fixed_endpoint([(status, {"error": {"message": "fixed failure"}}, 0)]) as (base, captured):
                with self.assertRaises(Exception) as raised:
                    llm(base).completion(messages=[Message(role="user", content="probe")], tools=[TOOL])
                error = raised.exception
                chain = []
                while error is not None and len(chain) < 5:
                    chain.append({"type": type(error).__name__, "status_code": getattr(error, "status_code", None)})
                    error = error.__cause__ if error.__cause__ is not None else error.__context__
                OBSERVED.setdefault("http_failure_errors", {})[str(status)] = chain
                self.assertTrue(any(item["status_code"] == status for item in chain), chain)
                self.assertEqual(len(captured), 1)
                counts[str(status)] = len(captured)
        OBSERVED["http_failure_attempts"] = counts

    def test_timeout_has_one_attempt(self):
        with fixed_endpoint([(200, completion({"role": "assistant", "content": "late"}), 1.6)]) as (base, captured):
            started = time.monotonic()
            with self.assertRaises(Exception) as raised:
                llm(base, timeout=1).completion(messages=[Message(role="user", content="probe")], tools=[TOOL])
            self.assertIn("Timeout", type(raised.exception).__name__)
            self.assertLess(time.monotonic() - started, 3)
            self.assertEqual(len(captured), 1)
            OBSERVED["timeout_attempts"] = len(captured)

    def test_missing_usage_is_not_a_platform_usage_ledger(self):
        reply = completion({"role": "assistant", "content": "no upstream usage"}, usage=False)
        with fixed_endpoint([(200, reply, 0)]) as (base, captured):
            response = llm(base).completion(messages=[Message(role="user", content="probe")], tools=[TOOL])
            usage = response.raw_response.usage
            OBSERVED["missing_upstream_usage_sdk_value"] = usage.model_dump() if usage is not None else None
            self.assertEqual(len(captured), 1)
            self.assertNotIn("usage", reply)


if __name__ == "__main__":
    suite = unittest.defaultTestLoader.loadTestsFromTestCase(SDKContractTests)
    result = unittest.TextTestRunner(stream=sys.stderr, verbosity=2).run(suite)
    print(json.dumps({"status": "passed" if result.wasSuccessful() else "failed", "tests_run": result.testsRun,
                      "scope": "SDK_LLM_fixed_response_contract_only_not_Gateway_Agent_or_sandbox_acceptance",
                      "observed": OBSERVED}, ensure_ascii=False, indent=2))
    sys.exit(0 if result.wasSuccessful() else 1)

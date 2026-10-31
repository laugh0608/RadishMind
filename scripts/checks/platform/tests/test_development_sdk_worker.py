"""Dependency-free checks for the optional worker's private wire boundary."""
import io
import json
import unittest
from unittest.mock import patch

from services.runtime.development_sdk.worker import (
    ACTION_FIELDS, Bridge, NoRedirect, WorkerFailure, decode, read_input,
)


class Input:
    def __init__(self, value):
        self.buffer = io.BytesIO(json.dumps(value).encode() + b"\n")


def config():
    return {
        "schema_version": "development_worker_input.v1", "attempt_id": "attempt_fixture",
        "executor_instance_id": "executor_fixture", "channel_url": "http://127.0.0.1:12345",
        "credential": "devtask_" + "a" * 43, "model": "fixture-model",
        "initial_messages": [{"role": "system", "content": "synthetic instructions"},
                             {"role": "user", "content": "synthetic task"}],
        "tools": [{"type": "function", "function": {"name": "development_tool"}}],
        "max_output_tokens": 50, "max_model_calls": 4,
    }


class DevelopmentSDKWorkerWireTests(unittest.TestCase):
    def test_private_input_and_exact_fields(self):
        value = config()
        with patch("sys.stdin", Input(value)):
            self.assertEqual(read_input(), value)
        for key in value:
            changed = dict(value)
            del changed[key]
            with self.subTest(key=key), patch("sys.stdin", Input(changed)), self.assertRaises(WorkerFailure):
                read_input()
        for extra in ("provider_url", "shell", "raw_events"):
            with patch("sys.stdin", Input(dict(value, **{extra: "forbidden"}))), self.assertRaises(WorkerFailure):
                read_input()

    def test_input_cannot_expand_transport_or_budget(self):
        for field, value in [("channel_url", "https://example.com"), ("channel_url", "http://user:secret@127.0.0.1:12345"),
                             ("channel_url", "http://127.0.0.1:12345?key=x"), ("max_model_calls", True),
                             ("max_model_calls", 9), ("max_output_tokens", 0), ("credential", "long-lived-api-key")]:
            with self.subTest(field=field, value=value), patch("sys.stdin", Input(dict(config(), **{field: value}))), self.assertRaises(WorkerFailure):
                read_input()
        with self.assertRaises(WorkerFailure):
            NoRedirect().redirect_request(None)

    def test_duplicate_null_nonfinite_and_oversized_wire(self):
        for raw in [b'{"x":1,"x":2}', b'{"x":null}', b'{"x":NaN}', b'"' + b'a' * 262144 + b'"']:
            with self.subTest(raw=raw[:25]), self.assertRaises(WorkerFailure):
                decode(raw)

    def test_sdk_report_is_metadata_only(self):
        sink = io.StringIO()
        bridge = Bridge(config())
        acks = [dict(schema_version="development_worker_reply.v1",sequence=i,request_id=rid,observation="") for i,rid in [(1,"worker_ready"),(2,"sdk_event")]]
        replies = Input(acks[0]);replies.buffer=io.BytesIO(b"".join(json.dumps(a).encode()+b"\n" for a in acks))
        with patch("services.runtime.development_sdk.worker.PROTOCOL_OUT", sink), patch("sys.stdin",replies):
            bridge.emit("ready")
            bridge.emit("event", event_type="user_message")
        frames = [json.loads(line) for line in sink.getvalue().splitlines()]
        self.assertEqual([frame["sequence"] for frame in frames], [1, 2])
        self.assertNotIn(config()["credential"], sink.getvalue())
        self.assertEqual(set(frames[1]), {"schema_version", "sequence", "kind", "event_type", "request_id"})
        self.assertEqual(len(ACTION_FIELDS), 6)

    def test_tool_wait_requires_matching_trusted_reply(self):
        bridge = Bridge(config())
        action = dict(schema_version="development_tool_action.v1", action="read_file", path="main.go",
                      expected_digest="", content="", candidate_digest="")
        bridge.pending = dict(schema_version="development_tool_call.v1", attempt_id="attempt_fixture",
                              request_id="tool_fixture", parent_request_id="model_1", tool_call_id="call_fixture",
                              arguments=action)
        reply = dict(schema_version="development_worker_reply.v1", sequence=1,
                     request_id="tool_wrong", observation="synthetic")
        with patch.object(bridge, "request", return_value={"status": "queued"}), patch("sys.stdin", Input(reply)), \
             patch("services.runtime.development_sdk.worker.PROTOCOL_OUT", io.StringIO()), self.assertRaises(WorkerFailure):
            bridge.observation(action)
        self.assertIsNotNone(bridge.pending)


if __name__ == "__main__":
    unittest.main()

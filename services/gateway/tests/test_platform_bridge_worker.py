from __future__ import annotations

import importlib.util
import io
import json
import unittest
from pathlib import Path
from unittest import mock


REPO_ROOT = Path(__file__).resolve().parents[3]
BRIDGE_SCRIPT = REPO_ROOT / "scripts/run-platform-bridge.py"
MODULE_SPEC = importlib.util.spec_from_file_location("radishmind_platform_bridge", BRIDGE_SCRIPT)
if MODULE_SPEC is None or MODULE_SPEC.loader is None:
    raise RuntimeError("platform bridge module could not be loaded")
PLATFORM_BRIDGE = importlib.util.module_from_spec(MODULE_SPEC)
MODULE_SPEC.loader.exec_module(PLATFORM_BRIDGE)


def worker_request(
    request_id: str,
    operation: str,
    *,
    request: dict[str, object] | None = None,
    options: dict[str, object] | None = None,
) -> dict[str, object]:
    frame: dict[str, object] = {
        "protocol_version": 1,
        "type": "request",
        "request_id": request_id,
        "operation": operation,
    }
    if request is not None:
        frame["request"] = request
    if options is not None:
        frame["options"] = options
    return frame


class PlatformBridgeWorkerTest(unittest.TestCase):
    def run_worker(self, frames: list[object]) -> tuple[int, str, list[dict[str, object]]]:
        worker_input = io.StringIO("".join(json.dumps(frame) + "\n" for frame in frames))
        worker_output = io.StringIO()
        exit_code = PLATFORM_BRIDGE.run_worker(worker_input, worker_output)
        output_text = worker_output.getvalue()
        output_frames = [json.loads(line) for line in output_text.splitlines()]
        return exit_code, output_text, output_frames

    def test_worker_handshake_and_credentials_are_request_scoped(self) -> None:
        observed_credentials: list[str] = []

        def fake_handle(request_document, *, options, stream_handler=None):
            self.assertIsNone(stream_handler)
            observed_credentials.append(options.api_key)
            return {"status": "ok", "request_id": request_document["request_id"]}

        frames = [
            worker_request(
                "bridge-001",
                "envelope",
                request={"request_id": "canonical-001"},
                options={"provider": "mock", "api_key": "request-secret-token"},
            ),
            worker_request(
                "bridge-002",
                "envelope",
                request={"request_id": "canonical-002"},
                options={"provider": "mock"},
            ),
        ]

        with mock.patch.object(PLATFORM_BRIDGE, "handle_copilot_request", side_effect=fake_handle):
            exit_code, output_text, output_frames = self.run_worker(frames)

        self.assertEqual(exit_code, 0)
        self.assertEqual(observed_credentials, ["request-secret-token", ""])
        self.assertNotIn("request-secret-token", output_text)
        self.assertEqual(output_frames[0]["type"], "ready")
        self.assertEqual(output_frames[0]["operations"], ["envelope", "stream", "providers", "inventory"])
        self.assertEqual([frame["request_id"] for frame in output_frames[1:]], ["bridge-001", "bridge-002"])
        self.assertTrue(all(frame["type"] == "result" for frame in output_frames[1:]))

    def test_stream_events_are_correlated_and_completed(self) -> None:
        def fake_handle(request_document, *, options, stream_handler=None):
            self.assertEqual(options.provider, "mock")
            self.assertIsNotNone(stream_handler)
            stream_handler({"type": "delta", "delta": "first"})
            stream_handler({"type": "delta", "delta": request_document["request_id"]})
            return {"status": "ok", "request_id": request_document["request_id"]}

        frame = worker_request(
            "bridge-stream-001",
            "stream",
            request={"request_id": "canonical-stream-001"},
            options={"provider": "mock"},
        )
        with mock.patch.object(PLATFORM_BRIDGE, "handle_copilot_request", side_effect=fake_handle):
            exit_code, _, output_frames = self.run_worker([frame])

        self.assertEqual(exit_code, 0)
        response_frames = output_frames[1:]
        self.assertEqual([item["type"] for item in response_frames], [
            "stream_event",
            "stream_event",
            "stream_event",
            "result",
        ])
        self.assertTrue(all(item["request_id"] == "bridge-stream-001" for item in response_frames))
        self.assertEqual(response_frames[-2]["event"]["type"], "completed")

    def test_worker_serves_provider_registry_and_inventory_without_credentials(self) -> None:
        exit_code, _, output_frames = self.run_worker(
            [
                worker_request("bridge-providers-001", "providers"),
                worker_request("bridge-inventory-001", "inventory"),
            ]
        )

        self.assertEqual(exit_code, 0)
        self.assertEqual([frame["type"] for frame in output_frames], ["ready", "result", "result"])
        self.assertIsInstance(output_frames[1]["payload"], list)
        self.assertIsInstance(output_frames[2]["payload"], dict)
        self.assertIn("providers", output_frames[2]["payload"])

    def test_worker_rejects_metadata_request_options(self) -> None:
        exit_code, _, output_frames = self.run_worker(
            [worker_request("bridge-providers-options", "providers", options={"provider": "mock"})]
        )

        self.assertEqual(exit_code, 0)
        self.assertEqual(output_frames[-1]["type"], "error")
        self.assertEqual(output_frames[-1]["request_id"], "bridge-providers-options")
        self.assertEqual(output_frames[-1]["error"]["code"], "BRIDGE_WORKER_PROTOCOL_ERROR")

    def test_worker_frame_limit_counts_utf8_bytes(self) -> None:
        worker_input = io.StringIO("界界界\n")
        worker_output = io.StringIO()

        with mock.patch.object(PLATFORM_BRIDGE, "WORKER_MAX_FRAME_BYTES", 8):
            exit_code = PLATFORM_BRIDGE.run_worker(worker_input, worker_output)
        output_frames = [json.loads(line) for line in worker_output.getvalue().splitlines()]

        self.assertEqual(exit_code, 1)
        self.assertEqual(output_frames[-1]["type"], "error")
        self.assertEqual(output_frames[-1]["error"]["code"], "BRIDGE_WORKER_PROTOCOL_ERROR")
        self.assertEqual(output_frames[-1]["error"]["message"], "bridge worker request frame exceeds limit")

    def test_worker_protocol_errors_are_stable_and_sanitized(self) -> None:
        invalid_json_with_secret = '{"api_key":"must-not-leak"'
        worker_input = io.StringIO(
            invalid_json_with_secret
            + "\n"
            + json.dumps(worker_request("bridge-unknown", "unknown"))
            + "\n"
        )
        worker_output = io.StringIO()

        exit_code = PLATFORM_BRIDGE.run_worker(worker_input, worker_output)
        output_text = worker_output.getvalue()
        output_frames = [json.loads(line) for line in output_text.splitlines()]

        self.assertEqual(exit_code, 0)
        self.assertNotIn("must-not-leak", output_text)
        self.assertEqual([frame["type"] for frame in output_frames], ["ready", "error", "error"])
        self.assertEqual(output_frames[1]["error"]["code"], "BRIDGE_WORKER_PROTOCOL_ERROR")
        self.assertEqual(output_frames[1]["error"]["message"], "bridge worker request frame is invalid")
        self.assertEqual(output_frames[2]["error"]["code"], "BRIDGE_WORKER_PROTOCOL_ERROR")


if __name__ == "__main__":
    unittest.main()

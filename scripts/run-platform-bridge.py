#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import math
import os
import sys
from pathlib import Path
from typing import Any, TextIO

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.gateway import GatewayOptions, handle_copilot_request  # noqa: E402
from services.runtime.inference_support import describe_provider_inventory  # noqa: E402
from services.runtime.provider_registry import describe_provider_registry  # noqa: E402


BRIDGE_API_KEY_ENV = "RADISHMIND_PLATFORM_BRIDGE_API_KEY"
WORKER_PROTOCOL_VERSION = 1
WORKER_MAX_FRAME_BYTES = 8 * 1024 * 1024
WORKER_OPERATIONS = ("envelope", "stream", "providers", "inventory")


class WorkerRequestError(ValueError):
    def __init__(self, code: str, message: str) -> None:
        super().__init__(message)
        self.code = code
        self.message = message


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Bridge Go platform requests to the canonical Python runtime.")
    subparsers = parser.add_subparsers(dest="command", required=True)

    def add_bridge_arguments(target_parser: argparse.ArgumentParser) -> None:
        target_parser.add_argument("--request-file", default="", help="Optional request json file. Reads stdin when omitted.")
        target_parser.add_argument("--provider", default="mock")
        target_parser.add_argument("--provider-profile", default="")
        target_parser.add_argument("--model", default="")
        target_parser.add_argument("--base-url", default="")
        target_parser.add_argument("--temperature", type=float, default=0.0)
        target_parser.add_argument("--request-timeout-seconds", type=float, default=120.0)

    envelope_parser = subparsers.add_parser("envelope", help="Read a CopilotRequest and emit a CopilotGatewayEnvelope.")
    add_bridge_arguments(envelope_parser)
    stream_parser = subparsers.add_parser("stream", help="Read a CopilotRequest and emit JSONL stream events.")
    add_bridge_arguments(stream_parser)
    subparsers.add_parser("providers", help="Emit the canonical provider registry description.")
    subparsers.add_parser("inventory", help="Emit the canonical provider and profile inventory.")
    subparsers.add_parser("worker", help="Run the versioned persistent JSONL bridge worker.")
    return parser.parse_args()


def load_request_document(path_text: str) -> dict[str, Any]:
    if path_text.strip():
        return json.loads(Path(path_text).read_text(encoding="utf-8"))
    return json.load(sys.stdin)


def run_envelope(args: argparse.Namespace) -> int:
    request_document = load_request_document(args.request_file)
    envelope = handle_copilot_request(
        request_document,
        options=GatewayOptions(
            provider=args.provider,
            provider_profile=args.provider_profile,
            model=args.model,
            base_url=args.base_url,
            api_key=os.environ.get(BRIDGE_API_KEY_ENV, ""),
            temperature=args.temperature,
            request_timeout_seconds=args.request_timeout_seconds,
        ),
    )
    json.dump(envelope, sys.stdout, ensure_ascii=False, indent=2)
    sys.stdout.write("\n")
    return 0


def run_stream(args: argparse.Namespace) -> int:
    request_document = load_request_document(args.request_file)

    def emit_event(event: dict[str, Any]) -> None:
        json.dump(event, sys.stdout, ensure_ascii=False)
        sys.stdout.write("\n")
        sys.stdout.flush()

    envelope = handle_copilot_request(
        request_document,
        options=GatewayOptions(
            provider=args.provider,
            provider_profile=args.provider_profile,
            model=args.model,
            base_url=args.base_url,
            api_key=os.environ.get(BRIDGE_API_KEY_ENV, ""),
            temperature=args.temperature,
            request_timeout_seconds=args.request_timeout_seconds,
        ),
        stream_handler=emit_event,
    )
    json.dump({"type": "completed", "envelope": envelope}, sys.stdout, ensure_ascii=False)
    sys.stdout.write("\n")
    return 0


def run_providers() -> int:
    json.dump(describe_provider_registry(), sys.stdout, ensure_ascii=False, indent=2)
    sys.stdout.write("\n")
    return 0


def run_inventory() -> int:
    json.dump(describe_provider_inventory(), sys.stdout, ensure_ascii=False, indent=2)
    sys.stdout.write("\n")
    return 0


def write_worker_frame(output_stream: TextIO, frame: dict[str, Any]) -> None:
    json.dump(frame, output_stream, ensure_ascii=False, separators=(",", ":"))
    output_stream.write("\n")
    output_stream.flush()


def worker_error_frame(request_id: str, code: str, message: str) -> dict[str, Any]:
    return {
        "protocol_version": WORKER_PROTOCOL_VERSION,
        "type": "error",
        "request_id": request_id,
        "error": {
            "code": code,
            "message": message,
        },
    }


def require_worker_string(document: dict[str, Any], key: str, default: str = "") -> str:
    value = document.get(key, default)
    if not isinstance(value, str):
        raise WorkerRequestError("BRIDGE_WORKER_PROTOCOL_ERROR", "bridge worker options are invalid")
    return value.strip()


def worker_gateway_options(document: Any) -> GatewayOptions:
    if document is None:
        document = {}
    if not isinstance(document, dict):
        raise WorkerRequestError("BRIDGE_WORKER_PROTOCOL_ERROR", "bridge worker options are invalid")
    allowed_fields = {
        "provider",
        "provider_profile",
        "model",
        "base_url",
        "api_key",
        "temperature",
        "request_timeout_seconds",
    }
    if set(document) - allowed_fields:
        raise WorkerRequestError("BRIDGE_WORKER_PROTOCOL_ERROR", "bridge worker options are invalid")

    temperature = document.get("temperature", 0.0)
    timeout_seconds = document.get("request_timeout_seconds", 120.0)
    if isinstance(temperature, bool) or not isinstance(temperature, (int, float)):
        raise WorkerRequestError("BRIDGE_WORKER_PROTOCOL_ERROR", "bridge worker options are invalid")
    if isinstance(timeout_seconds, bool) or not isinstance(timeout_seconds, (int, float)):
        raise WorkerRequestError("BRIDGE_WORKER_PROTOCOL_ERROR", "bridge worker options are invalid")
    if not math.isfinite(float(temperature)) or not math.isfinite(float(timeout_seconds)) or timeout_seconds <= 0:
        raise WorkerRequestError("BRIDGE_WORKER_PROTOCOL_ERROR", "bridge worker options are invalid")

    return GatewayOptions(
        provider=require_worker_string(document, "provider", "mock") or "mock",
        provider_profile=require_worker_string(document, "provider_profile"),
        model=require_worker_string(document, "model"),
        base_url=require_worker_string(document, "base_url"),
        api_key=require_worker_string(document, "api_key"),
        temperature=float(temperature),
        request_timeout_seconds=float(timeout_seconds),
    )


def validate_worker_request_frame(frame: Any) -> tuple[str, str, dict[str, Any] | None, GatewayOptions | None]:
    if not isinstance(frame, dict):
        raise WorkerRequestError("BRIDGE_WORKER_PROTOCOL_ERROR", "bridge worker request frame is invalid")
    allowed_fields = {"protocol_version", "type", "request_id", "operation", "request", "options"}
    if set(frame) - allowed_fields:
        raise WorkerRequestError("BRIDGE_WORKER_PROTOCOL_ERROR", "bridge worker request frame is invalid")
    if type(frame.get("protocol_version")) is not int or frame["protocol_version"] != WORKER_PROTOCOL_VERSION:
        raise WorkerRequestError("BRIDGE_WORKER_PROTOCOL_ERROR", "bridge worker protocol version is unsupported")
    if frame.get("type") != "request":
        raise WorkerRequestError("BRIDGE_WORKER_PROTOCOL_ERROR", "bridge worker request frame type is invalid")

    request_id = frame.get("request_id")
    if not isinstance(request_id, str) or not request_id.strip() or len(request_id) > 128:
        raise WorkerRequestError("BRIDGE_WORKER_PROTOCOL_ERROR", "bridge worker request id is invalid")
    request_id = request_id.strip()
    operation = frame.get("operation")
    if operation not in WORKER_OPERATIONS:
        raise WorkerRequestError("BRIDGE_WORKER_PROTOCOL_ERROR", "bridge worker operation is unsupported")

    if operation in ("providers", "inventory"):
        options_document = frame.get("options")
        if "request" in frame or options_document not in (None, {}):
            raise WorkerRequestError("BRIDGE_WORKER_PROTOCOL_ERROR", "bridge worker metadata operation input is invalid")
        return request_id, operation, None, None

    request_document = frame.get("request")
    if not isinstance(request_document, dict):
        raise WorkerRequestError("BRIDGE_WORKER_PROTOCOL_ERROR", "bridge worker canonical request is invalid")
    return request_id, operation, request_document, worker_gateway_options(frame.get("options"))


def handle_worker_request(frame: Any, output_stream: TextIO) -> None:
    request_id, operation, request_document, options = validate_worker_request_frame(frame)

    if operation == "providers":
        payload: Any = describe_provider_registry()
    elif operation == "inventory":
        payload = describe_provider_inventory()
    elif operation == "envelope":
        if request_document is None or options is None:
            raise WorkerRequestError("BRIDGE_WORKER_PROTOCOL_ERROR", "bridge worker request frame is invalid")
        payload = handle_copilot_request(request_document, options=options)
    else:
        if request_document is None or options is None:
            raise WorkerRequestError("BRIDGE_WORKER_PROTOCOL_ERROR", "bridge worker request frame is invalid")

        def emit_event(event: dict[str, Any]) -> None:
            write_worker_frame(
                output_stream,
                {
                    "protocol_version": WORKER_PROTOCOL_VERSION,
                    "type": "stream_event",
                    "request_id": request_id,
                    "event": event,
                },
            )

        envelope = handle_copilot_request(request_document, options=options, stream_handler=emit_event)
        emit_event({"type": "completed", "envelope": envelope})
        payload = None

    write_worker_frame(
        output_stream,
        {
            "protocol_version": WORKER_PROTOCOL_VERSION,
            "type": "result",
            "request_id": request_id,
            "payload": payload,
        },
    )


def process_worker_line(raw_line: str, output_stream: TextIO) -> None:
    request_id = ""
    try:
        frame = json.loads(raw_line)
        if isinstance(frame, dict) and isinstance(frame.get("request_id"), str):
            request_id = frame["request_id"].strip()[:128]
        handle_worker_request(frame, output_stream)
    except json.JSONDecodeError:
        write_worker_frame(
            output_stream,
            worker_error_frame("", "BRIDGE_WORKER_PROTOCOL_ERROR", "bridge worker request frame is invalid"),
        )
    except WorkerRequestError as exc:
        write_worker_frame(output_stream, worker_error_frame(request_id, exc.code, exc.message))
    except Exception:
        write_worker_frame(
            output_stream,
            worker_error_frame(request_id, "BRIDGE_WORKER_REQUEST_FAILED", "bridge worker request failed"),
        )


def run_worker(input_stream: TextIO | None = None, output_stream: TextIO | None = None) -> int:
    worker_input = input_stream or sys.stdin
    worker_output = output_stream or sys.stdout
    write_worker_frame(
        worker_output,
        {
            "protocol_version": WORKER_PROTOCOL_VERSION,
            "type": "ready",
            "operations": list(WORKER_OPERATIONS),
        },
    )

    while True:
        raw_line = worker_input.readline(WORKER_MAX_FRAME_BYTES + 1)
        if raw_line == "":
            return 0
        if len(raw_line.encode("utf-8")) > WORKER_MAX_FRAME_BYTES:
            raw_line = ""
            write_worker_frame(
                worker_output,
                worker_error_frame("", "BRIDGE_WORKER_PROTOCOL_ERROR", "bridge worker request frame exceeds limit"),
            )
            return 1
        if not raw_line.strip():
            raw_line = ""
            continue

        process_worker_line(raw_line, worker_output)
        raw_line = ""


def main() -> int:
    args = parse_args()
    if args.command == "envelope":
        return run_envelope(args)
    if args.command == "stream":
        return run_stream(args)
    if args.command == "providers":
        return run_providers()
    if args.command == "inventory":
        return run_inventory()
    if args.command == "worker":
        return run_worker()
    raise SystemExit(f"unsupported command: {args.command}")


if __name__ == "__main__":
    raise SystemExit(main())

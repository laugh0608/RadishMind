#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from typing import Any
from urllib.parse import urlsplit


MAX_REQUEST_BYTES = 256 * 1024
LOOPBACK_HOSTS = {"127.0.0.1", "::1", "localhost"}


def parse_listen(value: str) -> tuple[str, int]:
    parsed = urlsplit(f"//{value.strip()}")
    host = (parsed.hostname or "").strip().lower()
    try:
        port = parsed.port
    except ValueError as exc:
        raise argparse.ArgumentTypeError("listen port is invalid") from exc
    if host not in LOOPBACK_HOSTS or port is None or port < 1024 or port > 65535:
        raise argparse.ArgumentTypeError("listen must use a loopback host and an unprivileged port")
    return host, port


def success_response() -> dict[str, Any]:
    copilot_response = {
        "schema_version": 1,
        "status": "ok",
        "project": "radish",
        "task": "answer_docs_question",
        "summary": "Deterministic Provider Attempt fixture completed the reviewed fallback path.",
        "answers": [
            {
                "kind": "direct_answer",
                "text": "The deterministic backup target completed without contacting a real Provider.",
                "citation_ids": [],
            }
        ],
        "issues": [],
        "proposed_actions": [],
        "citations": [],
        "confidence": 1.0,
        "risk_level": "low",
        "requires_confirmation": False,
    }
    return {
        "id": "provider-attempt-fixture",
        "object": "chat.completion",
        "choices": [
            {
                "index": 0,
                "message": {
                    "role": "assistant",
                    "content": json.dumps(copilot_response, separators=(",", ":")),
                },
                "finish_reason": "stop",
            }
        ],
        "usage": {"prompt_tokens": 7, "completion_tokens": 5, "total_tokens": 12},
    }


class ProviderAttemptFixtureHandler(BaseHTTPRequestHandler):
    server_version = "RadishMindProviderAttemptFixture/1"
    sys_version = ""

    def do_GET(self) -> None:  # noqa: N802
        if self.path != "/healthz":
            self.write_json(404, {"error": {"code": "fixture_route_not_found"}})
            return
        self.write_json(200, {"status": "ok", "fixture": "provider_attempt.v1"})

    def do_POST(self) -> None:  # noqa: N802
        if self.path not in {"/eligible/v1/chat/completions", "/success/v1/chat/completions"}:
            self.write_json(404, {"error": {"code": "fixture_route_not_found"}})
            return
        request_document = self.read_request_document()
        if request_document is None:
            return
        if self.path == "/eligible/v1/chat/completions":
            self.write_json(
                503,
                {
                    "error": {
                        "code": "fixture_temporarily_unavailable",
                        "message": "deterministic Provider Attempt fixture failure",
                    }
                },
            )
            return
        self.write_json(200, success_response())

    def read_request_document(self) -> dict[str, Any] | None:
        try:
            content_length = int(self.headers.get("Content-Length", ""))
        except ValueError:
            content_length = -1
        if content_length < 1 or content_length > MAX_REQUEST_BYTES:
            self.write_json(400, {"error": {"code": "fixture_request_invalid"}})
            return None
        try:
            document = json.loads(self.rfile.read(content_length))
        except (UnicodeDecodeError, json.JSONDecodeError):
            self.write_json(400, {"error": {"code": "fixture_request_invalid"}})
            return None
        if (
            not isinstance(document, dict)
            or not isinstance(document.get("model"), str)
            or not isinstance(document.get("messages"), list)
        ):
            self.write_json(400, {"error": {"code": "fixture_request_invalid"}})
            return None
        return document

    def write_json(self, status: int, document: dict[str, Any]) -> None:
        payload = json.dumps(document, separators=(",", ":")).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(payload)))
        self.send_header("Cache-Control", "no-store")
        self.end_headers()
        self.wfile.write(payload)

    def log_message(self, _format: str, *_args: object) -> None:
        return


def main() -> int:
    parser = argparse.ArgumentParser(description="Run the loopback-only Provider Attempt product fixture.")
    parser.add_argument("--listen", default="127.0.0.1:7201", type=parse_listen)
    args = parser.parse_args()
    server = ThreadingHTTPServer(args.listen, ProviderAttemptFixtureHandler)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        pass
    finally:
        server.server_close()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

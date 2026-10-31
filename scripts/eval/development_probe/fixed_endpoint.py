"""Loopback-only, bounded Chat Completions fixture for SDK contract tests.

This is not the platform Gateway or a model provider. No tool is executed here.
"""
from __future__ import annotations

import json
import threading
import time
from contextlib import contextmanager
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer


def completion(message: dict, finish_reason: str = "stop", *, usage: bool = True) -> dict:
    response = {
        "id": "chatcmpl-fixed-probe", "object": "chat.completion", "created": 1,
        "model": "gpt-4o",  # Local SDK capability metadata only; no real model is called.
        "choices": [{"index": 0, "message": message, "finish_reason": finish_reason}],
    }
    if usage:
        response["usage"] = {"prompt_tokens": 11, "completion_tokens": 7, "total_tokens": 18}
    return response


@contextmanager
def fixed_endpoint(replies: list[tuple[int, dict, float]]):
    """Yield (base_url, captured_requests); each reply is (status, JSON, delay)."""
    captured: list[dict] = []
    lock = threading.Lock()

    class Handler(BaseHTTPRequestHandler):
        def setup(self):
            super().setup()
            self.connection.settimeout(3)

        def do_POST(self):
            size = int(self.headers.get("Content-Length", "0"))
            if self.path != "/v1/chat/completions" or size < 1 or size > 262144:
                self.send_error(400)
                return
            body = json.loads(self.rfile.read(size))
            with lock:
                index = len(captured)
                captured.append({"path": self.path, "body": body})
            if index >= len(replies):
                status, payload, delay = 500, {"error": {"message": "unexpected extra request"}}, 0
            else:
                status, payload, delay = replies[index]
            time.sleep(delay)
            data = json.dumps(payload, ensure_ascii=False).encode("utf-8")
            try:
                self.send_response(status)
                self.send_header("Content-Type", "application/json")
                self.send_header("Content-Length", str(len(data)))
                self.end_headers()
                self.wfile.write(data)
            except (BrokenPipeError, ConnectionResetError):
                # Expected when the timeout test disconnects before its delayed reply.
                if delay == 0:
                    raise

        def log_message(self, *_args):
            pass  # The fixture never logs request headers or bodies.

    server = ThreadingHTTPServer(("127.0.0.1", 0), Handler)
    server.daemon_threads = False
    thread = threading.Thread(target=server.serve_forever, kwargs={"poll_interval": 0.01})
    thread.start()
    try:
        yield f"http://127.0.0.1:{server.server_port}/v1", captured
    finally:
        server.shutdown()
        server.server_close()
        thread.join(timeout=2)
        if thread.is_alive():
            raise RuntimeError("fixed response server did not stop")

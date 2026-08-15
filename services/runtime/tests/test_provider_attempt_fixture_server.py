from __future__ import annotations

import http.client
import importlib.util
import json
import threading
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts" / "eval" / "provider_attempt_fixture_server.py"
MODULE_SPEC = importlib.util.spec_from_file_location("provider_attempt_fixture_server", FIXTURE_PATH)
assert MODULE_SPEC is not None and MODULE_SPEC.loader is not None
FIXTURE_MODULE = importlib.util.module_from_spec(MODULE_SPEC)
MODULE_SPEC.loader.exec_module(FIXTURE_MODULE)


class ProviderAttemptFixtureServerTest(unittest.TestCase):
    def setUp(self) -> None:
        self.server = FIXTURE_MODULE.ThreadingHTTPServer(
            ("127.0.0.1", 0), FIXTURE_MODULE.ProviderAttemptFixtureHandler
        )
        self.thread = threading.Thread(target=self.server.serve_forever, daemon=True)
        self.thread.start()

    def tearDown(self) -> None:
        self.server.shutdown()
        self.server.server_close()
        self.thread.join(timeout=5)

    def request(self, method: str, path: str, document: dict[str, object] | None = None):
        connection = http.client.HTTPConnection("127.0.0.1", self.server.server_port, timeout=5)
        body = None if document is None else json.dumps(document)
        headers = {"Authorization": "Bearer fixture-private-value"}
        if body is not None:
            headers["Content-Type"] = "application/json"
        connection.request(method, path, body=body, headers=headers)
        response = connection.getresponse()
        payload = response.read().decode("utf-8")
        connection.close()
        return response.status, json.loads(payload), payload

    def test_health_eligible_failure_and_success_are_deterministic_and_sanitized(self) -> None:
        status, health, _ = self.request("GET", "/healthz")
        self.assertEqual((status, health), (200, {"status": "ok", "fixture": "provider_attempt.v1"}))

        request = {"model": "fixture-model", "messages": [{"role": "user", "content": "private"}]}
        status, failure, failure_payload = self.request("POST", "/eligible/v1/chat/completions", request)
        self.assertEqual(status, 503)
        self.assertEqual(failure["error"]["code"], "fixture_temporarily_unavailable")
        self.assertNotIn("private", failure_payload)
        self.assertNotIn("fixture-private-value", failure_payload)

        status, success, success_payload = self.request("POST", "/success/v1/chat/completions", request)
        self.assertEqual(status, 200)
        self.assertEqual(success["usage"], {"prompt_tokens": 7, "completion_tokens": 5, "total_tokens": 12})
        response_document = json.loads(success["choices"][0]["message"]["content"])
        self.assertEqual(response_document["status"], "ok")
        self.assertNotIn("private", success_payload)
        self.assertNotIn("fixture-private-value", success_payload)

    def test_invalid_routes_and_payloads_fail_closed(self) -> None:
        status, document, _ = self.request("POST", "/success/v1/chat/completions", {"model": "fixture-model"})
        self.assertEqual((status, document["error"]["code"]), (400, "fixture_request_invalid"))
        status, document, _ = self.request("POST", "/unknown", {"model": "fixture-model", "messages": []})
        self.assertEqual((status, document["error"]["code"]), (404, "fixture_route_not_found"))


if __name__ == "__main__":
    unittest.main()

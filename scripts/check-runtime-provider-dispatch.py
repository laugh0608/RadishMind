#!/usr/bin/env python3
from __future__ import annotations

import json
import os
import sys
from pathlib import Path
from unittest.mock import patch
from typing import Any

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.inference_provider import run_inference  # noqa: E402
from services.runtime.inference_support import describe_provider_inventory  # noqa: E402
from services.runtime.provider_registry import list_provider_ids  # noqa: E402


SAMPLE_REQUEST_PATH = REPO_ROOT / "datasets/eval/radish/answer-docs-question-direct-answer-001.json"


class FakeHTTPResponse:
    def __init__(self, payload: dict[str, Any]) -> None:
        self._payload = payload

    def __enter__(self) -> "FakeHTTPResponse":
        return self

    def __exit__(self, exc_type, exc, tb) -> None:
        return None

    def read(self) -> bytes:
        return json.dumps(self._payload, ensure_ascii=False).encode("utf-8")


def load_sample_request() -> tuple[dict[str, Any], dict[str, Any]]:
    sample = json.loads(SAMPLE_REQUEST_PATH.read_text(encoding="utf-8"))
    request_document = sample.get("input_request")
    golden_response = sample.get("golden_response")
    if not isinstance(request_document, dict) or not isinstance(golden_response, dict):
        raise SystemExit(f"invalid sample fixture: {SAMPLE_REQUEST_PATH.relative_to(REPO_ROOT)}")
    return request_document, golden_response


def build_fake_response_payload(response_document: dict[str, Any]) -> dict[str, Any]:
    return {
        "choices": [
            {
                "message": {
                    "content": json.dumps(response_document, ensure_ascii=False),
                }
            }
        ]
    }


def run_provider_case(
    *,
    provider: str,
    env: dict[str, str],
    expected_model: str,
    expected_transport: str,
    expected_api_key: str | None,
) -> None:
    request_document, golden_response = load_sample_request()
    fake_response = FakeHTTPResponse(build_fake_response_payload(golden_response))
    captured_requests: list[Any] = []

    def fake_urlopen(request_obj: Any, timeout: float | None = None) -> FakeHTTPResponse:
        captured_requests.append(request_obj)
        return fake_response

    with patch.dict(os.environ, env, clear=False):
        with patch("services.runtime.inference_provider.request.urlopen", side_effect=fake_urlopen):
            result = run_inference(
                request_document,
                provider=provider,
                provider_profile=None,
                model=None,
                base_url=None,
                api_key=None,
                temperature=0.0,
                request_timeout_seconds=1.0,
            )

    if result.get("provider") != provider:
        raise SystemExit(f"{provider}: unexpected provider {result.get('provider')}")
    if result.get("model") != expected_model:
        raise SystemExit(f"{provider}: unexpected model {result.get('model')}")
    raw_request = result.get("raw_request") or {}
    if raw_request.get("transport") != expected_transport:
        raise SystemExit(f"{provider}: unexpected transport {raw_request.get('transport')}")
    if raw_request.get("api_style") != expected_transport:
        raise SystemExit(f"{provider}: unexpected api style {raw_request.get('api_style')}")
    if result.get("response", {}).get("summary") != golden_response.get("summary"):
        raise SystemExit(f"{provider}: unexpected response summary")

    if not captured_requests:
        raise SystemExit(f"{provider}: provider request was not captured")
    authorization = captured_requests[0].get_header("Authorization")
    if expected_api_key is None:
        if authorization:
            raise SystemExit(f"{provider}: expected no authorization header")
    else:
        if authorization != f"Bearer {expected_api_key}":
            raise SystemExit(f"{provider}: unexpected authorization header {authorization}")
    content_type = captured_requests[0].get_header("Content-Type") or captured_requests[0].get_header("Content-type")
    if content_type != "application/json":
        raise SystemExit(f"{provider}: missing content type header")


def assert_inventory_profiles(*, env: dict[str, str], expected_profiles: set[tuple[str, str]]) -> None:
    with patch.dict(os.environ, env, clear=False):
        inventory = describe_provider_inventory()
    observed_profiles = {
        (str(item.get("provider_id") or "").strip(), str(item.get("profile") or "").strip())
        for item in inventory.get("profiles") or []
    }
    missing_profiles = sorted(expected_profiles - observed_profiles)
    if missing_profiles:
        raise SystemExit(f"inventory missing expected provider profiles: {missing_profiles}")


def main() -> int:
    supported_provider_ids = set(list_provider_ids())
    required_provider_ids = {"mock", "openai-compatible", "huggingface", "ollama"}
    if not required_provider_ids.issubset(supported_provider_ids):
        raise SystemExit(
            f"provider registry missing expected ids: {sorted(required_provider_ids - supported_provider_ids)}"
        )

    run_provider_case(
        provider="huggingface",
        env={
            "RADISHMIND_HUGGINGFACE_PROFILE": "hf",
            "RADISHMIND_HUGGINGFACE_PROFILE_HF_NAME": "meta-llama/Meta-Llama-3.1-8B-Instruct",
            "RADISHMIND_HUGGINGFACE_PROFILE_HF_BASE_URL": "https://example.huggingface.invalid/v1",
            "RADISHMIND_HUGGINGFACE_PROFILE_HF_API_KEY": "hf-token",
            "RADISHMIND_HUGGINGFACE_PROFILE_HF_API_STYLE": "huggingface-chat-completions",
        },
        expected_model="meta-llama/Meta-Llama-3.1-8B-Instruct",
        expected_transport="huggingface-chat-completions",
        expected_api_key="hf-token",
    )

    run_provider_case(
        provider="ollama",
        env={
            "RADISHMIND_OLLAMA_NAME": "qwen2.5:7b-instruct",
            "RADISHMIND_OLLAMA_BASE_URL": "http://localhost:11434",
            "RADISHMIND_OLLAMA_API_STYLE": "ollama-chat-completions",
        },
        expected_model="qwen2.5:7b-instruct",
        expected_transport="ollama-chat-completions",
        expected_api_key=None,
    )

    assert_inventory_profiles(
        env={
            "RADISHMIND_HUGGINGFACE_PROFILE": "hf",
            "RADISHMIND_HUGGINGFACE_PROFILE_HF_NAME": "meta-llama/Meta-Llama-3.1-8B-Instruct",
            "RADISHMIND_HUGGINGFACE_PROFILE_HF_BASE_URL": "https://example.huggingface.invalid/v1",
            "RADISHMIND_HUGGINGFACE_PROFILE_HF_API_KEY": "hf-token",
            "RADISHMIND_HUGGINGFACE_PROFILE_HF_API_STYLE": "huggingface-chat-completions",
            "RADISHMIND_OLLAMA_NAME": "qwen2.5:7b-instruct",
            "RADISHMIND_OLLAMA_BASE_URL": "http://localhost:11434",
            "RADISHMIND_OLLAMA_API_STYLE": "ollama-chat-completions",
        },
        expected_profiles={
            ("huggingface", "hf"),
            ("ollama", "default"),
        },
    )

    print("runtime provider dispatch checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

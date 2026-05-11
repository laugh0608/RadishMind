from __future__ import annotations

import time
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Callable

import jsonschema

from services.runtime.inference import run_inference, validate_request_document, validate_response_document
from services.runtime.inference_support import load_schema, make_failed_response, utc_now_iso
from services.runtime.provider_registry import is_supported_provider


REPO_ROOT = Path(__file__).resolve().parent.parent.parent
GATEWAY_ENVELOPE_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-gateway-envelope.schema.json"
SUPPORTED_ROUTES = {
    ("radish", "answer_docs_question"),
    ("radishflow", "suggest_flowsheet_edits"),
    ("radishflow", "suggest_ghost_completion"),
}


@dataclass(frozen=True)
class GatewayOptions:
    provider: str = "mock"
    provider_profile: str = ""
    model: str = ""
    base_url: str = ""
    api_key: str = ""
    temperature: float = 0.0
    request_timeout_seconds: float = 120.0


def route_key(copilot_request: dict[str, Any]) -> tuple[str, str]:
    return (
        str(copilot_request.get("project") or "").strip(),
        str(copilot_request.get("task") or "").strip(),
    )


def build_gateway_metadata(
    *,
    started_at: float,
    copilot_request: dict[str, Any] | None,
    options: GatewayOptions,
    request_validated: bool,
    response_validated: bool,
) -> dict[str, Any]:
    project, task = route_key(copilot_request or {})
    return {
        "handled_at": utc_now_iso(),
        "duration_ms": max(0, round((time.perf_counter() - started_at) * 1000)),
        "route": f"{project}/{task}" if project and task else "",
        "request_validated": request_validated,
        "response_validated": response_validated,
        "advisory_only": True,
        "provider": {
            "name": options.provider,
            "profile": options.provider_profile,
            "model": options.model,
            "request_timeout_seconds": options.request_timeout_seconds,
        },
    }


def build_gateway_envelope(
    *,
    started_at: float,
    copilot_request: dict[str, Any] | None,
    options: GatewayOptions,
    status: str,
    response: dict[str, Any] | None,
    error: dict[str, str] | None = None,
    request_validated: bool = False,
    response_validated: bool = False,
) -> dict[str, Any]:
    request = copilot_request or {}
    project, task = route_key(request)
    return {
        "schema_version": 1,
        "status": status,
        "request_id": str(request.get("request_id") or "").strip(),
        "project": project,
        "task": task,
        "response": response,
        "error": error,
        "metadata": build_gateway_metadata(
            started_at=started_at,
            copilot_request=copilot_request,
            options=options,
            request_validated=request_validated,
            response_validated=response_validated,
        ),
    }


def validate_gateway_envelope(document: Any) -> None:
    jsonschema.validate(document, load_schema(GATEWAY_ENVELOPE_SCHEMA_PATH))


def validated_gateway_envelope(**kwargs: Any) -> dict[str, Any]:
    envelope = build_gateway_envelope(**kwargs)
    validate_gateway_envelope(envelope)
    return envelope


def failed_copilot_response(copilot_request: dict[str, Any], *, code: str, message: str) -> dict[str, Any]:
    response = make_failed_response(copilot_request, message, message, code)
    validate_response_document(response)
    return response


def handle_copilot_request(
    copilot_request: dict[str, Any],
    *,
    options: GatewayOptions | None = None,
    stream_handler: Callable[[dict[str, Any]], None] | None = None,
) -> dict[str, Any]:
    gateway_options = options or GatewayOptions()
    started_at = time.perf_counter()

    try:
        validate_request_document(copilot_request)
    except jsonschema.ValidationError as exc:
        return validated_gateway_envelope(
            started_at=started_at,
            copilot_request=copilot_request if isinstance(copilot_request, dict) else None,
            options=gateway_options,
            status="failed",
            response=None,
            error={
                "code": "REQUEST_SCHEMA_INVALID",
                "message": exc.message,
            },
        )

    project, task = route_key(copilot_request)
    if not is_supported_provider(gateway_options.provider):
        message = f"unsupported provider: {gateway_options.provider or '<empty>'}"
        response = failed_copilot_response(copilot_request, code="UNSUPPORTED_PROVIDER", message=message)
        return validated_gateway_envelope(
            started_at=started_at,
            copilot_request=copilot_request,
            options=gateway_options,
            status="failed",
            response=response,
            error={
                "code": "UNSUPPORTED_PROVIDER",
                "message": message,
            },
            request_validated=True,
            response_validated=True,
        )

    if (project, task) not in SUPPORTED_ROUTES:
        message = f"unsupported copilot route: {project} / {task}"
        response = failed_copilot_response(copilot_request, code="UNSUPPORTED_TASK", message=message)
        return validated_gateway_envelope(
            started_at=started_at,
            copilot_request=copilot_request,
            options=gateway_options,
            status="failed",
            response=response,
            error={
                "code": "UNSUPPORTED_TASK",
                "message": message,
            },
            request_validated=True,
            response_validated=True,
        )

    try:
        inference_result = run_inference(
            copilot_request,
            provider=gateway_options.provider,
            provider_profile=gateway_options.provider_profile or None,
            model=gateway_options.model or None,
            base_url=gateway_options.base_url or None,
            api_key=gateway_options.api_key or None,
            temperature=gateway_options.temperature,
            request_timeout_seconds=gateway_options.request_timeout_seconds,
            stream_handler=stream_handler,
        )
        response = inference_result["response"]
        validate_response_document(response)
        return validated_gateway_envelope(
            started_at=started_at,
            copilot_request=copilot_request,
            options=gateway_options,
            status=str(response.get("status") or "ok"),
            response=response,
            request_validated=True,
            response_validated=True,
        )
    except Exception as exc:
        message = f"gateway inference failed: {exc}"
        response = failed_copilot_response(copilot_request, code="GATEWAY_INFERENCE_FAILED", message=message)
        return validated_gateway_envelope(
            started_at=started_at,
            copilot_request=copilot_request,
            options=gateway_options,
            status="failed",
            response=response,
            error={
                "code": "GATEWAY_INFERENCE_FAILED",
                "message": message,
            },
            request_validated=True,
            response_validated=True,
        )

from __future__ import annotations

import http.client
from dataclasses import asdict, dataclass
from urllib import error


PROVIDER_ATTEMPT_FAILURE_SCHEMA_VERSION = "gateway_provider_attempt_failure.v1"


@dataclass(frozen=True)
class ProviderAttemptFailureObservation:
    schema_version: str
    failure_class: str
    fallback_disposition: str
    provider_response_started: bool
    outcome: str
    code: str
    http_status_class: str = ""

    def as_document(self) -> dict[str, object]:
        return asdict(self)


class ProviderAttemptError(RuntimeError):
    def __init__(self, message: str, observation: ProviderAttemptFailureObservation) -> None:
        super().__init__(message)
        self.observation = observation


def provider_http_failure_observation(status_code: int) -> ProviderAttemptFailureObservation:
    status_class = f"{status_code // 100}xx" if 400 <= status_code <= 599 else ""
    failure_class = "provider_invalid_request"
    disposition = "ineligible"
    code = "PROVIDER_INVALID_REQUEST"
    if status_code == 429:
        failure_class = "provider_rate_limited"
        disposition = "eligible"
        code = "PROVIDER_RATE_LIMITED"
    elif status_code in (502, 503, 504):
        failure_class = "provider_upstream_gateway_unavailable"
        disposition = "eligible"
        code = "PROVIDER_UPSTREAM_GATEWAY_UNAVAILABLE"
    elif status_code in (401, 403):
        failure_class = "provider_authentication_failed"
        code = "PROVIDER_AUTHENTICATION_FAILED"
    elif status_code == 404:
        failure_class = "provider_model_not_found"
        code = "PROVIDER_MODEL_NOT_FOUND"
    elif 500 <= status_code <= 599:
        failure_class = "provider_temporarily_unavailable"
        code = "PROVIDER_TEMPORARILY_UNAVAILABLE"
    return ProviderAttemptFailureObservation(
        schema_version=PROVIDER_ATTEMPT_FAILURE_SCHEMA_VERSION,
        failure_class=failure_class,
        fallback_disposition=disposition,
        provider_response_started=False,
        outcome="failed",
        code=code,
        http_status_class=status_class,
    )


def provider_unknown_outcome_observation() -> ProviderAttemptFailureObservation:
    return ProviderAttemptFailureObservation(
        schema_version=PROVIDER_ATTEMPT_FAILURE_SCHEMA_VERSION,
        failure_class="provider_temporarily_unavailable",
        fallback_disposition="ineligible",
        provider_response_started=False,
        outcome="unknown",
        code="PROVIDER_OUTCOME_UNKNOWN",
    )


def normalized_provider_attempt_error(exc: BaseException) -> ProviderAttemptError:
    if isinstance(exc, error.HTTPError):
        return ProviderAttemptError(
            f"provider request failed with HTTP {exc.code}",
            provider_http_failure_observation(exc.code),
        )
    reason = exc.reason if isinstance(exc, error.URLError) else None
    observation = provider_unknown_outcome_observation()
    if isinstance(exc, TimeoutError) or isinstance(reason, TimeoutError):
        return ProviderAttemptError("provider request timed out", observation)
    if isinstance(exc, (http.client.RemoteDisconnected, http.client.IncompleteRead)):
        return ProviderAttemptError("provider connection terminated", observation)
    return ProviderAttemptError("provider request could not reach upstream", observation)

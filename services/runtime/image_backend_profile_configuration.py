from __future__ import annotations

import copy
import hashlib
import json
import math
import re
from collections.abc import Mapping
from dataclasses import dataclass
from functools import lru_cache
from pathlib import Path
from typing import Any

import jsonschema


REPO_ROOT = Path(__file__).resolve().parents[2]
PROFILE_SOURCE_SCHEMA_PATH = REPO_ROOT / "contracts/image-backend-profile-source.schema.json"

FAILURE_PROFILE_SOURCE_INVALID = "image_backend_profile_source_invalid"
FAILURE_PROFILE_SOURCE_BUDGET = "image_backend_profile_source_budget_exceeded"
FAILURE_PROFILE_SENSITIVE_MATERIAL = "image_backend_profile_sensitive_material_rejected"
FAILURE_PROFILE_ENVIRONMENT_FORBIDDEN = "image_backend_profile_environment_forbidden"
FAILURE_PROFILE_BINDING_INVALID = "image_backend_profile_binding_invalid"
FAILURE_PROFILE_DIGEST_DRIFT = "image_backend_profile_digest_drift"

MAX_PROFILE_SOURCE_BYTES = 16 * 1024
MAX_TIMEOUT_SECONDS = 300.0

IDENTIFIER_RE = re.compile(r"^[a-z][a-z0-9._-]{2,63}$")
PROFILE_VALUE_RE = re.compile(r"^[A-Za-z0-9][A-Za-z0-9._:/-]{0,127}$")
DIGEST_RE = re.compile(r"^sha256:[a-f0-9]{64}$")
REFERENCE_RE = re.compile(
    r"^ref:radishmind/(development|test)/image-backends/[a-z0-9._/-]+$"
)
FORBIDDEN_KEYS = {
    "api_key",
    "authorization",
    "authorization_header",
    "base_url",
    "cookie",
    "credential_handle",
    "credential_value",
    "dsn",
    "endpoint",
    "endpoint_url",
    "headers",
    "model_dir",
    "model_path",
    "password",
    "private_key",
    "provider_config",
    "provider_runtime",
    "raw_credential",
    "raw_secret",
    "runtime_config",
    "secret",
    "secret_value",
    "system_prompt",
    "token",
}
SENSITIVE_VALUE_PATTERNS = (
    re.compile(r"https?://", re.IGNORECASE),
    re.compile(r"\b(?:postgres(?:ql)?|mysql|mongodb(?:\+srv)?|redis)://", re.IGNORECASE),
    re.compile(
        r"\b(?:authorization|api[_-]?key|credentials?|token|cookie|password|"
        r"secret|endpoint|dsn|model[_-]?(?:dir|path)|"
        r"provider[_-]?(?:config|runtime)|runtime[_-]?config|"
        r"system[_-]?prompt)\s*[:=]",
        re.IGNORECASE,
    ),
    re.compile(r"\bbearer\s+[A-Za-z0-9._~+/=-]{8,}", re.IGNORECASE),
    re.compile(r"\bsk-[A-Za-z0-9_-]{12,}", re.IGNORECASE),
    re.compile(r"-----BEGIN [A-Z ]*PRIVATE KEY-----"),
    re.compile(r"^(?:/|[A-Za-z]:[\\/]|\\\\)"),
    re.compile(r"\$\{[^}]+\}|%[A-Za-z_][A-Za-z0-9_]*%"),
)


@dataclass(frozen=True)
class ImageBackendProfile:
    profile_id: str
    profile_version: int
    environment: str
    enabled: bool
    backend_id: str
    model: str
    adapter_profile: str
    runtime_mode: str
    endpoint_ref: str | None
    model_dir_ref: str | None
    credential_requirement: str
    secret_ref: str | None
    timeout_seconds: float
    profile_digest: str


@dataclass(frozen=True)
class ImageBackendProfileCompilationResult:
    ok: bool
    profile: ImageBackendProfile | None = None
    failure_code: str | None = None
    failure_message: str = ""


def compile_image_backend_profile(
    source_document: Mapping[str, Any],
) -> ImageBackendProfileCompilationResult:
    source = copy_mapping(source_document)
    if source is None:
        return compilation_failure(
            FAILURE_PROFILE_SOURCE_INVALID,
            "image backend profile source must be an object",
        )
    if not contains_only_valid_utf8(source):
        return compilation_failure(
            FAILURE_PROFILE_SOURCE_INVALID,
            "image backend profile source must contain valid UTF-8 text",
        )
    source_bytes = canonical_json_bytes(source)
    if source_bytes is None:
        return compilation_failure(
            FAILURE_PROFILE_SOURCE_INVALID,
            "image backend profile source must be JSON serializable",
        )
    if len(source_bytes) > MAX_PROFILE_SOURCE_BYTES:
        return compilation_failure(
            FAILURE_PROFILE_SOURCE_BUDGET,
            "image backend profile source exceeds the UTF-8 byte budget",
        )
    if contains_sensitive_material(source):
        return compilation_failure(
            FAILURE_PROFILE_SENSITIVE_MATERIAL,
            "image backend profile source contains sensitive or concrete runtime material",
        )
    if source.get("environment") not in {"development", "test"}:
        return compilation_failure(
            FAILURE_PROFILE_ENVIRONMENT_FORBIDDEN,
            "image backend profile source is limited to development and test",
        )

    schema_error = next(iter(profile_source_validator().iter_errors(source)), None)
    if schema_error is not None:
        return compilation_failure(
            FAILURE_PROFILE_SOURCE_INVALID,
            "image backend profile source failed the strict schema",
        )
    binding_failure = validate_source_binding(source)
    if binding_failure:
        return compilation_failure(FAILURE_PROFILE_BINDING_INVALID, binding_failure)

    backend = mapping_at(source, "backend")
    runtime = mapping_at(source, "runtime")
    credential = mapping_at(source, "credential")
    limits = mapping_at(source, "limits")
    canonical_source = canonical_profile_source(source)
    profile = ImageBackendProfile(
        profile_id=source["profile_id"],
        profile_version=source["profile_version"],
        environment=source["environment"],
        enabled=source["enabled"],
        backend_id=backend["id"],
        model=backend["model"],
        adapter_profile=backend["adapter_profile"],
        runtime_mode=runtime["mode"],
        endpoint_ref=runtime["endpoint_ref"],
        model_dir_ref=runtime["model_dir_ref"],
        credential_requirement=credential["requirement"],
        secret_ref=credential["secret_ref"],
        timeout_seconds=float(limits["timeout_seconds"]),
        profile_digest=digest_document(canonical_source),
    )
    if not compiled_profile_is_valid(profile):
        return compilation_failure(
            FAILURE_PROFILE_DIGEST_DRIFT,
            "compiled image backend profile did not preserve its canonical digest",
        )
    return ImageBackendProfileCompilationResult(ok=True, profile=profile)


def compiled_profile_is_valid(profile: Any) -> bool:
    if not isinstance(profile, ImageBackendProfile):
        return False
    if (
        not isinstance(profile.profile_version, int)
        or isinstance(profile.profile_version, bool)
        or profile.profile_version <= 0
        or profile.profile_version > 2_147_483_647
        or not isinstance(profile.enabled, bool)
        or profile.environment not in {"development", "test"}
        or not valid_identifier(profile.profile_id)
        or not all(
            valid_profile_value(value)
            for value in (profile.backend_id, profile.model, profile.adapter_profile)
        )
        or not valid_timeout(profile.timeout_seconds)
        or not DIGEST_RE.fullmatch(profile.profile_digest)
    ):
        return False
    if not profile_binding_is_valid(
        environment=profile.environment,
        runtime_mode=profile.runtime_mode,
        endpoint_ref=profile.endpoint_ref,
        model_dir_ref=profile.model_dir_ref,
        credential_requirement=profile.credential_requirement,
        secret_ref=profile.secret_ref,
    ):
        return False
    canonical = profile_to_canonical_source(profile)
    return (
        not contains_sensitive_material(canonical)
        and profile.profile_digest == digest_document(canonical)
    )


def profile_to_reference_document(profile: ImageBackendProfile) -> dict[str, Any]:
    document = profile_to_canonical_source(profile)
    document["profile_digest"] = profile.profile_digest
    return document


def canonical_profile_source(source: Mapping[str, Any]) -> dict[str, Any]:
    backend = mapping_at(source, "backend")
    runtime = mapping_at(source, "runtime")
    credential = mapping_at(source, "credential")
    limits = mapping_at(source, "limits")
    return {
        "schema_version": 1,
        "kind": "image_backend_profile_source",
        "profile_id": source["profile_id"],
        "profile_version": source["profile_version"],
        "environment": source["environment"],
        "enabled": source["enabled"],
        "backend": {
            "id": backend["id"],
            "model": backend["model"],
            "adapter_profile": backend["adapter_profile"],
        },
        "runtime": {
            "mode": runtime["mode"],
            "endpoint_ref": runtime["endpoint_ref"],
            "model_dir_ref": runtime["model_dir_ref"],
        },
        "credential": {
            "requirement": credential["requirement"],
            "secret_ref": credential["secret_ref"],
        },
        "limits": {
            "timeout_seconds": normalize_number(limits["timeout_seconds"]),
        },
    }


def profile_to_canonical_source(profile: ImageBackendProfile) -> dict[str, Any]:
    return {
        "schema_version": 1,
        "kind": "image_backend_profile_source",
        "profile_id": profile.profile_id,
        "profile_version": profile.profile_version,
        "environment": profile.environment,
        "enabled": profile.enabled,
        "backend": {
            "id": profile.backend_id,
            "model": profile.model,
            "adapter_profile": profile.adapter_profile,
        },
        "runtime": {
            "mode": profile.runtime_mode,
            "endpoint_ref": profile.endpoint_ref,
            "model_dir_ref": profile.model_dir_ref,
        },
        "credential": {
            "requirement": profile.credential_requirement,
            "secret_ref": profile.secret_ref,
        },
        "limits": {
            "timeout_seconds": normalize_number(profile.timeout_seconds),
        },
    }


def validate_source_binding(source: Mapping[str, Any]) -> str:
    backend = mapping_at(source, "backend")
    runtime = mapping_at(source, "runtime")
    credential = mapping_at(source, "credential")
    limits = mapping_at(source, "limits")
    if any(".." in value or "://" in value for value in backend.values()):
        return "image backend identity fields are not canonical"
    if not valid_timeout(limits["timeout_seconds"]):
        return "image backend timeout exceeds the development budget"
    if not profile_binding_is_valid(
        environment=source["environment"],
        runtime_mode=runtime["mode"],
        endpoint_ref=runtime["endpoint_ref"],
        model_dir_ref=runtime["model_dir_ref"],
        credential_requirement=credential["requirement"],
        secret_ref=credential["secret_ref"],
    ):
        return "image backend runtime and credential references are inconsistent"
    return ""


def profile_binding_is_valid(
    *,
    environment: str,
    runtime_mode: str,
    endpoint_ref: Any,
    model_dir_ref: Any,
    credential_requirement: str,
    secret_ref: Any,
) -> bool:
    if runtime_mode == "remote_https":
        return (
            valid_reference(endpoint_ref, environment, "endpoints")
            and model_dir_ref is None
            and credential_requirement == "required"
            and valid_reference(secret_ref, environment, "credentials")
        )
    if runtime_mode == "local_model":
        return (
            endpoint_ref is None
            and valid_reference(model_dir_ref, environment, "model-dirs")
            and credential_requirement == "not_required"
            and secret_ref is None
        )
    return False


def valid_reference(value: Any, environment: str, reference_kind: str) -> bool:
    if not isinstance(value, str) or not REFERENCE_RE.fullmatch(value):
        return False
    expected_prefix = f"ref:radishmind/{environment}/image-backends/{reference_kind}/"
    reference_key = value.removeprefix(expected_prefix)
    return (
        value.startswith(expected_prefix)
        and bool(re.fullmatch(r"[a-z0-9][a-z0-9._/-]{0,127}", reference_key))
        and not reference_key.endswith("/")
        and len(value) <= 256
        and ".." not in value
        and "//" not in value
        and "\\" not in value
    )


def valid_identifier(value: Any) -> bool:
    return isinstance(value, str) and bool(IDENTIFIER_RE.fullmatch(value))


def valid_profile_value(value: Any) -> bool:
    return (
        isinstance(value, str)
        and bool(PROFILE_VALUE_RE.fullmatch(value))
        and "://" not in value
        and ".." not in value
    )


def valid_timeout(value: Any) -> bool:
    return (
        isinstance(value, (int, float))
        and not isinstance(value, bool)
        and math.isfinite(float(value))
        and 1 <= float(value) <= MAX_TIMEOUT_SECONDS
    )


def contains_sensitive_material(value: Any) -> bool:
    if isinstance(value, Mapping):
        for key, item in value.items():
            if isinstance(key, str) and key.casefold() in FORBIDDEN_KEYS:
                return True
            if contains_sensitive_material(item):
                return True
        return False
    if isinstance(value, (list, tuple)):
        return any(contains_sensitive_material(item) for item in value)
    if isinstance(value, str):
        return any(pattern.search(value) for pattern in SENSITIVE_VALUE_PATTERNS)
    return False


def contains_only_valid_utf8(value: Any) -> bool:
    if isinstance(value, Mapping):
        return all(
            contains_only_valid_utf8(key) and contains_only_valid_utf8(item)
            for key, item in value.items()
        )
    if isinstance(value, (list, tuple)):
        return all(contains_only_valid_utf8(item) for item in value)
    if isinstance(value, str):
        try:
            value.encode("utf-8")
        except UnicodeEncodeError:
            return False
    return True


def copy_mapping(value: Any) -> dict[str, Any] | None:
    if not isinstance(value, Mapping):
        return None
    return copy.deepcopy(dict(value))


def mapping_at(document: Mapping[str, Any], key: str) -> dict[str, Any]:
    value = document[key]
    if not isinstance(value, Mapping):
        raise TypeError(f"{key} must be a mapping")
    return dict(value)


def normalize_number(value: int | float) -> int | float:
    numeric = float(value)
    return int(numeric) if numeric.is_integer() else numeric


def canonical_json_bytes(value: Mapping[str, Any]) -> bytes | None:
    try:
        serialized = json.dumps(
            value,
            ensure_ascii=False,
            sort_keys=True,
            separators=(",", ":"),
            allow_nan=False,
        )
        return serialized.encode("utf-8")
    except (TypeError, ValueError, UnicodeEncodeError):
        return None


def digest_document(value: Mapping[str, Any]) -> str:
    canonical = canonical_json_bytes(value)
    if canonical is None:
        raise ValueError(FAILURE_PROFILE_SOURCE_INVALID)
    return f"sha256:{hashlib.sha256(canonical).hexdigest()}"


@lru_cache(maxsize=1)
def profile_source_validator() -> jsonschema.Draft202012Validator:
    schema = json.loads(PROFILE_SOURCE_SCHEMA_PATH.read_text(encoding="utf-8"))
    jsonschema.Draft202012Validator.check_schema(schema)
    return jsonschema.Draft202012Validator(
        schema,
        format_checker=jsonschema.FormatChecker(),
    )


def compilation_failure(
    code: str,
    message: str,
) -> ImageBackendProfileCompilationResult:
    return ImageBackendProfileCompilationResult(
        ok=False,
        failure_code=code,
        failure_message=message,
    )

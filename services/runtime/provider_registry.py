from __future__ import annotations

from dataclasses import asdict, dataclass
from typing import Any


MOCK_PROVIDER_ID = "mock"
OPENAI_COMPATIBLE_PROVIDER_ID = "openai-compatible"
HUGGINGFACE_PROVIDER_ID = "huggingface"
OLLAMA_PROVIDER_ID = "ollama"
OPENAI_COMPATIBLE_API_STYLES = (
    "openai-compatible",
    "gemini-native",
    "anthropic-messages",
)


@dataclass(frozen=True)
class ProviderCapability:
    transport: str
    local_or_remote: str
    chat: bool
    responses: bool
    messages: bool
    models_list: bool
    streaming: bool
    json_schema_output: bool
    tool_calling: bool
    image_input: bool
    image_output: bool
    auth_mode: str
    timeout_policy: str
    retry_policy: str
    cost_profile: str
    latency_profile: str
    deployment_mode: str


@dataclass(frozen=True)
class ProviderSpec:
    provider_id: str
    display_name: str
    default_api_style: str
    supported_api_styles: tuple[str, ...]
    capabilities: ProviderCapability
    profile_driven: bool = False
    notes: str = ""


def _build_mock_provider_spec() -> ProviderSpec:
    return ProviderSpec(
        provider_id=MOCK_PROVIDER_ID,
        display_name="Mock provider",
        default_api_style=MOCK_PROVIDER_ID,
        supported_api_styles=(MOCK_PROVIDER_ID,),
        capabilities=ProviderCapability(
            transport="deterministic-rule-based",
            local_or_remote="local",
            chat=False,
            responses=False,
            messages=False,
            models_list=False,
            streaming=False,
            json_schema_output=True,
            tool_calling=False,
            image_input=False,
            image_output=False,
            auth_mode="none",
            timeout_policy="none",
            retry_policy="none",
            cost_profile="none",
            latency_profile="deterministic",
            deployment_mode="embedded",
        ),
        notes="Internal deterministic runtime for supported Copilot tasks.",
    )


def _build_openai_compatible_provider_spec() -> ProviderSpec:
    return ProviderSpec(
        provider_id=OPENAI_COMPATIBLE_PROVIDER_ID,
        display_name="OpenAI-compatible provider family",
        default_api_style="openai-compatible",
        supported_api_styles=OPENAI_COMPATIBLE_API_STYLES,
        capabilities=ProviderCapability(
            transport="openai-chat-completions",
            local_or_remote="remote",
            chat=True,
            responses=False,
            messages=False,
            models_list=False,
            streaming=True,
            json_schema_output=False,
            tool_calling=False,
            image_input=False,
            image_output=False,
            auth_mode="profile",
            timeout_policy="per-request",
            retry_policy="caller-managed",
            cost_profile="provider-defined",
            latency_profile="provider-defined",
            deployment_mode="remote_api",
        ),
        profile_driven=True,
        notes="Profile-based family for openai-compatible chat, Gemini native, and Anthropic messages transports.",
    )


def _build_huggingface_provider_spec() -> ProviderSpec:
    return ProviderSpec(
        provider_id=HUGGINGFACE_PROVIDER_ID,
        display_name="Hugging Face chat-completions provider",
        default_api_style="huggingface-chat-completions",
        supported_api_styles=("huggingface-chat-completions",),
        capabilities=ProviderCapability(
            transport="huggingface-chat-completions",
            local_or_remote="remote",
            chat=True,
            responses=False,
            messages=False,
            models_list=False,
            streaming=True,
            json_schema_output=False,
            tool_calling=False,
            image_input=False,
            image_output=False,
            auth_mode="profile",
            timeout_policy="per-request",
            retry_policy="caller-managed",
            cost_profile="provider-defined",
            latency_profile="provider-defined",
            deployment_mode="remote_api",
        ),
        profile_driven=True,
        notes="Provider-specific chat-completions transport for Hugging Face hosted endpoints.",
    )


def _build_ollama_provider_spec() -> ProviderSpec:
    return ProviderSpec(
        provider_id=OLLAMA_PROVIDER_ID,
        display_name="Ollama chat-completions provider",
        default_api_style="ollama-chat-completions",
        supported_api_styles=("ollama-chat-completions",),
        capabilities=ProviderCapability(
            transport="ollama-chat-completions",
            local_or_remote="local",
            chat=True,
            responses=False,
            messages=False,
            models_list=False,
            streaming=True,
            json_schema_output=False,
            tool_calling=False,
            image_input=False,
            image_output=False,
            auth_mode="optional",
            timeout_policy="per-request",
            retry_policy="caller-managed",
            cost_profile="local",
            latency_profile="local_daemon",
            deployment_mode="local_daemon",
        ),
        profile_driven=True,
        notes="Provider-specific chat-completions transport for local Ollama deployments.",
    )


PROVIDER_REGISTRY: dict[str, ProviderSpec] = {
    spec.provider_id: spec
    for spec in (
        _build_mock_provider_spec(),
        _build_openai_compatible_provider_spec(),
        _build_huggingface_provider_spec(),
        _build_ollama_provider_spec(),
    )
}


def build_provider_registry() -> dict[str, ProviderSpec]:
    return dict(PROVIDER_REGISTRY)


def get_provider_spec(provider_id: str) -> ProviderSpec:
    normalized_provider_id = str(provider_id or "").strip()
    try:
        return PROVIDER_REGISTRY[normalized_provider_id]
    except KeyError as exc:
        supported = ", ".join(list_provider_ids())
        raise ValueError(f"unsupported provider: {normalized_provider_id or '<empty>'}; supported providers: {supported}") from exc


def is_supported_provider(provider_id: str) -> bool:
    normalized_provider_id = str(provider_id or "").strip()
    return normalized_provider_id in PROVIDER_REGISTRY


def list_provider_ids() -> list[str]:
    return sorted(PROVIDER_REGISTRY)


def describe_provider_registry() -> list[dict[str, Any]]:
    descriptions: list[dict[str, Any]] = []
    for provider_id in list_provider_ids():
        spec = PROVIDER_REGISTRY[provider_id]
        descriptions.append(
            {
                "provider_id": spec.provider_id,
                "display_name": spec.display_name,
                "default_api_style": spec.default_api_style,
                "supported_api_styles": list(spec.supported_api_styles),
                "profile_driven": spec.profile_driven,
                "notes": spec.notes,
                "capabilities": asdict(spec.capabilities),
            }
        )
    return descriptions


__all__ = [
    "MOCK_PROVIDER_ID",
    "HUGGINGFACE_PROVIDER_ID",
    "OPENAI_COMPATIBLE_API_STYLES",
    "OPENAI_COMPATIBLE_PROVIDER_ID",
    "OLLAMA_PROVIDER_ID",
    "PROVIDER_REGISTRY",
    "ProviderCapability",
    "ProviderSpec",
    "build_provider_registry",
    "describe_provider_registry",
    "get_provider_spec",
    "is_supported_provider",
    "list_provider_ids",
]

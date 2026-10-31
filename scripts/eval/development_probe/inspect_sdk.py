"""Print candidate API/dependency metadata inside the isolated SDK image only."""
from __future__ import annotations

import importlib.metadata as metadata
import inspect
import json
import platform
import signal


def main() -> None:
    signal.alarm(45)
    from openhands.sdk import Agent, Conversation, LLM
    from openhands.sdk.conversation import LocalConversation
    from openhands.sdk.llm import Message
    from openhands.sdk.tool import Action, Observation, ToolDefinition, ToolExecutor

    if metadata.version("openhands-sdk") not in ("1.59.0", "1.59.0+radishmind.1"):
        raise ValueError("unexpected SDK candidate")
    apis = {}
    for cls in (LLM, Agent, Conversation, LocalConversation, Message, Action, Observation, ToolDefinition, ToolExecutor):
        apis[cls.__name__] = {
            "signature": str(inspect.signature(cls)),
            "fields": {key: str(field.annotation) for key, field in getattr(cls, "model_fields", {}).items()},
        }
    methods = {}
    for cls, names in [(LLM, ["completion", "close"]), (LocalConversation, ["run", "pause", "close"]),
                       (ToolDefinition, ["create", "to_openai_tool"]), (Observation, ["to_llm_content"])]:
        for name in names:
            method = getattr(cls, name, None)
            methods[f"{cls.__name__}.{name}"] = (
                {"kind": "property", "signature": str(inspect.signature(method.fget))}
                if isinstance(method, property) else str(inspect.signature(method)) if method else None
            )
    distributions = []
    for distribution in metadata.distributions():
        item = distribution.metadata
        distributions.append({
            "name": item["Name"], "version": distribution.version,
            "license_expression": item.get("License-Expression"),
            "license_field": item.get("License"),
            "license_classifiers": [value for value in item.get_all("Classifier", []) if value.startswith("License ::")],
        })
    print(json.dumps({
        "sdk_version": metadata.version("openhands-sdk"), "python": platform.python_version(),
        "system": platform.system(), "machine": platform.machine(),
        "apis": apis, "methods": methods,
        "distributions": sorted(distributions, key=lambda item: item["name"].lower()),
        "scope": "metadata_only_not_license_approval_or_runtime_adoption",
    }, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()

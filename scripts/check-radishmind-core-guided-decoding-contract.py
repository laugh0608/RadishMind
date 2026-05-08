#!/usr/bin/env python3
from __future__ import annotations

import importlib.util
import json
import sys
import tempfile
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))


DRY_RUN_MANIFEST = REPO_ROOT / "scripts/checks/fixtures/radishmind-core-candidate-dry-run-manifest.json"
RESPONSE_SCHEMA = REPO_ROOT / "contracts/copilot-response.schema.json"


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must be a JSON object")
    return document


def load_module(module_name: str, relative_path: str) -> Any:
    module_path = REPO_ROOT / relative_path
    spec = importlib.util.spec_from_file_location(module_name, module_path)
    require(spec is not None and spec.loader is not None, f"{relative_path} module spec must be loadable")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


class _GuidedGenerationConfig:
    def __init__(self) -> None:
        self.guided_decoding = None


class _SupportedTransformers:
    GenerationConfig = _GuidedGenerationConfig


class _MissingHookGenerationConfig:
    pass


class _UnsupportedTransformers:
    GenerationConfig = _MissingHookGenerationConfig


class _CustomGenerateGenerationMixin:
    def generate(self, *args: Any, custom_generate: Any | None = None, **kwargs: Any) -> Any:
        return None


class _CustomGenerateTransformers:
    GenerationConfig = _MissingHookGenerationConfig
    GenerationMixin = _CustomGenerateGenerationMixin


def main() -> int:
    runner = load_module("radishmind_core_candidate_runner", "scripts/run-radishmind-core-candidate.py")
    from services.runtime import inference_provider as runtime_provider

    manifest = load_json(DRY_RUN_MANIFEST)
    response_schema = load_json(RESPONSE_SCHEMA)

    parser = runner.parse_args
    require(callable(parser), "candidate runner must expose parse_args")
    original_argv = sys.argv[:]
    try:
        sys.argv = ["run-radishmind-core-candidate.py", "--guided-decoding", "json_schema"]
        parsed = parser()
    finally:
        sys.argv = original_argv
    require(parsed.guided_decoding == "json_schema", "parse_args must accept --guided-decoding json_schema")

    guided_request = runtime_provider.build_guided_decoding_request(
        mode="json_schema",
        json_schema=response_schema,
    )
    require(
        guided_request["mode"] == runtime_provider.GUIDED_DECODING_MODE_JSON_SCHEMA,
        "guided decoding request must preserve json_schema mode",
    )
    require(
        runtime_provider.describe_guided_decoding_request(guided_request) == "json_schema",
        "guided decoding request description mismatch",
    )
    payload = runtime_provider.build_local_transformers_guided_decoding_payload(
        guided_decoding_request=guided_request
    )
    require(payload.get("type") == "json_schema", "guided decoding payload type mismatch")
    require(payload.get("json_schema") == response_schema, "guided decoding payload must reuse response schema")

    supported = runtime_provider.resolve_local_transformers_guided_decoding_support(
        transformers_module=_SupportedTransformers(),
        guided_decoding_request=guided_request,
    )
    require(supported.get("supported") is True, "supported transformers stub must pass guided decoding support check")
    require(
        supported.get("hook") == "GenerationConfig.guided_decoding",
        "supported transformers stub must report guided_decoding hook",
    )
    require(
        supported.get("backend") == "generation_config_guided_decoding",
        "supported transformers stub must report native guided decoding backend",
    )

    unsupported = runtime_provider.resolve_local_transformers_guided_decoding_support(
        transformers_module=_UnsupportedTransformers(),
        guided_decoding_request=guided_request,
    )
    require(unsupported.get("supported") is False, "unsupported transformers stub must fail support check")
    require(
        "GenerationConfig.guided_decoding" in str(unsupported.get("reason") or ""),
        "unsupported transformers reason must mention missing GenerationConfig.guided_decoding support",
    )
    require(
        "GenerationMixin.generate" in str(unsupported.get("reason") or ""),
        "unsupported transformers reason must mention missing custom_generate fallback support",
    )

    custom_generate_supported = runtime_provider.resolve_local_transformers_guided_decoding_support(
        transformers_module=_CustomGenerateTransformers(),
        guided_decoding_request=guided_request,
    )
    require(
        custom_generate_supported.get("supported") is True,
        "custom_generate transformers stub must pass guided decoding support check",
    )
    require(
        custom_generate_supported.get("backend") == "custom_generate_callable",
        "custom_generate transformers stub must report custom_generate backend",
    )
    require(
        custom_generate_supported.get("hook") == "GenerationMixin.generate(custom_generate=callable)",
        "custom_generate transformers stub must report custom_generate hook",
    )

    args = type(
        "Args",
        (),
        {
            "repair_hard_fields": False,
            "inject_hard_fields": False,
            "build_suggest_edits_response": False,
            "build_task_scoped_response": False,
            "guided_decoding": "json_schema",
        },
    )()
    require(
        runner.active_structured_variant_name(args) == "raw_guided_json_schema",
        "guided decoding variant name mismatch",
    )

    source_eval_manifest = load_json(REPO_ROOT / str(manifest.get("source_eval_manifest") or ""))
    selected_samples = runner.iter_selected_samples(source_eval_manifest, sample_id_filter=None)
    require(selected_samples, "source eval manifest must expose at least one sample")
    sample, _selected_sample = selected_samples[0]
    guided_plan = runner.build_guided_json_schema_scaffold_plan(sample)
    require(isinstance(guided_plan, list) and guided_plan, "guided plan must be a non-empty list")
    require(
        any(isinstance(segment, dict) and segment.get("kind") == "text_slot" for segment in guided_plan),
        "guided plan must expose at least one text slot",
    )
    require(
        any(isinstance(segment, dict) and segment.get("kind") == "static" for segment in guided_plan),
        "guided plan must expose static scaffold segments",
    )

    guided_policy = runner.build_postprocess_policy(
        args=args,
        guided_decoding_request=guided_request,
        postprocessed_output_count=0,
        postprocessed_path_counts={},
    )
    require(isinstance(guided_policy, dict), "guided decoding policy must be generated")
    require(guided_policy.get("guided_decoding") == "json_schema", "guided decoding policy mode mismatch")
    require(
        guided_policy.get("candidate_track") == "raw_guided_json_schema",
        "guided decoding policy candidate track mismatch",
    )
    require(guided_policy.get("requires_runtime_support") is True, "guided decoding policy must require runtime support")

    with tempfile.TemporaryDirectory(prefix="check-guided-decoding-contract-") as tmp_dir:
        summary = runner.build_candidate_run(
            type(
                "BuildArgs",
                (),
                {
                    "manifest": str(DRY_RUN_MANIFEST.relative_to(REPO_ROOT)),
                    "output_dir": tmp_dir,
                    "summary_output": None,
                    "check_summary": None,
                    "provider": "echo",
                    "model_dir": None,
                    "max_new_tokens": 1200,
                    "sample_timeout_seconds": 0.0,
                    "sample_id": None,
                    "allow_invalid_output": False,
                    "validate_task": False,
                    "repair_hard_fields": False,
                    "inject_hard_fields": False,
                    "build_suggest_edits_response": False,
                    "build_task_scoped_response": False,
                    "guided_decoding": None,
                },
            )(),
            manifest,
        )
    require(summary.get("candidate_track") is None, "plain dry-run summary must not set candidate_track")

    print("radishmind core guided decoding contract check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

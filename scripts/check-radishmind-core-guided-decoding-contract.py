#!/usr/bin/env python3
from __future__ import annotations

import importlib.util
import inspect
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


class _SampleSignatureGenerationMixin:
    def _sample(
        self,
        input_ids: Any,
        logits_processor: Any,
        stopping_criteria: Any,
        generation_config: Any,
        synced_gpus: bool = False,
        streamer: Any | None = None,
        **model_kwargs: Any,
    ) -> Any:
        del (
            input_ids,
            logits_processor,
            stopping_criteria,
            generation_config,
            synced_gpus,
            streamer,
            model_kwargs,
        )
        return None


class _CustomGenerateTransformers:
    GenerationConfig = _MissingHookGenerationConfig
    GenerationMixin = _CustomGenerateGenerationMixin


def main() -> int:
    runner = load_module("radishmind_core_candidate_runner", "scripts/run-radishmind-core-candidate.py")
    from services.runtime import inference_provider as runtime_provider
    from services.runtime.guided_decoding import GuidedJsonSchemaCustomGenerator

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

    class _TokenizerStub:
        all_special_ids = [0]

        def encode(self, text: str, add_special_tokens: bool = False) -> list[int]:
            del add_special_tokens
            if text == '"':
                return [1]
            return [2]

        def decode(
            self,
            token_ids: list[int],
            skip_special_tokens: bool = False,
            clean_up_tokenization_spaces: bool = False,
        ) -> str:
            del skip_special_tokens, clean_up_tokenization_spaces
            return "".join('"' if token_id == 1 else "x" for token_id in token_ids)

    custom_generator = GuidedJsonSchemaCustomGenerator(
        tokenizer=_TokenizerStub(),
        plan=[{"kind": "static", "text": "{}"}],
    )
    usual_mode_kwargs = set(inspect.signature(_SampleSignatureGenerationMixin._sample).parameters.keys())
    custom_generate_kwargs = set(inspect.signature(custom_generator).parameters.keys())
    require(
        custom_generate_kwargs - usual_mode_kwargs == {"model"},
        "guided decoding custom_generate callable must not expose custom-only kwargs beyond model",
    )
    require(
        "attention_mask" not in custom_generate_kwargs,
        "guided decoding custom_generate callable must read attention_mask from **model_kwargs",
    )

    class _FakeIds:
        ndim = 2
        shape = (1, 1)

        def clone(self) -> "_FakeIds":
            return self

        def new_ones(self, _shape: Any) -> "_FakeIds":
            return self

    class _GenerationConfigStub:
        max_new_tokens = 2

    class _ReserveClosingQuoteGenerator(GuidedJsonSchemaCustomGenerator):
        def __init__(self) -> None:
            super().__init__(
                tokenizer=_TokenizerStub(),
                plan=[
                    {
                        "kind": "text_slot",
                        "path": "$.summary",
                        "fallback_text": "x",
                        "min_non_space_chars": 1,
                        "max_chars": 8,
                    }
                ],
            )
            self.appended_token_ids: list[int] = []

        def _prime_model(
            self,
            model: Any,
            *,
            input_ids: Any,
            attention_mask: Any,
        ) -> tuple[Any, Any]:
            del model, input_ids, attention_mask
            return None, object()

        def _append_chunk(
            self,
            model: Any,
            *,
            current_ids: Any,
            attention_mask: Any,
            past_key_values: Any,
            token_ids: list[int],
        ) -> tuple[Any, Any, Any, Any]:
            next_attention_mask = attention_mask
            del model, past_key_values
            self.appended_token_ids.extend(token_ids)
            return current_ids, next_attention_mask, None, object()

        def _pick_slot_token(
            self,
            logits: Any,
            *,
            slot_text: str,
            min_non_space_chars: int,
            max_chars: int,
            require_non_space_progress: bool = False,
        ) -> tuple[str, int | None, str | None] | None:
            del logits, slot_text, min_non_space_chars, max_chars, require_non_space_progress
            return ("token", 2, "x")

    reserve_quote_generator = _ReserveClosingQuoteGenerator()
    reserve_quote_generator(
        model=object(),
        input_ids=_FakeIds(),
        logits_processor=None,
        stopping_criteria=None,
        generation_config=_GenerationConfigStub(),
    )
    require(
        reserve_quote_generator.appended_token_ids == [2, reserve_quote_generator.quote_token_id],
        "guided decoding custom_generate callable must reserve the last token for closing the current JSON string slot",
    )

    class _CharTokenizer:
        all_special_ids = [0]

        def encode(self, text: str, add_special_tokens: bool = False) -> list[int]:
            del add_special_tokens
            if text == '"':
                return [1]
            return [1000 + ord(character) for character in text]

        def decode(
            self,
            token_ids: list[int],
            skip_special_tokens: bool = False,
            clean_up_tokenization_spaces: bool = False,
        ) -> str:
            del skip_special_tokens, clean_up_tokenization_spaces
            pieces: list[str] = []
            for token_id in token_ids:
                if token_id == 1:
                    pieces.append('"')
                elif token_id == 2001:
                    pieces.append("zzzz")
                elif token_id >= 1000:
                    pieces.append(chr(token_id - 1000))
                else:
                    pieces.append("x")
            return "".join(pieces)

    class _BudgetAwareGenerator(GuidedJsonSchemaCustomGenerator):
        def __init__(self) -> None:
            super().__init__(
                tokenizer=_CharTokenizer(),
                plan=[
                    {"kind": "static", "text": '"'},
                    {
                        "kind": "text_slot",
                        "path": "$.summary",
                        "fallback_text": "fallback-summary-text",
                        "min_non_space_chars": 2,
                        "max_chars": 8,
                    },
                    {"kind": "static", "text": ',"'},
                    {
                        "kind": "text_slot",
                        "path": "$.issues[0].message",
                        "fallback_text": "fallback-message-text",
                        "min_non_space_chars": 2,
                        "max_chars": 8,
                    },
                ],
            )
            self.appended_token_ids: list[int] = []

        @staticmethod
        def _compact_fallback_text(*, slot_path: str) -> str:
            if slot_path == "$.summary":
                return "请复核。"
            if slot_path == "$.issues[0].message":
                return "请复核。"
            return ""

        def _prime_model(
            self,
            model: Any,
            *,
            input_ids: Any,
            attention_mask: Any,
        ) -> tuple[Any, Any]:
            del model, input_ids, attention_mask
            return None, object()

        def _append_chunk(
            self,
            model: Any,
            *,
            current_ids: Any,
            attention_mask: Any,
            past_key_values: Any,
            token_ids: list[int],
        ) -> tuple[Any, Any, Any, Any]:
            del model, past_key_values
            self.appended_token_ids.extend(token_ids)
            return current_ids, attention_mask, None, object()

        def _pick_slot_token(
            self,
            logits: Any,
            *,
            slot_text: str,
            min_non_space_chars: int,
            max_chars: int,
            require_non_space_progress: bool = False,
        ) -> tuple[str, int | None, str | None] | None:
            del logits, max_chars, require_non_space_progress
            if len(slot_text) < min_non_space_chars:
                return ("token", 2001, "zzzz")
            return ("end", self.quote_token_id, None)

    budget_aware_generator = _BudgetAwareGenerator()
    minimum_budget = 0
    for segment in budget_aware_generator.plan:
        if segment.get("kind") == "static":
            minimum_budget += len(budget_aware_generator._encode_static_text(str(segment.get("text") or "")))
        else:
            minimum_budget += budget_aware_generator._minimum_slot_completion_tokens(
                slot_path=str(segment.get("path") or ""),
                fallback_text=str(segment.get("fallback_text") or ""),
            )

    class _BudgetAwareGenerationConfig:
        max_new_tokens = minimum_budget

    budget_aware_generator(
        model=object(),
        input_ids=_FakeIds(),
        logits_processor=None,
        stopping_criteria=None,
        generation_config=_BudgetAwareGenerationConfig(),
    )
    require(
        budget_aware_generator.last_report.get("guided_empty_slot_count") == 0,
        "guided decoding custom_generate shim must not collapse future text slots into empty strings when only the minimum budget remains",
    )
    require(
        budget_aware_generator.last_report.get("guided_compact_fallback_paths") == ["$.summary", "$.issues[0].message"],
        "guided decoding custom_generate shim must reserve enough budget for later slots and use compact non-empty fallbacks when the budget is tight",
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

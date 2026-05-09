from __future__ import annotations

import json
from typing import Any


GUIDED_SEGMENT_KIND_STATIC = "static"
GUIDED_SEGMENT_KIND_TEXT_SLOT = "text_slot"


def append_guided_static_segment(segments: list[dict[str, Any]], text: str) -> None:
    if not text:
        return
    if segments and segments[-1].get("kind") == GUIDED_SEGMENT_KIND_STATIC:
        segments[-1]["text"] = str(segments[-1]["text"]) + text
        return
    segments.append(
        {
            "kind": GUIDED_SEGMENT_KIND_STATIC,
            "text": text,
        }
    )


def build_guided_json_schema_plan(
    document: Any,
    *,
    slot_specs: dict[str, dict[str, Any]],
) -> list[dict[str, Any]]:
    segments: list[dict[str, Any]] = []

    def append_value(value: Any, path: str) -> None:
        slot_spec = slot_specs.get(path)
        if slot_spec is not None:
            if not isinstance(value, str):
                raise ValueError(f"guided text slot must map to a string value: {path}")
            append_guided_static_segment(segments, '"')
            segments.append(
                {
                    "kind": GUIDED_SEGMENT_KIND_TEXT_SLOT,
                    "path": path,
                    "fallback_text": value,
                    "min_non_space_chars": int(slot_spec.get("min_non_space_chars") or 1),
                    "max_chars": int(slot_spec.get("max_chars") or 240),
                }
            )
            return
        if isinstance(value, dict):
            append_guided_static_segment(segments, "{")
            items = list(value.items())
            for index, (key, child) in enumerate(items):
                if index:
                    append_guided_static_segment(segments, ",")
                append_guided_static_segment(segments, json.dumps(str(key), ensure_ascii=False, separators=(",", ":")))
                append_guided_static_segment(segments, ":")
                append_value(child, f"{path}.{key}")
            append_guided_static_segment(segments, "}")
            return
        if isinstance(value, list):
            append_guided_static_segment(segments, "[")
            for index, child in enumerate(value):
                if index:
                    append_guided_static_segment(segments, ",")
                append_value(child, f"{path}[{index}]")
            append_guided_static_segment(segments, "]")
            return
        append_guided_static_segment(segments, json.dumps(value, ensure_ascii=False, separators=(",", ":")))

    append_value(document, "$")
    return segments


class GuidedJsonSchemaCustomGenerator:
    def __init__(
        self,
        *,
        tokenizer: Any,
        plan: list[dict[str, Any]],
        top_k: int = 128,
    ) -> None:
        self.tokenizer = tokenizer
        self.plan = plan
        self.top_k = max(8, int(top_k))
        quote_token_ids = tokenizer.encode('"', add_special_tokens=False)
        if len(quote_token_ids) != 1:
            raise ValueError(
                "guided decoding custom_generate shim currently requires a tokenizer where '\"' encodes as a single token"
            )
        self.quote_token_id = int(quote_token_ids[0])
        self.special_token_ids = {int(token_id) for token_id in (getattr(tokenizer, "all_special_ids", None) or [])}
        self._token_text_cache: dict[int, str] = {}
        self._static_token_cache: dict[str, list[int]] = {}
        self.last_report: dict[str, Any] = {}

    def _encode_static_text(self, text: str) -> list[int]:
        cached = self._static_token_cache.get(text)
        if cached is not None:
            return list(cached)
        token_ids = [int(token_id) for token_id in self.tokenizer.encode(text, add_special_tokens=False)]
        self._static_token_cache[text] = token_ids
        return list(token_ids)

    def _decode_token(self, token_id: int) -> str:
        cached = self._token_text_cache.get(token_id)
        if cached is not None:
            return cached
        text = self.tokenizer.decode(
            [token_id],
            skip_special_tokens=False,
            clean_up_tokenization_spaces=False,
        )
        self._token_text_cache[token_id] = text
        return text

    @staticmethod
    def _non_space_char_count(text: str) -> int:
        return sum(1 for character in text if not character.isspace())

    @staticmethod
    def _is_valid_slot_piece(piece: str) -> bool:
        if not piece:
            return False
        if '"' in piece or "\\" in piece:
            return False
        return not any(ord(character) < 0x20 for character in piece)

    @staticmethod
    def _build_attention_mask_like(input_ids: Any) -> Any:
        return input_ids.new_ones(input_ids.shape)

    def _prime_model(
        self,
        model: Any,
        *,
        input_ids: Any,
        attention_mask: Any,
    ) -> tuple[Any, Any]:
        outputs = model(input_ids=input_ids, attention_mask=attention_mask, use_cache=True)
        return outputs.past_key_values, outputs.logits[:, -1, :]

    def _append_chunk(
        self,
        model: Any,
        *,
        current_ids: Any,
        attention_mask: Any,
        past_key_values: Any,
        token_ids: list[int],
    ) -> tuple[Any, Any, Any, Any]:
        if not token_ids:
            return current_ids, attention_mask, past_key_values, None
        import torch

        chunk = torch.tensor([token_ids], device=current_ids.device, dtype=current_ids.dtype)
        next_ids = torch.cat([current_ids, chunk], dim=-1)
        next_attention_mask = torch.cat([attention_mask, attention_mask.new_ones((attention_mask.shape[0], chunk.shape[-1]))], dim=-1)
        outputs = model(
            input_ids=chunk,
            attention_mask=next_attention_mask,
            past_key_values=past_key_values,
            use_cache=True,
        )
        return next_ids, next_attention_mask, outputs.past_key_values, outputs.logits[:, -1, :]

    def _fallback_slot_tokens(self, fallback_text: str) -> list[int]:
        fallback_ids = self._encode_static_text(fallback_text)
        return [*fallback_ids, self.quote_token_id]

    def _slot_completion_options(
        self,
        *,
        slot_path: str,
        fallback_text: str,
    ) -> list[tuple[str, list[int]]]:
        options: list[tuple[str, list[int]]] = []
        if fallback_text:
            options.append(("fallback", self._fallback_slot_tokens(fallback_text)))
        compact_fallback_text = self._compact_fallback_text(slot_path=slot_path)
        if compact_fallback_text:
            options.append(("compact_fallback", self._fallback_slot_tokens(compact_fallback_text)))
        if not options:
            options.append(("empty", [self.quote_token_id]))
        return options

    def _minimum_slot_completion_tokens(
        self,
        *,
        slot_path: str,
        fallback_text: str,
    ) -> int:
        options = self._slot_completion_options(slot_path=slot_path, fallback_text=fallback_text)
        option_kind, token_ids = min(
            options,
            key=lambda item: (
                len(item[1]),
                0 if item[0] == "compact_fallback" else 1 if item[0] == "fallback" else 2,
            ),
        )
        del option_kind
        return len(token_ids)

    @staticmethod
    def _compact_fallback_text(*, slot_path: str) -> str:
        if slot_path == "$.summary":
            return "请按诊断复核。"
        if slot_path == "$.answers[0].text":
            return "请按诊断与引用复核。"
        if slot_path.endswith(".message"):
            return "请复核当前项。"
        if slot_path.endswith(".title"):
            return "请复核"
        if slot_path.endswith(".rationale"):
            return "建议人工确认后处理。"
        return ""

    def _pick_slot_token(
        self,
        logits: Any,
        *,
        slot_text: str,
        min_non_space_chars: int,
        max_chars: int,
        require_non_space_progress: bool = False,
    ) -> tuple[str, int | None, str | None] | None:
        import torch

        vocab_size = int(logits.shape[-1])
        top_k = min(self.top_k, vocab_size)
        top_values, top_indices = torch.topk(logits[0], k=top_k)
        candidate_scores = {
            int(token_id): float(score)
            for token_id, score in zip(top_indices.tolist(), top_values.tolist(), strict=True)
        }
        if self.quote_token_id not in candidate_scores:
            candidate_scores[self.quote_token_id] = float(logits[0, self.quote_token_id].item())
        ordered_candidates = sorted(candidate_scores.items(), key=lambda item: item[1], reverse=True)
        non_space_chars = self._non_space_char_count(slot_text)
        for token_id, _score in ordered_candidates:
            if token_id == self.quote_token_id:
                if non_space_chars >= min_non_space_chars:
                    return ("end", token_id, None)
                continue
            if token_id in self.special_token_ids:
                continue
            piece = self._decode_token(token_id)
            if not self._is_valid_slot_piece(piece):
                continue
            if require_non_space_progress and self._non_space_char_count(piece) <= 0:
                continue
            if len(slot_text) + len(piece) > max_chars:
                continue
            return ("token", token_id, piece)
        return None

    def __call__(
        self,
        model: Any,
        input_ids: Any,
        logits_processor: Any,
        stopping_criteria: Any,
        generation_config: Any,
        synced_gpus: bool = False,
        streamer: Any | None = None,
        **model_kwargs: Any,
    ) -> Any:
        del logits_processor, stopping_criteria, synced_gpus, streamer
        if input_ids.ndim != 2 or int(input_ids.shape[0]) != 1:
            raise ValueError("guided decoding custom_generate shim currently supports batch_size=1 only")
        max_new_tokens = int(getattr(generation_config, "max_new_tokens", 0) or 0)
        if max_new_tokens <= 0:
            raise ValueError("guided decoding custom_generate shim requires generation_config.max_new_tokens > 0")
        attention_mask = model_kwargs.get("attention_mask")
        current_ids = input_ids.clone()
        current_attention_mask = (
            attention_mask.clone() if attention_mask is not None else self._build_attention_mask_like(current_ids)
        )
        past_key_values, next_logits = self._prime_model(
            model,
            input_ids=current_ids,
            attention_mask=current_attention_mask,
        )
        generated_token_count = 0
        static_segment_count = 0
        static_token_count = 0
        slot_count = 0
        fallback_paths: list[str] = []
        compact_fallback_paths: list[str] = []
        empty_slot_paths: list[str] = []
        relaxed_close_paths: list[str] = []

        suffix_min_required_tokens = [0] * (len(self.plan) + 1)
        for index in range(len(self.plan) - 1, -1, -1):
            segment = self.plan[index]
            kind = str(segment.get("kind") or "").strip()
            if kind == GUIDED_SEGMENT_KIND_STATIC:
                token_count = len(self._encode_static_text(str(segment.get("text") or "")))
            elif kind == GUIDED_SEGMENT_KIND_TEXT_SLOT:
                slot_path = str(segment.get("path") or "")
                fallback_text = str(segment.get("fallback_text") or "")
                token_count = self._minimum_slot_completion_tokens(
                    slot_path=slot_path,
                    fallback_text=fallback_text,
                )
            else:
                raise ValueError(f"unsupported guided segment kind: {kind or '<empty>'}")
            suffix_min_required_tokens[index] = suffix_min_required_tokens[index + 1] + token_count

        for segment_index, segment in enumerate(self.plan):
            kind = str(segment.get("kind") or "").strip()
            if kind == GUIDED_SEGMENT_KIND_STATIC:
                token_ids = self._encode_static_text(str(segment.get("text") or ""))
                if generated_token_count + len(token_ids) + suffix_min_required_tokens[segment_index + 1] > max_new_tokens:
                    raise ValueError("guided decoding custom_generate shim exceeded max_new_tokens while writing fixed scaffold")
                if generated_token_count + len(token_ids) > max_new_tokens:
                    raise ValueError("guided decoding custom_generate shim exceeded max_new_tokens while writing fixed scaffold")
                if token_ids:
                    current_ids, current_attention_mask, past_key_values, next_logits = self._append_chunk(
                        model,
                        current_ids=current_ids,
                        attention_mask=current_attention_mask,
                        past_key_values=past_key_values,
                        token_ids=token_ids,
                    )
                    generated_token_count += len(token_ids)
                    static_token_count += len(token_ids)
                static_segment_count += 1
                continue
            if kind != GUIDED_SEGMENT_KIND_TEXT_SLOT:
                raise ValueError(f"unsupported guided segment kind: {kind or '<empty>'}")
            slot_count += 1
            slot_path = str(segment.get("path") or "")
            fallback_text = str(segment.get("fallback_text") or "")
            min_non_space_chars = int(segment.get("min_non_space_chars") or 1)
            max_chars = int(segment.get("max_chars") or 240)
            remaining_min_tokens_after_slot = suffix_min_required_tokens[segment_index + 1]
            minimum_slot_completion_tokens = self._minimum_slot_completion_tokens(
                slot_path=slot_path,
                fallback_text=fallback_text,
            )
            if generated_token_count + minimum_slot_completion_tokens + remaining_min_tokens_after_slot > max_new_tokens:
                raise ValueError(
                    f"guided decoding custom_generate shim could not fit fallback text for slot {slot_path} within max_new_tokens"
                )
            if generated_token_count + minimum_slot_completion_tokens + remaining_min_tokens_after_slot == max_new_tokens:
                for fallback_kind, fallback_token_ids in self._slot_completion_options(
                    slot_path=slot_path,
                    fallback_text=fallback_text,
                ):
                    if generated_token_count + len(fallback_token_ids) + remaining_min_tokens_after_slot <= max_new_tokens:
                        current_ids, current_attention_mask, past_key_values, next_logits = self._append_chunk(
                            model,
                            current_ids=current_ids,
                            attention_mask=current_attention_mask,
                            past_key_values=past_key_values,
                            token_ids=fallback_token_ids,
                        )
                        generated_token_count += len(fallback_token_ids)
                        if fallback_kind == "fallback":
                            fallback_paths.append(slot_path)
                        elif fallback_kind == "compact_fallback":
                            compact_fallback_paths.append(slot_path)
                        else:
                            empty_slot_paths.append(slot_path)
                        break
                else:
                    raise ValueError(
                        f"guided decoding custom_generate shim could not fit fallback text for slot {slot_path} within max_new_tokens"
                    )
                continue
            slot_text = ""
            slot_started = False
            slot_closed = False
            while generated_token_count < max_new_tokens:
                choice = self._pick_slot_token(
                    next_logits,
                    slot_text=slot_text,
                    min_non_space_chars=min_non_space_chars,
                    max_chars=max_chars,
                    require_non_space_progress=self._non_space_char_count(slot_text) < min_non_space_chars,
                )
                if choice is None:
                    break
                choice_kind, token_id, piece = choice
                if choice_kind == "end":
                    current_ids, current_attention_mask, past_key_values, next_logits = self._append_chunk(
                        model,
                        current_ids=current_ids,
                        attention_mask=current_attention_mask,
                        past_key_values=past_key_values,
                        token_ids=[self.quote_token_id],
                    )
                    generated_token_count += 1
                    slot_closed = True
                    break
                if token_id is None or piece is None:
                    break
                piece_non_space_chars = self._non_space_char_count(piece)
                # Always leave room to close the current JSON string slot and to
                # finish the remaining fixed scaffold. When budget gets tight, the
                # slot will be shortened instead of corrupting the whole response.
                remaining_required_non_space_chars = max(
                    0,
                    min_non_space_chars - self._non_space_char_count(slot_text) - piece_non_space_chars,
                )
                if (
                    generated_token_count
                    + 1
                    + 1
                    + remaining_required_non_space_chars
                    + remaining_min_tokens_after_slot
                    > max_new_tokens
                ):
                    break
                slot_started = True
                current_ids, current_attention_mask, past_key_values, next_logits = self._append_chunk(
                    model,
                    current_ids=current_ids,
                    attention_mask=current_attention_mask,
                    past_key_values=past_key_values,
                    token_ids=[token_id],
                )
                generated_token_count += 1
                slot_text += piece
                if len(slot_text) >= max_chars:
                    break
            if not slot_started:
                appended_fallback = False
                for fallback_kind, fallback_token_ids in self._slot_completion_options(
                    slot_path=slot_path,
                    fallback_text=fallback_text,
                ):
                    if generated_token_count + len(fallback_token_ids) + remaining_min_tokens_after_slot <= max_new_tokens:
                        current_ids, current_attention_mask, past_key_values, next_logits = self._append_chunk(
                            model,
                            current_ids=current_ids,
                            attention_mask=current_attention_mask,
                            past_key_values=past_key_values,
                            token_ids=fallback_token_ids,
                        )
                        generated_token_count += len(fallback_token_ids)
                        appended_fallback = True
                        if fallback_kind == "fallback":
                            fallback_paths.append(slot_path)
                        elif fallback_kind == "compact_fallback":
                            compact_fallback_paths.append(slot_path)
                        else:
                            empty_slot_paths.append(slot_path)
                        break
                if appended_fallback:
                    continue
                raise ValueError(
                    f"guided decoding custom_generate shim could not fit fallback text for slot {slot_path} within max_new_tokens"
                )
            if not slot_closed:
                if self._non_space_char_count(slot_text) <= 0:
                    raise ValueError(
                        f"guided decoding custom_generate shim exhausted max_new_tokens before opening slot {slot_path}"
                    )
                if generated_token_count + 1 + remaining_min_tokens_after_slot > max_new_tokens:
                    raise ValueError(
                        f"guided decoding custom_generate shim exhausted max_new_tokens before closing slot {slot_path}"
                    )
                current_ids, current_attention_mask, past_key_values, next_logits = self._append_chunk(
                    model,
                    current_ids=current_ids,
                    attention_mask=current_attention_mask,
                    past_key_values=past_key_values,
                    token_ids=[self.quote_token_id],
                )
                generated_token_count += 1
                if self._non_space_char_count(slot_text) < min_non_space_chars:
                    relaxed_close_paths.append(slot_path)

        self.last_report = {
            "backend": "custom_generate_scaffold_slots",
            "guided_static_segment_count": static_segment_count,
            "guided_static_token_count": static_token_count,
            "guided_slot_count": slot_count,
            "guided_fallback_slot_count": len(fallback_paths),
            "guided_fallback_paths": fallback_paths,
            "guided_compact_fallback_slot_count": len(compact_fallback_paths),
            "guided_compact_fallback_paths": compact_fallback_paths,
            "guided_empty_slot_count": len(empty_slot_paths),
            "guided_empty_slot_paths": empty_slot_paths,
            "guided_relaxed_close_slot_count": len(relaxed_close_paths),
            "guided_relaxed_close_slot_paths": relaxed_close_paths,
        }
        return current_ids

#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.inference import (  # noqa: E402
    build_candidate_response_dump,
    run_inference,
    validate_request_document,
    validate_response_document,
)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Run the minimal RadishMind inference flow for radish / answer_docs_question.",
    )
    input_group = parser.add_mutually_exclusive_group(required=True)
    input_group.add_argument("--sample", help="Path to an eval sample file; inference uses sample.input_request.")
    input_group.add_argument("--request", help="Path to a CopilotRequest json file.")
    parser.add_argument("--provider", choices=["mock", "openai-compatible"], default="mock")
    parser.add_argument("--model", default="", help="Provider model name. Required for openai-compatible when env is absent.")
    parser.add_argument("--base-url", default="", help="Provider base URL or /v1 endpoint for openai-compatible.")
    parser.add_argument("--api-key", default="", help="Provider API key for openai-compatible.")
    parser.add_argument("--temperature", type=float, default=0.0)
    parser.add_argument("--sample-id", default="", help="Optional sample_id override when using --request.")
    parser.add_argument("--response-output", default="", help="Optional path to write the normalized response json.")
    parser.add_argument("--dump-output", default="", help="Optional path to write a raw candidate response dump.")
    return parser.parse_args()


def resolve_path(path_value: str) -> Path:
    path = Path(path_value)
    if not path.is_absolute():
        path = (REPO_ROOT / path).resolve()
    return path


def load_json(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json '{path}': {exc}") from exc


def write_json(path_value: str, document: Any) -> None:
    path = resolve_path(path_value)
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(document, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def main() -> int:
    args = parse_args()

    if args.sample:
        sample_path = resolve_path(args.sample)
        sample = load_json(sample_path)
        copilot_request = sample.get("input_request")
        sample_id = str(sample.get("sample_id") or "").strip()
        if not isinstance(copilot_request, dict):
            raise SystemExit("--sample must point to a json file containing input_request")
    else:
        request_path = resolve_path(args.request)
        copilot_request = load_json(request_path)
        sample_id = args.sample_id.strip() or f"adhoc-{copilot_request.get('task', 'request')}"

    validate_request_document(copilot_request)
    result = run_inference(
        copilot_request,
        provider=args.provider,
        model=args.model.strip() or None,
        base_url=args.base_url.strip() or None,
        api_key=args.api_key.strip() or None,
        temperature=args.temperature,
    )
    validate_response_document(result["response"])

    if args.response_output:
        write_json(args.response_output, result["response"])
    if args.dump_output:
        dump_document = build_candidate_response_dump(
            copilot_request=copilot_request,
            sample_id=sample_id,
            inference_result=result,
        )
        write_json(args.dump_output, dump_document)

    print(json.dumps(result["response"], ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

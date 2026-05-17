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

from services.gateway import GatewayOptions, handle_copilot_request  # noqa: E402
from services.runtime.inference_support import describe_provider_inventory  # noqa: E402
from services.runtime.provider_registry import describe_provider_registry  # noqa: E402


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Bridge Go platform requests to the canonical Python runtime.")
    subparsers = parser.add_subparsers(dest="command", required=True)

    def add_bridge_arguments(target_parser: argparse.ArgumentParser) -> None:
        target_parser.add_argument("--request-file", default="", help="Optional request json file. Reads stdin when omitted.")
        target_parser.add_argument("--provider", default="mock")
        target_parser.add_argument("--provider-profile", default="")
        target_parser.add_argument("--model", default="")
        target_parser.add_argument("--base-url", default="")
        target_parser.add_argument("--api-key", default="")
        target_parser.add_argument("--temperature", type=float, default=0.0)
        target_parser.add_argument("--request-timeout-seconds", type=float, default=120.0)

    envelope_parser = subparsers.add_parser("envelope", help="Read a CopilotRequest and emit a CopilotGatewayEnvelope.")
    add_bridge_arguments(envelope_parser)
    stream_parser = subparsers.add_parser("stream", help="Read a CopilotRequest and emit JSONL stream events.")
    add_bridge_arguments(stream_parser)
    subparsers.add_parser("providers", help="Emit the canonical provider registry description.")
    subparsers.add_parser("inventory", help="Emit the canonical provider and profile inventory.")
    return parser.parse_args()


def load_request_document(path_text: str) -> dict[str, Any]:
    if path_text.strip():
        return json.loads(Path(path_text).read_text(encoding="utf-8"))
    return json.load(sys.stdin)


def run_envelope(args: argparse.Namespace) -> int:
    request_document = load_request_document(args.request_file)
    envelope = handle_copilot_request(
        request_document,
        options=GatewayOptions(
            provider=args.provider,
            provider_profile=args.provider_profile,
            model=args.model,
            base_url=args.base_url,
            api_key=args.api_key,
            temperature=args.temperature,
            request_timeout_seconds=args.request_timeout_seconds,
        ),
    )
    json.dump(envelope, sys.stdout, ensure_ascii=False, indent=2)
    sys.stdout.write("\n")
    return 0


def run_stream(args: argparse.Namespace) -> int:
    request_document = load_request_document(args.request_file)

    def emit_event(event: dict[str, Any]) -> None:
        json.dump(event, sys.stdout, ensure_ascii=False)
        sys.stdout.write("\n")
        sys.stdout.flush()

    envelope = handle_copilot_request(
        request_document,
        options=GatewayOptions(
            provider=args.provider,
            provider_profile=args.provider_profile,
            model=args.model,
            base_url=args.base_url,
            api_key=args.api_key,
            temperature=args.temperature,
            request_timeout_seconds=args.request_timeout_seconds,
        ),
        stream_handler=emit_event,
    )
    json.dump({"type": "completed", "envelope": envelope}, sys.stdout, ensure_ascii=False)
    sys.stdout.write("\n")
    return 0


def run_providers() -> int:
    json.dump(describe_provider_registry(), sys.stdout, ensure_ascii=False, indent=2)
    sys.stdout.write("\n")
    return 0


def run_inventory() -> int:
    json.dump(describe_provider_inventory(), sys.stdout, ensure_ascii=False, indent=2)
    sys.stdout.write("\n")
    return 0


def main() -> int:
    args = parse_args()
    if args.command == "envelope":
        return run_envelope(args)
    if args.command == "stream":
        return run_stream(args)
    if args.command == "providers":
        return run_providers()
    if args.command == "inventory":
        return run_inventory()
    raise SystemExit(f"unsupported command: {args.command}")


if __name__ == "__main__":
    raise SystemExit(main())

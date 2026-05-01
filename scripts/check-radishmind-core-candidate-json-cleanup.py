#!/usr/bin/env python3
from __future__ import annotations

import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from scripts.eval.core_candidate_json import extract_json_object  # noqa: E402


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def main() -> int:
    extracted = extract_json_object(
        "```json\n"
        "{\n"
        '  "schema_version": 1,\n'
        '  "answers": [\n'
        '    {"kind": "direct_answer", "text": "ok",},\n'
        "  ],\n"
        "}\n"
        "```"
    )
    require(extracted.cleanup_applied is True, "trailing comma cleanup must be reported")
    require(extracted.document["answers"][0]["kind"] == "direct_answer", "cleaned JSON document was not parsed")

    clean = extract_json_object('{"schema_version": 1, "answers": []}')
    require(clean.cleanup_applied is False, "valid JSON must not report cleanup")

    print("radishmind core candidate json cleanup check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

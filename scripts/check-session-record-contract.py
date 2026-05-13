#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent
SESSION_SCHEMA_PATH = REPO_ROOT / "contracts/session-record.schema.json"
SESSION_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/session-record-basic.json"


def load_json_document(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json document '{path.relative_to(REPO_ROOT)}': {exc}") from exc


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def check_session_record(record: dict[str, Any]) -> None:
    require(record.get("kind") == "conversation_session_record", "session record kind mismatch")
    require(record.get("session_id") == record.get("conversation_id"), "v1 fixture keeps session_id aligned with conversation_id")

    history_policy = record.get("history_policy")
    require(isinstance(history_policy, dict), "history_policy must be an object")
    require(history_policy.get("mode") in {"stateless", "windowed", "summary_only", "disabled"}, "unexpected history policy mode")
    max_turns = int(history_policy.get("max_turns") or 0)
    if history_policy.get("mode") == "windowed":
        require(1 <= max_turns <= 20, "windowed history policy must keep a bounded reviewable window")
    else:
        require(max_turns == 0 or max_turns <= 20, "non-windowed history policy must not declare a large retained window")

    compression = history_policy.get("compression")
    require(isinstance(compression, dict), "history compression must be an object")
    if compression.get("strategy") == "none":
        require(compression.get("summary_artifact_id") is None, "uncompressed history must not point to a summary artifact")
    else:
        require(bool(str(compression.get("summary_artifact_id") or "").strip()), "compressed history must name its summary artifact")

    state_policy = record.get("state_policy")
    if state_policy is not None:
        require(isinstance(state_policy, dict), "state_policy must be an object when present")
        require(state_policy.get("durable_memory_enabled") is False, "session state policy must not enable durable memory")
        require(
            state_policy.get("session_state_scope") in {"request_local", "northbound_metadata"},
            "session state scope must stay local or metadata-only",
        )
        require(
            state_policy.get("tool_result_cache_scope") in {"disabled", "request_local", "session_recovery_checkpoint"},
            "unexpected tool result cache scope",
        )
        if state_policy.get("tool_result_cache_scope") == "session_recovery_checkpoint":
            require(
                state_policy.get("recovery_checkpoint_scope") in {"audit_refs_only", "audit_and_result_refs"},
                "session recovery cache must declare checkpoint reference scope",
            )

    recovery_record = record.get("recovery_record")
    require(isinstance(recovery_record, dict), "recovery_record must be an object")
    checkpoints = recovery_record.get("checkpoints")
    require(isinstance(checkpoints, list), "recovery checkpoints must be a list")
    if recovery_record.get("status") == "available":
        require(recovery_record.get("replayable") is True, "available recovery record must be replayable")
        require(bool(str(recovery_record.get("last_stable_turn_id") or "").strip()), "available recovery record must name last stable turn")
        require(len(checkpoints) > 0, "available recovery record must keep at least one checkpoint reference")
    if recovery_record.get("status") == "not_required":
        require(recovery_record.get("last_stable_turn_id") is None, "not-required recovery must not invent a stable turn")

    audit = record.get("audit")
    require(isinstance(audit, dict), "audit must be an object")
    require(audit.get("advisory_only") is True, "session record must remain advisory-only")
    require(audit.get("writes_business_truth") is False, "session record must not write business truth")
    require(isinstance(audit.get("notes"), list) and len(audit["notes"]) > 0, "session audit must include notes")


def main() -> int:
    schema = load_json_document(SESSION_SCHEMA_PATH)
    fixture = load_json_document(SESSION_FIXTURE_PATH)

    jsonschema.Draft202012Validator.check_schema(schema)
    jsonschema.validate(fixture, schema)
    if not isinstance(fixture, dict):
        raise SystemExit("session record fixture must be an object")
    check_session_record(fixture)

    print("session record contract smoke passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

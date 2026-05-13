#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent
CHECKPOINT_SCHEMA_PATH = REPO_ROOT / "contracts/session-recovery-checkpoint.schema.json"
MANIFEST_SCHEMA_PATH = REPO_ROOT / "contracts/session-recovery-checkpoint-manifest.schema.json"
READ_SCHEMA_PATH = REPO_ROOT / "contracts/session-recovery-checkpoint-read.schema.json"
CHECKPOINT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/session-recovery-checkpoint-basic.json"
MANIFEST_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/session-recovery-checkpoint-manifest-basic.json"
READ_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/session-recovery-checkpoint-read-basic.json"
TOOL_AUDIT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/tool-audit-record-basic.json"


def load_json_document(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json document '{path.relative_to(REPO_ROOT)}': {exc}") from exc


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def check_checkpoint(checkpoint: dict[str, Any]) -> None:
    storage_policy = checkpoint.get("storage_policy")
    require(isinstance(storage_policy, dict), "storage_policy must be an object")
    require(storage_policy.get("durable_memory_enabled") is False, "checkpoint must not enable durable memory")
    require(storage_policy.get("writes_business_truth") is False, "checkpoint must not write business truth")
    require(
        storage_policy.get("materialized_result_storage") == "metadata_only",
        "v1 checkpoint fixture must store tool result metadata only",
    )

    replay_policy = checkpoint.get("replay_policy")
    require(isinstance(replay_policy, dict), "replay_policy must be an object")
    require(replay_policy.get("auto_replay_enabled") is False, "checkpoint must not enable automatic replay")
    if replay_policy.get("replayable") is True:
        require(
            replay_policy.get("requires_confirmation_for_actions") is True,
            "replayable checkpoint must require confirmation for actions",
        )

    refs = checkpoint.get("refs")
    require(isinstance(refs, list) and len(refs) > 0, "checkpoint refs must be a non-empty list")
    ref_kinds = {str(ref.get("kind")) for ref in refs if isinstance(ref, dict)}
    for expected_kind in ("request", "session_record", "tool_audit"):
        require(expected_kind in ref_kinds, f"checkpoint must include {expected_kind} ref")
    allowed_ref_kinds = set(replay_policy.get("allowed_ref_kinds") or [])
    require(ref_kinds.issubset(allowed_ref_kinds), "checkpoint refs must stay within replay policy allowed_ref_kinds")

    state_summary = checkpoint.get("state_summary")
    require(isinstance(state_summary, dict), "state_summary must be an object")
    require(
        state_summary.get("contains_tool_result_metadata") is True,
        "checkpoint fixture must declare tool result metadata",
    )
    require(
        state_summary.get("contains_materialized_tool_results") is False,
        "checkpoint must not contain materialized tool results",
    )
    require(state_summary.get("contains_business_truth") is False, "checkpoint must not contain business truth")

    audit = checkpoint.get("audit")
    require(isinstance(audit, dict), "audit must be an object")
    require(audit.get("advisory_only") is True, "checkpoint audit must remain advisory-only")
    require(audit.get("writes_business_truth") is False, "checkpoint audit must not write business truth")
    require(audit.get("durable_memory_written") is False, "checkpoint audit must not write durable memory")


def check_manifest(manifest: dict[str, Any], checkpoint: dict[str, Any]) -> None:
    storage_policy = manifest.get("storage_policy")
    require(isinstance(storage_policy, dict), "manifest storage_policy must be an object")
    require(storage_policy.get("durable_memory_enabled") is False, "manifest must not enable durable memory")
    require(storage_policy.get("auto_replay_enabled") is False, "manifest must not enable automatic replay")

    checkpoints = manifest.get("checkpoints")
    require(isinstance(checkpoints, list) and len(checkpoints) > 0, "manifest checkpoints must be non-empty")
    matching = [
        item
        for item in checkpoints
        if isinstance(item, dict) and item.get("checkpoint_id") == checkpoint.get("checkpoint_id")
    ]
    require(len(matching) == 1, "manifest must index the checkpoint fixture exactly once")
    entry = matching[0]
    require(entry.get("session_id") == checkpoint.get("session_id"), "manifest session_id must match checkpoint")
    require(entry.get("turn_id") == checkpoint.get("turn_id"), "manifest turn_id must match checkpoint")
    require(entry.get("replayable") == checkpoint.get("replay_policy", {}).get("replayable"), "manifest replayable mismatch")
    require(
        entry.get("record_ref") == "scripts/checks/fixtures/session-recovery-checkpoint-basic.json",
        "manifest record_ref must point at committed checkpoint fixture",
    )

    audit = manifest.get("audit")
    require(isinstance(audit, dict), "manifest audit must be an object")
    require(audit.get("advisory_only") is True, "manifest audit must remain advisory-only")
    require(audit.get("writes_business_truth") is False, "manifest audit must not write business truth")


def check_read_result(read_result: dict[str, Any], checkpoint: dict[str, Any], tool_audit: dict[str, Any]) -> None:
    api_boundary = read_result.get("api_boundary")
    require(isinstance(api_boundary, dict), "read api_boundary must be an object")
    require(api_boundary.get("implemented") is False, "checkpoint read API boundary must not claim implementation")
    require(api_boundary.get("response_shape") == "metadata_refs_only", "checkpoint read response must be metadata refs only")

    request = read_result.get("request")
    require(isinstance(request, dict), "read request must be an object")
    require(request.get("checkpoint_id") == checkpoint.get("checkpoint_id"), "read request checkpoint_id mismatch")
    require(request.get("session_id") == checkpoint.get("session_id"), "read request session_id mismatch")
    require(request.get("include_materialized_results") is False, "read request must not include materialized results")

    result = read_result.get("result")
    require(isinstance(result, dict), "read result must be an object")
    require(result.get("status") == "found", "read fixture must describe a found checkpoint")
    require(result.get("checkpoint_id") == checkpoint.get("checkpoint_id"), "read result checkpoint_id mismatch")
    require(result.get("session_id") == checkpoint.get("session_id"), "read result session_id mismatch")
    require(result.get("turn_id") == checkpoint.get("turn_id"), "read result turn_id mismatch")
    require(
        result.get("checkpoint_ref") == "scripts/checks/fixtures/session-recovery-checkpoint-basic.json",
        "read result must point at committed checkpoint fixture",
    )

    checkpoint_refs = {
        (str(ref.get("kind")), str(ref.get("ref")), bool(ref.get("required_for_recovery")))
        for ref in checkpoint.get("refs", [])
        if isinstance(ref, dict)
    }
    read_refs = {
        (str(ref.get("kind")), str(ref.get("ref")), bool(ref.get("required_for_recovery")))
        for ref in result.get("refs", [])
        if isinstance(ref, dict)
    }
    require(read_refs == checkpoint_refs, "read result refs must mirror checkpoint refs")

    tool_audit_summary = result.get("tool_audit_summary")
    require(isinstance(tool_audit_summary, dict), "read result must include tool_audit_summary")
    require(tool_audit_summary.get("audit_id") == tool_audit.get("audit_id"), "tool audit summary audit_id mismatch")
    require(tool_audit_summary.get("tool_id") == tool_audit.get("tool_id"), "tool audit summary tool_id mismatch")
    require(
        tool_audit_summary.get("policy_decision") == tool_audit.get("policy_decision", {}).get("decision"),
        "tool audit summary policy decision mismatch",
    )
    require(
        tool_audit_summary.get("requires_confirmation") == tool_audit.get("policy_decision", {}).get("requires_confirmation"),
        "tool audit summary confirmation policy mismatch",
    )
    require(tool_audit_summary.get("execution_enabled") is False, "read route must not expose executable tooling")
    require(tool_audit_summary.get("execution_status") == "not_executed", "read route must keep tool execution not_executed")
    require(tool_audit_summary.get("result_cache_mode") == "metadata_only", "read route must expose metadata-only cache mode")
    require(tool_audit_summary.get("result_ref") is None, "read route must not expose a materialized result ref")
    require(tool_audit_summary.get("durable_memory_written") is False, "read route must not write durable memory")
    require(tool_audit_summary.get("writes_business_truth") is False, "read route must not write business truth")

    replay_policy = result.get("replay_policy")
    checkpoint_replay_policy = checkpoint.get("replay_policy") or {}
    require(isinstance(replay_policy, dict), "read replay_policy must be an object")
    require(replay_policy.get("replayable") == checkpoint_replay_policy.get("replayable"), "read replayable mismatch")
    require(replay_policy.get("auto_replay_enabled") is False, "read result must not enable automatic replay")
    require(
        replay_policy.get("requires_confirmation_for_actions")
        == checkpoint_replay_policy.get("requires_confirmation_for_actions"),
        "read confirmation policy mismatch",
    )

    state_summary = result.get("state_summary")
    checkpoint_state_summary = checkpoint.get("state_summary") or {}
    require(isinstance(state_summary, dict), "read state_summary must be an object")
    require(
        state_summary.get("contains_materialized_tool_results") is False,
        "read result must not expose materialized tool results",
    )
    require(state_summary.get("contains_business_truth") is False, "read result must not expose business truth")
    require(
        state_summary.get("contains_tool_result_metadata")
        == checkpoint_state_summary.get("contains_tool_result_metadata"),
        "read tool metadata summary mismatch",
    )

    access_policy = read_result.get("access_policy")
    require(isinstance(access_policy, dict), "read access_policy must be an object")
    require(access_policy.get("metadata_only") is True, "read access policy must be metadata-only")
    require(access_policy.get("materialized_results_included") is False, "read access policy must not include results")
    require(access_policy.get("durable_memory_enabled") is False, "read access policy must not enable durable memory")
    require(access_policy.get("writes_business_truth") is False, "read access policy must not write business truth")
    require(access_policy.get("auto_replay_enabled") is False, "read access policy must not enable auto replay")

    audit = read_result.get("audit")
    require(isinstance(audit, dict), "read audit must be an object")
    require(audit.get("advisory_only") is True, "read audit must remain advisory-only")


def main() -> int:
    checkpoint_schema = load_json_document(CHECKPOINT_SCHEMA_PATH)
    manifest_schema = load_json_document(MANIFEST_SCHEMA_PATH)
    read_schema = load_json_document(READ_SCHEMA_PATH)
    checkpoint_fixture = load_json_document(CHECKPOINT_FIXTURE_PATH)
    manifest_fixture = load_json_document(MANIFEST_FIXTURE_PATH)
    read_fixture = load_json_document(READ_FIXTURE_PATH)
    tool_audit_fixture = load_json_document(TOOL_AUDIT_FIXTURE_PATH)

    for schema in (checkpoint_schema, manifest_schema, read_schema):
        jsonschema.Draft202012Validator.check_schema(schema)

    jsonschema.validate(checkpoint_fixture, checkpoint_schema)
    jsonschema.validate(manifest_fixture, manifest_schema)
    jsonschema.validate(read_fixture, read_schema)
    if not isinstance(checkpoint_fixture, dict):
        raise SystemExit("session recovery checkpoint fixture must be an object")
    if not isinstance(manifest_fixture, dict):
        raise SystemExit("session recovery checkpoint manifest fixture must be an object")
    if not isinstance(read_fixture, dict):
        raise SystemExit("session recovery checkpoint read fixture must be an object")
    if not isinstance(tool_audit_fixture, dict):
        raise SystemExit("tool audit fixture must be an object")

    check_checkpoint(checkpoint_fixture)
    check_manifest(manifest_fixture, checkpoint_fixture)
    check_read_result(read_fixture, checkpoint_fixture, tool_audit_fixture)

    print("session recovery checkpoint contract smoke passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

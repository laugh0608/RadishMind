"""Canonical control contracts and independent cancellation/accounting examples."""
from __future__ import annotations

import copy
import json
import unittest
from pathlib import Path

from jsonschema import Draft202012Validator
from referencing import Registry, Resource

ROOT = Path(__file__).resolve().parents[4]
FIXTURES = ROOT / "scripts/eval/fixtures/development-control"
NAMES = ["delegation", "reservation", "receipt", "event", "tool-action", "tool-call", "plan"]


def schema_name(name):
    return "development-" + ("attempt-" if name in NAMES[:4] or name == "plan" else "") + name


class DevelopmentControlContractTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.schemas = {name: json.loads((ROOT / "contracts" / (schema_name(name) + ".schema.json")).read_text()) for name in NAMES}
        registry = Registry().with_resources((value["$id"], Resource.from_contents(value)) for value in cls.schemas.values())
        cls.validators = {}
        for name, value in cls.schemas.items():
            Draft202012Validator.check_schema(value)
            cls.validators[name] = Draft202012Validator(value, registry=registry)

    def test_positive_fixtures_and_exact_fields(self):
        for name in ["delegation", "reservation", "receipt", "tool-action", "tool-call", "plan"]:
            data = json.loads((FIXTURES / (name + ".json")).read_text())
            validator = self.validators[name]
            validator.validate(data)
            for key in data:
                changed = copy.deepcopy(data)
                del changed[key]
                self.assertFalse(validator.is_valid(changed), (name, key))
                changed[key] = None
                self.assertFalse(validator.is_valid(changed), (name, key, "null"))
            self.assertFalse(validator.is_valid(dict(data, api_key="forbidden")), name)

    def test_no_authority_or_budget_expansion(self):
        d = json.loads((FIXTURES / "delegation.json").read_text())
        for field, value in [("audience", "public_chat"), ("revision", 0), ("executor_instance_id", "http://outside"), ("token", "secret")]:
            changed = dict(d, **{field: value})
            self.assertFalse(self.validators["delegation"].is_valid(changed))
        for field, value in [("max_model_calls", 9), ("max_tool_calls", 33), ("max_tokens", 0), ("max_cost_microusd", -1)]:
            changed = copy.deepcopy(d)
            changed["limits"][field] = value
            self.assertFalse(self.validators["delegation"].is_valid(changed))
        r = json.loads((FIXTURES / "reservation.json").read_text())
        for field, value in [("kind", "shell"), ("token_bound", 0), ("payload_digest", "sha256:" + "a" * 64 + "\n"), ("sequence", 41)]:
            self.assertFalse(self.validators["reservation"].is_valid(dict(r, **{field: value})))
        p = json.loads((FIXTURES / "receipt.json").read_text())
        self.assertFalse(self.validators["receipt"].is_valid(dict(p, status="missing")))
        self.validators["receipt"].validate(dict(p, status="missing", tokens=0, cost_microusd=0))

    def test_tool_payload_has_no_transport_metadata_or_shell(self):
        action = json.loads((FIXTURES / "tool-action.json").read_text())
        for path in ["../list.go", "/list.go", "a/../list.go", "a//b", "a/", "C:target", "a\\b", "a\x00b"]:
            self.assertFalse(self.validators["tool-action"].is_valid(dict(action, path=path)), path)
        for extra in ["summary", "command", "api_key"]:
            self.assertFalse(self.validators["tool-action"].is_valid(dict(action, **{extra: "forbidden"})))
        self.assertFalse(self.validators["tool-action"].is_valid(dict(action, action="execute_shell")))
        call = json.loads((FIXTURES / "tool-call.json").read_text())
        call["arguments"]["summary"] = "SDK display metadata"
        self.assertFalse(self.validators["tool-call"].is_valid(call))

    def test_durable_control_events(self):
        base = json.loads((FIXTURES / "cancel-trace.json").read_text())[0]
        for event_type in ["tool_offer_registered", "tool_offer_queued"]:
            event = dict(base, type=event_type, request_id="tool_fixture")
            self.validators["event"].validate(event)
            self.assertFalse(self.validators["event"].is_valid(dict(event, request_id="")))
        for event_type, phase, reason in [
            ("attempt_interrupted", "outcome_unknown", "runtime_interrupted"),
            ("attempt_failed", "failed", "native_response_invalid"),
        ]:
            event = dict(base, type=event_type, phase=phase, reason=reason, termination_status="pending")
            self.validators["event"].validate(event)
            self.assertFalse(self.validators["event"].is_valid(dict(event, phase="active")))

    def test_plan_does_not_grant_ambient_tools(self):
        plan = json.loads((FIXTURES / "plan.json").read_text())
        for extra in ["command", "provider_url", "credential", "authority"]:
            self.assertFalse(self.validators["plan"].is_valid(dict(plan, **{extra: "forbidden"})))
        for field, value in [("read_paths", ["../main.go"]), ("read_paths", ["main.go", "main.go"]), ("verification_profile", "shell")]:
            changed = copy.deepcopy(plan)
            changed["tools"][field] = value
            self.assertFalse(self.validators["plan"].is_valid(changed))

    def test_worker_wire_schema_and_no_arbitrary_reports(self):
        schema = json.loads((ROOT / "contracts/development-sdk-worker.schema.json").read_text())
        Draft202012Validator.check_schema(schema)
        registry = Registry().with_resources((value["$id"], Resource.from_contents(value)) for value in [schema, *self.schemas.values()])
        validator = Draft202012Validator(schema, registry=registry)
        frames = json.loads((FIXTURES / "worker-events.json").read_text())
        for frame in frames:
            validator.validate(frame)
            for name in frame:
                changed = dict(frame)
                del changed[name]
                self.assertFalse(validator.is_valid(changed))
            self.assertFalse(validator.is_valid(dict(frame, raw_text="forbidden")))
        self.assertFalse(validator.is_valid(dict(frames[1], event_type="cancelled")))
        self.assertFalse(validator.is_valid(dict(frames[0], kind="tool")))
        error = dict(frames[1], event_type="error", error_code="model_channel")
        validator.validate(error)
        self.assertFalse(validator.is_valid(dict(frames[1], event_type="error")))
        self.assertFalse(validator.is_valid(dict(error, error_code="raw exception with secret")))
        self.assertFalse(validator.is_valid(dict(error, event_type="security")))

    def test_sandbox_reports_cannot_promote_incomplete_tests(self):
        schema = json.loads((ROOT / "contracts/development-code-sandbox.schema.json").read_text())
        Draft202012Validator.check_schema(schema)
        validator = Draft202012Validator(schema)
        observations = json.loads((FIXTURES / "sandbox-observations.json").read_text())
        ledger = json.loads((FIXTURES / "sandbox-resources.json").read_text())
        for value in [*observations, ledger]:
            validator.validate(value)
            for field in value:
                changed = copy.deepcopy(value)
                del changed[field]
                self.assertFalse(validator.is_valid(changed), field)
            self.assertFalse(validator.is_valid(dict(value, command="forbidden")))
        for field, value in [("exit_code", 137), ("oom_killed", True), ("output_truncated", True)]:
            changed = copy.deepcopy(observations[-1])
            changed["verification"][field] = value
            self.assertFalse(validator.is_valid(changed), field)
        changed = copy.deepcopy(observations[1])
        changed["path"] = "go.mod"
        self.assertFalse(validator.is_valid(changed))

    def test_cancel_trace_preserves_unknown_budget_until_late_receipt(self):
        trace = json.loads((FIXTURES / "cancel-trace.json").read_text())
        for event in trace:
            self.validators["event"].validate(event)
            self.assertFalse(self.validators["event"].is_valid(dict(event, raw_error="forbidden")))
        self.assertEqual([e["sequence"] for e in trace], list(range(1, 7)))
        self.assertEqual(len({e["event_id"] for e in trace}), 6)
        cancelled, late = trace[-2:]
        self.assertEqual(cancelled["phase"], late["phase"], "cancelled")
        self.assertEqual(cancelled["usage_status"], "unknown")
        self.assertEqual(cancelled["termination_status"], "confirmed")
        self.assertFalse(self.validators["event"].is_valid(dict(cancelled, termination_status="pending")))
        self.assertEqual(cancelled["budget"]["held_cost_microusd"], 1000)
        self.assertEqual(late["budget"]["spent_cost_microusd"], 500)
        self.assertEqual(late["budget"]["held_cost_microusd"], 0)
        self.assertFalse(self.validators["event"].is_valid(dict(cancelled, type="attempt_completed")))


if __name__ == "__main__":
    unittest.main()

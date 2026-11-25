"""Public delivery schema validation using only checked-in canonical contracts."""

import copy
import json
import unittest
from pathlib import Path

from jsonschema import Draft202012Validator, FormatChecker
from referencing import Registry, Resource

ROOT = Path(__file__).resolve().parents[4]
SCHEMA_NAME = "development-task-delivery.schema.json"


class DevelopmentTaskDeliveryContractTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        contracts = ROOT / "contracts"
        cls.schema = json.loads((contracts / SCHEMA_NAME).read_text())
        cls.registry = Registry().with_resources(
            (schema["$id"], Resource.from_contents(schema))
            for schema in (json.loads(path.read_text()) for path in contracts.glob("development-*.schema.json"))
        )
        cls.bundle = json.loads((ROOT / "scripts/eval/fixtures/development-control/artifact.json").read_text())
        Draft202012Validator.check_schema(cls.schema)

    def validator(self, name):
        return Draft202012Validator(
            {"$ref": f"{self.schema['$id']}#/$defs/{name}"},
            registry=self.registry,
            format_checker=FormatChecker(),
        )

    def test_write_requests_are_bounded_and_closed(self):
        schema = self.validator("create_task_request")
        limits = dict(max_model_calls=4, max_tool_calls=3, max_tokens=1000, max_cost_microusd=1000)
        request = dict(request_id="new_req", source_profile_id="probe_list_filter", model_ref="profile:fixture", limits=limits)
        schema.validate(request)
        for changes in (
            {"account_id": "other_actor"}, {"workspace_id": "other_workspace"}, {"source_profile_id": "git"},
            {"source_profile_id": None}, {"REQUEST_ID": "new_req"}, {"request_id": "foo\n"},
            {"model_ref": "https://provider.invalid"}, {"limits": dict(limits, max_tool_calls=4)},
            {"limits": dict(limits, max_model_calls=5)}, {"limits": dict(limits, max_cost_microusd=0)},
            {"limits": dict(limits, shell=True)},
        ):
            with self.subTest(changes=changes):
                self.assertFalse(schema.is_valid(dict(request, **changes)))
        for field in request:
            self.assertFalse(schema.is_valid({key: value for key, value in request.items() if key != field}), field)

    def test_plan_reuses_attempt_contract_and_limits_fixed_tools(self):
        attempt = copy.deepcopy(self.bundle["plan"])
        attempt["limits"].update(max_model_calls=4, max_tool_calls=3)
        digest = self.bundle["digest"]
        plan = dict(
            plan_revision=1, plan_digest=digest, attempt_plan=attempt, source_profile_id="probe_list_filter",
            source_version="list_filter_v1", execution_mode="fixed_response_only", sdk_image_digest=digest,
            go_image_digest=digest, executor_digest=digest, retention_policy_version="retention_v1",
        )
        schema = self.validator("plan_revision")
        schema.validate(plan)
        for field, value in (
            ("write_paths", ["list_test.go"]),
            ("read_paths", ["go.mod", "list.go", "list.go"]),
            ("read_paths", ["go.mod", "list.go", "extra.go"]),
            ("verification_profile", "run_unrestricted_shell"),
        ):
            changed = copy.deepcopy(plan)
            changed["attempt_plan"]["tools"][field] = value
            self.assertFalse(schema.is_valid(changed), field)
        changed = copy.deepcopy(plan)
        del changed["attempt_plan"]["initial_messages_digest"]
        self.assertFalse(schema.is_valid(changed))

    def test_read_shapes_reuse_existing_contracts_and_keep_unknowns(self):
        digest = self.bundle["digest"]
        schema = self.validator("artifact_preview")
        preview = dict(bundle=self.bundle, digest=digest, verification_state=self.bundle["verification_state"], expires_at_ms=1)
        schema.validate(preview)
        self.assertFalse(schema.is_valid(dict(preview, bundle=dict(self.bundle, patch_digest=None))))
        raw_event = json.loads((ROOT / "scripts/eval/fixtures/development-control/cancel-trace.json").read_text())[0]
        event_page = dict(events=[raw_event], next_sequence=raw_event["sequence"], has_more=False, record_version=1)
        self.validator("event_page").validate(event_page)
        self.assertFalse(self.validator("event_page").is_valid(dict(event_page, events=[dict(raw_event, usage_status="guessed")])))
        attempt = dict(
            attempt_id="attempt_fixture", record_version=1, phase="outcome_unknown", termination_status="unconfirmed",
            usage_status="unknown", budget=raw_event["budget"], resource_check_status="pending",
            resource_status=dict(model_connection="unconfirmed", sdk_worker="confirmed", code_environment="stopping"),
            active_tool={"state": "idle"},
        )
        attempt_schema = self.validator("attempt")
        attempt_schema.validate(attempt)
        self.assertFalse(attempt_schema.is_valid(dict(attempt, active_tool=None)))
        self.assertFalse(attempt_schema.is_valid(dict(attempt, resource_status=dict(attempt["resource_status"], pid=42))))

    def test_action_reasons_pages_and_receipt_references_are_explicit(self):
        schema = self.validator("eligibility")
        schema.validate({"allowed": False, "reason": "resource_termination_unconfirmed"})
        self.assertFalse(schema.is_valid({"allowed": True, "reason": "permission_denied"}))
        self.assertFalse(schema.is_valid({"allowed": False, "reason": "allowed"}))
        schema = self.validator("task_page")
        schema.validate(dict(tasks=[], next_cursor="", has_more=False))
        self.assertFalse(schema.is_valid(dict(tasks=[], next_cursor="", has_more=True)))
        self.assertFalse(self.validator("page_query").is_valid(dict(cursor="https://external.invalid", limit=51)))
        schema = self.validator("request_receipt")
        receipt = dict(request_id="save_req", request_digest=self.bundle["digest"], task_id="task_sample")
        schema.validate(dict(receipt, operation="create_task"))
        self.assertFalse(schema.is_valid(dict(receipt, operation="prepare", attempt_id="attempt_sample")))
        self.assertFalse(schema.is_valid(dict(receipt, operation="start")))
        self.assertFalse(schema.is_valid(dict(receipt, operation="save_artifact", attempt_id="attempt_sample")))
        schema.validate(dict(receipt, operation="save_artifact", attempt_id="attempt_sample", artifact_id="devart_fixture"))

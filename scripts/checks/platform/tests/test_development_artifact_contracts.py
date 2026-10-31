"""Development package canonical schema; no SDK, containers or network."""
from __future__ import annotations

import copy
import json
import unittest
from pathlib import Path

from jsonschema import Draft202012Validator, FormatChecker
from referencing import Registry, Resource

ROOT = Path(__file__).resolve().parents[4]


class DevelopmentArtifactContractTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        schemas = [json.loads((ROOT / "contracts" / f"{name}.schema.json").read_text()) for name in (
            "development-artifact", "development-attempt-plan", "development-task-probe-plan", "development-tool-action", "development-attempt-delegation"
        )]
        registry = Registry().with_resources((s["$id"], Resource.from_contents(s)) for s in schemas)
        for schema in schemas:
            Draft202012Validator.check_schema(schema)
        cls.validator = Draft202012Validator(schemas[0], registry=registry, format_checker=FormatChecker())
        cls.bundle = json.loads((ROOT / "scripts/eval/fixtures/development-control/artifact.json").read_text())

    def test_package_record_export_and_retention(self):
        b = self.bundle
        record = dict(schema_version="development_artifact_record.v1", artifact_id="devart_fixture",
                      request_id="save_fixture", saved_at=b["snapshot_at"], bundle=b)
        export = dict(schema_version="development_artifact_export.v1", record=record,
                      exported_at=b["snapshot_at"], request_id="export_fixture", digest=b["digest"])
        retention = dict(schema_version="development_artifact_retention.v1", session_id="sandbox_" + "a" * 32,
                         created_at=b["source_created_at"], expires_at=b["expires_at"])
        for document in [b, record, export, retention]:
            self.validator.validate(document)
            for field in document:
                changed = copy.deepcopy(document)
                del changed[field]
                self.assertFalse(self.validator.is_valid(changed), field)
                changed[field] = None
                self.assertFalse(self.validator.is_valid(changed), field)
            self.assertFalse(self.validator.is_valid(dict(document, download_path="/tmp/arbitrary")))

    def test_scope_files_and_evidence_are_closed(self):
        for container in ["scope", "origin", "plan", "probe"]:
            for field in self.bundle[container]:
                changed = copy.deepcopy(self.bundle)
                del changed[container][field]
                self.assertFalse(self.validator.is_valid(changed), (container, field))
        for container in ["files", "tests"]:
            for field in self.bundle[container][0]:
                changed = copy.deepcopy(self.bundle)
                del changed[container][0][field]
                self.assertFalse(self.validator.is_valid(changed), (container, field))
        for field, value in [("acceptance_state", "accepted"), ("review_state", "approved"),
                             ("attempt_state", "active"), ("source_kind", "arbitrary_directory")]:
            self.assertFalse(self.validator.is_valid(dict(self.bundle, **{field: value})))
        changed = copy.deepcopy(self.bundle)
        changed["files"][0]["path"] = "../private"
        self.assertFalse(self.validator.is_valid(changed))
        changed = copy.deepcopy(self.bundle)
        changed["tests"][0]["argv"] = ["sh", "-c", "go test"]
        self.assertFalse(self.validator.is_valid(changed))
        changed = copy.deepcopy(self.bundle)
        changed["tests"][0]["exit_known"] = False
        changed["tests"][0]["exit_code"] = 1
        self.assertFalse(self.validator.is_valid(changed))


if __name__ == "__main__":
    unittest.main()

"""Behavior checks for the offline probe contract and candidate dependency lock."""
from __future__ import annotations

import copy
import hashlib
import importlib.util
import json
import re
import unittest
from pathlib import Path

from jsonschema import Draft202012Validator

ROOT = Path(__file__).resolve().parents[4]
spec = importlib.util.spec_from_file_location("development_probe_lock", ROOT / "scripts/eval/development_probe/lock_from_report.py")
assert spec and spec.loader
lock_module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(lock_module)


class DevelopmentProbeContractTests(unittest.TestCase):
    def test_frozen_candidate_lock_and_sdk_wire_fixtures(self):
        root = ROOT / "scripts/eval/development_probe"
        lines = (root / "requirements-linux-arm64-py312.lock").read_text().splitlines()
        packages = {}
        for line in lines:
            if line.startswith("#"):
                continue
            match = re.fullmatch(r"([a-z0-9][a-z0-9-]*)==([A-Za-z0-9.!+_-]+) --hash=sha256:([0-9a-f]{64})", line)
            self.assertIsNotNone(match, line)
            name, version, digest = match.groups()
            self.assertNotIn(name, packages)
            packages[name] = (version, digest)
        self.assertEqual(list(packages), sorted(packages))
        self.assertEqual(packages["openhands-sdk"][0], "1.59.0")
        self.assertLessEqual(len(packages), 300)
        for name in ["sdk-first-request.json", "sdk-observation-request.json"]:
            body = json.loads((root / "fixtures" / name).read_text())
            self.assertEqual(body["model"], "gpt-4o")
            self.assertEqual([tool["function"]["name"] for tool in body["tools"]], ["development_probe"])
            self.assertIn("summary", body["tools"][0]["function"]["parameters"]["properties"])
            self.assertNotIn("api_key", body)
        messages = body["messages"]
        self.assertEqual(messages[-1]["tool_call_id"], messages[-2]["tool_calls"][0]["id"])

    def test_frozen_task_and_plan_contract(self):
        fixture = ROOT / "scripts/eval/fixtures/development-task"
        plan = json.loads((fixture / "list-filter-plan.json").read_text())
        schema = json.loads((ROOT / "contracts/development-task-probe-plan.schema.json").read_text())
        Draft202012Validator.check_schema(schema)
        validator = Draft202012Validator(schema)
        validator.validate(plan)
        for source in plan["source_files"]:
            actual = hashlib.sha256((fixture / "list-filter" / source["path"]).read_bytes()).hexdigest()
            self.assertEqual(source["sha256"], "sha256:" + actual)
        for field, value in [("mode", "real_provider"), ("writable_paths", ["list.go", "list_test.go"]),
                             ("verify_argv", ["sh", "-c", "go test ./..."])]:
            changed = copy.deepcopy(plan)
            changed[field] = value
            self.assertFalse(validator.is_valid(changed), field)
        changed = copy.deepcopy(plan)
        changed["sdk_candidate"]["version"] = "latest"
        self.assertFalse(validator.is_valid(changed))

    def test_tool_request_contract_rejects_scope_expansion(self):
        schema = json.loads((ROOT / "contracts/development-task-probe-tool-request.schema.json").read_text())
        Draft202012Validator.check_schema(schema)
        validator = Draft202012Validator(schema)
        request = {"schema_version": "development_task_probe_tool_request.v1", "probe_id": "probe_list_filter",
                   "sequence": 1, "action": "read_file", "path": "list.go", "expected_digest": "",
                   "content": "", "candidate_digest": ""}
        validator.validate(request)
        for field, value in [("path", "../list.go"), ("action", "execute_shell"), ("sequence", 9),
                             ("content", "unrequested"), ("command", "go test")]:
            changed = copy.deepcopy(request)
            changed[field] = value
            self.assertFalse(validator.is_valid(changed), field)
        replacement = dict(request, action="replace_file", expected_digest="sha256:" + "a" * 64,
                           content="package list\n// 合成示例")
        validator.validate(replacement)
        self.assertFalse(validator.is_valid(dict(replacement, content="bad\x00text")))
        validator.validate(dict(request, action="run_tests", path="", candidate_digest="sha256:" + "b" * 64))
        del request["candidate_digest"]
        self.assertFalse(validator.is_valid(request))

    def wheel(self, name="openhands-sdk", version="1.59.0"):
        return {"metadata": {"name": name, "version": version}, "download_info": {
            "url": "https://files.pythonhosted.org/packages/example.whl",
            "archive_info": {"hashes": {"sha256": "a" * 64}}}}

    def test_resolution_lock_requires_fixed_hashed_wheels(self):
        report = {"version": "1", "install": [self.wheel(), self.wheel("jsonschema", "4.23.0")]}
        lock = lock_module.build_lock(report)
        self.assertIn("openhands-sdk==1.59.0 --hash=sha256:" + "a" * 64, lock)
        self.assertEqual(lock, lock_module.build_lock({"version": "1", "install": list(reversed(report["install"]))}))
        invalid = []
        for name, version in [("openhands-sdk", "latest"), ("openhands-sdk", "1.58.2"), ("bad;package", "1.0")]:
            invalid.append({"version": "1", "install": [self.wheel(name, version)]})
        for url in ["http://files.pythonhosted.org/example.whl", "https://other.example/example.whl",
                    "https://files.pythonhosted.org/example.tar.gz", "https://secret@files.pythonhosted.org/example.whl"]:
            changed = copy.deepcopy(report)
            changed["install"][0]["download_info"]["url"] = url
            invalid.append(changed)
        changed = copy.deepcopy(report)
        del changed["install"][0]["download_info"]["archive_info"]
        invalid.append(changed)
        invalid.append({"version": "1", "install": [self.wheel(), self.wheel()]})
        for value in invalid:
            with self.subTest(value=value), self.assertRaises(ValueError):
                lock_module.build_lock(value)


if __name__ == "__main__":
    unittest.main()

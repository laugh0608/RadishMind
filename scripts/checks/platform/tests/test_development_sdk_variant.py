"""Offline source-archive boundaries; never install or import the SDK."""
import importlib.util
import hashlib
import io
import json
from pathlib import Path
import shutil
import tarfile
import tempfile
import unittest
from unittest.mock import patch

ROOT = Path(__file__).resolve().parents[4]
spec = importlib.util.spec_from_file_location("sdk_variant_prepare", ROOT / "scripts/eval/development_probe/prepare_sdk_variant.py")
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)
notice_spec = importlib.util.spec_from_file_location("sdk_notice_delivery", ROOT / "scripts/eval/development_probe/sdk_notice_delivery.py")
notices = importlib.util.module_from_spec(notice_spec)
notice_spec.loader.exec_module(notices)


def notice_fixture(variant):
    (variant / "notice-supplements").mkdir(parents=True)
    body = b"synthetic notice\n"
    (variant / "notice-supplements/package-LICENSE.txt").write_bytes(body)
    (variant / "requirements-linux-arm64-py312.lock").write_text("package==1 --hash=sha256:" + "a" * 64 + "\n")
    evidence = {"schema_version": "development_notice_supplement_evidence.v1", "items": [
        {"name": "package", "version": "1", "wheel_sha256": "a" * 64, "supplements": [
            {"file": "notice-supplements/package-LICENSE.txt", "bytes": len(body),
             "sha256": hashlib.sha256(body).hexdigest(),
             "candidate_image_destination": notices.DESTINATION + "/package-LICENSE.txt"}]}]}
    (variant / "license-delivery-20261105.json").write_text(json.dumps(evidence))
    return evidence


class SDKVariantSourceTests(unittest.TestCase):
    def members(self, entries):
        data = io.BytesIO()
        with tarfile.open(fileobj=data, mode="w") as archive:
            for entry in entries:
                archive.addfile(entry)
        data.seek(0)
        return tarfile.open(fileobj=data)

    def test_archive_rejects_escape_aliases_links_and_duplicate_files(self):
        for name in ("../outside", "/absolute", "other/file", "source/../file", "source/./file", "source//file"):
            with self.subTest(name=name), self.members([tarfile.TarInfo(name)]) as archive, self.assertRaises(ValueError):
                module.archive_members(archive, "source")
        for kind in (tarfile.SYMTYPE, tarfile.LNKTYPE, tarfile.CHRTYPE, tarfile.FIFOTYPE):
            entry = tarfile.TarInfo("source/file")
            entry.type, entry.linkname = kind, "../../outside"
            with self.members([entry]) as archive, self.assertRaises(ValueError):
                module.archive_members(archive, "source")
        with self.members([tarfile.TarInfo("source/file")] * 2) as archive, self.assertRaises(ValueError):
            module.archive_members(archive, "source")
        with self.members([tarfile.TarInfo("source/file")]) as archive:
            self.assertEqual(len(module.archive_members(archive, "source")), 1)

    def test_wrong_source_digest_creates_nothing(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            source = root / "wrong.tar.gz"
            source.write_bytes(b"wrong source")
            with self.assertRaisesRegex(ValueError, "digest mismatch"):
                module.prepare(source, root / "must-not-exist")
            self.assertFalse((root / "must-not-exist").exists())

    def test_preparation_applies_patch_and_records_only_a_new_context(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory).resolve()
            probe = root / "scripts/eval/development_probe"
            variant = probe / "sdk_variant"
            variant.mkdir(parents=True)
            patch_text = "--- a/pyproject.toml\n+++ b/pyproject.toml\n@@ -1 +1 @@\n-before\n+after\n"
            (variant / "optional-laminar.patch").write_text(patch_text)
            archive_path = root / "input.tar.gz"
            with tarfile.open(archive_path, "w:gz") as archive:
                for name, body in {"pyproject.toml": b"before\n", "PKG-INFO": b"old metadata",
                                   "openhands_sdk.egg-info/PKG-INFO": b"old metadata"}.items():
                    item = tarfile.TarInfo("source/" + name)
                    item.size = len(body)
                    archive.addfile(item, io.BytesIO(body))
            sha = lambda data: hashlib.sha256(data).hexdigest()
            manifest = {"archive_root": "source", "source_sha256": module.digest(archive_path),
                        "patch_sha256": sha(patch_text.encode()), "variant_version": "synthetic",
                        "files_before": {"pyproject.toml": sha(b"before\n")},
                        "files_after": {"pyproject.toml": sha(b"after\n")}}
            (variant / "source.json").write_text(json.dumps(manifest))
            for file in [root / "deploy/Dockerfile.development-sdk-variant",
                         root / "contracts/development-task-probe-tool-request.schema.json",
                         variant / "build-requirements.lock"] + [probe / name for name in (
                             "lock_from_report.py", "inspect_sdk.py", "fixed_endpoint.py",
                             "test_sdk_contract.py", "test_sdk_optional_laminar.py")]:
                file.parent.mkdir(parents=True, exist_ok=True)
                file.write_text("synthetic fixture; never executed\n")
            (probe / "requirements-linux-arm64-py312.lock").write_text(
                "# fixture\nopenhands-sdk==1.59.0 --hash=sha256:old\njsonschema==4.23.0 --hash=sha256:test\n")
            shutil.copyfile(ROOT / "scripts/eval/development_probe/sdk_notice_delivery.py", probe / "sdk_notice_delivery.py")
            notice_fixture(variant)
            target = root / "tmp/context"
            with patch.object(module, "ROOT", root), patch.object(module, "PROBE", probe), patch.object(module, "VARIANT", variant):
                module.prepare(archive_path, target)
                with self.assertRaisesRegex(ValueError, "new directory"):
                    module.prepare(archive_path, target)
            self.assertEqual((target / "source/pyproject.toml").read_text(), "after\n")
            self.assertFalse((target / "source/PKG-INFO").exists())
            self.assertFalse((target / "source/openhands_sdk.egg-info").exists())
            self.assertEqual((target / "constraints.txt").read_text(), "jsonschema==4.23.0\n")
            report = json.loads((target / "preparation.json").read_text())
            self.assertEqual(report["files"]["source/pyproject.toml"], sha(b"after\n"))
            self.assertEqual(report["files"]["third-party/package-LICENSE.txt"], sha(b"synthetic notice\n"))


class SDKNoticeDeliveryTests(unittest.TestCase):
    def test_preparation_rejects_incorrect_wheel_notice_and_path(self):
        for failure in ("version", "wheel", "missing", "digest", "path", "destination", "duplicate", "unlisted", "symlink"):
            with self.subTest(failure=failure), tempfile.TemporaryDirectory() as temp:
                root = Path(temp)
                variant = root / "variant"
                evidence = notice_fixture(variant)
                item = evidence["items"][0]
                supplement = item["supplements"][0]
                source = variant / supplement["file"]
                if failure == "version": item["version"] = "2"
                if failure == "wheel": item["wheel_sha256"] = "b" * 64
                if failure == "missing": source.unlink()
                if failure == "digest": source.write_bytes(b"tampered! notice\n")
                if failure == "path": supplement["file"] = "notice-supplements/../package-LICENSE.txt"
                if failure == "destination": supplement["candidate_image_destination"] = "/outside.txt"
                if failure == "duplicate": item["supplements"].append(dict(supplement))
                if failure == "unlisted": (source.parent / "extra.txt").write_text("extra")
                if failure == "symlink":
                    source.unlink()
                    source.symlink_to(root / "outside")
                (variant / "license-delivery-20261105.json").write_text(json.dumps(evidence))
                with self.assertRaises(ValueError): notices.prepare(variant, root / "output")
                self.assertFalse((root / "output").exists())

    def test_delivery_matches_installed_wheels_and_retains_partial_scope(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            variant = root / "variant"
            evidence = notice_fixture(variant)
            directory = root / "delivery"
            notices.prepare(variant, directory)
            manifest = json.loads((directory / "manifest.json").read_text())
            self.assertFalse(manifest["license_gate_closed"])
            self.assertEqual(directory.stat().st_mode & 0o777, 0o755)
            self.assertEqual((directory / "package-LICENSE.txt").stat().st_mode & 0o777, 0o644)
            report = {"version": "1", "install": [{"metadata": {"name": "Package", "version": "1"},
                      "download_info": {"archive_info": {"hashes": {"sha256": "a" * 64}}}}]}
            report_path = root / "installation.json"
            report_path.write_text(json.dumps(report))
            notices.verify(directory, report_path)
            for failure in ("missing", "version", "wheel", "duplicate"):
                modified = json.loads(json.dumps(report))
                if failure == "missing": modified["install"] = []
                if failure == "version": modified["install"][0]["metadata"]["version"] = "2"
                if failure == "wheel": modified["install"][0]["download_info"]["archive_info"]["hashes"]["sha256"] = "b" * 64
                if failure == "duplicate": modified["install"].append(modified["install"][0])
                report_path.write_text(json.dumps(modified))
                with self.subTest(failure=failure), self.assertRaises(ValueError): notices.verify(directory, report_path)
            report_path.write_text(json.dumps(report))
            (directory / "package-LICENSE.txt").write_bytes(b"tampered! notice\n")
            with self.assertRaisesRegex(ValueError, "digest mismatch"): notices.verify(directory, report_path)

    def test_repository_notices_match_frozen_evidence(self):
        with tempfile.TemporaryDirectory() as temp:
            directory = Path(temp) / "delivery"
            notices.prepare(ROOT / "scripts/eval/development_probe/sdk_variant", directory)
            manifest = json.loads((directory / "manifest.json").read_text())
            self.assertEqual(len(manifest["packages"]), 6)
            self.assertEqual(sum(len(p["files"]) for p in manifest["packages"]), 14)


if __name__ == "__main__":
    unittest.main()

"""四项工具元数据核对器的离线正反例；所有响应均为合成样例。"""

import copy
import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

import pillow_tool_metadata as metadata


def fixture(tool="meson", version="1.0.0", tag="py3-none-any"):
    filename = f"{tool}-{version}-{tag}.whl"
    return {"info": {"name": tool, "version": version, "requires_python": ">=3.8",
                     "requires_dist": ['packaging>=20; python_version >= "3.8"',
                                       'pytest; extra == "test"']},
            "urls": [{"filename": filename, "packagetype": "bdist_wheel", "yanked": False,
                      "requires_python": ">=3.8", "size": 100,
                      "url": "https://files.pythonhosted.org/packages/test/" + filename,
                      "digests": {"sha256": metadata.SETUPTOOLS_SHA256}}]}


def validate(document, tool="meson"):
    return metadata.validate_metadata(json.dumps(document).encode(), tool)


class MetadataTests(unittest.TestCase):
    def test_four_tools_and_target_independent_of_host(self):
        for tool in metadata.TOOLS:
            with self.subTest(tool=tool):
                document = fixture(tool, "83.0.0" if tool == "setuptools" else "1.0.0")
                result = validate(document, tool)
                self.assertEqual(result["status"], "metadata_reviewed_not_adopted")
                self.assertEqual(result["requires_dist"], document["info"]["requires_dist"])
                self.assertEqual(len(result["candidate_wheels"]), 1)
                if tool == "ninja":
                    self.assertIn("binary_version_status", result)

    def test_target_tags(self):
        accepted = ["cp312-cp312-manylinux_2_28_aarch64", "cp38-abi3-manylinux2014_aarch64",
                    "py2.py3-none-any", "py3-none-manylinux_2_17_aarch64"]
        rejected = ["cp313-cp313-manylinux_2_28_aarch64", "cp312-cp312-manylinux_2_29_aarch64",
                    "cp312-cp312-manylinux_2_28_x86_64", "py3-none-musllinux_1_2_aarch64",
                    "cp311-cp311-manylinux_2_28_aarch64", "py3-none-win_amd64",
                    "py3-none-manylinux_2_16_aarch64", "cp312-cp312-linux_aarch64"]
        for tag in accepted:
            with self.subTest(tag=tag):
                self.assertTrue(validate(fixture(tag=tag))["candidate_wheels"])
        for tag in rejected:
            with self.subTest(tag=tag), self.assertRaisesRegex(ValueError, "no_compatible"):
                validate(fixture(tag=tag))

    def test_invalid_project_fields(self):
        cases = [("name", "unrelated"), ("version", "not a version"),
                 ("version", "1.0rc1"), ("version", "1.0+local"),
                 ("requires_python", ">=3.13"), ("requires_python", "invalid"),
                 ("requires_python", []), ("requires_dist", "packaging"),
                 ("requires_dist", ["invalid >="]), ("requires_dist", [1])]
        for key, value in cases:
            document = fixture()
            document["info"][key] = value
            with self.subTest(key=key, value=value), self.assertRaises((ValueError, TypeError)):
                validate(document)
        document = fixture()
        del document["info"]["requires_dist"]
        with self.assertRaises(KeyError):
            validate(document)

    def test_nullable_declarations_are_preserved(self):
        document = fixture()
        document["info"]["requires_python"] = None
        document["info"]["requires_dist"] = None
        document["urls"][0]["requires_python"] = None
        result = validate(document)
        self.assertIsNone(result["requires_dist"])
        self.assertIsNone(result["requires_python"])

    def test_invalid_distribution_fields(self):
        cases = [("filename", "other-1.0.0-py3-none-any.whl"), ("size", True), ("size", 0),
                 ("yanked", "false"), ("yanked", True), ("requires_python", ">=3.13"),
                 ("digests", {"sha256": "g" * 64}), ("digests", {"sha256": "a" * 63}),
                 ("url", "https://files.pythonhosted.org.evil/packages/x.whl"),
                 ("url", "http://files.pythonhosted.org/packages/x.whl"),
                 ("url", "https://user@files.pythonhosted.org/packages/x.whl"),
                 ("url", "https://files.pythonhosted.org/packages/wrong.whl"),
                 ("packagetype", "unknown")]
        for key, value in cases:
            document = fixture()
            document["urls"][0][key] = value
            with self.subTest(key=key, value=value), self.assertRaises(ValueError):
                validate(document)
        for suffix in ("?token=x", "#fragment", "\n"):
            document = fixture()
            document["urls"][0]["url"] += suffix
            with self.subTest(suffix=suffix), self.assertRaises(ValueError):
                validate(document)

    def test_lock_minimum_and_candidate_exclusions(self):
        with self.assertRaisesRegex(ValueError, "version_mismatch"):
            validate(fixture("setuptools", "82.0.0"), "setuptools")
        document = fixture("setuptools", "83.0.0")
        document["urls"][0]["digests"]["sha256"] = "a" * 64
        with self.assertRaisesRegex(ValueError, "no_compatible"):
            validate(document, "setuptools")
        with self.assertRaisesRegex(ValueError, "below_known"):
            validate(fixture(version="0.59.9"))
        document = fixture()
        other = fixture(tag="py3-none-win_amd64")["urls"][0]
        document["urls"].append(other)
        self.assertEqual(validate(document)["excluded_wheels"][0]["reason"], "target_tags")
        document["urls"].append(copy.deepcopy(other))
        with self.assertRaisesRegex(ValueError, "duplicate_wheel"):
            validate(document)

    def test_json_fail_closed(self):
        for raw in (b'{"info":{},"info":{}}', b'{"value":NaN}', b'{', b'[]', b'\xff',
                    b' ' * (metadata.MAX_BODY_BYTES + 1)):
            with self.subTest(prefix=raw[:30]), self.assertRaises(ValueError):
                metadata.validate_metadata(raw, "meson")

    def test_cli_with_network_and_process_audit_guard(self):
        script = str(Path(metadata.__file__).resolve())
        # 子进程在导入核对器前禁止网络、派生进程；失败时不能打印成功 JSON。
        wrapper = ("import sys,runpy\n"
                   "def guard(event,args):\n"
                   " if event.startswith(('socket.', 'subprocess.')) or event in "
                   "('os.system','os.posix_spawn','os.fork'): raise RuntimeError(event)\n"
                   "sys.addaudithook(guard)\n"
                   "sys.argv=sys.argv[1:]\nrunpy.run_path(sys.argv[0],run_name='__main__')")
        with tempfile.TemporaryDirectory() as directory:
            source = Path(directory) / "response.json"
            source.write_text(json.dumps(fixture()))
            for args in (["--preflight"], ["--tool", "meson", "--input", str(source)]):
                completed = subprocess.run([sys.executable, "-B", "-c", wrapper, script, *args],
                                           capture_output=True, text=True, timeout=10)
                self.assertEqual(completed.returncode, 0, completed.stderr)
                self.assertIsInstance(json.loads(completed.stdout), dict)
            source.write_text('{}')
            completed = subprocess.run([sys.executable, "-B", "-c", wrapper, script,
                                        "--tool", "meson", "--input", str(source)],
                                       capture_output=True, text=True, timeout=10)
            self.assertEqual(completed.returncode, 1)
            self.assertEqual(completed.stdout, "")
            self.assertIn("metadata_validation_failed", completed.stderr)
            self.assertEqual(source.read_text(), '{}')

    def test_missing_parser_stops_before_validation(self):
        # -I -S 排除 site-packages，模拟解析器缺失，不安装或尝试其它解析器。
        completed = subprocess.run([sys.executable, "-I", "-S", "-B", metadata.__file__,
                                    "--preflight"], capture_output=True, text=True, timeout=10)
        self.assertNotEqual(completed.returncode, 0)
        self.assertEqual(completed.stdout, "")
        self.assertIn("ModuleNotFoundError", completed.stderr)


if __name__ == "__main__":
    unittest.main()

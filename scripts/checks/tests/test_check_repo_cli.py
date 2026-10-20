from __future__ import annotations

import importlib.util
import sys
import unittest
from pathlib import Path
from unittest.mock import patch


SCRIPT = Path(__file__).resolve().parents[2] / "check-repo.py"
SPEC = importlib.util.spec_from_file_location("repo_checker", SCRIPT)
assert SPEC is not None and SPEC.loader is not None
CHECKER = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(CHECKER)


class CheckRepoCLITest(unittest.TestCase):
    def test_go_delegation_is_explicit_for_full_and_fast_entrypoints(self) -> None:
        for fast in (False, True):
            for delegated in (False, True):
                flags = (["--fast"] if fast else []) + (["--skip-platform-go-tests"] if delegated else [])
                with self.subTest(fast=fast, delegated=delegated), \
                        patch.object(sys, "argv", [str(SCRIPT), *flags]), \
                        patch.object(CHECKER, "run_python_script") as text_checks, \
                        patch.object(CHECKER, "check_fast_baseline") as fast_check, \
                        patch.object(CHECKER, "check_full_baseline") as full_check:
                    self.assertEqual(CHECKER.main(), 0)
                    selected, other = (fast_check, full_check) if fast else (full_check, fast_check)
                    selected.assert_called_once_with(skip_platform_go_tests=delegated)
                    other.assert_not_called()
                    self.assertEqual([call.args[0] for call in text_checks.call_args_list], [
                        "check-text-files.py", "check-markdown-size.py", "check-doc-language-policy-v1.py",
                    ])


if __name__ == "__main__":
    unittest.main()

from __future__ import annotations

import importlib.util
import sys
import unittest
from pathlib import Path
from unittest.mock import patch


SCRIPT = Path(__file__).resolve().parents[3] / "check-platform-ops-smoke.py"
SPEC = importlib.util.spec_from_file_location("platform_ops_smoke", SCRIPT)
assert SPEC is not None and SPEC.loader is not None
SMOKE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(SMOKE)


class PlatformOpsSmokeTest(unittest.TestCase):
    def test_go_delegation_keeps_both_bridge_checks_and_defaults_to_full_smoke(self) -> None:
        for flags in ([], ["--skip-go-tests"]):
            with self.subTest(flags=flags), patch.object(sys, "argv", [str(SCRIPT), *flags]), \
                    patch.object(SMOKE, "check_go_platform_tests") as go_tests, \
                    patch.object(SMOKE, "check_bridge_provider_registry") as registry, \
                    patch.object(SMOKE, "check_bridge_inventory") as inventory:
                self.assertEqual(SMOKE.main(), 0)
                self.assertEqual(go_tests.call_count, 0 if flags else 1)
                registry.assert_called_once()
                inventory.assert_called_once_with(registry.call_args.args[0])

    def test_delegated_smoke_still_propagates_bridge_failure(self) -> None:
        with patch.object(sys, "argv", [str(SCRIPT), "--skip-go-tests"]), \
                patch.object(SMOKE, "check_go_platform_tests") as go_tests, \
                patch.object(SMOKE, "check_bridge_provider_registry", side_effect=SystemExit(7)), \
                patch.object(SMOKE, "check_bridge_inventory") as inventory:
            with self.assertRaises(SystemExit) as failure:
                SMOKE.main()
            self.assertEqual(failure.exception.code, 7)
            go_tests.assert_not_called()
            inventory.assert_not_called()

    def test_default_smoke_propagates_go_failure_before_bridge_checks(self) -> None:
        with patch.object(sys, "argv", [str(SCRIPT)]), \
                patch.object(SMOKE, "check_go_platform_tests", side_effect=SystemExit(9)), \
                patch.object(SMOKE, "check_bridge_provider_registry") as registry:
            with self.assertRaises(SystemExit) as failure:
                SMOKE.main()
            self.assertEqual(failure.exception.code, 9)
            registry.assert_not_called()

    def test_go_tests_inherit_cache_configuration(self) -> None:
        with patch.object(SMOKE, "subprocess") as subprocess:
            subprocess.run.return_value.returncode = 0
            SMOKE.check_go_platform_tests()
            call = subprocess.run.call_args
            self.assertEqual(call.args[0], ["go", "test", "./..."])
            self.assertEqual(call.kwargs["cwd"], SMOKE.PLATFORM_DIR)
            self.assertIsNone(call.kwargs["env"])


if __name__ == "__main__":
    unittest.main()

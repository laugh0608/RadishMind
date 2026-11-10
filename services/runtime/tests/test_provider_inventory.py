import os
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

from services.runtime import inference_support


class ProviderInventoryTest(unittest.TestCase):
    def test_workbench_disables_implicit_env_file(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            env_file = Path(directory) / ".env"
            env_file.write_text("RADISHMIND_MODEL_PROFILE=implicit\n", encoding="utf-8")
            with patch.dict(os.environ, {"RADISHMIND_DISABLE_ENV_FILE": "1"}, clear=True):
                inference_support.load_env_file(env_file)
                self.assertNotIn("RADISHMIND_MODEL_PROFILE", os.environ)
            with patch.dict(os.environ, {}, clear=True):
                inference_support.load_env_file(env_file)
                self.assertEqual(os.environ["RADISHMIND_MODEL_PROFILE"], "implicit")

    def test_inventory_distinguishes_active_from_enabled(self) -> None:
        env = {
            "RADISHMIND_MODEL_PROVIDER": "openai-compatible",
            "RADISHMIND_MODEL_PROFILE_FALLBACKS": "primary,backup,missing",
            "RADISHMIND_MODEL_PROFILE_PRIMARY_NAME": "fixture-model",
            "RADISHMIND_MODEL_PROFILE_PRIMARY_BASE_URL": "http://127.0.0.1:7201/primary",
            "RADISHMIND_MODEL_PROFILE_PRIMARY_API_KEY": "primary-key",
            "RADISHMIND_MODEL_PROFILE_BACKUP_NAME": "fixture-model",
            "RADISHMIND_MODEL_PROFILE_BACKUP_BASE_URL": "http://127.0.0.1:7201/backup",
            "RADISHMIND_MODEL_PROFILE_BACKUP_API_KEY": "backup-key",
            "RADISHMIND_MODEL_PROFILE_MISSING_NAME": "fixture-model",
            "RADISHMIND_MODEL_PROFILE_MISSING_BASE_URL": "http://127.0.0.1:7201/missing",
        }
        with patch.dict("os.environ", env, clear=True):
            with patch.object(inference_support, "ENV_FILE_PATH", Path("/missing/provider-inventory.env")):
                profiles = inference_support.describe_provider_inventory()["profiles"]

        observed = {profile["profile"]: profile for profile in profiles}
        self.assertTrue(observed["primary"]["active"])
        self.assertTrue(observed["primary"]["enabled"])
        self.assertFalse(observed["backup"]["active"])
        self.assertTrue(observed["backup"]["fallback"])
        self.assertTrue(observed["backup"]["enabled"])
        self.assertEqual(observed["missing"]["credential_state"], "missing")
        self.assertFalse(observed["missing"]["enabled"])


if __name__ == "__main__":
    unittest.main()

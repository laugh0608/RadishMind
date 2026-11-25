"""Controller inspection tests. No container, SDK or privilege change is needed."""
from copy import deepcopy
from pathlib import Path
import unittest
from unittest.mock import patch

from services.runtime.development_sdk import workbench_launcher as launcher


class WorkbenchSDKLauncherTests(unittest.TestCase):
    def test_inspection_requires_exact_isolated_controller(self):
        data = Path("/private-data")
        cid = "a" * 64
        sdk = "sha256:" + "b" * 64
        original = {
            "Id": cid, "Image": launcher.SDK_CONFIG, "State": {"Running": True},
            "HostConfig": {"NetworkMode": "none", "ReadonlyRootfs": True, "Privileged": False,
                           "PidMode": "", "IpcMode": "private", "CapAdd": ["CAP_KILL", "CAP_SETUID", "CAP_SETGID", "CAP_CHOWN", "CAP_DAC_OVERRIDE"],
                           "CapDrop": ["ALL"], "PidsLimit": 128, "Memory": 1 << 30, "MemorySwap": 1 << 30,
                           "NanoCpus": 2_000_000_000, "Tmpfs": {"/tmp": "rw", "/control": "rw"},
                           "SecurityOpt": ["no-new-privileges:true"]},
            "Mounts": [{"Type": "bind", "Source": str(data), "Destination": str(data), "RW": True}],
        }
        with patch.object(launcher.os, "listdir", return_value=["lo"]):
            launcher.verify_controller(original, cid, sdk, data, "/opt/docker")
            for path, value in [
                (("Id",), "e"*64), (("Image",), "other"), (("State", "Running"), False),
                (("HostConfig", "NetworkMode"), "bridge"), (("HostConfig", "Privileged"), True),
                (("HostConfig", "ReadonlyRootfs"), False), (("HostConfig", "PidMode"), "host"),
                (("HostConfig", "PidsLimit"), -1), (("HostConfig", "Tmpfs"), {"/usr": "rw"}),
                (("HostConfig", "CapAdd"), ["CAP_SYS_ADMIN"]), (("HostConfig", "SecurityOpt"), []),
                (("Mounts", 0, "Source"), "/other"), (("Mounts", 0, "RW"), False),
                (("Mounts",), original["Mounts"] + [{"Type":"bind", "Source":"/", "Destination":"/host", "RW":False}]),
            ]:
                changed = deepcopy(original)
                obj = changed
                for key in path[:-1]:
                    obj = obj[key]
                obj[path[-1]] = value
                with self.subTest(path=path), self.assertRaises(ValueError):
                    launcher.verify_controller(changed, cid, sdk, data, "/opt/docker")
        with patch.object(launcher.os, "listdir", return_value=["lo","eth0"]), self.assertRaises(ValueError):
            launcher.verify_controller(original, cid, sdk, data, "/opt/docker")

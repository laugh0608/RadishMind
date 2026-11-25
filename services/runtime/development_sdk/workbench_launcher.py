#!/usr/local/bin/python -I
"""Run the frozen SDK in an explicitly provisioned, network-isolated ARM64 controller."""
import hashlib
import importlib.metadata
import json
import os
from pathlib import Path
import platform
import re
import socket
import stat
import subprocess
import sys

WORKER = Path("/opt/sdk-worker/worker.py")
SOCKET = "/control/docker.sock"
WORKER_DIGEST = "3334fb5fe439c9bdcf863adc797525e9c4ebac72964d507c44d21661aa64a095"
SDK_CONFIG = "sha256:45d15b588f80a2e44a70906ee917ddafb914667c74e4e398e99cd614a41b40a6"


def require(condition):
    if not condition:
        raise ValueError("unqualified workbench execution environment")


def inspect(cli, kind, identity):
    result = subprocess.run([cli, "--host", "unix://" + SOCKET, kind, "inspect", identity],
                            check=True, capture_output=True, timeout=2)
    items = json.loads(result.stdout)
    require(len(items) == 1)
    return items[0]


def verify_controller(own, container_id, sdk, data, cli):
    host = own["HostConfig"]
    require(own["Id"] == container_id and own["State"]["Running"] and own["Image"] in {sdk, SDK_CONFIG})
    require(host["NetworkMode"] == "none" and set(os.listdir("/sys/class/net")) == {"lo"})
    require(host["ReadonlyRootfs"] and not host["Privileged"] and not host["PidMode"]
            and host["IpcMode"] != "host" and "ALL" in host["CapDrop"])
    caps = {name.removeprefix("CAP_") for name in host.get("CapAdd") or []}
    require(caps == {"CHOWN", "SETUID", "SETGID", "DAC_OVERRIDE", "KILL"})
    require(any(opt.split(":")[0].split("=")[0] == "no-new-privileges" for opt in host["SecurityOpt"]))
    require(0 < host["PidsLimit"] <= 128 and 0 < host["Memory"] <= 1 << 30
            and host["MemorySwap"] == host["Memory"] and 0 < host["NanoCpus"] <= 2_000_000_000)
    require(set(host.get("Tmpfs") or {}) == {"/tmp", "/control"})
    require(all(m["Destination"] in {"/tmp", "/control"} for m in own["Mounts"] if m["Type"] == "tmpfs"))
    mounts = [m for m in own["Mounts"] if m["Type"] != "tmpfs"]
    for m in mounts:
        dest = Path(m["Destination"])
        require(m["Type"] == "bind")
        # The SDK cannot read protected controller inputs. Its only public
        # bind is the hash-checked worker file (or its read-only directory).
        require(dest == Path(SOCKET) or ((dest == data or dest.is_relative_to(data)) and Path(m["Source"]) == dest)
                or (not m["RW"] and dest in {WORKER, WORKER.parent, Path(cli)}))
    # Code snapshots and the pinned runner are Docker bind sources. The host and
    # controller must resolve their common private directory at the same path.
    require(any(m["Destination"] == str(data) and m["Source"] == str(data) and m["RW"] for m in mounts))


def verify(sdk, code, cli, data, config, key):
    require(sys.platform == "linux" and platform.machine() == "aarch64" and os.geteuid() == 0)
    require(data.is_absolute() and data.stat().st_uid == 0 and stat.S_IMODE(data.stat().st_mode) == 0o700)
    require(Path(config).is_file() and (not key or Path(key).is_file()))
    # Docker binds this container's own hostname file. Container IDs chosen
    # by a client or discovered from a different daemon cannot stand in for it.
    roots = [line.split()[3] for line in Path("/proc/self/mountinfo").read_text().splitlines()
             if line.split()[4] == "/etc/hostname"]
    require(len(roots) == 1)
    match = re.search(r"/containers/([0-9a-f]{64})/hostname$", roots[0])
    require(match is not None)
    own = inspect(cli, "container", match[1])
    verify_controller(own, match[1], sdk, data, cli)
    code_image = inspect(cli, "image", code)
    require(code_image["Id"] == code and code_image["Os"] == "linux" and code_image["Architecture"] == "arm64")
    require(hashlib.sha256(WORKER.read_bytes()).hexdigest() == WORKER_DIGEST)
    require(importlib.metadata.version("openhands-sdk") == "1.59.0+radishmind.1")
    result = subprocess.run([cli, "--host", "unix://" + SOCKET, "info", "--format", "{{.ID}}"],
                            check=True, capture_output=True, text=True, timeout=2)
    daemon = result.stdout.strip()
    require(re.fullmatch(r"[a-zA-Z0-9:-]{12,128}", daemon) is not None)
    return daemon


def drop_identity(data, config, key, run):
    if run:
        cwd = Path.cwd()
        require(cwd.stat().st_uid == 0 and stat.S_IMODE(cwd.stat().st_mode) == 0o700 and not list(cwd.iterdir()))
        os.chown(cwd, 65532, 65532)
    os.setgroups([])
    os.setgid(65532)
    os.setuid(65532)
    status = dict(line.split(":", 1) for line in Path("/proc/self/status").read_text().splitlines())
    require(status["NoNewPrivs"].strip() == "1" and all(int(status[k].strip(), 16) == 0
            for k in ("CapInh", "CapPrm", "CapEff", "CapAmb")))
    require(not os.access(data, os.X_OK) and not os.access(config, os.R_OK)
            and (not key or not os.access(key, os.R_OK)))
    with socket.socket(socket.AF_UNIX) as sock:
        try:
            sock.connect(SOCKET)
        except PermissionError:
            pass
        else:
            raise ValueError("Docker socket accessible to SDK")


def main():
    mode, sdk, code, cli, data, config, key = sys.argv[1:]
    require(mode in {"verify", "run"})
    daemon = verify(sdk, code, cli, Path(data), config, key)
    drop_identity(data, config, key, mode == "run")
    if mode == "verify":
        print(daemon)
    else:
        os.execv(sys.executable, [sys.executable, "-I", str(WORKER)])


if __name__ == "__main__":
    main()

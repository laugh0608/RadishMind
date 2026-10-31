"""Authorized Linux joint-test launcher; never a general worker entry point."""

import json
import os
from pathlib import Path
import socket
import stat
import sys


def main():
    if sys.platform != "linux" or os.geteuid() != 0 or len(sys.argv) != 3:
        raise RuntimeError("joint launcher requires the authorized Linux controller")
    worker, private = sys.argv[1:]
    directory = Path.cwd()
    info = directory.stat()
    if info.st_uid != 0 or stat.S_IMODE(info.st_mode) != 0o700 or list(directory.iterdir()):
        raise RuntimeError("worker directory is not new and private")
    os.chown(directory, 65532, 65532)
    os.setgroups([])
    os.setgid(65532)
    os.setuid(65532)
    status = dict(line.split(":", 1) for line in Path("/proc/self/status").read_text().splitlines())
    for name in ("CapInh", "CapPrm", "CapEff", "CapAmb"):
        if int(status[name].strip(), 16) != 0:
            raise RuntimeError("SDK retained capabilities")
    if status["NoNewPrivs"].strip() != "1":
        raise RuntimeError("no-new-privileges missing")
    for target in (str(Path(private) / "joint.db"), "/control/docker.sock"):
        try:
            descriptor = os.open(target, os.O_RDONLY)
            os.close(descriptor)
        except PermissionError:
            pass
        else:
            Path("boundary-failure.json").write_text(json.dumps({"target": target, "uid": os.getuid()}))
            raise RuntimeError("SDK can access controller data")
    with socket.socket(socket.AF_UNIX) as connection:
        try:
            connection.connect("/control/docker.sock")
        except PermissionError:
            pass
        else:
            raise RuntimeError("SDK can connect to Docker")
    Path("boundaries.json").write_text(json.dumps({
        "uid": os.getuid(), "gid": os.getgid(), "groups": os.getgroups(),
        "capabilities_zero": True, "no_new_privileges": True,
        "docker_denied": True, "controller_data_denied": True,
    }) + "\n")
    os.execv(sys.executable, [sys.executable, worker])


if __name__ == "__main__":
    main()

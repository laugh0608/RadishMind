"""Run explicitly authorized joint acceptance with prepared, pinned local inputs.

Does not download, build, pull images or start Docker itself. Each invocation
retains logs/inspection and removes only its exact controller container.
"""
import argparse
import hashlib
import json
import re
from pathlib import Path
import subprocess

SDK_IMAGE = "sha256:d80a7496e7c659f1fc76ffb5f429530bcfcd1fef5a790fa06b789ff27eb23842"
CODE_IMAGE = "sha256:b17af760035fc2f338eed92d448a6c67f2d45438844fc6c60678fa5f99e44b57"


def run(args, **kwargs):
    kwargs.setdefault("timeout", 15)
    return subprocess.run(args, check=True, **kwargs)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--directory", type=Path, required=True)
    parser.add_argument("--socket", type=Path, required=True)
    parser.add_argument("--run-id", required=True, type=int)
    parser.add_argument("--sdk-image", default=SDK_IMAGE, help="exact local SDK image sha256; never a tag")
    args = parser.parse_args()
    if not re.fullmatch(r"sha256:[0-9a-f]{64}", args.sdk_image):
        parser.error("sdk-image must be an immutable local image digest")
    root = args.directory.resolve(strict=True)
    repo = Path(__file__).resolve().parents[3]
    if not root.is_relative_to(repo / "tmp") or args.run_id < 1:
        parser.error("directory must be inside repository tmp; run-id must be positive")
    root.chmod(0o700)
    name = f"radishmind-joint-20261031-{args.run_id:02d}"
    prefix = root / f"run-{args.run_id:02d}"
    log = prefix.with_suffix(".log")
    if prefix.with_suffix(".command.json").exists():
        parser.error("run-id already has evidence")
    files = {
        "test": root / "joint.test", "runner": root / "code-runner",
        "cli": root / "tools/docker/docker",
        "launcher": repo / "scripts/eval/development_probe/joint_worker.py",
        "worker": repo / "services/runtime/development_sdk/worker.py",
    }
    hashes = {name: hashlib.sha256(path.read_bytes()).hexdigest() for name, path in files.items()}
    cmd = ["docker", "container", "create", "--pull=never", "--name", name,
           "--network", "none", "--read-only", "--user", "0:0", "--cap-drop", "ALL"]
    for cap in ("CHOWN", "SETUID", "SETGID", "DAC_OVERRIDE", "KILL"):
        cmd += ["--cap-add", cap]
    cmd += ["--security-opt", "no-new-privileges", "--memory", "1g", "--memory-swap", "1g",
            "--cpus", "2", "--pids-limit", "128", "--restart", "no", "--log-driver", "none",
            "--tmpfs", "/tmp:rw,exec,size=512m,mode=1777,nosuid,nodev",
            "--tmpfs", "/control:rw,size=1m,mode=0700,nosuid,nodev,noexec",
            "--tmpfs", f"{root.parent}:rw,size=1m,mode=0700,nosuid,nodev,noexec"]
    mounts = [(root, str(root), False), (files["test"], "/opt/joint-test", True),
              (files["cli"], "/opt/docker", True), (files["launcher"], "/opt/joint-worker.py", True),
              (files["worker"].parent, "/opt/sdk-worker", True),
              (args.socket.resolve(strict=True), "/control/docker.sock", False)]
    for source, target, readonly in mounts:
        cmd += ["--mount", f"type=bind,src={source},dst={target}" + (",readonly" if readonly else "")]
    cmd += ["-e", f"RADISHMIND_JOINT_ROOT={root}", "-e", f"RADISHMIND_SANDBOX_IMAGE={CODE_IMAGE}",
            "-e", "DOCKER_HOST=unix:///control/docker.sock", args.sdk_image,
            "/opt/joint-test", "-test.run", "^TestJointSDKCandidateDelivery$", "-test.v", "-test.timeout", "480s"]
    prefix.with_suffix(".command.json").write_text(json.dumps({"command": cmd, "sha256": hashes}, indent=2) + "\n")
    cases_before = set(root.glob("case-*"))
    created = False
    try:
        run(cmd, capture_output=True, text=True)
        created = True
        # Inspect before executing: audit the actual controller rather than flags alone.
        inspected = run(["docker", "inspect", name], capture_output=True, text=True).stdout
        prefix.with_suffix(".container.json").write_text(inspected)
        item = json.loads(inspected)[0]
        host = item["HostConfig"]
        actual_mounts = {(m["Source"], m["Destination"], m["RW"]) for m in item["Mounts"] if m["Type"] == "bind"}
        expected_mounts = {(str(source), target, not readonly) for source, target, readonly in mounts}
        if (item["Image"] != args.sdk_image or item["Config"]["User"] != "0:0"
                or not host["ReadonlyRootfs"] or host["NetworkMode"] != "none"
                or host["Privileged"] or host["PortBindings"]
                or host["Memory"] != 1 << 30 or host["MemorySwap"] != 1 << 30
                or host["NanoCpus"] != 2_000_000_000 or host["PidsLimit"] != 128
                or host["CapDrop"] != ["ALL"]
                or set(host["CapAdd"]) != {"CAP_CHOWN", "CAP_SETUID", "CAP_SETGID", "CAP_DAC_OVERRIDE", "CAP_KILL"}
                or "no-new-privileges" not in host["SecurityOpt"]
                or actual_mounts != expected_mounts):
            raise RuntimeError("controller configuration differs from authorized topology")
        with log.open("x") as output:
            result = subprocess.run(["docker", "start", "--attach", name], stdout=output, stderr=subprocess.STDOUT, timeout=510)
        state = json.loads(run(["docker", "inspect", name], capture_output=True, text=True).stdout)[0]["State"]
        prefix.with_suffix(".state.json").write_text(json.dumps(state, indent=2) + "\n")
        print(log.read_text())
        if result.returncode or state["ExitCode"] or state["OOMKilled"]:
            raise RuntimeError("joint acceptance failed; retained evidence")
    finally:
        if created:
            try:
                run(["docker", "rm", "--force", name], capture_output=True)
            finally:
                # Inspect only new ledgers from this invocation; no global prune.
                audit = []
                for case in sorted(set(root.glob("case-*")) - cases_before):
                    ledger_path = case / "sandbox/resources.json"
                    if not ledger_path.exists():
                        continue
                    ledger = json.loads(ledger_path.read_text())
                    owner = ledger["session_id"]
                    if ledger["schema_version"] != "development_sandbox_resources.v1" or ledger["image"] != CODE_IMAGE:
                        raise RuntimeError("invalid cleanup ledger")
                    query = ["docker", "ps", "-a", "--no-trunc", "--filter",
                             f"label=local.radishmind.development.sandbox={owner}", "--format", "{{.ID}}"]
                    ids = run(query, capture_output=True, text=True).stdout.split()
                    for identifier in ids:
                        item = json.loads(run(["docker", "inspect", identifier], capture_output=True, text=True).stdout)[0]
                        matching = [r for r in ledger["resources"] if item["Name"] == "/" + r["name"] and r["container_id"] in ("", identifier)]
                        labels = item["Config"]["Labels"]
                        if (len(matching) != 1 or item["Image"] != CODE_IMAGE
                                or labels.get("local.radishmind.development.sandbox") != owner
                                or labels.get("local.radishmind.development.attempt") != ledger["binding"]["attempt_id"]):
                            raise RuntimeError("cleanup ownership mismatch")
                        run(["docker", "rm", "--force", identifier], capture_output=True)
                    if run(query, capture_output=True, text=True).stdout.strip():
                        raise RuntimeError("owned code containers remain")
                    audit.append({"ledger": str(ledger_path.relative_to(root)), "removed_after_controller": ids, "remaining": []})
                prefix.with_suffix(".cleanup.json").write_text(json.dumps(audit, indent=2) + "\n")
                if any(entry["removed_after_controller"] for entry in audit):
                    raise RuntimeError("controller left code resources; host reconciled them, acceptance failed")


if __name__ == "__main__":
    main()

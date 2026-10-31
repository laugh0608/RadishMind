"""Prepare a new repository-tmp build context from a locally supplied pinned sdist.

No downloads, package installation, SDK imports, builds or containers. Failures
retain the private incomplete directory for inspection; existing paths are refused.
"""
from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path, PurePosixPath
import shutil
import subprocess
import tarfile

ROOT = Path(__file__).resolve().parents[3]
PROBE = ROOT / "scripts/eval/development_probe"
VARIANT = PROBE / "sdk_variant"


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def archive_members(archive: tarfile.TarFile, expected_root: str):
    members = archive.getmembers()
    if len(members) > 2000 or sum(item.size for item in members) > 16 * 1024 * 1024:
        raise ValueError("source archive exceeds bounds")
    seen = set()
    for item in members:
        path = PurePosixPath(item.name)
        if (not path.parts or path.is_absolute() or path.parts[0] != expected_root
                or item.name.rstrip("/") != path.as_posix()
                or ".." in path.parts or "\\" in item.name or item.name in seen
                or not (item.isfile() or item.isdir()) or item.size > 4 * 1024 * 1024):
            raise ValueError("unsafe or duplicate source archive member")
        seen.add(item.name)
    return members


def prepare(source_archive: Path, directory: Path) -> None:
    manifest = json.loads((VARIANT / "source.json").read_text())
    patch = VARIANT / "optional-laminar.patch"
    if source_archive.stat().st_size > 8 * 1024 * 1024 or digest(source_archive) != manifest["source_sha256"]:
        raise ValueError("source archive digest mismatch")
    if digest(patch) != manifest["patch_sha256"]:
        raise ValueError("variant patch digest mismatch")
    directory = directory.resolve()
    if not directory.is_relative_to(ROOT / "tmp") or directory == ROOT / "tmp" or directory.exists():
        raise ValueError("a new directory inside repository tmp is required")
    directory.parent.mkdir(parents=True, exist_ok=True)
    with tarfile.open(source_archive, "r:gz") as archive:
        members = archive_members(archive, manifest["archive_root"])
        directory.mkdir(mode=0o700)
        source = directory / "source"
        source.mkdir(mode=0o700)
        for item in members:
            target = source.joinpath(*PurePosixPath(item.name).parts[1:])
            if item.isdir():
                target.mkdir(parents=True, exist_ok=True, mode=0o700)
            else:
                target.parent.mkdir(parents=True, exist_ok=True, mode=0o700)
                with archive.extractfile(item) as incoming, target.open("xb") as outgoing:
                    shutil.copyfileobj(incoming, outgoing)
                target.chmod(0o600)
    for name, expected in manifest["files_before"].items():
        if digest(source / name) != expected:
            raise ValueError(f"upstream file differs: {name}")
    prefix = str(source.relative_to(ROOT))
    # Exact pre/post hashes guard zero-context hunks; the patch itself contains no
    # whitespace-only context lines that would fail repository text hygiene.
    command = ["git", "apply", "--unidiff-zero", "--whitespace=error-all", f"--directory={prefix}", str(patch)]
    subprocess.run(command[:2] + ["--check"] + command[2:], cwd=ROOT, check=True, timeout=15)
    subprocess.run(command, cwd=ROOT, check=True, timeout=15)
    for name, expected in manifest["files_after"].items():
        if digest(source / name) != expected:
            raise ValueError(f"patched file differs: {name}")
    # Regenerate distribution metadata from the patched project, not the sdist's
    # old Requires-Dist / SOURCES records. These are inside this new context only.
    (source / "PKG-INFO").unlink()
    shutil.rmtree(source / "openhands_sdk.egg-info")
    copies = {
        ROOT / "deploy/Dockerfile.development-sdk-variant": "Dockerfile",
        VARIANT / "source.json": "source.json",
        VARIANT / "build-requirements.lock": "build-requirements.lock",
        PROBE / "lock_from_report.py": "lock_from_report.py",
        PROBE / "inspect_sdk.py": "inspect_sdk.py",
        PROBE / "fixed_endpoint.py": "fixed_endpoint.py",
        PROBE / "test_sdk_contract.py": "test_sdk_contract.py",
        PROBE / "test_sdk_optional_laminar.py": "test_sdk_optional_laminar.py",
        ROOT / "contracts/development-task-probe-tool-request.schema.json": "tool-request.schema.json",
    }
    for original, name in copies.items():
        shutil.copyfile(original, directory / name)
    # Constraints retain existing dependency versions without installing unused
    # distributions. The new dependency closure is produced by pip's resolver.
    stock = (PROBE / "requirements-linux-arm64-py312.lock").read_text().splitlines()
    constraints = [line.split(" --hash=")[0] for line in stock
                   if line and not line.startswith(("#", "openhands-sdk=="))]
    (directory / "constraints.txt").write_text("\n".join(constraints) + "\n")
    context_files = sorted(path for path in directory.rglob("*") if path.is_file())
    report = {"schema_version": "development_sdk_variant_preparation.v1",
              "source_sha256": digest(source_archive), "patch_sha256": digest(patch),
              "variant_version": manifest["variant_version"],
              "files": {str(path.relative_to(directory)): digest(path) for path in context_files}}
    (directory / "preparation.json").write_text(json.dumps(report, indent=2) + "\n")


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--source-archive", type=Path, required=True)
    parser.add_argument("--directory", type=Path, required=True)
    args = parser.parse_args()
    prepare(args.source_archive, args.directory)


if __name__ == "__main__":
    main()

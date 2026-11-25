"""Prepare and verify partial SDK notices; never download, install or approve licenses."""
from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path, PurePosixPath
import re
from urllib.parse import urlsplit


DESTINATION = "/usr/local/share/radishmind/third-party"
CPYTHON_EXTENSION_DIRECTORY = Path("/usr/local/lib/python3.12/lib-dynload")
# CPython 3.12.16 notices and extension fingerprints from the retained SDK image evidence.
CPYTHON_NOTICES = {
    "Expat-COPYING.txt": (1144, "31b15de82aa19a845156169a17a5488bf597e561b2c318d159ed583139b25e87"),
    "libmpdec-LICENSE.txt": (1359, "b07528d8b1dbf1e2d2741052996f0876e23342ce2d30d0effa39c5457716c25a"),
}
CPYTHON_EXTENSIONS = {
    "pyexpat.cpython-312-aarch64-linux-gnu.so": (
        271152, "074bf13e0f10b9203ee1191dff15eecec584fff8a2df7ec1c55083a1b00ed0cf"),
    "_decimal.cpython-312-aarch64-linux-gnu.so": (
        339552, "60a6a97503e9ea31efb35bf2aabfffd7fb2b0bfc1c8fb9a08757206622bae008"),
}


def normalized_name(name: str) -> str:
    return re.sub(r"[-_.]+", "-", name).lower()


def frozen_packages(lock: str) -> dict[str, tuple[str, str]]:
    packages = {}
    for line in lock.splitlines():
        if not line or line.startswith("#"):
            continue
        entry = re.fullmatch(r"([a-z0-9]+(?:-[a-z0-9]+)*)==([A-Za-z0-9.+!_-]+) --hash=sha256:([0-9a-f]{64})", line)
        if entry is None or entry[1] in packages:
            raise ValueError(f"invalid or duplicate frozen requirement: {line}")
        name, version, digest = entry.groups()
        packages[name] = (version, digest)
    if not packages:
        raise ValueError("empty frozen wheel lock")
    return packages


def checked_bytes(root: Path, name: str, size: int, digest: str) -> bytes:
    path = root / name
    if root.is_symlink() or path.is_symlink() or not path.is_file():
        raise ValueError(f"{name} must be a regular file")
    if not 0 < size <= 1024 * 1024 or path.stat().st_size != size:
        raise ValueError(f"{name} size mismatch")
    body = path.read_bytes()
    if hashlib.sha256(body).hexdigest() != digest:
        raise ValueError(f"{name} digest mismatch")
    return body


def notice_bytes(root: Path, name: str, size: int, digest: str) -> bytes:
    if not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9._-]*\.txt", name):
        raise ValueError("invalid notice filename")
    return checked_bytes(root, name, size, digest)


def cpython_notices(directory: Path) -> dict[str, bytes]:
    bodies = {name: notice_bytes(directory, name, *expected) for name, expected in CPYTHON_NOTICES.items()}
    if bodies.keys() != {path.name for path in directory.iterdir()}:
        raise ValueError("CPython notice inventory mismatch")
    return bodies


def verify_cpython(directory: Path, extension_directory: Path) -> None:
    cpython_notices(directory / "cpython")
    for name, expected in CPYTHON_EXTENSIONS.items():
        checked_bytes(extension_directory, name, *expected)


def pillow_supplements(variant: Path, wheel_lock: str) -> tuple[dict, dict[str, bytes], str]:
    directory = variant / "pillow-notice-supplements"
    manifest_path = directory / "manifest.json"
    manifest = json.loads(manifest_path.read_text())
    evidence = (variant / "pillow-notice-sources-20261121.json").read_bytes()
    if (manifest["schema_version"] != "development_sdk_notice_delivery.v1"
            or manifest["scope"] != "partial_supplements_only"
            or manifest["license_gate_closed"] is not False
            or hashlib.sha256(evidence).hexdigest() != manifest["evidence_sha256"]):
        raise ValueError("invalid Pillow notice evidence")
    package, = manifest["packages"]
    wheel = (package["version"], package["wheel_sha256"])
    if package["name"] != "pillow" or frozen_packages(wheel_lock).get("pillow") != wheel:
        raise ValueError("Pillow notices do not match the delivery lock")
    files = {}
    for item in package["files"]:
        name = item["file"]
        if name in files:
            raise ValueError("duplicate Pillow notice")
        files[name] = notice_bytes(directory, name, item["bytes"], item["sha256"])
    if len(files) != 5 or set(files) != {path.name for path in directory.glob("*.txt")}:
        raise ValueError("Pillow notice inventory mismatch")
    return package, files, manifest["evidence_sha256"]


def prepare(variant: Path, directory: Path, wheel_lock: str | None = None) -> None:
    evidence_path = variant / "license-delivery-20261105.json"
    evidence = json.loads(evidence_path.read_text())
    if evidence["schema_version"] != "development_notice_supplement_evidence.v1":
        raise ValueError("unsupported notice evidence")
    lock = (wheel_lock if wheel_lock is not None
            else (variant / "requirements-linux-arm64-py312.lock").read_text()).splitlines()
    packages, files, names = [], {}, set()
    for item in evidence["items"]:
        name = normalized_name(item["name"])
        expected = f"{name}=={item['version']} --hash=sha256:{item['wheel_sha256']}"
        if name in names or lock.count(expected) != 1:
            raise ValueError("notice package does not match frozen wheel lock")
        names.add(name)
        package = {key: item[key] for key in ("name", "version", "wheel_sha256")}
        package["files"] = []
        for supplement in item["supplements"]:
            path = PurePosixPath(supplement["file"])
            if (len(path.parts) != 2 or path.parts[0] != "notice-supplements"
                    or path.as_posix() != supplement["file"] or path.name in files
                    or supplement["candidate_image_destination"] != f"{DESTINATION}/{path.name}"):
                raise ValueError("invalid or duplicate notice destination")
            files[path.name] = notice_bytes(variant / "notice-supplements", path.name,
                                           supplement["bytes"], supplement["sha256"])
            package["files"].append({"file": path.name, "bytes": supplement["bytes"],
                                     "sha256": supplement["sha256"]})
        if not package["files"]:
            raise ValueError("package has no supplemental notices")
        packages.append(package)
    if not packages or set(files) != {p.name for p in (variant / "notice-supplements").glob("*.txt")}:
        raise ValueError("notice inventory is empty or incomplete")
    evidence_bindings = {"evidence_sha256": hashlib.sha256(evidence_path.read_bytes()).hexdigest()}
    if wheel_lock is not None:
        package, bodies, pillow_evidence = pillow_supplements(variant, wheel_lock)
        if package["name"] in names or files.keys() & bodies.keys():
            raise ValueError("duplicate Pillow notice package or destination")
        packages.append(package)
        files.update(bodies)
        evidence_bindings["pillow_evidence_sha256"] = pillow_evidence
    cpython = cpython_notices(variant / "cpython-notice-supplements")
    directory.mkdir(mode=0o700, exist_ok=False)
    cpython_directory = directory / "cpython"
    cpython_directory.mkdir(mode=0o700)
    for destination, bodies in ((directory, files), (cpython_directory, cpython)):
        for name, body in bodies.items():
            target = destination / name
            target.write_bytes(body)
            target.chmod(0o644)
    cpython_directory.chmod(0o755)
    manifest = {"schema_version": "development_sdk_notice_delivery.v1", "scope": "partial_supplements_only",
                "license_gate_closed": False, **evidence_bindings,
                "packages": packages}
    (directory / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
    (directory / "manifest.json").chmod(0o644)
    directory.chmod(0o755)


def verify(directory: Path, installation_report: Path, wheel_lock: Path | None = None) -> None:
    manifest = json.loads((directory / "manifest.json").read_text())
    report = json.loads(installation_report.read_text())
    if (manifest["schema_version"] != "development_sdk_notice_delivery.v1"
            or manifest["scope"] != "partial_supplements_only" or manifest["license_gate_closed"] is not False
            or report.get("version") != "1"):
        raise ValueError("unsupported notice manifest or pip report")
    installed = {}
    for item in report["install"]:
        name = normalized_name(item["metadata"]["name"])
        if name in installed:
            raise ValueError("duplicate installed package")
        download = item["download_info"]
        if wheel_lock is not None:
            location = urlsplit(download["url"])
            if location.scheme != "file" or location.netloc or not location.path.endswith(".whl"):
                raise ValueError("frozen installation must use local wheels")
        installed[name] = (item["metadata"]["version"],
                           download["archive_info"]["hashes"]["sha256"])
    if wheel_lock is not None and installed != frozen_packages(wheel_lock.read_text()):
        raise ValueError("installation report does not match the full frozen wheel lock")
    names, files = set(), set()
    for package in manifest["packages"]:
        name = normalized_name(package["name"])
        if name in names or installed.get(name) != (package["version"], package["wheel_sha256"]):
            raise ValueError("installed wheel does not match notice evidence")
        names.add(name)
        if not package["files"]:
            raise ValueError("package has no supplemental notices")
        for item in package["files"]:
            if item["file"] in files:
                raise ValueError("duplicate notice file")
            files.add(item["file"])
            notice_bytes(directory, item["file"], item["bytes"], item["sha256"])
    if not names or files != {p.name for p in directory.glob("*.txt")}:
        raise ValueError("notice inventory is empty or incomplete")


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--directory", type=Path, required=True)
    mode = parser.add_mutually_exclusive_group(required=True)
    mode.add_argument("--variant", type=Path)
    mode.add_argument("--installation-report", type=Path)
    parser.add_argument("--frozen-wheel-lock", type=Path)
    args = parser.parse_args()
    if args.variant is not None:
        prepare(args.variant, args.directory, args.frozen_wheel_lock.read_text() if args.frozen_wheel_lock else None)
    else:
        verify(args.directory, args.installation_report, args.frozen_wheel_lock)
        verify_cpython(args.directory, CPYTHON_EXTENSION_DIRECTORY)


if __name__ == "__main__":
    main()

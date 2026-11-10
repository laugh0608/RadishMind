"""Prepare pip vendor notices and verify their exact installed source tree offline."""
from __future__ import annotations

import argparse
from email.parser import Parser
import hashlib
import json
from pathlib import Path, PurePosixPath
import re


def regular_bytes(root: Path, relative: str) -> bytes:
    path = PurePosixPath(relative)
    if path.is_absolute() or not path.parts or ".." in path.parts or path.as_posix() != relative:
        raise ValueError("invalid relative file path")
    current = root
    for part in ("", *path.parts):
        current = current / part
        if current.is_symlink():
            raise ValueError("symlink is not a source file")
    if not current.is_file() or current.stat().st_size > 4 * 1024 * 1024:
        raise ValueError("missing or oversized source file")
    return current.read_bytes()


def check_digest(body: bytes, item: dict) -> None:
    if len(body) != item["bytes"] or hashlib.sha256(body).hexdigest() != item["sha256"]:
        raise ValueError("source size or digest mismatch")


def installed_files(items: list[dict]) -> dict[str, dict]:
    expected = {}
    for item in items:
        path = PurePosixPath(item["path"])
        if (len(path.parts) < 2 or path.parts[0] != "_vendor" or ".." in path.parts
                or path.as_posix() != item["path"] or item["path"] in expected
                or type(item["bytes"]) is not int or not 0 <= item["bytes"] <= 4 * 1024 * 1024
                or not re.fullmatch(r"[0-9a-f]{64}", item["sha256"])):
            raise ValueError("invalid or duplicate pip source record")
        expected[item["path"]] = item
    if not expected:
        raise ValueError("empty pip source inventory")
    return expected


def notice_name(name: str) -> bool:
    return re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9._-]*\.txt", name) is not None


def prepare(evidence_path: Path, directory: Path) -> None:
    evidence_bytes = regular_bytes(evidence_path.parent, evidence_path.name)
    evidence = json.loads(evidence_bytes)
    if (evidence["schema_version"] != "development_pip_notice_evidence.v1"
            or evidence["name"] != "pip" or not re.fullmatch(r"\d+\.\d+\.\d+", evidence["version"])):
        raise ValueError("unsupported pip evidence")
    files, seen_parts = [], set()
    for part in evidence["installed_files_parts"]:
        if (not re.fullmatch(r"pip-notices-20261105\.parts/vendor-\d{2}\.json", part["file"])
                or part["file"] in seen_parts):
            raise ValueError("invalid or duplicate source part")
        seen_parts.add(part["file"])
        body = regular_bytes(evidence_path.parent, part["file"])
        if hashlib.sha256(body).hexdigest() != part["sha256"]:
            raise ValueError("source part digest mismatch")
        files.extend(json.loads(body)["files"])
    installed_files(files)
    notices, bodies = [], {}
    for item in evidence["notices"]:
        path = PurePosixPath(item["file"])
        if (len(path.parts) != 2 or path.parts[0] != "pip-notice-supplements"
                or not notice_name(path.name) or path.name in bodies
                or item["candidate_image_destination"] != "/usr/local/share/radishmind/third-party/pip/" + path.name):
            raise ValueError("invalid or duplicate pip notice destination")
        body = regular_bytes(evidence_path.parent, item["file"])
        check_digest(body, item)
        bodies[path.name] = body
        notices.append({"file": path.name, "bytes": len(body), "sha256": item["sha256"]})
    if not notices or set(bodies) != {p.name for p in (evidence_path.parent / "pip-notice-supplements").glob("*.txt")}:
        raise ValueError("pip notice inventory is empty or incomplete")
    manifest = {"schema_version": "development_pip_notice_delivery.v1", "scope": "pip_vendor_notices_only",
                "license_gate_closed": False, "version": evidence["version"],
                "evidence_sha256": hashlib.sha256(evidence_bytes).hexdigest(),
                "installed_files": files, "notices": notices}
    directory.mkdir(mode=0o700, exist_ok=False)
    for name, body in bodies.items():
        (directory / name).write_bytes(body)
        (directory / name).chmod(0o644)
    (directory / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
    (directory / "manifest.json").chmod(0o644)
    directory.chmod(0o755)


def verify(directory: Path, pip_root: Path, metadata_path: Path) -> None:
    manifest = json.loads(regular_bytes(directory, "manifest.json"))
    if (manifest["schema_version"] != "development_pip_notice_delivery.v1"
            or manifest["scope"] != "pip_vendor_notices_only" or manifest["license_gate_closed"] is not False):
        raise ValueError("unsupported pip delivery manifest")
    # Read distribution metadata without importing pip or any vendor code.
    if (pip_root.name != "pip" or metadata_path.parent.parent != pip_root.parent
            or metadata_path.parent.name != f"pip-{manifest['version']}.dist-info"
            or metadata_path.name != "METADATA"
            or sorted(p.name for p in pip_root.parent.glob("pip-*.dist-info")) != [metadata_path.parent.name]):
        raise ValueError("pip metadata is not in the same installation")
    metadata = Parser().parsestr(regular_bytes(pip_root.parent, str(metadata_path.relative_to(pip_root.parent))).decode("utf-8"))
    if metadata.get_all("Name") != ["pip"] or metadata.get_all("Version") != [manifest["version"]]:
        raise ValueError("installed pip version mismatch")
    expected = installed_files(manifest["installed_files"])
    for path, record in expected.items():
        check_digest(regular_bytes(pip_root, path), record)
    actual = set()
    for path in (pip_root / "_vendor").rglob("*"):
        if path.is_symlink() or not (path.is_dir() or path.is_file()):
            raise ValueError("unexpected pip source entry")
        if path.is_file():
            actual.add(path.relative_to(pip_root).as_posix())
    if actual != set(expected):
        raise ValueError("pip source inventory mismatch")
    names = set()
    for item in manifest["notices"]:
        if not notice_name(item["file"]) or item["file"] in names:
            raise ValueError("invalid or duplicate pip notice")
        names.add(item["file"])
        check_digest(regular_bytes(directory, item["file"]), item)
    if not names or names != {p.name for p in directory.glob("*.txt")}:
        raise ValueError("pip notice inventory is empty or incomplete")


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--directory", type=Path, required=True)
    mode = parser.add_mutually_exclusive_group(required=True)
    mode.add_argument("--evidence", type=Path)
    mode.add_argument("--pip-root", type=Path)
    parser.add_argument("--metadata", type=Path)
    args = parser.parse_args()
    if args.evidence is not None:
        if args.metadata is not None:
            parser.error("metadata is only valid for verification")
        prepare(args.evidence, args.directory)
    else:
        if args.metadata is None:
            parser.error("verification requires --metadata")
        verify(args.directory, args.pip_root, args.metadata)


if __name__ == "__main__":
    main()

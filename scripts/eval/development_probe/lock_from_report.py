"""Freeze wheel hashes from pip's isolated candidate resolution report.

This utility does not resolve, download or install packages itself.
"""
from __future__ import annotations

import argparse
import hashlib
import json
import re
from pathlib import Path
from urllib.parse import unquote, urlsplit


VARIANT_VERSION = "1.59.0+radishmind.1"
VARIANT_WHEEL = "openhands_sdk-1.59.0+radishmind.1-py3-none-any.whl"


def build_lock(report: dict, sdk_wheel: Path | None = None) -> str:
    if report.get("version") != "1" or not isinstance(report.get("install"), list):
        raise ValueError("unsupported pip resolution report")
    packages: dict[str, tuple[str, str]] = {}
    local_digest = ""
    if sdk_wheel is not None:
        if sdk_wheel.name != VARIANT_WHEEL or sdk_wheel.is_symlink() or not sdk_wheel.is_file():
            raise ValueError("exact locally built SDK variant wheel required")
        if sdk_wheel.stat().st_size > 8 * 1024 * 1024:
            raise ValueError("SDK wheel exceeds validation scope")
        local_digest = hashlib.sha256(sdk_wheel.read_bytes()).hexdigest()
    for item in report["install"]:
        metadata = item.get("metadata", {})
        name, version = metadata.get("name", ""), metadata.get("version", "")
        if not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9._-]*", name) or not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9.!+_-]*", version):
            raise ValueError("invalid distribution identity")
        name = re.sub(r"[-_.]+", "-", name).lower()
        download = item.get("download_info", {})
        url = urlsplit(download.get("url", ""))
        digest = download.get("archive_info", {}).get("hashes", {}).get("sha256", "")
        local_sdk = sdk_wheel is not None and name == "openhands-sdk"
        valid_url = (url.scheme == "file" and not url.netloc and not url.query and not url.fragment
                     and unquote(url.path) == str(sdk_wheel.resolve()) and digest == local_digest) if local_sdk else (
            url.scheme == "https" and url.hostname == "files.pythonhosted.org" and not url.username and not url.password
            and url.port in (None, 443) and url.path.endswith(".whl") and not url.query and not url.fragment)
        if not valid_url or not re.fullmatch(r"[0-9a-f]{64}", digest) or name in packages:
            raise ValueError("resolution requires unique hashed PyPI wheels or the exact local SDK wheel")
        packages[name] = (version, digest)
    if packages.get("openhands-sdk", (None,))[0] != (VARIANT_VERSION if sdk_wheel else "1.59.0"):
        raise ValueError("resolution does not contain the fixed SDK candidate")
    if sdk_wheel and {"lmnr", "lmnr-claude-code-proxy"}.intersection(packages):
        raise ValueError("default SDK variant must not resolve Laminar distributions")
    if len(packages) > 300:
        raise ValueError("candidate dependency count exceeds validation scope")
    return "# Resolved for this Python/platform image; require hashes when installing.\n" + "".join(
        f"{name}=={version} --hash=sha256:{digest}\n"
        for name, (version, digest) in sorted(packages.items())
    )


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--report", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--sdk-wheel", type=Path, help="accept only the exact local optional-Laminar variant")
    args = parser.parse_args()
    if args.report.stat().st_size > 32 * 1024 * 1024:
        raise ValueError("resolution report exceeds validation scope")
    lock = build_lock(json.loads(args.report.read_text(encoding="utf-8")), args.sdk_wheel)
    with args.output.open("x", encoding="utf-8") as target:
        target.write(lock)


if __name__ == "__main__":
    main()

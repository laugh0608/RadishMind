"""Freeze wheel hashes from pip's isolated candidate resolution report.

This utility does not resolve, download or install packages itself.
"""
from __future__ import annotations

import argparse
import json
import re
from pathlib import Path
from urllib.parse import urlsplit


def build_lock(report: dict) -> str:
    if report.get("version") != "1" or not isinstance(report.get("install"), list):
        raise ValueError("unsupported pip resolution report")
    packages: dict[str, tuple[str, str]] = {}
    for item in report["install"]:
        metadata = item.get("metadata", {})
        name, version = metadata.get("name", ""), metadata.get("version", "")
        if not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9._-]*", name) or not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9.!+_-]*", version):
            raise ValueError("invalid distribution identity")
        name = re.sub(r"[-_.]+", "-", name).lower()
        download = item.get("download_info", {})
        url = urlsplit(download.get("url", ""))
        digest = download.get("archive_info", {}).get("hashes", {}).get("sha256", "")
        if (url.scheme != "https" or url.hostname != "files.pythonhosted.org" or url.username or url.password
                or url.port not in (None, 443) or not url.path.endswith(".whl") or url.query or url.fragment
                or not re.fullmatch(r"[0-9a-f]{64}", digest) or name in packages):
            raise ValueError("resolution must contain unique hashed PyPI wheels only")
        packages[name] = (version, digest)
    if packages.get("openhands-sdk", (None,))[0] != "1.59.0":
        raise ValueError("resolution does not contain the fixed SDK candidate")
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
    args = parser.parse_args()
    if args.report.stat().st_size > 32 * 1024 * 1024:
        raise ValueError("resolution report exceeds validation scope")
    lock = build_lock(json.loads(args.report.read_text(encoding="utf-8")))
    with args.output.open("x", encoding="utf-8") as target:
        target.write(lock)


if __name__ == "__main__":
    main()

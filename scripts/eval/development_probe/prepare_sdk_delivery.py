"""Prepare a fresh offline SDK delivery context using the already verified wheels."""
from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path
import shutil

import pip_notice_delivery
import sdk_notice_delivery


ROOT = Path(__file__).resolve().parents[3]
PROBE = Path(__file__).resolve().parent
VARIANT = PROBE / "sdk_variant"


def digest(path: Path) -> str:
    with path.open("rb") as stream:
        return hashlib.file_digest(stream, "sha256").hexdigest()


def select_wheels(wheelhouse: Path, frozen: dict[str, tuple[str, str]]) -> dict[str, Path]:
    by_digest = {sha256: name for name, (_, sha256) in frozen.items()}
    wheels = {}
    for path in sorted(wheelhouse.iterdir()):
        if path.suffix != ".whl" or path.is_symlink() or not path.is_file():
            raise ValueError(f"unexpected wheelhouse entry: {path.name}")
        name = by_digest.get(digest(path))
        if name is None or name in wheels:
            raise ValueError(f"unknown or duplicate frozen wheel: {path.name}")
        # The pinned bytes establish identity. Also reject renamed files that pip
        # would skip or interpret as a different project or version.
        fields = path.stem.split("-")
        if (len(fields) not in (5, 6) or sdk_notice_delivery.normalized_name(fields[0]) != name
                or fields[1].lower() != frozen[name][0].lower()):
            raise ValueError(f"frozen wheel name or version mismatch: {path.name}")
        wheels[name] = path
    if wheels.keys() != frozen.keys():
        raise ValueError(f"missing frozen wheels: {sorted(frozen.keys() - wheels.keys())}")
    return wheels


def prepare(wheelhouse: Path, pillow_wheel: Path, directory: Path) -> None:
    acceptance = json.loads((VARIANT / "acceptance-20261031.json").read_text())
    build = json.loads((VARIANT / "pillow-build-window-launchd-result-20261121.json").read_text())
    recheck = json.loads((VARIANT / "pillow-sdk-recheck-result-20261121.json").read_text())
    original_path = VARIANT / "requirements-linux-arm64-py312.lock"
    original_lock = original_path.read_text()
    frozen = sdk_notice_delivery.frozen_packages(original_lock)
    if digest(original_path) != acceptance["lock_sha256"] or len(frozen) != 119:
        raise ValueError("historical wheel lock mismatch")
    if (pillow_wheel.is_symlink() or not pillow_wheel.is_file()
            or pillow_wheel.name != Path(build["wheel"]["path"]).name
            or pillow_wheel.stat().st_size != build["wheel"]["bytes"]
            or digest(pillow_wheel) != build["wheel"]["sha256"]
            or recheck["wheel"] != build["wheel"] or recheck["sdk_image"] != acceptance["image_id"]
            or recheck["runtime"]["exit_code"] != 0):
        raise ValueError("Pillow wheel does not match the build and target SDK evidence")
    pillow_version = recheck["runtime"]["installation"]["version"]
    old_version, old_sha256 = frozen["pillow"]
    old_line = f"pillow=={old_version} --hash=sha256:{old_sha256}\n"
    if original_lock.count(old_line) != 1:
        raise ValueError("expected one Pillow requirement in the original lock")
    pillow_sha256 = build["wheel"]["sha256"]
    lock = original_lock.replace(old_line, f"pillow=={pillow_version} --hash=sha256:{pillow_sha256}\n")
    lock_sha256 = hashlib.sha256(lock.encode()).hexdigest()
    source_index_path = VARIANT / "sdk-delivery-sources-20261121.json"
    source_index = json.loads(source_index_path.read_text())
    if source_index["wheel_lock_sha256"] != lock_sha256 or source_index["base_image"] != acceptance["base_image"]:
        raise ValueError("source inventory is for another SDK delivery context")
    recipe = ROOT / "deploy/Dockerfile.development-sdk-delivery"
    # Treat the retained base and its architecture as build inputs.
    from_image = "FROM docker.io/library/" + acceptance["base_image"]
    if [line for line in recipe.read_text().splitlines() if line.startswith("FROM ")] != [from_image]:
        raise ValueError("delivery recipe does not use the accepted base image")

    directory = directory.resolve()
    temp = ROOT / "tmp"
    if not directory.is_relative_to(temp) or directory == temp or directory.exists():
        raise ValueError("a new directory inside repository tmp is required")
    if wheelhouse.is_symlink() or not wheelhouse.is_dir():
        raise ValueError("a regular historical wheelhouse is required")
    wheels = select_wheels(wheelhouse, frozen)
    wheels["pillow"] = pillow_wheel
    frozen["pillow"] = (pillow_version, pillow_sha256)

    directory.mkdir(parents=True, mode=0o700)
    wheel_dir = directory / "wheels"
    wheel_dir.mkdir()
    for name, path in wheels.items():
        destination = wheel_dir / path.name
        if destination.exists():
            raise ValueError("duplicate delivery filename")
        shutil.copyfile(path, destination)
        destination.chmod(0o644)
        if digest(destination) != frozen[name][1]:
            raise ValueError(f"wheel changed during preparation: {path.name}")
    (directory / "requirements.lock").write_text(lock)
    shutil.copyfile(recipe, directory / "Dockerfile")
    for name in ("sdk_notice_delivery.py", "pip_notice_delivery.py", "inspect_sdk.py", "fixed_endpoint.py",
                 "test_sdk_contract.py", "test_sdk_optional_laminar.py"):
        shutil.copyfile(PROBE / name, directory / name)
    shutil.copyfile(ROOT / "contracts/development-task-probe-tool-request.schema.json",
                    directory / "tool-request.schema.json")
    shutil.copyfile(VARIANT / "source.json", directory / "source.json")
    notices = directory / "third-party"
    sdk_notice_delivery.prepare(VARIANT, notices, lock)
    pip_notice_delivery.prepare(VARIANT / "license-pip-notices-20261105.json", notices / "pip")
    source_delivery = notices / "source-delivery"
    source_delivery.mkdir()
    source_delivery.chmod(0o755)
    instructions = source_delivery / "README.txt"
    instructions.write_text(
        "取得本镜像时，同时向交付方取得 sources/<SDK-image-sha256>/。\n"
        "目录名使用交付镜像完整 SHA-256 的十六进制部分。\n"
        "在该目录读取 manifest.json，先校对镜像及 wheel 锁摘要，再核对清单中每个源码文件的字节数及 SHA-256。\n",
        encoding="utf-8",
    )
    instructions.chmod(0o644)
    report = {
        "schema_version": "development_sdk_delivery_preparation.v1",
        "platform": "linux/arm64",
        "python": acceptance["python"],
        "base_image": acceptance["base_image"],
        "historical_lock_sha256": acceptance["lock_sha256"],
        "wheel_lock_sha256": lock_sha256,
        "source_index_sha256": digest(source_index_path),
        "reused_wheels": len(wheels) - 1,
        "wheels": {name: {"file": wheels[name].name, "sha256": sha256}
                   for name, (_, sha256) in sorted(frozen.items())},
        "files": {path.relative_to(directory).as_posix(): digest(path)
                  for path in sorted(directory.rglob("*")) if path.is_file() and path.parent != wheel_dir},
    }
    (directory / "delivery-inputs.json").write_text(json.dumps(report, indent=2) + "\n")


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--wheelhouse", type=Path, required=True, help="retained original 119-wheel closure")
    parser.add_argument("--pillow-wheel", type=Path, required=True, help="exact repaired Pillow candidate")
    parser.add_argument("--directory", type=Path, required=True, help="new output directory in repository tmp")
    args = parser.parse_args()
    prepare(args.wheelhouse, args.pillow_wheel, args.directory)


if __name__ == "__main__":
    main()

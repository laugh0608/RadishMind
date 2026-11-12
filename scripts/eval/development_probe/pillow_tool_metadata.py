"""离线核对四项 Pillow 工具的 PyPI JSON；不含网络或安装入口。

显式复用仓库 .venv 中 pip 内嵌的 packaging；无导入回退。
运行前保存 --preflight 输出并核对解析器指纹，不改写历史获取回执。
"""

import argparse
import hashlib
import json
import re
import sys
from pathlib import Path
from urllib.parse import unquote, urlsplit

import pip
from pip._vendor import packaging
from pip._vendor.packaging.requirements import Requirement
from pip._vendor.packaging.specifiers import SpecifierSet
from pip._vendor.packaging.tags import compatible_tags, cpython_tags
from pip._vendor.packaging.utils import canonicalize_name, parse_wheel_filename
from pip._vendor.packaging.version import Version


TOOLS = ("setuptools", "pybind11", "meson", "ninja")
MAX_BODY_BYTES = 4 * 1024 * 1024
SETUPTOOLS_SHA256 = "29b23c360f22f414dc7336bb39178cc7bcbf6021ed2733cde173f09dba19abb3"
TARGET_PYTHON = "3.12.16"
# 仅目标声明标签，不读取 macOS 宿主的 sys_tags，也不声称镜像实物已验。
PLATFORMS = [f"manylinux_2_{minor}_aarch64" for minor in range(28, 16, -1)]
PLATFORMS.append("manylinux2014_aarch64")
TARGET_TAGS = set(cpython_tags((3, 12), abis=["cp312"], platforms=PLATFORMS))
TARGET_TAGS.update(compatible_tags((3, 12), interpreter="cp312", platforms=PLATFORMS))


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def parser_identity() -> dict:
    """本机解析器可用性及源码绑定；版本改变必须重新预检。"""
    root = Path(packaging.__file__).parent
    files = []
    for path in sorted(root.rglob("*.py")):
        data = path.read_bytes()
        files.append({"path": str(path.relative_to(root)), "bytes": len(data),
                      "sha256": hashlib.sha256(data).hexdigest()})
    return {"python_version": sys.version.split()[0], "pip_version": pip.__version__,
            "packaging_version": packaging.__version__,
            "provider": "pip._vendor.packaging", "source_files": files,
            "target_python": TARGET_PYTHON, "target_platforms": PLATFORMS,
            "target_tag_count": len(TARGET_TAGS), "network_requests": 0}


def unique_object(pairs: list) -> dict:
    result = {}
    for key, value in pairs:
        require(key not in result, f"duplicate_json_key: {key}")
        result[key] = value
    return result


def reject_constant(value: str) -> None:
    raise ValueError(f"invalid_json_constant: {value}")


def python_allowed(value: object) -> bool:
    require(value is None or isinstance(value, str), "invalid_requires_python_type")
    return value is None or SpecifierSet(value).contains(TARGET_PYTHON)


def validate_metadata(data: bytes, tool: str) -> dict:
    require(tool in TOOLS, "unsupported_tool")
    require(0 < len(data) <= MAX_BODY_BYTES, "metadata_body_size")
    document = json.loads(data.decode("utf-8"), object_pairs_hook=unique_object,
                          parse_constant=reject_constant)
    require(isinstance(document, dict), "metadata_object_required")
    info, distributions = document["info"], document["urls"]
    require(isinstance(info, dict), "info_object_required")
    require(isinstance(info["name"], str), "invalid_project_name")
    require(canonicalize_name(info["name"], validate=True) == tool, "project_mismatch")
    require(isinstance(info["version"], str), "invalid_version_type")
    version = Version(info["version"])
    require(not (version.is_prerelease or version.is_devrelease or version.local),
            "non_final_release")
    require(python_allowed(info["requires_python"]), "project_python_incompatible")
    dependencies = info["requires_dist"]
    require(dependencies is None or isinstance(dependencies, list), "invalid_requires_dist")
    for dependency in dependencies or []:
        require(isinstance(dependency, str), "invalid_requirement_type")
        Requirement(dependency)  # 校验语法，保留 marker，不按宿主求值或解析依赖闭包。
    if tool == "setuptools":
        require(version == Version("83.0.0"), "setuptools_version_mismatch")
    if tool == "meson":
        require(version >= Version("0.60.0"), "meson_below_known_source_minimum")
    require(isinstance(distributions, list) and 0 < len(distributions) <= 4096,
            "invalid_distribution_list")
    candidates, excluded, filenames = [], [], set()
    for distribution in distributions:
        require(isinstance(distribution, dict), "invalid_distribution_record")
        package_type = distribution["packagetype"]
        require(package_type in ("sdist", "bdist_wheel"), "unknown_distribution_type")
        if package_type == "sdist":
            continue
        filename = distribution["filename"]
        require(isinstance(filename, str) and "/" not in filename and "\\" not in filename,
                "invalid_wheel_filename")
        require(filename not in filenames, "duplicate_wheel_filename")
        filenames.add(filename)
        name, wheel_version, _build, tags = parse_wheel_filename(filename)
        require(name == tool and wheel_version == version, "wheel_identity_mismatch")
        require(type(distribution["yanked"]) is bool, "invalid_yanked_flag")
        allowed_python = python_allowed(distribution["requires_python"])
        url = distribution["url"]
        require(isinstance(url, str) and not any(ord(c) < 33 or ord(c) == 127 for c in url),
                "invalid_wheel_url")
        parsed = urlsplit(url)
        require(parsed.scheme == "https" and parsed.hostname == "files.pythonhosted.org"
                and parsed.port in (None, 443) and not parsed.username and not parsed.password
                and not parsed.query and not parsed.fragment and parsed.path.startswith("/packages/")
                and unquote(parsed.path.rsplit("/", 1)[-1]) == filename, "wheel_url_mismatch")
        digest = distribution["digests"]["sha256"]
        require(isinstance(digest, str) and re.fullmatch(r"[0-9a-f]{64}", digest) is not None,
                "invalid_sha256")
        size = distribution["size"]
        require(type(size) is int and size > 0, "invalid_distribution_size")
        matched = tags & TARGET_TAGS
        reason = None
        if distribution["yanked"]:
            reason = "yanked"
        elif not allowed_python:
            reason = "requires_python"
        elif not matched:
            reason = "target_tags"
        elif tool == "setuptools" and digest != SETUPTOOLS_SHA256:
            reason = "setuptools_locked_digest"
        if reason:
            excluded.append({"filename": filename, "reason": reason})
            continue
        candidates.append({"filename": filename, "url": url, "bytes": size, "sha256": digest,
                           "tags": sorted(map(str, tags)), "matched_tags": sorted(map(str, matched)),
                           "requires_python": distribution["requires_python"]})
    require(bool(candidates), "no_compatible_wheel_candidate")
    result = {"name": tool, "version": str(version), "requires_python": info["requires_python"],
              "requires_dist": dependencies, "metadata_bytes": len(data),
              "metadata_sha256": hashlib.sha256(data).hexdigest(),
              "candidate_wheels": candidates, "excluded_wheels": excluded,
              "status": "metadata_reviewed_not_adopted",
              "boundary": "只校验声明、依赖语法与目标标签；未获取 wheel、解析依赖闭包或验证构建器实物。"}
    if tool == "ninja":
        result["binary_version_status"] = "Python 分发包版本不等于内含 Ninja 实测版本。"
    return result


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--preflight", action="store_true", help="只检查本地解析器并输出指纹")
    parser.add_argument("--tool", choices=TOOLS)
    parser.add_argument("--input", type=Path, help="已留存的 PyPI JSON，最多 4 MiB")
    args = parser.parse_args()
    if args.preflight:
        if args.tool or args.input:
            parser.error("--preflight 不能与 --tool / --input 混用")
    elif not args.tool or not args.input:
        parser.error("需要 --tool 和 --input")
    try:
        if args.preflight:
            result = parser_identity()
        else:
            with args.input.open("rb") as stream:
                data = stream.read(MAX_BODY_BYTES + 1)
            result = validate_metadata(data, args.tool)
        print(json.dumps(result, ensure_ascii=False, indent=2))
        return 0
    except (ValueError, KeyError, TypeError, OSError, RecursionError) as error:
        print(f"metadata_validation_failed: {error}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())

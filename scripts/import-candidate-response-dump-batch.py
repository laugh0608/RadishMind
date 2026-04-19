#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import subprocess
import sys
from pathlib import Path

try:
    import jsonschema  # noqa: F401
except ModuleNotFoundError:
    print(
        "python package 'jsonschema' is required for import-candidate-response-dump-batch.py. "
        "Install it in the active environment before importing dump batches.",
        file=sys.stderr,
    )
    raise SystemExit(2)

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.candidate_records import (  # noqa: E402
    build_candidate_record_batch_manifest,
    candidate_response_record_from_dump,
    load_json_document,
    make_repo_relative,
    resolve_relative_to_repo,
    write_json_document,
)
from services.runtime.inference import recanonicalize_response_dump  # noqa: E402


TASK_SAMPLE_DIRS = {
    "radish-docs-qa": REPO_ROOT / "datasets/eval/radish",
    "radishflow-ghost-completion": REPO_ROOT / "datasets/eval/radishflow",
    "radishflow-suggest-edits": REPO_ROOT / "datasets/eval/radishflow",
}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Import a directory of raw candidate response dumps into formal candidate_response_record files, "
            "then build manifest and audit metadata."
        ),
    )
    parser.add_argument("task", choices=sorted(TASK_SAMPLE_DIRS), help="Eval task name to audit against.")
    parser.add_argument("--dump-dir", required=True, help="Directory containing raw dump json files.")
    parser.add_argument("--output-dir", required=True, help="Directory to write imported record files into.")
    parser.add_argument("--dump-pattern", default="*.dump.json", help="Glob used to select dump files.")
    parser.add_argument(
        "--manifest-output",
        default="",
        help="Optional manifest output path override. Defaults to <output-dir>/<collection_batch>.manifest.json.",
    )
    parser.add_argument("--manifest-description", default="", help="Optional manifest description.")
    parser.add_argument(
        "--report-output",
        default="",
        help="Optional audit report output path override. Defaults to <output-dir>/<collection_batch>.audit.json.",
    )
    parser.add_argument("--sample-dir", default="", help="Optional eval sample directory override for audit.")
    parser.add_argument(
        "--recanonicalize-response",
        action="store_true",
        help="Rebuild each dump.response from raw_response using the current runtime canonicalization before import.",
    )
    parser.add_argument("--fail-on-audit-violation", action="store_true", help="Fail when audit finds violations.")
    return parser.parse_args()


def derive_record_output_path(output_dir: Path, dump_path: Path) -> Path:
    name = dump_path.name
    if name.endswith(".dump.json"):
        return output_dir / (name[: -len(".dump.json")] + ".record.json")
    if name.endswith(".json"):
        return output_dir / (name[: -len(".json")] + ".record.json")
    return output_dir / (name + ".record.json")


def resolve_default_manifest_path(output_dir: Path, manifest: dict[str, object]) -> Path:
    collection_batch = str(manifest.get("collection_batch") or "").strip() or output_dir.name
    return output_dir / f"{collection_batch}.manifest.json"


def resolve_default_report_path(output_dir: Path, manifest: dict[str, object]) -> Path:
    collection_batch = str(manifest.get("collection_batch") or "").strip() or output_dir.name
    return output_dir / f"{collection_batch}.audit.json"


def run_audit(
    *,
    task: str,
    manifest_path: Path,
    sample_dir: str,
    report_path: Path,
    fail_on_audit_violation: bool,
) -> int:
    command = [
        sys.executable,
        str(REPO_ROOT / "scripts/audit-candidate-record-batch.py"),
        task,
        "--manifest",
        str(manifest_path),
        "--report-output",
        str(report_path),
    ]
    if sample_dir.strip():
        command.extend(["--sample-dir", sample_dir.strip()])
    if fail_on_audit_violation:
        command.append("--fail-on-violation")

    result = subprocess.run(command, cwd=REPO_ROOT, capture_output=True, text=True)
    if result.stdout:
        print(result.stdout, end="")
    if result.stderr:
        print(result.stderr, end="", file=sys.stderr)
    return result.returncode


def main() -> int:
    args = parse_args()
    dump_dir = resolve_relative_to_repo(args.dump_dir)
    output_dir = resolve_relative_to_repo(args.output_dir)

    if not dump_dir.is_dir():
        raise SystemExit(f"dump directory not found: {args.dump_dir}")

    dump_paths = sorted(path for path in dump_dir.glob(args.dump_pattern) if path.is_file())
    if not dump_paths:
        raise SystemExit(f"no dump files matched '{args.dump_pattern}' under: {args.dump_dir}")

    record_paths: list[Path] = []
    for dump_path in dump_paths:
        dump = load_json_document(dump_path)
        dump_label = make_repo_relative(dump_path)
        if args.recanonicalize_response:
            dump = recanonicalize_response_dump(dump, label=dump_label)

        record = candidate_response_record_from_dump(dump, label=dump_label)
        output_path = derive_record_output_path(output_dir, dump_path)
        write_json_document(output_path, record)
        record_paths.append(output_path)
        print(f"imported {dump_label} -> {make_repo_relative(output_path)}", file=sys.stderr)

    manifest = build_candidate_record_batch_manifest(
        record_paths,
        description=args.manifest_description,
    )
    manifest_path = (
        resolve_relative_to_repo(args.manifest_output)
        if args.manifest_output.strip()
        else resolve_default_manifest_path(output_dir, manifest)
    )
    report_path = (
        resolve_relative_to_repo(args.report_output)
        if args.report_output.strip()
        else resolve_default_report_path(output_dir, manifest)
    )
    write_json_document(manifest_path, manifest)

    audit_exit_code = run_audit(
        task=args.task,
        manifest_path=manifest_path,
        sample_dir=args.sample_dir,
        report_path=report_path,
        fail_on_audit_violation=args.fail_on_audit_violation,
    )
    if audit_exit_code != 0:
        return audit_exit_code

    summary = {
        "schema_version": 1,
        "task": args.task,
        "dump_dir": make_repo_relative(dump_dir),
        "output_dir": make_repo_relative(output_dir),
        "imported_record_count": len(record_paths),
        "manifest_path": make_repo_relative(manifest_path),
        "audit_report_path": make_repo_relative(report_path),
        "recanonicalized_response": bool(args.recanonicalize_response),
    }
    print(json.dumps(summary, ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

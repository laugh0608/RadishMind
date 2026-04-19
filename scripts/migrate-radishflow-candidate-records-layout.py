#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import shutil
import sys
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.candidate_records import (  # noqa: E402
    derive_candidate_batch_artifact_summary_path,
    derive_candidate_batch_audit_path,
    derive_candidate_batch_dump_path,
    derive_candidate_batch_manifest_path,
    derive_candidate_batch_output_root,
    derive_candidate_batch_record_path,
    derive_candidate_batch_response_path,
    load_json_document,
    make_repo_relative,
    resolve_manifest_record_path,
    write_json_document,
)


RADISHFLOW_ROOT = REPO_ROOT / "datasets/eval/candidate-records/radishflow"
FIXTURE_PATHS = [
    REPO_ROOT / "scripts/checks/fixtures/radishflow-ghost-real-batches.json",
    REPO_ROOT / "scripts/checks/fixtures/radishflow-suggest-edits-poc-batches.json",
    REPO_ROOT / "scripts/checks/fixtures/required-files.json",
]


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Migrate committed RadishFlow candidate-record batches to the short-layout path scheme.",
    )
    parser.add_argument("--check", action="store_true", help="Only report planned changes; do not write files.")
    return parser.parse_args()


def load_json(path: Path) -> Any:
    return json.loads(path.read_text(encoding="utf-8"))


def collect_fixture_batches() -> list[dict[str, Any]]:
    batches: list[dict[str, Any]] = []
    for fixture_path in FIXTURE_PATHS[:2]:
        payload = load_json(fixture_path)
        if not isinstance(payload, list):
            raise SystemExit(f"fixture must be a list: {make_repo_relative(fixture_path)}")
        for item in payload:
            if not isinstance(item, dict):
                raise SystemExit(f"fixture entry must be a json object: {make_repo_relative(fixture_path)}")
            manifest_value = str(item.get("manifest") or "").strip()
            if not manifest_value:
                raise SystemExit(f"fixture entry is missing manifest: {make_repo_relative(fixture_path)}")
            batches.append(
                {
                    "fixture_path": fixture_path,
                    "entry": item,
                    "manifest_path": REPO_ROOT / manifest_value,
                }
            )
    return batches


def safe_unlink(path: Path) -> None:
    if path.exists():
        path.unlink()


def safe_rmtree(path: Path) -> None:
    if path.is_dir():
        shutil.rmtree(path)


def rewrite_manifest_records(manifest_path: Path, manifest: dict[str, Any], new_output_root: Path) -> dict[str, Any]:
    rewritten = dict(manifest)
    rewritten["output_root"] = make_repo_relative(new_output_root)
    rewritten["batch_key"] = new_output_root.name
    rewritten_records: list[dict[str, Any]] = []
    for entry in manifest.get("records") or []:
        if not isinstance(entry, dict):
            raise SystemExit(f"{make_repo_relative(manifest_path)} contains non-object record entry")
        source_record_path = resolve_manifest_record_path(manifest_path, manifest, entry)
        source_record = load_json_document(source_record_path)
        if not isinstance(source_record, dict):
            raise SystemExit(f"record must be object: {make_repo_relative(source_record_path)}")
        sample_id = str(source_record.get("sample_id") or "").strip()
        task = str(source_record.get("task") or "").strip()
        target_record_path = derive_candidate_batch_record_path(
            output_root=new_output_root,
            task=task,
            sample_id=sample_id,
        )
        rewritten_entry = dict(entry)
        rewritten_entry["path"] = make_repo_relative(target_record_path)
        rewritten_entry["sample_key"] = target_record_path.stem[: -len(".record")] if target_record_path.name.endswith(".record.json") else target_record_path.stem
        rewritten_entry["record_relpath"] = str(target_record_path.relative_to(new_output_root)).replace("\\", "/")
        rewritten_records.append(rewritten_entry)
    rewritten["records"] = rewritten_records
    return rewritten


def rewrite_audit_document(audit_path: Path, document: dict[str, Any], new_manifest_path: Path) -> dict[str, Any]:
    rewritten = dict(document)
    rewritten["manifest_path"] = make_repo_relative(new_manifest_path)
    return rewritten


def rewrite_artifact_summary(document: dict[str, Any], *, new_output_root: Path, new_manifest_path: Path, new_audit_path: Path) -> dict[str, Any]:
    rewritten = dict(document)
    rewritten["output_root"] = make_repo_relative(new_output_root)
    artifacts = dict((rewritten.get("artifacts") or {}))
    for key, path in (
        ("manifest", new_manifest_path),
        ("audit_report", new_audit_path),
        ("output_root", new_output_root),
    ):
        artifact = dict((artifacts.get(key) or {}))
        artifact["path"] = make_repo_relative(path)
        artifacts[key] = artifact
    for key, rel in (
        ("records", "r"),
        ("responses", "o"),
        ("dumps", "d"),
    ):
        artifact = dict((artifacts.get(key) or {}))
        artifact["path"] = f"{make_repo_relative(new_output_root)}/{rel}"
        artifacts[key] = artifact
    rewritten["artifacts"] = artifacts
    return rewritten


def migrate_batch(batch: dict[str, Any], *, check_only: bool) -> dict[str, str]:
    manifest_path = Path(batch["manifest_path"]).resolve()
    manifest = load_json(manifest_path)
    if not isinstance(manifest, dict):
        raise SystemExit(f"manifest must be object: {make_repo_relative(manifest_path)}")
    collection_batch = str(manifest.get("collection_batch") or "").strip()
    project = str(manifest.get("project") or "").strip()
    new_output_root = derive_candidate_batch_output_root(project=project, collection_batch=collection_batch)
    old_output_root = manifest_path.parent
    path_map: dict[str, str] = {}
    if old_output_root.resolve() == new_output_root.resolve():
        return path_map

    rewritten_manifest = rewrite_manifest_records(manifest_path, manifest, new_output_root)
    new_manifest_path = derive_candidate_batch_manifest_path(new_output_root)
    old_audit_path = old_output_root / "audit.json"
    if not old_audit_path.is_file():
        old_audit_path = next(iter(sorted(old_output_root.glob("*.audit.json"))), Path())
    new_audit_path = derive_candidate_batch_audit_path(new_output_root)
    old_artifact_path = old_output_root / "artifacts.json"
    if not old_artifact_path.is_file():
        old_artifact_path = next(iter(sorted(old_output_root.glob("*.artifacts.json"))), Path())
    new_artifact_path = derive_candidate_batch_artifact_summary_path(new_output_root)

    path_map[make_repo_relative(old_output_root)] = make_repo_relative(new_output_root)
    path_map[make_repo_relative(manifest_path)] = make_repo_relative(new_manifest_path)
    if old_audit_path.is_file():
        path_map[make_repo_relative(old_audit_path)] = make_repo_relative(new_audit_path)
    if old_artifact_path.is_file():
        path_map[make_repo_relative(old_artifact_path)] = make_repo_relative(new_artifact_path)

    planned_copies: list[tuple[Path, Path]] = []
    original_entries = [entry for entry in manifest.get("records") or [] if isinstance(entry, dict)]
    rewritten_entries = [entry for entry in rewritten_manifest.get("records") or [] if isinstance(entry, dict)]
    if len(original_entries) != len(rewritten_entries):
        raise SystemExit(f"record count mismatch while rewriting manifest: {make_repo_relative(manifest_path)}")

    for original_entry, rewritten_entry in zip(original_entries, rewritten_entries):
        source_record_path = resolve_manifest_record_path(manifest_path, manifest, original_entry)
        target_record_path = REPO_ROOT / str(rewritten_entry["path"])
        source_record = load_json_document(source_record_path)
        if not isinstance(source_record, dict):
            raise SystemExit(f"record must be object: {make_repo_relative(source_record_path)}")
        sample_id = str(source_record.get("sample_id") or "").strip()
        task = str(source_record.get("task") or "").strip()
        target_response_path = derive_candidate_batch_response_path(
            output_root=new_output_root,
            task=task,
            sample_id=sample_id,
        )
        target_dump_path = derive_candidate_batch_dump_path(
            output_root=new_output_root,
            task=task,
            sample_id=sample_id,
        )

        source_record_name = source_record_path.name
        if source_record_name.endswith(".record.json"):
            source_base_name = source_record_name[: -len(".record.json")]
        elif source_record_name.endswith(".json"):
            source_base_name = source_record_name[: -len(".json")]
        else:
            source_base_name = source_record_path.stem
        source_response_candidates = list(old_output_root.rglob(f"{source_base_name}.response.json"))
        source_dump_candidates = list(old_output_root.rglob(f"{source_base_name}.dump.json"))
        planned_copies.append((source_record_path, target_record_path))
        path_map[make_repo_relative(source_record_path)] = make_repo_relative(target_record_path)
        if source_response_candidates:
            source_response_path = sorted(source_response_candidates)[0]
            planned_copies.append((source_response_path, target_response_path))
            path_map[make_repo_relative(source_response_path)] = make_repo_relative(target_response_path)
        if source_dump_candidates:
            source_dump_path = sorted(source_dump_candidates)[0]
            planned_copies.append((source_dump_path, target_dump_path))
            path_map[make_repo_relative(source_dump_path)] = make_repo_relative(target_dump_path)

    if check_only:
        return path_map

    safe_rmtree(new_output_root)
    for source_path, target_path in planned_copies:
        target_path.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(source_path, target_path)

    write_json_document(new_manifest_path, rewritten_manifest)

    if old_audit_path.is_file():
        audit_document = load_json(old_audit_path)
        if isinstance(audit_document, dict):
            write_json_document(new_audit_path, rewrite_audit_document(old_audit_path, audit_document, new_manifest_path))
    if old_artifact_path.is_file():
        artifact_document = load_json(old_artifact_path)
        if isinstance(artifact_document, dict):
            write_json_document(
                new_artifact_path,
                rewrite_artifact_summary(
                    artifact_document,
                    new_output_root=new_output_root,
                    new_manifest_path=new_manifest_path,
                    new_audit_path=new_audit_path,
                ),
            )

    safe_rmtree(old_output_root)
    return path_map


def rewrite_fixtures(mappings: dict[str, str], *, check_only: bool) -> None:
    for fixture_path in FIXTURE_PATHS:
        payload = load_json(fixture_path)
        serialized = json.dumps(payload, ensure_ascii=False, indent=2) + "\n"
        rewritten = serialized
        for old, new in sorted(mappings.items(), key=lambda item: len(item[0]), reverse=True):
            rewritten = rewritten.replace(old, new)
        if check_only:
            if rewritten != serialized:
                print(f"[plan] fixture update: {make_repo_relative(fixture_path)}")
            continue
        if rewritten != serialized:
            fixture_path.write_text(rewritten, encoding="utf-8")


def main() -> int:
    args = parse_args()
    batches = collect_fixture_batches()
    mappings: dict[str, str] = {}
    for batch in batches:
        batch_mappings = migrate_batch(batch, check_only=args.check)
        for old_path, new_path in sorted(batch_mappings.items()):
            mappings[old_path] = new_path
        if batch_mappings:
            old_root = str(next(iter(sorted(batch_mappings))))
            print(f"{old_root} -> {next(iter(sorted(batch_mappings.values())))}")
    rewrite_fixtures(mappings, check_only=args.check)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

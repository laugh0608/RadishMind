from __future__ import annotations

from .regression_shared import *  # noqa: F401,F403
from .regression_control_plane import validate_control_plane_request, validate_control_plane_response
from .regression_diagnostics_suggest import (
    validate_diagnostics_request,
    validate_diagnostics_response,
    validate_suggest_request,
    validate_suggest_response,
)
from .regression_docs import (
    validate_radish_docs_negative_replay,
    validate_radish_docs_response,
    validate_radish_docs_retrieval,
)
from .regression_ghost import validate_ghost_completion_request, validate_ghost_completion_response


def collect_sample_files(sample_dir: Path, sample_paths: list[Path]) -> list[Path]:
    if sample_paths:
        files: list[Path] = []
        for path in sample_paths:
            if not path.is_file():
                raise SystemExit(f"missing sample file: {path}")
            files.append(path)
        return files

    if not sample_dir.is_dir():
        raise SystemExit(f"missing sample directory: {sample_dir}")
    files = sorted(sample_dir.glob("*.json"))
    if not files:
        raise SystemExit("no sample files found for regression")
    return files


def run_radish_docs_qa(config: dict[str, Any], sample_dir: Path, sample_paths: list[Path], fail_on_violation: bool) -> int:
    sample_files = collect_sample_files(sample_dir, sample_paths)
    all_violations: list[str] = []
    matched_sample_count = 0

    for sample_file in sample_files:
        sample_name = sample_file.name
        violations: list[str] = []
        try:
            sample = json.loads(sample_file.read_text(encoding="utf-8"))
        except Exception as exc:
            add_violation(violations, f"{sample_name}: failed to parse json: {exc}")
            sample = None

        if sample is not None and str(sample.get("task") or "") == config["sample_task"]:
            matched_sample_count += 1
            test_document_against_schema(sample, config["sample_schema"], f"{sample_name} sample", violations)
            test_document_against_schema(sample.get("input_request"), config["request_schema"], f"{sample_name} input_request", violations)
            test_document_against_schema(sample.get("golden_response"), config["response_schema"], f"{sample_name} golden_response", violations)
            validate_radish_docs_retrieval(sample, sample_name, violations)
            validate_radish_docs_response(sample, sample["golden_response"], "golden_response", sample_name, violations)

            candidate_response, record_violations = load_candidate_response_from_record(sample, config, sample_name)
            violations.extend(record_violations)
            if candidate_response is not None:
                test_document_against_schema(candidate_response, config["response_schema"], f"{sample_name} candidate_response", violations)
                validate_radish_docs_response(sample, candidate_response, "candidate_response", sample_name, violations)
        elif sample is None:
            pass
        else:
            continue

        if violations:
            print(f"FAIL {sample_name}")
            for violation in violations:
                print(f"  - {violation}")
                all_violations.append(violation)
            continue

        print(f"PASS {sample_name}")

    if matched_sample_count == 0:
        raise SystemExit(config["no_sample_message"])
    if all_violations:
        if fail_on_violation:
            return 1
        print(f"WARNING: {config['warning_prefix']} {len(all_violations)} violation(s).", file=sys.stderr)
        return 0

    print(config["success_message"])
    return 0


def run_radish_docs_qa_negative(
    config: dict[str, Any],
    sample_dir: Path,
    sample_paths: list[Path],
    fail_on_violation: bool,
) -> int:
    sample_files = collect_sample_files(sample_dir, sample_paths)
    all_violations: list[str] = []
    matched_sample_count = 0

    for sample_file in sample_files:
        sample_name = sample_file.name
        violations: list[str] = []
        try:
            sample = json.loads(sample_file.read_text(encoding="utf-8"))
        except Exception as exc:
            add_violation(violations, f"{sample_name}: failed to parse json: {exc}")
            sample = None

        if sample is not None and str(sample.get("task") or "") == config["sample_task"]:
            matched_sample_count += 1
            test_document_against_schema(sample, config["sample_schema"], f"{sample_name} sample", violations)
            test_document_against_schema(sample.get("input_request"), config["request_schema"], f"{sample_name} input_request", violations)
            test_document_against_schema(sample.get("golden_response"), config["response_schema"], f"{sample_name} golden_response", violations)
            validate_radish_docs_retrieval(sample, sample_name, violations)
            validate_radish_docs_response(sample, sample["golden_response"], "golden_response", sample_name, violations)
            validate_radish_docs_negative_replay(sample, config, sample_name, violations)
        elif sample is None:
            pass
        else:
            continue

        if violations:
            print(f"FAIL {sample_name}")
            for violation in violations:
                print(f"  - {violation}")
                all_violations.append(violation)
            continue

        print(f"PASS {sample_name}")

    if matched_sample_count == 0:
        raise SystemExit(config["no_sample_message"])
    if all_violations:
        if fail_on_violation:
            return 1
        print(f"WARNING: {config['warning_prefix']} {len(all_violations)} violation(s).", file=sys.stderr)
        return 0

    print(config["success_message"])
    return 0


def run_radishflow_diagnostics(config: dict[str, Any], sample_dir: Path, sample_paths: list[Path], fail_on_violation: bool) -> int:
    sample_files = collect_sample_files(sample_dir, sample_paths)
    all_violations: list[str] = []
    matched_sample_count = 0

    for sample_file in sample_files:
        sample_name = sample_file.name
        violations: list[str] = []
        try:
            sample = json.loads(sample_file.read_text(encoding="utf-8"))
        except Exception as exc:
            add_violation(violations, f"{sample_name}: failed to parse json: {exc}")
            sample = None

        if sample is not None and str(sample.get("task") or "") == config["sample_task"]:
            matched_sample_count += 1
            test_document_against_schema(sample, config["sample_schema"], f"{sample_name} sample", violations)
            test_document_against_schema(sample.get("input_request"), config["request_schema"], f"{sample_name} input_request", violations)
            test_document_against_schema(sample.get("golden_response"), config["response_schema"], f"{sample_name} golden_response", violations)
            validate_diagnostics_request(sample, sample_name, violations)
            validate_diagnostics_response(sample, sample["golden_response"], "golden_response", sample_name, violations)

            candidate_response, record_violations = load_candidate_response_from_record(sample, config, sample_name)
            violations.extend(record_violations)
            if candidate_response is not None:
                test_document_against_schema(candidate_response, config["response_schema"], f"{sample_name} candidate_response", violations)
                validate_diagnostics_response(sample, candidate_response, "candidate_response", sample_name, violations)
        elif sample is None:
            pass
        else:
            continue

        if violations:
            print(f"FAIL {sample_name}")
            for violation in violations:
                print(f"  - {violation}")
                all_violations.append(violation)
            continue

        print(f"PASS {sample_name}")

    if matched_sample_count == 0:
        raise SystemExit(config["no_sample_message"])
    if all_violations:
        if fail_on_violation:
            return 1
        print(f"WARNING: {config['warning_prefix']} {len(all_violations)} violation(s).", file=sys.stderr)
        return 0

    print(config["success_message"])
    return 0


def run_radishflow_suggest_edits(config: dict[str, Any], sample_dir: Path, sample_paths: list[Path], fail_on_violation: bool) -> int:
    sample_files = collect_sample_files(sample_dir, sample_paths)
    all_violations: list[str] = []
    matched_sample_count = 0

    for sample_file in sample_files:
        sample_name = sample_file.name
        violations: list[str] = []
        try:
            sample = json.loads(sample_file.read_text(encoding="utf-8"))
        except Exception as exc:
            add_violation(violations, f"{sample_name}: failed to parse json: {exc}")
            sample = None

        if sample is not None and str(sample.get("task") or "") == config["sample_task"]:
            matched_sample_count += 1
            test_document_against_schema(sample, config["sample_schema"], f"{sample_name} sample", violations)
            test_document_against_schema(sample.get("input_request"), config["request_schema"], f"{sample_name} input_request", violations)
            test_document_against_schema(sample.get("golden_response"), config["response_schema"], f"{sample_name} golden_response", violations)
            validate_suggest_request(sample, sample_name, violations)
            validate_suggest_response(sample, sample["golden_response"], "golden_response", sample_name, violations)

            candidate_response, record_violations = load_candidate_response_from_record(sample, config, sample_name)
            violations.extend(record_violations)
            if candidate_response is not None:
                test_document_against_schema(candidate_response, config["response_schema"], f"{sample_name} candidate_response", violations)
                validate_suggest_response(sample, candidate_response, "candidate_response", sample_name, violations)
        elif sample is None:
            pass
        else:
            continue

        if violations:
            print(f"FAIL {sample_name}")
            for violation in violations:
                print(f"  - {violation}")
                all_violations.append(violation)
            continue

        print(f"PASS {sample_name}")

    if matched_sample_count == 0:
        raise SystemExit(config["no_sample_message"])
    if all_violations:
        if fail_on_violation:
            return 1
        print(f"WARNING: {config['warning_prefix']} {len(all_violations)} violation(s).", file=sys.stderr)
        return 0

    print(config["success_message"])
    return 0


def run_radishflow_ghost_completion(
    config: dict[str, Any],
    sample_dir: Path,
    sample_paths: list[Path],
    fail_on_violation: bool,
) -> int:
    sample_files = collect_sample_files(sample_dir, sample_paths)
    all_violations: list[str] = []
    matched_sample_count = 0

    for sample_file in sample_files:
        sample_name = sample_file.name
        violations: list[str] = []
        try:
            sample = json.loads(sample_file.read_text(encoding="utf-8"))
        except Exception as exc:
            add_violation(violations, f"{sample_name}: failed to parse json: {exc}")
            sample = None

        if sample is not None and str(sample.get("task") or "") == config["sample_task"]:
            matched_sample_count += 1
            test_document_against_schema(sample, config["sample_schema"], f"{sample_name} sample", violations)
            test_document_against_schema(sample.get("input_request"), config["request_schema"], f"{sample_name} input_request", violations)
            test_document_against_schema(sample.get("golden_response"), config["response_schema"], f"{sample_name} golden_response", violations)
            validate_ghost_completion_request(sample, sample_name, violations)
            validate_ghost_completion_response(sample, sample["golden_response"], "golden_response", sample_name, violations)

            candidate_response, record_violations = load_candidate_response_from_record(sample, config, sample_name)
            violations.extend(record_violations)
            if candidate_response is not None:
                test_document_against_schema(candidate_response, config["response_schema"], f"{sample_name} candidate_response", violations)
                validate_ghost_completion_response(sample, candidate_response, "candidate_response", sample_name, violations)
        elif sample is None:
            pass
        else:
            continue

        if violations:
            print(f"FAIL {sample_name}")
            for violation in violations:
                print(f"  - {violation}")
                all_violations.append(violation)
            continue

        print(f"PASS {sample_name}")

    if matched_sample_count == 0:
        raise SystemExit(config["no_sample_message"])
    if all_violations:
        if fail_on_violation:
            return 1
        print(f"WARNING: {config['warning_prefix']} {len(all_violations)} violation(s).", file=sys.stderr)
        return 0

    print(config["success_message"])
    return 0


def run_radishflow_control_plane(config: dict[str, Any], sample_dir: Path, sample_paths: list[Path], fail_on_violation: bool) -> int:
    sample_files = collect_sample_files(sample_dir, sample_paths)
    all_violations: list[str] = []
    matched_sample_count = 0

    for sample_file in sample_files:
        sample_name = sample_file.name
        violations: list[str] = []
        try:
            sample = json.loads(sample_file.read_text(encoding="utf-8"))
        except Exception as exc:
            add_violation(violations, f"{sample_name}: failed to parse json: {exc}")
            sample = None

        if sample is not None and str(sample.get("task") or "") == config["sample_task"]:
            matched_sample_count += 1
            test_document_against_schema(sample, config["sample_schema"], f"{sample_name} sample", violations)
            test_document_against_schema(sample.get("input_request"), config["request_schema"], f"{sample_name} input_request", violations)
            test_document_against_schema(sample.get("golden_response"), config["response_schema"], f"{sample_name} golden_response", violations)
            validate_control_plane_request(sample, sample_name, violations)
            validate_control_plane_response(sample, sample["golden_response"], "golden_response", sample_name, violations)

            candidate_response, record_violations = load_candidate_response_from_record(sample, config, sample_name)
            violations.extend(record_violations)
            if candidate_response is not None:
                test_document_against_schema(candidate_response, config["response_schema"], f"{sample_name} candidate_response", violations)
                validate_control_plane_response(sample, candidate_response, "candidate_response", sample_name, violations)
        elif sample is None:
            pass
        else:
            continue

        if violations:
            print(f"FAIL {sample_name}")
            for violation in violations:
                print(f"  - {violation}")
                all_violations.append(violation)
            continue

        print(f"PASS {sample_name}")

    if matched_sample_count == 0:
        raise SystemExit(config["no_sample_message"])
    if all_violations:
        if fail_on_violation:
            return 1
        print(f"WARNING: {config['warning_prefix']} {len(all_violations)} violation(s).", file=sys.stderr)
        return 0

    print(config["success_message"])
    return 0


def parse_args(argv: list[str]) -> tuple[str, Path | None, list[Path], bool, dict[str, Any]]:
    if len(argv) < 2 or argv[1] not in TASK_CONFIG:
        task_names = ", ".join(sorted(TASK_CONFIG))
        raise SystemExit(f"usage: {Path(argv[0]).name} <task> [options]\navailable tasks: {task_names}")

    task_name = argv[1]
    sample_dir: Path | None = None
    sample_paths: list[Path] = []
    fail_on_violation = False
    negative_replay_index: Path | None = None
    batch_artifact_summary: Path | None = None
    group_ids: list[str] = []
    record_ids: list[str] = []
    replay_mode = ""
    recommended_groups_top: int | None = None
    index = 2

    while index < len(argv):
        arg = argv[index]
        if arg in {"-FailOnViolation", "--fail-on-violation"}:
            fail_on_violation = True
            index += 1
            continue
        if arg in {"-SampleDir", "--sample-dir"}:
            if index + 1 >= len(argv):
                raise SystemExit(f"missing value for {arg}")
            sample_dir = (REPO_ROOT / argv[index + 1]).resolve() if not Path(argv[index + 1]).is_absolute() else Path(argv[index + 1])
            index += 2
            continue
        if arg in {"-SamplePaths", "--sample-paths"}:
            index += 1
            if index >= len(argv):
                raise SystemExit(f"missing value for {arg}")
            while index < len(argv) and not argv[index].startswith("-"):
                path = Path(argv[index])
                sample_paths.append((REPO_ROOT / path).resolve() if not path.is_absolute() else path)
                index += 1
            continue
        if arg in {"-NegativeReplayIndex", "--negative-replay-index"}:
            if index + 1 >= len(argv):
                raise SystemExit(f"missing value for {arg}")
            path = Path(argv[index + 1])
            negative_replay_index = (REPO_ROOT / path).resolve() if not path.is_absolute() else path
            index += 2
            continue
        if arg in {"-BatchArtifactSummary", "--batch-artifact-summary"}:
            if index + 1 >= len(argv):
                raise SystemExit(f"missing value for {arg}")
            path = Path(argv[index + 1])
            batch_artifact_summary = (REPO_ROOT / path).resolve() if not path.is_absolute() else path
            index += 2
            continue
        if arg in {"-GroupId", "--group-id"}:
            if index + 1 >= len(argv):
                raise SystemExit(f"missing value for {arg}")
            group_ids.append(argv[index + 1].strip())
            index += 2
            continue
        if arg in {"-RecordId", "--record-id"}:
            if index + 1 >= len(argv):
                raise SystemExit(f"missing value for {arg}")
            record_ids.append(argv[index + 1].strip())
            index += 2
            continue
        if arg in {"-ReplayMode", "--replay-mode"}:
            if index + 1 >= len(argv):
                raise SystemExit(f"missing value for {arg}")
            replay_mode = argv[index + 1].strip()
            index += 2
            continue
        if arg in {"-RecommendedGroupsTop", "--recommended-groups-top"}:
            if index + 1 >= len(argv):
                raise SystemExit(f"missing value for {arg}")
            try:
                recommended_groups_top = int(argv[index + 1])
            except ValueError as exc:
                raise SystemExit(f"invalid integer value for {arg}: {argv[index + 1]}") from exc
            index += 2
            continue
        raise SystemExit(f"unsupported argument: {arg}")

    selection = {
        "negative_replay_index": negative_replay_index,
        "batch_artifact_summary": batch_artifact_summary,
        "group_ids": [group_id for group_id in group_ids if group_id],
        "record_ids": [record_id for record_id in record_ids if record_id],
        "replay_mode": replay_mode,
        "recommended_groups_top": recommended_groups_top,
    }
    return task_name, sample_dir, sample_paths, fail_on_violation, selection


def ensure_required_paths(config: dict[str, Any]) -> None:
    for required_path in (
        config["sample_schema"],
        config.get("candidate_record_schema"),
        config["request_schema"],
        config["response_schema"],
        config["task_card"],
    ):
        if required_path is None:
            continue
        if not Path(required_path).is_file():
            raise SystemExit(f"missing required file: {required_path}")


def main(argv: list[str]) -> int:
    task_name, sample_dir, sample_paths, fail_on_violation, selection = parse_args(argv)
    config = TASK_CONFIG[task_name]
    ensure_required_paths(config)
    resolved_sample_dir = sample_dir or config["sample_dir"]
    negative_replay_index = selection["negative_replay_index"]
    batch_artifact_summary = selection["batch_artifact_summary"]
    recommended_groups_top = selection["recommended_groups_top"]
    if recommended_groups_top is not None:
        if task_name != "radish-docs-qa-negative":
            raise SystemExit("--recommended-groups-top is only supported for radish-docs-qa-negative")
        if batch_artifact_summary is None:
            raise SystemExit("--recommended-groups-top requires --batch-artifact-summary")
        if sample_paths:
            raise SystemExit("--recommended-groups-top cannot be used together with --sample-paths")
        if negative_replay_index is not None:
            raise SystemExit("--recommended-groups-top cannot be used together with --negative-replay-index")
        if selection["group_ids"]:
            raise SystemExit("--recommended-groups-top cannot be used together with --group-id")
        recommended_group_ids, default_replay_mode = resolve_recommended_negative_replay_groups_from_batch_artifact_summary(
            batch_artifact_summary,
            recommended_groups_top,
            selection["replay_mode"],
        )
        selection["group_ids"] = recommended_group_ids
        if not selection["replay_mode"]:
            selection["replay_mode"] = default_replay_mode
    if batch_artifact_summary is not None:
        if task_name != "radish-docs-qa-negative":
            raise SystemExit("--batch-artifact-summary is only supported for radish-docs-qa-negative")
        if sample_paths:
            raise SystemExit("--batch-artifact-summary cannot be used together with --sample-paths")
        if negative_replay_index is not None:
            raise SystemExit("--batch-artifact-summary cannot be used together with --negative-replay-index")
        negative_replay_index = resolve_negative_replay_index_from_batch_artifact_summary(
            batch_artifact_summary,
            selection["replay_mode"],
        )
    if negative_replay_index is not None:
        if task_name != "radish-docs-qa-negative":
            raise SystemExit("--negative-replay-index is only supported for radish-docs-qa-negative")
        if sample_paths:
            raise SystemExit("--negative-replay-index cannot be used together with --sample-paths")
        sample_paths = resolve_negative_replay_sample_paths(
            negative_replay_index,
            group_ids=selection["group_ids"],
            record_ids=selection["record_ids"],
            replay_mode=selection["replay_mode"],
        )

    if task_name == "radish-docs-qa":
        return run_radish_docs_qa(config, resolved_sample_dir, sample_paths, fail_on_violation)
    if task_name == "radish-docs-qa-negative":
        return run_radish_docs_qa_negative(config, resolved_sample_dir, sample_paths, fail_on_violation)
    if task_name == "radishflow-diagnostics":
        return run_radishflow_diagnostics(config, resolved_sample_dir, sample_paths, fail_on_violation)
    if task_name == "radishflow-ghost-completion":
        return run_radishflow_ghost_completion(config, resolved_sample_dir, sample_paths, fail_on_violation)
    if task_name == "radishflow-control-plane":
        return run_radishflow_control_plane(config, resolved_sample_dir, sample_paths, fail_on_violation)
    return run_radishflow_suggest_edits(config, resolved_sample_dir, sample_paths, fail_on_violation)

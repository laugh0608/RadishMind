from __future__ import annotations

from typing import Any


def parse_regression_output(stdout_text: str, stderr_text: str) -> dict[str, Any]:
    samples: list[dict[str, Any]] = []
    current_sample: dict[str, Any] | None = None
    summary_line = ""
    warning_line = ""

    for raw_line in stdout_text.splitlines():
        line = raw_line.rstrip()
        if line.startswith("PASS "):
            sample_name = line[len("PASS ") :].strip()
            current_sample = {
                "sample_file": sample_name,
                "status": "pass",
                "violations": [],
            }
            samples.append(current_sample)
            continue
        if line.startswith("FAIL "):
            sample_name = line[len("FAIL ") :].strip()
            current_sample = {
                "sample_file": sample_name,
                "status": "fail",
                "violations": [],
            }
            samples.append(current_sample)
            continue
        if line.startswith("  - ") and current_sample is not None:
            current_sample["violations"].append(line[4:])
            continue
        if "regression passed." in line:
            summary_line = line
            continue
        if line.startswith("WARNING: "):
            warning_line = line

    passed_count = sum(1 for sample in samples if sample["status"] == "pass")
    failed_count = sum(1 for sample in samples if sample["status"] == "fail")
    violation_count = sum(len(sample["violations"]) for sample in samples)
    warning_line = warning_line or next((line for line in stderr_text.splitlines() if line.startswith("WARNING: ")), "")
    return {
        "passed_count": passed_count,
        "failed_count": failed_count,
        "violation_count": violation_count,
        "summary_line": summary_line,
        "warning_line": warning_line,
        "samples": samples,
    }

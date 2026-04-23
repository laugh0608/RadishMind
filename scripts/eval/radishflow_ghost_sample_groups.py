#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any

DEFAULT_SAMPLE_PATHS = [
    "datasets/eval/radishflow/suggest-ghost-completion-flash-vapor-outlet-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-valve-ambiguous-no-tab-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-stop-no-legal-outlet-001.json",
]

SAMPLE_GROUP_PATHS = {
    "default-poc-trio": list(DEFAULT_SAMPLE_PATHS),
    "high-value-chain-recovery-boundaries": [
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-tab-after-dismiss-cooldown-and-other-skip-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-stop-no-legal-outlet-after-mixed-history-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001.json",
    ],
    "high-value-alternate-and-latest-action": [
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-stop-no-legal-outlet-after-mixed-history-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-alternate-candidate-tab-after-other-reject-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlet-latest-reject-after-skip-no-retab-001.json",
    ],
    "high-value-dismiss-and-empty-template-spread": [
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-latest-dismiss-after-reject-no-retab-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-alternate-candidate-tab-after-other-dismiss-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-stop-no-legal-outlet-after-mixed-history-001.json",
    ],
    "high-value-latest-dismiss-and-alternate-dismiss": [
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-alternate-candidate-tab-after-other-dismiss-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-outlet-latest-dismiss-after-reject-no-retab-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlet-latest-dismiss-after-reject-no-retab-001.json",
    ],
    "high-value-latest-dismiss-cooldown-recovery": [
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json",
    ],
    "high-value-reject-and-skip-cooldown-recovery": [
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-tab-after-latest-skip-cooldown-with-older-reject-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json",
    ],
    "remaining-latest-action-precedence": [
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-latest-reject-after-skip-no-retab-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-latest-skip-after-reject-no-retab-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-outlet-latest-reject-after-skip-no-retab-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-outlet-latest-skip-after-reject-no-retab-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlet-latest-skip-after-reject-no-retab-001.json",
    ],
    "remaining-other-candidate-recovery": [
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-alternate-candidate-tab-after-other-reject-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-alternate-candidate-tab-after-other-skip-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-alternate-candidate-tab-after-other-reject-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-alternate-candidate-tab-after-other-skip-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-alternate-candidate-tab-after-other-dismiss-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-alternate-candidate-tab-after-other-skip-001.json",
    ],
    "remaining-basic-no-retab-and-cooldown": [
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-skip-no-retab-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-tab-after-skip-cooldown-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-outlet-reject-no-retab-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-outlet-tab-after-reject-cooldown-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlet-dismiss-no-retab-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlet-tab-after-dismiss-cooldown-001.json",
    ],
    "remaining-foundation-and-conflict-basics": [
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-cooler-outlet-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-ranking-ambiguous-no-tab-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-heater-outlet-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-stop-no-legal-outlet-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-flash-outlets-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlets-name-conflict-no-tab-001.json",
    ],
    "remaining-general-basics": [
        "datasets/eval/radishflow/suggest-ghost-completion-context-gap-none-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-flash-inlet-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-flash-liquid-outlet-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-flash-nearby-node-ranking-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-heater-stream-name-001.json",
        "datasets/eval/radishflow/suggest-ghost-completion-mixer-standard-outlet-001.json",
    ],
}

HIGH_VALUE_PRIORITY_GROUPS = [
    "remaining-latest-action-precedence",
    "remaining-other-candidate-recovery",
    "remaining-basic-no-retab-and-cooldown",
    "remaining-foundation-and-conflict-basics",
    "remaining-general-basics",
]


def load_group_sample_id_map(repo_root: Path, group_names: list[str] | None = None) -> dict[str, list[str]]:
    normalized_group_names = group_names or list(SAMPLE_GROUP_PATHS)
    sample_id_map: dict[str, list[str]] = {}
    for group_name in normalized_group_names:
        raw_paths = SAMPLE_GROUP_PATHS.get(group_name)
        if raw_paths is None:
            continue
        sample_ids: list[str] = []
        for raw_path in raw_paths:
            document_path = (repo_root / raw_path).resolve()
            try:
                document = json.loads(document_path.read_text(encoding="utf-8"))
            except Exception as exc:
                raise SystemExit(f"failed to parse ghost sample '{raw_path}': {exc}") from exc
            if not isinstance(document, dict):
                raise SystemExit(f"expected json object in ghost sample: {raw_path}")
            sample_id = str(document.get("sample_id") or "").strip()
            if not sample_id:
                raise SystemExit(f"ghost sample is missing sample_id: {raw_path}")
            sample_ids.append(sample_id)
        sample_id_map[group_name] = sample_ids
    return sample_id_map


def build_pending_group_summaries(
    repo_root: Path,
    *,
    covered_sample_ids: set[str],
    group_names: list[str] | None = None,
) -> list[dict[str, Any]]:
    sample_id_map = load_group_sample_id_map(repo_root, group_names)
    pending_groups: list[dict[str, Any]] = []
    for group_name in group_names or list(sample_id_map):
        sample_ids = sample_id_map.get(group_name) or []
        missing_sample_ids = [sample_id for sample_id in sample_ids if sample_id not in covered_sample_ids]
        if not missing_sample_ids:
            continue
        pending_groups.append(
            {
                "group_name": group_name,
                "group_sample_count": len(sample_ids),
                "missing_sample_count": len(missing_sample_ids),
                "missing_sample_ids": missing_sample_ids,
            }
        )
    return pending_groups

#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import subprocess
import sys
from pathlib import Path

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent
GHOST_REQUEST_ASSEMBLY_FIXTURES = [
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-flash-basic-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-flash-basic-001.json",
                "request_id": "rf-ghost-request-flash-basic-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-flash-basic-001-debug-full.json",
                "request_id": "rf-ghost-request-flash-basic-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-valve-ambiguous-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-valve-ambiguous-001.json",
                "request_id": "rf-ghost-request-valve-ambiguous-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-valve-ambiguous-001-debug-full.json",
                "request_id": "rf-ghost-request-valve-ambiguous-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-flash-outlets-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-flash-outlets-001.json",
                "request_id": "rf-ghost-request-chain-feed-valve-flash-flash-outlets-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-flash-outlets-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-valve-flash-flash-outlets-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-stop-no-legal-outlet-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-stop-no-legal-outlet-001.json",
                "request_id": "rf-ghost-request-chain-feed-valve-flash-stop-no-legal-outlet-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-stop-no-legal-outlet-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-valve-flash-stop-no-legal-outlet-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlets-name-conflict-no-tab-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-name-conflict-no-tab-001.json",
                "request_id": "rf-ghost-request-chain-feed-valve-flash-outlets-name-conflict-no-tab-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-name-conflict-no-tab-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-valve-flash-outlets-name-conflict-no-tab-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlets-ranking-ambiguous-no-tab-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-ranking-ambiguous-no-tab-001.json",
                "request_id": "rf-ghost-request-chain-feed-valve-flash-outlets-ranking-ambiguous-no-tab-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-ranking-ambiguous-no-tab-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-valve-flash-outlets-ranking-ambiguous-no-tab-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-reject-no-retab-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-reject-no-retab-001.json",
                "request_id": "rf-ghost-request-chain-feed-valve-flash-outlet-reject-no-retab-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-reject-no-retab-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-valve-flash-outlet-reject-no-retab-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-dismiss-no-retab-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-dismiss-no-retab-001.json",
                "request_id": "rf-ghost-request-chain-feed-valve-flash-outlet-dismiss-no-retab-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-dismiss-no-retab-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-valve-flash-outlet-dismiss-no-retab-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-skip-no-retab-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-skip-no-retab-001.json",
                "request_id": "rf-ghost-request-chain-feed-valve-flash-outlet-skip-no-retab-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-skip-no-retab-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-valve-flash-outlet-skip-no-retab-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-reject-cooldown-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-reject-cooldown-001.json",
                "request_id": "rf-ghost-request-chain-feed-valve-flash-outlet-tab-after-reject-cooldown-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-reject-cooldown-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-valve-flash-outlet-tab-after-reject-cooldown-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-dismiss-cooldown-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-dismiss-cooldown-001.json",
                "request_id": "rf-ghost-request-chain-feed-valve-flash-outlet-tab-after-dismiss-cooldown-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-dismiss-cooldown-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-valve-flash-outlet-tab-after-dismiss-cooldown-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-skip-cooldown-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-skip-cooldown-001.json",
                "request_id": "rf-ghost-request-chain-feed-valve-flash-outlet-tab-after-skip-cooldown-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-skip-cooldown-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-valve-flash-outlet-tab-after-skip-cooldown-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-alternate-candidate-tab-after-other-reject-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-reject-001.json",
                "request_id": "rf-ghost-request-chain-feed-valve-flash-alternate-candidate-tab-after-other-reject-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-reject-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-valve-flash-alternate-candidate-tab-after-other-reject-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-alternate-candidate-tab-after-other-dismiss-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-dismiss-001.json",
                "request_id": "rf-ghost-request-chain-feed-valve-flash-alternate-candidate-tab-after-other-dismiss-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-dismiss-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-valve-flash-alternate-candidate-tab-after-other-dismiss-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-alternate-candidate-tab-after-other-skip-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-skip-001.json",
                "request_id": "rf-ghost-request-chain-feed-valve-flash-alternate-candidate-tab-after-other-skip-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-skip-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-valve-flash-alternate-candidate-tab-after-other-skip-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-heater-outlet-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-heater-outlet-001.json",
                "request_id": "rf-ghost-request-chain-feed-heater-flash-heater-outlet-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-heater-outlet-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-heater-flash-heater-outlet-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-alternate-candidate-tab-after-other-reject-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-reject-001.json",
                "request_id": "rf-ghost-request-chain-feed-heater-flash-alternate-candidate-tab-after-other-reject-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-reject-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-heater-flash-alternate-candidate-tab-after-other-reject-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-alternate-candidate-tab-after-other-dismiss-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-dismiss-001.json",
                "request_id": "rf-ghost-request-chain-feed-heater-flash-alternate-candidate-tab-after-other-dismiss-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-dismiss-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-heater-flash-alternate-candidate-tab-after-other-dismiss-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-alternate-candidate-tab-after-other-skip-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-skip-001.json",
                "request_id": "rf-ghost-request-chain-feed-heater-flash-alternate-candidate-tab-after-other-skip-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-skip-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-heater-flash-alternate-candidate-tab-after-other-skip-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-reject-cooldown-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-reject-cooldown-001.json",
                "request_id": "rf-ghost-request-chain-feed-heater-flash-outlet-tab-after-reject-cooldown-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-reject-cooldown-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-heater-flash-outlet-tab-after-reject-cooldown-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-dismiss-cooldown-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-dismiss-cooldown-001.json",
                "request_id": "rf-ghost-request-chain-feed-heater-flash-outlet-tab-after-dismiss-cooldown-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-dismiss-cooldown-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-heater-flash-outlet-tab-after-dismiss-cooldown-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-skip-cooldown-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-skip-cooldown-001.json",
                "request_id": "rf-ghost-request-chain-feed-heater-flash-outlet-tab-after-skip-cooldown-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-skip-cooldown-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-heater-flash-outlet-tab-after-skip-cooldown-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-reject-no-retab-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-reject-no-retab-001.json",
                "request_id": "rf-ghost-request-chain-feed-heater-flash-outlet-reject-no-retab-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-reject-no-retab-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-heater-flash-outlet-reject-no-retab-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-dismiss-no-retab-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-dismiss-no-retab-001.json",
                "request_id": "rf-ghost-request-chain-feed-heater-flash-outlet-dismiss-no-retab-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-dismiss-no-retab-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-heater-flash-outlet-dismiss-no-retab-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-skip-no-retab-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-skip-no-retab-001.json",
                "request_id": "rf-ghost-request-chain-feed-heater-flash-outlet-skip-no-retab-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-skip-no-retab-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-heater-flash-outlet-skip-no-retab-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-name-conflict-no-tab-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-name-conflict-no-tab-001.json",
                "request_id": "rf-ghost-request-chain-feed-heater-flash-outlet-name-conflict-no-tab-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-name-conflict-no-tab-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-heater-flash-outlet-name-conflict-no-tab-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-stop-no-legal-outlet-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-stop-no-legal-outlet-001.json",
                "request_id": "rf-ghost-request-chain-feed-heater-flash-stop-no-legal-outlet-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-stop-no-legal-outlet-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-heater-flash-stop-no-legal-outlet-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-ranking-ambiguous-no-tab-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-ranking-ambiguous-no-tab-001.json",
                "request_id": "rf-ghost-request-chain-feed-heater-flash-outlet-ranking-ambiguous-no-tab-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-ranking-ambiguous-no-tab-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-heater-flash-outlet-ranking-ambiguous-no-tab-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-cooler-outlet-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-cooler-outlet-001.json",
                "request_id": "rf-ghost-request-chain-feed-cooler-flash-cooler-outlet-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-cooler-outlet-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-cooler-flash-cooler-outlet-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-alternate-candidate-tab-after-other-reject-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-reject-001.json",
                "request_id": "rf-ghost-request-chain-feed-cooler-flash-alternate-candidate-tab-after-other-reject-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-reject-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-cooler-flash-alternate-candidate-tab-after-other-reject-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-alternate-candidate-tab-after-other-dismiss-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-dismiss-001.json",
                "request_id": "rf-ghost-request-chain-feed-cooler-flash-alternate-candidate-tab-after-other-dismiss-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-dismiss-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-cooler-flash-alternate-candidate-tab-after-other-dismiss-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-alternate-candidate-tab-after-other-skip-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-skip-001.json",
                "request_id": "rf-ghost-request-chain-feed-cooler-flash-alternate-candidate-tab-after-other-skip-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-skip-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-cooler-flash-alternate-candidate-tab-after-other-skip-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-reject-cooldown-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-reject-cooldown-001.json",
                "request_id": "rf-ghost-request-chain-feed-cooler-flash-outlet-tab-after-reject-cooldown-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-reject-cooldown-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-cooler-flash-outlet-tab-after-reject-cooldown-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-dismiss-cooldown-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-dismiss-cooldown-001.json",
                "request_id": "rf-ghost-request-chain-feed-cooler-flash-outlet-tab-after-dismiss-cooldown-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-dismiss-cooldown-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-cooler-flash-outlet-tab-after-dismiss-cooldown-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-skip-cooldown-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-skip-cooldown-001.json",
                "request_id": "rf-ghost-request-chain-feed-cooler-flash-outlet-tab-after-skip-cooldown-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-skip-cooldown-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-cooler-flash-outlet-tab-after-skip-cooldown-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-reject-no-retab-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-reject-no-retab-001.json",
                "request_id": "rf-ghost-request-chain-feed-cooler-flash-outlet-reject-no-retab-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-reject-no-retab-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-cooler-flash-outlet-reject-no-retab-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-dismiss-no-retab-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-dismiss-no-retab-001.json",
                "request_id": "rf-ghost-request-chain-feed-cooler-flash-outlet-dismiss-no-retab-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-dismiss-no-retab-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-cooler-flash-outlet-dismiss-no-retab-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-skip-no-retab-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-skip-no-retab-001.json",
                "request_id": "rf-ghost-request-chain-feed-cooler-flash-outlet-skip-no-retab-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-skip-no-retab-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-cooler-flash-outlet-skip-no-retab-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001.json",
                "request_id": "rf-ghost-request-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-stop-no-legal-outlet-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-stop-no-legal-outlet-001.json",
                "request_id": "rf-ghost-request-chain-feed-cooler-flash-stop-no-legal-outlet-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-stop-no-legal-outlet-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-cooler-flash-stop-no-legal-outlet-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
    {
        "candidate_set": "datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-ranking-ambiguous-no-tab-001.json",
        "requests": [
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-ranking-ambiguous-no-tab-001.json",
                "request_id": "rf-ghost-request-chain-feed-cooler-flash-outlet-ranking-ambiguous-no-tab-001",
                "assembly_profile": "model-minimal",
            },
            {
                "output": "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-ranking-ambiguous-no-tab-001-debug-full.json",
                "request_id": "rf-ghost-request-chain-feed-cooler-flash-outlet-ranking-ambiguous-no-tab-001-debug-full",
                "assembly_profile": "debug-full",
            },
        ],
    },
]

RADISH_DOCS_QA_REAL_BATCHES = [
    {
        "audit_report": "datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.audit.json",
        "artifact_summary": "datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.artifacts.json",
        "manifest": "datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.manifest.json",
        "negative_output_dir": "datasets/eval/radish-negative",
        "replay_index": "datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.negative-replay-index.json",
        "recommended_summary": "datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.recommended-negative-replay-top4-same_sample.summary.json",
        "recommended_top": "4",
    },
    {
        "audit_report": "datasets/eval/candidate-records/radish/2026-04-05-radish-docs-qa-real-batch-v1/2026-04-05-radish-docs-qa-real-batch-v1.audit.json",
        "artifact_summary": "datasets/eval/candidate-records/radish/2026-04-05-radish-docs-qa-real-batch-v1/2026-04-05-radish-docs-qa-real-batch-v1.artifacts.json",
        "manifest": "datasets/eval/candidate-records/radish/2026-04-05-radish-docs-qa-real-batch-v1/2026-04-05-radish-docs-qa-real-batch-v1.manifest.json",
        "negative_output_dir": "datasets/eval/radish-negative/2026-04-05-radish-docs-qa-real-batch-v1",
        "replay_index": "datasets/eval/candidate-records/radish/2026-04-05-radish-docs-qa-real-batch-v1/2026-04-05-radish-docs-qa-real-batch-v1.negative-replay-index.json",
        "cross_sample_negative_output_dir": "datasets/eval/radish-negative",
        "cross_sample_replay_index": "datasets/eval/candidate-records/radish/2026-04-05-radish-docs-qa-real-batch-v1/2026-04-05-radish-docs-qa-real-batch-v1.cross-sample-replay-index.json",
        "recommended_summary": "datasets/eval/candidate-records/radish/2026-04-05-radish-docs-qa-real-batch-v1/2026-04-05-radish-docs-qa-real-batch-v1.recommended-negative-replay-top5-same_sample.summary.json",
        "recommended_top": "5",
        "cross_sample_recommended_summary": "datasets/eval/candidate-records/radish/2026-04-05-radish-docs-qa-real-batch-v1/2026-04-05-radish-docs-qa-real-batch-v1.recommended-negative-replay-top2-cross_sample.summary.json",
        "cross_sample_recommended_top": "2",
    },
]

RADISH_DOCS_QA_REAL_DERIVED_NEGATIVES = {
    "manifest": "datasets/eval/candidate-records/radish-negative/2026-04-04-radish-docs-qa-simulated-negatives-v1.manifest.json",
    "negative_sample_dir": "datasets/eval/radish-negative",
    "index": "datasets/eval/candidate-records/radish-negative/2026-04-04-radish-docs-qa-simulated-negatives-v1.real-derived-index.json",
}

REQUIRED_FILES = [
    "AGENTS.md",
    "CLAUDE.md",
    "LICENSE",
    ".editorconfig",
    ".env.example",
    ".gitattributes",
    ".github/PULL_REQUEST_TEMPLATE.md",
    ".github/rulesets/README.md",
    ".github/rulesets/master-protection.json",
    ".github/workflows/pr-check.yml",
    ".github/workflows/release-check.yml",
    "README.md",
    "docs/README.md",
    "docs/task-cards/README.md",
    "docs/task-cards/radishflow-explain-diagnostics.md",
    "docs/task-cards/radishflow-suggest-flowsheet-edits.md",
    "docs/task-cards/radishflow-explain-control-plane-state.md",
    "docs/task-cards/radishflow-suggest-ghost-completion.md",
    "docs/task-cards/radish-answer-docs-question.md",
    "docs/radishmind-product-scope.md",
    "docs/radishmind-architecture.md",
    "docs/radishmind-roadmap.md",
    "docs/radishmind-integration-contracts.md",
    "docs/adr/0001-branch-and-pr-governance.md",
    "docs/devlogs/README.md",
    "docs/devlogs/2026-W14.md",
    "contracts/README.md",
    "contracts/copilot-request.schema.json",
    "contracts/copilot-response.schema.json",
    "contracts/radishflow-ghost-candidate-set.schema.json",
    "datasets/README.md",
    "datasets/examples/README.md",
    "datasets/examples/radishflow-ghost-candidate-set-flash-basic-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-valve-ambiguous-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-flash-outlets-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-stop-no-legal-outlet-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlets-name-conflict-no-tab-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlets-ranking-ambiguous-no-tab-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-reject-no-retab-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-dismiss-no-retab-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-skip-no-retab-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-reject-cooldown-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-dismiss-cooldown-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-skip-cooldown-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-alternate-candidate-tab-after-other-reject-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-alternate-candidate-tab-after-other-dismiss-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-alternate-candidate-tab-after-other-skip-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-heater-outlet-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-alternate-candidate-tab-after-other-reject-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-alternate-candidate-tab-after-other-dismiss-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-alternate-candidate-tab-after-other-skip-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-reject-cooldown-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-dismiss-cooldown-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-skip-cooldown-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-reject-no-retab-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-dismiss-no-retab-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-skip-no-retab-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-name-conflict-no-tab-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-stop-no-legal-outlet-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-ranking-ambiguous-no-tab-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-cooler-outlet-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-alternate-candidate-tab-after-other-reject-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-alternate-candidate-tab-after-other-dismiss-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-alternate-candidate-tab-after-other-skip-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-reject-cooldown-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-dismiss-cooldown-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-skip-cooldown-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-reject-no-retab-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-dismiss-no-retab-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-skip-no-retab-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-stop-no-legal-outlet-001.json",
    "datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-ranking-ambiguous-no-tab-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-flash-basic-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-flash-basic-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-valve-ambiguous-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-valve-ambiguous-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-flash-outlets-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-flash-outlets-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-stop-no-legal-outlet-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-stop-no-legal-outlet-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-name-conflict-no-tab-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-name-conflict-no-tab-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-ranking-ambiguous-no-tab-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-ranking-ambiguous-no-tab-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-reject-no-retab-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-reject-no-retab-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-dismiss-no-retab-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-dismiss-no-retab-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-skip-no-retab-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-skip-no-retab-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-reject-cooldown-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-reject-cooldown-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-dismiss-cooldown-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-dismiss-cooldown-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-skip-cooldown-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-skip-cooldown-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-reject-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-reject-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-dismiss-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-dismiss-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-skip-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-skip-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-heater-outlet-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-heater-outlet-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-reject-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-reject-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-dismiss-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-dismiss-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-skip-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-skip-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-reject-cooldown-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-reject-cooldown-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-dismiss-cooldown-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-dismiss-cooldown-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-skip-cooldown-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-skip-cooldown-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-reject-no-retab-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-reject-no-retab-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-dismiss-no-retab-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-dismiss-no-retab-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-skip-no-retab-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-skip-no-retab-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-name-conflict-no-tab-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-name-conflict-no-tab-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-stop-no-legal-outlet-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-stop-no-legal-outlet-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-ranking-ambiguous-no-tab-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-ranking-ambiguous-no-tab-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-cooler-outlet-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-cooler-outlet-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-reject-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-reject-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-dismiss-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-dismiss-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-skip-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-skip-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-reject-cooldown-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-reject-cooldown-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-dismiss-cooldown-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-dismiss-cooldown-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-skip-cooldown-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-skip-cooldown-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-reject-no-retab-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-reject-no-retab-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-dismiss-no-retab-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-dismiss-no-retab-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-skip-no-retab-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-skip-no-retab-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-stop-no-legal-outlet-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-stop-no-legal-outlet-001-debug-full.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-ranking-ambiguous-no-tab-001.json",
    "datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-ranking-ambiguous-no-tab-001-debug-full.json",
    "datasets/eval/README.md",
    "datasets/eval/batch-orchestration-summary.schema.json",
    "datasets/eval/candidate-record-batch.schema.json",
    "datasets/eval/candidate-response-dump.schema.json",
    "datasets/eval/negative-replay-index.schema.json",
    "datasets/eval/real-derived-negative-index.schema.json",
    "datasets/eval/recommended-negative-replay-summary.schema.json",
    "datasets/eval/candidate-records/radish/2026-04-03-radish-docs-qa-real-captures-v1.manifest.json",
    "datasets/eval/candidate-records/radish-negative/2026-04-04-radish-docs-qa-simulated-negatives-v1.manifest.json",
    "datasets/eval/candidate-records/radish-negative/2026-04-04-radish-docs-qa-simulated-negatives-v1.real-derived-index.json",
    "datasets/eval/radishflow-task-sample.schema.json",
    "datasets/eval/radish-task-sample.schema.json",
    "datasets/eval/radishflow/explain-diagnostics-global-balance-gap-001.json",
    "datasets/eval/radishflow/explain-diagnostics-multi-object-feed-conflict-001.json",
    "datasets/eval/radishflow/explain-diagnostics-stream-spec-missing-001.json",
    "datasets/eval/radishflow/explain-diagnostics-unit-not-converged-001.json",
    "datasets/eval/radishflow/explain-control-plane-conflicting-signals-001.json",
    "datasets/eval/radishflow/explain-control-plane-entitlement-expired-001.json",
    "datasets/eval/radishflow/explain-control-plane-package-sync-warning-001.json",
    "datasets/eval/radishflow/explain-control-plane-upstream-403-boundary-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-flash-inlet-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-flash-vapor-outlet-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-flash-liquid-outlet-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-flash-outlets-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-stop-no-legal-outlet-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-valve-outlet-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlets-name-conflict-no-tab-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlets-ranking-ambiguous-no-tab-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlet-reject-no-retab-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlet-dismiss-no-retab-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlet-skip-no-retab-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlet-tab-after-reject-cooldown-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlet-tab-after-dismiss-cooldown-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlet-tab-after-skip-cooldown-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-alternate-candidate-tab-after-other-reject-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-alternate-candidate-tab-after-other-dismiss-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-alternate-candidate-tab-after-other-skip-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-heater-outlet-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-alternate-candidate-tab-after-other-reject-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-alternate-candidate-tab-after-other-dismiss-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-alternate-candidate-tab-after-other-skip-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-outlet-tab-after-reject-cooldown-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-outlet-tab-after-dismiss-cooldown-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-outlet-tab-after-skip-cooldown-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-outlet-reject-no-retab-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-outlet-dismiss-no-retab-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-outlet-skip-no-retab-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-outlet-name-conflict-no-tab-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-stop-no-legal-outlet-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-outlet-ranking-ambiguous-no-tab-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-cooler-outlet-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-alternate-candidate-tab-after-other-reject-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-alternate-candidate-tab-after-other-dismiss-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-alternate-candidate-tab-after-other-skip-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-tab-after-reject-cooldown-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-tab-after-dismiss-cooldown-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-tab-after-skip-cooldown-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-reject-no-retab-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-dismiss-no-retab-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-skip-no-retab-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-stop-no-legal-outlet-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-ranking-ambiguous-no-tab-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-flash-nearby-node-ranking-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-heater-stream-name-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-mixer-standard-outlet-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-valve-ambiguous-no-tab-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-context-gap-none-001.json",
    "datasets/eval/radishflow/suggest-flowsheet-edits-stream-spec-placeholder-001.json",
    "datasets/eval/radishflow/suggest-flowsheet-edits-reconnect-outlet-001.json",
    "datasets/eval/radish/answer-docs-question-attachment-mixed-001.json",
    "datasets/eval/radish/answer-docs-question-docs-attachments-faq-001.json",
    "datasets/eval/radish/answer-docs-question-docs-attachments-forum-conflict-001.json",
    "datasets/eval/radish/answer-docs-question-direct-answer-001.json",
    "datasets/eval/radish/answer-docs-question-docs-faq-forum-conflict-001.json",
    "datasets/eval/radish/answer-docs-question-docs-faq-mixed-001.json",
    "datasets/eval/radish/answer-docs-question-evidence-gap-001.json",
    "datasets/eval/radish/answer-docs-question-forum-supplement-001.json",
    "datasets/eval/radish/answer-docs-question-navigation-001.json",
    "datasets/eval/radish/answer-docs-question-role-example-boundary-001.json",
    "datasets/eval/radish/answer-docs-question-wiki-faq-mixed-001.json",
    "prompts/tasks/radish-answer-docs-question-system.md",
    "services/__init__.py",
    "services/runtime/__init__.py",
    "services/runtime/candidate_records.py",
    "services/runtime/inference.py",
    "services/runtime/eval_regression.py",
    "scripts/check-radishflow-control-plane-eval.ps1",
    "scripts/check-radishflow-control-plane-eval.sh",
    "scripts/check-radishflow-diagnostics-eval.ps1",
    "scripts/check-radishflow-diagnostics-eval.sh",
    "scripts/check-radishflow-ghost-completion-eval.ps1",
    "scripts/check-radishflow-ghost-completion-eval.sh",
    "scripts/check-radishflow-suggest-edits-eval.ps1",
    "scripts/check-radishflow-suggest-edits-eval.sh",
    "scripts/check-radish-docs-qa-eval.ps1",
    "scripts/check-radish-docs-qa-eval.sh",
    "scripts/check-radish-docs-qa-real-batch-cross-sample-only-summary.py",
    "scripts/check-radish-docs-qa-real-batch-dual-recommended-summary.py",
    "scripts/check-radish-docs-qa-real-derived-negative-index.py",
    "scripts/check-radish-docs-qa-real-batch-same-sample-only-summary.py",
    "scripts/check_radish_docs_qa_real_batch_summary_common.py",
    "scripts/check-text-files.py",
    "scripts/audit-candidate-record-batch.py",
    "scripts/build-radish-docs-negative-replay.py",
    "scripts/build-real-derived-negative-index.py",
    "scripts/build-negative-replay-index.py",
    "scripts/build-candidate-record-batch.py",
    "scripts/build-radishflow-ghost-request.py",
    "scripts/import-candidate-response-dump.py",
    "scripts/run-copilot-inference.py",
    "scripts/run-radish-docs-qa-real-batch.py",
    "scripts/run-radish-docs-qa-real-batch.ps1",
    "scripts/run-radish-docs-qa-real-batch.sh",
    "scripts/run-radish-docs-qa-negative-recommended.py",
    "scripts/run-radish-docs-qa-negative-recommended.ps1",
    "scripts/run-radish-docs-qa-negative-recommended.sh",
    "scripts/check-repo.py",
    "scripts/run-eval-regression.py",
    "scripts/run-radishflow-control-plane-regression.ps1",
    "scripts/run-radishflow-control-plane-regression.sh",
    "scripts/run-radishflow-diagnostics-regression.ps1",
    "scripts/run-radishflow-diagnostics-regression.sh",
    "scripts/run-radishflow-ghost-completion-regression.ps1",
    "scripts/run-radishflow-ghost-completion-regression.sh",
    "scripts/run-radishflow-suggest-edits-regression.ps1",
    "scripts/run-radishflow-suggest-edits-regression.sh",
    "scripts/run-radish-docs-qa-regression.ps1",
    "scripts/run-radish-docs-qa-regression.sh",
    "scripts/check-text-files.ps1",
    "scripts/check-text-files.sh",
    "scripts/check-repo.ps1",
    "scripts/check-repo.sh",
]
REQUIRED_FILES.extend(
    [
        relative_path
        for batch in RADISH_DOCS_QA_REAL_BATCHES
        for relative_path in (
            batch["audit_report"],
            batch["artifact_summary"],
            batch["manifest"],
            batch["replay_index"],
            batch.get("cross_sample_replay_index"),
            batch["recommended_summary"],
            batch.get("cross_sample_recommended_summary"),
        )
        if relative_path
    ]
)


def run_python_script(script_name: str, args: list[str]) -> None:
    result = subprocess.run([sys.executable, str(REPO_ROOT / "scripts" / script_name), *args], cwd=REPO_ROOT)
    if result.returncode != 0:
        raise SystemExit(result.returncode)


def check_required_files() -> None:
    for relative_path in REQUIRED_FILES:
        if not (REPO_ROOT / relative_path).is_file():
            raise SystemExit(f"missing required file: {relative_path}")


def check_content_baseline() -> None:
    agents_content = (REPO_ROOT / "AGENTS.md").read_text(encoding="utf-8")
    if "当前常态开发分支为 `dev`" not in agents_content:
        raise SystemExit("AGENTS.md does not mention dev as the default development branch")

    pr_template_content = (REPO_ROOT / ".github/PULL_REQUEST_TEMPLATE.md").read_text(encoding="utf-8")
    if "默认目标分支为 `dev`" not in pr_template_content:
        raise SystemExit("PULL_REQUEST_TEMPLATE.md does not mention dev as the default target branch")

    ruleset = json.loads((REPO_ROOT / ".github/rulesets/master-protection.json").read_text(encoding="utf-8"))
    include_refs = (((ruleset.get("conditions") or {}).get("ref_name") or {}).get("include") or [])
    if "refs/heads/master" not in include_refs:
        raise SystemExit("master-protection.json does not target refs/heads/master")

    required_check_rule = next((rule for rule in ruleset.get("rules", []) if rule.get("type") == "required_status_checks"), None)
    if required_check_rule is None:
        raise SystemExit("master-protection.json is missing required_status_checks")

    contexts = [
        item.get("context")
        for item in ((required_check_rule.get("parameters") or {}).get("required_status_checks") or [])
    ]
    if "Repo Hygiene" not in contexts:
        raise SystemExit("master-protection.json is missing Repo Hygiene required check")
    if "Planning Baseline" not in contexts:
        raise SystemExit("master-protection.json is missing Planning Baseline required check")

    pr_workflow = (REPO_ROOT / ".github/workflows/pr-check.yml").read_text(encoding="utf-8")
    for pattern in ("name: PR Checks", "- master", "name: Repo Hygiene", "name: Planning Baseline"):
        if pattern not in pr_workflow:
            raise SystemExit(f".github/workflows/pr-check.yml is missing expected content: {pattern}")

    release_workflow = (REPO_ROOT / ".github/workflows/release-check.yml").read_text(encoding="utf-8")
    for pattern in (
        "name: Release Checks",
        "v*-dev",
        "v*-test",
        "v*-release",
        "name: Release Repo Hygiene",
        "name: Release Planning Baseline",
    ):
        if pattern not in release_workflow:
            raise SystemExit(f".github/workflows/release-check.yml is missing expected content: {pattern}")


def check_contract_schemas() -> None:
    contract_schema_paths = [
        REPO_ROOT / "contracts/copilot-request.schema.json",
        REPO_ROOT / "contracts/copilot-response.schema.json",
        REPO_ROOT / "contracts/radishflow-ghost-candidate-set.schema.json",
    ]
    for schema_path in contract_schema_paths:
        document = json.loads(schema_path.read_text(encoding="utf-8"))
        jsonschema.Draft202012Validator.check_schema(document)

    ghost_candidate_schema = json.loads(
        (REPO_ROOT / "contracts/radishflow-ghost-candidate-set.schema.json").read_text(encoding="utf-8")
    )
    copilot_request_schema = json.loads((REPO_ROOT / "contracts/copilot-request.schema.json").read_text(encoding="utf-8"))
    for fixture in GHOST_REQUEST_ASSEMBLY_FIXTURES:
        ghost_candidate_example = json.loads((REPO_ROOT / fixture["candidate_set"]).read_text(encoding="utf-8"))
        jsonschema.validate(ghost_candidate_example, ghost_candidate_schema)

        for request_fixture in fixture["requests"]:
            ghost_request_example = json.loads((REPO_ROOT / request_fixture["output"]).read_text(encoding="utf-8"))
            jsonschema.validate(ghost_request_example, copilot_request_schema)

            run_python_script(
                "build-radishflow-ghost-request.py",
                [
                    "--input",
                    fixture["candidate_set"],
                    "--output",
                    request_fixture["output"],
                    "--artifact-uri",
                    "artifact://flowsheet/current",
                    "--request-id",
                    request_fixture["request_id"],
                    "--assembly-profile",
                    request_fixture["assembly_profile"],
                    "--check",
                ],
            )


def check_generated_eval_metadata() -> None:
    schema = json.loads(
        (REPO_ROOT / "datasets/eval/negative-replay-index.schema.json").read_text(encoding="utf-8")
    )
    real_derived_index_schema = json.loads(
        (REPO_ROOT / "datasets/eval/real-derived-negative-index.schema.json").read_text(encoding="utf-8")
    )
    artifact_summary_schema = json.loads(
        (REPO_ROOT / "datasets/eval/batch-orchestration-summary.schema.json").read_text(encoding="utf-8")
    )
    recommended_summary_schema = json.loads(
        (REPO_ROOT / "datasets/eval/recommended-negative-replay-summary.schema.json").read_text(encoding="utf-8")
    )
    for batch in RADISH_DOCS_QA_REAL_BATCHES:
        run_python_script(
            "build-negative-replay-index.py",
            [
                "--audit-report",
                batch["audit_report"],
                "--negative-sample-dir",
                batch["negative_output_dir"],
                "--output",
                batch["replay_index"],
                "--check",
            ],
        )

        document = json.loads((REPO_ROOT / batch["replay_index"]).read_text(encoding="utf-8"))
        jsonschema.validate(document, schema)

        artifact_summary_document = json.loads((REPO_ROOT / batch["artifact_summary"]).read_text(encoding="utf-8"))
        jsonschema.validate(artifact_summary_document, artifact_summary_schema)

        recommended_summary_document = json.loads((REPO_ROOT / batch["recommended_summary"]).read_text(encoding="utf-8"))
        jsonschema.validate(recommended_summary_document, recommended_summary_schema)

        cross_sample_recommended_summary = batch.get("cross_sample_recommended_summary")
        if cross_sample_recommended_summary:
            cross_sample_recommended_summary_document = json.loads(
                (REPO_ROOT / cross_sample_recommended_summary).read_text(encoding="utf-8")
            )
            jsonschema.validate(cross_sample_recommended_summary_document, recommended_summary_schema)

        run_python_script(
            "build-radish-docs-negative-replay.py",
            [
                "--index",
                batch["replay_index"],
                "--output-dir",
                batch["negative_output_dir"],
                "--check",
            ],
        )

        cross_sample_replay_index = batch.get("cross_sample_replay_index")
        cross_sample_negative_output_dir = batch.get("cross_sample_negative_output_dir")
        if cross_sample_replay_index and cross_sample_negative_output_dir:
            run_python_script(
                "build-negative-replay-index.py",
                [
                    "--audit-report",
                    batch["audit_report"],
                    "--negative-sample-dir",
                    cross_sample_negative_output_dir,
                    "--output",
                    cross_sample_replay_index,
                    "--check",
                ],
            )

            cross_sample_document = json.loads((REPO_ROOT / cross_sample_replay_index).read_text(encoding="utf-8"))
            jsonschema.validate(cross_sample_document, schema)

            run_python_script(
                "build-radish-docs-negative-replay.py",
                [
                    "--index",
                    cross_sample_replay_index,
                    "--replay-mode",
                    "cross_sample",
                    "--check",
                ],
            )

            run_python_script(
                "run-eval-regression.py",
                [
                    "radish-docs-qa-negative",
                    "--negative-replay-index",
                    cross_sample_replay_index,
                    "--replay-mode",
                    "cross_sample",
                    "--fail-on-violation",
                ],
            )

            if cross_sample_recommended_summary:
                run_python_script(
                    "run-eval-regression.py",
                    [
                        "radish-docs-qa-negative",
                        "--batch-artifact-summary",
                        batch["artifact_summary"],
                        "--recommended-groups-top",
                        "1",
                        "--replay-mode",
                        "cross_sample",
                        "--fail-on-violation",
                    ],
                )

                run_python_script(
                    "run-radish-docs-qa-negative-recommended.py",
                    [
                        "--batch-artifact-summary",
                        batch["artifact_summary"],
                        "--top",
                        batch["cross_sample_recommended_top"],
                        "--replay-mode",
                        "cross_sample",
                        "--fail-on-violation",
                        "--summary-output",
                        cross_sample_recommended_summary,
                        "--check",
                    ],
                )

        run_python_script(
            "run-eval-regression.py",
            [
                "radish-docs-qa-negative",
                "--batch-artifact-summary",
                batch["artifact_summary"],
                "--recommended-groups-top",
                "1",
                "--fail-on-violation",
            ],
        )

        run_python_script(
            "run-radish-docs-qa-negative-recommended.py",
            [
                "--batch-artifact-summary",
                batch["artifact_summary"],
                "--top",
                batch["recommended_top"],
                "--fail-on-violation",
                "--summary-output",
                batch["recommended_summary"],
                "--check",
            ],
        )

    run_python_script("check-radish-docs-qa-real-batch-dual-recommended-summary.py", [])
    run_python_script("check-radish-docs-qa-real-batch-cross-sample-only-summary.py", [])
    run_python_script("check-radish-docs-qa-real-batch-same-sample-only-summary.py", [])

    run_python_script(
        "build-real-derived-negative-index.py",
        [
            "--manifest",
            RADISH_DOCS_QA_REAL_DERIVED_NEGATIVES["manifest"],
            "--negative-sample-dir",
            RADISH_DOCS_QA_REAL_DERIVED_NEGATIVES["negative_sample_dir"],
            "--output",
            RADISH_DOCS_QA_REAL_DERIVED_NEGATIVES["index"],
            "--check",
        ],
    )
    real_derived_index_document = json.loads(
        (REPO_ROOT / RADISH_DOCS_QA_REAL_DERIVED_NEGATIVES["index"]).read_text(encoding="utf-8")
    )
    jsonschema.validate(real_derived_index_document, real_derived_index_schema)
    run_python_script("check-radish-docs-qa-real-derived-negative-index.py", [])


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--skip-text-files", action="store_true")
    return parser.parse_args()


def main() -> int:
    args = parse_args()

    if not args.skip_text_files:
        run_python_script("check-text-files.py", [])

    run_python_script("run-eval-regression.py", ["radish-docs-qa", "--fail-on-violation"])
    run_python_script("run-eval-regression.py", ["radishflow-control-plane", "--fail-on-violation"])
    run_python_script("run-eval-regression.py", ["radishflow-diagnostics", "--fail-on-violation"])
    run_python_script("run-eval-regression.py", ["radishflow-ghost-completion", "--fail-on-violation"])
    run_python_script("run-eval-regression.py", ["radishflow-suggest-edits", "--fail-on-violation"])

    check_required_files()
    check_content_baseline()
    check_contract_schemas()
    check_generated_eval_metadata()

    print("repository baseline checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

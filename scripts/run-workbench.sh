#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"
ACTION="${1:-help}"
if [[ $# -gt 0 ]]; then shift; fi
BIN_DIR="${ROOT_DIR}/tmp/workbench-bin"
if [[ "${ACTION}" == "help" || "${ACTION}" == "--help" ]]; then
  cat <<'EOF'
Usage: ./start.sh workbench <action> [options]
  build                         Build using installed Go/Node dependencies; no installation.
  check | init | serve          --config var/workbench/platform.json
  bootstrap-admin               --config <file> --user-id <id> --audit-ref <reference>
  backup | upgrade              --config <file> --output <new-backup-directory>
  restore                       --source <backup-directory> --output <new-workbench-directory>
Serve stays in the foreground. Stop it before bootstrap, backup, upgrade or restore.
Prepare platform.json from deploy/workbench-native.example.json. No default account is created.
EOF
  exit 0
fi
if [[ "${ACTION}" == "build" ]]; then
  if [[ $# -ne 0 ]]; then echo "build does not accept arguments" >&2; exit 2; fi
  if [[ ! -d apps/radishmind-web/node_modules ]]; then echo "Install the documented Web dependencies explicitly first." >&2; exit 1; fi
  (cd apps/radishmind-web && npm run build -- --mode workbench)
  mkdir -p "${BIN_DIR}"
  BUILD_VERSION="$(git rev-parse --short=12 HEAD)"
  if [[ -n "$(git status --porcelain)" ]]; then BUILD_VERSION="${BUILD_VERSION}-dirty"; fi
  # Resolve only already installed modules. Missing dependencies fail explicitly.
  (cd services/platform && GOPROXY=off GOTOOLCHAIN=local go build -ldflags "-X main.buildVersion=${BUILD_VERSION}" -o "${BIN_DIR}/radishmind-platform" ./cmd/radishmind-platform)
  exit 0
fi
if [[ ! -x "${BIN_DIR}/radishmind-platform" ]]; then
  echo "Build the workbench first: ./start.sh workbench build" >&2
  exit 1
fi
exec "${BIN_DIR}/radishmind-platform" workbench "${ACTION}" "$@"

# Production Secret Backend Cloud Secret Service Selection Readiness v1 计划

## 背景

`Production Secret Backend Production Resolver Runtime Blocker Consolidation v1` 已把 `cloud_secret_service_selection` 固定为 production resolver runtime task card 的 blocker。当前 resolver backend profile selection 只定义 reserved backend kind 和 reference-only profile shape，仍没有 cloud secret service selection 的独立准入证据。

本批固定 cloud secret service selection readiness，让后续 provider choice 或 production resolver runtime task card 只能消费 metadata-only selection evidence，不绕过 credential handle、operator approval、audit store、backend health 和 no leakage smoke runtime blocker。

当前状态：`cloud_secret_service_selection_readiness_defined`。

## 范围

- 新增平台专题 `docs/platform/production-secret-backend-cloud-secret-service-selection-readiness-v1.md`。
- 新增 fixture `scripts/checks/fixtures/production-secret-backend-cloud-secret-service-selection-readiness-v1.json`。
- 新增 checker `scripts/check-production-ops-secret-backend-cloud-secret-service-selection-readiness-v1.py`。
- 接入 `scripts/check-repo.py`、`production-ops-secret-backend-implementation-readiness.json` 与 implementation readiness checker。
- 同步 `docs/radishmind-current-focus.md`、`docs/features/`、`docs/platform/README.md`、`docs/radishmind-integration-contracts.md`、`docs/task-cards/README.md`、`scripts/README.md` 和周志。

## Selection Requirements

cloud secret service selection readiness 必须固定：

- concrete cloud vendor 仍为 `not_selected`。
- 只允许 future reserved candidate kind，不绑定 AWS / GCP / Azure / Vault 等 SDK。
- provider profile、environment、secret ref namespace 和 policy version 都必须是 metadata-only binding。
- credential handle、operator approval、audit store、backend health 和 no leakage smoke runtime 仍是 production resolver runtime 前置 blocker。
- failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard 可由 checker 复验。

## 禁止事项

- 不选择具体云厂商。
- 不导入、声明或绑定云 SDK。
- 不创建 cloud secret client。
- 不调用云 secret service、provider、fake resolver 或 production resolver。
- 不读取真实 secret、环境 secret、provider raw URL、DSN、cloud credential 或 credential payload。
- 不创建 production resolver runtime implementation task card。
- 不创建 credential handle runtime、operator approval runtime、audit store runtime、backend health runtime、no leakage smoke runtime、DB provider、repository mode runtime、production API、executor、confirmation、writeback 或 replay。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-cloud-secret-service-selection-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-production-resolver-runtime-blocker-consolidation-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-resolver-backend-profile-selection-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```

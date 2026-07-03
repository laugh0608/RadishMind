# Production Secret Backend Audit Store Runtime Event Schema Artifact Implementation v1 计划

更新时间：2026-06-28

## 任务目标

本任务卡固定 `production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1` 的静态实施计划。目标是把后续 runtime event schema artifact 的 path、schema version pin、event kind allowlist、required / optional fields、reference-only policy、negative fixtures、schema checker、writer input compatibility smoke、no fallback、no side effects 和 artifact guard 固定为可复验要求。

结论状态为 `audit_store_runtime_event_schema_artifact_implementation_task_card_defined`。

本批不创建 `contracts/production-secret-audit-event.schema.json`，不创建 runtime schema validator，不创建 audit writer runtime、audit store runtime、delivery runtime、idempotency runtime、production resolver runtime、cloud secret client、DB provider、repository mode 或 public production API。

## 输入

- [Production Secret Backend Audit Store Runtime Event Schema Artifact Implementation Entry Review v1](../platform/production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-entry-review-v1.md)
- [Production Secret Backend Audit Store Contract / Event Schema Readiness v1](../platform/production-secret-backend-audit-store-contract-event-schema-readiness-v1.md)
- [Production Secret Backend Audit Store Runtime Event Schema Materialization Readiness v1](../platform/production-secret-backend-audit-store-runtime-event-schema-materialization-readiness-v1.md)
- [Production Secret Backend Audit Store Writer Runtime Boundary Readiness v1](../platform/production-secret-backend-audit-store-writer-runtime-boundary-readiness-v1.md)
- [Production Secret Backend Audit Store Runtime Implementation Entry Refresh v4](../platform/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4.md)
- `scripts/checks/fixtures/production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-entry-review-v1.json`
- `scripts/checks/fixtures/production-secret-backend-audit-store-contract-event-schema-readiness-v1.json`
- `scripts/checks/fixtures/production-secret-backend-audit-store-runtime-event-schema-materialization-readiness-v1.json`
- `scripts/checks/fixtures/production-secret-backend-audit-store-writer-runtime-boundary-readiness-v1.json`
- `scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4.json`
- `scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json`

## 交付物

1. 新增平台专题：
   - `docs/platform/production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1.md`
2. 新增 task card：
   - `docs/task-cards/production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1-plan.md`
3. 新增 fixture：
   - `scripts/checks/fixtures/production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1.json`
4. 新增 checker：
   - `scripts/check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1.py`
5. 更新聚合与入口：
   - `docs/radishmind-current-focus.md`
   - `docs/features/README.md`
   - `docs/features/workflow/README.md`
   - `docs/features/workflow-agent-runtime.md`
   - `docs/features/workflow/saved-workflow-draft-v1.md`
   - `docs/platform/README.md`
   - `docs/radishmind-roadmap.md`
   - `docs/radishmind-integration-contracts.md`
   - `docs/task-cards/README.md`
   - `docs/task-cards/production-secret-backend-implementation-v1-plan.md`
   - `scripts/README.md`
   - `scripts/check-repo.py`
6. 更新关联静态证据：
   - implementation readiness fixture 与 checker
   - 本周周志

## Artifact Requirements

| gate | 后续 artifact 实现要求 |
| --- | --- |
| schema artifact path | `contracts/production-secret-audit-event.schema.json` |
| schema version pin | `audit-event-schema-v1` |
| event kind allowlist | 必须等于 contract readiness allowlist |
| required fields | 必须等于 contract readiness required fields |
| optional fields | 必须等于 contract readiness optional fields |
| reference-only fields | 只允许 opaque reference / status / policy version |
| forbidden fields | 必须覆盖 contract forbidden fields 与 raw writer / event payload、schema payload、payload hash、secret-derived hash |
| positive fixture | 必须覆盖完整 metadata-only event |
| negative fixtures | 必须覆盖 missing required、forbidden field、additionalProperties、event kind invalid |
| schema checker | 必须验证 artifact、正例、负例、writer input compatibility 和 check-repo 注册顺序 |
| writer input compatibility | 只做静态 smoke，不创建 writer runtime |

## 停止线

- 不创建 runtime event schema artifact 或 runtime schema validator。
- 不创建 audit store runtime implementation task card、audit store runtime、audit writer runtime、audit event writer、writer result fixture、delivery runtime、idempotency runtime、duplicate detector runtime 或 retry executor。
- 不执行 audit event write、delivery、duplicate detection、approval runtime、backend health check、no leakage smoke runtime 或 production resolver runtime。
- 不调用 provider、云 secret 服务、数据库、production API 或网络服务。
- 不读取真实 secret、developer env secret、credential payload、DSN、provider raw URL、raw request / response / audit / writer / event payload、schema payload、payload hash 或 secret-derived hash。
- 不创建 production resolver runtime、cloud secret client、credential handle runtime、operator approval runtime、backend health runtime、no leakage smoke runtime、DB provider、SQL migration、schema marker、repository mode runtime 或 public production API。
- 不把本 task card 写成 audit store runtime ready、runtime schema artifact created、writer runtime ready、delivery runtime ready、idempotency runtime ready、production resolver ready 或 production ready。

## 验证

本批至少执行：

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-contract-event-schema-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```

若触及阶段边界、真相源或仓库级验证入口，应补跑：

```bash
./scripts/check-repo.sh
```

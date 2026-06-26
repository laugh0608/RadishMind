# `Production Secret Backend Audit Store Contract / Event Schema Readiness` v1 计划

更新时间：2026-06-21

## 任务目标

本任务卡固定 `production-secret-backend-audit-store-contract-event-schema-readiness-v1` 的静态 readiness。目标是把 audit store handoff 与 audit store runtime entry review 中仍不够具体的 event schema、writer input/output、idempotency、delivery、retention 和 redaction contract 收束为可检查证据，并确认 audit store runtime implementation task card 仍未打开。

结论状态为 `audit_store_contract_event_schema_readiness_defined`。本批不解除 `audit_store_runtime_implementation_blocked_before_task_card`。

本批不创建 audit store runtime implementation task card，不创建 audit store runtime，不创建 audit writer，不写 audit event，不创建 runtime event schema，不访问 provider，不调用云 secret 服务，不读取真实 secret，不连接数据库，不创建 credential payload，不创建 credential handle runtime，不创建或执行 approval runtime，不创建 backend health runtime，不创建 no leakage smoke runtime，不实现 production resolver runtime，不启用 repository mode，不新增 public production API。

## 输入

- [Production Secret Backend Audit Store Handoff Readiness v1](../platform/production-secret-backend-audit-store-handoff-readiness-v1.md)
- [Production Secret Backend Audit Store Runtime Implementation Entry Review v1](../platform/production-secret-backend-audit-store-runtime-implementation-entry-review-v1.md)
- [Production Secret Backend Operator Approval Runtime Implementation Entry Review v1](../platform/production-secret-backend-operator-approval-runtime-implementation-entry-review-v1.md)
- [Production Secret Backend Credential Handle Runtime Implementation Entry Review v1](../platform/production-secret-backend-credential-handle-runtime-implementation-entry-review-v1.md)
- [Production Secret Backend Resolver Backend Health Runtime Implementation Entry Review v1](../platform/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.md)
- [Production Secret Backend Real Resolver No Secret Leakage Smoke Runtime Implementation Entry Review v1](../platform/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.md)
- `scripts/checks/fixtures/production-secret-backend-audit-store-handoff-readiness-v1.json`
- `scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-review-v1.json`
- `scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json`

## 交付物

1. 新增平台专题：
   - `docs/platform/production-secret-backend-audit-store-contract-event-schema-readiness-v1.md`
2. 新增 task card：
   - `docs/task-cards/production-secret-backend-audit-store-contract-event-schema-readiness-v1-plan.md`
3. 新增 fixture：
   - `scripts/checks/fixtures/production-secret-backend-audit-store-contract-event-schema-readiness-v1.json`
4. 新增 checker：
   - `scripts/check-production-ops-secret-backend-audit-store-contract-event-schema-readiness-v1.py`
5. 更新聚合与入口：
   - `docs/radishmind-current-focus.md`
   - `docs/features/README.md`
   - `docs/platform/README.md`
   - `docs/task-cards/README.md`
   - `docs/task-cards/production-secret-backend-implementation-v1-plan.md`
   - `scripts/README.md`
   - `scripts/check-repo.py`
6. 更新关联静态证据：
   - implementation readiness fixture 与 checker
   - 本周周志

## Contract Gate

| gate | 目标状态 |
| --- | --- |
| audit store contract readiness | `defined_without_store_runtime` |
| audit event schema | `static_contract_defined` |
| runtime event schema | `not_created` |
| runtime task card | `still_blocked_before_task_card` |
| audit store runtime | `not_created` |
| audit writer | `not_created` |
| audit event delivery | `not_executed` |
| idempotency key ref | `reference_required_before_runtime` |
| retention / redaction | `policy_binding_required_before_runtime` |
| operator approval runtime | `blocked_runtime_not_created` |
| credential handle runtime | `blocked_runtime_not_created` |
| backend health runtime | `blocked_runtime_not_created` |
| no leakage smoke runtime | `blocked_runtime_not_created` |
| production resolver runtime | `not_created` |
| cloud secret service | `forbidden` |
| database / repository / API | `blocked` |

## Event Schema Requirements

后续 future audit writer 至少必须使用：

- metadata-only event envelope
- event version
- event kind allowlist
- required / optional field allowlist
- forbidden field denylist
- reference-only secret ref、provider profile、backend profile、approval、credential boundary、health 和 smoke references
- idempotency key reference
- fail-closed delivery result envelope
- retention policy reference
- redaction profile reference
- sanitized diagnostic allowlist
- side effect counters

## 停止线

- 不创建 audit store runtime implementation task card。
- 不创建 audit store runtime、audit writer、audit event writer、durable audit backend 或 runtime event schema。
- 不写 audit event，不执行 delivery，不执行 store retention / redaction runtime。
- 不调用 provider、云 secret 服务、数据库、production API 或网络服务。
- 不读取真实 secret、developer env secret、credential payload、DSN、provider raw URL 或 backend endpoint URL。
- 不创建 credential payload、credential handle runtime、operator approval runtime、approval executor、backend health runtime、no leakage smoke runtime、production resolver runtime、DB provider、SQL migration、schema marker、repository mode runtime 或 public production API。
- 不把本 readiness 写成 audit store runtime ready、audit writer ready、audit event written、credential resolved、production resolver runtime ready、production secret backend ready 或 production API ready。

## 验证

本批至少执行：

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-contract-event-schema-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-handoff-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/check-repo.sh --fast
```

若触及 check-repo 注册、阶段边界或文档真相源，应补跑：

```bash
./scripts/check-repo.sh
```

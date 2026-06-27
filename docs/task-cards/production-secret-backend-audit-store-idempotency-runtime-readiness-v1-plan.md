# Production Secret Backend Audit Store Idempotency Runtime Readiness v1 计划

更新时间：2026-06-27

## 背景

`Production Secret Backend Audit Store Delivery Runtime Readiness v1` 已固定 future delivery runtime contract，但 audit store runtime task card 仍被 idempotency runtime、approval、credential handle、backend health 和 no leakage runtime 等依赖阻塞。

本批只推进 `production-secret-backend-audit-store-idempotency-runtime-readiness-v1`，把 future idempotency runtime 的职责边界、metadata-only input / result envelope、idempotency key reference policy、duplicate detection policy、replay decision policy、delivery runtime dependency、runtime event schema dependency、writer runtime dependency、durable backend dependency、禁止 secret material / payload、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard 固定为可检查证据。

结论状态为 `audit_store_idempotency_runtime_readiness_defined`。

## 范围

- 新增平台专题 `docs/platform/production-secret-backend-audit-store-idempotency-runtime-readiness-v1.md`。
- 新增 fixture `scripts/checks/fixtures/production-secret-backend-audit-store-idempotency-runtime-readiness-v1.json`。
- 新增 checker `scripts/check-production-ops-secret-backend-audit-store-idempotency-runtime-readiness-v1.py`。
- 同步 `production-ops-secret-backend-implementation-readiness.json`、implementation readiness checker 和 `scripts/check-repo.py`。
- 同步 current focus、platform 入口、features 入口、integration contracts、task card 入口、scripts 入口、implementation task card 和本周周志。

## Boundary Requirements

本批必须固定：

- idempotency runtime readiness 为 `defined_static_only`。
- idempotency runtime implementation task card 仍为 `not_created`。
- idempotency runtime、key store、duplicate detector、replay executor 和 retry runtime 仍为 `not_created`。
- idempotency owner 只负责 future idempotency decision owner reference，不负责 delivery runtime、writer runtime、runtime event schema materialization、durable backend selection 或 audit event write。
- idempotency input envelope 只能包含 reference、status、policy version 和 sanitized diagnostic。
- idempotency result reference 只定义 future result shape，不表示 duplicate decision 已持久化。
- idempotency key policy 只能引用 key ref，不派生 payload hash 或 secret-derived hash。
- duplicate detection policy 只能定义 fail-closed duplicate semantics，不执行 duplicate detection。
- replay decision policy 只能定义 fail-closed replay / suppression 语义，不执行 replay decision。
- delivery runtime、runtime event schema、writer runtime 和 durable backend 依赖必须保持静态边界状态，不能被 idempotency readiness 绕过。
- forbidden material 覆盖 raw secret、credential payload、provider raw URL、DSN、cloud credential、raw request / response / audit / writer / event / delivery / idempotency payload、payload hash、idempotency payload hash 和 secret-derived hash。
- failure mapping 必须 fail closed。
- sanitized diagnostics 只输出 status、failure code、failure boundary、request / audit reference 和 policy version。
- no fallback 禁止 memory store、fake resolver、developer env、fixture credential、mock provider、sample、历史证据、static boundary、delivery sample 或 duplicate sample 替代 idempotency runtime dependency。
- no side effects 必须保持所有 secret read、provider call、cloud call、DB call、audit write、delivery execution、idempotency decision、duplicate detection、runtime creation 和 repository enablement 计数为 0。
- artifact guard 只允许新增本批四个静态证据文件。

## 停止线

- 不创建 audit store runtime implementation task card。
- 不创建 idempotency runtime implementation task card。
- 不创建 delivery runtime implementation task card。
- 不创建 idempotency runtime、idempotency key store、duplicate detector、replay executor、retry executor 或 retry runtime。
- 不创建 runtime event schema implementation task card、runtime event schema artifact、runtime schema、writer type、writer runtime task card、writer runtime、audit store runtime、audit writer runtime 或 audit event writer。
- 不写 audit event，不创建 writer result fixture，不执行 delivery，不持久化 delivery result，不执行 duplicate detection，不持久化 duplicate decision，不执行 replay decision。
- 不选择或启用 durable audit backend。
- 不调用 provider、云 secret 服务、数据库、production API 或网络服务。
- 不读取真实 secret、developer env secret、credential payload、DSN、provider raw URL、backend endpoint URL、raw payload hash、raw idempotency payload、raw duplicate probe、raw replay payload、raw delivery payload、raw event payload 或 secret-derived hash。
- 不创建 production resolver runtime、cloud secret client、credential handle runtime、operator approval runtime、backend health runtime、no leakage smoke runtime、DB provider、SQL migration、schema marker、repository mode runtime 或 public production API。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-idempotency-runtime-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-delivery-runtime-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```

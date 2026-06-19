# Production Secret Backend Fake Resolver Implementation Task Card Entry Readiness Review v1

更新时间：2026-06-19

## 文档目的

本文档评审 production secret backend 在已具备 fake resolver static contract 与 no secret leakage smoke strategy 后，是否可以进入下一张 fake resolver implementation task card 的创建。

结论：状态为 `fake_resolver_implementation_task_card_entry_readiness_review_defined`，entry decision 为 `fake_resolver_implementation_task_card_ready_for_next_task`。这只表示下一步可以创建 fake resolver implementation task card；本批不创建 fake resolver runtime，不实现 resolver runtime，不解析 secret，不创建 credential handle，不连接数据库，不调用云 secret 服务，不启用 repository mode。

## 输入证据

- `production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1` 已把 test fixture strategy 与 fake resolver implementation entry 的 blocked 原因写成可检查证据。
- `production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1` 已固定 fake resolver static contract、no secret leakage smoke strategy、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard。
- `production-secret-backend-implementation-readiness` 仍显示 resolver implementation 为 `not_started`，fake resolver 为 `not_created`，test fixture strategy 为 `required_before_implementation`。
- `production-secret-reference-basic` 仍为 reference-only manifest，不保存 secret value，不启用 resolver，不允许 cloud calls。
- `workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1` 仍确认 saved draft database secret resolver implementation entry 不打开。

## Entry Readiness Decision

| gate | 本次结论 | 说明 |
| --- | --- | --- |
| fake resolver contract | `satisfied_static_contract` | 输入 / 输出 allowlist、禁止字段和失败语义已定义为静态证据 |
| no secret leakage smoke strategy | `satisfied_static_strategy` | 泄漏扫描策略、离线约束和 fail-closed 条件已定义，但 runtime smoke 未创建 |
| reference-only secret fixture | `satisfied_reference_only` | 只允许 placeholder secret ref key，不允许 secret value |
| implementation task card | `ready_for_next_task` | 下一步可以创建 fake resolver implementation task card |
| fake resolver runtime | `not_opened` | 本批不得新增 runtime type、resolver function 或 opaque credential handle runtime |
| no secret leakage smoke runtime | `not_opened` | 本批不得创建 smoke runner 或 runtime smoke fixture |
| resolver runtime | `forbidden` | production resolver runtime 仍不打开 |
| cloud secret service | `forbidden` | 不选择 vendor、不调用云 secret 服务 |
| database connection provider | `blocked` | DB provider、DB driver、connection factory、SQL 和 schema marker 均不打开 |
| repository mode runtime | `blocked` | workflow saved draft repository mode 仍不能通过 fake resolver 获得成功路径 |

## 下一张任务卡必须覆盖

下一张 fake resolver implementation task card 如果创建，必须只覆盖以下 implementation 入口要求：

- 创建 fake resolver implementation 的明确 scope、输入 / 输出 contract 和 disabled-by-default runtime gate。
- 创建 no secret leakage runtime smoke 的离线验证计划，但不得在任务卡创建评审内提前实现。
- 固定 placeholder secret ref fixture shape、environment binding、sanitized diagnostics emission boundary 和 artifact guard。
- 明确 fake resolver 只能服务测试与 fast baseline，不得被 production resolver、cloud secret backend、DB credential 或 repository mode 当作 fallback。
- 明确 side effect counters 必须保持零，直到独立 implementation 批次真正打开 runtime smoke。

下一张任务卡不得包含 resolver runtime、真实云 secret service、真实 credential、DB provider、DB driver、connection factory、SQL、schema marker、migration runner、repository mode、production secret audit store 或 public production API。

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `fake_resolver_task_card_contract_missing` | `task_card_entry` | 下一张任务卡缺少 fake resolver input / output contract 或禁止字段 |
| `fake_resolver_task_card_no_leakage_strategy_missing` | `task_card_entry` | 下一张任务卡缺少 no secret leakage runtime smoke 计划或离线边界 |
| `fake_resolver_task_card_runtime_created_in_entry_review` | `artifact_guard` | 本批评审直接新增 fake resolver runtime、smoke runner 或 credential handle runtime |
| `fake_resolver_task_card_secret_value_detected` | `artifact_guard` | 文档、fixture、checker 或 smoke record 出现 secret-looking value |
| `fake_resolver_task_card_cloud_call_forbidden` | `no_side_effects` | checker、fixture 或任务卡要求联网、provider call 或云 secret call |
| `fake_resolver_task_card_repository_mode_forbidden` | `repository_mode` | fake resolver 被用于打开 workflow saved draft repository mode 成功路径 |
| `fake_resolver_task_card_scope_overreach` | `implementation_boundary` | 下一张任务卡把 DB provider、resolver runtime、production API 或 audit store 合并进 fake resolver 实现 |

所有失败都必须 fail closed，不返回 secret value、DSN、provider raw URL、database hostname、database error detail、opaque credential handle、完整用户 claim 或草案主体。

## Sanitized Diagnostics

允许输出：

- `fake_resolver_task_card_entry_status`
- `fake_resolver_contract_status`
- `no_secret_leakage_smoke_strategy_status`
- `test_fixture_strategy_status`
- `resolver_state`
- `fake_resolver_runtime_status`
- `secret_ref_status`
- `environment_binding_status`
- `failure_code`
- `sanitized_diagnostic`
- `request_id`
- `audit_ref`
- `policy_version`

禁止输出 raw secret、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail、opaque credential handle、resolver backend URL、full secret ref value 或完整用户 claim。本批不允许 runtime emission；所有 diagnostics 只存在于 committed fixture / checker 的静态评审语义中。

## No Fallback

- 不允许 fake resolver task card fallback 到 production resolver。
- 不允许 production resolver fallback 到 fake resolver。
- 不允许 test secret ref fallback 到 production secret ref。
- 不允许 resolver failure fallback 到 committed value、developer env plaintext、fixture credential、mock provider、local-smoke profile、fake query executor 或 sample。
- 不允许 fake resolver task card readiness 被解释为 credential resolved、production secret backend ready、cloud secret backend ready、repository mode ready 或 production ready。

## No Side Effects

本批 checker 只读取 committed 文档、fixture、schema、已有 readiness fixture 和 `check-repo.py` 注册顺序，不读取真实环境变量、不连接网络、不调用云 secret 服务、不访问 provider、不调用 fake resolver、不调用 resolver、不创建 credential handle、不连接数据库、不打开 driver、不运行 SQL、不读写 schema marker、不启用 repository mode、不调用 production API。

side effect counters 必须保持：

- `secret_resolver_call_count=0`
- `fake_resolver_call_count=0`
- `secret_value_read_count=0`
- `cloud_secret_call_count=0`
- `provider_call_count=0`
- `credential_handle_created_count=0`
- `database_connection_count=0`
- `driver_open_count=0`
- `sql_execution_count=0`
- `schema_marker_read_count=0`
- `schema_marker_write_count=0`
- `repository_mode_enablement_count=0`
- `production_api_call_count=0`

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.md`
- `docs/task-cards/production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.json`
- `scripts/check-production-ops-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.py`

不得新增或启用以下 artifact：

- fake resolver implementation task card
- resolver runtime
- fake resolver runtime
- no secret leakage smoke runtime
- no secret leakage smoke fixture
- cloud secret SDK
- production credential file
- secret value fixture
- opaque credential handle runtime
- database connection provider
- DB driver / DSN parser
- connection factory
- SQL migration
- schema marker reader / writer
- migration runner
- workflow saved draft repository mode runtime
- production secret audit store
- public production API

## 后续推进

下一步可以创建 `Production Secret Backend Fake Resolver Implementation v1` 任务卡，但必须先把本评审作为输入证据。该任务卡创建后仍需要独立评审 runtime 实现边界与 no secret leakage smoke runtime，不能把 resolver runtime、DB provider、cloud secret service 或 repository mode 合并进同一批次。

`test-fixture-strategy` 在本批后仍为 `required_before_implementation`；只有 fake resolver implementation task card、runtime smoke、sanitized diagnostics runtime 和 artifact guard 都形成可复验实现后，才能重新评审是否从 blocked 变为 satisfied。

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-contract.py
./scripts/run-python.sh scripts/check-production-secret-reference-contract.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-secret-resolver-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```

# Saved Workflow Draft Schema Marker Runtime Dependency Refresh v1

更新时间：2026-06-23

## 专题定位

`Saved Workflow Draft Schema Marker Runtime Dependency Refresh v1` 承接 schema marker contract entry review、manual migration runner entry refresh、database connection provider entry refresh v2、Radish OIDC upstream evidence refresh 和 repository mode runtime boundary review，用于把 future schema marker runtime 之前的依赖矩阵重新收束。

结论：状态为 `draft_schema_marker_runtime_dependency_refresh_defined`。entry decision 为 `schema_marker_runtime_still_blocked_before_implementation_task_card`。本批只固定 marker reader / writer、manual runner、connection provider、connection lifecycle、repository mode、auth / membership、idempotency / lock、rollback observability 和 negative marker smoke 的静态依赖关系；不创建 schema marker runtime implementation task card、schema marker reader / writer、schema version table、SQL migration、manual migration runner、database connection provider、repository mode runtime、OIDC middleware、token validation、membership adapter 或 production API。

## 输入证据

- `workflow-saved-draft-schema-marker-contract-implementation-entry-review-v1` 已确认 marker reader / writer contract implementation task card 仍为 blocked。
- `workflow-saved-draft-manual-migration-runner-implementation-entry-refresh-v1` 已确认 runner command、dry-run、apply result、idempotency / lock 和 rollback observability 仍不足以打开 runner task card。
- `workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v2` 已确认 driver / DSN / TLS、role policy、connection smoke strategy 和 lifecycle readiness 只满足静态前置，connection provider task card 仍 blocked。
- `workflow-saved-draft-database-connection-lifecycle-readiness-v1` 与 `workflow-saved-draft-database-connection-smoke-strategy-v1` 只定义 future lifecycle / smoke contract，不创建 runtime。
- `radish-oidc-token-membership-upstream-evidence-refresh-v1` 已补 reviewed issuer、JWKS、client registration、auth middleware ownership、membership source ownership 和 negative auth smoke matrix 的静态契约，但不创建 auth runtime。
- `workflow-saved-draft-repository-mode-runtime-boundary-review-v1` 已确认 repository mode runtime task card 仍 blocked。

## Dependency Refresh Decision

| dependency | 本次结论 | 说明 |
| --- | --- | --- |
| marker contract shape | `static_contract_defined_no_runtime` | marker 状态与失败码已固定，但 reader / writer 未创建 |
| applied marker source | `static_contract_defined_no_marker_table` | future schema version table 名称可引用，但没有 SQL 或 table |
| marker read path | `blocked_no_marker_reader` | repository preflight 仍不能读取 marker |
| marker write path | `blocked_no_marker_writer` | writer 只能归 future manual runner 持有 |
| manual runner handoff | `blocked_no_runner_runtime` | runner command、dry-run、apply result 和 lock runtime 未创建 |
| database connection dependency | `blocked_no_connection_provider` | connection provider、secret resolver、driver runtime 和 query executor 均未创建 |
| connection lifecycle dependency | `static_readiness_defined_no_runtime` | timeout / pool / health / close / diagnostics 仍为静态前置 |
| connection smoke dependency | `static_strategy_defined_no_runtime_smoke` | smoke strategy 已有，但没有 smoke runner / output |
| repository mode dependency | `blocked_no_repository_mode_runtime` | repository mode 仍 fail closed |
| auth membership dependency | `static_upstream_evidence_defined_no_auth_runtime` | OIDC upstream evidence 已有，但 token validation / membership runtime 未创建 |
| idempotency lock dependency | `static_contract_defined_no_runtime` | duplicate / lock 语义已定义，未实现 runtime |
| rollback observability dependency | `static_contract_defined_no_runtime` | rollback 仍是静态 evidence |

## Runtime Blockers

- `schema_marker_reader_not_created`
- `schema_marker_writer_not_created`
- `schema_version_table_not_created`
- `manual_migration_runner_not_created`
- `database_connection_provider_not_created`
- `repository_mode_runtime_disabled`
- `token_validation_schema_missing`
- `auth_middleware_not_created`
- `membership_adapter_not_created`
- `connection_smoke_runtime_missing`

## Future Task Requirements

后续若要创建 schema marker runtime implementation task card，必须先独立补齐：

1. marker read / write interface owner、输入输出和 sanitized failure envelope。
2. applied marker table 或 schema version table 的 committed SQL / 等价 apply plan。
3. marker read preflight 与 repository query execution 的 failure mapping。
4. marker write ownership：只能由 manual runner 持有，startup / route / selector 不得写 marker。
5. manual runner handoff、dry-run、apply result envelope、idempotency key、lock 和 duplicate handling。
6. rollback / forward-only observability 与 evidence owner。
7. connection provider、query role、connection lifecycle 和 smoke runtime 的可复验证据。
8. token validation、auth middleware、membership context 和 negative auth smoke runtime 的可复验证据。
9. marker missing、version mismatch、unavailable、connection unavailable 和 auth context mismatch 的负向 smoke matrix。

## Failure Mapping

| 条件 | failure code |
| --- | --- |
| repository mode reserved / runtime 未打开 | `repository_store_disabled` |
| unknown store mode | `invalid_draft_store_mode` |
| marker 缺失或 migration 未应用 | `draft_schema_migration_not_applied` |
| marker version mismatch | `draft_store_schema_version_mismatch` |
| marker 不可读、runner 状态不可观测或 lock 状态不可信 | `draft_store_migration_unavailable` |
| connection provider、secret resolver、driver、lifecycle 或 smoke runtime 不可用 | `draft_store_unavailable` |
| request id / audit ref / purpose / policy version 缺失 | `draft_audit_context_missing` |
| stored record contract mismatch | `draft_store_contract_mismatch` |
| repository actor context 不可信 | `draft_auth_context_contract_mismatch` |

失败时不得返回草案主体、SQL detail、DSN、secret value、database host、raw database error、完整 claim、raw request payload 或 audit payload；不得回退 `memory_dev`、sample、fixture、dev route、test auth、fake query executor 或 test-only fake resolver runtime。

## No Fallback / No Side Effects

- 不允许 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 因本 refresh 被启用。
- 不允许 service startup、HTTP route、store selector 或 saved draft request path 创建 marker、运行 migration、打开数据库连接、读取 secret、校验 OIDC token 或查 membership。
- 不允许 static schema artifact、fake query executor smoke、connection smoke strategy、connection lifecycle readiness、OIDC upstream evidence 或 rollback evidence 替代真实 applied marker runtime。
- 本批 checker 只读取 committed 文档、fixture、源码和 `check-repo.py` 注册顺序；不启动服务、不读取 env secret、不调用 fake resolver、不调用 production resolver、不调用云 secret 服务、不连接数据库、不导入 driver、不运行 SQL、不读写 schema marker、不创建 runner、不调用 OIDC、不校验 token、不创建 membership adapter、不创建 production API、不执行 workflow。
- side effect counters 必须保持 `schema_marker_read_count=0`、`schema_marker_write_count=0`、`schema_version_table_create_count=0`、`migration_runner_call_count=0`、`runner_command_count=0`、`sql_execution_count=0`、`database_connection_count=0`、`connection_provider_call_count=0`、`repository_mode_enablement_count=0`、`token_validation_count=0`、`membership_lookup_count=0`、`production_api_call_count=0`、`executor_call_count=0`、`confirmation_call_count=0`、`business_writeback_count=0` 和 `replay_call_count=0`。

## 后续方向

本次 refresh 后，schema marker runtime implementation task card、manual migration runner implementation task card、connection provider implementation task card 和 repository mode runtime task card 仍不创建。后续若继续 durable store 上游，应从以下方向选择一个独立推进：

1. `secret resolver runtime dependency refresh`：复评 production resolver、credential handle、approval、audit、backend health 和 no leakage runtime 是否足以打开 resolver task card。
2. `token validation schema / auth middleware runtime entry review`：消费 upstream evidence，评审 token schema、middleware、membership adapter 和 runtime smoke 是否可拆 task card。
3. `manual migration runner runtime dependency refresh`：在 runner command / dry-run / apply result / lock / rollback evidence 可复验后重新评审。

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-marker-runtime-dependency-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v2.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-marker-contract-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-radish-oidc-token-membership-upstream-evidence-refresh-v1.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 `docs/task-cards/workflow-saved-draft-schema-marker-implementation-v1-plan.md` 或 `docs/task-cards/workflow-saved-draft-schema-marker-contract-implementation-v1-plan.md`。
- 不创建 SQL migration、schema version table、schema marker reader / writer、migration runner、runner command、dry-run output、database query executor、database connection provider、DB driver、connection smoke runner 或 lifecycle runtime。
- 不启用 repository mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
- 不把 `draft_schema_marker_runtime_dependency_refresh_defined` 解释为 schema marker ready、migration ready、database ready、repository mode ready、durable persistence ready、production API ready 或 production ready。

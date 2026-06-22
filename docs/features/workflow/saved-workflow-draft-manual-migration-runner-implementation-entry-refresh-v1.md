# Saved Workflow Draft Manual Migration Runner Implementation Entry Refresh v1

更新时间：2026-06-22

## 专题定位

`Saved Workflow Draft Manual Migration Runner Implementation Entry Refresh v1` 承接 [Saved Workflow Draft Schema Marker Contract Implementation Entry Review v1](saved-workflow-draft-schema-marker-contract-implementation-entry-review-v1.md)，用于复评 future manual migration runner 是否可以进入实现任务卡。

结论：状态为 `draft_manual_migration_runner_implementation_entry_refresh_defined`。entry decision 仍为 `blocked_before_implementation_task_card`。本批只固定 migration artifact / apply plan、dry-run output、apply result envelope、idempotency / lock、rollback observability 和 implementation blocker 的刷新结论；不创建 migration runner implementation task card、runner runtime、runner command、dry-run output、SQL migration、schema version table、schema marker reader / writer、database connection provider、query executor 或 repository mode runtime。

## 输入证据

- `workflow-saved-draft-schema-marker-contract-implementation-entry-review-v1` 已确认 marker reader / writer contract implementation task card 当前仍为 blocked。
- `workflow-saved-draft-schema-marker-migration-runner-readiness-refresh-v1` 已固定 manual runner boundary、dry-run output shape、apply result envelope、idempotency key owner、duplicate handling 和 rollback observability 的静态 readiness。
- `workflow-saved-draft-schema-migration-runner-implementation-entry-review-v1` 已确认 executable migration artifact、schema version marker contract、manual runner command、migration apply smoke 和 rollback observability 仍为 blocked。
- `workflow-saved-draft-schema-artifact-materialization-v1` 已物化静态 schema artifact，但没有可执行 SQL 或等价 apply plan。
- `workflow-saved-draft-database-connection-provider-implementation-entry-review-v1` 与 `workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1` 已确认 connection provider 和 secret resolver implementation 仍不打开。
- `workflow-saved-draft-repository-mode-runtime-boundary-review-v1` 已确认 repository mode runtime task card 仍 blocked。

## Entry Refresh Decision

| candidate | 本次结论 | 阻塞原因 |
| --- | --- | --- |
| executable migration artifact / apply plan | `blocked` | 当前只有静态 schema artifact 和 DDL review，没有 committed SQL 或等价 apply plan |
| dry-run output contract | `defined_static_only` | output shape 已定义，但没有 runner command、database role、connection provider 或 dry-run artifact |
| apply result envelope | `defined_static_only` | allowed result 已定义，但没有 apply command、marker writer、lock runtime 或 applied-marker smoke |
| idempotency / lock policy | `blocked` | idempotency key owner 已定义为 future runner，但没有 duplicate detector、lock runtime 或 conflict smoke |
| rollback / forward-only observability | `blocked` | rollback 仍是静态 evidence，没有 rollback command、backup policy、restore verification 或 forward-only exception record |
| manual migration runner implementation task card | `blocked_before_implementation_task_card` | artifact、connection、marker writer、lock 和 rollback 依赖未满足 |

后续若要打开 manual migration runner implementation task card，必须先独立定义 committed migration artifact 或 apply plan、runner command ownership、dry-run artifact、database connection provider dependency、migration role policy、marker writer handoff、lock / idempotency policy、apply smoke 和 rollback / forward-only observability。

## Runner Contract Refresh

future manual migration runner 必须保持以下边界：

- 只能由显式运维入口触发，不能由 service startup、HTTP route、store selector 或 saved draft request path 自动触发。
- dry-run 必须只输出 metadata：artifact id、target schema version、environment、database role、lock policy、expected marker transition、rollback / forward-only policy 和 no side effects summary。
- apply result 只能返回 `applied`、`already_applied`、`blocked_lock_conflict`、`blocked_version_mismatch`、`blocked_connection_unavailable` 或 `failed_rollback_unavailable`。
- same artifact + same target version 可以返回 `already_applied` metadata；不同 artifact 或 version 必须 fail closed。
- repository adapter 只能消费 applied marker，不能拥有 migration apply idempotency key，也不能在 marker 缺失时运行 migration。
- runner failure 不得返回 SQL detail、DSN、secret value、database host、raw database error、raw request payload、raw response payload 或 audit payload。

## Failure Mapping

implementation entry refresh 继续固定 fail-closed 语义：

| 条件 | failure code |
| --- | --- |
| repository mode reserved / runtime 未打开 | `repository_store_disabled` |
| unknown store mode | `invalid_draft_store_mode` |
| marker 缺失或 migration 未应用 | `draft_schema_migration_not_applied` |
| marker version mismatch | `draft_store_schema_version_mismatch` |
| marker 不可读、runner 状态不可观测或 lock 状态不可信 | `draft_store_migration_unavailable` |
| database connection / query executor 不可用 | `draft_store_unavailable` |
| stored record contract mismatch | `draft_store_contract_mismatch` |
| repository actor context 不可信 | `draft_auth_context_contract_mismatch` |

## No Fallback / No Side Effects

- 不允许 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 因本 refresh 被启用。
- 不允许 service startup、HTTP route、store selector 或 saved draft request path 自动运行 migration、创建 marker、打开数据库连接或写 repository。
- 不允许 static schema artifact、fake query executor smoke、rollback evidence 或 container smoke 替代真实 migration apply smoke。
- 不允许失败后回退到 `memory_dev`、sample、fixture、dev HTTP route、test auth、fake query executor 或 test-only fake resolver runtime。
- 本批 checker 只读取 committed 文档、fixture、源码和 `check-repo.py` 注册顺序；不启动服务、不读取 env secret、不调用云 secret 服务、不连接数据库、不导入 driver、不运行 SQL、不读写 schema marker、不创建 runner、不调用 OIDC、不校验 token、不创建 membership adapter、不创建 production API、不执行 workflow。
- side effect counters 必须保持 `database_connection_count=0`、`driver_open_count=0`、`sql_execution_count=0`、`schema_marker_read_count=0`、`schema_marker_write_count=0`、`migration_runner_invocation_count=0`、`dry_run_output_count=0`、`repository_mode_enablement_count=0`、`repository_write_count=0`、`oidc_call_count=0`、`token_validation_count=0`、`membership_lookup_count=0`、`production_api_call_count=0`、`executor_call_count=0`、`confirmation_call_count=0`、`business_writeback_count=0` 和 `replay_call_count=0`。

## 后续方向

本次 refresh 后，下一步可以在以下方向中选择一个独立推进：

1. `database connection provider implementation entry refresh`：复评 secret resolver、driver / DSN / TLS policy、role policy 和 connection smoke。
2. `Radish OIDC token / membership upstream evidence refresh`：等待 reviewed issuer、JWKS、client registration 和 membership data source ownership 后再复评。
3. `production API consumer readiness`：只有 repository mode、auth、database 和 smoke 都可复验后才重新评审。

如果上述依赖仍 blocked，manual migration runner implementation task card、schema marker contract implementation task card 和 repository mode runtime implementation task card 继续保持不创建。

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-manual-migration-runner-implementation-entry-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-marker-contract-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-marker-migration-runner-readiness-refresh-v1.py
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 `docs/task-cards/workflow-saved-draft-schema-migration-runner-implementation-v1-plan.md`。
- 不创建 SQL migration、schema version table、schema marker reader / writer、migration runner、runner command、dry-run output、database query executor、database connection provider、DB driver 或 connection smoke。
- 不启用 repository mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
- 不把 `draft_manual_migration_runner_implementation_entry_refresh_defined` 解释为 runner ready、migration ready、database ready、repository mode ready、durable persistence ready、production API ready 或 production ready。

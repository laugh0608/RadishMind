# Saved Workflow Draft Schema Marker Contract Implementation Entry Review v1

更新时间：2026-06-22

## 专题定位

`Saved Workflow Draft Schema Marker Contract Implementation Entry Review v1` 承接 [Saved Workflow Draft Schema Marker / Migration Runner Readiness Refresh v1](saved-workflow-draft-schema-marker-migration-runner-readiness-refresh-v1.md)，用于评审 future schema marker reader / writer contract 是否可以进入实现任务卡。

结论：状态为 `draft_schema_marker_contract_implementation_entry_review_defined`。entry decision 为 `blocked_before_implementation_task_card`。本批只固定 implementation entry review、候选阻塞项、failure mapping、no fallback、no side effects 和 artifact guard；不创建 schema marker implementation task card、schema marker runtime、schema version table、marker reader / writer、migration runner、SQL、database connection provider、query executor 或 repository mode runtime。

## 输入证据

- `workflow-saved-draft-schema-marker-migration-runner-readiness-refresh-v1` 已固定 applied marker read / write contract、manual runner boundary、dry-run output shape、idempotency / lock、duplicate handling 和 rollback observability 的静态 readiness。
- `workflow-saved-draft-schema-migration-runner-implementation-entry-review-v1` 已确认 executable migration artifact、schema version marker contract、manual runner command、dry-run plan、migration apply smoke 和 rollback observability 仍为 blocked。
- `workflow-saved-draft-database-connection-schema-marker-preconditions-v1` 已定义 future database connection provider、secret ref、query role、environment isolation 和 schema marker read / write contract，但不创建 runtime。
- `workflow-saved-draft-database-connection-provider-implementation-entry-review-v1` 与 `workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1` 已确认 connection provider 和 secret resolver implementation 仍不打开。
- `workflow-saved-draft-repository-mode-runtime-boundary-review-v1` 已确认 repository mode runtime task card 仍 blocked。

## Entry Decision

| candidate | 本次结论 | 阻塞原因 |
| --- | --- | --- |
| schema marker reader contract task card | `blocked` | marker store ownership、connection provider、schema version table 和 sanitized marker read failure envelope 未落地 |
| schema marker writer contract task card | `blocked` | writer 只能由 future manual runner 持有，但 runner command、lock / idempotency policy 和 apply result envelope 仍未实现 |
| schema marker table / record shape | `blocked` | 当前只有静态 schema artifact 和 refresh contract，没有 committed SQL 或等价 apply plan |
| repository preflight consumption binding | `blocked` | repository mode runtime 未打开，真实 query executor 和 marker reader 均未创建 |
| negative marker smoke | `blocked` | 没有测试数据库、marker runtime、connection smoke 或 applied-marker smoke |
| schema marker contract implementation task card | `blocked_before_implementation_task_card` | 上述依赖未满足，不能创建 implementation task card |

后续若要打开 schema marker contract implementation task card，必须先独立定义 marker store ownership、connection provider contract、manual runner writer ownership、schema marker record shape、negative marker smoke matrix 和 sanitized diagnostics。

## Contract Boundary

future schema marker contract 必须保持以下边界：

- reader 只读 applied marker，不自动创建 marker、运行 migration 或修复版本状态。
- writer 只能由 future manual migration runner 持有，service startup、HTTP route、store selector 和 saved draft request path 不得写 marker。
- marker 状态只能返回 `applied`、`not_applied`、`version_mismatch` 或 `unavailable`。
- `not_applied` 映射为 `draft_schema_migration_not_applied`。
- `version_mismatch` 映射为 `draft_store_schema_version_mismatch`。
- `unavailable` 映射为 `draft_store_migration_unavailable`。
- marker failure 必须阻止 repository query execution，且不得返回草案主体、SQL detail、DSN、database host、raw database error、secret value、raw claim、raw request payload 或 audit payload。

## Failure Mapping

implementation entry review 继续固定 fail-closed 语义：

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

- 不允许 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 因本 entry review 被启用。
- 不允许 service startup、HTTP route、store selector 或 saved draft request path 自动创建 marker、运行 migration、打开数据库连接或写 schema marker。
- 不允许 static schema artifact、fake query executor smoke、rollback evidence 或 container smoke 替代真实 applied marker。
- 不允许失败后回退到 `memory_dev`、sample、fixture、dev HTTP route、test auth、fake query executor 或 test-only fake resolver runtime。
- 本批 checker 只读取 committed 文档、fixture、源码和 `check-repo.py` 注册顺序；不启动服务、不读取 env secret、不调用云 secret 服务、不连接数据库、不导入 driver、不运行 SQL、不读写 schema marker、不创建 runner、不调用 OIDC、不校验 token、不创建 membership adapter、不创建 production API、不执行 workflow。
- side effect counters 必须保持 `database_connection_count=0`、`driver_open_count=0`、`sql_execution_count=0`、`schema_marker_read_count=0`、`schema_marker_write_count=0`、`migration_runner_invocation_count=0`、`repository_mode_enablement_count=0`、`repository_write_count=0`、`oidc_call_count=0`、`token_validation_count=0`、`membership_lookup_count=0`、`production_api_call_count=0`、`executor_call_count=0`、`confirmation_call_count=0`、`business_writeback_count=0` 和 `replay_call_count=0`。

## 后续方向

本次 entry review 后，下一步可以在以下方向中选择一个独立推进：

1. `manual migration runner implementation entry refresh`：复评 migration artifact、dry-run、apply result、idempotency、lock 和 rollback evidence 是否足以进入任务卡。
2. `database connection provider implementation entry refresh`：复评 secret resolver、driver / DSN / TLS policy、role policy 和 connection smoke。
3. `Radish OIDC token / membership upstream evidence refresh` 已固定为 `radish_oidc_token_membership_upstream_evidence_refresh_defined`；后续 auth runtime 仍需独立 entry review。

如果上述依赖仍 blocked，schema marker contract implementation task card 和 repository mode runtime implementation task card 继续保持不创建。

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-marker-contract-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-marker-migration-runner-readiness-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-mode-runtime-boundary-review-v1.py
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 `docs/task-cards/workflow-saved-draft-schema-marker-contract-implementation-v1-plan.md`。
- 不创建 SQL migration、schema version table、schema marker reader / writer、migration runner、runner command、dry-run output、database query executor、database connection provider、DB driver 或 connection smoke。
- 不启用 repository mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
- 不把 `draft_schema_marker_contract_implementation_entry_review_defined` 解释为 schema marker ready、migration ready、database ready、repository mode ready、durable persistence ready、production API ready 或 production ready。

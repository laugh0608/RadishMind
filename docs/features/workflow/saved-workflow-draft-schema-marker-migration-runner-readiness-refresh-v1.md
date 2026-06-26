# Saved Workflow Draft Schema Marker / Migration Runner Readiness Refresh v1

更新时间：2026-06-21

## 专题定位

`Saved Workflow Draft Schema Marker / Migration Runner Readiness Refresh v1` 承接 schema migration runner readiness、runner implementation entry review、database connection / schema marker preconditions 和 repository mode runtime boundary review，用于把 future schema marker contract 与 manual migration runner 的下一批前置边界重新收束。

结论：状态为 `draft_schema_marker_migration_runner_readiness_refresh_defined`。本批只固定 applied marker read / write contract、manual runner boundary、dry-run output shape、idempotency / lock / duplicate apply failure、rollback observability 和 repository preflight consumption 的静态前置；不创建 schema marker implementation task card、migration runner implementation task card、SQL migration、schema version table、runner command、database connection provider、database query executor 或 repository mode runtime。

## 输入证据

- `workflow-saved-draft-schema-migration-runner-readiness-v1` 已定义 manual runner boundary、config gate、schema preflight、failure mapping 和 rollback 停止线。
- `workflow-saved-draft-schema-migration-runner-implementation-entry-review-v1` 已确认 executable migration artifact、schema version marker contract、manual runner command、dry-run plan、migration apply smoke 和 rollback observability 仍为 blocked。
- `workflow-saved-draft-database-connection-schema-marker-preconditions-v1` 已定义 future database connection provider、secret ref、query role、environment isolation 和 schema marker read / write contract，但不创建 runtime。
- `workflow-saved-draft-database-connection-provider-implementation-entry-review-v1` 与 `workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1` 已确认 connection provider 和 secret resolver implementation 仍不打开。
- `workflow-saved-draft-repository-mode-runtime-boundary-review-v1` 已确认 repository mode runtime task card 仍 blocked。
- repository adapter、adapter smoke 和 production auth runtime bridge 已可复验，但仍不能替代真实 migration runner、schema marker、database connection provider、OIDC middleware、token validation 或 membership adapter。

## Refresh Boundary

| gate | 当前结论 | 说明 |
| --- | --- | --- |
| schema marker owner | `static_contract_owner_defined` | owner 只负责 future marker contract，不创建 marker runtime |
| applied marker read semantics | `defined_static` | 缺失、版本不匹配、不可读都必须阻止 query execution |
| applied marker write semantics | `manual_runner_only_static` | 只有 future manual runner 可写 marker，startup / route / selector 不可写 |
| manual migration runner boundary | `refreshed_manual_only` | runner 只能由显式运维入口触发 |
| dry-run output shape | `defined_static` | 输出 artifact、target schema version、lock policy、rollback policy 和 side effect preview |
| idempotency / lock semantics | `defined_static_fail_closed` | duplicate apply、concurrent apply 和 lock failure 必须 fail closed 或返回 already-applied metadata |
| rollback observability | `static_evidence_only` | rollback 仍是静态 evidence，不创建 rollback command |
| database connection provider | `blocked` | secret resolver、driver policy、role policy、connection smoke 未满足 |
| repository mode runtime task card | `blocked_before_task_card` | 本批不创建 runtime task card |

## Schema Marker Contract Refresh

future schema marker contract 必须回答以下状态：

- `applied`：store schema version 与 `saved_workflow_drafts_store_v1` 一致，marker 可读，runner record 指向已审核 artifact。
- `not_applied`：marker 缺失或目标 version 未出现，映射为 `draft_schema_migration_not_applied`。
- `version_mismatch`：marker version 与 adapter preflight 期望不一致，映射为 `draft_store_schema_version_mismatch`。
- `unavailable`：marker 不可读、marker store 不可观测、lock 状态不可信或 runner record 不可信，映射为 `draft_store_migration_unavailable`。

runtime repository preflight 只能读取 marker；只有 manual runner 可写 marker。service startup、HTTP route、store selector 和 saved draft request path 都不能写 marker，也不能在 marker 缺失时自动运行 migration。

## Manual Runner Readiness Refresh

后续若进入 runner implementation task card，必须先明确：

1. migration artifact 或等价 apply plan 的 committed 形态。
2. dry-run output：artifact id、target schema version、environment、database role、lock policy、expected marker transition、rollback / forward-only policy、no side effects summary。
3. apply result envelope：`applied`、`already_applied`、`blocked_lock_conflict`、`blocked_version_mismatch`、`blocked_connection_unavailable`、`failed_rollback_unavailable`。
4. idempotency key owner：future runner owns apply idempotency key，repository adapter 只消费 applied marker。
5. duplicate handling：same artifact + same target version 可返回 `already_applied` metadata；不同 artifact 或 version 必须 fail closed。
6. rollback observability：rollback command 或 forward-only exception 必须在独立实现批次中可复验。

## Failure Mapping

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

失败时不得返回草案主体、SQL detail、DSN、secret value、database host、raw database error、完整 claim、raw request payload 或 audit payload；不得回退 `memory_dev`、sample、fixture、dev route、test auth、fake query executor 或 test-only fake resolver runtime。

## No Fallback / No Side Effects

- 不允许 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 因 marker refresh 被启用。
- 不允许 service startup、HTTP route、store selector 或 saved draft request path 自动创建 marker、运行 migration 或打开数据库连接。
- 不允许 fake query executor smoke、static schema artifact、rollback evidence 或 container smoke 替代真实 applied marker。
- 本批 checker 只读取 committed 文档、fixture、源码和 `check-repo.py` 注册顺序；不启动服务、不读取 env secret、不调用云 secret 服务、不连接数据库、不导入 driver、不运行 SQL、不读写 schema marker、不创建 runner、不调用 OIDC、不校验 token、不创建 membership adapter、不创建 production API、不执行 workflow。
- side effect counters 必须保持 `database_connection_count=0`、`driver_open_count=0`、`sql_execution_count=0`、`schema_marker_read_count=0`、`schema_marker_write_count=0`、`migration_runner_invocation_count=0`、`repository_mode_enablement_count=0`、`repository_write_count=0`、`oidc_call_count=0`、`token_validation_count=0`、`membership_lookup_count=0`、`production_api_call_count=0`、`executor_call_count=0`、`confirmation_call_count=0`、`business_writeback_count=0`、`replay_call_count=0`。

## 后续方向

本次 refresh 后，下一步可以在以下方向中选择一个独立推进：

1. `schema marker contract implementation entry review`：只评审 marker reader / writer contract 是否可以创建任务卡，不实现 marker runtime。
2. `manual migration runner implementation entry refresh`：只复评 migration artifact、dry-run、apply result、idempotency、lock 和 rollback evidence 是否足以进入任务卡。
3. `database connection provider implementation entry refresh`：先复评 secret resolver、driver / DSN / TLS policy、role policy 和 connection smoke。
4. `Radish OIDC token / membership readiness`：先定义真实 auth middleware 和 membership adapter 前置。

如果上述依赖仍 blocked，repository mode runtime implementation task card 继续保持不创建。

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-marker-migration-runner-readiness-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-mode-runtime-boundary-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-schema-marker-preconditions-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-migration-runner-implementation-entry-review-v1.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```

## 停止线

- 不创建 `docs/task-cards/workflow-saved-draft-schema-marker-contract-implementation-v1-plan.md` 或 `docs/task-cards/workflow-saved-draft-schema-migration-runner-implementation-v1-plan.md`。
- 不创建 SQL migration、schema version table、schema marker reader / writer、migration runner、runner command、dry-run output、database query executor、database connection provider、DB driver 或 connection smoke。
- 不启用 repository mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。

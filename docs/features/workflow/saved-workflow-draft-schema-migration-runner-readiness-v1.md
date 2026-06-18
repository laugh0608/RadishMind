# Saved Workflow Draft Schema Migration Runner Readiness v1

更新时间：2026-06-18

## 专题定位

`Saved Workflow Draft Schema Migration Runner Readiness v1` 承接 [Saved Workflow Draft Repository Mode Enablement v1](saved-workflow-draft-repository-mode-enablement-v1.md)，用于定义 future schema migration runner 的实现准入边界。

结论：状态为 `draft_schema_migration_runner_readiness_defined`。当前只固定 runner boundary、config gate、schema preflight、adapter smoke / production auth runtime dependency、failure mapping、rollback、no fallback 和 no side effects；不创建 runner、不写 SQL、不连接数据库，也不启用 repository mode。

## 输入证据

- `workflow-saved-draft-schema-migration-preconditions-v1`：已定义 logical schema、index strategy、migration gate 和 failure mapping。
- `workflow-saved-draft-schema-artifact-materialization-v1`：已物化 `manifest.json`、`ddl-review.md`、`rollback-evidence.json` 和 `migration-smoke.json` 静态证据，但没有 SQL migration、schema version table 或 runner。
- `workflow-saved-draft-repository-adapter-implementation-v1`：adapter 已实现 schema preflight 和 store schema version failure mapping。
- `workflow-saved-draft-adapter-smoke-v1`：adapter smoke 仍使用 injected fake query executor，不代表真实 DB smoke。
- `workflow-saved-draft-production-auth-runtime-v1`：已实现 verified auth context 到 repository actor context 的投影，但没有 OIDC middleware、token validation 或 membership adapter。
- `workflow-saved-draft-repository-mode-enablement-v1`：已评审 repository mode enablement，结论仍是不启用 repository mode。

## Runner Boundary

future schema migration runner 必须是显式人工触发的离线 / 运维入口，不允许由 platform service startup、HTTP route、store selector 或 saved draft request path 自动运行。

runner 进入实现前必须先有：

1. 可执行 migration artifact 或等价 schema apply plan，且与 `manifest.json` 的 `store_schema_version` 一致。
2. schema version / applied migration marker 的读写契约。
3. dry-run / plan 输出，能说明即将 apply 的 artifact、目标 schema version、lock policy 和 rollback policy。
4. apply 成功、已应用、版本不匹配、migration 状态不可读和 rollback 不可用的 failure mapping。
5. migration smoke，覆盖 schema version、tenant / workspace / application predicate、owner list projection、version conflict predicate 和 no side effects。

当前 `services/platform/migrations/workflow_saved_drafts/` 只能保留四个静态证据文件：`manifest.json`、`ddl-review.md`、`rollback-evidence.json` 和 `migration-smoke.json`。任何 `.sql`、runner Go 文件或 database query executor 都不在本批范围内。

## Config Gate

| gate | 当前状态 | 结论 |
| --- | --- | --- |
| `runtime_repository_store_still_disabled` | `satisfied` | `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 仍返回 `repository_store_disabled` |
| `runner_config_separated_from_runtime_store_mode` | `satisfied` | future runner 配置不得复用 runtime store mode 作为启用条件 |
| `service_startup_auto_migration_forbidden` | `satisfied` | platform service startup 不自动运行 migration |
| `manual_runner_entry_defined` | `not_satisfied` | 本批不创建命令入口 |
| `database_connection_provider_defined` | `not_satisfied` | 本批不定义真实数据库连接 |

## Schema Preflight

adapter 当前只能消费静态 manifest 和 `SavedWorkflowDraftRepositorySchemaPreflight`。这足以把 schema migration 未应用、store schema version mismatch 和 migration 状态不可用映射为失败，但不足以证明真实数据库已经完成 migration。

repository mode 未来进入成功路径前，schema preflight 必须能读取实际 applied marker，并在任何不确定状态下阻止 query execution。缺少 applied marker、schema version table、runner 状态或数据库连接时，必须 fail closed。

## Failure Mapping

| 条件 | failure code |
| --- | --- |
| migration 未应用 | `draft_schema_migration_not_applied` |
| store schema version mismatch | `draft_store_schema_version_mismatch` |
| migration 状态不可用 | `draft_store_migration_unavailable` |
| runner 未配置或不可用 | `draft_store_migration_unavailable` |
| repository query executor 不可用 | `draft_store_unavailable` |
| stored record contract mismatch | `draft_store_contract_mismatch` |
| reserved repository mode | `repository_store_disabled` |
| unknown store mode | `invalid_draft_store_mode` |
| auth context contract mismatch | `draft_auth_context_contract_mismatch` |

失败时不得返回草案主体，不得回退 memory dev、sample、fixture、dev route 或 test auth，不得把 fake query executor smoke 提升为真实 DB smoke。

## Rollback / No Fallback / No Side Effects

- rollback 当前只存在静态 `rollback-evidence.json`；没有 rollback runner、rollback command 或 runtime rollback smoke。
- future runner 若采用 forward-only migration，必须在独立实现批次里提供可审查 exception、备份要求和恢复策略。
- 本批不创建数据库连接、不运行 SQL、不写 schema version marker、不调用 OIDC、不写 repository、不启动服务、不创建 production API。
- side effect counters 必须保持：`database_connection_count=0`、`sql_execution_count=0`、`schema_marker_write_count=0`、`repository_write_count=0`、`executor_call_count=0`、`confirmation_call_count=0`、`business_writeback_count=0`、`replay_call_count=0`。

## 验收方式

- `workflow-saved-draft-schema-migration-runner-readiness-v1` fixture / checker 固定 runner boundary、config gate、schema preflight、dependency、failure mapping、rollback、no fallback、no side effects、artifact guard 和 check-repo 顺序。
- Go tests 继续覆盖 repository adapter、adapter smoke、production auth runtime bridge 和 selector fail-closed 行为。
- 验证链路至少包含 `go test ./internal/httpapi`、schema migration runner readiness checker、repository mode enablement checker、schema artifact materialization checker、`./scripts/check-repo.sh --fast`；由于本批回写真相源，补跑全量 `./scripts/check-repo.sh`。

## 停止线

- 不创建 SQL migration、schema version table、migration runner、database query executor、数据库连接或 runner command。
- 不启用 repository mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
- 不把 `draft_schema_migration_runner_readiness_defined` 解释为 migration ready、database ready、repository mode ready、durable persistence ready、production auth ready、production API ready 或 production ready。

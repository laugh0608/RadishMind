# Saved Workflow Draft Database Connection / Schema Marker Preconditions v1

更新时间：2026-06-19

## 专题定位

`Saved Workflow Draft Database Connection / Schema Marker Preconditions v1` 承接 [Saved Workflow Draft Schema Migration Runner Implementation Entry Review v1](saved-workflow-draft-schema-migration-runner-implementation-entry-review-v1.md)，用于在后续打开 migration runner、真实 query executor 或 repository mode runtime 前，先固定数据库连接来源、secret ref、query role、schema applied marker 和环境隔离契约。

结论：状态为 `draft_database_connection_schema_marker_preconditions_defined`。当前只定义 future database connection provider / schema marker contract 的前置条件、failure mapping、no fallback、no side effects 和 artifact guard；不创建数据库连接、secret resolver、SQL migration、schema version table、schema marker 读写、database query executor、migration runner、runner command，也不启用 repository mode。

## 输入证据

- `workflow-saved-draft-schema-migration-runner-implementation-entry-review-v1` 已确认 runner implementation 当前不打开，SQL artifact、schema marker、manual runner、dry-run、apply smoke、rollback observability 和 repository mode runtime 均保持 blocked。
- `workflow-saved-draft-schema-migration-runner-readiness-v1` 已定义 manual runner boundary、schema preflight、failure mapping 和 rollback 停止线。
- `workflow-saved-draft-repository-mode-enablement-v1` 已确认 `repository` store mode 仍不启用。
- `workflow-saved-draft-schema-artifact-materialization-v1` 只物化静态 manifest、DDL review、rollback evidence 和 migration smoke，没有可执行 SQL。
- `workflow-saved-draft-repository-adapter-implementation-v1` 已实现 repository adapter 与 schema preflight，但 query executor 仍为注入式边界。
- `workflow-saved-draft-production-auth-runtime-v1` 已实现 auth context 投影，但没有 OIDC middleware、token validation 或 membership adapter。

## Database Connection Provider Contract

future database connection provider 必须是独立契约，不得由 service startup、HTTP route、store selector 或 saved draft request path 隐式创建连接。后续进入实现前必须明确：

| contract | 要求 |
| --- | --- |
| secret ref | 只记录 `RADISHMIND_WORKFLOW_SAVED_DRAFT_DATABASE_SECRET_REF` 一类引用，不在文档、fixture、日志或源码中保存 secret value |
| environment isolation | dev / test / prod secret ref、database name、role 和 migration target 必须分离，禁止跨环境 fallback |
| query role policy | runtime query role 只允许 saved draft read / write 所需 DML，migration role 独立承担 DDL / marker write |
| connection lifecycle | 明确 timeout、pool policy、health check、request id / audit ref 透传和关闭责任 |
| failure policy | secret 缺失、环境不匹配、连接不可用、role 不足或 driver 未配置时全部 fail closed |

本批不创建 `workflow_saved_draft_database_connection.go`、不导入 `database/sql` 或 `pgxpool`，不读取 secret，不连接数据库。

## Schema Marker Contract

future schema marker 必须能回答已应用、未应用、版本不匹配和 marker 不可读四类状态，并且先于 query execution 执行。进入实现前必须明确：

| marker contract | 要求 |
| --- | --- |
| marker location | future table 暂定为 `workflow_saved_draft_schema_versions`，但本批不创建表 |
| read semantics | runtime 只能读取 marker；缺失、不可读或版本不匹配时阻止 repository query |
| write semantics | 只有 manual migration runner 可写 marker，service startup、HTTP route 和 store selector 不可写 |
| idempotency / locking | manual runner 必须定义已应用、重复 apply、并发 apply 和 lock failure 的行为 |
| failure mapping | marker 缺失映射为 `draft_schema_migration_not_applied`，marker 不可读映射为 `draft_store_migration_unavailable`，版本不匹配映射为 `draft_store_schema_version_mismatch` |

本批不创建 schema version table、schema marker reader / writer、runner command 或 dry-run output。

## Failure Mapping

database connection / schema marker 前置契约继续使用 fail-closed 语义：

| 条件 | failure code |
| --- | --- |
| database secret ref 缺失、连接不可用或 role 不足 | `draft_store_unavailable` |
| migration marker 缺失 | `draft_schema_migration_not_applied` |
| marker 不可读或 migration 状态不可用 | `draft_store_migration_unavailable` |
| store schema version mismatch | `draft_store_schema_version_mismatch` |
| stored record contract mismatch | `draft_store_contract_mismatch` |
| reserved repository mode | `repository_store_disabled` |
| unknown store mode | `invalid_draft_store_mode` |
| auth / tenant / workspace / scope contract mismatch | `draft_auth_context_contract_mismatch` / `draft_tenant_binding_missing` / `draft_workspace_membership_denied` / `draft_scope_grant_missing` |

失败时不得返回草案主体，不得回退到 `memory_dev`、sample、fixture、dev route、test auth 或 fake query executor。

## No Fallback / No Side Effects

- 不允许 repository mode、runner、store selector 或 saved draft request path 因数据库连接缺失而回退到 `memory_dev`。
- 不允许文档或 fixture 承载 secret value，只能承载 secret ref 形状。
- 不允许 fake query executor smoke 被提升为真实 database smoke。
- 不连接数据库，不读取 secret，不运行 SQL，不读写 schema marker，不写 repository，不调用 OIDC，不校验 token，不创建 production API。
- side effect counters 必须保持：`database_connection_count=0`、`secret_resolver_call_count=0`、`sql_execution_count=0`、`schema_marker_read_count=0`、`schema_marker_write_count=0`、`repository_read_count=0`、`repository_write_count=0`、`executor_call_count=0`、`confirmation_call_count=0`、`business_writeback_count=0`、`replay_call_count=0`。

## 后续准入

后续若继续 durable store，应在独立任务卡中选择一个实现方向：

1. database connection provider implementation：只实现 secret ref resolver、connection factory、role policy 和 connection smoke，不创建 runner 或 repository mode 成功路径。
2. schema marker contract implementation：只实现 marker read / write contract、manual runner 依赖和 failure mapping，不创建完整 SQL migration。
3. migration runner implementation：必须同时满足 executable artifact、dry-run、apply smoke、marker contract、rollback observability 和 database connection provider。
4. repository mode runtime enablement：必须先满足真实 database smoke、OIDC middleware、token validation、membership adapter 和 repository mode rollback 停止线。

## 验收方式

- 新增 `workflow-saved-draft-database-connection-schema-marker-preconditions-v1` fixture / checker，固定 connection contract、schema marker contract、blocked implementation matrix、failure mapping、no fallback、no side effects、artifact guard 和 check-repo 顺序。
- checker 接入 `./scripts/check-repo.sh --fast`，顺序位于 `workflow-saved-draft-schema-migration-runner-implementation-entry-review-v1` 之后。
- 本批至少运行专项 checker、schema migration runner implementation entry review checker、schema migration runner readiness checker、repository mode enablement checker、schema artifact materialization checker 和 `./scripts/check-repo.sh --fast`。
- 因同步 current focus、产品范围和能力矩阵，补跑全量 `./scripts/check-repo.sh`。

## 停止线

- 不创建数据库连接、secret resolver、database query executor、SQL migration、schema version table、schema marker reader / writer、migration runner、runner command 或 dry-run output。
- 不启用 repository mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
- 不把 `draft_database_connection_schema_marker_preconditions_defined` 解释为 database ready、schema marker ready、migration ready、repository mode ready、durable persistence ready、production API ready 或 production ready。

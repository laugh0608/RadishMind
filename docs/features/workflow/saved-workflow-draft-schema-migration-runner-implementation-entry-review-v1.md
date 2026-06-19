# Saved Workflow Draft Schema Migration Runner Implementation Entry Review v1

更新时间：2026-06-19

## 专题定位

`Saved Workflow Draft Schema Migration Runner Implementation Entry Review v1` 承接 [Saved Workflow Draft Schema Migration Runner Readiness v1](saved-workflow-draft-schema-migration-runner-readiness-v1.md)，用于评审 future schema migration runner 是否可以进入真实实现批次。

结论：状态为 `draft_schema_migration_runner_implementation_entry_review_defined`。当前只固定 implementation entry review、候选阻塞项、failure mapping、no fallback 和 no side effects；不创建 SQL migration、schema version table、migration runner、runner command、database query executor 或数据库连接，也不启用 repository mode。

## 输入证据

- `workflow-saved-draft-schema-migration-runner-readiness-v1` 已定义 manual runner boundary、config gate、schema preflight、applied marker 缺口、failure mapping 和 rollback 停止线。
- `workflow-saved-draft-repository-mode-enablement-v1` 已确认 `repository` store mode 仍不启用，reserved mode 必须返回 `repository_store_disabled`。
- `workflow-saved-draft-schema-artifact-materialization-v1` 已物化 manifest、DDL review、rollback evidence 和 migration smoke 静态证据，但没有可执行 SQL。
- `workflow-saved-draft-repository-adapter-implementation-v1` 已实现 schema preflight 和 injected query executor adapter，但没有真实 database query executor。
- `workflow-saved-draft-adapter-smoke-v1` 只使用 injected fake query executor，不代表真实 DB smoke。
- `workflow-saved-draft-production-auth-runtime-v1` 已实现 verified auth context 到 repository actor context 的投影，但没有 OIDC middleware、token validation 或 membership adapter。

## Entry Review Decision

| candidate | 本次结论 | 阻塞原因 |
| --- | --- | --- |
| executable migration artifact | `blocked` | 只有静态 manifest / DDL review / rollback evidence / migration smoke，没有 SQL 或等价 apply plan |
| schema version marker contract | `blocked` | 没有 schema version table、applied marker 读写契约和真实 store schema version reader |
| manual runner command | `blocked` | 没有独立 runner config、database connection provider、dry-run plan 输出或 apply command boundary |
| migration apply smoke | `blocked` | 没有真实 query executor、database role、idempotency smoke 或 applied-marker smoke |
| rollback observability smoke | `blocked` | rollback 仍是静态 evidence，没有 rollback command、backup policy 或恢复验证 |
| repository mode runtime enablement | `blocked` | runner、DB connection、applied marker、OIDC middleware 和 membership adapter 均未满足 |

本批不创建 `workflow-saved-draft-schema-migration-runner-implementation-v1` 任务卡。后续若要打开 runner implementation，必须先独立定义可执行 migration artifact、schema marker contract、manual runner command、dry-run output、database connection provider、rollback policy 和真实 smoke。

## Failure Mapping

implementation entry review 继续固定 fail-closed 语义：

| 条件 | failure code |
| --- | --- |
| migration 未应用 | `draft_schema_migration_not_applied` |
| store schema version mismatch | `draft_store_schema_version_mismatch` |
| migration 状态不可用或 runner 不可用 | `draft_store_migration_unavailable` |
| repository query executor 不可用 | `draft_store_unavailable` |
| stored record contract mismatch | `draft_store_contract_mismatch` |
| reserved repository mode | `repository_store_disabled` |
| unknown store mode | `invalid_draft_store_mode` |
| auth context contract mismatch | `draft_auth_context_contract_mismatch` |

失败时不得返回草案主体，不得回退 `memory_dev`、sample、fixture、dev route 或 test auth，不得把 fake query executor smoke 提升为真实 DB smoke。

## No Fallback / No Side Effects

- 不允许 service startup、HTTP route、store selector 或 saved draft request path 自动运行 migration。
- 不允许 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 触发 migration 或进入成功路径。
- 不允许失败后回退到 `memory_dev`、sample、fixture、dev HTTP route 或 test auth。
- 不连接数据库，不运行 SQL，不写 schema marker，不写 repository，不调用 OIDC，不校验 token，不创建 production API。
- side effect counters 必须保持：`database_connection_count=0`、`sql_execution_count=0`、`schema_marker_write_count=0`、`repository_write_count=0`、`executor_call_count=0`、`confirmation_call_count=0`、`business_writeback_count=0`、`replay_call_count=0`。

## 后续准入

后续若继续 durable store，应先创建独立 `Saved Workflow Draft Schema Migration Runner Implementation v1` 任务卡，并在该任务卡中明确：

- migration artifact 或等价 apply plan 的 committed 形态。
- schema version / applied marker 读写契约。
- manual runner command、dry-run 输出、lock policy 和 apply / already-applied / mismatch / unavailable failure mapping。
- database connection provider、query role、secret reference 和 environment isolation。
- rollback / forward-only exception policy。
- migration smoke：schema version、tenant / workspace / application predicate、owner list projection、version conflict predicate 和 no side effects。

即使未来打开 runner implementation，repository mode runtime、Radish OIDC token validation、membership adapter、production API、publish、run、executor、confirmation、writeback 和 replay 仍必须作为后续独立专题推进。

## 验收方式

- 新增 `workflow-saved-draft-schema-migration-runner-implementation-entry-review-v1` fixture / checker，固定 entry decision、candidate matrix、failure mapping、no fallback、no side effects、artifact guard 和 check-repo 顺序。
- checker 接入 `./scripts/check-repo.sh --fast`，顺序位于 `workflow-saved-draft-schema-migration-runner-readiness-v1` 之后。
- 本批至少运行专项 checker、schema migration runner readiness checker、repository mode enablement checker、schema artifact materialization checker 和 `./scripts/check-repo.sh --fast`。
- 因同步 current focus、产品范围和能力矩阵，若时间允许补跑全量 `./scripts/check-repo.sh`。

## 停止线

- 不创建 SQL migration、schema version table、migration runner、runner command、database query executor、数据库连接或 runner task card。
- 不启用 repository mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
- 不把 `draft_schema_migration_runner_implementation_entry_review_defined` 解释为 runner ready、migration ready、database ready、repository mode ready、durable persistence ready、production API ready 或 production ready。

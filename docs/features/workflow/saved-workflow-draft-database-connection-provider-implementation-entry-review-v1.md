# Saved Workflow Draft Database Connection Provider Implementation Entry Review v1

更新时间：2026-06-19

## 专题定位

`Saved Workflow Draft Database Connection Provider Implementation Entry Review v1` 承接 [Saved Workflow Draft Database Connection / Schema Marker Preconditions v1](saved-workflow-draft-database-connection-schema-marker-preconditions-v1.md)，用于评审 future database connection provider 是否可以进入实现批次。

结论：状态为 `draft_database_connection_provider_implementation_entry_review_defined`。当前只固定 connection provider implementation entry review、候选阻塞项、failure mapping、no fallback、no side effects 和 artifact guard；不创建 database connection provider、secret resolver、DB driver、connection factory、role policy、connection smoke、SQL migration、schema marker、database query executor、migration runner，也不启用 repository mode。

## 输入证据

- `workflow-saved-draft-database-connection-schema-marker-preconditions-v1` 已定义 future database connection provider、secret ref、query role、environment isolation 和 schema marker contract 的前置条件。
- `workflow-saved-draft-schema-migration-runner-implementation-entry-review-v1` 已确认 runner implementation 当前不打开。
- `workflow-saved-draft-repository-mode-enablement-v1` 已确认 `repository` store mode 仍不启用。
- `workflow-saved-draft-repository-adapter-implementation-v1` 已实现 repository adapter，但只消费 injected query executor。
- `production-secret-backend-implementation-readiness` 已固定 production secret backend 仍未实现，resolver status 为 `not_started`。
- `production-secret-reference-basic` 只证明 secret reference manifest，不启用 resolver，不保存 secret value，不允许云调用。

## Entry Review Decision

| candidate | 本次结论 | 阻塞原因 |
| --- | --- | --- |
| secret ref resolver boundary | `blocked` | production secret backend resolver 未实现，当前只有 reference-only manifest |
| database driver policy | `blocked` | 未选择 driver / DSN contract / TLS policy，也没有测试环境 smoke |
| connection factory | `blocked` | secret resolver、driver policy、timeout / pool policy 和 sanitized diagnostics 未满足 |
| runtime query role policy | `blocked` | runtime DML role、migration role 分离和 least privilege review 还没有实现证据 |
| connection smoke | `blocked` | 没有真实测试数据库、secret resolver、role policy 或 no secret leakage smoke |
| repository query executor binding | `blocked` | connection provider、schema marker reader、OIDC middleware 和 membership adapter 均未满足 |

本批不创建 `workflow-saved-draft-database-connection-provider-implementation-v1` 任务卡。后续若要打开 connection provider implementation，必须先独立定义 secret resolver interface、driver / DSN policy、connection lifecycle、runtime / migration role policy、sanitized diagnostics 和 offline-safe connection smoke。

## Failure Mapping

entry review 继续固定 fail-closed 语义：

| 条件 | failure code |
| --- | --- |
| secret ref 缺失、resolver 不可用或 resolver disabled | `draft_store_unavailable` |
| DB driver / DSN / TLS policy 未配置 | `draft_store_unavailable` |
| database connection unavailable | `draft_store_unavailable` |
| role 不足或环境不匹配 | `draft_store_unavailable` |
| migration marker 缺失 | `draft_schema_migration_not_applied` |
| marker 不可读或 migration 状态不可用 | `draft_store_migration_unavailable` |
| store schema version mismatch | `draft_store_schema_version_mismatch` |
| reserved repository mode | `repository_store_disabled` |
| unknown store mode | `invalid_draft_store_mode` |

失败时不得返回草案主体，不得回退 `memory_dev`、sample、fixture、dev route 或 test auth，不得把 fake query executor smoke 提升为真实 DB smoke。

## No Fallback / No Side Effects

- 不允许 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 触发 connection provider 创建或进入成功路径。
- 不允许 service startup、HTTP route、store selector 或 saved draft request path 自动解析 secret 或连接数据库。
- 不允许从 committed 文档、fixture、env example 或日志读取 secret value。
- 不连接数据库，不解析 secret，不导入 DB driver，不运行 SQL，不读写 schema marker，不写 repository，不调用 OIDC，不校验 token，不创建 production API。
- side effect counters 必须保持：`database_connection_count=0`、`secret_resolver_call_count=0`、`driver_open_count=0`、`connection_smoke_count=0`、`sql_execution_count=0`、`schema_marker_read_count=0`、`schema_marker_write_count=0`、`repository_read_count=0`、`repository_write_count=0`、`executor_call_count=0`、`confirmation_call_count=0`、`business_writeback_count=0`、`replay_call_count=0`。

## 后续准入

后续若继续 connection provider implementation，应先创建独立任务卡，并在任务卡中明确：

- secret resolver interface 只消费 secret ref，不暴露 secret value 到日志、fixture、response 或 audit。
- driver / DSN / TLS policy、environment isolation 和 runtime / migration role 分离。
- connection lifecycle：timeout、pool、health check、close responsibility、request id / audit ref propagation。
- sanitized diagnostics：只输出 configured / missing / unavailable / denied 等状态，不输出 raw DSN、token、password、provider URL 或 credential。
- connection smoke：使用显式测试环境和 fake resolver 或人工运行记录，不在 fast baseline 中联网或连接真实 production DB。

即使未来打开 connection provider implementation，schema marker reader / writer、SQL migration、migration runner、repository mode runtime、Radish OIDC token validation、membership adapter、production API、publish、run、executor、confirmation、writeback 和 replay 仍必须作为后续独立专题推进。

## 验收方式

- 新增 `workflow-saved-draft-database-connection-provider-implementation-entry-review-v1` fixture / checker，固定 entry decision、candidate matrix、secret backend dependency、failure mapping、no fallback、no side effects、artifact guard 和 check-repo 顺序。
- checker 接入 `./scripts/check-repo.sh --fast`，顺序位于 `workflow-saved-draft-database-connection-schema-marker-preconditions-v1` 之后。
- 本批至少运行专项 checker、database connection / schema marker preconditions checker、schema migration runner implementation entry review checker、repository mode enablement checker 和 `./scripts/check-repo.sh --fast`。
- 因同步 current focus、产品范围和能力矩阵，补跑全量 `./scripts/check-repo.sh`。

## 停止线

- 不创建 connection provider implementation task card、database connection provider、secret resolver、DB driver、connection factory、role policy、connection smoke、database query executor、SQL migration、schema marker reader / writer、migration runner、runner command 或 dry-run output。
- 不启用 repository mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
- 不把 `draft_database_connection_provider_implementation_entry_review_defined` 解释为 database ready、connection provider ready、secret resolver ready、repository mode ready、durable persistence ready、production API ready 或 production ready。

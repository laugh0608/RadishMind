# Saved Workflow Draft Database Secret Resolver Readiness v1

更新时间：2026-06-19

## 专题定位

`Saved Workflow Draft Database Secret Resolver Readiness v1` 承接 [Saved Workflow Draft Database Connection Provider Implementation Entry Review v1](saved-workflow-draft-database-connection-provider-implementation-entry-review-v1.md)，用于在 future database connection provider implementation 之前，先固定 saved draft database secret ref resolver 的输入、输出、禁用态、failure taxonomy、sanitized diagnostics 和离线验证边界。

结论：状态为 `draft_database_secret_resolver_readiness_defined`。当前只定义 secret resolver readiness；不创建 resolver implementation、fake resolver、DB driver、connection provider、connection factory、connection smoke、database query executor、schema marker、SQL migration、migration runner，也不启用 repository mode。

## 输入证据

- `workflow-saved-draft-database-connection-provider-implementation-entry-review-v1` 已确认 connection provider implementation 当前不打开。
- `workflow-saved-draft-database-connection-schema-marker-preconditions-v1` 已固定 future database connection provider、secret ref、query role、environment isolation 和 schema marker contract。
- `production-secret-backend-implementation-readiness` 已固定 production secret backend resolver 仍为 `not_started`。
- `production-secret-reference-basic` 只证明 reference-only manifest，不保存 secret value，不启用 resolver，不允许云调用。
- `workflow-saved-draft-repository-mode-enablement-v1` 已确认 `repository` store mode 仍不启用。

## Secret Resolver Readiness

| contract | 本批结论 |
| --- | --- |
| secret ref key | future saved draft database secret ref 使用 `RADISHMIND_WORKFLOW_SAVED_DRAFT_DATABASE_SECRET_REF`，但本批不读取 env、不解析 secret |
| input shape | resolver 未来只接收 environment、secret ref、required credential type、request id / audit ref 和 caller purpose |
| result shape | resolver 未来只返回 `resolved_for_connection_factory` 状态和 opaque credential handle，不把 raw secret 暴露给日志、fixture、response 或 audit |
| disabled runtime | 当前 runtime 默认禁用；service startup、HTTP route、store selector 和 saved draft request path 不得触发 secret resolution |
| sanitized diagnostics | 只能输出 configured / missing / disabled / unavailable / denied / environment_mismatch 等状态 |
| fake resolver strategy | fake resolver 只能在后续 implementation task 中显式创建；fast baseline 不联网、不调用云 SDK、不要求真实 credential |
| environment binding | dev / test / prod secret ref、database role 和 database name 必须分离，禁止跨环境 fallback |

## Failure Taxonomy

| 条件 | failure code | 诊断口径 |
| --- | --- | --- |
| secret ref missing | `draft_store_unavailable` | `missing_secret_ref` |
| secret backend disabled | `draft_store_unavailable` | `secret_backend_disabled` |
| resolver unavailable | `draft_store_unavailable` | `secret_resolver_unavailable` |
| resolution denied | `draft_store_unavailable` | `secret_resolution_denied` |
| credential missing | `draft_store_unavailable` | `credential_missing` |
| environment mismatch | `draft_store_unavailable` | `environment_mismatch` |
| repository mode disabled | `repository_store_disabled` | `repository_mode_disabled` |
| unknown store mode | `invalid_draft_store_mode` | `invalid_store_mode` |

失败时不得返回 secret value、DSN、provider URL、raw token、草案主体或数据库错误细节；不得回退 `memory_dev`、sample、fixture、dev route、test auth 或 fake query executor。

## No Fallback / No Side Effects

- 不允许 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 触发 secret resolution 或 connection provider 成功路径。
- 不允许从 committed 文档、fixture、env example、日志或 response 读取 secret value。
- 不允许 committed fixture 包含 raw DSN、password、token、API key、provider URL 或 cloud credential。
- 不连接数据库，不解析 secret，不调用云 secret 服务，不导入 DB driver，不打开 driver，不运行 SQL，不读写 schema marker，不写 repository，不调用 OIDC，不校验 token，不创建 production API。
- side effect counters 必须保持：`secret_resolver_call_count=0`、`secret_value_read_count=0`、`cloud_secret_call_count=0`、`database_connection_count=0`、`driver_open_count=0`、`connection_smoke_count=0`、`sql_execution_count=0`、`schema_marker_read_count=0`、`schema_marker_write_count=0`、`repository_read_count=0`、`repository_write_count=0`、`executor_call_count=0`、`confirmation_call_count=0`、`business_writeback_count=0`、`replay_call_count=0`。

## 后续准入

后续若打开 secret resolver implementation，必须先创建独立任务卡，并满足：

- 只消费 secret ref，不保存、不打印、不返回 secret value。
- 明确 resolver interface、opaque credential handle、sanitized diagnostics、environment binding 和 failure taxonomy。
- 使用 fake resolver 或人工运行记录验证离线 smoke；fast baseline 不联网、不调用真实 cloud secret backend。
- 明确 resolver disabled 默认态和 operator enablement gate。

即使未来 secret resolver implementation 打开，database connection provider、DB driver / DSN policy、connection lifecycle、role policy、connection smoke、schema marker reader / writer、SQL migration、migration runner、repository mode runtime、OIDC middleware、token validation、membership adapter、production API、publish、run、executor、confirmation、writeback 和 replay 仍必须作为后续独立专题推进。

## 验收方式

- 新增 `workflow-saved-draft-database-secret-resolver-readiness-v1` fixture / checker，固定 resolver readiness、dependency alignment、failure taxonomy、sanitized diagnostics、no fallback、no side effects、artifact guard 和 check-repo 顺序。
- checker 接入 `./scripts/check-repo.sh --fast`，顺序位于 `workflow-saved-draft-database-connection-provider-implementation-entry-review-v1` 之后。
- 本批至少运行专项 checker、database connection provider implementation entry review checker、production secret backend implementation readiness checker、production secret reference contract checker 和 `./scripts/check-repo.sh --fast`。
- 因同步 current focus、产品范围和能力矩阵，补跑全量 `./scripts/check-repo.sh`。

## 停止线

- 不创建 secret resolver implementation task card、resolver implementation、fake resolver、database connection provider、DB driver、connection factory、role policy、connection smoke、database query executor、schema marker reader / writer、SQL migration、migration runner、runner command 或 dry-run output。
- 不启用 repository mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
- 不把 `draft_database_secret_resolver_readiness_defined` 解释为 secret resolver ready、database connection provider ready、database ready、repository mode ready、durable persistence ready、production API ready 或 production ready。

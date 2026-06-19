# Saved Workflow Draft Database Secret Resolver Implementation Entry Review v1

更新时间：2026-06-19

## 专题定位

`Saved Workflow Draft Database Secret Resolver Implementation Entry Review v1` 承接 [Saved Workflow Draft Database Secret Resolver Readiness v1](saved-workflow-draft-database-secret-resolver-readiness-v1.md)，用于评审 saved draft database secret resolver implementation 是否可以进入实现批次。

结论：状态为 `draft_database_secret_resolver_implementation_entry_review_defined`，entry decision 为 `secret_resolver_implementation_entry_not_opened`。当前依赖仍不满足，secret resolver implementation entry 不打开；不创建 resolver implementation、fake resolver、DB provider、DB driver、connection factory、SQL、schema marker、migration runner、repository mode 或 production API。

## 输入证据

- `workflow-saved-draft-database-secret-resolver-readiness-v1` 已定义 secret ref key、resolver input / result shape、disabled runtime、sanitized diagnostics、failure taxonomy、environment binding 和 offline fake resolver strategy。
- `production-secret-backend-implementation-readiness` 仍显示 production secret backend resolver implementation 为 `not_started`，默认 runtime state 为 `disabled_until_explicit_secret_backend_task`。
- `production-secret-backend-config-secret-ref-readiness-v1` 已固定 `config_secret_ref_readiness_defined`，只满足 config 注入点和 reference-only manifest 消费边界，不满足 provider profile binding、resolver interface、operator runbook 或 production backend ready。
- `production-secret-reference-basic` 仍是 reference-only manifest：不保存 secret value，不启用 resolver，不允许 cloud calls，不声明 production secret backend ready。
- `workflow-saved-draft-database-connection-provider-implementation-entry-review-v1` 已确认 connection provider implementation 当前不打开。
- `workflow-saved-draft-repository-mode-enablement-v1` 已确认 `repository` store mode 仍不启用。

## Entry Review Decision

| candidate | 本次结论 | 阻塞原因 |
| --- | --- | --- |
| resolver interface implementation | `blocked` | production secret backend resolver 未实现，当前只有 reference-only manifest、config secret ref readiness 和 readiness contract |
| offline fake resolver | `blocked` | fake resolver 只能在独立 implementation task 中显式创建，当前没有 no secret leakage smoke |
| sanitized diagnostics runtime | `blocked` | diagnostics 口径已定义，但没有 resolver interface、fake resolver 或 runtime emission gate |
| environment binding enforcement | `blocked` | dev / test / prod secret ref 分离、operator enablement gate 和 backend binding 尚无实现证据 |
| connection factory handoff | `blocked` | connection provider implementation entry review 仍为 blocked，不能把 opaque handle 交给不存在的 factory |
| repository mode integration | `blocked` | repository mode 仍 fail closed，不能触发 secret resolution 或 connection provider 成功路径 |

本批不创建 `workflow-saved-draft-database-secret-resolver-implementation-v1` 任务卡。后续若要打开 implementation，必须先独立满足 provider profile secret binding、disabled resolver interface、offline fake resolver、sanitized diagnostics runtime、environment binding、no secret leakage smoke 和 operator enablement gate。

## Failure Mapping

| 条件 | failure code | sanitized diagnostic |
| --- | --- | --- |
| secret ref 缺失 | `draft_store_unavailable` | `missing_secret_ref` |
| secret backend disabled | `draft_store_unavailable` | `secret_backend_disabled` |
| resolver unavailable | `draft_store_unavailable` | `secret_resolver_unavailable` |
| resolution denied | `draft_store_unavailable` | `secret_resolution_denied` |
| credential missing | `draft_store_unavailable` | `credential_missing` |
| environment mismatch | `draft_store_unavailable` | `environment_mismatch` |
| repository mode disabled | `repository_store_disabled` | `repository_mode_disabled` |
| unknown store mode | `invalid_draft_store_mode` | `invalid_store_mode` |

失败时不得返回 secret value、DSN、provider URL、raw token、数据库错误细节或草案主体；不得回退 `memory_dev`、sample、fixture、dev route、test auth、fake resolver 或 fake query executor。

## Sanitized Diagnostics

后续 implementation 只能输出以下脱敏状态：`configured`、`missing_secret_ref`、`secret_backend_disabled`、`secret_resolver_unavailable`、`secret_resolution_denied`、`credential_missing`、`environment_mismatch`、`repository_mode_disabled` 和 `invalid_store_mode`。

diagnostics 不得包含 raw secret、password、token、API key、provider URL、DSN、cloud credential、database hostname、database error detail 或完整用户 claim。任何需要 operator 排查的细节只能通过 request id / audit ref 与后续人工运行记录关联。

## No Fallback / No Side Effects

- `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 不得触发 secret resolution，也不得把 resolver failure 回退到 `memory_dev`。
- resolver disabled、secret backend unavailable、missing secret ref 或 resolution denied 均不得回退到 fixture、sample、test auth、env plaintext 或 fake query executor。
- 不解析 secret，不读取 secret value，不调用云 secret 服务，不连接数据库，不导入或打开 DB driver，不运行 SQL，不读写 schema marker，不写 repository，不调用 OIDC，不校验 token，不创建 production API。
- side effect counters 必须保持：`secret_resolver_call_count=0`、`fake_resolver_call_count=0`、`secret_value_read_count=0`、`cloud_secret_call_count=0`、`database_connection_count=0`、`driver_open_count=0`、`connection_smoke_count=0`、`sql_execution_count=0`、`schema_marker_read_count=0`、`schema_marker_write_count=0`、`repository_read_count=0`、`repository_write_count=0`、`executor_call_count=0`、`confirmation_call_count=0`、`business_writeback_count=0`、`replay_call_count=0`。

## 后续准入

后续若打开 secret resolver implementation，必须先创建独立 implementation task card，并满足：

- production secret backend implementation 不再停留在 config / reference readiness，provider profile binding、resolver interface、默认禁用态、operator enablement gate 和 offline fast baseline 均明确。
- resolver interface 只消费 secret ref、environment、required credential type、request id / audit ref 和 caller purpose。
- resolver result 只返回 `resolved_for_connection_factory` 状态和 opaque credential handle，不返回 raw secret。
- offline fake resolver 只能用于 implementation task 的离线 smoke，不得被提升为 production resolver。
- sanitized diagnostics、environment binding、failure taxonomy、no fallback 和 no side effects 均有 checker 守护。

即使未来 secret resolver implementation 打开，database connection provider、DB driver / DSN policy、connection lifecycle、role policy、connection smoke、schema marker reader / writer、SQL migration、migration runner、repository mode runtime、OIDC middleware、token validation、membership adapter、production API、publish、run、executor、confirmation、writeback 和 replay 仍必须作为后续独立专题推进。

## 验收方式

- 新增 `workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1` fixture / checker，固定 entry decision、blocked reasons、failure mapping、sanitized diagnostics、no fallback、no side effects、artifact guard 和 check-repo 顺序。
- checker 接入 `./scripts/check-repo.sh --fast`，顺序位于 `workflow-saved-draft-database-secret-resolver-readiness-v1` 之后。
- 本批至少运行专项 checker、database secret resolver readiness checker、production secret backend implementation readiness checker、production secret reference contract checker 和 `./scripts/check-repo.sh --fast`。
- 因同步 current focus、产品范围和能力矩阵，补跑全量 `./scripts/check-repo.sh`。

## 停止线

- 不创建 secret resolver implementation task card、resolver implementation、fake resolver、database connection provider、DB driver、connection factory、role policy、connection smoke、database query executor、schema marker reader / writer、SQL migration、migration runner、runner command 或 dry-run output。
- 不启用 repository mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
- 不把 `draft_database_secret_resolver_implementation_entry_review_defined` 解释为 secret resolver ready、database connection provider ready、database ready、repository mode ready、durable persistence ready、production API ready 或 production ready。

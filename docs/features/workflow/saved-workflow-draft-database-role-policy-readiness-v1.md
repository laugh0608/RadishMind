# Saved Workflow Draft Database Role Policy Readiness v1

更新时间：2026-06-22

## 专题定位

`Saved Workflow Draft Database Role Policy Readiness v1` 承接 [Saved Workflow Draft Database Driver / DSN / TLS Policy Readiness v1](saved-workflow-draft-database-driver-dsn-tls-policy-readiness-v1.md)，只定义 future saved draft database connection provider 之前的 runtime / migration role policy 边界。

结论：状态为 `draft_database_role_policy_readiness_defined`。本批只固定 future runtime DML role、migration DDL / schema marker role、least privilege review、cross-environment denial smoke 前置、environment binding、role claim / role id metadata-only shape、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard；不创建 role policy runtime、DB provider、DB driver import、DSN parser runtime、connection factory、connection smoke、SQL、schema marker、migration runner、repository mode runtime、OIDC middleware、token validator、membership adapter、production API、executor、confirmation、writeback 或 replay。

## 输入证据

- `workflow-saved-draft-database-driver-dsn-tls-policy-readiness-v1` 已确认 driver / DSN / TLS policy 只定义 future 策略，role policy 仍需独立 readiness。
- `workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v1` 已确认 connection provider implementation task card 当前仍为 `blocked_before_implementation_task_card`。
- `workflow-saved-draft-database-connection-schema-marker-preconditions-v1` 已定义 future database connection provider、query role、environment isolation 和 schema marker read / write contract。
- `workflow-saved-draft-manual-migration-runner-implementation-entry-refresh-v1` 与 `workflow-saved-draft-schema-marker-contract-implementation-entry-review-v1` 已确认 manual runner、marker reader / writer 和 schema version table 仍 blocked。
- `workflow-saved-draft-repository-mode-runtime-boundary-review-v1` 已确认 repository mode runtime task card 仍 blocked，不能通过本批绕过 database、auth、marker、audit 或 production API 前置。
- `radish-oidc-token-membership-implementation-entry-review-v1` 已确认 OIDC middleware、token validator 和 membership adapter 仍 blocked；本批只能定义 role claim metadata shape，不创建 auth runtime。
- `production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1` 已确认 no leakage smoke runtime 仍 blocked；本批不能执行 connection smoke 或 diagnostics runtime scanner。

## Readiness Decision

| policy area | 本次结论 | 阻塞或停止线 |
| --- | --- | --- |
| runtime DML role | `defined_no_runtime_role` | 只定义 future repository adapter 可消费的 DML role；本批不创建 role policy runtime、不创建数据库 role、不授予权限 |
| migration DDL / marker role | `defined_no_runtime_role` | 只定义 future manual runner / marker writer role；本批不创建 DDL、marker table、runner 或 role binding |
| least privilege review | `defined_static_review_required` | 后续实现任务卡必须证明 runtime role 不能执行 DDL / marker write，migration role 不能被 repository adapter 消费 |
| environment binding | `defined_no_runtime_binding` | role policy environment、secret ref environment、credential handle environment、TLS policy environment 和 connection smoke environment 必须一致 |
| role claim / role id metadata shape | `defined_metadata_only_no_auth_runtime` | 只允许 policy id、role class、environment id、role fingerprint 和 sanitized failure code；不输出完整 claim、token、credential 或 raw grants |
| cross-environment denial smoke | `blocked_before_connection_smoke_strategy` | 必须等 connection smoke strategy 定义 explicit test database、placeholder credential handoff 和 no leakage scan |
| connection provider dependency | `blocked_before_implementation_task_card` | provider task card 仍需等待 role policy、connection smoke、secret resolver、schema marker 和 repository mode runtime 依赖可复验 |

## Role Policy Boundary

future connection provider 必须遵守以下 role policy：

- runtime DML role 只能被 repository adapter 的 save / read / list 路径消费，且只能访问 saved draft 正常 DML 范围。
- runtime DML role 不得执行 DDL、schema marker write、migration lock write、schema version update、role grant 或跨环境连接。
- migration DDL / marker role 只能被 future manual migration runner、schema marker reader / writer 或显式 migration apply path 消费。
- migration DDL / marker role 不得被 saved draft repository adapter、HTTP request path、store selector、dev route 或 workflow executor 消费。
- role policy 输入只能包含 `environment`、`required_role`、`policy_version`、`role_policy_ref`、`role_id_ref`、`request_id`、`audit_ref` 和 auth context reference。
- role policy 输出只能包含 metadata-only role readiness / denial result、policy id、role class、environment id、redacted role fingerprint 和 sanitized failure code。
- role policy 不得输出完整 user claim、authorization header、cookie、raw grant、database role password、DSN、secret value、credential payload、database host 或 SQL detail。
- environment mismatch、role escalation attempt、missing role policy、unknown role class 和 cross-environment role binding 必须 fail closed，并返回 `draft_store_unavailable`。

## Failure Mapping

| 条件 | failure code |
| --- | --- |
| repository mode reserved / runtime 未打开 | `repository_store_disabled` |
| unknown store mode | `invalid_draft_store_mode` |
| runtime DML role policy 缺失、角色未绑定或 role class 不可信 | `draft_store_unavailable` |
| migration DDL / marker role policy 缺失、角色混用或 marker write 权限不可证明 | `draft_store_unavailable` |
| least privilege review 缺失或 runtime / migration role 混用 | `draft_store_unavailable` |
| role policy environment 与 secret / credential / TLS / smoke environment 不一致 | `draft_store_unavailable` |
| cross-environment denial smoke 前置缺失或 no leakage smoke runtime 未创建 | `draft_store_unavailable` |
| marker 缺失或 migration 未应用 | `draft_schema_migration_not_applied` |
| marker version mismatch | `draft_store_schema_version_mismatch` |
| marker 不可读、runner 状态不可观测或 lock 状态不可信 | `draft_store_migration_unavailable` |
| stored record contract mismatch | `draft_store_contract_mismatch` |
| repository actor context 不可信 | `draft_auth_context_contract_mismatch` |

失败时不得返回 raw role grant、database role name、database error detail、raw DSN、secret value、credential payload、完整 secret ref、完整 credential handle、完整 user claim、authorization header、cookie、SQL detail 或草案主体。

## No Fallback / No Side Effects

- 不允许 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 因本 readiness 被启用。
- 不允许 service startup、HTTP route、store selector 或 saved draft request path 自动解析 role claim、选择 database role、授予权限、连接数据库、读取 marker 或写 repository。
- 不允许 test-only fake resolver runtime、fake query executor smoke、static schema artifact、rollback evidence、driver / DSN / TLS policy 或 container smoke 替代 role policy readiness 后续实现前置。
- 本批 checker 只读取 committed 文档、fixture、源码和 `check-repo.py` 注册顺序；不启动服务、不读取 env secret、不调用 fake resolver、不调用 production resolver、不调用云 secret 服务、不导入 driver、不创建 role policy runtime、不创建 DSN parser、不连接数据库、不运行 connection smoke、不运行 SQL、不读写 schema marker、不创建 runner、不调用 OIDC、不校验 token、不创建 membership adapter、不创建 production API、不执行 workflow。
- side effect counters 必须保持 `secret_resolver_call_count=0`、`fake_resolver_call_count=0`、`production_resolver_call_count=0`、`secret_value_read_count=0`、`credential_handle_created_count=0`、`cloud_secret_call_count=0`、`network_call_count=0`、`database_connection_count=0`、`driver_import_count=0`、`driver_open_count=0`、`dsn_parse_count=0`、`dsn_build_count=0`、`tls_config_create_count=0`、`role_policy_evaluation_count=0`、`role_claim_parse_count=0`、`role_grant_count=0`、`connection_smoke_count=0`、`sql_execution_count=0`、`schema_marker_read_count=0`、`schema_marker_write_count=0`、`repository_mode_enablement_count=0`、`repository_read_count=0`、`repository_write_count=0`、`oidc_call_count=0`、`token_validation_count=0`、`membership_lookup_count=0`、`production_api_call_count=0`、`executor_call_count=0`、`confirmation_call_count=0`、`business_writeback_count=0` 和 `replay_call_count=0`。

## 后续方向

本批后，connection provider implementation task card 仍不创建。后续若继续 durable store 上游，应从以下方向选择一个独立推进：

1. `connection smoke strategy`：定义 explicit test database、placeholder credential handoff、smoke output shape、role denial cases、no leakage scan 和 zero production side effect 记录。
2. `Radish OIDC upstream evidence refresh`：补 reviewed issuer、JWKS、client registration、auth middleware ownership 和 membership data source ownership。
3. `connection provider implementation entry refresh v2`：只有 driver / DSN / TLS、role policy、connection smoke、secret resolver、schema marker 和 repository mode runtime 依赖均可复验后再复评。

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-role-policy-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-driver-dsn-tls-policy-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-mode-runtime-boundary-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-radish-oidc-token-membership-implementation-entry-review-v1.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 `docs/task-cards/workflow-saved-draft-database-connection-provider-implementation-v1-plan.md`。
- 不创建 role policy runtime、database role、role grant、database connection provider、DB driver import、DSN parser runtime、connection factory、connection smoke、database query executor、schema marker reader / writer、SQL migration、migration runner、runner command 或 dry-run output。
- 不启用 repository mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
- 不把 `draft_database_role_policy_readiness_defined` 解释为 database role ready、role runtime ready、connection provider ready、database ready、secret resolver ready、repository mode ready、durable persistence ready、production API ready 或 production ready。

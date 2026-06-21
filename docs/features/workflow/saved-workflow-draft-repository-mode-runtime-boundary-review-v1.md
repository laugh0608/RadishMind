# Saved Workflow Draft Repository Mode Runtime Boundary Review v1

更新时间：2026-06-21

## 专题定位

`Saved Workflow Draft Repository Mode Runtime Boundary Review v1` 承接 repository adapter implementation、adapter smoke execution、production auth runtime bridge、repository mode enablement、schema migration runner implementation entry review、database connection / schema marker preconditions、connection provider implementation entry review、database secret resolver readiness / implementation entry review，以及 production secret backend audit store runtime entry refresh v3，用于重新评审 `repository` store mode 是否可以进入 runtime enablement 或实现任务卡。

结论：状态为 `draft_repository_mode_runtime_boundary_review_defined`。entry decision 为 `repository_mode_runtime_still_blocked_before_task_card`。当前已有 repository adapter、adapter smoke 和 production auth runtime bridge 的受控运行层证据，但这些证据仍不能替代真实 schema migration runner、database connection provider、database secret resolver、schema marker、真实 query executor、Radish OIDC middleware、token validation、membership adapter 或 production API consumer。因此本批不创建 repository mode runtime implementation task card，也不启用 repository mode。

本批只固定 runtime boundary review 和下一步依赖选择，不创建 repository mode runtime、不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径、不创建数据库连接、不读取 secret、不调用云 secret 服务、不导入 DB driver、不运行 SQL、不读写 schema marker、不创建 migration runner、不接 OIDC middleware、不校验 token、不创建 membership adapter、不创建 production API、不创建 executor / confirmation / writeback / replay。

## 输入证据

- `workflow-saved-draft-repository-adapter-implementation-v1` 已实现 `SavedWorkflowDraftRepository` interface、注入式 query executor adapter、schema preflight、auth context failure mapping 和 adapter unit tests。
- `workflow-saved-draft-adapter-smoke-v1` 已用 injected fake query executor 执行 save / read / list adapter smoke；该 smoke 不代表真实数据库 query executor 或 repository mode runtime。
- `workflow-saved-draft-production-auth-runtime-v1` 已实现 verified auth context + workspace binding 到 repository actor context 的纯函数投影；它不创建 OIDC middleware、token validation 或 membership adapter。
- `workflow-saved-draft-repository-mode-enablement-v1` 已确认 `repository` 与 `repository_disabled` store mode 必须 fail closed 为 `repository_store_disabled`。
- `workflow-saved-draft-schema-migration-runner-implementation-entry-review-v1` 已确认 runner implementation 当前不打开。
- `workflow-saved-draft-database-connection-schema-marker-preconditions-v1` 已固定 future database connection provider 与 schema marker contract，但不创建 runtime。
- `workflow-saved-draft-database-connection-provider-implementation-entry-review-v1` 已确认 connection provider implementation 当前不打开。
- `workflow-saved-draft-database-secret-resolver-readiness-v1` 与 `workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1` 已固定 secret resolver readiness 与 implementation entry blocked 结论。
- `production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3` 已确认 audit store runtime task card 仍 blocked；它不能作为 repository mode runtime 的 audit backend 替代品。
- `production-ops-secret-backend-implementation-readiness` 仍保持 production secret backend 为 `not_satisfied`，production resolver runtime、credential handle runtime、approval runtime、backend health runtime、audit store runtime、no leakage smoke runtime 均未创建。

## Runtime Boundary Decision

| gate | 当前结论 | 说明 |
| --- | --- | --- |
| repository adapter | `satisfied_adapter_boundary` | adapter 已实现，但只消费注入式 query executor |
| adapter smoke | `satisfied_fake_query_executor_smoke` | 已验证 adapter contract，不代表真实 DB smoke |
| production auth runtime bridge | `satisfied_projection_only` | 只接受已验证输入，不创建 OIDC runtime |
| repository mode selector | `satisfied_fail_closed` | `repository` / `repository_disabled` 仍返回 `repository_store_disabled` |
| schema migration runner implementation | `blocked` | 无 SQL / apply plan、schema marker、runner command |
| database connection provider | `blocked` | secret resolver、driver policy、role policy、connection smoke 未满足 |
| database secret resolver implementation | `blocked` | production secret backend resolver 仍未实现 |
| schema marker runtime | `not_created` | 不读写 applied marker |
| real query executor | `not_created` | 不创建 SQL / database query executor |
| OIDC / token / membership runtime | `not_created` | 不接 Radish OIDC、不校验 token、不查 membership |
| production API consumer | `not_created` | 不创建 public production API |
| repository mode runtime task card | `blocked_before_task_card` | 本批不创建实现任务卡 |

## Blocked Conditions

后续如需创建 repository mode runtime implementation task card，至少必须先解决：

- `schema_migration_runner_implementation_not_opened`
- `executable_migration_artifact_missing`
- `schema_marker_contract_not_implemented`
- `database_connection_provider_implementation_not_opened`
- `database_secret_resolver_implementation_not_opened`
- `real_query_executor_not_created`
- `repository_mode_rollback_smoke_missing`
- `oidc_middleware_not_created`
- `token_validation_not_created`
- `workspace_membership_adapter_not_created`
- `production_api_consumer_not_created`
- `production_secret_backend_not_satisfied`
- `audit_store_runtime_not_created`

这些阻塞项不能用 injected fake query executor、memory dev store、test-only fake resolver runtime、static schema artifact、static rollback evidence、operator runbook、audit store handoff、audit event schema、backend health boundary、历史 adapter smoke 或本地 container smoke 替代。

## 下一步依赖选择

本次复评之后，合理的下一批开发不应直接启用 repository mode，而应在以下依赖中选择一个独立推进：

1. `schema marker contract / migration runner implementation readiness`：先定义 applied marker read / write contract、manual runner command、dry-run output、idempotency / lock failure 和 rollback observability。
2. `database connection provider implementation entry refresh`：先复评 secret resolver、driver / DSN / TLS policy、role policy、connection lifecycle、connection smoke 是否足以创建实现任务卡。
3. `database secret resolver implementation entry refresh`：先消费 production secret backend 最新 evidence，确认 saved draft DB secret resolver 是否仍 blocked，且不得读取 secret。
4. `Radish OIDC token validation / membership adapter readiness`：先定义真实 auth middleware、tenant / workspace / application / owner binding、scope projection 和 failure taxonomy。
5. `production API consumer readiness`：只有 repository mode、auth、database 和 smoke 都可复验后才重新评审。

如果以上依赖仍 blocked，repository mode runtime implementation task card 继续保持不创建。

## Failure Mapping

| 条件 | failure code |
| --- | --- |
| repository mode reserved / runtime 未打开 | `repository_store_disabled` |
| unknown store mode | `invalid_draft_store_mode` |
| migration 未应用 | `draft_schema_migration_not_applied` |
| schema version mismatch | `draft_store_schema_version_mismatch` |
| migration 状态不可用 | `draft_store_migration_unavailable` |
| database connection / query executor 不可用 | `draft_store_unavailable` |
| stored record contract mismatch | `draft_store_contract_mismatch` |
| repository actor context 不可信 | `draft_auth_context_contract_mismatch` |
| actor subject 缺失 | `draft_identity_context_missing` |
| tenant / workspace / application / owner binding 失败 | `draft_tenant_binding_missing` / `draft_workspace_membership_denied` / `draft_application_scope_denied` / `draft_owner_scope_denied` |
| scope grant 缺失 | `draft_scope_grant_missing` |
| audit context 缺失 | `draft_audit_context_missing` |

失败时不得返回草案主体、secret value、DSN、provider URL、database hostname、database error detail、完整 claim、raw request payload、raw response payload 或 audit payload；不得回退 `memory_dev`、sample、fixture、dev route、test auth、fake query executor 或 test-only fake resolver runtime。

## No Fallback / No Side Effects

- `repository` store mode 不允许通过 fallback 到 `memory_dev`、sample、fixture、dev HTTP route、test auth、fake query executor、test-only fake resolver runtime 或 static schema artifact 变成成功路径。
- schema preflight failure、auth failure、connection failure、secret resolver disabled 和 audit backend missing 都必须 fail closed。
- 本批 checker 只读取 committed 文档、fixture、源码和 `check-repo.py` 注册顺序；不启动服务、不解析 env secret、不调用云 secret 服务、不连接数据库、不打开 driver、不运行 SQL、不读写 schema marker、不调用 OIDC、不校验 token、不创建 membership adapter、不写 repository、不创建 production API、不执行 workflow。
- side effect counters 必须保持 `database_connection_count=0`、`secret_resolver_call_count=0`、`cloud_secret_call_count=0`、`driver_open_count=0`、`sql_execution_count=0`、`schema_marker_read_count=0`、`schema_marker_write_count=0`、`repository_mode_enablement_count=0`、`repository_write_count=0`、`oidc_call_count=0`、`token_validation_count=0`、`membership_lookup_count=0`、`production_api_call_count=0`、`executor_call_count=0`、`confirmation_call_count=0`、`business_writeback_count=0`、`replay_call_count=0`。

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/features/workflow/saved-workflow-draft-repository-mode-runtime-boundary-review-v1.md`
- `docs/task-cards/workflow-saved-draft-repository-mode-runtime-boundary-review-v1-plan.md`
- `scripts/checks/fixtures/workflow-saved-draft-repository-mode-runtime-boundary-review-v1.json`
- `scripts/checks/control_plane/check-workflow-saved-draft-repository-mode-runtime-boundary-review-v1.py`

不得新增 repository mode runtime implementation task card、repository mode runtime、SQL query executor、database connection provider、database secret resolver、fake database secret resolver、DB driver policy、schema marker reader / writer、migration runner、runner command、SQL migration、OIDC middleware、token validation、membership adapter、production API consumer、executor、confirmation、writeback 或 replay artifact。

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-mode-runtime-boundary-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-provider-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-migration-runner-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```

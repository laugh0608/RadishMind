# Saved Workflow Draft Repository Mode Enablement v1

更新时间：2026-06-18

## 专题定位

`Saved Workflow Draft Repository Mode Enablement v1` 是 repository mode 的准入评审。它承接 store selector、静态 schema artifact、repository adapter、adapter smoke execution 和 production auth runtime bridge，判断 `repository` store mode 是否可以从 reserved fail-closed 进入 runtime enablement。

结论：状态为 `draft_repository_mode_enablement_review_defined`，当前不启用 repository mode。`repository` 和 `repository_disabled` 仍必须返回 `repository_store_disabled`，不得回退 `memory_dev`、sample、fixture、dev route 或 test auth。

## 输入证据

- `workflow-saved-draft-store-selector-smoke-v1`：已实现 formal config 和 `SelectWorkflowSavedDraftStore`，但 reserved repository mode 仍 fail closed。
- `workflow-saved-draft-schema-artifact-materialization-v1`：已物化 manifest、DDL review、rollback evidence 和 migration smoke 静态证据，但没有 SQL migration、schema version table 或 migration runner。
- `workflow-saved-draft-repository-adapter-implementation-v1`：已实现 repository interface、注入式 query executor adapter、schema preflight 和 auth context failure mapping，但没有数据库连接或 runtime selector binding。
- `workflow-saved-draft-adapter-smoke-v1`：已用 injected fake query executor 执行 save / read / list adapter smoke，但没有真实 repository query executor 或数据库 smoke。
- `workflow-saved-draft-production-auth-runtime-v1`：已实现 verified auth context + workspace binding 到 repository actor context 的纯函数投影，但没有 OIDC middleware、token validation 或 membership adapter。

## Runtime Boundary

当前 runtime boundary 只允许继续使用 `memory_dev` dev store。`repository` mode 进入 runtime 前必须同时满足：

1. config gate 明确启用 repository mode，并与 dev-only HTTP route 分离。
2. schema preflight 可验证已应用 migration、store schema version 和 rollback 状态。
3. repository adapter 通过真实 query executor smoke，而不是 injected fake executor。
4. production auth runtime bridge 的输入来自真实 OIDC middleware、token validation 和 membership adapter。
5. failure mapping 能把 schema、auth、scope、store unavailable、contract mismatch 和 rollback failure 全部 fail closed。

以上条件当前均不能被自动推导为已满足；因此本批不把 selector 的 `repository` 分支接到 `NewSavedWorkflowDraftRepositoryAdapter`。

## Config Gate

| gate | 当前状态 | 结论 |
| --- | --- | --- |
| `memory_dev_default` | `satisfied` | 默认仍为 dev store |
| `repository_reserved_fail_closed` | `satisfied` | `repository` / `repository_disabled` 返回 `repository_store_disabled` |
| `unknown_mode_fail_closed` | `satisfied` | unknown mode 返回 `invalid_draft_store_mode` |
| `repository_runtime_enablement_config` | `not_satisfied` | 不新增 production enablement config |
| `selector_adapter_binding` | `not_satisfied` | 不把 selector 接到 repository adapter |

## Schema Preflight

静态 schema artifact 已能支撑 adapter preflight 和文档评审，但 runtime enablement 还缺：

- 可执行 SQL migration 或等价 migration artifact。
- schema version table / applied migration marker。
- migration runner 与失败可观测性。
- repository mode rollback smoke。

缺任一项时，repository mode 必须返回 `repository_store_disabled` 或 schema failure，不允许静默成功。

## Failure Mapping

准入评审要求继续保留以下 fail-closed 语义：

| 条件 | failure code |
| --- | --- |
| reserved repository mode | `repository_store_disabled` |
| unknown mode | `invalid_draft_store_mode` |
| schema migration 未应用 | `draft_schema_migration_not_applied` |
| store schema version mismatch | `draft_store_schema_version_mismatch` |
| migration 状态不可用 | `draft_store_migration_unavailable` |
| auth context contract mismatch | `draft_auth_context_contract_mismatch` |
| tenant / workspace / application / owner binding 失败 | `draft_tenant_binding_missing` / `draft_workspace_membership_denied` / `draft_application_scope_denied` / `draft_owner_scope_denied` |
| scope grant 缺失 | `draft_scope_grant_missing` |
| repository query executor 不可用 | `draft_store_unavailable` |
| stored record contract mismatch | `draft_store_contract_mismatch` |

失败时不得返回草案主体，不得回退 memory dev、sample 或 fixture，不得创建 run record、confirmation decision、writeback 或 replay。

## Rollback / No Fallback / No Side Effects

- rollback 当前只停留在静态 `rollback-evidence.json`；没有 migration runner rollback action，也没有 repository mode rollback smoke。
- `repository` mode enablement 不允许通过“失败后退回 `memory_dev`”实现可用性。
- 本批不创建数据库连接、不运行 SQL、不调用 OIDC、不写 repository、不启动服务、不创建 production API。
- side effect counters 必须保持：`database_connection_count=0`、`sql_execution_count=0`、`repository_write_count=0`、`executor_call_count=0`、`confirmation_call_count=0`、`business_writeback_count=0`、`replay_call_count=0`。

## 验收方式

- `workflow-saved-draft-repository-mode-enablement-v1` fixture / checker 固定准入结论、config gate、schema preflight、dependency gate、failure mapping、rollback、no fallback、no side effects 和 check-repo 顺序。
- Go tests 继续覆盖 selector fail-closed、repository adapter、adapter smoke 和 production auth runtime bridge。
- 验证链路至少包含 `go test ./internal/httpapi`、repository mode enablement checker、adapter smoke checker、production auth runtime checker、`./scripts/check-repo.sh --fast`；由于本批回写真相源，补跑全量 `./scripts/check-repo.sh`。

## 停止线

- 不启用 repository mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不连接数据库，不写 SQL，不创建 migration runner、schema version table 或 real query executor。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
- 不把 `draft_repository_mode_enablement_review_defined` 解释为 durable persistence ready、repository mode ready、production auth ready、database ready、production API ready 或 production ready。

# Saved Workflow Draft Production Auth Runtime v1

更新时间：2026-06-18

## 专题定位

`Saved Workflow Draft Production Auth Runtime v1` 承接 [Saved Workflow Draft Production Auth Readiness v1](saved-workflow-draft-production-auth-readiness-v1.md)、repository adapter implementation 和 adapter smoke execution，把已验证的 Radish auth context 与已验证 workspace / application / owner binding 投影为 `SavedWorkflowDraftRepositoryActorContext`。

本专题只实现 runtime bridge：`BuildSavedWorkflowDraftRepositoryActorContextForOperation` 和 `BuildSavedWorkflowDraftRepositoryActorContextFromProductionAuth`。它不创建 OIDC middleware、不校验 token、不创建 membership adapter、不启用 repository store mode、不连接数据库、不创建 production API，也不打开 publish、run、executor、confirmation、writeback 或 replay。

状态：`draft_production_auth_runtime_bridge_implemented`

## 当前输入事实

- `draft_production_auth_readiness_defined` 已固定 issuer / token / claim / tenant / workspace / scope / failure mapping 证据链。
- `draft_repository_adapter_implemented` 已实现 repository interface、query executor adapter、schema preflight 和 auth context failure mapping。
- `draft_adapter_smoke_executed` 已用 static contract smoke cases、repository adapter 和 injected fake query executor 执行 save / read / list。
- 当前仍没有真实 OIDC client、issuer network call、token validation、workspace membership adapter、repository mode enablement、数据库连接或 production API。

## Runtime Bridge

新增类型：

- `SavedWorkflowDraftVerifiedAuthContext`：承载已由上游认证层验证过的 request、auth source、tenant、actor subject、scope grants 和 audit ref。
- `SavedWorkflowDraftWorkspaceBinding`：承载已验证 tenant / workspace / application / owner binding，以及 workspace membership、application scope 和 owner scope 的 verified 标记。
- `SavedWorkflowDraftProductionAuthRuntimeResult`：返回 repository actor context 或 fail-closed failure code。

新增函数：

- `BuildSavedWorkflowDraftRepositoryActorContextForOperation`：按 `SaveWorkflowDraftRecord` / `ReadWorkflowDraftRecord` / `ListWorkflowDraftRecords` 选择 required scope。
- `BuildSavedWorkflowDraftRepositoryActorContextFromProductionAuth`：归一化字段、投影 `SavedWorkflowDraftRepositoryActorContext`，并在缺失或不可信时返回 failure code。

当前唯一允许的 auth source 是 `radish_oidc_verified_context`。这只是上游已验证上下文的来源标记，不代表本批实现 OIDC token validation。

## Operation Scope Matrix

| operation | required scope | 结果 |
| --- | --- | --- |
| `SaveWorkflowDraftRecord` | `workflow_drafts:write` | 可投影 repository actor context |
| `ReadWorkflowDraftRecord` | `workflow_drafts:read` | 可投影 repository actor context |
| `ListWorkflowDraftRecords` | `workflow_drafts:read` | 可投影 repository actor context |

未知 operation 返回 `draft_auth_context_contract_mismatch`，不回退默认 scope。

## Failure Mapping

| 条件 | failure code |
| --- | --- |
| dev fake auth source / unknown auth source | `draft_auth_context_contract_mismatch` |
| request id 缺失 | `draft_auth_context_contract_mismatch` |
| actor subject 缺失 | `draft_identity_context_missing` |
| audit ref 缺失 | `draft_audit_context_missing` |
| tenant 缺失或 binding mismatch | `draft_tenant_binding_missing` |
| workspace 缺失或 membership 未验证 | `draft_workspace_membership_denied` |
| application 缺失或 scope 未验证 | `draft_application_scope_denied` |
| owner 缺失或 scope 未验证 | `draft_owner_scope_denied` |
| operation scope 缺失 | `draft_scope_grant_missing` |

失败时不返回草案主体、不写 repository、不回退 dev auth、test auth、memory dev store、sample 或 fixture。

## 验收方式

- Go tests 覆盖 actor context 投影、operation scope 和 failure mapping。
- `workflow-saved-draft-production-auth-runtime-v1` fixture / checker 固定状态、源码边界、failure mapping、no fake fallback、no side effects、文档引用和 check-repo 顺序。
- 验证链路至少包含 `go test ./internal/httpapi`、专项 checker、production auth readiness checker、adapter smoke checker 和 `./scripts/check-repo.sh --fast`。

## 停止线

- 不创建 OIDC middleware、token validator、membership adapter、repository mode enablement、production API、SQL migration、migration runner、数据库连接或 schema version table。
- 不把 `draft_production_auth_runtime_bridge_implemented` 解释为 Radish OIDC ready、token validation ready、membership adapter ready、repository mode ready、durable persistence ready、production API ready 或 production ready。
- 不实现 publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。

## 下一批建议

下一步若继续 durable store，应进入 `Saved Workflow Draft Repository Mode Enablement` 的准入评审，先判断 schema artifact、selector、repository adapter、adapter smoke 和 production auth runtime bridge 是否足以打开 repository mode 的受控入口；真实数据库连接、SQL migration runner、OIDC middleware、membership adapter 和 production API 仍需独立专题。

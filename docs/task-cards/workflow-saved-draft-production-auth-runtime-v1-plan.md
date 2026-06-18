# Workflow Saved Draft Production Auth Runtime v1

更新时间：2026-06-18

## 任务定位

任务 ID：`workflow-saved-draft-production-auth-runtime-v1`

当前状态：`draft_production_auth_runtime_bridge_implemented`

本任务承接 `Saved Workflow Draft Production Auth Readiness v1`、repository adapter implementation 和 adapter smoke execution，实现已验证 auth context 到 `SavedWorkflowDraftRepositoryActorContext` 的 runtime bridge。它只固定生产权限上下文进入 repository adapter 前的纯函数投影，不创建真实 OIDC middleware、token validation、membership adapter、repository mode、数据库连接或 production API。

## 输入

- `workflow-saved-draft-production-auth-readiness-v1`
- `workflow-saved-draft-auth-context-preconditions-v1`
- `workflow-saved-draft-repository-adapter-implementation-v1`
- `workflow-saved-draft-adapter-smoke-v1`

## 实现范围

- 新增 `SavedWorkflowDraftVerifiedAuthContext`。
- 新增 `SavedWorkflowDraftWorkspaceBinding`。
- 新增 `SavedWorkflowDraftProductionAuthRuntimeResult`。
- 新增 `BuildSavedWorkflowDraftRepositoryActorContextForOperation`。
- 新增 `BuildSavedWorkflowDraftRepositoryActorContextFromProductionAuth`。
- 新增 Go tests 覆盖成功投影、operation scope 和 failure mapping。
- 新增 fixture / checker / guard，并接入 `check-repo.py`。

## Operation Matrix

| operation | required scope | 说明 |
| --- | --- | --- |
| `SaveWorkflowDraftRecord` | `workflow_drafts:write` | 投影 write scope 到 repository actor context |
| `ReadWorkflowDraftRecord` | `workflow_drafts:read` | 投影 read scope 到 repository actor context |
| `ListWorkflowDraftRecords` | `workflow_drafts:read` | 投影 read scope 到 repository actor context |

## Failure Mapping

- `draft_auth_context_contract_mismatch`：auth source、request id、operation 或 required scope contract 不满足。
- `draft_identity_context_missing`：actor subject 缺失。
- `draft_audit_context_missing`：audit ref 缺失。
- `draft_tenant_binding_missing`：tenant 缺失或 binding mismatch。
- `draft_workspace_membership_denied`：workspace 缺失或 membership 未验证。
- `draft_application_scope_denied`：application 缺失或 scope 未验证。
- `draft_owner_scope_denied`：owner 缺失或 scope 未验证。
- `draft_scope_grant_missing`：operation required scope 缺失。

## 验证

- `go test ./internal/httpapi`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-production-auth-runtime-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-production-auth-readiness-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-adapter-smoke-v1.py`
- `./scripts/check-repo.sh --fast`

## 停止线

- 不创建 OIDC middleware、token validation、membership adapter、repository store mode enablement、production API、SQL migration、migration runner、数据库连接或 schema version table。
- 不写 repository、不调用 query executor、不启动服务、不访问网络。
- 不实现 publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。

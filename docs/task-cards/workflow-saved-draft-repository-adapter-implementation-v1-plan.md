# Workflow Saved Draft Repository Adapter Implementation v1 任务卡

更新时间：2026-06-18

## 任务标识

- 任务 ID：`workflow-saved-draft-repository-adapter-implementation-v1`
- 状态：`draft_repository_adapter_implementation_task_card_defined`

## 任务定位

本任务卡承接 `Saved Workflow Draft Repository Adapter Implementation Entry Review v1`，用于固定下一批 repository adapter implementation 的实现边界、验证链路和停止线。

当前只定义实现任务卡，不创建 repository interface、repository adapter、database query、adapter unit tests、adapter smoke fixture、SQL migration、migration runner、OIDC middleware、token validation、membership adapter、production API、publish、run、executor、confirmation、writeback 或 replay。

## 输入

- `docs/features/workflow/saved-workflow-draft-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-adapter-implementation-plan-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-adapter-implementation-entry-review-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-contract-preconditions-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-contract-smoke-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-contract-smoke-runner-implementation-v1.md`
- `docs/features/workflow/saved-workflow-draft-store-selector-implementation-v1.md`
- `docs/features/workflow/saved-workflow-draft-schema-artifact-materialization-v1.md`
- `docs/features/workflow/saved-workflow-draft-production-auth-readiness-v1.md`
- `docs/features/workflow/saved-workflow-draft-adapter-smoke-readiness-v1.md`
- `services/platform/internal/httpapi/workflow_saved_draft.go`
- `services/platform/internal/httpapi/workflow_saved_draft_store_selector.go`
- `services/platform/internal/httpapi/workflow_saved_draft_repository_contract_smoke_runner.go`
- `services/platform/migrations/workflow_saved_drafts/manifest.json`

## 后续实现范围

后续代码实现批次只允许打开以下内容：

- `SavedWorkflowDraftRepository` interface，覆盖 `SaveWorkflowDraftRecord`、`ReadWorkflowDraftRecord` 和 `ListWorkflowDraftRecords`。
- repository actor context 输入校验，必须消费 `tenant_ref`、`workspace_id`、`application_id`、`actor_subject_ref`、`owner_subject_ref` 和 `scope_grants`。
- adapter 边界和可测试 query policy，使用注入式 query executor / transaction boundary，不创建进程级数据库连接，不启动真实数据库。
- schema preflight，消费静态 schema artifact manifest，不运行 SQL migration，不写 schema version table。
- sanitized projection，确保 read / list 不返回 secret、token、完整 claim、内部数据库 detail 或 sample payload。
- adapter unit tests，使用 fake query executor 覆盖 save / read / list、version conflict、scope denied、not found、schema mismatch、store unavailable、contract mismatch、no fallback 和 no side effects。

建议文件落点：

- `services/platform/internal/httpapi/workflow_saved_draft_repository.go`
- `services/platform/internal/httpapi/workflow_saved_draft_repository_adapter.go`
- `services/platform/internal/httpapi/workflow_saved_draft_repository_adapter_test.go`

## Operation Matrix

| operation | scope | 实现要求 |
| --- | --- | --- |
| `SaveWorkflowDraftRecord` | `workflow_drafts:write` | 校验 actor context、tenant / workspace / application predicate、owner、schema preflight、version compare-and-set、sanitized payload 和 audit metadata |
| `ReadWorkflowDraftRecord` | `workflow_drafts:read` | 校验 actor context、scope predicate、owner visibility、schema preflight、not found no sample fallback 和 sanitized projection |
| `ListWorkflowDraftRecords` | `workflow_drafts:read` | 校验 workspace / application predicate、owner visibility、stable ordering、summary projection、empty list no sample fallback 和 schema preflight |

## Failure Mapping

后续实现必须保留以下 fail-closed code：

- `draft_scope_denied`
- `draft_not_found`
- `draft_schema_version_unsupported`
- `draft_payload_invalid`
- `draft_version_conflict`
- `draft_store_unavailable`
- `draft_store_contract_mismatch`
- `repository_store_disabled`
- `invalid_draft_store_mode`
- `draft_auth_context_contract_mismatch`
- `draft_schema_migration_not_applied`
- `draft_store_schema_version_mismatch`
- `draft_store_migration_unavailable`
- `draft_identity_context_missing`
- `draft_tenant_binding_missing`
- `draft_workspace_membership_denied`
- `draft_scope_grant_missing`

`draft_version_conflict` 不得降级为 generic adapter failure；`draft_not_found` 不得回退 sample；`draft_scope_denied` 不得返回草案主体；schema、auth 或 store failure 不得泄露数据库、OIDC、token、claim dump 或内部 adapter detail。

## 验收方式

后续代码实现完成时至少运行：

```bash
cd services/platform
go test ./internal/httpapi
cd ../..
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-contract-smoke-runner-implementation-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-store-selector-smoke-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-artifact-materialization-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-production-auth-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-adapter-smoke-readiness-v1.py
./scripts/check-repo.sh --fast
```

若后续实现新增 fixture / checker、schema 真相源或阶段口径变化，补跑全量：

```bash
./scripts/check-repo.sh
```

## 停止线

- 不启用 `repository` store mode；当前 selector 仍只允许 `memory_dev` 成功，`repository_disabled` / `repository` / unknown mode 必须 fail closed。
- 不创建 SQL migration、schema version table、migration runner、真实数据库连接或外部数据库 fixture。
- 不创建 adapter smoke fixture、adapter smoke checker 或 adapter smoke execution 入口。
- 不创建 Radish OIDC middleware、token validation、session cookie、workspace membership adapter 或 production auth runtime。
- 不创建 public production API consumer、production CORS policy、API key lifecycle、quota enforcement、billing 或 cost ledger。
- 不实现 durable publish、run、workflow executor、node executor、tool executor、agent loop、confirmation decision、business writeback、replay、resume 或 materialized result reader。
- 不把 `draft_repository_adapter_implementation_task_card_defined` 解释为 repository adapter ready、database ready、repository mode ready、adapter smoke ready、OIDC ready、production API ready 或 production ready。

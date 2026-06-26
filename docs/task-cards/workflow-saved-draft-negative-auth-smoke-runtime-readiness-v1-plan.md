# Workflow Saved Draft Negative Auth Smoke Runtime Readiness v1 任务卡

状态：`draft_negative_auth_smoke_runtime_readiness_defined`

## 目标

固定 Saved Workflow Draft future negative auth smoke runtime 的 readiness，明确负向 auth / membership 场景、failure code、脱敏诊断、no fallback、no side effects 和 artifact guard。

本任务卡不创建 negative auth smoke runtime fixture / checker / runner，不创建 auth middleware / membership adapter implementation task card，不实现 OIDC middleware、token validator、auth middleware、membership adapter、repository mode runtime、database runtime 或 production API。

## 输入

- `draft_auth_middleware_membership_adapter_task_card_entry_readiness_defined`
- `draft_token_validation_schema_artifact_implemented`
- `draft_token_validation_auth_middleware_runtime_entry_review_defined`
- `radish_oidc_token_membership_upstream_evidence_refresh_defined`
- `draft_production_auth_runtime_bridge_implemented`
- `draft_database_secret_resolver_runtime_dependency_refresh_defined`
- `draft_schema_marker_runtime_dependency_refresh_defined`
- `draft_repository_mode_runtime_boundary_review_defined`

## 本批交付

1. 新增 feature doc：`docs/features/workflow/saved-workflow-draft-negative-auth-smoke-runtime-readiness-v1.md`。
2. 新增 fixture：`scripts/checks/fixtures/workflow-saved-draft-negative-auth-smoke-runtime-readiness-v1.json`。
3. 新增 checker：`scripts/checks/control_plane/check-workflow-saved-draft-negative-auth-smoke-runtime-readiness-v1.py`。
4. 将 checker 接入 `./scripts/check-repo.sh --fast`。
5. 同步 `docs/radishmind-current-focus.md`、`docs/features/README.md`、`docs/features/workflow/README.md`、`docs/features/workflow-agent-runtime.md`、`docs/features/workflow/saved-workflow-draft-v1.md`、`docs/radishmind-integration-contracts.md`、`docs/task-cards/README.md`、`scripts/README.md` 和周志。

## 准入判断

- negative auth case matrix：已满足静态 readiness，future runtime smoke 必须覆盖 13 个负向场景。
- failure mapping：必须 fail closed before repository query，并映射到 Saved Workflow Draft 既有 failure code。
- sanitized diagnostics：只允许 request / audit reference、policy version、case id 和 failure code。
- future runtime artifact：本批只声明路径，`workflow-saved-draft-negative-auth-smoke-runtime-v1` fixture / checker 仍必须不存在。
- auth middleware / membership adapter implementation task card：仍 blocked，不能由本批 readiness 直接创建。

## 验证

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-negative-auth-smoke-runtime-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-auth-middleware-membership-adapter-task-card-entry-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-token-validation-schema-artifact-implementation-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-radish-oidc-token-membership-upstream-evidence-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-production-auth-runtime-v1.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 `scripts/checks/fixtures/workflow-saved-draft-negative-auth-smoke-runtime-v1.json`。
- 不创建 `scripts/checks/control_plane/check-workflow-saved-draft-negative-auth-smoke-runtime-v1.py`。
- 不创建 auth middleware / membership adapter implementation task card。
- 不创建 OIDC middleware、token validator、auth middleware、membership adapter、membership cache、repository mode runtime、database runtime、production API、executor、confirmation、writeback 或 replay。
- 不 fetch issuer discovery，不下载 JWKS，不校验真实 token，不查询 membership，不连接数据库，不运行 SQL，不读写 schema marker，不解析 secret。
- 不把本批 readiness 写成 auth runtime ready、membership ready、repository mode ready、database ready、production API ready 或 production ready。

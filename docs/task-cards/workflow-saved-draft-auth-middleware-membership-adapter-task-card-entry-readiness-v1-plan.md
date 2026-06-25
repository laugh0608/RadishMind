# Workflow Saved Draft Auth Middleware / Membership Adapter Task Card Entry Readiness v1 任务卡

状态：`draft_auth_middleware_membership_adapter_task_card_entry_readiness_defined`

## 目标

固定 Saved Workflow Draft future auth middleware / membership adapter implementation task card 的准入要求，明确哪些输入已满足、哪些 runtime 前置仍 blocked，以及下一批应优先补齐的 negative auth smoke runtime readiness。

本任务卡不创建 auth middleware / membership adapter implementation task card，不实现 OIDC middleware、token validator、auth middleware、membership adapter、negative auth smoke runtime、repository mode runtime、database runtime 或 production API。

## 输入

- `draft_token_validation_schema_artifact_implemented`
- `draft_token_validation_auth_middleware_runtime_entry_review_defined`
- `radish_oidc_token_membership_upstream_evidence_refresh_defined`
- `draft_production_auth_runtime_bridge_implemented`
- `draft_database_secret_resolver_runtime_dependency_refresh_defined`
- `draft_schema_marker_runtime_dependency_refresh_defined`
- `draft_repository_mode_runtime_boundary_review_defined`

## 本批交付

1. 新增 feature doc：`docs/features/workflow/saved-workflow-draft-auth-middleware-membership-adapter-task-card-entry-readiness-v1.md`。
2. 新增 fixture：`scripts/checks/fixtures/workflow-saved-draft-auth-middleware-membership-adapter-task-card-entry-readiness-v1.json`。
3. 新增 checker：`scripts/checks/control_plane/check-workflow-saved-draft-auth-middleware-membership-adapter-task-card-entry-readiness-v1.py`。
4. 将 checker 接入 `./scripts/check-repo.sh --fast`。
5. 同步 `docs/radishmind-current-focus.md`、`docs/features/README.md`、`docs/features/workflow/README.md`、`docs/features/workflow-agent-runtime.md`、`docs/features/workflow/saved-workflow-draft-v1.md`、`docs/task-cards/README.md`、`scripts/README.md` 和周志。

## 准入判断

- token validation schema artifact：已满足，可作为 future validator 输出契约。
- auth middleware ownership：静态 owner contract 已满足，可写入后续任务卡要求。
- membership adapter contract：tenant / workspace / application / owner / scope projection 边界已满足，可写入后续任务卡要求。
- negative auth smoke runtime readiness：未完成，后续实现任务卡仍 blocked。
- repository actor context handoff：复用 existing bridge，不允许新增私有 handoff runtime。
- repository mode / DB runtime / schema marker / secret resolver：仍 blocked，不能由 auth/membership task card 绕过。

## 验证

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-auth-middleware-membership-adapter-task-card-entry-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-token-validation-schema-artifact-implementation-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-radish-oidc-token-membership-upstream-evidence-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-production-auth-runtime-v1.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 auth middleware / membership adapter implementation task card。
- 不创建 OIDC middleware、token validator、auth middleware、membership adapter、membership cache、negative auth smoke runtime、login / logout route、session cookie、repository mode runtime、database runtime、production API、executor、confirmation、writeback 或 replay。
- 不 fetch issuer discovery，不下载 JWKS，不校验真实 token，不查询 membership，不连接数据库，不运行 SQL，不读写 schema marker，不解析 secret。
- 不把本批 readiness 写成 auth runtime ready、membership ready、repository mode ready、database ready、production API ready 或 production ready。

# Workflow Saved Draft Token Validation Schema / Auth Middleware Runtime Entry Review v1 任务卡

状态：`draft_token_validation_auth_middleware_runtime_entry_review_defined`

## 背景

Saved Workflow Draft durable store 已完成 Radish OIDC token / membership readiness、implementation entry review、upstream evidence refresh、production auth runtime bridge、repository mode boundary、schema marker runtime dependency refresh 和 database secret resolver runtime dependency refresh。当前缺口集中在 token validation schema、auth middleware、membership adapter、negative auth smoke 与 repository actor context handoff 能否进入 runtime 实现任务卡。

本任务卡只承接 runtime entry review。本批不创建 schema、middleware、token validator、membership adapter、runtime smoke、repository mode、database runtime 或 production API。

## 输入

- `docs/integrations/radish-oidc-token-membership-readiness-v1.md`
- `docs/integrations/radish-oidc-token-membership-implementation-entry-review-v1.md`
- `docs/integrations/radish-oidc-token-membership-upstream-evidence-refresh-v1.md`
- `docs/features/workflow/saved-workflow-draft-production-auth-runtime-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-mode-runtime-boundary-review-v1.md`
- `docs/features/workflow/saved-workflow-draft-schema-marker-runtime-dependency-refresh-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-secret-resolver-runtime-dependency-refresh-v1.md`

## 输出

- 新增 `Saved Workflow Draft Token Validation Schema / Auth Middleware Runtime Entry Review v1` 细专题。
- 新增 `workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1` fixture / checker。
- 更新 workflow 入口、features 入口、Saved Workflow Draft v1、Workflow / Agent Runtime、current focus、task card index、scripts README 和周志。
- 将 checker 接入 `./scripts/check-repo.sh --fast`。

## 验收

- review matrix 覆盖 token validation schema、auth middleware ownership、membership adapter、negative auth smoke、repository actor context handoff、failure mapping、no fallback、no side effects 和 artifact guard。
- failure mapping 固定 `draft_auth_context_contract_mismatch`、`draft_identity_context_missing`、`draft_tenant_binding_missing`、`draft_workspace_membership_denied`、`draft_application_scope_denied`、`draft_owner_scope_denied`、`draft_scope_grant_missing`、`draft_audit_context_missing`、`repository_store_disabled`、`invalid_draft_store_mode`、`draft_store_unavailable` 和 `draft_store_migration_unavailable`。
- artifact guard 能检测 token schema、OIDC middleware、token validator、membership adapter、negative auth smoke runtime、repository mode、database runtime、production API、executor、confirmation、writeback 或 replay artifact 是否被提前创建。

## 验证

本批完成后运行：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-radish-oidc-token-membership-upstream-evidence-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-production-auth-runtime-v1.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 token validation schema implementation task card、auth middleware implementation task card、membership adapter implementation task card 或 runtime smoke task card。
- 不创建 `contracts/radish-oidc-token-validation.schema.json`、OIDC middleware、token validator、membership adapter、membership cache、negative auth smoke runtime、login / logout route、session cookie、repository mode runtime、database runtime、production API、executor、confirmation、writeback 或 replay。
- 不 fetch issuer discovery，不下载 JWKS，不校验 token，不查询 membership，不连接数据库，不运行 SQL，不读写 schema marker，不解析 secret。

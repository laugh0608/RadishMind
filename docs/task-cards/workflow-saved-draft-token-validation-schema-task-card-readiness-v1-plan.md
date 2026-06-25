# Workflow Saved Draft Token Validation Schema Task Card Readiness v1 任务卡

状态：`draft_token_validation_schema_task_card_readiness_defined`

## 背景

`Saved Workflow Draft Token Validation Schema / Auth Middleware Runtime Entry Review v1` 已确认 auth runtime task card 仍 blocked，但后续可以先拆出 token validation schema implementation task card readiness。该方向只评审 schema artifact 任务卡是否具备创建条件，不实现 token validation runtime。

本任务卡只承接 schema task card readiness。本批不创建 `contracts/radish-oidc-token-validation.schema.json`，不创建 schema implementation task card，不实现 OIDC middleware、token validator、auth middleware、membership adapter、negative auth smoke runtime、repository mode、database runtime 或 production API。

## 输入

- `docs/features/workflow/saved-workflow-draft-token-validation-auth-middleware-runtime-entry-review-v1.md`
- `docs/integrations/radish-oidc-token-membership-upstream-evidence-refresh-v1.md`
- `docs/features/workflow/saved-workflow-draft-production-auth-runtime-v1.md`
- `docs/features/workflow/saved-workflow-draft-production-auth-readiness-v1.md`

## 输出

- 新增 `Saved Workflow Draft Token Validation Schema Task Card Readiness v1` 细专题。
- 新增 `workflow-saved-draft-token-validation-schema-task-card-readiness-v1` fixture / checker。
- 更新 workflow 入口、features 入口、Saved Workflow Draft v1、Workflow / Agent Runtime、current focus、task card index、scripts README 和周志。
- 将 checker 接入 `./scripts/check-repo.sh --fast`。

## 验收

- readiness matrix 覆盖 prior runtime entry review、upstream issuer evidence、verified token output contract、schema implementation task card、schema artifact、token validator runtime、auth middleware runtime、membership adapter runtime、negative auth smoke runtime 和 repository mode runtime。
- future task card requirements 固定 schema scope、required fields、forbidden raw-material fields、sanitized failure envelope、consumer handoff 和 artifact guard。
- artifact guard 能检测 schema 文件、OIDC middleware、token validator、membership adapter、negative auth smoke runtime、repository mode、database runtime、production API、executor、confirmation、writeback 或 replay artifact 是否被提前创建。

## 验证

本批完成后运行：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-token-validation-schema-task-card-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-radish-oidc-token-membership-upstream-evidence-refresh-v1.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 token validation schema implementation task card。
- 不创建 `contracts/radish-oidc-token-validation.schema.json`、OIDC middleware、token validator、auth middleware、membership adapter、membership cache、negative auth smoke runtime、login / logout route、session cookie、repository mode runtime、database runtime、production API、executor、confirmation、writeback 或 replay。
- 不 fetch issuer discovery，不下载 JWKS，不校验 token，不查询 membership，不连接数据库，不运行 SQL，不读写 schema marker，不解析 secret。

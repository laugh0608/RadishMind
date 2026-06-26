# Workflow Saved Draft Manual Migration Runner Implementation Entry Refresh v1 任务卡

状态：`draft_manual_migration_runner_implementation_entry_refresh_defined`

## 背景

Saved Workflow Draft durable store 已完成 schema marker contract implementation entry review，确认 marker reader / writer contract implementation task card 仍 blocked。本任务卡承接其后续方向，复评 manual migration runner 是否具备进入实现任务卡的条件。

本任务卡只固定 implementation entry refresh，不实现 runner，不创建 runner command、dry-run output、SQL migration、schema marker runtime、database connection provider 或 repository mode runtime。

## 输入

- `docs/features/workflow/saved-workflow-draft-schema-marker-contract-implementation-entry-review-v1.md`
- `docs/features/workflow/saved-workflow-draft-schema-marker-migration-runner-readiness-refresh-v1.md`
- `docs/features/workflow/saved-workflow-draft-schema-migration-runner-implementation-entry-review-v1.md`
- `docs/features/workflow/saved-workflow-draft-schema-artifact-materialization-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-connection-provider-implementation-entry-review-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-secret-resolver-implementation-entry-review-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-mode-runtime-boundary-review-v1.md`

## 输出

- 新增 `Saved Workflow Draft Manual Migration Runner Implementation Entry Refresh v1` 细专题。
- 新增 `workflow-saved-draft-manual-migration-runner-implementation-entry-refresh-v1` fixture / checker。
- 更新 workflow 入口、features 入口、Saved Workflow Draft v1、Workflow / Agent Runtime、current focus、产品范围、能力矩阵、task card index 和周志。
- 将 checker 接入 `./scripts/check-repo.sh --fast`。

## 验收

- entry refresh 明确记录 manual migration runner implementation task card 当前不创建。
- candidate matrix 覆盖 migration artifact / apply plan、dry-run output contract、apply result envelope、idempotency / lock policy、rollback / forward-only observability 和 implementation task card。
- runner contract 固定 manual trigger、dry-run metadata-only output、apply result allowlist、duplicate handling、repository adapter 只消费 applied marker 和 sanitized diagnostics。
- no fallback、no side effects 和 artifact guard 能检测 runner task card、runner runtime、runner command、dry-run output、SQL、schema marker、DB provider / driver、OIDC、membership、production API、executor、confirmation、writeback 或 replay artifact 是否被提前创建。

## 验证

本批完成后运行：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-manual-migration-runner-implementation-entry-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-marker-contract-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-marker-migration-runner-readiness-refresh-v1.py
./scripts/check-repo.sh --fast
```

由于本批同步阶段真相源和 `check-repo.py` 注册顺序，若时间允许补跑：

```bash
./scripts/check-repo.sh
```

## 停止线

- 不创建 manual migration runner implementation task card。
- 不创建 database connection provider、database secret resolver、DB driver、connection factory、connection smoke、database query executor、schema marker reader / writer、SQL migration、migration runner、runner command 或 dry-run output。
- 不启用 `repository` store mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。

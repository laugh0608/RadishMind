# Workflow RAG Regression Review 与 Evaluation Profile（开发 / 测试态）v1 实施任务卡

更新时间：2026-07-18

状态：`completed`

## 目标

按 [功能设计](../features/workflow/workflow-rag-regression-review-evaluation-profile-dev-test-v1.md) 将 `workflow_run_record.v3` 接入既有 Run Comparison、Evaluation Cases、Baseline / Case Versioning 和 Evaluation Suite / Release Review，并补齐 SQLite evaluation resources durable store。全过程保持 metadata-only，不执行 retrieval / Gateway，不改变 HTTP Tool v2、executor v0 或生产能力边界。

## 批次 A：Comparison v2 与评测领域接入

1. 定义 v3 comparison compatibility 和 `workflow_run_retrieval_profile_incompatible`。
2. 实现 `workflow_run_comparison.v2` retrieval diff、citation loss 分类和稳定 findings。
3. Evaluation create / revise / review 接受同 binding v3，返回 run profile 与 comparison schema。
4. Suite 接受 retrieval case，review item / digest 纳入 run profile。
5. 保持 v0 / v1 Comparison v1 与 HTTP Tool v2 unsupported 行为。

完成条件：Go domain / HTTP 精准测试、metadata-only 序列化和 Platform 相关包通过。

## 批次 B：SQLite durable evaluation resources

1. 增加 SQLite `0006_workflow_evaluation_resources` migration。
2. 实现 SQLite case / revision repository 和 suite / decision repository。
3. factory 对 `sqlite_dev` 显式选择 SQLite，未知实现失败关闭。
4. 覆盖 CAS、scope、游标、重启恢复、损坏记录、关闭和 no-fallback。
5. PostgreSQL 复用既有表，补 v3 case / suite 重启和 no-fallback 测试，不新增空 migration。

完成条件：SQLite migration / repository 测试、PostgreSQL 专项测试和 Platform 全包通过。

## 批次 C：Web 审查与连续验收

1. strict consumer 接受 Comparison v1 / v2，并拒绝 query、fragment content、answer、prompt 和 credential。
2. Run History 为 v3 提供兼容 baseline / candidate 选择与 retrieval diff 面板。
3. Evaluation / Baseline / Suite 显示 retrieval profile、comparison schema 和 finding codes。
4. 覆盖 offline 零请求、旧 run 行为、Web tests / build、SQLite / PostgreSQL 重启与 no-fallback。
5. 更新功能入口、current focus、路线图、能力矩阵和 2026-W29 周志。

完成条件：双数据库连续链、必要真实浏览器验收、敏感内容扫描和仓库 fast / full 通过。

## 关键文件

- Platform：`services/platform/internal/httpapi/workflow_run_comparison.go`、`workflow_evaluation*.go`。
- SQLite：`services/platform/migrations/sqlite/workflow_runs/` 与 evaluation SQLite repositories。
- PostgreSQL：既有 workflow run migration / pool / evaluation repositories。
- Web：`workflowRunComparison*`、`workflowEvaluation*`、`workflowRunHistoryPanel.tsx`。
- 文档：本任务卡、功能专题、Workflow 入口、current focus、roadmap、capability matrix、W29 周志。

## 必要验证

```bash
cd services/platform
go test ./internal/httpapi ./migrations/sqlite/workflow_runs
go test ./...

cd ../..
./scripts/run-workflow-saved-draft-postgres-dev-test.sh check

cd apps/radishmind-web
npm test
npm run build

cd ../..
git diff --check
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```

## 停止线

- 不重新执行 retrieval / Gateway，不做 replay / retry / resume。
- 不持久化 query、fragment、answer、prompt、credential 或 provider raw response。
- 不放开 HTTP Tool v2 comparison、executor v0 retrieval 或生产能力。
- 不新增第二张任务卡、同层 readiness 文档或独立 checker 链。

## 完成记录

- 批次 A：`workflow_rag_regression_review_evaluation_profile_dev_test_v1_batch_a_completed`。
- 批次 B：`workflow_rag_regression_review_evaluation_profile_dev_test_v1_batch_b_completed`。
- 批次 C：`workflow_rag_regression_review_evaluation_profile_dev_test_v1_batch_c_completed`。
- 总状态：`workflow_rag_regression_review_evaluation_profile_dev_test_v1_completed`。
- 完成范围包括 Comparison v2、Evaluation / Baseline / Suite profile、SQLite durable evaluation resources、PostgreSQL 重启 / no-fallback、Web metadata-only 审查与仓库级验证；未打开任何停止线外能力。
- 浏览器证据覆盖同 binding v3 比较、case / exact-version suite review、SQLite 服务重启恢复和 run 数量不变；验收中修正 draft digest 误纳入读时 audit 投影的根因，并补充稳定性测试，未重写历史 run。

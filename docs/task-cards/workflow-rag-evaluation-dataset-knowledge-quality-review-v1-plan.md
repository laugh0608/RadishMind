# Workflow RAG 评测数据集与知识质量审查 v1 实施任务卡

更新时间：2026-07-18

状态：`workflow_rag_evaluation_dataset_knowledge_quality_review_v1_completed`

## 目标

按[功能设计](../features/workflow/workflow-rag-evaluation-dataset-knowledge-quality-review-v1.md)交付 synthetic-public、精确 snapshot / profile binding 的离线 RAG 评测链。复用现有 lexical ranker、snapshot digest 和 profile digest，生成 metadata-only quality review；不执行 Workflow / Gateway，不创建 run，也不新增数据库、HTTP route 或 Web 写面。

## 批次 A：契约、领域评审器与 CLI（已完成）

1. 新增 `workflow_rag_evaluation_dataset.v1` 与 `workflow_rag_quality_review.v1` JSON Schema。
2. 在独立领域文件实现 strict decode、dataset digest、binding、sample、expected citation evidence 和阈值校验。
3. 对每个样本恰好调用一次 `RankWorkflowRAGFragments`，计算 hit@k、expected recall、required official recall、MRR、no-evidence accuracy 与 sample pass rate。
4. 生成 duplicate content、official absence、fragment expectation coverage 等稳定 findings。
5. 新增 `cmd/radishmind-workflow-rag-eval`，支持 generate 与 `--check`，命令输出保持 metadata-only。
6. 测试 strict JSON、digest / binding、positive / negative、指标、threshold、finding、determinism 和敏感内容禁入。

完成证据：strict contract、digest / binding、一次 ranker 调用、正负样本、指标、threshold、finding、确定性和 metadata-only 测试已通过；CLI generate / check 已通过。

## 批次 B：starter dataset 与仓库接线（已完成）

1. 提交 synthetic-public snapshot fixture，覆盖 Latin、CJK、official / supplemental source 和稳定 tie-break。
2. 提交不少于 6 个样本，覆盖单证据、多证据、official required、CJK、tie-break 与 no-evidence。
3. 生成并提交 deterministic metadata-only report，状态必须为 `passed`。
4. `--check` 复算 committed report；schema validator 覆盖 snapshot、dataset 与 report。
5. 更新 evaluation README、功能入口、Workflow 入口、current focus、roadmap、能力矩阵、任务索引和 W29 周志。

完成证据：已提交 6 个 synthetic-public fragment、7 个样本与六项指标均为 `1` 的 `passed` report；committed report check、contract schema validation、Go 全包、`git diff --check`、仓库 fast / full 结果记录于 W29 周志。

## 关键文件

- 契约：`contracts/workflow-rag-evaluation-dataset.schema.json`、`contracts/workflow-rag-quality-review.schema.json`。
- 领域：`services/platform/internal/httpapi/workflow_rag_quality_review.go`。
- CLI：`services/platform/cmd/radishmind-workflow-rag-eval/`。
- 数据：`datasets/eval/workflow-rag/`。
- 文档：本任务卡、功能专题、Workflow / evaluation 入口、current focus、roadmap、capability matrix、W29 周志。

## 必要验证

```bash
cd services/platform
go test ./internal/httpapi ./cmd/radishmind-workflow-rag-eval
go test ./...
go vet ./...
go run ./cmd/radishmind-workflow-rag-eval \
  --snapshot ../../datasets/eval/workflow-rag/snapshots/dev_core_v1.json \
  --dataset ../../datasets/eval/workflow-rag/datasets/dev_core_v1.json \
  --output ../../datasets/eval/workflow-rag/reports/dev_core_v1.review.json \
  --check

cd ../..
git diff --check
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```

## 停止线

- 不复制 lexical 算法到 Python、Web 或第二套 provider。
- 不保存真实用户 query、workspace_internal fragment、credential、prompt、answer 或模型响应。
- 不调用 Gateway，不创建 run，不修改 snapshot / Evaluation / Suite。
- 不新增 HTTP route、store mode、migration、Web 写面或后台 job。
- 不接 crawler、connector、online search、embedding、vector、reranker、自动 baseline / release 或生产能力。
- 不新增第二张 task card 或同层 readiness / checker 文档链。

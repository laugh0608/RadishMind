# Workflow RAG 离线评测资产

本目录承载 `Workflow RAG 评测数据集与知识质量审查 v1` 的 synthetic-public、可提交、可复算资产：

- `snapshots/`：严格符合 `workflow_rag_snapshot.v1` 的不可变测试知识；
- `datasets/`：人工审查的 query、expected citation evidence、阈值与精确 snapshot / profile binding；
- `reports/`：由 Go CLI 生成的 metadata-only 确定性质量报告。

这些文件是离线 starter / 回归资产，不是应用运行时 dataset current projection、promotion candidate 或应用配置 binding。应用作用域 durable dataset 必须通过 Platform API 创建并由服务端重新读取 exact snapshot / profile；committed report 通过不自动创建 dataset resource、candidate review、promotion 或 baseline。

生成或复查 starter report：

```bash
cd services/platform
go run ./cmd/radishmind-workflow-rag-eval \
  --snapshot ../../datasets/eval/workflow-rag/snapshots/dev_core_v1.json \
  --dataset ../../datasets/eval/workflow-rag/datasets/dev_core_v1.json \
  --output ../../datasets/eval/workflow-rag/reports/dev_core_v1.review.json

go run ./cmd/radishmind-workflow-rag-eval \
  --snapshot ../../datasets/eval/workflow-rag/snapshots/dev_core_v1.json \
  --dataset ../../datasets/eval/workflow-rag/datasets/dev_core_v1.json \
  --output ../../datasets/eval/workflow-rag/reports/dev_core_v1.review.json \
  --check
```

只允许提交人工编写的 synthetic-public query 和 fragment。禁止复制真实用户 query、`workspace_internal` 知识、credential、prompt、answer 或模型原始响应。报告不得出现 query、fragment 正文、excerpt 或 review note。

CLI 对每个样本恰好调用一次现有确定性 lexical ranker，不调用 Gateway、不创建 workflow run，也不写应用 repository。`--check` 只比较 committed report 与重新计算结果，不修改文件；需要更新报告时去掉 `--check`，复核 diff 后再提交。

durable dataset、candidate review、权限、版本 CAS、双数据库和 Web 操作说明见 [Workflow RAG 开发测试态使用与资源治理指南](../../../docs/features/workflow/workflow-rag-dev-test-usage-governance-guide.md)。

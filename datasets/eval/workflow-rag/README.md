# Workflow RAG 离线评测资产

本目录承载 `Workflow RAG 评测数据集与知识质量审查 v1` 的 synthetic-public、可提交、可复算资产：

- `snapshots/`：严格符合 `workflow_rag_snapshot.v1` 的不可变测试知识；
- `datasets/`：人工审查的 query、expected citation evidence、阈值与精确 snapshot / profile binding；
- `reports/`：由 Go CLI 生成的 metadata-only 确定性质量报告。

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

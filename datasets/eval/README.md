# RadishMind 最小评测样本说明

更新时间：2026-04-01

当前目录用于存放第一阶段的最小离线评测样本。

第一阶段先做两件事：

1. 冻结样本结构
2. 让样本能与任务卡、契约文件一一对应

当前样本至少应描述：

- 输入请求是什么
- 召回输入应满足哪些边界与元数据约束
- 期望输出应满足哪些结构约束
- 一份可作为对照基线的 `golden_response`
- 风险等级应如何判定
- 哪些字段必须出现
- 哪些越界字段或行为不得出现

当前先使用以下 schema 约束样本格式：

- `radishflow-task-sample.schema.json`
- `radish-task-sample.schema.json`

当前 `Radish` 文档问答样本已新增 `retrieval_expectations`，用于把任务卡中的召回输入边界落成可执行检查。

当前已补最小回归脚本：

- `scripts/run-radishflow-diagnostics-regression.ps1`
- `scripts/run-radishflow-diagnostics-regression.sh`
- `scripts/check-radishflow-diagnostics-eval.ps1`
- `scripts/check-radishflow-diagnostics-eval.sh`
- `scripts/run-radish-docs-qa-regression.ps1`
- `scripts/run-radish-docs-qa-regression.sh`
- `scripts/check-radish-docs-qa-eval.ps1`
- `scripts/check-radish-docs-qa-eval.sh`

关系说明：

- `run-radishflow-diagnostics-regression.*` 负责执行 `RadishFlow explain_diagnostics` 样本回归
- `check-radishflow-diagnostics-eval.*` 负责把该回归接入仓库基线
- `run-radish-docs-qa-regression.*` 是真正执行样本回归的 runner
- `check-radish-docs-qa-eval.*` 是仓库基线入口，对 runner 做包装
- `check-repo.*` 继续通过上述入口脚本把各任务回归纳入仓库级校验链路

后续再补更完整的离线回归脚本和真实候选输出对照输入。

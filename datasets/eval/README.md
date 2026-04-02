# RadishMind 最小评测样本说明

更新时间：2026-04-02

当前目录用于存放第一阶段的最小离线评测样本。

第一阶段先做两件事：

1. 冻结样本结构
2. 让样本能与任务卡、契约文件一一对应

当前样本至少应描述：

- 输入请求是什么
- 召回输入应满足哪些边界与元数据约束
- 期望输出应满足哪些结构约束
- 一份可作为对照基线的 `golden_response`
- 可选的 `candidate_response`，用于接入真实候选输出或模拟模型输出
- 风险等级应如何判定
- 哪些字段必须出现
- 哪些越界字段或行为不得出现

当前先使用以下 schema 约束样本格式：

- `radishflow-task-sample.schema.json`
- `radish-task-sample.schema.json`

当前 `Radish` 文档问答样本已新增 `retrieval_expectations`，用于把任务卡中的召回输入边界落成可执行检查。

当前已补最小回归脚本：

- `scripts/run-radishflow-control-plane-regression.ps1`
- `scripts/run-radishflow-control-plane-regression.sh`
- `scripts/check-radishflow-control-plane-eval.ps1`
- `scripts/check-radishflow-control-plane-eval.sh`
- `scripts/run-radishflow-diagnostics-regression.ps1`
- `scripts/run-radishflow-diagnostics-regression.sh`
- `scripts/check-radishflow-diagnostics-eval.ps1`
- `scripts/check-radishflow-diagnostics-eval.sh`
- `scripts/run-radishflow-suggest-edits-regression.ps1`
- `scripts/run-radishflow-suggest-edits-regression.sh`
- `scripts/check-radishflow-suggest-edits-eval.ps1`
- `scripts/check-radishflow-suggest-edits-eval.sh`
- `scripts/run-radish-docs-qa-regression.ps1`
- `scripts/run-radish-docs-qa-regression.sh`
- `scripts/check-radish-docs-qa-eval.ps1`
- `scripts/check-radish-docs-qa-eval.sh`

关系说明：

- `run-radishflow-control-plane-regression.*` 负责执行 `RadishFlow explain_control_plane_state` 样本回归
- `check-radishflow-control-plane-eval.*` 负责把控制面状态说明回归接入仓库基线
- `run-radishflow-diagnostics-regression.*` 负责执行 `RadishFlow explain_diagnostics` 样本回归
- `check-radishflow-diagnostics-eval.*` 负责把该回归接入仓库基线
- `run-radishflow-suggest-edits-regression.*` 负责执行 `RadishFlow suggest_flowsheet_edits` 样本回归
- `check-radishflow-suggest-edits-eval.*` 负责把候选编辑回归接入仓库基线
- `run-radish-docs-qa-regression.*` 是真正执行样本回归的 runner
- `check-radish-docs-qa-eval.*` 是仓库基线入口，对 runner 做包装
- `check-repo.*` 继续通过上述入口脚本把各任务回归纳入仓库级校验链路
- 当前 `ps1` / `sh` runner 都通过 `scripts/run-eval-regression.py` 共享同一份 Python 回归核心
- 因此执行这些回归脚本时，当前环境需要具备可用的 Python 启动器与 `jsonschema`

`RadishFlow` 的回归 runner 当前已覆盖 `explain_control_plane_state`、`explain_diagnostics` 与 `suggest_flowsheet_edits` 三个任务，并支持样本内可选 `candidate_response` 校验，用于为后续真实模型输出接入预留稳定输入口。

其中 `explain_control_plane_state` 当前已覆盖：

- entitlement 过期阻塞
- package sync 轻度异常
- 控制面冲突态
- 上游 403 的授权边界对抗样本

`Radish` 的 docs QA runner 当前也已支持样本内可选 `candidate_response` 校验，方便在保持召回输入约束不变的前提下接入真实候选回答。

当前 `Radish answer_docs_question` 已覆盖的最小样本类型包括：

- 直接回答
- 证据不足降级
- 低风险导航回答
- 角色示例不等于最终授权的授权边界样本
- `docs + attachments` 混合召回
- `docs + forum` 混合召回，且要求正式文档优先于社区经验
- `docs + faq` 混合召回，且 FAQ 只作为正式文档的补充说明
- `wiki + faq` 混合召回，且恢复/操作流程仍以 wiki 主结论为准
- `docs + faq + forum` 三路混合召回，且需显式拒绝用社区经验覆盖正式发布规则

后续再补更完整的离线回归脚本和真实候选输出对照输入。

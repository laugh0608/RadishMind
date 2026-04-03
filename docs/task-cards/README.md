# 首批任务卡

更新时间：2026-03-31

本目录用于把路线图中的高优先级任务，从“任务名”收口到“可实现、可评测、可对齐协议”的正式任务卡。

当前已冻结的任务：

## `RadishFlow`

1. [explain_diagnostics](radishflow-explain-diagnostics.md)
2. [suggest_flowsheet_edits](radishflow-suggest-flowsheet-edits.md)
3. [explain_control_plane_state](radishflow-explain-control-plane-state.md)

## `Radish`

1. [answer_docs_question](radish-answer-docs-question.md)

使用原则：

- 任务卡定义的是任务边界、最小输入、输出要求和评测口径，不等同于最终实现代码
- 任务卡与 [跨项目集成契约](../radishmind-integration-contracts.md) 和 [真实契约文件](../../contracts/README.md) 保持一致
- 若未来实现发现字段命名或结构需要调整，应先同步更新任务卡和契约，再改实现
- 当前阶段优先保证“状态优先、结构化输出、显式风险分级、人工确认”

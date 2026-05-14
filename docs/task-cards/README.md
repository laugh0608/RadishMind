# 首批任务卡

更新时间：2026-05-10

本目录用于把路线图中的高优先级任务，从“任务名”收口到“可实现、可评测、可对齐协议”的正式任务卡或前置条件清单。

当前说明：

- 当前仓库主线已经切回平台本体建设；因此这里的文档更多承担“应用面任务边界”和“真实接入前置条件”职责。
- 如果你现在要判断应该优先做什么，先读 `docs/radishmind-current-focus.md`、`docs/radishmind-capability-matrix.md` 和 `docs/radishmind-roadmap.md`，不要把这些任务卡误解成当前唯一主线。

当前已冻结的应用面任务：

## `RadishFlow`

1. [explain_diagnostics](radishflow-explain-diagnostics.md)
2. [suggest_flowsheet_edits](radishflow-suggest-flowsheet-edits.md)
3. [explain_control_plane_state](radishflow-explain-control-plane-state.md)
4. [suggest_ghost_completion](radishflow-suggest-ghost-completion.md)
5. [接入前置条件与阻塞项](radishflow-first-real-integration.md)

## `Radish`

1. [answer_docs_question](radish-answer-docs-question.md)

## 跨项目

1. [接入前置条件总表](cross-project-integration-readiness.md)
2. [`P2 Session & Tooling` 实现前置条件](session-tooling-implementation-preconditions.md)

使用原则：

- 任务卡定义的是任务边界、最小输入、输出要求和评测口径，不等同于最终实现代码
- 前置条件型任务卡定义的是当前不能继续前推的阻塞项、已有门禁和后续触发条件，不等同于已经完成上层接线
- 当前平台主线仍以 runtime、session、tooling、governance 和 model adaptation 为主；任务卡不替代这些平台级规划文档
- 任务卡与 [跨项目集成契约](../radishmind-integration-contracts.md) 和 [真实契约文件](../../contracts/README.md) 保持一致
- 若未来实现发现字段命名或结构需要调整，应先同步更新任务卡和契约，再改实现
- 当前阶段优先保证“状态优先、结构化输出、显式风险分级，以及对会写回真相源的动作保留人工确认”
- 编辑器内 ghost 补全和正式候选 patch 必须分开建模，不应混用为同一任务

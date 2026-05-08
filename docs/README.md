# RadishMind 文档入口

更新时间：2026-05-08

## 阅读原则

`docs/` 是 RadishMind 的正式文档源。关键入口文档只保留项目定位、当前阶段、最近进度、下一步和阻塞项；历史推演、批次细节、长实验观察和一次性讨论应沉淀到 `docs/devlogs/`、任务卡、manifest 或 run record 中。

新会话优先按以下顺序读取：

1. 本文件
2. [当前推进焦点](radishmind-current-focus.md)
3. [阶段路线图](radishmind-roadmap.md)
4. 与当次任务直接相关的产品、架构、代码规范、契约、任务卡、评测或周志

## 当前状态

- 近期主线集中在 `M3/M4` 收口：维持现有服务/API smoke 作为未来上层接入门禁，并继续推进 `RadishMind-Core` 结构化输出决策实验。
- `RadishFlow` 的 `suggest_flowsheet_edits` 与 `suggest_ghost_completion` 已具备 committed eval、真实 candidate record、audit、replay 与服务 smoke 基线；当前不再默认扩同类真实 capture。
- `Radish` 已优先落在 docs QA、文档检索增强和结构化问答评测上；真实上层接入仍等待。
- `RadishCatalyst` 仍只做文档级预留，不扩真实 schema、adapter 或 gateway smoke。
- 当前正式仓库边界仍是“四仓独立 + 协议接入”；不在 `RadishMind` 与 `RadishFlow` / `Radish` / `RadishCatalyst` 之间默认采用双向 `git submodule`。
- `RadishMind-Core` 当前重点不是训练放量，而是继续验证 raw、repair、hard-field injection、task-scoped builder、自然语言 audit 和人工 review 的分工边界，并把 broader task-scoped builder review 的通过结论收口为稳定路线信号。
- `RadishMind-Core` broader task-scoped builder review 已完成两段本地执行、审计与 15 条样本人工复核；full-holdout-9 和 holdout6-v2-non-overlap 的 machine gate / offline eval / natural-language audit 均通过，当前记录集已更新为 15/15 `reviewed_pass`。
- 当前首要动作不再是重跑 broader review，而是继续维护 service/API smoke 矩阵，并先把 constrained/guided decoding 固定为下一轮 `M4` 主线；待该轨结果明确后，再决定是否进入更大样本面或 `3B/4B` 对照。
- 图片生成能力由 `RadishMind-Image Adapter` 与独立 backend 承接；主模型只输出结构化意图、约束、审查和 artifact metadata。

## 文档约束

- 入口文档必须简短，优先描述“现在是什么、下一步做什么、什么不能做”。
- 不在入口文档重复长批次流水、历史失败细节或完整命令输出；这些内容放入周志、实验记录、summary 或 task card。
- 回答“今天要做什么”这类问题时，默认读 [当前推进焦点](radishmind-current-focus.md)，不要默认展开长契约或长评测文档。
- 新增或更新文档时，优先更新既有正式文档；只有当内容有长期复用价值且无法自然归入现有文档时，才新增文档。
- 文档中提到外部项目时，默认使用项目名和在线仓库 URL，不写个人机器路径。

## 语言与实现约束

- 代码应趋近对应语言的惯用、清晰和可维护实践；本仓库当前主实现栈为 `Python`。
- 命名必须表达真实职责和领域含义，避免 `process_data`、`handle_item`、`manager`、`helper` 这类无法说明边界的泛名。
- 抽象只在能稳定表达职责边界、消除真实重复或收敛复杂度时引入；禁止为了“看起来通用”增加晦涩封装、空转方法或多层转发。
- 能用语言标准库、结构化 schema、明确数据类型和小而直接的函数解决的问题，不应写成难追踪的动态包装或隐式 fallback 链。
- 修复问题优先定位根因；不要用连续叠加兜底逻辑掩盖协议、数据或职责边界错误。

## 关键文档

- [当前推进焦点](radishmind-current-focus.md)
- [产品范围与目标](radishmind-product-scope.md)
- [系统架构](radishmind-architecture.md)
- [阶段路线图](radishmind-roadmap.md)
- [代码规范](radishmind-code-standards.md)
- [跨项目集成契约](radishmind-integration-contracts.md)
- [ADR 0002: 仓库集成边界](adr/0002-repository-integration-boundary.md)
- [RadishMind-Core 首版基座评估](radishmind-core-baseline-evaluation.md)
- [ADR 0001: 分支与 PR 治理](adr/0001-branch-and-pr-governance.md)
- [开发日志说明](devlogs/README.md)
- [任务卡入口](task-cards/README.md)
- [统一契约文件说明](../contracts/README.md)
- [数据集与评测目录说明](../datasets/README.md)
- [训练目录说明](../training/README.md)
- [脚本目录说明](../scripts/README.md)

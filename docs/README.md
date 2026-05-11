# RadishMind 文档入口

更新时间：2026-05-10

## 阅读原则

`docs/` 是 RadishMind 的正式文档源。关键入口文档只保留项目定位、当前阶段、最近进度、下一步和停止线；历史推演、批次细节、长实验观察和一次性讨论应沉淀到 `docs/devlogs/`、任务卡、manifest 或 run record 中。

新会话优先按以下顺序读取：

1. 本文件
2. [项目总览与使用指南](radishmind-project-guide.md)
3. [当前推进焦点](radishmind-current-focus.md)
4. [产品范围与目标](radishmind-product-scope.md)
5. [战略定义](radishmind-strategy.md)
6. [能力矩阵](radishmind-capability-matrix.md)
7. [阶段路线图](radishmind-roadmap.md)
8. 与当次任务直接相关的架构、契约、任务卡、评测或周志

## 当前状态

- `RadishMind` 已正式从“模型实验 / 接入准备仓库”的狭义口径，收口为“协议驱动、可审计、可本地部署、可工具化的 Copilot / Agent runtime platform”。
- 当前仓库主线不再只是等待其他项目真实接入；接下来可以独立推进五条平台主线：`Runtime Service`、`Conversation & Session`、`Tooling Framework`、`Evaluation & Governance`、`Model Adaptation`。
- 当前项目的更强正式定义已经固定在 [战略定义](radishmind-strategy.md)：`RadishMind` 是 `AI Middleware / AI Runtime`，核心价值是把多模型、多协议、多任务收口成可控产品能力。
- 平台后续必须同时具备两类兼容能力：南向接入自研模型与外部模型，北向对外提供常见 AI 协议接口；当前仓库已落地最小 `provider registry` 骨架，并由同一 southbound 入口收口 `openai-compatible / gemini-native / anthropic-messages` 调用基础，同时已落地 `/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models` 的第一版 bridge-backed 兼容面，但还没有正式的 `HuggingFace / Ollama` 服务接入。
- 平台表层实现分工已固定为：`UI=React + Vite + TypeScript`、`Platform Service Layer=Go`、`Model Side=Python`，并且所有层都只消费 `contracts/` 里的 canonical protocol。
- `services/platform/` 下的最小 `Go` 平台服务层 bootstrap 已落地，当前已固定 `HTTP` 服务启动、`/healthz`、`/v1/models`、`/v1/chat/completions`、`/v1/responses` 和 `/v1/messages` 的第一版 bridge；真正还未完成的是流式转发、更完整的 provider 选择和动态 model/profile inventory。
- 既有 `M3` service/API smoke matrix 与 `M4` broader review、`3B/4B` capacity review 继续保留为冻结证据和门禁；它们不再是当前唯一主线，也不再默认继续深挖同一批样本。
- `RadishFlow` 仍是第一优先应用面，但当前只冻结 gateway、UI consumption 和 candidate handoff 门禁；上层尚未具备真实接入能力前，不继续细化假想接线。
- `Radish` 当前保留 docs QA、文档检索增强和结构化问答资产；真实上层接入仍等待。
- `RadishCatalyst` 仍只做文档级预留，不扩真实 schema、adapter、gateway smoke 或模型接线。
- 图片生成能力继续由 `RadishMind-Image Adapter` 与独立 backend 承接；主模型只负责结构化 intent、约束、审查和 artifact metadata。

## 文档约束

- 入口文档必须简短，优先描述“项目是什么、现在能推进什么、哪些边界不能越”。
- 不在入口文档重复长批次流水、历史失败细节或完整命令输出；这些内容放入周志、实验记录、summary 或 task card。
- 回答“今天要做什么”这类问题时，默认读 [当前推进焦点](radishmind-current-focus.md) 与 [能力矩阵](radishmind-capability-matrix.md)，不要默认展开长契约或长评测文档。
- 新增或更新文档时，优先更新既有正式文档；只有当内容有长期复用价值且无法自然归入现有文档时，才新增文档。
- 文档中提到外部项目时，默认使用项目名和在线仓库 URL，不写个人机器路径。

## 语言与实现约束

- 代码应趋近对应语言的惯用、清晰和可维护实践；本仓库按职责分层：模型训练、评测和脚本优先 `Python`，前端 UI 默认 `React + Vite + TypeScript`，服务 / `gateway` / `API` 可按职责采用 `Go`。
- 命名必须表达真实职责和领域含义，避免 `process_data`、`handle_item`、`manager`、`helper` 这类无法说明边界的泛名。
- 抽象只在能稳定表达职责边界、消除真实重复或收敛复杂度时引入；禁止为了“看起来通用”增加晦涩封装、空转方法或多层转发。
- 能用语言标准库、结构化 schema、明确数据类型和小而直接的函数解决的问题，不应写成难追踪的动态包装或隐式 fallback 链。
- 修复问题优先定位根因；不要用连续叠加兜底逻辑掩盖协议、数据或职责边界错误。

## 关键文档

- [当前推进焦点](radishmind-current-focus.md)
- [项目总览与使用指南](radishmind-project-guide.md)
- [产品范围与目标](radishmind-product-scope.md)
- [战略定义](radishmind-strategy.md)
- [能力矩阵](radishmind-capability-matrix.md)
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

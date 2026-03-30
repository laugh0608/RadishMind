# RadishMind

RadishMind 是 `Radish` 体系下独立演进的 AI / Copilot 项目，目标是围绕 `RadishFlow` 与 `Radish` 提供统一的结构化智能层、评测与模型实验能力。

当前阶段不把它定义为“替代业务内核的大模型项目”，而是定义为：

- 一个独立维护协议、数据、评测、工具编排与模型实验的仓库
- 一个面向 `RadishFlow` 与 `Radish` 的外部智能层
- 一个默认输出解释、证据、结构化建议和候选动作的工程系统

## 当前定位

基于已审查的两个参考仓库，`RadishMind` 当前优先关注以下能力：

- 面向 `RadishFlow`
  - 基于 `FlowsheetDocument + selection + diagnostics + solve state` 的解释与建议
  - 控制面 / entitlement / lease / package sync 状态解释
  - 候选编辑提案生成，并保持 `requires_confirmation`
  - 画布截图理解作为补充输入，而不是第一阶段唯一主线

- 面向 `Radish`
  - 固定文档与在线文档问答
  - Console / 权限 / 运营能力说明与流程引导
  - 论坛 / 文档内容的结构化摘要、标签和分类建议
  - 附件 / 截图理解与引用解释

- 共享基础
  - 统一输入输出协议
  - 检索、工具调用、规则校验与评测链路
  - Teacher / Student 模型实验和版本对比

当前不把以下内容作为第一阶段目标：

- 直接替代 `RadishFlow` 的数值求解内核、控制面或 CAPE-OPEN 适配层
- 直接接管 `Radish` 的 Auth / Gateway / API / Console 业务真相源
- 让模型在无校验前提下直接写入上层项目状态
- 在第一阶段先锁定最终技术栈或主模型路线

## 文档入口

- [规划总览](docs/README.md)
- [产品范围与目标](docs/radishmind-product-scope.md)
- [系统架构草案](docs/radishmind-architecture.md)
- [阶段路线图](docs/radishmind-roadmap.md)
- [跨项目集成契约草案](docs/radishmind-integration-contracts.md)

## 当前分支策略

- 常态开发分支：`dev`
- `master` 仅作为稳定主线
- 日常规划、设计与实现优先在 `dev` 上推进

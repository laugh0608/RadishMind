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
- 在第一阶段先锁定最终部署形态

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

## 当前模型路线

当前模型路线已先收口为一条主线加两条对照线：

- `minimind-v`：默认 `student/base` 主线，承担当前领域适配、蒸馏承接与后续训练实验
- `Qwen2.5-VL`：默认 `teacher` / 多模态强基线，承担首轮高质量对照评测、标注参考和复杂图文任务 PoC
- `SmolVLM`：默认轻量本地对照组，用于低资源环境、部署下限和小模型回归比较

当前不把 `Molmo` 或其他视觉语言模型写成默认主线；如需引入，先作为补充对照或专项研究路线，再由离线评测结果决定是否升级优先级。

## 开发与验证环境

当前仓库同时保留 Windows / PowerShell 与 Linux / WSL 两套验证入口，并要求长期保持语义一致：

- Windows / PowerShell：优先使用 `pwsh ./scripts/check-repo.ps1`
- Linux / WSL：优先使用 `./scripts/check-repo.sh`

需要注意：

- 当前仓库主实现栈已收口为 `Python`
- 评测、回归、仓库级校验与后续模型/数据工具链默认以 `Python` 为核心实现
- Windows 侧 `.ps1` 与 Linux / WSL 侧 `.sh` 继续保留，但它们当前只作为平台入口包装，不再承载核心校验逻辑
- 当前验证链路需要可用的 Python 环境；至少应提供 `python3` 或等价 Python 启动器，以及 `jsonschema`

## 当前评测进展

当前最小离线回归已覆盖：

- `RadishFlow explain_diagnostics`
- `RadishFlow suggest_flowsheet_edits`
- `RadishFlow explain_control_plane_state`
- `Radish answer_docs_question`

其中：

- `RadishFlow` 三个首批任务都已具备最小回归闭环
- `Radish answer_docs_question` 已具备召回输入约束、`golden_response` 对照和离线 `candidate_response` 输入口

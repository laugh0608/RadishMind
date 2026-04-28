# RadishMind-Core 首版基座评估矩阵

更新时间：2026-04-28

## 文档目的

本文档用于固定 `M4` 前置阶段的模型基座评估口径。

当前目标不是下载模型、启动训练或承诺最终生产模型，而是先明确 `RadishMind-Core`、`minimind-v`、`Qwen2.5-VL` 与 `SmolVLM` 在首轮评估中的角色、准入门槛、退出条件和非目标。

## 当前判断

`RadishMind-Core` 是基座适配型自研主模型，不是从零预训练基础大模型。首版应优先比较 `3B` 与 `4B` 的协议遵循、中文任务理解、结构化输出、citation 对齐、风险确认和本地部署成本；只有当 `3B/4B` 在关键任务上明显不足，且部署资源与量化策略可接受时，才评估 `7B`。

图片输入理解可以进入 `RadishMind-Core` 或视觉适配路线；图片像素生成不进入主模型参数目标，继续由 `RadishMind-Image Adapter` 和独立生图 backend 承接。

## 评估对象

| 对象 | 角色 | 首轮定位 | 进入条件 | 退出或暂缓条件 |
| --- | --- | --- | --- | --- |
| `minimind-v` | 默认 `student/base` 主线 | 承接领域适配、训练样本格式和协议遵循验证 | 能稳定输出 `CopilotResponse` 结构，支持中文任务说明和风险确认 | 结构化协议遵循长期不达标，或需要过多任务级兜底 |
| `radishmind-core-3b` | 首选本地友好档 | 首轮默认自研主模型尺寸 | 在 32GB 级开发 / 部署机上优先满足核心文本任务 | 协议遵循、citation 或多步骤推理明显低于门槛 |
| `radishmind-core-4b` | 首版增强候选 | `3B` 不足时的首个对照档 | 相比 `3B` 明显改善关键任务且成本可接受 | 质量收益不足以抵消部署成本 |
| `radishmind-core-7b` | 长期本地部署上限 | 增强档，不作为首轮默认 | `3B/4B` 均无法满足关键任务，且量化、延迟和内存预算可接受 | 资源占用超出本地部署边界，或评测收益不明确 |
| `Qwen2.5-VL` | teacher / 强基线 | 质量上界、标注参考和蒸馏参考 | 用于复杂图文任务、对照评测和数据审核 | 不作为 `RadishMind-Core` 本地部署目标 |
| `SmolVLM` | 轻量对照组 | 低资源下限和小模型回归基线 | 用于验证轻量场景可接受下限 | 不替代 `minimind-v` 或 `RadishMind-Core` 主线 |

## 评估维度

首轮离线评估至少覆盖：

- `protocol_following`：是否稳定遵循 `CopilotRequest -> CopilotResponse`
- `structured_response_validity`：是否能产出 schema-valid JSON
- `chinese_task_understanding`：是否能理解中文任务说明和中文上下文
- `citation_alignment`：是否能保持 citation id、顺序和证据指向
- `risk_confirmation`：是否能保留 `risk_level` 与 `requires_confirmation`
- `action_boundary`：是否避免把 advisory proposal 误写成可执行动作
- `training_sample_fit`：是否适合接收 `CopilotTrainingSample`
- `image_intent_planning`：是否只输出结构化图片生成意图，不直接生成图片像素
- `local_deployment_cost`：是否符合 32GB 级开发 / 小型部署设备预算

## 进入下一步的门槛

`3B` 或 `4B` 进入首轮微调 / 蒸馏实验前，必须满足：

- 能通过 `CopilotTrainingSample` 最小样本格式检查
- 在核心文本任务上达到可比较的 schema-valid 输出
- 中高风险候选动作必须保留 `requires_confirmation=true`
- 不把图片像素生成纳入主模型目标
- 能用现有 eval / gateway smoke 作为回归入口复查协议行为

`7B` 只有在以下条件同时成立时进入评估：

- `3B/4B` 在关键任务上无法达标
- 失败点不是简单数据补齐、prompt 修正或规则校验能解决的问题
- 32GB 级环境下的量化、延迟和内存预算仍可接受
- 有明确评测差距证明扩尺寸有收益

## 仓库级门禁

当前矩阵以 `scripts/checks/fixtures/radishmind-core-baseline-matrix.json` 作为机器可读真相源，并由 `scripts/check-radishmind-core-baseline-matrix.py` 接入 `check-repo`。

该 smoke 只检查评估口径和边界是否稳定，不下载模型、不启动训练、不访问外部 provider。

## 暂不做

- 不下载 `minimind-v`、`Qwen2.5-VL`、`SmolVLM` 或任何权重
- 不启动微调、蒸馏、量化或长驻推理服务
- 不把 `14B/32B` 写成默认自研主模型目标
- 不把图片像素生成并入 `RadishMind-Core`
- 不为尚未接入的上层 UI / 命令层继续新增模拟 summary

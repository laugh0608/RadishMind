# RadishMind 跨项目集成契约

更新时间：2026-05-13

## 文档目的

本文档是 `RadishMind` 与上层项目之间的集成契约入口，只保留当前结论、协议索引和阅读路径。详细字段、样例和长任务口径拆入 `docs/contracts/` 专题页，避免默认读取时消耗过多阅读和 token 预算。

当前目标不是一次性定死全部字段，而是先建立足够稳定的抽象，避免 `RadishFlow`、`Radish` 和后续 `RadishCatalyst` 各自演化出不兼容的接入方式。

当前文档口径已经同步落成仓库内真实契约文件：

- `contracts/copilot-request.schema.json`
- `contracts/copilot-response.schema.json`
- `contracts/copilot-gateway-envelope.schema.json`
- `contracts/copilot-training-sample.schema.json`
- `contracts/session-record.schema.json`
- `contracts/tool.schema.json`
- `contracts/tool-registry.schema.json`
- `contracts/tool-audit-record.schema.json`
- `contracts/session-recovery-checkpoint.schema.json`
- `contracts/session-recovery-checkpoint-manifest.schema.json`
- `contracts/image-generation-intent.schema.json`
- `contracts/image-generation-backend-request.schema.json`
- `contracts/image-generation-artifact.schema.json`

## 当前协议原则

- 统一骨架 + 项目专属上下文块。
- 结构化 JSON 优先，图像和附件作为 artifact 补充。
- 默认 advisory mode，不做直接写回。
- 所有高风险输出都必须带 `requires_confirmation`。
- 兼容层只做翻译，不另起第二套真相源。
- 上层项目只消费建议、解释、候选动作和审计信息，最终业务真相源仍由上层维护。

## 专题索引

- [服务/API 接入契约](contracts/service-api.md)：northbound / southbound 兼容边界、`CopilotGatewayEnvelope`、`RadishFlow` UI consumption、candidate edit handoff、上层接入等待口径和仓库集成边界。
- [会话记录契约](contracts/session.md)：`Conversation & Session` 的 `session_id / turn_id`、history policy、recovery record、northbound session metadata 和 advisory-only audit 边界。
- [工具框架契约](contracts/tooling.md)：`Tooling Framework` 的 tool definition、registry、policy/audit record 和不执行真实工具的 v1 停止线。
- [训练 / 蒸馏样本契约](contracts/training-samples.md)：`CopilotTrainingSample`、训练集合治理、candidate record 转换、offline eval runner、本地模型 candidate wrapper 和 M4 builder/tooling 证据边界。
- [图片生成契约](contracts/image-generation.md)：`RadishMind-Image Adapter`、image intent、backend request、artifact metadata 和最小评测 manifest。
- [输入与项目上下文契约](contracts/input-context.md)：`CopilotRequest`、artifact 抽象和项目上下文专题索引。
- [RadishFlow 上下文契约](contracts/radishflow-context.md)：`RadishFlow` export snapshot、ghost completion、上游实现清单和任务级上下文要求。
- [Radish 上下文契约](contracts/radish-context.md)：`Radish` docs QA 的知识上下文、检索来源和 artifact metadata 约束。
- [RadishCatalyst 上下文预留契约](contracts/radishcatalyst-context.md)：第三项目的游戏知识、运行状态摘要和 spoiler policy 预留。
- [输出与候选动作契约](contracts/output-actions.md)：`CopilotResponse`、candidate action、任务枚举、脱敏要求、关键边界和推荐原则。

## 默认阅读路径

- 判断当前平台或服务接入边界时，读本文件后进入 [服务/API 接入契约](contracts/service-api.md)。
- 修改 schema、字段或任务上下文时，读对应专题页，并同步检查 `contracts/` 下的 schema 真相源。
- 查历史实验、长样本列表和批次流水时，优先读周志、task card、manifest、summary 或 run record，不把这些内容追加回本入口。

## 当前停止线

- 不为 `RadishFlow`、`Radish` 或 `RadishCatalyst` 新增第二套项目私有协议。
- 不把兼容层字段当作业务真相源。
- 不让模型输出直接写回上层项目。
- 不在 `RadishCatalyst` 进入真实任务、adapter skeleton 和最小 eval sample 前扩 schema 枚举或 gateway route。

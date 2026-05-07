# RadishMind 系统架构

更新时间：2026-05-06

## 架构目标

`RadishMind` 的目标架构固定为六层：

1. Client Adapters & Context Packers
2. Copilot Gateway / Task Router
3. Retrieval & Tool Layer
4. Model Runtime Layer
5. Rule Validation & Response Builder
6. Data / Evaluation / Training Pipeline

这个拆分用于隔离项目语义、工具编排、模型推理、安全校验和评测闭环，让 `RadishFlow`、`Radish` 与后续 `RadishCatalyst` 能通过统一协议接入，而不是各自私接模型。

## 请求流程

```text
上层项目状态 / 文档 / 附件 / 图像
        ↓
Adapter 打包为 CopilotRequest
        ↓
Gateway 识别 project / task / schema_version
        ↓
Retrieval / Tools 获取证据并压缩上下文
        ↓
Model Runtime 生成解释、问题和候选意图
        ↓
Rule Validation / Response Builder 收口结构和风险
        ↓
CopilotResponse / GatewayEnvelope 返回给上层
```

## 分层职责

### 1. Client Adapters & Context Packers

- 将上层状态转换为统一 `CopilotRequest`。
- 裁剪、脱敏或摘要敏感字段。
- 将 `CopilotResponse` 映射回 UI、日志或候选提案。
- 当前 `RadishFlow` 优先维护 `export -> adapter -> request` 链路；`Radish` 优先维护文档和内容上下文；`RadishCatalyst` 暂不落真实 adapter。

### 2. Copilot Gateway / Task Router

- 统一校验请求、识别任务、选择 provider/profile，并返回 `CopilotGatewayEnvelope`。
- 当前对外形态优先是进程内 Python API；HTTP JSON 只作为后续包装形态。
- 服务/API smoke 继续锁定 advisory-only、schema validation、route metadata、error envelope 和 handoff 不执行这些不变量。

### 3. Retrieval & Tool Layer

- 承载文档检索、附件/Markdown/JSON 解析、项目语义转换和本地合法候选生成。
- 能规则化的逻辑优先留在工具层，例如 ghost completion 的合法候选空间和 recent-actions suppress 信号。
- 工具层只生成证据或候选动作，不直接写业务真相源。

### 4. Model Runtime Layer

- Teacher models 用于强基线、标注参考、蒸馏和复杂任务对照。
- Student models 用于本地化、小成本部署和回归实验。
- `RadishMind-Core` 负责理解、推理、结构化建议、候选排序、风险标记和可选图片输入理解。
- Image Generation Runtime 独立负责图片像素生成；主模型只输出结构化 image intent 和约束。

### 5. Rule Validation & Response Builder

- 校验响应结构、目标对象、风险等级、确认要求、citation 和 action shape。
- 对可规则化字段保持确定性，例如 `status / risk_level / requires_confirmation / proposed_actions / patch / issue code`。
- 当前 `task-scoped response builder` 是 `M4` 决策实验和 tooling 分工证据，不是 raw 模型晋级或 production contract。

### 6. Data / Evaluation / Training Pipeline

- 管理 eval sample、candidate record、audit、replay、offline eval、training sample manifest 和 review records。
- 真实模型输出、生成 JSONL 和大体积实验产物默认留在 `tmp/`。
- 训练、蒸馏和模型晋级必须同时看 raw 输出、后处理轨、离线评测、自然语言 audit、人工 review 和 holdout 泄漏边界。

## 当前进度

- `contracts/` 已具备 Copilot request / response / gateway envelope / training sample / image generation intent / backend request / artifact schema。
- `RadishFlow` 的 gateway demo、service smoke matrix、UI consumption 和 candidate edit handoff 已作为未来接入门禁保留。
- `suggest_flowsheet_edits` 与 `suggest_ghost_completion` 的真实 candidate record、audit、replay 和治理链已阶段性收口；新增真实 capture 需要先说明非重复 drift 假设。
- `RadishMind-Core` 本地小模型观测显示 raw 仍 blocked；当前优先验证 task-scoped builder / tooling 分工、citation tightened rerun 和 human review records。
- `RadishMind-Image Adapter` 已具备 intent、backend request、artifact metadata 与最小评测 manifest；暂不调用真实生图 backend。

## 工程约束

- 层之间通过 schema、明确数据类型和稳定函数边界连接，避免隐式全局状态或字符串拼装协议。
- 代码优先使用对应语言的标准库和惯用结构；本仓库 Python 代码应保持直接、可测试、易读。
- 方法名和模块名必须说明真实职责；不要用空泛 helper、manager、processor 掩盖边界不清。
- 不为简单调用链增加多层抽象；只有当职责稳定、复用真实存在或复杂度明显下降时才抽 helper、builder 或 adapter。
- 修复结构漂移时优先修正 schema、builder 或任务边界，不用无限 fallback 包裹模型输出。

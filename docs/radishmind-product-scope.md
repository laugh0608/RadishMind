# RadishMind 产品范围与目标

更新时间：2026-09-06

## 核心定义

RadishMind 是 Radish 体系下的 AI 工具、工作流、模型网关和 Copilot 集成平台。平台把模型接入、受控运行、结构化建议、人工审查和运行证据组织为可重复的用户流程，不成为上层业务真相源。

当前成熟度为**内部开发者预览**。能力状态由[能力矩阵](radishmind-capability-matrix.md)与对应功能专题维护，当前目标由[当前推进焦点](radishmind-current-focus.md)维护；本文件不保存批次流水。

## 目标用户与核心任务

首要用户是 Radish 体系内部开发者和团队成员。核心任务是：**独立创建可复用 AI 应用，完成受控运行、结果审查与回归验证。**

Workflow 的创建、编辑、校验、保存、恢复和审查是首要入口；Gateway 提供模型调用支撑，Admin 提供完成该流程所需的身份、成员、路由和治理能力。上层项目暂未具备真实挂载点，不应拖慢 RadishMind 已明确的内部任务，平台自身仍可按边界推进。

下一功能应说明目标用户、遇到的具体阻塞、现有路径与缺口，以及完成后的可观察改善。功能数量、画板数量、批次关闭和 checker 通过不能单独证明用户收益。

## 四个产品面

| 产品面 | 产品职责 | 当前范围与长期边界 |
| --- | --- | --- |
| `User Workspace` | 应用创建、配置、调用、结果和回归审查 | 正式用户端代码位于 `apps/radishmind-web/`；提供开发测试态 Workflow、Prompt、Agent / Copilot、RAG、API key、运行与结果工作区 |
| `Admin Control Plane` | 本地账户、工作区成员、权限和运行治理 | 管理 tenant、provider profile、路由、配额与价格策略、审计和 deployment status；生产身份、secret 与部署分别验收 |
| `Model Gateway / API Distribution` | 统一模型发现与多协议调用 | 北向兼容 OpenAI-compatible / Responses / Messages / Models，南向接入多个 Provider；开发测试态 quota、trace、用量与价格快照不构成 billing ledger |
| `Workflow / Agent Runtime` | 编排受控执行与可追溯结果 | 已实现的 Definition、RAG、Prompt、Agent / Copilot、HTTP Tool 按独立执行档案约束；通用自治循环和业务写回继续关闭 |

四个产品面共同服务用户任务。图片理解与图片生成是横切适配能力，不另立第五条一级产品主线。细节进入[功能设计文档入口](features/README.md)。

## 平台自有数据与外部真相

- RadishMind 拥有平台本地账户、凭证、Web Session、角色、权限和工作区成员关系，以及应用配置、Workflow draft / definition、run record、trace、reported usage 和运行审计。
- Radish 保持其自身身份、组织成员关系和业务数据真相。联合登录以 `(issuer, subject)` 显式绑定本地 `user_id`，不按 email 自动合并，不同步身份数据库，不让上游 claim 隐式授予本地权限。
- RadishMind 可以作为 Radish 的 OIDC client；当前不成为第二个 OIDC issuer。真实 issuer、client registration、secret、部署资源与负责人仍是联调条件。
- 上层输入只用于解释、诊断、结构化建议和候选动作；高风险动作保留 `requires_confirmation`，由规则或人工审查约束，不由模型直接写入业务真相。
- `memory_dev`、`sqlite_dev` 与 `postgres_dev_test` 的存在不授予 production 能力；作用域、原子并发、migration、重启恢复和失败不回退分别验证。

## 能力范围的区分

应用 Session / Turn、显式结果资产和运行审查已有开发测试态持久化；历史通用 Session / Checkpoint 的 metadata-only 契约并不因此具备长期记忆或跨轮恢复执行器。

独立 Workflow HTTP Tool 已有受控确认与执行链；通用 `/v1/tools/actions` 的 blocked shell 不因该专题完成而开放。相邻 owner 的授权、状态和审计不能互相代替。

Image Path 已完成开发测试态 fixture client、adapter、一次性 binary delivery、私有存储协调和成功引用延后释放。metadata-only response builder 独立消费产物元数据，不读取图片二进制。真实生图 backend、引用解析、生产对象存储、public URL、HTTP / Gateway / Web 交付继续未实现。

## 使用与效果验收

后续内部试用先在对应功能专题中明确用户任务、参与范围、测试输入、模型 / Provider、数据处理和运行窗口。首轮记录基线后再确定改进目标，不预填达标结论。

| 维度 | 记录方式 |
| --- | --- |
| 独立完成 | 成功完成次数 / 有效尝试次数；同时记录样本量、环境和失败原因 |
| 使用成本 | 完成耗时、求助次数、卡住位置、恢复动作与人工修改量 |
| 输出质量 | 区分结构合规、事实正确、建议有用和人工审查结论 |
| 运行质量 | 记录模型 / 配置版本、延迟、失败类型及可信上报用量；缺失值不推算为零 |
| 可重复性 | 固定任务集与版本，保留 Run / Campaign / Case / Suite 的精确引用和回归结果 |

人工浏览器验收、自动浏览器回归、真实 Provider 验证、真实团队使用和生产验收分别记录。mock / fixture、少量人工样本和部分模块覆盖率不代表全部场景达标。

## 模型与评测

`RadishMind-Core` 采用开源基座加自有协议、数据与评测偏好适配，不从零预训练基础大模型。训练与蒸馏后置，需有新能力假设、评测基线和明确运行窗口。

raw、guided、builder 与 repaired 结果分别报告；结构修复通过不等于模型语义正确或 raw 晋级。先利用既有应用评测能力积累有代表性的任务与质量基线，再决定模型适配投入。详细记录见[训练目录](../training/README.md)。

图片像素生成由独立 `RadishMind-Image Adapter` 和生图后端承担，主模型负责理解、规划、约束、审查和结构化意图。

## 外部接入顺序

1. `RadishFlow`：优先验证解释、诊断、候选编辑与 ghost completion；真实挂载点、owner、协议和验收环境明确后恢复接入。
2. `Radish`：保留文档问答与检索资产；OIDC 联调和业务 Copilot 接入分别确认范围。
3. `RadishCatalyst`：保留文档级扩展边界，暂不扩真实 schema、adapter 或运行接线。

外部协议见[跨项目集成契约](radishmind-integration-contracts.md)与[集成专题](integrations/README.md)。本地外部路径只作当次只读输入，不进入长期文档。

## 历史契约索引

以下标识对应早期设计证据，不能覆盖上文已实现的开发测试态能力，也不提供当前排期：

- `product-surface-v1-boundary`：四产品面的初始资源与职责划分。
- `control-plane-data-boundary`：平台自有数据和外部真相边界。
- `radish-oidc-client-preconditions`：issuer、claim mapping、tenant binding 等联合身份前置条件。
- `gateway-api-key-quota-readiness`：API key、quota、trace 与生产准入边界。
- `workflow-definition-run-record-boundary`：Workflow 状态流转与审计证据。
- `control-plane-read-formal-ui-boundary-v1`：早期正式 UI 边界；`control-plane-read-consumer-contract-v1`：只读消费者合同。

历史协议从[Control Plane 契约](contracts/control-plane-read-side.md)追溯；当前功能从[功能专题](features/README.md)读取。

## 非目标与实现原则

- 不做无边界自治 agent、通用 SaaS 聊天前端或基础大模型预训练；不由模型、客户端或人工批准直接授予 Action Safety 有效级别。
- 不开放无限制工具、业务写回、自动确认提交、自动发布或 replay；生产身份、secret、quota / billing 和交付按独立范围验收。
- 后端服务与 control plane 默认 Go；模型、评测和 AI 生态适配优先 Python；前端采用 React + Vite + TypeScript。
- 通过 canonical schema、明确类型、稳定函数边界和行为测试维持一致性；模块职责收敛按[架构](radishmind-architecture.md)和[工程健康专题](platform/engineering-health-productization-remediation-v1.md)逐项评审实施。
- 当前邀请批次 E 的独立授权线保持不变；本文的产品目标与验收建议不自动开启代码变更、后台服务、依赖安装或外部系统操作。

# RadishMind 阶段路线图

更新时间：2026-09-08

## 文档职责

本路线图只维护产品方向、阶段顺序、当前执行顺位和停止线。功能流程、数据边界与验收方式进入 [功能设计文档入口](features/README.md)，平台横切能力进入 [平台专题入口](platform/README.md)，具体实现批次进入 [任务卡入口](task-cards/README.md)，历史完成流水进入 [开发周志](devlogs/README.md) 或既有专题。

当前成熟度统一称为“内部开发者预览”。历史 `M3 / M4` 和 `P1` 至 `P7` 编号只用于定位既有证据与长期专题，不再解释为必须逐级晋升的成熟度等级，也不决定今天的开发顺位。当前执行决策以 [当前推进焦点](radishmind-current-focus.md) 和 [工程健康与产品化整改专题 v1](platform/engineering-health-productization-remediation-v1.md) 为准。

## 产品方向

`RadishMind` 是 `Radish` 体系下的 AI 工具、工作流、模型网关和 Copilot 集成平台。它拥有平台自身的本地账户、角色与工作区成员关系，但不替代或复制 `Radish` 自身身份、组织成员关系和业务数据真相源；模型输出默认是解释、诊断、结构化建议或候选动作，高风险动作必须经过人工确认或规则复核。

长期产品面保持四个：

1. `User Workspace`：应用、Prompt、Workflow、Agent / Copilot、RAG、API key、调用量、运行记录和成本摘要。
2. `Admin Control Plane`：本地注册、用户、会话、角色、权限、租户、provider/profile、模型路由、quota、price、secret backend、审计和部署状态；以本地账户为 owner，并可作为 OIDC client 接入 `Radish`。
3. `Model Gateway / API Distribution`：OpenAI-compatible / Responses / Messages / Models API 分发、多 provider / profile / model 路由，以及开发测试态 quota、价格快照、trace 与受控 fallback；生产限流、billing 与真实 health 分别后置验收。
4. `Workflow / Agent Runtime`：Prompt、LLM、condition、output，以及独立开发测试态 HTTP Tool 与 RAG 执行档案；高风险动作默认要求确认，通用 agent loop 继续后置。

图片输入理解与图片生成是横切适配能力，不是第五个一级产品面。图片像素生成继续由独立 `RadishMind-Image Adapter` 与 backend 承接，主模型只负责理解、规划、约束、审查和结构化意图。

四个产品面共同服务“内部开发者独立创建可复用 AI 应用、受控运行、审查结果与回归验证”。后续优先级由该流程的真实阻塞、维护风险和使用证据决定，不以新增功能、批次或画板数量衡量产品成熟度。

## 能力建设顺序

以下是职责与依赖顺序，不是仍待从头实施的清单；已完成事实见[能力矩阵](radishmind-capability-matrix.md)。

### 1. 平台与协议基座

- 统一 canonical contract、模型服务、Gateway、provider/profile 和基础评测。
- 保持结构化 JSON、失败关闭、密钥脱敏和上层业务真相源只读边界。
- 历史 `P1 Runtime Foundation`、`M3` 和 `M4` 证据继续可复验，但不再扩同层 readiness 链。

### 2. 工作流与用户闭环

- 优先完成 Workflow 草案创建、编辑、校验、保存、恢复、审查和安全执行。
- 运行历史、失败诊断、对比、评测 case / baseline / suite 与 Gateway 请求审查形成连续产品链。
- 用户工作区应用接入、配置草案、发布治理、应用目录和 API 密钥生命周期按真实使用路径验收。

### 3. 开发测试态持久化与身份边界

- RadishMind 自有运行数据允许使用明确命名的 SQLite / PostgreSQL 开发测试态 repository。
- migration、作用域、原子并发、重启恢复、运行角色和 no fallback 必须有可执行证据。
- 先完成本地账户、session、角色与 workspace membership，再完成确定性 OIDC Relying Party；真实 Radish 联调与 production auth 仍必须等待上游 client registration、secret、部署资源和负责人。

### 4. 运行时与工程治理收敛

- Gateway 使用受控 `stdio` worker pool，保留显式 process 回滚模式。
- 测试、覆盖率、性能预算、PR / release CI 与仓库门禁按风险分层维护。
- 入口文档、活动 checker、fixture 和 task card 必须回到各自职责；历史证据可索引、可手工复验，但不占当前主线。

### 5. 外部接入、生产化与模型适配

- `RadishFlow` 优先于 `Radish`；真实挂载点、owner、协议和验收环境明确后再恢复接入。
- production secret、环境隔离、process supervisor、quota / billing、生产 API key 和公开生产声明分别验收。
- `RadishMind-Core` 采用开源基座加自有协议、数据与评测偏好适配；没有评测和运行窗口时不启动训练或长跑。

## 当前执行顺位

1. 产品线：[工作区成员邀请、认领与到期治理（开发 / 测试态）v1](features/admin-control-plane/workspace-member-invitation-claim-expiry-governance-dev-test-v1.md)已完成 A 至 E，状态为 `workspace_member_invitation_claim_expiry_governance_dev_test_v1_completed`。React 消费层、S7 管理端与 Authentication 认领入口、双数据库产品链、三视口与双标签验收已闭合，下一步按下表选择实际用户阻塞或维护目标，不派生邀请批次 F。[应用定时回归评测](features/user-workspace/application-evaluation-scheduled-regression-campaign-dev-test-v1.md)保持完成关闭；真实 Provider、production worker、真实 Radish 和 production IAM 尚未进入。
2. 工程线：`R2` 至 `R6` 已完成。R6 关闭评审确认活动 checker 数量和代码量均下降超过 `15%`；Provider、Production Ops 与 Control Plane formal UI 因仍缺少等价行为证据继续保留，不再派生独立清理批次。
3. `P3 Local Product Shell / Ops Surface` 保持 `local usable / read-only close`。普通只读 console 页面、evidence 面板和布局整理不自动形成新任务卡、fixture 或 checker。
4. [本地账户与 Radish OIDC 联合登录 v1](features/admin-control-plane/local-account-radish-oidc-federated-login-v1.md)已完成本地可执行的批次 A 至 D：identity owner、三种开发测试仓储、本地 Web Session HTTP、确定性 Authorization Code + PKCE、当前账户 / revoke API、完整 Pencil、Web strict consumer、S7 当前账户 owner 与浏览器连续链已经闭合。真实 Radish 批次 E 保持 `real_radish_integration_deferred`，不与当前本地成员管理专题耦合。dev header、signed-test membership 与 loopback issuer 不能作为 production 授权来源；production secret backend、真实 provider credential / endpoint、自动路由、process supervisor、console production packaging、生产认证、production API key、production quota 和 billing 继续为 `not_satisfied`。
5. 当前没有独立工程整改批次；后续只在真实功能实现中复用、补强或替代相关行为证据，不自动删除历史 fixture，也不新建同层治理入口。

## 邀请闭环后的决策顺序

2026-09-06 审阅建议已纳入文档规划，尚未启动下表的代码、依赖、CI、服务或外部操作。邀请批次 E 已于 2026-09-08 独立获授权并完成；下表候选未因此自动启动，R2 至 R6 保持关闭。

| 顺序 | 工作方向 | 进入条件与验收 |
| --- | --- | --- |
| 1 | 用户主流程阻塞修正 | 在既有专题记录谁执行什么任务、卡在哪里、已有证据与修复后的可重复成功路径 |
| 2 | 前端状态与后端领域职责收敛 | 先选一个范围明确的工作区或领域，列出依赖与迁移边界；以作用域、并发、恢复和旧行为验证验收 |
| 3 | 自动浏览器回归 | 选择少量已实现的关键流程，明确环境、数据隔离、启动清理与 CI 成本；不把人工记录改写为自动覆盖 |
| 4 | 内部真实任务试用 | 先确认用户、输入、模型 / Provider 与运行窗口，建立任务完成率、耗时、求助点和输出质量基线 |
| 5 | 有明确窗口的接入与交付 | 依据试用结论与外部条件选择功能扩展、真实集成或生产验收，不从历史 readiness 尾链自动续批 |

具体建议与验证边界集中在[工程健康专题的后续方向](platform/engineering-health-productization-remediation-v1.md#2026-09-06-复审与后续方向)。保持一条产品线、必要时一条工程线；候选顺序不代表同时开工或已批准架构变更。

## 权威入口

- 当前决策：[当前推进焦点](radishmind-current-focus.md)
- 工程整改：[工程健康与产品化整改专题 v1](platform/engineering-health-productization-remediation-v1.md)
- 功能设计：[功能设计文档入口](features/README.md)
- 产品边界：[产品范围](radishmind-product-scope.md)
- 系统边界：[架构](radishmind-architecture.md)
- 协议边界：[集成契约](radishmind-integration-contracts.md)
- 能力状态：[能力矩阵](radishmind-capability-matrix.md)
- 当前开发记录：[2026-W36 周志](devlogs/2026-W36.md)

## 停止线

- 不把开发测试态数据库、fake / local adapter、离线 smoke、静态 schema artifact、当前本地 console 或部分覆盖率写成 production ready。
- 不在缺少本地账户 / session / membership owner、真实 issuer / client registration、生产数据库资源、secret backend、部署环境、负责人和发布复核时启用生产能力。
- 不把 Control Plane、Gateway 和 Workflow Executor 混成隐式单体，也不因服务拆分引入新的默认语言栈。
- 不把 task card、fixture、checker、readiness 或周志当成功能设计文档和当前决策入口的替代品。
- 不继续派生“readiness 之后的 readiness”，不把历史 next dependency 恢复为当前开发顺位。
- 不在上层项目没有真实挂载点时细化假想接线，不跨工作区修改 `RadishFlow`、`Radish` 或 `RadishCatalyst`。
- 不让模型建议直接写入上层业务真相源，不启用 unrestricted tool、业务写回、自动确认提交或 replay。
- 不让模型、客户端或人工批准直接授予 Action Safety 有效级别；v1 不实现 `write_allowed_by_policy`、写方法 Tool、通用动作执行器或新的 Confirmation / Run / Audit owner。
- 不让 API key、token、DSN、provider 原始响应或异常正文进入 argv、公开错误、日志或 committed 资产。
- 不在缺少评测基线和明确运行窗口时下载模型、长跑真实模型、扩训练 JSONL、蒸馏或权重工作。

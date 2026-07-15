# RadishMind 阶段路线图

更新时间：2026-07-15

## 文档职责

本路线图只维护产品方向、阶段顺序、当前执行顺位和停止线。功能流程、数据边界与验收方式进入 [功能设计文档入口](features/README.md)，平台横切能力进入 [平台专题入口](platform/README.md)，具体实现批次进入 [任务卡入口](task-cards/README.md)，历史完成流水进入 [开发周志](devlogs/README.md) 或既有专题。

当前成熟度统一称为“内部开发者预览”。历史 `M3 / M4` 和 `P1` 至 `P7` 编号只用于定位既有证据与长期专题，不再解释为必须逐级晋升的成熟度等级，也不决定今天的开发顺位。当前执行决策以 [当前推进焦点](radishmind-current-focus.md) 和 [工程健康与产品化整改专题 v1](platform/engineering-health-productization-remediation-v1.md) 为准。

## 产品方向

`RadishMind` 是 `Radish` 体系下的 AI 工具、工作流、模型网关和 Copilot 集成平台。它不替代 `Radish` 的身份、组织成员关系和业务数据真相源；模型输出默认是解释、诊断、结构化建议或候选动作，高风险动作必须经过人工确认或规则复核。

长期产品面保持四个：

1. `User Workspace`：应用、Prompt、Workflow、Agent / Copilot、RAG、API key、调用量、运行记录和成本摘要。
2. `Admin Control Plane`：租户、用户、角色、权限、provider/profile、模型路由、quota、price、secret backend、审计和部署状态；未来作为 OIDC client 接入 `Radish`。
3. `Model Gateway / API Distribution`：OpenAI-compatible / Responses / Messages / Models API 分发，多 provider / profile / model 路由，以及后续 quota、限流、成本、trace、受控 fallback 和 health。
4. `Workflow / Agent Runtime`：Prompt、LLM、condition、output，以及后续受控 HTTP tool、RAG retrieval 和 agent loop；高风险动作默认要求确认。

图片输入理解与图片生成是横切适配能力，不是第五个一级产品面。图片像素生成继续由独立 `RadishMind-Image Adapter` 与 backend 承接，主模型只负责理解、规划、约束、审查和结构化意图。

## 阶段顺序

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
- verified identity、负向认证和确定性 OIDC integration test 可以独立完成；真实 Radish 联调与 production auth 必须等待上游资源和负责人。

### 4. 运行时与工程治理收敛

- Gateway 使用受控 `stdio` worker pool，保留显式 process 回滚模式。
- 测试、覆盖率、性能预算、PR / release CI 与仓库门禁按风险分层维护。
- 入口文档、活动 checker、fixture 和 task card 必须回到各自职责；历史证据可索引、可手工复验，但不占当前主线。

### 5. 外部接入、生产化与模型适配

- `RadishFlow` 优先于 `Radish`；真实挂载点、owner、协议和验收环境明确后再恢复接入。
- production secret、环境隔离、process supervisor、quota / billing、生产 API key 和公开生产声明分别验收。
- `RadishMind-Core` 采用开源基座加自有协议、数据与评测偏好适配；没有评测和运行窗口时不启动训练或长跑。

## 当前执行顺位

1. 产品线：工作流审查、Gateway 请求审查、用户工作区应用接入 / 配置 / 发布 / 目录 / API 密钥，以及本地 SQLite 与 PostgreSQL 开发测试态证据链均已关闭。下一项回到 Workflow 安全执行主线，先建立“Workflow 受控 HTTP Tool 与人工确认执行（开发/测试态）v1”功能设计并完成边界评审，通过后再拆 API、schema 与执行边界实现批次。
2. 工程线：`R2` 至 `R5` 已完成。`R6` 前五批已完成入口文档、Session / Tooling、Image Path 与 Control Plane Read 活动门禁收敛；入口不再重复保存功能 fixture、readiness 流水和完整运行配置。Control Plane Read 只退出已有 Go 行为与跨端消费证据承接的八项早期静态检查，formal UI 与页面链因 TypeScript 行为覆盖不足继续活动。
3. `P3 Local Product Shell / Ops Surface` 保持 `local usable / read-only close`。普通只读 console 页面、evidence 面板和布局整理不自动形成新任务卡、fixture 或 checker。
4. 真实 Radish 联调保持 `real_radish_integration_deferred`；production secret backend、process supervisor、部署环境隔离、console production packaging、生产认证、生产 API key、quota 和 billing 继续为 `not_satisfied`。
5. 下一工程批做 R6 关闭评审：复算活动 checker / task card 的收敛幅度，并逐项确认 Provider、Production Ops 与保留 formal UI 门禁是否有等价行为证据；没有证据就保持现状，不自动删除历史 fixture，也不新建同层治理入口。

## 权威入口

- 当前决策：[当前推进焦点](radishmind-current-focus.md)
- 工程整改：[工程健康与产品化整改专题 v1](platform/engineering-health-productization-remediation-v1.md)
- 功能设计：[功能设计文档入口](features/README.md)
- 产品边界：[产品范围](radishmind-product-scope.md)
- 系统边界：[架构](radishmind-architecture.md)
- 协议边界：[集成契约](radishmind-integration-contracts.md)
- 能力状态：[能力矩阵](radishmind-capability-matrix.md)
- 历史完成记录：[2026-W29 周志](devlogs/2026-W29.md)

## 停止线

- 不把开发测试态数据库、fake / local adapter、离线 smoke、静态 schema artifact、当前本地 console 或部分覆盖率写成 production ready。
- 不在缺少真实 issuer、membership source、生产数据库资源、secret backend、部署环境、负责人和发布复核时启用生产能力。
- 不把 Control Plane、Gateway 和 Workflow Executor 混成隐式单体，也不因服务拆分引入新的默认语言栈。
- 不把 task card、fixture、checker、readiness 或周志当成功能设计文档和当前决策入口的替代品。
- 不继续派生“readiness 之后的 readiness”，不把历史 next dependency 恢复为当前开发顺位。
- 不在上层项目没有真实挂载点时细化假想接线，不跨工作区修改 `RadishFlow`、`Radish` 或 `RadishCatalyst`。
- 不让模型建议直接写入上层业务真相源，不启用 unrestricted tool、业务写回、自动确认提交或 replay。
- 不让 API key、token、DSN、provider 原始响应或异常正文进入 argv、公开错误、日志或 committed 资产。
- 不在缺少评测基线和明确运行窗口时下载模型、长跑真实模型、扩训练 JSONL、蒸馏或权重工作。

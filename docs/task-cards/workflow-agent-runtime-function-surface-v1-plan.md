# `Workflow / Agent Runtime Function Surface` v1 计划

更新时间：2026-06-07

## 任务目标

本任务卡用于把当前默认推进方向从 Docker / production preflight 等部署复核窗口，切回产品功能骨架建设。它只定义 `Workflow / Agent Runtime` 在真实执行器、真实数据库和生产认证接入前，可以先推进的用户可见功能面、资源关系、API 草案边界和 blocked 状态。

当前任务不实现 workflow executor、不实现 workflow builder 写能力、不接真实数据库、不接 `Radish` OIDC、不实现 confirmation flow、不做 business writeback 或 replay。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [产品范围与目标](../radishmind-product-scope.md)
- [阶段路线图](../radishmind-roadmap.md)
- [能力矩阵](../radishmind-capability-matrix.md)
- [`Control Plane / User Workspace / Workflow` v1 计划](control-plane-user-workspace-workflow-v1-plan.md)
- [`P2 Session & Tooling` executor / storage / confirmation enablement plan](session-tooling-executor-storage-confirmation-enablement-plan.md)
- [`P2 Session & Tooling` stop-line manifest](session-tooling-stop-line-manifest.md)

## 当前判断

- Production Ops 静态边界和一次 `docker_local` smoke 运行记录已经足够支撑现阶段；没有测试或生产前复核窗口时，不继续把 Docker / deployment 作为默认推进方向。
- read-side repository / database / auth 迁移当前没有 implementation trigger satisfied；不应直接实现 repository interface、adapter、SQL、migration、store selector、OIDC token validation 或 production API consumer。
- 项目目标仍是 AI 工具、工作流、模型网关和 Copilot 集成平台；现阶段应优先补用户能理解的功能骨架，而不是继续扩同层治理切片。

## v1 功能面边界

1. `Application`
   - 展示 AI 应用 / Prompt 应用 / Workflow 应用的名称、类型、状态、provider profile、风险摘要和最近运行。
   - 不提供创建、编辑、删除或发布写能力。
2. `Workflow Definition`
   - 展示 workflow definition detail、节点列表、节点类型、输入输出摘要、风险等级、requires confirmation 标记和版本信息。
   - 不提供 workflow builder、节点编辑、版本发布或 definition mutation。
3. `Run Record`
   - 展示 run detail、状态流转、输入输出摘要、cost / token snapshot、trace id、failure code 和 audit refs。
   - 不提供 run replay、run resume、result materialization reader 或 durable run store。
4. `Tool Action`
   - 展示候选 tool action、risk level、requires_confirmation、policy reason、审计引用和 blocked state。
   - 不执行工具，不写上层业务真相源。
5. `Confirmation`
   - 展示未来 confirmation decision 的数据形状、风险说明和人工确认入口占位。
   - 不实现确认提交、决策持久化、执行解锁或写回。

## 建议切片

1. `workflow-function-surface-boundary-v1`
   - 已落地为 `function_surface_boundary_defined`，固定 application detail、workflow definition detail、run detail、tool action preview 和 confirmation placeholder 的字段边界。
   - 不新增真实 API、不实现 executor、不创建数据库 schema。
2. `workflow-definition-detail-read-v1`
   - 在现有 read-side shell 上增加 workflow definition detail 的只读视图或 route 草案，优先消费 fixture / fake-store-backed dev path。
   - 只展示节点结构、风险和 blocked action，不提供编辑或执行。
3. `workflow-run-detail-read-v1`
   - 已落地为 `workflow_run_detail_read_defined`，增加 `apps/radishmind-web/` 内 run detail 离线只读视图，展示状态流转、trace、成本摘要、failure code 和 audit refs。
   - 不提供 replay、resume、materialized result reader 或真实 run store。
4. `workflow-blocked-action-preview-v1`
   - 已落地为 `workflow_blocked_action_preview_defined`，复用 tooling blocked action shell，展示 tool/action 为什么被阻止、需要什么 confirmation、当前缺哪些前置条件。
   - 不执行 tool、不写回、不创建 confirmation decision。
5. `workflow-confirmation-placeholder-read-v1`
   - 已落地为 `workflow_confirmation_placeholder_read_defined`，展示 required action ref、risk summary、required decision shape、human review requirement、disabled reason、route / request / audit metadata 和 missing prerequisites。
   - 不提交 confirmation，不持久化 decision，不解锁执行，不写回。

## 验收口径

- 当前焦点、路线图、能力矩阵和周志都明确：无 Docker 运行窗口时，Production Ops 降为等待项，功能骨架成为默认下一步。
- `workflow-function-surface-boundary-v1` 已用 fixture / checker 固定 `function_surface_boundary_defined`，后续 detail read 切片不得越过该边界。
- 任务卡明确下一批功能切片优先从只读 detail、blocked action preview 和 fake-store dev path 开始。
- 文档继续保留 read-side implementation trigger 未满足的结论。
- 文档继续明确不实现 repository interface / adapter、SQL、migration、store selector、OIDC token validation、auth middleware、真实数据库、production API consumer、workflow executor、confirmation、writeback 或 replay。

## 非目标

- 不实现 workflow builder 写能力。
- 不实现 workflow executor、node executor、tool executor 或 agent loop。
- 不实现 durable run store、durable result store、materialized result reader、run replay 或 run resume。
- 不实现 confirmation decision 持久化、执行解锁或业务写回。
- 不实现真实数据库、repository adapter、schema migration、store selector、Radish OIDC、token validation 或 production API consumer。
- 不启动 Docker、不拉镜像、不发布镜像、不写入 secret。

## 停止线

- blocked action preview 只能解释“为什么不能执行”和“未来需要什么条件”，不能触发真实工具调用。
- workflow definition detail 和 run detail 只能读取 fixture / fake-store dev 数据，不得声明 production API consumer ready。
- confirmation placeholder 不能被解释为 confirmation flow 已接通。
- 任何写能力、执行能力、持久化能力或生产身份能力，必须另开实现任务卡并先满足对应 gate。

# RadishMind 跨项目集成契约

更新时间：2026-06-07

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
- `contracts/session-recovery-checkpoint-read.schema.json`
- `contracts/image-generation-intent.schema.json`
- `contracts/image-generation-backend-request.schema.json`
- `contracts/image-generation-artifact.schema.json`
- `contracts/production-secret-reference.schema.json`

## 当前协议原则

- 统一骨架 + 项目专属上下文块。
- 结构化 JSON 优先，图像和附件作为 artifact 补充。
- 默认 advisory mode，不做直接写回。
- 所有高风险输出都必须带 `requires_confirmation`。
- 兼容层只做翻译，不另起第二套真相源。
- 上层项目只消费建议、解释、候选动作和审计信息，最终业务真相源仍由上层维护。
- 用户端、管理端、模型网关和 workflow runtime 都必须复用同一套 canonical contract，不为每个产品面另起一套私有协议。
- 部署、数据库和登录 / 授权默认参考 `Radish`；未来 RadishMind 作为 OIDC client 接入 `Radish`，不把用户身份和权限真相源放进模型 runtime，也不默认引入 `.NET` / ASP.NET Core。
- `P2 Session & Tooling Foundation` 当前只声明 close candidate / governance-only；negative regression governance suite、deny-by-default gates、negative coverage rollup、route negative coverage matrix、route smoke readiness rollup、short close readiness delta、readiness consistency rollup、enablement plan 和 stop-line manifest 都是治理证据链，不代表真实执行、持久化、结果读取、confirmation 接线或 replay 已启用。
- `P3 Local Product Shell / Ops Surface` 已暴露只读 `GET /v1/platform/overview` 与 `GET /v1/platform/local-smoke`，并已有 overview / local-smoke console consumer smoke、`contracts/typescript/platform-overview-api.ts`、`contracts/typescript/platform-local-smoke-api.ts`、本地 console 壳、Dev Diagnostics、`Local Readiness` 面板、Provider/Profile Details、Stop-line Details、overview / local-smoke failure surface、behavior / visual smoke record / dev entry / production boundary gate 和 P3 checklist；本地只读产品壳已达到 `local usable / read-only close`。它只聚合和消费服务状态、本地 readiness、model/profile inventory、session/tooling metadata、blocked action route 和停止线，不引入真实 executor、durable store、confirmation 接线、长期记忆、业务写回或 replay；production secret backend、process supervisor、部署环境隔离和 console production packaging 仍是后续 hardening 缺口。
- `Provider Runtime & Health v1` 已把 provider capability matrix、provider health smoke、provider selection policy、provider retry/fallback policy 和 docs refresh 五个切片接入 fast baseline。它固定 `/v1/models`、provider/profile selection、diagnostics selectable model ids、credential state、deployment mode、offline health smoke、no implicit fallback、`retry_policy=caller-managed` 与 `fallback_policy=disabled` 的说明口径；它不代表 optional live health、retry/fallback execution、production secret backend、tool executor、confirmation/writeback/replay 或 production readiness 已完成。
- `Production Secret Reference` 已用 `contracts/production-secret-reference.schema.json` 固定 reference-only manifest：只允许 `ref:` 形式的 secret 引用和脱敏 readiness 字段，不保存 secret value，不启用 resolver，不调用云 API，也不声明 production secret backend ready。
- `Control Plane Read-Side` 已用 read model、read-only route contract、response fixture、negative contract、implementation preconditions、fake-store-backed handler plan、七条 handler implementation、`control-plane-read-auth-db-preconditions-v1`、`control-plane-read-consumer-contract-v1`、`control-plane-read-formal-ui-boundary-v1`、`control-plane-read-formal-ui-implementation-readiness-v1`、`control-plane-read-shared-shell-v1`、`control-plane-read-admin-tenant-overview-v1`、`control-plane-read-admin-audit-log-v1`、`control-plane-read-workspace-applications-v1`、`control-plane-read-workspace-api-keys-v1`、`control-plane-read-workspace-usage-quota-v1`、`control-plane-read-workspace-workflow-definitions-v1`、`control-plane-read-workspace-run-history-v1`、`control-plane-read-formal-ui-readiness-close-v1`、`control-plane-read-dev-live-consumer-v1` 和 `control-plane-read-auth-store-transition-preconditions-v1` 固定 user workspace / admin control plane 的只读查询边界、正式 UI 边界、只读 `admin-tenant-overview`、只读 `admin-audit-log`、只读 `workspace-applications`、只读 `workspace-api-keys`、只读 `workspace-usage-quota`、只读 `workspace-workflow-definitions`、只读 `workspace-run-history`、页面集合聚合收口、dev-only live consumer 和 auth/store transition preconditions。当前只使用 in-memory fixture fake store 与 test-only fake auth context；`contracts/typescript/control-plane-read-api.ts` 定义上层消费契约，`apps/radishmind-web/` 默认消费离线 read-side view model，显式 `dev_live_http` 只允许连接 fake-store-backed handler 和测试身份上下文，`future control plane read store repository` 仍只是前置条件，不实现完整 read-side API、production API consumer、数据库 query、OIDC、token validation、repository migration、repository implementation、API key lifecycle、quota enforcement、rate limit、billing、cost ledger、workflow builder、workflow lifecycle mutation、workflow executor、run replay、run resume、materialized result reader、confirmation、writeback 或 replay。
- read-side UI 门禁已从逐页专项证明转向能力边界与聚合门禁优先：普通只读展示页由 surface matrix / checker 管理；dev-only live read path 只验证 HTTP consumer shape；auth/store transition preconditions 只定义未来迁移 gates。三者都不改变真实数据库、Radish OIDC、API key / quota、workflow executor、confirmation、writeback 或 replay 的停止线。
- `Control Plane Read-Side repository transition` 当前只固定未来 read store 的 contract、受控类型边界、静态 runner、interface readiness、adapter readiness refresh、selector enablement preconditions、schema migration implementation preconditions、repository adapter implementation plan、schema artifact manifest readiness、store selector smoke readiness、production auth readiness、adapter smoke readiness 和 implementation trigger review：repository contract preconditions、disabled database guard、repository contract smoke、repository implementation readiness、store selection readiness、schema migration readiness、repository contract types readiness、repository contract types implementation、`control-plane-read-repository-contract-smoke-runner-readiness-v1`、`control-plane-read-repository-contract-smoke-runner-implementation-v1`、`control-plane-read-repository-interface-readiness-v1`、`control-plane-read-repository-adapter-implementation-readiness-refresh-v1`、`control-plane-read-store-selector-enablement-preconditions-v1`、`control-plane-read-schema-migration-implementation-preconditions-v1`、`control-plane-read-repository-adapter-implementation-plan-v1`、`control-plane-read-schema-artifact-manifest-readiness-v1`、`control-plane-read-store-selector-smoke-readiness-v1`、`control-plane-read-production-auth-readiness-v1`、`control-plane-read-adapter-smoke-readiness-v1` 和 `control-plane-read-implementation-trigger-review-v1` 已进入 fast baseline。它们共同要求七条 read route 未来迁移时保持统一 input/output envelope、tenant-scoped repository context、route request / result type、runner type matrix 消费、future interface method matrix、adapter gate、selector enablement / smoke gates、migration artifact manifest、DDL review evidence、rollback fixture evidence、schema version / tenant index / read-only role smoke、repository adapter implementation plan、schema artifact manifest readiness、production auth readiness、adapter smoke readiness、implementation trigger review、failure mapping、`database_read_disabled` / `invalid_read_store_mode` / `schema_migration_not_applied` / `schema_version_mismatch` / `tenant_binding_missing` / `scope_denied` 等 fail-closed code、no fake fallback 和 no side effects；当前只落地 Go contract type 文件、静态 contract/type runner、interface readiness、repository adapter implementation readiness refresh、selector enablement preconditions、schema migration implementation preconditions、repository adapter implementation plan、schema artifact manifest readiness、store selector smoke readiness、production auth readiness、adapter smoke readiness 和 implementation trigger review，不实现 repository interface、SQL、migration runner、repository adapter、store selector、adapter smoke、真实数据库、Radish OIDC、token validation、auth middleware 或 production API consumer。
- `Workflow / Agent Runtime Function Surface` 当前已用 `workflow-function-surface-boundary-v1` 固定 `function_surface_boundary_defined`，并用 `workflow-definition-detail-read-v1` 固定 `workflow_definition_detail_read_defined`、用 `workflow-run-detail-read-v1` 固定 `workflow_run_detail_read_defined`、用 `workflow-blocked-action-preview-v1` 固定 `workflow_blocked_action_preview_defined`：workflow definition detail、run detail 与 blocked action preview 已能在 `apps/radishmind-web/` 以离线只读 surface 展示 nodes、edges、input/output summary、state timeline、cost/token snapshot、trace/failure/audit metadata、blocked action preview、missing prerequisites、confirmation placeholder、audit trail 和 blocked replay / result preview；它不新增跨项目私有协议，不实现 workflow executor、tool executor、confirmation decision、execution unlock、business writeback、run replay、run resume、durable result store、真实数据库、Radish OIDC 或 production API consumer。

## 专题索引

- [服务/API 接入契约](contracts/service-api.md)：northbound / southbound 兼容边界、`CopilotGatewayEnvelope`、`RadishFlow` UI consumption、candidate edit handoff、上层接入等待口径和仓库集成边界。
- [Control Plane Read-Side 契约](contracts/control-plane-read-side.md)：control plane / user workspace 的只读 summary、route、response fixture、negative contract、fake-store-backed handler implementation、auth/db preconditions、consumer contract、formal UI boundary/readiness、repository/read store transition readiness、脱敏输出和停止线。
- [会话记录契约](contracts/session.md)：`Conversation & Session` 的 `session_id / turn_id`、history policy、recovery record、northbound session metadata、metadata-only checkpoint read、promotion gate、readiness rollup、stop-line manifest、负向查询和 advisory-only audit 边界。
- [工具框架契约](contracts/tooling.md)：`Tooling Framework` 的 tool definition、registry、policy/audit record、metadata-only result cache、negative regression governance suite、deny-by-default gates、result materialization policy、executor/storage 边界和不执行真实工具的 v1 停止线。
- [Production Secret Reference 契约](contracts/production-secret-reference.md)：provider profile 到 secret reference 的 reference-only manifest、脱敏字段、禁止字段和 production secret backend 未就绪停止线。
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
- 不让用户端、管理端、模型网关或 workflow runtime 私自分叉协议。
- 不自建与 `Radish` 冲突的身份、授权、数据库和部署真相源。
- 不把跨项目协议对齐解释成后端语言栈复制。
- 不把 secret reference manifest 写成真实 secret backend、secret resolver、provider credential readiness 或 production ready。
- 不让模型输出直接写回上层项目。
- 不把 checkpoint read route smoke 写成 durable checkpoint store、materialized result reader、executor ref reader、durable memory reader 或 replay executor。
- 不把 P2 design gate 写成上层确认流接线、真实 executor、durable session/checkpoint/audit/result store、长期记忆、业务写回或完整负向回归已经完成。
- 不在 `RadishCatalyst` 进入真实任务、adapter skeleton 和最小 eval sample 前扩 schema 枚举或 gateway route。

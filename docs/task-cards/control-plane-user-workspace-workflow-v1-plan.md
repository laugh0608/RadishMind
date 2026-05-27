# `Control Plane / User Workspace / Workflow` v1 计划

更新时间：2026-05-27

## 任务目标

本任务卡用于把平台重定义后的四个正式产品面，收口为下一阶段可实施、可验证、可停止的 v1 边界。

当前任务不实现真实用户端、不实现生产管理端、不接 `Radish` OIDC、不创建数据库 schema、不实现 workflow executor，也不新增 confirmation / writeback / replay。它只固定 `User Workspace`、`Admin Control Plane`、`Model Gateway / API Distribution` 和 `Workflow / Agent Runtime` 的服务边界、数据边界、前置条件和阶段停止线，避免后续直接从本地 ops console 堆功能。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [产品范围与目标](../radishmind-product-scope.md)
- [阶段路线图](../radishmind-roadmap.md)
- [能力矩阵](../radishmind-capability-matrix.md)
- [系统架构](../radishmind-architecture.md)
- [战略定义](../radishmind-strategy.md)
- [Production Secret Backend Implementation v1 计划](production-secret-backend-implementation-v1-plan.md)
- [Provider Runtime & Health v1 计划](provider-runtime-health-v1-plan.md)
- `services/platform/`
- `apps/radishmind-console/`
- `contracts/`

## 当前事实

- `P1 Runtime Foundation` 已达到 short close。
- `P2 Session & Tooling Foundation` 处于 `close candidate / governance-only`，当前只有 metadata / blocked shell，不具备真实 executor、durable store、confirmation flow、materialized result reader、长期记忆或 replay。
- `P3 Local Product Shell / Ops Surface` 已达到 `local usable / read-only close`，但 `apps/radishmind-console/` 仍只是本地 ops surface，不是正式用户端或生产管理端。
- `Production Ops Hardening v1` 已达到 static boundary close，并已有一次本地 `docker_local` container smoke 运行记录；测试环境 smoke、production preflight、process supervisor、正式 auth / CORS policy、真实镜像发布和 production secret backend 仍未完成。
- `Provider Runtime & Health v1` 已进入 close candidate；provider capability、health smoke、selection policy、retry/fallback audit metadata 和 docs refresh 已可检查，但 optional live health、retry/fallback execution、cost / quota / tenant binding 和 production secret backend 仍未实现。
- 正式产品面已固定为 `User Workspace`、`Admin Control Plane`、`Model Gateway / API Distribution` 和 `Workflow / Agent Runtime`。
- 长期技术栈固定为 Go control plane / gateway、Python 模型与评测、TypeScript/Vite 前端；不因参考 `Radish` 默认引入 `.NET`。

## v1 边界

### 1. `User Workspace`

v1 只定义用户工作区的产品和数据边界：

- AI 应用、Prompt 应用、Workflow、Agent / Copilot 应用和 RAG / 知识问答应用的最小资源模型。
- 用户 API key、调用量、运行记录、成本摘要的展示与查询边界。
- 应用运行只能通过受控 runtime / gateway 进入，不直接写上层业务真相源。
- 用户端只消费已授权的 tenant / project / application 资源，不拥有最终身份真相。

当前不实现正式用户端 UI、不实现 API key 生命周期、不实现运行记录存储，也不把本地 ops console 改名成 user workspace。

### 2. `Admin Control Plane`

v1 只定义管理端和平台控制面的职责：

- tenant、user、role、permission、provider/profile、model route、quota、price、secret backend、audit、deployment status 的管理边界。
- `Radish` Auth 是未来身份真相源，RadishMind 作为 OIDC client 接入，不自建第二套身份与权限真相源。
- Control Plane 默认使用 Go，可按职责拆服务，但不因为拆分而新增默认后端语言栈。
- 管理端只能配置和审计平台能力，不直接绕过 runtime 执行业务写回。

当前不实现 OIDC client、不建数据库、不实现 tenant / quota / billing API，也不把本地 ops console 写成 production admin console。

### 3. `Model Gateway / API Distribution`

v1 只定义模型 API 分发从现有 provider runtime 继续前进的边界：

- API key、tenant binding、quota、rate limit、cost ledger、trace、provider profile、secret ref 和 model route 的最小控制面关系。
- northbound API 继续兼容 `/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models`，但业务 truth 仍归 canonical protocol。
- provider selection、health signal、retry/fallback execution 和 cost / latency policy 必须显式配置和审计。
- production secret 只通过 `secret_ref` 和未来 resolver 边界进入，不提交 secret value。

当前不实现真实 API key 分发、不实现 quota / billing、不启用自动 fallback、不接真实 production secret backend，也不声明完整模型 API 分发平台已完成。

### 4. `Workflow / Agent Runtime`

v1 只定义 workflow runtime 的资源边界和执行停止线：

- Prompt、LLM、HTTP tool、RAG retrieval、condition、output、后续受控 code / sandbox / agent loop 的最小节点分类。
- 每次 workflow run 必须有 trace、输入输出摘要、错误分类、成本字段、风险边界和审计引用。
- 高风险 tool/action 默认 `requires_confirmation`，未经确认不得执行写回。
- workflow definition、run record、tool audit、result materialization 和 confirmation decision 必须分层存储和审计。

当前不实现 workflow builder、不实现 executor、不接 tool calling、不实现 durable result store、不启用 confirmation flow、不做 writeback 或 replay。

## 建议切片

1. `product-surface-v1-boundary`
   - 固定四个产品面的 v1 resource、read model、write boundary 和停止线。
   - 不新增 UI 或 API 实现。
   - 当前已落地：`scripts/checks/fixtures/product-surface-v1-boundary.json` 与 `scripts/check-product-surface-v1-boundary.py`，状态为 `governance_boundary_satisfied`；该切片只固定四个产品面的资源、读模型、写边界、blocked capability 和 shared stop-line，不实现正式用户端、生产管理端、OIDC、数据库、API key / quota、workflow executor、confirmation、writeback 或 replay。
2. `control-plane-data-boundary`
   - 明确 tenant、user、role、permission、provider profile、model route、quota、price、audit、secret ref、deployment status 的 ownership。
   - 不创建数据库迁移。
3. `radish-oidc-client-preconditions`
   - 固定未来接入 `Radish` Auth 前必须具备的 issuer、client、claim mapping、tenant binding、logout、audit 和 failure taxonomy。
   - 不接真实 OIDC。
4. `gateway-api-key-quota-readiness`
   - 定义 API key、quota、rate limit、cost ledger 和 trace 的契约前置条件。
   - 不发放真实 API key。
5. `workflow-definition-run-record-boundary`
   - 定义 workflow definition、run record、node execution、tool audit、result materialization 和 confirmation decision 的最小关系。
   - 不实现 executor。

## 验收口径

- 有任务卡固定正式产品面的 v1 边界、输入事实源、建议切片、非目标和停止线。
- `product-surface-v1-boundary.json` 与 `check-product-surface-v1-boundary.py` 已进入 `check-repo --fast`。
- 当前焦点、路线图、能力矩阵、任务卡入口和周志同步说明：下一条平台主线是先固定 control plane / user workspace / workflow v1 边界。
- 文档继续明确本地 ops console 不等同于正式用户端或生产管理端。
- 文档继续明确 `Provider Runtime & Health v1`、`Production Ops Hardening v1`、P2 / P3 / UI / P4 都不再默认扩同层小切片。
- `pwsh ./scripts/check-repo.ps1 -Fast` 通过。

## 非目标

- 不实现正式用户端。
- 不实现 production admin console。
- 不接 `Radish` OIDC。
- 不创建数据库 schema 或 migration。
- 不实现 API key 生命周期、quota、billing、cost ledger 或 tenant binding。
- 不实现 workflow builder、workflow executor、tool executor、confirmation flow、writeback、replay 或 materialized result reader。
- 不接真实 production secret backend，不写入真实 secret。
- 不重开真实模型长跑、训练 JSONL、蒸馏或权重相关工作。
- 不重复跑同一 `docker_local` smoke，除非 compose、Dockerfile、console build 或 platform startup 有新变更。

## 停止线

- 不把 `apps/radishmind-console/` 改写成正式用户端或生产管理端。
- 不让 Control Plane、Gateway 和 Workflow Executor 混成单体。
- 不自建与 `Radish` 冲突的身份、权限和部署真相源。
- 不把微服务拆分解释成新增后端语言栈的理由。
- 不把 provider health、local smoke、Docker 静态边界或本地 mock smoke 写成 production ready。
- 不让模型、workflow 或 tool 直接写上层业务真相源。

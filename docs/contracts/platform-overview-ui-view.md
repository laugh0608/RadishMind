# Platform Overview UI View 契约

更新时间：2026-05-20

## 文档目的

本文档固定本地 console 或上层 UI 消费 `P3 Local Product Shell / Ops Surface` overview 时的首版视图边界。当前目标是说明 UI 应如何展示平台状态、model/profile inventory、session/tooling 产品面、停止线、refresh 状态、Dev Diagnostics 和连接失败诊断；`apps/radishmind-console/` 已提供 React + Vite + TypeScript console 壳，但仍不接入真实 executor、durable store、confirmation flow 或 replay。

TypeScript 消费类型真相源为 `contracts/typescript/platform-overview-api.ts`。开发者可用 `scripts/run-platform-overview-consumer-smoke.py --check` 生成离线消费视图；本地平台服务启动后可加 `--base-url http://127.0.0.1:7000` 请求真实 `GET /v1/platform/overview`，或启动 `apps/radishmind-console/` 查看同一只读视图。`scripts/check-radishmind-console-behavior.py` 固定 ready、refresh、连接失败诊断和只读停止线；`scripts/check-radishmind-console-production-boundary.py` 固定 console production packaging 仍未完成；`scripts/check-p3-local-product-shell-short-close-checklist.py` 固定 P3 short close 仍为 `not_ready`，不需要启动浏览器或长驻服务。

平台服务当前只允许 `http://127.0.0.1:4000` 与 `http://localhost:4000` 作为本地 console origin 读取 API，并处理 `OPTIONS` preflight；这只是本地开发边界，不代表 production CORS policy 或正式鉴权。

`GET /v1/platform/local-smoke` 是 overview 的本地 readiness 配套入口。它不替代 overview，也不提供更多产品能力；它只把 healthz、overview contract、model inventory、session/tooling metadata、blocked action no-side-effects、本地 CORS origin、默认端口和停止线汇总成开发期可读状态。UI 如要展示 local-smoke，应按 [Platform Local Smoke UI View 契约](platform-local-smoke-ui-view.md) 消费，不得把 `local_console_ready=true` 展示成 production ready、supervisor ready 或 executor ready。

## 输入来源

来源：`GET /v1/platform/overview`

该 route 聚合已有只读面：

- `/healthz` 的 service status
- `/v1/models` 与 `/v1/models/{id}` 的 provider/profile inventory
- `/v1/session/metadata` 的 session metadata route
- `/v1/tools/metadata` 的 tool metadata route
- `/v1/tools/actions` 的 blocked action route
- 当前 executor、durable store、confirmation、long-term memory、business writeback 和 replay 停止线

overview 只复用既有 metadata / blocked shell，不是第二套业务真相源。

## 视图模型

### Service status

UI 只展示：

- `serviceName`
- `version`
- `status`
- `stage`
- `mode=local_read_only_product_shell`
- `overviewRoute=/v1/platform/overview`
- `healthyForLocalConsole`

该视图只说明本地产品壳可被读取，不代表 production deployment、process supervision、secret backend 或正式鉴权已完成。

### Model inventory

UI 只展示：

- `status`
- `inventoryKind=bridge_backed_provider_profile_inventory`
- `modelsRoute=/v1/models`
- `detailRoute=/v1/models/{id}`
- `defaultProvider`
- `defaultProfile`
- `defaultModel`
- `modelCount`
- `providerCount`
- `profileCount`
- `selectableModelIds`
- `activeProfileChain`
- `canShowProfileSelector`

`canShowProfileSelector=true` 只表示 UI 可展示 provider/profile 选择信息；它不表示 provider health check、production secret backend、真实外部调用策略或成本/重试策略已经完成。

### Session / tooling surface

UI 只展示：

- `sessionMetadataRoute=/v1/session/metadata`
- `toolsMetadataRoute=/v1/tools/metadata`
- `blockedActionRoute=/v1/tools/actions`
- `metadataOnly=true`
- `executionEnabled=false`
- `actionStatusLabel=blocked`
- `toolCount`
- `requiresConfirmationPath`

该视图只说明 session/tooling shell 可展示，不代表真实工具执行器、confirmation handoff、durable audit store 或 result materialization 已启用。

### Stop lines

UI 只展示：

- `allStopLinesEnforced`
- `blockedCapabilityIds`
- `canExecuteActions=false`
- `canUseDurableStore=false`
- `canWriteBusinessTruth=false`
- `canReplayAutomatically=false`

停止线必须被展示为不可用能力或当前阻塞能力，不得被转成隐藏的执行按钮、重试执行按钮、写回按钮或 replay 控件。

### Refresh / error state

本地 console 可维护最近一次成功加载的 overview，并在 refresh 或连接失败时继续展示上一份只读视图。UI 只允许展示：

- 当前 load status：`idle / loading / ready / error`
- 当前请求的 `endpoint`
- 最近一次成功加载时间
- refresh 期间的 `showing last overview` 状态
- 连接失败后的 `showing last overview` 状态
- 面向开发者的诊断项，例如服务未启动、URL 不匹配、CORS / preflight 或 overview contract mismatch

连接失败诊断只能帮助开发者恢复本地只读连接；不得触发自动重试执行、工具 action、业务写回或 replay。

### Dev Diagnostics / Connection Surface

本地 console 可额外展示只读 Dev Diagnostics，用于把本地开发连接状态和常见排障路径放在同一视图中。UI 只允许展示：

- 当前 `Platform URL`
- 当前 overview endpoint
- 当前 load status
- 最近一次成功加载时间
- service status
- console connection 状态
- 本地 probe 命令：`pwsh ./scripts/run-radishmind-console-dev.ps1 -VerifyOnly` 与 `./scripts/run-radishmind-console-dev.sh --verify-only`
- 本地失败分类：端口冲突、CORS / preflight、浏览器 unsafe port、overview contract mismatch
- 可选 local-smoke readiness 链接或摘要：`/v1/platform/local-smoke`、默认端口 `7000/4000`、local CORS origin 和停止线状态

该区域只是本地开发排障面，不是 production ops supervisor、process supervisor、provider health dashboard、正式部署环境隔离或自动恢复控制面。不得在该区域添加执行按钮、确认按钮、写回按钮、replay 控件、后台 supervisor 状态或 durable store 状态。即使 local-smoke 显示 `local_console_ready=true`，也只能解释为本地只读 console 链路可读。

### Audit boundary

UI 可展示 overview 中的 `audit.notes`，并把 `writes_business_truth=false` 显示为 advisory-only 状态。该区域只用于解释当前平台输出边界，不得变成确认按钮、执行按钮、写回入口或长期审计存储入口。

## 当前停止线

- `apps/radishmind-console/` 只能作为只读消费壳，不能承载执行器、store、confirmation 或 replay 逻辑。
- 不把 overview view model 当作业务状态源。
- 不把 `selectableModelIds` 解释成 provider health 或 credential readiness。
- 不把 `blockedActionRoute` 解释成可执行工具入口。
- 不在 UI 层模拟 executor、durable store、confirmation、replay 或业务写回。
- 不把 stop-line false 值隐藏成“尚未加载”或“稍后自动启用”状态。
- 不把 refresh/error 诊断解释成自动恢复执行链路。

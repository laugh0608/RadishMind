# Platform Overview UI View 契约

更新时间：2026-05-17

## 文档目的

本文档固定未来本地 console 或上层 UI 消费 `P3 Local Product Shell / Ops Surface` overview 时的首版视图边界。当前目标是说明 UI 应如何展示平台状态、model/profile inventory、session/tooling 产品面和停止线，而不是创建前端工程、接入真实 executor、接入 durable store、实现 confirmation flow 或启用 replay。

TypeScript 消费类型真相源为 `contracts/typescript/platform-overview-api.ts`。开发者可用 `scripts/run-platform-overview-consumer-smoke.py --check` 生成离线消费视图；本地平台服务启动后可加 `--base-url http://127.0.0.1:8080` 请求真实 `GET /v1/platform/overview`。

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

## 当前停止线

- 不创建 `React + Vite` 工程。
- 不把 overview view model 当作业务状态源。
- 不把 `selectableModelIds` 解释成 provider health 或 credential readiness。
- 不把 `blockedActionRoute` 解释成可执行工具入口。
- 不在 UI 层模拟 executor、durable store、confirmation、replay 或业务写回。
- 不把 stop-line false 值隐藏成“尚未加载”或“稍后自动启用”状态。

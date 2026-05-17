# Session / Tooling UI View 契约

更新时间：2026-05-17

## 文档目的

本文档固定未来平台控制台或上层 UI 消费 `P2 Session & Tooling Foundation` metadata shell 时的首版视图边界。当前目标是定义 UI 应如何展示 session 状态、tool registry 和 blocked action，而不是创建前端工程、接入真实 executor、接入 durable store 或实现确认流。

TypeScript 消费类型真相源为 `contracts/typescript/session-tooling-api.ts`。开发者可用 `scripts/run-platform-session-tooling-consumer-smoke.py --check` 生成离线消费视图；本地平台服务启动后可加 `--base-url http://127.0.0.1:8080` 请求真实 API surface。

## 视图模型

### Session status

来源：`GET /v1/session/metadata`

UI 只展示：

- `stateScope`：当前固定为 `northbound_metadata`
- `metadataOnly`：固定为 `true`
- `disabledCapabilities`：列出 `durable_session_store`、`durable_checkpoint_store`、`long_term_memory`、`automatic_replay`、`business_truth_write`
- `durableSessionStoreEnabled=false`
- `longTermMemoryEnabled=false`
- `automaticReplayEnabled=false`

该视图只说明平台能暴露 session metadata，不代表存在 durable session store、长期记忆或 replay API。

### Tool registry

来源：`GET /v1/tools/metadata`

UI 只展示：

- `registryId`
- `executionEnabled=false`
- `networkDefault=disabled`
- 工具列表中的 `toolId`、`displayName`、`projectScope`、`riskLevel`
- `requiresConfirmation`
- `canRequestAction=true`
- `executionMode=contract_only`
- 单个工具的 `executionEnabled=false`

`canRequestAction=true` 只表示 UI 可以提交一次会被平台阻断的 action 请求，用于获得统一 blocked response；它不表示工具可执行。

### Blocked action banner

来源：`POST /v1/tools/actions`

UI 只展示：

- `visible=true`
- `tone=blocked`
- `canExecute=false`
- `primaryCode`
- `requiresConfirmation`
- `noSideEffects=true`

该 banner 必须表达“动作已被平台阻断”，不得提供执行按钮、重试执行按钮或业务写回按钮。后续如果需要确认流，也必须先接入上层 confirmation handoff、独立审计、result materialization policy 和完整负向回归。

## 当前停止线

- 不创建 `React + Vite` 工程。
- 不把 view model 当作业务状态源。
- 不把 `canRequestAction` 解释成 `canExecute`。
- 不在 UI 层模拟 executor、durable store、confirmation、replay 或业务写回。
- 不把 blocked action response 转成 confirmed action。

# Platform Local Smoke UI View 契约

更新时间：2026-05-22

## 文档目的

本文档固定本地 console、开发者脚本或未来上层 UI 消费 `GET /v1/platform/local-smoke` 时的只读视图边界。当前 `apps/radishmind-console/` 已在 Dev Diagnostics 和 `Local Readiness` 面板中消费该摘要。

`local-smoke` 是 `P3 Local Product Shell / Ops Surface` 的本地开发 readiness 摘要，不是生产健康面，也不是 supervisor。它用于回答“当前默认本地 console 链路是否可读、失败时先查哪里”，而不是声明平台可以执行工具、写业务真相源或自动恢复。

TypeScript 消费类型真相源为 `contracts/typescript/platform-local-smoke-api.ts`。开发者可用以下命令生成离线视图：

```bash
./scripts/run-python.sh scripts/run-platform-local-smoke.py --check
```

本地平台服务启动后可请求真实 API：

```bash
./scripts/run-python.sh scripts/run-platform-local-smoke.py \
  --base-url http://127.0.0.1:7000 \
  --check
```

## 输入来源

来源：`GET /v1/platform/local-smoke`

该 route 汇总已有只读面：

- `/healthz` 的可读状态
- `/v1/platform/overview` 的 contract kind 和 UI consumable 状态
- `/v1/models` 与 `/v1/models/{id}` 的 provider/profile inventory 可读状态
- `/v1/session/metadata` 与 `/v1/tools/metadata` 的 metadata-only 状态
- `/v1/tools/actions` 的 blocked action no-side-effects 摘要
- 本地 console CORS origin：`http://127.0.0.1:4000` 与 `http://localhost:4000`
- 默认本地端口：platform `7000`，console `4000`
- 当前 executor、durable store、confirmation、materialized result reader、long-term memory、business writeback 和 replay 停止线

`local-smoke` 只复用既有 metadata / blocked shell，不启动进程、不守护进程、不访问外部 provider，也不替代 production health dashboard。

## 视图模型

### Summary

UI 或脚本只展示：

- `status`
- `localConsoleReady`
- `readOnly=true`
- `platformPort=7000`
- `consolePort=4000`

`localConsoleReady=true` 只表示本地只读 console 所需的基础 route 和 contract 摘要可读。它不表示 production deployment ready、process supervisor ready、production secret backend ready、provider health ready 或 executor ready。

### Route Readiness

UI 或脚本只展示：

- `healthzOk`
- `overviewContractReadable`
- `modelInventoryReadable`
- `sessionToolingMetadataReadable`
- `blockedActionNoSideEffects`

这些字段只用于排查本地 console 读取链路。若任一字段为 `false`，调用侧应提示开发者检查本地服务、bridge inventory 或 metadata route，而不是尝试自动执行 action、写入状态或 replay。

### Local Console

UI 或脚本只展示：

- `allowedCorsOrigins`
- `corsScope=local_dev_only`
- `defaultFrontendOrigin=http://127.0.0.1:4000`
- `defaultBackendUrl=http://127.0.0.1:7000`
- `unsafePortHintPresent`

本地 CORS 只服务开发期 console；不得把这些 origin 当作 production CORS policy、正式鉴权或外部公开部署策略。

### Failure Hints

UI 可把 `failure_hints` 映射为开发者排障提示：

- `PORT_IN_USE`：默认 `7000/4000` 被占用或现有进程不是预期服务
- `CORS_ORIGIN_NOT_ALLOWED`：浏览器 origin 不在本地 console allowlist 中
- `ERR_UNSAFE_PORT`：浏览器在请求发出前拦截 unsafe port

这些提示只解释本地读取失败，不触发自动修复、后台重启、进程杀停或系统配置修改。

### Console Failure Surface

本地 console 同时读取 `GET /v1/platform/overview` 与 `GET /v1/platform/local-smoke`。如果 overview 已经可读，但 local-smoke 请求失败或 contract mismatch，UI 只能显示 `Local-smoke readiness unavailable`、`failureSurface=platform_local_smoke` 和 local-smoke 专属诊断；如已有上一份成功加载的只读视图，可继续展示该 stale view。

该失败态不表示 production incident、process supervisor 失效或 executor 状态变化，也不得触发自动重启、执行按钮、确认按钮、durable store 控件、业务写回或 replay 控件。

### Stop Lines

UI 或脚本只展示：

- `allStopLinesEnforced`
- `canExecuteActions=false`
- `canUseDurableStore=false`
- `canWriteBusinessTruth=false`
- `canReplayAutomatically=false`

停止线必须展示为不可用能力或当前阻塞能力。不得把 stop-line false 值隐藏成“稍后自动启用”，也不得把 local-smoke 视图转成执行按钮、确认按钮、写回按钮或 replay 控件。

## 与 Overview 的关系

`GET /v1/platform/overview` 是产品展示入口，适合 console 渲染 service status、model inventory、session/tooling surface、stop-lines 和 audit boundary。

`GET /v1/platform/local-smoke` 是本地开发 readiness 摘要，适合 Dev Diagnostics、脚本 smoke 或排障页展示当前本地链路是否可读。

两者都只读、无副作用。overview 不能被解释为业务真相源；local-smoke 不能被解释为 production health 或 supervisor。overview 可读但 local-smoke 失败时，UI 仍只进入只读诊断态，不升级为执行链路或生产运维状态。

## 当前停止线

- 不把 `localConsoleReady=true` 写成 production ready。
- 不把 `healthzOk=true` 写成 provider health ready。
- 不把 `modelInventoryReadable=true` 写成 credential ready 或外部 provider 可调用。
- 不把 blocked action no-side-effects 写成确认流已接通。
- 不在 UI 层模拟 executor、durable store、confirmation、replay 或业务写回。
- 不把 failure hints 变成自动杀进程、自动改端口或自动重启服务。

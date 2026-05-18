# RadishMind Console

本目录是 `P3 Local Product Shell / Ops Surface` 的最小本地 console 壳。

它只读取 `GET /v1/platform/overview`，并复用 `contracts/typescript/platform-overview-api.ts` 中的 `PlatformOverviewResponse` 与 `toPlatformOverviewConsoleViewModel`。当前页面只展示 service status、model/profile inventory、session/tooling blocked 状态和 stop-lines。

## 本地运行

先启动平台服务：

```powershell
pwsh ../../scripts/run-platform-service.ps1 serve
```

再启动 console：

```powershell
npm install
npm run dev
```

默认读取 `http://127.0.0.1:8080/v1/platform/overview`。如需改地址，可设置：

```powershell
$env:VITE_RADISHMIND_PLATFORM_BASE_URL="http://127.0.0.1:8080"
```

平台服务只允许 `http://127.0.0.1:5173` 与 `http://localhost:5173` 这两个本地 console origin 读取 API；这只是本地开发 CORS 边界，不代表生产鉴权或公开部署已完成。

## 停止线

- 不实现真实工具执行器
- 不接 durable session/checkpoint/audit/result store
- 不接 confirmation flow
- 不读取 materialized result
- 不写上层业务真相源
- 不启用 automatic replay

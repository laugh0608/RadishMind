# RadishMind Console

本目录是 `P3 Local Product Shell / Ops Surface` 的最小本地 console 壳。

它当前读取 `GET /v1/platform/overview` 与 `GET /v1/platform/local-smoke`，并复用 `contracts/typescript/platform-overview-api.ts` 中的 `PlatformOverviewResponse` / `toPlatformOverviewConsoleViewModel` 以及 `contracts/typescript/platform-local-smoke-api.ts` 中的 `PlatformLocalSmokeResponse` / `toPlatformLocalSmokeReadinessViewModel`。当前页面只展示 service status、model/profile inventory、session/tooling blocked 状态、`Local Readiness` 面板、stop-lines、audit boundary、refresh 状态、Dev Diagnostics、failure surface 和连接失败诊断。

`GET /v1/platform/local-smoke` 只作为本地开发 readiness 摘要，用于排查 healthz、overview contract、model inventory、session/tooling metadata、blocked action no-side-effects、CORS 和停止线状态。它不是 production health、process supervisor、真实 executor 或 durable store。

## 本地运行

在本目录安装依赖：

```powershell
npm install
```

先从本目录启动平台服务：

```powershell
pwsh ../../scripts/run-platform-service.ps1 serve
```

再启动 console：

```powershell
npm run dev
```

默认读取 `http://127.0.0.1:7000/v1/platform/overview` 与 `http://127.0.0.1:7000/v1/platform/local-smoke`。如需改地址，可设置：

```powershell
$env:VITE_RADISHMIND_PLATFORM_BASE_URL="http://127.0.0.1:7000"
```

平台服务只允许 `http://127.0.0.1:4000` 与 `http://localhost:4000` 这两个本地 console origin 读取 API；这只是本地开发 CORS 边界，不代表生产鉴权或公开部署已完成。

也可以从仓库根目录使用本地开发一键入口启动并验证后端与前端：

```powershell
pwsh ./scripts/run-radishmind-console-dev.ps1
```

Linux / WSL 使用：

```bash
./scripts/run-radishmind-console-dev.sh
```

该入口按既有 `scripts/run-platform-service.ps1` / `scripts/run-platform-service.sh` 启动或复用 platform 后端，按 `npm run dev` 启动或复用本目录前端，然后探测 `http://127.0.0.1:7000/healthz`、`http://127.0.0.1:7000/v1/platform/overview`、`http://127.0.0.1:7000/v1/platform/local-smoke`、本地 console CORS preflight 和 `http://127.0.0.1:4000`。本地 readiness 摘要仍可用 `python scripts/run-platform-local-smoke.py --base-url http://127.0.0.1:7000 --check` 做脚本验证。如只验证已有进程，可执行：

```powershell
pwsh ./scripts/run-radishmind-console-dev.ps1 -VerifyOnly
```

或：

```bash
./scripts/run-radishmind-console-dev.sh --verify-only
```

如需做一次启动后自动清理的本地验证，可执行 `pwsh ./scripts/run-radishmind-console-dev.ps1 -ExitAfterProbe` 或 `./scripts/run-radishmind-console-dev.sh --exit-after-probe`。

常见失败边界：端口冲突时先确认 `7000` 和 `4000` 是否被其他进程占用；CORS / preflight 失败时确认浏览器 origin 是 `http://127.0.0.1:4000` 或 `http://localhost:4000`；浏览器报 `ERR_UNSAFE_PORT` / `unsafe port` 时改回默认 `4000/7000`。该脚本只是本地 dev wrapper，不是 production supervisor，不实现真实 executor、durable store、confirmation、业务写回或 replay。

## Production packaging 边界

当前 console production packaging 仍未完成，本目录仍是本地只读产品壳，不是 production package。当前边界：

- `package.json` 必须保持 `private=true`
- 不生成、提交或发布 production package
- 不添加 deploy / publish / release 脚本
- 不提交 `dist/` 或 `node_modules/`
- 不声明 production secret backend、正式鉴权、进程守护或环境隔离已完成

## 连接失败诊断

页面会在 refresh 期间保留上一份已加载 overview 和 local-smoke readiness；如果连接失败，会继续展示上一份只读视图并显示诊断项。`Dev Diagnostics` 区域会展示当前 `Platform URL`、overview endpoint、local-smoke endpoint、load status、failure surface、最近加载时间、service status、console connection、readiness status、`ps1` / `sh` 本地 probe 命令，以及端口冲突、CORS / preflight、unsafe port、overview contract mismatch 和 local-smoke contract mismatch 的本地排障分类。若 overview 可读但 local-smoke readiness 或 contract 失败，页面只显示 `Local-smoke readiness unavailable` 和 local-smoke 专属诊断，不升级为 production incident、supervisor 或执行链路状态。它只是本地连接排障面，不是 production ops supervisor。

常见处理顺序：

1. 确认平台服务已通过 `pwsh ../../scripts/run-platform-service.ps1 serve` 启动。
2. 打开 `http://127.0.0.1:7000/v1/platform/overview`，确认返回 JSON。
3. 打开 `http://127.0.0.1:7000/v1/platform/local-smoke`，确认 readiness 摘要返回 `platform_local_smoke`。
4. 若服务端口或主机不同，更新页面里的 `Platform URL` 或设置 `VITE_RADISHMIND_PLATFORM_BASE_URL`。
5. 若浏览器报 CORS / preflight，确认 console 使用 `http://127.0.0.1:4000` 或 `http://localhost:4000`。
6. 若提示 overview contract 不匹配，运行 `python ../../scripts/run-platform-overview-consumer-smoke.py --base-url http://127.0.0.1:7000 --check`。
7. 若提示 local-smoke contract 不匹配，运行 `python ../../scripts/run-platform-local-smoke.py --base-url http://127.0.0.1:7000 --check`。

## 验证

```powershell
npm run build
python ../../scripts/check-radishmind-console-shell.py
python ../../scripts/check-radishmind-console-behavior.py
python ../../scripts/check-platform-local-smoke-contract.py
python ../../scripts/check-radishmind-console-production-boundary.py
python ../../scripts/check-p3-local-product-shell-short-close-checklist.py
```

## 停止线

- 不实现真实工具执行器
- 不接 durable session/checkpoint/audit/result store
- 不接 confirmation flow
- 不读取 materialized result
- 不写上层业务真相源
- 不启用 automatic replay

# RadishMind Console

本目录是 `P3 Local Product Shell / Ops Surface` 的最小本地 console 壳。

它当前读取 `GET /v1/platform/overview` 与 `GET /v1/platform/local-smoke`，并复用 `contracts/typescript/platform-overview-api.ts` 中的 `PlatformOverviewResponse` / `toPlatformOverviewConsoleViewModel` 以及 `contracts/typescript/platform-local-smoke-api.ts` 中的 `PlatformLocalSmokeResponse` / `toPlatformLocalSmokeReadinessViewModel`。当前页面只展示 service status、model/profile inventory、`Provider/Profile Details` 只读详情、session/tooling blocked 状态、`Local Readiness` 面板、stop-lines、`Stop-line Details` 只读详情、audit boundary、refresh 状态、Dev Diagnostics、failure surface 和连接失败诊断。

`GET /v1/platform/local-smoke` 只作为本地开发 readiness 摘要，用于排查 healthz、overview contract、model inventory、session/tooling metadata、blocked action no-side-effects、CORS 和停止线状态。它不是 production health、process supervisor、真实 executor 或 durable store。

当前样式已开始按 [RadishMind UI 设计规范](../../docs/radishmind-ui-design-spec.md) 收敛到 `--rm-*` 语义 token。后续视觉调整应优先复用这些 token，不在页面样式中继续散落硬编码颜色；token 收敛不等于正式 ops surface UI 定稿。

## 界面结构

当前 console 是 `P3` 本地只读 ops surface，不是营销页、管理后台或 production console。页面固定为三层信息结构：

1. 左侧导航栏
   - 展示 `RadishMind Console` 产品身份。
   - 提供 Overview、Local Readiness、Provider/Profile、Session & Tooling、Stop-line Details 的锚点导航。
   - 显示当前加载状态和 `read-only shell` 边界。
2. 主工作区
   - `Runtime overview` 标题区提供 `Platform URL` 输入和 refresh 按钮。
   - 全局状态条展示 overview endpoint、local-smoke endpoint、最近加载时间和 stale 状态。
   - 摘要指标卡展示 Service、Boundary、Readiness 和 Stop lines。
   - 主列依次展示 Service Status、Model Inventory、Session And Tooling 和 Dev Diagnostics。
3. 右侧辅助栏
   - `Local Readiness` 展示 healthz、overview contract、model inventory、session/tooling metadata、blocked action no-side-effects、CORS origin 和 failure hints。
   - `Stop Lines` 与 `Stop-line Details` 展示当前被阻止的能力、原因和 evidence。
   - `Audit Boundary` 展示 advisory-only 和 writes_business_truth 边界。

窄屏下页面改为单列顺序展示，而不是压缩桌面三栏。长 endpoint、provider id、profile id 和诊断命令允许换行，不应横向撑破页面。

## 当前面板说明

- `Service Status`：只读展示 service name、version、stage、mode 和 overview route。
- `Model Inventory`：只读展示 model / provider / profile 计数、默认 provider/profile/model、selectable model ids、active profile chain 和 Provider/Profile Details。
- `Provider/Profile Details`：只解析 overview 已返回的 model id 来源，不额外请求详情 route，不声明 credential readiness。
- `Session And Tooling`：只读展示 session/tooling metadata route 和 blocked action route；Blocked Action Detail 明确不渲染 execute / confirm / apply / replay 控件。
- `Dev Diagnostics`：展示本地排障字段、verify-only 命令和端口 / CORS / unsafe port / contract mismatch 分类。
- `Local Readiness`：展示 `GET /v1/platform/local-smoke` 生成的本地 readiness 摘要，不表示 production ready。
- `Stop Lines`：展示 executor、durable store、confirmation、materialized result reader、long-term memory、business truth write、automatic replay 和 production secret backend 等停止线。

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
- `npm run build` 只做本地或 CI smoke：执行 `tsc --noEmit && vite build`，生成的 `dist/` 仍不得提交
- `npm run preview` 只做本地 build preview：固定 `127.0.0.1:4000`，不是 production hosting
- 不生成、提交或发布 production package
- 不添加 deploy / publish / release 脚本
- 不提交 `dist/` 或 `node_modules/`
- 不声明 production secret backend、正式鉴权、进程守护或环境隔离已完成

`scripts/checks/fixtures/production-ops-console-package-smoke.json` 与 `scripts/check-production-ops-console-package-smoke.py` 固定上述 package smoke 边界。该检查不发布、不上传、不启动公开服务，也不把 Vite preview 解释为 production deployment。

## Docker local compose 边界

`apps/radishmind-console/Dockerfile` 和 `apps/radishmind-console/nginx.local.conf` 只服务 `docker-local-compose` 本地容器 smoke。它们通过本地 nginx 容器托管 build 后的只读 console，默认把 `VITE_RADISHMIND_PLATFORM_BASE_URL` 固定到 `http://127.0.0.1:7000`，与 `deploy/docker-compose.local.yaml` 中发布到宿主机的 platform 端口配套。

这不是 production hosting policy，也不是 console production package ready。测试 / 生产部署态、runtime config、正式 auth / CORS policy、镜像发布和外部反代仍待后续 `docker-test-prod-compose` 切片定义。

## 连接失败诊断

页面会在 refresh 期间保留上一份已加载 overview 和 local-smoke readiness；如果连接失败，会继续展示上一份只读视图并显示诊断项。`Dev Diagnostics` 区域会展示当前 `Platform URL`、overview endpoint、local-smoke endpoint、load status、failure surface、最近加载时间、service status、console connection、readiness status、`ps1` / `sh` 本地 probe 命令，以及端口冲突、CORS / preflight、unsafe port、overview contract mismatch 和 local-smoke contract mismatch 的本地排障分类。若 overview 可读但 local-smoke readiness 或 contract 失败，页面只显示 `Local-smoke readiness unavailable` 和 local-smoke 专属诊断，不升级为 production incident、supervisor 或执行链路状态。它只是本地连接排障面，不是 production ops supervisor。

`Stop-line Details` 只解释每个 blocked capability 为什么保持禁用，以及 overview、local-smoke 和 audit boundary 中哪些只读证据支撑该结论。它不提供 action button，不调用 `POST /v1/tools/actions`，也不把 executor、durable store、confirmation、业务写回或 replay 标成 ready。

`Provider/Profile Details` 只解释 overview 已返回的 `selectable_model_ids`、默认 provider/profile/model、inventory kind、`/v1/models` 和 `/v1/models/{id}` route。它不会额外请求模型详情 route，也不会把 `canShowProfileSelector=true` 解释成 provider health check、credential readiness、production secret backend 或真实外部调用策略已经完成。

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

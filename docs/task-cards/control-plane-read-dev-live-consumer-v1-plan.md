# Control Plane Read Dev Live Consumer v1 计划

## 目标

`control-plane-read-dev-live-consumer-v1` 在 `apps/radishmind-web/` 中建立 dev-only live read consumer 路径，使现有七个只读页面可以在显式 opt-in 时通过 HTTP 消费 fake-store-backed read handlers 与测试身份上下文。

默认路径仍是离线 fixture / view model，继续作为可验证基线。

## 范围

- 前端默认使用 `offline_fixture`。
- 当 `VITE_RADISHMIND_READ_SOURCE=dev-live-http` 时，前端进入 `dev_live_http` 模式，从 `VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL` 读取七条 read route。
- live 请求只发送 test-only dev fake auth header，不发送真实 credential。
- 后端只有在 `RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1` 时才接受 dev fake auth header 并注入测试身份上下文。
- live route 仍消费现有 in-memory fixture fake store。
- 页面 view model 可接受 live collection override，但仍保留离线 fixture fallback。

## 程序化证据

- `scripts/checks/fixtures/control-plane-read-dev-live-consumer-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-dev-live-consumer-v1.py`
- `apps/radishmind-web/src/features/control-plane-read/devLiveReadConsumer.ts`
- `services/platform/internal/httpapi/control_plane_read.go`
- `services/platform/internal/httpapi/control_plane_read_test.go`

## 验证

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-dev-live-consumer-v1.py
cd apps/radishmind-web && npm run build
./scripts/run-python.sh scripts/run-control-plane-read-consumer-smoke.py --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不接真实数据库。
- 不接 `Radish` OIDC。
- 不实现 repository migration。
- 不实现 API key lifecycle、quota enforcement、billing 或 cost ledger。
- 不实现 workflow executor、confirmation、writeback 或 replay。
- 不声明 production API consumer、production admin console 或完整 formal user workspace ready。

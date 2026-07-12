# Gateway Bridge stdio Worker Pool v1 任务卡

更新时间：2026-07-11

## 任务标识

- 任务 ID：`gateway-bridge-stdio-worker-pool-v1`
- 状态：`completed`
- 对应功能专题：`docs/features/gateway/python-bridge-runtime-v1.md`
- 基线证据：`docs/features/gateway/evidence/process-per-request-baseline-2026-07-11.json`

## 用户结果

内部开发者通过现有 Gateway northbound API 发起请求时，平台复用一组受控 Python worker，不再为每个请求重复启动 Python 与 import 模块；请求、credential、stream event、失败和取消仍保持明确隔离。

## 实现范围

1. 在 `scripts/run-platform-bridge.py` 增加版本化 `worker` 模式：启动后先输出 ready handshake，再以 JSONL 读取 request frame、输出 result / stream event / error frame。
2. worker 启动环境必须移除请求 credential；credential 只存在于单个 request frame 和本次 `GatewayOptions` 生命周期，不缓存、不写 stdout / stderr、不进入错误或观测摘要。
3. 在现有 Go `bridge.Client` 后增加受控 `stdio_pool` 模式，同时保留 `process_per_request` 显式回滚模式；不改变 HTTP northbound contract、canonical request 或 provider registry。
4. pool 固定 worker 上限和有界等待容量。每个 worker 同时只处理一个 unary 或 stream 请求，stream event 不跨请求交错。
5. request timeout、context cancellation、worker EOF、协议错帧或回调失败会终止当前 worker；当前请求不自动 retry，后续请求使用重建后的健康 worker。
6. 平台 shutdown 先停止接收新请求，再关闭 worker stdin 并等待退出；超出优雅退出预算时才终止进程，不遗留后台 worker。
7. 扩展既有 bridge benchmark，在同一 fixture、mock provider、请求数和并发度下比较 `process_per_request` 与 `stdio_pool`。

## Worker protocol v1

- ready frame 必须包含 `protocol_version=1`、`type=ready` 和受支持 operation；Go 在 worker 进入 idle pool 前完成握手校验。
- request frame 必须包含内部 request ID、operation、canonical request 和受控 options；provider registry / inventory 请求不携带 credential。
- unary 请求以单个 result 或 error 结束；stream 请求可以输出多个 stream event，但最终必须以 result 结束。
- response frame 的 request ID 必须与当前在途请求一致；错 ID、未知 frame type、非法 JSON、EOF 或超限 frame 都视为 worker protocol failure。
- worker protocol error 只返回稳定 code 与脱敏 message，不回显请求正文、credential、provider 原始响应、stderr 或本机路径。

## 配置与回滚

- 新增 `bridge_mode`、`bridge_worker_count`、`bridge_queue_capacity` 和 `bridge_handshake_timeout` 的 config file / env / sanitized summary 口径。
- 推荐默认值为 `stdio_pool / 4 / 64 / 5s`；只有本批正确性、race、恢复和性能证据全部通过后，才把 `stdio_pool` 固定为仓库默认模式。
- `process_per_request` 长期保留为显式诊断与回滚模式；未知 mode、非法 worker 数量、非法 queue 容量和非法 handshake timeout 必须 fail closed。

## 稳定失败语义

至少覆盖以下内部错误码，并由 northbound error envelope 保留稳定映射：

- `BRIDGE_WORKER_QUEUE_FULL`
- `BRIDGE_WORKER_TIMEOUT`
- `BRIDGE_WORKER_CANCELED`
- `BRIDGE_WORKER_EXITED`
- `BRIDGE_WORKER_PROTOCOL_ERROR`
- `BRIDGE_WORKER_UNAVAILABLE`
- `BRIDGE_CLIENT_CLOSED`

错误正文不得包含 stderr、request payload、credential、base URL 或本机绝对路径。

## 必要测试与实测

- Python worker protocol unittest：ready、unary、stream、未知 operation、credential 不跨请求保留、错误脱敏。
- Go pool behavior tests：并发隔离、有界排队、queue full、timeout / cancellation、worker crash、错帧、后续请求重建、Close 和 stale credential environment 清除。
- `go test ./...`、`go test -race ./...`、`go vet ./...`。
- Gateway Python unittest、Gateway service smoke、平台 northbound HTTP tests。
- process 与 pool 各完成 2 次预热、20 次顺序请求、20 次四并发请求；Go route benchmark 保留为非主要成本参照。
- `./scripts/check-repo.sh --fast`；本批改变默认运行时、配置与执行边界，完成时补跑全量 `./scripts/check-repo.sh`。

## 完成条件

- 新旧模式输出等价的 Gateway envelope / stream event 语义，provider registry 与 inventory 一致。
- 并发请求没有 credential、canonical request、response 或 stream event 串线。
- worker 崩溃、超时、取消和协议错误都有稳定 code；失败请求不自动重放，下一请求可通过重建 worker 恢复。
- pool 关闭后没有 Python worker 残留。
- 相对已提交 process baseline，顺序与并发 bridge 自身 p95 开销均至少下降 70%；否则不切换默认模式，并在证据中记录阻塞原因。

## 停止线

- 不接真实 provider，不读取或写入真实 API key。
- 不实现自动 retry、provider fallback、load balancing、动态扩缩容或跨实例 worker 调度。
- 不新增网络端口、内部 HTTP Python 服务、第二套 provider registry 或第二套 Gateway contract。
- 不启用 production secret、quota、billing、公开生产 API 或 production deployment 声明。
- 不新增 readiness / refresh checker 链；标准单元测试、行为测试、benchmark、Gateway smoke 与仓库门禁承载本批验证。

## 2026-07-11 完成记录

- Python worker protocol v1 已覆盖 ready handshake、unary / stream / providers / inventory、request ID 关联、8 MiB frame 上限和脱敏 error frame。
- Go `stdio_pool` 已覆盖默认 `4 / 64 / 5s` 配置、有界 admission、单 worker 单在途请求、timeout / cancellation、crash / protocol failure 后重建、并行优雅 close 和稳定 northbound 错误码。
- credential 只通过 request frame 进入当次 `GatewayOptions`；Python 测试证明下一请求不会继承且处理结束会释放 frame 引用，Go 测试证明 worker 环境会移除陈旧 credential、写入可随 context 取消且 JSON / 传输缓冲会在写后清空。
- back-to-back mock 对照中，顺序 / 四并发 bridge 自身 p95 分别从 `62.796 / 75.439 ms` 降到 `4.098 / 4.252 ms`，降幅 `93.5% / 94.4%`；吞吐提升 `128.0% / 152.2%`。
- `stdio_pool` 已成为仓库默认模式，`process_per_request` 继续作为显式回滚模式；完整参数与限制见 [对照证据](../features/gateway/evidence/stdio-worker-pool-comparison-2026-07-11.json)。
- 本批没有调用真实 provider、读取真实 API key、增加网络端口、启用自动 retry/fallback 或扩大 production 声明。

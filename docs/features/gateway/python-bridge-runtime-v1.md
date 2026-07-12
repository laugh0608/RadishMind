# Gateway Python Bridge Runtime v1

更新时间：2026-07-11

状态：`gateway_bridge_stdio_worker_pool_completed`

## 用户结果

内部开发者通过现有 northbound API 调用 mock 或受控 provider 时，不应为每个请求重复承担完整 Python 进程启动与模块 import 成本；优化后的 bridge 必须保持现有协议、错误边界和请求级 credential 隔离。

## 当前实现

`services/platform/internal/bridge/client.go` 的 `HandleEnvelope`、`StreamEnvelope`、`DescribeProviders` 和 `DescribeInventory` 已默认通过四个受控 `stdio` worker 复用 Python runtime；`process_per_request` 保留为显式诊断与回滚模式。

- worker 启动完成 protocol v1 ready handshake 后才进入 idle pool；每个 worker 同时只承载一个 unary 或 stream 请求。
- pool 默认 worker / queue / handshake 配置为 `4 / 64 / 5s`；queue、timeout、cancellation、EOF、错帧与 close 均有明确行为。
- credential 不进入 worker 环境，只存在于单个 request frame 和本次 provider 调用生命周期；下一请求不会继承上一请求 credential。
- 当前请求在 timeout、取消、worker crash 或协议错误后不会自动 retry；后续请求通过重建后的 worker 恢复。
- mock provider 基线只用于定位本地运行时成本，不能外推真实 provider SLA。

## 基线与选型范围

1. 在 Gateway envelope 内新增受控 `provider_duration_ms` 观测字段；`duration_ms` 继续表示 Python Gateway 总耗时，两者都不包含 secret 或 provider 原始响应。
2. 新增稳定的 bridge benchmark 命令，直接复用真实 Go `bridge.Client` 与 mock provider，输出顺序 / 并发 total、process / IPC、Python Gateway、provider 的统计摘要。
3. 新增 northbound Go route benchmark，以内存 fake bridge 排除 Python 与 provider 成本。
4. 运行本机基线并提交脱敏 run record；记录运行时版本和参数，不记录用户名、绝对路径、环境变量内容或机器标识。
5. 比较候选形态并选择下一实现批次；该批选定受控 `stdio` worker pool，后续实现结果见本页下方。

## 计时模型

| 分段 | 计算方式 | 说明 |
| --- | --- | --- |
| Go route | `httpapi` benchmark | request decode、selection、canonical request、response translation |
| total bridge | Go 调用 `bridge.Client.HandleEnvelope` 的 wall time | 当前包含进程启动、IPC、Python 与 mock provider |
| process / IPC | `total bridge - metadata.duration_ms` | 当前主要是进程启动、import、stdin/stdout 与 Go decode |
| Python Gateway | `metadata.duration_ms - metadata.provider_duration_ms` | schema validation、route、envelope 与 response validation |
| provider | `metadata.provider_duration_ms` | `run_inference` 调用段；mock 结果不能外推真实 provider SLA |

所有差值小于零时归零。毫秒统计使用 nearest-rank p50 / p95，并同时记录平均值、最小值、最大值和 wall-clock throughput。

## 2026-07-11 实测结果

完整脱敏记录见 [process-per-request baseline](evidence/process-per-request-baseline-2026-07-11.json)。本次使用 repository `.venv`、mock provider、2 次预热、20 次顺序请求和 20 次四并发请求：

| 场景 | total p50 / p95 | process / IPC p50 / p95 | Python Gateway p50 / p95 | provider p50 / p95 | 吞吐 |
| --- | --- | --- | --- | --- | --- |
| 顺序 | 114.455 / 120.712 ms | 70.548 / 73.712 ms | 22 / 24 ms | 22 / 23 ms | 8.555 req/s |
| 四并发 | 164.121 / 273.888 ms | 100.104 / 175.854 ms | 33 / 51 ms | 32 / 47 ms | 20.289 req/s |

Go northbound route 五轮 benchmark 的 p50 / p95 分别为 `0.011772 ms` / `0.013559 ms`。按 `process / IPC ÷ (process / IPC + Python Gateway)` 计算，进程与 IPC 在顺序和并发场景分别占 bridge 自身 p95 开销的 `75.4%` 与 `77.5%`；它是 R4 的首要优化对象，并为新实现达到 p95 至少下降 70% 留出了足够空间。

## stdio worker pool 实测结果

完整对照见 [stdio worker pool comparison](evidence/stdio-worker-pool-comparison-2026-07-11.json)。同一 fixture、mock provider、2 次预热、20 次顺序请求和 20 次四并发请求的 back-to-back 对照如下：

| 场景 | process total p95 | pool total p95 | process bridge overhead p95 | pool bridge overhead p95 | 自身开销降幅 | pool 吞吐 |
| --- | --- | --- | --- | --- | --- | --- |
| 顺序 | 107.646 ms | 48.265 ms | 62.796 ms | 4.098 ms | 93.5% | 22.331 req/s |
| 四并发 | 120.339 ms | 47.843 ms | 75.439 ms | 4.252 ms | 94.4% | 85.710 req/s |

pool 只启动 4 个 Python worker，process 模式为同批请求启动 42 个进程；顺序 / 并发总 p95 分别下降 `55.2% / 60.2%`，吞吐分别提升 `128.0% / 152.2%`。相对最初 committed baseline 的 bridge 自身 p95 降幅也达到 `94.4% / 97.6%`，默认模式切换条件已经满足。

## 候选形态

| 候选 | 优点 | 主要代价 | 本批判断 |
| --- | --- | --- | --- |
| 受控 stdio worker pool | 不新增端口；一个 worker 一次只处理一个请求，容易隔离 stream 与 credential；可按 worker 粒度取消 / 重启 | 需要 worker 生命周期、池并发和协议握手 | 已实现并成为默认模式 |
| 单 worker 多路复用 stdio | 进程最少，理论吞吐高 | 必须实现请求关联、并发写、stream 交错、单请求取消，隔离复杂 | 本阶段不选 |
| 内部 HTTP Python 服务 | HTTP 工具链成熟，可自然并发 | 新增端口、服务发现、部署、鉴权和第二运行单元 | 当前阶段不选 |
| 保持 process-per-request | 隔离直接，回滚简单 | 重复启动与 import，吞吐和尾延迟不可控 | 保留为基线与回滚模式 |

## 实现完成边界

2026-07-11 [Gateway Bridge stdio Worker Pool v1 任务卡](../../task-cards/gateway-bridge-stdio-worker-pool-v1-plan.md) 已完成：

- protocol v1 handshake、受控 worker 上限、有界排队、并行优雅 shutdown 和超时终止均已实现。
- unary / stream 单 worker 独占、request ID 关联和 concurrent isolation 行为测试已通过。
- timeout、cancellation、worker crash、错帧、queue full 和 client close 均映射稳定错误码；失败请求不自动重试。
- 请求级 credential 不进入 worker 环境、不缓存、不记录、不回显；Go 侧分别清空 JSON 与传输缓冲，Python 侧在单次 frame 作用域结束后释放引用，阻塞写入可被 context 取消；process 回滚模式也不再回显 stderr 原文。
- process / pool 同口径 benchmark、Python unittest、Go race、Gateway smoke 与仓库门禁共同承载验证，不新增 checker 链。

## 验收

- provider timing schema 正向 / 负向验证通过。
- benchmark 统计函数有确定性单元测试，非法参数和失败请求明确退出。
- 顺序与并发基线均完成，run record 包含 p50 / p95、吞吐、请求数和进程启动数。
- Go route benchmark 至少运行五轮，记录稳定量级，不用单次最优值做结论。
- `go test ./...`、相关 Python / Gateway smoke、`go test -race ./...` 与仓库快速门禁通过。

## 停止线

- 不继续扩动态 worker scaling、内部 HTTP bridge、自动 retry 或 fallback。
- 不更改 OpenAI-compatible、Responses、Messages、Models 的公开 request / response 语义。
- 不运行真实 provider，不读取或写入真实 API key。
- 不把 benchmark 结果写入产品 API，也不承诺 production SLA。
- 不新增同层 checker；现有 schema validation、Go tests、Gateway smoke 和仓库基线负责验证。

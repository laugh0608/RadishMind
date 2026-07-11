# Model Gateway / API Distribution 设计与开发文档

更新时间：2026-07-11

## 功能定位

`Model Gateway / API Distribution` 负责对外提供 OpenAI-compatible、Responses、Messages、Models 等 northbound API，并统一分发到多 provider、多 profile 和多模型。Gateway 只负责协议适配、路由、运行时治理与观测，不成为 provider 配置或上层业务数据的第二真相源。

## 当前状态

- 平台已有 `/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models` 和 `/v1/models/{id}` 的第一版 bridge-backed 兼容面。
- `apps/radishmind-web/` 已有 Model Gateway Overview、Route Evidence、Usage/Audit Evidence 和 Evidence Review / Readiness。
- provider capability、health smoke、selection policy、retry/fallback policy 和 runtime docs 已进入仓库快速门禁。
- Go Gateway 当前每次 bridge 操作都会启动独立 Python 子进程；凭证不进入 argv。2026-07-11 的 mock 基线已确认进程启动 / IPC 约占 bridge 自身 p95 开销的八成。
- 当前不执行真实 API key 生命周期、quota enforcement、rate limit、billing、cost ledger、provider retry/fallback execution、production gateway 或 load balancing。

## 当前开发目标

R4 第一批 [Gateway Python Bridge Runtime v1](gateway/python-bridge-runtime-v1.md) 已完成 mock provider 的顺序 / 并发基线、四段成本拆分和候选评审，唯一推荐形态为受控 `stdio` worker pool。

下一批应在现有 `bridgeClient` 后实现有界 worker pool、版本化握手、排队、超时 / 取消、崩溃后重建、优雅退出和请求级 credential / stream 隔离。它不改变 northbound request / response 语义，不接真实 provider，不启用生产 secret、自动 retry/fallback 或新的公开 API。

## 设计边界

- gateway 只按 canonical contract 与 provider/profile metadata 分发，不把任一 provider 写成唯一方向。
- capability 不等于 health，health smoke 不等于 production readiness。
- 默认 retry policy 为 caller-managed，fallback policy 为 disabled；任何自动 fallback 都需要独立设计和审计。
- key、quota、billing 和 cost ledger 必须有明确失败语义和审计记录，不能只做 UI 展示。
- bridge 优化必须保留现有 `bridgeClient` 边界和 mock / offline 路径；不能通过绕过 canonical request、schema validation 或 provider registry 换取性能。
- request credential 只能通过受控请求级通道进入 Python，不得出现在 argv、公开错误、日志、benchmark 结果或 committed run record。

## 推进顺序

1. 已建立 process-per-request 的可复现顺序 / 并发基线，并记录 p50、p95、吞吐和进程启动次数。
2. 已用同一请求和 mock provider 分离 Go 路由、子进程启动 / IPC、Python Gateway 与 provider 路径耗时。
3. 已比较受控 stdio worker pool、单 worker 多路复用与内部 HTTP 服务，选定受控 `stdio` worker pool。
4. 下一批独立实现选定方案，覆盖健康握手、并发上限、排队、超时 / 取消、崩溃恢复、优雅退出和 credential 隔离。
5. 只有新实现相对基线的 bridge 自身 p95 开销下降至少 70%，才允许切换默认运行模式。

## 验收方式

- Go route：标准 Go benchmark，使用内存 fake bridge，报告 `ns/op`、`B/op` 与 `allocs/op`。
- bridge：真实 `bridge.Client` + mock provider，报告顺序 / 并发 total、process / IPC、Python Gateway、provider 的 p50 / p95。
- correctness：Gateway schema / smoke、Go 单元测试、并发 race、凭证负向检查与仓库快速门禁。
- 阶段收口：功能专题、任务卡、run record、当前焦点和周志一致；真实测量产物不包含本机绝对路径、环境变量值或 secret。

## 停止线

- 不在 worker 正确性、race、崩溃恢复和性能证据形成前切换默认运行模式；process-per-request 继续作为显式回滚模式。
- 不新增第二套 northbound contract、provider registry、selection policy 或 Gateway 业务真相源。
- 不把 mock provider 性能解释为真实 provider SLA。
- 不在本批启用 production API key、quota、billing、自动 fallback、load balancing 或 production deployment。
- 不为基线与选型新增 readiness / refresh checker 链；现有单元测试、benchmark、Gateway smoke 和仓库门禁足以承载。

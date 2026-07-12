# Model Gateway / API Distribution 设计与开发文档

更新时间：2026-07-12

## 功能定位

`Model Gateway / API Distribution` 负责对外提供 OpenAI-compatible、Responses、Messages、Models 等 northbound API，并统一分发到多 provider、多 profile 和多模型。Gateway 只负责协议适配、路由、运行时治理与观测，不成为 provider 配置或上层业务数据的第二真相源。

## 当前状态

- 平台已有 `/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models` 和 `/v1/models/{id}` 的第一版 bridge-backed 兼容面。
- `apps/radishmind-web/` 已有 Model Gateway Overview、Route Evidence、Usage/Audit Evidence 和 Evidence Review / Readiness。
- provider capability、health smoke、selection policy、retry/fallback policy 和 runtime docs 已进入仓库快速门禁。
- Go Gateway 已默认使用受控 `stdio` worker pool，复用四个 Python worker；`process_per_request` 仅保留为显式回滚模式，凭证不进入 argv 或 worker 环境。
- [Model Gateway Request History / Usage & Failure Review v1](gateway/model-gateway-request-history-usage-failure-review-v1.md) 已完成 `memory_dev`、PostgreSQL dev/test、分页详情、重启恢复和完整失败 / 取消终态证据。
- 当前不执行真实 API key 生命周期、quota enforcement、rate limit、billing、cost ledger、provider retry/fallback execution、production gateway 或 load balancing。

## 当前开发目标

R4 第一批 [Gateway Python Bridge Runtime v1](gateway/python-bridge-runtime-v1.md) 已完成 mock provider 的顺序 / 并发基线、四段成本拆分和候选评审，唯一推荐形态为受控 `stdio` worker pool。

[Gateway Bridge stdio Worker Pool v1 任务卡](../task-cards/gateway-bridge-stdio-worker-pool-v1-plan.md) 已在现有 `bridgeClient` 后完成有界 worker pool、版本化握手、排队、超时 / 取消、崩溃后重建、优雅退出和请求级 credential / stream 隔离。它没有改变 northbound request / response 语义，没有接真实 provider，也没有启用生产 secret、自动 retry/fallback 或新的公开 API。

Workflow 产品链、Gateway Request History 与 [Gateway Playground / Request Review Loop v1](gateway/gateway-playground-request-review-loop-v1.md) 均已关闭。内部开发者现在可以从 Web 真实调用三个现有 northbound 协议、取消 stream、查看当前响应，并按同一 request id 进入 sanitized history detail。

该功能只增加 Web consumer / lazy panel 与 request-id handoff，复用现有 API、dev/test caller scope 和 history，不新增 schema、repository、provider contract 或生产授权。输入输出只存在于当前组件内存，Request History 继续只保存 sanitized operational metadata。

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
4. 已实现健康握手、并发上限、排队、超时 / 取消、崩溃恢复、优雅退出和 credential 隔离。
5. 新实现相对 back-to-back process 基线的顺序 / 并发 bridge 自身 p95 开销下降 `93.5% / 94.4%`，已切换默认模式。
6. Request History 与 Playground v1 已完成三协议 request → response → history 交互闭环；下一步重新比较四个一级产品面的未完成用户价值。

## 验收方式

- Go route：标准 Go benchmark，使用内存 fake bridge，报告 `ns/op`、`B/op` 与 `allocs/op`。
- bridge：真实 `bridge.Client` + mock provider，报告顺序 / 并发 total、process / IPC、Python Gateway、provider 的 p50 / p95。
- correctness：Gateway schema / smoke、Go 单元测试、并发 race、凭证负向检查与仓库快速门禁。
- 阶段收口：功能专题、任务卡、run record、当前焦点和周志一致；真实测量产物不包含本机绝对路径、环境变量值或 secret。

## 停止线

- process-per-request 继续作为显式回滚模式；不移除该路径，也不把 worker pool 扩为动态集群调度。
- 不新增第二套 northbound contract、provider registry、selection policy 或 Gateway 业务真相源。
- 不把 mock provider 性能解释为真实 provider SLA。
- 不在本批启用 production API key、quota、billing、自动 fallback、load balancing 或 production deployment。
- 不为基线与选型新增 readiness / refresh checker 链；现有单元测试、benchmark、Gateway smoke 和仓库门禁足以承载。
- Playground 与 Request History 只服务开发 / 测试交互和审查，不等于 production API key、quota enforcement、billing、cost ledger、自动 retry/fallback、load balancing 或 production gateway ready。

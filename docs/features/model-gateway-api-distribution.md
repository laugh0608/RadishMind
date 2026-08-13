# Model Gateway / API Distribution 设计与开发文档

更新时间：2026-08-13

## 功能定位

`Model Gateway / API Distribution` 负责对外提供 OpenAI-compatible、Responses、Messages、Models 等 northbound API，并统一分发到多 provider、多 profile 和多模型。Gateway 只负责协议适配、路由、运行时治理与观测，不成为 provider 配置或上层业务数据的第二真相源。

## 当前状态

- 平台已有 `/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models` 和 `/v1/models/{id}` 的第一版 bridge-backed 兼容面。
- `apps/radishmind-web/` 已有 Model Gateway Overview、Route Evidence、Usage/Audit Evidence 和 Evidence Review / Readiness。
- provider capability、health smoke、selection policy、retry/fallback policy 和 runtime docs 已进入仓库快速门禁。
- Go Gateway 已默认使用受控 `stdio` worker pool，复用四个 Python worker；`process_per_request` 仅保留为显式回滚模式，凭证不进入 argv 或 worker 环境。
- [Model Gateway Request History / Usage & Failure Review v1](gateway/model-gateway-request-history-usage-failure-review-v1.md) 已完成 `memory_dev`、SQLite、PostgreSQL dev/test、分页详情、重启恢复和完整失败 / 取消终态证据；[Provider 上报用量规范化与应用用量审查 v1](gateway/provider-reported-usage-normalization-application-review-dev-test-v1.md)进一步完成五类 usage、三协议投影与应用当前窗口审查。
- [User Workspace Application API Integration & Invocation v1](user-workspace/application-api-integration-invocation-v1.md) 已复用 `/v1/models`、Playground 与 History，让当前选中 application 完成模型发现、接入示例、dev/test 调用和同 request id 审查；没有扩 Gateway API 或 schema。
- [Application Configuration Draft & Review v1](user-workspace/application-configuration-draft-review-v1.md) 已把经过模型 / 协议校验的 application draft 配置交给既有 Integration / Playground；Gateway 仍只消费 application / protocol / model，不读取草案描述、不保存测试输入输出，也没有新增 northbound schema。
- [Application Publish Governance & Promotion v1](user-workspace/application-publish-governance-promotion-v1.md) 只把 sanitized Gateway `request_id` 作为 candidate evidence reference，并复用既有 Integration / Playground / History handoff；Gateway 不读取 candidate、review 或 eligibility，也没有新增协议、schema、provider registry 或发布职责。
- [用户工作区 API 密钥生命周期与 Gateway 开发测试态认证 v1](user-workspace/api-key-lifecycle-gateway-dev-test-auth-v1.md) 已完成密钥领域、管理 API、五条 northbound 路由的显式 `api_key_dev_test` 认证、可信调用上下文、脱敏请求历史、最近使用更新、聚合 SQLite 本地产品链、真实 PostgreSQL 专项门禁、Web 一次性交接与浏览器连续验收，专题关闭；聚合 runtime 现已随 Admin Provider / Route 与 application request quota 扩展为十一组件。
- [Admin Provider Profile / Model Route 受控启用（开发 / 测试态）v1](admin-control-plane/provider-profile-model-route-controlled-activation-dev-test-v1.md) 批次 A 至 E 已完成并关闭，建立配置领域、三模式 durable repository、人工 review、显式 activation、可恢复 active snapshot、verified Admin API、Admin Web 和只读 Gateway consumer。`static_config` 继续作为默认模式；显式 `admin_snapshot_dev_test` 模式按租户、工作区、环境、配置、protocol 与 model 精确选路，固定请求开始时的 generation / digest，inventory 或 route 漂移在 bridge 前失败且不回退；Request History 页面展示精确快照谱系。
- 当前已执行开发测试态 application request quota admission；仍不执行生产 API 密钥生命周期、production quota、rate limit、billing、cost ledger、provider retry/fallback execution、production gateway 或 load balancing。
- [Provider 价格策略版本与应用成本审查（开发 / 测试态）v1](gateway/provider-pricing-policy-version-application-cost-review-dev-test-v1.md) 已完成设计、后端、Visual R1、React strict consumer、双数据库与真实浏览器连续链：价格保持独立版本化 owner，请求只绑定精确 selection 对应的不可变 USD 快照，成本只由合法 reported usage 以整数算法估算并进入 Request History / Application Operations；不改变请求准入、路由或 Provider 调用结果。
- [Gateway Provider Attempt 受控重试与降级执行（开发 / 测试态）v1](gateway/provider-attempt-controlled-retry-fallback-execution-dev-test-v1.md) 已获批准并完成批次 A：Route v2、可序列化冻结 attempt plan、Python adapter → bridge 类型化失败、Request History v3 memory owner 与原子 checkpoint 已成立。下一步按唯一任务卡推进 SQLite / PostgreSQL；现有 `retry_policy=caller-managed`、`fallback_policy=disabled` 仍是 northbound 运行事实，HTTP、Web、真实 Provider 调用和 production capability 尚未打开。

## 当前开发目标

R4 第一批 [Gateway Python Bridge Runtime v1](gateway/python-bridge-runtime-v1.md) 已完成 mock provider 的顺序 / 并发基线、四段成本拆分和候选评审，唯一推荐形态为受控 `stdio` worker pool。

[Gateway Bridge stdio Worker Pool v1 任务卡](../task-cards/gateway-bridge-stdio-worker-pool-v1-plan.md) 已在现有 `bridgeClient` 后完成有界 worker pool、版本化握手、排队、超时 / 取消、崩溃后重建、优雅退出和请求级 credential / stream 隔离。它没有改变 northbound request / response 语义，没有接真实 provider，也没有启用生产 secret、自动 retry/fallback 或新的公开 API。

Workflow 产品链、Gateway Request History、[Gateway Playground / Request Review Loop v1](gateway/gateway-playground-request-review-loop-v1.md) 与 application-scoped API Integration 均已关闭。内部开发者现在可以从选中 application 读取模型目录、生成三协议接入示例、调用三个现有 northbound 协议、取消 stream、查看当前响应，并按同一 request id 与 application scope 进入 sanitized history detail。

该功能只增加 Web consumer / lazy panel 与 request-id handoff，复用现有 API、dev/test caller scope 和 history，不新增 schema、repository、provider contract 或生产授权。输入输出只存在于当前组件内存，Request History 继续只保存 sanitized operational metadata。

Provider Attempt 批次 B 已把 Route v2 与 Request History v3 同构落入 SQLite / PostgreSQL，并证明旧版本兼容、重启恢复、并发单赢家、损坏 payload 失败关闭、runtime role 无 DDL 与 no fallback。当前开发目标已切换为批次 C：只扩既有 Route draft / candidate / review / activation API 与激活链；northbound fallback、Web 与 production capability 继续关闭。

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
6. Request History、Playground、Application API Integration、Application Configuration Draft / Review 与 Publish Governance 已完成 application → validated configuration → models / examples → request → response → history → immutable candidate / review 的开发测试路径。
7. API 密钥 Gateway 认证、本地连续链、PostgreSQL migration / 角色 / 方言 / 并发门禁、Web 一次性交接和浏览器重启复验均已通过；Provider reported usage 已进入 canonical envelope、三协议、历史与应用审查。独立 application request quota 已完成三模式 owner、Admin API、六条 route provider 前准入、S9 完整 Pencil、React 严格 consumer、CAS 确认与真实浏览器连续链；不提前打开 production distribution、token 估算、价格、production quota 或计费。
8. 价格与成本专题设计、S9 / S10 Visual R3、S7 Pricing / S5 Cost Review Visual R1、领域、三模式 owner、Admin API、Request History v2、React strict consumer、数据库实例和真实浏览器连续链均已关闭。
9. Provider Attempt 批次 A、B 已关闭；当前按唯一高风险任务卡推进批次 C 的既有 Admin Route API 与激活链，不从静态 retry / fallback policy 派生同层 checker 链。

## 验收方式

- Go route：标准 Go benchmark，使用内存 fake bridge，报告 `ns/op`、`B/op` 与 `allocs/op`。
- bridge：真实 `bridge.Client` + mock provider，报告顺序 / 并发 total、process / IPC、Python Gateway、provider 的 p50 / p95。
- correctness：Gateway schema / smoke、Go 单元测试、并发 race、凭证负向检查与仓库快速门禁。
- 阶段收口：功能专题、任务卡、run record、当前焦点和周志一致；真实测量产物不包含本机绝对路径、环境变量值或 secret。

## 停止线

- process-per-request 继续作为显式回滚模式；不移除该路径，也不把 worker pool 扩为动态集群调度。
- 不新增第二套 northbound contract、provider registry、selection policy 或 Gateway 业务真相源。
- 不把 mock provider 性能解释为真实 provider SLA。
- 批次 D 及其独立 gate 完成前不执行 fallback；后续也只允许开发测试态、请求显式允许、三个非流式 API 和最多一次不同 Profile 顺序降级，不启用 production API key、production quota、billing、load balancing 或 production deployment。
- 不为基线与选型新增 readiness / refresh checker 链；现有单元测试、benchmark、Gateway smoke 和仓库门禁足以承载。
- Playground 与 Request History 只服务开发 / 测试交互和审查，不等于 production API key、quota enforcement、billing、cost ledger、自动 retry/fallback、load balancing 或 production gateway ready。

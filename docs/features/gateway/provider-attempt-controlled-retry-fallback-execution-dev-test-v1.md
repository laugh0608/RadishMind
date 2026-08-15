# Gateway Provider Attempt 受控重试与降级执行（开发 / 测试态）v1

更新时间：2026-08-15

状态：`gateway_provider_attempt_dev_test_v1_batch_e_product_continuity_next`

## 功能定位

本专题为内部开发者预览建立一条显式、有限、可审计的 Provider 降级执行链：管理员在既有 Provider Profile / Model Route 原子配置中维护主目标与最多一个备用目标，候选经过人工审查并显式激活后，调用者还必须在单次非流式请求中明确允许使用该策略；只有主目标返回服务端判定为可降级的类型化失败时，Gateway 才能在同一不可变 Route snapshot 下发起一次备用 Provider attempt。

v1 的“重试”只表示同一根请求内的第二次受控 Provider attempt，不允许对同一个 Provider Profile 原地重试。平台现有 `retry_policy=caller-managed` 对普通请求继续成立；没有显式 Route policy 与请求级允许时，`fallback_policy=disabled` 继续成立。

本专题解决的真实阻塞是：Provider Route、逐 attempt quota、请求级价格快照和 Request History 已分别成立，但当前 Route 只能选择一个 Profile；主 Provider 返回临时失败时，即使管理员已经维护另一个兼容 Profile，Gateway 也只能结束根请求。直接增加循环会破坏四项现有事实：

1. quota 使用 request id 作为唯一 attempt identity，第二次调用会与首次准入冲突；
2. pricing 只为单一 selection 保存一个请求级快照；
3. Request History v2 只保存一个 Provider / Profile / Model 和一个 cost estimate；
4. 当前 `GATEWAY_INFERENCE_FAILED` 不能证明失败是否适合降级，Go 层不得解析错误文案猜测。

因此，本专题先建立版本化 attempt plan、类型化 Provider failure、逐 attempt quota / pricing 和 Request History v3，再接真实执行；不把自动切换藏进现有 bridge wrapper。

## 用户任务

### 平台管理员

1. 在 development / test 的既有 Provider Route 草案中，为一个 `protocol + model_id` 维护有序 attempt targets。
2. 首个 target 是主目标，第二个 target 是唯一备用目标；两者必须引用不同 `provider_profile_id`。
3. 选择固定的 `execution_mode=single_attempt | sequential_fallback`；v1 不允许自定义失败码列表。
4. 把完整配置生成不可变候选，审查主备能力、额外 quota / 成本风险和失败边界，再显式激活新 generation。
5. 从激活快照与 Request History 审查真实 attempt lineage；回滚继续生成新的 generation。

### 内部应用开发者

1. 使用 active application 的开发测试态 API Key 调用三个标准 northbound unary API。
2. 在 `radishmind` 扩展中显式设置 `fallback_mode=allow_configured`；省略或设置 `disabled` 时只执行主目标。
3. 主目标失败后，只有服务端类型化分类允许时才进入备用目标；调用者不能临时指定 Provider、Profile、失败码或备用顺序。
4. 从响应头和同一根 `request_id` 的 Request History 查看是否使用备用目标、实际 attempt 数、每次 quota admission、价格 / usage 覆盖与终态。

## 范围与启用前置

v1 只覆盖：

- `POST /v1/chat/completions`，且 `stream=false`；
- `POST /v1/responses`，且 `stream=false`；
- `POST /v1/messages`，且 `stream=false`；
- `api_key_dev_test` 可信调用上下文；
- `admin_snapshot_dev_test` Provider Route source；
- `development | test` 单一进程环境。

启用时必须同时满足：

- API Key 开发测试态认证已启用；
- application request quota enforcement 已启用且 policy 可用；
- Gateway Request History v3 store 已启用；
- Provider Route active snapshot 为 v2；
- pricing capture 已启用，以便每个实际 attempt 都留下价格覆盖证据；
- 独立 `RADISHMIND_GATEWAY_PROVIDER_FALLBACK_DEV=1` gate 已开启。

任一前置缺失都在主 Provider side effect 前失败关闭，不回退 single-attempt 静态配置、旧 snapshot、memory store 或未受控 bridge 调用。

## 领域所有权

| 领域 | 唯一 owner | 本专题行为 |
| --- | --- | --- |
| Provider inventory | 既有 runtime inventory | 继续拥有 provider、profile、capability、resolved model、enabled 与 digest |
| Route 草案 / 候选 / 快照 | 既有 Admin Provider Route owner | 以 v2 表达有序 attempt targets 和执行模式，不创建第二套路由 owner |
| request attempt plan | Gateway request-local owner | 从同一 active snapshot 编译不可变 plan，不持久化为独立配置资源 |
| quota policy / usage | 既有 Gateway Request Quota owner | 每个实际 Provider attempt 使用独立 attempt id 原子准入，不预留、不退款 |
| pricing policy | 既有 Gateway Pricing owner | 在首个 Provider side effect 前为全部 plan targets 固定 request-local snapshot |
| Provider 调用 | 既有 bridge / provider adapter | 每个 target 仍只调用一次；返回类型化、脱敏的 attempt failure observation |
| Request History | 既有 Gateway Request History owner | 升级 v3，在同一根记录中保存最多两个 attempt 与成本覆盖汇总 |

## Provider Route v2

既有 v1 草案、候选、快照和 activation history 保持可读、可回滚，并按单目标投影；不原地改写历史。

`admin_provider_route_configuration_draft.v2` 的每条 Model Route 使用：

```json
{
  "route_id": "route_chat_default",
  "protocol": "chat_completions",
  "model_id": "platform-model",
  "execution_mode": "sequential_fallback",
  "attempt_targets": [
    {"ordinal": 1, "provider_profile_id": "profile_primary"},
    {"ordinal": 2, "provider_profile_id": "profile_backup"}
  ]
}
```

固定约束：

- `single_attempt` 只允许一个 target；`sequential_fallback` 必须恰好两个 target；
- ordinal 只能是连续的 `1..2`，数组顺序必须与 ordinal 一致；
- 两个 `provider_profile_id` 必须不同，不允许同 Profile 原地重试；
- 两个 target 都必须支持同一 protocol capability，并在 candidate 创建、activation 与请求开始时重验 inventory binding / digest；
- 任一 target 缺失、禁用、重复、环境不匹配、capability 不兼容或 digest 漂移时，整个 plan 在首次 Provider side effect 前失败；不能静默降级为单目标执行；
- v2 candidate / snapshot digest 覆盖执行模式、target 顺序、两项 inventory binding 与 resolved model；
- approval 仍不改变 Gateway 行为，activation 才切换后续请求；在途请求始终固定开始时的 generation 与 snapshot digest。

v1 snapshot 只能编译 `single_attempt` plan。旧快照不能因新 gate 开启而获得备用目标。

## 不可变 attempt plan

Gateway 在请求开始时一次性读取 active snapshot，编译 `gateway_provider_attempt_plan.v1`：

- 根 `request_id`、route、protocol、requested model；
- configuration id、generation、snapshot digest；
- `fallback_mode` 与调用者是否显式允许；
- 有序 targets，每项保存 ordinal、provider、profile、selected model、upstream model、inventory digest；
- 每项 request-local pricing snapshot；
- `max_attempts=1|2`。

全部 target 必须在主 attempt 前解析完成。执行中不重新读取 active Route、inventory 或 pricing current pointer，也不因新 generation、价格更新或 health 变化改写 plan。

attempt id 由服务端确定性生成：

```text
<root_request_id>.pa1
<root_request_id>.pa2
```

它只用于 quota admission、attempt checkpoint 与内部 bridge correlation；northbound 根 `request_id` 不变。客户端不能提交 attempt id。

## 类型化 Provider failure

Python Provider adapter 与 bridge 必须输出内部 `gateway_provider_attempt_failure.v1`，至少包含：

- `failure_class`；
- `fallback_disposition=eligible | ineligible`；
- `provider_response_started`；
- `outcome=failed | unknown`；
- 脱敏稳定 code 与可选 HTTP status class；
- 不含 endpoint、credential、请求正文、响应正文或原始异常。

Go 只能消费该结构，禁止根据错误消息、HTTP message 或 Provider 名称推断是否可降级。结构缺失、非法、未知或与 bridge 状态矛盾时统一视为 `ineligible`。

### v1 可降级失败

只有以下类型化分类允许进入备用 target：

| failure class | 条件 | 说明 |
| --- | --- | --- |
| `provider_rate_limited` | Provider 明确返回受限终态，未产生成功响应 | 只换到不同 Profile，不重试原目标 |
| `provider_temporarily_unavailable` | Provider 明确返回临时不可用终态 | 只接受 adapter 允许列表映射 |
| `provider_upstream_gateway_unavailable` | Provider 明确返回上游 `502 / 503 / 504` 类终态 | 不按错误正文判断 |

### 永不降级

- northbound JSON、schema、model、protocol 或 canonical request 错误；
- API Key、application、workspace、permission 或 membership 失败；
- Route snapshot / inventory / digest / override 失败；
- quota missing、exceeded、conflict 或 store failure；
- pricing / Request History store failure；
- Provider auth、credential、endpoint 配置或 capability 失败；
- safety、content policy、invalid request、model not found 或 unsupported；
- client cancellation、context deadline、bridge queue full、worker timeout / crash / protocol、client closed；
- `outcome=unknown`、类型化结构缺失或 Provider response 已开始；
- platform response translation、tool、confirmation、Workflow、business writeback 或 replay 失败。

v1 特意不把 `BRIDGE_WORKER_TIMEOUT`、`PLATFORM_BRIDGE_FAILED` 或笼统 `GATEWAY_INFERENCE_FAILED` 设为可降级。这些错误不能证明换 Provider 能解决问题，也不能证明主 Provider outcome。

## 执行语义

1. 完成认证、application / scope、Route v2 plan 编译和全部 target 校验。
2. 创建 Request History v3 根记录，并保存完整 plan 摘要。
3. 为全部 plan targets 固定价格快照；价格缺失或不可用不改变路由，但必须进入对应 attempt evidence。
4. 使用 `.pa1` 执行第一次 quota admission；准入成功后 checkpoint attempt 1 为 `running`。
5. 调用主 target 一次。成功则结束；失败则写入类型化 observation 与 attempt 1 终态。
6. 仅当 Route mode、调用者允许、失败分类、未取消状态和剩余 target 全部满足时，先 durable checkpoint `fallback_pending`。
7. 使用 `.pa2` 独立执行 quota admission；拒绝时不调用备用 Provider，根请求以精确 quota failure 结束。
8. 准入成功后 checkpoint attempt 2 为 `running`，再调用备用 target 一次。
9. 任一 attempt 成功则根请求成功；两次均失败时以第二次稳定失败作为根终态，同时保留第一次失败。

每个实际 attempt 消耗一次 quota。Provider 失败、history terminal update 失败或客户端随后断开都不退款。没有执行的备用 target 不消耗 quota。

Request History 在主失败与备用调用之间的 checkpoint 是安全前置；该 checkpoint 失败时不得执行备用 target。备用成功后的最终 history update 失败不能改写已经得到的 Provider 成功响应，但最后 durable checkpoint 必须保持 `running | outcome_unknown`，不能伪装成完整成功。

## Request History v3

`gateway_request_record.v3` 继续使用既有 store 与根 request repository，不创建第二套历史 owner。v1 / v2 历史保持可读并投影为单 attempt。

v3 在根记录新增：

- `attempt_count`、`fallback_allowed`、`fallback_used`；
- `terminal_attempt_id`；
- 最多两个 `gateway_provider_attempt_record.v1`；
- `gateway_request_attempt_cost_summary.v1`。

每个 attempt record 保存：

- attempt id、ordinal 与状态；
- Provider / Profile / selected / upstream model；
- Route generation / digest 与 inventory digest；
- quota admission id 或稳定拒绝 code；
- started / completed 时间和 duration；
- 类型化 failure class、disposition、outcome、failure boundary；
- usage 与既有 `gateway_request_cost_estimate.v1`。

根记录既有 `selected_provider / selected_profile / selected_model` 对 v3 固定表示主 selection，避免静默改变历史字段语义；terminal selection 只从 `terminal_attempt_id` 对应记录读取。

成本汇总只累加已有 `estimated` evidence，并同时返回：

- `known_cost_micros`；
- `coverage=complete | partial | none`；
- `estimated_attempt_count`、`unknown_attempt_count`。

`partial` 不能冒充完整请求成本。Application Operations 只汇总当前已加载窗口中的 known cost，并单独展示 partial coverage；不推测失败 attempt 的 token 或账单。

现有 Provider / Profile / Model filter 对 v3 继续匹配主 selection；新增 `fallback_used` 与 terminal Provider / Profile filter。分页 cursor 必须绑定新过滤条件，v1 / v2 统一投影 `fallback_used=false`。

## Northbound 与响应边界

三个请求中的 `radishmind` 扩展新增：

```json
{"fallback_mode": "disabled | allow_configured"}
```

- 默认与省略值均为 `disabled`；
- 调用者只能允许或拒绝 active policy，不能传 target、失败码、attempt 数、Provider 或 Profile；
- Route 为 `single_attempt` 时传 `allow_configured` 仍只执行一次；
- streaming 请求携带 `allow_configured` 时以稳定 `GATEWAY_PROVIDER_FALLBACK_STREAM_UNSUPPORTED` 在 Provider 前拒绝，不静默忽略。

响应新增脱敏头：

- `X-RadishMind-Provider-Attempts: 1|2`；
- `X-RadishMind-Fallback-Used: true|false`。

错误 envelope metadata 可返回 attempt count、fallback used、terminal Provider / Profile 与 Route generation / digest，但不得返回原始 Provider 错误、endpoint 或 credential。成功响应正文保持现有 OpenAI-compatible / Responses / Messages 合同，不嵌入 RadishMind 私有 attempt 结构。

## 权限、持久化与运行配置

- Route v2 继续复用 `admin_provider_routes:read | draft | review | activate`，不新增隐式权限；
- fallback 执行只消费 API Key 既有 route scope，不增加客户端管理权限；
- Route、Request History、quota 与 pricing 的 memory / SQLite / PostgreSQL 三模式继续互斥且 no fallback；
- SQLite / PostgreSQL migration 需要支持 Route v2 与 Request History v3，runtime role 仍无 DDL；
- active v1 snapshot、历史 request v1 / v2 和旧 activation history 不迁移成伪 v2 / v3 事实；
- 新 gate 默认关闭，且不能由聚合 SQLite mode 自动打开。

## Web 与 Pencil

五维评分为 `1 / 2 / 2 / 2 / 2 = 9`，采用 `A / 完整 Pencil`，但不建立 S11。设计只在既有 S7 Route 与 S5 Request History 页面族增加代表面：

- S7 Route：主备 target 顺序、执行模式、能力兼容、额外 quota / 成本风险、candidate diff、review 与 activation；
- S5 Playground：请求级 `Allow configured fallback` 显式开关和非流式限制；
- S5 Request History：根请求、连续 attempt rows、主失败、备用准入、terminal target、quota 与成本 coverage；
- Desktop、关键断点、`390×844` 的 context → plan → attempts → cost / boundary 顺序；
- single attempt、fallback success、fallback rejected by quota、two failures、history checkpoint failure 和旧 v2 record 代表状态。

颜色不能成为主 / 备用、失败 / 成功或 partial coverage 的唯一通道。普通 target 保持中性；只有当前审查 attempt 使用墨蓝选中轨，失败和风险使用文字、图标与语义状态。

Pencil 通过前不实现 React；稳定同族组件可直接复用，只有上述结构性决策进入设计基准。

2026-08-13 已在同一 Family UI 设计源中完成 Visual R1，共新增七个横向相邻的顶层基准面：

- S7 Route：Desktop `h41DNz`、Narrow `Q5dMjv`；冻结有序主 / 备 target、`sequential_fallback`、能力兼容、额外 quota / partial cost 风险、candidate review 与独立 activation；
- S5 Playground：Desktop `DY5HB`、Narrow `o9Btk`；桌面表达非流式请求显式允许 `allow_configured`，窄屏表达 stream 开启时开关锁定且零隐式 fallback；
- S5 Request History：Desktop `KsXpp`、Narrow `BRzOE`；桌面表达主失败 → durable checkpoint → 备用成功，窄屏表达第二次 quota 拒绝且零备用 Provider 调用；
- Decision R17 `GfqT6`：统一冻结 `single attempt`、fallback success、quota rejected、two failures、history checkpoint failure 与旧 v2 单 attempt 投影六类代表状态。

七个根节点已完成逐节点边界扫描和实际截图复核，结果为零裁切、零越界、零 placeholder；颜色不是状态唯一通道，只有当前审查 target / attempt 使用墨蓝选中轨。Visual R1 已于 2026-08-13 获得人工视觉批准；React strict consumer 随后于 2026-08-15 完成。

## 实施批次

设计已于 2026-08-13 获得人工批准；当前由[唯一高风险任务卡](../../task-cards/gateway-provider-attempt-controlled-retry-fallback-execution-dev-test-v1-plan.md)承接以下批次，不派生同层 readiness / refresh 链：

### 批次 A：合同、Route v2 与 memory owner

- 实现 Route v2、attempt plan、类型化 Provider failure 与 Request History v3 领域合同；
- 扩展 deterministic fake bridge / Provider adapter，覆盖主成功、可降级失败、不可降级失败与 outcome unknown；
- memory repository 与领域测试先闭合，不接 HTTP / Web。

### 批次 B：SQLite / PostgreSQL

- 增加 Route v2 与 Request History v3 migration、marker、repository parity、重启和并发；
- 验证 v1 / v2 兼容、runtime DDL 拒绝、checkpoint 原子性、损坏 payload 与 no fallback。

已完成：SQLite 与 PostgreSQL 已分别推进 Route store v2 与 Request History store v3。两种数据库的 Request History 更新都在事务锁内重读当前记录并执行与 memory owner 相同的 attempt 状态迁移校验；Route v1、历史 v1 / v2、Route v2、v3 checkpoint、重启、损坏 payload、并发单赢家、runtime role 无 DDL 和 no fallback 已形成同构证据。当前运行时仍不执行 fallback。

### 批次 C：Admin API 与激活链

- 扩展既有 Route draft / candidate / review / activation API，不创建第二套 endpoint；
- 覆盖 strict JSON、权限、inventory 漂移、双标签 CAS、rollback 与 v1 snapshot。

已完成：既有 Admin Route endpoint、verified identity、开发测试态门禁和 `admin_provider_routes:read | draft | review | activate` 权限原样复用，`PUT` 草案现可严格消费 v1 或 v2，不接受混合版本。HTTP 证据覆盖有序双 target、两项 inventory digest、嵌套 strict JSON、能力不兼容、draft / review / generation CAS、v1 → v2 显式 activation 与回滚到 v1；review 不改变 active snapshot，activation 前任一 target 漂移都保持零运行时切换。真实 PostgreSQL 17 也已通过 Route v2 Admin HTTP 完整链与聚合 configured profile；northbound fallback 仍未接入。

### 批次 D：northbound unary 执行

- 建立 request-local executor，串联 per-attempt quota、pricing、bridge 和 history checkpoints；
- 接三个标准 API Key unary route；保持 stream、Application RAG、Prompt、Agent、Workflow 与 Campaign 单 attempt；
- 覆盖取消、quota second-attempt rejection、history failure、价格更新、Route generation 漂移和零隐式降级。

已完成：`/v1/chat/completions`、`/v1/responses` 与 `/v1/messages` 的非流式 API Key 路径已共享同一个 request-local executor。调用者省略或传入 `disabled` 时保持单 attempt；只有独立开发测试 gate、Route v2 `sequential_fallback` 和显式 `allow_configured` 同时成立，且主目标返回完整、合法、未开始响应的 eligible 类型化失败时，才会执行一次不同 Profile 的备用 target。executor 在首次 Provider side effect 前固定 active Route、全部 inventory binding 和逐 target pricing snapshot，并以 `.pa1` / `.pa2` 串联 quota admission、Request History v3 checkpoint 与 bridge。第二次 quota 拒绝、取消、unknown / ineligible failure、history checkpoint 失败、stream、gate 关闭和 target 漂移均保持零备用 Provider 调用；备用成功后的 terminal history 写失败不改写 Provider 成功响应，也不伪造 durable success。成功正文继续使用既有三协议合同，只增加脱敏 attempt 头；真实 Provider、非 API Key、stream fallback、其它应用运行链和 production capability 仍关闭。

### 批次 E：Pencil、React 与产品连续链

- 先完成人工设计复核，再实现 Route、Playground 与 Request History strict consumer；
- memory / SQLite 浏览器覆盖主失败 → 备用成功、双失败、quota 阻断、旧记录与服务重启；
- PostgreSQL 实例覆盖 migration、并发、checkpoint 与重连；
- 完成 Web 测试、build、Go race、三视口、隐私扫描和全量仓库门禁。

当前进度：Visual R1 七个代表面已经冻结并于 2026-08-13 获得人工视觉批准。2026-08-15 又在既有 owner 上完成三块 React strict consumer：Route editor 以严格判别联合消费 v1 / v2 并审查有序主备计划；Playground 只允许 API Key unary 显式提交 `allow_configured` 并严格核对两个脱敏响应头；Request History 严格消费 v1 / v2 / v3，展示 durable attempt lineage、逐 attempt quota / cost 与 terminal selection，旧记录不补造 v3 字段。Application Operations 只汇总当前窗口 known cost，并把 partial attempt coverage 单独展示。下一小步进入 memory / SQLite 产品连续链，不创建第二套 UI owner。

## 验收

- 未激活 v2 policy、请求未显式允许或 gate 关闭时，bridge 调用最多一次；
- 两个 target 在首次 Provider side effect 前完成相同 Route snapshot 和 inventory digest 校验；
- 主成功时备用 Provider、备用 quota admission 和备用 history attempt 均为零；
- 可降级主失败时最多执行一次不同 Profile 的备用 attempt；
- ineligible、unknown、取消、bridge / platform、stream 与 provider-response-started 失败绝不降级；
- 每个实际 attempt 使用唯一确定性 attempt id，并独立消耗 quota；
- second-attempt quota 拒绝时备用 Provider 调用为零，主 attempt 证据保留；
- 每个 target 使用请求开始时冻结的价格快照，策略随后更新不重算；
- Request History v3 能恢复完整 attempt lineage，旧 v1 / v2 保持单 attempt 兼容；
- fallback success 的 HTTP 根结果为成功，但主失败、partial cost 与 fallback used 不被隐藏；
- API、数据库、日志、浏览器和 committed fixture 不包含 token、credential、endpoint、请求正文、响应正文或 Provider raw error；
- memory、SQLite、PostgreSQL 和三协议行为一致。

## 停止线

- 不实现 production Gateway、production API Key、production quota、production price、billing 或生产 secret resolver；
- 不实现同 Profile 自动 retry、指数退避、后台重试、并行竞速、hedging、负载均衡、随机 / 权重路由、熔断或 live health routing；
- 不支持 stream fallback，也不在已发送响应字节后切换 Provider；
- 不把 bridge worker、client、schema、auth、quota、pricing、history、safety 或 platform response failure 当作 Provider fallback 信号；
- 不扩 Application RAG、Prompt、Agent、Workflow、Session 或 Evaluation Campaign 的自动降级；
- 不允许客户端覆盖 Provider / Profile、target 顺序、eligible failure、attempt id 或最大次数；
- 不把 known cost、reported usage 或 Request History 解释为账单、退款、SLA 或完整财务成本；
- 不创建第二套 Route、quota、pricing、Request History、provider inventory 或 selection owner；
- 不从本专题派生 S11 或同层 gate-only 文档。

## 当前评审结论

本设计已经确认四项 v1 决策：

1. 平台不做同目标自动 retry，只允许一次不同 Profile 的顺序降级；
2. Route activation 与请求级 `allow_configured` 必须同时成立；
3. 只覆盖 API Key 认证的三个非流式 northbound API；
4. Request History v3、逐 attempt quota 与逐 attempt pricing 是实现前置，不以单 selection 字段勉强承载多 attempt。

功能设计与批次 E Visual R1 均已于 2026-08-13 获得人工批准，唯一高风险任务卡继续承接实现。批次 A 至 D 已完成领域合同、三存储同构持久化、既有 Admin 人工激活链与三个 API Key unary API 的显式受控 fallback；七个 Visual R1 代表面和 Route / Playground / Request History React strict consumer 也已关闭。下一顺位是 memory / SQLite、PostgreSQL 与真实浏览器产品连续链；专题尚未完成。stream、非 API Key、真实 Provider 调用、其它应用运行链与 production capability 仍未打开。

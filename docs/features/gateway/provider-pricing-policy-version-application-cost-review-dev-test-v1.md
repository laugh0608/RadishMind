# Provider 价格策略版本与应用成本审查（开发 / 测试态）v1

更新时间：2026-08-12

状态：`provider_pricing_policy_version_application_cost_review_dev_test_v1_completed`

实施任务卡：[Provider 价格策略版本与应用成本审查（开发 / 测试态）v1 实施任务卡](../../task-cards/provider-pricing-policy-version-application-cost-review-dev-test-v1-plan.md)

## 功能目标

为内部开发者预览建立一条可复验但不冒充计费的成本证据链：管理员在精确 tenant、workspace、environment、provider、profile 与 model 作用域中维护版本化价格策略；Gateway 在一次请求已经选定精确 Provider 路由后绑定不可变价格快照；只有 Provider 明确上报合法 token usage 时，Request History 才按固定整数算法生成估算成本；Application Operations 只汇总当前已加载窗口中的同一份估算证据。

本专题回答以下问题：

- 当前选中的 Provider / Profile / Model 是否有适用于本环境的价格策略；
- 某次 Gateway 请求使用了哪个价格策略版本和摘要；
- Provider 明确上报的 input / output token 在该价格快照下对应多少开发测试态估算成本；
- 当前应用已加载的 Request History 窗口中，有多少记录可估算、缺少 usage、缺少价格或无法读取价格；
- 当价格策略更新后，历史请求为什么仍保持原来的估算结果。

本专题不创建账单、成本账本、发票、结算、余额、预算告警、token quota 或 production price。估算成本只是 Gateway Request History 的脱敏运行证据，不是 Provider 发票或财务凭证。

## 当前依据与真实缺口

现有能力已经形成可靠前置：

- Provider Adapter 已把五类可信 usage 规范化为 `reported | not_reported`，Platform 再次校验 input / output / total token；缺失或非法 usage 不会被估算或写成零用量。
- Gateway Request History `gateway_request_record.v1` 已保存 tenant、workspace、application、Provider / Profile / Model selection、Provider Route generation / digest、usage 与稳定失败边界，并有 memory、SQLite、PostgreSQL 开发测试态 owner。
- Admin Provider Route 已提供精确 route snapshot，但它只拥有运行路由，不应混入商业价格字段。
- Application Operations 已复用 Request History 当前分页窗口，能够展示 Provider reported usage、coverage 与 `has_more`，但明确禁止推导成本。
- 开发测试态 application request quota 已有独立 policy / usage owner，只统计 provider attempt，不读取 token 或价格。

当前缺口是没有可信价格 owner，也没有请求级不可变价格快照。仓库内旧 read-side fixture 中的 `estimated_cost`、`cost_limit` 或 readiness 文案只是离线展示证据，不能作为实现输入、fallback 或迁移来源。

## 真相源与职责边界

| 职责 | 唯一 owner | 本专题允许 | 本专题禁止 |
| --- | --- | --- | --- |
| Provider / Profile / Model 路由 | 既有 Admin Provider Route 与 Gateway selection | 读取请求开始时已冻结的精确 selection lineage | 把价格写入 route draft / candidate / snapshot，或因价格缺失改写选路 |
| 价格策略 | 新 `GatewayModelPricingRepository` | 版本化保存开发测试态 USD input / output rate 与不可变摘要 | 保存 credential、endpoint、Provider invoice、折扣合同原文或 production price |
| Provider usage | 既有 Gateway usage contract | 只消费 `availability=reported` 且校验通过的 token counts | 本地 tokenizer 估算、正文长度推算、把 `not_reported` 当作零 |
| 单请求成本证据 | 既有 Gateway Request History | 在 `gateway_request_record.v2` 中保存价格快照和确定性估算 | 建立第二套 request repository、写 cost ledger 或回改 quota |
| 应用成本审查 | 既有 Application Operations | 聚合当前已加载 Gateway 窗口并显示 coverage | 跨全部分页冒充全历史，或与 Workflow run 猜测关联 |
| 请求准入 | 既有 Gateway Request Quota | 保持 provider attempt 原子准入不变 | 用估算成本扣减 quota、拒绝请求或回退 admission |

价格 owner、Provider Route owner、quota owner 和 Request History owner 必须保持独立。价格策略缺失或 store 不可用不会阻断 Provider 请求，因为本专题只提供审查证据；但失败必须显式记录为不可估算，不能静默回退静态表、旧 fixture、当前价格或零成本。

## 价格策略领域契约

### 作用域与身份

价格策略唯一键固定为：

```text
tenant_ref + workspace_id + environment + provider_id + profile_id + model_id
```

- `environment` 只允许 `development | test`。
- `provider_id`、`profile_id` 与 `model_id` 必须精确匹配 Gateway selection，不做别名、通配、继承或 fallback。
- `policy_id` 由服务端对 canonical scope 计算稳定短 ID，客户端不能指定另一个身份。
- 同一作用域只有一个 current version；每次成功 PUT 创建新的不可变 revision，并原子切换 current pointer。
- repository 必须保留历史 revision，不能原地覆盖旧版本。

### `gateway_model_pricing_policy.v1`

最小公开字段固定为：

```json
{
  "schema_version": "gateway_model_pricing_policy.v1",
  "policy_id": "gmp_...",
  "record_version": 1,
  "tenant_ref": "tenant_demo",
  "workspace_id": "workspace_demo",
  "environment": "development",
  "provider_id": "provider_demo",
  "profile_id": "profile_demo",
  "model_id": "model_demo",
  "currency": "USD",
  "token_unit": 1000000,
  "input_price_micros_per_token_unit": 1000000,
  "output_price_micros_per_token_unit": 3000000,
  "policy_digest": "sha256:...",
  "reason": "initial development pricing evidence",
  "updated_at": "2026-08-12T00:00:00Z",
  "updated_by_actor_ref": "subject_admin",
  "request_id": "request_...",
  "audit_ref": "audit_..."
}
```

约束如下：

- v1 只允许 `currency=USD`，不声称支持汇率、税、折扣、阶梯价、缓存 token 独立价格、图像 / 音频单位或批量价。
- `token_unit` 固定为 `1_000_000`，调用方不能修改。
- input / output rate 使用非负 `int64` 微美元整数，零价是有效价格，不等同于缺少策略。
- rate、reason、scope 与版本共同进入 canonical digest；字段顺序不影响摘要。
- `reason` 必填且只用于脱敏人工说明，不允许包含 credential、endpoint、合同正文、联系人、账单或自由异常。
- 创建要求 `expected_version=0`；更新必须精确匹配 current version。并发只有一个赢家，失败返回当前版本 metadata。

## Admin API 与权限

开发测试态 API 固定为：

```text
GET /v1/admin/gateway-model-pricing-policy
PUT /v1/admin/gateway-model-pricing-policy
```

GET 以唯一的 `provider_id`、`profile_id`、`model_id` query 精确读取；tenant、workspace、environment、actor 与 permission 来自 verified identity、active workspace 和显式开发测试环境头。PUT body 只允许：

```json
{
  "expected_version": 0,
  "provider_id": "provider_demo",
  "profile_id": "profile_demo",
  "model_id": "model_demo",
  "currency": "USD",
  "input_price_micros_per_token_unit": 1000000,
  "output_price_micros_per_token_unit": 3000000,
  "reason": "initial development pricing evidence"
}
```

权限独立为：

- `admin_gateway_pricing:read`
- `admin_gateway_pricing:write`

写入前必须在 Web 展示精确作用域、旧版本、新 rate、单位、币种、reason 与“只影响后续请求价格快照”的确认。更新不会重算历史请求，不修改 Provider Route、quota、API Key、application、Workflow run 或发布状态。

稳定失败至少包括：

- `gateway_pricing_disabled`
- `gateway_pricing_scope_denied`
- `gateway_pricing_environment_forbidden`
- `gateway_pricing_payload_invalid`
- `gateway_pricing_policy_not_found`
- `gateway_pricing_policy_version_conflict`
- `gateway_pricing_policy_scope_conflict`
- `gateway_pricing_store_unavailable`

API 响应不得包含 Provider credential、endpoint、真实合同、invoice、Authorization、输入输出、raw membership、SQL、DSN 或内部异常。

## 持久化与运行配置

价格 owner 由 `RADISHMIND_GATEWAY_MODEL_PRICING_STORE=memory_dev|sqlite_dev|postgres_dev_test` 显式选择。聚合 `RADISHMIND_LOCAL_PERSISTENCE_MODE=sqlite_dev` 会把它投影为 `sqlite_dev` 并应用共享 SQLite migration，但不会自动开启 Admin 或 capture gate。

PostgreSQL 模式严格分离 runtime DSN `RADISHMIND_GATEWAY_MODEL_PRICING_DEV_TEST_DATABASE_URL` 与 migration DSN `RADISHMIND_GATEWAY_MODEL_PRICING_DEV_TEST_MIGRATION_DATABASE_URL`；DDL 只允许 migration identity 执行：

```bash
cd services/platform
go run ./cmd/radishmind-gateway-model-pricing-migrate up
go run ./cmd/radishmind-gateway-model-pricing-migrate status
```

管理 HTTP、写入、请求捕获与环境分别由 `RADISHMIND_GATEWAY_MODEL_PRICING_DEV_HTTP`、`RADISHMIND_GATEWAY_MODEL_PRICING_DEV_WRITE`、`RADISHMIND_GATEWAY_MODEL_PRICING_CAPTURE_DEV` 和 `RADISHMIND_GATEWAY_MODEL_PRICING_ENVIRONMENT=development|test` 控制。`postgres_dev_test` 未迁移、marker / checksum 漂移、runtime DDL 权限过大或连接失败时必须失败关闭，不回退 memory。

## 请求级价格快照与成本估算

### 绑定时机

Gateway 保持现有顺序：身份 / API Key → application / route 资格 → 精确 Provider selection → quota admission → bridge / Provider。价格解析只在精确 selection 已存在后执行，并必须在 Provider side effect 前固定结果；它不改变 selection 或 quota decision。

请求 trace 记录以下两类结果之一：

1. 精确找到合法策略：复制 policy identity、version、digest、currency、token unit 与 input / output rate 作为不可变 request-local snapshot。
2. 未找到、scope mismatch、store unavailable 或 contract invalid：记录稳定不可用 reason，不 fallback；请求继续按既有路径执行。

Provider 未被调用的 Models、auth、scope、quota、route 或其它 provider 前失败固定为 `not_applicable`。请求已经进入 Provider 但没有合法 reported usage 时固定为 `usage_not_reported`。

### `gateway_request_cost_estimate.v1`

Request History v2 中嵌套：

```json
{
  "availability": "estimated",
  "reason": "",
  "currency": "USD",
  "estimated_cost_micros": 42,
  "token_unit": 1000000,
  "input_price_micros_per_token_unit": 1000000,
  "output_price_micros_per_token_unit": 3000000,
  "pricing_policy_id": "gmp_...",
  "pricing_policy_version": 1,
  "pricing_policy_digest": "sha256:...",
  "rounding_mode": "half_up_to_currency_micro"
}
```

`availability` 只允许：

- `estimated`：价格快照与 reported usage 都合法；零成本也必须保留该状态。
- `usage_not_reported`：进入 Provider，但 usage 没有可信报告。
- `price_not_configured`：精确 selection 没有对应策略。
- `price_unavailable`：store、scope、digest 或计算失败，不能安全估算。
- `not_applicable`：没有 Provider side effect 或协议不适用。
- `legacy_not_captured`：历史 `gateway_request_record.v1` 没有成本字段。

金额只使用 checked integer 或 `math/big` 计算：

```text
numerator = input_tokens × input_rate + output_tokens × output_rate
estimated_cost_micros = round_half_up(numerator / 1_000_000)
```

不使用浮点数。溢出、负数、总 token 不一致、非法 rate 或 digest 漂移统一失败关闭为 `price_unavailable`，不返回部分金额。成功和 Provider 失败只要拥有可信 reported usage 与请求级价格快照，都可以估算；估算状态不能改变 HTTP 结果或业务响应。

## Request History v2 与兼容边界

- 新请求写入 `gateway_request_record.v2`，保留 v1 全部字段并新增 `cost_estimate`。
- memory、SQLite 与 PostgreSQL 继续复用同一 `gatewayRequestStore` 语义；必要 migration 只扩 v2 持久化和索引，不创建第二套 request table。
- v1 历史必须继续可读，并投影 `legacy_not_captured`；禁止按当前价格回算旧请求。
- list 直接返回 sanitized cost summary，Application Operations 不逐条请求 detail。
- detail 返回精确 rate snapshot 与 policy lineage；不返回 Provider 原始价格响应或商业合同。
- pricing store 或 Request History store 失败不能改写已经完成的 Provider 响应；Request History 仍保持现有 no-fallback 和脱敏日志语义。

## Application Operations 与 Web 信息架构

### Admin Pricing owner

Admin Control Plane 复用 S7 的管理上下文与单 owner 模式，增加 Pricing 任务：

- 从当前 Provider Route / inventory 中选择精确 Provider / Profile / Model；
- 读取 current policy、version、rate、digest 与更新 lineage；
- 缺少 policy、permission、environment、scope drift、CAS conflict 与 store unavailable 均失败关闭；
- 编辑使用真实整数输入和 USD / 1M token 说明，review 后显式 confirm；
- 不显示 cost limit、remaining budget、invoice 或 production activation。

### Request History 与 Application Operations

- Request History list / detail 显示 availability、USD 估算、policy version / digest 与稳定 reason。
- Application Operations 当前 Gateway 窗口统计 estimated、usage missing、price missing、price unavailable、not applicable 与 legacy 数量。
- 只对 `estimated` 记录汇总当前窗口 `estimated_cost_micros`；明确展示 `has_more` 与 coverage，不冒充全历史。
- Timeline 保留每条请求的 cost availability 与金额；Workflow run 不与 Gateway 金额相加，也不猜测一对一关系。
- application / workspace 切换先清空旧成本证据，迟到响应不得进入新 scope。
- offline fixture 中现有 `estimated_cost` / `cost_limit` 不参与 live owner，也不作为 fallback。

## Pencil 分级与设计前置

本功能五维评分为 `1 / 2 / 2 / 2 / 2 = 9`，采用 `A / 完整 Pencil`：

- 信息架构只在既有 S7 增加 Pricing task，并在 S5 Request / Evidence owner 增加成本证据，不创建新一级页面。
- CAS 价格更新、明确确认和历史不可重算构成新的高风险交互。
- `estimated / usage_not_reported / price_not_configured / price_unavailable / not_applicable / legacy_not_captured` 与窄屏顺序需要代表设计。
- 同一成本契约约束 Admin、Request History 与 Application Operations 三处消费面。

设计只新增 S7 Pricing owner 与 S5 Cost Review 的桌面代表面、必要窄屏和风险状态，不建立 `S11`，不重画完整 S5 / S7。S9 / S10 Visual R3 已先完成人工视觉复核；当前 Visual R1 节点为 S7 Desktop `wQ2t0`、Narrow `Z5Iqv`、S5 Desktop `Ue7hq`、Narrow `i50xIV` 与 Decision R16 `VAToA`。五张画板均已显式保存，通过全树无裁切、无越界、无 placeholder 检查和人工视觉复核；React strict consumer 已据此完成，当前进入产品连续链验收。

## 实施批次

实现涉及新 API、permission、schema version、双数据库 migration、repository 和请求级运行边界。功能设计与 S9 / S10 Visual R3 已于 2026-08-12 人工通过，唯一高风险任务卡已经建立；现按以下批次推进：

### 批次 A：领域、memory owner 与确定性计算

- 定义 pricing policy、immutable revision / current pointer、cost estimate 与稳定 failure。
- 实现 memory repository、CAS、唯一 scope、canonical digest、整数 rate 和 half-up 计算。
- 覆盖零价、零 token、missing usage、missing price、scope drift、digest drift、并发与溢出。

### 批次 B：SQLite / PostgreSQL 持久化

- 新增同构 migration、repository、store selector、schema marker 与 no-fallback。
- 验证 revision append-only、current pointer 原子切换、重启恢复、作用域隔离、受限运行角色和并发单赢家。
- 不复用 quota table、Provider Route JSON 或 Request History 作为价格 owner。

### 批次 C：Admin API、Gateway 绑定与 Request History v2

- 接入独立 GET / PUT、两项 permission、verified identity、strict JSON 与显式环境 gate。
- 在 selection 后、Provider 前绑定 request-local 价格快照，不改变 route / quota / Provider side effect。
- 升级 Request History v2 与 v1 read compatibility，三种 store 保存同一 cost estimate。
- 覆盖 unary / stream、API Key / dev headers、quota rejection、Provider failure、usage missing 与 price store failure。

### 批次 D：Pencil、严格 Web consumer 与产品路径

- 在 S9 / S10 Visual R3 人工复核关闭后冻结 S7 Pricing 与 S5 Cost Review 代表面。
- 实现 Admin price review / confirm、Request History cost detail 与 Application Operations 当前窗口汇总。
- 严格拒绝额外字段、非法整数、scope drift、digest mismatch、混合来源与敏感材料。

### 批次 E：双数据库与真实浏览器收口

- memory、SQLite、PostgreSQL 完成 create → update → request v2 → history → application cost review 连续链。
- 复验历史 v1、价格更新不重算旧请求、服务重启、CAS、price unavailable、usage missing 与隐私。
- 覆盖桌面、关键断点与 `390×844`；确认无横向溢出、迟到响应、控制台 warning / error 或后台残留进程。

批次 A 至 E 已全部完成。PostgreSQL 实例已经证明 migration / runtime role 分离、并发 CAS、服务重建与历史价格快照不重算；SQLite 产品浏览器已经证明价格 v1 → v2、stale conflict、重启恢复、API Key / 开发身份 Gateway、Request History v2 与 Application Operations 当前窗口连续链。真实 mock Provider 不报告 usage，因此浏览器证据保持 `usage_not_reported`，reported usage 的整数估算由 memory、stream、quota 与 PostgreSQL 连续测试证明；两类证据不互相冒充。未触发 Provider 的请求统一落为 `usage=not_applicable` 与 `cost=not_applicable`，避免严格 Web consumer 因终态语义错位拒绝整页。

## 验收方式

- Go domain：canonical digest、policy CAS、immutable revision、checked integer、rounding、availability matrix 与 no side effects。
- store：memory / SQLite / PostgreSQL contract parity、migration、runtime role、scope isolation、concurrency、restart 与 no fallback。
- HTTP / auth：strict JSON、两项 permission、membership、environment、stable failure、sensitive field absence。
- Gateway：selection exact match、price snapshot pin、quota / provider 调用计数、unary / stream usage、v1 / v2 history compatibility。
- Web：strict consumer、scope generation、current-window subtotal、partial coverage、CAS conflict、offline zero request 与 forbidden field guard。
- 产品：Pencil 人工复核、production build、真实 SQLite 浏览器、PostgreSQL integration、仓库 fast 与 full 门禁。

## 停止线

- 不实现 production price、billing ledger、invoice、结算、余额、预算告警、财务审计或 Provider 对账。
- 不实现 token estimation、缓存 token 独立计价、图片 / 音频单位、阶梯价、折扣、税、汇率或多币种换算。
- 不用成本估算执行 quota、rate limit、请求拒绝、自动降级、自动路由、retry / fallback 或 load balancing。
- 不把 Provider reported usage 解释为计费凭证；它只是在既有校验边界内可用于开发测试态估算。
- 不把当前分页窗口写成全历史成本，不逐条 detail 回算，不创建 aggregate table、materialized view 或 cost ledger。
- 不按当前价格重算历史 v1 / v2 请求，不修改历史价格快照。
- 不把价格塞入 Provider Route、quota、API Key、application、Workflow run 或发布 owner。
- 不打开真实 Radish OIDC、production secret、Provider credential / endpoint、公开生产 API 或外部项目接线。
- 不建立 S11，不在 S9 / S10 Visual R3 未完成人工复核时继续堆叠未审完整设计面。

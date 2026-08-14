# Provider 上报用量规范化与应用用量审查（开发 / 测试态）v1

更新时间：2026-07-27

状态：`provider_reported_usage_normalization_application_review_dev_test_v1_completed`

## 功能目标

把 Provider 已明确上报的 token 用量规范化为统一 Gateway usage contract，并沿既有 Request History 真相源持久化、查询和展示，使内部开发者能够按 application 审查可信的输入、输出与总 token 用量。

本专题不估算 token。Provider 没有上报、流式响应没有携带最终 usage、字段不完整或不满足整数约束时，统一保留 `not_reported`，不得用零伪装真实用量。

## 当前缺口

- Gateway Adapter 已持有 OpenAI-compatible、Gemini native 与 Anthropic 原始响应，但 `GatewayEnvelope` 尚未输出经校验的 usage。
- Platform recorder 创建记录时固定写入 `not_reported`，无法承接真实 Provider 用量。
- `/v1/responses` 与 `/v1/messages` 当前用零值填充 northbound usage，缺少来源证据；`/v1/chat/completions` 尚未返回 usage。
- Request History detail 已具备 reported usage 结构，但列表和 Application Operations 只能查看 availability，不能审查 token 汇总。

## 规范化契约

Gateway metadata 新增必需字段 `usage`：

```json
{
  "availability": "reported",
  "source": "gemini_usage_metadata",
  "input_tokens": 120,
  "output_tokens": 32,
  "total_tokens": 152
}
```

规则如下：

- `availability` 只允许 `reported` 或 `not_reported`；`not_applicable` 继续由 Platform 请求生命周期表达 Provider 未执行场景，不由成功 Provider 响应推导。
- `reported` 要求 `source` 为稳定 allowlist 值，三个计数均为非负整数，且 `total_tokens = input_tokens + output_tokens`。
- `not_reported` 的 `source` 为空且三个计数为零；零只表示 contract 的空值形态，不表示 Provider 实际使用零 token。
- Provider 字段缺失、类型错误、负数、布尔值、总数不一致或部分存在时，只降级 usage availability，不改变已经通过响应 schema 校验的业务响应。
- Gateway envelope、Request History 与 Web consumer 只传递规范化字段，不暴露 Provider raw response、endpoint、credential 或自由文本。

## Provider 来源映射

- OpenAI-compatible、HuggingFace OpenAI-compatible 与 Ollama OpenAI-compatible：读取 `usage.prompt_tokens`、`usage.completion_tokens` 和 `usage.total_tokens`。
- Gemini native：以 `promptTokenCount` 作为 input，以 `totalTokenCount - promptTokenCount` 作为规范化 output，从而包含 Provider 已计入总量的 thinking tokens；`candidatesTokenCount` 必须不大于规范化 output。该关系依据 [Gemini UsageMetadata](https://ai.google.dev/api/generate-content#UsageMetadata)。
- Anthropic Messages：规范化 input 为 `input_tokens + cache_creation_input_tokens + cache_read_input_tokens`，output 使用 `output_tokens`；缺失的 cache 分项按 Provider 可选字段的零值处理。该关系依据 [Claude Messages usage](https://platform.claude.com/docs/en/api/messages)。
- Ollama 原生兼容结果如存在 `prompt_eval_count` 与 `eval_count`，读取两项并计算总数。
- 流式调用只在已收到的终态 chunk 明确携带有效 usage 时标记 `reported`；本批次不为兼容性不明的 Provider 强制增加专有请求参数。
- deterministic mock 不执行真实 Provider，因此继续为 `not_reported`，不根据字符串长度或本地 tokenizer 生成伪用量。

## Gateway 与 Request History

- `GatewayEnvelope.metadata.usage` 是 Provider Adapter 到 Platform 的唯一 usage 交接边界。
- recorder 在 envelope 通过解码后复验 usage；合法 `reported` 写入现有 `gateway_request_record.v1`，非法或缺失保持 `not_reported`。
- SQLite 与 PostgreSQL 继续持久化现有完整 record JSON 和 `usage_availability` 物化列，不新增 migration；现有 record schema 已包含计数与 source。
- Request History list 增加 reported token summary，detail 保留 source 与三个精确计数；非 reported 状态不展示零值为实际用量。
- usage store 或展示失败不得改写 Provider 响应结果，也不得回退到估算值。

## Northbound 协议

- `/v1/chat/completions`、`/v1/responses` 与 `/v1/messages` 的 unary 响应只在 Gateway usage 为 `reported` 时输出协议 usage。
- stream 在终态事件可取得 reported usage 时输出协议允许的最终 usage；中间 delta 不重复累计。
- Gateway usage 为 `not_reported` 时省略可选 usage，或按协议既有必需结构返回明确不可误解的空形态；不得继续把缺失证据写成实际零用量。
- northbound usage 只是 Provider 报告值的兼容投影，不构成计费凭证、quota 扣减或账本。

## Application Operations 审查

- timeline item 增加 usage source 与输入、输出、总 token；只对 `reported` 记录展示。
- 当前查询窗口的 metrics 增加 reported request token 汇总，并保留 reported / not reported / not applicable 数量。
- 汇总只覆盖已加载且与当前 application scope 精确绑定的 Request History 页面，不把 Workflow duration、离线 quota snapshot 或其他 application 混入。
- UI 明确标注“Provider 上报”“当前查询窗口”，不表达价格、预算剩余或生产用量结算。

## 实施批次

唯一实施入口见[实施任务卡](../../task-cards/provider-reported-usage-normalization-application-review-dev-test-v1-plan.md)。

1. 批次 A：usage schema、Provider 来源映射与 Gateway envelope。
2. 批次 B：Platform recorder、三类 northbound unary / stream 投影。
3. 批次 C：Request History 与 Application Operations consumer / UI。
4. 批次 D：Python、Go、Web、SQLite / PostgreSQL 与仓库门禁收口。

## 验收

- 各 Provider shape 的合法、缺失、部分、负数、布尔值和总数不一致 fixture 均有确定结果。
- mock 与没有 usage 的 stream 保持 `not_reported`。
- 三类 northbound unary 和可取得终态 usage 的 stream 返回精确协议字段；没有证据时不伪造零用量。
- memory、SQLite 与 PostgreSQL 记录重启前后保持 source 和三个计数一致，既有 availability filter 不回归。
- Request History list / detail 与 Application Operations 的 token 汇总来自同一 record，跨 application scope 不混入。
- Web consumer 拒绝非法 reported usage，页面不暴露 raw response、credential、endpoint 或正文。

## 完成结果

- Python Provider Adapter 已规范化 OpenAI-compatible、HuggingFace、Ollama、Gemini native 与 Anthropic Messages usage；Gemini thinking tokens 和 Anthropic cache input tokens 均保留在统一计数内，缺失或非法 shape 退回 `not_reported`。
- `CopilotGatewayEnvelope.metadata.usage` 已成为严格 contract；Gateway 二次净化 Provider 输出，非法 usage 不改变已验证业务响应。
- Platform recorder、memory、SQLite 与 PostgreSQL 复用既有 `gateway_request_record.v1` 保存 source 和三个计数；stream 终态 envelope 与 unary 使用同一落库语义。
- `/v1/chat/completions`、`/v1/responses` 与 `/v1/messages` 只在 `reported` 时投影 usage；没有证据时不再返回伪造零值。
- Request History list / detail 与 Application Operations 已展示同一份 reported usage；应用面只汇总当前已加载窗口，并明确区分 availability、source 与完整历史。
- Python、Platform 全包 Go、SQLite、PostgreSQL integration、Web 255 项测试与 production build 均已通过；既有 Gateway Request History migration 无需变化。

## 停止线

- 不实现 token 估算、本地 tokenizer 补算、价格表、成本换算、quota 扣减、billing ledger 或预算告警。
- 不改变 Provider selection、retry、fallback、load balancing 或请求调度。
- 不为了取得流式 usage 强制所有 OpenAI-compatible endpoint 接受专有参数。
- 不打开 production API key、production OIDC、production repository 或生产能力声明。
- 不新增同层 readiness / refresh / checker 链；由相邻单元测试、现有数据库集成与聚合门禁承载。

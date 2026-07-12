# Model Gateway Request History / Usage & Failure Review v1 任务卡

更新时间：2026-07-12

## 任务标识

- 任务 ID：`model-gateway-request-history-usage-failure-review-v1`
- 状态：`in_progress_postgres_dev_test_vertical_slice_implemented`
- 对应功能专题：`docs/features/gateway/model-gateway-request-history-usage-failure-review-v1.md`

## 当前批次进度

- 已完成：caller context、record lifecycle / taxonomy、`memory_dev` store、三个 northbound route recorder、list / detail API、scope / filter / cursor、Web consumer、Evidence Review lazy panel，以及 Go / race / vet / Web test / build 验证。
- 已确认：缺少 caller context 时不创建 unscoped record；store create failure 不改写 provider 成功；usage 无可信来源时保持 `not_reported`；Web 默认 offline 零请求且拒绝嵌套 forbidden field。
- 第二批已完成：独立 `postgres_dev_test` selector、manual migration、schema marker / checksum preflight、runtime role、pool close、no-fallback integration、27 条真实浏览器分页 / 详情 / 过滤和 Platform / Web 重启恢复。
- 待完成：真实 provider 缺席时继续保持 usage `not_reported`；补 queue / timeout / cancel / stream cancel 的浏览器终态证据与最终功能 close。不得为此新增同层 repository、API、面板或 checker。

## 用户结果

内部开发者可以从现有 Model Gateway Evidence Review 查看真实 northbound 请求列表与详情，按路由、provider、状态、失败和时间定位问题，审查经验证的 usage 与 timing，并在 PostgreSQL 开发 / 测试模式下跨 Platform 重启恢复记录；默认 offline 路径仍零请求，记录不包含输入、输出或 credential。

## 实现范围

1. 新增 `GatewayCallerContextProvider`，只实现显式 opt-in dev/test provider；完整 caller scope 为 tenant + workspace + consumer，可选 application。
2. 新增 `gateway_request_record.v1` domain、稳定 lifecycle / failure / usage taxonomy 和 `gatewayRequestStore` contract。
3. 实现 `memory_dev` 默认 store，覆盖版本并发、终态保护、scope、过滤、keyset cursor、FIFO、clone 和 forbidden-field guard。
4. 在现有 request trace / observed response 边界增加 recorder，覆盖 chat completions、responses、messages 的 unary、stream、invalid body、selection、bridge/provider/translation failure 和 cancellation。
5. 新增 scoped list / detail read API，不改变现有 northbound request / response contract。
6. 新增独立 `postgres_dev_test` selector、manual migration、marker/checksum preflight、运行角色权限、连接池关闭和 no-fallback integration。
7. 新增真实 Web consumer 与 lazy Request History panel，接入现有 Gateway Evidence Review，保留 offline evidence 与错误状态。
8. 完成 Go / race / vet、PostgreSQL、Web、浏览器和仓库门禁验证，并同步功能文档、当前焦点和周志。

## 关键实现约束

- caller scope 不从 request body、query、model alias 或 `radishmind` extension 自报；production provider 缺席时 production mode fail closed。
- caller context 不完整时不创建 unscoped record；现有 compatibility request 可继续，但必须记录稳定 `caller_context_unavailable` observation。
- `started → succeeded|failed|canceled` 为唯一终态迁移；终态不可逆，record version 原子递增。
- stream 只有 terminal event 成功写出后才能 succeeded；client / context cancellation 映射 canceled。
- record store failure 不改写 provider outcome，但必须显式可观测；PostgreSQL 失败不得 fallback memory。
- token usage 只有经校验且确认由 response contract 报告时才为 `reported`；不估算 token、cost 或 quota。
- 不为本批新增平行 checker、readiness fixture 或第二套 Gateway contract。

## API 与 persistence 资产

```text
GET /v1/model-gateway/requests
GET /v1/model-gateway/requests/{request_id}

services/platform/migrations/gateway_requests/
gateway_request_schema_versions
gateway_request_records
```

默认 `memory_dev`；`postgres_dev_test` 必须显式选择，并要求独立 dev/test DSN、manual migration 和 caller-context gate。

## 验收矩阵

| 场景 | 必须结果 |
| --- | --- |
| caller context 缺失 | northbound compatibility 行为不变；不创建 unscoped record，稳定 observation 可见 |
| invalid JSON / body too large | 有效 caller scope 下保存 failed record，boundary 为 northbound request |
| unary success | terminal succeeded，selection / timing 可审查，usage 按来源标记 |
| stream success / cancel | terminal event 后 succeeded；中途取消为 canceled，不保存 delta |
| queue / timeout / worker / provider failure | 稳定 code / boundary / category，不含 raw error |
| store create / terminal update failure | provider outcome 不被改写；store failure 可观测；无 memory fallback |
| list / detail scope | tenant/workspace/consumer/application binding 隔离，不暴露跨 scope 存在性 |
| filter / cursor | keyset 稳定，cursor 绑定 scope 与过滤摘要，篡改 fail closed |
| PostgreSQL restart | 列表与详情恢复，started stale 派生正确 |
| Web offline | 零 HTTP，明确 offline evidence |
| Web dev/test | 列表、过滤、分页、详情、usage unavailable 和失败状态完整 |
| 敏感材料 | prompt、response、headers、credential、endpoint、raw envelope / error 均不存在 |
| 禁止副作用 | retry/fallback、quota/billing、tool、confirmation、business write、replay 均为 0 |

## 必要验证

- 精准 Go domain / store / HTTP tests 与 `go test -race`。
- `go test ./...`、`go test -race ./...`、`go vet ./...`。
- PostgreSQL migration fresh apply、rollback / reapply、runtime role DDL 拒绝、scope / paging / concurrency / restart / no fallback integration。
- Web consumer unit tests、`npm test`、`npm run build` 与 chunk budget。
- 真实浏览器列表 / 详情 / 失败 / stream cancel / restart 路径，完成后关闭 Platform、Web、浏览器和 PostgreSQL。
- `./scripts/check-repo.sh --fast` 与 `./scripts/check-repo.sh`。

## 提交拆分

1. Platform domain / memory store / recorder / read API 与测试。
2. PostgreSQL migration / selector / integration 与双端 launcher / CI 资产。
3. Web consumer / lazy panel / browser 验证。
4. 功能完成状态、当前焦点、专题入口和周志收口。

## 停止线

- 不实现 production API key / OIDC auth、quota、rate limit、billing、cost ledger、自动 retry/fallback 或 load balancing。
- 不持久化输入输出、credential、endpoint、provider raw material 或完整 headers。
- 不复用 Workflow run repository、Saved Draft repository、control-plane fake store 或离线 cost snapshot。
- 不实现 retention 后台任务、replay / resume、自动修复、production repository 或 public production ready 声明。

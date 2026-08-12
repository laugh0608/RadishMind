# Provider 价格策略版本与应用成本审查（开发 / 测试态）v1 实施任务卡

更新时间：2026-08-12

状态：`provider_pricing_policy_version_application_cost_review_dev_test_v1_completed`

对应功能文档：[Provider 价格策略版本与应用成本审查（开发 / 测试态）v1](../features/gateway/provider-pricing-policy-version-application-cost-review-dev-test-v1.md)

## 任务目标

建立独立于 Provider Route、Gateway quota 与业务账单的开发测试态价格策略 owner，让每次 Gateway 请求在精确 Provider / Profile / Model selection 后固定不可变价格快照，并只对 Provider 合法上报的 usage 生成确定性整数成本估算；同一份脱敏证据进入 Request History v2、Admin Pricing 与 Application Operations 当前窗口审查。

本任务卡是该专题唯一高风险实施卡。新增 API、permission、schema version、SQLite / PostgreSQL migration、repository、Gateway 请求边界与 Web 消费都在这里收口；不派生同层 readiness、refresh、review 或 checker 链。

## 人工批准记录

- 2026-08-12：功能设计人工通过。
- 2026-08-12：S9 Desktop `C7pkb`、Narrow `x8lESc` 与 S10 Desktop `Um8Zh`、Narrow `ZxJd7` 的 Visual R3 人工通过。
- 2026-08-12：S7 Pricing Desktop `wQ2t0` / Narrow `Z5Iqv`、S5 Cost Review Desktop `Ue7hq` / Narrow `i50xIV` 与 Decision R16 `VAToA` 的 Visual R1 人工通过，授权 React strict consumer 采用。
- 本批准关闭功能设计、S9 / S10 Visual R3 与本专题 Visual R1 前置条件，不代表数据库实例、React 或产品连续链已完成。

## 固定架构

```text
Admin exact pricing scope
  -> immutable revision + atomic current pointer
  -> Gateway exact selection
  -> request-local immutable pricing snapshot
  -> validated Provider reported usage
  -> deterministic integer cost estimate
  -> Gateway Request History v2
  -> Request detail + Application Operations current-window review
```

- `GatewayModelPricingRepository` 是价格策略唯一 owner；Provider Route、quota、API Key、application 与 Workflow run 不保存价格。
- scope 固定为 `tenant_ref + workspace_id + environment + provider_id + profile_id + model_id`，不允许通配、继承、别名或 fallback。
- `development | test` 只允许 USD / 1M token 的非负 `int64` 微美元 input / output rate。
- update 追加不可变 revision 并以 CAS 原子切换 current pointer；旧 request snapshot 永不重算。
- 价格解析发生在精确 selection 后、Provider side effect 前，但缺价格或 store 失败不改写 selection、quota 或 Provider 调用结果。
- 成本只消费合法 `reported` usage；不使用 tokenizer、正文长度或当前价格推算。
- Request History v1 继续可读并投影 `legacy_not_captured`；新请求写 v2。
- Application Operations 只汇总当前已加载窗口，并始终展示 coverage 与 `has_more`。

## 修改范围

### 允许

- `services/platform/internal/httpapi/` 中相邻 pricing、Gateway request、HTTP、store 与 server 装配；
- `services/platform/internal/config/` 中显式开发测试态 gate 和 store selector；
- `services/platform/migrations/` 与共享 SQLite migration 中的 pricing owner、Request History v2 兼容迁移；
- `apps/radishmind-web/` 中 Admin Pricing、Request History cost detail 与 Application Operations current-window review；
- 对应功能文档、Family UI 产品化专题、当前焦点、路线图、任务卡和本周周志。

### 禁止

- production price、billing ledger、invoice、结算、余额、预算告警、财务审计或 Provider 对账；
- token estimation、缓存 token 单独计价、图片 / 音频单位、阶梯价、折扣、税、汇率或多币种；
- 用估算成本执行 quota、rate limit、请求拒绝、自动路由、retry、fallback 或 load balancing；
- 把价格写入 Provider Route、quota、API Key、application、Workflow run 或发布 owner；
- 按当前价格回算历史记录，或把当前分页窗口冒充全历史；
- 接真实 Radish OIDC、production secret、Provider credential / endpoint 或公开生产 API；
- 建立 S11，或新增现有单元、集成、Web 和仓库聚合门禁可以承载的专项 checker。

## 批次 A：领域、memory owner 与确定性计算

- [x] 定义 pricing scope、policy revision、current pointer、request snapshot、cost estimate 与稳定 failure。
- [x] 实现 canonical scope、policy ID、canonical digest、严格 rate / reason 校验和深复制边界。
- [x] 实现 memory repository 的 append-only revision、current read 与 CAS 单赢家。
- [x] 实现无浮点、checked integer / `math/big` half-up 成本计算和六态 availability。
- [x] 覆盖零价、零 token、missing usage、missing price、scope / digest drift、并发与溢出。

退出条件：领域与 memory repository 能独立证明不可变 revision、作用域隔离、CAS、确定性计算、无 fallback 和无外部副作用。

## 批次 B：SQLite / PostgreSQL 持久化

- [x] 新增同构 migration、repository、schema marker、独立 PostgreSQL migration runner 与 store selector。
- [x] 验证 append-only revision、current pointer 原子切换、重启恢复、作用域隔离与并发单赢家。
- [x] 验证 SQLite 聚合 runtime 与 PostgreSQL migration / runtime role 边界，无 memory fallback。

退出条件：memory、SQLite 与 PostgreSQL repository contract 等价；未迁移、marker 漂移、连接失败和 corruption 全部失败关闭。

## 批次 C：Admin API、Gateway 绑定与 Request History v2

- [x] 接入 GET / PUT、`admin_gateway_pricing:read | write`、verified identity、strict JSON 和环境 gate。
- [x] 在 selection 后、Provider 前固定价格 snapshot 或稳定 unavailable reason。
- [x] 对 unary / stream 的 reported usage 计算 terminal estimate，不改变 Provider 响应。
- [x] 升级 Request History v2，并让三种 store 保存同一 cost estimate；v1 投影 `legacy_not_captured`。
- [x] 覆盖 quota / route / auth 前失败、Provider failure、usage missing、price missing / unavailable 与历史兼容。

退出条件：Admin 更新只影响后续请求；同一请求的 selection、snapshot、usage 与 estimate 谱系可复验，价格失败不改变调用副作用。

## 批次 D：Pencil 与严格 Web consumer

- [x] 在已通过 S9 / S10 Visual R3 基准上冻结 S7 Pricing 与 S5 Cost Review 代表面（Visual R1 已保存、通过结构检查并完成人工视觉复核）。
- [x] 实现 Admin exact-scope price read / review / confirm 与 CAS conflict 恢复。
- [x] 实现 Request History list / detail 的 availability、金额、version / digest 与稳定 reason。
- [x] 实现 Application Operations 当前窗口 subtotal、六态 coverage 与 `has_more` 提示。
- [x] 拒绝额外字段、非法整数、scope drift、digest mismatch、混合来源与敏感材料。

退出条件：React 只消费 live owner 的脱敏数据；scope 切换清空旧证据，迟到响应不能进入新 owner。

## 批次 E：双数据库与真实浏览器收口

- [x] memory、SQLite、PostgreSQL 完成 create → update → request v2 → history → application cost review 连续链。
- [x] 复验 v1 history、价格更新不重算旧请求、重启、CAS、price unavailable 与 usage missing。
- [x] 覆盖 Desktop、关键断点与 `390×844`，无横向溢出或控制台 warning / error。
- [x] 完成精准测试、Platform Go、Web、production build、仓库 fast 与 full 门禁。

退出条件：双数据库和真实浏览器证据关闭后，专题才可标记完成；不得用 memory 或静态 fixture 代替 durable owner。

关闭证据：真实 PostgreSQL 通过独立 migration runner、运行角色、并发 CAS 和服务重建，确认请求绑定 v1 价格快照后即使 current policy 更新为 v2，重连读取仍保留 v1 digest 与 `20` 微美元估算。SQLite 产品浏览器完成 v1 创建、v2 更新、双标签 stale CAS conflict、服务重启恢复、API Key 与开发身份 Gateway 请求、Request History v2 和 Application Operations 当前窗口审查；mock Provider 不上报 usage 时保持 `usage_not_reported`，`/v1/models` 等未触发 Provider 的记录同时使用 `usage=not_applicable` 与 `cost=not_applicable`。`1440×900`、`720×900`、`390×844` 无横向溢出，最终页面控制台无 warning / error，临时 API Key 已撤销且自启动服务已停止。

## 验证顺序

1. 每个批次先运行相邻 Go / Web 单元与集成测试。
2. 并发 CAS 与 request snapshot 批次运行 `go test -race`。
3. SQLite 和 PostgreSQL 分别证明本地产品连续性与数据库事务 / 角色边界。
4. Web 批次运行 strict consumer、完整 Web tests、production build 和真实浏览器链。
5. 批次 A、持久化完成、Gateway 接入和最终关闭时运行仓库快速门禁；最终关闭运行全量门禁。

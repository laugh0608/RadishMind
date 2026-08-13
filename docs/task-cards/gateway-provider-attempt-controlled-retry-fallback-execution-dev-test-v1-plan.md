# Gateway Provider Attempt 受控重试与降级执行（开发 / 测试态）v1 实施任务卡

更新时间：2026-08-13

状态：`gateway_provider_attempt_controlled_retry_fallback_execution_dev_test_v1_batch_d_completed_batch_e_next`

对应功能文档：[Gateway Provider Attempt 受控重试与降级执行（开发 / 测试态）v1](../features/gateway/provider-attempt-controlled-retry-fallback-execution-dev-test-v1.md)

## 任务目标

在既有 Provider Route、application request quota、Gateway Pricing、bridge 与 Request History owner 上建立一条显式、有限、可审计的开发测试态 Provider 顺序降级执行链：只有 Route v2 已激活、调用者逐请求允许、主目标产生类型化 eligible failure 且尚未开始 northbound 响应时，才对不同 Provider Profile 发起最多一次备用 attempt。

本任务卡是该专题唯一高风险实施入口。Route v2、attempt plan、类型化 Provider failure、Request History v3、双数据库、Admin API、三个 unary northbound API、Pencil、React 与产品连续链全部在这里收口；不派生同层 readiness、refresh、review、manifest-only 或 gate-only 子链。

## 人工批准记录

- 2026-08-13：功能设计获得人工批准，可以创建唯一高风险任务卡并进入批次 A。
- 本次批准只授权既定开发测试态范围，不代表 runtime fallback、数据库迁移、HTTP、Pencil、React、真实 Provider 或 production capability 已完成。
- 批次 E 的完整 Pencil 仍需单独人工视觉批准；批准前不得实现 React strict consumer。

## 固定架构

```text
active Provider Route v2 snapshot
  -> immutable request-local attempt plan
  -> pricing snapshot for every planned target
  -> attempt .pa1 quota admission + durable checkpoint
  -> typed Provider failure observation
  -> durable fallback_pending checkpoint
  -> attempt .pa2 quota admission + durable checkpoint
  -> one root Gateway Request History v3 record
```

- Route 草案、候选、激活与快照继续由既有 Admin Provider Route owner 承载；v2 只增加固定执行模式与最多两个有序 target。
- `gateway_provider_attempt_plan.v1` 是 request-local 不可变值，不成为新配置真相源。
- attempt id 固定为 `<root_request_id>.pa1` 与 `<root_request_id>.pa2`；根 request id 不变。
- 每个实际 attempt 独立执行 quota admission；不预留、不退款，未执行的备用 target 不计数。
- 每个 plan target 在首个 Provider side effect 前固定价格 snapshot；执行中不重新读取 current pointer。
- Request History v3 在同一根记录内保存最多两个 attempts；v1 / v2 继续按单 attempt 兼容读取。
- Go 只消费类型化、脱敏的 Provider failure，禁止解析错误消息猜测是否降级。

## 修改范围

### 允许

- `services/platform/internal/httpapi/` 中既有 Provider Route、Gateway request、quota、pricing、bridge 与 server 相邻边界；
- `services/platform/internal/bridge/` 与 `services/runtime/` 中类型化、脱敏 Provider failure 合同；
- `services/platform/internal/config/` 中独立开发测试态 gate；
- `services/platform/migrations/` 与共享 SQLite migration 中 Route v2、Request History v3 的兼容持久化；
- `apps/radishmind-web/` 中既有 S7 Route、S5 Playground / Request History strict consumer；
- 对应功能文档、任务卡、当前焦点、路线图、能力矩阵和本周周志。

### 禁止

- production Gateway、production API Key、production quota、billing、生产 secret resolver 或真实生产发布；
- 同 Profile retry、指数退避、后台重试、并行竞速、hedging、负载均衡、随机 / 权重路由、熔断或 live health routing；
- stream fallback，或 northbound 响应已开始后的 Provider 切换；
- 把 bridge worker、client、schema、auth、quota、pricing、history、safety 或 platform response failure 当作 fallback 信号；
- Application RAG、Prompt、Agent、Workflow、Session 或 Evaluation Campaign 自动降级；
- 客户端提交 Provider / Profile、target 顺序、eligible failure、attempt id 或最大次数；
- 创建第二套 Route、quota、pricing、Request History、inventory 或 selection owner；
- 建立 S11 或新增现有单元、集成、Web 与仓库聚合门禁可以承载的专项 checker。

## 批次 A：领域合同、Route v2 与 memory owner

状态：已完成；批次 B 已完成。

- [x] 定义 Route v2 execution mode、attempt targets，并保持 v1 snapshot 单 attempt 兼容。
- [x] 定义不可变 attempt plan、target、确定性 attempt id 与完整前置校验。
- [x] 定义类型化 Provider failure、eligible allowlist、ineligible / unknown / response-started 边界。
- [x] 定义 Request History v3 root / attempt / cost summary 与 v1 / v2 兼容边界。
- [x] 扩展 memory owner，支持根记录、attempt checkpoint 与 `fallback_pending` 原子状态推进。
- [x] 扩展 deterministic fake bridge / adapter，覆盖主成功、eligible、ineligible 与 unknown outcome。
- [x] 覆盖深复制、作用域隔离、非法状态转换、重复 attempt、终态不可改写和敏感材料拒绝。

退出条件：领域与 memory 测试能独立证明 Route v2、不可变 plan、类型化 failure、最多两次 attempt、durable checkpoint、旧记录兼容和零外部副作用；不得以 HTTP 或数据库集成补偿领域缺口。

## 批次 B：SQLite / PostgreSQL

状态：已完成；批次 C 已完成。

- [x] 增加 Route v2 / Request History v3 同构 migration、marker、repository 与 store selector 兼容；旧 v1 / v2 迁移链保持可升级。
- [x] 验证 v1 / v2 历史、Route v1 snapshot、Route v2 与 attempt checkpoint 重启恢复、原子状态迁移及损坏 payload 失败关闭。
- [x] 验证 PostgreSQL runtime role 无 DDL、并发单赢家、连接失败和 no fallback；SQLite 同步覆盖文件数据库并发与重启。

批次证据：SQLite Route 与 Request History 分别推进到 `admin_provider_routes_store_v2`、`gateway_requests_store_v3`；PostgreSQL 分别推进到 `admin_provider_routes_store_v2`、`gateway_request_store_v3`。Request History 双数据库更新都在事务内读取并锁定当前根记录，再复用 memory owner 的 Provider Attempt 状态迁移校验；v3 另保存 `provider_attempt_count`、`fallback_used` 与终态 Provider / Profile 可查询摘要，但完整 lineage 仍以严格 JSON record 为真相源。真实 PostgreSQL 17 完整平台集成门禁通过，容器已关闭。

退出条件：memory、SQLite、PostgreSQL 对 Route v2 与 Request History v3 的 repository contract 等价。

## 批次 C：Admin API 与激活链

状态：已完成；批次 D 已完成。

- [x] 扩展既有 Route draft / candidate / review / activation API，不创建第二套 endpoint 或 permission。
- [x] 覆盖 strict JSON、target 顺序、能力兼容、inventory digest、draft / review / generation CAS、rollback 与 v1 snapshot。
- [x] 确保 review 不改变运行时，只有 activation 切换后续请求的 generation。

批次证据：既有 `PUT /v1/admin/provider-route-configurations/{configuration_id}` 已移除批次 A 的 v1-only 停止线，直接复用同一 strict decoder、领域规范化、四项权限和三存储 owner 消费 Route v2。HTTP 连续链已证明 v1 generation 1 → v2 review 后仍保持 generation 1 → 显式 activation 切换 generation 2 → rollback 回到 v1 generation 3；嵌套未知 / 重复字段、target 数组与 ordinal 不一致、备用 Profile capability 不兼容、陈旧 revision / review version / generation 和任一 target inventory digest 漂移均在运行时切换前失败关闭。真实 PostgreSQL 17 聚合门禁已用 Route v2 完成同一 Admin HTTP draft → candidate → review → activation → history 链，容器随后关闭并保留命名数据卷。

退出条件：Route v2 可由既有人工审查链安全激活，任一 target 漂移都在 Provider 前失败关闭。

## 批次 D：三个 northbound unary API

状态：已完成；批次 E 下一顺位。

- [x] 接入 `fallback_mode=disabled | allow_configured`，省略时保持 disabled。
- [x] 建立 request-local executor，串联 plan、pricing、逐 attempt quota、bridge 与 history checkpoint。
- [x] 仅接 `/v1/chat/completions`、`/v1/responses`、`/v1/messages` 非流式 API Key 请求。
- [x] 返回脱敏 attempt count / fallback used 头与错误 metadata；成功正文保持既有兼容合同。
- [x] 覆盖主成功、eligible → fallback success、双失败、second quota denial、取消、history failure、价格更新和零隐式降级。

批次证据：新增默认关闭的 `RADISHMIND_GATEWAY_PROVIDER_FALLBACK_DEV`，只有 API Key、Admin snapshot、Request History、quota 与 pricing 五项开发测试态前置同时成立且环境一致时才可启用。三个 unary handler 共享 request-local executor，在首个 Provider side effect 前冻结 Route generation / digest、全部 inventory binding 和逐 target pricing snapshot，并以 `.pa1` / `.pa2` 完成逐 attempt quota 与 Request History v3 checkpoint；只有合法的类型化 eligible Provider failure 能进入备用目标。三协议测试已覆盖主成功、显式降级成功、双失败、第二次 quota 拒绝、取消、history checkpoint / terminal failure、Route / pricing 请求内固定、未知或不可降级失败、stream 拒绝、gate 关闭和 target 漂移，成功正文保持既有协议合同，响应只增加脱敏 attempt 头。Platform 全包、`go vet` 与相关 race 均通过；本批没有 UI、真实 Provider 或 production capability，因此未启动浏览器。

退出条件：备用 Provider 最多调用一次；stream、非 API Key、ineligible、unknown 和 response-started 路径全部保持零 fallback。

## 批次 E：Pencil、React 与产品连续链

状态：下一顺位；等待 Pencil 代表面与产品连续链。

- [ ] 在既有 S7 Route 与 S5 Playground / Request History 页面族冻结完整 Pencil 代表面并完成人工批准。
- [ ] 实现主备 Route、请求级允许和连续 attempt evidence strict consumer。
- [ ] 完成 memory / SQLite 浏览器主失败 → 备用成功、双失败、quota 阻断、旧记录与重启链。
- [ ] 完成 PostgreSQL migration、并发、checkpoint、重连与三协议一致性。
- [ ] 覆盖 Desktop、关键断点与 `390×844`，并完成隐私、Web、build、race 和仓库门禁。

退出条件：双数据库、三个协议、真实浏览器和人工设计证据全部关闭后，专题才可标记完成。

## 验证顺序

1. 每批先运行相邻 Go / Python / Web 单元与集成测试。
2. attempt checkpoint、quota 和 Route activation 并发路径运行 `go test -race`。
3. SQLite 与 PostgreSQL 分别证明本地产品连续性和事务 / 角色边界。
4. Web 批次运行 strict consumer、完整 Web tests、production build 和真实浏览器链。
5. 批次 A、持久化完成、northbound 接入与最终关闭时运行仓库快速门禁；最终关闭运行全量门禁。

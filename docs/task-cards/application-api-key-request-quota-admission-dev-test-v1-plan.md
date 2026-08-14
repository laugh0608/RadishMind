# 应用 API Key 请求配额与 Provider Attempt 准入实施任务卡

更新时间：2026-08-09

状态：`application_api_key_request_quota_admission_dev_test_v1_completed`

## 目标

按[功能专题](../features/gateway/application-api-key-request-quota-admission-dev-test-v1.md)交付开发测试态应用级 UTC 日请求预算：管理员显式维护 policy，API Key inference 在真正进入 provider 前原子 admission，超额失败关闭且 provider 副作用为零。

本任务卡是该高风险执行边界的唯一实现入口，不派生 readiness、refresh、manifest-only 或 gate-only 子链。

## 前置事实

- API Key 生命周期、作用域、active application 校验和三模式 repository 已完成。
- Gateway Request History、provider reported usage 和三协议 unary / stream 已完成，但不是 quota owner。
- Prompt、Agent 与 Application RAG 已具备 exact authority、run record 和 provider side-effect 计数。
- 真实 workspace quota 当前固定 `quota_policy_unavailable`；实现必须新增 owner、migration、API 和 provider 前 admission。
- 2026-08-09 已获得用户对 API、schema、repository 与唯一高风险任务卡的明确扩展授权。

## 批次 A：领域与 memory

状态：已完成。

- 定义 policy、usage、admission input / decision、稳定 failure 与严格校验。
- 实现 memory repository 的 policy CAS、UTC bucket、request id 唯一性与并发 admission。
- 固定 admitted provider attempt 计数；provider 失败不退款，quota denial 不调用 provider。

退出条件：领域与 memory 测试覆盖剩余一席并发单赢家、跨作用域、重复 attempt、周期切换和 store failure。

## 批次 B：双数据库

状态：已完成。

- 新增 PostgreSQL migration / manifest / runner 与 SQLite 聚合 migration。
- 实现 SQLite / PostgreSQL policy、usage 与 admission transaction。
- 增加 store mode selector、配置、migration 校验、unknown mode / schema mismatch no-fallback。
- 验证重启恢复、runtime DDL 拒绝、并发单赢家和敏感信息扫描。

退出条件：memory、SQLite、PostgreSQL 语义一致，双数据库门禁通过。

## 批次 C：Admin API

状态：已完成。

- 新增 GET / PUT policy route、workspace / environment 开发测试 header、read / write 权限。
- verified tenant / subject 与 workspace membership 在 repository 前失败关闭。
- PUT 使用 strict JSON、expected version CAS 和 `Cache-Control: no-store`。
- 响应只返回 policy、当前 UTC usage、remaining、lineage 和稳定 failure。

退出条件：权限、作用域、production environment、unknown field、CAS、missing policy 与 store failure 的 HTTP 负向测试通过。

## 批次 D：Gateway admission

状态：已完成。

- 用受控 bridge wrapper 在 provider 前执行 quota admission；没有 API Key quota context 时保持既有行为。
- 为三条标准 inference route 与 Application RAG、Prompt、Agent API Key invocation 注入可信 quota context。
- 确保输入 / authority / idempotency / retrieval 失败不计数；实际 provider attempt 无论结果都计数。
- 标准 API 使用稳定 `429`；三类应用 invocation 的严格 envelope 和 run diagnostic 保留相同 failure。

退出条件：六条路由 exact-limit / over-limit / provider failure / replay 测试通过，拒绝时 bridge / provider 调用为零。

## 批次 E：Pencil、Web 与产品验收

状态：已完成。

- 五维初评 `1 / 2 / 2 / 2 / 1 = 8`，按 `A / 完整 Pencil` 处理。
- 已在 Pencil 空闲后完成 Admin Quota Desktop / Narrow R1 与共享 Decision R14，并显式保存设计源。
- React 复用 S1–S8 Workbench 壳层、唯一选中语义、注意色和开发测试态边界。
- 自启动服务完成 `1440×900`、关键断点和 `390×844` 严格浏览器验收后关闭。

实际证据：Web `304/304`、production build、三视口零横向溢出和控制台零 warning / error 通过；memory_dev 连续链覆盖 missing policy、正整数创建、显式确认、并发 CAS 冲突、旧 owner 保留和 reload 权威版本。permission、environment、missing policy、version conflict 与 store failure 的严格失败 envelope 由 consumer 测试和现有后端 HTTP 测试共同覆盖。

退出条件：Pencil、Web 测试、production build、真实浏览器、控制台、横向溢出和交互通过。

## 验证入口

- 精准 Go 单元 / repository / HTTP / Gateway 测试
- SQLite 本地产品重启与并发验证
- PostgreSQL migration / role / repository / Gateway 集成
- Web `npm test` 与 `npm run build`
- `./scripts/check-repo.sh --fast`
- 本专题涉及协议、schema、架构和阶段真相，收口前运行完整 `./scripts/check-repo.sh`

## 停止线

- 不新增 production secret、OIDC membership、production API Key、quota inheritance 或 production enablement。
- 不实现 token quota、估算、price、cost、billing、rate limit、排队、retry / fallback 或自动路由。
- 不修改旧 quota fixture 冒充真实 owner，不从 Request History 分页窗口聚合总量。
- 不为本任务继续新增专项 task card、fixture 或 checker；测试、migration smoke、Web build 和仓库聚合门禁承载证据。
- 已冻结的 Pencil 只承载结构、交互、风险和响应式顺序；功能事实继续以专题、API 契约和代码为准，不复制完整状态画板。

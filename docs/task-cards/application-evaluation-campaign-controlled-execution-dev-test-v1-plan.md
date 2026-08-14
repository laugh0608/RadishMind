# 应用评测计划、受控执行与证据归档（开发 / 测试态）v1 实施任务卡

更新时间：2026-08-10

状态：`application_evaluation_campaign_controlled_execution_dev_test_v1_completed`

对应功能文档：[应用评测计划、受控执行与证据归档（开发 / 测试态）v1](../features/user-workspace/application-evaluation-campaign-controlled-execution-dev-test-v1.md)

## 任务目标

建立 development / test application-scoped evaluation plan 与 campaign owner，顺序复用四类现有执行服务产生 durable Run，并把同计划版本的 baseline / candidate campaign 显式交接到既有 Comparison、Evaluation Case、Suite 与人工 decision 链。

本任务卡是该专题唯一高风险实施卡。新增 schema、migration、权限、执行编排和跨 owner handoff 都在这里收口，不派生同层 readiness、refresh 或 review 卡。

## 实现边界

### 允许修改

- `services/platform/internal/httpapi/application_evaluation_*.go` 及相邻 owner / server 装配；
- workflow run store selector、SQLite / PostgreSQL workflow run migrations；
- `services/platform/internal/config/` 中的显式开发测试态 gate；
- 既有四类 invocation service 的可注入确定性 Run ID 或 authority snapshot 小范围接缝；
- 既有 Workflow Evaluation service 的受控复用入口；
- `apps/radishmind-web/` 中对应 strict consumer 与页面；
- 功能入口、Family UI 产品化专题、当前焦点、任务卡和本周周志。

### 不允许修改

- 既有 Run v5 / v4 / v6 / v7 schema 与 Comparison 语义；
- Case / Suite / decision schema 和已有 route 的行为；
- Gateway quota、provider route、Application Publish / Assignment 状态机；
- production auth、membership / OIDC、secret backend、billing 或 cost ledger；
- 自动发布、自动 activation、业务写回、工具执行扩权、schedule / replay / retry；
- Pencil 被其它项目占用时的任何设计源读取或修改；
- 没有证据证明现有测试与聚合门禁无法承载的新 checker / fixture。

## 冻结合同

1. Scope 固定为 `tenant_ref + workspace_id + environment + application_id`，environment 只允许 `development | test`。
2. 一个 immutable plan version 只允许一个 execution profile，最多 `20` 个有序 item。
3. fixture 由用户显式提交，持久化前执行严格结构、大小、UTF-8 与 secret material 校验；不得从 Run 恢复输入。
4. campaign 顺序执行，`failure_policy=stop_on_failure`，不并行、不自动 retry / resume / replay。
5. campaign 开始冻结 authority snapshot；每项执行前和 Run 落库后都必须证明 authority digest 未漂移。
6. 每项只保存确定性 Run ref 和 metadata，不复制输入、输出、prompt、context、artifact content 或 provider raw response。
7. baseline / candidate 必须引用同一 exact plan version；preview 复用 Comparison，materialize 显式复用 Case / Suite。
8. handoff 发生部分成功时返回和持久化 exact partial refs，不删除 append-only evidence，也不声称原子回滚。
9. API body 拒绝未知字段，query 严格 allowlist，cursor 与 version token 绑定 scope，mutation 使用 CAS。
10. memory / SQLite / PostgreSQL 行为等价；PostgreSQL 不自动迁移，SQLite 使用共享 runtime。
11. campaign 必须选择当前 actor 在当前 application 下仍为 active 的 API Key；gate 强制启用 quota enforcement。逐项确定性 Run ID 是 durable Run 根身份；单 provider 调用可直接使用该身份，多 LLM 节点 Workflow 必须按 Run 与 node 派生独立、确定性的 provider-attempt request identity。
12. handoff 只在 candidate campaign 保存；每个 exact Case version 成功后立即以 `partial` checkpoint，Suite 成功后推进 `complete`，partial 不自动续跑或补偿。

## 批次 A：正式设计与领域 owner

- [x] 审计现有 Run、Comparison、Case、Suite、四类 invocation service 与三模式装配。
- [x] 冻结功能文档、Profile / fixture / authority / permission / privacy / failure / handoff 合同。
- [x] 实现 plan current、immutable version、campaign 与 item 类型和 contract validator。
- [x] 实现 memory repository、CAS、幂等 campaign key、scope-bound cursor 和 fault injection 测试。
- [x] 实现服务层 create / revise / archive / read / list / version read / list。

批次 A 退出条件：领域记录不能持久化 secret 或运行输出；版本不可变；CAS、scope 隔离、cursor、幂等和 store failure 均有负向证据。

## 批次 B：HTTP 与 durable repositories

- [x] 增加显式 `ApplicationEvaluationCampaignDevEnabled` gate 和 sanitized config summary。
- [x] 实现 Plan / Version / Campaign strict HTTP、权限与 workspace / environment binding。
- [x] 增加 SQLite migration 与 repository，覆盖事务、restart read、CAS 和 corruption failure。
- [x] 增加 PostgreSQL migration 与 repository，覆盖 row lock、CAS、restart、rollback / reapply 和 scope 隔离。
- [x] 更新 store selector、server lifecycle 与 migration identity。

批次 B 退出条件：三模式行为一致；未迁移、错版本、环境不符、权限不足、冲突和 store failure 全部失败关闭，无 memory fallback。

## 批次 C：受控 campaign executor

- [x] 为四类 Profile 实现 fixture → 既有 invocation input 的显式映射。
- [x] 实现 profile-specific authority snapshot、campaign digest 和逐项 checkpoint。
- [x] 实现确定性 Run ID / client invocation key 并复用现有 Run owner。
- [x] 验证 API Key inference provider 前 quota admission 不被内部 campaign 调用绕过。
- [x] 实现 stop-on-failure、连接取消、store checkpoint failure 和 interrupted reconciliation。
- [x] 覆盖 application archived、assignment missing / revoked、definition drift、quota missing / exceeded / conflict / store unavailable。

批次 C 退出条件：每次 provider attempt 都能对应一个 exact campaign item 和既有 durable Run；authority 漂移与 quota 失败不会执行剩余项；重启不重放 provider。

## 批次 D：Comparison 与 Evaluation handoff

- [x] 实现 baseline / candidate campaign pair 校验和即时 Comparison preview。
- [x] 实现双 campaign expected-version 确认与逐项 Case materialize。
- [x] 把 exact case versions 组成既有 Suite，并保存 handoff refs / audit。
- [x] 覆盖跨 plan / version / scope / profile、非终态 Run、expected classification mismatch 和 partial handoff。

批次 D 退出条件：handoff 不复制 Comparison / Case / Suite 真相，不触发 publish / activation；部分成功有可审计 refs 和明确下一步。

## 批次 E：Pencil、React 与真实浏览器

- [x] 按五维评分 `8`、`A / 完整 Pencil` 冻结 Desktop `Um8Zh`、Narrow `ZxJd7` 与共享 Decision R15 `UNMOS`。
- [x] 复用 S1–S9 全视口 Workbench 实现 Plan、Campaign、Pair Review 与 Handoff 页面。
- [x] 实现 strict response validation、permission / environment / archived / authority drift / quota / conflict / store / interrupted / partial 状态。
- [x] 完成执行和 handoff 二次确认；普通行中性，只有当前详情 owner 选中。
- [x] Web tests、production build、三视口、memory 与 SQLite 浏览器链已通过；SQLite 完成 Plan → 两次 succeeded Campaign → Pair → exact Case / Suite Handoff，并在服务重启后恢复同一证据。
- [x] 已检查横向溢出、键盘可聚焦语义、凭据不回显、SQLite 持久记录和 console，并停止全部自启动开发服务。

批次 E 的设计、实现与 SQLite 重启复验均已完成，不以临时 React 页面替代设计基准面。任务卡关闭，不派生平行 task card 或同层 gate-only 切片。

## 验证矩阵

- Go domain / HTTP / repository / executor / handoff tests；
- SQLite migration、restart、corruption、CAS、deterministic Run reconciliation；
- PostgreSQL migration / rollback / reapply、runtime-role DDL rejection、并发与 restart；
- `go test -race` 定向包、`go vet`；
- Web strict consumer / interaction tests 与 production build；
- 真实 memory / SQLite 浏览器 campaign → pair preview → case / suite handoff；
- `git diff --check`、`./scripts/check-repo.sh --fast`；
- 因新增 API、schema、permission、migration 与阶段真相，完成前运行全量 `./scripts/check-repo.sh`。

## 完成定义

- 四个 Profile 都可从 exact immutable plan version 产生 durable campaign Run；
- 三种 store、CAS、cursor、idempotency、authority snapshot、quota 和 restart 证据齐全；
- 同计划版本的 campaign pair 可显式形成既有 Case / Suite refs；
- React 与真实浏览器覆盖完整状态、响应式、选中语义、交互和 console；
- production、billing、token / cost、自动路由、自动发布、retry / replay / schedule 继续关闭；
- 正式文档、Family UI 覆盖记录、current focus 和周志与代码一致。

## 当前阻塞

- 无。设计、实现、测试、memory / SQLite 浏览器链和服务重启恢复均已完成。
- 下一项工作回到功能设计入口选择新的真实用户需求；不得从已关闭任务卡扩 production、自动执行或新 owner。

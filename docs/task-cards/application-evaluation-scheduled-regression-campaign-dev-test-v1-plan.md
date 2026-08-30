# 应用定时回归评测与受控 Campaign 调度（开发 / 测试态）v1 实施任务卡

更新时间：2026-08-30

- 任务 ID：`application-evaluation-scheduled-regression-campaign-dev-test-v1`
- 状态：`batch_a_completed_batch_b_awaiting_approval`
- 功能设计：[应用定时回归评测与受控 Campaign 调度（开发 / 测试态）v1](../features/user-workspace/application-evaluation-scheduled-regression-campaign-dev-test-v1.md)

## 准入结论

项目所有者已于 2026-08-30 批准开发 / 测试态 `system actor + schedule-scoped delegation`：后台 runner 不冒用用户 actor，不保存 Session、cookie、Bearer Token、API Key token 或通用执行令牌；每个 occurrence 都必须重新读取当前账户、workspace membership、权限、API Key owner / lifecycle / scope 与 quota，失效时立即失败关闭。

本卡是该专题唯一高风险实施入口。Schedule、Version、Occurrence、HTTP、双数据库、runner 生命周期、Pencil 和 React 都在本卡分批收口，不派生平行 schema-only、UI-only、readiness 或 gate-only 任务卡。

## 完成目标

让内部 Prompt Application 开发者把一个 exact immutable Application Evaluation Plan version 配置为受限 `daily_utc` 周期，并在页面关闭后由 system runner 最多创建一个既有 Campaign。用户能审查授权主体、quota consumer、下次触发、Occurrence 与 Campaign refs，并显式暂停、恢复或归档；任何 authority、membership、API Key、quota、application、Plan 或 store 漂移都必须在 Provider 前失败关闭。

## 不可变边界

1. v1 只接受 `prompt_application_invocation_v1`，最多 `20` 个 item，`max_provider_attempts=item_count`。
2. Schedule owner 只保存调度意图、不可变版本、Occurrence claim 和既有 Campaign ref；不复制 Plan fixture、Run、Comparison、Case、Suite、decision 或 quota 状态。
3. 周期只允许 `daily_utc + hour + minute`；不接受 cron、时区名称、秒级周期、任意 interval、动态 payload 或用户代码。
4. 授权模型固定为 `system_actor_schedule_scoped_delegation_v1`；system actor 与 delegated user 分开记录，required permissions 固定且有序。
5. 委托不可转移、不可作为 bearer；角色或 permission snapshot 不能成为长期授权，Schedule version 变化、撤权、账户停用、membership 失效、API Key 轮换或失效都必须立即重新判定。
6. Occurrence 身份固定为 `schedule_id + schedule_version + scheduled_for_utc`；`client_campaign_key` 必须确定性派生。
7. 单赢家 claim 后不释放 lease、不自动重领；claim 后、Campaign 创建前崩溃时只能查 exact key，找不到则 `interrupted`，不得重放 Provider。
8. 状态只允许 `due -> claimed -> campaign_created -> observing -> terminal`；`claimed` 可在 Provider 前进入 `failed | interrupted | skipped`，terminal 不可返回非终态。
9. missed window 和 overlap 只允许形成 `skipped` 证据，不 catch up；其它权限、quota 或 authority 失败不能伪装为 skipped。
10. memory / SQLite / PostgreSQL 必须同构；数据库未迁移、损坏或不可用时失败关闭，不回退 memory。
11. Runner 默认关闭，只能在显式开发测试 gate 下启动；`production` 永远拒绝。
12. 不自动 retry、fallback、resume、replay、baseline、Case / Suite materialize、decision、release、deploy 或业务写回。

## P0：受限委托授权模型

状态：`completed_owner_approved`。

- system runner 使用独立 `system_actor_ref`，审计和 Occurrence 另存 `delegated_by_user_ref`。
- 委托只属于一个 Schedule，不可转移、不可导出、不可复用于其它 application / Plan / operation。
- 每个 occurrence 重新检查本地账户 active、workspace membership、`application_evaluations:execute + workflow_runs:execute`、API Key 当前 owner / active / expiry / application scope 与 quota。
- pause、archive、Schedule version 变化、账户 / membership / permission / API Key 变化立即使后续执行失效。
- 禁止保存用户 HTTP auth 上下文、Session、cookie、Bearer Token、API Key token、Provider credential 或通用 execution token。

## 批次 A：正式合同、领域与 memory owner

状态：`completed`。

- [x] 新增 `application_evaluation_schedule.v1`、`application_evaluation_schedule_version.v1` 与 `application_evaluation_schedule_occurrence.v1` canonical JSON Schema。
- [x] 实现 Prompt-only Plan / quota binding、固定 `daily_utc`、exact attempt budget、授权引用、canonical UTC 和 schedule digest validator。
- [x] 实现 Schedule current / immutable version、draft / active / paused / archived CAS 与 UTC next-due projection。
- [x] 实现 Occurrence deterministic key、单赢家 system claim、严格状态机、terminal no-replay 与 sanitized failure allowlist。
- [x] 实现独立 memory repository；用户 actor 只能管理 Schedule 生命周期，system actor 只能 claim / 推进 projection / Occurrence。
- [x] 覆盖不可变版本、stale CAS、并发 claim、错误 actor、permission order drift、secret-shaped ref、Profile / budget 漂移、missed skip、terminal replay、corruption 和 store outage。

批次 A 没有 route、migration、selector、server 装配、runner、goroutine、Provider、Pencil 或 React。它只证明 canonical owner 与受限委托的领域可执行性，不声明后台调度已经可用。

## 批次 B：strict HTTP 与双数据库

状态：`pending_owner_approval`。

- 实现 create / revise / activate / pause / resume / archive / list / exact version read 与 Occurrence read；body 拒绝 unknown / duplicate field，query 使用严格 allowlist，mutation 使用 CAS。
- 激活必须重新读取 exact active Prompt Plan version、digest、item count、Prompt assignment authority 与 actor-owned API Key，并显式确认周期 Provider 消耗。
- 增加 SQLite / PostgreSQL migration、marker / checksum、runtime role、transaction、唯一 occurrence、并发 claim、restart / reconnect、corruption 与 no-fallback。
- 复用既有数据库连接、selector 和 migration family；不新增 DSN、pool、database file 或跨 store join。
- 本批仍不启动 runner、不调用 Campaign / Provider、不修改 Pencil / React。

批次 B 退出条件：HTTP 与三种 store 行为同构；错误 scope、environment、permission、Plan、authority、API Key、CAS、migration 和 store 状态全部失败关闭。

## 批次 C：Runner 与既有 Campaign 交接

状态：`pending`。

- 增加显式 development / test gate、固定低频 poll、单 worker cancel / join 和“先停 worker、后关 store”的 Server 生命周期。
- 每个 occurrence 以 system actor 执行，同时携带独立 delegated user；逐次重读账户、membership、permission、Plan / assignment、API Key 与 quota。
- 只调用既有 `applicationEvaluationCampaignService.Execute` 一次，不调用 HTTP handler、不创建第二套 Campaign executor。
- 覆盖 crash before / after Campaign create、exact-key reconciliation、pause / archive / revoke、missed / overlap、quota / authority drift、store reconnect 与零重复 Provider。

## 批次 D：完整 Pencil、React 与产品验收

状态：`pending`。

- 按 `A / 完整 Pencil` 复用 S10 Application Evaluation Workspace，先取得项目所有者人工批准再实现 React。
- 实现 Schedule / Version / Occurrence strict consumer、生命周期确认、Campaign handoff 和完整失败态。
- 完成三视口、双标签 CAS、SQLite 重启、PostgreSQL no-fallback / reconnect、console / URL / storage / cookie / database 隐私与副作用审计。

## 验证矩阵

- Contract：三份 JSON Schema、Go validator、digest、UTC、permission order、nullable state fields、unknown / secret material。
- Domain：Schedule CAS / immutable version / lifecycle、Occurrence identity / transition / failure allowlist / no replay。
- Concurrency：单赢家 claim、stale expected version、capacity、store unavailable、corruption fail-close。
- Persistence：memory / SQLite / PostgreSQL、migration、runtime role、restart / reconnect、unique occurrence、no fallback。
- Runner：cancel / join、multi-runner claim、authorization recheck、Campaign reconciliation、zero duplicate Provider。
- Web / Product：strict consumer、Pencil approval、three viewports、two-tab stale、privacy、exact Campaign / Run refs。
- Repository：每批精准测试、`go test -race`、`go vet`、`git diff --check`、fast baseline；新增 schema / migration / 阶段边界时补全量 baseline。

## 当前下一步

Batch A 已完成。下一步只在项目所有者批准 Batch B 后进入 strict HTTP 与 SQLite / PostgreSQL；不得提前启动后台 runner、调用 Campaign / Provider、修改 Pencil / React 或声明定时回归可用。

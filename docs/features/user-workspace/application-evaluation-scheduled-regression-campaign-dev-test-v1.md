# 应用定时回归评测与受控 Campaign 调度（开发 / 测试态）v1

更新时间：2026-08-30

状态：`application_evaluation_scheduled_regression_campaign_dev_test_v1_design_reviewed_authorization_model_required`

## 设计结论

本专题选择“定时回归评测”作为 Action Safety Ladder 关闭后的下一条长期产品目标，但当前只完成产品设计与实现准入评审，不创建任务卡、不进入代码实现。

现有 Application Evaluation Plan 已保存不可变测试 fixture，Campaign 已能顺序调用既有应用运行服务并生成 durable Run，Comparison、Case、Suite 与人工 decision 也已闭合。新的用户价值不在复制这些 owner，而在让内部应用开发者把一个 exact Plan version 配置为受限 UTC 周期，在无人停留页面时仍能形成可审查的 Campaign 证据。

代码审计同时确认：现有 Campaign 只能在用户 HTTP 权限上下文中同步执行，API Key 必须属于当前 actor；Platform 没有后台 scheduler 生命周期、service principal、受限委托或 occurrence lease。后台进程若直接把创建者 `actor_ref` 填入请求，将构成隐式冒用。该授权模型属于新的跨模块高风险边界，必须由项目所有者明确选择后才能进入实现。

## 用户任务

1. 在当前 application 下选择一个 active Application Evaluation Plan 的 exact immutable version。
2. 配置一个受限 UTC 日周期，看到下次触发时间、fixture 数、最大 Provider attempt 数和当前 quota consumer。
3. 显式审查 application、Plan digest、Prompt assignment authority、API Key、workspace membership 与自动消耗配额的影响后启用 schedule。
4. 到期时由受控 runner 原子 claim 一个 occurrence，并最多创建一个既有 Campaign；页面关闭或服务重启不能制造重复 Campaign。
5. authority、membership、API Key、quota、application lifecycle 或 store 漂移时在 Provider 前失败关闭，并留下脱敏 occurrence 证据。
6. 用户查看 schedule、occurrence 和 exact Campaign / Run refs，显式暂停、恢复或归档 schedule。
7. Campaign 结果继续由既有 Comparison、Case、Suite 和人工 decision owner 审查；schedule 不自动选择 baseline、不自动批准或发布。

## 选题与首版收窄

四个产品面评审后的结论如下：

| 候选 | 结论 | 原因 |
| --- | --- | --- |
| 通用 Workflow schedule | 后置 | 会同时打开持久运行输入、Tool / RAG、后台执行和多 Profile authority，首版边界过宽 |
| 受控在线知识源 | 后置 | 现有 HTTP Tool 只接受 `application/json`；在线 Markdown / HTML 需要新的 transport、解析与来源安全边界 |
| 真实 Radish OIDC | 等待外部条件 | reviewed client registration、issuer / audience 与真实联调环境仍未成立 |
| production secret / billing | 不进入 | production owner、资源与发布条件未满足 |
| Prompt Application 定时回归 Campaign | 选中但实现待授权 | fixture、单次 Provider 调用、Run 与评测证据链已经存在，能形成最小可验证纵向闭环 |

v1 只接受 `prompt_application_invocation_v1` Plan version。每个 item 最多产生一次 Prompt Provider attempt，最多 `20` 个 item；Workflow Definition、RAG、Agent、HTTP Tool、结构化多 LLM Definition 和其它 Profile 均不在首版范围内。该收窄让请求配额预算可以由 `item_count` 精确表达，不把估算成本或未知节点数写成授权事实。

## 现有事实审计

| 事实 | 当前 owner / 实现 | 对本专题的影响 |
| --- | --- | --- |
| immutable fixture | `application_evaluation_plan_version.v1` | schedule 只保存 exact ref 与 digest，不复制 fixture |
| Campaign execution | `applicationEvaluationCampaignService.Execute` | runner 只能调用同一服务，不创建第二套 Campaign executor |
| actor / permission | `ApplicationEvaluationContext.ActorRef` 与 HTTP auth | 仅能表达交互式用户 actor，不能表达 system actor + delegated user |
| quota consumer | 当前 actor-owned active API Key | occurrence 必须重新验证 owner、lifecycle、scope 与 quota，不保存 token |
| idempotency | `client_campaign_key` | occurrence 可派生稳定 key，但仍需要 durable claim 防止多实例竞争 |
| reconciliation | existing interrupted Campaign reconcile | 只观察既有 Run，不重放 Provider；可被 occurrence 恢复流程复用 |
| Server lifecycle | `Server.Close` 只关闭 bridge / stores | 新 runner 必须具有 cancel、join、先停 worker 后关 store 的明确顺序 |
| storage modes | memory / SQLite / PostgreSQL | schedule / occurrence 必须同构且数据库失败不回退 memory |

## 领域所有权

| 领域 | owner | 允许保存 | 禁止保存 |
| --- | --- | --- | --- |
| Schedule current / version | 新 `Application Evaluation Schedule` owner | exact Plan ref / digest、周期、quota API Key id、lifecycle、授权引用、next due projection | fixture、API Key token、Provider credential、Run payload |
| Occurrence / claim | 同一 Schedule owner | exact schedule version、scheduled time、claim / terminal state、Campaign ref、稳定 failure | 输入、输出、Provider raw response、任意 retry payload |
| Plan / fixture | 既有 Application Evaluation Plan owner | 保持现状 | schedule 不复制或修改 |
| Campaign | 既有 Application Evaluation Campaign owner | 保持现状 | schedule 不复制 item / Run 状态机 |
| Run / quota / route | 既有 Run 与 Gateway owner | 保持现状 | schedule 不预扣额度、不选择 Provider |
| Comparison / Case / Suite / Decision | 既有 Evaluation owner | 保持现状 | 不自动 materialize、approve 或 release |
| identity / membership | 既有 local identity 与 workspace membership owner | occurrence 时重新读取 | schedule 不保存 role / permission snapshot 作为长期授权 |

Schedule owner 只能拥有调度意图、版本、claim 和交接引用。它不能成为第二套 Plan、Campaign、Run、quota、Action Safety 或审计真相源。

## 拟议 v1 合同

### Schedule current record

`application_evaluation_schedule.v1` 计划保存：

- `schedule_id`、`record_version`、`latest_schedule_version`；
- 完整 `tenant_ref + workspace_id + environment + application_id`；
- exact `plan_id + plan_version + plan_digest`；
- `execution_profile=prompt_application_invocation_v1`；
- `quota_api_key_id`，只保存 id；
- `lifecycle_state=draft|active|paused|archived`；
- `next_due_at` 只作为可重算投影；
- create / update / actor / request / audit refs。

current record 不保存 fixture、权限快照、cookie、session、bearer token 或 secret。

### Immutable schedule version

`application_evaluation_schedule_version.v1` 计划冻结：

- exact Plan identity 与 digest；
- `daily_utc` 规则，只包含 `hour` 与 `minute`；
- `item_count` 与 `max_provider_attempts=item_count`；
- `missed_window_policy=record_only_no_catch_up`；
- `overlap_policy=skip_while_campaign_non_terminal`；
- stable `schedule_digest`；
- 激活前要求重新读取的 authority / membership / API Key 条件。

v1 不接受 cron 表达式、时区名称、秒级周期、任意 interval、节假日、动态 payload 或用户代码。

### Occurrence

`application_evaluation_schedule_occurrence.v1` 的唯一键固定为：

```text
schedule_id + schedule_version + scheduled_for_utc
```

状态只允许：

```text
due -> claimed -> campaign_created -> observing -> succeeded|failed|interrupted|skipped
```

- claim 必须由 repository 原子完成；SQLite 单写事务和 PostgreSQL 行锁 / 唯一约束提供单赢家。
- `client_campaign_key` 由 occurrence identity 确定性派生；相同 occurrence 永远不能创建第二个 Campaign。
- 服务在 claim 后、Campaign 创建前崩溃时，恢复只能查找同 key Campaign；不存在则标记 `interrupted`，不得重新执行。
- 旧过期窗口只记录 `skipped / missed_window`，不 catch up。
- 当前 schedule 已有非终态 Campaign 时，新窗口记录 `skipped / overlap_blocked`。
- schedule pause / archive、application archive、Plan drift、assignment drift、API Key revoke、membership deny、quota deny 或 store unavailable 均不调用 Provider。

## 授权模型准入问题

后台 occurrence 同时涉及“谁授权了周期执行”和“当前由谁执行”。现有 `ActorRef` 不能安全表达这两层身份。v1 实现前必须在以下边界上获得项目所有者批准：

1. 是否允许建立 schedule-scoped、不可转移、非 bearer 的受限委托记录；
2. system runner 是否使用独立 `system_actor_ref`，并另存 `delegated_by_user_ref`；
3. 每次 occurrence 是否强制重读本地账户 active 状态、当前 workspace membership 和 `application_evaluations:execute + workflow_runs:execute`；
4. API Key 是否必须继续属于委托用户，并在每次 occurrence 重新检查 active / expiry / application scope；
5. 用户失去权限、账户停用、membership 撤销、API Key 轮换或 schedule version 变化时，委托是否立即失效；
6. HTTP、领域上下文、audit 和 Run metadata 如何同时记录 system actor 与 delegated user，而不把委托伪装成用户交互请求。

明确禁止的方案：持久化 Web Session / cookie / API Key token、后台直接冒用创建者 `actor_ref`、只检查激活时权限、把角色或 permission snapshot 当永久授权、引入可在其它资源复用的通用 execution token。

在该问题获得批准前，Schedule schema、route、migration、repository、runner、Pencil 和 React 均不进入实现。

## Runner 生命周期与失败语义

若授权模型获批，runner 仍须满足：

- 只在显式开发测试 gate 下启动；默认关闭，`production` 永远拒绝。
- 启动顺序为 store / migration preflight → runner；关闭顺序为停止接收 claim → cancel / join worker → 关闭 store / bridge。
- 固定低频 poll，不为每个 schedule 创建 goroutine；多实例通过 repository claim 收敛。
- occurrence 只调用既有 Campaign service 一次，不调用 HTTP handler，也不绕过 authority、membership、API Key 或 quota 校验。
- store unavailable 时停止 claim 且不回退 memory；恢复后只观察已存在 occurrence / Campaign。
- 不自动 retry、fallback、resume、replay、catch-up 或补偿 Provider 调用。

稳定失败至少包括：

- `application_evaluation_schedule_authorization_unavailable`；
- `application_evaluation_schedule_membership_denied`；
- `application_evaluation_schedule_plan_changed`；
- `application_evaluation_schedule_authority_changed`；
- `application_evaluation_schedule_quota_consumer_invalid`；
- `application_evaluation_schedule_quota_denied`；
- `application_evaluation_schedule_overlap_blocked`；
- `application_evaluation_schedule_missed_window`；
- `application_evaluation_schedule_claim_conflict`；
- `application_evaluation_schedule_store_unavailable`；
- `application_evaluation_schedule_store_contract_mismatch`。

## Web 与设计覆盖

本专题若进入实现，应复用 S10 Application Evaluation Workspace，在现有 Plan / Campaign / Handoff 任务模型中增加 schedule owner，而不建立 S11。

五维预评估为 `1 / 2 / 2 / 1 / 2 = 8`，达到 `A / 完整 Pencil`：周期启用属于持续 Provider 副作用授权；需要表达 exact Plan、quota consumer、授权主体、下次触发、暂停 / 归档、occurrence / Campaign handoff、权限失效和服务重启状态。Pencil 必须在 React 前由项目所有者人工批准，但当前授权模型未批准，因此暂不修改设计源。

## 拟议实施顺序

### P0：授权模型决策

- 项目所有者选择 system actor + schedule-scoped delegation 的边界，或明确拒绝后台委托。
- 若拒绝，不把专题降级为“页面打开时自动运行”或伪 schedule；专题保持 blocked。
- 只有 P0 通过后才创建唯一高风险任务卡。

### 批次 A：合同、领域与 memory owner

- 冻结 schedule / version / occurrence schema、UTC 计算、digest、状态机与授权引用。
- 实现单赢家 claim、missed / overlap、无 retry 和 memory repository。

### 批次 B：strict HTTP 与双数据库

- 实现 create / revise / activate / pause / archive / list / exact read 与 occurrence read。
- 增加 SQLite / PostgreSQL migration、CAS、唯一 occurrence、restart、multi-runner claim 和 no-fallback。

### 批次 C：runner 与既有 Campaign 交接

- 实现 cancel / join 生命周期、低频 poll、当前授权重读和 deterministic Campaign key。
- 覆盖 crash before / after Campaign create、reconnect、quota / authority drift 和零重复 Provider 调用。

### 批次 D：Pencil、React 与产品验收

- 先完成人工批准的 S10 完整设计，再实现单一 strict consumer。
- 验证三视口、双标签 CAS、SQLite 重启、PostgreSQL no-fallback / reconnect、console / URL / storage 隐私与零自动 release。

## 当前准入结论

| 准入项 | 结论 |
| --- | --- |
| 真实用户任务 | 满足：周期回归发现模型 / Profile 漂移 |
| canonical Plan / Campaign / Run owner | 满足：现有链可复用 |
| fixture 与 Provider attempt 预算 | 首版 Prompt-only 后满足 |
| durable schedule / occurrence owner | 可设计，尚未实现 |
| background service lifecycle | 设计可行，尚未实现 |
| system actor / delegated authorization | **未满足，必须由项目所有者决定** |
| implementation task card | blocked |

当前状态保持 `design_reviewed_authorization_model_required`。下一步不是写代码，而是项目所有者审查 P0：是否允许开发测试态 system runner 使用 schedule-scoped、每次重验的受限委托。未经批准，不创建执行授权、任务卡或运行时。

## 停止线

- 不实现通用 Workflow scheduler、任意 cron、queue platform、job framework 或 production worker。
- 不支持 Workflow Definition、RAG、Agent、HTTP Tool、write action、connector、sandbox、replay 或业务写回。
- 不保存 Web Session、cookie、API Key token、Provider credential、fixture 副本、输入、输出或 raw response。
- 不冒用用户 actor，不以激活时 permission snapshot 替代 occurrence 时重新授权。
- 不自动 retry、fallback、catch-up、resume、replay、baseline、Case / Suite materialize、decision、release 或 deploy。
- 不创建第二套 Plan、Campaign、Run、quota、route、Action Safety、Comparison、Case、Suite 或 decision owner。
- 不声明 production authentication、production quota、billing、cost limit、process supervisor 或 production readiness。

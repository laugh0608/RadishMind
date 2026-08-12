# 应用评测计划、受控执行与证据归档（开发 / 测试态）v1

更新时间：2026-08-11

状态：`application_evaluation_campaign_controlled_execution_dev_test_v1_completed`

## 功能定位

本专题为内部应用开发者提供一条可复验的应用回归执行链：把经过显式审查的评测输入冻结为不可变计划版本，在 `development | test` 环境内按单一执行 Profile 顺序产生 durable Run，再把两个 campaign 的同项 Run 精确交接给既有 Comparison、Evaluation Case、Suite 与人工 decision owner。

它解决的是当前真实断点：现有 Run History、Comparison、Evaluation Case 和 Suite 可以审查已经存在的 Run，却不能保存一组受控评测输入、重复产生同构 Run、证明一批 Run 使用同一 authority snapshot，也不能把逐项 Run 安全地组织成可供现有评测 owner 消费的证据集合。

本专题不创建第二套 Run、Comparison、Case、Suite 或 release decision。执行仍由现有四类应用运行服务完成，评测 owner 只保存计划、执行批次、逐项状态和既有资源引用。

## 用户任务

1. 在当前 application 下创建评测计划，选择一个真实执行 Profile，并录入一组明确用于评测的输入。
2. 保存计划版本；每次修改通过 `expected_version` CAS 生成新的不可变版本，历史版本不被覆盖。
3. 对 exact plan version 发起 campaign，选择当前 actor 在该 application 下仍为 active 的 API Key 作为 quota consumer，并确认环境、项目范围、项目数、当前 authority 摘要和会消耗的 UTC 日请求配额。
4. 服务顺序执行计划项，每项只调用对应现有执行服务一次，并保存 durable `run_id`、状态、失败边界和 authority digest；不保存 provider 原始响应或运行输出。
5. authority 在批次中漂移、应用归档、权限不足、quota admission、store failure 或运行服务合同不匹配时失败关闭，不切换到其它 candidate、assignment、definition、provider 或 model。
6. 选择同一 plan version 的 baseline campaign 与 candidate campaign，读取逐项兼容性和既有 Comparison 结果。
7. 显式确认后，把逐项 exact Run refs 交接给既有 Evaluation Case，再把 exact case versions 组成既有 Suite；交接只生成审查证据，不触发发布或运行时切换。

## 范围与环境

唯一资源作用域为：

```text
tenant_ref + workspace_id + environment + application_id
```

- `environment` 只允许 `development | test`，且必须与当前服务进程声明的 provider route / quota environment 一致；`production` 永远拒绝。
- plan、plan version、campaign、campaign item 和 handoff 必须保持同一完整作用域。
- application lifecycle、application kind 与执行 Profile 必须在保存计划和每次执行前重新读取。
- 跨 tenant、workspace、environment 或 application 的引用不返回存在性细节，也不回退到其它资源。

## 领域所有权

| 领域 | owner | 本专题行为 |
| --- | --- | --- |
| 评测计划 | Application Evaluation Plan owner | 保存当前记录、不可变版本、输入 fixture 与 CAS 修订 |
| 执行批次 | Application Evaluation Campaign owner | 保存 exact plan version、authority snapshot、逐项执行状态和 Run refs |
| 应用调用 | 既有 Definition / RAG / Prompt / Agent invocation service | 继续负责 authority checkpoint、provider 前 quota admission、Run 写入和终态 |
| Run | 既有 Workflow / Application Run store | 继续是运行状态、usage、failure 和 side effect metadata 的唯一真相源 |
| Comparison | Workflow Run Comparison | 按既有 Profile 兼容矩阵即时比较 exact Run refs |
| Case / Suite / Decision | Workflow Evaluation owner | 接收显式确认后的 exact refs，不被 campaign owner 复制 |
| Provider / quota | Gateway owner | 继续负责 route selection、provider attempt 与原子 quota admission |

## 资源模型

### Plan current record

`application_evaluation_plan.v1` 保存：

- `plan_id`、`record_version`、`latest_plan_version`；
- 完整作用域、名称、执行 Profile；
- `item_count`、当前版本 digest、创建 / 更新时间；
- actor、request、audit refs；
- lifecycle 固定为 `active | archived`，v1 不提供删除。

Plan current record 不重复保存 fixture 正文。读取具体输入必须读取 exact immutable version。

### Immutable plan version

`application_evaluation_plan_version.v1` 保存：

- `plan_id + plan_version`、`previous_plan_version`；
- `execution_profile` 和 profile-specific target；
- 有序 items；
- canonical `plan_digest`；
- 创建 actor、request、audit refs 和时间。

每个 item 具有稳定 `item_key`、名称、预期 Comparison classification 和一种 profile-specific fixture。顺序属于版本合同；同一版本内 `item_key` 唯一。

### Campaign

`application_evaluation_campaign.v1` 保存：

- `campaign_id`、`client_campaign_key`、exact `plan_id + plan_version + plan_digest`；
- 完整作用域、执行 Profile；
- 只用于 quota 归属的 `quota_api_key_id`，不保存 credential token；
- 发起时冻结的 authority snapshot 与 `authority_digest`；
- `pending | running | succeeded | failed | interrupted` 状态；
- 当前 item index、成功 / 失败计数、失败码和失败摘要；
- 有序 campaign items；
- 创建、开始、完成、actor、request 和 audit refs。

每个 campaign item 只保存 `item_key`、确定性 `run_id`、运行状态、Run schema / profile、failure code / boundary、authority digest 和时间。不得保存输入正文、输出、回答、context、artifact content、prompt 或 provider raw response。

`client_campaign_key` 在完整作用域内提供幂等创建；同 key 携带不同 plan version 或 digest 时返回冲突，不生成第二批运行。

## Profile 与 fixture 合同

一个 plan version 只允许一个执行 Profile，避免一个 campaign 混合不同 authority owner。

| 执行 Profile | application / target | fixture | 既有 Run |
| --- | --- | --- | --- |
| `workflow_definition_executor_v1` | workflow application + exact definition id / pointer version / definition version / digest | `input_text`、`condition_values`、可选 model / temperature | v5 |
| `workflow_definition_executor_v2` | workflow application + exact definition / pointer / input contract id 与 digest | `inputs`；按不可变合同校验的扁平 typed fixture | v8 |
| `application_rag_invocation_v1` | RAG-capable application + active RAG assignment | `input` | v4 |
| `prompt_application_invocation_v1` | `prompt_application` + active Prompt assignment | `variables` | v6 |
| `agent_copilot_suggestion_v1` | `agent` + active Agent assignment | `task`、`locale`、可选 conversation id、artifacts、context | v7 |

fixture 是开发者主动提交的测试材料，不从历史 Run 反推。服务复用各执行 Profile 的既有大小、UTF-8、结构、secret material 和输出合同校验；计划 owner 还必须在持久化前拒绝 credential、token、Authorization header、cookie、DSN、私钥和明显 endpoint secret。

v1 不保存图片二进制、文件上传、外部 URI 抓取或任意 connector 凭据。Agent artifact 只允许现有 canonical contract 可接受的结构化、无 secret 测试材料。

## Authority 一致性

campaign 开始前读取一次目标 authority 并生成 profile-specific snapshot：

- Definition：definition id / version / digest、activation pointer version、application record version；
- RAG：runtime assignment id / version / digest、publish candidate、draft、binding、application record version；
- Prompt：assignment id / version / digest、candidate、draft、template、protocol / model eligibility、application record version；
- Agent：assignment id / version / digest、candidate、draft、profile、policy、protocol / model eligibility、application record version。

每项执行前重新读取 authority；摘要必须与 campaign snapshot 一致。现有执行服务仍进行 provider 前 checkpoint。Run 落库后，campaign 再校验 Run 内 authority metadata 与 campaign snapshot 一致，任一不一致立即终止剩余项。

campaign 不锁住发布或 assignment owner，也不通过长事务包住 provider 调用；它通过 snapshot、逐项 checkpoint 和失败关闭证明一致性。

## 执行语义

- v1 仅支持显式同步、顺序执行，最大 `20` 项。
- `failure_policy` 固定为 `stop_on_failure`；不提供并行、continue-on-error、自动重试、fallback、resume 或 replay。
- 每项执行前先持久化确定性 `run_id` 和 `running` 状态，再调用既有服务；Prompt / Agent 使用由 campaign 生成的 `client_invocation_key`，Definition / RAG 注入相同确定性 Run ID。
- 专题 gate 要求 API Key lifecycle、API Key dev/test auth 与 Gateway quota enforcement 已显式启用，且 campaign environment 与 quota environment 一致；未满足时服务拒绝启动该能力。
- HTTP 发起前重读当前 actor-owned API Key；key 不存在、跨 application、已撤销、已过期或 store 不可用均失败关闭。服务只保存 `api_key_id`，不读取或返回 credential token。
- 每个 item 将所选 `api_key_id`、确定性 `run_id`、完整 quota scope 与既有 inference route 注入 Gateway bridge。`run_id` 保持 durable Run 根身份；单 provider 调用可直接使用该根身份，多 LLM 节点 Workflow 则按 `run_id + node_id` 派生确定性的 provider-attempt request identity，保证同一 Run 内每次真实 provider attempt 独立准入。provider 调用仍在现有原子 admission 之后发生，campaign 不预留或绕过额度。
- quota exhausted、policy missing、attempt conflict 和 quota store failure 作为 item / campaign 失败证据保存，剩余项不执行。
- HTTP 连接取消不会触发后台继续运行。若进程中断，重启 reconciliation 只根据确定性 Run ID 收口 campaign 状态，不重放 provider 调用。

## Comparison 与 Evaluation 交接

- baseline 与 candidate campaign 必须引用同一 `plan_id + plan_version + plan_digest`、完整作用域和 Profile。
- 每个 item 按稳定 `item_key` 配对，Run 必须存在、终态、Profile 匹配且满足既有副作用限制。
- preview 即时调用既有 Comparison，不在 campaign owner 复制 finding 或分类真相。
- handoff 必须由用户显式确认，使用两个 campaign 的 expected record version 防止并发漂移。
- 每个 item 创建一个既有 Evaluation Case：baseline Run 为 baseline campaign item，candidate Run 为 candidate campaign item，expected classification 来自 immutable plan version。
- candidate campaign 是 pair handoff 的唯一持久化锚点；baseline / candidate 的 expected record version 都要先匹配，避免以漂移 pair 创建证据。
- 每创建一个 Case 就把 exact `case_id + version` checkpoint 为 `partial`；全部 Case 成功后才创建 Suite 并把同一锚点推进为 `complete`。candidate campaign 只保存 exact case refs、suite id 和 handoff audit，不复制 review digest 或 decision。
- 任一 Case / Suite 创建失败时停止，不声称原子回滚已经进入 append-only owner 的证据；已成功资源作为 partial handoff 明示返回，允许人工审计，不自动补偿或删除。

## 权限

- 读取 plan / version / campaign：`application_evaluations:read`。
- 创建、修订、归档 plan：`application_evaluations:write`。
- 发起或 reconcile campaign：`application_evaluations:execute + workflow_runs:execute`，Definition 额外要求 `workflow_definitions:read`。
- preview / handoff：`application_evaluations:read + workflow_runs:read`；materialize 额外要求 `workflow_evaluations:write`。
- 所有 mutation 继续经过 workspace membership 与 resource binding 校验；dev signed header 只用于开发测试态，不宣称 production membership / OIDC。

## API 轮廓

统一位于 `/v1/user-workspace/applications/{application_id}/evaluation-*`：

- Plan：create、list、read、revise、archive、version list / read；
- Campaign：execute、list、read、reconcile；
- Review：`POST .../evaluation-campaign-pairs/preview` 做 baseline / candidate 即时比较，`POST .../evaluation-campaign-pairs/handoff` 显式 materialize exact Case / Suite refs。

所有 JSON body 拒绝未知字段，query 使用严格 allowlist，list 使用 scope-bound cursor，mutation 使用 `expected_version`。失败 envelope 必须返回稳定 `failure_code`、`failure_summary`、当前版本和 `audit_ref`；权限、环境、scope、store contract 和不存在均失败关闭。

## 持久化

三种模式保持同一合同：

- `memory_dev`：有界内存 owner，服务重启不承诺恢复；
- `sqlite_dev`：共享 local persistence runtime，新 migration，事务化 plan revision / campaign checkpoint / handoff refs；
- `postgres_dev_test`：独立 migration，行锁与 CAS，显式 migration 前置，不自动迁移。

数据库只保存经过合同校验的 plan / version / campaign JSON 和必要索引列。SQLite / PostgreSQL 都必须覆盖 restart read、并发修订、campaign 幂等、scope 隔离、部分写失败和 interrupted reconciliation。

## Web 产品面

该页面族包含新的结构性任务模型，按五维评分预计为 `8`，覆盖级别为 `A / 完整 Pencil`：

- Plan 列表、版本与 fixture 编辑；
- authority / environment / quota 边界确认；
- campaign 逐项执行进度与 Run refs；
- baseline / candidate 配对、Comparison preview 与显式 handoff；
- permission、archived application、environment mismatch、authority drift、quota、version conflict、missing policy、store failure、interrupted / partial handoff 状态；
- Desktop、关键断点与 `390×844` 的响应式顺序、选中语义和无横向溢出。

2026-08-10 已在用户确认设计源空闲后完成 `S10` Desktop、Narrow 与共享 Decision R15，并据此实现 React 功能纵向切片；但首版 R1 后续因 dashboard 式进度卡、稀疏工作面和页面骨架没有继承 S1–S8 被人工退回。Visual R2 在相同根节点恢复 `264px` 产品导航、薄页眉、evidence path、selected campaign 单一 owner、连续 item rows 和单一 handoff rail 后，又因业务表面全部硬方形、棱角语言与 S1–S8 不一致被人工退回。Visual R3 已保留连续信息结构，并按 `8–11px` 职责圆角修订任务、上下文、owner、boundary 与 handoff 表面，于 2026-08-12 完成人工视觉复核。功能契约、strict consumer 与真实数据库证据不因视觉修订失效，但 React 暂不声明逐项采用 Visual R3。

普通 plan、campaign 和 item 行保持中性；只有驱动当前详情的对象使用墨蓝选中轨。failed、quota exceeded、interrupted、partial 或 blocked 只使用文字、图标与状态色，不冒充选中。

## 实施批次

### 批次 A：设计合同与领域 owner

- 冻结 Profile / fixture / authority / failure matrix、API、权限、隐私和 handoff 语义。
- 实现 plan current + immutable version、campaign 领域合同、memory repository 和领域测试。

### 批次 B：HTTP 与三模式持久化

- 实现 strict HTTP、workspace binding、权限与 environment gate。
- 增加 SQLite / PostgreSQL migration、repository、CAS、cursor、restart 和并发测试。

### 批次 C：受控 campaign executor

- 复用四类现有执行服务和统一 Run owner。
- 完成 authority snapshot / checkpoint、确定性 Run ID、quota failure、stop-on-failure 与 interrupted reconciliation。

### 批次 D：Comparison / Case / Suite handoff

- 实现 campaign pair preview。
- 显式确认后交接既有 Case / Suite，保留 partial handoff 真实状态和 exact refs。

### 批次 E：Pencil、React 与浏览器验收（已完成）

- Desktop `Um8Zh`、Narrow `ZxJd7` 和共享 Decision R15 `UNMOS` 的 R1 与 Visual R2 均已被人工视觉退回；Visual R3 已显式保存，完成实际截图、零折叠、零硬编码色、零 placeholder 检查和 2026-08-12 人工视觉复核。历史 2× PNG 只属于 R1 证据，不再作为当前视觉基准。
- 已实现 strict consumer、四任务单 owner Workbench、完整失败关闭、Plan / Campaign / Pair / Handoff 交互和全视口响应式样式。
- Web `316/316`、S10 定向 `7/7` 与 production build 已通过；memory 真实浏览器已完成 Plan → 两次 Campaign → Pair → exact Case / Suite Handoff，`1440×900`、`1024×900`、`390×844` 无横向溢出，控制台零 warning / error，刷新后仍恢复 exact evidence。
- SQLite 真实页面以 `app_cssvwuvwodxmxecz` 创建 Plan `aeplan_lkqe7gr7kjobmf73 v1`。首个 Campaign `aecamp_slj6slzdz35qvhne` 在第二个 LLM 节点以 `workflow_run_gateway_failed / gateway` 失败并保留真实证据，由此定位同一 Run 复用 quota request identity 的根因；修正为逐 LLM 节点派生 provider-attempt identity 后，baseline `aecamp_2xrptdhto6nbj7vc` 与 candidate `aecamp_qr2wmglzcj5eortb` 均成功。Pair Preview 得到 `comparable`、expected `unchanged`、actual `changed`、mismatch `1`，显式 Handoff 生成 Case `eval_034d69aec0d7a2323c7f222f v1` 与 Suite `suite_9a8017d686be57009c7ad973`。再次重启服务后精确恢复同一 Case / Suite；`1440×900`、`1024×900`、`390×844` 均无横向溢出，控制台零 warning / error，页面未回显原始凭据。

## 当前实现进度

- 批次 A 至 D 已实现：领域合同、memory / SQLite / PostgreSQL repository、SQLite `0016` / PostgreSQL `0019` migration、严格 Plan / Campaign / Pair HTTP、四 Profile 显式 executor 映射、authority checkpoint、确定性 Run ID、quota binding、interrupted reconciliation，以及 Comparison / Case / Suite handoff。
- Workflow Definition 结构化运行输入专题批次 D 在原 owner 上增加 Plan / Plan Version / Campaign v2 与第五个 `workflow_definition_executor_v2` Profile；typed fixture 只使用 `{inputs}`，Campaign 逐项重读 exact Definition / contract authority，Run v8 pair 交给 Comparison v7 和既有 Case / Suite。SQLite `0019` 与 PostgreSQL `0022` 扩展共享 Workflow Run Store；v1 资源不迁移、不回退，Campaign 与 Run 不复制 fixture 正文。
- PostgreSQL repository 使用行锁与 CAS；SQLite 与 PostgreSQL 都拒绝版本覆盖、资源删除和非法状态迁移。Case / Suite partial handoff 以 candidate campaign 为单一锚点逐项 checkpoint。
- 相邻测试覆盖计划版本、campaign 幂等与顺序执行、quota binding、多 LLM 节点独立 provider-attempt admission、authority drift、quota failure、reconcile 不重放、strict HTTP、权限、active API Key 归属、SQLite restart / corruption、PostgreSQL migration 与 repository contract、Comparison preview、complete / partial handoff。
- 批次 E 的设计、React、strict consumer、测试、build、三视口、memory 与 SQLite 真实浏览器连续链均已完成；开发服务均已停止。
- SQLite 已完成 exact Plan → 两次 succeeded Campaign → Pair → Handoff，并在服务重启后恢复同一 Case / Suite。首个失败 Campaign 作为多节点 quota identity 根因证据保留；修正后双 LLM 节点回归证明每个 provider attempt 独立准入。专题关闭，不扩张 production membership / OIDC、production secret、billing、自动执行、自动重试 / 续跑、发布 / 部署或业务写回。

## 验收方式

- Go：领域、HTTP、memory、SQLite、PostgreSQL、四 Profile executor、authority drift、quota、store failure、reconciliation 与 handoff 相邻测试。
- Web：strict decoder、状态矩阵、交互、响应式和 build。
- 浏览器：`1440×900`、关键断点、`390×844`；验证层级、选中、执行确认、campaign pair、partial / blocked、横向溢出、storage 和 console。
- 仓库：`git diff --check`、`go test` / `go vet`、`./scripts/check-repo.sh --fast`；因新增 API、schema、migration、permission 与阶段真相，关闭专题前运行全量 `./scripts/check-repo.sh`。

## 停止线

- 不支持 production environment、production authentication / membership / OIDC、production quota 或 production enablement。
- 不支持 billing、token / cost limit、额度预留、自动提额、自动禁用或自动路由。
- 不支持 schedule、cron、queue worker、并行 fan-out、自动 retry、fallback、resume、replay 或长期 agent loop。
- 不保存 provider raw response、完整输出、对话 transcript、历史 Run 输入或从 metadata 反推的内容。
- 不自动 approve candidate、activate assignment、发布、部署、写业务真相或执行 proposed action。
- 不新增第二套 Run、Comparison、Case、Suite、decision、quota 或 route selection owner。
- 不为普通 UI 和既有测试可承载的行为新增专项 checker、fixture 或 readiness 链。

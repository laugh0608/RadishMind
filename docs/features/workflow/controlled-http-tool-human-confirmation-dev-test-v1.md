# Workflow 受控 HTTP Tool 与人工确认执行（开发 / 测试态）v1

更新时间：2026-07-17

状态：`workflow_controlled_http_tool_human_confirmation_dev_test_v1_completed`

## 设计结论

本专题定义 Workflow 首个可执行外部工具纵向切片：用户从精确的已保存草案版本创建不可变 action plan，人工审查并批准后，再显式一次性消费批准，执行一个服务端 allowlist 中的只读 HTTP `GET`，最后把 schema 校验后的脱敏投影交给既有 LLM / output 链并写入 durable Run History。

边界评审与[唯一高风险实施任务卡](../../task-cards/workflow-controlled-http-tool-human-confirmation-dev-test-v1-plan.md)的三个批次均已完成。当前状态为 `workflow_controlled_http_tool_human_confirmation_dev_test_v1_completed`：版本化契约、durable action plan / confirmation、受控 transport、原子 claim、`workflow_run_record.v2`、三种 store、增量 migration、诊断、不确定结果对账、`/executions`、Web 执行链和真实浏览器重启复验已经形成完整证据链。

首版采用独立 action plan / confirmation 资源，不把人工等待写成 `running` run，也不保持 HTTP 请求等待人工操作。拒绝、暂缓、取消或过期均不创建伪 run；只有已批准计划被成功 claim 后才创建 `workflow_run_record.v2`。

## 用户价值与适用范围

目标用户是内部开发者和具备独立确认权限的工作区操作者。首版解决一条真实路径：

1. 在已保存、无本地脏改动的 Workflow 草案中选择一个已登记的只读 HTTP Tool。
2. 提交符合工具输入 schema 的公开参数，服务端生成确定、可审查、可过期的 action plan。
3. 人工查看工具身份、用途、参数投影、目标策略键、风险和输出边界，选择批准、拒绝或暂缓。
4. 具备执行权限的操作者显式执行已批准计划；服务端最多 claim 一次、最多发起一次网络尝试。
5. Run History 展示确认、工具尝试、失败边界、脱敏输出和副作用计数；合法工具 run 不再被误判为契约损坏。

本专题只属于开发 / 测试态，不提供 production capability enablement 或公开生产 API。

## 与现有契约的兼容边界

以下历史事实保持不变：

- `contracts/tool-registry.schema.json` 的 `execution_enabled=false` 和通用 `/v1/tools/actions` 的 blocked 行为继续冻结；它们是 Session / Tooling v1 的 contract-only 真相源，不原地放宽。
- `contracts/tool-audit-record.schema.json` v1 继续只表达 `not_executed / blocked`，且保留真实 session / turn 绑定；Workflow 不伪造 `session_id` 或 `turn_id`。
- `scripts/checks/fixtures/workflow-definition-run-record-boundary.json` 中旧 run-bound `confirmation-decision` 与 `blocked_confirmation_required` 状态机已经标记为 superseded / archived，只允许历史读取，不能提交决定。批次 A 已物化新的 Workflow HTTP Tool 版本化契约，并同步更新关联 checker、function-surface fixture 与离线 placeholder；不得重新激活第二套 confirmation 真相源，也不得把旧字段映射成伪 run。
- `workflow_run_record.v0` 与 `workflow_run_record.v1` 继续要求 `tool_calls=0`、`confirmation_calls=0`，历史记录不回填伪造数据。
- Executor v0 的 Prompt / LLM / condition / output 行为、同步终态响应和既有 API 保持兼容；工具执行只进入新的版本化执行路径。
- Draft Designer 中的 preview-only `tool:workflow-preview-readonly` 不是可执行 registry id，不能被自动迁移或解释成真实工具。

新增契约必须版本化，名称固定为：

- `workflow_http_tool_definition.v1`：工具定义、公开输入 schema、脱敏输出 schema、风险与确认要求。
- `workflow_http_tool_execution_profile.v1`：环境 enablement、固定目标策略、预算与网络策略。
- `workflow_http_tool_action_plan.v1`：不可变计划主体、状态、版本和摘要。
- `workflow_http_tool_confirmation_decision.v1`：append-only 人工 / 系统决定及其 plan、digest、scope 和 CAS 绑定。
- `workflow_http_tool_execution_audit.v1`：Workflow scope 下的确认与执行审计，不复用 Session audit v1。
- `workflow_run_record.v2`：允许一个受控工具尝试和一个已消费确认的运行记录。

可执行 `tool_id` 统一使用 `workflow.http.<stable_key>.v1`，必须满足 `^[a-z0-9][a-z0-9._-]*\.v[0-9]+$`。草案只能引用完整、精确版本，不接受 alias、前缀匹配或客户端自定义工具。

## 权威来源

工具定义与环境执行 profile 分层，但不形成两个 enablement 真相源：

| 层 | 唯一职责 | 不得包含 |
| --- | --- | --- |
| definition registry | 固定 tool id / version、用途、公开参数 schema、输出投影 schema、`requires_confirmation=true`、`risk=medium`、只读语义 | 环境 endpoint、credential、动态启用开关 |
| runtime execution profile | 固定 definition digest、启用环境、method、scheme / host / port / path、query 映射、网络策略、预算、credential policy | 用户输入 URL、模型生成参数、业务写回策略 |

只有 definition 与当前环境 profile 都存在、版本和 digest 匹配、profile 显式启用，工具才可生成计划。未登记、profile disabled、版本漂移、非法 tool ref 或环境不匹配均 fail closed。

## 首版允许拓扑

首版只接受一条有界主路径：

```text
prompt root -> one http_tool -> one or more existing llm nodes -> output terminal
```

- 精确草案版本只能有一个 `http_tool` 节点和一次工具尝试。
- `prompt` 仍是唯一 root，`output` 仍是唯一 terminal；首版工具路径不允许 condition 分支、并行 fan-out 或循环。
- 工具参数只来自创建 action plan 时的显式 typed public arguments；`input_text`、Prompt packet、LLM 输出和确认后的新输入都不能改变 method、target、query 或 header。
- 工具节点的 schema-valid 脱敏投影可作为后续 LLM packet；既有 LLM / output 预算和 advisory-only 输出边界继续适用。
- 草案中的 `http_tool` 必须继续满足 `requires_confirmation=true`；缺失时计划创建直接失败。

## Action plan 与摘要

计划创建时，服务端必须重新读取同 scope 下的精确已保存草案版本，验证图资格、definition、execution profile、权限和 public arguments；客户端不得提交已规范化计划或目标 URL。

计划至少固定：

- tenant、workspace、application、draft id / version、node id；
- tool id / version、definition digest、profile id / version / digest；
- 固定 `GET`、target policy key、规范化 query 参数和输出 schema digest；
- credential policy、timeout、response budget、计划创建者、创建时间与过期时间；
- `tool_plan_digest`、`record_version` 和当前状态。

独立 confirmation decision 不绑定尚不存在的 run。每条决定至少包含 `confirmation_id`、`plan_id`、tenant / workspace / application、draft id / version、node id、tool id / version、`tool_plan_digest`、`outcome`、`decided_by_actor_ref`、`decided_at`、`audit_ref`、决定前后的 plan `record_version`；系统过期或失效同样写明确 outcome 和稳定 actor source。允许 outcome 固定为 `approve / reject / defer / cancel / expire / invalidate`。同一 transition 必须在一个事务中完成 action plan CAS 与 append-only decision 写入，任何一侧失败都不改变状态。

`tool_plan_digest` 使用 `SHA-256`，输入为 RFC 8785 JSON Canonicalization Scheme 规范化后的上述稳定字段。摘要包含完整 scope、草案与节点、工具和 profile 版本、method、target policy key、公开参数、输出策略、预算、credential policy 与 expiry；不包含 secret、解析后的 IP、request id 或展示文案。

执行时服务端必须从持久计划、当前草案和当前 profile 重新构建摘要并逐项匹配，不信任客户端回传的 digest。草案版本、definition、profile、参数、预算或 scope 任一变化都使原计划失效，必须重新创建并确认。

## Confirmation 生命周期

计划默认 15 分钟过期，`defer` 不延长 expiry。所有状态改变都要求 `expected_record_version` 与完整 scope，使用数据库 CAS；重复点击或并发决策最多一个成功。

| 当前状态 | 允许转移 | 说明 |
| --- | --- | --- |
| `pending` | `approved`、`rejected`、`deferred`、`canceled`、`expired`、`invalidated` | 创建后等待人工决定 |
| `deferred` | `approved`、`rejected`、`canceled`、`expired`、`invalidated` | 仍可在原 expiry 前继续审查 |
| `approved` | `consumed`、`canceled`、`expired`、`invalidated` | 批准只产生一次执行资格，不发送网络请求 |
| `rejected` | 无 | 终态，不创建 run |
| `canceled` | 无 | 终态，不创建 run |
| `expired` | 无 | 终态，不创建 run |
| `invalidated` | 无 | 草案、definition、profile、参数、预算或 scope 漂移后的终态，不创建 run |
| `consumed` | 无 | 终态，禁止再次 claim 或执行 |

开发 / 测试态允许计划创建者自批，但必须独立持有 `workflow_tool_actions:confirm`；服务端同时记录 `planned_by`、`decision_by` 和 `executed_by`。这不是 production dual control 声明。

批准不能绕过 registry、profile、scope、目标、预算或草案复核。执行时发现 expiry 时将计划 CAS 为 `expired`；发现草案、definition、profile、参数、预算、scope 或 digest drift 时将计划 CAS 为 `invalidated` 并写 `confirmation_invalidated`。两者都不消费计划、不创建 run、不发网络请求；CAS 或 decision append 失败则保持原状态并 fail closed。

## API 与权限设计

首版沿用 User Workspace 资源族：

```text
POST /v1/user-workspace/workflow-drafts/{draft_id}/tool-action-plans
GET  /v1/user-workspace/workflow-tool-action-plans/{plan_id}?workspace_id=...&application_id=...
POST /v1/user-workspace/workflow-tool-action-plans/{plan_id}/decisions
POST /v1/user-workspace/workflow-tool-action-plans/{plan_id}/executions
```

- create 要求 `workflow_drafts:read` 与 `workflow_tool_actions:plan`。
- detail 要求 `workflow_tool_actions:read`。
- approve / reject / defer / cancel 要求 `workflow_tool_actions:confirm`。
- execute 要求 `workflow_tool_actions:execute`、`workflow_runs:execute` 与 `workflow_drafts:read`。
- 所有 route 继续要求已验证 actor context、tenant / workspace / application binding、strict JSON、request id、audit ref 和显式 dev gate。
- confirmation grant 与 execute grant 必须独立；只持有 `workflow_runs:execute` 不能批准或消费计划。
- 首版不新增 plan list、任意 actor 查询、删除、修改、reopen 或批量执行 API。

具体 request / response schema 与 dev gate 名称由同一个实现任务卡物化，不得另开平行 API 设计链。

## HTTP 与 SSRF 边界

首版 runtime profile 只允许：

- method 固定为 `GET`，无 request body；
- target 使用 profile 固定的 `https` scheme、精确 host、port 和规范化静态 path；公开参数只映射到 allowlist query key；
- redirect、ambient proxy、automatic retry、fallback 和 connection reuse across profiles 全部关闭；
- request header 只由服务端生成 `Accept: application/json`、固定 User-Agent 与追踪 id；用户不能控制 `Host`、`Authorization`、`Cookie`、hop-by-hop header 或任意自定义 header；
- DNS 解析出的全部 A / AAAA 地址都必须通过策略检查，拒绝 loopback、private、link-local、multicast、reserved、unspecified 和 cloud metadata；连接只能拨号到本次验证的地址，并保持原 TLS server name，防止 DNS rebinding；
- 单元 / 集成测试只能通过注入 transport 或显式 `test_only_loopback` profile 访问 `httptest`，且同时要求 test environment 与独立 test gate；该例外不得进入开发产品 profile 或 production 配置。

首版 `credential_policy=none`，不读取 outbound secret，不复用 Gateway northbound API key，也不允许客户端提交 credential handle。未来 credential binding 必须升级 profile 与审计版本并重新进行生产 secret 边界评审。

## 预算、响应与脱敏

- public arguments 规范化 JSON 最多 8 KiB，首版只允许 `resource_key` 与可选 `locale` 两个字段；不接受 URL、header 或 raw query string 字段。
- 单次网络 timeout 为 5 秒；包括后续 LLM 的整条同步 run 总预算仍为 30 秒。
- `max_attempts=1`；任何失败、超时或不确定结果都不自动重试。
- 仅接受 `200..299` 和 `application/json`；其它状态按稳定 response status failure 处理，解压后 raw response 最多 64 KiB。
- raw body 只存在于有界内存，先完成 JSON 解析、工具输出 schema 校验、字段 allowlist 投影和脱敏，再进入后续节点。
- 可持久投影最多 16 KiB，节点 preview 最多 512 字符；不得保存 endpoint、query 原文、header、credential、raw body、远端错误、DNS 信息或 stack。
- 不创建 `result_ref`、materialized URI、通用 tool result store / reader 或 checkpoint tool output。

## Store、claim 与崩溃语义

action plan、confirmation decision、execution attempt 和 `workflow_run_record.v2` 必须使用同一 backend mode，并共享原子 claim 边界：memory 使用同一锁，SQLite / PostgreSQL 使用同一数据库事务。selector 不匹配、marker / migration 不匹配或任一必需 pre-dispatch 写入失败时 fail closed，不回退 memory，也不发送网络请求。

运行时按职责分为四层：action service 只创建计划并记录人工决定；execution service 重读权威状态、构造受限 DAG、协调 claim、节点执行与终态提交；transport 只负责 DNS / IP 策略、TLS 绑定、单次 HTTPS GET 和响应投影；execution store 在同一事务中维护 plan、confirmation、attempt、run 与 audit。HTTP handler 只作为 consumer adapter 调用 execution service，不复制 claim、transport 或 reconciliation 逻辑；`/v1/user-workspace/workflow-tool-action-plans/{plan_id}/executions` 仅在开发 / 测试双重门禁、已验证身份和完整执行 scope 同时满足时注册。

显式 execute 按以下顺序进行：

1. 重读计划、草案、definition、profile 与权限，复算 digest、expiry 和 policy。
2. 在一个原子事务中把 `approved` CAS 为 `consumed`，创建唯一 execution attempt、`workflow_run_record.v2` running 记录和 `tool_execution_started` 审计。
3. 事务成功后最多发送一次 HTTP 请求；claim 后的 `tool_calls` 记为 1，表示“最多一次可能已对外发起的尝试”。
4. schema-valid 成功响应继续执行后续 LLM / output，最后写入终态 run 与 audit；已知本地拒绝或已知远端失败进入稳定失败终态。
5. 无法判断远端是否已处理时进入 `outcome_unknown`，计划保持 consumed，禁止 retry、resume 或再次执行。

工具响应成功后，后续 LLM 或 output 仍可能失败；此时 attempt 保持 `succeeded`，run 按真实下游失败进入 `failed` / `canceled`，工具完成审计仍记成功，不把 Gateway / provider 失败伪装成工具失败。只有 transport 已可能对外发起且 timeout / connection error 无法判定结果时才直接进入 `outcome_unknown`；DNS、策略、请求组装等明确发生在 dispatch 前的拒绝属于已知失败。

`workflow_run_record.v2` 增加终态 `outcome_unknown`。进程在 claim 后崩溃或终态写入失败时，重启后的 metadata-only reconciliation 只把超出 30 秒预算且仍为 claimed / running 的记录 CAS 为 `outcome_unknown`；它不得发送网络请求、恢复节点或重放计划。

| 失败点 | 持久状态 | 网络尝试 | 后续行为 |
| --- | --- | --- | --- |
| claim 事务前失败 | `approved` 或原状态 | 0 | 可修复原因后重新显式 execute |
| claim 事务失败 | 不改变 | 0 | fail closed，不回退 |
| claim 成功、发送前进程崩溃 | `consumed` + claimed run | 记为最多 1 次 | reconciliation 标记 `outcome_unknown`，禁止重试 |
| 请求确定未离开客户端 | `consumed` + failed run | 1 | 记录 transport failure，不自动重试 |
| timeout / 连接中断，远端结果不明 | `consumed` + `outcome_unknown` | 1 | 人工审查；新动作必须重新计划与确认 |
| 收到响应后终态写失败 | `consumed` + claimed / running | 1 | reconciliation 标记 `outcome_unknown`，禁止重试 |

本设计只承诺本地计划最多消费一次、最多 claim 一个网络尝试；不声明远端 exactly-once。

## Run History 与诊断兼容

`workflow_run_record.v2` 只为本专题打开受控副作用：

- `tool_calls=1` 表示已 claim 的唯一外部尝试；claim 前为 0。
- `confirmation_calls=1` 表示一个有效 approval 被成功消费，不统计页面点击或无效 CAS。
- `business_writes=0`、`replay_writes=0` 必须继续由领域、store 与数据库 CHECK 共同保证。
- 工具节点只保留 tool id、node id、confirmation ref、plan digest、profile digest、attempt status、HTTP status 类别、响应字节数、耗时和脱敏投影；不保存 URL、header 或 raw request / response。

memory、SQLite、PostgreSQL、Go validator / codec 和 Web history / detail 必须同时支持 v2；SQLite / PostgreSQL 通过新 migration 增量放宽 v2，不能改写 v0/v1 的零工具 invariant。旧记录继续原样读取。

Run History 继续使用 `/v1/user-workspace/workflow-runs`，status filter 增加 `outcome_unknown`，但不增加新过滤字段，因此既有 cursor 编码和过滤摘要结构保持不变。Web 必须把 v2 的 tool / confirmation 计数展示为“已授权受控副作用”，不能继续报 `forbidden side effect`。

Run Comparison 与 Evaluation 首版不比较含工具副作用的 v2 run，统一返回 `workflow_run_side_effect_profile_unsupported`；不能返回 `store_contract_mismatch`，也不能把工具输出纳入 case、revision、suite 或 baseline promotion。Diagnostics 的 `failure_boundary` 增加 `tool_policy`、`tool_confirmation`、`tool_transport`、`tool_response` 与 `tool_store`，并新增独立 `tool_failure_category`，不得复用 `gateway_failure_category`。工具 attempt 已成功但下游模型失败时，`tool_failure_category=none`，run 继续使用真实 Gateway / provider 边界和类别。

## 失败码与审计

稳定失败分类至少包括：

| failure code | 边界与行为 |
| --- | --- |
| `workflow_tool_not_registered` | definition 不存在或 tool ref 非法；不创建计划 |
| `workflow_tool_profile_disabled` | 环境 profile 缺失、disabled 或版本漂移；fail closed |
| `workflow_tool_target_denied` | scheme / host / port / path、DNS / IP 或 redirect 策略拒绝 |
| `workflow_tool_arguments_invalid` | public arguments 违反 schema、大小或 query 映射 |
| `workflow_tool_confirmation_required` | 未批准；不 claim、不创建 run |
| `workflow_tool_confirmation_rejected` | 已拒绝；终态，不创建 run |
| `workflow_tool_confirmation_expired` | 计划已过期；终态，不创建 run |
| `workflow_tool_confirmation_stale` | 草案、profile、expiry 或 record version 已变化 |
| `workflow_tool_confirmation_mismatch` | plan digest 或 scope 不匹配 |
| `workflow_tool_confirmation_invalidated` | 草案、definition、profile、参数、预算、scope 或 digest 漂移；计划进入 `invalidated` |
| `workflow_tool_action_canceled` | 计划已取消；终态，不创建 run |
| `workflow_tool_action_consumed` | 已被唯一执行者 claim；禁止再次执行 |
| `workflow_tool_transport_failed` | 已知连接 / 协议失败，返回 allowlist 摘要 |
| `workflow_tool_timeout` | 超过 5 秒预算；若远端结果不明则转 outcome unknown |
| `workflow_tool_response_status_invalid` | HTTP status 不在 `200..299`；不向后续节点传递 raw body |
| `workflow_tool_response_too_large` | 解压后超过 64 KiB，丢弃 raw body |
| `workflow_tool_response_invalid` | content type、JSON 或输出 schema 不合法 |
| `workflow_tool_outcome_unknown` | 可能已到达远端但结果不可判定；不自动重试 |
| `workflow_tool_store_unavailable` | plan / audit / run store 或 claim 事务失败；no fallback |
| `workflow_tool_store_contract_mismatch` | scope、版本、状态迁移或记录契约不满足 |

审计事件包括 `confirmation_requested`、`confirmation_recorded`、`confirmation_rejected`、`confirmation_deferred`、`confirmation_canceled`、`confirmation_expired`、`confirmation_invalidated`、`tool_execution_started`、`tool_execution_succeeded`、`tool_execution_failed` 与 `tool_execution_outcome_unknown`。审计只保存稳定 metadata 和 policy ref，不保存 endpoint、参数原文、响应正文或底层错误。

## 副作用矩阵

| 阶段 | governance metadata write | network/tool | provider | business truth | replay |
| --- | ---: | ---: | ---: | ---: | ---: |
| 计划创建 / 查看 | 有 | 0 | 0 | 0 | 0 |
| reject / defer / cancel / expire / invalidate | 有 | 0 | 0 | 0 | 0 |
| approve 未执行 | 有 | 0 | 0 | 0 | 0 |
| consume 并执行 | 有 | 最多 1 | 沿用现有上限 | 0 | 0 |
| failure / outcome unknown 审查 | 有 | 不新增 | 0 | 0 | 0 |

## Web 交互

Draft Designer / Review Handoff 只在真实 dev source、已保存精确版本和服务 gate 同时满足时展示入口。UI 必须分别呈现：

- 工具身份、用途、固定 `GET`、target policy key、公开参数投影、输出字段和预算；不展示可复制 endpoint 或内部 DNS 信息。
- `pending / deferred / approved / rejected / canceled / expired / invalidated / consumed` 状态、expiry、决定者与版本冲突。
- approve 与 execute 两个独立动作和权限提示；批准后不自动执行。
- execute 期间禁用重复提交；刷新后从 durable plan / run 记录恢复，不依赖浏览器内存。
- Run History 中的确认引用、工具 attempt、受控副作用计数、脱敏结果、失败分类和 `outcome_unknown` 人工审查提示。

默认 offline / sample surface 不发送请求，也不得伪造已批准或已执行记录。

## 实施拆分与完成定义

批次 A 当前已完成代码实施与本地可执行证据：版本化 definition / profile / plan / decision / audit / run v2 schema、`tool_version=1` 与摘要绑定、create / detail / decisions route、三种 store、CAS / expiry / invalidation、SQLite 增量 migration、PostgreSQL `0006` migration、Web durable review 与独立 grant 提示均已落地。Go 定向 / race、SQLite restart、契约 checker、Web 92 项测试与 build 已通过。

批次 A 的真实 PostgreSQL 专项已通过 `./scripts/run-workflow-saved-draft-postgres-dev-test.sh check`：完整 Platform integration suite、`0006_workflow_http_tool_actions` migration、configured startup profile、三种 store 一致性、并发 CAS、作用域、事务、重启和 no-fallback 证据均通过。批次 B 随后完成受控 transport、SSRF / DNS / rebinding 策略、原子 claim、run v2 runtime、诊断、comparison / evaluation 隔离和 metadata-only reconciliation；同一 PostgreSQL 专项已应用 `0007_workflow_http_tool_execution`，schema version 为 `workflow_run_store_v7`。批次 C 已注册独立 execution route，完成 Web approve / execute 分离、共享 v0 / v1 / v2 run consumer、Run History v2、确定性 TLS transport、SQLite 连续链、PostgreSQL 复验和真实浏览器重启恢复；批准阶段网络为 0，显式执行最多一次，consumed 计划不能重试。当前完成锚点为 `workflow_controlled_http_tool_human_confirmation_dev_test_v1_completed`。

[唯一高风险实施任务卡](../../task-cards/workflow-controlled-http-tool-human-confirmation-dev-test-v1-plan.md)已经定义，并在卡内按依赖顺序承接三个可验证子批次：

1. 版本化 definition / profile、durable action plan / confirmation、CAS、权限、memory / SQLite / PostgreSQL 与 Web 审查；本子批不包含 network transport。
2. SSRF-safe 单次 `GET` transport、原子 claim、audit、`workflow_run_record.v2`、diagnostics、outcome reconciliation 和三种 store 一致性。
3. Web approve / execute / history 纵向链、测试 transport 与显式 HTTPS dev profile、浏览器连续验收和 SQLite / PostgreSQL 重启复验。

每个子批都做精准验证，但只有三批全部完成后才能声明 `workflow_controlled_http_tool_human_confirmation_dev_test_v1_completed`。第一批不能以“设计完成”或“confirmation store 已完成”冒充工具功能完成，也不新增 readiness-after-readiness 链。

最终验收必须覆盖：领域 / CAS / race、HTTP strict JSON 与 scope、SSRF allow / deny、DNS rebinding、大小 / timeout、崩溃矩阵、三种 store no-fallback、v0/v1/v2 兼容、Web consumer、comparison / evaluation 显式不支持、浏览器批准与执行、重启恢复、fast 与 full 仓库门禁。

## 停止线

- 不启用通用 `/v1/tools/actions`，不原地修改 Tooling v1 const。
- 不允许任意 URL、用户 header、request body、write-capable method、模型生成目标或确认后修改参数。
- 不实现 shell、code、sandbox、RAG、agent loop、多工具、并行工具或 background execution。
- 不实现 automatic retry / fallback、replay、resume、checkpoint tool output 或 generic result reader。
- 不写 `RadishFlow`、`Radish`、Saved Draft、workflow definition 或其它业务真相源。
- 不接 production OIDC、production secret resolver、production credential、production repository mode、public production API、quota 或 billing。
- 不进行真实 Radish / RadishFlow 外部联调，不把开发 / 测试态完成写成 production ready。

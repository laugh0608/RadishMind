# Workflow 受控 HTTP Tool 与人工确认执行（开发 / 测试态）v1 实施任务卡

更新时间：2026-07-17

- 任务 ID：`workflow-controlled-http-tool-human-confirmation-dev-test-v1`
- 状态：`workflow_controlled_http_tool_human_confirmation_dev_test_v1_completed`
- 功能设计：[Workflow 受控 HTTP Tool 与人工确认执行（开发 / 测试态）v1](../features/workflow/controlled-http-tool-human-confirmation-dev-test-v1.md)

## 准入结论

本卡是该功能唯一的高风险实现入口，承接版本化契约、持久 action plan / confirmation、受控 HTTP transport、`workflow_run_record.v2`、Web 纵向链和双数据库复验。三个批次已经全部完成：批次 A 证明计划和确认阶段的网络请求、provider 调用和 workflow run 创建全部为 0；批次 B 证明 claim 前零网络、原子单次 claim、无 retry / resume / replay 与不确定结果对账；批次 C 完成独立 `/executions`、Web、Run History v2、SQLite / PostgreSQL 和真实浏览器重启链。完成锚点为 `workflow_controlled_http_tool_human_confirmation_dev_test_v1_completed`，不派生同层 readiness 任务卡。

## 用户目标

内部开发者可以从精确的已保存 Workflow 草案版本创建一个不可变 HTTP Tool action plan，由具备独立确认权限的操作者审查并批准，再由具备执行权限的操作者显式消费批准，最多执行一次服务端 allowlist 中的只读 HTTPS `GET`。最终运行记录必须能解释确认、工具尝试、脱敏结果、失败边界和不确定结果，同时继续保证业务真相源写入与 replay 为 0。

## 不可变架构边界

1. action plan 与 confirmation decision 在 run 创建前独立持久化；人工等待、拒绝、暂缓、取消、过期或失效都不创建 run，也不保持 `running` 请求。
2. approve 只授予一次执行资格，不发送网络请求；显式 execute 才能原子消费 `approved` 计划并创建 `workflow_run_record.v2`。
3. `workflow_run_record.v0/v1`、Executor v0、通用 Tooling v1 与 `/v1/tools/actions` 保持冻结；不得把 `http_tool` 加入既有 `workflowExecutorAllowsNode` 或 `POST .../runs`，不得原地放宽零工具计数或伪造 session / turn / run 绑定。
4. 首版只支持 `prompt -> http_tool -> one or more llm -> output`、一个版本化 tool、一个固定 `GET` target policy 和最多一次网络尝试；不支持 condition、并行、循环、多工具或后台运行。
5. definition 只定义稳定工具身份与公开输入 / 脱敏输出 schema；execution profile 是环境目标、网络策略和预算的唯一运行配置来源。客户端、LLM 与确认后输入都不能提交或改变 URL、method、header、credential 或 raw query。
6. action plan、confirmation decision、execution attempt、execution audit 与 run v2 必须共享现有 `workflow_runs` backend mode 和原子边界：memory 使用同一 owner lock，SQLite 使用同一 shared database transaction，PostgreSQL 使用同一 pool transaction。不得创建平行 DSN、独立 selector 或跨 store 补偿事务。
7. claim 的单一线性化点必须完成 `approved -> consumed` CAS、唯一 attempt、run v2 `running` 和 `tool_execution_started` audit；事务失败时网络尝试为 0，事务成功后禁止自动 retry、fallback、resume 或 replay。
8. 旧 `workflow-definition-run-record-boundary` 中 run-bound confirmation 与 `blocked_confirmation_required` 只保留为 archived legacy 占位。批次 A 必须在新契约物化时同步标记 superseded，并更新关联 checker、Workflow function-surface fixture 和离线 placeholder 页面，不能留下第二套可执行确认真相源。

## 版本化契约与稳定标识

实现只使用以下版本，不复用历史 Tooling schema：

| 契约 | 静态物化 / 运行启用 | 核心约束 |
| --- | --- | --- |
| `workflow_http_tool_definition.v1` | A | 精确 `tool_id`、公开输入 schema、脱敏输出 schema、`risk=medium`、`requires_confirmation=true` |
| `workflow_http_tool_execution_profile.v1` | A | definition digest、环境、固定 HTTPS target policy、预算、`credential_policy=none` |
| `workflow_http_tool_action_plan.v1` | A | 完整 scope、草案 / 节点 / tool / profile 版本、规范化参数、expiry、digest、CAS 版本和状态 |
| `workflow_http_tool_confirmation_decision.v1` | A | append-only outcome、actor、plan / digest / scope、决定前后版本和 audit ref |
| `workflow_http_tool_execution_audit.v1` | A | 确认与执行的脱敏稳定 metadata；不保存 endpoint、原始参数、响应正文或底层错误 |
| `workflow_run_record.v2` | A / B | 一个已 claim tool attempt、一个已消费 confirmation、`outcome_unknown`；business / replay 继续为 0 |

可执行 `tool_id` 必须匹配 `^[a-z0-9][a-z0-9._-]*\.v[0-9]+$`，首版登记值使用 `workflow.http.<stable_key>.v1`。所有 JSON Schema 放入 `contracts/`，Go domain、HTTP codec 与 Web consumer 必须读取同一字段口径；契约 artifact 存在不等于 runtime 已启用。

## HTTP API 与请求边界

服务端 pre-run plan / decision 门禁固定为 `RADISHMIND_WORKFLOW_TOOL_ACTION_DEV=1`；HTTP execution 另要求默认关闭的 `RADISHMIND_WORKFLOW_HTTP_TOOL_EXECUTION_DEV=1`。两者都要求既有 verified dev auth、Saved Draft HTTP、Workflow Executor 和兼容 `workflow_runs` store 已启用；execution 还必须通过当前环境 profile enablement。Web source 固定为 `VITE_RADISHMIND_WORKFLOW_HTTP_TOOL_SOURCE=dev-workflow-http-tool-http`，测试 loopback 独立门禁固定为 `RADISHMIND_WORKFLOW_HTTP_TOOL_TEST_LOOPBACK=1` 且只能在测试进程使用。默认或 offline / sample 模式不注册真实动作，不发送请求，也不伪造计划、批准或执行记录。

资源路由固定为：

```text
POST /v1/user-workspace/workflow-drafts/{draft_id}/tool-action-plans
GET  /v1/user-workspace/workflow-tool-action-plans/{plan_id}?workspace_id=...&application_id=...
POST /v1/user-workspace/workflow-tool-action-plans/{plan_id}/decisions
POST /v1/user-workspace/workflow-tool-action-plans/{plan_id}/executions
```

请求字段固定如下：

- create plan：`workspace_id`、`application_id`、`draft_version`、`node_id`、typed `public_arguments`；精确 `tool_id` 必须由服务端从已保存节点读取，客户端不能重复提交 tool、profile、target、method、header、digest、expiry 或 normalized plan。
- decision：`workspace_id`、`application_id`、`expected_record_version`、`decision`；人工只允许 `approve / reject / defer / cancel`，`expire / invalidate` 只能由服务端产生。
- execute：`workspace_id`、`application_id`、`expected_record_version`，以及沿用 executor 预算的 `input_text / model / temperature`；这些运行输入只能进入既有 Prompt / LLM 路径，不能改变或重新提交 tool arguments、target、profile、digest、draft version 或 confirmation ref，也不得持久化原始 `input_text`。
- detail：只接受精确 `plan_id` 与 query scope，不提供 list、任意 actor 查询、删除、修改、reopen 或批量执行。

所有响应使用严格 envelope：`request_id`、`workspace_id`、`application_id`、`action_plan` 或 `run`、`failure_code`、`failure_summary`、`audit_ref`。计划投影只能展示工具身份、固定 method、target policy key、公开参数审查投影、输出字段、预算、expiry、状态、决定者和 record version；不得返回 endpoint、DNS / IP、header、raw query、credential 或内部错误。

权限固定为：create 需要 `workflow_drafts:read` 与 `workflow_tool_actions:plan`，detail 需要 `workflow_tool_actions:read`，decision 需要 `workflow_tool_actions:confirm`，execute 需要 `workflow_tool_actions:execute`、`workflow_runs:execute` 与 `workflow_drafts:read`。确认和执行授权必须独立；开发测试态允许同一 actor 同时持有两项权限，但必须分别记录 planned / decided / executed actor。

## 批次 A：版本化治理资源与三种持久化

状态：`completed`；完成锚点为 `workflow_controlled_http_tool_human_confirmation_dev_test_v1_batch_a_completed`。

### 实现范围

1. 物化 definition、profile、action plan、confirmation decision、execution audit v1 与 run record v2 的静态 schema、validator、positive / negative fixtures；run v2 runtime 仍留在批次 B，旧 run-bound confirmation fixture / checker 同批标记为 superseded。
2. 服务端从精确 saved draft version、节点、definition、当前环境 profile 与 typed public arguments 构建 RFC 8785 canonical payload 和 SHA-256 `tool_plan_digest`；客户端提供的派生字段全部拒绝。
3. 建立 action plan 状态机：`pending / deferred / approved / rejected / canceled / expired / invalidated / consumed`。映射固定为 `approve -> approved`、`reject -> rejected`、`defer -> deferred`、`cancel -> canceled`、`expire -> expired`、`invalidate -> invalidated`；批次 A 不能进入 `consumed`，它只由批次 B 的 execution claim 产生。
4. 计划默认 15 分钟过期；`defer` 不延长 expiry。所有决定必须以完整 scope 和 `expected_record_version` 做 CAS，并在同一事务完成 plan CAS、append-only decision 与 execution audit；三者任一失败均不落库，并发决定最多一个成功。
5. action plan store 归入现有 `workflow_runs` owner：memory、SQLite、PostgreSQL 共享相同领域 contract、失败码和 no-fallback 语义。PostgreSQL 从现有 v5 之后追加 `0006` tool action migration；SQLite 从现有 `0001` 之后追加 `0002` 增量 migration，不能改写已应用文件；批次 B 如需继续扩表，只能顺序追加后续 migration。
6. 实现 create / detail / decisions 三条 route、strict JSON、权限与 scope 绑定、稳定失败映射和脱敏审计。`/executions` 在本批不注册，不创建 blocked 假 run；action service 也不得复用既有 executor service 的 nil-to-memory lazy fallback。
7. Web 新增独立 action-plan consumer / review panel，接入已保存且无本地脏改动的 Draft Designer / Review Handoff；展示计划、expiry、状态、权限和 CAS 冲突，批准后不自动执行。
8. 用领域、HTTP、store contract、SQLite / PostgreSQL migration / restart、race 与 Web tests 证明 network、provider、run、business write、replay 均为 0。

### 当前实施证据（2026-07-16）

1. 已物化 definition、profile、action plan、confirmation decision、execution audit 与 run record v2 schema；`tool_id / tool_version=1`、definition / profile / output schema digest 与不可变 `tool_plan_digest` 已贯通 Go、JSON Schema、fixture、数据库和 Web strict consumer。
2. 已实现 create / detail / decisions 三条 route、严格请求和脱敏响应、plan / read / confirm 独立授权、15 分钟 expiry、源漂移失效、单时间采样和 append-only decision / audit；批次 A 未注册 `/executions`。
3. memory、SQLite 与 PostgreSQL store 共享既有 workflow runtime owner；CAS 状态矩阵拒绝重复 `deferred -> deferred`，decision / audit 与 plan 状态在同一事务或同一 owner lock 内原子提交。
4. Web 已提供 durable plan 审查、刷新恢复、权限矩阵、CAS 冲突刷新与独立 approve / execute 提示；批次 C 已接通显式 execute，offline 和缺少授权时请求数仍为 0。
5. 已通过 Go 定向与 race、SQLite migration / restart、PostgreSQL integration tag 编译、契约聚合 checker、Web 92 项测试与生产构建；旧 confirmation placeholder 已标记为 archived / superseded。
6. 真实 PostgreSQL 专项已通过 `./scripts/run-workflow-saved-draft-postgres-dev-test.sh check`：完整 Platform integration suite、`0006_workflow_http_tool_actions` migration、configured startup profile、并发 CAS、三层 scope、事务回滚、重启和 no-fallback 证据均通过，临时容器与网络已由 runner 清理。

### 完成门禁

- 三种 store 的 create / read / decide / expire / invalidate、scope isolation、CAS race、restart recovery 与 no-fallback 语义一致。
- definition / profile 缺失、disabled、digest 漂移、参数非法、草案漂移或权限不足都失败关闭，且不创建计划或网络副作用。
- 旧 confirmation placeholder 不再是活动执行契约；历史展示明确为 archived / legacy，不能提交决定。
- Web 刷新后从 durable plan 恢复审查状态；默认 offline 模式请求数为 0。
- 本批完成锚点固定为 `workflow_controlled_http_tool_human_confirmation_dev_test_v1_batch_a_completed`，不得写成完整功能完成。

## 批次 B：单次 transport、原子 claim 与 run v2

状态：`completed`；完成锚点为 `workflow_controlled_http_tool_human_confirmation_dev_test_v1_batch_b_completed`。

### 实现范围

1. 实现注入式 HTTP transport 和固定 HTTPS `GET` profile：关闭 redirect、ambient proxy、automatic retry、fallback 与跨 profile 连接复用；只生成固定 `Accept`、User-Agent 和 trace header。
2. 对全部 A / AAAA 结果执行 SSRF policy，拒绝 loopback、private、link-local、multicast、reserved、unspecified 和 cloud metadata；只拨号到本次校验地址并保持原 TLS server name，覆盖 DNS rebinding。
3. 在同一 backend 原子 claim `approved` 计划，创建唯一 execution attempt、run v2 running 和 started audit；唯一约束和 CAS 必须同时阻止重复点击、并发进程与服务重启后的第二次执行。
4. 单次 timeout 5 秒、总 run 预算 30 秒、`max_attempts=1`；只接受 `2xx` 与 `application/json`，解压后 raw body 上限 64 KiB，脱敏投影上限 16 KiB、节点 preview 上限 512 字符。
5. raw body 只存在于有界内存，完成 JSON parse、输出 schema 校验、字段 allowlist 投影和脱敏后才能进入 LLM / output；不得创建通用 result store / reader。
6. 扩展 run v2 validator / codec / memory / SQLite / PostgreSQL，允许 `tool_calls=1`、`confirmation_calls=1`、`business_writes=0`、`replay_writes=0` 和终态 `outcome_unknown`；v0/v1 零工具 invariant 保持不变。
7. 扩展 diagnostics 的 `tool_policy / tool_confirmation / tool_transport / tool_response / tool_store` 与独立 `tool_failure_category`。Run Comparison / Evaluation 对含工具副作用的 v2 run 返回 `workflow_run_side_effect_profile_unsupported`。
8. 增加 metadata-only reconciliation：claim 后超过预算仍未终态的 attempt / run 只 CAS 为 `outcome_unknown`，不得发网络请求、恢复节点或重放计划。

### 完成门禁

- SSRF allow / deny、DNS rebinding、redirect、proxy、header、大小、content type、timeout 和 response schema 均有正负行为测试。
- claim 前失败网络为 0；claim 后任何不确定结果都保持 consumed 并进入 `outcome_unknown`，没有 retry / resume / replay 路径。
- memory、SQLite、PostgreSQL 对唯一 claim、崩溃矩阵、restart reconciliation、v0/v1/v2 兼容和 no-fallback 结果一致。
- 本批完成锚点固定为 `workflow_controlled_http_tool_human_confirmation_dev_test_v1_batch_b_completed`，仍不得写成浏览器纵向链或完整功能完成。

### 当前实施证据（2026-07-16）

1. 已实现注入式单次 HTTPS `GET` transport，关闭 redirect、ambient proxy、retry / fallback 与跨 profile 连接复用；固定 header、DNS 全量解析、地址策略、绑定已校验地址拨号和 TLS server name 共同约束 SSRF 与 rebinding。
2. execute service 在原子 claim 前重读 approved plan、confirmation、精确 draft、definition、profile 与 digest；memory 共享 owner lock，SQLite / PostgreSQL 共享原数据库事务，一次完成 `approved -> consumed`、唯一 attempt、run v2 running 与 started audit。
3. 成功投影只保留 allowlist 字段并进入既有 LLM / output 链；已知 transport / response 失败稳定终止，可能已发出的 timeout / transport 与终态写入失败进入 `outcome_unknown`，不自动重试、恢复或重放。
4. metadata-only reconciliation 只读取 claimed attempt / running run，并在超出预算后以系统 actor 写入 `outcome_unknown`；SQLite 重启与 PostgreSQL 重启、并发八路 claim 已证明不会产生第二次网络执行。
5. `workflow_run_record.v2` 已贯通 validator、memory / SQLite / PostgreSQL、history / diagnostics；Run Comparison 与 Evaluation 对工具副作用统一返回 `workflow_run_side_effect_profile_unsupported`，v0 / v1 零工具不变量保持不变。
6. PostgreSQL 专项已应用 `0007_workflow_http_tool_execution`，schema version 为 `workflow_run_store_v7`，checksum 为 `sha256:b62ab78c1289bd6d840c2abf109c79224e6321ce3853d985d9300564847d13db`；Platform 全量 Go test、专项 race、Go vet 与仓库 fast / full 门禁共同作为本批关闭证据。

## 批次 C：Web 执行纵向链与双数据库复验

状态：`completed`；完成锚点为 `workflow_controlled_http_tool_human_confirmation_dev_test_v1_completed`。

### 实现范围

1. 注册 `/executions`，Web 分离 approve 与 execute 动作，执行中禁用重复提交；刷新后从 durable plan / run 恢复，而不是依赖浏览器内存。
2. Run History / detail 支持 run v2、受控副作用计数、confirmation ref、tool attempt、脱敏投影、失败分类和 `outcome_unknown` 人工审查提示，不再把合法 tool count 报为 forbidden side effect。
3. 提供注入 transport 的确定性集成路径，以及仅测试环境和独立 test gate 可用的 `test_only_loopback`；正式开发路径只使用显式 HTTPS dev profile，loopback 例外不能进入产品 profile 或 production 配置。
4. 完成 SQLite 本地产品连续链与重启恢复；PostgreSQL 单独复验 migration / rollback / reapply、角色、并发 claim、marker mismatch、重启和 no-fallback，不能用 SQLite 结果替代。
5. 真实浏览器连续验收：保存精确草案 → 创建计划 → 审查 → 批准但不执行 → 显式执行 → Run History / detail → 重启恢复 → 重复执行拒绝。

### 当前实施证据（2026-07-17）

1. 已注册独立 `/executions` handler，严格请求 / envelope、完整 tenant / workspace / application 作用域和三项 execute scope 均在进入批次 B execution service 前失败关闭；测试 loopback 只允许显式 `TestOnly` server 与独立 test gate 同时启用。
2. Web 已分离 approve 与 execute，执行中和 consumed 状态禁用重复提交；共享严格 run consumer 同时兼容 v0 / v1 / v2，Run History / detail 展示 confirmation、tool attempt、受控副作用、脱敏投影和工具失败分类。
3. 确定性 TLS transport 集成、SQLite 创建草案到执行 / 重启 / 重复拒绝的完整链，以及 PostgreSQL migration / rollback / reapply、角色、并发 claim、marker、重启和 no-fallback 专项均通过。
4. 真实浏览器完成“精确 4 节点 / 3 边草案 → 保存 → 计划 → 批准但不执行 → 显式执行一次 → v2 历史 / 详情 → 服务重启 → 草案、consumed 计划和同一运行记录恢复”；测试目标的确定性 transport 失败形成稳定 `tool_transport` 诊断，计划保持 consumed 且没有 retry / resume。
5. 浏览器重启后没有新增 HTTP / runtime 错误；SQLite 主库、WAL 和 SHM 未发现原始输入、目标 URL、header 或 raw response。Web 96 项测试、生产构建、Platform 全量 Go 测试、PostgreSQL 专项与仓库 fast / full 门禁共同作为关闭证据。

### 完成门禁

- 浏览器可以明确区分 planned、approved、consumed、failed 与 outcome unknown，批准后网络计数仍为 0，执行后最多一次。
- 页面、URL、Web Storage、日志、数据库和浏览器产物不包含 endpoint、raw query、header、credential、raw response 或底层错误。
- 三批已经全部通过，状态已同步为 `workflow_controlled_http_tool_human_confirmation_dev_test_v1_completed`；该状态不代表 production ready。

## 主要实现落点

- 契约与兼容：`contracts/workflow-http-tool-*.schema.json`、`contracts/README.md`、`scripts/checks/fixtures/workflow-definition-run-record-boundary.json`、关联 workflow boundary / function-surface checker 与 fixture。
- Platform：`services/platform/internal/httpapi/` 中新增 tool contract、action service、HTTP route、store、transport 与 policy 文件；更新 `server.go`、`workflow_executor.go`、`workflow_run_store.go`、storage codec、history、diagnostics、comparison 和 evaluation。
- 配置：`services/platform/internal/config/`、本地启动脚本与 sanitized config summary；默认关闭，不输出 endpoint。
- 数据库：`services/platform/migrations/workflow_runs/` 的 A=`0006`、B=`0007` 与 `services/platform/migrations/sqlite/workflow_runs/` 的 A=`0002`、B=`0003` 只追加 migration / manifest / tests；复用现有 workflow run migration runner，并把 backend selector 收口为共享 workflow runtime backend owner。
- Web：`apps/radishmind-web/src/features/control-plane-read/` 中独立 action / execution consumer 与 panel、共享 run record consumer，以及 Review Handoff、Executor、Run History / detail 的严格类型和测试；不继续扩大单个已有长文件。

## 必要验证

- 批次 A：Go domain / HTTP / store contract / race，SQLite 与 PostgreSQL migration / restart / no-fallback，Web consumer tests / build，旧契约 supersession 检查。
- 批次 B：HTTP policy 与注入 transport、SSRF / rebinding、claim race、崩溃矩阵、run v0/v1/v2、comparison / evaluation、三种 store 与 reconciliation。
- 批次 C：Platform 全量 Go test / race / vet、Web test / build、SQLite 本地连续链、PostgreSQL 专项门禁、真实浏览器批准 / 执行 / 历史 / 重启链。
- 每批执行 `./scripts/check-repo.sh --fast`；批次 B 的执行边界和批次 C 的专题完成均补跑完整 `./scripts/check-repo.sh`。
- 优先扩现有 Go / TypeScript 行为测试和聚合门禁；只允许一个承接新版本契约与旧占位 supersession 的边界证据，不派生同层 readiness-after-readiness checker 链。

## 停止线

- 不启用任意 URL、用户 header、request body、write-capable method、credential、模型生成目标、多工具、并行、循环或 background execution。
- 不实现 shell、code、sandbox、RAG、agent loop、automatic retry / fallback、replay、resume、checkpoint tool output 或 generic result reader。
- 不写 `RadishFlow`、`Radish`、Saved Draft、workflow definition 或其它业务真相源。
- 不接 production OIDC、production secret resolver、production credential、production repository mode、public production API、quota、billing 或真实外部项目联调。
- 不把任务卡已定义、批次 A 的治理资源或开发测试态纵向链写成 production ready。

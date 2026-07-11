# Workflow Executor v0

更新时间：2026-07-11

状态：`workflow_executor_v0_implemented`

## 功能定位

`Workflow Executor v0` 是 `Saved Workflow Draft v1` 之后的首个受控执行能力。它让开发 / 测试态用户把已保存草案中的 `prompt`、`llm`、`condition` 和 `output` 节点交给 RadishMind Platform 顺序执行，经现有 Model Gateway 获取模型结果，并得到可重新读取的受限 run record。

本功能验证的是“草案可以在明确边界内形成真实运行结果”，不是发布系统、通用编排引擎或生产执行面。模型结果仍是 advisory output，不写入 `RadishFlow`、`Radish` 或其它业务真相源。

## 解锁判断

此前全面关闭 executor 的阶段门禁可以结束，但只能放宽为本专题定义的开发 / 测试态 v0：

- Saved Draft 已具备内存开发存储与 PostgreSQL 开发 / 测试持久化、版本冲突和作用域隔离；executor 只读取已保存版本，不执行浏览器中的未保存草案。
- Gateway R4 已具备稳定 bridge、stdio worker pool、超时、队列背压和 provider 选择；executor 不再自建第二套模型调用适配层。
- 执行节点、状态、失败语义、资源预算和副作用计数可在本专题内完整定义并自动验证。
- tool、confirmation commit、业务写回、publish、replay / resume 和 production enablement 仍缺少独立安全与持久化前置，因此继续关闭。

## 用户路径

1. 用户从当前 application 创建一个 executor v0 草案。默认结构是可编辑的 `Prompt → LLM → Output`，不会覆盖现有复杂草案。
2. 用户编辑节点说明和 provider ref，保存草案，得到明确的 `draft_version`。
3. Web 对当前已保存版本执行本地资格预检，展示可运行或被阻止的具体原因。
4. 用户输入本次运行文本；存在 condition 节点时，为每个 condition 节点显式选择布尔值。
5. Platform 重新按 scope 读取已保存草案、复核资格、建立 running record，并在总时限内执行节点。
6. POST 返回终态 run record；Web 展示节点状态、Gateway 选择、最终 advisory output、失败码和副作用计数。
7. Web 可通过 run id 再次读取同一进程内的记录，证明结果来自 run store，而不是仅存在于按钮回调中。

## 执行输入

运行请求只允许以下字段：

| 字段 | 约束 | 说明 |
| --- | --- | --- |
| `workspace_id` | 必填，必须与 dev header 一致 | 草案和 run record 作用域 |
| `application_id` | 必填，必须与 dev header 一致 | 草案和 run record 作用域 |
| `input_text` | 必填，UTF-8，去除首尾空白后 1 到 8 KiB | 本次用户输入；不写入 run record |
| `condition_values` | 可选，最多 16 个 `node_id → boolean` | condition 节点的显式输入，不由模型推断 |
| `model` | 可选，最长 256 字符 | 交给既有 Gateway inventory / provider 选择逻辑 |
| `temperature` | 可选，`0..2` | 不提供时使用 Platform 配置值 |

executor 不接受 secret、tool payload、HTTP endpoint、业务写回内容、confirmation decision、checkpoint 或 replay state。

## 草案准入

执行服务必须重新读取并校验已保存草案，不信任前端 eligibility 结果。草案同时满足以下条件才可进入运行：

- schema 为 `saved_workflow_draft.v1`，存在正整数 `draft_version`，validation state 为 `valid_for_review`。
- 节点数为 3 到 16，边数为 2 到 32；图是 DAG，所有节点可由唯一 prompt 根节点到达，并最终到达唯一 output 终点。
- 节点类型只包含 `prompt`、`llm`、`condition`、`output`；至少一个 `llm`，最多四个 `llm`。
- 节点不能携带 `tool_ref`、`rag_ref`、`requires_confirmation=true` 或 `risk_level=medium/high`。
- `requested_capabilities` 必须为空。草案顶层未被节点引用的 provider / tool / RAG catalog ref 只作为设计 metadata 保留，不会被解析或执行。
- condition 节点必须在请求中有显式布尔值；其出边只接受 `when:true`、`when:false` 或 `always` 三种机器可读条件摘要。
- 不允许孤立节点、循环、多个根 prompt、多个终点 output、从非 condition 节点使用条件路由，或没有活动路径到达 output。

资格失败时不得调用 Gateway，也不得创建伪成功记录。服务返回稳定 failure code 和可展示的 sanitized summary。

## 节点语义

### Prompt

- 输入为本次 `input_text`。
- 节点 `input_summary` 作为固定指令说明；executor 将其与用户输入组成有界 prompt packet。
- 不读取 workspace 外部数据，不进行 RAG，不展开 secret ref。

### LLM

- 输入为所有已激活前驱节点的有界输出，按拓扑顺序稳定拼接，并附加当前节点 `input_summary`。
- 只调用现有 Gateway `HandleEnvelope`，强制 `allow_retrieval=false`、`allow_tool_calls=false`、`allow_image_reasoning=false` 和 advisory safety。
- request-level `model` 交给既有 Gateway 选择；节点 `provider_ref` 进入 run record 的审查 metadata，但 v0 不把它解析成 endpoint 或 credential。
- 不自动 retry，不切换 fallback provider；失败即终止本次运行。

### Condition

- condition 的结果只来自 `condition_values[node_id]`，不由自然语言、模型输出或风险等级隐式推断。
- `when:true` / `when:false` 只激活与显式布尔值匹配的边，`always` 始终激活。
- 未被激活且没有其它活动入边的节点记录为 `skipped`。

### Output

- 汇总所有活动前驱输出并形成最终 advisory output。
- output 不调用 provider，不发送消息，不写业务状态。
- 最终结果最多 16 KiB；超出预算时以明确 failure code 终止，不静默截断成成功。

## 执行状态与失败语义

run 状态为 `running`、`succeeded`、`failed`、`canceled`；节点状态为 `pending`、`running`、`succeeded`、`skipped`、`failed`。同步 POST 在返回前必须把 run 推进到终态，但服务会先写入 running record，再随节点执行更新记录。

首批稳定 failure code：

| failure code | 语义 |
| --- | --- |
| `WORKFLOW_EXECUTOR_DEV_DISABLED` | executor dev gate 未打开时由统一 Platform error envelope 返回 |
| `workflow_run_scope_denied` | 身份、scope、workspace 或 application 不匹配 |
| `workflow_run_draft_not_found` | 作用域内找不到已保存草案 |
| `workflow_run_draft_version_unavailable` | 草案没有可执行的持久版本 |
| `workflow_run_draft_not_eligible` | validation、节点类型、风险、确认或图结构不满足 v0 |
| `workflow_run_input_invalid` | 输入、condition、model 或 temperature 不合法 |
| `workflow_run_graph_invalid` | DAG、根 / 终点、可达性或条件路由无效 |
| `workflow_run_budget_exceeded` | 节点、边、LLM 次数、输入、输出或总时限超预算 |
| `workflow_run_gateway_failed` | Gateway 调用失败或返回非成功 envelope |
| `workflow_run_output_unavailable` | 活动路径没有形成最终 output |
| `workflow_run_canceled` | request context 取消或总执行时限到期 |
| `workflow_run_record_not_found` | 当前作用域内找不到 run record |
| `workflow_run_store_unavailable` | run store 读写失败 |

错误摘要不得携带原始输入、credential、provider raw response、endpoint、stack trace 或内部错误串。

## Run record

run record schema 固定为 `workflow_run_record.v0`，至少包含：

- `run_id`、`draft_id`、`draft_version`、workspace / application scope。
- `status`、`failure_code`、sanitized `failure_summary`、`started_at`、`completed_at`。
- 输入字节数和 condition node id 列表，不保存 `input_text` 或 condition 布尔值。
- requested / selected model、provider、profile 和 selection source，不保存 credential 或 endpoint。
- 每个节点的类型、状态、耗时、前驱 id、有限 output preview 和 failure code。
- 最终 advisory output、request id、audit ref。
- `provider_calls`、`tool_calls`、`confirmation_calls`、`business_writes`、`replay_writes` 副作用计数。

executor v0 初始 run store 是进程内、最多 100 条、FIFO 淘汰的开发 / 测试存储。[Workflow Run History / Durable Dev-Test Run Store v1](workflow-run-history-durable-dev-test-store-v1.md) 已在不改变执行能力的前提下补齐 scoped list、分页和独立 PostgreSQL 开发 / 测试持久化；Saved Draft repository 没有被改造成 run repository，replay 与 production audit 仍未打开。

## API 与配置

- `POST /v1/user-workspace/workflow-drafts/{draft_id}/runs`
- `GET /v1/user-workspace/workflow-runs/{run_id}?workspace_id=...&application_id=...`
- 显式服务端 gate：`RADISHMIND_WORKFLOW_EXECUTOR_DEV=1`，默认关闭。
- 显式 Web consumer source：`VITE_RADISHMIND_WORKFLOW_EXECUTOR_SOURCE=dev-workflow-executor-http`，默认关闭。
- POST 要求 `workflow_runs:execute` 和 `workflow_drafts:read`；GET 要求 `workflow_runs:read`。
- 继续复用 dev auth、workspace / application headers、request id、audit ref 和 local-console CORS；不声明 public production API 或 production auth ready。

## 资源预算

- 单 run 总时限 30 秒，并受 request context 取消约束。
- 节点最多 16，边最多 32，LLM 调用最多 4。
- `input_text` 最多 8 KiB；内部节点 packet 最多 16 KiB；节点 preview 最多 512 字符；最终 output 最多 16 KiB。
- 进程内 run record 最多 100 条。
- v0 串行按拓扑执行，不建立后台 job、长驻 goroutine、自动 retry 或并行 fan-out。

## Web 交互

Web 在 Draft Designer 下方提供单独的 Executor v0 区域：

- 创建新的受控草案，不改写当前复杂草案。
- 展示已保存版本、服务 gate、图资格、阻止原因和运行输入。
- 运行期间锁定草案切换和重复提交；终态后展示节点时间线、模型选择、输出、失败和副作用计数。
- 支持按最新 `run_id` 重新读取记录。
- sample-only / route disabled / unsaved / dirty / conflict / blocked graph 状态均有明确文案，不伪装成可运行。

## 实现结果

2026-07-11 已完成首批实现：Platform 新增草案重新读取、服务端图准入、稳定拓扑执行、condition 分支、Gateway 调用、终态 run record、tenant / workspace / application 隔离的 100 条进程内 FIFO store，以及默认关闭的 POST / GET dev route；Web 新增受控草案创建、保存资格检查、运行输入、节点时间线、advisory output、副作用计数和 scoped record 回读。

真实浏览器验证已完成 `Create executor v0 draft → Save → Start bounded run → Reload run record`。验证记录显示 3 个节点均成功、provider call 为 1，tool / confirmation / business write / replay 均为 0，原始输入未进入 run record；完整运行时计划与就绪面板继续单独表达 tool、confirmation、writeback 等后续能力的阻塞状态。

## 验收

- Go 单元测试覆盖线性成功、condition 分支、节点跳过、Gateway 失败、取消 / 超时、预算、循环、工具 / RAG / confirmation / high-risk 拒绝和副作用计数。
- HTTP 测试覆盖显式 gate、scope、strict JSON、已保存版本读取、成功执行、scoped run read、not found、无 Gateway side effect 的负向路径和 CORS。
- Web 测试覆盖配置、受控草案构建、eligibility、request / response 映射、失败状态、run record 再读取和敏感输入不进入展示 record。
- 执行 Go tests、race、vet、Web build / test、相关 HTTP smoke、`./scripts/check-repo.sh --fast`；因本批改变执行边界、API、配置与阶段停止线，最终补跑全量 `./scripts/check-repo.sh`。

## 停止线

- 不执行 `http_tool`、RAG、code、sandbox、agent loop 或任意外部 tool。
- 不实现 publish、schedule、background job、parallel fan-out、automatic retry、fallback provider、checkpoint、replay 或 resume。
- 不提交 confirmation decision，不因 condition=true 解锁高风险动作。
- 不写 `RadishFlow`、`Radish`、Saved Draft、workflow definition 或其它业务真相源。
- 不把已完成的开发 / 测试 durable run repository 扩张为 production audit store、production auth / OIDC、public production route、quota / billing enforcement 或 cost ledger。
- 不把 v0 成功记录解释为 production ready、通用 workflow runtime ready 或 agent runtime ready。

## 后续顺位

Run History、Failure Review、Run Comparison、[Workflow Evaluation Cases / Batch Regression Review v1](workflow-evaluation-cases-batch-regression-review-v1.md)、[Baseline / Case Versioning v1](workflow-evaluation-baseline-case-versioning-v1.md) 与 [Evaluation Suite / Release Review v1](workflow-evaluation-suite-release-review-v1.md) 已完成。suite decision 只是审查证据，不解锁 executor 或发布；tool 与 confirmation 必须作为独立高风险功能设计推进，不能在 v0 上直接放开。

# Workflow Definition 绑定受控 HTTP Tool 的版本化发布与人工确认执行（开发 / 测试态）v1

更新时间：2026-08-15

状态：`workflow_definition_http_tool_v1_batch_b_persistence_next`

## 功能定位

本专题让包含一个受控 HTTP Tool 节点的 Workflow 从精确 Saved Draft 进入不可变 Definition Version、显式激活、action plan、人工确认、单次执行和 Run History。它补齐“工具草案可以执行，但不能成为可重复使用的 active Definition”这一产品缺口。

本专题复用既有 Workflow Definition release owner 和 HTTP Tool action / confirmation / transport owner，不建立第二套版本、激活、确认、工具注册、网络策略或运行历史真相源。Definition 只提供不可变运行来源；批准 Definition 不等于批准一次工具调用，每次执行仍必须创建独立 action plan 并取得人工确认。

功能只在显式开发 / 测试态门禁下开放，不声明生产工具执行、生产凭据、公开 API 或上层业务写回能力。

## 当前缺口

- 通用 `workflow_definition_executor_v1/v2` 只接受 `prompt | llm | condition | output`，并保持 `tool_calls=0`。
- 既有 HTTP Tool 已具备精确 Saved Draft 来源、action plan、人工确认、单次 claim、受限 HTTPS transport、`workflow_run_record.v2` 和三种开发测试态存储。
- HTTP Tool action plan 仍绑定 `draft_id / draft_version`；Saved Draft 归档、派生或继续编辑后，缺少面向 active Application 的不可变 Definition 运行来源。
- 直接放宽通用 Definition executor 会绕过独立 confirmation、tool profile 和网络策略，因此必须增加显式 execution profile，并继续委托给既有 HTTP Tool owner。

## 用户路径

1. Builder 保存一份符合受控工具拓扑的精确 Workflow Draft。
2. Builder 以 `workflow_definition_http_tool_v1` 创建 Definition candidate；服务端重读草案、工具定义和当前 execution profile。
3. Reviewer 查看不可变 snapshot、工具身份、风险、确认要求和兼容 blocker，追加人工决定。
4. `approve` 只物化不可变 Definition Version，不创建 action plan、run 或网络请求。
5. Operator 显式激活精确版本；激活前再次校验工具定义、profile digest 和图资格。
6. 用户从 active Definition 创建一次 action plan，提交结构化 public arguments；服务端不再依赖 Saved Draft 作为运行权威。
7. Confirm actor 审查并批准计划；批准仍不发送网络请求。
8. Execute actor 显式 claim 已批准计划，既有 transport 最多执行一次受控 HTTPS `GET`，随后进入 LLM / output 链。
9. Run History 展示 Definition 来源、确认、工具 attempt、失败边界、脱敏投影和副作用计数。
10. 服务重启后可恢复 Definition、activation、plan、decision、attempt 和 run metadata；不得恢复或重放网络副作用。

## Owner 与架构边界

### Workflow Definition release owner

继续负责 candidate、review、不可变 version、activation pointer / event / audit。新增 profile 只改变资格分类和 snapshot 语义，不让 release owner 创建 action plan、调用工具或持久化工具结果。

### HTTP Tool action owner

继续负责 action plan、confirmation、claim、execution audit、transport 和不确定结果对账。计划来源升级为显式版本化 union：

- `saved_workflow_draft`：兼容既有计划；
- `workflow_definition`：绑定 active Definition id、version、digest 和 pointer version。

两类来源共享相同状态机、权限、tool registry、execution profile、transport 和 store selector，不允许互相 fallback。

### Run owner

Definition-bound tool run 使用新的 strict run schema，保留 Definition provenance 与既有 tool attempt / confirmation 语义。不得把 Definition id 填入旧 `draft_id`，也不得改写历史 `workflow_run_record.v2`。

## Execution profile 与拓扑

profile 固定为 `workflow_definition_http_tool_v1`。通用 `workflow_definition_executor_v1/v2` 保持原能力边界。

首版支持有界线性拓扑：

```text
prompt -> one http_tool -> one or more llm -> output
```

要求：

- 恰好一个 `prompt`、一个登记的 `http_tool`、至少一个 `llm` 和一个 `output`；
- 工具节点必须引用精确 `workflow.http.<stable_key>.vN`，风险和 `requires_confirmation=true` 与 registry 一致；
- 其它节点不得声明 tool / RAG ref 或 confirmation；
- 不允许 condition、RAG、code、sandbox、循环、并行、fan-out 或多个工具；
- public arguments 只在创建 action plan 时由用户显式提交，不能由模型输出、URL、header 或确认后的附加输入替换。

## 版本化合同

批次 A 物化以下新版本：

- Definition candidate / version：显式表达 `workflow_definition_http_tool_v1`、精确 tool ref 和 confirmation requirement；tool definition / profile digest 在 action plan 创建时读取当前 registry 后绑定；
- action plan v2：作为严格 `workflow_definition` 来源成员，增加 `source_kind`、Definition version / digest 与 activation pointer；现有 Draft writer 在批次 B 数据库迁移前继续写 v1；
- confirmation / audit v2：绑定同一来源摘要，避免只靠 `draft_id` 解释；
- Definition-bound HTTP Tool run：独立 strict schema，保留 source、confirmation 和 attempt 证据。

旧 action plan、confirmation、audit 和 run v2 保持不可变兼容。Definition 来源新写入不得降级为旧 schema；批次 B 完成后，Draft 新写入也升级到带显式 `saved_workflow_draft` 来源的 v2，历史 v1 只读。

## Authority 与漂移

candidate 创建、review approve、activation 和 action plan 创建均重新读取当前权威，但只有 action plan 创建和 execute 关心本次执行资格。

execute 前必须重新读取并比较：

- Application lifecycle；
- activation pointer version 与 active Definition version / digest；
- Definition execution profile 与工具节点；
- tool definition version / digest；
-当前环境 execution profile version / digest；
- plan digest、expiry、scope、confirmation 和 expected record version。

任何漂移必须在 claim 和网络副作用前失败关闭或使计划进入 `invalidated`。claim 成功后的 transport 不确定结果继续进入 `outcome_unknown`，不得换回 Saved Draft、旧 Definition 或上一个 active version。

## API 与权限

Definition release 继续复用既有 candidate / review / version / activation API。创建 candidate 时由客户端显式选择允许的 execution profile，服务端根据草案和 registry 证明资格，不能由 node type 隐式放宽。

Definition 来源计划新增独立、语义明确的路由：

```text
POST /v1/user-workspace/workflow-definitions/{definition_id}/tool-action-plans
```

detail、decision 和 execution 继续复用既有 plan 资源路由。权限要求：

- candidate / review / activation 沿用 `workflow_definitions:*`；
- create plan 要求 `workflow_definitions:read` 与 `workflow_tool_actions:plan`；
- detail、confirm、execute 沿用既有独立 grants；
- execute 继续要求 `workflow_runs:execute`，但不再要求读取 Saved Draft。

默认 offline Web 零请求，未知字段、未知 query、跨 scope、错误 source union、非法 profile 和权限缺失均失败关闭。

## 数据与隐私边界

允许持久化：Definition / activation ref、tool/profile digest、规范化 public arguments、确认 metadata、attempt metadata、input digest / bytes、脱敏输出投影、failure code、request / audit ref。

禁止持久化：原始 prompt、input 正文、模型回答、provider raw response、HTTP raw body、endpoint、DNS / IP、header、credential、token、cookie、完整 query、业务写回 payload 或浏览器 transcript。

## 实施批次

### 批次 A：功能合同、显式 profile 与 memory 领域链

状态：已完成。

- 物化 Definition HTTP Tool profile 和版本化 contract。
- 让 candidate eligibility 显式区分通用 executor 与 HTTP Tool profile。
- 抽取可复用的工具图资格校验，避免 Definition 与 action plan 两份规则漂移。
- action plan 增加严格来源 union，完成 Definition authority resolver、memory create / decide 与 claim 前漂移失效行为测试。
- 计划、批准和激活阶段证明网络、provider 和 run 均为 0。

当前证据：显式 `workflow_definition_http_tool_v1` 已进入 Definition release 领域层；v3 candidate / version strict schema 已物化。旧请求省略 profile 时继续使用原 v1 / v2 executor profile，HTTP Tool 草案不能借默认 profile 晋级。Definition 来源 action plan 使用 v2 strict schema，绑定 active version、Definition digest、activation pointer 以及创建时读取的 tool / profile digest；confirmation 与 audit 同步使用 v2 来源摘要。memory 与 HTTP 路径已覆盖 create、独立人工 approve、active pointer 漂移失效、CAS 和零网络 / provider / run，Saved Draft v1 计划保持兼容。批次 A 已关闭，下一步进入双数据库顺序迁移。

### 批次 B：SQLite / PostgreSQL 持久化与兼容迁移

状态：下一顺位。

- 在既有 workflow run migration family 中顺序追加 schema，不创建 DSN、pool、database file 或 selector。
- 覆盖 v1 历史读取、v2 新写入、source union 约束、CAS、唯一 claim、rollback / reapply、重启、损坏投影和 no-fallback。

### 批次 C：Definition-bound 受控执行与运行审查

- 执行前重读 active Definition、tool definition/profile 和 confirmation。
- 复用既有单次 transport、SSRF policy、预算、output projection 和 reconciliation。
- 接入 Run History、Diagnostics、Comparison / Evaluation 的显式兼容策略，不重新执行工具。

### 批次 D：React 产品路径与专题关闭

- 完成 Draft → Candidate → Review → Version → Activation → Plan → Confirm → Execute → History。
- 完成 SQLite / PostgreSQL 连续链、服务重启、三视口真实浏览器、console / network / Web Storage 和敏感内容复验。
- 更新功能入口、当前焦点、路线图、能力矩阵和周志，运行全量仓库门禁后关闭专题。

## 验收方式

- contract：错误 source union、未知字段、非法 profile、digest 漂移和 forbidden material 全部拒绝。
- authority：review / activate 不调用网络；execute 前的任何 authority 漂移都在 claim 前失败。
- concurrency：同一 expected version 的决定或 claim 只有一个成功，重启后不会产生第二次网络执行。
- persistence：memory / SQLite / PostgreSQL 语义一致，无跨 store fallback。
- compatibility：通用 Definition、Saved Draft HTTP Tool、RAG、Application Session 和 run v0–v8 不回归。
- product：真实浏览器完成完整用户路径，并能区分 approved 与 executed。
- privacy：数据库、日志、URL、Web Storage 和 committed evidence 不包含禁止材料。

## 停止线

- 不放宽通用 `workflow_definition_executor_v1/v2`，不把 HTTP Tool 直接加入 executor v0。
- 不实现自动 approve / activate / plan / confirm / execute，不实现 schedule、background job、retry、fallback、replay 或 resume。
- 不实现多工具、condition + tool、RAG + tool、code、sandbox、agent loop 或业务写回。
- 不接 outbound credential、生产 secret、生产 auth、public production API、quota、billing 或 SLA。
- 不从本专题派生多张同层 readiness 卡、独立 selector、平行 confirmation owner 或专用数据库。

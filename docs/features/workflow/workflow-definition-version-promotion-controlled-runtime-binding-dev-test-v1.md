# Workflow 不可变版本晋级与受控运行绑定（开发 / 测试态）v1

更新时间：2026-07-19

状态：`workflow_definition_version_promotion_controlled_runtime_binding_dev_test_v1_batch_c_ready_for_implementation`

## 功能定位

本专题让内部 Workflow Builder 把一份已经保存、校验并人工审查的 Saved Draft 晋级为不可变 Workflow Definition Version，再由独立人工动作激活某个精确版本，最后让受控执行器从该版本运行并在 Run History、Comparison 与 Evaluation 中保留可复验的版本来源。

它补齐当前“可编辑草案直接进入 executor v0”与“只读 Workflow Definition summary”之间缺少正式版本权威的问题。Saved Draft 继续是可编辑设计记录；Workflow Definition Version 成为 RadishMind 自身的不可变运行配置真相源；Run Record 继续只保存运行证据，不反向修改草案或 definition。

本专题只承载开发 / 测试态内部产品能力，不声明 production publish、production repository、公开执行 API 或上层项目正式接入。

## 当前缺口与设计结论

- Saved Draft 已支持创建、编辑、校验、memory / SQLite / PostgreSQL 保存、版本冲突、恢复和 Review Handoff，但其正式边界明确不是 published workflow definition。
- executor v0 当前从精确 Saved Draft 版本执行，物理运行来源为 `workflow_draft`；这条兼容路径保留，但不继续承担 definition-bound execution。
- `WorkflowDefinitionSummary` 与 `/v1/user-workspace/workflow-definitions` 已提供历史只读投影和页面上下文，当前没有由 reviewed Saved Draft 晋级而来的 durable definition owner。
- HTTP Tool、Workflow RAG 与 Application RAG 已分别拥有 action plan / confirmation、retrieval authority 和 application runtime assignment，不能通过通用 definition activation 绕过这些独立控制面。
- definition-bound executor 必须使用新的 run schema 与 `execution_source_kind=workflow_definition`，不能把 definition id 填进旧 `draft_id` 或把旧 v0 record 改写成另一个语义。

## 目标用户路径

1. Builder 选择当前应用下一个精确 Saved Draft 版本；服务端重新读取草案、校验结果、图结构、provider / tool / RAG refs 和版本 metadata。
2. Builder 创建不可变 release candidate。candidate 保存 source draft id / version / digest、完整脱敏 definition snapshot、执行能力分类、blocker 和创建审计。
3. Reviewer 读取 candidate、source evidence、与当前 active version 的差异和 blocker，追加 `approve` 或 `reject` 决定；决定使用 `expected_review_version` CAS。
4. `approve` 只物化不可变 definition version，不自动激活、不执行节点、不调用 provider / tool / retrieval。
5. Operator 对一个已批准且仍满足 authority 的 version 执行 `activate`。每个 tenant / workspace / application / definition lineage 同时最多一个 active version；activation 使用 current pointer CAS 并追加事件与审计。
6. 用户从 active definition version 启动受控运行。服务端在创建 running record 前重新读取 active pointer、definition version 和 execution profile，并拒绝漂移、停用、归档或不支持的能力。
7. Run History 以新的 schema 显示 definition id / version / digest、source draft evidence、节点时间线、诊断和副作用计数；原始输入和模型回答仍不持久化。
8. Comparison 与 Evaluation 只比较兼容的 definition-bound runs，不重新执行 definition，也不自动改变 active version。
9. 用户从 active definition 派生新 Saved Draft 时，草案携带 `source_definition_id` 与 `base_definition_version`；新编辑必须重新创建 candidate、review 和 activation，不能原地修改旧 version。

## 资源与真相源

### Saved Workflow Draft

继续由既有 Saved Draft repository 拥有，允许编辑和版本 CAS。definition service 只能按精确 id / version 重读，不得修改、归档或提升 Saved Draft 状态。

### Workflow Definition Lineage

稳定标识一个应用内 Workflow 的版本序列，包含：

- `definition_id`、tenant / workspace / application / owner scope；
- 当前 active version pointer 与 pointer version；
- created / updated actor 与时间；
- lifecycle：`active | inactive | archived`。

lineage 不保存可变图 payload。active pointer 只是精确 version ref，不把 version 内容复制成第二份可编辑配置。

### Workflow Definition Release Candidate

candidate 是不可变审查快照，包含：

- candidate id、candidate state 与 review version；
- source draft id / version / digest；
- definition id（新 lineage 可由服务端分配，现有 lineage 必须精确匹配 scope）；
- canonical definition snapshot 与 digest；
- execution profile、risk summary、requires-confirmation summary；
- eligibility blocker、evidence refs、created actor / time、request / audit ref；
- append-only review decisions。

candidate state 固定为 `pending | approved | rejected | superseded`。客户端不得提交 digest、derived blocker、risk summary、review actor 或 candidate state。

### Workflow Definition Version

只由 approved candidate 物化，内容不可变：

- definition id / version / definition digest；
- exact candidate id / review version；
- exact source draft id / version / digest；
- schema version、name、description、nodes、edges、input / output contracts；
- provider / tool / RAG refs、requested capabilities 与受控 additional fields；
- execution profile 与 activation eligibility；
- created actor / time、request / audit ref。

version 不提供 update 或 delete。后续修改必须来自新 Saved Draft 和新 candidate。

### Activation Pointer、Event 与 Audit

activation pointer 是每个 definition lineage 的 current projection；`activate | replace | deactivate` 使用 `expected_pointer_version` CAS。每次成功决定必须在同一 owner lock / 数据库事务中写 current projection、append-only event 和 audit，任何一步失败都不得留下 partial write。

## 执行能力分类

首个受控运行 profile 固定为 `workflow_definition_executor_v1`，完整支持既有 executor v0 的 `prompt | llm | condition | output` 图、预算、取消、Gateway 调用和零业务副作用约束。

definition snapshot 可以保存 Saved Draft schema 允许的 tool / RAG refs，但以下情况不得激活到该 profile：

- 存在 `http_tool`、RAG、code、sandbox、agent loop、writeback、replay 或 resume capability；
- 图不满足现有 executor 稳定拓扑、contract 或预算规则；
- 节点要求 confirmation，但没有独立 execution authority；
- provider / profile ref 不能由既有 Gateway policy 解释。

HTTP Tool 与 Workflow RAG 后续只有在独立兼容评审完成后，才允许消费 definition version；它们继续保留自己的 plan / confirmation / snapshot / authority 资源和权限，不并入通用 activation。

## Run Record 与兼容策略

- 新增 `workflow_run_record.v5`，execution profile 固定为 `workflow_definition_executor_v1`。
- execution source 固定为 `workflow_definition`，包含 definition id、version 与 digest；source draft 只作为 provenance，不作为执行来源。
- v5 复用 v0 的节点状态、diagnostics、provider selection、duration 和 side-effect 语义，但以独立 strict schema 表达，不修改 v0–v4 contract。
- v0–v4 历史记录继续读取；旧 `workflow_draft` filter 只返回原草案运行，definition filter 只返回 v5。
- v5 持久化 input digest / bytes、节点摘要、配置 refs、request / audit refs 和副作用计数，不保存原始 input、prompt、模型原始响应、完整 answer、credential 或 header。
- Comparison / Evaluation 新增显式 `workflow_definition_executor.v1` profile；不同 definition lineage 或不兼容 profile 必须失败关闭。

## API 与权限边界

开发测试态管理 API 预计为：

- `POST /v1/user-workspace/workflow-definition-candidates`
- `GET /v1/user-workspace/workflow-definition-candidates`
- `GET /v1/user-workspace/workflow-definition-candidates/{candidate_id}`
- `POST /v1/user-workspace/workflow-definition-candidates/{candidate_id}/decisions`
- `GET /v1/user-workspace/workflow-definitions/{definition_id}/versions`
- `GET /v1/user-workspace/workflow-definitions/{definition_id}/versions/{version}`
- `GET /v1/user-workspace/workflow-definitions/{definition_id}/activation`
- `POST /v1/user-workspace/workflow-definitions/{definition_id}/activation-decisions`
- `POST /v1/user-workspace/workflow-definition-runs`

权限固定拆分为：

- `workflow_definitions:read`
- `workflow_definitions:write`
- `workflow_definitions:review`
- `workflow_definitions:activate`
- definition-bound run 继续要求 `workflow_runs:execute` 与 `workflow_runs:read`

所有管理 route 使用 verified actor、tenant / workspace / application / owner scope 与 strict JSON。默认 gate 关闭；offline Web 零请求；OIDC membership 未成立时继续失败关闭。

## 失败与并发语义

至少固定以下稳定失败类别：

- source draft 不存在、version 漂移、digest 漂移或 scope 不匹配；
- draft 不是 `valid_for_review`、图不可执行或包含 profile 不支持能力；
- candidate 不存在、已决定、被取代或 review version conflict；
- definition version 不存在、digest / candidate evidence 损坏；
- activation pointer version conflict、version 未批准、version 被取代或 application archived；
- definition 已 deactivate、active pointer 漂移或 execution profile unsupported；
- store unavailable / contract mismatch / schema migration missing；
- scope grant、tenant binding、workspace membership、application 或 owner scope 拒绝。

失败不得 fallback 到 sample、fake summary、memory store、旧 Saved Draft execution 或上一个 active version。相同 expected review / pointer version 的并发写只能有一个成功。

## Web 信息架构

- Draft Designer：显示 source definition provenance、candidate eligibility 和“创建晋级候选”入口。
- Definition Review：candidate snapshot、与 active version 的结构差异、blocker、append-only review 和人工决定。
- Definition Versions：version history、digest、source draft、review evidence、activation state 和派生草案入口。
- Runtime Binding：active pointer、activate / replace / deactivate、CAS conflict 和明确风险说明。
- Run：只允许从 exact active version 启动，展示 profile eligibility、definition evidence 和结果交接。
- Run History / Comparison / Evaluation / Application Operations：识别 v5 与 definition source，不把它归类为 Saved Draft 或 Application RAG。

应用切换必须清空 candidate、version、activation、run input / answer、conflict 和迟到响应。任何一次性运行输入或模型回答不得进入浏览器持久化介质。

## 实施批次

### 批次 A：strict contract、memory repository 与人工审查 API

状态：已完成。六份 strict contract、canonical digest、精确 Saved Draft authority reload、memory owner lock、人工 review / activation CAS、append-only event / audit 与八条管理 API 已实现；`RADISHMIND_WORKFLOW_DEFINITION_RELEASE_DEV` 默认关闭。批准只物化 version，activation eligibility 独立判断，未创建 migration、run v5、Web 或 launcher。

- 物化 candidate、candidate decision、definition version、activation pointer / event / audit strict contract。
- 实现 canonical digest、source draft authority reload、candidate eligibility 和 memory owner lock。
- 实现 candidate create / list / detail / decision、version list / detail、activation read / decision。
- 实现 review / activation CAS、append-only 与无 partial write 测试。
- 本批不实现 run v5、migration、Web 或 launcher。

### 批次 B：SQLite / PostgreSQL durable repository

状态：已完成。SQLite `0010` / marker v10 与 PostgreSQL `0013` / marker v13 已在既有 workflow shared backend 中物化 candidate、decision、version、activation pointer / event / audit；三种 repository 共用 domain contract，数据库事务保证 review / version 和 pointer / event / audit 原子提交。正式 Workflow Definition summary 已切换到 tenant / owner 隔离的 live projection，active lineage 精确投影 active version，inactive lineage 投影最新 approved version；offline sample 继续独立，不与 live 结果混合。

- 复用既有 workflow shared database / pool、selector 与 migration family，不创建新 DSN、pool 或数据库文件。
- 完成 candidate / decision / version / pointer / event / audit 事务、索引、marker、rollback / reapply、restart、corruption 与 no-fallback。
- 把正式 Workflow Definition summary read projection 接到新 owner；fake store 只保留明确 offline sample，不与 live repository 混合。

### 批次 C：definition-bound executor 与 run v5

- 新增 exact active version authority resolver、execution checkpoint 与 `workflow_definition_executor_v1`。
- 新增 run v5 strict contract、memory / SQLite / PostgreSQL persistence、history / detail / diagnostics。
- 接入 Comparison、Evaluation、Baseline 与 Suite 的显式 profile，不重新执行。
- v0–v4、旧 draft execution、HTTP Tool、RAG 与 Application RAG 不回归。

### 批次 D：Web、连续链与专题收口

- 完成 candidate / review / version / activation / run / history / comparison / evaluation Web 路径。
- 完成统一本地产品档、SQLite / PostgreSQL 连续链、服务重启、CAS 冲突和真实浏览器复验。
- 同步当前焦点、功能入口、路线图、能力矩阵、架构、契约和周志并关闭专题。

## 验收方式

- contract：未知字段、非法状态、digest 漂移、forbidden material 与跨资源 scope 必须拒绝。
- repository：同一 expected version 并发单一成功，projection / event / audit 原子，append-only 不可修改。
- authority：所有来源在写入和运行前重新读取；任何 drift 在副作用前失败关闭。
- runtime：成功恰好执行受支持图，provider call 符合节点计划；tool、retrieval、confirmation、business write、replay 均为 0。
- privacy：原始 input、answer、prompt、credential、header、provider raw payload 不进入 definition、run、audit、日志或浏览器 storage。
- compatibility：Saved Draft、v0–v4 history、HTTP Tool、Workflow RAG、Application RAG、Gateway history 与应用观测保持兼容。
- persistence：memory / SQLite / PostgreSQL 语义一致，migration / rollback / reapply、restart、corruption 和 no-fallback 可复验。
- product：真实浏览器完成 draft → candidate → review → version → activation → run → history → comparison / evaluation → derive new draft。

## 停止线

- 不自动 approve、activate、replace、deactivate、publish、schedule 或 release。
- 不实现 background job、parallel fan-out、automatic retry、fallback、checkpoint、replay、resume 或 agent loop。
- 不通过 definition activation 绕过 HTTP Tool confirmation、RAG snapshot authority 或 Application RAG assignment。
- 不持久化原始 input、answer、prompt、provider response、credential、token、header 或业务写回 payload。
- 不修改 Radish / RadishFlow 真相源，不实现 production auth、production repository、public production API、quota、billing 或 SLA。
- 不为每个批次派生多张任务卡、同层 readiness 文档或专项 checker 链；统一任务卡、行为测试、现有聚合门禁和必要 contract / migration 证据承载实施。

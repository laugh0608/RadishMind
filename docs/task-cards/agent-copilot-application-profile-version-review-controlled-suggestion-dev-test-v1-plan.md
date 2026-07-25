# Agent / Copilot 应用档案版本审查与受控建议（开发 / 测试态）v1 实施任务卡

更新时间：2026-07-25

状态：`agent_copilot_application_profile_version_review_controlled_suggestion_dev_test_v1_task_card_defined_batch_a_ready`

## 目标与准入结论

按[功能设计](../features/user-workspace/agent-copilot-application-profile-version-review-controlled-suggestion-dev-test-v1.md)交付“Copilot Profile Draft → immutable Profile Version → Configuration Draft v4 精确引用 → Publish Candidate v4 人工审查 → 显式 Runtime Assignment → Application API / Session 单次受控建议 → Run / Evaluation 审查”的开发测试态路径。

本任务卡是该专题唯一实施入口。Profile owner、配置、发布候选、assignment、canonical contract、Gateway、Session、Run 与 Evaluation 保持独立职责。首版不是自治 Agent：每次 invocation 最多一次计划内 Gateway / provider 副作用，工具、检索、业务写回、confirmation 执行和 replay 计数固定为 0。

准入结论：功能设计、现有代码兼容审计和版本分配已完成，可以进入批次 A。批次 A 只实现 strict contracts、确定性 Profile policy compiler、memory owner 和默认关闭 API；不得打开 provider 调用、Session、Run、配置绑定、发布候选 v4 或 assignment runtime。

## 实现基线与兼容审计

- 实现基线为 `c472a2ad`，分支为 `dev`，开始审计时工作区干净。
- Application Catalog 已允许 `workflow_copilot | docs_qa | agent | prompt_application`；本专题不增加 kind、不改 ID、生命周期、列表筛选、CAS 或归档语义。
- Application Configuration Draft 当前严格支持 v1 未绑定、v2 Workflow RAG binding、v3 Prompt Template binding；v4 只允许 `application_kind=agent` 和一个 Agent / Copilot Profile ref。
- Application Publish Candidate 当前严格支持 v1 / v2 / v3 并共享唯一 review / supersede 状态机；v4 只增加 Agent Profile 精确引用，不建立平行审批。
- Runtime Authority v1 只允许 Workflow Definition / Application RAG，v2 只允许 Prompt Application；Agent 使用独立 v3，不修改 v1 / v2。
- Application Session / Turn v1 只承载 Workflow Definition / Application RAG，v2 只承载 Prompt Application；Agent 使用独立 v3。
- Workflow Run Store 与消费端当前严格识别 v0–v6；Agent 使用 v7，旧 lineage、比较和评测规则不原地放宽。
- API key 当前允许 `models:read`、`chat:invoke`、`responses:invoke`、`messages:invoke`、`application_rag:invoke`、`prompt_application:invoke`；既有密钥不得自动获得 Agent 调用权限。
- Application Development Workspace 当前在 configure、promotion、controlled test 和 readiness 中使用 Prompt / 非 Prompt 二分，使 `agent` 继承 Workflow RAG owner；批次 E 必须改为显式 kind routing，批次 A 不改 Web。
- `copilot-request.schema.json` 与 `copilot-response.schema.json` 已固定 project、task、artifact、context、tool hints、安全模式、结构化回答和候选动作；本专题消费而不复制或放宽它们。

## 冻结的 contract 与版本

| 领域 | 新版本 | 冻结边界 |
| --- | --- | --- |
| Profile Draft | `agent_copilot_profile_draft.v1` | 完整规范化结构化策略只属于 Profile owner |
| Profile Version | `agent_copilot_profile_version.v1` | 从精确有效草案创建，创建后不可更新或删除 |
| Configuration Draft | `application_configuration_draft.v4` | 只为 `agent` 增加 `agent_copilot_profile_ref`，不得携带 RAG / Prompt binding |
| Publish Candidate | `application_publish_candidate.v4` | 保存配置与 Profile 精确引用 / digest，不复制 Profile source |
| Runtime Assignment | `agent_copilot_runtime_assignment.v1` | ref-only 当前指针、CAS、`activate | replace | revoke` |
| Assignment Event | `agent_copilot_runtime_assignment_event.v1` | 只追加 metadata，不调用 provider |
| Runtime Authority | `application_runtime_authority.v3` | 只允许 `agent_copilot_suggestion_v1`，v1 / v2 不原地放宽 |
| Application Session | `application_session.v3` | Agent profile 使用 v3，既有 v1 / v2 保持原语义 |
| Application Session Turn | `application_session_turn.v3` | Agent turn 引用 authority v3 与 run v7，不保存正文 |
| Workflow Run | `workflow_run_record.v7` | exact Profile lineage、canonical request 摘要、调用与失败 metadata |

Profile ID 使用 `acpf_[a-z2-7]{16}`，assignment ID 使用 `acra_[a-z2-7]{16}`。所有 digest 使用 `sha256:<64 lowercase hex>`；通用 ref 与既有 application / request / audit 规则保持一致。

`application_runtime_authority.v3` 必须包含 Application record version / lifecycle、assignment version / digest、candidate / review version、draft version / digest、Profile Version ref、project / allowed task policy digest、协议策略、模型资格摘要和整体 authority digest。旧 owner 继续写原版本，禁止用 nullable 字段把 Agent lineage 塞进 v1 / v2。

## Profile source 与确定性策略

### Profile 字段

`agent_copilot_profile_draft.v1` 与 immutable version 共享以下规范化源码：

- `profile_name`、`description`；
- `project`：`radishflow | radish`；
- `allowed_tasks`：只能来自所选 project 的 canonical task enum，至少 1 项、去重并按 canonical 顺序保存；
- `default_locale`、`allowed_locales`；
- `context_policy`：允许的 canonical context 分类、必填策略与最大字节数；
- `artifact_policy`：允许 kind / role、最大数量、单项与总字节预算；
- `response_policy`：允许 answer / issue / citation / proposed action kind 与最大条目数；
- `risk_policy`：候选动作确认规则和禁止形态；
- `tool_hints_policy`：首版三个 canonical hint 均固定为 false；
- draft / version、digest、校验、作用域、actor、request、audit 与时间 metadata。

Profile 不保存自由形式系统提示词、provider / profile / model、credential、运行 context、artifact content、完整 request / response 或业务权限。协议和默认模型继续属于 Configuration Draft。

### Canonical policy compiler

批次 A 实现纯函数式 policy compiler：

1. 按 `project` 从 canonical contract 投影任务集合，不维护第二份开放式 task registry。
2. 规范化任务、locale、artifact / response enum 与数值预算，计算稳定 `profile_digest` 和 `policy_digest`。
3. 强制 `tool_hints.allow_retrieval=false`、`allow_tool_calls=false`、`allow_image_reasoning=false`。
4. 强制 request `safety.mode=advisory`、`requires_confirmation_for_actions=true`，客户端不能放宽。
5. 响应中 `candidate_edit | candidate_operation | ghost_completion` 必须 `requires_confirmation=true`；整体响应在存在此类动作时也必须为 true。
6. `read_only_check` 可以不要求确认，但本专题仍不执行该动作；任何 `apply` 字段只作为候选描述返回。

同一规范化 Profile 必须产生相同 digest。Profile 校验不得调用 Gateway、provider、tool、retrieval、Application Session 或 Run Store。

## 领域预算

所有预算按 UTF-8 bytes 或规范化条目计算：

- Profile name 2–80 字符，description 最多 512 字符，规范化 Profile source 最多 64 KiB。
- `allowed_tasks` 最多 16 项；`allowed_locales` 最多 8 项；单个 locale 最多 32 字符。
- 单次 invocation context 最多 128 KiB。
- artifact 最多 16 项；单项 inline content 最多 128 KiB；全部 artifact content 合计最多 256 KiB。
- 完整 canonical `CopilotRequest` JSON 最多 512 KiB；完整 `CopilotResponse` JSON 最多 256 KiB。
- answers、issues、proposed actions、citations 各最多 64 项；单个用户可见文本字段最多 16 KiB。
- Profile Draft list 只返回 metadata 与 policy 摘要；读取完整 source 必须使用独立权限。

超出预算必须在 provider 副作用前失败关闭，不截断输入、不删字段后继续调用，也不把正文写入失败诊断。

## 冻结的 API 与权限

### Profile owner

- `POST /v1/user-workspace/agent-copilot-profiles/validate`
- `POST /v1/user-workspace/agent-copilot-profiles`
- `GET /v1/user-workspace/agent-copilot-profiles`
- `GET /v1/user-workspace/agent-copilot-profiles/{profile_id}`
- `POST /v1/user-workspace/agent-copilot-profiles/{profile_id}/versions`
- `GET /v1/user-workspace/agent-copilot-profiles/{profile_id}/versions`
- `GET /v1/user-workspace/agent-copilot-profiles/{profile_id}/versions/{profile_version}`

权限固定为：

- `agent_copilot_profiles:read`：读取脱敏摘要与版本引用；
- `agent_copilot_profiles:read_source`：读取草案或不可变版本的结构化源码；
- `agent_copilot_profiles:write`：validate / save；
- `agent_copilot_profiles:version`：从精确有效草案创建不可变版本；
- `agent_copilot_profiles:bind`：把精确版本附着到配置草案。

### 配置、发布与 assignment

- `POST /v1/user-workspace/application-configuration-drafts/{draft_id}/agent-copilot-profile-binding`
- 既有 Publish Candidate create / read / list / review 路由承载 v4。
- `GET /v1/user-workspace/applications/{application_id}/agent-copilot-runtime-assignment`
- `GET /v1/user-workspace/applications/{application_id}/agent-copilot-runtime-assignment/events`
- `POST /v1/user-workspace/applications/{application_id}/agent-copilot-runtime-assignment/decisions`

assignment 权限为 `agent_copilot_runtime:read | write`。Configuration binding 同时要求既有 draft write 与 `agent_copilot_profiles:bind`；Publish Candidate v4 源码审查额外要求 `agent_copilot_profiles:read_source`。

### 受控调用

- `POST /v1/agent-copilot/invocations`
- API key scope 固定为 `agent_copilot:invoke`。
- Application Session execute 继续使用 `application_sessions:execute`，但只有 v3 session 的 `agent_copilot_suggestion_v1` profile 可以委托唯一 invocation service。
- 客户端只能提交 application scope、canonical task / locale / artifacts / context、client invocation key 和必要 request metadata；不得提交 Profile、authority、candidate、provider、credential 或重试策略。

## 稳定失败语义

Profile owner 至少固定：

- `agent_copilot_profile_scope_denied`
- `agent_copilot_profile_payload_invalid`
- `agent_copilot_profile_secret_material_forbidden`
- `agent_copilot_profile_project_task_invalid`
- `agent_copilot_profile_policy_invalid`
- `agent_copilot_profile_not_found`
- `agent_copilot_profile_version_conflict`
- `agent_copilot_profile_store_unavailable`
- `agent_copilot_profile_write_disabled`
- `agent_copilot_profile_immutable_conflict`
- `agent_copilot_profile_digest_drift`
- `agent_copilot_profile_binding_ineligible`

后续 runtime 至少固定：

- `agent_copilot_runtime_assignment_not_found`
- `agent_copilot_runtime_assignment_version_conflict`
- `agent_copilot_runtime_candidate_ineligible`
- `agent_copilot_runtime_authority_changed`
- `agent_copilot_invocation_input_invalid`
- `agent_copilot_invocation_duplicate_running`
- `agent_copilot_invocation_canceled`
- `agent_copilot_invocation_outcome_unknown`
- `agent_copilot_response_contract_failed`

错误响应只返回稳定 failure code、脱敏 summary、公开版本及 request / audit ref；不得回显 Profile source、context、artifact content、完整 request / response、provider raw response、credential、token、header、endpoint 或 DSN。

## 兼容矩阵与必须保留的行为

| 消费面 | 既有版本 / 行为 | Agent 增量 | 防漂移验证 |
| --- | --- | --- | --- |
| Application Catalog | 四种 kind、原生命周期 / CAS | 不改 schema | `agent` 不自动获得 runtime，旧 kind CRUD 全通过 |
| Configuration Draft | v1 unbound、v2 RAG、v3 Prompt | v4 Agent Profile ref | 四版 strict decode；binding 互斥；旧 digest 不变 |
| Publish Candidate | v1–v3、统一 review / supersede | v4 Profile ref | 旧 transition 全通过；v4 不复制 source |
| Runtime Authority | v1 Workflow / RAG、v2 Prompt | v3 Agent only | 三版 exact profile；禁止跨版字段 |
| Session / Turn | v1 Workflow / RAG、v2 Prompt | v3 Agent only | 旧 session 可读写；新 turn 只引用 run v7 |
| Workflow Run | v0–v6 严格 lineage | v7 Agent | History / Detail 不误判；跨 lineage comparison 拒绝 |
| API key | 现有 6 scopes | 新增 `agent_copilot:invoke` | 旧 key 无隐式授权；scope / app / lifecycle fail closed |
| Canonical contract | `CopilotRequest / CopilotResponse` v1 | 只消费 | 项目 / task / context 正负样本继续通过 |
| Web surface | Prompt / 非 Prompt 二分 | 批次 E 显式 kind routing | `agent` 不挂 RAG / Prompt owner，未知 kind 失败关闭 |

批次 A 必须先把该矩阵固化为 Go / schema 相邻测试；后续各批只在对应列增加 runtime 行为，不以可选字段或 fallback 兼容。

## 实施批次

### 批次 A：strict contract、policy compiler 与 memory Profile owner

状态：`ready`。

1. A1：新增 Profile Draft / Version、Configuration v4、Candidate v4、Assignment / Event、Authority v3、Session / Turn v3 与 Run v7 strict schema 和 Go contract-only codec；先落兼容矩阵测试，不注册后续 runtime。
2. A2：实现 Profile canonical normalization、project / task 校验、policy compiler、预算、safety / confirmation invariant、secret guard 和相邻单元测试。
3. A3：实现 memory Draft / Version repository、作用域、CAS、不可变版本、stable list、corruption / unavailable、默认关闭 API 与 read / read_source 权限分离。

批次 A 完成条件：

- validate 零写入，save / version create 只写 Profile owner；
- 非 Profile 路由、全部失败路径和 contract-only codec 零 Gateway / provider 调用；
- Application Catalog 必须重读为 active `agent`；
- 旧配置、候选、Authority、Session、Turn、Run 和 API key 用例全部通过；
- 不存在内存 fallback、unknown-field 宽松解码或 source 跨 owner 复制。

批次 A 验证：schema 元校验、相邻 Go 单元测试、定向 race、`go test ./internal/httpapi`、`go vet ./...`、`git diff --check` 与 `./scripts/check-repo.sh --fast`。strict contract 和 compatibility 影响仓库真相时补跑全量 `./scripts/check-repo.sh`。

### 批次 B：SQLite / PostgreSQL 开发测试态持久化

- Profile owner 使用独立 migration family；assignment、Session / Turn v3 与 run v7 复用共享 Workflow runtime 的后续 migration。
- 完成 migration / rollback / reapply、marker / checksum、运行角色、重启、并发、corruption、no-fallback 和数据库敏感内容扫描。

### 批次 C：配置、发布审查与显式 assignment

- 启用 Configuration v4 ref-only binding、Publish Candidate v4 exact reload 与既有 review 状态机。
- 实现 assignment `activate | replace | revoke`、事件 CAS、read-time eligibility、drift / supersede；approve 不自动激活。

### 批次 D：单次受控建议、Session、Run 与 Evaluation

- 实现唯一 `agent_copilot_suggestion_v1` invocation service 与 provider 前 exact authority checkpoint。
- API key 与 Session v3 只委托同一 service；每次成功 invocation 恰好一次计划内 Gateway 调用。
- Run v7、History、Comparison、Evaluation 与 Operations 只保存 metadata；取消、幂等与 `outcome_unknown` 不 replay。

### 批次 E：类型专属 Web 与双数据库真实验收

- 显式 kind routing 挂载 Profile、binding、candidate review、assignment、controlled test、Session 与 Run / Evaluation handoff。
- SQLite / PostgreSQL 各完成连续浏览器链、重启、CAS、authority drift、取消、迟到响应、URL / storage / network / database 隐私复验。

## 当前下一步

直接进入批次 A1：先创建 10 份 strict schema、Go contract-only codec 与兼容矩阵测试。A1 通过前不实现 Profile repository / API；A1–A3 全部通过前不进入数据库、配置绑定、发布审查、assignment、Session、Run、provider 或 Web。

## 停止线

- 不实现 agent loop、多步自治规划、tool / retrieval executor、connector、在线搜索、业务写回、confirmation executor 或自动应用候选动作。
- 不实现 retry / fallback、schedule、replay / resume、长期记忆、quota、billing、生产认证、生产密钥或生产能力声明。
- 不改 canonical task 集合，不建立 Agent 私有 request / response 格式，不把 metadata history 当作可恢复回答。
- 不从 Prompt Application、Application RAG 或 Workflow Definition 复制 executor、Session、Run、发布审批或 Gateway 算法。
- 不因 v4 / v3 / v7 引入而重写、迁移或放宽既有配置 v1–v3、候选 v1–v3、Authority v1/v2、Session / Turn v1/v2 或 Run v0–v6。

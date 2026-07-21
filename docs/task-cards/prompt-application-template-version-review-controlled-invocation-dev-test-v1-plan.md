# 提示词应用模板版本审查与受控调用（开发 / 测试态）v1 实施任务卡

更新时间：2026-07-21

状态：`prompt_application_template_version_review_controlled_invocation_dev_test_v1_batch_a_in_progress`

## 目标与准入结论

按[功能设计](../features/user-workspace/prompt-application-template-version-review-controlled-invocation-dev-test-v1.md)交付“Template Draft → immutable Template Version → Configuration Draft 精确引用 → Publish Candidate v3 人工审查 → 显式 Runtime Assignment → Application API / Session 单次受控调用 → Run / Evaluation 审查”的开发测试态路径。

本任务卡是该专题唯一实施入口。Template owner、配置绑定、发布候选、运行时 assignment、Gateway、Session、Run 与 Evaluation 保持独立职责；模板正文不得复制到配置、候选、assignment、Session、Run、Gateway Request History 或 Operations。

准入结论：功能设计已确认，可以进入批次 A；批次 A 的 strict contract、确定性渲染器与 memory Template owner 通过前，数据库、发布绑定运行实现、provider 调用、Session Prompt profile 和 Web 全链保持关闭。

## 实现基线与现有兼容边界

- 实现基线为 `c6b4d80d`，分支为 `dev`，开始任务时工作区干净并与 `origin/dev` 同步。
- Application Configuration Draft 当前支持 `application_configuration_draft.v1` 与 `.v2`；v2 只承载 Workflow RAG binding。
- Application Publish Candidate 当前支持 `application_publish_candidate.v1` 与 `.v2`；v2 只承载 Workflow RAG binding，既有 review 状态机继续作为唯一应用人工审批 owner。
- Application Session、Turn 与 Runtime Authority 当前分别为 `application_session.v1`、`application_session_turn.v1`、`application_runtime_authority.v1`，只允许 Workflow Definition v5 与 Application RAG v4 profile。
- Workflow Run Store 当前兼容 `workflow_run_record.v0` 至 `.v5`；新 Prompt lineage 不修改既有版本。
- API key 当前允许 `models:read`、三种通用协议调用和 `application_rag:invoke`；Prompt 调用必须使用新增独立 scope。
- SQLite Workflow migration marker 当前为 v12，PostgreSQL Workflow marker 为 v15；配置草案与发布候选使用独立现有 migration family。

## 冻结的 contract 与版本

| 领域 | 新版本 | 冻结边界 |
| --- | --- | --- |
| Template Draft | `prompt_application_template_draft.v1` | 完整规范化源码、变量与输出契约只属于 Template owner |
| Template Version | `prompt_application_template_version.v1` | 从精确有效草案版本创建，创建后不可更新或删除 |
| Configuration Draft | `application_configuration_draft.v3` | 只为 `prompt_application` 增加 `prompt_template_ref`；不得同时携带 RAG binding |
| Publish Candidate | `application_publish_candidate.v3` | 保存配置与模板精确引用及 digest，不复制模板正文 |
| Runtime Assignment | `prompt_application_runtime_assignment.v1` | ref-only 当前指针、CAS、activate / replace / revoke |
| Assignment Event | `prompt_application_runtime_assignment_event.v1` | 只追加、metadata-only，不调用 provider |
| Runtime Authority | `application_runtime_authority.v2` | 增加 `prompt_application_invocation_v1` 分支；v1 不原地放宽 |
| Application Session | `application_session.v2` | Prompt profile 使用 v2；既有 v1 session 保持原语义 |
| Application Session Turn | `application_session_turn.v2` | Prompt turn 引用 v2 authority 与 run v6，不保存 input / output |
| Workflow Run | `workflow_run_record.v6` | Prompt exact lineage、input digest / bytes、变量名摘要、调用与失败 metadata |

`application_runtime_authority.v2` 的 Prompt 分支必须包含 Application record version / lifecycle、assignment version、candidate / review version、draft version / digest、template version / digest、协议与模型资格摘要以及整体 authority digest。旧 Workflow / RAG Session 继续写 v1；不得为兼容而把 Prompt 字段加入任何 v1 schema。

## 冻结的 API 与权限

### Template owner

- `POST /v1/user-workspace/prompt-application-templates/validate`
- `POST /v1/user-workspace/prompt-application-templates`
- `GET /v1/user-workspace/prompt-application-templates`
- `GET /v1/user-workspace/prompt-application-templates/{template_id}`
- `POST /v1/user-workspace/prompt-application-templates/{template_id}/versions`
- `GET /v1/user-workspace/prompt-application-templates/{template_id}/versions`
- `GET /v1/user-workspace/prompt-application-templates/{template_id}/versions/{template_version}`

权限固定为：

- `prompt_application_templates:read`：只读脱敏摘要与版本引用；
- `prompt_application_templates:read_source`：读取草案或不可变版本源码；
- `prompt_application_templates:write`：validate / save；
- `prompt_application_templates:version`：从精确有效草案创建不可变版本；
- `prompt_application_templates:bind`：把精确版本引用附着到配置草案。

### 配置、发布与 assignment

- `POST /v1/user-workspace/application-configuration-drafts/{draft_id}/prompt-template-binding`
- 既有 Publish Candidate create / read / list / review 路由承载 v3，不新建平行审批路由。
- `GET /v1/user-workspace/applications/{application_id}/prompt-runtime-assignment`
- `GET /v1/user-workspace/applications/{application_id}/prompt-runtime-assignment/events`
- `POST /v1/user-workspace/applications/{application_id}/prompt-runtime-assignment/decisions`

assignment 权限固定为 `prompt_application_runtime:read` 与 `prompt_application_runtime:write`。Configuration binding 必须同时具备既有 draft 写权限和 `prompt_application_templates:bind`；Publish Candidate v3 源码审查必须额外具备 `prompt_application_templates:read_source`。

### 受控调用

- `POST /v1/prompt-applications/invocations`
- API key scope 固定为 `prompt_application:invoke`。
- Application Session execute 继续使用 `application_sessions:execute`，但只有 v2 session 的显式 `prompt_application_invocation_v1` profile 可以委托 Prompt invocation service。
- 客户端只提交 application scope、变量、`client_invocation_key` 和必要 request metadata；不得提交模板、版本、digest、authority、provider、credential 或重试策略。

## 领域预算与失败关闭

首版冻结以下预算，所有计数按 UTF-8 bytes 或规范化条目计算：

- 每个模板最多 16 条消息；每条消息源码最多 16 KiB；完整模板源码最多 64 KiB。
- 最多 64 个变量；变量名最长 64 字符；单个字符串变量最多 16 KiB；`string_list` 最多 128 项且规范 JSON 后最多 32 KiB。
- 最终 canonical rendered messages 总计最多 128 KiB。
- `text` 输出最多 64 KiB；`json_object` 原始输出最多 64 KiB；JSON Schema 最多 32 KiB、对象层级最多 8、属性总数最多 128。
- 角色只允许 `system | developer | user`；变量类型只允许 `string | integer | number | boolean | string_list`。

稳定失败语义至少包括：

- `prompt_template_scope_denied`
- `prompt_template_payload_invalid`
- `prompt_template_secret_material_forbidden`
- `prompt_template_syntax_invalid`
- `prompt_template_variable_invalid`
- `prompt_template_output_contract_invalid`
- `prompt_template_not_found`
- `prompt_template_version_conflict`
- `prompt_template_store_unavailable`
- `prompt_template_write_disabled`
- `prompt_template_immutable_conflict`
- `prompt_template_digest_drift`
- `prompt_template_binding_ineligible`
- `prompt_runtime_assignment_not_found`
- `prompt_runtime_assignment_version_conflict`
- `prompt_runtime_candidate_ineligible`
- `prompt_runtime_authority_changed`
- `prompt_invocation_input_invalid`
- `prompt_invocation_duplicate_running`
- `prompt_invocation_canceled`
- `prompt_invocation_outcome_unknown`
- `prompt_invocation_output_contract_failed`

错误响应只返回稳定 failure code、脱敏 summary、公开版本、request / audit ref；不得回显模板、变量值、渲染消息、模型输出、provider raw response、credential、token、header、endpoint 或 DSN。

## 批次 A：strict contract、确定性渲染器与 memory owner

状态：`in_progress`。

### A1：受限模板领域内核

状态：`completed`。

- 实现 `{{ variable_name }}` 单轮 parser，拒绝表达式、属性访问、函数、过滤器、控制结构、include、宏、转义执行和嵌套解释。
- 实现变量声明、canonical normalization、deterministic renderer、rendered digest、`text` / `json_object` 输出契约 validator。
- 实现模板、消息、变量、渲染结果预算和 credential-like 敏感材料守卫。
- 同一模板与规范化变量必须产生完全相同的 canonical messages 与 digest。

完成证据：

- 已建立单轮受限 parser、canonical variable normalization、deterministic renderer 与 rendered digest；变量值中的占位符只作为字面量，不递归解释。
- 已覆盖 `string | integer | number | boolean | string_list`、required / safe default、额外 / 缺失变量、重复 / 未使用声明、非法表达式和三层 byte budget。
- 已实现 `text | json_object` 输出契约、受控 object / array / scalar schema 子集、required / additional properties、深度 / 属性 / schema 大小预算和输出终态校验。
- 模板源码、变量说明 / 默认值和运行输入均拒绝 credential、token、header、cookie 与 DSN-like 材料；失败结果不返回 rendered messages 或 digest。
- 定向单元测试、定向 race、完整 `internal/httpapi`、`go vet` 与 `git diff --check` 已通过；没有新增 HTTP、repository、migration、provider 或 Web 接线。

### A2：strict schema、memory Template owner 与 API

状态：`completed`。

- 新增 Template Draft / Version strict JSON Schema 与 Go domain validation。
- memory repository 实现 application / tenant / workspace / owner scope、CAS、不可变版本、稳定排序和故障注入。
- Template API 默认关闭，read summary 与 read source 权限严格分离。
- validate 不写入；save 只写 Template owner；version create 必须重读精确 draft version 并验证 digest。

完成条件：非 Template 路由零写入，所有失败路径零 provider 调用；不存在内存 fallback、正文跨 owner 复制或未知字段兼容。

完成证据：

- 新增 `prompt_application_template_draft.v1` 与 `prompt_application_template_version.v1` Draft 2020-12 strict schema，并接入既有仓库级 schema 元校验；Go strict codec 同时拒绝未知字段与 credential-bearing 扩展字段。
- memory Template owner 按 tenant / workspace / application / owner 隔离，支持草案 expected-version CAS、稳定摘要排序、精确草案版本重读、不可变版本和显式 unavailable 故障；corruption、digest drift 与 no-fallback 均失败关闭。
- 七条 Template API 使用独立默认关闭 HTTP / write gate；validate 零写入，摘要读取、草案 / 版本源码读取、写入和版本创建分别使用 `read`、`read_source`、`write` 与 `version` 权限。
- save / version create 在写入前重读 Application Catalog active record 并要求 `application_kind=prompt_application`；列表和版本列表不返回 messages、content 或变量值。
- 定向单元测试、定向 race、完整 `internal/config`、完整 `internal/httpapi`、`go vet` 和 schema 元校验均已通过；所有 Template HTTP 路径与失败用例的 Gateway 调用计数保持为零。

### A3：后续版本 contract 占位边界

状态：`next`。

- 只创建配置 v3、候选 v3、assignment v1 / event v1、authority v2、Session / Turn v2、run v6 strict schema 与 codec 测试。
- 本子批不得接入配置保存、候选 review、assignment repository、Session coordinator、Run Store 或 provider。

完成条件：JSON Schema、Go codec 与后续 TypeScript strict consumer 共享字段表；v1/v2/v0–v5 兼容用例保持通过。

批次 A 总门禁：Platform 相邻单元测试、contract 校验、`go test ./internal/httpapi`、`git diff --check` 与 `./scripts/check-repo.sh --fast`。A1–A3 均完成前不得进入批次 B。

## 批次 B：SQLite / PostgreSQL 开发测试态持久化

状态：`blocked_by_batch_a`。

- 新增独立 `prompt_application_templates` migration family：SQLite / PostgreSQL 均从 `0001` 开始，承载 Draft 与 immutable Version owner。
- Workflow shared runtime 下一 migration 固定为 SQLite `0013`、PostgreSQL `0016`，承载 Prompt assignment、v2 Session / Turn 与 v6 run 所需投影。
- Configuration Draft v3 与 Publish Candidate v3 继续使用既有 payload 表和独立 owner，不因 JSON schema 升级新增平行表。
- 完成 migration / rollback / reapply、marker / checksum、运行角色、重启、并发、corruption、no-fallback 和数据库敏感材料扫描。

## 批次 C：配置、发布审查与 runtime assignment

状态：`blocked_by_batch_b`。

- 实现 Configuration Draft v3 ref-only binding、Publish Candidate v3 exact reload 与既有 review 状态机组合。
- 实现 assignment activate / replace / revoke、事件 CAS、read-time eligibility 与 drift / supersede 失败关闭。
- Candidate approve 不自动创建、替换或恢复 assignment。

## 批次 D：受控调用、Session、Run 与 Evaluation

状态：`blocked_by_batch_c`。

- 实现 `prompt_application_invocation_v1` 唯一 invocation service 与 provider 前 exact authority checkpoint。
- API key 与 v2 Session 只委托同一 service；每次成功 invocation 恰好一次计划内 Gateway 调用。
- v6 Run / History / Comparison / Evaluation / Operations 只保存 metadata；取消、幂等与 `outcome_unknown` 不 replay。

## 批次 E：Web、双数据库连续链与专题关闭

状态：`blocked_by_batch_d`。

- 完成 Template 创作、版本、binding、候选源码审查、assignment、受控测试与 Run / Evaluation handoff。
- SQLite / PostgreSQL launcher profile、重启恢复和真实浏览器连续链必须分别验收。
- 浏览器审查覆盖 application / identity / lifecycle drift、CAS 冲突、批准未激活、显式激活、取消、URL / storage / console / network 与敏感扫描。
- 同步功能专题、任务卡、能力矩阵、current focus、集成契约和周志，执行全量仓库门禁后关闭专题。

## 兼容矩阵

| 现有能力 | 保持方式 | Prompt 新路径 |
| --- | --- | --- |
| Configuration Draft v1 | 原样读写 | 不允许 template ref |
| Configuration Draft v2 | 原样承载 RAG binding | 不允许 template ref |
| Configuration Draft v3 | 新增 strict consumer | 仅 Prompt ref，RAG ref 必须为空 |
| Publish Candidate v1/v2 | 原 review、晋级与读取语义 | 不解释为 Prompt authority |
| Publish Candidate v3 | 新增 exact template reload | 继续复用唯一 review 状态机 |
| Application Session / Turn v1 | Workflow v5 / RAG v4 不变 | 不接受 Prompt profile |
| Application Session / Turn v2 | 新 strict consumer | Prompt profile 与 authority v2 |
| Runtime Authority v1 | Workflow / RAG 不变 | 不增加字段 |
| Runtime Authority v2 | 新 strict consumer | Prompt exact authority |
| Workflow Run v0–v5 | 存储、History、Comparison、Evaluation 兼容 | 不改写、不迁移 |
| Workflow Run v6 | 新 lineage 与 strict consumer | 只用于 Prompt invocation |
| API key 既有 scopes | 权限与 northbound 路由不变 | 不隐式获得 Prompt 调用权 |
| `prompt_application:invoke` | 新 scope、默认未授予 | 只调用当前 assignment |

## 总停止线

- 不把 Template owner 合并到 Configuration Draft、Publish Candidate、Session、Run、Gateway 或 Operations。
- 不新增第二套应用审批状态机，不因 candidate approved 自动 binding、activation、assignment、release 或 invocation。
- 不支持完整 Jinja / Handlebars、任意代码、文件 / 环境 / 网络读取、动态工具、RAG、HTTP Tool、agent loop 或业务写回。
- 不实现自动 retry / fallback、schedule、replay / resume、长期记忆、quota、billing、生产认证、生产 secret、生产 repository 或生产声明。
- 不下载模型、不接真实外部 provider 账户，不把 fake / mock / memory / dev database 写成生产可用。
- 不为本专题新增同层 readiness 链；专项 checker 只有在现有单元、集成与聚合门禁无法承载高风险证据时才允许增加。

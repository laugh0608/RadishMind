# Agent / Copilot 应用档案版本审查与受控建议（开发 / 测试态）v1

更新时间：2026-07-25

状态：`agent_copilot_application_profile_version_review_controlled_suggestion_dev_test_v1_batch_d_completed_batch_e_ready`

## 功能定位

本专题把 Application Catalog 中已经允许声明的 `agent` 类型建设为真实可用的 Agent / Copilot 应用能力。内部应用开发者可以为一个 Agent 应用维护结构化 Copilot Profile、生成不可变版本、把精确版本纳入应用配置与发布候选审查，并在人工批准和显式运行时激活后，通过 Application API key 或 Application Interaction Session 发起一次受控建议调用。

首版产品语义是“面向特定项目与任务的 Copilot 建议应用”，不是自治 Agent。每次调用只允许一次计划内 Gateway / provider 委托；系统不自行规划多步循环、不执行工具、不写回业务真相源，也不根据模型结果自动重试、发布或应用候选动作。`CopilotResponse.proposed_actions` 仍是候选建议，高风险动作必须保留 `requires_confirmation`。

## 现状、根因与本轮决策

- 专题启动前，Application Catalog 和 Application Configuration Draft 已允许 `application_kind=agent`，但没有 Agent 专属的应用档案 owner、不可变版本、发布候选引用、运行时 assignment 或 invocation service；当前批次 A 至批次 D 已补齐 Profile、配置、发布审查、assignment、单次 invocation、Session / Turn v3、Run v7 与 metadata-only 审查链，类型专属 Web 仍待批次 E 实现。
- Application Development Workspace 目前只区分 `prompt_application` 与“非 Prompt”；`agent` 因而继承 Workflow RAG 的配置和晋级界面。这不是 Agent 能力复用，而是应用类型与产品界面语义不一致。
- 仓库已经具备 canonical `CopilotRequest / CopilotResponse`、Gateway、应用目录、配置草案、发布审查、API key、Application Interaction Session、Run History、Comparison、Evaluation 和双数据库开发测试态持久化基础。新专题应组合这些 owner，不复制协议适配、会话、运行记录或评测算法。
- 产品范围与路线图已经长期承诺 Agent / Copilot 应用。本轮选择该方向，优先补齐用户可识别的应用能力，不转向缺少可信 usage 的计费、不提前接真实 OIDC，也不在 backend / artifact store 尚未成立时扩图片生成运行时。

因此，功能设计先固定长期边界，再由[唯一实施任务卡](../../task-cards/agent-copilot-application-profile-version-review-controlled-suggestion-dev-test-v1-plan.md)冻结版本、API、权限、失败语义、兼容矩阵和分批验证。批次 A 至批次 D 已完成 strict contracts、canonical policy compiler、三种 Profile owner、Configuration / Candidate v4、显式 assignment、唯一受控建议 service、Session / Turn v3、Run v7 与 metadata-only 审查链；candidate approve 不自动激活，调用在 provider 前持续重验 exact authority。下一步只进入批次 E 类型专属 Web 与双数据库真实浏览器连续验收。

## 目标用户与主要任务

首批目标用户是内部应用开发者、应用审查人和开发测试态调用者：

1. 开发者在 Application Catalog 中创建或选择 `agent` 应用。
2. 开发者创建或恢复结构化 Copilot Profile Draft，选择目标项目、允许任务、默认语言、上下文 / 产物策略、响应策略和风险策略。
3. 系统根据 canonical contract 执行确定性校验；未知项目、越界任务、互斥策略、敏感材料和不受支持的执行声明必须失败关闭。
4. 开发者从已保存且校验通过的精确草案版本生成不可变 Copilot Profile Version。
5. 开发者把精确 Profile Version 引用附着到新的 Application Configuration Draft 版本；配置草案不复制 profile source。
6. Application Publish Candidate 绑定精确配置与 Profile Version，审查人读取结构化源码、摘要和兼容结果后执行既有人工审查状态机。
7. 候选批准不自动激活。具备独立权限的参与者显式创建、替换或撤销 Agent / Copilot Runtime Assignment。
8. API key 或 Application Interaction Session 提交 canonical request data；唯一 invocation service 重读 exact authority，确定性构造 `CopilotRequest`，并只委托一次既有 Gateway。
9. 当前响应返回调用方；Run History、Comparison、Evaluation 和 Application Operations 只保存和消费 metadata-only 运行证据，不恢复原始输入、上下文、产物正文或完整回答。

## 领域所有权

| 领域 | 真相源 / owner | 本专题职责 | 明确不承担 |
| --- | --- | --- | --- |
| 应用身份与生命周期 | Application Catalog | 确认作用域、`agent` 类型、当前 revision 和生命周期 | 不保存 Profile、assignment 或运行输入 |
| Agent / Copilot 源码 | 新增 Copilot Profile owner | 保存结构化草案、CAS、校验结果和不可变版本 | 不保存 provider credential、运行正文、发布决定或业务状态 |
| 应用公开运行配置 | Application Configuration Draft | 保存协议、模型和精确 Profile Version 引用 | 不复制 Profile source，不保存上下文或回答 |
| 应用发布审查 | 既有 Application Publish Candidate owner | 绑定精确配置与 Profile Version，承担唯一人工审查状态机 | 不复制 Profile source，不因批准自动激活 |
| 运行时激活 | 新增 Agent / Copilot Runtime Assignment owner | 保存当前 approved candidate 与 Profile Version 的 ref-only 指针、CAS 和事件 | 不调用 provider，不复制配置或 Profile |
| 受控建议调用 | 新增唯一 Agent / Copilot invocation service | 解析 exact authority、构造 canonical request、委托一次 Gateway、校验 canonical response | 不实现 agent loop、工具执行、业务写回或 provider fallback |
| 协议与模型调用 | 既有 canonical contracts 与 Gateway | 校验请求 / 响应、适配协议、选择 provider / profile / model、执行上行调用 | 不管理 Profile、候选或 assignment |
| 交互编排 | Application Interaction Session | 通过显式 Agent profile 委托唯一 invocation service，维护 metadata-only Session / Turn | 不持久化 transcript，不复制 invocation 或 Gateway 算法 |
| 运行证据 | Workflow Run Store / Application Operations | 保存 authority、状态、diagnostic、usage availability 与副作用计数 | 不保存原始 context、artifact content、完整 response 或 provider raw response |
| 质量评测 | 既有 Evaluation owner | 对 exact Profile lineage 和明确任务做只读比较与开发测试态评测 | 不修改 Profile、候选、assignment 或运行记录 |

## Copilot Profile 设计边界

### 结构化源码

Copilot Profile Draft 首版在逻辑上需要表达：

- `project`：只允许 canonical contract 已支持的 `radishflow | radish`；
- `allowed_tasks`：只能从所选项目对应的 canonical task 集合中选择；
- `default_locale` 与允许语言策略；
- `context_policy`：允许的 context 字段类别、必填约束和大小预算；
- `artifact_policy`：允许的 artifact kind / role、数量、大小与引用边界；
- `response_policy`：允许的 answer、issue、citation 与 proposed action 类型；
- `risk_policy`：哪些任务或候选动作必须 `requires_confirmation`，以及何时禁止返回可执行形态；
- `read_only_tool_hints_policy`：只允许表达 canonical request 中的提示，不授予工具执行权限；
- 描述、版本、校验状态和脱敏审计引用。

首版不把 provider / profile / model credential、自由形式系统提示词、运行输入、artifact content、完整 context 或业务权限写入 Copilot Profile。协议与默认模型继续属于 Application Configuration Draft；运行输入仍由每次调用提供并按 Profile 与 canonical contract 校验。

### Canonical task 约束

Profile 不创建第二份开放式 task 注册表。首版只能消费 `copilot-request.schema.json` 已固定的任务：

- `radishflow`：`explain_diagnostics`、`suggest_flowsheet_edits`、`suggest_ghost_completion`、`summarize_selection`、`explain_control_plane_state`、`inspect_canvas_snapshot`；
- `radish`：`answer_docs_question`、`summarize_doc_or_thread`、`suggest_forum_metadata`、`explain_console_capability`、`interpret_attachment`。

项目与任务不匹配、请求 task 不在 Profile allowlist、context 不满足 task 条件或 artifact 超出策略时，必须在创建 provider 副作用前失败关闭。未来新增 task 必须先演进 canonical contract 与评测，不允许只在 Profile 或 Web 中单独放宽。

### 不可变版本

只有已保存、校验状态为 `valid` 的精确 Copilot Profile Draft 版本可以生成不可变 Profile Version。版本至少保存规范化结构化源码、来源草案版本、服务端摘要、创建参与者、时间、请求与审计引用。

不可变版本创建后不能修改或删除。后续变更必须通过草案 CAS 生成新草案版本和新不可变版本。Profile Version 不建立平行审批状态；人工批准、拒绝、要求修改与撤回继续统一由绑定该版本的 Application Publish Candidate 承担。

## 与配置、发布治理和运行时的组合

### Ref-only 配置绑定

后续 Application Configuration Draft 新 schema 版本只允许 `application_kind=agent` 携带精确 Copilot Profile Version 引用。绑定操作必须由服务端重读 Profile Version、校验应用作用域并计算 / 核对摘要，再通过既有 expected-version CAS 生成下一版配置草案。

客户端不能提交 Profile source 或伪造 digest。`prompt_application` 的 v3 binding、Workflow RAG v2 binding 与未绑定 v1 草案继续保持原语义；新版本不能让一个配置同时携带多个互斥可执行源码引用。

### 发布候选审查

后续 Publish Candidate schema 版本复用既有创建、读取、列表、人工 review、supersede 与 eligibility 状态机，只增加 Agent / Copilot 的精确 Profile 引用。创建、审查与 eligibility 读取至少重读：

1. Application Catalog 当前记录、`agent` 类型、生命周期与 revision；
2. 精确 Application Configuration Draft 版本、digest 与 Profile binding；
3. 精确 Copilot Profile Version、digest 与作用域；
4. Profile 的 canonical project / task 兼容结果；
5. 默认协议、模型和现有发布 blocker。

审查面必须读取结构化 Profile source，而不是只看摘要。source 不可读、digest 漂移、作用域漂移、应用类型改变或 canonical contract 不再兼容时，候选不得获准或继续保持运行资格。

### 显式运行时 assignment

Agent / Copilot Runtime Assignment 是当前运行 authority 的 ref-only 指针。只有当前 `approved`、未漂移、类型匹配且 canonical 兼容的候选才能通过显式 `activate | replace` 建立 assignment；`revoke` 只撤销当前指针，不修改候选、配置或 Profile Version。

候选批准不得自动建立 assignment，assignment 写入不得调用 provider。所有变更使用 expected assignment version，事件只追加并保留 actor / request / audit 引用，不保存 Profile source 或运行正文。

## 受控建议调用

后续 implementation 必须建立一个明确命名且唯一的 Agent / Copilot execution profile。Application API key 与 Application Interaction Session 都委托同一个 invocation service：

- 调用者提交 application scope、canonical task、locale、artifacts、context、tool hints、安全输入和 client invocation key；不能提交 Profile、assignment、candidate、provider、credential 或重试策略；
- 服务端在运行预留前解析完整 authority，并在计划内 Gateway 调用前再次重读 Application、assignment、candidate、Configuration Draft、Profile Version 与模型资格；
- 请求经 Profile policy 收窄后构造成 `CopilotRequest`，响应必须通过 `CopilotResponse` 校验；
- 每次调用只允许一次计划内 Gateway / provider 副作用；同步或并发幂等重试只读取已有 running / terminal evidence；
- 输出不符合 canonical contract 时直接失败，不自动修复、不追加第二次 provider 调用，也不降级为未校验文本；
- 取消映射为明确 canceled 终态；provider 结果不确定或终态写入失败映射为 `outcome_unknown`，不得自动 replay。

`CopilotResponse.proposed_actions` 只作为候选动作返回。任何 `requires_confirmation=true` 的动作都不得由本专题执行；首版即使返回 `apply` 描述，也不能把它解释为已授权命令、工具调用或业务写回。

## 数据、隐私与运行证据

Profile owner 可以保存结构化策略源码，但不得保存 credential、token、header、cookie、DSN、运行输入或外部系统业务内容。Configuration Draft、Publish Candidate 与 Runtime Assignment 只保存精确 ref、digest 和必要审计 metadata。

原始 artifacts、artifact content、context、tool hints、渲染后的 Gateway 输入、完整 `CopilotResponse` 和 provider raw response 只存在于当前请求内存与当前响应交接中。Session、Turn、Run、Operations 与 Evaluation 只持久化：

- execution profile、application / assignment / candidate / draft / Profile refs 与 digest；
- project、task、locale、输入摘要 / 字节数、artifact / context 分类摘要；
- requested / selected protocol、provider、profile 和 model；
- started / completed time、status、failure code、diagnostic 与 usage availability；
- provider、tool、retrieval、confirmation、business write、replay 副作用计数；
- request、audit 和 actor ref。

历史读取、幂等重试、比较和评测不得根据 metadata 伪造原始回答或 transcript。Web 的应用切换、revision 变化、归档、身份变化和 surface 卸载必须清除未提交输入、当前回答和迟到请求；URL 与浏览器存储不得保留这些内容或一次性 credential。

## 兼容审计

后续任务卡在冻结新 schema 或路由前，必须先形成兼容矩阵并用相邻测试证明：

- Application Catalog 既有 `workflow_copilot | docs_qa | agent | prompt_application` 读取、筛选、更新与归档语义不变；
- Application Configuration Draft v1 / v2 / v3 继续按未绑定、Workflow RAG binding、Prompt Template binding 的原语义读取和校验；
- Publish Candidate v1 / v2 / v3、既有 review transition 与 supersede 语义不变；
- Application Interaction Session / Turn v1 / v2 和既有 Workflow Definition、Application RAG、Prompt Application profiles 不原地放宽；
- Workflow Run Record v0–v6、History、Detail、Comparison、Evaluation 与 Operations 继续严格识别原 lineage；
- 既有 API key scopes 不因 `agent` 类型自动获得新的 invocation 权限；
- Application Development Workspace 必须按 application kind 使用显式 surface routing，不能继续用“Prompt / 非 Prompt”二分法让 `agent` 继承 Workflow RAG owner；
- canonical `CopilotRequest / CopilotResponse` 与 Gateway envelope 仍是唯一 northbound / internal 语义来源，不建立 Agent 私有响应格式。

## 实施批次

### 批次 A：高风险任务卡、兼容矩阵与 memory owner

- A1 已完成唯一实施任务卡、兼容审计、10 份 strict schema、独立 Go contract-only codec 与旧版本防漂移测试。
- A2 已完成 Profile canonical normalization；task、context field、artifact kind / role 和 action kind 直接投影 canonical `CopilotRequest / CopilotResponse` 并由相邻 schema 对照测试防止形成第二套 registry。
- A2 已完成纯函数 policy compiler，输出规范化 source、稳定 `profile_digest / policy_digest / allowed_tasks_digest`；Profile Draft / Version codec 会复算 digest 并拒绝非 canonical source 或 digest drift。
- A2 已完成 UTF-8 source / locale 与策略数值预算、原始条目数量、project / task / context 可满足性、固定 advisory / confirmation / tool hints、敏感字段和配置形态守卫。
- A3 已完成 tenant / workspace / application / owner 作用域的 memory Draft / Version repository、原子 CAS、不可变版本、稳定列表、损坏 / 不可用失败关闭，以及默认关闭 API。
- A3 的 validate 零写入；save / version create 只写 Profile owner，并在写入前重读 active `agent` Application Catalog。摘要读取与完整 source 读取分别要求 `agent_copilot_profiles:read` 和 `agent_copilot_profiles:read_source`。
- 批次 A 关闭时，Configuration / Candidate v4、Assignment、Session、Run、`agent_copilot:invoke` 与 Gateway / provider 调用均未注册；后续注册状态以对应批次记录和本文“当前下一步”为准。

### 批次 B：SQLite / PostgreSQL 开发测试态持久化

- 状态：`completed`。
- Profile owner 已使用独立 SQLite / PostgreSQL `0001_agent_copilot_profiles` migration family；SQLite 复用聚合本地产品 runtime，PostgreSQL 复用既有 migration / pool 模式，没有新增 DSN / pool 管理层。
- memory / SQLite / PostgreSQL 已统一 Draft CAS、immutable Version、作用域、稳定列表、digest 重算、损坏失败和 no-fallback 语义；显式 `memory_dev | sqlite_dev | postgres_dev_test` 配置、启动 marker 校验和关闭链路已接通。
- 相邻与真实 PostgreSQL 门禁覆盖 migration / rollback / reapply、marker / checksum、受限运行角色、并发单写者、服务重建、跨作用域隔离、不可变触发器、损坏数据与数据库敏感材料扫描。
- Assignment、Session / Turn v3 与 Run v7 未在本批实现；它们继续按任务卡在后续批次复用共享 Workflow runtime migration。

### 批次 C：配置绑定、发布审查与显式 assignment

- 状态：`completed`。
- Configuration Draft v4 已启用独立 Profile binding，写入只保存 exact `profile_id / profile_version / profile_digest / policy_digest`，与 RAG / Prompt binding 严格互斥；绑定要求 draft write 与 `agent_copilot_profiles:bind`。
- Publish Candidate v4 复用唯一 create / read / list / review / supersede 状态机；创建、审查和运行资格读取均重读精确 Profile Version，源码审查额外要求 `agent_copilot_profiles:read_source`，digest drift 与 supersede 失败关闭。
- Runtime Assignment memory / SQLite / PostgreSQL owner 支持 `activate | replace | revoke`、expected-version CAS、只追加事件和 read-time exact authority 重验；candidate approve 不自动激活，已撤销 assignment 不原地恢复。
- SQLite / PostgreSQL 共享 Workflow runtime 前滚到 `0014_agent_copilot_runtime_assignments` / `0017_agent_copilot_runtime_assignments`；真实 PostgreSQL 已覆盖迁移、受限角色、并发 CAS、重启、损坏守卫和敏感材料扫描。

### 批次 D：API key、Session、单次调用与审查链

- 状态：`completed`。
- 唯一 `agent_copilot_suggestion_v1` service 已完成运行预留前 authority 解析和 Gateway 前 exact authority checkpoint；输入按 Profile policy 收窄为 canonical `CopilotRequest`，响应严格校验 canonical `CopilotResponse`、引用关系、UTF-8 byte budget 与候选动作确认。
- API key `agent_copilot:invoke` 已独立注册，既有 scope 不隐式继承；Application Session / Turn v3 只委托同一 service。并发重复只读取 running / terminal evidence，不恢复回答或触发 retry / replay。
- `workflow_run_record.v7` 与独立 Agent Session / Turn / Run 投影已启用；History、Detail、`workflow_run_comparison.v6`、Evaluation 与 Operations 显式识别同一 Profile / project / task lineage且只消费 metadata。
- SQLite / PostgreSQL 共享 Workflow runtime 前滚到 `0015_agent_copilot_invocation_projections` / `0018_agent_copilot_invocation_projections`；真实 PostgreSQL 已验证受限运行角色、重启恢复、append-only、运行正文不落盘和 configured startup。
- 相邻测试覆盖权限、严格 body、并发幂等、取消、transport 不确定、authority drift、canonical response / confirmation / citation 失败、成功调用恰好一次 Gateway 副作用，以及旧 Session / Run / API key 行为不漂移。

### 批次 E：类型专属 Web 与双数据库连续验收

- 为 `agent` 建立独立 Profile、版本、binding、候选审查、assignment、受控测试、Session 和 Run / Evaluation handoff。
- 把 Application Development Workspace 从 Prompt / 非 Prompt 二分改为显式 application kind surface routing，不让不适用 owner 伪装成可用。
- 使用 SQLite 与 PostgreSQL 各完成一条连续真实浏览器链，并验证服务重启、CAS、authority drift、取消、迟到响应和 transient 清理。
- 复验 URL、浏览器存储、console、network、SQLite 和 PostgreSQL 中不存在 credential、原始输入、artifact content、完整回答或 transcript 泄漏。

## 验收标准

- 产品真实性：`agent` 类型拥有专属 Profile 与调用路径，不再显示 Workflow RAG 专属创作和晋级 owner。
- 所有权：Profile、Configuration Draft、Publish Candidate、Runtime Assignment、Session、Run、Evaluation 和 Gateway 职责独立，没有复制 source 或执行算法。
- 兼容性：既有配置、候选、Session、Run、API key、Workflow / RAG、Prompt Application 和 Gateway 行为保持原语义。
- 安全性：所有漂移在 provider 副作用前失败关闭；一次调用最多一次计划内 Gateway 副作用；候选动作不被自动执行。
- 持久化：memory / SQLite / PostgreSQL 在作用域、CAS、不可变版本、重启和 no-fallback 上语义一致。
- 隐私：durable owner 与浏览器持久介质不保存原始 context、artifact content、完整响应、transcript 或 credential。
- 可复验性：相邻单元 / 集成测试、Go / Web 验证、双数据库 launcher、真实浏览器连续链、快速与全量仓库门禁形成完整证据。

## 当前下一步

下一步直接进入批次 E：把 Application Development Workspace 从 Prompt / 非 Prompt 二分改为显式 application kind routing，为 `agent` 挂载 Profile、版本、binding、候选源码审查、assignment、受控测试、Session v3 和 Run / Evaluation handoff。随后补齐 SQLite / PostgreSQL launcher 与两条真实浏览器连续链，验证服务重启、CAS、authority drift、取消、迟到响应和 transient 清理；不扩 agent loop、工具执行、retry / replay 或生产能力。

## 停止线

- 不实现自治规划、agent loop、多轮自驱执行、tool executor、connector、在线搜索、业务写回、自动 confirmation 或自动应用候选动作。
- 不实现 provider retry / fallback、负载均衡、schedule、replay / resume、长期记忆、quota、billing、cost ledger 或生产能力声明。
- 不引入真实 Radish OIDC、workspace membership、外部 provider 账户、生产 secret backend 或生产 API key。
- 不修改 `CopilotRequest / CopilotResponse` 来绕过现有项目、任务、风险和确认边界；新增 canonical task 必须另行设计与评测。
- 不把 Profile Version、approved candidate 或 assignment 中任一状态单独解释为正式发布；运行资格必须由 exact authority 联合决定。
- 不从已关闭的 Prompt Application、Application RAG 或 Workflow Definition 专题原地复制一套 Agent 执行器。

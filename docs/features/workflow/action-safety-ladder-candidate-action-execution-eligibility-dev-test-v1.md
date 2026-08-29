# Action Safety Ladder 与候选动作执行资格（开发 / 测试态）v1

更新时间：2026-08-29

状态：`action_safety_ladder_candidate_action_execution_eligibility_dev_test_v1_design_approved_implementation_pending`

## 功能定位

本专题把战略文档中的 `Action Safety Ladder` 从抽象原则推进为可审查、可失败关闭、可进入既有执行边界的开发 / 测试态产品能力。内部应用开发者和审查人可以看到一个回答、候选动作或受控 Tool 计划实际请求什么能力、当前规则最多允许什么能力、最终生效到哪一级，以及哪些 blocker 阻止它继续提升。

首版只负责候选动作的执行资格，不创建通用动作执行器。模型、客户端、Prompt、Profile、Workflow 草案、候选审查结果或人工批准都不能直接授予执行权限；确定性规则层必须在每个高风险检查点重读精确 authority，并把结果收口为 `requested_level / maximum_allowed_level / effective_level`。既有 Workflow HTTP Tool 是唯一可进入 `tool_callable` 的执行路径，而且仍必须经过其独立 action plan、人工 confirmation、原子 claim 与单次 allowlisted `GET`。

本专题状态已由项目所有者批准为“设计完成、实现待单独批准”。当前只建立功能专题与唯一高风险任务卡，不修改 canonical schema、Go runtime、数据库、HTTP、React 或 Pencil。

## 用户价值与目标用户

目标用户是内部应用开发者、Workflow Builder、人工审查人和运行证据审查人。当前 `risk_level`、`requires_confirmation`、Profile 风险策略、Tool confirmation 与 Run 副作用计数分别存在，但用户无法从同一条产品链判断“建议为什么只能看、为什么可以交接、为什么可以调用工具、为什么写入被阻断”。

首版提供三项用户价值：

1. 在 Builder / Profile / Candidate Review 阶段预览当前动作能力上限和阻塞原因，避免把 `valid`、`approved` 或 `requires_confirmation=true` 误读为可执行。
2. 在显式 activation、action plan 创建和 Tool 执行前由服务端重新计算资格，防止旧摘要、客户端状态或人工批准绕过当前 authority。
3. 在 Run History / Detail 中读取当时冻结的规则版本、有效级别与副作用证据，解释一次动作为什么执行、为什么失败关闭或为什么只能交给人工继续处理。

## 首版安全梯度

安全级别是带显式转换规则的状态集合，不是客户端可比较的数字，也不能只靠字符串顺序决定是否晋级。

| level | 首版语义 | 允许结果 | 禁止结果 |
| --- | --- | --- | --- |
| `answer_only` | 只能回答、解释或给出问题说明 | answer、issue、citation | proposed action、handoff、tool plan、外部副作用 |
| `proposal_only` | 可以返回候选建议，但不能形成可执行资格 | canonical proposed action、风险与确认提示 | 命令、action plan、execution token、tool dispatch、业务写入 |
| `handoff_ready` | 可以把经过脱敏和作用域绑定的结构化候选交给既有人工 owner | ref-only handoff、目标 owner 重新读取 | 自动审查、自动确认、自动计划、自动执行 |
| `tool_callable` | 可以进入既有受控 Tool action plan / confirmation / execution owner | 当前唯一为人工确认后的只读 HTTPS `GET` | 任意 URL、写方法、多工具、自动确认、retry / replay |
| `write_blocked` | 请求包含业务写入、命令提交或其它 v1 禁止能力，规则层明确阻断 | blocker、失败码、人工解释 | action plan、execution、业务 truth mutation |
| `write_allowed_by_policy` | 为未来规则明确允许写入保留的协议级状态 | v1 中不可达 | v1 中的任何成功判定、持久授权或执行 |

`requested_level` 表示根据 source kind、action shape、目标类型和副作用声明由服务端推导出的能力需求，不接受模型或客户端自报。`maximum_allowed_level` 表示当前精确 Profile、Application、Definition、Tool policy、membership 与环境 gate 联合允许的最高正向资格。`effective_level` 是规则状态机的最终结果；请求含写入而 v1 无写入 policy 时必须为 `write_blocked`，不能简单降级成 `proposal_only` 后隐藏原始风险。

## 首版兼容矩阵

| 当前来源 | 可达结果 | 资格条件 |
| --- | --- | --- |
| 普通 Prompt / RAG / Copilot 回答且无候选动作 | `answer_only` | canonical response 合法，未声明动作或外部副作用 |
| Agent / Copilot canonical proposed action | `proposal_only` | action kind、risk、confirmation 与 source refs 均合法；不生成计划 |
| 结构化候选交给既有人工审查面 | `handoff_ready` | handoff shape 在 allowlist，完整 scope 与 exact source refs 可重读，目标 owner 明确接受，载荷只含脱敏 metadata |
| 既有 Workflow HTTP Tool action | `tool_callable` | exact Definition / Tool definition / runtime profile / plan digest / confirmation / membership / execute grants / dev gate 均通过，并在 dispatch 前再次复核 |
| `POST / PUT / PATCH / DELETE`、业务真相写入、shell、code、sandbox、agent loop、connector mutation、自动 apply | `write_blocked` | 无例外；人工批准也不能改变结果 |
| 任意来源申请 `write_allowed_by_policy` | `write_blocked` | v1 没有可用 policy row、permission、执行器或生产 gate |

`handoff_ready` 不是通用命令包。只有已经存在、职责明确且会自行重读 exact resource 的人工 owner 才能作为 handoff 目标；上层项目尚无真实挂载点时不得用假想 command 或 callback 冒充该级别。

## 规则所有权与现有 owner

本专题新增的是纯规则职责，不新增第二套领域真相源：

| 领域事实 | 唯一 owner | Safety Ladder 如何消费 |
| --- | --- | --- |
| canonical response 与 proposed action | `CopilotResponse` contract / response builder | 读取规范化 action shape；不接受模型自报有效级别 |
| Agent / Copilot Profile 与风险策略 | Copilot Profile owner | 重读 exact version、policy digest、allowed task / action 与 confirmation 约束 |
| Application 配置、候选、assignment | 各自既有 owner | 重读 application kind、candidate decision、eligibility、exact binding 与当前 assignment |
| Workflow Definition 与 activation | Definition release / activation owner | 重读 exact immutable source、profile、digest 与当前 activation |
| HTTP Tool definition / profile / plan / confirmation | Workflow HTTP Tool owner | 复用唯一 action plan、confirmation、claim、transport 与 audit |
| 身份、permission、workspace membership | local identity / membership provider | 在每个 mutation / execution checkpoint 重新授权 |
| Run 与副作用计数 | Workflow Run Store | 保存当时的 safety snapshot / refs 与真实 side-effect counters |
| 页面组合与 readiness | Application Development Workspace / Workflow Workbench | 只读组织 decision，不重新计算或持久化第二份资格 |

`ActionSafetyPolicyCompiler` 只接受上述 owner 返回的严格、脱敏、版本化输入，输出确定性 decision。它不能修改来源资源、创建 confirmation、发起 Tool、写 Run 或替代 permission decision。若未来 decision 需要持久化，只能作为既有 candidate、action plan 或 Run 的不可变 snapshot / ref 进入对应 owner；不得创建通用 `action_safety_decisions` 主表、平行 audit store 或新的跨 store aggregate owner。

## Canonical decision 边界

后续批次 A 需要版本化 `action_safety_decision.v1`，至少包含：

- `schema_version`、`decision_id` 或 owner 内稳定引用；
- tenant、workspace、environment、application 的 exact scope；
- `source_kind / source_id / source_version / source_digest`；
- canonical task / action kind 与目标类别；
- `requested_level / maximum_allowed_level / effective_level`；
- `requires_confirmation / confirmation_state`；
- `writes_business_truth` 与允许的副作用上限；
- 有序、去重、稳定的 `blockers`；
- `policy_version / policy_digest`；
- actor、request、audit 与 UTC 时间引用。

decision 不保存 input、answer、prompt、context、artifact content、Tool arguments 原文、URL、header、credential、token、cookie、provider raw response、remote error、业务写入载荷或完整候选正文。严格 decoder 必须拒绝 unknown / duplicate field、非法 level、非 canonical blocker 顺序、scope 漂移、digest 漂移与不受支持的 source kind。

## 资格计算与重新验证

规则编译必须按以下顺序失败关闭：

1. 校验 strict payload、source kind 与 exact scope。
2. 重读来源 owner，复算 source digest，拒绝 missing、stale、superseded、archived 或 contract drift。
3. 重读 Application / Profile / Definition / Tool 等当前 authority，构造不可伪造的 capability demand。
4. 读取当前 policy version、环境 gate、membership 与 operation grants，计算 `maximum_allowed_level`。
5. 通过显式 compatibility / transition matrix 计算 `effective_level`；禁止只比较枚举顺序。
6. 输出稳定 blocker 与副作用预算；任何规则缺失、矛盾或不可读都降低资格或进入 `write_blocked`，不使用默认允许。
7. 在形成持久候选、activation / assignment、action plan、Tool dispatch 和 Run snapshot 前按对应检查点再次执行；旧 decision 只能用于审查，不能作为当前授权。

必须覆盖以下检查点：

- response normalization：模型输出校验后、返回 consumer 前；
- candidate review：人工批准前和 decision 提交事务内；
- activation / assignment：写入当前 runtime pointer 前；
- action plan creation：生成不可变 Tool plan 前；
- pre-dispatch：既有 Tool 原子 claim 与网络尝试前；
- run projection：终态只记录实际使用的 policy snapshot 与 side-effect counters。

人工 review、candidate approve 或 confirmation 只满足各自 owner 的一个前置条件。它们不能提升 `maximum_allowed_level`，不能把 `write_blocked` 改为 `tool_callable`，也不能让 `write_allowed_by_policy` 在 v1 可达。

## API、权限与持久化边界

- 首版不增加通用 `/actions/execute`、`/action-safety/decisions` mutation 或万能 execution permission。
- 现有 candidate / assignment / Definition / Tool route 通过版本化 contract 消费 decision 或 decision summary；每个 owner 继续使用自己的 read / review / activate / plan / confirm / execute permission。
- `tool_callable` 仍要求 `workflow_tool_actions:execute + workflow_runs:execute + workflow_drafts:read`，并保持 confirmation grant 独立；Safety Ladder 不新增能够覆盖这些权限的上位 grant。
- memory、SQLite、PostgreSQL 若需要保存 snapshot，只扩展已有 owner 的版本化记录和 migration family；三种实现保持同构、原子、重启可读与 no-fallback。
- 历史记录没有 safety snapshot 时必须明确显示 `not_recorded_legacy` 或对应版本兼容状态，不能按当前 policy 反算并伪造历史决定。

## 失败语义

| failure code | 语义与副作用 |
| --- | --- |
| `action_safety_scope_denied` | tenant / workspace / application / actor 不匹配；owner 与外部副作用为 0 |
| `action_safety_payload_invalid` | strict payload、source kind 或 level 非法；不创建 decision snapshot |
| `action_safety_source_unavailable` | exact source 不存在或 owner 不可用；fail closed |
| `action_safety_source_changed` | source version / digest / lifecycle 漂移；拒绝当前 transition |
| `action_safety_policy_unavailable` | policy version 或兼容矩阵不可读；不得默认允许 |
| `action_safety_policy_changed` | 审查后、激活前、计划前或执行前 policy digest 漂移；要求重新审查 |
| `action_safety_level_escalation_denied` | 请求能力超过当前最大允许级别；不生成更高权限资源 |
| `action_safety_confirmation_required` | 需要既有 confirmation 但当前没有有效 approval；不 claim、不 dispatch |
| `action_safety_confirmation_changed` | confirmation scope、plan digest 或状态漂移；不 dispatch |
| `action_safety_tool_authority_unavailable` | Tool definition / profile / plan / grants / gate 不满足；网络为 0 |
| `action_safety_write_blocked` | 检测到写入或 v1 禁止动作；业务写入、Tool 与 Run 执行为 0 |
| `action_safety_store_contract_mismatch` | 既有 owner 中的 snapshot / ref 损坏；不回退 memory、fixture 或当前 policy 推算 |

失败响应与审计只返回稳定 code、policy ref、source ref 和脱敏 blocker，不返回底层存储错误、模型正文、Tool 参数、endpoint 或权限详情。

## 副作用不变量

| effective level | provider | handoff | Tool / network | confirmation consumption | business write | replay |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| `answer_only` | 按既有回答路径上限 | 0 | 0 | 0 | 0 | 0 |
| `proposal_only` | 按既有建议路径上限 | 0 | 0 | 0 | 0 | 0 |
| `handoff_ready` | 不新增 | 最多 1 个 ref-only 交接 | 0 | 0 | 0 | 0 |
| `tool_callable` | 按既有 HTTP Tool DAG 上限 | 0 | 最多 1 次 allowlisted GET | 恰好 1 个已批准计划被消费 | 0 | 0 |
| `write_blocked` | 不为被阻断动作新增 | 0 | 0 | 0 | 0 | 0 |
| `write_allowed_by_policy` | v1 不可达 | 0 | 0 | 0 | 0 | 0 |

普通回答或建议路径已经发生的单次 Gateway 调用不因后处理 Safety decision 被重复调用。任何资格重算、刷新、重启、CAS 冲突、policy drift 或历史读取都不得 replay provider、Tool 或 confirmation。

## Web 产品面

首版不建立 S11。UI 复用 Application Development Workspace、Workflow Workbench、Human Promotion、Controlled Test 与 Run / Evaluation Review 的既有 application / workspace context：

- Builder 显示当前 source 的 requested / maximum / effective level、policy version 与 blocker，不把 validation 或 risk summary当作执行授权。
- Reviewer 在 candidate / activation / action plan 的当前 owner 内审查相同 decision；提交后由服务端重算，CAS / drift 只展示稳定失败，不自动重试。
- Run History / Detail 显示当时冻结的 effective level、confirmation / Tool refs 与真实副作用计数；legacy 记录不回填。
- application / workspace / actor / source 切换用既有 generation + abort 清空 selection、confirmation、decision 与迟到响应。
- URL、Web Storage、IndexedDB、Cache、cookie 和跨标签消息不得保存 decision payload、候选正文、Tool 参数或 confirmation 内容；跨标签只允许 metadata-only invalidation。

完整 Pencil 必须覆盖 Desktop 与不能直接推导的 Narrow，以及 `answer_only`、`proposal_only`、`handoff_ready`、`tool_callable`、`write_blocked`、policy drift、confirmation missing 和 legacy no-snapshot。Pencil 与 Decision Record 必须先由项目所有者人工批准，之后才能进入 React。

## 实施拆分

### 批次 A：strict decision contract 与确定性 policy compiler

- 物化 `action_safety_decision.v1`、level / blocker / source compatibility matrix 与 canonical digest。
- 实现纯函数 compiler、严格 codec、unknown / duplicate / ordering / drift 负向测试和 memory fixture。
- 对照 canonical `CopilotResponse`、Agent Profile、Workflow Definition 与 HTTP Tool contract，证明没有第二套 task / action / permission registry。
- 本批不注册 HTTP、不修改 owner record、不创建 migration、Pencil、React 或运行副作用。

### 批次 B：既有 owner 检查点与开发测试态 runtime 组合

- 将 response normalization、candidate review、activation / assignment、Tool action plan 和 pre-dispatch 接入同一 compiler。
- 版本化扩展现有 candidate / plan / run projection；不建立通用 decision mutation API 或新 store。
- memory owner 证明 scope、CAS、authority / policy drift、confirmation 与零越权副作用。

### 批次 C：SQLite / PostgreSQL 同构证据

- 仅在需要历史 snapshot 的既有 migration family 中追加字段或关联记录。
- 覆盖 migration / rollback / reapply、runtime role、并发 CAS、重启、corruption、no-fallback、legacy read 与 policy snapshot 不重算。
- 证明 `tool_callable` 仍只消费既有 HTTP Tool owner，`write_allowed_by_policy` 无数据库成功路径。

### 批次 D：完整 Pencil 与人工批准

- 在现有产品面完成 Safety Ladder 的 Desktop / Narrow、关键状态与 Decision Record。
- 未获项目所有者人工批准不得开始 React；批准不自动授权批次 E。

### 批次 E：React strict consumer 与双数据库产品连续链

- 实现单一 strict consumer 和 Builder → Reviewer → Controlled Tool / blocked write → Run History 连续链。
- SQLite 页面链、PostgreSQL configured Server、双标签 CAS、服务重启、三视口、隐私与 side-effect audit 全部复验。
- 关闭专题时同步真相源并清理服务、容器、数据库与临时材料；不自动派生批次 F。

每个批次都需要项目所有者单独批准。唯一实施入口是[Action Safety Ladder 与候选动作执行资格 v1 高风险任务卡](../../task-cards/action-safety-ladder-candidate-action-execution-eligibility-dev-test-v1-plan.md)。

## 验收方式

- Contract：strict schema / codec、canonical digest、level transition、blocker ordering、unknown / duplicate / legacy compatibility。
- Domain：source authority reload、policy drift、scope、membership、permission、CAS、confirmation、Tool authority 与 zero-escalation tests。
- Runtime：回答 / 建议不重复 provider；Tool 仍最多一次 GET；所有 write request 在 plan / network / business owner 前失败。
- Persistence：memory / SQLite / PostgreSQL 同构、migration、role、restart、corruption、no-fallback 与历史 snapshot 不重算。
- Web：offline 零请求、strict parser、generation + abort、双标签 stale、Desktop / tablet / mobile、URL / Storage / cookie / response / database 敏感扫描。
- Repository：每批精准测试、`git diff --check`、`./scripts/check-repo.sh --fast`；schema、migration、架构、协议、阶段状态与最终关闭补全量 `./scripts/check-repo.sh`。

## 停止线

- `write_allowed_by_policy` 在 v1 必须不可达；不新增写入 policy、写入 permission、写入 Tool 或业务 owner adapter。
- 不开放 `POST / PUT / PATCH / DELETE` Tool，不允许任意 URL、header、body、credential 或模型生成目标。
- 不实现自动 confirmation、candidate approve 后自动 activation、action plan 自动执行、agent loop、多工具、connector、shell、code、sandbox、schedule、retry / fallback、replay 或 resume。
- 不创建通用 Action、Confirmation、Run、Audit、Result 或 decision store，不复制既有 owner 的 source、状态机或权限。
- 不让模型、客户端、Profile、Workflow 草案、candidate decision、assignment 或页面状态自报或提升 effective level。
- 不写 `RadishFlow`、`Radish`、`RadishCatalyst` 或其它业务真相源，不细化不存在的上层 command handoff。
- 不打开 production OIDC、production secret、production repository、public API、quota、billing、正式发布或 production capability enablement。

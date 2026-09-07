# Action Safety Ladder 与候选动作执行资格（开发 / 测试态）v1 实施任务卡

更新时间：2026-08-30

- 任务 ID：`action-safety-ladder-candidate-action-execution-eligibility-dev-test-v1`
- 状态：`completed`
- 功能设计：[Action Safety Ladder 与候选动作执行资格（开发 / 测试态）v1](../features/workflow/action-safety-ladder-candidate-action-execution-eligibility-dev-test-v1.md)

## 准入结论

项目所有者已批准长期功能目标与首版设计边界：把战略层的动作梯度实现为规则层确定性资格，不让模型、客户端或人工批准直接授予执行能力；`tool_callable` 只复用既有 Workflow HTTP Tool 的人工确认只读 `GET`，`write_allowed_by_policy` 在 v1 始终不可达。

本卡是唯一跨模块高风险实施入口。项目所有者已经逐批批准并完成 A 至 E，不得创建平行 schema-only、UI-only、readiness 或 gate-only 任务卡；本卡现已关闭，不派生批次 F。

## 完成目标

形成可复验的 Builder / Profile → Candidate Reviewer → Activation / Action Plan → 受控 Tool → Run History 产品链，并让写入请求在原候选 / 审查 owner 中明确阻断且不创建 Run。用户能看到服务端推导的 `requested_level / maximum_allowed_level / effective_level`、规则版本、blocker 与真实副作用，同时保持现有领域 owner 和执行停止线不变。

## 不可变边界

1. `ActionSafetyPolicyCompiler` 是纯规则职责，不拥有 Application、Profile、Definition、Tool、Confirmation、Run 或 Audit 状态。
2. requested level 由服务端从 source / action / target / side-effect shape 推导；模型和客户端不能提交 effective / allowed authority。
3. decision 只作为现有 owner 的易失投影或不可变 snapshot / ref；不创建通用 action safety 主表或第二套 audit store。
4. level transition 使用显式兼容矩阵，不用枚举顺序或字符串比较决定授权。
5. `tool_callable` 只进入现有 Workflow HTTP Tool action plan / confirmation / claim / single GET path；不修改通用 `/v1/tools/actions` blocked contract。
6. `write_blocked` 必须保留被阻断的写入意图和稳定 blocker，不能静默降级成看似安全的 proposal。
7. `write_allowed_by_policy` 没有 v1 成功路径、permission、policy row、migration 或 UI action。
8. 历史记录缺少 snapshot 时明确标记 legacy，不按当前 policy 反算历史结论。
9. 所有 scope、authority、policy、membership、permission 与 confirmation 在当前 transition / dispatch 前重读；旧 decision 只供审查。
10. 不持久化 input、answer、prompt、context、artifact content、Tool arguments、URL、header、credential、token、cookie、provider raw response 或业务写入载荷。

## 批次 A：strict decision contract 与确定性 policy compiler

状态：`completed`。

### 允许实现

- 新增版本化 `action_safety_decision.v1` strict schema、Go 类型 / codec 与必要 TypeScript 类型生成或镜像。
- 固定六级枚举、source / action compatibility matrix、transition matrix、blocker allowlist / order 与 RFC 8785 canonical digest。
- 实现纯函数 `ActionSafetyPolicyCompiler`，只接受 caller 已从权威 owner 读取的严格输入，不自行访问 store、HTTP 或 provider。
- 建立 canonical contract 对照测试，证明 task / action kind / risk / confirmation / permission 不形成第二套 registry。
- 实现 memory fixture 和 contract / compiler 精准测试。

### 必须证明

- unknown / duplicate field、非法 level、非法 transition、非 canonical blocker 顺序、scope / source / policy digest 漂移全部失败关闭。
- `answer_only`、`proposal_only`、`handoff_ready`、`tool_callable`、`write_blocked` 的正负矩阵明确；`write_allowed_by_policy` 所有输入都不可产生成功 decision。
- client / model 自报 level、人工批准、candidate approved 或 assignment active 不能单独提升 maximum / effective level。
- 任意 write-capable action 明确进入 `write_blocked`，provider / handoff / plan / confirmation / Tool / Run / business / replay 写入均为 0。
- 既有 `CopilotResponse`、Agent Profile、Workflow Definition、HTTP Tool 与历史 Run schema 回归不漂移。

### 批次 A 停止线

- 不注册或修改 HTTP route，不接 candidate / activation / plan / Run owner。
- 不创建 SQLite / PostgreSQL migration、repository、Pencil、React、CSS、launcher、服务或浏览器证据。
- 不修改通用 Tooling v1 blocked 行为，不发送 provider / Tool / network 请求。

完成后只能推进为 `batch_a_completed_batch_b_ready`，不得自动进入批次 B。

完成证据：`contracts/action-safety-decision.schema.json` 已接入仓库 schema 元校验；Platform 新增 strict decision codec、显式 compatibility / transition matrix 与纯函数 `ActionSafetyPolicyCompiler`。精准测试覆盖五个可达结果、全部 v1 禁止写入类别、四种写方法、scope / source / policy drift、confirmation 缺失 / 漂移、unknown / duplicate / missing field、blocker ordering、digest drift、非法 transition、现有 task / action / risk / Tool registry 对齐和非授权人工状态；完整 `internal/httpapi` 回归、`go vet` 与仓库 fast baseline 已通过。没有 route、store、migration、Pencil、React 或网络副作用。

## 批次 B：既有 owner 检查点与开发测试态 runtime 组合

状态：`completed`。

- response normalization 在 canonical response 校验后编译 decision，不重复 Gateway / provider。
- Candidate Review、activation / assignment、HTTP Tool action plan 和 pre-dispatch 复用同一 compiler，并在各自事务或 claim 前重读 exact authority。
- 版本化扩展现有 candidate / plan / run projection；不创建通用 decision mutation route、万能 execute grant 或新 store。
- memory owner 覆盖 scope、membership、permission、policy / authority drift、CAS、confirmation、duplicate submission、late response 与 zero escalation。
- 本批不创建 SQLite / PostgreSQL migration、Pencil 或 React。

完成后只能推进为 `batch_b_completed_batch_c_ready`。

完成证据：同一 compiler 已接入 Agent Copilot response normalization、易失 candidate review、既有 Runtime Assignment CAS、canonical Definition-bound HTTP Tool action plan、pre-dispatch 和 Workflow Run memory projection 六个检查点。每次 transition / dispatch 都重新读取当前 source、policy、scope、membership、permission、confirmation 或 assignment authority；迟到响应、provider 后 authority 漂移、source / policy drift、confirmation 缺失 / 漂移、stale CAS、重复执行和并发 claim 均失败关闭。内部 projection 统一使用 `json:"-"`，memory owner 深拷贝并校验 projection digest 与真实 side-effect counters；既有 HTTP contract、SQLite / PostgreSQL 编码和 migration 未改变，legacy workflow application identifier 保持无 snapshot 兼容。精准测试、定向 race、完整 `internal/httpapi` 回归均已通过，provider 不重复、Tool / confirmation 最多各一次、business write / replay 始终为 0。

## 批次 C：SQLite / PostgreSQL 同构 snapshot

状态：`completed`。

- 只在确需冻结历史 decision 的既有 candidate / Tool plan / Workflow Run migration family 中追加版本化 projection。
- memory、SQLite、PostgreSQL 共用同一 domain contract，不新增 selector、pool、DSN、database file、跨 store join 或 fallback。
- 覆盖 migration / rollback / reapply、marker / checksum、runtime role、并发 CAS、事务回滚、restart / reconnect、corruption、legacy read 与 no-fallback。
- 历史 snapshot 使用当时 policy version / digest；policy 升级不重算旧记录。
- 数据库必须证明 `write_allowed_by_policy` 零成功记录，blocked write 零 action plan / Tool / Run / business mutation。

完成后只能推进为 `batch_c_completed_batch_d_pencil_ready`。

完成证据：PostgreSQL `0028_action_safety_snapshots / workflow_run_store_v28` 与 SQLite `0025_action_safety_snapshots / workflow_run_store_sqlite_v25` 在五张既有 owner 表内追加 all-or-none 的 `schema_version / projection_digest / sanitized_snapshot`，覆盖 Agent assignment / event、Agent Run v7、Workflow HTTP Tool plan 与 Workflow Run。三种 store 共用严格 projection contract；全空三元组只读为 `not_recorded_legacy`，不按当前 policy 回填，部分三元组、未知 / 重复字段、尾随内容、schema / digest / payload 损坏都失败关闭。既有 transaction / CAS 同时封存 projection，Tool plan transition 额外绑定 projection digest；SQLite restart 与 PostgreSQL reconnect 保留当时 policy version / digest，损坏 Agent Run 的 replay 不回退、不重复 provider。PostgreSQL 真实门禁覆盖 runtime role DML / DDL 边界、rollback → `not_applied` → reapply、marker / checksum、corruption 与 no-fallback；门禁过程中发现并修复 SQL `NULL` 三值逻辑允许部分三元组的缺口。没有新 selector、pool、DSN、database、跨 store join、HTTP contract、Pencil 或 React；`write_allowed_by_policy` 无数据库成功记录，business write / replay 为 0。

## 批次 D：完整 Pencil 与人工批准

状态：`completed_owner_approved`。

- 复用 Application Development Workspace、Workflow Workbench、Human Promotion、Controlled Test 与 Run / Evaluation Review，不建立 S11。
- 完成 Builder、Reviewer、Tool plan / confirmation、Run History 的 Desktop 与不能直接推导的 Narrow。
- 关键状态至少覆盖六级 ladder、policy drift、confirmation missing、scope denied、Tool unavailable、legacy no-snapshot 和 write-blocked danger。
- Decision Record 必须说明信息层级、状态语言、响应式边界、隐私与 owner handoff。
- 项目所有者必须人工审查并明确批准 Pencil；静态 QA、截图或自动检查不能替代人工批准。

项目所有者已于 2026-08-29 完成人工视觉与边界审查并明确批准，状态推进为 `batch_d_pencil_approved_batch_e_ready`；批准不自动授权批次 E。

当前证据：正式设计源 `docs/designs/radishmind-web-family-ui-v1.pen` 已新增 Builder Desktop `JtBu1` / write-blocked Narrow `EIWxV`、Candidate Review Desktop `l3dr1K` / policy-drift Narrow `wyqof`、Tool Plan & Confirmation Desktop `N0jBm` / confirmation-missing Narrow `CoI4i`、Run History Desktop `rHr7a` / legacy Narrow `XRRpD`，以及 R23 Decision `OvCRE`。设计覆盖六级 ladder、scope denied、Tool unavailable、confirmation missing、policy drift、write blocked、legacy no-snapshot、exact pre-dispatch recheck、零副作用 blocker 和 frozen history；R23 五维评分为 `2 / 1 / 2 / 2 / 2 = 9`。9 个根画板共 1055 个节点，Pencil 原生结构检查未发现 placeholder、布局裁切、未命名节点或缺失文字填充。本证据只证明设计完整性，不替代项目所有者人工批准。

## 批次 E：React strict consumer 与双数据库产品连续链

状态：`completed`。

项目所有者已于 2026-08-30 批准本批方案与实施。HTTP 只新增统一的 `action_safety_read_projection.v1` read projection，并嵌入既有 Agent turn、Agent Runtime Assignment、Workflow HTTP Tool plan 与 Workflow Run 响应；不注册 action safety route。projection 使用 owner exact ref、recorded / `not_recorded_legacy`、decision / source / policy digest、三段 level、confirmation、canonical blocker、副作用预算、observed side effects 与 projection digest 的严格 allowlist。Agent response candidate 没有跨请求可重读的独立 owner，本批只在当前 response owner 内易失展示，不增加 candidate mutation / store，不允许客户端提交 decision、authority 或 effective level。

- 实现单一 strict consumer，拒绝 unknown / sensitive field、scope / source / policy drift、duplicate decision、非法 blocker 和 legacy 伪造。
- application / workspace / actor / source 切换使用 generation + abort，清空 selection、decision、confirmation、handoff 与迟到响应。
- Builder → Reviewer → proposal / handoff / controlled Tool → Run History 使用各 owner exact ref，目标 owner必须重新读取；blocked write 停留在原候选 / 审查 owner，不创建 plan、Tool attempt 或 Run。
- SQLite 页面链与服务重启恢复、PostgreSQL configured Server migration / runtime role / no-fallback / reconnect、双标签 CAS 与 policy drift 必须形成连续证据。
- 浏览器覆盖 `1440×900`、`720×900`、`390×844`，检查横向溢出、console / network、URL、Storage、IndexedDB、Cache、cookie、响应与数据库敏感材料。
- 副作用审计证明回答 / 建议不重复 provider，handoff 最多一个 metadata ref，Tool 最多一次既有 GET，所有 business write / replay 为 0。
- 收口时同步专题、入口、当前焦点、路线图、能力矩阵、架构、集成契约、任务卡与周志，并清理服务、容器、数据库和临时文件。

完成后关闭本卡，不派生批次 F、通用动作执行器、写入 policy 或 production 续批。

完成证据：既有 HTTP envelope 新增单一 `action_safety_read_projection.v1` read projection，单一 TypeScript strict consumer 与共享面板接入 Agent Session / Assignment、Workflow HTTP Tool plan / execution 和 Run History；没有新增 route、mutation owner 或执行 token。Web 全量 `410/410`、production build、Go `internal/httpapi`、仓库 fast / full gate 均通过；PostgreSQL 17 configured gate 通过 migration `0028`、runtime role、reconnect 与 no-fallback。SQLite 真实页面覆盖重启精确回读、legacy 不反算、authority drift 与双标签 review CAS conflict；`1440×900`、`720×900`、`390×844` 无横向溢出，浏览器 console 无 warning / error。请求日志不含 payload，storage 只允许既有 exact plan / run metadata ref；provider 不重复，Tool 最多一次既有 GET，business write / replay 为 0。

## 验证矩阵

- Contract：strict schema / codec、canonical digest、compatibility / transition matrix、blocker order、legacy compatibility。
- Go：compiler、authority reload、scope、membership、permission、CAS、policy drift、confirmation、Tool pre-dispatch、side-effect counters、race、`go vet`。
- Persistence：memory / SQLite / PostgreSQL、migration、runtime role、transaction、restart、reconnect、corruption、no-fallback、legacy snapshot。
- Web：offline、strict parser、Builder / Reviewer / Run projection、generation + abort、late response、双标签 stale、privacy、production build。
- Product：canonical response / candidate → review → activation / plan recheck → controlled GET or write blocked → exact Run History snapshot。
- Repository：每批精准测试、`git diff --check`、`./scripts/check-repo.sh --fast`；schema、migration、架构、协议、阶段状态与最终关闭补全量 `./scripts/check-repo.sh`。

## 当前下一步

本卡已完成并关闭。下一步回到[当前推进焦点](../radishmind-current-focus.md)选择新的长期功能目标；不得从本卡派生批次 F、通用动作执行器、写入 policy、production 续批或平行 readiness 卡。

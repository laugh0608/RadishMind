# Prompt / Agent / Copilot 类型工作区产品化 v1

更新时间：2026-08-09

状态：`prompt_agent_copilot_type_workspace_productization_v1_implemented`

## 功能定位

本专题把已经成立的 Prompt Application 与 Agent / Copilot 开发测试态能力组织成连续类型工作区，让内部应用开发者在一个明确的 application / workspace 上下文中完成：

1. 定位类型源码与不可变版本；
2. 审查 Application Configuration 与 Publish Candidate；
3. 显式管理 runtime assignment；
4. 进入既有 API / Key 前置与受控调用；
5. 将 exact Run 交给既有 Comparison、Evaluation Case / Suite 与人工 decision owner。

本专题不创建新的领域 owner。Prompt Template、Agent Profile、Application Configuration、Publish Candidate、Runtime Assignment、Application Session、Workflow Run、Comparison、Evaluation 与 Release Decision 继续分别拥有自己的真相。

## 事实审计

### Prompt Application

| 任务 | 真实 consumer / owner | 作用域与权限 | 分页 / 详情 | 失败关闭与停止线 |
| --- | --- | --- | --- | --- |
| Template Draft / Version | `promptApplicationTemplateConsumer.ts`；Prompt Template owner | tenant + workspace + application；`prompt_application_templates:read`、`read_source`、`write`、`version` | Draft / Version 列表消费 owner 当前返回窗口；源码只走精确读取 | offline、作用域漂移、CAS、非法模板和敏感材料均失败关闭 |
| Configuration / Candidate | 既有 Application Configuration 与 Publish Candidate owner | verified identity、active workspace、资源条件 membership permission | 精确 draft / candidate 与现有列表语义不变 | candidate approval 不自动建立 assignment |
| Runtime Assignment | `promptApplicationRuntimeConsumer.ts`；Prompt Runtime Assignment owner | `prompt_application_runtime:read`；写入额外要求 `prompt_application_runtime:write` | 当前 assignment 与 append-only event；不引入第二套列表 | `activate / replace / revoke` 使用 expected version，漂移时 provider 副作用为 0 |
| 单次调用 | `promptApplicationInvocationConsumer.ts`；既有 Prompt invocation service | API Key 独立 `prompt_application:invoke`；application scope 必须精确匹配 | 单次易失输入 / 输出和 exact Run handoff | 不提交模板、authority、provider、credential 或 retry 策略 |
| Session | `promptApplicationSessionConsumer.ts`；Application Session v2 owner | `application_sessions:read / write / execute` 与对应 membership permission | active Session 当前窗口固定 `limit=100`；返回 cursor 不在本批扩全分页 UI | metadata-only 持久化，不恢复变量、回答或 transcript |
| Run / Evaluation | Workflow Run v6、Comparison v5、既有 Evaluation owner | tenant + workspace + application exact scope | Run 保持现有 cursor window；Case / Suite 使用 exact version refs | decision 只追加人工证据，不自动 candidate、assignment、release 或 deploy |

### Agent / Copilot

| 任务 | 真实 consumer / owner | 作用域与权限 | 分页 / 详情 | 失败关闭与停止线 |
| --- | --- | --- | --- | --- |
| Profile Draft / Version | `agentCopilotProfileConsumer.ts`；Agent Profile owner | tenant + workspace + application；`agent_copilot_profiles:read`、`read_source`、`write`、`version` | Draft / Version 列表消费 owner 当前返回窗口；完整 Profile 只走精确读取 | 首版强制 advisory、候选动作确认与 tool hints 关闭 |
| Configuration / Candidate | 既有 Application Configuration v4 与 Publish Candidate v4 owner | verified identity、active workspace、资源条件 membership permission | 精确 draft / candidate；不复制 Prompt / RAG 配置 | approved candidate 不自动建立 assignment |
| Runtime Assignment | `agentCopilotRuntimeConsumer.ts`；Agent Runtime Assignment owner | `agent_copilot_runtime:read`；写入额外要求 `agent_copilot_runtime:write` | 当前 assignment 与 append-only event | `activate / replace / revoke` 使用 expected version；revoked assignment 不原地恢复 |
| 受控建议 | `agentCopilotSessionConsumer.ts`；Application Session v3 与唯一 Agent service | `application_sessions:write / execute`，运行资格来自 exact assignment；API Key scope 为独立 `agent_copilot:invoke` | 当前 Session / Turn / response 由现有 owner 管理 | 只允许一次 advisory suggestion；`requires_confirmation` 不会自动执行动作 |
| Run / Evaluation | Workflow Run v7、Comparison v6、既有 Evaluation owner | tenant + workspace + application + Profile / project / task exact scope | Run 保持现有 cursor window；Case / Suite 使用 exact version refs | 不扩 agent loop、工具执行、在线搜索、retry、replay 或业务写回 |

### 共同管理边界

- application、workspace、lifecycle 或 generation 改变时，当前 owner state、未保存输入、易失回答和迟到响应必须清空。
- archived application 允许读取既有脱敏证据，但 source 写入、candidate mutation、assignment mutation 与受控调用保持阻塞；API Key 仍沿 S4 既有 read / revoke 边界。
- offline 模式显示 owner 未启用和缺失证据，不使用旧 application 或示例记录填充当前上下文。
- 普通列表保持中性；只有驱动当前 owner、详情或任务导航的对象使用选中轨。ready、partial、blocked、approved 等状态继续使用独立文字、图标和标签。
- 首分页、`limit=100` Session 窗口、当前 Run cursor window 和已加载 Evaluation 记录都不得显示为全历史统计。

## 产品化根因

现有 React 已经能逐项使用上述 owner，但类型能力仍散落在 S2 通用阶段、折叠的 controlled-test 详情、S4 Access 和 S6 Review 中：

- 类型源码、Application Configuration、Candidate 与 Assignment 在同一阶段内可能同时挂载，用户无法判断当前任务 owner；
- Prompt 单次调用与 Session、Agent Session / suggestion 的入口层级不一致；
- Run / Evaluation 虽已兼容 v6 / v7，但从类型 authority 到 exact Run 的交接缺少连续任务路径；
- archived、offline、authority drift 与 production closed 边界由各 panel 分散表达，没有类型工作区统一上下文。

因此本批新增的只是类型任务编排层：它根据 application kind 和当前 hash 选择一个 owner，复用 S2 application context、S4 access、S6 review 和既有 handoff，不复制数据或执行逻辑。

### 2026-08-09 真实路径复核

S1–S8 连续产品路径复核发现：从 Agent / Copilot application 切换到 Prompt Application，或反向切换时，application context 和可见 owner 会正确更新，但类型专属 hash 仍可能保留上一种 application kind 的 anchor。当前实现随后依赖 stage fallback 选出可见 owner，因此页面表面正确，刷新、分享和深链接身份却已经失真。

本轮只补类型切换时的 anchor 规范化：若当前 hash 是上一类型的精确 S8 owner anchor，则在不增加浏览历史的前提下替换为新类型的等价任务 anchor；共享 Configuration / Candidate anchor、S6 深层 review anchor 与无关页面 hash 保持原样。Prompt Session 切到 Agent 时落到其唯一 Session v3 Suggestion owner。该修正不改变 owner、权限、分页、API、schema 或执行行为，也不需要新的 Pencil 基准面。

### 2026-08-09 受控使用资格失败交接

真实 SQLite 路径进一步发现：Agent Session 在当前 assignment 已漂移时由服务端正确返回 `application_session_authority_changed`，且 Gateway / provider 副作用为零；但 Controlled Use owner 只显示技术失败码，用户无法判断应返回哪个既有 owner 修复资格。Prompt 单次调用和 Prompt Session 也存在相同信息断点。

本轮按真实 consumer 冻结允许列表式解释：

- Prompt 单次调用仅将 `prompt_runtime_assignment_not_found`、`prompt_runtime_candidate_ineligible`、`prompt_runtime_authority_changed` 交给 Prompt Assignment；
- Prompt Session 与 Agent Session 仅将 `application_session_authority_not_found`、`application_session_authority_changed`、`application_session_profile_ineligible` 交给各自 Assignment；
- 解释同时说明当前尝试在 provider 调用前失败关闭，不声称 assignment 已修复，也不自动 activate、replace、revoke 或重试；
- scope、输入、传输、存储、响应契约、取消和 outcome unknown 等失败保持原 owner 与原技术证据，不错误引导到 Assignment；
- 前端不预读或复制 runtime eligibility，不因本地推测禁用请求；服务端继续是当前 authority 与副作用边界的唯一判定者。

## 任务拓扑

Prompt Application 使用以下任务顺序：

1. `Template`：Prompt Template Draft / Version owner；
2. `Configuration`：Application Configuration v3；
3. `Candidate`：Publish Candidate v3 与人工 review；
4. `Assignment`：Prompt runtime assignment；
5. `Access`：复用 S4 API Integration / Key；
6. `Invocation`：单次 Prompt invocation；
7. `Session`：Prompt Application Session v2；
8. `Evaluation`：交给 S6 Runs / Compare / Cases / Release。

Agent / Copilot 使用同一语法，但类型任务为 `Profile → Configuration → Candidate → Assignment → Access → Suggestion → Evaluation`。Suggestion 继续由 Session v3 owner 承载，不另建 invocation owner。

任一时刻只挂载当前任务 owner。Evaluation 任务只显示交接边界并让 S6 挂载其当前 owner；S8 不复制 Run、Comparison、Case、Suite 或 Release Decision panel。

## Pencil 覆盖判定

五维评分为 `1 / 2 / 2 / 2 / 2 = 9`：

| 维度 | 分数 | 判断 |
| --- | ---: | --- |
| 结构新颖度 | 1 | 复用 Workbench 壳与单 owner rail，但增加类型任务拓扑 |
| 交互新颖度 | 2 | 将源码、治理、assignment、受控调用和评测交接组织为多阶段工作区 |
| 边界风险 | 2 | 存在 assignment mutation、易失输入输出、确认语义和生产误导风险 |
| 状态与窄屏复杂度 | 2 | ready / partial / offline / archived / drift / blocked 并存，窄屏必须按任务顺序重排 |
| 复用杠杆 | 2 | 同时约束 Prompt、Agent 与 Copilot 类型页面，并连接 S2、S4、S6 |

覆盖级别为 `A / 完整 Pencil`。Pencil 只维护：

- 一个 Agent / Copilot 桌面代表面，用更复杂的 advisory / confirmation 边界冻结类型任务拓扑；
- 一个窄屏代表面，冻结“类型上下文 → 全部任务 → 当前 owner → 停止线”的顺序；
- Prompt 差异直接写入设计决策，不复制完整画板。

本轮受控使用资格失败交接五维评分为 `0 / 0 / 0 / 1 / 1 = 2`：不改变 S8 结构、交互模型或风险语义，只复用既有失败表达、Assignment task 和窄屏顺序增加精确恢复入口，因此采用 `C / 直接实现`，不修改 Pencil。

## React 纵向切片

1. 新增纯 view model，按 application kind、stage 与 hash 确定任务集合和当前 owner；非法类型或未知 hash 失败关闭。
2. 新增 S8 类型工作区组件，复用 S1–S7 壳层和既有 panel，只挂载当前任务 owner。
3. 从 S2 通用 stage surface 移出 Prompt / Agent 重复挂载；Workflow / RAG 路径保持原 owner 与布局。
4. 为 Prompt assignment 与 Session 补稳定 anchor，不修改其 API、状态机或 mutation 行为。
5. Evaluation 继续交给 S6；S8 只说明 exact Run handoff、当前窗口和人工 evidence 停止线。
6. application kind 改变时，只规范化上一类型的精确 S8 anchor；普通 hashchange 和 S6 深层 owner anchor 不改写。
7. 对允许列表内的 assignment / authority 失败显示确定性原因、零 provider 副作用和对应 Assignment 锚点；其它失败不提供该交接。

## 验收方式

- 纯 view model 测试覆盖 Prompt / Agent 任务集合、hash 选择、stage fallback、跨类型 anchor 规范化、archived blocked action 与未知类型拒绝。
- Web 全量测试与 production build 通过。
- 真实浏览器覆盖 `1440×900`、关键断点和 `390×844`；检查任务选中、状态与选中分离、单 owner、响应式顺序、交互、横向溢出和控制台。
- 真实失败链验证 Controlled Use 显示精确资格说明，点击后只切换到当前类型的 Assignment task；不得发起自动 mutation 或 provider 重试。
- 运行 `./scripts/check-repo.sh --fast`；本批更新阶段真相源与产品化顺位，提交前补跑全量 `./scripts/check-repo.sh`。

## 实现结果

- Pencil 已新增 `S8 Prompt / Agent Type Workspace — Desktop / Dev Test · R1`、`S8 Prompt / Agent Type Workspace — Narrow / Dev Test · R1`，共享决策记录升级为 `S1 + S2 + S3 + S4 + S5 + S6 + S7 + S8 Visual Language — Design Decision Record · R13`；全树检查无裁切、越界或占位节点。
- React 已新增类型任务 view model 与持续 S8 工作面。Prompt 使用八任务，Agent 使用七任务；任一时刻只挂载一个既有 owner，Evaluation 只形成 S6 交接说明。S2 通用 surface 不再重复挂载 Prompt / Agent owner，Workflow / RAG 路径保持原状。
- Prompt Candidate 原组件继续持有候选数据，只新增可关闭的内嵌 assignment 组合和当前候选回传，供 S8 将 Candidate 与 Assignment 分成两个真实任务；没有复制 candidate 或 runtime assignment 状态机。
- Web `298/298` 单元测试与 production build 通过。共享 `ControlledUseFailureGuidance` chunk 为 `2.66 KiB`，S8 入口 chunk 为 `9.44 KiB`，主入口为 `462.50 KiB`，均在现有预算内。
- 应用内浏览器使用自启动 SQLite Agent / Prompt 本地产品链完成两类任务逐项切换；严格覆盖 `1440×900`、`1100×900`、`900×900`、`760×844` 与 `390×844`。各宽度 document / body / S8 scroll width 与 viewport 一致，当前任务和 owner 均唯一，响应式顺序正确，控制台零 warning / error。
- 当前产品数据没有 archived application 记录，因此浏览器没有改写真实数据制造归档态；归档 source / governance 只读和 controlled use 阻塞由纯模型测试覆盖。开发服务在验收后关闭。
- 2026-08-09 应用内浏览器真实复验 Agent Invocation → Prompt Invocation、Prompt Session → Agent Suggestion、Profile → Template、Assignment 与 Access 双向切换，并确认刷新后按真实选中 application 保留等价任务身份；共享 owner、S6 深层 review 与普通页面 hash 不被改写。既有 Pencil 基准面继续有效。
- 2026-08-09 第二次真实复验以 Agent `application_session_authority_changed` 触发资格卡，确认技术码、确定性解释、零 provider 副作用和 `#agent-copilot-runtime-assignment` 精确交接同时可见；点击后只有 Assignment task 与 owner 被选中。Prompt Invocation / Session 无对应失败时不显示伪恢复入口。`1440×900`、`1100×900`、`760×844` 与 `390×844` 均无横向溢出，控制台零 warning / error，开发服务已关闭。

## 停止线

- 不新增 API、schema、migration、repository、permission、task card、fixture 或专项 checker。
- 不创建第二套 Prompt / Agent source、configuration、candidate、assignment、session、run、comparison、evaluation 或 release owner。
- 不把 approved candidate、active assignment、successful Run 或 approved decision解释为生产发布或生产启用。
- 不实现自动 assignment、自动 release、自动 deploy、retry / fallback、replay / resume、schedule、长期记忆、agent loop、工具执行、在线搜索或业务写回。
- 不保存或恢复 Prompt 变量、渲染消息、回答、Agent context、artifact content、完整 `CopilotResponse`、credential 或 provider raw payload。
- 不启用 production membership、正式 OIDC、production API Key、quota、billing、production secret、provider 自动接入或自动路由。

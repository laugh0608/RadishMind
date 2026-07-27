# Workspace-scoped Mutation Authorization / 工作区写入与审查动作成员资格绑定（开发 / 测试态）v1

更新时间：2026-07-27

状态：`workspace_scoped_mutation_authorization_dev_test_v1_batch_a1_complete_batch_a2_ready`

## 文档目的

本专题承接 [Workspace-scoped Read Transition](workspace-scoped-read-transition-dev-test-v1.md) 已建立的 verified identity、active workspace、membership assertion、`WorkspaceMembershipProvider` 与既有 durable owner 读投影，把 User Workspace 中由人发起的写入、审查和执行动作推进到统一、可复验的工作区授权边界。

本专题不创建新的用户、tenant、role、membership、Application、Workflow、Run、Evaluation 或 API key 真相源，不把 active workspace 选择解释成授权证明，也不把开发测试态 header 或 signed-test assertion 外推为 production OIDC。首批实现前必须先完成全量 mutation inventory、拒绝顺序、权限映射、副作用顺序和零副作用证据设计。

## 当前问题

- Workspace-scoped Read Transition 已让 Applications、API keys、Workflow definitions 与 Runs 的读操作调用共享 membership provider；现有写入、审查和执行路由仍有多套 context builder。
- Application Catalog 三条 mutation 已在 body 解码前调用共享 membership provider，并只从 verified binding 建立 tenant、subject 与 workspace context；API Key Lifecycle 与 Application Interaction Session 等后续 owner 仍主要从 body / query 获取 workspace。
- Workflow Draft、Definition、RAG、Evaluation、Prompt 与 Agent 路由大多从 request context 读取 identity，并用 body workspace 与历史 dev workspace header 相等作为作用域证明；该相等关系不是 membership proof。
- Prompt、Agent 与 Application RAG 的直接 invocation 使用 application API key。API key 是应用运行凭据，不是当前人类成员关系 assertion；不能为了表面统一而在 invocation handler 中伪造 membership。
- 多数旧 handler 会把 identity、workspace 或 permission failure 压平为领域 `scope_denied`，无法稳定区分身份失效、工作区未选择、非成员、成员过期、workspace mismatch 与 membership permission denied。
- 当前 `WorkspaceMembershipProvider` 已接受 read 与 A1 Application Catalog 的单一 mutation permission；后续复杂执行路由可能要求多个业务 permission，不能通过循环调用 provider 或只校验其中一项来近似授权。

## 目标用户与核心流程

目标用户是内部开发者预览中已通过 dev header 或 signed-test token 建立 verified identity，并在 Web 中显式选择 active workspace 的工作区成员。

1. 用户发起写入、审查或执行请求。
2. 路由先确认对应开发测试态能力已显式启用，再验证 identity、identity permission、active workspace 和 membership。
3. membership decision 输出 tenant、subject、workspace、permission、source、policy version 与 expiry 均已绑定的 `ControlPlaneResourceBinding`。
4. handler 严格解析 path / body 中的资源坐标，并要求其中的 workspace 与 active workspace 精确相等；body、path、旧 dev header 都不能覆盖 active workspace。
5. 既有业务 owner 只按 verified tenant、active workspace、verified subject 与资源 ID 查询，随后执行领域校验和 CAS。
6. 只有以上步骤全部成功后，才能发生 credential 生成、Run 创建、Gateway / provider 调用、HTTP Tool 网络请求或其它不可逆副作用。
7. 返回继续使用既有脱敏 envelope，不返回 membership assertion、raw token、credential digest、DSN、SQL、provider raw body 或内部异常正文。

## 授权来源与职责

| 层级 | 权威输入 | owner | 允许输出 | 禁止替代 |
| --- | --- | --- | --- | --- |
| verified identity | dev auth headers / signed-test token / future OIDC token | auth boundary | issuer、subject、tenant、identity permission、expiry、session ref | body、query、Web state、role 名 |
| active workspace | `X-RadishMind-Active-Workspace` | 当前请求 | 单一 workspace selection | membership proof、默认 workspace |
| membership assertion | dev-test projection / signed-test claim / future reviewed adapter | assertion issuer | tenant + subject + workspace + permissions + expiry | 业务 body、旧 dev workspace header |
| membership decision | identity + selection + assertions + required permissions | `WorkspaceMembershipProvider` | verified `ControlPlaneResourceBinding` | handler 自行比对 header |
| resource ownership | binding + resource ID | 既有 durable owner | subject-owned resource / CAS state | 新聚合表、跨 owner scan |
| runtime credential | application API key | API Key Lifecycle owner | application / workspace / capability-bound调用上下文 | 人类 membership assertion |

### Identity permission 与 membership permission

- 两者是独立 allowlist，路由要求的每项 permission 必须同时存在于 verified identity 和 selected workspace membership。
- v1 不引入 `workspace:write`、`workspace:admin` 等粗粒度角色推导；membership permission 直接复用现有 operation permission 字符串。
- 多 permission 路由必须由同一次 membership decision 原子校验全部 permission；不能循环调用 provider 后拼接多个 decision。
- 实现允许扩展 `WorkspaceMembershipRequest` 以承载规范化、去重且稳定排序的 required permission 集合，但必须保持唯一 `WorkspaceMembershipProvider` owner，不创建 mutation 专用 provider。
- 现有 read permission 行为和失败码必须保持兼容；mutation permission registry 由一处维护，不继续扩 `workspaceReadPermissionAllowlist` 这类 read-only 命名和职责。

### Active workspace 与资源坐标

- `X-RadishMind-Active-Workspace` 是所有人类交互式 mutation 的唯一 workspace selector。
- body、query、path 派生资源、`X-RadishMind-Dev-Workflow-Workspace` 与其它历史 dev workspace header 只作为待核对的资源坐标；它们必须与 active workspace 精确相等。
- 缺少 active workspace 返回 `workspace_selection_missing`；重复、非法或与其它坐标不一致返回 `workspace_binding_mismatch`。
- 不从 application、draft、candidate、session、run 或 API key 反推出 active workspace，也不在请求缺失 selector 时回退资源记录中的 workspace。
- Web selector 继续只存在 React 进程内；不写 URL、cookie、local storage 或 session storage。

## 统一拒绝与副作用顺序

人类交互式 mutation 统一使用以下顺序：

1. 路由级开发测试态 enablement gate。
2. verified identity 存在、结构、issuer / signature / audience / time 与 tenant / subject 绑定。
3. identity operation permissions 全部满足。
4. active workspace 单值、语法与选择存在。
5. membership provider 对 tenant、subject、workspace、expiry 和全部 membership permissions 做一次原子 decision。
6. 有界 strict JSON、path 与 query 解析；拒绝未知字段、重复字段、尾随文档和 forbidden secret material。
7. body / path / 历史 dev header 资源坐标与 active workspace、verified subject 精确匹配。
8. 既有 owner 以 tenant + workspace + owner subject + resource ID 重读依赖资源。
9. 领域状态、不可变引用、expected version、review version、generation 与 CAS 校验。
10. durable write。
11. 只有相应业务确有需要时才生成一次性 credential、创建 Run、调用 Gateway / provider 或执行外部 HTTP 请求。
12. 输出脱敏 response 与 audit reference。

补充规则：

- 第 2 至第 7 步失败时，业务 repository query、CAS、token 随机数生成、hash、Run 写入、Gateway / provider bridge 和网络请求都必须为 0。
- membership provider 自身的未来 membership lookup 与业务 owner query 分开计数；`zero business repository query` 不得伪装成 `zero membership lookup`。
- 第 8 至第 9 步失败允许发生必要的只读 owner query，但 durable write 和外部副作用必须为 0。
- credential 必须在 application owner 和 API key CAS 准入全部通过后生成；失败 response 永远不含 token。
- 执行路由必须在 Run 占位或外部调用前固定幂等键和 authority snapshot；是否先写 Run 再调用 provider 由各既有 owner 的终态协议决定，本专题不改写其最多一次语义。

## 稳定失败语义

| failure code | HTTP | 业务 owner query | durable / external side effect |
| --- | --- | --- | --- |
| `identity_context_missing` | `401` | `0` | `0` |
| `auth_context_contract_mismatch` | `401` | `0` | `0` |
| `tenant_binding_missing` | `401 / 403` | `0` | `0` |
| `scope_denied` | `403` | `0` | `0` |
| `workspace_selection_missing` | `400` | `0` | `0` |
| `workspace_binding_mismatch` | `403` | `0` | `0` |
| `workspace_membership_denied` | `403` | `0` | `0` |
| `workspace_membership_expired` | `403` | `0` | `0` |
| `workspace_permission_denied` | `403` | `0` | `0` |
| `workspace_membership_unavailable` | `503` | `0` | `0` |

通过共享授权边界后，路由继续使用既有领域 failure，例如 payload invalid、resource not found、version conflict、transition invalid、write disabled、store unavailable、authority drift、outcome unknown。实现批次需要为旧 envelope 明确映射，不再把上述 workspace failure 全部压平为领域 `scope_denied`。

## 全量 Mutation Inventory

以下矩阵以 `services/platform/internal/httpapi/server.go` 的当前注册路由为准。`workspace 来源` 描述设计前现状；`membership permission` 是本专题冻结的目标。`只校验` 表示 POST 但不应产生持久或外部副作用。

### Application Catalog 与 API Key Lifecycle

| 路由 | identity permission | 当前 workspace 来源 | membership permission | durable owner 与副作用 |
| --- | --- | --- | --- | --- |
| `POST /v1/user-workspace/applications` | `applications:write` | body | `applications:write` | Application Catalog；create write |
| `PUT /v1/user-workspace/applications/{application_id}` | `applications:write` | body + path resource | `applications:write` | Application Catalog；expected-version CAS |
| `POST /v1/user-workspace/applications/{application_id}/archive` | `applications:archive` | body + path resource | `applications:archive` | Application Catalog；lifecycle CAS |
| `POST /v1/user-workspace/api-keys` | `api_keys:write` | body | `api_keys:write` | API Key owner + Application Catalog；record write、一次性 token 生成 / hash / handoff |
| `POST /v1/user-workspace/api-keys/{api_key_id}/revoke` | `api_keys:revoke` | body + path resource | `api_keys:revoke` | API Key owner；expected-version CAS |

### Workflow Draft、Application Configuration 与 Publish Review

| 路由 | identity permission | 当前 workspace 来源 | membership permission | durable owner 与副作用 |
| --- | --- | --- | --- | --- |
| `POST /v1/user-workspace/workflow-drafts/validate` | `workflow_drafts:write` | body + workflow dev header | `workflow_drafts:write` | Saved Draft codec / validator；只校验 |
| `POST /v1/user-workspace/workflow-drafts` | `workflow_drafts:write` | body + workflow dev header | `workflow_drafts:write` | Saved Draft owner；draft CAS |
| `POST /v1/user-workspace/application-drafts/validate` | `application_drafts:write` | body + application draft dev header | `application_drafts:write` | Application Draft validator；只校验 |
| `POST /v1/user-workspace/application-drafts` | `application_drafts:write` | body + application draft dev header | `application_drafts:write` | Application Draft owner + Application Catalog；draft CAS |
| `POST /v1/user-workspace/application-configuration-drafts/{draft_id}/prompt-template-binding` | `application_drafts:write` + `prompt_application_templates:bind` | body + path + application draft dev header | 同 identity permissions | Application Draft + Prompt Template Version owners；binding CAS |
| `POST /v1/user-workspace/application-configuration-drafts/{draft_id}/agent-copilot-profile-binding` | `application_drafts:write` + `agent_copilot_profiles:bind` | body + path + application draft dev header | 同 identity permissions | Application Draft + Agent Profile Version owners；binding CAS |
| `POST /v1/user-workspace/application-publish-candidates` | `application_publish_candidates:write`，按 candidate kind 增加 `workflow_rag_promotions:read` / `prompt_application_templates:read_source` / `agent_copilot_profiles:read_source` | application publish dev headers | 同 identity permissions | Publish Candidate + Draft + Catalog + binding source owners；immutable candidate write |
| `POST /v1/user-workspace/application-publish-candidates/{candidate_id}/reviews` | `application_publish_candidates:review` | application publish dev headers + path | `application_publish_candidates:review` | Publish Candidate owner；review CAS，不自动 assignment / release |

### Prompt Application

| 路由 | identity / runtime permission | 当前 workspace 来源 | membership permission | durable owner 与副作用 |
| --- | --- | --- | --- | --- |
| `POST /v1/user-workspace/prompt-application-templates/validate` | `prompt_application_templates:write` | body | `prompt_application_templates:write` | Template validator；只校验 |
| `POST /v1/user-workspace/prompt-application-templates` | `prompt_application_templates:write` | body | `prompt_application_templates:write` | Template + Application Catalog owners；draft CAS |
| `POST /v1/user-workspace/prompt-application-templates/{template_id}/versions` | `prompt_application_templates:version` | body + path | `prompt_application_templates:version` | Template owner；immutable version write |
| `POST /v1/user-workspace/applications/{application_id}/prompt-runtime-assignment/decisions` | `prompt_application_runtime:write` | body + path | `prompt_application_runtime:write` | Runtime Assignment + Publish Candidate / Template owners；assignment CAS |
| `POST /v1/prompt-applications/invocations` | API key `prompt_application:invoke` | API key binding + request application | 不在逐请求 membership rollout 内 | API Key + Runtime Assignment + Run owners；Gateway 单次调用、Run 终态写入 |

### Agent / Copilot

| 路由 | identity / runtime permission | 当前 workspace 来源 | membership permission | durable owner 与副作用 |
| --- | --- | --- | --- | --- |
| `POST /v1/user-workspace/agent-copilot-profiles/validate` | `agent_copilot_profiles:write` | body | `agent_copilot_profiles:write` | Profile validator / policy compiler；只校验 |
| `POST /v1/user-workspace/agent-copilot-profiles` | `agent_copilot_profiles:write` | body | `agent_copilot_profiles:write` | Profile + Application Catalog owners；draft CAS |
| `POST /v1/user-workspace/agent-copilot-profiles/{profile_id}/versions` | `agent_copilot_profiles:version` | body + path | `agent_copilot_profiles:version` | Profile owner；immutable version write |
| `POST /v1/user-workspace/applications/{application_id}/agent-copilot-runtime-assignment/decisions` | `agent_copilot_runtime:write` | body + path | `agent_copilot_runtime:write` | Runtime Assignment + Publish Candidate / Profile owners；assignment CAS |
| `POST /v1/agent-copilot/invocations` | API key `agent_copilot:invoke` | API key binding + request application | 不在逐请求 membership rollout 内 | API Key + Assignment + Run owners；Gateway 单次调用、Run 终态写入 |

### Application Session / Turn

| 路由 | identity permission | 当前 workspace 来源 | membership permission | durable owner 与副作用 |
| --- | --- | --- | --- | --- |
| `POST /v1/user-workspace/application-sessions` | `application_sessions:write` | body | `application_sessions:write` | Session + Application / authority owners；session CAS |
| `POST /v1/user-workspace/application-sessions/{session_id}/close` | `application_sessions:write` | body + path | `application_sessions:write` | Session owner；close CAS |
| `POST /v1/user-workspace/application-sessions/{session_id}/turns` | `application_sessions:execute` | body + path | `application_sessions:execute` | Session / Turn + selected runtime + Run owners；至多一次受控委托、易失正文 |

### Workflow Definition 与 Run

| 路由 | identity permission | 当前 workspace 来源 | membership permission | durable owner 与副作用 |
| --- | --- | --- | --- | --- |
| `POST /v1/user-workspace/workflow-definition-candidates` | `workflow_definitions:write` | workflow dev headers | `workflow_definitions:write` | Definition Release + Saved Draft + Catalog owners；immutable candidate write |
| `POST /v1/user-workspace/workflow-definition-candidates/{candidate_id}/decisions` | `workflow_definitions:review` | workflow dev headers + path | `workflow_definitions:review` | Definition Release owner；review CAS |
| `POST /v1/user-workspace/workflow-definitions/{definition_id}/activation-decisions` | `workflow_definitions:activate` | workflow dev headers + path | `workflow_definitions:activate` | Definition Release owner；activation CAS，不自动执行 |
| `POST /v1/user-workspace/workflow-drafts/{draft_id}/runs` | `workflow_runs:execute` | body + workflow dev headers + path | `workflow_runs:execute` | Saved Draft + Run owners；受控 Gateway 调用与 Run 终态写入 |
| `POST /v1/user-workspace/workflow-definition-runs` | `workflow_runs:execute` + `workflow_definitions:read` | body + workflow dev headers | 同 identity permissions | Definition Release + Run owners；固定 version 后受控执行 |

### Workflow RAG

| 路由 | identity / runtime permission | 当前 workspace 来源 | membership permission | durable owner 与副作用 |
| --- | --- | --- | --- | --- |
| `POST /v1/user-workspace/workflow-drafts/{draft_id}/retrieval-executions` | `workflow_rag:execute` + `workflow_runs:execute` | body + workflow dev headers + path | 同 identity permissions | Saved Draft + Snapshot + Run owners；检索、Gateway 调用、Run 写入 |
| `POST /v1/user-workspace/workflow-retrieval-snapshots` | `workflow_rag_snapshots:write` | body + workflow dev headers | `workflow_rag_snapshots:write` | Retrieval Snapshot owner；draft write |
| `POST /v1/user-workspace/workflow-retrieval-snapshots/{snapshot_id}/versions` | `workflow_rag_snapshots:write` | body + workflow dev headers + path | `workflow_rag_snapshots:write` | Snapshot owner；immutable version write |
| `POST /v1/user-workspace/workflow-retrieval-snapshots/{snapshot_id}/archive` | `workflow_rag_snapshots:archive` | body + workflow dev headers + path | `workflow_rag_snapshots:archive` | Snapshot owner；archive CAS |
| `POST /v1/user-workspace/workflow-rag-evaluation-datasets` | `workflow_rag_evaluation_datasets:write` + `workflow_rag_snapshots:read` | body + workflow dev headers | 同 identity permissions | Evaluation Dataset + Snapshot owners；dataset write |
| `POST /v1/user-workspace/workflow-rag-evaluation-datasets/{dataset_id}/versions` | `workflow_rag_evaluation_datasets:write` + `workflow_rag_snapshots:read` | body + workflow dev headers + path | 同 identity permissions | Dataset + Snapshot owners；immutable version write |
| `POST /v1/user-workspace/workflow-rag-evaluation-datasets/{dataset_id}/archive` | `workflow_rag_evaluation_datasets:archive` | body + workflow dev headers + path | `workflow_rag_evaluation_datasets:archive` | Dataset owner；archive CAS |
| `POST /v1/user-workspace/workflow-rag-evaluation-datasets/{dataset_id}/candidate-reviews` | `workflow_rag_evaluation_datasets:review` + `workflow_rag_evaluation_datasets:read` + `workflow_rag_snapshots:read` | body + workflow dev headers + path | 同 identity permissions | Dataset + Snapshot + Candidate Review owners；review record write |
| `POST /v1/user-workspace/workflow-rag-knowledge-promotion-candidates` | `workflow_rag_promotions:write` + `workflow_rag_evaluation_datasets:read` + `workflow_rag_snapshots:read` + `application_drafts:read` | body + workflow dev headers | 同 identity permissions | Promotion + Dataset + Snapshot + Application Draft owners；candidate write |
| `POST /v1/user-workspace/workflow-rag-knowledge-promotion-candidates/{candidate_id}/decisions` | `workflow_rag_promotions:review` | body + workflow dev headers + path | `workflow_rag_promotions:review` | Promotion owner；review CAS |
| `POST /v1/user-workspace/applications/{application_id}/workflow-rag-runtime-assignment/decisions` | `workflow_rag_runtime:write` | body + workflow dev headers + path | `workflow_rag_runtime:write` | Runtime Assignment + Promotion + Application owners；assignment CAS |
| `POST /v1/application-rag/invocations` | API key `application_rag:invoke` | API key binding + request application | 不在逐请求 membership rollout 内 | API Key + Assignment + Run owners；检索、Gateway 单次调用、Run 写入 |

### Workflow HTTP Tool

| 路由 | identity permission | 当前 workspace 来源 | membership permission | durable owner 与副作用 |
| --- | --- | --- | --- | --- |
| `POST /v1/user-workspace/workflow-drafts/{draft_id}/tool-action-plans` | `workflow_drafts:read` + `workflow_tool_actions:plan` | body + workflow dev headers + path | 同 identity permissions | Saved Draft + Tool Action owner；计划写入，零网络 |
| `POST /v1/user-workspace/workflow-tool-action-plans/{plan_id}/decisions` | `workflow_tool_actions:confirm` | body + workflow dev headers + path | `workflow_tool_actions:confirm` | Tool Action owner；人工 decision CAS，零网络 |
| `POST /v1/user-workspace/workflow-tool-action-plans/{plan_id}/executions` | `workflow_tool_actions:execute` + `workflow_runs:execute` + `workflow_drafts:read` | body + workflow dev headers + path | 同 identity permissions | Tool Action + Execution + Run owners；确认重读、受控 HTTP 网络、Run 终态写入 |

### Workflow Evaluation

| 路由 | identity permission | 当前 workspace 来源 | membership permission | durable owner 与副作用 |
| --- | --- | --- | --- | --- |
| `POST /v1/user-workspace/workflow-evaluation-cases` | `workflow_evaluations:write` + `workflow_runs:read` | body + workflow dev headers | 同 identity permissions | Evaluation Case + Run owners；case write |
| `POST /v1/user-workspace/workflow-evaluation-cases/{case_id}/revisions` | `workflow_evaluations:write` + `workflow_runs:read` | body + workflow dev headers + path | 同 identity permissions | Evaluation Case + Run owners；immutable revision write |
| `POST /v1/user-workspace/workflow-evaluation-suites` | `workflow_evaluations:write` + `workflow_runs:read` | body + workflow dev headers | 同 identity permissions | Evaluation Suite + Case / Run owners；suite write |
| `POST /v1/user-workspace/workflow-evaluation-suites/{suite_id}/decisions` | `workflow_evaluations:write` + `workflow_runs:read` | body + workflow dev headers + path | 同 identity permissions | Evaluation Suite owner；人工 decision CAS，不自动发布 |

## 不进入逐请求 Membership Rollout 的路由

以下路径仍需保留在风险审计中，但不由本专题给每次请求追加人类 membership assertion：

- `/v1/chat/completions`、`/v1/responses`、`/v1/messages` 与三类 application invocation 使用 application API key；其 workspace、application 与 capability 必须只来自验证后的 key binding。成员离职、key 吊销和应用 credential 生命周期的生产策略需要独立 policy owner。
- `POST /v1/tools/actions` 当前稳定返回 blocked action response，没有业务 mutation；不借本专题启用 unrestricted tool。
- Admin Provider Route mutation 属于 Admin Control Plane 的 environment-scoped 授权，不是 User Workspace active workspace；继续复用其已关闭专题，不并入本批。
- repository migration、reconciler 与进程恢复属于内部系统动作，必须使用固定 system actor 和既有 store contract，不伪造交互式 membership。

## 第一实现批次

### 批次 A1：Application Catalog

范围：

- `POST /v1/user-workspace/applications`
- `PUT /v1/user-workspace/applications/{application_id}`
- `POST /v1/user-workspace/applications/{application_id}/archive`

要求：

- 建立共享 mutation authorization 入口，复用唯一 `WorkspaceMembershipProvider`。
- active workspace 成为唯一 selector，body workspace 只做精确一致性校验。
- create / update / archive 分别要求 `applications:write` / `applications:write` / `applications:archive` 的 identity 与 membership 双重授权。
- 跨 tenant / subject、非成员、过期 identity / membership、workspace mismatch、permission denied 在 Application Catalog repository 查询前失败关闭。
- 保持 memory / SQLite / PostgreSQL owner、CAS、soft archive 和下游归档只读语义不变。

实施结果（2026-07-27）：

- 共享 workspace authorization 入口已从 read-only 命名边界抽为通用入口；既有 read wrapper 与行为保持不变，没有创建第二个 provider。
- `applications:write` 与 `applications:archive` 已进入 membership permission registry。create / update / archive 在 JSON body 解码前完成 identity、identity permission、active workspace 与 membership decision，body workspace 只做 verified binding 精确一致性校验。
- 稳定失败码不再压平为 Application Catalog 领域 `scope_denied`；identity、selection、membership 与 body binding 负向矩阵均由 repository spy 证明业务查询和写入为 0。
- dev headers、signed-test token、过期 signed identity 与 OIDC unavailable 已覆盖；Web mutation 发送 active workspace，只有 dev mode 发送 membership proof，signed-test token 只携带内存 token 与 active selection。
- Application Catalog memory、SQLite、本地产品重启链、完整 Platform HTTP、定向 race、`go vet`、Web 245 项测试 / production build 和 PostgreSQL integration suite 均通过；record schema、CAS、soft archive、cursor 与 read route 保持不变。

### 批次 A2：API Key Lifecycle

范围：

- `POST /v1/user-workspace/api-keys`
- `POST /v1/user-workspace/api-keys/{api_key_id}/revoke`
- 三类 application invocation 与 Gateway API key 认证只做负向回归，不改变逐请求授权模型。

要求：

- create / revoke 分别要求 `api_keys:write` / `api_keys:revoke` 的 identity 与 membership 双重授权。
- API Key owner 与 Application Catalog owner 只在 membership 成功后查询。
- token 随机数生成、hash、credential record write 和一次性交接均在完整授权、active application 与资源绑定校验之后。
- failure、日志、request history、Web state 与持久化介质不得出现原始 token；成功 response 继续 `Cache-Control: no-store`。
- revoke 继续以 API key record version 做 CAS，不修改历史 Gateway request / Run。

批次 A1 和 A2 使用同一任务卡并按顺序实现和验证。A1 已形成稳定 shared context、错误映射与双数据库证据；当前允许进入 A2，但不得借 A2 扩大到后续 owner。

## 后续批次顺序

1. 批次 B：Workflow Draft、Application Configuration Draft、Prompt Template 与 Agent Profile 等创作 owner。
2. 批次 C：Publish Candidate、Definition Candidate、RAG Promotion 与三类 Runtime Assignment 等审查 / 激活 owner。
3. 批次 D：Workflow Run、Application Session / Turn 与人类发起的受控执行。
4. 批次 E：RAG Dataset / Snapshot、HTTP Tool 与 Evaluation 组合 owner。

每批开始前必须回看 inventory 中的真实 owner 依赖。若需要修改 schema、公开 API、运行时幂等、外部 provider 或高风险网络边界，先更新本专题和唯一任务卡，不把同一专题拆成平行 readiness 链。

## 验收方式

设计阶段：

- inventory 与 `server.go` 注册路由逐项对应。
- 文档入口、当前焦点、用户工作区专题和周志同步。
- `./scripts/check-repo.sh --fast` 与全量 `./scripts/check-repo.sh` 通过。

批次 A：

- `WorkspaceMembershipProvider` 单元测试覆盖 mutation allowlist、多 permission 原子 decision、expiry、tenant / subject / workspace mismatch。
- Application Catalog 与 API Key handler 测试覆盖 dev headers、signed-test token、OIDC unavailable、缺失 / 重复 active workspace、identity / membership permission 差异和稳定 HTTP / failure code。
- repository spy、credential generator spy、Gateway bridge spy 分别证明拒绝路径零业务查询、零 token / hash、零 invocation。
- memory、SQLite 与 PostgreSQL 连续链验证 workspace / owner 隔离、CAS、重启恢复与 no fallback。
- Web 授权消费测试确认 active workspace 与 membership proof 分离，切换后不复用旧 response 或 credential。
- 相邻 Go tests、定向 race、`go vet ./...`、Web tests / build、fast 与全量仓库检查按任务卡执行。

## 隐私边界

- raw identity token、membership assertion、dev membership headers 与 API key 不进入日志、audit record、error body、Run input 或 committed fixture。
- response 只允许返回 sanitized workspace ref、业务 record、稳定 failure code、必要 current version 和 audit ref。
- membership source / policy version 只用于服务端授权与测试观测，默认不返回 Web。
- 一次性 API key 只在成功创建 response 中出现一次，不进入浏览器持久化介质。
- Session / Turn、Run、RAG 和 HTTP Tool 继续遵守既有正文最小化、provider raw body 禁止和业务真相源不写回边界。

## 生产态准入

开发测试态完成只证明 deterministic dev / signed-test membership 与既有 durable owner 的授权顺序可复验。production enablement 必须另行满足：

- reviewed Radish membership owner / endpoint、撤销 / 过期语义和稳定 subject / tenant / workspace mapping；
- 真实 OIDC issuer、audience、key rotation、session / auth time 与 negative integration test；
- production repository、secret backend、部署环境、审计 owner、负责人和发布复核；
- API key 与成员撤销、应用归档、quota / billing、credential rotation 的正式 policy owner；
- production capability enablement 与运行配置失败关闭。

以上条件未满足时，OIDC mutation 统一返回 `workspace_membership_unavailable`，不得回退 dev headers、signed-test assertion、旧 resource binding 或默认 workspace。

## 停止线

- 任务卡已冻结且 A1 已完成；当前只进入 A2 API Key Lifecycle，不进入批次 B 或后续 mutation owner。
- 不把 active workspace、body workspace、旧 dev header、application owner 或 API key 解释成 membership proof。
- 不创建新的用户、tenant、role、membership、Application、Workflow、Run、Evaluation 或 credential owner。
- 不通过统一授权专题顺带启用 production OIDC、production API key、quota / billing、自动发布、自动确认、replay、unrestricted tool 或业务写回。
- 不改变三类 application API key invocation 的逐请求授权模型，除非后续独立专题明确 credential 与 membership 撤销策略。
- 不为每个 handler 新建 wrapper、checker、fixture 或任务卡；共享授权由一处实现，行为证据优先进入相邻测试和聚合门禁。

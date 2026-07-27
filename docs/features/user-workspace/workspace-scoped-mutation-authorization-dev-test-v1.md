# Workspace-scoped Mutation Authorization / 工作区写入与审查动作成员资格绑定（开发 / 测试态）v1

更新时间：2026-07-27

状态：`workspace_scoped_mutation_authorization_dev_test_v1_batch_d_complete_batch_e_ready`

## 文档目的

本专题承接 [Workspace-scoped Read Transition](workspace-scoped-read-transition-dev-test-v1.md) 已建立的 verified identity、active workspace、membership assertion、`WorkspaceMembershipProvider` 与既有 durable owner 读投影，把 User Workspace 中由人发起的写入、审查和执行动作推进到统一、可复验的工作区授权边界。

本专题不创建新的用户、tenant、role、membership、Application、Workflow、Run、Evaluation 或 API key 真相源，不把 active workspace 选择解释成授权证明，也不把开发测试态 header 或 signed-test assertion 外推为 production OIDC。首批实现前必须先完成全量 mutation inventory、拒绝顺序、权限映射、副作用顺序和零副作用证据设计。

## 当前问题

- Workspace-scoped Read Transition 已让 Applications、API keys、Workflow definitions 与 Runs 的读操作调用共享 membership provider；现有写入、审查和执行路由仍有多套 context builder。
- 批次 A 至 D 共 32 条 mutation 已在 body 解码前调用共享 membership provider，并只从 verified binding 建立 tenant、subject 与 workspace context；Evaluation、RAG Dataset / Snapshot 与 HTTP Tool 等后续组合 owner 仍需按 inventory 复核。
- Workflow Draft、Definition、RAG、Prompt、Agent、Session、Turn 与两类人类受控 Run 入口已完成迁移；尚未迁移的 Evaluation、RAG Dataset / Snapshot 与 HTTP Tool 路由仍需按 inventory 复核真实 owner 和副作用顺序。
- Prompt、Agent 与 Application RAG 的直接 invocation 使用 application API key。API key 是应用运行凭据，不是当前人类成员关系 assertion；不能为了表面统一而在 invocation handler 中伪造 membership。
- 尚未迁移的旧 handler 仍可能把 identity、workspace 或 permission failure 压平为领域 `scope_denied`，无法稳定区分身份失效、工作区未选择、非成员、成员过期、workspace mismatch 与 membership permission denied。
- 当前 `WorkspaceMembershipProvider` 已接受 read、批次 A 至 C 的单项和原子组合 mutation permission；后续复杂执行路由同样不能通过循环调用 provider 或只校验其中一项来近似授权。

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
| `POST /v1/user-workspace/application-publish-candidates/{candidate_id}/reviews` | `application_publish_candidates:review`，`approve` 按 candidate kind 增加 `workflow_rag_promotions:read` / `prompt_application_templates:read_source` / `agent_copilot_profiles:read_source` | application publish dev headers + path | 同 identity permissions | Publish Candidate owner；review CAS，不自动 assignment / release |

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

实施结果（2026-07-27）：

- `api_keys:write` 与 `api_keys:revoke` 已进入共享 membership permission registry；create / revoke 在 JSON body 解码前完成 identity、identity permission、active workspace 与 membership decision，body workspace 只做 verified binding 精确一致性校验。
- API Key owner 与 Application Catalog owner 只在 membership 成功后进入。签发顺序由相邻 spy 固定为 Application owner reload → identifier generation → credential generation / hash → record write；成功 response 才执行一次性交接并继续要求 `Cache-Control: no-store`。
- create / revoke 双操作负向矩阵覆盖 identity、selection、membership、body binding 与 OIDC unavailable；Application Catalog 和 API Key repository 对全部授权拒绝路径均记录为 0。既有 inactive / cross-owner Application 测试继续证明 credential generator 为 0。
- dev headers、signed-test token 与 Web header 分离已覆盖；read route、三类 application invocation、Gateway API key auth、record schema、repository interface、migration、CAS 与历史 request / Run 均保持不变。
- memory、SQLite、PostgreSQL、吊销、重启恢复、敏感信息扫描、完整 Platform HTTP、定向 race、`go vet`、Web 246 项测试和 production build 均通过。

批次 A1 和 A2 使用同一任务卡并已按顺序完成和验证。

### 批次 B：创作 owner

范围：

- Workflow Draft validate / save；
- Application Configuration Draft validate / save，以及 Prompt Template / Agent Profile 两条显式 binding；
- Prompt Template validate / save / immutable version；
- Agent Profile validate / save / immutable version。

实施结果（2026-07-27）：

- 共享 workspace authorization 已支持权限集合。identity 权限在 provider 前一次性校验，membership provider 每个请求只调用一次并原子验证全部 required permissions；两条 binding 路由分别要求 `application_drafts:write` + `prompt_application_templates:bind`、`application_drafts:write` + `agent_copilot_profiles:bind`。
- 12 条 mutation 均在 JSON body 解码前完成 identity、identity permission、active workspace 与 membership decision；body workspace 只与 verified binding 精确比较，tenant、workspace、actor 与 owner subject 不再来自旧 owner dev header。
- Configuration Draft save 中已有的可选 RAG / Prompt / Agent binding 能力只按 verified identity grants 与 verified membership grants 的交集启用；不循环调用 provider，也不因基础草案保存隐式授予 binding permission。
- 跨四类 owner 的拒绝矩阵覆盖 identity permission、selection、membership 缺失 / 过期 / permission、workspace mismatch 与 OIDC unavailable；owner spy 全部为 0。组合权限测试证明 identity 第二权限缺失时 provider 为 0，membership 第二权限缺失时 provider 恰好调用一次且 owner 为 0。
- signed-test permission projection 已补齐 Workflow Draft 与 Application Draft；四类 validate 路由均通过真实签名 token + signed membership assertion。Web mutation 发送 active workspace 和精确 dev membership permission，读请求仍沿既有 header / owner 边界。
- memory、SQLite、PostgreSQL owner contract、完整 Platform Go tests、Web 246 项测试 / production build 与 PostgreSQL integration suite 已通过；repository interface、schema、migration、CAS、application API key invocation 和批次 C 审查 / 激活 owner 均未改变。

批次 B 已关闭。

### 批次 C：审查与激活 owner

范围：

- Application Publish Candidate create / review；
- Workflow Definition Candidate create / review 与 Definition activation；
- Workflow RAG Promotion Candidate create / review；
- Workflow RAG、Prompt Application、Agent Copilot 三类 Runtime Assignment decision。

设计复核结论（2026-07-27）：

- 本批共 10 条 mutation。除 Application Publish Candidate create 和 `approve` review 的资源条件 source permission 外，其余入口的完整 permission 集合都能在业务 owner 前确定，必须在 JSON body 解码前通过同一次 membership decision。
- RAG Promotion create 原子要求 `workflow_rag_promotions:write`、`workflow_rag_evaluation_datasets:read`、`workflow_rag_snapshots:read` 与 `application_drafts:read`；缺少任一 identity permission 时 provider 为 0，缺少任一 membership permission 时 provider 只调用一次，全部业务 owner 为 0。
- Application Publish Candidate create 先原子要求 `application_publish_candidates:write`，review 先要求 `application_publish_candidates:review`。附加 `workflow_rag_promotions:read`、`prompt_application_templates:read_source` 或 `agent_copilot_profiles:read_source` 只能由权威草案或候选类型决定，因此允许在基础 authorization 后由 create 最小重读 Application Draft、由 `approve` 最小重读 Publish Candidate，再从同一 verified identity 与 membership grant 交集派生对应能力；不得第二次调用 provider，也不得要求调用者持有与实际候选类型无关的 source permission。非 `approve` review 不追加 source permission。
- 上述条件权限缺失时，只允许发生确定候选类型所需的 Application Draft 或 Publish Candidate 最小只读；Publish Candidate durable write、Application Catalog 与 RAG / Prompt / Agent source owner 重读均必须为 0。其它 identity、selection、membership 与 body workspace 拒绝仍要求全部业务 owner 为 0。
- review、activation 与三类 assignment decision 只要求各自单项 permission；既有 service 继续负责按 verified tenant、workspace、subject 与资源 ID 重读 candidate、draft、version、application、promotion 等权威 owner，并维持 CAS、不可变版本、append-only event / audit 和“不因 approve 自动 activate / execute”的状态机。
- body workspace 只与 active workspace 精确比较；Application Publish、Workflow Definition 的历史 workspace / application header 以及三类 Runtime Assignment 的旧 dev header 不再建立 mutation authority。path application / definition / candidate ID 继续由既有 owner 做作用域内精确重读。

实施结果（2026-07-27）：

- 10 条 Publish Candidate、Definition Candidate / activation、RAG Promotion 与三类 Runtime Assignment mutation 已接入共享 authorization；所有可预先确定的 permission 都在 body 前由一次 membership decision 原子验证。
- 跨 10 条入口的拒绝矩阵覆盖 identity permission、active workspace、membership 缺失 / permission、workspace mismatch 与 OIDC unavailable，primary owner 调用均为 0；RAG Promotion 四权限回归固定 identity 缺权限时 provider 为 0、membership 缺权限时 provider 恰好为 1。
- Application Publish create 与 `approve` review 只在基础授权后最小重读权威 Draft / Candidate，再从同一 verified identity 与 membership grants 交集派生实际 source permission；缺权限时 baseline、source owner 与 durable write 均为 0。
- 上游 signed permission projection、Web 六类 consumer 的 active workspace / dev membership headers、完整 Platform Go tests、定向 race、`go vet`、Web 246 项测试与 PostgreSQL integration suite 均已通过；既有 schema、migration、CAS、append-only event / audit 和 application API key invocation 未改变。
- 批次 C 已关闭；专题累计 27 条 mutation，下一步只进入批次 D 的 Workflow Run、Application Session / Turn 与人类受控执行设计复核。

### 批次 D：Session / Turn 与人类受控执行

范围：

- Application Interaction Session create / close；
- Application Interaction Turn execute；
- Saved Workflow Draft run；
- Workflow Definition-bound run。

设计复核结论（2026-07-27）：

- 本批共 5 条 mutation。Session create / close 原子要求 `application_sessions:write`，Turn execute 原子要求 `application_sessions:execute`，Saved Draft run 原子要求 `workflow_runs:execute` + `workflow_drafts:read`，Definition-bound run 原子要求 `workflow_runs:execute` + `workflow_definitions:read`；完整权限集合均可在 body 解码与业务 owner 前确定，必须由一次 membership decision 验证。
- Session create / close 在完整授权与 body workspace / application binding 通过后，才能重读 Application、Definition 或 Runtime Assignment authority 并写 Session CAS；不得调用 Gateway / provider 或创建 Run。
- Turn execute 在完整授权与 body binding 通过后，才允许重读 Session、校验 active state / expected version / client turn key、预留 metadata-only Turn，并按 Session 已持久化的 exact authority 至多委托一次 Workflow Definition、Application RAG、Prompt 或 Agent runtime。客户端不能提交 API key、runtime credential 或 authority override；直接 application invocation 的 API key 边界保持不变。
- Saved Draft run 与 Definition-bound run 在完整授权与 body binding 通过后，才能分别重读 exact saved draft，或 activation / immutable version / Application authority；随后沿用既有 planned Run 写入 → 至多一次 Gateway / provider 调用 → terminal Run 写入顺序。
- identity、selection、membership 与 body workspace / application binding 拒绝时，Session、Draft、Definition、Application、Runtime Assignment、Run owner、Turn reservation 和 Gateway / provider 调用均为 0。严格 payload、资源缺失、CAS 或 authority drift 的后续失败继续由既有 service 负责，不改变 schema、repository interface、幂等语义或失败恢复。

实施结果（2026-07-27）：

- 5 条 Session / Turn / Run mutation 已接入共享 authorization；单项权限和两类双权限 Run 均在 body 前由一次 membership decision 原子验证，identity 缺第二权限时 provider 为 0，membership 缺第二权限时 provider 恰好为 1。
- 跨 5 条入口的拒绝矩阵覆盖 identity permission、active workspace、membership 缺失 / permission、workspace binding 与 OIDC unavailable；Session、Draft、Definition、Application、Runtime Assignment、Run owner、Turn reservation 及 Gateway / provider 调用均为 0。
- Session create / close 保持零 Run 与零外部调用；Turn 继续只按 persisted Session authority 至多委托一次；两类 Run 保持 planned write → 单次外部调用 → terminal write，不接受客户端 credential 或 authority override。
- 上游 permission projection 与 Web 的 Application Interaction、Prompt Session、Agent Session、Saved Draft Run、Definition Run consumer 已补齐 active workspace 和精确 dev membership permission；read route 与 application API key invocation 未改变。
- 完整 Platform Go tests、定向 race、`go vet`、Web 246 项测试 / production build 与 PostgreSQL integration suite 均已通过；repository interface、schema、migration、CAS、idempotent replay 与 metadata-only persistence 未改变。
- 批次 D 已关闭；专题累计 32 条 mutation，下一步只进入批次 E 的 RAG Dataset / Snapshot、HTTP Tool 与 Evaluation 组合 owner 设计复核。

## 后续批次顺序

1. 批次 C 已完成：Publish Candidate、Definition Candidate、RAG Promotion 与三类 Runtime Assignment 等审查 / 激活 owner。
2. 批次 D 已完成：Workflow Run、Application Session / Turn 与人类发起的受控执行。
3. 下一批次 E：RAG Dataset / Snapshot、HTTP Tool 与 Evaluation 组合 owner。

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

批次 B：

- 四类创作 owner 的 validate / save / version / binding 正向链保持原有 CAS、不可变版本和 ref-only binding 语义。
- identity、selection、membership 与 body binding 负向矩阵在业务 owner 前失败，组合权限由一次 provider decision 原子判断。
- dev headers、signed-test 与 OIDC unavailable 失败关闭均可复验；Web mutation 只在 dev mode 携带 membership proof。
- memory、SQLite、PostgreSQL、定向 race、`go vet`、Web tests / build、fast 与全量仓库检查通过。

批次 C：

- 10 条审查 / 激活 mutation 保持既有 candidate immutable、review / activation / assignment CAS、append-only event / audit 与 ref-only runtime selection。
- 可预先确定的 permission 在业务 owner 前由一次 membership decision 原子验证；Publish Candidate create / approve 的资源条件权限只允许一次最小草案 / 候选重读，且不得触发无关 source owner 或 durable write。
- identity、selection、membership、body binding 与 OIDC unavailable 失败关闭可复验；三类 assignment approve / activate 不自动 invocation 或创建 Run。
- memory、SQLite、PostgreSQL、定向 race、`go vet`、Web tests / build、fast 与全量仓库检查通过。

批次 D：

- 5 条 Session / Turn / Run mutation 保持既有 Session CAS、Turn reservation、Run planned / terminal 写入、幂等重放与 metadata-only persistence。
- 单项和双项 permission 均在业务 owner 与外部调用前由一次 membership decision 原子验证；授权和 binding 拒绝时业务 owner、Turn reservation、Run write 与 Gateway / provider 均为 0。
- Web mutation 在 dev mode 携带 active workspace 与精确 membership permission；read route 与 application API key invocation 不继承人类 membership proof。
- memory、SQLite、PostgreSQL、定向 race、`go vet`、Web tests / build、fast 与全量仓库检查通过。

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

- 任务卡批次 A 至 D 已完成；下一步只复核批次 E 的 RAG Dataset / Snapshot、HTTP Tool 与 Evaluation 组合 owner，不在设计冻结前直接迁移。
- 不把 active workspace、body workspace、旧 dev header、application owner 或 API key 解释成 membership proof。
- 不创建新的用户、tenant、role、membership、Application、Workflow、Run、Evaluation 或 credential owner。
- 不通过统一授权专题顺带启用 production OIDC、production API key、quota / billing、自动发布、自动确认、replay、unrestricted tool 或业务写回。
- 不改变三类 application API key invocation 的逐请求授权模型，除非后续独立专题明确 credential 与 membership 撤销策略。
- 不为每个 handler 新建 wrapper、checker、fixture 或任务卡；共享授权由一处实现，行为证据优先进入相邻测试和聚合门禁。

# RadishMind 系统架构

更新时间：2026-09-14

## 架构职责

本文描述当前进程、模块、请求流与数据归属，并区分实现事实、逻辑职责和后续收敛方向。能力成熟度见[能力矩阵](radishmind-capability-matrix.md)，执行顺位见[当前推进焦点](radishmind-current-focus.md)，功能细节见[功能设计入口](features/README.md)。

RadishMind 是 AI 工具、工作流、模型网关和 Copilot 集成平台，不是上层业务真相源。RadishMind 本地账户、角色与工作区成员关系是平台内部真相；Radish 自身身份、组织成员关系和业务数据仍由 Radish 负责。

## 当前进程与代码映射

| 组件 | 当前落点 | 职责与边界 |
| --- | --- | --- |
| 产品 Web | `apps/radishmind-web/` | React + Vite + TypeScript；应用、Workflow、Gateway 和管理端用户流程 |
| 本地运维 Console | `apps/radishmind-console/` | 平台发现、诊断与只读证据；不代替正式产品 Web |
| Go 平台服务 | `services/platform/` | HTTP、身份与权限、领域编排、持久化、配额、路由、运行与审计 |
| 工作区策略 | `services/platform/internal/workspacepolicy/` | 标准库依赖的权限目录、四个内建角色、固定 digest 与独立策略测试；由 HTTP、身份服务和既有仓储直接消费 |
| 桥接层 | `services/platform/internal/bridge/` | 默认四个受控 Python 标准输入输出工作进程；健康握手、有界排队、取消、超时、崩溃重建、优雅关闭和请求隔离 |
| Python 网关与运行时 | `services/gateway/`、`services/runtime/` | 规范请求分发、提供方适配、模型推理、规则与响应构建；`process_per_request` 保留显式回滚 |
| 契约 | `contracts/`、`contracts/typescript/` | 规范 schema、产品资源协议和跨语言消费合同 |
| 评测与实验 | `datasets/`、`scripts/eval/`、`training/` | 固定任务、候选记录、离线回归与模型路线证据 |

当前大量领域、存储与执行逻辑仍在 `services/platform/internal/httpapi/` 同包，产品工作区状态也部分集中在 `App.tsx` 和 `features/control-plane-read/`。下文的逻辑职责不表示已经完成 Go 包拆分、前端目录迁移或独立服务部署。

Prompt 应用调用在 Python 运行时中由 `prompt_application_inference.py` 解包 Go 已渲染的有序消息，并将答案原文封装回既有规范响应；Go 继续负责精确运行权限、输出契约和结果保存。该适配不是新进程或公开协议，具体边界见[Prompt 提供方适配](features/user-workspace/prompt-application-template-version-review-controlled-invocation-dev-test-v1.md#真实-provider-的消息与输出适配)。

## 四个产品面与逻辑职责

- `User Workspace`：应用、草案、运行、结果和评测的用户入口；页面消费领域状态，不复制服务端授权或发布资格。
- `Admin Control Plane`：账户、租户、工作区成员、角色、API 密钥、提供方档案、路由、配额、价格策略和运行审计；配置与执行保持明确交接。
- `Model Gateway / API Distribution`：北向协议翻译、模型发现、请求准入、提供方调用和脱敏请求历史；不用兼容协议建立第二套业务真相。
- `Workflow / Agent Runtime`：按明确执行档案编排 Definition、RAG、Prompt、Agent / Copilot 与受控 HTTP Tool，记录 run record、状态流转和失败边界。

核心逻辑层继续是 Client Adapters & Context Packers（客户端适配与上下文打包）、Copilot Gateway / Task Router（Copilot 网关与任务路由）、Retrieval & Tool Layer（检索与工具）、Model Runtime Layer（模型运行时）、Rule Validation & Response Builder（规则校验与响应构建）、Data / Evaluation / Training Pipeline（数据、评测与训练流水线）。它们是职责划分，不要求每次请求依次经过所有层；检索与工具仅在对应执行档案允许时发生。

## 请求与执行边界

```text
Web / SDK / 上层项目
  → HTTP 请求解析与协议适配
  → 验证身份、工作区成员资格、资源作用域和操作权限
  → 对应领域服务重读精确版本、资格与执行配置
  → Gateway 请求准入 / quota / 固定路由与价格快照
  → Go bridge → Python worker → Provider
  → 规则与响应构建 → 运行 / 请求 / 审计记录
  → 严格消费者与对应用户工作区
```

- 北向已有 `/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models` 兼容接口；内部 Copilot 链使用 `CopilotRequest / CopilotResponse / CopilotGatewayEnvelope`。产品管理资源有各自规范契约，不强行归一为模型请求。
- Go 负责请求准入与领域执行边界，Python 负责模型和 AI 生态适配。一次运行可能产生多个明确的提供方调用尝试；作用域、权限与配额不能由模型输出决定。
- 默认重试与回退继续关闭或由调用方管理；已实现的开发测试态回退仅在独立门禁、允许的 API 密钥非流式请求、冻结路由和符合回退条件的类型化失败同时成立时执行。流式请求、非 API 密钥请求与生产态不自动继承此能力。
- 所有高风险交接重读精确资源与当前权限，跨负责模块仅传稳定引用；客户端快照、旧资格或准入状态投影不能代替执行授权。
- API key、token、DSN 和 Provider 原始错误不进入 argv、公开日志或响应；错误保持稳定分类与脱敏上下文。取消、超时、关闭与持久化失败不伪造成功或回退样例。

## 数据归属与持久化

| 数据 | 当前归属 | 一致性要求 |
| --- | --- | --- |
| 本地账户、凭证、外部身份绑定、Web 会话 | 本地身份存储库 | 显式 `(issuer, subject)` 绑定，不按邮箱合并；认证失败不回退其它身份来源 |
| 工作区成员、角色与邀请 | 既有身份 / 授权负责模块 | 邀请认领在单事务中建立成员与角色并消费邀请；不另设权限真相 |
| 应用配置、模板、发布候选、运行时绑定 | 对应领域仓储 | 不可变版本、CAS 审查、精确作用域和当前资格重验 |
| 草案 / 定义、运行、会话 / 回合、评测 | 对应工作流 / 应用领域负责模块 | 执行来源、版本、审计与历史兼容明确；跨负责模块交接不复制完整状态 |
| 结果资产与生命周期 | 显式结果资产负责模块 | 原始结果按已批准保存策略处理；会话不自动成为对话全文或通用结果库 |
| 网关请求、配额、价格快照 | 各自独立负责模块 | 提供方调用前原子准入、不可变价格、可信用量报告；不构成成本总账（cost ledger）或账单 |

`memory_dev` 用于不要求跨进程恢复的测试。聚合 `sqlite_dev` 通过共享文件与运行时管理组件迁移和仓储；显式 `postgres_dev_test` 为组件保留连接、迁移、标记 / 校验和与运行角色边界。

Prompt / Agent Runtime Assignment、Session / Turn、Run 投影、Action Safety snapshot、Evaluation Plan / Campaign / Schedule / Occurrence 复用 Workflow Run Store 已审查的持久化边界。复用连接不改变各资源的领域职责。

未知选择器、禁用模式、迁移不兼容、查询或存储失败均失败关闭；不能回退内存存储。双数据库的事务、并发、权限、重启与查询行为分别验证，不能用其中一个替代另一个。详细配置见[SQLite 专题](platform/local-sqlite-dev-persistence-v1.md)与[运行手册](platform/platform-service-operations-runbook-v1.md)。

## Session、工具与后台调度

应用 Session / Turn、显式结果资产和运行链已有开发测试态持久化。历史通用 `/v1/session/metadata` 与 checkpoint metadata-only 路由仍不提供长期记忆、任意跨轮恢复或 replay 执行。

通用 `/v1/tools/actions` 保持 blocked shell。Workflow HTTP Tool 通过独立 action plan、confirmation、执行档案和既有 Run / Audit owner 工作，不开放通用工具执行器或业务写回。

应用定时评测运行器仅在明确开发测试门禁下启动；区分系统执行主体与受委托用户，每次调度触发时重新验证权限和配置，使用确定性 Campaign / Run 交接。关闭时先取消并等待任务结束，再释放桥接层与存储；不等于生产工作进程或通用调度器。

## 图片产物链

开发测试态链已实现：

```text
strict image intent + reference-only profile
  → image_generation_adapter
  → test-only contract_fixture client
  → 单次 binary delivery
  → image_artifact_delivery_coordinator
  → 本机私有 content-addressed store 重验
  → 成功后释放产物引用
```

`image_backend_profile_configuration.py` 编译稳定摘要与超时配置，不解析凭证、端点或模型目录。`image_generation_adapter.py` 独占规范元数据构造；`image_artifact_delivery_coordinator.py` 已接通一次性交付与私有存储，失败不释放成功引用。

`image_artifact_runtime_mapper.py`、`image_artifact_response_consumer.py` 与 `inference_response.py` 的仅元数据响应构建器另行消费产物元数据，不读取二进制或替代存储。真实生图后端、引用解析器、生产存储、公开 URL 与 HTTP / 网关 / Web 交付仍未实现。详见[图片功能专题](features/image-generation-artifact-return.md)。

## 部署与生产缺口

Go Platform 与 Python worker 构成当前运行部署单元。Docker compose 编排 Platform 与本地 Console，不承担身份、secret、授权或业务执行决策。产品 Web 的正式交付需独立收口。

当前 `P3 Local Product Shell / Ops Surface` 保持 `local usable / read-only close`，历史复验入口为 `scripts/check-p3-local-product-shell-short-close-checklist.py`。它不证明生产包装、环境隔离、process supervisor 或部署恢复成立。

production secret resolver 与其 audit storage adapter 仍是后置边界；这不否定已实现的应用审计、SQL migration、本地身份和开发测试态数据库。真实 Radish、外部 provider live health、生产凭据、API key / quota / billing 和发布运行验收分别取得证据。

## 后续模块收敛方向

2026-09-06 文档审阅建议先保护现有行为，再按一个领域逐步形成实际代码边界。候选包括身份 / 成员领域提取、前端工作流状态归属、严格消费者的版本兼容集中与关键浏览器回归。

前端草案状态收敛、三条 Workflow 与三条 Prompt 自动浏览器回归已完成，并复用既有 PR / Release 配置。2026-09-14 已建立标准库依赖的 `internal/workspacepolicy`，集中权限目录与内建角色策略，并直接迁移调用点；身份记录、应用服务、assignment 生命周期与仓储事务继续保留在 `internal/httpapi`。新包不反向依赖 HTTP 或数据库，不增加转发层、角色来源或运行生命周期。部署形态、权限、API、schema 与数据库角色均未改变；具体边界和验收见[包边界与实现](platform/engineering-health-productization-remediation-v1.md#身份与成员领域包边界与实现)。

## 契约与历史证据路由

以下是仍可复验的历史设计边界，不代表当前全平台实现范围：

- `Control Plane / User Workspace / Workflow v1` 的 `product-surface-v1-boundary` 是初始规划证据；其“不实现 OIDC、数据库、API key / quota、workflow executor、confirmation、writeback 或 replay”仅限定该历史批次。
- `control-plane-data-boundary` 与 `radish-oidc-client-preconditions` 分别定义数据归属和联合身份前置条件；`workflow-definition-run-record-boundary` 定义状态流转与审计。
- `gateway-api-key-quota-readiness` 定义生产准入；`control-plane-read-formal-ui-boundary-v1` 定义早期正式 UI 边界。
- `control-plane-read-consumer-contract-v1` 对应 `contracts/typescript/control-plane-read-api.ts`，历史协议详见[Control Plane 契约](contracts/control-plane-read-side.md)。

Provider 基础证据保留在 `scripts/checks/fixtures/`，检查器位于 `scripts/`：

| 证据文件 | 检查器 | 证明范围 |
| --- | --- | --- |
| `provider-capability-matrix-v1.json` | `check-provider-capability-matrix.py` | 能力声明与 registry 一致 |
| `provider-health-smoke-v1.json` | `check-provider-health-smoke.py` | mock / 配置级健康 |
| `provider-selection-policy-v1.json` | `check-provider-selection-policy.py` | 显式选择与负向边界 |
| `provider-retry-fallback-policy-v1.json` | `check-provider-retry-fallback-policy.py` | 默认策略与历史审计合同 |
| `provider-runtime-docs-refresh.json` | `check-provider-runtime-docs-refresh.py` | 基础文档引用一致性，不证明全部叙述仍有效 |

当前受控 fallback、权限、持久化与用户流程事实优先读取对应功能专题和行为测试。历史 next dependency 不恢复为当前开发顺位。

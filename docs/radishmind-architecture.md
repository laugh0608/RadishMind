# RadishMind 系统架构

更新时间：2026-09-06

## 架构职责

本文描述当前进程、模块、请求流与数据归属，并区分实现事实、逻辑职责和后续收敛方向。能力成熟度见[能力矩阵](radishmind-capability-matrix.md)，执行顺位见[当前推进焦点](radishmind-current-focus.md)，功能细节见[功能设计入口](features/README.md)。

RadishMind 是 AI 工具、工作流、模型网关和 Copilot 集成平台，不是上层业务真相源。RadishMind 本地账户、角色与工作区成员关系是平台内部真相；Radish 自身身份、组织成员关系和业务数据仍由 Radish 负责。

## 当前进程与代码映射

| 组件 | 当前落点 | 职责与边界 |
| --- | --- | --- |
| 产品 Web | `apps/radishmind-web/` | React + Vite + TypeScript；应用、Workflow、Gateway 和管理端用户流程 |
| 本地运维 Console | `apps/radishmind-console/` | 平台发现、诊断与只读证据；不代替正式产品 Web |
| Go Platform | `services/platform/` | HTTP、身份与权限、领域编排、持久化、配额、路由、运行与审计 |
| Bridge | `services/platform/internal/bridge/` | 默认四个受控 Python stdio worker；健康握手、有界排队、取消、超时、崩溃重建、优雅关闭和请求隔离 |
| Python Gateway / Runtime | `services/gateway/`、`services/runtime/` | canonical 请求分发、Provider 适配、模型推理、规则与响应构建；`process_per_request` 保留显式回滚 |
| 契约 | `contracts/`、`contracts/typescript/` | canonical schema、产品资源协议和跨语言消费合同 |
| 评测与实验 | `datasets/`、`scripts/eval/`、`training/` | 固定任务、候选记录、离线回归与模型路线证据 |

当前大量领域、存储与执行逻辑仍在 `services/platform/internal/httpapi/` 同包，产品工作区状态也部分集中在 `App.tsx` 和 `features/control-plane-read/`。下文的逻辑职责不表示已经完成 Go 包拆分、前端目录迁移或独立服务部署。

## 四个产品面与逻辑职责

- `User Workspace`：应用、草案、运行、结果和评测的用户入口；页面消费领域状态，不复制服务端授权或发布资格。
- `Admin Control Plane`：账户、tenant、工作区成员、角色、API key、provider profile、路由、quota、价格策略和运行审计；配置与执行保持明确交接。
- `Model Gateway / API Distribution`：北向协议翻译、模型发现、请求准入、Provider 调用和脱敏请求历史；不用兼容协议建立第二套业务真相。
- `Workflow / Agent Runtime`：按明确执行档案编排 Definition、RAG、Prompt、Agent / Copilot 与受控 HTTP Tool，记录 run record、状态流转和失败边界。

核心逻辑层继续是 Client Adapters & Context Packers、Copilot Gateway / Task Router、Retrieval & Tool Layer、Model Runtime Layer、Rule Validation & Response Builder、Data / Evaluation / Training Pipeline。它们是职责划分，不要求每次请求依次经过所有层；检索与工具仅在对应执行档案允许时发生。

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

- 北向已有 `/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models` 兼容接口；内部 Copilot 链使用 `CopilotRequest / CopilotResponse / CopilotGatewayEnvelope`。产品管理资源有各自 canonical 合同，不强行归一为模型请求。
- Go 负责请求准入与领域执行边界，Python 负责模型和 AI 生态适配。一次运行可能产生多个明确的 Provider attempt；作用域、权限与 quota 不能由模型输出决定。
- 默认 retry / fallback 继续关闭或由 caller 管理；已实现的开发测试态 fallback 仅在独立 gate、允许的 API Key unary 请求、冻结路由和类型化 eligible failure 同时成立时执行。stream、非 API Key 与 production 不自动继承此能力。
- 所有高风险交接重读精确资源与当前权限，跨 owner 仅传稳定引用；客户端快照、旧资格或 readiness 投影不能代替执行授权。
- API key、token、DSN 和 Provider 原始错误不进入 argv、公开日志或响应；错误保持稳定分类与脱敏上下文。取消、超时、关闭与持久化失败不伪造成功或回退样例。

## 数据归属与持久化

| 数据 | 当前归属 | 一致性要求 |
| --- | --- | --- |
| 本地账户、凭证、外部身份绑定、Web Session | local identity repository | 显式 `(issuer, subject)` 绑定，不按 email 合并；认证失败不回退其它身份来源 |
| 工作区成员、角色与邀请 | 既有身份 / 授权 owner | invitation claim 单事务建立成员与角色并消费邀请；不另设权限真相 |
| 应用配置、模板、发布候选、运行时绑定 | 对应领域仓储 | 不可变版本、CAS 审查、精确作用域和当前资格重验 |
| Draft / Definition、Run、Session / Turn、评测 | 对应 Workflow / 应用领域 owner | 执行来源、版本、审计与历史兼容明确；跨 owner 交接不复制完整状态 |
| 结果资产与生命周期 | 显式结果资产 owner | 原始结果按已批准保存策略处理；Session 不自动成为 transcript 或通用结果库 |
| Gateway Request、quota、价格快照 | 各自独立 owner | Provider 前原子准入、不可变价格、可信 reported usage；不构成 cost ledger 或 invoice |

`memory_dev` 用于不要求跨进程恢复的测试。聚合 `sqlite_dev` 通过共享文件与 runtime 管理组件 migration 和仓储；显式 `postgres_dev_test` 为组件保留连接、migration、marker / checksum 与运行角色边界。

Prompt / Agent Runtime Assignment、Session / Turn、Run 投影、Action Safety snapshot、Evaluation Plan / Campaign / Schedule / Occurrence 复用 Workflow Run Store 已审查的持久化边界。复用连接不改变各资源的领域职责。

未知 selector、禁用模式、migration 不兼容、查询或存储失败均失败关闭；不能回退 memory。双数据库的事务、并发、权限、重启与查询行为分别验证，不能用其中一个替代另一个。详细配置见[SQLite 专题](platform/local-sqlite-dev-persistence-v1.md)与[运行手册](platform/platform-service-operations-runbook-v1.md)。

## Session、工具与后台调度

应用 Session / Turn、显式结果资产和运行链已有开发测试态持久化。历史通用 `/v1/session/metadata` 与 checkpoint metadata-only 路由仍不提供长期记忆、任意跨轮恢复或 replay 执行。

通用 `/v1/tools/actions` 保持 blocked shell。Workflow HTTP Tool 通过独立 action plan、confirmation、执行档案和既有 Run / Audit owner 工作，不开放通用工具执行器或业务写回。

应用定时评测 runner 仅在明确开发测试 gate 下启动；区分 system actor 与 delegated user，每次 occurrence 重新验证权限和配置，使用确定性 Campaign / Run 交接。关闭时先 cancel / join，再释放 bridge 与 store；不等于 production worker 或通用 scheduler。

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

`image_backend_profile_configuration.py` 编译稳定 digest 与 timeout，不解析 credential、endpoint 或 model directory。`image_generation_adapter.py` 独占 canonical metadata 构造；`image_artifact_delivery_coordinator.py` 已接通一次性交付与私有存储，失败不释放成功引用。

`image_artifact_runtime_mapper.py`、`image_artifact_response_consumer.py` 与 `inference_response.py` 的 metadata-only response builder 另行消费 artifact metadata，不读取二进制或替代 store。真实生图 backend、reference resolver、production storage、public URL 与 HTTP / Gateway / Web 交付仍未实现。详见[图片功能专题](features/image-generation-artifact-return.md)。

## 部署与生产缺口

Go Platform 与 Python worker 构成当前运行部署单元。Docker compose 编排 Platform 与本地 Console，不承担身份、secret、授权或业务执行决策。产品 Web 的正式交付需独立收口。

当前 `P3 Local Product Shell / Ops Surface` 保持 `local usable / read-only close`，历史复验入口为 `scripts/check-p3-local-product-shell-short-close-checklist.py`。它不证明生产包装、环境隔离、process supervisor 或部署恢复成立。

production secret resolver 与其 audit storage adapter 仍是后置边界；这不否定已实现的应用审计、SQL migration、本地身份和开发测试态数据库。真实 Radish、外部 provider live health、生产凭据、API key / quota / billing 和发布运行验收分别取得证据。

## 后续模块收敛方向

2026-09-06 文档审阅建议先保护现有行为，再按一个领域逐步形成实际代码边界。候选包括身份 / 成员领域提取、前端工作流状态归属、严格消费者的版本兼容集中与关键浏览器回归。

当前包名、部署形态、权限、API、schema、数据库角色和 CI 尚未改变。具体依赖分析、迁移顺序和验收在[工程健康专题](platform/engineering-health-productization-remediation-v1.md#2026-09-06-复审与后续方向)中确认；不把目标目录写成已经存在的模块。

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

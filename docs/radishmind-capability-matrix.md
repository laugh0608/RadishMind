# RadishMind 能力矩阵

更新时间：2026-09-06

## 读取口径

当前成熟度为**内部开发者预览**。本矩阵区分实现、验证环境、外部验证和生产边界；`completed` 只关闭对应专题范围，不自动升级平台成熟度。

表中“已有证据”引用各专题的历史验收，不表示本次文档更新重新执行了浏览器、PostgreSQL 或真实 Provider 联调。2026-09-06 的实际审阅验证见[本周周志](devlogs/2026-W36.md)。

## 当前能力

| 能力 | 开发测试态实现与证据 | 仍未成立或未完成 | 权威专题 |
| --- | --- | --- | --- |
| Workflow 草案与定义 | 创建、编辑、校验、版本 / 生命周期、三存储、冲突 / 重启、审查与受控运行；模板目录和结构化输入已关闭 | 任意节点执行、通用 replay、自动 release 与业务写回 | [Workflow](features/workflow/README.md) |
| 应用工作区 | 配置、Prompt / Agent 档案、受控会话、结果保存 / 资产库、精确运行交接与浏览器链 | 通用 transcript、公开分享、永久 purge、自治 agent loop | [User Workspace](features/user-workspace/README.md) |
| 应用评测 | Comparison、Case / Suite、人工 decision、Plan / Campaign、受控 Schedule / Occurrence 与双数据库证据 | 真实 Provider 调度验收、production worker、通用 scheduler；持续团队收益指标仍待建立 | [定时回归评测](features/user-workspace/application-evaluation-scheduled-regression-campaign-dev-test-v1.md) |
| Gateway | 三类北向推理协议、Models、SSE、Provider 选择、stdio worker pool；开发测试态 API key / quota、reported usage、价格快照与受控 unary fallback | 真实 Provider fallback、生产 API key / quota / price、限流、billing ledger、optional live health | [模型网关](features/model-gateway-api-distribution.md) |
| 本地身份与成员 | 账户、Web Session、角色、成员、凭据轮换、确定性 OIDC 与双数据库 / 浏览器证据 | 真实 Radish 联调 `real_radish_integration_deferred`；production auth、MFA、账户恢复 | [管理端](features/admin-control-plane/README.md) |
| 工作区邀请 | A / B canonical 与三存储、C 五条 strict HTTP、D Pencil 与人工批准已完成 | React strict consumer 与双数据库产品 / 浏览器链待批次 E 独立授权；邮件、目录搜索、管理员邀请关闭 | [邀请专题](features/admin-control-plane/workspace-member-invitation-claim-expiry-governance-dev-test-v1.md) |
| 受控工具与 Action Safety | Workflow HTTP Tool 的独立确认 / 执行链与安全资格快照已实现 | 通用 `/v1/tools/actions` 仍 blocked；写方法 Tool、`write_allowed_by_policy`、通用动作执行器未开放 | [HTTP Tool](features/workflow/workflow-definition-http-tool-v1.md)、[Action Safety](features/workflow/action-safety-ladder-candidate-action-execution-eligibility-dev-test-v1.md) |
| 通用 Session / Checkpoint | metadata-only 契约与只读 shell；与应用 Session / Turn 区分 | 通用 durable checkpoint、长期记忆、跨轮恢复执行器 | [架构](radishmind-architecture.md) |
| 图片产物 | fixture client → adapter → 一次性 binary delivery → 私有存储 coordinator 已完成；metadata-only response builder 独立存在 | 真实生图 backend、reference resolver、生产存储、public URL、HTTP / Gateway / Web 交付 | [图片专题](features/image-generation-artifact-return.md) |
| 持久化与审计 | memory / SQLite / PostgreSQL 开发测试仓储、migration、CAS、scope、no fallback 与重启测试 | 生产数据库资源与运行验收；production secret backend 的 audit store 未因应用审计完成而成立 | [SQLite](platform/local-sqlite-dev-persistence-v1.md)、[运行手册](platform/platform-service-operations-runbook-v1.md) |
| UI 与工程验证 | Family UI、人工浏览器验收、消费层测试、构建预算、Go race / vet、PR / release CI 与 PostgreSQL integration 配置 | CI 尚无自动浏览器回归；部分模块覆盖率不等于全 UI 覆盖率；结构收敛待实施 | [工程健康](platform/engineering-health-productization-remediation-v1.md) |
| 部署与运维 | 本地 Console、启动入口、Docker 静态边界和历史本地容器 smoke | production secret、环境隔离、process supervisor、产品 Web 正式包装、真实镜像发布和恢复演练 | [部署说明](../deploy/README.md) |
| 模型适配 | 原始 / 修复双轨、小规模 holdout 和 builder 审查记录 | raw 晋级、训练准入、稳定业务收益和生产模型声明 | [训练目录](../training/README.md) |
| 外部业务接入 | RadishFlow / Radish 的协议、样例、candidate 与离线回归证据；Catalyst 文档预留 | 真实挂载点与验收环境未齐备，不建立模拟接入完成声明 | [集成入口](integrations/README.md) |

## 下一动作与验收区分

当前执行项只在[当前推进焦点](radishmind-current-focus.md)维护。邀请状态保持 `workspace_member_invitation_claim_expiry_governance_dev_test_v1_batch_d_pencil_approved_batch_e_ready`；R2 至 R6 与已关闭产品专题不因本次复审重开批次。

后续专题分别记录代码 / 契约、单元 / 存储测试、浏览器连续链、真实 Provider、内部用户使用和生产验收。未执行填“未验证”，外部依赖缺失填具体阻塞，不从相邻能力推算 ready。使用指标与模型效果口径见[产品范围](radishmind-product-scope.md#使用与效果验收)。

## 历史证据索引

以下基础证据仍由既有门禁复验。表内标识不是完整产品能力声明，也不是下一顺位；长 readiness / review / refresh 链从[任务卡索引](task-cards/README.md)和[平台专题](platform/README.md)追溯。

fixture 默认位于 `scripts/checks/fixtures/`；Control Plane checker 位于 `scripts/checks/control_plane/`，Provider checker 位于 `scripts/`。

| 契约 / 证据 | 文件 | 检查器 |
| --- | --- | --- |
| `product-surface-v1-boundary` | `product-surface-v1-boundary.json` | `check-product-surface-v1-boundary.py` |
| `control-plane-data-boundary` | `control-plane-data-boundary.json` | `check-control-plane-data-boundary.py` |
| `radish-oidc-client-preconditions` | `radish-oidc-client-preconditions.json` | `check-radish-oidc-client-preconditions.py` |
| `gateway-api-key-quota-readiness` | `gateway-api-key-quota-readiness.json` | `check-gateway-api-key-quota-readiness.py` |
| `workflow-definition-run-record-boundary` | `workflow-definition-run-record-boundary.json` | `check-workflow-definition-run-record-boundary.py` |
| provider capability matrix | `provider-capability-matrix-v1.json` | `check-provider-capability-matrix.py` |
| provider health smoke | `provider-health-smoke-v1.json` | `check-provider-health-smoke.py` |
| provider selection policy | `provider-selection-policy-v1.json` | `check-provider-selection-policy.py` |
| provider retry/fallback policy | `provider-retry-fallback-policy-v1.json` | `check-provider-retry-fallback-policy.py` |
| `provider-runtime-docs-refresh` | `provider-runtime-docs-refresh.json` | `check-provider-runtime-docs-refresh.py` |

- `control-plane-read-formal-ui-boundary-v1` 固定早期正式 UI 边界；`control-plane-read-consumer-contract-v1` 是 TypeScript consumer contract，不代表全部 UI 的实时能力。
- production secret backend contract 对应 `production-ops-secret-backend-contract.json`；secret ref schema 为 `contracts/production-secret-reference.schema.json`。真实 production secret backend 仍未完成，不能用静态引用或 test-only fake resolver 替代。
- Provider 基础策略的默认关闭口径不覆盖独立开发测试 gate 下已实现的受控 fallback；当前事实以 Gateway 功能专题为准。

### 历史只读页面与图片消费合同

以下标识定位既有 fixture 的原始证明范围。相关页面在 `apps/radishmind-web/`，当前应用写入、身份和执行链以上表为准，不由历史只读合同推断。

| 检查标识 | 原始页面、状态或合同 |
| --- | --- |
| `control-plane-read-shared-shell-v1` | shared shell |
| `control-plane-read-admin-tenant-overview-v1` | `admin-tenant-overview` |
| `control-plane-read-admin-audit-log-v1` | `admin-audit-log` |
| `control-plane-read-workspace-applications-v1` | `workspace-applications` |
| `control-plane-read-workspace-api-keys-v1` | `workspace-api-keys` |
| `control-plane-read-workspace-usage-quota-v1` | `workspace-usage-quota` |
| `control-plane-read-workspace-workflow-definitions-v1` | `workspace-workflow-definitions` |
| `control-plane-read-workspace-run-history-v1` | `workspace-run-history` |
| `control-plane-read-formal-ui-readiness-close-v1` | surface matrix |
| `control-plane-read-dev-live-consumer-v1` | dev-only live read consumer |
| `workflow-function-surface-boundary-v1` | `function_surface_boundary_defined` |
| `workflow-application-detail-read-v1` | `workflow_application_detail_read_defined` |
| `workflow-definition-detail-read-v1` | `workflow_definition_detail_read_defined` |
| `workflow-run-detail-read-v1` | `workflow_run_detail_read_defined` |
| `workflow-blocked-action-preview-v1` | `workflow_blocked_action_preview_defined` |
| `workflow-confirmation-placeholder-read-v1` | `workflow_confirmation_placeholder_read_defined` |
| `workflow-draft-designer-offline-v1` | 本地草案设计；上层缺少挂载点不阻塞 RadishMind 已明确的内部任务 |
| `workflow-draft-validation-inspector-offline-v1` | `workflow_draft_validation_inspector_offline_defined` |
| `workflow-execution-plan-preview-offline-v1` | `workflow_execution_plan_preview_offline_defined` |
| `workflow-runtime-readiness-inspector-offline-v1` | `workflow_runtime_readiness_inspector_offline_defined` |
| `workflow-function-surface-readiness-close-v1` | `workflow_function_surface_readiness_closed` |
| `image-artifact-runtime-mapper-runtime-implementation-v1` | metadata-only artifact runtime mapper |
| `image-artifact-response-consumer-runtime-implementation-v1` | metadata-only response consumer runtime |
| `image-artifact-response-builder-runtime-integration-implementation-v1` | `image_artifact_response_builder_runtime_integration_implemented` |

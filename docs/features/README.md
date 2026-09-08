# RadishMind 功能设计文档入口

更新时间：2026-09-08

## 文档目的

本目录承载产品能力与功能专题。入口只提供当前目标、专题路由和设计要求；完整实现事实与边界在对应专题维护，历史批次进入任务卡与周志。

当前成熟度为内部开发者预览。四个产品面共同服务内部开发者的应用创建、受控运行、结果审查与回归验证；详细定位和使用指标见[产品范围](../radishmind-product-scope.md)。

## 当前目标

[工作区成员邀请、认领与到期治理（开发 / 测试态）v1](admin-control-plane/workspace-member-invitation-claim-expiry-governance-dev-test-v1.md)已完成 A 至 E，状态为 `workspace_member_invitation_claim_expiry_governance_dev_test_v1_completed`。本地成员入会的前端、双数据库与浏览器链已闭合；下一目标回到真实用户阻塞和维护边界的选择，不派生同层续批。

[应用定时回归评测](user-workspace/application-evaluation-scheduled-regression-campaign-dev-test-v1.md)、[工作区 Workflow 模板目录](workflow/workspace-workflow-template-catalog-review-controlled-derivation-dev-test-v1.md)和[Action Safety Ladder](workflow/action-safety-ladder-candidate-action-execution-eligibility-dev-test-v1.md)保持完成关闭。真实 Radish OIDC 为 `real_radish_integration_deferred`，应用运行观测后续仍为 `no_entry`。

结构收敛、自动浏览器回归和内部试用见[工程健康专题](../platform/engineering-health-productization-remediation-v1.md#2026-09-06-复审与后续方向)，均为待确定实施范围的候选，不自动创建批次。当前排期与停止线只以[当前推进焦点](../radishmind-current-focus.md)为准。

## 专题分层

| 层级 | 路径 | 职责 |
| --- | --- | --- |
| 产品面 | `docs/features/*.md` | 长期目标、用户和范围 |
| 功能与复杂页面 | 对应产品面的子目录 | 页面 / 界面专题承载流程、状态、数据归属、实施与验收 |
| 平台横切能力 | `docs/platform/` | 身份、持久化、运行与工程边界 |
| 外部集成 | `docs/integrations/` | 外部项目或后端接入的上游、协议与验收条件 |
| 实现批次 | `docs/task-cards/` | 高风险实施范围、前置条件与完成记录 |
| 历史证据 | `docs/devlogs/`、manifest、summary、run record | 批次事实、实验与完整验证输出 |

## 产品面导航

| 产品面 | 大方向 | 具体功能入口 |
| --- | --- | --- |
| `User Workspace` | [用户工作区](user-workspace.md) | [应用、配置、会话、结果与评测](user-workspace/README.md) |
| `Admin Control Plane` | [管理端](admin-control-plane.md) | [本地身份、成员、角色与管理配置](admin-control-plane/README.md) |
| `Model Gateway / API Distribution` | [模型网关与 API 分发](model-gateway-api-distribution.md) | [Provider、准入、用量、价格与 fallback](gateway/README.md) |
| `Workflow / Agent Runtime` | [工作流与运行时](workflow-agent-runtime.md) | [草案、Definition、执行、模板与 RAG](workflow/README.md) |
| 横切适配 | [图片生成与产物返回](image-generation-artifact-return.md) | fixture / 私有存储已完成；真实 backend 和公开交付独立验收，不另立第五产品主线 |

## 常用用户流程

- 创建与恢复：[草案设计](workflow/draft-designer-editing-model-v2.md)、[保存草案库生命周期](workflow/saved-workflow-draft-library-lifecycle-organization-dev-test-v1.md)、[模板目录与派生](workflow/workspace-workflow-template-catalog-review-controlled-derivation-dev-test-v1.md)。
- 接入与受控运行：[应用 API 接入与调用 v1](user-workspace/application-api-integration-invocation-v1.md)、[应用发布治理与晋级审查 v1](user-workspace/application-publish-governance-promotion-v1.md)、[运行开发测试指南](user-workspace/application-controlled-runtime-dev-test-guide.md)、[Definition HTTP Tool](workflow/workflow-definition-http-tool-v1.md)。
- 结果与回归：[结果资产库](user-workspace/application-result-artifact-library-controlled-export-dev-test-v1.md)、[应用评测计划与 Campaign](user-workspace/application-evaluation-campaign-controlled-execution-dev-test-v1.md)、[定时回归](user-workspace/application-evaluation-scheduled-regression-campaign-dev-test-v1.md)。
- 团队协作：[本地账户与联合登录](admin-control-plane/local-account-radish-oidc-federated-login-v1.md)、[工作区邀请](admin-control-plane/workspace-member-invitation-claim-expiry-governance-dev-test-v1.md)。
- 平台支撑：[SQLite 持久化](../platform/local-sqlite-dev-persistence-v1.md)、[运行手册](../platform/platform-service-operations-runbook-v1.md)、[Family UI](user-workspace/radishmind-family-ui-productization-v1.md)。

## 设计与验收要求

1. 先写目标用户、具体任务、现有路径与阻塞证据，说明完成后用户能做什么；没有真实使用样本时明确为待验证假设。
2. 功能专题写清核心流程、状态、数据边界、当前实现、下一步和停止线。复杂页面明确单一职责、作用域切换、异步回调、冲突与恢复动作。
3. 代码 / 契约、存储测试、人工浏览器、自动浏览器回归、真实 Provider、团队试用与生产验收分别标明。批次关闭、checker 通过和部分模块覆盖率不能替代实际用户收益。
4. 使用与效果指标按[产品范围的验收口径](../radishmind-product-scope.md#使用与效果验收)记录样本量、环境、失败原因与精确版本；先建立基线，再确定改进目标。
5. 普通 UI、文案、布局和只读证据组织复用既有测试、web 构建、消费端冒烟验证与聚合门禁。只有协议、schema、执行边界、生产声明、外部服务或高风险能力变化时，才考虑专项 task card / fixture / checker。
6. 实现前按当前任务确认范围。文档规划不自动授权依赖安装、后台服务、模块迁移、CI、真实模型、外部接入或生产部署。
7. 功能文档正文默认中文；schema / fixture / checker ID、路径、状态锚点和必要专业标识保留原文，详见[文档语言治理 v1](../document-language-governance-v1.md)。
8. 完成后回写对应专题，当前焦点只保留当前目标与下一动作；长验证和历史流水进入周志，不复制进本入口。

## 历史证据路由

早期只读页面、readiness 与实现准入链从[Workflow 子目录](workflow/README.md)、[平台目录](../platform/README.md)和[历史任务卡索引](../task-cards/README.md)按需追溯，不再在总入口逐项复制长状态。

Production Secret Backend / Storage Adapter 的历史材料见[Storage Adapter Evidence Rollup v1](../platform/production-secret-backend-audit-store-storage-adapter-evidence-rollup-v1.md)。旧依赖与 review / refresh 不提供当前开发顺位；生产 secret audit store 的缺口不否定已实现的应用审计与开发测试态数据库。

部分既有检查器仍引用以下草案编辑专题名称；这些是历史功能合同入口，不代表新增工作：

| 历史专题 | 路由 |
| --- | --- |
| Workflow Draft Node Attribute Editing Model v1 | [节点属性编辑](workflow/draft-node-attribute-editing-model-v1.md) |
| Workflow Review Handoff Active Draft v1 | [活动草案审查交接](workflow/review-handoff-active-draft-v1.md) |
| Workflow Node Designer Review Handoff v1 | [设计器审查交接](workflow/node-designer-review-handoff-v1.md) |
| Workflow Node Designer Graph Review Handoff Refinement v1 | [图审查交接细化](../task-cards/workflow-node-designer-graph-review-handoff-refinement-v1-plan.md) |
| Workflow Node Designer Persisted Layout v1 | [布局持久化](workflow/node-designer-persisted-layout-v1.md) |
| Workflow Node Designer Edge Editing Save Preconditions v1 | [连线编辑与保存前置](workflow/node-designer-edge-editing-save-preconditions-v1.md) |
| Workflow Node Designer Layout Review Findings v1 | [布局审查发现](../task-cards/workflow-node-designer-layout-review-findings-v1-plan.md) |
| Workflow Node Designer Builder Interaction Polish v1 | [编辑器交互](../task-cards/workflow-node-designer-builder-interaction-polish-v1-plan.md) |
| Workflow Node Designer Validation Overlay Navigation v1 | [校验叠层导航](../task-cards/workflow-node-designer-validation-overlay-navigation-v1-plan.md) |

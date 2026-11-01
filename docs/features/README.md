# RadishMind 功能设计文档入口

更新时间：2026-11-01

## 文档目的

本目录承载产品能力与功能专题。入口只提供当前目标、专题路由和设计要求；完整实现事实与边界在对应专题维护，历史批次进入任务卡与周志。

当前成熟度为内部开发者预览。四个产品面共同服务内部开发者的应用创建、受控运行、结果审查与回归验证；详细定位和使用指标见[产品范围](../radishmind-product-scope.md)。

## 当前目标

当前开发主线与停止线由[当前推进焦点](../radishmind-current-focus.md)维护；受控开发团队的阶段 0 工作集中在[任务卡](../task-cards/development-task-probe-v1-plan.md)，四项退出条件见[实施与评测子专题](workflow/controlled-software-development-team/implementation-and-evaluation-plan-v1.md#阶段-0-采用复评与剩余退出条件)。登记与局部技术验收不代表正式采用。

[本机持久工作台](user-workspace/local-persistent-workbench-delivery-v1.md)是交付基座；技术证据、独立用户及未验平台按该专题分别记录。真实模型继续暂停；既有完成专题不由本次导航整理派生续批。未来 Gateway 解释与尝试轨迹随受控开发阶段 1 统筹，应用运行观测保持 `no_entry`。

## 专题分层

[产品范围](../radishmind-product-scope.md)与[路线图](../radishmind-roadmap.md)构成整体规划；下表和产品面导航是能力地图，产品面是大专题，具体功能及职责专题承接子专题。逻辑层级不强制等于目录深度：受控开发保留主功能及五个职责文件，阶段 0 使用一张集中任务卡。

每个功能指定一个主要归属，跨产品面只登记协同；平台与集成是支撑主线，不新增第五产品面。任务状态只在已指定任务卡或功能文档维护，地图只链接来源。任务完成、阶段退出、真实质量和生产准入分别验收。

| 层级 | 路径 | 职责 |
| --- | --- | --- |
| 产品面与跨页面功能 | `docs/features/*.md` | 长期目标、用户和范围；跨页面功能不另立产品面 |
| 功能与复杂页面 | 对应产品面的子目录 | 页面 / 界面专题承载流程、状态、数据归属、实施与验收 |
| 平台横切能力 | `docs/platform/` | 身份、持久化、运行与工程边界 |
| 外部集成 | `docs/integrations/` | 外部项目或后端接入的上游、协议与验收条件 |
| 实现批次 | `docs/task-cards/` | 高风险实施范围、前置条件与完成记录 |
| 历史证据 | `docs/devlogs/`、清单、摘要、运行记录 | 批次事实、实验与完整验证输出 |

## 产品面导航

| 产品面 | 大方向 | 具体功能入口 |
| --- | --- | --- |
| `User Workspace` | [用户工作区](user-workspace.md) | [应用、配置、会话、结果与评测](user-workspace/README.md) |
| `Admin Control Plane` | [管理端](admin-control-plane.md) | [本地身份、成员、角色与管理配置](admin-control-plane/README.md) |
| `Model Gateway / API Distribution` | [模型网关与 API 分发](model-gateway-api-distribution.md) | [提供方、准入、用量、价格与回退](gateway/README.md) |
| `Workflow / Agent Runtime` | [工作流与运行时](workflow-agent-runtime.md) | [草案、定义、执行、模板与 RAG](workflow/README.md) |
| 横切适配 | [图片生成与产物返回](image-generation-artifact-return.md) | 固定响应测试后端与私有存储已完成；真实后端和公开交付独立验收，不另立第五产品主线 |
| 跨页面体验 | [Web 中英国际化与本地化](web-internationalization-zh-en-v1.md) | A 至 F 在登记的可达范围完成关闭；界面语言不改写业务语言、协议字段、用户原文或导出 JSON；日程双语浏览器证据限定合成投影 |

## 活跃专题的任务与验收来源

| 主要归属 | 子专题 | 任务与验收唯一来源 | 协同 |
| --- | --- | --- | --- |
| User Workspace | [本机持久工作台交付](user-workspace/local-persistent-workbench-delivery-v1.md) | 原文“实施范围与验证安排”“连续实施与完成依据”及实际证据；不另建任务卡 | 管理端身份、Workflow、Gateway、平台运维 |
| Workflow / Agent Runtime | [受控软件开发团队](workflow/controlled-software-development-team-v1.md)及五个职责子专题 | [阶段 0 集中任务卡](../task-cards/development-task-probe-v1-plan.md)维护任务结果；[实施子专题](workflow/controlled-software-development-team/implementation-and-evaluation-plan-v1.md)维护阶段 0–4 的准入 / 退出 | User Workspace 审阅、Gateway 传输、平台隔离 |
| Model Gateway / API Distribution | [路由决策解释与尝试轨迹](gateway/gateway-routing-decision-explanation-attempt-trace-v1.md) | 原专题的阶段范围与验收；尚不独立开实施线 | 受控开发阶段 1 运行详情 |
| User Workspace | [跨版本回归审查](user-workspace/application-cross-version-regression-review-dev-test-v1.md) | 原候选专题的触发依据，未获选前不建任务批次 | Workflow / Prompt 评测 |

平台支撑由[平台索引](../platform/README.md)路由身份、存储、运行和工程合同；外部接入由[集成索引](../integrations/README.md)路由上游与验收条件。跨主线任务引用原合同，不在两处勾选同一完成状态。历史细案继续沿原目录查阅，本轮只梳理活跃工作链。

## 常用用户流程

- 创建与恢复：[草案设计](workflow/draft-designer-editing-model-v2.md)、[保存草案库生命周期](workflow/saved-workflow-draft-library-lifecycle-organization-dev-test-v1.md)、[模板目录与派生](workflow/workspace-workflow-template-catalog-review-controlled-derivation-dev-test-v1.md)。
- 接入与受控运行：[应用 API 接入与调用 v1](user-workspace/application-api-integration-invocation-v1.md)、[应用发布治理与晋级审查 v1](user-workspace/application-publish-governance-promotion-v1.md)、[运行开发测试指南](user-workspace/application-controlled-runtime-dev-test-guide.md)、[工作流定义的 HTTP 工具](workflow/workflow-definition-http-tool-v1.md)。
- 结果与回归：[结果资产库](user-workspace/application-result-artifact-library-controlled-export-dev-test-v1.md)、[应用评测计划与受控批次](user-workspace/application-evaluation-campaign-controlled-execution-dev-test-v1.md)、[定时回归](user-workspace/application-evaluation-scheduled-regression-campaign-dev-test-v1.md)。
- 团队协作：[本地账户与联合登录](admin-control-plane/local-account-radish-oidc-federated-login-v1.md)、[工作区邀请](admin-control-plane/workspace-member-invitation-claim-expiry-governance-dev-test-v1.md)。
- Agent 协作开发：[功能主专题与五个子专题](workflow/controlled-software-development-team-v1.md)，覆盖角色、模型、受限执行、任务 / 讨论 / 交付及阶段安排；设计登记不代表已经可执行。
- 平台支撑：[SQLite 持久化](../platform/local-sqlite-dev-persistence-v1.md)、[运行手册](../platform/platform-service-operations-runbook-v1.md)、[Family UI](user-workspace/radishmind-family-ui-productization-v1.md)。

## 设计与验收要求

1. 先写目标用户、具体任务、现有路径与阻塞证据，说明完成后用户能做什么；没有真实使用样本时明确为待验证假设。
2. 功能专题写清核心流程、状态、数据边界、当前实现、下一步和停止线。复杂页面明确单一职责、作用域切换、异步回调、冲突与恢复动作。
3. 代码 / 契约、存储测试、人工浏览器、自动浏览器回归、真实提供方、团队试用与生产验收分别标明。批次关闭、检查器通过和部分模块覆盖率不能替代实际用户收益。
4. 使用与效果指标按[产品范围的验收口径](../radishmind-product-scope.md#使用与效果验收)记录样本量、环境、失败原因与精确版本；先建立基线，再确定改进目标。
5. 普通 UI、文案、布局和只读证据组织复用既有测试、web 构建、消费端冒烟验证与聚合门禁。只有协议、schema、执行边界、生产声明、外部服务或高风险能力变化时，才考虑专项任务卡、测试样例或检查器。
6. 实现前按当前任务确认范围。文档规划不自动授权依赖安装、后台服务、模块迁移、CI、真实模型、外部接入或生产部署。
7. 功能文档正文默认中文；schema / fixture / checker ID、路径、状态锚点和必要专业标识保留原文，详见[文档语言治理 v1](../document-language-governance-v1.md)。
8. 完成后回写对应专题，当前焦点只保留当前目标与下一动作；长验证和历史流水进入周志，不复制进本入口。

## 历史证据路由

早期只读页面、准备状态与实现准入链从[Workflow 子目录](workflow/README.md)、[平台目录](../platform/README.md)和[历史任务卡索引](../task-cards/README.md)按需追溯，不再在总入口逐项复制长状态。

Production Secret Backend / Storage Adapter 的历史材料见[存储适配器证据汇总 v1](../platform/production-secret-backend-audit-store-storage-adapter-evidence-rollup-v1.md)。旧依赖与审查 / 更新记录不提供当前开发顺位；生产密钥审计存储的缺口不否定已实现的应用审计与开发测试态数据库。

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

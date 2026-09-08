# RadishMind 文档入口

更新时间：2026-09-08

## 阅读原则

`docs/` 是 RadishMind 的正式文档源。入口负责定位和路由，功能 / 平台专题负责当前事实与长期边界，任务卡负责实施批次，周志与运行记录负责历史证据。

当前成熟度为**内部开发者预览**。工作区成员邀请的开发测试态专题已完成关闭；当前状态、下一目标选择和停止线只以[当前推进焦点](radishmind-current-focus.md)为准，专题完成不扩大生产能力。

## 从哪里开始

| 要回答的问题 | 优先阅读 |
| --- | --- |
| 今天推进什么、有什么阻塞 | [当前推进焦点](radishmind-current-focus.md) |
| 项目为谁服务、要完成什么任务 | [产品范围与目标](radishmind-product-scope.md) |
| 哪些能力已实现、在哪种环境验证 | [能力矩阵](radishmind-capability-matrix.md) |
| 下一阶段如何选择工作 | [路线图](radishmind-roadmap.md)、[工程健康与产品化](platform/engineering-health-productization-remediation-v1.md) |
| 功能流程、页面与实施范围 | [功能设计入口](features/README.md) |
| 进程、模块、请求与数据归属 | [系统架构](radishmind-architecture.md) |
| 本地使用、配置与故障 | [项目指南](radishmind-project-guide.md)、[运行手册](platform/platform-service-operations-runbook-v1.md) |
| 跨项目接入与模型后端 | [集成契约](radishmind-integration-contracts.md)、[集成专题](integrations/README.md) |
| 协作、授权与验证粒度 | [Agent 协作与执行规则](agent-collaboration.md) |
| 代码组织、UI 与提交规则 | [代码规范](radishmind-code-standards.md)、[UI 差异附录](ui-addendum.md)、[分支 ADR](adr/0001-branch-and-pr-governance.md) |

长期战略见[战略定义](radishmind-strategy.md)，未排期想法见[产品机会池](radishmind-product-ideas.md)。二者不替代当前焦点。只在追溯时展开[任务卡](task-cards/README.md)、[周志](devlogs/README.md)与长实验记录。

## 专题与资产入口

- 产品：[用户工作区](features/user-workspace/README.md)、[管理端](features/admin-control-plane/README.md)、[Workflow](features/workflow/README.md)、[Gateway](features/gateway/README.md)、[图片适配](features/image-generation-artifact-return.md)。
- 横切能力：[平台专题](platform/README.md)、[契约专题](contracts/README.md)、[部署说明](../deploy/README.md)。
- 数据与验证：[canonical 契约](../contracts/README.md)、[数据集](../datasets/README.md)、[训练与模型实验](../training/README.md)、[脚本入口](../scripts/README.md)。
- 设计：[Family UI 产品化](features/user-workspace/radishmind-family-ui-productization-v1.md)、[语言治理](document-language-governance-v1.md)。
- 历史专题：[Control Plane Read-Side](contracts/control-plane-read-side.md)、[Production Secret Storage Adapter 静态准入](platform/production-secret-backend-audit-store-storage-adapter-static-readiness-guide-v1.md)。

## 事实与验证口径

- 功能 `completed` 只对应该专题的已批准范围；分别说明代码、存储测试、人工浏览器、自动浏览器、真实 Provider、团队试用和生产验收，不推算未验证层次。
- 当前应用与 Workflow 已有开发测试态数据库、身份、受控执行和结果链。历史 read-only / metadata-only 契约的停止线只约束其原始批次。
- 人工浏览器验收记录不等于 CI 已有自动浏览器回归；消费层覆盖率不等于全 UI 覆盖率。具体审阅证据见[2026-W36](devlogs/2026-W36.md)。
- 缺失值、外部阻塞和未执行测试分别写明；不把 mock、静态 schema 或本地试验写成 production ready。

## 历史契约导航

下表保留当前 checker 引用的早期设计入口。停止线只限定对应历史文档或批次，不覆盖[能力矩阵](radishmind-capability-matrix.md)中的后续实现；旧 next dependency 不恢复为排期。

| 历史入口 | 原始证明范围 |
| --- | --- |
| Control Plane / User Workspace / Workflow v1 · `product-surface-v1-boundary` | 初始四产品面设计，不直接实现 OIDC、数据库、API key / quota、workflow executor、confirmation、writeback 或 replay |
| `control-plane-data-boundary` | 数据归属设计，不创建数据库 schema 或 migration |
| `radish-oidc-client-preconditions` | 身份前置条件，不接真实 OIDC |
| `gateway-api-key-quota-readiness` | 生产准入设计，不发放真实 API key |
| `workflow-definition-run-record-boundary` | workflow definition、run record、状态流转与审计边界 |
| `control-plane-read-consumer-contract-v1` | TypeScript consumer contract，不实现完整 read-side API |
| `control-plane-read-formal-ui-boundary-v1` | 早期正式 UI 边界，不直接实现 OIDC、数据库、API key / quota、workflow executor、confirmation、writeback 或 replay |
| `control-plane-read-formal-ui-implementation-readiness-v1` | 正式 UI 实现 readiness，只证明当时的工程前置 |

Provider Runtime & Health v1 的基础证据包括 `provider-capability-matrix-v1`、`provider-health-smoke-v1`、`provider-selection-policy-v1`、`provider-retry-fallback-policy-v1` 与 `provider-runtime-docs-refresh`。不把 provider health 写成 production readiness；当前受控 fallback 等后续能力以 Gateway 专题为准。

早期 `apps/radishmind-web/` 只读页面的复验定位如下；它们的合同不描述后续应用写入与执行能力：

| 页面 / 聚合 | 历史检查标识 |
| --- | --- |
| shared shell | `control-plane-read-shared-shell-v1` |
| `admin-tenant-overview` | `control-plane-read-admin-tenant-overview-v1` |
| `admin-audit-log` | `control-plane-read-admin-audit-log-v1` |
| `workspace-applications` | `control-plane-read-workspace-applications-v1` |
| `workspace-api-keys` | `control-plane-read-workspace-api-keys-v1` |
| `workspace-usage-quota` | `control-plane-read-workspace-usage-quota-v1` |
| `workspace-workflow-definitions` | `control-plane-read-workspace-workflow-definitions-v1` |
| `workspace-run-history` | `control-plane-read-workspace-run-history-v1` |
| surface matrix | `control-plane-read-formal-ui-readiness-close-v1` |

Workflow 早期聚合 `workflow-function-surface-readiness-close-v1` 的状态为 `workflow_function_surface_readiness_closed`，只说明离线页面合同收口。图片历史 `image-artifact-runtime-mapper-runtime-implementation-v1` 是 metadata-only runtime mapper；`image-artifact-response-builder-runtime-integration-implementation-v1` 的 `image_artifact_response_builder_runtime_integration_implemented` 只证明元数据响应构建接线，不替代后续私有存储 coordinator 或真实 backend 验收。

## 文档维护

1. 更新事实时同时消除入口中的相反断言。未实现、已实现但未启用、开发测试验证、外部待验收必须明确区分。
2. 短入口只保留结论、必要边界和路由；批次状态、画板节点、实验样本和长命令回到已有专题或周志。
3. 行数只是自动检查的一部分；人工复核还检查总信息量、超长段落、重复状态与历史清单，避免把历史内容压成单行满足预算。
4. 优先维护既有文档，不为一次讨论新增平行报告或 gate-only 专题；结构化字段与有效链接优先于多处复制长叙述。
5. 保留仍有消费者的契约 ID 与证据路由。检查器的文案耦合若需解除，作为明确的后续代码任务评审，不通过删断言或写入失真结论使检查通过。
6. 文档正文默认中文；没有稳定中文对应的专业名词、代码标识、路径和必要原文引用保留英文。历史英文工程短语按入口文档、专题文档、任务卡和周志顺序逐批收口，详见[文档语言治理 v1](document-language-governance-v1.md)。外部项目使用项目名和在线仓库链接，不记录本机外部路径。
7. 真相源、阶段和治理口径变更完成后运行全量仓库检查；本轮验证记录写入周志。

# RadishMind 功能设计文档入口

更新时间：2026-06-15

## 文档目的

本目录用于承载长期可复用的产品能力与功能专题文档。后续推进应先明确“正在做哪个专题”，再判断它属于大方向、具体功能、页面 / surface、平台能力还是外部扩展，最后再拆实现任务、测试和必要门禁。

`docs/features/*.md` 现在只作为产品面大方向入口，不再承担所有专题粒度。具体功能和复杂页面应进入对应子目录；跨产品面的平台能力进入 `docs/platform/`；外部项目或 backend 接入进入 `docs/integrations/`。

## 专题分层

| 层级 | 默认位置 | 用途 | 何时新增 |
| --- | --- | --- | --- |
| 产品面大方向专题 | `docs/features/*.md` | 固定 User Workspace、Admin、Gateway、Workflow、Image Path 等长期产品面目标和停止线 | 产品面边界、阶段顺序或长期目标变化 |
| 功能专题 | `docs/features/<area>/<feature>.md` | 固定某个功能的目标用户、流程、数据边界、失败语义、验收方式 | 功能需要跨多个实现批次推进，或会影响用户真实工作流 |
| 页面 / Surface 专题 | `docs/features/<area>/<surface>.md` | 固定复杂页面的信息架构、状态、交互和跨功能消费关系 | 页面承担复杂交互、状态组织、保存 / 审查 / handoff 等真实流程 |
| 平台专题 | `docs/platform/*.md` | 固定跨产品面的 auth、store、repository、provider、deployment、dev-only write path 等平台边界 | 能力不属于单一产品页面，或影响多个产品面和服务边界 |
| 扩展 / 集成专题 | `docs/integrations/*.md` | 固定 RadishFlow、Radish、Radish OIDC、image backend 等外部接入前置条件和停止线 | 需要依赖外部项目、外部 backend、真实 provider 或上层挂载点 |
| 实现批次 / 高风险边界 | `docs/task-cards/*.md` | 固定具体实现批次、输入输出、前置条件、专项 gate 和验证记录 | 新增 API、schema、执行边界、生产声明、外部 provider 风险或高风险能力 |

## 当前口径

- 产品面大方向专题描述长期目标、现有能力、下一批方向和停止线。
- 功能专题描述一个可持续推进的产品能力，必须写清目标用户、核心流程、数据边界、当前实现、下一批开发和验收方式。
- 页面 / Surface 专题只在页面承担复杂状态组织或真实用户流程时新增；普通展示、文案和布局不单独建专题。
- 平台专题和扩展专题不是 task card；它们负责长期边界和准入条件，具体实现仍用 task card 承接。
- 功能设计文档不是禁止清单；它应明确本功能允许打开的实现范围，以及哪些能力需要作为后续独立目标推进。
- `docs/radishmind-current-focus.md` 只回答当前阶段与下一顺位，不再承载长功能细节。
- `docs/task-cards/` 只承载具体实现批次、前置条件或高风险边界，不再作为产品功能的默认主文档。
- `scripts/checks/fixtures/` 与专项 checker 只在协议、schema、执行边界、生产声明、外部 provider 风险或高风险能力变化时新增。
- 普通 UI、文案、布局、只读 evidence 组织和使用性整理优先复用现有聚合门禁、web build、consumer smoke 和仓库基线。

## 产品面大方向导航

| 功能文档 | 当前作用 | 下一步默认入口 |
| --- | --- | --- |
| [User Workspace](user-workspace.md) | 用户端 AI 应用、API key、用量、运行记录和审查入口 | 从只读工作区转向真实用户工作流前先更新 |
| [Admin Control Plane](admin-control-plane.md) | 租户、权限、provider/profile、quota、secret、审计和部署证据 | 进入真实管理端、OIDC 或数据库前先更新 |
| [Model Gateway / API Distribution](model-gateway-api-distribution.md) | northbound API、provider/profile route、key/quota、trace 和审计 | 进入真实 API 分发、quota 或 billing 前先更新 |
| [Workflow / Agent Runtime](workflow-agent-runtime.md) | workflow draft、validation、execution plan、readiness、review 和 future executor | 进入 builder persistence、executor 或 confirmation 前先更新 |
| [Image Generation / Artifact Return](image-generation-artifact-return.md) | image intent、artifact metadata、response merge 和 future backend adapter | 进入 store / reader / public URL / backend adapter 前先更新 |

## 细专题导航

| 专题 | 类型 | 当前用途 |
| --- | --- | --- |
| [Workflow 细专题入口](workflow/README.md) | 功能专题目录 | 承接 workflow 具体功能、页面 / surface 和实现专题 |
| [Saved Workflow Draft v1](workflow/saved-workflow-draft-v1.md) | 功能专题 | 固定草案保存、读取、校验、版本冲突和失败语义 |
| [Workflow Draft Designer Surface](workflow/draft-designer-surface.md) | 页面 / Surface 专题 | 固定 draft designer 的 sample / unsaved / saved 状态和后续 consumer 接线边界 |
| [Dev-only Saved Draft Consumer](workflow/dev-only-saved-draft-consumer.md) | 实现专题 | 固定下一批 dev-only HTTP route + web consumer 的准入、验收和停止线 |
| [Saved Workflow Draft Durable Store Preconditions v1](workflow/saved-workflow-draft-durable-store-preconditions-v1.md) | 前置设计专题 | 固定 saved draft durable store 迁移前的 scope、owner / workspace、version conflict、no sample fallback 和 store 切换停止线 |
| [Saved Workflow Draft Repository Contract Preconditions v1](workflow/saved-workflow-draft-repository-contract-preconditions-v1.md) | 前置设计专题 | 固定 future saved draft repository contract 的 actor context、operation matrix、failure 和 projection 边界 |
| [Saved Workflow Draft Schema / Migration Preconditions v1](workflow/saved-workflow-draft-schema-migration-preconditions-v1.md) | 前置设计专题 | 固定 future saved draft durable store 的 logical schema、index strategy、migration gate 和 artifact guard |
| [Saved Workflow Draft Auth Context Preconditions v1](workflow/saved-workflow-draft-auth-context-preconditions-v1.md) | 前置设计专题 | 固定 future saved draft repository actor context 的身份来源、membership、owner、scope 和 audit 边界 |
| [Saved Workflow Draft Store Selector Enablement Preconditions v1](workflow/saved-workflow-draft-store-selector-enablement-preconditions-v1.md) | 前置设计专题 | 固定 future saved draft store mode、selector gate、failure、no fallback 和 artifact guard |
| [平台专题入口](../platform/README.md) | 平台专题目录 | 承接 auth、store、repository、provider、deployment 等跨产品面能力 |
| [扩展 / 集成专题入口](../integrations/README.md) | 扩展专题目录 | 承接外部项目、外部 backend 和真实接入前置条件 |

## 使用方式

1. 先判断本次推进属于产品面、具体功能、页面 / surface、平台能力还是外部扩展。
2. 大方向只更新 `docs/features/*.md`；具体功能和复杂页面优先进入对应子目录，例如 `docs/features/workflow/`。
3. 平台横切能力进入 `docs/platform/`；外部项目或 backend 接入进入 `docs/integrations/`。
4. 只有当实现批次需要固定输入输出、协议、前置条件或停止线时，才新增或更新 task card。
5. 只有当实现批次改变可复验契约或高风险能力时，才新增 fixture / checker。
6. 完成实现后，回写对应专题文档和当前焦点；历史验证细节写入周志或 run record。

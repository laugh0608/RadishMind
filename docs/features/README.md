# RadishMind 功能设计文档入口

更新时间：2026-06-14

## 文档目的

本目录用于承载长期可复用的功能设计与开发文档。后续推进应先明确“正在做哪个功能或开发目标”，再按该功能文档拆实现任务、测试和必要门禁。

## 当前口径

- 功能设计文档是产品能力的长期载体，描述目标用户、核心流程、数据边界、实现状态、下一批开发和停止线。
- 功能设计文档不是禁止清单；它应明确本功能允许打开的实现范围，以及哪些能力需要作为后续独立目标推进。
- `docs/radishmind-current-focus.md` 只回答当前阶段与下一顺位，不再承载长功能细节。
- `docs/task-cards/` 只承载具体实现批次、前置条件或高风险边界，不再作为产品功能的默认主文档。
- `scripts/checks/fixtures/` 与专项 checker 只在协议、schema、执行边界、生产声明、外部 provider 风险或高风险能力变化时新增。
- 普通 UI、文案、布局、只读 evidence 组织和使用性整理优先复用现有聚合门禁、web build、consumer smoke 和仓库基线。

## 功能文档导航

| 功能文档 | 当前作用 | 下一步默认入口 |
| --- | --- | --- |
| [User Workspace](user-workspace.md) | 用户端 AI 应用、API key、用量、运行记录和审查入口 | 从只读工作区转向真实用户工作流前先更新 |
| [Admin Control Plane](admin-control-plane.md) | 租户、权限、provider/profile、quota、secret、审计和部署证据 | 进入真实管理端、OIDC 或数据库前先更新 |
| [Model Gateway / API Distribution](model-gateway-api-distribution.md) | northbound API、provider/profile route、key/quota、trace 和审计 | 进入真实 API 分发、quota 或 billing 前先更新 |
| [Workflow / Agent Runtime](workflow-agent-runtime.md) | workflow draft、validation、execution plan、readiness、review 和 future executor | 进入 builder persistence、executor 或 confirmation 前先更新 |
| [Image Generation / Artifact Return](image-generation-artifact-return.md) | image intent、artifact metadata、response merge 和 future backend adapter | 进入 store / reader / public URL / backend adapter 前先更新 |

## 使用方式

1. 新功能或大目标先新增或更新对应 `docs/features/*.md`。
2. 功能文档明确目标、非目标、当前实现、下一批开发和验收方式。
3. 只有当实现批次需要固定输入输出、协议、前置条件或停止线时，才新增 task card。
4. 只有当实现批次改变可复验契约或高风险能力时，才新增 fixture / checker。
5. 完成实现后，回写功能文档和当前焦点；历史验证细节写入周志或 run record。

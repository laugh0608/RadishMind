# Workflow 细专题入口

更新时间：2026-06-15

## 文档目的

本目录承接 `Workflow / Agent Runtime` 的细粒度专题。上层大方向仍在 [Workflow / Agent Runtime 设计与开发文档](../workflow-agent-runtime.md)，本目录只放需要长期跟踪的具体功能、复杂页面 / surface 和实现专题。

普通只读展示、文案整理和布局调整不在本目录单独建文档，默认复用现有 workflow 聚合门禁、web build 和仓库基线。

## 当前专题

| 专题 | 类型 | 状态 | 作用 |
| --- | --- | --- | --- |
| [Saved Workflow Draft v1](saved-workflow-draft-v1.md) | 功能专题 | `domain_service_implemented` | 固定草案保存、读取、校验、版本冲突、失败语义和 saved / sample 区分 |
| [Workflow Draft Designer Surface](draft-designer-surface.md) | 页面 / Surface 专题 | `offline_surface_defined` | 固定 draft designer 的页面状态、数据来源和后续 saved draft 接线边界 |
| [Dev-only Saved Draft Consumer](dev-only-saved-draft-consumer.md) | 实现专题 | `implemented` | 固定 dev-only HTTP route + web consumer 的准入、验收和停止线 |

## 选题规则

- 功能专题：当一个 workflow 能力需要跨多批实现推进，且涉及数据边界、失败语义或用户真实工作流时创建。
- 页面 / Surface 专题：当页面承担复杂状态组织、跨功能消费、保存 / 审查 / handoff 流程时创建。
- 实现专题：当下一批实现需要在功能和任务卡之间先固定方案、准入和验收时创建。
- Task card：当实现批次开始落代码、route、schema、checker 或高风险 gate 时更新或新增。

## 当前下一步

`Saved Workflow Draft v1` 的 dev-only consumer integration 已实现。下一步不直接进入 executor / confirmation / writeback / replay，而是根据真实使用反馈选择是否补更细的 conflict UI、consumer smoke 或 route contract 固化；若新增专项 gate，应先更新对应 task card。

## 停止线

- 不实现 workflow executor、node executor、tool executor、agent loop、publish、run、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把 saved draft、validation summary、risk summary 或 readiness summary 解释为 publish ready、run ready 或 production ready。
- 不接真实数据库、repository adapter、schema migration、store selector、Radish OIDC、token validation、API key lifecycle、quota enforcement、billing 或 public production API。

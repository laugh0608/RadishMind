# Workflow 细专题入口

更新时间：2026-06-15

## 文档目的

本目录承接 `Workflow / Agent Runtime` 的细粒度专题。上层大方向仍在 [Workflow / Agent Runtime 设计与开发文档](../workflow-agent-runtime.md)，本目录只放需要长期跟踪的具体功能、复杂页面 / surface 和实现专题。

普通只读展示、文案整理和布局调整不在本目录单独建文档，默认复用现有 workflow 聚合门禁、web build 和仓库基线。

## 当前专题

| 专题 | 类型 | 状态 | 作用 |
| --- | --- | --- | --- |
| [Saved Workflow Draft v1](saved-workflow-draft-v1.md) | 功能专题 | `dev_consumer_stabilized` | 固定草案保存、读取、校验、版本冲突、失败语义和 saved / sample 区分 |
| [Workflow Draft Designer Surface](draft-designer-surface.md) | 页面 / Surface 专题 | `local_editing_entry_implemented` | 固定 draft designer 的页面状态、数据来源和后续 saved draft 接线边界 |
| [Workflow Draft Editing Entry v1](draft-editing-entry-v1.md) | 功能 / 页面专题 | `implemented` | 固定草案名称、说明、节点名称和边条件摘要的受控本地编辑入口 |
| [User Workspace Draft Creation v1](user-workspace-draft-creation-v1.md) | 功能 / 页面专题 | `implemented` | 固定从 Workspace Home / workflow definitions 创建本地草案并进入 Draft Designer 的入口 |
| [Saved Workflow Draft Durable Store Preconditions v1](saved-workflow-draft-durable-store-preconditions-v1.md) | 前置设计专题 | `draft_durable_store_preconditions_defined` | 固定 durable store 迁移前的 draft scope、owner / workspace、版本冲突、no sample fallback 和 store 切换停止线 |
| [Dev-only Saved Draft Consumer](dev-only-saved-draft-consumer.md) | 实现专题 | `implemented` | 固定 dev-only HTTP route + web consumer 的准入、验收和停止线 |

## 选题规则

- 功能专题：当一个 workflow 能力需要跨多批实现推进，且涉及数据边界、失败语义或用户真实工作流时创建。
- 页面 / Surface 专题：当页面承担复杂状态组织、跨功能消费、保存 / 审查 / handoff 流程时创建。
- 实现专题：当下一批实现需要在功能和任务卡之间先固定方案、准入和验收时创建。
- Task card：当实现批次开始落代码、route、schema、checker 或高风险 gate 时更新或新增。

## 当前下一步

`Saved Workflow Draft v1` 的 dev-only consumer integration 已实现，并已补 route contract、consumer smoke 和 `version_conflict` 状态；`Workflow Draft Editing Entry v1` 已补受控本地编辑入口，并让 validate / save / read 使用当前本地草案；`User Workspace Draft Creation v1` 已补 Workspace Home / workflow definitions 创建本地草案入口；`Saved Workflow Draft Durable Store Preconditions v1` 已固定 durable store 迁移前置设计和 checker。下一步不直接进入 executor / confirmation / writeback / replay；若继续推进 durable store，只能先补 repository contract / schema migration / auth contract 等独立准入，不得绕过 dev store 与未来 repository adapter 的切换停止线。

## 停止线

- 不实现 workflow executor、node executor、tool executor、agent loop、publish、run、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把 saved draft、validation summary、risk summary 或 readiness summary 解释为 publish ready、run ready 或 production ready。
- 不接真实数据库、repository adapter、schema migration、store selector、Radish OIDC、token validation、API key lifecycle、quota enforcement、billing 或 public production API。

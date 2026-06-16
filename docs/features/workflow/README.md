# Workflow 细专题入口

更新时间：2026-06-16

## 文档目的

本目录承接 `Workflow / Agent Runtime` 的细粒度专题。上层大方向仍在 [Workflow / Agent Runtime 设计与开发文档](../workflow-agent-runtime.md)，本目录只放需要长期跟踪的具体功能、复杂页面 / surface 和实现专题。

普通只读展示、文案整理和布局调整不在本目录单独建文档，默认复用现有 workflow 聚合门禁、web build 和仓库基线。

## 当前专题

| 专题 | 类型 | 状态 | 作用 |
| --- | --- | --- | --- |
| [Saved Workflow Draft v1](saved-workflow-draft-v1.md) | 功能专题 | `dev_consumer_stabilized` | 固定草案保存、读取、校验、版本冲突、失败语义和 saved / sample 区分 |
| [Workflow Draft Designer Surface](draft-designer-surface.md) | 页面 / Surface 专题 | `structure_editing_model_v2_implemented` | 固定 draft designer 的页面状态、数据来源、本地结构编辑和后续 saved draft 接线边界 |
| [Workflow Draft Editing Entry v1](draft-editing-entry-v1.md) | 功能 / 页面专题 | `implemented` | 固定草案名称、说明、节点名称和边条件摘要的受控本地编辑入口 |
| [User Workspace Draft Creation v1](user-workspace-draft-creation-v1.md) | 功能 / 页面专题 | `implemented` | 固定从 Workspace Home / workflow definitions 创建本地草案并进入 Draft Designer 的入口 |
| [User Workspace Saved Draft List v1](user-workspace-saved-draft-list-v1.md) | 功能 / 页面专题 | `implemented` | 固定 Workspace Home 中已保存 dev draft 的 sanitized summary、empty / failure state 和恢复入口 |
| [Workflow Draft Designer Editing Model v2](draft-designer-editing-model-v2.md) | 功能 / 页面专题 | `workflow_draft_designer_editing_model_v2_implemented` | 固定 Draft Designer 本地节点新增、删除保护、重排、边重建和 active draft 下游预览 |
| [Workflow Draft Node Attribute Editing Model v1](draft-node-attribute-editing-model-v1.md) | 功能 / 页面专题 | `workflow_draft_node_attribute_editing_model_v1_implemented` | 固定 Draft Designer 节点属性编辑、保存映射、恢复映射和下游 validation / plan 消费 |
| [Workflow Review Handoff Active Draft v1](review-handoff-active-draft-v1.md) | 页面 / Surface 专题 | `workflow_review_handoff_active_draft_v1_implemented` | 固定 Review Handoff 对 active draft validation / plan / readiness 的可交接审查记录 |
| [Saved Workflow Draft Durable Store Preconditions v1](saved-workflow-draft-durable-store-preconditions-v1.md) | 前置设计专题 | `draft_durable_store_preconditions_defined` | 固定 durable store 迁移前的 draft scope、owner / workspace、版本冲突、no sample fallback 和 store 切换停止线 |
| [Saved Workflow Draft Repository Contract Preconditions v1](saved-workflow-draft-repository-contract-preconditions-v1.md) | 前置设计专题 | `draft_repository_contract_preconditions_defined` | 固定未来 repository contract 的 actor context、operation matrix、request / result、failure 和 projection 边界 |
| [Saved Workflow Draft Schema / Migration Preconditions v1](saved-workflow-draft-schema-migration-preconditions-v1.md) | 前置设计专题 | `draft_schema_migration_preconditions_defined` | 固定 future durable store 的 logical schema、index strategy、migration gate、failure 和 artifact guard |
| [Saved Workflow Draft Auth Context Preconditions v1](saved-workflow-draft-auth-context-preconditions-v1.md) | 前置设计专题 | `draft_auth_context_preconditions_defined` | 固定 future repository actor context 的身份来源、workspace membership、owner policy、scope grants 和 audit 边界 |
| [Saved Workflow Draft Store Selector Enablement Preconditions v1](saved-workflow-draft-store-selector-enablement-preconditions-v1.md) | 前置设计专题 | `draft_store_selector_enablement_preconditions_defined` | 固定 future store mode、selector gate、failure、no fallback 和 artifact guard |
| [Saved Workflow Draft Schema Artifact Evidence v1](saved-workflow-draft-schema-artifact-evidence-v1.md) | 前置证据专题 | `draft_schema_artifact_evidence_defined` | 固定 future schema artifact manifest、DDL review、rollback、migration smoke 和 artifact guard |
| [Saved Workflow Draft Store Selector Smoke Readiness v1](saved-workflow-draft-store-selector-smoke-readiness-v1.md) | 前置证据专题 | `draft_store_selector_smoke_readiness_defined` | 固定 future selector smoke 的 mode matrix、operation matrix、failure、no fallback 和 artifact guard |
| [Dev-only Saved Draft Consumer](dev-only-saved-draft-consumer.md) | 实现专题 | `implemented` | 固定 dev-only HTTP route + web consumer 的准入、验收和停止线 |

## 选题规则

- 功能专题：当一个 workflow 能力需要跨多批实现推进，且涉及数据边界、失败语义或用户真实工作流时创建。
- 页面 / Surface 专题：当页面承担复杂状态组织、跨功能消费、保存 / 审查 / handoff 流程时创建。
- 实现专题：当下一批实现需要在功能和任务卡之间先固定方案、准入和验收时创建。
- Task card：当实现批次开始落代码、route、schema、checker 或高风险 gate 时更新或新增。

## 当前下一步

`Saved Workflow Draft v1` 的 dev-only consumer integration 已实现，并已补 route contract、consumer smoke 和 `version_conflict` 状态；`Workflow Draft Editing Entry v1` 已补受控本地编辑入口，并让 validate / save / read 使用当前本地草案；`User Workspace Draft Creation v1` 已补 Workspace Home / workflow definitions 创建本地草案入口；`User Workspace Saved Draft List v1` 已补 Workspace Home 的 saved dev draft summary、empty / failure state、refresh 和恢复入口；`Workflow Draft Designer Editing Model v2` 已补本地节点新增、删除保护、重排、边重建和 active draft validation / plan / readiness 派生；`Workflow Draft Node Attribute Editing Model v1` 已补节点属性编辑、dev-only 保存映射、恢复映射和 validation / plan 消费；`Workflow Review Handoff Active Draft v1` 已把 active draft validation inspector、execution plan preview 和 runtime readiness inspector 汇总成可交接审查记录；`Saved Workflow Draft Durable Store Preconditions v1`、`Saved Workflow Draft Repository Contract Preconditions v1`、`Saved Workflow Draft Schema / Migration Preconditions v1`、`Saved Workflow Draft Auth Context Preconditions v1`、`Saved Workflow Draft Store Selector Enablement Preconditions v1`、`Saved Workflow Draft Schema Artifact Evidence v1` 和 `Saved Workflow Draft Store Selector Smoke Readiness v1` 已固定 durable store 迁移前置设计、future repository contract preconditions、future schema / migration preconditions、auth context preconditions、store selector enablement preconditions、schema artifact evidence 和 selector smoke readiness。下一步若继续推进 durable store，只能先补 repository contract smoke 或 repository adapter implementation plan 等独立准入；若继续用户审查体验，应选择新的明确用户审查增强点，不得绕过 dev store 与未来 repository adapter 的切换停止线，也不直接进入 executor / confirmation / writeback / replay。

## 停止线

- 不实现 workflow executor、node executor、tool executor、agent loop、publish、run、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把 saved draft、validation summary、risk summary 或 readiness summary 解释为 publish ready、run ready 或 production ready。
- 不接真实数据库、repository adapter、schema migration、store selector、Radish OIDC、token validation、API key lifecycle、quota enforcement、billing 或 public production API。

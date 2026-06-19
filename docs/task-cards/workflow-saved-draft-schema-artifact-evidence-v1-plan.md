# Workflow Saved Draft Schema Artifact Evidence v1 任务卡

更新时间：2026-06-15

## 任务标识

- 切片：`workflow-saved-draft-schema-artifact-evidence-v1`
- 轨道：`Workflow / Agent Runtime`
- 状态：`draft_schema_artifact_evidence_defined`

## 目标

在 schema / migration preconditions、auth context preconditions 和 store selector enablement preconditions 之后，固定 future saved workflow draft schema artifact manifest / DDL review evidence 的证据链、字段映射、索引映射、migration smoke、failure mapping、no fallback、no side effects 和 implementation artifact guard。

本任务卡只定义 schema artifact evidence，不创建 migration root、manifest 文件、DDL review 文件、SQL migration、schema version table、migration runner、真实数据库、repository interface、repository adapter、store selector、Radish OIDC middleware、production API、saved draft list、publish、run、executor、confirmation、writeback 或 replay。

## 输入事实源

- [Saved Workflow Draft v1 功能专题](../features/workflow/saved-workflow-draft-v1.md)
- [Saved Workflow Draft Repository Contract Preconditions v1 专题](../features/workflow/saved-workflow-draft-repository-contract-preconditions-v1.md)
- [Saved Workflow Draft Schema / Migration Preconditions v1 专题](../features/workflow/saved-workflow-draft-schema-migration-preconditions-v1.md)
- [Saved Workflow Draft Auth Context Preconditions v1 专题](../features/workflow/saved-workflow-draft-auth-context-preconditions-v1.md)
- [Saved Workflow Draft Store Selector Enablement Preconditions v1 专题](../features/workflow/saved-workflow-draft-store-selector-enablement-preconditions-v1.md)
- [Saved Workflow Draft Schema Artifact Evidence v1 专题](../features/workflow/saved-workflow-draft-schema-artifact-evidence-v1.md)

## 本轮交付

- 新增 schema artifact evidence 细专题，固定 evidence-only 状态。
- 新增 `workflow-saved-draft-schema-artifact-evidence-v1` fixture / checker。
- checker 接入 fast baseline，校验 dependency、manifest contract、logical entity mapping、index mapping、DDL review gate、migration smoke evidence、failure mapping、no fallback、no side effects、文档引用和 forbidden implementation artifact。
- 同步更新 workflow 入口、Saved Draft 主专题、schema / selector preconditions、User Workspace、当前焦点、任务卡索引、脚本说明和周志。

## Evidence Contract

future schema artifact materialization 前必须具备：

- `manifest.json`：artifact id、store schema version、logical entity、field mapping、index mapping、migration id、DDL review ref、rollback ref、migration smoke ref、failure mapping 和 no side effect counters。
- `ddl-review.md`：人工审查、manual apply command、backup requirement、migration lock、schema version table、rollback / forward-only 策略和 destructive change gate。
- rollback evidence：rollback command 或 forward-only exception、备份要求、失败迁移恢复和 lock release。
- migration smoke：schema version table、tenant / workspace / application predicate、owner list projection、version conflict predicate、migration failure mapping、no sample fallback 和 no side effects。

## 验收口径

- `workflow-saved-draft-schema-artifact-evidence-v1` checker 通过。
- `./scripts/check-repo.sh --fast` 通过。
- 文档和 fixture 均保持 evidence-only，不声明 schema artifact ready、DDL ready、migration ready、database ready、repository adapter ready 或 production ready。

## 停止线

- 不创建 migration root、manifest 文件、DDL review 文件、SQL migration、schema version table、migration runner、数据库连接、repository interface、repository adapter、store selector、Radish OIDC middleware、token validation、public production API、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把本任务卡、fixture 或 checker 解释为 durable persistence、saved draft list、schema migration、store selector、repository adapter、saved draft list、publish、run 或 production readiness。

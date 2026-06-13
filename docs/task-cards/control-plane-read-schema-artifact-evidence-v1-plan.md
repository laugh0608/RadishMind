# Control Plane Read Schema Artifact Evidence v1 任务卡

## 任务标识

- 切片：`control-plane-read-schema-artifact-evidence-v1`
- 轨道：`Control Plane / User Workspace / Workflow v1`
- 状态：`schema_artifact_evidence_defined`

## 目标

在未来创建 durable read store schema artifact、migration manifest、SQL、migration runner 或 repository adapter 之前，先把 schema 证据链固定为可检查 artifact evidence contract。

该切片补齐：

- DDL review evidence 的字段、审查要求和人工复核边界。
- rollback fixture evidence 的回滚、备份、锁释放和失败恢复要求。
- schema version evidence 的目标版本、版本表、缺失 / mismatch failure mapping。
- tenant index evidence 的 `tenant_ref` leading index、跨租户拒绝和 fail-closed 要求。
- read-only role evidence 的只读权限、禁止写入、禁止 migration 和 secret reference-only 要求。
- 七条 read route 到未来 schema artifact / projection 的映射关系。

本切片只创建任务卡、fixture 和 checker，不创建真实 schema artifact 文件。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [能力矩阵](../radishmind-capability-matrix.md)
- [Control Plane Read-Side 契约](../contracts/control-plane-read-side.md)
- [`Control Plane Read Schema Artifact Manifest Readiness` v1 计划](control-plane-read-schema-artifact-manifest-readiness-v1-plan.md)
- [`Control Plane Read Implementation Trigger Review` v1 计划](control-plane-read-implementation-trigger-review-v1-plan.md)
- [`Product Surface Readiness / Implementation Trigger Recheck` v1 计划](product-surface-readiness-implementation-trigger-recheck-v1-plan.md)

## 验收口径

- `scripts/checks/fixtures/control-plane-read-schema-artifact-evidence-v1.json` 固定 schema artifact evidence boundary、evidence gate matrix、DDL review、rollback fixture、schema version、tenant index、read-only role、七条 route mapping、failure mapping、no fake fallback 和 no side effects。
- `scripts/checks/control_plane/check-control-plane-read-schema-artifact-evidence-v1.py` 已进入 `./scripts/check-repo.sh --fast`。
- checker 必须跨读既有 schema artifact manifest readiness、implementation trigger review 和 product surface recheck，确认新证据链不会把 implementation trigger 改写为 satisfied。
- current focus、能力矩阵、integration contracts、read-side contract、task card index、scripts README、platform README 和 W24 周志同步该切片。

## 非目标

- 不新增同层只读 UI 面板。
- 不启动开发服务器，不做真实联调。
- 不创建 migration 目录、manifest、SQL、DDL review artifact、rollback fixture、schema smoke artifact 或 database fixture。
- 不实现 migration runner、store selector、repository interface、repository adapter、database query 或 production API consumer。
- 不接真实数据库、Radish OIDC、repository adapter、production gateway、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

## 停止线

- 不能把 schema evidence contract 写成 schema artifact manifest ready。
- 不能把 DDL review / rollback / schema version / tenant index / read-only role evidence 写成已经执行过的 smoke。
- 不能把七条 route mapping 写成 runtime store selector、repository adapter 或 SQL projection 已实现。
- 不能把 fake-store-backed dev path 写成 durable read store。

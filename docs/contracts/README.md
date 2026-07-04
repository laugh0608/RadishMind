# RadishMind 契约专题目录

更新时间：2026-06-28

本目录承载 `docs/radishmind-integration-contracts.md` 拆出的稳定专题。入口文档只保留当前结论、索引和停止线；需要修改字段、任务上下文或长样例时，优先修改本目录下的对应专题页。

## 阅读入口

- [服务/API 接入契约](service-api.md)
- [Control Plane Read-Side 契约](control-plane-read-side.md)
- [Platform Overview UI View 契约](platform-overview-ui-view.md)
- [Platform Local Smoke UI View 契约](platform-local-smoke-ui-view.md)
- [会话记录契约](session.md)
- [工具框架契约](tooling.md)
- [Session / Tooling UI View 契约](session-tooling-ui-view.md)
- [Production Secret Reference 契约](production-secret-reference.md)
- [Radish OIDC Token Validation 契约](radish-oidc-token-validation.md)
- [Production Secret Audit Event 契约](production-secret-audit-event.md)
- [Production Secret Audit Storage Adapter Metadata Contract 契约](production-secret-audit-storage-adapter-metadata-contract.md)
- [Production Secret Audit Storage Adapter Table Schema 契约](production-secret-audit-storage-adapter-table-schema.md)
- [训练 / 蒸馏样本契约](training-samples.md)
- [图片生成契约](image-generation.md)
- [输入与项目上下文契约](input-context.md)
- [RadishFlow 上下文契约](radishflow-context.md)
- [Radish 上下文契约](radish-context.md)
- [RadishCatalyst 上下文预留契约](radishcatalyst-context.md)
- [输出与候选动作契约](output-actions.md)

## 维护规则

- 本目录文档不替代 `contracts/` 下的 schema 真相源。
- Control Plane Read-Side 的 TypeScript consumer contract 由 `contracts/typescript/control-plane-read-api.ts` 与 `control-plane-read-consumer-contract-v1` checker 固定；正式 UI 边界由 `control-plane-read-formal-ui-boundary-v1` checker 固定；正式 UI 实现 readiness 由 `control-plane-read-formal-ui-implementation-readiness-v1` checker 固定；`apps/radishmind-web/` 的 shared shell、七个只读页面、formal UI readiness close、dev-only live consumer、RadishFlow / Radish Docs 产品样例一致性、User Workspace Home、Workflow Review Handoff、`workflowWorkspaceContext` 派生一致性、auth/store transition preconditions、repository contract smoke、repository implementation readiness、store selection readiness、schema migration readiness、repository contract types implementation、静态 contract smoke runner、repository interface readiness、`ControlPlaneReadRepository` interface + fake-store bridge、adapter implementation readiness refresh、selector enablement preconditions、schema migration implementation preconditions、repository adapter implementation plan、schema artifact manifest readiness 和 store selector smoke readiness 由对应 `control-plane-read-*-v1` / `workflow-*-v1` checker 或聚合 gate 固定。专题页只解释语义、停止线和推进顺序。
- Image Generation / Artifact Return 的 schema、runtime mapper、response consumer 和 `coerce_response_document` metadata-only hook 统一解释在 [图片生成契约](image-generation.md)；该专题只说明 artifact metadata 如何进入现有 `CopilotResponse.citations`，不声明 artifact store、binary reader、public URL 或真实 backend ready。
- Radish OIDC Token Validation 的 schema、positive / negative fixtures 和 artifact checker 统一解释在 [Radish OIDC Token Validation 契约](radish-oidc-token-validation.md)；该专题只说明 verified token context 的脱敏投影、forbidden raw-material 字段、消费规则和验证方式，不声明 OIDC middleware、token validator、membership adapter、repository mode 或 production API ready。
- Production Secret Audit Event 的 schema、positive / negative fixtures 和 artifact checker 统一解释在 [Production Secret Audit Event 契约](production-secret-audit-event.md)；该专题只说明 future audit writer 的 metadata-only event 输入，不声明 audit writer runtime、audit store runtime、delivery、idempotency、durable backend、repository mode 或 production API ready。
- Production Secret Audit Storage Adapter Metadata Contract 的 contract artifact、positive / negative fixtures、writer compatibility smoke 和 artifact checker 统一解释在 [Production Secret Audit Storage Adapter Metadata Contract 契约](production-secret-audit-storage-adapter-metadata-contract.md)；该专题只说明 future storage adapter 的 metadata-only input / result envelope、record identity、failure taxonomy 和 writer output handoff。后续 review 已静态选择 `managed_database_append_only_table` product class，并定义 database policy 与 logical table schema boundary，但仍不声明具体 backend product / vendor、storage adapter runtime、DB provider、audit store runtime、repository mode 或 production API ready。
- Production Secret Audit Storage Adapter Table Schema 的 metadata-only logical schema artifact、positive / negative fixtures、metadata contract compatibility smoke 和 no secret material scan 统一解释在 [Production Secret Audit Storage Adapter Table Schema 契约](production-secret-audit-storage-adapter-table-schema.md)；该专题状态为 `audit_store_storage_adapter_table_schema_artifact_materialized`，下一依赖为 `storage_adapter_offline_adapter_smoke_strategy_readiness`，只说明 future storage adapter 的 logical field groups，不声明 SQL、DDL、物理表名、列名、列类型、schema marker runtime、migration runner、storage adapter runtime、DB provider、audit store runtime、repository mode 或 production API ready。
- `P2 Session & Tooling Foundation` 的晋级口径同时由 `scripts/checks/fixtures/session-tooling-promotion-gates.json` 固定；修改 session/tooling promotion gate 时，应同步更新对应专题页和该 fixture。
- P2 负向门禁消费关系由 `scripts/checks/fixtures/session-tooling-negative-consumption-summary.json` 固定；新增 denied query、promotion gate 或对应消费者时，应同步更新该 summary。
- Checkpoint read route smoke 覆盖关系由 `scripts/checks/fixtures/session-recovery-checkpoint-route-smoke-coverage-summary.json` 固定；修改 route 正向/负向 smoke 时，应同步更新该 summary。
- 长示例、批次流水和运行记录继续进入 fixture、manifest、summary、run record 或 task card 附件。
- 单个专题接近 `500` 行时，应优先继续按稳定职责拆分，而不是添加 `markdown-size-allow:`。

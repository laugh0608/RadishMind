# Saved Workflow Draft Store Selector Enablement Preconditions v1 专题

更新时间：2026-06-15

## 专题定位

`Saved Workflow Draft Store Selector Enablement Preconditions v1` 承接 [Saved Workflow Draft Auth Context Preconditions v1](saved-workflow-draft-auth-context-preconditions-v1.md)，用于在实现 store selector、repository adapter、真实数据库或 production API 前固定 saved workflow draft store mode、切换 gate、failure mapping、no dev fallback 和 artifact guard。

本专题只定义 store selector enablement preconditions，不创建正式 config entry、selector 函数、selector 类型、repository interface、repository adapter、SQL migration、真实数据库、Radish OIDC middleware、production API、saved draft list、publish、run、executor、confirmation、writeback 或 replay。

状态：`draft_store_selector_enablement_preconditions_defined`

## 当前输入事实

- `Saved Workflow Draft v1` 已有 memory dev store、dev-only route + web consumer、save / read / validate、版本冲突、no sample fallback 和 no side effects tests。
- `Saved Workflow Draft Repository Contract Preconditions v1` 已固定 future repository operation matrix。
- `Saved Workflow Draft Schema / Migration Preconditions v1` 已固定 future durable store logical schema、index strategy、migration gate、failure mapping 和 artifact guard。
- `Saved Workflow Draft Auth Context Preconditions v1` 已固定 future repository actor context 的身份来源、workspace membership、owner policy、scope grants 和 audit / sanitization policy。
- `Saved Workflow Draft Schema Artifact Evidence v1` 已固定 `draft_schema_artifact_evidence_defined`，覆盖 future schema artifact manifest、DDL review、rollback evidence、migration smoke、failure mapping 和 artifact guard。
- 当前仍没有 store selector、repository interface、repository adapter、schema artifact manifest、database schema、SQL migration、Radish OIDC middleware、token validation 或 production API consumer。

## Store Mode Contract

future selector 只能识别以下 store mode：

| mode | 当前状态 | 行为 |
| --- | --- | --- |
| `memory_dev` | current default | 继续使用 platform memory dev store，仅用于 dev-only saved draft route |
| `repository_disabled` | reserved disabled | 明确禁用 future repository store，返回 `repository_store_disabled` |
| `repository` | future reserved | 在 adapter / schema / auth / smoke gate 满足前返回 `repository_store_disabled` |
| `unknown` | fail closed | 返回 `invalid_draft_store_mode` |

默认仍是 `memory_dev`。`repository_disabled` 和 `repository` 都不得 fallback 到 memory dev store，也不得把 sample / fixture 包装成 durable saved record。

## Selector Gate

future `SelectWorkflowSavedDraftStore` 或等价 selector 进入实现前，必须满足：

1. 正式 config key 已设计并审查，例如 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE`；本专题不创建该 config entry。
2. dev HTTP enablement 与 write enablement 仍由 `RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP` 和 `RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE` 单独控制。
3. repository adapter gate 未满足前，`repository` mode 必须 fail closed。
4. schema / migration gate 未满足前，`repository` mode 必须 fail closed。
5. auth context gate 未满足前，`repository` mode 必须 fail closed。
6. selector smoke 必须覆盖 unset / `memory_dev` / `repository_disabled` / `repository` / unknown mode。
7. 所有 reserved mode 都必须证明 no sample fallback、no memory dev fallback 和 no side effects。

## Failure Mapping

selector 相关 failure 必须保持 fail closed：

- `repository_store_disabled`
- `invalid_draft_store_mode`
- `draft_store_unavailable`
- `draft_store_contract_mismatch`
- `draft_auth_context_contract_mismatch`
- `draft_schema_migration_not_applied`
- `draft_store_schema_version_mismatch`

这些 failure 不得保存草案、不得返回草案主体、不得调用 provider/tool、不得创建 run record、不得提交 confirmation decision、不得写业务真相源。

## 后续准入

本专题之后仍不能直接创建 repository adapter。后续若继续 durable store 方向，应选择一个独立批次：

1. repository contract smoke / selector smoke readiness，覆盖 selector mode、schema artifact failure、failure mapping、no fallback 和 no side effects。
2. repository interface / adapter implementation plan，必须消费 schema、auth、selector、schema artifact evidence 和 smoke gate。
3. schema artifact materialization review，另行决定是否创建 manifest / DDL review / migration smoke artifact；进入该批前仍不得连接真实数据库。

## 验收方式

- 新增 task card 固定本专题为 precondition-only。
- 新增 fixture / checker 固定 store mode contract、selector gate、failure mapping、no fallback、no side effects、dev flag boundary 和 implementation artifact guard。
- checker 接入 `./scripts/check-repo.sh --fast`。
- 本批至少运行专项 checker 和 `./scripts/check-repo.sh --fast`。

## 停止线

- 不创建正式 config entry、selector 函数、selector 类型、repository interface、repository adapter、SQL migration、真实数据库、Radish OIDC middleware、token validation、production API consumer 或 saved draft list。
- 不实现 durable persistence、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把 store selector enablement preconditions 解释为 store selector ready、repository mode ready、repository adapter ready、database ready 或 production ready。

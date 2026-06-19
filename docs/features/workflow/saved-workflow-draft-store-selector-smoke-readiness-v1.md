# Saved Workflow Draft Store Selector Smoke Readiness v1 专题

更新时间：2026-06-15

## 专题定位

`Saved Workflow Draft Store Selector Smoke Readiness v1` 承接 [Saved Workflow Draft Store Selector Enablement Preconditions v1](saved-workflow-draft-store-selector-enablement-preconditions-v1.md) 和 [Saved Workflow Draft Schema Artifact Evidence v1](saved-workflow-draft-schema-artifact-evidence-v1.md)，用于在实现正式 store selector、config entry、repository adapter、数据库或 production API 前，固定 future selector smoke 的 mode matrix、operation matrix、failure mapping、no fallback 和 no side effects。

本专题只定义 selector smoke readiness，不创建正式 config entry、selector 函数、selector 类型、selector smoke fixture、selector smoke checker、Go selector test、repository interface、repository adapter、SQL migration、真实数据库、Radish OIDC middleware、production API、saved draft list、publish、run、executor、confirmation、writeback 或 replay。

状态：`draft_store_selector_smoke_readiness_defined`

后续 `Saved Workflow Draft Store Selector Implementation v1` 已在独立任务卡中完成，状态为 `draft_store_selector_smoke_implemented`。本专题仍保留 smoke readiness 当批未创建 selector smoke artifact 的历史事实。

## 当前输入事实

- `Saved Workflow Draft v1` 已有 memory dev store、dev-only route + web consumer、save / read / validate、版本冲突、no sample fallback 和 no side effects tests。
- `Saved Workflow Draft Repository Contract Preconditions v1` 已固定 future repository operation matrix：`SaveWorkflowDraftRecord`、`ReadWorkflowDraftRecord` 和 `ListWorkflowDraftRecords`。
- `Saved Workflow Draft Store Selector Enablement Preconditions v1` 已固定 `memory_dev` / `repository_disabled` / `repository` / unknown mode、selector gate、failure mapping、no fallback 和 dev flag boundary。
- `Saved Workflow Draft Schema Artifact Evidence v1` 已固定 schema artifact manifest、DDL review、rollback evidence、migration smoke、schema failure mapping 和 no side effects evidence。
- 当前已有正式 store config、store selector 和 selector smoke fixture；仍没有 repository interface、repository adapter、schema artifact manifest、SQL migration、Radish OIDC middleware、token validation 或 production API consumer。

## Selector Smoke Contract

future selector smoke 必须覆盖以下 mode：

| mode | 输入 | 预期 |
| --- | --- | --- |
| `memory_dev` | unset 或显式 `memory_dev` | 选择 platform memory dev store，仅用于 dev-only saved draft route |
| `repository_disabled` | `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository_disabled` | fail closed，返回 `repository_store_disabled` |
| `repository` | `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` | 在 adapter / schema / auth / smoke gate 满足前 fail closed，返回 `repository_store_disabled` |
| `unknown` | 任意未识别值 | fail closed，返回 `invalid_draft_store_mode` |

reserved mode 和 unknown mode 都不得 fallback 到 memory dev store、sample 或 fixture，也不得产生数据库写入、workflow execution、confirmation decision、business writeback 或 replay side effect。

## Operation Smoke Matrix

future selector smoke 必须覆盖以下 repository operation：

- `SaveWorkflowDraftRecord`
- `ReadWorkflowDraftRecord`
- `ListWorkflowDraftRecords`

每个 operation 都必须证明：

- unset / `memory_dev` 仍选择 dev-only memory store。
- `repository_disabled` 和 `repository` 在 gate 未满足前返回 `repository_store_disabled`。
- unknown mode 返回 `invalid_draft_store_mode`。
- schema artifact failure 进入对应 fail-closed mapping，不覆盖 `draft_version_conflict`、`draft_scope_denied` 或 `draft_store_unavailable`。
- denied / failed selector path 不调用 provider/tool，不创建 run record，不提交 confirmation decision，不写业务真相源。

## Failure Mapping

selector smoke readiness 必须保留以下 failure code：

- `repository_store_disabled`
- `invalid_draft_store_mode`
- `draft_store_unavailable`
- `draft_store_contract_mismatch`
- `draft_auth_context_contract_mismatch`
- `draft_schema_migration_not_applied`
- `draft_store_schema_version_mismatch`
- `draft_store_migration_unavailable`

## 后续准入

本专题之后的 repository contract smoke 已由 [Saved Workflow Draft Repository Contract Smoke v1](saved-workflow-draft-repository-contract-smoke-v1.md) 固定为 `draft_repository_contract_smoke_defined`。后续仍不能直接创建 repository adapter；若继续 durable store 方向，应选择一个独立批次：

1. repository contract smoke runner readiness / implementation，消费 selector smoke readiness、schema artifact evidence、auth context、repository operation matrix 和 repository contract smoke。
2. repository adapter implementation plan，必须消费 schema preconditions、schema artifact evidence、auth context、store selector readiness 和 repository contract smoke。
3. selector implementation entry review，另行决定是否创建 formal config、selector 函数、selector tests 和 selector smoke fixture；进入该批前仍不得连接真实数据库。

## 验收方式

- 新增 task card 固定本专题为 readiness-only。
- 新增 fixture / checker 固定 selector smoke boundary、mode matrix、operation matrix、schema artifact failure matrix、failure mapping、no fallback、no side effects 和 implementation artifact guard。
- checker 接入 `./scripts/check-repo.sh --fast`。
- 本批至少运行专项 checker 和 `./scripts/check-repo.sh --fast`。

## 停止线

- 不创建正式 config entry、selector 函数、selector 类型、selector smoke fixture、selector smoke checker、Go selector test、repository interface、repository adapter、migration root、SQL migration、真实数据库、Radish OIDC middleware、token validation、production API consumer 或 saved draft list。
- 不实现 durable persistence、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把 selector smoke readiness 解释为 selector smoke ready、store selector ready、repository mode ready、repository adapter ready、database ready 或 production ready。

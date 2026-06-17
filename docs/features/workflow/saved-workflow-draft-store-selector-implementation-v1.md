# Saved Workflow Draft Store Selector Implementation v1

更新时间：2026-06-17

## 专题定位

本专题承接 `Saved Workflow Draft Store Selector Implementation Entry Review v1` 之后的正式 selector implementation 批次。

它只打开 saved workflow draft store mode 的配置入口、`SelectWorkflowSavedDraftStore` 运行时选择器、selector 单元测试、HTTP fail-closed 覆盖和 selector smoke fixture/checker。它不实现 repository adapter、durable persistence、数据库 schema、SQL migration、OIDC/token validation、production API、publish/run/executor、confirmation、writeback 或 replay。

## 当前状态

- 状态：`draft_store_selector_smoke_implemented`。
- 正式配置入口已落地：配置文件键为 `workflow_saved_draft_store`，环境变量为 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE`，sanitized summary 字段为 `workflow_saved_draft_store_mode`，默认值为 `memory_dev`。
- `SelectWorkflowSavedDraftStore` 已在 `services/platform/internal/httpapi/workflow_saved_draft_store_selector.go` 中实现，并由 `newSavedWorkflowDraftStoreFromConfig` 接入 `NewServer`。
- selector tests 覆盖 `memory_dev`、`repository_disabled`、`repository` 和 unknown mode；HTTP route tests 覆盖 reserved / unknown mode 的 save / read / list fail-closed。
- 新增 `workflow-saved-draft-store-selector-smoke-v1.json` 与 `check-workflow-saved-draft-store-selector-smoke-v1.py`，并接入 fast baseline。
- 既有 readiness / review checker 已改为允许本批 selector implementation 固定产物存在，同时继续禁止 repository adapter、migration root、SQL、OIDC 和 production API artifact。

## Mode 语义

| mode | 结果 | 行为 |
| --- | --- | --- |
| `memory_dev` | memory dev store | 保持现有开发态 save / read / list 行为 |
| `repository_disabled` | disabled store | 返回 `repository_store_disabled`，不 fallback 到 memory dev / sample / fixture |
| `repository` | disabled store | 在 adapter / schema / auth / smoke gate 满足前返回 `repository_store_disabled` |
| unknown | disabled store | 返回 `invalid_draft_store_mode`，不 fallback |

## 失败语义

- `repository_store_disabled`：reserved repository store 被显式禁用或尚未满足 adapter / schema / auth / smoke gate。
- `invalid_draft_store_mode`：配置值不是允许的 store mode。
- `draft_store_unavailable`：非 selector failure 的存储错误仍保持原语义。

上述失败码必须在 save / read / list 中直接返回，不得被重写为成功、sample state 或 generic fixture fallback。

## 验收方式

- `go test ./internal/config ./internal/httpapi` 覆盖 config summary、field source、selector service failure 和 HTTP route fail-closed。
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-store-selector-smoke-v1.py` 校验 formal config、selector 源码、selector tests、HTTP tests、fixture/checker、文档引用和 forbidden artifact。
- 相邻 workflow checker 继续覆盖 selector smoke readiness、implementation entry review、adapter smoke readiness 和 schema artifact materialization review。
- `./scripts/check-repo.sh --fast` 作为仓库快速基线。
- 若触及 current focus、能力矩阵或阶段真相源，补跑 `./scripts/check-repo.sh`。

## 停止线

- 不创建 `SavedWorkflowDraftRepository` interface、repository adapter 或 adapter smoke。
- 不创建 migration root、schema artifact manifest file、DDL review、rollback evidence、migration smoke、SQL migration、schema version table 或 migration runner。
- 不接 Radish OIDC middleware、token validation、session cookie、workspace membership adapter 或 production API consumer。
- 不打开 publish、run、executor、confirmation decision、business writeback、replay、resume 或 materialized result reader。

## 下一步

若继续推进 durable store，应在 `Saved Workflow Draft Schema Artifact Materialization` 或后续 repository adapter implementation 中选择一个方向，并先补齐对应 implementation task card、fixture/checker 与验证链路。selector 当前只提供 fail-closed store mode 选择，不代表 repository mode 可用。

# `P2 Session & Tooling` 任务卡：Storage Backend Design

更新时间：2026-05-14

## 任务目标

本任务卡固定 `P2 Session & Tooling Foundation` 的 `storage_backend_design` 设计边界。

当前目标不是实现 durable session store、checkpoint store、audit store、result store、长期记忆、materialized result reader、业务写回或 replay，而是先定义 session / checkpoint / audit / result store 的职责分离、retention、redaction、secret handling、写入边界和禁止项。

程序化真相源为 `scripts/checks/fixtures/session-tooling-storage-backend-design.json`，快速门禁为 `scripts/check-session-tooling-storage-backend-design.py`。

## 当前边界

当前状态：`design_only_not_connected`。

当前 storage 只允许表达设计级 metadata：

- `durable_session_store_enabled=false`
- `durable_checkpoint_store_enabled=false`
- `durable_audit_store_enabled=false`
- `durable_result_store_enabled=false`
- `long_term_memory_enabled=false`
- `materialized_result_reader_enabled=false`
- `writes_business_truth=false`
- `replay_enabled=false`

因此本任务卡不会创建数据库、文件持久化后端、result reader、长期记忆或 replay state。

## Store 职责分离

`session_store` 未来只负责 bounded session metadata 和 history / state policy 摘要，不保存完整用户历史、长期记忆、raw tool output、credential 或业务真相源状态。

`checkpoint_store` 未来只负责 recovery checkpoint refs 和 metadata summaries，不保存 materialized result、result ref、executor ref、durable memory 或 replay plan。

`audit_store` 未来只负责 independent audit records，并且必须等待 retention / redaction policy 明确；不保存 raw tool output、credential、unredacted payload 或 materialized result payload。

`result_store` 未来只负责真实执行结果的 result refs；当前完全禁用，不允许 `result_ref`、`output_ref`、raw tool output、`materialized_result_uri` 或 `executor_ref`。

## Retention / Redaction / Secret

启用 durable storage 前必须定义：

- `retention_policy`
- `redaction_policy`
- `secret_handling_policy`
- `encryption_at_rest_policy`
- `access_control_policy`
- `deletion_policy`

当前禁止保存 credential、raw tool output、unredacted payload、完整用户历史作为长期记忆，或上层业务真相源状态。

## 写入边界

当前只允许 committed fixture、`tmp/` 检查产物和 metadata-only test outputs。

当前禁止：

- durable session store write
- durable checkpoint store write
- durable audit store write
- durable result store write
- long-term memory write
- business truth write
- replay state write

`RadishFlow`、`Radish`、`RadishCatalyst` 都仍是上层业务真相源，不得由本任务卡写入。

## 失败边界

当前必须保留三类失败边界：

- `materialized-result-read-denied`：返回 `CHECKPOINT_MATERIALIZED_RESULTS_DISABLED`，不返回 materialized result、result ref、output ref 或 raw output。
- `durable-memory-write-denied`：返回 `CHECKPOINT_DURABLE_MEMORY_DISABLED`，不写 durable memory 或长期记忆。
- `business-truth-write-denied`：返回 `BUSINESS_TRUTH_WRITE_DISABLED`，不写业务真相源，不执行 confirmed action。

这些 case 仍是 design-level gate，不代表完整 `negative_regression_suite` 已满足。

## 与 Session 的边界

session metadata 可以被 checkpoint read 引用，但当前没有 durable session store。

history policy 必须保持 bounded，不创建 long-term memory。

## 与 Checkpoint 的边界

checkpoint read 仍是 metadata-only 和 fixture-backed。

checkpoint store design 不能创建 materialized result reader，也不能在当前范围内返回 `executor_ref`、`result_ref` 或 replay plan。

## 与 Audit 的边界

audit store 当前是 `future_only_disabled`。

independent audit records 可以作为 metadata ref 被引用，但当前不是 durable write。审计记录不保存 raw tool output 或 credential。

## 与 Result 的边界

result store 当前是 `future_only_disabled`。

`result_ref` 和 `materialized_result_uri` 当前全部禁止。未来 result storage 必须依赖 executor boundary、result materialization policy、independent audit records 和 explicit reader authorization。

## 与 Executor 的边界

executor boundary 不创建 durable stores。

真实执行器未启用前，execution output 不能被持久化。storage writes 不能绕过 sandbox、allowlist 或 confirmation gate。

## 非目标

- 不实现 durable session store。
- 不实现 durable checkpoint store。
- 不实现 durable audit store。
- 不实现 durable result store。
- 不实现长期记忆。
- 不实现 materialized result reader。
- 不实现真实 executor。
- 不写 `RadishFlow`、`Radish` 或其它上层业务真相源。
- 不启用 automatic replay。

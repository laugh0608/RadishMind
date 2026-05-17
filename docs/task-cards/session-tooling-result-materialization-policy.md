# `P2 Session & Tooling` 任务卡：Result Materialization Policy

更新时间：2026-05-14

## 任务目标

本任务卡固定 `P2 Session & Tooling Foundation` 的 `result_materialization_policy` 设计边界。

当前目标不是实现真实 executor、durable result store、materialized result reader、长期记忆、业务写回或 replay，而是先定义 `metadata_only`、未来 `result_ref` 和未来 materialized result 的分层口径，并确认 checkpoint read 当前继续拒绝结果读取类请求。

程序化真相源为 `scripts/checks/fixtures/session-tooling-result-materialization-policy-design.json`，快速门禁为 `scripts/check-session-tooling-result-materialization-policy.py`。

## 当前边界

当前状态：`design_only_not_connected`。

当前唯一允许的模式是 `metadata_only`，且只代表 fixture / checkpoint read 的治理元数据边界：

- 可引用 `request_ref`、`session_record_ref`、`tool_audit_ref`、`tool_state_metadata_ref` 和 `checkpoint_ref`。
- 不可返回 `result_ref`、`output_ref`、`executor_ref`、`materialized_result_uri`、`business_truth_ref` 或 `replay_ref`。
- checkpoint read 只能返回治理摘要和 metadata refs，不返回真实工具输出。

## Result Ref

`result_ref` 当前是 `future_only_disabled`。

启用前至少需要：

- `real_tool_executor`
- `durable_result_store_design`
- `independent_audit_records_implemented`
- `upper_layer_confirmation_flow`
- `negative_regression_suite`

在当前 P2 范围内，`include_result_ref=true`、`result_ref=...`、`output_ref=...` 和 `executor_ref=...` 必须继续被 checkpoint read route 拒绝。

## Materialized Result

materialized result 当前是 `future_only_disabled`。

启用前至少需要：

- `retention_policy`
- `redaction_policy`
- `secret_handling_policy`
- `durable_result_store_design`
- `explicit_reader_authorization`
- `negative_regression_suite`

在当前 P2 范围内，`include_materialized_results=true`、`include_tool_results=1`、`materialize_results=on`、raw tool output 和 business truth write 都不得进入 checkpoint read response。

## 与 Confirmation 的边界

confirmation approval 不能在当前范围内创建 `result_ref` 或 materialized output。

即使后续存在 approved confirmation，动作也必须先停在 `approved_pending_execution_boundary`，不能直接变成 `execution_status=executed` 或结果已物化。

## 与 Executor 的边界

当前真实 executor 仍是 `not_implemented`，`executor_enabled=false`。

`executor_ref` 当前必须被 checkpoint read 和 result materialization policy 拒绝。未来如果启用 executor，执行输入输出 envelope 必须与模型 prompt 文本、checkpoint metadata response 分离。

## 与 Storage 的边界

当前 `durable_result_store_enabled=false`，`materialized_result_reader_enabled=false`。

metadata-only checkpoint read 不是 result store。未来 result storage 必须先定义 retention、redaction、secret handling 和 explicit reader authorization。

## 与 Audit 的边界

independent audit records 当前只能记录 result materialization denied 这类治理 metadata。

审计记录不保存 raw tool output 或 credential。未来任何 `result_ref` 或 materialized result access 都必须产生独立审计记录。

## 仍未满足的前置条件

完成本设计后，以下前置条件仍保持未满足：

- `upper_layer_confirmation_flow`
- `executor_boundary`
- `storage_backend_design`
- `negative_regression_suite`

`result_materialization_policy` 只达到 `design_boundary_defined_not_implemented`，不代表 result reader、durable result store 或 production policy 已完成。

## 非目标

- 不实现真实 executor。
- 不创建 `result_ref`。
- 不实现 materialized result reader。
- 不实现 durable result store。
- 不实现 durable session store 或 durable tool store。
- 不实现长期记忆。
- 不写 `RadishFlow`、`Radish` 或其它上层业务真相源。
- 不启用 automatic replay。

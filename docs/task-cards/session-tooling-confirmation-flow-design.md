# `P2 Session & Tooling` 任务卡：Confirmation Flow Design

更新时间：2026-05-14

## 任务目标

本任务卡固定 `P2 Session & Tooling Foundation` 的 confirmation flow 设计边界。

当前目标不是接入真实上层 UI、执行 confirmed action、写业务真相源或实现 replay，而是先定义确认记录、确认结果、失败边界和审计事件的最小设计口径。

程序化真相源为 `scripts/checks/fixtures/session-tooling-confirmation-flow-design.json`，快速门禁为 `scripts/check-session-tooling-confirmation-flow-design.py`。

## 当前边界

当前状态：`design_only_not_connected`。

当前确认流只允许记录设计级 metadata：

- `approve`：只记录 approval，进入 `approved_pending_execution_boundary`，不执行动作。
- `reject`：终止候选动作，不执行动作。
- `defer`：保持候选等待确认，不执行动作。

即使 outcome 为 `approve`，也不得启用 executor、result materialization、business truth write、durable memory 或 replay。

## Confirmation Record

确认记录必须至少包含：

- `confirmation_id`
- `session_id`
- `turn_id`
- `action_ref`
- `action_hash`
- `outcome`
- `actor_ref`
- `confirmed_at`
- `audit_ref`

`action_hash` 是防止 stale / mismatched confirmation 的最小稳定键。后续真实上层确认流如果使用不同 payload identity，需要先更新本任务卡、fixture 和负向回归。

## 负向边界

当前必须保留三类失败边界：

- missing confirmation：返回 `CONFIRMATION_REQUIRED`，不执行动作，不写业务真相源。
- stale confirmation：返回 `CONFIRMATION_STALE`，不 replay，不执行动作。
- mismatched confirmation payload：返回 `CONFIRMATION_PAYLOAD_MISMATCH`，不执行动作，不返回 result ref。

这些 case 仍是 design-level gate，不代表完整 `negative_regression_suite` 已满足。

## Audit Events

当前只固定审计事件名，不实现 durable audit store：

- `confirmation_requested`
- `confirmation_recorded`
- `confirmation_rejected`
- `confirmation_deferred`
- `confirmation_invalidated`

后续真实 audit record 必须独立于 checkpoint read response，不得用 checkpoint metadata response 冒充独立审计记录。

## 仍未满足的前置条件

完成本设计后，以下前置条件仍保持 `not_satisfied`：

- `upper_layer_confirmation_flow`
- `independent_audit_records`
- `negative_regression_suite`
- `result_materialization_policy`

因此本任务卡不会解除 [实现前置条件](session-tooling-implementation-preconditions.md) 中 confirmation 的 `not_ready` 状态。

## 非目标

- 不实现真实 executor。
- 不执行 confirmed action。
- 不写 `RadishFlow`、`Radish` 或其它上层业务真相源。
- 不实现 durable confirmation store。
- 不实现 durable audit store。
- 不实现 materialized result reader。
- 不启用 automatic replay。

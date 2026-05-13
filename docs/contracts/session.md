# RadishMind 会话记录契约

更新时间：2026-05-13

## 文档目的

本文档固定 `Conversation & Session` 的首版最小契约。当前目标是把 `conversation_id` 从简单透传升级为可审计的 session record，而不是立即实现持久化 session store、长期记忆或自治恢复循环。

Schema 真相源为 `contracts/session-record.schema.json`，最小 fixture 为 `scripts/checks/fixtures/session-record-basic.json`，快速门禁为 `scripts/check-session-record-contract.py`。

## 最小结构

`SessionRecord` 必须表达：

- `session_id` / `conversation_id`：会话身份；v1 默认保持两者一致，便于兼容已有调用方。
- `turn_id` / `parent_turn_id`：当前轮次和上一稳定轮次引用。
- `history_policy`：历史窗口、是否包含 system/tool 结果、是否使用 summary 压缩。
- `recovery_record`：恢复状态、最后稳定轮次、是否可 replay 和 checkpoint 引用。
- `audit`：必须保持 advisory-only，不写 `RadishFlow`、`Radish` 或其他上层业务真相源。

## Northbound 兼容层

`/v1/chat/completions`、`/v1/responses` 与 `/v1/messages` 的 `radishmind` 扩展当前支持以下会话字段：

```json
{
  "conversation_id": "conv-123",
  "turn_id": "turn-002",
  "parent_turn_id": "turn-001",
  "history_policy": "windowed",
  "history_window": 6
}
```

当这些字段存在时，`Go` 平台层会在 canonical `CopilotRequest.context.northbound.session` 中写入 `conversation_session_record`。该记录只用于审计、窗口策略和恢复边界说明，不代表平台已经拥有 durable session store。

## 当前停止线

- 不把 session record 当作上层业务状态源。
- 不在 v1 中引入共享长期记忆。
- 不让 northbound compatibility layer 自行持久化用户历史。
- 不把 recovery record 写成自动恢复执行计划；它只记录可审计边界和 checkpoint 引用。

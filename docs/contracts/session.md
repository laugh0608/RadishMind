# RadishMind 会话记录契约

更新时间：2026-05-16

## 文档目的

本文档固定 `Conversation & Session` 的首版最小契约。当前目标是把 `conversation_id` 从简单透传升级为可审计的 session record，而不是立即实现持久化 session store、长期记忆或自治恢复循环。

Schema 真相源为 `contracts/session-record.schema.json`、`contracts/session-recovery-checkpoint.schema.json`、`contracts/session-recovery-checkpoint-manifest.schema.json` 与 `contracts/session-recovery-checkpoint-read.schema.json`。最小 fixture 为 `scripts/checks/fixtures/session-record-basic.json`、`scripts/checks/fixtures/session-recovery-checkpoint-basic.json`、`scripts/checks/fixtures/session-recovery-checkpoint-manifest-basic.json` 与 `scripts/checks/fixtures/session-recovery-checkpoint-read-basic.json`。快速门禁为 `scripts/check-session-record-contract.py` 与 `scripts/check-session-recovery-checkpoint-contract.py`。

## 最小结构

`SessionRecord` 必须表达：

- `session_id` / `conversation_id`：会话身份；v1 默认保持两者一致，便于兼容已有调用方。
- `turn_id` / `parent_turn_id`：当前轮次和上一稳定轮次引用。
- `history_policy`：历史窗口、是否包含 system/tool 结果、是否使用 summary 压缩。
- `state_policy`：会话状态落点、tool result cache 范围、recovery checkpoint 引用范围；v1 只允许 request-local / northbound metadata / session recovery checkpoint 这类受限边界，不启用 durable memory。
- `recovery_record`：恢复状态、最后稳定轮次、是否可 replay 和 checkpoint 引用。
- `audit`：必须保持 advisory-only，不写 `RadishFlow`、`Radish` 或其他上层业务真相源。

`SessionRecoveryCheckpoint` 必须表达：

- `checkpoint_id`、`session_id`、`turn_id`：checkpoint 身份与所属会话轮次。
- `storage_policy`：当前只允许 fixture / local artifact / session recovery store 这类引用落点，且不启用 durable memory、不写业务真相源。
- `replay_policy`：是否 replayable、是否要求人工确认；v1 强制 `auto_replay_enabled=false`。
- `refs`：request、session record、tool audit、tool state、tool result metadata 等引用，不保存真实工具执行结果。
- `state_summary` 与 `audit`：明确不包含 materialized tool results，不包含业务真相源。

`SessionRecoveryCheckpointReadResult` 必须表达：

- `api_boundary`：当前只定义平台 metadata 读取边界，`implemented=false` 表示没有 durable checkpoint store、materialized result 读取或 replay executor；平台层可以暴露 fixture-backed route smoke，但不得把它声明为真实恢复 API。
- `request`：按 `checkpoint_id / session_id / turn_id` 查询，并强制 `include_materialized_results=false`。
- `result`：只返回 checkpoint ref、metadata refs、tool audit 治理摘要、replay policy 摘要和 state summary。
- `tool_audit_summary`：只暴露 execution disabled、not executed、metadata-only cache、无 result ref、无 durable memory 和不写业务真相源等治理元数据；不得返回真实工具输出。
- `access_policy`：必须保持 metadata-only、不返回真实工具结果、不写业务真相源、不启用 durable memory 或 automatic replay。

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

## Promotion 门禁分层

当前 session / checkpoint 只能按以下层级晋级，不允许跳层声明能力：

| 层级 | 当前门禁 | 可声明能力 | 不可声明能力 |
| --- | --- | --- | --- |
| Contract gate | `session-record`、`session-recovery-checkpoint*` schema、正向 fixture、`session-recovery-checkpoint-read-denied-queries` 负向 fixture | 会话身份、history/state/recovery policy、checkpoint refs、metadata-only read result 结构稳定 | durable session store、长期记忆、真实 checkpoint storage |
| Platform route smoke | `GET /v1/session/recovery/checkpoints/{checkpoint_id}` fixture-backed route、禁止 materialized result / replay / durable memory 查询参数 | 平台能稳定暴露 metadata-only checkpoint read shape，并拒绝越界读取和执行请求 | materialized result reader、跨轮 replay executor、自动恢复 API |
| Fast check | `check-session-record-contract.py`、`check-session-recovery-checkpoint-contract.py`、`go test ./...` 经 `check-repo --fast` 间接覆盖 | 日常开发能复验 session/checkpoint 治理不变量 | 生产级持久化、外部 provider 健康、真实上层确认流 |
| Governance rollup gate | `session-tooling-short-close-readiness-delta`、`session-tooling-route-smoke-readiness-rollup`、`session-tooling-readiness-consistency-rollup`、`session-tooling-stop-line-manifest` | close candidate 后的剩余 blocker、future route smoke 缺口、跨 rollup 漂移和停止线可复验 | `P2 short close`、真实 checkpoint store、materialized result reader、automatic replay |
| Future implementation gate | 上层确认流接线、independent audit、result materialization policy、executor boundary、storage backend、enablement plan entry condition 和完整负向回归同时满足后再定义 | 可讨论有限 durable store、受控 result reader 或受控 replay 的实现条件 | 在本阶段直接启用自动 replay、长期记忆、业务写回或持久化读取 |

## 治理证据链

`P2 Session & Tooling Foundation` 的 session / checkpoint 能力当前由 `scripts/checks/fixtures/session-tooling-short-close-readiness-delta.json`、`session-tooling-route-smoke-readiness-rollup.json`、`session-tooling-readiness-consistency-rollup.json` 和 `session-tooling-stop-line-manifest.json` 共同约束。它们只证明停止线、阻塞项和 route smoke 缺口可检查，不证明 durable checkpoint store、materialized result reader 或 replay executor 已实现。

## 当前停止线

- 不把 session record 当作上层业务状态源。
- 不在 v1 中引入共享长期记忆。
- 不让 northbound compatibility layer 自行持久化用户历史。
- 不把 recovery record 写成自动恢复执行计划；它只记录可审计边界和 checkpoint 引用。
- 不把 tool result cache 升级为长期记忆；当前只允许 request-local metadata 或 session recovery checkpoint 引用。
- 不让 recovery checkpoint 自动 replay；当前只固定 record / manifest 与可审计引用。
- 不把 checkpoint read route smoke 写成 durable checkpoint store、materialized result reader 或 replay executor；当前只冻结 response shape 和安全边界。
- checkpoint read route 必须拒绝 materialized result、result ref、executor ref、replay 和 durable memory 类查询参数，并在平台 smoke 中覆盖这些治理门禁。

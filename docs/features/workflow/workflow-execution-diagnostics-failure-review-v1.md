# Workflow Execution Diagnostics / Failure Review v1

更新时间：2026-07-16

状态：`workflow_execution_diagnostics_failure_review_v1_completed`

## 功能目标

让内部开发者在现有 Workspace Run History 中定位 executor v0 的失败边界、失败节点和后续人工审查动作，并用稳定、脱敏、可过滤的诊断记录复验 Gateway、provider、执行预算、取消和 run store 故障。该能力只扩展现有 `/v1/user-workspace/workflow-runs` 列表与详情，不创建第二套运行真相源。

本专题不把错误诊断变成自动修复。v0 / v1 仍保持零工具副作用；独立受控 HTTP Tool v2 只复用并扩展诊断 schema，不开放自动 retry / fallback、replay / resume、checkpoint、业务写回或 production enablement。

## 用户路径

1. 用户从符合资格的 saved draft 启动受控运行；显式开发 / 测试诊断模式可以选择固定故障场景。
2. 运行创建后，Platform 在每次节点推进时保存最后完成节点；失败时记录失败边界、失败阶段、失败节点、Gateway 失败类别和稳定诊断摘要。
3. Workspace Run History 展示成功、失败、取消和 stale running 指标，并允许按失败码、失败边界、provider、model 与 stale running 过滤。
4. 用户打开详情审查节点时间线、最后完成节点、失败节点、sanitized diagnostic、推荐人工动作、request id 和 audit ref。
5. `workflow_run_record.v0` 历史记录继续可读，Web 明确显示“旧记录无结构化诊断”，不推断不存在的数据。
6. 默认 offline 模式不发 HTTP；真实 dev / test 模式只读取现有 workflow-runs 资源族。

## 失败边界与分类

`failure_boundary` 使用固定枚举：`draft_read`、`executor`、`gateway`、`provider`、`run_store`、`request`、`tool_policy`、`tool_confirmation`、`tool_transport`、`tool_response`、`tool_store`。资格失败发生在 run id 创建前时不创建记录；run store create 失败同样没有伪历史。

`gateway_failure_category` 使用 `none|queue_full|timeout|canceled|worker_crash|protocol|provider_failed|output_unavailable|unavailable`。通用 `failure_code` 保持 executor v0 稳定码，结构化类别用于审查和过滤，不暴露 bridge / provider 原始错误。

v2 的 `tool_failure_category` 独立使用 `none|policy|confirmation|transport|timeout|response_status|response_too_large|response_invalid|store|outcome_unknown`。它只描述工具 attempt，不替代 Gateway 类别：工具 attempt 成功、后续模型节点失败时，`tool_failure_category=none`，run 继续记录真实 `gateway|provider` 边界；可能已 dispatch 且结果不明时才记录 `outcome_unknown`。

## 记录生命周期与兼容

- 新运行写入 `workflow_run_record.v1`；repository 同时接受和读取已存在的 `workflow_run_record.v0`。
- v1 增加 `diagnostic`：`failure_boundary`、`failure_stage`、`failed_node_id`、`last_completed_node_id`、`terminal_write_state`、`gateway_failure_category`、`summary`、`recommended_review_action` 和 `observed_at`。
- `terminal_write_state` 只允许 `pending|stored`。运行中的最后快照为 `pending`；终态写入前改为 `stored` 并随终态原子提交。若终态写失败，数据库仍保留上一版 `pending` running 记录，API 返回 store failure，不伪造持久化成功。
- `running` 超过 30 秒派生为 `stale_running`；它不是新持久化状态，不自动终态化、不恢复、不重放。
- `last_completed_node_id` 只在节点成功或明确 skipped 后推进；`failed_node_id` 只在节点失败时出现。
- v0 记录没有 `diagnostic`；读取端不得回填推断值。
- v2 保留 v1 字段并增加 `tool_failure_category`；`outcome_unknown` 是持久化终态，不等同于 v1 派生的 `stale_running`。若终态提交失败，调用结果可返回 `terminal_write_state=pending` 的 outcome unknown 提示，而 durable 记录保持 claimed / running，等待超出 30 秒预算后的 metadata-only reconciliation 以 CAS 写入终态。

## 推荐审查动作

`recommended_review_action` 只允许 `review_draft`、`check_gateway_capacity`、`check_provider_configuration`、`check_run_store`、`start_new_run`、`check_tool_policy`、`review_tool_outcome`。这些值是只读建议，不对应执行 API；尤其 `review_tool_outcome` 不授权重试或恢复已 consumed 的计划。

## API、过滤与分页

现有 list 新增 `failure_code`、`failure_boundary`、`provider`、`model` 和 `stale_running=true|false` 单值过滤，可与原有状态、草案和时间条件组合。provider / model 精确匹配，最多 256 字符且禁止 endpoint 形式。cursor 继续使用 `started_at DESC, run_id DESC` keyset，并将全部过滤条件纳入摘要。

summary 增加失败边界、失败节点、最后完成节点、Gateway 类别、推荐动作和既有 stale 标记。detail 返回完整 v1 diagnostic。list / detail 的 tenant、workspace、application scope 与 no-fallback 行为不变。

## 开发 / 测试故障注入

故障注入只在 executor dev gate、mock provider 和独立 `RADISHMIND_WORKFLOW_DIAGNOSTICS_DEV=true` 同时满足时接受。请求字段 `dev_failure_scenario` 只允许：`gateway_timeout`、`gateway_queue_full`、`gateway_worker_crash`、`gateway_protocol_failure`、`provider_failed`、`output_unavailable`、`request_canceled`、`run_store_unavailable`、`terminal_write_conflict`、`budget_exceeded`、`stale_running`。

默认关闭；非 allowlist、非 mock provider 或未开 gate 均 fail closed。场景不接受附加 payload、endpoint、credential、命令、持续时间或 provider envelope。真实 Gateway 错误仍通过同一稳定分类器映射。

## 持久化与 migration

新增独立 `workflow_runs` 0002 manual migration，为列表过滤物化 `failure_code`、`failure_boundary`、`selected_provider` 和 `selected_model` 列及 scope 索引；既有 v0 行通过 JSONB 中已有稳定字段回填，缺失诊断列保持空值。migration marker 升级到 v2，支持从 0001 递进 apply、新库直接 apply 和完整 rollback / reapply；runtime 仍只做 marker preflight，不执行 DDL。

Saved Draft repository 不参与 run diagnostics。`memory_dev` 继续是默认；`postgres_dev_test` 继续独立显式 opt-in，数据库失败不得回退内存成功。

## 脱敏、保留与可观测性

诊断只保存固定枚举和服务端 allowlist 摘要。禁止保存原始 input、condition value、credential、endpoint、DSN、provider raw envelope、stderr、stack trace、SQL、bridge 帧和任何 tool / confirmation / writeback / replay payload。摘要最多 256 字符。既有 30 天 / 每 scope 10,000 条 dev-test 保留声明不变。

服务日志只记录稳定 outcome、failure boundary / category、store mode、duration、request / audit ref 和 scope hash。Web 展示 failed、canceled、stale、Gateway / provider 和 store failure 指标；详情突出 failed node 与 last completed node，提供复制 request id / audit ref，v0 明确 diagnostic unavailable。诊断 UI 留在 Run History lazy chunk，并继续拆分非首屏静态证据，使主包本批低于 650 KiB，后续再收敛到 500 KiB。

## 验收

- Go：taxonomy、v0 / v1 decode、record contract、失败节点、last completed、terminal write、stale、filter / cursor、scope、并发与 no side effects。
- PostgreSQL：0001 → 0002、fresh apply、rollback / reapply、重启恢复、诊断过滤、版本冲突、连接失败与 no fallback。
- HTTP：strict query、故障注入 gate、list / detail 对齐、失败前无记录与终态写失败的 stale 记录。
- Web：v0 / v1 mapping、指标、过滤、详情、copy refs、offline 不请求、禁止字段拒绝和 lazy chunk。
- 浏览器：成功、Gateway timeout、provider failure、取消、store create failure、terminal conflict / stale、过滤、详情、重启恢复与敏感字段缺失。
- v0 / v1 路径继续验证 provider call 有界且 tool / confirmation / business write / replay 为 0；v2 单独验证一个已授权 tool / confirmation attempt、业务写回和 replay 为 0，以及下游失败不覆盖工具 attempt 结果。

## 完成结果

- 新运行写入 `workflow_run_record.v1` structured diagnostic，PostgreSQL 与 Web 保持 `workflow_run_record.v0` 只读兼容；terminal write 失败不声明存储成功，最后 running 快照由 stale 标记审查。
- 现有 list / detail 已支持失败码、失败边界、provider、model 与 stale 过滤，cursor 升级到 v2 并绑定全部过滤条件；scope 与 no-fallback 不变。
- 0002 migration 已覆盖 fresh apply、0001 pending marker 递进、rollback / reapply、物化诊断列和索引；真实集成测试覆盖重启、v0 / v1、过滤、CAS、运行角色 DDL 拒绝与 no fallback。
- 显式 `RADISHMIND_WORKFLOW_DIAGNOSTICS_DEV` 与 mock provider gate 支持固定故障场景；shell / PowerShell launcher 都要求独立 opt-in，未开放任意 payload。
- Web 已完成失败 / 取消 / stale / Gateway / store 指标、诊断过滤、固定场景生成、failed / last completed 节点、legacy 状态和 request / audit ref 复制；默认 offline 不请求。
- R5 同批把 Node Designer 与三个非首屏审查面板拆为 lazy chunks，主包从 850.39 KiB 降到 624.57 KiB；下一工程目标继续收敛到 500 KiB。
- 真实浏览器验证成功、Gateway timeout、provider failure、取消、store create failure、terminal conflict / stale、stale-only 过滤和 Platform 重启恢复。恢复后 5 条记录完整，详情禁止字段扫描为 0，tool / confirmation / business write / replay 为 0。

## 停止线

- 不新增平行运行 API，不读取旧 fake `/v1/user-workspace/runs` 作为真实历史。
- 不从本诊断专题新增 tool、RAG、confirmation 或执行 route；受控 HTTP Tool v2 只消费既有诊断扩展。不实现业务写回、自动 retry / fallback、replay、resume、checkpoint 或自动修复。
- 不启用 production repository、production auth、production secret、audit store 或公开生产 API。
- 不新增同层 checker；使用现有测试、浏览器验证和聚合门禁。

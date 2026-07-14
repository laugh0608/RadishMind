# Workflow Run History / Durable Dev-Test Run Store v1

更新时间：2026-07-14

状态：`workflow_run_history_durable_dev_test_store_v1_sqlite_repository_completed`

## 功能目标

让内部开发者在 Workspace Run History 中审查真实 executor v0 运行历史，并在显式 PostgreSQL 开发 / 测试模式下跨平台服务重启恢复运行记录。该能力延续 `workflow_run_record.v0` 和 `/v1/user-workspace/workflow-runs` 资源族，不把旧 fake `/v1/user-workspace/runs` 作为运行真相源。

本专题只扩展历史读取、受限记录持久化与执行可观测性。tool、RAG、confirmation commit、业务写回、replay / resume、生产身份和 production enablement 继续关闭。

## 用户路径

1. 用户从已保存且符合 executor v0 资格的草案启动运行。
2. Platform 在运行开始时创建 `running` 记录，并在节点推进和终态时更新同一记录。
3. Workspace Run History 按当前 tenant / workspace / application 加载真实记录，展示状态、草案、开始时间、耗时、模型选择和副作用摘要。
4. 用户可按状态、草案和时间区间过滤，并使用稳定游标加载更早记录。
5. 用户选择一条历史记录进入现有详情审查，查看节点时间线、sanitized output、失败分类、request / audit ref 和零副作用证据。
6. PostgreSQL 开发 / 测试模式下，平台服务重启后列表和详情仍能恢复；默认离线 Web 模式继续展示明确标记的 sample，不发真实请求。

## 运行记录生命周期

- `running`：run id 创建后、首个节点执行前必须持久化。后续写入只能更新相同 scope 下的相同 run id。
- `succeeded`、`failed`、`canceled`：终态不可逆；数据库层以终态保护谓词拒绝回到 `running`，也拒绝从一个终态改成另一个终态。
- 同一记录允许 `running → running` 的节点进度更新，并使用单调递增 `record_version` 防止并发旧快照覆盖新快照。
- create 要求 `record_version=1`；每次成功 update 原子递增一。版本冲突返回 store unavailable 类失败并使本次执行失败关闭，不回退内存成功。
- executor 同步 POST 仍在终态返回。若终态持久化失败，不得返回成功记录；允许数据库中保留可观测的最后一个 `running` 快照，列表以 `stale_running` 派生标记提示人工审查，但不自动恢复、重放或改写终态。
- 记录不提供删除、修改、replay、resume 或业务写回 API。

## Scope 与授权

所有 create / update / read / list predicate 必须包含：

- `tenant_ref`
- `workspace_id`
- `application_id`
- `run_id`（详情或写入）

owner 不作为 run 可见性边界；同一 workspace / application 内具备 `workflow_runs:read` 的成员可审查该 scope 的历史。记录保留 `actor_ref` 作为 metadata，但 API 不提供 actor 任意文本过滤。POST 继续要求 `workflow_runs:execute` 与 `workflow_drafts:read`；list / detail 要求 `workflow_runs:read`。任何 header、query 与身份 binding 不一致都返回 scope denied，且不暴露记录是否存在。

## List API、过滤与游标

新增：

```text
GET /v1/user-workspace/workflow-runs?workspace_id=...&application_id=...&limit=...&cursor=...&status=...&draft_id=...&started_from=...&started_to=...
```

规则如下：

- 默认 `limit=25`，允许 `1..100`。
- 排序固定为 `started_at DESC, run_id DESC`。
- cursor 是不透明、版本化、URL-safe base64 JSON，只编码上一页末项的 `started_at` 与 `run_id`，并带当前过滤条件摘要；服务端必须校验版本、字段 allowlist 和过滤摘要，篡改或跨过滤复用返回 `workflow_run_cursor_invalid`。
- 下一页使用严格 keyset predicate：`started_at < cursor.started_at OR (started_at = cursor.started_at AND run_id < cursor.run_id)`，不得使用 offset。
- `status` 可选单值：`running|succeeded|failed|canceled`；`draft_id` 可选精确匹配。
- `started_from` / `started_to` 使用 RFC3339 UTC 边界，均为包含关系；from 晚于 to、未来偏差超过 5 分钟或区间超过 90 天返回 `workflow_run_filter_invalid`。
- list 返回 sanitized summary、`next_cursor` 和 `has_more`。读取 `limit+1` 行决定是否有下一页；空列表是成功，不回退 fake / sample。
- detail API 保持 `GET /v1/user-workspace/workflow-runs/{run_id}`，与 list 使用同一 scope、失败语义和 store。

## 持久化、保留与脱敏

默认 `memory_dev` 继续使用每个进程最多 100 条 FIFO 记录，并新增同语义 scoped list。独立 opt-in 模式为：

```text
RADISHMIND_WORKFLOW_RUN_STORE=postgres_dev_test
RADISHMIND_WORKFLOW_RUN_DEV_TEST_DATABASE_URL=...
RADISHMIND_WORKFLOW_RUN_DATABASE_TIMEOUT=5s
```

`postgres_dev_test` 还要求现有 control-plane dev auth、Saved Draft dev read 与 executor dev gate 显式开启。它使用独立 `workflow_run_records` 表、schema marker、manual migration 命令和连接池；可以复用 `pgx` 依赖与数据库连接做法，但不复用、改名或扩张 Saved Draft repository。

### SQLite 开发 repository

S2 增加组件级 `sqlite_dev` 选择，只供专项测试和后续聚合本地启动档注入共享 SQLite runtime。它要求 control-plane dev auth、Saved Draft dev read 与 executor dev gate 同时开启；缺少共享 runtime、缺少或不匹配本组件 migration marker、数据库关闭或查询失败时必须失败关闭，不创建组件私有连接，也不回退 `memory_dev`。

SQLite 使用独立 STRICT `workflow_run_records` 表，主键仍为 tenant、workspace、application 与 run id。表内同时保存脱敏 JSON 记录和 draft / record version、schema、状态、开始 / 完成时间、失败分类、provider / model 与审计 metadata 等筛选列；读取和列表必须复验物理列与 JSON 一致，任一记录损坏时拒绝整次结果，不返回部分列表。

开始与完成时间使用 UTC Unix 纳秒整数，超出纳秒可表达范围的时间拒绝写入；排序固定为 `started_at_unix_nano DESC, run_id DESC`，时间区间、stale running 与游标均使用物理整数列。create 只接受 version 0 的 `running` 记录并写为 version 1；update 必须同时命中完整 scope、run id、预期版本、原始开始时间和 `running` 状态，再由数据库原子递增版本。终态不可逆，不以进程锁替代数据库 CAS。

本批只持久化工作流运行记录，不把 PostgreSQL migration 中的 evaluation case、revision、suite 或 release decision 扩入 SQLite。聚合 `sqlite_dev`、跨平台本地启动档和连续 HTTP 产品链后续均已完成；PostgreSQL migration、运行角色、tuple pagination、多连接并发、重启恢复和 no-fallback 也已由真实数据库门禁复验。下一项只进入 API 密钥 Web 一次性交接与浏览器连续验收，不扩张运行历史职责。

开发 / 测试 v1 不在请求路径自动删除记录。默认保留策略声明为 30 天、每 scope 最多 10,000 条，由后续显式维护命令承接；本批 schema 和查询必须支持 `completed_at / started_at` 保留索引，但不实现后台清理 goroutine。超过策略的环境应由操作者清理或重建 disposable 数据库，不得在 list 时隐式删除。

禁止持久化：

- 原始 `input_text` 或可逆输入摘要；只保留 UTF-8 字节数。
- condition 布尔值；只保留排序后的 condition node id。
- credential、secret ref 解引用结果、endpoint、DSN 或 provider raw envelope / raw error。
- tool payload/result、confirmation decision、业务写回 payload、checkpoint、replay / resume state。

允许持久化 `workflow_run_record.v0` 的 sanitized node preview 和最终 advisory output，分别受 512 字符和 16 KiB 预算约束。失败摘要只能来自稳定 allowlist 文案。`tool_calls`、`confirmation_calls`、`business_writes`、`replay_writes` 必须持续为 0；repository 在写入前复核这些计数和 record schema，违反时 fail closed。

## Schema 与 migration

独立 migration source：

```text
services/platform/migrations/workflow_runs/
```

物化 `workflow_run_schema_versions` 与 `workflow_run_records`：完整 scope、run / draft identity、record / draft version、状态、started / completed 时间、sanitized JSONB record、request / audit / actor metadata。主键为完整 scope + run id；列表索引覆盖 scope + started_at + run_id，并为 scope + status、scope + draft_id、保留时间提供索引。

一次性 migration 读取独立 migration URL 环境变量。平台启动只做连接与 marker / checksum preflight，不自动 migration；runtime role 无 DDL 权限。up 必须幂等，reviewed down 只供 disposable 集成测试验证 rollback / reapply。

## 失败语义

| failure code | 行为 |
| --- | --- |
| `workflow_run_scope_denied` | 身份、scope 或 header / query binding 不一致；不暴露记录存在性 |
| `workflow_run_record_not_found` | 当前 scope 不存在详情记录；不回退 fake route |
| `workflow_run_filter_invalid` | limit、status、draft 或时间过滤非法 |
| `workflow_run_cursor_invalid` | 游标格式、版本或过滤摘要不匹配 |
| `workflow_run_store_unavailable` | store、连接、marker、scan 或写入失败；no fallback |
| `workflow_run_store_contract_mismatch` | record schema、scope、状态迁移、版本或零副作用约束不满足 |
| `workflow_run_store_mode_invalid` | selector 不在 allowlist；fail closed |
| `workflow_run_store_mode_disabled` | production / reserved mode 未启用；fail closed |

数据库原始错误、SQL、连接信息和 record JSON 不进入 HTTP failure summary。list 失败不返回部分结果；executor 任一必需记录写入失败不声明成功。

## 可观测性

- 每次 create / update / read / list 记录 operation、store mode、稳定 outcome、duration、request id、audit ref、scope hash、result count 和 has_more；不记录输入、输出正文、cursor 原文或数据库错误。
- run record 保留 started / completed time、节点耗时、provider selection、request / audit ref 和副作用计数，Web 对 `running` 超过 executor 预算的记录显示 `stale_running` 审查提示。
- list envelope 返回当前 request id / audit ref；单条 record 仍保留创建运行时的 request / audit ref，二者不得混淆。

## Web 边界

- 新增独立 run-history client、domain mapping 和 UI 模块，不继续扩张 `App.tsx`；通过动态 import / route-level lazy chunk 将历史审查从当前约 862 KiB 主包拆出。
- Workspace Run History 在真实 dev source 下只调用 `/v1/user-workspace/workflow-runs` list / detail；旧 `/v1/user-workspace/runs` 只保留历史 sample surface，不得合并成真实结果。
- 列表提供状态、草案、时间过滤和“加载更早记录”；详情复用 executor v0 的节点时间线、Gateway selection、advisory output、failure 与零副作用审查。
- 默认 offline source 不发 HTTP，明确展示 sample / offline 标识。route disabled、空列表、过滤失败、游标失败和 store unavailable 都有独立 UI 状态。

## 验收

- Go domain / store：生命周期、并发 update、FIFO、克隆、scope、过滤、稳定 keyset 分页、游标篡改与零副作用契约。
- HTTP：strict query、auth/scope、list/detail 对齐、空列表、失败映射、fake route 不参与。
- PostgreSQL：migration 幂等、rollback / reapply、重启恢复、并发版本保护、scope 隔离、分页、marker mismatch、连接中断和 no fallback。
- Web：配置、list/detail mapping、过滤、分页、offline 不请求、敏感字段拒绝和独立 chunk。
- 真实浏览器：创建多条成功 / 失败记录，按 scope 与过滤分页审查；重启平台后列表和详情恢复；确认四类禁止副作用计数均为 0。
- 最终执行 Go test / race / vet、Web test / build、仓库 fast / full 门禁，并关闭本批服务。

## 完成结果

- 已实现带 `record_version` 原子保护和终态不可逆约束的独立 run store contract；`memory_dev` 保持默认 100 条 FIFO，并支持 scope、过滤和 keyset 分页。
- 已实现独立 `postgres_dev_test` selector、`workflow_run_records` schema、marker / checksum / advisory lock、manual `status / up` runner、reviewed rollback、runtime role DDL 拒绝、连接池生命周期和 no fallback。
- `GET /v1/user-workspace/workflow-runs` 已与 detail API 对齐 tenant / workspace / application scope，支持状态、草案、时间区间与不透明游标；旧 `/v1/user-workspace/runs` 不参与真实历史。
- Go 单元与 PostgreSQL 集成测试覆盖并发版本保护、scope、分页、重启恢复、migration / rollback / reapply、marker mismatch、运行角色权限和数据库失败；PR / release PostgreSQL job 使用 `-count=1` 禁止缓存真实集成结果。
- Web 真实 history / detail consumer 已拆为约 9.9 KiB lazy chunk；主包从约 862 KiB 降到 848.4 KiB。默认 offline 不发请求，dev/test 历史只读取新资源族。
- 真实浏览器完成 26 条记录的 25+1 分页、草案过滤、详情节点时间线和 Platform / Web 重启恢复；详情显示 provider call 为 1，tool / confirmation / business write / replay 合计为 0，原始输入和 condition value 未出现。最终全新会话无 console error / warning。
- 浏览器截图保存在本地 `output/playwright/workflow-run-history-durable-dev-test/`，不作为 committed 真相源；复验后已关闭浏览器、Platform / Web 与 PostgreSQL 容器和网络。
- 已新增独立 SQLite migration、STRICT 表和 `sqlite_dev` run store；memory、SQLite 与 PostgreSQL 复用同一严格 JSON 编解码、领域校验、完整 scope、筛选、稳定 keyset 顺序、版本 CAS 和终态不可逆语义。
- SQLite 专项验证覆盖等时刻 run id 排序、完整 scope 隔离、16 路终态单写者、真实 executor 运行、重启恢复、关闭不回退、marker mismatch、未知 document 字段、物理列漂移、损坏记录无部分列表、原始输入禁入和超出纳秒范围时间拒绝。evaluation case / suite 未扩入 SQLite。
- S2 第七组 repository、聚合 shared runtime、跨平台本地启动档、SQLite 连续产品链与 PostgreSQL 专项门禁均已完成；下一项是 API 密钥 Web 一次性交接和浏览器连续验收，production 与工作流 Web 新能力没有开放。

## 停止线

- 不实现 tool、RAG、confirmation commit、业务写回、publish、schedule、retry/fallback、replay、resume、checkpoint 或 materialized result reader。
- 不启用 production repository、OIDC、membership、production secret、audit store、公开生产 API 或自动保留任务。
- 不持久化原始输入、condition value、credential、endpoint 或 provider raw envelope。
- 不把 Saved Draft repository 改造成 run repository，不把旧 fake run API 作为运行真相源。

后续 [Workflow Execution Diagnostics / Failure Review v1](workflow-execution-diagnostics-failure-review-v1.md) 已在本资源族上完成 v1 诊断、失败过滤、受控故障场景和 Web 失败审查；本专题的持久化与 scope 边界保持不变。

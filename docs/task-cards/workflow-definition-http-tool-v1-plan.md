# Workflow Definition 绑定受控 HTTP Tool v1 实施任务卡

更新时间：2026-08-15

- 任务 ID：`workflow-definition-http-tool-v1`
- 状态：`workflow_definition_http_tool_v1_completed`
- 功能设计：[Workflow Definition 绑定受控 HTTP Tool 的版本化发布与人工确认执行（开发 / 测试态）v1](../features/workflow/workflow-definition-http-tool-v1.md)

## 准入结论

本卡是该功能唯一的高风险实施入口。现有 Definition release、HTTP Tool action / confirmation / transport 和三种 durable store 均已有行为证据，可以在不建立新 owner 的前提下增加 Definition 来源。任务先完成显式 profile、合同和 memory 领域链，再进入数据库、执行和 Web；不派生同层 readiness 卡。

## 完成目标

内部开发者能够把精确 HTTP Tool 草案晋级为不可变 Definition、显式激活，从 active Definition 创建一次 action plan，经人工确认后最多执行一次，并在 Run History 审查 Definition 来源、确认、attempt、失败和副作用证据。

## 不可变边界

1. `workflow_definition_executor_v1/v2` 保持零工具语义；新增能力只通过 `workflow_definition_http_tool_v1` 打开。
2. Definition review / activation 不创建 plan、run，不调用 provider 或网络。
3. Definition approval 不能替代每次工具执行确认；action plan、confirmation、claim 和 transport 继续由既有 HTTP Tool owner 负责。
4. action plan 来源必须是严格 union；`saved_workflow_draft` 和 `workflow_definition` 不得同时出现、缺失或互相 fallback。
5. Definition 来源执行前必须重读 active pointer、version / digest、tool definition/profile 和 Application lifecycle。
6. claim 继续是单一线性化点；claim 前失败网络为 0，claim 后结果不确定进入 `outcome_unknown`。
7. memory、SQLite、PostgreSQL 复用既有 backend mode、pool、migration family 和 selector，不增加专用存储。
8. 不持久化 endpoint、header、credential、raw body、原始 prompt / answer 或业务写回 payload。

## 批次 A：合同、profile 与 memory 领域链

状态：`completed`。

- 新增 Definition HTTP Tool candidate / version 与 action plan source v2 strict schema。
- candidate create 显式接收 execution profile；服务端根据精确草案和 registry 校验，不从节点类型自动推断授权。
- 抽取共享图资格校验，保证 release 与 plan 使用同一规则。
- 增加 active Definition authority resolver 和 Definition 来源 plan create。
- 覆盖 profile 隔离、review / activation 零副作用、source union、scope、drift、CAS 与既有 Draft 计划兼容测试。

完成门禁：目标 Go 测试、合同校验、race 和仓库 fast 通过；批次状态只写为 `workflow_definition_http_tool_v1_batch_a_completed`。

完成证据：v3 candidate / version、Definition 来源 action plan / confirmation / audit v2、active Definition authority resolver、独立人工确认、pointer 漂移失效和严格 HTTP 路由均已落地；目标测试、目标 race、schema 校验和仓库 fast 已通过，计划 / 审批阶段网络、provider 与 run 均为 0。

## 批次 B：双数据库持久化

状态：`completed`。

- 为 action plan / confirmation / audit 顺序追加 SQLite v20 与 PostgreSQL v23 migration；Definition-bound run schema 随批次 C 的真实执行行为一起增加。
- Definition 来源新写入使用 v2；既有与新建 Draft 计划继续使用 v1 payload，并由数据库规范化 `saved_workflow_draft` 来源列。旧 run v2 继续可读且不可改写。
- 覆盖 migration / rollback / reapply、marker、CAS、restart、corruption、组合权限和 no-fallback。

完成门禁：双数据库专项通过，批次状态写为 `workflow_definition_http_tool_v1_batch_b_completed`。

完成证据：SQLite `0020_workflow_http_tool_definition_sources` 和 PostgreSQL `0023_workflow_http_tool_definition_sources` 已进入既有 migration family；计划、决定和审计的 payload / 规范化来源投影受数据库与 repository 双重约束。SQLite active Definition → plan → approve 已跨服务重启恢复，PostgreSQL 完整开发测试集成套件已通过 migration、runtime role、rollback / reapply、source corruption、重连和 no-fallback；两条路径均未触发网络、provider 或 run。

## 批次 C：执行与只读审查消费者

状态：`completed`。

- 在原子 claim 前重读 Definition authority、Application lifecycle、confirmation 和 plan digest；claim 后、网络前再次复核 Definition 与 Application authority。
- 复用 transport、SSRF、预算、output projection、终态提交与 reconciliation；同一 approved plan 并发只有一个 claim 和一次网络执行。
- 新增 strict `workflow_run_record.v9`，接入 History / detail / diagnostics；Comparison / Evaluation 对有副作用 profile 稳定拒绝且不重新执行。
- SQLite `0021` 与 PostgreSQL `0024` 复用既有 migration family；覆盖 migration、并发 claim、重启、来源过滤、计划消费恢复、authority drift、no-fallback 和隐私扫描。

完成门禁：Platform 全量、目标 race、`go vet`、双数据库产品链和仓库 full 通过。

完成证据：memory、SQLite 与 PostgreSQL 已贯通 active Definition → v2 plan → approve → v9 execute → History；执行权限按计划来源动态要求 `workflow_drafts:read` 或 `workflow_definitions:read`。claim 后 authority 漂移会写入失败 v9 且网络 / provider 为 0；成功链只保存 input digest / bytes、Definition authority、脱敏工具投影、attempt 和副作用 metadata。PostgreSQL 完整开发测试集成套件、迁移启动检查和重启重复执行拒绝已通过，测试容器已关闭。

## 批次 D：React 与产品连续链

状态：`completed`。

- 在 Definition 工作区提供工具型候选、版本、activation、plan、confirm、execute 与 history handoff。
- application / workspace 切换清空易失输入和迟到响应；offline 零请求，strict consumer 拒绝字段和 scope 漂移。
- 完成桌面、中宽、窄屏浏览器复核、重启恢复、Web Storage / URL / console / network 与数据库敏感扫描。
- 同步专题、入口、当前焦点、路线图、能力矩阵和周志并提交。

完成证据：既有 Definition 工作区已接通 Candidate → Review → Version → Activation → Plan → Confirm → Execute → History；plan / decision strict v2 与 run strict v9 均按精确 application / workspace / Definition scope 消费。application / workspace 切换会清空易失参数和迟到响应；刷新及 SQLite 服务重启通过 plan / run 短引用重读同一 consumed plan 与 v9 detail，不重新执行。

内置浏览器已覆盖 `1440×900`、`1024×768`、`390×844`，均无横向溢出且最终控制台 warning / error 为零。服务端登记的 `.invalid` 开发目标按预期产生 `workflow_tool_transport_failed`，只记录一个 attempt、一个 confirmation、零业务写入、零 replay 与零 retry / fallback。会话引用的允许字段、run id 形状和未知字段拒绝由 strict consumer 测试覆盖；数据库与运行记录敏感扫描沿用批次 C 并由本批实际 SQLite 记录复核。

完成锚点：`workflow_definition_http_tool_v1_completed`。本卡归档，不派生批次 E、平行 owner、专用数据库或 readiness 卡。

## 明确不做

- code / sandbox、agent loop、RAG + tool、多工具、并行或循环。
- 自动确认、自动执行、retry / fallback、replay / resume、业务写回。
- production credential、production secret、production auth、public production API、quota、billing 或 SLA。

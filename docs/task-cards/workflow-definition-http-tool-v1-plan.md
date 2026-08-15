# Workflow Definition 绑定受控 HTTP Tool v1 实施任务卡

更新时间：2026-08-15

- 任务 ID：`workflow-definition-http-tool-v1`
- 状态：`workflow_definition_http_tool_v1_batch_b_persistence_next`
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

状态：下一顺位。

- 为 action plan / confirmation / audit / run 顺序追加 SQLite 与 PostgreSQL migration。
- 新写入只使用新 schema；旧 v1 / run v2 继续可读且不可改写。
- 覆盖 migration / rollback / reapply、marker、并发、restart、corruption、权限和 no-fallback。

完成门禁：双数据库专项通过，批次状态写为 `workflow_definition_http_tool_v1_batch_b_completed`。

## 批次 C：执行与只读审查消费者

- 在原子 claim 前重读 Definition authority 和 plan digest。
- 复用 transport、SSRF、预算、output projection、终态提交与 reconciliation。
- 新 run schema 接入 History / detail / diagnostics；Comparison / Evaluation 采用显式支持或稳定拒绝，不能重新执行。
- 覆盖取消、终态写入失败、authority drift、并发 claim、重启和隐私扫描。

完成门禁：Platform 全量、目标 race、`go vet`、双数据库产品链和仓库 full 通过。

## 批次 D：React 与产品连续链

- 在 Definition 工作区提供工具型候选、版本、activation、plan、confirm、execute 与 history handoff。
- application / workspace 切换清空易失输入和迟到响应；offline 零请求，strict consumer 拒绝字段和 scope 漂移。
- 完成桌面、中宽、窄屏浏览器复核、重启恢复、Web Storage / URL / console / network 与数据库敏感扫描。
- 同步专题、入口、当前焦点、路线图、能力矩阵和周志并提交。

完成锚点：`workflow_definition_http_tool_v1_completed`。

## 明确不做

- code / sandbox、agent loop、RAG + tool、多工具、并行或循环。
- 自动确认、自动执行、retry / fallback、replay / resume、业务写回。
- production credential、production secret、production auth、public production API、quota、billing 或 SLA。

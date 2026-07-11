# Workflow Run Comparison / Regression Review v1

更新时间：2026-07-11

状态：`workflow_run_comparison_regression_review_v1_completed`

## 功能目标

让内部开发者从现有 Workspace Run History 选择一个基线运行和一个候选运行，在同一 tenant / workspace / application scope 内审查状态、草案版本、provider / model、失败诊断、节点结果、耗时和零副作用证据的变化。比较只读取 `/v1/user-workspace/workflow-runs` 真相源，不创建运行、不保存比较结果，也不开放 retry、replay、resume、tool、confirmation commit 或业务写回。

## 用户路径

1. 用户在真实 dev / test Run History 中加载并过滤 durable 运行记录。
2. 用户选择基线与候选；两者必须是不同 run id，且都能在当前 scope 读取。
3. Web 调用候选运行的 scoped comparison route，展示确定性分类、关键变化、节点级差异和人工审查建议。
4. 用户仍可打开任一原始详情核对 request id、audit ref 和结构化诊断；比较结果不替代 run record。
5. 默认 offline 模式不发 HTTP，也不把旧 fake `/v1/user-workspace/runs` 混入选择集。

## API 与 scope

```text
GET /v1/user-workspace/workflow-runs/{candidate_run_id}/comparison?baseline_run_id=...&workspace_id=...&application_id=...
```

route 要求 `workflow_runs:read`，使用现有 dev executor gate、身份头、workspace / application 头和 audit 规则。两个记录均通过同一个 run repository 和同一个 `WorkflowRunContext` 读取；任一记录不存在或越界统一返回 `workflow_run_record_not_found`，不泄漏其它 scope 的存在性。相同 run id、空 id、重复或未知 query 返回 `workflow_run_comparison_invalid`。repository 失败返回既有 store failure，绝不回退 memory 成功。

比较是二元定点读取，不使用分页游标。运行选择继续复用既有 list keyset cursor 和过滤条件，comparison route 不引入第二套列表或索引。

## 比较模型与分类

响应使用 `workflow_run_comparison.v1`，包含 baseline / candidate 的安全身份摘要、`classification`、`comparison_state`、状态与诊断变化、草案 / provider / model 变化、总耗时差、节点差异、固定 review findings、推荐人工动作和双方零副作用计数。

`classification` 固定为：

- `regression`：基线成功而候选失败、取消或 stale running；
- `improvement`：基线失败、取消或 stale running，而候选成功；
- `changed`：可审查的终态没有成功性反转，但稳定字段或节点结果发生变化；
- `unchanged`：全部允许比较的稳定字段一致；
- `inconclusive`：任一运行仍在进行且未 stale，无法给出终态判断。

耗时增减只作为观察项，不单独推断回归。`comparison_state` 为 `comparable|legacy_partial|running_inconclusive`；任一 v0 记录缺少 diagnostic 时仍比较状态、选择和节点，但明确标记 `legacy_partial`，不得反推失败边界。

节点按 `node_id` 对齐，返回 `added|removed|changed|unchanged`、节点类型、双方状态和耗时差。响应不包含 label、output、output preview、输入字节、condition node / value、actor ref、credential、endpoint、provider raw envelope 或任意 payload。

## 生命周期、保留、脱敏与可观测性

比较结果不持久化、无独立保留期，不新增 schema migration。它在请求时从已通过既有 30 天 / 每 scope 10,000 条开发测试保留规则管理的 run records 派生；记录过期后返回 not found。服务只记录 outcome、classification、comparison state、双方 run id 的不可逆摘要、scope hash、duration、request id 和 audit ref。

双方 `tool_calls`、`confirmation_calls`、`business_writes`、`replay_writes` 必须全部为 0；否则比较失败为 store contract mismatch，不输出可信结论。provider call 只返回计数差。review finding 和建议动作来自服务端固定枚举，不接受客户端阈值或任意文本。

## Web 与性能边界

Run History 在真实 dev / test 模式提供基线 / 候选选择和比较审查；offline 保持零请求。比较 consumer 与面板放入独立 lazy 模块，不继续扩大 `App.tsx`。本批同时把 React runtime 拆为稳定 vendor chunk，使 R5 主包低于 500 KiB；不新增现有测试、Web build 或聚合门禁无法承载的同层 checker。

## 验收

- Go domain：五种分类、legacy / running、节点 added / removed / changed、耗时和固定 findings。
- HTTP：strict query、auth / scope、同 id、not found、store failure、敏感字段缺失和零副作用 contract。
- repository：memory 并发读取；PostgreSQL 重启后比较、跨 scope not found 和连接失败 no fallback。
- Web：offline 零请求、请求契约、响应验证、禁止字段拒绝、选择与比较详情。
- 浏览器：真实 durable history 选择成功 / 失败记录，完成比较、原始详情核对和服务重启后复验。
- 最终运行 Go test / race / vet、Web test / build、仓库 fast / full 门禁，并关闭本批服务。

## 停止线

- 不持久化 comparison，不新增 comparison table、migration 或长期快照。
- 不根据耗时噪声自动宣称回归，不做自动基线选择或跨 scope 比较。
- 不开放 retry / replay / resume、tool、RAG、confirmation commit、业务写回、publish 或 production enablement。

## 完成结果

- nested comparison route 通过同一 scoped repository 读取双方记录；strict query、同 id、not found、scope、store failure 和零副作用 contract 均 fail closed。
- `workflow_run_comparison.v1` 覆盖五种分类、legacy partial、stale / running 语义、节点差异和固定 findings；耗时变化只作为观察证据。
- comparison 不持久化、不新增 migration；PostgreSQL 集成验证连接池重开后仍可比较 durable records，跨 scope 返回 not found，数据库失败不回退 memory。
- Web 提供显式 baseline / candidate 选择和独立 lazy comparison panel；offline 与同 run 选择零请求，实际响应按 key 递归执行敏感字段检查。
- 真实浏览器完成 saved executor draft、成功运行、Gateway timeout、历史详情与成功 → 失败比较，结果为 `regression / comparable`，provider call delta 为 0，四类禁止副作用为 0。
- R5 将 React runtime 拆为稳定 vendor chunk；主入口由 624.57 KiB 降到 430.39 KiB，comparison chunk 为 6.96 KiB。
- Go test / race / vet、Web 14 项测试 / build、PostgreSQL integration 与真实浏览器均通过；服务、浏览器、容器、网络和 Docker Desktop 已关闭。

下一产品顺位先设计 `Workflow Evaluation Cases / Batch Regression Review v1`：把案例生命周期、显式 baseline、批量候选审查、聚合证据与保留 / 脱敏设计清楚，再决定新 API 和 schema；不得借此打开 replay、tool 或业务写回。

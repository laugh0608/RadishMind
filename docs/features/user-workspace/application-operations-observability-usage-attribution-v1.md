# 应用运行观测与用量归因 v1

更新时间：2026-07-19

状态：`application_operations_observability_usage_attribution_v1_batch_a_completed`

## 功能定位

本专题为内部应用开发者提供应用作用域的运行观测入口，把现有 Model Gateway Request History 与 Workflow Run History 放到同一个用户工作区视图中，回答以下问题：

- 当前应用最近有哪些 northbound Gateway 请求和 Workflow / Application RAG 运行；
- 两个通道分别成功、失败、取消或仍在运行多少条；
- 当前加载窗口内使用了哪些 provider / model，Workflow 实际记录了多少 provider、retrieval 和 tool 调用；
- Gateway usage 是 provider 已报告、未报告还是不适用；
- 哪些失败需要进入既有请求详情、运行详情、Comparison 或 Evaluation 继续审查。

本专题不创建新的运行真相源。Gateway request 与 Workflow run 继续由各自 repository 拥有，应用观测面只读取、校验、聚合和交接 metadata。

## 现状判断

- Gateway Request History 已按 tenant / workspace / consumer / optional application 保存 northbound 请求，列表公开 route、protocol、状态、provider / profile / model、耗时、failure boundary 和 usage availability。
- Workflow Run History 已按 tenant / workspace / application 保存 v0–v4 运行，列表公开 execution source、状态、provider / model、诊断、RAG metadata 和副作用计数。
- 两个资源没有稳定的一对一关联键。Workflow 内部 Gateway bridge request id 与 northbound Gateway history request id 语义不同；首批不得按时间、model、request id 或 audit ref 推测关联。
- Gateway list 不公开 token counts；只有精确 detail 在 `usage.availability=reported` 时公开经验证的 token metadata。首批不得通过逐条 N+1 detail 请求伪造全局 token 汇总。
- 两条 list route 均默认返回 25 条并有独立 cursor。首批聚合只能表示“当前加载窗口”，不能写成应用全部历史或全量用量。

## 目标用户路径

1. 用户在应用目录选择一个应用，进入 Application Operations 面板。
2. 面板以该应用、当前 workspace 和既有可信开发测试态身份并行读取 Gateway request 与 Workflow run 的首个分页窗口。
3. 顶部显示两个通道各自的来源状态、已加载条数和 `has_more`，并明确说明当前摘要是否覆盖全部可见记录。
4. 归因摘要分别展示 Gateway 请求状态 / usage availability，以及 Workflow 状态 / provider、retrieval、tool 调用数；不把两组数字相加为“总调用量”。
5. 合并时间线只按 `started_at` 排序，保留 `gateway_request` / `workflow_run` 来源标签、独立资源 id、provider / model、耗时、失败和审查引用。
6. 一个通道失败时，另一个通道的有效结果仍可审查，并显示 `partial_failure`；两个通道都失败时才进入整体失败状态。
7. 应用切换立即丢弃旧时间线、统计、错误和请求结果；迟到响应不得覆盖新应用。
8. 默认 offline 模式零请求，只解释需要显式开发测试态 source 才能加载真实数据。

## 数据与归因语义

### Gateway 通道

当前加载窗口允许归因：

- loaded request count；
- `started / succeeded / failed / canceled` 数量；
- `reported / not_reported / not_applicable` usage availability 数量；
- provider、profile、model、protocol、route、duration 与 failure boundary 的逐条 metadata。

不得归因 token 总数、价格、成本、quota 消耗或 billing。`not_reported` 不能按零 token 处理。

### Workflow 通道

当前加载窗口允许归因：

- loaded run count；
- `running / succeeded / failed / canceled / outcome_unknown` 数量；
- `provider_calls / retrieval_calls / tool_calls / confirmation_calls` 求和；
- execution kind / source、provider / profile / model、duration、failure boundary 和 review action；
- `business_writes / replay_writes` 停止线计数，非零必须显示为越界而不是隐藏。

Workflow side-effect count 只描述 Workflow run 自身，不能与 Gateway request count 相加，也不能推导 provider 账单。

### 合并时间线

合并时间线条目固定包含：

- `source=gateway_request|workflow_run`；
- source record id、started / completed time、status；
- route / protocol 或 execution kind / schema version；
- provider / profile / model；
- duration、failure code / boundary、request / audit ref；
- Gateway usage availability，或 Workflow side-effect counts。

排序只用于阅读，不建立跨资源因果关系。相同 request id、model 或相邻时间都不能自动写成 related / parent / child。

## 状态模型

整体状态固定为：

- `offline`：两个 source 均未启用，零网络请求；
- `loading`：至少一个已启用通道正在读取；
- `ready`：至少一个通道有记录且没有读取失败；
- `empty`：已启用通道均成功但没有记录；
- `partial_failure`：一个通道失败，另一个通道仍可展示有效结果；
- `failed`：全部已启用通道失败；
- `application_unavailable`：没有可用 application id，零网络请求。

每个通道独立保留 `offline / loading / ready / empty / failed`、request / audit ref、failure code、`has_more` 和已加载条数。

## Web 信息架构

Application Operations 面板包含：

1. 应用与 coverage header：应用名 / id、整体状态、两个 source 状态、加载窗口说明。
2. Attribution summary：Gateway 状态与 usage availability；Workflow 状态与受控调用数；禁止显示虚构总成本。
3. Unified timeline：最多展示当前两个首分页窗口的 50 条 metadata，按时间倒序。
4. Boundary note：两个真相源不自动关联，摘要不是 billing / quota / 全量 usage。
5. Refresh：只刷新当前应用；刷新中旧数据不冒充新快照，错误不回退 offline sample。

面板复用选中应用、Gateway Request History consumer 和 Workflow Run History consumer，不创建第二套 route、身份头、过滤器或 schema decoder。

## 当前实现

- 批次 A 已在用户工作区接入独立 lazy panel，并复用当前 application selection、Gateway Request History 与 Workflow Run History 的严格消费端。
- 两个通道并行读取、独立失败；一个通道失败时保留另一个通道，只有全部已启用通道失败才进入整体失败。应用切换和刷新会丢弃迟到结果。
- 当前加载窗口分别统计 Gateway 状态 / usage availability 与 Workflow 状态 / side effects，并以来源标签合并时间线；相同 request id 仍保留为两条独立记录。
- coverage 只有在全部已启用通道成功且游标耗尽时才声明当前窗口完成；offline、partial failure 和 `has_more=true` 均不会伪装为完整覆盖。
- 本批没有新增 API、schema、migration、repository、task card、fixture 或 checker，也没有启用 token、成本、配额或计费推导。

## 实施批次

### 批次 A：当前窗口聚合与用户工作区接线

- 新增纯聚合 consumer，复用两条严格 list consumer；并行加载、独立失败和应用切换隔离。
- 新增 Application Operations lazy panel、coverage、归因摘要、合并时间线和 refresh。
- 接入当前应用选择；默认 offline 零请求，显式本地产品档复用既有 source。
- 覆盖排序、计数、usage availability、Workflow side effects、`has_more`、partial failure、双失败、空状态和敏感字段边界。
- 执行 Web 单测、production build、`git diff --check` 与仓库快速门禁。

批次 A 不新增 API、schema、migration、repository、task card 或专项 checker。

批次 A 已完成，后续不继续派生同层 UI 汇总批次；只有满足下述服务端 summary 评审条件时，才重新打开本专题的实现范围。

### 后续评审条件

只有出现以下真实需求，才评审后续服务端 summary API：

- 需要跨全部分页窗口的稳定统计；
- provider adapter 已提供可信 `reported` usage，且需要避免 N+1 detail；
- 需要稳定时间桶、游标一致性或大数据量性能预算；
- quota / billing 已有独立正式设计、可信价格和 ledger owner。

满足条件前不创建 aggregate table、materialized view、cost ledger、quota counter 或跨 store join。

## 验收方式

- offline：两个 source 未启用时 fetch 次数为 0。
- scope：所有读取使用当前 application id；应用切换后旧响应不能进入新状态。
- aggregation：两个通道分别计数，合并时间线稳定排序，不生成跨通道关联。
- partial failure：任一通道失败不清除另一通道的有效 metadata。
- coverage：`has_more=true` 时必须显示窗口未覆盖全部记录。
- privacy：不出现 prompt、input、answer、response body、token、credential、header、endpoint、provider raw payload 或 fragment 正文。
- regression：既有 Gateway Request History 与 Workflow Run History 页面、过滤、详情和分页保持不变。

## 停止线

- 不把加载窗口摘要写成全量历史、可信用量、成本、quota 或 billing。
- 不按时间、request id、model 或 audit ref 推测 Gateway request 与 Workflow run 的一对一关系。
- 不保存新的运行副本，不把 UI 聚合状态写回任一 repository。
- 不逐条读取 Gateway detail 计算 token，不估算 token 或价格。
- 不启用生产认证、生产 API key、quota enforcement、billing、外部 connector、自动 retry / fallback、业务写回或 replay。
- 不为普通 UI 聚合新增 task card、fixture 或 checker；现有 Web 测试、build 和仓库门禁足以承载批次 A。

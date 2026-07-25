# Prompt / Agent 应用回归评测与发布审查（开发 / 测试态）v1

更新时间：2026-07-25

状态：`prompt_agent_application_regression_evaluation_release_review_dev_test_v1_completed`

## 功能定位

本专题让内部应用开发者把 Prompt Application v6 与 Agent / Copilot v7 的 durable Run 纳入既有 Evaluation Case、Evaluation Suite 和人工 Release Decision 链，在版本继续演进前回答：

- 同一 Prompt Template lineage 的两次受控调用是否发生状态、authority、变量契约、协议、provider、model 或 usage metadata 变化；
- 同一 Agent Profile、project 与 task 的两次受控建议是否发生状态、authority、策略摘要、response metadata、provider 或 model 变化；
- 一组明确预期能否形成可复验 case，再进入 exact case-version suite 与人工发布审查证据。

本专题不创建第二套应用评测真相源。Run Comparison 继续拥有运行比较语义，Workflow Evaluation Case / Suite 继续拥有用例、版本、聚合 review 和 decision evidence；本专题只补齐它们对既有 v6 / v7 metadata profile 的严格兼容和用户入口。

## 根因与选择

- `workflow_run_comparison.v5` 已支持 `prompt_application_invocation_v1`，`workflow_run_comparison.v6` 已支持 `agent_copilot_suggestion_v1`。
- Go Evaluation Case 创建路径已经按 Prompt Template lineage、Agent Profile / project / task 进行可比性校验，但 review 的 Agent 不兼容失败映射不完整。
- Web Evaluation Case 与 Suite 严格 decoder 仍只接受 comparison v1–v4 和旧 run profile，导致合法 v5 / v6 review 被客户端拒绝。
- Evaluation Run selector 已能收到 v6 / v7 summary，但显示标签没有完整区分 Agent v7。

因此首批不新增 route、schema、migration、repository 或执行器，只修复同一 owner 内的版本支持矩阵、失败语义、严格消费和产品标签，并通过真实 SQLite 浏览器链证明 case → suite → decision 可用。

## 目标用户路径

1. 用户在同一应用下产生至少两个同 lineage 的 Prompt Application v6 或 Agent / Copilot v7 terminal Run。
2. 用户在 Run History 的 Evaluation Cases 中选择 baseline 与一个或多个兼容 candidate，并为每个 candidate 指定预期分类。
3. 服务端重读 durable Run，校验 scope、终态、零禁止副作用和 profile-specific lineage，创建或修订既有 `workflow_evaluation_case.v2`。
4. 用户打开 case review；服务即时复用 Comparison v5 / v6，返回实际分类、finding、run profile 与推荐审查动作。
5. 用户把一个或多个 exact case version 组成既有 `workflow_evaluation_suite.v1`，读取稳定 review digest。
6. 用户追加 `approved | rejected | needs_review` decision evidence。批准只表示开发测试态审查结论，不触发 candidate、assignment、release、deploy 或 provider 调用。

## 领域所有权

| 领域 | 既有 owner | 本专题行为 |
| --- | --- | --- |
| Prompt / Agent 运行 | Workflow Run Store v6 / v7 | 只读取 metadata-only terminal Run |
| 单次比较 | Workflow Run Comparison v5 / v6 | 继续负责分类、finding 与兼容失败 |
| Evaluation Case | Workflow Evaluation owner | 继续保存 run refs、预期、revision 与审计 metadata |
| Evaluation Suite | Workflow Evaluation Suite owner | 继续保存 exact case-version refs、review digest 与 decision evidence |
| Web | Run History 的既有 lazy panels | 增加严格版本支持与清晰类型标签 |
| 发布 / assignment | Application Publish / Runtime Assignment owner | 不被 Evaluation decision 自动修改 |

## Profile 兼容规则

### Prompt Application

baseline 与 candidate 必须同时满足：

- `schema_version=workflow_run_record.v6`；
- `execution_profile=prompt_application_invocation_v1`；
- `execution_source.kind=prompt_application_template`；
- `execution_source.id` 相同。

Template version、authority digest、变量契约、协议、provider、model、状态、失败或 usage metadata 可以变化，并由 Comparison v5 显式分类与形成 finding。跨 Template lineage 必须返回 `prompt_application_execution_profile_incompatible`。

### Agent / Copilot

baseline 与 candidate 必须同时满足：

- `schema_version=workflow_run_record.v7`；
- `execution_profile=agent_copilot_suggestion_v1`；
- `execution_source.kind=agent_copilot_profile`；
- `execution_source.id`、`project` 与 `task` 相同。

Profile version、authority / profile / policy digest、locale、response status / digest、action count、risk、confirmation、provider、model、状态或失败可以变化，并由 Comparison v6 显式分类与形成 finding。跨 Profile、project 或 task 必须返回 `agent_copilot_execution_profile_incompatible`。

### 通用资格

- 所有 Run 必须位于同一 tenant / workspace / application scope。
- Run 必须为 terminal 或超过既有 stale budget 的 running record。
- v6 / v7 只允许一次 provider side effect，tool、retrieval、confirmation、business write 与 replay 必须为 0。
- Prompt 与 Agent 不能互相比较，也不能与 v0–v5 profile 混合。
- comparison / evaluation 只读取 metadata，不恢复输入、变量值、context、artifact content、prompt、answer、provider raw response 或 transcript。

## Case、Suite 与 decision 语义

- Case 继续使用 `workflow_evaluation_case.v2`，不增加应用私有字段。
- Review item 允许 `workflow_run_comparison.v5 | workflow_run_comparison.v6`，对应 run profile 必须精确匹配。
- Suite 可混合包含 Workflow、RAG、Prompt 与 Agent 的 exact case versions；每个 item 保留自己的 run profile，canonical review digest 已包含 profile。
- `approved` 仍只在 suite review `passed` 且 unavailable 为 0 时允许。
- decision 是 append-only review evidence，不授权发布、activation、执行、重放或业务写回。

## 数据与隐私

新增支持不改变持久化形态。Case、Suite 与 decision 仍只保存 run / case 引用、预期分类、聚合计数、review digest、actor / request / audit ref 和时间。

以下材料不得进入 evaluation API、Web state 的持久介质、数据库、日志或 committed 证据：

- Prompt 模板正文、变量值、渲染结果或完整回答；
- Agent context、artifact content、完整 `CopilotResponse` 或 proposed action payload；
- credential、token、header、cookie、endpoint、DSN 或 provider raw payload；
- 根据 metadata 猜测或重建的输入、输出或跨资源因果关系。

## 实施批次

### 批次 A：兼容矩阵、严格消费与相邻测试

- 补齐 Go review 对 Agent 不兼容失败的稳定映射和 summary。
- 为 Prompt v6 / Agent v7 增加 Case create / review、跨 profile、跨 lineage、project / task drift 和零副作用负向测试。
- 扩展 Web Case / Suite decoder，只接受 Comparison v5 + Prompt profile、Comparison v6 + Agent profile 的精确组合。
- 补齐 Agent v7 Run selector 标签和非法 schema / profile 组合拒绝测试。

### 批次 B：真实用户链与专题关闭

- 使用现有 SQLite 本地产品档产生同 lineage Agent v7 Run。
- 在真实浏览器完成 case create / review → suite create / review → human decision。
- 验证页面离开、刷新与 browser storage 不保留输入、回答或凭证。
- 同步功能入口、当前焦点、路线图、能力矩阵和周志，关闭专题。

## 验收方式

- Go：Evaluation domain / HTTP / Suite 相邻测试、跨 profile / lineage 负向、定向 race 和 `go vet`。
- Web：Case / Suite strict consumer、Run History compatibility 与 build。
- 浏览器：SQLite Agent v7 case、suite、decision、Run detail 与隐私清理。
- 仓库：`git diff --check`、`./scripts/check-repo.sh --fast`；阶段真相源变化后运行全量 `./scripts/check-repo.sh`。

## 完成证据

1. Go Evaluation review 已补齐 Agent Profile / project / task 不兼容失败映射；相邻测试覆盖合法 Prompt / Agent case 与 suite、跨 task 漂移、review 失败摘要和人工批准策略。
2. Web Case strict consumer 只接受 Comparison v5 + Prompt profile、Comparison v6 + Agent profile 的精确组合；Suite strict consumer 接受两种新 profile，未知 profile 与错配组合继续失败关闭。
3. Run History 已显式标识 Prompt v6 与 Agent v7；Web 229 项测试、生产构建、Go 相邻测试、定向 race 和 `go vet` 均通过。
4. SQLite 真实浏览器使用同一 Agent Profile / project / task 产生两个 v7 Run，Comparison v6 得到 `changed`，创建 `eval_46be5143a56cadd4d15b18e0` 并 review 为 `passed / 1 matched`。
5. exact case version 组成 `suite_a40a6cf0a739a3f27f332eb9`，suite review 为 `passed · agent_copilot_suggestion_v1`，随后记录 `approved v1`；该 decision 没有修改 candidate、assignment、release 或 provider 状态。
6. 浏览器 `localStorage` 与 `sessionStorage` 均为空，console 0 error；开发服务在验收后关闭，没有保留后台进程。

## 停止线

- 不新增第二套 evaluation case、suite、decision、comparison 或 run profile registry。
- 不重新执行、批量执行、retry、fallback、replay 或 resume Run。
- 不把评测 decision 连接为自动 candidate approve、assignment activation、release 或 deploy。
- 不保存或恢复 Prompt / Agent 输入输出正文。
- 不启用生产认证、生产 API key、quota、billing、cost ledger、外部 connector、在线搜索、工具执行或业务写回。
- 不为普通 UI 标签和已有 owner 兼容扩展新增专项 checker 或 fixture。

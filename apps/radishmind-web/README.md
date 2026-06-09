# RadishMind Web

`apps/radishmind-web/` 是正式产品 UI 的首个落点。当前承载两组离线只读 surface：

- `Control Plane Read-Side`：`control-plane-read-shared-shell-v1`、`control-plane-read-admin-tenant-overview-v1`、`control-plane-read-admin-audit-log-v1`、`control-plane-read-workspace-applications-v1`、`control-plane-read-workspace-api-keys-v1`、`control-plane-read-workspace-usage-quota-v1`、`control-plane-read-workspace-workflow-definitions-v1`、`control-plane-read-workspace-run-history-v1`、`control-plane-read-formal-ui-readiness-close-v1`、`control-plane-read-dev-live-consumer-v1` 和 `control-plane-read-auth-store-transition-preconditions-v1`。
- `Workflow / Agent Runtime Function Surface`：`workflow-function-surface-boundary-v1`、`workflow-application-detail-read-v1`、`workflow-definition-detail-read-v1`、`workflow-run-detail-read-v1`、`workflow-blocked-action-preview-v1`、`workflow-confirmation-placeholder-read-v1`、`workflow-draft-designer-offline-v1`、`workflow-draft-validation-inspector-offline-v1`、`workflow-execution-plan-preview-offline-v1`、`workflow-runtime-readiness-inspector-offline-v1`、`workflow-function-surface-readiness-close-v1`、`workflow-workspace-context-consistency-v1`、`workflow-workspace-review-offline-v1`、`workflow-user-workspace-home-offline-v1` 和普通离线 Workflow Review Handoff。

当前边界：

- 默认只消费 `contracts/typescript/control-plane-read-api.ts` 的离线 read-side contract。
- 前端离线 view model 默认数据必须与 `control-plane-read-response-fixtures-v1` 的 RadishFlow Copilot / Radish Docs Assistant success 样例保持一致；`control-plane-read-product-sample-consistency-v1` 会校验该 response fixture、Go fake store、consumer smoke product refs 和前端离线默认 envelope 没有漂移。
- 当显式设置 `VITE_RADISHMIND_READ_SOURCE=dev-live-http` 时，可通过 dev-only HTTP consumer 消费 fake-store-backed read handlers；后端必须同时设置 `RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1` 才会接受测试身份 header。
- 只渲染 read route catalog、共享状态组件、forbidden output guard、只读 `admin-tenant-overview`、只读 `admin-audit-log`、只读 `workspace-applications`、只读 `workspace-api-keys`、只读 `workspace-usage-quota`、只读 `workspace-workflow-definitions`、只读 `workspace-run-history`、User Workspace Home 和 workflow function surface 面板。
- `admin-tenant-overview` 只消费 `tenant-summary-route` 的离线 view model，展示租户摘要、route metadata、request / audit ref 和状态预览。
- `admin-audit-log` 只消费 `audit-summary-list-route` 的离线 view model，展示 audit ref、actor、event kind、resource、decision、failure code、trace id、recorded timestamp、route metadata、request / audit ref、cursor 和状态预览。
- `workspace-applications` 只消费 `application-summary-list-route` 的离线 view model，展示应用摘要列表、cursor、route metadata、request / audit ref 和状态预览。
- `workspace-api-keys` 只消费 `api-key-summary-list-route` 的离线 view model，展示 API key id、owner、scope、state、时间字段、route metadata、request / audit ref 和状态预览，不展示 key value 或 hash。
- `workspace-usage-quota` 只消费 `quota-summary-route` 的离线 view model，展示 quota id、period、request / token / cost limit、usage snapshot、over quota failure code、route metadata、request / audit ref 和状态预览。
- `workspace-workflow-definitions` 只消费 `workflow-definition-summary-list-route` 的离线 view model，展示 workflow definition id、application ref、version、definition status、node count、risk level、requires confirmation capable、updated at、route metadata、request / audit ref 和状态预览。
- `workspace-run-history` 只消费 `run-record-summary-list-route` 的离线 view model，展示 run id、workflow definition ref、application ref、status、failure code、cost summary、trace id、started / completed timestamp、route metadata、request / audit ref、cursor 和状态预览。
- workflow application detail 由 `workspace-applications` summary 和离线 enrichment 派生，展示 application identity、tenant / owner ref、application type、status、provider profile、risk summary、latest run ref、route / request / audit metadata 和 blocked capability preview。
- workflow definition detail 由 `workspace-workflow-definitions` summary 派生，展示 definition identity、application ref、version、nodes、edges、input / output summary、risk summary、blocked action preview 和 audit metadata。
- workflow run detail 由 `workspace-run-history` summary 派生，展示 run identity、state timeline、cost / token snapshot、trace / failure / audit metadata、blocked replay / result preview 和 request / route metadata。
- workflow blocked action preview 与 confirmation placeholder 只展示未来动作和确认流的形状、风险、human review requirement、missing prerequisites 和 audit trail，不提供 decision submit、approve、reject、defer 或 execution unlock。
- workflow draft designer、draft validation inspector、execution plan preview 与 runtime readiness inspector 是 offline-only inspection surface，只允许在本地查看 draft template、node / edge、readiness、risk、structural checks、contract checks、stage order、node-to-stage mapping、provider/profile requirements、confirmation/audit gates、runtime prerequisites、readiness blockers、implementation gates 和 blocked capability checks，不持久化 draft、validation result、execution plan 或 runtime readiness result。
- `workflowWorkspaceContext` 是 workflow 离线组合层的共享入口，统一解析 application、workflow definition、run、draft 和 scenario selection，并统一构建 detail、blocked action、confirmation placeholder、draft validation、execution plan、runtime readiness、surface overview、scenario inspector、review workspace、User Workspace Home 和 Review Handoff；`workflow-workspace-context-consistency-v1` 会校验 App 不重新手拼这些派生链路。
- Workflow Surface Overview 是普通离线只读总览区域，复用 workflow application detail、definition detail、run detail、selected draft、validation inspector、execution plan preview 和 runtime readiness inspector view model，把 application、definition、draft、validation、plan、readiness、latest run 和 blocked capability 的关系集中展示；它不新增专项 gate，不请求 live backend，不新增 Go route，不创建持久化结果。
- Workflow workspace context selection 允许用户在本地选择 application、workflow definition、run record 和 draft template，并让 detail、draft validation、execution plan、runtime readiness、blocked action preview、confirmation placeholder、scenario inspector、review workspace 与 overview 随当前上下文联动；该选择只改变浏览器内查看状态，不保存、不发布、不执行、不提交确认、不写回业务数据。
- Workflow Scenario Inspector 是普通离线只读场景检查区域，复用当前 application、definition、run、draft、validation、execution plan、runtime readiness 和 overview view model，展示 RadishFlow / Radish Docs 场景 intent、input contract、expected advisory output、risk / confirmation requirement、relationship map、blocked reason 和 stop lines；场景选择只改变浏览器内查看状态，不保存、不发布、不执行、不请求 live backend、不写回业务数据。
- Workflow Review Workspace 是正式离线只读审查工作区，复用当前 application、workflow definition、run、draft、validation、execution plan、runtime readiness、overview 和 scenario inspector view model，把当前选中 application + definition + run + draft + scenario 的上下文、review stage、关系链、blocked capability rollup 和 stop line rollup 集中展示；它不新增专项 gate，不请求 live backend，不新增 Go route，不创建 review / draft / plan / readiness 持久化结果，不提供保存、发布、执行、确认提交、写回或 replay 控件。
- User Workspace Home 是正式离线只读首页，复用 applications、workflow definitions、run history、API key / quota summary、Workflow Review Workspace、Workflow Surface Overview 和 Scenario Inspector 的 view model，集中展示应用组合、审查路径、最近 run、优先 readiness、主要 route evidence、blocked capability 和关键 stop line rollup；完整 selected context、关系链和停止线明细继续由 Workflow Review Workspace、Workflow Surface Overview 和 Scenario Inspector 承接。它不新增专项 gate，不请求 live backend，不新增 Go route，不创建首页持久化结果，不提供保存、发布、执行、确认提交、写回或 replay 控件。
- Workflow Review Handoff 是普通离线只读审查交接摘要，复用 User Workspace Home、Workflow Review Workspace、Workflow Surface Overview、Scenario Inspector、Runtime Readiness、Blocked Action Preview 和 Confirmation Placeholder 的 view model，集中展示 review recipients、key findings、read-side evidence checklist、decision blockers 和 boundary locks。它不导出、不发送、不保存 handoff，不请求 live backend，不新增 Go route，不提交 confirmation decision，不解锁执行，不写回或 replay。
- `control-plane-read-formal-ui-readiness-close-v1` 已用聚合 surface matrix / checker 固定七个只读页面的 route binding、状态预览、request / audit ref 和 forbidden output guard；后续普通只读展示页不再默认逐页新增专项门禁。
- `workflow-function-surface-readiness-close-v1` 已用 workflow surface matrix / checker 固定当前 workflow 离线产品面的 builder、render anchor、CSS selector、关闭项和停止线；后续普通离线 workflow 展示，包括 Workflow Review Workspace、User Workspace Home 和 Workflow Review Handoff，优先复用该聚合 gate、`npm run build` 和 fast baseline。
- 不请求生产后端，不接 `Radish` OIDC，不接数据库，不实现 API key lifecycle、quota enforcement、rate limit、billing、cost ledger、workflow builder mutation、draft persistence、validation result persistence、execution plan persistence、runtime readiness persistence、publish、workflow executor、confirmation decision、execution unlock、writeback、run replay、run resume 或 repository adapter。
- `control-plane-read-dev-live-consumer-v1` 只能连接 fake-store-backed handler 和测试身份上下文；不得解释为 production API consumer、真实 auth/db、repository、API key / quota 或 workflow executor ready。
- `control-plane-read-auth-store-transition-preconditions-v1` 只固定未来 auth middleware / read store repository 迁移前置条件；不得解释为 Radish OIDC ready、token validation ready、database ready、repository implementation ready 或 production admin console ready。
- 不替代 `apps/radishmind-console/`；后者仍是本地 ops surface。

User Workspace Home / Workflow Review Workspace 读法：

- 左侧导航按 `Workspace`、`Workflow Review`、`Admin` 和 `Contract` 分组；用户端工作区入口和 workflow 审查入口不再与管理端、契约 guard 混排。
- 先看 User Workspace Home，确认应用组合、审查路径、最近 run、优先 readiness、主要 route evidence 和关键 stop line rollup。
- 先使用本地 context selection 选择 application、workflow definition、run record 和 draft template，再选择要审查的 scenario；blocked action preview 和 confirmation placeholder 应跟随当前 run / workflow definition，而不是停留在默认 fixture。
- 先看 selected context 和 review stage，确认当前审查对象；再按 scenario、draft、validation、execution plan、runtime readiness 的顺序阅读证据链。
- 最后看 Review Handoff，确认给人工审查的 recipients、key findings、evidence checklist、decision blockers 和 boundary locks 已经与当前选中上下文一致。
- blocked capability rollup 和 stop line rollup 是审查结论入口，只解释为什么当前不能 publish / execute / confirm / writeback / replay，不提供解锁或提交按钮。

本地启动从仓库根目录执行：

```bash
./start.sh web-live
./start.sh web-offline
pwsh ./start.ps1 -Command web-live
```

`web-live` 会启动或复用 platform 后端和 `apps/radishmind-web/` 前端，并集中设置 dev-only live read 所需的本地环境变量。它只连接 fake-store-backed handler 和测试身份上下文，不代表 production API consumer、真实数据库、Radish OIDC、repository adapter 或 workflow executor ready。

底层 wrapper 也可单独执行：

```bash
./scripts/run-radishmind-web-dev.sh --mode dev-live
./scripts/run-radishmind-web-dev.sh --mode offline
```

常规包命令仍保留在 `apps/radishmind-web/` 下：

```bash
cd apps/radishmind-web
npm run build
npm run preview
```

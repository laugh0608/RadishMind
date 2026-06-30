# RadishMind Web

`apps/radishmind-web/` 是正式产品 UI 的首个落点。当前承载三组离线优先 / dev-only product surface：

- `Control Plane Read-Side`：`control-plane-read-shared-shell-v1`、`control-plane-read-admin-tenant-overview-v1`、`control-plane-read-admin-audit-log-v1`、普通离线 Admin Operations Review / Readiness、普通离线 Admin Provider/Profile & Deployment Evidence Review / Readiness、`control-plane-read-workspace-applications-v1`、`control-plane-read-workspace-api-keys-v1`、`control-plane-read-workspace-usage-quota-v1`、`control-plane-read-workspace-workflow-definitions-v1`、`control-plane-read-workspace-run-history-v1`、`control-plane-read-formal-ui-readiness-close-v1`、`control-plane-read-dev-live-consumer-v1` 和 `control-plane-read-auth-store-transition-preconditions-v1`。
- `Model Gateway / API Distribution`：普通离线 Model Gateway Overview、Route Evidence、Usage/Audit Evidence 与 Evidence Review / Readiness，复用 shared read shell、API key summary、quota summary、run history、audit log、provider runtime、`gateway-api-key-quota-readiness` 和前三个网关 view model 证据，展示 northbound API compatibility surfaces、provider/profile inventory、route binding、selection cases、key scope、quota / cost snapshot、trace / failure、audit decision、readiness rollup、evidence checklist、route / usage / audit key risks 和 locked distribution capabilities。
- `Workflow / Agent Runtime Function Surface`：`workflow-function-surface-boundary-v1`、`workflow-application-detail-read-v1`、`workflow-definition-detail-read-v1`、`workflow-run-detail-read-v1`、`workflow-blocked-action-preview-v1`、`workflow-confirmation-placeholder-read-v1`、`workflow-draft-designer-offline-v1`、`workflow-draft-validation-inspector-offline-v1`、`workflow-execution-plan-preview-offline-v1`、`workflow-runtime-readiness-inspector-offline-v1`、`workflow-function-surface-readiness-close-v1`、`workflow-workspace-context-consistency-v1`、`workflow-workspace-review-offline-v1`、`workflow-user-workspace-home-offline-v1` 和普通离线 Workflow Review Handoff。

当前边界：

- 默认只消费 `contracts/typescript/control-plane-read-api.ts` 的离线 read-side contract。
- 前端离线 view model 默认数据必须与 `control-plane-read-response-fixtures-v1` 的 RadishFlow Copilot / Radish Docs Assistant success 样例保持一致；`control-plane-read-product-sample-consistency-v1` 会校验该 response fixture、Go fake store、consumer smoke product refs 和前端离线默认 envelope 没有漂移。
- 当显式设置 `VITE_RADISHMIND_READ_SOURCE=dev-live-http` 时，可通过 dev-only HTTP consumer 消费 fake-store-backed read handlers；后端必须同时设置 `RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1` 才会接受测试身份 header。
- Workflow saved draft consumer 独立使用 `VITE_RADISHMIND_WORKFLOW_SAVED_DRAFT_SOURCE=dev-saved-draft-http` 开关；默认仍是 sample-only，显式启用后通过 `POST /v1/user-workspace/workflow-drafts`、`GET /v1/user-workspace/workflow-drafts`、`GET /v1/user-workspace/workflow-drafts/{draft_id}` 和 `POST /v1/user-workspace/workflow-drafts/validate` 连接 platform memory dev store。后端仍必须设置 `RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1` 与 `RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP=1`，保存还需要 `RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE=1`。
- 只渲染 read route catalog、共享状态组件、forbidden output guard、只读 `admin-tenant-overview`、只读 `admin-audit-log`、普通离线 Admin Operations Review / Readiness、普通离线 Admin Provider/Profile & Deployment Evidence Review / Readiness、只读 `workspace-applications`、只读 `workspace-api-keys`、只读 `workspace-usage-quota`、只读 `workspace-workflow-definitions`、只读 `workspace-run-history`、User Workspace Home、Model Gateway Overview、Model Gateway Route Evidence、Model Gateway Usage/Audit Evidence、Model Gateway Evidence Review / Readiness 和 workflow function surface 面板。
- `admin-tenant-overview` 只消费 `tenant-summary-route` 的离线 view model，展示租户摘要、route metadata、request / audit ref 和状态预览。
- `admin-audit-log` 只消费 `audit-summary-list-route` 的离线 view model，展示 audit ref、actor、event kind、resource、decision、failure code、trace id、recorded timestamp、route metadata、request / audit ref、cursor 和状态预览。
- Admin Operations Review / Readiness 是普通离线只读管理端汇总面，复用 tenant overview、audit log、Model Gateway Evidence Review 和 Production Ops 静态证据，集中展示 readiness rollup、evidence checklist、operational risks 和 boundary locks；它不新增专项 gate，不请求 live backend，不接数据库、Radish auth、repository adapter、production secret resolver 或 production gateway，不提供 tenant mutation、raw audit export、API key lifecycle、quota enforcement、cost record write、deployment preflight、workflow executor、writeback 或 replay。
- Admin Provider/Profile & Deployment Evidence Review / Readiness 是普通离线只读管理端证据组织面，复用 Model Gateway Route Evidence、Model Gateway Evidence Review、Admin Operations Review、tenant overview 和 audit log，集中展示 provider/profile readiness、model route readiness、secret / deployment evidence、operator risks 和 locked capabilities；它不新增专项 gate，不请求 live backend，不执行 provider call，不接 production gateway、production secret resolver、数据库、Radish auth 或 repository adapter，不提供 provider/profile mutation、model route change、deployment preflight、workflow executor、writeback 或 replay。
- `workspace-applications` 只消费 `application-summary-list-route` 的离线 view model，展示应用摘要列表、cursor、route metadata、request / audit ref 和状态预览。
- `workspace-api-keys` 只消费 `api-key-summary-list-route` 的离线 view model，展示 API key id、owner、scope、state、时间字段、route metadata、request / audit ref 和状态预览，不展示 key value 或 hash。
- `workspace-usage-quota` 只消费 `quota-summary-route` 的离线 view model，展示 quota id、period、request / token / cost limit、usage snapshot、over quota failure code、route metadata、request / audit ref 和状态预览。
- `workspace-workflow-definitions` 只消费 `workflow-definition-summary-list-route` 的离线 view model，展示 workflow definition id、application ref、version、definition status、node count、risk level、requires confirmation capable、updated at、route metadata、request / audit ref 和状态预览。
- `workspace-run-history` 只消费 `run-record-summary-list-route` 的离线 view model，展示 run id、workflow definition ref、application ref、status、failure code、cost summary、trace id、started / completed timestamp、route metadata、request / audit ref、cursor 和状态预览。
- workflow application detail 由 `workspace-applications` summary 和离线 enrichment 派生，展示 application identity、tenant / owner ref、application type、status、provider profile、risk summary、latest run ref、route / request / audit metadata 和 blocked capability preview。
- workflow definition detail 由 `workspace-workflow-definitions` summary 派生，展示 definition identity、application ref、version、nodes、edges、input / output summary、risk summary、blocked action preview 和 audit metadata。
- workflow run detail 由 `workspace-run-history` summary 派生，展示 run identity、state timeline、cost / token snapshot、trace / failure / audit metadata、blocked replay / result preview 和 request / route metadata。
- workflow blocked action preview 与 confirmation placeholder 只展示未来动作和确认流的形状、风险、human review requirement、missing prerequisites 和 audit trail，不提供 decision submit、approve、reject、defer 或 execution unlock。
- workflow draft designer、draft validation inspector、execution plan preview 与 runtime readiness inspector 默认是 offline inspection surface；Draft Designer 现在支持草案名称、说明、节点名称、边条件摘要、本地结构和节点属性的受控编辑，并可在显式 dev-only saved draft 配置下 validate / save / read / restore。该保存只写 platform memory dev store，用于 sample / local / saved / failed / version conflict 状态区分，不代表 durable draft persistence、production API、publish、run 或 executor ready。
- `workflowWorkspaceContext` 是 workflow 离线组合层的共享入口，统一解析 application、workflow definition、run、draft 和 scenario selection，并统一构建 detail、blocked action、confirmation placeholder、draft validation、execution plan、runtime readiness、surface overview、scenario inspector、review workspace、User Workspace Home 和 Review Handoff；`workflow-workspace-context-consistency-v1` 会校验 App 不重新手拼这些派生链路。
- Product Surface Usage Gap Triage 已由 `product-surface-usage-gap-triage-v1` 固定：User Workspace、Workflow Review、Model Gateway 和 Admin 的使用走查只允许在发现真实阅读缺口后修正现有 view model、canonical fixture、文案、导航分组或文档读法，不新增同层产品面，不打开实现入口。
- Workflow Surface Overview 是普通离线只读总览区域，复用 workflow application detail、definition detail、run detail、selected draft、validation inspector、execution plan preview 和 runtime readiness inspector view model，把 application、definition、draft、validation、plan、readiness、latest run 和 blocked capability 的关系集中展示；它不新增专项 gate，不请求 live backend，不新增 Go route，不创建持久化结果。
- Workflow workspace context selection 允许用户在本地选择 application、workflow definition、run record 和 draft template，并让 detail、draft validation、execution plan、runtime readiness、blocked action preview、confirmation placeholder、scenario inspector、review workspace 与 overview 随当前上下文联动；该选择只改变浏览器内查看状态，不保存、不发布、不执行、不提交确认、不写回业务数据。
- Workflow Scenario Inspector 是普通离线只读场景检查区域，复用当前 application、definition、run、draft、validation、execution plan、runtime readiness 和 overview view model，展示 RadishFlow / Radish Docs 场景 intent、input contract、expected advisory output、risk / confirmation requirement、relationship map、blocked reason 和 stop lines；场景选择只改变浏览器内查看状态，不保存、不发布、不执行、不请求 live backend、不写回业务数据。
- Workflow Review Workspace 是正式离线只读审查工作区，复用当前 application、workflow definition、run、draft、validation、execution plan、runtime readiness、overview 和 scenario inspector view model，把当前选中 application + definition + run + draft + scenario 的上下文、review stage、关系链、blocked capability rollup 和 stop line rollup 集中展示；它不新增专项 gate，不请求 live backend，不新增 Go route，不创建 review / draft / plan / readiness 持久化结果，不提供保存、发布、执行、确认提交、写回或 replay 控件。
- User Workspace Home 是正式离线优先首页，复用 applications、workflow definitions、run history、API key / quota summary、Workflow Review Workspace、Workflow Surface Overview 和 Scenario Inspector 的 view model，集中展示应用组合、审查路径、最近 run、优先 readiness、主要 route evidence、blocked capability 和关键 stop line rollup；当前还提供本地 `Create draft` 入口、saved dev draft list、empty / failure state、refresh 和 restore。创建动作只写浏览器本地状态，saved list 只显示 sanitized summary / metadata，恢复动作通过既有 read route 回到 Draft Designer；不创建首页持久化结果，不提供发布、执行、确认提交、写回或 replay 控件。
- Model Gateway Overview、Route Evidence、Usage/Audit Evidence 和 Evidence Review / Readiness 是普通离线只读网关证据区，复用 read shell、API key、quota、run history、audit、provider/gateway readiness 和前三个网关 view model 证据，展示 `/v1/models`、`/v1/models/{id}`、`/v1/chat/completions`、`/v1/responses`、`/v1/messages` 的 northbound API distribution 关系，以及 provider/profile credential state、deployment mode、auth mode、streaming、route metadata、selection cases、key scope、quota policy、cost snapshot、trace records、failure code、audit decision、readiness rollup、evidence checklist、route / usage / audit key risks 和 boundary locks；它们不新增专项 gate，不请求 production gateway，不发放或验证真实 API key，不执行 quota / rate limit，不写 cost record，不解析 production secret，不启用 retry/fallback execution，不接数据库、Radish auth 或 repository adapter。
- Workflow Review Handoff 是普通离线只读审查交接摘要，复用 User Workspace Home、Workflow Review Workspace、Workflow Surface Overview、Scenario Inspector、Validation Inspector、Execution Plan Preview、Runtime Readiness、Blocked Action Preview 和 Confirmation Placeholder 的 view model，集中展示 active draft review record、review recipients、key findings、read-side evidence checklist、decision blockers 和 boundary locks。它不导出、不发送、不保存 handoff，不请求 live backend，不新增 Go route，不提交 confirmation decision，不解锁执行，不写回或 replay。
- `control-plane-read-formal-ui-readiness-close-v1` 已用聚合 surface matrix / checker 固定七个只读页面的 route binding、状态预览、request / audit ref 和 forbidden output guard；后续普通只读展示页不再默认逐页新增专项门禁。
- `workflow-function-surface-readiness-close-v1` 已用 workflow surface matrix / checker 固定当前 workflow 离线产品面的 builder、render anchor、CSS selector、关闭项和停止线；后续普通离线 workflow 展示，包括 Workflow Review Workspace、User Workspace Home 和 Workflow Review Handoff，优先复用该聚合 gate、`npm run build` 和 fast baseline。
- 不请求生产后端，不接 `Radish` auth，不接数据库，不实现 API key lifecycle、quota enforcement、rate limit、cost record writes、完整 workflow builder mutation、durable draft persistence、validation result persistence、execution plan persistence、runtime readiness persistence、publish、workflow executor、confirmation decision、execution unlock、writeback、run replay、run resume 或 repository adapter。
- `control-plane-read-dev-live-consumer-v1` 只能连接 fake-store-backed handler 和测试身份上下文；不得解释为 production API consumer、真实 auth/db、repository、API key / quota 或 workflow executor ready。
- `control-plane-read-auth-store-transition-preconditions-v1` 只固定未来 auth middleware / read store repository 迁移前置条件；不得解释为 Radish auth ready、token validation ready、database ready、repository implementation ready 或 production admin console ready。
- 不替代 `apps/radishmind-console/`；后者仍是本地 ops surface。

源码组织与派生关系：

- `modelGatewayOverview.ts` / `modelGatewayOverviewPanel.tsx` 是模型网关证据根视图，汇总 northbound API surface、provider/profile inventory、route metadata 和 gateway readiness 证据。
- `modelGatewayRouteEvidence.ts` / `modelGatewayRouteEvidencePanel.tsx` 只从 Overview 与 read shell 派生 route binding、selection case、streaming、auth mode、secret ref 和 route risk，不创建新的产品真相源。
- `modelGatewayUsageAuditEvidence.ts` / `modelGatewayUsageAuditEvidencePanel.tsx` 只从 Overview、Route Evidence、API key、quota、run history 和 audit log 派生 usage / audit 证据，不执行 quota、rate limit、billing 或 cost write。
- `modelGatewayEvidenceReview.ts` / `modelGatewayEvidenceReviewPanel.tsx` 只复用前三个 Model Gateway view model，集中生成 readiness rollup、evidence checklist、route / usage / audit risk 和 locked capability。
- `adminOperationsReview.ts` / `adminOperationsReviewPanel.tsx` 只复用 tenant overview、admin audit log、Model Gateway Evidence Review 和 Production Ops 静态证据，生成管理端 review/readiness 摘要。
- `adminProviderDeploymentReview.ts` / `adminProviderDeploymentReviewPanel.tsx` 只复用 Model Gateway Route Evidence、Model Gateway Evidence Review、Admin Operations Review、tenant overview 和 audit log，生成 provider/profile、model route、secret ref readiness、deployment status、operator risk 和 locked capability 摘要。
- `savedWorkflowDraftConsumer.ts` 只在 `VITE_RADISHMIND_WORKFLOW_SAVED_DRAFT_SOURCE=dev-saved-draft-http` 下连接 dev-only saved draft route，负责 sample / unsaved / validating / saving / reading / saved / version conflict / failed，以及 saved draft list `sample` / `loading` / `ready` / `empty` / `list_failed` / `restore_failed` 状态映射；默认 sample-only，不承担 production persistence。
- `workflowDraftDesigner.ts` 与 `App.tsx` 负责受控本地编辑、本地节点新增 / 移动 / 删除保护、边重建、节点属性编辑、active draft validate / save / read、版本冲突时保留本地草案，以及 saved dev draft restore 后进入 Draft Designer；`workflowUserWorkspaceHome.ts` / `workflowUserWorkspaceHomePanel.tsx` 负责从 Workspace Home 与 workflow definitions 派生本地草案，并展示 saved draft list / restore 入口。
- `App.tsx` 只负责把这些 view model 接入分组导航和页面渲染；如果新增真实后端 route、持久化状态或执行能力，应先落契约、fixture、checker 和边界文档，而不是直接在 App 或 panel 中接线。

User Workspace Home / Workflow Review Workspace 读法：

- 左侧导航按 `Workspace`、`Model Gateway`、`Workflow Review`、`Admin` 和 `Contract` 分组；用户端工作区入口、模型网关证据、workflow 审查入口和管理端入口不再与契约 guard 混排。
- 先看 User Workspace Home，确认应用组合、审查路径、最近 run、优先 readiness、主要 route evidence 和关键 stop line rollup。
- 再看 Model Gateway Overview、Route Evidence、Usage/Audit Evidence 和 Evidence Review / Readiness，确认 northbound API surfaces、provider/profile、route binding、selection cases、API key / quota / trace / audit evidence、readiness rollup、evidence checklist、route / usage / audit key risks 和 locked distribution capabilities；该区域只解释模型网关分发证据，不提供 key lifecycle、quota enforcement、cost record writes、secret resolver 或 fallback execution。
- 再看 Admin Operations Review / Readiness，确认 tenant、audit、gateway 和 Production Ops 静态证据是否足以支撑管理端只读审查，并明确 production backend、tenant mutation、raw audit export、数据库、Radish auth、repository、secret resolver、production gateway、deployment preflight、executor、writeback 和 replay 仍保持锁定。
- 再看 Admin Provider/Profile & Deployment Evidence Review / Readiness，确认 provider/profile、model route、secret ref readiness、deployment status、operator risk 和 locked capability 的管理端阅读路径；该区域只解释证据，不修改 provider/profile，不切换 model route，不解析 secret，不运行 deployment preflight。
- 先使用本地 context selection 选择 application、workflow definition、run record 和 draft template，再选择要审查的 scenario；blocked action preview 和 confirmation placeholder 应跟随当前 run / workflow definition，而不是停留在默认 fixture。
- 先看 selected context 和 review stage，确认当前审查对象；再按 scenario、draft、validation、execution plan、runtime readiness 的顺序阅读证据链。
- 最后看 Review Handoff，确认给人工审查的 recipients、key findings、evidence checklist、decision blockers 和 boundary locks 已经与当前选中上下文一致。
- blocked capability rollup 和 stop line rollup 是审查结论入口，只解释为什么当前不能 publish / execute / confirm / writeback / replay，不提供解锁或提交按钮。

Saved draft 冲突读法：

- 只有 dev-only saved draft consumer 启用后，Draft Designer 的保存、读取、校验和列表才会连接 platform memory dev store；离线模式仍只展示 sample / local draft。
- 保存返回 `version_conflict` 时，页面必须保留当前本地 active draft，并展示 saved version metadata、validation state 和 blocked capability count；它不是保存成功，也不是自动覆盖。
- 选择继续本地草案后，consumer 状态进入 `conflict_local_continued`，后续保存会使用当前 saved version 作为 expected version；这仍不是 auto merge。
- 恢复 saved version 必须由用户显式触发，并依赖冲突后刷新的当前 application saved draft list；列表只包含 sanitized summary，不暴露 secret、token、完整 claim 或 runtime material。
- Review Handoff 会显示同一份 conflict review summary，帮助 reviewer 理解冲突来源、下一步选择和 auto overwrite / auto merge 停止线；它不保存、不导出、不发送 handoff。

本地启动从仓库根目录执行：

```bash
./start.sh web-live
./start.sh web-offline
pwsh ./start.ps1 -Command web-live
```

`web-live` 会启动或复用 platform 后端和 `apps/radishmind-web/` 前端，并集中设置 dev-only live read 所需的本地环境变量。它只连接 fake-store-backed handler 和测试身份上下文，不代表 production API consumer、真实数据库、Radish auth、repository adapter 或 workflow executor ready。

如果 macOS `Control Center` / AirPlay 占用了默认 backend 端口 `7000`，不要继续用交互菜单重试；改用显式端口：

```bash
./start.sh web-live --backend-url http://127.0.0.1:7100
```

PowerShell 使用：

```powershell
pwsh ./scripts/run-radishmind-web-dev.ps1 -Mode dev-live -BackendUrl http://127.0.0.1:7100
```

如需同时验证 saved draft dev-only 保存路径，后端还需要显式启用：

```bash
RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1
RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP=1
RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE=1
VITE_RADISHMIND_WORKFLOW_SAVED_DRAFT_SOURCE=dev-saved-draft-http
```

这些开关只服务本地开发态 memory dev store，不代表 durable persistence 或 production API。

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

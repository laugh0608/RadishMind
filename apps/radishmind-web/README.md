# RadishMind Web

`apps/radishmind-web/` 是正式产品 UI 的首个落点。当前承载三组离线优先 / dev-only product surface：

- `Control Plane Read-Side`：`control-plane-read-shared-shell-v1`、`control-plane-read-admin-tenant-overview-v1`、`control-plane-read-admin-audit-log-v1`、普通离线 Admin Operations Review / Readiness、普通离线 Admin Provider/Profile & Deployment Evidence Review / Readiness、`control-plane-read-workspace-applications-v1`、`control-plane-read-workspace-api-keys-v1`、`control-plane-read-workspace-usage-quota-v1`、`control-plane-read-workspace-workflow-definitions-v1`、`control-plane-read-workspace-run-history-v1`、`control-plane-read-formal-ui-readiness-close-v1`、`control-plane-read-dev-live-consumer-v1` 和 `control-plane-read-auth-store-transition-preconditions-v1`。
- `Model Gateway / API Distribution`：普通离线 Model Gateway Overview、Route Evidence、Usage/Audit Evidence 与 Evidence Review / Readiness，复用 shared read shell、API key summary、quota summary、run history、audit log、provider runtime、`gateway-api-key-quota-readiness` 和前三个网关 view model 证据，展示 northbound API compatibility surfaces、provider/profile inventory、route binding、selection cases、key scope、quota / cost snapshot、trace / failure、audit decision、readiness rollup、evidence checklist、route / usage / audit key risks 和 locked distribution capabilities。
- `Workflow / Agent Runtime Function Surface`：既有 workflow detail / review / Draft Designer / Validation、`workflow-execution-plan-preview-offline-v1`、`workflow-runtime-readiness-inspector-offline-v1` 等完整运行时预览面板，以及显式 dev-only 的 Saved Draft consumer 与受控 Workflow Executor v0。

当前边界：

- 默认只消费 `contracts/typescript/control-plane-read-api.ts` 的离线 read-side contract。
- 前端离线 view model 默认数据必须与 `control-plane-read-response-fixtures-v1` 的 RadishFlow Copilot / Radish Docs Assistant success 样例保持一致；`control-plane-read-product-sample-consistency-v1` 会校验该 response fixture、Go fake store、consumer smoke product refs 和前端离线默认 envelope 没有漂移。
- 当显式设置 `VITE_RADISHMIND_READ_SOURCE=dev-live-http` 时，可通过 dev-only HTTP consumer 消费 fake-store-backed read handlers；后端必须同时设置 `RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1` 才会接受测试身份 header。
- Control Plane live consumer 通过 `VITE_RADISHMIND_READ_AUTH_MODE` 区分 `dev_headers`、`signed_test_token` 与 `radish_oidc_integration_test`。OIDC integration 模式只读取页面内存中的 token provider；consumer 仍请求七条 read route，但只把 Tenant Summary 与 Audit 作为允许成功的数据读取，其余五条 workspace operation 的 `workspace_membership_unavailable` 是预期 fail-closed 结果，不会回退 dev headers、signed token 或 fake repository。
- Workflow saved draft consumer 独立使用 `VITE_RADISHMIND_WORKFLOW_SAVED_DRAFT_SOURCE=dev-saved-draft-http` 开关；默认仍是 sample-only，显式启用后通过 `POST /v1/user-workspace/workflow-drafts`、`GET /v1/user-workspace/workflow-drafts`、`GET /v1/user-workspace/workflow-drafts/{draft_id}` 和 `POST /v1/user-workspace/workflow-drafts/validate` 连接 platform memory dev store。后端仍必须设置 `RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1` 与 `RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP=1`，保存还需要 `RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE=1`。
- Workflow Executor v0 独立使用 `VITE_RADISHMIND_WORKFLOW_EXECUTOR_SOURCE=dev-workflow-executor-http`；只有 active draft 是已保存、未修改且通过 bounded graph eligibility 的 executor v0 草案时，才能调用 Platform POST run，随后可用 GET scoped read 回读 record。服务端仍会重新读取并校验草案；Web 预检不构成执行授权。
- Gateway Request History 独立使用 `VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_SOURCE=dev-gateway-request-history-http`；默认 offline evidence 零请求。显式启用后，现有 Model Gateway Evidence Review 内的独立 lazy panel 读取 `/v1/model-gateway/requests` list / detail，展示 sanitized caller refs、route / protocol、selection、timing、usage availability 和稳定 failure，不回退旧 quota / cost 或 Workflow run fixture。后端必须同时启用 `RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1` 与 `RADISHMIND_GATEWAY_REQUEST_HISTORY_DEV=1`。
- Gateway Playground 使用 `VITE_RADISHMIND_GATEWAY_PLAYGROUND_SOURCE=dev-gateway-playground-http`；默认 offline 零请求。显式启用后可从 Web 调用 Chat Completions、Responses、Messages 的 unary / stream，用户可取消请求并按同一 request id 打开 sanitized history。输入输出只保留在组件内存，不写 URL 或浏览器 storage。
- Application API Integration 复用同一 Gateway Playground 开关与 caller scope；从当前 Application Detail 主动加载 `/v1/models`，生成 Chat Completions、Responses、Messages 的 cURL / Python / TypeScript 环境变量占位示例，并把当前 application / protocol / model 交给既有 Playground。示例不展示真实 key、hash 或内部 dev caller headers，application 切换会清空旧目录与选择。
- Application Catalog 使用独立的 `VITE_RADISHMIND_APPLICATION_CATALOG_SOURCE=dev-application-catalog-http` 开关；默认沿用离线应用摘要。显式启用后，`workspace-applications` 可创建、读取、编辑和归档当前主体拥有的应用，使用 `expected_version` 处理并发冲突，并在操作成功后让 Workspace Home 与 Application Detail 消费同一目录状态。它不实现删除、正式 promotion、跨主体管理或生产 repository。
- API Key Lifecycle 使用独立的 `VITE_RADISHMIND_API_KEY_LIFECYCLE_SOURCE=dev-api-key-lifecycle-http` 开关；默认仍显示离线脱敏摘要。显式启用后，`workspace-api-keys` 按当前活跃应用列出、签发、查看和 CAS 吊销开发测试态密钥，原始令牌只在签发成功视图出现一次，并可通过内存事件交给 Gateway Playground。该消费端不提供令牌恢复、浏览器持久化、生产授权、配额或计费。
- Application Configuration Draft 使用独立的 `VITE_RADISHMIND_APPLICATION_DRAFT_SOURCE=dev-application-draft-http` 开关；默认 offline 零请求且编辑状态只在当前组件内存。显式启用后可在当前 application scope 下校验、保存、列出和恢复配置草案，使用 expected-version 处理并发冲突，并把协议与模型继续交给 API Integration 或既有 Playground。草案不保存 secret、Gateway 测试输入输出，也不创建、发布或删除正式 application。
- Application Publish Review 使用独立的 `VITE_RADISHMIND_APPLICATION_PUBLISH_SOURCE=dev-application-publish-http` 开关；默认 offline 零请求。显式启用后可从当前 application 的 saved valid draft 创建不可变 candidate、恢复 snapshot / digest、追加 review CAS、查看漂移和 promotion blocker，并复用 Integration / Playground / exact History handoff。approved 仍不执行正式 application mutation。
- 渲染 read route catalog、共享状态组件、forbidden output guard、只读 `admin-tenant-overview`、只读 `admin-audit-log`、普通离线 Admin Operations Review / Readiness、普通离线 Admin Provider/Profile & Deployment Evidence Review / Readiness、可显式连接开发测试目录的 `workspace-applications`、可显式连接生命周期 API 的 `workspace-api-keys`、只读 `workspace-usage-quota`、只读 `workspace-workflow-definitions`、只读 `workspace-run-history`、User Workspace Home、Model Gateway Overview、Model Gateway Route Evidence、Model Gateway Usage/Audit Evidence、Model Gateway Evidence Review / Readiness 和 workflow function surface 面板。
- `admin-tenant-overview` 只消费 `tenant-summary-route` 的离线 view model，展示租户摘要、route metadata、request / audit ref 和状态预览。
- `admin-audit-log` 只消费 `audit-summary-list-route` 的离线 view model，展示 audit ref、actor、event kind、resource、decision、failure code、trace id、recorded timestamp、route metadata、request / audit ref、cursor 和状态预览。
- Admin Operations Review / Readiness 是普通离线只读管理端汇总面，复用 tenant overview、audit log、Model Gateway Evidence Review 和 Production Ops 静态证据，集中展示 readiness rollup、evidence checklist、operational risks 和 boundary locks；它不新增专项 gate，不请求 live backend，不接数据库、Radish auth、repository adapter、production secret resolver 或 production gateway，不提供 tenant mutation、raw audit export、API key lifecycle、quota enforcement、cost record write、deployment preflight、workflow executor、writeback 或 replay。
- Admin Provider/Profile & Deployment Evidence Review / Readiness 是普通离线只读管理端证据组织面，复用 Model Gateway Route Evidence、Model Gateway Evidence Review、Admin Operations Review、tenant overview 和 audit log，集中展示 provider/profile readiness、model route readiness、secret / deployment evidence、operator risks 和 locked capabilities；它不新增专项 gate，不请求 live backend，不执行 provider call，不接 production gateway、production secret resolver、数据库、Radish auth 或 repository adapter，不提供 provider/profile mutation、model route change、deployment preflight、workflow executor、writeback 或 replay。
- `workspace-applications` 默认消费 `application-summary-list-route` 的离线 view model；显式 Application Catalog source 下改为消费 dev/test list / create / read / update / archive API，展示目录状态、版本冲突、empty / loading / failure 和操作结果，不回退离线样例掩盖服务端错误。
- `workspace-api-keys` 默认消费 `api-key-summary-list-route` 的离线 view model；显式生命周期 source 下改为消费应用作用域的 list / issue / detail / revoke API，展示状态、作用域、版本、最近使用时间、冲突和操作结果。只有签发成功响应可在独立面板一次性展示原始令牌；列表、详情、吊销、错误和后续读取都不得展示令牌或摘要。
- `workspace-usage-quota` 只消费 `quota-summary-route` 的离线 view model，展示 quota id、period、request / token / cost limit、usage snapshot、over quota failure code、route metadata、request / audit ref 和状态预览。
- `workspace-workflow-definitions` 只消费 `workflow-definition-summary-list-route` 的离线 view model，展示 workflow definition id、application ref、version、definition status、node count、risk level、requires confirmation capable、updated at、route metadata、request / audit ref 和状态预览。
- `workspace-run-history` 只消费 `run-record-summary-list-route` 的离线 view model，展示 run id、workflow definition ref、application ref、status、failure code、cost summary、trace id、started / completed timestamp、route metadata、request / audit ref、cursor 和状态预览。
- workflow application detail 由 `workspace-applications` summary 和离线 enrichment 派生，展示 application identity、tenant / owner ref、application type、status、provider profile、risk summary、latest run ref、route / request / audit metadata 和 blocked capability preview。
- workflow definition detail 由 `workspace-workflow-definitions` summary 派生，展示 definition identity、application ref、version、nodes、edges、input / output summary、risk summary、blocked action preview 和 audit metadata。
- workflow run detail 由 `workspace-run-history` summary 派生，展示 run identity、state timeline、cost / token snapshot、trace / failure / audit metadata、blocked replay / result preview 和 request / route metadata。
- workflow blocked action preview 与 confirmation placeholder 只展示未来动作和确认流的形状、风险、human review requirement、missing prerequisites 和 audit trail，不提供 decision submit、approve、reject、defer 或 execution unlock。
- workflow draft designer 与 draft validation inspector 支持 executor v0 受控草案的本地编辑、保存和资格展示；原 execution plan preview 与 runtime readiness inspector 已明确标注为完整运行时边界。Saved Draft 可连接 `memory_dev`、聚合 `sqlite_dev` 或显式 `postgres_dev_test`，Executor v0 可在开发态运行已保存版本；这些都不代表 production API、publish、tool、confirmation、writeback 或 replay ready。
- `workflowWorkspaceContext` 是 workflow 离线组合层的共享入口，统一解析 application、workflow definition、run、draft 和 scenario selection，并统一构建 detail、blocked action、confirmation placeholder、draft validation、execution plan、runtime readiness、surface overview、scenario inspector、review workspace、User Workspace Home 和 Review Handoff；`workflow-workspace-context-consistency-v1` 会校验 App 不重新手拼这些派生链路。
- Product Surface Usage Gap Triage 已由 `product-surface-usage-gap-triage-v1` 固定：User Workspace、Workflow Review、Model Gateway 和 Admin 的使用走查只允许在发现真实阅读缺口后修正现有 view model、canonical fixture、文案、导航分组或文档读法，不新增同层产品面，不打开实现入口。
- Workflow Surface Overview 是普通离线只读总览区域，复用 workflow application detail、definition detail、run detail、selected draft、validation inspector、execution plan preview 和 runtime readiness inspector view model，把 application、definition、draft、validation、plan、readiness、latest run 和 blocked capability 的关系集中展示；它不新增专项 gate，不请求 live backend，不新增 Go route，不创建持久化结果。
- Workflow workspace context selection 允许用户在本地选择 application、workflow definition、run record 和 draft template，并让 detail、draft validation、execution plan、runtime readiness、blocked action preview、confirmation placeholder、scenario inspector、review workspace 与 overview 随当前上下文联动；该选择只改变浏览器内查看状态，不保存、不发布、不执行、不提交确认、不写回业务数据。
- Workflow Scenario Inspector 是普通离线只读场景检查区域，复用当前 application、definition、run、draft、validation、execution plan、runtime readiness 和 overview view model，展示 RadishFlow / Radish Docs 场景 intent、input contract、expected advisory output、risk / confirmation requirement、relationship map、blocked reason 和 stop lines；场景选择只改变浏览器内查看状态，不保存、不发布、不执行、不请求 live backend、不写回业务数据。
- Workflow Review Workspace 是正式离线只读审查工作区，复用当前 application、workflow definition、run、draft、validation、execution plan、runtime readiness、overview 和 scenario inspector view model，把当前选中 application + definition + run + draft + scenario 的上下文、review stage、关系链、blocked capability rollup 和 stop line rollup 集中展示；它不新增专项 gate，不请求 live backend，不新增 Go route，不创建 review / draft / plan / readiness 持久化结果，不提供保存、发布、执行、确认提交、写回或 replay 控件。
- User Workspace Home 是正式离线优先首页，复用 applications、workflow definitions、run history、API key / quota summary、Workflow Review Workspace、Workflow Surface Overview 和 Scenario Inspector 的 view model，集中展示应用组合、审查路径、最近 run、优先 readiness、主要 route evidence、blocked capability 和关键 stop line rollup；当前还提供本地 `Create draft` 入口、saved dev draft list、empty / failure state、refresh 和 restore。创建动作只写浏览器本地状态，saved list 只显示 sanitized summary / metadata，恢复动作通过既有 read route 回到 Draft Designer；不创建首页持久化结果，不提供发布、执行、确认提交、写回或 replay 控件。
- Model Gateway Overview、Route Evidence、Usage/Audit Evidence 和 Evidence Review / Readiness 是普通离线只读网关证据区，复用 read shell、API key、quota、run history、audit、provider/gateway readiness 和前三个网关 view model 证据，展示 `/v1/models`、`/v1/models/{id}`、`/v1/chat/completions`、`/v1/responses`、`/v1/messages` 的 northbound API distribution 关系，以及 provider/profile credential state、deployment mode、auth mode、streaming、route metadata、selection cases、key scope、quota policy、cost snapshot、trace records、failure code、audit decision、readiness rollup、evidence checklist、route / usage / audit key risks 和 boundary locks；它们不新增专项 gate，不请求 production gateway，不在 Web 中签发或吊销 API 密钥，不执行 quota / rate limit，不写 cost record，不解析 production secret，不启用 retry/fallback execution，不接 Radish auth 或 production repository。Platform 的 API 密钥 Gateway 认证只属于独立的开发测试 HTTP 链路。
- Workflow Review Handoff 是普通离线只读审查交接摘要，复用 User Workspace Home、Workflow Review Workspace、Workflow Surface Overview、Scenario Inspector、Validation Inspector、Execution Plan Preview、Runtime Readiness、Blocked Action Preview 和 Confirmation Placeholder 的 view model，集中展示 active draft review record、review recipients、key findings、read-side evidence checklist、decision blockers 和 boundary locks。它不导出、不发送、不保存 handoff，不请求 live backend，不新增 Go route，不提交 confirmation decision，不解锁执行，不写回或 replay。
- `control-plane-read-formal-ui-readiness-close-v1` 已用聚合 surface matrix / checker 固定七个只读页面的 route binding、状态预览、request / audit ref 和 forbidden output guard；后续普通只读展示页不再默认逐页新增专项门禁。
- `workflow-function-surface-readiness-close-v1` 已用 workflow surface matrix / checker 固定当前 workflow 离线产品面的 builder、render anchor、CSS selector、关闭项和停止线；后续普通离线 workflow 展示，包括 Workflow Review Workspace、User Workspace Home 和 Workflow Review Handoff，优先复用该聚合 gate、`npm run build` 和 fast baseline。
- 不请求生产后端，不接正式 `Radish` auth，不实现 quota enforcement、rate limit、cost record writes、production repository、validation result persistence、execution plan persistence、runtime readiness persistence、正式 publish、unrestricted workflow executor、confirmation decision、execution unlock、writeback、run replay 或 run resume。显式开发测试态例外包括 Application Catalog、API Key Lifecycle / Gateway Bearer、配置草案 / 发布审查、Saved Draft / Executor v0 和 Request History；这些能力都受独立 gate 与存储模式约束。
- `control-plane-read-dev-live-consumer-v1` 的基础七路 read consumer 仍只使用测试身份上下文；Application Catalog、API Key Lifecycle、Workflow 与 Gateway History 由各自严格消费端和显式 gate 承载。任何 dev-live 组合都不得解释为 production API consumer、正式 auth/db、production repository、quota 或完整 workflow runtime ready。
- `control-plane-read-auth-store-transition-preconditions-v1` 只固定未来 auth middleware / read store repository 迁移前置条件；不得解释为 Radish auth ready、token validation ready、database ready、repository implementation ready 或 production admin console ready。
- 不替代 `apps/radishmind-console/`；后者仍是本地 ops surface。

源码组织与派生关系：

- `modelGatewayOverview.ts` / `modelGatewayOverviewPanel.tsx` 是模型网关证据根视图，汇总 northbound API surface、provider/profile inventory、route metadata 和 gateway readiness 证据。
- `modelGatewayRouteEvidence.ts` / `modelGatewayRouteEvidencePanel.tsx` 只从 Overview 与 read shell 派生 route binding、selection case、streaming、auth mode、secret ref 和 route risk，不创建新的产品真相源。
- `modelGatewayUsageAuditEvidence.ts` / `modelGatewayUsageAuditEvidencePanel.tsx` 只从 Overview、Route Evidence、API key、quota、run history 和 audit log 派生 usage / audit 证据，不执行 quota、rate limit、billing 或 cost write。
- `modelGatewayEvidenceReview.ts` / `modelGatewayEvidenceReviewPanel.tsx` 只复用前三个 Model Gateway view model，集中生成 readiness rollup、evidence checklist、route / usage / audit risk 和 locked capability。
- `modelGatewayRequestHistoryConsumer.ts` / `modelGatewayRequestHistoryPanel.tsx` 是与离线 evidence 分离的显式 dev/test consumer 和 lazy review panel；consumer 负责 scope query、专用 Gateway headers、strict mapping、forbidden-field scan、过滤、分页和详情，panel 不持有 production auth、quota、billing、retry / fallback 或写入能力。
- `apiKeyLifecycleConsumer.ts` / `apiKeyLifecyclePanel.tsx` 负责应用作用域密钥列表、签发、一次性交接、详情、CAS 吊销、严格响应映射和 forbidden-field scan；原始令牌只存在于签发成功状态与显式 Playground 内存交接，不进入通用 read shell。
- `modelGatewayPlaygroundConsumer.ts` / `modelGatewayPlaygroundPanel.tsx` 在 `dev_headers` 与 `api_key_dev_test` 之间严格二选一；API 密钥模式只发送 Bearer 凭据，并在路由离开、应用切换、显式清除或卸载时中止活动请求并丢弃凭据。通过校验的模型目录可以脱敏事件复用于 API 接入区和配置草案，事件不包含凭据。
- `adminOperationsReview.ts` / `adminOperationsReviewPanel.tsx` 只复用 tenant overview、admin audit log、Model Gateway Evidence Review 和 Production Ops 静态证据，生成管理端 review/readiness 摘要。
- `adminProviderDeploymentReview.ts` / `adminProviderDeploymentReviewPanel.tsx` 只复用 Model Gateway Route Evidence、Model Gateway Evidence Review、Admin Operations Review、tenant overview 和 audit log，生成 provider/profile、model route、secret ref readiness、deployment status、operator risk 和 locked capability 摘要。
- `savedWorkflowDraftConsumer.ts` 只在 `VITE_RADISHMIND_WORKFLOW_SAVED_DRAFT_SOURCE=dev-saved-draft-http` 下连接 dev-only saved draft route，负责 sample / unsaved / validating / saving / reading / saved / version conflict / `conflict_local_continued` / failed，以及 saved draft list `sample` / `loading` / `ready` / `empty` / `list_failed` / `restore_failed` 状态映射；`savedWorkflowDraftLifecycle.ts` 负责 persisted base version、validate / failure version preservation 和 unresolved conflict blocking，冲突审查 summary 只派生 `savedMetadataState`、`restoreActionState`、`restoreUnavailableReason`、本地草案保留说明和 reviewer 下一步，默认 sample-only，不承担 production persistence。
- `workflowExecutorConsumer.ts` 负责受控草案构建、bounded graph eligibility、dev HTTP request / response 映射与 record guard；`workflowExecutorPanel.tsx` 只负责运行输入、condition、状态、节点时间线、advisory output 和副作用计数展示，不持有执行授权或生产配置。
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

- 只有 dev-only saved draft consumer 启用后，Draft Designer 的保存、读取、校验和列表才会连接所选 `memory_dev`、聚合 `sqlite_dev` 或显式 `postgres_dev_test` store；离线模式仍只展示 sample / local draft。
- 保存返回 `version_conflict` 时，页面必须保留当前本地 active draft，并展示 saved version metadata、validation state 和 blocked capability count；它不是保存成功，也不是自动覆盖。
- 选择继续本地草案后，consumer 状态进入 `conflict_local_continued`，后续保存会使用当前 saved version 作为 expected version；这仍不是 auto merge。
- 恢复 saved version 必须由用户显式触发，并依赖冲突后刷新的当前 application saved draft list；列表只包含 sanitized summary，不暴露 secret、token、完整 claim 或 runtime material。metadata 刷新中、列表为空、列表失败或缺少匹配 summary 时，恢复入口保持禁用并显示原因。
- Review Handoff 会显示同一份 conflict review summary，帮助 reviewer 理解冲突来源、下一步选择和 auto overwrite / auto merge 停止线；它不保存、不导出、不发送 handoff。

本地启动从仓库根目录执行：

```bash
./start.sh web-live
./start.sh web-offline
pwsh ./start.ps1 -Command web-live
```

`web-live` 会启动或复用 platform 后端和 `apps/radishmind-web/` 前端，并集中设置 dev-only live read 所需的本地环境变量。默认模式只连接 fake-store-backed handler；Saved Draft、Application Draft / Publish、Application Catalog 和 Gateway History 等写入或运行能力都必须通过各自参数显式启用。所有模式都不代表 production API consumer、Radish auth、production repository 或完整 workflow runtime ready。

如果 macOS `Control Center` / AirPlay 占用了默认 backend 端口 `7000`，不要继续用交互菜单重试；改用显式端口：

```bash
./start.sh web-live --backend-url http://127.0.0.1:7100
```

PowerShell 使用：

```powershell
pwsh ./scripts/run-radishmind-web-dev.ps1 -Mode dev-live -BackendUrl http://127.0.0.1:7100
```

如需同时验证 Saved Draft dev-only 保存路径，优先通过 launcher 的显式开发态开关启动：

```bash
./scripts/run-radishmind-web-dev.sh --mode dev-live --saved-draft-dev --backend-url http://127.0.0.1:7100
```

```powershell
pwsh ./scripts/run-radishmind-web-dev.ps1 -Mode dev-live -SavedDraftDev -BackendUrl http://127.0.0.1:7100
```

该开关会集中设置以下六个环境变量；默认 `dev-live` 仍只打开 fake read consumer，不隐式开放 Saved Draft 写入或 executor：

```bash
RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1
RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP=1
RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE=1
VITE_RADISHMIND_WORKFLOW_SAVED_DRAFT_SOURCE=dev-saved-draft-http
RADISHMIND_WORKFLOW_EXECUTOR_DEV=1
VITE_RADISHMIND_WORKFLOW_EXECUTOR_SOURCE=dev-workflow-executor-http
```

这些开关只服务本地开发态 Saved Draft 与受控 executor v0；`--saved-draft-postgres-dev-test` 可显式使用 PostgreSQL dev/test repository。两者都不代表 production persistence、production auth 或 production API。

Application Configuration Draft 通过独立开关启用；可与 Gateway PostgreSQL dev/test 联调组合使用：

```bash
./scripts/run-radishmind-web-dev.sh --mode dev-live --application-draft-dev --gateway-request-postgres-dev-test --backend-url http://127.0.0.1:7100
```

```powershell
pwsh ./scripts/run-radishmind-web-dev.ps1 -Mode dev-live -ApplicationDraftDev -GatewayRequestPostgresDevTest -BackendUrl http://127.0.0.1:7100
```

launcher 会设置 application draft 的 dev HTTP / write gate、当前 workspace 和 Web consumer source；如同时选择 PostgreSQL dev/test，会先检查 application draft migration marker，再启动 Platform。默认 offline 与未传入该开关的 dev-live 都不会发出 application draft 请求。

Application Publish Review 可使用 memory dev 或与 Application Draft 共用同一 PostgreSQL dev/test 实例中的独立 schema：

```bash
./scripts/run-radishmind-web-dev.sh --mode dev-live --application-publish-postgres-dev-test --gateway-request-postgres-dev-test --backend-url http://127.0.0.1:7100
```

```powershell
pwsh ./scripts/run-radishmind-web-dev.ps1 -Mode dev-live -ApplicationPublishPostgresDevTest -GatewayRequestPostgresDevTest -BackendUrl http://127.0.0.1:7100
```

launcher 会同时启用 application draft 与 publish candidate 的 dev HTTP / write gate，并检查两个 migration marker；`--application-publish-dev` / `-ApplicationPublishDev` 使用 memory dev。两种模式都不启用 promotion endpoint、production auth 或正式 application repository。

Application Catalog 的独立 PostgreSQL 开发测试模式可通过以下参数启用：

```bash
./scripts/run-radishmind-web-dev.sh --mode dev-live --application-catalog-postgres-dev-test --backend-url http://127.0.0.1:7100
```

```powershell
pwsh ./scripts/run-radishmind-web-dev.ps1 -Mode dev-live -ApplicationCatalogPostgresDevTest -BackendUrl http://127.0.0.1:7100
```

launcher 会检查 Application Catalog migration marker，打开目录 HTTP / write gate，并设置 Web consumer source。该页面支持创建、编辑和归档应用；该 PostgreSQL 参数不自动启用 API 密钥页面，完整 API 密钥产品链使用下面的 SQLite 本地产品档。

API 密钥一次性交接与 Gateway 开发测试态认证使用独立的本地产品档参数：

```bash
./scripts/run-radishmind-web-dev.sh --mode dev-live --api-key-local-product --backend-url http://127.0.0.1:7100
```

```powershell
pwsh ./scripts/run-radishmind-web-dev.ps1 -Mode dev-live -APIKeyLocalProduct -BackendUrl http://127.0.0.1:7100
```

该参数必须单独使用。launcher 通过默认 `local-product` 平台档选择七组件共享 `sqlite_dev`，显式启用 Application Catalog、Configuration Draft、Publish Review、API Key Lifecycle、Gateway Playground 与 Request History 的 Web consumer，并把 Gateway 认证固定为 `api_key_dev_test`。页面支持应用作用域列表、签发、一次性令牌复制 / 清除、内存交接、模型目录、单次 / 流式 / 取消调用、脱敏历史和吊销；Playground 成功读取的模型目录可通过严格校验的脱敏事件复用于同应用配置，Bearer 凭据不会传播到配置面板。原始令牌只存在于签发响应与当前 React 组件内存，离开 Playground 路由即清除，不进入 URL、浏览器存储、cookie、日志或后续响应。

该模式只属于内部开发者预览，不启用生产 API 密钥、正式成员关系、配额、限流、计费、provider credential 或公开生产 Gateway。完整 HTTP 边界与手工排障仍见[应用目录与 API 密钥开发测试指南](../../docs/features/user-workspace/application-catalog-api-key-dev-test-guide.md)。

Gateway Request History 与 Playground 可以通过 launcher 的 `--gateway-request-postgres-dev-test` / `-GatewayRequestPostgresDevTest` 一起开启；launcher 会为 Platform 与 Web 绑定同一 caller scope，并先执行 Saved Draft、Workflow Run、Gateway Request、Application Draft 与 Application Publish 五套 PostgreSQL migration status preflight。真实 northbound 请求仍必须携带同组 `X-RadishMind-Dev-Gateway-*` header 才会形成记录。手动联调 `memory_dev` 时可使用以下必要开关：

```bash
RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1
RADISHMIND_GATEWAY_REQUEST_HISTORY_DEV=1
VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_SOURCE=dev-gateway-request-history-http
VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_BASE_URL=http://127.0.0.1:7100
VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_TENANT_REF=tenant_demo
VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_WORKSPACE_ID=workspace_demo
VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_CONSUMER_REF=consumer_web_dev
VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_SUBJECT_REF=subject_web_dev
```

`memory_dev` 随 Platform 重启清空；`postgres_dev_test` 已完成独立 migration、no-fallback 和重启恢复验收，但仍只属于开发 / 测试态 history，不是 production audit、billing 或合规账本。

Control Plane 的 signed test / deterministic OIDC 联调属于独立鉴权路径。Web 侧至少需要：

```bash
VITE_RADISHMIND_READ_SOURCE=dev-live-http
VITE_RADISHMIND_READ_AUTH_MODE=radish_oidc_integration_test
VITE_RADISHMIND_READ_STORE_MODE=postgres_dev_test
VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL=http://127.0.0.1:7100
VITE_RADISHMIND_DEV_READ_TENANT_REF=tenant_demo
```

Platform 必须同时使用 `RADISHMIND_CONTROL_PLANE_READ_AUTH_MODE=radish_oidc_integration_test`、`RADISHMIND_CONTROL_PLANE_READ_STORE=postgres_dev_test` 和完整 reviewed OIDC integration 配置启动。短期 token 只能由受控联调 harness 通过 `globalThis.__RADISHMIND_CONTROL_PLANE_OIDC_INTEGRATION_TOKEN__` 的内存 closure 提供；该 hook 不是登录 UI 或长期 session，不得把 token 写入 URL、源码、`.env`、`localStorage`、`sessionStorage`、日志、截图或构建产物。仓库当前没有 reviewed Radish evidence 或真实 token 时，不执行这条真实联调。

底层 wrapper 也可单独执行：

```bash
./scripts/run-radishmind-web-dev.sh --mode dev-live
./scripts/run-radishmind-web-dev.sh --mode offline
```

常规包命令仍保留在 `apps/radishmind-web/` 下：

```bash
cd apps/radishmind-web
npm test
npm run test:coverage
npm run build
npm run preview
```

`npm run test:coverage` 是 PR / release 的可发现覆盖率入口，当前预算为行 `90%`、分支 `78%`、函数 `85%`；它约束已进入 Node 严格消费端测试的模块，不替代 React 页面构建和真实浏览器验收。

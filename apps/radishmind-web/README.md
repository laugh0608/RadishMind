# RadishMind Web

`apps/radishmind-web/` 是正式产品 UI 的首个落点，当前承载 `control-plane-read-shared-shell-v1`、`control-plane-read-admin-tenant-overview-v1`、`control-plane-read-admin-audit-log-v1`、`control-plane-read-workspace-applications-v1`、`control-plane-read-workspace-api-keys-v1`、`control-plane-read-workspace-usage-quota-v1`、`control-plane-read-workspace-workflow-definitions-v1`、`control-plane-read-workspace-run-history-v1`、`control-plane-read-formal-ui-readiness-close-v1` 和 `control-plane-read-dev-live-consumer-v1`。

当前边界：

- 默认只消费 `contracts/typescript/control-plane-read-api.ts` 的离线 read-side contract。
- 当显式设置 `VITE_RADISHMIND_READ_SOURCE=dev-live-http` 时，可通过 dev-only HTTP consumer 消费 fake-store-backed read handlers；后端必须同时设置 `RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1` 才会接受测试身份 header。
- 只渲染 read route catalog、共享状态组件、forbidden output guard、只读 `admin-tenant-overview`、只读 `admin-audit-log`、只读 `workspace-applications`、只读 `workspace-api-keys`、只读 `workspace-usage-quota`、只读 `workspace-workflow-definitions` 和只读 `workspace-run-history` 页面切片。
- `admin-tenant-overview` 只消费 `tenant-summary-route` 的离线 view model，展示租户摘要、route metadata、request / audit ref 和状态预览。
- `admin-audit-log` 只消费 `audit-summary-list-route` 的离线 view model，展示 audit ref、actor、event kind、resource、decision、failure code、trace id、recorded timestamp、route metadata、request / audit ref、cursor 和状态预览。
- `workspace-applications` 只消费 `application-summary-list-route` 的离线 view model，展示应用摘要列表、cursor、route metadata、request / audit ref 和状态预览。
- `workspace-api-keys` 只消费 `api-key-summary-list-route` 的离线 view model，展示 API key id、owner、scope、state、时间字段、route metadata、request / audit ref 和状态预览，不展示 key value 或 hash。
- `workspace-usage-quota` 只消费 `quota-summary-route` 的离线 view model，展示 quota id、period、request / token / cost limit、usage snapshot、over quota failure code、route metadata、request / audit ref 和状态预览。
- `workspace-workflow-definitions` 只消费 `workflow-definition-summary-list-route` 的离线 view model，展示 workflow definition id、application ref、version、definition status、node count、risk level、requires confirmation capable、updated at、route metadata、request / audit ref 和状态预览。
- `workspace-run-history` 只消费 `run-record-summary-list-route` 的离线 view model，展示 run id、workflow definition ref、application ref、status、failure code、cost summary、trace id、started / completed timestamp、route metadata、request / audit ref、cursor 和状态预览。
- `control-plane-read-formal-ui-readiness-close-v1` 已用聚合 surface matrix / checker 固定七个只读页面的 route binding、状态预览、request / audit ref 和 forbidden output guard；后续普通只读展示页不再默认逐页新增专项门禁。
- 不请求生产后端，不接 `Radish` OIDC，不接数据库，不实现 API key lifecycle、quota enforcement、rate limit、billing、cost ledger、workflow builder、workflow executor、confirmation、writeback 或 replay。
- `control-plane-read-dev-live-consumer-v1` 只能连接 fake-store-backed handler 和测试身份上下文；不得解释为 production API consumer、真实 auth/db、repository、API key / quota 或 workflow executor ready。
- 不替代 `apps/radishmind-console/`；后者仍是本地 ops surface。

本地命令：

```bash
npm run dev
npm run build
npm run preview
```

dev-only live read 示例：

```bash
RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1 go run ./services/platform/cmd/radishmind-platform
VITE_RADISHMIND_READ_SOURCE=dev-live-http VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL=http://127.0.0.1:7000 npm run dev
```

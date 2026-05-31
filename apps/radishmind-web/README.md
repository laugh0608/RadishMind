# RadishMind Web

`apps/radishmind-web/` 是正式产品 UI 的首个落点，当前承载 `control-plane-read-shared-shell-v1`、`control-plane-read-admin-tenant-overview-v1`、`control-plane-read-workspace-applications-v1` 和 `control-plane-read-workspace-api-keys-v1`。

当前边界：

- 只消费 `contracts/typescript/control-plane-read-api.ts` 的离线 read-side contract。
- 只渲染 read route catalog、共享状态组件、forbidden output guard、只读 `admin-tenant-overview`、只读 `workspace-applications` 和只读 `workspace-api-keys` 页面切片。
- `admin-tenant-overview` 只消费 `tenant-summary-route` 的离线 view model，展示租户摘要、route metadata、request / audit ref 和状态预览。
- `workspace-applications` 只消费 `application-summary-list-route` 的离线 view model，展示应用摘要列表、cursor、route metadata、request / audit ref 和状态预览。
- `workspace-api-keys` 只消费 `api-key-summary-list-route` 的离线 view model，展示 API key id、owner、scope、state、时间字段、route metadata、request / audit ref 和状态预览，不展示 key value 或 hash。
- 不请求真实后端，不接 `Radish` OIDC，不接数据库，不实现 API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。
- 不替代 `apps/radishmind-console/`；后者仍是本地 ops surface。

本地命令：

```bash
npm run dev
npm run build
npm run preview
```

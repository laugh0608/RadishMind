# RadishMind Web

`apps/radishmind-web/` 是正式产品 UI 的首个落点，当前只承载 `control-plane-read-shared-shell-v1`。

当前边界：

- 只消费 `contracts/typescript/control-plane-read-api.ts` 的离线 read-side contract。
- 只渲染 read route catalog、共享状态组件和 forbidden output guard。
- 不请求真实后端，不接 `Radish` OIDC，不接数据库，不实现 API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。
- 不替代 `apps/radishmind-console/`；后者仍是本地 ops surface。

本地命令：

```bash
npm run dev
npm run build
npm run preview
```

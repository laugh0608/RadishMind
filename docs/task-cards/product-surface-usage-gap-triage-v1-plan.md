# Product Surface Usage Gap Triage v1 计划

更新时间：2026-06-13

## 任务目标

本任务卡用于在 `product-surface-readiness-implementation-trigger-recheck-v1`、`control-plane-read-schema-artifact-evidence-v1` 和 `control-plane-read-implementation-entry-review-v1` 之后，固定四个产品面的使用走查口径。

本任务不新增同层只读 UI 面板，不启动开发服务器，不接真实后端，不创建数据库、OIDC、repository adapter、store selector 或 runtime artifact。它只定义如何从现有 User Workspace、Workflow Review、Model Gateway 和 Admin 阅读路径中识别真实阅读缺口，并把后续动作限制为定向修正或等待实现触发条件。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [能力矩阵](../radishmind-capability-matrix.md)
- [跨项目集成契约](../radishmind-integration-contracts.md)
- [Control Plane Read-Side 契约](../contracts/control-plane-read-side.md)
- [`Product Surface Readiness / Implementation Trigger Recheck` v1 计划](product-surface-readiness-implementation-trigger-recheck-v1-plan.md)
- [`Control Plane Read Schema Artifact Evidence` v1 计划](control-plane-read-schema-artifact-evidence-v1-plan.md)
- [`Control Plane Read Implementation Entry Review` v1 计划](control-plane-read-implementation-entry-review-v1-plan.md)
- `apps/radishmind-web/README.md`
- `apps/radishmind-web/src/features/control-plane-read/`

## 走查范围

`product-surface-usage-gap-triage-v1` 状态固定为 `product_surface_usage_gap_triage_defined`。

走查范围：

- User Workspace：确认应用组合、最近 run、review context、API key / quota、route evidence 和 stop line 能从首页进入现有详情。
- Workflow Review：确认 application、definition、run、draft、scenario、blocked action、confirmation placeholder、handoff 和 stop line 的关系由 `workflowWorkspaceContext` 统一派生。
- Model Gateway / API Distribution：确认 provider/profile、route binding、selection case、API key、quota、trace、audit 和 locked capability 的证据链可读。
- Admin Control Plane：确认 tenant、audit、provider/profile、secret ref、deployment status、operator risk 和 boundary lock 的管理端阅读路径可读。
- Schema evidence：确认七条 read route 到未来 schema artifact / projection 的映射只作为证据，不被提升为 schema artifact manifest ready 或 implementation trigger satisfied。

## 验收口径

- `scripts/checks/fixtures/product-surface-usage-gap-triage-v1.json` 固定四个产品面的 usage walkthrough matrix、用户决策问题、证据来源、schema route crosscheck、定向修正策略、实现入口等待策略、禁止产物和无副作用口径。
- `scripts/checks/control_plane/check-product-surface-usage-gap-triage-v1.py` 进入 `./scripts/check-repo.sh --fast`。
- checker 必须跨读 product surface recheck、schema artifact evidence 和 implementation entry review，确认 usage triage 不打开实现入口、不新增同层产品面、不创建 runtime artifact。
- current focus、能力矩阵、integration contracts、read-side contract、task card index、scripts README、platform web README 和 W24 周志同步该切片。

## 非目标

- 不新增产品 UI 面板。
- 不改 `apps/radishmind-web/` 产品页面或 view model。
- 不启动开发服务器，不做真实浏览器联调。
- 不创建 implementation task card。
- 不创建 migration 目录、manifest、SQL、DDL review artifact、rollback fixture、schema smoke artifact 或 database fixture。
- 不实现 migration runner、store selector、repository interface、repository adapter、database query、auth middleware、token validation 或 production API consumer。
- 不接真实数据库、Radish OIDC、repository adapter、production gateway、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

## 后续动作规则

- 若真实使用暴露阅读缺口，优先修正现有 view model、fixture、文案、导航分组或 README 读法。
- 普通 copy、布局和信息层级修正继续复用 web build、聚合 read-side checker 和 fast baseline。
- 只有新增 API、执行边界、生产声明、数据格式、外部 provider 风险或高风险能力时，才新增专项 gate。
- 若 implementation trigger 仍为 `not_satisfied`，不得创建 implementation task card。
- 若未来某个 trigger 满足，只能先按 implementation entry review 的顺序选择一个实现方向进入任务卡。

## 停止线

- 不能把 usage triage 写成 production ready。
- 不能把产品阅读路径完整写成 implementation trigger satisfied。
- 不能把 schema evidence 写成 schema artifact manifest ready。
- 不能把 fake-store-backed dev path 写成 durable read store 或 production API consumer。
- 不能为了弥补普通阅读问题新增同层只读面板。

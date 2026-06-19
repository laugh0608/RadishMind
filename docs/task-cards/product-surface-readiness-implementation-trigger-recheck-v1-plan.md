# Product Surface Readiness / Implementation Trigger Recheck v1 计划

更新时间：2026-06-13

## 任务目标

本任务卡用于在 User Workspace、Workflow Review、Model Gateway 和 Admin 四个产品面完成第一版离线只读证据组织后，做一次产品面 readiness 与 read-side implementation trigger 复核。

本任务不新增 UI 面板、不启动开发服务器、不接真实后端、不创建数据库或 auth 实现、不进入 executor / writeback / replay。它只固定当前判断：四个产品面没有新的同层阅读缺口需要默认追加页面；read-side implementation trigger 仍未满足。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [能力矩阵](../radishmind-capability-matrix.md)
- [跨项目集成契约](../radishmind-integration-contracts.md)
- [`Control Plane / User Workspace / Workflow` v1 计划](control-plane-user-workspace-workflow-v1-plan.md)
- [`Control Plane Read Implementation Trigger Review` v1 计划](control-plane-read-implementation-trigger-review-v1-plan.md)
- [`Workflow Function Surface Readiness Close` v1 计划](workflow-function-surface-readiness-close-v1-plan.md)
- `apps/radishmind-web/README.md`
- `apps/radishmind-web/src/features/control-plane-read/`

## 当前事实

- User Workspace 已有 User Workspace Home，能汇总 applications、workflow definitions、run history、API key / quota summary、Workflow Review、Gateway route evidence 和 stop line。
- Workflow Review 已有 Workflow Review Workspace、Workflow Review Handoff、Workflow Surface Overview、Scenario Inspector、Runtime Readiness、Blocked Action Preview 和 Confirmation Placeholder。
- Model Gateway 已有 Overview、Route Evidence、Usage/Audit Evidence 和 Evidence Review / Readiness。
- Admin 已有 Tenant Overview、Audit Log、Admin Operations Review / Readiness 和 Admin Provider/Profile & Deployment Evidence Review / Readiness。
- `control-plane-read-implementation-trigger-review-v1` 当前仍明确：schema artifact、store selector、production auth 和 adapter smoke 四类候选均为 `not_satisfied`，没有 read-side implementation trigger satisfied。

## 复核结论

`product-surface-readiness-implementation-trigger-recheck-v1` 状态固定为 `product_surface_readiness_trigger_recheck_defined`。

复核结论：

- 四个产品面当前都已有离线只读阅读路径。
- 没有新的默认同层只读面板需要继续追加。
- 后续只在真实使用暴露阅读缺口时做定向修正。
- 没有新的 read-side implementation trigger satisfied。
- 不创建 repository adapter、SQL、migration、store selector、Radish OIDC、production API consumer、API key lifecycle、quota enforcement、deployment preflight、workflow executor、confirmation、writeback 或 replay 实现任务。

## 验收口径

- `scripts/checks/fixtures/product-surface-readiness-implementation-trigger-recheck-v1.json` 固定四个产品面的 readiness recheck、reading gap decision、implementation trigger decision 和停止线。
- `scripts/checks/control_plane/check-product-surface-readiness-implementation-trigger-recheck-v1.py` 已进入 `./scripts/check-repo.sh --fast`。
- checker 必须验证依赖 gate 状态、四个产品面矩阵、implementation trigger recheck、真实联调开发服务器口径、禁止产物、无副作用策略和关键文档引用。
- current focus、能力矩阵、跨项目集成契约、任务卡索引、脚本说明和周志同步该切片。

## 非目标

- 不新增产品 UI 面板。
- 不新增 live backend consumer。
- 不启动或保留开发服务器。
- 不实现 production API consumer、数据库、Radish OIDC、repository adapter、store selector、API key lifecycle、quota enforcement、billing、secret resolver、deployment preflight、workflow builder、workflow executor、confirmation、writeback 或 replay。

## 停止线

- 不能把只读产品面 readiness recheck 写成 production ready。
- 不能把阅读路径完整写成实现触发条件满足。
- 不能把 fake-store-backed dev path 写成真实数据库或生产 API。
- 不能在没有明确真实联调窗口时启动开发服务器。

# RadishMind 平台专题入口

更新时间：2026-06-15

## 文档目的

本目录用于承接跨产品面的平台专题。平台专题不属于某一个页面，也不应塞进单个产品面大方向文档；它们负责固定 auth、store、repository、provider、deployment、dev-only write path 等长期边界和准入条件。

具体实现批次仍进入 `docs/task-cards/`。平台专题只说明为什么要做、允许打开什么、依赖什么证据、哪些能力必须作为独立目标。

## 何时放在这里

- 能力会影响多个产品面，例如 User Workspace、Admin Control Plane 和 Workflow 同时依赖的 auth / store / repository。
- 能力属于服务层或运行层，例如 provider runtime、deployment runtime config、secret backend、store selector。
- 能力需要明确 dev-only、test、production 的启用边界。
- 能力不是单个页面可独立验收的 UI 组织问题。

## 当前平台专题候选

| 专题 | 当前状态 | 现有事实源 | 下一步 |
| --- | --- | --- | --- |
| Auth / Store Transition | readiness 已有证据 | `docs/contracts/control-plane-read-side.md`、相关 control-plane read task cards | 等真实 OIDC 或 store 迁移入口满足后再拆细专题 |
| Repository Adapter / Store Selector | readiness 已有证据，implementation trigger 未满足 | `control-plane-read-*repository*` 与 `*store-selector*` task cards | 不在当前默认推进 |
| Provider Runtime & Health | close candidate | [Provider Runtime & Health v1 任务卡](../task-cards/provider-runtime-health-v1-plan.md) | 不继续扩同层 provider 小切片 |
| Production Ops / Deployment | 静态边界已 close | [Production Ops Hardening v1](../task-cards/production-ops-hardening-v1-plan.md)、[Docker Deployment v1](../task-cards/production-ops-docker-deployment-v1-plan.md) | 等测试或生产前复核窗口 |
| Dev-only Write Path | planned | [Dev-only Saved Draft Consumer](../features/workflow/dev-only-saved-draft-consumer.md) | 先服务 saved draft consumer integration |

## 停止线

- 不用平台专题替代 task card；进入代码实现、route、schema、checker 或高风险 gate 时仍必须有实现批次记录。
- 不把 dev-only 能力写成 production ready。
- 不在没有明确触发条件时打开真实数据库、Radish OIDC、repository adapter、store selector、secret backend、executor、confirmation、writeback 或 replay。

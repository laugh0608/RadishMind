# RadishMind 平台专题入口

更新时间：2026-06-19

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
| Repository Adapter / Store Selector | control-plane readiness 已有证据；saved workflow draft selector 已按 fail-closed mode selection 落地 | `control-plane-read-*repository*`、`*store-selector*` task cards、`docs/features/workflow/saved-workflow-draft-store-selector-implementation-v1.md` | 后续 repository adapter / database 仍需独立实现专题 |
| Provider Runtime & Health | close candidate | [Provider Runtime & Health v1 任务卡](../task-cards/provider-runtime-health-v1-plan.md) | 不继续扩同层 provider 小切片 |
| Production Ops / Deployment | 静态边界已 close；secret ref config readiness 与 provider profile binding readiness 已定义 | [Production Ops Hardening v1](../task-cards/production-ops-hardening-v1-plan.md)、[Docker Deployment v1](../task-cards/production-ops-docker-deployment-v1-plan.md)、[Production Secret Backend Config / Secret Ref Readiness v1](production-secret-backend-config-secret-ref-readiness-v1.md)、[Production Secret Backend Provider Profile Secret Binding Readiness v1](production-secret-backend-provider-profile-secret-binding-readiness-v1.md) | secret backend 后续只能从 disabled resolver interface 或 runbook / negative gates 中选择单一方向 |
| Dev-only Write Path | implemented | [Dev-only Saved Draft Consumer](../features/workflow/dev-only-saved-draft-consumer.md) | 已服务 saved draft consumer integration；后续按 conflict UI / smoke / contract 固化拆批次 |

## Saved Workflow Draft Store Selector

Saved workflow draft 的平台服务配置已经新增 `workflow_saved_draft_store` / `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE`。当前唯一可成功读写的 mode 是 `memory_dev`，它只服务 dev-only saved draft route；`repository_disabled` 和 `repository` 都返回 `repository_store_disabled`，unknown mode 返回 `invalid_draft_store_mode`。这些失败必须直接暴露给 save / read / list 调用方，不得回退到 memory dev、sample 或 fixture。

`services/platform/migrations/workflow_saved_drafts/` 当前只承载 `manifest.json`、`ddl-review.md`、`rollback-evidence.json` 和 `migration-smoke.json` 四个静态 schema artifact 证据文件。它们说明 future durable store 的 logical schema、predicate、review 和 rollback 边界，不是 SQL migration，不会被 platform service 自动执行，也不表示 repository mode、真实数据库或 production auth 已可用。

## Production Secret Backend Config / Secret Ref Readiness

`Production Secret Backend Config / Secret Ref Readiness v1` 已固定 `config_secret_ref_readiness_defined`，对应切片为 `production-secret-backend-config-secret-ref-readiness-v1`。它只把 `production-secret-backend-implementation-readiness` 中的 `config-secret-ref-readiness` 前置条件推进到可检查状态，要求后续配置层只能处理 `secret_ref` 存在性、`secret_backend_configured`、`secret_ref_present`、`missing_secret_refs` 和 `field_sources` 等脱敏状态。

该专题不修改 `config.LoadFromEnv` runtime，不实现 resolver、不创建 fake resolver、不调用云 secret 服务、不读取 secret value、不接 database connection provider，也不启用 workflow saved draft repository mode。`production_secret_backend` 仍为 `not_satisfied`，`resolver_implementation_status` 仍为 `not_started`。

## Production Secret Backend Provider Profile Secret Binding Readiness

`Production Secret Backend Provider Profile Secret Binding Readiness v1` 已固定 `provider_profile_secret_binding_readiness_defined`，对应切片为 `production-secret-backend-provider-profile-secret-binding-readiness-v1`。它只把 `production-secret-backend-implementation-readiness` 中的 `provider-profile-secret-binding` 前置条件推进到可检查状态，要求后续 provider/profile inventory 只能声明 `credential_requirement`、`secret_ref_status`、`secret_ref_present`、`missing_secret_refs` 和 `field_sources` 等脱敏状态。

该专题不修改 provider runtime，不实现 resolver、不创建 fake resolver、不调用云 secret 服务、不读取 secret value、不把 `secret_ref_status=present` 写成 credential resolved、不接 database connection provider，也不启用 workflow saved draft repository mode。`production_secret_backend` 仍为 `not_satisfied`，`resolver_implementation_status` 仍为 `not_started`。后续若继续 production secret backend，应在 disabled resolver interface 或 operator runbook / negative gates 中选择单一方向。

## 停止线

- 不用平台专题替代 task card；进入代码实现、route、schema、checker 或高风险 gate 时仍必须有实现批次记录。
- 不把 dev-only 能力写成 production ready。
- 不在没有明确触发条件时打开真实数据库、Radish OIDC、repository adapter、store selector、secret backend、executor、confirmation、writeback 或 replay。

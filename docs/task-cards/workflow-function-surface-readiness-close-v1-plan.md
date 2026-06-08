# `Workflow Function Surface Readiness Close` v1 计划

更新时间：2026-06-08

## 任务目标

本任务卡用于落实 `workflow-function-surface-readiness-close-v1`，把 `Workflow / Agent Runtime Function Surface v1` 已完成的离线只读页面收束为一个聚合 readiness gate。

状态固定为：`workflow_function_surface_readiness_closed`。

该状态只证明当前 workflow 产品面已有可复验的 surface matrix、依赖链和停止线，不代表 runtime API、workflow builder、executor、confirmation flow、数据库、Radish OIDC、repository adapter、business writeback、replay 或 production API consumer 已可用。

## 输入事实源

- [`Workflow / Agent Runtime Function Surface` v1 计划](workflow-agent-runtime-function-surface-v1-plan.md)
- [`Workflow Runtime Readiness Inspector Offline` v1 计划](workflow-runtime-readiness-inspector-offline-v1-plan.md)
- [当前推进焦点](../radishmind-current-focus.md)
- [阶段路线图](../radishmind-roadmap.md)
- [能力矩阵](../radishmind-capability-matrix.md)
- [跨项目集成契约](../radishmind-integration-contracts.md)

## 关闭范围

- 聚合依赖 `workflow-function-surface-boundary-v1`、application detail、definition detail、run detail、blocked action preview、confirmation placeholder、draft designer、validation inspector、execution plan preview 和 runtime readiness inspector。
- 在 `scripts/checks/fixtures/workflow-function-surface-readiness-close-v1.json` 中固定 surface matrix、关闭项、停止线和验证策略。
- 在 `scripts/checks/control_plane/check-workflow-function-surface-readiness-close-v1.py` 中校验 fixture、依赖状态、`apps/radishmind-web/` 源码、样式、文档引用和 `scripts/check-repo.py` fast baseline 接入。
- 在 `apps/radishmind-web/README.md`、current focus、roadmap、capability matrix、integration contracts、任务卡入口、scripts README 和本周周志中同步状态。

## Surface Matrix

当前聚合 gate 覆盖以下可见功能面：

1. `workflow-application-detail-read`
2. `workflow-definition-detail-read`
3. `workflow-run-detail-read`
4. `workflow-blocked-action-preview-read`
5. `workflow-confirmation-placeholder-read`
6. `workflow-draft-designer-offline`
7. `workflow-draft-validation-inspector-offline`
8. `workflow-execution-plan-preview-offline`
9. `workflow-runtime-readiness-inspector-offline`

矩阵只要求这些页面继续以离线 fixture 或 fixture-derived view model 渲染，并保留 request / route / audit metadata、blocked capability preview、execution plan preview 和 runtime readiness inspector。矩阵不允许页面新增 live backend request、持久化、发布、执行、确认提交、写回、replay、数据库、生产认证或 repository adapter 能力。

## 验收口径

- 新 fixture 的 slice id 为 `workflow-function-surface-readiness-close-v1`，状态为 `workflow_function_surface_readiness_closed`。
- 新 checker 已接入 `scripts/check-repo.py`，并在 `workflow-runtime-readiness-inspector-offline-v1` 之后执行。
- checker 必须读取 10 个依赖 fixture，并确认各自状态没有漂移。
- checker 必须确认九个 `apps/radishmind-web/` workflow surface 的 builder、render component、render anchor、CSS selector 和只读 / blocked flags 仍在。
- checker 必须拒绝 live backend request、持久化、发布、执行、确认提交、业务写回、replay、生产认证、数据库和 repository adapter 的源码字面量。
- 文档必须明确：此 gate 是离线产品面 readiness 收束，不是实现触发条件满足。

## 非目标

- 不新增 Go route 或 runtime API。
- 不实现 workflow builder mutation、draft persistence、validation result persistence、execution plan persistence 或 runtime readiness persistence。
- 不实现 workflow publish、workflow executor、node executor、tool executor 或 agent loop。
- 不实现 confirmation decision、decision store、execution unlock 或业务写回。
- 不实现 durable run/result store、materialized result reader、run replay 或 run resume。
- 不实现真实数据库、repository adapter、store selector、Radish OIDC、token validation 或 production API consumer。
- 不启动 Docker、不访问网络、不下载模型、不写入 secret。

## 验证方式

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-function-surface-readiness-close-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-runtime-readiness-inspector-offline-v1.py
cd apps/radishmind-web && npm run build
./scripts/check-repo.sh --fast
```

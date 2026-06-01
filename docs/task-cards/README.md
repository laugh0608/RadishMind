# 首批任务卡

更新时间：2026-06-01

本目录用于把路线图中的高优先级任务，从“任务名”收口到“可实现、可评测、可对齐协议”的正式任务卡或前置条件清单。

当前说明：

- 当前仓库主线已经切回平台本体建设；因此这里的文档更多承担“应用面任务边界”和“真实接入前置条件”职责。
- 如果你现在要判断应该优先做什么，先读 `docs/radishmind-current-focus.md`、`docs/radishmind-capability-matrix.md` 和 `docs/radishmind-roadmap.md`，不要把这些任务卡误解成当前唯一主线。

当前已冻结的应用面任务：

## `RadishFlow`

1. [explain_diagnostics](radishflow-explain-diagnostics.md)
2. [suggest_flowsheet_edits](radishflow-suggest-flowsheet-edits.md)
3. [explain_control_plane_state](radishflow-explain-control-plane-state.md)
4. [suggest_ghost_completion](radishflow-suggest-ghost-completion.md)（附件：[PoC 进展](radishflow-suggest-ghost-completion-progress.md)、[候选上下文与样本覆盖](radishflow-suggest-ghost-completion-candidate-context.md)）
5. [接入前置条件与阻塞项](radishflow-first-real-integration.md)

## `Radish`

1. [answer_docs_question](radish-answer-docs-question.md)

## 跨项目

1. [接入前置条件总表](cross-project-integration-readiness.md)
2. [`P2 Session & Tooling` 实现前置条件](session-tooling-implementation-preconditions.md)
3. [`P2 Session & Tooling` confirmation flow design](session-tooling-confirmation-flow-design.md)
4. [`P2 Session & Tooling` independent audit records design](session-tooling-independent-audit-records.md)
5. [`P2 Session & Tooling` result materialization policy design](session-tooling-result-materialization-policy.md)
6. [`P2 Session & Tooling` executor boundary design](session-tooling-executor-boundary.md)
7. [`P2 Session & Tooling` storage backend design](session-tooling-storage-backend-design.md)
8. [`P2 Session & Tooling` close-candidate readiness rollup](session-tooling-close-candidate-readiness-rollup.md)
9. [`P2 Session & Tooling` negative regression suite readiness](session-tooling-negative-regression-suite-readiness.md)
10. [`P2 Session & Tooling` negative regression suite](session-tooling-negative-regression-suite.md)
11. [`P2 Session & Tooling` deny-by-default implementation gates](session-tooling-deny-by-default-implementation-gates.md)
12. [`P2 Session & Tooling` negative coverage rollup](session-tooling-negative-coverage-rollup.md)
13. [`P2 Session & Tooling` route smoke readiness rollup](session-tooling-route-smoke-readiness-rollup.md)
14. [`P2 Session & Tooling` short close readiness delta](session-tooling-short-close-readiness-delta.md)
15. [`P2 Session & Tooling` readiness consistency rollup](session-tooling-readiness-consistency-rollup.md)
16. [`P2 Session & Tooling` executor / storage / confirmation enablement plan](session-tooling-executor-storage-confirmation-enablement-plan.md)
17. [`P2 Session & Tooling` route negative coverage matrix](session-tooling-route-negative-coverage-matrix.md)
18. [`P2 Session & Tooling` stop-line manifest](session-tooling-stop-line-manifest.md)
19. [`P2 Session & Tooling` short close entry checklist](session-tooling-short-close-entry-checklist.md)
20. [`P2 Session & Tooling` upper-layer confirmation flow readiness](session-tooling-upper-layer-confirmation-flow-readiness.md)

## Product Platform

1. [`Control Plane / User Workspace / Workflow` v1 计划](control-plane-user-workspace-workflow-v1-plan.md)
2. [`Control Plane Read Model` v1 计划](control-plane-read-model-v1-plan.md)（`control-plane-read-model-v1`）
3. [`Control Plane Read Route Contract` v1 计划](control-plane-read-route-contract-v1-plan.md)（`control-plane-read-route-contract-v1`）
4. [`Control Plane Read Response Fixtures` v1 计划](control-plane-read-response-fixtures-v1-plan.md)（`control-plane-read-response-fixtures-v1`）
5. [`Control Plane Read Negative Contract` v1 计划](control-plane-read-negative-contract-v1-plan.md)（`control-plane-read-negative-contract-v1`）
6. [`Control Plane Read Implementation Preconditions` v1 计划](control-plane-read-implementation-preconditions-v1-plan.md)（`control-plane-read-implementation-preconditions-v1`）
7. [`Control Plane Read Fake-Store Handler Plan` v1 计划](control-plane-read-fake-store-handler-plan-v1-plan.md)（`control-plane-read-fake-store-handler-plan-v1`）
8. [`Control Plane Read Fake-Store Handler Implementation` v1 计划](control-plane-read-fake-store-handler-implementation-v1-plan.md)（`control-plane-read-fake-store-handler-implementation-v1`）
9. [`Control Plane Read Auth/DB Preconditions` v1 计划](control-plane-read-auth-db-preconditions-v1-plan.md)（`control-plane-read-auth-db-preconditions-v1`）
10. [`Control Plane Read Consumer Contract` v1 计划](control-plane-read-consumer-contract-v1-plan.md)（`control-plane-read-consumer-contract-v1`）
11. [`Control Plane Read Formal UI Boundary` v1 计划](control-plane-read-formal-ui-boundary-v1-plan.md)（`control-plane-read-formal-ui-boundary-v1`）
12. [`Control Plane Read Formal UI Implementation Readiness` v1 计划](control-plane-read-formal-ui-implementation-readiness-v1-plan.md)（`control-plane-read-formal-ui-implementation-readiness-v1`）
13. [`Control Plane Read Shared Shell` v1 计划](control-plane-read-shared-shell-v1-plan.md)（`control-plane-read-shared-shell-v1`）
14. [`Control Plane Read Admin Tenant Overview` v1 计划](control-plane-read-admin-tenant-overview-v1-plan.md)（`control-plane-read-admin-tenant-overview-v1`）
15. [`Control Plane Read Workspace Applications` v1 计划](control-plane-read-workspace-applications-v1-plan.md)（`control-plane-read-workspace-applications-v1`）
16. [`Control Plane Read Workspace API Keys` v1 计划](control-plane-read-workspace-api-keys-v1-plan.md)（`control-plane-read-workspace-api-keys-v1`）
17. [`Control Plane Read Workspace Usage Quota` v1 计划](control-plane-read-workspace-usage-quota-v1-plan.md)（`control-plane-read-workspace-usage-quota-v1`）
18. [`Control Plane Read Workspace Workflow Definitions` v1 计划](control-plane-read-workspace-workflow-definitions-v1-plan.md)（`control-plane-read-workspace-workflow-definitions-v1`）
19. [`Control Plane Read Workspace Run History` v1 计划](control-plane-read-workspace-run-history-v1-plan.md)（`control-plane-read-workspace-run-history-v1`）
20. [`Control Plane Read Admin Audit Log` v1 计划](control-plane-read-admin-audit-log-v1-plan.md)（`control-plane-read-admin-audit-log-v1`）
21. [`Control Plane Read Formal UI Readiness Close` v1 计划](control-plane-read-formal-ui-readiness-close-v1-plan.md)（`control-plane-read-formal-ui-readiness-close-v1`）

## UI 设计专题

1. [`UI Design Topic` ops surface 设计到实现计划](ui-design-ops-surface-implementation-plan.md)

## Production Ops

1. [`Production Ops Hardening` v1 计划](production-ops-hardening-v1-plan.md)
2. [`Provider Runtime & Health` v1 计划](provider-runtime-health-v1-plan.md)

## Model Adaptation

1. [`P4 Model Adaptation` v1 前置计划](model-adaptation-v1-preflight-plan.md)

配套 runbook：`training/experiments/radishmind-core-model-adaptation-v1-preflight-runbook-v0.json`

使用原则：

- 任务卡定义的是任务边界、最小输入、输出要求和评测口径，不等同于最终实现代码
- 前置条件型任务卡定义的是当前不能继续前推的阻塞项、已有门禁和后续触发条件，不等同于已经完成上层接线
- 当前平台主线已完成 `Production Ops Hardening v1` 静态边界收口，并把 `Provider Runtime & Health v1` 推进到 close candidate；任务卡不替代当前焦点、路线图和能力矩阵
- 任务卡与 [跨项目集成契约](../radishmind-integration-contracts.md) 和 [真实契约文件](../../contracts/README.md) 保持一致
- 若未来实现发现字段命名或结构需要调整，应先同步更新任务卡和契约，再改实现
- 当前阶段优先保证“状态优先、结构化输出、显式风险分级，以及对会写回真相源的动作保留人工确认”
- 编辑器内 ghost 补全和正式候选 patch 必须分开建模，不应混用为同一任务

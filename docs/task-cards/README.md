# 首批任务卡

更新时间：2026-06-23

本目录用于把路线图中的高优先级任务，从“任务名”收口到“可实现、可评测、可对齐协议”的正式任务卡或前置条件清单。

当前说明：

- 当前仓库主线已经切回平台本体建设；因此这里的文档更多承担“应用面任务边界”和“真实接入前置条件”职责。
- 如果你现在要判断应该优先做什么，先读 `docs/radishmind-current-focus.md`、`docs/radishmind-capability-matrix.md` 和 `docs/radishmind-roadmap.md`，不要把这些任务卡误解成当前唯一主线。
- 2026-06-14 起，长期功能设计默认写入 `docs/features/`；任务卡只用于具体实现批次、前置条件或高风险边界。普通只读展示、文案、布局和 evidence 组织不再默认逐项新增任务卡。

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
22. [`Control Plane Read Dev Live Consumer` v1 计划](control-plane-read-dev-live-consumer-v1-plan.md)（`control-plane-read-dev-live-consumer-v1`）
23. [`Control Plane Read Auth/Store Transition Preconditions` v1 计划](control-plane-read-auth-store-transition-preconditions-v1-plan.md)（`control-plane-read-auth-store-transition-preconditions-v1`）
24. [`Control Plane Read Repository Contract Preconditions` v1 计划](control-plane-read-repository-contract-preconditions-v1-plan.md)（`control-plane-read-repository-contract-preconditions-v1`）
25. [`Control Plane Read Disabled Database Guard` v1 计划](control-plane-read-disabled-database-guard-v1-plan.md)（`control-plane-read-disabled-database-guard-v1`）
26. [`Control Plane Read Repository Contract Smoke` v1 计划](control-plane-read-repository-contract-smoke-v1-plan.md)（`control-plane-read-repository-contract-smoke-v1`）
27. [`Control Plane Read Repository Implementation Readiness` v1 计划](control-plane-read-repository-implementation-readiness-v1-plan.md)（`control-plane-read-repository-implementation-readiness-v1`）
28. [`Control Plane Read Store Selection Readiness` v1 计划](control-plane-read-store-selection-readiness-v1-plan.md)（`control-plane-read-store-selection-readiness-v1`）
29. [`Control Plane Read Schema Migration Readiness` v1 计划](control-plane-read-schema-migration-readiness-v1-plan.md)（`control-plane-read-schema-migration-readiness-v1`）
30. [`Control Plane Read Repository Contract Types Readiness` v1 计划](control-plane-read-repository-contract-types-readiness-v1-plan.md)（`control-plane-read-repository-contract-types-readiness-v1`）
31. [`Control Plane Read Repository Contract Types Implementation` v1 计划](control-plane-read-repository-contract-types-implementation-v1-plan.md)（`control-plane-read-repository-contract-types-implementation-v1`）
32. [`Control Plane Read Repository Contract Smoke Runner Readiness` v1 计划](control-plane-read-repository-contract-smoke-runner-readiness-v1-plan.md)（`control-plane-read-repository-contract-smoke-runner-readiness-v1`）
33. [`Control Plane Read Repository Contract Smoke Runner Implementation` v1 计划](control-plane-read-repository-contract-smoke-runner-implementation-v1-plan.md)（`control-plane-read-repository-contract-smoke-runner-implementation-v1`）
34. [`Control Plane Read Repository Interface Readiness` v1 计划](control-plane-read-repository-interface-readiness-v1-plan.md)（`control-plane-read-repository-interface-readiness-v1`）
35. [`Control Plane Read Repository Adapter Implementation Readiness Refresh` v1 计划](control-plane-read-repository-adapter-implementation-readiness-refresh-v1-plan.md)（`control-plane-read-repository-adapter-implementation-readiness-refresh-v1`）
36. [`Control Plane Read Store Selector Enablement Preconditions` v1 计划](control-plane-read-store-selector-enablement-preconditions-v1-plan.md)（`control-plane-read-store-selector-enablement-preconditions-v1`）
37. [`Control Plane Read Schema Migration Implementation Preconditions` v1 计划](control-plane-read-schema-migration-implementation-preconditions-v1-plan.md)（`control-plane-read-schema-migration-implementation-preconditions-v1`）
38. [`Control Plane Read Repository Adapter Implementation Plan` v1 计划](control-plane-read-repository-adapter-implementation-plan-v1-plan.md)（`control-plane-read-repository-adapter-implementation-plan-v1`）
39. [`Control Plane Read Schema Artifact Manifest Readiness` v1 计划](control-plane-read-schema-artifact-manifest-readiness-v1-plan.md)（`control-plane-read-schema-artifact-manifest-readiness-v1`）
40. [`Control Plane Read Store Selector Smoke Readiness` v1 计划](control-plane-read-store-selector-smoke-readiness-v1-plan.md)（`control-plane-read-store-selector-smoke-readiness-v1`）
41. [`Control Plane Read Production Auth Readiness` v1 计划](control-plane-read-production-auth-readiness-v1-plan.md)（`control-plane-read-production-auth-readiness-v1`）
42. [`Control Plane Read Adapter Smoke Readiness` v1 计划](control-plane-read-adapter-smoke-readiness-v1-plan.md)（`control-plane-read-adapter-smoke-readiness-v1`）
43. [`Control Plane Read Implementation Trigger Review` v1 计划](control-plane-read-implementation-trigger-review-v1-plan.md)（`control-plane-read-implementation-trigger-review-v1`）
44. [`Workflow / Agent Runtime Function Surface` v1 计划](workflow-agent-runtime-function-surface-v1-plan.md)（`workflow-agent-runtime-function-surface-v1`）
45. [`Workflow Function Surface Boundary` v1 计划](workflow-function-surface-boundary-v1-plan.md)（`workflow-function-surface-boundary-v1`）
46. [`Workflow Definition Detail Read` v1 计划](workflow-definition-detail-read-v1-plan.md)（`workflow-definition-detail-read-v1`）
47. [`Workflow Run Detail Read` v1 计划](workflow-run-detail-read-v1-plan.md)（`workflow-run-detail-read-v1`）
48. [`Workflow Blocked Action Preview` v1 计划](workflow-blocked-action-preview-v1-plan.md)（`workflow-blocked-action-preview-v1`）
49. [`Workflow Application Detail Read` v1 计划](workflow-application-detail-read-v1-plan.md)（`workflow-application-detail-read-v1`）
50. [`Workflow Confirmation Placeholder Read` v1 计划](workflow-confirmation-placeholder-read-v1-plan.md)（`workflow-confirmation-placeholder-read-v1`）
51. [`Workflow Draft Designer Offline` v1 计划](workflow-draft-designer-offline-v1-plan.md)（`workflow-draft-designer-offline-v1`）
52. [`Workflow Draft Validation Inspector Offline` v1 计划](workflow-draft-validation-inspector-offline-v1-plan.md)（`workflow-draft-validation-inspector-offline-v1`）
53. [`Workflow Execution Plan Preview Offline` v1 计划](workflow-execution-plan-preview-offline-v1-plan.md)（`workflow-execution-plan-preview-offline-v1`）
54. [`Workflow Runtime Readiness Inspector Offline` v1 计划](workflow-runtime-readiness-inspector-offline-v1-plan.md)（`workflow-runtime-readiness-inspector-offline-v1`）
55. [`Workflow Function Surface Readiness Close` v1 计划](workflow-function-surface-readiness-close-v1-plan.md)（`workflow-function-surface-readiness-close-v1`）
56. [`Product Surface Readiness / Implementation Trigger Recheck` v1 计划](product-surface-readiness-implementation-trigger-recheck-v1-plan.md)（`product-surface-readiness-implementation-trigger-recheck-v1`）
57. [`Control Plane Read Schema Artifact Evidence` v1 计划](control-plane-read-schema-artifact-evidence-v1-plan.md)（`control-plane-read-schema-artifact-evidence-v1`）
58. [`Control Plane Read Implementation Entry Review` v1 计划](control-plane-read-implementation-entry-review-v1-plan.md)（`control-plane-read-implementation-entry-review-v1`）
59. [`Product Surface Usage Gap Triage` v1 计划](product-surface-usage-gap-triage-v1-plan.md)（`product-surface-usage-gap-triage-v1`）
60. [`Control Plane Durable Read Foundation` v1 任务卡](control-plane-durable-read-foundation-v1-plan.md)（`control-plane-durable-read-foundation-v1`）
61. [`Workflow Saved Draft` v1 implementation 任务卡](workflow-saved-draft-v1-implementation-plan.md)（`workflow-saved-draft-v1-implementation`）
62. [`Workflow Saved Draft Consumer Integration` v1 任务卡](workflow-saved-draft-consumer-integration-v1-plan.md)（`workflow-saved-draft-consumer-integration-v1`）
63. [`Workflow Draft Editing Entry` v1 任务卡](workflow-draft-editing-entry-v1-plan.md)（`workflow-draft-editing-entry-v1`）
64. [`User Workspace Draft Creation` v1 任务卡](user-workspace-draft-creation-v1-plan.md)（`user-workspace-draft-creation-v1`）
65. [`User Workspace Saved Draft List` v1 任务卡](user-workspace-saved-draft-list-v1-plan.md)（`user-workspace-saved-draft-list-v1`）
66. [`Workflow Draft Designer Editing Model` v2 任务卡](workflow-draft-designer-editing-model-v2-plan.md)（`workflow-draft-designer-editing-model-v2`）
67. [`Workflow Draft Node Attribute Editing Model` v1 任务卡](workflow-draft-node-attribute-editing-model-v1-plan.md)（`workflow-draft-node-attribute-editing-model-v1`）
68. [`Workflow Review Handoff Active Draft` v1 任务卡](workflow-review-handoff-active-draft-v1-plan.md)（`workflow-review-handoff-active-draft-v1`）
69. [`Workflow Saved Draft Durable Store Preconditions` v1 任务卡](workflow-saved-draft-durable-store-preconditions-v1-plan.md)（`workflow-saved-draft-durable-store-preconditions-v1`）
70. [`Workflow Saved Draft Repository Contract Preconditions` v1 任务卡](workflow-saved-draft-repository-contract-preconditions-v1-plan.md)（`workflow-saved-draft-repository-contract-preconditions-v1`）
71. [`Workflow Saved Draft Schema / Migration Preconditions` v1 任务卡](workflow-saved-draft-schema-migration-preconditions-v1-plan.md)（`workflow-saved-draft-schema-migration-preconditions-v1`）
72. [`Workflow Saved Draft Auth Context Preconditions` v1 任务卡](workflow-saved-draft-auth-context-preconditions-v1-plan.md)（`workflow-saved-draft-auth-context-preconditions-v1`）
73. [`Workflow Saved Draft Store Selector Enablement Preconditions` v1 任务卡](workflow-saved-draft-store-selector-enablement-preconditions-v1-plan.md)（`workflow-saved-draft-store-selector-enablement-preconditions-v1`）
74. [`Workflow Saved Draft Schema Artifact Evidence` v1 任务卡](workflow-saved-draft-schema-artifact-evidence-v1-plan.md)（`workflow-saved-draft-schema-artifact-evidence-v1`）
75. [`Workflow Saved Draft Store Selector Smoke Readiness` v1 任务卡](workflow-saved-draft-store-selector-smoke-readiness-v1-plan.md)（`workflow-saved-draft-store-selector-smoke-readiness-v1`）
76. [`Workflow Saved Draft Repository Contract Smoke` v1 任务卡](workflow-saved-draft-repository-contract-smoke-v1-plan.md)（`workflow-saved-draft-repository-contract-smoke-v1`）
77. [`Workflow Saved Draft Repository Contract Smoke Runner Readiness` v1 任务卡](workflow-saved-draft-repository-contract-smoke-runner-readiness-v1-plan.md)（`workflow-saved-draft-repository-contract-smoke-runner-readiness-v1`）
78. [`Workflow Saved Draft Repository Contract Smoke Runner Implementation` v1 任务卡](workflow-saved-draft-repository-contract-smoke-runner-implementation-v1-plan.md)（`workflow-saved-draft-repository-contract-smoke-runner-implementation-v1`）
79. [`Workflow Saved Draft Repository Adapter Implementation Plan` v1 任务卡](workflow-saved-draft-repository-adapter-implementation-plan-v1-plan.md)（`workflow-saved-draft-repository-adapter-implementation-plan-v1`）
80. [`Workflow Saved Draft Schema Artifact Manifest` v1 任务卡](workflow-saved-draft-schema-artifact-manifest-v1-plan.md)（`workflow-saved-draft-schema-artifact-manifest-v1`）
81. [`Workflow Saved Draft Adapter Smoke Readiness` v1 任务卡](workflow-saved-draft-adapter-smoke-readiness-v1-plan.md)（`workflow-saved-draft-adapter-smoke-readiness-v1`）
82. [`Workflow Saved Draft Store Selector Implementation Entry Review` v1 任务卡](workflow-saved-draft-store-selector-implementation-entry-review-v1-plan.md)（`workflow-saved-draft-store-selector-implementation-entry-review-v1`）
83. [`Workflow Saved Draft Schema Artifact Materialization Review` v1 任务卡](workflow-saved-draft-schema-artifact-materialization-review-v1-plan.md)（`workflow-saved-draft-schema-artifact-materialization-review-v1`）
84. [`Workflow Saved Draft Store Selector Implementation` v1 任务卡](workflow-saved-draft-store-selector-implementation-v1-plan.md)（`workflow-saved-draft-store-selector-smoke-v1`）
85. [`Workflow Saved Draft Schema Artifact Materialization` v1 任务卡](workflow-saved-draft-schema-artifact-materialization-v1-plan.md)（`workflow-saved-draft-schema-artifact-materialization-v1`）
86. [`Workflow Saved Draft Production Auth Readiness` v1 任务卡](workflow-saved-draft-production-auth-readiness-v1-plan.md)（`workflow-saved-draft-production-auth-readiness-v1`）
87. [`Workflow Saved Draft Repository Adapter Implementation Entry Review` v1 任务卡](workflow-saved-draft-repository-adapter-implementation-entry-review-v1-plan.md)（`workflow-saved-draft-repository-adapter-implementation-entry-review-v1`）
88. [`Workflow Saved Draft Repository Adapter Implementation` v1 任务卡](workflow-saved-draft-repository-adapter-implementation-v1-plan.md)（`workflow-saved-draft-repository-adapter-implementation-v1`）
89. [`Workflow Saved Draft Adapter Smoke` v1 任务卡](workflow-saved-draft-adapter-smoke-v1-plan.md)（`workflow-saved-draft-adapter-smoke-v1`）
90. [`Workflow Saved Draft Production Auth Runtime` v1 任务卡](workflow-saved-draft-production-auth-runtime-v1-plan.md)（`workflow-saved-draft-production-auth-runtime-v1`）
91. [`Workflow Saved Draft Repository Mode Enablement` v1 任务卡](workflow-saved-draft-repository-mode-enablement-v1-plan.md)（`workflow-saved-draft-repository-mode-enablement-v1`）
92. [`Workflow Saved Draft Schema Migration Runner Readiness` v1 任务卡](workflow-saved-draft-schema-migration-runner-readiness-v1-plan.md)（`workflow-saved-draft-schema-migration-runner-readiness-v1`）
93. [`Workflow Saved Draft Schema Migration Runner Implementation Entry Review` v1 任务卡](workflow-saved-draft-schema-migration-runner-implementation-entry-review-v1-plan.md)（`workflow-saved-draft-schema-migration-runner-implementation-entry-review-v1`）
94. [`Workflow Saved Draft Database Connection / Schema Marker Preconditions` v1 任务卡](workflow-saved-draft-database-connection-schema-marker-preconditions-v1-plan.md)（`workflow-saved-draft-database-connection-schema-marker-preconditions-v1`）
95. [`Workflow Saved Draft Database Connection Provider Implementation Entry Review` v1 任务卡](workflow-saved-draft-database-connection-provider-implementation-entry-review-v1-plan.md)（`workflow-saved-draft-database-connection-provider-implementation-entry-review-v1`）
96. [`Workflow Saved Draft Database Secret Resolver Readiness` v1 任务卡](workflow-saved-draft-database-secret-resolver-readiness-v1-plan.md)（`workflow-saved-draft-database-secret-resolver-readiness-v1`）
97. [`Workflow Saved Draft Database Secret Resolver Implementation Entry Review` v1 任务卡](workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1-plan.md)（`workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1`）
98. [`Workflow Saved Draft Repository Mode Runtime Boundary Review` v1 任务卡](workflow-saved-draft-repository-mode-runtime-boundary-review-v1-plan.md)（`workflow-saved-draft-repository-mode-runtime-boundary-review-v1`）
99. [`Workflow Saved Draft Schema Marker / Migration Runner Readiness Refresh` v1 任务卡](workflow-saved-draft-schema-marker-migration-runner-readiness-refresh-v1-plan.md)（`workflow-saved-draft-schema-marker-migration-runner-readiness-refresh-v1`）
100. [`Workflow Saved Draft Schema Marker Contract Implementation Entry Review` v1 任务卡](workflow-saved-draft-schema-marker-contract-implementation-entry-review-v1-plan.md)（`workflow-saved-draft-schema-marker-contract-implementation-entry-review-v1`）
101. [`Workflow Saved Draft Manual Migration Runner Implementation Entry Refresh` v1 任务卡](workflow-saved-draft-manual-migration-runner-implementation-entry-refresh-v1-plan.md)（`workflow-saved-draft-manual-migration-runner-implementation-entry-refresh-v1`）
102. [`Workflow Saved Draft Database Connection Provider Implementation Entry Refresh` v1 任务卡](workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v1-plan.md)（`workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v1`）
103. [`Workflow Saved Draft Database Driver / DSN / TLS Policy Readiness` v1 任务卡](workflow-saved-draft-database-driver-dsn-tls-policy-readiness-v1-plan.md)（`workflow-saved-draft-database-driver-dsn-tls-policy-readiness-v1`）
104. [`Workflow Saved Draft Database Role Policy Readiness` v1 任务卡](workflow-saved-draft-database-role-policy-readiness-v1-plan.md)（`workflow-saved-draft-database-role-policy-readiness-v1`）
105. [`Workflow Saved Draft Database Connection Smoke Strategy` v1 任务卡](workflow-saved-draft-database-connection-smoke-strategy-v1-plan.md)（`workflow-saved-draft-database-connection-smoke-strategy-v1`）
106. [`Workflow Saved Draft Database Connection Lifecycle Readiness` v1 任务卡](workflow-saved-draft-database-connection-lifecycle-readiness-v1-plan.md)（`workflow-saved-draft-database-connection-lifecycle-readiness-v1`）
107. [`Workflow Saved Draft Database Connection Provider Implementation Entry Refresh v2` 任务卡](workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v2-plan.md)（`workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v2`）
108. [`Radish OIDC Token / Membership Readiness` v1 任务卡](radish-oidc-token-membership-readiness-v1-plan.md)（`radish-oidc-token-membership-readiness-v1`）
109. [`Radish OIDC Token / Membership Implementation Entry Review` v1 任务卡](radish-oidc-token-membership-implementation-entry-review-v1-plan.md)（`radish-oidc-token-membership-implementation-entry-review-v1`）
110. [`Radish OIDC Token / Membership Upstream Evidence Refresh` v1 任务卡](radish-oidc-token-membership-upstream-evidence-refresh-v1-plan.md)（`radish-oidc-token-membership-upstream-evidence-refresh-v1`）
111. [`Workflow Saved Draft Schema Marker Runtime Dependency Refresh` v1 任务卡](workflow-saved-draft-schema-marker-runtime-dependency-refresh-v1-plan.md)（`workflow-saved-draft-schema-marker-runtime-dependency-refresh-v1`）
112. [`Workflow Saved Draft Database Secret Resolver Runtime Dependency Refresh` v1 任务卡](workflow-saved-draft-database-secret-resolver-runtime-dependency-refresh-v1-plan.md)（`workflow-saved-draft-database-secret-resolver-runtime-dependency-refresh-v1`）
113. [`Workflow Saved Draft Token Validation Schema / Auth Middleware Runtime Entry Review` v1 任务卡](workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1-plan.md)（`workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1`）
114. [`Workflow Saved Draft Token Validation Schema Task Card Readiness` v1 任务卡](workflow-saved-draft-token-validation-schema-task-card-readiness-v1-plan.md)（`workflow-saved-draft-token-validation-schema-task-card-readiness-v1`）
115. [`Workflow Saved Draft Token Validation Schema Implementation` v1 任务卡](workflow-saved-draft-token-validation-schema-implementation-v1-plan.md)（`workflow-saved-draft-token-validation-schema-implementation-v1`）
116. [`Workflow Saved Draft Token Validation Schema Artifact Implementation` v1 任务卡](workflow-saved-draft-token-validation-schema-artifact-implementation-v1-plan.md)（`workflow-saved-draft-token-validation-schema-artifact-implementation-v1`）
117. [`Workflow Saved Draft Auth Middleware / Membership Adapter Task Card Entry Readiness` v1 任务卡](workflow-saved-draft-auth-middleware-membership-adapter-task-card-entry-readiness-v1-plan.md)（`workflow-saved-draft-auth-middleware-membership-adapter-task-card-entry-readiness-v1`）
118. [`Workflow Saved Draft Negative Auth Smoke Runtime Readiness` v1 任务卡](workflow-saved-draft-negative-auth-smoke-runtime-readiness-v1-plan.md)（`workflow-saved-draft-negative-auth-smoke-runtime-readiness-v1`）
119. [`Workflow Node Designer Surface Implementation` v1 任务卡](workflow-node-designer-surface-implementation-v1-plan.md)（`workflow-node-designer-surface-implementation-v1`）
120. [`Workflow Node Designer Saved Draft Mapping Implementation` v1 任务卡](workflow-node-designer-saved-draft-mapping-implementation-v1-plan.md)（`workflow-node-designer-saved-draft-mapping-implementation-v1`）
121. [`Workflow Node Designer Review Handoff` v1 任务卡](workflow-node-designer-review-handoff-v1-plan.md)（`workflow-node-designer-review-handoff-v1`）
122. [`Workflow Node Designer Persisted Layout` v1 任务卡](workflow-node-designer-persisted-layout-v1-plan.md)（`workflow-node-designer-persisted-layout-v1`）
123. [`Workflow Node Designer Edge Editing Save Preconditions` v1 任务卡](workflow-node-designer-edge-editing-save-preconditions-v1-plan.md)（`workflow-node-designer-edge-editing-save-preconditions-v1`）
124. [`Workflow Node Designer Controlled Edge Mutation Implementation` v1 任务卡](workflow-node-designer-controlled-edge-mutation-implementation-v1-plan.md)（`workflow-node-designer-controlled-edge-mutation-implementation-v1`）
125. [`Workflow Node Designer Layout Review Findings` v1 任务卡](workflow-node-designer-layout-review-findings-v1-plan.md)（`workflow-node-designer-layout-review-findings-v1`）
126. [`Workflow Node Designer Builder Interaction Polish` v1 任务卡](workflow-node-designer-builder-interaction-polish-v1-plan.md)（`workflow-node-designer-builder-interaction-polish-v1`）
127. [`Workflow Node Designer Validation Overlay Navigation` v1 任务卡](workflow-node-designer-validation-overlay-navigation-v1-plan.md)（`workflow-node-designer-validation-overlay-navigation-v1`）
128. [`Workflow Node Designer Graph Review Handoff Refinement` v1 任务卡](workflow-node-designer-graph-review-handoff-refinement-v1-plan.md)（`workflow-node-designer-graph-review-handoff-refinement-v1`）

## UI 设计专题

1. [`UI Design Topic` ops surface 设计到实现计划](ui-design-ops-surface-implementation-plan.md)

## Production Ops

1. [`Production Ops Hardening` v1 计划](production-ops-hardening-v1-plan.md)
2. [`Production Secret Backend Implementation` v1 计划](production-secret-backend-implementation-v1-plan.md)
3. [`Production Secret Backend Config / Secret Ref Readiness` v1 计划](production-secret-backend-config-secret-ref-readiness-v1-plan.md)（`production-secret-backend-config-secret-ref-readiness-v1`）
4. [`Production Secret Backend Provider Profile Secret Binding Readiness` v1 计划](production-secret-backend-provider-profile-secret-binding-readiness-v1-plan.md)（`production-secret-backend-provider-profile-secret-binding-readiness-v1`）
5. [`Production Secret Backend Secret Resolver Interface Disabled Readiness` v1 计划](production-secret-backend-secret-resolver-interface-disabled-readiness-v1-plan.md)（`production-secret-backend-secret-resolver-interface-disabled-readiness-v1`）
6. [`Production Secret Backend Operator Runbook / Negative Gates Readiness` v1 计划](production-secret-backend-operator-runbook-negative-gates-readiness-v1-plan.md)（`production-secret-backend-operator-runbook-negative-gates-readiness-v1`）
7. [`Production Secret Backend Rotation / Audit Policy Readiness` v1 计划](production-secret-backend-rotation-audit-policy-readiness-v1-plan.md)（`production-secret-backend-rotation-audit-policy-readiness-v1`）
8. [`Production Secret Backend Test Fixture Strategy / Fake Resolver Implementation Entry Review` v1 计划](production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1-plan.md)（`production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1`）
9. [`Production Secret Backend Fake Resolver Contract / No Secret Leakage Smoke Strategy` v1 计划](production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1-plan.md)（`production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1`）
10. [`Production Secret Backend Fake Resolver Implementation Task Card Entry Readiness Review` v1 计划](production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1-plan.md)（`production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1` / `fake_resolver_implementation_task_card_entry_readiness_review_defined`）
11. [`Production Secret Backend Fake Resolver Implementation` v1 计划](production-secret-backend-fake-resolver-implementation-v1-plan.md)（`production-secret-backend-fake-resolver-implementation-v1` / `fake_resolver_implementation_task_card_defined`）
12. [`Production Secret Backend Fake Resolver Runtime Implementation Entry Review` v1 计划](production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1-plan.md)（`production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1` / `fake_resolver_runtime_implementation_entry_review_defined`）
13. [`Production Secret Backend Fake Resolver Runtime Implementation` v1 计划](production-secret-backend-fake-resolver-runtime-implementation-v1-plan.md)（`production-secret-backend-fake-resolver-runtime-implementation-v1` / `fake_resolver_runtime_test_only_implemented`）
14. [`Production Secret Backend Real Resolver Runtime Preconditions` v1 计划](production-secret-backend-real-resolver-runtime-preconditions-v1-plan.md)（`production-secret-backend-real-resolver-runtime-preconditions-v1` / `real_resolver_runtime_preconditions_defined`）
15. [`Production Secret Backend Real Resolver Runtime Implementation Entry Review` v1 计划](production-secret-backend-real-resolver-runtime-implementation-entry-review-v1-plan.md)（`production-secret-backend-real-resolver-runtime-implementation-entry-review-v1` / `real_resolver_runtime_implementation_entry_review_defined`）
16. [`Production Secret Backend Resolver Backend Profile Selection Readiness` v1 计划](production-secret-backend-resolver-backend-profile-selection-readiness-v1-plan.md)（`production-secret-backend-resolver-backend-profile-selection-readiness-v1` / `resolver_backend_profile_selection_readiness_defined`）
17. [`Production Secret Backend Real Resolver No Secret Leakage Smoke Runtime Strategy` v1 计划](production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1-plan.md)（`production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1` / `real_resolver_no_secret_leakage_smoke_runtime_strategy_defined`）
18. [`Production Secret Backend Real Resolver No Secret Leakage Smoke Runtime Implementation Entry Review` v1 计划](production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1-plan.md)（`production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1` / `real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_defined`）
19. [`Production Secret Backend Credential Handle Runtime Boundary Readiness` v1 计划](production-secret-backend-credential-handle-runtime-boundary-readiness-v1-plan.md)（`production-secret-backend-credential-handle-runtime-boundary-readiness-v1` / `credential_handle_runtime_boundary_readiness_defined`）
20. [`Production Secret Backend Operator Approval Runtime Evidence Readiness` v1 计划](production-secret-backend-operator-approval-runtime-evidence-readiness-v1-plan.md)（`production-secret-backend-operator-approval-runtime-evidence-readiness-v1` / `operator_approval_runtime_evidence_readiness_defined`）
21. [`Production Secret Backend Audit Store Handoff Readiness v1` 计划](production-secret-backend-audit-store-handoff-readiness-v1-plan.md)（`production-secret-backend-audit-store-handoff-readiness-v1` / `audit_store_handoff_readiness_defined`）
22. [`Production Secret Backend Resolver Backend Health Boundary Readiness` v1 计划](production-secret-backend-resolver-backend-health-boundary-readiness-v1-plan.md)（`production-secret-backend-resolver-backend-health-boundary-readiness-v1` / `resolver_backend_health_boundary_readiness_defined`）
23. [`Production Secret Backend Resolver Backend Health Runtime Implementation Entry Review` v1 计划](production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1-plan.md)（`production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1` / `resolver_backend_health_runtime_implementation_entry_review_defined`）
24. [`Production Secret Backend Audit Store Runtime Implementation Entry Review` v1 计划](production-secret-backend-audit-store-runtime-implementation-entry-review-v1-plan.md)（`production-secret-backend-audit-store-runtime-implementation-entry-review-v1` / `audit_store_runtime_implementation_entry_review_defined`）
25. [`Production Secret Backend Audit Store Contract / Event Schema Readiness` v1 计划](production-secret-backend-audit-store-contract-event-schema-readiness-v1-plan.md)（`production-secret-backend-audit-store-contract-event-schema-readiness-v1` / `audit_store_contract_event_schema_readiness_defined`）
26. [`Production Secret Backend Audit Store Runtime Implementation Entry Refresh` v2 计划](production-secret-backend-audit-store-runtime-implementation-entry-refresh-v2-plan.md)（`production-secret-backend-audit-store-runtime-implementation-entry-refresh-v2` / `audit_store_runtime_implementation_entry_refresh_v2_defined`）
27. [`Production Secret Backend Audit Store Ownership Boundary Readiness` v1 计划](production-secret-backend-audit-store-ownership-boundary-readiness-v1-plan.md)（`production-secret-backend-audit-store-ownership-boundary-readiness-v1` / `audit_store_ownership_boundary_readiness_defined`）
28. [`Production Secret Backend Audit Store Delivery / Idempotency Runtime Boundary Readiness` v1 计划](production-secret-backend-audit-store-delivery-idempotency-runtime-boundary-readiness-v1-plan.md)（`production-secret-backend-audit-store-delivery-idempotency-runtime-boundary-readiness-v1` / `audit_store_delivery_idempotency_runtime_boundary_readiness_defined`）
29. [`Production Secret Backend Audit Store Runtime Implementation Entry Refresh` v3 计划](production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3-plan.md)（`production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3` / `audit_store_runtime_implementation_entry_refresh_v3_defined`）
30. [`Production Secret Backend Audit Store Durable Backend Boundary Readiness` v1 计划](production-secret-backend-audit-store-durable-backend-boundary-readiness-v1-plan.md)（`production-secret-backend-audit-store-durable-backend-boundary-readiness-v1` / `audit_store_durable_backend_boundary_readiness_defined`）
31. [`Production Secret Backend Audit Store Writer Runtime Boundary Readiness` v1 计划](production-secret-backend-audit-store-writer-runtime-boundary-readiness-v1-plan.md)（`production-secret-backend-audit-store-writer-runtime-boundary-readiness-v1` / `audit_store_writer_runtime_boundary_readiness_defined`）
32. [`Production Secret Backend Audit Store Runtime Event Schema Materialization Readiness` v1 计划](production-secret-backend-audit-store-runtime-event-schema-materialization-readiness-v1-plan.md)（`production-secret-backend-audit-store-runtime-event-schema-materialization-readiness-v1` / `audit_store_runtime_event_schema_materialization_readiness_defined`）
33. [`Production Secret Backend Audit Store Delivery Runtime Readiness` v1 计划](production-secret-backend-audit-store-delivery-runtime-readiness-v1-plan.md)（`production-secret-backend-audit-store-delivery-runtime-readiness-v1` / `audit_store_delivery_runtime_readiness_defined`）
34. [`Production Secret Backend Audit Store Idempotency Runtime Readiness` v1 计划](production-secret-backend-audit-store-idempotency-runtime-readiness-v1-plan.md)（`production-secret-backend-audit-store-idempotency-runtime-readiness-v1` / `audit_store_idempotency_runtime_readiness_defined`）
35. [`Production Secret Backend Audit Store Runtime Implementation Entry Refresh` v4 计划](production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4-plan.md)（`production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4` / `audit_store_runtime_implementation_entry_refresh_v4_defined`）
36. [`Production Secret Backend Operator Approval Runtime Implementation Entry Review` v1 计划](production-secret-backend-operator-approval-runtime-implementation-entry-review-v1-plan.md)（`production-secret-backend-operator-approval-runtime-implementation-entry-review-v1` / `operator_approval_runtime_implementation_entry_review_defined`）
37. [`Production Secret Backend Credential Handle Runtime Implementation Entry Review` v1 计划](production-secret-backend-credential-handle-runtime-implementation-entry-review-v1-plan.md)（`production-secret-backend-credential-handle-runtime-implementation-entry-review-v1` / `credential_handle_runtime_implementation_entry_review_defined`）
38. [`Production Secret Backend Real Resolver Runtime Implementation Entry Refresh` v1 计划](production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1-plan.md)（`production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1` / `real_resolver_runtime_implementation_entry_refresh_defined`）
39. [`Production Secret Backend Production Resolver Runtime Blocker Consolidation` v1 计划](production-secret-backend-production-resolver-runtime-blocker-consolidation-v1-plan.md)（`production-secret-backend-production-resolver-runtime-blocker-consolidation-v1` / `production_resolver_runtime_blocker_consolidation_defined`）
40. [`Production Secret Backend Production Resolver Runtime Implementation Entry Refresh` v2 计划](production-secret-backend-production-resolver-runtime-implementation-entry-refresh-v2-plan.md)（`production-secret-backend-production-resolver-runtime-implementation-entry-refresh-v2` / `production_resolver_runtime_implementation_entry_refresh_v2_defined`）
41. [`Production Secret Backend Audit Store Runtime Event Schema Artifact Implementation Entry Review` v1 计划](production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-entry-review-v1-plan.md)（`production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-entry-review-v1` / `audit_store_runtime_event_schema_artifact_implementation_entry_review_defined`）
42. [`Production Secret Backend Audit Store Runtime Event Schema Artifact Implementation v1` 计划](production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1-plan.md)（`production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1` / `audit_store_runtime_event_schema_artifact_implementation_task_card_defined`）
43. [`Production Secret Backend Audit Store Runtime Event Schema Artifact v1` 计划](production-secret-backend-audit-store-runtime-event-schema-artifact-v1-plan.md)（`production-secret-backend-audit-store-runtime-event-schema-artifact-v1` / `audit_store_runtime_event_schema_artifact_implemented`）
44. [`Production Secret Backend Audit Store Runtime Blocker Matrix` v1 计划](production-secret-backend-audit-store-runtime-blocker-matrix-v1-plan.md)（`production-secret-backend-audit-store-runtime-blocker-matrix-v1` / `audit_store_runtime_blocker_matrix_defined`）
45. [`Production Secret Backend Audit Store Durable Backend Selection Readiness` v1 计划](production-secret-backend-audit-store-durable-backend-selection-readiness-v1-plan.md)（`production-secret-backend-audit-store-durable-backend-selection-readiness-v1` / `audit_store_durable_backend_selection_readiness_defined`）
46. [`Production Secret Backend Audit Store Writer Runtime Implementation Entry Review` v1 计划](production-secret-backend-audit-store-writer-runtime-implementation-entry-review-v1-plan.md)（`production-secret-backend-audit-store-writer-runtime-implementation-entry-review-v1` / `audit_store_writer_runtime_implementation_entry_review_defined`）
47. [`Production Secret Backend Audit Store Idempotency Runtime Implementation Entry Review` v1 计划](production-secret-backend-audit-store-idempotency-runtime-implementation-entry-review-v1-plan.md)（`production-secret-backend-audit-store-idempotency-runtime-implementation-entry-review-v1` / `audit_store_idempotency_runtime_implementation_entry_review_defined`）
48. [`Production Secret Backend Audit Store Delivery Runtime Implementation Entry Review` v1 计划](production-secret-backend-audit-store-delivery-runtime-implementation-entry-review-v1-plan.md)（`production-secret-backend-audit-store-delivery-runtime-implementation-entry-review-v1` / `audit_store_delivery_runtime_implementation_entry_review_defined`）
49. [`Production Secret Backend Audit Store Concrete Durable Backend Selection Review` v1 计划](production-secret-backend-audit-store-concrete-durable-backend-selection-review-v1-plan.md)（`production-secret-backend-audit-store-concrete-durable-backend-selection-review-v1` / `audit_store_concrete_durable_backend_selection_review_defined`）
50. [`Production Secret Backend Audit Store Runtime Implementation Entry Refresh` v5 计划](production-secret-backend-audit-store-runtime-implementation-entry-refresh-v5-plan.md)（`production-secret-backend-audit-store-runtime-implementation-entry-refresh-v5` / `audit_store_runtime_implementation_entry_refresh_v5_defined`）
49. [`Production Secret Backend Credential Handle Runtime Implementation Entry Refresh` v1 计划](production-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1-plan.md)（`production-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1` / `credential_handle_runtime_implementation_entry_refresh_defined`）
50. [`Production Secret Backend Operator Approval Runtime Implementation Entry Refresh` v1 计划](production-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1-plan.md)（`production-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1` / `operator_approval_runtime_implementation_entry_refresh_defined`）
51. [`Production Secret Backend Cloud Secret Service Selection Readiness` v1 计划](production-secret-backend-cloud-secret-service-selection-readiness-v1-plan.md)（`production-secret-backend-cloud-secret-service-selection-readiness-v1` / `cloud_secret_service_selection_readiness_defined`）
52. [`Production Secret Backend Resolver Backend Health Runtime Implementation Entry Refresh` v1 计划](production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1-plan.md)（`production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1` / `resolver_backend_health_runtime_implementation_entry_refresh_defined`）
53. [`Production Secret Backend Real Resolver No Secret Leakage Smoke Runtime Implementation Entry Refresh` v1 计划](production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh-v1-plan.md)（`production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh-v1` / `real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_refresh_defined`）
54. [`Provider Runtime & Health` v1 计划](provider-runtime-health-v1-plan.md)

## Model Adaptation

1. [`P4 Model Adaptation` v1 前置计划](model-adaptation-v1-preflight-plan.md)

配套 runbook：`training/experiments/radishmind-core-model-adaptation-v1-preflight-runbook-v0.json`

## Image Path

1. [`Image Adapter Handshake / Safety Gate` v1 计划](image-adapter-handshake-safety-gate-v1-plan.md)（`image-adapter-handshake-safety-gate-v1`）
2. [`Image Artifact Return Runbook / Metadata Evidence` v1 计划](image-artifact-return-runbook-evidence-v1-plan.md)（`image-artifact-return-runbook-evidence-v1`）
3. [`Image Safety Runbook Evidence` v1 计划](image-safety-runbook-evidence-v1-plan.md)（`image-safety-runbook-evidence-v1`）
4. [`Image Backend Adapter Readiness Evidence` v1 计划](image-backend-adapter-readiness-evidence-v1-plan.md)（`image-backend-adapter-readiness-evidence-v1`）
5. [`Image Artifact Runtime Mapping Readiness` v1 计划](image-artifact-runtime-mapping-readiness-v1-plan.md)（`image-artifact-runtime-mapping-readiness-v1`）
6. [`Image Artifact Runtime Mapping Implementation Entry Review` v1 计划](image-artifact-runtime-mapping-implementation-entry-review-v1-plan.md)（`image-artifact-runtime-mapping-implementation-entry-review-v1`）
7. [`Image Artifact Store / Binary Reader Boundary Readiness` v1 计划](image-artifact-store-binary-reader-boundary-readiness-v1-plan.md)（`image-artifact-store-binary-reader-boundary-readiness-v1`）
8. [`Image Artifact Runtime Mapper Implementation Plan` v1 计划](image-artifact-runtime-mapper-implementation-plan-v1-plan.md)（`image-artifact-runtime-mapper-implementation-plan-v1`）
9. [`Image Artifact Runtime Mapper Implementation Entry` v1 计划](image-artifact-runtime-mapper-implementation-entry-v1-plan.md)（`image-artifact-runtime-mapper-implementation-entry-v1`）
10. [`Image Artifact Runtime Mapper Implementation` v1 计划](image-artifact-runtime-mapper-implementation-v1-plan.md)（`image-artifact-runtime-mapper-implementation-v1`）
11. [`Image Artifact Runtime Mapper Runtime Implementation` v1 计划](image-artifact-runtime-mapper-runtime-implementation-v1-plan.md)（`image-artifact-runtime-mapper-runtime-implementation-v1`）
12. [`Image Artifact Runtime Mapper Response Consumer Integration Review` v1 计划](image-artifact-runtime-mapper-response-consumer-integration-review-v1-plan.md)（`image-artifact-runtime-mapper-response-consumer-integration-review-v1`）
13. [`Image Artifact Response Consumer Implementation Readiness` v1 计划](image-artifact-response-consumer-implementation-readiness-v1-plan.md)（`image-artifact-response-consumer-implementation-readiness-v1`）
14. [`Image Artifact Response Consumer Implementation` v1 计划](image-artifact-response-consumer-implementation-v1-plan.md)（`image-artifact-response-consumer-implementation-v1`）
15. [`Image Artifact Response Consumer Runtime Implementation` v1 计划](image-artifact-response-consumer-runtime-implementation-v1-plan.md)（`image-artifact-response-consumer-runtime-implementation-v1`）
16. [`Image Artifact Response Builder Integration Entry Review` v1 计划](image-artifact-response-builder-integration-entry-review-v1-plan.md)（`image-artifact-response-builder-integration-entry-review-v1`）
17. [`Image Artifact Response Builder Integration` v1 计划](image-artifact-response-builder-integration-v1-plan.md)（`image-artifact-response-builder-integration-v1`）
18. [`Image Artifact Response Builder Runtime Integration Entry Review` v1 计划](image-artifact-response-builder-runtime-integration-entry-review-v1-plan.md)（`image-artifact-response-builder-runtime-integration-entry-review-v1`）
19. [`Image Artifact Response Builder Runtime Integration Implementation` v1 任务卡](image-artifact-response-builder-runtime-integration-implementation-v1-plan.md)（`image-artifact-response-builder-runtime-integration-implementation-v1`）

使用原则：
- 任务卡定义的是任务边界、最小输入、输出要求和评测口径，不等同于最终实现代码
- 前置条件型任务卡定义的是当前不能继续前推的阻塞项、已有门禁和后续触发条件，不等同于已经完成上层接线
- 当前平台主线已完成 `Production Ops Hardening v1` 静态边界收口，并把 `Provider Runtime & Health v1` 推进到 close candidate；没有 Docker 运行窗口时，下一步默认回到 `Workflow / Agent Runtime Function Surface v1` 和离线 workflow 产品功能，且 `workflow-definition-detail-read-v1`、`workflow-run-detail-read-v1`、`workflow-blocked-action-preview-v1`、`workflow-application-detail-read-v1`、`workflow-confirmation-placeholder-read-v1`、`workflow-draft-designer-offline-v1`、`workflow-draft-validation-inspector-offline-v1`、`workflow-execution-plan-preview-offline-v1` 与 `workflow-runtime-readiness-inspector-offline-v1` 已固定 definition / run / blocked action / application detail / confirmation placeholder / offline draft designer / offline validation inspector / offline execution plan preview / runtime readiness inspector surface；`workflow-function-surface-readiness-close-v1` 已把上述离线 workflow surface 收束为 `workflow_function_surface_readiness_closed`；任务卡不替代当前焦点、路线图和能力矩阵
- 任务卡与 [跨项目集成契约](../radishmind-integration-contracts.md) 和 [真实契约文件](../../contracts/README.md) 保持一致
- 若未来实现发现字段命名或结构需要调整，应先同步更新任务卡和契约，再改实现
- 当前阶段优先保证“状态优先、结构化输出、显式风险分级，以及对会写回真相源的动作保留人工确认”
- 编辑器内 ghost 补全和正式候选 patch 必须分开建模，不应混用为同一任务

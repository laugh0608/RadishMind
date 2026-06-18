# RadishMind 当前推进焦点

更新时间：2026-06-18

## 文档目的

本文档用于回答“根据项目规划和开发进度，今天要做什么以推进开发”。它是新会话短入口，只保留当前阶段、最近结论、下一顺位和停止线。

功能细节默认先进入 [功能设计文档入口](features/README.md) 所定义的专题层级：产品面大方向进入 `docs/features/*.md`，具体功能和复杂页面进入对应子目录，平台横切能力进入 `docs/platform/`，外部接入进入 `docs/integrations/`。实现批次进入 `docs/task-cards/`，长验证记录进入周志、manifest、summary 或 run record。

## 当前阶段

当前处于“平台本体建设期”。`P1 Runtime Foundation` 已达到 short close，`P2 Session & Tooling Foundation` 进入 close candidate / governance-only，`P3 Local Product Shell / Ops Surface` 达到 local usable / read-only close，Production Ops 静态边界和 Provider Runtime & Health v1 已可检查。2026-06-16 已完成 `User Workspace Saved Draft List v1`、`Workflow Draft Designer Editing Model v2`、`Workflow Draft Node Attribute Editing Model v1`、`Workflow Review Handoff Active Draft v1`、`Saved Workflow Draft Repository Contract Smoke v1`、`Saved Workflow Draft Repository Contract Smoke Runner Readiness v1` 和 `Saved Workflow Draft Repository Contract Smoke Runner Implementation v1`，让用户能在 Workspace Home 查看 saved dev draft summary、恢复到 Draft Designer，在本地草案中新增、移动、删除非受保护节点，编辑节点属性，把 active draft validation / plan / readiness 汇总为可交接审查记录，并固定 future saved draft repository contract smoke、runner readiness 和 static runner implementation 的准入证据。2026-06-17 已完成 `Saved Workflow Draft Repository Adapter Implementation Plan v1`、`Saved Workflow Draft Schema Artifact Manifest v1`、`Saved Workflow Draft Adapter Smoke Readiness v1`、`Saved Workflow Draft Store Selector Implementation Entry Review v1`、`Saved Workflow Draft Schema Artifact Materialization Review v1`、`Saved Workflow Draft Store Selector Implementation v1`、`Saved Workflow Draft Schema Artifact Materialization v1`、`Saved Workflow Draft Production Auth Readiness v1` 和 `Saved Workflow Draft Repository Adapter Implementation Entry Review v1`，固定 future repository adapter implementation plan、operation adapter matrix、schema artifact manifest shape、operation predicate coverage、adapter smoke dependency gate、selector implementation entry review、schema artifact materialization review、formal store selector、静态 schema artifact materialization、production auth readiness、repository adapter implementation entry review、failure mapping、no fallback 和 no side effects。2026-06-18 已完成 `Saved Workflow Draft Repository Adapter Implementation v1`、`Saved Workflow Draft Adapter Smoke Execution v1`、`Saved Workflow Draft Production Auth Runtime v1` 和 `Saved Workflow Draft Repository Mode Enablement v1`，实现 repository interface、注入式 query executor adapter、schema preflight、auth context failure mapping、adapter unit tests、adapter smoke execution、verified auth context + workspace binding 到 repository actor context 的 runtime bridge，并固定 repository mode enablement 准入评审；当前仍不启用 repository mode，不创建 SQL migration、schema version table、migration runner、数据库连接、OIDC middleware、token validation、membership adapter 或 production API artifact。

`RadishMind` 的长期定位保持不变：它是 `Radish` 体系下的 AI 工具、工作流、模型网关和 Copilot 集成平台，不是替代上层业务真相源的大模型项目。长期产品面仍是：

1. `User Workspace`
2. `Admin Control Plane`
3. `Model Gateway / API Distribution`
4. `Workflow / Agent Runtime`
5. `Image Generation / Artifact Return`

部署方式、数据库和登录 / 授权优先参考 `Radish`；未来 RadishMind 作为 OIDC client 接入 `Radish`。参考 `Radish` 不代表默认引入 `.NET`，当前长期语言边界仍是 Go control plane / gateway、Python 模型与评测、TypeScript/Vite 前端。

## 节奏调整

2026-06-14 起，开发节奏从“按小 task card / fixture / checker 连续推进”调整为“按专题文档推进”；2026-06-15 起，专题粒度进一步拆成产品面大方向、具体功能、页面 / surface、平台能力和外部集成：

- 产品面大方向写入 `docs/features/*.md`。
- 具体功能和复杂页面 / surface 写入对应子目录，例如 `docs/features/workflow/`。
- 平台横切能力写入 `docs/platform/`。
- 外部项目、外部 backend 或真实 provider 接入写入 `docs/integrations/`。
- `docs/task-cards/` 只服务具体实现批次、前置条件或高风险边界。
- 普通 UI、文案、布局、只读 evidence 组织和使用性整理，不再默认逐项新增 task card / fixture / checker。
- 只有新增 API、执行边界、生产声明、数据格式、外部 provider 风险、schema 变化或高风险能力时，才新增专项 gate。
- 实现完成后回写对应专题文档、当前焦点和必要周志；不再把入口文档写成长流水。

这个调整的目标是让后续工作围绕“一个专题如何设计、如何交付、如何验收”推进，而不是在同一层停止线和 readiness gate 中反复打转。

## 当前已完成的关键事实

- Image Path 已完成 `coerce_response_document` metadata-only response builder runtime integration，状态为 `image_artifact_response_builder_runtime_integration_implemented`。它只把 request artifact metadata 合并为现有 `CopilotResponse.citations` artifact citation，不改 response schema，不接 artifact store / binary reader / public URL / backend adapter，不调用真实生图 backend。
- Control Plane Read 已完成 `control-plane-durable-read-foundation-v1`，状态为 `durable_read_foundation_implemented`。七条 read handlers 已通过 `ControlPlaneReadRepository` interface 消费 fake store，response envelope、fake auth 和 dev-only 边界保持不变。
- Workflow v1 的普通离线审查流已覆盖 Workspace Home、Review Workspace、Review Handoff、surface overview、scenario inspector、draft designer、validation inspector、execution plan preview 和 runtime readiness inspector；`Saved Workflow Draft v1` 已完成首个 platform Go domain service、dev-only HTTP route + web consumer，并补 route contract / consumer smoke / `version_conflict` 状态；`Workflow Draft Editing Entry v1`、`User Workspace Draft Creation v1`、`User Workspace Saved Draft List v1`、`Workflow Draft Designer Editing Model v2`、`Workflow Draft Node Attribute Editing Model v1` 和 `Workflow Review Handoff Active Draft v1` 已覆盖本地草案创建、保存、恢复、结构编辑、节点属性编辑和 advisory-only handoff；`workflow-saved-draft-durable-store-preconditions-v1` 到 `workflow-saved-draft-repository-adapter-implementation-entry-review-v1` 已固定 durable store 迁移证据链、static runner、adapter plan、schema manifest、adapter smoke readiness、formal store selector、静态 schema artifact 和 production auth readiness；`workflow-saved-draft-repository-adapter-implementation-v1` 已固定 `draft_repository_adapter_implemented`，实现 `SavedWorkflowDraftRepository` interface、注入式 query executor adapter、schema preflight、auth context failure mapping、adapter unit tests 和 implementation guard；`workflow-saved-draft-adapter-smoke-v1` 已固定 `draft_adapter_smoke_executed`，用 static contract smoke cases、repository adapter 和 injected fake query executor 执行 save / read / list adapter smoke；`workflow-saved-draft-production-auth-runtime-v1` 已固定 `draft_production_auth_runtime_bridge_implemented`，实现 verified auth context + workspace binding 到 repository actor context 的投影；`workflow-saved-draft-repository-mode-enablement-v1` 已固定 `draft_repository_mode_enablement_review_defined`，确认 repository mode 仍不启用。当前能力只覆盖开发态 save / read / validate / list / restore、版本冲突、失败语义、sanitized response、sample / saved record 区分、no sample fallback、受控本地编辑、本地草案创建、本地结构编辑、节点属性编辑、active draft handoff record、durable store 迁移证据链、formal store selector、静态 schema artifact、production auth readiness、repository adapter boundary、adapter smoke execution、production auth runtime bridge、repository mode enablement 准入评审和 no side effects tests，不代表 durable persistence、production API、publish、run 或 executor ready。
- Model Gateway 和 Admin Control Plane 已完成第一版普通离线 evidence / readiness 组织层。
- Provider Runtime & Health v1 已完成 `provider-capability-matrix-v1`、`provider-health-smoke-v1`、`provider-selection-policy-v1`、`provider-retry-fallback-policy-v1` 和 `provider-runtime-docs-refresh`，由 `provider-capability-matrix-v1.json` / `check-provider-capability-matrix.py`、`provider-health-smoke-v1.json` / `check-provider-health-smoke.py`、`provider-selection-policy-v1.json` / `check-provider-selection-policy.py`、`provider-retry-fallback-policy-v1.json` / `check-provider-retry-fallback-policy.py` 和 `check-provider-runtime-docs-refresh.py` 固定证据；当前进入 close candidate，不继续默认新增 provider 同层小切片，也不把 provider health 写成 production readiness。
- 当前 repository adapter implementation 和 adapter smoke execution 已满足 interface / adapter / unit test / adapter smoke 边界，但没有数据库 / OIDC / repository mode runtime trigger satisfied；SQL migration、schema version table、migration runner、Radish OIDC middleware、token validation、production API consumer、API key lifecycle、quota enforcement、billing、workflow executor、confirmation、writeback 和 replay 都仍未打开。formal store config、store selector implementation 和静态 schema artifact materialization 已打开；selector 只允许 `memory_dev` 成功，reserved / unknown mode 必须 fail closed。

## 已完成证据锚点

本节只保留仓库门禁需要的短锚点，完整细节仍在对应专题文档、任务卡、fixture、checker 和周志中。

- `Control Plane / User Workspace / Workflow v1`：`product-surface-v1-boundary`、`product-surface-v1-boundary.json`、`check-product-surface-v1-boundary.py`；`control-plane-data-boundary`、`control-plane-data-boundary.json`、`check-control-plane-data-boundary.py`；`radish-oidc-client-preconditions`、`radish-oidc-client-preconditions.json`、`check-radish-oidc-client-preconditions.py`；`gateway-api-key-quota-readiness`、`gateway-api-key-quota-readiness.json`、`check-gateway-api-key-quota-readiness.py`；`workflow-definition-run-record-boundary`、`workflow-definition-run-record-boundary.json`、`check-workflow-definition-run-record-boundary.py`。
- Control Plane Read：`control-plane-read-model-v1`、`control-plane-read-model-v1.json`、`check-control-plane-read-model-v1.py`；`control-plane-read-route-contract-v1`、`control-plane-read-route-contract-v1.json`、`check-control-plane-read-route-contract-v1.py`；`control-plane-read-response-fixtures-v1`、`control-plane-read-response-fixtures-v1.json`、`check-control-plane-read-response-fixtures-v1.py`；`control-plane-read-negative-contract-v1`、`control-plane-read-negative-contract-v1.json`、`check-control-plane-read-negative-contract-v1.py`；`control-plane-read-implementation-preconditions-v1` 和 negative route smoke readiness。
- Control Plane Read implementation：`control-plane-read-fake-store-handler-plan-v1`、fake-store-backed read handler plan、`control-plane-read-fake-store-handler-implementation-v1`、fake-store-backed read handler implementation；`control-plane-read-auth-db-preconditions-v1`、`control-plane-read-auth-db-preconditions-v1.json`、`check-control-plane-read-auth-db-preconditions-v1.py`；`control-plane-read-consumer-contract-v1`、`control-plane-read-consumer-contract-v1.json`、`check-control-plane-read-consumer-contract-v1.py`；`control-plane-read-formal-ui-boundary-v1`、`control-plane-read-formal-ui-boundary-v1.json`、`check-control-plane-read-formal-ui-boundary-v1.py`；`control-plane-read-formal-ui-implementation-readiness-v1`、`control-plane-read-formal-ui-implementation-readiness-v1.json`、`check-control-plane-read-formal-ui-implementation-readiness-v1.py`。这些证据不直接实现 OIDC、数据库、API key / quota、workflow executor、confirmation、writeback 或 replay，也不直接实现数据库、OIDC、executor、confirmation、writeback 或 replay。
- Control Plane Read surfaces：`control-plane-read-shared-shell-v1`、`apps/radishmind-web/`、`control-plane-read-admin-tenant-overview-v1`、admin-tenant-overview、`control-plane-read-workspace-applications-v1`、workspace-applications、`control-plane-read-workspace-api-keys-v1`、workspace-api-keys、`control-plane-read-workspace-usage-quota-v1`、workspace-usage-quota、`control-plane-read-workspace-workflow-definitions-v1`、workspace-workflow-definitions、`control-plane-read-workspace-run-history-v1`、workspace-run-history、`control-plane-read-admin-audit-log-v1`、admin-audit-log、`control-plane-read-formal-ui-readiness-close-v1`、surface matrix、`control-plane-read-dev-live-consumer-v1`、dev-only live read consumer、`control-plane-read-product-sample-consistency-v1`、response fixture 和 Go fake store。
- Control Plane Read store transition：`control-plane-read-auth-store-transition-preconditions-v1`、auth/store transition preconditions、`control-plane-read-repository-contract-preconditions-v1`、repository_contract_preconditions_defined、`control-plane-read-disabled-database-guard-v1`、disabled_database_guard_defined、`control-plane-read-repository-contract-smoke-v1`、repository_contract_smoke_defined、`control-plane-read-repository-implementation-readiness-v1`、repository_implementation_readiness_defined、`control-plane-read-store-selection-readiness-v1`、store_selection_readiness_defined、`control-plane-read-schema-migration-readiness-v1`、schema_migration_readiness_defined、`control-plane-read-repository-contract-types-readiness-v1`、repository_contract_types_readiness_defined、`control-plane-read-repository-contract-types-implementation-v1`、repository_contract_types_implemented。
- Control Plane Read implementation readiness：`control-plane-read-repository-contract-smoke-runner-readiness-v1`、repository_contract_smoke_runner_readiness_defined、`control-plane-read-repository-contract-smoke-runner-implementation-v1`、repository_contract_smoke_runner_implemented、`control-plane-read-repository-interface-readiness-v1`、repository_interface_readiness_defined、`control-plane-read-repository-adapter-implementation-readiness-refresh-v1`、repository_adapter_implementation_readiness_refreshed、`control-plane-read-store-selector-enablement-preconditions-v1`、store_selector_enablement_preconditions_defined、`control-plane-read-schema-migration-implementation-preconditions-v1`、schema_migration_implementation_preconditions_defined、`control-plane-read-repository-adapter-implementation-plan-v1`、repository_adapter_implementation_plan_defined、`control-plane-read-schema-artifact-manifest-readiness-v1`、schema_artifact_manifest_readiness_defined、`control-plane-read-store-selector-smoke-readiness-v1`、store_selector_smoke_readiness_defined、`control-plane-read-production-auth-readiness-v1`、production_auth_readiness_defined、`control-plane-read-adapter-smoke-readiness-v1`、adapter_smoke_readiness_defined、`control-plane-read-implementation-trigger-review-v1`、implementation_trigger_review_defined、`control-plane-read-schema-artifact-evidence-v1`、schema_artifact_evidence_defined、`control-plane-read-implementation-entry-review-v1`、implementation_entry_review_defined、`product-surface-readiness-implementation-trigger-recheck-v1`、product_surface_readiness_trigger_recheck_defined、`product-surface-usage-gap-triage-v1`、product_surface_usage_gap_triage_defined、`control-plane-durable-read-foundation-v1`、durable_read_foundation_implemented。
- Workflow / Agent Runtime evidence：`workflow-function-surface-boundary-v1`、function_surface_boundary_defined、`workflow-definition-detail-read-v1`、workflow_definition_detail_read_defined、`workflow-run-detail-read-v1`、workflow_run_detail_read_defined、`workflow-blocked-action-preview-v1`、workflow_blocked_action_preview_defined、`workflow-application-detail-read-v1`、workflow_application_detail_read_defined、`workflow-confirmation-placeholder-read-v1`、workflow_confirmation_placeholder_read_defined、`workflow-draft-designer-offline-v1`、RadishFlow、不阻塞、`workflow-draft-validation-inspector-offline-v1`、workflow_draft_validation_inspector_offline_defined、`workflow-execution-plan-preview-offline-v1`、workflow_execution_plan_preview_offline_defined、`workflow-runtime-readiness-inspector-offline-v1`、workflow_runtime_readiness_inspector_offline_defined、`workflow-function-surface-readiness-close-v1`、workflow_function_surface_readiness_closed、`workflow-workspace-context-consistency-v1`、workflowWorkspaceContext、Workflow Review Handoff，以及 `workflow-saved-draft-v1-implementation`、saved_workflow_draft_domain_service_implemented、Go domain service / memory dev store / save-read-validate tests、`workflow-saved-draft-consumer-smoke-v1`、route contract / consumer smoke / version conflict guard、`workflow-draft-editing-entry-v1`、workflow_draft_editing_entry_implemented、`user-workspace-draft-creation-v1`、user_workspace_draft_creation_implemented、`user-workspace-saved-draft-list-v1`、user_workspace_saved_draft_list_implemented、`workflow-draft-designer-editing-model-v2`、workflow_draft_designer_editing_model_v2_implemented、`workflow-draft-node-attribute-editing-model-v1`、workflow_draft_node_attribute_editing_model_v1_implemented、`workflow-review-handoff-active-draft-v1`、workflow_review_handoff_active_draft_v1_implemented、`workflow-saved-draft-durable-store-preconditions-v1`、draft_durable_store_preconditions_defined、`workflow-saved-draft-repository-contract-preconditions-v1`、draft_repository_contract_preconditions_defined、`workflow-saved-draft-schema-migration-preconditions-v1`、draft_schema_migration_preconditions_defined、`workflow-saved-draft-auth-context-preconditions-v1`、draft_auth_context_preconditions_defined、`workflow-saved-draft-store-selector-enablement-preconditions-v1`、draft_store_selector_enablement_preconditions_defined、`workflow-saved-draft-schema-artifact-evidence-v1`、draft_schema_artifact_evidence_defined、`workflow-saved-draft-store-selector-smoke-readiness-v1`、draft_store_selector_smoke_readiness_defined、`workflow-saved-draft-repository-contract-smoke-v1`、draft_repository_contract_smoke_defined、`workflow-saved-draft-repository-contract-smoke-runner-readiness-v1`、draft_repository_contract_smoke_runner_readiness_defined、`workflow-saved-draft-repository-contract-smoke-runner-implementation-v1`、draft_repository_contract_smoke_runner_implemented、`workflow-saved-draft-repository-adapter-implementation-plan-v1`、draft_repository_adapter_implementation_plan_defined、`workflow-saved-draft-schema-artifact-manifest-v1`、draft_schema_artifact_manifest_defined、`workflow-saved-draft-adapter-smoke-readiness-v1`、draft_adapter_smoke_readiness_defined、`workflow-saved-draft-store-selector-implementation-entry-review-v1`、draft_store_selector_implementation_entry_review_defined、`workflow-saved-draft-schema-artifact-materialization-review-v1`、draft_schema_artifact_materialization_review_defined、`workflow-saved-draft-store-selector-smoke-v1`、draft_store_selector_smoke_implemented、`workflow-saved-draft-schema-artifact-materialization-v1`、draft_schema_artifact_materialized_static、`workflow-saved-draft-production-auth-readiness-v1`、draft_production_auth_readiness_defined、`workflow-saved-draft-repository-adapter-implementation-entry-review-v1`、draft_repository_adapter_implementation_entry_review_defined。
- Workflow / Agent Runtime latest adapter/auth/enablement evidence：`workflow-saved-draft-repository-adapter-implementation-v1`、draft_repository_adapter_implemented、`check-workflow-saved-draft-repository-adapter-implementation-v1.py`；`workflow-saved-draft-adapter-smoke-v1`、draft_adapter_smoke_executed、`check-workflow-saved-draft-adapter-smoke-v1.py`；`workflow-saved-draft-production-auth-runtime-v1`、draft_production_auth_runtime_bridge_implemented、`check-workflow-saved-draft-production-auth-runtime-v1.py`；`workflow-saved-draft-repository-mode-enablement-v1`、draft_repository_mode_enablement_review_defined、`check-workflow-saved-draft-repository-mode-enablement-v1.py`。
- Image Path evidence：`image-adapter-handshake-safety-gate-v1`、image_adapter_handshake_safety_gate_defined、`image-artifact-return-runbook-evidence-v1`、image_artifact_return_runbook_evidence_defined、`image-safety-runbook-evidence-v1`、image_safety_runbook_evidence_defined、`image-backend-adapter-readiness-evidence-v1`、image_backend_adapter_readiness_defined、`image-artifact-runtime-mapping-readiness-v1`、image_artifact_runtime_mapping_readiness_defined、`image-artifact-runtime-mapping-implementation-entry-review-v1`、image_artifact_runtime_mapping_entry_review_defined、`image-artifact-store-binary-reader-boundary-readiness-v1`、image_artifact_store_binary_reader_boundary_readiness_defined、`image-artifact-runtime-mapper-implementation-plan-v1`、image_artifact_runtime_mapper_implementation_plan_defined、`image-artifact-runtime-mapper-implementation-entry-v1`、image_artifact_runtime_mapper_implementation_entry_review_defined、`image-artifact-runtime-mapper-implementation-v1`、image_artifact_runtime_mapper_implementation_task_card_defined、`image-artifact-runtime-mapper-runtime-implementation-v1`、image_artifact_runtime_mapper_runtime_implemented、`image-artifact-runtime-mapper-response-consumer-integration-review-v1`、image_artifact_runtime_mapper_response_consumer_integration_review_defined、`image-artifact-response-consumer-implementation-readiness-v1`、image_artifact_response_consumer_implementation_readiness_defined、`image-artifact-response-consumer-implementation-v1`、image_artifact_response_consumer_implementation_task_card_defined、`image-artifact-response-consumer-runtime-implementation-v1`、image_artifact_response_consumer_runtime_implemented、`image-artifact-response-builder-integration-entry-review-v1`、image_artifact_response_builder_integration_entry_review_defined、`image-artifact-response-builder-integration-v1`、image_artifact_response_builder_integration_task_card_defined、`image-artifact-response-builder-runtime-integration-entry-review-v1`、image_artifact_response_builder_runtime_integration_entry_review_defined、`image-artifact-response-builder-runtime-integration-implementation-v1`、image_artifact_response_builder_runtime_integration_implemented。
- `P2 Session & Tooling Foundation` 保持 close candidate / governance-only，且不声明 `P2 short close`；证据锚点包括 `session-tooling-readiness-summary.json`、`session-tooling-foundation-status-summary.json`、`session-tooling-implementation-preconditions.json`、`session-tooling-negative-regression-skeleton.json`、`session-tooling-negative-regression-suite.json`、`session-tooling-negative-regression-suite-readiness.json`、`session-tooling-deny-by-default-implementation-gates.json`、`session-tooling-negative-coverage-rollup.json`、`session-tooling-route-negative-coverage-matrix.json`、`session-tooling-route-smoke-readiness-rollup.json`、`session-tooling-close-candidate-readiness-rollup.json`、`session-tooling-short-close-readiness-delta.json`、`session-tooling-readiness-consistency-rollup.json`、`session-tooling-executor-storage-confirmation-enablement-plan.json`、`session-tooling-stop-line-manifest.json`、`session-tooling-upper-layer-confirmation-flow-readiness.json`、`session-tooling-short-close-entry-checklist.json`、`session-tooling-confirmation-flow-design.json`、`session-tooling-independent-audit-records-design.json`、`session-tooling-result-materialization-policy-design.json`、`session-tooling-executor-boundary-design.json`、`session-tooling-storage-backend-design.json`、`session-recovery-checkpoint-route-smoke-coverage-summary.json`、`tool-registry-basic.json` 和 `tool-audit-record-basic.json`；这些只固定 `negative_regression_suite`、route smoke、metadata-only、blocked shell、enablement plan 和 stop-line manifest，不实现真实 executor、durable store、confirmation 接线、materialized result reader、长期记忆或 replay。
- `P3 Local Product Shell / Ops Surface` 已达到 local usable / read-only close；`UI Design Topic / Pencil Draft`、`p3-local-product-shell-short-close-checklist.json` 和 `check-p3-local-product-shell-short-close-checklist.py` 只固定本地只读产品壳和设计证据，不声明生产管理端 ready，不再默认继续补同类只读 console 小切片；P3 checklist 固定 production secret backend、process supervisor、部署环境隔离和 console production packaging 仍为 `not_satisfied`。
- `Production Ops Hardening v1` 静态边界证据包括 `config-secret-boundary`、`production-ops-config-secret-boundary.json`、`production-secret-backend-contract`、`production-ops-secret-backend-contract.json`、`production-secret-backend-implementation-readiness`、`production-ops-secret-backend-implementation-readiness.json`、`production-secret-backend-implementation-v1-plan.md`、`secret-ref-schema-and-fixtures`、`production-secret-reference.schema.json`、`production-secret-reference-basic.json`；这些证据不实现真实云 secret 服务、不写入真实 secret、不声明 production ready。
- Production Ops deployment evidence：`startup-supervisor-boundary`、`production-ops-startup-supervisor-boundary.json`、`environment-isolation`、`production-ops-environment-isolation-boundary.json`、`console-production-package-smoke`、`production-ops-console-package-smoke.json`、`docker-deployment-mode-definition`、`production-ops-docker-deployment-mode.json`、`docker-local-compose`、`production-ops-docker-local-compose.json`、`docker-test-prod-compose`、`production-ops-docker-test-prod-compose.json`、`deploy/.env.example`、`deploy/docker-compose.yaml`、`docker-image-build-publish`、`production-ops-docker-image-build-publish.json`、`v*-dev`、`v*-test`、`v*-release`、`deployment-readiness-smoke`、`docker compose config`、`production-ops-deployment-readiness-smoke.json`、`container-smoke-runbook`、`production-ops-container-smoke-runbook.json`、`container-smoke-record-template`、`production-ops-container-smoke-record-template.json` 和 `container_smoke_ready`。

## 当前优先做什么

第一顺位不是继续扩同层 gate，而是把已经完成的产品面证据收束到合适粒度的专题文档，并按专题选择下一批真实开发目标。

当前功能入口：

1. [User Workspace](features/user-workspace.md)
2. [Admin Control Plane](features/admin-control-plane.md)
3. [Model Gateway / API Distribution](features/model-gateway-api-distribution.md)
4. [Workflow / Agent Runtime](features/workflow-agent-runtime.md)
5. [Image Generation / Artifact Return](features/image-generation-artifact-return.md)

当前细专题入口：

1. [Workflow 细专题入口](features/workflow/README.md)
2. [Saved Workflow Draft v1](features/workflow/saved-workflow-draft-v1.md)
3. [Workflow Draft Designer Surface](features/workflow/draft-designer-surface.md)
4. [Workflow Draft Editing Entry v1](features/workflow/draft-editing-entry-v1.md)
5. [User Workspace Draft Creation v1](features/workflow/user-workspace-draft-creation-v1.md)
6. [User Workspace Saved Draft List v1](features/workflow/user-workspace-saved-draft-list-v1.md)
7. [Workflow Draft Designer Editing Model v2](features/workflow/draft-designer-editing-model-v2.md)
8. [Workflow Draft Node Attribute Editing Model v1](features/workflow/draft-node-attribute-editing-model-v1.md)
9. [Workflow Review Handoff Active Draft v1](features/workflow/review-handoff-active-draft-v1.md)
10. [Saved Workflow Draft Durable Store Preconditions v1](features/workflow/saved-workflow-draft-durable-store-preconditions-v1.md)
11. [Saved Workflow Draft Repository Contract Preconditions v1](features/workflow/saved-workflow-draft-repository-contract-preconditions-v1.md)
12. [Saved Workflow Draft Schema / Migration Preconditions v1](features/workflow/saved-workflow-draft-schema-migration-preconditions-v1.md)
13. [Saved Workflow Draft Auth Context Preconditions v1](features/workflow/saved-workflow-draft-auth-context-preconditions-v1.md)
14. [Saved Workflow Draft Store Selector Enablement Preconditions v1](features/workflow/saved-workflow-draft-store-selector-enablement-preconditions-v1.md)
15. [Saved Workflow Draft Schema Artifact Evidence v1](features/workflow/saved-workflow-draft-schema-artifact-evidence-v1.md)
16. [Saved Workflow Draft Store Selector Smoke Readiness v1](features/workflow/saved-workflow-draft-store-selector-smoke-readiness-v1.md)
17. [Saved Workflow Draft Repository Contract Smoke v1](features/workflow/saved-workflow-draft-repository-contract-smoke-v1.md)
18. [Saved Workflow Draft Repository Contract Smoke Runner Readiness v1](features/workflow/saved-workflow-draft-repository-contract-smoke-runner-readiness-v1.md)
19. [Saved Workflow Draft Repository Contract Smoke Runner Implementation v1](features/workflow/saved-workflow-draft-repository-contract-smoke-runner-implementation-v1.md)
20. [Saved Workflow Draft Repository Adapter Implementation Plan v1](features/workflow/saved-workflow-draft-repository-adapter-implementation-plan-v1.md)
21. [Saved Workflow Draft Schema Artifact Manifest v1](features/workflow/saved-workflow-draft-schema-artifact-manifest-v1.md)
22. [Saved Workflow Draft Adapter Smoke Readiness v1](features/workflow/saved-workflow-draft-adapter-smoke-readiness-v1.md)
23. [Saved Workflow Draft Store Selector Implementation Entry Review v1](features/workflow/saved-workflow-draft-store-selector-implementation-entry-review-v1.md)
24. [Saved Workflow Draft Schema Artifact Materialization Review v1](features/workflow/saved-workflow-draft-schema-artifact-materialization-review-v1.md)
25. [Saved Workflow Draft Store Selector Implementation v1](features/workflow/saved-workflow-draft-store-selector-implementation-v1.md)
26. [Saved Workflow Draft Schema Artifact Materialization v1](features/workflow/saved-workflow-draft-schema-artifact-materialization-v1.md)
27. [Saved Workflow Draft Production Auth Readiness v1](features/workflow/saved-workflow-draft-production-auth-readiness-v1.md)
28. [Saved Workflow Draft Repository Adapter Implementation Entry Review v1](features/workflow/saved-workflow-draft-repository-adapter-implementation-entry-review-v1.md)
29. [Saved Workflow Draft Adapter Smoke Execution v1](features/workflow/saved-workflow-draft-adapter-smoke-execution-v1.md)
30. [Saved Workflow Draft Production Auth Runtime v1](features/workflow/saved-workflow-draft-production-auth-runtime-v1.md)
31. [Saved Workflow Draft Repository Mode Enablement v1](features/workflow/saved-workflow-draft-repository-mode-enablement-v1.md)
32. [Dev-only Saved Draft Consumer](features/workflow/dev-only-saved-draft-consumer.md)
33. [平台专题入口](platform/README.md)
34. [扩展 / 集成专题入口](integrations/README.md)

推荐下一批开发目标从以下方向选择一个：

1. `Workflow / Agent Runtime`：`Saved Workflow Draft v1` consumer integration、受控编辑、用户工作区创建、saved draft list / restore、本地结构编辑、节点属性编辑、active draft review record、durable store 迁移证据链、static runner implementation、schema artifact、formal store selector、production auth readiness、repository adapter implementation、`Saved Workflow Draft Adapter Smoke Execution v1`、`Saved Workflow Draft Production Auth Runtime v1` 和 `Saved Workflow Draft Repository Mode Enablement v1` 已完成。下一步若继续 durable store，应在真实数据库连接、SQL migration runner、OIDC middleware / token validation、membership adapter 或 production API consumer 中选择单一前置方向独立评审；若继续用户审查体验，应选择新的明确用户审查增强点。不得绕过 dev auth / write enablement、scope check、no sample fallback、repository mode enablement 评审结论、真实数据库 / SQL migration runner 和 public production API 停止线。
2. `User Workspace`：若继续用户端路径，可推进恢复后的审查交接，但不能绕过 saved draft scope、owner / workspace 和 no sample fallback 约束。
3. `Model Gateway / API Distribution`：推进真实 API distribution 前的 key / quota / trace 设计，而不是继续新增 evidence 面板。
4. `Admin Control Plane`：为未来 Radish OIDC 或 read store adapter 选择单一实现方向，不能和管理端写入并行打开。
5. `Image Generation / Artifact Return`：若继续推进，只能在 store、reader、public URL 或 backend adapter 中选择一个方向独立设计。

## 当前不要做

- 不继续为普通只读展示页、evidence review、文案和布局逐项新增 task card / fixture / checker。
- 不把 task card 当成功能长期设计文档。
- 不在没有对应专题文档更新的情况下启动新的大功能或高风险实现。
- 不把 Image Path metadata-only 接线解释为 artifact store、public delivery 或真实 backend ready。
- 不把 durable read foundation 解释为 repository adapter、真实数据库、OIDC、production API consumer 或完整 read-side API ready。
- 不把 Workflow / Gateway / Admin 的普通离线 evidence surface 写成生产能力 ready。
- 不在上层项目没有真实挂载点时继续细化假想接线。
- 不默认启动 Docker、下载模型、长跑真实模型、生成图片或访问真实 backend。

## 最小读取路径

回答“今天做什么”时，默认读取：

1. `AGENTS.md` 或 `CLAUDE.md`
2. [文档入口](README.md)
3. 本文件
4. [功能设计文档入口](features/README.md)
5. 与当次专题直接相关的细专题，例如 [Workflow 细专题入口](features/workflow/README.md)
6. 必要时读取 [产品范围](radishmind-product-scope.md)、[路线图](radishmind-roadmap.md)、[能力矩阵](radishmind-capability-matrix.md)

实施具体功能时，先读产品面大方向和对应细专题，再读相关 contract、task card、checker 或周志。

## 验证基线

文档或治理改动完成后，macOS / Linux / WSL 环境优先执行：

```bash
./scripts/bootstrap-dev.sh
./scripts/check-repo.sh --fast
```

Windows / PowerShell 环境使用：

```powershell
pwsh ./scripts/bootstrap-dev.ps1
pwsh ./scripts/check-repo.ps1 -Fast
```

若改动触及阶段边界、协作规则、验证入口或文档真相源，应补跑全量仓库检查。

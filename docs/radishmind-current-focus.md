# RadishMind 当前推进焦点

更新时间：2026-07-19

## 文档目的

本文档用于回答“根据项目规划和开发进度，今天要做什么以推进开发”。它是新会话短入口，只保留当前阶段、最近结论、下一顺位和停止线；正文默认中文，代码标识符、路径、配置键和状态锚点保留原文。

功能细节默认先进入 [功能设计文档入口](features/README.md) 所定义的专题层级：产品面大方向进入 `docs/features/*.md`，具体功能和复杂页面进入对应子目录，平台横切能力进入 `docs/platform/`，外部接入进入 `docs/integrations/`。实现批次进入 `docs/task-cards/`，长验证记录进入周志、清单、摘要或运行记录。

## 当前结论（默认读取到本节结束）

- 当前成熟度：内部开发者预览，不使用 `M2` 编号，不声明生产就绪。
- 产品焦点：工作流审查链、Gateway 请求历史与调试台、应用 API 接入、配置草案、发布治理、应用目录与 API 密钥开发测试态认证均已完成；Workflow HTTP Tool、RAG Retrieval / 应用知识快照、RAG Regression Review、离线 RAG 质量评测、应用作用域 durable dataset / candidate review、知识基线晋级与配置绑定，以及 [Workflow RAG 应用运行时激活与受控调用（开发 / 测试态）v1](features/workflow/workflow-rag-application-runtime-activation-controlled-invocation-dev-test-v1.md)均已关闭。[应用运行观测与用量归因 v1](features/user-workspace/application-operations-observability-usage-attribution-v1.md)批次 A 也已完成：它在应用作用域并列消费 Gateway Request History 与 Workflow Run History，分别表达当前窗口状态、usage availability 和受控调用计数，不推测跨来源关联，也不估算 token、成本、配额或计费。真实 Radish 联调保持 `real_radish_integration_deferred`，外部 connector、在线搜索、embedding / reranker、自动 activation / baseline / release、生产认证、正式晋级、生产密钥、配额和计费继续关闭。
- `R2 正确性与安全清零`、`R3 工作流草案审查闭环`、`R4 Gateway 运行时产品化`、`R5 测试、CI 与性能预算`、`R6 文档与检查器收敛` 均已完成。R6 关闭评审确认活动 checker 从 `132` 项、`38,644` 行降至 `111` 项、`28,486` 行，分别下降约 `15.9%` 与 `26.3%`；Provider、Production Ops 和 Control Plane formal UI 因仍有独立证据责任继续活动，不再派生第六批或同层 readiness 链。
- `P3 Local Product Shell / Ops Surface` 保持 `local usable / read-only close`，不再默认继续补同类只读 console 小切片。production secret backend、process supervisor、部署环境隔离和 console production packaging 仍为 `not_satisfied`。
- 四个正式一级产品面保持为“用户工作区”“管理控制面”“模型网关 / API 分发”“工作流 / Agent 运行时”；图片路径是横切适配能力，不作为当前第五条一级主线。
- 旧生产凭据后端 / 存储适配器准入链已冻结为历史证据，`storage_adapter_runtime_implementation_entry_refresh_after_provider_account_resource_endpoint_review` 不再是当前开发下一步。

当前最多两条在制主线：

1. 产品线：Application RAG runtime 及应用运行观测批次 A 已关闭，当前没有在制实现批次。下一步先从[功能设计入口](features/README.md)选择并评审一个新的真实用户流程；应用观测只有在全分页稳定统计、可信 reported usage 或正式 quota / billing owner 成立时才评审服务端 summary，不继续派生同层 UI 汇总。connector、在线搜索、embedding / reranker、自动 activation / baseline / release、多实例和生产能力均不预设为下一项；`real_radish_integration_deferred` 不占用该主线。
2. 工程线：R2 至 R6 均已关闭，当前没有独立整改批次。后续只在真实功能实现中复用或替代对应门禁；没有等价行为证据的 Provider、Production Ops 与 formal UI 检查继续保留，不按数量清理，也不新建同层治理入口。

R3 与 [工作流草案 PostgreSQL 开发测试态存储库 v1](features/workflow/saved-workflow-draft-postgresql-dev-test-repository-v1.md) 已于 2026-07-11 完成。`postgres_dev_test` 已覆盖迁移 / 回滚 / 重新应用、运行角色 DDL 拒绝、服务重启恢复、原子预期版本校验、租户 / 工作区 / 应用 / 所有者作用域、不回退、CI 与真实浏览器双标签冲突审查。该完成不启用生产存储库模式，也不代表 OIDC、生产凭据、审计存储或公开生产 API 已就绪。

持久草案存储库、稳定 Gateway、执行器 v0 与持久开发测试态运行历史均已成立。真实浏览器已验证“创建 → 保存 → 启动受限运行 → 分页历史 → 详情 → 服务重启恢复”，运行记录中模型服务调用为 1，工具、确认、业务写入和重放均为 0，原始输入与条件值未持久化。无限制工具、业务写回、自动确认提交、重放和恢复继续关闭。

总入口与证据：

1. [工程健康与产品化整改专题 v1](platform/engineering-health-productization-remediation-v1.md)
2. [Gateway Python Bridge Runtime v1](features/gateway/python-bridge-runtime-v1.md)
3. [stdio worker pool 对照证据](features/gateway/evidence/stdio-worker-pool-comparison-2026-07-11.json)
4. [Saved Workflow Draft v1](features/workflow/saved-workflow-draft-v1.md)
5. [Workflow Executor v0](features/workflow/workflow-executor-v0.md)
6. [Workflow Run History / Durable Dev-Test Run Store v1](features/workflow/workflow-run-history-durable-dev-test-store-v1.md)
7. [Workflow Execution Diagnostics / Failure Review v1](features/workflow/workflow-execution-diagnostics-failure-review-v1.md)
8. [Workflow Run Comparison / Regression Review v1](features/workflow/workflow-run-comparison-regression-review-v1.md)
9. [Workflow Evaluation Cases / Batch Regression Review v1](features/workflow/workflow-evaluation-cases-batch-regression-review-v1.md)
10. [Workflow Evaluation Baseline & Case Versioning v1](features/workflow/workflow-evaluation-baseline-case-versioning-v1.md)
11. [Workflow Evaluation Suite / Release Review v1](features/workflow/workflow-evaluation-suite-release-review-v1.md)
12. [Model Gateway Request History / Usage & Failure Review v1](features/gateway/model-gateway-request-history-usage-failure-review-v1.md)
13. [用户工作区应用 API 接入与调用 v1](features/user-workspace/application-api-integration-invocation-v1.md)
14. [用户工作区应用配置草案与审查 v1](features/user-workspace/application-configuration-draft-review-v1.md)
15. [用户工作区应用发布治理与晋级审查 v1](features/user-workspace/application-publish-governance-promotion-v1.md)
16. [用户工作区应用目录与生命周期（开发/测试态）v1](features/user-workspace/application-catalog-lifecycle-dev-test-v1.md)
17. [Admin Control Plane Authenticated Read Store Transition v1](features/admin-control-plane/authenticated-read-store-transition-v1.md)
18. [Workflow 受控 HTTP Tool 与人工确认执行（开发 / 测试态）v1](features/workflow/controlled-http-tool-human-confirmation-dev-test-v1.md)
19. [Workflow 受控 HTTP Tool 与人工确认执行（开发 / 测试态）v1 实施任务卡](task-cards/workflow-controlled-http-tool-human-confirmation-dev-test-v1-plan.md)
20. [Workflow RAG Retrieval 与应用知识快照 v1 实施任务卡](task-cards/workflow-rag-retrieval-application-knowledge-snapshot-dev-test-v1-plan.md)
21. [Workflow RAG Regression Review 与 Evaluation Profile v1](features/workflow/workflow-rag-regression-review-evaluation-profile-dev-test-v1.md)
22. [Workflow RAG Regression Review 与 Evaluation Profile v1 实施任务卡](task-cards/workflow-rag-regression-review-evaluation-profile-dev-test-v1-plan.md)
23. [Workflow RAG 评测数据集与知识质量审查 v1](features/workflow/workflow-rag-evaluation-dataset-knowledge-quality-review-v1.md)
24. [Workflow RAG 评测数据集与知识质量审查 v1 实施任务卡](task-cards/workflow-rag-evaluation-dataset-knowledge-quality-review-v1-plan.md)
25. [Workflow RAG 评测数据集应用资源化与候选快照审查 v1](features/workflow/workflow-rag-evaluation-dataset-application-resource-candidate-snapshot-review-v1.md)
26. [Workflow RAG 评测数据集应用资源化与候选快照审查 v1 实施任务卡](task-cards/workflow-rag-evaluation-dataset-application-resource-candidate-snapshot-review-v1-plan.md)
27. [Workflow RAG 知识基线晋级与应用配置绑定审查 v1](features/workflow/workflow-rag-knowledge-baseline-promotion-application-binding-review-v1.md)
28. [Workflow RAG 知识基线晋级与应用配置绑定审查 v1 实施任务卡](task-cards/workflow-rag-knowledge-baseline-promotion-application-binding-review-v1-plan.md)
29. [Workflow RAG 应用运行时激活与受控调用（开发 / 测试态）v1](features/workflow/workflow-rag-application-runtime-activation-controlled-invocation-dev-test-v1.md)
30. [Workflow RAG 应用运行时激活与受控调用（开发 / 测试态）v1 实施任务卡](task-cards/workflow-rag-application-runtime-activation-controlled-invocation-dev-test-v1-plan.md)
31. [应用运行观测与用量归因 v1](features/user-workspace/application-operations-observability-usage-attribution-v1.md)
32. [本周周志](devlogs/2026-W29.md)

## 当前不要做

- 不继续为普通只读展示页、evidence review、文案和布局逐项新增 task card / fixture / checker。
- 不把 task card 当成功能长期设计文档。
- 不在没有对应专题文档更新的情况下启动新的大功能或高风险实现。
- 不把 Image Path 仅元数据接线解释为 artifact store、public delivery 或真实后端已就绪。
- 不把 durable read foundation 解释为 repository adapter、真实数据库、OIDC、production API consumer 或完整 read-side API ready。
- 不把 Workflow / Gateway / Admin 的普通离线证据界面写成生产能力已就绪。
- 不在上层项目没有真实挂载点时继续细化假想接线。
- 不默认启动 Docker、下载模型、长跑真实模型、生成图片或访问真实后端。

## 默认读取路径

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

# RadishMind 当前推进焦点

更新时间：2026-08-10

## 文档目的

本文档用于回答“根据项目规划和开发进度，今天要做什么以推进开发”。它是新会话短入口，只保留当前阶段、最近结论、下一顺位和停止线；正文默认中文，代码标识符、路径、配置键和状态锚点保留原文。

功能细节默认先进入 [功能设计文档入口](features/README.md) 所定义的专题层级：产品面大方向进入 `docs/features/*.md`，具体功能和复杂页面进入对应子目录，平台横切能力进入 `docs/platform/`，外部接入进入 `docs/integrations/`。实现批次进入 `docs/task-cards/`，长验证记录进入周志、清单、摘要或运行记录。

## 当前结论（默认读取到本节结束）

- 当前成熟度：内部开发者预览，不使用 `M2` 编号，不声明生产就绪。
- 当前产品专题：[Workflow Definition 结构化运行输入（开发 / 测试态）v1](features/workflow/workflow-definition-structured-runtime-inputs-dev-test-v1.md) 正在实施批次 C，状态为 `workflow_definition_structured_runtime_inputs_dev_test_v1_batch_c_pencil_visual_r4_review_next`。批次 B 的 strict HTTP union、executor v2、metadata-only Run v8 / History 与 memory、SQLite、PostgreSQL durable chain 已关闭；首版 `B / 局部 Pencil · R1` 因沿用 S9 之后偏离的表单看板语言被人工退回，Visual R2 因硬方形业务表面再次退回，Visual R3 虽恢复职责圆角，字段区仍像整页横向切割的静态列表。Desktop `W3O4tV`、Narrow `t39foq` 已修订为 Visual R4 表单画布，当前先完成人工视觉复核，再实现 strict contract decoder、Session v4 与 Definition Run / Session 共享 `StructuredRuntimeInputEditor`。
- [应用评测计划、受控执行与证据归档（开发 / 测试态）v1](features/user-workspace/application-evaluation-campaign-controlled-execution-dev-test-v1.md) 功能状态仍为 `application_evaluation_campaign_controlled_execution_dev_test_v1_completed`。后端 A 至 D、React strict consumer、定向 `7/7`、Web `316/316`、production build、memory 与 SQLite 浏览器均已完成；SQLite exact Plan `aeplan_lkqe7gr7kjobmf73 v1` 产生两次 succeeded Campaign，Pair Preview 后交接 Case `eval_034d69aec0d7a2323c7f222f v1` 与 Suite `suite_9a8017d686be57009c7ad973`，服务重启后恢复同一证据。S10 Pencil R1 因页面骨架偏离被退回，Visual R2 又因硬方形表面与 S1–S8 形态语言不一致被退回；Desktop `Um8Zh`、Narrow `ZxJd7`、Decision R15 `UNMOS` 已修订为 Visual R3，React 尚未声明与该视觉稿对齐。
- [应用 API Key 请求配额与 Provider Attempt 准入（开发 / 测试态）v1](features/gateway/application-api-key-request-quota-admission-dev-test-v1.md)批次 A 至 E 已全部完成，功能状态仍为 `application_api_key_request_quota_admission_dev_test_v1_completed`。后端三模式 quota owner、Admin GET / PUT、独立权限和六条 provider 前原子准入继续成立；正整数 CAS 确认、missing / permission / environment / conflict / store failure 失败关闭和真实浏览器连续链已经完成。S9 Pencil R1 因没有继承 S8 连续 Workbench 被退回，Visual R2 又因上下文、任务、owner 与边界全部使用硬方形表面被退回；Desktop `C7pkb`、Narrow `x8lESc`、Decision R14 `tCWCW` 已修订为 Visual R3，React 尚未声明与该视觉稿对齐。旧 User Workspace `QuotaSummary` 仍为 `quota_policy_unavailable`，生产 quota、rate limit、token / cost、billing、正式 membership / OIDC 与自动路由未打开。
- [RadishMind Family UI 产品化设计与迁移 v1](features/user-workspace/radishmind-family-ui-productization-v1.md) 已完成 `v26.7.3` 通用参考基线与 `S1 R8` 至 `S8 R1` 的已审视觉语言。Pencil 顶层 26 个基准面按 S1 → S10 横向排列；S9、S10 保持 Visual R3 并等待人工视觉复核。结构化输入局部稿在同一 Workbench 骨架上继续修订为 Visual R4：字段区使用非对称表单画布、真实输入形态、精确错误归属和布尔开关，避免整页横线切割或等权卡片墙。不把 S9 / S10 Visual R3 或结构化输入 Visual R4 写成 React 已落地；不建立 S11。
- S9 后续真实 API Key 高频链已复现 quota 内调用 → `429 / gateway_quota_exceeded / quota_admission` → 同 request id Request History → Admin Quota owner。审计发现 Playground 原本只显示技术码，现按 `C / 直接实现` 复用既有失败引导模式，明确 UTC 日预算、零 provider 调用和无自动重试，并精确打开同一 application 的 Admin Quota；用户工作区不读取或推算 used / remaining。该功能修正没有要求新的 Pencil 决策；随后 S9 的 Visual R2 因形态语言偏离被退回并修订为 Visual R3，API、schema、repository、permission 与生产停止线不变。
- SQLite 本地产品继续复验 Saved Draft → Definition candidate / review / activation → v5 run → Comparison → Evaluation Case / Suite / Human Decision。审计实际发现 RAG 草案可进入不兼容 candidate、append-only audit 后续创建误取首条 audit、activation 只信任旧 eligibility 标记，以及 definition 派生草案按存储顺序而非图拓扑排列四处阻塞；现已在现有 owner 内失败关闭并修正。RAG 草案只交接独立 Workflow RAG Promotion owner，通用 Definition candidate 只接收 `prompt | llm | condition | output` 节点，candidate 和 activation 都复核 executor graph，派生草案按拓扑排序；两条真实 v5 run、Comparison、Case、Suite 与 append-only approved decision 已贯通。五维评分 `0 / 0 / 0 / 1 / 1 = 2`，采用 `C / 直接实现`，未操作正被其它项目占用的 Pencil，也未新增 API、schema、task card、fixture 或 checker。
- Workflow RAG Promotion → Application Configuration Draft 的真实交接已复验并收口。原入口只切换 hash，未打开配置 owner，也未带入当前 approved binding；现复用 Application Development Workspace 的单一易失 handoff，只传精确 `candidateId`，由配置 owner 重新读取并仅选择同 application 下仍为 `approved + eligible` 的 binding，任何缺失、blocked、撤销、scope 或 store 失败均不回退其它记录。浏览器进一步确认本地来源草案为 `v1`、当前草案已为 `v2`，显式恢复后以 `workflow_rag_promotion_draft_changed` 失败关闭，未自动恢复、挂载或保存。五维评分 `0 / 0 / 0 / 1 / 1 = 2`，采用 `C / 直接实现`；Web `308/308`、production build 与 `1440×900`、`900×900`、`720×900`、`390×844` 验收通过，无横向溢出或控制台 warning / error。Pencil 仍被其它项目占用，本批未读取或修改设计源，也未新增 API、schema、migration、repository、permission、task card、fixture 或 checker。
- Workflow RAG 应用运行时的当前 SQLite 权威链已从过期 promotion 恢复到 configuration draft `v3`、binding `wragb_xektdumcd2i2ow7h` 与 active assignment `wragra_ftrte2fisc7t7os5 v3`；两次相同输入的 v4 run 已通过 Comparison v3、versioned Evaluation Case、digest-bound Suite 和 append-only `approved v1`，正式 production blockers 仍使 publish candidate 保持 `promotion_blocked`。真实链发现 Comparison v2 / v3 顶层 `run_profile` 错保留为 standard profile，以及 application 切换时排队 evidence updater 可让 React 根节点失败；现已在既有 contract 与 workspace owner 内分别对齐严格 profile 校验和 state-apply scope guard。独立 Gateway Key 另行验证 `/v1/models` 与 `/v1/responses` sanitized Request History；三把临时验收 Key 已撤销。Web `309/309`、Platform `internal/httpapi`、production build 与 `1440×900`、`900×900`、`390×844` 浏览器复核通过。Pencil 仍被其它项目占用且未操作，没有新增 API、schema、migration、repository、permission、task card、fixture 或 checker。
- [用户工作区设计与开发文档](features/user-workspace.md)中的首批真实路径 UI 一致性治理已完成，状态为 `user_workspace_real_path_ui_coherence_v1_completed`。SQLite 本地产品真实链确认 Saved Draft、应用重新启用与 API Key 轮换的领域行为正确；Web 已补齐精确打开后的 Designer 交接、长标题信息密度、解除归档重新打开说明、开发测试态环境标签、`api_key_application_unavailable` 稳定脱敏解释和替代 Key 精确验证后的列表即时刷新。没有新增 API、schema、repository、任务卡或 checker。
- [API 密钥引导式轮换与验证后退役（开发 / 测试态）v1](features/user-workspace/api-key-guided-rotation-verified-retirement-dev-test-v1.md) 批次 A、B 已完成，状态为 `api_key_guided_rotation_verified_retirement_dev_test_v1_completed`。易失脱敏会话、同应用 / 同 owner / 同 scopes 替代、`last_used_at` 验证门槛、原 Key 精确重读与 revoke CAS、Web 和真实浏览器连续链已有可执行证据；未新增 rotate API、schema 或持久 rotation owner。
- [应用解除归档与安全重新启用（开发 / 测试态）v1](features/user-workspace/application-unarchive-safe-reactivation-dev-test-v1.md) 批次 A 至 C 已完成，状态为 `application_unarchive_safe_reactivation_dev_test_v1_completed`。三种 store CAS、`applications:archive + applications:write` 单次组合权限、显式影响确认、Gateway 资格回归、Web 与真实浏览器连续链已有可执行证据；目录 owner 不级联改写 API Key、运行时绑定、会话、草案、候选或运行记录。
- [已保存 Workflow 草案库生命周期与组织（开发 / 测试态）v1](features/workflow/saved-workflow-draft-library-lifecycle-organization-dev-test-v1.md) 批次 A 至 E 已完成，状态为 `saved_workflow_draft_library_lifecycle_organization_dev_test_v1_completed`。领域、三种 store owner、严格 cursor、超过 `200` 条分页 / 组合筛选、双版本并发、双数据库 `0003`、原子 transition / event、HTTP lifecycle API、独立 archive permission、相邻操作 active lifecycle 资格、Web 活动 / 归档库，以及 SQLite 重启和真实浏览器 archive → 只读审查 → unarchive 连续链已有可执行证据；[唯一高风险任务卡](task-cards/saved-workflow-draft-library-lifecycle-organization-dev-test-v1-plan.md)已关闭。
- 产品焦点：[Provider 上报用量规范化与应用用量审查（开发 / 测试态）v1](features/gateway/provider-reported-usage-normalization-application-review-dev-test-v1.md) 已完成，状态为 `provider_reported_usage_normalization_application_review_dev_test_v1_completed`。OpenAI-compatible、Gemini、Anthropic、HuggingFace 与 Ollama 的可信 reported usage 已通过 Gateway envelope、三类 northbound unary / stream、Request History memory / SQLite / PostgreSQL 和 Application Operations 当前窗口审查形成连续证据；缺失或非法 usage 保持 `not_reported`。
- [Workspace-scoped Mutation Authorization / 工作区写入与审查动作成员资格绑定（开发 / 测试态）v1](features/user-workspace/workspace-scoped-mutation-authorization-dev-test-v1.md) 已完成批次 A 至 E，状态为 `workspace_scoped_mutation_authorization_dev_test_v1_complete`。47 条人类交互式 mutation 已复用唯一 membership provider，专题关闭。
- `R2 正确性与安全清零`、`R3 工作流草案审查闭环`、`R4 Gateway 运行时产品化`、`R5 测试、CI 与性能预算`、`R6 文档与检查器收敛` 均已完成。R6 关闭评审确认活动 checker 从 `132` 项、`38,644` 行降至 `111` 项、`28,486` 行，分别下降约 `15.9%` 与 `26.3%`；Provider、Production Ops 和 Control Plane formal UI 因仍有独立证据责任继续活动，不再派生第六批或同层 readiness 链。
- `P3 Local Product Shell / Ops Surface` 保持 `local usable / read-only close`，不再默认继续补同类只读 console 小切片。production secret backend、process supervisor、部署环境隔离和 console production packaging 仍为 `not_satisfied`。
- 四个正式一级产品面保持为“用户工作区”“管理控制面”“模型网关 / API 分发”“工作流 / Agent 运行时”；图片路径是横切适配能力，不作为当前第五条一级主线。
- 旧生产凭据后端 / 存储适配器准入链已冻结为历史证据，`storage_adapter_runtime_implementation_entry_refresh_after_provider_account_resource_endpoint_review` 不再是当前开发下一步。

当前最多两条在制主线：

1. 产品线：继续实施 Workflow Definition 结构化运行输入批次 C。批次 B 的 HTTP / executor、三模式 durable chain、Run History metadata-only 与双数据库复验已关闭；Desktop `W3O4tV` / Narrow `t39foq` 已在 Visual R3 列表形态退回后修订为 Visual R4 表单画布。下一步先完成人工视觉复核，再实现 strict contract decoder、Session v4 与 Definition Run / Session 两个共享编辑器 consumer。Evaluation Plan / Campaign v2 留在批次 D，不创建 S11 页面族。
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
32. [Workflow 不可变版本晋级与受控运行绑定（开发 / 测试态）v1](features/workflow/workflow-definition-version-promotion-controlled-runtime-binding-dev-test-v1.md)
33. [Workflow 不可变版本晋级与受控运行绑定实施任务卡](task-cards/workflow-definition-version-promotion-controlled-runtime-binding-dev-test-v1-plan.md)
34. [应用交互会话与受控运行编排（开发 / 测试态）v1](features/user-workspace/application-interaction-session-controlled-runtime-orchestration-dev-test-v1.md)
35. [应用交互会话与受控运行编排实施任务卡](task-cards/application-interaction-session-controlled-runtime-orchestration-dev-test-v1-plan.md)
36. [应用开发工作区与发布准备审查 v1](features/user-workspace/application-development-workspace-release-readiness-review-v1.md)
37. [提示词应用模板版本审查与受控调用（开发 / 测试态）v1](features/user-workspace/prompt-application-template-version-review-controlled-invocation-dev-test-v1.md)
38. [提示词应用模板版本审查与受控调用实施任务卡](task-cards/prompt-application-template-version-review-controlled-invocation-dev-test-v1-plan.md)
39. [Agent / Copilot 应用档案版本审查与受控建议（开发 / 测试态）v1](features/user-workspace/agent-copilot-application-profile-version-review-controlled-suggestion-dev-test-v1.md)
40. [Agent / Copilot 应用档案版本审查与受控建议实施任务卡](task-cards/agent-copilot-application-profile-version-review-controlled-suggestion-dev-test-v1-plan.md)
41. [Prompt / Agent 应用回归评测与发布审查（开发 / 测试态）v1](features/user-workspace/prompt-agent-application-regression-evaluation-release-review-dev-test-v1.md)
42. [Prompt / Agent 应用回归评测与发布审查实施任务卡](task-cards/prompt-agent-application-regression-evaluation-release-review-dev-test-v1-plan.md)
43. [图片生成 / 产物返回](features/image-generation-artifact-return.md)
44. [Image Adapter 受控调用与 artifact 返回实施任务卡](task-cards/image-adapter-controlled-invocation-artifact-return-dev-test-v1-plan.md)
45. [Provider Profile / Model Route 配置草案、版本审查与受控启用（开发 / 测试态）v1](features/admin-control-plane/provider-profile-model-route-controlled-activation-dev-test-v1.md)
46. [Admin Provider Profile / Model Route 受控启用实施任务卡](task-cards/admin-provider-route-controlled-activation-dev-test-v1-plan.md)
47. [本周周志](devlogs/2026-W32.md)
48. [Workspace-scoped Read Transition / 工作区选择与成员资格绑定（开发 / 测试态）v1](features/user-workspace/workspace-scoped-read-transition-dev-test-v1.md)
49. [工作区运营收件箱（开发 / 测试态）v1](features/user-workspace/workspace-operations-inbox-dev-test-v1.md)
50. [Workspace-scoped Mutation Authorization / 工作区写入与审查动作成员资格绑定（开发 / 测试态）v1](features/user-workspace/workspace-scoped-mutation-authorization-dev-test-v1.md)
51. [Workspace-scoped Mutation Authorization 实施任务卡](task-cards/workspace-scoped-mutation-authorization-dev-test-v1-plan.md)
52. [已保存 Workflow 草案派生（开发 / 测试态）v1](features/workflow/saved-workflow-draft-derivation-dev-test-v1.md)
53. [已保存 Workflow 草案派生实施任务卡](task-cards/saved-workflow-draft-derivation-dev-test-v1-plan.md)
54. [已保存 Workflow 草案修订历史、版本比较与显式恢复（开发 / 测试态）v1](features/workflow/saved-workflow-draft-revision-history-restore-dev-test-v1.md)
55. [已保存 Workflow 草案修订历史、版本比较与显式恢复实施任务卡](task-cards/saved-workflow-draft-revision-history-restore-dev-test-v1-plan.md)
56. [Provider 上报用量规范化与应用用量审查（开发 / 测试态）v1](features/gateway/provider-reported-usage-normalization-application-review-dev-test-v1.md)
57. [Provider 上报用量规范化与应用用量审查实施任务卡](task-cards/provider-reported-usage-normalization-application-review-dev-test-v1-plan.md)
58. [已保存 Workflow 草案库生命周期与组织（开发 / 测试态）v1](features/workflow/saved-workflow-draft-library-lifecycle-organization-dev-test-v1.md)
59. [已保存 Workflow 草案库生命周期与组织实施任务卡](task-cards/saved-workflow-draft-library-lifecycle-organization-dev-test-v1-plan.md)
60. [应用解除归档与安全重新启用（开发 / 测试态）v1](features/user-workspace/application-unarchive-safe-reactivation-dev-test-v1.md)
61. [应用解除归档与安全重新启用实施任务卡](task-cards/application-unarchive-safe-reactivation-dev-test-v1-plan.md)
62. [API 密钥引导式轮换与验证后退役（开发 / 测试态）v1](features/user-workspace/api-key-guided-rotation-verified-retirement-dev-test-v1.md)
63. [API 密钥引导式轮换与验证后退役实施任务卡](task-cards/api-key-guided-rotation-verified-retirement-dev-test-v1-plan.md)
64. [应用 API Key 请求配额与 Provider Attempt 准入（开发 / 测试态）v1](features/gateway/application-api-key-request-quota-admission-dev-test-v1.md)
65. [应用 API Key 请求配额与 Provider Attempt 准入实施任务卡](task-cards/application-api-key-request-quota-admission-dev-test-v1-plan.md)
66. [应用评测计划、受控执行与证据归档（开发 / 测试态）v1](features/user-workspace/application-evaluation-campaign-controlled-execution-dev-test-v1.md)
67. [应用评测计划、受控执行与证据归档实施任务卡](task-cards/application-evaluation-campaign-controlled-execution-dev-test-v1-plan.md)

## 当前不要做

- 不继续为普通只读展示页、evidence review、文案和布局逐项新增 task card / fixture / checker。
- 不把 task card 当成功能长期设计文档。
- 不在没有对应专题文档更新的情况下启动新的大功能或高风险实现。
- 不让 Admin Provider / Route 草案复制 provider runtime inventory；该专题已关闭，不从现有 Web 原地扩真实 credential / endpoint、production、自动路由、quota 或 billing。
- 不把新的开发测试态 application request quota 改写成生产 quota、token / cost、billing、rate limit 或旧 tenant-only `QuotaSummary` 已就绪。
- Prompt Application 专题已经关闭；不从其现有 assignment / invocation 原地增加 provider retry / fallback、自动 activation / release、replay、agent loop 或生产能力声明。
- Agent / Copilot 专题已经关闭；不从既有 Profile、assignment、Session 或 Run 原地扩 agent loop、工具 / 检索执行、业务写回、自动 activation / release、retry / fallback、replay 或生产能力。
- Prompt / Agent 回归评测专题已经关闭；不从 Case、Suite 或人工 decision 原地扩批量执行、自动发布、重放或生产晋级。
- 不把 Image Adapter 批次 A 至 E 的受控 handoff、本机私有 storage、reference-only profile、test-only fixture client 与一次性私有交付协调解释为真实 backend、credential / endpoint resolver、production object store、public delivery 或图片生成已就绪。
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

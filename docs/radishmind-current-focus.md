# RadishMind 当前推进焦点

更新时间：2026-09-02

## 文档目的

本文档用于回答“根据项目规划和开发进度，今天要做什么以推进开发”。它是新会话短入口，只保留当前阶段、最近结论、下一顺位和停止线；正文默认中文，代码标识符、路径、配置键和状态锚点保留原文。

功能细节默认先进入 [功能设计文档入口](features/README.md) 所定义的专题层级：产品面大方向进入 `docs/features/*.md`，具体功能和复杂页面进入对应子目录，平台横切能力进入 `docs/platform/`，外部接入进入 `docs/integrations/`。实现批次进入 `docs/task-cards/`，长验证记录进入周志、清单、摘要或运行记录。

## 当前结论（默认读取到本节结束）

- 下一长期目标为[工作区成员邀请、认领与到期治理（开发 / 测试态）v1](features/admin-control-plane/workspace-member-invitation-claim-expiry-governance-dev-test-v1.md)，状态为 `workspace_member_invitation_claim_expiry_governance_dev_test_v1_batch_b_completed`。批次 A 已完成 `workspace_invitation.v1`、一次性 256-bit secret policy、filter-bound cursor 与 memory 原子链；批次 B 已追加 SQLite / PostgreSQL `0005_workspace_invitations` durable owner，以 `BEGIN IMMEDIATE` / workspace advisory lock、CAS 和单事务创建既有 membership + catalog-derived role assignment + claimed invitation。双库 migration、v4 → v5、受限 runtime、并发单胜者、重启、损坏载荷、rollback / reapply、database unavailable 与 no-fallback 已通过。下一步停在批次 C 准入前，未经项目所有者再次明确推进不得注册 HTTP，也不得抢跑 Pencil、React、邮件 / 目录搜索、`workspace_admin` 邀请或 production 能力。
- 最新关闭产品顺位为[应用定时回归评测与受控 Campaign 调度（开发 / 测试态）v1](features/user-workspace/application-evaluation-scheduled-regression-campaign-dev-test-v1.md)，状态为 `application_evaluation_scheduled_regression_campaign_dev_test_v1_completed`。P0 与 Batch A 至 D 已完成 canonical / 三存储 owner、strict HTTP、显式 dev/test runner、逐次授权重验、deterministic Campaign / durable Run 交接、完整 Pencil、单一 React strict consumer、SQLite 产品链与重启、PostgreSQL configured gate、双标签 CAS、三视口和隐私审计；任务卡关闭，不派生 Batch E。下一产品顺位回到[功能设计文档入口](features/README.md)，按四产品面真实阻塞重新选择长期目标。真实 Provider、production worker、通用 scheduler、retry / replay、自动 release、业务写回与 production capability 继续关闭。Action Safety Ladder 保持完成关闭，不派生批次 F。
- 最新关闭产品顺位为[工作区 Workflow 模板目录、审查与受控派生（开发 / 测试态）v1](features/workflow/workspace-workflow-template-catalog-review-controlled-derivation-dev-test-v1.md)，状态为 `workspace_workflow_template_catalog_review_controlled_derivation_dev_test_v1_completed`。批次 A 至 E 已完成 strict contract、三种 catalog owner、获人工批准的 Pencil、configured target binding、React strict consumer、双数据库连续链与真实浏览器；SQLite 页面贯通 approved Definition → reviewed/listed template → target Application derive → Draft validate/save → 重启恢复，PostgreSQL configured Server 覆盖 runtime role、no-fallback 与 reconnect。双标签 CAS、workspace / application scope、三视口、隐私与零执行副作用均已复验，唯一任务卡关闭，不派生批次 F。
- 2026-08-19 已完成[应用运行观测与用量归因 v1](features/user-workspace/application-operations-observability-usage-attribution-v1.md)后续准入评审，状态为 `application_operations_observability_usage_attribution_v1_followup_reviewed_no_entry`。仓库没有首分页窗口阻塞真实任务的证据，Gateway Request 与 Workflow Run 仍是独立 owner / cursor，也没有统一 snapshot、时间桶、数据规模、性能预算或正式 billing ledger；因此不启动服务端 summary，不创建 aggregate table、materialized view、跨 store join、API、schema、migration、任务卡或 checker。批次 A 的当前窗口审查与 reported usage 完成事实保持不变。
- 最新关闭产品顺位：[应用结果资产库与受控导出（开发 / 测试态）v1](features/user-workspace/application-result-artifact-library-controlled-export-dev-test-v1.md) 状态为 `application_result_artifact_library_controlled_export_dev_test_v1_completed`。批次 A 至 C 已在唯一 artifact / lifecycle owner 上完成 application-scoped 严格列表、filter-bound cursor、canonical export、独立 export 权限、双数据库读取索引、strict Web consumer、S5 单一 Result Workspace，以及共享双 Session fixture 下的 SQLite 页面重启链与 PostgreSQL configured Server no-fallback / 重启链。专题关闭，不派生批次 D / E、通用 result store、transcript、public share、永久 purge、业务写回或 production 能力。
- 最近关闭产品顺位：[应用会话运行结果资产显式保存与恢复（开发 / 测试态）v1](features/user-workspace/application-session-result-artifact-explicit-retention-dev-test-v1.md) 状态为 `application_session_result_artifact_explicit_retention_dev_test_v1_completed`。批次 A 至 D 已完成五类 session profile、memory / SQLite / PostgreSQL 不可变 artifact 与独立 lifecycle、共享 strict Web consumer、三类 Session 接入、SQLite 服务重启页面恢复，以及 PostgreSQL 配置化 Server 保存 → 归档 → 关闭 no-fallback → 重启读取 → 解除归档产品链；永久 `DELETE` route 仍不存在。专题保持关闭，不派生批次 E、通用 result store 或 transcript。
- 最近关闭产品顺位：[Workflow Definition 绑定受控 HTTP Tool v1](features/workflow/workflow-definition-http-tool-v1.md) 状态为 `workflow_definition_http_tool_v1_completed`。批次 A 至 D 已完成独立 profile、v3 Definition、Definition 来源 action plan / confirmation / audit v2、strict `workflow_run_record.v9`、三种 store、React strict consumer、SQLite 重启恢复与真实浏览器 `1440×900` / `1024×768` / `390×844`。实际开发目标按预期以 transport failure 终结，只产生一个 attempt 和一个 confirmation，零业务写入、零 retry / fallback；刷新和重启均恢复同一 consumed plan 与 v9 run，不重新执行。下一产品顺位回到 [功能设计文档入口](features/README.md)，先选择并更新新的长期功能设计文档，不从本专题派生批次 E、平行 owner 或 gate-only 切片。
- 最近关闭产品顺位：[Gateway Provider Attempt 受控重试与降级执行（开发 / 测试态）v1](features/gateway/provider-attempt-controlled-retry-fallback-execution-dev-test-v1.md) 状态为 `gateway_provider_attempt_dev_test_v1_completed`。批次 A 至 E、七个 Visual R1 代表面、三块 React strict consumer、memory / SQLite / PostgreSQL 产品连续链、三个 unary 协议、真实浏览器 `1440×900` / `720×900` / `390×844` 和最终门禁均已闭合；真实 Provider、非 API Key、同 Profile retry、stream fallback、隐式切换和 production enablement 全部关闭。下一产品顺位回到 [功能设计文档入口](features/README.md) 选择新的长期功能目标，不从本专题派生 S11 或同层 gate-only 切片。
- 最近关闭专题：[Provider 价格策略版本与应用成本审查（开发 / 测试态）v1](features/gateway/provider-pricing-policy-version-application-cost-review-dev-test-v1.md) 状态为 `provider_pricing_policy_version_application_cost_review_dev_test_v1_completed`。批次 A 至 E、memory / SQLite / PostgreSQL、Admin GET / PUT、Request History v2、不可变价格快照、reported usage 整数估算、quota / stream、Visual R1、React strict consumer 和真实浏览器均已关闭。SQLite 完成 v1 → v2、双标签 CAS、重启、API Key / 开发身份、Request History 与 Application Operations；PostgreSQL 完成 migration / runtime role、并发和重连后旧快照不重算。production price、token quota、billing ledger、invoice、全历史成本、自动路由和请求拒绝全部关闭。
- 最近关闭产品顺位为 [RadishMind Family UI 产品化设计与迁移 v1](features/user-workspace/radishmind-family-ui-productization-v1.md)：S9 / S10 React 已分别完成 Visual R3 纵向迁移与真实浏览器复核；本轮不改 Pencil、不建立 S11，也没有新增同层 gate-only 任务卡。该专题关闭后的入口回流已完成，当前新专题以上方 Provider Attempt 设计为准。
- 当前成熟度：内部开发者预览，不使用 `M2` 编号，不声明生产就绪。
- 最近关闭专题：[Workflow RAG 本地知识材料导入、审查与快照构建（开发 / 测试态）v1](features/workflow/workflow-rag-local-material-import-review-snapshot-building-dev-test-v1.md) 已完成批次 A 至 C，状态为 `workflow_rag_local_material_import_review_snapshot_building_dev_test_v1_completed`。Desktop `U4tmEg` 与 Narrow `nI3RW` 局部 Pencil 已人工通过；单一结构化 editor、来源 / fragment 审查、create v1、full replacement v2、双标签 CAS、SQLite 重启、隐私和 `1440×900` / `720×900` / `390×844` 均已形成真实证据。Web `338/338`、production build 和正式 `--workflow-rag-dev` launcher probe 通过；没有新增 API、schema、migration、repository、permission、持久 staging、自动执行或生产声明。下一产品顺位回到 `docs/features/README.md` 选择新的长期功能设计文档，不从本专题派生 S11 或同层 gate-only 切片。
- 最近关闭专题：[Workflow Definition 结构化运行输入（开发 / 测试态）v1](features/workflow/workflow-definition-structured-runtime-inputs-dev-test-v1.md) 已完成批次 A 至 E，状态为 `workflow_definition_structured_runtime_inputs_dev_test_v1_completed`。Draft / Definition v2、executor v2、Run v8、Comparison v7、Session v4、Evaluation Plan / Campaign v2 已在 memory、SQLite、PostgreSQL 形成连续链；SQLite 产品浏览器又完成 Direct Run → Session → 两次 Campaign → Pair Preview → Case / Suite handoff、服务重启、隐私、v1 历史和三视口复核，最终控制台无 warning / error。该历史专题保持关闭；工作区 Workflow 模板目录也已于本日关闭，下一顺位回到功能设计入口。
- [应用评测计划、受控执行与证据归档（开发 / 测试态）v1](features/user-workspace/application-evaluation-campaign-controlled-execution-dev-test-v1.md) 功能状态仍为 `application_evaluation_campaign_controlled_execution_dev_test_v1_completed`。后端 A 至 D、React strict consumer、memory / SQLite exact handoff 与服务重启证据继续成立；SQLite exact Plan `aeplan_lkqe7gr7kjobmf73 v1` 产生两次 succeeded Campaign，Pair Preview 后交接 Case `eval_034d69aec0d7a2323c7f222f v1` 与 Suite `suite_9a8017d686be57009c7ad973`。S10 Desktop `Um8Zh`、Narrow `ZxJd7` 与 Decision R15 `UNMOS` 的 Visual R3 React 迁移已完成：页面使用 selected campaign context、campaign 主 owner、连续 item evidence rows 和单一 Handoff rail；`1440×900`、`720×900`、`390×844` 无横向溢出，campaign 切换与 exact Handoff 交接正常，控制台无 warning / error。
- [应用 API Key 请求配额与 Provider Attempt 准入（开发 / 测试态）v1](features/gateway/application-api-key-request-quota-admission-dev-test-v1.md)批次 A 至 E 已全部完成，功能状态仍为 `application_api_key_request_quota_admission_dev_test_v1_completed`。后端三模式 quota owner、Admin GET / PUT、独立权限和六条 provider 前原子准入继续成立；S9 React 已采用 Visual R3 的 selected application context、单一 quota owner、连续 policy rows、admission rail 与职责圆角。SQLite 真实浏览器完成 missing → create v1 → 双标签 stale CAS → reload v2，`1440×900`、`720×900`、`390×844` 无横向溢出且控制台无 warning / error。旧 User Workspace `QuotaSummary` 仍为 `quota_policy_unavailable`，生产 quota、rate limit、token / cost、billing、正式 membership / OIDC 与自动路由未打开。
- [RadishMind Family UI 产品化设计与迁移 v1](features/user-workspace/radishmind-family-ui-productization-v1.md) 已完成 `v26.7.3` 通用参考基线、`S1 R8` 至 `S8 R1` 的已审视觉语言，以及 S9 / S10 Visual R3 React 迁移与真实浏览器复核。结构化输入 Visual R4 与价格专题 S7 / S5 Visual R1 也已关闭；本专题不继续派生 S11，后续 UI 工作只由新的功能专题和真实使用证据产生。
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
- 外部阻塞的[本地账户与 Radish OIDC 联合登录 v1](features/admin-control-plane/local-account-radish-oidc-federated-login-v1.md)已推进到 `local_account_radish_oidc_federated_login_v1_batch_d_completed_batch_e_external_blocked`。批次 A 至 D 已完成本地身份 owner、三种 repository、Web Session、确定性 browser OIDC、当前账户 / external identity revoke HTTP、显式 opt-in Web gateway、S7 User / Role 当前账户 owner、完整 Pencil 与真实浏览器连续链。批次 E 仍等待真实 Radish 注册条件，不把它改写为当前可执行本地任务。
- 最新关闭产品顺位为[本地用户、角色与工作区成员管理（开发 / 测试态）v1](features/admin-control-plane/local-user-role-workspace-membership-administration-dev-test-v1.md)，状态为 `local_user_role_workspace_membership_administration_dev_test_v1_completed`。批次 A 至 E 已在现有 local identity owner 上完成 canonical 四角色、三存储管理 service、`0003` durable metadata / 顺序索引、显式 one-shot bootstrap CLI、七条 local-session-only strict Admin HTTP、已批准 S7 User / Role Pencil、单一 React strict consumer，以及 SQLite / PostgreSQL configured Server 产品连续链、三视口、双标签与隐私审计；专题关闭，不打开全局账户搜索、自定义角色、真实 Radish 或 production IAM。
- 最新关闭产品顺位：[本地账户凭证轮换与自助会话治理（开发 / 测试态）v1](features/admin-control-plane/local-account-credential-rotation-self-service-session-governance-dev-test-v1.md)状态为 `local_account_credential_rotation_self_service_session_governance_dev_test_v1_completed`。批次 A 至 E 已完成三存储 owner、ordered index、credential replacement + source-bound revoke 原子链、四条 strict HTTP、批准 Pencil、单一 React strict consumer、SQLite / PostgreSQL configured Server 产品链、三视口、双标签和隐私审计；真实窄屏发现并修复 session group 压缩裁切与 account trigger 被 sticky mobile navigation 覆盖两个 CSS 根因。专题关闭，不派生批次 F、S11、设备管理、全局 session console 或 production auth。
- 旧生产凭据后端 / 存储适配器准入链已冻结为历史证据，`storage_adapter_runtime_implementation_entry_refresh_after_provider_account_resource_endpoint_review` 不再是当前开发下一步。

当前最多两条在制主线：

1. 产品线：定时回归评测 P0 与 Batch A 至 D 已完成并关闭；runner 只在显式 dev/test gate 下复用既有 Campaign，并以逐次受限委托、deterministic key 和 no-replay 失败关闭。下一步回到功能设计入口重新选择长期目标，不从已关闭专题派生 Batch E；联合登录批次 E 继续等待 reviewed 真实 Radish 注册条件。
2. 工程线：R2 至 R6 均已关闭，当前没有独立整改批次。后续只在真实功能实现中复用或替代对应门禁；没有等价行为证据的 Provider、Production Ops 与 formal UI 检查继续保留，不按数量清理，也不新建同层治理入口。

## 2026-09-01 今日完成

1. 定时回归评测 Batch D 已完成单一 React strict consumer，Schedule / Version / Occurrence、生命周期确认、exact Campaign handoff 和完整失败态继续复用既有 Plan、Campaign、Run、quota 与 audit owner；local Session 只传 tenant / workspace / product binding，不传 dev identity、permission 或 bearer proof。
2. SQLite 产品连续链完成 Prompt Template v1 → Application Draft v2 → Publish Candidate v3 review → Runtime Assignment v1 activation → Evaluation Plan v2 → quota policy → Schedule v5。exact Occurrence 只创建一个 deterministic Campaign 和一个 durable Prompt Run，终态 `succeeded`，retry / replay 为 `0`；归档后同库重启仍恢复同一 Schedule / Occurrence / Campaign。
3. 双标签 stale CAS、`1440×900` / `720×900` / `390×844`、零横向溢出、选中步骤占满轨道、干净重启 console warning / error `0 / 0`、URL / storage / database 隐私均已通过；页面产品链请求均为 `200`。
4. PostgreSQL 17 configured gate 已通过完整 `internal/httpapi` integration、迁移 `0029_application_evaluation_schedules / workflow_run_store_v29`、受限 runtime role、restart / reconnect、corruption 与 no-fallback；验收后容器和网络已停止，命名 volume 按仓库策略保留。
5. 验收中修复 quota owner 与 Saved Draft 的 local Session transport、Schedule 专用 launcher 的无关通用读请求、Prompt owner 最小 read grant，以及 invocation 未返回 Run 时 Campaign item 未同步进入失败终态的问题；本地角色目录升为 v2，历史 assignment 不重写。专题与任务卡更新为 `completed`，不派生 Batch E；下一步回到四产品面功能设计入口，production 停止线继续关闭。

## 2026-08-23 今日完成

1. 已完成批次 D 完整设计覆盖并修正 Family UI 语言：五维评分 `2 / 1 / 2 / 2 / 2 = 9`，正确设计源第二排的 Authentication Gateway Desktop `scHoA`、Narrow `uR4Yd` 已统一为 Visual R2 的 `Inter + Geist Mono`、冷灰白工作区、深蓝主操作与紧凑信息密度，Decision `SQPBB` 更新为 R19；布局检查无问题，修正提交为 `cb9df403`。
2. Platform 已增加 `GET /v1/auth/account` 与 `POST /v1/auth/external-identities/{binding_id}/revoke`。当前账户 profile 由 memory / SQLite / PostgreSQL 的唯一 repository contract 投影；解绑要求近期认证、ownership、CSRF / Origin、版本 CAS 与至少一种剩余登录方式。
3. 联合登录批次 D 的 Web 已增加显式 opt-in 本地身份 gateway、strict consumer、登录 / 注册 / Radish 登录入口、当前账户、link / revoke / logout 和 metadata-only 跨标签同步；该批次当时的 S7 User / Role 只读取当前本地账户、role assignment 与 membership，不使用离线 fixture 冒充目录事实，随后已由本地成员管理专题升级为 workspace 成员与角色管理面。
4. 真实浏览器完成错误密码、正确登录、当前账户面板、S7 User / Role、双标签登出、刷新、SQLite 服务重启恢复及 `1440×900`、`720×900`、`390×844`；无横向溢出，三标签 console 无 warning / error。临时账户、SQLite 文件、服务与浏览器标签均已清理。
5. “本地用户、角色与工作区成员管理 v1”批次 E 已完成并关闭专题：SQLite 与 PostgreSQL configured Server 均完成双账户注册、显式 bootstrap、membership / role create / revoke、业务权限 `200 → 403` 即时失效、Server 停止 no-fallback 与重启恢复；实链路发现并修复七条 Admin 请求缺少 active tenant header 的 scope mismatch。User 三视口、Role Narrow、双标签 logout、URL / Storage / IndexedDB / Cache / service worker / cookie 属性审计通过，临时账号、数据库、浏览器日志、服务、容器、网络和 volume 已清理。真实 Radish、production session store、MFA、恢复、速率限制、refresh token 与 production auth 继续关闭。
6. Batch B 的 SQLite v2 → v3 与 PostgreSQL v1 → v3、`121` 条同时间戳 cursor、并发 CAS、原子 revoke、受限 runtime role、重启、rollback / reapply、回滚后 no-fallback 和双数据库显式 bootstrap 均通过；`local_workspace_memberships_directory_idx` 已由 SQLite `EXPLAIN QUERY PLAN` 与 PostgreSQL `ANALYZE + EXPLAIN` 验证。PostgreSQL 测试容器和网络已关闭。

## 2026-08-25 今日推进

1. 已从四个一级产品面选择并批准本地账户凭证轮换与自助会话治理，优先补齐注册 / 登录后的用户可感知安全闭环，不恢复普通只读 console、同层 gate-only 或生产 secret 历史链。
2. 批次 A 已完成独立 capability interface、session summary / page / snapshot-bound cursor、exact revoke、revoke others、credential rotation result 与 memory 聚合事务。
3. `121` 条同时间戳分页、snapshot 过期边界、cursor owner / filter 绑定、跨账户 target、recent-auth、CAS、bulk / rotation 坏目标零部分写入、密码复用、当前 local / OIDC session 分流和四争用者并发单胜者已通过。
4. 批次 B 已完成 SQLite / PostgreSQL 同构 owner；只追加 `0004_local_identity_self_service_sessions` 与 ordered index。SQLite / PostgreSQL query-plan、并发 CAS、数据库冲突整事务回滚、v3 / v1 → v4、重启、受限角色、rollback / reapply 与 no-fallback 已通过。
5. 批次 C 已完成四条 strict HTTP、local-session-only actor、CSRF / Origin、recent-auth、current-password proof、confirmation、稳定失败与 cookie 清理语义；success / scope / stale / strict JSON / duplicate submission / zero-side-effect 和响应 / 日志敏感扫描均通过。下一步只进入批次 D，不提前实现双数据库产品连续链。
6. 批次 D 已在现有 Authentication Gateway 页面族完成 Desktop `pOLcz`、Narrow `LMi7H`、credential rotation danger state `n2O8A5` 与 R21 Decision `DASE0` 的 React 采用。单一 strict consumer 严格消费四条 canonical route，security scope 使用 generation / abort 拒绝迟到响应，跨标签只发 metadata-only invalidation；password 和 confirmation 只保留在组件内存并在所有退出路径清理。Web `398/398`、production build、Platform `internal/httpapi` 普通测试和 race 均已通过。

## 2026-08-26 今日完成

1. 在现有 Authentication Gateway 内完成单一 self-service security owner，落地 session directory、exact revoke、aggregate revoke others 与 credential rotation；没有建立 S11、device owner 或第二套 session / credential 状态机。
2. strict consumer 严格校验 canonical schema、稳定 failure boundary 和 actor scope；state projection 将 current / other active / ended 分组，generation / abort 管理迟到响应，mutation 和跨标签 metadata-only signal 失效旧 cursor、selection 与 confirmation。
3. password、confirmation 和 pending credential 只存在组件内存，进入 danger review 前从可见 input state 清空，并在取消、成功、失败、scope 变化、路由离开和卸载时清理。Desktop / Narrow 使用语义 token 与 `900px` / `620px` 响应式边界，不冒充批次 E 的真实三视口验收。
4. 批次 D 的 Web `398/398`、production build、Platform `go test ./internal/httpapi/...`、`go test -race ./internal/httpapi/...` 和仓库 fast / full gate 已通过；随后按任务卡继续进入批次 E，没有提前打开停止线外能力。
5. SQLite `local-product` 完成 session list、exact revoke `1`、revoke others `2`、credential rotation 撤销 `3` 条 local-password session、旧密码 `401`、新密码 `200`、OIDC 保留、双标签失效、服务停止 no-fallback 与同库重启恢复。
6. PostgreSQL 17 由独立 migration 角色应用 `0004`，受限 runtime 角色 DDL 被拒绝；configured Server 完成同构 exact / bulk / rotation、旧密码失败、新密码登录与 OIDC 保留。数据库 hard pause 时没有 fallback，恢复后连接池重新读取 durable session 事实。
7. in-app Browser 完成 `1440×900`、`720×900`、`390×844`、危险确认、forced re-login、双标签 metadata-only logout、console / network / URL / storage source / cookie 属性审计。真实窄屏修复 grid implicit row 压缩与 account trigger stacking 两个 CSS 根因，复验后三视口无横向溢出。
8. Platform、Vite、浏览器标签、PostgreSQL 容器、隔离 SQLite 与 cookie jar 均已清理，`4100`、`7100`、`55439` 无监听。专题关闭；production auth、MFA、恢复、设备管理、全局 session console 与真实 Radish 继续关闭。

## 2026-08-27 今日推进

1. 已完成四产品面顺位评审，并以真实用户闭环、canonical owner、现有事实、外部依赖、本仓库端到端证据和同层延伸风险收敛候选；项目所有者批准工作区内部 Workflow 模板目录作为新的长期目标。
2. 已建立功能专题，固定 approved immutable Definition → template candidate → human review → immutable template version → explicit listing → target Application recheck → Saved Draft v1 的连续链。模板目录只拥有分发与上架事实，不复制 Definition、Draft、Run、Evaluation、membership 或 Application owner。
3. 项目所有者已明确批准并完成批次 A。`contracts/` 新增 candidate / decision、version、lineage、listing event、audit 与 `derivation_v2` 七份 strict schema；Platform 已完成 memory owner、canonical digest、authority reload、portability / target binding validator、review / listing 原子 CAS、strict cursor 与十条默认关闭 HTTP。
4. derive 在全部预检通过后只调用 Saved Draft owner 单次创建独立 v1 草案；既有 `derivation_v1` 兼容保留并与 `derivation_v2` 互斥。精准 / 相邻测试、race、`go vet` 与完整 Platform config / HTTP API 测试已通过，未创建数据库、Pencil、React、服务或浏览器证据。
5. 2026-08-28 首先由项目所有者审查并批准同卡前置修正：让 configured Server 从既有 canonical authority 验证 exact target Provider / Profile / Model binding，并让十条 route 对 query key / cardinality fail closed；完成精准测试、race 与仓库门禁后，再单独决定是否进入批次 D React strict consumer。公开 Marketplace、跨 workspace / tenant、HTTP Tool / RAG 模板、自动模型重绑定、自动上架 / 派生 / activation / run、评分推荐、计费和 production 能力继续关闭；账户安全批次 F、S11 与真实 Radish 外部联调也不恢复。

## 2026-08-30 今日推进

1. Action Safety Ladder 批次 E 已完成单一 HTTP read projection、TypeScript strict consumer、SQLite 页面重启 / legacy / authority drift / 双标签 CAS、PostgreSQL runtime role / reconnect / no-fallback 与三视口产品证据；专题和任务卡关闭，不派生批次 F。
2. 四产品面重新排位后，选择 Prompt Application 定时回归评测作为新长期目标；项目所有者批准非 bearer、schedule-scoped、每次 occurrence 重验的 system actor + delegated user 模型，不采用用户 Session 持久化或页面打开时伪调度。
3. Batch A 已完成 Schedule current / immutable version / Occurrence canonical schema、UTC 日周期、digest、状态机、memory owner、单赢家 claim、missed / overlap 与 no-replay 领域边界。
4. Batch B 已完成十条 strict HTTP、激活时 exact Plan / assignment / API Key 重读、SQLite `0026`、PostgreSQL `0029 / workflow_run_store_v29`、并发、重启 / 重连、corruption 与 no-fallback。
5. Batch C 已完成显式 dev/test gate、固定 `30s` 单 worker、cancel / join、逐次账户 / membership / permission / Plan / assignment / API Key / quota 重读、deterministic Campaign handoff、Run 双 actor metadata 与 crash no-replay；真实 PostgreSQL configured gate、race、vet 与仓库 fast / full 均通过。
6. 日终反查今日五个提交后，已补齐 contracts 索引、Platform 路由 / runner 运维说明、Workflow Run Store 的 Action Safety `0025 / 0028` 与 Schedule `0026 / 0029` 迁移口径、共享 SQLite 生命周期、系统架构、项目总览和跨项目入口中的内部 persistence boundary；未发现需要修改根入口、Web README、脚本或部署边界的代码事实，Schedule 也没有新增任何上层外部执行协议。

## 2026-08-19 今日评审

1. 应用运行观测后续四项准入条件均未满足：没有真实跨页阻塞、没有统一 snapshot / cursor 语义、没有时间桶与性能预算、没有正式 quota / billing owner。评审结论已写回功能专题，不进入服务端 summary 实现。
2. SQLite 真实页面已贯通结构化 Session 执行、逐 turn 显式保存、Application Operations 独立窗口、Result Workspace exact read、Run detail 与 compatible Comparison；输入值未持久化，过期 Session 在 Provider 前失败关闭，页面无 console warning / error。
3. 审计发现的既有 owner 恢复交接缺口已修正：exact Run action 在 `React.StrictMode` 下仍保持一次稳定 owner-scope 初始化，有效目标直接打开精确详情；目标不存在或不在当前 Application scope 时显示稳定说明，不回退其它 Run，也不合成 canonical evidence。三类 Session authority 失败会显示只读 reload、显式选择当前 authority 或审查后新建 Session 的恢复说明，且不会自动切换、创建、重试或调用 Provider。
4. 身份 owner 已修正为 RadishMind 本地账户、角色与 workspace membership；Radish OIDC 只按 `(issuer, subject)` 绑定本地 `user_id`，不按 email 自动合并，也不直接采用 Radish role / permission claim。
5. 联合身份专题批次 A、B 已完成，下一实现入口为批次 C 的确定性 browser OIDC client；工作区运营收件箱批次 B、结果资产批次 D / E、S11、真实 Provider、production auth / secret、billing 与同层 gate-only 切片继续关闭。

## 2026-08-17 今日完成与明天事项

1. 已完成批次 C1：lifecycle current state、append-only event、summary v2、cursor v2、archive / unarchive route、独立组合权限，以及 memory / SQLite / PostgreSQL 同构 CAS 均已成立；既有 artifact payload 与数据库不可变触发器未修改。
2. 已完成批次 C2：单一 strict artifact consumer 与共享面板接入三类 Session surface；`save_result` 逐 turn 默认关闭，执行成功与保存失败分离，active / archived list、exact read、Run handoff 和 archive / unarchive 已贯通。五维评分 `0 / 0 / 0 / 0 / 2 = 2`，采用 `C / 直接实现`，未修改 Pencil、未建立 S11。
3. SQLite 浏览器完成 Prompt 与结构化 Workflow Application 的保存、metadata list、exact read、archive / unarchive；Agent 漂移 assignment 在 Provider 调用前失败关闭。`1440×900`、`720×900`、`390×844`、Web `371/371` 与 production build 通过。
4. 已完成批次 D：memory / SQLite / PostgreSQL 复用同一五 profile fixture；PostgreSQL 17 配置化 Server 完成显式保存、幂等、归档、关闭后 no-fallback、重启精确读取与解除归档，永久 `DELETE` route 返回 `405`。SQLite 页面在服务重启后恢复同一 `appres_oz5tcssa3ysdvdya` 正文与 digest，并完成 active v3 → archived v4 → active v5；当前页面无横向溢出且日志无 warning / error。
5. 已选择新的独立长期专题“应用结果资产库与受控导出 v1”并进入批次 A：只打开 application-scoped discovery 和单 artifact canonical JSON export；public share、批量导出、永久删除、自动清理、transcript、长期记忆、replay / resume、真实 Provider、production secret / auth / store、自动执行和业务写回继续关闭。
6. 已完成新专题批次 A：summary v2 / export v1 合同、application cursor v3、三存储 application-scoped list、独立 export permission、SQLite `0024` / PostgreSQL `0027` 索引与 HTTP route 已落地。memory 超过 `100` 条同时间戳分页、SQLite 重启、PostgreSQL migration / runtime role / configured Server no-fallback 与重启、完整 `internal/httpapi` 均已通过；测试容器已关闭。
7. 已完成批次 B / C：S5 单一 Result Workspace、strict Web consumer、筛选、exact inspector、Run handoff、lifecycle CAS 与二阶段校验后下载已经闭合。共享 fixture 覆盖两个 Session、两种 profile、两种 content type 和 active / archived；SQLite 页面完成 archive / unarchive 与重启恢复，PostgreSQL configured Server 完成关闭 no-fallback 与重启恢复。Web `375/375`、production build、三视口和 PostgreSQL 17 聚合集成通过，测试服务与容器已关闭。
8. 应用结果资产库专题状态推进为 `application_result_artifact_library_controlled_export_dev_test_v1_completed`。下一轮先选择新的长期功能设计文档；工作区运营收件箱批次 B、真实 Provider、production auth / secret 和 Radish 接线仍等待各自真实需求或外部前置，不以 gate-only 续批代替产品推进。
9. 原定 2026-08-18 开始的应用运行观测后续功能设计评审已于 2026-08-19 完成；因真实任务、统一 snapshot / cursor、性能预算与正式 billing owner 均不足，结论为不进入实现。

R3 与 [工作流草案 PostgreSQL 开发测试态存储库 v1](features/workflow/saved-workflow-draft-postgresql-dev-test-repository-v1.md) 已于 2026-07-11 完成。`postgres_dev_test` 已覆盖迁移 / 回滚 / 重新应用、运行角色 DDL 拒绝、服务重启恢复、原子预期版本校验、租户 / 工作区 / 应用 / 所有者作用域、不回退、CI 与真实浏览器双标签冲突审查。该完成不启用生产存储库模式，也不代表 OIDC、生产凭据、审计存储或公开生产 API 已就绪。

持久草案存储库、稳定 Gateway、执行器 v0 与持久开发测试态运行历史均已成立。真实浏览器已验证“创建 → 保存 → 启动受限运行 → 分页历史 → 详情 → 服务重启恢复”，运行记录中模型服务调用为 1，工具、确认、业务写入和重放均为 0，原始输入与条件值未持久化。无限制工具、业务写回、自动确认提交、重放和恢复继续关闭。

总入口与证据：

最新完成入口：[应用定时回归评测与受控 Campaign 调度（开发 / 测试态）v1](features/user-workspace/application-evaluation-scheduled-regression-campaign-dev-test-v1.md)。P0 与 Batch A 至 D 已完成，[唯一实施任务卡](task-cards/application-evaluation-scheduled-regression-campaign-dev-test-v1-plan.md)已关闭；当前回到[功能设计文档入口](features/README.md)选择新的长期目标。

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
47. [本周周志](devlogs/2026-W35.md)
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

- 定时回归评测 P0 与 Batch A 至 D 已完成并关闭；不得从已关闭专题派生 Batch E，也不得越界启动真实 Provider、production worker、通用 scheduler、cron / queue、retry / replay 或 production capability。始终不冒用用户 actor、不持久化 Session / token、不引入通用 execution token。Action Safety Ladder 继续关闭且不派生批次 F。
- 不让模型、客户端、Profile、Workflow 草案、candidate decision、assignment、人工批准或页面状态自报或提升 effective level；`write_allowed_by_policy` 在 v1 始终不可达。
- 不为 Safety Ladder 创建第二套 Action、Confirmation、Run、Audit、Result 或通用 decision owner，也不新增万能 execute permission 或通用动作执行 API。
- 不把工作区模板目录扩成公开 Marketplace、跨作用域共享、推荐 / 评分、自动重绑定、自动运行或生产发布。
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

# 用户工作区细专题入口

更新时间：2026-10-26

本目录承接用户工作区中跨应用、模型发现、接入、调用与审查的具体功能专题。产品面长期边界继续以 [用户工作区设计与开发文档](../user-workspace.md) 为准。

## 当前专题

- [本机持久工作台交付 v1](local-persistent-workbench-delivery-v1.md)：当前主线，实施已确认，统一配置、启动命令、Web 接线和恢复已实现；宿主机浏览器主链、重启与恢复已验收；`linux/arm64` Docker 运行与备份迁移已验收，独立用户验收待完成。统一正式 Web 的创建、运行、结果保存与重启恢复路径。
- [应用跨版本回归审查（开发 / 测试态）v1](application-cross-version-regression-review-dev-test-v1.md)：候选、未排期、未实现；持续使用证明批量版本审查负担后再决策，既有配对规则保持。
- [应用定时回归评测与受控 Campaign 调度（开发 / 测试态）v1](application-evaluation-scheduled-regression-campaign-dev-test-v1.md)：状态为 `application_evaluation_scheduled_regression_campaign_dev_test_v1_completed`。受限委托 P0 与批次 A 至 D 已完成规范契约与三种存储实现、严格 HTTP 接口、逐次授权重验、显式开发测试态运行器、既有评测批次 / 运行记录交接、崩溃后不重放、完整 Pencil、单一 React 严格消费端、双数据库产品链、CAS、重启 / 重连、三视口与隐私审计；任务卡关闭，不派生批次 E。
- [应用结果资产库与受控导出（开发 / 测试态）v1](application-result-artifact-library-controlled-export-dev-test-v1.md)：状态为 `application_result_artifact_library_controlled_export_dev_test_v1_completed`；批次 A 至 C 的应用作用域严格列表、过滤游标、规范导出、独立导出权限、双数据库索引、严格 Web 消费端、三视口页面与双数据库产品连续链已完成，专题关闭且不复制结果资产 / 生命周期负责模块。
- [应用会话运行结果资产显式保存与恢复（开发 / 测试态）v1](application-session-result-artifact-explicit-retention-dev-test-v1.md)：批次 A 至 D 已完成并关闭；默认关闭的显式保存、三种存储中的不可变结果资产、版本化生命周期、共享严格 Web 消费端、SQLite 重启页面恢复与 PostgreSQL 配置化产品链均已成立，不改变运行历史 / 会话仅元数据契约，也不派生批次 E、通用结果存储或会话正文。
- [RadishMind Family UI 参考图产品面映射 v1](radishmind-family-ui-reference-mapping-v1.md)：已把 family-ui `references.md` 的 `ref-01` 至 `ref-27` 逐项映射到 S1–S8 八个产品面，固定实际查看、共享转译、禁止照搬内容、Pencil 构件与版权停止线。
- [RadishMind Family UI 产品化设计与迁移 v1](radishmind-family-ui-productization-v1.md)：family-ui `v26.7.3` 参考基线、RadishMind Workbench 选择和项目语义层已经对齐；`S1 R8` 至 `S8 R1` 已完成设计、React 与真实浏览器验收，S9 / S10 功能、Pencil Visual R3、React 迁移与真实浏览器复核也已完成。
- [应用评测计划、受控执行与证据归档（开发 / 测试态）v1](application-evaluation-campaign-controlled-execution-dev-test-v1.md)：后端 A 至 D、S10 React 严格消费端、内存 / SQLite 三视口精确交接和重启恢复均已完成；结构化输入专题又在原负责模块上补齐定义 v2 类型化测试样例、比较 v7、三种存储的连续链与用例 / 套件精确交接。2026-10-15 新增 Prompt 计划结构化编辑已完成前端实现与无服务验证；新增双语连续交互、焦点及三视口浏览器验收已于 2026-10-21 完成。Workflow 结构化样本已使用表单 / JSON 单一状态、全部样本校验和修改后撤销确认；Definition 引用须显式读取、核对并采用，双语隔离验收完成。
- [Prompt / Agent / Copilot 类型工作区产品化 v1](prompt-agent-copilot-type-workspace-productization-v1.md)：S8 已完成。既有模板 / 档案、配置、候选版本、运行时绑定、接入、会话 / 调用、运行记录与评测负责模块被编排为七 / 八任务、单一负责模块的工作区；开发测试态停止线、`A` 级 Pencil 与真实浏览器证据均已关闭。
- [Workspace-scoped Mutation Authorization / 工作区写入与审查动作成员资格绑定（开发 / 测试态）v1](workspace-scoped-mutation-authorization-dev-test-v1.md)：批次 A 至 E 共 47 条写入操作已完成身份 / 成员资格双重权限、当前活动工作区唯一选择、原子组合与条件权限、稳定拒绝映射和零业务 / 外部副作用证据；专题关闭。
- [工作区运营收件箱（开发 / 测试态）v1](workspace-operations-inbox-dev-test-v1.md)：批次 A 已完成四类既有负责模块的首分页关注项、来源覆盖、稳定严重度、Web 既有详情跳转和工作区切换失败关闭；不新增运营或修复真相源。
- [Workspace-scoped Read Transition / 工作区选择与成员资格绑定（开发 / 测试态）v1](workspace-scoped-read-transition-dev-test-v1.md)：批次 A、B 已完成共享成员资格提供方、五类路由授权、四类持久化负责模块的读取投影、工作区范围运行记录游标与非持久化 Web 选择器；配额和生产成员资格继续关闭。
- [Prompt / Agent 应用回归评测与发布审查（开发 / 测试态）v1](prompt-agent-application-regression-evaluation-release-review-dev-test-v1.md)：Prompt v6 / Agent v7 已严格接入既有比较、评测用例、套件与人工决策；SQLite Agent 用例 → 套件 → `approved v1` 和隐私复验均已完成，专题关闭。
- [Agent / Copilot 应用档案版本审查与受控建议（开发 / 测试态）v1](agent-copilot-application-profile-version-review-controlled-suggestion-dev-test-v1.md)：批次 A 至批次 E 已完成；档案、发布与运行时绑定、唯一受控建议、会话 / 运行记录、类型专属 Web 与双数据库真实验收形成闭环，专题关闭。
- [Agent / Copilot 开发测试态使用指南](agent-copilot-dev-test-usage-guide.md)：说明档案 → 配置草案 v4 → 候选版本 v4 → 运行时绑定 → API 密钥 / 会话 v3 → 运行记录 v7 的操作顺序、启动配置、身份权限、CAS、持久化、隐私和排障边界。
- [提示词应用模板版本审查与受控调用（开发 / 测试态）v1](prompt-application-template-version-review-controlled-invocation-dev-test-v1.md)：批次 A 至 E 均已完成，受限模板、双数据库模板负责模块、配置草案 v3、发布候选版本 v3、显式运行时绑定、受控调用、会话 / 交互轮次 v2、运行记录 v6、Web 与真实浏览器验收均有可复验证据，专题关闭。
- [提示词应用开发测试态使用指南](prompt-application-dev-test-usage-guide.md)：说明完整的模板 → 配置草案 → 候选版本审查 → 运行时绑定 → 调用 / 会话 → 运行记录审查顺序，以及启动配置、身份权限、CAS、持久化与故障处理；所有能力仅限开发测试态。
- [应用开发工作区与发布准备审查 v1](application-development-workspace-release-readiness-review-v1.md)：新增 [Workflow 创建、版本迭代与回归审查方案](application-development-workspace-release-readiness-review-v1.md#workflow-创建版本迭代与回归审查方案2026-10-21)，四项问题已完成实现和隔离验收；局部 Pencil、v2 草案编写 / 历史 / 派生、晋级隔离 / 冲突恢复及评测编辑已接通，后续范围见当前推进焦点。原批次 A 至 C 已完成并关闭；路由作用域证据、草案 / 运行记录负责模块的精确重读、离线修订版本失败关闭、真实浏览器连续路径与 URL / 控制台 / 网络隐私审计均有可复验证据。
- [应用受控运行开发测试态指南](application-controlled-runtime-dev-test-guide.md)：说明 Application RAG、Workflow Definition v1 / v2、Application Interaction Session v1 / v4、v4 / v5 / v8 运行记录与 Application Operations 的启动、资源准备、作用域、恢复、失败语义和隐私边界。
- [应用交互会话与受控运行编排（开发 / 测试态）v1](application-interaction-session-controlled-runtime-orchestration-dev-test-v1.md)：严格契约、三种会话 / 交互轮次负责模块、权威来源精确重读、v5 / v4 单次委托、Web 易失交互工作区、双数据库启动器连续链、重启恢复、真实浏览器和敏感信息扫描均已完成，专题关闭。
- [API 密钥生命周期与 Gateway 开发测试态认证 v1](api-key-lifecycle-gateway-dev-test-auth-v1.md)：Gateway 认证、统一 `sqlite_dev` 存储库 / 聚合运行时、双数据库门禁、Web 一次性交接、真实浏览器连续路径、重启恢复与敏感信息复验均已完成，专题关闭。
- [API 密钥引导式轮换与验证后退役（开发 / 测试态）v1](api-key-guided-rotation-verified-retirement-dev-test-v1.md)：批次 A、B 已完成；易失脱敏会话、同作用域替代、`last_used_at` 验证门槛、精确退役 CAS 与真实浏览器连续链均有可复验证据，专题关闭。
- [应用目录与生命周期（开发/测试态）v1](application-catalog-lifecycle-dev-test-v1.md)：核心生命周期、内存与 PostgreSQL 开发测试态存储、Web 管理、下游归档只读约束和真实浏览器连续验收均已完成。
- [应用解除归档与安全重新启用（开发 / 测试态）v1](application-unarchive-safe-reactivation-dev-test-v1.md)：批次 A 至 C、三种存储的 CAS、组合权限、显式影响确认、Gateway 资格回归、Web 与真实浏览器连续验收均已完成，专题关闭。
- [应用 API 接入与调用 v1](application-api-integration-invocation-v1.md)：把选中应用、`/v1/models` 模型目录、三协议接入示例、现有 Gateway 调试台调用与脱敏请求历史审查串成连续的内部开发者路径。
- [应用配置草案与审查 v1](application-configuration-draft-review-v1.md)：为当前应用建立独立配置草案、校验、开发测试态持久化、版本冲突、比较和 API 接入交接。
- [应用发布治理与晋级审查 v1](application-publish-governance-promotion-v1.md)：已完成不可变候选版本、版本绑定、审查 CAS、漂移识别、阻塞式晋级资格判断，以及既有接入区、调试台和请求历史交接；不直接发布正式应用。
- [应用运行观测与用量归因 v1](application-operations-observability-usage-attribution-v1.md)：已完成应用作用域 Gateway Request History 与 Workflow Run History 的独立来源覆盖、当前窗口归因摘要和合并时间线；2026-08-19 后续准入评审为 `no_entry`，不启动跨页 summary，也不推测跨来源关联或估算 token、成本、配额、计费。

## 下一步

[本机持久工作台](local-persistent-workbench-delivery-v1.md)的启动、配置、数据、身份、Docker 资产与恢复实现已提交；2026-10-31 已按明确运行授权完成宿主机主链、冷重启与新目录恢复验收，所属进程已清理；另行授权的 `linux/arm64` Docker 构建 / 运行、挂载限制和容器与宿主机备份迁移也已验收并清理；独立用户试用仍待安排。Prompt 与 Workflow 共用交付基础，保留各自运行规则；真实模型质量另在恢复窗口后验证。

已完成的应用、会话、结果、评测、调度与授权专题继续作为实现事实源。技术交付需要的修正回写对应专题，不重新建立平行领域模块，不为每个普通步骤追加任务卡或检查器。

跨版本批次审查、跨页服务端汇总、长期记忆、Agent 循环和生产能力不因本次交付自动进入。候选的触发依据见[路线图](../../radishmind-roadmap.md#后续开发方向2026-10-26)；应用运行观测后续仍为 `no_entry`。

## 目录停止线

- 应用配置草案只允许建立独立开发测试态存储库，不改变 Gateway 上行协议 schema、模型服务注册表或正式应用真相源。
- 开发测试态 API 密钥必须显式启用、只绑定活跃应用且失败关闭；不把它晋级为生产 API 密钥，也不并行打开配额、计费、自动回退、负载均衡或生产认证。
- 应用创建、发布、删除和业务写回必须由独立功能设计承接，不并入既有接入与调用工作区。
- 应用解除归档只恢复目录活动态；不自动创建、吊销、激活或改写 API Key、运行时绑定、会话、草案、候选、定义和运行记录。

# RadishMind 当前推进焦点

更新时间：2026-11-25（阶段 0 复评通过，进入正式 Web 单 Agent 技术交付）

## 文档职责

本文只回答当前目标、下一动作、阻塞和停止线。产品范围见[产品范围与目标](radishmind-product-scope.md)，实现事实见[能力矩阵](radishmind-capability-matrix.md)与[功能专题](features/README.md)，阶段选择见[路线图](radishmind-roadmap.md)，历史证据见[周志](devlogs/README.md)。

正文默认中文，代码标识符、路径、配置键和状态锚点保留原文。

## 当前结论（默认读取到本节结束）

- 2026-11-14 已停用 Pencil。单 Agent 已确认的六幅稿继续用于阶段 1，新 UI 任务按[代表设计协作模型](features/user-workspace/radishmind-family-ui-productization-v1.md#代表设计协作模型)评审。

- **当前开发主线为阶段 1 的正式 Web 单 Agent 技术交付。** 阶段 0 四项退出条件已通过，正式采用冻结的 OpenHands SDK `1.59.0+radishmind.1` 作为固定 Go 单 Agent 的受限编码执行层。公开合同、范围与 A 级画板已获项目所有者确认。[阶段 0 结论](task-cards/development-task-probe-v1-plan.md#2026-11-25-sdk-接收确认与阶段-0-采用复评)
- [本机持久工作台](features/user-workspace/local-persistent-workbench-delivery-v1.md)继续作为交付基座；宿主机与 `linux/arm64` 容器已有技术证据，独立用户验收未完成且按要求暂不安排。Windows PowerShell、`linux/amd64` 和真实模型质量不由已有证据推定通过。
- 成熟度保持**内部开发者预览**。本机交付、真实任务效果与按需共享 / 业务接入的顺序保持；跨版本回归审查仍为候选。真实模型试用继续暂停，通道默认关闭，未开放正式执行能力。

<a id="next-actions"></a>
<a id="接下来事项2026-11-05"></a>
<a id="接下来事项2026-10-31"></a>

<a id="接下来事项2026-11-12"></a>

<a id="接下来事项2026-11-18"></a>

<a id="接下来事项2026-11-19"></a>

<a id="接下来事项2026-11-21"></a>

## 接下来事项（2026-11-25）

已落地公开 canonical schema 和角色目录 v3。管理员可通过现有角色分配流程显式授予 `workspace_developer`，已有 v1 / v2 角色分配保留原授权。[公开交付合同](contracts/development-task-delivery-v1.md)

共享 SQLite 已落地任务、不可变计划修订、确认、六类请求回执及工作台唯一执行槽。隔离库已验证并发启动只签发一次，取消与保存失败可回滚，重启后旧尝试保留原身份和未决资源。

**下一步接入正式 Web 会话与严格 HTTP。** 先把当前会话、精确工作区权限和固定响应档案接到任务协调器，按公开合同提供元数据写入、回执恢复和精确读取。随后接入提交后的工作进程交接、逐项资源观测和旧资源实际核对，再验收 Web 与隔离执行。[阶段 1 实施与验收](features/workflow/controlled-software-development-team/implementation-and-evaluation-plan-v1.md#阶段-1单-agent-隔离开发交付)

已有原生 HTTP 固定上游技术条件、SDK 工具往返与联合任务证据保持，只有候选 / 配置变化或相关失败触发复验。真实厂商资格、完整输入计量和费用窗口仍未验收，真实模型暂停、通道默认关闭。

[来源裁决](features/workflow/controlled-software-development-team/framework-and-gateway-integration-v1.md#closeout-decision)分别标清实际使用 / 交付依据、既定来源要求、额外复现研究；不是把全部未知自动叠成同一门槛，也不自动豁免已接受要求。下一批须明确投入上限、复核点和结束决定；路径不成立就改选或等待，不维护多个候选，不继续派生来源链或通用构建体系。

阶段 1 只用固定三文件 Go、单模型串行工具、一次尝试、受保护测试和显式保存；通用 Git、任意命令 / 依赖、团队 / 群聊返修、额外数据库 / 平台、完整路由解释 / 回放与自动合入后置。运行详情仅交付合同必需且可追溯的信息；未知值和未确认终止不得伪装为成功。

**阶段 1 技术出口后，先做小范围真实单 Agent 价值复核，再决定是否进入阶段 2。** 模型窗口仍暂停；恢复须明确提供方资格、有限任务、数据和费用授权，记录完成质量、人工干预、失败与成本，形成继续 / 修正 / 暂缓团队的决定。未取得这些依据时保留技术成果，不用新增团队系统填补等待期。完整团队对比仍在阶段 4，详见[价值复核点](features/workflow/controlled-software-development-team/implementation-and-evaluation-plan-v1.md#single-agent-value-review)。

本机工作台的独立用户首次 / 持续使用仍是交付待办，按所有者要求暂不安排；Windows PowerShell、linux/amd64 与真实模型质量不由已有证据推定。

## 阶段顺位与进入条件

| 顺位 | 目标 | 完成或进入条件 |
| --- | --- | --- |
| 交付待办 | 本机持久工作台独立使用 | 宿主机与容器技术证据已完成；补独立用户按指南持续使用的证据 |
| 已完成 | 受控软件开发团队阶段 0 | 四项退出条件复评通过，冻结 SDK 在已确认的范围正式采用 |
| 当前开发 | 阶段 1 单 Agent 隔离开发交付 | Web 可发起 / 取消 / 审阅 / 保存候选并精确重读；技术出口后先作真实价值复核，再决定团队 |
| 阶段 1 后的决定点 | 真实单 Agent 价值复核 | 技术交付后恢复获准模型窗口，记录有限任务的质量 / 人工成本 / 失败 / 费用；据此确认是否进入固定团队，窗口缺失则等待 |
| 后续团队与完整比较 | 固定团队及阶段 4 效果比较 | 单 Agent 价值复核支持继续投入并确认团队范围后才推进；完整比较沿原可比基线 |
| 按需求选择 | 团队共享或业务接入 | 实际用户 / 环境 / 负责人或上层挂载点明确；选一个目标交付，不同时铺开 |
| 候选 | 跨版本回归审查等增强 | 有持续使用证据，能说明现有手工路径的具体阻塞与收益 |

外部条件提前齐备时可以明确重排下一窗口；不自动并行开多条主线。真实故障与安全问题优先处理，具体条件见[路线图](radishmind-roadmap.md#后续开发方向2026-10-26)。

已登记未来能力：[Gateway 路由决策解释与尝试轨迹](features/gateway/gateway-routing-decision-explanation-attempt-trace-v1.md)。随阶段 1 运行详情设计统筹，首版仅采用公开交付合同必需且有可信关联的运行信息；完整调用链解释、历史回放和策略优化后置，不阻塞单 Agent 出口；应用观测 `no_entry` 保持。

## 已完成范围与剩余边界

- 邀请治理、中英国际化、定时回归、Workflow 模板与 Action Safety 保持关闭；工程 R2 至 R6、身份策略首个切片与既有双数据库回归不派生续批。实现事实从[功能入口](features/README.md)追溯。
- 2026-10-15 RAG 晋级列表失败已修复并完成本地复验；2026-10-21 Prompt / Workflow 编辑与恢复、10 月 26 日首次使用指引及悬停修正均保持完成。分窗口证据见 [W42](devlogs/2026-W42.md)、[W43](devlogs/2026-W43.md)、[W44](devlogs/2026-W44.md)，不合称全站或真实用户验收。
- 本地控制台保持 `local usable / read-only close`，不再默认继续补同类只读 console 小切片；production secret backend、process supervisor、部署环境隔离和 console production packaging 仍为 `not_satisfied`。正式 Web 的本机交付不默认扩展旧 Console 或授予生产能力。
- 现有评测比较运行元数据；同计划版本批次配对不等于跨版本审查，也不判断答案质量。结果正文按 Session / Result Artifact 的既有显式保留规则处理。

## 保持关闭的边界

- 邀请不发送邮件、不搜索目录，不邀请 `workspace_admin`；邀请码不直接授予权限，认领仍由既有身份 / 成员 / 角色负责模块原子处理。
- 真实 Radish OIDC 联调保持 `real_radish_integration_deferred`。本地会话、开发身份头、签名测试成员断言和回环地址签发方不能作为生产授权证据。
- 生产密钥后端、生产认证、生产 API 密钥 / 配额 / 计费、真实模型提供方故障回退、生产工作进程、进程监管和生产交付继续未满足准入条件。
- 通用自治智能体循环、无限制工具、写方法工具、业务写回、自动确认提交、自动发布、重试 / 重放扩围继续关闭；Action Safety 不开放 `write_allowed_by_policy`。受控开发团队仅规划专属有界角色回合、隔离副本写入与命令档案，不放宽旧 Agent / Copilot 或通用工具入口，当前仅实现固定 Go 探针的专属执行档案，未开放通用仓库、真实模型或正式执行入口。
- [应用运行观测与用量归因](features/user-workspace/application-operations-observability-usage-attribution-v1.md)后续仍为 `no_entry`：没有首分页窗口阻塞真实任务的证据时，不创建服务端汇总、跨存储库关联查询或计费账本。
- 普通 UI、文案、布局和只读证据整理复用既有测试与门禁；不恢复准入 / 审查 / 刷新历史尾链，不因专题关闭派生新任务卡或检查器。
- 未明确授权不启动常驻服务、Docker、真实模型长跑、模型或数据下载、生图后端、发布部署；不跨外部项目工作区写入。
- API key、token、DSN、一次性凭据、未经脱敏的真实敏感输入和模型提供方原始错误不得进入公开日志、URL 或已提交资产。

## 读取与验证

历史索引：[R6 文档与检查器收敛](platform/engineering-health-productization-remediation-v1.md#r6文档checker-与资产收敛)、[用户工作区应用目录与生命周期（开发/测试态）v1](features/user-workspace/application-catalog-lifecycle-dev-test-v1.md)保持关闭，不提供执行顺位。

1. 先读本文件与当前交付专题；架构、身份、存储和使用指南只按任务需要读取。历史任务卡与周志用于追溯，不决定排期。
2. 日常验证使用 `./scripts/check-repo.sh --fast`，Windows 使用 `pwsh ./scripts/check-repo.ps1 -Fast`；变更真相源、阶段或治理口径时补跑全量检查。
3. 依赖准备、服务、模型与外部操作按[协作规则](agent-collaboration.md)落实范围；不因文档规划自动执行。

# RadishMind 当前推进焦点

更新时间：2026-11-01（本次整理导航，阶段决策沿用原记录）

## 文档职责

本文只回答当前目标、下一动作、阻塞和停止线。产品范围见[产品范围与目标](radishmind-product-scope.md)，实现事实见[能力矩阵](radishmind-capability-matrix.md)与[功能专题](features/README.md)，阶段选择见[路线图](radishmind-roadmap.md)，历史证据见[周志](devlogs/README.md)。

正文默认中文，代码标识符、路径、配置键和状态锚点保留原文。

## 当前结论（默认读取到本节结束）

- **当前开发主线为受控软件开发团队阶段 0。** [集中任务卡](task-cards/development-task-probe-v1-plan.md)记录任务结果，[实施专题](features/workflow/controlled-software-development-team/implementation-and-evaluation-plan-v1.md#阶段-0-采用复评与剩余退出条件)维护退出条件。联合固定响应技术链及遥测可选化候选已复验；剩余许可、原生 HTTP 传输、阶段 1 范围及 A 级视觉依据尚未闭合，正式采用暂不通过，阶段 0 不退出。
- [本机持久工作台](features/user-workspace/local-persistent-workbench-delivery-v1.md)继续作为交付基座；宿主机与 `linux/arm64` 容器已有技术证据，独立用户验收未完成且按要求暂不安排。Windows PowerShell、`linux/amd64` 和真实模型质量不由已有证据推定通过。
- 成熟度保持**内部开发者预览**。本机交付、真实任务效果与按需共享 / 业务接入的顺序保持；跨版本回归审查仍为候选。真实模型试用继续暂停，通道默认关闭，未开放正式执行能力。
- 既有授权窗口已结束，当前规划不延续下载、Docker 或服务运行授权；详细运行事实与清理记录沿任务卡和[W44 周志](devlogs/2026-W44.md)查证。

## 接下来事项（2026-10-31）

接下来按以下顺序推进，不并行扩张实施主线：

1. **先收口许可证据与交付范围。** 优先核对 cryptography 内嵌 OpenSSL 4.0.3 的精确来源 / 构建映射与 regex 历史 CNRI 条款，再落实 fastmcp-slim 等缺失告知、原生 / 内嵌组件清单及最终告知 / 覆盖源码交付映射。先复用已保留材料列出准确缺口；新增下载、依赖调整、重锁或重建须另报范围并取得授权。源码已取得不等于整镜像许可条件通过。
2. **确认并补齐原生 HTTP 传输。** 按[框架专题](features/workflow/controlled-software-development-team/framework-and-gateway-integration-v1.md#原生传输的建议实施范围待确认)收口单一非流式档案，实施前确认范围；用固定上游验证实际 HTTP 工具往返、错误 / 超时 / 取消、未知用量及凭据边界，保留真实厂商 / 模型资格待验。
3. **确认阶段 1 的单 Agent 固定 Go 交付合同与 A 级视觉依据。** 明确公开权限 / API、生命周期与重启行为、候选审阅 / 保存 / 导出，再复评阶段 0；退出条件齐备前不提前开发正式 Web 或固定团队。

上述事项是后续推进计划，不延续已结束的下载、Docker 或服务运行授权。四项退出条件见[实施专题](features/workflow/controlled-software-development-team/implementation-and-evaluation-plan-v1.md#阶段-0-采用复评与剩余退出条件)；正式采用暂不通过，阶段 0 不退出。SDK 暂停不是立即取消，缺失 usage 的零值不能用作平台账本；SDK / 代码沙箱分层窗口及本次联合窗口均已授权、验证并清理，所属容器与测试监听器全部关闭，镜像保留；未运行真实模型或用户试用。

**工作台剩余交付：独立用户按运行指南完成首次使用与持续使用；按项目所有者要求暂不安排。** 记录完成情况、求助点和恢复成本后修复，AI 技术验收不代替该项；它与新功能设计分别记录。Windows PowerShell 与 `linux/amd64` 容器仍未验收，不由已有平台结果推定通过。

技术证据、临时材料路径与清理记录见[工作台专题](features/user-workspace/local-persistent-workbench-delivery-v1.md)；入口整理前的日期化摘要保留在[规划整理记录](devlogs/2026-W44.parts/planning-structure-20261101.md)。

不自动并行开启多条实施主线。当前只按明确要求推进受控软件开发团队阶段 0；真实模型继续暂停，独立使用证据未齐前不关闭工作台整体交付。已有配置 / 身份 / 存储和高风险执行契约保持，通用仓库准入与正式长任务交付仍需独立设计和验收，固定 Go 探针证据不替代这些条件。

## 阶段顺位与进入条件

| 顺位 | 目标 | 完成或进入条件 |
| --- | --- | --- |
| 交付待办 | 本机持久工作台独立使用 | 宿主机与容器技术证据已完成；补独立用户按指南持续使用的证据 |
| 当前开发 | 受控软件开发团队阶段 0 | [四项退出条件](features/workflow/controlled-software-development-team/implementation-and-evaluation-plan-v1.md#阶段-0-采用复评与剩余退出条件)闭合并复评通过后进入阶段 1；任务结果只在集中任务卡维护 |
| 下一实施目标 | 单 Agent 隔离开发交付 | 阶段 0 条件满足并确认实现范围；Web 可发起 / 取消 / 审阅 / 保存候选，再推进固定团队 |
| 后续 | 真实任务效果验证 | 恢复模型窗口，明确任务集、提供方 / 模型、数据与费用；记录质量和使用成本 |
| 按需求选择 | 团队共享或业务接入 | 实际用户 / 环境 / 负责人或上层挂载点明确；选一个目标交付，不同时铺开 |
| 候选 | 跨版本回归审查等增强 | 有持续使用证据，能说明现有手工路径的具体阻塞与收益 |

外部条件提前齐备时可以明确重排下一窗口；不自动并行开多条主线。真实故障与安全问题优先处理，具体条件见[路线图](radishmind-roadmap.md#后续开发方向2026-10-26)。

已登记未来能力：[Gateway 路由决策解释与尝试轨迹](features/gateway/gateway-routing-decision-explanation-attempt-trace-v1.md)。随阶段 1 运行详情设计统筹，先只读解释与历史回放，策略优化后置；当前 SDK / 隔离执行顺位及应用观测 `no_entry` 保持，不自动开启第二条实施主线。

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

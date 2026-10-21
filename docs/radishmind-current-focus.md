# RadishMind 当前推进焦点

更新时间：2026-10-21

## 文档职责

本文只回答当前目标、下一动作、阻塞和停止线。产品范围见[产品范围与目标](radishmind-product-scope.md)，实施事实见[能力矩阵](radishmind-capability-matrix.md)和对应[功能专题](features/README.md)，审阅数据与批次记录见[2026-W42 周志](devlogs/2026-W42.md)。

正文默认中文，代码标识符、路径、配置键和状态锚点保留原文。

## 当前结论（默认读取到本节结束）

- 当前成熟度为**内部开发者预览**。工作流、应用运行与评测、模型网关、本地身份和开发测试态持久化已有连续实现；专题完成不等于真实模型提供方、团队试用或生产运行已通过验收。
- 2026-10-15 已关闭开发阶段 RAG 晋级列表失败：修复共享请求标识与敏感材料检查的碰撞，确定性回归及四组本地中英 44/44 浏览器复验通过，所属进程、端口和临时数据库已清理。未重跑远端 CI 或调用真实模型；证据见[2026-W42 周志](devlogs/2026-W42.md)。
- [Web 中英国际化与本地化 v1](features/web-internationalization-zh-en-v1.md)已在登记的现有可达 Web 范围内完成 A 至 F 并关闭。身份、管理、Prompt / Workflow / RAG / HTTP Tool、Agent、Gateway、评测、结果库及主壳证据均按切片通过双语验收；F 最终三窗口 76 次复验、519 项单测、原包体预算与最终仓库检查通过。服务和临时库已清理，分窗口证据见[本周周志](devlogs/2026-W41.md#2026-10-11-f-阶段全站双语收口)。
- [工作区成员邀请、认领与到期治理（开发 / 测试态）v1](features/admin-control-plane/workspace-member-invitation-claim-expiry-governance-dev-test-v1.md)已完成并关闭，状态为 `workspace_member_invitation_claim_expiry_governance_dev_test_v1_completed`。批次 E 已于 2026-09-08 获授权并完成 React、双数据库产品链、三视口、双标签与凭据隐私验收。
- 2026-09-09 已修复 SQLite / 模拟工作流主链阻塞、完成草案前端状态归属收敛与工作区响应式裁切修正。三条自动浏览器回归已连续两轮通过，当时接入 PR / Release 配置；浏览器验证现已改为本地按需运行，历史证据详见[2026-W37 周志](devlogs/2026-W37.md)。该证据不等于真实模型提供方、团队试用或生产验收。
- 2026-09-14 已完成[身份与成员领域首个切片](platform/engineering-health-productization-remediation-v1.md#身份与成员领域包边界与实现)：`internal/workspacepolicy` 已集中权限与四角色策略，调用点及既有覆盖率映射已迁移，身份、成员和邀请事务保留。独立策略、内存 / SQLite / HTTP、平台全量测试与 race / vet、覆盖率预算及仓库全量检查通过；PostgreSQL 四项隔离测试通过且环境已清理，详细结果见[2026-W38 周志](devlogs/2026-W38.md)。邀请[任务卡](task-cards/workspace-member-invitation-claim-expiry-governance-dev-test-v1-plan.md)不派生批次 F。
- [应用定时回归评测](features/user-workspace/application-evaluation-scheduled-regression-campaign-dev-test-v1.md)、工作区 Workflow 模板目录与 Action Safety Ladder 保持完成关闭，不派生新的同层批次。
- 工程整改 R2 至 R6 保持完成。前端状态收敛与三条工作流浏览器回归已分别获批实施，后者现作为本地按需验证，不再进入 PR `Candidate Quality`；不创建新工程批次。后端首个策略切片与双数据库回归已完成，进一步迁移须有具体维护问题，详见[工程健康专题的后续方向](platform/engineering-health-productization-remediation-v1.md#2026-09-06-复审与后续方向)。
- 项目所有者已暂缓真实模型试用。[故障诊断试用素材](features/user-workspace/prompt-application-dev-test-usage-guide.md#内部故障诊断试用)与已完成的 [模型提供方消息 / 输出衔接修复](features/user-workspace/prompt-application-template-version-review-controlled-invocation-dev-test-v1.md#真实-provider-的消息与输出适配)保留；恢复试用前再落实模型、费用和操作窗口。隔离测试不能作为真实模型提供方或人工答案质量验收。
- 2026-09-23 已完成 Prompt 结构化输出契约的网页编辑、精确版本审查与非空结果保存恢复；五条 Prompt 与原三条工作流连续两轮通过，测试服务和临时数据库已清理。实施与验收见[工程健康专题](platform/engineering-health-productization-remediation-v1.md#workflow-自动浏览器回归)及[2026-W39 周志](devlogs/2026-W39.md)，不新增同层任务卡或检查器。
- 本地控制台保持 `local usable / read-only close`，不再默认继续补同类只读 console 小切片；production secret backend、process supervisor、部署环境隔离和 console production packaging 仍为 `not_satisfied`。这些未满足项分别指生产密钥后端、进程监管、部署环境隔离和控制台生产打包；英文名称保留与现有检查器一致。

## 当前完成与下一动作（2026-10-15）

[Prompt 结构化输出契约的 UI 编辑](features/user-workspace/prompt-application-template-version-review-controlled-invocation-dev-test-v1.md#结构化输出契约的-ui-编辑)已完成。开发者可在网页配置五字段诊断契约，经草案保存 / 恢复、不可变版本与源码审查进入受控运行；非空结果显式保存和刷新恢复、缺字段 / 额外字段拒绝且不重调已通过隔离浏览器验收。三条工作流与五条 Prompt 连续两轮通过，服务与临时数据库已清理；验证记录见[2026-W39 周志](devlogs/2026-W39.md)。

项目所有者随后暂停手工测试，指定 i18n 为下一重点；本轮手工测试服务已停止，独立测试库与日志保留，未形成独立用户使用或答案质量验收。

项目所有者批准的首轮常读文档中文化已完成，范围与结果见[文档语言治理专题](document-language-governance-v1.md)。[中英 i18n 专题](features/web-internationalization-zh-en-v1.md)A、B 阶段已完成，保持原包体预算。草案编辑、冲突与修订恢复、候选审查 / 激活、Definition 运行、模板派生、RAG 与 HTTP Tool 已分窗口通过双语验收；离线投影解释于 2026-10-02 完成两语各两轮验收，服务与临时库已清理。C 阶段认证与账户安全已完成两语各两轮验收，会话冲突 / 撤销、密码轮换与双标签重新登录通过，服务与临时库已清理。工作区 / 成员 / 邀请与其余管理主页面亦已通过双语验收，C 阶段完成，D 阶段 Agent、Gateway 与运营亦已完成双语验收；不另起工程批次，E 阶段已登记切片与 F 全站覆盖、术语及帮助指南收口均已完成。真实模型试用继续暂缓，恢复前另行落实模型提供方 / 模型、费用与操作窗口。

2026-10-07 已完成 C 阶段工作区、成员与邀请双语迁移，身份组六条流程中英各两轮共 24 次通过。范围与证据见[本周周志](devlogs/2026-W41.md)；Provider、路由、配额、价格与租户 / 审计主页面随后完成双语迁移，管理组四条流程中英各两轮共 16 次通过，C 阶段关闭。D 阶段 Agent 的 Profile、运行分配与 Session 单次建议随后完成中英各两轮共 12 次验收，服务和临时库已清理。Gateway 三协议与 API 示例、请求历史 / 用量及运营随后完成 12 次双语验收，D 阶段关闭。E 阶段随后按下述切片验收，F 最终收口亦已完成。

2026-10-11 已完成 E 阶段覆盖清单及首个运行审查切片：运行历史、精确详情与对比两条浏览器流程中英各两轮共 8 次通过，覆盖语言切换、等待 / 失败恢复、迟到响应隔离与三视口内部布局。隔离服务及临时数据库已清理。Case / Suite 用例与套件随后通过中英各两轮共 8 次浏览器验收，包含真实冲突、等待切换、历史版本、人工决定和三视口，所属服务与临时库已清理；应用评测 Plan / Campaign / Schedule 三条流程中英各两轮共 12 次通过，计划 / 批次使用真实隔离 API，日程 / 发生记录限定为合成投影；结果库 / 导出、RAG 数据集与其余可达辅助证据完成双语迁移，四条流程中英各两轮共 16 次通过，真实隔离 API、导出与冲突恢复及三视口已验收；辅助页面限定读侧投影，所属服务和临时库已清理。E 阶段已登记切片完成，其后 F 补齐主壳 / 会话与术语、帮助指南，最终结果见本节结论。详见[本周周志](devlogs/2026-W41.md#2026-10-11-e-阶段结果库与可达证据双语迁移)。

[十三组浏览器回归](platform/engineering-health-productization-remediation-v1.md#workflow-自动浏览器回归)已完成本地中英 100/100 验收和入口遮挡修复。远端后续重跑在中文 RAG 晋级列表断言处失败；2026-10-15 已修复请求标识碰撞并通过四组本地复验，详见下节。PR / Release 浏览器任务及 `Candidate Quality` 对应依赖已移除，保留本地按需入口和其余六项 CI 检查。配置提交 `75bd9c34` 已正常推送，两套远端检查依次全部通过，PR 耗时 19 分 30 秒、Release 耗时 15 分 13 秒；当时本地仓库全量检查通过。真实模型试用继续暂停，决策与结果见[2026-W41 周志](devlogs/2026-W41.md#2026-10-11-浏览器验证退出-ci-门禁)。

## 接下来事项（2026-10-21）

[Prompt 调用与会话结构化变量输入](features/user-workspace/prompt-application-template-version-review-controlled-invocation-dev-test-v1.md#受控调用与会话的结构化变量输入2026-10-21)已完成设计、实现及独立获授权的隔离浏览器验收，新增表面已关闭；Prompt 子集、评测兼容及最终布局复验分别通过，环境已清理。真实模型试用继续暂停，详细结果见 [W43 周志](devlogs/2026-W43.md#2026-10-21-prompt-调用与会话结构化变量输入)。

1. [Prompt 计划结构化编辑与版本回归体验](features/user-workspace/application-evaluation-campaign-controlled-execution-dev-test-v1.md#隔离浏览器与-react-验收2026-10-21)已完成获授权的隔离验收：原评测流程 16/16、参考变更与作用域最终补测 4/4、样式修正后的错误 / 冲突复验 8/8，以及 Prompt 子集 36/36，均为中英各两轮的独立窗口。三视口与实际截图已复核，所属环境已清理；详细证据见[本周周志](devlogs/2026-W43.md#2026-10-21-prompt-计划编辑浏览器验收)。不派生同层续批，下一项按下节的实际用户阻塞或维护证据选择。
2. [PR #22](https://github.com/laugh0608/RadishMind/pull/22) 的阶段合入与 `master → dev` 回同步、Go 缓存与重复测试修复、[race 热点与夹具优化](platform/engineering-health-productization-remediation-v1.md#race-热点与测试夹具优化2026-10-20)保持完成。本地 `httpapi` race 由 414 降至 240 秒，原覆盖率与 race 标准保持；远端耗时留待后续正常 CI 观察，不重复启动远端工作流。
3. 真实模型试用继续暂停。恢复前须明确提供方 / 模型、费用、数据与运行窗口；未发现具体阻塞时不自动扩展后端拆包、定时执行或新功能批次。本轮浏览器验收不构成生产交付或真实答案质量证据。

## 下一顺位的选择

邀请产品链与项目所有者指定的中英 i18n 专题已完成。不自动派生同层续批；下一步仍围绕“内部开发者独立创建可复用 AI 应用、受控运行、审查结果与回归验证”按下表选择目标。四个产品面共同服务这条用户流程，不分别维持功能扩张队列。

| 顺序 | 候选工作 | 进入条件与结果 |
| --- | --- | --- |
| 1 | 修复实际用户流程中的阻塞 | 记录目标用户、任务、现有路径和失败证据；完成后能重复验证任务成功 |
| 2 | 按实际维护问题选择后端工作 | 首个权限 / 角色策略切片已完成；进一步迁移须有具体维护问题和行为验收依据，不自动连续拆包 |
| 3 | 持续验证关键浏览器流程 | 本地按需使用现有十三组入口，不进入 PR / Release 门禁；已登记的 Prompt、Workflow、模板、RAG、HTTP Tool、离线投影、身份、管理、Agent 及 Gateway / 运营流程按各自隔离窗口验收，运行环境与次数见工程专题和周志。新切片按实际风险选择回归，不能把分窗口通过合称一次全量或远端 CI 通过 |
| 4 | 内部真实任务试用与质量评测 | 明确参与者、模型 / 提供方、数据范围和运行窗口；记录完成率、耗时、求助点与输出质量 |
| 5 | 外部接入或生产交付 | 只有挂载点、负责人、资源和验收环境明确时进入独立专题 |

该表是决策顺序，不表示同时启动五项工作。仍按一条产品线、必要时一条工程线控制在制范围；实现前在既有功能或平台专题中确认具体范围。

## 保持关闭的边界

- 邀请不发送邮件、不搜索目录，不邀请 `workspace_admin`；邀请码不直接授予权限，认领仍由既有身份 / 成员 / 角色负责模块原子处理。
- 真实 Radish OIDC 联调保持 `real_radish_integration_deferred`。本地会话、开发身份头、签名测试成员断言和回环地址签发方不能作为生产授权证据。
- 生产密钥后端、生产认证、生产 API 密钥 / 配额 / 计费、真实模型提供方故障回退、生产工作进程、进程监管和生产交付继续未满足准入条件。
- 通用自治智能体循环、无限制工具、写方法工具、业务写回、自动确认提交、自动发布、重试 / 重放扩围继续关闭；Action Safety 不开放 `write_allowed_by_policy`。
- [应用运行观测与用量归因](features/user-workspace/application-operations-observability-usage-attribution-v1.md)后续仍为 `no_entry`：没有首分页窗口阻塞真实任务的证据时，不创建服务端汇总、跨存储库关联查询或计费账本。
- 普通 UI、文案、布局和只读证据整理复用既有测试与门禁；不恢复准入 / 审查 / 刷新历史尾链，不因专题关闭派生新任务卡或检查器。
- 未明确授权不启动常驻服务、Docker、真实模型长跑、模型或数据下载、生图后端、发布部署；不跨外部项目工作区写入。
- API key、token、DSN、一次性凭据、未经脱敏的真实敏感输入和模型提供方原始错误不得进入公开日志、URL 或已提交资产。

## 读取与验证

历史入口：[R6 文档与检查器收敛](platform/engineering-health-productization-remediation-v1.md#r6文档checker-与资产收敛)与[用户工作区应用目录与生命周期（开发/测试态）v1](features/user-workspace/application-catalog-lifecycle-dev-test-v1.md)均保持完成，不提供新的执行顺位。

1. 先读本文件与[功能设计入口](features/README.md)，再进入当前专题；仅在追溯时读取任务卡、周志和历史清单。
2. 当前技术边界见[架构](radishmind-architecture.md)，后续决策见[路线图](radishmind-roadmap.md)，执行规则见[Agent 协作与执行规则](agent-collaboration.md)。
3. 日常验证使用 `./scripts/check-repo.sh --fast`，Windows 使用 `pwsh ./scripts/check-repo.ps1 -Fast`；变更真相源、阶段或治理口径时补跑全量检查。
4. 虚拟环境缺失时按[项目指南](radishmind-project-guide.md)单独准备，依赖安装不作为每次文档验证的默认步骤。

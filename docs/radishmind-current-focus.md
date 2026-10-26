# RadishMind 当前推进焦点

更新时间：2026-10-26

## 文档职责

本文只回答当前目标、下一动作、阻塞和停止线。产品范围见[产品范围与目标](radishmind-product-scope.md)，实现事实见[能力矩阵](radishmind-capability-matrix.md)与[功能专题](features/README.md)，阶段选择见[路线图](radishmind-roadmap.md)，历史证据见[周志](devlogs/README.md)。

正文默认中文，代码标识符、路径、配置键和状态锚点保留原文。

## 当前结论（默认读取到本节结束）

- 项目所有者已接受全项目复审建议并批准正式规划整合。当前唯一产品主线改为[本机持久工作台交付 v1](features/user-workspace/local-persistent-workbench-delivery-v1.md)：通过正式 Web 创建 / 打开应用、受控运行、显式保存结果，关闭重启后继续使用。方向与规划已确认，具体启动 / 配置契约待设计，新交付尚未实现。
- 后续顺序是本机技术交付、真实任务效果验证，再根据实际需求选择共享交付或业务接入。跨版本回归审查调整为候选，尚未实现，不继续沿用先 Workflow 后 Prompt 的旧优先排期。
- 当前成熟度保持**内部开发者预览**。已有本地账户 / Session、SQLite、应用、Workflow、Gateway、评测和结果链；模块完成不等于统一入口、持续使用、真实模型质量或生产交付已经验收。
- 真实模型试用继续暂停。可先完成不依赖真实模型的设计、实现与技术验收；技术里程碑完成后如实记录质量待验证，不因等待模型窗口继续扩功能。
- 本轮只整合规划与文档，不修改运行代码，不沿用此前服务授权。审阅与决策依据见[工程健康专题](platform/engineering-health-productization-remediation-v1.md#2026-10-26-全项目复审交付后移与开发循环)与[W44 周志](devlogs/2026-W44.md)。

## 接下来事项（2026-10-26）

下一项具体工作是在[交付专题](features/user-workspace/local-persistent-workbench-delivery-v1.md#具体设计需要确定的内容)中确定本机启动、配置、身份、数据目录和恢复契约，并列明必要代码改动。复用既有 Go Platform、Python Runtime、正式 Web、本地 Session 与共享 SQLite；保留 PostgreSQL 已有能力，不重建平行实现。

设计应能直接进入实现：明确入口与关闭方式、配置来源及优先级、首次账户 / 工作区初始化、用户资源创建、Prompt / Workflow 同配置运行、结果保留和备份恢复。具体运行边界确认后，连续完成启动接线、Web 路径、恢复与验收，不再为每一层追加准备状态专题。

首轮以既有故障诊断 Prompt 素材检验结果保留，并在同一配置下验证 Workflow 版本与运行链。空环境到首次运行、关闭到继续使用、配置 / 存储 / 身份 / 提供方失败均进入验收；固定响应的技术通过与真实任务效果分别记录。

## 阶段顺位与进入条件

| 顺位 | 目标 | 完成或进入条件 |
| --- | --- | --- |
| 当前 | 本机持久工作台 | 统一正式 Web 入口与明确配置；正常使用、重启、备份恢复及失败路径有实际证据 |
| 后续 | 真实任务效果验证 | 恢复模型窗口，明确任务集、提供方 / 模型、数据与费用；记录质量和使用成本 |
| 按需求选择 | 团队共享或业务接入 | 实际用户 / 环境 / 负责人或上层挂载点明确；选一个目标交付，不同时铺开 |
| 候选 | 跨版本回归审查等增强 | 有持续使用证据，能说明现有手工路径的具体阻塞与收益 |

外部条件提前齐备时可以明确重排下一窗口；不自动并行开多条主线。真实故障与安全问题优先处理，具体条件见[路线图](radishmind-roadmap.md#后续开发方向2026-10-26)。

## 已完成范围与剩余边界

- 邀请治理、中英国际化、定时回归、Workflow 模板与 Action Safety 保持关闭；工程 R2 至 R6、身份策略首个切片与既有双数据库回归不派生续批。实现事实从[功能入口](features/README.md)追溯。
- 2026-10-15 RAG 晋级列表失败已修复并完成本地复验；2026-10-21 Prompt / Workflow 编辑与恢复、10 月 26 日首次使用指引及悬停修正均保持完成。分窗口证据见 [W42](devlogs/2026-W42.md)、[W43](devlogs/2026-W43.md)、[W44](devlogs/2026-W44.md)，不合称全站或真实用户验收。
- 本地控制台保持 `local usable / read-only close`，不再默认继续补同类只读 console 小切片；production secret backend、process supervisor、部署环境隔离和 console production packaging 仍为 `not_satisfied`。正式 Web 的本机交付不默认扩展旧 Console 或授予生产能力。
- 现有评测比较运行元数据；同计划版本批次配对不等于跨版本审查，也不判断答案质量。结果正文按 Session / Result Artifact 的既有显式保留规则处理。

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

历史索引：[R6 文档与检查器收敛](platform/engineering-health-productization-remediation-v1.md#r6文档checker-与资产收敛)、[用户工作区应用目录与生命周期（开发/测试态）v1](features/user-workspace/application-catalog-lifecycle-dev-test-v1.md)保持关闭，不提供执行顺位。

1. 先读本文件与当前交付专题；架构、身份、存储和使用指南只按任务需要读取。历史任务卡与周志用于追溯，不决定排期。
2. 日常验证使用 `./scripts/check-repo.sh --fast`，Windows 使用 `pwsh ./scripts/check-repo.ps1 -Fast`；变更真相源、阶段或治理口径时补跑全量检查。
3. 依赖准备、服务、模型与外部操作按[协作规则](agent-collaboration.md)落实范围；不因文档规划自动执行。

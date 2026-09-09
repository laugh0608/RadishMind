# RadishMind 当前推进焦点

更新时间：2026-09-09

## 文档职责

本文只回答当前目标、下一动作、阻塞和停止线。产品范围见[产品范围与目标](radishmind-product-scope.md)，实施事实见[能力矩阵](radishmind-capability-matrix.md)和对应[功能专题](features/README.md)，审阅数据与批次记录见[2026-W37 周志](devlogs/2026-W37.md)。

正文默认中文，代码标识符、路径、配置键和状态锚点保留原文。

## 当前结论（默认读取到本节结束）

- 当前成熟度为**内部开发者预览**。Workflow、应用运行与评测、Gateway、本地身份和开发测试态持久化已有连续实现；专题完成不等于真实 Provider、团队试用或生产运行已通过验收。
- [工作区成员邀请、认领与到期治理（开发 / 测试态）v1](features/admin-control-plane/workspace-member-invitation-claim-expiry-governance-dev-test-v1.md)已完成并关闭，状态为 `workspace_member_invitation_claim_expiry_governance_dev_test_v1_completed`。批次 E 已于 2026-09-08 获授权并完成 React、双数据库产品链、三视口、双标签与凭据隐私验收。
- 2026-09-09 已走查首条 SQLite / mock Workflow 用户链，修复编辑器迟到响应串入新草案、刷新后新建执行草案 ID 冲突、Definition v5 / v8 响应消费遗漏既有字段三处阻塞；已完成保存至精确运行历史的浏览器复验，详见[本周周志](devlogs/2026-W37.md)。该证据不等于真实 Provider、团队试用或新增自动浏览器回归。
- 下一步回到**真实用户流程阻塞与后续范围选择**。邀请[任务卡](task-cards/workspace-member-invitation-claim-expiry-governance-dev-test-v1-plan.md)不派生批次 F；前端草案库与 Designer 的状态归属收敛已于 2026-09-09 按批准范围完成并通过验收；后端领域拆分、自动浏览器回归与内部试用仍需分别选定具体目标和边界。
- [应用定时回归评测](features/user-workspace/application-evaluation-scheduled-regression-campaign-dev-test-v1.md)、工作区 Workflow 模板目录与 Action Safety Ladder 保持完成关闭，不派生新的同层批次。
- 工程整改 R2 至 R6 保持完成。当前文档已承接结构收敛、自动浏览器回归和内部使用验证的建议；它们是后续实施候选，尚无新工程批次或新增门禁。问题、依赖与验收见[工程健康专题的后续方向](platform/engineering-health-productization-remediation-v1.md#2026-09-06-复审与后续方向)。
- 本地 Console 保持 `local usable / read-only close`，不再默认继续补同类只读 console 小切片；production secret backend、process supervisor、部署环境隔离和 console production packaging 仍为 `not_satisfied`。

## 下一顺位的选择

邀请产品链已完成。围绕“内部开发者独立创建可复用 AI 应用、受控运行、审查结果与回归验证”选择下一目标。四个产品面共同服务这条用户流程，不分别维持功能扩张队列。

| 顺序 | 候选工作 | 进入条件与结果 |
| --- | --- | --- |
| 1 | 修复实际用户流程中的阻塞 | 记录目标用户、任务、现有路径和失败证据；完成后能重复验证任务成功 |
| 2 | 收敛前端工作流状态与一个后端领域 | 先明确受影响 owner、依赖与迁移边界；保护作用域、并发、持久化和敏感状态清理 |
| 3 | 固定少量自动浏览器回归 | 先选择已实现的关键流程与可重复测试环境；依赖、服务和 CI 变更按任务授权 |
| 4 | 内部真实任务试用与质量评测 | 明确参与者、模型 / Provider、数据范围和运行窗口；记录完成率、耗时、求助点与输出质量 |
| 5 | 外部接入或生产交付 | 只有挂载点、负责人、资源和验收环境明确时进入独立专题 |

该表是决策顺序，不表示同时启动五项工作。仍按一条产品线、必要时一条工程线控制在制范围；实现前在既有功能或平台专题中确认具体范围。

## 保持关闭的边界

- 邀请不发送邮件、不搜索目录，不邀请 `workspace_admin`；邀请码不直接授予权限，认领仍由既有身份 / 成员 / 角色 owner 原子处理。
- 真实 Radish OIDC 联调保持 `real_radish_integration_deferred`。本地 Session、dev header、signed-test membership 和 loopback issuer 不能作为 production 授权证据。
- production secret backend、生产认证、生产 API key / quota / billing、真实 Provider fallback、production worker、process supervisor 和生产交付继续未满足准入条件。
- 通用自治 agent loop、无限制工具、写方法 Tool、业务写回、自动确认提交、自动 release、retry / replay 扩围继续关闭；Action Safety 不开放 `write_allowed_by_policy`。
- [应用运行观测与用量归因](features/user-workspace/application-operations-observability-usage-attribution-v1.md)后续仍为 `no_entry`：没有首分页窗口阻塞真实任务的证据时，不创建服务端 summary、跨 store join 或 billing ledger。
- 普通 UI、文案、布局和只读证据整理复用既有测试与门禁；不恢复 readiness / review / refresh 历史尾链，不因专题关闭派生新任务卡或 checker。
- 未明确授权不启动常驻服务、Docker、真实模型长跑、模型或数据下载、生图后端、发布部署；不跨外部项目工作区写入。
- API key、token、DSN、一次性凭据、未经脱敏的真实敏感输入和 Provider 原始错误不得进入公开日志、URL 或 committed 资产。

## 读取与验证

历史入口：[R6 文档与检查器收敛](platform/engineering-health-productization-remediation-v1.md#r6文档checker-与资产收敛)与[用户工作区应用目录与生命周期（开发/测试态）v1](features/user-workspace/application-catalog-lifecycle-dev-test-v1.md)均保持完成，不提供新的执行顺位。

1. 先读本文件与[功能设计入口](features/README.md)，再进入当前专题；仅在追溯时读取任务卡、周志和历史清单。
2. 当前技术边界见[架构](radishmind-architecture.md)，后续决策见[路线图](radishmind-roadmap.md)，执行规则见[Agent 协作与执行规则](agent-collaboration.md)。
3. 日常验证使用 `./scripts/check-repo.sh --fast`，Windows 使用 `pwsh ./scripts/check-repo.ps1 -Fast`；变更真相源、阶段或治理口径时补跑全量检查。
4. 虚拟环境缺失时按[项目指南](radishmind-project-guide.md)单独准备，依赖安装不作为每次文档验证的默认步骤。

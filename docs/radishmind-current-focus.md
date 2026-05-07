# RadishMind 当前推进焦点

更新时间：2026-05-07

## 文档目的

本文档用于回答“根据项目规划和开发进度，今天要做什么以推进开发”。它是新会话的短入口，不承载历史细节、完整实验日志或长契约说明。

默认只读本文件仍应能判断下一步方向；只有准备实施具体改动时，才跳转到下方按需文档。

## 当前阶段

当前主线集中在 `M3/M4`：

- `M3`：维护现有 gateway、service smoke、UI consumption 与 candidate handoff，作为未来上层接入门禁。
- `M4`：继续验证 `RadishMind-Core` 的结构化输出路线，重点是 task-scoped response builder / tooling 分工、broader review runbook、natural-language audit 和 human review records。

当前不启动训练放量，不默认扩同类真实 capture，不把 builder / repaired / injected 轨通过解释成 raw 模型能力晋级。

## 今天优先做什么

当前 broader task-scoped builder review 的两段本地执行和 15 条样本人工复核都已经完成。当前结论是该批 records 收口为 `reviewed_changes_required`：5 条样本可接受，10 条样本仍卡在 task-grounded natural-language、citation explanation、evidence-gap wording 与 ambiguity/no-tab ghost explanation；不要把这次 builder 结果写成 raw 晋级或训练准入。

1. 围绕 10 条 `reviewed_changes_required` broader review 样本收口修复面，重点治理 task-grounded natural-language、citation explanation、evidence-gap wording 与 ambiguity/no-tab ghost explanation。
2. 把修复后的 blocked 样本按同一 runbook 重跑，并回填 broader review records，确认它们是否能推进到 `reviewed_pass`。
3. 保持 broader review records 为 `reviewed_changes_required`，在 blocked 样本修复前不写 `reviewed_pass`。
4. 继续维护 service/API smoke 矩阵，不新增散落 UI / 命令层模拟 summary。

## 为什么是这个任务

- 当前 raw 小模型仍 blocked，后处理和 builder 轨只能作为 tooling 分工证据。
- full-holdout-9 与 holdout6-v2-non-overlap 两段本地执行都已完成，machine gate / offline eval / natural-language audit 均通过；15 条样本人工复核也已完成，但 broader review records 当前结论是 `reviewed_changes_required`，不能把 builder 结果写成 raw 晋级、训练准入或 production contract 接受证据。
- broader review 的可执行样本面、执行清单和 review records 已经接入仓库级验证；当前重点是围绕 10 条 blocked 样本收口修复，而不是继续设计新的入口。
- 在 broader review blocked 样本修复并复核通过前，仍不要把 builder 结果写成 raw 晋级、训练准入或 production contract 接受证据。

## 默认不要做

- 不继续加长同一批 prompt/scaffold 当作默认推进。
- 不扩 `RadishFlow` 同类真实 capture，除非先写清楚非重复 drift 假设。
- 不把 `RadishCatalyst` 从文档预留提前扩成真实 schema、adapter 或 gateway smoke。
- 不下载模型、数据集或权重。
- 不把真实模型输出、训练 JSONL 或大体积实验产物提交入仓。
- 不直接修改 `RadishFlow`、`Radish` 或 `RadishCatalyst` 外部工作区。

## 最小读取路径

回答“今天做什么”时，默认读取：

1. `AGENTS.md` 或 `CLAUDE.md`
2. `docs/README.md`
3. `docs/radishmind-current-focus.md`
4. 必要时读取 `docs/radishmind-roadmap.md`

## 按需读取

- 动产品边界：读 `docs/radishmind-product-scope.md`
- 动架构或服务分层：读 `docs/radishmind-architecture.md`
- 动协议、schema 或 API：读 `docs/radishmind-integration-contracts.md`
- 动 `RadishMind-Core` 评测：读 `docs/radishmind-core-baseline-evaluation.md`
- 动代码风格、抽象或脚本组织：读 `docs/radishmind-code-standards.md`
- 查最近执行细节：读本周周志 `docs/devlogs/2026-W19.md`

## 验证基线

文档或治理改动完成后，优先执行：

```bash
./scripts/check-repo.sh
```

Windows / PowerShell 环境使用：

```powershell
pwsh ./scripts/check-repo.ps1
```

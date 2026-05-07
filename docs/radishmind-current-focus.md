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

## 当前优先做什么

当前 broader task-scoped builder review 的两段本地执行和 15 条样本人工复核都已经完成。今天已把 10 条 blocked 样本对应的 deterministic builder 收口和回归断言补齐，但 broader review records 仍停留在 `reviewed_changes_required`：5 条样本可接受，10 条样本仍待用新一轮 `tmp/` 产物复核；不要把这次 builder 结果写成 raw 晋级或训练准入。

1. `2026-05-08` 先按原 runbook 重跑 `full-holdout-9` 的 `--build-task-scoped-response` broader review 段。
2. 再重跑 `holdout6-v2-non-overlap` 同一轨段，并继续保持 `--sample-timeout-seconds 300`。
3. 读取两段新的 `candidate summary / offline eval run / natural-language audit`，重点复核 `tmp/radishmind-core-broader-review-qwen15b-task-scoped-builder-full-holdout-timeout300/` 与 `tmp/radishmind-core-broader-review-qwen15b-task-scoped-builder-v2-timeout300/` 下产物。
4. 只有在 blocked 样本的新产物通过人工复核后，才更新 broader review records；在此之前继续保持 `reviewed_changes_required`。
5. 继续维护 service/API smoke 矩阵，不新增散落 UI / 命令层模拟 summary。

## 为什么是这个任务

- 当前 raw 小模型仍 blocked，后处理和 builder 轨只能作为 tooling 分工证据。
- 今天已经补齐 10 条 `reviewed_changes_required` 样本的 deterministic builder 收口与回归断言，但 broader review records 仍对应重跑前的真实 `tmp/` 产物，不能直接手改成 `reviewed_pass`。
- full-holdout-9 与 holdout6-v2-non-overlap 两段本地执行都已完成，machine gate / offline eval / natural-language audit 均通过；15 条样本人工复核也已完成，但当前下一步必须先重跑两段 broader review，而不是继续设计新的入口。
- 在 broader review blocked 样本修复并复核通过前，仍不要把 builder 结果写成 raw 晋级、训练准入或 production contract 接受证据。

## 2026-05-08 先看这些产物

- `tmp/radishmind-core-broader-review-qwen15b-task-scoped-builder-full-holdout-timeout300/summary.json`
- `tmp/radishmind-core-broader-review-qwen15b-task-scoped-builder-full-holdout-timeout300-run.json`
- `tmp/radishmind-core-broader-review-qwen15b-task-scoped-builder-full-holdout-timeout300-natural-language-audit.json`
- `tmp/radishmind-core-broader-review-qwen15b-task-scoped-builder-v2-timeout300/summary.json`
- `tmp/radishmind-core-broader-review-qwen15b-task-scoped-builder-v2-timeout300-run.json`
- `tmp/radishmind-core-broader-review-qwen15b-task-scoped-builder-v2-timeout300-natural-language-audit.json`
- `training/datasets/radishmind-core-task-scoped-builder-broader-review-records-v0.json`

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

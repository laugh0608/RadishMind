# RadishMind 当前推进焦点

更新时间：2026-05-10

## 文档目的

本文档用于回答“根据项目规划和开发进度，今天要做什么以推进开发”。它是新会话的短入口，不承载历史细节、完整实验日志或长契约说明。

默认只读本文件仍应能判断下一步方向；只有准备实施具体改动时，才跳转到下方按需文档。

## 当前阶段

当前主线集中在 `M3/M4`：

- `M3`：维护现有 gateway、service smoke、UI consumption 与 candidate handoff，作为未来上层接入门禁。
- `M4`：继续验证 `RadishMind-Core` 的结构化输出路线，重点是 task-scoped response builder / tooling 分工、broader review runbook、natural-language audit 和 human review records。

当前不启动训练放量，不默认扩同类真实 capture，不把 builder / repaired / injected 轨通过解释成 raw 模型能力晋级。

## 当前优先做什么

当前 broader task-scoped builder review 的两段本地执行、6 份 `tmp/` 产物读取和 15 条样本人工复核都已经完成，records 现已更新为 15/15 `reviewed_pass`。这说明 task-scoped response builder / tooling 分工在 broader 15 样本面上已经有稳定 tooling-route evidence；但仍不要把这次 builder 结果写成 raw 晋级或训练准入。

1. 继续维护 `M3` 的 service/API smoke 矩阵，不新增散落 UI / 命令层模拟 summary。
2. 把 broader 15 样本 `reviewed_pass` 结果作为当前 builder/tooling 路线的正式人工复核依据，而不是继续补同一批 blocked 样本。
3. 在不把 builder 结果写成 raw 晋级或训练准入的前提下，继续保留 constrained/guided decoding 为已验证的同边界改善信号，并把 `holdout6-v2-non-overlap` + `300s` 的对照结论写实到实验记录。
4. 当前 `.venv` 的 `transformers 5.7.0` 已通过 `custom_generate` scaffold-slot shim 接入 guided-decoding runtime，不再要求本机先暴露 `GenerationConfig.guided_decoding` 才能开始 guided 轨。
5. 2026-05-09 的 `Qwen2.5-1.5B-Instruct` guided smoke / `holdout6-v2-non-overlap` 已完成，candidate summary 与 offline eval 均为 6/6 通过；但 candidate responses 里仍有自然语言退化、重复和 `max_new_tokens` 打满样本，因此该结果只说明结构化约束有效，不等于路线已经收口。
6. 2026-05-10 的 `Qwen3-4B-Instruct-2507` raw / guided 也已完成：raw 在同一 holdout 上仍被两条复杂样本卡住；guided 虽然 6/6 机器通过、`timeout_count=0`，并在较简单的 docs/ghost 样本上减少了 `3B guided` 的重复 filler，但 6 条 inspected `summary` 都泄漏了 `','answers':[...` 形态的 JSON 片段，且 cross-object / no-tab 语义没有真正收口。
7. 2026-05-10 的 `Qwen2.5-3B-Instruct` raw / guided 也已完成：raw 在 `suggest_flowsheet_edits` 上仍 blocked，且保留 1 条超时与 `advisory_action_boundary_rate=0.5`；guided 虽然把结构化指标拉回到 6/6 机器通过，但仍保留跨对象与 ghost/docs 的自然语言退化。当前优先级仍是 `3B/4B` 对照，但现在已经有 3B 和 4B 两端证据，而且当前结论是“4B guided 只有局部去噪改善，并未形成质量晋级”，所以下一步应把它们收口成容量-质量曲线，而不是继续扩同一 1.5B guided 样本面。

## 为什么是这个任务

- 当前 raw 小模型仍 blocked，后处理和 builder 轨只能作为 tooling 分工证据。
- broader 15 样本现在已经完成 machine gate、offline eval、natural-language audit 和人工复核，且当前 records 为 15/15 `reviewed_pass`；这意味着继续停留在“重跑同一批 broader review”已不再是最高价值动作。
- 当前更高价值的是利用这批 broader `reviewed_pass` 结果和 3B/4B guided 机器通过结果，进入更清晰的模型容量对照；其中 `4B guided` 已证明“简单去重改善”和“系统性 summary 泄漏”可以同时存在，因此不能只看 machine pass 就把它写成质量晋级。
- 即便 broader review 已通过，也仍不要把 builder 结果写成 raw 晋级、训练准入或 production contract 接受证据。

## 2026-05-08 已确认这些产物

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
- 不默认下载大于当前决策所需范围的模型、数据集或权重。
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

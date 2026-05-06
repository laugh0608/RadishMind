# RadishMind 当前推进焦点

更新时间：2026-05-06

## 文档目的

本文档用于回答“根据项目规划和开发进度，今天要做什么以推进开发”。它是新会话的短入口，不承载历史细节、完整实验日志或长契约说明。

默认只读本文件仍应能判断下一步方向；只有准备实施具体改动时，才跳转到下方按需文档。

## 当前阶段

当前主线集中在 `M3/M4`：

- `M3`：维护现有 gateway、service smoke、UI consumption 与 candidate handoff，作为未来上层接入门禁。
- `M4`：继续验证 `RadishMind-Core` 的结构化输出路线，重点是 task-scoped response builder / tooling 分工、citation tightened rerun、natural-language audit 和 human review records。

当前不启动训练放量，不默认扩同类真实 capture，不把 builder / repaired / injected 轨通过解释成 raw 模型能力晋级。

## 今天优先做什么

优先任务是推进 `M4` 的 citation tightened full-holdout 闭环：

1. 等用户完成 citation tightened full-holdout-9 task-scoped builder 本地重跑。
2. 读取 `tmp/` 下新的 candidate summary、offline eval、natural-language audit 和相关产物。
3. 更新 review records，判断 `compressor-parameter-update` 是否可以从 `reviewed_changes_required` 改为 `reviewed_pass`。
4. 若仍未通过，先定位 citation、answer、issue、action rationale 或 fixture/scaffold 的根因。
5. 若全部通过，再评估是否扩大 task-scoped builder 样本面，或进入 constrained/guided decoding、`minimind-v`、`3B/4B` 对照。

## 为什么是这个任务

- 当前 raw 小模型仍 blocked，后处理和 builder 轨只能作为 tooling 分工证据。
- full-holdout-9 的机器门禁和 deterministic audit 曾通过，但人工复核发现 broad citation 仍未满足接受标准。
- citation fixture/scaffold 已收紧；下一步必须用新的本地重跑产物更新结论，不能追认旧产物。
- 在该 review 缺口关闭前，扩大样本、启动训练或切换更大模型都会提前消耗复杂度。

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

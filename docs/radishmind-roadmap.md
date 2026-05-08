# RadishMind 阶段路线图

更新时间：2026-05-08

## 路线图原则

路线图只记录阶段目标、当前进度、下一步和停止线。批次细节、历史失败、完整实验输出和长命令不放在本入口文档中，应进入周志、实验 manifest、run record 或任务卡。

当前长期目标保持不变：`RadishMind` 是受控 Copilot / Agent 系统 + 可替换模型能力，不是单一万能模型。

## 当前主线

近期实际推进集中在 `M3/M4`：

- `M3`：把已有 gateway、service smoke、UI consumption 与 candidate handoff 保持为服务/API 接入验收门禁。
- `M4`：继续验证 `RadishMind-Core` 的结构化输出路线，重点是 task-scoped builder、natural-language audit、human review records 和 broader review 执行；broader 15 样本两段本地执行与人工复核现已完成并更新为 15/15 `reviewed_pass`，下一步应利用这批 broader 通过结果来稳定 builder/tooling 路线，而不是继续重跑同一批样本。
- 当前不启动训练放量，不继续默认扩同类真实 capture，不把 builder 轨通过解释成 raw 模型能力晋级。

## 阶段

### M0：真实上下文审查与规划冻结

目标：建立文档真相源，收口产品定位、边界和任务矩阵。

状态：已完成基础定位和跨项目边界；后续只在边界变化时更新。

### M1：统一协议与上下文打包

目标：建立 `CopilotRequest` / `CopilotResponse`、artifact、citation、风险分级和 `requires_confirmation`。

状态：核心 schema 和 `RadishFlow export -> adapter -> request` 链路已具备仓库级回归；继续保持短路径、manifest 元数据和 schema 校验口径。

### M2：`RadishFlow` 首个 PoC

目标：证明结构化 flowsheet 状态、诊断摘要、控制面状态和 ghost completion 可以形成稳定 Copilot 任务。

状态：`suggest_flowsheet_edits` 与 `suggest_ghost_completion` 已具备 committed eval、真实 capture、audit、replay 和治理资产；当前不再默认扩同类真实样本。

### M3：服务/API 收口

目标：把 runtime、gateway、UI consumption 和 candidate handoff 统一成可复跑服务切片。

状态：进程内 Python gateway、service smoke matrix、RadishFlow gateway demo、UI consumption 与 candidate edit handoff 已作为未来上层接入门禁；HTTP server 暂不作为当前默认推进项。

### M4：模型路线与评测闭环

目标：明确 `RadishMind-Core` 的基座适配、结构化输出、response builder / tooling 分工和评测晋级标准。

状态：本地小模型 raw 仍 blocked；repair、hard-field injection 与 task-scoped builder 已提供路线信号，但不能替代 raw 晋级。citation tightened full-holdout-9 已完成并通过 review；broader task-scoped builder 的 15 样本 review entry、两段本地执行 runbook 和 review records 已接入仓库级验证，且 broader 15 样本现已完成 machine gate、offline eval、natural-language audit 和人工复核，records 当前为 15/15 `reviewed_pass`。当前下一步不再是重跑同一批 broader review，而是先在同一 `holdout6-v2-non-overlap` + `300s` 边界上推进 constrained/guided decoding，再根据该轨结果决定是否进入更大样本面或 `minimind-v` / `3B/4B` 对照。

### M5：`Radish` 首批任务接入

目标：在共享协议和评测基线上扩展 `Radish` 文档问答、Console/运营辅助、内容结构化和附件解释。

状态：docs QA 与训练样本转换已有基础资产；真实上层接入仍等待，不提前扩自动治理或权限写回。

### M6：双项目工程化收口

目标：在 `RadishFlow` 和 `Radish` 都有可用场景后，再收口多项目路由、审计、缓存、部署和 provider 策略。

状态：暂不提前压入当前阶段。

### M7：`RadishCatalyst` 预留评估

目标：未来在首个真实任务明确后，评估玩家知识问答、进度解释、生产链规划或开发侧静态数据一致性检查。

状态：当前只保留文档级口径，不扩 schema、adapter、eval 或 gateway smoke。

## 下一步

1. 继续维护服务/API smoke 矩阵，不新增散落 UI / 命令层模拟 summary。
2. 把 broader 15 样本 `reviewed_pass` 结果作为当前 builder/tooling 路线的正式人工复核依据。
3. 在不把 builder 结果写成 raw 晋级或训练准入的前提下，先补 constrained/guided decoding 的执行前置与 runbook，并在 `holdout6-v2-non-overlap` 上完成同边界对照。
4. 仅在 guided/constrained 轨仍不能明显改善 raw 后，再决定是否进入更大样本面或 `minimind-v` / `3B/4B` 对照。
5. 训练数据继续只提交治理 manifest、summary、复核记录和实验说明；JSONL 默认输出到 `tmp/`。
6. 图片生成继续沿 intent、backend request、artifact metadata 和 safety gate 推进，不下载模型、不生成图片。

## 停止线

- 不把 repaired、injected 或 builder 轨通过写成 raw 模型能力通过。
- 不把机器指标通过写成人工可接受度通过。
- 不在没有非重复 drift 假设时继续扩真实 capture。
- 不让模型直接写上层业务真相源。
- 不用晦涩抽象、空泛 helper 或多层 fallback 掩盖代码职责不清。

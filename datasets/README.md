# RadishMind 数据集与评测目录

更新时间：2026-04-01

本目录用于承载 `RadishMind` 的样本、评测和后续训练输入。

当前阶段先建立最小评测骨架，不急着铺大规模数据。

推荐子目录职责：

- `synthetic/`: 合成样本
- `annotated/`: 人工校正样本
- `examples/`: 面向契约与适配器的最小示例对象
- `eval/`: 离线评测样本、样本 schema 和回归输入

当前已经先落地：

- `examples/radishflow-ghost-candidate-set-flash-basic-001.json`
- `examples/radishflow-copilot-request-ghost-flash-basic-001.json`
- `eval/radishflow-task-sample.schema.json`
- `eval/radish-task-sample.schema.json`
- `eval/radishflow/*.json`
- `eval/radish/*.json`
- `eval/README.md`

当前原则：

- 优先让样本直接服务于任务卡和契约校验
- 优先积累 `RadishFlow` 的状态优先样本
- 在 `Radish` 首个场景收口前，不急着平铺多任务数据
- 当前 `RadishFlow explain_diagnostics` 已补首批 `golden_response` 样本与最小回归 runner，后续优先扩展更多真实诊断场景
- 当前 `Radish` 文档问答样本已补最小召回输入约束与回归 runner，后续优先扩展真实样本和候选输出对照入口
- 当前 `examples/` 目录开始承接“schema + 实例”双校验，避免新契约只有结构没有真实对象参照
- 当前 `build-radishflow-ghost-request.py` 已可把本地 ghost 候选集示例装配成最小 `CopilotRequest`，默认使用 `model-minimal` profile 裁剪本地排序/冲突证据，同时保留 `debug-full` 对照示例，并由 `check-repo` 校验 `FlashDrum` 与 `Valve ambiguous` 两组输出都不漂移

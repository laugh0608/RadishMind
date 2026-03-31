# RadishMind 数据集与评测目录

更新时间：2026-03-31

本目录用于承载 `RadishMind` 的样本、评测和后续训练输入。

当前阶段先建立最小评测骨架，不急着铺大规模数据。

推荐子目录职责：

- `synthetic/`: 合成样本
- `annotated/`: 人工校正样本
- `eval/`: 离线评测样本、样本 schema 和回归输入

当前已经先落地：

- `eval/radishflow-task-sample.schema.json`
- `eval/radish-task-sample.schema.json`
- `eval/radish/*.json`
- `eval/README.md`

当前原则：

- 优先让样本直接服务于任务卡和契约校验
- 优先积累 `RadishFlow` 的状态优先样本
- 在 `Radish` 首个场景收口前，不急着平铺多任务数据
- 当前 `Radish` 文档问答样本已补最小召回输入约束与回归 runner，后续优先扩展真实样本和候选输出对照入口

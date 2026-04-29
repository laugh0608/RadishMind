# training/datasets 说明

更新时间：2026-04-29

本目录预留给训练 / 蒸馏数据集的轻量治理文件。

当前只允许提交：

- 小型训练集合 manifest
- 转换 summary
- 抽样复核记录
- 样本来源索引
- 与离线评测接线有关的说明文件

默认不提交：

- 大规模 JSONL
- 模型权重
- tokenizer、checkpoint、adapter 二进制产物
- provider 原始输出临时 dump
- 本地探测缓存

首批 JSONL 应继续由 `scripts/build-copilot-training-samples.py` 从 committed eval 样本或 audit pass candidate record 生成，并默认输出到 `tmp/`。

# RadishMind 项目总览与使用指南

更新时间：2026-05-10

## 这份文档讲什么

这是一份面向新读者的项目说明书，回答三个问题：

- `RadishMind` 是什么
- 仓库怎么分工
- 常见入口和参数怎么用

它不替代 `docs/radishmind-current-focus.md`、`docs/devlogs/` 或任务卡，也不记录阶段推进流水。

## 项目定位

`RadishMind` 是 `Radish` 体系下的外部智能层仓库，不是上层业务真相源，也不替代 `RadishFlow`、`Radish` 或 `RadishCatalyst` 的业务决策权。

当前主要职责是：

- 接收结构化输入
- 输出解释、诊断、结构化建议和候选动作
- 维护统一协议、评测门禁和训练治理
- 通过可替换模型路线支持后续接入

## 架构速览

当前架构固定为六层：

1. `Client Adapters & Context Packers`：把上层状态整理成统一请求
2. `Copilot Gateway / Task Router`：校验请求、识别任务、选择 provider/profile
3. `Retrieval & Tool Layer`：检索文档、解析附件、整理本地证据
4. `Model Runtime Layer`：生成解释、候选意图和结构化输出
5. `Rule Validation & Response Builder`：收口风险、确认边界和响应结构
6. `Data / Evaluation / Training Pipeline`：管理样本、评测、复核和训练治理

细节说明见 [系统架构](radishmind-architecture.md)。

## 目录速览

- `docs/`：正式文档源
- `contracts/`：JSON Schema 真相源
- `scripts/`：检查、转换、评测和本地运行入口
- `datasets/`：eval 样本、示例对象和 candidate record
- `training/`：训练治理、实验说明和复核记录
- `services/`：运行时实现
- `adapters/`：上游项目适配层
- `tmp/`：本地临时产物，不提交

## 常见入口

### 仓库级检查

日常先跑：

```bash
./scripts/check-repo.sh --fast
```

需要全量门禁时再跑：

```bash
./scripts/check-repo.sh
```

Windows / PowerShell 对应入口是 `pwsh ./scripts/check-repo.ps1` 和 `pwsh ./scripts/check-repo.ps1 -Fast`。

### 协议和样本

- 看结构定义：`contracts/README.md`
- 看任务边界：`docs/task-cards/README.md`
- 看示例对象：`datasets/examples/`

### 本地候选输出

本地小模型候选输出通常先走这个入口：

```bash
python3 scripts/run-radishmind-core-candidate.py \
  --provider local_transformers \
  --model-dir /path/to/model \
  --output-dir tmp/radishmind-core-candidate-local \
  --allow-invalid-output \
  --validate-task
```

常见参数含义：

- `--provider local_transformers`：本地模型提供者
- `--model-dir`：本地模型目录
- `--output-dir`：本次输出目录
- `--allow-invalid-output`：保留无效输出用于审计
- `--validate-task`：开启任务级校验
- `--sample-id`：只跑单条样本
- `--sample-timeout-seconds`：限制单样本超时

其它实验轨按需叠加，例如 `--repair-hard-fields`、`--inject-hard-fields`、`--build-task-scoped-response`、`--guided-decoding json_schema`。这些开关按实验轨选择，通常不要混用。

### 模型下载

如果 `hf download` 在代理环境里卡在 `transfer.xethub.hf.co`，优先用这个：

```bash
HF_HUB_DISABLE_XET=1 hf download Qwen/Qwen2.5-3B-Instruct \
  --local-dir /home/luobo/Code/Models/Qwen2.5-3B-Instruct \
  --max-workers 1
```

### 训练样本转换

把 committed eval 样本或 audit pass candidate record 转成训练样本时，优先用：

```bash
python3 scripts/build-copilot-training-samples.py \
  --manifest scripts/checks/fixtures/copilot-training-sample-conversion-manifest.json \
  --output-jsonl tmp/copilot-training-samples-golden.jsonl
```

生成结果默认留在 `tmp/`，不要把大规模 JSONL 直接入仓。

### 离线评测

离线评测的基本顺序是：

1. 先生成 candidate response
2. 再交给 `scripts/run-radishmind-core-offline-eval.py`
3. 最后看 `tmp/` 下的 summary 和 run record

## 读文档顺序

如果你刚接触这个仓库，建议按这个顺序读：

1. [文档入口](README.md)
2. 项目总览与使用指南
3. [当前推进焦点](radishmind-current-focus.md)
4. [产品范围](radishmind-product-scope.md)
5. [系统架构](radishmind-architecture.md)
6. [跨项目集成契约](radishmind-integration-contracts.md)
7. [脚本目录说明](../scripts/README.md)
8. [数据集目录说明](../datasets/README.md)
9. [训练目录说明](../training/README.md)

## 不做什么

- 不把 `RadishMind` 做成上层业务真相源
- 不默认下载大模型、数据集或权重
- 不把 `tmp/` 里的临时产物当成 committed 资产
- 不用 builder 或 repaired 结果冒充 raw 晋级证据

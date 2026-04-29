# scripts/ 目录说明

更新时间：2026-04-29

## 目录目标

`scripts/` 用于承载仓库级检查、评测回归、数据构建与最小运维入口。

当前约定不再继续把所有实现直接平铺在 `scripts/` 根目录，而是采用“根目录稳定入口 + 浅层分类子目录”的方式收口。

## 当前分组

- `scripts/`
  - 保留稳定入口脚本，以及 `ps1` / `sh` 平台包装
  - 例如 `check-repo.py`、`run-eval-regression.py`、`check-repo.sh`
  - 当前还提供 `build-copilot-training-samples.py`，用于把 committed eval 样本中的 `input_request + golden_response` 转换为 `CopilotTrainingSample` JSONL，也支持把 audit pass candidate record 转换为 `teacher_capture` 训练样本，并用 summary fixture 固定首批转换结果；生成 JSONL 默认输出到 `tmp/`，不直接作为 committed 训练资产
- `scripts/checks/`
  - 放仓库检查相关的内部模块与静态 fixture
  - 当前已用于承载 `check-repo` 的 fixture JSON，以及 `radish docs QA real batch summary` 的内部 helper
- `scripts/eval/`
  - 放评测回归 runner 的内部实现模块
  - 当前 `run-eval-regression.py` 的具体实现已拆到这里
  - 当前也承载 `report_real_batch_governance_status.py` 这类只读治理报表，用于统一盘点 `suggest_flowsheet_edits`、`suggest_ghost_completion` 与 `Radish docs QA` 的 formal real batch、coverage、replay / real-derived 连通性，以及当前优先级队列

## 维护约定

- 单个 `Python` 脚本和单个 committed `JSON` 文件默认不超过 `1000` 行
- 当根目录入口脚本变长时，优先把实现迁入对应分类子目录，根目录只保留薄 wrapper
- 只有“用户会直接执行”的稳定命令才应长期留在根目录
- 内部 helper、共享实现和大型 fixture 不应继续堆在根目录
- 目录层级保持浅，优先两层；非必要不要继续嵌套更深
- 训练 / 蒸馏相关脚本应优先输出 summary 与 manifest；大规模 JSONL、权重、checkpoint 和本地 provider dump 默认放在 `tmp/` 或外部工作区，不提交到仓库

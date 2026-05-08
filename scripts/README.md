# scripts/ 目录说明

更新时间：2026-05-07

## 目录目标

`scripts/` 用于承载仓库级检查、评测回归、数据构建与最小运维入口。

当前约定不再继续把所有实现直接平铺在 `scripts/` 根目录，而是采用“根目录稳定入口 + 浅层分类子目录”的方式收口。

## 当前分组

- `scripts/`
  - 保留稳定入口脚本，以及 `ps1` / `sh` 平台包装
  - 例如 `check-repo.py`、`run-eval-regression.py`、`check-repo.sh`、`check-repo-fast.sh`
  - `check-repo.py` 支持 `--fast`，用于日常快速验证；`check-repo.sh --fast`、`check-repo-fast.sh`、`pwsh ./scripts/check-repo.ps1 -Fast` 与 `pwsh ./scripts/check-repo-fast.ps1` 都会跳过慢速回归和批量元数据重跑，但仍保留核心静态门禁
  - 当前还提供 `build-copilot-training-samples.py`，用于把 committed eval 样本中的 `input_request + golden_response` 转换为 `CopilotTrainingSample` JSONL，也支持把 audit pass candidate record 转换为 `teacher_capture` 训练样本，并用 summary fixture 固定首批转换结果；生成 JSONL 默认输出到 `tmp/`，不直接作为 committed 训练资产
  - 当前还提供 `run-radishmind-core-offline-eval.py`，用于运行 `RadishMind-Core` 首个离线评测 runner；当前既支持读取 committed eval 样本的 `golden_response` 作为 fixture 候选输出，也支持读取 `run-radishmind-core-candidate.py` 生成的 candidate response 文件和 summary，生成符合 `radishmind-core-offline-eval-run` schema 的 completed run record；评测阶段不重新运行模型、不访问 provider、不下载权重，schema-invalid raw 输出会进入指标失败统计而不是阻断整批记录生成
  - 当前还提供 `run-radishmind-core-candidate.py`，用于把离线评测样本转换成候选模型 prompt 与 candidate response 文件；仓库级检查只使用 `golden_fixture` dry-run provider，真实本地模型必须显式传入 `--provider local_transformers` 和 `--model-dir` 或 `RADISHMIND_MODEL_DIR`，且脚本使用 `local_files_only`，不自动下载模型；调试真实小模型时可用 `--sample-id` 限定单条样本，用 `--sample-timeout-seconds` 为单条本地生成设置超时边界，并用 `--allow-invalid-output` 把 schema-invalid、JSON parse error 或 timeout 原始失败保留在 `tmp/` 下审计；本地模型运行会打印 `runtime-ready / batch-start / sample-start / sample-done` 进度，便于区分正在生成、单样本超时和脚本无响应；当前 `Qwen2.5-1.5B-Instruct` 本地 9 条 M4 fixture raw / repaired 全量复跑推荐使用 `--sample-timeout-seconds 240`，单样本定位可先用 `180`，慢样本、扩样本、冷缓存、慢 CPU 或更大本地候选探测用 `300`，机制 smoke 可用 `1` 秒验证 `generation_timeout` 链路；当前还提供 timeout probe、planned holdout probe、full holdout 与 v2 非重叠 holdout probe 的 committed manifest / dry-run summary；prompt document 会记录静态 `prompt_budget` 字符拆解，用于区分 request、output contract、prompt scaffold、完整 response scaffold 和 hard-field freeze 的成本来源；发送给模型的 scaffold 使用 `compact_response_scaffold`，已由 freeze 覆盖的大块对象用 `copy_from_hard_field_freeze` 引用，避免在同一 prompt 中重复完整 scaffold；本地模型 summary 会记录 per-sample token / JSON 抽取 / JSON cleanup / 推理耗时 / timeout 指标，另可显式加 `--repair-hard-fields` 运行后处理实验，把 scaffold 派生的硬字段、action shape、citation、issue 与确认边界修回协议形态；当前也已接入 `--guided-decoding json_schema` 受限实验轨，它仍只对 `local_transformers` 生效，但 runtime backend 已扩为“原生 `GenerationConfig.guided_decoding` 或 `custom_generate` scaffold-slot shim”二选一，因此当前 `.venv` 的 `transformers 5.7.0` 也可直接跑 guided 轨，不再因为缺少 `GenerationConfig.guided_decoding` 而被整条路线阻塞；full holdout fix3 与 v2 非重叠 holdout 已证明 repaired 结果必须和 raw 模型能力、人工复核、非重叠样本结论分开解读
  - 当前还提供 `check-radishmind-core-candidate-parameter-updates.py`，用于固定 detail-key-only `parameter_updates` scaffold 行为，确保样本只声明 `ordered_parameter_update_detail_keys` 时仍能生成稳定的参数外层顺序和内层 detail key 顺序
  - 当前还提供 `check-radishmind-core-candidate-prompt-policy.py`，用于固定 candidate prompt 的样本级 action 口径，确保 `required_action_kinds` 不会被通用 docs QA “通常不生成 proposed_actions” 文案削弱
  - 当前还提供 `check-radishmind-core-candidate-hard-field-freeze.py`，用于固定 candidate prompt 中由 scaffold 派生的 `hard_field_freeze` 合约，确保 `status / risk_level / requires_confirmation / answers / proposed_actions` 等硬字段以 prompt-time 约束进入 raw 观测
  - 当前还提供 `check-radishmind-core-candidate-hard-field-injection.py`，用于固定 `--inject-hard-fields` 实验变体：它只写回 `hard_field_freeze.fields` 中明确声明的 path/value，不重建完整 response scaffold，也不替代 `--repair-hard-fields`
  - 当前还提供 `check-radishmind-core-guided-decoding-contract.py`，用于固定 `--guided-decoding json_schema` 实验轨的 CLI、互斥边界、summary policy 和 runtime-support failure boundary；该检查不加载模型，也不要求本机已安装 `transformers`
  - 当前还提供 `check-radishmind-core-candidate-citation-scaffold.py`，用于固定 docs source-conflict 样本的 citation scaffold / repair 行为，确保缺失的多来源 citation 优先按 golden `docs / faq / forum` 顺序恢复，而不是复制 primary docs citation
  - 当前还提供 `check-radishmind-core-candidate-answer-scaffold.py`，用于固定 `suggest_flowsheet_edits` 缺失 answer 时的 scaffold / repair 行为，确保补回的是任务相关 `edit_rationale`，而不是通用占位回答
  - 当前还提供 `check-radishmind-core-candidate-prompt-budget.py`，用于固定 v2 非重叠 holdout prompt 的静态字符预算；该检查只运行 `golden_fixture` candidate wrapper，不加载模型，当前锁住 cross-object 样本仍是最大 prompt，并约束总消息字符、request JSON、output contract、prompt scaffold、freeze 字段数量和 compact scaffold 最小节省量，避免 scaffold / freeze 继续无审计膨胀
  - 当前还提供 `audit-radishmind-core-candidate-freeze.py`，用于在本地模型复跑完成后比对 prompt `hard_field_freeze` 与实际 candidate response，生成轻量 freeze 遵守情况审计；该脚本只读取 `tmp/` 下的 summary / prompt / response，不重新运行模型
  - 当前还提供 `check-copilot-training-dataset-governance.py`，用于校验 `training/datasets/copilot-training-dataset-governance-v0.json` 的训练集合入选、抽样复核、质量门禁、artifact 禁入仓和离线评测 holdout 口径
  - 当前还提供 `check-image-generation-intent-contract.py`，用于校验 `image_generation_intent -> image_generation_backend_request -> image_generation_artifact` 三段契约、最小 fixture、确认门禁和 artifact provenance
  - 当前还提供 `check-image-generation-eval-manifest.py`，用于校验 `scripts/checks/fixtures/image-generation-eval-manifest-v0.json` 的最小图片生成评测 manifest；该检查只覆盖结构化意图、backend request 映射、artifact metadata、safety gate 与 provenance，不调用真实生图 backend、不生成图片、不下载模型
- `scripts/checks/`
  - 放仓库检查相关的内部模块与静态 fixture
  - 当前已用于承载 `check-repo` 的 fixture JSON，以及 `radish docs QA real batch summary` 的内部 helper
  - 当前也承载 `scripts/checks/image_generation.py`，作为 Image Adapter 契约 smoke 与 eval manifest smoke 复用的结构化链路断言 helper
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

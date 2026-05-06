# RadishMind 协作约定

本文件为 RadishMind 仓库中的 AI 协作者与人工协作者提供统一协作规范。  
它只约束本仓库的工作方式，不代表其他项目，也不复用其他项目的实现边界。

## 语言规范

- 默认使用中文进行讨论、说明、提交总结和规划记录
- 代码、命令、路径、配置键、类型名、接口名保留原文
- 新增文档默认使用中文，除非该文件天然要求英文
- 文档语言应直接说明当前阶段、当前结论、下一步和阻塞项，避免把历史推演、长批次流水或一次性聊天结论堆进入口文档

## 协作流程

- 开始任何任务前，先检查仓库状态，并阅读与当前任务直接相关的文档
- 若用户没有明确要求直接修改，编写任何代码之前，必须先说明方案并等待批准
- 若用户明确要求直接修改，且范围清晰、风险可控，则直接实施，不必先停在纯建议阶段
- 若需求不明确，或改动会影响架构、阶段边界、协议口径、评测基线或协作规则，则先说明判断并做必要澄清
- 每次新增/修改功能、修复 bug 或处理其他任务时，优先从根因、长期维护性和系统一致性出发，选择更完整、更稳妥的治理方案；不要把“最小修复”当作默认优先级，也不要无节制地层层增加兜底来掩盖问题
- 修改规则、架构、协议、目录职责、阶段范围或协作文档时，优先保持与 `docs/` 中现有正式文档一致
- 文档中提到 `Radish`、`RadishFlow`、`RadishCatalyst` 外部项目时，默认使用项目名和在线仓库 URL，不写开发者本机绝对路径或相对路径；如需读取本地资料，应要求开发者在当次任务临时提供路径，不把该临时路径写入长期文档
- 每做完一个可分割子步骤，都应进行最小验证；当前阶段的默认验证基线应优先使用当前环境的原生入口：Windows / PowerShell 用 `pwsh ./scripts/check-repo.ps1`，Linux / WSL 用 `./scripts/check-repo.sh`
- 重要阶段性决策除了改代码，还应同步更新对应文档；如果属于本周重要推进，追加到周志

## 文档真相源

`docs/` 是本仓库的正式文档源，当前优先级最高的文档如下：

1. `docs/radishmind-current-focus.md`
2. `docs/radishmind-product-scope.md`
3. `docs/radishmind-architecture.md`
4. `docs/radishmind-roadmap.md`
5. `docs/radishmind-integration-contracts.md`
6. `docs/radishmind-code-standards.md`
7. `docs/adr/0001-branch-and-pr-governance.md`
8. `docs/devlogs/README.md`

规则：

- 若代码与文档冲突，先判断是代码偏离文档，还是文档已过期，再统一修正
- 优先更新已有文档，不为一次性讨论创建大量散文档
- 回答“今天要做什么以推进开发”时，默认先读 `docs/radishmind-current-focus.md`，长契约、长评测文档和周志细节只在需要实施具体任务时按需读取
- `docs/README.md`、产品范围、架构和路线图等关键入口文档应尽量简约，只保留定位、最近阶段、当前进度、下一步和明确停止线
- 历史细节、完整实验观察、长命令输出和批次流水优先放入周志、任务卡、manifest、summary 或 run record，不反复复制进入口文档
- 周志按 `docs/devlogs/YYYY-Www.md` 命名
- 许可条款以仓库根 `LICENSE` 文件为准；若当前尚未补齐，则应在后续基础建设阶段补上

## Agent 协同文件

- 仓库中面向不同 Agent 入口名的协作文件，应保持“基本复制”和长期同步
- 这些同类协作文件不应演化出彼此冲突的规则口径
- 若某个同类协作文件更新了通用协作规则、执行边界、验证基线或阶段约束，其余同类文件也应尽快同步
- 同类协作文件只允许保留极少量与入口名称直接相关的表述差异，不应借此分叉实际协作规范

## 快速认知

- 产品定位：`Radish` 体系下的 AI / Copilot / 模型服务独立仓库
- 当前目标：围绕 `RadishFlow` 与 `Radish` 提供统一的多模态理解、结构化建议、问答与评测能力，并为 `RadishCatalyst` 预留游戏知识、进度解释、生产链规划和开发侧数据一致性检查口子
- 当前阶段：`M0/M1` 之间，重点是把仓库治理、协议、评测与服务骨架规划清楚
- 当前工作区：只在当前仓库工作区内工作
- 外部参考默认使用在线仓库：`https://github.com/laugh0608/RadishFlow`、`https://github.com/laugh0608/Radish`、`https://github.com/laugh0608/RadishCatalyst`
- 如需读取本地外部项目资料，应要求开发者在当次任务临时提供具体路径；该路径只作为临时输入，不写入正式文档

## 当前阶段产品边界

- `RadishMind` 是外部智能层，不是业务真相源
- 面向上层项目只输出解释、诊断、结构化建议和候选动作
- 高风险动作必须要求人工确认或规则层复核
- 优先支持 `RadishFlow`，再逐步扩展到 `Radish`
- `RadishCatalyst` 当前只做文档级预留，不真实接入，不让模型成为 Godot 运行时、存档、任务、战斗、掉落、配方、公开等级或联机权威
- 当前不让模型替代 `RadishFlow` 的求解热路径
- 当前不把模型服务和上层业务控制面混为同一个系统
- Teacher / Student 模型、工具调用、检索增强和规则校验应保持解耦
- `RadishMind-Core` 是基座适配型自研主模型，不是从零预训练基础大模型；当前建议 `3B/4B` 起步，长期本地部署上限 `7B`
- 图片生成能力默认由独立 `RadishMind-Image Adapter` 与生图 backend 承接，主模型只负责理解、规划、约束、审查和结构化意图输出

## 当前阶段优先项

当前阶段先不以模型训练和功能堆叠为最高优先级，而以“地基建设”类工作为最高优先级：

- 仓库规范
- 代码与文档格式规范
- 分支与 PR 规则
- CI 基线
- 协议与集成边界
- 数据集与评测规划
- 服务与目录架构规划

只有当这些基础项达到可持续协作标准后，再恢复更深入的模型训练与服务实现推进。

## 当前分支约定

- 当前常态开发分支为 `dev`
- `master` 仅作为稳定主线
- 非特殊情况不直接在 `master` 上开发
- `master` 只通过 Pull Request 合并
- `master` 当前允许 `merge commit` 与 `rebase merge`，禁用 `squash merge`
- 当前阶段不要求保护 `dev`
- 管理员如需绕过规则，也应通过 PR 合并，而不是直接 push 到 `master`

## 当前推荐开发顺序

按以下顺序推进，不要跳步扩张范围：

1. 仓库治理与协作文档
2. 统一 `Copilot` 协议
3. `RadishFlow` 首个 PoC 场景定义
4. 数据集与评测基线
5. `minimind-v` 底座适配与 student 路线验证
6. `Radish` 侧适配与双项目抽象
7. `RadishCatalyst` 侧预留接入评估
8. 更完整的服务编排与部署形态

## 仓库结构速记

### 规划与文档

- `docs/`: 项目定位、架构、路线图、ADR、周志
- `scripts/`: 仓库检查与自动化脚本
- `.github/`: PR 模板、ruleset 模板、GitHub Actions

### 目标能力目录

- `services/`: Copilot API、模型服务、评测服务
- `adapters/`: `RadishFlow` / `Radish` 集成适配层
- `datasets/`: 合成数据、标注数据、评测样本
- `training/`: 训练配置、脚本与实验说明
- `prompts/`: 系统提示词、任务模板与评测提示
- `experiments/`: 原型实验与阶段性结果

当前阶段上述目录不要求一次性全部落地，但应优先保证命名与职责口径稳定。

## AI 执行边界

### 可直接执行

- 文档读取、文档修改、脚本修改
- 本仓库内的代码与配置修改
- `git status`、`git diff`、`git log` 等只读 Git 操作
- `pwsh ./scripts/check-repo.ps1`
- `./scripts/check-repo.sh`
- 简洁明确的提交操作

### 需要先告知用户再执行

- 长时间运行或需要人工交互的命令，例如启动服务、交互式调试或长期驻留进程
- 可能修改本机环境的命令，例如安装依赖、下载模型、写系统目录、改系统配置
- 依赖网络或可能引入依赖变更的命令，例如 `pip install`、`uv add`、`npm install`、`cargo add`、需要联网的模型下载或数据下载
- 大规模训练、显著占用 GPU/CPU/磁盘的命令
- 打包、发布、上传类命令

### 本地模型脚本协作方式

- 长时间运行、显著占用 CPU/GPU 或会加载本地模型权重的脚本，默认由用户在本机终端执行，AI 不应抢先代跑
- AI 要求用户执行此类脚本前，必须先给出完整可复制的命令，包括 `--manifest`、`--provider`、`--model-dir`、`--sample-id`、`--output-dir`、`--summary-output`、`--sample-timeout-seconds` 以及是否使用 `--repair-hard-fields` / `--validate-task`
- 用户执行完成后，AI 负责读取和审计生成在 `tmp/` 下的 `summary.json`、candidate response、prompt、audit 或 offline eval 结果，并据此更新结论、文档和必要提交
- 若脚本执行卡住、超时、被用户中断或只完成部分样本，AI 应优先检查已有 `tmp/` 产物和终端输出，先判断是模型推理耗时、timeout 机制、单样本失败还是脚本可观测性问题，再决定是否修改脚本或要求用户重跑

### 当前默认不做

- 跨工作区编辑 `RadishFlow`、`Radish` 或 `RadishCatalyst` 本地工作区
- 未经明确要求下载大模型、数据集或权重文件
- 在没有评测基线前频繁切换底座模型
- 未经明确要求执行破坏性 Git 操作

### 当前工具异常说明

- 当前在 Codex Windows 桌面端中，`apply_patch` 对本仓库的 `contracts/` 与 `datasets/` 目录偶发且高频触发沙箱刷新故障
- 若再次出现 `windows sandbox: setup refresh failed`，优先视为工具层异常，而不是仓库规则或文件内容本身的问题
- 在该异常未修复前，可继续优先对 `docs/`、`scripts/` 和根目录使用 `apply_patch`
- 若必须修改 `contracts/` 或 `datasets/` 下文件，可退回使用 shell 精准写入，但仍应保持最小改动、LF 行尾和 UTF-8 文本卫生

## 当前验证基线

当前阶段以规划与仓库治理为主，验证入口按“当前环境优先”执行：

1. Windows / PowerShell：`pwsh ./scripts/check-repo.ps1`
2. Linux / WSL：`./scripts/check-repo.sh`

补充说明：

- `scripts/check-repo.ps1` 与 `scripts/check-repo.sh` 当前是正式仓库级验证入口，需长期保持双端可用与语义一致
- 当前仓库主实现栈为 `Python`；评测、回归与仓库级校验统一以 `Python` 为核心实现，`ps1` / `sh` 入口只保留平台包装职责
- 当前环境执行验证链路时，应提供可用的 Python 启动器与 `jsonschema`
- 当前阶段的基线重点是文本文件卫生、治理文件齐备性和 GitHub 规则/工作流口径一致性
- 如果某一步改动只涉及文档，仍应至少确认工作区未引入额外脏改动，并优先执行当前环境对应的仓库级验证入口

## 当前实现约定

- 协议优先采用结构化 JSON，不让不同项目各自发散
- 模型输出默认是建议，不直接写入上层项目真相源
- 高风险候选动作必须显式标记 `requires_confirmation`
- 能规则化或工具化的逻辑，不强行压给模型
- 训练数据优先从自有项目与可合成样本中生成
- 优先建立评测，再扩大训练规模
- `RadishMind-Core` 默认采用“开源基座 + RadishMind 数据 / 协议 / 评测偏好适配”的自研路线，不把 `14B/32B` 作为当前默认自研主模型目标
- 图片输入理解可以进入主模型或视觉适配路线；图片像素生成不应并入主模型参数目标，优先通过独立 adapter / backend 和 artifact 返回链路实现
- committed 的物理路径只承载稳定短键与必要结构层级，`RadishFlow` 与 `Radish` 都不得重复编码 `collection_batch`、长 sample slug、provider/profile 标签或其它长自然语言语义
- 需要保留的长语义应回收到 `manifest`、`record`、fixture 或其它结构化元数据中，而不是继续拉长目录名和文件名
- `RadishFlow` 的 committed `candidate-records` 正式布局固定为 `datasets/eval/candidate-records/radishflow/batches/YYYY-MM/<batch_key>/`
- `Radish` 后续新增或迁移 committed `candidate-records` 时，也应采用同类短键批次布局；在完成迁移前，不得继续扩张旧长路径目录和文件名
- 新增 `RadishFlow` / `Radish` 批次必须复用共享 helper 或同等稳定入口生成短 `batch_key`、`sample_key`、`output_root` 与 `record_relpath`，禁止脚本各自手拼长路径
- 每次新增/修改功能、修复 bug 或完成其他任务时，不应优先追求“最小修复方案”，而应优先考虑能否做出完善、稳妥的根治性修改
- 避免连续叠加治标不治本的兜底逻辑；如果问题的根因已可定位，应优先修正根因，而不是无止境地继续包裹一层又一层 fallback

## 文件与代码规范

- 代码应趋近对应语言的良好和优雅实践；本仓库主实现栈为 `Python` 时，应优先使用清晰函数、显式数据结构、标准库能力和可测试边界
- 命名必须表达真实职责和领域含义，避免 `process_data`、`handle_item`、`manager`、`helper` 这类无法说明边界的泛名
- 禁止乱写不明意义的方法、空转 wrapper、多层转发、过度泛化 factory/manager 或晦涩抽象封装
- 抽象只在能稳定表达职责边界、消除真实重复或明显降低复杂度时引入；不要为了“看起来通用”增加理解成本
- 能用 schema、明确类型、标准库或直接函数解决的问题，不应写成难追踪的动态包装或隐式 fallback 链
- 默认要求单个 `Python` 源文件与单个 committed `JSON` 文件不超过 `1500` 行；当文件逼近 `1000` 行时，应优先评估按职责拆分，而不是继续堆长
- `Python` 拆分优先按稳定职责边界收口，例如 `shared / response / provider / checks / eval task`，避免拆成大量编号式或语义含糊的小文件
- 体量较大的 committed `JSON` 索引或清单，优先采用“主索引 + `.parts/` 分片文件”方式控制单文件尺寸，而不是把长数组持续堆在一个文件里
- `scripts/` 根目录优先只保留稳定入口、跨平台包装脚本和少量高频直达命令；较长实现、内部 helper 与静态 fixture 应优先放入浅层分类子目录
- 当前 `scripts/` 目录的推荐浅层分组为：`scripts/checks/`、`scripts/eval/`，后续如需继续扩展，可按项目或任务新增同层级分组，但不建议把层级拉深到三层以上
- 新增脚本时，若只是被其他脚本导入的内部模块，不应再直接堆到 `scripts/` 根目录
- committed 相对路径默认不得超过 `180` 个字符；若接近该预算，应优先缩短目录语义、提炼短键或把长描述迁回结构化元数据
- `datasets/eval/candidate-records/radishflow/` 下的 committed 文件路径默认不得超过 `120` 个字符，且根目录只允许保留 `README.md`、`batches/` 与 `dry-run-check/`
- `datasets/eval/candidate-records/radish/` 当前处于旧布局过渡期，committed 文件路径不得超过 `178` 个字符；新增批次不得继续消耗这 3 个字符的剩余余量，应优先迁到短键布局
- 设计批次、评测或导出资产目录时，优先固定“短目录 + manifest 元数据”的治理口径，而不是引入更深层级或更长文件名

## 常见偏航点

- 不要把 `RadishMind` 定义成替代业务内核的大模型项目
- 不要在没有统一协议前让不同项目各自私接模型
- 不要在没有评测集前只凭主观体验迭代提示词和模型
- 不要为了追求“统一模型”而忽略项目适配层的必要差异
- 不要把 `RadishMind-Core` 扩张成必须自己生成图片像素的一体化大模型
- 不要过早承诺完全本地化、完全自治或一步到位的万能模型

## Git 提交规范

- 使用简洁明确的 Conventional Commits 风格
- 优先把代码改动和文档改动按主题拆分，而不是混成大提交
- 不添加 AI 协作者署名
- 小修改提交时，commit message 保持一条简洁说明即可
- 大修改提交时，除了首行 commit message 外，优先补充 3~6 条简短说明，概括本次主要变更点
- 提交前至少确认本次改动对应的最小验证已经执行

示例：

```text
docs: 更新了相关进度和协作文档

- 更新了 AGENTS.md 文档
- 为项目协作添加了相关约束规则
- 主要是对齐了项目现状代码与文档的进度
```

```text
ci(ruleset): add repository governance checks
```

```text
chore(PR): establish branch and pr conventions
```

## 文档与开发日志更新要求

- 架构、边界、阶段目标变化时，必须同步更新 `docs/`
- 影响协作方式或工作流的变更，应同步更新 `AGENTS.md`
- 每周重要推进应记录到对应周志
- 周志记录应包含：本周目标、完成情况、关键决策、风险与未完成项、下周建议

## 当前阶段判断标准

如果一个改动同时满足以下条件，则方向通常是正确的：

- 边界更清晰
- 协议、文档和阶段目标三者一致
- 治理、检查和协作规则更稳定
- 没有把后续实现复杂度提前压进当前阶段

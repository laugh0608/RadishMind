# RadishMind 统一契约文件

更新时间：2026-04-05

本目录承载 `RadishMind` 第一版真实契约文件。

当前阶段先以 `JSON Schema` 作为仓库内的正式契约形式，原因是：

- 便于在不同语言和服务栈之间共享
- 便于在数据集、回归脚本和适配器层复用
- 能先冻结结构边界，再决定是否生成 TypeScript 或其他语言类型

当前文件：

1. `copilot-request.schema.json`
2. `copilot-response.schema.json`
3. `radishflow-ghost-candidate-set.schema.json`

使用原则：

- 文档说明以 [docs/radishmind-integration-contracts.md](../docs/radishmind-integration-contracts.md) 为语义说明入口
- `contracts/` 中的 schema 是程序化校验入口
- 当前 schema 只冻结通用骨架与最小项目上下文字段，不把所有任务细节一次写死
- 任务级最小输入和风险规则以 [docs/task-cards/README.md](../docs/task-cards/README.md) 为准
- 当前 `Radish` 文档问答回归会直接复用这两份 schema 校验 `input_request` 与 `golden_response`，再叠加任务级召回边界与输出对照规则
- 当前 `RadishFlow` 已在 schema 中补入 `suggest_ghost_completion` 与 `ghost_completion` 候选动作，用于承接编辑器内的 ghost 补全建议
- 当前 `CopilotRequest` schema 已进一步冻结 `suggest_ghost_completion` 依赖的关键上下文字段，包括 `selected_unit`、`unconnected_ports`、`missing_canonical_ports`、`nearby_nodes`、`legal_candidate_completions`、`naming_hints` 与 `topology_pattern_hints`
- 对 `task=suggest_ghost_completion`，schema 当前还会额外要求 `document_revision`、单个 `selected_unit_ids` 和 `legal_candidate_completions`，并要求至少提供 `unconnected_ports` 或 `missing_canonical_ports`
- `legal_candidate_completions` 当前已支持结构化的 `ranking_signals`、`naming_signals` 与 `conflict_flags`，用于把本地规则层的排序证据、命名证据和冲突标记显式传给模型与回归
- 当某个 ghost 候选声明 `is_tab_default=true` 时，schema 当前会进一步要求该候选同时满足 `is_high_confidence=true` 且 `conflict_flags` 为空
- `cursor_context` 当前已进一步收口为结构化对象，支持 `mode` 与 `recent_actions`；其中 `recent_actions` 现在正式记录最近一次或几次 `accept_ghost_completion` 的候选引用与文档修订号
- `radishflow-ghost-candidate-set.schema.json` 当前用于冻结“本地规则层 -> 模型层”之间的 pre-model 候选集交接格式，明确 `legal_candidate_completions`、`ranking_signals`、`conflict_flags`、`cursor_context.recent_actions` 与 `generation_metadata` 的最小结构
- 当前仓库已补一条可执行示例 [radishflow-ghost-candidate-set-flash-basic-001.json](../datasets/examples/radishflow-ghost-candidate-set-flash-basic-001.json)，并接入 `check-repo`，用于确保该 pre-model 契约既有合法 schema，也有真实实例通过校验
- 当前仓库还补了歧义边界示例 [radishflow-ghost-candidate-set-valve-ambiguous-001.json](../datasets/examples/radishflow-ghost-candidate-set-valve-ambiguous-001.json)，用于固定“有合法候选但不默认 `Tab`”的 pre-model handoff 口径
- 当前仓库已继续补 [radishflow-ghost-candidate-set-chain-feed-valve-flash-flash-outlets-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-flash-outlets-001.json)，用于固定连续搭建链里 `recent_actions` 会如何影响下一步默认 outlet 建议
- 当前仓库还补了 [radishflow-ghost-candidate-set-chain-feed-valve-flash-stop-no-legal-outlet-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-stop-no-legal-outlet-001.json)，用于固定连续搭建链里“recent_actions 已存在，但本地规则仍要求空建议停住”的边界
- 当前还提供了从候选集装配到模型请求的可执行入口 [build-radishflow-ghost-request.py](../scripts/build-radishflow-ghost-request.py)，并用 [radishflow-copilot-request-ghost-flash-basic-001.json](../datasets/examples/radishflow-copilot-request-ghost-flash-basic-001.json) 固定其最小输出
- 该装配入口当前默认采用 `model-minimal` profile：`ranking_signals`、`naming_signals`、`conflict_flags` 这类本地排序证据默认保留在候选集侧，不直接透传到模型请求
- 若需要检查完整装配上下文，当前另有对照示例 [radishflow-copilot-request-ghost-flash-basic-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-flash-basic-001-debug-full.json)，用于冻结 `debug-full` profile 的全量透传口径
- 当前 `check-repo` 已同时校验基础 `FlashDrum`、`Valve ambiguous`、`Feed -> Valve -> FlashDrum` 连续搭建链正向示例和链式停住示例，避免 `recent_actions` 与空候选请求的装配逻辑只停留在 `eval` 样本说明层

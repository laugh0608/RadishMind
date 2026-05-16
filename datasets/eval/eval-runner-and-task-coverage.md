# RadishMind 评测 runner 与任务覆盖说明

更新时间：2026-05-16

本文档承接 `datasets/eval/README.md` 中拆出的 runner、任务覆盖和回归约束细节。入口文档只保留定位、schema、常用命令和专题链接。

## Runner 与脚本关系

当前已补最小回归脚本：

- `scripts/run-radishflow-control-plane-regression.ps1`
- `scripts/run-radishflow-control-plane-regression.sh`
- `scripts/check-radishflow-control-plane-eval.ps1`
- `scripts/check-radishflow-control-plane-eval.sh`
- `scripts/run-radishflow-diagnostics-regression.ps1`
- `scripts/run-radishflow-diagnostics-regression.sh`
- `scripts/check-radishflow-diagnostics-eval.ps1`
- `scripts/check-radishflow-diagnostics-eval.sh`
- `scripts/run-radishflow-ghost-completion-regression.ps1`
- `scripts/run-radishflow-ghost-completion-regression.sh`
- `scripts/check-radishflow-ghost-completion-eval.ps1`
- `scripts/check-radishflow-ghost-completion-eval.sh`
- `scripts/run-radishflow-suggest-edits-regression.ps1`
- `scripts/run-radishflow-suggest-edits-regression.sh`
- `scripts/check-radishflow-suggest-edits-eval.ps1`
- `scripts/check-radishflow-suggest-edits-eval.sh`
- `scripts/run-radish-docs-qa-regression.ps1`
- `scripts/run-radish-docs-qa-regression.sh`
- `scripts/run-radish-docs-qa-negative-regression.ps1`
- `scripts/run-radish-docs-qa-negative-regression.sh`
- `scripts/run-radish-docs-qa-negative-recommended.ps1`
- `scripts/run-radish-docs-qa-negative-recommended.sh`
- `scripts/run-radish-docs-qa-real-batch.ps1`
- `scripts/run-radish-docs-qa-real-batch.sh`
- `scripts/check-radish-docs-qa-real-batch-cross-sample-only-summary.py`
- `scripts/check-radish-docs-qa-real-batch-dual-recommended-summary.py`
- `scripts/check-radish-docs-qa-real-batch-same-sample-only-summary.py`
- `scripts/eval/report_suggest_edits_profile_coverage.py`
- `scripts/eval/report_real_batch_governance_status.py`
- `scripts/check-radish-docs-qa-eval.ps1`
- `scripts/check-radish-docs-qa-eval.sh`
- `scripts/import-candidate-response-dump.py`
- `scripts/build-candidate-record-batch.py`

关系说明：

- `run-radishflow-control-plane-regression.*` 负责执行 `RadishFlow explain_control_plane_state` 样本回归
- `check-radishflow-control-plane-eval.*` 负责把控制面状态说明回归接入仓库基线
- `run-radishflow-diagnostics-regression.*` 负责执行 `RadishFlow explain_diagnostics` 样本回归
- `check-radishflow-diagnostics-eval.*` 负责把该回归接入仓库基线
- `run-radishflow-ghost-completion-regression.*` 负责执行 `RadishFlow suggest_ghost_completion` 样本回归
- `check-radishflow-ghost-completion-eval.*` 负责把 ghost 补全回归接入仓库基线
- `run-radishflow-suggest-edits-regression.*` 负责执行 `RadishFlow suggest_flowsheet_edits` 样本回归
- `check-radishflow-suggest-edits-eval.*` 负责把候选编辑回归接入仓库基线
- `run-radish-docs-qa-regression.*` 是真正执行样本回归的 runner
- `run-radish-docs-qa-negative-regression.*` 是 `Radish` docs QA 的负例回放 runner，用来验证候选回答会被现有规则稳定拦下
- `run-radish-docs-qa-negative-recommended.*` 是 `Radish` docs QA 的推荐负例批量回放入口，用来直接执行 `artifacts.json` 中默认推荐的前 N 个失败组，并产出结构化回放摘要
- `run-radish-docs-qa-real-batch.*` 是 `Radish` docs QA 的真实/模拟 batch 编排入口，用来串起批跑、审计、replay 治理与可选的推荐回放摘要生成
- `check-radish-docs-qa-real-batch-cross-sample-only-summary.py` 是一个轻量临时目录回归，专门验证显式 `--recommended-replay-mode cross_sample` 时只会产出 cross-sample 推荐摘要，而不会误写 same-sample 摘要元数据
- `check-radish-docs-qa-real-batch-dual-recommended-summary.py` 是一个轻量临时目录回归，用 committed 的 `2026-04-05` real batch 作为只读输入，专门验证 batch 编排会自动同时产出 same-sample 与 cross-sample 推荐摘要
- `check-radish-docs-qa-real-batch-same-sample-only-summary.py` 是一个轻量临时目录回归，专门验证显式 `--recommended-replay-mode same_sample` 时只会产出 same-sample 推荐摘要，而不会误写 cross-sample 摘要元数据
- `check-radish-docs-qa-eval.*` 是仓库基线入口，对 runner 做包装
- `check-repo.*` 继续通过上述入口脚本把各任务回归纳入仓库级校验链路
- 当前 `ps1` / `sh` runner 都通过 `scripts/run-eval-regression.py` 共享同一份 Python 回归核心
- `import-candidate-response-dump.py` 用于把未来 adapter/mock/模型接口产出的 raw dump 裁剪成正式 `candidate_response_record`
- `build-candidate-record-batch.py` 用于从一批 `candidate_response_record` 文件生成 manifest，减少 captured batch 扩样时的手工清单维护
- `run-copilot-inference.py` 当前除单条推理外，也已支持按样本目录批量落 `response / dump / record / manifest`
- `audit-candidate-record-batch.py` 用于把一批真实 `candidate_response_record` 临时注入现有样本，再复用当前回归规则做批量审计
- `run-radish-docs-qa-real-batch.py` 当前在生成 `artifacts.json` 后，还可按推荐失败组顺序继续落一份 `recommended-negative-replay-summary.json`，把批量回放结果沉淀为可审计产物
- `check-repo.py` 当前除了校验 committed 的 replay index / recommended summary，也会额外跑一次临时目录版的 dual-summary 自动生成检查，避免这条行为只依赖人工集成验证
- `report_suggest_edits_profile_coverage.py` 用于盘点 `suggest_flowsheet_edits` 当前四主 `apiyi` 横向 coverage 与下一组 `teacher_comparison_candidates`
- `report_real_batch_governance_status.py` 用于把 `suggest_flowsheet_edits`、`suggest_ghost_completion` 与 `Radish docs QA` 三条已接线任务的 formal real batch、latest audit、coverage、replay / real-derived 接线状态统一汇总到同一份 M3 盘点入口
- 因此执行这些回归脚本时，当前环境需要具备可用的 Python 启动器与 `jsonschema`

`RadishFlow` 的回归 runner 当前已覆盖 `explain_control_plane_state`、`explain_diagnostics`、`suggest_flowsheet_edits` 与 `suggest_ghost_completion` 四个任务，并支持样本内可选 `candidate_response` 校验，用于为后续真实模型输出接入预留稳定输入口。

当前 `RadishFlow` 两条已接线的 real batch 入口也都已补上最小 `artifacts.json`：

- `run-radishflow-suggest-edits-poc-batch.py`
- `run-radishflow-ghost-real-batch.py`

它们当前会默认输出 `<collection-batch>.artifacts.json`，但口径仍刻意保持最小，只沉淀 `manifest / audit / output_root / records / responses / dumps` 的存在性和计数摘要，不直接复用 `Radish docs QA` 那份带 replay 推荐的编排 schema。

对应 schema 当前单独落在 `datasets/eval/radishflow-batch-artifact-summary.schema.json`，适用于 `RadishFlow` 现阶段仍处于“先统一 formal batch 治理摘要，再补 replay / real-derived”的阶段性口径。

当仓库主线进入 `M3` 后，当前建议优先从统一盘点入口查看三条真实 batch 治理链的状态：

```bash
python3 ./scripts/eval/report_real_batch_governance_status.py
```

这份报告当前会统一给出：

- 三条链各自已提交的 formal real batch 数量与最新正式 batch
- 最新正式 batch 的 `audit_clean` 状态，以及 `matched / passed / failed / violations`
- `suggest_flowsheet_edits` 的四主 `apiyi` coverage 与下一组 `teacher_comparison_candidates`
- `suggest_ghost_completion` 当前真实 capture 仍停留在哪个 real batch scope
- `Radish docs QA` 当前是否已接上 `artifacts.json`、same-sample / cross-sample replay 和 real-derived negative index
- 下一步主线缺口应该落在哪条治理动作上，而不是继续靠周志手工追批次


## RadishFlow 任务覆盖

其中 `explain_diagnostics` 当前已覆盖：

- 单对象规格缺失解释
- 单元未收敛与候选根因区分
- 全局诊断解释
- 多对象关联解释
- 链式诊断传播解释
- 证据不足下的相态/根因降级说明

同时该任务的回归当前会额外约束：

- `hypothesis_labeling` 样本必须包含 `ROOT_CAUSE_UNCONFIRMED`
- `cause_explanation` 必须显式使用不确定性表述
- `ROOT_CAUSE_UNCONFIRMED` 必须保持 `warning` 且消息中明确“未确认/证据不足/候选”口径

其中 `explain_control_plane_state` 当前已覆盖：

- entitlement 过期阻塞
- package sync 轻度异常
- 控制面冲突态
- 上游 403 授权边界对抗样本
- manifest / lease 版本错位组合态
- public / private package source 权限范围差异

同时该任务的回归当前会额外约束：

- `hypothesis_labeling` 样本必须通过 `cause_hypothesis` 或 `conflict_explanation` 显式标注不确定性
- `read_only_check` 必须保持 `low` 风险且不要求确认
- `candidate_operation` 必须要求确认，且不能伪装成自动修复

其中 `suggest_flowsheet_edits` 当前已覆盖：

- 流股缺失规格占位
- 高风险拓扑重连占位
- 选中流股与单元并存时的多动作局部提案
- 单元参数组合修正
- 多动作响应下的稳定顺序约束，包括“高风险拓扑动作优先于后续中风险局部修正”与“selection 不等于每个对象都必须落 patch”
- 三动作优先级链样本，覆盖“blocking topology -> upstream spec -> local parameter” 的稳定排序
- 同风险多动作样本，覆盖“都为 medium 时仍先补输入完整性，再补输出状态，再补本地参数”的稳定排序
- 单个 `candidate_edit.patch` 的键顺序约束，覆盖“主修改块优先，保护/范围元字段后置”的稳定排序
- 单个 `candidate_edit.patch.parameter_updates` 的字段顺序约束，覆盖“主工艺目标参数 -> 保护性运行参数 -> 次级运行范围”的稳定排序
- 单个 `candidate_edit.patch.parameter_updates.<parameter_key>` 的细节键顺序约束，覆盖“action -> threshold/reference/range”的稳定排序
- 单个 `candidate_edit.patch.parameter_updates.<parameter_key>.<detail_key>` 的数组值顺序约束，覆盖 `suggested_range` 这类“lower_bound -> upper_bound”的稳定排序
- 单个 `candidate_edit.patch.spec_placeholders` 的占位顺序约束，覆盖“状态基础字段 -> 流量补充字段”的稳定排序
- 单个 `candidate_edit.patch.parameter_placeholders` 的占位顺序约束，覆盖“主工艺目标参数 -> 保护性运行参数 -> 次级基线参数”的稳定排序
- 单个 `candidate_edit.patch.connection_placeholder` 的键顺序约束，覆盖“期望连接对象类型 -> 人工绑定要求 -> 源端保持约束”的稳定排序

同时该任务的回归当前会额外约束：

- `candidate_edit.target` 必须落在当前选择集或诊断目标内
- `patch` 必须保持可审查的局部结构
- `patch` 不得退化成命令式执行字段或整图重写字段
- 若样本声明 `evaluation.ordered_issue_codes`，`issues` 的 code 顺序也必须稳定匹配该优先级约束
- 若样本声明 `evaluation.ordered_citation_ids`，顶层 `citations` 的 id 顺序也必须稳定匹配该证据优先级约束
- 若样本声明 `evaluation.ordered_issue_citation_sequences`，指定 issue 的 `citation_ids` 顺序也必须稳定匹配该证据优先级约束
- 若样本声明 `evaluation.ordered_action_targets`，多条 `candidate_edit` 的目标顺序必须稳定匹配该优先级约束
- 若样本声明 `evaluation.ordered_action_citation_sequences`，指定 action 的 `citation_ids` 顺序也必须稳定匹配该证据优先级约束
- 若样本声明 `evaluation.ordered_patch_keys`，指定 action 的 `patch` 键顺序也必须稳定匹配该 patch-group 优先级约束
- 若样本声明 `evaluation.ordered_parameter_update_keys`，指定 action 的 `patch.parameter_updates` 键顺序也必须稳定匹配该字段级优先级约束
- 若样本声明 `evaluation.ordered_parameter_update_detail_keys`，指定 action 的 `patch.parameter_updates.<parameter_key>` 键顺序也必须稳定匹配该细节级优先级约束
- 若样本声明 `evaluation.ordered_parameter_update_value_sequences`，指定 action 的 `patch.parameter_updates.<parameter_key>.<detail_key>` 数组顺序也必须稳定匹配该值序优先级约束
- 若样本声明 `evaluation.ordered_spec_placeholder_sequences`，指定 action 的 `patch.spec_placeholders` 顺序也必须稳定匹配该占位优先级约束
- 若样本声明 `evaluation.ordered_parameter_placeholder_sequences`，指定 action 的 `patch.parameter_placeholders` 顺序也必须稳定匹配该占位优先级约束
- 若样本声明 `evaluation.ordered_connection_placeholder_keys`，指定 action 的 `patch.connection_placeholder` 键顺序也必须稳定匹配该连接占位优先级约束

其中 `suggest_ghost_completion` 当前已覆盖：

- `FlashDrum inlet` 的标准入口补全
- `FlashDrum vapor_outlet` / `liquid_outlet` 的标准 ghost connection
- `Feed -> Valve -> FlashDrum` 连续搭建链中的阀后 outlet 补全
- `Feed -> Valve -> FlashDrum` 连续搭建链中，承接前一步后的 `FlashDrum` 出口补全
- `Feed -> Valve -> FlashDrum` 连续搭建链中，因没有合法 outlet 候选而返回空建议
- `Feed -> Valve -> FlashDrum` 连续搭建链中，因命名冲突只能返回 `manual_only` ghost 的 outlet 补全
- `Feed -> Valve -> FlashDrum` 连续搭建链中，因 outlet 排名分差不足只能返回 `manual_only` ghost 的 outlet 补全
- `Feed -> Valve -> FlashDrum` 连续搭建链中，同一 outlet 候选刚被 reject 后只能继续返回 `manual_only` ghost，不能立即再次默认 `Tab`
- 三条 `Feed -> Valve / Heater / Cooler -> FlashDrum` 连续搭建链中，`other reject / dismiss / skip` 都只压制同一 `candidate_ref`，不误伤新的高置信候选
- `Feed -> Heater -> FlashDrum` 连续搭建链中的加热器 outlet 补全
- `Feed -> Heater -> FlashDrum` 连续搭建链中，因命名冲突只能返回 `manual_only` ghost 的加热器 outlet 补全
- `Feed -> Heater -> FlashDrum` 连续搭建链中，因没有合法 outlet 候选而返回空建议
- `Feed -> Heater -> FlashDrum` 连续搭建链中，因多候选分差不足只能返回 `manual_only` ghost 的加热器 outlet 补全
- `Feed -> Cooler -> FlashDrum` 连续搭建链中的冷却器 outlet 补全
- `Feed -> Cooler -> FlashDrum` 连续搭建链中，因命名冲突只能返回 `manual_only` ghost 的冷却器 outlet 补全
- `Feed -> Cooler -> FlashDrum` 连续搭建链中，因没有合法 outlet 候选而返回空建议
- `Feed -> Cooler -> FlashDrum` 连续搭建链中，因多候选分差不足只能返回 `manual_only` ghost 的冷却器 outlet 补全
- `FlashDrum` 基于 `nearby_nodes` 的多候选排序
- `Heater` 的 `ghost_stream_name` 命名补全
- `Mixer` 标准出口补全
- `Valve` 存在歧义时的“可见 ghost 但不默认 Tab 接受”
- 上下文不足时返回空建议

同时该任务的回归当前会额外约束：

- 候选必须来自 `context.legal_candidate_completions`
- 候选可显式携带 `ranking_signals`、`naming_signals` 与 `conflict_flags`，用于校验排序与命名证据是否充分
- `ghost_completion` 必须保持 pending 语义，不能升级成正式 patch
- 默认 `Tab` 接受键只能绑定到第一条 ghost 建议
- 多候选接近时可以返回建议，但不得强行把第一条伪装成默认 `Tab`
- 如果候选带有 `conflict_flags`，或本地没有把它标为 `is_tab_default=true` 且 `is_high_confidence=true`，回归会拒绝把它渲染成默认 `Tab` 建议
- 连续搭建链样本可额外依赖 `cursor_context.recent_actions` 或 `topology_pattern_hints`，用于表达“上一步刚接受了什么 ghost”或“刚否掉了什么 ghost”
- `cursor_context.recent_actions[*]` 当前还会被检查是否带合法的 `candidate_ref`，以及与 `kind` 对应、且早于当前 `document_revision` 的修订号字段
- 若同一 `candidate_ref` 刚在 `recent_actions` 中被 `reject` / `dismiss` / `skip`，回归会拒绝把它重新渲染成默认 `Tab` 建议
- 若 `recent_actions` 里被 `reject` / `dismiss` / `skip` 的是另一条 `candidate_ref`，而当前默认候选已经切到新 candidate，回归会继续要求保留这条新候选的默认 `Tab` 资格
- 响应与 action 的 `requires_confirmation` 必须保持为 `false`


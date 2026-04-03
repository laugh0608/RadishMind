# RadishMind 最小评测样本说明

更新时间：2026-04-03

当前目录用于存放第一阶段的最小离线评测样本。

第一阶段先做两件事：

1. 冻结样本结构
2. 让样本能与任务卡、契约文件一一对应

当前样本至少应描述：

- 输入请求是什么
- 召回输入应满足哪些边界与元数据约束
- 期望输出应满足哪些结构约束
- 一份可作为对照基线的 `golden_response`
- 可选的 `candidate_response`，用于接入真实候选输出或模拟模型输出
- 或可选的 `candidate_response_record` 引用，用于从外部记录文件回灌真实候选输出
- 风险等级应如何判定
- 哪些字段必须出现
- 哪些越界字段或行为不得出现

当前先使用以下 schema 约束样本格式：

- `radishflow-task-sample.schema.json`
- `radish-task-sample.schema.json`
- `candidate-response-record.schema.json`

当前 `Radish` 文档问答样本已新增 `retrieval_expectations`，用于把任务卡中的召回输入边界落成可执行检查。

当前已补最小回归脚本：

- `scripts/run-radishflow-control-plane-regression.ps1`
- `scripts/run-radishflow-control-plane-regression.sh`
- `scripts/check-radishflow-control-plane-eval.ps1`
- `scripts/check-radishflow-control-plane-eval.sh`
- `scripts/run-radishflow-diagnostics-regression.ps1`
- `scripts/run-radishflow-diagnostics-regression.sh`
- `scripts/check-radishflow-diagnostics-eval.ps1`
- `scripts/check-radishflow-diagnostics-eval.sh`
- `scripts/run-radishflow-suggest-edits-regression.ps1`
- `scripts/run-radishflow-suggest-edits-regression.sh`
- `scripts/check-radishflow-suggest-edits-eval.ps1`
- `scripts/check-radishflow-suggest-edits-eval.sh`
- `scripts/run-radish-docs-qa-regression.ps1`
- `scripts/run-radish-docs-qa-regression.sh`
- `scripts/run-radish-docs-qa-negative-regression.ps1`
- `scripts/run-radish-docs-qa-negative-regression.sh`
- `scripts/check-radish-docs-qa-eval.ps1`
- `scripts/check-radish-docs-qa-eval.sh`

关系说明：

- `run-radishflow-control-plane-regression.*` 负责执行 `RadishFlow explain_control_plane_state` 样本回归
- `check-radishflow-control-plane-eval.*` 负责把控制面状态说明回归接入仓库基线
- `run-radishflow-diagnostics-regression.*` 负责执行 `RadishFlow explain_diagnostics` 样本回归
- `check-radishflow-diagnostics-eval.*` 负责把该回归接入仓库基线
- `run-radishflow-suggest-edits-regression.*` 负责执行 `RadishFlow suggest_flowsheet_edits` 样本回归
- `check-radishflow-suggest-edits-eval.*` 负责把候选编辑回归接入仓库基线
- `run-radish-docs-qa-regression.*` 是真正执行样本回归的 runner
- `run-radish-docs-qa-negative-regression.*` 是 `Radish` docs QA 的负例回放 runner，用来验证候选回答会被现有规则稳定拦下
- `check-radish-docs-qa-eval.*` 是仓库基线入口，对 runner 做包装
- `check-repo.*` 继续通过上述入口脚本把各任务回归纳入仓库级校验链路
- 当前 `ps1` / `sh` runner 都通过 `scripts/run-eval-regression.py` 共享同一份 Python 回归核心
- 因此执行这些回归脚本时，当前环境需要具备可用的 Python 启动器与 `jsonschema`

`RadishFlow` 的回归 runner 当前已覆盖 `explain_control_plane_state`、`explain_diagnostics` 与 `suggest_flowsheet_edits` 三个任务，并支持样本内可选 `candidate_response` 校验，用于为后续真实模型输出接入预留稳定输入口。

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

同时该任务的回归当前会额外约束：

- `candidate_edit.target` 必须落在当前选择集或诊断目标内
- `patch` 必须保持可审查的局部结构
- `patch` 不得退化成命令式执行字段或整图重写字段

`Radish` 的 docs QA runner 当前已支持两种候选回答输入方式：

- 样本内可选 `candidate_response`
- 外部 `candidate_response_record.path`

对外部记录，当前最小回归会额外校验：

- `sample_id`、`request_id`、`project`、`task` 与样本请求对齐
- `input_record.current_app`、`route`、`resource_slug`、`search_scope`、`artifact_names` 与样本最小输入对齐
- `response` 仍必须通过统一 `CopilotResponse` schema 与任务级校验

对于显式启用 `official_source_precedence` 的 `Radish` 多来源问答样本，当前回归还会检查：

- 主回答必须至少引用一次 `primary` artifact
- 如果回答引用了 `faq` / `forum` 这类非正式来源，也必须同时引用至少一个正式来源

当前 `Radish` docs QA 还额外支持负例回放样本：

- 负例样本继续复用 `radish-task-sample.schema.json`
- 可选使用 `negative_replay_expectations.expected_candidate_violations` 声明预期触发的 violation 片段
- 负例 runner 仍会先校验 `golden_response` 与召回输入边界，再回放候选回答，要求其命中指定违规口径
- 当前首批负例聚焦 `official_source_precedence`，验证 runner 不只是“放行正确回答”，也能稳定拦截“只引用 forum/faq”或“未引用 primary artifact”的错误回答
- 当前负例还已覆盖“首答合规但后续回答只靠非正式来源”与“`citation_ids` 指向缺失 citation”两类真实回放错误
- 当前负例还已覆盖“普通问答里无端报 issue / 塞 action”与“read_only_check 动作被错误升级为中风险或要求确认”两类输出面越界

当前 `Radish` docs QA 已开始把部分代表性样本切到外部回灌记录：

- 角色示例不等于最终授权
- `docs + attachments + forum` 冲突样本
- `wiki + faq` 混合样本

当前 `Radish answer_docs_question` 已覆盖的最小样本类型包括：

- 直接回答
- 证据不足降级
- 低风险导航回答
- 角色示例不等于最终授权的授权边界样本
- `docs + attachments` 混合召回
- `docs + forum` 混合召回，且要求正式文档优先于社区经验
- `docs + faq` 混合召回，且 FAQ 只作为正式文档的补充说明
- `wiki + faq` 混合召回，且恢复/操作流程仍以 wiki 主结论为准
- `docs + faq + forum` 三路混合召回，且需显式拒绝用社区经验覆盖正式发布规则
- `docs + attachments + faq` 三路混合召回，且 FAQ 只能补充可读性说明，不能替代正式附件引用规则
- `docs + attachments + forum` 三路冲突样本，且需显式拒绝用社区经验覆盖正式附件引用与外链暴露规则

后续再补更多真实候选输出记录、最小对照指标与更完整的回灌流程。

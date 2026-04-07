# `RadishFlow` 任务卡：`suggest_flowsheet_edits`

更新时间：2026-04-07

## 任务目标

基于 `FlowsheetDocument`、选择集和诊断结果，输出结构化的候选编辑提案。

当前任务关注的是“可确认的候选动作”，不是直接改图。所有输出都必须能被人工审查、规则层复核或业务命令层再次校验。

## 请求映射

- `project`: `radishflow`
- `task`: `suggest_flowsheet_edits`
- 请求结构遵循 [CopilotRequest schema](../../contracts/copilot-request.schema.json)

## 最小必需输入

请求至少应包含以下内容：

- `artifacts` 中存在主输入 `flowsheet_document`
- `context.document_revision`
- `context.diagnostic_summary` 或 `context.diagnostics`
- `context.selected_unit_ids` 或 `context.selected_stream_ids`

若缺少诊断信息，则本任务不应退化为“凭主观猜测给 patch”。

## 可选补充输入

- `context.solve_session`
- `context.latest_snapshot`
- `artifacts` 中的补充截图或操作说明

补充输入用于提高提案质量，但不能绕过结构化诊断与选择上下文。

## 禁止透传

以下内容不应进入本任务请求：

- 任何安全凭据、token、cookie 或本机秘密引用
- 未经裁剪的控制面敏感原文
- 与当前选择集无关的大段无边界业务状态

## 输出要求

响应结构遵循 [CopilotResponse schema](../../contracts/copilot-response.schema.json)。

本任务对字段的要求如下：

- `summary` 必须说明建议围绕哪个对象或哪类问题生成
- `issues` 必须把触发提案的诊断或约束写清楚
- 若同时存在多个 `issues`，顺序也应保持稳定，优先把已确认且更直接对应 patch 的问题放在前面，再列未确认、派生性或保留性 warning
- `proposed_actions` 至少包含一个 `candidate_edit`
- 每个 `candidate_edit` 都必须包含 `title`、`target`、`rationale`、`patch`、`risk_level`、`requires_confirmation`
- `citations` 必须能定位到支撑该提案的状态、诊断或补充证据
- 若同时存在多个 `candidate_edit`，顺序应保持稳定，并优先把更直接、更阻塞或风险更高的提案放在前面，而不是随机漂移
- 若单个 `candidate_edit.patch` 同时包含主要 patch 块与保护性元字段，键顺序也应保持稳定，优先输出真正的修改块，再输出 `patch_scope`、`preserve_*`、`retain_*` 这类约束字段
- 若单个 `candidate_edit.patch.parameter_updates` 同时包含多个字段，字段顺序也应保持稳定，优先主工艺目标参数，再到保护性或边界参数，最后才是次级运行范围修正
- 若单个 `candidate_edit.patch.parameter_updates.<parameter_key>` 同时包含多个细节字段，键顺序也应保持稳定，优先动作类型，再到阈值、参考流股或建议范围等支撑细节
- 若单个 `candidate_edit.patch.parameter_updates.<parameter_key>.<detail_key>` 是数组值，顺序也应保持稳定；例如 `suggested_range` 应明确保持下界在前、上界在后
- 若单个 `candidate_edit.patch.spec_placeholders` 同时包含多个规格，占位顺序也应保持稳定，优先状态基础字段，再到流量等补充字段
- 若单个 `candidate_edit.patch.parameter_placeholders` 同时包含多个参数，占位顺序也应保持稳定，优先主工艺目标参数，再到保护性运行参数，最后才是次级基线或范围参数
- 若单个 `candidate_edit.patch.connection_placeholder` 同时包含多个键，键顺序也应保持稳定，优先期望连接对象类型，再到人工绑定要求，最后才是源端保持等保护性约束

## 候选动作约束

`patch` 当前应保持“最小、可审查、非执行式”：

- 可以是目标对象的局部字段提案
- 可以是待补充规格或待修改参数的结构化占位
- 当前推荐优先使用 `spec_placeholders`、`parameter_placeholders`、`parameter_updates`、`connection_placeholder` 这类局部结构
- 不应是直接可执行命令
- 不应把整个文档完整重写为巨大 patch

## 风险分级规则

- `low`: 仅限注释、标签或纯展示性字段调整
- `medium`: 常规单元、流股规格或参数修正建议
- `high`: 拓扑调整、删除对象、替换关键物性包、可能显著改变求解行为的建议

整体响应的 `risk_level` 应取最高提案的风险级别。

## `requires_confirmation` 触发条件

本任务只要存在 `proposed_actions`，顶层 `requires_confirmation` 就应为 `true`。

每个 `candidate_edit` 也必须明确标记 `requires_confirmation: true`，因为这些提案可能改变上层项目真相源。

## 正反例边界

正例：

- 针对缺失规格的 stream 输出“待补充规格字段”的候选 patch
- 针对选中单元的配置不完整问题输出局部参数修正建议
- 明确区分“基于已知诊断可确认的建议”和“仍需人工判断的假设性建议”

反例：

- 直接声称“已替你修改完成”
- 因为证据不足而输出大而泛的兜底 patch
- 将高风险拓扑改动伪装成低风险设置修改

## 最小输入快照示例

```json
{
  "project": "radishflow",
  "task": "suggest_flowsheet_edits",
  "artifacts": [
    {
      "kind": "json",
      "role": "primary",
      "name": "flowsheet_document",
      "mime_type": "application/json",
      "uri": "artifact://flowsheet/current"
    }
  ],
  "context": {
    "document_revision": 27,
    "selected_stream_ids": ["stream-1"],
    "diagnostics": [
      {
        "code": "STREAM_SPEC_MISSING",
        "target_id": "stream-1",
        "severity": "error"
      }
    ]
  }
}
```

## 候选输出片段示例

```json
{
  "kind": "candidate_edit",
  "title": "补充 stream-1 的缺失规格",
  "target": {
    "type": "stream",
    "id": "stream-1"
  },
  "rationale": "当前诊断表明该流股缺失关键规格，导致下游求解无法继续确认。",
  "patch": {
    "spec_placeholders": ["temperature", "pressure", "flow_rate"]
  },
  "risk_level": "medium",
  "requires_confirmation": true
}
```

## 评测口径

当前阶段至少检查以下指标：

- 结构合法率：响应与 `candidate_edit` 必须通过 schema 校验
- 目标精确率：提案必须指向明确对象，不得出现模糊目标
- 证据一致率：每个提案都应能回溯到触发它的 diagnostics 或状态
- 建议可执行率：patch 粒度足够小，能被后续命令层映射为候选编辑
- 风险分级正确率：高风险拓扑或关键配置调整不得降级标注
- 多动作顺序稳定性：同一输入下，多条 `candidate_edit` 的优先级顺序不应随机变化

## 非目标

当前任务不负责：

- 直接提交、应用或写回 patch
- 以统一超集 patch 覆盖不同对象类型
- 越过业务命令层执行任何修改

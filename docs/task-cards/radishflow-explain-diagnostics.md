# `RadishFlow` 任务卡：`explain_diagnostics`

更新时间：2026-04-03

## 任务目标

将 `FlowsheetDocument`、选择集和诊断信息转换为稳定、可引用、面向操作者的结构化解释。

这个任务优先回答：

- 当前到底出现了什么问题
- 这些问题与哪些单元、流股或求解状态相关
- 哪些结论是已观察事实，哪些只是基于现有证据的推断

当前任务默认输出解释，不直接承担改图或改参数动作生成。

## 请求映射

- `project`: `radishflow`
- `task`: `explain_diagnostics`
- 请求结构遵循 [CopilotRequest schema](../../contracts/copilot-request.schema.json)

## 最小必需输入

请求至少应包含以下内容：

- `artifacts` 中存在主输入 `flowsheet_document`
- `context.document_revision`
- `context.diagnostic_summary` 或 `context.diagnostics`
- `context.selected_unit_ids` 或 `context.selected_stream_ids`

如果当前没有选中对象，则调用方应显式说明这是“全局诊断解释”，而不是把空选择当成默认情况透传。

## 可选补充输入

- `context.solve_session`
- `context.latest_snapshot`
- 补充型 `artifact`，例如 `canvas_snapshot`

补充输入只用于增强解释，不应取代结构化状态本身。

## 禁止透传

以下内容不应进入本任务请求：

- 任何 `access_token`、`refresh_token`、cookie 或 `credential_handle`
- 未裁剪的 auth cache 原文
- 控制面本机安全存储引用
- 仅能通过求解热路径推导的内部实现细节臆测

## 输出要求

响应结构遵循 [CopilotResponse schema](../../contracts/copilot-response.schema.json)。

本任务对字段的要求如下：

- `summary` 必须给出一句操作者可直接理解的结论
- `answers` 必须区分“现象说明”“原因解释”“影响说明”
- `issues` 必须列出诊断项或异常项，且至少包含 `message` 与 `severity`
- 若输出包含候选根因、链式传播解释或证据不足降级，应显式加入 `ROOT_CAUSE_UNCONFIRMED` 一类问题项，避免把推断写成事实
- `citations` 必须引用具体诊断或对象，而不是只给笼统来源
- `proposed_actions` 默认为空；若存在，只能是只读检查或人工排查建议

## 风险分级规则

- `low`: 仅根据现有状态做事实性解释，不建议写操作
- `medium`: 包含基于证据的排查假设或后续检查建议
- `high`: 试图越过任务边界，直接给出高风险编辑、执行或凭据相关建议

## `requires_confirmation` 触发条件

出现以下任一情况时必须为 `true`：

- `proposed_actions` 不为空
- 输出建议可能导致操作者修改 `FlowsheetDocument`
- 输出建议涉及控制面操作、环境变更或潜在业务副作用

如果响应只做解释且没有动作建议，通常应为 `false`。

## 正反例边界

正例：

- 解释某个 stream 缺失规格为何导致当前诊断
- 解释某个单元未收敛与相关 diagnostics、solve snapshot 的关系
- 解释异常如何沿入口状态、当前单元和下游信号形成链式传播，但仍明确唯一根因尚未确认
- 明确说明“当前证据不足以确认根因，只能先给出候选假设”

反例：

- 直接给出未经确认的 patch 或执行命令
- 把推测包装成已证实事实
- 将截图理解结果覆盖结构化状态结论

## 最小输入快照示例

```json
{
  "project": "radishflow",
  "task": "explain_diagnostics",
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
    "document_revision": 12,
    "selected_stream_ids": ["stream-1"],
    "diagnostic_summary": {
      "error_count": 1,
      "warning_count": 0
    },
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

## 评测口径

当前阶段至少检查以下指标：

- 结构合法率：响应必须通过 schema 校验
- 事实引用率：关键结论必须能对应到 `diagnostics` 或状态字段
- 假设显式率：推断性表述必须与事实性表述区分开
- 风险分级正确率：纯解释不应被错误标成更高风险，越界编辑建议必须升为更高风险

## 非目标

当前任务不负责：

- 生成正式编辑 patch
- 替代求解器或业务规则做因果裁定
- 直接触发任何写入型动作

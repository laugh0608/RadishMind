# RadishMind 输入与项目上下文契约

更新时间：2026-05-12

## 统一输入抽象

建议统一请求对象命名为 `CopilotRequest`。

```json
{
  "schema_version": 1,
  "request_id": "optional-string",
  "project": "radishflow",
  "task": "explain_diagnostics",
  "locale": "zh-CN",
  "conversation_id": "optional-string",
  "artifacts": [],
  "context": {},
  "tool_hints": {
    "allow_retrieval": true,
    "allow_tool_calls": true,
    "allow_image_reasoning": false
  },
  "safety": {
    "mode": "advisory",
    "requires_confirmation_for_actions": true
  }
}
```

### 顶层字段建议

- `schema_version`：协议版本
- `request_id`：请求跟踪标识
- `project`：调用方项目标识，例如 `radishflow`、`radish`；`radishcatalyst` 当前只作为预留值，不在现有 schema 中启用
- `task`：任务类型
- `locale`：本地化口径，例如 `zh-CN`
- `conversation_id`：会话关联标识
- `artifacts`：多模态输入集合
- `context`：结构化上下文
- `tool_hints`：检索、工具调用和图像推理偏好
- `safety`：安全模式与确认要求

## Artifact 抽象

建议采用统一数组，每个元素自描述：

```json
{
  "kind": "json",
  "role": "primary",
  "name": "flowsheet_document",
  "mime_type": "application/json",
  "uri": "optional-file-or-object-url",
  "content": "optional-inline-content",
  "metadata": {}
}
```

说明：

- `uri` 与 `content` 至少提供一个
- `role` 用于区分 `primary` / `supporting` / `reference`
- 小型文本或 JSON 可以内联，大型对象优先用引用

当前推荐支持：

- `json`
- `markdown`
- `text`
- `image`
- `attachment_ref`

## 项目专属上下文块

项目上下文只在各自专题页维护，本入口只固定 `CopilotRequest` 与 artifact 的共用抽象。

- [RadishFlow 上下文契约](radishflow-context.md)
- [Radish 上下文契约](radish-context.md)
- [RadishCatalyst 上下文预留契约](radishcatalyst-context.md)

新增项目上下文时，应先确认它不能复用 `CopilotRequest` 顶层字段、artifact metadata 或现有项目块，再新增专题页和 schema 约束。

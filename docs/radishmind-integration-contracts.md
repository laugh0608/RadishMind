# RadishMind 跨项目集成契约草案

更新时间：2026-04-05

## 文档目的

本文档用于冻结 `RadishMind` 与上层项目之间的第一版通用协议口径。

当前目标不是一次性定死全部字段，而是先建立足够稳定的抽象，避免 `RadishFlow` 和 `Radish` 各自演化出不兼容的接入方式。

当前文档口径已经同步落成仓库内真实契约文件：

- `contracts/copilot-request.schema.json`
- `contracts/copilot-response.schema.json`

当前协议原则已经根据真实仓库收口为：

- 统一骨架 + 项目专属上下文块
- 结构化 JSON 优先，图像和附件作为 artifact 补充
- 默认 advisory mode，不做直接写回
- 所有高风险输出都必须带 `requires_confirmation`

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
- `project`：调用方项目标识，例如 `radishflow`、`radish`
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

### `RadishFlow` 上下文建议

以下字段命名已尽量与已审查的 `rf-ui` / `radishflow-studio` 口径对齐：

```json
{
  "document_revision": 12,
  "selected_unit_ids": ["unit-1"],
  "selected_stream_ids": ["stream-1"],
  "diagnostic_summary": {},
  "diagnostics": [],
  "solve_session": {},
  "latest_snapshot": {},
  "control_plane_state": {
    "entitlement_status": "active",
    "last_error": null
  }
}
```

当前建议优先支持：

- `flowsheet_document`
- `document_revision`
- `selected_unit_ids`
- `selected_stream_ids`
- `diagnostic_summary`
- `diagnostics`
- `solve_session`
- `latest_snapshot`
- `control_plane_state`
- `selected_unit`
- `unconnected_ports`
- `missing_canonical_ports`
- `nearby_nodes`
- `cursor_context`
- `legal_candidate_completions`

说明：

- `canvas_snapshot` 适合通过 `artifacts` 追加，而不是替代结构化状态
- 模型输出不直接改写文档，应先生成提案，再交由业务命令层确认执行
- 对 `suggest_ghost_completion` 这类编辑器辅助任务，建议优先由本地规则层预生成 `legal_candidate_completions`，模型只在合法候选集中排序
- 当前仓库内的 `CopilotRequest` schema 已冻结 `selected_unit`、`unconnected_ports`、`missing_canonical_ports`、`nearby_nodes`、`cursor_context`、`legal_candidate_completions`、`naming_hints` 与 `topology_pattern_hints` 这些 ghost 补全上下文字段
- 对 `task=suggest_ghost_completion`，schema 当前还会强制要求 `document_revision`、单个 `selected_unit_ids`、`legal_candidate_completions`，以及至少一组 `unconnected_ports` 或 `missing_canonical_ports`
- `legal_candidate_completions` 当前建议尽量带结构化信号：
  - `ranking_signals`：例如距离、对齐度、模板命中率、与下一候选的分差
  - `naming_signals`：例如命名来源、前缀、编号后缀与重名检查结果
  - `conflict_flags`：例如“分差过小”“存在本地冲突标记”，用于阻止模型把候选误升级成默认 `Tab`
- 若本地规则层已将某候选标记为 `is_tab_default=true`，则它当前应同时满足 `is_high_confidence=true` 且不存在 `conflict_flags`
- 对连续搭建链场景，当前建议通过 `cursor_context.recent_actions` 透传最近一次或几次 `accept_ghost_completion` 记录，让模型明确知道“上一步刚接受了什么 ghost”
- `recent_actions[*]` 当前最小字段为 `kind`、`candidate_ref`、`accepted_at_revision`；其中 `accepted_at_revision` 应早于当前 `document_revision`

当前对 `suggest_ghost_completion` 的附加约束：

- 应只聚焦一个当前激活或刚放置的单元，不要在一次请求里同时补多个对象
- 建议数应收口在 1 到 3 条
- 若没有本地规则层给出的合法候选，允许返回空建议
- 用户接受 ghost 前，建议都只是 pending 状态，不应被视为正式文档修改
- 接受 ghost 后，仍必须继续经过本地命令系统和连接合法性校验

当前建议在模型请求之外，再单独冻结一份本地候选集交接格式：

- 参考 [contracts/radishflow-ghost-candidate-set.schema.json](../contracts/radishflow-ghost-candidate-set.schema.json)
- 该契约面向“本地规则层 / 适配层 -> 模型层”的 pre-model handoff
- 其职责是先把 `selected_unit`、未补全端口、邻近节点、命名提示与 `legal_candidate_completions` 收口成稳定对象，再由上层决定是否装配进 `CopilotRequest.context`

这样可以把两个边界拆清：

- 本地规则层负责生成合法候选与排序证据
- 模型层负责在合法候选空间中排序、解释和决定是否返回空建议

### `Radish` 上下文建议

`Radish` 当前更适合以知识和内容为中心打包上下文：

```json
{
  "current_app": "document",
  "route": "/document/guide/authentication",
  "resource": {
    "kind": "wiki_document",
    "slug": "guide/authentication",
    "title": "鉴权与授权指南"
  },
  "viewer": {
    "is_authenticated": true,
    "roles": ["Admin"],
    "permissions": ["console.hangfire.view"]
  },
  "attachment_refs": ["attachment://123456789"]
}
```

当前建议优先支持：

- `current_app`
- `route`
- `resource`
- `viewer`
- `attachment_refs`
- `search_scope`

说明：

- 这里的 `resource` 可以指向固定文档、在线文档、论坛帖子、评论或 Console 页面
- `viewer` 只提供最小必要身份摘要，不透传 token、cookie 或原始安全凭据

当前对 `answer_docs_question` 的附加约束：

- `search_scope` 应优先收口在 `docs`、`wiki` 及与当前 `resource` 直接相关的受控来源
- 已召回内容应优先以片段级 `artifacts` 传入，而不是透传整篇长文或整段论坛线程
- 若补充论坛或 FAQ 内容，应能让响应层区分其与正式文档的来源差异
- 当前页面已经足够回答问题时，不应继续扩张召回范围
- 当前阶段建议把 `search_scope` 收口到 `docs`、`wiki`、`faq`、`forum`、`attachments` 这组受控枚举中，不把 `console` 内部状态本身作为本任务的检索 scope
- 当前最小回归样本已要求召回 artifact 带 `metadata.source_type`、`metadata.page_slug`、`metadata.fragment_id`、`metadata.retrieval_rank`、`metadata.is_official`
- `primary` artifact 应与当前 `resource.slug` 对齐；`faq` / `forum` 仅作为 supporting 或 reference 补充来源，不应抢占主证据位置
- 若接入外部 `candidate_response_record`，当前最小回归还会校验 `sample_id`、`request_id`、`input_record` 摘要与样本请求对齐
- 若外部记录来自真实候选输出快照，当前建议补 `capture_metadata.capture_origin`、`capture_metadata.collection_batch` 与 `capture_metadata.tags`
- 负例回放继续复用与正例相同的响应校验逻辑；当前允许将真实 captured record 跨样本回放，以验证 record 对齐与响应规则会共同拒绝错配输出

## 统一输出抽象

建议统一响应对象命名为 `CopilotResponse`。

```json
{
  "schema_version": 1,
  "status": "ok",
  "project": "radishflow",
  "task": "explain_diagnostics",
  "summary": "string",
  "answers": [],
  "issues": [],
  "proposed_actions": [],
  "citations": [],
  "confidence": 0.0,
  "risk_level": "medium",
  "requires_confirmation": true
}
```

### 顶层字段建议

- `status`：`ok` / `partial` / `failed`
- `summary`：面向用户的简短说明
- `answers`：适合直接展示的解释或回答片段
- `issues`：发现的问题列表
- `proposed_actions`：候选动作
- `citations`：证据、来源或输入引用
- `confidence`：整体置信度
- `risk_level`：`low` / `medium` / `high`
- `requires_confirmation`：是否必须人工确认

## Candidate Action 抽象

当前建议把高风险输出统一压成候选动作：

```json
{
  "kind": "candidate_edit",
  "title": "补充流股规格",
  "target": {
    "type": "stream",
    "id": "stream-1"
  },
  "rationale": "当前诊断显示该流股缺失关键规格。",
  "patch": {},
  "risk_level": "medium",
  "requires_confirmation": true
}
```

说明：

- `patch` 是候选提案，不是直接执行的命令
- `kind` 可以按项目扩展，但都必须经过项目适配层确认

对于 `RadishFlow suggest_ghost_completion`，当前建议新增 `ghost_completion` 这类候选动作：

```json
{
  "kind": "ghost_completion",
  "title": "补全 FlashDrum 的 vapor outlet ghost 连线",
  "target": {
    "type": "unit_port",
    "unit_id": "U-12",
    "port_key": "vapor_outlet"
  },
  "rationale": "当前 canonical port 尚未连接，且本地规则已提供合法 ghost 候选。",
  "patch": {
    "ghost_kind": "ghost_connection",
    "candidate_ref": "cand-vapor-stub",
    "ghost_stream_name": "V-12"
  },
  "preview": {
    "ghost_color": "gray",
    "accept_key": "Tab",
    "render_priority": 1
  },
  "apply": {
    "command_kind": "accept_ghost_completion",
    "payload": {
      "candidate_ref": "cand-vapor-stub"
    }
  },
  "risk_level": "low",
  "requires_confirmation": false
}
```

## 当前推荐任务枚举

### `RadishFlow`

- `explain_diagnostics`
- `suggest_flowsheet_edits`
- `suggest_ghost_completion`
- `summarize_selection`
- `explain_control_plane_state`
- `inspect_canvas_snapshot`

### `Radish`

- `answer_docs_question`
- `summarize_doc_or_thread`
- `suggest_forum_metadata`
- `explain_console_capability`
- `interpret_attachment`

## 禁止透传与脱敏要求

以下内容当前不应进入 `CopilotRequest`：

- `access_token`
- `refresh_token`
- cookie
- `credential_handle`
- 本机安全存储引用
- 未裁剪的 auth cache 索引原文
- 本地绝对密钥路径和证书密码

补充约定：

- `RadishFlow` 控制面相关只传状态摘要、manifest 摘要和错误信息
- `Radish` 权限相关只传角色 / 权限键摘要和当前页面语义，不传安全凭据

## 关键边界

- `RadishMind` 只返回建议、解释或结构化候选动作
- 最终真相源仍由上层项目维护
- 高风险建议必须要求人工确认
- 若模型能力不足，应允许退化到检索、规则或模板响应
- 协议的统一对象是“外部智能层”，不是两个项目的内部 DTO

## 当前推荐原则

- 先统一协议，再分别做项目适配
- 先支持 `RadishFlow`，再让 `Radish` 逐步接入
- 先做可消费的 JSON，再谈复杂自治代理
- 不把 `RadishFlow` 和 `Radish` 强行拉成同一业务字段集合

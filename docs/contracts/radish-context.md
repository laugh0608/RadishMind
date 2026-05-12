# Radish 上下文契约

更新时间：2026-05-12

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

# 已保存 Workflow 草案派生（开发 / 测试态）v1

更新时间：2026-07-27

状态：`saved_workflow_draft_derivation_dev_test_v1_completed`

## 功能目标

内部开发者应能从当前活动应用中的一个精确已保存 Workflow 草案版本派生新草案，复用已有图结构、节点属性、受控布局和执行档案，再独立编辑、校验、保存和审查。

派生不是覆盖、复制存储记录或发布动作。源草案保持不变，新草案取得独立 `draft_id`，初始版本固定为 `0`，首次保存继续经过既有 workspace membership、owner、原子版本比较和 Saved Draft repository。

## 用户流程

1. 用户从 Saved Drafts 恢复草案，或完成当前草案保存。
2. 只有活动草案没有未保存修改、没有冲突且 consumer 指向精确已保存版本时，`派生新草案` 才可用。
3. 派生动作在浏览器内深复制允许字段，生成不与当前工作区草案冲突的新 `draft_id`。
4. 新草案保留 application、workflow definition、base definition version、节点、边、layout、风险、blocked capability 和 execution profile。
5. 新草案只新增脱敏来源引用 `source_draft_id` 与 `source_draft_version`；不复制源请求 / 审计引用作为新动作证据。
6. 用户继续编辑并通过既有 Validate / Save / Review Handoff；首次保存使用 `expected_draft_version=0`。
7. 保存并恢复后，来源引用仍可审查，但不自动建立版本合并、同步或继承关系。

## 数据边界

`saved_workflow_draft.v1.additional_fields` 新增可选 `derivation_v1`：

```json
{
  "derivation_v1": {
    "version": 1,
    "source_kind": "saved_workflow_draft",
    "source_draft_id": "draft_source",
    "source_draft_version": 3
  }
}
```

约束：

- `source_draft_id` 必须是非空脱敏引用，不能等于派生草案自身 `draft_id`。
- `source_draft_version` 必须大于等于 `1`。
- 服务端归一化非法结构时删除 `derivation_v1`，不从名称、时间或图摘要猜测来源。
- 来源引用不携带 source payload、digest、token、actor、membership、request body、run input / output 或 provider 响应。
- 派生链只记录直接父草案，不递归复制祖先链。

## 实施批次

唯一实施任务卡见 [已保存 Workflow 草案派生（开发 / 测试态）v1 实施任务卡](../../task-cards/saved-workflow-draft-derivation-dev-test-v1-plan.md)。

本批交付：

- 纯派生 builder、稳定短 ID 和深复制边界。
- Draft Designer 的显式派生入口、来源摘要和不可用条件。
- `derivation_v1` Web 编解码与 Go 归一化。
- Web 单元测试、Go 相邻单元测试、Web build 和仓库门禁。

实施结果：纯派生 builder、Draft Designer 入口、直接来源摘要、`derivation_v1` Web 编解码和 Go 归一化均已完成。派生只在 `saved_dev_record + exact version + clean local state` 下开放，新草案以版本 `0` 进入既有保存链。

## 验收

- 未保存、正在操作、版本冲突或版本为 `0` 时不能派生。
- 派生草案修改节点、边、布局或数组时不改变源草案。
- 同一来源连续派生时 ID 稳定递增且不与现有本地草案冲突。
- 新草案初始 consumer 版本为 `0`，首次保存不会覆盖源草案。
- 保存 / 读取往返后 `derivation_v1` 保留直接来源；非法来源结构不被持久化。
- 派生不产生后端请求、provider 调用、工具调用、确认、业务写回或 replay。

## 停止线

- 不实现草案合并、自动同步、分支图、祖先链、批量复制、跨 application 派生或跨 workspace 派生。
- 不复制源 owner、actor、request / audit ref、运行记录、评测、发布候选、激活状态或密钥。
- 不绕过既有 Saved Draft mutation authorization、owner scope、CAS、repository selector 或 no-fallback 规则。
- 不新增 API、repository、migration、业务真相源或专项 checker。
- 不把开发 / 测试态派生解释为 production OIDC、membership adapter、生产存储或公开 API 就绪。

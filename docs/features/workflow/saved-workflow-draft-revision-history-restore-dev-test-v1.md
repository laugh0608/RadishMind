# 已保存 Workflow 草案修订历史、版本比较与显式恢复（开发 / 测试态）v1

更新时间：2026-07-27

状态：`saved_workflow_draft_revision_history_restore_dev_test_v1_completed`

## 功能目标

内部开发者应能在同一 application、owner 与 draft scope 内查看已保存 Workflow 草案的不可变修订历史，读取任一精确版本，在浏览器内比较结构化差异，并把历史版本显式恢复为新的当前版本。

本专题扩展既有 Saved Draft owner，不创建第二套草案真相源。当前草案记录继续承担最新可编辑状态，修订记录承担 append-only 历史证据；恢复不会倒退版本号、覆盖历史记录或移动隐藏 pointer。

## 用户流程

1. 每次成功保存草案时，当前记录 CAS 更新与不可变 revision 追加在同一原子操作中完成。
2. 用户从 Draft Designer 打开当前草案的修订历史，按版本倒序分页查看摘要。
3. 用户选择任一版本读取精确 sanitized snapshot，并与当前版本比较节点、边、布局、引用、校验状态和来源 metadata。
4. 用户可以显式请求恢复历史版本；请求必须携带当前精确版本并重新通过现行校验策略。
5. 恢复成功时创建 `current_version + 1`，revision kind 为 `restored`，并记录直接 `restored_from_version`。
6. 原历史版本与恢复前的当前版本继续不可变可读；服务重启后版本序列和当前记录保持一致。

## 数据与存储边界

修订记录 schema 固定为 `saved_workflow_draft_revision.v1`，包含：

- tenant、workspace、application、owner 与 `draft_id` scope；
- 正整数 `draft_version`；
- `revision_kind=saved|restored|backfilled_current`；
- `restored_from_version`，仅恢复 revision 为正整数，其余为 `0`；
- 精确 `saved_workflow_draft.v1` sanitized snapshot。

SQLite 与 PostgreSQL 新增 `saved_workflow_draft_revisions` append-only 表。既有数据库升级时，只把当前主表快照回填为 `backfilled_current`；迁移前已经被覆盖的历史版本不可恢复，不根据时间、审计引用或版本号猜测内容。

新保存必须满足：

- 当前主表 CAS 与 revision insert 同一事务；
- revision 主键为精确 scope + `draft_id` + `draft_version`；
- revision insert 失败时当前主表不得提交；
- runtime 只允许读取和追加 revision，不提供 update / delete API；
- stored snapshot 继续经过现有 forbidden-field、大小预算、scope 和 schema contract。

## API 与权限

开发测试态新增：

- `GET /v1/user-workspace/workflow-drafts/{draft_id}/revisions`
- `GET /v1/user-workspace/workflow-drafts/{draft_id}/revisions/{version}`
- `POST /v1/user-workspace/workflow-drafts/{draft_id}/revisions/{version}/restore`

历史列表与读取要求 `workflow_drafts:read`。恢复由一次 membership decision 原子要求 `workflow_drafts:read + workflow_drafts:write`，并继续受显式 Saved Draft dev write gate 约束。

分页 cursor 只编码当前 draft scope 内的下一版本位置，不携带 payload。恢复 body 只允许 `workspace_id`、`application_id` 与 `expected_current_draft_version`。

## 比较语义

版本比较在 Web 纯函数中完成，不创建服务端 diff owner。比较结果按稳定类别输出：

- draft metadata、base definition 与 execution profile；
- 节点新增、删除和字段变化；
- 边新增、删除和条件变化；
- designer layout 变化；
- provider / tool / RAG refs；
- validation、blocked capability、derivation 与 restore provenance。

比较结果只用于审查，不自动合并、不生成 patch、不改变当前草案。

## 实施批次

唯一实施入口见[实施任务卡](../../task-cards/saved-workflow-draft-revision-history-restore-dev-test-v1-plan.md)。

1. 批次 A：设计、领域模型、失败语义与内存 owner。
2. 批次 B：SQLite / PostgreSQL migration、原子保存与重启恢复。
3. 批次 C：分页读取、精确版本与显式恢复 API / authorization。
4. 批次 D：Web 修订历史、结构化比较与恢复确认。
5. 批次 E：双数据库、并发、浏览器、隐私和仓库门禁收口。

## 验收

- 连续保存 `v1 → v2 → v3` 后三个版本均可精确读取。
- 历史列表稳定倒序分页，非法或跨 draft cursor 失败关闭。
- 任一版本 payload 损坏、scope 漂移或 schema 不兼容时不返回部分数据。
- 恢复 `v1` 在当前 `v3` 上生成 `v4`；`v2`、`v3` 保持可读。
- 并发恢复只有一个 CAS 成功，失败方返回真实当前版本。
- SQLite 与 PostgreSQL 在重启、migration 回填和 no-fallback 场景下语义一致。
- Web 比较不修改任一 snapshot，未保存编辑、冲突和 pending operation 时恢复保持禁用。
- revision 不包含 token、credential、header、原始运行输入输出、provider raw payload 或业务写回数据。

## 完成结果

- Memory、SQLite 与 PostgreSQL 已统一实现不可变 revision owner；每次普通保存或恢复都把当前记录 CAS 与 revision append 放在同一原子边界内。
- `0002_saved_workflow_draft_revisions` 会为既有数据库回填唯一可证明的当前快照，并保留 `backfilled_current` 来源；不推测迁移前已覆盖历史。
- 历史列表使用绑定 draft 与页大小的 opaque cursor，精确版本读取和恢复均复验 tenant、workspace、application、owner、schema 与 snapshot。
- 恢复历史版本会重新执行现行草案校验，并以 `restored` revision 创建严格递增的新版本；并发或过期 expected version 保持真实 CAS 冲突。
- Web 已提供懒加载历史面板、结构化版本比较、未保存编辑警告与两步恢复确认；恢复成功后明确进入新当前版本。
- 恢复路由在 body 解码和 owner 调用前原子要求 `workflow_drafts:read + workflow_drafts:write`；PostgreSQL runtime role 对 revision 表仅保留 `SELECT / INSERT`，明确拒绝 `UPDATE / DELETE`。
- Platform 定向测试、Web 254 项测试与 production build、真实 PostgreSQL migration / restore / restart / role denial 集成链均已通过。

## 停止线

- 不实现自动保存、自动恢复、自动合并、三方合并、分支图或祖先图。
- 不允许修改、删除、压缩或重写 revision。
- 不实现跨 application / workspace / owner 恢复。
- 不从历史版本直接运行、发布、确认或写回业务真相源。
- 不启用 replay、resume、schedule、agent loop、production repository、真实 OIDC、quota 或 billing。
- 不为普通比较 UI 新增独立 checker；新增 schema、API 和存储风险由任务卡、相邻测试与聚合门禁承载。

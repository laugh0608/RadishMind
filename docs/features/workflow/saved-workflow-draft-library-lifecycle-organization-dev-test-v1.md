# 已保存 Workflow 草案库生命周期与组织（开发 / 测试态）v1

更新时间：2026-08-08

状态：`saved_workflow_draft_library_lifecycle_organization_dev_test_v1_completed`

实施准入：`implementation_task_card_closed`

当前实现进度：批次 A 至 E 已完成领域类型、三种 store lifecycle owner、严格 cursor、keyset 分页 / 筛选、双版本并发、双数据库 `0003`、原子 transition / event、HTTP list / archive / unarchive、`workflow_drafts:archive` 单次 membership 判定、保存 / 恢复双版本请求、相邻派生 / 晋级 / 执行 active lifecycle 资格，以及 Web 活动 / 归档独立查询、筛选、加载更多、只读归档审查、两步 archive、显式 unarchive、迟到响应隔离和“打开 / 恢复历史版本”术语分离。超过 `200` 条的三存储同时间记录矩阵、真实 PostgreSQL、SQLite 本地产品启动与浏览器连续链均已复验，专题关闭。

## 功能目标

内部开发者应能在当前 tenant、workspace、application 与 owner 作用域内，把不断增长的已保存 Workflow 草案组织为活动草案和归档草案，通过稳定分页与明确筛选找到目标草案，并以可逆、可审计的方式归档或解除归档。

本专题扩展既有 Saved Draft owner，不创建第二套草案真相源。草案内容、不可变修订、生命周期状态和 Workflow Definition 是四类不同责任：

- 当前草案记录继续承担最新已保存内容；
- revision 继续承担不可变内容历史；
- lifecycle 只承担草案库中的活动 / 归档组织状态；
- 已晋级的不可变 Workflow Definition 继续承担运行权威。

归档不是删除、内容保存、内容修订、发布、停用 Definition 或取消 Run。生命周期变化不得借用 `draft_version` 伪装成图内容变化，也不得改写既有 revision。

## 设计时审计与当前治理结果

### 服务与存储

- 设计时 `ListWorkflowDraftsRequest` 为空、SQLite / PostgreSQL 固定最多读取 `200` 条且 HTTP 无截断信号；批次 A 至 C 已把它替换为三种 store 一致的 lifecycle、limit、opaque cursor、name prefix、validation 与 provenance 查询。
- memory、SQLite 与 PostgreSQL 现统一使用 `library_updated_at DESC, draft_id ASC` keyset 排序和 `limit + 1`，HTTP envelope 已返回 `next_cursor` 与 `has_more`，不再保留固定 `200` 条静默截断。
- `draft_status` 继续只承担 `valid_for_review`、`invalid_draft`、`blocked_capability`、`schema_unsupported` 校验状态；活动 / 归档由独立 lifecycle metadata、CAS 与 event owner 承担。
- 既有内容 CAS、scope / owner 隔离、no fallback、schema preflight、数据库升级与重启恢复保持不变；双数据库 `0003` 已补齐 projection、索引、原子 transition / event 与 runtime role 约束。
- revision 列表已经具备绑定 draft 与页大小的 opaque cursor，可复用 cursor 校验思路，但不能直接复用只按整数版本倒序的 cursor 结构。

### Web 与相邻流程

- Web consumer 已严格消费 lifecycle metadata、服务端筛选、`next_cursor` / `has_more`、双版本 failure envelope 和 archive / unarchive；活动 / 归档拥有独立结果、cursor 与加载状态，分页合并按 `draft_id` 防御重复。
- UI 已把“读取一条已保存草案并进入 Draft Designer”统一改称“打开草案”，归档态使用“只读审查”；“恢复”只保留给 revision restore。
- workspace、application、owner、lifecycle 或 filter 变化会清理 cursor、选择与 pending，并用 generation + scope key 拒绝迟到响应；解除归档后刷新两侧列表但不自动打开，必须重新读取当前双版本。
- 校验响应不具备 lifecycle 权威，Web 会保留调用前精确 lifecycle 版本与状态，避免已保存活动草案在 validate 后被错误锁成未知态。
- 已保存草案派生首次保存已复验来源精确内容版本、生命周期版本与 active 状态；无未保存修改和无 pending operation 的既有要求保持不变。
- revision 历史的 list / read / compare 继续可读；restore request 已携带 expected lifecycle version，并在归档状态失败关闭。
- Workflow Definition candidate、Saved Draft 受控执行、RAG retrieval 和 HTTP Tool 已在各自业务 owner 内统一复验 active lifecycle。
- 已晋级 Definition 和既有 Run 都保留 source draft provenance；来源草案归档后，这些不可变证据不得失效或被级联修改。

### S3 UI 消费

- S3 Workflow Designer 只消费本专题已经成立的活动 / 归档列表、严格 cursor、筛选、双版本 lifecycle 和精确打开，不新增草案库 API、schema、repository 或生命周期动作。
- 活动草案进入可编辑 Designer；归档草案进入只读审查。解除归档后必须从活动草案库重新精确打开，旧浏览器快照不得自动恢复为可编辑状态。
- Saved Draft Library 作为 `B` 级产品面复用 S1 列表和 S3 Designer 上下文模式；普通状态、筛选和分页不派生独立 Pencil 画板或专项 checker。

## 目标用户与核心场景

目标用户是正在一个 application 内设计、审查和晋级 Workflow 的内部开发者。

核心场景：

1. 默认查看最近变化的活动草案，不混入归档草案。
2. 使用名称、validation state 或直接 provenance 缩小结果集。
3. 通过稳定 cursor 继续加载，不依赖一次性固定数量窗口。
4. 打开活动草案进入现有编辑、校验、保存、派生与审查路径。
5. 显式归档暂时不再编辑的草案，使其离开默认工作集。
6. 在归档视图继续查看 sanitized summary、当前内容、不可变历史和版本比较。
7. 显式解除归档后，再重新进入既有 mutation、晋级和受控运行资格检查。

## 草案库信息架构

Saved Draft Library 继续位于 User Workspace 的 Workflow 区域，不在本批创建新的一级产品面。

首版信息架构：

- `活动草案`：默认视图，显示 `lifecycle_state=active`。
- `归档草案`：独立视图，只显示 `lifecycle_state=archived`。
- 筛选：名称、validation state、provenance kind。
- 列表项：名称、`draft_id`、内容版本、生命周期版本、validation state、provenance、内容更新时间、草案库更新时间、归档时间和受控操作。
- 详情入口：
  - 活动草案使用“打开草案”，进入可编辑 Draft Designer；
  - 归档草案使用“只读审查”，不得生成可保存的 active editor state；
  - revision 面板继续使用“查看版本”和“恢复此版本”，但归档状态下恢复按钮禁用并解释原因。
- 分页：首屏和“加载更多”均消费服务端 cursor，不在 Web 对截断窗口进行伪分页。

workspace、application 或 owner 变化时，Web 必须清理列表、cursor、筛选、选择项、pending transition 与迟到响应。活动 / 归档切换使用各自独立的结果和 cursor，不拼接不同查询。

## 生命周期模型

### 当前状态

当前草案记录增加以下生命周期 metadata：

| 字段 | 语义 |
| --- | --- |
| `lifecycle_state` | `active` 或 `archived` |
| `lifecycle_version` | 正整数生命周期 CAS；既有记录迁移后为 `1` |
| `archived_at` | 活动态为 `null`，归档态为 RFC 3339 UTC 时间 |
| `library_updated_at` | 最近一次内容保存或生命周期变化时间，用于草案库排序 |
| `lifecycle_updated_by_actor_ref` | 最近一次 lifecycle transition 的脱敏 actor ref；未发生 transition 的迁移记录为空 |

`draft_version` 只在内容保存或 revision restore 成功时递增。`lifecycle_version` 只在 archive / unarchive 成功时递增。普通读取、列表、比较和校验不改变任一版本。

### 追加式生命周期事件

SQLite 与 PostgreSQL 增加 `saved_workflow_draft_lifecycle_events`：

- 主键包含完整 scope、`draft_id` 与 `lifecycle_version`；
- 记录 `from_state`、`to_state`、`transition_kind=archived|unarchived`；
- 记录发生时间、actor ref、request id 与 audit ref；
- 不包含草案 payload、名称、说明、节点、边、运行输入输出或 provider 数据；
- runtime role 只允许 `SELECT / INSERT`，拒绝 `UPDATE / DELETE`；
- lifecycle 当前记录更新与 event insert 必须处于同一原子操作。

既有记录迁移为 `active + lifecycle_version=1`，但不伪造从未发生过的 lifecycle event。

这一设计保留完整 archive / unarchive 审计，同时避免污染内容 revision。

### 内容与生命周期并发

会改变草案内容或从草案创建下游权威记录的请求，必须携带或在同一 owner 内复验精确 `draft_version` 与 `lifecycle_version`。

- save、revision restore 和 candidate creation 必须要求当前 lifecycle 为 `active`。
- archive / unarchive 必须携带 `expected_draft_version` 与 `expected_lifecycle_version`。
- archive 与 save / restore / promotion 并发时，只允许首先完成精确 CAS 的一方成功；另一方返回当前内容版本、生命周期版本和状态。
- archive commit 之后发起的新执行、派生保存、恢复或晋级请求必须失败关闭。
- 已在 archive commit 前完成资格判定并进入既有原子 claim 的操作不被本功能隐式取消；本功能不增加取消、补偿或进程协调器。

## 列表、分页与筛选契约

既有 `GET /v1/user-workspace/workflow-drafts` 扩展为服务端集合查询：

- `workspace_id`、`application_id`：继续必填并与 active workspace / membership binding 一致；
- `lifecycle_state`：可选，默认 `active`；
- `limit`：默认 `25`，最小 `1`，最大 `100`；
- `cursor`：可选 opaque cursor；
- `name_prefix`：可选，trim 后按 UTF-8 精确前缀匹配，最大 `80` 字符；
- `validation_state`：可选 allowlist；
- `provenance_kind`：可选 `unversioned|workflow_definition|saved_draft_derivation`。

provenance 只依据当前记录的直接事实确定：

1. 合法 `derivation_v1` 优先归类为 `saved_draft_derivation`；
2. 否则 `base_definition_version > 0` 归类为 `workflow_definition`；
3. 其余归类为 `unversioned`。

不递归读取祖先，不根据名称、时间、ID 或 revision 猜测 provenance。

数据库增加可校验的 `draft_name` 与 `provenance_kind` 投影列，避免 memory、SQLite 与 PostgreSQL 分别从 JSON 实现不同筛选语义。写入时投影与 sanitized payload 必须一致；不一致按 store contract mismatch 失败关闭。

排序固定为 `library_updated_at DESC, draft_id ASC`。下一页谓词固定为：

```text
library_updated_at < last_library_updated_at
OR (library_updated_at = last_library_updated_at AND draft_id > last_draft_id)
```

cursor 必须绑定 schema version、tenant、workspace、application、owner、lifecycle、全部筛选、limit 和最后排序键。cursor 不携带名称、payload、actor、request 或 audit 内容。任何 scope、筛选、limit、版本或结构漂移都返回 `draft_list_cursor_invalid`，不回退第一页。

集合查询使用 `limit + 1` 判断 `next_cursor`。并发变化下不承诺跨请求 snapshot isolation，但必须保证未变化记录不会因同一 cursor 重复出现；新保存或新解除归档而排到 cursor 之前的记录只在用户刷新后出现。

## API 与权限

开发 / 测试态新增：

- `POST /v1/user-workspace/workflow-drafts/{draft_id}/archive`
- `POST /v1/user-workspace/workflow-drafts/{draft_id}/unarchive`

请求 body 只允许：

```json
{
  "workspace_id": "workspace_demo",
  "application_id": "app_demo",
  "expected_draft_version": 3,
  "expected_lifecycle_version": 1
}
```

响应只返回 lifecycle metadata、当前内容版本、failure code、request id 与 audit ref，不返回完整草案 payload。

权限：

- list、当前记录只读、revision list / read：`workflow_drafts:read`；
- archive / unarchive：一次 membership decision 原子要求 `workflow_drafts:read + workflow_drafts:archive`；
- save 与 revision restore：继续要求既有 read / write 权限，并增加 active lifecycle 资格；
- candidate creation、执行和其它组合入口继续使用各自既有权限，同时在业务 owner 内复验 active lifecycle；
- 新增 `workflow_drafts:archive` 只表示草案生命周期转换，不授予内容写入、发布、运行或跨作用域能力。

dev header、signed-test membership 和 production OIDC 的边界保持不变。

## 归档后操作资格

| 操作 | 活动草案 | 归档草案 |
| --- | --- | --- |
| 列表、sanitized summary | 允许 | 允许，仅归档视图 |
| 当前记录读取 | 允许，可进入编辑器 | 允许，只读审查 |
| revision list / read / compare | 允许 | 允许 |
| 内容校验结果审查 | 允许 | 允许，只读 |
| save | 允许，复用既有门禁 | 拒绝 `draft_archived` |
| 从已保存草案派生 | 允许，要求精确双版本 | 拒绝 `draft_archived` |
| revision restore | 允许，创建新内容版本 | 拒绝 `draft_archived` |
| 创建 Workflow Definition candidate | 允许，复验精确双版本 | 拒绝 `draft_archived` |
| 直接 Saved Draft 受控执行 | 允许，复用既有执行门禁 | 拒绝 `draft_archived` |
| archive | 允许，双版本 CAS | 拒绝 lifecycle conflict |
| unarchive | 拒绝 lifecycle conflict | 允许，双版本 CAS |
| 已晋级 Definition 执行 | 不依赖当前草案生命周期 | 不依赖当前草案生命周期 |
| 既有 Run / candidate / version 审查 | 允许 | 允许，不级联改写 |

归档不撤销、隐藏或重写既有 candidate、Definition version、activation pointer、Run、evaluation、request history 或 audit reference。

## 失败语义

新增稳定失败码：

- `draft_archived`：操作要求活动草案，但当前已归档；
- `draft_lifecycle_version_conflict`：expected lifecycle version 漂移；
- `draft_lifecycle_state_conflict`：archive / unarchive 的起始状态不匹配；
- `draft_list_cursor_invalid`：cursor 与 scope、筛选、limit 或 schema 不一致；
- `draft_list_filter_invalid`：筛选不在 allowlist、超长或组合非法；
- `draft_lifecycle_event_write_failed`：当前状态与 event 不能原子提交；
- `draft_lifecycle_store_contract_mismatch`：投影列、payload 与 lifecycle metadata 不一致。

既有 `draft_version_conflict`、scope、membership、schema、store unavailable 和 no fallback 语义保持不变。所有失败响应必须返回 sanitized metadata；失败不得产生部分 lifecycle update、event、revision、candidate、Run 或外部调用。

## 隐私与审计边界

- 列表继续只返回 sanitized summary，不返回完整节点、边条件、secret、token、header、原始输入输出或 provider raw payload。
- lifecycle event 只保留 scope、状态变化和脱敏审计引用。
- cursor 不编码名称、actor、request / audit ref 或 payload。
- 归档确认界面不得展示 secret、完整 claim、运行正文或 provider 响应。
- archive / unarchive 自身的 provider、tool、confirmation、Run、publish、business writeback、replay 与 resume 副作用必须为 `0`。

## 实施拆分与准入结论

本设计明确需要新增 API、permission、schema、migration、集合查询契约和执行资格，因此已创建[唯一高风险实施任务卡](../../task-cards/saved-workflow-draft-library-lifecycle-organization-dev-test-v1-plan.md)，不允许从 Web 按钮开始零散实现。

唯一任务卡按以下批次承接：

1. 批次 A：领域类型、生命周期 CAS、事件、分页 / 筛选契约和 memory owner。
2. 批次 B：SQLite / PostgreSQL `0003` migration、投影列、索引、原子 transition 与 runtime role。
3. 批次 C：HTTP list / archive / unarchive、membership permission 和相邻 save / restore / promotion / execution 资格。
4. 批次 D：Saved Draft Library 活动 / 归档视图、筛选、加载更多、只读归档审查和术语修正。
5. 批次 E：已完成双数据库、升级、重启、并发、浏览器连续链、隐私与仓库门禁收口。

不为普通布局、文案或只读列表另建平行任务卡、fixture 或 checker。专项证据只服务新增 API、schema、permission 和高风险操作资格。

## 开发 / 测试态验收

- 既有数据库升级后全部草案为 `active + lifecycle_version=1`，内容版本和 revision 不变化，不伪造 lifecycle event。
- 活动 / 归档列表按相同 scope、筛选和 keyset 语义在 memory、SQLite、PostgreSQL 中一致。
- 超过 `200` 条记录可通过 cursor 完整遍历，不再静默截断；分页中无重复的未变化记录。
- name、validation 与 provenance 筛选在三种 store 返回相同结果。
- archive / unarchive 在重启后保持状态与 event；并发同 transition 只有一个成功。
- archive 与 save / restore / promotion 并发时只有精确双版本 CAS 一方成功。
- 归档后 list / read / revision / compare 可用，save / derive / restore / promotion / direct execution 失败关闭。
- 解除归档后必须重新读取当前双版本，不能自动重放归档前 pending mutation。
- workspace / application / owner 切换清理列表、cursor、选择和 pending transition；迟到响应不能污染新作用域。
- 已晋级 Definition 与既有 Run 在来源草案归档后仍可审查，Definition-bound execution 不被错误阻断。
- archive / unarchive 失败路径的 provider、tool、confirmation、Run、publish 和业务写入计数均为 `0`。
- migration、cursor、event 与响应通过 forbidden-field 和敏感信息扫描。

批次 E 已使用 `231` 条相同时间记录复验三种 store 的完整分页、筛选和稳定顺序；真实浏览器通过 SQLite 本地产品链复验服务重启、`25 + 1` 分页、组合筛选、workspace 切换隔离、归档只读审查、解除归档和双版本重新取得资格。完整 Go、race、`go vet`、Web tests / build、真实 PostgreSQL 与仓库门禁共同构成关闭证据。

## 停止线

- 不实现永久删除、批量清理、批量归档、自动归档、保留期执行器或垃圾回收。
- 不实现自动保存、自动恢复、自动合并、三方合并、分支图、祖先图或跨 application / workspace / owner 移动。
- 不从归档草案直接创建 candidate、运行、发布、确认或写回业务真相源。
- 不因草案归档自动停用 Definition、取消 Run、撤销 API key、修改应用配置或删除历史证据。
- 不创建第二套草案 owner、搜索服务、全文索引、通用资源生命周期框架或跨产品面抽象。
- 不启用 production repository、真实 OIDC、production membership adapter、quota、billing 或公开生产 API。
- 不恢复已冻结的 storage adapter readiness 链，也不新增同层 gate-only 文档链。

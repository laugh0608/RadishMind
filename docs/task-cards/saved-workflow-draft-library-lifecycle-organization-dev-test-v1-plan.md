# 已保存 Workflow 草案库生命周期与组织（开发 / 测试态）v1 实施任务卡

更新时间：2026-07-28

状态：`in_progress`

当前入口：`batch_b_ready`

## 目标

在既有 Saved Draft owner 内交付活动 / 归档草案库、稳定服务端分页、受控筛选、可逆 archive / unarchive 和归档后操作资格，并保持内容版本、不可变 revision、生命周期审计、Workflow Definition 与 Run 权威边界一致。

权威功能设计见[已保存 Workflow 草案库生命周期与组织（开发 / 测试态）v1](../features/workflow/saved-workflow-draft-library-lifecycle-organization-dev-test-v1.md)。

本任务卡是该功能唯一实现入口，不派生平行 lifecycle、分页、筛选、archive permission、数据库或 Web 任务卡。

## 前置条件与已知事实

- Memory、SQLite 与 PostgreSQL 已共享 Saved Draft domain、repository adapter、scope / owner、内容 CAS、schema preflight 和 no fallback。
- 当前 `ListWorkflowDraftsRequest` 无分页 / 筛选；Memory 返回全部，SQLite / PostgreSQL 固定最多 `200` 条，HTTP 无 `next_cursor`。
- `draft_status` 是 validation state，不是生命周期。
- revision 已有 append-only owner、稳定版本 cursor、精确读取、结构化比较和显式恢复。
- 已保存草案派生、Workflow Definition candidate、Saved Draft execution、RAG retrieval 和 HTTP Tool 都会消费精确草案，必须统一增加 active lifecycle 资格。
- Workspace-scoped Mutation Authorization 已提供唯一 membership provider；新增 mutation 必须复用它。
- 生产 repository、真实 OIDC、production membership adapter 和公开生产 API 仍关闭。

## 不变量

1. `draft_version` 只表示内容版本，archive / unarchive 不递增它，也不追加内容 revision。
2. `lifecycle_version` 只表示生命周期转换；既有记录迁移为 `active + v1`。
3. lifecycle 当前记录更新与 append-only event insert 必须原子提交。
4. archive / unarchive 同时校验 expected 内容版本和 expected lifecycle 版本。
5. 归档后只保留 list / read / revision / compare 与 unarchive；save、derive、restore、promotion 和 direct execution 失败关闭。
6. 来源草案归档不修改既有 candidate、Definition、activation pointer 或 Run。
7. cursor 绑定完整 scope、lifecycle、筛选、limit 与排序锚点；漂移不回退第一页。
8. archive / unarchive 不产生 provider、tool、confirmation、Run、publish、业务写回、replay 或 resume。
9. 三种 store 的筛选、排序、分页、失败码和 projection contract 必须一致。
10. 不创建第二套草案 owner、通用生命周期框架或全文搜索服务。

## 批次 A：领域、集合契约与 Memory owner

状态：`completed`

完成证据：

- Saved Draft domain 已增加 lifecycle、library、provenance、transition / event 与稳定 failure code。
- Memory owner 已实现双版本 CAS、当前状态与 event 原子提交、活动 / 归档列表和三类组合筛选。
- cursor 已绑定完整 scope、筛选、limit 与排序锚点，并严格拒绝未知字段、结构 / schema 漂移和锚点篡改。
- `231` 条同时间记录已通过三页完整遍历；archive 与 save / restore 并发均只有一方成功。
- 保存、修订恢复、Workflow executor、RAG execution、HTTP Tool planning 与 Definition candidate 已统一复验 active lifecycle。

### 实现

- 在 Saved Draft domain 增加：
  - `lifecycle_state=active|archived`
  - `lifecycle_version`
  - `archived_at`
  - `library_updated_at`
  - `lifecycle_updated_by_actor_ref`
  - `provenance_kind`
- 增加 lifecycle transition request / result / event 和稳定 failure code。
- 扩展 `ListWorkflowDraftsRequest`，支持：
  - lifecycle，默认 `active`
  - `limit`，默认 `25`、最大 `100`
  - opaque cursor
  - `name_prefix`
  - `validation_state`
  - `provenance_kind`
- Memory store 使用 `library_updated_at DESC, draft_id ASC` keyset 分页和 `limit + 1`。
- 实现 Memory archive / unarchive 双版本 CAS 与 append-only event。
- save、revision restore 和直接消费草案的领域入口复验 active lifecycle。
- 既有内存 fixture 默认补齐 `active + lifecycle_version=1`，不创建伪 event。

### 验证

- 默认 active、显式 archived、三筛选及组合筛选。
- 空页、最大页、超过 `200` 条完整遍历、相同时间的 `draft_id` tie-break。
- cursor 的 scope、lifecycle、filter、limit 和 schema 漂移全部失败关闭。
- archive → read-only → unarchive；重复 archive / unarchive 返回 state conflict。
- save / restore 与 archive 并发只有精确双版本 CAS 一方成功；candidate / promotion 的跨 owner 原子资格留在批次 C。
- 归档后相邻 mutation 和 execution owner 返回 `draft_archived`，副作用计数为 `0`。

### 停止线

- 不修改 SQL、HTTP、Web 或生产配置。
- 不增加删除、批量 lifecycle 或 snapshot isolation。

## 批次 B：SQLite / PostgreSQL schema 与 repository

状态：`ready`

### 实现

- 新增 SQLite 与 PostgreSQL `0003` migration：
  - current table lifecycle / library / projection columns；
  - `saved_workflow_draft_lifecycle_events`；
  - scope + lifecycle + `library_updated_at DESC` + `draft_id ASC` list index；
  - validation / provenance / name prefix 相邻索引。
- 迁移既有记录：
  - lifecycle 固定为 active v1；
  - `library_updated_at=updated_at`；
  - `draft_name` 与 `provenance_kind` 从 sanitized payload 的直接事实回填；
  - 不回填 lifecycle event。
- repository interface 和 query executor 承接 list page 与 lifecycle transition。
- transition 主记录 CAS 与 event insert 使用同一事务。
- runtime role 对 event 表只允许 `SELECT / INSERT`，拒绝 `UPDATE / DELETE`。
- projection 与 payload 不一致时 fail closed，不静默修复。

### 验证

- `0001 → 0002 → 0003`、重复 apply、reviewed rollback / reapply。
- 旧数据库升级保持内容版本、revision、scope、owner 和 payload 不变。
- 三种 store 的 page / filter golden matrix 一致。
- PostgreSQL 与 SQLite 并发 transition、重启恢复、损坏 projection、event insert 失败原子回滚。
- runtime role DML 拒绝和敏感信息扫描。

### 停止线

- 不启用 production repository。
- 不选择新数据库产品、secret resolver 或连接 provider。

## 批次 C：HTTP、权限与相邻操作资格

状态：`blocked_by_batch_b`

### 实现

- 扩展 `GET /v1/user-workspace/workflow-drafts` query 与 envelope，返回 `next_cursor`。
- 新增：
  - `POST /v1/user-workspace/workflow-drafts/{draft_id}/archive`
  - `POST /v1/user-workspace/workflow-drafts/{draft_id}/unarchive`
- mutation body 只允许 workspace、application 和 expected 双版本。
- 新增 `workflow_drafts:archive` permission。
- archive / unarchive 一次 membership decision 原子要求 read + archive。
- save / revision restore request 增加 expected lifecycle version。
- 已保存草案派生首次保存、Workflow Definition candidate、Saved Draft executor、RAG retrieval 和 HTTP Tool 在业务 owner 内复验 active lifecycle。
- HTTP error envelope 返回当前内容版本、生命周期版本与状态，不返回完整 payload。

### 验证

- query allowlist、limit、cursor、filter、path / body / header / active workspace binding。
- missing、stale、cross tenant / workspace / application / owner、OIDC 未接入和 permission denial。
- membership provider 每个组合入口只判定一次；denial 时业务 owner、数据库写、Run、Gateway、provider、tool、network 均为 `0`。
- archive commit 后的新 mutation / execution 失败关闭；archive 前已进入既有原子 claim 的操作不被伪取消。

### 停止线

- 不新增 public production route、API key lifecycle、quota 或 billing。
- 不复用 `workflow_drafts:write` 代替独立 archive permission。

## 批次 D：Saved Draft Library Web

状态：`blocked_by_batch_c`

### 实现

- 把现有 Saved Draft list 组织为活动 / 归档两个独立查询状态。
- 增加 name prefix、validation state、provenance 筛选和“加载更多”。
- 合并分页时按 `draft_id` 防御重复，但不对服务端截断窗口做伪分页。
- 活动草案使用“打开草案”；归档草案使用“只读审查”。
- 把现有 saved record `restore` UI 术语和 handler 改为 `open`；revision restore 保留“恢复历史版本”。
- archive 使用两步确认；unarchive 显式执行，成功后刷新两侧列表但不自动打开。
- 归档草案 revision list / read / compare 可用，restore 禁用并解释 `draft_archived`。
- workspace / application / owner / lifecycle / filter 切换清理 cursor、选择、pending 与迟到响应。

### 验证

- consumer parser 严格校验 lifecycle metadata、next cursor 与 failure envelope。
- active / archived empty、loading、failed、load more、filter reset、迟到响应和 scope switch。
- 归档只读状态不能生成可保存 editor state。
- Web tests、production build 与既有 lazy chunk 预算不回退。

### 停止线

- 不创建第二个 Workflow 一级页面。
- 不实现全文搜索、批量选择、拖拽整理、文件夹或标签系统。

## 批次 E：连续验证与文档收口

状态：`blocked_by_batch_d`

### 产品连续链

1. 旧数据库升级；
2. 创建并连续保存多个草案；
3. 跨 `200` 条稳定分页与组合筛选；
4. 打开活动草案并查看 revision；
5. archive；
6. 只读当前记录、历史与比较；
7. 验证 save / derive / restore / promotion / direct execution 全部失败关闭；
8. 服务重启；
9. unarchive；
10. 重新读取双版本并恢复既有编辑 / 晋级资格。

### 验证

- Memory、SQLite、PostgreSQL 的领域 / HTTP / Web 语义一致。
- 定向 Go tests、race、`go vet`、Web tests / build。
- SQLite 聚合产品链和真实 PostgreSQL migration / role / restart integration。
- 真实浏览器 scope switch、分页、筛选、archive / unarchive 与归档只读链。
- forbidden-field、cursor、event、日志和响应敏感信息扫描。
- `git diff --check`、仓库快速门禁与全量门禁。

### 文档收口

- 更新功能设计、Workflow / task card 入口、当前焦点、路线图、能力矩阵和周志。
- 完成后状态改为 `saved_workflow_draft_library_lifecycle_organization_dev_test_v1_completed`。
- 不派生同层批次 F、readiness refresh 或普通 UI checker。

## 推荐验证命令

```bash
(cd services/platform && go test ./internal/httpapi)
(cd services/platform && go test -race ./internal/httpapi)
(cd services/platform && go vet ./...)
npm --prefix apps/radishmind-web test
npm --prefix apps/radishmind-web run build
./scripts/run-workflow-saved-draft-postgres-dev-test.sh check
git diff --check
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```

真实 PostgreSQL、浏览器 dev server 或其它需要长期驻留的进程按仓库协作约定由开发者显式启动，除非当次任务另有授权。

## 总停止线

- 不实现永久删除、批量清理、自动归档、自动保存、自动恢复、自动合并、分支图或跨作用域移动。
- 不级联停用 Definition、取消 Run、修改应用配置、撤销 API key 或删除历史。
- 不创建通用资源 lifecycle framework、搜索服务、全文索引或第二套 audit owner。
- 不启用 production repository、真实 OIDC、production membership adapter、quota、billing 或公开生产声明。

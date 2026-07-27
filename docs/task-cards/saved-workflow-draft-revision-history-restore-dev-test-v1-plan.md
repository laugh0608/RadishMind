# 已保存 Workflow 草案修订历史、版本比较与显式恢复（开发 / 测试态）v1 实施任务卡

更新时间：2026-07-27

状态：`completed`

## 目标

在既有 Saved Draft owner 内交付 append-only 修订历史、精确版本读取、稳定分页、结构化比较和显式恢复，并保持当前记录 CAS、revision insert、权限与审计证据一致。

## 批次

### A：领域与内存 owner

- 定义 revision / summary / page / restore request 与稳定 failure code。
- memory store 每次成功保存原子追加 revision。
- 历史读取、分页、恢复为新版本和并发 CAS 单元测试。

### B：双数据库

- PostgreSQL / SQLite 新增 `saved_workflow_draft_revisions` migration 与现有 current snapshot 回填。
- 当前记录写入与 revision insert 使用同一事务。
- runtime revision 表只读 / 追加，覆盖重启、损坏、scope、migration 与 no-fallback。

### C：HTTP 与授权

- 新增 revision list / read / restore route。
- list / read 使用 `workflow_drafts:read`；restore 原子使用 read + write。
- body / path / active workspace / application binding 在 owner 前失败关闭。

### D：Web

- Saved Draft consumer 增加 revision list、read、restore。
- Draft Designer 增加历史分页、版本选择、结构化比较和显式恢复确认。
- 切换 draft / workspace、迟到响应、未保存编辑、冲突与 pending operation 均失败关闭。

### E：产品连续验证

- `v1 → v2 → v3 → compare → restore v1 as v4 → restart`。
- memory / SQLite / PostgreSQL、并发、隐私、Web build、快速与全量仓库门禁。
- 同步功能入口、当前焦点、路线图、能力矩阵和周志。

## 验证

```bash
(cd services/platform && go test ./internal/httpapi)
(cd services/platform && go test -race ./internal/httpapi)
npm --prefix apps/radishmind-web test
npm --prefix apps/radishmind-web run build
./scripts/run-workflow-saved-draft-postgres-dev-test.sh check
git diff --check
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```

完成证据：

- 领域、HTTP、SQLite migration 与 repository 定向 Go tests 通过。
- Web 254 项测试与 production build 通过，修订面板保持独立懒加载 chunk。
- 真实 PostgreSQL 门禁通过 `0001 → 0002`、重复 apply、current snapshot 回填、历史分页、精确读取、恢复为新版本、服务重启、运行角色 `UPDATE / DELETE` 拒绝和 reviewed rollback / reapply。
- 仓库快速与全量门禁结果记录在本周周志；没有新增独立 checker。

## 停止线

- 不新增第二套草案 owner、diff repository、merge engine 或 branch graph。
- 不允许 revision update / delete，也不伪造迁移前历史。
- 不绕过 membership、write gate、owner scope、CAS、schema preflight 或 no-fallback。
- 不产生 Run、provider、tool、confirmation、business writeback、replay 或 resume。
- 不扩 production OIDC、production repository、quota、billing 或公开生产能力声明。

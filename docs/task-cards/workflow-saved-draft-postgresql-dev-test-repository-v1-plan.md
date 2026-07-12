# Workflow Saved Draft PostgreSQL Dev/Test Repository v1 任务卡

更新时间：2026-07-11

## 任务标识

- 任务 ID：`workflow-saved-draft-postgresql-dev-test-repository-v1`
- 状态：`completed`
- 对应功能专题：`docs/features/workflow/saved-workflow-draft-postgresql-dev-test-repository-v1.md`

## 用户目标

Radish 体系内部开发者保存 Workflow 草案后，即使平台服务重启，也能在 Workspace Home 找回并恢复同一草案；并发保存、scope 错误或数据库故障必须给出稳定失败行为，不能覆盖他人记录或回退 sample / memory store。

## 输入

- `docs/features/workflow/saved-workflow-draft-v1.md`
- `docs/features/workflow/saved-workflow-draft-postgresql-dev-test-repository-v1.md`
- `services/platform/internal/httpapi/workflow_saved_draft.go`
- `services/platform/internal/httpapi/workflow_saved_draft_repository.go`
- `services/platform/internal/httpapi/workflow_saved_draft_repository_adapter.go`
- `services/platform/internal/httpapi/workflow_saved_draft_store_selector.go`
- `services/platform/migrations/workflow_saved_drafts/manifest.json`
- `docs/platform/engineering-health-productization-remediation-v1.md`

## 实现范围

1. 为 Saved Draft 请求补齐 request context、tenant、owner 和 scope，并贯穿 store / repository / query executor。
2. 新增显式 `postgres_dev_test` selector mode，production `repository` 保持关闭。
3. 引入 `pgx/v5`，实现受控连接池、schema preflight、关闭回收和脱敏失败。
4. 物化真实 PostgreSQL migration、schema marker、checksum、advisory lock、表和索引。
5. 实现 create / update CAS、read 和 owner-scoped list query executor，并复用现有 repository adapter。
6. 提供一次性 migration 入口、双端开发启动参数和 PostgreSQL route probe。
7. 增加真实 PostgreSQL 集成测试及 CI service，覆盖 migration、重启、CAS、隔离和故障。
8. 用真实 Web consumer 完成 PostgreSQL dev-live 浏览器验收。
9. 退出与本阶段冲突的历史 absence checker 调用，不新增 readiness checker 链。

## 提交拆分

1. 功能设计、任务卡和活动门禁口径。
2. context / scope / lifecycle 架构修正。
3. migration、connection pool、repository 和 selector。
4. PostgreSQL 集成测试、CI 与开发入口。
5. 浏览器证据和阶段文档收口。

## 验收矩阵

| 场景 | 必须结果 |
| --- | --- |
| 空库 migration | 事务成功并写入正确 marker / checksum |
| 重复 migration | 幂等，无重复表、索引或 marker |
| 首次保存 | version `1`，记录可 read / list |
| 服务重启 | 新进程 / 新连接池可恢复同一记录 |
| 并发更新 | 一个成功，其余 `draft_version_conflict` |
| owner 或 scope 不匹配 | 不返回、不覆盖目标记录 |
| marker / checksum 不匹配 | 启动 preflight 失败 |
| 数据库运行中断 | `draft_store_unavailable`，无 fallback |
| production `repository` | `repository_store_disabled` |
| 浏览器用户路径 | 保存、重启、恢复、冲突处置和 Handoff 均可复验 |

## 必要验证

- `go test ./...`
- `go test -race ./...`
- `go vet ./...`
- PostgreSQL integration test 入口
- `npm test` 与 `npm run build`（`apps/radishmind-web`）
- Saved Draft consumer / Review Handoff 既有聚合检查
- `./scripts/check-repo.sh --fast`
- `./scripts/check-repo.sh`
- PostgreSQL dev-live 浏览器复验

## 完成记录

- `pgx/v5` PostgreSQL repository、manual migration runner、schema marker、role separation、连接池生命周期和 `postgres_dev_test` selector 已实现。
- 真实 PostgreSQL 集成测试覆盖 migration / rollback / reapply、服务重建恢复、16 路 CAS、scope / owner 隔离、运行角色 DDL 拒绝、连接中断、marker mismatch 和 no fallback。
- Web Restore 状态跨 draft selection 的版本丢失已在浏览器验收中发现并修复；恢复后能从 version 1 正确续存到 version 2。
- 双标签冲突验收确认旧 version 保存返回 `draft_version_conflict`，本地内容不被覆盖，显式继续后以最新 version 3 保存到 version 4；Review Handoff 同步展示冲突审查证据。
- `npm test`、`npm run build`、Go 单元测试、`go test -race ./...`、`go vet ./...`、PostgreSQL 真实集成入口以及仓库 fast / full 门禁均已通过。

## 停止线

- 不接 production OIDC、membership、secret resolver、audit store 或 production database resource。
- 不启用 production `repository`，不把 `postgres_dev_test` 写成 production ready。
- 不自动 migration，不把数据库 URL 或原始错误写入日志、响应和 committed 资产。
- 不扩 Control Plane Read 数据库，不创建第二套 Saved Draft domain service。
- 不打开 publish、run、executor、confirmation、业务写回、replay 或 resume。

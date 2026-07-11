# Workflow Run History / Durable Dev-Test Run Store v1 任务卡

更新时间：2026-07-11

## 任务标识

- 任务 ID：`workflow-run-history-durable-dev-test-store-v1`
- 状态：`completed`
- 对应功能专题：`docs/features/workflow/workflow-run-history-durable-dev-test-store-v1.md`

## 用户目标

内部开发者能在 Workspace Run History 中分页审查真实 executor v0 运行记录，并在显式 PostgreSQL 开发 / 测试模式下跨服务重启恢复列表和详情；任何数据库或 scope 失败都必须失败关闭，不能回退内存成功或旧 fake route。

## 实现范围

1. 扩展独立 `workflowRunStore` contract：request context、record version、create / update、scoped read / list、过滤和 keyset cursor。
2. 保留 `memory_dev` 默认路径及 100 条 FIFO，增加稳定 scoped list 与并发安全测试。
3. 新增 `RADISHMIND_WORKFLOW_RUN_STORE=postgres_dev_test` selector、独立连接池、schema preflight 和 no-fallback failure mapping。
4. 新增 `services/platform/migrations/workflow_runs/` manual up/down migration、marker、checksum、索引和一次性 CLI。
5. 增加 `GET /v1/user-workspace/workflow-runs`，并与现有 detail route 对齐 scope、auth、envelope 和失败语义。
6. 在 repository 写入前复核 sanitized record、状态迁移、版本和 tool / confirmation / business write / replay 零计数。
7. 增加 migration / rollback、重启恢复、并发、scope、分页、连接失败和 no fallback PostgreSQL 集成测试。
8. 将 Web 真实 Run History list / detail consumer 拆入独立模块和 lazy chunk，保持 offline 默认模式，不扩大 `App.tsx` 与主包。
9. 完成真实浏览器历史 / 详情 / 重启恢复验证，并同步专题、当前焦点与周志。

## 验收矩阵

| 场景 | 必须结果 |
| --- | --- |
| memory 默认模式 | 100 条 FIFO，稳定分页，进程重启不承诺恢复 |
| PostgreSQL opt-in | migration 已应用且显式 dev gate 满足时启用 |
| create / update | record version 单调，终态不可逆，并发旧版本不能覆盖 |
| list | scope 内 `started_at DESC, run_id DESC`，过滤和游标稳定 |
| detail | 与 list 使用同一 store / scope，不读取旧 fake route |
| 服务重启 | 新 Server / pool 可恢复同一列表和详情 |
| scope 不匹配 | 不返回记录或数量，不泄露存在性 |
| 数据库 / marker 失败 | `workflow_run_store_unavailable`，无 memory fallback |
| 敏感字段 | 原始 input、condition value、credential、endpoint、raw envelope 不落库 |
| 副作用 | tool / confirmation / business write / replay 计数始终为 0 |
| Web offline | 不发 HTTP，sample 有明确标识 |
| Web dev/test | 真实分页列表与现有详情审查可复验，历史模块独立 chunk |

## 必要验证

- `go test ./services/platform/internal/httpapi/...`
- `go test -race ./services/platform/internal/httpapi/...`
- `go test` PostgreSQL workflow run integration 入口
- `go test ./...`
- `go test -race ./...`
- `go vet ./...`
- `npm test` 与 `npm run build`（`apps/radishmind-web`）
- 既有 Workflow / Web consumer 聚合门禁
- 真实浏览器 history / detail / restart 路径
- `./scripts/check-repo.sh --fast`
- `./scripts/check-repo.sh`

## 提交拆分

1. 实现主题：Go store / migration / API、集成测试、Web consumer 与浏览器修正。
2. 文档主题：功能状态、任务卡完成记录、当前焦点、专题入口和周志。

## 停止线

- 不复用 Saved Draft repository 领域接口，不启用 production repository / auth / API。
- 不接 tool、RAG、confirmation、业务写回、replay / resume 或后台保留任务。
- 不新增同层 checker；复用 Go / Web 测试、浏览器验证和聚合门禁。

## 完成记录

- 独立 memory / PostgreSQL run store、record version、终态保护、scope list / detail、过滤和 keyset cursor 已实现。
- 独立 workflow run migration / runner、runtime DML 角色、marker preflight、rollback / reapply、重启恢复、并发 CAS、marker mismatch 和 no-fallback 集成测试已通过。
- Web 历史与详情 consumer 已通过 lazy chunk 接入现有 Workspace Run History；旧 fake route 不参与真实列表，offline 默认不请求。
- 真实浏览器已复验 26 条记录分页、过滤、详情、零禁止副作用和服务重启恢复；最终会话无 error / warning。
- Go test / race / vet、Web test / build、PostgreSQL integration、仓库 fast / full 门禁作为最终完成证据记录在本周周志。

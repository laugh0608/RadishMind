# Workflow Execution Diagnostics / Failure Review v1 任务卡

更新时间：2026-07-11

## 任务标识

- 任务 ID：`workflow-execution-diagnostics-failure-review-v1`
- 状态：`completed`
- 对应功能专题：`docs/features/workflow/workflow-execution-diagnostics-failure-review-v1.md`

## 用户目标

内部开发者能在现有 Workspace Run History 中过滤并审查真实失败，明确失败边界、失败节点、最后完成节点、稳定诊断和人工审查动作；开发 / 测试环境能用受控故障场景复验这些路径，且不保存敏感原始材料、不触发外部副作用。

## 实现范围

1. 新写 `workflow_run_record.v1` diagnostic，兼容读取 v0。
2. 扩展 domain taxonomy、节点进度、terminal write 与 stale running 语义。
3. 扩展现有 list filter / cursor / summary 和 detail，不新增 API 资源族。
4. 新增 0002 manual migration、递进 marker、诊断过滤列和 PostgreSQL 集成测试。
5. 增加显式 dev-test mock failure selector，覆盖 Gateway、provider、取消、预算、output 与 store 边界。
6. 扩展 Run History lazy Web 模块的指标、过滤、详情、legacy 状态和复制 refs。
7. 补齐 Go / race / HTTP / PostgreSQL / Web / 浏览器验证，并同步正式文档。
8. 继续 R5 拆包，先把主包降到 650 KiB 以下，不新增 checker。

## 验收矩阵

| 场景 | 必须结果 |
| --- | --- |
| v0 历史 | list / detail 可读，明确 diagnostic unavailable |
| v1 成功 | diagnostic 无 failure，terminal write stored |
| Gateway / provider 失败 | 稳定 boundary / category / node / summary，不含 raw error |
| 请求取消 | canceled 终态，request boundary，零禁止副作用 |
| create store failure | API fail closed，不创建伪历史、不回退 memory |
| terminal write conflict | API 返回 store failure，最后 running 快照可标 stale，不自动终态化 |
| list filter | 新旧过滤组合、scope 与 cursor digest 稳定 |
| PostgreSQL 重启 | v0 / v1 与诊断过滤均恢复 |
| Web offline | 不发 HTTP |
| Web dev/test | 指标、过滤、时间线、诊断与 refs 可审查 |
| 敏感材料 | input、condition value、credential、endpoint、raw envelope、stderr / stack / SQL 均不存在 |
| 副作用 | tool / confirmation / business write / replay 均为 0 |

## 必要验证

- 精准 Go domain / store / HTTP 测试与 race。
- PostgreSQL integration：migration / rollback / reapply、filter、restart、conflict、no fallback。
- `go test ./...`、`go test -race ./...`、`go vet ./...`。
- Web consumer test、`npm test`、`npm run build` 和 bundle 体积审查。
- 真实浏览器失败列表 / 详情 / 过滤 / restart 路径。
- `./scripts/check-repo.sh --fast` 与 `./scripts/check-repo.sh`。

## 提交拆分

1. 实现主题：Go domain / store / migration / API、Web consumer / UI、测试和必要脚本。
2. 文档主题：功能完成状态、当前焦点、专题入口和周志证据。

## 停止线

- 不打开 tool、RAG、confirmation、业务写回、retry / fallback、replay / resume、checkpoint 或 production enablement。
- 不新增平行 API、后台自动终态化或同层 checker。
- 不把 Saved Draft repository 改造成 run repository，不回退旧 fake run source。

## 完成记录

- Go domain、store、HTTP 与 PostgreSQL 已实现 v1 diagnostic、v0 compatibility、失败过滤、cursor v2、固定 dev-test 故障场景和 0002 递进 migration。
- 单元 / race / HTTP / migration 与真实 PostgreSQL 集成覆盖 Gateway / provider / cancellation / budget / store、终态写失败、stale、scope、CAS、restart、rollback / reapply 和 no fallback。
- Web 11 项消费者测试覆盖 offline、v0 / v1 mapping、失败过滤、禁止副作用和 raw provider material 拒绝；build 主包为 624.57 KiB。
- 真实浏览器完成 5 条持久记录的指标、过滤、详情与 Platform 重启恢复；store create failure 未产生伪历史，terminal conflict 只保留 stale running 快照。
- 最终 Go test / race / vet、Web test / build、仓库 fast / full 门禁通过；本批浏览器、Platform / Web 和 PostgreSQL 服务均已关闭。

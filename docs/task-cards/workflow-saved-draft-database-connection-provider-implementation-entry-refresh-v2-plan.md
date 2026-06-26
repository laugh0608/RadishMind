# Workflow Saved Draft Database Connection Provider Implementation Entry Refresh v2 任务卡

更新时间：2026-06-23

## 背景

`Saved Workflow Draft Database Connection Lifecycle Readiness v1` 已补齐 connection provider 前置中的 timeout、pool、health check、close responsibility、request / audit propagation 和 sanitized diagnostics runtime 边界。现在需要重新复评 connection provider implementation entry，把已完成的静态 readiness 和仍 blocked 的 runtime 依赖分开。

## 范围

- 新增 `Saved Workflow Draft Database Connection Provider Implementation Entry Refresh v2` 专题文档。
- 新增 `workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v2` fixture / checker。
- 在 `check-repo.py` 注册 v2 checker。
- 同步 workflow 入口、功能入口、current focus、roadmap、capability matrix、product scope、scripts README 和周志。

## 明确不做

- 不创建 `docs/task-cards/workflow-saved-draft-database-connection-provider-implementation-v1-plan.md`。
- 不创建 database connection provider、secret resolver、DB driver、DSN parser、connection factory、pool runtime、health check runtime、role policy runtime、connection smoke runner、SQL、schema marker、migration runner、repository mode runtime、OIDC middleware、membership adapter 或 production API。
- 不启用 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 成功路径。
- 不连接数据库、不读取 secret、不调用 resolver、不调用 OIDC、不执行 workflow。

## 验收

- v2 文档明确 driver / DSN / TLS、role policy、connection smoke strategy 和 connection lifecycle 已达到静态 readiness / strategy，但仍不等价于 runtime ready。
- v2 fixture 固定依赖、entry gate、candidate decision、failure mapping、no fallback、no side effects、artifact guard 和 required doc references。
- v2 checker 能验证 fixture、自身文档引用、依赖状态、禁止 artifact、源码禁止 literal 和 `check-repo.py` 注册顺序。
- 入口文档同步当前状态为 `draft_database_connection_provider_implementation_entry_refresh_v2_defined`。

## 建议验证

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v2.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-lifecycle-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-smoke-strategy-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v1.py
git diff --check
./scripts/check-repo.sh --fast
```

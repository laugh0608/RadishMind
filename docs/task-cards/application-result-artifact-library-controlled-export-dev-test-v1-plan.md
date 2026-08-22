# 应用结果资产库与受控导出（开发 / 测试态）v1 实施任务卡

更新时间：2026-08-17

状态：`application_result_artifact_library_controlled_export_dev_test_v1_completed`

## 唯一功能真相源

[应用结果资产库与受控导出（开发 / 测试态）v1](../features/user-workspace/application-result-artifact-library-controlled-export-dev-test-v1.md)

本任务卡只承接新增 API、export 权限、schema 与双数据库索引 migration 的高风险实现边界，不替代功能文档。

## 批次 A

状态：已完成。memory、SQLite、PostgreSQL、HTTP、migration、configured Server no-fallback 与重启证据均已通过；没有新增平行 owner 或 export store。

目标：在不复制 artifact / lifecycle owner 的前提下，建立 application-scoped 严格读取与受控 export 后端纵向切片。

实现范围：

1. 将 `application-result-artifact-summary.schema.json` 对齐 `application_result_artifact_summary.v2`。
2. 新增 `application-result-artifact-export.schema.json` 与同构 Go validator。
3. repository 支持 application-scoped lifecycle list、execution profile / content type 过滤和稳定 cursor；既有 session list 保持兼容。
4. SQLite / PostgreSQL 顺序追加 application history index migration，不新增组件、DSN、pool 或 fallback。
5. 新增 application list / export route；export 要求 `application_sessions:read + application_result_artifacts:export`。
6. 覆盖跨 scope、非法 query、cursor 漂移、权限拒绝、corruption、no-store、export 零写入与不存在的 DELETE / share route。

验收：

- Go 精准单元测试与 `internal/httpapi` 包测试；
- SQLite migration / restart / corruption / no-fallback；
- PostgreSQL migration / rollback / reapply、runtime role、configured Server restart；
- `./scripts/check-repo.sh --fast`；
- 因 schema、API、permission、migration 和真相源变化，专题收口前运行全量 `./scripts/check-repo.sh`。

## 批次 B

状态：已完成。设计复核结论为 `0 / 0 / 1 / 1 / 0 = 2`、`C / 直接实现`；strict consumer、单一 owner、受控下载、Web 测试、production build 和三视口浏览器证据均已通过。

目标：建立 Application Result Workspace strict consumer 与真实用户路径。

前置：

- 完成 `B / 局部 Pencil`，或记录可复用既有基准面并把评分复核为 `C / 直接实现`；
- 批次 A contract 与 application list / export route 已稳定。

实现范围：

1. application-scoped list / export strict parser、offline 零请求和 generation guard。
2. filter rail、metadata list、exact inspector、Run handoff、archive / unarchive 与下载动作。
3. export 文件名只使用稳定短 identity，不使用用户内容、自然语言标题或 provider 标签。
4. 下载前重新核对 artifact ID、content digest、lifecycle version、scope 和 export digest。
5. 不写 URL、localStorage、sessionStorage、IndexedDB 或日志。

验收：Web `375/375`、production build、`1440×900` / `720×900` / `390×844`、筛选 route、零横向溢出、单一选中任务和控制台零 warning / error 均已通过。当前 Agent assignment 漂移按合同失败关闭；带真实 artifact 的 exact read / export / lifecycle 页面连续链属于批次 C。

## 批次 C

状态：已完成。共享 fixture、SQLite 真实页面连续链、PostgreSQL configured Server 重启 / no-fallback、export 零持久化与专题真相源收口均已通过。

目标：形成双数据库产品连续链并关闭专题。

实现范围：

1. 同一 application 下至少两个 Session、两种 profile、active / archived 与两种 content type。
2. SQLite 浏览器链：list → filter → exact read → export → archive / unarchive → 服务重启恢复。
3. PostgreSQL configured Server：list / export、关闭 no-fallback、重启恢复与 runtime role 索引证据。
4. export 不落库；artifact content / digest / lineage 不变；DELETE / public share route 继续不存在。
5. 同步功能索引、产品面、current focus、roadmap、capability matrix、integration contracts 与周志。

验收：

- memory / SQLite / PostgreSQL 使用同一定义的双 Session、双 profile、双 content type 与双 lifecycle fixture；重复启动不覆盖已推进的 lifecycle；
- SQLite 页面完成列表、组合过滤、精确读取、archive / unarchive、重启恢复和校验后下载，三视口无横向溢出且控制台零 warning / error；
- PostgreSQL 配置化 Server 完成相同核心读取 / export、关闭 no-fallback、重启恢复与 lifecycle 往返；
- Go 定向回归、Web `375/375`、production build、PostgreSQL 17 聚合集成、仓库 fast 与全量门禁通过；
- 没有新增 export store、永久 `DELETE`、public share、真实 Provider 或 production enablement。

## 停止线

- 不创建第二套 result store、export store、transcript store 或共享链接 owner。
- 不允许客户端上传或覆盖 artifact content。
- 不为 application library 放宽 owner、membership、scope 或 cursor 约束。
- 不执行外部写入、candidate action、replay / resume、自动清理或 production enablement。

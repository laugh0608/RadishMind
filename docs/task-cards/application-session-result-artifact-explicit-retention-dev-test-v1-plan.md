# 应用会话运行结果资产显式保存与恢复 v1 实施任务卡

更新时间：2026-08-17

状态：`application_session_result_artifact_explicit_retention_dev_test_v1_batch_c1_lifecycle_backend_completed`

对应功能：[应用会话运行结果资产显式保存与恢复（开发 / 测试态）v1](../features/user-workspace/application-session-result-artifact-explicit-retention-dev-test-v1.md)

## 批次 A 目标

在不改变 Run History 与 Session / Turn metadata-only 契约的前提下，完成批次 A 的可调用内存纵向链：用户在新 turn 上显式选择保存，服务端从同一次成功执行的 canonical result 创建独立 artifact，并提供 metadata-only list 与精确 content read。

## 批次 A 允许改动

- 新增 `application_result_artifact.v1` domain、summary、memory repository、service、cursor 与稳定 failure mapping。
- `ApplicationInteractionTurnExecutionInput` 和 HTTP body 增加默认 false 的 `save_result`。
- coordinator 在 terminal turn 成功后调用 artifact owner；支持五类现有 session profile。
- 新增 session-scoped list / read route，复用 `application_sessions:read|execute`。
- 增加 Go 单元、HTTP、并发 / 幂等、权限、隐私和兼容测试。
- 同步当前功能文档、索引、规划入口、能力矩阵和周志。

## 批次 A 禁止改动

- 不新增 SQLite / PostgreSQL migration、repository selector、DSN、连接池或生产 store。
- 不把 content 写入 run、session、turn、request history、audit ref、日志、URL、cursor 或 committed fixture。
- 不新增外部 create / upload route，不接受客户端 content、digest 或 run ref。
- 不改变 provider、runtime authority、Gateway 路由、自动 retry / fallback、HTTP Tool 或 RAG 算法。
- 不实现 Web 页面、archive / purge、公开分享、replay / resume、自动执行、业务写回或 production enablement。

## 批次 A 数据与风险门禁

- 单份 content 最大 `64 KiB`、必须 UTF-8 且非空。
- artifact 只能绑定已成功 terminal turn 与严格 run ref；每 turn 唯一。
- list 只能返回 summary；read 响应设置 `Cache-Control: no-store`。
- capture store failure 必须与 turn outcome 分开，不回滚 run、不返回伪成功、不重复 provider。
- `save_result` 省略 / false 必须保持零 artifact repository 调用。
- 幂等 retry 只能读取已有 artifact summary；首次未保存时不得补建。

## 批次 A 验收

- [x] domain / memory repository / strict cursor 与 failure mapping 完成。
- [x] 五类 profile canonical capture 完成。
- [x] turn `save_result`、artifact summary、artifact failure 表达完成。
- [x] session-scoped list / read route 完成。
- [x] 默认不保存、每 turn 唯一、scope / owner、store failure、no replay 与隐私测试完成。
- [x] Platform 精准测试、race、`go vet` 与仓库快速 / 全量门禁通过。
- [x] current focus、功能索引、路线图、能力矩阵和周志同步。

## 提交停止线

批次 A 提交只关闭内存纵向链；其历史停止线保持不变。

## 批次 B：双数据库 durable repository（已完成）

本批复用 Workflow Run Store 的共享 SQLite runtime、PostgreSQL pool、selector 和 migration family，增加独立不可变结果资产 repository；没有新增第十二个本地持久化组件、独立 DSN、连接池或 fallback。

允许改动：

- SQLite / PostgreSQL artifact 表、完整 scope 主键、session / turn 唯一键、cursor index、JSON 投影一致性和 update / delete 拒绝；
- backend selector 与 `Server` repository 注入；
- SQLite 真实文件、PostgreSQL 17、迁移、并发、重启、损坏 payload、运行角色和 no-fallback 测试；
- 功能文档、当前焦点、功能索引、路线图、能力矩阵、本地 SQLite 边界和周志。

验收结果：

- [x] SQLite `0022_application_result_artifacts` 与 PostgreSQL `0025_application_result_artifacts` 顺序迁移完成；
- [x] memory / SQLite / PostgreSQL selector 同构且无 backend fallback；
- [x] 每 turn 唯一、幂等重试、冲突、scope / owner、不可变、cursor 顺序和严格 payload 投影完成；
- [x] SQLite 重启 / 并发 / corruption / closed-store 与 PostgreSQL runtime role / 重连 / rollback / reapply 完成；
- [x] 聚合 PostgreSQL 开发测试链和仓库分层门禁通过。

本提交只关闭批次 B。Web consumer、archive / purge、浏览器产品链和 production capability 属于批次 C / D；批次 B 不创建其 route、migration、页面或声明。

## 批次 C1：生命周期后端纵向切片（已完成）

目标：在不修改 `application_result_artifact.v1` content / provenance 和现有不可变表的前提下，为结果资产建立 memory / SQLite / PostgreSQL 同构的 archive / unarchive owner。

允许改动：

- 新增 `application_result_artifact_lifecycle.v1`、`application_result_artifact_lifecycle_event.v1` 与 `application_result_artifact_summary.v2`；
- 新增独立 current lifecycle state、append-only event、expected-version CAS、active / archived list filter 与 cursor 绑定；
- 新增 `application_result_artifacts:archive` workspace permission，以及 archive / unarchive HTTP route；
- SQLite `0023_application_result_artifact_lifecycle`、PostgreSQL `0026_application_result_artifact_lifecycle` 和既有资产 active v1 回填；
- memory / SQLite / PostgreSQL 的并发、重启、scope、损坏投影、运行角色、rollback / reapply 与 no-fallback 验证。

禁止改动：

- 不修改或删除既有 artifact payload，不撤销数据库 update / delete 拒绝；
- 不新增 purge route、永久删除、自动 retention、级联删除或批量生命周期操作；
- 不以 `application_sessions:read|write|execute` 替代独立 archive 权限；
- 不在本子批复制三套 React consumer，不打开 transcript、replay / resume、真实 Provider 或 production capability。

退出条件：

- [x] 新资产与 active v1 lifecycle 同事务创建，旧资产由顺序 migration 回填；
- [x] 默认 list 仅返回 active，archived filter 与 cursor 完整绑定 scope；
- [x] archive / unarchive 原子完成 CAS 与 event，重复状态、陈旧版本和并发写失败关闭；
- [x] 精确 read 可读取同 owner archived artifact，跨 owner / application / session 继续不可见；
- [x] 三种 store、HTTP 组合权限、隐私、迁移和 no-fallback 证据通过；
- [x] 精准测试、race、`go vet`、仓库快速与全量门禁通过。

验收结果：SQLite marker 已推进到 `0023_application_result_artifact_lifecycle`，PostgreSQL marker 已推进到 `0026_application_result_artifact_lifecycle`；PostgreSQL 17 聚合集成与 configured profile 已复验。生命周期状态与事件不包含 artifact content，既有 artifact update / delete 拒绝保持不变。

批次 C1 关闭后才进入共享 Web strict consumer 与三类 Session surface；不创建 S11，是否需要局部 Pencil 由真实页面结构变化决定。

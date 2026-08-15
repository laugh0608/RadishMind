# 应用会话运行结果资产显式保存与恢复 v1 实施任务卡

更新时间：2026-08-15

状态：`completed`

对应功能：[应用会话运行结果资产显式保存与恢复（开发 / 测试态）v1](../features/user-workspace/application-session-result-artifact-explicit-retention-dev-test-v1.md)

## 批次目标

在不改变 Run History 与 Session / Turn metadata-only 契约的前提下，完成批次 A 的可调用内存纵向链：用户在新 turn 上显式选择保存，服务端从同一次成功执行的 canonical result 创建独立 artifact，并提供 metadata-only list 与精确 content read。

## 允许改动

- 新增 `application_result_artifact.v1` domain、summary、memory repository、service、cursor 与稳定 failure mapping。
- `ApplicationInteractionTurnExecutionInput` 和 HTTP body 增加默认 false 的 `save_result`。
- coordinator 在 terminal turn 成功后调用 artifact owner；支持五类现有 session profile。
- 新增 session-scoped list / read route，复用 `application_sessions:read|execute`。
- 增加 Go 单元、HTTP、并发 / 幂等、权限、隐私和兼容测试。
- 同步当前功能文档、索引、规划入口、能力矩阵和周志。

## 禁止改动

- 不新增 SQLite / PostgreSQL migration、repository selector、DSN、连接池或生产 store。
- 不把 content 写入 run、session、turn、request history、audit ref、日志、URL、cursor 或 committed fixture。
- 不新增外部 create / upload route，不接受客户端 content、digest 或 run ref。
- 不改变 provider、runtime authority、Gateway 路由、自动 retry / fallback、HTTP Tool 或 RAG 算法。
- 不实现 Web 页面、archive / purge、公开分享、replay / resume、自动执行、业务写回或 production enablement。

## 数据与风险门禁

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

本提交只关闭批次 A。SQLite / PostgreSQL、Web、生命周期和真实浏览器属于后续独立批次；没有对应实现与验证前不得把专题标记为完成。

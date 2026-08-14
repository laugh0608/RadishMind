# 应用解除归档与安全重新启用（开发 / 测试态）v1 实施任务卡

更新时间：2026-07-29

状态：`application_unarchive_safe_reactivation_dev_test_v1_completed`

## 任务目标

按照[应用解除归档与安全重新启用（开发 / 测试态）v1](../features/user-workspace/application-unarchive-safe-reactivation-dev-test-v1.md)，为已归档应用建立显式影响确认、组合权限、记录版本 CAS、三种存储和 Web 审查连续链，使同一应用身份能够安全恢复，同时保留下游 owner 的独立状态和资格规则。

## 准入与边界

- 现有应用目录已完成 `active -> archived`，但恢复会让部分既有凭据、绑定和会话重新获得资格，属于新增 API 与高风险生命周期边界。
- 本任务卡只承接一个功能专题的批次 A 至 C，不拆分同层 readiness / refresh 链。
- 复用现有 `application_catalog_record.v1`、数据表、store selector、权限投影和响应 envelope；不新增 migration 或 schema 版本。
- 开发测试态与 production 分开验收；本任务不声明生产授权或生产恢复流程。

## 批次 A：领域、存储与资格回归

状态：已完成

实现：

1. 为 `applicationCatalogRepository` 和 service 新增 `Unarchive`。
2. service 要求合法 application ID、正整数 `expected_version` 和显式影响确认。
3. Memory owner 原子执行 `archived -> active`、版本加一、更新时间 / actor / request / audit 更新和 `archived_at` 清空。
4. SQLite / PostgreSQL 使用 `record_version + lifecycle_state` CAS，并精确区分版本冲突和非法转换。
5. 扩展三种 owner 的 archive → unarchive、重复转换、陈旧版本、并发与重启回归。
6. 扩展 Gateway API Key 鉴权回归，证明未吊销 Key 只在应用活动时有效，已吊销 Key 不因恢复应用而复活。

批次 A 完成条件：

- 三种 owner 的转换语义一致；
- 数据库无需 migration，旧记录可直接恢复；
- 同一版本并发最多一个成功；
- 应用恢复不修改 API Key、assignment、Session、Draft、Candidate 或 Run owner。

## 批次 B：HTTP、组合权限与失败关闭

状态：已完成

实现：

1. 注册 `POST /v1/user-workspace/applications/{application_id}/unarchive`。
2. 请求体严格包含 `workspace_id`、`expected_version`、`acknowledge_existing_access_reactivation`。
3. 使用 `applications:archive + applications:write` 的单次 membership provider decision。
4. 继续使用现有开发 HTTP / write enablement、活动工作区和 owner scope。
5. 成功返回活动记录；版本冲突返回当前版本与状态；重复恢复返回 `application_catalog_transition_invalid`。
6. 覆盖未知字段、确认缺失 / 为假、workspace 漂移、scope / membership 任一权限缺失、OIDC membership unavailable 和 store unavailable。

批次 B 完成条件：

- 认证和 membership 拒绝发生在业务 repository 之前；
- 确认失败不查询或写入 application repository；
- 请求不调用 Gateway、Provider、模型、工具或网络；
- HTTP route 与 service / repository 失败码保持一套语义。

## 批次 C：Web、连续产品链与阶段收口

状态：已完成

实现：

1. Web consumer 增加 unarchive request builder、严格响应解析和 `unarchived` 状态。
2. dev header 同时发送 read、archive、write scopes，membership proof 同时包含 archive、write。
3. 已归档详情展示重新启用影响；最终确认才发送确认字段。
4. 成功后以服务端记录切换活动筛选与选中应用上下文；冲突要求显式刷新。
5. 补齐 consumer 回归和 production build。
6. 使用 SQLite 本地产品链与浏览器验证归档阻断、解除归档、API Key 恢复资格、新配置继续和历史保持可读。
7. 复验 PostgreSQL 连接 / 重启恢复、快速与完整仓库门禁、敏感信息和服务清理。

批次 C 完成条件：

- 用户能理解恢复影响并完成可审计操作；
- Web 不在失败、冲突或迟到响应下伪造活动状态；
- 既有下游资源只按自身状态恢复资格，不被目录 owner 级联改写；
- 本批启动的服务、容器和浏览器页面在提交前关闭；
- 功能设计、当前焦点、入口、周志与任务卡统一关闭。

## 验证矩阵

| 层级 | 必须验证 |
| --- | --- |
| 领域 / Memory | archive → unarchive、重复转换、陈旧版本、并发、所有者隔离、确认失败零查询 |
| SQLite | 原子 CAS、并发单赢家、服务重建读取、审计字段、无 migration |
| PostgreSQL | 真实数据库 CAS、并发、重新连接恢复、运行角色无 DDL |
| HTTP | strict body、组合权限、workspace、OIDC 失败关闭、stable failure metadata |
| 下游 | API Key 归档拒绝 / 恢复、已吊销 Key 不恢复、配置与候选保持既有基线规则 |
| Web | 离线零请求、精确 headers / body、strict envelope、两步确认、冲突刷新、活动上下文交接 |
| 产品链 | 真实浏览器 archive → blocked → unarchive → eligible、历史可读、敏感信息检查 |
| 仓库 | 定向 Go / race / vet、Web test / build、PostgreSQL integration、fast 与 full `check-repo` |

## 停止线

- 不新增生产 adapter、production auth、quota、billing 或发布声明。
- 不实现物理删除、批量操作、自动恢复、自动资源重建或集中式下游事务。
- 不让目录 owner 直接修改 API Key、assignment、Session、Draft、Candidate、Definition、Run 或 Request History。
- 不新建同层 checker；优先复用平台 Go 回归、Web 测试、PostgreSQL 集成、consumer smoke 和仓库聚合门禁。

## 关闭结论

批次 A 至 C 已全部完成。三种 store CAS、组合权限、显式影响确认、Gateway 资格回归、Web 两步确认、SQLite 真实浏览器连续链、PostgreSQL 集成和仓库门禁已经形成可复验证据；本任务卡关闭，不继续派生平行任务卡。

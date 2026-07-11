# Workflow Evaluation Baseline & Case Versioning v1

更新时间：2026-07-11

状态：`workflow_evaluation_baseline_case_versioning_v1_completed`

## 功能目标

让内部开发者在保留既有评测用例历史的前提下，显式提升 baseline 或修订候选预期，并能复查任一历史版本的定义、审计信息和即时 batch review。版本变更只引用既有 durable run，不复制运行载荷，不自动选择 baseline，也不重新执行、重试、重放或恢复工作流。

## 用户路径

1. 用户从 Workspace Run History 打开一个 evaluation case，查看当前版本、baseline、候选预期和历史修订。
2. 用户选择“提升 baseline”或“修订用例”，提交完整目标快照和当前 `expected_version`；服务端重新校验所有 run 的 scope、资格和零禁止副作用。
3. repository 原子创建下一版本。并发期间版本已变化时返回 conflict 和当前版本，用户必须重新审查，不自动覆盖或合并。
4. 用户可分页查看修订链，打开任一版本并即时运行只读 batch review；旧版本的定义和审计引用保持不变。
5. 默认 offline 模式不请求；旧 fake `/v1/user-workspace/runs` 不参与 baseline、revision 或 review。

## 版本模型与生命周期

一个 `case_id` 表示稳定用例家族，`version` 从 1 单调递增。每个 `workflow_evaluation_case.v2` 是不可变完整快照，包含名称、baseline、候选预期、revision kind、前一版本、服务端生成的 change codes、稳定 family `created_at`、逐版 `revised_at`、actor / request / audit ref。v1 记录读取时兼容为 version 1、`created`，并以原 `created_at` 作为首版 `revised_at`，不回写或伪造历史操作。

允许的 revision kind：

- `baseline_promotion`：baseline 必须变化；提交者同时给出新的完整候选预期集合。
- `case_revision`：baseline 必须保持不变，名称、候选集合或 expected classification 至少一项变化。

创建为 `created`。禁止无变化 revision、跳号、删除、回滚覆盖和历史修改。所谓回到旧定义也必须基于当前版本创建新 revision，历史仍保留。

## Baseline promotion 语义

baseline promotion 是人工审查后的显式元数据决策，不从 passed / improvement 自动推导。新 baseline 可以是当前 candidate 或同 scope 的其它 eligible run，但不能同时作为新版本 candidate。所有引用必须为终态或 stale，且 tool、confirmation、business write、replay 计数均为 0。

服务端只生成 `baseline_changed`、`name_changed`、`candidate_added`、`candidate_removed`、`expectation_changed` 等稳定 change code，不持久化自由文本理由。actor、request 和 audit ref 提供修订责任链；v1 不引入独立生产 audit store 声明。

## API、scope 与分页

- `POST /v1/user-workspace/workflow-evaluation-cases/{case_id}/revisions`
- `GET /v1/user-workspace/workflow-evaluation-cases/{case_id}/revisions?workspace_id=...&application_id=...&limit=...&cursor=...`
- `GET /v1/user-workspace/workflow-evaluation-cases/{case_id}/revisions/{version}?workspace_id=...&application_id=...`
- 既有 detail 返回当前版本；既有 review 增加可选、严格的 `version` 查询参数。

写入要求 `workflow_evaluations:write` 与 `workflow_runs:read`，读取要求 `workflow_evaluations:read` 与 `workflow_runs:read`。tenant / workspace / application scope 与既有资源族一致；跨 scope 和不存在统一 not found。

revision list 按 `version DESC`，默认 25、最大 100；opaque cursor 绑定 case id 与 limit。case list 继续只列每个家族的当前版本，并沿用 `created_at DESC, case_id DESC` 游标，因此 revision 不会制造重复列表项。

## 并发、失败与恢复

revision 请求必须携带正整数 `expected_version`。memory 与 PostgreSQL repository 都以 case family 为原子边界执行 compare-and-swap；PostgreSQL 在单事务中锁定 family、插入 revision、更新当前快照和版本。冲突返回 `workflow_evaluation_version_conflict`、当前版本和当前 case，不得自动 retry、merge 或 fallback。

其它稳定失败包括 invalid、run not eligible、not found、revision cursor invalid、store unavailable 和 store contract mismatch。terminal database failure 不回退 memory；revision 已提交但响应中断时，客户端通过 current detail / revision history复核，不重复猜测成功。

## 持久化与 migration

`memory_dev` 保持默认，并在既有 100 个 case family 预算内保存每个 family 的完整 revision 链。`postgres_dev_test` 增加 workflow-runs 0004 manual migration：为 `workflow_evaluation_cases` 增加当前版本列，并新增独立 `workflow_evaluation_case_revisions` 表；迁移把既有 v1 sanitized record 作为 version 1 回填。

当前 family 快照继续支撑 list / detail，revision 表只保存不可变版本。两者复用既有 run-store selector 与连接池生命周期，但 repository 仍独立于 Saved Draft。声明式保留沿用 case 90 天 / 每 scope 2,000 family；只要 family 可读，其 revision 链不可被部分清除。v1 不自动清理。

## 脱敏与可观测性

禁止持久化或返回原始 input、input bytes、condition value、draft payload、output / preview、credential、endpoint、provider raw envelope、comparison snapshot、自由文本变更理由或业务 payload。名称继续执行 96 字符与 secret / URI 拒绝规则。

日志只允许 case id hash、from / to version、revision kind、change codes、candidate count、store mode、scope hash、duration、request id、audit ref 和 outcome。review 继续验证四类禁止副作用计数为 0。

## Web 与性能

在既有 lazy Evaluation Cases 面板内增加 current version、revision action、expected-version conflict、revision history、历史版本详情和历史 review。编辑器以完整目标快照提交，不做隐式 patch 或自动合并。offline 零请求；功能继续留在独立 lazy chunk，主入口保持低于 500 KiB。

## 验收与停止线

- Go：v1 兼容升级、revision domain、baseline / case revision、change codes、CAS 并发、scope、游标、历史 review、敏感字段和失败映射。
- PostgreSQL：fresh、0001 / 0002 / 0003 pending、回填、rollback / reapply、并发 CAS、重启恢复、scope、运行角色和 no fallback。
- Web：offline、current / history / revision / conflict / historical review、strict response 和零副作用。
- 浏览器：创建 case、修订预期、提升 baseline、制造并发 conflict、审查旧 / 新版本并在服务重启后恢复。
- 不实现自动 baseline、scheduled / batch execution、retry、replay / resume、delete、history rewrite、tool、RAG、confirmation commit、业务写回、production audit store 或 production enablement。

## 完成结果

- `workflow_evaluation_case.v2` 已落地稳定 family `created_at`、逐版 `revised_at`、version、previous version、revision kind 与服务端 change codes；既有 v1 读取和 PostgreSQL migration 回填为 version 1。
- 新增完整快照 revision、人工 baseline promotion、原子 `expected_version` CAS、分页 revision history、指定版本 detail / review；跨 scope、无变化、kind 不匹配、非法游标和 store failure 均 fail closed。
- PostgreSQL workflow-runs 0004 增加 `current_version` 与不可变 revision 表，覆盖 fresh、0001 / 0002 / 0003 pending、v1 回填、rollback / reapply、并发 CAS、重启恢复、scope、runtime role 和 no fallback。
- Web 在既有 lazy evaluation chunk 内完成 current version、完整快照修订、baseline promotion、冲突刷新、五版审计链、严格响应校验和历史 review；chunk 为 14.12 KiB，主入口保持 430.39 KiB，offline 零请求。
- 真实 PostgreSQL 浏览器路径完成 v1 创建、v2 expected 修订、v3 baseline promotion、外部并发 v4 / v5、旧版本冲突拒绝和 Platform/Web 重启恢复；新会话除 React DevTools 提示外无 console error / warning。
- Go test / race / vet、Web 19 项测试 / build、PostgreSQL integration 和仓库 fast / full 门禁通过；浏览器、Platform/Web 与 PostgreSQL 容器 / 网络均已关闭。

下一产品设计优先进入 `Workflow Evaluation Suite / Release Review v1`：把明确 case version 组成不可变 suite，提供聚合只读审查和人工 release decision evidence；不自动执行、部署、promote 或写回业务系统。

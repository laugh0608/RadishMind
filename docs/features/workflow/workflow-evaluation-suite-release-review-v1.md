# Workflow Evaluation Suite / Release Review v1

更新时间：2026-07-11

状态：`workflow_evaluation_suite_release_review_v1_completed`

## 功能目标

让内部开发者把 1–50 个明确的 evaluation case version 组成不可变 suite，获得可复验的聚合只读 review，并追加与 review digest 绑定的人工 release decision evidence。decision 只表达审查结论，不触发部署、发布、baseline promotion、工作流执行或业务写回。

## 用户路径

1. 用户在 Run History 的 evaluation 区域选择若干 `case_id + version`，填写短名称并创建 suite。
2. 服务校验所有 case revision 位于同一 tenant / workspace / application scope、引用互异且可读，创建不可变 suite。
3. 用户打开 suite review；服务即时复用每个历史 case version 的 batch review，聚合 passed / mismatch / inconclusive / unavailable 和稳定 review digest。
4. 用户基于当前 review 显式提交 `approved|rejected|needs_review`，携带当前 digest 与 `expected_decision_version`。服务重新计算 review 后原子追加 decision evidence。
5. 用户分页查看 suite、decision history 和审查引用。默认 offline 不请求，旧 fake `/v1/user-workspace/runs` 不参与。

## Suite 与 decision 生命周期

`workflow_evaluation_suite.v1` 创建后不可修改、删除或把 case ref 指向 current；每项必须固定正整数 version。需要调整组成时创建新 suite。

`workflow_evaluation_release_decision.v1` 是 append-only evidence，decision version 从 1 单调递增。请求必须携带当前 `expected_decision_version` 和 review digest；并发冲突或 digest 已变化时 fail closed。`approved` 仅在 suite review 为 passed 且 unavailable 为 0 时允许；rejected / needs_review 可记录其它结果。后续 decision 不覆盖历史，也不代表 production release authorization。

## 聚合 review 与 digest

每项返回 case id、version、case name、case outcome、matched / mismatched / inconclusive / unavailable 数和 case audit ref。suite outcome：全部 case passed 为 passed；任一 mismatch 为 mismatch；否则为 inconclusive。case revision 或引用 run 不可读时 item 为 unavailable。

digest 由 schema version、suite id、精确 case refs、逐项 outcome/counts 和聚合 outcome 的 canonical JSON 计算 SHA-256；不包含时间、actor、输出或自由文本。review 本身不持久化，decision 只保存 digest、聚合计数和引用。

## API、scope 与分页

- `POST /v1/user-workspace/workflow-evaluation-suites`
- `GET /v1/user-workspace/workflow-evaluation-suites?workspace_id=...&application_id=...&limit=...&cursor=...`
- `GET /v1/user-workspace/workflow-evaluation-suites/{suite_id}?workspace_id=...&application_id=...`
- `GET /v1/user-workspace/workflow-evaluation-suites/{suite_id}/review?workspace_id=...&application_id=...`
- `POST /v1/user-workspace/workflow-evaluation-suites/{suite_id}/decisions`
- `GET /v1/user-workspace/workflow-evaluation-suites/{suite_id}/decisions?workspace_id=...&application_id=...&limit=...&cursor=...`

读要求 `workflow_evaluations:read` 与 `workflow_runs:read`，创建 suite / decision 要求 `workflow_evaluations:write` 与 `workflow_runs:read`。跨 scope 与不存在统一 not found。suite list 按 `created_at DESC, suite_id DESC` keyset cursor；decision list 按 version DESC，游标绑定 suite id 与 limit。

## 持久化、并发与 migration

`memory_dev` 保持默认，提供进程内 100 个 suite family FIFO 和各自完整 decision trail，所有读写仍按 scope 隔离。`postgres_dev_test` 增加 workflow-runs 0005 manual migration，新增独立 `workflow_evaluation_suites` 与 `workflow_evaluation_suite_decisions` 表；不改造 run、case revision 或 Saved Draft 表。

decision CAS 在单事务中锁定 suite、比较 current decision version、插入 evidence 并更新 current projection。数据库失败不回退 memory，不自动 retry。声明式保留为 90 天 / 每 scope 1,000 suite；suite 存续期间 decision trail 不部分清除，v1 不自动清理。

## 脱敏、可观测性与失败语义

只保存 suite / decision id、scope、名称、精确 case refs、decision enum、decision version、review digest、聚合 outcome/counts、actor / request / audit ref 和时间。禁止 input、condition、draft、output、preview、credential、endpoint、provider envelope、comparison body、run payload、自由文本理由和部署信息。

稳定 failure 包含 invalid、not found、case not eligible、cursor invalid、review changed、decision conflict、approval blocked、store unavailable 和 store contract mismatch。日志仅允许 suite id hash、case count、outcome/counts、digest 前缀、decision/version、store mode、scope hash、duration、request/audit ref。

## Web、验收与停止线

新增独立 lazy Suite / Release Review 面板，支持 exact case version 选择、创建、列表、聚合 review、人工 decision、冲突和 decision history；不扩大主入口，offline 零请求。

- Go：domain、exact ref、聚合/digest、decision policy/CAS、scope、分页、敏感字段和失败。
- PostgreSQL：fresh、0001–0004 pending、rollback / reapply、并发 decision、重启恢复、scope、runtime role 和 no fallback。
- Web：offline、create/list/review/decision/conflict/history、strict response 和 lazy build。
- 浏览器：创建多 case-version suite、review、批准阻塞、needs_review / approved evidence、并发 conflict 和重启恢复。
- 本专题原始版本不实现 tool 或 RAG；后续 RAG 只由独立 profile 专题以显式 run profile 接入。仍不实现 suite update/delete、自动 suite、定时/批量执行、真实 release/deploy、retry、replay/resume、confirmation commit、业务写回或 production enablement。

## 完成证据

1. Go 已实现 memory / PostgreSQL suite repository、exact revision 资格、稳定聚合 digest、approval policy、append-only decision CAS、scope、两类 keyset cursor、strict HTTP 和稳定失败语义；数据库路径不会回退 memory。
2. workflow-runs 0005 manual migration 已覆盖 fresh apply、0001–0004 pending upgrade、rollback / reapply、runtime DML role、并发 decision 和 pool 重开恢复。
3. Web 以 13.78 KiB 独立 lazy chunk 接入 Run History，支持 case / suite 显式刷新、exact version 创建、聚合审查、decision 与历史证据；主入口保持 430.39 KiB，默认 offline 零请求。
4. 真实浏览器完成 `Save executor draft → two runs → run history/detail → case → suite → mismatch review → needs_review v1`。详情显示原始 input / condition value 未保留，tool / confirmation / business write / replay 合计为 0。
5. 本专题只完成开发 / 测试态 release review evidence，不授权发布或生产启用；下一产品任务需回到四个一级产品面重新排位，不从本文自动派生 production gate 链。

2026-07-18 独立的 [Workflow RAG Regression Review 与 Evaluation Profile v1](workflow-rag-regression-review-evaluation-profile-dev-test-v1.md) 已允许 suite 聚合普通 case 与 RAG case；每个 item 显式返回 `run_profile`，canonical review digest 将 profile 纳入签名。SQLite shared database 补齐 durable suite / decision repository，PostgreSQL 继续复用既有表；HTTP Tool v2 仍明确 unsupported，decision 仍不触发 release、执行或 baseline promotion。

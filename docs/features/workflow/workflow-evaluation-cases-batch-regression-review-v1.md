# Workflow Evaluation Cases / Batch Regression Review v1

更新时间：2026-07-11

状态：`workflow_evaluation_cases_batch_regression_review_v1_completed`

## 功能目标

让内部开发者把一个 durable baseline run、1–20 个 durable candidate run 和每个候选的预期分类保存为可复用 evaluation case，并在 Workspace Run History 中批量审查实际 comparison 是否符合预期。case 只保存稳定引用和预期，不保存输入、condition、输出或 comparison 快照，不重新执行、重放或恢复 run。

## 用户路径

1. 用户在真实 Run History 中选择一个 baseline 和多个 candidates，填写短名称并为每个 candidate 选择预期 `regression|improvement|changed|unchanged`。
2. Platform 在同一 tenant / workspace / application scope 校验所有 run ref、终态或 stale 状态、零禁止副作用和数量预算后创建不可变 case。
3. 用户分页查看 evaluation cases，打开 batch review；服务即时复用 `workflow_run_comparison.v1` 比较每个 candidate。
4. review 展示 passed / mismatch / inconclusive / unavailable item、聚合计数、失败边界和人工动作；用户仍可回到原始 run 详情。
5. 默认 offline 模式不请求；旧 fake `/v1/user-workspace/runs` 不参与 case 或 review。

## 生命周期与 scope

case schema 为 `workflow_evaluation_case.v1`，创建后不可修改、删除或替换 baseline。不可变定义避免预期被静默改写；需要新预期时创建新 case。case 本身为 `active`，当任一引用因既有 run 保留策略不可读取时 review 派生 `expired`，但记录继续可审查，不伪造 run。

所有操作要求相同 tenant / workspace / application scope。创建要求 `workflow_evaluations:write` 与 `workflow_runs:read`；list / detail / review 要求 `workflow_evaluations:read` 与 `workflow_runs:read`。跨 scope 与不存在统一 not found，不泄漏引用存在性。

## API、分页与失败语义

- `POST /v1/user-workspace/workflow-evaluation-cases`
- `GET /v1/user-workspace/workflow-evaluation-cases?workspace_id=...&application_id=...&limit=...&cursor=...&baseline_run_id=...`
- `GET /v1/user-workspace/workflow-evaluation-cases/{case_id}?workspace_id=...&application_id=...`
- `GET /v1/user-workspace/workflow-evaluation-cases/{case_id}/review?workspace_id=...&application_id=...`

list 使用 `created_at DESC, case_id DESC` keyset cursor，并绑定 baseline filter 与 limit。未知、重复或越界字段 fail closed。稳定 failure code 包含 invalid、run not eligible、record not found、cursor invalid、store unavailable 和 store contract mismatch。数据库失败不得回退 memory。

## Batch review 语义

每项返回 expected / actual classification、`matched`、comparison state、status transition、finding codes 和推荐动作。实际为 `inconclusive` 或 run 已过期时 item 为 inconclusive / unavailable，不计为通过。聚合 outcome：全部匹配为 `passed`；任一 mismatch 为 `mismatch`；没有 mismatch 但存在 inconclusive / unavailable 为 `inconclusive`。

review 是即时派生结果，不持久化；同一 case 在 run 保留期内可复验，过期后明确降级。耗时变化仍不单独宣称 regression。

## 持久化、保留与 migration

`memory_dev` 随现有 run store mode 提供 100 条 FIFO case store；`postgres_dev_test` 在 workflow-runs 组件增加独立 `workflow_evaluation_cases` 表和 0003 manual migration，复用同一连接池与 selector 生命周期，但不把 run record 表改造成 case 表。case 声明保留 90 天、每 scope 2,000 条；v1 不自动清理，只提供可观察约束，后续清理必须独立设计。

## 脱敏与可观测性

只保存 case id、名称、scope、baseline / candidate run ids、预期枚举、actor / request / audit ref 和 created_at。禁止保存 input text / bytes、condition、draft payload、output / preview、credential、endpoint、provider envelope、comparison body或业务 payload。名称最多 96 字符且拒绝 URI、换行和疑似 secret 形态。

服务日志只记录 outcome、case id hash、candidate count、matched / mismatch / inconclusive 数、store mode、scope hash、duration、request id 和 audit ref。所有引用 run 的 tool / confirmation / business write / replay 必须为 0。

## Web 与性能

在现有 Run History 增加独立 lazy Evaluation Cases 面板，支持 baseline、candidate 集合、expected classification、创建、分页列表和 batch review。comparison 单项详情仍复用现有面板。offline 保持零请求，新增模块不进入主入口 chunk；既有 430.39 KiB 主入口预算不得回退到 500 KiB 以上。

## 验收与停止线

- Go：domain、不可变 store、并发、scope、分页 cursor、资格、聚合、过期和禁止字段。
- PostgreSQL：0001→0002→0003、fresh、rollback / reapply、重启恢复、scope、并发 create、runtime DDL 拒绝、连接失败 no fallback。
- Web：offline、创建 / list / review、strict response、零副作用和 lazy chunk。
- 浏览器：创建成功与失败候选 case、batch review、详情、分页 / 重启恢复和敏感字段缺失。
- 不实现 case update / delete、定时调度、自动 baseline、批量执行、retry、replay / resume、tool、RAG、confirmation commit、业务写回或 production enablement。

## 完成结果

- 已实现不可变 `workflow_evaluation_case.v1`、1–20 个 candidate 预期、终态 / stale 与零副作用资格校验、memory FIFO repository、scope 和 keyset cursor。
- 新增 create / list / detail / review 资源族；batch review 即时复用 comparison domain，返回 passed / mismatch / inconclusive / unavailable 聚合，不保存 comparison 或重新执行 run。
- workflow-runs 0003 manual migration 新增独立 `workflow_evaluation_cases` 表，支持 fresh、0001 / 0002 pending 递进、rollback / reapply、运行角色和 marker 校验；selector 与 run store mode / pool 生命周期一致，数据库失败不回退 memory。
- Web 新增 8.55 KiB 独立 lazy panel，支持 case 名称、baseline、多个 candidate、预期分类、创建、列表与 batch review；主入口保持 430.39 KiB，offline 零请求。
- PostgreSQL 真实浏览器创建一个成功 baseline、Gateway timeout 与 provider failure 两个 candidates 的 case，review 为 `passed / 2 matched / 0 mismatch / 0 inconclusive / 0 unavailable`；Platform/Web 重启后 case 与 review 完整恢复。
- Go test / race / vet、Web 18 项测试 / build、PostgreSQL integration 与真实浏览器通过；浏览器、Platform/Web、PostgreSQL 容器和网络已关闭。Docker Desktop 关闭操作被本机拒绝，但没有 RadishMind 容器残留。

后续 [Workflow Evaluation Baseline & Case Versioning v1](workflow-evaluation-baseline-case-versioning-v1.md) 已完成 baseline promotion、case revision、expected classification 变更审查、并发版本和审计链，[Workflow Evaluation Suite / Release Review v1](workflow-evaluation-suite-release-review-v1.md) 也已完成 exact-version 聚合审查与人工 decision evidence；仍不保存输入、不自动执行或 replay。

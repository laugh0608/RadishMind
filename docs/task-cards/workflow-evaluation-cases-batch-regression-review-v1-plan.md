# Workflow Evaluation Cases / Batch Regression Review v1 任务卡

更新时间：2026-07-11

状态：`workflow_evaluation_cases_batch_regression_review_v1_completed`

## 目标

交付 durable evaluation case、scoped list / detail / batch review API 和真实 Web 审查路径，使多个既有 run comparison 可按显式预期复验，同时保持输入、执行与外部副作用关闭。

## 实现范围

1. 实现 case domain、资格、不可变 lifecycle、keyset cursor 与 batch 聚合。
2. 实现 memory store 和随 run store selector 联动的独立 PostgreSQL case repository。
3. 增加 workflow-runs 0003 manual migration、rollback / reapply 和 integration coverage。
4. 增加 scoped POST / list / detail / review route 与 stable failure envelope。
5. 增加独立 Web consumer / lazy panel及真实浏览器路径，不扩大主入口。
6. 更新当前焦点、runtime 专题、周志和任务索引，按实现 / 文档主题提交。

## 门禁

- case 不保存或返回 input、condition、output、preview、credential、endpoint、raw envelope 或 comparison snapshot。
- 创建时所有 run 同 scope、互异、终态或 stale，禁止副作用为 0；1–20 candidates。
- PostgreSQL 失败 no fallback；comparison item 过期显式 unavailable。
- 完成 Go test / race / vet、Web test / build、PostgreSQL integration、浏览器、仓库 fast / full。

## 完成证据

- Go domain / memory / HTTP 覆盖资格、不可变 case、并发安全、scope、cursor、matched / mismatch / inconclusive / unavailable 和敏感字段。
- PostgreSQL 0003 覆盖 fresh、v1 / v2 pending、rollback / reapply、runtime DDL、重启恢复、scope 与 no fallback。
- Web 18 项测试与 build 通过；evaluation lazy chunk 8.55 KiB，主入口保持 430.39 KiB。
- 浏览器在 PostgreSQL 模式完成 2-candidate passed review 和 Platform/Web 重启恢复；相关服务与容器已关闭。

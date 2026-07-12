# Workflow Evaluation Suite / Release Review v1 任务卡

更新时间：2026-07-11

状态：`workflow_evaluation_suite_release_review_v1_completed`

## 目标

交付 exact-case-version suite、即时聚合 review、digest-bound append-only 人工 decision evidence 和真实 Web 审查路径，保持 ref-only、no fallback 和零执行副作用。

## 实现范围

1. 实现 suite / decision domain、资格、聚合 review、canonical digest、审批策略和 decision CAS。
2. 实现 memory / PostgreSQL repository、suite / decision 分页和 workflow-runs 0005 migration。
3. 增加 scoped create/list/detail/review/decision API 与稳定 failure envelope。
4. 增加独立 Web consumer / lazy panel及真实浏览器路径。
5. 更新焦点、runtime、周志和索引，按实现 / 文档主题提交。

## 门禁

- suite 只引用明确 case version；definition 和 decision history 不可变。
- decision 必须绑定重新计算的 review digest 与 expected decision version；approved 只接受 passed review。
- 不保存运行载荷、comparison body、自由文本理由、credential、endpoint、provider raw envelope 或部署信息。
- PostgreSQL no fallback；完成 Go test/race/vet、Web test/build、PostgreSQL integration、浏览器和仓库 fast/full。

## 完成结果

- exact refs、review digest、approval 阻断、decision CAS、scope、分页、migration、rollback、重启恢复与 no fallback 已落地。
- Web consumer、独立 lazy panel、offline 零请求和真实浏览器历史 / 详情 / decision 路径已完成。
- 本卡不继续派生 deployment、production authorization、自动执行或同层 checker；后续产品任务重新排位后另建功能设计。

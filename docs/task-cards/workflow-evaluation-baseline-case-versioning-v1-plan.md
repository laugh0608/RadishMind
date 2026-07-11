# Workflow Evaluation Baseline & Case Versioning v1 任务卡

更新时间：2026-07-11

状态：`workflow_evaluation_baseline_case_versioning_v1_completed`

## 目标

交付 evaluation case 不可变修订链、人工 baseline promotion、expected-version 并发控制、历史版本 review 和真实 Web 审查路径，同时保持 ref-only、no fallback 与禁止副作用边界。

## 实现范围

1. 将 case domain 演进为兼容 v1 的 `workflow_evaluation_case.v2` 完整快照，增加 version、revision kind、previous version 和服务端 change codes。
2. 扩展 memory / PostgreSQL evaluation repository，提供原子 revision CAS、current read、分页 revision history 和历史版本读取。
3. 增加 workflow-runs 0004 manual migration、既有 v1 回填、rollback / reapply 和 PostgreSQL integration coverage。
4. 增加 scoped revision create / list / detail API，并让既有 review 支持明确 version；稳定暴露 version conflict。
5. 扩展独立 Web consumer / lazy panel，覆盖修订、提升、冲突、历史审查与 offline 零请求。
6. 更新当前焦点、runtime 专题、周志和任务索引，按实现 / 文档主题提交。

## 门禁

- revision 必须是完整快照且原子 `expected_version` CAS；旧版本不可变、可读、可 review。
- baseline promotion 只能显式人工触发；所有引用 run 同 scope、eligible、互异且四类禁止副作用为 0。
- 不保存 input、condition、output、credential、endpoint、raw envelope、comparison snapshot 或自由文本理由。
- PostgreSQL 失败不得回退 memory；完成 Go test / race / vet、Web test / build、PostgreSQL migration / integration、真实浏览器、仓库 fast / full。

## 完成证据

- Go domain / memory / HTTP 覆盖 v1 兼容、完整快照修订、baseline promotion、稳定 family 排序、revision cursor、历史 review、scope、敏感字段和 16 路 CAS。
- PostgreSQL 0004 覆盖 fresh、v1 / v2 / v3 pending、既有 case 回填、rollback / reapply、runtime DDL、8 路 CAS、重启恢复、scope 与 no fallback。
- Web 19 项测试与 build 通过；evaluation lazy chunk 14.12 KiB，主入口保持 430.39 KiB，offline 零请求。
- 浏览器完成 v1→v5 审计链、人工 baseline promotion、外部并发 conflict 全量刷新和 Platform/Web 重启恢复；相关服务与容器已关闭。

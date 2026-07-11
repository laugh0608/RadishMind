# Workflow Run Comparison / Regression Review v1 任务卡

更新时间：2026-07-11

状态：`workflow_run_comparison_regression_review_v1_completed`

## 目标

交付同 scope durable run 的只读基线 / 候选比较 API 和 Workspace Run History 审查路径，以确定性规则识别状态回归、改善、其它变化、无变化与暂不可判断，并保持敏感材料和外部副作用关闭。

## 实现范围

1. 增加 comparison domain、稳定分类、节点差异、legacy / running 语义与 contract validation。
2. 增加 scoped nested GET route，复用既有 repository、auth、dev gate、failure envelope 和 audit。
3. 覆盖 memory 并发、scope、HTTP strict query、PostgreSQL 重启恢复和 no-fallback；不新增数据库 schema。
4. 增加独立 Web consumer 与 lazy comparison panel，接入真实 Run History，offline 不请求。
5. 拆分 React vendor chunk，把 R5 主包控制到 500 KiB 以下。
6. 更新功能专题、current focus、runtime 专题、周志和任务索引，并按实现 / 文档主题提交。

## 风险与门禁

- 任一 run 越界或不存在统一 not found；任一 repository 错误 fail closed。
- 相同 run id 和非 allowlist query 拒绝；不接受阈值、字段选择或任意比较表达式。
- 响应禁止 output / preview、input、condition、actor、credential、endpoint 和 raw envelope。
- 双方 tool / confirmation / business write / replay 必须为 0，否则 contract mismatch。
- 精准 Go / Web 测试通过后，补 PostgreSQL、真实浏览器、race / vet、Web build、fast / full。

## 完成证据

- Go domain / HTTP / memory 并发 / PostgreSQL 重启与 scope 集成测试通过；comparison 不新增 migration，复用现有 v0 / v1 durable record。
- Web consumer、禁止字段、offline、同 run、零副作用测试与 build 通过；主入口 430.39 KiB，comparison lazy chunk 6.96 KiB。
- 真实浏览器验证成功基线与 Gateway timeout 候选的 regression、节点差异、详情、人工动作和零副作用。
- Go test / race / vet、Web test / build、PostgreSQL integration 完成；本批服务与 Docker 资源已关闭。

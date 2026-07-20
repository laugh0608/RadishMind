# Application Configuration Draft & Review v1 任务卡

更新时间：2026-07-12

## 任务标识

- 任务 ID：`application-configuration-draft-review-v1`
- 状态：`completed`
- 对应功能专题：`docs/features/user-workspace/application-configuration-draft-review-v1.md`

## 用户目标

内部开发者可以从当前 application 建立、校验、保存、恢复和比较配置草案，在版本冲突时显式审查，并把有效的 application / protocol / model 交给现有 API Integration 与 Gateway Playground。

## 实现范围

1. 新增 application draft 独立领域、稳定失败、scope / owner 约束和 CAS。
2. 新增 dev-only validate / save / read / list routes，不修改 northbound Gateway schema。
3. 新增 memory dev 与 PostgreSQL dev/test repository，显式 migration 和 no fallback。
4. 新增 lazy Application Configuration Workspace，复用现有模型目录 consumer和 Playground handoff。
5. 补齐恢复、比较、冲突审查、application reset、secret guard 与 browser storage boundary。
6. 同步功能入口、当前焦点、路线图、能力矩阵和本周周志。

## 验收矩阵

| 场景 | 必须结果 |
| --- | --- |
| offline | 零 fetch，输入只在组件内存 |
| 本地校验 | 非法字段、协议和 secret 被阻止 |
| 首次 / 后续保存 | version 1 / CAS version + 1 |
| scope / owner 错误 | 不读取、不返回、不覆盖 |
| 冲突 | 返回当前版本，保留本地修改，要求显式处置 |
| store unavailable | 稳定失败且不回退 memory |
| application 切换 | 旧表单、目录、版本和冲突完全清除 |
| Integration handoff | 使用当前 application、已校验 protocol / model |
| browser | 保存、恢复、冲突、调用与 History 可连续复验且无存储泄漏 |

## 必要验证

- Web unit tests 与 production build
- Go unit / HTTP / PostgreSQL integration tests
- migration apply / rollback / reapply
- 真实浏览器验收与 console / URL / storage 检查
- `git diff --check`
- `./scripts/check-repo.sh --fast`
- `./scripts/check-repo.sh`

## 停止线

- 不修改正式 application 真相源，不开放 create / publish / delete。
- 不实现 production auth、API key、quota、billing、provider credential 或 endpoint 管理。
- 不复制 Gateway protocol adapter、SSE parser、request state 或 History repository。
- 不扩 Workflow 高风险执行能力。

## 完成记录

- 独立 domain、HTTP route、memory / PostgreSQL dev-test repository、migration runner、scope / owner / CAS 和 no-fallback 已实现。
- Web 已完成配置编辑、模型兼容校验、保存 / 列表 / 恢复、配置比较、双标签冲突审查和 API Integration / Playground handoff。
- PostgreSQL integration、Platform 全量 Go 测试、Web 49 项测试、production build 和真实浏览器连续流程均通过。
- 正式 application create / update / publish / delete、production auth、API key、quota 与 billing 继续关闭。

# Prompt / Agent 应用回归评测与发布审查（开发 / 测试态）v1 实施任务卡

更新时间：2026-07-25

状态：`prompt_agent_application_regression_evaluation_release_review_dev_test_v1_completed`

对应功能文档：[Prompt / Agent 应用回归评测与发布审查（开发 / 测试态）v1](../features/user-workspace/prompt-agent-application-regression-evaluation-release-review-dev-test-v1.md)

## 任务目标

在不新增 route、schema、migration、repository 或执行器的前提下，把 Prompt Application v6 与 Agent / Copilot v7 的 metadata-only Run 接入既有 Evaluation Case、Suite 和人工 decision 用户链。

## 实现边界

### 允许修改

- `services/platform/internal/httpapi/workflow_evaluation*.go`
- Evaluation / Comparison / Run History 相邻 Go 测试
- `apps/radishmind-web/src/features/control-plane-read/workflowEvaluation*.ts*`
- `apps/radishmind-web/src/features/control-plane-read/workflowRunHistoryPanel.tsx`
- 对应 Web consumer 测试
- 功能入口、当前焦点、路线图、能力矩阵、任务卡索引和本周周志

### 不允许修改

- Workflow Evaluation Case / Suite schema 与数据库表
- route、permission、store selector、migration、repository 接口
- Prompt / Agent invocation service、Gateway、provider、assignment 或 publish 状态机
- canonical `CopilotRequest / CopilotResponse`
- replay、batch execution、automatic release、production auth、quota、billing 或业务写回

## 冻结兼容矩阵

| Run record | Comparison | Run profile | Evaluation |
| --- | --- | --- | --- |
| v0–v2 | v1 | `workflow_standard.v1` | 保持既有语义 |
| v3 | v2 | `workflow_rag_retrieval.v1` | 保持既有语义 |
| v4 | v3 | `workflow_rag_application_invocation.v1` | 保持既有语义 |
| v5 | v4 | `workflow_definition_executor.v1` | 保持既有语义 |
| v6 | v5 | `prompt_application_invocation_v1` | 本任务补齐 Web strict review |
| v7 | v6 | `agent_copilot_suggestion_v1` | 本任务补齐 Go failure 与 Web strict review |

schema 与 run profile 必须按表精确配对。未知版本、跨版本组合、跨 Prompt Template lineage、跨 Agent Profile / project / task 和非零禁止副作用都失败关闭。

## 批次 A

1. Go：
   - 在 review 中映射 `WorkflowRunFailureAgentCopilotIncompatible`；
   - 为 Agent failure summary 增加稳定文本；
   - 补 Prompt / Agent create、review、revision compatibility 与负向测试；
   - 确认 suite review digest 对新 profile 稳定且 decision policy 不变。
2. Web：
   - Case decoder 接受 comparison v5 / v6 与对应 profile；
   - Suite decoder 接受 Prompt / Agent profile；
   - Run selector 显示 Prompt v6 与 Agent v7；
   - 非法 schema/profile 配对、未知字段与敏感字段继续拒绝。
3. 验证：
   - 相邻 Go / Web tests；
   - 定向 race、Web build、`go vet`；
   - `git diff --check`。

## 批次 B

1. 用现有 `--agent-copilot-local-product` 启动 SQLite 本地产品档。
2. 创建或选择 active Agent 应用，建立两个同 Profile / project / task 的 terminal v7 Run。
3. 浏览器创建预期为实际分类的 Evaluation Case，确认 review matched。
4. 创建 exact case-version Suite，确认 review digest 稳定并记录允许的人工 decision。
5. 离开页面后确认瞬态输入 / 回答清除，`localStorage` / `sessionStorage` 无敏感材料，console 0 error。
6. 同步阶段文档并运行 fast / full 仓库门禁。

## 完成定义

- Prompt v6 与 Agent v7 均可创建、修订并 review 既有 Evaluation Case。
- Suite strict consumer 可读取包含 Prompt / Agent case 的 review item。
- 跨 profile、跨 lineage、project / task drift、schema/profile 错配和禁止副作用均失败关闭。
- 真实 SQLite 浏览器完成 Agent v7 case → suite → decision。
- 没有新增 schema、API、migration、repository、provider 调用或自动发布能力。
- 文档状态推进为 `prompt_agent_application_regression_evaluation_release_review_dev_test_v1_completed`。

## 完成记录

- 批次 A 已完成 Go Agent 不兼容失败映射、Prompt / Agent Case 与 Suite 邻接测试、Web Comparison/Profile 精确配对、Agent v7 标签、229 项 Web 测试、生产构建、定向 race 和 `go vet`。
- 批次 B 已在 SQLite 本地产品档完成 Agent v7 Run → Comparison v6 → `eval_46be5143a56cadd4d15b18e0` → `suite_a40a6cf0a739a3f27f332eb9` → `approved v1`。
- suite item profile 为 `agent_copilot_suggestion_v1`，review 为 `passed`；浏览器 storage 为空、console 0 error，开发进程已关闭。
- 实现仅扩展既有 Evaluation owner 与 strict consumer，没有新增 route、schema、migration、repository、provider 调用或自动发布状态迁移。

# Workflow / Agent Runtime 设计与开发文档

更新时间：2026-06-14

## 功能定位

`Workflow / Agent Runtime` 承载 AI 应用执行链路，包括 Prompt、LLM、HTTP tool、RAG retrieval、condition、output、后续受控 code / sandbox 和 agent loop。

## 当前状态

- `apps/radishmind-web/` 已有 workflow application detail、definition detail、run detail、blocked action preview、confirmation placeholder、draft designer、validation inspector、execution plan preview、runtime readiness inspector、surface overview、context selection、scenario inspector、Review Workspace、Workspace Home 和 Review Handoff。
- 当前能力全部是 offline-only / read-only / advisory-only / blocked capability 组织层。
- `workflowWorkspaceContext` 已作为本地选择和派生关系的共享入口。
- 当前没有 builder persistence、draft persistence、validation result persistence、execution plan persistence、runtime readiness persistence、workflow executor、confirmation decision store、writeback、replay 或 resume。

## 设计边界

- draft、validation、plan、readiness、review 和 execution 是不同阶段，不能在 UI 中混成一个“已可执行”状态。
- 高风险 tool/action 默认要求 confirmation。
- executor 之前必须先有 run record、audit、failure taxonomy、materialized result 和 no side effects 策略。
- 上层 `RadishFlow` / `Radish` 未提供挂载点时，不阻塞本仓库设计离线草案、审查和 readiness，但不能声明真实接入 ready。

## 下一批开发方向

1. 下一步优先从功能设计角度定义“可保存 workflow draft”和“可执行 run”的最小产品路径。
2. 若继续优化现有只读审查组织，不再新增逐项 task card / fixture / checker。
3. 若进入 draft persistence、publish、executor、confirmation decision 或 writeback，必须先更新本功能文档，再拆 task card。
4. executor 不能和 confirmation、writeback、replay 一次性打开；每个能力都需要独立失败语义和回归策略。

## 验收方式

- 只读 surface：web build、聚合 workflow checker、fast baseline。
- draft / validation persistence：schema、unit tests、negative tests、storage boundary checks。
- executor / confirmation：专门 task card、audit tests、no side effects tests、人工确认路径和全量仓库验证。

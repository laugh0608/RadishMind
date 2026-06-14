# User Workspace 设计与开发文档

更新时间：2026-06-14

## 功能定位

`User Workspace` 是 RadishMind 面向终端用户和项目成员的主工作区。它用于查看和管理 AI 应用、Prompt 应用、Workflow、Agent / Copilot 应用、API key、调用量、运行记录、成本摘要和人工审查入口。

## 当前状态

- `apps/radishmind-web/` 已有 read-side shell、Workspace Home、applications、API keys、usage quota、workflow definitions、run history、Workflow Review Workspace 和 Workflow Review Handoff。
- dev-only live read consumer 只能在显式 opt-in 下读取 fake-store-backed Go handlers。
- `ControlPlaneReadRepository` interface 已落地，七条 read handlers 已通过 fake-store repository bridge 消费数据。
- 当前仍不具备 production API consumer、真实数据库、Radish OIDC、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

## 设计边界

- 用户端默认只输出建议、解释、审查包和候选动作，不直接写业务真相源。
- 高风险动作必须保留 `requires_confirmation`。
- read-side 与未来 write/execution side 必须分开设计；展示 ready 不等于执行 ready。
- API key 页面不得展示 key value、hash、authorization header 或 secret material。

## 下一批开发方向

1. 继续整理真实用户工作流：从“看应用 / 看运行 / 看审查包”走向“创建或选择一个可执行应用草案”。
2. 在进入任何写入前，先补用户工作区功能设计更新，明确创建、保存、发布、执行、确认和回滚边界。
3. 若下一步只改展示、分组、文案或使用性，不新增专项 gate，复用 web build、consumer smoke 和仓库基线。
4. 若新增 API、写入、真实 auth、真实数据源或执行能力，必须新增 task card，并按风险补 fixture / checker。

## 验收方式

- 功能展示类：`npm run build`、必要浏览器布局检查、`./scripts/check-repo.sh --fast`。
- read contract 类：consumer smoke、Go handler tests、read-side contract checker。
- 写入或执行类：先补设计文档和 task card，再补单测、负向测试、仓库级检查和人工确认路径。

# Workflow Executor v0 实现任务卡

更新时间：2026-07-11

状态：`workflow_executor_v0_implementation_in_progress`

## 目标

按照 [Workflow Executor v0](../features/workflow/workflow-executor-v0.md) 落地一个开发 / 测试态受控执行批次：从已保存草案读取 `prompt`、`llm`、`condition`、`output` 图，经现有 Gateway 运行，生成可重新读取的有界 run record，并在 Web 提供草案创建、执行与结果审查入口。

## 前置条件

- `Saved Workflow Draft v1` 和 PostgreSQL dev/test repository 已完成。
- Gateway R4 与 stdio worker pool 已完成，executor 必须复用 `bridgeClient` 与既有 provider selection。
- 功能设计已固定节点语义、图准入、状态、failure code、预算、副作用和停止线。
- 本批只在 RadishMind 工作区修改，不触碰外部项目工作区。

## 实施范围

### 1. Platform domain 与 store

- 新增 executor request、run / node state、failure code、run record、side effect count 和 bounded memory store。
- 新增草案资格与 DAG 检查、稳定拓扑排序、条件路由、节点状态推进和总执行时限。
- Prompt 只组合有界说明与输入；LLM 只调用现有 Gateway；Output 只形成 advisory result。
- running 与 terminal record 都经过 scope-aware store，最多保留 100 条。

### 2. Dev/test HTTP 与配置

- 新增 POST run 和 GET run record 路由、strict JSON、scope 映射、sanitized envelope 与 CORS。
- 新增默认关闭的 `RADISHMIND_WORKFLOW_EXECUTOR_DEV`，进入 config summary / source / tests。
- dev launcher 只在显式 saved-draft live mode 下同时打开 server gate 和 Web consumer source。

### 3. Web consumer 与交互

- 新建独立 consumer 文件，不继续扩大已超过 1200 行的 saved draft consumer。
- 从当前 application 派生新的 `Prompt → LLM → Output` executor v0 草案，保留草案编辑 / 保存流程。
- 对已保存且未修改的 active draft执行本地 eligibility；服务端仍重新检查。
- 展示 input、condition 值、运行状态、run / node 记录、Gateway selection、output、failure 和副作用计数。
- 支持用最新 run id 重新读取 record。

### 4. 测试与文档

- Go domain、store、HTTP、config 和 side-effect negative tests。
- Web consumer 单元测试与 build；只在现有页面结构中新增必要样式，不拆新的 gate-only checker 链。
- 更新当前焦点、Workflow 大专题 / 入口、当前阶段停止线和周志。

## 明确不做

- 不执行 tool、RAG、code、sandbox、agent loop。
- 不实现 confirmation decision、writeback、publish、replay / resume 或 materialized result reader。
- 不把 Saved Draft PostgreSQL repository 扩展为 run repository。
- 不增加 production auth、production route、production enablement 或真实外部项目联调。
- 不保存原始运行输入、condition 值、raw provider envelope、credential 或 endpoint。

## 分步验证

1. domain / store：`go test ./services/platform/internal/httpapi -run 'WorkflowExecutor|WorkflowRun'`。
2. HTTP / config：`go test ./services/platform/internal/config ./services/platform/internal/httpapi`。
3. 并发与静态检查：`go test -race ./services/platform/internal/httpapi`、`go vet ./services/platform/...`。
4. Web：在 `apps/radishmind-web` 执行现有 test 与 build 入口。
5. 仓库：`./scripts/check-repo.sh --fast`，最终 `./scripts/check-repo.sh`。

## 完成条件

- 成功路径能够从保存草案运行到终态 record，并通过 GET 重新读取。
- condition 分支有显式输入与 skipped evidence；无活动 output 必须失败。
- 所有拒绝路径的 Gateway call count 为 0，所有路径的 tool / confirmation / business write / replay count 为 0。
- Web 能创建合规草案、保存后执行、审查结果和重新读取记录。
- 相关精准验证、fast 和 full 门禁通过，文档与周志同步完成。

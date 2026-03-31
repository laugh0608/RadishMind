# RadishMind 统一契约文件

更新时间：2026-03-31

本目录承载 `RadishMind` 第一版真实契约文件。

当前阶段先以 `JSON Schema` 作为仓库内的正式契约形式，原因是：

- 便于在不同语言和服务栈之间共享
- 便于在数据集、回归脚本和适配器层复用
- 能先冻结结构边界，再决定是否生成 TypeScript 或其他语言类型

当前文件：

1. `copilot-request.schema.json`
2. `copilot-response.schema.json`

使用原则：

- 文档说明以 [docs/radishmind-integration-contracts.md](../docs/radishmind-integration-contracts.md) 为语义说明入口
- `contracts/` 中的 schema 是程序化校验入口
- 当前 schema 只冻结通用骨架与最小项目上下文字段，不把所有任务细节一次写死
- 任务级最小输入和风险规则以 [docs/task-cards/README.md](../docs/task-cards/README.md) 为准
- 当前 `Radish` 文档问答回归会直接复用这两份 schema 校验 `input_request` 与 `golden_response`，再叠加任务级召回边界与输出对照规则

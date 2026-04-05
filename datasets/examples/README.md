# RadishMind 契约示例

更新时间：2026-04-05

本目录用于存放可直接被 schema 校验的最小示例对象。

当前用途：

- 为新增契约提供稳定、可执行验证的示例输入
- 避免 schema 长期只有结构说明，没有真实实例参照
- 让 `check-repo` 同时校验“schema 本身合法”和“示例能通过 schema”

当前示例：

1. `radishflow-ghost-candidate-set-flash-basic-001.json`

说明：

- 该示例对应 [radishflow-ghost-candidate-set.schema.json](../../contracts/radishflow-ghost-candidate-set.schema.json)
- 它表示“本地规则层已为一个 `FlashDrum` 生成候选 ghost completion 集”的最小交接对象
- 该示例不是模型响应，也不是最终 `CopilotRequest`，而是 pre-model 候选集对象

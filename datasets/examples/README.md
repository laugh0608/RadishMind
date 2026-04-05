# RadishMind 契约示例

更新时间：2026-04-05

本目录用于存放可直接被 schema 校验的最小示例对象。

当前用途：

- 为新增契约提供稳定、可执行验证的示例输入
- 避免 schema 长期只有结构说明，没有真实实例参照
- 让 `check-repo` 同时校验“schema 本身合法”和“示例能通过 schema”

当前示例：

1. `radishflow-ghost-candidate-set-flash-basic-001.json`
2. `radishflow-copilot-request-ghost-flash-basic-001.json`
3. `radishflow-copilot-request-ghost-flash-basic-001-debug-full.json`
4. `radishflow-ghost-candidate-set-valve-ambiguous-001.json`
5. `radishflow-copilot-request-ghost-valve-ambiguous-001.json`
6. `radishflow-copilot-request-ghost-valve-ambiguous-001-debug-full.json`

说明：

- 该示例对应 [radishflow-ghost-candidate-set.schema.json](../../contracts/radishflow-ghost-candidate-set.schema.json)
- 它表示“本地规则层已为一个 `FlashDrum` 生成候选 ghost completion 集”的最小交接对象
- 该示例不是模型响应，也不是最终 `CopilotRequest`，而是 pre-model 候选集对象

- `radishflow-copilot-request-ghost-flash-basic-001.json`
  - 该示例对应 [copilot-request.schema.json](../../contracts/copilot-request.schema.json)
  - 它表示“候选集经适配层装配后发给模型”的最小 `CopilotRequest`
  - 当前应可由 [build-radishflow-ghost-request.py](../../scripts/build-radishflow-ghost-request.py) 以默认 `model-minimal` profile 从上面的候选集示例稳定生成

- `radishflow-copilot-request-ghost-flash-basic-001-debug-full.json`
  - 该示例对应同一份 [copilot-request.schema.json](../../contracts/copilot-request.schema.json)
  - 它表示同一候选集在 `debug-full` profile 下的调试装配结果
  - 当前用于固定“默认最小透传”和“完整本地证据透传”的对照边界，并由 `check-repo` 校验输出不漂移

- `radishflow-ghost-candidate-set-valve-ambiguous-001.json`
  - 该示例对应 [radishflow-ghost-candidate-set.schema.json](../../contracts/radishflow-ghost-candidate-set.schema.json)
  - 它表示 `Valve` 出口存在两个接近合法下游时的 pre-model 候选集
  - 当前用于固定“允许展示 ghost，但不默认 Tab”的本地规则层交接边界

- `radishflow-copilot-request-ghost-valve-ambiguous-001.json`
  - 该示例对应 [copilot-request.schema.json](../../contracts/copilot-request.schema.json)
  - 它表示上述 `Valve ambiguous` 候选集在默认 `model-minimal` profile 下的模型请求
  - 当前用于固定“歧义场景默认只透传最小候选信息”的请求边界

- `radishflow-copilot-request-ghost-valve-ambiguous-001-debug-full.json`
  - 该示例对应同一份 [copilot-request.schema.json](../../contracts/copilot-request.schema.json)
  - 它表示上述 `Valve ambiguous` 候选集在 `debug-full` profile 下的调试请求
  - 当前用于保留歧义评分与冲突标记，便于对照 `model-minimal` 请求中被裁剪的字段

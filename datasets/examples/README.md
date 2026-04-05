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
7. `radishflow-ghost-candidate-set-chain-feed-valve-flash-flash-outlets-001.json`
8. `radishflow-copilot-request-ghost-chain-feed-valve-flash-flash-outlets-001.json`
9. `radishflow-copilot-request-ghost-chain-feed-valve-flash-flash-outlets-001-debug-full.json`
10. `radishflow-ghost-candidate-set-chain-feed-valve-flash-stop-no-legal-outlet-001.json`
11. `radishflow-copilot-request-ghost-chain-feed-valve-flash-stop-no-legal-outlet-001.json`
12. `radishflow-copilot-request-ghost-chain-feed-valve-flash-stop-no-legal-outlet-001-debug-full.json`
13. `radishflow-ghost-candidate-set-chain-feed-valve-flash-outlets-name-conflict-no-tab-001.json`
14. `radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-name-conflict-no-tab-001.json`
15. `radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-name-conflict-no-tab-001-debug-full.json`
16. `radishflow-ghost-candidate-set-chain-feed-heater-flash-heater-outlet-001.json`
17. `radishflow-copilot-request-ghost-chain-feed-heater-flash-heater-outlet-001.json`
18. `radishflow-copilot-request-ghost-chain-feed-heater-flash-heater-outlet-001-debug-full.json`
19. `radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-name-conflict-no-tab-001.json`
20. `radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-name-conflict-no-tab-001.json`
21. `radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-name-conflict-no-tab-001-debug-full.json`
22. `radishflow-ghost-candidate-set-chain-feed-heater-flash-stop-no-legal-outlet-001.json`
23. `radishflow-copilot-request-ghost-chain-feed-heater-flash-stop-no-legal-outlet-001.json`
24. `radishflow-copilot-request-ghost-chain-feed-heater-flash-stop-no-legal-outlet-001-debug-full.json`
25. `radishflow-ghost-candidate-set-chain-feed-cooler-flash-cooler-outlet-001.json`
26. `radishflow-copilot-request-ghost-chain-feed-cooler-flash-cooler-outlet-001.json`
27. `radishflow-copilot-request-ghost-chain-feed-cooler-flash-cooler-outlet-001-debug-full.json`
28. `radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001.json`
29. `radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001.json`
30. `radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001-debug-full.json`
31. `radishflow-ghost-candidate-set-chain-feed-cooler-flash-stop-no-legal-outlet-001.json`
32. `radishflow-copilot-request-ghost-chain-feed-cooler-flash-stop-no-legal-outlet-001.json`
33. `radishflow-copilot-request-ghost-chain-feed-cooler-flash-stop-no-legal-outlet-001-debug-full.json`

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

- `radishflow-ghost-candidate-set-chain-feed-valve-flash-flash-outlets-001.json`
  - 该示例对应 [radishflow-ghost-candidate-set.schema.json](../../contracts/radishflow-ghost-candidate-set.schema.json)
  - 它表示 `Feed -> Valve -> FlashDrum` 连续搭建链中，`Valve -> FlashDrum` 入口 ghost 刚被接受后，下一步默认建议转向 `FlashDrum` 标准出口补全的 pre-model 候选集
  - 当前用于把 `cursor_context.recent_actions` 从 `eval` 样本推进到 `examples/` 基线，固定“上一步已接受 ghost 会影响下一步默认建议”的 handoff 口径

- `radishflow-copilot-request-ghost-chain-feed-valve-flash-flash-outlets-001.json`
  - 该示例对应 [copilot-request.schema.json](../../contracts/copilot-request.schema.json)
  - 它表示上述连续搭建链候选集在默认 `model-minimal` profile 下的模型请求
  - 当前用于固定 recent actions、命名提示与最小候选摘要会如何一起装配进模型请求

- `radishflow-copilot-request-ghost-chain-feed-valve-flash-flash-outlets-001-debug-full.json`
  - 该示例对应同一份 [copilot-request.schema.json](../../contracts/copilot-request.schema.json)
  - 它表示上述连续搭建链候选集在 `debug-full` profile 下的调试请求
  - 当前用于保留 outlet 排序证据、命名证据与几何细节，方便对照 `model-minimal` 裁剪后的连续搭建链请求边界

- `radishflow-ghost-candidate-set-chain-feed-valve-flash-stop-no-legal-outlet-001.json`
  - 该示例对应 [radishflow-ghost-candidate-set.schema.json](../../contracts/radishflow-ghost-candidate-set.schema.json)
  - 它表示 `Feed -> Valve -> FlashDrum` 连续搭建链里，虽然最近两步 ghost 已连续接受，但当前 `FlashDrum` outlet 不存在任何合法候选的 pre-model 候选集
  - 当前用于固定“recent_actions 已存在，但 local rules 要求停住并返回空建议”的 handoff 边界

- `radishflow-copilot-request-ghost-chain-feed-valve-flash-stop-no-legal-outlet-001.json`
  - 该示例对应 [copilot-request.schema.json](../../contracts/copilot-request.schema.json)
  - 它表示上述链式停住候选集在默认 `model-minimal` profile 下的模型请求
  - 当前用于固定空候选链路里 `recent_actions`、`naming_hints` 与裁剪后的邻近节点仍会如何进入请求

- `radishflow-copilot-request-ghost-chain-feed-valve-flash-stop-no-legal-outlet-001-debug-full.json`
  - 该示例对应同一份 [copilot-request.schema.json](../../contracts/copilot-request.schema.json)
  - 它表示上述链式停住候选集在 `debug-full` profile 下的调试请求
  - 当前用于保留邻近节点的阻塞原因与几何细节，避免“为何停住”只留在口头约定

- `radishflow-ghost-candidate-set-chain-feed-valve-flash-outlets-name-conflict-no-tab-001.json`
  - 该示例对应 [radishflow-ghost-candidate-set.schema.json](../../contracts/radishflow-ghost-candidate-set.schema.json)
  - 它表示 `Feed -> Valve -> FlashDrum` 连续搭建链里，`FlashDrum` 的 outlet 候选已经存在，但建议命名与现有流股冲突，因此只能显示可见 ghost 而不能默认 `Tab`
  - 当前用于固定“recent_actions 已存在、legal candidates 非空，但仍需 manual-only”的 handoff 边界

- `radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-name-conflict-no-tab-001.json`
  - 该示例对应 [copilot-request.schema.json](../../contracts/copilot-request.schema.json)
  - 它表示上述链式命名冲突候选集在默认 `model-minimal` profile 下的模型请求
  - 当前用于固定 non-tab 链式候选在最小请求里仍会保留的字段集合

- `radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-name-conflict-no-tab-001-debug-full.json`
  - 该示例对应同一份 [copilot-request.schema.json](../../contracts/copilot-request.schema.json)
  - 它表示上述链式命名冲突候选集在 `debug-full` profile 下的调试请求
  - 当前用于保留 naming_signals 与 conflict_flags，避免“为何不能默认 Tab”再次退回成口头说明

- `radishflow-ghost-candidate-set-chain-feed-heater-flash-heater-outlet-001.json`
  - 该示例对应 [radishflow-ghost-candidate-set.schema.json](../../contracts/radishflow-ghost-candidate-set.schema.json)
  - 它表示 `Feed -> Heater -> FlashDrum` 连续搭建链里，`Feed -> Heater` 入口 ghost 刚被接受后，下一步默认建议转向 `Heater` 的 outlet 接入 `FlashDrum` inlet 的 pre-model 候选集
  - 当前用于验证 `recent_actions` 与 `topology_pattern_hints` 的链式作用并未写死在 `Valve` 模板上

- `radishflow-copilot-request-ghost-chain-feed-heater-flash-heater-outlet-001.json`
  - 该示例对应 [copilot-request.schema.json](../../contracts/copilot-request.schema.json)
  - 它表示上述 `Feed -> Heater -> FlashDrum` 链式候选集在默认 `model-minimal` profile 下的模型请求
  - 当前用于固定第二条链式模板在最小请求里的装配边界

- `radishflow-copilot-request-ghost-chain-feed-heater-flash-heater-outlet-001-debug-full.json`
  - 该示例对应同一份 [copilot-request.schema.json](../../contracts/copilot-request.schema.json)
  - 它表示上述 `Feed -> Heater -> FlashDrum` 链式候选集在 `debug-full` profile 下的调试请求
  - 当前用于保留几何评分与命名证据，便于对照第二条链式模板的装配结果

- `radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-name-conflict-no-tab-001.json`
  - 该示例对应 [radishflow-ghost-candidate-set.schema.json](../../contracts/radishflow-ghost-candidate-set.schema.json)
  - 它表示 `Feed -> Heater -> FlashDrum` 连续搭建链里，`Heater` outlet 候选已经存在，但建议命名与现有流股冲突，因此只能显示可见 ghost 而不能默认 `Tab`
  - 当前用于验证第二条模板同样支持 `manual_only` 边界，而不是只存在顺风正例

- `radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-name-conflict-no-tab-001.json`
  - 该示例对应 [copilot-request.schema.json](../../contracts/copilot-request.schema.json)
  - 它表示上述第二模板命名冲突候选集在默认 `model-minimal` profile 下的模型请求
  - 当前用于固定第二模板 `manual_only` 候选在最小请求里的装配边界

- `radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-name-conflict-no-tab-001-debug-full.json`
  - 该示例对应同一份 [copilot-request.schema.json](../../contracts/copilot-request.schema.json)
  - 它表示上述第二模板命名冲突候选集在 `debug-full` profile 下的调试请求
  - 当前用于保留第二模板的 naming_signals 与 conflict_flags，避免“为何不能默认 Tab”再次退回成口头说明

- `radishflow-ghost-candidate-set-chain-feed-heater-flash-stop-no-legal-outlet-001.json`
  - 该示例对应 [radishflow-ghost-candidate-set.schema.json](../../contracts/radishflow-ghost-candidate-set.schema.json)
  - 它表示 `Feed -> Heater -> FlashDrum` 连续搭建链里，`Feed -> Heater` 入口 ghost 刚被接受后，当前 `Heater` outlet 不存在任何合法候选的 pre-model 候选集
  - 当前用于固定第二条模板里“recent_actions 已存在，但 local rules 仍要求停住并返回空建议”的 handoff 边界

- `radishflow-copilot-request-ghost-chain-feed-heater-flash-stop-no-legal-outlet-001.json`
  - 该示例对应 [copilot-request.schema.json](../../contracts/copilot-request.schema.json)
  - 它表示上述第二模板链式停住候选集在默认 `model-minimal` profile 下的模型请求
  - 当前用于固定第二模板空候选链路里 `recent_actions`、`naming_hints` 与裁剪后的邻近节点仍会如何进入请求

- `radishflow-copilot-request-ghost-chain-feed-heater-flash-stop-no-legal-outlet-001-debug-full.json`
  - 该示例对应同一份 [copilot-request.schema.json](../../contracts/copilot-request.schema.json)
  - 它表示上述第二模板链式停住候选集在 `debug-full` profile 下的调试请求
  - 当前用于保留第二模板里邻近节点的阻塞原因与几何细节，避免“为何停住”再次退回成口头约定

- `radishflow-ghost-candidate-set-chain-feed-cooler-flash-cooler-outlet-001.json`
  - 该示例对应 [radishflow-ghost-candidate-set.schema.json](../../contracts/radishflow-ghost-candidate-set.schema.json)
  - 它表示 `Feed -> Cooler -> FlashDrum` 连续搭建链里，`Feed -> Cooler` 入口 ghost 刚被接受后，下一步默认建议转向 `Cooler` 的 outlet 接入 `FlashDrum` inlet 的 pre-model 候选集
  - 当前用于验证第三条链式模板同样可以复用 `recent_actions` handoff、最小装配和默认 `Tab` 口径

- `radishflow-copilot-request-ghost-chain-feed-cooler-flash-cooler-outlet-001.json`
  - 该示例对应 [copilot-request.schema.json](../../contracts/copilot-request.schema.json)
  - 它表示上述 `Feed -> Cooler -> FlashDrum` 链式候选集在默认 `model-minimal` profile 下的模型请求
  - 当前用于固定第三条链式模板在最小请求里的装配边界

- `radishflow-copilot-request-ghost-chain-feed-cooler-flash-cooler-outlet-001-debug-full.json`
  - 该示例对应同一份 [copilot-request.schema.json](../../contracts/copilot-request.schema.json)
  - 它表示上述 `Feed -> Cooler -> FlashDrum` 链式候选集在 `debug-full` profile 下的调试请求
  - 当前用于保留第三条链式模板的几何评分与命名证据，便于对照最小请求的裁剪边界

- `radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001.json`
  - 该示例对应 [radishflow-ghost-candidate-set.schema.json](../../contracts/radishflow-ghost-candidate-set.schema.json)
  - 它表示 `Feed -> Cooler -> FlashDrum` 连续搭建链里，`Cooler` outlet 候选已经存在，但建议命名与现有流股冲突，因此只能显示可见 ghost 而不能默认 `Tab`
  - 当前用于验证第三条模板同样支持 `manual_only` 边界，而不是只存在顺风正例

- `radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001.json`
  - 该示例对应 [copilot-request.schema.json](../../contracts/copilot-request.schema.json)
  - 它表示上述第三模板命名冲突候选集在默认 `model-minimal` profile 下的模型请求
  - 当前用于固定第三模板 `manual_only` 候选在最小请求里的装配边界

- `radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001-debug-full.json`
  - 该示例对应同一份 [copilot-request.schema.json](../../contracts/copilot-request.schema.json)
  - 它表示上述第三模板命名冲突候选集在 `debug-full` profile 下的调试请求
  - 当前用于保留第三模板的 `naming_signals` 与 `conflict_flags`，避免“为何不能默认 Tab”再次退回成口头说明

- `radishflow-ghost-candidate-set-chain-feed-cooler-flash-stop-no-legal-outlet-001.json`
  - 该示例对应 [radishflow-ghost-candidate-set.schema.json](../../contracts/radishflow-ghost-candidate-set.schema.json)
  - 它表示 `Feed -> Cooler -> FlashDrum` 连续搭建链里，`Feed -> Cooler` 入口 ghost 刚被接受后，当前 `Cooler` outlet 不存在任何合法候选的 pre-model 候选集
  - 当前用于固定第三条模板里“recent_actions 已存在，但 local rules 仍要求停住并返回空建议”的 handoff 边界

- `radishflow-copilot-request-ghost-chain-feed-cooler-flash-stop-no-legal-outlet-001.json`
  - 该示例对应 [copilot-request.schema.json](../../contracts/copilot-request.schema.json)
  - 它表示上述第三模板链式停住候选集在默认 `model-minimal` profile 下的模型请求
  - 当前用于固定第三模板空候选链路里 `recent_actions`、`naming_hints` 与裁剪后的邻近节点仍会如何进入请求

- `radishflow-copilot-request-ghost-chain-feed-cooler-flash-stop-no-legal-outlet-001-debug-full.json`
  - 该示例对应同一份 [copilot-request.schema.json](../../contracts/copilot-request.schema.json)
  - 它表示上述第三模板链式停住候选集在 `debug-full` profile 下的调试请求
  - 当前用于保留第三模板里邻近节点的阻塞原因与几何细节，避免“为何停住”再次退回成口头约定

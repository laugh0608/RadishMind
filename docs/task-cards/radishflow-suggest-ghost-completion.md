# `RadishFlow` 任务卡：`suggest_ghost_completion`

更新时间：2026-04-05

## 任务目标

基于 `RadishFlow` 当前画布中的单元放置、canonical ports、本地规则筛出的合法候选和邻近拓扑上下文，输出 1 到 3 条可供前端渲染为灰态 ghost 的补全建议。

当前任务关注的是“编辑器内 pending 建议”，不是正式 patch，也不是直接改图。所有输出都必须先经过本地规则系统过滤，并在用户接受后继续走本地命令与校验流程。

## 请求映射

- `project`: `radishflow`
- `task`: `suggest_ghost_completion`
- 请求结构遵循 [CopilotRequest schema](../../contracts/copilot-request.schema.json)

## 任务定位

本任务和 [`suggest_flowsheet_edits`](radishflow-suggest-flowsheet-edits.md) 不同：

- `suggest_flowsheet_edits` 面向正式候选编辑提案，默认会影响上层真相源，因此必须 `requires_confirmation: true`
- `suggest_ghost_completion` 面向前端灰态占位补全，接受前只是 pending ghost，不应被视为正式写回

因此本任务优先解决：

- 放置标准单元后，下一个最可能补全的 canonical port
- 哪个邻近节点或物流线最适合作为 ghost connection 候选
- 哪个 ghost stream name 最符合当前局部命名习惯

## 最小必需输入

请求至少应包含以下内容：

- `artifacts` 中存在主输入 `flowsheet_document`
- `context.document_revision`
- `context.selected_unit_ids`，且当前任务应只聚焦一个被放置或当前激活的单元
- `context.unconnected_ports` 或 `context.missing_canonical_ports`
- `context.legal_candidate_completions`

若没有本地规则层提供的合法候选集，本任务应优先返回空建议，而不是凭主观猜测拼接拓扑。

## 推荐补充输入

- `context.selected_unit`
- `context.nearby_nodes`
- `context.cursor_context`
- `context.naming_hints`
- `context.topology_pattern_hints`

推荐先由本地规则层把“合法候选空间”压缩出来，再由模型做排序、命名和空结果判断。

## 禁止透传

以下内容不应进入本任务请求：

- 任何安全凭据、token、cookie 或本机秘密引用
- 绕过本地规则系统才能成立的候选连接
- 未经裁剪的整页无关状态和与当前单元无关的大范围拓扑噪声

## 第一阶段优先支持

当前阶段优先覆盖标准化程度高的单元：

- `FlashDrum`
- `Mixer`
- `Valve`
- `Heater`
- `Cooler`
- `Feed`

其中第一优先是 `FlashDrum`：

1. `inlet`
2. `vapor_outlet`
3. `liquid_outlet`

## 输出要求

响应结构遵循 [CopilotResponse schema](../../contracts/copilot-response.schema.json)。

本任务对字段的要求如下：

- `summary` 必须说明当前是否存在高置信 ghost 建议，以及建议围绕哪个单元生成
- `proposed_actions` 可以为空；若存在建议，数量应限制在 1 到 3 条
- 每条建议都必须使用 `kind: "ghost_completion"`
- 每条 `ghost_completion` 都必须包含 `title`、`target`、`rationale`、`patch`、`preview`、`apply`、`risk_level`、`requires_confirmation`
- `citations` 必须能回溯到当前单元、缺失端口、本地候选集或邻近拓扑证据

## `ghost_completion` 语义

`ghost_completion` 不是直接 patch，而是可供前端渲染的待接受占位建议。

建议使用以下结构表达：

- `target`
  - 指向 unit、port 或局部 candidate slot
- `patch`
  - 只描述 ghost 语义，不描述正式文档写回
  - 推荐包含 `ghost_kind`、`target_port_key`、`candidate_ref`、`ghost_stream_name`
- `preview`
  - 前端渲染提示
  - 推荐包含 `ghost_color`、`accept_key`、`render_priority`
- `apply`
  - 用户接受 ghost 后交给本地命令层执行的命令描述
  - 推荐包含 `command_kind` 与 `payload`

## 风险与确认规则

本任务默认只输出低副作用 ghost 建议。

- `low`
  - 标准 canonical port 补全
  - 标准 ghost connection
  - 低歧义命名补全
- `medium`
  - 多候选接近、但仍能给出一条领先建议的局部补全
- `high`
  - 当前阶段原则上不应输出；若候选具有明显拓扑副作用或规则冲突风险，应直接返回空建议

本任务默认 `requires_confirmation: false`，因为：

- 建议本身只是 pending ghost
- 用户接受后仍要经过本地命令系统与本地校验
- 这一步不是绕过规则的直接写回

若某条建议只有在越过本地规则前提下才能成立，则不应输出。

## `Tab` 接受规则

本任务不应默认把“第一条建议”直接等同于“可按 Tab 接受”。

建议至少满足以下条件之一，前端才将第一条视为默认 `Tab` 候选：

- 本地规则层已将其标记为唯一高置信合法候选
- 或模型排序后显著领先第二候选，且无本地 conflict 标记

若不存在高置信候选，应返回空建议或仅返回可见 ghost，但不应强推默认接受。

## 正反例边界

正例：

- `FlashDrum` 刚放置后，针对缺失的 `vapor_outlet` 返回 ghost connection 建议
- `Mixer` 已有两个入口候选时，只返回本地规则认可的那个标准出口补全
- 在合法候选集中为新 ghost stream 生成符合当前命名习惯的补全名

反例：

- 直接声称“已为你创建并落盘”
- 忽略本地 canonical ports，自行发明一个不存在的 port
- 明明上下文不足，还输出看起来聪明但不可验证的连线猜测
- 把高副作用拓扑调整伪装成低风险 ghost 建议

## 最小输入快照示例

```json
{
  "project": "radishflow",
  "task": "suggest_ghost_completion",
  "artifacts": [
    {
      "kind": "json",
      "role": "primary",
      "name": "flowsheet_document",
      "mime_type": "application/json",
      "uri": "artifact://flowsheet/current"
    }
  ],
  "context": {
    "document_revision": 31,
    "selected_unit_ids": ["U-12"],
    "selected_unit": {
      "id": "U-12",
      "kind": "FlashDrum"
    },
    "unconnected_ports": [
      "inlet",
      "vapor_outlet",
      "liquid_outlet"
    ],
    "legal_candidate_completions": [
      {
        "candidate_ref": "cand-vapor-stub",
        "ghost_kind": "ghost_connection",
        "target_port_key": "vapor_outlet"
      },
      {
        "candidate_ref": "cand-liquid-stub",
        "ghost_kind": "ghost_connection",
        "target_port_key": "liquid_outlet"
      }
    ]
  }
}
```

## 候选输出片段示例

```json
{
  "kind": "ghost_completion",
  "title": "补全 FlashDrum 的 vapor outlet ghost 连线",
  "target": {
    "type": "unit_port",
    "unit_id": "U-12",
    "port_key": "vapor_outlet"
  },
  "rationale": "FlashDrum 的 canonical ports 中 vapor outlet 尚未连接，且本地规则已给出合法 ghost connection 候选。",
  "patch": {
    "ghost_kind": "ghost_connection",
    "candidate_ref": "cand-vapor-stub",
    "target_port_key": "vapor_outlet",
    "ghost_stream_name": "V-12"
  },
  "preview": {
    "ghost_color": "gray",
    "accept_key": "Tab",
    "render_priority": 1
  },
  "apply": {
    "command_kind": "accept_ghost_completion",
    "payload": {
      "candidate_ref": "cand-vapor-stub"
    }
  },
  "risk_level": "low",
  "requires_confirmation": false
}
```

## 评测口径

当前阶段至少检查以下指标：

- 结构合法率：`ghost_completion` 必须通过 schema 校验
- 本地合法率：建议必须来自本地规则认可的候选空间
- `Tab` 命中率：第一条默认建议被接受后应尽量命中用户真实下一步
- 空建议正确率：信息不足时，应稳定返回空建议而不是硬猜
- 建议后校验通过率：用户接受 ghost 后，本地命令与校验的通过率应足够高

## 非目标

当前任务不负责：

- 直接修改 `FlowsheetDocument`
- 绕过本地 canonical ports、连接校验或命令系统
- 把复杂多步搭建一次性扩成自治代理
- 替代 `suggest_flowsheet_edits` 去生成正式 patch

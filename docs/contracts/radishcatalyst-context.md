# RadishCatalyst 上下文预留契约

更新时间：2026-05-12

### `RadishCatalyst` 上下文预留

`RadishCatalyst` 当前更适合以结构化游戏数据、玩家知识源和脱敏运行状态摘要为中心打包上下文。该段是未来契约预留，不代表现有 schema 已接受 `project=radishcatalyst`。

```json
{
  "current_surface": "in_game",
  "route": "world/vertical_slice",
  "resource": {
    "kind": "game_state",
    "id": "world_20260505_001",
    "title": "第一可玩切片"
  },
  "player_context": {
    "mode": "offline_single_player",
    "locale": "zh-CN"
  },
  "static_data_summary": {},
  "world_state_summary": {},
  "character_state_summary": {},
  "quest_state_summary": {},
  "inventory_summary": {},
  "progression_context": {},
  "spoiler_policy": {
    "allowed_public_levels": ["public"],
    "hide_spoiler_by_default": true
  }
}
```

当前建议优先支持：

- `current_surface`
- `route`
- `resource`
- `player_context`
- `static_data_summary`
- `wiki_sources`
- `official_tool_sources`
- `world_state_summary`
- `character_state_summary`
- `quest_state_summary`
- `inventory_summary`
- `progression_context`
- `spoiler_policy`
- `search_scope`

说明：

- `static_data_summary` 应来自 `client/data/items.json`、`recipes.json`、`buildings.json`、`equipment.json`、`enemies.json`、`regions.json`、`quests.json` 等结构化数据的裁剪摘要，而不是整包无差别透传。
- `wiki_sources` 和 `official_tool_sources` 应只提供玩家可见或当前任务允许的片段，并保留 `public_level` / `source_type` / `page_slug` / `fragment_id` / `retrieval_rank`。
- `world_state_summary`、`character_state_summary`、`quest_state_summary` 与 `inventory_summary` 只能是只读摘要，不应包含完整存档原文、可直接覆盖写回的状态块或平台本地路径。
- `spoiler_policy.allowed_public_levels` 默认只允许 `public`；若未来允许 `spoiler`，调用侧必须显式标记目标 surface、用户确认状态和展示约束。
- 游戏截图、HUD 截图或地图截图适合通过 `artifacts` 追加为辅助输入，但不应取代静态数据和状态摘要。

当前预留任务建议：

- `answer_game_knowledge_question`
  - 用于玩家侧物品、配方、设备、区域、敌人、任务和 Wiki 知识问答。
  - 必须遵守 `spoiler_policy`，默认不回答 `internal` 或隐藏 `spoiler` 内容。
- `explain_player_progress_state`
  - 用于解释玩家当前为什么卡住、下一步目标、缺失材料、任务前置或区域解锁条件。
  - 输出只能是建议或说明，不直接推进任务或改存档。
- `suggest_production_chain_plan`
  - 用于基于目标物品、库存、配方、设备和解锁状态给出生产链规划建议。
  - 若涉及未解锁或隐藏内容，必须按 `spoiler_policy` 降级或要求确认。
- `inspect_static_data_consistency`
  - 用于开发侧检查配方引用、任务解锁、物品用途、公开等级、Wiki 可见性和官方工具反查缺口。
  - 输出应是 `read_only_check` 或候选修正建议，不直接修改 `client/data`。
- `summarize_wiki_or_design_content`
  - 用于整理玩家 Wiki、官方工具说明或设计文档摘要、标签和缺口建议。

当前非目标：

- 不让模型直接写入 Godot 存档、静态数据、任务状态、角色状态或世界状态
- 不让模型替代战斗命中、敌人行为、掉落、任务完成、存档迁移、联机同步或服务端权威
- 不把 `internal` 或默认隐藏的 `spoiler` 内容训练成玩家侧默认可见回答
- 不把模型输出当成配方、解锁、任务或数值平衡真相源

# `UI Design Topic` 任务卡：ops surface 设计到实现计划

更新时间：2026-05-23

## 任务目标

本任务卡把 `docs/designs/radishmind-console-ops-surface-v0.pen`、[UI 设计规范](../radishmind-ui-design-spec.md) 和后续 `apps/radishmind-console/` React 实现切片对齐起来。

当前任务不是立即重构 console，而是固定从设计稿到实现的顺序、缺口、验收和停止线，避免绕过设计源文件直接扩正式 UI。

## 输入事实源

- [UI 设计规范](../radishmind-ui-design-spec.md)
- [UI 设计参考](../radishmind-ui-design-reference.md)
- `docs/designs/radishmind-console-ops-surface-v0.pen`
- [当前推进焦点](../radishmind-current-focus.md)
- [阶段路线图](../radishmind-roadmap.md)
- `apps/radishmind-console/`

## 当前设计稿覆盖

当前 `.pen` 已覆盖：

- `Desktop Ready 1920x1080`
  - 本地 ops surface ready 态
  - service status
  - provider/profile details
  - dev diagnostics
  - local readiness
  - session/tooling surface
  - stop-line details
- `Desktop Failure/Stale 1920x1080`
  - overview 或 local-smoke 拉取失败时的 stale snapshot
  - last good 标记
  - failure diagnostics
  - blocked boundary 仍保持
- `Narrow 9:16 1080x1920`
  - 单列窄屏信息优先级
  - 状态、关键指标、readiness、session/tooling、stop-line、failure/stale rule
- `Loading / Empty 1920x1080`
  - 首次加载
  - refresh loading
  - 无 provider/profile inventory
  - 无 session/tool metadata
- `Settings / Permissions 1920x1080`
  - provider config placeholder
  - permission / confirmation policy placeholder
  - local environment note
  - 明确当前不声明 production secret backend 或 confirmation flow 已完成
- `Blocked Action Detail 1920x1080`
  - `POST /v1/tools/actions` 当前 blocked response 的可读解释
  - blocked reason、evidence、future gate
  - 不提供 execute / confirm / replay 按钮
- `Token Mapping Notes 1920x1080`
  - 将主要颜色、状态、边框、背景、文本映射到 `--rm-*` 语义 token
  - 左侧导航使用浅中性 `--rm-bg-sidebar` / `--rm-bg-sidebar-active` / `--rm-bg-sidebar-note`，不使用大面积深绿
  - 设计稿色值只作为 token 初值参考，不作为实现硬编码许可

## 设计稿剩余缺口

当前 `.pen` 已补齐正式 React UI 重排前要求的主要状态画面。后续进入第二批实现前，还需要人工评审上述新增画面与导出的可读截图，并确认是否需要追加更细的窄屏 loading / empty 专页。

## 第一批实现切片

第一批 React 实现只做基础视觉治理，不重排全部页面：

1. 在 `apps/radishmind-console/src/styles.css` 建立 `--rm-*` token。
2. 将现有 console 主要背景、面板、文本、边框、ready / stale / failed / blocked 状态色收敛到 token。
3. 保持现有 overview / local-smoke API 消费逻辑不变。
4. 保持现有 behavior / visual smoke 记录可通过。

第一批不做：

- 不新增 API。
- 不新增确认流。
- 不新增工具执行器。
- 不新增 replay、writeback、apply、execute、confirm 类按钮。
- 不大面积改组件结构。

## 第二批实现切片

第二批在设计稿评审后，再进入 ops surface 结构重排：

- 全局状态条
- readiness summary
- provider/profile inventory
- dev diagnostics
- session/tooling surface
- stop-line details
- failure/stale view
- 窄屏单列布局

第二批仍不改变接口语义，只调整信息架构和视觉层级。

## 验收口径

设计稿验收：

- Pencil `snapshot_layout` 无 layout problems。
- 关键画面导出 PNG 后可读，无文本溢出、遮挡、错位。
- 画面能区分 ready、loading、empty、stale、failed、blocked。
- 未实现能力没有被画成可点击执行路径。

React 实现验收：

- `pwsh ./scripts/check-repo.ps1 -Fast`
- `scripts/check-radishmind-console-behavior.py`
- `scripts/check-radishmind-console-visual-smoke-record.py`
- 必要时补 console 截图或浏览器人工复核记录

## 停止线

- Pencil 设计稿评审前，不进入正式 ops surface 结构重排。
- token 收敛不等于正式 UI 定稿。
- Settings / Permissions 当前只能是占位和边界说明，不声明真实 provider secret backend、permission store 或 confirmation flow。
- Blocked Action Detail 只能解释 blocked response，不提供执行入口。
- 窄屏稿必须是 `1080x1920` 单列优先级，不得把桌面三栏等比缩放。

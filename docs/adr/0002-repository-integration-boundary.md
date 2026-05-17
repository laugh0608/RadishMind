# ADR 0002: Repository Integration Boundary

更新时间：2026-05-08

## 状态

Accepted

## 背景

当前需要在以下两种思路之间做仓库级决策：

- 在 `RadishMind` 中加入 `Radish`、`RadishFlow`、`RadishCatalyst` 的 `git submodule`
- 反过来在 `Radish`、`RadishFlow`、`RadishCatalyst` 中加入 `RadishMind` 的 `git submodule`

但 `RadishMind` 当前已经明确定位为 `Radish` 体系下的外部智能层独立仓库，不是上层业务真相源，也不是承载全部上层代码的聚合仓库。

现有正式架构与契约也已经收口为：

- 上层项目通过 adapter 提供 `CopilotRequest`
- `RadishMind` 通过 gateway / runtime / response builder 返回 `CopilotResponse`
- 高风险输出继续由人工确认或上层规则层复核

在这一阶段，如果直接通过 `git submodule` 形成源码嵌套，会过早把代码分发、联调方式和仓库边界固化为长期结构，而 `RadishFlow`、`Radish` 与 `RadishCatalyst` 当前的真实接入成熟度并不一致。

## 决策

当前正式方案是：**不采用任一方向的 `git submodule` 作为 `RadishMind` 与 `Radish` / `RadishFlow` / `RadishCatalyst` 的默认集成方式。**

仓库边界固定为：

- `RadishMind`、`RadishFlow`、`Radish`、`RadishCatalyst` 继续保持独立仓库
- 正式集成通过统一协议、gateway、adapter、评测门禁和版本化发布制品完成
- 文档中的外部项目引用继续使用项目名和在线仓库 URL，而不是把外部仓库嵌入本仓库
- 若需要多仓联调、演示或集成测试，优先单独建立 integration workspace / super-repo，而不是让任一业务仓库或 `RadishMind` 充当长期聚合仓库

## 否决的方案

### 方案 A：在 `RadishMind` 中加入 `Radish`、`RadishFlow`、`RadishCatalyst` 子模块

否决原因：

- 这会把外部智能层反向变成上层业务仓库的聚合入口，和当前“独立仓库、只在本仓库工作区内工作”的边界相冲突
- 会把还未准备真实接入的 `Radish`、尤其是仍停留在文档预留阶段的 `RadishCatalyst`，一起拉入日常仓库治理面
- 会扩大工作区、提交审查面和版本矩阵，削弱 `RadishMind` 作为协议与评测中枢的清晰职责

### 方案 B：在 `Radish`、`RadishFlow`、`RadishCatalyst` 中加入 `RadishMind` 子模块

当前也不采用，原因是：

- 三个上层项目当前接入时机不同；现在同时铺开会把尚未稳定的接入形态提前固化
- 每个上层仓库都维护一份 `RadishMind` 子模块，会增加版本同步、CI、发布和回滚复杂度
- 当前更需要先冻结协议、gateway 与 smoke 门禁，而不是先决定源码分发形态

## 例外与后续

若未来出现明确的真实消费方、部署边界和版本分发要求，可以单独评估某一个上层项目是否需要使用 `RadishMind` 的版本化制品或受控源码引用。

但在那之前：

- 不在四个正式仓库之间建立双向 `git submodule`
- 不让任一正式业务仓库承担长期 super-repo 角色
- 多仓本地联调优先通过独立 integration workspace 解决

如果未来必须先做单点接入试点，应优先选择已经具备 gateway / service smoke / UI consumption 门禁的 `RadishFlow`，而不是同时推动三个上层仓库。

## 影响

正面影响：

- 仓库边界、职责归属和评测/协议治理保持清晰
- 避免把接入时机不同的三个上层项目一起绑定到同一套源码嵌套结构
- 后续可以按项目成熟度分别决定接入方式，而不是一次性锁死

代价：

- 多仓联调需要额外的 workspace / super-repo 设计，而不是直接依赖现成子模块结构
- 版本对齐要继续通过协议、发布制品、提交引用或集成清单治理

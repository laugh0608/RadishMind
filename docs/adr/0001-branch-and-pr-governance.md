# ADR 0001: Branch And PR Governance

更新时间：2026-07-12

## 状态

Accepted

## 背景

当前仓库刚进入规划与治理初始化阶段，但分支治理、PR 流程和自动化检查还未建立。  
如果继续在 `master` 上直接累积提交，后续很难形成稳定的发布基线，也不利于多人或多代理协作。

同时，`RadishMind` 当前已收口以 `Python` 作为主实现栈，但仓库内当前阶段的检查重点仍不是语言细节堆叠，而是：

- 文本文档卫生
- 治理文件齐备性
- 规则、PR 流程和工作流口径一致性

## 决策

仓库采用以下分支与 PR 治理策略：

### 分支角色

- `master`: 稳定主线，只接受 Pull Request 合并
- `dev`: 日常集成分支，功能、文档、规范类分支默认合并到这里
- `feature/*`: 功能开发分支
- `docs/*`: 文档与规范分支
- `chore/*`: 基础设施、脚本、CI、仓库治理分支
- `experiment/*`: 原型实验、提示词和模型路线验证分支

### 合并策略

- 默认开发流程为 `feature/*` / `docs/*` / `chore/*` / `experiment/*` -> `dev`
- 阶段性稳定后，通过 PR 将 `dev` 合并到 `master`，形成受完整门禁保护的阶段晋级
- 每次 `dev -> master` PR 合并后，必须在下一批开发继续前把最新 `master` 回同步到 `dev`，形成 `feature/* -> dev -> master -> dev` 闭环
- 仅在必须修复主线问题时，才允许 `hotfix/*` 直接向 `master` 发 PR；hotfix 合并后同样必须回同步到 `dev`
- `master -> dev` 只承担稳定主线和提交拓扑回灌，不表示允许在 `master` 上形成与 `dev` 并行的日常开发线

### 回同步方式

- 常态 `dev -> master` 阶段 PR 优先使用 `merge commit`；若合并期间 `dev` 未继续前进，PR 合并后的 `master` 应可直接 fast-forward 回 `dev`
- 若阶段 PR 使用 `rebase merge`，或 PR 合并期间 `dev` 已出现后续提交，则使用普通 merge 将 `master` 合入 `dev`
- 禁止通过 rebase、force push 或其它重写方式让共享 `dev` 追随 `master`
- 纯历史回同步且文件树没有变化时，不重复已经在 `PR -> master` 上通过的完整门禁
- 如回同步需要解决冲突或产生实际文件变化，应先在 `dev` 完成冲突审查和对应风险级别验证，再推送远端

### `master` 规则

- 禁止直接 push
- 必须通过 PR 合并
- 必须通过仓库检查
- 当前允许 `merge commit` 与 `rebase merge`，禁用 `squash merge`
- 常态 `dev -> master` 阶段 PR 优先使用 `merge commit`，降低合并后回同步 `dev` 的拓扑复杂度
- 管理员仅可通过 PR 方式绕过规则
- 允许在单人开发阶段保留管理员 PR 直过能力

### `dev` 规则

- 允许作为当前阶段默认目标分支
- 当前阶段不启用分支保护
- `push -> dev` 与 `pull_request -> dev` 不自动触发 `PR Checks`
- 开发阶段按改动风险执行本地分层验证；完整自动门禁统一收口到 `pull_request -> master`
- 如需复验远端环境，可手动触发 `PR Checks`
- `master` PR 合并后必须接收 `master -> dev` 回同步；该回同步不依赖自动 CI 触发，发生实际内容变化时由操作者先补对应验证

## 需要在 GitHub 仓库设置中完成的动作

以下规则不能仅靠仓库文件完全强制，需要仓库管理员在 GitHub Settings 中启用：

1. 创建远端 `dev` 分支
2. 将默认分支切换为 `dev`，或至少把开发 PR 默认目标改为 `dev`
3. 对 `master` 启用 branch protection / ruleset
4. 要求 `master` 通过 `Repo Hygiene`、`Repository Baseline`、`RadishMind Web Build` 与 `Platform Go Tests` 状态检查
5. 对 `master` 开启 “Require a pull request before merging”
6. 配置管理员仅通过 PR 绕过，不开放直接 push
7. 仓库 Merge options 中启用 `Merge commits` 与 `Rebase merging`，关闭 `Squash merging`
8. `dev` 当前不配置 branch protection

## 仓库内已落地的支撑项

为配合该决策，仓库内已同步增加：

- `AGENTS.md`
- PR 模板
- GitHub Actions PR 检查工作流
  - `PR Checks` 仅在目标分支为 `master` 的 Pull Request 上自动运行，并保留手动触发
  - 当前包含 `Repo Hygiene`、`Repository Baseline`、`RadishMind Web Build`、`Platform Go Tests` 与 `Platform PostgreSQL Integration` 五个 job
  - `master` required checks 当前按 job 名配置为 `Repo Hygiene` / `Repository Baseline` / `RadishMind Web Build` / `Platform Go Tests`
  - PR 页面可能展示 workflow 前缀或 `(pull_request)` 后缀，但它们不属于 ruleset 中需要手动配置的 check context
  - 规范 tag push 与手动补跑改由独立的 `Release Checks` workflow 承担，并使用 `Release Repo Hygiene` / `Release Repository Baseline` / `Release RadishMind Web Build` / `Release Platform Go Tests` 独立 job 名，避免与 PR required check 名称漂移或混淆
- 文本编码与文件格式检查脚本
- 仓库治理基线检查脚本
- `master` ruleset 模板

## 影响

正面影响：

- `master` 可以保持稳定
- `dev` 可以作为当前阶段的真实集成面
- 文档、规范、脚本和后续代码都能纳入统一 PR 检查
- 单人开发阶段仍保留必要的管理员 PR 绕过能力

代价：

- 需要维护远端 `master` 保护设置
- 开发节奏从“直接提交”切换为“分支 + PR”
- CI 需要随着产品面推进继续保持分层：仓库治理、正式 web 产品面、Go 平台服务和后续发布验证不应混成一个难定位的状态

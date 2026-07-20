# GitHub Rulesets

本目录存放 RadishMind 的仓库规则模板。  
当前维护默认分支 `master` / `main` 的保护规则模板，`dev` 作为常态开发分支，不启用强制保护。

## 建议流程

1. 日常开发提交到 `dev` 或功能分支
2. 功能、文档、规范类变更默认先合并到 `dev`
3. 阶段性稳定后，再从 `dev` 发起到默认分支（当前为 `master`，如切换可适配 `main`）的 Pull Request
4. 默认分支 PR 必须通过仓库检查
5. PR 合并后，立即把最新默认分支回同步到 `dev`，再继续下一批开发
6. 管理员如需绕过规则，也只能通过 Pull Request，不开放直接 push

## 默认分支规则说明

- 禁止直接推送到默认分支（`master` / `main`）
- 禁止 force push
- 禁止删除分支
- 仅允许通过 Pull Request 合并
- 要求 `Repo Hygiene`、`Repository Baseline`、`RadishMind Web Build`、`RadishMind Console Build` 与 `Platform Go Tests` 检查通过
- `PR Checks` 当前包含仓库卫生、仓库治理基线、Web 覆盖率与构建、Console 构建、Go 平台分包覆盖率 / race / vet 和 PostgreSQL 集成验证，保留拆分式门禁
- 默认分支 ruleset required checks 应按 job 名配置为 `Repo Hygiene`、`Repository Baseline`、`RadishMind Web Build`、`RadishMind Console Build` 与 `Platform Go Tests`
- PR 页面里可能显示 workflow 前缀或 `(pull_request)` 后缀，但这些不属于 ruleset 中需要配置的 check context
- `PR Checks` 自动响应 `pull_request -> dev/master`，并保留手动触发；如默认分支切换为 `main`，需先同步调整 workflow 触发分支，再在远端套用该模板
- tag push 与手动补跑使用独立的 `Release Checks` workflow，并显式使用 `Release Repo Hygiene`、`Release Repository Baseline`、`Release RadishMind Web Build`、`Release RadishMind Console Build` 与 `Release Platform Go Tests` job 名，避免与 PR required checks 混淆
- 允许 `merge` 与 `rebase` 两种合并方式，禁用 `squash`
- 常态 `dev -> master` 阶段 PR 优先使用 `merge commit`；PR 合并后必须完成 `master -> dev` 回同步
- 管理员仅可通过 Pull Request 方式绕过规则，不开放直接 push

## dev 策略说明

- `dev` 是当前常态开发分支
- 当前阶段不启用 branch protection
- `push -> dev` 不自动触发 `PR Checks`；目标为 `dev` 的 Pull Request 自动运行同一套检查，为其他开发者提供合并前反馈
- `dev` 当前不启用 required checks 或 branch protection；完整强制门禁仍统一收口到 `pull_request -> master`
- `master -> dev` 是每次阶段 PR 或 hotfix PR 合并后的必需回同步步骤，不是独立开发方向
- 回同步可 fast-forward 时优先 fast-forward；如因 `rebase merge` 或 `dev` 后续提交无法 fast-forward，则普通 merge `master` 到 `dev`，不得重写 `dev` 历史
- 纯拓扑回同步不重复完整门禁；冲突解决或实际内容变化必须先完成对应验证
- 如后续进入多人并行开发，再评估是否对 `dev` 追加保护

## master 合并后回同步

在工作区干净且远端引用已更新的前提下，优先执行：

```bash
git fetch origin
git switch dev
git pull --ff-only origin dev
git merge --ff-only origin/master
git push origin dev
```

如果 `git merge --ff-only origin/master` 因 `rebase merge` 或 `dev` 已继续前进而失败，改用普通 merge：

```bash
git merge --no-ff origin/master
# 审查并解决冲突；如产生实际内容变化，先完成对应验证
git push origin dev
```

## 检查入口

- `scripts/check-repo.ps1` 与 `scripts/check-repo.sh` 当前是正式仓库级验证入口
- `scripts/check-text-files.ps1` 与 `scripts/check-text-files.sh` 负责文本文件卫生检查
- `apps/radishmind-web/` 的 CI 使用 `npm ci`、`npm run test:coverage` 与 `npm run build`，`apps/radishmind-console/` 使用 `npm ci` 与 `npm run build`
- `services/platform/` 的 CI 使用核心包分层覆盖率入口、`go test -race ./...` 与 `go vet ./...`
- 当前基线重点是文本文件、治理文件齐备性、GitHub 规则 / workflow 口径一致性、Web 覆盖率与构建、Console 构建和 Go 平台分层测试

## 应用方式

如果仓库还没有对应 ruleset，可以使用 GitHub CLI 或 REST API 导入：

```bash
gh api repos/<owner>/<repo>/rulesets --method POST --input .github/rulesets/master-protection.json
```

如果仓库中已存在旧 ruleset，建议改用 `PUT /repos/{owner}/{repo}/rulesets/{ruleset_id}` 更新。

本目录模板还包含 Conventional Commits 的远端校验规则。当前远端 ruleset 未启用该规则；仅调整 Actions 触发策略、required checks 或合并方式时，应基于远端现状构造精确更新，不直接导入完整模板扩大门禁范围。

`master-protection.json` 中的 `actor_id: 5` 按“RepositoryRole = Admin”模板生成，表示管理员只能通过 PR 绕过规则。

## 配套仓库设置

- 仓库 Merge options 中启用 `Rebase merging`
- 仓库 Merge options 中启用 `Merge commits`
- 关闭 `Squash merging`
- 如后续增加 `CODEOWNERS`，再决定是否开启 code owner review

# GitHub Rulesets

本目录存放 RadishMind 的仓库规则模板。  
当前只维护 `master` 分支保护规则，`dev` 作为常态开发分支，不启用强制保护。

## 建议流程

1. 日常开发提交到 `dev` 或功能分支
2. 功能、文档、规范类变更默认先合并到 `dev`
3. 阶段性稳定后，再从 `dev` 发起到 `master` 的 Pull Request
4. `master` PR 必须通过仓库检查
5. 管理员如需绕过规则，也只能通过 Pull Request，不开放直接 push

## master 规则说明

- 禁止直接推送到 `master`
- 禁止 force push
- 禁止删除分支
- 仅允许通过 Pull Request 合并
- 要求 `Repo Hygiene` 与 `Planning Baseline` 检查通过
- `PR Checks` 当前拆分为 `Repo Hygiene` 与 `Planning Baseline` 两个 job，保留拆分式门禁
- `master` required checks 应按 job 名配置为 `Repo Hygiene` 与 `Planning Baseline`
- PR 页面里可能显示 workflow 前缀或 `(pull_request)` 后缀，但这些不属于 ruleset 中需要配置的 check context
- `PR Checks` 只响应 `pull_request -> master`
- tag push 与手动补跑使用独立的 `Release Checks` workflow，并显式使用 `Release Repo Hygiene` 与 `Release Planning Baseline` job 名，避免与 PR required checks 混淆
- 限制合并方式为 `squash` / `rebase`
- 管理员仅可通过 Pull Request 方式绕过规则，不开放直接 push

## dev 策略说明

- `dev` 是当前常态开发分支
- 当前阶段不启用 branch protection
- 当前默认不要求 push 到 `dev` 时自动触发仓库检查
- 当前也不强制收口到 `pull_request -> dev`；tag 与手动补跑改由独立的 `Release Checks` workflow 承担
- 如后续进入多人并行开发，再评估是否对 `dev` 追加保护

## 检查入口

- `scripts/check-repo.ps1` 与 `scripts/check-repo.sh` 当前是正式仓库级验证入口
- `scripts/check-text-files.ps1` 与 `scripts/check-text-files.sh` 负责文本文件卫生检查
- 当前基线重点是文本文件、治理文件齐备性和 GitHub 规则/工作流口径一致性

## 应用方式

如果仓库还没有对应 ruleset，可以使用 GitHub CLI 或 REST API 导入：

```bash
gh api repos/<owner>/<repo>/rulesets --method POST --input .github/rulesets/master-protection.json
```

如果仓库中已存在旧 ruleset，建议改用 `PUT /repos/{owner}/{repo}/rulesets/{ruleset_id}` 更新。

`master-protection.json` 中的 `actor_id: 5` 按“RepositoryRole = Admin”模板生成，表示管理员只能通过 PR 绕过规则。

## 配套仓库设置

- 仓库 Merge options 中启用 `Squash merging` 与 `Rebase merging`
- 关闭 `Merge commits`
- 如后续增加 `CODEOWNERS`，再决定是否开启 code owner review

## 变更说明

请简要说明本次 PR 的目标、范围和原因。

## 检查清单

- [ ] 本次改动符合当前阶段“地基建设优先”的方向，或已明确说明为何需要例外
- [ ] 已执行对应的最小验证
- [ ] 如修改了架构、边界、协议、流程或规范，已同步更新 `docs/` / `AGENTS.md`
- [ ] 未直接向 `master` 提交功能改动
- [ ] 默认目标分支为 `dev`；只有阶段性集成或发版时才面向 `master`

## 验证记录

请列出实际执行过的命令，例如：

```text
./scripts/bootstrap-dev.sh
./scripts/check-repo.sh --fast
npm --prefix apps/radishmind-web run build
(cd services/platform && go test ./...)
pwsh ./scripts/bootstrap-dev.ps1
pwsh ./scripts/check-repo.ps1 -Fast
```

## 风险与后续

请说明当前已知风险、未完成项和后续建议。

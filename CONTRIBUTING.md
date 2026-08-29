# 参与 RadishMind

感谢你参与 RadishMind。本项目当前是 `Radish` 体系下的内部开发者预览智能层，贡献应优先保持协议、权限、执行边界、评测证据与文档事实一致，不把开发 / 测试态能力提前描述为生产就绪。

## 开始之前

建议按以下顺序阅读：

1. [当前推进焦点](docs/radishmind-current-focus.md)
2. [功能设计文档入口](docs/features/README.md)
3. [产品范围与目标](docs/radishmind-product-scope.md)
4. [系统架构](docs/radishmind-architecture.md)
5. [跨项目集成契约](docs/radishmind-integration-contracts.md)
6. [代码规范](docs/radishmind-code-standards.md)
7. 与改动直接相关的功能专题、任务卡、契约和 ADR

安全漏洞不要创建公开 Issue 或 Pull Request，请遵循[安全策略](SECURITY.md)私下报告。参与讨论和审查时同时遵循[社区行为准则](CODE_OF_CONDUCT.md)。

本仓库采用 [RadishMind Source-Available License 1.0](LICENSE)，不是开放源码许可证。外部提交者不得推定仓库内容已经获得复制、修改、再分发、衍生开发或商业使用授权；提交贡献即表示接受 `LICENSE` 第 4 节的贡献授权条款。

## 贡献方式

- 缺陷修复：说明复现条件、实际结果、期望结果和影响范围，并优先修正根因。
- 功能开发：先更新对应 `docs/features/` 功能设计文档，再拆分小而完整、可复验的纵向切片。
- 协议、架构或高风险能力：先通过 Issue、设计讨论、ADR 或任务卡明确兼容性、迁移、失败语义和停止线。
- 文档与 UI：保持产品事实、页面状态和当前实现一致，不用文案或离线 evidence 暗示尚未实现的生产能力。
- 测试与评测：覆盖正例、负例、边界条件、失败关闭和必要的跨存储一致性，不把单次模型体验当作评测结论。

`docs/task-cards/` 只承接具体实现批次、前置条件或高风险门禁，不作为长期功能的默认主文档。普通 UI、文案、布局和只读 evidence 整理应复用现有测试与聚合门禁，不为每次修改派生新的 checker 链。

## 项目边界

- RadishMind 只输出解释、诊断、结构化建议和候选动作，不替代上层项目的业务真相源。
- 高风险动作必须保留 `requires_confirmation`、人工确认或规则层复核，不得由模型输出直接触发业务写回。
- `RadishFlow` 是当前优先支持对象；`Radish` 逐步扩展；`RadishCatalyst` 在没有明确任务前只保留文档级预留。
- 图片生成由独立 `RadishMind-Image Adapter` 与生图 backend 承接，不把像素生成并入 `RadishMind-Core` 的默认职责。
- 开发 / 测试态与 production 分别验收。真实 Provider、凭据、存储、身份和发布能力必须有明确准入，不得通过隐式 fallback 或扩大声明绕过停止线。

## 分支、提交与 Pull Request

- `master` 是受保护的稳定主线，`dev` 是日常集成分支。
- 项目所有者或已授权维护者串行推进普通任务时直接在 `dev` 开发和提交；外部贡献、并行写入、确有隔离价值的高风险改动或明确需要评审时，从主题分支向 `dev` 发起 PR。
- 需要主题分支时，使用 `feature/*`、`fix/*`、`docs/*`、`chore/*` 或 `experiment/*`；Agent 不因默认流程自动创建 `codex/*` 分支或额外 worktree。
- 只有阶段性 `dev` 晋级或明确的 `hotfix/*` 才向 `master` 发起 PR；不直接 push 到 `master`，不 force push 共享分支。
- `dev -> master` 合并后必须在下一批开发前将 `master` 回同步到 `dev`。完整规则见 [ADR 0001](docs/adr/0001-branch-and-pr-governance.md)。
- 提交遵循 Conventional Commits，例如 `feat(gateway): 增加受控 Provider 路由`、`fix(auth): 拒绝跨工作区 Session`、`docs(governance): 完善贡献规范`。
- 提交使用贡献者自己的 Git 身份，不添加 AI 协作者署名。

PR 描述应说明目标、范围、明确非目标、架构或协议影响、数据与安全影响、实际验证、未验证风险和必要的回滚方式。目标为 `master` 时，还应写明 `master -> dev` 回同步负责人和预期方式。

## 实现与数据要求

- 沿用现有模块职责、结构化 JSON 协议和明确类型；新增抽象必须消除真实重复或稳定表达职责边界。
- 错误应在正确 owner 内失败关闭，不通过吞异常、返回默认值、多层 fallback 或切换存储掩盖问题。
- 不提交真实 API Key、Provider 凭据、访问令牌、个人数据、专有输入、未经脱敏的日志或无权提供的第三方代码、模型、数据集和资产。
- 长语义应放入 manifest、record 或 fixture 元数据；committed 路径使用稳定短键并遵循仓库路径和文件体量预算。
- 代码、契约、文档、fixture 与检查器必须保持一致；阶段事实、停止线或协作方式变化时同步更新对应 `docs/`、周志或协作文件。

## 本地验证

首次拉取或缺少 `.venv` 时先准备开发环境：

```bash
./scripts/bootstrap-dev.sh
```

Windows PowerShell：

```powershell
pwsh ./scripts/bootstrap-dev.ps1
```

日常改动优先运行快速仓库检查：

```bash
./scripts/check-repo.sh --fast
```

```powershell
pwsh ./scripts/check-repo.ps1 -Fast
```

改动较大、准备发布，或影响评测 / 治理口径、协议、架构、阶段边界与文档真相源时，应补跑不带快速参数的完整检查，并执行与改动范围匹配的格式化、单元测试、集成测试、Web build 或真实浏览器验证。PR 只记录实际执行过的命令，并明确列出未执行或受环境阻塞的验证。

长时间运行、加载本地模型、下载数据或显著占用 CPU / GPU / 磁盘的脚本默认由贡献者在本机显式执行，不应加入普通仓库检查或在未说明资源影响时启动。

## 许可证

除非另有明确书面约定，贡献依照 [LICENSE](LICENSE) 第 4 节授权项目所有者使用。提交贡献即表示你有权提供相关内容；第三方代码、数据、模型、字体和资产必须标明来源并遵守各自许可证。

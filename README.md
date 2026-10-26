# RadishMind

`RadishMind` 是 `Radish` 体系下的 AI 工具、工作流、模型网关和 Copilot 集成平台，不是上层业务真相源。

## 入口

- [文档入口](docs/README.md)
- [当前推进焦点](docs/radishmind-current-focus.md)
- [产品范围与目标](docs/radishmind-product-scope.md)
- [系统架构](docs/radishmind-architecture.md)
- [阶段路线图](docs/radishmind-roadmap.md)
- [跨项目集成契约](docs/radishmind-integration-contracts.md)
- [Family UI 产品化设计](docs/features/user-workspace/radishmind-family-ui-productization-v1.md)
- [RadishMind UI 差异附录](docs/ui-addendum.md)
- [应用目录与 API 密钥开发测试指南](docs/features/user-workspace/application-catalog-api-key-dev-test-guide.md)
- [代码规范](docs/radishmind-code-standards.md)
- [参与贡献](CONTRIBUTING.md)
- [社区行为准则](CODE_OF_CONDUCT.md)
- [安全策略](SECURITY.md)
- [本地 Console 说明](apps/radishmind-console/README.md)

## 本地启动

- macOS / Linux / WSL：`./start.sh`
- Windows / PowerShell：`pwsh ./start.ps1`
- 产品 UI 开发联调：`./start.sh web-live`（默认使用开发假数据读取后端）
- 产品 UI 离线预览：`./start.sh web-offline`

这些是现有开发入口，持久化产品链需要按[项目指南](docs/radishmind-project-guide.md)选择对应配置。当前[本机持久工作台交付](docs/features/user-workspace/local-persistent-workbench-delivery-v1.md)已实现统一启动 / 配置与恢复命令，使用见[工作台指南](docs/platform/local-persistent-workbench-operations-v1.md)；实际浏览器 / 容器交付验收尚未完成。

## 仓库约定

- 当前常态开发分支为 `dev`
- `master` 仅作为稳定主线
- 首次拉取后先创建仓库虚拟环境：macOS / Linux / WSL 用 `./scripts/bootstrap-dev.sh`，Windows / PowerShell 用 `pwsh ./scripts/bootstrap-dev.ps1`
- 仓库级检查默认使用 `.venv`：macOS / Linux / WSL 用 `./scripts/check-repo.sh`，Windows / PowerShell 用 `pwsh ./scripts/check-repo.ps1`
- 文本文件默认走 UTF-8 + LF，规则以 `.editorconfig` 和 `.gitattributes` 为准
- 本地模型配置以仓库根 `.env.example` 为示例，真实 `.env` 只留本地
- 产品 UI 通用参考基线采用 RadishX `docs/design/family-ui/` `v26.7.3`；Profile、配色与精确 token 镜像是 RadishMind 的项目级选择，详见 `docs/ui-addendum.md`

## 说明

- 阶段事实、近期重点和停止线以 `docs/` 为准

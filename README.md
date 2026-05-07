# RadishMind

`RadishMind` 是 `Radish` 体系下的外部智能层仓库，负责协议、评测、工具编排与模型实验，不是上层业务真相源。

## 入口

- [文档入口](docs/README.md)
- [当前推进焦点](docs/radishmind-current-focus.md)
- [产品范围与目标](docs/radishmind-product-scope.md)
- [系统架构](docs/radishmind-architecture.md)
- [阶段路线图](docs/radishmind-roadmap.md)
- [跨项目集成契约](docs/radishmind-integration-contracts.md)
- [代码规范](docs/radishmind-code-standards.md)

## 仓库约定

- 当前常态开发分支为 `dev`
- `master` 仅作为稳定主线
- 仓库级检查入口：Linux / WSL 用 `./scripts/check-repo.sh`，Windows / PowerShell 用 `pwsh ./scripts/check-repo.ps1`
- 文本文件默认走 UTF-8 + LF，规则以 `.editorconfig` 和 `.gitattributes` 为准
- 本地模型配置以仓库根 `.env.example` 为示例，真实 `.env` 只留本地

## 说明

- 阶段事实、近期重点和停止线以 `docs/` 为准

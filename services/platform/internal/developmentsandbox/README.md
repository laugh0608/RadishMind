# 固定 Go 探针代码沙箱

本包是可信 Go 宿主显式启用的受限执行器，接入 `WorkerLaunch.Sandbox`，实现 `Binding / HandleTool / Close`。它只接受三文件代表任务与独立确认的控制计划，不是宿主 shell 或通用执行 API。边界、私有 schema、归属与停止语义见[沙箱合同](../../../../docs/contracts/development-code-sandbox-v1.md)。

`Config.Directory` 必须是新的私有绝对路径，调用方明确来源、镜像、Docker 程序及 Linux watchdog 摘要；watchdog 复制到该目录后固定挂载，不复用可变外部程序。实际验证快照全只读，候选替换由 Go 完成。该目录保存候选、快照、日志与资源账册；清理容器不删除唯一证据，创建时固定 24 小时临时保留期限；关闭后通过 `NewDeliverySource` 提供可信开发包来源，`RemoveExpiredArtifacts` 显式清理到期自有材料。保存 / 导出由[开发产物模块](../developmentartifact/README.md)负责，没有后台清理或公开下载入口。

常规验证不安装依赖、启动 Docker 或运行 SDK：

```sh
# 从 services/platform 执行
GOPROXY=off go test -race ./internal/developmentsandbox ./internal/developmentsandboxrunner ./internal/developmentcontrol
```

实际 Docker 测试只在具体运行窗口授权后设置以下环境；普通环境明确 skip，不算实际隔离证据。2026-10-31 使用已有 `golang:1.26.7-alpine3.23` arm64 镜像身份，不构建或下载镜像，所有命名容器记录后清理，保留已有镜像与私有产物：

```sh
# 从 services/platform 执行；先创建仓库 tmp/development-sandbox-20261031/cases 目录。
GOPROXY=off GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o ../../tmp/development-sandbox-20261031/code-runner ./cmd/radishmind-development-sandbox
RADISHMIND_SANDBOX_DOCKER=/usr/local/bin/docker \
RADISHMIND_SANDBOX_IMAGE=sha256:b17af760035fc2f338eed92d448a6c67f2d45438844fc6c60678fa5f99e44b57 \
RADISHMIND_SANDBOX_RUNNER="$PWD/../../tmp/development-sandbox-20261031/code-runner" \
RADISHMIND_SANDBOX_CASES="$PWD/../../tmp/development-sandbox-20261031/cases" \
GOPROXY=off go test -race -cover -count=1 -run '^TestDockerSandbox' -v -timeout 8m ./internal/developmentsandbox
```

没有真实模型或用户试用。其他架构、Windows 控制端、通用仓库、恶意代码防护、团队共享资源及正式交付仍未验证；不能从本机固定任务结果扩展声明。`CleanupRecovered` 只用于可信宿主显式收口旧账册，不自动重放命令。

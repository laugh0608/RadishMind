# 固定 Go 探针代码沙箱

本包是可信 Go 宿主显式启用的受限执行器，接入 `WorkerLaunch.Sandbox`，实现 `Binding / HandleTool / Close`。它只接受三文件代表任务与独立确认的控制计划，不是宿主 shell 或通用执行 API。边界、私有 schema、归属与停止语义见[沙箱合同](../../../../docs/contracts/development-code-sandbox-v1.md)。

`Config.Directory` 必须是新的私有绝对路径，调用方明确来源、镜像、Docker 程序及 Linux watchdog 摘要。watchdog 复制到该目录后固定挂载。实际验证快照全只读，候选替换由 Go 完成。该目录保存候选、快照、日志与资源账册。创建时固定 24 小时临时保留期限，关闭后由 `NewDeliverySource` 提供可信开发包来源，`RemoveExpiredArtifacts` 显式清理到期自有材料。已保存包由[开发产物模块](../developmentartifact/README.md)读取和导出。

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

工作台持锁启动时可调用 `TaskWorkers.RecoverCode` 按原实例、尝试、Docker 实例和容器身份核对遗留资源。此次恢复路径使用 Docker CLI 替身做了隔离测试，实际容器回收与冻结 SDK 联调另排。其他架构、Windows 控制端、通用仓库及团队共享资源仍待验收。[工作台核对与交接](../../../../docs/contracts/development-task-delivery-v1.md#sqlite-实现与宿主接线)

真实 SDK 同次任务的候选修改、测试、保存 / 导出、旧测试失效和命令中取消已固定响应联合验收，见[运行入口](../../../../scripts/eval/development_probe/README.md#sdk代码沙箱与产物联合验收2026-10-31)。联合用例默认 skip，仅在明确授权的可信控制容器中启用；不替代真实提供方、正式身份入口、Web 或模型效果验收。

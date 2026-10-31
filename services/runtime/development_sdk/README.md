# 受限 SDK 工作进程

本目录是可选的可信 SDK 执行层，不是工作台依赖，也不承载代码文件 / 命令沙箱。正常平台构造不导入或启动它。合同见[尝试控制](../../../docs/contracts/development-attempt-control-v1.md)，阶段与边界见[集中任务卡](../../../docs/task-cards/development-task-probe-v1-plan.md)。

## 入口与信任边界

可信 Go 宿主通过 `Runtime.RunWorker` 使用明确的绝对程序 / 参数、当前任务凭据和全新私有空目录启动一次尝试。当前仅实现 Linux / macOS 的独立进程组，目录要求当前用户拥有且权限为 `0700`；其他平台明确拒绝。调用方负责临时目录的分配和回收，不能使用用户仓库、工作台数据目录或秘密目录作为 SDK 工作目录。

Go 先提交不可重放的 `worker_claimed`，再启动进程；失败不自动重试。环境只保留明确的运行变量，凭据只经 stdin 交接。额外 FD 3 是父进程存活管道，父进程消失时 Python 终止自身所属进程组；不依赖 SDK pause 或 PID 猜测。stdout 只承载有界、连续序号的私有协议，SDK 常规日志重定向后由 Go 丢弃，不进入事件或公开错误。

`worker.py` 固定 `openhands-sdk==1.59.0`，使用真实 `Agent` / `LocalConversation`、内存会话存储及唯一自定义 `development_tool`。默认工具、MCP、插件、动态上下文、condenser、critic 和 fallback 关闭；只接受一条经确认的系统消息和一条用户任务，后续上下文仍由 Go 核对摘要。SDK 的同步 `generate` 显式转换为专属通道请求，其他同步 / 异步模型入口拒绝。`openai/gpt-4o` 只用于 SDK 本地类型 / 能力初始化，实际传输模型名称来自 Go 冻结选择，不调用该模型。

工具描述来自 Go 的 canonical 目录，不直接透传 SDK 自动增加的参数。SDK 报告与工具请求先由 Go 校验、持久提交，再通过 stdin 确认；确认前不能继续下一步。工具观察值只能来自可信 Go `ToolHandler` 经 `DispatchTool / CompleteTool` 的核对结果；默认没有 handler，不产生执行成功。该回调必须遵守取消上下文，固定 Go 任务现可由 `WorkerLaunch.Sandbox` 接到独立[代码沙箱](../../platform/internal/developmentsandbox/README.md)，实际文件由 Go 私有工具处理，固定验证在全只读快照容器运行；不能借此直接在宿主机执行用户代码。

SDK 事件只转换为明确的观察类型，错误仅带固定诊断码，不传任意异常文本、需求、代码、隐藏推理或原始 SDK transcript。Go 原始用量与预算为权威；SDK 本地 metrics 和缺失 usage 补出的零值不结算账本。

最终模型响应先关闭新调用入口并记录 `model_final_response`；SDK 完成报告、OS 正常退出、所属进程组为空且当前资格仍有效后，Go 才同事务记录退出与 `completed`。异常或清理未确认不冒充成功。重启只读已有状态，旧凭据失效，不恢复 SDK 会话或重放模型 / 工具。

## 验证方式

无 SDK 环境时，普通 Go 进程测试和 Python 私有协议测试可以运行，真实 SDK 测试明确 skip：

```sh
(cd services/platform && GOPROXY=off go test -race ./internal/developmentcontrol)
.venv/bin/python -m unittest discover -s scripts/checks/platform/tests -p 'test_development*.py'
```

真实 SDK 验证只在明确授权的隔离镜像内，将固定响应 Go 测试程序和本目录只读挂载；依赖锁与镜像说明见[SDK 检查入口](../../../scripts/eval/development_probe/README.md)。2026-10-31 已验证完整工具往返、模型失败不重试、工具失败、缺失用量、等待中取消和父进程存活管道断开。它不证明实际文件 / 命令隔离、真实提供方原生适配、生产安全或模型质量。

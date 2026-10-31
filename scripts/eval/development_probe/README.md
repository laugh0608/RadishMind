# SDK 候选依赖与固定响应检查

本目录只服务于[受控开发团队阶段 0](../../../docs/features/workflow/controlled-software-development-team/framework-and-gateway-integration-v1.md)，不属于工作台或既有 Python runtime 的依赖。

## 固定输入与运行入口

候选为 `sdk-candidate.txt` 中的 `openhands-sdk==1.59.0`，完整 138 个依赖的版本 / wheel SHA-256 已冻结在 `requirements-linux-arm64-py312.lock`。`deploy/Dockerfile.development-probe-sdk` 使用带摘要的 Python 3.12 基础镜像，直接以 `--require-hashes --no-deps --only-binary=:all:` 安装此锁并执行 `pip check`，不在构建时重新解析依赖。当前只验证 `linux/arm64` / Python 3.12.16；其他架构明确拒绝，不自动选择另一份依赖。

`lock_from_report.py` 保留为显式更新工具：从 pip 解析报告生成稳定排序的精确版本 / 哈希锁，拒绝重复包、非 PyPI wheel、无哈希分发与候选版本漂移，本身不联网。更新依赖需要另行授权并重新保留解析 / 安装报告、锁和兼容证据，不修改工作台或宿主机 `.venv`。

以下入口涉及构建 / 下载或容器启动，须有对应运行窗口。本轮已于 2026-10-31 获得授权并执行完毕，后续运行不自动沿用：

```sh
docker build -f deploy/Dockerfile.development-probe-sdk -t radishmind-development-probe-sdk:1.59.0 .
docker run --rm --network none --read-only --cap-drop ALL --security-opt no-new-privileges --tmpfs /tmp:size=64m,mode=1777 --memory 1g --cpus 2 --pids-limit 128 radishmind-development-probe-sdk:1.59.0 python /opt/probe/test_sdk_contract.py
docker run --rm --network none --read-only --cap-drop ALL --security-opt no-new-privileges --tmpfs /tmp:size=64m,mode=1777 --memory 1g --cpus 2 --pids-limit 128 radishmind-development-probe-sdk:1.59.0 python /opt/probe/inspect_sdk.py
```

镜像默认命令只导入 SDK 并打印安装版本。检查以 UID / GID `65532:65532` 运行，不发布端口、不挂载仓库或凭据，只有 `/tmp` 临时文件系统可写。契约测试的 HTTP 服务只监听同一容器的回环地址；`--network none` 阻断外部网络。脚本总时限 60 秒，LLM 请求上限 2 秒，超时用例为 1 秒。没有真实模型或代码任务执行。

## 实际结果与适用边界

2026-10-31，最终镜像的 7 项 SDK LLM 契约测试通过：

- 一个显式自定义工具、无 executor，完成工具请求 → 固定观察值 → 最终响应，两次请求的工具名、标识、结束原因与已给出的 usage 一致。
- 规范消费者拒绝畸形 / 重复字段、跨探针、路径越界、超预算序号、任意命令和未知工具。普通文本 HTTP 成功不能通过工具试验。SDK 自动声明的可选 `summary` 不属于 canonical 工具请求，也作为负例拒绝；正式适配需显式区分元数据与授权字段。
- 显式关闭重试与 fallback 后，400 / 429 / 500 及超时均只出现一次请求。
- 缺失上游 usage 时，当前 SDK 依赖链填入零 token，测试保留该观察；平台必须独立记录未知用量，不能采用 SDK 零值作为账本。

`fixtures/` 保存两个真实 SDK 捕获的合成请求体，无请求头、密钥或真实用户正文。`gpt-4o` 仅用于 SDK 本地能力元数据和固定响应协议标识，没有调用对应模型。Go 的 `TestSDKToolWireDoesNotGrantLegacyToolExecution` 只将模型选择改为本地 mock，再用内存 handler / fake bridge 核对两份请求：旧 Gateway 仍返回建议文本、`allow_tool_calls=false`，不传递工具标识。这是“不兼容”的证据，不是 Gateway 工具接入验收。

接口核对另确认 SDK 默认重试为 5，`LocalConversation.pause()` 不立即中断正在等待的模型调用。SDK 本地 Conversation 是工厂返回的具体实现，检查脚本同时记录工厂与 `LocalConversation` 的真实签名。当前未运行完整 Agent、事件流、命令取消、任务级委托或代码沙箱，不能宣布 SDK 已正式采用。

依赖元数据已导出，包括许可证字段；该记录不代替完整许可证核验。实际运行限制、摘要、首次失败及修正、安装一致性和清理记录见 [W44 周志](../../../docs/devlogs/2026-W44.md#2026-10-31-sdk-候选镜像与固定响应契约验收)，剩余工作见[集中任务卡](../../../docs/task-cards/development-task-probe-v1-plan.md)。

原始产物位于忽略目录 `tmp/development-probe-20261031/`：初始解析报告、冻结锁、最终安装报告、API / 依赖元数据、测试结果、容器限制与退出状态。临时容器已全部删除，镜像及构建缓存保留，未执行全局清理。完整 wheel 锁与安装版本集合一致只证明本次依赖复建，镜像时间 / 构建元数据不承诺逐位复现。

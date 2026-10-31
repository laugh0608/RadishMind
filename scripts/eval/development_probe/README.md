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

接口核对另确认 SDK 默认重试为 5，`LocalConversation.pause()` 不立即中断正在等待的模型调用。SDK 本地 Conversation 是工厂返回的具体实现，检查脚本同时记录工厂与 `LocalConversation` 的真实签名。该候选依赖窗口未运行完整 Agent、事件流、命令取消、任务级委托或代码沙箱。后续受限 Agent 验收见下节；仍不能宣布 SDK 已正式采用。

依赖元数据已导出，包括许可证字段；该记录不代替完整许可证核验。实际运行限制、摘要、首次失败及修正、安装一致性和清理记录见 [W44 周志](../../../docs/devlogs/2026-W44.md#2026-10-31-sdk-候选镜像与固定响应契约验收)，剩余工作见[集中任务卡](../../../docs/task-cards/development-task-probe-v1-plan.md)。

原始产物位于忽略目录 `tmp/development-probe-20261031/`：初始解析报告、冻结锁、最终安装报告、API / 依赖元数据、测试结果、容器限制与退出状态。临时容器已全部删除，镜像及构建缓存保留，未执行全局清理。完整 wheel 锁与安装版本集合一致只证明本次依赖复建，镜像时间 / 构建元数据不承诺逐位复现。

## 受限 Agent 工作进程验收（2026-10-31）

项目所有者另行授权复用摘要固定的已有镜像，只读挂载本次 Go 测试二进制与 `services/runtime/development_sdk/`，没有重新构建、下载依赖或安装宿主机依赖。Go 测试程序在同一无外网容器内创建回环固定响应通道、隔离 SQLite 和私有工作目录；SDK 以镜像原有非 root 用户运行，只有 `/tmp` 可写。完整限制、首次失败修正与清理记录见[运行记录](../../../docs/devlogs/2026-W44.parts/development-sdk-worker-20261031.md)。以下命令仅供下一次明确授权的相同范围复验，不是常规仓库门禁自动运行项：

```sh
# 从 services/platform 编译，使用已有模块缓存，不下载依赖。
GOPROXY=off GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go test -c ./internal/developmentcontrol -o ../../tmp/development-sdk-worker-20261031/sdk-check.test

# 以下从仓库根目录运行；需已有隔离目录和上述二进制。
docker run --name radishmind-sdk-worker-20261031-check \
  --network none --read-only --cap-drop ALL --security-opt no-new-privileges \
  --tmpfs /tmp:size=256m,mode=1777 --memory 1g --cpus 2 --pids-limit 128 \
  --mount "type=bind,src=$PWD/tmp/development-sdk-worker-20261031/sdk-check.test,dst=/opt/sdk-check,readonly" \
  --mount "type=bind,src=$PWD/services/runtime/development_sdk,dst=/opt/sdk-worker,readonly" \
  -e RADISHMIND_SDK_PYTHON=/usr/local/bin/python \
  -e RADISHMIND_SDK_SCRIPT=/opt/sdk-worker/worker.py \
  sha256:d80a7496e7c659f1fc76ffb5f429530bcfcd1fef5a790fa06b789ff27eb23842 \
  /opt/sdk-check -test.run '^TestSDKWorker(AgentFixedResponseRoundTrip|FailureDoesNotRetry|ModelWaitCancellation)$' -test.v -test.timeout 180s
# 无论检查成功或失败，记录状态后删除本次命名容器；保留已有镜像。
docker rm -f radishmind-sdk-worker-20261031-check
```

最终全部通过：真实 `Agent / LocalConversation` 经 Go 完成两次模型请求与合成工具观察值往返，核对系统提示、用户 / 助手消息、动作与观察事件；原生失败仅一次请求，工具失败保留未知状态，缺失 usage 不采用 SDK 零值；模型等待中取消后进程组为空，单独关闭父存活管道也能终止 SDK。合成观察值不访问文件、不执行代码。普通 Go 测试未配置 SDK 路径时会明确 skip，不能将其算作镜像验收。

Go 对每条 SDK 报告先提交事件再确认，最终完成须核对真实退出和最新资格。Python 仅使用内存会话、唯一自定义工具与受限模型转换，关闭默认工具 / MCP / 动态上下文等旁路；模块边界见[SDK 工作进程说明](../../../services/runtime/development_sdk/README.md)。实际代码沙箱、真实提供方原生适配、完整 SDK 正式采用及阶段 0 退出仍待完成。

## SDK、代码沙箱与产物联合验收（2026-10-31）

`run_joint.py` 只运行已准备的本地输入；不下载、构建、拉取镜像或启动 Docker 守护进程。需要另行明确授权可信控制容器访问 Docker socket；SDK 镜像与代码镜像分别固定为前述 `d80a7496…` 和 `b17af760…`。本次 Linux Docker CLI 固定 `29.8.2`，只解压可执行文件到隔离目录，不安装到系统或修改 SDK 依赖锁。

以下从仓库根目录执行，仅用于已授权窗口。`joint_root` 必须位于仓库 `tmp/`；首次准备使用新目录，后续复验递增 `--run-id`，保留失败记录。`--socket` 必须指向当前本机 Docker context 的 Unix socket，先用 `docker context inspect --format '{{.Endpoints.docker.Host}}'` 核对，不传 TCP 管理端口。

```sh
joint_root="$PWD/tmp/development-joint-20261031"
mkdir -p "$joint_root/tools" "$joint_root/input"
curl --fail --location --max-time 120 -o "$joint_root/tools/docker.tgz" https://download.docker.com/linux/static/stable/aarch64/docker-29.8.2.tgz
# 本次归档 SHA-256：76a624e4a8e5da654d1150e808175125efb5a6f1b6aa1cbd9caee18f51047a50
shasum -a 256 "$joint_root/tools/docker.tgz"
tar -xzf "$joint_root/tools/docker.tgz" -C "$joint_root/tools" docker/docker
cp -R scripts/eval/fixtures/development-task/list-filter scripts/eval/fixtures/development-task/list-filter-plan.json "$joint_root/input/"
(cd services/platform && GOPROXY=off CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go test -c -o "$joint_root/joint.test" ./internal/developmentsandbox)
(cd services/platform && GOPROXY=off CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o "$joint_root/code-runner" ./cmd/radishmind-development-sandbox)
# 仅接受本机 Unix socket；脚本会核对目标文件。
socket_path="$(docker context inspect --format '{{.Endpoints.docker.Host}}')"
socket_path="${socket_path#unix://}"
python3 scripts/eval/development_probe/run_joint.py --directory "$joint_root" --socket "$socket_path" --run-id 1
```

可信 Go 控制程序以 root 运行，只增补 `CHOWN / SETUID / SETGID / DAC_OVERRIDE / KILL`；`joint_worker.py` 在私有目录降为 `65532:65532`、清空附加组，核对 SDK 有效 / 许可 / 继承 / ambient capability 全为零及 `no-new-privileges`，实际打开控制数据库和连接 Docker socket 必须被拒绝。管理 socket 的父目录及材料挂载点的父目录均使用 Linux `0700` tmpfs；本机共享挂载的模式位曾未阻止 SDK 读取，不能单独作为隔离边界。该启动器只用于联合试验，未接入正式平台启动。

控制容器无外网、无宿主端口、只读根、1 GiB / 2 CPU / 128 进程、512 MiB `/tmp`；两处私有目录各 1 MiB。只挂载本次材料、明确的程序 / SDK 适配代码及已授权管理 socket，不挂载原工作树、日常数据库或秘密。代码容器继续使用独立无网络、非 root、只读源码的既定档案，无法接触控制数据或管理 socket。固定响应服务仅在控制容器回环临时端口监听。

三个真实用例为正常修复、先测试后修改、运行中取消；保存失败重试、摘要确认、幂等、关闭重开数据库、精确导出重读与独立导出权限撤销贯穿各例。资格读取和提供方是合成适配，SDK / SQLite / 文件工具 / 测试容器 / 产物负责模块均为真实实现。普通仓库测试不设置 `RADISHMIND_JOINT_ROOT` 时明确 skip。

每次保留 `run-*.command.json / .container.json / .state.json / .cleanup.json / .log`，case 内保存事件、SDK 边界检查、取消耗时、资源账册、验证报告、SQLite 和导出包。脚本结束移除精确控制容器，并按本次新增账册核对代码资源；如果需要宿主额外清理，则整轮仍判失败。最终运行与全部资源核对见[联合记录](../../../docs/devlogs/2026-W44.parts/development-joint-20261031.md)。本次运行窗口已结束，后续不自动沿用。

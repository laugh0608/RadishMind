# SDK 遥测可选依赖候选

状态：`1.59.0+radishmind.1` 的源码补丁、离线准备、版本边界和构建配方已准备；尚未构建、解析新运行依赖或执行真实 SDK。原始 `1.59.0` 锁和镜像保持历史试验基线，不能把原基线的验收结果归给本变体。

## 来源与改动

`source.json` 固定 PyPI `1.59.0` 源码归档 URL / SHA-256、精确上游文件摘要、补丁摘要和修改后文件摘要。许可证来自上游 `v1.59.0/LICENSE`，Git blob 也有记录。`optional-laminar.patch` 只有以下改动：

- 改为独立本地版本 `1.59.0+radishmind.1`；声明 MIT 并把原许可证与修改记录放入发行物。
- 将 `lmnr` 从必需依赖移至 `laminar` 可选安装项；不改版本范围，不修改或分发专有代理代码。
- 显式启用 Laminar / OTel 观测但缺少依赖时，初始化和直接创建根 span 均明确报错；默认关闭时沿用上游延迟导入和直接执行路径。

`prepare_sdk_variant.py` 只消费本地归档，不联网、不安装、不构建、不导入 SDK。它只创建仓库 `tmp/` 内的新目录，拒绝错误摘要、路径逃逸、重复路径、链接与特殊文件；先检查补丁再应用，核对结果摘要，移除副本中的旧发行元数据并记录完整构建输入摘要。失败保留不完整目录，不覆盖已有证据。

`build-requirements.lock` 仅锁定构建工具；原 138 项锁派生的 `constraints.txt` 仅约束既有版本，不强制安装未被使用的包。正常 pip 解析决定新的依赖闭包；只有明确构建的本地 SDK wheel 可使用本地 URL，其余发行物必须为有 SHA-256 的 PyPI wheel，出现 `lmnr` 或代理包即拒绝生成候选锁。新运行锁与实际安装报告尚待授权窗口产生，不预填依赖数量或轮子摘要。

`--no-deps` 仅用于单独构建 SDK wheel。依赖解析、下载和最终安装均保留依赖检查，最终执行 `pip check`；没有安装后删包。最终镜像的冻结安装阶段禁用网络，只读取导出的 wheel 集合与锁。

## 拟执行的依赖与运行窗口

以下操作均待单独运行授权。只在 `linux/arm64`、既有 Docker 环境中执行，不启动 Docker 守护进程，不安装宿主机或工作台依赖。隔离根目录为 `tmp/development-sdk-variant-20261031/`；构建阶段允许固定源码、PyPI wheel 和固定基础镜像访问，运行阶段无外网、无宿主端口。构建缓存、镜像、wheel、报告和失败材料保留，所有本轮运行容器结束删除；不推送或发布。

### 固定源码、解析与冻结安装

从仓库根目录执行。源地址和哈希由 `source.json` 给出；先下载到新文件，再由准备器核验。首次构建导出 wheel 与解析报告；第二次构建只消费该冻结集合。构建使用 Docker 现有资源配额，容器运行上限另见下节。

```sh
variant_root="$PWD/tmp/development-sdk-variant-20261031"
curl --fail --location --max-time 120 \
  -o "$variant_root/openhands_sdk-1.59.0.tar.gz" \
  https://files.pythonhosted.org/packages/77/2c/3c6a59e4088deeae25bf7255844e53374c337e69ec2446c0549df269af92/openhands_sdk-1.59.0.tar.gz
python3 scripts/eval/development_probe/prepare_sdk_variant.py \
  --source-archive "$variant_root/openhands_sdk-1.59.0.tar.gz" \
  --directory "$variant_root/context"
docker build --platform linux/arm64 --target artifacts \
  --output "type=local,dest=$variant_root/context/resolution" "$variant_root/context"
docker build --platform linux/arm64 --network none --target runtime \
  --iidfile "$variant_root/image-id.txt" \
  -t radishmind-development-sdk-variant:1.59.0-radishmind.1 "$variant_root/context"
```

构建成功后，检查 `resolution.json / requirements.lock / source.json`、wheel 许可证与镜像实际安装集合；实际本地 SDK wheel 摘要必须与解析、锁和安装报告相同。使用镜像真实 digest 做后续运行，不依赖可移动 tag。若缺失依赖、构建失败或摘要不一致，保留失败并停止，不放宽约束重试。

### 镜像内正反例与 LLM 子合同

以下每条命令保存独立日志。每轮检查实际容器配置、退出状态和 OOM；即使命令失败也删除对应精确容器，不能只依赖成功路径的 `--rm`。

```sh
set -e
variant_image="$(cat "$variant_root/image-id.txt")"
for check in optional_laminar contract inspect; do
  name="radishmind-sdk-variant-$check"
  script="/opt/probe/test_sdk_$check.py"
  if [ "$check" = inspect ]; then script=/opt/probe/inspect_sdk.py; fi
  docker create --name "$name" \
    --network none --read-only --cap-drop ALL --security-opt no-new-privileges \
    --tmpfs /tmp:size=64m,mode=1777 --memory 1g --memory-swap 1g --cpus 2 --pids-limit 128 \
    "$variant_image" python "$script" > "$variant_root/$check.container-id.txt"
  trap 'docker rm -f "$name"' EXIT
  docker inspect "$name" > "$variant_root/$check.container.json"
  status=0
  docker start --attach "$name" > "$variant_root/$check.log" 2>&1 || status=$?
  docker inspect --format '{{json .State}}' "$name" > "$variant_root/$check.state.json"
  docker rm "$name"
  trap - EXIT
  test "$status" -eq 0
  python3 -c 'import json,sys; s=json.load(open(sys.argv[1])); assert s["ExitCode"] == 0 and not s["OOMKilled"]' "$variant_root/$check.state.json"
done
```

可选依赖测试覆盖精确版本 / MIT 告知、默认环境无 `lmnr` / 代理分发包、可选依赖 marker、SDK 导入、关闭遥测的正常调用、四个显式遥测环境入口与直接创建根 span 的拒绝。LLM 测试继续覆盖真实 SDK 工具往返、非法输入、错误不重试、超时及缺失 usage；固定响应只在容器回环临时端口监听。

### 完整 SDK 与联合任务复验

本轮已用现有 Go 缓存准备 `sdk-check.test`、`joint/joint.test` 和 `joint/code-runner`，未执行它们；Linux Docker CLI 复用此前保留文件，不重新安装。SDK 层沿用[原工作进程验收命令](../README.md#受限-agent-工作进程验收2026-10-31)，替换本轮镜像 digest、二进制挂载路径和独立容器名称，维持非 root、无外网、只读根、1 GiB / 2 CPU / 128 进程、256 MiB 临时目录，覆盖完整事件、失败、等待取消和父管道退出。

联合任务使用已生成的 `joint/input/list-filter-plan.json`，其中候选版本必须为 `1.59.0+radishmind.1`，其余三文件摘要 / 工具 / 预算保持；不改写历史冻结计划。复用已验固定代码镜像与 Docker CLI：

```sh
socket_path="$(docker context inspect --format '{{.Endpoints.docker.Host}}')"
socket_path="${socket_path#unix://}"
python3 scripts/eval/development_probe/run_joint.py \
  --directory "$variant_root/joint" --socket "$socket_path" --run-id 1 \
  --sdk-image "$variant_image"
```

此步需要明确授权可信控制容器访问本机 Docker 管理 socket。权限、Linux 私有挂载父目录与 SDK 降权拒绝检查沿用[联合拓扑](../README.md#sdk代码沙箱与产物联合验收2026-10-31)：控制容器仅增补必要五项 capability，无外网、无端口、只读根，1 GiB / 2 CPU / 128 进程和 512 MiB 临时空间；代码容器继续独立无网络 / 非 root。覆盖候选修改 / 测试 / 保存 / 导出、旧测试失效和命令中取消。结束必须核对归属容器为空，管理 socket、原工作区和数据库不得进入 SDK 或代码沙箱。

## 采用限制

只有新闭包、冻结安装、无专有代理的发行物检查及上述实际 SDK / 联合回归均通过，才将本变体登记为技术候选。启用 `laminar` 可选安装项会重新引入原许可问题，不属于本镜像或当前授权范围。MPL 对应源码、其他依赖及基础镜像的分发义务仍需闭合；本轮不接真实提供方、不实现正式 Web、不宣布 SDK 采用或阶段 0 退出。

# SDK 候选依赖检查

本目录只服务于[受控开发团队阶段 0](../../../docs/features/workflow/controlled-software-development-team/framework-and-gateway-integration-v1.md)，不属于工作台或现有 Python runtime 的依赖。

候选输入为 `sdk-candidate.txt` 中的 `openhands-sdk==1.59.0`。它固定了 SDK 试验版本，**尚非采用结论或完整依赖锁**。`lock_from_report.py` 从 pip 的解析报告生成按名称排序的精确版本 / wheel SHA-256 锁，拒绝重复包、非 PyPI wheel、无哈希分发及候选版本漂移，本身不联网或安装。后续安装使用 `--require-hashes --no-deps`，并执行 `pip check`。

可审阅的镜像配方是 `deploy/Dockerfile.development-probe-sdk`，使用带摘要的 Python 3.12 基础镜像；候选依赖不写入宿主机 `.venv` 或工作台。解析结果与哈希锁在镜像 `/opt/probe/{resolution.json,requirements.lock,installation.json}`。当前配方每次重新解析可能得到不同的间接依赖；首次实测后应保存完整锁与镜像摘要，再以该锁复建，不能宣称仅固定 SDK 就已可复现全部依赖。

以下命令需要本轮明确的下载 / 容器授权，授权前不执行：

```sh
docker build -f deploy/Dockerfile.development-probe-sdk -t radishmind-development-probe-sdk:1.59.0 .
docker run --rm --network none --read-only --cap-drop ALL --security-opt no-new-privileges --tmpfs /tmp:size=64m,mode=1777 --memory 1g --cpus 2 --pids-limit 128 radishmind-development-probe-sdk:1.59.0
```

默认命令只导入 SDK 并打印安装版本，以非 root 运行、不发布端口、不挂载仓库或凭据；外部网络关闭，模型缓存使用离线设置。它不执行完整 Agent 或代码任务，导入成功也不等于工具协议兼容。后续固定响应测试需显式核对接口、关闭重试 / 隐藏调用，记录工具身份、拒绝路径、事件与预算。

本轮申请的产物目录为忽略目录 `tmp/development-probe-20261031/`；临时容器退出删除，镜像与构建缓存保留，不执行全局清理。当前状态为配方和锁生成测试已完成，SDK 未安装运行；授权及后续证据见[任务卡](../../../docs/task-cards/development-task-probe-v1-plan.md)。

# 单尝试控制合成样例

本目录只保存合成计划、委托、预留、回执、工具及取消轨迹，不含真实账户凭据、模型正文或用户代码。canonical 源为 `contracts/`；语义见[尝试控制合同](../../../../docs/contracts/development-attempt-control-v1.md)。`plan.json` 另由 Go 与 Python 严格消费；写路径属于读路径、UTF-8 字节限制与运行时关系由 Go 补充核对。

从仓库根目录使用现有环境验证：

```sh
.venv/bin/python -m unittest discover -s scripts/checks/platform/tests -p test_development_control_contracts.py
(cd services/platform && GOPROXY=off go test -race ./internal/developmentcontrol)
```

SQLite 事务故障注入、实际数据库关闭重开、并发提议消费、事件分页与原生固定响应测试位于 Go 控制模块；真实负责模块 / Session / 回环监听器生命周期测试位于 `httpapi/development_channel_test.go`。测试创建隔离临时目录，结束关闭数据库与监听器，不启动 Docker 或真实模型。`worker-events.json` 是私有 worker 协议合成轨迹，由同一 schema 聚合测试核对。Go 另有真实 SDK Agent 固定响应测试，须显式设置 SDK 路径并在授权镜像内运行；普通环境 skip 不算真实 SDK 验收。该镜像测试已验证工具往返、允许事件、模型等待取消与父管道失联退出，见[运行说明](../../development_probe/README.md)。实际文件 / 命令工具、代码容器终止和开发产物仍未验收；`202 queued` 不证明工具执行。

# 单次开发尝试的合同样例

这些 JSON 是纯合成输入，不包含 bearer、用户正文或真实执行结果。时间 `1000` 等数值只供可重复测试，不是当前有效委托。

`delegation / reservation / receipt / tool-action / tool-call.json` 分别对应 `contracts/` 中的专属开发合同。`cancel-trace.json` 是独立预期轨迹：预留 → 记录派发 → 请求取消 → 确认本地终止但保留未知用量 → 迟到可信回执只核账。Go 状态规则的输出与轨迹逐字段比较，Python 以 canonical schema 验证同一组样例；不是 SDK 或沙箱运行记录。

必要验证在仓库根目录使用：

```sh
(cd services/platform && GOPROXY=off GOTOOLCHAIN=local go test -race ./internal/developmentcontrol ./internal/strictjson ./internal/developmentprobe)
.venv/bin/python -m unittest discover -s scripts/checks/platform/tests -p 'test_development*_contracts.py'
```

字段、边界及尚未接线的执行职责见[尝试控制合同](../../../../docs/contracts/development-attempt-control-v1.md)。本目录没有执行、恢复或凭据签发入口。

# 受控开发代表任务：列表筛选与条数边界

本目录提供阶段 0 的合成 Go 任务，不对应 RadishMind 线上缺陷，不引用外部项目。

## 冻结输入

`list-filter-plan.json` 是符合 `contracts/development-task-probe-plan.schema.json` 的固定响应探针计划。`list-filter/` 恰好包含 `go.mod`、`list.go`、`list_test.go`；文件摘要写入计划，基准变化必须显式更新任务版本与证据，不静默接受新内容。模块仅使用标准库，不需要下载依赖。

任务要求：接口先按 `state=open|done` 筛选，再按合法 `limit` 返回对应数量，保持原顺序。空结果返回数组；非法状态、非法数字及非正数返回 400。`list.go` 特意保留 `result[:limit-1]` 的边界错误，基准测试中的 `filter_before_limit` 和 `limit_two` 应失败；修复后 8 项子用例应全部通过。不要将基准修正为成功版本。

只允许候选修改 `list.go`；`go.mod`、`list_test.go` 和验证参数受保护。首个探针使用预置回归测试，后续真实开发任务再验证 Agent 补充测试的能力。

## 离线准备与检查

在仓库根目录执行，输出目录必须尚不存在：

```sh
(cd services/platform && GOPROXY=off GOTOOLCHAIN=local go build -o ../../tmp/development-probe ./cmd/radishmind-development-probe)
tmp/development-probe check --plan scripts/eval/fixtures/development-task/list-filter-plan.json
tmp/development-probe prepare --plan scripts/eval/fixtures/development-task/list-filter-plan.json --source scripts/eval/fixtures/development-task/list-filter --output tmp/development-probe-case
tmp/development-probe inspect --plan scripts/eval/fixtures/development-task/list-filter-plan.json --candidate tmp/development-probe-case/candidate
```

`prepare` 只创建私有的基准副本、候选副本和 `workspace.json`，不启动进程。`inspect` 校验文件集合、保护文件和候选身份，不判断测试通过。目录内源文件大小上限为 32768 字节，拒绝额外文件、符号链接和非普通文件。失败时不覆盖既有输出；中途 IO 失败留下不完整目录供诊断，不据此续跑。

可信的人工检查可以在候选目录执行 `GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local CGO_ENABLED=0 go test -json -count=1 ./...`，先保留基准失败，再修改候选并保存实际测试输出。**该命令不是用于运行模型生成代码的宿主机执行入口。** 文件只读模式和事后摘要检查不能代替容器隔离；固定 Go 探针现已另由[代码沙箱](../../../../services/platform/internal/developmentsandbox/README.md)实现并验收只读测试挂载、资源限制、取消及同候选证据绑定；不改变本节人工宿主命令的用途。

2026-10-31 已执行人工修正的离线复验：基准两项预期失败、候选 8 项通过且基准未变。该人工结果不证明 SDK、Gateway、Agent 编码、代码沙箱或真实模型质量；后续真实容器验收按[沙箱合同](../../../../docs/contracts/development-code-sandbox-v1.md)另记，仍不证明真实编码质量。实施状态见[集中任务卡](../../../../docs/task-cards/development-task-probe-v1-plan.md)。

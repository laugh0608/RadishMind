# Gateway Bridge Runtime Baseline v1 任务卡

更新时间：2026-07-11

## 任务标识

- 任务 ID：`gateway-bridge-runtime-baseline-v1`
- 状态：`completed`
- 对应功能专题：`docs/features/gateway/python-bridge-runtime-v1.md`

## 目标

建立 process-per-request Python bridge 的可复现实测基线，按 Go route、process / IPC、Python Gateway 和 provider 分段记录成本，并据此确定持久 bridge 的唯一推荐形态。

## 实现范围

1. 在 `CopilotGatewayEnvelope.metadata` 增加 `provider_duration_ms`，保持 schema validation 与现有 envelope 语义。
2. 新增 `radishmind-bridge-benchmark`，支持 warmup、顺序请求、并发请求、并发度、request file 和 timeout 参数。
3. benchmark 只允许 mock provider 作为 committed 基线，输出脱敏 JSON，不包含请求正文、secret 或绝对路径。
4. 新增 Go route benchmark 和 benchmark 统计函数测试。
5. 运行基线，生成 `docs/features/gateway/evidence/` 下的短 run record，并完成候选选型。

## 必要验证

- `go test ./...`
- `go test -race ./...`
- `go vet ./...`
- Gateway envelope / service smoke
- bridge benchmark 顺序与并发实跑
- Go route benchmark 五轮
- `./scripts/check-repo.sh --fast`

本批涉及 Gateway contract 和执行性能口径；阶段结论写入当前焦点后补跑全量 `./scripts/check-repo.sh`。

## 完成条件

- 指标可复验，p50 / p95 统计方法固定。
- 四段成本均有数据或明确计算来源。
- 选型结论由数据支持，写清未选方案原因和后续实现停止线。
- 未新增真实 provider 调用、credential 材料、自动 fallback 或新的生产声明。

## 2026-07-11 完成记录

- `radishmind-bridge-benchmark` 使用 mock provider 完成 2 次预热、20 次顺序请求和 20 次四并发请求，共启动 42 个 Python 进程；顺序 total / process p95 为 `120.712 / 73.712 ms`，并发 total / process p95 为 `273.888 / 175.854 ms`。
- Go route benchmark 连续运行五轮，p50 / p95 为 `11772 / 13559 ns/op`；路由本身不是当前主要成本。
- process / IPC 在顺序和并发场景分别占 bridge 自身 p95 开销的 `75.4%` 与 `77.5%`，已选定受控 `stdio` worker pool 进入后续独立实现批次。
- 完整参数、四段统计、运行时版本、候选结论与限制见 [脱敏 run record](../features/gateway/evidence/process-per-request-baseline-2026-07-11.json)。
- 本任务只完成计时字段、基线工具、标准测试、实测与选型，没有实现或默认启用持久 worker。

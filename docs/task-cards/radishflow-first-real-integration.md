# `RadishFlow` 任务卡：接入前置条件与阻塞项

更新时间：2026-05-10

## 任务目标

把当前已经稳定的 `gateway -> UI consumption -> candidate handoff` 服务切片，收口为“在上层尚未具备真实接入能力前，本仓库应停止在哪一层”的前置条件与阻塞项清单。

当前目标不是新增模型能力、扩新样本或模拟更多 UI / 命令层，也不是提前设计不存在的真实接线，而是明确已有门禁、当前阻塞和后续何时才值得重开接入准备。

## 当前前提

- 进程内 Python gateway、service smoke matrix、gateway demo、UI consumption summary 与 candidate edit handoff summary 均已接入仓库级验证。
- `RadishFlow` 当前正式接入切片固定在 `radishflow/suggest_flowsheet_edits`。
- `M4` 的 broader review 与 `3B/4B` 审计记录当前只作为路线证据保留，不是这次阻塞盘点的主线。
- HTTP server、真实命令执行和上层写回流不属于当前最小范围。
- `RadishFlow`、`Radish` 与 `RadishCatalyst` 当前都还没有在本仓库内提供可直接落地的真实接线位置、确认流或命令层承接面。

## 当前冻结边界

当前本仓库在 `M3` 只冻结到以下边界：

1. `RadishFlow export -> adapter/request -> gateway envelope` 可复跑。
2. UI consumption 语义已通过 summary 固定为 advisory-only。
3. `candidate_edit` handoff 已固定为 non-executing proposal。
4. 仓库级 smoke matrix 可以证明 route、provider、`requires_confirmation`、UI 不写回和 handoff 不执行这些不变量。

在这四条之外，不继续在本仓库内前推接线设计。

## 当前已冻结的门禁资产

当前保留的正式门禁如下：

- `python3 scripts/check-radishflow-service-smoke-matrix.py --check-summary scripts/checks/fixtures/radishflow-service-smoke-matrix-summary.json`
- `python3 scripts/check-gateway-service-smoke.py --check-summary scripts/checks/fixtures/gateway-service-smoke-summary.json`
- `python3 scripts/run-radishflow-gateway-demo.py --manifest scripts/checks/fixtures/radishflow-gateway-demo-fixtures.json --check-summary scripts/checks/fixtures/radishflow-gateway-demo-summary.json --check`
- `python3 scripts/check-radishflow-gateway-ui-consumption.py --check-summary scripts/checks/fixtures/radishflow-gateway-ui-consumption-summary.json`
- `python3 scripts/check-radishflow-candidate-edit-handoff.py --check-summary scripts/checks/fixtures/radishflow-candidate-edit-handoff-summary.json`

文档或治理改动完成后，仓库级最小验证继续优先使用：

```bash
./scripts/check-repo.sh --fast
```

## 当前阻塞项

以下阻塞项没有明确前，不应继续推进真实接入设计：

- `RadishFlow` 中尚未明确真实承载 proposal view 的 UI 位置。
- 上层尚未明确用户确认 `candidate_edit` 的真实交互入口和确认流。
- 上层尚未明确接收 handoff proposal 的命令层接口或中间结构。
- 上层尚未明确审计日志的记录位置、保留周期和排障入口。
- `Radish` 与 `RadishCatalyst` 当前也都没有形成可复用的真实接线成熟度，不能反向为 `RadishFlow` 接入提供现成模式。

这些属于“上层项目接入前提”，不应继续在 `RadishMind` 仓库里用更多模拟 summary 或更细的接线文案替代。

## 何时重新打开接入准备

只有出现以下任一触发条件，才重新把这条任务从“阻塞盘点”推进到“接入准备”：

- 某个上层项目明确给出真实 UI 挂载点和确认流。
- 某个上层项目明确给出命令层 / handoff 承接接口。
- 现有 `M3` 门禁在真实接入试点中暴露出新的非重复 drift。

## 非目标

- 不新增新的 gateway route、HTTP server 或更多本仓库内模拟 UI / 命令层 summary。
- 不把当前任务扩成新的 `M4` 结构化输出实验。
- 不把 builder / guided 轨结果写成 raw 晋级、训练准入或 production contract 通过。
- 不让 `RadishMind` 直接执行 patch、写回 `FlowsheetDocument` 或承担上层命令层职责。
- 不在上层没有给出真实接线点前，继续细化 UI 字段映射、确认流文案或命令层接口。

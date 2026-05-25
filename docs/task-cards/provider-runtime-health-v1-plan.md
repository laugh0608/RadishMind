# `Provider Runtime & Health` v1 计划

更新时间：2026-05-25

## 任务目标

本任务卡用于把 `Runtime Service` 中已有的 provider/profile 分流能力，推进到可解释、可检查、可维护的 provider runtime / health 基础层。

当前任务不接真实 tool executor，不做 confirmation / writeback / replay，不启动模型训练放量，也不把 provider health 写成 production ready。它只处理 provider capability、profile selection、健康检查边界、错误分类和调用策略的可验证收口。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [阶段路线图](../radishmind-roadmap.md)
- [能力矩阵](../radishmind-capability-matrix.md)
- [系统架构](../radishmind-architecture.md)
- [战略定义](../radishmind-strategy.md)
- `services/runtime/provider_registry.py`
- `services/runtime/inference_provider.py`
- `services/platform/`
- `scripts/check-runtime-provider-dispatch.py`
- `scripts/check-platform-diagnostics.py`

## 当前事实

- `P1 Runtime Foundation` 已达到 short close，provider registry、northbound bridge、provider/profile inventory、diagnostics、request observability 和 error taxonomy 已有基础。
- `P3 Local Product Shell / Ops Surface` 已达到 `local usable / read-only close`。
- `Production Ops Hardening v1` 的静态部署边界已 close：docker local/test/prod compose、镜像命名治理、deployment readiness 静态 smoke、container smoke runbook 和 record template 已可检查；后续只保留一次明确运行窗口下的本地 container smoke 或测试环境 smoke。
- 当前仍缺 provider capability matrix、provider health policy、profile selection policy、fallback / retry / timeout 口径和外部 provider 健康检查边界。
- `provider-capability-matrix-v1` 已用 `scripts/checks/fixtures/provider-capability-matrix-v1.json` 与 `scripts/check-provider-capability-matrix.py` 固定为第一版可检查证据；它逐项比对 `services/runtime/provider_registry.py`，不联网、不要求 credential、不下载模型，也不声明 provider health 或 production ready。
- `provider-health-smoke-v1` 已用 `scripts/checks/fixtures/provider-health-smoke-v1.json` 与 `scripts/check-provider-health-smoke.py` 固定为第二版可检查证据；默认 fast baseline 只跑 mock runtime smoke 与 config-level inventory smoke，optional live health 仍是未来手动切片。

## v1 范围

1. `provider-capability-matrix`
   - 固定 provider / profile 的能力维度：transport、local_or_remote、chat、responses、messages、models_list、streaming、json_schema_output、tool_calling、image_input、image_output、auth_mode、timeout_policy、retry_policy、cost_profile、latency_profile、deployment_mode。
   - 不要求真实外部 provider key，不默认联网。
2. `provider-health-smoke`
   - 固定 mock / config-level / optional live health 三层检查。
   - 默认 fast baseline 只跑 mock 与离线配置检查。
   - optional live health 必须显式传 provider profile / base URL / credential source，失败不应被误写成 production outage。
3. `provider-selection-policy`
   - 固定 request-side profile/model selection、默认 profile、错误分类、fallback 禁止 / 允许条件。
   - 明确何时返回配置错误、credential 缺失、provider unavailable、timeout 或 unsupported capability。
4. `runtime-provider-docs-refresh`
   - 更新项目指南、架构、能力矩阵和脚本说明，让 provider runtime / health 不只存在于 fixture 或代码里。

## 非目标

- 不实现 production secret backend。
- 不实现 process supervisor。
- 不新增真实 tool executor、confirmation、writeback、replay 或 materialized result reader。
- 不默认联网探测外部 provider。
- 不下载模型、数据集或权重。
- 不把 provider health 写成 production readiness。
- 不把单一 provider 或单一模型绑定成平台唯一方向。

## 建议切片

1. `provider-capability-matrix-v1`
   - 已新增稳定 fixture / checker，固定 provider capability matrix 的字段、默认 provider/profile、可选择 model id 和不支持能力。
   - 复用现有 provider registry / diagnostics 输出，不另起第二套 provider truth。
2. `provider-health-smoke-v1`
   - 已新增离线 health policy、mock runtime smoke 与 config-level inventory smoke。
   - optional live health 只作为手动命令或显式参数，不进入默认 fast baseline。
3. `provider-selection-policy-v1`
   - 固定 profile/model selection 的负向场景：未知 profile、未知 model、credential missing、unsupported streaming / schema / tool capability、timeout。
   - 不启用自动 fallback，除非策略明确允许。
4. `provider-runtime-docs-refresh`
   - 同步说明类文档和任务卡入口。

## 验收口径

- provider capability matrix 有结构化证据和 checker。
- `provider-capability-matrix-v1.json` 与 `check-provider-capability-matrix.py` 已进入 `check-repo --fast`。
- provider health 分清 offline / optional live / production readiness。
- `provider-health-smoke-v1.json` 与 `check-provider-health-smoke.py` 已进入 `check-repo --fast`。
- request-side selection 与 diagnostics / `/v1/models` 的 profile id 口径一致。
- 默认快速检查不联网、不要求真实 credential、不下载模型。
- 保持 `Production Ops Hardening v1` 静态边界 close，不继续新增同类静态 governance 切片。
- `pwsh ./scripts/check-repo.ps1 -Fast` 通过。

## 下一步

下一步优先做 `provider-selection-policy-v1`，固定未知 profile / model、credential missing、unsupported capability、timeout 和 fallback 条件。若有明确 Docker 运行窗口，可以另行补一次本地 container smoke 运行记录；否则继续留在 provider runtime / health，不再补 production ops 同类静态边界。

## 停止线

- 不把 provider capability 写成真实 health。
- 不把 optional live health 写成默认门禁。
- 不把 health check failure 写成上层业务事件。
- 不把 provider fallback 做成隐式行为。
- 不接真实 executor、confirmation、writeback 或 replay。

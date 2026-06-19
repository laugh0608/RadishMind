# `Production Secret Backend Implementation` v1 计划

更新时间：2026-06-19

## 任务目标

本任务卡用于把 `production-secret-backend-contract` 之后的真实实现工作拆成可执行前置条件。当前只定义实现入口、输入输出、停止线、验收顺序和风险边界；不直接写 resolver 代码、不接真实云 secret 服务、不写入真实 secret、不声明 production ready。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [阶段路线图](../radishmind-roadmap.md)
- [Production Ops Hardening v1 计划](production-ops-hardening-v1-plan.md)
- `scripts/checks/fixtures/production-ops-secret-backend-contract.json`
- `scripts/check-production-ops-secret-backend-contract.py`
- `services/platform/internal/config/config.go`
- `services/platform/README.md`
- `deploy/.env.example`

## 当前事实

- `production-secret-backend-contract` 已固定 future external secret backend adapter contract、secret reference、脱敏输出和禁止项。
- 真实 production secret backend 仍为 `not_satisfied`。
- 当前平台配置仍按 `default < config file < env` 分层。
- `RADISHMIND_PLATFORM_API_KEY` 仍只允许作为 developer env override。
- `RADISHMIND_SECRET_SOURCE` 只能表示部署态外部 secret 来源要求，不是 secret backend 本身。
- 生产环境仍缺少 secret rotation policy、production secret audit store、provider health policy、process supervisor、正式 auth / CORS policy 和生产前复核记录。

## v1 实现目标

第一版实现目标不是接某个云厂商 SDK，而是先在平台层建立稳定的 secret reference 解析边界：

- 定义 provider profile 到 `secret_ref` 的绑定格式。
- 在 config / provider profile 选择链路中只传递 `secret_ref` 和脱敏状态。
- 预留 `SecretResolver` 接口或等价边界，但默认 backend 仍为 disabled。
- 失败时返回明确 failure boundary，例如 missing secret ref、backend disabled、resolution denied、resolver unavailable。
- diagnostics、config-summary、provider inventory 和 request metadata 只输出 `credential_state`、`secret_backend_configured`、`secret_ref_present`、`missing_secret_refs` 和 `field_sources`。

## 前置条件

1. `secret-ref-schema`
   - 明确 `environment`、`provider`、`provider_profile`、`secret_ref`、`required_fields` 和 `sanitized_fields`。
   - 禁止把 secret value、provider raw URL、token、cookie 或 authorization header 写入 schema fixture。
2. `config-injection-point`
   - 明确 `config.LoadFromEnv`、provider inventory 和 request-side selection 哪一层只接收 `secret_ref`，哪一层未来可调用 resolver。
   - 默认行为必须保持 env override dev-only，不自动启用 production secret backend。
   - 当前已落地：`production-secret-backend-config-secret-ref-readiness-v1` / `config_secret_ref_readiness_defined` 用 [Production Secret Backend Config / Secret Ref Readiness v1](../platform/production-secret-backend-config-secret-ref-readiness-v1.md)、`scripts/checks/fixtures/production-secret-backend-config-secret-ref-readiness-v1.json` 与 `scripts/check-production-ops-secret-backend-config-secret-ref-readiness-v1.py` 固定配置注入点、脱敏字段、failure mapping、no fallback 和 no side effects；这不实现 resolver runtime、不创建 fake resolver、不调用云 secret 服务、不读取 secret value、不启用 production secret backend。
3. `provider-profile-binding`
   - 明确 provider profile 如何声明需要 credential，以及缺失 secret ref 时的错误分类。
   - 不把 `mock`、`local-smoke` 或 demo profile 写成生产 provider。
4. `sanitized-audit-fields`
   - 固定 diagnostics、config-summary、provider inventory、request metadata 的脱敏字段。
   - 不输出 secret 原文或 provider base URL 原文。
5. `failure-taxonomy`
   - 固定 missing ref、backend disabled、resolver unavailable、resolution denied、credential missing 的错误码和 failure boundary。
   - 不吞掉 secret 解析错误，也不 fallback 到 mock provider。
6. `test-fixture-strategy`
   - 用 fake resolver / placeholder secret ref 做单元测试。
   - fast baseline 不联网、不要求真实 credential、不调用云 SDK。
7. `operator-runbook`
   - 明确测试环境和生产环境如何提供 secret source、如何验证脱敏状态、如何记录 smoke 结果。
   - 不把 `.env.example`、local env override 或 docker compose config 写成 secret backend。
8. `rotation-and-audit-policy`
   - 先定义 rotation / audit 的最小记录要求，再允许声明 production secret backend ready。
   - 当前不实现 rotation、不写 audit store。

## 建议切片

1. `secret-ref-schema-and-fixtures`
   - 新增 committed schema / fixture，只保存 secret reference 和脱敏状态。
   - 不保存 secret value。
   - 当前已落地：`contracts/production-secret-reference.schema.json`、`scripts/checks/fixtures/production-secret-reference-basic.json` 与 `scripts/check-production-secret-reference-contract.py` 固定 `environment`、`provider`、`provider_profile`、`secret_ref`、`required_fields`、`sanitized_fields` 和禁止字段；这不实现 resolver，也不代表 production secret backend ready。
2. `config-secret-ref-readiness`
   - 扩展 config summary / check 的脱敏字段，让缺失 secret ref 可被检查。
   - 不启用 resolver。
   - 当前已落地 readiness 定义：`docs/platform/production-secret-backend-config-secret-ref-readiness-v1.md`、`docs/task-cards/production-secret-backend-config-secret-ref-readiness-v1-plan.md`、`scripts/checks/fixtures/production-secret-backend-config-secret-ref-readiness-v1.json` 与 `scripts/check-production-ops-secret-backend-config-secret-ref-readiness-v1.py` 固定 `secret_backend_configured`、`secret_ref_present`、`missing_secret_refs` 和 `field_sources` 的未来脱敏语义；当前不添加 runtime 字段、不调用 resolver、不声明 credential resolved。
3. `provider-profile-secret-binding`
   - 让 provider/profile inventory 能声明 credential requirement 与 secret ref 状态。
   - 不访问 provider。
4. `secret-resolver-interface-disabled`
   - 加最小接口和 disabled backend，测试错误边界。
   - 不接云、不读本机真实 secret。
5. `operator-runbook-and-negative-gates`
   - 固定运行手册、负向门禁和 production ready 停止线。

## 验收口径

- 有任务卡、readiness fixture 和 checker 固定前置条件。
- `production_secret_backend` 仍为 `not_satisfied`，直到真实 resolver、运行手册、负向门禁、测试环境 smoke 和生产前复核记录都完成。
- fast baseline 不联网、不要求真实 credential、不写入真实 secret。
- `pwsh ./scripts/check-repo.ps1 -Fast` 通过。

## 停止线

- 不实现真实云 secret 服务。
- 不写入真实 secret。
- 不提交任何 API key、token、cookie、authorization header、证书或 provider raw dump。
- 不把 `.env.example`、developer env override、fake resolver 或 disabled resolver 写成 production secret backend。
- 不声明 production ready。
- 不接 executor、confirmation、writeback、replay 或 materialized result reader。

# `Production Secret Backend Implementation` v1 计划

更新时间：2026-06-20

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
   - 当前已落地：`production-secret-backend-provider-profile-secret-binding-readiness-v1` / `provider_profile_secret_binding_readiness_defined` 用 [Production Secret Backend Provider Profile Secret Binding Readiness v1](../platform/production-secret-backend-provider-profile-secret-binding-readiness-v1.md)、`scripts/checks/fixtures/production-secret-backend-provider-profile-secret-binding-readiness-v1.json` 与 `scripts/check-production-ops-secret-backend-provider-profile-secret-binding-readiness-v1.py` 固定 `credential_requirement`、`secret_ref_status`、环境绑定、failure mapping、sanitized diagnostics、no fallback 和 no side effects；这不实现 resolver runtime、不创建 fake resolver、不调用云 secret 服务、不读取 secret value、不启用 production secret backend。
4. `sanitized-audit-fields`
   - 固定 diagnostics、config-summary、provider inventory、request metadata 的脱敏字段。
   - 不输出 secret 原文或 provider base URL 原文。
   - 当前已落地：`production-secret-backend-secret-resolver-interface-disabled-readiness-v1` / `secret_resolver_interface_disabled_readiness_defined` 用 [Production Secret Backend Secret Resolver Interface Disabled Readiness v1](../platform/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.md)、`scripts/checks/fixtures/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.json` 与 `scripts/check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py` 固定 future resolver interface 的 disabled result、sanitized diagnostics、no fallback 和 no side effects；这不实现 resolver runtime、不创建 fake resolver、不读取 secret value、不启用 production secret backend。
5. `failure-taxonomy`
   - 固定 missing ref、backend disabled、resolver unavailable、resolution denied、credential missing 的错误码和 failure boundary。
   - 不吞掉 secret 解析错误，也不 fallback 到 mock provider。
   - 当前已落地：`production-secret-backend-secret-resolver-interface-disabled-readiness-v1` / `secret_resolver_interface_disabled_readiness_defined` 已把 `secret_resolver_secret_ref_missing`、`secret_resolver_backend_disabled`、`secret_resolver_unavailable`、`secret_resolution_denied`、`secret_resolver_environment_mismatch` 和 `secret_resolver_invocation_disabled` 固定到 `configuration` failure boundary。
6. `test-fixture-strategy`
   - 用 fake resolver / placeholder secret ref 做单元测试。
   - fast baseline 不联网、不要求真实 credential、不调用云 SDK。
   - 当前已落地评审证据：`production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1` / `test_fixture_strategy_fake_resolver_entry_review_defined` 用 [Production Secret Backend Test Fixture Strategy / Fake Resolver Implementation Entry Review v1](../platform/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.md)、`scripts/checks/fixtures/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.json` 与 `scripts/check-production-ops-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.py` 固定 `test-fixture-strategy` 与 fake resolver implementation entry review；结论仍是 `test-fixture-strategy` 保持 `required_before_implementation`，fake resolver implementation entry 不打开。这不实现 resolver runtime、不实现 fake resolver runtime、不解析 secret、不调用云 secret 服务、不接数据库、不启用 repository mode。
   - 当前已落地策略证据：`production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1` / `fake_resolver_contract_no_secret_leakage_smoke_strategy_defined` 用 [Production Secret Backend Fake Resolver Contract / No Secret Leakage Smoke Strategy v1](../platform/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.md)、`scripts/checks/fixtures/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.json` 与 `scripts/check-production-ops-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.py` 固定 fake resolver static contract、no secret leakage smoke strategy、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard；结论仍是 `test-fixture-strategy` 未 satisfied，fake resolver implementation entry 不打开。这不实现 resolver runtime、不实现 fake resolver runtime、不创建 no secret leakage smoke runtime、不解析 secret、不调用云 secret 服务、不接数据库、不启用 repository mode。
   - 当前已落地 task card entry readiness：`production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1` / `fake_resolver_implementation_task_card_entry_readiness_review_defined` 用 [Production Secret Backend Fake Resolver Implementation Task Card Entry Readiness Review v1](../platform/production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.md)、`scripts/checks/fixtures/production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.json` 与 `scripts/check-production-ops-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.py` 固定下一张 fake resolver implementation task card 的准入。
   - 当前已落地静态任务卡：`production-secret-backend-fake-resolver-implementation-v1` / `fake_resolver_implementation_task_card_defined` 用 [Production Secret Backend Fake Resolver Implementation v1](../platform/production-secret-backend-fake-resolver-implementation-v1.md)、`scripts/checks/fixtures/production-secret-backend-fake-resolver-implementation-v1.json` 与 `scripts/check-production-ops-secret-backend-fake-resolver-implementation-v1.py` 固定 fake resolver implementation 任务卡、disabled-by-default runtime gate、no leakage smoke plan、sanitized diagnostics boundary 和 artifact guard；结论仍是不实现 resolver runtime、不实现 fake resolver runtime、不创建 no secret leakage smoke runtime、不解析 secret、不调用云 secret 服务、不接数据库、不启用 repository mode。
   - 当前已落地真实 resolver runtime 前置：`production-secret-backend-real-resolver-runtime-preconditions-v1` / `real_resolver_runtime_preconditions_defined` 用 [Production Secret Backend Real Resolver Runtime Preconditions v1](../platform/production-secret-backend-real-resolver-runtime-preconditions-v1.md)、`scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-preconditions-v1.json` 与 `scripts/check-production-ops-secret-backend-real-resolver-runtime-preconditions-v1.py` 固定真实 resolver runtime 的启用条件、secret ref / provider profile / environment binding、operator approval、audit / rotation dependency、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard；这不实现 production resolver runtime、不读取真实 secret、不调用云 secret 服务、不连接数据库、不创建 credential payload、不启用 repository mode、不新增 public production API。
   - 当前已落地真实 resolver runtime implementation entry review：`production-secret-backend-real-resolver-runtime-implementation-entry-review-v1` / `real_resolver_runtime_implementation_entry_review_defined` 用 [Production Secret Backend Real Resolver Runtime Implementation Entry Review v1](../platform/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1.md)、`scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1.json` 与 `scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-review-v1.py` 固定 entry decision 为 blocked before runtime task card；backend profile selection、no leakage smoke runtime gate、credential handle runtime boundary、operator approval runtime evidence、audit store handoff 和 backend health boundary 仍是 blocker。这不创建 production resolver runtime implementation task card、不实现 production resolver runtime、不读取真实 secret、不调用云 secret 服务、不连接数据库、不创建 credential payload、不启用 repository mode、不新增 public production API。
   - 当前已落地 resolver backend profile selection readiness：`production-secret-backend-resolver-backend-profile-selection-readiness-v1` / `resolver_backend_profile_selection_readiness_defined` 用 [Production Secret Backend Resolver Backend Profile Selection Readiness v1](../platform/production-secret-backend-resolver-backend-profile-selection-readiness-v1.md)、`scripts/checks/fixtures/production-secret-backend-resolver-backend-profile-selection-readiness-v1.json` 与 `scripts/check-production-ops-secret-backend-resolver-backend-profile-selection-readiness-v1.py` 固定 backend profile shape、reserved backend kind allowlist、environment / provider profile / policy version binding、operator approval、audit / rotation dependency、backend health reference、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard；这不创建 backend runtime、不实现 production resolver runtime、不读取真实 secret、不调用云 secret 服务、不连接数据库、不创建 credential payload 或 credential handle、不启用 repository mode、不新增 public production API。
7. `operator-runbook`
   - 明确测试环境和生产环境如何提供 secret source、如何验证脱敏状态、如何记录 smoke 结果。
   - 不把 `.env.example`、local env override 或 docker compose config 写成 secret backend。
   - 当前已落地：`production-secret-backend-operator-runbook-negative-gates-readiness-v1` / `operator_runbook_negative_gates_readiness_defined` 用 [Production Secret Backend Operator Runbook / Negative Gates Readiness v1](../platform/production-secret-backend-operator-runbook-negative-gates-readiness-v1.md)、`scripts/checks/fixtures/production-secret-backend-operator-runbook-negative-gates-readiness-v1.json` 与 `scripts/check-production-ops-secret-backend-operator-runbook-negative-gates-readiness-v1.py` 固定 operator runbook、negative gates、sanitized diagnostics、no fallback 和 no side effects；这不实现 resolver runtime、不创建 fake resolver、不调用云 secret 服务、不读取 secret value、不启用 production secret backend。
8. `rotation-and-audit-policy`
   - 先定义 rotation / audit 的最小记录要求，再允许声明 production secret backend ready。
   - 当前已落地：`production-secret-backend-rotation-audit-policy-readiness-v1` / `rotation_audit_policy_readiness_defined` 用 [Production Secret Backend Rotation / Audit Policy Readiness v1](../platform/production-secret-backend-rotation-audit-policy-readiness-v1.md)、`scripts/checks/fixtures/production-secret-backend-rotation-audit-policy-readiness-v1.json` 与 `scripts/check-production-ops-secret-backend-rotation-audit-policy-readiness-v1.py` 固定 rotation trigger、audit event fields、secret ref version reference、rollback policy、failure mapping、no fallback 和 no side effects；这不实现 rotation runtime、不写 audit store、不实现 resolver runtime、不创建 fake resolver、不调用云 secret 服务、不读取 secret value、不启用 production secret backend。

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
   - 当前已落地 readiness 定义：`docs/platform/production-secret-backend-provider-profile-secret-binding-readiness-v1.md`、`docs/task-cards/production-secret-backend-provider-profile-secret-binding-readiness-v1-plan.md`、`scripts/checks/fixtures/production-secret-backend-provider-profile-secret-binding-readiness-v1.json` 与 `scripts/check-production-ops-secret-backend-provider-profile-secret-binding-readiness-v1.py` 固定 provider/profile 到 `secret_ref` 的 reference-only 绑定；当前不添加 runtime 字段、不调用 resolver、不声明 credential resolved。
4. `secret-resolver-interface-disabled`
   - 固定 reference-only interface、disabled result、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard。
   - 当前已落地 readiness 定义：`docs/platform/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.md`、`docs/task-cards/production-secret-backend-secret-resolver-interface-disabled-readiness-v1-plan.md`、`scripts/checks/fixtures/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.json` 与 `scripts/check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py` 固定 `secret-resolver-interface-disabled`；当前不添加 resolver runtime、不创建 fake resolver、不解析 secret、不接云、不读本机真实 secret。
5. `operator-runbook-and-negative-gates`
   - 固定运行手册、负向门禁和 production ready 停止线。
   - 当前已落地 readiness 定义：`docs/platform/production-secret-backend-operator-runbook-negative-gates-readiness-v1.md`、`docs/task-cards/production-secret-backend-operator-runbook-negative-gates-readiness-v1-plan.md`、`scripts/checks/fixtures/production-secret-backend-operator-runbook-negative-gates-readiness-v1.json` 与 `scripts/check-production-ops-secret-backend-operator-runbook-negative-gates-readiness-v1.py` 固定 `operator-runbook-and-negative-gates`；当前不添加 resolver runtime、不创建 fake resolver、不解析 secret、不接云、不读本机真实 secret。
6. `rotation-and-audit-policy`
   - 固定 rotation / audit policy、failure mapping 和 production ready 停止线。
   - 当前已落地 readiness 定义：`docs/platform/production-secret-backend-rotation-audit-policy-readiness-v1.md`、`docs/task-cards/production-secret-backend-rotation-audit-policy-readiness-v1-plan.md`、`scripts/checks/fixtures/production-secret-backend-rotation-audit-policy-readiness-v1.json` 与 `scripts/check-production-ops-secret-backend-rotation-audit-policy-readiness-v1.py` 固定 `rotation-and-audit-policy`；当前不添加 rotation runtime、不写 audit store、不添加 resolver runtime、不创建 fake resolver、不解析 secret、不接云、不读本机真实 secret。
7. `test-fixture-strategy / fake-resolver-entry-review`
   - 固定 test fixture strategy 与 fake resolver implementation entry 是否打开的评审结论。
   - 当前已落地 entry review：`docs/platform/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.md`、`docs/task-cards/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1-plan.md`、`scripts/checks/fixtures/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.json` 与 `scripts/check-production-ops-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.py` 固定 `test_fixture_strategy_fake_resolver_entry_review_defined`；当前 `test-fixture-strategy` 仍未 satisfied，不添加 fake resolver implementation task card、不创建 fake resolver runtime、不添加 no secret leakage smoke runtime、不接云、不读 secret、不接数据库。
8. `fake-resolver-contract / no-secret-leakage-smoke-strategy`
   - 固定 fake resolver implementation 前必须具备的静态输入 / 输出 allowlist、禁止 secret-bearing 字段、泄漏扫描策略、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard。
   - 当前已落地策略定义：`docs/platform/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.md`、`docs/task-cards/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1-plan.md`、`scripts/checks/fixtures/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.json` 与 `scripts/check-production-ops-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.py` 固定 `fake_resolver_contract_no_secret_leakage_smoke_strategy_defined`；当前 `test-fixture-strategy` 仍未 satisfied，不添加 fake resolver implementation task card、不把 test-only fake resolver runtime 写成 production resolver、不创建 no secret leakage smoke runtime、不接云、不读 secret、不接数据库。
9. `fake-resolver-implementation-task-card-entry-readiness-review`
   - 评审 fake resolver implementation task card 是否可以作为下一步单独创建。
   - 当前已落地 entry readiness：`docs/platform/production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.md`、`docs/task-cards/production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1-plan.md`、`scripts/checks/fixtures/production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.json` 与 `scripts/check-production-ops-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.py` 固定 `fake_resolver_implementation_task_card_entry_readiness_review_defined`；后续任务卡已在下一切片单独创建。
10. `fake-resolver-implementation`
   - 创建 fake resolver implementation 静态任务卡，固定后续 runtime implementation 的 disabled gate、placeholder secret ref fixture、no leakage smoke、sanitized diagnostics、no fallback、no side effects 和 artifact guard。
   - 当前已落地任务卡：`docs/platform/production-secret-backend-fake-resolver-implementation-v1.md`、`docs/task-cards/production-secret-backend-fake-resolver-implementation-v1-plan.md`、`scripts/checks/fixtures/production-secret-backend-fake-resolver-implementation-v1.json` 与 `scripts/check-production-ops-secret-backend-fake-resolver-implementation-v1.py` 固定 `fake_resolver_implementation_task_card_defined`；当前不实现 resolver runtime、不实现 fake resolver runtime、不创建 no secret leakage smoke runtime、不接云、不读 secret、不接数据库。
11. `fake-resolver-runtime-implementation-entry-review`
   - 评审 fake resolver implementation 静态任务卡是否足以进入下一张 test-only runtime implementation 任务卡。
   - 当前已落地 entry review：`docs/platform/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.md`、`docs/task-cards/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1-plan.md`、`scripts/checks/fixtures/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.json` 与 `scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.py` 固定 `fake_resolver_runtime_implementation_entry_review_defined`；当前只允许下一步单独创建 runtime implementation 任务卡，不实现 resolver runtime、不实现 fake resolver runtime、不创建 no secret leakage smoke runtime、不接云、不读 secret、不接数据库。
12. `fake-resolver-runtime-implementation`
   - 创建 test-only fake resolver runtime，固定后续 test-only runtime gate、placeholder secret ref fixture、environment binding、opaque credential handle metadata、sanitized diagnostics、offline no leakage smoke 和 side effect counters。
   - 当前已落地实现：`docs/platform/production-secret-backend-fake-resolver-runtime-implementation-v1.md`、`docs/task-cards/production-secret-backend-fake-resolver-runtime-implementation-v1-plan.md`、`scripts/checks/fixtures/production-secret-backend-fake-resolver-runtime-implementation-v1.json`、`scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-v1.py`、`services/platform/internal/secretbackend/fake_resolver.go` 与 `services/platform/internal/secretbackend/fake_resolver_test.go` 固定 `fake_resolver_runtime_test_only_implemented`；当前只实现 test-only、默认 disabled 的 fake resolver runtime，不实现 production resolver runtime、不创建 no secret leakage smoke runtime、不接云、不读 secret、不接数据库。
13. `real-resolver-runtime-preconditions`
   - 固定真实 resolver runtime implementation 之前的启用条件、secret ref / provider profile / environment binding、operator approval、audit / rotation dependency、failure mapping、sanitized diagnostics、no fallback、no side effects、artifact guard 和后续实现拆分。
   - 当前已落地前置定义：`docs/platform/production-secret-backend-real-resolver-runtime-preconditions-v1.md`、`docs/task-cards/production-secret-backend-real-resolver-runtime-preconditions-v1-plan.md`、`scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-preconditions-v1.json` 与 `scripts/check-production-ops-secret-backend-real-resolver-runtime-preconditions-v1.py` 固定 `real_resolver_runtime_preconditions_defined`；当前不实现 production resolver runtime、不读取真实 secret、不调用云 secret 服务、不连接数据库、不创建 credential payload、不启用 repository mode、不新增 public production API。
14. `real-resolver-runtime-implementation-entry-review`
   - 评审真实 resolver runtime preconditions 是否足以创建 production resolver runtime implementation task card。
   - 当前已落地 entry review：`docs/platform/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1.md`、`docs/task-cards/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1-plan.md`、`scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1.json` 与 `scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-review-v1.py` 固定 `real_resolver_runtime_implementation_entry_review_defined`；结论为 blocked before runtime task card。
15. `resolver-backend-profile-selection-readiness`
   - 固定真实 resolver runtime 之前的 backend profile selection 前置条件、reserved backend kind allowlist、environment / provider profile / policy version binding、operator approval、audit / rotation dependency、backend health reference、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard。
   - 当前已落地 readiness：`docs/platform/production-secret-backend-resolver-backend-profile-selection-readiness-v1.md`、`docs/task-cards/production-secret-backend-resolver-backend-profile-selection-readiness-v1-plan.md`、`scripts/checks/fixtures/production-secret-backend-resolver-backend-profile-selection-readiness-v1.json` 与 `scripts/check-production-ops-secret-backend-resolver-backend-profile-selection-readiness-v1.py` 固定 `resolver_backend_profile_selection_readiness_defined`；当前仍不创建 backend runtime、production resolver runtime、credential handle runtime、cloud secret service、DB provider、repository mode 或 public production API。下一步必须先选择 no leakage smoke runtime strategy、credential handle runtime boundary、operator approval runtime evidence、audit store handoff 或 backend health boundary 等剩余 blocker 单独推进。

## 验收口径

- 有任务卡、readiness fixture 和 checker 固定前置条件。
- `production_secret_backend` 仍为 `not_satisfied`，直到真实 resolver、test fixture strategy、测试环境 smoke 和生产前复核记录都完成。
- `test-fixture-strategy` 当前有 blocked entry review、fake resolver contract / no secret leakage smoke strategy 静态证据、fake resolver implementation task card entry readiness、fake resolver implementation 静态任务卡、fake resolver runtime implementation entry review、test-only fake resolver runtime、真实 resolver runtime preconditions、真实 resolver runtime implementation entry review 和 resolver backend profile selection readiness；这只能解释为当前需要继续拆剩余 blocker，不能解释为 production resolver runtime ready 或 production resolver runtime task card ready。
- fast baseline 不联网、不要求真实 credential、不写入真实 secret。
- `pwsh ./scripts/check-repo.ps1 -Fast` 通过。

## 停止线

- 不实现真实云 secret 服务。
- 不写入真实 secret。
- 不提交任何 API key、token、cookie、authorization header、证书或 provider raw dump。
- 不把 `.env.example`、developer env override、fake resolver 或 disabled resolver 写成 production secret backend。
- 不把 fake resolver static contract、no secret leakage smoke strategy、fake resolver implementation task card entry readiness、fake resolver runtime implementation entry review、fake resolver runtime implementation task card、真实 resolver runtime preconditions、真实 resolver runtime implementation entry review 或 resolver backend profile selection readiness 写成 production resolver runtime ready。
- 不创建 production resolver runtime implementation task card，直到 no leakage runtime gate、credential handle runtime boundary、operator approval runtime evidence、audit handoff 和 backend health boundary 形成可执行证据；backend profile selection readiness 当前只是静态前置证据，不代表 backend runtime created。
- 不声明 production ready。
- 不接 executor、confirmation、writeback、replay 或 materialized result reader。

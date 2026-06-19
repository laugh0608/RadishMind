# Production Secret Backend Operator Runbook / Negative Gates Readiness v1

更新时间：2026-06-19

## 文档目的

本文档用于固定 production secret backend 在进入 resolver implementation、真实云 secret backend 或 production ready 声明之前，operator runbook 与 negative gates 的最低证据边界。

本专题只定义人工启用流程、test / production secret source 要求、脱敏验证、smoke record、负向门禁、failure mapping、no fallback、no side effects 和 artifact guard；不实现 resolver runtime，不创建 fake resolver，不读取 secret value，不调用云 secret 服务，不访问 provider，不接 database connection provider，也不启用 workflow saved draft repository mode。

## 当前状态

状态：`operator_runbook_negative_gates_readiness_defined`

本批承接：

- `production-secret-backend-implementation-readiness`
- `secret-ref-schema-and-fixtures`
- `production-secret-backend-config-secret-ref-readiness-v1`
- `production-secret-backend-provider-profile-secret-binding-readiness-v1`
- `production-secret-backend-secret-resolver-interface-disabled-readiness-v1`

结论：

- `operator-runbook-and-negative-gates` 已推进为可检查证据。
- `operator-runbook` 前置条件已由本批 evidence 满足。
- `production_secret_backend` 仍为 `not_satisfied`。
- `resolver_implementation_status` 仍为 `not_started`。
- `resolver_runtime_status` 仍为 `not_created`。
- `fake_resolver_status` 仍为 `not_created`。
- `rotation-and-audit-policy` 仍未满足，不能声明 production secret backend ready。

## Operator Runbook Contract

后续真实 production secret backend 或 resolver implementation 如果打开，必须先具备人工启用与复核证据。runbook 只能记录 reference-only 和脱敏状态，不得记录 secret value、provider raw URL、DSN、cloud credential 或 opaque credential handle。

必须定义的 runbook sections：

- `purpose_and_scope`
- `environment_selection`
- `secret_source_inventory`
- `operator_approval_evidence`
- `sanitized_verification_steps`
- `negative_gate_review`
- `smoke_record_template`
- `rollback_or_disable_procedure`
- `production_ready_stop_line`

允许进入 operator evidence 的字段只有：

- `environment`
- `provider`
- `provider_profile`
- `secret_ref_status`
- `secret_backend_configured`
- `resolver_state`
- `operator_id_ref`
- `approval_ticket_ref`
- `request_id`
- `audit_ref`
- `runbook_version`
- `verification_status`
- `smoke_record_ref`
- `negative_gate_results`
- `timestamp`

禁止进入 operator evidence 的字段包括 raw secret、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail、opaque credential handle 和 resolver backend URL。

## Negative Gates

以下 gate 任一触发时，必须 fail closed，不能 fallback 到 developer env、mock provider、local-smoke profile、fixture credential、跨环境 `secret_ref`、fake resolver、fake query executor 或 committed secret value。

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `operator_runbook_missing` | `operator_gate` | 没有可审计 runbook |
| `operator_approval_missing` | `operator_gate` | 缺少人工 approval evidence |
| `operator_environment_mismatch` | `operator_gate` | runbook environment 与 secret ref / runtime environment 不一致 |
| `operator_secret_source_missing` | `operator_gate` | test 或 production secret source 未明确 |
| `operator_sanitized_verification_missing` | `operator_gate` | 缺少脱敏验证结果 |
| `operator_smoke_record_missing` | `operator_gate` | 缺少 smoke record reference |
| `operator_resolver_invocation_disabled` | `operator_gate` | 当前阶段尝试调用 resolver |
| `operator_secret_value_exposure_detected` | `operator_gate` | runbook、diagnostic 或 smoke record 暴露 secret-looking value |
| `operator_fallback_forbidden` | `operator_gate` | 出现 env / mock / fixture / fake resolver fallback |
| `operator_production_ready_claim_forbidden` | `operator_gate` | 在 test fixture strategy、真实 smoke 和生产前复核之前声明 production ready |

## Sanitized Diagnostics

允许输出：

- `credential_state`
- `secret_backend_configured`
- `secret_ref_present`
- `secret_ref_status`
- `resolver_state`
- `operator_gate_status`
- `negative_gate_results`
- `failure_code`
- `sanitized_diagnostic`
- `field_sources`
- `smoke_record_ref`
- `audit_ref`

禁止输出 raw secret、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail、opaque credential handle、resolver backend URL 或完整用户 claim。

## No Fallback

- 不从 `RADISHMIND_PLATFORM_API_KEY` 回退 production provider credential。
- 不从 `.env.example`、fixture、task card、README、runbook 或 committed config 中读取 secret value。
- 不把 disabled resolver、operator runbook 或 negative gate evidence 写成 production secret backend ready。
- 不把 smoke record reference 写成 credential resolved。
- 不允许 test `secret_ref` fallback 到 production，也不允许 production fallback 到 test。
- 不使用 fake resolver、fake query executor、mock provider 或 local-smoke profile 代替 production secret backend。

## No Side Effects

本批 checker 只读取 committed 文档、schema、fixture 和 `check-repo.py` 注册顺序，不读取真实环境变量、不连接网络、不调用云 secret 服务、不访问 provider、不创建文件、不修改 runtime 状态。

side effect counters 必须保持：

- `runbook_execution_count=0`
- `negative_gate_runtime_call_count=0`
- `secret_resolver_call_count=0`
- `fake_resolver_call_count=0`
- `secret_value_read_count=0`
- `cloud_secret_call_count=0`
- `provider_call_count=0`
- `database_connection_count=0`
- `credential_handle_created_count=0`
- `repository_mode_enablement_count=0`

## Artifact Guard

本批不得新增或启用以下 artifact：

- secret resolver runtime
- fake resolver
- operator runbook executor
- negative gate runtime
- cloud secret SDK
- secret value fixture
- production credential file
- provider credential runtime binding
- opaque credential handle runtime
- database connection provider
- DB driver / DSN parser
- connection factory
- workflow saved draft repository mode runtime
- SQL migration runner
- schema marker reader / writer
- public production API

## 后续推进

当前 readiness 只解除 `production-secret-backend-implementation-readiness` 中的 `operator-runbook-and-negative-gates` 阻塞。rotation / audit policy 已由 `production-secret-backend-rotation-audit-policy-readiness-v1` 单独固定；本文档不解除 `test-fixture-strategy`、真实 resolver implementation、云 secret backend 或 production ready 阻塞。

下一批如继续 production secret backend，应重新评审 test fixture strategy / fake resolver implementation 是否打开。任何 resolver runtime、cloud secret SDK、credential handle、DB provider、connection factory、SQL、schema marker、migration runner、repository mode 或 public production API 都必须作为独立实现目标重新开题。

## 验证

本专题由以下证据固定：

- `docs/task-cards/production-secret-backend-operator-runbook-negative-gates-readiness-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-operator-runbook-negative-gates-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-operator-runbook-negative-gates-readiness-v1.py`
- `scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json`
- `scripts/checks/fixtures/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py`

建议验证顺序：

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-operator-runbook-negative-gates-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-provider-profile-secret-binding-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-config-secret-ref-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/run-python.sh scripts/check-production-secret-reference-contract.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```

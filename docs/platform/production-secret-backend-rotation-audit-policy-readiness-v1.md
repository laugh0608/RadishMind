# Production Secret Backend Rotation / Audit Policy Readiness v1

更新时间：2026-06-19

## 文档目的

本文档用于固定 production secret backend 在声明 production ready 之前，secret rotation 与 audit policy 的最低证据边界。

本专题只定义 rotation trigger、approval / change window、secret ref versioning、rollback / disable policy、sanitized verification、audit event fields、failure mapping、no fallback、no side effects 和 artifact guard；不实现 rotation runtime，不写 audit store，不实现 resolver runtime，不创建 fake resolver，不读取 secret value，不调用云 secret 服务，不访问 provider，不接 database connection provider，也不启用 workflow saved draft repository mode。

## 当前状态

状态：`rotation_audit_policy_readiness_defined`

本批承接：

- `production-secret-backend-implementation-readiness`
- `secret-ref-schema-and-fixtures`
- `production-secret-backend-config-secret-ref-readiness-v1`
- `production-secret-backend-provider-profile-secret-binding-readiness-v1`
- `production-secret-backend-secret-resolver-interface-disabled-readiness-v1`
- `production-secret-backend-operator-runbook-negative-gates-readiness-v1`

结论：

- `rotation-and-audit-policy` 已推进为可检查证据。
- `rotation-and-audit-policy` 前置条件已由本批 evidence 满足。
- `production_secret_backend` 仍为 `not_satisfied`。
- `resolver_implementation_status` 仍为 `not_started`。
- `resolver_runtime_status` 仍为 `not_created`。
- `fake_resolver_status` 仍为 `not_created`。
- `test-fixture-strategy` 仍未满足，不能打开 resolver implementation。
- 本批不创建 production secret audit store，不能声明 production secret backend ready。

## Rotation Policy Contract

后续真实 production secret backend 或 resolver implementation 如果打开，必须先具备可审计的轮换策略。rotation policy 只能记录 reference-only 和脱敏状态，不得记录 secret value、provider raw URL、DSN、cloud credential、database hostname、database error detail 或 opaque credential handle。

必须定义的 rotation policy sections：

- `rotation_scope`
- `rotation_trigger_matrix`
- `approval_and_change_window`
- `secret_ref_versioning_policy`
- `rollback_or_disable_policy`
- `sanitized_rotation_verification`
- `post_rotation_smoke_record`
- `rotation_failure_mapping`
- `production_ready_stop_line`

必须覆盖的 rotation trigger 包括：

- operator-requested rotation
- provider credential compromise suspicion
- scheduled rotation window
- provider profile binding changed
- secret ref version changed
- resolver backend policy changed
- failed sanitized smoke after rotation

## Audit Policy Contract

audit policy 只允许记录脱敏字段、引用字段和策略版本，不得记录 secret payload、credential handle、provider raw URL 或完整用户 claim。

允许进入 audit event 的字段只有：

- `event_id`
- `event_kind`
- `environment`
- `provider`
- `provider_profile`
- `secret_ref_status`
- `secret_ref_version_ref`
- `operator_id_ref`
- `approval_ticket_ref`
- `request_id`
- `audit_ref`
- `runbook_version`
- `policy_version`
- `rotation_window_ref`
- `verification_status`
- `failure_code`
- `timestamp`

禁止进入 audit event 的字段包括 raw secret、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail、opaque credential handle、resolver backend URL 和完整用户 claim。

## Failure Mapping

以下 code 任一触发时，必须 fail closed，不能 fallback 到 developer env、mock provider、local-smoke profile、fixture credential、跨环境 `secret_ref`、fake resolver、fake query executor 或 committed secret value。

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `rotation_policy_missing` | `rotation_audit_policy` | 缺少可审计 rotation policy |
| `rotation_trigger_missing` | `rotation_audit_policy` | 缺少适用的 rotation trigger |
| `rotation_approval_missing` | `rotation_audit_policy` | 缺少 approval evidence |
| `rotation_window_missing` | `rotation_audit_policy` | 缺少 change window 或 rotation window reference |
| `rotation_secret_ref_version_missing` | `rotation_audit_policy` | 缺少 secret ref version reference |
| `rotation_sanitized_verification_missing` | `rotation_audit_policy` | 缺少脱敏验证结果 |
| `rotation_smoke_record_missing` | `rotation_audit_policy` | 缺少 post-rotation smoke record reference |
| `rotation_audit_event_missing` | `rotation_audit_policy` | 缺少 audit event policy evidence |
| `rotation_audit_secret_exposure_detected` | `rotation_audit_policy` | rotation 或 audit evidence 暴露 secret-looking value |
| `rotation_fallback_forbidden` | `rotation_audit_policy` | 出现 env / mock / fixture / fake resolver fallback |
| `rotation_production_ready_claim_forbidden` | `rotation_audit_policy` | 在 fake resolver strategy、真实 smoke 和生产前复核之前声明 production ready |

## Sanitized Diagnostics

允许输出：

- `rotation_policy_status`
- `audit_policy_status`
- `secret_backend_configured`
- `secret_ref_present`
- `secret_ref_status`
- `secret_ref_version_status`
- `resolver_state`
- `operator_gate_status`
- `rotation_gate_status`
- `failure_code`
- `sanitized_diagnostic`
- `audit_ref`
- `policy_version`
- `rotation_window_ref`
- `smoke_record_ref`
- `timestamp`

禁止输出 raw secret、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail、opaque credential handle、resolver backend URL 或完整用户 claim。

## No Fallback

- 不从 `RADISHMIND_PLATFORM_API_KEY` 回退 production provider credential。
- 不从 `.env.example`、fixture、task card、README、runbook、rotation policy 或 committed config 中读取 secret value。
- 不把 rotation policy、audit policy、operator runbook、negative gate evidence 或 disabled resolver 写成 production secret backend ready。
- 不把 audit reference、policy version、secret ref version reference 或 smoke record reference 写成 credential resolved。
- 不允许 test `secret_ref` fallback 到 production，也不允许 production fallback 到 test。
- 不使用 fake resolver、fake query executor、mock provider 或 local-smoke profile 代替 production secret backend。
- audit policy evidence 不代表 production secret audit store ready。

## No Side Effects

本批 checker 只读取 committed 文档、schema、fixture 和 `check-repo.py` 注册顺序，不读取真实环境变量、不连接网络、不调用云 secret 服务、不访问 provider、不执行 rotation、不写 audit store、不创建文件、不修改 runtime 状态。

side effect counters 必须保持：

- `rotation_execution_count=0`
- `audit_event_write_count=0`
- `audit_store_call_count=0`
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

- rotation runtime
- production secret audit store
- audit writer
- cloud secret SDK
- secret resolver runtime
- fake resolver
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

当前 readiness 只解除 `production-secret-backend-implementation-readiness` 中的 `rotation-and-audit-policy` 阻塞，不解除 `test-fixture-strategy`、真实 resolver implementation、fake resolver implementation、云 secret backend 或 production ready 阻塞。后续已由 `production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1` 固定 `test_fixture_strategy_fake_resolver_entry_review_defined`，但该评审结论仍是 fake resolver implementation entry 不打开。

下一批如继续 production secret backend，应在 fake resolver implementation task card、真实 resolver runtime preconditions 或其它单一前置方向中重新开题。任何 resolver runtime、cloud secret SDK、credential handle、DB provider、connection factory、SQL、schema marker、migration runner、repository mode、production secret audit store、audit writer 或 public production API 都必须作为独立实现目标重新开题。

## 验证

本专题由以下证据固定：

- `docs/task-cards/production-secret-backend-rotation-audit-policy-readiness-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-rotation-audit-policy-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-rotation-audit-policy-readiness-v1.py`
- `scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json`
- `scripts/checks/fixtures/production-secret-backend-operator-runbook-negative-gates-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-operator-runbook-negative-gates-readiness-v1.py`

建议验证顺序：

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-rotation-audit-policy-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-operator-runbook-negative-gates-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-provider-profile-secret-binding-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-config-secret-ref-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/run-python.sh scripts/check-production-secret-reference-contract.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```

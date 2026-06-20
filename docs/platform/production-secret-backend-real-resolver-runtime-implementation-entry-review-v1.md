# Production Secret Backend Real Resolver Runtime Implementation Entry Review v1

更新时间：2026-06-20

## 文档目的

本文档评审 production secret backend 是否可以从真实 resolver runtime preconditions 进入 production resolver runtime implementation task card。

对应切片：`production-secret-backend-real-resolver-runtime-implementation-entry-review-v1`。

结论：状态为 `real_resolver_runtime_implementation_entry_review_defined`，entry decision 为 `real_resolver_runtime_implementation_blocked_before_task_card`。当前只能确认真实 resolver runtime 的启用条件和停止线已经可检查；仍不能创建 production resolver runtime implementation task card，原因是 resolver backend profile selection、no secret leakage smoke runtime gate、credential handle runtime boundary、operator approval runtime evidence、production secret audit store handoff 和 backend health boundary 尚未形成可执行证据。

本批不实现 production resolver runtime，不读取真实 secret，不调用云 secret 服务，不连接数据库，不创建 credential payload，不创建 credential handle runtime，不创建 no secret leakage smoke runtime，不启用 workflow saved draft repository mode，也不新增 public production API。

## 输入证据

- `production-secret-backend-real-resolver-runtime-preconditions-v1` 已固定 `real_resolver_runtime_preconditions_defined`。
- `production-secret-backend-fake-resolver-runtime-implementation-v1` 已固定 `fake_resolver_runtime_test_only_implemented`，但只服务离线测试边界。
- `production-secret-backend-secret-resolver-interface-disabled-readiness-v1` 仍确认 production resolver interface 默认 disabled。
- `production-secret-backend-provider-profile-secret-binding-readiness-v1` 只固定 reference-only provider/profile binding，不表示 credential resolved。
- `production-secret-backend-operator-runbook-negative-gates-readiness-v1` 和 `production-secret-backend-rotation-audit-policy-readiness-v1` 只固定运行手册与策略，不提供 runtime audit store 或真实 operator approval execution。
- `production-ops-secret-backend-implementation-readiness` 仍保持 `production_secret_backend=not_satisfied`、`resolver_runtime_status=not_created`。

## Entry Review Decision

| gate | 本次结论 | 说明 |
| --- | --- | --- |
| real resolver preconditions | `satisfied_static_preconditions` | 启用条件、停止线、failure mapping、diagnostics 和 artifact guard 已定义 |
| implementation task card | `blocked_before_task_card` | 不允许创建 production resolver runtime implementation task card |
| resolver backend profile | `blocked_missing_backend_profile_selection` | 尚未选择 backend kind、allowed environment、policy version 和 health boundary |
| no secret leakage smoke runtime | `blocked_missing_runtime_gate` | 尚未创建真实 resolver 的离线泄漏扫描 runtime gate |
| credential handle boundary | `blocked_missing_runtime_boundary` | 只定义 future result 字段，尚无 runtime handle boundary / lifecycle |
| operator approval runtime evidence | `blocked_missing_runtime_evidence` | 只有 runbook policy，没有真实启用执行证据 |
| audit / rotation runtime handoff | `blocked_missing_runtime_handoff` | 只有 policy，没有 production audit store、audit writer 或 rotation execution |
| production resolver runtime | `not_created` | 本批不创建 runtime |
| database / repository mode | `blocked` | DB provider、SQL、schema marker、repository mode 和 production API 均不打开 |

## Blocked Conditions

后续创建真实 resolver runtime implementation task card 前，至少需要单独完成或评审：

- `resolver-backend-profile-selection-readiness`
- `no-secret-leakage-smoke-runtime-strategy`
- `credential-handle-runtime-boundary-readiness`
- `operator-approval-runtime-evidence-readiness`
- `production-secret-audit-store-handoff-readiness`
- `resolver-backend-health-boundary-readiness`

这些 blocker 不能用 fake resolver、developer env、fixture credential、mock provider、local-smoke profile、DB provider 或 repository memory store 替代。

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `real_resolver_runtime_entry_preconditions_missing` | `implementation_entry` | 真实 resolver runtime preconditions 不存在或未被消费 |
| `real_resolver_runtime_entry_task_card_blocked` | `implementation_gate` | 当前仍不能创建 production resolver runtime implementation task card |
| `real_resolver_runtime_entry_backend_profile_missing` | `backend_profile` | resolver backend profile selection 尚未完成 |
| `real_resolver_runtime_entry_no_leakage_gate_missing` | `no_secret_leakage` | 真实 resolver no leakage runtime gate 尚未定义 |
| `real_resolver_runtime_entry_credential_handle_boundary_missing` | `credential_boundary` | credential handle runtime boundary 尚未定义 |
| `real_resolver_runtime_entry_operator_approval_evidence_missing` | `operator_gate` | operator approval runtime evidence 尚未定义 |
| `real_resolver_runtime_entry_audit_handoff_missing` | `audit_policy` | production audit store / writer handoff 尚未定义 |
| `real_resolver_runtime_entry_secret_value_detected` | `artifact_guard` | 文档、fixture 或 checker 出现 secret-looking value |
| `real_resolver_runtime_created_in_entry_review` | `artifact_guard` | 本批创建 production resolver runtime、cloud client、credential payload 或 credential handle runtime |
| `real_resolver_runtime_entry_cloud_call_forbidden` | `no_side_effects` | checker、fixture 或任务卡要求联网、provider call 或云 secret call |
| `real_resolver_runtime_entry_repository_mode_forbidden` | `repository_mode` | entry review 被用于打开 workflow saved draft repository mode 成功路径 |
| `real_resolver_runtime_entry_scope_overreach` | `implementation_boundary` | 本批把 DB provider、SQL、schema marker、audit store 或 production API 合入 entry review |

所有失败都必须 fail closed，不返回 secret value、DSN、provider raw URL、database hostname、database error detail、credential payload、完整 secret ref value、完整 credential handle 或完整用户 claim。

## Sanitized Diagnostics

允许输出：

- `real_resolver_runtime_entry_status`
- `real_resolver_runtime_preconditions_status`
- `resolver_runtime_status`
- `resolver_backend_profile_status`
- `no_secret_leakage_gate_status`
- `credential_handle_boundary_status`
- `operator_approval_status`
- `audit_policy_status`
- `rotation_policy_status`
- `failure_code`
- `sanitized_diagnostic`
- `request_id`
- `audit_ref`
- `policy_version`

禁止输出 raw secret、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail、credential payload、resolver backend URL、full secret ref value、full credential handle、完整用户 claim、authorization header 或 cookie。

## No Fallback

- 不允许 production resolver fallback 到 fake resolver。
- 不允许 fake resolver fallback 到 production resolver。
- 不允许 resolver failure fallback 到 `RADISHMIND_PLATFORM_API_KEY`、developer env plaintext、fixture credential、committed value、mock provider、local-smoke profile、fake query executor、sample 或 repository memory store。
- 不允许 test `secret_ref` fallback 到 production，也不允许 production fallback 到 test。
- 不把 real resolver runtime preconditions 或 entry review 写成 credential resolved、production resolver ready、production secret backend ready 或 production API ready。

## No Side Effects

本批 checker 只读取 committed 文档、fixture、schema、已有 readiness fixture 和 `check-repo.py` 注册顺序，不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不调用 fake resolver、不调用 resolver、不创建 credential payload、不创建 credential handle、不连接数据库、不打开 driver、不运行 SQL、不读写 schema marker、不启用 repository mode、不写 audit store、不调用 production API。

side effect counters 必须保持：

- `real_secret_read_count=0`
- `secret_resolver_call_count=0`
- `fake_resolver_call_count=0`
- `cloud_secret_call_count=0`
- `provider_call_count=0`
- `credential_payload_created_count=0`
- `credential_handle_created_count=0`
- `database_connection_count=0`
- `driver_open_count=0`
- `sql_execution_count=0`
- `schema_marker_read_count=0`
- `schema_marker_write_count=0`
- `repository_mode_enablement_count=0`
- `production_audit_store_write_count=0`
- `production_api_call_count=0`

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1.md`
- `docs/task-cards/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1.json`
- `scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-review-v1.py`

不得新增或启用以下 artifact：

- production resolver runtime
- production resolver implementation task card
- production resolver backend client
- cloud secret SDK
- secret value fixture
- production credential file
- credential payload runtime
- credential handle runtime
- no secret leakage smoke runtime
- database connection provider
- DB driver / DSN parser
- connection factory
- SQL migration
- schema marker reader / writer
- migration runner
- workflow saved draft repository mode runtime
- production secret audit store
- audit writer
- public production API

## 后续推进

下一步若继续 production secret backend，不应直接创建 resolver runtime implementation task card。应在以下方向中选择一个单独开题：

1. `resolver-backend-profile-selection-readiness`
2. `no-secret-leakage-smoke-runtime-strategy`
3. `credential-handle-runtime-boundary-readiness`
4. `operator-approval-runtime-evidence-readiness`
5. `production-secret-audit-store-handoff-readiness`

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-preconditions-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-contract.py
./scripts/run-python.sh scripts/check-production-secret-reference-contract.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```

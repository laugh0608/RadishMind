# Production Secret Backend Secret Resolver Interface Disabled Readiness v1

更新时间：2026-06-19

## 文档目的

本文档用于固定 production secret backend 在进入 resolver implementation 或 operator runbook 之前，secret resolver interface 的默认禁用边界。

本专题只定义 future resolver interface 的输入形状、disabled result、sanitized diagnostics、failure mapping、no fallback、no side effects 和 artifact guard；不实现 resolver runtime，不创建 fake resolver，不读取 secret value，不调用云 secret 服务，不访问 provider，不接 database connection provider，也不启用 workflow saved draft repository mode。

## 当前状态

状态：`secret_resolver_interface_disabled_readiness_defined`

本批承接：

- `production-secret-backend-implementation-readiness`
- `secret-ref-schema-and-fixtures`
- `production-secret-backend-config-secret-ref-readiness-v1`
- `production-secret-backend-provider-profile-secret-binding-readiness-v1`
- `production-secret-reference-basic`

结论：

- `secret-resolver-interface-disabled` 已推进为可检查证据。
- `production_secret_backend` 仍为 `not_satisfied`。
- `resolver_implementation_status` 仍为 `not_started`。
- `resolver_runtime_status` 仍为 `not_created`。
- `fake_resolver_status` 仍为 `not_created`。
- disabled interface 只返回脱敏失败状态，不返回 credential handle。

## Disabled Resolver Interface Contract

后续 resolver implementation 如果打开，只能消费 reference-only 输入，不得把 secret value、provider raw URL、DSN、cloud provider SDK 参数或本机 credential path 作为输入。

允许进入 resolver input contract 的字段只有：

- `environment`
- `provider`
- `provider_profile`
- `credential_requirement`
- `secret_ref_status`
- `secret_backend_configured`
- `request_id`
- `audit_ref`
- `caller_purpose`

允许进入 disabled result contract 的字段只有：

- `resolver_state`
- `secret_backend_configured`
- `resolver_available`
- `credential_handle_present`
- `failure_code`
- `sanitized_diagnostic`
- `retryable`

当前 disabled result 必须满足：

| field | value |
| --- | --- |
| `resolver_state` | `disabled` |
| `secret_backend_configured` | `false` |
| `resolver_available` | `false` |
| `credential_handle_present` | `false` |
| `retryable` | `false` |

## Failure Mapping

后续实现必须使用 fail-closed 失败语义：

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `secret_resolver_secret_ref_missing` | `configuration` | provider/profile 需要 credential，但 secret ref 缺失 |
| `secret_resolver_backend_disabled` | `configuration` | secret ref 存在，但 secret backend 仍禁用 |
| `secret_resolver_unavailable` | `configuration` | future resolver interface 不可用或未启用 |
| `secret_resolution_denied` | `configuration` | caller purpose、audit ref 或环境绑定不允许解析 |
| `secret_resolver_environment_mismatch` | `configuration` | resolver input environment 与 secret ref environment 不一致 |
| `secret_resolver_invocation_disabled` | `configuration` | 当前切片或 readiness 检查中出现 resolver invocation 企图 |

这些失败不能 fallback 到 `RADISHMIND_PLATFORM_API_KEY`、developer env plaintext、`mock` provider、`local-smoke` profile、fixture credential、committed secret value、test / production 跨环境 `secret_ref`、fake resolver 或 fake query executor。

## Sanitized Diagnostics

允许输出：

- `credential_state`
- `secret_backend_configured`
- `secret_ref_present`
- `secret_ref_status`
- `resolver_state`
- `resolver_available`
- `credential_handle_present`
- `failure_code`
- `sanitized_diagnostic`
- `field_sources`

禁止输出 raw secret、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail、opaque credential handle、resolver backend URL 或完整用户 claim。

## No Fallback

- 不从 `RADISHMIND_PLATFORM_API_KEY` 回退 production provider credential。
- 不从 `.env.example`、fixture、task card、README 或 committed config 中读取 secret value。
- 不把 disabled resolver 写成 production secret backend ready。
- 不把 disabled result 写成 credential resolved。
- 不允许 test `secret_ref` fallback 到 production，也不允许 production fallback 到 test。
- 不使用 fake resolver、fake query executor、mock provider 或 local-smoke profile 代替 production secret backend。

## No Side Effects

本批 checker 只读取 committed 文档、schema、fixture 和 `check-repo.py` 注册顺序，不读取真实环境变量、不连接网络、不调用云 secret 服务、不访问 provider、不创建文件、不修改 runtime 状态。

side effect counters 必须保持：

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

## 后续推进

当前 readiness 只解除 `production-secret-backend-implementation-readiness` 中的 `secret-resolver-interface-disabled` 阻塞。operator runbook / negative gates 已由 `production-secret-backend-operator-runbook-negative-gates-readiness-v1` 单独固定；本文档仍不解除 rotation / audit policy、test fixture strategy / fake resolver implementation 或 production ready 阻塞。

下一批如继续 production secret backend，应推进 rotation / audit policy，或重新评审 test fixture strategy / fake resolver implementation 是否打开。

## 验证

本专题由以下证据固定：

- `docs/task-cards/production-secret-backend-secret-resolver-interface-disabled-readiness-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py`
- `scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json`
- `scripts/checks/fixtures/production-secret-backend-provider-profile-secret-binding-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-provider-profile-secret-binding-readiness-v1.py`

建议验证顺序：

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-provider-profile-secret-binding-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-config-secret-ref-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/run-python.sh scripts/check-production-secret-reference-contract.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```

# Production Secret Backend Fake Resolver Runtime Implementation v1

更新时间：2026-06-20

## 文档目的

本文档固定 production secret backend 的 test-only fake resolver runtime 实现结果，明确该 runtime 的启用门禁、输入输出、诊断、泄漏检查、side effect counters 和停止线。

对应切片：`production-secret-backend-fake-resolver-runtime-implementation-v1`。

结论：状态为 `fake_resolver_runtime_test_only_implemented`。本批实现 `services/platform/internal/secretbackend/fake_resolver.go`，并以 `services/platform/internal/secretbackend/fake_resolver_test.go` 覆盖默认 disabled、placeholder secret ref、environment binding、opaque metadata、sanitized diagnostics、offline no leakage smoke 和 zero external side effects。该 runtime 默认 disabled，只能由测试显式启用；不实现 production resolver runtime，不解析真实 secret，不读取真实环境 secret，不调用云 secret 服务，不访问 provider，不创建 credential payload，不连接数据库，不启用 workflow saved draft repository mode 或 production API。

## 输入证据

- `production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1` 已固定 `fake_resolver_runtime_implementation_entry_review_defined`。
- `production-secret-backend-fake-resolver-implementation-v1` 已固定 fake resolver runtime 必须 test-only、disabled-by-default、offline no leakage smoke。
- `production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1` 已固定 input / output allowlist、禁止字段和 no leakage strategy。
- `production-ops-secret-backend-implementation-readiness` 已推进到 `fake_resolver_runtime_test_only_implemented`，但 `production_secret_backend` 仍为 `not_satisfied`，真实 resolver runtime 仍为 `not_created`。

## 本批实现范围

本批只打开以下 test-only runtime 边界：

- 新增 `FakeResolver`，零值 / `NewDisabledFakeResolver()` 永远返回 `fake_resolver_runtime_disabled`。
- `NewFakeResolver(FakeResolverConfig{Enabled: true, ...})` 仅供测试显式启用。
- `Resolve` 只消费结构化请求字段：`environment`、`provider`、`provider_profile`、`secret_ref_key`、`secret_ref_version`、`purpose`、`request_id`、`audit_ref`、`policy_version`。
- 成功输出只返回 `credential_handle_id`、`credential_kind=opaque_test_credential_handle` 和绑定元数据，不返回 credential payload、raw secret、DSN、provider raw URL 或 database hostname。
- 失败输出只返回 `failure_code`、`sanitized_diagnostic`、`request_id`、`audit_ref` 和 `policy_version`，并 fail closed。
- Go 单测覆盖 no leakage 和 zero external side effects。

本批不接 `config.LoadFromEnv`、HTTP route、provider runtime、DB provider、repository mode selector、migration runner、audit store 或 public production API。

## 开发者使用说明

`FakeResolver` 位于 `services/platform/internal/secretbackend`，属于 Go `internal` 包，只能被 `services/platform` 模块内代码或测试消费。它不是 public SDK，也不是 production resolver adapter。

默认行为必须 fail closed：

- `FakeResolver{}` 与 `NewDisabledFakeResolver()` 都返回 `fake_resolver_runtime_disabled`。
- `NewFakeResolver(FakeResolverConfig{Enabled: true, ...})` 只能在测试或离线 smoke 中显式创建。
- allowlist 为空时不允许任何 environment、provider、provider profile 或 placeholder secret ref。
- `PolicyVersion` 未配置时使用 `fake-resolver-runtime-v1`，请求侧仍必须显式带同一 policy version。

测试中允许的最小调用形态如下：

```go
resolver := secretbackend.NewFakeResolver(secretbackend.FakeResolverConfig{
    Enabled:                 true,
    AllowedEnvironments:     []string{"test"},
    AllowedProviders:        []string{"mock"},
    AllowedProviderProfiles: []string{"local-smoke"},
    AllowedSecretRefKeys:    []string{"placeholder/provider/mock/local-smoke"},
    PolicyVersion:           secretbackend.DefaultFakeResolverPolicyVersion,
})

result := resolver.Resolve(secretbackend.FakeResolverInput{
    Environment:      "test",
    Provider:         "mock",
    ProviderProfile:  "local-smoke",
    SecretRefKey:     "placeholder/provider/mock/local-smoke",
    SecretRefVersion: "v1",
    Purpose:          "workflow-saved-draft-database-smoke",
    RequestID:        "req-runtime-smoke-001",
    AuditRef:         "audit-runtime-smoke-001",
    PolicyVersion:    secretbackend.DefaultFakeResolverPolicyVersion,
})
```

调用方只能把成功结果里的 `credential_handle_id` 当作 opaque test metadata，用于证明绑定和脱敏路径可运行；不得把它传给 provider、database driver、repository mode selector、audit writer 或 production API。失败结果不得被重试到真实 resolver，也不得回退到 developer env、fixture secret、mock provider credential 或 sample。

新增消费方必须同步验证：

- disabled / zero value fail closed；
- unsupported environment、provider、provider profile 和 placeholder secret ref fail closed；
- secret-looking input 不会被 echo；
- `SideEffectCounters` 全部为 `0`；
- JSON 结果不包含 credential payload、raw secret、provider raw URL、DSN、database hostname、token、cookie 或 authorization header。

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `fake_resolver_runtime_disabled` | `implementation_gate` | runtime 未显式 test enable |
| `fake_resolver_runtime_request_audit_missing` | `sanitized_diagnostics` | 缺少 request / audit metadata |
| `fake_resolver_runtime_policy_missing` | `implementation_gate` | policy version 缺失或不匹配 |
| `fake_resolver_runtime_environment_binding_missing` | `environment_binding` | 缺少 environment |
| `fake_resolver_runtime_environment_denied` | `environment_binding` | environment 不在 test allowlist |
| `fake_resolver_runtime_provider_binding_missing` | `provider_binding` | 缺少 provider |
| `fake_resolver_runtime_provider_denied` | `provider_binding` | provider 不在 allowlist |
| `fake_resolver_runtime_provider_profile_denied` | `provider_binding` | provider profile 不在 allowlist |
| `fake_resolver_runtime_placeholder_secret_ref_missing` | `test_fixture` | placeholder secret ref 缺失或不在 allowlist |
| `fake_resolver_runtime_secret_value_detected` | `artifact_guard` | 输入出现 secret-looking value |
| `fake_resolver_runtime_purpose_missing` | `credential_boundary` | 缺少 purpose |
| `fake_resolver_runtime_opaque_handle_metadata_missing` | `credential_boundary` | 无法生成 opaque metadata |
| `fake_resolver_runtime_scope_overreach` | `implementation_boundary` | runtime 引入生产 resolver、云、DB、SQL、repository mode 或 production API |

所有失败都必须 fail closed，不返回 secret value、provider raw URL、DSN、database hostname、database error detail、credential payload、完整 secret ref value 或完整用户 claim。

## Sanitized Diagnostics

允许输出：

- `fake_resolver_runtime_implementation_status`
- `fake_resolver_runtime_status`
- `no_secret_leakage_smoke_runtime_status`
- `secret_ref_status`
- `environment_binding_status`
- `failure_code`
- `sanitized_diagnostic`
- `request_id`
- `audit_ref`
- `policy_version`

禁止输出 raw secret、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail、credential payload、resolver backend URL、full secret ref value 或完整用户 claim。

## No Fallback

- 不允许 fake resolver fallback 到 production resolver。
- 不允许 production resolver fallback 到 fake resolver。
- 不允许 test secret ref fallback 到 production secret ref。
- 不允许 resolver failure fallback 到 committed value、developer env plaintext、fixture credential、mock provider、local-smoke profile、fake query executor 或 sample。
- `fake_resolver_runtime_test_only_implemented` 不表示 production credential resolved、production secret backend ready、cloud secret backend ready、repository mode ready 或 production ready。

## No Side Effects

runtime 只读取 request struct 字段，不读取环境变量，不连接网络，不调用云 secret 服务，不访问 provider，不调用 production resolver，不创建 credential payload，不连接数据库，不打开 driver，不执行 SQL，不读写 schema marker，不启用 repository mode，不写 audit store，不调用 production API。

side effect counters 必须保持：

- `secret_resolver_call_count=0`
- `secret_value_read_count=0`
- `cloud_secret_call_count=0`
- `provider_call_count=0`
- `credential_handle_created_count=0`
- `database_connection_count=0`
- `driver_open_count=0`
- `sql_execution_count=0`
- `schema_marker_read_count=0`
- `schema_marker_write_count=0`
- `repository_mode_enablement_count=0`
- `production_api_call_count=0`
- `production_audit_store_write_count=0`

## Artifact Guard

本批允许新增：

- `services/platform/internal/secretbackend/fake_resolver.go`
- `services/platform/internal/secretbackend/fake_resolver_test.go`

不得新增或启用：

- production resolver runtime
- no secret leakage smoke runtime / runner
- cloud secret SDK
- production credential file
- secret value fixture
- credential payload runtime
- database connection provider
- DB driver / DSN parser
- connection factory
- SQL migration
- schema marker reader / writer
- migration runner
- workflow saved draft repository mode runtime
- production secret audit store
- public production API

## 后续推进

后续已由 `production-secret-backend-real-resolver-runtime-preconditions-v1` 固定真实 resolver runtime 前置，由 `production-secret-backend-real-resolver-runtime-implementation-entry-review-v1` 确认 production resolver runtime task card 当前 blocked，并由 `production-secret-backend-resolver-backend-profile-selection-readiness-v1`、`production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1`、`production-secret-backend-credential-handle-runtime-boundary-readiness-v1`、`production-secret-backend-operator-approval-runtime-evidence-readiness-v1`、`production-secret-backend-audit-store-handoff-readiness-v1`、`production-secret-backend-resolver-backend-health-boundary-readiness-v1` 与 `production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1` 固定 backend profile selection、no leakage strategy、credential handle boundary、operator approval runtime evidence、audit store handoff、backend health boundary 和 backend health runtime entry review 静态前置。当前 test-only fake resolver runtime 只为离线验证提供安全替身，不是 production secret backend。

## 验证

建议验证顺序：

```bash
cd services/platform && go test ./internal/secretbackend
cd ../..
./scripts/run-python.sh scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-fake-resolver-implementation-v1.py
./scripts/check-repo.sh --fast
```

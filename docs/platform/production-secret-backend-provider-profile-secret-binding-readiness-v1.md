# Production Secret Backend Provider Profile Secret Binding Readiness v1

更新时间：2026-06-19

## 文档目的

本文档用于固定 production secret backend 在进入 resolver interface 或 operator runbook 之前，provider/profile 到 `secret_ref` 的绑定契约。

本专题只定义 provider profile 如何声明 credential requirement、`secret_ref_status`、环境绑定、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard；不实现 resolver runtime，不创建 fake resolver，不读取 secret value，不调用云 secret 服务，不接 database connection provider，也不启用 workflow saved draft repository mode。

## 当前状态

状态：`provider_profile_secret_binding_readiness_defined`

本批承接：

- `production-secret-backend-implementation-readiness`
- `secret-ref-schema-and-fixtures`
- `production-secret-backend-config-secret-ref-readiness-v1`
- `production-secret-reference-basic`

结论：

- `provider-profile-secret-binding` 已推进为可检查证据。
- `production_secret_backend` 仍为 `not_satisfied`。
- `resolver_implementation_status` 仍为 `not_started`。
- `production-secret-reference-basic` 仍为 reference-only manifest。
- `secret_ref_status=present` 只表示引用存在，不表示 credential resolved。

## Provider/Profile Binding Contract

后续 provider/profile inventory 或 request-side selection 只能把 credential requirement 与 secret reference 绑定状态作为脱敏状态传递，不得读取或解析 secret value。

允许进入绑定契约的字段只有：

- `environment`
- `provider`
- `provider_profile`
- `credential_requirement`
- `secret_ref_status`
- `secret_ref_present`
- `missing_secret_refs`
- `field_sources`

`credential_requirement` 的允许值为：

| value | 语义 |
| --- | --- |
| `required` | 该 provider/profile 在对应环境必须有 `secret_ref` |
| `optional` | 后续独立任务可定义的可选 credential，不得用于 production required provider |
| `not_required` | 只允许用于 `mock`、`local-smoke` 或不需要 credential 的开发态 profile |

`secret_ref_status` 的允许值为：

| value | 语义 |
| --- | --- |
| `present` | reference-only manifest 中存在 `ref:` 引用 |
| `missing` | credential required，但缺少 `secret_ref` |
| `not_required` | provider/profile 不需要 credential，不能升级为 production credential ready |

## 环境绑定

test 与 production 的 `secret_ref` 必须分离，不能跨环境 fallback。provider/profile 只允许使用短键，例如 `openai-compatible/default`；长语义、真实 credential、provider raw URL 和云 provider SDK 参数都不得写入 profile key 或 committed fixture。

当前 `production-secret-reference-basic` 只覆盖：

- `test` / `openai-compatible` / `default`
- `production` / `openai-compatible` / `default`

这两个 binding 都是 `credential_requirement=required` 且 `secret_ref_status=present` 的 reference-only 样例。它们不证明 resolver 可用，不证明 credential 已解析，也不证明 production secret backend ready。

## Failure Mapping

后续实现必须使用 fail-closed 失败语义：

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `provider_profile_binding_missing` | `configuration` | provider/profile 没有可验证 binding record |
| `provider_profile_credential_required` | `configuration` | provider/profile 要求 credential，但未声明 requirement 或引用策略 |
| `provider_profile_secret_ref_missing` | `configuration` | `credential_requirement=required` 但 `secret_ref_status=missing` |
| `provider_profile_environment_mismatch` | `configuration` | test / production binding 交叉或环境不匹配 |
| `provider_profile_secret_backend_disabled` | `configuration` | binding 存在，但 production secret backend 仍禁用 |
| `provider_profile_resolver_forbidden` | `configuration` | 当前切片或 readiness 检查中出现 resolver 调用企图 |

这些失败不能 fallback 到 `RADISHMIND_PLATFORM_API_KEY`、developer env plaintext、`mock` provider、`local-smoke` profile、fixture credential、committed secret value 或跨环境 `secret_ref`。

## Sanitized Diagnostics

允许输出：

- `credential_state`
- `secret_backend_configured`
- `secret_ref_present`
- `secret_ref_status`
- `missing_secret_refs`
- `field_sources`
- `environment`
- `provider`
- `provider_profile`

禁止输出 raw secret、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail 或完整用户 claim。`secret_ref` 本身只允许保留在 reference-only manifest 中，不进入 config summary、diagnostics、request metadata 或 audit payload。

## No Fallback

- 不从 `RADISHMIND_PLATFORM_API_KEY` 回退 production provider credential。
- 不把 `mock`、`local-smoke`、demo profile 或 `not_required` binding 写成 production provider credential ready。
- 不从 `.env.example`、fixture、task card、README 或 committed config 中读取 secret value。
- 不把 `secret_ref_status=present` 写成 credential resolved。
- 不允许 test `secret_ref` fallback 到 production，也不允许 production fallback 到 test。

## No Side Effects

本批 checker 只读取 committed 文档、schema、fixture 和 `check-repo.py` 注册顺序，不读取真实环境变量、不连接网络、不调用云 secret 服务、不访问 provider、不创建文件、不修改 runtime 状态。

side effect counters 必须保持：

- `secret_resolver_call_count=0`
- `fake_resolver_call_count=0`
- `secret_value_read_count=0`
- `cloud_secret_call_count=0`
- `provider_call_count=0`
- `database_connection_count=0`
- `repository_mode_enablement_count=0`

## Artifact Guard

本批不得新增或启用以下 artifact：

- secret resolver runtime
- fake resolver
- cloud secret SDK
- secret value fixture
- production credential file
- provider credential runtime binding
- database connection provider
- DB driver / DSN parser
- connection factory
- workflow saved draft repository mode runtime
- SQL migration runner
- schema marker reader / writer

## 后续推进

当前 readiness 只解除 `production-secret-backend-implementation-readiness` 中的 `provider-profile-binding` 阻塞。disabled resolver interface、operator runbook / negative gates 和 rotation / audit policy 已分别由后续 readiness 单独固定；本文档仍不解除 test fixture strategy / fake resolver implementation 或 production ready 阻塞。

下一批如继续 production secret backend，应重新评审 test fixture strategy / fake resolver implementation 是否打开。

每个方向都必须继续保持 no raw secret、no cloud call、no fallback 和 no production ready 声明。

## 验证

本专题由以下证据固定：

- `docs/task-cards/production-secret-backend-provider-profile-secret-binding-readiness-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-provider-profile-secret-binding-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-provider-profile-secret-binding-readiness-v1.py`
- `scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json`
- `scripts/checks/fixtures/production-secret-reference-basic.json`
- `scripts/check-production-secret-reference-contract.py`

建议验证顺序：

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-provider-profile-secret-binding-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-config-secret-ref-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/run-python.sh scripts/check-production-secret-reference-contract.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```

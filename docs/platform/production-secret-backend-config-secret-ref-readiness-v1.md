# Production Secret Backend Config / Secret Ref Readiness v1

更新时间：2026-06-19

## 文档目的

本文档用于固定 production secret backend 在进入 resolver 或 provider profile binding 之前，配置层和 secret reference 层必须先满足的 readiness 边界。后续 `Production Secret Backend Provider Profile Secret Binding Readiness v1` 已把 provider/profile binding 单独推进为 `provider_profile_secret_binding_readiness_defined`，但不改变本文档的配置注入停止线。

本专题只定义 `config-secret-ref-readiness` 的输入、脱敏输出、失败映射和停止线；不修改 platform runtime，不实现 secret resolver，不读取 secret value，不调用云 secret 服务，不启用 production secret backend，也不启用 workflow saved draft repository mode。

## 当前状态

状态：`config_secret_ref_readiness_defined`

本批承接：

- `config-secret-boundary`
- `production-secret-backend-contract`
- `production-secret-backend-implementation-readiness`
- `secret-ref-schema-and-fixtures`
- `production-secret-reference-basic`

结论：

- `production_secret_backend` 仍为 `not_satisfied`。
- `resolver_implementation_status` 仍为 `not_started`。
- `production-secret-reference-basic` 仍为 reference-only manifest。
- `RADISHMIND_SECRET_SOURCE` 仍只表示外部 secret 来源要求，不是 secret backend。
- `RADISHMIND_PLATFORM_API_KEY` 仍只允许作为 developer env override，不能作为生产 fallback。

## 配置注入边界

后续如果进入 runtime implementation，配置层只能消费 secret reference 的存在性和绑定状态，不得读取或解析 secret value。

允许进入配置 / 诊断 summary 的字段只有：

- `credential_state`
- `secret_backend_configured`
- `secret_ref_present`
- `missing_secret_refs`
- `field_sources`

当前 `config.LoadFromEnv`、`config-summary`、`config-check`、`diagnostics` 和 provider inventory 仍保持现有脱敏行为；本批不添加新 runtime 字段，只固定这些字段未来打开时的语义和失败边界。

## Secret Reference 绑定边界

`contracts/production-secret-reference.schema.json` 与 `scripts/checks/fixtures/production-secret-reference-basic.json` 仍是当前唯一 committed secret reference 契约。它只允许保存：

- `environment`
- `provider`
- `provider_profile`
- `secret_ref`
- `required_fields`
- `sanitized_fields`

`secret_ref` 必须以 `ref:` 开头，不能是 secret value、HTTP URL、云 provider SDK 参数、本机绝对路径或真实 credential。

## Failure Mapping

后续实现必须使用 fail-closed 失败语义：

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `secret_reference_manifest_missing` | `configuration` | 生产路径需要 secret reference manifest，但配置未提供 |
| `secret_reference_manifest_invalid` | `configuration` | manifest 不符合 schema 或包含禁止字段 |
| `secret_ref_missing` | `configuration` | provider/profile 需要 credential，但缺少 `secret_ref` |
| `secret_backend_disabled` | `configuration` | secret reference 存在，但 production secret backend 未启用 |
| `resolver_invocation_forbidden` | `configuration` | 当前切片或 readiness 检查中出现 resolver 调用企图 |

这些失败不能 fallback 到 `mock` provider、developer env credential、local-smoke profile、fixture credential 或 committed secret。

## No Fallback

- 不从 `RADISHMIND_PLATFORM_API_KEY` 回退生产 secret。
- 不从 `.env.example`、fixture、task card、README 或 committed config 中读取 secret value。
- 不把 `RADISHMIND_SECRET_SOURCE` 写成 backend ready。
- 不用 `mock`、`local-smoke`、demo profile 或 fake resolver 代替 production secret backend。
- 不把 secret reference present 写成 credential resolved。

## No Side Effects

本批 checker 只读取 committed 文档、schema、fixture 和 `check-repo.py` 注册顺序，不读取真实环境变量、不连接网络、不调用云 secret 服务、不创建文件、不修改 runtime 状态。

## Artifact Guard

本批不得新增或启用以下 artifact：

- secret resolver runtime
- fake resolver
- cloud secret SDK
- secret value fixture
- production credential file
- database connection provider
- DB driver / DSN parser
- workflow saved draft repository mode runtime
- SQL migration runner
- schema marker reader / writer

## 后续推进

当前 readiness 只解除 `production-secret-backend-implementation-readiness` 中的 `config-injection-point` 阻塞。provider profile binding 已由 `production-secret-backend-provider-profile-secret-binding-readiness-v1` 单独固定；本文档仍不解除 disabled resolver interface、operator runbook、rotation / audit policy 和 production ready 阻塞。

下一批如继续 production secret backend，应在以下方向中选择一个独立任务：

1. `secret-resolver-interface-disabled`
2. `operator-runbook-and-negative-gates`

每个方向都必须继续保持 no raw secret、no cloud call、no fallback 和 no production ready 声明。

## 验证

本专题由以下证据固定：

- `docs/task-cards/production-secret-backend-config-secret-ref-readiness-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-config-secret-ref-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-config-secret-ref-readiness-v1.py`
- `scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json`
- `scripts/checks/fixtures/production-secret-reference-basic.json`
- `scripts/check-production-secret-reference-contract.py`

建议验证顺序：

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-config-secret-ref-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/run-python.sh scripts/check-production-secret-reference-contract.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```

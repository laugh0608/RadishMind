# Production Secret Reference 契约

更新时间：2026-05-26

## 文档目的

本文档说明 `production-secret-reference.schema.json` 的长期契约语义。它用于在真实 production secret backend 尚未实现前，先冻结 provider profile 与 secret reference 的最小治理边界，避免把真实 credential、provider URL 或云 secret 服务细节写入仓库。

该契约是 `Production Ops Hardening v1` 的 reference-only 切片，不是生产 secret 服务实现，也不是 production ready 声明。

## 真相源

- Schema：`contracts/production-secret-reference.schema.json`
- Fixture：`scripts/checks/fixtures/production-secret-reference-basic.json`
- 检查入口：`scripts/check-production-secret-reference-contract.py`
- 仓库基线：`pwsh ./scripts/check-repo.ps1 -Fast`

## Manifest 结构

`production_secret_reference_manifest` 用于声明一组 provider profile 应绑定到哪些 secret reference。当前固定字段如下：

- `schema_version`：当前固定为 `1`。
- `kind`：固定为 `production_secret_reference_manifest`。
- `scope`：固定为 `secret_reference_only`。
- `bindings`：按环境、provider 与 provider profile 列出 secret reference。
- `policy`：声明禁止保存 secret value、禁止 resolver、禁止云调用和禁止 production ready 声明。

每个 `binding` 当前只允许表达：

- `environment`：`test` 或 `production`。
- `provider`：provider 稳定短键。
- `provider_profile`：provider profile 稳定短键。
- `secret_ref`：以 `ref:` 开头的引用名，例如 `ref:radishmind/production/openai-compatible/default-api-key`。
- `credential_requirement`：当前 fixture 用 `required` 固定 provider credential 必需状态。
- `secret_ref_status`：只表达 reference 是否存在，不表达 secret value 是否可读。
- `required_fields`：消费方必须保留的结构字段声明。
- `sanitized_fields`：允许出现在只读诊断和 readiness 输出中的脱敏字段声明。

## 允许输出

该契约允许后续配置、readiness 和诊断层输出以下脱敏状态：

- `credential_state`
- `secret_backend_configured`
- `secret_ref_present`
- `missing_secret_refs`
- `field_sources`

这些字段只说明配置引用和治理状态，不得包含真实 secret 内容、provider raw URL、HTTP header、cookie 或 bearer token。

## 禁止内容

`policy.forbidden_fields` 固定禁止以下字段进入 schema fixture、committed manifest 和只读诊断输出：

- `api_key`
- `token`
- `cookie`
- `authorization`
- `credential`
- `secret_value`
- `provider_base_url`

`secret_ref` 必须是引用，不得写成真实密钥、HTTP URL、云 provider SDK 参数或本机绝对路径。`deploy/.env.example` 只能提供非密钥配置样例，真实 `.env` 不提交。

## 当前停止线

- 不实现真实云 secret 服务。
- 不实现 secret resolver。
- 不保存 secret value。
- 不读取、写入或提交真实 secret。
- 不调用云 secret API。
- 不声明 production secret backend ready。
- 不声明 `production_secret_backend_ready=true`。
- 不把 `RADISHMIND_SECRET_SOURCE` 或 `deploy/.env.example` 写成 secret backend。
- 不把 reference fixture 写成 provider credential readiness。
- 不把该契约写成 production ready。

## 与部署和服务的关系

部署层可以通过 `RADISHMIND_SECRET_SOURCE` 表达未来 secret 来源类型，但当前只作为治理引用，不代表 backend 已接通。`services/platform/` 的只读 readiness 或 diagnostics 后续只能消费 `sanitized_fields` 中的字段，不能返回 secret value、provider raw URL 或 resolver 结果。

真实 production secret backend 仍需要后续独立任务补齐 adapter、resolver、rotation policy、audit store、provider health policy 和生产前复核记录。

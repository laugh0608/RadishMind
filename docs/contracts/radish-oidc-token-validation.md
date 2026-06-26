# Radish OIDC Token Validation 契约

更新时间：2026-06-25

## 契约定位

`contracts/radish-oidc-token-validation.schema.json` 是 Saved Workflow Draft future Radish OIDC token validation 的 verified token context schema。它描述的是“真实 token 已由未来 validator 校验完成之后，供 RadishMind 下游消费的脱敏投影”，不是 raw token、middleware input、Radish membership source record 或数据库状态。

该契约用于让后续 auth middleware、membership adapter、repository actor context handoff 和 negative auth smoke 在同一字段边界上推进。它只固定可传递给 RadishMind 的 reference / grant / audit metadata，不定义 OIDC discovery、JWKS fetch、JWT signature verification、membership query、cache、repository mode 或 production API。

## 文件入口

- Schema：`contracts/radish-oidc-token-validation.schema.json`
- 正例 fixture：`scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-positive-v1.json`
- required field 负例：`scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-missing-required-negative-v1.json`
- forbidden raw material 负例：`scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-forbidden-field-negative-v1.json`
- additional properties 负例：`scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-additional-properties-negative-v1.json`
- Schema checker：`scripts/checks/control_plane/check-workflow-saved-draft-token-validation-schema-artifact-implementation-v1.py`

## 字段分组

身份与归属 reference：

- `issuer_ref`
- `subject_ref`
- `tenant_ref`
- `owner_subject_ref`
- `key_id_ref`

授权与范围：

- `audience_refs`
- `scope_grants`
- `workspace_binding_refs`
- `application_scope_refs`

校验时间与策略：

- `algorithm`
- `issued_at`
- `expires_at`
- `auth_time`
- `policy_version`

追踪与审计：

- `request_id`
- `audit_ref`

根对象使用 `additionalProperties=false`。后续若需要新增字段，应先确认它是已验证 token context 的脱敏投影，再同步更新 schema、fixtures、checker、相关专题说明和消费方文档。

## 禁止字段

schema 显式禁止以下 raw-material 或敏感字段：

- `raw_token`
- `authorization_header`
- `cookie`
- `client_secret`
- `refresh_token`
- `authorization_code`
- `jwks_raw_dump`
- `raw_claim_dump`
- `membership_raw_record`
- `database_detail`
- `provider_error_detail`
- `secret_value`

这些字段即使作为 debug 信息、provider error 或 membership 旁路材料出现，也不得进入 committed fixture、runtime envelope、repository actor context 或 UI evidence。

## 消费规则

- 消费方只能读取 schema 中允许的 reference / grant / timestamp / audit 字段。
- `issuer_ref`、`subject_ref`、`tenant_ref`、workspace / application refs 和 `audit_ref` 应作为后续 actor context 与审计链路的稳定引用，不得反查或拼接 raw credential。
- `scope_grants` 只表示已验证后的授权授予摘要，不代表 RadishMind 可以绕过 membership adapter 或 policy evaluator。
- `algorithm`、`key_id_ref` 和时间戳只描述已完成校验的结果，不触发 JWKS refresh、token revalidation 或 provider network call。
- 任何缺字段、额外字段或 forbidden field 都应 fail closed；不能回退到 dev fake auth、sample user、memory dev store、test-only fake resolver、developer env plaintext 或 committed credential。

## 验证

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-token-validation-schema-artifact-implementation-v1.py
```

仓库级快速检查也会覆盖该 checker：

```bash
./scripts/check-repo.sh --fast
```

## 非目标

- 不实现 OIDC middleware、token validator、auth middleware、membership adapter、negative auth smoke runtime、repository mode、database runtime 或 production API。
- 不 fetch issuer discovery，不下载 JWKS，不校验真实 token，不查询 membership，不创建 membership cache。
- 不连接数据库，不运行 SQL，不读写 schema marker，不解析 secret。
- 不把 schema artifact 写成登录态可用、repository actor context 可用、durable store 可用或 production ready。

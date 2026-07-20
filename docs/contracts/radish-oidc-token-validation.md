# Radish OIDC Token Validation 契约

更新时间：2026-07-12

## 契约定位

`contracts/radish-oidc-token-validation.schema.json` 是跨 Saved Workflow Draft / workspace membership 的 verified token context schema。它描述的是“真实 token 已由 validator 校验完成之后，可供后续 workspace actor context 消费的脱敏投影”，不是 raw token、middleware input、Radish membership source record 或数据库状态。

该契约用于让后续 workspace auth middleware、membership adapter、repository actor context handoff 和 negative auth smoke 在同一字段边界上推进。它只固定可传递给 workspace consumer 的 reference / grant / audit metadata，不定义 membership query、repository mode 或 production API。

## 与当前 OIDC runtime 的关系

`radish_oidc_integration_test` 已实现受控 discovery、JWKS fetch/cache/rotation 和 JWT validation，但当前只为 Tenant Summary 与 Audit 投影内部 `VerifiedControlPlaneIdentity`。该 identity 只包含 route authorization 所需的脱敏 subject、tenant、permission、issuer / mapping reference 和 request / audit metadata；raw token 和 raw claims 在进入 handler 前被移除。完整运行语义见 [Control Plane 鉴权与只读运行时契约](control-plane-auth-read-runtime.md)。

当前 Admin OIDC runtime 不把 `contracts/radish-oidc-token-validation.schema.json` 当作 HTTP response 或 repository record，也不据此开放 Saved Workflow Draft、Applications、API Keys、Quota、Workflow Definitions 或 Runs。上述 workspace operation 仍缺正式 membership owner，在 OIDC integration 模式下返回 `workspace_membership_unavailable`。

未来如果 workspace membership adapter 消费本 schema，必须由独立功能设计明确 `workspace_binding_refs`、`application_scope_refs` 和 `scope_grants` 的来源与版本映射；不能把当前 Admin permission projection 自动提升为 workspace membership。

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

- 本 schema 本身不实现 OIDC middleware、token validator、auth middleware、membership adapter、negative auth smoke runtime、repository mode、database runtime 或 production API。
- 本 schema checker 不 fetch issuer discovery、不下载 JWKS、不校验 token、不查询 membership、不创建 membership cache；当前 Admin OIDC runtime 的网络与验证策略由 Platform runtime 配置和测试单独约束。
- 不连接数据库，不运行 SQL，不读写 schema marker，不解析 secret。
- 不把 schema artifact 写成登录态可用、repository actor context 可用、durable store 可用或 production ready。

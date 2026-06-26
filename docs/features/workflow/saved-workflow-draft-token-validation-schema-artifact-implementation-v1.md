# Saved Workflow Draft Token Validation Schema Artifact Implementation v1

更新时间：2026-06-25

## 专题定位

`Saved Workflow Draft Token Validation Schema Artifact Implementation v1` 承接上一轮 `draft_token_validation_schema_implementation_task_card_defined`，在独立批次物化 token validation schema artifact 和离线 schema validation fixtures。

结论：状态为 `draft_token_validation_schema_artifact_implemented`。本批创建 `contracts/radish-oidc-token-validation.schema.json`、positive verified token context fixture、missing required field negative fixture、forbidden raw-material field negative fixture、additionalProperties negative fixture 和 schema validation checker；不实现 OIDC middleware、token validator、auth middleware、membership adapter、negative auth smoke runtime、repository mode runtime、database runtime 或 production API。

## 输入证据

- `workflow-saved-draft-token-validation-schema-implementation-v1` 已固定 schema artifact implementation 的字段、forbidden raw-material fields、fixture 和 checker 要求。
- `workflow-saved-draft-token-validation-schema-task-card-readiness-v1` 已确认 schema implementation task card 可作为本批输入。
- `workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1` 已固定 auth middleware / membership / negative smoke runtime 仍 blocked。
- `radish-oidc-token-membership-upstream-evidence-refresh-v1` 已固定上游 issuer evidence、JWKS pin / refresh policy、client registration evidence 和 membership source ownership，但本批不 fetch discovery / JWKS。
- `draft_production_auth_runtime_bridge_implemented` 只提供 verified auth context 到 repository actor context 的投影，不解析真实 token。

## 本批打开范围

- 创建 `contracts/radish-oidc-token-validation.schema.json`。
- 创建四个 schema validation fixtures：
  - `scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-positive-v1.json`
  - `scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-missing-required-negative-v1.json`
  - `scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-forbidden-field-negative-v1.json`
  - `scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-additional-properties-negative-v1.json`
- 创建 `scripts/checks/control_plane/check-workflow-saved-draft-token-validation-schema-artifact-implementation-v1.py`。
- 将 checker 接入 `./scripts/check-repo.sh --fast`。
- 同步 workflow 入口、features 入口、Saved Workflow Draft v1、Workflow / Agent Runtime、current focus、integration contracts、task card index、scripts README 和周志。

## Schema 边界

schema 只描述已验证 token context 的脱敏投影，根对象 `additionalProperties=false`。

required fields 固定为：

- `issuer_ref`
- `subject_ref`
- `tenant_ref`
- `audience_refs`
- `scope_grants`
- `workspace_binding_refs`
- `application_scope_refs`
- `owner_subject_ref`
- `key_id_ref`
- `algorithm`
- `issued_at`
- `expires_at`
- `auth_time`
- `policy_version`
- `request_id`
- `audit_ref`

schema 必须显式禁止：

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

禁止策略为 `additionalProperties=false` 加 `allOf/not/required` 显式 forbidden field guards。schema 不描述 raw token、raw claims、middleware input、membership source record、database state、secret value 或 provider raw diagnostics。

## 验证覆盖

- 正例 verified token context 必须通过 schema validation。
- 缺少 required field 的负例必须失败。
- 带 forbidden raw-material field 的负例必须失败。
- 带 additional properties 的负例必须失败。
- checker 还会枚举每个 required field 的缺失情况和每个 forbidden field 的注入情况，避免 fixture 只覆盖单个样例。

## No Fallback / No Side Effects

- 不允许 schema artifact 被解释为 token validation runtime ready、auth middleware ready、membership adapter ready、repository mode ready、database ready、production API ready 或 production ready。
- 不允许缺少 production auth runtime 时回退 dev fake auth、memory dev、sample、fixture、test-only fake resolver、fake query executor、developer env plaintext 或 committed credential。
- checker 只读取 committed schema、fixture、文档和 `check-repo.py` 注册顺序；不启动服务、不 fetch issuer discovery、不下载 JWKS、不校验真实 token、不查询 membership、不创建 cache、不连接数据库、不运行 SQL、不读写 schema marker、不解析 secret、不调用 production API、不执行 workflow。

side effect counters 必须保持：

- `issuer_network_call_count=0`
- `jwks_fetch_count=0`
- `token_validation_call_count=0`
- `auth_middleware_invocation_count=0`
- `membership_query_count=0`
- `membership_cache_create_count=0`
- `repository_mode_enablement_count=0`
- `database_connection_count=0`
- `sql_execution_count=0`
- `secret_resolver_call_count=0`
- `production_api_call_count=0`
- `executor_call_count=0`
- `confirmation_call_count=0`
- `business_writeback_count=0`
- `replay_call_count=0`

## 后续推进

下一步若继续 durable store 上游，应在 auth middleware / membership adapter task card entry readiness、negative auth smoke runtime readiness、repository mode runtime blocker consolidation 或 production resolver runtime blocker consolidation 中选择一个独立方向。不能把本批 schema artifact 直接解释为 token validation、membership、repository mode、database 或 production API 可用。

## 验证

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-token-validation-schema-artifact-implementation-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-token-validation-schema-implementation-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-token-validation-schema-task-card-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-radish-oidc-token-membership-upstream-evidence-refresh-v1.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不实现 OIDC middleware、token validator、auth middleware、membership adapter、membership cache、negative auth smoke runtime、login / logout route、session cookie、repository mode runtime、database runtime、production API、executor、confirmation、writeback 或 replay。
- 不 fetch issuer discovery，不下载 JWKS，不校验真实 token，不查询 membership，不连接数据库，不运行 SQL，不读写 schema marker，不解析 secret。
- 不把 `draft_token_validation_schema_artifact_implemented` 解释为 auth ready、repository mode ready、durable persistence ready、production API ready 或 production ready。

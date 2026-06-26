# Saved Workflow Draft Token Validation Schema Implementation v1

更新时间：2026-06-25

## 专题定位

`Saved Workflow Draft Token Validation Schema Implementation v1` 承接上一轮 `draft_token_validation_schema_task_card_readiness_defined`，把 token validation schema 的正式实施批次先收束为可检查任务卡。它的目标是固定后续 schema artifact 必须满足的字段边界、JSON Schema 校验、负例夹具和 artifact guard，而不是在本批创建 schema 文件或运行时。

结论：状态为 `draft_token_validation_schema_implementation_task_card_defined`。本批只创建 schema implementation 静态任务卡、fixture、checker 和文档入口同步；不创建 `contracts/radish-oidc-token-validation.schema.json`，不实现 OIDC middleware、token validator、auth middleware、membership adapter、negative auth smoke runtime、repository mode runtime、database runtime 或 production API。

## 输入证据

- `workflow-saved-draft-token-validation-schema-task-card-readiness-v1` 已确认下一步可以创建 token validation schema implementation task card。
- `workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1` 已固定 schema、middleware、membership、negative smoke 与 repository actor context handoff 的静态边界。
- `radish-oidc-token-membership-upstream-evidence-refresh-v1` 已固定 reviewed issuer evidence、JWKS pin / refresh policy、client registration evidence、claim mapping、auth middleware ownership、membership source / cache policy 和 negative auth smoke matrix。
- `draft_production_auth_runtime_bridge_implemented` 已提供 verified auth context + workspace binding 到 repository actor context 的 runtime bridge，但不解析 token、不查询 membership。
- `draft_production_auth_readiness_defined` 已定义 issuer、claim mapping、workspace binding、scope projection 和 fail-closed 输出前置。

## 本批打开范围

本批只允许打开以下静态实施准备：

- 创建 `docs/task-cards/workflow-saved-draft-token-validation-schema-implementation-v1-plan.md`。
- 创建 `scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-implementation-v1.json`。
- 创建 `scripts/checks/control_plane/check-workflow-saved-draft-token-validation-schema-implementation-v1.py`。
- 将 checker 接入 `./scripts/check-repo.sh --fast` 的聚合入口。
- 更新 workflow 入口、features 入口、Saved Workflow Draft v1、Workflow / Agent Runtime、current focus、integration contracts、task card index、scripts README 和周志。

本批不创建 schema artifact。后续若进入 schema artifact implementation，必须由独立批次创建 `contracts/radish-oidc-token-validation.schema.json`、正例 / 负例 fixture 和 schema validation checker。

## Schema Artifact Implementation 要求

后续 schema artifact implementation 批次必须满足：

- schema 文件路径固定为 `contracts/radish-oidc-token-validation.schema.json`。
- schema 只描述已验证 token context 的脱敏投影，默认 `additionalProperties=false`。
- required fields 只能来自 `issuer_ref`、`subject_ref`、`tenant_ref`、`audience_refs`、`scope_grants`、`workspace_binding_refs`、`application_scope_refs`、`owner_subject_ref`、`key_id_ref`、`algorithm`、`issued_at`、`expires_at`、`auth_time`、`policy_version`、`request_id` 和 `audit_ref`。
- 必须提供 positive verified token context fixture、missing required field negative fixture、forbidden raw-material field negative fixture 和 additional properties negative fixture。
- sanitized failure envelope 必须只包含 failure code、sanitized diagnostic、request id、audit ref 和 policy version 等脱敏字段。
- consumer handoff 只面向 production auth runtime bridge 的 verified context，不把 middleware、membership adapter 或 repository mode runtime 合入 schema artifact 批次。

## Forbidden Fields

schema、fixture、diagnostics 和 future schema validation 输出不得包含：

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

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `token_validation_schema_implementation_task_card_missing` | `task_card` | schema implementation 任务卡不存在或缺少 scope |
| `token_validation_schema_implementation_scope_missing` | `schema_boundary` | 任务卡未固定 schema artifact path、schema version 或 output shape |
| `token_validation_schema_implementation_field_boundary_missing` | `schema_boundary` | 任务卡未列出 required fields 或字段分组 |
| `token_validation_schema_implementation_forbidden_field_missing` | `schema_boundary` | 任务卡未显式禁止 raw token / claim / membership / secret / database detail |
| `token_validation_schema_implementation_validation_plan_missing` | `validation` | 任务卡缺少 JSON Schema validation、positive fixture 或 negative fixture 计划 |
| `token_validation_schema_implementation_schema_file_created_in_task_card` | `artifact_guard` | 本批任务卡阶段直接创建 schema artifact |
| `token_validation_schema_implementation_runtime_scope_overreach` | `implementation_boundary` | 本批混入 validator、middleware、membership、repository mode、database runtime 或 production API |
| `token_validation_schema_implementation_dev_auth_fallback` | `no_fallback` | 任务卡允许 auth 缺失时回退 dev fake auth、sample、fixture、memory dev 或 test-only fake resolver |

所有失败必须 fail closed，不返回 raw token、authorization header、cookie、完整 claim、JWKS dump、membership raw record、secret、DSN、provider detail、database detail 或草案主体。

## No Fallback / No Side Effects

- 不允许 schema implementation task card 被解释为 token validation schema ready、token validation ready、auth middleware ready、membership adapter ready、repository mode ready、database ready、production API ready 或 production ready。
- 不允许缺少 production auth runtime 时回退 dev fake auth、memory dev、sample、fixture、test-only fake resolver、fake query executor、developer env plaintext 或 committed credential。
- checker 只读取 committed 文档、fixture、源码和 `check-repo.py` 注册顺序；不启动服务、不 fetch issuer discovery、不下载 JWKS、不校验 token、不创建 schema 文件、不运行 JSON Schema validation runtime、不查询 membership、不创建 cache、不连接数据库、不运行 SQL、不读写 schema marker、不解析 secret、不调用 production API、不执行 workflow。

side effect counters 必须保持：

- `issuer_network_call_count=0`
- `jwks_fetch_count=0`
- `token_validation_call_count=0`
- `schema_file_write_count=0`
- `schema_validation_call_count=0`
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

下一步若继续 durable store 上游，可以进入独立的 token validation schema artifact implementation 批次，创建 `contracts/radish-oidc-token-validation.schema.json`、schema positive / negative fixtures 和 schema validation checker。auth middleware、membership adapter、negative auth smoke runtime、repository mode runtime、production resolver runtime blocker consolidation 和 public production API 仍必须作为独立方向评审。

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-token-validation-schema-implementation-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-token-validation-schema-task-card-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-radish-oidc-token-membership-upstream-evidence-refresh-v1.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 `contracts/radish-oidc-token-validation.schema.json`。
- 不创建 schema positive / negative validation fixture。
- 不创建 OIDC middleware、token validator、auth middleware、membership adapter、membership cache、negative auth smoke runtime、login / logout route、session cookie、repository mode runtime、database runtime、production API、executor、confirmation、writeback 或 replay。
- 不 fetch issuer discovery，不下载 JWKS，不校验 token，不查询 membership，不连接数据库，不运行 SQL，不读写 schema marker，不解析 secret。
- 不把 `draft_token_validation_schema_implementation_task_card_defined` 解释为 auth ready、repository mode ready、durable persistence ready、production API ready 或 production ready。

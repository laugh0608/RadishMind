# Saved Workflow Draft Token Validation Schema Task Card Readiness v1

更新时间：2026-06-25

## 专题定位

`Saved Workflow Draft Token Validation Schema Task Card Readiness v1` 承接 Radish OIDC upstream evidence refresh、Saved Workflow Draft production auth runtime bridge，以及上一轮 token validation schema / auth middleware runtime entry review，用于判断下一步是否可以创建 token validation schema implementation task card。

结论：状态为 `draft_token_validation_schema_task_card_readiness_defined`。task card decision 为 `token_validation_schema_task_card_ready_for_next_task`。这只表示下一批可以创建 token validation schema implementation task card；本批不创建 `contracts/radish-oidc-token-validation.schema.json`，不实现 OIDC middleware、token validator、auth middleware、membership adapter、negative auth smoke runtime、repository mode runtime、database runtime 或 production API。

## 输入证据

- `draft_token_validation_auth_middleware_runtime_entry_review_defined` 已复核 token validation schema、auth middleware ownership、membership adapter、negative auth smoke、repository actor context handoff、failure mapping、no fallback、no side effects 和 artifact guard。
- `radish_oidc_token_membership_upstream_evidence_refresh_defined` 已固定 reviewed issuer evidence、JWKS pin / refresh policy、client registration evidence、claim mapping、auth middleware ownership、membership source / cache policy 和 negative auth smoke matrix。
- `draft_production_auth_runtime_bridge_implemented` 已提供 verified auth context + workspace binding 到 repository actor context 的 runtime bridge，但不解析 token、不查询 membership。
- `draft_production_auth_readiness_defined` 已定义 issuer、claim mapping、workspace binding、scope projection 和 fail-closed 输出前置。

## Readiness Decision

| gate | 本次结论 | 说明 |
| --- | --- | --- |
| prior runtime entry review | `satisfied_static_entry_review` | 已确认 schema、middleware、membership、negative smoke 与 actor context handoff 的静态边界 |
| upstream issuer evidence | `satisfied_reviewed_evidence` | 只消费 reviewed evidence，不 fetch discovery / JWKS |
| verified token output contract | `satisfied_for_schema_task_card` | required fields、forbidden fields、failure envelope 和 audit fields 已稳定到可写任务卡 |
| schema implementation task card | `ready_for_next_task` | 下一批可创建 schema implementation task card |
| schema artifact | `not_created` | 本批不得创建 schema 文件 |
| token validator runtime | `not_opened` | validator、signature check、issuer refresh 与 claim validation 仍不实现 |
| auth middleware runtime | `blocked` | route binding、public bearer token、session、login / logout 仍不打开 |
| membership adapter runtime | `blocked` | tenant / workspace / application / owner membership lookup 仍不打开 |
| negative auth smoke runtime | `blocked` | runtime smoke fixture、runner、HTTP route smoke 仍不创建 |
| repository mode runtime | `blocked` | schema task card readiness 不启用 repository store mode |

## 下一张任务卡必须覆盖

若下一批创建 token validation schema implementation task card，必须只覆盖以下内容：

- `contracts/radish-oidc-token-validation.schema.json` 的 JSON Schema scope、字段 allowlist、forbidden raw-material fields 和 schema version。
- verified token output shape：`issuer_ref`、`subject_ref`、`tenant_ref`、`audience_refs`、`scope_grants`、`workspace_binding_refs`、`application_scope_refs`、`owner_subject_ref`、`key_id_ref`、`algorithm`、`issued_at`、`expires_at`、`auth_time`、`policy_version`、`request_id` 和 `audit_ref`。
- sanitized failure envelope alignment：schema 缺失、required claim 缺失、time / issuer / audience / scope 不可信时必须 fail closed，不返回 raw token、raw claim 或 provider detail。
- consumer handoff：schema 只描述已验证 token context 的脱敏输出，后续 auth middleware 和 membership adapter 必须仍由独立任务卡打开。
- artifact guard：schema implementation task card 不得顺带创建 middleware、validator runtime、membership adapter、negative auth smoke runtime、repository mode runtime、database runtime 或 production API。

下一张任务卡不得包含：

- OIDC discovery / JWKS fetch runtime
- token signature validation runtime
- auth middleware route binding
- login / logout / session cookie
- membership adapter 或 membership cache
- negative auth smoke runtime
- repository mode runtime
- database connection、SQL、schema marker runtime 或 secret resolver runtime
- production API、executor、confirmation、writeback 或 replay

## Schema Field Boundary

允许字段只描述已验证上下文的脱敏投影，不保存任何 credential material。

required field groups：

- identity：`issuer_ref`、`subject_ref`、`tenant_ref`、`owner_subject_ref`
- audience / scope：`audience_refs`、`scope_grants`、`workspace_binding_refs`、`application_scope_refs`
- token metadata：`key_id_ref`、`algorithm`、`issued_at`、`expires_at`、`auth_time`
- audit：`policy_version`、`request_id`、`audit_ref`

forbidden fields：

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
| `token_validation_schema_task_card_contract_missing` | `task_card_entry` | 下一张任务卡缺少 schema scope、required fields 或 forbidden fields |
| `token_validation_schema_task_card_forbidden_field_missing` | `schema_boundary` | 下一张任务卡未显式禁止 raw token / claim / membership / secret / database detail |
| `token_validation_schema_task_card_runtime_scope_overreach` | `implementation_boundary` | 下一张任务卡混入 validator、middleware、membership、repository mode 或 production API runtime |
| `token_validation_schema_file_created_in_readiness` | `artifact_guard` | 本批 readiness 直接创建 schema 文件 |
| `token_validation_runtime_artifact_created_in_readiness` | `artifact_guard` | 本批 readiness 新增 validator、middleware、adapter、smoke runtime 或 production route |
| `token_validation_schema_task_card_dev_auth_fallback` | `no_fallback` | schema task card 允许 auth 缺失时回退 dev fake auth、sample、fixture 或 memory dev |

所有失败必须使用 sanitized diagnostics，不返回 raw token、authorization header、cookie、完整 claim、JWKS dump、membership raw record、secret、DSN、provider URL、database detail 或草案主体。

## No Fallback / No Side Effects

- 不允许 token validation schema readiness 被解释为 token validation ready、auth middleware ready、membership adapter ready、repository mode ready、database ready、production API ready 或 production ready。
- 不允许缺少 production auth runtime 时回退 dev fake auth、memory dev、sample、fixture、test-only fake resolver、fake query executor、developer env plaintext 或 committed credential。
- checker 只读取 committed 文档、fixture、源码和 `check-repo.py` 注册顺序；不启动服务、不 fetch issuer discovery、不下载 JWKS、不校验 token、不创建 schema、不查询 membership、不创建 cache、不连接数据库、不运行 SQL、不读写 schema marker、不解析 secret、不调用 production API、不执行 workflow。

side effect counters 必须保持：

- `issuer_network_call_count=0`
- `jwks_fetch_count=0`
- `token_validation_call_count=0`
- `schema_file_write_count=0`
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

## 后续方向

本批完成后，下一步可以创建 token validation schema implementation task card，但仍只能围绕 schema artifact、字段边界、JSON Schema 校验和静态 artifact guard 展开。auth middleware / membership adapter task card entry readiness、negative auth smoke runtime、repository mode runtime 和 production resolver runtime blocker consolidation 仍是后续独立方向。

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-token-validation-schema-task-card-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-radish-oidc-token-membership-upstream-evidence-refresh-v1.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 `contracts/radish-oidc-token-validation.schema.json`。
- 不创建 token validation schema implementation task card 之外的 runtime task card。
- 不创建 OIDC middleware、token validator、auth middleware、membership adapter、membership cache、negative auth smoke runtime、login / logout route、session cookie、repository mode runtime、database runtime、production API、executor、confirmation、writeback 或 replay。
- 不 fetch issuer discovery，不下载 JWKS，不校验 token，不查询 membership，不连接数据库，不运行 SQL，不读写 schema marker，不解析 secret。
- 不把 `draft_token_validation_schema_task_card_readiness_defined` 解释为 auth ready、repository mode ready、durable persistence ready、production API ready 或 production ready。

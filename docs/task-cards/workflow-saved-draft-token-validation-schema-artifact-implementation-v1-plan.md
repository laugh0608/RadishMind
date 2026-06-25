# Workflow Saved Draft Token Validation Schema Artifact Implementation v1 任务卡

状态：`draft_token_validation_schema_artifact_implemented`

## 背景

`Saved Workflow Draft Token Validation Schema Implementation v1` 已创建 schema implementation task card，并固定 schema artifact path、字段 allowlist、forbidden raw-material fields、JSON Schema validation 和 positive / negative fixtures。本任务卡负责实际物化 token validation schema artifact 与离线 schema validation checker。

本任务卡不实现 OIDC middleware、token validator、auth middleware、membership adapter、negative auth smoke runtime、repository mode、database runtime 或 production API。

## 输入

- `docs/features/workflow/saved-workflow-draft-token-validation-schema-implementation-v1.md`
- `docs/task-cards/workflow-saved-draft-token-validation-schema-implementation-v1-plan.md`
- `docs/features/workflow/saved-workflow-draft-token-validation-schema-task-card-readiness-v1.md`
- `docs/features/workflow/saved-workflow-draft-token-validation-auth-middleware-runtime-entry-review-v1.md`
- `docs/integrations/radish-oidc-token-membership-upstream-evidence-refresh-v1.md`
- `docs/features/workflow/saved-workflow-draft-production-auth-runtime-v1.md`

## 输出

- 新增 `contracts/radish-oidc-token-validation.schema.json`。
- 新增 positive verified token context fixture。
- 新增 missing required field negative fixture。
- 新增 forbidden raw-material field negative fixture。
- 新增 additionalProperties negative fixture。
- 新增 schema validation checker：`scripts/checks/control_plane/check-workflow-saved-draft-token-validation-schema-artifact-implementation-v1.py`。
- 更新 workflow 入口、features 入口、Saved Workflow Draft v1、Workflow / Agent Runtime、current focus、integration contracts、contracts README、task card index、scripts README 和周志。
- 将 checker 接入 `./scripts/check-repo.sh --fast`。

## 本批实施内容

1. 物化 verified token context schema：
   - `additionalProperties=false`
   - required fields 固定为脱敏 verified context 字段
   - forbidden raw-material fields 通过 `additionalProperties=false` 和显式 `allOf/not/required` guard 拦截

2. 固定 schema validation fixtures：
   - positive verified token context fixture
   - missing required field negative fixture
   - forbidden raw-material field negative fixture
   - additionalProperties negative fixture

3. 固定 schema validation checker：
   - 校验 schema 自身满足 draft 2020-12
   - 校验 positive / negative fixtures 的预期结果
   - 枚举所有 required field 缺失用例
   - 枚举所有 forbidden raw-material field 注入用例
   - 校验旧 guard 已允许本批 schema artifact 作为后续合法产物

## required fields

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

## Forbidden Fields

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

## 验收

- schema 只描述 verified token context 的脱敏投影。
- schema root `additionalProperties=false`。
- positive verified token context fixture 通过 JSON Schema validation。
- missing required field negative fixture、forbidden raw-material field negative fixture 和 additionalProperties negative fixture 均失败。
- schema validation checker 接入 fast baseline。
- 前置 checker 仍能通过，并且不再把本批 schema artifact 当作 runtime artifact。

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

- 不 fetch issuer discovery，不下载 JWKS，不校验真实 token，不查询 membership，不连接数据库，不运行 SQL。
- 不创建 OIDC middleware、token validator、auth middleware、membership adapter、membership cache、negative auth smoke runtime、login / logout route、session cookie、repository mode runtime、database runtime、production API、executor、confirmation、writeback 或 replay。
- 不把 `draft_token_validation_schema_artifact_implemented` 解释为 token validation runtime ready、auth middleware ready、membership adapter ready、repository mode ready、database ready、production API ready 或 production ready。

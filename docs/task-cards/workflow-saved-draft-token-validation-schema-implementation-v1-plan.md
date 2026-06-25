# Workflow Saved Draft Token Validation Schema Implementation v1 任务卡

状态：`draft_token_validation_schema_implementation_task_card_defined`

## 背景

`Saved Workflow Draft Token Validation Schema Task Card Readiness v1` 已确认下一批可以创建 token validation schema implementation task card。本任务卡负责把后续 schema artifact implementation 的输入、输出、字段边界、JSON Schema validation、负例 fixture、failure mapping 和 artifact guard 固定下来。

本任务卡本身不创建 `contracts/radish-oidc-token-validation.schema.json`，不创建 schema positive / negative validation fixture，不实现 OIDC middleware、token validator、auth middleware、membership adapter、negative auth smoke runtime、repository mode、database runtime 或 production API。

## 输入

- `docs/features/workflow/saved-workflow-draft-token-validation-schema-task-card-readiness-v1.md`
- `docs/features/workflow/saved-workflow-draft-token-validation-auth-middleware-runtime-entry-review-v1.md`
- `docs/integrations/radish-oidc-token-membership-upstream-evidence-refresh-v1.md`
- `docs/features/workflow/saved-workflow-draft-production-auth-runtime-v1.md`
- `docs/features/workflow/saved-workflow-draft-production-auth-readiness-v1.md`

## 输出

- 新增 `Saved Workflow Draft Token Validation Schema Implementation v1` 细专题。
- 新增 `workflow-saved-draft-token-validation-schema-implementation-v1` fixture / checker。
- 更新 workflow 入口、features 入口、Saved Workflow Draft v1、Workflow / Agent Runtime、current focus、integration contracts、task card index、scripts README 和周志。
- 将 checker 接入 `./scripts/check-repo.sh --fast`。

## 本批实施内容

1. 创建 token validation schema implementation 静态任务卡：
   - 固定 future schema artifact path
   - 固定 verified token output field allowlist
   - 固定 forbidden raw-material field list
   - 固定 JSON Schema validation / positive fixture / negative fixture 要求
   - 固定 sanitized failure envelope 和 artifact guard

2. 固定后续 schema artifact implementation 验收：
   - `contracts/radish-oidc-token-validation.schema.json`
   - positive verified token context fixture
   - missing required field negative fixture
   - forbidden raw-material field negative fixture
   - additional properties negative fixture
   - schema validation checker

3. 固定停止线：
   - schema artifact 不在本批创建
   - token validator、OIDC middleware、auth middleware、membership adapter 和 negative auth smoke runtime 不在本批创建
   - repository mode runtime、database runtime、production API、executor、confirmation、writeback 和 replay 不在本批创建

## Schema Artifact Implementation 要求

后续 schema artifact implementation 批次必须把以下项目作为验收项：

- schema version 与 `$schema` 明确，schema root 只描述 verified token context。
- `additionalProperties=false`，禁止 raw token、authorization header、cookie、client secret、refresh token、authorization code、JWKS raw dump、raw claim dump、membership raw record、database detail、provider detail 和 secret value。
- required fields 为 `issuer_ref`、`subject_ref`、`tenant_ref`、`audience_refs`、`scope_grants`、`workspace_binding_refs`、`application_scope_refs`、`owner_subject_ref`、`key_id_ref`、`algorithm`、`issued_at`、`expires_at`、`auth_time`、`policy_version`、`request_id` 和 `audit_ref`。
- JSON Schema validation 必须覆盖正例、required field 缺失、forbidden field 出现和 additional properties。
- failure envelope 只能返回 sanitized diagnostic，不暴露 raw credential、claim、membership、provider 或 database detail。
- consumer handoff 只到 production auth runtime bridge 的 verified context，不创建 middleware、membership adapter 或 repository mode success path。

## 验收

- fixture 能证明本任务卡只创建静态实施边界，不创建 schema artifact。
- checker 能检测 schema field boundary、forbidden field boundary、validation plan、failure mapping、no fallback、no side effects 和 artifact guard。
- 聚合检查顺序为 auth middleware runtime entry review、schema task card readiness、schema implementation task card、后续 product surface / control plane checks。
- 前一张 readiness 的 artifact guard 已允许本任务卡作为后续合法产物，但仍不改变 readiness 本批“不创建任务卡”的历史结论。

## 验证

本批完成后运行：

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
- 不创建 schema validation positive / negative fixture。
- 不 fetch issuer discovery，不下载 JWKS，不校验 token，不查询 membership。
- 不创建 OIDC middleware、token validator、auth middleware、membership adapter、membership cache、negative auth smoke runtime、login / logout route、session cookie、repository mode runtime、database runtime、production API、executor、confirmation、writeback 或 replay。

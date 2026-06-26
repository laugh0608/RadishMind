# Saved Workflow Draft Negative Auth Smoke Runtime Readiness v1

更新时间：2026-06-25

## 专题定位

`Saved Workflow Draft Negative Auth Smoke Runtime Readiness v1` 承接 token validation schema artifact implementation、auth middleware / membership adapter task card entry readiness、Radish OIDC upstream evidence refresh 和 production auth runtime bridge，用于固定 future negative auth smoke runtime 的准入矩阵。

结论：状态为 `draft_negative_auth_smoke_runtime_readiness_defined`。本批只定义 negative auth smoke runtime readiness，固定 13 个负向 auth / membership 场景、failure code、脱敏诊断、no fallback、no side effects 和 runtime artifact guard；不创建 negative auth smoke runtime fixture、runtime checker、OIDC middleware、token validator、auth middleware、membership adapter、repository mode runtime、database runtime 或 production API。

entry decision：`negative_auth_smoke_runtime_readiness_defined_before_runtime_artifact`。

readiness boundary：`negative_auth_smoke_runtime_readiness_defined_no_runtime`。

## 输入证据

- `draft_auth_middleware_membership_adapter_task_card_entry_readiness_defined` 已确认 auth middleware / membership adapter implementation task card 仍 blocked pending negative auth smoke runtime readiness。
- `draft_token_validation_schema_artifact_implemented` 已物化 verified token context 的脱敏 schema 投影。
- `draft_token_validation_auth_middleware_runtime_entry_review_defined` 已固定 token validation、auth middleware、membership adapter、negative auth smoke 和 repository actor context handoff 的 runtime entry review。
- `radish_oidc_token_membership_upstream_evidence_refresh_defined` 已固定 negative auth smoke matrix 的上游 case boundary。
- `draft_production_auth_runtime_bridge_implemented` 已提供 verified auth context 到 repository actor context 的桥接，但不解析 token、不查询 membership。

## Readiness Boundary

| 项目 | 结论 | 说明 |
| --- | --- | --- |
| negative auth case matrix | `ready_as_static_runtime_readiness` | 13 个场景已固定为 future runtime smoke 必须覆盖的输入矩阵 |
| failure mapping | `fail_closed_before_repository_query_required` | 每个负向场景必须在 repository query 前 fail closed |
| diagnostics | `sanitized_only` | 只允许 failure code、request / audit ref、policy version 和 case id |
| runtime smoke artifact | `not_allowed_in_this_slice` | future fixture / checker 路径已声明，但本批不创建 |
| implementation task card | `still_blocked_before_runtime_artifact` | 后续仍需独立 runtime smoke artifact / checker 才能评审 auth middleware / membership adapter implementation task card |

## Negative Auth Case Matrix

future smoke runtime 必须覆盖：

- `missing_authorization_header` -> `draft_identity_context_missing`
- `malformed_bearer_header` -> `draft_auth_context_contract_mismatch`
- `unknown_issuer` -> `draft_auth_context_contract_mismatch`
- `jwks_unavailable` -> `draft_auth_context_contract_mismatch`
- `invalid_signature` -> `draft_auth_context_contract_mismatch`
- `expired_token` -> `draft_auth_context_contract_mismatch`
- `missing_required_claim` -> `draft_identity_context_missing`
- `tenant_binding_denied` -> `draft_tenant_binding_missing`
- `workspace_membership_denied` -> `draft_workspace_membership_denied`
- `application_scope_denied` -> `draft_application_scope_denied`
- `owner_scope_denied` -> `draft_owner_scope_denied`
- `scope_grant_missing` -> `draft_scope_grant_missing`
- `membership_source_unavailable` -> `draft_auth_context_contract_mismatch`

future runtime fixture path：`scripts/checks/fixtures/workflow-saved-draft-negative-auth-smoke-runtime-v1.json`。future runtime checker path：`scripts/checks/control_plane/check-workflow-saved-draft-negative-auth-smoke-runtime-v1.py`。这两个路径本批必须保持不存在。

## Future Runtime Requirements

后续若创建 negative auth smoke runtime artifact，必须覆盖：

- 只使用合成、脱敏、静态输入，不使用真实 token、真实 issuer、真实 JWKS、真实 membership 或数据库。
- 每个场景必须先经过 auth / membership boundary，再进入 repository adapter 之前 fail closed。
- failure envelope 只能包含脱敏字段，禁止 raw token、authorization header、cookie、raw claim、membership raw record、provider detail、database detail、secret material 或 DSN。
- 不允许 fallback 到 dev fake auth、memory dev store、sample、fixture truth、test-only fake resolver 或 fake query executor。
- smoke runtime 不得启用 repository mode，不得连接数据库，不得运行 SQL，不得创建 production API consumer。

## 验证

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-negative-auth-smoke-runtime-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-auth-middleware-membership-adapter-task-card-entry-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-token-validation-schema-artifact-implementation-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-radish-oidc-token-membership-upstream-evidence-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-production-auth-runtime-v1.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 negative auth smoke runtime fixture / checker / runner。
- 不创建 auth middleware / membership adapter implementation task card。
- 不实现 OIDC middleware、token validator、auth middleware、membership adapter、membership cache、repository mode runtime、database runtime、production API、executor、confirmation、writeback 或 replay。
- 不 fetch issuer discovery，不下载 JWKS，不校验真实 token，不查询 membership，不连接数据库，不运行 SQL，不读写 schema marker，不解析 secret。
- 不把 `draft_negative_auth_smoke_runtime_readiness_defined` 解释为 auth runtime ready、membership ready、repository mode ready、database ready、production API ready 或 production ready。

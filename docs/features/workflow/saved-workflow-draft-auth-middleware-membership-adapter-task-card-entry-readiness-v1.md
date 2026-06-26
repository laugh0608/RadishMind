# Saved Workflow Draft Auth Middleware / Membership Adapter Task Card Entry Readiness v1

更新时间：2026-06-25

## 专题定位

`Saved Workflow Draft Auth Middleware / Membership Adapter Task Card Entry Readiness v1` 承接 token validation schema artifact implementation、Radish OIDC upstream evidence refresh 和 production auth runtime bridge，用于判断 future auth middleware / membership adapter implementation task card 是否可以创建。

结论：状态为 `draft_auth_middleware_membership_adapter_task_card_entry_readiness_defined`。本批固定 auth middleware、token validator、membership adapter、repository actor context handoff、negative auth smoke readiness 和 failure mapping 的实现任务卡准入要求；实现任务卡仍不创建，原因是 negative auth smoke runtime readiness、runtime route binding、membership source runtime 和 repository mode 依赖仍未完成。本批不实现 OIDC middleware、token validator、auth middleware、membership adapter、negative auth smoke runtime、repository mode runtime、database runtime 或 production API。

entry decision：`auth_middleware_membership_adapter_implementation_task_card_blocked_pending_negative_auth_smoke_readiness`。

## 输入证据

- `draft_token_validation_schema_artifact_implemented` 已物化 `contracts/radish-oidc-token-validation.schema.json` 和 schema validation fixtures。
- `draft_token_validation_auth_middleware_runtime_entry_review_defined` 已固定 auth middleware / membership adapter / negative auth smoke 的 runtime entry review，并确认 runtime task card 仍 blocked。
- `radish_oidc_token_membership_upstream_evidence_refresh_defined` 已固定 reviewed issuer evidence、JWKS pin / refresh policy、client registration evidence、auth middleware ownership、membership source ownership 和 negative auth smoke matrix。
- `draft_production_auth_runtime_bridge_implemented` 已提供 verified auth context + workspace binding 到 repository actor context 的 bridge，但不解析 token、不查询 membership。
- `draft_database_secret_resolver_runtime_dependency_refresh_defined`、`draft_schema_marker_runtime_dependency_refresh_defined` 和 `draft_repository_mode_runtime_boundary_review_defined` 仍确认 repository mode、DB runtime、schema marker runtime 和 secret resolver runtime 未打开。

## Readiness Decision

| 项目 | 结论 | 说明 |
| --- | --- | --- |
| token validation schema artifact | `satisfied_static_schema_artifact` | schema 已存在，可作为 future validator 输出契约 |
| auth middleware ownership | `ready_as_static_owner_contract` | route binding、validator、failure envelope、audit 和 dev fake auth isolation owner 已能进入任务卡要求 |
| membership adapter contract | `ready_as_static_contract` | tenant、workspace、application、owner 和 scope grant 投影边界已能进入任务卡要求 |
| negative auth smoke readiness | `required_before_implementation_task_card` | 负向 auth smoke runtime readiness 尚未定义，不能创建 runtime 实现任务卡 |
| repository actor context handoff | `existing_bridge_reused` | future runtime 必须复用 existing production auth bridge，不新增 token-to-repository 私有路径 |
| implementation task card | `blocked_before_task_card` | 先完成 negative auth smoke runtime readiness 或同等前置评审 |

## Future Task Card Requirements

后续若创建 auth middleware / membership adapter implementation task card，必须覆盖：

- token validator 输出必须符合 `contracts/radish-oidc-token-validation.schema.json`。
- auth middleware 只接收 future validator 的 verified token context，不把 raw token、authorization header、cookie 或 raw claim 写入 repository / audit / diagnostics。
- membership adapter 必须分开处理 tenant binding、workspace membership、application scope、owner scope 和 scope grant projection。
- membership cache 只允许脱敏 reference 与 expiry policy，不保存 raw membership record。
- failure envelope 必须在 repository adapter 前 fail closed，并映射到 Saved Workflow Draft 既有 failure code。
- repository actor context handoff 必须复用 `BuildSavedWorkflowDraftRepositoryActorContextFromProductionAuth`。
- negative auth smoke readiness 必须先定义 missing / malformed auth、unknown issuer、JWKS unavailable、invalid signature、expired token、missing claim、tenant denied、workspace denied、application denied、owner denied、scope missing 和 membership unavailable。
- no fallback / no side effects 必须禁止 dev fake auth、memory dev、fixture、sample、test-only resolver、fake query executor、database、OIDC network call 和 production API fallback。

## Blocked Preconditions

- `negative_auth_smoke_runtime_readiness` 尚未定义。
- auth middleware runtime route binding 尚未定义。
- token validator runtime 尚未创建。
- membership source runtime / adapter 尚未创建。
- repository mode 仍为 fail closed。
- database connection provider、schema marker runtime、secret resolver runtime 和 production resolver runtime 仍 blocked。

## 验证

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-auth-middleware-membership-adapter-task-card-entry-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-token-validation-schema-artifact-implementation-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-radish-oidc-token-membership-upstream-evidence-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-production-auth-runtime-v1.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 auth middleware / membership adapter implementation task card。
- 不实现 OIDC middleware、token validator、auth middleware、membership adapter、membership cache、negative auth smoke runtime、login / logout route、session cookie、repository mode runtime、database runtime、production API、executor、confirmation、writeback 或 replay。
- 不 fetch issuer discovery，不下载 JWKS，不校验真实 token，不查询 membership，不连接数据库，不运行 SQL，不读写 schema marker，不解析 secret。
- 不把 `draft_auth_middleware_membership_adapter_task_card_entry_readiness_defined` 解释为 auth ready、repository mode ready、durable persistence ready、production API ready 或 production ready。

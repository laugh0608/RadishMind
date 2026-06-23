# Radish OIDC Token / Membership Upstream Evidence Refresh v1 任务卡

状态：`radish_oidc_token_membership_upstream_evidence_refresh_defined`

## 任务目标

把 Radish OIDC token validation / membership adapter 进入 runtime 之前缺失的上游证据拆成可检查契约，并确认 runtime task card 仍然 blocked。

本任务只固定 reviewed issuer evidence、JWKS pin / refresh policy、client registration evidence、auth middleware ownership、membership data source ownership、membership cache policy、negative auth smoke matrix、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard。不创建任何 runtime artifact。

## 输入

- `radish-oidc-token-membership-readiness-v1`
- `radish-oidc-token-membership-implementation-entry-review-v1`
- `radish-oidc-client-preconditions`
- `control-plane-read-production-auth-readiness-v1`
- `workflow-saved-draft-production-auth-readiness-v1`
- `workflow-saved-draft-production-auth-runtime-v1`

## 本批输出

- `docs/integrations/radish-oidc-token-membership-upstream-evidence-refresh-v1.md`
- `scripts/checks/fixtures/radish-oidc-token-membership-upstream-evidence-refresh-v1.json`
- `scripts/checks/control_plane/check-radish-oidc-token-membership-upstream-evidence-refresh-v1.py`
- `scripts/check-repo.py` fast baseline 注册
- integrations / platform / workflow / current focus / roadmap / product scope / capability matrix / task card index / scripts README / 周志同步

## 验收要求

- fixture 必须固定状态为 `radish_oidc_token_membership_upstream_evidence_refresh_defined`。
- evidence contract matrix 必须覆盖 reviewed issuer evidence、JWKS pin / refresh policy、client registration evidence、auth middleware ownership、membership source ownership、membership cache policy 和 negative auth smoke matrix。
- checker 必须确认 readiness 与 implementation entry review 证据状态未漂移。
- checker 必须确认 future token / membership runtime artifact 没有提前出现。
- checker 必须确认本批没有 issuer network call、JWKS fetch、token validation、membership query、repository write、production API call 或 executor side effect。
- 新 checker 必须进入 `./scripts/check-repo.sh --fast`。

## 验证

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-radish-oidc-token-membership-upstream-evidence-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-radish-oidc-token-membership-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-radish-oidc-token-membership-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v2.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 OIDC middleware、OIDC client、token validation schema、token validator、membership adapter、login / logout route、session cookie 或 production auth smoke。
- 不 fetch issuer discovery、不下载 JWKS、不校验 token、不查询 membership、不创建 membership cache。
- 不创建 repository mode runtime、production API consumer、database query executor、schema marker runtime、migration runner、SQL、secret resolver、audit store runtime 或 backend runtime。
- 不实现 publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。

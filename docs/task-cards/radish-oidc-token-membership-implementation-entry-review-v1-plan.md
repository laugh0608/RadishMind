# Radish OIDC Token / Membership Implementation Entry Review v1 任务卡

状态：`radish_oidc_token_membership_implementation_entry_review_defined`

## 任务目标

评审 Radish OIDC token validation schema、auth middleware、membership adapter、negative auth smoke 和 production API consumer gate 是否可以进入实现任务。

本任务只固定 entry decision、candidate matrix、blocked conditions、future task card requirements、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard。不创建任何 runtime artifact。

## 输入

- `radish-oidc-token-membership-readiness-v1`
- `radish-oidc-client-preconditions`
- `control-plane-read-production-auth-readiness-v1`
- `workflow-saved-draft-production-auth-readiness-v1`
- `workflow-saved-draft-production-auth-runtime-v1`

## 本批输出

- `docs/integrations/radish-oidc-token-membership-implementation-entry-review-v1.md`
- `scripts/checks/fixtures/radish-oidc-token-membership-implementation-entry-review-v1.json`
- `scripts/checks/control_plane/check-radish-oidc-token-membership-implementation-entry-review-v1.py`
- `scripts/check-repo.py` fast baseline 注册
- integrations / platform / workflow / current focus / task card index / scripts README / 周志同步

## 验收要求

- fixture 必须固定 entry decision 为 `blocked_before_runtime_task_card`。
- candidate matrix 必须覆盖 token validation schema、auth middleware、membership adapter、negative auth smoke 和 production API consumer binding。
- checker 必须确认 readiness 与上游 auth 证据状态未漂移。
- checker 必须确认 future token / membership implementation artifact 没有提前出现。
- 新 checker 必须进入 `./scripts/check-repo.sh --fast`。

## 验证

- `./scripts/run-python.sh scripts/checks/control_plane/check-radish-oidc-token-membership-implementation-entry-review-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-radish-oidc-token-membership-readiness-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-radish-oidc-client-preconditions.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-production-auth-readiness-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-production-auth-runtime-v1.py`
- `./scripts/check-repo.sh --fast`

## 停止线

- 不创建 OIDC middleware、OIDC client、token validation schema、token validator、membership adapter、login / logout route、session cookie 或 production auth smoke。
- 不创建 repository mode runtime、production API consumer、database query executor、schema marker runtime、migration runner、SQL、secret resolver、audit store runtime 或 backend runtime。
- 不实现 publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。

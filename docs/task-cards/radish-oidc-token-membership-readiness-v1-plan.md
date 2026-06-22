# Radish OIDC Token / Membership Readiness v1 任务卡

状态：`radish_oidc_token_membership_readiness_defined`

## 任务目标

固定 RadishMind 后续接入 `Radish` OIDC token validation 与 tenant / workspace / application / owner membership adapter 前必须具备的跨产品面 readiness 证据。

本任务只定义 token validation contract、membership adapter contract、consumer matrix、failure mapping、no fallback、no side effects 和 artifact guard。不创建 OIDC client、auth middleware、token validator、membership adapter、production API consumer、repository mode runtime 或数据库 artifact。

## 输入

- `radish-oidc-client-preconditions`
- `control-plane-read-auth-store-transition-preconditions-v1`
- `control-plane-read-production-auth-readiness-v1`
- `workflow-saved-draft-production-auth-readiness-v1`
- `workflow-saved-draft-production-auth-runtime-v1`

## 本批输出

- `docs/integrations/radish-oidc-token-membership-readiness-v1.md`
- `scripts/checks/fixtures/radish-oidc-token-membership-readiness-v1.json`
- `scripts/checks/control_plane/check-radish-oidc-token-membership-readiness-v1.py`
- `scripts/check-repo.py` fast baseline 注册
- integrations / platform / workflow / current focus / task card index / scripts README / 周志同步

## 验收要求

- fixture 必须覆盖 issuer evidence、token validation gates、required / optional claim mapping、membership input / output、consumer matrix、failure mapping、no fallback 和 no side effects。
- checker 必须确认 upstream auth readiness 状态未漂移。
- checker 必须确认 future token / membership implementation artifact 没有提前出现。
- checker 必须确认本批不声明 Radish OIDC ready、token validation ready、membership adapter ready、repository mode ready、production API ready 或 production ready。
- 新 checker 必须进入 `./scripts/check-repo.sh --fast`。

## 验证

- `./scripts/run-python.sh scripts/checks/control_plane/check-radish-oidc-token-membership-readiness-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-radish-oidc-client-preconditions.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-production-auth-readiness-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-production-auth-readiness-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-production-auth-runtime-v1.py`
- `./scripts/check-repo.sh --fast`

## 停止线

- 不创建 OIDC middleware、OIDC client、token validation schema、token validator、membership adapter、login / logout route、session cookie 或 production auth smoke。
- 不创建 repository mode runtime、production API consumer、database query executor、schema marker runtime、migration runner、SQL、secret resolver、audit store runtime 或 backend runtime。
- 不实现 publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。

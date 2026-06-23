# Radish OIDC Token / Membership Implementation Entry Review v1

更新时间：2026-06-23

## 专题定位

`Radish OIDC Token / Membership Implementation Entry Review v1` 承接 [Radish OIDC Token / Membership Readiness v1](radish-oidc-token-membership-readiness-v1.md)，评审 token validation schema、auth middleware、membership adapter、negative auth smoke 和 production API consumer gate 是否可以进入实现任务。

本专题只做 implementation entry review，不创建 OIDC client、token validation schema、token validator、auth middleware、membership adapter、login / logout route、session cookie、production API consumer、repository mode runtime、数据库连接、workflow executor、confirmation、writeback 或 replay。

状态：`radish_oidc_token_membership_implementation_entry_review_defined`

## 输入事实

- `radish_oidc_token_membership_readiness_defined` 已固定 token validation contract、membership contract、consumer matrix、failure mapping、no fallback 和 no side effects。
- `radish-oidc-client-preconditions` 仍只是 governance boundary，不包含已 review 的 issuer discovery artifact、client registration artifact 或 JWKS pin。
- control plane read 与 saved draft 的 production auth readiness / runtime bridge 仍只消费已验证上下文，不校验 token、不查询 membership。
- [Radish OIDC Token / Membership Upstream Evidence Refresh v1](radish-oidc-token-membership-upstream-evidence-refresh-v1.md) 已固定 `radish_oidc_token_membership_upstream_evidence_refresh_defined`，把 reviewed issuer evidence、JWKS pin / refresh policy、client registration evidence、auth middleware ownership、membership source ownership、membership cache policy 和 negative auth smoke matrix 拆成静态契约；仍不创建 runtime artifact。

## Entry Decision

结论：`blocked_before_runtime_task_card`。

阻塞项：

- reviewed Radish issuer evidence 未落地。
- token validation schema 未落地。
- auth middleware route ownership 和 negative smoke 未定义。
- membership adapter 的 data source、cache、tenant / workspace / application / owner ownership 未定义。
- production API consumer gate 仍未打开。

## Candidate Matrix

| candidate | entry decision | runtime artifact |
| --- | --- | --- |
| token validation schema | blocked before task card | not created |
| auth middleware | blocked before task card | not created |
| membership adapter | blocked before task card | not created |
| negative auth smoke | blocked before task card | not created |
| production API consumer binding | blocked before task card | not created |

## 后续要求

future implementation task card 至少要先消费并复验：

- reviewed issuer discovery evidence、JWKS pin / refresh policy 与 client registration evidence。
- token validation schema 与 failure envelope。
- auth middleware route ownership、dev fake auth isolation 和 no public fake auth rule。
- membership adapter ownership、data source、cache / expiry、tenant / workspace / application / owner checks。
- negative auth smoke matrix：missing token、malformed header、invalid signature、expired token、missing tenant、scope denied、membership denied 和 sanitized diagnostics。

## 验收方式

- 新增 `radish-oidc-token-membership-implementation-entry-review-v1` fixture / checker。
- checker 消费 readiness fixture 和上游 auth readiness fixture，固定 blocked entry decision、candidate matrix、blocked conditions、failure mapping、diagnostic allowlist、no fallback、no side effects 和 artifact guard。
- checker 接入 `./scripts/check-repo.sh --fast`。

## 停止线

- 不创建 OIDC middleware、token validator、membership adapter、login / logout route、session cookie、production API consumer 或 auth runtime smoke。
- 不创建 token validation schema、runtime smoke fixture、runtime checker 或 production route binding。
- 不启用 repository mode，不创建真实 query executor、数据库连接、schema marker runtime、migration runner、SQL、secret resolver、audit store runtime 或 backend runtime。
- 不实现 workflow executor、publish、run、confirmation decision、writeback、replay、resume 或 materialized result reader。

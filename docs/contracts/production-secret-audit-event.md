# Production Secret Audit Event 契约

更新时间：2026-06-28

## 定位

`contracts/production-secret-audit-event.schema.json` 是 future production secret backend audit store runtime 的 metadata-only audit event 契约。它只固定未来 audit event writer 可以接收的字段形状、event kind allowlist、reference-only policy 和禁止字段，不打开 runtime、不写 audit event、不连接 durable store，也不声明 production secret backend ready。

## 字段边界

- `event_version` 固定为 `audit-event-schema-v1`。
- `event_kind` 只允许 contract readiness 已固定的十类 metadata event。
- required fields 覆盖 event identity、environment、provider/profile、backend profile ref、secret ref 状态、operator approval ref、credential handle boundary ref、request/audit refs、policy refs、idempotency key ref 和 fail-closed delivery mode。
- optional fields 只允许 failure code / boundary、sanitized diagnostic、rotation/runbook version、approval window/ticket ref、backend health ref 和 no leakage smoke ref。
- reference 字段只保存 opaque reference，不保存真实 secret、credential payload、operator raw claim、provider raw URL、DSN 或数据库细节。

## 禁止字段

schema 通过 `additionalProperties=false` 和显式 `allOf.not.required` guard 拒绝 secret value、raw secret、password、token、API key、authorization header、cookie、provider raw URL、resolver backend URL、DSN、database hostname、cloud credential、credential payload、full secret ref、full credential handle、full operator/user claim、raw request / response / audit / writer / event payload、schema payload、payload hash、event payload hash 和 secret-derived hash。

## 消费方式

- 正例 fixture：`scripts/checks/fixtures/production-secret-audit-event-positive-v1.json`
- 负例 fixture：missing required、forbidden field、additionalProperties、event kind invalid 四类 fixture
- 程序化校验：`scripts/check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-v1.py`
- 仓库级入口：`scripts/check-repo.py`

## 停止线

本契约不打开 runtime，不创建 audit writer、不创建 audit store、不执行 delivery、不创建 idempotency runtime、不选择 durable backend、不连接 production resolver、不调用云 secret 服务、不启用 DB / repository mode / production API。后续 runtime 仍必须按 durable backend、writer、delivery、idempotency、approval、credential handle、backend health 和 no leakage gate 逐项推进。

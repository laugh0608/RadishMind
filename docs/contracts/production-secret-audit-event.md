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

## 事件种类

`event_kind` 当前只允许以下 metadata-only 事件。它们用于描述 future resolver / audit store 的安全边界检查结果，不代表这些 runtime 已经存在。

| event kind | 语义 |
| --- | --- |
| `secret_resolution_requested` | 记录 secret resolution 请求已进入 future resolver 边界 |
| `secret_resolution_denied` | 记录请求因 policy / approval / scope 等前置未满足而被拒绝 |
| `secret_resolution_failed_closed` | 记录 resolver 或依赖证据不足时按 fail-closed 停止 |
| `credential_handle_boundary_checked` | 记录 credential handle boundary evidence 已被检查 |
| `operator_approval_evidence_checked` | 记录 operator approval evidence 已被检查 |
| `backend_profile_selected` | 记录 backend profile ref 已由静态策略选择或引用 |
| `backend_health_gate_checked` | 记录 backend health gate evidence 已被检查 |
| `no_leakage_gate_checked` | 记录 no leakage smoke / scanner evidence 已被检查 |
| `rotation_policy_checked` | 记录 rotation / runbook policy version 已被检查 |
| `audit_handoff_failed_closed` | 记录 audit handoff 缺少必要前置时按 fail-closed 停止 |

## 字段分组

| 分组 | 字段 | 说明 |
| --- | --- | --- |
| identity | `event_id`、`event_version`、`event_kind`、`occurred_at` | audit event 自身身份与 schema 版本 |
| environment / profile | `environment`、`provider`、`provider_profile`、`backend_profile_ref` | 只保存 provider/profile key 与 backend profile ref，不保存 provider raw URL |
| secret reference | `secret_ref_key_status`、`secret_ref_version_ref` | 只记录 secret ref 是否存在及其版本引用，不保存 secret ref 完整值或 secret value |
| approval / handle | `operator_approval_ref`、`credential_handle_boundary_ref`、`approval_ticket_ref`、`approval_window_ref` | 只保存 approval / handle 证据引用，不保存 raw ticket、operator claim 或 credential payload |
| request / audit refs | `request_id`、`audit_ref`、`idempotency_key_ref` | 只保存请求、审计和幂等 key 的 opaque ref |
| policy refs | `policy_version`、`retention_policy_ref`、`redaction_profile_ref`、`rotation_policy_version`、`runbook_version` | 只保存策略版本或策略 ref，便于 future audit store 复验来源 |
| delivery / failure | `delivery_mode`、`failure_code`、`failure_boundary`、`sanitized_diagnostic` | `delivery_mode` 只能是 `fail_closed`；诊断必须是脱敏短文本 |
| gate refs | `backend_health_ref`、`no_leakage_smoke_ref` | 只保存 health / leakage gate 的证据引用 |

## 未来 writer / consumer 使用规则

- future audit writer 只能把该 schema 视为输入契约：先构造 metadata-only event，再由 schema checker 验证，最后才允许进入后续 writer runtime 评审。
- future consumer 不能从该 event 反推 secret value、cloud credential、provider raw URL、DSN、operator raw claim、raw request payload、raw response payload 或 writer payload。
- `request_id`、`audit_ref`、`backend_profile_ref`、`operator_approval_ref`、`credential_handle_boundary_ref`、`idempotency_key_ref` 等字段都是 opaque reference；它们不是可解引用的生产 secret，也不是 durable store 主键设计。
- positive fixture 只证明 schema 可以接受一条合法 metadata event；负例 fixture 只证明禁止字段、额外字段、缺必填字段和非法 event kind 会被拒绝。fixture 不得被当作 runtime event log。
- 后续若新增 event kind 或字段，必须同步更新 schema、positive / negative fixtures、artifact checker、契约说明和相关 platform 专题，且继续保持 reference-only policy。

## 禁止字段

schema 通过 `additionalProperties=false` 和显式 `allOf.not.required` guard 拒绝 secret value、raw secret、password、token、API key、authorization header、cookie、provider raw URL、resolver backend URL、DSN、database hostname、cloud credential、credential payload、full secret ref、full credential handle、full operator/user claim、raw request / response / audit / writer / event payload、schema payload、payload hash、event payload hash 和 secret-derived hash。

## 消费方式

- 正例 fixture：`scripts/checks/fixtures/production-secret-audit-event-positive-v1.json`
- 负例 fixture：missing required、forbidden field、additionalProperties、event kind invalid 四类 fixture
- 程序化校验：`scripts/check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-v1.py`
- 仓库级入口：`scripts/check-repo.py`

## 复验矩阵

| 验证目标 | fixture / 入口 | 预期 |
| --- | --- | --- |
| 正例 event | `scripts/checks/fixtures/production-secret-audit-event-positive-v1.json` | 通过 `production-secret-audit-event.schema.json` 校验 |
| 缺必填字段 | `scripts/checks/fixtures/production-secret-audit-event-missing-required-negative-v1.json` | 被 schema 拒绝 |
| 禁止字段 | `scripts/checks/fixtures/production-secret-audit-event-forbidden-field-negative-v1.json` | 被 schema 拒绝 |
| 额外字段 | `scripts/checks/fixtures/production-secret-audit-event-additional-properties-negative-v1.json` | 被 schema 拒绝 |
| 非法 event kind | `scripts/checks/fixtures/production-secret-audit-event-event-kind-invalid-negative-v1.json` | 被 schema 拒绝 |
| artifact guard | `scripts/check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-v1.py` | 校验 schema、fixture、writer input compatibility smoke、forbidden runtime artifact 和 check-repo 注册 |

手动复验时可以运行：

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-v1.py
./scripts/check-repo.sh --fast
```

## 停止线

本契约不打开 runtime，不创建 audit writer、不创建 audit store、不执行 delivery、不创建 idempotency runtime、不选择 durable backend、不连接 production resolver、不调用云 secret 服务、不启用 DB / repository mode / production API。后续 runtime 仍必须按 durable backend、writer、delivery、idempotency、approval、credential handle、backend health 和 no leakage gate 逐项推进。

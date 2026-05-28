# Control Plane Read-Side 契约

更新时间：2026-05-27

## 契约目的

本专题说明 `Control Plane / User Workspace / Workflow v1` 的只读控制面契约层。它把未来用户工作区和管理端会消费的 summary、route、response 和 negative contract 先固定为可检查治理边界，避免在正式 Go API、数据库或 UI 尚未准备好时，从本地 ops console 直接堆出产品功能。

当前 read-side 契约只描述未来 API 应如何被设计和验证，不代表 API 已实现。

## 分层关系

当前 read-side 契约按四层固定：

1. `control-plane-read-model-v1`
   - 固定 tenant、application、API key、quota、workflow definition、run record 和 audit 的只读 summary 模型。
   - 固定访问策略、脱敏策略和停止线。
2. `control-plane-read-route-contract-v1`
   - 固定七类 tenant-scoped read-only route contract。
   - 固定 `GET` 方法、scope、分页 / 过滤、失败分类和 fail-closed 访问边界。
3. `control-plane-read-response-fixtures-v1`
   - 固定统一 response envelope：`request_id`、`tenant_ref`、`items`、`next_cursor`、`failure_code`、`audit_ref`。
   - 固定成功 / 失败样例、`failure_code` 来源和敏感字段脱敏。
4. `control-plane-read-negative-contract-v1`
   - 固定缺少身份、cross-tenant read、scope denied、invalid filter、forbidden method、forbidden query、forbidden fallback 和敏感字段投影拒绝。
   - 固定 fail-closed、无副作用、无 executor、无 database write、无 confirmation decision、无 writeback 和无 replay。

## 程序化证据

read-side 契约当前由以下 fixture 和 checker 固定：

- `scripts/checks/fixtures/control-plane-read-model-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-model-v1.py`
- `scripts/checks/fixtures/control-plane-read-route-contract-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-route-contract-v1.py`
- `scripts/checks/fixtures/control-plane-read-response-fixtures-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-response-fixtures-v1.py`
- `scripts/checks/fixtures/control-plane-read-negative-contract-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-negative-contract-v1.py`

这些 checker 已接入 `scripts/check-repo.py --fast`。它们的作用是防止契约、样例、负向边界和文档说明互相漂移，不负责启动服务或模拟真实数据库。

## 路由范围

当前固定的七类只读 route 是：

- `GET /v1/control-plane/tenants/{tenant_ref}/summary`
- `GET /v1/user-workspace/applications`
- `GET /v1/user-workspace/api-keys`
- `GET /v1/user-workspace/usage/quota-summary`
- `GET /v1/user-workspace/workflow-definitions`
- `GET /v1/user-workspace/runs`
- `GET /v1/control-plane/audit`

这些 route 都是未来 contract，不是当前已实现的 Go route handler。

## 输出边界

read-side 输出只允许暴露已脱敏 summary、计数、状态、成本摘要、trace id、audit ref 和 redacted secret reference。以下内容不得通过 read-side 输出或投影泄漏：

- raw secret value
- API key value 或 API key hash
- authorization header、bearer token、cookie value
- raw request body dump
- raw tool payload
- business writeback payload
- full prompt dump with secret

## 停止线

- 不把 read model 写成数据库 schema。
- 不把 route contract 写成 Go handler 已完成。
- 不把 response fixture 写成真实 API 返回。
- 不把 negative contract 写成 route smoke 已完成。
- 不通过 read-side 契约启用 OIDC、API key lifecycle、quota enforcement、billing ledger、workflow executor、confirmation、business writeback 或 replay。
- 不把本地 ops console 提升为正式 user workspace 或 production admin console。

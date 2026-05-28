# `Control Plane Read Formal UI Boundary` v1 计划

更新时间：2026-05-28

## 任务目标

本任务卡用于把 `control-plane-read-consumer-contract-v1` 之后的正式 UI 边界写清楚：哪些页面属于 `Admin Control Plane`，哪些页面属于 `User Workspace`，每个页面消费哪条 read route，以及 UI 必须保持哪些只读状态和停止线。

当前任务只定义 `control-plane-read-formal-ui-boundary-v1` 的正式 UI 边界，不实现 React 页面、不改 `apps/radishmind-console/`、不启动服务、不请求真实后端，也不把 fake-store-backed read route 写成生产管理端或正式用户端。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [Control Plane Read-Side 契约](../contracts/control-plane-read-side.md)
- [Control Plane Read Consumer Contract v1 计划](control-plane-read-consumer-contract-v1-plan.md)
- `contracts/typescript/control-plane-read-api.ts`
- `scripts/run-control-plane-read-consumer-smoke.py`
- `scripts/checks/fixtures/product-surface-v1-boundary.json`
- `scripts/checks/fixtures/control-plane-read-route-contract-v1.json`
- `scripts/checks/fixtures/control-plane-read-response-fixtures-v1.json`
- `scripts/checks/fixtures/control-plane-read-negative-contract-v1.json`
- `scripts/checks/fixtures/control-plane-read-consumer-contract-v1.json`

## 页面边界

1. `Admin Control Plane`
   - `admin-tenant-overview` 消费 `tenant-summary-route`。
   - `admin-audit-log` 消费 `audit-summary-list-route`。
   - 页面只允许 refresh、copy audit ref、过滤或跳转到相关只读视图。

2. `User Workspace`
   - `workspace-applications` 消费 `application-summary-list-route`。
   - `workspace-api-keys` 消费 `api-key-summary-list-route`。
   - `workspace-usage-quota` 消费 `quota-summary-route`。
   - `workspace-workflow-definitions` 消费 `workflow-definition-summary-list-route`。
   - `workspace-run-history` 消费 `run-record-summary-list-route`。
   - 页面只允许 refresh、copy audit ref、过滤或跳转到相关只读视图。

## UI 状态

每个正式 UI 页面必须覆盖以下状态：

- `loading`
- `ready`
- `empty`
- `denied`
- `stale`
- `partial_failure`
- `forbidden_projection`

`denied` 状态必须渲染空 items，并保留 `failure_code` 与 `audit_ref`。`stale` 状态只能保留上一份只读数据和诊断时间，不允许触发写入、执行或恢复。`forbidden_projection` 状态必须阻断敏感字段投影。

## 契约范围

- UI 只能消费 `ControlPlaneReadCollectionViewModel`、`ControlPlaneReadRouteCatalogViewModel` 和 `CONTROL_PLANE_READ_ROUTES`。
- UI 不得绕过 `contracts/typescript/control-plane-read-api.ts` 自行解释 response fixture。
- UI 不得渲染 `raw_secret_value`、`api_key_value`、`api_key_hash`、`authorization_header`、`bearer_token`、`cookie_value`、`raw_request_body_dump`、`raw_tool_payload`、`business_writeback_payload` 或 `full_prompt_dump_with_secret`。
- API key 页面只显示 summary、状态、scope、owner、时间和 audit ref，不显示 key value 或 hash。
- workflow definition 和 run history 页面只显示定义摘要、运行摘要、状态、失败分类、trace / audit ref，不创建、编辑、执行、恢复或 replay。

## 验收口径

- `scripts/checks/fixtures/control-plane-read-formal-ui-boundary-v1.json` 固定正式 UI 页面、路由分配、状态和只读不变量。
- `scripts/checks/control_plane/check-control-plane-read-formal-ui-boundary-v1.py` 校验依赖 gate、route 覆盖、surface 分配、敏感字段停止线、文档引用和 fast baseline 接入。
- checker 接入 `scripts/check-repo.py --fast`。
- 文档同步说明该切片只是正式 UI 边界，不代表 UI 实现 ready。

## 非目标

- 不实现正式 React 页面、生产管理端或正式 user workspace UI。
- 不修改 `apps/radishmind-console/`，不把本地 ops console 扩成 production console。
- 不启动服务、不调用真实 route、不暴露 test-only fake auth context。
- 不实现数据库 schema、migration、query、durable store 或 repository。
- 不实现 `Radish` OIDC middleware、token validation、login / logout route 或 production auth policy。
- 不实现 API key lifecycle、quota enforcement、rate limiting、billing ledger 或 cost calculation。
- 不实现 workflow builder、workflow executor、tool executor、confirmation flow、writeback、replay 或 materialized result reader。

## 停止线

- 不把 `control-plane-read-formal-ui-boundary-v1` 写成 React UI ready、formal user workspace UI ready、production admin console ready、real API consumer ready、database ready、OIDC ready 或 production ready。
- 不让任何页面暗示可以 create、edit、delete、issue、revoke、execute、confirm、writeback、materialize result 或 replay。
- 不把 `ControlPlaneReadCollectionViewModel` 写成真实数据库 read path。
- 不把 offline fixture、fake store 或 test-only fake auth context 写成正式 UI 数据源。

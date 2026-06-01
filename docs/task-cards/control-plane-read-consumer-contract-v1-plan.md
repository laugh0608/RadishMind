# `Control Plane Read Consumer Contract` v1 计划

更新时间：2026-05-28

## 任务目标

本任务卡用于把已实现的七条 fake-store-backed read route，收口成上层可消费的 TypeScript consumer contract 和离线消费 smoke。

当前任务不实现正式 user workspace UI、不实现 production admin console、不接真实 OIDC、不创建数据库 schema / migration / query、不实现 repository、不实现 API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。它只固定 `control-plane-read-consumer-contract-v1` 的 TypeScript 类型、route catalog、response envelope 消费、failure view、脱敏检查和上层消费停止线。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [Control Plane Read-Side 契约](../contracts/control-plane-read-side.md)
- [Control Plane Read Fake-Store Handler Implementation v1 计划](control-plane-read-fake-store-handler-implementation-v1-plan.md)
- [Control Plane Read Auth/DB Preconditions v1 计划](control-plane-read-auth-db-preconditions-v1-plan.md)
- `scripts/checks/fixtures/control-plane-read-fake-store-handler-implementation-v1.json`
- `scripts/checks/fixtures/control-plane-read-auth-db-preconditions-v1.json`
- `scripts/checks/fixtures/control-plane-read-route-contract-v1.json`
- `scripts/checks/fixtures/control-plane-read-response-fixtures-v1.json`
- `scripts/checks/fixtures/control-plane-read-negative-contract-v1.json`

## 契约范围

1. `TypeScript consumer contract`
   - 新增 `contracts/typescript/control-plane-read-api.ts`。
   - 固定七条 read route 的 route id、path、method、read model、required scope、分页模式、allowed filters、request 类型和 response envelope 类型。
   - 固定 `ControlPlaneReadCollectionViewModel` 和 `ControlPlaneReadRouteCatalogViewModel`，供未来 UI 或上层项目安全消费。

2. `Response consumption`
   - 上层只消费统一 envelope：`request_id`、`tenant_ref`、`items`、`next_cursor`、`failure_code`、`audit_ref`。
   - 成功响应可渲染脱敏 summary items；失败响应必须显示 denied 状态、无 items、保留 `failure_code` 和 `audit_ref`。
   - `next_cursor` 只能用于继续读取下一页，不代表执行、写入或 replay 能力。

3. `Sanitization and stop-lines`
   - consumer contract 必须显式保留 forbidden output keys，并提供 forbidden output 检测 helper。
   - view model 固定 `databaseBacked=false`、`formalUiReady=false`、`canMutate=false`、`canExecuteWorkflow=false`、`canWriteBusinessTruth=false`、`canRevealSecrets=false`。
   - 不把 fake-store-backed read route 写成正式 UI 或 production API ready。

4. `Offline consumer smoke`
   - 新增 `scripts/run-control-plane-read-consumer-smoke.py`。
   - 默认读取 committed response fixture，不启动服务、不绕过 test-only fake auth context、不请求真实后端。
   - smoke 验证 all routes consumed、envelope complete、failure views denied、failure no items、audit ref required、forbidden output absent、read-only views、secret projection disabled 和 negative contract invariants preserved。

## 验收口径

- `contracts/typescript/control-plane-read-api.ts` 固定 read-side 上层消费类型和 view model。
- `scripts/run-control-plane-read-consumer-smoke.py --check` 可离线验证 fixture 消费不变量。
- `scripts/checks/fixtures/control-plane-read-consumer-contract-v1.json` 固定契约、route catalog、consumer smoke 和停止线。
- `scripts/checks/control_plane/check-control-plane-read-consumer-contract-v1.py` 校验 TypeScript contract、response fixture、negative contract、fake-store implementation、文档引用和 fast baseline 接入。
- checker 接入 `scripts/check-repo.py --fast`。

## 非目标

- 不实现正式 React 页面、用户工作区页面、生产管理端或 control plane UI。
- 不启动服务、不调用真实 route、不暴露 test-only fake auth context。
- 不实现数据库 schema、migration、query、durable store 或 repository。
- 不实现 `Radish` OIDC middleware、token validation、login / logout route 或 production auth policy。
- 不实现 API key lifecycle、quota enforcement、rate limiting、billing ledger 或 cost calculation。
- 不实现 workflow builder、workflow executor、tool executor、confirmation flow、writeback、replay 或 materialized result reader。

## 停止线

- 不把 `control-plane-read-consumer-contract-v1` 写成 formal user workspace UI ready、production admin console ready、database ready、OIDC ready 或 production ready。
- 不让 TypeScript view model 暗示可以 mutate、execute workflow、write business truth、reveal secrets、confirm action 或 replay。
- 不把 offline consumer smoke 写成真实 API smoke。
- 不把 response fixture 写成真实数据库 read path。

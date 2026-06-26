# Production Secret Backend Fake Resolver Contract / No Secret Leakage Smoke Strategy v1 Plan

更新时间：2026-06-19

## 任务目的

本任务卡固定 production secret backend fake resolver implementation 前的静态 contract 和 no secret leakage smoke strategy。它承接 `production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1` 暴露的 `fake_resolver_contract_gate` 与 `no_secret_leakage_smoke_gate` 缺口，但不打开 fake resolver implementation。

## 当前结论

- 状态：`fake_resolver_contract_no_secret_leakage_smoke_strategy_defined`
- 本批新增：platform topic、task card、fixture、checker、check-repo 注册和文档入口同步
- 本批不新增：resolver runtime、fake resolver runtime、no secret leakage smoke runtime、cloud SDK、DB provider、DB driver、connection factory、SQL、schema marker、migration runner、repository mode、audit store 或 production API

## 输入

- `docs/platform/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.md`
- `scripts/checks/fixtures/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.json`
- `scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json`
- `scripts/checks/fixtures/production-secret-reference-basic.json`
- `scripts/checks/fixtures/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.json`
- `scripts/checks/fixtures/workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1.json`

## 输出

- `docs/platform/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.md`
- `docs/task-cards/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.json`
- `scripts/check-production-ops-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.py`

## 实施内容

1. 定义 fake resolver static contract：
   - 输入字段 allowlist
   - 输出字段 allowlist
   - failure mapping
   - sanitized diagnostics
   - forbidden secret-bearing fields

2. 定义 no secret leakage smoke strategy：
   - committed fixture / docs / checker 禁止 secret-looking value
   - future smoke 必须离线运行
   - future smoke 不读取环境 secret、不联网、不接 provider、不连接数据库
   - future smoke 不启用 repository mode

3. 更新 implementation readiness：
   - 增加 fake resolver contract 与 no secret leakage strategy 静态证据
   - 保持 `test-fixture-strategy` 仍为 `required_before_implementation`
   - 保持 `fake_resolver_status=not_created`

4. 更新前序 entry review：
   - 把 fake resolver contract 与 no secret leakage smoke 缺口从“未定义”推进到“静态策略已定义”
   - 保持 fake resolver implementation entry 不打开

## 停止线

- 不实现 fake resolver runtime
- 不实现 resolver runtime
- 不创建 `ProductionSecretBackendFakeResolver` 或同类 Go runtime type
- 不创建 no secret leakage smoke runtime / smoke runner
- 不读取真实环境变量
- 不保存 secret value
- 不连接数据库
- 不调用云 secret 服务
- 不启用 workflow saved draft repository mode
- 不新增 public production API

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-contract.py
./scripts/run-python.sh scripts/check-production-secret-reference-contract.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-secret-resolver-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```

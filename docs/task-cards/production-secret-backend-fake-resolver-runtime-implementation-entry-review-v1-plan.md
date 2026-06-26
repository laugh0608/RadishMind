# Production Secret Backend Fake Resolver Runtime Implementation Entry Review v1 Plan

更新时间：2026-06-20

## 任务目的

本任务卡承接 `fake_resolver_implementation_task_card_defined`，评审 fake resolver 是否可以从静态 implementation task card 进入下一张 test-only runtime implementation 任务卡。

## 当前结论

- 状态：`fake_resolver_runtime_implementation_entry_review_defined`
- Entry decision：`fake_resolver_runtime_implementation_ready_for_next_task`
- 本批新增：platform topic、task card、fixture、checker、check-repo 注册和文档入口同步
- 本批不新增：fake resolver runtime、resolver runtime、no secret leakage smoke runtime、runtime implementation task card、cloud SDK、DB provider、DB driver、connection factory、SQL、schema marker、migration runner、repository mode、audit store 或 production API

## 输入

- `docs/platform/production-secret-backend-fake-resolver-implementation-v1.md`
- `docs/task-cards/production-secret-backend-fake-resolver-implementation-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-fake-resolver-implementation-v1.json`
- `docs/platform/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.md`
- `scripts/checks/fixtures/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.json`
- `scripts/checks/fixtures/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.json`
- `scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json`
- `scripts/checks/fixtures/production-secret-reference-basic.json`
- `scripts/checks/fixtures/workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1.json`

## 输出

- `docs/platform/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.md`
- `docs/task-cards/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.json`
- `scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.py`

## 本批实施内容

1. 创建 runtime implementation entry review：
   - 判断现有 fake resolver implementation task card 是否足以进入下一张 runtime implementation 任务卡
   - 固定 disabled-by-default gate、placeholder secret ref fixture、environment binding、opaque handle metadata、sanitized diagnostics runtime emission 和 no leakage smoke 的准入结论
   - 明确本批只允许下一步创建 runtime implementation 任务卡，不创建 runtime

2. 更新 implementation readiness 总账：
   - `fake_resolver_runtime_implementation_entry_review_status=ready_for_next_task`
   - `fake_resolver_runtime_implementation_status=entry_review_defined_runtime_not_started`
   - 保持 `fake_resolver_status=not_created`
   - 保持 `no_secret_leakage_smoke_runtime_status=not_created`
   - 保持 `resolver_runtime_status=not_created`
   - 保持 `test_fixture_strategy_status=required_before_implementation`

3. 固定 failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard：
   - 不允许 secret-looking value
   - 不允许云 secret call、provider call、DB connection、SQL 或 repository mode
   - 不允许本批新增 fake resolver runtime、resolver runtime、runtime smoke runner 或 runtime implementation task card

4. 同步入口文档：
   - platform 入口、workflow 入口、Saved Workflow Draft v1、Workflow / Agent Runtime
   - current focus、能力矩阵、task card index
   - scripts README、services/platform README、deploy README 和周志

## 后续 Runtime 实现任务卡要求

下一张 runtime implementation 任务卡必须把以下项目作为验收项：

- fake resolver 默认 disabled，只能在显式 test fixture / smoke gate 下启用。
- 输入字段只允许 `environment`、`provider`、`provider_profile`、`secret_ref_key`、`secret_ref_version`、`purpose`、`request_id`、`audit_ref`、`policy_version`。
- 成功输出只能返回 opaque test credential handle metadata，不返回 credential payload、raw secret、DSN、provider raw URL 或 database hostname。
- 失败输出必须包含 `failure_code`、`sanitized_diagnostic`、`request_id`、`audit_ref` 和 `policy_version`，并 fail closed。
- no secret leakage smoke 必须扫描 committed fixture、runtime response、diagnostics 和 smoke record，且离线执行。
- side effect counters 必须保持零，不允许 provider call、cloud secret call、database connection、driver open、SQL execution 或 repository mode enablement。

## 停止线

- 不实现 fake resolver runtime。
- 不实现 resolver runtime。
- 不创建 no secret leakage smoke runtime / smoke runner。
- 不创建 fake resolver runtime implementation task card。
- 不读取真实环境变量。
- 不保存 secret value。
- 不连接数据库。
- 不调用云 secret 服务。
- 不启用 workflow saved draft repository mode。
- 不新增 public production API。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-fake-resolver-implementation-v1.py
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

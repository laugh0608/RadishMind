# Production Secret Backend Fake Resolver Runtime Implementation v1 Plan

更新时间：2026-06-20

## 任务目的

本任务卡承接 `fake_resolver_runtime_implementation_entry_review_defined`，完成 test-only fake resolver runtime 的代码实现，并保持 production secret backend、真实 resolver、云 secret、DB provider、repository mode 和 public production API 关闭。

## 当前结论

- 状态：`fake_resolver_runtime_test_only_implemented`
- Runtime scope：test-only、默认 disabled；`disabled-by-default runtime gate` 只能由 Go 测试显式启用
- 本批新增：`services/platform/internal/secretbackend/fake_resolver.go`、`services/platform/internal/secretbackend/fake_resolver_test.go`
- 本批不新增：production resolver runtime、cloud SDK、DB provider、DB driver、connection factory、SQL、schema marker、migration runner、repository mode、audit store 或 production API

## 输入

- `docs/platform/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.md`
- `docs/task-cards/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.json`
- `docs/platform/production-secret-backend-fake-resolver-implementation-v1.md`
- `scripts/checks/fixtures/production-secret-backend-fake-resolver-implementation-v1.json`
- `docs/platform/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.md`
- `scripts/checks/fixtures/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.json`
- `scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json`
- `scripts/checks/fixtures/production-secret-reference-basic.json`

## 输出

- `services/platform/internal/secretbackend/fake_resolver.go`
- `services/platform/internal/secretbackend/fake_resolver_test.go`
- 更新 `scripts/checks/fixtures/production-secret-backend-fake-resolver-runtime-implementation-v1.json`
- 更新 `scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-v1.py`
- 同步平台专题、workflow 专题、current focus、scripts README、services/platform README、deploy README 和周志

## 本批实施内容

1. 实现 fake resolver runtime：
   - `NewDisabledFakeResolver()` 和零值配置默认 disabled
   - `NewFakeResolver(FakeResolverConfig{Enabled: true, ...})` 仅供测试显式启用
   - `Resolve` 只读取结构化 request 字段，不读环境、不联网、不访问 provider 或 DB

2. 固定输入 / 输出边界：
   - 输入只允许 `environment`、`provider`、`provider_profile`、`secret_ref_key`、`secret_ref_version`、`purpose`、`request_id`、`audit_ref`、`policy_version`
   - 成功只输出 opaque test credential handle metadata
   - 失败只输出 `failure_code`、`sanitized_diagnostic`、`request_id`、`audit_ref`、`policy_version`

3. 固定 no leakage 和 side effect counters：
   - secret-looking input 必须 fail closed
   - failure result 不回显 secret-looking input
   - side effect counters 全部为零

4. 更新 implementation readiness：
   - `fake_resolver_runtime_implementation_status=fake_resolver_runtime_test_only_implemented`
   - `fake_resolver_runtime_status=implemented_test_only_disabled_by_default`
   - `test_fixture_strategy_status=satisfied_for_test_only_fake_resolver`
   - `production_secret_backend` 仍为 `not_satisfied`
   - `resolver_runtime_status` 仍为 `not_created`

## 停止线

- 不实现 production resolver runtime。
- 不读取真实环境变量。
- 不保存或返回 secret value。
- 不调用云 secret 服务。
- 不访问 provider。
- 不创建 credential payload。
- 不连接数据库。
- 不打开 DB driver / connection factory。
- 不执行 SQL。
- 不读取或写入 schema marker。
- 不启用 workflow saved draft repository mode。
- 不新增 public production API。

## 验证

```bash
cd services/platform && go test ./internal/secretbackend
cd ../..
./scripts/run-python.sh scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-fake-resolver-implementation-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.py
./scripts/check-repo.sh --fast
```

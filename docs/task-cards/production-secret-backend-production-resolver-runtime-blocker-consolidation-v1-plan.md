# Production Secret Backend Production Resolver Runtime Blocker Consolidation v1 计划

状态：`production_resolver_runtime_blocker_consolidation_defined`

## 目标

把 production resolver runtime 的剩余 blocker 从 real resolver entry refresh、audit store entry refresh v3、credential handle / operator approval / backend health / no leakage runtime entry review，以及 Saved Workflow Draft durable store 依赖中收束成可检查矩阵。

本任务卡不创建 production resolver runtime implementation task card，不实现 production resolver runtime，不读取真实 secret，不调用云 secret 服务，不连接数据库，不创建 repository mode runtime 或 production API。

## 输入

- `real_resolver_runtime_implementation_entry_refresh_defined`
- `audit_store_runtime_implementation_entry_refresh_v3_defined`
- `credential_handle_runtime_implementation_entry_review_defined`
- `operator_approval_runtime_implementation_entry_review_defined`
- `resolver_backend_health_runtime_implementation_entry_review_defined`
- `real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_defined`
- `draft_database_secret_resolver_runtime_dependency_refresh_defined`
- `draft_negative_auth_smoke_runtime_readiness_defined`

## 本批交付

1. 新增 platform doc：`docs/platform/production-secret-backend-production-resolver-runtime-blocker-consolidation-v1.md`。
2. 新增 fixture：`scripts/checks/fixtures/production-secret-backend-production-resolver-runtime-blocker-consolidation-v1.json`。
3. 新增 checker：`scripts/check-production-ops-secret-backend-production-resolver-runtime-blocker-consolidation-v1.py`。
4. 将 checker 接入 `./scripts/check-repo.sh --fast`。
5. 同步 current focus、platform / features / workflow 入口、integration contracts、scripts README、task card index 和周志。

## 准入判断

- production resolver runtime task card：仍 blocked，不能创建。
- credential handle、operator approval、audit store、backend health 和 no leakage smoke runtime：均仍 blocked before runtime task card。
- workflow database secret resolver：仍 blocked before implementation task card。
- negative auth smoke：只有 runtime readiness，runtime smoke artifact 仍未创建。
- repository mode / DB runtime / production API：仍 disabled / blocked / not created。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-production-resolver-runtime-blocker-consolidation-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-credential-handle-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-operator-approval-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-secret-resolver-runtime-dependency-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-negative-auth-smoke-runtime-readiness-v1.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 production resolver runtime implementation task card。
- 不创建 production resolver runtime、cloud secret client、credential payload、credential handle runtime、operator approval runtime、audit store runtime、backend health runtime、no leakage smoke runtime、database connection provider、repository mode runtime、production API、executor、confirmation、writeback 或 replay。
- 不读取真实 secret，不调用云 secret 服务，不访问 provider，不连接数据库，不运行 SQL，不读写 schema marker。

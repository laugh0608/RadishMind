# Control Plane Read Repository Contract Smoke v1 任务卡

## 任务标识

- 切片：`control-plane-read-repository-contract-smoke-v1`
- 轨道：`Control Plane / User Workspace / Workflow v1`
- 状态：`repository_contract_smoke_defined`

## 目标

在 repository implementation 之前固定未来 repository contract smoke 的输入输出、七条 read route 覆盖、failure mapping、no fake fallback、no side effects 和文档停止线。该切片只定义未来 smoke 应验证什么，不实现 SQL、migration、repository adapter、真实数据库、Radish OIDC、token validation 或 production API consumer。

## 范围

- 固定 `ControlPlaneReadRepositoryContractSmoke` 的输入字段、repository context 字段、request 字段和输出 envelope 字段。
- 固定七条 read route 到 repository operation 的 smoke matrix。
- 固定成功输出只能是 sanitized summary envelope，失败输出只能返回 failure code，不泄漏数据库内部细节。
- 固定 `tenant_binding_missing`、`scope_denied`、`invalid_filter`、`read_store_unavailable`、`read_store_contract_mismatch`、`database_read_disabled` 和 `auth_context_contract_mismatch` 的 failure mapping。
- 固定 database mode 不得回退到 fixture fake store，未知 filter 不得回退成功。
- 固定 repository write、executor、confirmation、business writeback 和 replay 计数必须保持为 `0`。
- 将 checker 接入仓库 fast baseline。

## 不在本次范围

- 不实现 repository interface、repository adapter、SQL query、database schema 或 migration。
- 不新增正式配置入口，也不启用 database / postgres / repository read mode。
- 不接 Radish OIDC、auth middleware、token validation、production API consumer、API key lifecycle 或 quota enforcement。
- 不实现 workflow executor、confirmation、business writeback、replay 或任何写能力。

## 验证

- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-repository-contract-smoke-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-disabled-database-guard-v1.py`
- `./scripts/check-repo.sh --fast`

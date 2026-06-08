# Control Plane Read Repository Implementation Readiness v1 任务卡

## 任务标识

- 切片：`control-plane-read-repository-implementation-readiness-v1`
- 轨道：`Control Plane / User Workspace / Workflow v1`
- 状态：`repository_implementation_readiness_defined`

## 目标

在真实 read repository implementation 前固定 implementation readiness：未来文件落点、实现准入 gate、七条 route 的 readiness matrix、dual smoke plan、failure mapping、no fake fallback、no side effects 和停止线。该切片只定义实现前置条件，不实现 repository interface、repository adapter、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer。

本任务卡中的 `repository implementation readiness` 只表示实现前置条件已定义，不表示 repository implementation ready。

## 范围

- 固定未来 repository contract / adapter 文件的计划落点，但不创建 Go repository 文件。
- 固定 repository contract types、auth context injection、store selection guard、schema / migration plan、dual smoke 和 production auth 的准入 gate。
- 固定七条 read route 的 future implementation readiness matrix。
- 固定 fake-store route smoke、repository contract smoke、disabled database guard 和 future durable adapter smoke 的 dual smoke 顺序。
- 固定 failure mapping 不能泄漏数据库内部错误，未知 filter 和缺失 tenant 不得成功。
- 将 checker 接入仓库 fast baseline。

## 不在本次范围

- 不实现 repository interface、repository adapter、SQL query、database schema 或 migration。
- 不新增正式配置入口，也不启用 database / postgres / repository read mode。
- 不接 Radish OIDC、auth middleware、token validation、production API consumer、API key lifecycle 或 quota enforcement。
- 不实现 workflow executor、confirmation、business writeback、replay 或任何写能力。

## 验证

- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-repository-implementation-readiness-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-repository-contract-smoke-v1.py`
- `./scripts/check-repo.sh --fast`

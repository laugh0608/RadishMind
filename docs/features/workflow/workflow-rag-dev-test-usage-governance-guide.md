# Workflow RAG 开发测试态使用与资源治理指南

更新时间：2026-07-18

## 适用范围

本文是 Workflow RAG 当前已实现能力的使用说明和长期资源边界参考，覆盖：

1. 应用知识快照及精确版本管理；
2. Saved Workflow Draft 绑定 `rag_ref` 后的单次检索执行；
3. Run History、Comparison、Evaluation、Baseline 与 Suite 对 RAG run v3 的审查；
4. synthetic-public 离线质量评测；
5. 应用作用域 durable evaluation dataset 与 baseline / candidate snapshot review；
6. 人工 knowledge promotion、不可变 application binding、配置草案 attach 与发布候选重校验。

这些能力只属于内部开发测试态。当前已批准的 publish candidate 不能直接通过应用 API key 发起 RAG 调用；runtime assignment、`application_rag:invoke` 与 `workflow_run_record.v4` 尚未实现。现有运行入口仍是精确 Saved Workflow Draft 的独立 retrieval execution route。

## 资源链与真相源

| 资源 | 真相源与职责 | 不承担的职责 |
| --- | --- | --- |
| `workflow_rag_snapshot.v1` | 保存应用作用域、不可变版本、fragment 和 `rag_ref` | 不自动抓取、同步或成为上层业务真相源 |
| Saved Workflow Draft | 保存精确 `rag_ref`、模型与受控图配置 | 不复制 snapshot fragment，不自动执行 |
| `workflow_run_record.v3` | 保存 retrieval / provider 调用、排名与 citation metadata | 不保存输入、fragment 正文、prompt、完整回答或模型原始响应 |
| `workflow_rag_evaluation_dataset_resource.v1` | 保存应用作用域 dataset current projection 和不可变版本 | 不保存 candidate review 结果或自动选择 baseline |
| `workflow_rag_candidate_snapshot_review.v1` | 保存 baseline / candidate 的确定性 lexical 对比结果 | 不调用 Gateway，不创建 workflow run，不修改 snapshot |
| promotion candidate / decision | 封印 exact dataset、review、两侧 snapshot、profile 与 source draft，承载人工决定 | 不自动 baseline、release、publish 或运行 |
| `workflow_rag_application_binding.v1` | 为批准结果签发不可变、ref-only 配置绑定资格 | 不复制 query、fragment、指标、配置正文或模型材料 |
| application draft v2 / publish candidate v2 | 只引用 `binding_id / binding_version / binding_digest` 并重新校验资格 | 不成为知识证据的第二真相源 |

所有资源都绑定 tenant、workspace、application 和 owner。跨作用域读取、版本 / digest / `rag_ref` 不一致、归档、取消、损坏记录或 store failure 均失败关闭，不回退 fake store、旧版本或内存实现。

## 本地 Web 使用入口

### 快照、检索执行与 run v3

在仓库根目录运行：

```bash
./scripts/run-radishmind-web-dev.sh --mode dev-live --workflow-rag-dev
```

Windows / PowerShell：

```powershell
pwsh ./scripts/run-radishmind-web-dev.ps1 -Mode dev-live -WorkflowRAGDev
```

该入口使用本地产品档的 shared SQLite database，启用 Saved Draft、snapshot、独立 retrieval execution、Gateway mock bridge、Run History 与 RAG comparison / evaluation。Web 使用：

1. 在 Application Detail 的“知识快照与精确版本”区域创建或读取 snapshot；
2. 在 Draft Designer 保存包含精确 `rag_ref` 的 eligible draft；
3. 在“精确知识版本与显式执行”区域恢复该 draft，提交一次输入并显式执行；
4. 从返回的 run id 进入 Run History，审查有效 snapshot、profile、fragment refs、ranks、citations、failure boundary 和副作用计数；
5. 选择同一 binding 的两个 v3 run 进入 Comparison / Evaluation，不会重新执行 retrieval 或 Gateway。

默认 provider 为本地 mock。该入口会启动开发服务，长期运行时应由开发者在本机终端保持，并在验证结束后停止。

### 评测数据集、晋级、配置绑定与发布重校验

SQLite 本地产品档使用：

```bash
./scripts/run-radishmind-web-dev.sh --mode dev-live --workflow-rag-promotion-local-product
```

Windows / PowerShell：

```powershell
pwsh ./scripts/run-radishmind-web-dev.ps1 -Mode dev-live -WorkflowRAGPromotionLocalProduct
```

该入口复用同一 workflow backend selector、shared SQLite database、snapshot repository、evaluation dataset repository、application configuration draft repository 与 publish governance。Web 使用顺序固定为：

1. 在 evaluation 区创建完整 dataset version，或读取已有 exact version；
2. 人工选择 baseline / candidate snapshot，显式创建 candidate review；
3. 从 exact dataset / review / snapshots / profile / source draft 创建 promotion candidate；
4. reviewer 以最新 `record_version` 和理由执行 `approve / reject / defer / cancel`；
5. `approve` 原子签发不可变 binding，但不修改任何来源资源；
6. 在 Application Configuration Draft 面板恢复 exact source draft，显式 attach binding，创建 draft v2；
7. 从该草案创建 publish candidate v2，并在 create、approve 和读取时重新校验 binding 资格。

以上步骤互不自动触发。promotion approve 不等于 attach，attach 不等于 publish approve，publish approve 也不等于正式发布或运行时激活。

## 开发身份与权限

当前管理与 Saved Draft retrieval execution 都使用受验证的开发身份头，不接受 API key：

```text
X-RadishMind-Dev-Read-Identity: <request identity>
X-RadishMind-Dev-Read-Tenant: <tenant ref>
X-RadishMind-Dev-Read-Subject: <owner subject>
X-RadishMind-Dev-Read-Scopes: <comma-separated scopes>
X-RadishMind-Dev-Read-Audit: <audit ref>
```

主要权限如下：

| 能力 | 权限 |
| --- | --- |
| snapshot | `workflow_rag_snapshots:read/write/archive` |
| retrieval execution | `workflow_rag:execute`、`workflow_runs:execute`、`workflow_drafts:read`、`workflow_rag_snapshots:read` |
| dataset metadata | `workflow_rag_evaluation_datasets:read` |
| dataset query / review note detail | `workflow_rag_evaluation_datasets:read_content` |
| dataset version / review / archive | `workflow_rag_evaluation_datasets:write/review/archive` |
| promotion list / create / decision | `workflow_rag_promotions:read/write/review`，create 还需 dataset、snapshot 与 application draft read |
| 配置草案 attach / replace | `application_drafts:read/write` 与 `workflow_rag_promotions:bind` |
| 发布候选重校验 | application publish 权限与 `workflow_rag_promotions:read` |

OIDC integration test 或 production auth 不会因这些开发权限自动启用。管理 API 和当前 retrieval execution route 也不会接受开发测试态 API key 替代上述身份上下文。

## HTTP 资源参考

| 方法与路径 | 用途 |
| --- | --- |
| `POST /v1/user-workspace/workflow-retrieval-snapshots` | 创建 snapshot v1 |
| `GET /v1/user-workspace/workflow-retrieval-snapshots` | metadata-only 列表 |
| `GET /v1/user-workspace/workflow-retrieval-snapshots/{snapshot_id}` | 读取精确版本；需要 read scope 才返回有界 fragment |
| `POST /v1/user-workspace/workflow-retrieval-snapshots/{snapshot_id}/versions` | 完整替换创建新版本 |
| `POST /v1/user-workspace/workflow-retrieval-snapshots/{snapshot_id}/archive` | 以 expected latest version 归档 |
| `POST /v1/user-workspace/workflow-drafts/{draft_id}/retrieval-executions` | 重读 exact draft / snapshot 后执行一次 retrieval 和一次 Gateway |
| `POST /v1/user-workspace/workflow-rag-evaluation-datasets` | 创建应用作用域 dataset |
| `GET /v1/user-workspace/workflow-rag-evaluation-datasets` | metadata-only dataset 列表 |
| `GET /v1/user-workspace/workflow-rag-evaluation-datasets/{dataset_id}` | 读取 exact dataset version；正文受 `read_content` 保护 |
| `POST /v1/user-workspace/workflow-rag-evaluation-datasets/{dataset_id}/versions` | 以 CAS 创建完整 replacement |
| `POST /v1/user-workspace/workflow-rag-evaluation-datasets/{dataset_id}/archive` | 以 CAS 归档 dataset |
| `POST /v1/user-workspace/workflow-rag-evaluation-datasets/{dataset_id}/candidate-reviews` | 创建 baseline / candidate review |
| `GET /v1/user-workspace/workflow-rag-evaluation-datasets/{dataset_id}/candidate-reviews` | metadata-only review 列表 |
| `GET /v1/user-workspace/workflow-rag-evaluation-datasets/{dataset_id}/candidate-reviews/{review_id}` | 读取 exact review evidence |
| `POST /v1/user-workspace/workflow-rag-knowledge-promotion-candidates` | 创建 exact promotion candidate |
| `GET /v1/user-workspace/workflow-rag-knowledge-promotion-candidates` | metadata-only candidate 列表 |
| `GET /v1/user-workspace/workflow-rag-knowledge-promotion-candidates/{candidate_id}` | 读取 evidence、decisions、binding 与动态 eligibility |
| `POST /v1/user-workspace/workflow-rag-knowledge-promotion-candidates/{candidate_id}/decisions` | 人工 decision 与 expected-version CAS |

请求体使用 strict JSON，未知字段会被拒绝。精确字段和 schema 以 `contracts/workflow-rag-*.schema.json`、`contracts/workflow-run-record-v3.schema.json` 及对应 Go request type 为准。

## 离线质量资产

committed starter dataset、snapshot 和 report 位于 `datasets/eval/workflow-rag/`。只读漂移检查：

```bash
cd services/platform
go run ./cmd/radishmind-workflow-rag-eval \
  --snapshot ../../datasets/eval/workflow-rag/snapshots/dev_core_v1.json \
  --dataset ../../datasets/eval/workflow-rag/datasets/dev_core_v1.json \
  --output ../../datasets/eval/workflow-rag/reports/dev_core_v1.review.json \
  --check
```

CLI 每个样本只调用一次既有确定性 lexical ranker，不调用 Gateway、不创建 workflow run。committed 资产只允许 synthetic-public query / fragment；workspace-internal query 只能进入受权限保护的 durable dataset detail，不得提交到仓库。

## 持久化与迁移

| 模式 | 用途 | 当前 Workflow RAG schema |
| --- | --- | --- |
| `memory_dev` | 单元测试与快速领域验证 | 进程退出后不保留 |
| `sqlite_dev` | 本地连续开发，共享一个 workflow database | `0005` execution audit、`0007` evaluation dataset、`0008` promotion；marker `workflow_run_store_sqlite_v8` |
| `postgres_dev_test` | migration、角色、方言、并发与重启同构验证 | `0009` execution audit、`0010` evaluation dataset、`0011` promotion；marker `workflow_run_store_v11` |

三种模式复用现有 workflow backend selector。SQLite 只通过 shared local-product runtime 启用；PostgreSQL 只复用既有 workflow pool。缺少 marker、database / pool、checksum 不一致、未知 selector 或记录损坏时启动或请求失败，不创建平行数据库、DSN、pool 或 memory fallback。

PostgreSQL 专项验证继续使用统一 runner：

```bash
./scripts/run-workflow-saved-draft-postgres-dev-test.sh check
./scripts/run-workflow-saved-draft-postgres-dev-test.sh down
```

`check` 会执行迁移、回滚 / 重应用、运行角色 DDL 拒绝、并发、重启、损坏记录和 no-fallback 验证；结束后使用 `down` 关闭容器。

## 隐私、失败与副作用

- snapshot exact detail 和 dataset content detail 是仅有的受权限正文读取路径；list、candidate review、promotion list、audit、日志、普通 Run History 与 Gateway history 均保持 metadata-only。
- 不保存真实用户 query、运行输入、fragment 正文、excerpt、prompt、完整 answer、模型原始响应、token、credential 或 secret。
- v3 成功执行固定 `retrieval_calls=1`、`provider_calls=1`；`tool_calls`、`confirmation_calls`、`business_writes` 与 `replay_writes` 为 0。
- candidate review 每个样本对 baseline 与 candidate 各调用一次 ranker，Gateway 与 workflow run 调用数为 0。
- promotion、binding attach 与 publish review 的 retrieval、Gateway、workflow run、业务写入和 replay 调用数均为 0。
- CAS 冲突不自动 retry / merge；先重新读取权威版本，再由用户决定是否提交。

## 常见问题

| 现象 | 检查项 |
| --- | --- |
| route 提示 dev capability disabled | 检查对应 `RADISHMIND_WORKFLOW_RAG_*_DEV=1` 与 launcher 模式 |
| `scope_denied` | 检查受验证主体、tenant / owner binding 和操作所需全部 scope |
| draft / snapshot mismatch | 重新读取 Saved Draft 与 exact `rag_ref`，不要用 latest snapshot 代替绑定版本 |
| dataset content 不返回 | 读取 exact version 时显式提供 `workflow_rag_evaluation_datasets:read_content` |
| version conflict | 读取当前 latest / record version 后重新人工决定，不盲目重试 |
| promotion eligibility 出现 blocker | 检查 dataset / review / snapshots / profile / source draft / application 是否漂移、取消或归档 |
| store unavailable / schema mismatch | 检查 backend selector、shared database / pool、migration marker 和 checksum；不会回退 memory |
| publish approved 但无法用 API key RAG 调用 | 符合当前边界；runtime assignment 与 `application_rag:invoke` 尚未实现 |

## 相关设计与契约

- [RAG Retrieval 与应用知识快照](rag-retrieval-application-knowledge-snapshot-dev-test-v1.md)
- [RAG Regression Review 与 Evaluation Profile](workflow-rag-regression-review-evaluation-profile-dev-test-v1.md)
- [RAG 评测数据集与知识质量审查](workflow-rag-evaluation-dataset-knowledge-quality-review-v1.md)
- [RAG 评测数据集应用资源化与候选快照审查](workflow-rag-evaluation-dataset-application-resource-candidate-snapshot-review-v1.md)
- [RAG 知识基线晋级与应用配置绑定审查](workflow-rag-knowledge-baseline-promotion-application-binding-review-v1.md)
- [Platform Service Layer](../../../services/platform/README.md)
- [Web 使用说明](../../../apps/radishmind-web/README.md)
- [契约索引](../../../contracts/README.md)

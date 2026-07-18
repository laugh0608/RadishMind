# Workflow RAG 评测数据集应用资源化与候选快照审查 v1

更新时间：2026-07-18

状态：`workflow_rag_evaluation_dataset_application_resource_candidate_snapshot_review_v1_completed`

## 功能目标

把已经验证的离线 RAG dataset 升级为应用作用域的开发测试态持久资源，并允许内部维护者以一份人工审查过的 dataset version 对照精确 baseline / candidate immutable snapshot。系统同步复算两侧 lexical retrieval，保存 metadata-only candidate review，使知识变更在进入 Workflow RAG execution 前具备可恢复、可比较、可审计的人工审查路径。

本专题不执行 Workflow、不调用 Gateway、不生成答案，也不自动晋级 snapshot、baseline、Evaluation Case、Suite 或 release decision。

## 用户路径

1. 维护者在一个有效应用下选择精确 baseline snapshot version，创建 dataset resource，提交数据分类、阈值、人工 review summary 与 1–200 个样本。
2. 服务端重新读取 baseline snapshot，校验 tenant / workspace / application 三层 scope、version / digest / `rag_ref`、profile digest、classification 与所有 expected refs，再原子创建 dataset v1。
3. 后续编辑提交完整 replacement 和 `expected_latest_version`；store 以事务 CAS 创建不可变 dataset version，旧版本仍可精确读取。
4. 维护者选择精确 candidate snapshot version，提交 dataset version / digest 与 candidate binding。服务端重新读取 dataset、baseline 与 candidate，不接受客户端指标、排名或 fragment 内容。
5. 服务端对每个样本分别对 baseline、candidate 恰好执行一次既有 lexical ranker，保存 metadata-only review；不 retry、fallback 或后台执行。
6. Web 展示 dataset lifecycle、版本 binding、baseline / candidate 指标差异、样本状态变化、finding 与稳定结论；workspace query 只在具备内容读取权限的 dataset detail 编辑面出现。

## 数据分类与内容边界

| dataset classification | snapshot classification | 允许内容 | 返回边界 |
| --- | --- | --- | --- |
| `synthetic_public` | `public` | 人工编写的公开测试 query / review note | 授权 detail 可返回正文；list、review、audit、日志只返回摘要 / digest |
| `workspace_internal` | `workspace_internal` | 已授权工作区内 query / review note，不得含 credential | 仅 `workflow_rag_evaluation_datasets:read_content` detail 返回正文；其它响应 metadata-only |

`workspace_internal` 只进入所选 Workflow durable backend，不进入 committed `datasets/eval/`、普通日志、candidate review、audit 或 Run History。所有模式继续拒绝 credential、token、任意 URL secret 和模型响应。

## 独立门禁与权限

开发测试态门禁固定为：

```text
RADISHMIND_WORKFLOW_RAG_EVALUATION_DEV=1
```

不得复用 snapshot 写入门禁、retrieval execution 门禁或 executor v0 门禁。门禁开启必须同时具备 verified actor 与现有 Workflow durable backend；`sqlite_dev` / `postgres_dev_test` 初始化失败时服务启动失败，不回退 memory。

权限矩阵：

| 操作 | 权限 |
| --- | --- |
| list / metadata read / review read | `workflow_rag_evaluation_datasets:read` |
| exact dataset content read | `workflow_rag_evaluation_datasets:read`、`workflow_rag_evaluation_datasets:read_content` |
| create / version | `workflow_rag_evaluation_datasets:write`、`workflow_rag_snapshots:read` |
| archive | `workflow_rag_evaluation_datasets:archive` |
| candidate review | `workflow_rag_evaluation_datasets:review`、`workflow_rag_evaluation_datasets:read`、`workflow_rag_snapshots:read` |

## API 与 strict JSON

首版注册以下独立接口：

```text
POST /v1/user-workspace/workflow-rag-evaluation-datasets
GET  /v1/user-workspace/workflow-rag-evaluation-datasets
GET  /v1/user-workspace/workflow-rag-evaluation-datasets/{dataset_id}
POST /v1/user-workspace/workflow-rag-evaluation-datasets/{dataset_id}/versions
POST /v1/user-workspace/workflow-rag-evaluation-datasets/{dataset_id}/archive
POST /v1/user-workspace/workflow-rag-evaluation-datasets/{dataset_id}/candidate-reviews
GET  /v1/user-workspace/workflow-rag-evaluation-datasets/{dataset_id}/candidate-reviews
GET  /v1/user-workspace/workflow-rag-evaluation-datasets/{dataset_id}/candidate-reviews/{review_id}
```

写请求使用有界 strict JSON。客户端只提交 scope、resource key / display name、classification、精确 snapshot binding、阈值、样本、人工 review summary、CAS version。不得提交 dataset id / digest、profile digest、selected fragments、排名、指标、finding、review id、actor / audit、candidate report 或任何模型内容。

## 资源、版本与 CAS

- resource 保存稳定 `dataset_id` / `dataset_key`、lifecycle、latest version / digest、baseline binding 摘要、sample count 与时间戳，不保存 query 正文。
- immutable version 使用 `workflow_rag_evaluation_dataset_resource.v1`，包含精确 baseline snapshot / lexical profile binding、阈值、review metadata 和样本；digest 覆盖全部不可变字段。
- create / version 必须在同一事务内写 resource projection、immutable version 和 metadata-only audit；archive 只改变 resource lifecycle，并追加 audit。
- version / archive 使用 `expected_latest_version` CAS；stale / future version 都返回权威 current version，不进行自动合并。
- candidate review 使用 `workflow_rag_candidate_snapshot_review.v1`，绑定精确 dataset / baseline / candidate / profile digest；review append-only，不更新 dataset 或 snapshot。

## 候选快照比较

baseline 使用 dataset version 已固定的 snapshot。candidate 必须属于同一 scope、classification 和 lexical profile，并由服务端按 id / version 重读且精确比对 digest / `rag_ref`。允许审查 active 或 archived immutable version，但 lifecycle 会进入 binding evidence；该行为不代表可执行或可发布。

review 保存两份既有 `workflow_rag_quality_review.v1` 的 metadata-only evidence，并派生：

- 六项 metric 的 candidate-minus-baseline delta；
- `improved / unchanged / regressed / needs_review` 结论；
- 每样本 `passed / failed` 状态变化、expected hit / official hit / first rank 变化；
- candidate 新增 / 移除 finding code；
- baseline / candidate fragment count、official count、source distribution。

任一侧 contract、binding、ranker 或 store 失败都 fail closed。`regressed` 包含 candidate 从通过变失败、任一指标下降、预期 / official hit 丢失或新增 `review_required` finding；只有无 regression 且至少一项指标或样本改善时为 `improved`。

## Store 与 migration

- `memory_dev`：有界 scoped resource / immutable versions / append-only reviews / audits。
- `sqlite_dev`：复用 workflow run shared SQLite database，新增下一号 migration；resource CAS 使用事务条件更新，version / review / audit append-only。
- `postgres_dev_test`：复用 workflow run pool，新增下一号 migration、marker / checksum / pending migration 链与重启测试。
- backend selector 从现有 workflow run store 派生；不创建平行 DSN、pool、database file 或 memory fallback。

## Web

Web 在应用知识快照面板之后新增 lazy-loaded dataset / candidate review panel：

- 默认 offline 零请求；只有显式 `VITE_RADISHMIND_WORKFLOW_RAG_EVALUATION_SOURCE=dev-workflow-rag-evaluation-http` 才访问 API。
- list 与 review history 默认 metadata-only；读取 / 编辑 exact dataset version 时请求内容读取权限。
- 创建 / 版本化使用完整 replacement，显示 CAS conflict 的权威 version / lifecycle。
- candidate review 由用户显式触发，显示 exact binding、指标 delta、样本变化和 finding；不显示 fragment 正文、完整 query、prompt、answer 或模型响应。

## 已完成实现

1. 批次 A 已完成两份资源契约、独立配置门禁、领域校验 / 对比、memory store、八条 strict HTTP route 与精准测试。
2. 批次 B 已完成 SQLite `0007`、PostgreSQL `0010`、三种 store selector、事务 CAS、append-only review / audit、重启恢复与 no-fallback；PostgreSQL schema marker 已推进为 `workflow_run_store_v10`。
3. 批次 C 已完成 strict Web consumer、lazy panel、离线零请求、完整替换 / CAS、metadata-only candidate review、Web tests / build 与仓库收口；按停止线未启动真实浏览器验收。

三个批次由同一张任务卡承接，完成锚点为 `workflow_rag_evaluation_dataset_application_resource_candidate_snapshot_review_v1_completed`。功能已关闭，不派生批次 D。

## 后续顺位

下一产品动作应先在功能设计层评审“Workflow RAG 知识基线晋级与应用配置绑定审查 v1”：由维护者基于精确 dataset version 与已完成 candidate review，人工形成不可变晋级候选和配置绑定审查，复用现有应用配置草案 / 发布治理的 CAS、漂移和审计边界。该方向当前仅为 `workflow_rag_knowledge_baseline_promotion_application_binding_review_v1_ready_for_design`，尚未获得实现准入；不得从本专题自动修改 snapshot、dataset baseline、应用配置草案或发布状态，也不得打开自动 promotion / release、Gateway execution、connector、在线搜索或生产能力。

## 验收

- gate、verified actor、权限矩阵、strict JSON、三层 scope；
- create / exact read / list / version / archive、CAS、immutable history；
- baseline / candidate id / version / digest / `rag_ref` / classification / profile 精确校验；
- synthetic-public / workspace-internal 正反例、secret rejection、空 query、非法 refs、预算；
- baseline / candidate 每样本各一次 ranker，comparison 正反例和稳定结论；
- dataset list / review / audit / log metadata-only，content detail 权限隔离；
- memory / SQLite / PostgreSQL 行为一致、重启恢复、append-only、损坏记录拒绝、no-fallback；
- Web strict consumer、CAS、candidate review、敏感字段拒绝、offline 零请求、tests / build；
- 既有 offline CLI / committed report、RAG execution、Run History v3、Regression Review、executor v0 与 HTTP Tool v2 不回归。

## 停止线

- 不启用 crawler、connector、在线搜索、文件扫描、embedding、vector database、reranker 或外部 provider。
- 不调用 Gateway，不生成 answer，不创建 run，不改变 executor v0 `allow_retrieval=false`。
- 不自动修改 snapshot、dataset baseline、Evaluation Case / Suite、release decision 或应用发布状态。
- 不实现自动 baseline、自动 promotion、自动 release、后台调度、retry、replay 或 production enablement。

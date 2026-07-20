# Workflow RAG 评测数据集与知识质量审查 v1

更新时间：2026-07-18

状态：`workflow_rag_evaluation_dataset_knowledge_quality_review_v1_completed`

## 功能目标

为内部开发者提供一条可提交、可复算、可审查的 RAG 离线质量链：人工维护 synthetic-public query 与 expected citation evidence，精确绑定一份不可变知识快照和 `workflow.rag.lexical-ngram-dev.v1` profile，使用运行时同一 lexical ranker 生成 metadata-only 质量报告，并以确定性指标判断当前快照是否满足开发测试态检索基线。

该功能独立于已经完成的 RAG Retrieval、run v3 和 Regression Review，不是 RAG 批次 D。它评审“某份知识快照能否为一组已审查问题召回预期证据”，不执行 Workflow、不调用 Gateway、不生成回答，也不自动创建 baseline、Evaluation Case 或 release decision。

## 用户路径

1. 维护者提交一份 synthetic-public `workflow_rag_snapshot.v1` fixture，正文只包含人工编写的测试知识，不复制真实 workspace 内部内容。
2. 维护者为该快照提交版本化 dataset，逐样本写入 query、`evidence_required / no_evidence` 预期、expected citation refs、必须命中的 official refs 和 `top_k`。
3. CLI 严格读取 snapshot 与 dataset，重新校验 fragment digest、snapshot digest、dataset digest 和 lexical profile digest。
4. 评审器对每个 query 恰好执行一次既有 deterministic lexical ranker；不调用 Gateway，不进行 retry、fallback、query expansion、embedding 或 rerank。
5. CLI 生成不含原始 query 和 fragment 正文的 `workflow_rag_quality_review.v1`，展示 selected refs / ranks / digests、样本判定、聚合指标和知识质量 finding。
6. `--check` 模式重新生成报告并与 committed report 精确比较；算法、数据、digest、阈值或报告漂移都会失败关闭。

## 权威来源与数据边界

| 资产 | 权威内容 | 允许持有 | 不得持有 |
| --- | --- | --- | --- |
| snapshot fixture | 本次离线评测使用的不可变测试知识 | synthetic-public fragment 正文、来源元数据、digest | workspace 内部内容、credential、线上抓取结果 |
| evaluation dataset | 人工审查过的问题与引用预期 | query 正文、expected citation refs、阈值、review metadata | 模型生成的 query、客户端排名、答案正文、生产声明 |
| quality report | 可提交的确定性审查结果 | query digest / bytes、fragment refs / ranks / digests、指标、finding | 原始 query、fragment 正文、excerpt、prompt、answer、credential |

dataset 与 snapshot fixture 是显式评测资产，因此可以保存 synthetic-public query 和正文；report、日志与普通命令输出继续 metadata-only。首版不接受 `workspace_internal` fixture，避免把运行时授权内容复制到仓库。

## 版本化契约

新增两份契约：

- `workflow_rag_evaluation_dataset.v1`：dataset id / version / digest、精确 snapshot 与 profile binding、阈值、review metadata 和 1–200 个样本。
- `workflow_rag_quality_review.v1`：精确 binding、聚合指标、逐样本 metadata-only evidence、知识覆盖摘要和稳定 findings。

dataset digest 对除自身 digest 字段外的完整 canonical dataset 求 SHA-256，包含人工 review metadata；任何 query、预期引用、阈值或绑定变化都必须产生新 digest。snapshot digest 和 profile digest 继续使用现有运行时算法，不另造第二套口径。

## 样本语义

每个样本固定包含：

- `sample_id`：dataset 版本内唯一稳定短键；
- `query_text`：synthetic-public UTF-8 文本，去除首尾空白后必须保持不变，最多 4 KiB；
- `expectation`：`evidence_required` 或 `no_evidence`；
- `expected_citation_refs`：人工认可为可支撑回答的 fragment refs；
- `required_official_refs`：必须命中的 official refs，且必须是 expected refs 子集；
- `top_k`：1–8，显式固定，不继承运行环境默认值；
- `review_note`：说明预期证据边界，不进入质量报告。

`evidence_required` 至少有一个 expected citation ref；`no_evidence` 两组 refs 都必须为空，并要求 lexical ranker 返回 `workflow_rag_no_evidence`。所有引用必须存在于精确 snapshot；required official ref 必须指向 `is_official=true` 的 fragment。

## 确定性指标

首版只评估 lexical retrieval 和证据可用性：

- `hit_at_k`：positive 样本中至少命中一个 expected citation ref 的比例；
- `expected_recall_at_k`：所有 positive 样本 expected refs 的微平均召回率；
- `required_official_recall_at_k`：所有 required official refs 的微平均召回率；
- `mean_reciprocal_rank`：每个 positive 样本第一个 expected ref 的 reciprocal rank 平均值；
- `no_evidence_accuracy`：negative 样本正确返回 no evidence 的比例；
- `sample_pass_rate`：满足各自完整预期的样本比例。

指标保留六位小数，阈值写入 dataset 并进入 dataset digest。positive 样本只有全部 expected refs 与 required official refs 都在本次 selected set 中才通过；negative 样本只有零 selected 且稳定失败码为 `workflow_rag_no_evidence` 才通过。延迟不进入离线质量阈值。

## 知识质量审查

评审器同时产生不依赖模型的 snapshot findings：

- content digest 重复：`duplicate_fragment_content`，`review_required`；
- 没有 official fragment：`official_evidence_absent`，`review_required`；
- fragment 未被任何 positive 样本列为 expected evidence：`fragment_expectation_uncovered`，`info`；
- expected refs 未覆盖全部 official fragments：`official_expectation_coverage_partial`，`info`；
- source type 分布和 expected coverage 作为摘要返回，不把论坛 / FAQ 自动判定为错误。

指标低于阈值或任一样本失败时报告状态为 `failed`；指标通过但存在 `review_required` finding 时为 `needs_review`；其余为 `passed`。CLI 的 `--check` 只接受 `passed` 且与 committed report 完全一致。

## CLI 与实现职责

稳定入口为：

```bash
cd services/platform
go run ./cmd/radishmind-workflow-rag-eval \
  --snapshot ../../datasets/eval/workflow-rag/snapshots/dev_core_v1.json \
  --dataset ../../datasets/eval/workflow-rag/datasets/dev_core_v1.json \
  --output ../../datasets/eval/workflow-rag/reports/dev_core_v1.review.json
```

追加 `--check` 时不改文件，只复算并比较 output。命令只输出 dataset id / version / digest、snapshot / profile digest、样本计数、状态和报告路径，不输出 query 或 fragment 正文。

领域实现放在独立 `workflow_rag_quality_review.go`，不继续堆入 `workflow_rag_snapshot.go`。CLI 只负责 strict file I/O、参数和退出码；检索、digest、指标、finding 与 report 由领域服务负责。首版继续复用 `RankWorkflowRAGFragments`、snapshot validator 和 profile digest。

## 数据布局

```text
datasets/eval/workflow-rag/
  README.md
  snapshots/<short_key>.json
  datasets/<short_key>.json
  reports/<short_key>.review.json
```

文件名使用短键，长语义保存在 JSON 内。starter dataset 至少覆盖 Latin、CJK、official boost、多个 expected refs、tie-break 和 no-evidence；不得只放一个必过样本。

## 验收

- strict JSON：未知字段、尾随 JSON、schema / id / digest / binding 漂移全部拒绝；
- dataset：重复 sample、非法 expectation、引用不存在、official 子集错误、query 为空 / 超限 / 含禁止材料拒绝；
- evaluator：positive / negative、top-k、hit / recall / MRR、official recall、threshold failure 和 finding 状态；
- metadata-only：report 与命令输出不含 query、fragment 正文、excerpt、prompt 或 answer；
- deterministic：相同输入字节级同报告，`--check` 检出 committed report 漂移；
- compatibility：现有 lexical provider、RAG execution、run v3、Regression Review、executor v0 和 HTTP Tool v2 测试不回归；
- repository：Go 精准 / 全包、CLI generate / check、schema validation、`git diff --check`、仓库 fast / full。

## 已完成实现

本专题只使用一张任务卡，两个实现批次均已完成：

1. 批次 A 已完成两份契约、独立领域 evaluator、strict loader、CLI、指标 / finding 测试和 metadata-only 输出；每个样本只调用一次既有 `RankWorkflowRAGFragments`。
2. 批次 B 已提交 6 个 synthetic-public fragment、7 个评测样本和确定性 `passed` report，覆盖 Latin、CJK、official boost、多个 expected refs、稳定 tie-break、补充来源和 no-evidence；仓库契约检查已接入 snapshot / dataset / report schema 与敏感内容扫描。

starter report 的六项指标均为 `1`，3 个 official fragment 与 6 个总 fragment 全部进入人工 expected evidence 覆盖，当前无 finding。报告仅含 query digest / bytes、selected refs / ranks / digests、source types、指标和状态，不含 query、fragment 正文、excerpt 或 review note。

本专题现已关闭。首版没有新增数据库 migration、HTTP route、Web 页面、后台 job 或第二张 task card；下一产品设计可评审“RAG 评测数据集应用资源化与候选快照审查 v1”，但必须先固定 query 数据分类、权限、版本 CAS、store、候选 snapshot binding 和 Web 边界，不能从本专题自动获得实现准入。

## 边界评审结论

1. 复用运行时 ranker 和 digest validator，禁止 Python / Web 复制排名算法。
2. committed query / fragment 只允许 synthetic-public；真实 workspace knowledge 不进入仓库。
3. report 和日志保持 metadata-only，expected citation evidence 不被解释为答案正确性或生产 relevance 声明。
4. v1 只做同步离线复算，不创建 run、不调用 Gateway、不更新 Evaluation Case / Suite。
5. 新 schema / 数据格式由一张高风险实现任务卡承接；不新增同层 readiness / checker 文档链。

## 停止线

- 不启用 crawler、connector、在线搜索、文件扫描或跨应用数据导入。
- 不接 embedding、vector database、reranker、hybrid search 或外部 provider。
- 不执行 Workflow / Gateway，不生成或评测 answer，不保存模型响应。
- 不自动修改 snapshot、dataset、baseline、case、suite 或 release decision。
- 不提交 workspace_internal、credential、真实用户 query 或未经授权的上游正文。
- 不新增生产 API、生产认证、数据库表、Web 写面、后台调度、自动 release 或 production relevance 声明。

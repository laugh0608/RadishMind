# Workflow 不可变版本晋级与受控运行绑定（开发 / 测试态）v1 实施任务卡

更新时间：2026-07-19

状态：`workflow_definition_version_promotion_controlled_runtime_binding_dev_test_v1_batch_a_ready_for_implementation`

## 目标与准入结论

按[功能设计](../features/workflow/workflow-definition-version-promotion-controlled-runtime-binding-dev-test-v1.md)交付“精确 Saved Draft → 不可变 candidate → 人工 review → definition version → 人工 activation → definition-bound run v5 → history / comparison / evaluation”的开发测试态产品路径。

设计、资源 owner、权限、执行来源、兼容策略和生产停止线已经冻结，可以进入批次 A。不得给 Saved Draft 增加可变 publish 状态，不得复用 application publish candidate，不得把 definition id 写入旧 `draft_id`，也不得通过 activation 绕开 HTTP Tool / RAG 独立 authority。

## 前置基线

- 实现准入基线为 `fce60c5c`，已完成 Saved Draft 三种 store、executor v0、run history / comparison / evaluation、HTTP Tool、Workflow RAG、Application RAG 与 Application Operations。
- Saved Draft 是可编辑来源，definition candidate / version repository 是新 owner；两者只通过 exact id / version / digest provenance 关联。
- 既有 Workflow Definition summary 是历史读模型，不作为写入 authority；批次 B 才允许切换 live projection。
- v0–v4 contract、`workflow_draft` 与 `application_configuration_draft` execution source 必须兼容保留。
- 默认 gates 关闭，所有管理动作使用 verified actor 与精确 scopes，所有输出 metadata-only / sanitized。

## 批次 A：strict contract、memory repository 与人工审查 API

状态：`ready_for_implementation`。

### 允许实现

- 新增 `workflow_definition_release_candidate.v1`、candidate decision、`workflow_definition_version.v1`、activation pointer / event / audit strict JSON Schema 与 Go codec。
- 新增 canonical definition digest；服务端从 exact Saved Draft 重读并派生 candidate snapshot、execution profile、risk 和 blocker。
- 新增 application-scoped memory repository，candidate / review / version / activation 共用 owner lock。
- candidate approve 在同一原子操作中追加 review 并物化一次 immutable version；reject 不物化 version。
- activation `activate / replace / deactivate` 使用 expected pointer version，原子写 current projection、event 和 audit。
- 注册默认关闭的 dev/test gate、五类 management scopes 和 candidate / version / activation 管理 API。

### 必须证明

- source draft missing / version drift / digest drift / scope mismatch / invalid / blocked 全部在 candidate write 前失败。
- candidate 与 version snapshot 不包含 secret、input、answer、prompt、provider raw payload 或 browser-only state。
- review / activation 并发相同 expected version 只有一个成功；current projection、append-only evidence 和 version materialization 无 partial write。
- approve 不自动 activate；activate 不执行节点；deactivate 不回退旧版本。
- route、method、strict JSON、verified actor、tenant / workspace / application / owner 和 scope grant 均失败关闭。
- Saved Draft、executor v0、run v0–v4、HTTP Tool、RAG 与 application publish 回归通过。

### 批次 A 停止线

- 不创建 SQLite / PostgreSQL migration、run v5、definition execution、Web、launcher 或真实浏览器路径。
- 不修改 Workflow Definition live read projection；不删除 fake-store offline sample。
- 不打开 production auth、public API、automatic approval / activation、schedule、retry、replay、writeback 或 agent loop。

完成后状态推进为 `workflow_definition_version_promotion_controlled_runtime_binding_dev_test_v1_batch_b_ready_for_implementation`。

## 批次 B：durable repository 与正式 read projection

状态：`blocked_by_batch_a`。

- 在 workflow shared SQLite / PostgreSQL migration family 中追加 candidate、decision、version、activation pointer、event 和 audit 表及 schema marker。
- memory / SQLite / PostgreSQL 共用同一 domain contract，不新增 DSN、pool、database file、selector 或 fallback。
- 覆盖 migration / rollback / reapply、运行角色、事务 CAS、append-only、restart、corruption、checksum 与 no-fallback。
- live Workflow Definition summary 从新 repository 投影；offline sample 继续显式隔离，禁止结果混合。
- 双数据库语义与 HTTP API 完成后推进批次 C。

## 批次 C：definition-bound executor 与 run v5

状态：`blocked_by_batch_b`。

- 物化 `workflow_run_record.v5` 和 `workflow_definition_executor.v1` profile。
- 执行前重读 activation pointer、definition version、digest、application lifecycle 和 profile eligibility。
- 完整复用 executor v0 的受支持拓扑、预算、取消、Gateway 和 diagnostics，不复制第二套图执行算法。
- run source 固定 `workflow_definition`；source draft 仅为 provenance。
- memory / SQLite / PostgreSQL run store、history / detail / filter / stale reconciliation 接受 v5。
- Comparison / Evaluation / Baseline / Suite 只读消费兼容 v5，不重新执行。
- v0–v4、旧 draft execution、HTTP Tool、Workflow RAG 与 Application RAG 保持行为兼容。

## 批次 D：Web、连续链与专题收口

状态：`blocked_by_batch_c`。

- Draft Designer 接入 candidate 创建与 source definition provenance。
- 新增 candidate review、version history、activation 管理与 definition-bound run 页面。
- Run History / Comparison / Evaluation / Application Operations 识别 v5 和 definition source。
- 应用切换清除全部易失状态；offline 零请求；strict consumer 拒绝 schema / scope / sensitive drift。
- 完成 SQLite / PostgreSQL 连续链、重启、冲突、停用阻断、派生新草案和真实浏览器复验。
- 同步正式文档和周志，关闭专题；不派生第二张任务卡。

## 验证矩阵

- Go：contract / codec、canonical digest、domain state machine、CAS / race、HTTP auth / strict JSON、memory / SQLite / PostgreSQL、migration、authority reload、executor、history、comparison、evaluation、no-fallback。
- Web：offline、candidate / review / version / activation / run、application switch、late response、strict schema、secret guard、production build。
- 连续链：draft save → candidate → approve → version → activate → run v5 → history → comparison / evaluation → derive new draft → replacement activation → deactivate blocked。
- 停止线：candidate / review / activation provider、tool、retrieval、confirmation、business write、replay 均为 0；v5 只允许受支持 profile 的计划内 provider calls。
- 仓库：每批 `git diff --check`、精准测试与 `./scripts/check-repo.sh --fast`；contract / schema / migration / execution /阶段真相源变化补完整 `./scripts/check-repo.sh`。

## 总停止线

- 不自动 review、activation、publish、release 或执行。
- 不把开发测试态 definition version 声明为 production workflow、公开 API 或上层项目正式真相。
- 不启用 schedule、background、retry / fallback、replay / resume、agent loop、business writeback、production auth / secret / repository、quota 或 billing。
- 不创建平行 Saved Draft、run、Gateway、HTTP Tool、RAG、application publish 或 audit 基础设施。

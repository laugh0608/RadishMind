# 工作区 Workflow 模板目录、审查与受控派生（开发 / 测试态）v1 实施任务卡

更新时间：2026-08-27

- 任务 ID：`workspace-workflow-template-catalog-review-controlled-derivation-dev-test-v1`
- 状态：`batch_b_completed_batch_c_pencil_ready`
- 功能设计：[工作区 Workflow 模板目录、审查与受控派生（开发 / 测试态）v1](../features/workflow/workspace-workflow-template-catalog-review-controlled-derivation-dev-test-v1.md)

## 准入结论

项目所有者已批准该长期功能目标与首版方向。本卡是唯一跨模块高风险实施入口，负责 Definition exact source、工作区模板 catalog owner、人工 review / listing、跨应用目标资格与 Saved Draft 原子派生；不再创建平行 readiness、schema-only 或 UI-only 任务卡。

项目所有者已明确批准并完成批次 A，随后批准并完成批次 B。strict contract、memory catalog owner、默认关闭的开发测试态 HTTP 与 Saved Draft `derivation_v2` 已落地；SQLite / PostgreSQL durable repository、共享 backend factory 和 `0005` migration 已通过双数据库验证。当前停在批次 C Pencil 明确授权前，不自动创建 Pencil 或进入 React。

## 完成目标

同一 workspace 的 Builder 能把一个 approved immutable Definition Version 提交为模板候选，经人工 review 和独立 listing pointer 上架后，由另一活动应用受控派生为独立 Saved Draft v1，并通过现有 Draft Designer、Validate、Save 与 Definition candidate handoff 继续工作。

## 不可变边界

1. Definition Version 是模板内容的唯一来源；catalog 只保存 exact ref / digest、分发元数据、资格摘要和决定，不保存第二份可变 graph。
2. Template candidate approve 不自动 list；list / replace / unlist 不自动派生；derive 不自动创建 Definition candidate、activate 或 run。
3. 派生只调用既有 Saved Draft owner 创建独立记录；不创建 template draft table、平行 repository 或跨 owner fallback。
4. 首版只允许 `workflow_definition_executor_v1` 的 `prompt | llm | condition | output`；HTTP Tool、RAG、code、sandbox、agent loop、writeback、replay、resume 和 schedule 全部失败关闭。
5. Provider / Profile / Model ref 必须在目标应用重新验证；不自动替换、不回退默认绑定。
6. 所有操作限定 exact tenant / workspace；目标应用必须处于同一 workspace 且 active。
7. v1 复用现有 Definition / Draft 组合 permissions，不新增 Marketplace permission，不静默改写 local role assignments。
8. memory、SQLite、PostgreSQL 复用既有 workflow backend mode、pool、selector 和 migration family，不新增 DSN 或数据库文件。
9. 不持久化 credential、endpoint、header、raw provider payload、运行 input / answer、cookie、确认或业务写回材料。

## 批次 A：strict contract、memory owner 与 HTTP

状态：`completed`。

### 允许实现

- 新增 template candidate / decision、version、lineage / listing pointer、event / audit strict schema 和 codec。
- 将 Saved Draft provenance 扩展为兼容保留 `derivation_v1` 的 strict `derivation_v2` union。
- 实现 template canonical digest、Definition exact authority reload、portable profile validator 与 target binding validator。
- 实现 workspace-scoped memory catalog repository、candidate review CAS、approve materialization、listing pointer CAS 与 append-only event / audit。
- 注册默认关闭的 strict dev/test routes，复用 verified identity、membership provider、Application / Definition / Draft owners 与现有组合 permissions。
- derive 在所有预检通过后通过 Saved Draft owner 原子创建 v1 draft；失败不留下 catalog 或 draft partial write。

### 必须证明

- source missing / version drift / digest drift / application archived / scope mismatch 在 candidate write 前失败。
- forbidden node / capability / material 在 candidate write 前失败，人工 review 不能解锁。
- approve 物化 version 与 review append 原子；相同 expected review version 并发最多一个成功。
- listing pointer / event / audit 原子；相同 expected pointer version 并发最多一个成功。
- derive 重读 listed pointer、template version、Definition、target Application、membership 与 binding；任何漂移在 draft write 前失败。
- candidate、review、listing、derive 的 provider / tool / retrieval / confirmation / run / evaluation / business write 均为 0。
- 既有 Saved Draft `derivation_v1`、Definition release、HTTP Tool、RAG、application publish 和 role catalog 回归通过。

### 完成证据

- `contracts/` 已冻结 candidate / decision、version、lineage、listing event、audit 与 Saved Draft template `derivation_v2` 七份 strict schema；Go codec 拒绝 unknown / duplicate / invalid transition / invalid provenance union。
- memory catalog owner 已实现 candidate create、单次 review CAS、approve + immutable version materialization、listing pointer + event + audit CAS、strict stored-record validation 和 workspace / filter / snapshot-bound cursor。
- 十条默认关闭的开发测试态 route 已注册，复用 verified identity、active membership 与 Definition / Draft 组合 permissions；`RADISHMIND_WORKFLOW_TEMPLATE_CATALOG_DEV` 还要求 Application Catalog、Definition Release 与 Saved Draft authority gates 同时打开。
- 派生会重读 listed pointer、template version、source Application / Definition、target active Application 与 exact provider profile ref 形状，通过 Saved Draft owner 单次创建 v1 草案；`derivation_v1` 兼容保留，二者严格互斥。
- source missing / version / digest / scope drift、archived Application、forbidden node / capability / material、target binding、draft id 冲突、review / listing 并发单胜者、十条 HTTP、权限和默认关闭门禁均有精准测试；相邻 Saved Draft / Definition / Application 回归、review / listing race、`go vet` 与完整 `go test ./internal/config ./internal/httpapi` 已通过。
- 本批没有新增 SQLite / PostgreSQL migration、Pencil、React、CSS、launcher、长驻服务、真实浏览器、Provider / Tool / Retrieval / Confirmation / Run / Evaluation 或业务写入。

### 批次 A 停止线

- 不创建 SQLite / PostgreSQL migration、Pencil、React、CSS、launcher、产品服务或真实浏览器记录。
- 不增加独立 permission、公开 API、跨 workspace sharing 或 production gate。
- 不实现 target binding replacement、template search ranking、ratings、download count 或 recommendation。

批次 A 完成后只能推进为 `batch_a_completed_batch_b_ready`，不得自动进入批次 B。

## 批次 B：SQLite / PostgreSQL durable repository

状态：`completed`。

- 在既有 workflow shared SQLite / PostgreSQL migration family 中追加 candidate、decision、version、lineage pointer、listing event / audit 与必要 cursor 索引。
- memory / SQLite / PostgreSQL 共用同一 domain contract，不新增 selector、pool、DSN、database file 或 fallback。
- review materialization 与 listing pointer / event / audit 分别使用单事务；runtime role 不得 DDL。
- 覆盖 migration / rollback / reapply、marker、checksum、keyset cursor、CAS、restart、corruption、pool reconnect 与 no-fallback。
- derive 继续调用 Saved Draft owner；同数据库也不允许绕过 domain owner 直接插入 draft table。

当前证据：

- SQLite `0004 → 0005` 升级、失败 migration 整体回滚、reapply、marker / checksum、review / listing 事务回滚、keyset cursor、并发 CAS、restart、corruption、workspace isolation 与 no-fallback 已执行通过。
- PostgreSQL 17 disposable integration 已执行通过，覆盖受限 runtime role、review / listing 原子回滚、并发 CAS、restart / reconnect、history append-only、corruption、closed-pool no-fallback、reviewed down 与 reapply；完整 `postgres_integration` suite 同时通过。
- `go vet`、完整 Platform Go 回归、template 精准测试与 race、仓库 fast baseline 已通过；未创建 Pencil、React、CSS、launcher、产品服务或真实浏览器记录。

完成后只能推进为 `batch_b_completed_batch_c_pencil_ready`。

## 批次 C：完整 Pencil 与人工批准

状态：`approval_required`。

- 以已实现功能事实完成 Catalog、Candidate Review、Version / Listing、Derive 的 Desktop 与不能直接推导的 Narrow。
- 覆盖 pending / rejected / approved-unlisted / listed / replace conflict / unlist danger / target binding unavailable / store unavailable。
- 复用现有 Workflow Workbench 与语义 token，不建立 S11、不复制 Saved Draft Library 或 Definition Review owner。
- Pencil 根画板、关键状态和 Decision 记录必须由项目所有者人工批准；未批准不得进入 React。

完成后只能推进为 `batch_c_pencil_approved_batch_d_ready`。

## 批次 D：React strict consumer

状态：`blocked_by_batch_c`。

- 在现有 Workflow 产品面接入单一 template catalog consumer，覆盖 catalog / candidate / version / listing / derive。
- strict parser 拒绝 unknown field、scope drift、invalid digest / cursor、duplicate record 与敏感字段。
- workspace / application / actor 切换以 generation + abort 清理 cursor、selection、target、confirmation 与迟到响应。
- 派生成功只交接 exact draft id / version 到既有 Draft Designer；不在前端复制 Definition payload 或调用多次 mutation 模拟原子性。
- offline source 零请求，target / confirmation 不写 URL、Storage、IndexedDB、Cache 或跨标签 payload。

完成门禁：Web 精准 / 全量测试、production build、Platform 相邻测试、源码敏感扫描、`git diff --check`、仓库 fast / full；本批不冒充真实数据库和浏览器产品链。

完成后只能推进为 `batch_d_react_completed_batch_e_ready`。

## 批次 E：双数据库产品连续链与收口

状态：`blocked_by_batch_d`。

- SQLite：exact Definition → candidate → approve → list → target application derive → Draft exact open / validate / save → service restart restore。
- PostgreSQL configured Server：migration / runtime role、同构链、hard failure no-fallback、reconnect 后 exact reload 与历史不可变。
- 双标签制造 review / listing stale CAS；workspace / application 切换不得串入旧 template、target 或 draft handoff。
- 浏览器覆盖 `1440×900`、`720×900`、`390×844`，复核横向溢出、console / network、URL、Storage、cookie 与响应 / 数据库敏感材料。
- 证明模板动作没有 provider、tool、retrieval、confirmation、run、evaluation、writeback、retry / fallback 或 replay。
- 同步功能专题、功能入口、Workflow 产品面、任务卡入口、当前焦点、路线图、能力矩阵和周志；清理所有测试服务、容器、数据库与临时文件。

完成锚点：`workspace_workflow_template_catalog_review_controlled_derivation_dev_test_v1_completed`。完成后关闭本卡，不派生批次 F、公开 Marketplace、跨 workspace、推荐、计费或 production 续批。

## 验证矩阵

- Go：schema / codec、digest、source reload、portability、target binding、memory / SQLite / PostgreSQL、review / listing CAS、cursor、strict HTTP、auth、draft create、migration、no-fallback、race。
- Web：offline、strict parsing、catalog / review / listing / derive、scope generation、late response、exact Draft handoff、secret guard、production build。
- 产品链：Definition approve → template candidate → template approve → list → target derive → Draft validate / save → restart → Definition candidate handoff。
- 停止线：所有模板管理与派生操作的执行副作用计数为 0，源 Definition、模板版本和已有草案保持不可变。
- 仓库：每批精准测试、`git diff --check`、`./scripts/check-repo.sh --fast`；schema / migration /阶段真相源和最终关闭补完整 `./scripts/check-repo.sh`。

## 当前下一步

批次 B 已完成并停在 `batch_b_completed_batch_c_pencil_ready`。下一步必须由项目所有者明确批准批次 C，才允许反查已实现事实并完成 Catalog、Candidate Review、Version / Listing 与 Derive 的完整 Pencil；未获批准前不修改 Pencil，不进入 React、产品服务或真实浏览器。

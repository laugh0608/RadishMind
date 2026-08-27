# 工作区 Workflow 模板目录、审查与受控派生（开发 / 测试态）v1

更新时间：2026-08-27

状态：`workspace_workflow_template_catalog_review_controlled_derivation_dev_test_v1_batch_a_completed_batch_b_ready`

## 功能定位

本专题让同一工作区内的 Workflow Builder 把一个已经人工批准的不可变 `Workflow Definition Version` 提交为团队模板候选，经独立人工审查和显式上架后，由其他 Builder 从当前上架的精确模板版本受控派生新的 Saved Workflow Draft。

它把现有“个人草案派生”和“应用内 Definition 晋级”向工作区内部复用扩展，但不建立公开 Marketplace，不复制 Definition、Saved Draft、运行、评测、权限或应用真相源。模板目录只拥有工作区级模板候选、不可变模板版本、上架指针与对应 append-only 决定；模板内容来源和运行资格仍由既有 Definition 与 Draft owner 重新判定。

本专题只承载内部开发者预览下的开发 / 测试态能力。当前设计已获项目所有者批准，并已沿[唯一高风险任务卡](../../task-cards/workspace-workflow-template-catalog-review-controlled-derivation-dev-test-v1-plan.md)完成批次 A 的 strict contract、memory owner 与 HTTP；当前停在批次 B 授权前，不自动创建 durable repository、Pencil、React 或产品连续链。

## 用户与真实任务

- `Workspace Builder`：希望把已经验证过的 Workflow 作为团队模板发布，也希望从团队模板快速建立独立草案，而不是复制 JSON 或覆盖源记录。
- `Workspace Reviewer`：审查模板来源、可移植资格、风险、目标用户说明和敏感内容边界，决定模板版本是否可以物化。
- `Workspace Operator`：显式上架、替换或下架一个精确模板版本，并能审查 pointer CAS、事件和失败原因。
- `Platform Maintainer`：保证模板目录复用唯一 Definition / Draft / membership owner，所有跨应用派生失败关闭且没有执行副作用。

## 已有事实与准入结论

- Saved Workflow Draft 已具备 memory、SQLite、PostgreSQL、严格保存 / 读取 / 校验、版本 CAS、生命周期、修订历史和同应用精确草案派生。
- Workflow Definition 已具备不可变 candidate / version、人工 review、activation、exact digest、三种 repository 与受控运行来源。
- 本地账户、四个 canonical workspace role、membership decision 与 Workflow mutation authorization 已形成开发测试态真实 owner。
- Definition、Run、Comparison、Evaluation 和 Application Operations 已能提供精确来源和可复验产品证据；模板目录不需要复制这些记录。
- 当前没有真实 Radish、production auth、production secret、公开发布或跨工作区共享前置，因此首版只打开同一 workspace 内部目录。

准入结论：该功能有真实团队复用闭环、明确 canonical owner、充分本仓库事实基础、零外部依赖，并能在 memory / SQLite / PostgreSQL 与真实浏览器形成端到端证据。它不是已关闭 Saved Draft 或 Definition 专题的同层小切片，允许作为新的长期功能目标进入设计和高风险实施准备。

## 产品闭环

1. Builder 从同一 workspace 下选择一个已批准的精确 `Workflow Definition Version`，提交模板标题、摘要和有限标签。
2. 服务端按 exact definition id / version / digest 重读不可变来源，重新检查 workspace、source application、execution profile、图结构和禁止能力；客户端不能提交 Definition payload、digest、risk 或 eligibility。
3. 服务端创建不可变模板候选。候选只保存来源引用、目录元数据、服务端派生的资格摘要、request / audit ref 和 append-only review 状态。
4. Reviewer 读取候选、精确 Definition evidence、当前同 lineage 模板版本差异和 blocker，使用 expected review version 追加 `approve | reject | request_changes | withdraw` 决定。
5. `approve` 只物化一个不可变模板版本，不自动上架、不创建草案、不激活 Definition、不执行 Workflow。
6. Operator 使用当前上架 pointer version 执行 `list | replace | unlist`。每个 `template_id` 同时最多一个上架版本；pointer、event 与 audit 在同一 owner 原子提交。
7. Builder 浏览当前 workspace 的上架模板目录，查看脱敏来源、版本、能力、目标绑定资格与使用说明，选择一个精确模板版本和目标活动应用。
8. 派生服务重新读取当前上架 pointer、模板版本、源 Definition、目标应用、membership 与目标 Provider / Profile / Model 可用性。任何漂移、下架、归档、scope 或 store 失败均在写草案前失败关闭。
9. 成功派生只调用既有 Saved Draft owner 创建一个独立版本 `1` 草案，并写入严格 `derivation_v2` provenance；源 Definition、模板版本和既有草案保持不变。
10. Web 精确打开新草案，用户继续通过现有 Draft Designer、Validate、Save、Review Handoff 和 Definition candidate 流程；模板上架不替代后续人工审查或 activation。

## 资源与 owner 边界

| 资源 | 唯一 owner | 本专题允许行为 | 明确禁止 |
| --- | --- | --- | --- |
| Saved Workflow Draft | 既有 Saved Draft owner | 派生成功后原子创建独立 v1 草案并保存模板 provenance | 模板目录直接改写、覆盖、归档或激活草案 |
| Workflow Definition Version | 既有 Definition owner | exact read、digest / profile / graph authority reload | 复制为可变模板正文、修改 Definition 或绕过 review |
| Template Candidate / Version / Listing | 新 workspace template catalog owner | 不可变候选、append-only review、不可变版本、CAS 上架 pointer / event | 保存 credential、运行输入输出、确认、Run 或业务数据 |
| Application Catalog | 既有 Application owner | 重读 source / target application lifecycle 与 scope | 模板动作创建、启用、归档或更新应用 |
| Provider / Profile / Model | 既有 Admin / Gateway owner | 派生前验证引用在目标应用下仍可解释 | 复制 credential、endpoint、raw provider config 或自动 fallback |
| Membership / Permission | 既有 workspace membership provider | 对每个 read / review / listing / derive 重新授权 | 模板记录缓存成员关系或客户端 grants |

模板版本是 Definition 的工作区分发授权记录，不是第二份 Workflow 内容真相源。它只保存 exact `source_definition_id + source_definition_version + source_definition_digest`、模板元数据和服务端派生摘要；详情与派生始终重新读取不可变 Definition。源记录不可用、digest 不匹配或 scope 漂移时不回退候选快照、旧版本或浏览器缓存。

## 首版可移植资格

首版模板只接受 `workflow_definition_executor_v1`，图节点限定为：

- `prompt`
- `llm`
- `condition`
- `output`

以下任一能力使候选在写入前失败，不能靠人工 review 解锁：

- HTTP Tool、RAG、code、sandbox、agent loop、business writeback、replay、resume、schedule 或 background execution；
- confirmation-specific、tool-specific 或 retrieval-specific authority；
- 非法拓扑、unsupported contract、损坏 digest 或未知 schema；
- credential、header、endpoint、raw provider payload、运行 input / answer 或其它禁止材料。

Provider / Profile / Model 只允许保存既有脱敏引用。跨应用派生前必须由服务端确认同一引用在目标活动应用中仍可用；不匹配时返回 `workflow_template_target_binding_unavailable`，不自动替换模型、不回退默认 Profile，也不创建 blocked 草案。目标绑定替换若未来需要，应由独立功能版本设计。

## 数据模型

### Template Lineage 与 Listing Pointer

`workspace_workflow_template_lineage.v1` 至少包含：

- `template_id`、`tenant_ref`、`workspace_id`；
- current listed version ref、pointer version 与 lifecycle；
- created / updated actor、request / audit ref 与时间。

lineage 不保存 Definition graph。listing pointer lifecycle 只允许 `unlisted | listed`，决定只允许 `list | replace | unlist`，并使用 `expected_pointer_version`。

### Template Candidate

`workspace_workflow_template_candidate.v1` 至少包含：

- candidate id、optional template lineage ref、candidate state 与 review version；
- exact source application、definition id / version / digest；
- server-owned execution profile、node kinds、risk / portability summary 和 ordered blockers；
- 用户提交的 sanitized title、summary、usage notes 与最多八个 normalized labels；
- created actor / time、request / audit ref 和 append-only review decisions。

客户端不得提交 source digest、graph、profile、eligibility、candidate state、review actor 或 audit 字段。候选创建后来源和目录元数据均不可修改；修改必须创建新候选。

### Template Version

`workspace_workflow_template_version.v1` 只由 approved candidate 物化，至少包含：

- template id / version / template digest；
- exact candidate id / review version；
- exact source Definition ref / digest；
- immutable title、summary、usage notes、labels 与 portability summary；
- created actor / time、request / audit ref。

版本不提供 update 或 delete。template digest 对 canonical distribution metadata、source ref / digest 和 portability summary 计算，不包含 actor、request、audit 或时间字段。

### Saved Draft Provenance

既有 `derivation_v1` 继续只表达 saved-draft direct parent。模板派生新增严格 union `derivation_v2`：

```json
{
  "version": 2,
  "source_kind": "workspace_workflow_template",
  "template_id": "wftpl_example",
  "template_version": 1,
  "template_digest": "sha256:...",
  "source_definition_id": "wf_example",
  "source_definition_version": 3,
  "source_definition_digest": "sha256:..."
}
```

该 provenance 只保存直接来源，不复制候选 review、listing event、actor、membership、request body、audit body、Run、Evaluation、credential 或 provider payload。

## 权限与身份

首版不新增通用 Marketplace permission，也不静默修改已冻结的 local role assignment grants。模板操作是既有 Definition 与 Draft 权限的组合：

- 目录 / candidate / version 读取：`workflow_definitions:read`；
- candidate 创建：`workflow_definitions:read + workflow_definitions:write`；
- candidate review：`workflow_definitions:read + workflow_definitions:review`；
- `list | replace | unlist`：`workflow_definitions:read + workflow_definitions:activate`；
- 派生：`workflow_definitions:read + workflow_drafts:write`。

所有操作还必须通过 verified identity、active workspace membership 和 exact tenant / workspace / source / target application scope。未来公开发布、跨 workspace 共享、收费或第三方安装必须建立独立 permission 与功能设计，不能复用本 v1 权限外推。

## 开发测试态 API 边界

批次 A 已新增以下默认关闭的 strict route：

- `POST /v1/user-workspace/workflow-template-candidates`
- `GET /v1/user-workspace/workflow-template-candidates`
- `GET /v1/user-workspace/workflow-template-candidates/{candidate_id}`
- `POST /v1/user-workspace/workflow-template-candidates/{candidate_id}/decisions`
- `GET /v1/user-workspace/workflow-templates`
- `GET /v1/user-workspace/workflow-templates/{template_id}`
- `GET /v1/user-workspace/workflow-templates/{template_id}/versions`
- `GET /v1/user-workspace/workflow-templates/{template_id}/versions/{version}`
- `POST /v1/user-workspace/workflow-templates/{template_id}/listing-decisions`
- `POST /v1/user-workspace/workflow-templates/{template_id}/derivations`

目录、候选和版本列表使用 workspace-scoped、filter-bound、snapshot-bound cursor，排序固定为 `updated_at DESC + stable id DESC`。派生请求只接受 exact template version、expected pointer version、target application id、new draft id / name 与显式确认；Definition body、draft body、digest、actor、grant 和 audit 字段一律由客户端禁止提交。

所有 route 只在显式 `RADISHMIND_WORKFLOW_TEMPLATE_CATALOG_DEV=1` 下注册或启用，默认离线 Web 零请求。unknown mode、disabled mode、store unavailable、schema mismatch 或 upstream owner failure 均不得回退 sample、memory、旧模板版本、源应用草案或默认模型绑定。

## 失败与并发语义

稳定失败至少包括：

- `workflow_template_scope_denied`
- `workflow_template_source_application_unavailable`
- `workflow_template_source_definition_not_found`
- `workflow_template_source_definition_changed`
- `workflow_template_source_profile_unsupported`
- `workflow_template_forbidden_capability`
- `workflow_template_payload_invalid`
- `workflow_template_secret_material_forbidden`
- `workflow_template_candidate_not_found`
- `workflow_template_candidate_version_conflict`
- `workflow_template_review_transition_invalid`
- `workflow_template_version_not_found`
- `workflow_template_pointer_version_conflict`
- `workflow_template_not_listed`
- `workflow_template_target_application_unavailable`
- `workflow_template_target_binding_unavailable`
- `workflow_template_draft_id_conflict`
- `workflow_template_store_unavailable`

相同 expected review / pointer version 的并发 mutation 最多一个成功。candidate approval 物化版本、listing pointer + event + audit 必须分别在 catalog owner 的单一 lock / transaction 中原子提交。派生只在所有 catalog、Definition、Application、membership 和 binding 预检通过后调用 Saved Draft owner；Draft create 失败时不得写部分 provenance，也不得修改模板 pointer。

## Web 与产品面

- 在现有 Workflow 产品面新增功能驱动的 `Templates` 任务区，不建立 Family UI `S11`，不恢复旧同层 UI 迁移计划。
- 页面组织 Catalog、Candidate Review、Version / Listing 与 Derive 四条连续任务轨，只挂载一个当前 workspace owner。
- Catalog 只显示当前 workspace 上架版本；Reviewer / Operator 视图按权限显示 pending、unlisted 与历史版本。
- application / workspace / actor 变化必须清空 selection、target application、confirmation、cursor 和迟到响应；易失字段不进入 URL、`localStorage`、`sessionStorage`、IndexedDB 或跨标签 payload。
- React 前必须完成 `A / 完整 Pencil` 的 Desktop、无法直接推导的 Narrow、candidate review、replace / unlist danger 和 derivation blocked 状态，并由项目所有者人工批准。

## 实施批次

### 批次 A：strict contract、memory owner 与 HTTP

- 定义 candidate / decision、version、lineage / listing、event / audit 与 `derivation_v2` strict contract。
- 实现 canonical digest、Definition authority reload、portability validator、memory repository、review / listing CAS 与 strict HTTP。
- 复用现有 permissions、membership provider、Application / Definition / Saved Draft owner；不创建数据库、Pencil 或 Web。
- 状态：`completed`。Go owner、七份 schema、十条 route、默认关闭配置门禁、Saved Draft provenance union、相邻回归与并发 race 已通过；没有越过本批停止线。

### 批次 B：SQLite / PostgreSQL durable repository

- 在既有 workflow shared backend / pool / selector / migration family 中增加 catalog records 与必要索引。
- 完成 migration / rollback / reapply、runtime role、transaction CAS、append-only、cursor、restart、corruption 和 no-fallback。
- 派生继续只通过 Saved Draft owner 创建记录，不建立第二套 draft table 或 repository。

### 批次 C：完整 Pencil 与人工批准

- 反查批次 A / B 的真实功能事实，完成 Catalog、Review、Listing 与 Derive 的 Desktop / Narrow / danger / blocked 设计。
- 只使用语义 token 和既有 Family UI 语言；人工批准前不进入 React。

### 批次 D：React strict consumer

- 接入单一 workspace template catalog consumer、strict schema、cursor、CAS、scope generation、late-response guard 与 Draft Designer exact handoff。
- offline 零请求，target application / confirmation 只存在组件内存；不在本批启动数据库产品服务或冒充真实浏览器证据。

### 批次 E：双数据库产品连续链与专题收口

- SQLite 完成 Definition → template candidate → approve → list → target application derive → exact Draft open / validate / save → restart restore。
- PostgreSQL configured Server 完成 migration / runtime role、同构连续链、hard failure no-fallback 与 reconnect 后 authoritative reload。
- 真实浏览器覆盖 `1440×900`、`720×900`、`390×844`、双标签 CAS、workspace / application 切换、URL / Storage / console / network 与敏感信息审计。
- 同步专题、入口、当前焦点、路线图、能力矩阵和周志，确认完成后关闭唯一任务卡。

## 验收方式

- Contract：unknown / duplicate field、invalid union、digest drift、forbidden material、scope drift 全部失败关闭。
- Domain：候选不可变、review / listing CAS、append-only、Definition reload、target binding check 和成功 / 失败零部分写入。
- Repository：memory / SQLite / PostgreSQL 同构，cursor、迁移、回滚、重启、并发、损坏和 no-fallback 可复验。
- Authorization：每条 mutation 复用唯一 membership decision 与组合 permission，跨 tenant / workspace / application 拒绝且无 side effects。
- Product：完整团队模板发布与派生链在双数据库和真实浏览器可复跑，derived draft 有 exact `derivation_v2` 且源资源未改变。
- Privacy：响应、日志、数据库和浏览器不包含 credential、header、endpoint、raw provider payload、运行 input / answer、cookie 或业务写回材料。
- Side effects：candidate、review、listing 和 derive 的 provider、tool、retrieval、confirmation、run、evaluation、business write、retry / fallback、replay 均为 `0`。
- 每批执行精准测试、相关 race、Web build（涉及 Web 时）、`git diff --check` 和 `./scripts/check-repo.sh --fast`；schema、migration、阶段真相源或最终收口补全量 `./scripts/check-repo.sh`。

## 停止线

- 不做公开 Marketplace、跨 workspace / tenant 共享、匿名访问、收费、评分、评论、推荐、下载量、搜索排行或第三方分发。
- 不从 Saved Draft 直接上架；来源必须是人工批准的精确不可变 Definition Version。
- 不复制、修改或替代 Definition、Saved Draft、Application、Provider、membership、Run、Evaluation 或 audit 真相源。
- 不支持 HTTP Tool、RAG、code、sandbox、agent loop、多工具、schedule、background、writeback、replay 或 resume 模板。
- 不自动 review、上架、替换、派生、保存后续编辑、创建 Definition candidate、activation 或运行。
- 不自动重绑定 Provider / Profile / Model，不 fallback 默认模型，不复制 credential / endpoint / raw provider config。
- 不创建 public production API、production repository、production auth / secret、production publish、quota、billing、SLA 或真实 Radish 接线声明。
- 不建立 Family UI `S11`，也不为普通列表、文案或 evidence panel 派生专项 checker / fixture 链。

## 当前下一步

批次 A 已完成并推进到 `workspace_workflow_template_catalog_review_controlled_derivation_dev_test_v1_batch_a_completed_batch_b_ready`。下一步必须由项目所有者显式批准批次 B 后，才允许在既有 workflow shared SQLite / PostgreSQL backend、pool、selector 与 migration family 中实现 durable repository；当前不创建 migration、Pencil、React、产品服务或真实浏览器记录。

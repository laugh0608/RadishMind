# User Workspace Application Publish Governance & Promotion v1

更新时间：2026-07-12

状态：`application_publish_governance_promotion_v1_complete`

## 功能目标

让内部开发者把当前 application 下已保存、已校验的配置草案固定为不可变发布候选，完成版本绑定、配置摘要校验、调用证据引用、审查决定、漂移识别和 promotion eligibility 复核。

本功能只建立 application 发布治理与开发测试态候选审查链。候选通过审查不等于正式 application 已发布；在真实 application repository、Radish OIDC / membership、发布 owner 和生产授权未成立时，promotion 必须保持 blocked。

## 用户流程

1. 用户从 Applications 选择 application，进入 Application Detail 和 Application Configuration Draft。
2. 用户恢复一个已保存、校验状态为 `valid` 的草案版本；未保存编辑不能直接创建候选。
3. Publish Review Workspace 重新从服务端读取 `draft_id + draft_version`，不接受浏览器提交的完整配置快照。
4. 服务端重新校验草案 scope、schema、validation state 和当前 application read baseline，并计算 sanitized configuration digest。
5. 用户可添加零个或多个 sanitized Gateway `request_id` 作为审查引用。候选只保存引用，不复制 History payload；精确记录仍由 Gateway Request History 校验和展示。
6. 创建成功后得到不可变 `application_publish_candidate.v1`。草案后续变化不会改写旧 candidate，只能创建新 candidate。
7. reviewer 在显式 dev/test 模式记录 `approve`、`reject`、`request_changes` 或 `withdraw`；每次决定使用 expected review version，冲突时不得覆盖。
8. Workspace 比较 candidate snapshot、当前草案版本和 application baseline，展示 candidate superseded、draft drift 与 application baseline drift。
9. eligibility 汇总 review state、版本漂移、正式 repository、auth / membership、发布 owner 和 promotion enablement。当前生产依赖未成立时返回稳定 blocker，不伪造发布成功。
10. 用户可以从 candidate 打开现有 API Integration、Playground 和精确 Request History detail；这些交接不新增 Gateway adapter、SSE parser 或 history store。

## 数据来源与职责

| 数据 | 真相源 / owner | 本功能用途 | 禁止替代来源 |
| --- | --- | --- | --- |
| application baseline | 现有 Control Plane Read `ApplicationSummary` repository | 创建时绑定 `updated_at`，读取时检查 baseline drift | 不从 URL、浏览器存储或 candidate body 伪造 |
| application draft | application configuration draft repository | 服务端读取精确保存版本并生成不可变快照 | 不接受客户端完整 snapshot，不复用 Workflow draft |
| publish candidate | 独立 publish candidate dev/test repository | 保存不可变配置快照、digest、review history 和 sanitized metadata | 不写正式 application 真相源 |
| Gateway evidence | 现有 Request History | 用同一 `request_id` 精确审查调用 | candidate 不复制请求、响应、prompt、usage payload 或 provider raw material |
| promotion blockers | publish governance service | 计算 eligibility 和稳定失败 | 不把 Web 文案当授权或发布结果 |

## Application scope 与角色

- create / list / read / review 必须绑定 `tenant_ref / workspace_id / application_id / owner_subject_ref`。
- request body、query、当前 application 和 dev scope header 必须一致。
- create 要求 `application_publish_candidates:write`；list / read 要求 `application_publish_candidates:read`；review 要求独立 `application_publish_candidates:review`。
- dev/test reviewer 可以与 candidate owner 相同，用于验证状态机和 CAS；这不证明生产职责分离或真实审批身份已经成立。
- candidate list 只返回当前 scope 与 owner 的 sanitized summary；跨 application、workspace、tenant 或 owner 必须 fail closed。

## Candidate 模型与不可变边界

schema 固定为 `application_publish_candidate.v1`，包含：

- `candidate_id`、`workspace_id`、`application_id`；
- `draft_id`、`draft_version`、`draft_digest`；
- `base_application_updated_at`；
- sanitized configuration snapshot：display name、description、application kind、default protocol / model、allowed protocols；
- 去重排序后的 sanitized `evidence_request_ids`；
- `candidate_state`、`review_version`、append-only review records；
- 创建 / 更新时间、actor ref、request / audit ref；
- 动态计算的 `promotion_eligibility`。

candidate 创建后，配置 snapshot、draft binding、digest、application baseline 和 evidence refs 不允许更新。review 只追加决定记录并更新 review state / version。草案或 application baseline 变化时，旧 candidate 仍可读取，但 eligibility 必须反映 drift 或 superseded。

digest 由服务端对规范化后的 sanitized configuration snapshot 进行确定性 JSON 编码后计算 SHA-256。digest 不包含 actor、request id、audit ref、review、时间戳或任何 secret。

## 审查状态机

candidate 初始状态为 `pending_review`，允许以下决定：

- `approve`：进入 `approved`；
- `reject`：进入 `rejected`；
- `request_changes`：进入 `changes_requested`；
- `withdraw`：进入 `withdrawn`。

`rejected`、`changes_requested` 和 `withdrawn` 为当前 candidate 终态；后续修改草案并创建新 candidate。`approved` 仍可被 owner 显式 withdraw，但不提供真实 promotion。所有决定必须带 4 到 500 字符的 sanitized reason，使用 expected review version，并追加 actor / request / audit metadata。

## Promotion eligibility

eligibility 输出 `eligible`、`status` 和有序 blockers。当前 v1 至少检查：

- candidate 是否 `approved`；
- 当前 application `updated_at` 是否与 candidate baseline 一致；
- 当前 saved draft version / digest 是否仍与 candidate binding 一致；
- candidate 是否已被后续同 draft candidate supersede；
- 正式 application repository 是否可用；
- production auth / membership 是否已验证；
- 发布 owner 是否已配置；
- promotion runtime 是否显式启用。

本批正式 repository、production auth、发布 owner 与 promotion runtime 均未启用，因此即使 candidate 已 approved，也必须返回 `status=promotion_blocked` 和对应 blockers，不创建 promotion endpoint，不修改 application read model。

## API 与 repository 边界

新增 dev-only routes：

- `POST /v1/user-workspace/application-publish-candidates`
- `GET /v1/user-workspace/application-publish-candidates`
- `GET /v1/user-workspace/application-publish-candidates/{candidate_id}`
- `POST /v1/user-workspace/application-publish-candidates/{candidate_id}/reviews`

create body 只允许 `candidate_id`、`draft_id`、`expected_draft_version` 和 `evidence_request_ids`。review body 只允许 `expected_review_version`、`decision` 和 `reason`。repository contract 只提供 create / read / list / append review；application baseline、draft validation、digest、eligibility 和 failure mapping 属于 service。

memory 与 PostgreSQL dev/test adapter 必须共享相同的 immutable create、review CAS、scope、排序和 sanitized projection 语义。PostgreSQL 使用独立 schema marker、manual migration runner、runtime / migration role separation；连接、marker 或 query 失败不得回退 memory。

## 失败语义

稳定 failure 至少包括：

- `publish_candidate_scope_denied`；
- `publish_candidate_not_found`；
- `publish_candidate_payload_invalid`；
- `publish_candidate_secret_material_forbidden`；
- `publish_candidate_draft_not_found`；
- `publish_candidate_draft_version_conflict`；
- `publish_candidate_draft_invalid`；
- `application_base_revision_changed`；
- `publish_candidate_immutable_conflict`；
- `publish_candidate_review_version_conflict`；
- `publish_candidate_review_transition_invalid`；
- `publish_candidate_store_unavailable`；
- `publish_candidate_write_disabled`。

store failure、scope denial、draft mismatch、baseline drift 和 CAS conflict 都不得创建或修改 candidate。冲突响应只返回当前 sanitized version / state，不返回他人 candidate 或草案。

## 隐私与前端状态边界

- 默认 offline 模式零 fetch；candidate id、review reason、evidence request id 和 selection 只在当前 React 组件内存。
- 不写 URL query/hash payload、`localStorage`、`sessionStorage`、草稿、运行记录、Gateway payload 或业务真相源。
- candidate 和 review reason 禁止 API key、Authorization、cookie、provider credential / endpoint、内部 dev header、prompt / message、模型输出、完整 HTTP payload 或 DSN。
- Web 对 create / list / read / review response 做 strict validation 和递归 forbidden-field scan；scope drift、schema drift 或敏感字段一律失败。
- Request History handoff 只携带 `request_id + application_id`；Integration / Playground handoff 只携带 candidate 的 application、protocol 和 model。

## 实施拆分

1. 建立 candidate / review / eligibility domain、memory repository、HTTP route、scope 与单元测试。
2. 建立 PostgreSQL schema、marker、manual runner、immutable create、review CAS 和集成测试。
3. 建立独立 lazy Publish Review Workspace，消费 saved draft binding、candidate list / detail / review 和 drift comparison。
4. 复用现有 Integration / Playground / Request History 事件，完成 application 切换隔离与 offline 零请求。
5. 完成真实浏览器多标签冲突、漂移、审查决定、精确 History handoff、泄漏检查和阶段文档收口。

## 验收方式

单元与集成测试至少覆盖：offline 零请求、服务端 draft reload、candidate digest、immutable create、scope / owner 隔离、review CAS、非法 transition、secret fail closed、application baseline drift、draft version / digest drift、superseded candidate、PostgreSQL migration / rollback / reapply、重启恢复、runtime role DDL denial和 no fallback。

真实浏览器应完成：选择 application、恢复已保存草案、创建 candidate、查看 digest / snapshot / blockers、添加并打开精确 History request id、在第二标签页制造 review CAS conflict、记录 approve / reject / request changes / withdraw 中的主要路径、验证 approved 仍为 promotion blocked、切换 application 后旧 candidate 与 reason 不串入，并检查 console、URL、localStorage 和 sessionStorage。

相称验证包括 Web 单测、production build、Go 单元 / race / vet、PostgreSQL integration、`git diff --check`、`./scripts/check-repo.sh --fast` 和完整 `./scripts/check-repo.sh`。

## 当前实现与复验结果

- Platform 已实现 candidate / review / eligibility domain、memory dev 与 PostgreSQL dev/test repository、四条 scoped dev-only route、独立 migration / marker / manual runner 和 fail-closed store selector；create 只接收 draft binding，由服务端重读草案并计算 digest。
- Web 已实现独立 lazy Publish Review Workspace，支持 saved valid draft 选择、candidate 创建 / 列表 / 恢复、snapshot 比较、review CAS、promotion blocker、Integration / Playground 与精确 History 交接；读取 candidate 与记录 review 使用不同状态语义。
- 单元和 PostgreSQL 集成覆盖 immutable create、scope / owner、secret guard、draft / baseline drift、superseded、review transition / CAS、重启恢复、runtime DDL denial、rollback / reapply 和 no fallback。
- 真实浏览器已完成 `app_flow_copilot` 草案保存、Responses unary、同一 `request_id` History detail、candidate 创建、第二标签页 approve、第一标签页 stale review conflict 恢复，以及切换到 `app_docs_assistant` 后旧 candidate / reason / evidence 清空。
- approved candidate 仍稳定显示 `promotion_blocked`，并列出正式 application repository、production auth、发布 owner 和 promotion runtime 未成立的四项 blocker；没有发生正式 application mutation。
- 浏览器控制台无 error / warning，URL 只保留稳定 section hash，`localStorage` 与 `sessionStorage` 均为空。

## 停止线

- 不新增正式 promotion endpoint，不创建、更新、发布或删除正式 application。
- 不接 production OIDC、membership、authorization、application repository、secret backend 或发布 owner runtime。
- 不实现 production API key、quota enforcement、billing、cost ledger、provider credential 或可信 SLA。
- 不新增 Gateway northbound protocol、provider registry、协议 adapter、SSE parser、fallback 或 load balancing。
- 不把 candidate approved、dev/test review、模型调用或 History 记录解释为 production publish。
- 不扩展 Workflow tool、confirmation、writeback、replay 或 resume。

## 后续顺位

下一产品设计转向 `Admin Control Plane Authenticated Read Store Transition v1`，一次只处理 authenticated read identity、workspace membership 与正式 read repository 的边界和迁移顺序。该专题形成设计与可复验前置条件前，不启用 application promotion，也不并行打开管理写入、production API key、quota 或 billing。

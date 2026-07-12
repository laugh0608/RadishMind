# User Workspace Application Configuration Draft & Review v1

更新时间：2026-07-12

状态：`application_configuration_draft_review_v1_complete`

## 当前实现结果

2026-07-12 已完成独立 application configuration draft 领域、dev-only validate / save / read / list API、memory dev 与 PostgreSQL dev/test repository、显式 migration、scope / owner 隔离、CAS 版本冲突、secret fail closed、Web 配置 / 校验 / 比较 / 恢复 / 冲突审查，以及到现有 API Integration 和 Playground 的 application / protocol / model handoff。

Web 49 项测试与 production build、Platform 全量 Go 测试和 PostgreSQL integration suite 通过。真实浏览器以 `app_docs_assistant` 加载 6 个模型，保存 version 1，在第二标签页保存 version 2，第一标签页触发冲突并显式恢复；随后完成 Responses unary、stream、用户取消，并以同一 `request_id` 打开 `408 / BRIDGE_WORKER_CANCELED / postgres_dev_test` History detail。console 为 0 error / 0 warning，URL 只含稳定 section hash，localStorage / sessionStorage 为空。

## 功能目标

让内部开发者从当前 User Workspace application 建立独立配置草案，完成编辑、模型与协议校验、开发测试态保存、恢复、版本冲突审查、配置差异比较，并把已校验的 application、protocol 和 model 交给现有 API Integration Workspace 与 Gateway Playground。

本功能建立 application configuration draft 的领域与持久化边界，但不直接修改 Control Plane Read application summary，不创建正式 application，也不提供发布、删除或生产授权。

## 用户流程

1. 用户从 Applications 选择一个 application，进入现有 Application Detail。
2. Application Configuration Workspace 以当前 application 的公开字段建立内存工作副本；application 切换会清除旧表单、校验、版本、冲突和 handoff 状态。
3. 用户编辑 display name、description、application kind、默认 protocol、默认 model 与允许协议集合。
4. 默认 offline 模式不发网络请求，允许本地编辑、校验和比较，但不保存、不恢复、不加载模型。
5. 显式 dev/test 模式下，用户从现有 `/v1/models` consumer 加载当前 application scope 的模型目录；只有通过响应校验的 model id 和协议能力可进入草案。
6. 用户执行本地校验；通过后可调用 dev-only application draft route 保存。首次保存得到 version `1`，后续保存必须携带 expected version。
7. 用户可以刷新草案列表、恢复当前 application 下自己的草案，并比较草案与只读 application 基线。
8. 版本冲突时保留当前内存修改，展示服务器当前版本；用户必须显式恢复已保存版本或以当前版本为新基线继续编辑，不能静默覆盖。
9. 已通过校验的 protocol 与 model 可以交给现有 Application API Integration / Playground；调用与 History 继续由既有 Gateway consumer 和同一 `request_id` 链路承载。

## 数据来源与职责

| 数据 | 真相源 / owner | 本功能用途 | 禁止替代来源 |
| --- | --- | --- | --- |
| 当前 application | User Workspace Applications / Application Detail | 建立只读比较基线和 application scope | 不从 URL 或浏览器存储恢复 |
| 模型目录 | 现有 `GET /v1/models` consumer | 校验默认模型和协议兼容性 | 不复制 provider registry 或静态伪造 live 目录 |
| application draft | application draft service / dev-test repository | 保存 sanitized configuration draft、版本和审计元数据 | 不复用 Workflow draft repository，不写 Control Plane Read fixture |
| 测试调用 | 现有 Gateway Playground | 消费已校验 application / protocol / model handoff | 不新增协议 adapter、SSE parser 或请求状态模型 |
| 调用审查 | 现有 Gateway Request History | 使用调用产生的同一 `request_id` | 不把草案内容补写到 Gateway history |

## Application scope 与身份

- 所有保存、读取和列表操作都必须绑定 `tenant_ref / workspace_id / application_id / owner_subject_ref`。
- application id 只能来自当前 Applications 选择；请求 body、query 和 dev scope header 必须一致。
- dev/test actor context 复用现有显式开发身份头，但 application draft 使用独立 scope：`application_drafts:read` 与 `application_drafts:write`。
- list 只返回当前 scope 与 owner 的 sanitized summary；read 和 save 不允许跨 application、workspace 或 owner。
- 本批次不接 Radish OIDC、真实 membership 或 production authorization；dev headers 不进入 UI、草案 payload、日志或示例。

## 配置草案模型

草案 schema 固定为 `application_configuration_draft.v1`，包含：

- `draft_id`：稳定短标识；
- `workspace_id`、`application_id`、`base_application_updated_at`；
- `display_name`、`description`、`application_kind`；
- `default_protocol`、`default_model`、`allowed_protocols`；
- `draft_version`、`validation_state`、创建 / 更新时间和 actor ref；
- sanitized validation findings 与 request audit metadata。

草案不允许包含 API key、key hash、Authorization、cookie、provider credential、provider endpoint、内部 caller header、prompt / message 输入、模型输出或业务数据。服务端必须递归检查禁止字段名，并限制字符串、数组和 payload 大小。

## 校验与比较

校验规则包括：

- display name、description、application kind、draft id 和 scope 字段的长度与字符约束；
- application kind 必须来自当前受控 allowlist；
- allowed protocols 只能包含 Chat Completions、Responses、Messages，且去重后非空；
- default protocol 必须属于 allowed protocols；
- dev/test 加载过模型目录时，default model 必须存在，default protocol 必须属于该模型公开支持协议；
- payload 任意层级出现 secret、credential、authorization、header、endpoint 或原始调用内容字段时 fail closed。

比较视图只展示公开配置字段的 before / after / unchanged 状态，不展示 actor header、repository metadata、credential 或测试输入输出。模型目录刷新后模型不可用时，草案保持原值但进入 blocking validation，不能交给 Playground。

## 状态与失败语义

前端状态固定覆盖：

- `offline`、`unsaved`、`validating`、`invalid`、`valid`；
- `saving`、`saved`、`loading`、`restored`；
- `version_conflict`、`store_failure`、`scope_denied`。

稳定服务端失败至少包括：

- `application_draft_scope_denied`；
- `application_draft_not_found`；
- `application_draft_payload_invalid`；
- `application_draft_secret_material_forbidden`；
- `application_draft_version_conflict`；
- `application_draft_store_unavailable`；
- `application_draft_write_disabled`。

repository 故障必须显式失败，`postgres_dev_test` 不得回退 memory；冲突响应只能返回当前版本和 sanitized finding，不能返回他人的草案。

## 持久化与隐私边界

- offline 模式必须保持零 fetch。
- 未保存编辑、模型目录、比较选择和 Playground 输入输出只存在当前 React 组件内存。
- 不写入 URL query/hash payload、`localStorage`、`sessionStorage`、Workflow draft、Workflow run、Gateway history payload 或业务真相源。
- memory dev store 只用于当前进程内显式开发路径；PostgreSQL dev/test store用于重启恢复与并发验证，不代表 production repository。
- committed 文档、测试 fixture 和浏览器证据不得保存真实 secret、DSN、用户输入或模型输出。

## API 与 repository 边界

新增 dev-only routes：

- `POST /v1/user-workspace/application-drafts/validate`
- `POST /v1/user-workspace/application-drafts`
- `GET /v1/user-workspace/application-drafts`
- `GET /v1/user-workspace/application-drafts/{draft_id}`

repository contract 只提供 validate 之外的 save / read / list；validation 属于领域 service。memory 与 PostgreSQL adapter 必须共享同一 CAS、scope 和 sanitized projection 语义。migration 只由显式 dev/test runner 执行，平台启动不自动迁移。

## 实施拆分

1. 建立 application draft 领域类型、校验、memory store、HTTP route、稳定失败与单元测试。
2. 建立 PostgreSQL schema、marker、manual migration runner、CAS repository、store selector 和集成测试。
3. 建立独立 lazy Application Configuration Workspace，复用模型目录 consumer 与 Playground handoff。
4. 完成草案列表、恢复、比较、版本冲突处置和 application 切换隔离。
5. 完成真实浏览器连续流程、泄漏检查、文档真相源和阶段收口。

## 验收方式

单元与集成测试至少覆盖：

- offline 零请求；
- 本地合法 / 非法校验与 secret fail closed；
- save / read / list 和 sanitized summary；
- tenant / workspace / application / owner 隔离；
- CAS 版本冲突且不覆盖；
- PostgreSQL migration apply / rollback / reapply、重启恢复和 no fallback；
- `/v1/models` 兼容性校验；
- application 切换清除旧状态；
- 草案到 API Integration / Playground 的 application、protocol、model handoff；
- Playground 到 History 的同 request id handoff继续通过既有测试。

真实浏览器在显式 dev/test 配置下完成：选择 application、建立草案、加载模型、编辑、校验、保存、刷新恢复、制造并审查版本冲突、比较配置、交给 Playground、完成 unary / stream / cancel 并打开精确 History detail。验收同时检查 console error、URL、localStorage 和 sessionStorage。

相称验证包括 Web 单测、production build、Go 单元与集成测试、migration 往返、`git diff --check`、`./scripts/check-repo.sh --fast` 和完整 `./scripts/check-repo.sh`。

## 停止线

- 不创建、发布、删除或直接更新正式 application。
- 不接 production API key、key lifecycle、quota、billing、cost ledger、production auth 或真实 membership。
- 不保存 provider credential、endpoint、Authorization、内部 dev header、测试输入输出或完整 HTTP payload。
- 不新增 northbound Gateway protocol、provider registry、协议 adapter、SSE parser、fallback 或 load balancing。
- 不把 dev/test 草案保存、模型校验或调用成功解释为 production ready、真实 provider SLA 或正式 application 已发布。
- 不扩展 Workflow tool、confirmation、writeback、replay 或 resume。

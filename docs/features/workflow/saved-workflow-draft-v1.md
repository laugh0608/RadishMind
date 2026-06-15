# Saved Workflow Draft v1 功能专题

更新时间：2026-06-15

## 专题定位

`Saved Workflow Draft v1` 是 `Workflow / Agent Runtime` 中第一个从只读审查面走向可保存用户草案的功能专题。它让用户工作区中的 workflow 草案可以被保存、读取和校验，并让 reviewer 看到同一份可审查记录。

本专题不是 workflow publish / run / executor 专题。成功保存只表示“草案已保存且可审查”，不表示 workflow 已发布、可执行或已接入上层项目。

## 当前状态

- Platform Go domain service 已实现，文件为 `services/platform/internal/httpapi/workflow_saved_draft.go`。
- 已覆盖 `SavedWorkflowDraft` v1 类型、memory dev store、`SaveDraft` / `ReadDraft` / `ValidateDraft`、版本冲突、失败语义、sanitized response、no sample fallback 和 no side effects tests。
- 当前已新增 dev-only HTTP route 和 web consumer 状态区分：`POST /v1/user-workspace/workflow-drafts`、`GET /v1/user-workspace/workflow-drafts/{draft_id}` 和 `POST /v1/user-workspace/workflow-drafts/validate` 默认关闭，只在显式 dev 配置下工作。
- 当前仍没有 durable persistence、repository adapter、schema migration、store selector、Radish OIDC 或 production API。
- 当前任务卡为 [Workflow Saved Draft v1 Implementation](../../task-cards/workflow-saved-draft-v1-implementation-plan.md)，状态是 `saved_workflow_draft_domain_service_implemented`。

## 目标用户

- `Workspace Builder`：保存未完成的 workflow 设计，后续继续编辑。
- `Workflow Reviewer`：读取同一份草案，审查结构、风险和 blocked capability。
- `Platform Maintainer`：维护 schema、校验规则、失败语义和停止线，避免后续 executor 或 confirmation 被隐式打开。

## 数据边界

Saved draft 是用户工作区中的可编辑设计记录，不是 published workflow definition，也不是 run record。

允许保存：

- `draft_id`、`workspace_id`、`application_id`、`source_definition_id`、`base_definition_version`、`draft_version`、`schema_version`、`draft_status`。
- 草案名称、说明、节点、边、输入契约、输出契约、provider / profile ref、tool ref、RAG ref、condition 摘要和 output mapping。
- validation summary、blocked capability summary、request / audit metadata。

明确排除：

- secret value、API key value、OAuth / OIDC token、完整用户 claim。
- 真实工具执行结果、materialized result、confirmation decision。
- execution plan persistence、runtime readiness persistence、scenario / review / handoff persistence。
- run input / output、business writeback payload、replay / resume state。

## 功能流程

保存流程：

1. UI 或 consumer 提交 sanitized draft payload，必须包含 workspace、application、schema version、draft version 和 graph 主体。
2. 服务端执行 normalize、field allowlist、forbidden field reject 和大小预算检查。
3. 校验 schema、graph、contract、capability 和 risk。
4. 成功时原子保存并递增 `draft_version`，返回 sanitized draft、validation summary、blocked capability summary 和 request / audit metadata。
5. 失败时不得产生部分写入，不得回退 sample / fixture，不得创建 run record 或 confirmation decision。

读取流程：

1. 按 workspace + application + draft scope 查询。
2. scope mismatch、not found 或 store unavailable 必须 fail closed。
3. 读取结果只返回 sanitized draft、版本信息、validation summary、blocked capability summary 和 request / audit metadata。
4. UI 可以继续展示 sample，但必须明确标记 `sample / unsaved`，不能混成 saved draft record。

校验流程：

- `valid_for_review` 只表示草案可审查，不表示 publish ready 或 run ready。
- `invalid_draft` 可用于可恢复的不完整业务草案。
- `blocked_capability` 可用于保留 code、sandbox、agent_loop、executor、confirmation decision、writeback、replay 等禁止能力的审查 finding。
- `schema_unsupported` 用于旧 schema 或不支持 schema 的可审查失败状态。

## 失败语义

必须保持以下 failure code 和行为一致：

| failure code | 行为 |
| --- | --- |
| `draft_scope_denied` | 拒绝读写，不返回草案主体 |
| `draft_not_found` | 返回 not found，不回退 sample |
| `draft_schema_version_unsupported` | 拒绝保存；读取时只返回可审查 metadata |
| `draft_payload_invalid` | 拒绝保存，无部分写入 |
| `draft_graph_invalid` | 结构损坏则拒绝；可恢复问题可保存为 invalid finding |
| `draft_contract_invalid` | 可保存为 invalid，不允许 publish / run |
| `draft_blocked_capability` | 可保存为 blocked finding 或拒绝危险字段 |
| `draft_version_conflict` | 拒绝覆盖，返回当前版本 metadata |
| `draft_payload_too_large` | 拒绝保存 |
| `draft_store_unavailable` | fail closed，不回退 fixture |
| `draft_write_disabled` | 拒绝保存，允许继续查看 sample |

## 下一批开发

dev-only consumer integration 已按 [Dev-only Saved Draft Consumer](dev-only-saved-draft-consumer.md) 落地。后续可选择补更细的 conflict UI、consumer smoke 或 route contract 固化；任何 durable persistence、public production API、database、OIDC、repository adapter 或 executor 仍必须作为独立专题和 task card 推进。

## 验收方式

- Go 单元测试覆盖 save / read / validate 成功和失败路径。
- Consumer 能区分 sample / unsaved draft 与 saved draft record。
- Web build 和 workflow 相关聚合检查通过。
- `./scripts/check-repo.sh --fast` 通过。
- 若新增 committed schema、route contract、fixture 或 checker，先更新 task card，并按风险补全专项验证。

## 停止线

- 不实现 publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不接真实数据库、repository adapter、schema migration、store selector、Radish OIDC、token validation、API key lifecycle、quota enforcement、billing 或 public production API。
- 不把 `valid_for_review`、validation summary、risk summary 或 readiness summary 当作执行解锁条件。

# Workflow Saved Draft v1 Implementation 任务卡

更新时间：2026-06-14

## 任务标识

- 切片：`workflow-saved-draft-v1-implementation`
- 轨道：`Workflow / Agent Runtime`
- 状态：`saved_workflow_draft_domain_service_implemented`

## 目标

在 `Workflow / Agent Runtime` 功能文档已经明确 `Saved Workflow Draft v1` 可打开范围后，完成首个实现批次：固定 draft schema / 类型、存储边界、save / read / validate 契约、版本冲突、失败语义和验证策略。

本任务卡用于约束后续代码实现批次。它允许打开草案保存、读取、校验、schema 和受控存储边界，但不把 saved draft 扩张为 publish、run、executor、confirmation decision、writeback、replay 或 production API。

## 本轮实现

- 已新增 `services/platform/internal/httpapi/workflow_saved_draft.go`，定义 `SavedWorkflowDraft` v1 结构化类型、failure code、validation summary、blocked capability summary、request / audit metadata、内存 dev store 和 `SaveDraft` / `ReadDraft` / `ValidateDraft` domain service。
- 已新增 `services/platform/internal/httpapi/workflow_saved_draft_test.go`，覆盖成功保存与读取、invalid 可保存、blocked capability 可审查、版本冲突、scope denied、not found、schema unsupported、payload too large、store unavailable、write disabled、no sample fallback、无部分写入和无 executor / confirmation / writeback / replay 副作用。
- 当前存储形态是 platform 内部 memory dev store boundary，不是 durable repository adapter、真实数据库、schema migration、store selector 或 public production API。
- 当前未新增 HTTP route 或 web consumer；下一步若把 domain service 接到 UI / dev-only route，应先明确 consumer contract、dev auth / write enablement 和 sample / saved record 区分方式。

## 输入事实源

- [Workflow / Agent Runtime 设计与开发文档](../features/workflow-agent-runtime.md)
- [功能设计文档入口](../features/README.md)
- [当前推进焦点](../radishmind-current-focus.md)
- [`Workflow Draft Designer Offline` v1 计划](workflow-draft-designer-offline-v1-plan.md)
- [`Workflow Draft Validation Inspector Offline` v1 计划](workflow-draft-validation-inspector-offline-v1-plan.md)
- [`Workflow Function Surface Readiness Close` v1 计划](workflow-function-surface-readiness-close-v1-plan.md)
- [`Control Plane Durable Read Foundation` v1 任务卡](control-plane-durable-read-foundation-v1-plan.md)

## 实现边界

1. Draft schema / 类型
   - 定义 `SavedWorkflowDraft` v1 的结构化字段：identity、scope、version、status、editable graph、contracts、provider / tool / RAG references、risk / validation summary 和 audit/request metadata。
   - 字段命名优先沿用功能文档：`draft_id`、`workspace_id`、`application_id`、`source_definition_id`、`base_definition_version`、`draft_version`、`schema_version`、`draft_status`。
   - 用户可编辑内容只保存结构化设计，不保存 secret value、API key value、token、真实工具执行结果、materialized result、confirmation decision、run input/output 或 writeback payload。

2. 存储边界
   - 实现批次必须先选择一个明确存储形态：local file、dev store、fake store bridge 或经任务卡复核后的正式存储入口。
   - 当前推荐从 repository-friendly 的 dev/fake store boundary 开始，保持与 `ControlPlaneReadRepository` 已有 interface 化方向兼容。
   - 不允许在没有 schema / scope / conflict / failure tests 的情况下直接接真实数据库、repository adapter、schema migration、store selector 或 public production API。

3. Save contract
   - save 必须接收 sanitized draft payload，执行 normalize、field allowlist、forbidden field reject、graph / contract / capability validation。
   - save 成功时返回 saved draft、`draft_version`、`validation_summary`、`blocked_capability_summary` 和 request / audit metadata。
   - save 失败不得产生部分写入，不得回退写入 sample / fixture，不得创建 run record 或 confirmation decision。

4. Read contract
   - read 必须按 workspace + application + draft scope 查询。
   - scope mismatch、not found、store unavailable 必须 fail closed。
   - read 不得把 offline sample 当作 saved draft 返回；sample / unsaved 状态必须在 UI 或 response metadata 中可区分。
   - read 应能复核当前 validation policy，并对旧 schema / policy drift 返回可审查状态。

5. Validate contract
   - validation 分层覆盖 schema、graph、contract、capability 和 risk。
   - `valid_for_review` 只表示草案可审查，不表示 publish ready、run ready 或 production ready。
   - blocked capability 可以随 draft 保存为审查 finding，但不能解锁 executor、confirmation decision、writeback 或 replay。

## 失败语义

实现批次应固定并测试以下 failure code：

| failure code | 触发条件 | 要求 |
| --- | --- | --- |
| `draft_scope_denied` | workspace / application / actor scope 不匹配 | 拒绝读写，不返回草案主体 |
| `draft_not_found` | `draft_id` 不存在或不属于当前 scope | 返回 not found，不回退 sample |
| `draft_schema_version_unsupported` | schema version 不被支持 | 拒绝保存；读取时返回可审查 metadata |
| `draft_payload_invalid` | JSON / 类型 / required fields 无法解析 | 拒绝保存，无部分写入 |
| `draft_graph_invalid` | 节点、边、入口、出口或循环策略不合法 | 按结构损坏程度拒绝或保存为 invalid finding |
| `draft_contract_invalid` | input / output / tool / RAG / condition 契约不一致 | 可保存为 invalid，不允许 publish / run |
| `draft_blocked_capability` | payload 包含 v1 禁止能力 | 可保存为 blocked finding 或拒绝危险字段 |
| `draft_version_conflict` | 客户端基于旧版本写入 | 拒绝覆盖，返回当前版本 metadata |
| `draft_payload_too_large` | 节点、边或文本字段超过预算 | 拒绝保存 |
| `draft_store_unavailable` | 存储不可用 | fail closed，不回退 fixture |
| `draft_write_disabled` | 环境只允许只读或 sample 模式 | 拒绝保存，允许继续查看 sample |

## 验收口径

- 新增或更新的 schema / type 定义能表达 saved draft identity、scope、version、status、editable graph、contracts、validation summary 和 blocked capability summary。
- save / read / validate 的实现入口覆盖成功、invalid、blocked、scope denied、not found、schema unsupported、version conflict、payload too large 和 store unavailable。
- 测试证明 save 失败无部分写入，read 不回退 sample，validation 不把 `valid_for_review` 写成 executable ready。
- UI 或 consumer 能区分 sample / unsaved draft 与 saved draft record。
- `./scripts/check-repo.sh --fast` 必须通过；若实现引入 schema 真相源、专项 checker 或阶段口径变化，补跑 `./scripts/check-repo.sh`。

## 非目标

- 不实现 workflow executor、node executor、tool executor、agent loop、publish、run、confirmation decision、decision store、execution unlock、business writeback、run replay、run resume 或 materialized result reader。
- 不直接实现真实数据库、repository adapter、schema migration、store selector、Radish OIDC、token validation、public production API consumer、API key lifecycle、quota enforcement、billing 或 cost ledger。
- 不新增同层只读 evidence 面板来替代 saved draft 实现。
- 不修改 `RadishFlow`、`Radish` 或 `RadishCatalyst` 外部仓库。

## 停止线

- Saved draft 的成功状态只能表示“草案已保存且可审查”。
- Draft storage boundary 不能被写成 durable repository adapter ready 或 production database ready。
- `valid_for_review`、validation summary、risk summary 和 readiness summary 不能被用作 publish / run / executor 的解锁条件。
- 任何 executor、confirmation decision、writeback 或 replay 能力都必须回到功能文档另开独立目标。

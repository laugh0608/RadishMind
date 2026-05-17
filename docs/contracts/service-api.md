# RadishMind 服务/API 接入契约

更新时间：2026-05-13

## 协议兼容边界

`CopilotRequest`、`CopilotResponse` 与 `CopilotGatewayEnvelope` 继续是本仓库内部的 canonical protocol。

这意味着后续即使平台要同时支持：

- 南向调用外部模型和外部 provider
- 北向对外提供常见 AI 协议接口

也仍然必须遵循“兼容层只做翻译，不另起第二套真相源”的规则。

实现语言可以分布在 `Go`、`Python` 和 `TypeScript`，但都必须从 `contracts/` 读取同一套 schema 和 canonical contract，不得各自重新定义业务真相源。

当前目标口径应固定为：

- 北向兼容：native Copilot API、`/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models`、`/v1/models/{id}`、`/v1/session/metadata`、`/v1/session/recovery/checkpoints/{checkpoint_id}`、`/v1/tools/metadata`、`/v1/tools/actions`
- 南向兼容：`RadishMind-Core`、`local_transformers / HuggingFace`、`Ollama`、OpenAI-compatible、Gemini native、Anthropic messages

当前真实状态是：

- `services/runtime/inference_provider.py` 已具备 `openai-compatible` 主入口，并可按 profile 分流到 `openai-compatible chat`、`gemini-native` 与 `anthropic-messages`；同时已补上 `HuggingFace` 与 `Ollama` 的第一版 chat-completions provider coverage
- `services/platform/` 已具备最小 `Go` 服务壳与 Python bridge-backed `HTTP` 路由，先固定 `HTTP` 服务启动、`/healthz`、`/v1/models`、`/v1/models/{id}`、`/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/session/metadata`、`/v1/session/recovery/checkpoints/{checkpoint_id}`、`/v1/tools/metadata` 与 `/v1/tools/actions`，并开始把 northbound 请求翻译并桥接到 canonical `CopilotRequest / CopilotResponse / CopilotGatewayEnvelope`
- `local_transformers` 当前主要存在于 `scripts/run-radishmind-core-candidate.py` 的本地 candidate/runtime 评测链路
- `HuggingFace` / `Ollama` 已有第一版 provider coverage，`/v1/models` 已暴露 provider-qualified profile inventory，并与请求选择、diagnostics、request observability 和 error taxonomy 共享平台门禁；正式 secret backend、环境隔离和外部 provider health check 仍未落地

当前第一版 `Go -> Python` bridge 的 northbound 切片仍然很窄：

- 文本消息仍是主要切片，但 `/v1/chat/completions`、`/v1/responses` 与 `/v1/messages` 已具备第一版 SSE 兼容与 bridge 增量转发 smoke
- 只把最后一条文本用户消息映射到 `radish/answer_docs_question`
- `GET /v1/models` 已从 provider 目录推进到第一版 bridge-backed provider/profile inventory，并补上 `GET /v1/models/{id}` 的精确 lookup；`/v1/chat/completions` 也已经把 request-side provider/profile 选择显式化并把流式路径推进到 bridge 增量转发；provider/profile discoverability、request observability 和 error taxonomy 已进入平台门禁
- 当前 northbound `model` 选择已经开始支持 configured default、provider id、legacy `profile:<name>` 与 provider-qualified `provider:<provider>:profile:<profile>`；同时 `radishmind.provider` / `radishmind.provider_profile` 扩展字段已具备显式覆盖能力

## 当前服务/API 接入切片

从 `suggest_flowsheet_edits v93` 与 `suggest_ghost_completion v25` 收口后，下一步不再默认通过继续跑真实样本推进，而是把现有协议和评测资产上提为服务/API 接入门禁。

当前最小切片建议固定为：

1. 上层或本地 smoke 提供 schema-valid `CopilotRequest`
2. `Go` northbound 入口把 OpenAI 请求翻译成最小 canonical request，并交给 Python bridge
3. `Copilot Gateway / Task Router` 校验 `project / task / schema_version`
4. Gateway 调用现有 `services/gateway/copilot_gateway.py` 路径生成 `CopilotGatewayEnvelope`
5. Response Builder 保持统一 `risk_level / requires_confirmation / citations / proposed_actions`
6. 服务层只输出 advisory response，不写回 `RadishFlow`、`Radish` 或 `RadishCatalyst` 真相源
7. 同一请求必须能复用既有 eval regression、candidate record audit 与治理报表作为验收门禁

真实 provider capture 只在服务/API 或集成演示暴露现有样本无法覆盖的新 drift 时触发；触发前应先写清楚新假设、覆盖缺口和退出条件。

### `CopilotGatewayEnvelope` 调用口径

当前 gateway 层返回值统一采用 `CopilotGatewayEnvelope`，schema 真相源为 `contracts/copilot-gateway-envelope.schema.json`。

首个对外调用形态当前固定为**Go HTTP + Python gateway bridge**：

```python
from services.gateway import GatewayOptions, handle_copilot_request

envelope = handle_copilot_request(
    copilot_request,
    options=GatewayOptions(provider="mock"),
)
```

HTTP JSON 现在由 `Go` 平台服务层承接，但它仍然只是这条 canonical bridge 的包装形态；长驻服务、鉴权、端口和部署生命周期尚未进入更完整切片前，不把更复杂的 API 表面当成真相源。未来若扩展更多 northbound 形态，也必须复用同一个 `CopilotGatewayEnvelope` 语义，而不是引入第二套响应协议。

调用侧口径建议固定为：

- 上层提交 schema-valid `CopilotRequest`，不直接调用任务 runtime 或 provider
- Gateway 返回 `schema_version / status / request_id / project / task / response / error / metadata`
- 当 `status=ok` 或 `status=partial` 时，`response` 必须存在，并继续按 `contracts/copilot-response.schema.json` 校验和消费
- 当 `status=failed` 时，调用侧必须优先读取 `error.code` 与 `error.message`；若同时存在 `response`，它也只能作为 failed advisory response 展示或记录，不能转成可执行动作
- `metadata.route` 固定表达 `project/task`，用于调用侧日志、审计和路由观测
- `metadata.provider` 只表达本次 gateway 使用的 provider/profile/model 摘要，不应被上层当作业务结果
- `metadata.advisory_only` 必须保持 `true`；上层不得绕过人工确认或规则层复核直接执行 `response.proposed_actions`
- `metadata.request_validated` 与 `metadata.response_validated` 用于确认 gateway 是否完成请求和响应 schema 校验；生产前调用侧应把任一 `false` 视为异常观测信号
- 对 `RadishFlow suggest_flowsheet_edits`，即使 gateway 返回 `partial`，只要 `response.requires_confirmation=true`，上层也只能呈现候选建议，不能写回 `FlowsheetDocument`

当前仓库用两层 smoke 固定这条调用口径：

- `scripts/check-gateway-service-smoke.py --check-summary scripts/checks/fixtures/gateway-service-smoke-summary.json`：固定进程内 Python API 的成功调用、schema-valid 但 unsupported route、schema-invalid request 三类 envelope 行为
- `scripts/run-radishflow-gateway-demo.py --manifest scripts/checks/fixtures/radishflow-gateway-demo-fixtures.json --check-summary scripts/checks/fixtures/radishflow-gateway-demo-summary.json --check`：固定 `RadishFlow export -> adapter/request -> gateway envelope` 的服务/API 集成 smoke

这些 summary 只锁定上层依赖的 envelope 行为字段，不锁死完整自然语言响应。

当前还新增 `scripts/check-radishflow-service-smoke-matrix.py --check-summary scripts/checks/fixtures/radishflow-service-smoke-matrix-summary.json` 作为矩阵级门禁。该门禁不新增业务模拟，而是把 `run-copilot-inference.py` CLI runtime、进程内 gateway、`RadishFlow export -> adapter/request -> gateway` demo、UI consumption summary 与 candidate edit handoff summary 收束到同一条 `radishflow/suggest_flowsheet_edits` 服务切片，检查 route、provider、`requires_confirmation`、advisory-only、UI 不写回和 handoff 不执行这些跨入口不变量。

当前平台级 `ops smoke` 由 `scripts/check-platform-ops-smoke.py` 承接，并已接入 `check-repo --fast`。它不启动长驻服务、不访问外部 provider，只固定 `Go` 平台层 `go test ./...`、Python bridge provider registry、openai-compatible fallback profile chain、HuggingFace profile 与 Ollama local profile inventory 这些可快速复验的不变量。该门禁证明平台 bootstrap、northbound handler 和 southbound discoverability 能在本地受控环境下同时成立，但仍不等同于 production deployment。

当前平台级结构化诊断由 `radishmind-platform diagnostics` 与 `scripts/check-platform-diagnostics.py` 承接，并已接入 `check-repo --fast`。它聚合启动配置、配置必填字段、Python bridge provider registry 和 provider/profile inventory，输出 `status`、`checks`、`failure_codes`、`bridge`、`providers` 与脱敏 `config`；其中 credential 只表达 `configured / missing / optional_missing / not_required` 这类状态，不输出 API key、token、cookie、secret 或 provider URL 原文。该诊断用于启动前 failure boundary 和本地部署排障，不替代生产 secret backend、进程守护或外部 provider 健康探测。

`/v1/models` 的 profile 条目必须暴露稳定 discoverability metadata：`capabilities`、`northbound_protocols`、`northbound_routes`、`credential_state`、`deployment_mode`、`auth_mode` 与 `streaming`。其中 `credential_state` 只允许表达 `configured / missing / optional_missing / not_required` 这类状态，不得泄漏 API key、token 或 secret 原文；`deployment_mode` 用于区分 `remote_api`、`local_daemon`、`embedded` 等运行形态。

`/v1/models`、northbound request selection 和 `diagnostics.providers` 必须复用同一套 provider/profile discoverability 口径：profile 可选择 ID 固定为 `profile:<profile>` 或 `provider:<provider>:profile:<profile>`，selection metadata 固定暴露 `selected_provider`、`selected_provider_profile`、`selected_model`、`upstream_model`、`selection_source`、`selection_inventory_kind`、`credential_state`、`deployment_mode`、`auth_mode`、`streaming`、`northbound_routes` 与 `northbound_protocols`。如果某个 profile 在 `/v1/models` 中可见，使用对应 model id 发起 `/v1/chat/completions`、`/v1/responses` 或 `/v1/messages` 时，canonical request 的 `context.northbound` 必须带出同源 selection metadata，避免“列表可见但请求选择不可解释”的漂移。

当前 northbound `radishmind` 扩展也已开始承接首版 session metadata：当请求提供 `conversation_id`、`turn_id`、`parent_turn_id`、`history_policy` 或 `history_window` 时，平台层会把 `conversation_session_record` 写入 canonical request 的 `context.northbound.session`。该记录遵循 [会话记录契约](session.md)，只固定 history policy、recovery record 和 advisory-only audit 边界；当前不实现 durable session store，也不把兼容层变成业务真相源。

当前 `GET /v1/session/recovery/checkpoints/{checkpoint_id}` 只承接 session recovery checkpoint 的 metadata-only route smoke。它返回固定 fixture 边界、checkpoint refs、tool audit refs、`tool_audit_summary`、replay policy 摘要和 state summary，并显式拒绝 materialized result / replay 类查询；这条路由不是 durable checkpoint store、materialized result reader 或 replay executor。

当前 `GET /v1/session/metadata`、`GET /v1/tools/metadata` 与 `POST /v1/tools/actions` 只承接最小 session/tooling 产品骨架。session metadata route 返回 northbound `radishmind` 扩展字段、history/state/recovery 边界和 disabled capability；tools metadata route 返回当前 contract-only registry view；tool action route 对工具 action 请求返回 `tool_action_blocked_response`，用于上层或 UI 消费 `status=blocked`、denial code、confirmation requirement 和 side-effect absence。该 blocked response 不代表 confirmation 接线、executor、durable store、materialized result reader、长期记忆、业务写回或 replay 已启用。

上层或未来 UI 的最小 TypeScript 消费口径已固定在 `contracts/typescript/session-tooling-api.ts`。该文件只提供 `SessionMetadataResponse`、`ToolingMetadataResponse`、`ToolActionBlockedResponse` 类型和 blocked view helper，调用侧应把 `canExecute=false`、`statusLabel=blocked`、`primaryCode`、`requiresConfirmation` 与 `noSideEffects` 作为展示字段；不得把 metadata shell 或 blocked response 转成可执行命令。

开发者可用 `scripts/run-platform-session-tooling-consumer-smoke.py --check` 在离线 fixture 模式下生成同样的消费视图；如果本地平台服务已经启动，可加 `--base-url http://127.0.0.1:8080` 直接请求真实 API surface。该脚本只验证上层展示语义，不启动或模拟 executor。

未来平台控制台或上层 UI 的首版视图模型固定在 [Session / Tooling UI View 契约](session-tooling-ui-view.md)：`SessionStatusViewModel`、`ToolRegistryViewModel` 和 `BlockedActionBannerViewModel`。这些 view model 只用于展示 metadata-only 与 blocked 状态，不声明真实执行、确认流、持久化或业务写回能力。

当前本地启动 runbook 固定在 `services/platform/README.md`，并由 `scripts/check-platform-runbook.py` 防止配置、路由和命令说明漂移。该检查会对齐 `RADISHMIND_PLATFORM_*` 环境变量、`/healthz`、`/v1/models`、`/v1/models/{id}`、`/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/session/metadata`、`/v1/session/recovery/checkpoints/{checkpoint_id}`、`/v1/tools/metadata`、`/v1/tools/actions` 和最小 curl smoke 命令；它只保证本地开发入口可复验，不代表 secret 管理、进程守护、部署观测或生产鉴权已经完成。

### `RadishFlow` UI 消费口径

`RadishFlow` 调用侧在消费 `CopilotGatewayEnvelope` 时，应先把 envelope 映射为 UI 可展示、可审计、不可直接执行的 consumption view：

- `status=ok/partial` 且存在 `response.proposed_actions` 时，可展示候选提案卡片，但必须保留 `requires_confirmation`
- `candidate_edit` 只能显示为 advisory proposal，后续真实修改仍交由上层命令层和人工确认流处理
- `status=failed` 时，UI 应优先展示 `error.code / error.message`，并把同时存在的 failed `response` 仅作为说明或日志，不转成候选动作
- `metadata.provider / route / request_validated / response_validated / advisory_only` 应进入调用侧审计或诊断信息，不作为业务真相源
- 任一消费路径都不得直接写回 `FlowsheetDocument`

当前仓库用 `scripts/check-radishflow-gateway-ui-consumption.py --check-summary scripts/checks/fixtures/radishflow-gateway-ui-consumption-summary.json` 固定这层消费口径。该 summary 覆盖 proposal-ready、unsupported route 与 schema-invalid 三类路径，确保上层可消费视图始终保持 advisory-only。

### `RadishFlow` 候选编辑 handoff 口径

当用户在 UI 中确认某条 `candidate_edit` 后，调用侧仍不应直接执行 patch，而应生成命令层可接收的候选 handoff：

- 只有 `kind=candidate_edit` 且 `requires_confirmation=true` 的动作可以进入 handoff
- handoff 输出是 `radishflow_candidate_edit_proposal`，不是已执行命令
- handoff 必须保留 `source_request_id`、`source_route`、`action_index`、`target`、`patch_keys`、`risk_level` 与 `citation_ids`
- handoff 必须显式标记 `requires_human_confirmation=true` 与 `can_execute=false`
- `status=failed`、schema-invalid request 或没有合格 `candidate_edit` 的响应不得产生 handoff

当前仓库用 `scripts/check-radishflow-candidate-edit-handoff.py --check-summary scripts/checks/fixtures/radishflow-candidate-edit-handoff-summary.json` 固定这层 handoff 口径。该 summary 只证明候选动作可被安全交接给上层命令层，不代表本仓库已经执行或实现 `RadishFlow` 真实命令。

### 当前上层接入等待口径

`RadishFlow`、`Radish` 与 `RadishCatalyst` 暂时都还没有进入真实模型 / Agent 接入阶段，因此当前不把真实上层调用位置作为 `RadishMind` 的阻塞项。

在上层项目准备好之前：

- 当前已落地的 gateway smoke、服务级 smoke 矩阵、`RadishFlow` UI consumption summary 与 candidate edit handoff summary 视为未来接入验收门禁
- 不继续为尚不存在的上层 UI / 命令层扩展更多本仓库内模拟 summary
- HTTP JSON 包装已由 `Go` 平台服务层承接，但它仍只是 canonical bridge 的第一版包装形态，不在没有真实消费方前继续膨胀协议表面
- 当前主线优先维护 `M3` 的 service/API smoke 矩阵，继续把 gateway、UI consumption 与 candidate edit handoff 作为同一条 `RadishFlow` 服务切片的验收门禁；`M4` 仅保留已收口的 builder/tooling 与 `3B/4B` 审计记录，不再继续扩同一批 guided 或 broader 样本面
- 未来 `RadishFlow`、`Radish` 或 `RadishCatalyst` 准备接入时，应优先复用现有 `CopilotRequest`、`CopilotResponse` 与 `CopilotGatewayEnvelope`，而不是新增第二套协议
- `RadishCatalyst` 当前仅作为第三项目预留，不修改 `contracts/copilot-request.schema.json`、`contracts/copilot-response.schema.json` 或 gateway route；等第一条真实任务、adapter skeleton 和最小 eval sample 一起进入时，再扩 `project=radishcatalyst` 的 schema 枚举和 smoke

### 当前仓库集成边界

当前正式口径是：`RadishMind`、`RadishFlow`、`Radish` 与 `RadishCatalyst` 保持独立仓库，通过协议、gateway、adapter、评测门禁和版本化制品集成，而不是在正式仓库之间互相建立 `git submodule`。

当前不采用以下做法作为默认方案：

- 不在 `RadishMind` 中加入 `RadishFlow`、`Radish`、`RadishCatalyst` 子模块
- 不在 `RadishFlow`、`Radish`、`RadishCatalyst` 中默认加入 `RadishMind` 子模块

原因是当前更需要先冻结服务/API 接口、schema、smoke 与验收门禁，而不是提前把多仓库源码分发方式固化为长期结构。若后续确实需要多仓联调或集成演示，应优先单独建立 integration workspace / super-repo，而不是让任一正式仓库承担长期聚合角色。

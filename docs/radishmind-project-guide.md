# RadishMind 项目总览与使用指南

更新时间：2026-05-17

## 这份文档讲什么

这是一份面向新读者的项目说明书，回答四个问题：

- `RadishMind` 现在被定义成什么
- 仓库怎么分工
- 今天可以怎样实际运行它
- 如果你要继续推进开发，先看哪几条主线

它不替代 `docs/radishmind-current-focus.md`、`docs/devlogs/` 或任务卡，也不记录阶段推进流水。

## 项目定位

`RadishMind` 现在的正式定位是：

协议驱动、可审计、可本地部署、可工具化的 Copilot / Agent runtime platform。

它不是上层业务真相源，不替代 `RadishFlow`、`Radish` 或 `RadishCatalyst` 的业务决策权，也不是“只放模型实验”的仓库。

当前主要职责是：

- 接收结构化上下文
- 运行最小推理链路
- 兼容多种上游模型接入方式与多种下游协议接口
- 提供本地只读产品发现面，供 console 或上层 UI 读取平台可展示能力和停止线
- 组织局部工具、规则和响应收口
- 输出解释、诊断、结构化建议和候选动作
- 维护统一协议、评测门禁、审计记录和训练治理

## 实现分工

- `UI`：`React + Vite + TypeScript`
- `平台服务层`：`Go`，覆盖 `HTTP API`、`gateway`、鉴权、流式转发、长驻进程、观测和部署壳
- `模型侧`：`Python`，覆盖训练、评测、`prompt / builder`、离线推理和校验逻辑
- `contracts/`：唯一 canonical protocol 真相源，所有语言只能消费，不得各自重新定义业务协议

## 当前五条主线

1. `Runtime Service`：本地启动、gateway、route、provider/profile、协议兼容、响应封装、部署基础；当前已达到 short close，request observability 和 error taxonomy 已进入平台门禁。
2. `Conversation & Session`：会话标识、历史压缩、恢复和审计边界；当前已有 session record、recovery checkpoint record/manifest/read result、northbound session metadata、metadata-only route smoke、confirmation / audit / result / executor / storage 设计门禁、short close readiness delta、stop-line manifest 和 close-candidate status summary。
3. `Tooling Framework`：检索、附件解析、候选生成、builder、tool policy 和 audit；当前已有 tool contract、registry、audit record、session binding、metadata-only result cache、result materialization policy、executor boundary、storage backend design、deny-by-default gates 和 executor/storage/confirmation enablement plan。
4. `Evaluation & Governance`：schema、smoke、offline eval、review、promotion gate、负向消费 summary、route smoke coverage summary、readiness summary、implementation preconditions、negative regression governance suite、negative coverage rollup、route negative coverage matrix 和 readiness consistency rollup。
5. `Model Adaptation`：基座选型、prompt/runtime 协同、蒸馏、训练样本治理和模型晋级。

如果你今天想推进开发，默认先检查 `P2 Session & Tooling Foundation` 的 stop-line manifest、short close readiness delta、route smoke readiness rollup 和 readiness consistency rollup；当前状态是 governance-only，不代表真实 executor、durable store、confirmation 接线、materialized result reader、长期记忆、业务写回或 replay 已经完成。

## 目录速览

- `docs/`：正式文档源
- `contracts/`：JSON Schema 真相源
- `scripts/`：检查、运行、转换、评测和最小运维入口
- `datasets/`：eval 样本、示例对象和 candidate record
- `training/`：训练治理、实验说明和复核记录
- `services/`：runtime 与 gateway 实现
- `adapters/`：上游项目适配层
- `tmp/`：本地临时产物，不提交

## 今天能怎么运行

### 1. 直接跑最小推理链路

当前最小运行入口是 `scripts/run-copilot-inference.py`。它不是长驻服务，而是单次 CLI runtime。

```bash
python3 scripts/run-copilot-inference.py \
  --sample datasets/eval/radishflow/suggest-flowsheet-edits-reconnect-outlet-001.json \
  --provider mock \
  --response-output tmp/rf-suggest-edit.response.json
```

如果后续要接真实 provider，再显式传 `--provider openai-compatible|huggingface|ollama`、`--provider-profile`、`--model`、`--base-url`、`--api-key`。当前这条入口已经能按 profile 分流到 `openai-compatible chat`、`gemini-native` 和 `anthropic-messages` 三类上游协议；`services/platform/` 也已把 `/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models`、SSE bridge、provider/profile inventory、request-side selection、`HuggingFace` / `Ollama` coverage、本地启动 wrapper、JSON 配置层级、deployment smoke、diagnostics / failure boundary 和 discoverability 对齐纳入第一版 runtime foundation。

### 2. 跑进程内 gateway demo

如果你要看 `RadishFlow export -> adapter -> request -> gateway` 整条链路，当前正式入口是：

```bash
python3 scripts/run-radishflow-gateway-demo.py \
  --check-summary scripts/checks/fixtures/radishflow-gateway-demo-summary.json
```

这条链路当前使用 mock provider，只验证 route、schema、advisory-only 和 `requires_confirmation` 等不变量。

### 3. 跑 service/API smoke matrix

当前 `RadishFlow` 的正式服务门禁入口是：

```bash
python3 scripts/check-radishflow-service-smoke-matrix.py \
  --check-summary scripts/checks/fixtures/radishflow-service-smoke-matrix-summary.json
```

这条矩阵会一起核对：

- CLI runtime
- gateway API
- gateway demo
- UI consumption
- candidate handoff

它现在是仓库里最接近“服务切片验收”的正式门禁。

### 3.5 运行 Go 平台服务层

当前 `Go` 平台服务层已落在 `services/platform/`，用于承载 `HTTP API`、`gateway`、鉴权、流式转发、观测和部署壳。日常本地运行优先使用 wrapper，而不是手动切换目录：

```bash
./scripts/run-platform-service.sh config-check
./scripts/run-platform-service.sh diagnostics
./scripts/run-platform-service.sh serve
```

Windows / PowerShell 使用对应的 `pwsh ./scripts/run-platform-service.ps1 config-check|diagnostics|serve`。

当前它固定以下 northbound / health 路由：

- `GET /healthz`
- `GET /v1/models`
- `GET /v1/models/{id}`
- `POST /v1/chat/completions`
- `POST /v1/responses`
- `POST /v1/messages`
- `GET /v1/session/recovery/checkpoints/{checkpoint_id}`

其中 `/v1/chat/completions` 已接到第一版 bridge，并支持非流式与 SSE 增量转发；`/v1/models`、请求侧 provider/profile selection 与 `diagnostics.providers.selectable_model_ids` 共享同一套 discoverability 口径，包括 `profile:<profile>` 与 `provider:<provider>:profile:<profile>`。当请求选中 profile 时，响应 `context.northbound` 会带出脱敏后的 `credential_state`、`deployment_mode`、`auth_mode`、`streaming`、`northbound_routes` 与 `northbound_protocols`，便于客户端和部署检查判断实际命中的 provider/profile。

`GET /v1/session/recovery/checkpoints/{checkpoint_id}` 当前只是 fixture-backed metadata-only route smoke：它返回 checkpoint refs、tool audit refs、`tool_audit_summary`、replay policy 摘要和 state summary，并拒绝 materialized result、result ref、output ref、executor ref、durable memory 与 replay 类查询；它不是 durable checkpoint store、materialized result reader、executor ref reader、durable memory reader 或 replay executor。

这仍然不是 production deployment：它已经能作为本地平台服务切片运行和诊断，但尚未具备生产级 secret backend、进程监管、环境隔离和正式发布包。

### 4. 跑本地候选模型输出

如果你要继续看 `RadishMind-Core` 本地候选输出，入口仍是：

```bash
python3 scripts/run-radishmind-core-candidate.py \
  --provider local_transformers \
  --model-dir /path/to/model \
  --output-dir tmp/radishmind-core-candidate-local \
  --allow-invalid-output \
  --validate-task
```

这仍属于模型评测 / 训练前治理链路，不等同于平台正式服务。

### 5. 当前协议兼容边界

当前平台必须区分两类兼容：

- 南向模型接入：平台去调用哪些模型和哪些 provider
- 北向协议输出：外部客户端如何调用 `RadishMind`

当前真实状态是：

- 南向已有一部分：`openai-compatible` 主入口、`HuggingFace`、`Ollama`、`gemini-native`、`anthropic-messages`，以及评测链路中的 `local_transformers`
- 北向已有第一版兼容面和只读产品面：`/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models`、`/v1/platform/overview`、`/v1/session/metadata`、`/v1/tools/metadata`、blocked `/v1/tools/actions`、SSE bridge、provider/profile selection metadata 和 diagnostics discoverability 已对齐
- `P1 Runtime Foundation` 已达到 short close，当前不应继续把 provider/config/diagnostics/observability 同层细节当作主线
- 当前仍是窄切片：还缺 production secret backend、部署隔离、外部 provider health check、真实本地 console 壳，以及 session/tooling 的真实确认流接线、存储、执行和完整负向回归；P2 现有 route / gate / negative regression 资产仍是 governance-only

## 今天还不能算完成的能力

当前仓库还没有这些正式能力：

- production deployment package
- production secret backend
- process supervisor 与环境隔离
- 外部 provider health check
- 真正的本地 console 前端工程
- 更完整的 route-level smoke、stream 组合和兼容性矩阵
- durable session/checkpoint/audit/result store、materialized checkpoint/result reader 和 recovery runbook
- 真实工具执行器、materialized tool result cache、上层确认流接线和完整 session/tooling 负向回归 implementation consumer

所以如果你问“现在怎么部署”，准确答案是：当前已有本地 CLI runtime、进程内 gateway、Go platform service、本地 runbook、启动 wrapper、config / deployment / diagnostics smoke、request observability、error taxonomy、bridge-backed provider/profile discoverability、`GET /v1/platform/overview` 只读产品 overview、overview consumer smoke、session/tooling metadata smoke、P2 design gates 和 P2 governance rollup checks，但还没有完整 production deployment 面、本地 console 前端工程、真实 executor、durable store、confirmation 接线、materialized result reader、长期记忆、业务写回或 replay。

## 读文档顺序

如果你刚接触这个仓库，建议按这个顺序读：

1. [文档入口](README.md)
2. 项目总览与使用指南
3. [当前推进焦点](radishmind-current-focus.md)
4. [产品范围](radishmind-product-scope.md)
5. [战略定义](radishmind-strategy.md)
6. [能力矩阵](radishmind-capability-matrix.md)
7. [系统架构](radishmind-architecture.md)
8. [跨项目集成契约](radishmind-integration-contracts.md)
9. [脚本目录说明](../scripts/README.md)
10. [数据集目录说明](../datasets/README.md)
11. [训练目录说明](../training/README.md)

## 默认不要做

- 不把 `RadishMind` 做成上层业务真相源
- 不默认把 builder/guided 结果当成 raw 晋级证据
- 不在上层项目还没具备真实挂载点时继续细化假想接线
- 不把 production deployment、session、tooling 或完整外部兼容矩阵写成“已经具备”
- 不默认下载大模型、数据集或权重

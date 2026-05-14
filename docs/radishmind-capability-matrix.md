# RadishMind 能力矩阵

更新时间：2026-05-14

本文档用于回答三件事：

- `RadishMind` 现在已经具备什么能力
- 还缺哪些关键能力
- 接下来应该先补哪一块，而不是继续在旧实验或假想接线上打转

如果你要先理解“项目为什么这样分层”，先读 [战略定义](radishmind-strategy.md)。

| 主线 | 当前已有 | 当前缺口 | 当前不做 | 下一步 |
| --- | --- | --- | --- | --- |
| `Runtime Service` | `scripts/run-copilot-inference.py`、`services/gateway/copilot_gateway.py`、`services/platform/` Go platform service、`RadishFlow` gateway demo、service smoke matrix、平台级 `ops smoke`、本地启动 runbook、启动 wrapper、JSON 配置层级、deployment smoke、structured diagnostics / failure boundary、provider/profile discoverability 对齐、request-level observability、error taxonomy；`P1` 已达到 short close | production secret backend、process supervision、环境隔离、可发布部署包 | 不承诺正式 production deployment，不接外部认证系统，不直接写回上层真相源，不继续把 provider/config/diagnostics/observability 同层细节无限扩张 | 切到 `P2 Session & Tooling Foundation` |
| `Southbound Model Access` | `services/runtime/provider_registry.py` 已落地最小 `provider registry` 骨架，并收口当前 `mock`、`openai-compatible`、`HuggingFace`、`Ollama` 与 `openai-compatible chat / gemini-native / anthropic-messages` 分流；`HuggingFace` / `Ollama` 已进入 profile inventory、diagnostics 和 readiness metadata；`run-radishmind-core-candidate.py` 已具备 `local_transformers` 本地模型实验入口 | 外部 provider health checks、production secret backend、retry/fallback/cost/latency policy | 不把单一 provider 或单一模型绑定成平台唯一方向，不把实验 wrapper 直接当正式服务能力 | 在现有 registry / diagnostics 骨架上继续补 health check、secret backend 和调用策略 |
| `Northbound API Compatibility` | internal canonical `CopilotRequest / CopilotResponse / CopilotGatewayEnvelope`、进程内 gateway、CLI runtime、`/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models` 第一版 bridge、SSE 兼容骨架、bridge-backed provider/profile inventory、request-side provider/profile selection、bridge 增量流式转发、`GET /v1/models/{id}` 精确 lookup、显式 provider/model 隔离；`/v1/models`、request selection 与 diagnostics selectable model ids 已共享 `profile:<profile>` / `provider:<provider>:profile:<profile>` 口径；三种 northbound 协议已共享 `request_id`、route、selection metadata、latency 日志和 error envelope | 更多 northbound / southbound stream 组合、session-aware route smoke | 不让兼容层演化成第二套业务真相源，不为每个协议单独复制一套核心业务逻辑 | 暂不把更多协议组合当作当前主线，先进入 session/tooling |
| `Conversation & Session` | request 已支持 `conversation_id`；`session-record`、recovery checkpoint record/manifest/read result、northbound session metadata、checkpoint metadata-only route smoke、denied query fixture、readiness summary、implementation preconditions、negative regression skeleton、independent audit records design、result materialization policy design 和 `session-tooling-foundation-status-summary.json` 已有首版门禁；当前为 `close candidate / governance-only` | durable session store、长期记忆、真实 checkpoint storage backend、materialized result reader、跨轮恢复执行器、完整 `negative_regression_suite` | 不做共享长期记忆，不做隐式自治规划循环，不把 checkpoint read route smoke 写成真实恢复 API 或 result reader，不把 close candidate 写成 P2 short close | 保持只读暴露和治理门禁，等待 executor/storage/confirmation 前置条件满足 |
| `Tooling Framework` | task-local 的检索、合法候选生成、deterministic builder、`tool_hints`；`tool`、registry、audit record、session binding、metadata-only result cache、checkpoint read `tool_audit_summary`、promotion gate 分层、负向消费 summary、route smoke coverage summary、readiness summary、implementation preconditions、negative regression skeleton、confirmation flow design、independent audit records design、result materialization policy design 和 close candidate status summary 已有首版契约与快速门禁 | 真实工具执行器、tool result materialization、durable result store、durable tool store、上层确认流接线、durable audit store、完整负向回归 | 不做 unrestricted tool calling，不把所有逻辑塞进模型提示词，不在 registry v1 后直接接 executor，不把 skeleton、audit design 或 materialization policy design 当成已满足负向回归 | 继续把 tooling audit 与 recovery checkpoint 读取边界对齐；真实执行前先补确认流接线、executor 边界、storage 设计和 durable audit 策略 |
| `Evaluation & Governance` | schema、offline eval、candidate record、review record、`check-repo`、service smoke、runtime provider dispatch smoke、platform config / deployment / diagnostics / runbook checks、platform request observability 单元测试、session/tooling/checkpoint contract smoke、checkpoint read route smoke、readiness summary check、implementation preconditions check、negative regression skeleton check、confirmation flow design check、independent audit records design check、result materialization policy design check 和 foundation status summary check | 平台级 promotion checklist、完整 session/tooling 负向回归、真实实现前的 durable audit / result store gate | 不用纯机器指标决定晋级，不把真实大产物直接提交入仓，不把 metadata smoke 升级成真实 store/executor，不把 governance-only close candidate 写成实现完成 | 维持 P2 close candidate 盘点，后续只在前置条件满足后进入真实设计 |
| `Model Adaptation` | raw/repaired/injected/guided/builder 轨、training sample conversion、训练治理文档 | 平台对齐的基座目标、晋级门槛、蒸馏计划、训练 runbook | 不做大规模训练放量，不把 `14B/32B` 设为默认目标，不把 builder/guided 写成 raw 晋级 | 等平台主线稳定后，再锁定 v1 训练路线 |
| `Image Path` | image intent、backend request、artifact schema、最小评测 manifest | 真实 backend adapter、artifact 返回链路、image safety runbook | 不下载图片模型，不直接接真实生图 backend | 先补 image adapter handshake 和 safety gate 文档 |
| 上层项目接入 | `RadishFlow` service/API 门禁冻结，`Radish` docs QA 资产具备，`RadishCatalyst` 文档预留 | 上层挂载点、确认流、命令承接接口、真实审计落点 | 不继续细化假想接线，不改外部仓库，不补不存在的 UI/命令承接层 | 等上层准备就绪后，只选一个切片真实接入 |

## 现阶段判断

- `P1 Runtime Foundation` 已经从“应该先补平台本体”推进到 short close。
- 当前最优先的缺口已经从“请求级观测 + 错误分类”转为 `P2 Session & Tooling Foundation`。
- `M3` 与 `M4` 已经提供了足够的冻结证据，说明仓库不是“什么都不能做”，而是应该把这些证据转化为正式平台能力。
- 如果后续仍然只围绕样本、prompt、假想接线，或 provider/config/diagnostics 同层细节打转，就会再次偏离主线。

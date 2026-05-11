# RadishMind 能力矩阵

更新时间：2026-05-10

本文档用于回答三件事：

- `RadishMind` 现在已经具备什么能力
- 还缺哪些关键能力
- 接下来应该先补哪一块，而不是继续在旧实验或假想接线上打转

如果你要先理解“项目为什么这样分层”，先读 [战略定义](radishmind-strategy.md)。

| 主线 | 当前已有 | 当前缺口 | 当前不做 | 下一步 |
| --- | --- | --- | --- | --- |
| `Runtime Service` | `scripts/run-copilot-inference.py`、`services/gateway/copilot_gateway.py`、`RadishFlow` gateway demo、service smoke matrix | 长驻服务、HTTP/worker 包装、配置分层、启动 runbook、部署 smoke | 不承诺正式 production deployment，不接外部认证系统，不直接写回上层真相源 | 先收口本地 service bootstrap、配置约定和最小调用路径 |
| `Southbound Model Access` | `services/runtime/provider_registry.py` 已落地最小 `provider registry` 骨架，并收口当前 `mock`、`openai-compatible` 与 `openai-compatible chat / gemini-native / anthropic-messages` 分流；`run-radishmind-core-candidate.py` 已具备 `local_transformers` 本地模型实验入口 | 正式 `HuggingFace` 服务接入、`Ollama` adapter、更完整的 provider capability discovery、统一 auth/config 分层 | 不把单一 provider 或单一模型绑定成平台唯一方向，不把实验 wrapper 直接当正式服务能力 | 在现有 registry 骨架上先补 northbound 兼容与 `HuggingFace` / `Ollama` |
| `Northbound API Compatibility` | internal canonical `CopilotRequest / CopilotResponse / CopilotGatewayEnvelope`、进程内 gateway、CLI runtime、`/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models` 第一版 bridge、SSE 兼容骨架、bridge-backed provider/profile inventory | request-side provider 选择、更细粒度的 model discovery、兼容性 smoke 的路由覆盖与协议细节收敛 | 不让兼容层演化成第二套业务真相源，不为每个协议单独复制一套核心业务逻辑 | 继续补 upstream token 级流式转发、request-side provider 选择和更细粒度 model discovery |
| `Conversation & Session` | request 已支持 `conversation_id`，`RadishFlow` snapshot 带局部会话语义 | session schema、history policy、恢复记录、跨轮审计、会话级 smoke | 不做共享长期记忆，不做隐式自治规划循环 | 先定义 session contract、fixture 和最小恢复/审计边界 |
| `Tooling Framework` | task-local 的检索、合法候选生成、deterministic builder、`tool_hints` | 通用 tool contract、registry、timeout/retry/policy、tool audit | 不做 unrestricted tool calling，不把所有逻辑塞进模型提示词 | 先做本地 tool contract 与 registry 原型 |
| `Evaluation & Governance` | schema、offline eval、candidate record、review record、`check-repo`、service smoke | runtime/session/tooling/deployment 级门禁、平台级 promotion checklist | 不用纯机器指标决定晋级，不把真实大产物直接提交入仓 | 先扩展 smoke gate 到 runtime、session、tooling 主线 |
| `Model Adaptation` | raw/repaired/injected/guided/builder 轨、training sample conversion、训练治理文档 | 平台对齐的基座目标、晋级门槛、蒸馏计划、训练 runbook | 不做大规模训练放量，不把 `14B/32B` 设为默认目标，不把 builder/guided 写成 raw 晋级 | 等平台主线稳定后，再锁定 v1 训练路线 |
| `Image Path` | image intent、backend request、artifact schema、最小评测 manifest | 真实 backend adapter、artifact 返回链路、image safety runbook | 不下载图片模型，不直接接真实生图 backend | 先补 image adapter handshake 和 safety gate 文档 |
| 上层项目接入 | `RadishFlow` service/API 门禁冻结，`Radish` docs QA 资产具备，`RadishCatalyst` 文档预留 | 上层挂载点、确认流、命令承接接口、真实审计落点 | 不继续细化假想接线，不改外部仓库，不补不存在的 UI/命令承接层 | 等上层准备就绪后，只选一个切片真实接入 |

## 现阶段判断

- 现在最该补的是平台本体，而不是继续深挖同一批 `M4` 实验。
- 其中最优先的缺口已经明确是“外部模型接入 + 多协议兼容”，因为这决定了平台能否真正成为平台，而不是只服务单一实验链路。
- `M3` 与 `M4` 已经提供了足够的冻结证据，说明仓库不是“什么都不能做”，而是应该把这些证据转化为正式平台能力。
- 如果后续仍然只围绕样本、prompt 或假想接线打转，就会再次偏离主线。

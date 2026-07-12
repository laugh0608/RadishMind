# User Workspace Application API Integration & Invocation v1

更新时间：2026-07-12

状态：`application_api_integration_invocation_v1_complete`

## 当前实现结果

2026-07-12 已完成 Application Detail 内的独立 lazy Integration Workspace、application-scoped `/v1/models` consumer、严格目录校验、loading / ready / empty / failure / refresh、三协议 × 三语言示例，以及到现有 Playground / Request History 的动态 scope handoff。没有新增 API、Gateway schema、repository、provider registry、协议适配器或 SSE parser。

Web 单测覆盖 offline 零请求、目录成功 / 空 / HTTP failure / 非法响应、workspace / application caller scope、九种示例与 secret guard、Playground handoff、request-id History handoff 和 application reset；44 项 Web 测试与 production build 通过。新 Integration chunk 为 11.82 KiB，Playground / History chunk 为 5.33 / 17.90 KiB，主入口为 431.36 KiB，均通过现有预算。

真实浏览器以 `app_docs_assistant` 完成 6 个模型加载、Messages / TypeScript 示例、Messages unary、Messages stream、用户取消和精确 History detail。取消请求在 Playground 显示 `gateway_playground_request_canceled / client / HTTP not observed`，同一 request id 的 sanitized History detail 显示 `408 / BRIDGE_WORKER_CANCELED / python_bridge`、`app_docs_assistant` 与 `gateway_request_record.v1 · version 3 · memory_dev`。控制台为 0 error / 0 warning，URL 只有稳定 section hash，localStorage / sessionStorage 均为空。

## 功能目标

让内部开发者从 User Workspace 的 Applications 选择一个 application 后，在同一 application detail 上完成模型发现、接入示例生成、测试调用和调用审查。该路径消费现有 `/v1/models`、Gateway Playground 与 sanitized Request History，不新增 northbound route、协议适配器、SSE parser、请求状态模型、repository 或 provider registry。

本功能仍属于内部开发者预览。dev/test 调用成功只证明当前本地 Gateway 路径可交互、可取消和可审查，不代表 production API 分发、真实 provider SLA、production auth、key lifecycle、quota 或 billing 已成立。

## 目标用户与连续流程

目标用户是需要把某个 workspace application 接到 RadishMind Gateway，并在提交真实集成改动前验证协议和审查调用记录的内部开发者。

1. 用户从 Applications 列表选择 application，并进入现有 Application Detail。
2. Application API Integration Workspace 显示当前 workspace / application scope；application 切换会立即清空旧模型、旧示例选择、旧失败和旧调用交接状态。
3. 默认 offline 模式只显示能力说明并保持零网络请求。显式 dev/test 模式下，用户主动加载或刷新 `/v1/models`。
4. 模型目录完成 loading、ready、empty 或 failure 状态转换；只有通过响应校验的 model id 才能进入选择器。
5. 用户选择 model、Chat Completions / Responses / Messages 和 cURL / Python / TypeScript，生成只含环境变量占位符的接入示例。
6. 用户将当前 application、protocol 和 model 交给现有 Gateway Playground；Playground 继续负责 unary、stream、取消、稳定失败和当前内存态输入输出。
7. 调用完成后，Playground 用同一 `request_id` 与 application scope 交给现有 Request History，并精确打开 sanitized detail。

## 数据来源与职责

| 数据 | 真相源 / owner | 本功能消费方式 | 不允许的替代来源 |
| --- | --- | --- | --- |
| application 列表与当前选择 | 现有 User Workspace Applications / workspace selection | 只消费当前 `applicationRef` 与展示名 | 不新增 application repository，不使用固定 application 配置覆盖当前选择 |
| workspace scope | 现有显式 dev/test Gateway 配置 | 模型目录、Playground 与 History 使用同一 `workspaceId` | 不从 URL、浏览器存储或示例文本恢复 scope |
| 模型目录 | 现有 `GET /v1/models` | 严格读取 `object=list` 与 `data[]` 的公开模型字段 | 不复制 provider registry，不从离线 evidence 或 application fixture 伪造 live 目录 |
| 协议调用 | 现有 Gateway Playground consumer | 通过 application handoff 复用三协议 builder、响应适配、SSE parser、abort 和状态模型 | 不在 Integration Workspace 新建第二套 fetch / parser / request state |
| 调用审查 | 现有 Gateway Request History | 通过同一 request id、workspace、consumer 与 application scope 精确读取 detail | 不把 Playground 输入输出补写到 history，不把旧 quota / usage evidence 当调用记录 |

## Application scope

- Integration Workspace 的 application 只能来自当前 Applications 选择，不能使用 launcher 中的固定 application 作为当前业务选择。
- `/v1/models` dev/test 请求必须携带当前 tenant、workspace、consumer、application、subject、scope 和 request id caller headers；这些内部 header 只存在于 consumer，不进入 UI、示例或持久化介质。
- Playground handoff detail 固定包含 `applicationId`、`protocol` 和 `model`，不包含输入、输出、credential 或完整 headers。
- Playground 发出的 northbound request 必须使用 handoff application 覆盖静态默认值；History handoff 必须继续使用同一 application scope，否则 detail read fail closed。
- application 切换时，正在加载的目录请求应被取消或结果被丢弃；新 application 不得显示旧 application 的 models、示例选择、失败或 request id。

## 模型目录状态与响应校验

模型目录状态固定为：

- `offline`：未启用显式 dev/test source；不调用 fetch。
- `idle`：已启用，但用户尚未加载当前 application 的模型。
- `loading`：当前 application 的 `/v1/models` 请求进行中；允许显式取消或被 application 切换淘汰。
- `ready`：响应通过校验且至少包含一个 model。
- `empty`：响应通过校验但 `data` 为空。
- `failed`：HTTP、network、JSON、schema、correlation 或 forbidden projection 校验失败。

响应只接受 `object: "list"`、`data: Model[]`，每个 model 必须含单行、非空且长度受限的 `id`，并校验 `object: "model"`、非负 `created` 与字符串 `owned_by`。UI 只保留 model id、owner 和公开协议能力；忽略 provider credential 状态、endpoint 与未知 metadata。顶层或 model 公共投影出现 credential、authorization、secret、raw error、endpoint 或 header payload 时 fail closed。

HTTP 非成功响应保留 `gateway_model_catalog_http_failed`，非法 JSON / envelope 保留 `gateway_model_catalog_response_invalid`，网络失败保留 `gateway_model_catalog_network_error`；UI 只显示稳定 code 与 sanitized summary，不显示 response body、stack、endpoint 或 header。

## 接入示例

协议固定为：

| 协议 | route | request shape |
| --- | --- | --- |
| Chat Completions | `/v1/chat/completions` | `model + messages[] + stream` |
| Responses | `/v1/responses` | `model + input + stream` |
| Messages | `/v1/messages` | `model + messages[] + max_tokens + stream` |

每种协议提供 cURL、Python 和 TypeScript 示例。示例约束：

- base URL 与 key 只使用 `${RADISHMIND_BASE_URL}`、`${RADISHMIND_API_KEY}` 或对应进程环境变量引用。
- model 使用当前经过目录校验的 id；临时输入使用 `${RADISHMIND_INPUT}` 或运行时变量，不嵌入 Playground 当前输入。
- 不显示真实 key、key hash、Authorization 值、cookie、secret ref、provider endpoint 或任何 `X-RadishMind-Dev-*` caller headers。
- 示例不是 credential issuance 或 production auth 文档；若环境变量缺失应由示例运行环境显式失败，不提供默认 secret。
- 示例纯函数生成，只存在当前组件内存，不写 clipboard history 以外的仓库或业务介质；v1 不提供保存、分享、导出或生成 application 配置文件。

## Playground 与 Request History 交接

- Integration Workspace 只派发 application / protocol / model handoff，并导航到现有 Playground。
- Playground 接收 handoff 后重置旧 input/output/result，保留默认临时输入模板；用户显式选择 unary 或 stream 并发送。
- Playground 继续使用既有 request id 生成、request builder、response adapter、SSE terminal parser、输出预算、abort 和稳定 failure mapping。
- 成功、失败或取消后，History handoff 只派发 `requestId` 与 `applicationId`。History 清空旧筛选，以对应 scope 读取列表和 detail，并精确展示同一 request id。
- History 不可用或 scope 不匹配时显示 handoff failure；Playground 当前结果仍留在内存，不伪造已持久化或已审查。

## 隐私与持久化边界

- 默认 offline 必须为零 fetch；离线模型目录为空且不回退 committed fixture。
- 模型选择、协议、语言、临时输入、当前输出、stream delta、失败与 request id 只存在于当前 React 组件内存。
- 上述值不得写入 URL query/hash payload、localStorage、sessionStorage、draft、workflow run、Gateway history payload、artifact、日志、周志或业务真相源。
- URL 只使用稳定 section hash，不包含 application、model、protocol、request id、输入或输出。
- Gateway Request History 继续只保存既有 sanitized operational metadata；本功能不改变其 schema、retention 或 repository。

## 实施拆分

1. 建立独立 lazy Integration Workspace panel 与纯 consumer：config、模型目录校验、示例生成和 application reset state。
2. 扩展现有浏览器内 Playground handoff event，使 panel 可接收 application / protocol / model；不复制调用逻辑。
3. 让现有 Playground config 在每次提交时绑定当前 application，并在 handoff / application 切换时清空旧状态。
4. 扩展现有 History handoff，使 detail read 使用调用对应的 application scope；不改变 backend API。
5. 补 consumer 单测、production build、必要 Go 测试、真实浏览器连续流程与存储泄漏检查。

## 验收方式

单元测试至少覆盖：

- offline 模式零 fetch；
- 模型目录成功、空结果、HTTP failure、network failure 和非法响应；
- `/v1/models` 与 Playground request 携带当前 workspace / application scope；
- 三协议 × 三语言示例的 route / request shape 与 secret 脱敏；
- application / protocol / model Playground handoff；
- request id / application History handoff；
- application 切换后旧目录、选择、失败和 request id 不进入新状态。

真实浏览器必须在显式 dev/test 配置下完成：选择 application、加载模型、切换协议与语言生成示例、至少一次 unary、至少一次 stream、取消请求、按同一 request id 精确打开 History detail，并检查控制台以及 URL、localStorage、sessionStorage 不含输入输出、key、application handoff payload 或 request id。

相称验证固定为 Web 单测、production build、与 scope / `/v1/models` 直接相关的 Go 测试、`git diff --check` 和 `./scripts/check-repo.sh --fast`。若实现中实际改变阶段边界、协议、schema 或治理真相源，再补完整 `./scripts/check-repo.sh`。

## 停止线

- 不实现真实 API key 生命周期、key issuance / rotation / revoke、production auth、quota enforcement、rate limit、billing 或 cost ledger。
- 不新增 northbound API、Gateway schema、request history schema / repository、provider registry、protocol adapter 或 SSE parser。
- 不实现自动 retry / fallback、load balancing、动态 provider route 或 production deployment。
- 不实现 application 创建、编辑、发布、删除或写入。
- 不保存输入输出，不实现 prompt history、分享、导出、批量调用或后台任务。
- 不把 dev/test 调用成功、模型目录可读或 History detail 存在解释为 production ready、真实 provider SLA 或可信 reported usage。
- 不扩展 Workflow tool、confirmation、writeback、replay 或 resume。

# Gateway Playground / Request Review Loop v1

更新时间：2026-07-12

状态：`gateway_playground_request_review_loop_v1_complete`

## 功能目标

让内部开发者在 RadishMind Web 中完成一次真实、可取消、可审查的 Gateway 请求，而不再依赖终端命令或浏览器调试脚本。用户可以选择现有 northbound 协议、输入模型和临时文本、选择 unary 或 stream，查看当前响应，并跳转到同一 `request_id` 的 Request History 详情。

本功能复用已完成的 Gateway runtime、三个 northbound route、显式 dev/test caller context 和 Request History。它不新增 API、schema、repository 或 provider contract，不把开发调用入口解释为 production API 分发。

## 当前实现结果

2026-07-12 已完成独立 lazy consumer / panel、三协议 unary / stream adapter、SSE terminal parser、用户 abort、稳定失败、20,000 字符输出预算、offline 零请求、双端 launcher 配置和 request-id history handoff。后续 Application API Integration 在不复制调用逻辑的前提下扩展 application / protocol / model handoff，Playground 请求与 History detail 现在都使用当前 application scope。Web consumer 单测覆盖三协议 request / response、caller scope、SSE、HTTP failure、correlation mismatch、malformed response、invalid input、abort、output budget、network failure和动态 application scope。

真实浏览器验证 Chat Completions 与 Responses unary、Messages stream、页面按钮取消和 history handoff。Playground 只把本地 abort 表达为 `client / gateway_playground_request_canceled / HTTP not observed`；同 request id 的 PostgreSQL detail 提供服务端事实 `408 / BRIDGE_WORKER_CANCELED / python_bridge`，record 为 `gateway_request_record.v1 · version 3 · postgres_dev_test`。URL、localStorage 与 sessionStorage 不含输入输出，全新会话为 0 error / 0 warning。功能 v1 已关闭，下一批不追加 prompt history、分享、导出或 production key / quota / billing。

## 目标用户与主路径

目标用户是调试模型路由、兼容协议和响应行为的内部开发者。

1. 用户从 Model Gateway 导航进入 Playground。
2. 选择 Chat Completions、Responses 或 Messages，输入 model 和临时请求文本，可选择 stream。
3. Web 生成稳定、可关联的 `request_id`，通过现有显式 dev/test caller scope 调用对应 northbound route。
4. unary 成功后展示协议适配后的文本；stream 逐段展示 delta，并允许用户取消仍在进行的请求。
5. 失败时只展示 HTTP status、稳定 failure code、failure boundary 和 sanitized message，不显示 provider raw error、endpoint 或 credential。
6. 成功、失败或取消后，用户可以跳转到 Request History；History consumer 用同一 `request_id` 刷新并打开 sanitized detail。

## 可打开范围

- Web 独立 lazy Playground panel、consumer、协议 request builder、response adapter 和 SSE parser。
- 现有 `/v1/chat/completions`、`/v1/responses`、`/v1/messages` 的 unary / stream 调用。
- 显式 dev/test caller headers、request id、abort signal 和 Request History handoff event。
- component-memory 内的临时输入、当前输出、稳定失败与最近一次关联信息。
- consumer unit tests、Web build / chunk budget、真实浏览器三协议与 history handoff 验证。

## 数据与安全边界

- 输入文本只存在于当前组件状态和 northbound HTTP request；不写 localStorage、sessionStorage、URL、日志、Request History、fixture、周志或 committed evidence。
- 当前输出只存在于组件内存；不写数据库、业务真相源、run record、artifact store 或浏览器持久存储。
- consumer 只允许返回 `outputText`、request id、route、protocol、stream、HTTP status、稳定 failure 和 completion state。
- authorization、API key、cookie、credential、secret ref、endpoint、provider raw envelope、raw error、完整 headers 和 stack trace 不进入 UI state。
- model 最大 160 字符且禁止换行；输入必填，最大 8,000 字符；累计输出最大 20,000 字符，超限时 fail closed。
- request id 由 Web 生成，并要求响应关联 header 缺失或与请求一致；不接受响应改写为另一 request id。

## 协议映射

| Playground 协议 | route | request shape | unary output | stream delta |
| --- | --- | --- | --- | --- |
| Chat Completions | `/v1/chat/completions` | `model + messages[] + stream` | `choices[0].message.content` | `choices[].delta.content` |
| Responses | `/v1/responses` | `model + input + stream` | `output_text` | `response.output_text.delta.delta` |
| Messages | `/v1/messages` | `model + messages[] + max_tokens + stream` | `content[].text` | `content_block_delta.delta.text` |

三种响应都必须经过 allowlist adapter。成功 HTTP 若缺少可识别文本、返回错误 envelope 或 request id 不一致，统一进入 `gateway_playground_response_invalid`，不回退原始 JSON dump。

## 状态与失败语义

Playground 状态固定为：

- `offline`：未启用显式 dev/test source，零 northbound 请求。
- `idle`：可编辑，尚未提交。
- `submitting`：请求或 stream 正在进行，可取消。
- `succeeded`：unary 完成或 stream 收到 terminal event。
- `failed`：输入、网络、HTTP、协议或 response adapter 失败。
- `canceled`：用户主动 abort，或浏览器以 `AbortError` 结束当前请求。

稳定 consumer failure 至少包括：

- `gateway_playground_disabled`
- `gateway_playground_input_invalid`
- `gateway_playground_network_error`
- `gateway_playground_request_canceled`
- `gateway_playground_response_invalid`
- `gateway_playground_output_too_large`
- northbound 返回的现有稳定 failure code

失败不得自动 retry、fallback、改协议、改 provider、改 model 或保存输入。用户可显式修改后重新提交，新提交使用新的 request id。

## Request History 交接

- Playground 只发送 `request_id` 的浏览器内事件，不发送输入或输出。
- Request History panel 收到事件后清除旧过滤，重新读取第一页，并直接读取对应 detail。
- History detail 仍以 Platform sanitized record 为真相源；Playground 当前输出不能补写或合并进 history。
- history 不可用时保留 Playground 结果，并显示 review handoff 失败；不得伪造记录已持久化。

## Offline 与 dev/test 模式

默认 offline：panel 显示能力说明与停止线，不发起任何请求。

显式 dev/test source：

```text
VITE_RADISHMIND_GATEWAY_PLAYGROUND_SOURCE=dev-gateway-playground-http
VITE_RADISHMIND_GATEWAY_PLAYGROUND_MODEL=radishmind-local-dev
```

base URL、tenant、workspace、consumer、application 和 subject 复用 Gateway Request History 的显式开发配置。launcher 只在 Gateway Request History dev/test 已启用时打开 Playground，确保 request 与 review 使用同一 caller scope。

## 实施拆分

本功能不新增 API、schema 或新的执行边界，不创建专项 task card 或 checker。首个纵向切片直接包含：

1. consumer config、输入校验、三协议 request / response adapter、SSE parser 与 abort。
2. 独立 lazy panel、offline / idle / submitting / success / failure / canceled UI。
3. Request History request-id handoff 与精确 detail 打开。
4. launcher 配置、consumer tests、Web build、真实浏览器和仓库门禁。

## 验收

- offline 模式零 fetch。
- 三种协议 unary request body、caller headers、request id 和输出映射正确。
- 三种协议 stream delta 可解析，terminal event 后成功，用户取消后进入 canceled。
- 非法输入、HTTP failure、malformed response、request id mismatch、output budget 和 network failure fail closed。
- 输入 / 输出不进入 URL、浏览器持久存储或 Request History detail。
- 成功、失败和取消均可按同一 request id 打开真实 history detail。
- lazy chunk 通过现有 64 KiB 普通 chunk 预算；不扩大 `App.tsx` 业务逻辑。

## 停止线

- 不实现 production API key、OIDC Gateway auth、quota、rate limit、billing、cost ledger 或 key issuance。
- 不实现自动 retry / fallback、动态 route、load balancing、provider secret 编辑或 production deployment。
- 不保存 prompt、response、stream delta、完整 request / response JSON 或可逆摘要。
- 不把 Playground 输出写入 Workflow run、evaluation、artifact、audit 或上层业务真相源。
- 不实现会话历史、prompt library、分享链接、导出、批量运行或后台任务。
- 不因 Playground 成功声明真实 provider SLA、production Gateway 或 reported usage ready。

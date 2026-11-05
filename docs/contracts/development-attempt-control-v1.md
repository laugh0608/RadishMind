# 受控开发尝试：工具通道、委托、事件与取消 / 预算合同 v1

更新时间：2026-11-05

状态：单尝试领域规则、SQLite 持久状态 / 提议 / 事件事务、当前资格读取、任务凭据与独立回环通道已实现并以固定响应验证。默认未启用；受限 SDK Agent 工作进程 / 事件、模型等待取消及所属 SDK 进程组终止已固定响应验证；固定 Go 探针代码沙箱、受限工具与归属清理现已真实验收，SDK 收尾接线已定向验证；真实 SDK 与固定代码容器同次任务的候选 / 测试 / 产物及命令中取消已联合验证；原生 HTTP 已通过固定上游验证，真实厂商资格及通用沙箱仍待收口。所属功能见[主专题](../features/workflow/controlled-software-development-team-v1.md)，实施记录见[集中任务卡](../task-cards/development-task-probe-v1-plan.md)。

## 接入与默认状态

`Server.ConnectDevelopmentControl(backend)` 是可信 Go 宿主的显式接线入口，只接受共享 SQLite 与明确的 `NativeBackend`。正常平台 / 工作台构造不调用它，不新增自动启用配置。连接后也须显式调用 `StartDevelopmentChannel(ctx)`，仅监听 `127.0.0.1:0`；地址交给可信 SDK 工作进程，不写死验收端口。

| 独立内部入口 | 职责 | 当前边界 |
| --- | --- | --- |
| `POST /internal/development/v1/attempts/{attempt_id}/chat/completions` | 严格原生模型请求，先提交预留 / 派发记录，再调用后端 | 单一非流式 HTTP 适配已通过固定上游验证；真实厂商资格待验，未配置后端则拒绝 |
| `POST /internal/development/v1/attempts/{attempt_id}/tool-calls` | 精确回显已登记的工具提议，事务内消费并预留 | 返回 `202 queued`；不执行文件或命令，不返回伪造观察值 |
| `GET /internal/development/v1/attempts/{attempt_id}/tool-offers/{tool_call_id}` | 读取 Go 分配的工具请求身份及摘要 | 不含工具正文 |
| `GET /internal/development/v1/attempts/{attempt_id}/tool-catalog` | 读取本通道唯一冻结工具目录 | 不开放 SDK 默认终端、MCP 或递归委派 |
| `GET /internal/development/v1/attempts/{attempt_id}/events` | `after_sequence` / `limit` 增量读取 | 最多 50 项，显式 `next_sequence / has_more / record_version`；读取不触发执行 |

这些路由不注册到公开 `/v1` 或主监听器，不扩大旧建议型 Agent Profile 的 30 秒边界。HTTP 同时核对回环来源与 Host、拒绝 Origin；浏览器 Session、Cookie、长期 API key 和自报身份均不能替代任务 bearer。没有 SDK 自报权威事件、签发或取消成功的入口。正式 Web 启动 / 取消 / 结果资源接口仍待阶段 1。

模型请求只接受 `model / messages / max_completion_tokens / tools` 及显式为 `false` 的 `stream / parallel_tool_calls`。未知、重复、空值、大小写别名、尾随 JSON 拒绝；总请求最多 262144 字节、消息最多 64 条、单条文本最多 65536 UTF-8 字节。消息仅文本及原生 assistant `tool_calls` / tool `tool_call_id`；不解析答案正文里的 JSON。模型必须精确匹配冻结模型名，单响应最多一个工具调用，多个调用整体拒绝。

`tools` 精确匹配目录，仅 `development_tool`，动作是 `read_file / replace_file / run_tests`。目录内参数来自 canonical schema 的内嵌分发副本，Go 测试逐字节对比源文件。旧 SDK 探针的 `development_probe` 保持原语义；受限 SDK 已显式关闭隐式重试、fallback、压缩和 critic，以固定响应验证代表任务的全部模型调用经过此通道；不外推到未开放的 SDK 插件或执行档案。

## 原生 HTTP 档案

可信宿主可显式调用 `Server.ConnectDevelopmentNativeHTTP(ctx, policy)`，只连接 Go 平台配置中已经明确的一个 `openai-compatible` 提供方 / 档案 / 模型。此方法复用 `buildBridgeEnvelopeOptions` 的地址与凭据来源，并核对当前目录；不解析 Python 环境中的其他档案，不使用路由回退，不从 SDK 请求取地址或密钥。配置中的 `api_key_file` 沿既有 `ResolveWorkbenchCredential` 启动时读取语义，不新增轮询、热更新或 secret backend。

`NativeHTTPConnection` 只在可信 Go 内存传递，不属于公开 JSON 合同。构造时冻结模型引用、实际模型名、地址、密钥与 `NativePolicy`；策略读取和实际派发均重读当前配置，任何变化或目录不匹配均拒绝。账户、应用、模型目录和价格修订继续由原尝试资格规则核对；正常工作台构造不调用接线入口，不自动授予真实模型资格。

- 地址由平台配置明确提供，路径追加 `/chat/completions`；只接受 HTTPS，拒绝用户信息、查询、片段及不明确路径。TLS 最低 1.2，证书验证保持开启；不使用环境代理、重定向、连接复用或自动重试 / 回退。
- DNS 返回地址全部校验并选择一个已校验地址直接拨号，拒绝私网、回环、链路本地及保留地址；不进行第二次主机解析。仅 `Options.TestOnly` 的固定上游测试允许有明确端口的 `http://127.0.0.1`，不是普通运行配置的外连例外。
- 请求显式发送 `stream=false / parallel_tool_calls=false`，避免提供方缺省开启并行工具。输入仍严格消费本通道唯一工具目录；正文与响应均最多 262144 字节，响应头最多 16384 字节，总期限由冻结策略限定且不超过两分钟。
- 响应为单个 `chat.completion`、单个索引 0 的 choice，模型名必须精确相同；结束原因仅 `stop / tool_calls`。正文只允许文本或工具调用时的空值；工具名称、调用标识与参数均验证，工具身份和观察值继续绑定尝试上下文。未知模式、重复字段、大小写别名、尾随值、超过 24 层的 JSON 及非法 / 超限响应均拒绝。
- `usage` 缺失或为 `null` 时保持未知；存在时三个 token 计数必须齐全、非负整数且总数相符。不完整用量或非法响应返回失败，并保留已派发请求的未知预留，不以零补齐。提供方额外用量细项不参与平台核账。
- 错误只返回固定类别，如 `native_authentication_failed / native_rate_limited / native_upstream_failed / native_redirect_denied / native_timeout / native_canceled / native_response_invalid`；不包含地址、密钥或上游响应原文。尝试及请求归因继续由现有预留 / 回执 / 事件保存，不另建用量账本。
- 取消中断本地 HTTP 连接，已派发请求保留未决预留；本地连接关闭不证明远端计算停止或不计费。其他后端错误继续映射为原有通用安全错误。

固定上游已经验证三次模型请求、两轮工具提议 / 合成观察值往返及错误 / 取消；真实 SQLite 身份、应用、价格和凭据文件接线同时参与测试。该证据不证明实际厂商 / 模型兼容、计量或输入 token 上界；`InputTokenCeiling` 当前只能用于经核对的固定响应试验，恢复真实模型前须为精确档案补足完整输入计量依据。实际 SDK 与代码容器的联合证据仍沿 10 月 31 日记录，本轮未重跑镜像。

## canonical 文件与可信来源

| 文件（位于 `contracts/`） | 职责 |
| --- | --- |
| `development-attempt-plan.schema.json` | 可信宿主确认的应用 / 模型 / 基准 / 工具路径 / 固定验证档案与预算计划 |
| `development-attempt-delegation.schema.json` | 单尝试委托、作用域 / 修订 / 摘要、有效期和上限；不含 bearer |
| `development-attempt-reservation.schema.json` | Go 生成的请求身份、连续序号、完整请求摘要与预留 |
| `development-attempt-receipt.schema.json` | 可信后端 / 执行器提供的实报、缺失、未知或未派发事实 |
| `development-attempt-event.schema.json` | Go 有序脱敏事件及账目快照 |
| `development-tool-action.schema.json` | 文件读取、替换及固定验证动作；不接受 shell |
| `development-tool-call.schema.json` | 尝试、父模型请求、工具调用、Go 请求身份与参数 |
| `development-sdk-worker.schema.json` | 私有 stdin 输入 / 确认、stdout 连续事件 / 工具提议；不是公开 API |
| `development-code-sandbox.schema.json` | 固定 Go 沙箱的私有工具观察值与归属资源账册；不是正式交付包 |

对象规范字段全部必填，未使用字段按动作明确为空字符串或零，不采用隐式默认值。路径 / 替换正文分别最多 256 / 32768 UTF-8 字节；Go 补充字节限制和跨字段关系。计划可写路径必须属于可读路径，验证档案只接受 `go_test_v1`；这只是授权描述，不证明实际文件或命令已受隔离。

`PayloadDigest` 对已校验的固定结构紧凑 JSON 计算 SHA-256，不信任 SDK 提供的摘要。工具请求摘要包含完整调用、父请求与工具标识；替换绑定旧文件摘要，验证绑定候选摘要。候选变化使旧测试 / 审查证据失效。

## 真实资格、确认与凭据

首个内部开发档案复用现有本地身份负责模块，要求当前活跃账户、工作区成员、`applications:read`、`local_identity_roles:assign` 及实际管理员管理资格。仅有成员或应用运行权限不足以签发；管理员身份也不自动发起任务，还须确认精确计划及当前选择。未新增全局角色或权限目录，团队角色资格后续另定。

`Server.PrepareDevelopmentAttempt(request, plan)` 与 `IssueDevelopmentAttempt(request, plan, confirmationID, prepared)` 只供可信 Go 宿主调用。请求必须已经过现有本地 Session 中间件验证并选定工作区；账户 / 租户 / 工作区从验证上下文取得。相同确认身份只能创建一次尝试，失败或响应丢失不恢复原文凭据，也不自动重建执行。

资格读取复用以下真实负责模块：

- 身份 repository 的 `AuthorizeWorkspace`，冻结账户 / 成员 / 角色修订及授权摘要；不保存登录名、密码或 Session 秘密。
- `ApplicationInteractionAuthoritySnapshot.AuthorityDigest`，精确读取活跃应用与 Prompt / Agent / Workflow 运行权威，不另造应用真相源。
- Gateway 目录中的稳定 `profile:` / `provider:...:profile:` 身份，精确匹配启用的模型档案；首轮仅接受 `openai-compatible`、已配置凭据的档案，不猜测或回退其它模型。
- 现有开发态 `GatewayModelPricingRepository`，冻结提供方 / 档案 / 模型作用域内的有效美元定价。缺失、不可用或全零价格均拒绝，不以零价代替未知。

`Scope` 冻结账户、工作区、应用权威、开发绑定、授权修订、运行 / 任务 / 尝试、计划 / 基准、模型 / 定价 / 工具策略摘要及 `executor_instance_id`。当前单尝试的运行 / 任务身份与尝试一致，修订为 1；成员变化通过真实摘要核对，不把常量修订当作身份负责模块的修订。

Go 生成 256 位随机不透明 `devtask_` bearer，仅可信 SDK 工作进程持有；数据库只保存 SHA-256，`Issued` 默认 JSON 序列化排除原文。凭据不进入描述、事件、URL、日志或代码容器，不能调用公开模型 / 管理 API。audience 固定为 `development-attempt-control.v1`，有效期最多十分钟，不自动续签；当前进程另用单调时钟限制期限。

模型预留、工具入队 / 派发及模型返回后，均读取最新资格并比较冻结快照。模型等待时每秒发起独立资格核对，失效则先提交停止状态、吊销凭据并中断调用上下文；读取失败按结果未知处理。已另以真实 SDK 进程验证模型等待中断与进程组退出；该 SDK 证据仍不证明真实上游计算停止，代码资源须另按沙箱归属核对。

## 受限 SDK 工作进程与私有事件

可信宿主显式调用 `Runtime.RunWorker`，按当前作用域 / 凭据、精确首次消息和私有空目录核对后，先持久提交 `worker_claimed` 再启动一次进程。已有认领、旧实例或旧凭据不能重放。当前仅支持 Linux / macOS 进程组，未开放 Windows 执行、公开启动路由或自动启用配置。详细运行边界见[SDK 模块说明](../../services/runtime/development_sdk/README.md)。

stdin 一次交接任务凭据与计划输入，后续只交回已提交的事件确认和可信工具观察值；stdout 帧只允许既定类型、最多 128 项和每帧 262144 字节。Go 校验 SDK 序号、类型、请求身份和工具摘要后才写事件 / 回复确认。SDK 普通日志不公开，错误只接受固定诊断码；未知帧或失败报告立即停止，不解析原始错误文本猜授权。

SDK 原基线 `1.59.0` 和已独立固定响应复验的精确变体 `1.59.0+radishmind.1` 使用真实 Agent / LocalConversation 与显式模型传输适配，只注册 `development_tool`，禁用默认工具、MCP、插件、动态上下文、压缩、critic、fallback 及其它模型入口。会话仅内存保存；首版限一条系统消息与一条用户任务，均纳入 Go 计划确认。工具仍只可先入队，由可信 Go 回调形成观察值；默认未配置实际工具；可信宿主可另行显式绑定固定 Go 探针的 `WorkerLaunch.Sandbox`，细节见[沙箱合同](development-code-sandbox-v1.md)。

每个生命周期和允许的 SDK 事件以 Go 观察事件追加，包括认领 / 启动 / 就绪、系统提示 / 用户 / 助手消息观察、动作 / 观察值 / 错误 / 安全报告观察、工具等待、SDK 完成报告、进程退出 / 失败 / 清理未确认。它们不复制 SDK 正文，也不直接赋予执行资格或把 SDK 自报完成当作终态。

SDK 模式的最终模型响应仅记录 `model_final_response` 并吊销新调用资格；完成报告、实际零码退出、空进程组和最新资格核对通过后，Go 才原子记录 `worker_exited / attempt_completed`。非 SDK 模型入口保持原完成语义。缺失用量保留全部预留，工具失败的未知效果不制造未知模型量。

SDK 观察事件为停止 / 清理保留事件容量，持久事务总事件上限为 200，超限拒绝并保持原记录可读。同毫秒的纯事件新增也必须提交，不因账目与时钟未变而丢弃。

## 工具提议、上下文与事务

Go 在返回真实后端模型响应之前，登记提议身份、父模型请求、工具调用标识、完整调用和参数摘要；不保存参数正文。SDK 回显时重新读取持久提议，核对参数、身份和路径策略，在同一事务内消费一次、创建工具预留并写事件。重复消费、跨尝试、父请求或参数替换均拒绝。

`ParseSDKToolArguments` 只识别已知可选 `summary`，限制 512 UTF-8 字节并拒绝控制字符，显式剥离后再规范化。它是未受信任的易失展示材料，不进入授权、事件、产物或公开 UI；其它附加字段仍拒绝。

可信宿主把首次 system / user 规范消息摘要写入 `initial_messages_digest` 并纳入计划确认；首次模型请求必须精确匹配该摘要，且末条为 user，防止 SDK 替换获准任务。后续请求必须完整回显前次规范上下文和模型响应，并附一条可信工具观察值；数据库仅保存前缀 / 观察值摘要。`DispatchTool / CompleteTool` 是可信 Go 执行器钩子，SDK 没有完成回执入口。虚构工具结果、改变历史、重放或静默压缩均拒绝；上下文窗口不足时拒绝，不能自动截断。

持久化复用共享 SQLite，新增组件 `development_attempts / 0001_development_attempts / development_attempt_store_v1`：

| 表 | 内容 |
| --- | --- |
| `development_attempts` | 作用域、确认唯一性、运行实例、版本、凭据摘要 / 吊销、私有版本化状态、绑定与上下文摘要 |
| `development_requests` | 不可变请求身份、类型及摘要 |
| `development_tool_offers` | 提议身份 / 摘要、单次消费状态及观察值摘要 |
| `development_events` | 从 1 连续递增的不可变事件与摘要 |

使用 `BEGIN IMMEDIATE`、预期版本 CAS 和同事务写入；失败丢弃私有状态副本，不能泄漏内存修改或部分消费。请求与事件不可更新 / 删除，提议身份不可修改或逆转消费。读取核对状态、账目、事件、请求索引和提议关系；不修补损坏记录、不从创建函数重建。SQLite 写等待不占满连接池，保留当前资格负责模块的读取连接。

新组件仅支持本机 SQLite；不增 PostgreSQL / memory repository fallback，不据已有其它模块的双库证据声明本功能双库支持。工作台旧目录需停服、备份并显式 `upgrade`，`serve` 遇待迁移继续拒绝；不得覆盖日常库或沿用旧构建打开新增标记。

## 预算、回执与事件

首个档案单尝试仅一项在途，最多 8 次模型和 32 次工具预留，实际取计划较小值；接受即消耗次数，未派发退款不退次数。预留原子核对 `spent + held + new_bound <= limit`、次数、连续序号和请求唯一性，提交后才派发。DB 失败不调用后端；`dispatch_recorded` 表示外部影响可能发生，回执丢失不得重放。

模型输入上界须由可信 `NativeBackend.Policy` 提供可验证的完整输入界限，输出由请求和冻结档案共同限制；当前固定响应后端已验，真实提供方实现仍需证明该界限。费用使用整数 microUSD；预留向上取整，实报结算复用现有定价的半入规则。工具的模型 token / 费用预留为零，执行资源另由沙箱限制。

| 回执 | 处理 |
| --- | --- |
| `reported` | 原始上游确实给出用量，按冻结价格核算，释放预留并记全部实报；超额不裁剪，停止后续执行 |
| `missing` | 响应无 usage，保留全部预留及未知标记；不采纳 SDK 补出的零值 |
| `unknown` | 派发后超时、断线或丢失回执，保留预留并记录结果未知 |
| `not_dispatched` | 仅无派发记录的操作可释放预留，不退次数 |
| `not_applicable` | 可信工具完成，不产生模型用量 |

迟到可信回执只结算一次，终态不可复活；未知工具效果不制造未知模型用量。团队共享预算、真实计费与多运行额度不在本实现范围。

事件包含原有控制类型，并增加 `tool_offer_registered / tool_offer_queued / attempt_interrupted / attempt_failed`，后两者分别对应 `runtime_interrupted / native_response_invalid`。仅保存身份、枚举原因和账目，不保存文件 / 模型 / 命令正文、隐藏推理、bearer、summary 或原始异常。时间回拨不倒排事件，排序以序号为准。

`active` 可预留 / 派发，`cancel_requested` 拒绝新动作；`cancelled / completed / failed / outcome_unknown` 均不可恢复。`completed` 只表示控制工作结束，不证明测试通过、用户接受或 Git 合入。独立 `termination_status` 保留归属清理的 `not_requested / pending / unconfirmed / confirmed`，清理成功不抹掉失败 / 未知结果。

## 取消、关闭与重启

`Runtime.Cancel` 以预期版本提交取消意图、吊销凭据并取消本次模型上下文；没有 SDK 自报取消成功入口。`CloseDevelopmentControl` 先停止运行 / 中断模型并记录未决项，再关闭内部 HTTP；`Server.Close` 在释放 bridge 与共享数据库前调用它。已增加 SDK 工作进程的取消 / 关闭跟踪，等待真实 OS 退出并核对所属进程组；固定 Go 代码容器已另行取得真实终止与归属证据；绑定 `WorkerLaunch.Sandbox` 后完成前必须关闭沙箱，失败 / 取消同时核对 SDK、本地模型连接及代码资源。真实 SDK 与固定 Go 任务的联合技术验收已完成，通用代码沙箱仍未验收。

SDK 进程由独立控制路径立即发出软停止，两秒未退出则强制终止，有界等待仍不能核对退出或空组时记录未确认；额外 FD 3 失联会终止 SDK 所属组。真实 SDK 等待取消和存活管道断开已验，Go 测试另验证拒绝软停止的子进程组。固定 Go 沙箱已验证取消后强制删除、空资源查询及显式遗留清理；五秒内强制终止、十秒内归属核对无法满足时如实记未确认。分层实际证据与联合任务分别记录，不能仅据 SDK 退出宣布完整本地取消成功。仅凭 SDK pause、浏览器断线、PID 不存在或自报布尔值不足以标记已取消；本地取消也不证明上游停止或零费用。

每次运行生成新的实例身份；重启只读回元数据、账目与事件，旧 bearer 不再可用，不自动认领、续签、补跑模型或命令。可信宿主显式 `Interrupt` 才将遗留活动尝试收口为未知并吊销，后续新尝试需新确认。当前没有自动恢复或持久租约调度器。

## 验证与下一实施

Go 验证覆盖隔离 SQLite 的创建 / 关闭重开、事件 / 请求 / 提议不可变、并发消费和取消、CAS、事务故障注入、坏状态拒绝、固定原生工具往返、可信观察值 / 上下文、防止缺失用量变零、等待时撤销与监听关闭。平台测试使用真实 SQLite 身份 / 应用 / 定价 repository、现有 Workflow 权威及本地 Session 中间件；资格当前性不是由测试请求自报。

受限 SDK 工作进程与代表任务的完整 Agent 事件链已固定响应验证。固定 Go 沙箱的文件 / 固定测试执行、所属资源终止及硬期限已独立真实验收。开发产物合同、内部 SQLite 链及真实 SDK / 固定 Go 代码容器的联合任务已验，见[联合记录](../devlogs/2026-W44.parts/development-joint-20261031.md)。原生 HTTP 已于 11 月 5 日取得固定上游技术证据，见 [W45 周志](../devlogs/2026-W45.md)；正式采用仍需收口许可、阶段 1 合同与视觉依据，真实厂商资格另验。固定响应证据不代替通用代码沙箱、真实模型质量或真实用户试用。阶段 0 尚未退出，真实模型与用户试用继续暂停。

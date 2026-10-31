# 受控开发尝试：工具通道、委托、事件与取消 / 预算合同 v1

更新时间：2026-10-31

状态：首个单 Agent 开发尝试的合同与离线领域规则已物化；不是已注册 HTTP API、凭据签发服务、持久执行器或沙箱。所属功能见[主专题](../features/workflow/controlled-software-development-team-v1.md)，实施记录沿用[集中任务卡](../task-cards/development-task-probe-v1-plan.md)。

## 已确认的接入选择

采用**独立内部监听器**，不扩展公开 `/v1/chat/completions`、旧建议型 Agent Profile 或 30 秒调用边界。首版本机内部开发者模式仅监听回环地址、运行时分配端口，由 Go 可信执行控制层告知 SDK 工作进程；不写死已有验收端口，不向代码容器开放网络或凭据。本轮不创建监听器或启动服务。

| 拟实施的内部入口 | 请求与职责 | 边界 |
| --- | --- | --- |
| `POST /internal/development/v1/attempts/{attempt_id}/chat/completions` | 受限 OpenAI Chat Completions 传输，Go 先鉴权 / 预留再请求模型 | 不转发到现有建议桥接，不允许选择任意提供方或 URL |
| `POST /internal/development/v1/attempts/{attempt_id}/tool-calls` | `development-tool-call.schema.json`，回显 Go 已登记的工具提议 | 工具只通过可信执行控制层进入隔离代码副本 |

这两条路径是下一实现的内部合同名称，本轮未注册路由。没有 SDK 可任意上报权威事件、签发委托或宣布取消成功的入口。正式 Web 的启动 / 取消 / 结果资源接口及权限接线另在阶段 1 实施，不把这些内部入口公开给浏览器。

模型传输只接受 `model / messages / max_completion_tokens / tools`，以及显式为 `false` 的 `stream / parallel_tool_calls`；未知字段拒绝，不静默忽略。`messages` 只允许文本及原生 assistant `tool_calls` / tool `tool_call_id`，不解析答案正文里的工具 JSON。模型必须精确匹配 Go 冻结引用解析出的传输模型名，输出上限不得超过模型档案与剩余预算。单次模型响应首版最多一个工具调用；多个调用整体拒绝，不偷偷取第一个。

`tools` 必须精确匹配冻结的 SDK 工具目录，只有 `development_tool`，动作仅 `read_file / replace_file / run_tests`。旧 SDK 探针继续使用 `development_probe`，新 `development_tool` 合同必须单独接线和复验。SDK 默认终端、MCP、递归委派、自动压缩 / critic / fallback 和隐式重试均关闭；不能证明所有模型调用经过此通道时不采用 SDK。已有 SDK LLM 子合同证据不代替完整 Agent 验收。

## canonical 文件与职责

| 文件（位于 `contracts/`） | 作用 |
| --- | --- |
| `development-attempt-delegation.schema.json` | 一次尝试的委托描述、资源 / 修订 / 摘要及预算上限；不含 bearer |
| `development-attempt-reservation.schema.json` | Go 计算的请求预算预留，绑定序号、请求身份与完整请求摘要 |
| `development-attempt-receipt.schema.json` | 可信 Gateway / 执行器给出的实际用量、缺失、未知或未派发事实 |
| `development-attempt-event.schema.json` | Go 控制状态与账目的有序脱敏事件，不含 SDK 正文 |
| `development-tool-action.schema.json` | 读取、替换及验证动作的明确参数，不接受 shell |
| `development-tool-call.schema.json` | 尝试、父模型请求、工具调用标识、Go 分配请求身份与参数 |

对象字段全部必填，未使用字段按各动作明确要求为空字符串或零；不是隐式默认值。未知、重复、空值、大小写别名、尾随 JSON 与过大输入均拒绝。JSON 总上限 262144 字节；路径 / 内容分别最多 256 / 32768 UTF-8 字节。JSON Schema 的字符限制之外，由 Go 负责字节限制、跨字段时间和运行时关系约束。

`PayloadDigest` 是 Go 对已校验、固定字段顺序的规范结构紧凑 JSON 计算的 SHA-256；不是 SDK 提供的任意 JSON 文本摘要。工具请求摘要包含完整 `ToolCall`，包括父请求和工具标识；替换另绑定当前文件摘要，验证绑定候选摘要。候选改动使旧测试与审查证据失效。

## 任务级委托

Go 在读取当前账户、工作区成员 / 授权修订、活跃应用绑定及模型 / 定价 / 工具策略后生成描述。`Scope` 绑定账户、工作区、应用、绑定及修订、授权修订、运行 / 任务 / 尝试、需求计划 / 基准摘要、模型引用、定价及工具策略摘要；另绑定 `executor_instance_id`。`audience` 固定为 `development-attempt-control.v1`，有效期最多十分钟，到期不自动续签。

仅有工作区成员或应用运行权限不足以签发：还须有显式开发执行资格和精确计划确认；资格目录与授权负责模块尚未实现时默认拒绝签发，不能借旧 Agent Profile 的批准状态开通工具。

运行时凭据采用 Go 生成的至少 256 位随机不透明 bearer，仅可信 SDK 执行进程持有；Go 保存单向摘要与描述关联，原文只交接一次，不进入 JSON 描述、事件、URL、公开日志或代码容器。凭据只对该内部 audience / 尝试有效，不能调用管理接口或公开模型 API；不复用浏览器 Session、用户长期 API key、提供方密钥，也不把模型消息当作凭据申请。

每次模型请求、工具派发和继续执行都读取最新负责模块快照，与描述精确比较。撤销、成员权限变化、绑定 / 模型 / 定价 / 工具 / 计划摘要变化或期限失效立即拒绝新动作，并由独立控制循环进入取消；不能只检查请求里自报的 `Scope`。首版控制循环核对间隔不超过一秒，不能等待被阻塞的 SDK 主动再发请求。

现有应用负责模块已有 `ApplicationInteractionAuthoritySnapshot` 及 Prompt / Agent 的运行权威快照，本合同以 `application_authority_digest` 绑定其 `AuthorityDigest`，不另造一份应用运行权威。 `model_ref` 直接复用 Gateway 目录的稳定身份（例如 `profile:probe` 或 `provider:huggingface:profile:dev`），由原负责模块核对存在性 / 资格，不新增模型别名表或 URL 入口。`binding_id / binding_revision` 描述开发运行与该应用的显式关联，不能冒充旧 Session 的 Profile Binding，也不继承旧 Agent 的代码工具资格。

本轮 `Authority` 是**由可信负责模块提供的当前快照**这一内部类型，不是认证实现。描述、`Authority`、可信工具提议与终止证据都不能从 SDK HTTP JSON 直接构造为“已认证”。签发、秘密保存、鉴权中间件及真实账户 / 应用绑定接线尚未实现。

## 工具身份与 SDK 元数据

Go 在把模型响应交给 SDK 前登记不可变提议，绑定本次模型请求、工具标识、规范参数和工具策略核对结果，并分配工具请求身份。SDK 工具回显该提议；接收端重新读取可信提议并逐字段比较，再绑定同一 `Reservation.payload_digest`，以 CAS 原子消费一次。跨尝试、父请求、工具标识、内容替换或重复消费均拒绝。

`ParseSDKToolArguments` 只显式识别已观察到的可选 `summary`，限制为 512 UTF-8 字节且无控制字符，并把它与规范参数分离。返回摘要仍是未受信任的易失展示材料，不写入权威事件 / 授权 / 产物，不直接作为公开 UI 文案。其他附加字段拒绝；旧探针 schema 对 `summary` 的拒绝保持。

`ToolAction` 的语法检查不授予路径访问权。Go 必须按冻结的工具策略核对可读 / 可写路径和角色所有权，实际文件类型、链接、摘要与挂载由隔离执行负责模块再验证。`run_tests` 只解析策略内的固定命令档案；请求没有命令、工作目录、环境或任意参数字段。模型提供的测试命令不得直接执行。

本轮已实现参数转换、工具提议 / 预留绑定校验；提议持久化、与真实父模型响应关联、路径策略负责模块及原子消费尚未接线。不能把传入同一个客户端对象两次当作可信提议比较成功。

## 预留、结算与缺失用量

`budget.unresolved` 是未决操作数；工具结果未知不制造未知模型用量，没有模型预留时 `usage_status` 保持 `not_applicable`。

首个单 Agent 档案只允许一项预留或派发在途。每个尝试最多 8 次模型预留、32 次工具预留，实际上限以委托较小值为准。接受预留即消耗次数，不因未派发释放金额而退还次数；这是保守的调用上限，不冒充已发生的网络次数。

模型预留由 Go 根据完整输入上界、最大输出和冻结定价计算 token / 费用上界；SDK 自报成本不参与。金额使用整数 `cost_microusd`，一单位为百万分之一美元；无可验证定价时不签发可执行委托，不能用零价替代未知。工具预留的模型 token / 费用字段必须为零，命令 / 输出 / 时间另受固定执行档案约束。

原子核对：`spent + held + new_bound <= limit`，同时核对次数、单在途、请求唯一性与连续序号，成功才提交预留事件。相同请求身份重发返回 `operation_duplicate`，不重新执行，也不从缓存制造一次新执行成功；合法新尝试须重新授权。

| 回执状态 | 可信来源与账目处理 |
| --- | --- |
| `reported` | Gateway 确认上游确实提供用量，以冻结价格核算；释放该预留，记入全部实际量，包括超过预留的部分 |
| `missing` | 模型响应已结束但上游无 usage；零字段只是无数量占位，保留全部预留，标为未知用量 |
| `unknown` | 已派发后超时 / 失联 / 回执丢失；保留预留，停止尝试并记录结果未知 |
| `not_dispatched` | 仅限没有派发记录的操作；可释放预留，不退还调用次数 |
| `not_applicable` | 工具已有可信完成回执；不产生模型用量 |

显式提供的真实零用量与缺失用量不同。不得把 SDK 补出的零值当作 `reported`；原始上游字段存在性须在 SDK 规范化之前由 Gateway 记录。HTTP 错误或断线不证明未发送 / 未计费，派发后不能套用 `not_dispatched`。

实际量超上界时不裁剪、不吞掉：记录全量，活动尝试进入 `failed / reservation_exceeded`，停止后续动作；已终止尝试只核账，不恢复执行。超额实报失败后仍可核对所属资源终止，未派发的新预留可释放，已派发的继续保留未知量；清理成功不把失败改写为取消成功。未知量的迟到可信回执仅结算一次。首版本合同为**单尝试预算**，团队共享预算、跨进程原子认领与持久账目仍须下一实施，不据内存锁宣称跨进程或全运行额度已生效。

## 事件与状态

Go 生成 `attempt_opened / reservation_created / dispatch_recorded / receipt_recorded / cancel_requested / termination_recorded / attempt_completed`。序号在同一尝试内从 1 连续递增；消费键包含工作区、尝试和事件身份，事件 ID 本身不是授权。时间由 Go 提供，回拨时展示时间不倒退，排序仍以序号为准；执行期限必须另用单调计时器，不能依赖展示时间。

事件只含稳定身份、枚举原因和预算快照，不含命令输出、文件正文、模型内容、隐藏推理、bearer、SDK summary 或原始异常。SDK 进度消息不能直接变成 Go 权威事件。读取按最后确认序号继续，重复去重、断档明确失败；事件读取不触发执行。

`active` 可预留 / 记录派发；`cancel_requested` 拒绝新动作；`cancelled / completed / failed / outcome_unknown` 均不可恢复派发。`completed` 只表示本次控制工作已结束，不证明测试通过、用户接受、Git 合入或生产就绪。迟到回执可以追加事件核账，但保持终止状态。 `termination_status` 独立记录 `not_requested / pending / unconfirmed / confirmed`；未确认后允许追加一次完整确认，结果未知或失败仍保持原结果。重复的未确认报告与已确认报告不重复生成事件。

状态、请求身份、预留、工具提议消费和事件必须同事务提交，之后才允许实际派发。DB 与外部效果不能原子回滚，`dispatch_recorded` 表示影响可能已经发生；提交失败不得继续，派发后回执缺失不得重放。本轮仅内存规则，尚无事务 / durable store，不能接真实执行器。

## 取消与重启

1. Go 提交取消意图、关闭新预留 / 派发入口；尚无派发记录的预留可释放，已派发的预算保持。
2. 同时中断本次模型连接和可信 SDK 执行进程，终止所属代码容器及进程树。软停止期限两秒，五秒内发出强制终止，十秒内仍无完整归属终止证据则记录 `outcome_unknown / termination_unconfirmed`。这些期限是待实现的运行合同，本轮没有实际 supervisor / timer 验收。
3. 只有可信负责模块核对本尝试 / executor 归属、进程已退出、代码容器已结束和本地模型连接已关闭，才可标记本地 `cancelled`。未创建的资源也需核对登记；SDK `pause()`、浏览器断线、PID 不存在或 SDK 自报布尔值均不足以构成证据。
4. 本地已取消不证明上游模型停止计算或零费用，也不回滚副本；保留未知预留及部分候选。清理失败 / 影响未知分别记录，不自动重跑命令。
5. 重启只读取记录与核对归属，不重新调用创建函数重置旧尝试、续签委托或恢复模型 / 工具；新尝试需要显式原因、新身份及当前资格。

## 已验证与下一实施

代码位于 `services/platform/internal/developmentcontrol/`；严格 JSON 解析提取到 `internal/strictjson/`，原探针继续使用同一规则。离线测试覆盖跨作用域 / 过期 / 撤销、预算并发、未知用量与晚到结算、取消与派发竞态、结果未知、超额实报、工具身份 / 摘要替换、SDK 元数据分离及严格输入。合成样例和 Go / Python 共用取消轨迹位于 `scripts/eval/fixtures/development-control/`，不含凭据或真实模型材料。

下一实施仍需：持久尝试 / 认领与事件事务、真实资源负责模块快照及任务凭据签发 / 撤销、专属 listener / 原生消息适配与工具提议原子消费、受限 executor / sandbox、独立取消控制循环与期限，以及开发产物保存合同。首次运行使用固定响应；真实模型和真实用户试用继续暂停，不沿用上一 SDK 容器窗口授权。

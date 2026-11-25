# 单 Agent 开发任务：公开交付合同 v1

更新时间：2026-11-10（确认回执、幂等及重启归属）

状态：阶段 0 的公开交付合同与固定 Go 单 Agent 首版范围已获项目所有者确认。公开 API、任务协调器和 Web 属于阶段 1 实现。

所属功能：[实施与评测](../features/workflow/controlled-software-development-team/implementation-and-evaluation-plan-v1.md)、[运行状态与证据](../features/workflow/controlled-software-development-team/team-runtime-state-and-artifacts-v1.md)、[协作界面](../features/workflow/controlled-software-development-team/collaboration-workspace-ui-v1.md)。本页维护新增公开任务合同；底层委托、工具、事件与产物字段继续由[尝试控制](development-attempt-control-v1.md)和[开发产物](development-artifact-v1.md)负责。

2026-11-10 确认回执读取权限、同请求幂等判断顺序和重启后的旧实例资源归属。首版技术交付后先做有限真实单 Agent 价值复核，再决定团队范围，见[实施与评测](../features/workflow/controlled-software-development-team/implementation-and-evaluation-plan-v1.md#single-agent-value-review)。真实模型资格另行验收。

## 交付范围与能力边界

- 入口为当前工作区内活动应用的“开发任务”；一个任务包含一份固定需求、可修订的执行计划和至多一次尝试。任务完成后可重新创建独立任务，不提供原地继续、自动返修或历史尝试重放。
- 唯一来源为 `probe_list_filter` 的冻结三文件 Go 树，保留来源和基准摘要。只允许修改 `list.go`；`go.mod`、`list_test.go` 及验证命令受保护。任务使用预置回归用例，不声称 Agent 已具备补写测试能力。
- 单模型、串行工具，复用 `read_file / replace_file / run_tests`；不接受任意仓库路径、URL、shell、依赖安装、自定义系统提示、团队定义或运行中聊天输入。
- 首轮技术交付使用明确标记的固定响应档案。真实提供方只有在工具资格、完整输入计量和费用窗口另行验收后才可选；固定响应档案不能冒用真实模型名称或作为默认回退。
- 候选、命令结果、资源终止、保存状态与用户接受分开。只交付审阅、显式保存和 JSON 导出，不增加接受、合并、推送或发布操作。

## 资源与负责模块

| 对象 | 负责模块与持久内容 | 关系 |
| --- | --- | --- |
| 开发任务 | Go 任务协调模块；作用域、固定需求引用、创建请求、当前计划修订、尝试引用 | 每个账户在应用内拥有自己的任务，不借用旧 Session / Workflow Run |
| 计划修订 | 同一任务模块；原尝试计划、执行档案、来源、镜像 / 执行器摘要及正文保留说明 | 修订不可变；启动后不再编辑 |
| 确认预览 | 任务模块；计划与当前资格摘要、生成 / 失效时间、一次消费状态 | 最多五分钟且不超过资格有效期；它是确认对象，不是执行凭据 |
| 尝试、预算与事件 | `developmentcontrol` | 沿用原状态机和单尝试账目，任务列表不能自行推算执行终态 |
| 沙箱和可信快照 | `developmentsandbox` | 私有候选、受保护测试、资源归属和有界清理 |
| 显式保存与导出 | `developmentartifact` | 保留现有包、摘要、配额、敏感规则和不可变存储 |

公开 `task_id`、`attempt_id` 由服务端生成。阶段 1 新委托的 `run_id = task_id`，`task_id` 指向任务，`attempt_id` 指向唯一尝试；与任务记录、资源账册、包内引用逐项一致。阶段 0 已存记录三者相同的历史事实不改写，也不自动迁移成可再次执行的公开任务。

任务身份包含 `tenant_ref / workspace_id / account_id / application_id`，从认证和应用负责模块取得；请求不能自报账户、所有者、执行器实例、委托或源文件正文。公共返回不能携带内部地址、任务 bearer、模型密钥、Docker socket、宿主路径或原始 SDK 上下文。

## 身份、权限与角色分配

所有路由复用正式 Web 的本地 Session、工作区选择、Origin 和写操作 CSRF 验证；拒绝开发身份头、长期 API key 和内部任务 bearer 替代用户会话。每次操作重读活跃账户、成员、精确权限、应用及任务所有者；跨用户、跨应用或跨工作区资源查询返回相同的不可访问结果。

| 操作 | 精确权限（均另需 `applications:read`） |
| --- | --- |
| 能力、自己的任务列表 / 详情 / 尝试 / 事件 | `development_tasks:read` |
| 创建任务、修订未启动计划、准备确认 | `development_tasks:read` + `development_tasks:write` |
| 确认并发起唯一尝试 | 上述两项 + `development_tasks:execute` |
| 请求取消自己的尝试 | `development_tasks:read` + `development_tasks:cancel` |
| 预览与读取自己的开发包 | `development_artifacts:read`；预览另需任务读取 |
| 保存、导出 | 产物读取分别加既有 `development_artifacts:save / export` |
| 查询自己的请求回执 | 任务操作回执需 `development_tasks:read`，保存回执需 `development_artifacts:read` |

拟在阶段 1 的版本化角色目录增加叠加角色 `workspace_developer`：仅授予本表四项任务权限、三项产物权限及 `applications:read`，不拥有成员管理或其它应用执行权限。应用创建仍由既有 Builder 等角色负责。现有 Reader / Builder / Reviewer / Admin 的已存授权不得自动获得代码执行权限；管理员使用现有角色分配流程显式授予新角色，预览新增权限与目录版本。

角色目录升级必须保留旧版本分配的明确识别和冻结权限，不按新目录重算旧授权；版本未知或摘要不符拒绝。阶段 1 在原身份模块实现目录升级与管理界面的精确分配，不能用阶段 0 的 `local_identity_roles:assign` 管理资格直接包装正式执行接口，也不另建身份系统。

取消不要求原模型配置仍可用，也不要求保留执行权限；它要求当前取消权限和原任务所有权。执行资格撤销、成员失效、应用归档或会话失效时，协调器主动停止新动作并请求终止，不依赖已失权用户再次点击取消。历史读取 / 导出不要求原模型在线；归档应用禁止新建、启动和新快照保存，自己的已保存包仍按现有合同读取。

回执查询在完整当前作用域内查找原请求，按回执所指资源重查所有权和当前读取权限。回执不存在、超出作用域或无该回执读取资格时统一返回不可访问的 404。重新提交原写请求仍要求对应操作的当前写权限。

## 公开 HTTP 面

基路径记为 `A = /v1/user-workspace/applications/{application_id}`，工作区选择沿既有 `workspace_id` 参数；每个对象仍重查完整作用域。以下为待实现路由，当前不存在。

| 方法与相对路径 | 输入及返回 | 行为 |
| --- | --- | --- |
| `GET A/development-capabilities` | 固定来源档案、精确可选模型引用、上限、各操作资格与稳定不可用原因 | 只读；关闭状态仍可说明，不启动执行器 |
| `POST A/development-tasks` | `request_id / source_profile_id / model_ref / limits` → 任务及计划修订 1 | 创建元数据，不执行；固定需求来自服务端版本 |
| `GET A/development-tasks` | `cursor / limit` → 自己的摘要行、`next_cursor / has_more` | 默认 20、最多 50，无全量统计；创建时间和身份稳定排序 |
| `GET A/development-tasks/{task_id}` | 任务、当前计划、尝试引用、操作资格与保存引用 | 不隐式准备新授权，不返回代码正文 |
| `POST A/development-tasks/{task_id}/plan-revisions` | `request_id / expected_version / model_ref / limits` → 新计划修订 | 仅未启动任务；使所有旧确认失效 |
| `POST A/development-tasks/{task_id}/preparations` | `request_id / expected_version / plan_digest` → 确认预览 | 重读当前应用、身份、档案、价格和执行环境；不发凭据 |
| `POST A/development-tasks/{task_id}/attempts` | `request_id / expected_version / preparation_id / confirmation_digest` → `202` 任务 / 尝试引用 | 明确确认并启动；提交后异步执行，一任务最多一次 |
| `GET A/development-tasks/{task_id}/attempts/{attempt_id}` | 原控制状态、版本、终止状态、预算 / 用量及资源核对结果 | 刷新只读取，不恢复执行 |
| `GET A/development-tasks/{task_id}/attempts/{attempt_id}/events` | `after_sequence / limit` → 有序允许事件、`next_sequence / has_more / record_version` | 复用最多 50 项 / 总计 200 项的既有事件合同 |
| `POST A/development-tasks/{task_id}/attempts/{attempt_id}/cancel` | `request_id / expected_version` → `202` 取消意图与当前状态 | 不把受理写成已停止；终态重复操作返回原状态 |
| `GET A/development-tasks/{task_id}/attempts/{attempt_id}/artifact-preview` | 可信包、实际摘要、证据状态、原到期时间 | 仅资源结束确认后；浏览器不能上传包 |
| `POST A/development-tasks/{task_id}/attempts/{attempt_id}/artifacts` | `request_id / expected_digest` → 已保存记录 | 调用既有保存服务并重新核对可信来源 |
| `GET A/development-artifacts` | `cursor / limit` → 自己的已保存包摘要 | 精确应用内列表，分页规则同任务；不合并旧结果正文 |
| `GET A/development-artifacts/{artifact_id}` | 已保存记录及包 | 精确读取、重算校验，不依赖临时目录 |
| `GET A/development-artifacts/{artifact_id}/export` | `request_id` → 单份 JSON 导出 | 既有独立导出权限；无公开链接、ZIP 或服务端导出历史 |
| `GET A/development-requests/{request_id}` | 操作类型、请求摘要、任务 / 尝试 / 保存记录引用 | 按当前作用域、所有权及回执资源的读取权限只读查询 |

列表游标是服务端编码的作用域、筛选和最后排序键，输入必须有界、严格校验；游标不能授予权限。首版只提供按创建顺序的有界历史，不增加跨账户搜索或多存储库汇总。事件采用短轮询，同一页面只允许一个在途请求；可见且活动时两秒一次，隐藏或失权即停，断线只重读同一游标。无 SSE、WebSocket 或浏览器直连内部监听器。

## JSON、摘要与错误

阶段 1 首项代码工作将上述字段物化为一个 `contracts/development-task-delivery.schema.json`，Go 与 TypeScript 消费同一版本，不各自扩字段。本页是设计源，不把尚未物化的 schema 标作已验证。已有 attempt / artifact 子对象直接引用其 canonical schema，不复制新定义。

| 字段组 | 设计约束 |
| --- | --- |
| 请求身份 | `request_id` 使用现有小写身份格式，3–64 字符；唯一键为完整用户 / 应用作用域与请求身份，跨操作复用同一身份也冲突；回执记录操作类别 |
| 任务元数据 | `task_id / source_profile_id / record_version / created_at_ms / updated_at_ms / plan_revision / attempt_id`；尚无尝试时 `attempt_id` 明确为空字符串 |
| 计划 | 嵌入既有 `development_attempt_plan.v1`；另记录固定来源版本、SDK / Go 镜像与执行器摘要、固定响应模式和保留规则版本 |
| 确认 | `preparation_id / task_id / task_version / plan_revision / confirmation_digest / prepared_at_ms / expires_at_ms`，并展示其绑定的完整计划及安全资格摘要 |
| 尝试投影 | 原 `phase / termination_status / record_version / budget`、明确 `usage_status`；资源核对 `not_requested / pending / unconfirmed / confirmed` 不合并进“成功” |
| 受控执行投影 | `resource_status` 按模型连接 / SDK / 代码环境给出有界枚举；`active_tool` 绑定工具序号、名称、候选摘要与起始时间，空闲时明确无活动项。由可信协调器记录，不从 SDK 文本猜测；枚举与空值形式在物化 schema 时统一 |
| 操作资格 | 每项操作分别为 `allowed` 与允许列表 `reason`；按钮和服务端使用相同原因语义，不能以缺字段当允许 |
| 时间与金额 | UTC 毫秒安全整数；金额为整数 microUSD；未知用量保留明确状态，前端不能把账目占位零显示为实际零消耗 |

拒绝未知 / 缺失 / 重复 / 大小写别名字段、非法 UTF-8、尾随 JSON、null 和不支持的枚举。元数据写请求最多 16 KiB；任务详情最多 64 KiB；事件页最多 256 KiB；开发包仍最多 1 MiB，读取 / 导出信封最多 2 MiB，全部以 UTF-8 字节计。

确认摘要由服务端对固定结构的紧凑 JSON 计算 SHA-256，包含完整作用域、任务版本、计划、资格选择、来源、SDK / Go 镜像、执行器、保留规则和失效时间；前端只回显。任一字段变化使确认失效。修订、确认与启动均使用预期版本 CAS；事件更新不直接消耗任务计划版本，取消使用尝试版本。

响应沿平台安全错误信封，稳定码使用 `development_task_` 前缀：`invalid / denied / not_found / conflict / preparation_expired / unavailable / capacity_exceeded / corrupt`。未登录为 401，缺权限为 403，不可访问资源为 404，CAS / 已启动 / 同请求异内容为 409，过期确认或来源为 410，体积超限为 413，容量不足为 429，存储或执行环境不可用为 503。产物沿既有错误码；不能把数据库异常、宿主路径或提供方正文透传出去。

## 计划上限与启动事务

首版上限取现有固定探针与控制合同的交集：模型调用 4 次、工具 3 次、单命令 120 秒、总期限 600 秒；token / 费用必须明确为正且不超过原合同上限。用户可降低预算，不得提高固定档案上限。预置任务消息与系统消息绑定精确摘要；配置成本不等于实际发生费用。

同一工作台的持久执行槽最多容纳一个未确认结束的开发尝试。SQLite 原子占用该槽，进程重启后仍沿用持久记录。忙时拒绝新启动。

创建、修订、确认、启动、取消和保存都持久记录请求结果。校验请求后，先重查当前 Session、完整作用域、资源所有权及原操作权限，再按 `request_id` 比对原操作与请求摘要。操作和摘要均相同则返回原回执，任一不同则返回冲突。只有首次写入才检查 CAS、确认有效期、执行槽、临时来源等适用条件。事务内以作用域内唯一请求身份收敛并发重提，重提返回原回执时不产生新的写入副作用。

启动按以下顺序实现：

1. 按上述规则检查当前权限并处理原回执。对首次启动，重查当前应用、模型、价格和执行档案，比较确认摘要、有效期与任务版本。
2. 在共享 SQLite 同一事务中消费确认、占用执行槽、建立任务到尝试引用、保存启动请求回执和原尝试记录。创建失败全部回滚；现有 `Issue` 需支持这一可信宿主事务，不能先签发再补写任务链接。
3. 提交后在当前进程向受控工作进程交接一次性凭据；认领记录先于进程启动。凭据只在内存和私有管道传递；响应只包含公开资源引用。
4. SDK 和代码环境仍由原执行负责模块运行，资格重验、模型预留、工具消费及终止使用原合同。任何准备或启动错误写明确失败 / 未知结果，保留已提交回执，不自动新建尝试。

事务提交与 OS 启动无法原子化。提交后未能确认启动、凭据丢失或平台崩溃时，记录可读取的 `outcome_unknown` 与资源核对状态；不能重新生成凭据启动第二次。只有归属资源结束已确认后释放执行槽，未知归属阻止下一次执行，但不阻止历史读取。

启动或保存回执丢失时，客户端先查询原请求。收到不可访问的 404 时保留结果未确认。需要重提时由用户按当前资格提交同一请求身份和原内容。客户端刷新只发送读请求。

## 生命周期与恢复

| 情况 | 权威处理 | 页面表现与允许后续动作 |
| --- | --- | --- |
| 未启动 | 任务计划可修订；旧确认随修订失效 | 审阅当前范围、重新确认；不显示执行进度 |
| 活动 | 原 `active` 状态、事件和预算驱动 | 查看时间线、请求取消；离开页面不自动取消 |
| 请求取消 | 先提交取消意图、停止新派发、吊销内部凭据，再停止模型连接 / SDK / 容器 | 显示取消中及未确认资源，允许只读刷新 |
| 取消未确认 | 保留 `outcome_unknown` 或原失败状态及 `termination_status=unconfirmed` | 不显示安全结束，不允许快照保存或新执行 |
| 正常结束 | 等待实际 SDK 退出、代码资源终止与终态提交 | 审阅候选；命令通过不等于需求验收通过 |
| 账户 / 成员 / 执行权限失效 | 独立资格监视停止新动作并请求终止，不能依赖后续工具调用才发现 | 清除失权正文；恢复资格后只能读取已有记录 |
| 进程关闭 | 停止接收启动、取消工作、按原有时限等待、持久记录未决项，最后关闭监听器和数据库 | 刷新显示原结果或未知，不提前宣称取消 |
| 服务重启 | 新进程使用新实例身份，按同一工作台账册记录的旧实例身份核对并终止可证明归属的旧资源 | 只恢复记录。旧活动标未知，清理确认单独追加 |
| 临时内容失效 | 24 小时期保持；无可验证快照则返回不可用，不从工作目录猜测重建 | 已保存包仍可读，未保存候选显示无法恢复 |
| 保存失败 / 回执丢失 | 保持原尝试与期限，先查原保存请求，获准重提时沿用原请求身份和包摘要 | 区分未保存、保存失败、结果未确认及已保存。已保存时打开原交付包 |

重启时，当前进程身份用于当前执行，旧实例身份只用于按本工作台持久账册核对旧资源。仅对账册已记录且旧实例、任务、适用标签和镜像身份均匹配的资源执行终止与结束核对。其他工作台的资源、缺少归属信息或不匹配的资源禁止终止或删除。旧尝试只恢复记录，禁止认领、续签或重放。归属或结束仍未确认时保留执行槽，包括运行环境不可用的情况。现有探针提供显式遗留清理，自动核对属于阶段 1。

## 保留、列表与结果库

新增任务 / 计划 / 请求回执采用共享 SQLite 的独立组件，旧库先备份并显式升级。首版每个工作区内每所有者最多 256 个任务，单任务计划最多 16 修订 / 请求回执最多 64 项；到限明确拒绝，不自动删除历史或重用请求身份。元数据保留至工作台数据由用户整体清理，首版不提供逐项删除或后台过期清理。

计划仅含固定模板及配置引用，不持久化任意用户需求、草稿消息、代码或模型上下文。事件沿既有脱敏元数据。临时候选 24 小时与已保存包 64 份 / 64 MiB、单包 1 MiB 的规则保持。用户在保存前可查看实际代码、diff、日志脱敏和证据状态；失败 / 取消后的部分成果只有内容完整且资源结束确认才可保存。

开发任务页包含自己的历史入口与已保存交付；应用结果库以明确“开发交付”类型链接精确开发包详情，不把开发包塞进旧 Session 文本字段，不展示不存在的跨类型总数。跨资产列表需要的开发摘要只由上述应用内列表提供，加载失败独立显示，不返回伪空目录。

## 实现拆分与验收

阶段 1 在同一个功能工作包内依次落实：公开 canonical schema 与权限目录 → SQLite 任务 / 请求事务和运行协调 → 正式 HTTP 严格消费 → Web 确认、运行与交付审阅 → 实际隔离执行与浏览器验收。普通范围内步骤连续推进，不再按每个路由创建任务卡或检查器。

验收必须涵盖：固定任务实际修改与受保护测试；重复确认 / 并发启动只执行一次；授权撤销与跨作用域拒绝；HTTP 回执丢失和保存重试；SDK / 容器启动失败与取消未确认；关闭 / 崩溃 / 重启不重放；事件游标重复 / 断档；候选变化使旧测试过期；包损坏 / 敏感正文拒绝；中英、键盘、窄屏和历史精确重读。

回执专项覆盖已提交启动的确认已过期、已提交保存的临时来源已过期，以及撤权后的查询和重提。重启专项覆盖本工作台的旧资源、其他工作台资源和缺失归属标识的资源，并检查未确认时执行槽仍占用。

阶段 0 退出条件齐备后进入上述实现。下载、镜像、服务、打包及模型运行按具体窗口授权。

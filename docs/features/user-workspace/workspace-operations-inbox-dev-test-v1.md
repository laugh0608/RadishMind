# 工作区运营收件箱（开发 / 测试态）v1

更新时间：2026-07-26

状态：`workspace_operations_inbox_dev_test_v1_batch_a_complete`

## 功能目标

工作区成员在切换 active workspace 后，应能先看到一个可解释、可跳转的运营关注队列，而不是分别浏览 Applications、API Keys、Workflow Definitions 与 Runs 才发现异常。

v1 只消费 Workspace-scoped Read Transition 已授权并脱敏的首分页读快照，在 Web 内构造确定性关注项。它不是新的业务真相源，不保存决定，不执行修复，也不把当前窗口解释为工作区全量统计。

## 目标用户与核心流程

目标用户是开发 / 测试环境中的工作区成员。

1. 用户通过既有 active workspace selector 选择工作区。
2. 五条 workspace read route 继续独立完成身份、成员资格和资源权限判断。
3. 收件箱只消费其中 Applications、API Keys、Workflow Definitions、Runs 四类成功的脱敏快照。
4. 投影按固定规则生成关注项、来源覆盖状态和稳定顺序。
5. 用户打开关注项时，只切换现有 application / workflow definition / run 选择并跳转到既有详情；API key 项只跳转到既有密钥区。
6. workspace 切换沿用既有失效语义，旧快照和旧选择不得留在新工作区收件箱。

Quota 不参与 v1。缺少可信 policy owner 时仍由既有 route 返回 `quota_policy_unavailable`，收件箱不得推算 quota、token、cost 或 billing。

## 数据边界

| 来源 | 允许消费 | 关注项规则 | 禁止推导 |
| --- | --- | --- | --- |
| Applications | application ref、显示名、类型、生命周期、最近运行状态摘要、更新时间 | `archived` 只作为低优先级历史提醒；旧摘要中的 `failed / blocked` 可作为应用级提醒 | 不推导 membership、最新 definition、key 归属或发布资格 |
| API Keys | key ID、状态、scope、创建 / 过期 / 最近使用时间 | `rotation_required`、`expired`；显式过期时间进入未来 14 天窗口 | 不读取 credential、digest、header，不把长期未使用解释为异常 |
| Workflow Definitions | definition ID、application ref、version、状态、风险、更新时间 | 非 `active / published` definition 进入审查队列 | 不推导可发布、可运行或已批准 |
| Runs | run ID、application ref、definition ID、状态、failure code、trace、时间 | `failure_code != none` 或 `failed / blocked / outcome_unknown` | 不读取输入输出，不推断根因，不执行 replay / resume |

所有规则只基于当前返回窗口。任一 collection 存在 `next_cursor` 时，对应 coverage 固定为 `partial_window`；没有关注项只能解释为“当前窗口未发现”，不能解释为“工作区健康”。

## 确定性投影

- projector 必须是纯函数，reference time 由调用侧显式传入，测试不得依赖系统时钟。
- item ID 由 `source + resource ref + reason` 构造，不使用随机数。
- 严重度固定为 `critical > high > medium > info`。
- 同严重度按资源时间倒序，再按 item ID 字典序排列。
- 同一资源的不同 reason 可分别保留；v1 不跨 owner 猜测关联，也不合并为新的 incident。
- 四类 coverage 分别为 `complete_window | partial_window | unavailable`。
- 全部来源 unavailable 时整体状态为 `blocked`；部分 unavailable 或 partial window 时为 `partial`；其余为 `ready`。
- denied / unavailable collection 不消费 partial rows，只生成来源不可用说明；公开说明只使用既有稳定 failure code。

## 导航与状态语义

- Application 项：选中 application，清空不属于该 application 的旧 definition / run / draft / scenario，再进入 Applications。
- Workflow Definition 项：复用既有 definition selection patch，再进入 Workflows。
- Run 项：复用既有 run selection patch，再进入 Run History。
- API Key 项：进入 API Keys，不创建第二套 key detail 或 mutation。
- 收件箱本身不写 URL 参数、cookie、`localStorage` 或 `sessionStorage`；页面 section anchor 只用于现有页面内导航。
- 切换 workspace 时，旧 read load 被 request lifecycle 淘汰；loading / failed 期间不得继续显示旧工作区关注项。

## 隐私与授权边界

- 收件箱不发独立后端请求，因此不能绕过 workspace route authorization。
- 输入只能是已通过 forbidden-output guard 的 view model；任一来源不能渲染时，不得读取其 rows。
- UI 不展示 raw token、credential、digest、Authorization、membership assertion、issuer、email、run input / output、provider endpoint、SQL、DSN 或内部异常。
- source request ID / audit ref 可作为脱敏可追溯信息展示；失败项不回显候选 workspace 或成员列表。
- dev headers、signed test token 和 OIDC 的生产边界完全沿用 Workspace-scoped Read Transition；本专题不改变授权来源。

## 批次 A：首分页运营关注队列

本批实现：

1. 新增纯 `WorkspaceOperationsInbox` projector、coverage、稳定严重度与排序。
2. 接入四类现有 workspace view model，不新增 API、schema、repository、migration 或 owner。
3. Web 增加收件箱 section、来源覆盖、关注项和既有详情跳转。
4. 覆盖跨来源排序、14 天到期边界、partial / unavailable、零关注项、敏感字段缺席和 workspace 切换后重新投影。

开发 / 测试态验收：

- 四来源均成功时，关注项只来自当前 active workspace 的当前读快照。
- 一项来源拒绝不遮蔽其他成功来源，但整体状态必须为 `partial`。
- 全部来源拒绝时不展示资源项，整体状态为 `blocked`。
- `next_cursor` 显式降级为 `partial_window`。
- 导航只复用现有 selection handler，不产生写请求或执行副作用。
- Web tests、production build、仓库 fast 检查通过；因本批新增功能真相源文档，补跑全量仓库检查。

实施状态：已完成。Web 已新增纯 projector 与运营收件箱 section；四类 coverage、严重度、14 天到期窗口、稳定排序、现有详情导航和读写停止线已接入。live workspace snapshot 在 loading / failed 期间固定为 `blocked`，不会消费离线 fixture 或上一工作区 rows；单来源拒绝只降级该来源，不遮蔽其他已授权关注项。

完成锚点：`workspace_operations_inbox_dev_test_v1_batch_a_complete`。

## 后续批次入口

批次 B 只有在真实用户需要跨全部分页窗口的服务端运营队列，且现有四类 owner 能提供统一、稳定、可审计的分页读取契约时才评审。届时应设计独立 read projection 和 cursor，但仍不得复制业务记录或引入自动修复。

本批完成后优先进行真实使用反馈，不继续派生同层 readiness、refresh、fixture 或 checker。

## 停止线

- 不接真实 Radish OIDC、membership adapter、quota / billing owner、production secret 或外部项目。
- 不新增 incident、task、notification、acknowledgement 或 remediation 真相源。
- 不自动撤销 / 轮换 API key，不激活 workflow，不 replay / resume run，不修改 application。
- 不把首分页计数、estimated cost 或缺少关注项解释为完整 usage、健康度、SLA 或生产结论。
- 不从 Applications、API Keys、Workflow Definitions、Runs 之外猜测跨资源关联。

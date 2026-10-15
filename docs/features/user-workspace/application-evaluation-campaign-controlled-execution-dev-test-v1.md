# 应用评测计划、受控执行与证据归档（开发 / 测试态）v1

更新时间：2026-10-15

状态：`application_evaluation_campaign_controlled_execution_dev_test_v1_completed`

既有执行与证据链保持完成；2026-10-15 新增的 Prompt 计划编辑体验处于方案设计阶段，尚未获准实现。其范围、交互和实施计划见下文，不改变已完成批次及生产边界。

## 功能定位

本专题为内部应用开发者提供一条可复验的应用回归执行链：把经过显式审查的评测输入冻结为不可变计划版本，在 `development | test` 环境内按单一执行 Profile 顺序产生 durable Run，再把两个 campaign 的同项 Run 精确交接给既有 Comparison、Evaluation Case、Suite 与人工 decision owner。

它解决的是当前真实断点：现有 Run History、Comparison、Evaluation Case 和 Suite 可以审查已经存在的 Run，却不能保存一组受控评测输入、重复产生同构 Run、证明一批 Run 使用同一 authority snapshot，也不能把逐项 Run 安全地组织成可供现有评测 owner 消费的证据集合。

本专题不创建第二套 Run、Comparison、Case、Suite 或 release decision。执行仍由现有四类应用运行服务完成，评测 owner 只保存计划、执行批次、逐项状态和既有资源引用。

## 用户任务

1. 在当前 application 下创建评测计划，选择一个真实执行 Profile，并录入一组明确用于评测的输入。
2. 保存计划版本；每次修改通过 `expected_version` CAS 生成新的不可变版本，历史版本不被覆盖。
3. 对 exact plan version 发起 campaign，选择当前 actor 在该 application 下仍为 active 的 API Key 作为 quota consumer，并确认环境、项目范围、项目数、当前 authority 摘要和会消耗的 UTC 日请求配额。
4. 服务顺序执行计划项，每项只调用对应现有执行服务一次，并保存 durable `run_id`、状态、失败边界和 authority digest；不保存 provider 原始响应或运行输出。
5. authority 在批次中漂移、应用归档、权限不足、quota admission、store failure 或运行服务合同不匹配时失败关闭，不切换到其它 candidate、assignment、definition、provider 或 model。
6. 选择同一 plan version 的 baseline campaign 与 candidate campaign，读取逐项兼容性和既有 Comparison 结果。
7. 显式确认后，把逐项 exact Run refs 交接给既有 Evaluation Case，再把 exact case versions 组成既有 Suite；交接只生成审查证据，不触发发布或运行时切换。

## 范围与环境

唯一资源作用域为：

```text
tenant_ref + workspace_id + environment + application_id
```

- `environment` 只允许 `development | test`，且必须与当前服务进程声明的 provider route / quota environment 一致；`production` 永远拒绝。
- plan、plan version、campaign、campaign item 和 handoff 必须保持同一完整作用域。
- application lifecycle、application kind 与执行 Profile 必须在保存计划和每次执行前重新读取。
- 跨 tenant、workspace、environment 或 application 的引用不返回存在性细节，也不回退到其它资源。

## 领域所有权

| 领域 | owner | 本专题行为 |
| --- | --- | --- |
| 评测计划 | Application Evaluation Plan owner | 保存当前记录、不可变版本、输入 fixture 与 CAS 修订 |
| 执行批次 | Application Evaluation Campaign owner | 保存 exact plan version、authority snapshot、逐项执行状态和 Run refs |
| 应用调用 | 既有 Definition / RAG / Prompt / Agent invocation service | 继续负责 authority checkpoint、provider 前 quota admission、Run 写入和终态 |
| Run | 既有 Workflow / Application Run store | 继续是运行状态、usage、failure 和 side effect metadata 的唯一真相源 |
| Comparison | Workflow Run Comparison | 按既有 Profile 兼容矩阵即时比较 exact Run refs |
| Case / Suite / Decision | Workflow Evaluation owner | 接收显式确认后的 exact refs，不被 campaign owner 复制 |
| Provider / quota | Gateway owner | 继续负责 route selection、provider attempt 与原子 quota admission |

## 资源模型

### Plan current record

`application_evaluation_plan.v1` 保存：

- `plan_id`、`record_version`、`latest_plan_version`；
- 完整作用域、名称、执行 Profile；
- `item_count`、当前版本 digest、创建 / 更新时间；
- actor、request、audit refs；
- lifecycle 固定为 `active | archived`，v1 不提供删除。

Plan current record 不重复保存 fixture 正文。读取具体输入必须读取 exact immutable version。

### Immutable plan version

`application_evaluation_plan_version.v1` 保存：

- `plan_id + plan_version`、`previous_plan_version`；
- `execution_profile` 和 profile-specific target；
- 有序 items；
- canonical `plan_digest`；
- 创建 actor、request、audit refs 和时间。

每个 item 具有稳定 `item_key`、名称、预期 Comparison classification 和一种 profile-specific fixture。顺序属于版本合同；同一版本内 `item_key` 唯一。

### Campaign

`application_evaluation_campaign.v1` 保存：

- `campaign_id`、`client_campaign_key`、exact `plan_id + plan_version + plan_digest`；
- 完整作用域、执行 Profile；
- 只用于 quota 归属的 `quota_api_key_id`，不保存 credential token；
- 发起时冻结的 authority snapshot 与 `authority_digest`；
- `pending | running | succeeded | failed | interrupted` 状态；
- 当前 item index、成功 / 失败计数、失败码和失败摘要；
- 有序 campaign items；
- 创建、开始、完成、actor、request 和 audit refs。

每个 campaign item 只保存 `item_key`、确定性 `run_id`、运行状态、Run schema / profile、failure code / boundary、authority digest 和时间。不得保存输入正文、输出、回答、context、artifact content、prompt 或 provider raw response。

`client_campaign_key` 在完整作用域内提供幂等创建；同 key 携带不同 plan version 或 digest 时返回冲突，不生成第二批运行。

## Profile 与 fixture 合同

一个 plan version 只允许一个执行 Profile，避免一个 campaign 混合不同 authority owner。

| 执行 Profile | application / target | fixture | 既有 Run |
| --- | --- | --- | --- |
| `workflow_definition_executor_v1` | workflow application + exact definition id / pointer version / definition version / digest | `input_text`、`condition_values`、可选 model / temperature | v5 |
| `workflow_definition_executor_v2` | workflow application + exact definition / pointer / input contract id 与 digest | `inputs`；按不可变合同校验的扁平 typed fixture | v8 |
| `application_rag_invocation_v1` | RAG-capable application + active RAG assignment | `input` | v4 |
| `prompt_application_invocation_v1` | `prompt_application` + active Prompt assignment | `variables` | v6 |
| `agent_copilot_suggestion_v1` | `agent` + active Agent assignment | `task`、`locale`、可选 conversation id、artifacts、context | v7 |

fixture 是开发者主动提交的测试材料，不从历史 Run 反推。服务复用各执行 Profile 的既有大小、UTF-8、结构、secret material 和输出合同校验；计划 owner 还必须在持久化前拒绝 credential、token、Authorization header、cookie、DSN、私钥和明显 endpoint secret。

v1 不保存图片二进制、文件上传、外部 URI 抓取或任意 connector 凭据。Agent artifact 只允许现有 canonical contract 可接受的结构化、无 secret 测试材料。

## Authority 一致性

campaign 开始前读取一次目标 authority 并生成 profile-specific snapshot：

- Definition：definition id / version / digest、activation pointer version、application record version；
- RAG：runtime assignment id / version / digest、publish candidate、draft、binding、application record version；
- Prompt：assignment id / version / digest、candidate、draft、template、protocol / model eligibility、application record version；
- Agent：assignment id / version / digest、candidate、draft、profile、policy、protocol / model eligibility、application record version。

每项执行前重新读取 authority；摘要必须与 campaign snapshot 一致。现有执行服务仍进行 provider 前 checkpoint。Run 落库后，campaign 再校验 Run 内 authority metadata 与 campaign snapshot 一致，任一不一致立即终止剩余项。

campaign 不锁住发布或 assignment owner，也不通过长事务包住 provider 调用；它通过 snapshot、逐项 checkpoint 和失败关闭证明一致性。

## 执行语义

- v1 仅支持显式同步、顺序执行，最大 `20` 项。
- `failure_policy` 固定为 `stop_on_failure`；不提供并行、continue-on-error、自动重试、fallback、resume 或 replay。
- 每项执行前先持久化确定性 `run_id` 和 `running` 状态，再调用既有服务；Prompt / Agent 使用由 campaign 生成的 `client_invocation_key`，Definition / RAG 注入相同确定性 Run ID。
- 专题 gate 要求 API Key lifecycle、API Key dev/test auth 与 Gateway quota enforcement 已显式启用，且 campaign environment 与 quota environment 一致；未满足时服务拒绝启动该能力。
- HTTP 发起前重读当前 actor-owned API Key；key 不存在、跨 application、已撤销、已过期或 store 不可用均失败关闭。服务只保存 `api_key_id`，不读取或返回 credential token。
- 每个 item 将所选 `api_key_id`、确定性 `run_id`、完整 quota scope 与既有 inference route 注入 Gateway bridge。`run_id` 保持 durable Run 根身份；单 provider 调用可直接使用该根身份，多 LLM 节点 Workflow 则按 `run_id + node_id` 派生确定性的 provider-attempt request identity，保证同一 Run 内每次真实 provider attempt 独立准入。provider 调用仍在现有原子 admission 之后发生，campaign 不预留或绕过额度。
- quota exhausted、policy missing、attempt conflict 和 quota store failure 作为 item / campaign 失败证据保存，剩余项不执行。
- HTTP 连接取消不会触发后台继续运行。若进程中断，重启 reconciliation 只根据确定性 Run ID 收口 campaign 状态，不重放 provider 调用。

## Comparison 与 Evaluation 交接

- baseline 与 candidate campaign 必须引用同一 `plan_id + plan_version + plan_digest`、完整作用域和 Profile。
- 每个 item 按稳定 `item_key` 配对，Run 必须存在、终态、Profile 匹配且满足既有副作用限制。
- preview 即时调用既有 Comparison，不在 campaign owner 复制 finding 或分类真相。
- handoff 必须由用户显式确认，使用两个 campaign 的 expected record version 防止并发漂移。
- 每个 item 创建一个既有 Evaluation Case：baseline Run 为 baseline campaign item，candidate Run 为 candidate campaign item，expected classification 来自 immutable plan version。
- candidate campaign 是 pair handoff 的唯一持久化锚点；baseline / candidate 的 expected record version 都要先匹配，避免以漂移 pair 创建证据。
- 每创建一个 Case 就把 exact `case_id + version` checkpoint 为 `partial`；全部 Case 成功后才创建 Suite 并把同一锚点推进为 `complete`。candidate campaign 只保存 exact case refs、suite id 和 handoff audit，不复制 review digest 或 decision。
- 任一 Case / Suite 创建失败时停止，不声称原子回滚已经进入 append-only owner 的证据；已成功资源作为 partial handoff 明示返回，允许人工审计，不自动补偿或删除。

## 权限

- 读取 plan / version / campaign：`application_evaluations:read`。
- 创建、修订、归档 plan：`application_evaluations:write`。
- 发起或 reconcile campaign：`application_evaluations:execute + workflow_runs:execute`，Definition 额外要求 `workflow_definitions:read`。
- preview / handoff：`application_evaluations:read + workflow_runs:read`；materialize 额外要求 `workflow_evaluations:write`。
- 所有 mutation 继续经过 workspace membership 与 resource binding 校验；dev signed header 只用于开发测试态，不宣称 production membership / OIDC。

## API 轮廓

统一位于 `/v1/user-workspace/applications/{application_id}/evaluation-*`：

- Plan：create、list、read、revise、archive、version list / read；
- Campaign：execute、list、read、reconcile；
- Review：`POST .../evaluation-campaign-pairs/preview` 做 baseline / candidate 即时比较，`POST .../evaluation-campaign-pairs/handoff` 显式 materialize exact Case / Suite refs。

所有 JSON body 拒绝未知字段，query 使用严格 allowlist，list 使用 scope-bound cursor，mutation 使用 `expected_version`。失败 envelope 必须返回稳定 `failure_code`、`failure_summary`、当前版本和 `audit_ref`；权限、环境、scope、store contract 和不存在均失败关闭。

## 持久化

三种模式保持同一合同：

- `memory_dev`：有界内存 owner，服务重启不承诺恢复；
- `sqlite_dev`：共享 local persistence runtime，新 migration，事务化 plan revision / campaign checkpoint / handoff refs；
- `postgres_dev_test`：独立 migration，行锁与 CAS，显式 migration 前置，不自动迁移。

数据库只保存经过合同校验的 plan / version / campaign JSON 和必要索引列。SQLite / PostgreSQL 都必须覆盖 restart read、并发修订、campaign 幂等、scope 隔离、部分写失败和 interrupted reconciliation。

## Web 产品面

该页面族包含新的结构性任务模型，按五维评分预计为 `8`，覆盖级别为 `A / 完整 Pencil`：

- Plan 列表、版本与 fixture 编辑；
- authority / environment / quota 边界确认；
- campaign 逐项执行进度与 Run refs；
- baseline / candidate 配对、Comparison preview 与显式 handoff；
- permission、archived application、environment mismatch、authority drift、quota、version conflict、missing policy、store failure、interrupted / partial handoff 状态；
- Desktop、关键断点与 `390×844` 的响应式顺序、选中语义和无横向溢出。

2026-08-10 已在用户确认设计源空闲后完成 `S10` Desktop、Narrow 与共享 Decision R15，并据此实现 React 功能纵向切片；但首版 R1 后续因 dashboard 式进度卡、稀疏工作面和页面骨架没有继承 S1–S8 被人工退回。Visual R2 在相同根节点恢复 `264px` 产品导航、薄页眉、evidence path、selected campaign 单一 owner、连续 item rows 和单一 handoff rail 后，又因业务表面全部硬方形、棱角语言与 S1–S8 不一致被人工退回。Visual R3 已保留连续信息结构，并按 `8–11px` 职责圆角修订任务、上下文、owner、boundary 与 handoff 表面，于 2026-08-12 完成人工视觉复核。2026-08-13 React 已逐项采用该基线：selected campaign context 驱动唯一 campaign 主 owner，item 使用连续 evidence rows，后续 Handoff 只占一个辅助 rail；功能契约、strict consumer 与真实数据库证据没有改变。

普通 plan、campaign 和 item 行保持中性；只有驱动当前详情的对象使用墨蓝选中轨。failed、quota exceeded、interrupted、partial 或 blocked 只使用文字、图标与状态色，不冒充选中。

## Prompt 计划结构化编辑与版本回归体验（2026-10-15 方案）

### 目标、依据与本轮授权

目标用户为内部应用开发者。目标任务是维护一组 Prompt 回归样本，形成明确的计划版本，按同一计划版本运行基线与候选批次，再进入既有比较、用例 / 套件和人工审查。当前 JSON 编辑方式对实际耗时、求助次数和输入错误率的影响尚未取得用户基线，不预填改进收益。

项目所有者本轮只授权核对代码与契约、完善本专题及必要 Pencil 设计，然后确认实现方案。暂不修改产品代码，不启动服务或真实模型，不扩大 API、schema、权限定义、持久化或执行范围。

已核对的实现依据如下；它们是本方案约束，不是待重写的领域：

| 代码与契约入口 | 已有事实 | 对方案的约束 |
| --- | --- | --- |
| `apps/radishmind-web/src/features/control-plane-read/ApplicationEvaluationPlanOwner.tsx` | Plan 使用目标和样本 JSON；仅 Definition v2 接入结构化输入编辑 | 在既有 S10 Plan 区域增加 Prompt 编辑，不重建工作区或扩展其它 Profile |
| `applicationEvaluationCampaignPanel.tsx` | 负责计划选择、精确版本读取、编辑、保存确认、批次和配对；当前版本读入编辑器是显式操作 | 编辑状态须绑定计划与读取版本；选中、载入、修订、执行不能混成同一动作 |
| `applicationEvaluationCampaignConsumer.ts` | 严格解析既有 Plan / Item 联合结构、1–20 项与分类；当前前端名称上限和标识校验比服务端宽松 | 本轮编辑反馈对齐服务端既有规则；复用原编码器，不把编辑器行标识、模板来源或本地校验结果加入正文 |
| `services/platform/internal/httpapi/application_evaluation_campaign.go` | Prompt target 为 `workflow_definition: null`；fixture 只有 `variables`；保存不解析当前模板 | 不把模板引用塞入 target，不以 UI 校验替代服务端合同，不宣称计划固定了模板版本 |
| `promptApplicationRuntimeConsumer.ts`、`promptApplicationTemplateConsumer.ts` | 可读当前 Assignment，再按 `promptTemplateRef` 精确读取不可变模板版本 | 复用既有只读请求与严格消费端，不从模板列表第一项、草案或旧缓存猜测合同 |
| `prompt_application_template_render.go`、`prompt_application_invocation.go` | 五种变量类型、必填 / 默认值、未知变量拒绝和字节预算已有规则 | 前端反馈与既有规则对齐，运行时检查仍由原服务负责 |
| `application_evaluation_campaign_runtime.go` 及本专题执行合同 | 每批冻结执行依据，逐项复核；配对要求相同精确计划版本，比较不读答案正文 | 模板可在两批之间显式变更；样本变化必须形成新计划版本，不能混配 |

表中未写目录前缀的 TypeScript 文件与首行同目录；Go 文件位于 `services/platform/internal/httpapi/`。

### 范围与不变条件

- 仅面向 `prompt_application_invocation_v1` 的 Plan 编辑：样本新增、复制、删除、上移 / 下移、选择，名称、`item_key`、预期分类及五类变量输入；保留高级 JSON。
- 新建和修订继续使用现有 Plan API。样本 `item_key` 在编辑、重排及修订中保持；复制样本生成新的唯一标识。编辑器内部行标识与业务标识分开，改名不重建 DOM，不进入 API。
- 既有 1–20 项、`item_key` 的 `^[a-z][a-z0-9_-]{0,63}$` 及唯一性、五种预期分类和原有隐私 / 大小限制保持。服务端计划名为 2–120 个 Unicode 字符，样本名为 1–120 个；当前前端均使用更宽的 160 上限，实施时在本编辑路径补齐与服务端一致的反馈。空计划可在页面编辑，但不能保存。
- 不新增样本库、文件导入导出、自动生成样本、定时执行、答案评分、结果正文捕获或批次重试。不把试用素材自动写入用户计划。
- Plan 的 `record_version` 用于并发写入，`plan_version + plan_digest` 表示不可变输入。模板版本、Assignment 版本、Campaign 版本分别属于原负责模块，界面不得混称为一个“版本”。
- 现有 Campaign / Pair / Handoff 只补必要的精确计划提示和入口衔接；执行服务、配额、幂等键、失败即停、Case / Suite 及人工决定的语义不变。其它 Profile 的 JSON 与 Definition v2 编辑器行为保持原状。

### 编辑参考模板的读取与失效

1. 用户显式选择“读取当前模板”，使用当前 tenant / workspace / application / actor 范围请求 `GET .../applications/{application_id}/prompt-runtime-assignment`，不读取事件列表。
2. 仅在 Assignment 为 `active` 时，按其 `template_id + template_version` 调用既有模板版本读取；严格核对作用域、模板 ID、版本和 digest，读取后再核对 Assignment 仍为同一版本 / digest。不同即返回“绑定已变化”，不混用两个响应。
3. 此操作需要已有 `prompt_application_runtime:read` 和 `prompt_application_templates:read_source`。没有权限、不存在、撤销、离线或读取失败时，不显示“已匹配模板”，也不使用示例变量替代；高级 JSON 仍按原评测权限提供已有功能，不授予模板源码访问权。
4. 页面只展示变量声明及精确来源摘要，不展示模板消息正文。既有 API 会返回源码，调用及临时校验所需数据仅留在当前作用域内存，不写浏览器持久化、URL、日志、计划或 Campaign。
5. 来源标记为“编辑参考”，不标为执行绑定。刷新参考或保存前的显式审查若发现 Assignment 改变，保留同作用域样本，标记“需要重新检查”；只有用户确认采用新参考后才重新计算字段反馈。缺参、旧字段和类型不匹配均保留原值供处理，不自动迁移。
6. 切换应用、工作区、身份或退出页面使参考读取和确认失效，拒绝迟到响应并清除敏感编辑内容。相同作用域的语言切换、样本选择和普通输入不触发网络读取、不更换编辑器身份。

不可变 Plan 不保存本次参考模板；刷新页面后不能宣称恢复了当时使用的模板合同。实际执行仍由 Campaign 在开始时读取当前有效依据并逐项检查。模板变化不改写已保存样本，缺省值的变化也可能影响实际渲染结果。

### 一份编辑状态与两种表达

采用 Prompt 专用、具有明确模式的编辑状态：`form` 保存可编辑样本与暂未合法的字段文本，`json` 保存完整样本数组 JSON 原文；同一时刻仅一个分支是待提交内容来源。Prompt target 固定沿用 `workflow_definition: null`，不再提供无效目标配置；计划名称继续由共同字段编辑。既有不可变版本只用于载入及脏状态比较，不作为隐藏的“最近合法内容”参与提交。

- 表单 → JSON：仅在所有字段可无损表达为 JSON 时生成一次原文并切换；数字中间态等不可表达内容会定位到字段，不能用上一次合法值代替。
- JSON → 表单：严格解析一次并核对结构；语法错误、未知协议字段、重复标识或不支持的表达保留原文并阻止切换。不得在每次输入时解析后重写文本，不对字符串、数组、空白或换行做静默清洗。
- 已声明字段类型不匹配、`null`、未声明变量须以“待处理原值”保留，可在高级 JSON 明确修正或删除；编辑其他字段不丢失这些值。重复 JSON 属性名和可能造成精度损失的数值不得经普通解析后静默覆盖 / 舍入再提交。
- 前端表单使用已有可安全表达的整数与有限数规则；无法无损承接的值保留原文并说明当前网页限制，不假装服务端禁止该类型，也不扩大数值协议或引入新的解析依赖。
- 模式切换、增删排序和字段编辑都会使先前保存确认失效；确认内容从当前完整编辑状态重新产生。部分合法字段不能绕过其余未完成编辑提交。

字段行为约定：

| 类型 / 状态 | 控件与反馈 | 序列化要求 |
| --- | --- | --- |
| `string` | 按变量说明提供文本框或多行输入，显示必填 / 可选 | 空字符串是已提供的值，与未提供不同；不自动 trim |
| `integer` / `number` | 使用文本输入与数字键盘提示；保留负号、小数点等输入中间态，失焦 / 审查时定位错误 | 不把空值、非法文本转换为 `0`；不提交舍入值 |
| `boolean` | 先区分“未提供”和“提供值”，再选择 true / false | `false` 是明确提供，不等于缺省 |
| `string_list` | 字符串条目列表，支持增删和上下移动 | 保留条目顺序、空字符串与换行；`[]` 与未提供不同，不按逗号拆分 |
| 可选与默认值 | 显示默认值来源；默认选项为“未提供”，可显式填写 | 未提供时省略该键，由运行时应用模板默认值；不得把默认值隐式写入样本 |
| 未声明 / 类型不符 | 显示变量名、原值和原因，提供明确编辑入口 | 不删除、不强制转换，不从其他样本或历史运行补值 |

本地反馈分两层：计划格式不合法时阻止保存；结构化模式只有在当前编辑参考可用且样本匹配时允许审查保存，否则定位字段或来源问题，提供显式转入高级 JSON 的入口。高级 JSON 中合法的既有计划可经现有保存确认保存，并明确显示“未完成模板匹配 / 存在不匹配”，保留原有反例样本用途；不能因此展示可运行或预期拒绝已通过。Campaign 仍失败即停，`expected_classification` 不是预期失败码。

### 计划版本、冲突与连续使用

- 当前选中计划、已读取不可变版本、正在编辑的基准计划及 CAS 版本须分别记录并在上下文中说明。创建新计划有明确入口；不得因选择列表第一项把新草稿当成该计划的修订。
- 同作用域选择另一计划、载入版本、切换 Profile 或丢弃编辑前，若存在未保存内容，提供“继续编辑 / 放弃并继续”；切换样本不丢弃草稿。身份失效及作用域切换直接按原隐私规则清理，不用确认框阻止撤权。
- 载入已有版本是用户显式操作。保存成功后使用服务端返回的规范版本确认当前基准，保持样本身份和选中项；列表刷新不能重新应用旧保存回执或覆盖新编辑。
- CAS 冲突保留本地内容并停止本次提交；显示读取时版本和服务端可取得的当前版本。只有显式重读并选择载入远端后才替换内容；不自动覆盖、合并、提升 `expected_version` 或重试。
- 新增历史版本下拉、版本差异引擎和跨计划克隆不在本轮范围；沿用已有精确版本读取，当前版本、已载入版本与批次引用必须可辨认。
- 保存后可显式进入既有 Campaign 任务，执行确认清楚列出所选 `plan_id / plan_version / digest`。未保存编辑不会进入执行请求，也不因保存自动执行。
- 基线与候选批次引用同一精确计划版本时才可比较；两批之间更新应用模板或模型仍走原发布与绑定流程。改变样本应保存新计划版本并重新建立该版本下的两批证据。
- Pair / Handoff 继续显示实际 Comparison 分类和 exact Case / Suite 引用；人工决定不激活绑定、不发布。答案质量与事实正确性仍需另行人工试用，不从 metadata 的 `changed / unchanged` 推导。

### 交互状态与响应式

| 状态 | 可见内容与允许动作 | 必须保持 |
| --- | --- | --- |
| 新建 / 无样本 | 空列表、添加样本、读取模板入口 | 无示例自动提交；0 项不能审查保存 |
| 参考读取中 / 不可用 | 来源加载状态或精确失败原因，可重读或进入已有 JSON 编辑 | 不使用旧来源冒充成功，不扩大权限 |
| 正常编辑 | 当前样本唯一选中，字段、元数据、增删排序，未保存标记 | 稳定焦点、单一待提交内容 |
| 字段 / JSON 错误 | 错误计数、具体位置与原文 | 不回退到旧合法值，不丢失错误输入 |
| 模板变化 / 未匹配 | 来源版本及需要复核的变量，显式刷新 / 采用 | 不修改样本、不宣称执行已锁定 |
| 保存审查 / 等待 | 操作种类、计划 / CAS 版本、样本数、匹配状态；明确确认 | 编辑使确认失效；等待禁止重复提交 |
| CAS 冲突 / 写入失败 | 保留草稿、稳定错误与恢复入口 | 不自动合并、重试或提升版本 |
| 归档 / 只读权限 | 已有计划和版本可读，写操作不可用 | UI 不授予能力，服务端继续鉴权 |
| 切换 / 退出 | 同作用域丢弃提示；身份 / 作用域失效清理 | 迟到响应不得恢复旧正文 |

桌面继承 S10 现有壳与任务导航；Plan 内为紧凑样本列表和当前样本编辑区，不新增并列产品导航或等权卡墙。主要动作是“审查保存”，复制、排序和删除是样本局部动作；“进入批次执行”只在已保存版本上下文出现。

在工作区可用宽度不足时，样本列表收为选择入口；窄屏按计划上下文 → 参考模板 → 样本选择 → 当前字段 → 错误与保存动作顺序排列。移动端保留字段标签、键盘访问和明确文字，不缩放桌面表格；高级 JSON 仅其代码区域可横向滚动，页面本身不得溢出。验收视口为 `1440×900`、`720×900`、`390×844`。

### Pencil 覆盖与待确认设计

五维评分为结构 `1`、交互 `1`、边界 `1`、状态与窄屏 `2`、复用 `0`，合计 `5`，采用 `B / 局部 Pencil`。复用 S10 和已评审结构化输入表面，仅补 Plan 桌面代表区域、窄屏重排及无法从正常态推导的关键失败 / 确认状态；不建立 S11，不重画 Campaign 或 S6。

设计源继续使用 `docs/designs/radishmind-web-family-ui-v1.pen`，新增四块局部画板，均标记为 R1 待确认：

| 节点 | 覆盖内容 |
| --- | --- |
| `q3ngo0` | 桌面 Plan 代表区域：版本上下文、参考模板、连续样本列表、字段表单和保存入口 |
| `a3sRe` | `390×844` 窄屏：样本选择、元数据与已填字段折叠、当前字段和保存动作 |
| `fFIKS` | JSON 错误、参考模板变化、保存审查、CAS 冲突及明确恢复入口 |
| `MsAk8` | 数字中间态、布尔缺省与 false、列表空字符串、未声明值保留和设计决定 |

2026-10-15 四块画板共 250 个节点的原生静态检查未发现裁切、placeholder 或硬编码 fill / stroke；四张实际导出预览已逐张复核，修正了空字符串占位提示与布尔两态示例标注，并显式保存设计源。预览与检查记录只属于本地设计证据；方案和画板尚待项目所有者确认，不等于 React 实现、浏览器验收或已采用视觉基线。

### 实施顺序与验收计划（待批准）

| 顺序 | 实现范围 | 对应验证与交付 |
| --- | --- | --- |
| 1. 编辑状态与编码 | Prompt 专用模式状态、稳定行身份、样本动作、单一序列化入口；复用现有 Item 编码器与 Prompt 变量规则 | 无损模式转换、无效中间态、未知值保留、数值边界、空 / 默认 / false / 列表、标识唯一与 1–20 项单测；不新增协议 |
| 2. 来源读取与字段编辑 | 复用 Assignment / exact Template 只读消费端、来源核对和请求代次；按已确认设计接入 S10，补齐中英资源和语义 token | 无权限、撤销、读取中切换、迟到响应、模板变更与五种类型；语言切换无请求、无丢值 |
| 3. 版本及保存交接 | 在既有面板整合基准版本、脏状态、保存确认、CAS 恢复与已保存版本进入 Campaign | 新建 / 修订、冲突、失败、刷新、列表选择与保存响应归属；保持既有 Profile 和配对语义 |
| 4. 产品路径验收 | 经另行授权启动隔离固定响应服务，在现有 `evaluation` / `workflow` 组补充连续操作验证 | 多样本维护 → 保存恢复 → 同计划版本两批 → Pair → Case / Suite → 原有人工决定入口；失败定位后再继续，核对进程、端口及临时库清理 |

每步执行与变更相称的单测、严格类型检查和必要构建 / 包体检查；状态、作用域与 API 正文的行为断言优先，不用源码文本断言代替。产品代码完成后按风险运行快速仓库检查；功能真相源收口补跑全量检查。没有后端改动时不机械重跑全部数据库组合，若核对暴露必须修改后端 / API / 存储的缺口，应说明影响并重新确认范围。

使用既有五份故障诊断素材验证字段反馈，其中四份有效输入可用于固定响应批次；第五份缺参样本用于编辑错误 / 离线拒绝，不把它混入成功批次或答案质量分母。另以有界合成输入覆盖 boolean、integer、number、string_list 与可选默认值。真实模型、用户计时和答案质量仍待独立窗口。

本轮完成条件是范围与代码合同一致、交互和失败恢复可评审、必要 Pencil 画板完成并通过静态 / 渲染复核，随后由项目所有者确认实现方案。确认前不进入上述代码实施步骤。

## 实施批次

### 批次 A：设计合同与领域 owner

- 冻结 Profile / fixture / authority / failure matrix、API、权限、隐私和 handoff 语义。
- 实现 plan current + immutable version、campaign 领域合同、memory repository 和领域测试。

### 批次 B：HTTP 与三模式持久化

- 实现 strict HTTP、workspace binding、权限与 environment gate。
- 增加 SQLite / PostgreSQL migration、repository、CAS、cursor、restart 和并发测试。

### 批次 C：受控 campaign executor

- 复用四类现有执行服务和统一 Run owner。
- 完成 authority snapshot / checkpoint、确定性 Run ID、quota failure、stop-on-failure 与 interrupted reconciliation。

### 批次 D：Comparison / Case / Suite handoff

- 实现 campaign pair preview。
- 显式确认后交接既有 Case / Suite，保留 partial handoff 真实状态和 exact refs。

### 批次 E：Pencil、React 与浏览器验收（已完成）

- Desktop `Um8Zh`、Narrow `ZxJd7` 和共享 Decision R15 `UNMOS` 的 R1 与 Visual R2 均已被人工视觉退回；Visual R3 已显式保存，完成实际截图、零折叠、零硬编码色、零 placeholder 检查和 2026-08-12 人工视觉复核。历史 2× PNG 只属于 R1 证据，不再作为当前视觉基准。
- 已实现 strict consumer、四任务单 owner Workbench、完整失败关闭、Plan / Campaign / Pair / Handoff 交互和全视口响应式样式。
- Web `316/316`、S10 定向 `7/7` 与 production build 已通过；memory 真实浏览器已完成 Plan → 两次 Campaign → Pair → exact Case / Suite Handoff，`1440×900`、`1024×900`、`390×844` 无横向溢出，控制台零 warning / error，刷新后仍恢复 exact evidence。
- SQLite 真实页面以 `app_cssvwuvwodxmxecz` 创建 Plan `aeplan_lkqe7gr7kjobmf73 v1`。首个 Campaign `aecamp_slj6slzdz35qvhne` 在第二个 LLM 节点以 `workflow_run_gateway_failed / gateway` 失败并保留真实证据，由此定位同一 Run 复用 quota request identity 的根因；修正为逐 LLM 节点派生 provider-attempt identity 后，baseline `aecamp_2xrptdhto6nbj7vc` 与 candidate `aecamp_qr2wmglzcj5eortb` 均成功。Pair Preview 得到 `comparable`、expected `unchanged`、actual `changed`、mismatch `1`，显式 Handoff 生成 Case `eval_034d69aec0d7a2323c7f222f v1` 与 Suite `suite_9a8017d686be57009c7ad973`。再次重启服务后精确恢复同一 Case / Suite；`1440×900`、`1024×900`、`390×844` 均无横向溢出，控制台零 warning / error，页面未回显原始凭据。
- 2026-08-13 完成 Visual R3 React 迁移。真实 SQLite 数据下可在两个 succeeded Campaign 间切换，selected context、主 owner、连续 item rows 与 Handoff rail 始终保持单一职责；`Review handoff` 精确进入既有 Handoff 任务，不创建自动执行或第二套状态 owner。`1440×900`、`720×900`、`390×844` 的 document、workspace、context、owner、item 与 rail 均无横向溢出，最终控制台无 warning / error。

## 当前实现进度

- 批次 A 至 D 已实现：领域合同、memory / SQLite / PostgreSQL repository、SQLite `0016` / PostgreSQL `0019` migration、严格 Plan / Campaign / Pair HTTP、四 Profile 显式 executor 映射、authority checkpoint、确定性 Run ID、quota binding、interrupted reconciliation，以及 Comparison / Case / Suite handoff。
- Workflow Definition 结构化运行输入专题批次 D 在原 owner 上增加 Plan / Plan Version / Campaign v2 与第五个 `workflow_definition_executor_v2` Profile；typed fixture 只使用 `{inputs}`，Campaign 逐项重读 exact Definition / contract authority，Run v8 pair 交给 Comparison v7 和既有 Case / Suite。SQLite `0019` 与 PostgreSQL `0022` 扩展共享 Workflow Run Store；v1 资源不迁移、不回退，Campaign 与 Run 不复制 fixture 正文。
- PostgreSQL repository 使用行锁与 CAS；SQLite 与 PostgreSQL 都拒绝版本覆盖、资源删除和非法状态迁移。Case / Suite partial handoff 以 candidate campaign 为单一锚点逐项 checkpoint。
- 相邻测试覆盖计划版本、campaign 幂等与顺序执行、quota binding、多 LLM 节点独立 provider-attempt admission、authority drift、quota failure、reconcile 不重放、strict HTTP、权限、active API Key 归属、SQLite restart / corruption、PostgreSQL migration 与 repository contract、Comparison preview、complete / partial handoff。
- 批次 E 的设计、Visual R3 React、strict consumer、测试、build、三视口、memory 与 SQLite 真实浏览器连续链均已完成；本轮迁移没有新增 API、schema、migration、repository、permission、任务卡、fixture 或专项 checker，开发服务均已停止。
- SQLite 已完成 exact Plan → 两次 succeeded Campaign → Pair → Handoff，并在服务重启后恢复同一 Case / Suite。首个失败 Campaign 作为多节点 quota identity 根因证据保留；修正后双 LLM 节点回归证明每个 provider attempt 独立准入。专题关闭，不扩张 production membership / OIDC、production secret、billing、自动执行、自动重试 / 续跑、发布 / 部署或业务写回。

## 验收方式

- Go：领域、HTTP、memory、SQLite、PostgreSQL、四 Profile executor、authority drift、quota、store failure、reconciliation 与 handoff 相邻测试。
- Web：strict decoder、状态矩阵、交互、响应式和 build。
- 浏览器：`1440×900`、关键断点、`390×844`；验证层级、选中、执行确认、campaign pair、partial / blocked、横向溢出、storage 和 console。
- 仓库：`git diff --check`、`go test` / `go vet`、`./scripts/check-repo.sh --fast`；因新增 API、schema、migration、permission 与阶段真相，关闭专题前运行全量 `./scripts/check-repo.sh`。

## 停止线

- 不支持 production environment、production authentication / membership / OIDC、production quota 或 production enablement。
- 不支持 billing、token / cost limit、额度预留、自动提额、自动禁用或自动路由。
- 不支持 schedule、cron、queue worker、并行 fan-out、自动 retry、fallback、resume、replay 或长期 agent loop。
- 不保存 provider raw response、完整输出、对话 transcript、历史 Run 输入或从 metadata 反推的内容。
- 不自动 approve candidate、activate assignment、发布、部署、写业务真相或执行 proposed action。
- 不新增第二套 Run、Comparison、Case、Suite、decision、quota 或 route selection owner。
- 不为普通 UI 和既有测试可承载的行为新增专项 checker、fixture 或 readiness 链。


### 2026-10-11 双语展示迁移

页面已接入随功能加载的中英资源，提示保存稳定状态与参数，界面语言不影响原始输入、精确版本、摘要、权限或显式确认。异常正文不直接展示，保留稳定失败码及本地恢复说明；作用域变化拒绝迟到响应，等待时禁止重复提交。版本冲突保留本地输入，显式重载后须重新审查。完整覆盖与浏览器验收边界见 [i18n 专题](../web-internationalization-zh-en-v1.md#e-阶段-plan--campaign--schedule-覆盖清单2026-10-11)。本批三条流程中英各两轮共 12 次通过，不改变本功能原生产停止线；日程合成投影验证不构成真实调度器验收。

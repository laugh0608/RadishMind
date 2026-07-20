# 提示词应用模板版本审查与受控调用（开发 / 测试态）v1

更新时间：2026-07-20

状态：`prompt_application_template_version_review_controlled_invocation_dev_test_v1_design_confirmed`

## 功能定位

本专题把应用目录中已经允许声明的 `prompt_application` 建设为可创作、可审查、可激活、可调用和可复验的正式开发测试态产品能力。内部开发者可以在一个应用作用域内维护提示词模板源码、声明输入变量与输出契约、生成不可变模板版本，把精确版本绑定到既有应用发布候选，并在人工审查和显式运行时激活后通过 Application API 或 Application Interaction Session 受控调用。

这不是为现有应用配置页增加一个自由文本输入框。提示词模板是可执行的应用源码，必须拥有独立领域、版本、摘要、权限、持久化和隐私边界；应用配置、发布候选、运行时 assignment、Gateway、Session、Run History 与 Evaluation 继续各自承担已有职责。

## 现状与问题

- Application Catalog 与 Application Configuration Draft 已允许 `prompt_application`，但当前没有应用作用域的模板草案、模板版本、变量契约或运行时 authority。
- Application Configuration Draft 明确不保存提示词、消息、用户输入或模型输出；把模板正文直接加入该 owner 会破坏现有脱敏配置边界。
- Workflow Definition 支持 `prompt` 节点，但当前不可变定义保存的是节点结构、输入 / 输出摘要与引用，不是提示词源码真相源；不能把摘要字段解释为可执行模板正文。
- Application Interaction Session 当前只支持 `workflow_definition_executor_v1` 与 `application_rag_invocation_v1`，没有 Prompt Application profile。
- 仓库 `prompts/` 保存的是 RadishMind 自身的静态任务提示词，不属于用户应用资源，也不提供多租户、版本、审查或运行时绑定。

因此，`prompt_application` 目前只是目录与配置允许值，并没有与产品声明相匹配的真实用户路径。

## 目标用户与主要任务

首批目标用户是内部应用开发者与开发测试态审查人：

1. 应用开发者为当前 `prompt_application` 创建或恢复模板草案。
2. 开发者编辑有序消息模板，声明受控变量、默认运行参数和输出契约。
3. 系统执行确定性校验与预览；缺失变量、非法占位符、超限内容、敏感材料和不支持的输出契约必须失败关闭。
4. 开发者从已保存且校验通过的精确草案版本生成不可变模板版本。
5. 开发者把精确模板版本引用附着到新的 Application Configuration Draft 版本，不在配置草案中复制模板正文。
6. 既有 Application Publish Candidate 创建路径生成绑定模板版本的候选；审查人读取精确模板版本完成批准、拒绝、要求修改或撤回。
7. 批准不自动激活。具备独立权限的参与者显式创建或替换 Prompt Application runtime assignment。
8. Application API key 或 Application Interaction Session 解析当前 assignment，重读完整 authority 后只委托一次既有 Gateway 调用。
9. 用户从 Run History、Comparison、Evaluation 与 Application Operations 审查 metadata-only 运行证据；原始变量、渲染消息、模型输出和 provider raw response 不从运行记录恢复。

## 领域所有权

| 领域 | 真相源 / owner | 本专题中的职责 | 不允许承担的职责 |
| --- | --- | --- | --- |
| 应用身份与生命周期 | Application Catalog | 确认应用存在、作用域一致、类型为 `prompt_application` 且生命周期允许编辑或调用 | 不保存模板正文或运行时 assignment |
| 应用公开运行配置 | Application Configuration Draft | 保存协议、模型、允许协议与精确模板版本引用 | 不保存提示词正文、变量值、渲染消息或输出 |
| 模板源码 | 新增 Prompt Application Template owner | 保存模板草案、变量声明、输出契约、CAS 版本和不可变模板版本 | 不保存 provider credential、运行输入、运行输出或发布决定 |
| 应用发布审查 | 既有 Application Publish Candidate owner | 通过新 schema 版本绑定配置草案与精确模板版本，承担唯一人工审查状态机 | 不复制模板正文，不因批准自动激活或发布 |
| 运行时激活 | 新增 Prompt Application Runtime Assignment owner | 保存当前已激活候选与模板版本的 ref-only 指针、CAS 和事件 | 不复制配置或模板快照，不调用 provider |
| 模型调用 | 既有 Gateway | 执行协议适配、provider / profile / model 选择与单次上行调用 | 不管理模板草案、发布审查或 assignment |
| 交互编排 | Application Interaction Session | 通过显式 Prompt profile 委托同一个 invocation service，维护 metadata-only session / turn | 不持久化 transcript，不复制渲染器或 Gateway 算法 |
| 运行证据 | Workflow Run Store / Application Operations | 保存 authority、digest、状态、diagnostic、usage availability 和副作用计数 | 不保存变量、完整 prompt、answer 或 provider raw response |
| 质量评测 | 既有 Evaluation owner | 对明确的 Prompt profile 与 exact template lineage 做只读比较和开发测试态评测 | 不修改模板、候选、assignment 或运行记录 |

## 模板领域模型

### 模板草案

`prompt_application_template_draft.v1` 至少包含：

- `template_id`、`workspace_id`、`application_id`、`owner_subject_ref`；
- `template_name`、`description`；
- 有序 `messages`，角色只允许 `system | developer | user`；
- `variables`，每个变量包含稳定名称、类型、是否必填、说明与可选的安全默认值；
- `output_contract`；
- `draft_version`、`validation_state`、创建 / 更新时间和脱敏审计引用。

一个应用可以维护多个模板资源，但一个 Application Publish Candidate 只能绑定一个精确模板版本。列表只返回脱敏摘要；读取正文必须使用独立 `prompt_application_templates:read_source` 权限，普通 Application、Operations 和 Run History 读取权限不得获得模板源码。

### 变量与受限模板语法

首版使用受限占位符语法 `{{ variable_name }}`，变量名必须满足稳定标识符规则。只允许一轮直接替换，不支持表达式、函数、过滤器、循环、条件、include、宏、属性访问、文件读取、环境变量、URL 请求或任意代码执行。

变量类型固定为 `string | integer | number | boolean | string_list`。渲染时：

- 未声明占位符、缺失必填变量和额外变量均失败关闭；
- 每个变量只按声明类型做确定性规范化，不递归解释变量值中的占位符；
- `string_list` 使用规范 JSON 数组文本渲染，其余标量使用明确的 canonical 表达；
- 默认值属于模板源码，可以进入模板 owner，但不得包含 credential、token、header、cookie、DSN 或 provider secret；
- 模板、单条消息、变量数量、变量值和最终渲染请求必须有明确大小预算。

受限语法只保证渲染行为可预测，不宣称解决 prompt injection。审查界面必须展示变量契约、模板源码和使用合成值生成的预览；真实运行值不得进入候选、审查记录或 committed 证据。

### 输出契约

`output_contract` 首版支持：

- `text`：返回文本，受最大长度与空结果规则约束；
- `json_object`：要求返回单个 JSON object，并按受控 JSON Schema 子集校验字段、类型、必填项和额外字段策略。

输出不满足契约时运行失败，不自动修复、不追加第二次 provider 调用，也不降级为未校验文本。输出契约属于模板版本摘要的一部分。

### 不可变模板版本

只有已保存、校验状态为 `valid` 的精确草案版本可以生成 `prompt_application_template_version.v1`。不可变版本至少保存：

- `template_id`、`template_version`、来源 `draft_version`；
- 完整规范化模板源码、变量契约与输出契约；
- 服务端计算的 `template_digest`；
- 创建参与者、时间、请求和审计引用。

模板版本创建后不可更新或删除。修改必须回到草案，通过 CAS 保存新草案版本并生成新的不可变模板版本。模板版本本身不再建立一套平行审批状态；人工审查统一由绑定该版本的 Application Publish Candidate 承担。

## 与配置和发布治理的组合

### 配置草案精确绑定

新增 Application Configuration Draft schema 版本，只允许 `application_kind=prompt_application` 时携带：

- `template_id`；
- `template_version`；
- `template_digest`。

附着操作必须由服务端读取当前已保存配置草案和精确模板版本，只修改 ref-only binding 并通过既有 expected-version CAS 生成下一版草案。浏览器不能提交模板正文、伪造 digest，批准模板版本也不能自动修改配置草案。

非 Prompt Application 不得携带该 binding；现有 v1 / v2 草案继续保持兼容和原语义。

### 发布候选精确绑定

新增 `application_publish_candidate.v3`，在既有配置快照、RAG binding 兼容边界和审查状态机基础上增加 Prompt Template ref。创建、批准和读取资格时，服务端必须重读：

- 当前应用类型、生命周期与 revision；
- 精确 Application Configuration Draft 版本和 digest；
- 精确模板版本及其 digest；
- 模板与应用作用域；
- 模板变量和输出契约的有效状态。

候选只保存模板 ref / digest，不复制模板正文。审查界面通过 Template owner 独立读取源码；源码不可用、digest 漂移、作用域漂移或应用类型改变时必须阻塞审查与运行资格。

批准候选仍不代表正式发布，也不自动建立 runtime assignment。现有 v1 / v2 发布候选继续可读，不因 v3 引入而改写。

## 运行时 assignment 与 exact authority

`prompt_application_runtime_assignment.v1` 是 ref-only 当前运行指针，至少包含：

- tenant / workspace / application / owner scope；
- `assignment_version` 与状态；
- `candidate_id`、`candidate_review_version`；
- `draft_id`、`draft_version`、`draft_digest`；
- `template_id`、`template_version`、`template_digest`；
- actor、request、audit 和时间引用。

只有当前 `approved`、未漂移且类型匹配的 v3 candidate 可以被显式 `activate | replace`；`revoke` 只撤销当前 assignment，不修改候选、配置草案或模板版本。所有动作使用 expected assignment version，事件只追加且不得保存模板正文。

Prompt invocation service 在创建运行前和每次计划内 provider 调用前重读并比较 exact authority：

1. Application Catalog 当前记录、类型、生命周期与 revision；
2. 当前 runtime assignment 与 CAS 版本；
3. 精确 approved publish candidate v3 与 review version；
4. 精确配置草案版本、digest 和 template binding；
5. 精确模板版本、digest、变量契约与输出契约；
6. 默认协议、模型及当前 Gateway model catalog eligibility；
7. API key 或 Session actor 的应用 scope 与 `prompt_application:invoke` 权限。

任何漂移、缺失、归档、吊销、存储故障或权限变化都必须在 provider 副作用前失败关闭，不回退历史 assignment、旧候选、旧模板版本、默认 fixture 或内存存储。

## 受控调用与运行记录

新增显式 execution profile `prompt_application_invocation_v1`。Application API key 路径和 Application Interaction Session 都委托同一个 invocation service：

- 每次 invocation 只允许一次计划内 Gateway provider 调用；
- request 只提交变量值、client invocation key 和必要的 application scope，不能提交模板正文、版本、digest、provider credential 或 authority snapshot；
- 服务端按 exact template version 渲染 canonical messages，再使用已批准配置中的协议和模型调用既有 Gateway；
- 不复制 Gateway 协议适配、provider selection、取消或 request history 算法；
- client invocation key 的同步或并发重试只能读取既有 running / terminal evidence，不重复调用 provider；
- 取消映射为 `canceled`，provider 结果不确定或终态写入失败映射为 `outcome_unknown`，不得自动 replay。

批次 A 必须冻结新的严格 authority、Session / Turn 和 run record schema 版本；既有 `application_runtime_authority.v1`、Application Session v1、Turn v1 与 workflow run v0–v5 不原地放宽。新的 metadata-only run lineage 至少保存：

- execution profile、application / assignment / candidate / draft / template refs 与 digest；
- input digest / bytes、变量名集合摘要，不保存变量值；
- requested / selected protocol、provider、profile 和 model；
- started / completed time、status、failure code、diagnostic 与 usage availability；
- provider / tool / retrieval / confirmation / business write / replay 副作用计数；
- request、audit 和 actor ref。

原始变量、渲染消息、模板正文副本、模型输出和 provider raw response 只在当前请求内存中存在。成功响应可把当前输出返回调用方；幂等重试和历史读取不得根据 metadata 伪造输出或 transcript。

## Web 产品路径

Prompt Application 工作区作为既有 Application Development Workspace 下的独立 feature-owned surface，至少覆盖：

1. Template 列表、创建、恢复和 CAS 冲突处理；
2. 消息模板、变量、输出契约编辑；
3. 本地确定性校验与合成值预览；
4. 不可变模板版本创建与精确版本详情；
5. Template Version → Configuration Draft binding handoff；
6. Publish Candidate v3 源码审查、漂移与 blocker；
7. runtime assignment 的显式 activate / replace / revoke；
8. Prompt Application Session / 单次受控测试；
9. exact Run History、Evaluation 与 Operations handoff。

应用切换、revision 变化、归档、身份变化和 surface 卸载必须清除未保存模板、变量值、渲染预览、响应和迟到请求。稳定 URL 只允许阶段锚点与短资源标识，不携带模板源码、变量值或输出；不得使用 `localStorage`、`sessionStorage`、IndexedDB 或 cookie 恢复这些内容。

## API、存储和权限方向

后续实施任务卡应冻结以下开发测试态边界：

- Template Draft：validate、save、list、read；
- Template Version：create、list、read；
- Configuration binding：从精确 draft + template version 生成下一版配置草案；
- Publish Candidate v3：复用既有 create / read / list / review 路由与状态机；
- Runtime Assignment：read current、list events、activate / replace / revoke；
- Invocation：API key 单次调用与 Application Session profile 委托；
- Run / Evaluation：复用既有历史、详情、比较和评测入口，按新 lineage 显式识别。

权限至少区分模板摘要读取、源码读取、写入、版本创建、发布候选审查、assignment 管理和 invocation。`memory_dev`、聚合 `sqlite_dev` 与 `postgres_dev_test` 必须共享作用域、CAS、不可变版本、排序、失败和 no-fallback 语义；生产 repository、production auth 和正式发布继续关闭。

## 稳定失败语义

实施契约至少需要覆盖：

- template scope / payload / secret material / syntax / variable / output contract invalid；
- template draft not found / version conflict / store unavailable / write disabled；
- template version not found / immutable conflict / digest drift；
- configuration binding kind mismatch / source drift；
- publish candidate template unavailable / review drift / superseded；
- runtime assignment absent / version conflict / candidate ineligible / revoked；
- model or protocol ineligible；
- invocation input invalid / duplicate running / authority drift / canceled / outcome unknown；
- output contract validation failed。

错误响应只返回稳定 failure code、脱敏 summary、当前可公开版本和 request / audit ref；不回显模板正文、变量值、渲染消息、模型输出、credential、token、header、endpoint、DSN 或 provider raw error。

## 实施批次

### 批次 A：strict contract、确定性渲染器与 memory owner

- 冻结 Template Draft / Version、配置 binding、Publish Candidate v3、Runtime Assignment、authority、Session / Turn 与 run lineage schema。
- 实现模板领域校验、受限占位符 parser / renderer、输出契约 validator 和相邻单元测试。
- 实现 memory repository、Template Draft / Version API、作用域、CAS、不可变版本和敏感材料守卫。
- 形成高风险实施任务卡；批次 A 未通过前不打开 provider 调用。

### 批次 B：SQLite / PostgreSQL 开发测试态持久化

- 在共享本地产品 runtime 中加入模板与 assignment migration，不新增 DSN、pool 或 fallback。
- 完成 SQLite / PostgreSQL repository、schema marker、manual migration、rollback / reapply、运行角色、重启、并发、corruption 和 no-fallback 验证。
- 数据库扫描证明运行变量、渲染消息、输出与 credential-like 材料不进入持久 owner。

### 批次 C：配置、发布审查与 runtime assignment

- 实现 ref-only Configuration Draft binding 与 Publish Candidate v3。
- 复用既有发布审查状态机，补 exact template reload、源码审查投影、漂移和 supersede 语义。
- 实现 assignment activate / replace / revoke、事件、CAS 和 read-time eligibility；candidate approve 不自动激活。

### 批次 D：受控调用、Session、Run 与 Evaluation

- 实现 `prompt_application_invocation_v1` exact authority resolver 和 provider 前 checkpoint。
- API key 与 Application Session 委托唯一 invocation service，单次调用既有 Gateway。
- 接入新的 metadata-only run lineage、History / Detail / Comparison / Evaluation / Operations，并保持旧 lineage 兼容。
- 覆盖幂等、并发、取消、漂移、终态不确定、输出契约失败和零自动重放。

### 批次 E：Web 工作区、双数据库连续链与浏览器验收

- 完成模板创作、版本、binding、发布审查、assignment、受控测试和运行 / 评测交接。
- 完成 SQLite / PostgreSQL launcher profile、服务重启恢复和连续用户路径。
- 真实浏览器验证应用切换、冲突、漂移、批准但未激活、显式激活、调用、取消、Run / Evaluation handoff、URL / storage / console / network 和敏感扫描。
- 同步功能专题、任务卡、能力矩阵、current focus、集成契约和周志后关闭专题。

## 验收方式

- contract：JSON Schema、Go codec、TypeScript strict consumer 均拒绝未知字段、旧 schema 漂移和敏感字段。
- renderer：同一模板版本和规范化变量生成相同 rendered digest；非法语法、缺失 / 额外变量、嵌套解释和超限内容失败关闭。
- repository：memory / SQLite / PostgreSQL 同组契约，覆盖 CAS、不可变版本、作用域、排序、重启、migration、rollback、运行角色和 no fallback。
- governance：配置 binding、候选创建 / review、assignment 激活为三个显式动作；任一 authority 漂移均阻塞运行。
- provider：非 invocation 路由零 provider 调用；每次成功 invocation 恰好一次计划内调用；幂等、取消与 outcome unknown 不重放。
- privacy：模板正文只进入 Template owner；运行变量、rendered messages、output、provider raw response、credential、token、header 和 DSN 不进入 run / session / assignment / candidate、日志或 committed evidence。
- compatibility：应用配置 v1 / v2、发布候选 v1 / v2、Application Session v1、run v0–v5、Gateway、Workflow / RAG、API key、Operations 和 Evaluation 继续通过。
- browser：完成从模板创作到受控调用和运行审查的连续链，页面无控制台错误，应用切换和服务重启不恢复 transient 内容。

涉及新 schema、API、持久化、API key capability 和 provider 执行边界，后续实现必须先新增一个对应任务卡，并在每批完成后执行相称验证；批次 D、E 与专题关闭补跑全量仓库门禁。

## 停止线

- 不把模板正文并入 Application Configuration Draft、Publish Candidate、Session、Run History、Gateway Request History 或 Application Operations。
- 不新增第二套应用发布审批状态机；模板版本的人工作业统一进入既有 Application Publish Candidate review。
- 不支持任意模板代码、Jinja / Handlebars 完整语法、函数、循环、条件、include、文件 / 环境变量 / 网络读取或动态工具注册。
- 不实现多轮 durable transcript、长期记忆、自动摘要、agent loop、RAG、HTTP Tool、connector、在线搜索、schedule、background execution、replay / resume、业务写回或自动确认。
- 不自动创建配置 binding、发布候选、review、assignment、release 或正式应用变更。
- 不实现 provider retry / fallback、负载均衡、quota、billing、cost ledger、生产 API key、生产认证、生产 secret 或生产能力声明。
- 不下载模型、不接真实外部 provider 账户、不把本地 fake / mock 调用写成生产可用。

## 下一步

设计已于 2026-07-20 确认。下一步先新增单一实施任务卡，冻结批次 A 至 E 的 schema / API 版本、迁移顺序、兼容矩阵、专项验证和停止条件；随后从批次 A 开始，不并行打开数据库、发布绑定、provider 调用与 Web 全链。

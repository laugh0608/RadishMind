# Workflow RAG 本地知识材料导入、审查与快照构建（开发 / 测试态）v1

更新时间：2026-08-11

状态：`workflow_rag_local_material_import_review_snapshot_building_dev_test_v1_batch_a_domain_completed_pencil_pending`

## 设计结论

本专题补齐 Workflow RAG 用户路径最前端的真实阻塞：现有知识快照后端已经支持严格作用域、完整替换、不可变版本、CAS、三种开发测试态存储和后续评测 / 晋级 / 运行链，但 Web 仍要求用户手写 `camelCase JSON array`。内部应用开发者无法从本地 Markdown / 纯文本材料可靠地产生可审查片段，也无法在提交前确认来源、顺序、重复、预算和最终持久化内容。

v1 在现有 `WorkflowRAGSnapshotPanel` owner 内增加浏览器本地材料暂存、确定性切分、结构化片段审查和显式提交。原始文件、文件名和解析中间态只存在于当前 React 组件内存；只有用户已经审查的最终 fragment replacement 才通过既有 create / version API 写入现有不可变知识快照。

本专题不新增 API、JSON Schema、permission、migration、repository、运行 profile 或业务真相源，不改变 lexical retrieval、dataset review、promotion、binding、assignment、Session 或 Run History。若实现审计发现现有 snapshot 契约不能完整表达最终片段，必须停止并重新评审，而不是在 Web 中隐藏第二套持久化格式。

## 当前实施进度

2026-08-11 已完成批次 A 的纯 TypeScript 导入领域：`workflowRAGLocalMaterialImporter.ts` 复用既有敏感材料规则，实现严格 UTF-8 / BOM / 换行规范化、稳定来源排序与引用、Markdown 围栏和标题边界、纯文本段落、Unicode code point 安全切分、内容 digest、重复和全部预算 findings；生成的 fragment 已通过既有 snapshot write validator。定向 `6/6` 与 Web production build 已通过。

当前 Pen 会话仍被另一项目占用，因此没有读取、覆盖或修改其设计源；本专题的 `B / 局部 Pencil` 仍待在 RadishMind 设计源内冻结，React 结构改造继续遵守“局部稿人工通过前不开始”的停止线。下一步先完成该局部稿，再进入批次 B 的单一结构化 owner。

## 用户价值与当前阻塞

目标用户是需要把内部文档、FAQ 或人工整理说明转为可审查 RAG 快照的应用开发者。

当前路径要求用户：

1. 自行切分材料；
2. 自行生成合法且唯一的 `fragmentRef`、`sourceRef` 与 `pageSlug`；
3. 手工编写完整 JSON replacement；
4. 在缺少分片级表单、重复提示和预算汇总的情况下直接创建不可变版本。

这使已经完成的“快照 → 质量评测 → 晋级 → 应用绑定 → 受控调用”链路在入口处仍偏向协议调试，而不是可持续使用的产品任务。

v1 完成后，用户路径固定为：

1. 在当前 application 的知识快照 owner 中选择 `.md`、`.markdown` 或 `.txt` 文件；
2. 浏览器本地完成 UTF-8 读取、换行规范化、稳定来源标识和确定性切分；
3. 用户查看来源、切分结果、重复、预算与阻塞项；
4. 用户逐条修改标题、来源类型、稳定来源引用、官方标记和正文，或删除不需要的片段；
5. 所有阻塞项清零后，用户确认当前 application、目标 snapshot、版本动作、来源数、fragment 数和正文总量；
6. Web 才把最终 `WorkflowRAGSnapshotWriteInput` 交给既有 create / version consumer；
7. 成功后仍由现有 snapshot record、dataset candidate review、promotion、binding 与 runtime owner 承担后续路径。

## 现有实现审计

本专题直接复用以下已经成立的边界：

| 现有能力 | 当前 owner | 本专题如何复用 |
| --- | --- | --- |
| snapshot create / version / archive | Platform `WorkflowRAGSnapshotRepository` 与既有 HTTP route | 只调用 create / version；不修改 route 或 store |
| fragment contract | `workflow_rag_fragment.v1` | 最终输出继续是既有七字段 `WorkflowRAGFragmentInput` |
| 预算与 secret scan | Go snapshot domain + Web strict consumer | 本地预检与服务端权威校验同时保留，服务端结果为最终结论 |
| 精确 application scope 与 permission | 既有 snapshot consumer | 导入不产生网络请求；提交仍要求 `workflow_rag_snapshots:write` |
| immutable version 与 CAS | snapshot repository | 新版本继续提交完整 replacement 与 `expected_latest_version` |
| 三种开发测试态存储 | memory / SQLite / PostgreSQL 既有 owner | 不新增 selector、表、连接、迁移或 fallback |
| 后续评测与运行 | dataset / promotion / binding / assignment / retrieval owner | 只消费成功写入的精确 snapshot version，不接收本地 staging |

当前 Web 的唯一编辑真相是 `SnapshotEditor.fragmentsJSON`。新实现应把它替换为类型化 staging / fragment editor；不得同时维护 JSON textarea 与结构化表单两套可写状态。读取既有 snapshot record 时，应直接映射到相同 fragment editor，使手工片段、文件导入片段和历史版本使用同一最终审查模型。

## 输入范围与预算

v1 只接受用户显式选择的本地文本文件：

| 项目 | v1 边界 |
| --- | --- |
| 文件类型 | `.md`、`.markdown`、`.txt` |
| 编码 | 严格 UTF-8，可选单个 UTF-8 BOM |
| 单次文件数 | `1..16` |
| 单文件原始大小 | 最大 `256 KiB` |
| 单次原始内容总量 | 最大 `1 MiB` |
| 最终 fragment 数 | 继续沿用 `1..256` |
| 单 fragment 正文 | 继续沿用最大 `8 KiB` UTF-8 |
| 最终正文总量 | 继续沿用最大 `1 MiB` |
| 切分目标 | 最大 `6 KiB` UTF-8，给人工修订保留余量 |

MIME 只作提示，不能替代扩展名和内容校验。文件选择包含目录、软链接、压缩包、二进制、PDF、Word、HTML、图片、音视频或未知扩展名时失败关闭；v1 不使用 File System Access API，不扫描目录，也不读取浏览器未由用户明确选择的文件。

## 浏览器内暂存模型

暂存模型只存在于当前 snapshot panel 的 React state，至少包含：

- 当前 application 与目标 snapshot identity；
- 本次选择的来源顺序、原始 basename、UTF-8 bytes、内容 digest 和解析状态；
- 每个来源的 `sourceType`、稳定 `sourceRef`、`isOfficial` 与生成片段；
- 每个片段的最终七字段输入、UTF-8 bytes、内容 digest 和 findings；
- 总来源数、fragment 数、正文总量、重复集合和阻塞项；
- 当前 staging generation，用于拒绝 application / snapshot 切换后的迟到读取结果。

以下数据不得写入 URL、hash payload、`localStorage`、`sessionStorage`、IndexedDB、日志、committed evidence、Request History、run record 或新的 repository：

- 原始 `File` / `Blob`；
- 原始 basename 和完整文件正文；
- 解析中间块、被删除片段与失败文件内容；
- 用户尚未确认的 fragment replacement。

application、snapshot、lifecycle filter 或 owner 切换时，必须清空文件、解析结果、用户修改、findings 和待提交确认；在途 `File.text()` / `arrayBuffer()` 与 digest 结果只能通过 generation guard 丢弃，不能回填到新作用域。

## 确定性解析与切分

实现采用纯 TypeScript、无网络、无外部解析服务的 `workflow.rag.local-material-sectioner.v1`。相同文件 bytes、相同来源设置和相同算法版本必须产生相同顺序、稳定引用与正文。

### 规范化

1. 使用 fatal UTF-8 decoder；非法编码直接拒绝，不做替换字符修复。
2. 移除开头单个 UTF-8 BOM。
3. 将 `CRLF` / `CR` 规范化为 `LF`，保留正文内部其它字符。
4. 拒绝 NUL、空文件、仅空白文件和既有 secret material 规则命中的内容。
5. 原始 basename 只用于本地 UI；持久 `sourceRef` 使用规范化 ASCII slug 与内容 digest 前缀，不包含本机路径。

### 来源顺序与标识

- 来源按规范化 basename、完整内容 digest、原选择序号稳定排序；同 basename 不依赖浏览器返回顺序。
- 完全相同的内容 digest 视为重复来源并阻塞提交，用户必须显式移除其中一项。
- `sourceRef` 默认为 `local_material/<slug>-<digest-prefix>`，必须通过既有 reference 校验；用户可在审查区修改为其它无 credential 的稳定引用。
- `fragmentRef` 由来源短键与顺序号生成，保持小写短键和 snapshot version 内唯一；人工编辑后仍需通过既有 pattern 与唯一性校验。
- `pageSlug` 由来源短键与 section 顺序生成，不包含本机路径或原始 query。

### Markdown 与纯文本

- Markdown 只把 fenced code block 之外的 ATX heading 识别为 section 边界；heading 前正文归入 `overview`。
- fenced code block 内容和内部 `#` 原样保留，不作为 section 标题。
- 纯文本按一个或多个空行分隔 paragraph block。
- 空 section 不生成 fragment；标题为空时使用来源 basename 的脱敏展示名与序号。
- section 依次打包到 `6 KiB` 目标；超长单块优先在换行或空白处切分，仍无法切分时只在 Unicode code point 边界切开，绝不按 UTF-16 code unit 截断。
- 切分不做语义摘要、LLM 改写、embedding、rerank 或自动删减；正文内容只能由确定性规则切分或用户显式编辑。

算法版本只属于 Web staging 事实，不写入 snapshot schema 或冒充 retrieval profile。最终权威仍是服务端保存的 fragment content、digest 与 snapshot digest。

## 结构化审查与提交

现有 JSON textarea 替换为单一结构化 owner：

1. 来源区：选择文件、显示来源状态、调整来源类型 / stable ref / official 标记、移除来源；
2. 预算与 findings：来源数、fragment 数、正文总量、重复、非法字段、secret 与超限；
3. fragment 列表：顺序、标题、来源、bytes、阻塞状态；
4. 当前 fragment inspector：编辑 `fragmentRef`、`pageSlug`、标题、来源属性与正文；
5. 提交摘要：application、snapshot key / version、create 或 full replacement、最终数量与 CAS 版本。

强选中只属于当前来源或当前 fragment。解析失败、重复、超限、secret、version conflict 与 lifecycle 状态使用独立文字、图标和结构表达，不能用选中颜色代替。

提交前必须重新从当前 staging 构造 `WorkflowRAGSnapshotWriteInput` 并调用既有 `validateWorkflowRAGSnapshotWriteInput`。本地校验成功只表示可以发请求；服务端仍重做全部权威校验。提交开始后，用户编辑、来源变化、application / snapshot 切换会使本次结果失效，迟到响应不得覆盖新 generation。

create / version 成功后：

- 清空原始文件和解析中间态；
- 用服务端返回的 exact record 重建结构化 editor；
- 刷新现有 active / archived collection；
- 不自动进入 dataset review、promotion、binding、activation 或 execution。

version conflict 保留当前本地 staging 供用户比较，但必须显示服务端最新版本；不得自动重放、自动 merge 或改写 expected version。

## 失败语义

本地稳定失败至少包括：

| failure code | 语义 |
| --- | --- |
| `workflow_rag_material_file_count_invalid` | 文件数为空或超过上限 |
| `workflow_rag_material_file_type_unsupported` | 扩展名或内容类型不在允许范围 |
| `workflow_rag_material_file_too_large` | 单文件或原始总量超限 |
| `workflow_rag_material_utf8_invalid` | 严格 UTF-8 解码失败 |
| `workflow_rag_material_content_invalid` | 空内容、NUL 或无法形成有效 section |
| `workflow_rag_material_source_duplicate` | 来源内容 digest 完全重复 |
| `workflow_rag_material_fragment_duplicate` | 最终 fragment 正文或引用重复 |
| `workflow_rag_material_budget_exceeded` | fragment 数、单片段或总正文超限 |
| `workflow_rag_material_staging_changed` | 异步结果不再属于当前 application / snapshot generation |

既有 `workflow_rag_secret_material_forbidden`、`workflow_rag_fragment_invalid`、`workflow_rag_snapshot_version_conflict` 与 store / scope failure 继续由 snapshot consumer / server 表达。本专题不复制或重新解释后端失败码。

## UI 覆盖级别

五维评分为 `1 / 1 / 1 / 1 / 1 = 5`，采用 `B / 局部 Pencil`：

- 结构只改造 S2 Application Development Workspace 内既有知识快照 owner；
- 交互是文件选择、结构化审查和既有 create / version 的组合，不建立新页面族；
- 风险属于可逆的开发测试态暂存和 append-only snapshot version；
- 状态集中在当前 owner，窄屏需调整顺序但不改变产品导航；
- 来源与 fragment 审查模式同时服务 create 与 version 两条路径。

Pencil 只冻结导入区、findings / budget、fragment list、当前 inspector、提交摘要和窄屏顺序，不新建 S11，不重画 S2 完整基准面，也不把代表性文件数或 fragment 数画成实时生产事实。局部稿人工通过前不开始 React 结构改造。

## 实施拆分

### 批次 A：局部设计与纯导入领域

1. 待完成：冻结 `B / 局部 Pencil` 的 Desktop / Narrow 区域与关键 blocked state。
2. 已完成：新增纯 TypeScript material reader / sectioner / staging projector。
3. 已完成：覆盖 UTF-8、BOM、换行、Markdown fence / heading、纯文本 paragraph、Unicode 边界、稳定排序 / ref、重复和全部预算。
4. 不修改后端、API、schema、migration、repository 或 launcher。

### 批次 B：结构化 owner 与既有 snapshot 写入

1. 用单一结构化 staging / fragment editor 替换 JSON textarea。
2. 接入本地文件选择、来源审查、fragment inspector、findings、提交摘要和 generation guard。
3. create / version 继续复用既有 strict consumer；读取历史 record 进入同一 editor。
4. 覆盖 application / snapshot 切换、permission、archived、version conflict、server failure 和迟到结果。

### 批次 C：真实产品链与关闭

1. 使用现有 SQLite local-product 启动入口，不增加新开关。
2. 真实浏览器完成 Markdown + Text 导入、审查、创建 v1、修改后创建 v2、精确 record 重读与后续现有 owner 可见性。
3. 服务重启后只恢复 committed snapshot v1 / v2，不恢复文件、basename、解析中间态或未提交 staging。
4. 覆盖 `1440×900`、`720×900` 与 `390×844`，检查响应式顺序、横向溢出、控制台和浏览器持久化介质。
5. 更新当前焦点、专题状态和周志；Web tests、production build、仓库 fast 通过。只有实际改变协议、schema、阶段边界或高风险执行边界时才补全量门禁。

本专题不创建独立 task card、fixture 或专项 checker；纯 importer tests、snapshot consumer tests、Web build、真实浏览器和现有仓库门禁足以承载证据。若实现被迫新增 API、schema、持久 staging、外部 parser 或执行边界，必须先新建高风险任务卡并重新评审。

## 完成定义

- 用户不再需要手写 fragments JSON 才能从 Markdown / Text 创建知识快照。
- 相同输入和设置产生相同 fragment 顺序、引用与正文。
- 所有最终 fragment 均可见、可编辑、可删除，并在提交前通过完整预算 / secret / duplicate 审查。
- 原始文件与未提交 staging 只存在于当前组件内存，作用域切换、刷新和服务重启后不可恢复。
- create / version 仍由现有 scope、CAS、immutable repository 与三种 store owner 保证。
- 成功写入的精确 snapshot version 可被既有 dataset review、promotion、binding 和 runtime 路径继续消费。
- 没有新增 API、schema、migration、repository、permission、自动执行或生产声明。

## 停止线

- 不支持 PDF、Word、HTML、图片 OCR、音视频、压缩包、目录扫描或浏览器 File System Access API。
- 不接 crawler、URL 抓取、connector、在线搜索、embedding、vector database、reranker 或外部 parser 服务。
- 不把文件名、本机路径、原始文件、失败内容或未提交 staging 持久化或上传。
- 不自动创建 snapshot、自动 version、自动 dataset review、自动 promotion、自动 binding、自动 activation 或自动 execution。
- 不做近似重复、语义摘要、LLM 切分、自动改写、自动分类或自动 official 判定。
- 不创建第二套 fragment schema、snapshot store、staging repository、历史 owner 或跨应用复制。
- 不扩 production secret、正式 membership / OIDC、production RAG、quota / billing、schedule、replay / resume、agent loop 或业务写回。

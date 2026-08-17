# 应用结果资产库与受控导出（开发 / 测试态）v1

更新时间：2026-08-17

状态：`application_result_artifact_library_controlled_export_dev_test_v1_completed`

## 功能目标

让内部开发者在选中一个 Application 后，跨 Session 找到自己显式保存的运行结果，核对 Session / Turn / Run 来源、生命周期和内容摘要，并把单个精确结果显式导出为可校验的本地 JSON 文件。

本专题不创建第二套结果真相源。`application_result_artifact.v1`、`application_result_artifact_lifecycle.v1` 与 append-only lifecycle event 继续是唯一 owner；新能力只增加 application-scoped 严格读取、canonical export projection 和一个复用既有 Workbench 模式的 Result Workspace。

## 为什么是当前下一顺位

结果资产显式保存专题已经证明五类 Session Profile、三种 store、精确正文读取与归档生命周期，但资产仍只能从单个 Session 内逐页发现。用户离开原 Session 后缺少 application 级的结果工作区，已保存结果无法形成稳定的“发现 → 核对 → 取用”产品闭环。

与其它候选相比，本专题：

- 不依赖真实 Provider、Radish OIDC、production secret 或上层项目挂载点；
- 不持久化 transcript、用户输入、prompt、Provider raw response 或长期记忆；
- 能复用已经验证的不可变 content、digest、lineage 与 lifecycle owner；
- 为未来 RadishFlow / Radish 的人工 handoff 保留可校验结果包，但不直接写入其业务真相源。

工作区运营收件箱批次 B 缺少跨 owner 统一 cursor 前置，真实 Provider / OIDC / Image Backend 仍依赖外部资源，因此不作为本轮默认实现顺位。

## 目标用户与核心流程

目标用户是开发 / 测试环境中拥有当前 workspace 与 application 读取权限的内部开发者。

1. 用户在既有 Application Context 中打开 Result Workspace。
2. 页面按当前 application、owner 和 `active | archived` 生命周期读取结果资产，不先枚举 Session，也不在前端拼接多条 Session 分页。
3. 用户可按 execution profile、content type 和生命周期过滤，cursor 必须绑定完整 scope、过滤条件与 limit。
4. 列表只返回 `application_result_artifact_summary.v2`，不返回正文；用户选择一项后才按 summary 中的精确 Session / artifact identity 读取正文与当前 lifecycle。
5. 页面展示 Session / Turn / Run lineage，并复用既有 Run History handoff；archive / unarchive 继续调用既有 CAS 路由。
6. 用户显式选择“导出 JSON”后，服务端以专用 export 权限重读同一 artifact 与 lifecycle，形成 `application_result_artifact_export.v1`。浏览器只把本次响应保存为本地文件，不缓存、不上传、不创建公开链接。
7. application、workspace、owner 或身份切换必须取消旧请求、清空正文与 export 文档，并拒绝迟到响应回填。

## 单一 owner 与数据边界

| 数据 | 唯一 owner | 本专题允许 | 本专题禁止 |
| --- | --- | --- | --- |
| artifact content / digest / lineage | 既有 Application Result Artifact repository | 精确读取、application-scoped metadata list、canonical export projection | 复制 content 表、改写 content、通用 result store |
| lifecycle current state / event | 既有 lifecycle repository | active / archived 过滤、既有 archive / unarchive CAS | delete、overwrite、第二套 lifecycle |
| Session / Turn / Run | 既有 Session / Run owner | 展示精确 reference、交接既有详情 | 重建 transcript、保存 input、replay / resume |
| export | 无持久 owner；按请求即时投影 | 单个 artifact JSON、请求与 audit lineage、本地下载 | export history、公开 URL、分享令牌、后台打包 |

application-scoped list 不能通过客户端循环所有 Session 得到。repository 必须在同一 application / owner scope 内完成过滤与稳定排序；数据库新增的索引只服务读取，不改变 artifact 或 lifecycle payload。

## 合同与 API

### `application_result_artifact_summary.v2`

`contracts/application-result-artifact-summary.schema.json` 必须与当前 Go / TypeScript 严格消费者一致，包含：

- 原有 metadata-only artifact identity、Session / Turn / Run lineage、content type / bytes / digest 与 created at；
- `lifecycle_state`、`lifecycle_version`、可空 `archived_at` 与 `lifecycle_updated_at`；
- 不包含 `content`、输入、prompt、Provider raw response、credential 或 header。

### `application_result_artifact_export.v1`

export 文档固定包含：

- export schema version 与本次 `exported_at / exported_by_actor_ref / request_id / audit_ref`；
- 完整 `application_result_artifact.v1`；
- 与读取时一致的 `application_result_artifact_lifecycle.v1`；
- `export_digest`，绑定 artifact identity、content digest、lifecycle state / version 和导出元数据。

服务端必须在构造后重新校验 artifact、lifecycle、scope、digest 和 export 文档。export 不落库，不改变 lifecycle，不产生 Session / Run 副作用。

### 路由

- `GET /v1/user-workspace/applications/{application_id}/result-artifacts`
- `GET /v1/user-workspace/applications/{application_id}/result-artifacts/{artifact_id}/export`

application list query 只允许：

- `workspace_id`
- `lifecycle_state=active|archived`
- `execution_profile`
- `content_type=text/markdown|application/json`
- `limit`
- `cursor`

export query 只允许 `workspace_id`。两个路由均默认关闭，复用 Application Session 开发测试态 gate、verified workspace membership 和 `Cache-Control: no-store`。

权限固定为：

- list：`application_sessions:read`
- export：`application_sessions:read + application_result_artifacts:export`

`application_result_artifacts:archive` 仍只用于既有 archive / unarchive mutation，不与 export 合并。

## Cursor、排序与过滤

- 排序固定为 `created_at DESC, artifact_id DESC`。
- application cursor 使用新版本，绑定 tenant / workspace / application / owner、scope kind、lifecycle、execution profile、content type、limit 和最后一项 identity。
- cursor 与任一过滤条件不一致时返回 `application_result_artifact_payload_invalid`，不隐式重置第一页。
- 默认 lifecycle 为 `active`，默认 limit 为 50，最大 100。
- 空 execution profile / content type 表示不过滤；非空值必须来自现有五 profile / 两 content type allowlist。
- 数据库查询和索引必须保持同一 owner scope；不得先取全租户数据再在应用层过滤。

## Result Workspace 与设计覆盖

页面继续位于既有 Application Workbench，不建立新的一级产品面或 S11。它复用：

- Application Context 与现有 application 切换语义；
- Saved Draft Library 的“列表 + inspector”层级；
- Session Result Artifact Panel 的 summary、exact read、lifecycle 与 Run handoff；
- Family UI Workbench token、职责圆角、窄屏渐进顺序和失败表达。

批次 B 复核后的五维评分为 `0 / 0 / 1 / 1 / 0 = 2`，覆盖级别为 `C / 直接实现`。Application Context、S5 单 owner / Run handoff、Saved Draft Library 的筛选列表、Session Result Artifact Panel 的 exact inspector / lifecycle，以及既有窄屏单列顺序已经冻结了本批全部结构和交互；新增风险只是在开发测试态显式下载前表达 digest 重校验，未产生新的页面拓扑、高风险确认或跨页面复用语义，因此不修改 Pencil、不建立 S11，以真实浏览器证据验收。

## 实施批次

### 批次 A：严格合同、application-scoped repository 与 API

状态：已完成。

- 对齐 summary v2 schema，新增 export v1 schema 与 Go validator。
- 扩展同一 repository 的 application-scoped list；session list 保持兼容。
- 为 SQLite / PostgreSQL 增加 application lifecycle history 索引与 migration marker。
- 注册 list / export 路由、严格 query、独立 export permission、稳定 failure mapping 和 no-store。
- 覆盖 memory、SQLite、PostgreSQL 的 scope、filter、cursor、corruption、migration、no-fallback 和 export 零副作用。

完成证据：

- `application_result_artifact_summary.v2` schema 已与 lifecycle-aware Go 投影对齐；`application_result_artifact_export.v1`、canonical digest 与构造后重校验已落地。
- application-scoped repository list 在同一 tenant / workspace / application / owner 内完成 lifecycle、execution profile 与 content type 过滤；cursor v3 绑定完整 scope、过滤条件与 limit，并严格拒绝未知字段、尾随 JSON 与条件漂移。
- SQLite `0024` 与 PostgreSQL `0027` 只顺序追加 application history 读取索引；既有 artifact / lifecycle 表、payload 与不可变约束未改写。
- list / export HTTP route、独立 `application_result_artifacts:export` 权限、strict query cardinality、`no-store`、跨 application 隔离和不存在的永久 `DELETE` route 已有单元证据。
- memory 覆盖超过 `100` 条同时间戳稳定分页；SQLite 覆盖跨 Session filter、export 与服务重启；PostgreSQL 17 覆盖 migration / rollback / reapply、运行角色、application list / export、configured Server 关闭 no-fallback 与重启恢复。完整 PostgreSQL 聚合集成和 Platform `internal/httpapi` 包测试均已通过，测试容器已关闭。

### 批次 B：Result Workspace strict consumer

状态：已完成。

- 增加 application-scoped strict TypeScript consumer 和单一 React owner。
- 接入 filter、分页、exact read、Run handoff、archive / unarchive 与显式本地 JSON 下载。
- application / workspace / identity 切换时 abort、清空并拒绝迟到响应。
- 按 `C / 直接实现` 复用既有基准面，并完成 production build 与浏览器多视口验收。

完成证据：

- application-scoped consumer 严格校验 envelope、跨 Session summary、完整 filter、scope、owner 与 schema；offline 保持零请求，application / filter / cursor / session / artifact / generation 任一漂移都会拒绝迟到响应。
- export consumer 只使用独立 export permission，按 Go 结构顺序重建 canonical 文档，通过 Web Crypto 复核 `content_digest` 与 `export_digest`；下载前再次比对列表、精确正文和当前 lifecycle，文件名只含稳定 artifact identity 与 lifecycle version。
- S5 Runtime Review 增加第四个 `Saved results` task，任一时刻只挂载一个 owner；列表只含 metadata，正文按 summary 的精确 Session / artifact 读取，Run 继续交接既有 S6 owner，archive / unarchive 继续调用既有 CAS 路由。
- Web `375/375` 与 production build 通过；应用内浏览器覆盖 `1440×900`、`720×900`、`390×844`，三种宽度均无横向溢出且只有一个选中任务。active / archived 与 content type 筛选切换、清除和 application-scoped route 命中正常，控制台零 warning / error。
- 当前 Agent 本地产品数据缺少可用 assignment，创建验证数据按既有合同返回 `application_session_authority_changed`，没有绕过 authority 或伪造 artifact；同一页面的 exact read / export / lifecycle / 重启连续链留在批次 C 使用稳定双 Session fixture 验收。

### 批次 C：双数据库产品连续链与专题收口

状态：已完成；专题关闭。

- memory / SQLite / PostgreSQL 复用同一 application library fixture。
- SQLite 页面完成跨至少两个 Session 的 application list、过滤、精确读取、导出、归档往返和服务重启恢复。
- PostgreSQL configured Server 完成相同 list / export 核心链、关闭连接 no-fallback 与重启恢复。
- 验证 export 文档不落库、永久 `DELETE` / public share route 不存在，最终同步真相源并关闭专题。

完成证据：

- 显式开发测试 fixture 以同一 application、两个 Session、Workflow / Prompt 两种 profile、Markdown / JSON 两种 content type 和 active / archived 两种初始状态进入 memory、SQLite 与 PostgreSQL；重复启动只补缺失记录，不重置已推进的 lifecycle。fixture 只提供稳定来源引用，不创建第二套 Session / Run owner，也不声明真实 Provider 执行成立。
- 正式本地产品启动入口增加 `--application-result-artifact-library-local-product`，只在完整 Application Session、目录读写、开发身份与 `sqlite_dev` 前置同时成立时启用 fixture；缺少任一前置都在 Server 构造阶段失败关闭。
- SQLite 真实页面完成跨 Session 默认列表、profile / content type / lifecycle 组合过滤、精确正文读取、archive / unarchive CAS、服务重启恢复和二阶段本地下载。重启后同一 Markdown artifact 保持 content digest 不变，lifecycle 恢复为 `active v3`；下载文件再次解析后 content / export digest 均匹配。
- 浏览器 `1440×900`、`720×900`、`390×844` 均无页面级横向溢出，窄屏筛选与 inspector 保持单列顺序，控制台零 warning / error。下载只在严格 consumer 完成 scope、artifact、lifecycle 与 digest 复核后生成一个易失 Blob URL；application、filter、selection、重新准备或卸载都会撤销该 URL。
- PostgreSQL 17 配置化 Server 复用同一 fixture，完成 filter、exact read、export、archive、关闭后 `store_unavailable` 且不回退 memory、重启恢复 `archived v2` 和 unarchive 至 `active v3`；聚合集成测试通过，容器与网络已关闭。
- export 没有数据库表、repository 或历史记录；既有永久 `DELETE` 与 public share route 继续不存在。专题没有打开 transcript、批量导出、公开分享、业务写回、真实 Provider 或 production 能力。

## 验收方式

- contract：JSON Schema、Go validator、TypeScript strict parser 对同一正负 fixture 一致。
- scope：跨 tenant / workspace / application / owner 均失败关闭，不泄漏资源存在性。
- pagination：超过 100 条、同时间戳 tie-break、过滤绑定 cursor 和非法 cursor 均可复验。
- persistence：migration / rollback / reapply、运行角色、索引、重启和 no fallback。
- privacy：list metadata-only；export 只含既有 artifact 内容和 allowlist 元数据，不进入日志、URL、浏览器持久化或 committed fixture。
- UI：至少 `1440×900`、`720×900`、`390×844`，无横向溢出，控制台无 warning / error。
- repository：精准 Go / Web tests、production build、PostgreSQL integration、仓库 fast；触及合同、migration 与阶段真相源时补全量仓库检查。

## 停止线

- 不实现 transcript、长期记忆、自动摘要、全文内容搜索或跨 artifact 聚合生成。
- 不实现永久 purge、自动 retention、覆盖写、批量导出或后台 export job。
- 不实现 public / signed URL、分享令牌、邮件 / Slack 发送或外部 object storage。
- 不执行 candidate action，不写 RadishFlow / Radish 业务真相源，不实现 replay / resume / writeback。
- 不打开 production repository、production auth、production secret、真实 Provider、quota / billing 或公开生产 API。

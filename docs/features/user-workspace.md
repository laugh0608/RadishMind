# 用户工作区设计与开发文档

更新时间：2026-07-19

## 功能定位

用户工作区是 RadishMind 面向终端用户和项目成员的主工作区，用于查看和管理 AI 应用、提示词应用、工作流、Agent / Copilot 应用、API 密钥、调用量、运行记录、成本摘要和人工审查入口。

## 当前状态

- `apps/radishmind-web/` 已有只读产品壳、工作区首页、应用列表、API 密钥、用量配额、工作流定义、运行历史、工作流审查区和审查交接区。
- [应用 API 接入与调用 v1](user-workspace/application-api-integration-invocation-v1.md) 已把当前选中应用、`/v1/models` 模型目录、三协议 × 三语言接入示例、现有 Gateway 调试台和脱敏请求历史串成连续的内部开发者路径；作用域不再依赖固定应用配置。
- [应用配置草案与审查 v1](user-workspace/application-configuration-draft-review-v1.md) 已为当前应用建立独立脱敏配置草案，完成模型 / 协议校验、memory / SQLite / PostgreSQL 开发测试态保存、恢复、配置比较、CAS 冲突审查，以及到 API 接入区和调试台的交接；正式应用真相源仍只读。
- [应用发布治理与晋级审查 v1](user-workspace/application-publish-governance-promotion-v1.md) 已把有效的已保存草案固定为不可变候选版本，完成服务端重读、摘要计算、审查 CAS、漂移 / 被取代检查、阻塞式晋级资格判断，以及到接入区、调试台和请求历史的交接；`approved` 仍不修改正式应用。
- [应用目录与生命周期（开发/测试态）v1](user-workspace/application-catalog-lifecycle-dev-test-v1.md) 已完成核心生命周期、memory / SQLite / PostgreSQL 开发测试持久化、Web 管理和真实浏览器纵向验收：应用唯一真相源、服务端标识、所有者作用域、完整元数据更新、软归档、原子 CAS、独立迁移、无回退、归档只读历史和重启恢复均已成立。
- [API 密钥生命周期与 Gateway 开发测试态认证 v1](user-workspace/api-key-lifecycle-gateway-dev-test-auth-v1.md) 已完成并关闭：活跃应用可以签发有期限、有受控作用域、只展示一次且可吊销的开发测试态密钥，五条 northbound 路由可显式启用 API 密钥认证，并记录可信调用上下文、脱敏请求历史与最近使用时间；七组件聚合 SQLite 本地产品链、真实 PostgreSQL 专项门禁、Web 一次性交接、真实浏览器连续验收、敏感信息扫描和重启恢复均已通过。
- [应用运行观测与用量归因 v1](user-workspace/application-operations-observability-usage-attribution-v1.md) 批次 A 已完成：当前应用可并列审查 Gateway 请求与 Workflow 运行的首分页窗口，分别查看状态、usage availability、受控调用计数、来源覆盖和合并时间线；两类记录不自动关联，当前窗口不冒充全量 usage、成本、配额或计费。
- 工作区首页和工作流定义已支持创建本地工作流草案并进入草案设计器；草案保存复用仅开发的已保存草案消费端，不代表生产持久化已成立。
- `User Workspace Saved Draft List v1` 已在工作区首页支持仅开发的已保存草案列表：显示当前应用下已保存草案的脱敏摘要、空结果 / 失败状态、刷新和恢复。默认内存、聚合 SQLite 与显式 PostgreSQL 开发测试态存储库均可承载该路径，但不代表生产持久化已成立。
- 草案设计器已支持本地节点新增、移动、删除保护、属性编辑和边重建；校验检查器、执行计划预览和运行时准入检查器使用当前活跃草案，不代表工作流可正式发布或执行。
- 工作流审查交接已把恢复后的活跃草案校验、执行计划和运行时准入结果汇总为可交接审查记录，仍不保存、不导出、不发送交接内容。
- 已保存草案与运行历史具备 memory / SQLite / PostgreSQL 开发测试态存储库；受控执行器 v0、失败审查、运行比较、评测用例 / 版本管理和评测套件 / 发布审查已接入工作区运行历史。
- 应用配置草案、发布候选和显式启用的应用目录均具备 memory、SQLite 与 PostgreSQL 开发测试态存储库；应用目录未启用时，历史只读列表仍来自预置假数据存储库。
- 当前仍不具备生产认证 / 存储库、Radish 工作区成员关系、正式应用生命周期 / 晋级、生产 API 密钥、配额执行、计费、工具调用、确认提交、业务写回或重放。

历史已保存草案准入专题继续作为证据索引保留，不再在本入口重复展开。需要追溯时，从 [工作流专题入口](workflow/README.md) 和对应实现专题进入。

## 设计边界

- 用户端默认只输出建议、解释、审查包和候选动作，不直接写业务真相源。
- 高风险动作必须保留 `requires_confirmation`。
- 只读侧与后续写入 / 执行侧必须分开设计；界面可展示不等于执行能力已就绪。
- API 密钥列表、详情和后续读取不得展示密钥值、摘要、`Authorization` 请求头或任何敏感材料；创建成功后由独立响应完成的一次性交接是唯一例外，且不得进入浏览器持久化介质。

## 下一批开发方向

1. 本地 SQLite、应用目录和 API 密钥纵向专题均已完成并关闭；不继续扩同层应用目录、密钥页面、准入文档、检查器或证据链。
2. 应用运行观测首批已完成；只有跨全部分页窗口的稳定统计、可信 reported usage 或正式 quota / billing owner 成立时，才评审服务端 summary。下一次产品推进继续先选择真实用户缺口并更新功能设计，不预设生产认证、成员关系、配额或计费为下一项。
3. 一次性令牌继续只保存在当前 Web 组件内存；刷新、路由离开、应用 / 身份切换、组件卸载和服务重启都不得恢复原始令牌。
4. 不把开发测试态应用目录或 API 密钥解释为生产存储库与生产授权；OIDC 模式在成员关系契约未成立时继续失败关闭。
5. 后续专题不得隐式打开生产认证、工作区成员关系适配器、正式晋级、生产 API 密钥、配额、计费、模型服务凭据或新的 Gateway 请求 / 响应 schema。

## 验收方式

- 功能展示类：`npm run build`、必要浏览器布局检查、`./scripts/check-repo.sh --fast`。
- 只读契约类：消费端冒烟验证、Go 处理器测试和只读侧契约检查。
- 写入或执行类：先补设计文档和任务卡，再补单元测试、负向测试、仓库级检查和人工确认路径。

<!--
历史检查器兼容字面量，仅供既有证据链读取，人工默认不读：
repository contract preconditions
saved draft list
durable store 迁移前置设计
owner / workspace
Workflow Draft Node Attribute Editing Model v1
Workflow Review Handoff Active Draft v1
Workspace Home / workflow definitions
创建本地 workflow 草案
dev-only saved draft consumer
User Workspace Saved Draft List v1
sanitized summary
-->

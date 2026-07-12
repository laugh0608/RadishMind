# 文档语言治理 v1

更新时间：2026-07-04

状态锚点：`doc_language_governance_topic_v1_defined`

## 文档目的

本文档固定 RadishMind 文档正文的中文优先口径、必要英文保留边界、工程短语中文化顺序和验收方式。它服务协作体验和长期可维护性，不替代产品面、平台专题、任务卡或运行时实现边界。

## 当前结论

- 新增和触碰的文档正文默认使用中文；代码标识符、路径、命令、配置键、协议字段、状态锚点、`schema` / `fixture` / `checker` ID、外部产品名和必要引用保留原文。
- 标题、表格列名、结论、下一步、阻塞项和停止线优先中文表达；必要英文标识符使用反引号，或先写中文概念再附原文标识符。
- 历史文档不做全仓机械翻译；优先治理入口文档、功能 / 平台专题、契约、任务卡和周志中影响阅读的英文工程短语。
- 不重命名状态锚点、fixture key、路径或机器检查依赖的 literal；确需改名时，必须同步更新 checker、fixture、文档入口和验证记录。

## 可保留英文

| 类别 | 保留方式 | 示例 |
| --- | --- | --- |
| 代码和配置标识符 | 原文保留，优先反引号 | `workflow_saved_draft_store`、`RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE` |
| 路径和命令 | 原文保留，优先反引号 | `docs/features/README.md`、`./scripts/check-repo.sh --fast` |
| 协议和机器检查字段 | 原文保留，避免随意翻译 | `schema_version`、`status`、`required_policy_anchors` |
| 状态锚点和 fixture / checker ID | 原文保留，不改写历史锚点 | `audit_store_storage_adapter_concrete_database_selection_review_defined` |
| 外部项目和产品名 | 原文保留，必要时补中文说明 | `RadishFlow`、`Radish`、`OpenAI` |
| 尚无稳定中文对应的专业名词 | 先给中文解释，再保留原文 | `OIDC`、`DSN`、`API` |

## 优先中文化短语

| 原文标识 | 推荐中文 | 使用边界 |
| --- | --- | --- |
| `readiness` | 准入边界 | 状态锚点和文件名保留原文，正文说明用“准入边界”或“准备状态” |
| `entry_review` | 入口评审 | 评审类专题正文优先写“入口评审” |
| `runtime` | 运行时 | 普通正文优先写“运行时”；路径和状态锚点保留原文 |
| `blocker_matrix` | 阻塞矩阵 | 聚合状态说明中写“阻塞矩阵” |
| `evidence_rollup` | 证据汇总 | 聚合事实说明中写“证据汇总” |
| `artifact_guard` | 产物守卫 | 停止线和检查规则中写“产物守卫” |
| `no_fallback` | 不回退 | 禁止隐式兜底时写“不回退” |
| `no_side_effects` | 无副作用 | 静态检查和只读边界中写“无副作用” |
| `smoke` | 冒烟验证 | 验证策略正文写“冒烟验证”；脚本名和 fixture ID 保留原文 |
| `public_production_api` | 公开生产 API | 停止线正文写“公开生产 API” |
| `surface` | 界面 | 页面分层正文写“页面 / 界面”；历史专题名保留原文 |
| `backend` | 后端 | 普通正文写“后端”；`secret backend` 等既有专题名可保留原文 |
| `future` | 后续 | 自然语言中写“后续”；状态锚点和文件名保留原文 |
| `metadata-only` | 仅元数据 | 决策边界正文写“仅元数据”；状态锚点和文件名保留原文 |
| `fail-closed` | 失败关闭 | 失败语义正文写“失败关闭”；fixture ID 保留原文 |
| `dev-only` | 仅开发 | 环境边界正文写“仅开发”；配置键和专题名保留原文 |

## 第二批治理边界

1. 新增或触碰文档时，优先把自然语言句子改为中文，不把多个英文工程词串成正文。
2. 入口文档只放短口径和索引；术语清单、替换顺序和停止线放在本文档与 `doc-language-policy-v1` fixture。
3. 修改状态锚点、fixture key、路径、脚本输出 literal 或 checker 依赖文本前，必须先确认是否会破坏现有检查链路。
4. 工程短语中文化按入口文档、功能 / 平台专题、契约、任务卡、周志顺序推进；每批只处理与当次任务相关的文本。
5. checker 只固定治理边界和入口引用，不扫描全仓英文词，也不把历史英文文档写成失败。

## 优先入口治理记录 v1

状态锚点：`doc_language_priority_documents_remediation_v1_defined`

本批优先处理常读入口中的自然语言英文工程短语，处理范围包括 `docs/radishmind-current-focus.md`、`docs/features/README.md`、`docs/platform/README.md`、`scripts/README.md` 和 `docs/radishmind-code-standards.md`。处理原则是替换正文说明中的 `surface`、`backend`、`readiness`、`entry review`、`metadata-only`、`smoke`、`literal`、`future`、`fail-closed`、`dev-only` 等高频短语；状态锚点、文件名、fixture key、checker ID、命令、路径和协议字段继续保留原文。

本批验收只证明优先入口已完成一轮定向中文化，不表示全仓历史文档已完成语言治理，也不要求重写长证据链、任务卡、脚本输出或机器检查依赖字面量。

## 停止线

- 不做全仓机械翻译。
- 不重命名状态锚点、fixture key、checker ID、路径或机器检查依赖 literal。
- 不把脚本、代码、协议字段和配置键翻译成中文。
- 不因文档语言治理创建 runtime、provider、database、API、repository mode 或产品功能。
- 不把普通文案整理升级为新的 task card / fixture / checker，除非它改变协议、执行边界、生产声明、schema、外部 provider 风险或高风险能力。

## 验收方式

- 运行 `./scripts/run-python.sh scripts/check-doc-language-policy-v1.py`，确认本文档、入口文档、fixture 和 checker 注册一致。
- 提交前运行 `git diff --check` 和 `./scripts/check-repo.sh --fast`。
- 若后续新增术语、保留英文类别或停止线，必须同步更新 `scripts/checks/fixtures/doc-language-policy-v1.json` 与 `scripts/check-doc-language-policy-v1.py`。

## 下一批建议

- 优先在新增文档和当次触碰文档中落实本口径。
- 对历史 storage adapter、audit store 和 workflow 专题只做与当前任务相关的局部中文化，不为了语言治理单独重写长证据链。
- 若发现高频英文工程短语影响阅读，先补入本文档术语表，再做定向替换和验证。

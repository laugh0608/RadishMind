# Workflow Definition 结构化运行输入（开发 / 测试态）v1

更新时间：2026-08-10

状态：`workflow_definition_structured_runtime_inputs_dev_test_v1_batch_c_pencil_frozen_implementation_next`

## 功能定位

本专题把 Workflow Definition 从单一 `input_text` 扩展为由不可变输入合同驱动的扁平强类型表单，使内部开发者能够在直接运行、Application Interaction Session 和 Application Evaluation Campaign 中提交同一种可校验输入，并让 Run、Comparison、Evaluation Case 与 Suite 保持可复验身份。

当前链路虽然已经完成 Saved Draft、不可变 Definition、受控运行、Session、Comparison、Evaluation 与 Campaign，但 `saved_workflow_draft.v1` 的 input contract 只保存字段名，`workflow_definition_executor_v1` 只接收一段 `input_text`。客户端无法根据不可变 Definition 生成可靠表单，服务端也无法证明直接运行、Session 与评测 fixture 使用同一输入结构。

本专题采用显式版本升级，不在 `additional_fields` 或旧 `input_text` 中暗塞结构化语义。现有 v1/v5 资源继续可读、可运行和可比较；新资源通过 v2 Definition、v2 executor、Run v8 与 Comparison v7 形成独立兼容域。

## 用户任务

1. 在 Draft Designer 中为 Workflow 定义一组有序、扁平、强类型的运行输入字段。
2. 保存和晋级时看到输入合同的字段、类型、必填性、说明与稳定 digest。
3. 在直接运行或 Application Interaction Session 中，由 exact active Definition 合同生成表单并提交输入。
4. 输入缺失、多余、类型错误、超出预算、疑似 secret 或合同漂移时，在任何执行副作用前失败关闭。
5. 在 Run History 中审查输入合同摘要、已提供字段名、类型、字节数与 digest，但不回显原始输入值。
6. 使用同一不可变评测计划反复执行结构化 fixture，并把兼容 Run 交给 Comparison、Evaluation Case 与 Suite。

## 版本与兼容域

| 领域 | 现有合同 | 新合同 | 兼容规则 |
| --- | --- | --- | --- |
| Saved Draft | `saved_workflow_draft.v1` | `saved_workflow_draft.v2` | v1 与 v2 独立保存；不自动改写历史草案 |
| Release Candidate | `workflow_definition_release_candidate.v1` | `workflow_definition_release_candidate.v2` | candidate 必须携带与源 Draft 完全一致的 schema version 与 input contract digest |
| Definition Version | `workflow_definition_version.v1` | `workflow_definition_version.v2` | v2 snapshot 只绑定 v2 contract 与 v2 executor |
| Executor | `workflow_definition_executor_v1` | `workflow_definition_executor_v2` | v1 只接受 `input_text`；v2 只接受 `inputs` |
| Run | `workflow_run_record.v5` | `workflow_run_record.v8` | v8 保存结构化输入 metadata，不保存值 |
| Comparison | `workflow_run_comparison.v4` | `workflow_run_comparison.v7` | v8 只与 v8 比较，且 contract digest 必须相同 |
| Evaluation Profile | `workflow_definition_executor.v1` | `workflow_definition_executor.v2` | Case、Baseline 与 Suite 保持 exact profile |
| Application Session | v1–v3 | v4 | v4 authority 与 turn 绑定 exact v2 Definition contract |
| Evaluation Plan / Campaign | v1 | v2 | v2 plan version 可显式保存结构化测试 fixture |

历史 v1 Definition 不因服务升级变成 v2。Draft 只有经过用户显式保存为 `saved_workflow_draft.v2`、重新校验、生成 candidate、晋级和 activation 后，才进入 v2 运行链。服务不得把 v1 `input_text` 猜测、解析或迁移为字段集合。

## 输入合同

`saved_workflow_draft.v2` 的 input contract 固定包含：

- `contract_id`：稳定标识，不承载 display label；
- `fields`：最多 `16` 个有序字段；
- `summary`：面向开发者的简短用途说明；
- 由服务端 canonicalization 生成的 `contract_digest`。

每个字段固定包含：

- `name`：匹配 `^[a-z][a-z0-9_]{0,63}$`，同一合同内唯一；
- `value_type`：只允许 `string | integer | number | boolean`；
- `required`：显式布尔值；
- `label`：用户可见短标签；
- `description`：可为空的用途说明。

首版不支持默认值、nullable、enum、正则、对象、数组、联合类型、动态 schema、条件显隐、字段间表达式、文件、图片、外部 URI、credential 或 secret 类型。字段顺序属于不可变合同，但 canonical digest 按稳定结构计算，不能依赖 Go map 或客户端对象顺序。

## 输入值与预算

v2 运行请求使用：

```json
{
  "inputs": {
    "customer_name": "示例用户",
    "retry_count": 2,
    "dry_run": true
  },
  "condition_values": {},
  "model": "optional-model",
  "temperature": 0.2
}
```

- `inputs` 必须是 JSON object，字段集合必须与 exact active contract 匹配；未知字段拒绝。
- 必填字段必须存在；可选字段通过缺席表达，不接受 `null`。
- `integer` 必须位于 JSON 安全整数范围；`number` 必须有限，不接受 NaN 或 infinity 的等价表示。
- 单个 string UTF-8 正文最多 `4096` bytes；canonical `inputs` 总体最多 `8192` bytes。
- 全部 string 值、字段名、label 与 description 进入现有 secret material 规则检查；命中 credential、token、Authorization、cookie、DSN、私钥或明显 endpoint secret 时拒绝。
- v2 请求出现 `input_text`、v1 请求出现 `inputs`，或同时出现两者时均失败关闭；不得静默丢弃字段或降级到另一 executor。

`condition_values` 继续作为 condition 节点的显式决策输入，不并入通用字段合同。model / temperature 继续服从既有开发测试态受控覆盖合同，不成为 input contract 的业务字段。

## Canonicalization 与隐私

服务端在执行前根据 exact Definition contract 完成类型校验，再生成字段名排序、稳定 JSON number 编码、无多余空白的 canonical payload。`input_digest` 只对这份 canonical payload 计算。

Run v8 只保存：

- `input_contract_id`、`input_contract_digest`；
- 按字段名排序的已提供字段名与声明类型；
- canonical payload 总字节数与 `input_digest`；
- 既有 condition node ids、authority、usage、failure 与 side-effect metadata。

直接运行与 Session 不得把原始输入值写入 Run、Session turn、audit summary、日志、metric label 或 HTTP failure summary。Application Evaluation Plan v2 是唯一例外：用户显式提交的合成 / 测试 fixture 可以在严格 secret 检查后进入不可变 plan version；Campaign item 与 Run 仍不复制 fixture 正文，也不得从历史 Run 反推输入。

## 执行与失败语义

请求处理顺序固定为：

1. 校验 scope、权限、environment 与 application lifecycle；
2. 读取 exact active Definition pointer、version、digest、executor profile 与 input contract；
3. 校验请求 union、字段集合、类型、预算与 secret material；
4. 生成 canonical payload、contract digest checkpoint 与 input digest；
5. 创建 Run 并进入既有 Prompt / LLM / condition / output executor；
6. provider 前继续执行现有 authority checkpoint、quota admission 与 route selection。

在步骤 5 前至少区分稳定 failure code：

- `workflow_input_schema_unsupported`；
- `workflow_input_contract_mismatch`；
- `workflow_input_required_field_missing`；
- `workflow_input_unknown_field`；
- `workflow_input_value_type_invalid`；
- `workflow_input_budget_exceeded`；
- `workflow_input_secret_material_forbidden`；
- `workflow_input_authority_drift`。

这些失败不创建成功 Run、不调用 provider、不写 application session turn，也不消费 quota。若 durable Run owner 需要记录拒绝证据，只能保存 sanitized metadata，不能保存被拒绝的值。

Prompt 根输入使用有界 canonical structured packet，并明确区分字段名、声明类型和值；节点图仍只允许现有 Prompt / LLM / condition / output。v2 不通过结构化输入扩展 tool、RAG、code、connector 或业务写回能力。

## Application Interaction Session

Application Session v4 只在 application active Definition 已绑定 `workflow_definition_executor_v2` 时启用：

- open session 时冻结 definition id / version / digest、pointer version、contract id / digest 与 application record version；
- 每个 turn 提交 `inputs`，服务在执行前重读同一 authority；
- turn record 只保存字段名 / 类型、bytes、digest、Run ref 和状态，不保存原始值；
- authority 或 contract 漂移时关闭后续 turn，不自动创建新 session、不回退 v1；
- 切换 application、definition、version、session、route 或组件卸载时，Web 清空尚未提交的字段值。

## Comparison、Evaluation 与 Campaign

- Comparison v7 只接受同 scope、同 Definition lineage、同 `workflow_definition_executor.v2`、同 `input_contract_digest` 的终态 Run v8。
- v5 与 v8、v1 与 v2 profile、不同 contract digest 或不同 Definition lineage 均返回明确 incompatible，不尝试字段映射。
- Evaluation Case、Baseline 与 Suite 保存 exact v2 profile 和 Run refs，继续复用既有 owner 与人工 decision。
- Application Evaluation Plan v2 的 Definition fixture 使用 `inputs`；一个 plan version 仍只允许一个 execution profile。
- Campaign v2 在开始和每项执行前校验 exact Definition authority 与 contract digest；漂移后停止剩余项。
- v1 Plan / Campaign 继续驱动 v1 Definition，不被自动升级；同一 campaign 不混合 v1 / v2 item。

## Web 与 Pencil 覆盖

本专题不创建新页面族，而是在 Definition 直接运行、Application Interaction Session 和 Application Evaluation Plan fixture 编辑三个既有任务区复用 `StructuredRuntimeInputEditor`。

五维评分为 `0 / 1 / 1 / 1 / 2 = 5`：没有新页面结构；增加类型化表单交互、secret / 合同漂移风险表达、窄屏字段顺序和三个 surface 的共享模式。因此覆盖级别固定为 `B / 局部 Pencil`，在 React 实施前冻结字段行、错误归属、不可变合同摘要、窄屏顺序和 value 清理提示；不升级为新的 S11 完整页面。

客户端只能使用服务端返回的 exact immutable contract 生成表单，不猜测字段，不缓存跨 authority 的输入值。颜色不是错误、必填、类型或选中状态的唯一通道。

2026-08-10 已在 [Family UI v1](../../designs/radishmind-web-family-ui-v1.pen) 完成 `B / 局部 Pencil`：Desktop `W3O4tV` 与 Narrow `t39foq`。首版 R1 因继承 S9 之后被退回的独立表单看板语言而失效；Visual R2 虽恢复薄页眉、输入路径、单一 editor owner 和 authority rail，仍因业务表面全部硬方形被再次退回；Visual R3 恢复职责圆角后，字段区仍像以整页横线切割出的静态列表。Visual R4 已把字段区改为有留白层级的表单画布，保留不可变合同、authority 与易失值边界，不再用表格替代真实输入控件。两张局部稿已于同日通过人工复核并冻结为批次 C 实现基准；它们仍只表达共享组件，不复制完整页面，也不建立 S11：

- Desktop 固定不可变合同上下文、Direct Run / Session 共用标识、非对称双列表单、字段级错误、布尔开关、authority / contract 失败归属、易失值说明和 Batch D 停止线；长文本字段独占整行，短字段依据任务关系成组，不堆成等权卡片；
- Narrow 固定 `context → path → editor owner → ephemeral boundary` 单列顺序，字段按 `process_description → target_temperature_c + include_recycle → operator_note` 渐进编排，不把桌面栏宽或整行分割机械压缩进 `390×844`；
- authority 切换后必须显示值已清理，类型 / required 错误贴在精确字段，contract drift 与 authority 失败接管提交区域并清空全部值；
- Pencil 中的示例合同 id、digest、字段名与占位文案只用于评审结构，不进入 React 或服务端真相源。

## 实施批次

### 批次 A：版本化合同与领域模型

- 增加 Draft v2、Candidate v2、Definition v2、Run v8 与 Comparison v7 schema。
- 实现 input contract / value canonicalization、digest、预算、secret 与兼容矩阵。
- 保留 v1/v5 完整读取、执行和比较路径，覆盖无自动迁移、无 fallback 负向测试。

批次 A 已完成。仓库已冻结五份 Draft / Candidate / Definition / Run / Comparison schema、`workflow_definition_executor.v2` 评测身份、最多 `16` 个字段的 Go 领域合同、稳定 canonical JSON / digest、预算与 secret 检查，以及八类稳定输入失败码。Saved Draft validation、Candidate / Version materialize、Run store / history 和 Comparison dispatch 已按显式版本矩阵分派；v1 不改写为 v2，v5 / v8 不混合比较，Comparison v7 只接受同 Definition lineage、同 contract digest 的终态 Run v8。

### 批次 B：HTTP、executor 与三模式持久化

- 实现 strict v1/v2 request union 与 `workflow_definition_executor_v2`。
- 增加 memory、SQLite、PostgreSQL migration / repository 支持和 restart / corruption / CAS 测试。
- 完成 Draft 保存、candidate、promotion、activation、direct Run 与 Run History 纵向链。

批次 B 已完成。Direct Run HTTP 现在按 exact Definition profile 执行严格 `input_text | inputs` 联合分派，executor v2 只把有界 typed packet 交给 provider；Run v8 与 History 仅保存合同、字段、字节数和 digest metadata。Saved Draft 与 Workflow Run 分别通过 SQLite `0004` / `0017`、PostgreSQL `0004` / `0020` 迁移接通 v2 / v8，并由数据库 payload / projection 约束与仓储解码双层失败关闭。memory、SQLite、PostgreSQL 已复验 Draft v2 → Candidate v2 → Definition v2 → Activation → Run v8 → Comparison v7，覆盖服务重启、隐私扫描、CAS / 行锁既有合同、腐化拒绝、关闭后无 memory fallback 和 v1 连续回归。

### 批次 C：Session 与共享 Web 输入编辑器

- 先冻结 `B / 局部 Pencil`，再实现共享 `StructuredRuntimeInputEditor`。
- 增加 Application Session v4 authority / turn 合同和 Definition direct Run / Session 两个 consumer。
- 验证值清理、错误归属、Desktop / 关键断点 / `390×844` 与 console。

局部 Pencil Visual R4 已完成结构、截图检查和人工视觉复核。批次 C 下一步只进入共享 `StructuredRuntimeInputEditor` strict contract decoder、Application Session v4 authority / turn 合同，以及 Definition Direct Run / Session 两个 consumer 的实现；Evaluation Plan / Campaign v2 继续留在批次 D。

### 批次 D：Evaluation 与 Campaign

- 让 Comparison、Evaluation Case / Baseline / Suite 支持 v2 profile。
- 增加 Application Evaluation Plan / Campaign v2 typed fixture、authority checkpoint 与 pair handoff。
- 保持 Campaign / Run metadata-only，不复制输入正文。

### 批次 E：连续产品链复验与收口

- memory、SQLite、PostgreSQL 覆盖 Draft → Definition → Run → Comparison / Evaluation。
- memory 与 SQLite 真实浏览器覆盖 Definition Run、Session 和 Campaign，并做服务重启恢复。
- 完成隐私扫描、无 fallback、旧 v1 兼容、三视口和全量仓库门禁后再关闭专题。

## 验收方式

- Go：contract、canonicalization、digest、domain、HTTP、executor、memory / SQLite / PostgreSQL repository 与 migration 测试。
- 兼容：v1 Draft / Definition / Run / Comparison / Session / Campaign 回归；v1/v2 混用明确失败。
- Web：strict decoder、共享 editor、错误状态、value 清理、交互与 production build。
- 浏览器：Desktop、关键断点、`390×844`；直接运行、Session、Campaign、刷新 / 切换清理、横向溢出和 console。
- 隐私：Run、Session、Campaign、日志与失败摘要均不得出现提交值或 secret fixture。
- 仓库：批次内先跑精准验证和 `./scripts/check-repo.sh --fast`；因新增 schema、migration、API、协议身份与阶段真相，专题关闭前运行全量 `./scripts/check-repo.sh`。

## 停止线

- 不支持 object、array、nullable、union、任意 JSON Schema、动态表单规则、默认值、enum、文件或图片上传。
- 不接受 secret、credential、Authorization、cookie、DSN、私钥或 connector token 作为 Workflow 输入。
- 不在 Run、Session、Campaign item、日志或审计摘要保存原始输入值。
- 不自动迁移 v1 Draft / Definition，不解析 `input_text`，不跨 v1/v2 fallback 或比较。
- 不扩 tool、RAG、code / sandbox、connector、外部 URI 抓取、业务写回或 proposed action 执行。
- 不支持 production enablement、schedule、queue、parallel fan-out、retry、resume、replay 或长期 agent loop。
- 不新增第二套 Run、Comparison、Evaluation、Session、Campaign、quota 或 route owner。
- 不为普通 UI 或现有测试可承载的行为派生同层 checker、fixture、readiness、review 或 refresh 链。

## 当前下一步

继续实施唯一任务卡批次 C：按已冻结的 Desktop `W3O4tV`、Narrow `t39foq` Visual R4 实现 strict contract decoder、共享 `StructuredRuntimeInputEditor`、Application Session v4 authority / turn 合同，以及 Definition Direct Run / Session 两个 consumer。当前不进入 Evaluation Plan / Campaign v2，不创建 S11 页面族，也不扩大 production、secret 输入、自动执行或业务写回边界。

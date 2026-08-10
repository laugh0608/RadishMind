# Workflow Definition 结构化运行输入（开发 / 测试态）v1 实施任务卡

更新时间：2026-08-10

状态：`workflow_definition_structured_runtime_inputs_dev_test_v1_batch_c_pencil_frozen_implementation_next`

对应功能文档：[Workflow Definition 结构化运行输入（开发 / 测试态）v1](../features/workflow/workflow-definition-structured-runtime-inputs-dev-test-v1.md)

## 任务目标

建立由不可变 Definition input contract 驱动的扁平强类型运行输入，使 Draft、Promotion、Direct Run、Application Interaction Session、Run History、Comparison、Evaluation 与 Application Evaluation Campaign 共享同一版本身份、校验和隐私边界。

本任务卡是该专题唯一高风险实施卡。schema、migration、executor、Session、Evaluation 与 Campaign 兼容升级都在这里收口，不派生同层 readiness、review、refresh 或 gate-only 任务。

## 允许修改

- `contracts/` 中 Saved Draft、Definition Candidate / Version、Run、Comparison、Application Session 与 Evaluation Plan / Campaign 的版本化 schema；
- `services/platform/internal/httpapi/` 中相邻 workflow definition、run、comparison、evaluation、application session / campaign owner；
- workflow saved draft / run 的 memory、SQLite、PostgreSQL repository、migration 与 selector；
- `apps/radishmind-web/` 中 Definition Run、Interaction Session、Evaluation Plan fixture 和共享输入组件；
- 对应功能入口、Family UI 产品化专题、current focus、roadmap、capability matrix 与周志。

## 不允许修改

- v1 Draft / Candidate / Definition、Run v5、Comparison v4 的既有含义和历史记录；
- Prompt、RAG、Agent application 的 fixture 与运行合同；
- Gateway quota、provider route、Application Publish / Assignment 状态机；
- tool、RAG、code / sandbox、connector、业务写回、production auth / membership / OIDC；
- schedule、queue、parallel fan-out、retry、resume、replay 或自动发布；
- 没有证明现有单元、集成或聚合门禁无法承载的新专项 checker / fixture。

## 冻结合同

1. v2 input contract 最多 `16` 个扁平字段，只允许 `string | integer | number | boolean`。
2. v2 请求只接受 `inputs`；v1 只接受 `input_text`；混合或跨版本请求失败关闭。
3. canonical payload 总计不超过 `8192` bytes，单 string 不超过 `4096` bytes；number 必须有限，integer 位于 JSON 安全范围。
4. 输入合同和值必须通过 secret material 检查，不允许 credential、token、Authorization、cookie、DSN、私钥或明显 endpoint secret。
5. Draft v2 → Candidate v2 → Definition v2 → executor v2 → Run v8 使用 exact schema、version、digest 和 contract digest，不自动迁移 v1。
6. Run v8、Session v4 与 Campaign item 只保存字段名 / 类型、bytes、digest 与 authority metadata，不保存原始值。
7. Evaluation Plan v2 可以保存用户显式提交且通过 secret 检查的 typed fixture；不得从 Run 恢复输入。
8. Comparison v7 只比较同 scope、Definition lineage、v2 profile 和 contract digest 的终态 Run v8。
9. 输入合同、authority 或 schema 漂移必须在 provider、quota、Session turn 等副作用前失败关闭。
10. memory、SQLite、PostgreSQL 行为一致；旧版本持续可读可运行，无 fallback。
11. Web 表单只能由 exact immutable contract 生成，并在 application / definition / version / session / route 切换或组件卸载时清空值。
12. UI 覆盖固定为 `B / 局部 Pencil`，五维评分 `0 / 1 / 1 / 1 / 2 = 5`；不创建 S11 完整页面。

## 批次 A：版本化 schema 与领域模型

- [x] 增加 Draft v2、Release Candidate v2、Definition Version v2 schema 与 Go 类型。
- [x] 增加 Run v8、Comparison v7 及 `workflow_definition_executor.v2` evaluation profile。
- [x] 实现 contract / input canonicalization、digest、预算、secret 与稳定 failure code。
- [x] 冻结 v1/v2 compatibility matrix，覆盖历史读取、无自动迁移、混合请求和无 fallback。
- [x] 更新 Saved Draft validation / release candidate / promotion domain tests。

批次 A 退出条件：相同合同与输入跨重复编码得到相同 digest；任何不合法值在执行副作用前失败；旧版本行为不变。

## 批次 B：HTTP、executor 与 durable store

- [x] 实现 strict v1/v2 request union 与 `workflow_definition_executor_v2`。
- [x] 扩展 memory repository 与服务层的 v2 Draft / Definition / Run 生命周期。
- [x] 增加 SQLite saved draft / workflow run migration、restart、CAS、corruption 与 no-fallback 测试。
- [x] 增加 PostgreSQL migration、rollback / reapply、row lock、restart 与 no-fallback 测试。
- [x] 完成 Direct Run 与 Run History metadata-only 展示合同。

批次 B 退出条件：三种 store 可完成 v2 Draft → Candidate → Promotion → Activation → Run v8，未迁移、错版本、store failure 与 authority drift 全部失败关闭。

批次 B 已通过 memory、SQLite 与 PostgreSQL 纵向复验。SQLite 使用 Saved Draft `0004` 与 Workflow Run `0017`，PostgreSQL 使用 Saved Draft `0004` 与 Workflow Run `0020`；Draft payload schema、Run contract projection 与 sanitized record 均有数据库约束和仓储复核。Direct Run / History 不回显值，provider 前 authority checkpoint 会复核 contract digest，v2 错误形状不回退 v1，关闭后的 repository 不回退 memory。

## 批次 C：Session 与共享 Web 输入编辑器

- [x] 在 Pencil 空闲且确认基准源后冻结 Visual R4 表单画布、合同摘要、错误归属、值清理提示与窄屏顺序。
- [ ] 实现共享 `StructuredRuntimeInputEditor` 及 strict contract decoder。
- [ ] 实现 Application Interaction Session v4 authority / turn 合同与 v2 executor bridge。
- [ ] 接入 Definition Direct Run 和 Session，覆盖 permission、contract drift、secret、type、budget 与 value clearing。
- [ ] 完成 Web tests、production build、三视口浏览器与 console 检查。

Pencil 前置证据：`docs/designs/radishmind-web-family-ui-v1.pen` 中 Desktop `W3O4tV`、Narrow `t39foq` 已冻结为 `B / 局部 Pencil · Visual R4` 并通过人工复核。两张局部稿覆盖 exact contract、四类真实输入控件、非对称桌面表单、渐进窄屏顺序、字段级错误、authority / contract 失败和值清理；不建立 S11，也不提前画 Evaluation Plan / Campaign v2。

批次 C 退出条件：三个 authority 切换边界不残留值；页面不猜测合同、不回显已提交输入；Desktop、关键断点与 `390×844` 无横向溢出。

## 批次 D：Evaluation 与 Campaign v2

- [ ] 让 Comparison、Evaluation Case / Baseline / Suite 接受 exact v2 profile / Run v8。
- [ ] 增加 Application Evaluation Plan / Campaign v2 typed Definition fixture。
- [ ] 实现 campaign authority / contract checkpoint、pair preview 与 existing Case / Suite handoff。
- [ ] 覆盖 v1/v2 plan 隔离、contract drift、secret fixture、partial handoff 与 restart。

批次 D 退出条件：同一 v2 plan version 可以产生兼容 Run v8 pair 并交接既有评测 owner；Campaign / Run 不复制 fixture 正文。

## 批次 E：连续产品链与关闭证据

- [ ] memory、SQLite、PostgreSQL 完成 Draft → Definition → Run → Comparison / Evaluation 连续链。
- [ ] memory 与 SQLite 浏览器完成 Direct Run、Session、Campaign，并在重启后恢复 exact evidence。
- [ ] 审计 Run、Session、Campaign、日志与失败摘要无原始输入或 secret 泄漏。
- [ ] 复验 v1 历史路径、三视口、no fallback、store identity 和 console。
- [ ] 同步正式文档并运行全量 `./scripts/check-repo.sh`。

批次 E 退出条件：纵向用户链、三存储、旧版本兼容、隐私与真实浏览器证据全部闭合，才允许把状态改为 completed。

## 验证矩阵

- Go contract / canonicalization / domain / HTTP / executor / repository / migration tests；
- SQLite / PostgreSQL migration、restart、CAS、corruption、rollback / reapply 与 scope 隔离；
- v1 Draft / Definition / Run / Comparison / Session / Campaign compatibility regression；
- Web strict consumer、共享 editor、交互、响应式与 production build；
- memory / SQLite 真实浏览器 Direct Run → Session → Campaign → Comparison / Evaluation；
- `git diff --check`、精准 `go test`、`./scripts/check-repo.sh --fast`；
- 因新增 schema、migration、API 与阶段真相，关闭专题前运行全量 `./scripts/check-repo.sh`。

## 当前下一步

继续实施批次 C。Pencil 前置已关闭，下一步实现 strict contract decoder、共享 `StructuredRuntimeInputEditor`、Application Session v4 authority / turn 合同与 Definition Run / Session 两个 consumer；本批不提前进入 Evaluation Plan / Campaign v2，也不建立 S11 页面族。

# 工程健康与产品化整改专题 v1

更新时间：2026-07-12

状态：`remediation_v1_in_progress`

## 专题目的

本专题是 2026-07-10 全仓审阅后的单一整改真相源，用于统一开发进度、产品规划、代码质量、安全边界、测试体系和文档治理的整改顺序。它不替代功能设计文档，也不为每一步再创建 readiness 专题、任务卡、fixture 或 checker。

整改目标不是扩大项目范围，而是把已经形成的平台、Workflow、Gateway、评测和安全能力收束成可维护、可复验、可持续迭代的产品基线。

## 审阅范围与基线

本次审阅覆盖：

- 根目录治理文件、正式规划文档、功能专题、平台专题、周志和任务卡。
- `services/platform/`、`services/gateway/`、`services/runtime/`、`apps/radishmind-web/`、`apps/radishmind-console/`。
- `scripts/`、`datasets/`、`contracts/`、Git 历史、CI 配置和仓库级验证入口。
- Go 竞态与覆盖率、Python 运行时行为检查、Web / Console 构建和仓库全量验证。

2026-07-10 审阅基线：

| 维度 | 事实 | 判断 |
| --- | --- | --- |
| 仓库规模 | 约 3,228 个 tracked 文件；`datasets/` 1,663、`scripts/` 825、`docs/` 514 | 资产丰富，但治理和验证体量已显著挤压产品实现空间 |
| 文件结构 | JSON 约 2,110、Markdown 521、Python 429、Go 44、TypeScript / TSX 52 | 评测和静态证据占主导，产品代码仍处于平台原型期 |
| 治理体量 | 约 273 张 task card、350 个 `check-*.py`、约 150,602 行 checker | 存在明显 gate-driven 漂移和重复维护成本 |
| 近期变化 | 2026-06-01 以来约 329 个提交、280,962 行新增、925 行删除 | 推进速度高，但新增远大于收敛，技术债净增长 |
| 文档入口 | `current-focus` 249 行但约 116 KB；`platform/README` 252 行约 69 KB | 行数预算未能阻止超长单行和证据链堆叠，入口可读性下降 |
| 后端质量 | 全量 Go test、`go test -race -cover ./...`、`go vet ./...` 通过 | 基础实现可靠，但 bridge / cmd 基线覆盖薄弱 |
| 前端质量 | Web 与 Console 构建通过；Web 主包约 833 KiB，Vite 给出大包 warning | 功能可构建，但需要路由级拆包与性能预算 |
| 仓库门禁 | fast 与 full `check-repo` 均通过 | 门禁强，但“门禁通过”不能替代真实用户路径验收 |

## 总体结论

当前项目应按“内部开发者预览”管理：安全边界、协议、评测和失败关闭意识较强，已经具备形成产品闭环的基础；但 durable persistence、正式身份、生产 secret、稳定运行时、端到端用户验收和性能预算仍未完成，因此不得声明 production ready。当前成熟度不再使用 `M2` 编号，避免与历史模型 / 评测阶段的 `M2`、`M3`、`M4` 发生语义复用。

主要矛盾已经从“缺少边界定义”转为以下四项：

1. 关键代码存在少量高影响正确性与安全缺口，需要先清零。
2. 产品纵向闭环少于横向 readiness 和 checker 证据链。
3. 文档、task card、fixture、checker 的数量增长快于实现收敛。
4. Gateway 仍依赖每请求启动 Python 子进程，产品化性能与稳定性不足。

以上四项是 2026-07-10 审阅基线；第 1 项已由 R2 清零，第 2 项已由 2026-07-11 的 Workflow 纵向产品链显著改善，第 4 项已由 R4 `stdio` worker pool 解决。R5 Web 包体预算已于 2026-07-12 接入现有 Vite build；当前剩余工程重点是低覆盖包，以及 R6 文档 / checker / 资产收敛。

## 2026-07-11 产品焦点与门禁调整决议

当前阶段不同时推进所有长期产品面。正式一级产品面继续保持四个：`User Workspace`、`Admin Control Plane`、`Model Gateway / API Distribution`、`Workflow / Agent Runtime`；`Image Generation / Artifact Return` 作为跨产品面的适配能力保留，不作为当前第五条一级产品主线。

当前首要用户是 Radish 体系内部开发者和团队成员，首要任务是创建、编辑、校验、保存、恢复和审查 Workflow 草案。Gateway 是支撑该用户任务的第一工程主线；Admin 只推进支撑 Workflow 与 Gateway 所必需的运维能力；Image Path、模型训练、quota / billing 和完整外部接入继续后置。

门禁从“限制实现文件是否可以存在”调整为“限制能力是否可以启用和声明”：

| 能力 | 开发 / 测试态 | 生产态 |
| --- | --- | --- |
| Saved Draft durable repository | R3 浏览器闭环完成后，可独立设计并实现显式开发 / 测试模式；必须有真实 migration、重启恢复、原子版本比较、scope 校验、no fallback 和集成测试 | 继续要求真实 OIDC、membership、生产数据库资源、secret、审计和部署复核，条件未满足时不得启用 |
| Repository adapter / migration source | 允许代码和测试存在，不再以“文件必须不存在”作为长期门禁 | production selector 必须默认关闭并失败关闭，不得因代码存在而声明 ready |
| Workflow executor | durable repository 与稳定 Gateway 成立后，可打开无外部副作用的安全执行切片，只允许 Prompt / LLM / condition / output 和 run record | unrestricted tool、业务写回、自动 confirmation commit、replay / resume 继续关闭 |
| OIDC / production secret / cloud provider | 可实现接口、fake / local adapter 和负向行为测试 | 只有真实 issuer、membership source、provider account / resource / endpoint 与负责人明确后才允许启用 |
| readiness / review / refresh 证据链 | 历史证据保留，禁止继续派生同层链 | 生产准入只保留聚合清单、行为证据和一次明确复核 |

数据 ownership 同步澄清：`Radish` 继续是用户身份、组织成员关系和上层业务数据的真相源；RadishMind 可以并且需要拥有自身运行数据，包括 Workflow draft、Workflow version、run record、trace、provider usage 和 RadishMind audit。该边界禁止复制上层身份与业务权威，但不再把 RadishMind 自身 operational database 解释为越界。

本决议只调整允许推进的范围和准入方式，不表示数据库、repository mode、executor、OIDC 或 production secret 已经实现或启用。

## 整改原则

- 功能设计文档先行，后续默认以一个用户目标驱动实现、测试和文档更新。
- 在制主线最多两条：一条产品纵向切片，一条工程健康整改；阻塞项不继续派生同层 readiness 文档。
- 新增 checker 必须证明现有单元测试、集成测试或聚合门禁无法承载；第一阶段 checker 数量净增必须为零。
- dev / test 能力与 production 能力分别验收，不用静态证据或 fake runtime 代替生产声明。
- 修复优先落在根因和稳定接口，不叠加吞错、默认值或隐式 fallback。
- 任何密钥不得进入命令行参数、公开错误、日志、fixture 或响应正文。

## 整改工作流

| 编号 | 工作流 | 优先级 | 当前状态 | 完成条件 |
| --- | --- | --- | --- | --- |
| R1 | 规划与在制品治理 | P0 | 进行中 | 当前焦点指向本专题；停止新增同层 readiness 链；每批只有一个明确用户目标 |
| R2 | 正确性与安全清零 | P0 | 已完成 | CAS、竞态、错误脱敏、密钥传递、请求体边界均有行为测试且仓库全量通过 |
| R3 | Workflow Draft Review Loop 产品闭环 | P1 | 已完成 | 创建、编辑、校验、保存、冲突、恢复、审查交接形成一条可重复端到端路径 |
| R4 | Gateway 运行时产品化 | P1 | 已完成 | 受控 `stdio` worker pool 已默认启用，生命周期、隔离、恢复、错误语义和 p95 降幅均有证据 |
| R5 | 测试、CI 与性能预算 | P1 | 进行中 | Gateway 性能预算已建立；Web 主入口降到 430.39 KiB，主入口 / named / lazy chunk 预算已接入 Vite build；下一步继续处理剩余低覆盖包 |
| R6 | 文档、checker 与资产收敛 | P2 | 进行中 | 入口文档恢复短入口；活跃 checker 和 task card 显著下降；历史证据可索引但不占当前主线 |

## R1：规划与在制品治理

立即执行：

- 本专题作为整改总入口；跨领域整改不再拆成同层专题。
- `docs/radishmind-current-focus.md` 只保留当前阶段、两个在制目标、下一动作和停止线。
- 旧 readiness / review 文档保留为历史证据，不再通过 `v2 / v3 / refresh after review` 延长静态链。
- 只有新增公开 API、schema、生产声明、外部 provider 或执行边界时，才允许独立高风险 task card。
- 门禁默认约束配置启用、运行行为和发布声明，不再把未来实现文件必须缺席作为长期完成条件。
- “不做玩具式最小实现”解释为交付小而完整、可复验的纵向切片，不解释为开发态功能必须先满足全部生产依赖。

每次迭代评审只回答四个问题：

1. 本批解决了哪个用户或运维人员的真实任务？
2. 哪条可执行测试证明行为成立？
3. 哪些边界仍未打开？
4. 是否删除、合并或替代了旧证据，而不是只新增文件？

## R2：正确性与安全清零

第一整改批覆盖以下问题：

| 问题 | 根因整改 | 行为证据 | 状态 |
| --- | --- | --- | --- |
| Saved Draft 并发 map 访问与丢失更新 | memory store 使用 `sync.RWMutex`；写入接口改为存储层原子 compare-and-swap | 并发 24 个相同 expected version 只允许一个成功；`go test -race` | 已实现 |
| information finding 被判为 invalid | validation state 只由 blocking finding 或 blocked capability 降级 | 高风险节点已要求确认时保留 info finding，状态仍为 `valid_for_review` | 已实现 |
| bridge API key 出现在 argv | Go 子进程仅通过请求级环境变量传入，显式移除继承的陈旧值；Python bridge 删除 `--api-key` | 参数不含 secret；无密钥请求不继承上一请求 secret | 已实现 |
| provider 原始错误正文进入公开 envelope | provider transport error 归一化；Gateway 使用稳定错误消息 | HTTP 429 响应正文和异常 secret 均不能出现在错误 envelope | 已实现 |
| HTTP 请求体无统一上限且允许尾随 JSON | 新增统一 decoder；公开路由 4 MiB、控制写入 1 MiB；只允许一个 JSON 文档 | 413、尾随 JSON、内部未知字段拒绝测试 | 已实现 |

本批不打开数据库、OIDC、repository mode、production secret resolver、executor、publish、writeback 或 replay。

## R3：Workflow Draft Review Loop 产品闭环

2026-07-10 第一批先修复 saved version 生命周期：编辑、validate 和非冲突失败不再丢失 persisted base version，未处理的 `version_conflict` 不能通过重复 Save 绕过显式 Continue / Restore；行为测试已进入 Web PR / release CI。

2026-07-11 已使用真实 Web consumer 与 Go dev-only route 完成 dev-live 收口：正常路径覆盖创建、节点 / 边 / 属性 / 布局编辑、校验、连续保存、列表刷新与显式恢复；冲突路径通过独立写入推进服务端版本，确认本地草案保持不变、未决冲突锁定编辑与 dev route 动作、Continue 使用当前 saved version 重试、Restore 只在显式选择后替换本地草案，Review Handoff 同步展示 conflict evidence。复验同时修复 Handoff layout evidence 重复 key，并为 launcher 增加显式 `--saved-draft-dev` / `-SavedDraftDev` 与 Saved Draft route probe；最终新浏览器会话为零 error / warning。

下一条产品主线使用现有能力形成一条可复验路径，不继续扩同层只读面板：

1. 用户从 Workspace Home 创建草案。
2. 在 Draft Designer 编辑节点、边、属性和布局。
3. 执行校验并在图上定位 blocking / info finding。
4. 保存草案、刷新列表并恢复同一 saved record。
5. 制造版本冲突，保留本地编辑并显式选择恢复 saved version。
6. 在 Review Handoff 查看 validation、plan、readiness 和 conflict evidence。

验收要求：

- 使用真实 Web consumer 与 Go dev-only route，不把 sample / fixture 当成功保存路径。
- 一条浏览器行为测试覆盖正常路径和版本冲突路径。
- 页面明确标记 dev-only、saved、unsaved、conflict 和 review-only 状态。
- 失败时不回退 sample，不执行 workflow，不创建 confirmation decision。
- 本批完成前不再新增 Workflow readiness / review 小切片。

R3 完成后已停止继续扩同层 Builder 小切片。显式开发 / 测试态 PostgreSQL durable repository 已完成 migration / rollback / reapply、运行账号权限隔离、重启恢复、原子 expected-version、tenant / workspace / application / owner scope、no fallback、CI 和双标签浏览器冲突验收。旧 Saved Draft / Secret Backend / Storage Adapter readiness checker，以及 Control Plane Read formal UI 之后的 repository readiness 尾链，均已退出活动仓库基线；历史脚本和证据继续保留。

## R4：Gateway 运行时产品化

2026-07-11 已完成 mock provider 顺序 / 并发分段基线、选型与迁移。受控 `stdio` worker pool 当前默认复用四个 Python worker，process 模式保留显式回滚：

1. 已建立 mock provider 的顺序与并发基线，分离 Go 路由、进程启动 / IPC、Python Gateway 和 provider 延迟。
2. 已在现有 `bridgeClient` 边界后评审候选，选择不新增端口且每个 worker 只处理一个在途请求的受控 `stdio` worker pool。
3. 已实现生命周期、版本化健康握手、有界排队、并发上限、取消、超时、崩溃后重建和并行优雅退出。
4. 已保留 mock / offline 与 process 回滚能力；真实 provider 仍只允许通过 secret ref 或受控请求级凭证进入。

验收要求：

- bridge 自身开销相对当前基线至少下降 70%，并报告 p50 / p95。
- 并发请求没有跨请求 credential、上下文或 stream event 污染。
- worker 崩溃、超时和取消都有稳定错误码，不泄漏 stderr 原文。
- 不新增第二套 Gateway 业务真相源。

完成结论：顺序 / 四并发 bridge 自身 p95 从 `62.796 / 75.439 ms` 降到 `4.098 / 4.252 ms`，降幅 `93.5% / 94.4%`；并发隔离、queue full、timeout、cancellation、crash recovery、协议错帧、credential 清除和 close 回收均有行为测试，R4 收口为已完成。

## R5：测试、CI 与性能预算

整改项：

- PR CI 增加 `go test -race ./...`、`go vet ./...`、Web build、Console build 和现有关键 Python 行为检查。
- bridge、HTTP request boundary、Saved Draft CAS 和公开错误脱敏必须由行为测试覆盖。
- Python 新测试优先进入可发现的标准测试层；既有 `check-*` 仅保留仓库聚合和跨资产验证职责。
- Web 按页面和重依赖拆包，首阶段把主入口 chunk 控制在 500 KiB 以内；若不能达到，记录具体依赖和下一步。
- 建立关键路径的延迟、失败率、包体和覆盖率基线，不用单一总覆盖率掩盖 0% 包。

2026-07-12 进展：React runtime、Node Designer、Run Comparison、Evaluation Cases 与 Evaluation Suite 已拆为独立 chunks，主入口从审阅时约 833 KiB 降到 430.39 KiB。Vite build 现以原始输出字节执行 500 KiB 主入口、220 KiB `react-vendor`、220 KiB Node Designer 和 64 KiB 普通 lazy chunk 预算；缺少关键 named chunk 或任一 chunk 超限都会使现有 `npm run build` 失败并输出实际值与预算值。预算逻辑有单元测试，不新增平行 gate-only checker。R5 下一步转向剩余低覆盖包，不继续机械拆包。

## R6：文档、checker 与资产收敛

第一阶段目标：

- `docs/radishmind-current-focus.md`、`docs/platform/README.md` 等入口回到短入口职责，长历史迁入专题索引或归档摘要。
- 普通入口避免超长单行；单行目标不超过 2,000 字符。
- `docs/radishmind-current-focus.md` 的人工默认阅读区先收敛到 10 KiB 以内；历史机器锚点在 checker 解耦前进入明确兼容区，不再混入当前决策。
- 两个迭代内把活跃 `check-*.py` 数量和 checker 代码量各降低至少 15%，优先合并相同 schema / fixture / readiness 扫描。
- task card 索引区分“活跃、阻塞、历史完成”，默认读取只展示活跃项。
- 大 JSON 继续使用主索引 + `.parts/`，不新增长路径语义编码。

删除或合并前必须先确认：调用方、fixture、CI 注册、文档引用和历史证据是否仍需要；不以批量删除换取表面指标。

## 推进顺序

### 第一阶段：P0 清零

- 完成 R2 代码与行为测试。
- 跑 Go race / vet、关键 Python check、Web / Console build、fast 和 full 仓库验证。
- 更新 Saved Draft 专题、当前焦点和周志。

### 第二阶段：产品闭环

- 只推进 R3。
- 完成浏览器端到端复验、冲突恢复体验和明确状态文案。
- 同步拆分 Web 主包并补 CI 构建。

### 第三阶段：运行时与治理收敛

- R4 受控 `stdio` worker pool 已完成；后续不扩同层 readiness 或 benchmark 链。
- 并行启动一批 R6 合并工作，但在制整改批次最多一个。
- production durable store、OIDC 和 production secret 只有在真实上游资源与负责人明确后才能进入启用评审；显式开发 / 测试态 durable store 已在 R3 后作为独立产品纵向切片完成。

### 第四阶段：开发态持久化与安全执行

- 产品线的显式开发 / 测试态 PostgreSQL durable repository 已完成，不依赖 production secret audit store；production repository 继续关闭。
- 当前工程线继续 R5：Gateway 性能预算已建立，Web 主包拆分和可发现 build 预算已完成；下一步处理剩余低覆盖包。
- durable repository 与稳定 Gateway 均已通过；下一产品批先做无外部副作用 executor v0 功能设计与评审。本阶段仍不打开 unrestricted tool、业务写回、自动确认提交或 replay。

## 每批完成定义

每个整改批次只有同时满足以下条件才可完成：

- 用户行为、失败行为和停止线均有可执行证据。
- 精准测试、相关构建和仓库门禁按风险通过。
- 没有新增 secret 泄漏、隐式 fallback、竞态或部分写入。
- 代码、功能专题、当前焦点和周志口径一致。
- 未验证项、外部阻塞和回滚方式明确记录。
- 新增文件数量有理由，且优先合并或替代旧证据。

## 风险与回滚

- CAS 接口变更只作用于当前 Saved Draft store 抽象；若未来 repository adapter 接入，必须实现相同 expected version 原子语义。
- bridge 凭证环境变量只存在于子进程环境；回滚不得恢复 argv 传密钥，可改为受控 stdin 或 secret handle。
- 请求体限制若影响合法大请求，应以采样数据调整 endpoint 预算，不得直接移除全局边界。
- 错误归一化会减少公开诊断细节；详细原因应进入受控内部观测，不得重新写回客户端。

## 停止线

- 不继续创建“readiness 之后的 readiness”来替代实现或真实阻塞说明。
- 不把 memory dev store、fake resolver、静态 schema artifact 或离线 smoke 写成 production ready。
- 不在缺少真实 issuer、membership source、生产数据库资源、secret backend 和负责人时启用 production repository mode；显式开发 / 测试模式必须独立命名、默认关闭生产入口并保持 no fallback。
- 不在本轮引入新语言栈、重写 Gateway 或跨工作区修改上层项目。
- 不把模型建议直接写入上层业务真相源。
- 不让 API key、token、DSN、provider 原始响应或异常正文进入 argv、公开错误、日志和 committed 资产。

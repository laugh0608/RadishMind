# Image Generation / Artifact Return 设计与开发文档

更新时间：2026-07-26

状态：`image_adapter_controlled_invocation_artifact_return_dev_test_v1_completed`

## 功能定位

`Image Generation / Artifact Return` 负责把模型侧结构化 image intent、安全约束和 artifact metadata 转成可审查、可追踪、可返回的图片生成结果引用。图片像素生成由独立 image adapter 和 backend 承接，不并入 `RadishMind-Core` 主模型职责。

## 当前状态

- Image Path 已完成 adapter handshake / safety gate、artifact return runbook、安全 runbook、backend adapter readiness、artifact runtime mapping readiness、store / binary reader boundary readiness、metadata-only runtime mapper、response consumer 和 `coerce_response_document` metadata-only response builder runtime integration。
- runtime integration 只从 request artifact metadata 发现 `image_generation_artifact`，通过 mapper / consumer 合并到现有 `CopilotResponse.citations` artifact citation。
- 批次 A 已完成受控调用纯领域 runtime；批次 B 已完成本机私有 content-addressed store；批次 C 已完成 reference-only backend profile 编译；批次 D 已完成 test-only `contract_fixture` 具体 client 与 canonical artifact metadata owner 收口；批次 E 已完成一次性 binary delivery、private store 重验与成功引用延后释放。
- 当前仍不改 `CopilotResponse` schema，不创建 public URL resolver，不解析 endpoint / credential / model-dir 引用，不调用真实生图 backend，也不生成或上传图片；fixture bytes 只在协调器内部单次进入本机私有 store。

2026-07-25 已选择下一实现方向为“开发测试态 Image Adapter 受控调用”，不再继续派生 metadata-only readiness 链。首批只实现独立 Python 领域运行时：

- 严格读取既有 `image_generation` intent，不建立第二套 intent / backend request / artifact contract。
- 纯函数编译 `image_generation_backend_request`，backend profile、request id 与 timeout 由调用侧显式提供。
- 在任何 backend side effect 前执行 schema、预算、敏感材料、profile exact match 与 safety gate。
- backend client 只通过注入式协议调用一次；不自动 retry，不 fallback，不解析 credential、endpoint 或 model dir。
- backend 返回 artifact metadata 与由 transport 观察到的 hash / MIME / dimensions；adapter 重验 lineage、generation、safety、provenance 和观察值，再复用既有 mapper 形成 artifact citation / 内部 metadata reference。
- 本批不创建 HTTP API、Gateway route、store、binary reader、public URL、Web、生产 profile 或真实 backend client。

## 设计边界

- artifact metadata 只允许作为 metadata-only reference 返回。
- `blocked / failed / pending_review` artifact 不能进入成功 response。
- public URL、signed URL、binary payload、provider raw dump、hash / mime / dimensions mismatch、safety review missing 或 provenance missing 都必须 fail closed。
- response builder 接线不等于 artifact store、public delivery 或 backend generation ready。

## 下一批开发方向

1. [Image Adapter 受控调用与 artifact 返回（开发 / 测试态）v1 实施任务卡](../task-cards/image-adapter-controlled-invocation-artifact-return-dev-test-v1-plan.md)批次 A 已完成。
2. 批次 B 的 `local_private_artifact_storage` owner 已完成：二进制容器观察与私有存储按稳定职责拆分，仍属于同一边界。
3. 批次 C 已选择 profile / credential 配置单一方向：strict source 只承载 reference-only endpoint / credential / model-dir，编译稳定 `profile_digest`，并成为批次 A 的 profile 与 timeout owner。
4. 批次 D 已选择 test-only `contract_fixture` 单一 client：只检查真实 fixture 图片容器并返回最小 observation，canonical artifact metadata 由 adapter 构造。
5. 批次 E 已完成 fixture 二进制到既有本机私有 store 的单次交付协调；只有持久化成功后才释放 citation / metadata reference，开发测试态 v1 随之关闭。
6. 生产 backend call 仍需要独立 credential resolver、endpoint/model-dir resolver、moderation、安全复核、运行配置和发布声明，不能由开发测试态 fixture client 代替。
7. 普通 metadata mapping 文案和 runbook 调整继续复用现有测试与仓库基线，不恢复同层 checker 链。

### 批次 B 固定边界

- store root 只能由调用侧显式注入绝对本机路径；模块不读取环境变量，不选择 production object store。
- blob 以 sha256 和 canonical format 形成 content-addressed 相对路径；artifact ref 单独绑定 `artifact_id / artifact:// URI / digest / MIME / dimensions / format / size`，两者均不可变。
- 写入前严格校验既有 artifact schema、安全状态、provenance、canonical URI、payload 大小、sha256、MIME 与 dimensions；不接受 base64、provider raw payload、public URL 或 signed URL。
- lookup 只返回内部 storage handle 和已验证 metadata，不读取图片二进制；binary reader 默认拒绝，只有显式 `allow_binary_read=true` 才能打开一次内部流。
- reader 每次读取都重新计算 sha256，并重新识别 MIME 与 dimensions；校验通过后只把临时 stream 交给一次内部 consumer，不在 result、日志、引用或响应中返回 bytes。
- 本批不提供 delete / overwrite、upload、public URL resolver、retry / fallback、后台服务、repository selector、migration 或生产存储声明。

### 批次 C 固定边界

- `image-backend-profile-source.schema.json` 只允许 `development | test`，不接受 production / staging、未知字段或第二套 backend request 配置。
- `remote_https` 只能绑定同环境 `endpoint_ref + secret_ref`；`local_model` 只能绑定同环境 `model_dir_ref` 且 credential 为 `not_required`。两种模式严格互斥。
- endpoint、credential 与 model-dir 均使用 `ref:radishmind/<environment>/image-backends/...` 引用；真实 URL、密钥、header、DSN、绝对路径、环境变量、自由 system prompt、provider / runtime 配置全部拒绝。
- 编译产物固定 profile identity、backend/model/adapter profile、runtime binding、credential policy、timeout 与稳定 `profile_digest`。相同 source 映射顺序不影响 digest，任何字段漂移都会改变或破坏 digest。
- `invoke_image_generation` 只接受 digest 可重算的编译 profile，并从 profile 读取 timeout；profile 缺失、disabled、digest 漂移或 backend preference mismatch 都在 client 调用前失败关闭。
- 本批不解析任何 reference，不读取 secret / 环境变量 / 文件系统，不连接 endpoint，不创建具体 backend client，也不启用 retry / fallback。

### 批次 D 固定边界

- `contract_fixture` profile 只允许 `test`，必须 disabled credential 且 endpoint / secret / model-dir refs 全部为 null；不得由 development 或 production 使用。
- `ContractFixtureImageBackendClient` 只消费 canonical backend request、编译 profile、固定 UTC 时间与一份调用侧注入的 PNG / JPEG / WebP fixture。
- client 创建时检查真实图片容器、hash、MIME、dimensions、大小和像素预算；调用时精确匹配 profile identity、timeout、output、空 input 与低风险 safety gate。
- client 只返回 deterministic artifact id、UTC 时间和二进制 observation，不返回 bytes、base64、provider response、title、purpose、safety 或 provenance。
- adapter 根据 intent、backend request 和 observation 构造唯一 canonical artifact metadata；backend 不再成为第二套 artifact metadata 真相源。
- 本批不持久化或返回 fixture bytes，不接 store coordinator，不解析 reference，不连接网络，不读取模型目录，不生成图片，也不启用 retry / fallback。

### 批次 E 固定边界

- 新增单一领域协调器，编排既有 adapter、test-only fixture client 与 `LocalPrivateImageArtifactStore`；不把持久化职责塞入 adapter 或 backend client，也不创建第二套 artifact store。
- fixture client 在构造时拥有不可变二进制，并只允许对本次已验证的 artifact identity、observation 和 canonical metadata 执行一次内部交付；bytes 不进入 invocation result、citation、metadata reference、日志或异常。
- 固定顺序为 adapter 完成 profile、request、safety、observation 与 canonical artifact 校验后，协调器请求一次 binary delivery，private store 再检查容器、大小、hash、MIME、dimensions、format、safety 与 provenance。
- 只有 private store 成功返回不可变内部记录后，协调器才释放既有 citation 与 metadata reference；交付拒绝、重复交付、observation 漂移、store 冲突、完整性失败或不可用均失败关闭，不返回成功引用。
- side-effect counter 必须区分 backend handoff、fixture binary delivery、private store write、binary revalidation、retry、fallback、upload 与 public URL；成功和所有失败路径都要求精确计数。
- 本批不修改三份 image schema 或 `CopilotResponse`，不实现 remote HTTP / local model client、reference resolver、API / Gateway / Web、production storage、upload、public URL、retry / fallback 或生产声明。

## 验收方式

- metadata-only runtime：runtime unit tests、image artifact checker、fast baseline。
- store / reader：hash / mime / dimensions revalidation、binary leak negative tests、no side effects checks。
- backend adapter：credential boundary、safety gate、timeout/failure taxonomy、no upload by default 和全量仓库验证。
- 受控调用批次 A：Python 相邻单元测试覆盖纯编译、单次调用、预算、安全、敏感材料、profile drift、artifact 观察值与 provenance；timeout 从批次 C 编译 profile 读取。
- Profile 配置批次 C：strict schema、确定性 digest、开发测试环境、远程 / 本机互斥绑定、引用作用域、UTF-8 / source byte / timeout budget、敏感材料和 digest drift 负向测试。
- Binary delivery 批次 E：相邻测试覆盖 PNG / JPEG / WebP 持久化、确定性引用、一次性交付、store 成功前不释放引用、重复 / 错 artifact 交付、payload / observation 漂移、store 冲突 / 不可用、二进制零泄露和精确副作用计数。

## 批次 A 完成结果

1. 新增 `services/runtime/image_generation_adapter.py`，复用三份既有 schema 和 metadata mapper；没有新增第二套 intent、backend request 或 artifact 类型。
2. `compile_image_backend_request` 对相同 intent、profile 和 request id 生成稳定请求；trace 固定包含 source request、intent 与 backend request lineage。
3. `invoke_image_generation` 在调用前完成 schema、UTF-8、尺寸 / 像素 / count、参数、列表、标识符、locale、敏感材料、profile exact match 和 low-risk safety gate；所有前置失败 `backend_call_count=0`。
4. 注入式 client 最多调用一次；timeout、unavailable 与未知异常不 retry / fallback，也不把异常原文、endpoint 或 credential 投影到结果。
5. 成功返回必须同时提供 schema-valid artifact metadata 与 transport 观察到的 sha256 / MIME / dimensions；adapter 重验 canonical URI、lineage、generation、title / purpose、safety 和 provenance 后才生成既有 artifact citation / metadata reference。
6. 11 项相邻单元测试已纳入 `services/runtime/tests` 和快速仓库入口；store lookup、binary read、upload、public URL、production storage、executor、confirmation、writeback 与 replay 均保持 0。

## 批次 B 完成结果

1. `services/runtime/image_artifact_binary_inspection.py` 使用标准库确定性观察 PNG / JPEG / WebP 容器、sha256、MIME、dimensions、format 与 size，不加载图片模型或第三方解码器。
2. `services/runtime/image_artifact_private_storage.py` 实现调用侧绝对路径注入、私有目录 / 文件权限、content-addressed blob、不可变 artifact ref、原子 no-overwrite 写入和 metadata-only lookup。
3. store 写入前复用 strict artifact schema 与既有 mapper，精确重验 canonical URI、低风险安全状态、provenance、hash、MIME、dimensions 和 format；相同 artifact 幂等，绑定漂移失败关闭。
4. binary reader 默认拒绝；显式授权后只读取一次，在 consumer 前再次重验容器、hash、MIME、dimensions、format 和 size。consumer 最多调用一次，result 不返回 bytes、绝对路径、base64 或 URL。
5. 15 项 storage 相邻测试覆盖三种格式、真实临时目录、并发幂等、payload / metadata 边界、ref / blob / symlink 篡改、读取授权、consumer 异常脱敏和零外部副作用；runtime 测试总数为 26。

## 批次 C 完成结果

1. 新增 `image-backend-profile-source.schema.json` 与 reference-only 基础 fixture，固定 profile identity、环境、backend、runtime、credential policy 和 timeout source。
2. `services/runtime/image_backend_profile_configuration.py` 纯函数编译远程 / 本机两种互斥配置，生成稳定 `sha256:` profile digest；不解析 reference 或读取任何外部配置。
3. 批次 A 的 `ImageBackendProfile` 已收口为编译产物，timeout 不再由调用点独立注入；disabled、digest drift、非法引用或 backend mismatch 均在 backend side effect 前拒绝。
4. 10 项 profile 配置测试与 12 项 adapter、15 项 storage 测试共同形成 37 项 runtime 测试；既有 Image contract / mapper / consumer / builder 检查继续兼容。
5. 状态推进为 `image_adapter_controlled_invocation_artifact_return_dev_test_v1_batch_c_completed_batch_d_review_required`。下一步只评审一个具体 backend client，不打开 HTTP / Gateway / Web 或 production capability。

## 批次 D 完成结果

1. Profile compiler 新增 `contract_fixture` 模式，只允许 `test` 且不绑定 endpoint、credential 或 model-dir reference。
2. 新增 `image_backend_contract_fixture_client.py`，以实际 fixture 容器计算 sha256、MIME 与 dimensions，并对 backend request、profile、timeout、output、input 和 safety 执行 exact match。
3. `ImageBackendInvocationResult` 缩减为 artifact identity、UTC 时间与 observation；adapter 统一构造 title、purpose、generation、safety、provenance 和 canonical URI。
4. 7 项具体 client、11 项 profile、12 项 adapter 与 15 项 storage 测试共同形成 45 项 runtime 测试；结果不泄露 bytes、endpoint、credential 或 provider raw material。
5. 状态推进为 `image_adapter_controlled_invocation_artifact_return_dev_test_v1_batch_d_completed_batch_e_review_required`；日终设计复核随后确认了批次 E 的一次性交付、私有持久化顺序和失败关闭语义，当前状态为 `image_adapter_controlled_invocation_artifact_return_dev_test_v1_batch_d_completed_batch_e_ready`。

## 批次 E 完成结果

1. 新增 `services/runtime/image_artifact_delivery_coordinator.py`，固定 adapter 校验成功 → fixture binary 单次交付 → private store 重验与持久化 → 释放既有 citation / metadata reference 的唯一顺序。
2. `ContractFixtureImageBackendClient` 复制并持有不可变 fixture bytes，使用线程安全状态精确绑定已完成调用、artifact identity、observation 与 canonical metadata；未调用、错 artifact、observation 漂移、重复和并发交付均失败关闭。
3. `LocalPrivateImageArtifactStore` 新增 `artifact_binary_revalidation_count`，并在成功、metadata / payload 拒绝、store 冲突、完整性失败与不可用路径记录真实重验和已发生的逻辑写入。
4. 协调结果只在 store 成功后返回 citation / metadata reference；所有失败结果均不含 backend request、artifact document、bytes、base64、绝对路径或 storage ref，store 与 consumer 内部异常使用稳定脱敏失败语义。
5. 新增 11 项协调器相邻测试，Image runtime 测试总数增至 56，覆盖 PNG / JPEG / WebP、确定性绑定、调用前与 backend 失败、错 artifact、observation / payload 漂移、重复 / 并发交付、consumer 异常、store 冲突 / 不可用、零泄露与精确副作用计数；定向测试、差异卫生、fast 与 full 仓库门禁均已通过。状态推进为 `image_adapter_controlled_invocation_artifact_return_dev_test_v1_completed`，不派生批次 F 或同层 checker。

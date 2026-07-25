# Image Adapter 受控调用与 artifact 返回（开发 / 测试态）v1 实施任务卡

更新时间：2026-07-25

状态：`image_adapter_controlled_invocation_artifact_return_dev_test_v1_batch_d_completed_batch_e_review_required`

对应功能文档：[Image Generation / Artifact Return 设计与开发文档](../features/image-generation-artifact-return.md)

## 任务目标

把既有 image intent → backend request → artifact metadata 契约从静态 fixture 推进为可测试的开发测试态领域运行时。首批只打开 Image Adapter 本身：调用者显式注入 backend profile、backend client、request id 和 timeout，adapter 在 side effect 前完成全部确定性校验，成功时只返回现有 artifact citation 与内部 metadata reference。

## 根因

- metadata-only mapper、response consumer 与 `coerce_response_document` hook 已完成，但只能消费调用前已经存在于 request metadata 的 artifact。
- 三份 image schema 与 safety / backend readiness 已冻结，缺少真正执行这些规则的单一 runtime owner。
- 继续新增 readiness / entry review 不会增加用户路径，也无法证明 backend 调用前阻断、单次调用和返回重验。

因此批次 A 创建一个独立、无 IO 所有权的 adapter service；真实 backend client、凭据、endpoint、model dir、store、binary reader、public URL 与产品 API 继续后置。

## 批次 A 实现边界

### 允许

- 新增 `services/runtime/image_generation_adapter.py`。
- 新增 `services/runtime/tests/test_image_generation_adapter.py`。
- 将 `services/runtime/tests` 纳入现有 Python unittest 仓库入口。
- 更新本功能文档、任务卡索引、当前焦点、路线图、能力矩阵和周志。

### 禁止

- 不修改三份 image schema 或 `CopilotRequest / CopilotResponse`。
- 不创建 HTTP / Gateway route、API key scope、应用配置、repository、migration、Session、Run 或 Web。
- 不创建 credential / endpoint / model-dir resolver，不读取环境变量。
- 不创建生产 backend client，不下载模型，不生成、保存、上传或公开图片。
- 不创建 artifact store、binary reader、public URL / signed URL resolver。
- 不自动 retry、fallback、replay、业务写回或发布。

## 固定领域流程

1. 使用 Draft 2020-12 schema 严格校验 intent。
2. 校验 UTF-8 prompt budget、尺寸 / 像素 / count、steps / guidance、列表数量、trace 与标识符预算。
3. 拒绝 credential、token、authorization header、cookie、endpoint URL、DSN、私钥或 provider raw 材料。
4. 只允许 `risk_level=low` 且 `requires_confirmation=false`；其它情况在 backend 调用前失败关闭。
5. `intent.backend.preferred` 必须与显式 profile backend id 精确一致；不自动选择或 fallback。
6. 纯函数编译 backend request，并通过既有 backend request schema。
7. client 最多调用一次；timeout、unavailable 与异常使用稳定失败语义，不 retry。
8. client 只返回 artifact metadata 和 transport 观察到的 sha256 / MIME / dimensions，不返回像素或 provider raw payload。
9. adapter 严格校验 artifact schema、intent / request lineage、backend / model / generation 参数、title / purpose、safety、provenance 与 transport 观察值。
10. 成功结果复用 `map_image_artifact_to_response_reference`，只暴露 artifact citation 与内部 metadata reference。

## 稳定失败语义

- `image_intent_invalid`
- `image_intent_budget_exceeded`
- `image_intent_sensitive_material_rejected`
- `image_intent_requires_confirmation`
- `image_intent_high_risk`
- `image_backend_profile_missing`
- `image_backend_profile_mismatch`
- `image_backend_safety_gate_blocked`
- `image_backend_timeout`
- `image_backend_unavailable`
- `image_backend_invalid_artifact_metadata`
- `image_backend_artifact_lineage_mismatch`
- `image_backend_artifact_hash_mismatch`
- `image_backend_response_untrusted`

所有失败结果都必须携带 `backend_call_count`；调用前失败为 0，调用后失败为 1。成功时 `backend_call_count=1`、`image_generation_count=1`；store lookup、binary read、upload、public URL、retry、fallback、executor、confirmation、writeback 和 replay 全部为 0。

## 测试矩阵

- 相同输入、profile 与 request id 生成完全相同的 backend request。
- UTF-8 byte 边界、最大尺寸 / 像素、count、steps、guidance、列表和标识符边界。
- 未知字段、重复 trace、profile mismatch、requires confirmation、medium / high risk 和敏感材料在调用前拒绝。
- 成功 client 恰好调用一次并返回 canonical citation / metadata reference。
- timeout、unavailable、任意 client 异常均不 retry。
- artifact schema、public URL、binary / provider raw 字段、lineage、generation、title / purpose、safety、provenance、hash、MIME 与 dimensions 漂移全部失败关闭。
- 输入、backend request、artifact 和结果均不携带 credential、endpoint、像素或 provider raw response。

## 完成定义

- 批次 A 领域运行时和相邻测试全部通过。
- `services/runtime/tests` 已进入快速仓库基线，不新增一次性 checker。
- 既有 image mapper / consumer / response builder 测试继续通过。
- 文档状态推进为 `image_adapter_controlled_invocation_artifact_return_dev_test_v1_batch_a_completed_batch_b_review_required`。
- 下一批必须重新选择单一实现方向；不得从本任务自动打开真实 backend、store、reader、public URL、API 或 Web。

## 完成记录

- `services/runtime/image_generation_adapter.py` 已实现纯 backend request compiler、调用前预算 / 敏感材料 / profile / safety gate、单次注入式 client 调用、稳定失败语义和 artifact transport observation 重验。
- 11 项相邻测试覆盖确定性、UTF-8 边界、尺寸 / 像素 / count、参数与集合预算、未知字段、标识符 / locale、confirmation、medium / high risk、profile mismatch、timeout / unavailable、artifact schema / lineage / safety / provenance 和 hash / MIME / dimensions 漂移。
- `services/runtime/tests` 已进入现有 `check-repo` Python unittest 入口；没有新增一次性 checker。
- 本批没有创建 backend client、credential / endpoint / model-dir resolver、HTTP、Gateway、store、binary reader、public URL、Web 或生产能力。

## 批次 B：本机私有 artifact storage

### 单一方向

批次 B 只打开 `local_private_artifact_storage` owner，并按稳定职责拆出 binary inspection 与 private storage 两个模块，实现 content-addressed blob、不可变 artifact ref、metadata-only lookup 与显式授权 binary reader。它不改变 Batch A 的 backend client envelope，也不并行打开真实 backend、profile / credential 配置、HTTP / Gateway 或 Web。

### 允许

- 新增 `services/runtime/image_artifact_binary_inspection.py` 与 `services/runtime/image_artifact_private_storage.py`。
- 扩展 `services/runtime/tests`，使用临时目录验证真实本机文件写入、lookup、读取与篡改失败关闭。
- 复用 `image-generation-artifact.schema.json` 与既有 artifact mapper，不新增第二套 artifact contract。

### 固定存储与读取规则

1. store root 必须由调用侧显式注入绝对路径；不读取环境变量，不自动选择存储实现。
2. payload 只接受有界 bytes-like 输入，不接受 base64 字符串或 provider response envelope。
3. 写入前重算 sha256，并从二进制容器识别 PNG / JPEG / WebP MIME 与 dimensions；所有值必须与 strict artifact metadata 精确一致。
4. blob 使用 sha256 content-addressed 相对路径；artifact ref 使用 `artifact_id` 独立绑定 blob、URI、digest、MIME、dimensions、format 与 size。
5. blob 与 ref 均不可覆盖；相同内容 / 相同 ref 幂等成功，任何既有内容或绑定漂移失败关闭。
6. lookup 只读取 ref metadata 和文件状态，不读取 blob 内容，不返回绝对路径或公开 URL。
7. binary reader 默认拒绝；显式授权后只读取一次，并在内部 consumer 执行前重新校验 hash / MIME / dimensions。
8. consumer 最多调用一次；reader result 不携带 bytes、base64、文件路径、provider raw payload、public URL 或 signed URL。

### 稳定失败语义

- `image_artifact_private_store_root_invalid`
- `image_artifact_binary_invalid`
- `image_artifact_binary_too_large`
- `image_artifact_store_unavailable`
- `image_artifact_store_conflict`
- `image_artifact_store_integrity_failure`
- `image_artifact_store_reference_missing`
- `image_artifact_binary_read_forbidden`
- `image_artifact_binary_reader_unavailable`
- `image_artifact_binary_consumer_failed`
- 既有 mapper 的 metadata、hash、MIME、dimensions、安全与 provenance 失败码继续复用。

### 非目标

- 不实现 delete / overwrite、artifact upload、public / signed URL、production object store 或跨主机复制。
- 不接真实生图 backend、credential / endpoint / model-dir resolver、moderation provider 或模型下载。
- 不新增 schema、API、API key、repository selector、migration、Session、Run、Gateway、Web 或浏览器链。
- 不把本机私有 store 写成 production storage 或图片生成能力。

### 验收

- 覆盖 PNG / JPEG / WebP header observation、原子不可变写入、幂等、ref lookup、默认拒绝读取和显式内部消费。
- 覆盖 payload/type/size、schema、URI、safety、provenance、hash、MIME、dimensions、format、ref、symlink 与磁盘篡改负向路径。
- 覆盖 consumer 异常脱敏、零 retry / fallback / upload / public URL / production write / backend 调用。
- 通过相邻 Python 单元测试、既有三项 Image runtime 检查、fast / full 仓库门禁与差异卫生检查。

### 完成记录

- PNG / JPEG / WebP 容器观察已独立收口，使用 sha256 与容器 header 生成稳定 observation，并拒绝 CRC、marker、RIFF size、dimensions 或格式漂移。
- 私有 store 已实现 content-addressed blob 与 artifact ref 两级不可变写入；相同 artifact 幂等，并发相同写入收敛，artifact id 重绑定、symlink、ref 或 blob 篡改失败关闭。
- metadata-only lookup 不读取 blob，不返回绝对路径；显式 binary reader 每次读取都重新验证，内部 stream 在 consumer 返回后关闭。
- 15 项 storage 测试与既有 11 项 adapter 测试共 26 项，均由 `services/runtime/tests` 聚合入口执行。
- 状态推进为 `image_adapter_controlled_invocation_artifact_return_dev_test_v1_batch_b_completed_batch_c_review_required`；下一批不得同时打开具体 backend client 与 profile / credential 配置。

## 批次 C：reference-only backend profile 配置

### 单一方向

批次 C 只打开 `image_backend_profile_configuration` owner。配置 source 负责 profile identity、开发测试环境、backend identity、runtime reference、credential policy 与 timeout；具体 backend client、reference resolver、HTTP / Gateway / Web 和生产配置继续关闭。

### 允许

- 新增 `contracts/image-backend-profile-source.schema.json` 和一个基础 fixture。
- 新增 `services/runtime/image_backend_profile_configuration.py` 与相邻单元测试。
- 将批次 A 的 `ImageBackendProfile` 收口为 profile compiler 产物，并让 adapter 从 profile 读取 timeout。
- 复用既有 `image_generation_backend_request` 的 backend 字段，不创建第二套 backend request。

### 固定配置规则

1. source 必须通过 strict schema、有效 UTF-8 和 `16 KiB` canonical JSON byte budget。
2. 环境只允许 `development | test`；production、staging 与未知环境失败关闭。
3. `remote_https` 必须绑定同环境 `endpoint_ref`、`credential_requirement=required` 与 `secret_ref`，且 `model_dir_ref=null`。
4. `local_model` 必须绑定同环境 `model_dir_ref`、`credential_requirement=not_required`，且 `endpoint_ref / secret_ref=null`。
5. 所有引用必须采用 `ref:radishmind/<environment>/image-backends/<kind>/<key>`；跨环境、错 kind、路径穿越、raw URL 与绝对路径均拒绝。
6. profile source 禁止 credential value、token、authorization / headers、cookie、endpoint URL、DSN、model path、环境变量、自由 system prompt、provider config 与 runtime config。
7. profile compiler 生成确定性 `profile_digest`；源字段顺序不影响 digest，identity / binding / limit 任一变化都会改变 digest。
8. adapter 只接受可重算 digest 的 enabled profile；profile timeout 是 client 调用的唯一 timeout source。

### 稳定失败语义

- `image_backend_profile_source_invalid`
- `image_backend_profile_source_budget_exceeded`
- `image_backend_profile_sensitive_material_rejected`
- `image_backend_profile_environment_forbidden`
- `image_backend_profile_binding_invalid`
- `image_backend_profile_digest_drift`
- adapter 侧新增 `image_backend_profile_invalid`，并继续复用 missing / mismatch 失败码。

### 非目标

- 不实现 endpoint / credential / model-dir reference resolver，不读取 secret、环境变量或本机模型目录。
- 不创建具体 backend client，不连接网络，不加载模型，不生成图片。
- 不新增 repository、selector、migration、应用配置、API key、HTTP、Gateway、Session、Run 或 Web。
- 不启用 production environment、retry、fallback、public URL、upload 或 production storage。

### 验收与完成记录

- 10 项 profile 配置测试覆盖确定性、远程 / 本机模式、strict schema、环境门禁、敏感材料、互斥绑定、引用作用域、timeout / source byte budget、digest drift 与输入不变性。
- 批次 A 相邻测试更新为编译 profile，新增 digest drift 调用前拒绝；12 项 adapter 测试全部通过。
- 15 项 storage 测试继续通过，`services/runtime/tests` 总数为 37；五项既有 Image 检查保持兼容。
- 状态推进为 `image_adapter_controlled_invocation_artifact_return_dev_test_v1_batch_c_completed_batch_d_review_required`；批次 D 只允许先评审一个具体 backend client。

## 批次 D：test-only contract fixture backend client

### 单一方向

批次 D 只打开 `image_backend_contract_fixture_client` owner。该 client 以调用侧注入的真实图片 fixture 承担离线 contract smoke，不解析批次 C 的 endpoint / credential / model-dir reference，不连接网络或加载模型。

### 根因修正

批次 A 的 `ImageBackendInvocationResult` 要求 backend client 返回完整 `image_generation_artifact`，但 canonical title、purpose、safety 与 provenance 属于 RadishMind intent / adapter 所有权，且未完整进入 backend request。批次 D 将 client 输出缩减为 artifact identity、UTC 时间与二进制 observation，由 adapter 唯一构造 canonical artifact metadata，避免 backend 成为第二套 artifact 真相源。

### 允许

- Profile source 新增仅限 `test` 的 `contract_fixture` runtime mode；credential 为 `not_required`，三类 reference 必须全部为 null。
- 新增 `services/runtime/image_backend_contract_fixture_client.py`、一份 profile fixture 和相邻测试。
- client 创建时检查调用侧注入的 PNG / JPEG / WebP fixture，调用时精确匹配 canonical backend request、profile、timeout、output、input 和 safety。
- adapter 从可信 observation、intent 与 backend request 构造既有 `image_generation_artifact`，不新增 artifact schema。

### 固定 client 规则

1. 只接受 digest 可重算、enabled、`environment=test`、`runtime_mode=contract_fixture` 的编译 profile。
2. fixture 最大 `32 MiB`，dimensions 最大 `2048 x 2048`，像素最大 `4,194,304`；hash、MIME 和 dimensions 必须来自实际容器检查。
3. backend request 最大 `64 KiB` canonical JSON，必须通过既有 strict schema、UTF-8 与敏感材料检查。
4. backend identity 与 profile exact match；timeout 必须等于 profile timeout。
5. 首版只支持单张 text-to-image fixture，reference / edit / mask inputs 必须为空，safety 必须是 low-risk、无需确认且已允许 backend。
6. artifact id 由 canonical backend request digest 确定性派生；created_at 只能是显式注入的 UTC 秒级时间。
7. client result 禁止 bytes、base64、provider raw response、endpoint、credential、title、purpose、safety 与 provenance。
8. fixture 成功只记录一次 backend handoff，`image_generation_count=0`；adapter 失败结果继续不返回 backend request 或 artifact metadata，不 retry / fallback。

### 稳定失败语义

- `image_backend_fixture_profile_invalid`
- `image_backend_fixture_binary_invalid`
- `image_backend_fixture_request_invalid`
- `image_backend_fixture_request_mismatch`
- 通过 adapter 调用时，上述内部失败统一脱敏为 `image_backend_response_untrusted`。

### 非目标

- 不持久化或返回 fixture bytes，不新增 binary delivery / store coordinator。
- 不实现 remote HTTP client、local model client、endpoint / credential / model-dir resolver。
- 不连接网络，不读取 secret、环境变量或模型目录，不下载模型，不生成图片。
- 不新增 API、Gateway、Web、repository、migration、upload、public URL、retry / fallback 或生产声明。

### 验收与完成记录

- 7 项具体 client 测试覆盖 PNG / JPEG / WebP observation、确定性、canonical artifact owner、profile / fixture / timestamp、strict request、binding drift、零 generation count 与 adapter 脱敏失败。
- Profile 测试增至 11 项，覆盖 `contract_fixture` 仅限 test 和零 reference；adapter 12 项继续覆盖调用前安全与调用后不可信结果。
- 15 项 storage 测试保持兼容，`services/runtime/tests` 总数为 45；既有五项 Image contract / mapper / consumer / builder 检查继续复用。
- 状态推进为 `image_adapter_controlled_invocation_artifact_return_dev_test_v1_batch_d_completed_batch_e_review_required`；批次 E 只先评审 fixture binary 的单次交付与既有本机私有 store 协调边界。

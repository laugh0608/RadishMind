# Image Adapter 受控调用与 artifact 返回（开发 / 测试态）v1 实施任务卡

更新时间：2026-07-25

状态：`image_adapter_controlled_invocation_artifact_return_dev_test_v1_batch_b_completed_batch_c_review_required`

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

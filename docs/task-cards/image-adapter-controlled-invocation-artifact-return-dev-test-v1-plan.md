# Image Adapter 受控调用与 artifact 返回（开发 / 测试态）v1 实施任务卡

更新时间：2026-07-25

状态：`image_adapter_controlled_invocation_artifact_return_dev_test_v1_batch_a_completed_batch_b_review_required`

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

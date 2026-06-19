# Image Artifact Response Consumer Runtime Implementation v1 计划

## 状态

- 切片：`image-artifact-response-consumer-runtime-implementation-v1`
- 轨道：`Image Path`
- 状态：`image_artifact_response_consumer_runtime_implemented`

## 目标

在 `image-artifact-response-consumer-implementation-v1` 已固定任务边界后，落地 metadata-only response consumer runtime code。

本切片只实现 `services/runtime/image_artifact_response_consumer.py` 中的 `apply_image_artifact_reference_to_response`：消费已有 `CopilotResponse` draft 与 `ImageArtifactMappingResult`，成功时把 mapper 返回的 artifact citation 合并进 `CopilotResponse.citations`，并把 `metadata_reference` 作为内部 handoff 随 result 返回。

本切片不修改现有 response builder，不改 `CopilotResponse` schema，不接 artifact store、binary reader、public URL resolver、backend adapter、gateway 或 platform HTTP route。

## 输入

- [`Image Artifact Response Consumer Implementation` v1 计划](image-artifact-response-consumer-implementation-v1-plan.md)
- [`Image Artifact Response Consumer Implementation Readiness` v1 计划](image-artifact-response-consumer-implementation-readiness-v1-plan.md)
- [`Image Artifact Runtime Mapper Runtime Implementation` v1 计划](image-artifact-runtime-mapper-runtime-implementation-v1-plan.md)
- `services/runtime/image_artifact_runtime_mapper.py`
- `contracts/copilot-response.schema.json`

## 范围

1. Runtime consumer
   - 新增 `services/runtime/image_artifact_response_consumer.py`。
   - 新增 `ImageArtifactResponseConsumerResult`。
   - 新增 `apply_image_artifact_reference_to_response`。
   - 新增 `response_consumer_side_effect_counters`。

2. 成功路径
   - mapper `ok=true` 且 citation schema 合法时，把 citation append 到 `CopilotResponse.citations`。
   - 保留已有 citations 顺序。
   - 不原地修改输入 response draft 或 mapper citation。
   - `metadata_reference` 只通过 result 的内部字段返回，不写入 `response_document`。

3. 失败路径
   - mapper failure 返回 `image_artifact_mapper_failed`。
   - citation id 冲突返回 `image_artifact_citation_id_conflict`。
   - citation shape 不符合 artifact citation 要求返回 `image_artifact_citation_schema_invalid`。
   - response document 或 metadata reference 试图暴露 public URL、signed URL、binary payload、provider raw data 或内部 handoff 字段时返回 `image_artifact_metadata_reference_leak`。

4. 验证
   - 新增 `scripts/checks/fixtures/image-artifact-response-consumer-runtime-implementation-v1.json`。
   - 新增 `scripts/check-image-artifact-response-consumer-runtime-implementation-v1.py`。
   - checker 接入 `./scripts/check-repo.sh --fast`，并在 `check-image-artifact-response-consumer-implementation-v1.py` 之后运行。

## 停止线

- 不修改 `services/runtime/inference_response.py`。
- 不修改 `services/runtime/inference_support.py`。
- 不修改 `services/gateway/copilot_gateway.py`。
- 不修改 `services/platform/internal/httpapi/responses.go`。
- 不创建 artifact store、binary reader、public URL resolver 或 backend adapter。
- 不读取 artifact 二进制、不上传 artifact、不调用真实生图 backend、不生成图片。
- 不进入 executor、confirmation、writeback 或 replay。

## 验收

- response consumer success case 产出的 `response_document` 通过 `contracts/copilot-response.schema.json`。
- failure case 返回明确 failure code，且不产生成功 citation。
- `metadata_reference` 不进入 public response surface。
- side-effect counters 保持为零。
- 前置 readiness / implementation checker 在 runtime module 已存在后仍通过，证明只有本切片放行该模块，现有 response builder 与平台路径仍未接线。

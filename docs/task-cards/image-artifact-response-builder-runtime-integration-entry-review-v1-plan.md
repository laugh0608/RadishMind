# Image Artifact Response Builder Runtime Integration Entry Review v1 计划

## 状态

- 切片：`image-artifact-response-builder-runtime-integration-entry-review-v1`
- 轨道：`Image Path`
- 状态：`image_artifact_response_builder_runtime_integration_entry_review_defined`

## 目标

在 `image-artifact-response-builder-integration-v1` 已固定未来 hook placement、request metadata discovery、post-merge validation 和 failure propagation 后，审查是否允许下一步创建单一 response builder runtime integration implementation 任务卡。

本切片只做 runtime integration entry review：确认前置证据、单一实现方向、未来任务卡范围、停止线和 checker 证据。当前不修改 `services/runtime/inference_response.py`，不改 `CopilotResponse` schema，不实现 artifact store、binary reader、public URL resolver、backend adapter，也不调用真实生图 backend。

## 输入

- [`Image Artifact Response Builder Integration` v1 计划](image-artifact-response-builder-integration-v1-plan.md)
- [`Image Artifact Response Builder Integration Entry Review` v1 计划](image-artifact-response-builder-integration-entry-review-v1-plan.md)
- [`Image Artifact Response Consumer Runtime Implementation` v1 计划](image-artifact-response-consumer-runtime-implementation-v1-plan.md)
- [`Image Artifact Runtime Mapper Runtime Implementation` v1 计划](image-artifact-runtime-mapper-runtime-implementation-v1-plan.md)
- `services/runtime/inference_response.py`
- `services/runtime/image_artifact_response_consumer.py`
- `services/runtime/image_artifact_runtime_mapper.py`
- `contracts/copilot-request.schema.json`
- `contracts/copilot-response.schema.json`
- `contracts/image-generation-artifact.schema.json`

## 范围

1. 前置证据复核
   - `image-artifact-runtime-mapper-runtime-implementation-v1` 必须保持 `image_artifact_runtime_mapper_runtime_implemented`。
   - `image-artifact-response-consumer-runtime-implementation-v1` 必须保持 `image_artifact_response_consumer_runtime_implemented`。
   - `image-artifact-response-builder-integration-v1` 必须保持 `image_artifact_response_builder_integration_task_card_defined`。
   - mapper / consumer 组合必须仍能从 request-side `image_generation_artifact` metadata 产出 schema-valid `CopilotResponse`。

2. 单一实现方向
   - 下一步只允许创建 `image-artifact-response-builder-runtime-integration-implementation-v1` 任务卡。
   - 未来 runtime code 唯一候选仍是 `services/runtime/inference_response.py#coerce_response_document`。
   - 未来 hook placement 继续固定为 response top-level filtering 之后、现有 `validate_response_document(coerced)` 之前。
   - 未来实现不得改 `coerce_response_document(document, copilot_request, raw_text)` 的调用形状；request metadata discovery 必须使用已有 `copilot_request` 参数。

3. 未来任务卡必须定义的实现边界
   - 从 `copilot_request.artifacts[*].metadata.image_generation_artifact` 按 request artifact 顺序发现 metadata。
   - 缺失 metadata 时保持 no-op，不把普通文本响应升级为失败。
   - 发现 metadata 后先校验 `contracts/image-generation-artifact.schema.json`，再调用 `map_image_artifact_to_response_reference` 和 `apply_image_artifact_reference_to_response`。
   - 合并后必须复用现有 `validate_response_document` 做 post-merge `CopilotResponse` schema validation。
   - mapper failure、consumer failure 和 post-merge schema validation failure 必须 fail closed，并保留内部 failure code。

4. 停止线
   - 本切片不创建 implementation task card，不实现 runtime code，不把 response builder 标记为 connected。
   - store、binary reader、public URL resolver、signed URL resolver、backend adapter、artifact upload、production storage、真实 backend call 和图片生成继续 deferred。
   - gateway、platform northbound route、`inference_support.py#build_citations` 不承接 image artifact metadata merge。

5. 验证
   - 新增 `scripts/checks/fixtures/image-artifact-response-builder-runtime-integration-entry-review-v1.json`。
   - 新增 `scripts/check-image-artifact-response-builder-runtime-integration-entry-review-v1.py`。
   - checker 接入 `./scripts/check-repo.sh --fast`，并在 `check-image-artifact-response-builder-integration-v1.py` 之后运行。

## 停止线

- 不修改 `services/runtime/inference_response.py`。
- 不修改 `services/runtime/inference_support.py`。
- 不修改 `services/gateway/copilot_gateway.py`。
- 不修改 `services/platform/internal/httpapi/responses.go`。
- 不改 `CopilotResponse` schema。
- 不创建 runtime integration implementation 任务卡以外的后续任务。
- 不创建 artifact store、binary reader、public URL resolver、signed URL resolver 或 backend adapter。
- 不读取 artifact 二进制、不上传 artifact、不调用真实生图 backend、不生成图片。
- 不进入 executor、confirmation、writeback 或 replay。

## 验收

- task card、fixture 和 checker 明确结论：下一步只允许创建单一 response builder runtime integration implementation 任务卡。
- checker 证明 exact hook placement、request metadata discovery input、post-merge validation 和 failure propagation 证据仍稳定。
- checker 证明当前 `inference_response.py`、`inference_support.py`、gateway 和 platform route 仍未接入 image artifact response consumer。
- checker 证明 mapper / consumer 组合无 side effects，且没有创建 store、reader、public URL、backend adapter 或 UI artifact。

# Image Artifact Response Builder Runtime Integration Implementation v1 任务卡

更新时间：2026-06-14

## 任务标识

- 切片：`image-artifact-response-builder-runtime-integration-implementation-v1`
- 轨道：`Image Path`
- 状态：`image_artifact_response_builder_runtime_integration_implementation_task_card_defined`

## 目标

在 `image-artifact-response-builder-runtime-integration-entry-review-v1` 已确认只允许单一 response builder runtime integration implementation 方向后，创建后续 runtime code 的实现任务卡、fixture 和 checker。

本切片只定义未来实现 `services/runtime/inference_response.py#coerce_response_document` 接线时必须满足的实现边界、测试矩阵、失败传播、no side effects 和停止线。当前不修改 `services/runtime/inference_response.py`，不改 `CopilotResponse` schema，不创建 artifact store、binary reader、public URL resolver 或 backend adapter，也不调用真实生图 backend。

## 输入事实源

- [图片生成契约](../contracts/image-generation.md)
- [`Image Artifact Response Builder Runtime Integration Entry Review` v1 计划](image-artifact-response-builder-runtime-integration-entry-review-v1-plan.md)
- [`Image Artifact Response Builder Integration` v1 计划](image-artifact-response-builder-integration-v1-plan.md)
- [`Image Artifact Response Consumer Runtime Implementation` v1 计划](image-artifact-response-consumer-runtime-implementation-v1-plan.md)
- [`Image Artifact Runtime Mapper Runtime Implementation` v1 计划](image-artifact-runtime-mapper-runtime-implementation-v1-plan.md)
- `services/runtime/inference_response.py`
- `services/runtime/image_artifact_response_consumer.py`
- `services/runtime/image_artifact_runtime_mapper.py`
- `contracts/copilot-request.schema.json`
- `contracts/copilot-response.schema.json`
- `contracts/image-generation-artifact.schema.json`

## 后续 runtime code 边界

1. 允许的实现落点
   - 未来 runtime code 只允许修改 `services/runtime/inference_response.py`。
   - 未来接线函数固定为 `coerce_response_document(document, copilot_request, raw_text)`，不得改变调用签名。
   - hook 固定在 `coerced = {key: value for key, value in coerced.items() if key in RESPONSE_TOP_LEVEL_KEYS}` 之后、现有 `validate_response_document(coerced)` 之前。

2. request metadata discovery
   - 未来实现只能从 `copilot_request.artifacts[*].metadata.image_generation_artifact` 按 request artifact 顺序发现 metadata。
   - 缺失 metadata 时必须 no-op，保留原 response builder 行为。
   - discovery 不得读取 response draft、`raw_text`、环境变量、artifact store、binary reader、public URL、signed URL 或 backend response。
   - 发现 metadata 后必须先通过 `contracts/image-generation-artifact.schema.json` 校验，再进入 mapper。

3. mapper / consumer / validation pipeline
   - schema-valid metadata 才能进入 `map_image_artifact_to_response_reference`。
   - mapper 成功后才允许进入 `apply_image_artifact_reference_to_response`。
   - consumer 成功后必须对合并后的 `response_document` 执行现有 `validate_response_document`。
   - 多个 metadata 必须按 request artifact 顺序依次合并，任一失败都必须 fail closed，不返回部分成功 response。

4. failure propagation
   - metadata schema invalid 使用 `image_artifact_metadata_schema_invalid`。
   - mapper failure 传播为 `image_artifact_mapper_failed`，并保留 mapper 原始 failure code 供内部审计。
   - consumer failure 保留 `image_artifact_citation_id_conflict`、`image_artifact_citation_schema_invalid` 或 `image_artifact_metadata_reference_leak`。
   - post-merge schema validation failure 使用 `image_artifact_response_schema_invalid`。
   - failure path 不触发 retry、fallback execution、backend call、artifact upload、binary read、store lookup 或 public URL resolution。

## 验收口径

- `scripts/checks/fixtures/image-artifact-response-builder-runtime-integration-implementation-v1.json` 固定 implementation task card boundary、future runtime hook、metadata discovery、merge pipeline、failure propagation、runtime test plan、forbidden artifact matrix 和 no side effects。
- `scripts/check-image-artifact-response-builder-runtime-integration-implementation-v1.py` 进入 `./scripts/check-repo.sh --fast`，并在 `check-image-artifact-response-builder-runtime-integration-entry-review-v1.py` 之后运行。
- checker 必须证明前置 entry review、integration task card、runtime mapper 和 response consumer 状态稳定。
- checker 必须用 request-side metadata 模拟 single artifact、multiple artifacts ordering、missing metadata no-op、schema invalid、mapper failure、consumer failure 和 post-merge schema invalid。
- checker 必须证明本切片未修改 `inference_response.py`，未改 response schema，未创建 store / reader / public URL / backend adapter，未调用真实 backend。

## 非目标

- 不修改 `services/runtime/inference_response.py`。
- 不修改 `services/runtime/inference_support.py`、gateway 或 platform response route。
- 不改 `CopilotResponse` schema，不新增 response 顶层 artifact 字段。
- 不创建 artifact store、binary reader、public URL resolver、signed URL resolver、backend adapter、artifact upload 或 production storage。
- 不读取 artifact 二进制，不调用真实生图 backend，不生成图片。
- 不新增 UI，不启动开发服务器，不进入 executor、confirmation、writeback 或 replay。

## 停止线

- 不能把本任务卡写成 response builder runtime integration 已实现。
- 后续 runtime code 只能接入 metadata-only mapper / consumer 链路，不得绕过 schema validation 或 fail-closed 结果生成成功 citation。
- `metadata_reference` 仍只允许作为内部 handoff，不得进入 public `CopilotResponse`。
- store / reader / public URL / backend adapter 仍需分别独立任务卡、fixture、checker 和验证链路。

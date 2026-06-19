# Image Artifact Response Builder Integration v1 计划

## 状态

- 切片：`image-artifact-response-builder-integration-v1`
- 轨道：`Image Path`
- 状态：`image_artifact_response_builder_integration_task_card_defined`

## 目标

在 `image-artifact-response-builder-integration-entry-review-v1` 已选择未来接入点后，固定 response builder integration 的任务卡边界。

本切片只定义未来接入 `services/runtime/inference_response.py#coerce_response_document` 时必须满足的 exact hook placement、request artifact metadata discovery input、post-merge `CopilotResponse` schema validation、failure propagation 和 no side effects。当前不修改 response builder，不改 `CopilotResponse` schema，不实现 artifact store、binary reader、public URL resolver、backend adapter，也不调用真实生图 backend。

## 输入

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

1. Exact hook placement
   - 未来接入点固定为 `coerce_response_document` 内 `coerced = {key: value for key, value in coerced.items() if key in RESPONSE_TOP_LEVEL_KEYS}` 之后。
   - 未来 hook 必须位于现有 `validate_response_document(coerced)` 之前。
   - hook 必须在 task-specific canonicalization、默认字段补齐、citation normalization 和 response top-level filtering 之后运行，避免后续 normalization 覆盖 artifact citation 或隐藏 metadata leak。

2. Request artifact metadata discovery input
   - v1 只允许从 `copilot_request.artifacts[*].metadata.image_generation_artifact` 发现 `image_generation_artifact` metadata。
   - discovery 必须保持 request artifact 顺序；缺失该 metadata 时保持 no-op，不把普通文本响应升级成失败。
   - discovery 不得读取 response draft、`raw_text`、环境变量、artifact store、binary reader、public URL、signed URL 或 backend response。
   - 发现的 metadata 必须先通过 `contracts/image-generation-artifact.schema.json`，再交给 `map_image_artifact_to_response_reference`。

3. Post-merge schema validation
   - mapper 成功后才允许调用 `apply_image_artifact_reference_to_response`。
   - consumer 成功后必须对合并后的 `response_document` 再执行现有 `contracts/copilot-response.schema.json` / `validate_response_document` 校验。
   - post-merge validation 失败时必须 fail closed，不得静默丢弃 artifact citation 后返回成功响应。
   - `metadata_reference` 只允许保留在内部 handoff / audit detail，不得进入 public `CopilotResponse`。

4. Failure propagation
   - mapper failure 必须作为 `image_artifact_mapper_failed` 传播到 response builder failure path，并保留原始 mapper failure detail 供内部审计。
   - consumer failure 必须保留 `image_artifact_citation_id_conflict`、`image_artifact_citation_schema_invalid` 或 `image_artifact_metadata_reference_leak`，不得转换成成功响应。
   - post-merge schema validation failure 必须使用独立失败边界，不触发 retry、fallback execution、backend call 或 artifact upload。

5. 验证
   - 新增 `scripts/checks/fixtures/image-artifact-response-builder-integration-v1.json`。
   - 新增 `scripts/check-image-artifact-response-builder-integration-v1.py`。
   - checker 接入 `./scripts/check-repo.sh --fast`，并在 `check-image-artifact-response-builder-integration-entry-review-v1.py` 之后运行。

## 停止线

- 不修改 `services/runtime/inference_response.py`。
- 不修改 `services/runtime/inference_support.py`。
- 不修改 `services/gateway/copilot_gateway.py`。
- 不修改 `services/platform/internal/httpapi/responses.go`。
- 不改 `CopilotResponse` schema。
- 不创建 artifact store、binary reader、public URL resolver 或 backend adapter。
- 不读取 artifact 二进制、不上传 artifact、不调用真实生图 backend、不生成图片。
- 不进入 executor、confirmation、writeback 或 replay。

## 验收

- task card、fixture 和 checker 明确定义未来 hook placement、discovery input、post-merge validation、failure propagation 和 no side effects。
- checker 证明 request-side `image_generation_artifact` metadata 可以经 mapper / consumer 合并成 schema-valid response。
- checker 证明缺失 metadata 保持 no-op，mapper / consumer / post-merge validation failure 均 fail closed。
- checker 证明现有 response builder、gateway 和 platform route 仍未接入 image artifact response consumer。

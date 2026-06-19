# Image Artifact Response Builder Integration Entry Review v1 计划

## 状态

- 切片：`image-artifact-response-builder-integration-entry-review-v1`
- 轨道：`Image Path`
- 状态：`image_artifact_response_builder_integration_entry_review_defined`

## 目标

在 `image-artifact-response-consumer-runtime-implementation-v1` 已落地 metadata-only response consumer runtime 后，审查未来是否允许把该 consumer 接入现有 response builder，以及应优先选择哪个接入点。

本切片只做 entry review：固定候选接入点、准入 gate、禁止接线项、failure propagation 和下一步任务卡条件。当前不修改 `services/runtime/inference_response.py`、`services/runtime/inference_support.py`、gateway 或 platform route。

## 输入

- [`Image Artifact Response Consumer Runtime Implementation` v1 计划](image-artifact-response-consumer-runtime-implementation-v1-plan.md)
- [`Image Artifact Response Consumer Implementation` v1 计划](image-artifact-response-consumer-implementation-v1-plan.md)
- [`Image Artifact Runtime Mapper Runtime Implementation` v1 计划](image-artifact-runtime-mapper-runtime-implementation-v1-plan.md)
- `services/runtime/image_artifact_response_consumer.py`
- `services/runtime/image_artifact_runtime_mapper.py`
- `services/runtime/inference_response.py`
- `contracts/copilot-response.schema.json`

## 范围

1. 候选接入点评审
   - 优先候选：`services/runtime/inference_response.py#coerce_response_document` 的 response normalization / schema validation 边界。
   - 不选 `services/runtime/inference_support.py#build_citations`，因为它负责请求上下文 citation，不负责生成 artifact 的 runtime handoff。
   - 不选 `services/gateway/copilot_gateway.py#handle_copilot_request`，因为 gateway 只处理 envelope 和 provider route，不应承载 artifact metadata merge。
   - 不选 `services/platform/internal/httpapi/responses.go#buildOpenAIResponsesResponse`，因为 northbound response route 只做协议转换，不应接 Python runtime artifact handoff。

2. 准入 gate
   - mapper 必须返回 `ok=true`、artifact citation 和内部 `metadata_reference`。
   - consumer 必须返回 `ok=true`，并且输出 `response_document` 通过 `contracts/copilot-response.schema.json`。
   - citation id conflict、citation schema invalid、mapper failure 和 metadata reference leak 必须 fail closed。
   - `metadata_reference` 不得进入 public response surface。

3. 验证
   - 新增 `scripts/checks/fixtures/image-artifact-response-builder-integration-entry-review-v1.json`。
   - 新增 `scripts/check-image-artifact-response-builder-integration-entry-review-v1.py`。
   - checker 接入 `./scripts/check-repo.sh --fast`，并在 `check-image-artifact-response-consumer-runtime-implementation-v1.py` 之后运行。

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

- entry candidate matrix 明确选择 Python response normalization / schema validation 边界作为未来接入候选。
- checker 证明 mapper + consumer 的 metadata-only 成功链路可产出 schema-valid response document。
- checker 证明 failure propagation、metadata handoff 和 no side effects 仍成立。
- checker 证明现有 response builder、gateway 和 platform route 没有接入 consumer runtime。

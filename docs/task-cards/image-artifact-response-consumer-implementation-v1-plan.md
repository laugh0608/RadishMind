# Image Artifact Response Consumer Implementation v1 任务卡

更新时间：2026-06-14

## 任务标识

- 切片：`image-artifact-response-consumer-implementation-v1`
- 轨道：`Image Path`
- 状态：`image_artifact_response_consumer_implementation_task_card_defined`

## 目标

在 `image-artifact-response-consumer-implementation-readiness-v1` 已固定未来 consumer 准入后，创建 metadata-only response consumer 的实现任务卡、fixture 和 checker。

本切片只定义后续 runtime code 的职责边界：未来 `services/runtime/image_artifact_response_consumer.py` 中的 `apply_image_artifact_reference_to_response` 只能消费已有 `CopilotResponse` draft 和 `ImageArtifactMappingResult`，把 mapper 成功结果合并进 `CopilotResponse.citations` 的 `kind=artifact` citation，并把 `metadata_reference` 继续保留为内部 handoff。

当前不创建 response consumer 模块，不修改 response builder，不改 `CopilotResponse` schema，也不接 artifact store、binary reader、public URL resolver 或 backend adapter。

## 输入事实源

- [图片生成契约](../contracts/image-generation.md)
- [`Image Artifact Response Consumer Implementation Readiness` v1 计划](image-artifact-response-consumer-implementation-readiness-v1-plan.md)
- [`Image Artifact Runtime Mapper Response Consumer Integration Review` v1 计划](image-artifact-runtime-mapper-response-consumer-integration-review-v1-plan.md)
- [`Image Artifact Runtime Mapper Runtime Implementation` v1 计划](image-artifact-runtime-mapper-runtime-implementation-v1-plan.md)
- `scripts/checks/fixtures/image-artifact-response-consumer-implementation-v1.json`
- `scripts/checks/fixtures/image-artifact-response-consumer-implementation-readiness-v1.json`
- `services/runtime/image_artifact_runtime_mapper.py`
- `contracts/copilot-response.schema.json`

## 后续 runtime code 边界

1. future module
   - 未来实现文件：`services/runtime/image_artifact_response_consumer.py`
   - 未来函数：`apply_image_artifact_reference_to_response`
   - 未来结果类型：`ImageArtifactResponseConsumerResult`

2. input / output
   - 输入为已有 `CopilotResponse` draft 和 `ImageArtifactMappingResult`。
   - 成功时只合并 `kind=artifact` citation 到 `CopilotResponse.citations`。
   - `metadata_reference` 不进入 `CopilotResponse` 顶层，不暴露 public URL、signed URL、binary payload 或 provider raw dump。

3. failure policy
   - mapper failure 必须 fail closed，response draft 保持不变。
   - citation id conflict、citation schema invalid 和 metadata reference leak 必须返回明确 failure code。
   - duplicate citation 不允许静默覆盖，schema 不匹配不允许降级为普通文本引用。

## 验收口径

- `scripts/checks/fixtures/image-artifact-response-consumer-implementation-v1.json` 固定 implementation task card boundary、future module、function contract、input / output contract、failure taxonomy、runtime test plan、forbidden artifact matrix 和 no side effects。
- `scripts/check-image-artifact-response-consumer-implementation-v1.py` 进入 `./scripts/check-repo.sh --fast`，并在 `check-image-artifact-response-consumer-implementation-readiness-v1.py` 之后运行。
- checker 必须跨读 readiness fixture、runtime mapper fixture、`CopilotResponse` schema 和 artifact fixture，确认后续 runtime code 可以覆盖 success / fail-closed case。
- checker 必须确认本切片未创建 response consumer 模块，未修改 response builder，未改 response schema，未接 store / reader / public URL / backend。
- current focus、能力矩阵、integration contracts、图片生成契约、task card index、contracts README、scripts README、roadmap、docs README 和 W24 周志同步该切片。

## 非目标

- 不创建 `services/runtime/image_artifact_response_consumer.py`。
- 不实现 `apply_image_artifact_reference_to_response`。
- 不修改 `services/runtime/inference_response.py`、`services/runtime/inference_support.py`、gateway 或 platform response route。
- 不改 `CopilotResponse` schema，不新增 response 顶层 artifact 字段。
- 不实现 artifact store、binary reader、public URL resolver、backend adapter、artifact upload 或 production storage。
- 不读取 artifact 二进制，不调用真实生图 backend，不生成图片。
- 不新增 UI，不启动开发服务器，不进入 executor、confirmation、writeback 或 replay。

## 停止线

- 不能把本任务卡写成 response consumer runtime code 已完成。
- 后续 runtime code 只能消费 mapper 的 metadata-only success result，不得绕过 mapper fail-closed 结果生成成功 citation。
- 后续若需要 store / reader / public URL / backend adapter，必须回到对应独立任务卡和验证链路。

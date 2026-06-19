# Image Artifact Response Consumer Implementation Readiness v1 任务卡

更新时间：2026-06-14

## 任务标识

- 切片：`image-artifact-response-consumer-implementation-readiness-v1`
- 轨道：`Image Path`
- 状态：`image_artifact_response_consumer_implementation_readiness_defined`

## 目标

在 `image-artifact-runtime-mapper-response-consumer-integration-review-v1` 已确认未来消费入口后，固定真实 response consumer implementation 前必须满足的准入条件、文件落点、函数边界、传播规则和测试策略。

本切片不创建 response consumer 代码，不修改现有 response builder，也不把 metadata-only runtime mapper 接入真实响应生成路径。

## 输入事实源

- [图片生成契约](../contracts/image-generation.md)
- [`Image Artifact Runtime Mapper Response Consumer Integration Review` v1 计划](image-artifact-runtime-mapper-response-consumer-integration-review-v1-plan.md)
- [`Image Artifact Runtime Mapper Runtime Implementation` v1 计划](image-artifact-runtime-mapper-runtime-implementation-v1-plan.md)
- `scripts/checks/fixtures/image-artifact-response-consumer-implementation-readiness-v1.json`
- `services/runtime/image_artifact_runtime_mapper.py`
- `contracts/copilot-response.schema.json`

## 实现准入边界

1. future implementation target
   - 未来实现任务卡只能选择 response consumer 单一路径。
   - 预期文件落点为 `services/runtime/image_artifact_response_consumer.py`。
   - 预期函数边界为 `apply_image_artifact_reference_to_response`，输入为已有 `CopilotResponse` draft 和 mapper result。

2. response contract
   - 成功路径只能向 `CopilotResponse.citations` 追加或合并 `kind=artifact` citation。
   - `metadata_reference` 只能作为内部 handoff 或审计 metadata，不能新增到 `CopilotResponse` 顶层。
   - citation id 冲突、schema 不匹配或 mapper failure 必须 fail closed。

3. implementation gate
   - 必须消费 runtime mapper implementation、integration review、`CopilotResponse` citation schema 和 failure taxonomy。
   - 必须有 success、blocked、failed、pending_review、public URL、binary payload、provider raw dump 和 no side effects 测试。
   - 必须继续证明现有 response builder 在本 readiness 切片中未被接线。

## 验收口径

- 新增 fixture 固定 implementation readiness、future target、function contract、gate matrix、test plan、forbidden scope 和 no side effects。
- 新增 checker 校验前置状态、schema 边界、未来文件未创建、现有 response code 未接线、mapper success / failure 可以支撑未来 consumer test plan。
- checker 接入 `./scripts/check-repo.sh --fast`，并在 response consumer integration review checker 之后运行。

## 非目标

- 不创建 `services/runtime/image_artifact_response_consumer.py`。
- 不实现 `apply_image_artifact_reference_to_response`。
- 不修改 `services/runtime/inference_response.py`、`services/runtime/inference_support.py` 或 gateway / platform response route。
- 不改 `CopilotResponse` schema，不新增 response 顶层 artifact 字段。
- 不实现 artifact store、binary reader、public URL resolver、backend adapter、artifact upload 或 production storage。
- 不读取 artifact 二进制，不调用真实生图 backend，不生成图片。
- 不新增 UI，不启动开发服务器，不进入 executor、confirmation、writeback 或 replay。

## 停止线

- readiness 只说明未来实现任务的准入条件，不代表真实 consumer implementation 已完成。
- 不允许用 public URL、signed URL、binary payload 或 provider raw dump 填补 `metadata_reference`。
- 不允许绕过 mapper fail-closed 结果生成成功 citation。

# Image Artifact Runtime Mapper Response Consumer Integration Review v1 任务卡

更新时间：2026-06-14

## 任务标识

- 切片：`image-artifact-runtime-mapper-response-consumer-integration-review-v1`
- 轨道：`Image Path`
- 状态：`image_artifact_runtime_mapper_response_consumer_integration_review_defined`

## 目标

在 metadata-only runtime mapper 已实现后，评审它未来是否、何时以及通过哪个 response consumer / response builder 入口被消费。本切片只固定集成入口判断和停止线，不把 mapper 接入真实 `CopilotResponse` 生成路径。

结论：未来消费只能沿现有 `CopilotResponse.citations` 的 artifact citation 形状和 mapper 返回的 `metadata_reference` 进行 metadata-only 传递；任何真实 consumer implementation、schema 变更、artifact store、binary reader、public URL resolver 或 backend adapter 都必须另开任务卡。

## 输入事实源

- [图片生成契约](../contracts/image-generation.md)
- [`Image Artifact Runtime Mapper Runtime Implementation` v1 计划](image-artifact-runtime-mapper-runtime-implementation-v1-plan.md)
- `scripts/checks/fixtures/image-artifact-runtime-mapper-response-consumer-integration-review-v1.json`
- `services/runtime/image_artifact_runtime_mapper.py`
- `contracts/copilot-response.schema.json`

## 评审范围

1. response surface
   - 只允许未来复用 `CopilotResponse.citations` 中已有的 `kind=artifact` citation 形状。
   - `metadata_reference` 仍是内部 metadata-only handoff，不新增 `CopilotResponse` 顶层字段。

2. 成功入口
   - 只有 mapper 返回 `ok=true`、citation 符合 `CopilotResponse` citation schema、metadata reference 不含 binary / public URL / provider raw dump 时，未来 consumer 才允许进入成功 response 路径。

3. 失败入口
   - `blocked / failed / pending_review`、invalid metadata、hash mismatch、public URL claim、binary payload、provider raw dump、store / reader 缺失和 provenance 缺失，都必须在 consumer 前或 consumer 内 fail closed。
   - 失败路径不得生成成功 citation，不得降级为部分成功 artifact reference。

## 验收口径

- 新增 fixture 固定 response consumer integration review、候选入口、gate matrix、成功 / 失败传播规则和禁止项。
- 新增 checker 校验 runtime mapper 已实现、`CopilotResponse` 仍只通过 citation schema 承载 artifact reference、失败 case 不产生成功 citation、现有 response builder 未被本切片接线。
- checker 接入 `./scripts/check-repo.sh --fast`，并在 runtime mapper runtime implementation checker 之后运行。

## 非目标

- 不改 `CopilotResponse` schema，不新增 response 顶层 artifact 字段。
- 不实现 response consumer、不改 response builder、不创建 runtime integration code。
- 不实现 artifact store、binary reader、public URL resolver、signed URL resolver、backend adapter 或 production storage。
- 不读取 artifact 二进制，不上传 artifact，不调用真实生图 backend，不生成图片。
- 不新增 UI，不启动开发服务器。
- 不进入 executor、confirmation decision、writeback、replay 或 production API consumer。

## 停止线

- 不能把 integration review 写成 response consumer implementation ready。
- 不能把 mapper 的 `metadata_reference` 暴露成 public URL、signed URL、binary payload 或 provider raw dump。
- 不能为了展示 artifact 而绕过 `blocked / failed / pending_review` 的 fail-closed 规则。

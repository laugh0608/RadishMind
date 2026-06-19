# Image Artifact Runtime Mapper Implementation Entry v1 任务卡

更新时间：2026-06-13

## 任务标识

- 切片：`image-artifact-runtime-mapper-implementation-entry-v1`
- 轨道：`Image Path`
- 状态：`image_artifact_runtime_mapper_implementation_entry_review_defined`

## 目标

在 `image-artifact-runtime-mapper-implementation-plan-v1` 之后，复核 runtime mapper implementation 的真实入口条件。该切片只判断前置证据是否足以让下一步创建 runtime mapper implementation task card，并固定单一实现方向、仍需延后的 store / reader / public URL / backend adapter 方向、禁止创建的 runtime artifact、no fake fallback、no side effects 和下一张任务卡必须消费的契约。

当前结论是：可以把下一张任务卡指向 `image-artifact-runtime-mapper-implementation-v1-plan.md`，但本切片不创建该实现任务卡，也不创建 runtime mapper 代码。后续实现任务卡只能围绕 metadata-only runtime mapper 展开，不得同时打开 artifact store、binary reader、public URL resolver、backend adapter 或 response schema 变更。

## 输入事实源

- [图片生成契约](../contracts/image-generation.md)
- [`Image Artifact Runtime Mapper Implementation Plan` v1 计划](image-artifact-runtime-mapper-implementation-plan-v1-plan.md)
- [`Image Artifact Store / Binary Reader Boundary Readiness` v1 计划](image-artifact-store-binary-reader-boundary-readiness-v1-plan.md)
- [`Image Artifact Runtime Mapping Readiness` v1 计划](image-artifact-runtime-mapping-readiness-v1-plan.md)
- [`Image Artifact Return Runbook / Metadata Evidence` v1 计划](image-artifact-return-runbook-evidence-v1-plan.md)
- [`Image Safety Runbook Evidence` v1 计划](image-safety-runbook-evidence-v1-plan.md)
- [`Image Backend Adapter Readiness Evidence` v1 计划](image-backend-adapter-readiness-evidence-v1-plan.md)

## 验收口径

- `scripts/checks/fixtures/image-artifact-runtime-mapper-implementation-entry-v1.json` 固定 entry boundary、entry gate matrix、implementation candidates、selected track contract、implementation preconditions reconciliation、next implementation constraints、forbidden task card matrix、forbidden artifact matrix、no fake fallback 和 no side effects。
- `scripts/check-image-artifact-runtime-mapper-implementation-entry-v1.py` 进入 `./scripts/check-repo.sh --fast`，并在 runtime mapper implementation plan checker 之后执行。
- checker 必须跨读 runtime mapper implementation plan、store / binary reader boundary readiness、runtime mapping readiness、artifact return runbook、safety runbook、backend adapter readiness、artifact schema fixture 和 `CopilotResponse` schema，确认本切片只打开下一张 metadata-only mapper 实现任务卡方向。
- checker 必须确认当前不创建 runtime mapper implementation task card，不创建 runtime mapper、artifact store、binary reader、public URL resolver、backend adapter implementation、response schema 变更、artifact upload 或真实二进制读取。
- current focus、能力矩阵、integration contracts、图片生成契约、task card index、contracts README、scripts README、roadmap、docs README 和 W24 周志同步该切片。

## 非目标

- 不改 `CopilotResponse` schema，不实现 runtime response mapper。
- 不创建 `image-artifact-runtime-mapper-implementation-v1-plan.md`。
- 不实现 artifact store、artifact upload、binary reader、public URL resolver、signed URL resolver 或 production storage。
- 不读取 artifact 二进制，不生成图片，不调用真实生图 backend，不下载模型，不上传 artifact。
- 不新增产品 UI 面板，不启动开发服务器，不做真实浏览器联调。
- 不进入 executor、confirmation decision、writeback、replay 或 production API consumer。

## 停止线

- 不能把 entry review 写成 runtime mapper ready、runtime mapper implemented、artifact store ready、binary reader ready 或 public URL ready。
- 不能把下一张 mapper implementation task card 扩成 store / reader / backend adapter / public URL 的并行实现入口。
- `blocked / failed / pending_review`、hash mismatch、mime mismatch、dimension mismatch、public URL claim、signed URL policy missing、binary payload、provider raw dump、missing store / reader 和 provenance missing 必须继续 fail closed。
- 后续如创建 `image-artifact-runtime-mapper-implementation-v1-plan.md`，仍需单独任务卡、fixture、checker 和验证，不得从本切片直接写实现。

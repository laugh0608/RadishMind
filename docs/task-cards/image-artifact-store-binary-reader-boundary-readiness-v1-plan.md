# Image Artifact Store / Binary Reader Boundary Readiness v1 任务卡

更新时间：2026-06-13

## 任务标识

- 切片：`image-artifact-store-binary-reader-boundary-readiness-v1`
- 轨道：`Image Path`
- 状态：`image_artifact_store_binary_reader_boundary_readiness_defined`

## 目标

在 `image-artifact-runtime-mapping-implementation-entry-review-v1` 之后，补齐 artifact runtime mapper 进入实现计划前必须具备的 store / binary reader 边界证据。该切片只固定 artifact store ownership、`artifact://` 解析边界、hash / mime type / dimensions revalidation、binary payload redaction、public URL / signed URL 禁止策略、failure taxonomy、no fake fallback、no side effects 和下一步推进条件。

当前结论是：可以定义 store / binary reader 边界 readiness，但仍不创建 artifact store、binary reader、public URL resolver、runtime mapper 或 backend adapter implementation。后续如继续推进，应先进入 runtime mapper implementation plan 评审，而不是直接写 runtime mapper。

## 输入事实源

- [图片生成契约](../contracts/image-generation.md)
- [`Image Artifact Runtime Mapping Implementation Entry Review` v1 计划](image-artifact-runtime-mapping-implementation-entry-review-v1-plan.md)
- [`Image Artifact Runtime Mapping Readiness` v1 计划](image-artifact-runtime-mapping-readiness-v1-plan.md)
- [`Image Artifact Return Runbook / Metadata Evidence` v1 计划](image-artifact-return-runbook-evidence-v1-plan.md)
- [`Image Safety Runbook Evidence` v1 计划](image-safety-runbook-evidence-v1-plan.md)
- [`Image Backend Adapter Readiness Evidence` v1 计划](image-backend-adapter-readiness-evidence-v1-plan.md)

## 验收口径

- `scripts/checks/fixtures/image-artifact-store-binary-reader-boundary-readiness-v1.json` 固定 boundary readiness、store contract、binary reader contract、validation matrix、public URL policy、failure taxonomy、runtime mapping plan inputs、forbidden artifact matrix、no fake fallback 和 no side effects。
- `scripts/check-image-artifact-store-binary-reader-boundary-readiness-v1.py` 进入 `./scripts/check-repo.sh --fast`，并在 entry review checker 之后执行。
- checker 必须跨读 entry review、runtime mapping readiness、artifact return runbook、safety runbook、backend adapter readiness 和 artifact schema fixture，确认本切片只解除 store / binary reader boundary blocker，不声明 runtime implementation ready。
- checker 必须确认当前不创建 artifact store、binary reader、public URL resolver、runtime mapper、backend adapter implementation、response schema 变更或真实 storage / binary read artifact。
- current focus、能力矩阵、integration contracts、图片生成契约、task card index、contracts README、scripts README、roadmap、docs README 和 W24 周志同步该切片。

## 非目标

- 不改 `CopilotResponse` schema，不实现 runtime response mapper。
- 不实现 artifact store、artifact upload、binary reader、public URL resolver、signed URL resolver 或 production storage。
- 不读取 artifact 二进制，不生成图片，不调用真实生图 backend，不下载模型，不上传 artifact。
- 不新增产品 UI 面板，不启动开发服务器，不做真实浏览器联调。
- 不进入 executor、confirmation decision、writeback、replay 或 production API consumer。

## 停止线

- 不能把 boundary readiness 写成 artifact store ready、binary reader ready、public URL ready 或 runtime mapper ready。
- `artifact://` 仍只是 metadata reference，不是 public URL、signed URL、production storage path 或 binary download endpoint。
- hash mismatch、mime type mismatch、dimension mismatch、public URL claim、binary payload、provider raw dump、pending review 和 blocked review 必须 fail closed。
- 后续如进入 runtime mapper implementation plan，仍需单独任务卡、fixture、checker 和验证，不得从本切片直接写实现。

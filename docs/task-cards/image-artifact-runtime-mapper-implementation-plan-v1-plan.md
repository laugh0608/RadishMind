# Image Artifact Runtime Mapper Implementation Plan v1 任务卡

更新时间：2026-06-13

## 任务标识

- 切片：`image-artifact-runtime-mapper-implementation-plan-v1`
- 轨道：`Image Path`
- 状态：`image_artifact_runtime_mapper_implementation_plan_defined`

## 目标

在 `image-artifact-store-binary-reader-boundary-readiness-v1` 之后，定义未来 artifact runtime mapper 的实现计划评审证据。该切片只固定 mapper 输入、输出、成功 / blocked 映射、fail-closed taxonomy、单一实现方向选择策略、禁止创建的 runtime artifact、no fake fallback、no side effects 和下一步实现入口复核条件。

当前结论是：runtime mapper 的计划输入已经可整理，但实现入口仍未打开。后续如继续推进，应先创建 runtime mapper implementation entry review，确认仍只选择一个实现方向，再决定是否进入真实 runtime mapper implementation。

## 输入事实源

- [图片生成契约](../contracts/image-generation.md)
- [`Image Artifact Store / Binary Reader Boundary Readiness` v1 计划](image-artifact-store-binary-reader-boundary-readiness-v1-plan.md)
- [`Image Artifact Runtime Mapping Implementation Entry Review` v1 计划](image-artifact-runtime-mapping-implementation-entry-review-v1-plan.md)
- [`Image Artifact Runtime Mapping Readiness` v1 计划](image-artifact-runtime-mapping-readiness-v1-plan.md)
- [`Image Artifact Return Runbook / Metadata Evidence` v1 计划](image-artifact-return-runbook-evidence-v1-plan.md)
- [`Image Safety Runbook Evidence` v1 计划](image-safety-runbook-evidence-v1-plan.md)
- [`Image Backend Adapter Readiness Evidence` v1 计划](image-backend-adapter-readiness-evidence-v1-plan.md)

## 验收口径

- `scripts/checks/fixtures/image-artifact-runtime-mapper-implementation-plan-v1.json` 固定 implementation plan boundary、mapper input contract、target response reference plan、mapping case plan、fail-closed plan、single track policy、implementation entry recommendation、forbidden artifact matrix、no fake fallback 和 no side effects。
- `scripts/check-image-artifact-runtime-mapper-implementation-plan-v1.py` 进入 `./scripts/check-repo.sh --fast`，并在 store / binary reader boundary readiness checker 之后执行。
- checker 必须跨读 store / binary reader boundary readiness、implementation entry review、runtime mapping readiness、artifact return runbook、safety runbook、backend adapter readiness、artifact schema fixture 和 `CopilotResponse` schema，确认计划只消费证据，不声明 mapper ready。
- checker 必须确认当前不创建 runtime mapper、artifact store、binary reader、public URL resolver、backend adapter implementation、response schema 变更、artifact upload 或真实二进制读取。
- current focus、能力矩阵、integration contracts、图片生成契约、task card index、contracts README、scripts README、roadmap、docs README 和 W24 周志同步该切片。

## 非目标

- 不改 `CopilotResponse` schema，不实现 runtime response mapper。
- 不创建 runtime mapper implementation entry review 之外的真实实现任务卡。
- 不实现 artifact store、artifact upload、binary reader、public URL resolver、signed URL resolver 或 production storage。
- 不读取 artifact 二进制，不生成图片，不调用真实生图 backend，不下载模型，不上传 artifact。
- 不新增产品 UI 面板，不启动开发服务器，不做真实浏览器联调。
- 不进入 executor、confirmation decision、writeback、replay 或 production API consumer。

## 停止线

- 不能把 implementation plan 写成 runtime mapper ready、artifact store ready、binary reader ready 或 public URL ready。
- 不能把 `image_artifact_store_binary_reader_boundary_readiness_defined` 写成 store / reader implementation complete。
- blocked / failed / pending_review、hash mismatch、mime mismatch、dimension mismatch、public URL claim、signed URL policy missing、binary payload、provider raw dump、missing store / reader 和 provenance missing 必须 fail closed。
- 后续如进入 runtime mapper implementation entry review，仍需单独任务卡、fixture、checker 和验证，不得从本切片直接写实现。

# Image Artifact Runtime Mapping Implementation Entry Review v1 任务卡

更新时间：2026-06-13

## 任务标识

- 切片：`image-artifact-runtime-mapping-implementation-entry-review-v1`
- 轨道：`Image Path`
- 状态：`image_artifact_runtime_mapping_entry_review_defined`

## 目标

在 `image-artifact-runtime-mapping-readiness-v1` 之后，复核 Image Path 是否可以进入真实 artifact runtime mapper 实现入口。该切片只固定实现入口判断、当前 blocker、后续推进顺序、禁止创建的实现 artifact、no fake fallback、no side effects 和停止线。

当前结论仍是：runtime mapping readiness 已具备评审证据，但 artifact store / binary reader 边界、runtime mapper implementation plan、backend adapter implementation 和 storage / URL failure gate 都未满足，因此不创建 implementation task card，不实现 runtime mapper。

## 输入事实源

- [图片生成契约](../contracts/image-generation.md)
- [`Image Artifact Runtime Mapping Readiness` v1 计划](image-artifact-runtime-mapping-readiness-v1-plan.md)
- [`Image Artifact Return Runbook / Metadata Evidence` v1 计划](image-artifact-return-runbook-evidence-v1-plan.md)
- [`Image Safety Runbook Evidence` v1 计划](image-safety-runbook-evidence-v1-plan.md)
- [`Image Backend Adapter Readiness Evidence` v1 计划](image-backend-adapter-readiness-evidence-v1-plan.md)

## 验收口径

- `scripts/checks/fixtures/image-artifact-runtime-mapping-implementation-entry-review-v1.json` 固定 implementation entry boundary、entry gate matrix、implementation candidates、runtime mapping readiness reconciliation、next development policy、review order、forbidden task card / artifact matrix、no fake fallback 和 no side effects。
- `scripts/check-image-artifact-runtime-mapping-implementation-entry-review-v1.py` 进入 `./scripts/check-repo.sh --fast`。
- checker 必须跨读 runtime mapping readiness、artifact return runbook、safety runbook 和 backend adapter readiness fixture，确认 readiness 不会被提升为 runtime mapper implementation ready。
- checker 必须确认当前不创建 runtime mapper、artifact store、binary reader、public URL resolver、backend adapter implementation 或 response schema 变更。
- current focus、能力矩阵、integration contracts、图片生成契约、task card index、contracts README、scripts README、roadmap 和 W24 周志同步该切片。

## 非目标

- 不改 `CopilotResponse` schema，不实现 runtime response mapper。
- 不创建 runtime mapper implementation task card。
- 不实现 artifact store、artifact upload、binary reader、public URL resolver 或 production storage。
- 不调用真实生图 backend，不下载图片模型，不生成图片，不提交图片像素。
- 不新增产品 UI 面板，不启动开发服务器，不做真实浏览器联调。
- 不进入 executor、confirmation decision、writeback、replay 或 production API consumer。

## 停止线

- 不能把 `image_artifact_runtime_mapping_readiness_defined` 写成 runtime mapper ready。
- 不能把 artifact metadata reference 写成 artifact store、binary reader 或 public URL ready。
- 若后续进入实现准备，应先补 artifact store / binary reader boundary readiness，再定义 runtime mapper implementation plan。
- 不能把 blocked / failed / pending_review、hash mismatch、public URL claim、binary payload 或 provider raw dump 映射为成功 response artifact citation。

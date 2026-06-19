# Image Artifact Runtime Mapping Readiness v1 任务卡

更新时间：2026-06-13

## 任务标识

- 切片：`image-artifact-runtime-mapping-readiness-v1`
- 轨道：`Image Path`
- 状态：`image_artifact_runtime_mapping_readiness_defined`

## 目标

在真正实现 image artifact runtime mapper、artifact store、public URL、binary reader 或 `CopilotResponse` schema 变更前，先把 `image_generation_artifact` metadata 到未来 response artifact citation / metadata reference 的准入证据链固定为可检查契约。

本切片补齐：

- `artifact://` URI、hash、mime type、dimensions、safety review 和 provenance 的必要字段准入。
- `generated` 且 safety review 通过或无需审查的 artifact 才能进入未来成功 response reference。
- `blocked`、`failed`、`pending_review` artifact 不能进入成功 response。
- invalid metadata、hash mismatch、public URL claim、binary payload 和 provider raw dump 的 fail-closed 映射。
- 与 `image-backend-adapter-readiness-evidence-v1`、`image-safety-runbook-evidence-v1` 和 `image-artifact-return-runbook-evidence-v1` 的依赖关系。

本切片只创建任务卡、fixture 和 checker，不改 `CopilotResponse` schema，不实现 runtime mapper，不创建 artifact store / public URL / binary reader，不调用真实生图 backend。

## 输入事实源

- [图片生成契约](../contracts/image-generation.md)
- [`Image Artifact Return Runbook / Metadata Evidence` v1 计划](image-artifact-return-runbook-evidence-v1-plan.md)
- [`Image Safety Runbook Evidence` v1 计划](image-safety-runbook-evidence-v1-plan.md)
- [`Image Backend Adapter Readiness Evidence` v1 计划](image-backend-adapter-readiness-evidence-v1-plan.md)
- `contracts/image-generation-artifact.schema.json`
- `contracts/copilot-response.schema.json`
- `scripts/checks/fixtures/image-generation-artifact-basic.json`
- `scripts/checks/fixtures/image-artifact-return-runbook-evidence-v1.json`
- `scripts/checks/fixtures/image-safety-runbook-evidence-v1.json`
- `scripts/checks/fixtures/image-backend-adapter-readiness-evidence-v1.json`

## 验收口径

- `scripts/checks/fixtures/image-artifact-runtime-mapping-readiness-v1.json` 固定 runtime mapping boundary、future mapping contract、metadata field matrix、success / blocked mapping matrix、fail-closed taxonomy、forbidden artifact、no fallback 和 no side effect 口径。
- `scripts/check-image-artifact-runtime-mapping-readiness-v1.py` 进入 `./scripts/check-repo.sh --fast`。
- checker 必须跨读 artifact schema、`CopilotResponse` schema、基础 artifact fixture、artifact return runbook、safety runbook 和 backend adapter readiness fixture。
- checker 必须确认 `artifact://` 不被提升为 public URL，binary payload 和 provider raw dump 不进入 future response reference。
- checker 必须确认 blocked / failed / pending_review artifact 不返回成功 reference，invalid metadata、hash mismatch、public URL claim、binary payload 和 provider raw dump 都 fail closed。
- current focus、能力矩阵、integration contracts、图片生成契约、task card index、contracts README、scripts README、roadmap 和 W24 周志同步该切片。

## 非目标

- 不改 `CopilotResponse` schema，不实现 runtime response mapper。
- 不实现 artifact store、artifact upload、binary reader、public URL resolver 或 production storage。
- 不调用真实生图 backend，不下载图片模型，不生成图片，不提交图片像素。
- 不新增产品 UI 面板，不启动开发服务器，不做真实浏览器联调。
- 不进入 executor、confirmation decision、writeback、replay 或 production API consumer。

## 停止线

- 不能把 runtime mapping readiness 写成 runtime mapper 已实现。
- 不能把 metadata reference 写成图片二进制可读取或 public URL ready。
- 不能把 `blocked`、`failed`、`pending_review` 或 hash mismatch artifact 写成成功 response artifact citation。
- 不能把 provider raw response、base64 image 或 pixel payload 写入未来 response metadata reference。

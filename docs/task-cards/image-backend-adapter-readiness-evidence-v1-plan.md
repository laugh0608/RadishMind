# Image Backend Adapter Readiness Evidence v1 任务卡

更新时间：2026-06-13

## 任务标识

- 切片：`image-backend-adapter-readiness-evidence-v1`
- 轨道：`Image Path`
- 状态：`image_backend_adapter_readiness_defined`

## 目标

在真实生图 backend client、backend profile registry、credential resolver、model-dir resolver 或 endpoint health probe 实现前，先把未来 backend adapter 的实现准入证据固定为可检查链路。

本切片补齐：

- backend adapter ownership、profile、request / response / failure envelope 的准入口径。
- backend credential、model dir、endpoint、timeout budget 和 response artifact metadata 的前置证据要求。
- backend unavailable、profile missing、credential missing、model dir missing、timeout、invalid artifact metadata、hash mismatch 和 safety gate blocked 的 fail-closed 映射。
- 与 `image-adapter-handshake-safety-gate-v1`、`image-artifact-return-runbook-evidence-v1` 和 `image-safety-runbook-evidence-v1` 的依赖关系。
- 真实 backend adapter 实现前必须保持 no backend call、no image generation、no artifact upload、no public URL 的停止线。

本切片只创建任务卡、fixture 和 checker，不创建真实 backend client，不调用 backend，不下载模型，不生成图片，不上传 artifact。

## 输入事实源

- [图片生成契约](../contracts/image-generation.md)
- [`Image Adapter Handshake / Safety Gate` v1 计划](image-adapter-handshake-safety-gate-v1-plan.md)
- [`Image Artifact Return Runbook / Metadata Evidence` v1 计划](image-artifact-return-runbook-evidence-v1-plan.md)
- [`Image Safety Runbook Evidence` v1 计划](image-safety-runbook-evidence-v1-plan.md)
- `contracts/image-generation-backend-request.schema.json`
- `contracts/image-generation-artifact.schema.json`
- `scripts/checks/fixtures/image-generation-backend-request-basic.json`
- `scripts/checks/fixtures/image-generation-artifact-basic.json`
- `scripts/checks/fixtures/image-generation-eval-manifest-v0.json`
- `scripts/checks/fixtures/image-adapter-handshake-safety-gate-v1.json`
- `scripts/checks/fixtures/image-artifact-return-runbook-evidence-v1.json`
- `scripts/checks/fixtures/image-safety-runbook-evidence-v1.json`

## 验收口径

- `scripts/checks/fixtures/image-backend-adapter-readiness-evidence-v1.json` 固定 adapter boundary、contract input fields、readiness prerequisites、profile matrix、readiness gate matrix、failure taxonomy、future smoke contract、forbidden artifact、no fallback 和 no side effect 口径。
- `scripts/check-image-backend-adapter-readiness-evidence-v1.py` 进入 `./scripts/check-repo.sh --fast`。
- checker 必须跨读 backend request / artifact schema 与基础 fixture，并确认仍复用前序 handshake、artifact return、safety runbook 和 eval manifest。
- checker 必须确认 backend request 的 backend/profile/parameters/safety/trace 与 artifact metadata 的 backend/hash/provenance 字段保持可追溯。
- checker 必须确认本切片不创建 backend adapter、backend client、profile resolver、credential resolver、model-dir resolver、endpoint health probe、artifact store、UI 面板、图片生成脚本或 deployment artifact。
- current focus、能力矩阵、integration contracts、图片生成契约、task card index、contracts README、scripts README、roadmap、architecture 和 W24 周志同步该切片。

## 非目标

- 不实现 `RadishMind-Image Adapter` runtime 或 image backend adapter。
- 不实现 backend client、profile registry、credential resolver、model-dir resolver、endpoint health probe 或 backend queue。
- 不调用真实生图 backend，不下载图片模型，不读取模型目录，不生成图片，不提交图片像素。
- 不改 `CopilotResponse` schema，不实现 runtime response mapping。
- 不实现 artifact store、artifact upload、binary reader、public URL resolver 或 production storage。
- 不新增产品 UI 面板，不启动开发服务器，不做真实浏览器联调。
- 不进入 executor、confirmation decision、writeback、replay 或 production API consumer。

## 停止线

- 不能把 backend adapter readiness 写成 backend adapter implementation ready。
- 不能把 profile / credential / model dir / endpoint 占位证据写成真实连接已可用。
- 不能把 `approved_for_backend` 写成当前允许真实 backend 调用。
- 不能把 backend unavailable、timeout、invalid metadata 或 hash mismatch 写成 retry loop、fallback execution 或 success artifact reference。

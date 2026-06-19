# Image Safety Runbook Evidence v1 任务卡

更新时间：2026-06-13

## 任务标识

- 切片：`image-safety-runbook-evidence-v1`
- 轨道：`Image Path`
- 状态：`image_safety_runbook_evidence_defined`

## 目标

在真实生图 backend、外部 moderation provider、runtime policy engine 或 artifact store 实现前，先把图片路径的安全审查 runbook 固定为可检查证据链。

本切片补齐：

- `image_generation` intent 的 prompt / constraint / safety precheck 证据。
- Adapter backend request 前的 safety gate 与 blocked confirmation / high risk / policy unknown 停止线。
- `image_generation_artifact` metadata 的 safety review 字段、blocked / pending review 状态和 provenance 要求。
- unsafe prompt、requires confirmation、high risk、policy unknown、backend unavailable、artifact safety pending / blocked 的失败分类。
- 与 `image-adapter-handshake-safety-gate-v1` 和 `image-artifact-return-runbook-evidence-v1` 的依赖关系。

本切片只创建任务卡、fixture 和 checker，不创建 runtime safety engine、不接 moderation provider、不调用真实生图 backend、不生成图片、不上传 artifact。

## 输入事实源

- [图片生成契约](../contracts/image-generation.md)
- [`Image Adapter Handshake / Safety Gate` v1 计划](image-adapter-handshake-safety-gate-v1-plan.md)
- [`Image Artifact Return Runbook / Metadata Evidence` v1 计划](image-artifact-return-runbook-evidence-v1-plan.md)
- `contracts/image-generation-intent.schema.json`
- `contracts/image-generation-backend-request.schema.json`
- `contracts/image-generation-artifact.schema.json`
- `scripts/checks/fixtures/image-generation-intent-basic.json`
- `scripts/checks/fixtures/image-generation-backend-request-basic.json`
- `scripts/checks/fixtures/image-generation-artifact-basic.json`
- `scripts/checks/fixtures/image-adapter-handshake-safety-gate-v1.json`
- `scripts/checks/fixtures/image-artifact-return-runbook-evidence-v1.json`

## 验收口径

- `scripts/checks/fixtures/image-safety-runbook-evidence-v1.json` 固定 safety boundary、policy input fields、runbook steps、decision matrix、failure taxonomy、artifact review requirements、forbidden artifact、no fallback 和 no side effect 口径。
- `scripts/check-image-safety-runbook-evidence-v1.py` 进入 `./scripts/check-repo.sh --fast`。
- checker 必须跨读 intent、backend request、artifact schema / fixture、handshake fixture 和 artifact return fixture。
- checker 必须确认 low risk intent、requires confirmation intent、high risk intent、blocked backend request 和 blocked artifact metadata 的安全字段保持一致。
- checker 必须确认本切片不创建 safety runtime、moderation provider、backend client、artifact store、public URL resolver、UI 面板、图片生成脚本或 deployment artifact。
- current focus、能力矩阵、integration contracts、图片生成契约、task card index、contracts README、scripts README、roadmap、architecture 和 W24 周志同步该切片。

## 非目标

- 不实现 runtime safety engine、policy engine、moderation provider 或真实 classifier。
- 不调用真实生图 backend，不下载图片模型，不生成图片，不提交图片像素。
- 不改 `CopilotResponse` schema，不实现 runtime response mapping。
- 不实现 artifact store、artifact upload、binary reader、public URL resolver 或 production storage。
- 不新增产品 UI 面板，不启动开发服务器，不做真实浏览器联调。
- 不进入 executor、confirmation decision、writeback、replay 或 production API consumer。

## 停止线

- 不能把安全 runbook 写成 safety runtime ready。
- 不能把 policy unknown / high risk / requires confirmation 写成可自动提交 backend。
- 不能把 artifact `review_status=pending_review` 或 `blocked` 写成可返回成功 artifact reference。
- 不能把 backend unavailable 写成自动 retry、fallback execution 或 backend health ready。

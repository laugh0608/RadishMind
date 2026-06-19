# Image Artifact Return Runbook / Metadata Evidence v1 任务卡

更新时间：2026-06-13

## 任务标识

- 切片：`image-artifact-return-runbook-evidence-v1`
- 轨道：`Image Path`
- 状态：`image_artifact_return_runbook_evidence_defined`

## 目标

在真实 artifact store、public URL、runtime response mapping 或 image backend adapter 实现前，先把图片生成结果回到上层响应的 metadata-only 返回证据固定为可检查 runbook。

本切片补齐：

- `image_generation_artifact` metadata 到未来上层 response artifact reference 的字段映射。
- `artifact://` URI、hash、尺寸、格式、backend、safety 和 provenance 的返回要求。
- backend unavailable、artifact metadata missing、hash mismatch、unsafe artifact blocked 的失败分类。
- metadata-only 返回、禁止 pixel payload、禁止 provider raw dump、禁止 public URL claim 的停止线。
- 与 `image-adapter-handshake-safety-gate-v1` 的依赖关系，确认 artifact return 不绕过 adapter safety gate。

本切片只创建任务卡、fixture 和 checker，不创建 response runtime mapping、不改 `CopilotResponse` schema、不上传 artifact、不接生产存储。

## 输入事实源

- [图片生成契约](../contracts/image-generation.md)
- [`Image Adapter Handshake / Safety Gate` v1 计划](image-adapter-handshake-safety-gate-v1-plan.md)
- `contracts/image-generation-artifact.schema.json`
- `contracts/copilot-response.schema.json`
- `scripts/checks/fixtures/image-generation-artifact-basic.json`
- `scripts/checks/fixtures/image-adapter-handshake-safety-gate-v1.json`
- `scripts/checks/fixtures/image-generation-eval-manifest-v0.json`

## 验收口径

- `scripts/checks/fixtures/image-artifact-return-runbook-evidence-v1.json` 固定 artifact return boundary、metadata reference shape、failure taxonomy、runbook steps、forbidden artifact、no fallback 和 no side effect 口径。
- `scripts/check-image-artifact-return-runbook-evidence-v1.py` 进入 `./scripts/check-repo.sh --fast`。
- checker 必须跨读 artifact schema、artifact fixture、image eval manifest 和 handshake / safety gate fixture。
- checker 必须确认 artifact URI 仍是 `artifact://` metadata reference，不是 `http(s)` public URL。
- checker 必须确认本切片不创建 artifact store、response mapper、public URL resolver、image UI 面板、backend client 或图片生成脚本。
- current focus、能力矩阵、integration contracts、图片生成契约、task card index、contracts README、scripts README、roadmap、architecture 和 W24 周志同步该切片。

## 非目标

- 不实现 artifact store、artifact upload、binary reader、public URL resolver 或 production storage。
- 不改 `CopilotResponse` schema，不实现 runtime response mapping。
- 不调用真实生图 backend，不下载图片模型，不生成图片，不提交图片像素。
- 不新增产品 UI 面板，不启动开发服务器，不做真实浏览器联调。
- 不进入 executor、confirmation decision、writeback、replay 或 production API consumer。

## 停止线

- 不能把 artifact metadata reference 写成图片像素可读取。
- 不能把 `artifact://` 写成 public URL 或 production storage ready。
- 不能绕过 `image-adapter-handshake-safety-gate-v1` 的 safety gate。
- 不能把 failure taxonomy 写成自动 retry、fallback execution 或 backend health ready。

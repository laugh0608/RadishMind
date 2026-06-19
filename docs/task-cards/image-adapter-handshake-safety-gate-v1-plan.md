# Image Adapter Handshake / Safety Gate v1 任务卡

更新时间：2026-06-13

## 任务标识

- 切片：`image-adapter-handshake-safety-gate-v1`
- 轨道：`Image Path`
- 状态：`image_adapter_handshake_safety_gate_defined`

## 目标

在真实 `RadishMind-Image Adapter` runtime、backend client、artifact store 或 artifact 返回链路实现前，先把 adapter handoff 和 safety gate 证据固定成可检查契约。

本切片补齐：

- `RadishMind-Core -> RadishMind-Image Adapter` 的结构化 intent handoff。
- Adapter safety gate 到 backend request materialization 的判断关系。
- `requires_confirmation=true`、高风险或 backend unavailable 场景的阻断边界。
- backend result 到 artifact metadata 的只读审计关系。
- artifact metadata 回到上层响应前的 metadata-only 返回边界。

本切片只创建任务卡、fixture 和 checker，不创建真实 adapter、backend client、图片生成任务、artifact upload 或公开 URL。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [能力矩阵](../radishmind-capability-matrix.md)
- [跨项目集成契约](../radishmind-integration-contracts.md)
- [图片生成契约](../contracts/image-generation.md)
- `contracts/image-generation-intent.schema.json`
- `contracts/image-generation-backend-request.schema.json`
- `contracts/image-generation-artifact.schema.json`
- `scripts/checks/fixtures/image-generation-intent-basic.json`
- `scripts/checks/fixtures/image-generation-backend-request-basic.json`
- `scripts/checks/fixtures/image-generation-artifact-basic.json`
- `scripts/checks/fixtures/image-generation-eval-manifest-v0.json`

## 验收口径

- `scripts/checks/fixtures/image-adapter-handshake-safety-gate-v1.json` 固定 handshake boundary、phase matrix、safety gate matrix、artifact return policy、forbidden artifact、no fallback 和 no side effect 口径。
- `scripts/check-image-adapter-handshake-safety-gate-v1.py` 进入 `./scripts/check-repo.sh --fast`。
- checker 必须跨读三份 image generation schema、三份基础 fixture 和 eval manifest，确认 handshake / safety gate 仍复用既有契约链路。
- checker 必须确认本切片不调用真实 backend、不生成图片、不下载模型、不上传 artifact、不创建 production storage、不把像素生成并入 `RadishMind-Core`。
- current focus、能力矩阵、integration contracts、图片生成契约、task card index、contracts README、scripts README、roadmap、architecture 和 W24 周志同步该切片。

## 非目标

- 不实现 `RadishMind-Image Adapter` runtime。
- 不实现 image backend client、provider adapter、backend queue、artifact store 或公开 URL。
- 不调用真实生图 backend，不下载图片模型，不生成图片，不提交图片像素。
- 不新增产品 UI 面板，不启动开发服务器，不做真实浏览器联调。
- 不把 artifact metadata URI 解释成已可公开访问的业务 URL。
- 不把图片像素生成并入 `RadishMind-Core`。
- 不进入 executor、confirmation decision、writeback、replay 或 production API consumer。

## 停止线

- 不能把 `approved_for_backend` 写成当前已经允许真实 backend 调用。
- 不能把 metadata-only artifact 返回写成 artifact upload、production storage 或 public URL ready。
- 不能把 safety gate 文档写成 image safety runbook 已完成。
- 不能把 eval manifest 的结构化链路检查写成图片质量评测或 backend latency 评测。

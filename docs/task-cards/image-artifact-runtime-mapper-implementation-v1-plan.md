# Image Artifact Runtime Mapper Implementation v1 任务卡

更新时间：2026-06-13

## 任务标识

- 切片：`image-artifact-runtime-mapper-implementation-v1`
- 轨道：`Image Path`
- 状态：`image_artifact_runtime_mapper_implementation_task_card_defined`

## 目标

在 `image-artifact-runtime-mapper-implementation-entry-v1` 已放行下一张任务卡之后，固定 metadata-only runtime mapper 的实现准入边界。该切片只创建实现任务卡、fixture 和 checker，定义后续 runtime 代码必须如何消费 `image_generation_artifact` metadata，并投影到未来 `CopilotResponse` artifact citation / metadata reference。

当前结论是：下一步可以进入 runtime mapper 代码实现，但范围只限 metadata-only mapper。后续代码必须从 artifact metadata 生成引用结构，不得读取 artifact 二进制、不得查 artifact store、不得创建 public URL / signed URL、不得调用真实生图 backend，也不得修改 `CopilotResponse` schema。

## 输入事实源

- [图片生成契约](../contracts/image-generation.md)
- [`Image Artifact Runtime Mapper Implementation Entry` v1 计划](image-artifact-runtime-mapper-implementation-entry-v1-plan.md)
- [`Image Artifact Runtime Mapper Implementation Plan` v1 计划](image-artifact-runtime-mapper-implementation-plan-v1-plan.md)
- [`Image Artifact Store / Binary Reader Boundary Readiness` v1 计划](image-artifact-store-binary-reader-boundary-readiness-v1-plan.md)
- [`Image Artifact Runtime Mapping Readiness` v1 计划](image-artifact-runtime-mapping-readiness-v1-plan.md)
- [`Image Artifact Return Runbook / Metadata Evidence` v1 计划](image-artifact-return-runbook-evidence-v1-plan.md)
- [`Image Safety Runbook Evidence` v1 计划](image-safety-runbook-evidence-v1-plan.md)
- [`Image Backend Adapter Readiness Evidence` v1 计划](image-backend-adapter-readiness-evidence-v1-plan.md)

## 验收口径

- `scripts/checks/fixtures/image-artifact-runtime-mapper-implementation-v1.json` 固定 implementation task card boundary、implementation scope、input metadata contract、target response reference contract、mapping case requirements、fail-closed requirements、dependency reconciliation、future runtime test plan、forbidden task card / artifact matrix、no fake fallback 和 no side effects。
- `scripts/check-image-artifact-runtime-mapper-implementation-v1.py` 进入 `./scripts/check-repo.sh --fast`，并在 runtime mapper implementation entry checker 之后执行。
- checker 必须跨读 implementation entry、implementation plan、store / binary reader boundary readiness、runtime mapping readiness、artifact return runbook、safety runbook、backend adapter readiness、artifact schema fixture 和 `CopilotResponse` schema，确认本切片只定义后续 metadata-only mapper runtime 代码边界。
- checker 必须确认 `artifact://`、`sha256`、`mime_type`、`width / height`、`safety.review_status`、`provenance` 是成功引用必要字段；`blocked / failed / pending_review` 不得进入成功 response。
- checker 必须确认 invalid metadata、hash mismatch、mime mismatch、dimension mismatch、public URL claim、binary payload、provider raw dump、missing store / reader、safety review 未通过和 provenance missing 均 fail closed。
- current focus、能力矩阵、integration contracts、图片生成契约、task card index、contracts README、scripts README、roadmap、docs README 和 W24 周志同步该切片。

## 非目标

- 不改 `CopilotResponse` schema，不创建新的 response schema 字段。
- 不在本切片实现 runtime mapper 代码，只定义后续实现准入和验证口径。
- 不实现 artifact store、artifact upload、binary reader、public URL resolver、signed URL resolver 或 production storage。
- 不读取 artifact 二进制，不生成图片，不调用真实生图 backend，不下载模型，不上传 artifact。
- 不新增产品 UI 面板，不启动开发服务器，不做真实浏览器联调。
- 不进入 executor、confirmation decision、writeback、replay 或 production API consumer。

## 停止线

- 不能把本任务卡写成 runtime mapper 已实现、response schema 已变更、artifact store ready、binary reader ready 或 public URL ready。
- 后续 runtime mapper 只能消费 metadata-only `image_generation_artifact`，不能隐式引入 store lookup、binary read、public URL resolution、backend retry 或 fallback execution。
- `blocked / failed / pending_review`、hash mismatch、mime mismatch、dimension mismatch、public URL claim、signed URL policy missing、binary payload、provider raw dump、missing store / reader 和 provenance missing 必须继续 fail closed。
- 后续如实现 mapper 代码，仍需单独代码提交和验证，不得从本任务卡直接创建 store / reader / backend adapter / public URL 实现。

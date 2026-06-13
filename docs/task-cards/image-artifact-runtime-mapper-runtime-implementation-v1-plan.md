# Image Artifact Runtime Mapper Runtime Implementation v1 任务卡

更新时间：2026-06-13

## 任务标识

- 切片：`image-artifact-runtime-mapper-runtime-implementation-v1`
- 轨道：`Image Path`
- 状态：`image_artifact_runtime_mapper_runtime_implemented`

## 目标

在 `image-artifact-runtime-mapper-implementation-v1` 已固定实现任务卡之后，落地 metadata-only runtime mapper 代码。该切片只实现纯 metadata mapper：消费 `image_generation_artifact` metadata，校验成功引用的必要字段，并输出 future `CopilotResponse` 可消费的 artifact citation / metadata reference 结构。

该实现不读取 artifact 二进制，不查 artifact store，不解析 public URL，不调用真实生图 backend，不修改 `CopilotResponse` schema，也不接入 executor、confirmation、writeback 或 replay。

## 输入事实源

- [图片生成契约](../contracts/image-generation.md)
- [`Image Artifact Runtime Mapper Implementation` v1 计划](image-artifact-runtime-mapper-implementation-v1-plan.md)
- [`Image Artifact Runtime Mapping Readiness` v1 计划](image-artifact-runtime-mapping-readiness-v1-plan.md)
- `scripts/checks/fixtures/image-artifact-runtime-mapper-runtime-implementation-v1.json`
- `services/runtime/image_artifact_runtime_mapper.py`

## 实现范围

1. `services/runtime/image_artifact_runtime_mapper.py`
   - 提供 `map_image_artifact_to_response_reference`。
   - 成功路径返回 `citation` 和 `metadata_reference`。
   - 失败路径返回稳定 `failure_code` 和 `failure_message`。

2. 成功映射
   - 支持 `generated + not_required`。
   - 支持 `generated + reviewed_pass`。
   - 保留 `artifact://`、sha256、mime type、dimensions、safety review、provenance 和 generation metadata。

3. 失败映射
   - `blocked / failed / pending_review` 不进入成功引用。
   - invalid metadata、hash mismatch、mime mismatch、dimension mismatch、public URL claim、signed URL policy missing、binary payload、provider raw dump、missing / unavailable store、missing / forbidden binary reader、safety review not passed 和 provenance missing 均 fail closed。

## 验收口径

- `scripts/checks/fixtures/image-artifact-runtime-mapper-runtime-implementation-v1.json` 固定 runtime implementation 范围、成功 case、fail-closed case、side-effect counter 和文档引用。
- `scripts/check-image-artifact-runtime-mapper-runtime-implementation-v1.py` 导入实际 mapper 并执行成功 / 失败 / no side effects 检查。
- `scripts/check-repo.py --fast` 在 `image-artifact-runtime-mapper-implementation-v1` 之后运行该 checker。

## 非目标

- 不改 `CopilotResponse` schema，不创建新 response 字段。
- 不实现 artifact store、binary reader、public URL resolver、signed URL resolver、backend adapter 或 production storage。
- 不读取 artifact 二进制，不生成图片，不下载模型，不上传 artifact，不调用真实 backend。
- 不新增 UI，不启动开发服务器，不做浏览器联调。
- 不进入 executor、confirmation decision、writeback、replay 或 production API consumer。

## 停止线

- 不能把 metadata-only mapper 写成完整 image generation runtime ready。
- 不能把 fail-closed 错误降级为成功 response citation。
- 不能为了补齐 metadata reference 引入二进制读取、store lookup、public URL resolution、backend retry 或 fallback execution。

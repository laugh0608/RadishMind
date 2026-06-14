# Image Generation / Artifact Return 设计与开发文档

更新时间：2026-06-14

## 功能定位

`Image Generation / Artifact Return` 负责把模型侧结构化 image intent、安全约束和 artifact metadata 转成可审查、可追踪、可返回的图片生成结果引用。图片像素生成由独立 image adapter 和 backend 承接，不并入 `RadishMind-Core` 主模型职责。

## 当前状态

- Image Path 已完成 adapter handshake / safety gate、artifact return runbook、安全 runbook、backend adapter readiness、artifact runtime mapping readiness、store / binary reader boundary readiness、metadata-only runtime mapper、response consumer 和 `coerce_response_document` metadata-only response builder runtime integration。
- runtime integration 只从 request artifact metadata 发现 `image_generation_artifact`，通过 mapper / consumer 合并到现有 `CopilotResponse.citations` artifact citation。
- 当前不改 `CopilotResponse` schema，不读取 artifact 二进制，不创建 artifact store、binary reader、public URL resolver、backend adapter，不调用真实生图 backend，不生成或上传图片。

## 设计边界

- artifact metadata 只允许作为 metadata-only reference 返回。
- `blocked / failed / pending_review` artifact 不能进入成功 response。
- public URL、signed URL、binary payload、provider raw dump、hash / mime / dimensions mismatch、safety review missing 或 provenance missing 都必须 fail closed。
- response builder 接线不等于 artifact store、public delivery 或 backend generation ready。

## 下一批开发方向

1. 后续如果继续推进，必须在本功能文档中选择单一方向：artifact store、binary reader、public URL resolver 或 backend adapter。
2. store / reader / public URL / backend adapter 不能并行打开。
3. 真实 backend call 需要独立 adapter design、credential/profile boundary、safety gate、timeout/failure taxonomy 和 no binary leak 验证。
4. 普通 metadata mapping 文案和 runbook 调整复用现有 checker 与 fast baseline。

## 验收方式

- metadata-only runtime：runtime unit tests、image artifact checker、fast baseline。
- store / reader：hash / mime / dimensions revalidation、binary leak negative tests、no side effects checks。
- backend adapter：credential boundary、safety gate、timeout/failure taxonomy、no upload by default 和全量仓库验证。

# Image Generation / Artifact Return 设计与开发文档

更新时间：2026-07-25

状态：`image_adapter_controlled_invocation_artifact_return_dev_test_v1_batch_b_completed_batch_c_review_required`

## 功能定位

`Image Generation / Artifact Return` 负责把模型侧结构化 image intent、安全约束和 artifact metadata 转成可审查、可追踪、可返回的图片生成结果引用。图片像素生成由独立 image adapter 和 backend 承接，不并入 `RadishMind-Core` 主模型职责。

## 当前状态

- Image Path 已完成 adapter handshake / safety gate、artifact return runbook、安全 runbook、backend adapter readiness、artifact runtime mapping readiness、store / binary reader boundary readiness、metadata-only runtime mapper、response consumer 和 `coerce_response_document` metadata-only response builder runtime integration。
- runtime integration 只从 request artifact metadata 发现 `image_generation_artifact`，通过 mapper / consumer 合并到现有 `CopilotResponse.citations` artifact citation。
- 批次 A 已完成受控调用纯领域 runtime；批次 B 已完成本机私有 content-addressed store、不可变 artifact ref、metadata-only lookup 与显式授权 binary reader。
- 当前仍不改 `CopilotResponse` schema，不创建 public URL resolver、backend adapter，不调用真实生图 backend，不生成或上传图片。

2026-07-25 已选择下一实现方向为“开发测试态 Image Adapter 受控调用”，不再继续派生 metadata-only readiness 链。首批只实现独立 Python 领域运行时：

- 严格读取既有 `image_generation` intent，不建立第二套 intent / backend request / artifact contract。
- 纯函数编译 `image_generation_backend_request`，backend profile、request id 与 timeout 由调用侧显式提供。
- 在任何 backend side effect 前执行 schema、预算、敏感材料、profile exact match 与 safety gate。
- backend client 只通过注入式协议调用一次；不自动 retry，不 fallback，不解析 credential、endpoint 或 model dir。
- backend 返回 artifact metadata 与由 transport 观察到的 hash / MIME / dimensions；adapter 重验 lineage、generation、safety、provenance 和观察值，再复用既有 mapper 形成 artifact citation / 内部 metadata reference。
- 本批不创建 HTTP API、Gateway route、store、binary reader、public URL、Web、生产 profile 或真实 backend client。

## 设计边界

- artifact metadata 只允许作为 metadata-only reference 返回。
- `blocked / failed / pending_review` artifact 不能进入成功 response。
- public URL、signed URL、binary payload、provider raw dump、hash / mime / dimensions mismatch、safety review missing 或 provenance missing 都必须 fail closed。
- response builder 接线不等于 artifact store、public delivery 或 backend generation ready。

## 下一批开发方向

1. [Image Adapter 受控调用与 artifact 返回（开发 / 测试态）v1 实施任务卡](../task-cards/image-adapter-controlled-invocation-artifact-return-dev-test-v1-plan.md)批次 A 已完成。
2. 批次 B 的 `local_private_artifact_storage` owner 已完成：二进制容器观察与私有存储按稳定职责拆分，仍属于同一边界。
3. 批次 C 进入实现前，只评审具体 backend client 或 profile / credential 配置中的一个方向；HTTP、Gateway、应用配置、API key、Session、Run History 与 Web 仍不进入范围。
4. 生产 backend call 仍需要独立 credential/profile resolver、endpoint/model-dir、moderation、安全复核、运行配置和发布声明，不能由开发测试态注入协议代替。
5. 普通 metadata mapping 文案和 runbook 调整继续复用现有测试与仓库基线，不恢复同层 checker 链。

### 批次 B 固定边界

- store root 只能由调用侧显式注入绝对本机路径；模块不读取环境变量，不选择 production object store。
- blob 以 sha256 和 canonical format 形成 content-addressed 相对路径；artifact ref 单独绑定 `artifact_id / artifact:// URI / digest / MIME / dimensions / format / size`，两者均不可变。
- 写入前严格校验既有 artifact schema、安全状态、provenance、canonical URI、payload 大小、sha256、MIME 与 dimensions；不接受 base64、provider raw payload、public URL 或 signed URL。
- lookup 只返回内部 storage handle 和已验证 metadata，不读取图片二进制；binary reader 默认拒绝，只有显式 `allow_binary_read=true` 才能打开一次内部流。
- reader 每次读取都重新计算 sha256，并重新识别 MIME 与 dimensions；校验通过后只把临时 stream 交给一次内部 consumer，不在 result、日志、引用或响应中返回 bytes。
- 本批不提供 delete / overwrite、upload、public URL resolver、retry / fallback、后台服务、repository selector、migration 或生产存储声明。

## 验收方式

- metadata-only runtime：runtime unit tests、image artifact checker、fast baseline。
- store / reader：hash / mime / dimensions revalidation、binary leak negative tests、no side effects checks。
- backend adapter：credential boundary、safety gate、timeout/failure taxonomy、no upload by default 和全量仓库验证。
- 受控调用批次 A：Python 相邻单元测试覆盖纯编译、单次调用、预算、安全、敏感材料、profile drift、timeout、artifact 观察值与 provenance；随后运行 fast / full 仓库门禁。

## 批次 A 完成结果

1. 新增 `services/runtime/image_generation_adapter.py`，复用三份既有 schema 和 metadata mapper；没有新增第二套 intent、backend request 或 artifact 类型。
2. `compile_image_backend_request` 对相同 intent、profile 和 request id 生成稳定请求；trace 固定包含 source request、intent 与 backend request lineage。
3. `invoke_image_generation` 在调用前完成 schema、UTF-8、尺寸 / 像素 / count、参数、列表、标识符、locale、敏感材料、profile exact match 和 low-risk safety gate；所有前置失败 `backend_call_count=0`。
4. 注入式 client 最多调用一次；timeout、unavailable 与未知异常不 retry / fallback，也不把异常原文、endpoint 或 credential 投影到结果。
5. 成功返回必须同时提供 schema-valid artifact metadata 与 transport 观察到的 sha256 / MIME / dimensions；adapter 重验 canonical URI、lineage、generation、title / purpose、safety 和 provenance 后才生成既有 artifact citation / metadata reference。
6. 11 项相邻单元测试已纳入 `services/runtime/tests` 和快速仓库入口；store lookup、binary read、upload、public URL、production storage、executor、confirmation、writeback 与 replay 均保持 0。

## 批次 B 完成结果

1. `services/runtime/image_artifact_binary_inspection.py` 使用标准库确定性观察 PNG / JPEG / WebP 容器、sha256、MIME、dimensions、format 与 size，不加载图片模型或第三方解码器。
2. `services/runtime/image_artifact_private_storage.py` 实现调用侧绝对路径注入、私有目录 / 文件权限、content-addressed blob、不可变 artifact ref、原子 no-overwrite 写入和 metadata-only lookup。
3. store 写入前复用 strict artifact schema 与既有 mapper，精确重验 canonical URI、低风险安全状态、provenance、hash、MIME、dimensions 和 format；相同 artifact 幂等，绑定漂移失败关闭。
4. binary reader 默认拒绝；显式授权后只读取一次，在 consumer 前再次重验容器、hash、MIME、dimensions、format 和 size。consumer 最多调用一次，result 不返回 bytes、绝对路径、base64 或 URL。
5. 15 项 storage 相邻测试覆盖三种格式、真实临时目录、并发幂等、payload / metadata 边界、ref / blob / symlink 篡改、读取授权、consumer 异常脱敏和零外部副作用；runtime 测试总数为 26。

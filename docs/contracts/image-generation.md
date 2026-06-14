# RadishMind 图片生成契约

更新时间：2026-06-14

## `RadishMind-Image Adapter` 第一版仓库级契约

图片生成能力通过独立 adapter / backend 提供。`RadishMind-Core` 只负责生成结构化意图、约束、风险确认和审查信息，不直接生成图片像素。

第一版 image generation intent 已落成仓库级可回归契约：

- Schema：`contracts/image-generation-intent.schema.json`
- Backend request schema：`contracts/image-generation-backend-request.schema.json`
- Artifact schema：`contracts/image-generation-artifact.schema.json`
- 最小 fixture：`scripts/checks/fixtures/image-generation-intent-basic.json`
- Backend request fixture：`scripts/checks/fixtures/image-generation-backend-request-basic.json`
- Artifact fixture：`scripts/checks/fixtures/image-generation-artifact-basic.json`
- Smoke：`scripts/check-image-generation-intent-contract.py`
- 最小评测 manifest：`scripts/checks/fixtures/image-generation-eval-manifest-v0.json`
- 评测 manifest smoke：`scripts/check-image-generation-eval-manifest.py`
- Handshake / safety gate fixture：`scripts/checks/fixtures/image-adapter-handshake-safety-gate-v1.json`
- Handshake / safety gate smoke：`scripts/check-image-adapter-handshake-safety-gate-v1.py`
- Artifact return runbook fixture：`scripts/checks/fixtures/image-artifact-return-runbook-evidence-v1.json`
- Artifact return runbook smoke：`scripts/check-image-artifact-return-runbook-evidence-v1.py`
- Safety runbook fixture：`scripts/checks/fixtures/image-safety-runbook-evidence-v1.json`
- Safety runbook smoke：`scripts/check-image-safety-runbook-evidence-v1.py`
- Backend adapter readiness fixture：`scripts/checks/fixtures/image-backend-adapter-readiness-evidence-v1.json`
- Backend adapter readiness smoke：`scripts/check-image-backend-adapter-readiness-evidence-v1.py`
- Artifact runtime mapping readiness fixture：`scripts/checks/fixtures/image-artifact-runtime-mapping-readiness-v1.json`
- Artifact runtime mapping readiness smoke：`scripts/check-image-artifact-runtime-mapping-readiness-v1.py`
- Artifact runtime mapping implementation entry review fixture：`scripts/checks/fixtures/image-artifact-runtime-mapping-implementation-entry-review-v1.json`
- Artifact runtime mapping implementation entry review smoke：`scripts/check-image-artifact-runtime-mapping-implementation-entry-review-v1.py`
- Artifact store / binary reader boundary readiness fixture：`scripts/checks/fixtures/image-artifact-store-binary-reader-boundary-readiness-v1.json`
- Artifact store / binary reader boundary readiness smoke：`scripts/check-image-artifact-store-binary-reader-boundary-readiness-v1.py`
- Artifact runtime mapper implementation plan fixture：`scripts/checks/fixtures/image-artifact-runtime-mapper-implementation-plan-v1.json`
- Artifact runtime mapper implementation plan smoke：`scripts/check-image-artifact-runtime-mapper-implementation-plan-v1.py`
- Artifact runtime mapper implementation entry fixture：`scripts/checks/fixtures/image-artifact-runtime-mapper-implementation-entry-v1.json`
- Artifact runtime mapper implementation entry smoke：`scripts/check-image-artifact-runtime-mapper-implementation-entry-v1.py`
- Artifact runtime mapper implementation fixture：`scripts/checks/fixtures/image-artifact-runtime-mapper-implementation-v1.json`
- Artifact runtime mapper implementation smoke：`scripts/check-image-artifact-runtime-mapper-implementation-v1.py`
- Artifact runtime mapper runtime implementation fixture：`scripts/checks/fixtures/image-artifact-runtime-mapper-runtime-implementation-v1.json`
- Artifact runtime mapper runtime implementation smoke：`scripts/check-image-artifact-runtime-mapper-runtime-implementation-v1.py`
- Artifact runtime mapper response consumer integration review fixture：`scripts/checks/fixtures/image-artifact-runtime-mapper-response-consumer-integration-review-v1.json`
- Artifact runtime mapper response consumer integration review smoke：`scripts/check-image-artifact-runtime-mapper-response-consumer-integration-review-v1.py`
- Artifact response consumer implementation readiness fixture：`scripts/checks/fixtures/image-artifact-response-consumer-implementation-readiness-v1.json`
- Artifact response consumer implementation readiness smoke：`scripts/check-image-artifact-response-consumer-implementation-readiness-v1.py`
- Artifact response consumer implementation fixture：`scripts/checks/fixtures/image-artifact-response-consumer-implementation-v1.json`
- Artifact response consumer implementation smoke：`scripts/check-image-artifact-response-consumer-implementation-v1.py`
- Artifact response consumer runtime implementation fixture：`scripts/checks/fixtures/image-artifact-response-consumer-runtime-implementation-v1.json`
- Artifact response consumer runtime implementation smoke：`scripts/check-image-artifact-response-consumer-runtime-implementation-v1.py`
- Artifact response builder integration entry review fixture：`scripts/checks/fixtures/image-artifact-response-builder-integration-entry-review-v1.json`
- Artifact response builder integration entry review smoke：`scripts/check-image-artifact-response-builder-integration-entry-review-v1.py`
- Artifact response builder integration fixture：`scripts/checks/fixtures/image-artifact-response-builder-integration-v1.json`
- Artifact response builder integration smoke：`scripts/check-image-artifact-response-builder-integration-v1.py`
- Artifact response builder runtime integration entry review fixture：`scripts/checks/fixtures/image-artifact-response-builder-runtime-integration-entry-review-v1.json`
- Artifact response builder runtime integration entry review smoke：`scripts/check-image-artifact-response-builder-runtime-integration-entry-review-v1.py`
- Artifact response builder runtime integration implementation fixture：`scripts/checks/fixtures/image-artifact-response-builder-runtime-integration-implementation-v1.json`
- Artifact response builder runtime integration implementation smoke：`scripts/check-image-artifact-response-builder-runtime-integration-implementation-v1.py`

当前 schema 固定的是 `RadishMind-Core -> RadishMind-Image Adapter -> Image Generation Backend -> artifact metadata` 的最小结构化链路，不承诺具体 backend 常驻、权重下载、图片质量或像素生成实现。第一版 intent 结构如下：

```json
{
  "schema_version": 1,
  "intent_id": "optional-string",
  "kind": "image_generation",
  "source_request_id": "copilot-request-id",
  "prompt": {
    "positive": "生成目标的正向描述",
    "negative": "需要避免的内容",
    "locale": "zh-CN"
  },
  "output": {
    "width": 1024,
    "height": 1024,
    "count": 1,
    "format": "png"
  },
  "style": {
    "preset": "diagram",
    "reference_artifact_ids": []
  },
  "constraints": {
    "must_include": [],
    "must_avoid": [],
    "edit_artifact_id": null,
    "mask_artifact_id": null
  },
  "backend": {
    "preferred": "sd15",
    "seed": 12345,
    "steps": 24,
    "guidance_scale": 7.0
  },
  "safety": {
    "requires_confirmation": false,
    "risk_level": "low",
    "review_notes": []
  },
  "artifact_metadata": {
    "proposed_title": "generated-image",
    "purpose": "visual_reference",
    "trace_ids": []
  }
}
```

字段边界：

- `prompt` 是主模型输出给 image adapter 的自然语言生成意图，不应直接等同于最终 backend prompt；adapter 可以做模板化、翻译或安全改写
- `constraints` 用于表达必须包含 / 避免 / 局部编辑输入，不承载业务真相源
- `backend` 只表达推理偏好；实际 backend 可以根据部署环境降级或忽略不支持的参数，但必须在 artifact metadata 中记录
- `safety.requires_confirmation=true` 时，调用侧必须先展示 intent，不得直接提交给生图 backend；当前 smoke 已把这一点纳入回归检查
- 生成结果应以 artifact 形式返回，并保留来源 intent、backend、seed、尺寸、格式和审计 metadata

第一版 backend request 只表达 Adapter 对 backend 的一次调度请求：

```json
{
  "schema_version": 1,
  "kind": "image_generation_backend_request",
  "request_id": "image-backend-request-id",
  "intent_id": "image-intent-id",
  "backend": {
    "id": "sd15",
    "model": "sd15-local-or-service",
    "adapter_profile": "diagram-default"
  },
  "prompt": {
    "positive": "adapter transformed positive prompt",
    "negative": "adapter transformed negative prompt",
    "locale": "en-US",
    "transformed_from_intent": true
  },
  "output": {
    "width": 1024,
    "height": 1024,
    "count": 1,
    "format": "png"
  },
  "parameters": {
    "seed": 12345,
    "steps": 24,
    "guidance_scale": 7.0
  },
  "inputs": {
    "reference_artifact_ids": [],
    "edit_artifact_id": null,
    "mask_artifact_id": null
  },
  "constraints": {
    "must_include": [],
    "must_avoid": [],
    "style_preset": "diagram"
  },
  "safety": {
    "gate": "approved_for_backend",
    "requires_confirmation": false,
    "risk_level": "low",
    "review_notes": []
  },
  "trace": {
    "source_request_id": "copilot-request-id",
    "trace_ids": []
  }
}
```

第一版 artifact metadata 只表达 backend 产物回到 `RadishMind` 后的可审计索引，不提交图片像素本体：

```json
{
  "schema_version": 1,
  "kind": "image_generation_artifact",
  "artifact_id": "image-artifact-id",
  "intent_id": "image-intent-id",
  "backend_request_id": "image-backend-request-id",
  "status": "generated",
  "artifact": {
    "uri": "artifact://radishmind/generated/image.png",
    "mime_type": "image/png",
    "width": 1024,
    "height": 1024,
    "format": "png",
    "sha256": "64-char-lowercase-hex",
    "title": "generated-image",
    "purpose": "visual_reference"
  },
  "generation": {
    "backend_id": "sd15",
    "model": "sd15-local-or-service",
    "seed": 12345,
    "steps": 24,
    "guidance_scale": 7.0
  },
  "safety": {
    "risk_level": "low",
    "requires_confirmation": false,
    "review_status": "not_required",
    "review_notes": []
  },
  "provenance": {
    "source_request_id": "copilot-request-id",
    "trace_ids": [],
    "backend_request_id": "image-backend-request-id",
    "intent_id": "image-intent-id"
  },
  "created_at": "2026-04-29T00:00:00Z"
}
```

链路边界：

- backend request 中的 seed、steps、guidance、尺寸、格式、输入 artifact 和约束必须可追溯到 intent 或 adapter profile
- `safety.gate=blocked_requires_confirmation` 时不得提交给真实 backend；当前 smoke 会构造该负向路径
- artifact metadata 必须保留 `intent_id`、`backend_request_id`、backend/model、seed、尺寸、格式、hash 和 provenance
- 当前不提交生成图片像素，也不把 artifact URI 当作已可公开访问的业务 URL

### 图片生成最小评测 manifest

当前 `scripts/checks/fixtures/image-generation-eval-manifest-v0.json` 是 `RadishMind-Image Adapter` 的首个评测 manifest 草案，状态为 `draft`。它的目标不是评价真实图片质量，而是把第一版 adapter 链路纳入仓库级回归：

- `structured_intent`：intent 必须符合 schema，并保留 prompt、output、style、constraints、backend 和 safety 的最小结构
- `backend_request_mapping`：backend request 的 backend、seed、steps、guidance、尺寸、输入 artifact、约束和 safety 必须能追溯到 intent
- `artifact_metadata`：artifact metadata 必须保留 intent/backend request 反链、尺寸、格式、hash、backend/model、seed 和 safety review
- `safety_gate`：`requires_confirmation=true` 的 intent 必须被 `blocked_requires_confirmation` 阻断，不得视为可提交 backend
- `provenance`：source request、intent 和 backend request 必须进入 trace/provenance

该 manifest 明确排除 `image_pixel_quality`、真实 backend 延迟、provider 渲染差异和模型权重质量；执行策略固定为不调用真实 backend、不生成图片、不下载模型、不启动训练。committed 资产只允许 manifest、小型 JSON fixture 和 expected summary，图片像素、provider raw dump、权重、checkpoint 与大规模 JSONL 均不得入仓。

### Adapter handshake / safety gate evidence

`image-adapter-handshake-safety-gate-v1` 已把 adapter handoff 和 safety gate 固定为 `image_adapter_handshake_safety_gate_defined`。该证据层只复用既有 intent、backend request、artifact metadata schema 和 eval manifest，说明五段关系：

- `RadishMind-Core -> RadishMind-Image Adapter` 只交付结构化 `image_generation` intent，intent 本身不直接触发 backend。
- Adapter 必须先执行 safety gate，只有低风险且 `requires_confirmation=false` 的 intent 才能映射到 `approved_for_backend` backend request；即便如此，当前仓库仍不调用真实 backend。
- `requires_confirmation=true`、高风险或 policy unknown 场景必须在 backend submission 前停住，并保持 `blocked_requires_confirmation` 或等价 blocked gate。
- backend result 只允许回写 `image_generation_artifact` metadata，当前不提交图片像素、provider raw dump、公开 URL 或 production artifact storage。
- artifact metadata 回到上层响应前仍是 metadata-only reference，不实现 runtime response mapping、artifact upload、executor、confirmation decision、writeback 或 replay。

### Artifact return runbook evidence

`image-artifact-return-runbook-evidence-v1` 已把 artifact 返回链路 runbook 固定为 `image_artifact_return_runbook_evidence_defined`。该证据层只定义 metadata reference，不改 `CopilotResponse` schema，不实现 runtime mapping，也不上传 artifact。

返回证据要求：

- metadata reference 必须来自 `image_generation_artifact`，保留 `artifact_id`、`intent_id`、`backend_request_id`、`artifact://` URI、mime type、尺寸、格式、hash、title、purpose、backend/model、seed、safety review、provenance 和 `created_at`。
- `artifact://` 只表示仓库契约里的 artifact reference，不是 public URL、signed URL、production storage location 或 binary download endpoint。
- 上层响应当前只能消费 metadata reference 证据；`pixel_payload`、`base64_image`、provider raw response、storage write result、executor ref、writeback ref 和 replay ref 都不得进入该返回层。
- 失败分类固定为 `image_backend_unavailable`、`image_artifact_metadata_missing`、`image_artifact_hash_mismatch` 和 `image_artifact_safety_blocked`；这些失败只能返回 blocked / failed metadata 状态，不触发自动 retry、fallback execution 或真实 backend health 声明。

### Safety runbook evidence

`image-safety-runbook-evidence-v1` 已把图片路径安全审查 runbook 固定为 `image_safety_runbook_evidence_defined`。该证据层只定义 intent precheck、adapter gate、artifact safety review 和 failure taxonomy，不接 moderation provider，不实现 runtime policy engine，也不调用真实生图 backend。

安全证据要求：

- intent precheck 必须读取 `prompt.positive`、`prompt.negative`、constraints、`safety.requires_confirmation`、`safety.risk_level` 和 `review_notes`；这些字段只进入 safety runbook 证据，不触发 backend。
- Adapter gate 必须在 backend request 前处理 `requires_confirmation=true`、high risk 和 policy unknown，输出 `blocked_requires_confirmation` 或等价 blocked 状态；不能写成自动提交 backend。
- artifact safety review 必须保留 `risk_level`、`requires_confirmation`、`review_status`、`review_notes`、hash、`artifact://` URI 和 provenance；`pending_review` 与 `blocked` 都不能返回成功 artifact reference。
- 失败分类固定为 `image_prompt_policy_unknown`、`image_intent_requires_confirmation`、`image_intent_high_risk`、`image_backend_safety_gate_blocked`、`image_backend_unavailable`、`image_artifact_safety_pending_review` 和 `image_artifact_safety_blocked`；这些失败不触发 retry loop、fallback execution、moderation provider 调用或真实 backend health 声明。

### Backend adapter readiness evidence

`image-backend-adapter-readiness-evidence-v1` 已把未来生图 backend adapter readiness 固定为 `image_backend_adapter_readiness_defined`。该证据层只定义 backend profile、credential / model-dir / endpoint 准入、failure envelope、artifact metadata 验收和 future smoke contract，不创建真实 backend client，也不调用真实生图 backend。

准入证据要求：

- backend adapter 必须消费 `image_generation_backend_request` 的 backend id、model、adapter profile、prompt、output、parameters、safety gate 和 trace，不得绕过 `image-safety-runbook-evidence-v1`。
- profile / credential / model dir / endpoint / timeout budget 当前都只是 implementation gate；缺失时必须映射到 fail-closed failure code，不能自动降级为可调用 backend。
- backend response 未来只能进入 `image_generation_artifact` metadata；必须校验 `artifact://` URI、hash、backend/model、provenance 和 safety review，不接收 pixel payload、provider raw response 或 public URL。
- 失败分类固定为 `image_backend_profile_missing`、`image_backend_credential_missing`、`image_backend_model_dir_missing`、`image_backend_endpoint_unavailable`、`image_backend_timeout`、`image_backend_safety_gate_blocked`、`image_backend_invalid_artifact_metadata`、`image_backend_artifact_hash_mismatch` 和 `image_backend_response_untrusted`；这些失败不触发 retry loop、fallback execution 或 success artifact reference。

### Artifact runtime mapping readiness

`image-artifact-runtime-mapping-readiness-v1` 已把 `image_generation_artifact` metadata 到未来 `CopilotResponse` artifact citation / metadata reference 的准入证据固定为 `image_artifact_runtime_mapping_readiness_defined`。该证据层只定义 future mapping gate，不改 `CopilotResponse` schema，不实现 runtime mapper，不创建 artifact store、public URL resolver 或 binary reader。

映射准入要求：

- 成功 reference 必须保留 `artifact://` URI、`sha256` hash、mime type、width / height、format、title、purpose、backend/model/seed、safety review、provenance 和 `created_at`。
- `artifact://` 仍只是 metadata reference，不是 public URL、signed URL、production storage location 或 binary download endpoint。
- 只有 `status=generated` 且 `safety.review_status=not_required` 或 `reviewed_pass` 的 artifact 才能进入未来成功 response artifact citation；`blocked / failed / pending_review` artifact 不得进入成功 response。
- invalid metadata、hash mismatch、public URL claim、binary payload 和 provider raw dump 必须 fail closed，映射到 `image_artifact_invalid_metadata`、`image_artifact_hash_mismatch`、`image_artifact_public_url_claim`、`image_artifact_binary_payload_rejected` 或 `image_artifact_provider_raw_dump_rejected`。
- 该 readiness 依赖 `image-backend-adapter-readiness-evidence-v1`、`image-safety-runbook-evidence-v1` 和 `image-artifact-return-runbook-evidence-v1`，不绕过 backend adapter readiness、artifact return runbook 或 safety runbook。

### Artifact runtime mapping implementation entry review

`image-artifact-runtime-mapping-implementation-entry-review-v1` 已把 runtime mapping 实现入口复核固定为 `image_artifact_runtime_mapping_entry_review_defined`。结论是 readiness 证据可被消费，但实现入口仍未打开；当前只能把下一步指向 artifact store / binary reader boundary readiness。

入口评审要求：

- checker 必须跨读 runtime mapping readiness、artifact return runbook、safety runbook 和 backend adapter readiness，确认这些证据不会被提升为 runtime mapper implementation ready。
- runtime mapper、artifact store、binary reader、public URL resolver 和 backend adapter implementation 五类候选当前全部保持 `blocked`，不得创建对应 implementation task card 或 runtime artifact。
- 当前不改 `CopilotResponse` schema，不创建 artifact store / public URL / binary reader，不调用真实 backend，不生成图片，不上传 artifact，也不进入 executor、confirmation、writeback 或 replay。
- 后续若继续推进 Image Path，应先补 runtime mapper implementation entry review，再评估单一 runtime mapping 实现方向。

### Artifact store / binary reader boundary readiness

`image-artifact-store-binary-reader-boundary-readiness-v1` 已把 store / binary reader 边界准入固定为 `image_artifact_store_binary_reader_boundary_readiness_defined`。该证据层只定义 future artifact store ownership、`artifact://` 解析边界、hash / mime type / dimensions revalidation、binary payload redaction、public URL / signed URL 禁止策略和 failure taxonomy，不实现 artifact store、binary reader、public URL resolver 或 runtime mapper。

边界准入要求：

- artifact store 未来只能消费 artifact metadata reference，不接收 `pixel_payload`、`base64_image`、provider raw response、public URL 或 signed public URL。
- binary reader 未来必须在 artifact store lookup、`artifact://` scheme、sha256、mime type、dimensions、safety review 和 public URL policy 全部通过后才可读；当前仍不允许读取 artifact 二进制。
- `artifact://` 仍不是 public URL、signed URL、production storage path 或 binary download endpoint；public URL / signed URL 行为必须等待 production storage policy 和 expiry policy。
- store missing、binary reader missing、invalid URI、hash mismatch、mime mismatch、dimension mismatch、public URL claim、signed URL policy missing、binary payload、provider raw dump、pending / blocked safety review 和 provenance missing 都必须 fail closed。
- 该 readiness 只允许下一步进入 runtime mapper implementation plan 评审，不允许直接写 runtime mapper、artifact store、binary reader、backend adapter implementation 或 response schema 变更。

### Artifact runtime mapper implementation plan

`image-artifact-runtime-mapper-implementation-plan-v1` 已把 future runtime mapper 的实现计划证据固定为 `image_artifact_runtime_mapper_implementation_plan_defined`。该证据层只组织 mapper 输入、future CopilotResponse artifact citation / metadata reference 目标、成功 / blocked 映射、fail-closed plan、单一实现方向策略和下一步 runtime mapper implementation entry review 条件；不改 `CopilotResponse` schema，不实现 runtime mapper，不创建 artifact store、binary reader、public URL resolver、backend adapter implementation 或 production storage。

计划准入要求：

- mapper 未来只能消费 `image_generation_artifact` metadata、store / binary reader boundary contract、runtime mapping readiness cases、artifact return runbook、image safety runbook 和 backend adapter readiness failure envelope。
- 成功路径必须保留 `artifact://`、sha256、mime type、dimensions、safety review、provenance 和 metadata reference；`blocked / failed / pending_review` artifact 不得进入成功 response。
- invalid metadata、hash mismatch、mime mismatch、dimension mismatch、public URL claim、signed URL policy missing、binary payload、provider raw dump、artifact store missing / unavailable、binary reader missing / forbidden、safety review not passed 和 provenance missing 都必须 fail closed。
- 下一步只能进入 runtime mapper implementation entry review，由该 entry review 决定是否选择一个实现方向；当前仍不读取 artifact 二进制、不调用真实 backend、不生成图片、不上传 artifact、不启动 UI 或 dev server。

### Artifact runtime mapper implementation entry review

`image-artifact-runtime-mapper-implementation-entry-v1` 已把 runtime mapper implementation entry review 固定为 `image_artifact_runtime_mapper_implementation_entry_review_defined`。该证据层复核 implementation plan、store / binary reader boundary readiness、runtime mapping readiness、artifact return runbook、safety runbook 和 backend adapter readiness 后，只选择下一步允许创建 metadata-only runtime mapper implementation task card；不创建该 task card，不实现 runtime mapper，不创建 artifact store、binary reader、public URL resolver、backend adapter implementation 或 production storage。

入口复核要求：

- 下一张 runtime mapper implementation task card 只能覆盖 metadata-only mapper input validation、`image_generation_artifact` 到 future CopilotResponse artifact citation / metadata reference 的 projection、成功映射测试、fail-closed 测试和 no side effects smoke。
- artifact store、binary reader、public URL resolver 和 backend adapter implementation 继续 deferred，不得与 runtime mapper implementation task card 并行打开。
- 当前不允许 runtime mapping execution、artifact store lookup、artifact binary read、artifact upload、public / signed URL resolution、backend call、image generation、response schema change、executor、confirmation、writeback 或 replay。

### Artifact runtime mapper implementation task card

`image-artifact-runtime-mapper-implementation-v1` 已把 metadata-only runtime mapper implementation task card 固定为 `image_artifact_runtime_mapper_implementation_task_card_defined`。该证据层创建下一步 runtime mapper 代码实现前的任务卡、fixture 和 checker，明确后续代码只能消费 `image_generation_artifact` metadata，并投影到未来 CopilotResponse artifact citation / metadata reference；本切片仍不创建 runtime mapper 文件、不改 `CopilotResponse` schema、不创建 artifact store、binary reader、public URL resolver、backend adapter implementation 或 production storage。

实现任务卡要求：

- 后续 metadata-only runtime mapper 必须保留 `artifact://`、sha256、mime type、dimensions、safety review、provenance 和 metadata reference；`blocked / failed / pending_review` artifact 不得进入成功 response。
- invalid metadata、hash mismatch、mime mismatch、dimension mismatch、public URL claim、signed URL policy missing、binary payload、provider raw dump、missing store / reader、safety review not passed 和 provenance missing 都必须 fail closed。
- 下一步可以进入 runtime mapper 代码实现，但只能实现 metadata-only mapper input validation、artifact citation projection、metadata reference projection、成功映射测试、fail-closed 测试和 no side effects smoke；store / reader / public URL / backend adapter 仍需独立任务卡。

### Artifact runtime mapper runtime implementation

`image-artifact-runtime-mapper-runtime-implementation-v1` 已把 metadata-only runtime mapper 代码固定为 `image_artifact_runtime_mapper_runtime_implemented`。该实现由 `services/runtime/image_artifact_runtime_mapper.py` 承载，只消费 `image_generation_artifact` metadata，并输出 future CopilotResponse artifact citation / metadata reference；不读取 artifact 二进制、不查 artifact store、不解析 public URL、不调用真实 backend、不上传 artifact、不修改 `CopilotResponse` schema。

runtime 实现要求：

- 成功路径只允许 `generated + not_required` 与 `generated + reviewed_pass`，并保留 `artifact://`、sha256、mime type、dimensions、safety review、provenance 和 generation metadata。
- `blocked / failed / pending_review`、invalid metadata、hash mismatch、mime mismatch、dimension mismatch、public URL claim、signed URL policy missing、binary payload、provider raw dump、missing / unavailable store、missing / forbidden binary reader、safety review not passed 和 provenance missing 均返回 fail-closed failure code，不产生成功 citation。
- 后续若要接入真实 response consumer、artifact store、binary reader、public URL resolver 或 backend adapter，仍需独立入口、fixture、checker 和验证。

### Artifact runtime mapper response consumer integration review

`image-artifact-runtime-mapper-response-consumer-integration-review-v1` 已把 metadata-only runtime mapper 到未来 response consumer 的入口评审固定为 `image_artifact_runtime_mapper_response_consumer_integration_review_defined`。该证据层只确认未来消费应沿现有 `CopilotResponse.citations` 的 `kind=artifact` citation 形状和 mapper 返回的 `metadata_reference` 进行 metadata-only handoff；不改 `CopilotResponse` schema，不实现 response consumer，不修改 response builder，不创建 artifact store、binary reader、public URL resolver 或 backend adapter。

response consumer 集成评审要求：

- 只有 mapper 返回 `ok=true` 且 citation 符合 `CopilotResponse` citation schema 时，未来 consumer 才允许进入成功 response 路径。
- `metadata_reference` 仍是内部 metadata-only handoff，不得作为 public URL、signed URL、binary payload、provider raw dump 或 production storage claim 暴露。
- `blocked / failed / pending_review`、public URL claim、binary payload、provider raw dump 和其它 fail-closed case 不得生成成功 citation；现有 response builder 本切片仍保持未接线。

### Artifact response consumer implementation readiness

`image-artifact-response-consumer-implementation-readiness-v1` 已把 metadata-only response consumer implementation readiness 固定为 `image_artifact_response_consumer_implementation_readiness_defined`。该证据层只定义未来 `services/runtime/image_artifact_response_consumer.py` 与 `apply_image_artifact_reference_to_response` 的准入条件、输入输出、failure propagation test plan、禁止实现 artifact 和禁止接线项；当前不创建 response consumer 模块，不修改 response builder，不改 `CopilotResponse` schema，不实现 store、binary reader、public URL resolver 或 backend adapter。

response consumer implementation readiness 要求：

- 未来实现只能把 mapper 成功结果合并进 `CopilotResponse.citations` 的 `kind=artifact` citation，`metadata_reference` 继续保持内部 handoff。
- citation id 冲突、citation schema 不匹配、mapper failure 或 `metadata_reference` 外泄都必须 fail closed，且不能生成成功 citation。
- 本 readiness 切片继续证明现有 `inference_response`、`inference_support`、gateway 和 platform response route 未接入未来 consumer。

### Artifact response consumer implementation task card

`image-artifact-response-consumer-implementation-v1` 已把 metadata-only response consumer implementation task card 固定为 `image_artifact_response_consumer_implementation_task_card_defined`。该证据层只创建后续 runtime code 的任务卡、fixture 与 checker，定义 `services/runtime/image_artifact_response_consumer.py`、`apply_image_artifact_reference_to_response` 和 `ImageArtifactResponseConsumerResult` 的职责边界；当前不创建 response consumer 模块，不修改 response builder，不改 `CopilotResponse` schema，不实现 store、binary reader、public URL resolver 或 backend adapter。

response consumer implementation task card 要求：

- 后续 runtime code 只能消费已有 `CopilotResponse` draft 和 `ImageArtifactMappingResult`，成功时把 artifact citation 合并进 `CopilotResponse.citations`。
- `metadata_reference` 只保留为内部 handoff；citation id conflict、citation schema invalid、mapper failure 和 metadata reference leak 都必须 fail closed。
- 本切片只允许下一步进入 metadata-only response consumer runtime code；store / reader / public URL / backend adapter 仍不得并行打开。

### Artifact response consumer runtime implementation

`image-artifact-response-consumer-runtime-implementation-v1` 已把 metadata-only response consumer runtime 固定为 `image_artifact_response_consumer_runtime_implemented`。该实现由 `services/runtime/image_artifact_response_consumer.py` 承载，只消费已有 `CopilotResponse` draft 与 `ImageArtifactMappingResult`，成功时把 artifact citation append 到 `CopilotResponse.citations`，并把 `metadata_reference` 作为内部 handoff 随 result 返回；不改 `CopilotResponse` schema，不修改现有 response builder，不接 artifact store、binary reader、public URL resolver、gateway、platform HTTP route 或 backend adapter。

response consumer runtime implementation 要求：

- 成功路径必须保留已有 citations 顺序，不原地修改输入 response draft 或 mapper citation，输出 response 仍通过 `contracts/copilot-response.schema.json`。
- mapper failure、citation id conflict、citation schema invalid 和 metadata reference leak 必须 fail closed，且不产生成功 citation。
- side-effect counters 继续保持为零；现有 `inference_response`、`inference_support`、gateway 和 platform response route 不得引用 response consumer runtime。

### Artifact response builder integration entry review

`image-artifact-response-builder-integration-entry-review-v1` 已把 metadata-only response builder integration entry review 固定为 `image_artifact_response_builder_integration_entry_review_defined`。该证据层只评审未来是否允许把 response consumer runtime 接入现有 response builder，并选择 `services/runtime/inference_response.py#coerce_response_document` 的 response normalization / schema validation 边界作为未来候选；当前不修改 response builder，不改 `CopilotResponse` schema，不接 gateway、platform route、artifact store、binary reader、public URL resolver 或 backend adapter。

response builder integration entry review 要求：

- 未来接线前必须证明 mapper success、consumer success、post-merge `CopilotResponse` schema validation、citation conflict fail-closed 和 `metadata_reference` internal-only 都成立。
- `services/runtime/inference_support.py#build_citations` 只负责请求上下文 citation，不承接生成 artifact runtime handoff；gateway 和 platform northbound route 也不承接 metadata merge。
- 本切片只允许下一步创建 response builder integration task card；仍不得直接修改 `inference_response`、`inference_support`、gateway 或 platform route。

### Artifact response builder integration task card

`image-artifact-response-builder-integration-v1` 已把 metadata-only response builder integration task card 固定为 `image_artifact_response_builder_integration_task_card_defined`。该证据层只定义未来接入 `services/runtime/inference_response.py#coerce_response_document` 的 exact hook placement、request artifact metadata discovery input、post-merge `CopilotResponse` schema validation、failure propagation 和 no side effects；当前不修改 response builder，不改 `CopilotResponse` schema，不接 artifact store、binary reader、public URL resolver、gateway、platform route 或 backend adapter。

response builder integration task card 要求：

- 未来 hook placement 固定为 `coerce_response_document` 内 response top-level filtering 之后、现有 `validate_response_document(coerced)` 之前。
- v1 只允许从 `copilot_request.artifacts[*].metadata.image_generation_artifact` 发现 request-side artifact metadata；缺失 metadata 时保持 no-op，发现 metadata 时必须先校验 `image_generation_artifact` schema，再走 mapper / consumer。
- mapper failure、consumer failure 和 post-merge schema validation failure 必须 fail closed，不触发 retry、fallback execution、backend call、artifact upload 或 public URL resolution。

### Artifact response builder runtime integration entry review

`image-artifact-response-builder-runtime-integration-entry-review-v1` 已把 metadata-only response builder runtime integration entry review 固定为 `image_artifact_response_builder_runtime_integration_entry_review_defined`。该证据层只判断是否允许下一步创建 single response builder runtime integration implementation 任务卡，结论是只允许单一选择 `services/runtime/inference_response.py#coerce_response_document`，继续复用既有函数签名、exact hook placement、request metadata discovery、mapper / consumer handoff 和 post-merge schema validation；当前不修改 response builder，不改 `CopilotResponse` schema，不接 artifact store、binary reader、public URL resolver、gateway、platform route 或 backend adapter。

response builder runtime integration entry review 要求：

- 下一步只允许创建 `image-artifact-response-builder-runtime-integration-implementation-v1` 任务卡，不在本切片实现 runtime code。
- checker 必须证明当前 `inference_response.py`、`inference_support.py`、gateway 和 platform route 仍未接入 image artifact consumer。
- store、binary reader、public URL、signed URL、backend adapter、artifact upload、production storage、真实 backend call、图片生成、executor、confirmation、writeback 和 replay 继续 deferred。

### Artifact response builder runtime integration implementation

`image-artifact-response-builder-runtime-integration-implementation-v1` 已把 metadata-only response builder runtime integration 固定为 `image_artifact_response_builder_runtime_integration_implemented`。该实现只在 `services/runtime/inference_response.py#coerce_response_document` 内接入 hook，覆盖 request metadata discovery、mapper / consumer merge pipeline、multiple metadata ordering、failure propagation、runtime test coverage 和 no side effects；该 hook 完成后不继续扩同层 Image gate。当前仍不改 `CopilotResponse` schema，不接 artifact store、binary reader、public URL resolver、gateway、platform route 或 backend adapter。

response builder runtime integration implementation 要求：

- runtime code 只修改 `services/runtime/inference_response.py`，并保持 `coerce_response_document(document, copilot_request, raw_text)` 签名不变。
- metadata discovery 只能从 `copilot_request.artifacts[*].metadata.image_generation_artifact` 读取，缺失 metadata no-op，多个 metadata 必须按 request artifact 顺序合并。
- metadata schema invalid、mapper failure、consumer failure 和 post-merge schema invalid 必须 fail closed，不触发 retry、fallback execution、binary read、store lookup、artifact upload、public URL resolution 或 backend call。

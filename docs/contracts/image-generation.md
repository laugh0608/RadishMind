# RadishMind 图片生成契约

更新时间：2026-05-12

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

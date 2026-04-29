# RadishMind 跨项目集成契约草案

更新时间：2026-04-29

## 文档目的

本文档用于冻结 `RadishMind` 与上层项目之间的第一版通用协议口径。

当前目标不是一次性定死全部字段，而是先建立足够稳定的抽象，避免 `RadishFlow` 和 `Radish` 各自演化出不兼容的接入方式。

当前文档口径已经同步落成仓库内真实契约文件：

- `contracts/copilot-request.schema.json`
- `contracts/copilot-response.schema.json`
- `contracts/copilot-gateway-envelope.schema.json`
- `contracts/copilot-training-sample.schema.json`
- `contracts/image-generation-intent.schema.json`

当前协议原则已经根据真实仓库收口为：

- 统一骨架 + 项目专属上下文块
- 结构化 JSON 优先，图像和附件作为 artifact 补充
- 默认 advisory mode，不做直接写回
- 所有高风险输出都必须带 `requires_confirmation`

## 当前服务/API 接入切片

从 `suggest_flowsheet_edits v93` 与 `suggest_ghost_completion v25` 收口后，下一步不再默认通过继续跑真实样本推进，而是把现有协议和评测资产上提为服务/API 接入门禁。

当前最小切片建议固定为：

1. 上层或本地 smoke 提供 schema-valid `CopilotRequest`
2. `Copilot Gateway / Task Router` 校验 `project / task / schema_version`
3. Gateway 调用现有 `services/runtime/inference.py` 路径生成 `CopilotResponse`
4. Response Builder 保持统一 `risk_level / requires_confirmation / citations / proposed_actions`
5. 服务层只输出 advisory response，不写回 `RadishFlow` 或 `Radish` 真相源
6. 同一请求必须能复用既有 eval regression、candidate record audit 与治理报表作为验收门禁

真实 provider capture 只在服务/API 或集成演示暴露现有样本无法覆盖的新 drift 时触发；触发前应先写清楚新假设、覆盖缺口和退出条件。

### `CopilotGatewayEnvelope` 调用口径

当前 gateway 层返回值统一采用 `CopilotGatewayEnvelope`，schema 真相源为 `contracts/copilot-gateway-envelope.schema.json`。

首个对外调用形态当前固定为**进程内 Python API**：

```python
from services.gateway import GatewayOptions, handle_copilot_request

envelope = handle_copilot_request(
    copilot_request,
    options=GatewayOptions(provider="mock"),
)
```

HTTP JSON 暂时只保留为后续包装形态；在长驻服务、鉴权、端口和部署生命周期尚未进入当前切片前，不把 HTTP server 作为 `M3` 的默认推进项。未来若添加 HTTP 包装，也必须复用同一个 `CopilotGatewayEnvelope` 语义，而不是引入第二套响应协议。

调用侧口径建议固定为：

- 上层提交 schema-valid `CopilotRequest`，不直接调用任务 runtime 或 provider
- Gateway 返回 `schema_version / status / request_id / project / task / response / error / metadata`
- 当 `status=ok` 或 `status=partial` 时，`response` 必须存在，并继续按 `contracts/copilot-response.schema.json` 校验和消费
- 当 `status=failed` 时，调用侧必须优先读取 `error.code` 与 `error.message`；若同时存在 `response`，它也只能作为 failed advisory response 展示或记录，不能转成可执行动作
- `metadata.route` 固定表达 `project/task`，用于调用侧日志、审计和路由观测
- `metadata.provider` 只表达本次 gateway 使用的 provider/profile/model 摘要，不应被上层当作业务结果
- `metadata.advisory_only` 必须保持 `true`；上层不得绕过人工确认或规则层复核直接执行 `response.proposed_actions`
- `metadata.request_validated` 与 `metadata.response_validated` 用于确认 gateway 是否完成请求和响应 schema 校验；生产前调用侧应把任一 `false` 视为异常观测信号
- 对 `RadishFlow suggest_flowsheet_edits`，即使 gateway 返回 `partial`，只要 `response.requires_confirmation=true`，上层也只能呈现候选建议，不能写回 `FlowsheetDocument`

当前仓库用两层 smoke 固定这条调用口径：

- `scripts/check-gateway-service-smoke.py --check-summary scripts/checks/fixtures/gateway-service-smoke-summary.json`：固定进程内 Python API 的成功调用、schema-valid 但 unsupported route、schema-invalid request 三类 envelope 行为
- `scripts/run-radishflow-gateway-demo.py --manifest scripts/checks/fixtures/radishflow-gateway-demo-fixtures.json --check-summary scripts/checks/fixtures/radishflow-gateway-demo-summary.json --check`：固定 `RadishFlow export -> adapter/request -> gateway envelope` 的服务/API 集成 smoke

这些 summary 只锁定上层依赖的 envelope 行为字段，不锁死完整自然语言响应。

### `RadishFlow` UI 消费口径

`RadishFlow` 调用侧在消费 `CopilotGatewayEnvelope` 时，应先把 envelope 映射为 UI 可展示、可审计、不可直接执行的 consumption view：

- `status=ok/partial` 且存在 `response.proposed_actions` 时，可展示候选提案卡片，但必须保留 `requires_confirmation`
- `candidate_edit` 只能显示为 advisory proposal，后续真实修改仍交由上层命令层和人工确认流处理
- `status=failed` 时，UI 应优先展示 `error.code / error.message`，并把同时存在的 failed `response` 仅作为说明或日志，不转成候选动作
- `metadata.provider / route / request_validated / response_validated / advisory_only` 应进入调用侧审计或诊断信息，不作为业务真相源
- 任一消费路径都不得直接写回 `FlowsheetDocument`

当前仓库用 `scripts/check-radishflow-gateway-ui-consumption.py --check-summary scripts/checks/fixtures/radishflow-gateway-ui-consumption-summary.json` 固定这层消费口径。该 summary 覆盖 proposal-ready、unsupported route 与 schema-invalid 三类路径，确保上层可消费视图始终保持 advisory-only。

### `RadishFlow` 候选编辑 handoff 口径

当用户在 UI 中确认某条 `candidate_edit` 后，调用侧仍不应直接执行 patch，而应生成命令层可接收的候选 handoff：

- 只有 `kind=candidate_edit` 且 `requires_confirmation=true` 的动作可以进入 handoff
- handoff 输出是 `radishflow_candidate_edit_proposal`，不是已执行命令
- handoff 必须保留 `source_request_id`、`source_route`、`action_index`、`target`、`patch_keys`、`risk_level` 与 `citation_ids`
- handoff 必须显式标记 `requires_human_confirmation=true` 与 `can_execute=false`
- `status=failed`、schema-invalid request 或没有合格 `candidate_edit` 的响应不得产生 handoff

当前仓库用 `scripts/check-radishflow-candidate-edit-handoff.py --check-summary scripts/checks/fixtures/radishflow-candidate-edit-handoff-summary.json` 固定这层 handoff 口径。该 summary 只证明候选动作可被安全交接给上层命令层，不代表本仓库已经执行或实现 `RadishFlow` 真实命令。

### 当前上层接入等待口径

`RadishFlow` 与 `Radish` 暂时都还没有进入真实模型 / Agent 接入阶段，因此当前不把真实上层调用位置作为 `RadishMind` 的阻塞项。

在上层项目准备好之前：

- 当前已落地的 gateway smoke、`RadishFlow` UI consumption summary 与 candidate edit handoff summary 视为未来接入验收门禁
- 不继续为尚不存在的上层 UI / 命令层扩展更多本仓库内模拟 summary
- HTTP JSON 包装仍保留为后续部署形态，不在没有真实消费方前强行提前实现
- 当前开发重心转向 `RadishMind-Core` 基座评估、训练 / 蒸馏样本格式和 `RadishMind-Image Adapter` schema
- 未来 `RadishFlow` 或 `Radish` 准备接入时，应优先复用现有 `CopilotRequest`、`CopilotResponse` 与 `CopilotGatewayEnvelope`，而不是新增第二套协议

## `CopilotTrainingSample` 训练 / 蒸馏样本契约

`RadishMind-Core` 的首版训练 / 蒸馏样本以 `CopilotRequest -> CopilotResponse` 为核心，不引入第二套任务协议。样本 wrapper 只负责记录训练模式、teacher/source、训练字段、质量门禁和来源 metadata。

当前已落成仓库级可回归契约：

- Schema：`contracts/copilot-training-sample.schema.json`
- 最小 fixture：`scripts/checks/fixtures/copilot-training-sample-basic.json`
- Smoke：`scripts/check-copilot-training-sample-contract.py`
- Eval 转换入口：`scripts/build-copilot-training-samples.py`
- 首批转换清单：`scripts/checks/fixtures/copilot-training-sample-conversion-manifest.json`
- 首批转换 summary：`scripts/checks/fixtures/copilot-training-sample-conversion-summary.json`
- 首批 candidate record 转换清单：`scripts/checks/fixtures/copilot-training-sample-candidate-record-conversion-manifest.json`
- 首批 candidate record 转换 summary：`scripts/checks/fixtures/copilot-training-sample-candidate-record-conversion-summary.json`
- 训练集合治理 manifest 草案：`training/datasets/copilot-training-dataset-governance-v0.json`
- 训练集合人工复核记录模板：`training/datasets/copilot-training-review-record-v0.json`
- 训练集合 offline eval holdout split：`training/datasets/copilot-training-holdout-split-v0.json`
- 训练集合治理 smoke：`scripts/check-copilot-training-dataset-governance.py`

当前 eval 转换入口先只使用 committed eval 样本中的 `input_request` 与 `golden_response`，不调用 provider、不下载模型、不读取 candidate record。首批转换固定三条主任务各 3 条样本：

- `radishflow/suggest_flowsheet_edits`
- `radishflow/suggest_ghost_completion`
- `radish/answer_docs_question`

该入口可输出 JSONL 训练样本，并用 summary fixture 固定样本数、任务分布、来源 eval 样本、生成 sample id、训练字段、质量门禁计数和确认边界统计。

当前训练输出布局固定为：

- `training/` 只提交训练集合治理说明、manifest、summary、抽样复核记录和实验说明
- 默认 JSONL 输出到 `tmp/`，作为本地生成产物，不直接入仓
- `datasets/` 继续承载 eval 样本、candidate record 和示例对象，不作为生成后训练 JSONL 的默认落点
- 若后续确需提交小型 JSONL fixture，必须先在 `training/` 中说明样本数、用途、来源、复核状态和退场条件
- 大规模 JSONL、权重、checkpoint、adapter 二进制和 provider 临时 dump 不进入本仓库 committed 资产

当前更大训练集合的首个治理 manifest 已固定为 `training/datasets/copilot-training-dataset-governance-v0.json`，状态为 `draft`。它不替代转换 manifest，也不包含生成后的 JSONL；职责是约束哪些样本未来能进入更大的训练 / 蒸馏集合：

- 当前 seed set 继续来自 committed eval `golden_response` 与 audit pass `teacher_capture` 两类转换 manifest
- candidate record 必须被转换 manifest 显式列入，并通过 batch manifest、audit report、record schema、response schema、`project/task/sample_id` 一致性、风险确认和 citation 检查
- 当前小型 seed set 默认全量复核；后续更大 candidate record 池按 `project / task / source / provider / model / risk_level / requires_confirmation / coverage_tags / batch_id` 分层抽样
- 默认每个分层至少复核 `20%` 且不少于 `5` 条；新项目/任务、新 provider/model、高风险动作、`requires_confirmation=true`、新 action/patch 结构、新 retrieval source、schema/contract 版本变化和历史失败模式必须全量复核
- 至少保留 `10%` 离线评测 holdout，且每个任务至少 `3` 条；holdout 不得进入同一训练 JSONL split
- 样本若出现 audit 失效、schema 失效、人工复核拒绝、确认边界弱化、来源不可追踪、无理由重复、holdout 泄漏或模式过期，应从训练集合退场

当前 planned 人工复核与 holdout 接线已补两份轻量资产：

- `training/datasets/copilot-training-review-record-v0.json` 只记录复核模板与三组 planned review batch：`golden_response` seed set、`teacher_capture` seed set 和 offline eval holdout 泄漏检查。在没有真实 reviewer、timestamp、逐维度结果和泄漏判断前，不得标记为 `reviewed_pass`
- `training/datasets/copilot-training-holdout-split-v0.json` 固定三条主任务各 3 条 planned holdout，且显式排除当前两份训练转换 manifest 中已经列入的样本，避免训练 / 评测泄漏。该 split 不生成 JSONL、不运行模型

当前 candidate record 转换入口只允许显式列入转换 manifest 的记录进入训练样本，且必须同时满足：

- 所属 batch manifest 与 audit report 都存在且互相指向一致
- 选中的 `sample_file` 在 audit report 中为 `pass`
- candidate record 通过 `datasets/eval/candidate-response-record.schema.json`
- candidate record 内嵌 `response` 通过 `contracts/copilot-response.schema.json`
- record 的 `project / task / sample_id` 必须与 committed eval 样本一致

首批 candidate record 转换同样固定三条主任务各 3 条样本，`distillation.source=teacher_capture`，`teacher.model` 与 `teacher.record_id` 来自真实 record。该入口仍不重新调用 provider、不下载模型、不启动训练；它只把已审计通过的真实候选响应转换成可复验的训练 / 蒸馏样本。

第一版结构如下：

```json
{
  "schema_version": 1,
  "kind": "copilot_training_sample",
  "sample_id": "radishflow-suggest-flowsheet-edits-training-basic-001",
  "training_mode": "distillation",
  "project": "radishflow",
  "task": "suggest_flowsheet_edits",
  "input_request": {},
  "target_response": {},
  "distillation": {
    "source": "golden_response",
    "teacher": {
      "provider": "fixture",
      "model": "golden-response",
      "record_id": "radishflow-suggest-flowsheet-edits-training-basic-001"
    },
    "train_fields": [
      "summary",
      "answers",
      "issues",
      "proposed_actions",
      "citations",
      "risk_level",
      "requires_confirmation"
    ]
  },
  "quality_gates": {
    "schema_validated": true,
    "risk_reviewed": true,
    "citation_checked": true,
    "human_review_required": false
  },
  "metadata": {
    "source_eval_sample": "datasets/eval/radishflow/suggest-flowsheet-edits-reconnect-outlet-001.json",
    "created_for": "teacher-student-distillation",
    "notes": []
  }
}
```

字段边界：

- `input_request` 必须继续通过 `contracts/copilot-request.schema.json` 校验
- `target_response` 必须继续通过 `contracts/copilot-response.schema.json` 校验
- `project / task` 必须在 wrapper、`input_request` 与 `target_response` 三处一致
- 训练样本必须保持 advisory safety：`input_request.safety.mode=advisory`
- 当目标响应包含需要确认的输出、`high` 风险动作，或非 `manual_only` 的中风险执行边界动作时，`input_request.safety.requires_confirmation_for_actions` 必须保持 `true`
- `quality_gates` 表示该样本进入训练 / 蒸馏集合前已经通过 schema、风险与 citation 检查，不表示真实业务动作可自动执行
- `teacher_capture` 样本必须来自 audit pass 的 committed candidate record；audit failed、warning-only 未复核、schema-invalid 或 citation/risk 边界不稳定的 record 不得进入训练集合
- `high` 风险 `proposed_actions` 仍必须保留 `requires_confirmation=true`，不得因进入训练样本而弱化人工确认边界
- `suggest_ghost_completion` 的 `manual_only` 中风险候选属于可见候选排序边界，不等同于直接写回动作；只要响应本身不要求确认，且没有默认 `Tab` 或高风险动作，训练样本应保留原始 eval request 的确认口径

## `RadishMind-Image Adapter` 第一版仓库级契约

图片生成能力通过独立 adapter / backend 提供。`RadishMind-Core` 只负责生成结构化意图、约束、风险确认和审查信息，不直接生成图片像素。

第一版 image generation intent 已落成仓库级可回归契约：

- Schema：`contracts/image-generation-intent.schema.json`
- 最小 fixture：`scripts/checks/fixtures/image-generation-intent-basic.json`
- Smoke：`scripts/check-image-generation-intent-contract.py`

当前 schema 固定的是 `RadishMind-Core -> RadishMind-Image Adapter` 的结构化意图边界，不承诺具体 backend 常驻、权重下载、图片质量或像素生成实现。第一版结构如下：

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

## 统一输入抽象

建议统一请求对象命名为 `CopilotRequest`。

```json
{
  "schema_version": 1,
  "request_id": "optional-string",
  "project": "radishflow",
  "task": "explain_diagnostics",
  "locale": "zh-CN",
  "conversation_id": "optional-string",
  "artifacts": [],
  "context": {},
  "tool_hints": {
    "allow_retrieval": true,
    "allow_tool_calls": true,
    "allow_image_reasoning": false
  },
  "safety": {
    "mode": "advisory",
    "requires_confirmation_for_actions": true
  }
}
```

### 顶层字段建议

- `schema_version`：协议版本
- `request_id`：请求跟踪标识
- `project`：调用方项目标识，例如 `radishflow`、`radish`
- `task`：任务类型
- `locale`：本地化口径，例如 `zh-CN`
- `conversation_id`：会话关联标识
- `artifacts`：多模态输入集合
- `context`：结构化上下文
- `tool_hints`：检索、工具调用和图像推理偏好
- `safety`：安全模式与确认要求

## Artifact 抽象

建议采用统一数组，每个元素自描述：

```json
{
  "kind": "json",
  "role": "primary",
  "name": "flowsheet_document",
  "mime_type": "application/json",
  "uri": "optional-file-or-object-url",
  "content": "optional-inline-content",
  "metadata": {}
}
```

说明：

- `uri` 与 `content` 至少提供一个
- `role` 用于区分 `primary` / `supporting` / `reference`
- 小型文本或 JSON 可以内联，大型对象优先用引用

当前推荐支持：

- `json`
- `markdown`
- `text`
- `image`
- `attachment_ref`

## 项目专属上下文块

### `RadishFlow` 上下文建议

以下字段命名已尽量与已审查的 `rf-ui` / `radishflow-studio` 口径对齐：

```json
{
  "document_revision": 12,
  "selected_unit_ids": ["unit-1"],
  "selected_stream_ids": ["stream-1"],
  "diagnostic_summary": {},
  "diagnostics": [],
  "solve_session": {},
  "latest_snapshot": {},
  "control_plane_state": {
    "entitlement_status": "active",
    "last_error": null
  }
}
```

当前建议优先支持：

- `flowsheet_document`
- `document_revision`
- `selected_unit_ids`
- `selected_stream_ids`
- `diagnostic_summary`
- `diagnostics`
- `solve_session`
- `latest_snapshot`
- `control_plane_state`
- `selected_unit`
- `unconnected_ports`
- `missing_canonical_ports`
- `nearby_nodes`
- `cursor_context`
- `legal_candidate_completions`

说明：

- `canvas_snapshot` 适合通过 `artifacts` 追加，而不是替代结构化状态
- 模型输出不直接改写文档，应先生成提案，再交由业务命令层确认执行
- 当前仓库也已新增 `radishflow-adapter-snapshot.schema.json` 与 `build-radishflow-request.py`，先为 `explain_diagnostics`、`suggest_flowsheet_edits` 与 `explain_control_plane_state` 提供最小 `adapter-radishflow` 上游快照 -> `CopilotRequest` 装配链
- 当前仓库还已再前推一层：新增 `radishflow-export-snapshot.schema.json` 与 `build-radishflow-adapter-snapshot.py`，先把更贴近 `document_state / selection_state / diagnostics_export / solve_snapshot / control_plane_snapshot` 的导出对象稳定转换为 adapter snapshot，再继续装配成 `CopilotRequest`
- 当前仓库也已新增 `build-radishflow-export-request.py`，用于把 export snapshot 直接装配为 `CopilotRequest`，并由 `check-repo` 校验其结果与既有 eval sample `input_request` 一致
- 当前仓库也已新增 `validate-radishflow-export-snapshot.py`，用于在真实接线前先对 export snapshot 做 schema、任务级语义与敏感信息 smoke 校验，并由 `check-repo` 统一回归
- 当前这条装配链已不只覆盖最小 happy path，还补进了 `multi-object diagnostics`、`control-plane conflicting signals`、“multi-selection 但只允许单 actionable target”、`multi-unit + multi-stream + 单 primary focus`、selection 顺序保持、同风险多动作时的 input-first 排序理由，以及纯 `uri + metadata.summary` 与 mixed support summary 的 control-plane supporting capture 这几类代表性样本，避免 context packer 只在单对象、单诊断样本上自洽

### `RadishFlowExportSnapshot` 上游导出映射约定

`RadishFlowExportSnapshot` 当前定位为“上游导出对象和 adapter 之间的稳定边界”。

它的职责不是直接替代 `CopilotRequest`，而是先把上游更贴近真实对象结构的状态冻结下来，再由 adapter 决定怎样装配到统一协议。

当前建议按以下口径对齐：

- `document_state.document_revision`
  - 来源应是当前文档修订号或等价版本号
  - 这是所有后续选择集、诊断和 recent state 的统一时间基线
- `document_state.flowsheet_document`
  - 来源应是当前可供解释或建议使用的结构化 `FlowsheetDocument`
  - 若对象过大、已有稳定对象存储，也可改为只给 `flowsheet_document_uri`
- `selection_state.selected_unit_ids`
  - 来源应是 UI 或编辑器当前选择集中的 unit id 列表
  - 对多选场景必须保留完整选择集，不能为了下游简化而提前裁掉
- `selection_state.selected_stream_ids`
  - 来源应是 UI 或编辑器当前选择集中的 stream id 列表
  - 与 `selected_unit_ids` 可同时存在，用于表达“单元 + 流股”联合选择态
- `selection_state.primary_selected_unit`
  - 只有当上游确实存在“主焦点 unit”语义时才提供
  - 当前 adapter 只会在这个字段显式存在时才写入 `selected_unit`，不得再从 `selected_unit_ids + flowsheet_document` 反推
  - 它当前允许与 `selected_stream_ids` 等联合选择态并存；若同时给了 `selected_unit_ids`，则应显式落在该列表中，但不要求与 selection chronology 的第一项重合
  - 当前 committed fixture 也已覆盖“`selected_unit_ids` 中存在多个 unit，且同时保留多条 stream selection，但只落一个 `primary_selected_unit`”的组合，避免这条口径只停留在文档约定
  - 当前 committed fixture 还已固定 `selected_unit_ids` 与 `selected_stream_ids` 的输入顺序不会在 export -> adapter -> request 链中被偷偷重排，避免 UI 侧 selection chronology 在装配阶段丢失
  - 当前 committed fixture 还已进一步覆盖“更复杂 selection chronology + 单 actionable target”的组合：主焦点 `selected_unit` 可独立于 chronology 第一项存在，但 candidate_edit 仍只能落到证据充分的单一局部对象
- `diagnostics_export.diagnostic_summary`
  - 来源应是当前诊断摘要，例如 error / warning 计数
  - 若上游没有摘要，可省略，由 `diagnostics` 独立成立
- `diagnostics_export.diagnostics`
  - 来源应是当前任务可见的诊断对象数组
  - 多对象诊断、`stream_pair` 这类 target 语义应原样保留，不在导出层擅自改写成单对象解释
- `solve_session_state`
  - 来源应是当前求解会话摘要，例如 `status`、iteration limit、blocked state
  - 该块表达“会话状态”，不负责代替诊断或控制面定因
- `solve_snapshot`
  - 来源应是最近一次求解快照或 snapshot 摘要
  - 它适合提供 residual、solver_status、last_updated 等“近实时状态”
- `control_plane_snapshot`
  - 来源应是 entitlement / lease / sync / manifest 等控制面摘要
  - 若上游状态彼此冲突，应原样保留冲突，不得在导出层提前归一成单一根因
- `support_artifacts`
  - 来源应是额外但必要的 supporting 证据，例如 UI note、lease summary、操作面文本提示
  - 只允许补充解释证据，不得夹带 token、credential、cookie 或其他敏感原文
  - 若 artifact 指向外部 capture 或对象引用，当前优先采用 `uri + metadata.summary` 这类最小脱敏摘要，而不是把原始控制面载荷整段内联；若上游更适合输出 `metadata.redaction_summary` 或 `metadata.sanitized_summary`，当前也视为同类摘要键
  - 当前下游消费口径也已先收口：若 artifact 带可用 `content`，citation/excerpt 默认仍取 `content`；仅在没有可用 `content` 时，才按 `metadata.summary -> metadata.sanitized_summary -> metadata.redaction_summary` 的顺序回退到脱敏摘要键
  - 当前 runtime 已进一步冻结 artifact-backed citation canonicalization：只要 citation 能映射回 request 中真实 artifact，最终 `label / locator / excerpt` 就必须以该 artifact 的 canonical 字段为准，而不是保留模型自己生成的 metadata 派生文案
  - 这意味着 `metadata.redactions`、`metadata.source_scope`、`metadata.summary_variant` 这类审计/辅助元数据可以随 artifact 一起进入请求契约，但它们不能被当作最终 citation excerpt 或 locator 的展示源
  - 当前 committed fixture 也已覆盖“纯 `uri + metadata.summary`、不再额外内联 `content`”的 control-plane capture 摘要形态，避免 export 侧默认回退到长文本或原始 JSON 载荷
  - 当前 committed fixture 也已覆盖 `attachment_ref + json summary + text note` 的 mixed summary 组合，以及 `sanitized_summary + summary/redaction_summary + 最小 json rollup + text note` 的更复杂 mixed summary 变体，避免一旦需要多种 supporting 视角或不同脱敏摘要键就退回到原始 telemetry 或长文本拼接
  - 当前 eval regression 也已把这条 canonicalization 前推成正式约束：`explain_control_plane_state` 样本会逐字段校验 artifact citation 的 `label / locator / excerpt` 是否与 request artifact 保持一致，并额外阻止 excerpt 泄漏 `redactions`、`source_scope`、`summary_variant` 这类 metadata 字段名

当前任务级最小导出要求建议固定为：

- `task=explain_diagnostics`
  - 至少提供 `document_state.document_revision`
  - 至少提供 `flowsheet_document` 或 `flowsheet_document_uri`
  - 至少提供一类 selection 信息和一类 diagnostics 信息
  - 可选补充 `solve_session_state`、`solve_snapshot`
- `task=suggest_flowsheet_edits`
  - 与 `explain_diagnostics` 共享相同最小导出口径
  - 即使最终只有单个对象可落 patch，也必须保留完整 selection，不能在导出层提前裁掉其他已选对象
- `task=explain_control_plane_state`
  - 至少提供 `document_state.document_revision`
  - 至少提供 `control_plane_snapshot`
  - 可选补充 `solve_session_state` 与 `support_artifacts`

当前导出层的非目标也应明确：

- 不在 export 层推导 `selected_unit`
- 不在 export 层把多对象诊断改写成单对象结论
- 不在 export 层把冲突控制面信号提前收口成单一根因
- 不在 export 层决定哪个 selection object 最终可落 `candidate_edit`
- 不在 export 层透传敏感控制面凭据或超出当前任务所需的大体量原始状态

### 上游实现清单

若上游项目准备首次接线，当前建议按以下清单实现 exporter：

1. 先按任务选择最小模板
   - 可直接用 `python3 scripts/init-radishflow-export-snapshot.py --task <task>`
   - 该脚本会生成一份 schema-valid 的最小 export snapshot 起步模板
2. 先跑预接线 smoke 校验
   - 可直接用 `python3 scripts/validate-radishflow-export-snapshot.py --input <snapshot.json>`
   - 该脚本除 schema 外，还会检查任务级必需状态块、selection 语义，以及明显敏感字段/疑似凭据透传
   - 若上游已经能一次导出多条 snapshot，当前也可直接用 `python3 scripts/run-radishflow-export-smoke.py --manifest <manifest.json> --summary-output <summary.json>` 批量跑 `validate -> adapter -> request` smoke；manifest 里既可只填 `export_snapshot` 做最小装配烟测，也可附带 `expected_adapter_snapshot` / `expected_request_sample` 做更严格对照
3. 再替换 request 级基础字段
   - 必填：`request_id`、`locale`
   - 推荐：`request_id` 使用可追踪、可回放的业务侧请求标识
4. 再补 `document_state`
   - 必填：`document_revision`
   - `explain_diagnostics` / `suggest_flowsheet_edits` 必填：`flowsheet_document` 或 `flowsheet_document_uri`
5. 再补 `selection_state`
   - `explain_diagnostics` / `suggest_flowsheet_edits` 至少要有 `selected_unit_ids` 或 `selected_stream_ids`
   - 若 UI 存在主焦点 unit，才补 `primary_selected_unit`
   - 若主焦点 unit 与 UI chronology 第一项不是同一对象，也应保留原始 `selected_unit_ids` / `selected_stream_ids` 顺序，不要为了对齐 focus 重排 selection
6. 再补任务所需状态块
   - `explain_diagnostics` / `suggest_flowsheet_edits`：补 `diagnostics_export`
   - `explain_control_plane_state`：补 `control_plane_snapshot`
   - 有近实时求解状态时，再补 `solve_session_state` / `solve_snapshot`
7. 最后再补 supporting 证据
   - 仅在主状态不足以解释当前 UI 表现时补 `support_artifacts`
   - 典型例子：UI note、lease summary、同步提示文本

当前任务级 checklist 建议固定为：

- `explain_diagnostics`
  - 必填：`request_id`、`locale`、`document_state.document_revision`
  - 必填：`flowsheet_document` 或 `flowsheet_document_uri`
  - 必填：`selected_unit_ids` 或 `selected_stream_ids`
  - 必填：`diagnostic_summary` 或 `diagnostics`
  - 可选：`solve_session_state`
  - 可选：`solve_snapshot`
- `suggest_flowsheet_edits`
  - 必填：`request_id`、`locale`、`document_state.document_revision`
  - 必填：`flowsheet_document` 或 `flowsheet_document_uri`
  - 必填：`selected_unit_ids` 或 `selected_stream_ids`
  - 必填：`diagnostic_summary` 或 `diagnostics`
  - 可选：`solve_session_state`
  - 可选：`solve_snapshot`
  - 额外约束：不得因为最终只想 patch 一个对象，就在 export 层删掉其他已选对象
- `explain_control_plane_state`
  - 必填：`request_id`、`locale`、`document_state.document_revision`
  - 必填：`control_plane_snapshot`
  - 可选：`solve_session_state`
  - 可选：`support_artifacts`

当前脱敏清单建议固定为：

- 不透传 access token、refresh token、cookie、license secret、credential 原文
- 不透传完整控制面报文或超出当前任务所需的诊断原始堆栈
- 对外链对象优先传 `uri + 最小摘要`，而不是大对象全文内联
- 若 artifact 只用于解释 UI 冲突现象，优先给文本摘要，不直接给敏感原始载荷

当前 `validate-radishflow-export-snapshot.py` 已先固定以下预接线校验口径：

- `primary_selected_unit` 只有在显式提供时才允许出现，且其 `id` 必须显式出现在 `selected_unit_ids` 中；它可以和 `selected_stream_ids` 这类联合选择态并存，也不要求与 chronology 第一项重合
- `explain_diagnostics` / `suggest_flowsheet_edits` 必须同时满足“有 selection”与“有 diagnostics 信息”
- `explain_control_plane_state` 若仍带 `selection_state`，当前只记 warning，不直接判失败
- 对 `token`、`secret`、`cookie`、`authorization`、`api_key` 等敏感 key 名，以及 `Bearer ...`、`sk-...`、`AIza...` 这类疑似凭据内容，当前直接判失败
- `support_artifacts` 若只给 `uri` 而没有内联 `content`，当前会提示补 `metadata.summary` / `metadata.redaction_summary` / `metadata.sanitized_summary` 这类最小脱敏摘要；若内联文本过长，也会提示缩成更短摘要
- `support_artifacts` 若出现 `raw_payload`、`response_body`、`headers`、`stack_trace` 等疑似原始载荷字段，当前直接判失败
- 当前 `radishflow-export-snapshot.schema.json` 与 `radishflow-adapter-snapshot.schema.json` 也已把 `support_artifacts.metadata` 的常用字段显式收口到 `summary`、`sanitized_summary`、`redaction_summary`、`redactions`、`source_scope` 与 `summary_variant`，并在 schema 层先拦截 `headers`、`raw_payload`、`response_body`、`stack_trace` 等原始载荷键名
- 当前 `services/runtime/artifact_summary.py` 也是这条摘要键优先级的共享真相源：当 runtime 或 validator 需要判断“哪个 metadata 摘要键可用、提示文案该怎么写”时，统一按 `summary -> sanitized_summary -> redaction_summary` 的既定口径复用，而不再各自维护一套顺序
- 当前 `CopilotRequest.artifacts[*].metadata` 也已同步纳入同一条正式契约：除了 `RadishFlow` control-plane 常用的 `summary`、`sanitized_summary`、`redaction_summary`、`redactions`、`source_scope` 与 `summary_variant` 外，也显式保留 `Radish docs QA` 侧已稳定使用的 `source_type`、`page_slug`、`fragment_id`、`retrieval_rank` 与 `is_official`；同时 request 层也一样禁止 `headers`、`raw_payload`、`response_body`、`stack_trace` 等原始载荷键名进入 artifact metadata
- 当前 `services/runtime/artifact_metadata.py` 已成为这组三层契约字段和禁用键名的共享真相源，`check-artifact-metadata-contract.py` 会在 `check-repo` 中持续核对 `radishflow-export-snapshot`、`radishflow-adapter-snapshot` 与 `copilot-request` 三份 schema 是否仍与该真相源一致

- 对 `suggest_ghost_completion` 这类编辑器辅助任务，建议优先由本地规则层预生成 `legal_candidate_completions`，模型只在合法候选集中排序
- 当前仓库内的 `CopilotRequest` schema 已冻结 `selected_unit`、`unconnected_ports`、`missing_canonical_ports`、`nearby_nodes`、`cursor_context`、`legal_candidate_completions`、`naming_hints` 与 `topology_pattern_hints` 这些 ghost 补全上下文字段
- 对 `task=suggest_ghost_completion`，schema 当前还会强制要求 `document_revision`、单个 `selected_unit_ids`、`legal_candidate_completions`，以及至少一组 `unconnected_ports` 或 `missing_canonical_ports`
- `legal_candidate_completions` 当前建议尽量带结构化信号：
  - `ranking_signals`：例如距离、对齐度、模板命中率、与下一候选的分差
  - `naming_signals`：例如命名来源、前缀、编号后缀与重名检查结果
  - `conflict_flags`：例如“分差过小”“存在本地冲突标记”，用于阻止模型把候选误升级成默认 `Tab`
- 若本地规则层已将某候选标记为 `is_tab_default=true`，则它当前应同时满足 `is_high_confidence=true` 且不存在 `conflict_flags`
- 对连续搭建链场景，当前建议通过 `cursor_context.recent_actions` 透传最近一次或几次 `accept_ghost_completion` / `reject_ghost_completion` / `dismiss_ghost_completion` / `skip_ghost_completion` 记录，让模型明确知道“上一步刚接受了什么 ghost”或“刚否掉了什么 ghost”
- `recent_actions[*]` 当前最小字段为 `kind`、`candidate_ref` 和对应 kind 的修订号字段：`accepted_at_revision` / `rejected_at_revision` / `dismissed_at_revision` / `skipped_at_revision`；它们都应早于当前 `document_revision`
- `recent_actions` 当前只表达链式上下文，不得凌驾于 `legal_candidate_completions` 之上；若本地规则层给出的合法候选为空，当前仍应允许空建议
- 若同一 `candidate_ref` 刚被 `reject` / `dismiss` / `skip`，当前应把它视为 suppress-Tab 信号，而不是继续把它当作默认 `Tab` 候选强推
- 当前第一版 recent-actions 语义先收口为：`reject` / `dismiss` / `skip` 都共享“同一 candidate 的即时 suppress-Tab”语义；若该候选在本地规则层看来仍然合法，可继续保留为 `manual_only`，但不应立即恢复默认 `Tab`
- 上述 suppress 语义当前只作用于同一 `candidate_ref`：若最近被拒绝的是旧 candidate，而本地规则层已经切换到另一条新的高置信 candidate，则不应被旧反馈一并压成 `manual_only`
- 当前最小恢复窗口也已先固定一条共用时间基线：同一 `candidate_ref` 的 `reject` / `dismiss` / `skip` suppress 当前都只压制下一帧；若 `document_revision` 与对应 recent-action 修订号之间已隔一帧，且该候选仍被本地规则层标为高置信默认候选，则可恢复默认 `Tab`
- 当 `recent_actions` 同时包含多条 ghost 反馈时，当前应以“当前 `candidate_ref` 的最近一条相关动作”作为 suppress / cooldown 的直接依据：更早的同 candidate 反馈只能作为背景，不应覆盖更新动作；而其他 `candidate_ref` 的更新反馈也不应外溢误伤当前候选
- 这条“最近一条相关动作优先”约束同样适用于 cooldown 恢复态：若同一 `candidate_ref` 的最新否定动作 suppress 窗口已过，则更早的同 candidate 反馈不应继续阻止默认 `Tab` 恢复
- 当前 recent-actions 基线还已补齐两类 stacked 对称组合：同一 candidate 若先 `skip`、后又被最新一帧 `reject`，则应继续以最新同 candidate `reject` 维持 `manual_only`；而若 same-candidate `skip` cooldown 已过、最新 `reject` 针对的是其他 candidate，则该 other-candidate 反馈不得外溢压制当前默认 `Tab`
- `datasets/eval/radishflow-task-sample.schema.json` 当前也已支持外部 `candidate_response_record` 回灌，可将真实或模拟的 ghost completion capture 重新接回同一条 audit / regression 口径，而不必只停留在样本内联 `candidate_response`
- 当前仓库已提供 `scripts/run-radishflow-ghost-real-batch.py` 作为轻量批次入口，先以 3 个代表样本覆盖 `Tab / manual_only / empty` 三条用户侧主路径，完成 `capture -> manifest -> audit` 的最小 PoC 闭环
- 当前仓库也已提供 `scripts/run-radishflow-suggest-edits-poc-batch.py` 作为 `suggest_flowsheet_edits` 的最小批次入口：在 `mock` 模式下可直接基于 curated eval sample 的 `golden_response` 生成 `candidate_response_record -> manifest -> audit` 正式治理资产，先把高风险重连、局部规格占位与三步优先级链 3 条主路径补成 committed PoC；后续若切到真实 provider，则沿同一脚本入口继续做真实 capture
- 该入口当前若未显式传 `--output-root`，默认会直接落到 `datasets/eval/candidate-records/radishflow/batches/YYYY-MM/<batch_key>/`，使后续真实 batch 可以按正式短路径目录直接产出
- 该入口当前也已补逐样本单进程 capture 与 openai-compatible provider 单样本硬超时，避免单条真实 provider 请求失控时把整批 capture 一并拖住；同时本地 provider profile 已扩到 `anyrouter / sub_jlypx / qaq / google_gemini` 主链，并保留 `deepseek` 作为历史回放或临时 fallback；当前默认不再走 `openrouter`
- 对于已采集但仍停留在临时目录、且可能早于当前 canonicalization 修复的 ghost raw dump，当前推荐再通过 `scripts/import-candidate-response-dump-batch.py` 做一次“按当前 runtime 重新归一化后再导入正式批次”的收口，而不是直接把 `/tmp` 下的旧 `record` / `manifest` / `audit` 复制进仓库
- 当前已正式入仓八批 ghost real batch，且统一迁到 `datasets/eval/candidate-records/radishflow/batches/YYYY-MM/<batch_key>/` 短路径布局；八批都收口同一组 3 条 record，分别覆盖 `Tab`、`manual_only` 与 `empty` 三条正式导入主路径，且 `audit` 都已收口到 `3/3 pass`
- 第二批 `v3` 当前也已固定一个真实 provider 失败面：非空 ghost 输出可能返回“几乎完整但多闭合一个 `}`”的 JSON；runtime 现已先在 `radishflow / suggest_ghost_completion` 下做窄范围 repair，再继续沿用同一条 canonicalization 与导入链
- 第三批 `v4` 当前则确认：在延续同一组 3 样本 capture 时，尚未出现新的可重新归一化结构坏法；新增观察项主要是批处理执行中 `manual_only` 样本一度出现 provider 卡顿，但拆为单样本后仍可沿同一条 capture -> recanonicalize -> import -> audit 链完成正式入仓
- 第四批 `v5` 当前继续固定了另一条真实 provider 坏法：`manual_only` 多动作输出可能在第一条 action 后提前关掉 `proposed_actions` / `answers` 作用域，导致第二条 action 漂到错误层级；runtime 现已先在 `radishflow / suggest_ghost_completion` 下做窄范围 repair，再继续沿用同一条 canonicalization 与导入链
- 第五批 `v6` 当前则确认：当默认 `openrouter` 在同一时间窗内出现 provider-wide `HTTP 429` 时，`deepseek` fallback profile 已可继续完成同一组 3 样本真实 capture，且当前未再暴露新的结构坏法，仍可沿同一条 capture -> recanonicalize -> import -> audit 链完成正式入仓
- 第六批 `v7` 当前则继续固定两条新观察项：一是 openrouter 默认 free 模型已被上游废弃，会直接 `404 deprecated`；二是 `deepseek` 在 `empty` 样本上可能把 `summary` / `answer.text` 写成内嵌 JSON 字符串。当前前者已通过配置口径更新与 provider fallback 绕过，后者已通过 `radishflow / suggest_ghost_completion` 的文本摘要窄修复收口
- 第七批 `v8` 当前则继续固定一条新的执行判定：若在“先试 openrouter”的前提下先轮换其他 openrouter 候选模型，仍持续遇到路由不存在或短窗口限流，这类现象仍应优先归为 provider/model 可用性阻塞；当前这轮已按既定口径切回 `deepseek` fallback profile 完成正式 capture，并确认未新增新的任务级结构坏法
- 第八批 `v9` 当前则继续固定一条新的模型选择判定：即使同 provider 的备选模型已经可调用，也不代表它已达到可正式入仓的任务质量门槛；本轮 `openrouter` 的 `nemotron-nano` 就在 `manual_only` 样本上产出 project 名错误、空动作与错误 citation 结构的 schema-invalid JSON。当前这类输出不应通过 runtime fallback 强行导入，而应直接判定为模型质量漂移并切回备用 provider 继续 capture
- 当前仓库已用 `Feed -> Valve -> FlashDrum` 连续搭建链 example 固定这条口径：
  - [radishflow-ghost-candidate-set-chain-feed-valve-flash-flash-outlets-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-flash-outlets-001.json)
  - [radishflow-copilot-request-ghost-chain-feed-valve-flash-flash-outlets-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-flash-outlets-001.json)
  - [radishflow-copilot-request-ghost-chain-feed-valve-flash-flash-outlets-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-flash-outlets-001-debug-full.json)
  - [radishflow-ghost-candidate-set-chain-feed-valve-flash-stop-no-legal-outlet-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-stop-no-legal-outlet-001.json)
  - [radishflow-copilot-request-ghost-chain-feed-valve-flash-stop-no-legal-outlet-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-stop-no-legal-outlet-001.json)
  - [radishflow-copilot-request-ghost-chain-feed-valve-flash-stop-no-legal-outlet-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-stop-no-legal-outlet-001-debug-full.json)
  - [radishflow-ghost-candidate-set-chain-feed-valve-flash-outlets-name-conflict-no-tab-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlets-name-conflict-no-tab-001.json)
  - [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-name-conflict-no-tab-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-name-conflict-no-tab-001.json)
  - [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-name-conflict-no-tab-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-name-conflict-no-tab-001-debug-full.json)
  - [radishflow-ghost-candidate-set-chain-feed-valve-flash-outlets-ranking-ambiguous-no-tab-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlets-ranking-ambiguous-no-tab-001.json)
  - [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-ranking-ambiguous-no-tab-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-ranking-ambiguous-no-tab-001.json)
  - [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-ranking-ambiguous-no-tab-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-ranking-ambiguous-no-tab-001-debug-full.json)
- 当前仓库也已用 `Feed -> Valve -> FlashDrum` 的最近 reject example 固定“同一候选刚被 reject 后不可立即 retab”边界：
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-reject-no-retab-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-reject-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-reject-no-retab-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-reject-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-reject-no-retab-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-reject-no-retab-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-dismiss-no-retab-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-dismiss-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-dismiss-no-retab-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-dismiss-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-dismiss-no-retab-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-dismiss-no-retab-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-skip-no-retab-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-skip-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-skip-no-retab-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-skip-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-skip-no-retab-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-skip-no-retab-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-reject-cooldown-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-reject-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-reject-cooldown-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-reject-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-reject-cooldown-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-reject-cooldown-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-dismiss-cooldown-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-dismiss-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-dismiss-cooldown-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-dismiss-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-dismiss-cooldown-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-dismiss-cooldown-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-skip-cooldown-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-skip-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-skip-cooldown-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-skip-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-skip-cooldown-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-skip-cooldown-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-alternate-candidate-tab-after-other-reject-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-alternate-candidate-tab-after-other-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-reject-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-reject-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-reject-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-alternate-candidate-tab-after-other-dismiss-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-alternate-candidate-tab-after-other-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-dismiss-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-dismiss-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-dismiss-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-alternate-candidate-tab-after-other-skip-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-alternate-candidate-tab-after-other-skip-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-skip-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-skip-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-skip-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-skip-001-debug-full.json)
- 当前仓库也已用 `Feed -> Heater -> FlashDrum` 连续搭建链 example 验证这条口径可复用于第二模板：
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-heater-outlet-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-heater-outlet-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-heater-outlet-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-heater-outlet-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-heater-outlet-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-heater-outlet-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-alternate-candidate-tab-after-other-reject-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-alternate-candidate-tab-after-other-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-reject-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-reject-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-reject-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-alternate-candidate-tab-after-other-dismiss-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-alternate-candidate-tab-after-other-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-dismiss-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-dismiss-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-dismiss-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-alternate-candidate-tab-after-other-skip-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-alternate-candidate-tab-after-other-skip-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-skip-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-skip-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-skip-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-skip-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-reject-cooldown-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-reject-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-reject-cooldown-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-reject-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-reject-cooldown-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-reject-cooldown-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-dismiss-cooldown-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-dismiss-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-dismiss-cooldown-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-dismiss-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-dismiss-cooldown-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-dismiss-cooldown-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-skip-cooldown-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-skip-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-skip-cooldown-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-skip-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-skip-cooldown-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-skip-cooldown-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-reject-no-retab-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-reject-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-reject-no-retab-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-reject-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-reject-no-retab-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-reject-no-retab-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-dismiss-no-retab-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-dismiss-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-dismiss-no-retab-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-dismiss-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-dismiss-no-retab-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-dismiss-no-retab-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-skip-no-retab-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-skip-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-skip-no-retab-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-skip-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-skip-no-retab-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-skip-no-retab-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-name-conflict-no-tab-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-name-conflict-no-tab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-name-conflict-no-tab-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-name-conflict-no-tab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-name-conflict-no-tab-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-name-conflict-no-tab-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-stop-no-legal-outlet-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-stop-no-legal-outlet-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-stop-no-legal-outlet-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-stop-no-legal-outlet-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-stop-no-legal-outlet-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-stop-no-legal-outlet-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-ranking-ambiguous-no-tab-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-ranking-ambiguous-no-tab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-ranking-ambiguous-no-tab-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-ranking-ambiguous-no-tab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-ranking-ambiguous-no-tab-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-ranking-ambiguous-no-tab-001-debug-full.json)
- 当前仓库也已用 `Feed -> Cooler -> FlashDrum` 连续搭建链 example 验证这条口径可复用于第三模板：
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-cooler-outlet-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-cooler-outlet-001.json)
  - [radishflow-copilot-request-ghost-chain-feed-cooler-flash-cooler-outlet-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-cooler-outlet-001.json)
  - [radishflow-copilot-request-ghost-chain-feed-cooler-flash-cooler-outlet-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-cooler-outlet-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-alternate-candidate-tab-after-other-reject-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-alternate-candidate-tab-after-other-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-reject-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-reject-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-reject-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-alternate-candidate-tab-after-other-dismiss-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-alternate-candidate-tab-after-other-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-dismiss-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-dismiss-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-dismiss-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-alternate-candidate-tab-after-other-skip-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-alternate-candidate-tab-after-other-skip-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-skip-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-skip-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-skip-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-skip-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-reject-cooldown-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-reject-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-reject-cooldown-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-reject-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-reject-cooldown-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-reject-cooldown-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-dismiss-cooldown-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-dismiss-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-dismiss-cooldown-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-dismiss-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-dismiss-cooldown-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-dismiss-cooldown-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-skip-cooldown-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-skip-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-skip-cooldown-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-skip-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-skip-cooldown-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-skip-cooldown-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-latest-skip-cooldown-with-older-reject-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-latest-skip-cooldown-with-older-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-latest-skip-cooldown-with-older-reject-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-latest-skip-cooldown-with-older-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-latest-skip-cooldown-with-older-reject-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-latest-skip-cooldown-with-older-reject-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-reject-no-retab-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-reject-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-reject-no-retab-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-reject-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-reject-no-retab-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-reject-no-retab-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-dismiss-no-retab-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-dismiss-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-dismiss-no-retab-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-dismiss-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-dismiss-no-retab-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-dismiss-no-retab-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-skip-no-retab-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-skip-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-skip-no-retab-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-skip-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-skip-no-retab-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-skip-no-retab-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-stop-no-legal-outlet-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-stop-no-legal-outlet-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-stop-no-legal-outlet-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-stop-no-legal-outlet-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-stop-no-legal-outlet-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-stop-no-legal-outlet-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-ranking-ambiguous-no-tab-001.json](../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-ranking-ambiguous-no-tab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-ranking-ambiguous-no-tab-001.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-ranking-ambiguous-no-tab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-ranking-ambiguous-no-tab-001-debug-full.json](../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-ranking-ambiguous-no-tab-001-debug-full.json)

当前对 `suggest_ghost_completion` 的附加约束：

- 应只聚焦一个当前激活或刚放置的单元，不要在一次请求里同时补多个对象
- 建议数应收口在 1 到 3 条
- 若没有本地规则层给出的合法候选，允许返回空建议
- 用户接受 ghost 前，建议都只是 pending 状态，不应被视为正式文档修改
- 接受 ghost 后，仍必须继续经过本地命令系统和连接合法性校验

当前建议在模型请求之外，再单独冻结一份本地候选集交接格式：

- 参考 [contracts/radishflow-ghost-candidate-set.schema.json](../contracts/radishflow-ghost-candidate-set.schema.json)
- 该契约面向“本地规则层 / 适配层 -> 模型层”的 pre-model handoff
- 其职责是先把 `selected_unit`、未补全端口、邻近节点、命名提示与 `legal_candidate_completions` 收口成稳定对象，再由上层决定是否装配进 `CopilotRequest.context`

这样可以把两个边界拆清：

- 本地规则层负责生成合法候选与排序证据
- 模型层负责在合法候选空间中排序、解释和决定是否返回空建议

当前仓库还提供了一条最小装配入口：

- [build-radishflow-ghost-request.py](../scripts/build-radishflow-ghost-request.py)
- 它负责把 `radishflow-ghost-candidate-set.schema.json` 对象装配为 `CopilotRequest`
- 当前默认使用 `model-minimal` profile，而不是无差别透传
- 当前同时保留 `debug-full` profile，仅用于调试、对照和问题定位，不作为默认模型输入
- `selected_unit -> selected_unit_ids`、`unconnected_ports`、`cursor_context`、`naming_hints`、`topology_pattern_hints` 继续按最小稳定映射装配
- `legal_candidate_completions` 当前默认只保留模型最小必要字段：`candidate_ref`、`ghost_kind`、`target_port_key`、目标引用、建议名以及 `is_high_confidence` / `is_tab_default`
- `ranking_signals`、`naming_signals`、`conflict_flags` 这类本地排序证据默认留在 pre-model 候选集对象中，不直接透传到 `CopilotRequest`
- `nearby_nodes` 当前默认只保留 `type`、`id` 与 `direction`，不默认透传几何评分细节
- 当前仓库已同时固定 `model-minimal` 和 `debug-full` 两个示例输出，避免后续对“哪些本地证据可以进模型请求”再次退回口头约定
- 连续搭建链 example 当前也同步固定了两种 profile 的输出，避免 `recent_actions`、命名提示与 outlet 排序证据再次退回只在 `eval` 样本里口头存在
- 同一条连续搭建链当前还固定了“空候选停住”示例，确保 `recent_actions` 不会被误解为“只要有上一跳就必须继续给下一跳建议”
- 同一条连续搭建链当前也固定了“候选存在但命名冲突 no-tab”示例，确保 `recent_actions` 不会被误解为“只要候选非空就可以默认 Tab”
- 同一条连续搭建链当前也固定了“候选存在但排序分差过小 no-tab”示例，确保 `recent_actions` 不会被误解为“只要候选非空就一定存在默认 Tab”
- 同一条连续搭建链当前也固定了“同一候选刚被 reject / dismiss / skip 后 no-retab”示例，确保 `recent_actions` 不会被误解为“候选刚被用户否掉、关闭或跳过也可以下一帧继续默认 Tab 强推”
- 第二条与第三条链式模板当前也已补 `reject / dismiss / skip no-retab` 示例，确保 suppress-Tab 语义不会只在 `Feed -> Valve -> FlashDrum` 这一条模板上成立
- 三条链式模板当前都已补“other reject / dismiss / skip does not suppress new candidate”示例，确保 suppress 信号不会从旧 `candidate_ref` 外溢到新的高置信候选
- 三条链式模板当前都已补“same candidate retab after reject / dismiss / skip cooldown”示例，确保 suppress-Tab 语义不会被误读成永久 manual-only，而是只压制下一帧
- 三条链式模板当前也已补“same-candidate dismiss cooldown 已过、但 latest other-candidate skip 不外溢时仍可恢复 `Tab`”与“same-candidate 先 reject、后 skip 时以最新 skip 继续 no-retab”两组 stacked recent-actions 交错示例，确保恢复窗口与 latest-action precedence 不会只在单一动作对上成立
- 上述 `no-tab` 边界当前也已推进到 `datasets/eval/radishflow/` 的 response-level 回归样本，避免这条规则只停留在 pre-model examples
- 上述“链式停住空建议”边界当前也已推进到 `datasets/eval/radishflow/` 的 response-level 回归样本，避免这条规则只停留在 pre-model examples
- `Feed -> Valve -> FlashDrum` 的“排序分差不足导致 manual-only”边界当前也已推进到 `datasets/eval/radishflow/` 的 response-level 回归样本，避免第一模板的分叉态只停留在 pre-model examples
- `Feed -> Heater -> FlashDrum` 的正向 `Tab` 边界当前也已推进到 `datasets/eval/radishflow/` 的 response-level 回归样本，避免第二条模板只停留在 pre-model examples
- `Feed -> Heater -> FlashDrum` 的 `manual_only` 命名冲突边界当前也已推进到 `datasets/eval/radishflow/` 的 response-level 回归样本，避免第二条模板仍只剩顺风正例
- `Feed -> Heater -> FlashDrum` 的空候选停住边界当前也已推进到 `datasets/eval/radishflow/` 的 response-level 回归样本，避免第二条模板缺少与第一模板对齐的 empty baseline
- `Feed -> Heater -> FlashDrum` 的“排序分差不足导致 manual-only”边界当前也已推进到 `datasets/eval/radishflow/` 的 response-level 回归样本，避免第二模板的分叉态只停留在 pre-model examples
- `Feed -> Cooler -> FlashDrum` 的正向 `Tab`、`manual_only` 与空候选停住边界当前也已推进到 `datasets/eval/radishflow/` 的 response-level 回归样本，避免第三条模板只停留在 pre-model examples
- `Feed -> Cooler -> FlashDrum` 的“排序分差不足导致 manual-only”边界当前也已推进到 `datasets/eval/radishflow/` 的 response-level 回归样本，避免第三模板的分叉态只停留在 pre-model examples

当前对 `suggest_flowsheet_edits` 的附加约束：

- `proposed_actions` 当前应以 `candidate_edit` 为主，并保持“最小、可审查、非执行式”的局部 patch 口径
- 若同时存在多条 `candidate_edit`，顺序必须稳定，优先更阻塞、更直接或风险更高的提案
- 若同时存在多个 `issues`，顺序也必须稳定，优先已确认且直接对应 patch 的问题，再列未确认或保留性 warning
- 顶层 `citations` 的顺序应稳定保持为“direct diagnostics -> target artifacts -> supporting snapshot/context”
- `issue.citation_ids` 与 `candidate_edit.citation_ids` 的顺序也应稳定保持为“direct diagnostic -> target artifact -> supporting snapshot/context”
- 单个 `candidate_edit.patch` 的键顺序当前也属于契约的一部分，不应把主修改块和保护性元字段随机重排
- `patch.parameter_updates`、`patch.parameter_updates.<parameter_key>`、`patch.parameter_updates.<parameter_key>.<detail_key>`、`patch.spec_placeholders`、`patch.parameter_placeholders` 与 `patch.connection_placeholder` 的多层键/数组顺序都应保持稳定，避免同一建议在不同回归轮次里只因为序列漂移而变成“看起来像不同答案”
- 上述稳定性约束当前已进入 `datasets/eval/radishflow-task-sample.schema.json` 与 `scripts/run-eval-regression.py` 的正式评测口径，而不是只停留在任务卡示例

### `Radish` 上下文建议

`Radish` 当前更适合以知识和内容为中心打包上下文：

```json
{
  "current_app": "document",
  "route": "/document/guide/authentication",
  "resource": {
    "kind": "wiki_document",
    "slug": "guide/authentication",
    "title": "鉴权与授权指南"
  },
  "viewer": {
    "is_authenticated": true,
    "roles": ["Admin"],
    "permissions": ["console.hangfire.view"]
  },
  "attachment_refs": ["attachment://123456789"]
}
```

当前建议优先支持：

- `current_app`
- `route`
- `resource`
- `viewer`
- `attachment_refs`
- `search_scope`

说明：

- 这里的 `resource` 可以指向固定文档、在线文档、论坛帖子、评论或 Console 页面
- `viewer` 只提供最小必要身份摘要，不透传 token、cookie 或原始安全凭据

当前对 `answer_docs_question` 的附加约束：

- `search_scope` 应优先收口在 `docs`、`wiki` 及与当前 `resource` 直接相关的受控来源
- 已召回内容应优先以片段级 `artifacts` 传入，而不是透传整篇长文或整段论坛线程
- 若补充论坛或 FAQ 内容，应能让响应层区分其与正式文档的来源差异
- 当前页面已经足够回答问题时，不应继续扩张召回范围
- 当前阶段建议把 `search_scope` 收口到 `docs`、`wiki`、`faq`、`forum`、`attachments` 这组受控枚举中，不把 `console` 内部状态本身作为本任务的检索 scope
- 当前最小回归样本已要求召回 artifact 带 `metadata.source_type`、`metadata.page_slug`、`metadata.fragment_id`、`metadata.retrieval_rank`、`metadata.is_official`
- `primary` artifact 应与当前 `resource.slug` 对齐；`faq` / `forum` 仅作为 supporting 或 reference 补充来源，不应抢占主证据位置
- 若接入外部 `candidate_response_record`，当前最小回归还会校验 `sample_id`、`request_id`、`input_record` 摘要与样本请求对齐
- 若外部记录来自真实候选输出快照，当前建议补 `capture_metadata.capture_origin`、`capture_metadata.collection_batch` 与 `capture_metadata.tags`
- 负例回放继续复用与正例相同的响应校验逻辑；当前允许将真实 captured record 跨样本回放，以验证 record 对齐与响应规则会共同拒绝错配输出

## 统一输出抽象

建议统一响应对象命名为 `CopilotResponse`。

```json
{
  "schema_version": 1,
  "status": "ok",
  "project": "radishflow",
  "task": "explain_diagnostics",
  "summary": "string",
  "answers": [],
  "issues": [],
  "proposed_actions": [],
  "citations": [],
  "confidence": 0.0,
  "risk_level": "medium",
  "requires_confirmation": true
}
```

### 顶层字段建议

- `status`：`ok` / `partial` / `failed`
- `summary`：面向用户的简短说明
- `answers`：适合直接展示的解释或回答片段
- `issues`：发现的问题列表
- `proposed_actions`：候选动作
- `citations`：证据、来源或输入引用
- `confidence`：整体置信度
- `risk_level`：`low` / `medium` / `high`
- `requires_confirmation`：是否必须人工确认

## Candidate Action 抽象

当前建议把高风险输出统一压成候选动作：

```json
{
  "kind": "candidate_edit",
  "title": "补充流股规格",
  "target": {
    "type": "stream",
    "id": "stream-1"
  },
  "rationale": "当前诊断显示该流股缺失关键规格。",
  "patch": {},
  "risk_level": "medium",
  "requires_confirmation": true
}
```

说明：

- `patch` 是候选提案，不是直接执行的命令
- `kind` 可以按项目扩展，但都必须经过项目适配层确认

对于 `RadishFlow suggest_ghost_completion`，当前建议新增 `ghost_completion` 这类候选动作：

```json
{
  "kind": "ghost_completion",
  "title": "补全 FlashDrum 的 vapor outlet ghost 连线",
  "target": {
    "type": "unit_port",
    "unit_id": "U-12",
    "port_key": "vapor_outlet"
  },
  "rationale": "当前 canonical port 尚未连接，且本地规则已提供合法 ghost 候选。",
  "patch": {
    "ghost_kind": "ghost_connection",
    "candidate_ref": "cand-vapor-stub",
    "ghost_stream_name": "V-12"
  },
  "preview": {
    "ghost_color": "gray",
    "accept_key": "Tab",
    "render_priority": 1
  },
  "apply": {
    "command_kind": "accept_ghost_completion",
    "payload": {
      "candidate_ref": "cand-vapor-stub"
    }
  },
  "risk_level": "low",
  "requires_confirmation": false
}
```

## 当前推荐任务枚举

### `RadishFlow`

- `explain_diagnostics`
- `suggest_flowsheet_edits`
- `suggest_ghost_completion`
- `summarize_selection`
- `explain_control_plane_state`
- `inspect_canvas_snapshot`

### `Radish`

- `answer_docs_question`
- `summarize_doc_or_thread`
- `suggest_forum_metadata`
- `explain_console_capability`
- `interpret_attachment`

## 禁止透传与脱敏要求

以下内容当前不应进入 `CopilotRequest`：

- `access_token`
- `refresh_token`
- cookie
- `credential_handle`
- 本机安全存储引用
- 未裁剪的 auth cache 索引原文
- 本地绝对密钥路径和证书密码

补充约定：

- `RadishFlow` 控制面相关只传状态摘要、manifest 摘要和错误信息
- `Radish` 权限相关只传角色 / 权限键摘要和当前页面语义，不传安全凭据

## 关键边界

- `RadishMind` 只返回建议、解释或结构化候选动作
- 最终真相源仍由上层项目维护
- 高风险建议必须要求人工确认
- 若模型能力不足，应允许退化到检索、规则或模板响应
- 协议的统一对象是“外部智能层”，不是两个项目的内部 DTO

## 当前推荐原则

- 先统一协议，再分别做项目适配
- 先支持 `RadishFlow`，再让 `Radish` 逐步接入
- 先做可消费的 JSON，再谈复杂自治代理
- 不把 `RadishFlow` 和 `Radish` 强行拉成同一业务字段集合

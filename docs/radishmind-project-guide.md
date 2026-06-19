# RadishMind 项目总览与使用指南

更新时间：2026-06-15

## 这份文档讲什么

这是一份面向新读者的项目说明书，回答四个问题：

- `RadishMind` 现在被定义成什么
- 仓库怎么分工
- 今天可以怎样实际运行它
- 如果你要继续推进开发，先看哪几条主线

它不替代 `docs/radishmind-current-focus.md`、`docs/devlogs/` 或任务卡，也不记录阶段推进流水。

2026-06-14 起，具体功能或长期开发目标先看 [功能设计文档入口](features/README.md)。任务卡只承载实现批次、前置条件或高风险边界，不再作为功能默认主文档。

## 项目定位

`RadishMind` 现在的正式定位是：

`Radish` 体系下的 AI 工具、工作流、模型网关和 Copilot 集成平台。

它不是上层业务真相源，不替代 `RadishFlow`、`Radish` 或 `RadishCatalyst` 的业务决策权，也不是“只放模型实验”的仓库。

当前主要职责是：

- 提供用户端 AI 应用 / 工作流 / API key / 调用记录工作台。
- 提供管理端 provider/profile / 模型路由 / 租户 / quota / secret / 审计控制面。
- 提供模型 API 分发和兼容网关。
- 接收结构化上下文
- 运行最小推理链路
- 兼容多种上游模型接入方式与多种下游协议接口
- 提供本地只读产品发现面，供 console 或上层 UI 读取平台可展示能力和停止线
- 组织局部工具、规则和响应收口
- 输出解释、诊断、结构化建议和候选动作
- 维护统一协议、评测门禁、审计记录和训练治理

## 实现分工

- `UI`：`React + Vite + TypeScript`
- `平台服务层`：`Go`，覆盖 `HTTP API`、`gateway`、control plane、鉴权、流式转发、长驻进程、观测和部署壳
- `模型侧`：`Python`，覆盖训练、评测、`prompt / builder`、离线推理和校验逻辑
- `contracts/`：唯一 canonical protocol 真相源，所有语言只能消费，不得各自重新定义业务协议
- `Radish` 集成：部署、数据库、登录 / 授权默认对齐 `Radish`；未来作为 OIDC client 接入 `Radish`；不默认引入 `.NET`

## 当前四个产品面

1. `User Workspace`：AI 应用、Workflow、Agent / Copilot、RAG、API key、调用量、成本和运行记录。
2. `Admin Control Plane`：租户、用户、角色、权限、provider/profile、模型路由、quota、secret、审计和部署状态。
3. `Model Gateway / API Distribution`：OpenAI-compatible / Responses / Messages / Models API，多 provider / profile / model 分发。
4. `Workflow / Agent Runtime`：Prompt、LLM、HTTP tool、RAG retrieval、condition、output 和受控 agent loop。

这些产品面及 Image Path 的后续开发入口已整理到 `docs/features/`。

## 当前五条主线

1. `Runtime Service`：本地启动、gateway、route、provider/profile、协议兼容、响应封装、部署基础；当前已达到 short close，request observability 和 error taxonomy 已进入平台门禁。
2. `Conversation & Session`：会话标识、历史压缩、恢复和审计边界；当前已有 session record、recovery checkpoint record/manifest/read result、northbound session metadata、metadata-only route smoke、confirmation / audit / result / executor / storage 设计门禁、short close readiness delta、stop-line manifest 和 close-candidate status summary。
3. `Tooling Framework`：检索、附件解析、候选生成、builder、tool policy 和 audit；当前已有 tool contract、registry、audit record、session binding、metadata-only result cache、result materialization policy、executor boundary、storage backend design、deny-by-default gates 和 executor/storage/confirmation enablement plan。
4. `Evaluation & Governance`：schema、smoke、offline eval、review、promotion gate、负向消费 summary、route smoke coverage summary、readiness summary、implementation preconditions、negative regression governance suite、negative coverage rollup、route negative coverage matrix 和 readiness consistency rollup。
5. `Model Adaptation`：基座选型、prompt/runtime 协同、蒸馏、训练样本治理和模型晋级。

如果你要继续推进开发，当前主线在 `Control Plane / User Workspace / Workflow v1`。已完成产品面边界、control plane 数据边界、Radish OIDC 前置条件、gateway API key / quota 前置条件、workflow definition / run record 边界，以及 read-side 的 read model、read-only route contract、response fixture、negative contract、implementation preconditions、fake-store-backed handler plan、七条 fake-store-backed handler implementation、auth/db preconditions、TypeScript consumer contract、formal UI boundary、formal UI implementation readiness、`apps/radishmind-web/` read-only shared shell、七个只读页面切片、formal UI readiness close、dev-only live read consumer、auth/store transition preconditions、repository contract/read store readiness、Go contract type、静态 contract smoke runner、repository interface readiness、adapter implementation readiness refresh、selector enablement preconditions、schema migration implementation preconditions、repository adapter implementation plan、schema artifact manifest readiness、store selector smoke readiness、production auth readiness、adapter smoke readiness、implementation trigger review、product sample consistency、workflow workspace context consistency 和 `Control Plane Durable Read Foundation v1`；说明入口见 [Control Plane Read-Side 契约](contracts/control-plane-read-side.md)。普通只读展示页不再默认逐页新增专项门禁；dev-only live read path 只验证 fake-store-backed handler 的 HTTP consumer shape；RadishFlow Copilot 与 Radish Docs Assistant 两组只读产品样例以 `control-plane-read-response-fixtures-v1` 的 success response 为 canonical source，再由 Go fake store、前端离线默认数据和 consumer smoke 一起校验一致性；repository/read store 当前已把七条 read handler 收束到 `ControlPlaneReadRepository` interface + fake-store bridge，但 repository adapter、selector、SQL、migration、真实数据库、Radish OIDC、token validation 和 production API consumer 仍未打开。`Workflow / Agent Runtime Function Surface v1` 已在 `apps/radishmind-web/` 增加 application detail、definition detail、run detail、blocked action preview、confirmation placeholder、Draft Designer、offline validation inspector、execution plan preview、runtime readiness inspector、surface overview、context selection、scenario inspector、review workspace、User Workspace Home 和 Review Handoff；这些都是 fixture-derived、blocked-capability-first 的产品面，共享 `workflowWorkspaceContext` 负责统一派生当前 application、definition、run、draft 和 scenario 的组合关系。`Saved Workflow Draft v1` 已实现 platform Go domain service、memory dev store、dev-only HTTP route 和 web consumer，提供 save / read / validate、版本冲突、失败语义、sanitized response、sample / local / saved / failed / version conflict 状态区分、Draft Designer 受控本地编辑、User Workspace 本地草案创建和 no sample fallback 测试；当前仍没有 durable persistence 或 production API，也不新增 executor、confirmation decision、writeback 或 replay。`Provider Runtime & Health v1` 已完成 `provider-capability-matrix-v1`、`provider-health-smoke-v1`、`provider-selection-policy-v1`、`provider-retry-fallback-policy-v1` 和 `provider-runtime-docs-refresh` 五个可检查切片并进入 close candidate，不继续默认新增 provider 同层小切片；`Production Ops Hardening v1` 的静态边界已经收口，没有明确运行窗口时降为等待项。P2 停止线继续作为背景证据保留，不代表真实 executor、durable store、confirmation 接线、materialized result reader、长期记忆、业务写回或 replay 已经完成。

`Model Gateway / API Distribution` 的当前产品 UI 也已进入离线证据组织层：Model Gateway Overview、Route Evidence、Usage/Audit Evidence 和 Evidence Review / Readiness 都位于 `apps/radishmind-web/`，复用 read shell、API key summary、quota summary、run history、audit log、provider runtime 与 gateway readiness 证据。它们只解释 northbound API surface、provider/profile、route binding、selection case、key scope、quota / cost snapshot、trace / failure、audit decision、readiness rollup、evidence checklist 和 locked capability，不新增真实网关 route、production gateway、secret resolver、API key lifecycle、quota enforcement、cost record write、retry/fallback execution、数据库、Radish auth 或 repository adapter。

`Admin Operations Review / Readiness` 是同一只读产品壳中的管理端汇总面：它复用 tenant overview、audit log、Model Gateway Evidence Review 和 Production Ops 静态证据，把 readiness、evidence checklist、operational risk 和 boundary lock 放在一个审查视图里。`Admin Provider/Profile & Deployment Evidence Review / Readiness` 继续复用 Model Gateway route / review、Admin Operations、tenant overview 和 audit log，把 provider/profile readiness、model route readiness、secret / deployment evidence、operator risks 和 locked capabilities 组织成管理端阅读路径。它们都不代表 production admin console，不提供 tenant mutation、provider/profile mutation、model route change、raw audit export、deployment preflight、secret resolver、workflow executor、writeback 或 replay。

`Image Path` 当前已有 metadata-only response builder 链路：`services/runtime/image_artifact_runtime_mapper.py` 只消费 `image_generation_artifact` metadata，`services/runtime/image_artifact_response_consumer.py` 只把 mapper 成功结果合并为现有 `CopilotResponse.citations` artifact citation，`services/runtime/inference_response.py#coerce_response_document` 只从 request artifact metadata 发现并接入该链路。它用于验证 `artifact://`、sha256、mime type、dimensions、safety review、provenance 和 fail-closed 语义，不读取图片二进制、不查 artifact store、不解析 public URL、不调用真实生图 backend、不上传 artifact、不改 response schema。

完整正式用户端、生产管理端、workflow builder、租户 / quota / billing、Radish OIDC client、repository adapter、read store repository implementation 和完整模型网关控制面仍未实现；当前本地 console 只是 ops surface 和只读产品壳，`apps/radishmind-web/` 默认是离线 read-side product UI shell，显式 dev-only live path 也不是 production API consumer。

## 目录速览

- `docs/`：正式文档源
- `docs/features/`：功能设计与开发文档
- `contracts/`：JSON Schema 真相源
- `scripts/`：检查、运行、转换、评测和最小运维入口
- `deploy/`：Docker local / test / prod 部署边界说明和 compose 资产
- `datasets/`：eval 样本、示例对象和 candidate record
- `training/`：训练治理、实验说明和复核记录
- `services/`：runtime 与 gateway 实现
- `adapters/`：上游项目适配层
- `tmp/`：本地临时产物，不提交

## 今天能怎么运行

### 1. 直接跑最小推理链路

当前最小运行入口是 `scripts/run-copilot-inference.py`。它不是长驻服务，而是单次 CLI runtime。

```bash
./scripts/run-python.sh scripts/run-copilot-inference.py \
  --sample datasets/eval/radishflow/suggest-flowsheet-edits-reconnect-outlet-001.json \
  --provider mock \
  --response-output tmp/rf-suggest-edit.response.json
```

如果后续要接真实 provider，再显式传 `--provider openai-compatible|huggingface|ollama`、`--provider-profile`、`--model`、`--base-url`、`--api-key`。当前这条入口已经能按 profile 分流到 `openai-compatible chat`、`gemini-native` 和 `anthropic-messages` 三类上游协议；`services/platform/` 也已把 `/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models`、`/v1/platform/overview`、`/v1/platform/local-smoke`、SSE bridge、provider/profile inventory、request-side selection、`HuggingFace` / `Ollama` coverage、本地启动 wrapper、JSON 配置层级、deployment smoke、diagnostics / failure boundary 和 discoverability 对齐纳入第一版 runtime foundation。

### 2. 跑进程内 gateway demo

如果你要看 `RadishFlow export -> adapter -> request -> gateway` 整条链路，当前正式入口是：

```bash
./scripts/run-python.sh scripts/run-radishflow-gateway-demo.py \
  --check-summary scripts/checks/fixtures/radishflow-gateway-demo-summary.json
```

这条链路当前使用 mock provider，只验证 route、schema、advisory-only 和 `requires_confirmation` 等不变量。

### 3. 跑 service/API smoke matrix

当前 `RadishFlow` 的正式服务门禁入口是：

```bash
./scripts/run-python.sh scripts/check-radishflow-service-smoke-matrix.py \
  --check-summary scripts/checks/fixtures/radishflow-service-smoke-matrix-summary.json
```

这条矩阵会一起核对：

- CLI runtime
- gateway API
- gateway demo
- UI consumption
- candidate handoff

它现在是仓库里最接近“服务切片验收”的正式门禁。

### 3.5 运行 Go 平台服务层

当前 `Go` 平台服务层已落在 `services/platform/`，用于承载 `HTTP API`、`gateway`、鉴权、流式转发、观测和部署壳。日常本地运行优先使用 wrapper：

```bash
./scripts/run-platform-service.sh config-check
./scripts/run-platform-service.sh diagnostics
./scripts/run-platform-service.sh serve
```

Windows / PowerShell 使用对应的 `pwsh ./scripts/run-platform-service.ps1 config-check|diagnostics|serve`。

当前固定 `/healthz`、`/v1/platform/overview`、`/v1/platform/local-smoke`、`/v1/models`、`/v1/models/{id}`、`/v1/chat/completions`、`/v1/responses`、`/v1/messages`、session/tooling metadata shell、metadata-only checkpoint read，以及七条 Control Plane Read-Side route。`/v1/platform/overview` 与 `/v1/platform/local-smoke` 只服务本地只读 console、Dev Diagnostics 和 readiness 摘要；checkpoint read 不是 durable store、materialized result reader 或 replay executor。

Control Plane Read-Side route 目前已注册到平台服务，但默认外部请求仍没有真实身份上下文，应得到 fail-closed envelope。只有显式设置 `RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1` 时，本地开发态才允许 `apps/radishmind-web/` 的 dev-only live consumer 通过 `X-RadishMind-Dev-Read-*` 测试身份 header 注入 test-only fake auth context。该路径只能读取 fake-store-backed handler，不能当成 production admin API、正式 user workspace API、真实 auth middleware 或数据库 read path。

read-side repository/read store 迁移当前已完成第一层受控实现：`control_plane_read_repository_contract.go` 定义 `ReadRepositoryContext`、七条 route request / result type、failure code、projection / filter / sort 和 route type matrix；`control_plane_read_repository_contract_smoke_runner.go` 消费该 type matrix，并对齐 repository contract smoke fixture；`control_plane_read_repository.go` 定义 `ControlPlaneReadRepository` interface 和 fake-store bridge，七条 read handlers 已通过该 interface 消费数据。后续 adapter/read-store 仍被 adapter readiness refresh、selector enablement preconditions、schema migration implementation preconditions、repository adapter implementation plan、schema artifact manifest readiness 和 store selector smoke readiness 卡住。它们用于证明未来 adapter / selector / migration 实现前的 contract shape、failure mapping、no fake fallback 和 no side effects 一致，不执行数据库查询，也不声明 adapter、store selector、SQL、manifest、migration、OIDC 或 production API consumer ready。

production auth readiness、adapter smoke readiness 和 implementation trigger review 位于 read-side 迁移尾部的静态说明与检查层：前两个分别说明未来 OIDC / token validation 和 durable adapter smoke 需要哪些证据，最后一个明确当前 schema artifact、store selector、production auth 与 adapter smoke 四类候选都没有实现触发条件。因此如果看到 `control_plane_read_auth_middleware.go`、`control_plane_read_repository_adapter.go`、`control_plane_read_store_selector.go`、read-side migration manifest 或 adapter smoke fixture 出现在源码树里，应先视为越过当前停止线，而不是已进入实现阶段；`control_plane_read_repository.go` 则属于已完成的 durable read foundation interface 边界。

这仍然不是 production deployment：它已经能作为本地平台服务切片运行和诊断，但尚未具备生产级 secret backend、进程监管、环境隔离和正式发布包。

### 3.6 运行本地 console 产品壳

本地 console 位于 `apps/radishmind-console/`，默认前端端口为 `4000`，后端平台端口为 `7000`。最省事的开发入口是从仓库根目录运行：

macOS / Linux / WSL 使用：

```bash
./scripts/run-radishmind-console-dev.sh
```

Windows / PowerShell 使用：

```powershell
pwsh ./scripts/run-radishmind-console-dev.ps1
```

该入口会启动或复用 platform 后端和 console 前端，并探测 `/healthz`、`/v1/platform/overview`、`/v1/platform/local-smoke`、本地 CORS preflight 和前端页面。它不是 production supervisor，不负责长期守护进程，也不实现 executor、durable store、confirmation、业务写回或 replay。

如果只想验证已有 platform 服务的本地 readiness，可运行：

```bash
./scripts/run-python.sh scripts/run-platform-local-smoke.py \
  --base-url http://127.0.0.1:7000 \
  --check
```

console 页面当前直接消费 `/v1/platform/overview` 与 `/v1/platform/local-smoke`，展示 Runtime overview、Service Status、Model Inventory、Provider/Profile Details、Session And Tooling、Blocked Action Detail、Dev Diagnostics、Local Readiness、Stop Lines 和 Audit Boundary。它仍是本地只读 ops surface，不是 production console、正式用户端或生产管理端。

### 3.7 运行正式 read-only product UI shell

正式产品 UI 的当前实现位于 `apps/radishmind-web/`。它默认是离线 read-side shell，只消费 `contracts/typescript/control-plane-read-api.ts`；当显式设置 `VITE_RADISHMIND_READ_SOURCE=dev-live-http` 且平台服务设置 `RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1` 时，可用 dev-only live read consumer 通过 HTTP 读取 fake-store-backed handler 和测试身份上下文。RadishFlow Copilot 与 Radish Docs Assistant 的只读产品样例来自同一组 response fixture，并由 `control-plane-read-product-sample-consistency-v1` 防止 Go fake store、前端默认 view model 和 consumer smoke 之间漂移。该 live path 不能解释为真实数据库、Radish OIDC、production API consumer、API key / quota、read store repository 或 workflow executor ready。

日常预览或前后端联调优先使用仓库根目录启动脚本，不再手动拼接环境变量：

```bash
./start.sh web-live
./start.sh web-offline
```

Windows / PowerShell 使用：

```powershell
pwsh ./start.ps1 -Command web-live
```

`web-live` 会启动或复用 `http://127.0.0.1:7000` platform 后端和 `http://127.0.0.1:4100` 产品 UI，并只启用 dev-only fake read auth。它不是 production supervisor，不接真实数据库、Radish OIDC、repository adapter、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

```bash
cd apps/radishmind-web
npm run build
```

如需本地查看页面，可在明确需要时运行：

```bash
cd apps/radishmind-web
npm run dev
```

当前页面包括：

- route catalog、shared states 和 forbidden output guard
- `admin-tenant-overview`
- `admin-audit-log`
- `workspace-applications`
- `workspace-api-keys`
- `workspace-usage-quota`
- `workspace-workflow-definitions`
- `workspace-run-history`
- workflow application detail
- workflow definition detail
- workflow run detail
- workflow blocked action preview
- workflow confirmation placeholder
- Workflow Draft Designer（默认离线，可显式使用 dev-only saved draft consumer）
- workflow offline draft validation inspector
- workflow execution plan preview
- workflow runtime readiness inspector
- Workflow Surface Overview
- workflow workspace context selection
- Workflow Scenario Inspector
- Workflow Review Workspace
- User Workspace Home
- Workflow Review Handoff
- Model Gateway Overview
- Model Gateway Route Evidence
- Model Gateway Usage/Audit Evidence
- Model Gateway Evidence Review / Readiness
- Admin Operations Review / Readiness
- Admin Provider/Profile & Deployment Evidence Review / Readiness

七个 read-side summary 页面展示 route metadata、request / audit ref、状态预览和脱敏 summary；默认使用离线 view model，dev-only live mode 也只能读取 fake-store-backed handler。workflow function surface 面板继续复用这些 summary 和离线 fixture，展示 application / definition / run / draft / validation / execution plan / runtime readiness 的只读详情、风险、审计引用和 blocked capability。当前读法是先看 User Workspace Home 的应用组合、当前 review、最近 run、优先 readiness、主要 route evidence 和 stop-line 摘要，再进入 Workflow Review Workspace 查看 selected application + definition + run + draft + scenario 的关系、scenario intent、draft validation、execution plan、runtime readiness、blocked capability rollup 和 stop-line rollup，最后用 Workflow Review Handoff 读取人工审查交接摘要。所有派生链路都经由 `workflowWorkspaceContext` 组合；这些选择只改变本地查看上下文，不接数据库、OIDC、repository implementation、API key lifecycle、quota enforcement、workflow executor、confirmation decision、draft / validation / execution plan / readiness / scenario / review / handoff persistence、writeback、run replay 或 run resume。

Model Gateway 四个页面的读法是先看 Overview 识别 northbound API surface 与 provider/profile inventory，再看 Route Evidence 确认 route binding、selection case 和 streaming / auth / secret ref 证据，再看 Usage/Audit Evidence 核对 key scope、quota snapshot、trace、failure 和 audit decision，最后看 Evidence Review / Readiness 汇总 readiness rollup、evidence checklist、route / usage / audit key risk 和 locked distribution capability。Admin Operations Review / Readiness 放在管理端视角收口，用于把 tenant、audit、gateway 和 Production Ops 证据串成审查摘要；Admin Provider/Profile & Deployment Evidence Review / Readiness 则继续查看 provider/profile、model route、secret ref readiness、deployment status、operator risk 和 locked capability。两者都只是 evidence-only review surface。

Draft Designer 的保存路径需要额外区分：默认 `apps/radishmind-web/` 仍展示 sample / local draft；只有设置 `VITE_RADISHMIND_WORKFLOW_SAVED_DRAFT_SOURCE=dev-saved-draft-http`，并让 platform 同时启用 `RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1`、`RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP=1` 和保存所需的 `RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE=1`，才会通过 dev-only route 写入 memory dev store。该路径用于本地验证保存、读取、校验和版本冲突，不是 durable persistence、production API、publish 或 run。

### 3.8 使用 Image Path metadata-only runtime mapper

图片生成路径当前只允许 metadata-only runtime helper。开发者可以在离线检查或后续 response consumer 评审中导入：

```python
from services.runtime.image_artifact_runtime_mapper import (
    map_image_artifact_to_response_reference,
)
```

输入必须是已符合 `contracts/image-generation-artifact.schema.json` 的 `image_generation_artifact` metadata。成功结果只返回 artifact citation 和 metadata reference；失败结果只返回 fail-closed failure code，不返回 citation。该 helper 不打开文件、不访问网络、不读取图片二进制、不查 store、不生成 public URL、不调用 backend。完整契约见 [图片生成契约](contracts/image-generation.md)。

### 3.9 使用 Docker 部署资产

Docker 资产位于 `deploy/`，长期说明见 [部署目录说明](../deploy/README.md)。当前固定三种模式：

- `host_dev`：日常开发仍在宿主机直跑，使用 `scripts/run-platform-service.*` 和 `scripts/run-radishmind-console-dev.*`。
- `docker_local`：使用 `deploy/docker-compose.local.yaml` 本地 build platform / console 镜像，默认 `mock` provider，只用于本地容器 smoke。
- `docker_test` / `docker_prod`：共用 `deploy/docker-compose.yaml`，只拉取预构建镜像，通过 `RADISHMIND_IMAGE_TRACK=test|release` 或固定 `RADISHMIND_IMAGE_TAG` 区分环境。

本地容器 smoke 的命令边界是：

```bash
docker compose -f deploy/docker-compose.local.yaml config
docker compose -f deploy/docker-compose.local.yaml up --build -d
./scripts/run-python.sh scripts/run-platform-local-smoke.py --base-url http://127.0.0.1:7000 --check
docker compose -f deploy/docker-compose.local.yaml ps
docker compose -f deploy/docker-compose.local.yaml down
```

执行后应按 `scripts/checks/fixtures/production-ops-container-smoke-record-template.json` 把运行记录写入 `tmp/production-ops/container-smoke/`。该记录目录不提交。

部署态 compose 的静态检查和 runbook 检查已经纳入 `./scripts/check-repo-fast.sh` 与 `pwsh ./scripts/check-repo.ps1 -Fast`。这些检查默认不启动 Docker、不拉镜像、不声明 `container_smoke_ready`。

### 4. 跑本地候选模型输出

如果你要继续看 `RadishMind-Core` 本地候选输出，入口仍是：

```bash
./scripts/run-python.sh scripts/run-radishmind-core-candidate.py \
  --provider local_transformers \
  --model-dir /path/to/model \
  --output-dir tmp/radishmind-core-candidate-local \
  --allow-invalid-output \
  --validate-task
```

这仍属于模型评测 / 训练前治理链路，不等同于平台正式服务。

### 5. 当前协议兼容边界

当前平台必须区分两类兼容：

- 南向模型接入：平台去调用哪些模型和哪些 provider
- 北向协议输出：外部客户端如何调用 `RadishMind`

当前真实状态是：

- 南向已有一部分：`openai-compatible` 主入口、`HuggingFace`、`Ollama`、`gemini-native`、`anthropic-messages`，以及评测链路中的 `local_transformers`
- 北向已有第一版兼容面和只读产品面：`/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models`、`/v1/models/{id}`、`/v1/platform/overview`、`/v1/platform/local-smoke`、`/v1/session/metadata`、`/v1/tools/metadata`、blocked `/v1/tools/actions`、七条 fake-store-backed Control Plane Read-Side route、SSE bridge、provider/profile selection metadata 和 diagnostics discoverability 已对齐
- `P1 Runtime Foundation` 已达到 short close，当前不应继续把 provider/config/diagnostics/observability 同层细节当作主线
- 当前仍是窄切片：已经具备 Docker local/test/prod 的静态部署边界、镜像命名治理、deployment readiness 静态 smoke、container smoke runbook、运行记录模板、一次 `docker_local` container smoke 运行记录、provider capability matrix、provider health smoke、provider selection policy、provider retry/fallback policy 和 `provider-runtime-docs-refresh` 文档收口，但还缺真实镜像发布 workflow、production secret backend、部署隔离、外部 provider live health check、retry/fallback execution、console runtime config、测试环境 smoke、生产前复核记录，以及 session/tooling 的真实确认流接线、存储、执行和完整负向回归；P3 checklist 已把本地只读产品壳标为可用，并把这些生产前置条件继续保持为 `not_satisfied`，P2 现有 route / gate / negative regression 资产仍是 governance-only。没有测试或生产前复核窗口时，不把 Docker / deployment 当作默认开发主线

## 当前还不能算完成的能力

当前仓库还没有这些正式能力：

- production deployment package 与镜像发布 workflow
- production secret backend
- process supervisor 与环境隔离
- 外部 provider health check
- console production packaging / runtime config
- 完整 user workspace / production admin control plane React UI
- `apps/radishmind-web/` 的生产 API consumer、Radish OIDC / auth middleware、read store repository、数据库 query 和用户端 / 管理端写入能力
- Control Plane Read-Side 的真实 Radish OIDC / auth middleware、repository adapter、read store repository implementation、数据库 query 和 migration
- Workflow / Agent Runtime 的完整 builder mutation、durable draft persistence、validation result persistence、publish、executor、confirmation decision、execution unlock、business writeback、run replay 和 run resume；当前 saved draft HTTP route / web consumer 只属于 dev-only memory store 路径
- Image Path 的 artifact store、binary reader、public URL resolver、真实 backend adapter、安全 runtime、图片生成和 artifact upload
- 测试环境 smoke 和生产前复核记录
- 更完整的 route-level smoke、stream 组合和兼容性矩阵
- durable session/checkpoint/audit/result store、materialized checkpoint/result reader 和 recovery runbook
- 真实工具执行器、materialized tool result cache、上层确认流接线和完整 session/tooling 负向回归 implementation consumer

所以如果你问“现在怎么部署”，准确答案是：当前已有本地 CLI runtime、进程内 gateway、Go platform service、本地 runbook、启动 wrapper、config / deployment / diagnostics smoke、Docker local compose、测试 / 生产共用部署态 compose、镜像命名治理、deployment readiness 静态 smoke、container smoke runbook、container smoke 记录模板、一次 `docker_local` container smoke 通过记录、request observability、error taxonomy、bridge-backed provider/profile discoverability、`GET /v1/platform/overview` 只读产品 overview、`GET /v1/platform/local-smoke` 本地 readiness 摘要、overview / local-smoke consumer smoke、本地 console 壳、Dev Diagnostics、`Local Readiness` 面板、Provider/Profile Details、Stop-line Details、overview / local-smoke failure surface、console shell / behavior / visual smoke record / dev entry / production boundary checks、P3 checklist、session/tooling metadata smoke、七条 fake-store-backed Control Plane Read-Side route、TypeScript read consumer contract、formal UI boundary/readiness、read-side repository contract type、静态 contract smoke runner、`ControlPlaneReadRepository` interface + fake-store bridge、`apps/radishmind-web/` 离线 read-only product UI shell、P2 design gates 和 P2 governance rollup checks；本地只读壳已可用，Docker 静态部署边界已可检查，本地 mock 容器 smoke 已跑通，但还没有完整 production deployment、真实镜像发布、测试环境 smoke、production preflight、console production packaging、真实 Radish OIDC / auth middleware、repository adapter、read store repository implementation、数据库 query、真实 executor、durable store、confirmation 接线、materialized result reader、长期记忆、业务写回或 replay。没有明确运行窗口时，部署资产保持等待，默认继续推进产品功能骨架。

## 读文档顺序

如果你刚接触这个仓库，建议按这个顺序读：

1. [文档入口](README.md)
2. 项目总览与使用指南
3. [当前推进焦点](radishmind-current-focus.md)
4. [产品范围](radishmind-product-scope.md)
5. [战略定义](radishmind-strategy.md)
6. [能力矩阵](radishmind-capability-matrix.md)
7. [系统架构](radishmind-architecture.md)
8. 按任务需要继续读 UI、集成契约、部署、脚本、数据集或训练专题。

## 默认不要做

- 不把 `RadishMind` 做成上层业务真相源
- 不默认把 builder/guided 结果当成 raw 晋级证据
- 不在上层项目还没具备真实挂载点时继续细化假想接线
- 不把 production deployment、session、tooling 或完整外部兼容矩阵写成“已经具备”
- 不默认下载大模型、数据集或权重

# RadishMind 部署目录说明

更新时间：2026-06-19

## 目录职责

`deploy/` 保存 RadishMind 当前可检查的 Docker 部署边界。它用于说明和验证 local / test / prod 三类 Docker 形态，不是 production ready 声明。

当前文件：

- `docker-compose.local.yaml`：本地容器 smoke 编排，允许本机 build platform 与 console 镜像，默认 `mock` provider，端口默认为 `7000` / `4000`。
- `docker-compose.saved-draft-dev-test.yaml`：Saved Draft PostgreSQL 17 开发 / 测试编排，只绑定 loopback `55432`，拆分 migration 与 runtime role。
- `postgres/saved-draft-dev-test-init.sql`：本地测试库 runtime role 初始化和默认表权限，不包含 secret。
- `docker-compose.yaml`：测试 / 生产共用的部署态 compose，只引用预构建镜像，不执行本地 build。
- `.env.example`：部署态 compose 的非密钥配置样例；真实 `.env` 不提交。

## Docker 模式

### `host_dev`

日常开发仍默认在宿主机直跑：

macOS / Linux / WSL 使用：

```bash
./scripts/run-platform-service.sh serve
./scripts/run-radishmind-console-dev.sh
```

Windows / PowerShell 使用：

```powershell
pwsh ./scripts/run-platform-service.ps1 serve
pwsh ./scripts/run-radishmind-console-dev.ps1
```

`host_dev` 不默认走 Compose。

### `docker_local`

本地容器验证使用 `docker-compose.local.yaml`：

```bash
docker compose -f deploy/docker-compose.local.yaml config
docker compose -f deploy/docker-compose.local.yaml up --build -d
./scripts/run-python.sh scripts/run-platform-local-smoke.py --base-url http://127.0.0.1:7000 --check
docker compose -f deploy/docker-compose.local.yaml ps
docker compose -f deploy/docker-compose.local.yaml down
```

这条路径只证明本地容器 smoke 可运行。它使用 `mock` provider，不代表测试环境、生产环境或 provider credential readiness。

运行记录应按 `scripts/checks/fixtures/production-ops-container-smoke-record-template.json` 的字段写入 `tmp/production-ops/container-smoke/`。该目录用于本地运行证据，不提交入仓。

Saved Draft PostgreSQL dev/test 使用独立入口：

```bash
./scripts/run-workflow-saved-draft-postgres-dev-test.sh check
./scripts/run-radishmind-web-dev.sh --mode dev-live --saved-draft-postgres-dev-test
./scripts/run-radishmind-web-dev.sh --mode dev-live --saved-draft-postgres-dev-test --workflow-diagnostics-dev
./scripts/run-workflow-saved-draft-postgres-dev-test.sh down
```

该 dev/test 入口会同时准备独立的 Saved Draft 与 Workflow Run schema；Platform runtime 只持有 DML 权限，migration runner 仍需显式执行。真实运行历史使用 `/v1/user-workspace/workflow-runs`，不复用旧 `/v1/user-workspace/runs` read fixture。

`check` 会启动数据库、执行真实集成测试，并重新应用 reviewed schema 供浏览器联调；`down` 保留命名卷。默认 `radishmind_migrator` 只用于显式 migration，`radishmind_runtime` 没有 schema `CREATE` 权限。该 Compose 不等于 test deployment 或 production database readiness。

### `docker_test`

测试环境复用 `docker-compose.yaml`，默认使用 `RADISHMIND_IMAGE_TRACK=test` 或显式 `RADISHMIND_IMAGE_TAG`：

```powershell
Copy-Item deploy/.env.example deploy/.env
docker compose --env-file deploy/.env -f deploy/docker-compose.yaml config
docker compose --env-file deploy/.env -f deploy/docker-compose.yaml pull
docker compose --env-file deploy/.env -f deploy/docker-compose.yaml up -d
docker compose --env-file deploy/.env -f deploy/docker-compose.yaml ps
docker compose --env-file deploy/.env -f deploy/docker-compose.yaml down
```

执行前必须已有测试镜像、测试 provider profile、非生产 secret 来源和明确测试部署窗口。当前仓库尚未实现镜像发布 workflow。

### `docker_prod`

生产部署也复用 `docker-compose.yaml`，默认使用 `RADISHMIND_IMAGE_TRACK=release` 或固定 `v*-release` tag。生产环境必须先满足：

- production secret backend
- 正式 provider profile 与 provider health policy
- 正式 auth / CORS policy
- 外部 HTTPS 反向代理
- 部署后 smoke 与生产前复核记录
- process supervisor 或等价运行托管方案
- console runtime config

当前仓库没有声明这些条件已满足。

## 配置边界

`deploy/.env.example` 只保存非密钥配置样例。真实 credential、API key、生产 secret、证书路径和 provider token 不得提交。

### Production secret backend contract

`production-secret-backend-contract` 已由 `scripts/checks/fixtures/production-ops-secret-backend-contract.json` 与 `scripts/check-production-ops-secret-backend-contract.py` 固定为治理切片。当前只定义未来 external secret backend adapter contract、`RADISHMIND_SECRET_SOURCE` 这类 secret reference、脱敏输出和禁止项；`deploy/.env.example` 不是 secret backend，也不保存真实 credential。

该切片不实现真实云 secret 服务、不写入真实 secret、不调用云 API、不声明 production ready。真实 production secret backend、secret rotation runtime、production secret audit store 和 provider health policy 仍必须在后续任务中另行实现和验证。

`contracts/production-secret-reference.schema.json` 进一步固定 provider profile 到 `secret_ref` 的 reference-only manifest。它只允许 `ref:` 引用、`credential_state`、`secret_backend_configured`、`secret_ref_present`、`missing_secret_refs` 和 `field_sources` 这类脱敏字段；`scripts/checks/fixtures/production-secret-reference-basic.json` 只作为 test / production 两类环境的结构样例，不包含真实 key、token、cookie、authorization、provider raw URL 或 secret value。

这层 schema 只说明未来部署配置应该如何引用 secret，不说明 secret backend 已经存在。部署或 readiness 文档不得把 `RADISHMIND_SECRET_SOURCE`、`.env.example`、secret reference fixture 写成 provider credential readiness。

`production-secret-backend-config-secret-ref-readiness-v1` 已把 `config-secret-ref-readiness` 固定为可检查证据：后续配置层只能报告 `secret_backend_configured`、`secret_ref_present`、`missing_secret_refs` 和 `field_sources` 等脱敏状态，不能读取 secret value、调用 resolver 或调用云 secret 服务。该 readiness 不改变 `.env.example` 的职责；`RADISHMIND_SECRET_SOURCE` 仍只是外部 secret 来源要求，不是 secret backend。

`production-secret-backend-provider-profile-secret-binding-readiness-v1` 已把 `provider-profile-secret-binding` 固定为可检查证据：后续 provider/profile 只能声明 `credential_requirement`、`secret_ref_status`、环境绑定和脱敏诊断状态，不能把 `secret_ref_status=present` 写成 credential resolved，不能从 test / production secret ref 之间 fallback，也不能调用 resolver 或云 secret 服务。该 readiness 不改变 compose 或 `.env.example` 的职责，仍不声明 production secret backend ready。

`production-secret-backend-secret-resolver-interface-disabled-readiness-v1` 已把 `secret-resolver-interface-disabled` 固定为可检查证据：后续 resolver interface 在未启用前只能返回 disabled / fail-closed 脱敏状态，不能创建 credential handle，不能 fallback 到 `RADISHMIND_PLATFORM_API_KEY`、mock provider、local-smoke profile、fixture credential、跨环境 `secret_ref`、fake resolver 或 fake query executor。该 readiness 不改变 compose 或 `.env.example` 的职责，仍不声明 production secret backend ready。

`production-secret-backend-operator-runbook-negative-gates-readiness-v1` 已把 `operator-runbook-and-negative-gates` 固定为可检查证据：后续 operator runbook 只能记录人工 approval、test / production secret source、脱敏验证、smoke record reference 和 negative gate results，不能保存 secret value、provider raw URL、DSN、cloud credential 或 credential handle。该 readiness 不改变 compose 或 `.env.example` 的职责，不创建 resolver runtime、fake resolver、cloud SDK、DB provider、repository mode 或 production API，仍不声明 production secret backend ready。

`production-secret-backend-rotation-audit-policy-readiness-v1` 已把 `rotation-and-audit-policy` 固定为可检查证据：后续 rotation / audit policy 只能记录 rotation trigger、approval / change window、secret ref version reference、rollback policy、sanitized verification、audit event fields 和 failure code，不能保存 secret value、provider raw URL、DSN、cloud credential 或 credential handle。该 readiness 不改变 compose 或 `.env.example` 的职责，不创建 rotation runtime、production secret audit store、audit writer、resolver runtime、fake resolver、cloud SDK、DB provider、repository mode 或 production API，仍不声明 production secret backend ready。

`production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1` 已把 `test-fixture-strategy` 与 fake resolver implementation entry review 固定为可检查证据：当前结论是 `test-fixture-strategy` 仍为 `required_before_implementation`，fake resolver implementation entry 不打开。该 readiness 不改变 compose 或 `.env.example` 的职责，不创建 resolver runtime、fake resolver runtime、no secret leakage smoke runtime、cloud SDK、DB provider、repository mode 或 production API，仍不声明 production secret backend ready。

`production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1` 已把 fake resolver static contract 与 no secret leakage smoke strategy 固定为可检查证据：后续 fake resolver 只能消费 allowlist 输入、输出脱敏状态和 credential handle metadata，不得接收或输出 secret value、provider raw URL、DSN、token、cookie 或 cloud credential。该 strategy 不改变 compose 或 `.env.example` 的职责，不创建 resolver runtime、fake resolver runtime、no secret leakage smoke runtime、cloud SDK、DB provider、repository mode 或 production API，仍不声明 production secret backend ready。

`production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1` 已把 `fake_resolver_implementation_task_card_entry_readiness_review_defined` 固定为可检查证据：当前只确认下一步可以创建 fake resolver implementation task card。该 review 不改变 compose 或 `.env.example` 的职责，不实现 resolver runtime、不把 test-only fake resolver runtime 写成 production resolver、不创建 no secret leakage smoke runtime、不调用云 secret 服务、不读取 secret value、不接 DB provider、不启用 repository mode 或 production API，仍不声明 production secret backend ready。

`production-secret-backend-fake-resolver-implementation-v1` 已把 `fake_resolver_implementation_task_card_defined` 固定为可检查证据：当前只创建 fake resolver implementation 的静态任务卡、fixture、checker 和 artifact guard。该 task card 不改变 compose 或 `.env.example` 的职责，不实现 fake resolver runtime、不创建 no secret leakage smoke runtime、不解析 secret、不创建 credential handle、不连接数据库、不调用云 secret 服务、不启用 repository mode 或 production API，仍不声明 production secret backend ready。

`production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1` 已把 `fake_resolver_runtime_implementation_entry_review_defined` 固定为可检查证据：当前只确认下一步可以单独创建 test-only fake resolver runtime。该 review 不改变 compose 或 `.env.example` 的职责，不实现 fake resolver runtime、不创建 no secret leakage smoke runtime、不解析 secret、不创建 credential handle、不连接数据库、不调用云 secret 服务、不启用 repository mode 或 production API，仍不声明 production secret backend ready。

`production-secret-backend-fake-resolver-runtime-implementation-v1` 已把 `fake_resolver_runtime_test_only_implemented` 固定为可检查证据：当前只实现 test-only、默认 disabled 的 fake resolver runtime，并由 Go 单测覆盖 placeholder secret ref、sanitized diagnostics、offline no leakage smoke 和 zero external side effects。该 runtime 不改变 compose 或 `.env.example` 的职责，不实现 production resolver runtime、不创建 no secret leakage smoke runtime、不解析真实 secret、不创建 credential payload、不连接数据库、不调用云 secret 服务、不启用 repository mode 或 production API，仍不声明 production secret backend ready。

部署态 compose 通过这些变量区分测试和生产：

- `RADISHMIND_IMAGE_REGISTRY`
- `RADISHMIND_IMAGE_TRACK`
- `RADISHMIND_IMAGE_TAG`
- `RADISHMIND_PUBLIC_BASE_URL`
- `RADISHMIND_SECRET_SOURCE`
- `RADISHMIND_PLATFORM_PROVIDER`
- `RADISHMIND_PLATFORM_PROVIDER_PROFILE`
- `RADISHMIND_PLATFORM_MODEL`
- `RADISHMIND_PLATFORM_HTTP_BIND`
- `RADISHMIND_CONSOLE_HTTP_BIND`

## 验证入口

静态边界检查：

```bash
./scripts/run-python.sh scripts/check-production-ops-docker-local-compose.py
./scripts/run-python.sh scripts/check-production-ops-docker-test-prod-compose.py
./scripts/run-python.sh scripts/check-production-ops-docker-image-build-publish.py
./scripts/run-python.sh scripts/check-production-ops-deployment-readiness-smoke.py
./scripts/run-python.sh scripts/check-production-ops-container-smoke-runbook.py
./scripts/run-python.sh scripts/check-production-ops-container-smoke-record-template.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-contract.py
./scripts/run-python.sh scripts/check-production-secret-reference-contract.py
```

仓库快速基线：

```bash
./scripts/check-repo-fast.sh
```

这些检查默认不启动 Docker、不拉镜像、不提交运行记录。

## 停止线

- 不把 `docker_local` 写成测试或生产 ready。
- 不把 `mock` provider 写成真实 provider readiness。
- 不把 `docker compose down` cleanup 写成 process supervisor。
- 不把 `.env.example` 写成 secret backend。
- 不把 `fake_resolver_implementation_task_card_defined`、`fake_resolver_runtime_implementation_entry_review_defined` 或 `fake_resolver_runtime_test_only_implemented` 写成 production secret backend ready；它们不调用云 secret 服务、不解析真实 secret、不连接数据库。
- 不把 deployment readiness 静态展开写成 container smoke。
- 不把 runbook 或 record template 写成已经完成的运行记录。

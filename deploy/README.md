# RadishMind 部署目录说明

更新时间：2026-05-26

## 目录职责

`deploy/` 保存 RadishMind 当前可检查的 Docker 部署边界。它用于说明和验证 local / test / prod 三类 Docker 形态，不是 production ready 声明。

当前文件：

- `docker-compose.local.yaml`：本地容器 smoke 编排，允许本机 build platform 与 console 镜像，默认 `mock` provider，端口默认为 `7000` / `4000`。
- `docker-compose.yaml`：测试 / 生产共用的部署态 compose，只引用预构建镜像，不执行本地 build。
- `.env.example`：部署态 compose 的非密钥配置样例；真实 `.env` 不提交。

## Docker 模式

### `host_dev`

日常开发仍默认在宿主机直跑：

```powershell
pwsh ./scripts/run-platform-service.ps1 serve
pwsh ./scripts/run-radishmind-console-dev.ps1
```

Linux / WSL 使用对应 `.sh` wrapper。`host_dev` 不默认走 Compose。

### `docker_local`

本地容器验证使用 `docker-compose.local.yaml`：

```powershell
docker compose -f deploy/docker-compose.local.yaml config
docker compose -f deploy/docker-compose.local.yaml up --build -d
python scripts/run-platform-local-smoke.py --base-url http://127.0.0.1:7000 --check
docker compose -f deploy/docker-compose.local.yaml ps
docker compose -f deploy/docker-compose.local.yaml down
```

这条路径只证明本地容器 smoke 可运行。它使用 `mock` provider，不代表测试环境、生产环境或 provider credential readiness。

运行记录应按 `scripts/checks/fixtures/production-ops-container-smoke-record-template.json` 的字段写入 `tmp/production-ops/container-smoke/`。该目录用于本地运行证据，不提交入仓。

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

该切片不实现真实云 secret 服务、不写入真实 secret、不调用云 API、不声明 production ready。真实 production secret backend、secret rotation policy、production secret audit store 和 provider health policy 仍必须在后续任务中另行实现和验证。

`contracts/production-secret-reference.schema.json` 进一步固定 provider profile 到 `secret_ref` 的 reference-only manifest。它只允许 `ref:` 引用、`credential_state`、`secret_backend_configured`、`secret_ref_present`、`missing_secret_refs` 和 `field_sources` 这类脱敏字段；`scripts/checks/fixtures/production-secret-reference-basic.json` 只作为 test / production 两类环境的结构样例，不包含真实 key、token、cookie、authorization、provider raw URL 或 secret value。

这层 schema 只说明未来部署配置应该如何引用 secret，不说明 secret backend 已经存在。部署或 readiness 文档不得把 `RADISHMIND_SECRET_SOURCE`、`.env.example`、secret reference fixture 写成 provider credential readiness。

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

```powershell
python scripts/check-production-ops-docker-local-compose.py
python scripts/check-production-ops-docker-test-prod-compose.py
python scripts/check-production-ops-docker-image-build-publish.py
python scripts/check-production-ops-deployment-readiness-smoke.py
python scripts/check-production-ops-container-smoke-runbook.py
python scripts/check-production-ops-container-smoke-record-template.py
python scripts/check-production-ops-secret-backend-contract.py
python scripts/check-production-secret-reference-contract.py
```

仓库快速基线：

```powershell
pwsh ./scripts/check-repo.ps1 -Fast
```

这些检查默认不启动 Docker、不拉镜像、不提交运行记录。

## 停止线

- 不把 `docker_local` 写成测试或生产 ready。
- 不把 `mock` provider 写成真实 provider readiness。
- 不把 `docker compose down` cleanup 写成 process supervisor。
- 不把 `.env.example` 写成 secret backend。
- 不把 deployment readiness 静态展开写成 container smoke。
- 不把 runbook 或 record template 写成已经完成的运行记录。

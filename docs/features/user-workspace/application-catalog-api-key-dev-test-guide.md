# 应用目录与 API 密钥开发测试指南

更新时间：2026-07-21

## 适用范围

本文说明如何在本地开发测试态完成以下闭环：

1. 使用受验证的管理身份创建应用。
2. 为应用签发 API 密钥。
3. 使用 Bearer 密钥调用模型网关。
4. 检查密钥最近使用时间并吊销密钥。
5. 验证吊销密钥或归档应用后，调用在进入 bridge / provider 前失败关闭。

当前实现支持 `memory_dev`、默认本地产品档 `sqlite_dev` 与显式 `postgres_dev_test`。它不代表生产身份系统、生产密钥服务、公开生产网关或正式配额计费已经可用。产品 Web 已支持应用目录创建、API 密钥签发与吊销、一次性令牌内存交接，以及使用 Bearer 凭据进入既有 Gateway Playground；真实浏览器连续验收与重启复验已经完成。

2026-07-14 已完成当时七组件共享 SQLite 本地产品链和真实 PostgreSQL 专项门禁：migration、角色隔离、类型 / 索引、稳定分页、advisory lock、多连接并发、认证 / 吊销与应用归档竞态、重启恢复和 no-fallback 均已通过。2026-07-21 聚合 runtime 已加入第八个 Prompt Template owner，Prompt Runtime Assignment / Event 复用 Workflow Run Store 投影。2026-07-15 已完成 Web 一次性交接、严格消费端、生产构建和真实浏览器连续链：应用配置、密钥签发、模型发现、单次 / 流式 / 取消调用、精确历史、最近使用、吊销后拒绝与重启恢复均成立；令牌未进入 URL、浏览器持久化介质、日志、浏览器产物或 SQLite 物理文件。

## 两类身份必须分开

应用目录和 API 密钥生命周期属于管理面，模型调用属于消费面，两者不得混用：

| 场景 | 凭证 | 用途 |
| --- | --- | --- |
| 管理应用和 API 密钥 | `X-RadishMind-Dev-Read-*` 开发身份头 | 表达租户、操作者和管理作用域 |
| 调用模型网关 | `Authorization: Bearer <API key>` | 从密钥记录恢复租户、工作区、应用、所有者和调用作用域 |

管理 API 会拒绝 RadishMind API 密钥。`api_key_dev_test` 模式下，模型网关也会拒绝同时携带 Bearer 密钥与 `X-RadishMind-Dev-Gateway-*` 身份头的请求，避免调用方覆盖可信身份。

## 启动 SQLite 本地产品链

需要同时启动 Platform 与 Web 并验证 API 密钥产品路径时，在仓库根目录执行：

```bash
./scripts/run-radishmind-web-dev.sh --mode dev-live --api-key-local-product --backend-url http://127.0.0.1:7100
```

Windows / PowerShell 使用：

```powershell
pwsh ./scripts/run-radishmind-web-dev.ps1 -Mode dev-live -APIKeyLocalProduct -BackendUrl http://127.0.0.1:7100
```

该入口统一启用 SQLite `local-product`、应用目录、配置 / 发布链、API 密钥生命周期、`api_key_dev_test` Gateway、模型目录、Playground 与请求历史，并会验证未携带密钥的 `/v1/models` 返回 `api_key_missing`。长期运行的开发服务应由开发者在本机终端保持，验证结束后停止。

只启动 Platform 或做配置检查时执行：

```bash
RADISHMIND_GATEWAY_AUTH_MODE=api_key_dev_test \
RADISHMIND_PLATFORM_LISTEN_ADDR=127.0.0.1:7100 \
./scripts/run-platform-service.sh config-check

RADISHMIND_GATEWAY_AUTH_MODE=api_key_dev_test \
RADISHMIND_PLATFORM_LISTEN_ADDR=127.0.0.1:7100 \
./scripts/run-platform-service.sh serve
```

这是长期运行的本地服务命令，应由开发者在本机终端启动，并在验证结束后停止。默认 `local-product` 档会一次性选择八组件 `sqlite_dev`、仓库根 `var/sqlite-dev/radishmind.db` 和所需开发门禁；`config-check` 只检查配置，不创建数据库。应用、密钥、调用记录、配置草案、Prompt Template 和工作流数据会在服务重启后恢复。需要回到显式内存或 PostgreSQL 组件配置时，改用 `--profile configured` 并完整提供对应门禁与 store 配置。

以下示例统一使用：

- 服务地址：`http://127.0.0.1:7100`
- 租户：`tenant_demo`
- 工作区：`workspace_demo`
- 操作者：`subject_demo`

管理请求需要携带完整开发身份头：

```text
X-RadishMind-Dev-Read-Identity: user-workspace-api-key-guide
X-RadishMind-Dev-Read-Tenant: tenant_demo
X-RadishMind-Dev-Read-Subject: subject_demo
X-RadishMind-Dev-Read-Scopes: applications:read,applications:write,applications:archive,api_keys:read,api_keys:write,api_keys:revoke
X-RadishMind-Dev-Read-Audit: local-guide
```

## 创建与维护应用

创建应用：

```bash
curl --fail-with-body --silent --show-error \
  --request POST http://127.0.0.1:7100/v1/user-workspace/applications \
  --header 'Content-Type: application/json' \
  --header 'X-RadishMind-Dev-Read-Identity: user-workspace-api-key-guide' \
  --header 'X-RadishMind-Dev-Read-Tenant: tenant_demo' \
  --header 'X-RadishMind-Dev-Read-Subject: subject_demo' \
  --header 'X-RadishMind-Dev-Read-Scopes: applications:read,applications:write,applications:archive,api_keys:read,api_keys:write,api_keys:revoke' \
  --header 'X-RadishMind-Dev-Read-Audit: local-guide' \
  --data '{"workspace_id":"workspace_demo","display_name":"本地网关示例","description":"用于验证应用与密钥闭环","application_kind":"agent"}'
```

从响应的 `record.application_id` 取得服务端生成的应用标识，并在后续命令中替换 `<APPLICATION_ID>`。应用类型只允许：

- `workflow_copilot`
- `docs_qa`
- `agent`
- `prompt_application`

列表查询支持 `lifecycle_state`、`application_kind`、`limit` 与 `cursor`：

```bash
curl --fail-with-body --silent --show-error \
  'http://127.0.0.1:7100/v1/user-workspace/applications?workspace_id=workspace_demo&lifecycle_state=active' \
  --header 'X-RadishMind-Dev-Read-Identity: user-workspace-api-key-guide' \
  --header 'X-RadishMind-Dev-Read-Tenant: tenant_demo' \
  --header 'X-RadishMind-Dev-Read-Subject: subject_demo' \
  --header 'X-RadishMind-Dev-Read-Scopes: applications:read' \
  --header 'X-RadishMind-Dev-Read-Audit: local-guide'
```

更新和归档采用 `expected_version` 乐观并发控制。调用前应读取最新 `record_version`，版本不一致时服务返回冲突，并提供当前版本和状态，不应盲目重试覆盖。

## 签发 API 密钥

签发密钥时，应用必须处于 `active` 状态：

```bash
curl --fail-with-body --silent --show-error \
  --request POST http://127.0.0.1:7100/v1/user-workspace/api-keys \
  --header 'Content-Type: application/json' \
  --header 'X-RadishMind-Dev-Read-Identity: user-workspace-api-key-guide' \
  --header 'X-RadishMind-Dev-Read-Tenant: tenant_demo' \
  --header 'X-RadishMind-Dev-Read-Subject: subject_demo' \
  --header 'X-RadishMind-Dev-Read-Scopes: api_keys:write' \
  --header 'X-RadishMind-Dev-Read-Audit: local-guide' \
  --data '{"workspace_id":"workspace_demo","application_id":"<APPLICATION_ID>","display_name":"本地 SDK 密钥","scopes":["models:read","responses:invoke"],"expires_in_days":30}'
```

成功响应使用 `Cache-Control: no-store`，原始令牌只在本次响应的 `credential.token` 返回。列表、详情和吊销响应不会再次返回令牌，也不会暴露摘要。请立即保存所需值：

- `record.api_key_id`：用于查询和吊销。
- `record.record_version`：用于乐观并发控制。
- `credential.token`：用于 Gateway Bearer 认证，不应写入文档、日志、Git、URL 或命令历史。

如需在当前终端临时使用，可通过隐藏输入读取，不在命令行直接粘贴令牌：

```bash
read -s RADISHMIND_API_KEY
export RADISHMIND_API_KEY
```

完成验证后执行 `unset RADISHMIND_API_KEY`。

## 调用作用域

| Gateway 路由 | 必需作用域 |
| --- | --- |
| `GET /v1/models` | `models:read` |
| `GET /v1/models/{id}` | `models:read` |
| `POST /v1/chat/completions` | `chat:invoke` |
| `POST /v1/responses` | `responses:invoke` |
| `POST /v1/messages` | `messages:invoke` |

使用刚签发的密钥读取模型目录：

```bash
curl --fail-with-body --silent --show-error \
  http://127.0.0.1:7100/v1/models \
  --header "Authorization: Bearer ${RADISHMIND_API_KEY}"
```

认证成功后，服务从密钥记录恢复可信调用方上下文，原子更新 `last_used_at`，并将脱敏调用方记录为 `api_key:<API_KEY_ID>`。令牌和令牌摘要都不会写入 Gateway 请求历史。

## 检查与吊销密钥

列出指定应用的密钥：

```bash
curl --fail-with-body --silent --show-error \
  'http://127.0.0.1:7100/v1/user-workspace/api-keys?workspace_id=workspace_demo&application_id=<APPLICATION_ID>' \
  --header 'X-RadishMind-Dev-Read-Identity: user-workspace-api-key-guide' \
  --header 'X-RadishMind-Dev-Read-Tenant: tenant_demo' \
  --header 'X-RadishMind-Dev-Read-Subject: subject_demo' \
  --header 'X-RadishMind-Dev-Read-Scopes: api_keys:read' \
  --header 'X-RadishMind-Dev-Read-Audit: local-guide'
```

列表支持 `effective_state=active|expired|revoked`、`limit` 与 `cursor`。使用详情响应中的最新 `record_version` 吊销：

```bash
curl --fail-with-body --silent --show-error \
  --request POST http://127.0.0.1:7100/v1/user-workspace/api-keys/<API_KEY_ID>/revoke \
  --header 'Content-Type: application/json' \
  --header 'X-RadishMind-Dev-Read-Identity: user-workspace-api-key-guide' \
  --header 'X-RadishMind-Dev-Read-Tenant: tenant_demo' \
  --header 'X-RadishMind-Dev-Read-Subject: subject_demo' \
  --header 'X-RadishMind-Dev-Read-Scopes: api_keys:revoke' \
  --header 'X-RadishMind-Dev-Read-Audit: local-guide' \
  --data '{"workspace_id":"workspace_demo","expected_version":<RECORD_VERSION>}'
```

吊销后再次调用 `/v1/models` 应在 bridge / provider 前被拒绝。归档应用也会使绑定密钥失效；应用归档不是密钥吊销的替代品，仍应显式吊销不再使用的凭证。

## PostgreSQL 开发测试模式

PostgreSQL 模式必须为八个产品组件分别提供 runtime DML 连接，并使用独立 migration 连接执行 DDL。仓库继续保留历史文件名的统一 runner；它实际覆盖八组产品 migration / repository、独立 Control Plane read schema 和 `configured` 启动档：

```bash
./scripts/run-workflow-saved-draft-postgres-dev-test.sh check
./scripts/run-workflow-saved-draft-postgres-dev-test.sh status
./scripts/run-workflow-saved-draft-postgres-dev-test.sh down
```

Windows / PowerShell 使用：

```powershell
pwsh ./scripts/run-workflow-saved-draft-postgres-dev-test.ps1 -Action check
pwsh ./scripts/run-workflow-saved-draft-postgres-dev-test.ps1 -Action status
pwsh ./scripts/run-workflow-saved-draft-postgres-dev-test.ps1 -Action down
```

`check` 会在 `127.0.0.1:55432` 启动 PostgreSQL 17，对数据库名包含 `test` 的实例执行破坏性 integration suite，覆盖 migration 重复应用、回滚 / 重应用、advisory lock、运行 / 迁移角色隔离、类型 / 索引、稳定分页、多连接并发、竞态、重启恢复和 no-fallback。通过后容器与已迁移 schema 会继续保留，便于显式联调；验证结束后必须执行 `down`，该动作停止容器但保留命名卷。

只需应用 migration 时使用 runner 的 `migrate`；只检查 marker 时使用 `status`。需要手动启动平台时，完整提供各组件 `*_STORE=postgres_dev_test`、runtime URL、开发门禁和 `RADISHMIND_GATEWAY_AUTH_MODE=api_key_dev_test`，再通过 `./scripts/run-platform-service.sh --profile configured config-check` 与 `serve` 启动；不要把 PostgreSQL 配置混入默认 `local-product` 档。

完整配置键、统一 migration runner 和角色边界见 [Platform Service Layer](../../../services/platform/README.md)。连接、schema marker、checksum、查询或认证更新失败时必须失败关闭，不得回退到 `memory_dev`。数据库连接字符串与密钥只保存在本地环境，不进入 `.env.example`、文档或提交。

本地 SQLite 与 PostgreSQL 并行的正式设计见 [本地 SQLite 开发持久化 v1](../../platform/local-sqlite-dev-persistence-v1.md)。应用目录、API 密钥、Gateway 请求历史及其关联的配置草案、发布候选、工作流草案和工作流运行都已支持 `memory_dev`、聚合 `sqlite_dev` 与显式 `postgres_dev_test`；三种模式保持相同的领域作用域、版本保护、稳定分页和失败关闭语义，但分别承担快速测试、本地连续开发和 PostgreSQL 同构验证职责。

## 常见失败语义

| 现象 | 检查项 |
| --- | --- |
| 路由提示显式启用 | 检查对应 `*_DEV_HTTP` 和 write gate |
| `scope_denied` | 检查管理身份作用域是否与操作匹配 |
| `credential_conflict` | 不要用 API 密钥调用管理 API，也不要混用 Gateway 开发身份头 |
| `application_unavailable` | 检查应用是否存在、属于当前主体且处于 `active` |
| `version_conflict` | 重新读取最新记录与 `record_version` 后再决定是否提交 |
| `revoked` 或 `expired` | 签发新密钥，不复用失效凭证 |
| `store_unavailable` | 检查 store selector、连接、迁移 marker 和 checksum；不会自动回退内存 |

## 相关设计与契约

- [服务/API 接入契约](../../contracts/service-api.md)
- [跨项目集成契约](../../radishmind-integration-contracts.md)
- [应用目录生命周期功能设计](application-catalog-lifecycle-dev-test-v1.md)
- [API 密钥生命周期与 Gateway 认证功能设计](api-key-lifecycle-gateway-dev-test-auth-v1.md)
- [Platform Service Layer](../../../services/platform/README.md)

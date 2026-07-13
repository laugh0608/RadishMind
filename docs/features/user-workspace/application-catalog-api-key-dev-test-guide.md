# 应用目录与 API 密钥开发测试指南

更新时间：2026-07-13

## 适用范围

本文说明如何在本地开发测试态完成以下闭环：

1. 使用受验证的管理身份创建应用。
2. 为应用签发 API 密钥。
3. 使用 Bearer 密钥调用模型网关。
4. 检查密钥最近使用时间并吊销密钥。
5. 验证吊销密钥或归档应用后，调用在进入 bridge / provider 前失败关闭。

当前实现支持 `memory_dev` 与显式 `postgres_dev_test`。它不代表生产身份系统、生产密钥服务、公开生产网关或正式配额计费已经可用。产品 Web 已支持应用目录的创建、编辑和归档；API 密钥签发与吊销目前通过 HTTP API 验证，Web 端仍只展示脱敏摘要。

## 两类身份必须分开

应用目录和 API 密钥生命周期属于管理面，模型调用属于消费面，两者不得混用：

| 场景 | 凭证 | 用途 |
| --- | --- | --- |
| 管理应用和 API 密钥 | `X-RadishMind-Dev-Read-*` 开发身份头 | 表达租户、操作者和管理作用域 |
| 调用模型网关 | `Authorization: Bearer <API key>` | 从密钥记录恢复租户、工作区、应用、所有者和调用作用域 |

管理 API 会拒绝 RadishMind API 密钥。`api_key_dev_test` 模式下，模型网关也会拒绝同时携带 Bearer 密钥与 `X-RadishMind-Dev-Gateway-*` 身份头的请求，避免调用方覆盖可信身份。

## 启动内存开发测试链

在仓库根目录执行：

```bash
RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1 \
RADISHMIND_CONTROL_PLANE_READ_AUTH_MODE=dev_headers \
RADISHMIND_APPLICATION_CATALOG_DEV_HTTP=1 \
RADISHMIND_APPLICATION_CATALOG_DEV_WRITE=1 \
RADISHMIND_API_KEY_LIFECYCLE_DEV_HTTP=1 \
RADISHMIND_API_KEY_LIFECYCLE_DEV_WRITE=1 \
RADISHMIND_GATEWAY_REQUEST_HISTORY_DEV=1 \
RADISHMIND_GATEWAY_AUTH_MODE=api_key_dev_test \
./scripts/run-platform-service.sh serve
```

这是长期运行的本地服务命令，应由开发者在本机终端启动，并在验证结束后停止。未显式设置 store 时，应用、密钥和调用记录都保存在当前进程内存中，服务重启后清空。

以下示例统一使用：

- 服务地址：`http://127.0.0.1:7000`
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
  --request POST http://127.0.0.1:7000/v1/user-workspace/applications \
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
  'http://127.0.0.1:7000/v1/user-workspace/applications?workspace_id=workspace_demo&lifecycle_state=active' \
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
  --request POST http://127.0.0.1:7000/v1/user-workspace/api-keys \
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
  http://127.0.0.1:7000/v1/models \
  --header "Authorization: Bearer ${RADISHMIND_API_KEY}"
```

认证成功后，服务从密钥记录恢复可信调用方上下文，原子更新 `last_used_at`，并将脱敏调用方记录为 `api_key:<API_KEY_ID>`。令牌和令牌摘要都不会写入 Gateway 请求历史。

## 检查与吊销密钥

列出指定应用的密钥：

```bash
curl --fail-with-body --silent --show-error \
  'http://127.0.0.1:7000/v1/user-workspace/api-keys?workspace_id=workspace_demo&application_id=<APPLICATION_ID>' \
  --header 'X-RadishMind-Dev-Read-Identity: user-workspace-api-key-guide' \
  --header 'X-RadishMind-Dev-Read-Tenant: tenant_demo' \
  --header 'X-RadishMind-Dev-Read-Subject: subject_demo' \
  --header 'X-RadishMind-Dev-Read-Scopes: api_keys:read' \
  --header 'X-RadishMind-Dev-Read-Audit: local-guide'
```

列表支持 `effective_state=active|expired|revoked`、`limit` 与 `cursor`。使用详情响应中的最新 `record_version` 吊销：

```bash
curl --fail-with-body --silent --show-error \
  --request POST http://127.0.0.1:7000/v1/user-workspace/api-keys/<API_KEY_ID>/revoke \
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

持久化模式必须为应用目录、API 密钥和 Gateway 请求历史分别配置 runtime DML 连接，并先使用独立 migration 连接执行手动迁移：

```bash
cd services/platform
go run ./cmd/radishmind-application-catalog-migrate status
go run ./cmd/radishmind-application-catalog-migrate up
go run ./cmd/radishmind-api-key-migrate status
go run ./cmd/radishmind-api-key-migrate up
go run ./cmd/radishmind-gateway-request-migrate status
go run ./cmd/radishmind-gateway-request-migrate up
```

完整配置键和角色边界见 [Platform Service Layer](../../../services/platform/README.md)。连接、schema marker、checksum、查询或认证更新失败时必须失败关闭，不得回退到 `memory_dev`。数据库连接字符串与密钥只保存在本地环境，不进入 `.env.example`、文档或提交。

本地 SQLite 与 PostgreSQL 并行的正式设计见 [本地 SQLite 开发持久化 v1](../../platform/local-sqlite-dev-persistence-v1.md)。当前应用目录、API 密钥和 Gateway 请求历史的可运行 store 仍只有 `memory_dev` 与 `postgres_dev_test`，不得把设计稿误写成已实现的 `sqlite_dev`。

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

# 应用解除归档与安全重新启用（开发 / 测试态）v1

更新时间：2026-07-29

状态：`application_unarchive_safe_reactivation_dev_test_v1_completed`

## 设计结论

开发测试态应用目录需要补齐一次由人工明确发起、可审计、可并发复验的 `archived -> active` 转换。解除归档不是只读历史操作：现有 API Key、运行时绑定和交互会话都以应用当前活动状态作为资格条件，其中仍处于自身有效状态的资源可能在应用恢复后重新获得调用资格。

因此，本专题采用以下安全边界：

- 复用现有 `active` / `archived` 两态，不新增无法解释下游资格的过渡状态；
- 解除归档要求 `applications:archive` 与 `applications:write` 的单次组合成员资格判定；
- 请求必须携带当前 `record_version`，并显式确认理解既有访问资格可能恢复；
- 应用目录只原子修改自己的记录，不创建、吊销、激活或重写任何下游资源；
- 下游资源恢复与否继续由“应用活动态 + 资源自身状态 + 原有作用域 / 漂移规则”共同决定；
- Web 必须先展示重新启用影响，再允许提交，不把解除归档伪装成无副作用的普通按钮。

## 准入证据

[应用目录与生命周期（开发/测试态）v1](application-catalog-lifecycle-dev-test-v1.md) 已完成创建、更新、归档、三种存储与下游活动态约束，但明确把反归档留给独立设计。当前连续产品路径已经具备以下事实：

1. 归档会阻断新的配置草案、发布候选、应用交互、Workflow Definition、Prompt / Agent Runtime 与 Gateway API Key 鉴权。
2. 归档不级联吊销 API Key、不删除运行时绑定、不关闭会话，也不改写草案、候选、定义或运行历史。
3. 因此，误归档或临时停用后的恢复不能通过重建应用安全替代；重建会产生新的 `application_id`，并割裂现有历史与资源绑定。
4. 直接清空 `archived_at` 又会让既有有效访问资格静默恢复，缺少权限、确认、并发与审计契约。

这构成当前应用生命周期的真实恢复缺口，准入独立的安全重新启用能力；不以此扩展物理删除、批量生命周期或生产授权。

## 用户目标

工作区应用所有者在确认重新启用影响后，可以恢复一个已归档应用，并继续使用同一 `application_id`、历史记录和既有受控资源。

完成后用户应能：

1. 在已归档应用详情中看到恢复影响与停止线；
2. 使用当前记录版本提交解除归档；
3. 在成功后回到活动应用视图，继续编辑元数据或创建新配置；
4. 清楚知道哪些既有凭据 / 绑定可能恢复资格，哪些草案 / 候选仍需按原规则复验；
5. 在版本冲突、权限不足、确认缺失或存储失败时看到稳定失败，不产生部分状态。

## 生命周期与数据语义

允许的新增转换只有：

```text
archived --unarchive--> active
```

成功转换必须在同一个应用目录 owner 内原子完成：

- `lifecycle_state` 从 `archived` 改为 `active`；
- `record_version` 加一；
- `updated_at` 使用服务端 UTC 时间；
- `archived_at` 清空；
- `updated_by_actor_ref`、`request_id` 与 `audit_ref` 写入本次请求；
- `application_id`、所有者、创建信息和可编辑元数据保持不变。

以下请求必须失败关闭：

- 对活动应用执行解除归档；
- `expected_version` 不是正整数或已过期；
- 未显式确认既有访问资格可能恢复；
- 请求工作区与已验证活动工作区不一致；
- 缺少任一组合权限；
- owner 作用域不匹配、记录不存在或存储不可用。

解除归档不是幂等成功。重复请求应返回 `application_catalog_transition_invalid`；并发陈旧请求返回 `application_catalog_version_conflict`，并携带当前记录版本和生命周期。

## 下游重新启用语义

| 下游资源 | 解除归档后的语义 |
| --- | --- |
| API Key | 未吊销、未过期且作用域仍有效的 Key 可在 Gateway 再次通过应用活动态检查；不会签发新凭据，也不会恢复已吊销 Key |
| Workflow / Prompt / Agent / RAG Runtime Assignment | 自身仍为活动且绑定证据仍满足既有规则时可重新获得运行资格；不会新建或改写 assignment |
| Application Interaction Session | 未关闭会话仍按既有会话状态与运行时权威复验；不会自动创建 turn 或 Run |
| 配置草案 | 既有记录保持原版本；恢复后可创建或保存新版本，但仍执行应用基线和 CAS |
| 发布候选 / Workflow Definition | 既有审查决定不被重写；由于应用 `updated_at` 变化，原有基线继续按现有漂移规则复验 |
| Run / Request History / Audit | 始终保持历史只读；解除归档不重放、不续跑、不补写历史 |

解除归档请求本身不得调用 Gateway、Provider、模型、工具或网络，不得生成 API Key、Session、Turn、Run、草案、候选或 assignment。

## HTTP 契约

新增开发测试态端点：

```text
POST /v1/user-workspace/applications/{application_id}/unarchive
```

请求体：

```json
{
  "workspace_id": "workspace_demo",
  "expected_version": 4,
  "acknowledge_existing_access_reactivation": true
}
```

规则：

- 严格拒绝未知字段；
- `acknowledge_existing_access_reactivation` 必须精确为 `true`，缺失或 `false` 返回 `application_catalog_payload_invalid`，且不调用 repository；
- 认证、活动工作区选择与 membership 使用现有 workspace mutation authorization；
- scope 与 membership 同时要求 `applications:archive`、`applications:write`；
- 成功与失败继续使用 `applicationCatalogEnvelope`，不增加第二套响应协议；
- 继续受 `ApplicationCatalogDevHTTPEnabled` 与 `ApplicationCatalogDevWriteEnabled` 约束。

## 存储与并发

- 扩展现有 `applicationCatalogRepository`，为 Memory、SQLite 与 PostgreSQL 增加 `Unarchive`；
- SQLite / PostgreSQL 复用现有表与 schema，不新增 migration；
- 数据库更新必须在 `record_version = expected_version AND lifecycle_state = 'archived'` 条件下 CAS；
- 受影响行数为零时，精确重读并区分 not found、version conflict 与 transition invalid；
- 存储异常不得回退到内存或预置应用列表；
- Memory、SQLite 与 PostgreSQL 必须通过同一轮生命周期语义回归。

## Web 交互

已归档应用详情增加“解除归档”入口和两步确认：

1. 第一层说明应用将恢复活动态，既有未吊销 API Key、活动运行时绑定和未关闭会话可能重新获得资格。
2. 第二层明确说明不会自动创建调用、运行或凭据，旧草案和候选仍按版本 / 漂移规则复验。
3. 最终提交固定发送 `acknowledge_existing_access_reactivation: true` 与当前 `record_version`。

成功后：

- 以服务端返回记录替换本地记录；
- 切换到活动应用筛选；
- 更新选中应用上下文，使下游面板按新的活动态重新读取；
- 不保留确认框，不伪造其他 owner 的刷新成功。

冲突时保留当前选择并展示服务端版本，用户必须显式刷新后重新审查影响；失败时不把本地记录改成活动态。

## 实施顺序

### 批次 A：领域与三种存储

- 新增 service / repository `Unarchive`；
- 固定确认、版本与状态转换语义；
- 完成 Memory、SQLite、PostgreSQL 的 CAS 与重启回归；
- 扩展 Gateway API Key 归档拒绝测试，证明同一未吊销 Key 在显式解除归档后按原资格恢复。

### 批次 B：HTTP 与权限

- 注册严格 unarchive route；
- 接入 `applications:archive + applications:write` 单次组合授权；
- 覆盖确认缺失、未知字段、workspace 漂移、权限拒绝、OIDC membership unavailable、版本冲突和重复转换；
- 证明授权 / payload 失败时 repository 与外部副作用为零。

### 批次 C：Web 与连续验收

- 扩展严格 consumer、状态映射、请求头和单元测试；
- 完成已归档应用两步恢复审查及活动上下文交接；
- 使用 SQLite 本地产品链和真实浏览器验证 archive → 下游拒绝 → unarchive → 既有 Key 恢复资格 → 新配置可继续；
- 完成 PostgreSQL 集成、仓库级门禁、隐私复验和阶段文档收口。

## 验收

必须形成以下可复验证据：

1. Memory、SQLite 与 PostgreSQL 均只允许 `archived -> active`，版本单调递增且 `archived_at` 清空。
2. 两个并发解除归档请求最多一个成功，另一个得到带当前状态的版本冲突。
3. `applications:archive` 或 `applications:write` 任一缺失时，单次 membership decision 失败且 application repository 零写入。
4. 确认字段缺失 / 为假、未知字段和 workspace 漂移均失败关闭。
5. 未吊销 Key 在归档时拒绝、解除归档后恢复；已吊销 Key 不恢复。
6. 解除归档不创建任何下游资源，不调用 Gateway / Provider / 模型 / 工具 / 网络。
7. Web 成功后切换活动视图并更新应用上下文；失败或冲突时不伪造活动态。
8. SQLite 重启和 PostgreSQL 重新连接后仍读取相同活动状态、版本与审计字段。
9. 浏览器控制台、URL、存储和提交产物不包含凭据、用户输入输出或内部完整载荷。

## 停止线

- 不实现物理删除、批量归档 / 恢复、自动恢复、定时生命周期或保留清理。
- 不引入第三种生命周期、不建立跨 owner 的“重新启用事务”或集中式资源状态镜像。
- 不自动吊销 / 轮换 / 新建 API Key，不自动关闭 / 新建 Session，不自动改写 assignment。
- 不自动批准候选、消除漂移、重放 Run、补发请求或调用 Provider。
- 不扩生产应用目录、生产 membership / OIDC、共享角色、所有权转移、quota、billing 或发布声明。
- 不把开发测试态解除归档解释为生产授权恢复或生产事故恢复流程。

## 实施结果

- Memory、SQLite 与 PostgreSQL 已实现相同的 `archived -> active` CAS；并发单赢家、重启 / 重新连接、版本冲突和非法重复转换均有回归证据。
- HTTP 已注册严格 unarchive route，组合 scope / membership、活动工作区、显式影响确认和失败关闭均在业务写入前完成。
- Gateway 回归证明同一未吊销 Key 在归档时被拒绝、解除归档后按原资格恢复，已吊销 Key 不恢复。
- Web 已完成影响说明、最终确认、服务端记录替换和活动应用上下文交接；离线模式不伪造成功。
- SQLite 本地产品链的真实浏览器连续验证已完成：同一 Key 在活动态加载到 6 个模型，归档后得到 `403`，解除归档后再次加载到 6 个模型；测试 Key 随后已吊销。
- 定向 / 完整 Go、竞态、`go vet`、Web 测试与生产构建、真实 PostgreSQL 集成以及仓库快速 / 完整门禁均作为关闭证据。

## 关闭结论

批次 A 至 C 已完成，[唯一实施任务卡](../../task-cards/application-unarchive-safe-reactivation-dev-test-v1-plan.md)同步关闭。后续只有出现物理删除、批量生命周期、生产授权恢复或新的下游状态协调需求时，才重新更新对应功能设计；不从本专题派生同层 readiness、review、refresh 或 gate-only 任务。

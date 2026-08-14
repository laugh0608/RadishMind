# API 密钥引导式轮换与验证后退役（开发 / 测试态）v1

更新时间：2026-08-08

状态：`api_key_guided_rotation_verified_retirement_dev_test_v1_completed`

## 设计结论

开发测试态 API 密钥需要补齐一条显式、可恢复、不会提前中断调用的轮换路径。现有领域能力已经支持签发多枚 Key、一次性交接、Gateway 成功认证后的 `last_used_at` 和版本化吊销，但 Web 只把这些动作分散展示，无法引导用户证明替代凭据已被 Gateway 接受后再退役原凭据。

本专题把轮换固定为两阶段用户流程：

```text
原 Key active
  -> 签发同应用、同 scopes 的替代 Key，原 Key 保持 active
  -> 替代 Key 经现有调用入口通过 Gateway 认证，last_used_at 成为权威验证证据
  -> 精确重读原 Key，再以当前 record_version 吊销
```

不增加原子 `rotate` 端点。把“签发替代凭据”和“吊销原凭据”压成一次服务端动作，会在用户尚未取得或验证一次性令牌时提前切断原凭据，不符合本专题的可恢复性目标。

## 准入证据

[API 密钥生命周期与 Gateway 开发测试态认证 v1](api-key-lifecycle-gateway-dev-test-auth-v1.md) 已明确要求“需要替换时先签发新密钥，验证后再吊销旧密钥”，并已经提供所需的全部权威能力：

1. 替代 Key 可绑定同一活动应用并继承受控 scopes；
2. 原始令牌只在签发响应出现一次，Web 只在内存交接；
3. Gateway 成功认证后原子写入 `last_used_at`；
4. 吊销使用 `expected_version` CAS，且吊销后 Gateway 失败关闭；
5. Memory、SQLite、PostgreSQL 和真实浏览器链已经证明上述基础能力成立。

当前缺口在用户编排而非后端真相源：页面没有保存“正在替换哪枚 Key”的脱敏会话，没有强制 scopes 等价，也没有在退役旧 Key 前要求精确读取替代 Key 的成功认证证据。这是实际安全与可用性缺口，准入独立的引导式轮换专题。

## 用户目标

内部开发者可以在不中断现有调用的前提下完成一次开发测试态凭据轮换：

1. 从一枚活动 Key 发起轮换；
2. 查看并确认替代 Key 将保持同一应用与同一 scopes；
3. 选择新的展示名与 1 至 90 天有效期；
4. 取得只展示一次的替代令牌，并通过现有 Playground、Application RAG、Prompt Application 或外部调用完成验证；
5. 返回 API Key 页面，精确检查替代 Key 的 `last_used_at`；
6. 只有验证成立后才允许确认退役原 Key；
7. 在任一步失败、取消、刷新或离开时，原 Key 不被自动吊销。

## 真相源与状态模型

API Key repository 继续是唯一持久真相源。本专题不新增 rotation 表、relation 字段、生命周期状态或审计 owner。

Web 只保留不含凭据的易失轮换会话：

- `application_id`；
- 原 Key 的 `api_key_id`、开始轮换时的 `record_version`、展示名和 scopes；
- 替代 Key 的 `api_key_id`、`created_at`；
- 当前阶段：`replacement_pending`、`verification_pending` 或 `verified`；完成时直接清除会话。

令牌不得进入轮换会话。它继续只存在于现有一次性凭据状态或现有受控 handoff 通道。

持久权威状态由两条既有记录表达：

- 原 Key 是否仍为 `active`；
- 替代 Key 是否为 `active` 且 `last_used_at` 非空。

轮换会话不声称记录之间存在服务端关系；会话丢失时，两枚 Key 仍可通过现有列表和详情独立管理。

## 轮换规则

### 发起

- 只允许从 `lifecycle_state=active` 且 `effective_state=active` 的 Key 发起；
- 应用必须处于活动态；
- 同一 Web 运行实例只保留一个轮换会话，新会话必须显式取消旧会话；
- 发起时锁定原 Key 的应用与 scopes；
- 替代 Key scopes 必须与原 Key排序去重后完全一致，不允许在轮换中增加、删除或替换权限；
- 用户只可修改替代 Key 的展示名和有效期。

### 签发替代 Key

- 复用现有 `POST /v1/user-workspace/api-keys`；
- 签发失败时原 Key 不变，轮换停留在待签发状态；
- 签发成功后原 Key继续活动，替代令牌按现有一次性交接规则展示；
- 替代记录必须与会话的 tenant、workspace、application、owner 和 scopes 精确一致，否则失败关闭并不得进入验证阶段；
- 页面不得自动吊销原 Key，也不得自动调用 Provider。

### 验证替代 Key

- 用户通过现有受控入口或外部调用使用替代令牌；
- Web 返回后使用现有详情 API 精确读取替代 Key；
- 只有替代 Key 仍为活动态且 `last_used_at` 非空，才把会话推进到 `verified`；
- `last_used_at` 证明 Gateway 已接受凭据、应用状态与对应路由 scope；它不证明 Provider、模型或业务结果成功；
- 列表缓存、来源 Key 的使用时间或用户口头确认不能替代精确读取。

### 退役原 Key

- `verified` 后展示第二层确认，说明退役不可逆且替代 Key 必须妥善保存；
- 提交前精确重读原 Key；只有它仍为活动态才继续；
- 使用精确重读得到的 `record_version` 调用现有 revoke API；
- CAS 冲突、已吊销、存储失败或作用域漂移均不得伪造完成；
- 成功后清除一次性令牌和轮换会话，替代 Key 保持活动态。

## Web 生命周期与恢复

- API Key 页面内切换筛选、刷新列表或打开详情不丢失当前轮换会话；
- 跳转到现有调用入口时只保留脱敏轮换会话，令牌由已有一次性 handoff 单独承载；
- 返回 API Key 页面后可以重新检查替代 Key 使用证据；
- application 切换、显式取消、成功完成或启动另一轮轮换时清理旧会话；
- 浏览器刷新、页面重载或 Web 进程重启会丢失轮换会话，但不会修改任一 Key；
- 丢失会话后用户仍可通过既有详情确认 Key 状态并手动吊销，不提供令牌恢复。

## 失败语义

- 原 Key 非活动、应用归档或作用域不合法：不开始轮换；
- 替代签发失败：原 Key 保持活动；
- 替代记录 scope / owner / application 漂移：不接受签发结果，不退役原 Key；
- 替代 Key 尚无 `last_used_at`：保持 `verification_pending`；
- 替代 Key 已过期或已吊销：阻塞退役原 Key；
- 原 Key在确认前已改变：展示当前状态，轮换不得伪造完成；
- 详情、列表、签发或吊销响应不可信：沿用严格 consumer 失败关闭；
- 离线模式不发请求、不生成令牌、不模拟验证或轮换成功。

## 实施顺序

### 批次 A：易失状态模型与严格决策

- 新增独立的纯 TypeScript 轮换会话模型；
- 固定 active source、同应用、同 owner、同 scopes 和替代记录验证；
- 固定 `last_used_at` 验证门槛、显式取消、application 切换与完成清理；
- 增加状态模型精准单元测试，不修改后端 API 或 schema。

### 批次 B：Web 连续路径

- 在 API Key 列表为活动 Key 增加轮换入口；
- 建立替代展示名 / 有效期表单和只读 scopes 审查；
- 复用现有 issue、一次性交接、详情读取与 revoke consumer；
- 增加验证证据检查、退役二次确认、CAS 失败处理和返回路径；
- 完成 Web 全量测试、生产构建、SQLite 本地产品浏览器连续链与仓库门禁。

## 验收

1. 只有活动 Key 和活动应用可以发起轮换。
2. 替代 Key 与原 Key 必须绑定同一 application、owner 和完全相同的 scopes。
3. 替代签发后原 Key仍可调用，且不会因 UI 离开或验证失败自动吊销。
4. 替代令牌继续只展示一次，不进入轮换会话、URL、浏览器持久存储、日志或 committed 产物。
5. 替代 Key 未形成 `last_used_at` 时不能退役原 Key。
6. 替代 Key成功认证后，精确详情读取允许进入退役确认。
7. 退役前精确重读原 Key，并以最新版本调用既有 revoke CAS。
8. 完成后原 Key被拒绝、替代 Key继续通过 Gateway；测试结束后可显式吊销替代 Key。
9. application 切换、显式取消、刷新与失败路径不产生隐式凭据状态修改。
10. 离线模式零请求，Web 全量测试、生产构建、真实浏览器和仓库门禁通过。

## 完成结果与证据

批次 A、B 已于 2026-07-29 完成：

- 独立纯 TypeScript 状态模型只保存脱敏轮换元数据，覆盖 active source、会话互斥、同 application / owner / scopes 替代、`last_used_at` 验证门槛、原记录精确版本、application 切换、取消和完成清理；
- API Key 管理面已提供活动 Key 的 Rotate 入口、锁定 scopes、替代展示名 / 有效期、一次性交接、精确认证证据检查和不可逆退役二次确认；轮换期间普通签发和来源 / 替代 Key 的通用吊销入口被阻塞；
- 实现只编排既有 issue、read、revoke consumer 与 Playground / Application RAG / Prompt Application handoff，没有修改 API、schema、repository、permission 或 Gateway 认证语义；
- Web 全量 `270` 项测试和 production build 通过；
- SQLite 本地产品真实浏览器链验证：原 Key 已成功加载 `6` 个模型；替代 Key 签发后原 Key 继续加载 `6` 个模型；替代 Key 未认证时不存在退役入口；替代 Key 经 Playground 加载 `6` 个模型并形成精确 `last_used_at` 后才解锁退役确认；退役完成后原 Key 再次访问得到 `HTTP 403`，替代记录保持活动且保留认证证据；
- 浏览器验收产生的原 Key 与替代 Key 均已显式吊销，Web、platform 服务和浏览器页签已关闭；
- `git diff --check`、仓库快速门禁与完整门禁均通过；仅保留既有 W28、W29、W30 周志尺寸 warning。

2026-08-08 的 Family UI `S4 R1` 继续复用这套易失状态机，没有新增 rotation owner。Connect API、Credentials、Validate、Verify / retire 只形成页面任务位置；`verification_pending` 仍要求精确读取替代 Key 的 `last_used_at`，退役按钮在证据成立前保持禁用。状态 badge 不承担选中语义，当前任务和驱动详情的 Key 行才使用墨蓝细轨。

archived application 仍可读取、打开详情和吊销已有 Key，但不能开始或继续引导式轮换；issue、integration 和 invocation 同样关闭。workspace context 与 lifecycle 配置不一致、应用切换或旧列表尚未返回时都失败关闭并清空当前投影，不用旧 application 的记录填充新上下文。S4 浏览器连续链完成一次签发与两步吊销，原始令牌未进入文档、日志或持久浏览器介质；没有修改 API、schema、repository、permission 或 production 停止线。

## 停止线

- 不新增 rotate API、rotation repository、relation schema、第三种 Key 生命周期或后台轮换任务。
- 不实现自动轮换、定时轮换、宽限期执行器、批量轮换、自动吊销或自动重试。
- 不在轮换中增加 scope、延长原 Key 有效期、恢复已吊销 Key或再次显示原始令牌。
- 不把 `last_used_at` 解释为 Provider、模型输出或业务执行成功。
- 不持久化轮换会话，不写 `localStorage`、`sessionStorage`、cookie 或 URL。
- 不扩 production API key、生产 membership / OIDC、quota、rate limit、billing、provider secret 或公开生产 Gateway。
- 不新增同层 checker；复用既有 Web tests / build、Gateway 回归、SQLite 本地产品链和仓库聚合门禁。

## 专题关闭

[唯一实施任务卡](../../task-cards/api-key-guided-rotation-verified-retirement-dev-test-v1-plan.md)已完成并关闭。本专题不继续派生自动轮换、持久 rotation owner、生产凭据或同层 gate-only 批次；下一轮回到当前焦点，依据真实使用证据选择新的功能设计。

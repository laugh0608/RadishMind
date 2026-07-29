# API 密钥引导式轮换与验证后退役（开发 / 测试态）v1 实施任务卡

更新时间：2026-07-29

状态：`api_key_guided_rotation_verified_retirement_dev_test_v1_completed`

## 任务目标

按照[API 密钥引导式轮换与验证后退役（开发 / 测试态）v1](../features/user-workspace/api-key-guided-rotation-verified-retirement-dev-test-v1.md)，在现有 API Key Web 管理面建立“签发同权限替代 Key → 观察成功认证证据 → 精确退役原 Key”的引导路径，避免提前中断和长期双凭据遗留。

## 准入与边界

- 现有 issue、一次性交接、Gateway 认证、`last_used_at` 与 revoke CAS 已成立，本任务不重复建立后端真相源。
- 轮换属于凭据安全路径，使用唯一任务卡固定状态与失败边界。
- 不新增 API、schema、migration、repository、permission、Gateway route 或后台任务。
- production API key、quota、billing、生产 OIDC / membership 与 provider secret 保持关闭。

## 批次 A：状态模型与决策边界

实现：

1. 新增不含令牌的易失轮换会话类型。
2. 只允许 active source；固定 application、owner、source version 和排序 scopes。
3. 替代记录必须是不同 Key、同 application / owner / scopes 且保持活动态。
4. 只有替代记录 `last_used_at` 非空才进入 `verified`。
5. application 切换、取消、完成和非法状态清理会话；浏览器刷新自然丢失。
6. 以纯单元测试覆盖正常顺序、scope 漂移、owner 漂移、未验证、替代失效和会话清理。

批次 A 完成条件：

- 状态模型不包含 token、digest、Authorization 或完整响应；
- 任一不可信记录不能推进阶段；
- 不修改 API Key repository 或 Gateway 认证语义。

## 批次 B：管理界面与连续验收

实现：

1. 活动 Key 行增加 Rotate 入口，非活动 Key 不展示可执行轮换。
2. 表单锁定原 scopes，只允许替代展示名和有效期。
3. 复用现有 issue consumer，签发后保留原 Key活动并展示一次性替代令牌。
4. 复用现有 Playground、Application RAG、Prompt Application 和复制路径完成用户验证。
5. 返回后精确读取替代 Key，按 `last_used_at` 推进验证。
6. 退役原 Key前再次精确读取原记录，二次确认后使用最新版本 revoke。
7. 处理冲突、失效、取消、迟到响应、application 切换和离线零请求。
8. 完成 Web 测试 / build、SQLite 本地产品真实浏览器链、敏感信息复验和仓库门禁。

批次 B 完成条件：

- 替代凭据未验证时不存在退役原 Key的可执行路径；
- 签发、验证或离开失败不会隐式吊销原 Key；
- 完成后原 Key拒绝、替代 Key可用；
- 本批测试 Key、服务、浏览器页面和本地进程在提交前清理。

## 验证矩阵

| 层级 | 验证 |
| --- | --- |
| 状态模型 | active source、严格 scope equality、replacement identity、`last_used_at`、cancel / complete / application switch |
| Web consumer | 继续复用 strict issue / read / revoke envelope、一次性 token 与 `no-store` |
| Web UI | 锁定 scopes、替代签发、验证门槛、退役二次确认、失败保留 |
| 产品链 | source 可用 → issue replacement → source 仍可用 → replacement 使用 → 精确验证 → source revoke → source `403` / replacement 可用 |
| 隐私 | token 不进入 session、URL、持久存储、日志、测试快照或提交产物 |
| 仓库 | Web 全量 tests / build、`git diff --check`、fast 与 full `check-repo` |

## 完成记录

- 批次 A、B 已完成，状态模型、Web 引导路径、精确认证证据与退役 CAS 已形成连续实现。
- Web 全量 `270` 项测试和 production build 通过。
- SQLite 本地产品浏览器链已验证：source 成功调用；replacement 签发后 source 继续调用；replacement 未认证时退役入口关闭；replacement 成功加载 `6` 个模型并形成 `last_used_at` 后解锁双重确认；source 退役后访问返回 `HTTP 403`，replacement 记录继续保持活动。
- 验收产生的 source / replacement 均已吊销，本批服务和浏览器页签均已关闭。
- `git diff --check`、仓库快速与完整门禁均通过；仅保留既有 W28、W29、W30 周志尺寸 warning。本任务卡关闭，不派生同层任务。

## 停止线

- 不创建后端 rotate / verify route 或 rotation owner。
- 不写 token 到轮换会话或持久浏览器状态。
- 不自动调用 Provider、不自动退役、不自动扩大 scopes。
- 不把测试态轮换解释为生产凭据轮换、零停机 SLA 或集中式 secret rotation。
- 不派生 readiness、review、refresh、fixture 或 checker 链。

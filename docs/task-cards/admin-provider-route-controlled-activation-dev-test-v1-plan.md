# Admin Provider Profile / Model Route 受控启用（开发 / 测试态）v1 实施任务卡

更新时间：2026-07-26

状态：`admin_provider_route_controlled_activation_dev_test_v1_batch_a_completed_batch_b_ready`

对应功能文档：[Provider Profile / Model Route 配置草案、版本审查与受控启用（开发 / 测试态）v1](../features/admin-control-plane/provider-profile-model-route-controlled-activation-dev-test-v1.md)

## 任务目标

在一个连续开发阶段内，把 Admin Provider / Route evidence 只读能力推进为开发测试态受控配置产品流程，并让 Gateway 最终只消费经过草案校验、人工审查和显式启用的不可变运行快照。

本任务卡统一承载领域、持久化、API、Gateway 消费和 Admin Web 五个批次。除非出现新的生产边界或外部系统依赖，不为每个批次继续派生同层任务卡。

## 根因

- Admin 当前只提供 readiness / evidence review，`canMutateProviderProfile` 与 `canChangeModelRoute` 均保持 false。
- Gateway 当前使用运行配置中的 provider/profile/model，缺少 Admin 持有的可审查版本和原子激活事实。
- Workflow、Gateway、用户工作区、Prompt / Agent 和 Image Adapter 当前专题已有明确停止线，继续增加同层切片不符合产品推进顺序。
- quota / billing 缺少可信 usage、价格和 ledger owner，真实 OIDC 缺少 Radish upstream evidence，两者都不具备当前实施条件。

## 固定架构

```text
Admin draft
  -> immutable candidate
  -> manual review
  -> explicit activation with expected generation
  -> immutable ProviderRouteSnapshot
  -> Gateway request pins generation/digest
```

- 草案、候选、审查、激活事实归 Admin。
- provider runtime inventory 继续归既有配置 / bridge owner。
- Gateway 只消费 snapshot，不读取 Admin draft / review repository。
- 审查不会自动激活，激活不会修改候选。

## 批次 A：领域与 memory_dev

### 允许

- 新增 Admin Provider Route 领域、repository interface、service 和内存实现。
- 引入只读 inventory resolver 接口。
- 实现 draft CAS、candidate immutable create、review CAS、activation generation CAS、rollback 和 history。
- 新增相邻 Go 单元测试和并发测试。
- 更新功能入口、当前焦点、Admin / Gateway 专题、路线图、能力矩阵和周志。

### 禁止

- 不新增 HTTP route、配置开关或 Web。
- 不修改现有 Gateway request selection。
- 不创建 SQLite / PostgreSQL 表、migration 或 selector。
- 不读取 credential、endpoint、环境变量或 provider raw config。
- 不启用 production、自动 activation、fallback、quota 或 billing。

### 完成条件

- 规范化、digest 与 inventory binding 可确定性重算。
- draft、review、generation 三类 CAS 语义稳定。
- approval 对 active snapshot 零影响。
- activate / rollback 形成 append-only history，rollback 使用新 generation。
- 敏感材料、跨环境 ref、重复 route、capability mismatch、inventory 漂移均失败关闭。
- 并发更新 / 审查 / 激活只有一个成功，不产生部分记录。
- 相邻测试、`go test`、`go test -race` 和 fast 仓库门禁通过。

### 完成记录

- 新增 `admin_provider_route_configuration.go`、`admin_provider_route_configuration_memory.go` 与独立完整性重验模块，领域、存储和完整性职责没有堆入现有 Gateway 或 control-plane read 文件。
- 草案、候选、review、snapshot 和 activation history 已固定 schema、状态、digest、三类 CAS 与深复制边界；approval 不改变 active snapshot，rollback 只接受历史已启用候选并生成新 generation。
- inventory resolver 只返回脱敏 binding；candidate 与 activation 分别解析并比较 digest，inventory not found / unavailable / mismatch 均无激活副作用。
- 11 项测试覆盖成功流程、规范化、权限、隔离、敏感输入、负向状态、repository 故障、存储漂移与并发；定向普通测试和 race 测试均通过。
- 批次 A 不包含 HTTP、配置开关、SQLite、PostgreSQL、Gateway selection 或 Web；下一步只进入批次 B durable dev/test repositories。

## 批次 B：durable dev/test repositories

### 工作项

- 实现共享 SQLite component migration 与 repository。
- 实现 PostgreSQL schema、marker、manual migration、运行 / 迁移角色和 repository。
- 接入互斥 `memory_dev | sqlite_dev | postgres_dev_test` selector。
- 验证真实文件、真实 PostgreSQL、多连接池、重启恢复、并发 CAS 和 no fallback。

### 完成条件

- 三模式 repository contract 行为一致。
- 草案、候选、review、active snapshot 和 activation history 全部可恢复。
- activation 与 history 在同一事务中提交。
- DB / WAL / 日志不包含 forbidden material。

## 批次 C：Admin HTTP / Auth

### 工作项

- 增加四项独立权限和 strict management payload。
- 暴露草案、候选、审查、激活、回滚、当前快照和历史 API。
- 接入 verified identity、request / audit lineage 和稳定 HTTP failure mapping。
- 增加启动配置与显式开发测试态写入门禁。

### 完成条件

- 权限不足、OIDC membership unavailable 和未知作用域在 repository 前失败。
- stale revision / review / generation 返回 409 与脱敏当前版本。
- 错误和响应不包含 inventory raw payload 或敏感材料。

## 批次 D：Gateway snapshot consumer

### 工作项

- 建立 Gateway 只读 snapshot provider。
- 按 protocol + model 精确匹配 route，取出 profile assignment。
- 每次请求固定 generation / digest，并进入脱敏 Request History。
- 增加 static config / admin snapshot 互斥选择，不按请求 fallback。

### 完成条件

- 激活前后请求分别固定旧 / 新 generation。
- 在途请求不因后续激活发生 provider 漂移。
- route missing、profile missing、inventory drift 和 snapshot unavailable 均在 bridge 前失败。
- 现有 northbound contract 和静态配置回滚模式保持兼容。

## 批次 E：Admin Web 与产品链

### 工作项

- 把既有 Provider Deployment Review 扩展为受控配置工作区。
- 提供草案编辑、校验、版本差异、审查、激活、回滚、当前快照和历史。
- 完成 SQLite 本地产品 launcher、PostgreSQL 专项门禁和浏览器连续链。

### 完成条件

- 浏览器可完成 draft → candidate → approve → activate → Gateway invocation → history → rollback。
- 服务重启后配置与历史恢复。
- 原始敏感材料不进入页面、URL、浏览器存储、日志或数据库。
- Web 单元测试、build、产品 smoke、fast 和 full 仓库检查通过。

## 验证策略

- 每个批次先跑相邻单元 / 集成测试，再跑平台 Go 回归。
- 并发与快照切换批次必须运行 `go test -race`。
- SQLite / PostgreSQL 分别承担本地产品连续性和数据库事务 / 角色证据，不能互相替代。
- Web 批次运行严格 consumer 测试、完整 Web 测试、生产 build 和真实浏览器链。
- 批次 A、持久化完成、Gateway 接入和最终关闭时运行快速仓库检查；最终关闭运行全量仓库检查。

## 总停止线

- 不创建第二套 provider registry、selection policy 或 northbound contract。
- 不保存或解析真实 credential / endpoint。
- 不启用 production、自动 activation、fallback、load balancing、quota、billing 或部署执行。
- 不接真实 Radish OIDC，不绕过 membership blocker。
- 不从本任务扩 Workflow executor、Agent loop、Image backend 或业务写回。

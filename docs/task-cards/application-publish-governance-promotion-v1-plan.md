# Application Publish Governance & Promotion v1 实施任务卡

更新时间：2026-07-12

状态：`application_publish_governance_promotion_v1_complete`

## 任务目标

按照 [Application Publish Governance & Promotion v1](../features/user-workspace/application-publish-governance-promotion-v1.md) 实现 application-scoped、dev/test only 的不可变 publish candidate 与审查链，复用现有 application draft、Control Plane read baseline、API Integration、Gateway Playground 和 Request History。

## 前置条件

- Application Configuration Draft & Review v1 已完成 saved valid draft、版本、CAS、memory / PostgreSQL dev-test repository。
- Application API Integration、Gateway Playground 与 Request History 已具备 application / protocol / model / request-id handoff。
- 当前成熟度为内部开发者预览；正式 application repository、production auth / membership、发布 owner 和 promotion runtime 未成立。

## 实现范围

1. 新增 `application_publish_candidate.v1`、review record、promotion eligibility 和稳定 failure。
2. create 必须由服务端读取精确 saved draft，计算 sanitized digest，并绑定当前 application baseline。
3. candidate configuration / draft binding / evidence refs 创建后不可变；review 使用 append-only record 和 expected review version。
4. 新增 scoped dev-only create / list / read / review routes，分别要求 read / write / review scope。
5. 新增 memory 与独立 PostgreSQL dev/test repository、schema marker、manual migration runner、store selector 和 no-fallback。
6. 新增 lazy Publish Review Workspace，支持 candidate 创建、列表、恢复、配置比较、review、eligibility、Integration / Playground / History handoff 和 application 切换隔离。
7. offline 模式零请求，所有未提交输入只保存在组件内存。

## 高风险边界

- 这是新增 API、schema、写入和审批边界，必须补 Go / Web 单测、PostgreSQL 集成、真实浏览器与完整仓库门禁。
- dev/test review scope 不代表生产 reviewer 身份、职责分离或授权。
- candidate approved 不调用任何 application mutation；eligibility 必须明确返回 production blocker。
- 不复制 application draft repository、Gateway history store、协议 adapter、SSE parser 或请求状态模型。

## 验收证据

- Go：domain、scope、server-side draft reload、digest、immutable conflict、review CAS / transition、baseline drift、store failure。
- PostgreSQL：migration apply / rollback / reapply、重启恢复、immutable create、review CAS、scope、runtime role DDL denial、no fallback。
- Web：offline zero fetch、strict response validation、secret scan、scope handoff、candidate / review conflict、application reset、Integration / Playground / History events。
- 浏览器：saved draft → candidate → compare → approve / conflict → promotion blocked → exact history detail；console 和 storage 无泄漏。
- 门禁：Web test / build、Go test / race / vet、`git diff --check`、fast 与完整 `check-repo`。

## 完成条件

- 用户可在当前 application 下创建并审查不可变 candidate，所有版本和漂移均可解释、可复验。
- dev/test repository 支持重启恢复和并发冲突，失败不回退。
- 正式 application mutation、production auth、API key / quota / billing 与 Workflow 高风险能力保持关闭。

## 完成结论

上述条件已于 2026-07-12 完成。Platform、PostgreSQL dev/test、Web consumer / workspace、真实浏览器多标签冲突和精确 History 交接均已验证；approved candidate 仍由正式 repository、auth、owner 与 promotion runtime blocker 阻止发布。

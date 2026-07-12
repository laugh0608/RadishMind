# User Workspace 细专题入口

更新时间：2026-07-12

本目录承接 `User Workspace` 中跨 Applications、模型发现、接入、调用与审查的具体功能专题。产品面长期边界继续以 [User Workspace 设计与开发文档](../user-workspace.md) 为准。

## 当前专题

- [Application API Integration & Invocation v1](application-api-integration-invocation-v1.md)：把选中 application、`/v1/models` 模型目录、三协议接入示例、现有 Gateway Playground 调用与 sanitized Request History 审查串成连续的内部开发者路径。
- [Application Configuration Draft & Review v1](application-configuration-draft-review-v1.md)：为当前 application 建立独立配置草案、校验、开发测试态持久化、版本冲突、比较和 API Integration handoff。
- [Application Publish Governance & Promotion v1](application-publish-governance-promotion-v1.md)：已完成不可变 publish candidate、版本绑定、review CAS、漂移识别、阻塞式 promotion eligibility 和既有 Integration / Playground / History 交接；不直接发布正式 application。

## 下一专题

- `Application Catalog & Lifecycle Dev/Test v1`：先形成正式功能设计，再实现 application 创建、列表、详情、受控元数据更新、归档、scope / owner / CAS、PostgreSQL dev/test persistence，以及到 Configuration Draft、Integration 和 Publish Review 的连续路径。OIDC 模式在 workspace membership 未成立时继续 fail closed，且不把开发测试态 catalog 解释为 production application repository。

## 目录停止线

- application draft 只允许建立独立 dev/test repository，不改变 northbound Gateway schema、provider registry 或正式 application 真相源。
- 不在内部开发者预览阶段打开 production API key、quota、billing、自动 fallback、load balancing 或 production auth。
- application 创建、发布、删除和业务写回必须作为后续独立功能设计，不并入当前接入与调用工作区。

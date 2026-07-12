# User Workspace 细专题入口

更新时间：2026-07-12

本目录承接 `User Workspace` 中跨 Applications、模型发现、接入、调用与审查的具体功能专题。产品面长期边界继续以 [User Workspace 设计与开发文档](../user-workspace.md) 为准。

## 当前专题

- [Application API Integration & Invocation v1](application-api-integration-invocation-v1.md)：把选中 application、`/v1/models` 模型目录、三协议接入示例、现有 Gateway Playground 调用与 sanitized Request History 审查串成连续的内部开发者路径。
- [Application Configuration Draft & Review v1](application-configuration-draft-review-v1.md)：为当前 application 建立独立配置草案、校验、开发测试态持久化、版本冲突、比较和 API Integration handoff。

## 目录停止线

- application draft 只允许建立独立 dev/test repository，不改变 northbound Gateway schema、provider registry 或正式 application 真相源。
- 不在内部开发者预览阶段打开 production API key、quota、billing、自动 fallback、load balancing 或 production auth。
- application 创建、发布、删除和业务写回必须作为后续独立功能设计，不并入当前接入与调用工作区。

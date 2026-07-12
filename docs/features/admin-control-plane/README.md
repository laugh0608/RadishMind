# Admin Control Plane 细专题入口

更新时间：2026-07-12

本目录承接 `Admin Control Plane` 中需要跨身份、权限、repository 和管理端使用路径推进的功能专题。产品面长期边界继续以 [Admin Control Plane 设计与开发文档](../admin-control-plane.md) 为准。

## 当前专题

- [Authenticated Read Store Transition v1](authenticated-read-store-transition-v1.md)：第一批 verified identity / negative auth runtime 已完成；下一批只设计并实现 Admin tenant / audit PostgreSQL dev/test read repository。

## 目录停止线

- Radish 继续拥有用户、tenant、role、permission 与 OIDC 真相；RadishMind 不复制身份系统。
- Admin read transition 不并入管理写入、application promotion、API key lifecycle、quota enforcement、billing、secret runtime 或部署执行。
- 每个实现批次只打开一个主要高风险边界；auth、membership、store 与真实 Radish 联调按顺序验收，不同时切换。

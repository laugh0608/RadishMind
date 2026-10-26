# RadishMind 扩展与集成专题入口

更新时间：2026-10-26

## 文档职责

本目录承接外部项目、模型提供方与图片后端的真实接入，记录目标、外部条件、协议和验收证据。当前先交付[本机持久工作台](../features/user-workspace/local-persistent-workbench-delivery-v1.md)；真实模型试用继续暂停，外部接入不作为本机技术交付的前置条件。

外部项目使用在线仓库与项目名，不把临时本机路径写入长期文档。上游产品已有进展不等于本项目真实挂载点、客户端注册和验收环境已齐备。

## 当前事实与进入条件

| 方向 | 当前事实 | 进入条件 |
| --- | --- | --- |
| RadishFlow | 已有协议、样例和离线建议交接证据，真实业务接入未验收 | 明确一个真实 UI / command / API 挂载点、负责人、确认流和验收环境，再设计接入 |
| Radish OIDC | [本地账户与联合登录](../features/admin-control-plane/local-account-radish-oidc-federated-login-v1.md)已完成本地身份、Session、成员授权及确定性 OIDC；真实联调为 `real_radish_integration_deferred` | 真实签发方、客户端注册、密钥、部署资源与负责人齐备后恢复；不重复创建已完成的身份准入链 |
| 真实模型 | 已有多提供方适配与诊断试用素材；当前暂停试用 | 明确提供方 / 模型、任务数据、费用与操作窗口；分别验证技术链与答案质量 |
| 图片后端 | [图片专题](../features/image-generation-artifact-return.md)已完成固定响应客户端、adapter、一次性二进制交付与私有存储协调 | 真实后端、引用解析、生产存储及 HTTP / Gateway / Web 交付尚未完成；按具体图片任务选择范围 |
| Radish 文档 / 业务 | 已有文档问答资产，真实业务产品接入未验收 | 明确任务与业务入口；联合身份和业务 Copilot 分别确认 |
| RadishCatalyst | 文档级预留 | 明确任务前不扩 schema、adapter 或运行接线 |

## 历史身份证据

[Radish OIDC Token / Membership Readiness v1](radish-oidc-token-membership-readiness-v1.md) 已固定 `radish_oidc_token_membership_readiness_defined`；[Radish OIDC Token / Membership Implementation Entry Review v1](radish-oidc-token-membership-implementation-entry-review-v1.md) 已固定 `radish_oidc_token_membership_implementation_entry_review_defined`；[Radish OIDC Token / Membership Upstream Evidence Refresh v1](radish-oidc-token-membership-upstream-evidence-refresh-v1.md) 已固定 `radish_oidc_token_membership_upstream_evidence_refresh_defined`。这些是早期静态准备证据，后续本地账户与确定性 OIDC 实现以现有功能专题为准；不再要求重做 token / middleware / membership 的同层入口评审。

## 停止线

- 不把固定响应、离线样例、本地回环签发方或开发身份头当成真实接入与生产授权证据。
- 不假定外部项目就绪，也不以外部依赖缺失无限延期本机交付。
- 不跨工作区编辑 `RadishFlow`、`Radish` 或 `RadishCatalyst`；真实外部操作按当次范围授权。
- 不让模型输出直接写业务真相，不以联合登录隐式授予本地权限。

# RadishMind 扩展与集成专题入口

更新时间：2026-06-15

## 文档目的

本目录用于承接外部项目、外部 backend 和真实 provider 接入专题。它记录接入目标、前置条件、依赖方、停止线和可验收证据，不直接替代产品面功能文档或实现任务卡。

外部项目默认使用在线仓库 URL 和项目名，不写开发者本机路径。

## 何时放在这里

- 需要依赖 `RadishFlow`、`Radish`、`RadishCatalyst` 的真实挂载点、命令承接接口或确认流。
- 需要接入 Radish OIDC、image backend、外部 provider live health、真实 secret backend 等外部能力。
- 需要区分本仓库可先做的离线 / dev-only 能力和外部系统就绪后才能做的真实接入。

## 当前集成专题候选

| 专题 | 当前状态 | 下一步 |
| --- | --- | --- |
| RadishFlow Integration | gateway / handoff 门禁冻结，真实挂载点未成熟 | 等上层提供稳定 UI、command 或 API 承接点后再重开 |
| Radish OIDC | 前置条件已固定；[Radish OIDC Token / Membership Readiness v1](radish-oidc-token-membership-readiness-v1.md) 已固定 `radish_oidc_token_membership_readiness_defined` | 后续若推进，应做 token / membership implementation entry review；不直接创建 middleware、validator、membership adapter 或 production API |
| Image Backend Adapter | metadata-only response builder 已完成，真实 backend 未接 | 后续只能在 store、reader、public URL 或 backend adapter 中选择一个方向独立推进 |
| Radish Docs / Knowledge Integration | docs QA 资产已有，真实产品接入仍等待 | 不把文档问答资产写成完整上层接入 ready |
| RadishCatalyst | 文档级预留 | 未明确任务面前不扩 schema、adapter、gateway smoke 或模型接线 |

## 停止线

- 不把外部系统未就绪写成本仓库平台本体停滞理由。
- 不继续细化假想接线；真实接入必须有上层挂载点、确认流和命令承接接口。
- 不跨工作区编辑 `RadishFlow`、`Radish` 或 `RadishCatalyst` 本地项目。
- 不把 metadata-only、offline sample、dev-only route 或 fixture smoke 写成真实集成 ready。

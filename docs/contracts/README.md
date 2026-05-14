# RadishMind 契约专题目录

更新时间：2026-05-14

本目录承载 `docs/radishmind-integration-contracts.md` 拆出的稳定专题。入口文档只保留当前结论、索引和停止线；需要修改字段、任务上下文或长样例时，优先修改本目录下的对应专题页。

## 阅读入口

- [服务/API 接入契约](service-api.md)
- [会话记录契约](session.md)
- [工具框架契约](tooling.md)
- [训练 / 蒸馏样本契约](training-samples.md)
- [图片生成契约](image-generation.md)
- [输入与项目上下文契约](input-context.md)
- [RadishFlow 上下文契约](radishflow-context.md)
- [Radish 上下文契约](radish-context.md)
- [RadishCatalyst 上下文预留契约](radishcatalyst-context.md)
- [输出与候选动作契约](output-actions.md)

## 维护规则

- 本目录文档不替代 `contracts/` 下的 schema 真相源。
- `P2 Session & Tooling Foundation` 的晋级口径同时由 `scripts/checks/fixtures/session-tooling-promotion-gates.json` 固定；修改 session/tooling promotion gate 时，应同步更新对应专题页和该 fixture。
- 长示例、批次流水和运行记录继续进入 fixture、manifest、summary、run record 或 task card 附件。
- 单个专题接近 `500` 行时，应优先继续按稳定职责拆分，而不是添加 `markdown-size-allow:`。

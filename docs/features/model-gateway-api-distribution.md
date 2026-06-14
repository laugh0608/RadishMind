# Model Gateway / API Distribution 设计与开发文档

更新时间：2026-06-14

## 功能定位

`Model Gateway / API Distribution` 负责对外提供 OpenAI-compatible、Responses、Messages、Models 等 northbound API，并统一分发到多 provider、多 profile 和多模型。

## 当前状态

- 平台已有 `/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models` 和 `/v1/models/{id}` 的第一版 bridge-backed 兼容面。
- `apps/radishmind-web/` 已有 Model Gateway Overview、Route Evidence、Usage/Audit Evidence 和 Evidence Review / Readiness。
- provider capability、health smoke、selection policy、retry/fallback policy 和 runtime docs 已进入 fast baseline。
- 当前不执行真实 API key lifecycle、quota enforcement、rate limit、billing、cost ledger、provider retry/fallback execution、production gateway 或 load balancing。

## 设计边界

- gateway 只按 canonical contract 与 provider/profile metadata 分发，不把任一 provider 写成唯一方向。
- capability 不等于 health，health smoke 不等于 production readiness。
- 默认 retry policy 为 caller-managed，fallback policy 为 disabled；任何自动 fallback 都需要独立设计和审计。
- key、quota、billing 和 cost ledger 必须有明确失败语义和审计记录，不能只做 UI 展示。

## 下一批开发方向

1. 真实 API distribution 前，先更新本功能文档，明确 northbound compatibility surface、auth boundary、quota boundary 和 trace model。
2. 若只调整 evidence 展示，复用现有 web build 与 fast baseline。
3. 若新增 key validation、quota check、rate limit、cost ledger 或 provider fallback，必须新增 task card、负向测试和专项 checker。
4. production gateway 与 secret backend 只能在明确部署 / secret 方案后推进，不与普通 UI review 混在同一批。

## 验收方式

- compatibility surface：Go tests、gateway smoke、contract smoke。
- provider selection：selection unit tests、diagnostics checks、fast baseline。
- key / quota / billing：schema、negative tests、audit checks、no secret leak checks 和全量仓库验证。

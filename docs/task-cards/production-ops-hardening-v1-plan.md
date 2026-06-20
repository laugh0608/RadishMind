# `Production Ops Hardening` v1 计划

更新时间：2026-05-26

## 任务目标

本任务卡用于把 `P3 Local Product Shell / Ops Surface` 从 `local usable / read-only close` 推进到可部署前置条件更清楚的 production ops hardening 轨道。

当前任务不接真实 executor、不做 confirmation / writeback / replay、不声明 production ready，也不重新打开真实模型长跑。它只处理平台运行、配置、密钥、环境隔离和 console packaging 的可验证边界。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [阶段路线图](../radishmind-roadmap.md)
- [系统架构](../radishmind-architecture.md)
- [集成契约](../radishmind-integration-contracts.md)
- `scripts/checks/fixtures/p3-local-product-shell-short-close-checklist.json`
- `scripts/check-p3-local-product-shell-short-close-checklist.py`
- `apps/radishmind-console/`
- `services/platform/`

## 当前事实

- `P1 Runtime Foundation` 已达到 short close。
- `P2 Session & Tooling Foundation` 仍为 `close candidate / governance-only`，真实 executor、durable store、confirmation 接线、materialized result reader、长期记忆和 replay 仍 blocked。
- `P3 Local Product Shell / Ops Surface` 已达到 `local usable / read-only close`，但 production secret backend、process supervisor、deployment environment isolation 和 console production packaging 仍为 `not_satisfied`。
- `P4 Model Adaptation` 已形成前置证据：`Qwen2.5-1.5B-Instruct` raw blocked、repaired comparison 只作后处理证据、`Qwen2.5-3B-Instruct` CPU 单样本 timeout。真实模型产出、3B/4B 长跑、训练 JSONL、蒸馏和权重相关工作转入后置专题。

## v1 范围

1. `production config / secret boundary`
   - 明确 local / dev / production 配置来源。
   - 明确密钥、provider credential、profile 覆盖和不可提交项。
   - 不实现真实 secret backend，但必须把当前缺口固定为可检查事实。
2. `process supervisor / startup`
   - 明确平台服务和 console 的启动、退出码、健康检查、日志路径和人工重启边界。
   - 不实现长期守护进程或自动恢复器。
3. `deployment environment isolation`
   - 区分 local readiness、dev smoke 和 production readiness。
   - 防止 `mock` / local-smoke / demo profile 被误读为 production ready。
4. `console production packaging`
   - 为 `apps/radishmind-console/` 固定 production build / package smoke 的最低可验证入口。
   - 不新增 API、不接 executor、不做 confirmation、writeback 或 replay。
5. `P3 checklist alignment`
   - 将现有 `not_satisfied` 项拆成可推进的 evidence item。
   - 保持普通 UI 展示改动复用现有 behavior / visual smoke / fast baseline。

## 非目标

- 不重启真实模型产出专题。
- 不运行本地大模型长跑。
- 不下载模型权重或数据集。
- 不生成训练 JSONL、蒸馏数据或 checkpoint。
- 不把 `repaired`、guided、builder 或后处理结果写成 raw 模型晋级。
- 不提前接真实上层项目写回。
- 不把 console 壳声明为 production console。

## 建议切片

1. `config-secret-boundary`
   - 梳理现有 config 文档、fixture 和平台检查。
   - 新增或更新最小门禁，固定 production secret backend 仍未实现但边界清楚。
   - 当前已落地 governance boundary：`scripts/checks/fixtures/production-ops-config-secret-boundary.json` 与 `scripts/check-production-ops-config-secret-boundary.py` 固定 local / dev / production 配置来源、密钥注入、provider/profile 边界和生产未就绪声明；这不等于 production secret backend ready。
2. `production-secret-backend-contract`
   - 在不实现真实云 secret 服务、不写入真实 secret、不声明 production ready 的前提下，固定 future external secret backend adapter contract。
   - 明确 secret reference、身份字段、脱敏输出、禁止输出和仍 blocked 的 production 条件。
   - 当前已落地 governance boundary：`scripts/checks/fixtures/production-ops-secret-backend-contract.json` 与 `scripts/check-production-ops-secret-backend-contract.py` 固定 `environment`、`provider`、`provider_profile`、`secret_ref`、`RADISHMIND_SECRET_SOURCE`、脱敏字段和禁止项；production secret backend 仍为 `not_satisfied`。
3. `production-secret-backend-implementation-readiness`
   - 新增 [Production Secret Backend Implementation v1 计划](production-secret-backend-implementation-v1-plan.md)，把真实实现前必须满足的 schema、注入点、profile binding、脱敏审计、failure taxonomy、测试策略、operator runbook 和 rotation / audit policy 写清。
   - 当前已落地 readiness gate：`scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json` 与 `scripts/check-production-ops-secret-backend-implementation-readiness.py` 固定前置条件；它不实现 resolver、不接真实云 secret 服务、不写入真实 secret、不声明 production ready。
4. `config-secret-ref-readiness`
   - 新增 [Production Secret Backend Config / Secret Ref Readiness v1](production-secret-backend-config-secret-ref-readiness-v1-plan.md)，把配置注入点、reference-only manifest 消费边界、脱敏字段、failure mapping、no fallback、no side effects 和 artifact guard 固定为可检查证据。
   - 当前已落地 readiness gate：`scripts/checks/fixtures/production-secret-backend-config-secret-ref-readiness-v1.json` 与 `scripts/check-production-ops-secret-backend-config-secret-ref-readiness-v1.py` 固定 `production-secret-backend-config-secret-ref-readiness-v1`；它不实现 resolver、不创建 fake resolver、不调用云 secret 服务、不读取 secret value、不接数据库、不启用 repository mode。
5. `provider-profile-secret-binding`
   - 新增 [Production Secret Backend Provider Profile Secret Binding Readiness v1](production-secret-backend-provider-profile-secret-binding-readiness-v1-plan.md)，把 provider/profile 的 `credential_requirement`、`secret_ref_status`、环境绑定、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard 固定为可检查证据。
   - 当前已落地 readiness gate：`scripts/checks/fixtures/production-secret-backend-provider-profile-secret-binding-readiness-v1.json` 与 `scripts/check-production-ops-secret-backend-provider-profile-secret-binding-readiness-v1.py` 固定 `production-secret-backend-provider-profile-secret-binding-readiness-v1`；它不实现 resolver、不创建 fake resolver、不调用云 secret 服务、不读取 secret value、不接数据库、不启用 repository mode。
6. `secret-resolver-interface-disabled`
   - 新增 [Production Secret Backend Secret Resolver Interface Disabled Readiness v1](production-secret-backend-secret-resolver-interface-disabled-readiness-v1-plan.md)，把 future resolver interface 的 reference-only 输入、disabled result、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard 固定为可检查证据。
   - 当前已落地 readiness gate：`scripts/checks/fixtures/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.json` 与 `scripts/check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py` 固定 `production-secret-backend-secret-resolver-interface-disabled-readiness-v1`；它不实现 resolver runtime、不创建 fake resolver、不调用云 secret 服务、不读取 secret value、不接数据库、不启用 repository mode。
7. `operator-runbook-and-negative-gates`
   - 新增 [Production Secret Backend Operator Runbook / Negative Gates Readiness v1](production-secret-backend-operator-runbook-negative-gates-readiness-v1-plan.md)，把 operator runbook、test / production secret source、operator approval evidence、sanitized verification、smoke record reference、negative gates、failure mapping、no fallback、no side effects 和 artifact guard 固定为可检查证据。
   - 当前已落地 readiness gate：`scripts/checks/fixtures/production-secret-backend-operator-runbook-negative-gates-readiness-v1.json` 与 `scripts/check-production-ops-secret-backend-operator-runbook-negative-gates-readiness-v1.py` 固定 `production-secret-backend-operator-runbook-negative-gates-readiness-v1`；它不实现 resolver runtime、不创建 fake resolver、不调用云 secret 服务、不读取 secret value、不接数据库、不启用 repository mode。
8. `rotation-and-audit-policy`
   - 新增 [Production Secret Backend Rotation / Audit Policy Readiness v1](production-secret-backend-rotation-audit-policy-readiness-v1-plan.md)，把 rotation trigger、approval / change window、secret ref versioning、rollback / disable policy、sanitized verification、audit event fields、failure mapping、no fallback、no side effects 和 artifact guard 固定为可检查证据。
   - 当前已落地 readiness gate：`scripts/checks/fixtures/production-secret-backend-rotation-audit-policy-readiness-v1.json` 与 `scripts/check-production-ops-secret-backend-rotation-audit-policy-readiness-v1.py` 固定 `production-secret-backend-rotation-audit-policy-readiness-v1`；它不实现 rotation runtime、不写 audit store、不实现 resolver runtime、不创建 fake resolver、不调用云 secret 服务、不读取 secret value、不接数据库、不启用 repository mode。
9. `test-fixture-strategy-fake-resolver-entry-review`
   - 新增 [Production Secret Backend Test Fixture Strategy / Fake Resolver Implementation Entry Review v1](production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1-plan.md)，把 `test-fixture-strategy` 与 fake resolver implementation entry 的 blocked 结论、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard 固定为可检查证据。
   - 新增 [Production Secret Backend Fake Resolver Contract / No Secret Leakage Smoke Strategy v1](production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1-plan.md)，把 fake resolver static contract、输入 / 输出 allowlist、禁止字段、no secret leakage smoke strategy、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard 固定为可检查证据；当前仍不创建 fake resolver runtime 或 no secret leakage smoke runtime。
   - 新增 [Production Secret Backend Fake Resolver Implementation Task Card Entry Readiness Review v1](production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1-plan.md)，把下一张 fake resolver implementation task card 的准入评审固定为 `fake_resolver_implementation_task_card_entry_readiness_review_defined`；当前只允许下一步创建任务卡，不实现 fake resolver runtime。
   - 当前已落地 entry review：`scripts/checks/fixtures/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.json` 与 `scripts/check-production-ops-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.py` 固定 `production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1`；它不实现 resolver runtime、不实现 fake resolver runtime、不创建 no secret leakage smoke runtime、不调用云 secret 服务、不读取 secret value、不接数据库、不启用 repository mode。
10. `startup-supervisor-boundary`
   - 梳理本地启动脚本和 deployment smoke。
   - 固定当前只支持人工启动 / smoke，不声明 process supervisor。
   - 当前已落地 governance boundary：`scripts/checks/fixtures/production-ops-startup-supervisor-boundary.json` 与 `scripts/check-production-ops-startup-supervisor-boundary.py` 固定 platform wrapper、console dev launcher、readiness probe、退出码、日志路径和 process supervisor 未完成口径；这不等于 process supervisor ready。
11. `environment-isolation`
   - 固定 local/dev/prod profile 语义和禁止误用的负向场景。
   - 确认 local-smoke 只能代表本地 readiness。
   - 当前已落地 governance boundary：`scripts/checks/fixtures/production-ops-environment-isolation-boundary.json` 与 `scripts/check-production-ops-environment-isolation-boundary.py` 固定 local readiness、dev smoke 和 production readiness 的区分；这不等于 deployment environment isolation ready。
11. `console-production-package-smoke`
   - 给 console production build / preview / package 形成最小可复验入口。
   - 继续保持只读边界。
   - 当前已落地 governance boundary：`scripts/checks/fixtures/production-ops-console-package-smoke.json` 与 `scripts/check-production-ops-console-package-smoke.py` 固定 build / preview / private package / artifact policy；这不等于 console production package ready。
12. `short-close-checklist-refresh`
   - 更新 P3 checklist，将已完成项和仍 blocked 项分清。
   - 只在新增生产声明、配置格式或执行边界时补专项门禁。
   - 当前已落地：`scripts/checks/fixtures/p3-local-product-shell-short-close-checklist.json` 增加 `production_ops_hardening_refresh`，`scripts/check-p3-local-product-shell-short-close-checklist.py` 跨读四个 production ops boundary fixture，确认四个切片只是 governance boundary，真实 production 条件仍为 `not_satisfied`。

## 验收口径

- 当前焦点和路线图已把真实模型产出转为后置专题。
- production hardening v1 范围、停止线和切片顺序已写清。
- 所有新增 production 相关声明必须有检查或明确证据来源。
- `P3` 的 `local usable / read-only close` 不被误写成 production ready。
- `P2` 的 executor / storage / confirmation / replay 停止线不变。
- `pwsh ./scripts/check-repo.ps1 -Fast` 通过。

## 下一步

已根据 Radish 的 docker local/test/prod 模式新增 [Production Ops Docker Deployment v1 计划](production-ops-docker-deployment-v1-plan.md)，并用 `docker-deployment-mode-definition` 固定后续部署方向：开发期宿主机直跑，本地容器验证使用 `docker-local-compose`，测试和生产共用部署态 compose。当前 `docker-local-compose`、`docker-test-prod-compose`、`docker-image-build-publish`、`deployment-readiness-smoke`、`container-smoke-runbook` 与 `container-smoke-record-template` 已分别用 `production-ops-docker-local-compose.json`、`production-ops-docker-test-prod-compose.json`、`production-ops-docker-image-build-publish.json`、`production-ops-deployment-readiness-smoke.json`、`production-ops-container-smoke-runbook.json` 和 `production-ops-container-smoke-record-template.json` 固定为可检查边界，部署态资产为 `deploy/docker-compose.yaml` 与 `deploy/.env.example`。本任务卡的静态边界已经 close，2026-05-26 已完成一次本地 `docker_local` container smoke 运行记录；production secret backend 已继续补齐 config / provider profile / disabled resolver / operator runbook / rotation audit / fake resolver strategy / test-only fake resolver runtime / 真实 resolver runtime preconditions / 真实 resolver runtime implementation entry review / resolver backend profile selection readiness / real resolver no leakage smoke runtime strategy / credential handle runtime boundary readiness / operator approval runtime evidence readiness 证据，但它们只定义 contract、前置条件、reference-only manifest、配置注入边界、provider/profile binding、disabled resolver interface、operator runbook、negative gates、rotation / audit policy、test-only fake resolver runtime、真实 resolver runtime 前置、blocked-before-task-card 结论、backend profile selection、no leakage、credential handle boundary 和 operator approval runtime evidence 静态前置，不实现真实云 secret 服务、不写入真实 secret、不实现 production resolver runtime、不创建 backend runtime、不创建 no secret leakage smoke runtime、不创建 credential handle runtime、不创建或执行 approval runtime、不声明 production ready。后续若继续真实 production secret backend，必须单独推进 audit store handoff、backend health boundary 或其它单一 blocker。除非新任务卡明确存储方案、部署目标和验证边界，否则不实现 production resolver runtime、process supervisor、deployment environment isolation 或 console production packaging。

## 停止线

- 不把 local-smoke 当作 production health。
- 不把 mock provider、demo profile 或本地 wrapper 写成生产可用。
- 不实现真实 secret backend，除非另开任务卡并明确存储方案。
- 不实现 process supervisor，除非另开任务卡并明确部署目标。
- 不新增 confirmation、writeback、replay 或 materialized result reader。

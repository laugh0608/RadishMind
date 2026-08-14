# Provider 上报用量规范化与应用用量审查（开发 / 测试态）v1 实施任务卡

更新时间：2026-07-27

状态：`completed`

## 目标

在既有 Provider Adapter、GatewayEnvelope、Request History 和 Application Operations 边界内交付可信 reported usage，不引入 token 估算、计费或第二套用量真相源。

## 批次

### A：Provider 与 Gateway contract

- 定义 `reported / not_reported` 规范化结构与稳定 source。
- 覆盖 OpenAI-compatible、Gemini、Anthropic、Ollama shape 及流式终态 chunk。
- Gateway envelope schema 必须包含经过验证的 usage；mock 保持 `not_reported`。

### B：Platform 与 northbound

- recorder 复验并写入现有 `gateway_request_record.v1`。
- 三类 northbound unary / stream 只投影有证据的 usage。
- 非法或缺失 usage 不改写业务响应，且不得落成 reported。

### C：Web 审查

- Request History list / detail 显示 reported source 与 token counts。
- Application Operations timeline 和当前查询窗口 metrics 汇总 reported tokens。
- consumer 严格校验 availability、source、计数关系与 application scope。

### D：产品连续验证

- Python contract fixture、Go recorder / protocol / repository、Web consumer / panel 测试。
- SQLite 与 PostgreSQL 持久化、过滤和重启恢复。
- 快速与全量仓库门禁，并同步功能入口、当前焦点、路线图、能力矩阵和周志。

## 验证

```bash
.venv/bin/python -m unittest services.gateway.tests.test_copilot_gateway services.runtime.tests.test_inference_provider
(cd services/platform && go test ./internal/httpapi ./internal/bridge)
(cd services/platform && go test -race ./internal/httpapi)
npm --prefix apps/radishmind-web test
npm --prefix apps/radishmind-web run build
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```

数据库验证复用既有 Gateway Request History SQLite tests 与 PostgreSQL 集成入口，不新增只检查文档存在性的脚本。

完成证据：

- Python 9 项 Gateway / Provider usage 测试覆盖五类来源、stream、thinking、cache、部分字段、负数、布尔值和总数不一致。
- Platform 全包 Go tests、Gateway usage / northbound / recorder / SQLite 定向测试通过。
- PostgreSQL integration suite 通过 reported usage 落库、过滤、重启恢复及既有 migration / role / no-fallback 链。
- Web 255 项测试与 production build 通过；Request History 与 Application Operations consumer 对 reported usage 执行严格 source 和计数关系校验。
- 仓库快速与全量门禁结果记录在本周周志。

## 停止线

- 不估算缺失 token，不把零值解释为 Provider 报告。
- 不新增 cost / quota / billing owner 或生产计费声明。
- 不改变 Provider 调度、retry、fallback 或 compatibility route 授权边界。
- 不保存 Provider raw usage envelope、请求响应正文、credential、endpoint 或自由文本。
- 不为普通 UI 汇总新增 migration、fixture repository 或独立 checker。

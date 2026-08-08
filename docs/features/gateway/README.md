# Gateway 细专题入口

更新时间：2026-08-08

本目录承接 `Model Gateway / API Distribution` 的具体运行时、协议兼容和用户路径专题。上层方向、优先级与生产停止线仍以 [Model Gateway / API Distribution](../model-gateway-api-distribution.md) 为准。

## 当前专题

| 专题 | 状态 | 当前作用 |
| --- | --- | --- |
| [Gateway Python Bridge Runtime v1](python-bridge-runtime-v1.md) | `gateway_bridge_stdio_worker_pool_completed` | 受控 `stdio` worker pool 已成为默认模式，完成生命周期、排队、取消、崩溃重建、请求隔离和性能验收；process 模式保留回滚 |
| [Model Gateway Request History / Usage & Failure Review v1](model-gateway-request-history-usage-failure-review-v1.md) | `model_gateway_request_history_usage_failure_review_v1_complete` | `memory_dev`、SQLite、PostgreSQL dev/test、失败 / 取消终态、分页详情、重启恢复和真实 Web 审查已完成；S5 以当前 application / workspace / consumer 编排该 owner，并拒绝切换后的迟到响应 |
| [Provider 上报用量规范化与应用用量审查（开发 / 测试态）v1](provider-reported-usage-normalization-application-review-dev-test-v1.md) | `provider_reported_usage_normalization_application_review_dev_test_v1_completed` | 五类 Provider usage 规范化、三协议 unary / stream 投影、Request History 持久化及应用当前窗口 token 审查已完成；估算、成本、quota 与 billing 关闭 |
| [Gateway Playground / Request Review Loop v1](gateway-playground-request-review-loop-v1.md) | `gateway_playground_request_review_loop_v1_complete` | 三协议 unary / stream、取消、稳定失败和 request-id history handoff 已完成；S5 只开放目录真实协议并要求 active application、workspace 一致与易失 credential，输入输出仅保留在组件内存 |

普通 Gateway UI 文案、只读 evidence 和布局改动不在本目录新增专题；它们继续复用 Web build、Gateway smoke 与仓库聚合门禁。

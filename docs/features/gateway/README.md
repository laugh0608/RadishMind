# Gateway 细专题入口

更新时间：2026-07-12

本目录承接 `Model Gateway / API Distribution` 的具体运行时、协议兼容和用户路径专题。上层方向、优先级与生产停止线仍以 [Model Gateway / API Distribution](../model-gateway-api-distribution.md) 为准。

## 当前专题

| 专题 | 状态 | 当前作用 |
| --- | --- | --- |
| [Gateway Python Bridge Runtime v1](python-bridge-runtime-v1.md) | `gateway_bridge_stdio_worker_pool_completed` | 受控 `stdio` worker pool 已成为默认模式，完成生命周期、排队、取消、崩溃重建、请求隔离和性能验收；process 模式保留回滚 |
| [Model Gateway Request History / Usage & Failure Review v1](model-gateway-request-history-usage-failure-review-v1.md) | `model_gateway_request_history_usage_failure_review_v1_defined` | 固定真实 northbound 请求历史的 caller scope、生命周期、sanitized usage / timing、失败语义、分页、独立 dev/test repository 和 Web 审查路径；下一步由单张纵向任务卡实施 |

普通 Gateway UI 文案、只读 evidence 和布局改动不在本目录新增专题；它们继续复用 Web build、Gateway smoke 与仓库聚合门禁。

# Gateway 细专题入口

更新时间：2026-07-11

本目录承接 `Model Gateway / API Distribution` 的具体运行时、协议兼容和用户路径专题。上层方向、优先级与生产停止线仍以 [Model Gateway / API Distribution](../model-gateway-api-distribution.md) 为准。

## 当前专题

| 专题 | 状态 | 当前作用 |
| --- | --- | --- |
| [Gateway Python Bridge Runtime v1](python-bridge-runtime-v1.md) | `gateway_bridge_runtime_baseline_completed` | 已完成 process-per-request 分段实测并选定受控 `stdio` worker pool；下一批实现生命周期、排队、取消、崩溃重建与请求隔离 |

普通 Gateway UI 文案、只读 evidence 和布局改动不在本目录新增专题；它们继续复用 Web build、Gateway smoke 与仓库聚合门禁。

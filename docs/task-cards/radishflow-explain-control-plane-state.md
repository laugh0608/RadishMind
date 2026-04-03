# `RadishFlow` 任务卡：`explain_control_plane_state`

更新时间：2026-04-03

## 任务目标

把 entitlement、manifest、lease、package sync 等控制面状态摘要，转换为操作者可读的解释、问题定位和候选排查建议。

当前任务面向的是“控制面状态说明”，不是凭据处理器，也不是自动修复器。

## 请求映射

- `project`: `radishflow`
- `task`: `explain_control_plane_state`
- 请求结构遵循 [CopilotRequest schema](../../contracts/copilot-request.schema.json)

## 最小必需输入

请求至少应包含以下内容：

- `context.control_plane_state`
- `context.document_revision` 或其他可追踪的本地上下文版本号

若状态说明依赖补充文本或摘要，建议通过 `artifacts` 传入，例如：

- manifest 摘要
- lease 刷新结果摘要
- sync 错误文本摘要

## 可选补充输入

- `context.solve_session`
- `context.latest_snapshot`
- 当前操作者看到的 UI 错误说明或日志摘录

这些补充输入用于增强上下文，但不应包含安全凭据或本机秘密。

## 禁止透传

以下内容不应进入本任务请求：

- `access_token`、`refresh_token`、cookie
- `credential_handle`
- 本机密钥路径、证书密码或安全存储句柄
- 未脱敏的 auth cache、原始授权响应或完整权限票据

## 输出要求

响应结构遵循 [CopilotResponse schema](../../contracts/copilot-response.schema.json)。

本任务对字段的要求如下：

- `summary` 必须说明当前控制面总体状态
- `answers` 必须分别覆盖“当前状态”“可能原因”“建议的人工排查方向”
- `issues` 必须把 entitlement、lease、manifest 或 sync 相关问题拆开列出
- 若状态之间存在冲突、版本错位或上游访问边界不清，应显式保留“未证实/仅为候选解释”的口径，不把外部 403 或单个字段直接写成已确认根因
- `citations` 必须引用状态摘要、错误文本或相关 artifact
- `proposed_actions` 可以包含只读检查或人工操作建议，但不能包含自动凭据处理

## 风险分级规则

- `low`: 纯状态解释或只读检查建议
- `medium`: 需要人工执行的排查、刷新或重试建议
- `high`: 涉及凭据替换、权限提升、绕过校验、包体信任边界或其他安全敏感操作

## `requires_confirmation` 触发条件

以下情况必须为 `true`：

- 建议操作者执行任何非只读动作
- 建议可能影响 entitlement、lease、manifest 或 package sync 状态
- 建议触及安全、授权或信任边界

如果响应只解释状态且没有操作建议，通常应为 `false`。

## 正反例边界

正例：

- 解释当前 lease 过期为何导致功能不可用
- 解释 manifest 与本地状态不一致对操作者的实际影响
- 解释 public / private package source 权限范围差异为何只影响部分包源，而不是直接宣称全局授权失效
- 明确说明“建议人工重试 sync”而不是伪装成已自动修复

反例：

- 输出完整凭据操作步骤或替代安全真相源
- 鼓励绕过校验、跳过授权检查或直接修改本机安全状态
- 把模糊 UI 报错解释成已确认的后端根因

## 最小输入快照示例

```json
{
  "project": "radishflow",
  "task": "explain_control_plane_state",
  "artifacts": [
    {
      "kind": "text",
      "role": "supporting",
      "name": "sync_error_summary",
      "mime_type": "text/plain",
      "content": "package sync failed: remote manifest unavailable"
    }
  ],
  "context": {
    "document_revision": 27,
    "control_plane_state": {
      "entitlement_status": "active",
      "lease_status": "expired",
      "sync_status": "failed",
      "last_error": "remote manifest unavailable"
    }
  }
}
```

## 评测口径

当前阶段至少检查以下指标：

- 结构合法率：响应必须通过 schema 校验
- 状态解释正确率：解释必须与控制面摘要一致
- 安全边界遵守率：不得泄露凭据或建议越权操作
- 风险分级正确率：只读解释与敏感操作建议要正确区分
- 操作建议可执行率：若给出人工建议，必须具体且可理解

## 非目标

当前任务不负责：

- 直接处理凭据、token 或 auth cache
- 自动执行重试、续租、同步或授权动作
- 替代控制面的最终状态判定

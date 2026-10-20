export const usageQuota = {
  "unavailable": "不可用",
  "monthly": "每月",
  "workspace": "用户工作区",
  "title": "用量与配额",
  "readOnlyReady": "只读就绪",
  "blocked": "已阻止",
  "route": "配额摘要路由",
  "model": "读取模型",
  "period": "周期",
  "request": "请求",
  "audit": "审计",
  "snapshot": "工作区用量配额快照",
  "quota": "配额",
  "window": "只读配额窗口",
  "failure": "失败代码",
  "noEnforcement": "仅报告，此页面不执行配额限制",
  "available": "可用",
  "references": "请求与审计引用保持可见",
  "noPolicy": "没有可用的配额策略摘要。",
  "limits": "工作区用量配额上限",
  "limit": "上限 {{limit}} / 已用 {{percent}}",
  "noCurrency": "来源未提供币种；按原数值展示，不换算或推断币种。",
  "overQuota": "超出配额失败代码",
  "boundary": "仅展示读取侧元数据；此页面不执行配额限制、限流或成本记录写入。",
  "states": "配额状态示例",
  "failureCount": "{{count}} 项 / 失败 {{code}}",
  "measures": {
    "requests": "请求数",
    "tokens": "Token 数",
    "cost": "成本数值"
  },
  "previews": {
    "ready": {
      "label": "就绪",
      "summary": "配额摘要使用当前读取投影，此页面不执行配额限制。"
    },
    "empty": {
      "label": "空态",
      "summary": "未返回配额策略摘要时，路由元数据仍保持可见。"
    },
    "denied": {
      "label": "拒绝",
      "summary": "用量权限拒绝时，不暴露部分配额或用量记录。"
    },
    "stale": {
      "label": "过期",
      "summary": "缓存配额摘要与实时成本记录及网关执行限制保持区分。"
    },
    "partial_failure": {
      "label": "部分失败",
      "summary": "无法完整读取配额详情时，明确保留失败元数据。"
    },
    "forbidden_projection": {
      "label": "禁止的投影",
      "summary": "共享输出守卫阻止成本写回载荷投影。"
    }
  }
} as const;

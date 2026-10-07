export const adminQuota = {
  "policiesLabel": "应用配额策略",
  "policies": "应用策略",
  "oneDetail": "当前应用详情",
  "selectionRequired": "需要选择应用",
  "applicationList": "当前工作区中的应用",
  "devTest": "开发 / 测试",
  "notLoaded": "未加载",
  "noApplications": "当前工作区没有可用应用。",
  "selectionNote": "选择应用以查看其配额详情；策略状态标签与当前选择分别展示。",
  "selectedApplication": "所选应用",
  "authority": "权限：admin_gateway_quotas:read / write",
  "eyebrow": "配额 · admin_gateway_quotas:read / write",
  "title": "UTC 每日模型提供方尝试策略",
  "usageSource": "用量由配额服务提供，不从请求历史或旧 QuotaSummary 推导。",
  "tenant": "租户",
  "workspace": "工作区",
  "environment": "环境",
  "application": "应用",
  "selectApplication": "请选择一个应用",
  "selectRequired": "需要选择应用",
  "boundary": "生产配额、token / 成本限制、计费、删除 / 禁用、自动提高上限、自动路由及生产成员 / OIDC 接入仍未开放。",
  "admittedAttempts": "已准入的模型提供方尝试",
  "remaining": "今日剩余",
  "policy": "策略",
  "period": "周期",
  "recordVersion": "记录版本",
  "updatedBy": "更新人",
  "missingTitle": "此精确应用范围尚无配额策略",
  "policyUpdate": "策略更新",
  "createLimit": "创建 UTC 每日请求上限",
  "changeLimit": "修改 UTC 每日请求上限",
  "limitHelp": "仅接受 1–1,000,000 的正整数，不支持删除或禁用",
  "requestLimit": "请求上限",
  "reload": "重新加载当前策略",
  "reviewUpdate": "审查更新",
  "confirmationLabel": "配额策略更新确认",
  "confirmationHelp": "变更仅应用于精确的租户、工作区、环境和应用范围；当前已准入次数不会重置或重新计算。",
  "staleHelp": "若预期版本已过期，写入将被拒绝，必须重新加载当前策略。",
  "cancel": "取消",
  "updating": "正在更新…",
  "confirmUpdate": "确认更新",
  "casLabel": "配额策略 CAS 边界",
  "casGuard": "CAS 保护",
  "reviewHelp": "审查会打开显式确认区；编辑器本身不会直接发送更新。",
  "loading": "正在加载精确配额策略…",
  "loadingHelp": "当前配额尚未确定，暂时不能更新策略。",
  "unavailableTitle": "配额响应不可用",
  "noAcceptedPolicy": "未接受任何策略或用量数据。",
  "retry": "重试精确配额查询",
  "version": "版本 {{version}}",
  "window": "UTC 窗口 · {{periodStart}}",
  "usage": "{{admitted}} / {{limit}}",
  "progress": "已准入次数占请求上限的 {{percent}}%",
  "expectedVersion": "预期版本 {{version}}",
  "confirmationVersion": "确认 · 预期版本 {{version}}",
  "createRequests": "创建 {{limit}} 次请求上限",
  "updateRequests": "请求上限 {{current}} → {{limit}} 次",
  "missingHelp": "以 expected_version = 0 创建正整数 UTC 每日请求上限。策略缺失时保持拒绝准入，不以请求历史或旧 QuotaSummary 替代。",
  "operations": {
    "invalidRead": "配额响应未通过严格校验，未接受策略或用量数据。",
    "invalidLimit": "请求上限必须为 1 至 1,000,000 的正整数。",
    "unchangedLimit": "请先输入不同的请求上限，再审查更新。",
    "conflict": "预期版本已过期；请重新加载当前配额策略，再审查更新。",
    "updated": "策略版本 {{version}} 已成为当前版本。",
    "invalidUpdate": "无法确认配额更新响应；请重新加载当前策略后再尝试。"
  },
  "states": {
    "ready": "策略已就绪",
    "missing": "策略缺失",
    "loading": "加载中",
    "failed": "已阻止",
    "limitReached": "已达上限",
    "active": "有效",
    "draft": "草案",
    "archived": "已归档",
    "unknown": "未知"
  },
  "failures": {
    "gateway_quota_disabled": {
      "shortLabel": "HTTP 未开放",
      "title": "开发 / 测试态配额管理未开放",
      "summary": "管理读取或写入入口未启用，没有可用的回退策略。"
    },
    "gateway_quota_scope_denied": {
      "shortLabel": "权限被拒绝",
      "title": "缺少配额权限",
      "summary": "此工作区需要精确的 admin_gateway_quotas 读取或写入权限。"
    },
    "gateway_quota_environment_forbidden": {
      "shortLabel": "环境不允许",
      "title": "环境超出开发 / 测试边界",
      "summary": "仅接受 development 或 test，生产配额仍未开放。"
    },
    "gateway_quota_payload_invalid": {
      "shortLabel": "请求被拒绝",
      "title": "配额更新请求被拒绝",
      "summary": "请求被拒绝；请检查正整数上限后再尝试。"
    },
    "gateway_quota_policy_not_found": {
      "shortLabel": "策略缺失",
      "title": "尚无配额策略",
      "summary": "以预期版本 0 创建首个策略；策略缺失时保持拒绝准入。"
    },
    "gateway_quota_policy_version_conflict": {
      "shortLabel": "版本冲突",
      "title": "配额策略版本已改变",
      "summary": "过期版本写入已被拒绝；请重新加载精确策略后再审查更新。"
    },
    "gateway_quota_attempt_conflict": {
      "shortLabel": "尝试冲突",
      "title": "配额准入尝试发生冲突",
      "summary": "模型提供方尝试未继续；请重新加载配额策略后再操作。"
    },
    "gateway_quota_exceeded": {
      "shortLabel": "已达上限",
      "title": "已达到 UTC 每日请求上限",
      "summary": "新的模型提供方尝试被阻止；此状态不会改变所选应用。"
    },
    "gateway_quota_store_unavailable": {
      "shortLabel": "存储不可用",
      "title": "配额存储不可用",
      "summary": "无法读取精确配额策略，不会将应用视为无限额。"
    }
  }
} as const;

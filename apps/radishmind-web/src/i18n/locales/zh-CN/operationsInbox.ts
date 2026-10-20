export const operationsInbox = {
  "heading": "运营收件箱",
  "items": "工作区运营关注项",
  "openSelected": "打开所选项",
  "viewWorkspace": "查看工作区",
  "readOnlyEvidence": "只读证据。请到对应功能页面继续审查，此处不执行任何动作。",
  "evidencePath": "证据路径",
  "projection": "来源投影",
  "severity": "关注级别",
  "authority": "权限",
  "readOnly": "只读",
  "mutation": "修改",
  "remediation": "修复",
  "businessTruth": "业务真相源",
  "openEvidence": "打开证据",
  "select": "选择关注项以检查其证据路径。",
  "noAttention": "无关注项",
  "noResources": "无可用资源项",
  "complete": "当前来源窗口完整。",
  "checkCoverage": "请先检查来源覆盖范围，再判断空队列的含义。",
  "unavailable": "不可用",
  "enabled": "已启用",
  "locked": "已锁定",
  "reviewOnly": "仅审查",
  "writable": "可写",
  "sourceCount": "已加载 {{count}} 个已授权来源投影",
  "sources": {
    "applications": "应用",
    "api_keys": "API 密钥",
    "workflow_definitions": "工作流",
    "runs": "运行"
  },
  "reasons": {
    "application_archived": {
      "title": "{{name}} 已归档",
      "summary": "已归档应用仍可用于只读历史审查。"
    },
    "application_run_attention": {
      "title": "{{name}} 报告状态：{{status}}",
      "summary": "请打开应用并检查其权威运行历史。"
    },
    "api_key_rotation_required": {
      "title": "API 密钥需要轮换",
      "summary": "签发或撤销凭据前，请先审查现有密钥生命周期。"
    },
    "api_key_expired": {
      "title": "API 密钥已过期",
      "summary": "请审查脱敏密钥记录，此处不提供凭据材料。"
    },
    "api_key_expiring": {
      "title": "API 密钥将在 14 天内过期",
      "summary": "请在明确的到期时间之前审查现有密钥生命周期。"
    },
    "workflow_definition_review": {
      "title": "{{name}} 状态：{{status}}",
      "summary": "请审查不可变定义状态；此收件箱不能批准或激活定义。"
    },
    "run_failure": {
      "title": "{{name}} 需要失败审查",
      "summary": "请检查失败 {{code}}；回放、恢复与自动修复仍关闭。"
    },
    "run_outcome_unknown": {
      "title": "{{name}} 结果未知",
      "summary": "请先检查运行元数据，再决定是否适合再次显式调用。"
    }
  }
} as const;

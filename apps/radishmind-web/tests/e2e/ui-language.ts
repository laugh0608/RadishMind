import type { Page } from "@playwright/test";

// Independent product expectations: do not import implementation dictionaries here.
const labels: Record<string, readonly [string, string]> = {
  "Create application": ["Create application", "创建应用"],
  "Server-generated identity": ["Server-generated identity", "由服务端生成标识"],
  "Display name": ["Display name", "显示名称"],
  "Application kind": ["Application kind", "应用类型"],
  "Create and select": ["Create and select", "创建并选中"],
  "Application development context": ["Application development context", "应用开发上下文"],
  "Prompt Application Session v2": ["Prompt Application Session v2", "Prompt 应用会话 v2"],
  "Application result artifacts": ["Application result artifacts", "应用结果资产"],
  "Load saved drafts": ["Load saved drafts", "加载已保存草案"],
  "Saved valid draft": ["Saved valid draft", "已保存且有效的草案"],
  "Create immutable candidate": ["Create immutable candidate", "创建不可变候选版本"],
  "Read exact source": ["Read exact source", "读取精确源码"],
  "审查输出契约": ["Review output contract", "审查输出契约"],
  "Review reason": ["Review reason", "审查理由"],
  "Record review decision": ["Record review decision", "记录审查决定"],
  "重读 assignment 与事件": ["Reload assignment and events", "重读运行绑定与事件"],
  "CAS action": ["CAS action", "按版本执行的动作"],
  "记录显式 runtime 决策": ["Record explicit runtime decision", "记录显式运行决策"],
  "Issue new credential": ["Issue new credential", "签发新凭据"],
  "Issue API key": ["Issue API key", "签发 API 密钥"],
  "Use in Playground": ["Use in Playground", "在调试台中使用"],
  "Load models": ["Load models", "加载模型"],
  "Validated model": ["Validated model", "已验证模型"],
  "Default protocol": ["Default protocol", "默认协议"],
  "Default model": ["Default model", "默认模型"],
  "Save draft": ["Save draft", "保存草案"],
  "Kind": ["Kind", "类型"],
  "输出 JSON Schema": ["Output JSON Schema", "输出 JSON Schema"],
  "Save with CAS": ["Save with CAS", "按版本保存"],
  "Refresh": ["Refresh", "刷新"],
  "Create immutable version": ["Create immutable version", "创建不可变版本"],
  "不可变版本输出契约": ["Immutable version output contract", "不可变版本输出契约"],
  "Load drafts": ["Load drafts", "加载草案"],
  "Valid Prompt Application draft": ["Valid Prompt Application draft", "有效的 Prompt 应用草案"],
  "Immutable template version": ["Immutable template version", "不可变模板版本"],
  "Bind and open Publish Review": ["Bind and open Publish Review", "绑定并打开发布审查"],
  "Create Session v2": ["Create Session v2", "创建会话 v2"],
  "Turn variables": ["Turn variables", "本轮变量"],
  "Execute Prompt turn": ["Execute Prompt turn", "执行 Prompt 轮次"],
  "Validate": ["Validate", "校验"],
  "Active Prompt Session v2 records": ["Active Prompt Session v2 records", "有效的 Prompt 会话 v2 记录"],
  "(没有当前 turn output)": ["(No output for the current turn)", "（没有当前轮次输出）"],
  "Refresh artifacts": ["Refresh artifacts", "刷新结果资产"],
  "Open exact run": ["Open exact run", "打开精确运行记录"],
  "active result artifacts": ["active result artifacts", "活跃结果资产"],
  "runs workflow review owner": ["Run and evaluation review owner", "运行与评测审查模块"],
  "Description": ["Description", "描述"],
};
const languages = new WeakMap<Page, "zh-CN" | "en-US">();
export function setTestLanguage(page: Page, locale: "zh-CN" | "en-US") { languages.set(page, locale); }
export function uiText(page: Page, source: string) {
  const entry = labels[source];
  if (!entry) return source;
  return entry[languages.get(page) === "zh-CN" ? 1 : 0];
}

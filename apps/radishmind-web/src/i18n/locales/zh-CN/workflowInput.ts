export const workflowInput = {
  "structuredRuntimeInput": "结构化运行输入",
  "notSet": "未设置",
  "clear": "清除",
  "inputsAreRetainedOnlyForTheCurrentRequestDurableRecords": "输入只在当前请求期间保留；持久化记录仅保存合同、字段名／类型、bytes 与 digest。",
  "error_unknown_field": "该字段不属于当前输入契约。",
  "error_required": "这是必填字段。",
  "error_boolean": "请选择 true 或 false。",
  "error_string": "该字段必须是字符串。",
  "error_integer": "该字段必须是整数。",
  "error_number": "该字段必须是数值。",
  "error_string_budget": "文本不得超过 4096 字节。",
  "error_secret_material": "输入中不得包含凭据、令牌、密码或连接串。",
  "error_integer_syntax": "请输入十进制整数。",
  "error_integer_range": "整数超出安全范围。",
  "error_number_syntax": "请输入 JSON 有限数值。"
} as const;

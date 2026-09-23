export type PromptTemplateJSONSchema = {
  type: "object" | "array" | "string" | "integer" | "number" | "boolean";
  properties?: Record<string, PromptTemplateJSONSchema>;
  required?: string[];
  additionalProperties: boolean;
  items?: PromptTemplateJSONSchema;
};

const schemaIssueMessages = {
  "depth": "嵌套深度不能超过 8 层（根对象为第 1 层）。",
  "object": "必须是 schema 对象。",
  "keyword": "包含不支持的 schema 关键字。",
  "additionalProperties": "additionalProperties 必须是布尔值。",
  "properties": "properties 必须是对象。",
  "required": "required 必须是无重复的合法字段名数组。",
  "objectItems": "object 不能声明 items。",
  "fields": "全部对象合计不能超过 128 个字段。",
  "requiredProperties": "required 中的字段必须在 properties 中声明。",
  "fieldName": "字段名须以英文字母开头，只含字母、数字或下划线，最长 64 字符。",
  "scalarProperties": "非 object 类型不能声明对象字段或允许额外字段。",
  "type": "类型仅支持 object、array、string、integer、number、boolean。",
  "scalarItems": "标量类型不能声明 items。",
  "root": "根类型必须是 object。",
  "bytes": "紧凑 JSON 不能超过 32 KiB。",
  "json": "请输入完整有效的 JSON，检查引号、逗号和括号。"
} as const;
export type PromptSchemaIssue = { code: keyof typeof schemaIssueMessages; path: string };

const namePattern = /^[A-Za-z][A-Za-z0-9_]{0,63}$/u;
const keys = new Set(["type", "properties", "required", "additionalProperties", "items"]);

// These are the existing Prompt output budgets, not general JSON Schema support.
export function validatePromptOutputSchema(value: unknown, requireExplicitAdditionalProperties = true, reportIssue?: (issue: PromptSchemaIssue) => void): string {
  let properties = 0;
  function issue(code: PromptSchemaIssue["code"], path: string): string {
    reportIssue?.({ code, path });
    return `${path}：${schemaIssueMessages[code]}`;
  }
  function visit(node: unknown, depth: number, path: string): string {
    const fail = (code: PromptSchemaIssue["code"]) => issue(code, path);
    if (depth > 8) return fail("depth");
    if (!record(node)) return fail("object");
    if (Object.keys(node).some((key) => !keys.has(key))) return fail("keyword");
    if (typeof node.additionalProperties !== "boolean" && (requireExplicitAdditionalProperties || node.additionalProperties !== undefined)) return fail("additionalProperties");
    if (node.properties !== undefined && !record(node.properties)) return fail("properties");
    if (node.required !== undefined && (!Array.isArray(node.required) ||
      !node.required.every((name) => typeof name === "string" && namePattern.test(name)) ||
      new Set(node.required).size !== node.required.length)) return fail("required");
    const fields = record(node.properties) ? node.properties : {};
    const required = (node.required ?? []) as string[];
    if (node.type === "object") {
      if (node.items !== undefined) return fail("objectItems");
      properties += Object.keys(fields).length;
      if (properties > 128) return fail("fields");
      if (required.some((name) => !Object.hasOwn(fields, name))) return fail("requiredProperties");
      for (const [name, nested] of Object.entries(fields)) {
        if (!namePattern.test(name)) return fail("fieldName");
        const error = visit(nested, depth + 1, `${path}.properties.${name}`);
        if (error) return error;
      }
      return "";
    }
    if (Object.keys(fields).length || required.length || node.additionalProperties) return fail("scalarProperties");
    if (node.type === "array") return visit(node.items, depth + 1, `${path}.items`);
    if (!["string", "integer", "number", "boolean"].includes(String(node.type))) return fail("type");
    return node.items === undefined ? "" : fail("scalarItems");
  }
  if (!record(value) || value.type !== "object") return issue("root", "schema");
  const error = visit(value, 1, "schema");
  if (error) return error;
  if (new TextEncoder().encode(JSON.stringify(value)).length > 32 * 1024) return issue("bytes", "schema");
  return "";
}

export function parsePromptOutputSchema(source: string): { schema?: PromptTemplateJSONSchema; error: string; issue?: PromptSchemaIssue } {
  let value: unknown;
  try {
    value = JSON.parse(source);
  } catch {
    return { error: `schema：${schemaIssueMessages.json}`, issue: { code: "json", path: "schema" } };
  }
  let issue: PromptSchemaIssue | undefined;
  const report = (value: PromptSchemaIssue) => { issue = value; };
  const error = validatePromptOutputSchema(value, false, report);
  if (error) return { error, issue };
  // The existing Go request contract defaults omitted additionalProperties to false.
  const schema = normalize(value as Record<string, unknown>);
  const normalizedError = validatePromptOutputSchema(schema, true, report);
  return normalizedError ? { error: normalizedError, issue } : { schema, error: "" };
}

function normalize(value: Record<string, unknown>): PromptTemplateJSONSchema {
  return {
    type: value.type as PromptTemplateJSONSchema["type"],
    additionalProperties: value.additionalProperties === true,
    ...(record(value.properties) ? { properties: Object.fromEntries(Object.entries(value.properties).map(([name, nested]) => [name, normalize(nested as Record<string, unknown>)])) } : {}),
    ...(value.required === undefined ? {} : { required: value.required as string[] }),
    ...(value.items === undefined ? {} : { items: normalize(value.items as Record<string, unknown>) }),
  };
}

export function formatPromptOutputSchema(schema?: PromptTemplateJSONSchema): string {
  return JSON.stringify(schema ?? { type: "object", additionalProperties: false, properties: {}, required: [] }, null, 2);
}

function record(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

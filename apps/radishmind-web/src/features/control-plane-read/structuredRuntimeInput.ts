const ID_PATTERN = /^[A-Za-z0-9][A-Za-z0-9._:-]{0,159}$/u;
const FIELD_NAME_PATTERN = /^[a-z][a-z0-9_]{0,63}$/u;
const DIGEST_PATTERN = /^sha256:[a-f0-9]{64}$/u;
const MAX_INPUT_FIELDS = 16;
const MAX_INPUT_BYTES = 8192;
const MAX_STRING_BYTES = 4096;
const MAX_SAFE_RUNTIME_INTEGER = 9_007_199_254_740_991;

export type StructuredRuntimeInputValueType = "string" | "integer" | "number" | "boolean";

export type StructuredRuntimeInputField = {
  name: string;
  valueType: StructuredRuntimeInputValueType;
  required: boolean;
  label: string;
  description: string;
};

export type StructuredRuntimeInputContract = {
  contractId: string;
  fields: StructuredRuntimeInputField[];
  summary: string;
  contractDigest: string;
};

export type StructuredRuntimeInputDrafts = Record<string, string | boolean | undefined>;
export type StructuredRuntimeInputValues = Record<string, string | number | boolean>;

export type StructuredRuntimeInputValidation = {
  ok: boolean;
  inputs: StructuredRuntimeInputValues;
  fieldErrors: Record<string, string>;
  failureCode: string;
  summary: string;
};

export function parseStructuredRuntimeInputContractDocument(value: unknown): StructuredRuntimeInputContract | null {
  if (!isExactRecord(value, ["contract_id", "fields", "summary", "contract_digest"])) return null;
  if (!ID_PATTERN.test(String(value.contract_id)) || !DIGEST_PATTERN.test(String(value.contract_digest)) ||
    typeof value.summary !== "string" || utf8Length(value.summary) > 4000 || containsSensitiveText(value.summary) ||
    !Array.isArray(value.fields) || value.fields.length < 1 || value.fields.length > MAX_INPUT_FIELDS) return null;

  const fields: StructuredRuntimeInputField[] = [];
  const names = new Set<string>();
  for (const candidate of value.fields) {
    if (!isExactRecord(candidate, ["name", "value_type", "required", "label", "description"]) ||
      !FIELD_NAME_PATTERN.test(String(candidate.name)) || !isValueType(candidate.value_type) || typeof candidate.required !== "boolean" ||
      typeof candidate.label !== "string" || candidate.label.length < 1 || utf8Length(candidate.label) > 160 ||
      typeof candidate.description !== "string" || utf8Length(candidate.description) > 4000 ||
      containsSensitiveText(String(candidate.name)) || containsSensitiveText(candidate.label) || containsSensitiveText(candidate.description) ||
      names.has(String(candidate.name))) return null;
    names.add(String(candidate.name));
    fields.push({
      name: String(candidate.name),
      valueType: candidate.value_type,
      required: candidate.required,
      label: candidate.label,
      description: candidate.description,
    });
  }
  if (containsSensitiveText(String(value.contract_id))) return null;
  return {
    contractId: String(value.contract_id),
    fields,
    summary: value.summary,
    contractDigest: String(value.contract_digest),
  };
}

export function validateStructuredRuntimeInputDrafts(
  contract: StructuredRuntimeInputContract,
  drafts: StructuredRuntimeInputDrafts,
): StructuredRuntimeInputValidation {
  const allowedNames = new Set(contract.fields.map((field) => field.name));
  const fieldErrors: Record<string, string> = {};
  const inputs: StructuredRuntimeInputValues = {};
  let valueFailureCode = "workflow_input_value_type_invalid";

  for (const name of Object.keys(drafts)) {
    if (!allowedNames.has(name)) {
      return failure("workflow_input_unknown_field", `输入包含合同之外的字段：${name}。`, { [name]: "该字段不属于当前输入合同。" });
    }
  }

  for (const field of contract.fields) {
    const provided = Object.hasOwn(drafts, field.name) && drafts[field.name] !== undefined;
    if (!provided) {
      if (field.required) fieldErrors[field.name] = "这是必填字段。";
      continue;
    }
    const draft = drafts[field.name];
    if (field.valueType === "boolean") {
      if (typeof draft !== "boolean") fieldErrors[field.name] = "请选择 true 或 false。";
      else inputs[field.name] = draft;
      continue;
    }
    if (typeof draft !== "string") {
      fieldErrors[field.name] = `该字段必须是 ${field.valueType}。`;
      continue;
    }
    if (field.valueType === "string") {
      if (utf8Length(draft) > MAX_STRING_BYTES) {
        fieldErrors[field.name] = `文本不得超过 ${MAX_STRING_BYTES} bytes。`;
        valueFailureCode = "workflow_input_budget_exceeded";
      } else if (containsSensitiveText(draft)) {
        fieldErrors[field.name] = "输入中不得包含凭据、token、密码或连接串。";
        valueFailureCode = "workflow_input_secret_material_forbidden";
      }
      else inputs[field.name] = draft;
      continue;
    }
    if (field.valueType === "integer") {
      if (!/^-?(?:0|[1-9][0-9]*)$/u.test(draft)) {
        fieldErrors[field.name] = "请输入十进制整数。";
      } else {
        const number = Number(draft);
        if (!Number.isSafeInteger(number) || Math.abs(number) > MAX_SAFE_RUNTIME_INTEGER) fieldErrors[field.name] = "整数超出安全范围。";
        else inputs[field.name] = number;
      }
      continue;
    }
    const number = Number(draft);
    if (!/^-?(?:0|[1-9][0-9]*)(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?$/u.test(draft) || !Number.isFinite(number)) fieldErrors[field.name] = "请输入 JSON 有限数值。";
    else inputs[field.name] = Object.is(number, -0) ? 0 : number;
  }

  if (Object.keys(fieldErrors).length > 0) {
    const missingRequired = contract.fields.some((field) => field.required && fieldErrors[field.name] === "这是必填字段。");
    return failure(missingRequired ? "workflow_input_required_field_missing" : valueFailureCode, "请修正结构化输入中的字段错误。", fieldErrors);
  }
  const inputBytes = utf8Length(JSON.stringify(inputs));
  if (inputBytes > MAX_INPUT_BYTES) return failure("workflow_input_budget_exceeded", `结构化输入不得超过 ${MAX_INPUT_BYTES} bytes。`);
  return { ok: true, inputs, fieldErrors: {}, failureCode: "", summary: "" };
}

export function structuredRuntimeInputAuthorityKey(contract: StructuredRuntimeInputContract): string {
  return `${contract.contractId}:${contract.contractDigest}`;
}

function failure(failureCode: string, summary: string, fieldErrors: Record<string, string> = {}): StructuredRuntimeInputValidation {
  return { ok: false, inputs: {}, fieldErrors, failureCode, summary };
}

function isValueType(value: unknown): value is StructuredRuntimeInputValueType {
  return value === "string" || value === "integer" || value === "number" || value === "boolean";
}

function containsSensitiveText(value: string): boolean {
  return /(authorization\s*:|bearer\s+|api[_ -]?key|token\s*[=:]|password\s*[=:]|secret\s*[=:]|cookie\s*:|client_secret\s*=|access_token\s*=|refresh_token\s*=|postgres(?:ql)?:\/\/|private[_ ]?key)/iu.test(value);
}

function utf8Length(value: string): number {
  return new TextEncoder().encode(value).length;
}

function isExactRecord(value: unknown, keys: string[]): value is Record<string, any> {
  return typeof value === "object" && value !== null && !Array.isArray(value) &&
    Object.keys(value).length === keys.length && keys.every((key) => Object.hasOwn(value, key));
}

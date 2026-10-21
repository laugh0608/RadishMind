import { parsePromptInputJSON, type PromptInputJSONIssue } from "./promptInputJSON.ts";
import { editPromptVariableValues, serializePromptVariableValues, type PromptVariableValues, type PromptVariableJSON } from "./promptVariableValues.ts";
import { parsePromptApplicationVariables } from "./promptApplicationInvocationConsumer.ts";
import { renderPromptTemplatePreview, type PromptTemplateDraftInput } from "./promptApplicationTemplateConsumer.ts";

export type PromptVariableInput = { mode: "form"; values: PromptVariableValues } | { mode: "json"; text: string };
export type PromptVariableIssue = {
  code: PromptInputJSONIssue["code"] | "numberValue" | "inputType" | "missingVariable" | "undeclaredInput" | "inputSecret" | "renderedBytes" | "referenceInvalid";
  field?: string; line?: number; column?: number;
};
export const emptyPromptVariableInput = (): PromptVariableInput => ({ mode: "form", values: {} });

export function readPromptVariableInput(input: PromptVariableInput): { text: string | null; issues: PromptVariableIssue[] } {
  if (input.mode === "form") {
    const result = serializePromptVariableValues(input.values);
    return { text: result.invalidNumbers.length ? null : JSON.stringify(result.values),
      issues: result.invalidNumbers.map(field => ({ code: "numberValue", field })) };
  }
  const parsed = parsePromptInputJSON(input.text);
  if (!parsed.ok) return { text: null, issues: [parsed.issue] };
  if (!parsed.value || typeof parsed.value !== "object" || Array.isArray(parsed.value)) return { text: null, issues: [{ code: "inputType" }] };
  return { text: input.text, issues: [] };
}

export function switchPromptVariableInput(input: PromptVariableInput, mode: PromptVariableInput["mode"]): { input: PromptVariableInput; issues: PromptVariableIssue[] } {
  if (mode === input.mode) return { input, issues: [] };
  const result = readPromptVariableInput(input);
  if (result.text === null) return { input, issues: result.issues };
  return { input: mode === "json" ? { mode, text: JSON.stringify(JSON.parse(result.text), null, 2) }
    : { mode, values: editPromptVariableValues(JSON.parse(result.text) as Record<string, PromptVariableJSON>) }, issues: [] };
}

export function validatePromptVariableInput(input: PromptVariableInput, source: PromptTemplateDraftInput | null) {
  const result = readPromptVariableInput(input);
  const parsed = parsePromptApplicationVariables(result.text ?? "");
  const issues: PromptVariableIssue[] = [...result.issues];
  if (result.text !== null && !parsed.isValid) issues.push({ code: "inputType" });
  if (input.mode === "form" && source && result.text !== null) {
    for (const finding of renderPromptTemplatePreview(source, JSON.parse(result.text)).findings) {
      const codes = ["missingVariable", "inputType", "undeclaredInput", "inputSecret", "renderedBytes"] as const;
      issues.push({ code: codes.find(code => code === finding.messageKey) ?? "referenceInvalid", field: finding.field.replace(/^input\./u, "") });
    }
  }
  return { variables: parsed.variables, issues, isValid: parsed.isValid && !issues.length && (input.mode === "json" || source !== null) };
}

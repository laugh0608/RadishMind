import { structuredInputContainsSensitiveText, type StructuredRuntimeInputContract, type StructuredRuntimeInputField } from "./structuredRuntimeInput.ts";

export type WorkflowDraftInputContract = Omit<StructuredRuntimeInputContract, "contractDigest"> & { contractDigest?: string };
export type WorkflowDraftInputContractIssue = { field: string; code: "identity" | "summary" | "count" | "name" | "duplicate" | "type" | "required" | "label" | "description" | "sensitive" };

export function validateWorkflowDraftInputContract(contract: WorkflowDraftInputContract): WorkflowDraftInputContractIssue[] {
  const issues: WorkflowDraftInputContractIssue[] = [];
  const bytes = (value: string) => new TextEncoder().encode(value).length;
  if (!/^[A-Za-z0-9][A-Za-z0-9._:-]{0,159}$/u.test(contract.contractId)) issues.push({ field: "contractId", code: "identity" });
  if (bytes(contract.summary) > 4000) issues.push({ field: "summary", code: "summary" });
  if (structuredInputContainsSensitiveText(contract.contractId)) issues.push({ field: "contractId", code: "sensitive" });
  if (structuredInputContainsSensitiveText(contract.summary)) issues.push({ field: "summary", code: "sensitive" });
  if (contract.fields.length < 1 || contract.fields.length > 16) issues.push({ field: "fields", code: "count" });
  const names = new Set<string>();
  contract.fields.forEach((field, index) => {
    const add = (key: keyof StructuredRuntimeInputField, code: WorkflowDraftInputContractIssue["code"]) => issues.push({ field: `fields.${index}.${key}`, code });
    if (!/^[a-z][a-z0-9_]{0,63}$/u.test(field.name)) add("name", "name");
    if (names.has(field.name)) add("name", "duplicate");
    names.add(field.name);
    if (!["string", "integer", "number", "boolean"].includes(field.valueType)) add("valueType", "type");
    if (typeof field.required !== "boolean") add("required", "required");
    if (!field.label.trim() || bytes(field.label) > 160) add("label", "label");
    if (bytes(field.description) > 4000) add("description", "description");
    for (const key of ["name", "label", "description"] as const) {
      if (structuredInputContainsSensitiveText(field[key])) add(key, "sensitive");
    }
  });
  return issues;
}

export function workflowDraftInputContractDocument(contract: WorkflowDraftInputContract) {
  if (validateWorkflowDraftInputContract(contract).length) throw new Error("workflow_input_contract_invalid");
  return {
    contract_id: contract.contractId,
    fields: contract.fields.map((field) => ({ name: field.name, value_type: field.valueType, required: field.required, label: field.label, description: field.description })),
    summary: contract.summary,
  };
}

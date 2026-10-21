import { parseStructuredRuntimeInputContractDocument } from "./structuredRuntimeInput.ts";

// v2 uses the existing server document shape. Do not accept extension keys as
// permission to downgrade a structured contract to a v1 required-fields list.
export function isSavedWorkflowDraftV2Content(value: unknown): boolean {
  const keys = ["draft_id", "workspace_id", "application_id", "source_definition_id", "base_definition_version", "schema_version", "draft_status", "name", "description", "nodes", "edges", "input_contract", "output_contract", "provider_refs", "tool_refs", "rag_refs", "requested_capabilities", "draft_version", "lifecycle_state", "lifecycle_version", "archived_at", "library_updated_at", "lifecycle_updated_by_actor_ref", "provenance_kind", "created_at", "updated_at", "created_by_actor_ref", "updated_by_actor_ref", "validation_summary", "blocked_capability_summary", "request_audit_metadata", "sample_or_unsaved_draft_status"];
  if (!recordWithKeys(value, keys, ["additional_fields"]) || value.schema_version !== "saved_workflow_draft.v2" ||
    !parseStructuredRuntimeInputContractDocument(value.input_contract) || !legacyContract(value.output_contract) ||
    !Number.isInteger(value.base_definition_version) || Number(value.base_definition_version) < 0 ||
    !["draft_id", "workspace_id", "application_id", "source_definition_id", "name", "description", "created_at", "updated_at", "created_by_actor_ref", "updated_by_actor_ref", "lifecycle_updated_by_actor_ref", "sample_or_unsaved_draft_status"].every((key) => typeof value[key] === "string") ||
    !["provider_refs", "tool_refs", "rag_refs", "requested_capabilities"].every((key) => strings(value[key])) ||
    (value.additional_fields !== undefined && !record(value.additional_fields)) ||
    !Array.isArray(value.nodes) || value.nodes.length < 1 || value.nodes.length > 50 || !value.nodes.every(node) ||
    !Array.isArray(value.edges) || value.edges.length > 120 || !value.edges.every(edge)) return false;
  return true;
}

export function isLegacySavedWorkflowInputContract(value: unknown): boolean {
  return legacyContract(value);
}

function node(value: unknown): boolean {
  const textKeys = ["node_id", "node_type", "label", "input_summary", "output_summary", "input_contract_ref", "output_contract_ref", "output_mapping_summary", "provider_ref", "tool_ref", "rag_ref", "risk_level"];
  return recordWithKeys(value, [...textKeys, "input_contract_fields", "output_contract_fields", "requires_confirmation"]) &&
    textKeys.every((key) => typeof value[key] === "string") &&
    ["prompt", "llm", "condition", "output"].includes(String(value.node_type)) &&
    strings(value.input_contract_fields) && strings(value.output_contract_fields) && typeof value.requires_confirmation === "boolean";
}

function edge(value: unknown): boolean {
  const keys = ["edge_id", "from_node_id", "to_node_id", "condition_summary"];
  return recordWithKeys(value, keys) && keys.every((key) => typeof value[key] === "string");
}

function legacyContract(value: unknown): boolean {
  return recordWithKeys(value, ["contract_id", "required_fields", "summary"]) &&
    typeof value.contract_id === "string" && strings(value.required_fields) && typeof value.summary === "string";
}

function strings(value: unknown): value is string[] {
  return Array.isArray(value) && value.every((entry) => typeof entry === "string");
}

function record(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function recordWithKeys(value: unknown, required: string[], optional: string[] = []): value is Record<string, unknown> {
  return record(value) && required.every((key) => Object.hasOwn(value, key)) && Object.keys(value).every((key) => required.includes(key) || optional.includes(key));
}

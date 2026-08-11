import assert from "node:assert/strict";
import test from "node:test";

import {
  parseStructuredRuntimeInputContractDocument,
  validateStructuredRuntimeInputDrafts,
} from "../src/features/control-plane-read/structuredRuntimeInput.ts";

const contractDocument = {
  contract_id: "contract_customer_retry",
  fields: [
    { name: "customer_name", value_type: "string", required: true, label: "Customer", description: "Bounded customer label." },
    { name: "retry_count", value_type: "integer", required: true, label: "Retries", description: "Safe integer retry count." },
    { name: "threshold", value_type: "number", required: false, label: "Threshold", description: "Optional finite threshold." },
    { name: "dry_run", value_type: "boolean", required: true, label: "Dry run", description: "Explicit true or false." },
  ],
  summary: "Typed runtime input without defaults.",
  contract_digest: `sha256:${"a".repeat(64)}`,
};

test("structured runtime decoder accepts only the exact bounded contract", () => {
  const contract = parseStructuredRuntimeInputContractDocument(contractDocument);
  assert.equal(contract?.fields[1]?.valueType, "integer");
  assert.equal(parseStructuredRuntimeInputContractDocument({ ...contractDocument, extra: true }), null);
  assert.equal(parseStructuredRuntimeInputContractDocument({ ...contractDocument, fields: [...contractDocument.fields, contractDocument.fields[0]] }), null);
  assert.equal(parseStructuredRuntimeInputContractDocument({ ...contractDocument, summary: "token=forbidden" }), null);
});

test("structured runtime validation preserves types and requires an explicit boolean", () => {
  const contract = parseStructuredRuntimeInputContractDocument(contractDocument)!;
  const missing = validateStructuredRuntimeInputDrafts(contract, { customer_name: "Acme", retry_count: "2" });
  assert.equal(missing.failureCode, "workflow_input_required_field_missing");
  assert.equal(missing.fieldErrors.dry_run, "这是必填字段。");

  const valid = validateStructuredRuntimeInputDrafts(contract, {
    customer_name: "Acme",
    retry_count: "2",
    threshold: "0.75",
    dry_run: false,
  });
  assert.equal(valid.ok, true);
  assert.deepEqual(valid.inputs, { customer_name: "Acme", retry_count: 2, threshold: 0.75, dry_run: false });
});

test("structured runtime validation rejects unknown, unsafe, and mistyped values", () => {
  const contract = parseStructuredRuntimeInputContractDocument(contractDocument)!;
  assert.equal(validateStructuredRuntimeInputDrafts(contract, { customer_name: "Acme", retry_count: "2", dry_run: true, unknown: "x" }).failureCode, "workflow_input_unknown_field");
  assert.equal(validateStructuredRuntimeInputDrafts(contract, { customer_name: "password=hunter2", retry_count: "2", dry_run: true }).fieldErrors.customer_name?.includes("不得包含"), true);
  assert.equal(validateStructuredRuntimeInputDrafts(contract, { customer_name: "Acme", retry_count: "2.5", dry_run: true }).failureCode, "workflow_input_value_type_invalid");
  assert.equal(validateStructuredRuntimeInputDrafts(contract, { customer_name: "Acme", retry_count: "2", threshold: "0x10", dry_run: true }).failureCode, "workflow_input_value_type_invalid");
  assert.equal(validateStructuredRuntimeInputDrafts(contract, { customer_name: "Acme", retry_count: "2", threshold: "1e309", dry_run: true }).failureCode, "workflow_input_value_type_invalid");
});

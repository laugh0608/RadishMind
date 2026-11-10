import assert from "node:assert/strict";
import test from "node:test";
import { emptyPromptVariableInput, switchPromptVariableInput, validatePromptVariableInput, type PromptVariableInput } from "../src/features/control-plane-read/promptVariableInput.ts";
import type { PromptTemplateDraftInput } from "../src/features/control-plane-read/promptApplicationTemplateConsumer.ts";

const source: PromptTemplateDraftInput = {
  templateId: "ptpl_aaaaaaaaaaaaaaaa", workspaceId: "workspace_demo", applicationId: "app_aaaaaaaaaaaaaaaa", templateName: "Variable contract", description: "",
  messages: [{ role: "user", content: "{{question}} {{count}} {{ratio}} {{enabled}} {{items}} {{tone}}" }],
  variables: [
    { name: "question", type: "string", required: true, description: "" },
    { name: "count", type: "integer", required: false, description: "" },
    { name: "ratio", type: "number", required: false, description: "" },
    { name: "enabled", type: "boolean", required: false, description: "" },
    { name: "items", type: "string_list", required: false, description: "" },
    { name: "tone", type: "string", required: false, description: "", defaultValue: "clear" },
  ], outputContract: { kind: "text", allowEmpty: false, maxBytes: 1024 },
};

test("Prompt input has no sample values and preserves omission, false, empty strings, arrays and whitespace", () => {
  assert.deepEqual(emptyPromptVariableInput(), { mode: "form", values: {} });
  assert.equal(validatePromptVariableInput(emptyPromptVariableInput(), source).isValid, false);
  const values = { question: "  line\n", count: 0, ratio: 0.25, enabled: false, items: ["", "a\nb", "  "] };
  const form = switchPromptVariableInput({ mode: "json", text: JSON.stringify(values) }, "form");
  assert.deepEqual(form.issues, []);
  const result = validatePromptVariableInput(form.input, source);
  assert.equal(result.isValid, true);
  assert.deepEqual(result.variables, values);
  assert.equal(Object.hasOwn(result.variables, "tone"), false);
  const json = switchPromptVariableInput(form.input, "json");
  assert.equal(json.input.mode, "json");
  if (json.input.mode === "json") assert.deepEqual(JSON.parse(json.input.text), values);
  for (const value of ["", "  "]) {
    const input = switchPromptVariableInput({ mode: "json", text: JSON.stringify({ question: value, items: [] }) }, "form").input;
    assert.deepEqual(validatePromptVariableInput(input, source).variables, { question: value, items: [] });
  }
});

test("invalid JSON and numeric drafts cannot be silently rounded, normalized or replaced", () => {
  for (const text of ['{"question":"a","question":"b"}', '{"count":9007199254740993}', '{"ratio":0.10000000000000001}', '{"question":"\\ud800"}', '{']) {
    const input: PromptVariableInput = { mode: "json", text };
    const next = switchPromptVariableInput(input, "form");
    assert.equal(next.input, input);
    assert.ok(next.issues.length);
    assert.equal(validatePromptVariableInput(input, null).isValid, false);
  }
  for (const text of ["", "-", "1.", "9007199254740993", "0.10000000000000001"]) {
    const input: PromptVariableInput = { mode: "form", values: { count: { kind: "number", integer: true, text } } };
    assert.equal(switchPromptVariableInput(input, "json").input, input);
    assert.equal(validatePromptVariableInput(input, source).isValid, false);
  }
});

test("mismatched and undeclared original values survive mode switches and block form execution", () => {
  const text = '{"question":null,"count":"old","undeclared":{"keep":[1,2]}}';
  const form = switchPromptVariableInput({ mode: "json", text }, "form");
  assert.deepEqual(form.issues, []);
  assert.equal(validatePromptVariableInput(form.input, source).isValid, false);
  const json = switchPromptVariableInput(form.input, "json");
  assert.equal(json.input.mode, "json");
  if (json.input.mode === "json") assert.deepEqual(JSON.parse(json.input.text), JSON.parse(text));
});

test("advanced JSON preserves existing invocation validation without requiring source access", () => {
  assert.equal(validatePromptVariableInput({ mode: "json", text: '{"question":"valid"}' }, null).isValid, true);
  assert.equal(validatePromptVariableInput({ mode: "form", values: { question: { kind: "value", value: "valid" } } }, null).isValid, false);
  for (const text of ['[]', 'null', '{"question":null}', '{"items":[1]}', '{"question":"Bearer sensitive"}', JSON.stringify({ question: "x".repeat(65536) })]) {
    assert.equal(validatePromptVariableInput({ mode: "json", text }, null).isValid, false);
  }
  const mismatched = switchPromptVariableInput({ mode: "json", text: '{"question":"valid","extra":true}' }, "form").input;
  assert.equal(validatePromptVariableInput(mismatched, source).isValid, false);
});

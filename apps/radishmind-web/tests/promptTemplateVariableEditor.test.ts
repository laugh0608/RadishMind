import assert from "node:assert/strict";
import test from "node:test";
import { readFileSync } from "node:fs";
import { createPromptTemplateEditorVariables } from "../src/features/control-plane-read/promptTemplateVariableEditor.ts";
import { renderPromptTemplatePreview, validatePromptTemplateLocally, type PromptTemplateDraftInput, type PromptTemplateVariable } from "../src/features/control-plane-read/promptApplicationTemplateConsumer.ts";
import { parsePromptOutputSchema } from "../src/features/control-plane-read/promptTemplateOutputSchema.ts";

const variable = (name: string): PromptTemplateVariable => ({ name, type: "string", required: true, description: "" });

test("saved variable normalization preserves row identities after rename and reorder", () => {
  const rows = createPromptTemplateEditorVariables([variable("question"), variable("tone")]);
  const renamed = [{ ...rows[0]!, name: "user_goal" }, { ...rows[1]!, name: "error_log" }];
  const restored = createPromptTemplateEditorVariables([variable("error_log"), variable("user_goal")], renamed);
  assert.equal(restored[0]?.editorKey, rows[1]?.editorKey);
  assert.equal(restored[1]?.editorKey, rows[0]?.editorKey);
  assert.notEqual(rows[0]?.editorKey, rows[1]?.editorKey);
});

test("removed rows do not donate their keys to new rows or another restored draft", () => {
  const rows = createPromptTemplateEditorVariables([variable("error_log"), variable("user_goal")]);
  const next = createPromptTemplateEditorVariables([variable("user_goal"), variable("environment_context")], [rows[1]!]);
  assert.equal(next[0]?.editorKey, rows[1]?.editorKey);
  assert.ok(!rows.some(row => row.editorKey === next[1]?.editorKey));
  const anotherDraft = createPromptTemplateEditorVariables([variable("user_goal")]);
  assert.notEqual(anotherDraft[0]?.editorKey, rows[1]?.editorKey);
});

test("diagnostic editor rows preserve the real trial variable contract and missing-input rejection", () => {
  const readFixture = (name: string) => JSON.parse(readFileSync(new URL(`../../../docs/features/user-workspace/prompt-application-dev-test-usage-guide.parts/${name}.json`, import.meta.url), "utf8"));
  const source = readFixture("diagnostics-trial-source");
  const input: PromptTemplateDraftInput = {
    templateId: "ptpl_abcdefghijklmnop", workspaceId: "workspace_demo", applicationId: "app_abcdefghijklmnop",
    templateName: "Synthetic diagnostic template", description: "", messages: source.messages,
    variables: createPromptTemplateEditorVariables(source.variables),
    outputContract: { kind: "json_object", allowEmpty: false, maxBytes: 8192, jsonSchema: parsePromptOutputSchema(JSON.stringify(source.output_contract.json_schema)).schema },
  };
  assert.equal(validatePromptTemplateLocally(input).isValid, true);
  for (const sample of readFixture("diagnostics-trial-samples").samples) {
    const result = renderPromptTemplatePreview(input, sample.variables);
    assert.equal(result.status, sample.expected_render_failure_code ? "invalid" : "valid", sample.id);
    if (sample.expected_render_failure_code) assert.equal(result.findings[0]?.code, sample.expected_render_failure_code);
    else assert.ok(result.messages[1]?.content.includes(sample.variables.error_log));
  }
});

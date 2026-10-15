import assert from "node:assert/strict";
import test from "node:test";
import { readFileSync } from "node:fs";
import { parsePromptInputJSON, promptNumberIsLossless } from "../src/features/control-plane-read/promptInputJSON.ts";
import {
  addPromptPlanSample, loadPromptPlanEditor, movePromptPlanSample, newPromptPlanEditor, promptPlanIsDirty,
  removePromptPlanSample, switchPromptPlanMode, updatePromptPlanSample, validatePromptPlanEditor,
  type PromptPlanEditor,
} from "../src/features/control-plane-read/promptEvaluationPlanDraft.ts";
import { serializeApplicationEvaluationItems, type ApplicationEvaluationPlan, type ApplicationEvaluationPlanVersion } from "../src/features/control-plane-read/applicationEvaluationCampaignConsumer.ts";
import type { PromptTemplateDraftInput } from "../src/features/control-plane-read/promptApplicationTemplateConsumer.ts";

const template: PromptTemplateDraftInput = {
  templateId: "ptpl_aaaaaaaaaaaaaaaa", applicationId: "app_aaaaaaaaaaaaaaaa", workspaceId: "workspace_demo", templateName: "Typed template", description: "",
  variables: [
    { name: "question", type: "string", required: true, description: "" },
    { name: "count", type: "integer", required: false, description: "", defaultValue: 2 },
    { name: "weight", type: "number", required: false, description: "" },
    { name: "flag", type: "boolean", required: false, description: "", defaultValue: false },
    { name: "tags", type: "string_list", required: false, description: "" },
  ],
  messages: [{ role: "user", content: "{{question}} {{count}} {{weight}} {{flag}} {{tags}}" }],
  outputContract: { kind: "text", allowEmpty: false, maxBytes: 8192 },
};

function sample(): PromptPlanEditor {
  const initial = addPromptPlanSample(newPromptPlanEditor());
  return updatePromptPlanSample({ ...initial, name: "Test plan" }, initial.selectedKey, { name: "Test sample", variables: { question: { kind: "value", value: "question\n原文" } } });
}

test("Prompt input rejects duplicate keys, loss of number precision and deep data before conversion", () => {
  for (const raw of ['{"a":1,"a":2}', '{"a":1,"\\u0061":2}', '[{"a":{"same":1,"same":2}}]']) {
    const result = parsePromptInputJSON(raw);
    assert.equal(result.ok, false);
    if (!result.ok) assert.equal(result.issue.code, "duplicateProperty");
  }
  for (const text of ["9007199254740993", "9007199254740992", "0.10000000000000001", "1e400", "1e-400", "-0", "-", "01", ".5"]) {
    assert.equal(promptNumberIsLossless(text), false, text);
  }
  for (const text of ["9007199254740991", "-12.5", "1.00", "0", "1e-7", "1.2e3", "0.00100", "-0.1", "1.2300"]) {
    assert.equal(promptNumberIsLossless(text), true, text);
    assert.equal(parsePromptInputJSON(text).ok, true);
  }
  const bad = parsePromptInputJSON('[\n  9007199254740993\n]');
  assert.deepEqual(bad, { ok: false, issue: { code: "numberPrecision", line: 2, column: 3 } });
  const syntax = parsePromptInputJSON('{"key":}');
  assert.equal(syntax.ok, false);
  if (!syntax.ok) assert.equal(syntax.issue.code, "jsonSyntax");
  assert.equal(parsePromptInputJSON('['.repeat(65) + '0' + ']'.repeat(65)).ok, false);
  assert.equal(parsePromptInputJSON('"\\ud800"').ok, false);
  assert.equal(parsePromptInputJSON('"\\ud83e\\udd55"').ok, true);
  assert.deepEqual(parsePromptInputJSON('{"text":"a \\" \\n 9007199254740993", "values":[{"a":1},{"a":2}]}').ok, true);
});

test("new Prompt drafts are empty, not dirty and cannot save a hidden sample", () => {
  const state = newPromptPlanEditor();
  assert.equal(promptPlanIsDirty(state), false);
  assert.equal(validatePromptPlanEditor(state, template).draft, null);
  assert.deepEqual(validatePromptPlanEditor(state, template).issues.map(issue => issue.code), ["planName", "sampleCount"]);
  const json = switchPromptPlanMode(state, "json");
  assert.deepEqual(json.issues, []);
  assert.equal(switchPromptPlanMode(json.state, "form").state.content.mode, "form");
});

test("sample rename, copying, removal and reordering preserve editor identity and business order", () => {
  let state = sample();
  const first = state.selectedKey;
  state = updatePromptPlanSample(state, first, { itemKey: "custom_key", name: "Renamed" });
  state = addPromptPlanSample(state, first);
  const second = state.selectedKey;
  assert.notEqual(first, second);
  assert.equal(state.content.mode, "form");
  if (state.content.mode !== "form") return;
  assert.deepEqual(state.content.samples.map(row => row.itemKey), ["custom_key", "sample_1"]);
  assert.notEqual(state.content.samples[0]?.variables, state.content.samples[1]?.variables);
  state = movePromptPlanSample(state, second, -1);
  const draft = validatePromptPlanEditor(state, template).draft!;
  assert.deepEqual(draft.items.map(item => item.itemKey), ["sample_1", "custom_key"]);
  assert.equal(JSON.stringify(draft).includes("editorKey"), false);
  assert.equal(movePromptPlanSample(state, second, -1), state);
  assert.equal(movePromptPlanSample(state, "missing", 1), state);
  state = removePromptPlanSample(state, second);
  assert.equal(state.selectedKey, first);
  assert.equal(promptPlanIsDirty(state), true);
  for (let i = 1; i < 20; i += 1) state = addPromptPlanSample(state);
  assert.equal(addPromptPlanSample(state), state);
});

test("form and JSON preserve absent/default, false, empty string, lists, null and unknown originals", () => {
  let state = sample();
  state = updatePromptPlanSample(state, state.selectedKey, { variables: {
    question: { kind: "value", value: "" }, flag: { kind: "value", value: false },
    count: { kind: "number", text: "0", integer: true }, weight: { kind: "number", text: "0.10", integer: false },
    tags: { kind: "list", rows: [{ editorKey: "a", value: "" }, { editorKey: "b", value: "line\nnext" }] },
    old: { kind: "value", value: null },
  } });
  const original = validatePromptPlanEditor(state, template);
  assert.deepEqual(original.issues, []);
  assert.deepEqual(original.templateIssues.map(issue => issue.code), ["undeclaredInput"]);
  const json = switchPromptPlanMode(state, "json");
  assert.equal(json.issues.length, 0);
  assert.equal(json.state.content.mode, "json");
  assert.equal(addPromptPlanSample(json.state), json.state);
  assert.equal(updatePromptPlanSample(json.state, "x", {}), json.state);
  assert.equal(movePromptPlanSample(json.state, "x", 1), json.state);
  assert.equal(removePromptPlanSample(json.state, "x"), json.state);
  const restored = switchPromptPlanMode(json.state, "form").state;
  assert.equal(restored.selectedKey, state.selectedKey);
  assert.deepEqual(validatePromptPlanEditor(restored, template).draft, original.draft);
  assert.equal(switchPromptPlanMode(restored, "form").state, restored);
  const omitted = validatePromptPlanEditor(sample(), template);
  assert.equal(omitted.matched, true);
  assert.deepEqual(Object.keys(omitted.draft!.items[0]!.promptApplication!.variables), ["question"]);
});

test("incomplete numbers and invalid JSON preserve the active input and block serialization", () => {
  const initial = sample();
  for (const text of ["", "-", "1.", "1.5", "9007199254740993"]) {
    const state = updatePromptPlanSample(initial, initial.selectedKey, { variables: { count: { kind: "number", integer: true, text } } });
    assert.equal(switchPromptPlanMode(state, "json").state, state);
    assert.equal(validatePromptPlanEditor(state, template).draft, null);
    assert.equal(validatePromptPlanEditor(state, template).issues[0]?.code, "numberValue");
  }
  for (const text of ["{", '[{"name":"unsupported"}]', "[9007199254740993]"]) {
    const state: PromptPlanEditor = { ...initial, content: { mode: "json", text, identities: [] } };
    assert.equal(switchPromptPlanMode(state, "form").state, state);
    assert.equal(validatePromptPlanEditor(state, null).draft, null);
    assert.equal(promptPlanIsDirty(state), true);
  }
});

test("Prompt validation aligns item keys, Unicode names and variable budgets with the existing contract", () => {
  let state = sample();
  state = { ...state, name: "萝".repeat(120) };
  state = updatePromptPlanSample(state, state.selectedKey, { name: "🥕".repeat(120) });
  assert.ok(validatePromptPlanEditor(state, template).draft);
  state = updatePromptPlanSample(state, state.selectedKey, { itemKey: "Bad KEY", name: "x".repeat(121), variables: { "bad-name": { kind: "value", value: "<".repeat(12000) } } });
  assert.deepEqual(validatePromptPlanEditor(state, null).issues.map(issue => issue.code), ["sampleKey", "sampleName", "variableName", "variableBytes"]);
  state = addPromptPlanSample(sample());
  state = updatePromptPlanSample(state, state.selectedKey, { itemKey: "sample_1", name: "Duplicate" });
  assert.equal(validatePromptPlanEditor(state, null).issues[0]?.code, "duplicateKey");
  const json = switchPromptPlanMode(state, "json").state;
  assert.equal(switchPromptPlanMode(json, "form").state, json);
});

test("template mismatches remain valid structural fixtures in JSON and never drop originals", () => {
  const base = sample();
  const state = updatePromptPlanSample(base, base.selectedKey, { variables: { question: { kind: "value", value: null }, unknown: { kind: "value", value: { nested: [true] } } } });
  const validation = validatePromptPlanEditor(state, template);
  assert.ok(validation.draft);
  assert.equal(validation.matched, false);
  assert.deepEqual(validation.templateIssues.map(issue => issue.code), ["undeclaredInput", "inputType"]);
  assert.equal(validatePromptPlanEditor(sample(), null).matched, false);
  const missing = updatePromptPlanSample(base, base.selectedKey, { variables: {} });
  assert.equal(validatePromptPlanEditor(missing, template).templateIssues[0]?.code, "missingVariable");
});

test("five diagnostic trial samples keep their original fields and expected missing-input feedback", () => {
  const read = (name: string) => JSON.parse(readFileSync(new URL(`../../../docs/features/user-workspace/prompt-application-dev-test-usage-guide.parts/${name}.json`, import.meta.url), "utf8"));
  const source = read("diagnostics-trial-source");
  const trial: PromptTemplateDraftInput = { ...template, variables: source.variables, messages: source.messages };
  for (const fixture of read("diagnostics-trial-samples").samples) {
    const state = sample();
    const populated = updatePromptPlanSample(state, state.selectedKey, { variables: Object.fromEntries(Object.entries(fixture.variables).map(([name, value]) => [name, { kind: "value", value }])) } as Parameters<typeof updatePromptPlanSample>[2]);
    const result = validatePromptPlanEditor(populated, trial);
    assert.equal(result.matched, !fixture.expected_render_failure_code, fixture.id);
    assert.deepEqual(result.draft!.items[0]!.promptApplication!.variables, fixture.variables);
    if (fixture.expected_render_failure_code) assert.equal(result.templateIssues[0]?.field, "error_log");
  }
});

test("reference checks retain field-level size, secret and total rendered-input failures", () => {
  const initial = sample();
  for (const [variables, expected] of [
    [{ question: { kind: "value", value: "x".repeat(16 * 1024 + 1) } }, "inputType"],
    [{ question: { kind: "value", value: "Authorization: Bearer synthetic" } }, "inputSecret"],
    [{ question: { kind: "value", value: "ok" }, tags: { kind: "list", rows: Array.from({ length: 129 }, (_, i) => ({ editorKey: String(i), value: "" })) } }, "inputType"],
  ] as const) {
    const state = updatePromptPlanSample(initial, initial.selectedKey, { variables: structuredClone(variables) } as Parameters<typeof updatePromptPlanSample>[2]);
    assert.equal(validatePromptPlanEditor(state, template).templateIssues[0]?.code, expected);
  }
  const repeated = { ...template, variables: [template.variables[0]!], messages: [{ role: "user" as const, content: "{{question}}".repeat(9) }] };
  const state = updatePromptPlanSample(initial, initial.selectedKey, { variables: { question: { kind: "value", value: "x".repeat(16 * 1024) } } });
  assert.equal(validatePromptPlanEditor(state, repeated).templateIssues[0]?.code, "renderedBytes");
  assert.equal(validatePromptPlanEditor(initial, { ...template, messages: [] }).templateIssues[0]?.code, "referenceInvalid");
});

test("saved versions bind exact CAS, stay clean, preserve sample identities and reject unrelated plans", () => {
  const original = sample();
  const draft = validatePromptPlanEditor(original, template).draft!;
  const digest = `sha256:${"a".repeat(64)}`;
  const plan = { planId: "plan_a", recordVersion: 4, latestPlanVersion: 3, latestPlanDigest: digest } as ApplicationEvaluationPlan;
  const version = { planId: "plan_a", planVersion: 3, planDigest: digest, ...draft } as ApplicationEvaluationPlanVersion;
  const loaded = loadPromptPlanEditor(plan, version);
  assert.equal(promptPlanIsDirty(loaded), false);
  assert.equal(loaded.base?.recordVersion, 4);
  const renamed = updatePromptPlanSample(loaded, loaded.selectedKey, { name: "Updated" });
  const saved = loadPromptPlanEditor({ ...plan, recordVersion: 5, latestPlanVersion: 4 }, { ...version, planVersion: 4, items: validatePromptPlanEditor(renamed, template).draft!.items }, renamed);
  assert.equal(saved.selectedKey, loaded.selectedKey);
  assert.equal(promptPlanIsDirty(saved), false);
  assert.equal(saved.base?.recordVersion, 5);
  assert.throws(() => loadPromptPlanEditor({ ...plan, planId: "other" }, version));
  const json = serializeApplicationEvaluationItems(saved.content.mode === "form" ? validatePromptPlanEditor(saved, template).draft!.items : []);
  assert.equal(json.includes("recordVersion"), false);
});

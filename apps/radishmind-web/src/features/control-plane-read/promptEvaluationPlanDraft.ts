import {
  parseApplicationEvaluationPlanDraft, serializeApplicationEvaluationItems,
  type ApplicationEvaluationPlan, type ApplicationEvaluationPlanDraft, type ApplicationEvaluationPlanItem,
  type ApplicationEvaluationPlanVersion, type JsonValue,
} from "./applicationEvaluationCampaignConsumer.ts";
import { renderPromptTemplatePreview, type PromptTemplateDraftInput } from "./promptApplicationTemplateConsumer.ts";
import { parsePromptInputJSON, promptNumberIsLossless, type PromptInputJSONIssue } from "./promptInputJSON.ts";

export const PROMPT_PLAN_PROFILE = "prompt_application_invocation_v1" as const;
export type PromptPlanValue =
  | { kind: "value"; value: JsonValue }
  | { kind: "number"; text: string; integer: boolean }
  | { kind: "list"; rows: Array<{ editorKey: string; value: string }> };
export type PromptPlanSample = Pick<ApplicationEvaluationPlanItem, "itemKey" | "name" | "expectedClassification"> & {
  editorKey: string;
  variables: Record<string, PromptPlanValue>;
};
export type PromptPlanContent =
  | { mode: "form"; samples: PromptPlanSample[] }
  | { mode: "json"; text: string; identities: Array<{ itemKey: string; editorKey: string }> };
export type PromptPlanBase = { planId: string; planVersion: number; planDigest: string; recordVersion: number };
export type PromptPlanEditor = {
  name: string;
  content: PromptPlanContent;
  selectedKey: string;
  base: PromptPlanBase | null;
  baseline: string;
};
export type PromptPlanIssueCode = PromptInputJSONIssue["code"] | "planShape" | "planName" | "sampleCount" |
  "sampleName" | "sampleKey" | "duplicateKey" | "variableName" | "variableBytes" | "numberValue" |
  "missingVariable" | "inputType" | "undeclaredInput" | "inputSecret" | "renderedBytes" | "referenceInvalid";
export type PromptPlanIssue = { code: PromptPlanIssueCode; itemKey?: string; field?: string; line?: number; column?: number };
export type PromptPlanValidation = {
  draft: ApplicationEvaluationPlanDraft | null;
  issues: PromptPlanIssue[];
  templateIssues: PromptPlanIssue[];
  matched: boolean;
};

export function newPromptPlanEditor(): PromptPlanEditor {
  const state: PromptPlanEditor = { name: "", content: { mode: "form", samples: [] }, selectedKey: "", base: null, baseline: "" };
  return { ...state, baseline: promptPlanFingerprint(state) };
}

export function loadPromptPlanEditor(
  plan: ApplicationEvaluationPlan, version: ApplicationEvaluationPlanVersion, previous?: PromptPlanEditor,
): PromptPlanEditor {
  if (plan.planId !== version.planId || plan.latestPlanVersion !== version.planVersion ||
    plan.latestPlanDigest !== version.planDigest || version.executionProfile !== PROMPT_PLAN_PROFILE) {
    throw new Error("Prompt plan version does not match the selected record.");
  }
  const identities = previous && (!previous.base || previous.base.planId === plan.planId) ? sampleIdentities(previous.content) : [];
  const samples = version.items.map(item => sampleFromItem(item, identities));
  const state: PromptPlanEditor = {
    name: version.name, content: previous?.content.mode === "json"
      ? { mode: "json", text: serializeApplicationEvaluationItems(version.items), identities: samples.map(({ itemKey, editorKey }) => ({ itemKey, editorKey })) }
      : { mode: "form", samples },
    selectedKey: samples.some(item => item.editorKey === previous?.selectedKey) ? previous!.selectedKey : samples[0]?.editorKey ?? "",
    base: { planId: plan.planId, recordVersion: plan.recordVersion, planVersion: version.planVersion, planDigest: version.planDigest },
    baseline: "",
  };
  return { ...state, baseline: promptPlanFingerprint(state) };
}

export function promptPlanFingerprint(state: Pick<PromptPlanEditor, "name" | "content">): string {
  const result = contentItems(state.content);
  // An invalid edit is still dirty; a previously valid serialization is never used.
  const value = result.issues.length ? state.content : serializeApplicationEvaluationItems(result.items);
  return JSON.stringify([state.name, value]);
}

export function promptPlanIsDirty(state: PromptPlanEditor): boolean {
  return promptPlanFingerprint(state) !== state.baseline;
}

export function switchPromptPlanMode(state: PromptPlanEditor, mode: PromptPlanContent["mode"]):
  { state: PromptPlanEditor; issues: PromptPlanIssue[] } {
  if (state.content.mode === mode) return { state, issues: [] };
  const result = contentItems(state.content);
  if (result.issues.length) return { state, issues: result.issues };
  const identities = sampleIdentities(state.content);
  const content: PromptPlanContent = mode === "json"
    ? { mode, text: serializeApplicationEvaluationItems(result.items), identities }
    : { mode, samples: result.items.map(item => sampleFromItem(item, identities)) };
  const selectedKey = content.mode === "form" && !content.samples.some(sample => sample.editorKey === state.selectedKey)
    ? content.samples[0]?.editorKey ?? "" : state.selectedKey;
  return { state: { ...state, content, selectedKey }, issues: [] };
}

export function addPromptPlanSample(state: PromptPlanEditor, copyKey?: string): PromptPlanEditor {
  if (state.content.mode !== "form" || state.content.samples.length >= 20) return state;
  const source = state.content.samples.find(item => item.editorKey === copyKey);
  const existing = new Set(state.content.samples.map(item => item.itemKey));
  let index = 1;
  while (existing.has(`sample_${index}`)) index += 1;
  const sample: PromptPlanSample = {
    itemKey: `sample_${index}`, name: source?.name ?? "", expectedClassification: source?.expectedClassification ?? "unchanged",
    editorKey: crypto.randomUUID(), variables: source ? structuredClone(source.variables) : {},
  };
  return { ...state, content: { mode: "form", samples: [...state.content.samples, sample] }, selectedKey: sample.editorKey };
}

export function updatePromptPlanSample(state: PromptPlanEditor, editorKey: string, update: Partial<Pick<PromptPlanSample, "itemKey" | "name" | "expectedClassification" | "variables">>): PromptPlanEditor {
  if (state.content.mode !== "form") return state;
  return { ...state, content: { mode: "form", samples: state.content.samples.map(item => item.editorKey === editorKey ? { ...item, ...update } : item) } };
}

export function removePromptPlanSample(state: PromptPlanEditor, editorKey: string): PromptPlanEditor {
  if (state.content.mode !== "form") return state;
  const samples = state.content.samples.filter(item => item.editorKey !== editorKey);
  return { ...state, content: { mode: "form", samples }, selectedKey: state.selectedKey === editorKey ? samples[0]?.editorKey ?? "" : state.selectedKey };
}

export function movePromptPlanSample(state: PromptPlanEditor, editorKey: string, direction: -1 | 1): PromptPlanEditor {
  if (state.content.mode !== "form") return state;
  const samples = [...state.content.samples];
  const index = samples.findIndex(item => item.editorKey === editorKey);
  if (index < 0 || !samples[index + direction]) return state;
  [samples[index], samples[index + direction]] = [samples[index + direction], samples[index]];
  return { ...state, content: { mode: "form", samples } };
}

export function validatePromptPlanEditor(state: PromptPlanEditor, reference: PromptTemplateDraftInput | null): PromptPlanValidation {
  const { items, issues } = contentItems(state.content);
  const readableItems = issues.length === 0;
  if ([...state.name.trim()].length < 2 || [...state.name.trim()].length > 120) issues.push({ code: "planName" });
  if (readableItems && (!items.length || items.length > 20)) issues.push({ code: "sampleCount" });
  const keys = new Set<string>();
  for (const item of items) {
    const itemKey = item.itemKey;
    if (!/^[a-z][a-z0-9_-]{0,63}$/u.test(itemKey)) issues.push({ code: "sampleKey", itemKey, field: "item_key" });
    if (keys.has(itemKey)) issues.push({ code: "duplicateKey", itemKey, field: "item_key" });
    keys.add(itemKey);
    if ([...item.name.trim()].length < 1 || [...item.name.trim()].length > 120) issues.push({ code: "sampleName", itemKey, field: "name" });
    const variables = item.promptApplication!.variables;
    for (const name of Object.keys(variables)) {
      if (!/^[A-Za-z][A-Za-z0-9_]{0,63}$/u.test(name)) issues.push({ code: "variableName", itemKey, field: name });
    }
    // Go's canonical JSON escapes HTML and the two line-separator characters.
    const canonical = JSON.stringify(variables).replace(/[<>&\u2028\u2029]/gu, char => `\\u${char.charCodeAt(0).toString(16).padStart(4, "0")}`);
    if (new TextEncoder().encode(canonical).byteLength > 64 * 1024) issues.push({ code: "variableBytes", itemKey });
  }
  const templateIssues: PromptPlanIssue[] = [];
  if (reference) for (const item of items) {
    const preview = renderPromptTemplatePreview(reference, item.promptApplication!.variables);
    for (const finding of preview.findings) {
      const keys = ["missingVariable", "inputType", "undeclaredInput", "inputSecret", "renderedBytes"] as const;
      const code = keys.find(key => key === finding.messageKey) ?? "referenceInvalid";
      templateIssues.push({ code, itemKey: item.itemKey, field: finding.field.replace(/^input\./u, "") });
    }
  }
  let draft: ApplicationEvaluationPlanDraft | null = null;
  if (!issues.length) {
    try { draft = parseApplicationEvaluationPlanDraft(state.name, PROMPT_PLAN_PROFILE, '{"workflow_definition":null}', serializeApplicationEvaluationItems(items)); }
    catch { issues.push({ code: "planShape" }); }
  }
  return { draft, issues, templateIssues, matched: reference !== null && items.length > 0 && !issues.length && !templateIssues.length };
}

function contentItems(content: PromptPlanContent): { items: ApplicationEvaluationPlanItem[]; issues: PromptPlanIssue[] } {
  if (content.mode === "json") {
    const parsed = parsePromptInputJSON(content.text);
    if (!parsed.ok) return { items: [], issues: [parsed.issue] };
    if (Array.isArray(parsed.value) && !parsed.value.length) return { items: [], issues: [] };
    try {
      const draft = parseApplicationEvaluationPlanDraft("Prompt plan", PROMPT_PLAN_PROFILE, '{"workflow_definition":null}', content.text);
      const keys = draft.items.map(item => item.itemKey);
      if (new Set(keys).size !== keys.length) return { items: [], issues: [{ code: "duplicateKey" }] };
      return { items: draft.items, issues: [] };
    } catch { return { items: [], issues: [{ code: "planShape" }] }; }
  }
  const issues: PromptPlanIssue[] = [];
  const items = content.samples.map(item => {
    const entries: Array<[string, JsonValue]> = [];
    for (const [name, value] of Object.entries(item.variables)) {
      if (value.kind === "number") {
        if (!promptNumberIsLossless(value.text) || value.integer && !Number.isSafeInteger(Number(value.text))) {
          issues.push({ code: "numberValue", itemKey: item.itemKey, field: name });
          continue;
        }
        entries.push([name, Number(value.text)]);
      } else entries.push([name, value.kind === "list" ? value.rows.map(row => row.value) : value.value]);
    }
    return { itemKey: item.itemKey, name: item.name, expectedClassification: item.expectedClassification,
      workflowDefinition: null, applicationRAG: null, agentCopilot: null, promptApplication: { variables: Object.fromEntries(entries) } };
  });
  return { items, issues };
}

function sampleIdentities(content: PromptPlanContent): Array<{ itemKey: string; editorKey: string }> {
  return content.mode === "json" ? content.identities : content.samples.map(({ itemKey, editorKey }) => ({ itemKey, editorKey }));
}

function sampleFromItem(item: ApplicationEvaluationPlanItem, identities: Array<{ itemKey: string; editorKey: string }>): PromptPlanSample {
  if (!item.promptApplication) throw new Error("Expected a Prompt fixture.");
  return { itemKey: item.itemKey, name: item.name, expectedClassification: item.expectedClassification,
    editorKey: identities.find(identity => identity.itemKey === item.itemKey)?.editorKey ?? crypto.randomUUID(),
    variables: Object.fromEntries(Object.entries(item.promptApplication.variables).map(([name, value]) => [name,
      Array.isArray(value) && value.every(entry => typeof entry === "string")
        ? { kind: "list", rows: value.map(entry => ({ editorKey: crypto.randomUUID(), value: entry })) }
        : { kind: "value", value },
    ])),
  };
}

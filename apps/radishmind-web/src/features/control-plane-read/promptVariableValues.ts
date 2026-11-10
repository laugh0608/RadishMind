import { promptNumberIsLossless } from "./promptInputJSON.ts";

export type PromptVariableJSON = string | number | boolean | null | PromptVariableJSON[] | { [key: string]: PromptVariableJSON };
export type PromptVariableValue =
  | { kind: "value"; value: PromptVariableJSON }
  | { kind: "number"; text: string; integer: boolean }
  | { kind: "list"; rows: Array<{ editorKey: string; value: string }> };
export type PromptVariableValues = Record<string, PromptVariableValue>;

export function editPromptVariableValues(values: Record<string, PromptVariableJSON>): PromptVariableValues {
  return Object.fromEntries(Object.entries(values).map(([name, value]) => [name,
    Array.isArray(value) && value.every(entry => typeof entry === "string")
      ? { kind: "list", rows: value.map(entry => ({ editorKey: crypto.randomUUID(), value: entry })) }
      : { kind: "value", value },
  ]));
}

export function serializePromptVariableValues(values: PromptVariableValues): {
  values: Record<string, PromptVariableJSON>; invalidNumbers: string[];
} {
  const invalidNumbers: string[] = [];
  const entries: Array<[string, PromptVariableJSON]> = [];
  for (const [name, value] of Object.entries(values)) {
    if (value.kind === "number") {
      if (!promptNumberIsLossless(value.text) || value.integer && !Number.isSafeInteger(Number(value.text))) {
        invalidNumbers.push(name);
        continue;
      }
      entries.push([name, Number(value.text)]);
    } else entries.push([name, value.kind === "list" ? value.rows.map(row => row.value) : value.value]);
  }
  return { values: Object.fromEntries(entries), invalidNumbers };
}

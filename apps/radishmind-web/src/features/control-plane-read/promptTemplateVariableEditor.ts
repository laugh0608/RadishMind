import type { PromptTemplateVariable } from "./promptApplicationTemplateConsumer.ts";

export type PromptTemplateEditorVariable = PromptTemplateVariable & { editorKey: string };

// Keys belong to the mounted editor, never to the template API or its digest.
// A saved draft may normalize the order; keep each existing row's DOM identity.
export function createPromptTemplateEditorVariables(
  variables: PromptTemplateVariable[],
  previous: PromptTemplateEditorVariable[] = [],
): PromptTemplateEditorVariable[] {
  return variables.map(variable => ({
    ...variable,
    editorKey: previous.find(row => row.name === variable.name)?.editorKey ?? globalThis.crypto.randomUUID(),
  }));
}

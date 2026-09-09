// The editor has one result slot shared by open, read, validate and save.
// A context change must invalidate requests even when the user returns to it.
export function createWorkflowDraftEditorRequests() {
  let scope = "";
  let generation = 0;
  return {
    setScope(nextScope: string) {
      if (scope === nextScope) return;
      scope = nextScope;
      generation += 1;
    },
    invalidate() {
      generation += 1;
    },
    begin() {
      const requestGeneration = ++generation;
      return () => requestGeneration === generation;
    },
  };
}

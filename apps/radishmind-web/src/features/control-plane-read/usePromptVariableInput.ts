import { useEffect, useMemo, useRef, useState } from "react";
import { readPromptApplicationRuntimeConfig } from "./promptApplicationRuntimeConsumer.ts";
import { readPromptTemplateConfig } from "./promptApplicationTemplateConsumer.ts";
import { readPromptTemplateReference, type PromptTemplateReference } from "./promptTemplateReference.ts";
import type { PromptApplicationSession } from "./promptApplicationSessionConsumer.ts";
import { emptyPromptVariableInput, switchPromptVariableInput, validatePromptVariableInput, type PromptVariableInput, type PromptVariableIssue } from "./promptVariableInput.ts";

type ReferenceState = {
  status: "idle" | "loading" | "ready" | "changed" | "unavailable";
  current: PromptTemplateReference | null; candidate: PromptTemplateReference | null; code: string;
};
const emptyReference = (): ReferenceState => ({ status: "idle", current: null, candidate: null, code: "" });

export function usePromptVariableInput(applicationId: string, workspaceId: string, session?: PromptApplicationSession | null) {
  const [input, setInput] = useState(emptyPromptVariableInput);
  const [reference, setReference] = useState(emptyReference);
  const [modeIssues, setModeIssues] = useState<PromptVariableIssue[]>([]);
  const [busy, setBusy] = useState(false);
  const generation = useRef(0);
  const inFlight = useRef(false);
  const scopeKey = JSON.stringify([applicationId, workspaceId, session === undefined ? "invocation" : session?.sessionId ?? "no-session",
    session?.assignmentId, session?.assignmentVersion, session?.templateId, session?.templateVersion]);
  const configs = useMemo(() => ({
    runtime: { ...readPromptApplicationRuntimeConfig(), workspaceId },
    template: { ...readPromptTemplateConfig(), workspaceId },
  }), [workspaceId]);
  const source = reference.status === "ready" ? reference.current?.source ?? null : null;
  const validation = validatePromptVariableInput(input, source);

  useEffect(() => {
    reset();
    return () => { generation.current += 1; inFlight.current = false; };
  }, [scopeKey]);

  function reset() {
    generation.current += 1;
    inFlight.current = false;
    setBusy(false);
    setInput(emptyPromptVariableInput());
    setReference(emptyReference());
    setModeIssues([]);
  }
  function change(next: PromptVariableInput) {
    if (inFlight.current) return;
    setInput(next);
    setModeIssues([]);
  }
  function changeMode(mode: PromptVariableInput["mode"]) {
    if (inFlight.current) return;
    const next = switchPromptVariableInput(input, mode);
    setInput(next.input);
    setModeIssues(next.issues);
  }
  async function refresh() {
    if (inFlight.current) return null;
    const request = ++generation.current;
    if (session === null) {
      setReference({ ...emptyReference(), status: "unavailable", code: "prompt_input_session_required" });
      return null;
    }
    inFlight.current = true;
    setBusy(true);
    setReference(current => ({ ...current, status: "loading", candidate: null, code: "" }));
    try {
      const result = await readPromptTemplateReference(configs.runtime, configs.template, applicationId, () => request === generation.current, session);
      if (request !== generation.current || result.status === "cancelled") return null;
      if (result.status === "unavailable") {
        setReference({ ...emptyReference(), current: reference.current, status: "unavailable", code: result.code });
        return null;
      }
      if (reference.current && reference.current.key !== result.reference.key) {
        setReference({ status: "changed", current: reference.current, candidate: result.reference, code: "" });
        return null;
      }
      setReference({ status: "ready", current: result.reference, candidate: null, code: "" });
      return result.reference;
    } catch {
      if (request === generation.current) setReference({ ...emptyReference(), current: reference.current, status: "unavailable", code: "prompt_input_reference_failed" });
      return null;
    } finally {
      if (request === generation.current) { inFlight.current = false; setBusy(false); }
    }
  }
  function adopt() {
    if (inFlight.current || session !== undefined || !reference.candidate) return;
    setReference({ status: "ready", current: reference.candidate, candidate: null, code: "" });
  }
  async function prepare() {
    if (inFlight.current || !validation.isValid) return null;
    if (input.mode === "json") return validation.variables;
    const next = await refresh();
    if (!next) return null;
    const checked = validatePromptVariableInput(input, next.source);
    return checked.isValid ? checked.variables : null;
  }
  return { input, reference, busy, validation, issues: [...modeIssues, ...validation.issues], change, changeMode, refresh, adopt, prepare, reset };
}

import { useEffect, useMemo, useRef, useState } from "react";
import {
  archiveApplicationEvaluationPlan, createApplicationEvaluationPlan, reviseApplicationEvaluationPlan,
  type ApplicationEvaluationConfig, type ApplicationEvaluationPlan, type ApplicationEvaluationPlanVersion,
} from "./applicationEvaluationCampaignConsumer.ts";
import { readPromptApplicationRuntimeConfig } from "./promptApplicationRuntimeConsumer.ts";
import { readPromptTemplateConfig } from "./promptApplicationTemplateConsumer.ts";
import { loadPromptPlanEditor, newPromptPlanEditor, promptPlanIsDirty, switchPromptPlanMode, validatePromptPlanEditor, type PromptPlanEditor, type PromptPlanIssue } from "./promptEvaluationPlanDraft.ts";
import { readPromptTemplateReference, type PromptTemplateReference } from "./promptTemplateReference.ts";
import { acceptPromptPlanSave, promptPlanReviewIsCurrent, reviewPromptPlan, type PromptPlanReview } from "./promptEvaluationPlanReview.ts";

type ReferenceState = { status: "idle" | "loading" | "ready" | "changed" | "unavailable"; current: PromptTemplateReference | null; candidate: PromptTemplateReference | null; code: string };
type Confirmation = { kind: "save"; review: PromptPlanReview } | { kind: "archive"; plan: ApplicationEvaluationPlan };
type Notice = { kind: "saved" | "loaded" | "archived"; id: string; version?: number } | { kind: "failed"; code: string };
type Discard = { kind: "new" } | { kind: "load"; plan: ApplicationEvaluationPlan; version: ApplicationEvaluationPlanVersion };

export function usePromptEvaluationPlanEditor({ config, applicationActive, onSaved, onBusyChange, onDirtyChange }: {
  config: ApplicationEvaluationConfig;
  applicationActive: boolean;
  onSaved: (planId: string) => Promise<void>;
  onBusyChange: (busy: boolean) => void;
  onDirtyChange: (dirty: boolean) => void;
}) {
  const [editor, setEditor] = useState(newPromptPlanEditor);
  const [reference, setReference] = useState<ReferenceState>({ status: "idle", current: null, candidate: null, code: "" });
  const [confirmation, setConfirmation] = useState<Confirmation | null>(null);
  const [notice, setNotice] = useState<Notice | null>(null);
  const [discard, setDiscard] = useState<Discard | null>(null);
  const [conflict, setConflict] = useState<{ expected: number; current: number } | null>(null);
  const [busy, setBusy] = useState(false);
  const [showIssues, setShowIssues] = useState(false);
  const [modeIssues, setModeIssues] = useState<PromptPlanIssue[]>([]);
  const generation = useRef(0);
  const inFlight = useRef(false);
  const configs = useMemo(() => {
    const scope = { tenantRef: config.tenantRef, workspaceId: config.workspaceId, subjectRef: config.subjectRef, authMode: config.authMode };
    return { runtime: { ...readPromptApplicationRuntimeConfig(), ...scope }, template: { ...readPromptTemplateConfig(), ...scope } };
  }, [config]);
  const activeReference = reference.status === "ready" ? reference.current : null;
  const validation = useMemo(() => validatePromptPlanEditor(editor, activeReference?.source ?? null), [editor, activeReference]);
  const dirty = promptPlanIsDirty(editor);
  const writeBlocked = busy || config.mode === "offline" || !applicationActive || conflict !== null ||
    notice?.kind === "failed" && notice.code === "application_evaluation_authority_changed";

  useEffect(() => { onDirtyChange(dirty); }, [dirty, onDirtyChange]);
  useEffect(() => { onBusyChange(busy); }, [busy, onBusyChange]);
  useEffect(() => {
    if (!dirty) return;
    const warn = (event: BeforeUnloadEvent) => { event.preventDefault(); event.returnValue = ""; };
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [dirty]);
  useEffect(() => () => {
    generation.current += 1;
    onDirtyChange(false);
    onBusyChange(false);
  }, [onDirtyChange, onBusyChange]);

  function change(next: PromptPlanEditor) {
    if (inFlight.current) return;
    setEditor(next);
    setConfirmation(null);
    setNotice(null);
    setModeIssues([]);
  }

  function changeMode(mode: "form" | "json") {
    if (inFlight.current) return;
    const next = switchPromptPlanMode(editor, mode);
    change(next.state);
    setModeIssues(next.issues);
    setShowIssues(next.issues.length > 0);
  }

  function replaceDraft(action: Discard) {
    if (inFlight.current) return;
    if (dirty) { setDiscard(action); setConfirmation(null); return; }
    applyDiscard(action);
  }

  function applyDiscard(action: Discard) {
    try {
      const next = action.kind === "new" ? newPromptPlanEditor() : loadPromptPlanEditor(action.plan, action.version);
      setEditor(next);
      setConfirmation(null);
      setDiscard(null);
      setConflict(null);
      setModeIssues([]);
      setShowIssues(false);
      setNotice(action.kind === "new" ? null : { kind: "loaded", id: action.plan.planId, version: action.version.planVersion });
    } catch { setNotice({ kind: "failed", code: "prompt_plan_version_mismatch" }); }
  }

  async function refreshReference(forReview = false) {
    if (inFlight.current) return;
    inFlight.current = true;
    const request = ++generation.current;
    setBusy(true);
    setConfirmation(null);
    setNotice(null);
    setReference(current => ({ ...current, status: "loading", candidate: null, code: "" }));
    try {
      const result = await readPromptTemplateReference(configs.runtime, configs.template, config.applicationId, () => request === generation.current);
      if (request !== generation.current || result.status === "cancelled") return;
      if (result.status === "unavailable") {
        setReference({ status: "unavailable", current: reference.current, candidate: null, code: result.code });
        if (forReview && editor.content.mode === "json") showReview(null);
        return;
      }
      if (reference.current && reference.current.key !== result.reference.key) {
        setReference({ status: "changed", current: reference.current, candidate: result.reference, code: "" });
        return;
      }
      setReference({ status: "ready", current: result.reference, candidate: null, code: "" });
      if (forReview) showReview(result.reference);
    } catch {
      if (request === generation.current) setReference({ status: "unavailable", current: reference.current, candidate: null, code: "prompt_plan_reference_failed" });
    } finally {
      if (request === generation.current) { inFlight.current = false; setBusy(false); }
    }
  }

  function adoptReference() {
    if (!reference.candidate || inFlight.current) return;
    setReference({ status: "ready", current: reference.candidate, candidate: null, code: "" });
    setConfirmation(null);
    setShowIssues(true);
  }

  function showReview(source: PromptTemplateReference | null) {
    setShowIssues(true);
    const review = reviewPromptPlan(editor, source);
    setConfirmation(review ? { kind: "save", review } : null);
  }

  function reviewSave() {
    if (writeBlocked) return;
    setModeIssues([]);
    setShowIssues(true);
    if (!validation.draft) return;
    if (editor.content.mode === "form" || reference.status === "ready") void refreshReference(true);
    else showReview(null);
  }

  async function confirm() {
    if (!confirmation || writeBlocked || inFlight.current) return;
    if (confirmation.kind === "save" && !promptPlanReviewIsCurrent(editor, confirmation.review, activeReference)) {
      setConfirmation(null);
      return;
    }
    inFlight.current = true;
    const request = ++generation.current;
    setBusy(true);
    setNotice(null);
    try {
      const review = confirmation.kind === "save" ? confirmation.review : null;
      const envelope = confirmation.kind === "archive"
        ? await archiveApplicationEvaluationPlan(config, confirmation.plan.planId, confirmation.plan.recordVersion)
        : review!.base
          ? await reviseApplicationEvaluationPlan(config, review!.base.planId, review!.base.recordVersion, review!.draft)
          : await createApplicationEvaluationPlan(config, review!.draft);
      if (request !== generation.current) return;
      setConfirmation(null);
      if (envelope.failureCode) {
        setNotice({ kind: "failed", code: envelope.failureCode });
        if (envelope.failureCode === "application_evaluation_version_conflict") {
          setConflict({ expected: confirmation.kind === "archive" ? confirmation.plan.recordVersion : review!.base?.recordVersion ?? 0, current: envelope.currentRecordVersion });
        }
        return;
      }
      if (confirmation.kind === "archive") {
        await onSaved("");
        if (request !== generation.current) return;
        setEditor(newPromptPlanEditor());
        setNotice({ kind: "archived", id: confirmation.plan.planId });
      } else {
        if (!envelope.plan || !envelope.version) throw new Error("Missing saved version.");
        const next = acceptPromptPlanSave(editor, review!, envelope.plan, envelope.version);
        setEditor(next);
        await onSaved(envelope.plan.planId);
        if (request !== generation.current) return;
        setNotice({ kind: "saved", id: envelope.plan.planId, version: envelope.version.planVersion });
      }
    } catch {
      if (request === generation.current) { setConfirmation(null); setNotice({ kind: "failed", code: "prompt_plan_request_failed" }); }
    } finally {
      if (request === generation.current) { inFlight.current = false; setBusy(false); }
    }
  }

  return {
    editor, reference, confirmation, notice, discard, conflict, busy, dirty, validation, writeBlocked,
    issues: modeIssues.length ? modeIssues : showIssues ? [...validation.issues, ...validation.templateIssues] : [],
    change, changeMode, refreshReference, adoptReference, reviewSave, confirm, replaceDraft,
    cancelConfirmation: () => { if (!inFlight.current) setConfirmation(null); },
    cancelDiscard: () => setDiscard(null),
    confirmDiscard: () => { if (discard && !inFlight.current) applyDiscard(discard); },
    reviewArchive: (plan: ApplicationEvaluationPlan) => { if (!writeBlocked && !dirty) setConfirmation({ kind: "archive", plan }); },
  };
}

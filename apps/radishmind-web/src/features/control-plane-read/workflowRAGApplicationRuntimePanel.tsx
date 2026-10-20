import { useTranslation } from "react-i18next";
import "../../i18n/workflowRAGApplicationResources.ts";
import { workflowRAGApplicationStatus, workflowRAGAssignmentFeedback, workflowRAGInvocationFeedback } from "./workflowRAGApplicationMessages.ts";
import { useEffect, useMemo, useRef, useState } from "react";

import {
  decideWorkflowRAGApplicationRuntimeAssignment,
  initialWorkflowRAGApplicationInvocationResult,
  initialWorkflowRAGApplicationRuntimeResult,
  invokeWorkflowRAGApplication,
  readWorkflowRAGApplicationRuntimeAssignment,
  readWorkflowRAGApplicationRuntimeConfig,
  type WorkflowRAGApplicationInvocationResult,
  type WorkflowRAGApplicationRuntimeResult,
} from "./workflowRAGApplicationRuntimeConsumer.ts";
import {
  WORKFLOW_RAG_APPLICATION_CREDENTIAL_HANDOFF_EVENT,
  type WorkflowRAGApplicationCredentialHandoffDetail,
} from "./workflowRAGApplicationRuntimeEvents.ts";
import type { ApplicationDevelopmentOwnerEvidence } from "./applicationDevelopmentReadiness.ts";

const config = readWorkflowRAGApplicationRuntimeConfig();

export function WorkflowRAGRuntimeAssignmentPanel({
  applicationId,
  publishCandidateId,
  candidateApproved,
  readOnly = false,
  onEvidenceChange,
}: {
  applicationId: string;
  publishCandidateId: string;
  candidateApproved: boolean;
  readOnly?: boolean;
  onEvidenceChange?: (evidence: ApplicationDevelopmentOwnerEvidence) => void;
}) {
  const { t } = useTranslation("workflow");
  const [runtime, setRuntime] = useState<WorkflowRAGApplicationRuntimeResult>(() => initialWorkflowRAGApplicationRuntimeResult(config));
  const [reason, setReason] = useState("");
  const [busy, setBusy] = useState<"" | "read" | "decision">("");
  const generation = useRef(0);

  useEffect(() => {
    generation.current += 1;
    setRuntime(initialWorkflowRAGApplicationRuntimeResult(config));
    setReason("");
    setBusy("");
  }, [applicationId, publishCandidateId]);

  const proposedDecision = useMemo(() => {
    if (!runtime.assignment) return "activate" as const;
    if (runtime.assignment.state === "active" && runtime.assignment.publishCandidateId === publishCandidateId) return "revoke" as const;
    return "replace" as const;
  }, [publishCandidateId, runtime.assignment]);
  const validReason = reason.trim().length >= 4 && reason.trim().length <= 500;
  const decisionAllowed = candidateApproved && !readOnly && config.mode === "dev_workflow_rag_application_runtime_http" &&
    validReason && !(proposedDecision === "replace" && runtime.assignment?.publishCandidateId === publishCandidateId);

  useEffect(() => {
    if (!onEvidenceChange) return;
    const assignment = runtime.assignment;
    const ownerFailed = runtime.status === "failed" || runtime.status === "version_conflict";
    const active = runtime.status === "ready" && assignment?.state === "active";
    onEvidenceChange({
      contributionId: "rag_assignment",
      status: ownerFailed ? "blocked" : active ? "available" : "incomplete",
      coverage: assignment || ownerFailed ? "complete" : "none",
      evidenceRefs: assignment ? [{ kind: "assignment", id: assignment.assignmentId, version: assignment.recordVersion }] : [],
      missingEvidence: active ? [] : ["Load and activate the exact Application RAG assignment."],
      blockers: ownerFailed ? [{ code: runtime.failureCode || "rag_assignment_blocked", summary: runtime.summary }] : [],
      failureCodes: ownerFailed && runtime.failureCode ? [runtime.failureCode] : [],
    });
  }, [onEvidenceChange, runtime]);

  async function loadAssignment(preserveReason = false) {
    const current = ++generation.current;
    setBusy("read");
    const next = await readWorkflowRAGApplicationRuntimeAssignment(config, applicationId);
    if (generation.current !== current) return;
    setBusy("");
    setRuntime(next);
    if (!preserveReason) setReason("");
  }

  async function decide() {
    if (!decisionAllowed) return;
    const current = ++generation.current;
    setBusy("decision");
    const next = await decideWorkflowRAGApplicationRuntimeAssignment(config, {
      applicationId,
      expectedRecordVersion: runtime.assignment?.recordVersion ?? runtime.currentRecordVersion,
      decision: proposedDecision,
      publishCandidateId: proposedDecision === "revoke" ? "" : publishCandidateId,
      reason,
    });
    if (generation.current !== current) return;
    setBusy("");
    setRuntime(next);
    if (next.status === "ready") setReason("");
  }

  return (
    <article className="workflow-rag-runtime-assignment" aria-labelledby="workflow-rag-runtime-assignment-title">
      <div className="application-api-card-heading">
        <div><p className="eyebrow">{t($ => $.ragApplication.assignment.eyebrow)}</p><h5 id="workflow-rag-runtime-assignment-title">{t($ => $.ragApplication.assignment.title)}</h5></div>
        <span className={`status-badge ${runtime.assignment?.state === "active" ? "good" : runtime.status === "failed" ? "bad" : "neutral"}`}>
          {workflowRAGApplicationStatus(t, busy === "read" ? "loading" : busy === "decision" ? "recording" : runtime.assignment?.state ?? (!runtime.failureCode ? "idle" : runtime.status))}
        </span>
      </div>
      <p className="boundary-note">{t($ => $.ragApplication.assignment.boundary)}</p>
      <button type="button" onClick={() => void loadAssignment()} disabled={busy !== "" || config.mode === "offline"}>
        {busy === "read" ? t($ => $.ragApplication.assignment.loading) : t($ => $.ragApplication.assignment.load)}
      </button>
      {runtime.assignment ? (
        <dl className="workflow-rag-runtime-metadata">
          <div><dt>{t($ => $.ragApplication.assignment.assignment)}</dt><dd>{runtime.assignment.assignmentId} · v{runtime.assignment.recordVersion}</dd></div>
          <div><dt>{t($ => $.ragApplication.assignment.candidate)}</dt><dd>{runtime.assignment.publishCandidateId} · {t($ => $.ragApplication.assignment.reviewVersion, { version: runtime.assignment.publishReviewVersion })}</dd></div>
          <div><dt>{t($ => $.ragApplication.assignment.draft)}</dt><dd>{runtime.assignment.draftId} · v{runtime.assignment.draftVersion}</dd></div>
          <div><dt>{t($ => $.ragApplication.assignment.binding)}</dt><dd>{runtime.assignment.bindingRef.bindingId} · v{runtime.assignment.bindingRef.bindingVersion}</dd></div>
          <div><dt>{t($ => $.ragApplication.assignment.updated)}</dt><dd>{runtime.assignment.updatedAt} · {runtime.assignment.updatedByActorRef}</dd></div>
        </dl>
      ) : null}
      {!readOnly ? (
        <div className="workflow-rag-runtime-decision">
          <label>{t($ => $.ragApplication.assignment.reason)}<textarea value={reason} onChange={(event) => setReason(event.target.value)} rows={3} maxLength={500} placeholder={t($ => $.ragApplication.assignment.reasonPlaceholder)} /></label>
          {reason && !validReason ? <p className="failure-summary">{t($ => $.ragApplication.assignment.invalidReason)}</p> : null}
          <button type="button" onClick={() => void decide()} disabled={!decisionAllowed || busy !== ""}>
            {busy === "decision" ? t($ => $.ragApplication.assignment.recording) : proposedDecision === "activate" ? t($ => $.ragApplication.assignment.activate) : proposedDecision === "replace" ? t($ => $.ragApplication.assignment.replace) : t($ => $.ragApplication.assignment.revoke)}
          </button>
        </div>
      ) : null}
      {runtime.failureCode ? <p className="failure-summary"><code>{runtime.failureCode}</code>: {workflowRAGAssignmentFeedback(t, runtime)}</p> : <p className="boundary-note">{workflowRAGAssignmentFeedback(t, runtime)}</p>}
      {runtime.status === "version_conflict" ? <button type="button" onClick={() => void loadAssignment(true)} disabled={busy !== ""}>{t($ => $.ragApplication.assignment.refresh)}</button> : null}
      {!candidateApproved ? <p className="failure-summary">{t($ => $.ragApplication.assignment.notApproved)}</p> : null}
    </article>
  );
}

export default function ApplicationRAGInvocationPanel({
  applicationId,
  applicationName,
  applicationActive,
  onRunRecorded,
  onOpenRun,
  onEvidenceChange,
}: {
  applicationId: string;
  applicationName: string;
  applicationActive: boolean;
  onRunRecorded: (runId: string) => void;
  onOpenRun?: (runId: string) => void;
  onEvidenceChange?: (evidence: ApplicationDevelopmentOwnerEvidence) => void;
}) {
  const { t } = useTranslation("workflow");
  const [credential, setCredential] = useState<{ apiKeyId: string; token: string } | null>(null);
  const [input, setInput] = useState("");
  const [result, setResult] = useState<WorkflowRAGApplicationInvocationResult>(() => initialWorkflowRAGApplicationInvocationResult(config));
  const [busy, setBusy] = useState(false);
  const generation = useRef(0);

  useEffect(() => {
    generation.current += 1;
    setCredential(null);
    setInput("");
    setResult(initialWorkflowRAGApplicationInvocationResult(config));
    setBusy(false);
  }, [applicationId]);

  useEffect(() => {
    if (!onEvidenceChange || !result.runId) return;
    const succeeded = result.status === "succeeded" && result.runStatus === "succeeded";
    onEvidenceChange({
      contributionId: "controlled_run",
      status: succeeded ? "available" : "blocked",
      coverage: "complete",
      evidenceRefs: [{ kind: "run", id: result.runId }],
      missingEvidence: succeeded ? [] : ["Review the terminal Application RAG run failure before continuing."],
      blockers: succeeded ? [] : [{
        code: result.failureCode || `controlled_run_${result.runStatus || "failed"}`,
        summary: result.failureSummary || result.summary,
      }],
      failureCodes: result.failureCode ? [result.failureCode] : [],
    });
  }, [onEvidenceChange, result]);

  useEffect(() => {
    function receiveCredential(event: Event) {
      const detail = (event as CustomEvent<WorkflowRAGApplicationCredentialHandoffDetail>).detail;
      if (!detail || detail.applicationId !== applicationId) return;
      generation.current += 1;
      setCredential({ apiKeyId: detail.apiKeyId, token: detail.token });
      setInput("");
      setResult(initialWorkflowRAGApplicationInvocationResult(config));
      setBusy(false);
    }
    function clearAfterRouteLeave() {
      if (window.location.hash !== "#application-rag-invocation") {
        generation.current += 1;
        setCredential(null);
        setInput("");
        setResult(initialWorkflowRAGApplicationInvocationResult(config));
        setBusy(false);
      }
    }
    window.addEventListener(WORKFLOW_RAG_APPLICATION_CREDENTIAL_HANDOFF_EVENT, receiveCredential);
    window.addEventListener("hashchange", clearAfterRouteLeave);
    return () => {
      generation.current += 1;
      window.removeEventListener(WORKFLOW_RAG_APPLICATION_CREDENTIAL_HANDOFF_EVENT, receiveCredential);
      window.removeEventListener("hashchange", clearAfterRouteLeave);
    };
  }, [applicationId]);

  async function invoke() {
    if (!credential || !input.trim() || !applicationActive || busy) return;
    const current = ++generation.current;
    setBusy(true);
    setResult(initialWorkflowRAGApplicationInvocationResult(config));
    const next = await invokeWorkflowRAGApplication(config, { applicationId, apiKeyId: credential.apiKeyId, token: credential.token, text: input });
    if (generation.current !== current) return;
    setBusy(false);
    setResult(next);
    if (next.runId) onRunRecorded(next.runId);
    setInput("");
  }

  return (
    <section className="surface-band workflow-rag-application-invocation" id="application-rag-invocation" aria-labelledby="application-rag-invocation-title">
      <div className="section-heading">
        <div><p className="eyebrow">{t($ => $.ragApplication.invocation.eyebrow)}</p><h3 id="application-rag-invocation-title">{t($ => $.ragApplication.invocation.title)}</h3></div>
        <span className={`status-badge ${result.status === "succeeded" ? "good" : result.failureCode ? "bad" : "neutral"}`}>{workflowRAGApplicationStatus(t, busy ? "invoking" : !result.failureCode && !result.runId ? "idle" : result.status)}</span>
      </div>
      <div className="workflow-rag-runtime-scope">
        <article><span>{t($ => $.ragApplication.invocation.application)}</span><strong>{applicationName || t($ => $.ragApplication.invocation.noApplication)}</strong><code>{applicationId || t($ => $.ragApplication.invocation.unbound)}</code></article>
        <article><span>{t($ => $.ragApplication.invocation.credential)}</span><strong>{credential?.apiKeyId ?? t($ => $.ragApplication.invocation.noCredential)}</strong><p>{t($ => $.ragApplication.invocation.memory)}</p></article>
        <article><span>{t($ => $.ragApplication.invocation.authority)}</span><strong>{t($ => $.ragApplication.invocation.serverSelected)}</strong><p>{t($ => $.ragApplication.invocation.authorityNote)}</p></article>
      </div>
      <label>{t($ => $.ragApplication.invocation.input)}<textarea value={input} onChange={(event) => setInput(event.target.value)} rows={5} maxLength={4096} disabled={!credential || !applicationActive || busy} placeholder={t($ => $.ragApplication.invocation.inputPlaceholder)} /></label>
      <div className="workflow-rag-runtime-actions">
        <button type="button" onClick={() => void invoke()} disabled={!credential || !input.trim() || !applicationActive || busy || config.mode === "offline"}>{busy ? t($ => $.ragApplication.invocation.invoking) : t($ => $.ragApplication.invocation.invoke)}</button>
        <button type="button" className="secondary-action" onClick={() => { generation.current += 1; setCredential(null); setInput(""); setResult(initialWorkflowRAGApplicationInvocationResult(config)); setBusy(false); }} disabled={!credential && !input && !result.answer}>{t($ => $.ragApplication.invocation.clear)}</button>
        {result.runId ? <button type="button" className="secondary-action" onClick={() => onOpenRun?.(result.runId)}>{t($ => $.ragApplication.invocation.history)}</button> : null}
      </div>
      {result.answer ? (
        <article className="workflow-rag-application-answer" aria-live="polite">
          <div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.ragApplication.invocation.answer)}</p><h4>{result.answer.confidence === "high" ? t($ => $.ragApplication.invocation.high) : result.answer.confidence === "medium" ? t($ => $.ragApplication.invocation.medium) : t($ => $.ragApplication.invocation.low)}</h4></div><code>{result.runId}</code></div>
          <p>{result.answer.answer}</p>
          <h5>{t($ => $.ragApplication.invocation.citations)}</h5>
          <ul>{result.answer.citations.map((citation) => <li key={citation.fragmentRef}><code>{citation.fragmentRef}</code><span>{citation.claimSummary}</span></li>)}</ul>
          {result.answer.limitations.length ? <><h5>{t($ => $.ragApplication.invocation.limitations)}</h5><ul>{result.answer.limitations.map((limitation) => <li key={limitation}>{limitation}</li>)}</ul></> : null}
        </article>
      ) : null}
      {result.failureCode ? <p className="failure-summary" role="alert"><code>{result.failureCode}</code>: {workflowRAGInvocationFeedback(t, result)}</p> : <p className="boundary-note">{workflowRAGInvocationFeedback(t, result)}</p>}
      {result.failureCode && result.failureSummary ? <details><summary>{t($ => $.ragApplication.invocation.diagnostic)}</summary><p>{result.failureSummary}</p></details> : null}
      <p className="boundary-note">{t($ => $.ragApplication.invocation.privacy)}</p>
    </section>
  );
}

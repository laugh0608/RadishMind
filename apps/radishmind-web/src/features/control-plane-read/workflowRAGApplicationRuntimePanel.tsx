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

const config = readWorkflowRAGApplicationRuntimeConfig();

export function WorkflowRAGRuntimeAssignmentPanel({
  applicationId,
  publishCandidateId,
  candidateApproved,
  readOnly = false,
}: {
  applicationId: string;
  publishCandidateId: string;
  candidateApproved: boolean;
  readOnly?: boolean;
}) {
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
        <div><p className="eyebrow">Application RAG runtime</p><h5 id="workflow-rag-runtime-assignment-title">Explicit assignment</h5></div>
        <span className={`status-badge ${runtime.assignment?.state === "active" ? "good" : runtime.status === "failed" ? "bad" : "neutral"}`}>
          {runtime.assignment?.state ?? runtime.status}
        </span>
      </div>
      <p className="boundary-note">Publish approval and runtime activation remain separate. The server reloads the exact candidate, draft, binding, promotion, dataset, snapshots, and profile before every decision.</p>
      <button type="button" onClick={() => void loadAssignment()} disabled={busy !== "" || config.mode === "offline"}>
        {busy === "read" ? "Loading assignment…" : "Load current assignment"}
      </button>
      {runtime.assignment ? (
        <dl className="workflow-rag-runtime-metadata">
          <div><dt>Assignment</dt><dd>{runtime.assignment.assignmentId} · v{runtime.assignment.recordVersion}</dd></div>
          <div><dt>Candidate</dt><dd>{runtime.assignment.publishCandidateId} · review v{runtime.assignment.publishReviewVersion}</dd></div>
          <div><dt>Draft</dt><dd>{runtime.assignment.draftId} · v{runtime.assignment.draftVersion}</dd></div>
          <div><dt>Binding</dt><dd>{runtime.assignment.bindingRef.bindingId} · v{runtime.assignment.bindingRef.bindingVersion}</dd></div>
          <div><dt>Updated</dt><dd>{runtime.assignment.updatedAt} · {runtime.assignment.updatedByActorRef}</dd></div>
        </dl>
      ) : null}
      {!readOnly ? (
        <div className="workflow-rag-runtime-decision">
          <label>Sanitized decision reason<textarea value={reason} onChange={(event) => setReason(event.target.value)} rows={3} maxLength={500} placeholder="Explain why this exact approved candidate should be activated, replaced, or revoked." /></label>
          <button type="button" onClick={() => void decide()} disabled={!decisionAllowed || busy !== ""}>
            {busy === "decision" ? "Recording decision…" : `${proposedDecision} runtime assignment`}
          </button>
        </div>
      ) : null}
      {runtime.failureCode ? <p className="failure-summary">{runtime.failureCode}: {runtime.summary}</p> : <p className="boundary-note">{runtime.summary}</p>}
      {runtime.status === "version_conflict" ? <button type="button" onClick={() => void loadAssignment(true)} disabled={busy !== ""}>Refresh current version and keep reason</button> : null}
      {!candidateApproved ? <p className="failure-summary">Only an approved publish candidate can be selected for runtime assignment.</p> : null}
    </article>
  );
}

export default function ApplicationRAGInvocationPanel({
  applicationId,
  applicationName,
  applicationActive,
  onRunRecorded,
}: {
  applicationId: string;
  applicationName: string;
  applicationActive: boolean;
  onRunRecorded: (runId: string) => void;
}) {
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
        <div><p className="eyebrow">User Workspace</p><h3 id="application-rag-invocation-title">Application RAG Invocation</h3></div>
        <span className={`status-badge ${result.status === "succeeded" ? "good" : result.failureCode ? "bad" : "neutral"}`}>{result.status}</span>
      </div>
      <div className="workflow-rag-runtime-scope">
        <article><span>Application</span><strong>{applicationName || "No application selected"}</strong><code>{applicationId || "unbound"}</code></article>
        <article><span>Credential</span><strong>{credential?.apiKeyId ?? "No one-time handoff"}</strong><p>Raw token remains only in current component memory.</p></article>
        <article><span>Authority</span><strong>Server selected</strong><p>Model, protocol, candidate, binding, snapshot, profile, ranking, and citations are not client controls.</p></article>
      </div>
      <label>Bounded input<textarea value={input} onChange={(event) => setInput(event.target.value)} rows={5} maxLength={4096} disabled={!credential || !applicationActive || busy} placeholder="Ask one bounded question against the active candidate snapshot. Do not include credentials or URLs." /></label>
      <div className="workflow-rag-runtime-actions">
        <button type="button" onClick={() => void invoke()} disabled={!credential || !input.trim() || !applicationActive || busy || config.mode === "offline"}>{busy ? "Invoking…" : "Invoke active RAG assignment"}</button>
        <button type="button" className="secondary-action" onClick={() => { generation.current += 1; setCredential(null); setInput(""); setResult(initialWorkflowRAGApplicationInvocationResult(config)); }} disabled={!credential && !input && !result.answer}>Clear sensitive memory</button>
        {result.runId ? <button type="button" className="secondary-action" onClick={() => { window.location.hash = "workspace-run-history"; }}>Open Run History</button> : null}
      </div>
      {result.answer ? (
        <article className="workflow-rag-application-answer" aria-live="polite">
          <div className="application-api-card-heading"><div><p className="eyebrow">Advisory answer</p><h4>{result.answer.confidence} confidence</h4></div><code>{result.runId}</code></div>
          <p>{result.answer.answer}</p>
          <h5>Citations</h5>
          <ul>{result.answer.citations.map((citation) => <li key={citation.fragmentRef}><code>{citation.fragmentRef}</code><span>{citation.claimSummary}</span></li>)}</ul>
          {result.answer.limitations.length ? <><h5>Limitations</h5><ul>{result.answer.limitations.map((limitation) => <li key={limitation}>{limitation}</li>)}</ul></> : null}
        </article>
      ) : null}
      {result.failureCode ? <p className="failure-summary" role="alert">{result.failureCode}: {result.failureSummary || result.summary}</p> : <p className="boundary-note">{result.summary}</p>}
      <p className="boundary-note">Input, answer, token, selected fragment content, prompt packet, and provider response are not written to URL, browser storage, logs, assignment history, or Workflow Run History.</p>
    </section>
  );
}

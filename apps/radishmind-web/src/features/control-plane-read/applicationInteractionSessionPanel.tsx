import { useEffect, useRef, useState } from "react";

import {
  applicationInteractionResponseMatchesScope,
  closeApplicationInteractionSession,
  createApplicationInteractionSession,
  executeApplicationInteractionTurn,
  initialApplicationInteractionSessionListResult,
  listApplicationInteractionSessions,
  listApplicationInteractionTurns,
  readApplicationInteractionSessionConfig,
  type ApplicationInteractionAnswer,
  type ApplicationInteractionExecutionProfile,
  type ApplicationInteractionSession,
  type ApplicationInteractionSessionListResult,
  type ApplicationInteractionSessionState,
  type ApplicationInteractionTurn,
} from "./applicationInteractionSessionConsumer.ts";
import type { ApplicationDevelopmentOwnerEvidence } from "./applicationDevelopmentReadiness.ts";
import StructuredRuntimeInputEditor from "./StructuredRuntimeInputEditor.tsx";
import {
  structuredRuntimeInputAuthorityKey,
  validateStructuredRuntimeInputDrafts,
  type StructuredRuntimeInputDrafts,
} from "./structuredRuntimeInput.ts";

const config = readApplicationInteractionSessionConfig();

type TranscriptEntry = {
  clientTurnKey: string;
  input: string;
  status: "running" | "succeeded" | "failed" | "canceled";
  output: string;
  answer: ApplicationInteractionAnswer | null;
  turn: ApplicationInteractionTurn | null;
  failureCode: string;
  failureSummary: string;
};

type RequestScope = {
  generation: number;
  applicationId: string;
  sessionId: string;
  clientTurnKey: string;
};

export default function ApplicationInteractionSessionPanel({
  applicationId,
  applicationName,
  applicationActive,
  suggestedDefinitionId,
  onRunRecorded,
  onOpenRun,
  onEvidenceChange,
}: {
  applicationId: string;
  applicationName: string;
  applicationActive: boolean;
  suggestedDefinitionId: string;
  onRunRecorded: (runId: string) => void;
  onOpenRun?: (runId: string) => void;
  onEvidenceChange?: (evidence: ApplicationDevelopmentOwnerEvidence) => void;
}) {
  const [listing, setListing] = useState<ApplicationInteractionSessionListResult>(() => initialApplicationInteractionSessionListResult(config));
  const [sessions, setSessions] = useState<ApplicationInteractionSession[]>([]);
  const [selectedSession, setSelectedSession] = useState<ApplicationInteractionSession | null>(null);
  const [turns, setTurns] = useState<ApplicationInteractionTurn[]>([]);
  const [profile, setProfile] = useState<ApplicationInteractionExecutionProfile>("workflow_definition_executor_v1");
  const [sessionStateFilter, setSessionStateFilter] = useState<ApplicationInteractionSessionState>("active");
  const [definitionId, setDefinitionId] = useState(suggestedDefinitionId);
  const [input, setInput] = useState("");
  const [structuredInputDrafts, setStructuredInputDrafts] = useState<StructuredRuntimeInputDrafts>({});
  const [structuredInputErrors, setStructuredInputErrors] = useState<Record<string, string>>({});
  const [conditionValues, setConditionValues] = useState("{}");
  const [model, setModel] = useState("");
  const [temperature, setTemperature] = useState("");
  const [transcript, setTranscript] = useState<TranscriptEntry[]>([]);
  const [pending, setPending] = useState<"" | "list" | "create" | "select" | "close" | "turn">("");
  const [operationFailure, setOperationFailure] = useState("");
  const abortRef = useRef<AbortController | null>(null);
  const generationRef = useRef(0);
  const requestScopeRef = useRef<RequestScope>({ generation: 0, applicationId: "", sessionId: "", clientTurnKey: "" });

  useEffect(() => {
    if (!onEvidenceChange) return;
    const latestTurn = [...turns, ...transcript.map((entry) => entry.turn).filter((turn): turn is ApplicationInteractionTurn => Boolean(turn))]
      .filter((turn) => turn.runRef)
      .sort((left, right) => right.sequence - left.sequence)[0];
    if (!latestTurn && !operationFailure) return;
    const ownerFailed = Boolean(operationFailure) && !latestTurn;
    const succeeded = latestTurn?.status === "succeeded";
    const terminalBlocked = Boolean(latestTurn && latestTurn.status !== "succeeded" && latestTurn.status !== "running");
    onEvidenceChange({
      contributionId: "controlled_run",
      status: ownerFailed || terminalBlocked ? "blocked" : succeeded ? "available" : "incomplete",
      coverage: latestTurn || ownerFailed ? "complete" : "none",
      evidenceRefs: latestTurn?.runRef ? [
        { kind: "session", id: latestTurn.sessionId },
        { kind: "run", id: latestTurn.runRef.runId },
      ] : [],
      missingEvidence: succeeded ? [] : ["Complete and review a terminal v4, v5 or v8 Application interaction run."],
      blockers: ownerFailed || terminalBlocked ? [{
        code: latestTurn?.failureCode || operationFailure || "controlled_run_blocked",
        summary: latestTurn?.failureSummary || "The Application interaction owner reports a failure or non-success terminal run.",
      }] : [],
      failureCodes: latestTurn?.failureCode ? [latestTurn.failureCode] : operationFailure ? [operationFailure] : [],
    });
  }, [onEvidenceChange, operationFailure, transcript, turns]);

  useEffect(() => {
    abortRef.current?.abort();
    const generation = ++generationRef.current;
    requestScopeRef.current = { generation, applicationId, sessionId: "", clientTurnKey: "" };
    setListing(initialApplicationInteractionSessionListResult(config));
    setSessions([]);
    setSelectedSession(null);
    setTurns([]);
    setProfile("workflow_definition_executor_v1");
    setSessionStateFilter("active");
    setDefinitionId(suggestedDefinitionId);
    setInput("");
    setConditionValues("{}");
    setModel("");
    setTemperature("");
    setTranscript([]);
    setPending("");
    setOperationFailure("");
    if (config.mode === "offline" || !applicationActive || !applicationId) return;
    const controller = new AbortController();
    abortRef.current = controller;
    setPending("list");
    void listApplicationInteractionSessions(config, applicationId, { state: "active" }, controller.signal).then((result) => {
      if (generationRef.current !== generation) return;
      abortRef.current = null;
      setPending("");
      setListing(result);
      setSessions(result.sessions);
      setOperationFailure(result.failureCode);
    });
    return () => {
      controller.abort();
      generationRef.current += 1;
    };
  }, [applicationActive, applicationId]);

  const structuredInputContract = selectedSession?.executionProfile === "workflow_definition_executor_v2"
    ? selectedSession.authority.workflowDefinition?.inputContract ?? null
    : null;

  useEffect(() => {
    setStructuredInputDrafts({});
    setStructuredInputErrors({});
  }, [selectedSession?.sessionId, structuredInputContract ? structuredRuntimeInputAuthorityKey(structuredInputContract) : "legacy"]);

  function beginOperation(next: typeof pending): AbortController {
    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;
    setPending(next);
    setOperationFailure("");
    return controller;
  }

  async function reloadSessions() {
    if (config.mode === "offline" || !applicationActive || !applicationId) return;
    const generation = ++generationRef.current;
    requestScopeRef.current = { generation, applicationId, sessionId: "", clientTurnKey: "" };
    setSelectedSession(null);
    setTurns([]);
    clearTransientInput();
    const controller = beginOperation("list");
    const result = await listApplicationInteractionSessions(config, applicationId, { state: sessionStateFilter }, controller.signal);
    if (generationRef.current !== generation) return;
    abortRef.current = null;
    setPending("");
    setListing(result);
    setSessions(result.sessions);
    setOperationFailure(result.failureCode);
  }

  async function createSession() {
    if (!applicationActive || pending || ((profile === "workflow_definition_executor_v1" || profile === "workflow_definition_executor_v2") && !definitionId.trim())) return;
    const generation = ++generationRef.current;
    requestScopeRef.current = { generation, applicationId, sessionId: "", clientTurnKey: "" };
    setSelectedSession(null);
    setTurns([]);
    clearTransientInput();
    const controller = beginOperation("create");
    const result = await createApplicationInteractionSession(config, { applicationId, executionProfile: profile, definitionId: profile === "application_rag_invocation_v1" ? "" : definitionId }, controller.signal);
    if (generationRef.current !== generation) return;
    abortRef.current = null;
    setPending("");
    setOperationFailure(result.failureCode);
    if (!result.session) return;
    setSelectedSession(result.session);
    setSessions((current) => mergeSessions(current, result.session as ApplicationInteractionSession));
    requestScopeRef.current = { generation, applicationId, sessionId: result.session.sessionId, clientTurnKey: "" };
  }

  async function selectSession(session: ApplicationInteractionSession) {
    const generation = ++generationRef.current;
    requestScopeRef.current = { generation, applicationId, sessionId: session.sessionId, clientTurnKey: "" };
    setSelectedSession(session);
    setTurns([]);
    clearTransientInput();
    const controller = beginOperation("select");
    const result = await listApplicationInteractionTurns(config, session, controller.signal);
    if (generationRef.current !== generation) return;
    abortRef.current = null;
    setPending("");
    setTurns(result.turns);
    setOperationFailure(result.failureCode);
  }

  async function closeSession() {
    if (!selectedSession || selectedSession.state !== "active" || pending) return;
    const generation = ++generationRef.current;
    requestScopeRef.current = { generation, applicationId, sessionId: selectedSession.sessionId, clientTurnKey: "" };
    const controller = beginOperation("close");
    const result = await closeApplicationInteractionSession(config, selectedSession, controller.signal);
    if (generationRef.current !== generation) return;
    abortRef.current = null;
    setPending("");
    setOperationFailure(result.failureCode);
    if (!result.session) return;
    setSelectedSession(result.session);
    setSessions((current) => mergeSessions(current, result.session as ApplicationInteractionSession));
    setInput("");
    setStructuredInputDrafts({});
    setStructuredInputErrors({});
  }

  async function submitTurn() {
    if (!selectedSession || selectedSession.state !== "active" || pending) return;
    const structuredValidation = structuredInputContract
      ? validateStructuredRuntimeInputDrafts(structuredInputContract, structuredInputDrafts)
      : null;
    if (structuredValidation && !structuredValidation.ok) {
      setStructuredInputErrors(structuredValidation.fieldErrors);
      setOperationFailure(structuredValidation.failureCode);
      return;
    }
    if (!structuredInputContract && !input.trim()) return;
    const parsedConditions = parseConditionValues(conditionValues);
    const parsedTemperature = parseTemperature(temperature);
    if (!parsedConditions || parsedTemperature === undefined) {
      setOperationFailure("application_session_workflow_options_invalid");
      return;
    }
    const clientTurnKey = createClientTurnKey();
    const generation = ++generationRef.current;
    const expectedScope = { generation, applicationId, sessionId: selectedSession.sessionId, clientTurnKey };
    requestScopeRef.current = expectedScope;
    const currentInput = structuredInputContract
      ? `${Object.keys(structuredValidation!.inputs).sort().join(", ") || "无可选字段"} 已按合同提交；输入值未写入 transcript。`
      : input.trim();
    setInput("");
    setStructuredInputDrafts({});
    setStructuredInputErrors({});
    setTranscript((current) => [...current, { clientTurnKey, input: currentInput, status: "running", output: "", answer: null, turn: null, failureCode: "", failureSummary: "" }]);
    const controller = beginOperation("turn");
    const result = await executeApplicationInteractionTurn(config, {
      session: selectedSession,
      clientTurnKey,
      ...(selectedSession.executionProfile === "workflow_definition_executor_v2" ? { inputs: structuredValidation!.inputs } : { inputText: currentInput }),
      conditionValues: selectedSession.executionProfile === "application_rag_invocation_v1" ? {} : parsedConditions,
      model: selectedSession.executionProfile === "application_rag_invocation_v1" ? "" : model,
      temperature: selectedSession.executionProfile === "application_rag_invocation_v1" ? null : parsedTemperature,
    }, controller.signal);
    if (!applicationInteractionResponseMatchesScope(expectedScope, requestScopeRef.current)) return;
    abortRef.current = null;
    setPending("");
    setOperationFailure(result.failureCode);
    setTranscript((current) => current.map((entry) => entry.clientTurnKey === clientTurnKey ? {
      ...entry,
      status: result.status === "succeeded" ? "succeeded" : result.status === "canceled" ? "canceled" : "failed",
      output: result.advisoryOutput || result.answer?.answer || "",
      answer: result.answer,
      turn: result.turn,
      failureCode: result.failureCode,
      failureSummary: result.failureSummary || result.summary,
    } : entry));
    if (result.session) {
      setSelectedSession(result.session);
      setSessions((current) => mergeSessions(current, result.session as ApplicationInteractionSession));
    }
    if (result.turn) setTurns((current) => mergeTurns(current, result.turn as ApplicationInteractionTurn));
    if (result.turn?.runRef?.runId) onRunRecorded(result.turn.runRef.runId);
  }

  function cancelTurn() {
    if (pending !== "turn") return;
    const clientTurnKey = requestScopeRef.current.clientTurnKey;
    abortRef.current?.abort();
    abortRef.current = null;
    const generation = ++generationRef.current;
    requestScopeRef.current = { generation, applicationId, sessionId: selectedSession?.sessionId ?? "", clientTurnKey: "" };
    setPending("");
    setOperationFailure("application_session_request_canceled");
    setTranscript((current) => current.map((entry) => entry.clientTurnKey === clientTurnKey && entry.status === "running"
      ? { ...entry, status: "canceled", failureCode: "application_session_request_canceled", failureSummary: "Request canceled locally; any late response is ignored." }
      : entry));
  }

  function clearTransientInput() {
    setInput("");
    setStructuredInputDrafts({});
    setStructuredInputErrors({});
    setConditionValues("{}");
    setModel("");
    setTemperature("");
    setTranscript([]);
  }

  const workflowProfile = selectedSession?.executionProfile === "workflow_definition_executor_v1" || selectedSession?.executionProfile === "workflow_definition_executor_v2";
  const canCreate = config.mode !== "offline" && applicationActive && !pending && (profile === "application_rag_invocation_v1" || Boolean(definitionId.trim()));
  const canSubmit = config.mode !== "offline" && applicationActive && selectedSession?.state === "active" && Boolean(structuredInputContract || input.trim()) && !pending;

  return (
    <section className="surface-band application-interaction-session" id="application-interaction-session" aria-labelledby="application-interaction-session-title">
      <div className="section-heading">
        <div><p className="eyebrow">User Workspace</p><h3 id="application-interaction-session-title">Application Interaction</h3></div>
        <span className={`status-badge ${selectedSession?.state === "active" ? "good" : operationFailure ? "bad" : "neutral"}`}>
          {config.mode === "offline" ? "offline" : selectedSession?.state ?? listing.status}
        </span>
      </div>

      <div className="application-interaction-scope">
        <article><span>Application</span><strong>{applicationName || "No application selected"}</strong><code>{applicationId || "unbound"}</code></article>
        <article><span>Session owner</span><strong>{selectedSession?.sessionId ?? "No session selected"}</strong><p>{selectedSession ? `${selectedSession.executionProfile} · v${selectedSession.recordVersion}` : "Choose an explicit runtime profile."}</p></article>
        <article><span>Retention</span><strong>metadata_only</strong><p>Transcript, input, answer, prompt, provider response, and credentials remain outside persistent stores.</p></article>
      </div>

      <div className="application-interaction-create">
        <label>Session state
          <select value={sessionStateFilter} onChange={(event) => setSessionStateFilter(event.target.value as ApplicationInteractionSessionState)} disabled={Boolean(pending)}>
            <option value="active">Active</option>
            <option value="closed">Closed</option>
          </select>
        </label>
        <label>Execution profile
          <select value={profile} onChange={(event) => setProfile(event.target.value as ApplicationInteractionExecutionProfile)} disabled={Boolean(pending)}>
            <option value="workflow_definition_executor_v1">Workflow Definition v5</option>
            <option value="workflow_definition_executor_v2">Workflow Definition v8 · structured inputs</option>
            <option value="application_rag_invocation_v1">Application RAG v4</option>
          </select>
        </label>
        {profile !== "application_rag_invocation_v1" ? <label>Active definition ID<input value={definitionId} onChange={(event) => setDefinitionId(event.target.value)} disabled={Boolean(pending)} placeholder="definition_support_flow" /></label> : null}
        <button type="button" onClick={() => void createSession()} disabled={!canCreate}>{pending === "create" ? "Creating…" : "Create bound session"}</button>
        <button type="button" className="secondary-action" onClick={() => void reloadSessions()} disabled={config.mode === "offline" || Boolean(pending) || !applicationActive}>Reload sessions</button>
      </div>

      <div className="application-interaction-session-list" aria-label="Application sessions">
        {sessions.length === 0 ? <p className="empty-state">{listing.summary}</p> : sessions.map((session) => (
          <button type="button" key={session.sessionId} className={selectedSession?.sessionId === session.sessionId ? "selected" : ""} onClick={() => void selectSession(session)} disabled={Boolean(pending)}>
            <span><strong>{session.executionProfile === "application_rag_invocation_v1" ? "Application RAG" : session.executionProfile === "workflow_definition_executor_v2" ? "Workflow Definition · structured" : "Workflow Definition"}</strong><code>{session.sessionId}</code></span>
            <span><small>{session.state} · v{session.recordVersion}</small><small>{session.turnCount} turns</small></span>
          </button>
        ))}
      </div>

      {selectedSession ? (
        <>
          <div className="application-interaction-authority">
            <div><span>Authority digest</span><code>{selectedSession.authority.authorityDigest}</code></div>
            <div><span>Exact binding</span><strong>{selectedSession.authority.workflowDefinition ? `${selectedSession.authority.workflowDefinition.definitionId} · definition v${selectedSession.authority.workflowDefinition.definitionVersion} · pointer v${selectedSession.authority.workflowDefinition.activationPointerVersion}` : `${selectedSession.authority.applicationRAG?.assignmentId} · assignment v${selectedSession.authority.applicationRAG?.assignmentVersion}`}</strong></div>
            {structuredInputContract ? <div><span>Input contract</span><strong>{structuredInputContract.contractId}</strong><code>{structuredInputContract.contractDigest}</code></div> : null}
            <button type="button" className="secondary-action" onClick={() => void closeSession()} disabled={selectedSession.state !== "active" || Boolean(pending)}>{pending === "close" ? "Closing…" : "Close session"}</button>
          </div>

          <div className="application-interaction-composer">
            {structuredInputContract ? <StructuredRuntimeInputEditor contract={structuredInputContract} drafts={structuredInputDrafts} fieldErrors={structuredInputErrors} disabled={selectedSession.state !== "active" || Boolean(pending)} onChange={(drafts) => { setStructuredInputDrafts(drafts); setStructuredInputErrors({}); }} /> : <label>Transient input<textarea value={input} onChange={(event) => setInput(event.target.value)} rows={5} maxLength={8192} disabled={selectedSession.state !== "active" || Boolean(pending)} placeholder="Send one bounded interaction. Do not include credentials, tokens, or authorization headers." /></label>}
            {workflowProfile ? (
              <div className="application-interaction-workflow-options">
                <label>Condition values JSON<textarea value={conditionValues} onChange={(event) => setConditionValues(event.target.value)} rows={2} disabled={Boolean(pending)} /></label>
                <label>Model override<input value={model} onChange={(event) => setModel(event.target.value)} disabled={Boolean(pending)} placeholder="Server default" /></label>
                <label>Temperature<input value={temperature} onChange={(event) => setTemperature(event.target.value)} disabled={Boolean(pending)} inputMode="decimal" placeholder="Server default" /></label>
              </div>
            ) : null}
            <div className="application-interaction-actions">
              <button type="button" onClick={() => void submitTurn()} disabled={!canSubmit}>{pending === "turn" ? "Running…" : "Run turn"}</button>
              {pending === "turn" ? <button type="button" className="secondary-action" onClick={cancelTurn}>Cancel current turn</button> : null}
              <button type="button" className="secondary-action" onClick={clearTransientInput} disabled={pending === "turn" || (!input && Object.keys(structuredInputDrafts).length === 0 && transcript.length === 0)}>Clear transient transcript</button>
            </div>
          </div>
        </>
      ) : null}

      {transcript.length ? (
        <div className="application-interaction-transcript" aria-live="polite" aria-label="Current transient transcript">
          {transcript.map((entry) => (
            <article key={entry.clientTurnKey}>
              <div className="application-api-card-heading"><div><p className="eyebrow">Transient turn</p><h4>{entry.status}</h4></div><code>{entry.turn?.runRef?.runId ?? entry.clientTurnKey}</code></div>
              <div className="application-interaction-message user-message"><span>User</span><p>{entry.input}</p></div>
              {entry.output ? <div className="application-interaction-message assistant-message"><span>Advisory response</span><p>{entry.output}</p></div> : null}
              {entry.answer?.citations.length ? <ul>{entry.answer.citations.map((citation) => <li key={citation.fragmentRef}><code>{citation.fragmentRef}</code><span>{citation.claimSummary}</span></li>)}</ul> : null}
              {entry.failureCode ? <p className="failure-summary" role="alert">{entry.failureCode}: {entry.failureSummary}</p> : null}
              {entry.turn?.runRef ? <button type="button" className="secondary-action" onClick={() => onOpenRun?.(entry.turn?.runRef?.runId ?? "")}>Open run {entry.turn.runRef.schemaVersion}</button> : null}
            </article>
          ))}
        </div>
      ) : null}

      <div className="application-interaction-turn-history">
        <div className="application-api-card-heading"><div><p className="eyebrow">Persisted evidence</p><h4>Turn metadata only</h4></div><span>{turns.length}</span></div>
        {turns.length === 0 ? <p className="empty-state">Select a session to load status, digest, byte count, run ref, failure metadata, and timestamps—never transcript content.</p> : (
          <ul>{turns.map((turn) => <li key={turn.turnId}><span><strong>#{turn.sequence} · {turn.status}</strong><code>{turn.turnId}</code></span><span><small>{turn.inputDigest} · {turn.inputBytes} bytes</small>{turn.inputFields.length ? <small>{turn.inputFields.map((field) => `${field.name}:${field.valueType}`).join(" · ")}</small> : null}<code>{turn.runRef ? `${turn.runRef.schemaVersion} · ${turn.runRef.runId}` : turn.failureCode || "no run ref"}</code></span></li>)}</ul>
        )}
      </div>

      {operationFailure ? <p className="failure-summary" role="alert">{operationFailure}</p> : <p className="boundary-note">Application changes, session changes, route unmount, and cancellation invalidate pending response generations. Late responses cannot repopulate this workspace.</p>}
    </section>
  );
}

function mergeSessions(current: ApplicationInteractionSession[], next: ApplicationInteractionSession): ApplicationInteractionSession[] {
  return [next, ...current.filter((session) => session.sessionId !== next.sessionId)].sort((left, right) => right.updatedAt.localeCompare(left.updatedAt));
}

function mergeTurns(current: ApplicationInteractionTurn[], next: ApplicationInteractionTurn): ApplicationInteractionTurn[] {
  return [...current.filter((turn) => turn.turnId !== next.turnId), next].sort((left, right) => left.sequence - right.sequence);
}

function parseConditionValues(value: string): Record<string, boolean> | null {
  try {
    const parsed: unknown = JSON.parse(value);
    if (typeof parsed !== "object" || parsed === null || Array.isArray(parsed)) return null;
    return Object.keys(parsed).length <= 16 && Object.entries(parsed).every(([key, item]) => /^[A-Za-z0-9][A-Za-z0-9_.:/-]{0,159}$/u.test(key) && typeof item === "boolean")
      ? parsed as Record<string, boolean>
      : null;
  } catch {
    return null;
  }
}

function parseTemperature(value: string): number | null | undefined {
  if (!value.trim()) return null;
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed >= 0 && parsed <= 2 ? parsed : undefined;
}

function createClientTurnKey(): string {
  const suffix = (globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random()}`).replaceAll("-", "").replaceAll(".", "").slice(0, 24);
  return `web_turn_${suffix}`;
}

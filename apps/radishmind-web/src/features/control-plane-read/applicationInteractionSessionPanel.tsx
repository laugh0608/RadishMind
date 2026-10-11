import "../../i18n/interactionResources.ts";
import { useTranslation } from "react-i18next";
import { interactionFailureCode, interactionFailureMessage, interactionStatus } from "./applicationInteractionMessages.ts";
import type { StructuredRuntimeInputFieldError } from "./structuredRuntimeInput.ts";
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
import ApplicationResultArtifactPanel from "./applicationResultArtifactPanel.tsx";
import type { ApplicationResultArtifactSummary } from "./applicationResultArtifactConsumer.ts";
import { applicationInteractionSessionFailureGuidance } from "./controlledUseFailureGuidance.ts";
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
  submittedFields?: string[];
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
  const { t } = useTranslation("applications");
  const [listing, setListing] = useState<ApplicationInteractionSessionListResult>(() => initialApplicationInteractionSessionListResult(config));
  const [sessions, setSessions] = useState<ApplicationInteractionSession[]>([]);
  const [selectedSession, setSelectedSession] = useState<ApplicationInteractionSession | null>(null);
  const [turns, setTurns] = useState<ApplicationInteractionTurn[]>([]);
  const [profile, setProfile] = useState<ApplicationInteractionExecutionProfile>("workflow_definition_executor_v1");
  const [sessionStateFilter, setSessionStateFilter] = useState<ApplicationInteractionSessionState>("active");
  const [definitionId, setDefinitionId] = useState(suggestedDefinitionId);
  const [input, setInput] = useState("");
  const [structuredInputDrafts, setStructuredInputDrafts] = useState<StructuredRuntimeInputDrafts>({});
  const [structuredInputErrors, setStructuredInputErrors] = useState<Record<string, StructuredRuntimeInputFieldError>>({});
  const [conditionValues, setConditionValues] = useState("{}");
  const [model, setModel] = useState("");
  const [temperature, setTemperature] = useState("");
  const [transcript, setTranscript] = useState<TranscriptEntry[]>([]);
  const [saveResult, setSaveResult] = useState(false);
  const [latestResultArtifact, setLatestResultArtifact] = useState<ApplicationResultArtifactSummary | null>(null);
  const [resultArtifactFailureCode, setResultArtifactFailureCode] = useState("");
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
    setSaveResult(false);
    setLatestResultArtifact(null);
    setResultArtifactFailureCode("");
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
      abortRef.current?.abort();
      abortRef.current = null;
      requestScopeRef.current = { generation: ++generationRef.current, applicationId: "", sessionId: "", clientTurnKey: "" };
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
    const currentInput = structuredInputContract ? "" : input.trim();
    const submittedFields = structuredInputContract ? Object.keys(structuredValidation!.inputs).sort() : undefined;
    const shouldSaveResult = saveResult;
    setSaveResult(false);
    setLatestResultArtifact(null);
    setResultArtifactFailureCode("");
    setInput("");
    setStructuredInputDrafts({});
    setStructuredInputErrors({});
    setTranscript((current) => [...current, { clientTurnKey, input: currentInput, submittedFields, status: "running", output: "", answer: null, turn: null, failureCode: "", failureSummary: "" }]);
    const controller = beginOperation("turn");
    const result = await executeApplicationInteractionTurn(config, {
      session: selectedSession,
      clientTurnKey,
      ...(selectedSession.executionProfile === "workflow_definition_executor_v2" ? { inputs: structuredValidation!.inputs } : { inputText: currentInput }),
      conditionValues: selectedSession.executionProfile === "application_rag_invocation_v1" ? {} : parsedConditions,
      model: selectedSession.executionProfile === "application_rag_invocation_v1" ? "" : model,
      temperature: selectedSession.executionProfile === "application_rag_invocation_v1" ? null : parsedTemperature,
      saveResult: shouldSaveResult,
    }, controller.signal);
    if (!applicationInteractionResponseMatchesScope(expectedScope, requestScopeRef.current)) return;
    abortRef.current = null;
    setPending("");
    setOperationFailure(result.failureCode);
    setLatestResultArtifact(result.resultArtifact);
    setResultArtifactFailureCode(result.resultArtifactFailureCode);
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
    setSaveResult(false);
    setLatestResultArtifact(null);
    setResultArtifactFailureCode("");
  }

  const workflowProfile = selectedSession?.executionProfile === "workflow_definition_executor_v1" || selectedSession?.executionProfile === "workflow_definition_executor_v2";
  const canCreate = config.mode !== "offline" && applicationActive && !pending && (profile === "application_rag_invocation_v1" || Boolean(definitionId.trim()));
  const canSubmit = config.mode !== "offline" && applicationActive && selectedSession?.state === "active" && Boolean(structuredInputContract || input.trim()) && !pending;
  const failureGuidance = applicationInteractionSessionFailureGuidance(operationFailure);

  return (
    <section className="surface-band application-interaction-session" id="application-interaction-session" aria-labelledby="application-interaction-session-title">
      <div className="section-heading">
        <div><p className="eyebrow">{t($ => $.interaction.eyebrow)}</p><h3 id="application-interaction-session-title">{t($ => $.interaction.title)}</h3></div>
        <span className={`status-badge ${selectedSession?.state === "active" ? "good" : operationFailure ? "bad" : "neutral"}`}>
          {interactionStatus(t, config.mode === "offline" ? "offline" : pending ? "loading" : selectedSession?.state ?? listing.status)}
        </span>
      </div>

      <div className="application-interaction-scope">
        <article><span>{t($ => $.interaction.application)}</span><strong>{applicationName || t($ => $.interaction.noApplication)}</strong><code>{applicationId || t($ => $.interaction.unbound)}</code></article>
        <article><span>{t($ => $.interaction.owner)}</span><strong>{selectedSession?.sessionId ?? t($ => $.interaction.noSession)}</strong><p>{selectedSession ? `${selectedSession.executionProfile} · v${selectedSession.recordVersion}` : t($ => $.interaction.chooseProfile)}</p></article>
        <article><span>{t($ => $.interaction.retention)}</span><strong>{t($ => $.interaction.defaultOff)}</strong><p>{t($ => $.interaction.retentionHint)}</p></article>
      </div>

      <div className="application-interaction-create">
        <label>{t($ => $.interaction.sessionState)}<select value={sessionStateFilter} onChange={(event) => setSessionStateFilter(event.target.value as ApplicationInteractionSessionState)} disabled={Boolean(pending)}>
            <option value="active">{t($ => $.interaction.states.active)}</option>
            <option value="closed">{t($ => $.interaction.states.closed)}</option>
          </select>
        </label>
        <label>{t($ => $.interaction.profile)}<select value={profile} onChange={(event) => setProfile(event.target.value as ApplicationInteractionExecutionProfile)} disabled={Boolean(pending)}>
            <option value="workflow_definition_executor_v1">{t($ => $.interaction.definitionV5)}</option>
            <option value="workflow_definition_executor_v2">{t($ => $.interaction.definitionV8)}</option>
            <option value="application_rag_invocation_v1">{t($ => $.interaction.ragV4)}</option>
          </select>
        </label>
        {profile !== "application_rag_invocation_v1" ? <label>{t($ => $.interaction.definitionId)}<input value={definitionId} onChange={(event) => setDefinitionId(event.target.value)} disabled={Boolean(pending)} placeholder="definition_support_flow" /></label> : null}
        <button type="button" onClick={() => void createSession()} disabled={!canCreate}>{pending === "create" ? t($ => $.interaction.creating) : t($ => $.interaction.create)}</button>
        <button type="button" className="secondary-action" onClick={() => void reloadSessions()} disabled={config.mode === "offline" || Boolean(pending) || !applicationActive}>{t($ => $.interaction.reload)}</button>
      </div>

      <div className="application-interaction-session-list" aria-label={t($ => $.interaction.sessions)}>
        {sessions.length === 0 ? <p className="empty-state">{pending === "list" ? t($ => $.interaction.loading) : listing.failureCode ? interactionFailureMessage(t, listing.failureCode) : t($ => $.interaction.empty)}</p> : sessions.map((session) => (
          <button type="button" key={session.sessionId} className={selectedSession?.sessionId === session.sessionId ? "selected" : ""} onClick={() => void selectSession(session)} disabled={Boolean(pending)}>
            <span><strong>{session.executionProfile === "application_rag_invocation_v1" ? t($ => $.interaction.rag) : session.executionProfile === "workflow_definition_executor_v2" ? t($ => $.interaction.structuredDefinition) : t($ => $.interaction.definition)}</strong><code>{session.sessionId}</code></span>
            <span><small>{interactionStatus(t, session.state)} · v{session.recordVersion}</small><small>{t($ => $.interaction.turns, { count: session.turnCount })}</small></span>
          </button>
        ))}
      </div>

      {selectedSession ? (
        <>
          <div className="application-interaction-authority">
            <div><span>{t($ => $.interaction.authorityDigest)}</span><code>{selectedSession.authority.authorityDigest}</code></div>
            <div><span>{t($ => $.interaction.exactBinding)}</span><strong>{selectedSession.authority.workflowDefinition ? t($ => $.interaction.definitionBinding, { id: selectedSession.authority.workflowDefinition.definitionId, version: selectedSession.authority.workflowDefinition.definitionVersion, pointer: selectedSession.authority.workflowDefinition.activationPointerVersion }) : selectedSession.authority.applicationRAG ? t($ => $.interaction.ragBinding, { id: selectedSession.authority.applicationRAG.assignmentId, version: selectedSession.authority.applicationRAG.assignmentVersion }) : t($ => $.interaction.unbound)}</strong></div>
            {structuredInputContract ? <div><span>{t($ => $.interaction.inputContract)}</span><strong>{structuredInputContract.contractId}</strong><code>{structuredInputContract.contractDigest}</code></div> : null}
            <button type="button" className="secondary-action" onClick={() => void closeSession()} disabled={selectedSession.state !== "active" || Boolean(pending)}>{pending === "close" ? t($ => $.interaction.closing) : t($ => $.interaction.close)}</button>
          </div>

          <div className="application-interaction-composer">
            {structuredInputContract ? <StructuredRuntimeInputEditor contract={structuredInputContract} drafts={structuredInputDrafts} fieldErrors={structuredInputErrors} disabled={selectedSession.state !== "active" || Boolean(pending)} onChange={(drafts) => { setStructuredInputDrafts(drafts); setStructuredInputErrors({}); }} /> : <label>{t($ => $.interaction.input)}<textarea value={input} onChange={(event) => setInput(event.target.value)} rows={5} maxLength={8192} disabled={selectedSession.state !== "active" || Boolean(pending)} placeholder={t($ => $.interaction.inputPlaceholder)} /></label>}
            {workflowProfile ? (
              <div className="application-interaction-workflow-options">
                <label>{t($ => $.interaction.conditions)}<textarea value={conditionValues} onChange={(event) => setConditionValues(event.target.value)} rows={2} disabled={Boolean(pending)} /></label>
                <label>{t($ => $.interaction.model)}<input value={model} onChange={(event) => setModel(event.target.value)} disabled={Boolean(pending)} placeholder={t($ => $.interaction.serverDefault)} /></label>
                <label>{t($ => $.interaction.temperature)}<input value={temperature} onChange={(event) => setTemperature(event.target.value)} disabled={Boolean(pending)} inputMode="decimal" placeholder={t($ => $.interaction.serverDefault)} /></label>
              </div>
            ) : null}
            <div className="application-interaction-actions">
              <button type="button" onClick={() => void submitTurn()} disabled={!canSubmit}>{pending === "turn" ? t($ => $.interaction.running) : t($ => $.interaction.run)}</button>
              {pending === "turn" ? <button type="button" className="secondary-action" onClick={cancelTurn}>{t($ => $.interaction.cancel)}</button> : null}
              <button type="button" className="secondary-action" onClick={clearTransientInput} disabled={pending === "turn" || (!input && Object.keys(structuredInputDrafts).length === 0 && transcript.length === 0)}>{t($ => $.interaction.clear)}</button>
            </div>
          </div>
          <ApplicationResultArtifactPanel
            config={config}
            applicationId={applicationId}
            sessionId={selectedSession.sessionId}
            saveResult={saveResult}
            onSaveResultChange={setSaveResult}
            latestArtifact={latestResultArtifact}
            latestArtifactFailureCode={resultArtifactFailureCode}
            disabled={selectedSession.state !== "active" || Boolean(pending)}
            onOpenRun={onOpenRun}
          />
        </>
      ) : null}

      {transcript.length ? (
        <div className="application-interaction-transcript" aria-live="polite" aria-label={t($ => $.interaction.transcript)}>
          {transcript.map((entry) => (
            <article key={entry.clientTurnKey}>
              <div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.interaction.transientTurn)}</p><h4>{interactionStatus(t, entry.status)}</h4></div><code>{entry.turn?.runRef?.runId ?? entry.clientTurnKey}</code></div>
              <div className="application-interaction-message user-message"><span>{t($ => $.interaction.user)}</span><p>{entry.submittedFields ? t($ => $.interaction.structuredSubmitted, { fields: entry.submittedFields.join(", ") || t($ => $.interaction.noFields) }) : entry.input}</p></div>
              {entry.output ? <div className="application-interaction-message assistant-message"><span>{t($ => $.interaction.advisory)}</span><p>{entry.output}</p></div> : null}
              {entry.answer?.citations.length ? <ul>{entry.answer.citations.map((citation) => <li key={citation.fragmentRef}><code>{citation.fragmentRef}</code><span>{citation.claimSummary}</span></li>)}</ul> : null}
              {entry.failureCode ? <p className="failure-summary" role="alert">{interactionFailureMessage(t, entry.failureCode)} <code>{interactionFailureCode(entry.failureCode)}</code></p> : null}
              {entry.turn?.runRef ? <button type="button" className="secondary-action" onClick={() => onOpenRun?.(entry.turn?.runRef?.runId ?? "")}>{t($ => $.interaction.openRun, { schema: entry.turn.runRef.schemaVersion })}</button> : null}
            </article>
          ))}
        </div>
      ) : null}

      <div className="application-interaction-turn-history">
        <div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.interaction.persisted)}</p><h4>{t($ => $.interaction.metadataOnly)}</h4></div><span>{turns.length}</span></div>
        {turns.length === 0 ? <p className="empty-state">{t($ => $.interaction.noTurns)}</p> : (
          <ul>{turns.map((turn) => <li key={turn.turnId}><span><strong>#{turn.sequence} · {interactionStatus(t, turn.status)}</strong><code>{turn.turnId}</code></span><span><small>{turn.inputDigest} · {t($ => $.interaction.bytes, { count: turn.inputBytes })}</small>{turn.inputFields.length ? <small>{turn.inputFields.map((field) => `${field.name}:${field.valueType}`).join(" · ")}</small> : null}<code>{turn.runRef ? `${turn.runRef.schemaVersion} · ${turn.runRef.runId}` : turn.failureCode || t($ => $.interaction.noRun)}</code></span></li>)}</ul>
        )}
      </div>

      {failureGuidance ? (
        <article className="controlled-use-failure-guidance" aria-label={t($ => $.interaction.guidance)}>
          <div className="application-api-card-heading">
            <div><p className="eyebrow">{t($ => $.interaction.recovery)}</p><h5>{t($ => $.interaction.recovery)}</h5></div>
            <span className="status-badge bad">{t($ => $.interaction.noProviderCall)}</span>
          </div>
          <p>{interactionFailureMessage(t, operationFailure)}</p>
          <p>{t($ => $.interaction.recoveryHint)}</p>
          <p className="boundary-note">{t($ => $.interaction.sideEffects)}</p>
          <button type="button" className="secondary-action" onClick={() => void reloadSessions()} disabled={Boolean(pending) || !applicationActive}>{t($ => $.interaction.reloadCurrent)}</button>
        </article>
      ) : null}
      {operationFailure ? <p className="failure-summary" role="alert">{interactionFailureMessage(t, operationFailure)} <code>{interactionFailureCode(operationFailure)}</code></p> : <p className="boundary-note">{t($ => $.interaction.scopeHint)}</p>}
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

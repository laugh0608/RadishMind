import { useTranslation } from "react-i18next";
import { useEffect, useRef, useState } from "react";

import "./promptExecutionRecovery.css";
import { isPromptClientKey, promptExecutionNeedsReview } from "./promptExecutionRecovery.ts";
import PromptVariableInputEditor from "./PromptVariableInputEditor.tsx";
import { usePromptVariableInput } from "./usePromptVariableInput.ts";
import ControlledUseFailureGuidance from "./ControlledUseFailureGuidance.tsx";
import ApplicationResultArtifactPanel from "./applicationResultArtifactPanel.tsx";
import {
  createPromptApplicationSession,
  executePromptApplicationSessionTurn,
  initialPromptApplicationSessionResult,
  listPromptApplicationSessions,
  readPromptApplicationSessionConfig,
  readPromptApplicationSession,
  type PromptApplicationSession,
  type PromptApplicationSessionResult,
} from "./promptApplicationSessionConsumer.ts";

const baseConfig = readPromptApplicationSessionConfig();

export default function PromptApplicationSessionPanel({
  applicationId,
  workspaceId,
  onRunRecorded,
  onOpenRun,
}: {
  applicationId: string;
  workspaceId: string;
  onRunRecorded?: (runId: string) => void;
  onOpenRun?: (runId: string) => void;
}) {
  const { t } = useTranslation("prompt");
  const config = { ...baseConfig, workspaceId };
  const [result, setResult] = useState<PromptApplicationSessionResult>(
    () => initialPromptApplicationSessionResult(config),
  );
  const [selectedSession, setSelectedSession] = useState<PromptApplicationSession | null>(null);
  const [metadataStale, setMetadataStale] = useState(false);
  const [unresolved, setUnresolved] = useState(false);
  const dispatchedRef = useRef(false);
  const editor = usePromptVariableInput(applicationId, workspaceId, selectedSession);
  const [clientTurnKey, setClientTurnKey] = useState(() => newClientTurnKey());
  const [sessions, setSessions] = useState<PromptApplicationSession[]>([]);
  const [listFailureCode, setListFailureCode] = useState("");
  const [saveResult, setSaveResult] = useState(false);
  const [pending, setPending] = useState<"" | "list" | "create" | "execute" | "refresh">("");
  const generationRef = useRef(0);
  const controllerRef = useRef<AbortController | null>(null);

  useEffect(() => {
    const requestGeneration = ++generationRef.current;
    controllerRef.current?.abort();
    const controller = new AbortController();
    controllerRef.current = controller;
    setResult(initialPromptApplicationSessionResult(config));
    setSelectedSession(null);
    setMetadataStale(false);
    setUnresolved(false);
    setSessions([]);
    setListFailureCode("");
    setClientTurnKey(newClientTurnKey());
    setSaveResult(false);
    setPending(config.mode === "offline" ? "" : "list");
    void listPromptApplicationSessions(config, applicationId, controller.signal).then((listed) => {
      if (requestGeneration !== generationRef.current || controller.signal.aborted) return;
      controllerRef.current = null;
      setPending("");
      setSessions(listed.sessions);
      setListFailureCode(listed.failureCode);
      if (listed.sessions[0]) selectSession(listed.sessions[0], listed.summary);
    });
    return () => {
      generationRef.current += 1;
      controllerRef.current?.abort();
      controllerRef.current = null;
    };
  }, [applicationId, workspaceId]);

  async function createSession() {
    editor.reset();
    setSelectedSession(null);
    setMetadataStale(false);
    setUnresolved(false);
    setClientTurnKey(newClientTurnKey());
    const requestGeneration = ++generationRef.current;
    const controller = replaceController();
    setPending("create");
    setSaveResult(false);
    setResult({
      ...initialPromptApplicationSessionResult(config),
      status: "ready",
      summary: "正在从当前 exact runtime authority 创建 Prompt Session v2。",
    });
    const next = await createPromptApplicationSession(config, applicationId, controller.signal);
    if (requestGeneration !== generationRef.current || controller.signal.aborted) return;
    controllerRef.current = null;
    setPending("");
    setResult(next);
    if (next.session) {
      setSelectedSession(next.session);
      setSessions((current) => [next.session!, ...current.filter((item) => item.sessionId !== next.session!.sessionId)]);
    }
  }

  async function executeTurn() {
    if (controllerRef.current || !selectedSession || selectedSession.state !== "active" || metadataStale || !isPromptClientKey(clientTurnKey) || !editor.validation.isValid || editor.busy || pending) return;
    const requestGeneration = ++generationRef.current;
    const controller = replaceController();
    const shouldSaveResult = saveResult;
    dispatchedRef.current = false;
    setPending("execute");
    const variables = await editor.prepare();
    if (requestGeneration !== generationRef.current || controller.signal.aborted) return;
    if (!variables) { controllerRef.current = null; setPending(""); return; }
    setSaveResult(false);
    dispatchedRef.current = true;
    const next = await executePromptApplicationSessionTurn(
      config,
      selectedSession,
      variables,
      clientTurnKey,
      shouldSaveResult,
      controller.signal,
    );
    if (requestGeneration !== generationRef.current || controller.signal.aborted) return;
    controllerRef.current = null;
    setPending("");
    setResult(next);
    setMetadataStale(!next.session);
    setUnresolved(previous => promptExecutionNeedsReview(previous, next.failureCode, next.turn?.status));
    if (next.session) {
      setSelectedSession(next.session);
      setSessions((current) => current.map((item) => item.sessionId === next.session!.sessionId ? next.session! : item));
    }
    if (next.turn?.runId) onRunRecorded?.(next.turn.runId);
  }

  function selectSession(session: PromptApplicationSession, summary = "已恢复 metadata-only Session；未恢复变量值或 prompt_output。") {
    editor.reset();
    setSelectedSession(session);
    setMetadataStale(false);
    setUnresolved(false);
    generationRef.current += 1;
    controllerRef.current?.abort();
    controllerRef.current = null;
    setPending("");
    setSaveResult(false);
    setResult({
      ...initialPromptApplicationSessionResult(config),
      status: "ready",
      session,
      summary,
    });
    setClientTurnKey(newClientTurnKey());
  }

  async function refreshSession() {
    if (!selectedSession || pending || editor.busy) return;
    const requestGeneration = ++generationRef.current;
    const controller = replaceController();
    setMetadataStale(true);
    setPending("refresh");
    const next = await readPromptApplicationSession(config, applicationId, selectedSession.sessionId, controller.signal);
    if (requestGeneration !== generationRef.current || controller.signal.aborted) return;
    controllerRef.current = null;
    setPending("");
    if (!next.session) { setResult(next); return; }
    setSelectedSession(next.session);
    setMetadataStale(false);
    setSessions(current => current.map(item => item.sessionId === next.session!.sessionId ? next.session! : item));
    // Reading metadata does not resolve an interrupted execution or restore its output.
  }

  function prepareNewTurn() {
    if (pending || editor.busy || metadataStale || unresolved || !selectedSession) return;
    setClientTurnKey(newClientTurnKey());
    setSaveResult(false);
    setResult({ ...initialPromptApplicationSessionResult(config), session: selectedSession });
  }

  function cancelRequest() {
    if (pending === "execute" && dispatchedRef.current) { setMetadataStale(true); setUnresolved(true); }
    generationRef.current += 1;
    controllerRef.current?.abort();
    controllerRef.current = null;
    setPending("");
    setSaveResult(false);
    setResult((current) => ({
      ...current,
      status: "blocked",
      output: "",
      resultArtifact: null,
      resultArtifactFailureCode: "",
      failureCode: "application_session_request_canceled",
      failureSummary: "当前浏览器请求已取消；迟到响应会被丢弃。",
      summary: "当前浏览器请求已取消；未自动重试。",
    }));
  }

  function replaceController(): AbortController {
    controllerRef.current?.abort();
    const next = new AbortController();
    controllerRef.current = next;
    return next;
  }

  return (
    <section className="prompt-application-session-panel" id="prompt-application-session" aria-label={t($ => $.sessionTitle)}>
      <div className="section-heading compact-heading">
        <div><p className="eyebrow">{t($ => $.sessionTitle)}</p><h4>{t($ => $.sessionMetadata)}</h4></div>
        <span className={`status-badge ${result.turn?.status === "succeeded" ? "good" : result.failureCode ? "bad" : "neutral"}`}>{t($ => $.states[selectedSession?.state ?? result.status])}</span>
      </div>
      <div className="application-publish-layout">
        <article className="application-publish-create prompt-session-controls">
          <strong>{selectedSession?.sessionId ?? t($ => $.noSession)}</strong>
          {selectedSession ? (
            <>
              <code>{selectedSession.assignmentId} · {t($ => $.assignmentVersion, { version: selectedSession.assignmentVersion })}</code>
              <code>{selectedSession.templateId} · {t($ => $.templateVersionLabel, { version: selectedSession.templateVersion })}</code>
              <p>{t($ => $.recordTurns, { version: selectedSession.recordVersion, count: selectedSession.turnCount })}</p>
            </>
          ) : null}
          <button type="button" onClick={() => void createSession()} disabled={config.mode === "offline" || Boolean(pending) || editor.busy}>
            {pending === "create" ? t($ => $.creatingSession) : t($ => $.createSession)}
          </button>
          <button type="button" className="secondary-action" onClick={() => void refreshSession()} disabled={!selectedSession || Boolean(pending) || editor.busy}>{t($ => $.refreshSession)}</button>
          {metadataStale ? <p role="status">{t($ => $.sessionMetadataStale)}</p> : null}
          <button type="button" className="secondary-action" onClick={cancelRequest} disabled={!pending}>{t($ => $.cancelRequest)}</button>
          <div className="application-publish-list" aria-label={t($ => $.activeSessions)}>
            {sessions.map((session) => (
              <button
                type="button"
                key={session.sessionId}
                className={selectedSession?.sessionId === session.sessionId ? "is-selected" : ""}
                disabled={Boolean(pending) || editor.busy}
                onClick={() => session.sessionId === selectedSession?.sessionId ? void refreshSession() : selectSession(session)}
              >
                <strong>{session.sessionId}</strong>
                <small>{t($ => $.recordTurns, { version: session.recordVersion, count: session.turnCount })}</small>
              </button>
            ))}
          </div>
          <p className="boundary-note">{listFailureCode ? <>{t($ => $.listFailed)} <code>{listFailureCode}</code></> : pending === "list" ? t($ => $.listLoading) : t($ => $.listedCount, { count: sessions.length })}</p>
        </article>
        <article className="application-publish-review prompt-turn-controls">
          <PromptVariableInputEditor editor={editor} disabled={Boolean(pending) || editor.busy} jsonLabel={t($ => $.turnVariables)} />
          <label>client_turn_key<input disabled={Boolean(pending) || editor.busy || unresolved} value={clientTurnKey} onChange={(event) => setClientTurnKey(event.target.value)} maxLength={160} /></label>
          {!isPromptClientKey(clientTurnKey) ? <p role="alert">{t($ => $.clientKeyInvalid)}</p> : null}
          <p className="boundary-note">{t($ => $.executionKeyHelp)}</p>
          {unresolved ? <p role="status">{t($ => $.unknownExecutionHelp)} <a href="#workspace-run-history">{t($ => $.reviewRunHistory)}</a></p> : null}
          <button type="button" className="secondary-action" onClick={prepareNewTurn} disabled={!selectedSession || Boolean(pending) || editor.busy || metadataStale || unresolved}>{t($ => $.prepareNewTurn)}</button>
          <button type="button" onClick={() => void executeTurn()} disabled={!selectedSession || metadataStale || !isPromptClientKey(clientTurnKey) || !editor.validation.isValid || editor.busy || selectedSession.state !== "active" || Boolean(pending)}>
            {pending === "execute" ? t($ => $.executingTurn) : t($ => $.executeTurn)}
          </button>
        </article>
      </div>
      <ApplicationResultArtifactPanel
        config={config}
        applicationId={applicationId}
        sessionId={selectedSession?.sessionId ?? ""}
        saveResult={saveResult}
        onSaveResultChange={setSaveResult}
        latestArtifact={result.resultArtifact}
        latestArtifactFailureCode={result.resultArtifactFailureCode}
        disabled={metadataStale || !selectedSession || selectedSession.state !== "active" || Boolean(pending)}
        onOpenRun={onOpenRun}
      />
      <article className="application-publish-snapshot">
        <strong>{t($ => $.transientPromptOutput)}</strong>
        <pre>{result.output || t($ => $.noTurnOutput)}</pre>
        {result.turn ? (
          <>
            <p>{t($ => $.turnSummary, { sequence: result.turn.sequence, status: t($ => $.states[result.turn!.status]), id: result.turn.turnId })}</p>
            <button type="button" onClick={() => onOpenRun?.(result.turn!.runId)}>{t($ => $.openRunEvidence)}</button>
          </>
        ) : null}
        {result.failureCode ? <p className="failure-summary">{result.failureCode}: {result.failureCode === "prompt_invocation_output_contract_failed" ? t($ => $.outputRejected) : t($ => $.failureHelp)}</p> : null}
        <ControlledUseFailureGuidance owner="prompt_session" failureCode={result.failureCode} />
        <p className="boundary-note">{pending === "create" ? t($ => $.creatingSession) : pending === "execute" ? t($ => $.executingTurn) : result.failureCode === "application_session_request_canceled" ? t($ => $.requestCanceled) : result.status === "offline" ? t($ => $.sessionOffline) : result.failureCode ? t($ => $.sessionFailed) : result.status === "succeeded" ? t($ => $.sessionSucceeded) : result.status === "replayed" ? t($ => $.sessionReplayed) : selectedSession ? t($ => $.sessionRestored) : t($ => $.sessionCreateHelp)}</p>
      </article>
      <p className="boundary-note">
        {t($ => $.sessionBoundary)}</p>
    </section>
  );
}

function newClientTurnKey(): string {
  const suffix = (globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random()}`).replaceAll("-", "").slice(0, 16);
  return `prompt-turn-${suffix}`;
}

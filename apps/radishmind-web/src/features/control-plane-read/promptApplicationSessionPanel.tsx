import { useEffect, useMemo, useRef, useState } from "react";

import { parsePromptApplicationVariables } from "./promptApplicationInvocationConsumer.ts";
import ControlledUseFailureGuidance from "./ControlledUseFailureGuidance.tsx";
import ApplicationResultArtifactPanel from "./applicationResultArtifactPanel.tsx";
import {
  createPromptApplicationSession,
  executePromptApplicationSessionTurn,
  initialPromptApplicationSessionResult,
  listPromptApplicationSessions,
  readPromptApplicationSessionConfig,
  type PromptApplicationSession,
  type PromptApplicationSessionResult,
} from "./promptApplicationSessionConsumer.ts";

const config = readPromptApplicationSessionConfig();

export default function PromptApplicationSessionPanel({
  applicationId,
  onRunRecorded,
  onOpenRun,
}: {
  applicationId: string;
  onRunRecorded?: (runId: string) => void;
  onOpenRun?: (runId: string) => void;
}) {
  const [result, setResult] = useState<PromptApplicationSessionResult>(
    () => initialPromptApplicationSessionResult(config),
  );
  const [variablesText, setVariablesText] = useState('{"question":"请给出发布审查清单","tone":"简洁"}');
  const [clientTurnKey, setClientTurnKey] = useState(() => newClientTurnKey());
  const [sessions, setSessions] = useState<PromptApplicationSession[]>([]);
  const [listSummary, setListSummary] = useState("");
  const [saveResult, setSaveResult] = useState(false);
  const [pending, setPending] = useState<"" | "list" | "create" | "execute">("");
  const generationRef = useRef(0);
  const controllerRef = useRef<AbortController | null>(null);
  const variables = useMemo(() => parsePromptApplicationVariables(variablesText), [variablesText]);

  useEffect(() => {
    const requestGeneration = ++generationRef.current;
    controllerRef.current?.abort();
    const controller = new AbortController();
    controllerRef.current = controller;
    setResult(initialPromptApplicationSessionResult(config));
    setSessions([]);
    setListSummary("");
    setVariablesText('{"question":"请给出发布审查清单","tone":"简洁"}');
    setClientTurnKey(newClientTurnKey());
    setSaveResult(false);
    setPending(config.mode === "offline" ? "" : "list");
    void listPromptApplicationSessions(config, applicationId, controller.signal).then((listed) => {
      if (requestGeneration !== generationRef.current || controller.signal.aborted) return;
      controllerRef.current = null;
      setPending("");
      setSessions(listed.sessions);
      setListSummary(listed.summary);
      if (listed.sessions[0]) selectSession(listed.sessions[0], listed.summary);
    });
    return () => {
      generationRef.current += 1;
      controller.abort();
      controllerRef.current = null;
    };
  }, [applicationId]);

  async function createSession() {
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
      setSessions((current) => [next.session!, ...current.filter((item) => item.sessionId !== next.session!.sessionId)]);
    }
  }

  async function executeTurn() {
    if (!result.session || !variables.isValid || pending) return;
    const requestGeneration = ++generationRef.current;
    const controller = replaceController();
    const shouldSaveResult = saveResult;
    setSaveResult(false);
    setPending("execute");
    const next = await executePromptApplicationSessionTurn(
      config,
      result.session,
      variables.variables,
      clientTurnKey,
      shouldSaveResult,
      controller.signal,
    );
    if (requestGeneration !== generationRef.current || controller.signal.aborted) return;
    controllerRef.current = null;
    setPending("");
    setResult(next);
    if (next.session) {
      setSessions((current) => current.map((item) => item.sessionId === next.session!.sessionId ? next.session! : item));
    }
    if (next.turn?.runId) onRunRecorded?.(next.turn.runId);
  }

  function selectSession(session: PromptApplicationSession, summary = "已恢复 metadata-only Session；未恢复变量值或 prompt_output。") {
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
    setVariablesText('{"question":"请给出发布审查清单","tone":"简洁"}');
    setClientTurnKey(newClientTurnKey());
  }

  function cancelRequest() {
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
    <section className="prompt-application-session-panel" id="prompt-application-session" aria-label="Prompt Application Session v2">
      <div className="section-heading compact-heading">
        <div><p className="eyebrow">Prompt Application Session v2</p><h4>Metadata-only multi-turn owner</h4></div>
        <span className={`status-badge ${result.turn?.status === "succeeded" ? "good" : result.failureCode ? "bad" : "neutral"}`}>{result.session?.state ?? result.status}</span>
      </div>
      <div className="application-publish-layout">
        <article className="application-publish-create">
          <strong>{result.session?.sessionId ?? "No Prompt Session selected"}</strong>
          {result.session ? (
            <>
              <code>{result.session.assignmentId} · assignment v{result.session.assignmentVersion}</code>
              <code>{result.session.templateId} · template v{result.session.templateVersion}</code>
              <p>record v{result.session.recordVersion} · {result.session.turnCount} turn(s)</p>
            </>
          ) : null}
          <button type="button" onClick={() => void createSession()} disabled={config.mode === "offline" || Boolean(pending)}>
            {pending === "create" ? "Creating…" : "Create Session v2"}
          </button>
          <button type="button" className="secondary-action" onClick={cancelRequest} disabled={!pending}>Cancel current request</button>
          <div className="application-publish-list" aria-label="Active Prompt Session v2 records">
            {sessions.map((session) => (
              <button
                type="button"
                key={session.sessionId}
                className={result.session?.sessionId === session.sessionId ? "is-selected" : ""}
                disabled={Boolean(pending)}
                onClick={() => selectSession(session)}
              >
                <strong>{session.sessionId}</strong>
                <small>record v{session.recordVersion} · {session.turnCount} turn(s)</small>
              </button>
            ))}
          </div>
          <p className="boundary-note">{listSummary}</p>
        </article>
        <article className="application-publish-review">
          <label>Turn variables<textarea rows={6} value={variablesText} onChange={(event) => setVariablesText(event.target.value)} /></label>
          {variables.failureCode ? <p className="failure-summary">{variables.failureCode}</p> : null}
          <label>client_turn_key<input value={clientTurnKey} onChange={(event) => setClientTurnKey(event.target.value)} maxLength={160} /></label>
          <button type="button" onClick={() => void executeTurn()} disabled={!result.session || !variables.isValid || result.session.state !== "active" || Boolean(pending)}>
            {pending === "execute" ? "Executing…" : "Execute Prompt turn"}
          </button>
        </article>
      </div>
      <ApplicationResultArtifactPanel
        config={config}
        applicationId={applicationId}
        sessionId={result.session?.sessionId ?? ""}
        saveResult={saveResult}
        onSaveResultChange={setSaveResult}
        latestArtifact={result.resultArtifact}
        latestArtifactFailureCode={result.resultArtifactFailureCode}
        disabled={!result.session || result.session.state !== "active" || Boolean(pending)}
        onOpenRun={onOpenRun}
      />
      <article className="application-publish-snapshot">
        <strong>Transient prompt_output</strong>
        <pre>{result.output || "(没有当前 turn output)"}</pre>
        {result.turn ? (
          <>
            <p>turn #{result.turn.sequence} · {result.turn.status} · {result.turn.turnId}</p>
            <button type="button" onClick={() => onOpenRun?.(result.turn!.runId)}>Open Run v6 evidence</button>
          </>
        ) : null}
        {result.failureCode ? <p className="failure-summary">{result.failureCode}: {result.failureSummary}</p> : null}
        <ControlledUseFailureGuidance owner="prompt_session" failureCode={result.failureCode} />
        <p className="boundary-note">{result.summary}</p>
      </article>
      <p className="boundary-note">
        Session / Turn v2 仅持久化 authority、input digest/bytes、状态与 Run v6 引用；variables 始终易失，prompt_output 仅在逐 turn 显式选择时另存为结果资产。
      </p>
    </section>
  );
}

function newClientTurnKey(): string {
  const suffix = (globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random()}`).replaceAll("-", "").slice(0, 16);
  return `prompt-turn-${suffix}`;
}

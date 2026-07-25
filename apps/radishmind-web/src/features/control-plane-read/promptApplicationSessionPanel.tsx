import { useEffect, useMemo, useState } from "react";

import { parsePromptApplicationVariables } from "./promptApplicationInvocationConsumer.ts";
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
  const variables = useMemo(() => parsePromptApplicationVariables(variablesText), [variablesText]);

  useEffect(() => {
    const controller = new AbortController();
    let active = true;
    setResult(initialPromptApplicationSessionResult(config));
    setSessions([]);
    setListSummary("");
    setVariablesText('{"question":"请给出发布审查清单","tone":"简洁"}');
    setClientTurnKey(newClientTurnKey());
    void listPromptApplicationSessions(config, applicationId, controller.signal).then((listed) => {
      if (!active) return;
      setSessions(listed.sessions);
      setListSummary(listed.summary);
      if (listed.sessions[0]) selectSession(listed.sessions[0], listed.summary);
    });
    return () => {
      active = false;
      controller.abort();
    };
  }, [applicationId]);

  async function createSession() {
    setResult({
      ...initialPromptApplicationSessionResult(config),
      status: "ready",
      summary: "正在从当前 exact runtime authority 创建 Prompt Session v2。",
    });
    const next = await createPromptApplicationSession(config, applicationId);
    setResult(next);
    if (next.session) {
      setSessions((current) => [next.session!, ...current.filter((item) => item.sessionId !== next.session!.sessionId)]);
    }
  }

  async function executeTurn() {
    if (!result.session || !variables.isValid) return;
    const next = await executePromptApplicationSessionTurn(
      config,
      result.session,
      variables.variables,
      clientTurnKey,
    );
    setResult(next);
    if (next.session) {
      setSessions((current) => current.map((item) => item.sessionId === next.session!.sessionId ? next.session! : item));
    }
    if (next.turn?.runId) onRunRecorded?.(next.turn.runId);
  }

  function selectSession(session: PromptApplicationSession, summary = "已恢复 metadata-only Session；未恢复变量值或 prompt_output。") {
    setResult({
      ...initialPromptApplicationSessionResult(config),
      status: "ready",
      session,
      summary,
    });
    setVariablesText('{"question":"请给出发布审查清单","tone":"简洁"}');
    setClientTurnKey(newClientTurnKey());
  }

  return (
    <section className="prompt-application-session-panel" aria-label="Prompt Application Session v2">
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
          <button type="button" onClick={() => void createSession()} disabled={config.mode === "offline"}>
            Create Session v2
          </button>
          <div className="application-publish-list" aria-label="Active Prompt Session v2 records">
            {sessions.map((session) => (
              <button
                type="button"
                key={session.sessionId}
                className={result.session?.sessionId === session.sessionId ? "is-selected" : ""}
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
          <button type="button" onClick={() => void executeTurn()} disabled={!result.session || !variables.isValid || result.session.state !== "active"}>
            Execute Prompt turn
          </button>
        </article>
      </div>
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
        <p className="boundary-note">{result.summary}</p>
      </article>
      <p className="boundary-note">
        Session / Turn v2 仅持久化 authority、input digest/bytes、状态与 Run v6 引用；variables 与 prompt_output 只保留于当前请求和组件内存。
      </p>
    </section>
  );
}

function newClientTurnKey(): string {
  const suffix = (globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random()}`).replaceAll("-", "").slice(0, 16);
  return `prompt-turn-${suffix}`;
}

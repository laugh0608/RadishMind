import { useEffect, useRef, useState } from "react";

import type { ApplicationDevelopmentOwnerEvidence } from "./applicationDevelopmentReadiness.ts";
import ControlledUseFailureGuidance from "./ControlledUseFailureGuidance.tsx";
import {
  createAgentCopilotSession,
  executeAgentCopilotSessionTurn,
  initialAgentCopilotSessionResult,
  readAgentCopilotSessionConfig,
  type AgentCopilotSessionResult,
} from "./agentCopilotSessionConsumer.ts";

const config = readAgentCopilotSessionConfig();

export default function AgentCopilotSessionPanel({
  applicationId,
  applicationName,
  onRunRecorded,
  onOpenRun,
  onEvidenceChange,
}: {
  applicationId: string;
  applicationName: string;
  onRunRecorded?: (runId: string) => void;
  onOpenRun?: (runId: string) => void;
  onEvidenceChange?: (evidence: ApplicationDevelopmentOwnerEvidence) => void;
}) {
  const [result, setResult] = useState<AgentCopilotSessionResult>(
    () => initialAgentCopilotSessionResult(config),
  );
  const [task, setTask] = useState("suggest_flowsheet_edits");
  const [locale, setLocale] = useState("zh-CN");
  const [contextText, setContextText] = useState(
    '{\n  "selected_unit_ids": ["unit-101"],\n  "diagnostics": [{"code": "not_converged"}]\n}',
  );
  const [clientTurnKey, setClientTurnKey] = useState(() => createClientTurnKey());
  const [inputFailure, setInputFailure] = useState("");
  const [pending, setPending] = useState<"" | "create" | "execute">("");
  const generation = useRef(0);
  const controller = useRef<AbortController | null>(null);

  useEffect(() => {
    generation.current += 1;
    controller.current?.abort();
    controller.current = null;
    setResult(initialAgentCopilotSessionResult(config));
    setTask("suggest_flowsheet_edits");
    setLocale("zh-CN");
    setContextText('{\n  "selected_unit_ids": ["unit-101"],\n  "diagnostics": [{"code": "not_converged"}]\n}');
    setClientTurnKey(createClientTurnKey());
    setInputFailure("");
    setPending("");
    return () => {
      generation.current += 1;
      controller.current?.abort();
      controller.current = null;
    };
  }, [applicationId]);

  useEffect(() => {
    const runId = result.turn?.runId ?? "";
    const complete = result.status === "succeeded" && Boolean(runId);
    onEvidenceChange?.({
      contributionId: "controlled_run",
      status: complete ? "available" : result.failureCode ? "blocked" : "incomplete",
      coverage: runId ? "complete" : result.failureCode ? "partial" : "none",
      evidenceRefs: runId ? [{ kind: "run", id: runId }] : [],
      missingEvidence: complete ? [] : ["Record a reviewable Agent Copilot Run v7."],
      blockers: result.failureCode ? [{ code: result.failureCode, summary: result.summary }] : [],
      failureCodes: result.failureCode ? [result.failureCode] : [],
    });
  }, [onEvidenceChange, result]);

  async function createSession() {
    const requestGeneration = generation.current;
    const nextController = replaceController();
    setPending("create");
    const next = await createAgentCopilotSession(config, applicationId, nextController.signal);
    if (requestGeneration !== generation.current || nextController.signal.aborted) return;
    setResult(next);
    setPending("");
  }

  async function executeTurn() {
    if (!result.session) return;
    let context: unknown;
    try {
      context = JSON.parse(contextText);
    } catch {
      setInputFailure("Context 必须是 JSON object。");
      return;
    }
    if (!context || typeof context !== "object" || Array.isArray(context)) {
      setInputFailure("Context 必须是 JSON object。");
      return;
    }
    setInputFailure("");
    const requestGeneration = generation.current;
    const nextController = replaceController();
    setPending("execute");
    const next = await executeAgentCopilotSessionTurn(
      config,
      result.session,
      {
        task,
        locale,
        conversationId: "",
        artifacts: [],
        context: context as Record<string, unknown>,
        clientTurnKey,
      },
      nextController.signal,
    );
    if (requestGeneration !== generation.current || nextController.signal.aborted) return;
    setResult(next);
    setPending("");
    if (next.turn?.runId) onRunRecorded?.(next.turn.runId);
  }

  function cancelRequest() {
    generation.current += 1;
    controller.current?.abort();
    controller.current = null;
    setPending("");
    setResult((current) => ({
      ...current,
      response: null,
      status: "blocked",
      failureCode: "application_session_request_canceled",
      summary: "当前浏览器请求已取消；迟到响应会被丢弃。",
    }));
  }

  function replaceController(): AbortController {
    controller.current?.abort();
    const next = new AbortController();
    controller.current = next;
    return next;
  }

  return (
    <section className="prompt-application-session-panel" id="agent-copilot-invocation" aria-labelledby="agent-session-title">
      <div className="section-heading compact-heading">
        <div><p className="eyebrow">Agent Copilot · Controlled test</p><h4 id="agent-session-title">Session v3 与单次 advisory suggestion</h4></div>
        <span className={`status-badge ${result.status === "succeeded" ? "good" : result.failureCode ? "bad" : "neutral"}`}>{result.status}</span>
      </div>
      <div className="prompt-template-scope" id="agent-copilot-session">
        <article><span>Application</span><strong>{applicationName}</strong><code>{applicationId}</code></article>
        <article><span>Session</span><strong>{result.session?.sessionId ?? "not created"}</strong><code>{result.session ? `v${result.session.recordVersion}` : config.mode}</code></article>
        <article><span>Retention</span><strong>metadata only</strong><p>Context、artifact 与完整回答只在当前组件内存。</p></article>
      </div>
      <div className="application-draft-actions">
        <button type="button" onClick={() => void createSession()} disabled={Boolean(pending) || config.mode === "offline"}>{pending === "create" ? "Creating…" : "Create exact-authority Session v3"}</button>
        <button type="button" onClick={cancelRequest} disabled={!pending}>Cancel current request</button>
      </div>
      {result.session ? (
        <dl className="tenant-meta">
          <div><dt>Assignment</dt><dd>{result.session.assignmentId} · v{result.session.assignmentVersion}</dd></div>
          <div><dt>Profile</dt><dd>{result.session.profileId} · v{result.session.profileVersion}</dd></div>
          <div><dt>Project</dt><dd>{result.session.project}</dd></div>
          <div><dt>Turns</dt><dd>{result.session.turnCount}</dd></div>
        </dl>
      ) : null}
      <div className="prompt-template-layout">
        <article className="prompt-template-editor">
          <label>Canonical task<input value={task} onChange={(event) => setTask(event.target.value)} /></label>
          <label>Locale<input value={locale} onChange={(event) => setLocale(event.target.value)} /></label>
          <label>Client turn key<input value={clientTurnKey} onChange={(event) => setClientTurnKey(event.target.value)} /></label>
          <label>Transient context<textarea rows={9} value={contextText} onChange={(event) => { setContextText(event.target.value); setInputFailure(""); }} /></label>
          <button type="button" onClick={() => void executeTurn()} disabled={!result.session || Boolean(pending)}>{pending === "execute" ? "Invoking once…" : "Execute one controlled turn"}</button>
          {inputFailure ? <p className="failure-summary">{inputFailure}</p> : null}
        </article>
        <article className="prompt-template-review">
          <div className="application-api-card-heading"><div><p className="eyebrow">Current response</p><h5>{result.summary}</h5></div><span className={`status-badge ${result.response ? "good" : result.failureCode ? "bad" : "neutral"}`}>{result.response?.status ?? "none"}</span></div>
          {result.failureCode ? <p className="failure-summary">{result.failureCode} · {result.failureSummary}</p> : null}
          <ControlledUseFailureGuidance owner="agent_session" failureCode={result.failureCode} />
          {result.response ? (
            <>
              <p>{result.response.summary}</p>
              <dl className="tenant-meta">
                <div><dt>Answers</dt><dd>{result.response.answers.length}</dd></div>
                <div><dt>Issues</dt><dd>{result.response.issues.length}</dd></div>
                <div><dt>Candidate actions</dt><dd>{result.response.proposedActions.length}</dd></div>
                <div><dt>Confirmation</dt><dd>{String(result.response.requiresConfirmation)}</dd></div>
              </dl>
              {result.response.proposedActions.map((action, index) => (
                <div className="prompt-template-summary" key={`${action.kind}-${index}`}>
                  <strong>{action.title}</strong>
                  <span>{action.kind} · {action.riskLevel}</span>
                  <small>requires_confirmation={String(action.requiresConfirmation)}</small>
                </div>
              ))}
            </>
          ) : null}
          {result.turn?.runId ? <button type="button" onClick={() => onOpenRun?.(result.turn?.runId ?? "")}>Open Run v7 evidence</button> : null}
          <p className="boundary-note">响应不会写入 URL 或 browser storage；离开 stage、应用 revision 变化或取消请求会清除当前输入与回答并拒绝迟到响应。</p>
        </article>
      </div>
    </section>
  );
}

function createClientTurnKey(): string {
  return `agent-turn-${Date.now()}-${(globalThis.crypto?.randomUUID?.() ?? Math.random().toString(16).slice(2)).replaceAll("-", "").slice(0, 8)}`;
}

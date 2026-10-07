import { agentSessionMessage } from "./agentCopilotMessages.ts";
import "../../i18n/agentResources.ts";
import { useTranslation } from "react-i18next";
import { useEffect, useRef, useState } from "react";

import type { ApplicationDevelopmentOwnerEvidence } from "./applicationDevelopmentReadiness.ts";
import ApplicationResultArtifactPanel from "./applicationResultArtifactPanel.tsx";
import ActionSafetyReadPanel from "./ActionSafetyReadPanel.tsx";
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
  const { t } = useTranslation("agent");
  const [result, setResult] = useState<AgentCopilotSessionResult>(
    () => initialAgentCopilotSessionResult(config),
  );
  const [task, setTask] = useState("suggest_flowsheet_edits");
  const [locale, setLocale] = useState("zh-CN");
  const [contextText, setContextText] = useState(
    '{\n  "selected_unit_ids": ["unit-101"],\n  "diagnostics": [{"code": "not_converged"}]\n}',
  );
  const [clientTurnKey, setClientTurnKey] = useState(() => createClientTurnKey());
  const [inputFailure, setInputFailure] = useState(false);
  const [saveResult, setSaveResult] = useState(false);
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
    setInputFailure(false);
    setSaveResult(false);
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
    setSaveResult(false);
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
      setInputFailure(true);
      return;
    }
    if (!context || typeof context !== "object" || Array.isArray(context)) {
      setInputFailure(true);
      return;
    }
    setInputFailure(false);
    const requestGeneration = generation.current;
    const nextController = replaceController();
    const shouldSaveResult = saveResult;
    setSaveResult(false);
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
        saveResult: shouldSaveResult,
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
    setSaveResult(false);
    setResult((current) => ({
      ...current,
      response: null,
      actionSafety: null,
      resultArtifact: null,
      resultArtifactFailureCode: "",
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
        <div><p className="eyebrow">{t($ => $.session.eyebrow)}</p><h4 id="agent-session-title">{t($ => $.session.title)}</h4></div>
        <span className={`status-badge ${result.status === "succeeded" ? "good" : result.failureCode ? "bad" : "neutral"}`}>{t($ => $.states[result.status])}</span>
      </div>
      <div className="prompt-template-scope" id="agent-copilot-session">
        <article><span>{t($ => $.session.application)}</span><strong>{applicationName}</strong><code>{applicationId}</code></article>
        <article><span>{t($ => $.session.session)}</span><strong>{result.session?.sessionId ?? t($ => $.session.notCreated)}</strong><code>{result.session ? `v${result.session.recordVersion}` : config.mode}</code></article>
        <article><span>{t($ => $.session.retention)}</span><strong>{t($ => $.session.defaultOff)}</strong><p>{t($ => $.session.retentionNote)}</p></article>
      </div>
      <div className="application-draft-actions">
        <button type="button" onClick={() => void createSession()} disabled={Boolean(pending) || config.mode === "offline"}>{t($ => pending === "create" ? $.session.creating : $.session.create)}</button>
        <button type="button" onClick={cancelRequest} disabled={!pending}>{t($ => $.session.cancel)}</button>
      </div>
      {result.session ? (
        <dl className="tenant-meta">
          <div><dt>{t($ => $.session.assignment)}</dt><dd>{result.session.assignmentId} · v{result.session.assignmentVersion}</dd></div>
          <div><dt>{t($ => $.session.profile)}</dt><dd>{result.session.profileId} · v{result.session.profileVersion}</dd></div>
          <div><dt>{t($ => $.session.project)}</dt><dd>{result.session.project}</dd></div>
          <div><dt>{t($ => $.session.turns)}</dt><dd>{result.session.turnCount}</dd></div>
        </dl>
      ) : null}
      <div className="prompt-template-layout">
        <article className="prompt-template-editor">
          <label>{t($ => $.session.task)}<input value={task} onChange={(event) => setTask(event.target.value)} /></label>
          <p className="boundary-note" id="agent-session-locale-note">{t($ => $.session.localeNote)}</p>
          <label>{t($ => $.session.locale)}<input aria-describedby="agent-session-locale-note" value={locale} onChange={(event) => setLocale(event.target.value)} /></label>
          <label>{t($ => $.session.key)}<input value={clientTurnKey} onChange={(event) => setClientTurnKey(event.target.value)} /></label>
          <label>{t($ => $.session.context)}<textarea rows={9} value={contextText} onChange={(event) => { setContextText(event.target.value); setInputFailure(false); }} /></label>
          <button type="button" onClick={() => void executeTurn()} disabled={!result.session || Boolean(pending)}>{t($ => pending === "execute" ? $.session.invoking : $.session.execute)}</button>
          {inputFailure ? <p className="failure-summary">{t($ => $.session.invalidContext)}</p> : null}
        </article>
        <article className="prompt-template-review">
          <div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.session.response)}</p><h5>{agentSessionMessage(t, result)}</h5></div><span className={`status-badge ${result.response ? "good" : result.failureCode ? "bad" : "neutral"}`}>{t($ => $.states[result.response?.status ?? "none"])}</span></div>
          {result.failureCode ? <p className="failure-summary">{result.failureCode} · {result.failureSummary}</p> : null}
          <ControlledUseFailureGuidance owner="agent_session" failureCode={result.failureCode} />
          {result.response ? (
            <>
              <p>{result.response.summary}</p>
              <dl className="tenant-meta">
                <div><dt>{t($ => $.session.answers)}</dt><dd>{result.response.answers.length}</dd></div>
                <div><dt>{t($ => $.session.issues)}</dt><dd>{result.response.issues.length}</dd></div>
                <div><dt>{t($ => $.session.actions)}</dt><dd>{result.response.proposedActions.length}</dd></div>
                <div><dt>{t($ => $.session.confirmation)}</dt><dd>{String(result.response.requiresConfirmation)}</dd></div>
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
          <ActionSafetyReadPanel projection={result.actionSafety} title={t($ => $.session.safety)} transient />
          {result.turn?.runId ? <button type="button" onClick={() => onOpenRun?.(result.turn?.runId ?? "")}>{t($ => $.session.openRun)}</button> : null}
          <p className="boundary-note">{t($ => $.session.boundary)}</p>
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
    </section>
  );
}

function createClientTurnKey(): string {
  return `agent-turn-${Date.now()}-${(globalThis.crypto?.randomUUID?.() ?? Math.random().toString(16).slice(2)).replaceAll("-", "").slice(0, 8)}`;
}

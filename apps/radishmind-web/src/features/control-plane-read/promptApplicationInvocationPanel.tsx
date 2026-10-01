import { useTranslation } from "react-i18next";
import { useEffect, useMemo, useRef, useState } from "react";

import type { ApplicationDevelopmentOwnerEvidence } from "./applicationDevelopmentReadiness.ts";
import ControlledUseFailureGuidance from "./ControlledUseFailureGuidance.tsx";
import {
  initialPromptApplicationInvocationResult,
  invokePromptApplication,
  parsePromptApplicationVariables,
  readPromptApplicationInvocationConfig,
  type PromptApplicationInvocationResult,
} from "./promptApplicationInvocationConsumer.ts";
import {
  PROMPT_APPLICATION_CREDENTIAL_HANDOFF_EVENT,
  type PromptApplicationCredentialHandoffDetail,
} from "./promptApplicationInvocationEvents.ts";

const config = readPromptApplicationInvocationConfig();

export default function PromptApplicationInvocationPanel({
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
  const { t } = useTranslation("prompt");
  const credentialRef = useRef("");
  const abortRef = useRef<AbortController | null>(null);
  const [apiKeyId, setAPIKeyId] = useState("");
  const [variablesText, setVariablesText] = useState('{"question":"如何审查本次发布？","tone":"清晰"}');
  const [clientInvocationKey, setClientInvocationKey] = useState(() => newClientInvocationKey());
  const [result, setResult] = useState<PromptApplicationInvocationResult>(
    () => initialPromptApplicationInvocationResult(),
  );
  const variables = useMemo(() => parsePromptApplicationVariables(variablesText), [variablesText]);

  useEffect(() => {
    clearTransient();
    setVariablesText('{"question":"如何审查本次发布？","tone":"清晰"}');
    setClientInvocationKey(newClientInvocationKey());
    setResult(initialPromptApplicationInvocationResult());
  }, [applicationId]);

  useEffect(() => {
    function receiveCredential(event: Event) {
      const detail = (event as CustomEvent<PromptApplicationCredentialHandoffDetail>).detail;
      if (!detail || detail.applicationId !== applicationId) return;
      credentialRef.current = detail.token;
      setAPIKeyId(detail.apiKeyId);
      setResult({
        ...initialPromptApplicationInvocationResult(),
        summary: "一次性 Prompt Application credential 已接收，仅保留于当前组件内存。",
      });
    }
    function clearAfterRouteLeave() {
      if (window.location.hash !== "#prompt-application-invocation") clearTransient();
    }
    window.addEventListener(PROMPT_APPLICATION_CREDENTIAL_HANDOFF_EVENT, receiveCredential);
    window.addEventListener("hashchange", clearAfterRouteLeave);
    return () => {
      window.removeEventListener(PROMPT_APPLICATION_CREDENTIAL_HANDOFF_EVENT, receiveCredential);
      window.removeEventListener("hashchange", clearAfterRouteLeave);
      clearTransient(false);
    };
  }, [applicationId]);

  useEffect(() => {
    if (!onEvidenceChange) return;
    const failed = result.status === "blocked" || result.status === "failed";
    onEvidenceChange({
      contributionId: "controlled_run",
      status: result.run ? "available" : failed ? "blocked" : "incomplete",
      coverage: result.run || failed ? "complete" : "none",
      evidenceRefs: result.run
        ? [{ kind: "run", id: result.run.runId, version: result.run.recordVersion }]
        : [],
      missingEvidence: result.run ? [] : ["Record a reviewable Prompt Application Run v6."],
      blockers: failed ? [{ code: result.failureCode || "prompt_invocation_blocked", summary: result.summary }] : [],
      failureCodes: failed && result.failureCode ? [result.failureCode] : [],
    });
  }, [onEvidenceChange, result]);

  function clearTransient(updateState = true) {
    abortRef.current?.abort();
    abortRef.current = null;
    credentialRef.current = "";
    if (updateState) setAPIKeyId("");
  }

  async function invoke() {
    if (!credentialRef.current || !variables.isValid) return;
    const controller = new AbortController();
    abortRef.current = controller;
    setResult({
      ...initialPromptApplicationInvocationResult(),
      status: "running",
      summary: "服务端正在重读 exact authority、确定性渲染并执行一次 Gateway 调用。",
    });
    const next = await invokePromptApplication(
      config,
      credentialRef.current,
      variables.variables,
      clientInvocationKey,
      controller.signal,
    );
    if (abortRef.current !== controller) return;
    abortRef.current = null;
    setResult(next);
    if (next.run) onRunRecorded?.(next.run.runId);
  }

  return (
    <section className="prompt-application-invocation-panel" id="prompt-application-invocation" aria-label={t($ => $.invocationRegion)}>
      <div className="section-heading compact-heading">
        <div><p className="eyebrow">{t($ => $.invocationHeading)}</p><h4>{t($ => $.invocationTitle)}</h4></div>
        <span className={`status-badge ${result.status === "succeeded" ? "good" : result.failureCode ? "bad" : "neutral"}`}>{t($ => $.states[result.status])}</span>
      </div>
      <div className="application-publish-scope">
        <article><span>{t($ => $.application)}</span><strong>{applicationName}</strong><code>{applicationId}</code></article>
        <article><span>{t($ => $.credential)}</span><strong>{apiKeyId || t($ => $.noCredential)}</strong><p>{t($ => $.credentialPrivacy)}</p></article>
        <article><span>{t($ => $.idempotency)}</span><strong>{clientInvocationKey}</strong><p>{t($ => $.terminalRetry)}</p></article>
      </div>
      <div className="application-publish-layout">
        <article className="application-publish-create">
          <label>
            {t($ => $.templateVariables)}<textarea
              rows={8}
              value={variablesText}
              onChange={(event) => setVariablesText(event.target.value)}
              disabled={result.status === "running"}
            />
          </label>
          {variables.failureCode ? <p className="failure-summary">{variables.failureCode} · {t($ => $.invalidVariables)}</p> : null}
          <label>
            client_invocation_key
            <input
              value={clientInvocationKey}
              onChange={(event) => setClientInvocationKey(event.target.value)}
              maxLength={160}
              disabled={result.status === "running"}
            />
          </label>
          <div className="application-draft-handoff">
            <button type="button" onClick={() => void invoke()} disabled={!apiKeyId || !variables.isValid || result.status === "running"}>
              {t($ => $.invoke)}</button>
            {result.status === "running" ? <button type="button" onClick={() => abortRef.current?.abort()}>{t($ => $.cancel)}</button> : null}
            <button type="button" className="secondary-action" onClick={() => {
              clearTransient();
              setResult(initialPromptApplicationInvocationResult());
            }}>{t($ => $.clearTransient)}</button>
          </div>
        </article>
        <article className="application-publish-review">
          <strong>{t($ => $.transientOutput)}</strong>
          <pre>{result.output || t($ => $.noOutput)}</pre>
          {result.failureCode ? <p className="failure-summary">{result.failureCode}: {result.failureCode === "prompt_invocation_output_contract_failed" ? t($ => $.outputRejected) : t($ => $.failureHelp)}</p> : null}
          <ControlledUseFailureGuidance owner="prompt_invocation" failureCode={result.failureCode} />
          <p className="boundary-note">{result.status === "idle" && apiKeyId ? t($ => $.credentialReceived) : t($ => $.invocationStates[result.status])}</p>
        </article>
      </div>
      {result.run ? (
        <article className="application-publish-snapshot">
          <div className="application-api-card-heading">
            <div><p className="eyebrow">{t($ => $.runRecord)}</p><h5>{result.run.runId}</h5></div>
            <span className="status-badge good">{result.run.status}</span>
          </div>
          <code>{result.run.assignmentId} · {t($ => $.assignmentVersion, { version: result.run.assignmentVersion })}</code>
          <code>{result.run.templateId} · {t($ => $.templateVersionLabel, { version: result.run.templateVersion })}</code>
          <p>{result.run.selectedProtocol} · {result.run.selectedModel} · {t($ => $.providerCalls, { count: result.run.providerCalls })}</p>
          <button type="button" onClick={() => onOpenRun?.(result.run!.runId)}>{t($ => $.openExactRun)}</button>
        </article>
      ) : null}
      <p className="boundary-note">
        {t($ => $.invocationBoundary)}</p>
    </section>
  );
}

function newClientInvocationKey(): string {
  const suffix = (globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random()}`).replaceAll("-", "").slice(0, 16);
  return `prompt-web-${suffix}`;
}

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
    <section className="prompt-application-invocation-panel" id="prompt-application-invocation" aria-label="Prompt Application controlled invocation">
      <div className="section-heading compact-heading">
        <div><p className="eyebrow">Prompt Application Controlled Test</p><h4>Exact authority invocation and Run v6 handoff</h4></div>
        <span className={`status-badge ${result.status === "succeeded" ? "good" : result.failureCode ? "bad" : "neutral"}`}>{result.status}</span>
      </div>
      <div className="application-publish-scope">
        <article><span>Application</span><strong>{applicationName}</strong><code>{applicationId}</code></article>
        <article><span>Credential</span><strong>{apiKeyId || "not handed off"}</strong><p>raw token 仅存在于组件内存。</p></article>
        <article><span>Idempotency</span><strong>{clientInvocationKey}</strong><p>终态重试不会恢复 output。</p></article>
      </div>
      <div className="application-publish-layout">
        <article className="application-publish-create">
          <label>
            Template variables (JSON object)
            <textarea
              rows={8}
              value={variablesText}
              onChange={(event) => setVariablesText(event.target.value)}
              disabled={result.status === "running"}
            />
          </label>
          {variables.failureCode ? <p className="failure-summary">{variables.failureCode}</p> : null}
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
              执行受控调用
            </button>
            {result.status === "running" ? <button type="button" onClick={() => abortRef.current?.abort()}>取消</button> : null}
            <button type="button" className="secondary-action" onClick={() => {
              clearTransient();
              setResult(initialPromptApplicationInvocationResult());
            }}>清除 transient 状态</button>
          </div>
        </article>
        <article className="application-publish-review">
          <strong>Transient output</strong>
          <pre>{result.output || "(没有可展示的当前响应 output)"}</pre>
          {result.failureCode ? <p className="failure-summary">{result.failureCode}: {result.failureSummary}</p> : null}
          <ControlledUseFailureGuidance owner="prompt_invocation" failureCode={result.failureCode} />
          <p className="boundary-note">{result.summary}</p>
        </article>
      </div>
      {result.run ? (
        <article className="application-publish-snapshot">
          <div className="application-api-card-heading">
            <div><p className="eyebrow">Workflow Run Record v6</p><h5>{result.run.runId}</h5></div>
            <span className="status-badge good">{result.run.status}</span>
          </div>
          <code>{result.run.assignmentId} · assignment v{result.run.assignmentVersion}</code>
          <code>{result.run.templateId} · template v{result.run.templateVersion}</code>
          <p>{result.run.selectedProtocol} · {result.run.selectedModel} · provider calls {result.run.providerCalls}</p>
          <button type="button" onClick={() => onOpenRun?.(result.run!.runId)}>Open exact Run History evidence</button>
        </article>
      ) : null}
      <p className="boundary-note">
        页面不提交 model、provider、template、version 或 authority override；Run、History、Comparison、Evaluation 与 Operations 只接收 metadata。
      </p>
    </section>
  );
}

function newClientInvocationKey(): string {
  const suffix = (globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random()}`).replaceAll("-", "").slice(0, 16);
  return `prompt-web-${suffix}`;
}

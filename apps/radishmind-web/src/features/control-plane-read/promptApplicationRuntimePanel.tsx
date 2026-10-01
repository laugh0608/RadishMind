import { useTranslation } from "react-i18next";
import { useEffect, useMemo, useState } from "react";

import type { ApplicationDevelopmentOwnerEvidence } from "./applicationDevelopmentReadiness.ts";
import {
  decidePromptApplicationRuntime,
  initialPromptApplicationRuntimeState,
  readPromptApplicationRuntime,
  readPromptApplicationRuntimeConfig,
  type PromptApplicationRuntimeAction,
  type PromptApplicationRuntimeState,
} from "./promptApplicationRuntimeConsumer.ts";

const config = readPromptApplicationRuntimeConfig();

export default function PromptApplicationRuntimePanel({
  applicationId,
  publishCandidateId,
  candidateApproved,
  readOnly,
  onEvidenceChange,
}: {
  applicationId: string;
  publishCandidateId: string;
  candidateApproved: boolean;
  readOnly: boolean;
  onEvidenceChange?: (evidence: ApplicationDevelopmentOwnerEvidence) => void;
}) {
  const { t } = useTranslation("prompt");
  const [runtime, setRuntime] = useState<PromptApplicationRuntimeState>(
    () => initialPromptApplicationRuntimeState(config),
  );
  const [action, setAction] = useState<PromptApplicationRuntimeAction>("activate");

  useEffect(() => {
    setRuntime(initialPromptApplicationRuntimeState(config));
    setAction("activate");
  }, [applicationId]);

  useEffect(() => {
    if (!onEvidenceChange) return;
    const authorityReady = runtime.assignment?.state === "active" && !runtime.failureCode;
    const failed = runtime.status === "blocked" || runtime.status === "failed" || runtime.status === "version_conflict";
    onEvidenceChange({
      contributionId: "prompt_assignment",
      status: authorityReady ? "available" : failed ? "blocked" : "incomplete",
      coverage: runtime.assignment || failed ? "complete" : "none",
      evidenceRefs: runtime.assignment
        ? [{ kind: "candidate", id: runtime.assignment.candidateId, version: runtime.assignment.assignmentVersion }]
        : [],
      missingEvidence: authorityReady ? [] : ["Activate an approved Prompt Application candidate through the Runtime Assignment owner."],
      blockers: failed ? [{ code: runtime.failureCode || "prompt_runtime_blocked", summary: runtime.summary }] : [],
      failureCodes: failed && runtime.failureCode ? [runtime.failureCode] : [],
    });
  }, [onEvidenceChange, runtime]);

  const nextAction = useMemo<PromptApplicationRuntimeAction>(() => {
    if (!runtime.assignment) return "activate";
    return runtime.assignment.state === "active" ? "replace" : "revoke";
  }, [runtime.assignment]);

  useEffect(() => {
    setAction(nextAction);
  }, [nextAction]);

  const canDecide = config.mode === "dev_prompt_application_http" && !readOnly && (
    action === "revoke"
      ? runtime.assignment?.state === "active"
      : candidateApproved && (!runtime.assignment || runtime.assignment.state === "active")
  );

  async function loadRuntime() {
    setRuntime((current) => ({ ...current, status: "loading", failureCode: "", summary: "正在重读当前 Prompt runtime authority 与事件。" }));
    setRuntime(await readPromptApplicationRuntime(config, applicationId, true));
  }

  async function decide() {
    if (!canDecide) return;
    const candidateId = action === "revoke" ? "" : publishCandidateId;
    setRuntime((current) => ({ ...current, status: "loading", failureCode: "", summary: `正在执行 ${action} CAS 决策。` }));
    setRuntime(await decidePromptApplicationRuntime(
      config,
      applicationId,
      runtime.currentAssignmentVersion,
      action,
      candidateId,
    ));
  }

  return (
    <section className="prompt-application-runtime-panel" id="prompt-application-runtime-assignment" aria-label={t($ => $.runtimeRegion)}>
      <div className="application-api-card-heading">
        <div>
          <p className="eyebrow">{t($ => $.runtimeHeading)}</p>
          <h5>{t($ => $.runtimeTitle)}</h5>
        </div>
        <span className={`status-badge ${runtime.assignment?.state === "active" && !runtime.failureCode ? "good" : runtime.failureCode ? "bad" : "neutral"}`}>
          {t($ => $.states[runtime.assignment?.state ?? runtime.status])}
        </span>
      </div>

      <div className="application-publish-layout">
        <article className="application-publish-create">
          <strong>{t($ => $.currentAssignment)}</strong>
          {runtime.assignment ? (
            <>
              <code>{runtime.assignment.assignmentId} · v{runtime.assignment.assignmentVersion}</code>
              <p>{runtime.assignment.candidateId} · {t($ => $.candidateReviewVersion, { version: runtime.assignment.candidateReviewVersion })}</p>
              <code>{runtime.assignment.promptTemplateRef.templateId} · {t($ => $.templateVersionLabel, { version: runtime.assignment.promptTemplateRef.templateVersion })}</code>
              <code>{runtime.assignment.promptTemplateRef.templateDigest}</code>
            </>
          ) : (
            <p className="boundary-note">{t($ => $.noAssignment)}</p>
          )}
          <button type="button" onClick={() => void loadRuntime()} disabled={config.mode === "offline" || runtime.status === "loading"}>
            {t($ => $.reloadAssignment)}</button>
        </article>

        <article className="application-publish-review">
          <label>
            {t($ => $.casAction)}<select
              value={action}
              onChange={(event) => setAction(event.target.value as PromptApplicationRuntimeAction)}
              disabled={readOnly}
            >
              <option value="activate">{t($ => $.activateCandidate)}</option>
              <option value="replace">{t($ => $.replaceCandidate)}</option>
              <option value="revoke">{t($ => $.revokeAssignment)}</option>
            </select>
          </label>
          <p>{t($ => $.expectedAssignmentVersion)}<strong>{runtime.currentAssignmentVersion}</strong></p>
          <p>{t($ => $.candidate)}<code>{action === "revoke" ? t($ => $.omitCandidate) : publishCandidateId}</code></p>
          <button type="button" onClick={() => void decide()} disabled={!canDecide || runtime.status === "loading"}>
            {t($ => $.recordDecision)}</button>
          {runtime.failureCode ? <p className="failure-summary">{runtime.failureCode}</p> : null}
          <p className="boundary-note">{t($ => $.runtimeStates[runtime.status])}</p>
        </article>
      </div>

      <article className="application-publish-review-log">
        <div className="application-api-card-heading">
          <div><p className="eyebrow">{t($ => $.assignmentEvents)}</p><h5>{t($ => $.eventCount, { count: runtime.events.length })}</h5></div>
        </div>
        {runtime.events.length ? runtime.events.map((event) => (
          <div key={event.eventId}>
            <strong>#{event.eventSequence} · {event.action} · v{event.resultingAssignmentVersion}</strong>
            <span>{event.actorRef} · {event.occurredAt}</span>
            <p>{event.candidateId} · {event.promptTemplateRef.templateId} v{event.promptTemplateRef.templateVersion}</p>
          </div>
        )) : <p className="boundary-note">{t($ => $.noEvents)}</p>}
      </article>
      <p className="boundary-note">
        {t($ => $.runtimeBoundary)}</p>
    </section>
  );
}

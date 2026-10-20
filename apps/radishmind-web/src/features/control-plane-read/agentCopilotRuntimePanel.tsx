import { useLocalePreference } from "../../i18n/LocaleProvider.tsx";
import { formatDisplayDate } from "../../i18n/formatters.ts";
import "../../i18n/agentResources.ts";
import { useTranslation } from "react-i18next";
import { useEffect, useMemo, useState } from "react";

import type { ApplicationDevelopmentOwnerEvidence } from "./applicationDevelopmentReadiness.ts";
import ActionSafetyReadPanel from "./ActionSafetyReadPanel.tsx";
import {
  initialApplicationPublishListState,
  listApplicationPublishCandidates,
  readApplicationPublishCandidateConfig,
  type ApplicationPublishCandidateListState,
} from "./applicationPublishCandidateConsumer.ts";
import {
  decideAgentCopilotRuntime,
  initialAgentCopilotRuntimeState,
  readAgentCopilotRuntime,
  readAgentCopilotRuntimeConfig,
  type AgentCopilotRuntimeState,
} from "./agentCopilotRuntimeConsumer.ts";

const runtimeConfig = readAgentCopilotRuntimeConfig();
const publishConfig = readApplicationPublishCandidateConfig();

export default function AgentCopilotRuntimePanel({
  applicationId,
  applicationName,
  applicationActive,
  onEvidenceChange,
}: {
  applicationId: string;
  applicationName: string;
  applicationActive: boolean;
  onEvidenceChange?: (evidence: ApplicationDevelopmentOwnerEvidence) => void;
}) {
  const { t } = useTranslation("agent");
  const { locale: uiLocale } = useLocalePreference();
  const [runtime, setRuntime] = useState<AgentCopilotRuntimeState>(
    () => initialAgentCopilotRuntimeState(runtimeConfig),
  );
  const [candidates, setCandidates] = useState<ApplicationPublishCandidateListState>(
    () => initialApplicationPublishListState(publishConfig),
  );
  const [candidateId, setCandidateId] = useState("");
  const [busy, setBusy] = useState(false);
  const approved = useMemo(
    () => candidates.summaries.filter((candidate) =>
      candidate.candidateState === "approved" && Boolean(candidate.agentCopilotProfileRef)
    ),
    [candidates.summaries],
  );
  const action = runtime.assignment?.state === "active"
    ? candidateId && candidateId !== runtime.assignment.candidateId ? "replace" : "revoke"
    : "activate";
  const enabled = applicationActive && runtimeConfig.mode === "dev_agent_copilot_http";

  useEffect(() => {
    setRuntime(initialAgentCopilotRuntimeState(runtimeConfig));
    setCandidates(initialApplicationPublishListState(publishConfig));
    setCandidateId("");
  }, [applicationId]);

  useEffect(() => {
    const active = runtime.assignment?.state === "active" && !runtime.failureCode;
    const blockingFailure = runtime.status === "version_conflict" ||
      runtime.status === "blocked" ||
      runtime.status === "failed";
    onEvidenceChange?.({
      contributionId: "agent_assignment",
      status: active ? "available" : blockingFailure ? "blocked" : "incomplete",
      coverage: runtime.assignment ? "complete" : blockingFailure ? "partial" : "none",
      evidenceRefs: runtime.assignment
        ? [{ kind: "assignment", id: runtime.assignment.assignmentId, version: runtime.assignment.assignmentVersion }]
        : [],
      missingEvidence: active ? [] : ["Activate the exact approved Agent Copilot candidate."],
      blockers: blockingFailure ? [{ code: runtime.failureCode, summary: runtime.summary }] : [],
      failureCodes: blockingFailure ? [runtime.failureCode] : [],
    });
  }, [onEvidenceChange, runtime]);

  async function load() {
    setBusy(true);
    const [nextRuntime, nextCandidates] = await Promise.all([
      readAgentCopilotRuntime(runtimeConfig, applicationId),
      listApplicationPublishCandidates(publishConfig, applicationId),
    ]);
    setRuntime(nextRuntime);
    setCandidates(nextCandidates);
    setCandidateId(nextCandidates.summaries.find((candidate) =>
      candidate.candidateState === "approved" && Boolean(candidate.agentCopilotProfileRef)
    )?.candidateId ?? "");
    setBusy(false);
  }

  async function decide() {
    if (!enabled || busy) return;
    setBusy(true);
    const result = await decideAgentCopilotRuntime(
      runtimeConfig,
      applicationId,
      runtime.currentAssignmentVersion,
      action,
      candidateId,
    );
    setRuntime(result);
    setBusy(false);
  }

  return (
    <section className="prompt-application-runtime-panel" id="agent-copilot-runtime-assignment" aria-labelledby="agent-runtime-title">
      <div className="section-heading compact-heading">
        <div><p className="eyebrow">{t($ => $.runtime.eyebrow)}</p><h4 id="agent-runtime-title">{t($ => $.runtime.title)}</h4></div>
        <span className={`status-badge ${runtime.assignment?.state === "active" && !runtime.failureCode ? "good" : runtime.failureCode ? "bad" : "neutral"}`}>
          {t($ => $.states[runtime.assignment?.state ?? runtime.status])}
        </span>
      </div>
      <div className="prompt-template-scope">
        <article><span>{t($ => $.runtime.application)}</span><strong>{applicationName}</strong><code>{applicationId}</code></article>
        <article><span>{t($ => $.runtime.version)}</span><strong>{runtime.currentAssignmentVersion}</strong><code>{runtime.currentState}</code></article>
        <article><span>{t($ => $.runtime.mode)}</span><strong>{runtimeConfig.mode}</strong><p>{t($ => $.runtime.approvalNote)}</p></article>
      </div>
      <button type="button" onClick={() => void load()} disabled={busy}>{t($ => busy ? $.runtime.loading : $.runtime.load)}</button>
      {runtime.failureCode ? <p className="failure-summary">{runtime.failureCode} · {t($ => $.runtime.failure)}</p> : null}
      {runtime.assignment ? (
        <dl className="tenant-meta">
          <div><dt>{t($ => $.runtime.assignment)}</dt><dd>{runtime.assignment.assignmentId} · v{runtime.assignment.assignmentVersion}</dd></div>
          <div><dt>{t($ => $.runtime.candidate)}</dt><dd>{runtime.assignment.candidateId} · {t($ => $.runtime.reviewVersion, { version: runtime.assignment.candidateReviewVersion })}</dd></div>
          <div><dt>{t($ => $.runtime.profile)}</dt><dd>{runtime.assignment.profile.profileId} · v{runtime.assignment.profile.profileVersion}</dd></div>
          <div><dt>{t($ => $.runtime.updated)}</dt><dd title={runtime.assignment.updatedAt}>{formatDisplayDate(runtime.assignment.updatedAt, uiLocale, { timeZone: "UTC" }) ?? runtime.assignment.updatedAt}</dd></div>
        </dl>
      ) : <p className="boundary-note">{t($ => $.runtime.empty)}</p>}
      <ActionSafetyReadPanel projection={runtime.actionSafety} title={t($ => $.runtime.safety)} />
      <label>{t($ => $.runtime.approved)}<select value={candidateId} onChange={(event) => setCandidateId(event.target.value)}><option value="">{t($ => $.runtime.noCandidate)}</option>{approved.map((candidate) => <option key={candidate.candidateId} value={candidate.candidateId}>{candidate.candidateId} · profile v{candidate.agentCopilotProfileRef?.profileVersion}</option>)}</select></label>
      <button type="button" onClick={() => void decide()} disabled={!enabled || busy || (action !== "revoke" && !candidateId)}>
        {t($ => $.runtime.decide, { action: t($ => $.actions[action]), version: runtime.currentAssignmentVersion })}
      </button>
      <div className="workflow-run-history-live-list" aria-label={t($ => $.runtime.events)}>
        {runtime.events.map((event) => (
          <div className="workflow-run-history-live-row" key={event.eventId}>
            <span><strong>#{event.eventSequence} · {t($ => $.actions[event.action])}</strong><small>{event.eventId}</small></span>
            <span><small>CAS</small><strong>{event.expectedAssignmentVersion} → {event.resultingAssignmentVersion}</strong></span>
            <span><small>{t($ => $.runtime.candidate)}</small><strong>{event.candidateId || t($ => $.runtime.revoked)}</strong></span>
          </div>
        ))}
      </div>
      <p className="boundary-note">{t($ => $.runtime.boundary)}</p>
    </section>
  );
}

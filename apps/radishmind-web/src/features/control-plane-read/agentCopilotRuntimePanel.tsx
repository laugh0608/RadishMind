import { useEffect, useMemo, useState } from "react";

import type { ApplicationDevelopmentOwnerEvidence } from "./applicationDevelopmentReadiness.ts";
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
        <div><p className="eyebrow">Agent Copilot · Runtime authority</p><h4 id="agent-runtime-title">显式 assignment 与漂移审查</h4></div>
        <span className={`status-badge ${runtime.assignment?.state === "active" && !runtime.failureCode ? "good" : runtime.failureCode ? "bad" : "neutral"}`}>
          {runtime.assignment?.state ?? runtime.status}
        </span>
      </div>
      <div className="prompt-template-scope">
        <article><span>Application</span><strong>{applicationName}</strong><code>{applicationId}</code></article>
        <article><span>Current version</span><strong>{runtime.currentAssignmentVersion}</strong><code>{runtime.currentState}</code></article>
        <article><span>Runtime mode</span><strong>{runtimeConfig.mode}</strong><p>approve 不自动 activation。</p></article>
      </div>
      <button type="button" onClick={() => void load()} disabled={busy}>{busy ? "Loading…" : "Load assignment and approved candidates"}</button>
      {runtime.failureCode ? <p className="failure-summary">{runtime.failureCode} · {runtime.summary}</p> : null}
      {runtime.assignment ? (
        <dl className="tenant-meta">
          <div><dt>Assignment</dt><dd>{runtime.assignment.assignmentId} · v{runtime.assignment.assignmentVersion}</dd></div>
          <div><dt>Candidate</dt><dd>{runtime.assignment.candidateId} · review v{runtime.assignment.candidateReviewVersion}</dd></div>
          <div><dt>Profile</dt><dd>{runtime.assignment.profile.profileId} · v{runtime.assignment.profile.profileVersion}</dd></div>
          <div><dt>Updated</dt><dd>{runtime.assignment.updatedAt}</dd></div>
        </dl>
      ) : <p className="boundary-note">当前没有 assignment。请选择 approved v4 candidate 并显式 activate。</p>}
      <label>Approved Agent candidate<select value={candidateId} onChange={(event) => setCandidateId(event.target.value)}><option value="">No candidate selected</option>{approved.map((candidate) => <option key={candidate.candidateId} value={candidate.candidateId}>{candidate.candidateId} · profile v{candidate.agentCopilotProfileRef?.profileVersion}</option>)}</select></label>
      <button type="button" onClick={() => void decide()} disabled={!enabled || busy || (action !== "revoke" && !candidateId)}>
        {action} with CAS v{runtime.currentAssignmentVersion}
      </button>
      <div className="workflow-run-history-live-list" aria-label="Agent Copilot assignment events">
        {runtime.events.map((event) => (
          <div className="workflow-run-history-live-row" key={event.eventId}>
            <span><strong>#{event.eventSequence} · {event.action}</strong><small>{event.eventId}</small></span>
            <span><small>CAS</small><strong>{event.expectedAssignmentVersion} → {event.resultingAssignmentVersion}</strong></span>
            <span><small>Candidate</small><strong>{event.candidateId || "revoked"}</strong></span>
          </div>
        ))}
      </div>
      <p className="boundary-note">Assignment 仅保存 lineage metadata；不会调用 provider，也不会复制 Profile source。</p>
    </section>
  );
}

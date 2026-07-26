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
    <section className="prompt-application-runtime-panel" aria-label="Prompt Application Runtime Assignment">
      <div className="application-api-card-heading">
        <div>
          <p className="eyebrow">Prompt Runtime Assignment</p>
          <h5>显式激活、替换或撤销精确 Prompt authority</h5>
        </div>
        <span className={`status-badge ${runtime.assignment?.state === "active" && !runtime.failureCode ? "good" : runtime.failureCode ? "bad" : "neutral"}`}>
          {runtime.assignment?.state ?? runtime.status}
        </span>
      </div>

      <div className="application-publish-layout">
        <article className="application-publish-create">
          <strong>当前 assignment</strong>
          {runtime.assignment ? (
            <>
              <code>{runtime.assignment.assignmentId} · v{runtime.assignment.assignmentVersion}</code>
              <p>{runtime.assignment.candidateId} · candidate review v{runtime.assignment.candidateReviewVersion}</p>
              <code>{runtime.assignment.promptTemplateRef.templateId} · template v{runtime.assignment.promptTemplateRef.templateVersion}</code>
              <code>{runtime.assignment.promptTemplateRef.templateDigest}</code>
            </>
          ) : (
            <p className="boundary-note">尚未加载或不存在 Runtime Assignment。</p>
          )}
          <button type="button" onClick={() => void loadRuntime()} disabled={config.mode === "offline" || runtime.status === "loading"}>
            重读 assignment 与事件
          </button>
        </article>

        <article className="application-publish-review">
          <label>
            CAS action
            <select
              value={action}
              onChange={(event) => setAction(event.target.value as PromptApplicationRuntimeAction)}
              disabled={readOnly}
            >
              <option value="activate">activate approved candidate</option>
              <option value="replace">replace with approved candidate</option>
              <option value="revoke">revoke current assignment</option>
            </select>
          </label>
          <p>expected assignment version: <strong>{runtime.currentAssignmentVersion}</strong></p>
          <p>candidate: <code>{action === "revoke" ? "(must be omitted)" : publishCandidateId}</code></p>
          <button type="button" onClick={() => void decide()} disabled={!canDecide || runtime.status === "loading"}>
            记录显式 runtime 决策
          </button>
          {runtime.failureCode ? <p className="failure-summary">{runtime.failureCode}</p> : null}
          <p className="boundary-note">{runtime.summary}</p>
        </article>
      </div>

      <article className="application-publish-review-log">
        <div className="application-api-card-heading">
          <div><p className="eyebrow">Append-only assignment events</p><h5>{runtime.events.length} events</h5></div>
        </div>
        {runtime.events.length ? runtime.events.map((event) => (
          <div key={event.eventId}>
            <strong>#{event.eventSequence} · {event.action} · v{event.resultingAssignmentVersion}</strong>
            <span>{event.actorRef} · {event.occurredAt}</span>
            <p>{event.candidateId} · {event.promptTemplateRef.templateId} v{event.promptTemplateRef.templateVersion}</p>
          </div>
        )) : <p className="boundary-note">尚无已加载的 assignment event。</p>}
      </article>
      <p className="boundary-note">
        Runtime owner 每次读取都重验 candidate、draft、template digest 与应用生命周期；漂移时失败关闭，不回退旧 authority。
      </p>
    </section>
  );
}

import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import type {
  ApplicationEvaluationConfig,
  ApplicationEvaluationPlan,
  ApplicationEvaluationPlanVersion,
} from "./applicationEvaluationCampaignConsumer.ts";
import {
  createApplicationEvaluationSchedule,
  listApplicationEvaluationSchedules,
  readApplicationEvaluationScheduleOccurrence,
  readApplicationEvaluationScheduleVersion,
  reviseApplicationEvaluationSchedule,
  transitionApplicationEvaluationSchedule,
  type ApplicationEvaluationSchedule,
  type ApplicationEvaluationScheduleAction,
  type ApplicationEvaluationScheduleFailureCode,
  type ApplicationEvaluationScheduleOccurrence,
  type ApplicationEvaluationScheduleVersion,
} from "./applicationEvaluationScheduleConsumer.ts";

type PendingOperation = "create" | "revise" | ApplicationEvaluationScheduleAction;
type ScheduleView = "schedule" | "occurrence";

export default function ApplicationEvaluationScheduleOwner({
  config,
  view,
  plans,
  selectedPlan,
  selectedPlanVersion,
  applicationActive,
  onSelectPlan,
  onOpenCampaign,
}: {
  config: ApplicationEvaluationConfig;
  view: ScheduleView;
  plans: ApplicationEvaluationPlan[];
  selectedPlan: ApplicationEvaluationPlan | null;
  selectedPlanVersion: ApplicationEvaluationPlanVersion | null;
  applicationActive: boolean;
  onSelectPlan: (planId: string) => void;
  onOpenCampaign: (campaignId: string, planId: string) => void;
}) {
  const promptPlans = useMemo(
    () => plans.filter((plan) => plan.executionProfile === "prompt_application_invocation_v1"),
    [plans],
  );
  const [schedules, setSchedules] = useState<ApplicationEvaluationSchedule[]>([]);
  const [selectedScheduleId, setSelectedScheduleId] = useState("");
  const [version, setVersion] = useState<ApplicationEvaluationScheduleVersion | null>(null);
  const [occurrence, setOccurrence] = useState<ApplicationEvaluationScheduleOccurrence | null>(null);
  const [occurrenceTime, setOccurrenceTime] = useState("");
  const [quotaAPIKeyId, setQuotaAPIKeyId] = useState("");
  const [hour, setHour] = useState("02");
  const [minute, setMinute] = useState("30");
  const [pending, setPending] = useState<PendingOperation | null>(null);
  const [operationPending, setOperationPending] = useState(false);
  const [failureCode, setFailureCode] = useState<ApplicationEvaluationScheduleFailureCode | null>(null);
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(config.mode !== "offline");
  const requestGeneration = useRef(0);
  const requestAbort = useRef<AbortController | null>(null);

  const selectedSchedule = schedules.find((schedule) => schedule.scheduleId === selectedScheduleId) ?? null;
  const selectedPromptPlan = selectedPlan?.executionProfile === "prompt_application_invocation_v1" ? selectedPlan : null;
  const exactPromptVersion = selectedPromptPlan && selectedPlanVersion?.planId === selectedPromptPlan.planId &&
    selectedPlanVersion.executionProfile === "prompt_application_invocation_v1" ? selectedPlanVersion : null;
  const writeBlocked = config.mode === "offline" || !applicationActive || operationPending;
  const definitionBlocked = writeBlocked || !selectedPromptPlan || !exactPromptVersion || !quotaAPIKeyId.trim();

  const beginRequest = useCallback(() => {
    requestAbort.current?.abort();
    const controller = new AbortController();
    requestAbort.current = controller;
    return controller;
  }, []);

  const loadExactSchedule = useCallback(async (
    schedule: ApplicationEvaluationSchedule,
    generation = ++requestGeneration.current,
  ) => {
    const controller = beginRequest();
    setSelectedScheduleId(schedule.scheduleId);
    setVersion(null);
    setOccurrence(null);
    setFailureCode(null);
    setMessage("");
    setLoading(true);
    try {
      const envelope = await readApplicationEvaluationScheduleVersion(config, schedule.scheduleId, schedule.latestScheduleVersion, controller.signal);
      if (generation !== requestGeneration.current) return;
      if (envelope.failureCode || !envelope.version) {
        setFailureCode(envelope.failureCode);
        setMessage(envelope.failureSummary || "The exact immutable schedule version is unavailable.");
        return;
      }
      setVersion(envelope.version);
      setQuotaAPIKeyId(envelope.version.quotaAPIKeyId);
      setHour(String(envelope.version.schedule.hour).padStart(2, "0"));
      setMinute(String(envelope.version.schedule.minute).padStart(2, "0"));
      setOccurrenceTime(schedule.nextDueAt ?? "");
      if (promptPlans.some((plan) => plan.planId === schedule.planId) && selectedPlan?.planId !== schedule.planId) {
        onSelectPlan(schedule.planId);
      }
    } catch (error) {
      if (generation !== requestGeneration.current) return;
      if (isAbortError(error)) return;
      setMessage(strictFailure(error));
    } finally {
      if (generation === requestGeneration.current) setLoading(false);
    }
  }, [beginRequest, config, onSelectPlan, promptPlans, selectedPlan?.planId]);

  const loadScheduleOwner = useCallback(async (preferredScheduleId = "") => {
    const generation = ++requestGeneration.current;
    const controller = beginRequest();
    setPending(null);
    setLoading(config.mode !== "offline");
    setFailureCode(null);
    setMessage("");
    setOccurrence(null);
    try {
      const lifecycleStates = ["draft", "active", "paused", "archived"] as const;
      const envelopes = await Promise.all(lifecycleStates.map((lifecycleState) => (
        listApplicationEvaluationSchedules(config, controller.signal, lifecycleState)
      )));
      if (generation !== requestGeneration.current) return;
      const failedEnvelope = envelopes.find((envelope) => envelope.failureCode);
      if (failedEnvelope) {
        setSchedules([]);
        setSelectedScheduleId("");
        setVersion(null);
        setFailureCode(failedEnvelope.failureCode);
        setMessage(failedEnvelope.failureSummary);
        return;
      }
      const availableSchedules = envelopes
        .flatMap((envelope) => envelope.schedules)
        .sort((left, right) => right.updatedAt.localeCompare(left.updatedAt) || right.scheduleId.localeCompare(left.scheduleId));
      setSchedules(availableSchedules);
      const schedule = availableSchedules.find((candidate) => candidate.scheduleId === preferredScheduleId) ?? availableSchedules[0] ?? null;
      if (!schedule) {
        setSelectedScheduleId("");
        setVersion(null);
        return;
      }
      await loadExactSchedule(schedule, generation);
    } catch (error) {
      if (generation !== requestGeneration.current) return;
      if (isAbortError(error)) return;
      setSchedules([]);
      setSelectedScheduleId("");
      setVersion(null);
      setMessage(strictFailure(error));
    } finally {
      if (generation === requestGeneration.current) setLoading(false);
    }
  }, [beginRequest, config, loadExactSchedule]);

  useEffect(() => {
    void loadScheduleOwner();
    return () => {
      requestGeneration.current += 1;
      requestAbort.current?.abort();
    };
  }, [loadScheduleOwner]);

  function startNewSchedule() {
    setSelectedScheduleId("");
    setVersion(null);
    setOccurrence(null);
    setPending(null);
    setFailureCode(null);
    setMessage("New schedule draft. No server mutation has occurred.");
    setQuotaAPIKeyId("");
    setHour("02");
    setMinute("30");
  }

  function reviewOperation(kind: PendingOperation) {
    if ((kind === "create" || kind === "revise") && definitionBlocked) {
      setMessage("Choose one exact active Prompt plan, actor-owned API key id, and valid UTC cadence before review.");
      return;
    }
    if (kind !== "create" && !selectedSchedule) {
      setMessage("Select an exact schedule before reviewing this operation.");
      return;
    }
    if (!validCadence(hour, minute)) {
      setMessage("Daily UTC hour must be 00–23 and minute must be 00–59.");
      return;
    }
    setPending(kind);
    setFailureCode(null);
    setMessage("");
  }

  async function confirmOperation() {
    if (!pending || writeBlocked) return;
    setOperationPending(true);
    const controller = beginRequest();
    setFailureCode(null);
    setMessage("");
    try {
      let envelope;
      if (pending === "create" || pending === "revise") {
        if (!selectedPromptPlan || !exactPromptVersion) return;
        const input = {
          plan: selectedPromptPlan,
          version: exactPromptVersion,
          quotaAPIKeyId,
          hour: Number(hour),
          minute: Number(minute),
        };
        envelope = pending === "create"
          ? await createApplicationEvaluationSchedule(config, input, controller.signal)
          : await reviseApplicationEvaluationSchedule(config, selectedSchedule as ApplicationEvaluationSchedule, input, controller.signal);
      } else {
        envelope = await transitionApplicationEvaluationSchedule(config, selectedSchedule as ApplicationEvaluationSchedule, pending, controller.signal);
      }
      if (envelope.failureCode || !envelope.schedule) {
        setFailureCode(envelope.failureCode);
        setMessage(envelope.failureSummary || "The schedule operation failed closed.");
        return;
      }
      const operation = pending;
      setPending(null);
      await loadScheduleOwner(envelope.schedule.scheduleId);
      setMessage(`${operationLabel(operation)} completed at record v${envelope.schedule.recordVersion}.`);
    } catch (error) {
      if (isAbortError(error)) return;
      setMessage(strictFailure(error));
    } finally {
      setOperationPending(false);
    }
  }

  async function loadOccurrence() {
    if (!selectedSchedule || !version || operationPending) return;
    const generation = ++requestGeneration.current;
    const controller = beginRequest();
    setOperationPending(true);
    setOccurrence(null);
    setFailureCode(null);
    setMessage("");
    try {
      const envelope = await readApplicationEvaluationScheduleOccurrence(
        config,
        selectedSchedule.scheduleId,
        version.scheduleVersion,
        occurrenceTime.trim(),
        controller.signal,
      );
      if (generation !== requestGeneration.current) return;
      if (envelope.failureCode || !envelope.occurrence) {
        setFailureCode(envelope.failureCode);
        setMessage(envelope.failureSummary || "The exact occurrence is unavailable.");
        return;
      }
      setOccurrence(envelope.occurrence);
      setMessage(`Exact occurrence ${envelope.occurrence.scheduledForUTC} loaded; no execution or replay was requested.`);
    } catch (error) {
      if (generation !== requestGeneration.current) return;
      if (isAbortError(error)) return;
      setMessage(strictFailure(error));
    } finally {
      if (generation === requestGeneration.current) setOperationPending(false);
    }
  }

  return (
    <section
      className="application-evaluation-schedule-owner"
      id={view === "schedule" ? "application-evaluation-schedule" : "application-evaluation-occurrence"}
      data-view={view}
      data-schedule-state={selectedSchedule?.lifecycleState ?? "none"}
    >
      <ScheduleContext
        schedules={schedules}
        selectedSchedule={selectedSchedule}
        loading={loading}
        onSelect={(scheduleId) => {
          const schedule = schedules.find((candidate) => candidate.scheduleId === scheduleId);
          if (schedule) void loadExactSchedule(schedule);
        }}
        onNew={startNewSchedule}
      />

      {view === "schedule" ? (
        <ScheduleSurface
          promptPlans={promptPlans}
          selectedPlan={selectedPromptPlan}
          selectedPlanVersion={exactPromptVersion}
          schedule={selectedSchedule}
          version={version}
          quotaAPIKeyId={quotaAPIKeyId}
          hour={hour}
          minute={minute}
          pending={pending}
          writeBlocked={writeBlocked}
          definitionBlocked={definitionBlocked}
          operationPending={operationPending}
          onSelectPlan={onSelectPlan}
          onQuotaChange={setQuotaAPIKeyId}
          onHourChange={setHour}
          onMinuteChange={setMinute}
          onReview={reviewOperation}
          onCancel={() => setPending(null)}
          onConfirm={() => void confirmOperation()}
        />
      ) : (
        <OccurrenceSurface
          schedule={selectedSchedule}
          version={version}
          occurrence={occurrence}
          occurrenceTime={occurrenceTime}
          operationPending={operationPending}
          onOccurrenceTimeChange={setOccurrenceTime}
          onLoad={() => void loadOccurrence()}
          onOpenCampaign={() => {
            if (occurrence?.campaignId && selectedSchedule) onOpenCampaign(occurrence.campaignId, selectedSchedule.planId);
          }}
        />
      )}

      {failureCode || message ? (
        <div className={`application-evaluation-operation ${failureCode ? "is-failure" : ""}`} role="status">
          <strong>{failureCode ? scheduleFailureLabel(failureCode) : "Owner update"}</strong>
          <span>{message}</span>
          {failureCode === "application_evaluation_schedule_version_conflict" && selectedSchedule ? (
            <button type="button" onClick={() => void loadScheduleOwner(selectedSchedule.scheduleId)}>Reload exact owner</button>
          ) : null}
        </div>
      ) : null}

      <p className="application-evaluation-boundary-note">
        A schedule stores recurring intent and exact references, never a bearer credential. Every occurrence rechecks current account,
        membership, permissions, Plan authority, actor-owned API Key and quota; retry, replay, catch-up and production remain closed.
      </p>
    </section>
  );
}

function ScheduleContext({
  schedules,
  selectedSchedule,
  loading,
  onSelect,
  onNew,
}: {
  schedules: ApplicationEvaluationSchedule[];
  selectedSchedule: ApplicationEvaluationSchedule | null;
  loading: boolean;
  onSelect: (scheduleId: string) => void;
  onNew: () => void;
}) {
  return <header className="application-evaluation-schedule-context">
    <div>
      <span>{selectedSchedule ? "Selected schedule" : "Schedule owner"}</span>
      <label>
        <select value={selectedSchedule?.scheduleId ?? ""} onChange={(event) => onSelect(event.target.value)} disabled={loading}>
          <option value="">{loading ? "Loading exact schedules…" : "New schedule draft"}</option>
          {schedules.map((schedule) => <option key={schedule.scheduleId} value={schedule.scheduleId}>
            {schedule.scheduleId} · plan v{schedule.planVersion} · {schedule.lifecycleState}
          </option>)}
        </select>
      </label>
    </div>
    <div>
      <strong>{selectedSchedule ? `${selectedSchedule.scheduleId} · version ${selectedSchedule.latestScheduleVersion}` : "No persisted schedule selected"}</strong>
      <small>{selectedSchedule ? `${selectedSchedule.systemActorRef} + delegated user` : "review before create"}</small>
      <button type="button" className="secondary-action" onClick={onNew}>New schedule</button>
    </div>
  </header>;
}

function ScheduleSurface({
  promptPlans,
  selectedPlan,
  selectedPlanVersion,
  schedule,
  version,
  quotaAPIKeyId,
  hour,
  minute,
  pending,
  writeBlocked,
  definitionBlocked,
  operationPending,
  onSelectPlan,
  onQuotaChange,
  onHourChange,
  onMinuteChange,
  onReview,
  onCancel,
  onConfirm,
}: {
  promptPlans: ApplicationEvaluationPlan[];
  selectedPlan: ApplicationEvaluationPlan | null;
  selectedPlanVersion: ApplicationEvaluationPlanVersion | null;
  schedule: ApplicationEvaluationSchedule | null;
  version: ApplicationEvaluationScheduleVersion | null;
  quotaAPIKeyId: string;
  hour: string;
  minute: string;
  pending: PendingOperation | null;
  writeBlocked: boolean;
  definitionBlocked: boolean;
  operationPending: boolean;
  onSelectPlan: (planId: string) => void;
  onQuotaChange: (value: string) => void;
  onHourChange: (value: string) => void;
  onMinuteChange: (value: string) => void;
  onReview: (kind: PendingOperation) => void;
  onCancel: () => void;
  onConfirm: () => void;
}) {
  const state = schedule?.lifecycleState ?? "draft";
  const nextAction: ApplicationEvaluationScheduleAction | null = state === "draft" ? "activate" : state === "active" ? "pause" : state === "paused" ? "resume" : null;
  return <div className="application-evaluation-schedule-surface">
    <main className="application-evaluation-schedule-main">
      <header>
        <div><span>Exact Plan version</span><h4>{selectedPlan?.name ?? "Choose an active Prompt plan"}{selectedPlanVersion ? ` · plan v${selectedPlanVersion.planVersion}` : ""}</h4></div>
        <code>{shortRef(version?.planDigest ?? selectedPlanVersion?.planDigest ?? "")}</code>
      </header>
      <div className="application-evaluation-schedule-metrics">
        <div><strong>{schedule?.nextDueAt ? formatDate(schedule.nextDueAt) : "Daily"}</strong><span>{schedule?.nextDueAt ? `next due · ${formatTime(schedule.nextDueAt)} UTC` : `${hour.padStart(2, "0")}:${minute.padStart(2, "0")} UTC`}</span></div>
        <div><strong>{version?.itemCount ?? selectedPlan?.itemCount ?? "—"}</strong><span>fixtures</span></div>
        <div><strong>{version?.maxProviderAttempts ?? selectedPlan?.itemCount ?? "—"}</strong><span>max attempts</span></div>
        <div><strong>v{schedule?.latestScheduleVersion ?? 1}</strong><span>schedule version</span></div>
      </div>
      <dl className="application-evaluation-schedule-details">
        <div><dt>Plan</dt><dd>{schedule ? `${schedule.planId} · v${schedule.planVersion}` : selectedPlan ? `${selectedPlan.planId} · v${selectedPlan.latestPlanVersion}` : "not selected"}</dd></div>
        <div><dt>Actors</dt><dd>{schedule ? `${shortRef(schedule.systemActorRef)} + ${shortRef(schedule.delegatedByUserRef)}` : "system actor + current delegated user"}</dd></div>
        <div><dt>Quota</dt><dd>{quotaAPIKeyId || "actor-owned API key id required"}</dd></div>
        <div><dt>Cadence</dt><dd>daily UTC · {hour.padStart(2, "0")}:{minute.padStart(2, "0")}</dd></div>
      </dl>
      <div className="application-evaluation-schedule-editor">
        <label><span>Exact Prompt plan</span><select value={selectedPlan?.planId ?? ""} onChange={(event) => onSelectPlan(event.target.value)} disabled={operationPending}><option value="">Choose active Prompt plan</option>{promptPlans.map((plan) => <option key={plan.planId} value={plan.planId}>{plan.name} · v{plan.latestPlanVersion}</option>)}</select></label>
        <label><span>Actor-owned API key id</span><input value={quotaAPIKeyId} onChange={(event) => onQuotaChange(event.target.value)} placeholder="key_…" disabled={operationPending} /></label>
        <label><span>Daily UTC hour</span><input type="number" min="0" max="23" value={hour} onChange={(event) => onHourChange(event.target.value)} disabled={operationPending} /></label>
        <label><span>Minute</span><input type="number" min="0" max="59" value={minute} onChange={(event) => onMinuteChange(event.target.value)} disabled={operationPending} /></label>
      </div>
      <div className="application-evaluation-schedule-rule">
        <strong>Occurrence rule</strong>
        <p>One deterministic occurrence may create at most one existing Campaign. Missed windows are recorded without catch-up; overlap is skipped while a Campaign remains non-terminal.</p>
      </div>
    </main>

    <aside className="application-evaluation-schedule-rail">
      <header><span>Schedule authorization</span><em>{state}</em></header>
      <strong>{hour.padStart(2, "0")}:{minute.padStart(2, "0")}</strong>
      <small>next due UTC · daily</small>
      <div><span>Delegated user</span><strong>{shortRef(schedule?.delegatedByUserRef ?? "current user")}</strong><small>system actor remains separate</small></div>
      <div className="is-attention"><span>Quota consumer</span><strong>{quotaAPIKeyId || "required"}</strong><small>{version?.maxProviderAttempts ?? selectedPlan?.itemCount ?? "—"} attempts maximum</small></div>
      {!pending ? <div className="application-evaluation-schedule-actions">
        {!schedule ? <button type="button" onClick={() => onReview("create")} disabled={definitionBlocked}>Review create</button> : null}
        {schedule && (state === "draft" || state === "paused") ? <button type="button" className="secondary-action" onClick={() => onReview("revise")} disabled={definitionBlocked}>Review revision</button> : null}
        {schedule && nextAction ? <button type="button" onClick={() => onReview(nextAction)} disabled={writeBlocked}>{operationLabel(nextAction)}</button> : null}
        {schedule && state !== "archived" ? <button type="button" className="attention-action" onClick={() => onReview("archive")} disabled={writeBlocked}>Review archive</button> : null}
      </div> : <ScheduleConfirmation operation={pending} schedule={schedule} version={version} onCancel={onCancel} onConfirm={onConfirm} disabled={writeBlocked} />}
      <small>CAS · expected record v{schedule?.recordVersion ?? "new"}</small>
    </aside>
  </div>;
}

function ScheduleConfirmation({
  operation,
  schedule,
  version,
  onCancel,
  onConfirm,
  disabled,
}: {
  operation: PendingOperation;
  schedule: ApplicationEvaluationSchedule | null;
  version: ApplicationEvaluationScheduleVersion | null;
  onCancel: () => void;
  onConfirm: () => void;
  disabled: boolean;
}) {
  const recurring = operation === "activate" || operation === "resume" || operation === "create" || operation === "revise";
  return <div className="application-evaluation-confirmation">
    <span>CONFIRM {operation.toUpperCase()}</span>
    <strong>{schedule?.scheduleId ?? "New daily UTC schedule"}</strong>
    <p>{recurring
      ? `This records recurring intent for at most ${version?.maxProviderAttempts ?? "the exact Plan item count"} Provider attempts per occurrence.`
      : operation === "archive" ? "No future occurrences will be created. Existing versions and occurrences remain readable." : "The active cadence is paused without replay or catch-up."}</p>
    <p>Every future occurrence still revalidates current authority and quota. No credential, retry, replay, release or deploy is authorized.</p>
    <div><button type="button" className="secondary-action" onClick={onCancel}>Cancel</button><button type="button" onClick={onConfirm} disabled={disabled}>Confirm {operation}</button></div>
  </div>;
}

function OccurrenceSurface({
  schedule,
  version,
  occurrence,
  occurrenceTime,
  operationPending,
  onOccurrenceTimeChange,
  onLoad,
  onOpenCampaign,
}: {
  schedule: ApplicationEvaluationSchedule | null;
  version: ApplicationEvaluationScheduleVersion | null;
  occurrence: ApplicationEvaluationScheduleOccurrence | null;
  occurrenceTime: string;
  operationPending: boolean;
  onOccurrenceTimeChange: (value: string) => void;
  onLoad: () => void;
  onOpenCampaign: () => void;
}) {
  const terminal = occurrence ? ["succeeded", "failed", "interrupted", "skipped"].includes(occurrence.state) : false;
  return <div className="application-evaluation-occurrence-surface">
    <header>
      <div><span>Exact occurrence read</span><h4>{occurrence ? formatTimestamp(occurrence.scheduledForUTC) : "Enter the canonical scheduled UTC time"}</h4></div>
      <em>{occurrence?.state ?? "read only"}</em>
    </header>
    <div className="application-evaluation-occurrence-reader">
      <label><span>Schedule version</span><input value={version ? `v${version.scheduleVersion}` : "—"} readOnly /></label>
      <label><span>scheduled_for_utc</span><input value={occurrenceTime} onChange={(event) => onOccurrenceTimeChange(event.target.value)} placeholder="2026-08-31T02:30:00Z" disabled={operationPending} /></label>
      <button type="button" onClick={onLoad} disabled={!schedule || !version || !occurrenceTime.trim() || operationPending}>Read exact occurrence</button>
    </div>
    {occurrence ? <>
      <div className="application-evaluation-occurrence-metrics">
        <div><strong>{occurrence.campaignId ? 1 : 0}</strong><span>Campaign refs</span></div>
        <div><strong>{occurrence.failureCode ? 1 : 0}</strong><span>fail-closed result</span></div>
        <div><strong>0</strong><span>client replays</span></div>
        <div><strong>v{occurrence.recordVersion}</strong><span>record version</span></div>
      </div>
      <dl className="application-evaluation-occurrence-details">
        <div><dt>Exact schedule</dt><dd>{occurrence.scheduleId} · v{occurrence.scheduleVersion}</dd></div>
        <div><dt>Deterministic key</dt><dd>{occurrence.clientCampaignKey}</dd></div>
        <div><dt>Actors</dt><dd>{shortRef(occurrence.systemActorRef)} + {shortRef(occurrence.delegatedByUserRef)}</dd></div>
        <div><dt>Claimed</dt><dd>{occurrence.claimedAt ? formatTimestamp(occurrence.claimedAt) : "not claimed"}</dd></div>
        <div><dt>Failure code</dt><dd>{occurrence.failureCode ?? "none"}</dd></div>
        <div><dt>Campaign</dt><dd>{occurrence.campaignId ?? "none created"}</dd></div>
      </dl>
      <div className={`application-evaluation-occurrence-result is-${occurrence.state}`}>
        <div><span>{terminal ? "Terminal evidence" : "Open occurrence"}</span><strong>{occurrence.state}</strong></div>
        {occurrence.campaignId ? <button type="button" onClick={onOpenCampaign}>Open exact Campaign</button> : <span>No Campaign handoff exists; the browser cannot create or replay one.</span>}
      </div>
    </> : <div className="application-evaluation-pair-empty"><strong>No occurrence loaded</strong><span>The API intentionally has no occurrence list route. Read only the exact schedule version and canonical UTC identity you intend to inspect.</span></div>}
  </div>;
}

function validCadence(hour: string, minute: string): boolean {
  const hourValue = Number(hour);
  const minuteValue = Number(minute);
  return Number.isInteger(hourValue) && hourValue >= 0 && hourValue <= 23 &&
    Number.isInteger(minuteValue) && minuteValue >= 0 && minuteValue <= 59;
}

function operationLabel(operation: PendingOperation): string {
  const labels: Record<PendingOperation, string> = {
    create: "Create schedule",
    revise: "Revise schedule",
    activate: "Activate schedule",
    pause: "Pause schedule",
    resume: "Resume schedule",
    archive: "Archive schedule",
  };
  return labels[operation];
}

function scheduleFailureLabel(code: ApplicationEvaluationScheduleFailureCode): string {
  if (code.includes("version_conflict")) return "Version conflict";
  if (code.includes("membership_denied")) return "Membership revoked";
  if (code.includes("authority_changed") || code.includes("plan_changed")) return "Authority changed";
  if (code.includes("quota")) return "Quota admission denied";
  if (code.includes("missed_window")) return "Missed window recorded";
  if (code.includes("overlap_blocked")) return "Overlap skipped";
  if (code.includes("store")) return "Schedule store unavailable";
  if (code.includes("not_found")) return "Exact record unavailable";
  return "Schedule request failed closed";
}

function formatTimestamp(value: string): string {
  return value.replace("T", " · ").replace("Z", " UTC");
}

function formatDate(value: string): string {
  return new Intl.DateTimeFormat("en", { day: "2-digit", month: "short", timeZone: "UTC" }).format(new Date(value));
}

function formatTime(value: string): string {
  return new Intl.DateTimeFormat("en", { hour: "2-digit", minute: "2-digit", hour12: false, timeZone: "UTC" }).format(new Date(value));
}

function shortRef(value: string): string {
  if (!value) return "—";
  return value.length <= 28 ? value : `${value.slice(0, 13)}…${value.slice(-10)}`;
}

function strictFailure(error: unknown): string {
  return error instanceof Error
    ? `${error.message} No response data was accepted and no execution was retried.`
    : "Schedule strict validation failed. No response data was accepted.";
}

function isAbortError(error: unknown): boolean {
  return error instanceof Error && error.name === "AbortError";
}

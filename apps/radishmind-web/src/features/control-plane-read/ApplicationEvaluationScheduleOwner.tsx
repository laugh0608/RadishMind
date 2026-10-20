import "../../i18n/evaluationScheduleResources.ts";
import { useTranslation } from "react-i18next";
import { campaignState, campaignOperation, campaignFailure, campaignMessage, type CampaignMessage } from "./applicationEvaluationMessages.ts";
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
  const { t } = useTranslation("evaluation");
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
  const [message, setMessage] = useState<CampaignMessage | null>(null);
  const [loading, setLoading] = useState(config.mode !== "offline");
  const requestGeneration = useRef(0);
  const scopeGeneration = useRef(0);
  const operationInFlight = useRef(false);
  const planSelection = useRef({ promptPlans, selectedPlan, onSelectPlan });
  planSelection.current = { promptPlans, selectedPlan, onSelectPlan };
  const requestAbort = useRef<AbortController | null>(null);

  const selectedSchedule = schedules.find((schedule) => schedule.scheduleId === selectedScheduleId) ?? null;
  const selectedPromptPlan = selectedPlan?.executionProfile === "prompt_application_invocation_v1" ? selectedPlan : null;
  const exactPromptVersion = selectedPromptPlan && selectedPlanVersion?.planId === selectedPromptPlan.planId &&
    selectedPlanVersion.executionProfile === "prompt_application_invocation_v1" ? selectedPlanVersion : null;
  const writeBlocked = config.mode === "offline" || !applicationActive || operationPending || loading ||
    failureCode === "application_evaluation_schedule_version_conflict";
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
    setPending(null);
    setVersion(null);
    setOccurrence(null);
    setFailureCode(null);
    setMessage(null);
    setLoading(true);
    try {
      const envelope = await readApplicationEvaluationScheduleVersion(config, schedule.scheduleId, schedule.latestScheduleVersion, controller.signal);
      if (generation !== requestGeneration.current) return;
      if (envelope.failureCode || !envelope.version) {
        setFailureCode(envelope.failureCode);
        setMessage(envelope.failureCode ? null : { kind: "strictFailure" });
        return;
      }
      setVersion(envelope.version);
      setQuotaAPIKeyId(envelope.version.quotaAPIKeyId);
      setHour(String(envelope.version.schedule.hour).padStart(2, "0"));
      setMinute(String(envelope.version.schedule.minute).padStart(2, "0"));
      setOccurrenceTime(schedule.nextDueAt ?? "");
      const selection = planSelection.current;
      if (selection.promptPlans.some((plan) => plan.planId === schedule.planId) && selection.selectedPlan?.planId !== schedule.planId) {
        selection.onSelectPlan(schedule.planId);
      }
    } catch (error) {
      if (generation !== requestGeneration.current) return;
      if (isAbortError(error)) return;
      setMessage({ kind: "strictFailure" });
    } finally {
      if (generation === requestGeneration.current) setLoading(false);
    }
  }, [beginRequest, config]);

  const loadScheduleOwner = useCallback(async (preferredScheduleId = "") => {
    const generation = ++requestGeneration.current;
    const controller = beginRequest();
    setPending(null);
    setLoading(config.mode !== "offline");
    setFailureCode(null);
    setMessage(null);
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
        setMessage(null);
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
      setMessage({ kind: "strictFailure" });
    } finally {
      if (generation === requestGeneration.current) setLoading(false);
    }
  }, [beginRequest, config, loadExactSchedule]);

  useEffect(() => {
    operationInFlight.current = false;
    setOperationPending(false);
    void loadScheduleOwner();
    return () => {
      requestGeneration.current += 1;
      scopeGeneration.current += 1;
      requestAbort.current?.abort();
    };
  }, [loadScheduleOwner]);

  function startNewSchedule() {
    if (operationInFlight.current) return;
    requestGeneration.current += 1;
    requestAbort.current?.abort();
    setLoading(false);
    setSelectedScheduleId("");
    setVersion(null);
    setOccurrence(null);
    setPending(null);
    setFailureCode(null);
    setMessage({ kind: "newScheduleMessage" });
    setQuotaAPIKeyId("");
    setHour("02");
    setMinute("30");
  }

  function reviewOperation(kind: PendingOperation) {
    if ((kind === "create" || kind === "revise") && definitionBlocked) {
      setMessage({ kind: "chooseScheduleInputs" });
      return;
    }
    if (kind !== "create" && !selectedSchedule) {
      setMessage({ kind: "selectSchedule" });
      return;
    }
    if (!validCadence(hour, minute)) {
      setMessage({ kind: "invalidCadence" });
      return;
    }
    setPending(kind);
    setFailureCode(null);
    setMessage(null);
  }

  async function confirmOperation() {
    if (!pending || writeBlocked || operationInFlight.current) return;
    operationInFlight.current = true;
    const generation = scopeGeneration.current;
    setOperationPending(true);
    const controller = beginRequest();
    setFailureCode(null);
    setMessage(null);
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
      if (generation !== scopeGeneration.current) return;
      if (envelope.failureCode || !envelope.schedule) {
        setPending(null);
        setFailureCode(envelope.failureCode);
        setMessage(envelope.failureCode ? null : { kind: "strictFailure" });
        return;
      }
      const operation = pending;
      setPending(null);
      await loadScheduleOwner(envelope.schedule.scheduleId);
      if (generation !== scopeGeneration.current) return;
      setMessage({ kind: "scheduleSavedMessage", operation, version: envelope.schedule.recordVersion });
    } catch (error) {
      if (generation !== scopeGeneration.current) return;
      if (isAbortError(error)) return;
      setMessage({ kind: "strictFailure" });
    } finally {
      if (generation === scopeGeneration.current) {
        operationInFlight.current = false;
        setOperationPending(false);
      }
    }
  }

  async function loadOccurrence() {
    if (!selectedSchedule || !version || operationInFlight.current) return;
    operationInFlight.current = true;
    const generation = ++requestGeneration.current;
    const controller = beginRequest();
    setOperationPending(true);
    setOccurrence(null);
    setFailureCode(null);
    setMessage(null);
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
        setMessage(envelope.failureCode ? null : { kind: "strictFailure" });
        return;
      }
      setOccurrence(envelope.occurrence);
      setMessage({ kind: "occurrenceLoadedMessage", time: envelope.occurrence.scheduledForUTC });
    } catch (error) {
      if (generation !== requestGeneration.current) return;
      if (isAbortError(error)) return;
      setMessage({ kind: "strictFailure" });
    } finally {
      if (generation === requestGeneration.current) {
        operationInFlight.current = false;
        setOperationPending(false);
      }
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
        loading={loading || operationPending}
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

      {operationPending || loading ? <p role="status">{t($ => $.campaign.pendingOperation)}</p> : null}
      {!loading && !failureCode && schedules.length === 0 ? <p>{t($ => $.campaign.schedulesEmpty)}</p> : null}
      {failureCode || message ? (
        <div className={`application-evaluation-operation ${failureCode ? "is-failure" : ""}`} role="status">
          <strong>{failureCode ? campaignFailure(t, failureCode) : t($ => $.campaign.ownerUpdate)}</strong>
          <span>{campaignMessage(t, message)}</span>{failureCode ? <code>{failureCode}</code> : null}
          {failureCode || message?.kind === "strictFailure" ? (
            <button type="button" onClick={() => void loadScheduleOwner(selectedSchedule?.scheduleId ?? "")} disabled={operationPending}>{t($ => $.campaign.reloadOwner)}</button>
          ) : null}
        </div>
      ) : null}

      <p className="application-evaluation-boundary-note">{t($ => $.campaign.scheduleBoundary)}</p>
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
  const { t } = useTranslation("evaluation");
  return <header className="application-evaluation-schedule-context">
    <div>
      <span>{selectedSchedule ? t($ => $.campaign.selectedSchedule) : t($ => $.campaign.scheduleOwner)}</span>
      <label>
        <select aria-label={t($ => $.campaign.selectedSchedule)} value={selectedSchedule?.scheduleId ?? ""} onChange={(event) => onSelect(event.target.value)} disabled={loading}>
          <option value="">{loading ? t($ => $.campaign.schedulesLoading) : t($ => $.campaign.newScheduleDraft)}</option>
          {schedules.map((schedule) => <option key={schedule.scheduleId} value={schedule.scheduleId}>
            {t($ => $.campaign.scheduleOption, { id: schedule.scheduleId, version: schedule.planVersion, state: campaignState(t, schedule.lifecycleState) })}
          </option>)}
        </select>
      </label>
    </div>
    <div>
      <strong>{selectedSchedule ? t($ => $.campaign.scheduleSelected, { id: selectedSchedule.scheduleId, version: selectedSchedule.latestScheduleVersion }) : t($ => $.campaign.noSchedule)}</strong>
      <small>{selectedSchedule ? t($ => $.campaign.delegation, { actor: selectedSchedule.systemActorRef }) : t($ => $.campaign.reviewBeforeCreate)}</small>
      <button type="button" className="secondary-action" onClick={onNew} disabled={loading}>{t($ => $.campaign.newSchedule)}</button>
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
  const { t } = useTranslation("evaluation");
  const { i18n } = useTranslation("evaluation");
  const state = schedule?.lifecycleState ?? "draft";
  const nextAction: ApplicationEvaluationScheduleAction | null = state === "draft" ? "activate" : state === "active" ? "pause" : state === "paused" ? "resume" : null;
  return <div className="application-evaluation-schedule-surface">
    <main className="application-evaluation-schedule-main">
      <header>
        <div><span>{t($ => $.campaign.exactPlanVersion)}</span><h4>{selectedPlan?.name ?? t($ => $.campaign.choosePromptPlan)}{selectedPlanVersion ? t($ => $.campaign.planSuffix, { version: selectedPlanVersion.planVersion }) : ""}</h4></div>
        <code>{shortRef(version?.planDigest ?? selectedPlanVersion?.planDigest ?? "")}</code>
      </header>
      <div className="application-evaluation-schedule-metrics">
        <div><strong>{schedule?.nextDueAt ? formatDate(schedule.nextDueAt, i18n.language) : t($ => $.campaign.daily)}</strong><span>{schedule?.nextDueAt ? t($ => $.campaign.nextDue, { time: formatTime(schedule.nextDueAt) }) : `${hour.padStart(2, "0")}:${minute.padStart(2, "0")} UTC`}</span></div>
        <div><strong>{version?.itemCount ?? selectedPlan?.itemCount ?? "—"}</strong><span>{t($ => $.campaign.fixtures)}</span></div>
        <div><strong>{version?.maxProviderAttempts ?? selectedPlan?.itemCount ?? "—"}</strong><span>{t($ => $.campaign.maxAttempts)}</span></div>
        <div><strong>v{schedule?.latestScheduleVersion ?? 1}</strong><span>{t($ => $.campaign.scheduleVersion)}</span></div>
      </div>
      <dl className="application-evaluation-schedule-details">
        <div><dt>{t($ => $.campaign.plan)}</dt><dd>{schedule ? `${schedule.planId} · v${schedule.planVersion}` : selectedPlan ? `${selectedPlan.planId} · v${selectedPlan.latestPlanVersion}` : t($ => $.campaign.notSelected)}</dd></div>
        <div><dt>{t($ => $.campaign.actors)}</dt><dd>{schedule ? `${shortRef(schedule.systemActorRef)} + ${shortRef(schedule.delegatedByUserRef)}` : t($ => $.campaign.systemDelegation)}</dd></div>
        <div><dt>{t($ => $.campaign.quota)}</dt><dd>{quotaAPIKeyId || t($ => $.campaign.keyRequired)}</dd></div>
        <div><dt>{t($ => $.campaign.cadence)}</dt><dd>{t($ => $.campaign.dailyCadence, { time: `${hour.padStart(2, "0")}:${minute.padStart(2, "0")}` })}</dd></div>
      </dl>
      <div className="application-evaluation-schedule-editor">
        <label><span>{t($ => $.campaign.exactPromptPlan)}</span><select value={selectedPlan?.planId ?? ""} onChange={(event) => onSelectPlan(event.target.value)} disabled={operationPending}><option value="">{t($ => $.campaign.chooseActivePromptPlan)}</option>{promptPlans.map((plan) => <option key={plan.planId} value={plan.planId}>{plan.name} · v{plan.latestPlanVersion}</option>)}</select></label>
        <label><span>{t($ => $.campaign.actorKey)}</span><input value={quotaAPIKeyId} onChange={(event) => onQuotaChange(event.target.value)} placeholder="key_…" disabled={operationPending} /></label>
        <label><span>{t($ => $.campaign.dailyHour)}</span><input type="number" min="0" max="23" value={hour} onChange={(event) => onHourChange(event.target.value)} disabled={operationPending} /></label>
        <label><span>{t($ => $.campaign.minute)}</span><input type="number" min="0" max="59" value={minute} onChange={(event) => onMinuteChange(event.target.value)} disabled={operationPending} /></label>
      </div>
      <div className="application-evaluation-schedule-rule">
        <strong>{t($ => $.campaign.occurrenceRule)}</strong>
        <p>{t($ => $.campaign.occurrenceRuleBody)}</p>
      </div>
    </main>

    <aside className="application-evaluation-schedule-rail">
      <header><span>{t($ => $.campaign.scheduleAuthorization)}</span><em>{campaignState(t, state)}</em></header>
      <strong>{hour.padStart(2, "0")}:{minute.padStart(2, "0")}</strong>
      <small>{t($ => $.campaign.nextDaily)}</small>
      <div><span>{t($ => $.campaign.delegatedUser)}</span><strong>{shortRef(schedule?.delegatedByUserRef ?? t($ => $.campaign.currentUser))}</strong><small>{t($ => $.campaign.separateActor)}</small></div>
      <div className="is-attention"><span>{t($ => $.campaign.quotaConsumer)}</span><strong>{quotaAPIKeyId || t($ => $.campaign.required)}</strong><small>{version || selectedPlan ? t($ => $.campaign.maxAttemptCount, { count: version?.maxProviderAttempts ?? selectedPlan?.itemCount ?? 0 }) : t($ => $.campaign.unknownAttempts)}</small></div>
      {!pending ? <div className="application-evaluation-schedule-actions">
        {!schedule ? <button type="button" onClick={() => onReview("create")} disabled={definitionBlocked}>{t($ => $.campaign.reviewCreate)}</button> : null}
        {schedule && (state === "draft" || state === "paused") ? <button type="button" className="secondary-action" onClick={() => onReview("revise")} disabled={definitionBlocked}>{t($ => $.campaign.reviewRevision)}</button> : null}
        {schedule && nextAction ? <button type="button" onClick={() => onReview(nextAction)} disabled={writeBlocked}>{campaignOperation(t, nextAction)}</button> : null}
        {schedule && state !== "archived" ? <button type="button" className="attention-action" onClick={() => onReview("archive")} disabled={writeBlocked}>{t($ => $.campaign.reviewArchive)}</button> : null}
      </div> : <ScheduleConfirmation operation={pending} schedule={schedule} version={version} onCancel={onCancel} onConfirm={onConfirm} disabled={writeBlocked} />}
      <small>{schedule ? t($ => $.campaign.expectedRecord, { version: schedule.recordVersion }) : t($ => $.campaign.newRecord)}</small>
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
  const { t } = useTranslation("evaluation");
  const recurring = operation === "activate" || operation === "resume" || operation === "create" || operation === "revise";
  return <div className="application-evaluation-confirmation">
    <span>{t($ => $.campaign.confirmOperation, { operation: campaignOperation(t, operation) })}</span>
    <strong>{schedule?.scheduleId ?? t($ => $.campaign.newDailySchedule)}</strong>
    <p>{recurring
      ? version ? t($ => $.campaign.recurringConfirmation, { count: version.maxProviderAttempts }) : t($ => $.campaign.recurringPlanConfirmation)
      : operation === "archive" ? t($ => $.campaign.scheduleArchive) : t($ => $.campaign.schedulePause)}</p>
    <p>{t($ => $.campaign.scheduleConfirmBoundary)}</p>
    <div><button type="button" className="secondary-action" onClick={onCancel} disabled={disabled}>{t($ => $.campaign.cancel)}</button><button type="button" onClick={onConfirm} disabled={disabled}>{t($ => $.campaign.confirmOperation, { operation: campaignOperation(t, operation) })}</button></div>
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
  const { t } = useTranslation("evaluation");
  const terminal = occurrence ? ["succeeded", "failed", "interrupted", "skipped"].includes(occurrence.state) : false;
  return <div className="application-evaluation-occurrence-surface">
    <header>
      <div><span>{t($ => $.campaign.exactOccurrence)}</span><h4>{occurrence ? formatTimestamp(occurrence.scheduledForUTC) : t($ => $.campaign.enterUTC)}</h4></div>
      <em>{occurrence ? campaignState(t, occurrence.state) : t($ => $.campaign.readOnly)}</em>
    </header>
    <div className="application-evaluation-occurrence-reader">
      <label><span>{t($ => $.campaign.scheduleVersionLabel)}</span><input value={version ? `v${version.scheduleVersion}` : "—"} readOnly /></label>
      <label><span>scheduled_for_utc</span><input value={occurrenceTime} onChange={(event) => onOccurrenceTimeChange(event.target.value)} placeholder="2026-08-31T02:30:00Z" disabled={operationPending} /></label>
      <button type="button" onClick={onLoad} disabled={!schedule || !version || !occurrenceTime.trim() || operationPending}>{t($ => $.campaign.readOccurrence)}</button>
    </div>
    {occurrence ? <>
      <div className="application-evaluation-occurrence-metrics">
        <div><strong>{occurrence.campaignId ? 1 : 0}</strong><span>{t($ => $.campaign.campaignRefs)}</span></div>
        <div><strong>{occurrence.failureCode ? 1 : 0}</strong><span>{t($ => $.campaign.failClosedResult)}</span></div>
        <div><strong>0</strong><span>{t($ => $.campaign.clientReplays)}</span></div>
        <div><strong>v{occurrence.recordVersion}</strong><span>{t($ => $.campaign.recordVersionLabel)}</span></div>
      </div>
      <dl className="application-evaluation-occurrence-details">
        <div><dt>{t($ => $.campaign.exactSchedule)}</dt><dd>{occurrence.scheduleId} · v{occurrence.scheduleVersion}</dd></div>
        <div><dt>{t($ => $.campaign.deterministicKey)}</dt><dd>{occurrence.clientCampaignKey}</dd></div>
        <div><dt>{t($ => $.campaign.actors)}</dt><dd>{shortRef(occurrence.systemActorRef)} + {shortRef(occurrence.delegatedByUserRef)}</dd></div>
        <div><dt>{t($ => $.campaign.claimed)}</dt><dd>{occurrence.claimedAt ? formatTimestamp(occurrence.claimedAt) : t($ => $.campaign.notClaimed)}</dd></div>
        <div><dt>{t($ => $.campaign.failureCode)}</dt><dd>{occurrence.failureCode ? `${campaignFailure(t, occurrence.failureCode)} · ${occurrence.failureCode}` : t($ => $.campaign.none)}</dd></div>
        <div><dt>{t($ => $.campaign.campaign)}</dt><dd>{occurrence.campaignId ?? t($ => $.campaign.noneCreated)}</dd></div>
      </dl>
      <div className={`application-evaluation-occurrence-result is-${occurrence.state}`}>
        <div><span>{terminal ? t($ => $.campaign.terminalEvidence) : t($ => $.campaign.openOccurrence)}</span><strong>{campaignState(t, occurrence.state)}</strong></div>
        {occurrence.campaignId ? <button type="button" onClick={onOpenCampaign}>{t($ => $.campaign.openCampaign)}</button> : <span>{t($ => $.campaign.noCampaignHandoff)}</span>}
      </div>
    </> : <div className="application-evaluation-pair-empty"><strong>{t($ => $.campaign.noOccurrence)}</strong><span>{t($ => $.campaign.occurrenceReadBoundary)}</span></div>}
  </div>;
}

function validCadence(hour: string, minute: string): boolean {
  if (!hour.trim() || !minute.trim()) return false;
  const hourValue = Number(hour);
  const minuteValue = Number(minute);
  return Number.isInteger(hourValue) && hourValue >= 0 && hourValue <= 23 &&
    Number.isInteger(minuteValue) && minuteValue >= 0 && minuteValue <= 59;
}

function formatTimestamp(value: string): string {
  return value.replace("T", " · ").replace("Z", " UTC");
}

function formatDate(value: string, locale: string): string {
  return new Intl.DateTimeFormat(locale, { day: "2-digit", month: "short", timeZone: "UTC" }).format(new Date(value));
}

function formatTime(value: string): string {
  return new Intl.DateTimeFormat("en", { hour: "2-digit", minute: "2-digit", hour12: false, timeZone: "UTC" }).format(new Date(value));
}

function shortRef(value: string): string {
  if (!value) return "—";
  return value.length <= 28 ? value : `${value.slice(0, 13)}…${value.slice(-10)}`;
}

function isAbortError(error: unknown): boolean {
  return error instanceof Error && error.name === "AbortError";
}

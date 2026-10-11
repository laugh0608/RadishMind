import "../../i18n/evaluationCampaignResources.ts";
import { useTranslation } from "react-i18next";
import { campaignFailure, campaignMessage, type CampaignMessage } from "./applicationEvaluationMessages.ts";
import { lazy, Suspense, useCallback, useEffect, useMemo, useRef, useState } from "react";

import {
  applicationEvaluationPlanTemplate,
  archiveApplicationEvaluationPlan,
  createApplicationEvaluationPlan,
  executeApplicationEvaluationCampaign,
  listApplicationEvaluationCampaigns,
  listApplicationEvaluationPlans,
  materializeApplicationEvaluationHandoff,
  parseApplicationEvaluationPlanDraft,
  previewApplicationEvaluationPair,
  readApplicationEvaluationConfig,
  readApplicationEvaluationPlanVersion,
  reconcileApplicationEvaluationCampaign,
  reviseApplicationEvaluationPlan,
  serializeApplicationEvaluationItems,
  serializeApplicationEvaluationTarget,
  type ApplicationEvaluationCampaign,
  type ApplicationEvaluationFailureCode,
  type ApplicationEvaluationPairEnvelope,
  type ApplicationEvaluationPlan,
  type ApplicationEvaluationPlanDraft,
  type ApplicationEvaluationPlanVersion,
  type ApplicationEvaluationProfile,
} from "./applicationEvaluationCampaignConsumer.ts";

type EvaluationTask = "plan" | "schedule" | "occurrence" | "campaign" | "pair" | "handoff";
export type LoadState = "offline" | "loading" | "ready" | "empty" | "failed";
export type PendingPlanOperation =
  | { kind: "create" | "revise"; draft: ApplicationEvaluationPlanDraft }
  | { kind: "archive"; draft: null };

const TASKS: Array<{ id: EvaluationTask; number: string; anchor: string }> = [
  { id: "plan", number: "01", anchor: "application-evaluation-plan" },
  { id: "schedule", number: "02", anchor: "application-evaluation-schedule" },
  { id: "occurrence", number: "03", anchor: "application-evaluation-occurrence" },
  { id: "campaign", number: "04", anchor: "application-evaluation-campaign" },
];
const taskSummaryKeys = { plan: "taskPlanSummary", schedule: "taskScheduleSummary", occurrence: "taskOccurrenceSummary", campaign: "taskCampaignSummary" } as const;
const TASK_ROUTES: Array<{ id: EvaluationTask; anchor: string }> = [
  ...TASKS,
  { id: "pair", anchor: "application-evaluation-pair" },
  { id: "handoff", anchor: "application-evaluation-handoff" },
];
const CampaignOwner = lazy(() => import("./ApplicationEvaluationCampaignOwner.tsx"));
const PlanList = lazy(() => import("./ApplicationEvaluationPlanOwner.tsx").then(module => ({ default: module.PlanList })));
const PlanOwner = lazy(() => import("./ApplicationEvaluationPlanOwner.tsx").then(module => ({ default: module.PlanOwner })));
const ApplicationEvaluationScheduleOwner = lazy(() => import("./ApplicationEvaluationScheduleOwner.tsx"));
const ApplicationEvaluationPairHandoffOwner = lazy(() => import("./ApplicationEvaluationPairHandoffOwner.tsx"));

export default function ApplicationEvaluationCampaignPanel({
  applicationId,
  applicationName,
  applicationKind,
  workspaceId,
  applicationActive,
}: {
  applicationId: string;
  applicationName: string;
  applicationKind: string;
  workspaceId: string;
  applicationActive: boolean;
}) {
  const { t } = useTranslation("evaluation");
  const config = useMemo(
    () => readApplicationEvaluationConfig({ workspaceId, applicationId }),
    [applicationId, workspaceId],
  );
  const initialTemplate = useMemo(() => applicationEvaluationPlanTemplate(profileForKind(applicationKind)), [applicationKind]);
  const [activeTask, setActiveTask] = useState<EvaluationTask>(() => taskForHash(window.location.hash));
  const [loadState, setLoadState] = useState<LoadState>(config.mode === "offline" ? "offline" : "loading");
  const [plans, setPlans] = useState<ApplicationEvaluationPlan[]>([]);
  const [selectedPlanId, setSelectedPlanId] = useState("");
  const [selectedVersion, setSelectedVersion] = useState<ApplicationEvaluationPlanVersion | null>(null);
  const [campaigns, setCampaigns] = useState<ApplicationEvaluationCampaign[]>([]);
  const [selectedCampaignId, setSelectedCampaignId] = useState("");
  const [failureCode, setFailureCode] = useState<ApplicationEvaluationFailureCode | null>(null);
  const [message, setMessage] = useState<CampaignMessage | null>(null);
  const [operationPending, setOperationPending] = useState(false);
  const [planName, setPlanName] = useState(initialTemplate.name);
  const [profile, setProfile] = useState<ApplicationEvaluationProfile>(initialTemplate.executionProfile);
  const [targetJSON, setTargetJSON] = useState(() => serializeApplicationEvaluationTarget(initialTemplate.target));
  const [itemsJSON, setItemsJSON] = useState(() => serializeApplicationEvaluationItems(initialTemplate.items));
  const [pendingPlanOperation, setPendingPlanOperation] = useState<PendingPlanOperation | null>(null);
  const [quotaAPIKeyId, setQuotaAPIKeyId] = useState("");
  const [clientCampaignKey, setClientCampaignKey] = useState(() => newCampaignKey(applicationId));
  const [campaignConfirmation, setCampaignConfirmation] = useState(false);
  const [baselineCampaignId, setBaselineCampaignId] = useState("");
  const [candidateCampaignId, setCandidateCampaignId] = useState("");
  const [pairEnvelope, setPairEnvelope] = useState<ApplicationEvaluationPairEnvelope | null>(null);
  const [handoffConfirmation, setHandoffConfirmation] = useState(false);
  const requestGeneration = useRef(0);
  const scopeGeneration = useRef(0);
  const operationInFlight = useRef(false);

  const selectedPlan = plans.find((plan) => plan.planId === selectedPlanId) ?? null;
  const selectedCampaign = campaigns.find((campaign) => campaign.campaignId === selectedCampaignId) ?? null;
  const terminalCampaigns = campaigns.filter((campaign) => campaign.state === "succeeded");
  const writeBlocked = config.mode === "offline" || !applicationActive || operationPending ||
    failureCode === "application_evaluation_version_conflict" || failureCode === "application_evaluation_authority_changed";

  const loadPlanOwner = useCallback(async (preferredPlanId = "") => {
    const generation = ++requestGeneration.current;
    setLoadState(config.mode === "offline" ? "offline" : "loading");
    setFailureCode(null);
    setMessage(null);
    setSelectedVersion(null);
    setCampaigns([]);
    setPairEnvelope(null);
    try {
      const envelope = await listApplicationEvaluationPlans(config);
      if (generation !== requestGeneration.current) return;
      if (envelope.failureCode) {
        setPlans([]);
        setSelectedPlanId("");
        setFailureCode(envelope.failureCode);
        setLoadState(config.mode === "offline" ? "offline" : "failed");
        return;
      }
      setPlans(envelope.plans);
      const nextPlanId = envelope.plans.some((plan) => plan.planId === preferredPlanId)
        ? preferredPlanId
        : envelope.plans[0]?.planId ?? "";
      setSelectedPlanId(nextPlanId);
      setLoadState(envelope.plans.length ? "ready" : "empty");
      if (nextPlanId) await loadExactPlan(nextPlanId, envelope.plans, generation);
    } catch {
      if (generation !== requestGeneration.current) return;
      setPlans([]);
      setSelectedPlanId("");
      setLoadState("failed");
      setMessage({ kind: "strictFailure" });
    }
  }, [config]);

  const loadExactPlan = useCallback(async (
    planId: string,
    currentPlans = plans,
    generation = ++requestGeneration.current,
  ) => {
    const plan = currentPlans.find((candidate) => candidate.planId === planId);
    if (!plan) return;
    setSelectedPlanId(planId);
    setPendingPlanOperation(null);
    setCampaignConfirmation(false);
    setHandoffConfirmation(false);
    setSelectedVersion(null);
    setCampaigns([]);
    setSelectedCampaignId("");
    setPairEnvelope(null);
    setFailureCode(null);
    setMessage(null);
    try {
      const [versionEnvelope, campaignEnvelope] = await Promise.all([
        readApplicationEvaluationPlanVersion(config, planId, plan.latestPlanVersion),
        listApplicationEvaluationCampaigns(config, planId),
      ]);
      if (generation !== requestGeneration.current) return;
      if (versionEnvelope.failureCode || !versionEnvelope.version) {
        setFailureCode(versionEnvelope.failureCode);
        setMessage(versionEnvelope.failureCode ? null : { kind: "strictFailure" });
        return;
      }
      if (campaignEnvelope.failureCode) {
        setFailureCode(campaignEnvelope.failureCode);
        setMessage(null);
        return;
      }
      setSelectedVersion(versionEnvelope.version);
      setCampaigns(campaignEnvelope.campaigns);
      setSelectedCampaignId(campaignEnvelope.campaigns[0]?.campaignId ?? "");
      selectDefaultPair(campaignEnvelope.campaigns);
    } catch {
      if (generation !== requestGeneration.current) return;
      setMessage({ kind: "strictFailure" });
    }
  }, [config, plans]);

  useEffect(() => {
    operationInFlight.current = false;
    setOperationPending(false);
    void loadPlanOwner();
    return () => {
      requestGeneration.current += 1;
      scopeGeneration.current += 1;
    };
  }, [loadPlanOwner]);

  useEffect(() => {
    function synchronizeTask() {
      const task = taskForHash(window.location.hash);
      setActiveTask(task);
      if (TASK_ROUTES.some((candidate) => `#${candidate.anchor}` === window.location.hash)) {
        window.requestAnimationFrame(() => {
          document.querySelector<HTMLElement>(".application-evaluation-workspace")?.scrollIntoView({ block: "start" });
        });
      }
    }
    synchronizeTask();
    window.addEventListener("hashchange", synchronizeTask);
    return () => window.removeEventListener("hashchange", synchronizeTask);
  }, []);

  function chooseTask(task: EvaluationTask) {
    const target = TASK_ROUTES.find((candidate) => candidate.id === task);
    if (!target) return;
    setActiveTask(task);
    window.location.hash = target.anchor;
  }

  async function openExactScheduledCampaign(campaignId: string, planId: string) {
    const generation = requestGeneration.current + 1;
    await loadExactPlan(planId);
    if (generation !== requestGeneration.current) return;
    setSelectedCampaignId(campaignId);
    chooseTask("campaign");
  }

  function applyTemplate(nextProfile: ApplicationEvaluationProfile) {
    const template = applicationEvaluationPlanTemplate(nextProfile);
    setProfile(nextProfile);
    setPlanName(template.name);
    setTargetJSON(serializeApplicationEvaluationTarget(template.target));
    setItemsJSON(serializeApplicationEvaluationItems(template.items));
    setPendingPlanOperation(null);
    setMessage(null);
  }

  function loadVersionIntoEditor() {
    if (!selectedVersion) return;
    setPlanName(selectedVersion.name);
    setProfile(selectedVersion.executionProfile);
    setTargetJSON(serializeApplicationEvaluationTarget(selectedVersion.target));
    setItemsJSON(serializeApplicationEvaluationItems(selectedVersion.items));
    setPendingPlanOperation(null);
    setMessage({ kind: "planLoadedMessage", id: selectedVersion.planId, version: selectedVersion.planVersion });
  }

  function reviewPlan(kind: "create" | "revise") {
    try {
      const draft = parseApplicationEvaluationPlanDraft(planName, profile, targetJSON, itemsJSON);
      if (kind === "revise" && (!selectedPlan || !selectedVersion)) {
        setMessage({ kind: "selectPlan" });
        return;
      }
      setPendingPlanOperation({ kind, draft });
      setMessage(null);
    } catch {
      setPendingPlanOperation(null);
      setMessage({ kind: "invalidPlan" });
    }
  }

  async function confirmPlanOperation() {
    if (!pendingPlanOperation || writeBlocked) return;
    if (operationInFlight.current) return;
    operationInFlight.current = true;
    const generation = scopeGeneration.current;
    setOperationPending(true);
    setFailureCode(null);
    setMessage(null);
    try {
      if (pendingPlanOperation.kind === "archive") {
        if (!selectedPlan) return;
        const envelope = await archiveApplicationEvaluationPlan(config, selectedPlan.planId, selectedPlan.recordVersion);
        if (generation !== scopeGeneration.current) return;
        if (envelope.failureCode) {
          applyFailure(envelope.failureCode);
        } else {
          setPendingPlanOperation(null);
          await loadPlanOwner();
          if (generation !== scopeGeneration.current) return;
          setMessage({ kind: "planArchivedMessage", id: selectedPlan.planId });
        }
        return;
      }
      const envelope = pendingPlanOperation.kind === "create"
        ? await createApplicationEvaluationPlan(config, pendingPlanOperation.draft)
        : await reviseApplicationEvaluationPlan(
          config,
          selectedPlan?.planId ?? "",
          selectedPlan?.recordVersion ?? 0,
          pendingPlanOperation.draft,
        );
      if (generation !== scopeGeneration.current) return;
      if (envelope.failureCode || !envelope.plan || !envelope.version) {
        applyFailure(envelope.failureCode);
        return;
      }
      setPendingPlanOperation(null);
      await loadPlanOwner(envelope.plan.planId);
      if (generation !== scopeGeneration.current) return;
      setMessage({ kind: "planSavedMessage", operation: pendingPlanOperation.kind, id: envelope.plan.planId, version: envelope.version.planVersion });
    } catch {
      if (generation !== scopeGeneration.current) return;
      setMessage({ kind: "strictFailure" });
    } finally {
      if (generation === scopeGeneration.current) {
        operationInFlight.current = false;
        setOperationPending(false);
      }
    }
  }

  async function confirmCampaign() {
    if (!campaignConfirmation || !selectedPlan || !selectedVersion || writeBlocked) return;
    if (operationInFlight.current) return;
    operationInFlight.current = true;
    const generation = scopeGeneration.current;
    setOperationPending(true);
    setFailureCode(null);
    setMessage(null);
    try {
      const envelope = await executeApplicationEvaluationCampaign(config, {
        plan: selectedPlan,
        version: selectedVersion,
        clientCampaignKey,
        quotaAPIKeyId,
      });
      if (generation !== scopeGeneration.current) return;
      if (envelope.failureCode || !envelope.campaign) {
        applyFailure(envelope.failureCode);
        return;
      }
      setCampaignConfirmation(false);
      setClientCampaignKey(newCampaignKey(applicationId));
      await loadExactPlan(selectedPlan.planId);
      if (generation !== scopeGeneration.current) return;
      setSelectedCampaignId(envelope.campaign.campaignId);
      setMessage({ kind: "campaignFinishedMessage", id: envelope.campaign.campaignId, state: envelope.campaign.state });
    } catch {
      if (generation !== scopeGeneration.current) return;
      setMessage({ kind: "strictFailure" });
    } finally {
      if (generation === scopeGeneration.current) {
        operationInFlight.current = false;
        setOperationPending(false);
      }
    }
  }

  async function reconcileCampaign() {
    if (!selectedCampaign || writeBlocked) return;
    if (operationInFlight.current) return;
    operationInFlight.current = true;
    const generation = scopeGeneration.current;
    setOperationPending(true);
    setFailureCode(null);
    setMessage(null);
    try {
      const envelope = await reconcileApplicationEvaluationCampaign(config, selectedCampaign);
      if (generation !== scopeGeneration.current) return;
      if (envelope.failureCode || !envelope.campaign) {
        applyFailure(envelope.failureCode);
        return;
      }
      if (selectedPlan) await loadExactPlan(selectedPlan.planId);
      if (generation !== scopeGeneration.current) return;
      setMessage({ kind: "campaignReconciledMessage", id: envelope.campaign.campaignId, state: envelope.campaign.state });
    } catch {
      if (generation !== scopeGeneration.current) return;
      setMessage({ kind: "strictFailure" });
    } finally {
      if (generation === scopeGeneration.current) {
        operationInFlight.current = false;
        setOperationPending(false);
      }
    }
  }

  async function previewPair() {
    if (operationInFlight.current) return;
    operationInFlight.current = true;
    const generation = scopeGeneration.current;
    setOperationPending(true);
    setFailureCode(null);
    setPairEnvelope(null);
    setHandoffConfirmation(false);
    setMessage(null);
    try {
      const envelope = await previewApplicationEvaluationPair(config, baselineCampaignId, candidateCampaignId);
      if (generation !== scopeGeneration.current) return;
      setPairEnvelope(envelope);
      if (envelope.failureCode) applyFailure(envelope.failureCode);
    } catch {
      if (generation !== scopeGeneration.current) return;
      setMessage({ kind: "strictFailure" });
    } finally {
      if (generation === scopeGeneration.current) {
        operationInFlight.current = false;
        setOperationPending(false);
      }
    }
  }

  async function confirmHandoff() {
    const baseline = campaigns.find((campaign) => campaign.campaignId === baselineCampaignId);
    const candidate = campaigns.find((campaign) => campaign.campaignId === candidateCampaignId);
    if (!handoffConfirmation || !baseline || !candidate || writeBlocked) return;
    if (operationInFlight.current) return;
    operationInFlight.current = true;
    const generation = scopeGeneration.current;
    setOperationPending(true);
    setFailureCode(null);
    setMessage(null);
    try {
      const envelope = await materializeApplicationEvaluationHandoff(config, baseline, candidate);
      if (generation !== scopeGeneration.current) return;
      setPairEnvelope(envelope);
      setHandoffConfirmation(false);
      if (envelope.failureCode && envelope.failureCode !== "application_evaluation_handoff_partial") {
        applyFailure(envelope.failureCode);
        return;
      }
      setMessage(envelope.handoff ? { kind: "handoffMessage", state: envelope.handoff.state, count: envelope.handoff.caseRefs.length, suite: envelope.handoff.suiteId } : { kind: "strictFailure" });
    } catch {
      if (generation !== scopeGeneration.current) return;
      setMessage({ kind: "strictFailure" });
    } finally {
      if (generation === scopeGeneration.current) {
        operationInFlight.current = false;
        setOperationPending(false);
      }
    }
  }

  function applyFailure(code: ApplicationEvaluationFailureCode | null) {
    setFailureCode(code);
    setMessage(code ? null : { kind: "strictFailure" });
    if (code === "application_evaluation_version_conflict" || code === "application_evaluation_authority_changed") {
      setCampaignConfirmation(false);
      setPendingPlanOperation(null);
      setHandoffConfirmation(false);
    }
  }

  function selectDefaultPair(nextCampaigns: ApplicationEvaluationCampaign[]) {
    const terminal = nextCampaigns.filter((campaign) => campaign.state === "succeeded");
    setBaselineCampaignId(terminal[1]?.campaignId ?? terminal[0]?.campaignId ?? "");
    setCandidateCampaignId(terminal[0]?.campaignId ?? "");
  }

  return (
    <section
      className="application-evaluation-workspace"
      id="application-evaluation-campaign"
      aria-labelledby="application-evaluation-title"
      data-active-task={activeTask}
      data-load-state={loadState}
      data-source={config.mode}
    >
      <header className="application-evaluation-heading">
        <div>
          <p className="eyebrow">{t($ => $.campaign.eyebrow)}</p>
          <h3 id="application-evaluation-title">{t($ => $.campaign.title)}</h3>
          <p>{t($ => $.campaign.intro)}</p>
        </div>
        <div className="application-evaluation-boundaries">
          <span className="status-badge neutral">{config.environment}</span>
          <span className={`status-badge ${applicationActive ? "good" : "neutral"}`}>
            {applicationActive ? t($ => $.campaign.activeApplication) : t($ => $.campaign.archivedApplication)}
          </span>
        </div>
      </header>

      <dl className="application-evaluation-scope">
        <div><dt>{t($ => $.campaign.tenant)}</dt><dd>{config.tenantRef}</dd></div>
        <div><dt>{t($ => $.campaign.workspace)}</dt><dd>{config.workspaceId}</dd></div>
        <div><dt>{t($ => $.campaign.environment)}</dt><dd>{config.environment}</dd></div>
        <div><dt>{t($ => $.campaign.application)}</dt><dd>{applicationId}</dd></div>
      </dl>

      <div className="application-evaluation-stage-shell">
        <aside className="application-evaluation-path">
          <span>{t($ => $.campaign.scheduleOwner)}</span>
          <div className="application-evaluation-task-switcher" role="tablist" aria-label={t($ => $.campaign.path)}>
            {TASKS.map((task) => (
              <button
                key={task.id}
                type="button"
                role="tab"
                aria-selected={activeTask === task.id || task.id === "campaign" && (activeTask === "pair" || activeTask === "handoff")}
                onClick={() => chooseTask(task.id)}
                disabled={operationPending}
              >
                <b>{task.number}</b><span><strong>{t($ => $.campaign[task.id as "plan" | "schedule" | "occurrence" | "campaign"])}</strong><small>{t($ => $.campaign[taskSummaryKeys[task.id as keyof typeof taskSummaryKeys]])}</small></span>
              </button>
            ))}
          </div>
          <p><strong>{t($ => $.campaign.devOnly)}</strong><span>{t($ => $.campaign.pathBoundary)}</span></p>
        </aside>

        <div className="application-evaluation-stage-owner">
          {activeTask === "plan" ? <Suspense fallback={<p>{t($ => $.campaign.planLoading)}</p>}><div className="application-evaluation-workbench">
            <PlanList
              plans={plans}
              selectedPlanId={selectedPlanId}
              loadState={loadState}
              disabled={operationPending}
              onSelect={(planId) => void loadExactPlan(planId)}
            />
            <main className="application-evaluation-owner">
              <PlanOwner
                profile={profile}
                planName={planName}
                targetJSON={targetJSON}
                itemsJSON={itemsJSON}
                selectedPlan={selectedPlan}
                selectedVersion={selectedVersion}
                pending={pendingPlanOperation}
                writeBlocked={writeBlocked}
                operationPending={operationPending}
                onProfileChange={applyTemplate}
                onPlanNameChange={setPlanName}
                onTargetJSONChange={setTargetJSON}
                onItemsJSONChange={setItemsJSON}
                onLoadVersion={loadVersionIntoEditor}
                onReview={reviewPlan}
                onReviewArchive={() => setPendingPlanOperation({ kind: "archive", draft: null })}
                onCancel={() => setPendingPlanOperation(null)}
                onConfirm={() => void confirmPlanOperation()}
              />
            </main>
          </div></Suspense> : null}

          {activeTask === "schedule" || activeTask === "occurrence" ? (
            <Suspense fallback={<div className="application-evaluation-pair-empty"><strong>{t($ => $.campaign.scheduleLoading)}</strong><span>{t($ => $.campaign.consumerLoading)}</span></div>}>
              <ApplicationEvaluationScheduleOwner
                config={config}
                view={activeTask}
                plans={plans}
                selectedPlan={selectedPlan}
                selectedPlanVersion={selectedVersion}
                applicationActive={applicationActive}
                onSelectPlan={(planId) => void loadExactPlan(planId)}
                onOpenCampaign={(campaignId, planId) => void openExactScheduledCampaign(campaignId, planId)}
              />
            </Suspense>
          ) : null}

          {activeTask === "campaign" ? <main className="application-evaluation-owner"><Suspense fallback={<p>{t($ => $.campaign.evidenceLoading)}</p>}>
            <CampaignOwner
              plan={selectedPlan}
              version={selectedVersion}
              campaigns={campaigns}
              selectedCampaign={selectedCampaign}
              quotaAPIKeyId={quotaAPIKeyId}
              clientCampaignKey={clientCampaignKey}
              confirmed={campaignConfirmation}
              writeBlocked={writeBlocked}
              operationPending={operationPending}
              onSelectCampaign={setSelectedCampaignId}
              onNewCampaign={() => {
                setSelectedCampaignId("");
                setCampaignConfirmation(false);
                setClientCampaignKey(newCampaignKey(applicationId));
              }}
              onQuotaAPIKeyIdChange={setQuotaAPIKeyId}
              onClientCampaignKeyChange={setClientCampaignKey}
              onReview={() => setCampaignConfirmation(true)}
              onCancel={() => setCampaignConfirmation(false)}
              onConfirm={() => void confirmCampaign()}
              onReconcile={() => void reconcileCampaign()}
              onOpenHandoff={() => chooseTask("handoff")}
            />
          </Suspense></main> : null}

          {activeTask === "pair" || activeTask === "handoff" ? <main className="application-evaluation-owner">
            <Suspense fallback={<div className="application-evaluation-pair-empty"><strong>{t($ => $.campaign.evidenceLoading)}</strong><span>{t($ => $.campaign.refsUnchanged)}</span></div>}>
              <ApplicationEvaluationPairHandoffOwner
                task={activeTask}
                campaigns={terminalCampaigns}
                baselineCampaignId={baselineCampaignId}
                candidateCampaignId={candidateCampaignId}
                envelope={pairEnvelope}
                confirmed={handoffConfirmation}
                writeBlocked={writeBlocked}
                operationPending={operationPending}
                onBaselineChange={(value) => { setBaselineCampaignId(value); setPairEnvelope(null); setHandoffConfirmation(false); }}
                onCandidateChange={(value) => { setCandidateCampaignId(value); setPairEnvelope(null); setHandoffConfirmation(false); }}
                onPreview={() => void previewPair()}
                onReviewHandoff={() => setHandoffConfirmation(true)}
                onCancelHandoff={() => setHandoffConfirmation(false)}
                onConfirmHandoff={() => void confirmHandoff()}
              />
            </Suspense>
          </main> : null}
        </div>
      </div>

      {operationPending ? <p role="status">{t($ => $.campaign.pendingOperation)}</p> : null}
      {failureCode || message ? (
        <div className={`application-evaluation-operation ${failureCode ? "is-failure" : ""}`} role="status">
          <strong>{failureCode ? campaignFailure(t, failureCode) : t($ => $.campaign.ownerUpdate)}</strong>
          <span>{campaignMessage(t, message)}</span>{failureCode ? <code>{failureCode}</code> : null}
          {failureCode || message?.kind === "strictFailure" ? (
            <button type="button" onClick={() => void loadPlanOwner(selectedPlanId)} disabled={operationPending}>{t($ => $.campaign.reloadOwner)}</button>
          ) : null}
        </div>
      ) : null}

      <p className="application-evaluation-boundary-note">{t($ => $.campaign.boundary)}</p>
    </section>
  );
}

function profileForKind(kind: string): ApplicationEvaluationProfile {
  if (kind === "prompt_application") return "prompt_application_invocation_v1";
  if (kind === "agent") return "agent_copilot_suggestion_v1";
  if (kind === "docs_qa") return "application_rag_invocation_v1";
  return "workflow_definition_executor_v1";
}

function taskForHash(hash: string): EvaluationTask {
  const normalized = hash.trim().replace(/^#/u, "");
  return TASK_ROUTES.find((task) => task.anchor === normalized)?.id ?? "schedule";
}

function newCampaignKey(applicationId: string): string {
  const suffix = `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`;
  return `evaluation-${applicationId.replace(/[^A-Za-z0-9._:-]/gu, "-")}-${suffix}`.slice(0, 150);
}

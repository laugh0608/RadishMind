import { lazy, Suspense, useCallback, useEffect, useMemo, useRef, useState } from "react";

import {
  APPLICATION_EVALUATION_PROFILES,
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
import StructuredRuntimeInputEditor from "./StructuredRuntimeInputEditor.tsx";
import {
  structuredRuntimeInputAuthorityKey,
  validateStructuredRuntimeInputDrafts,
  type StructuredRuntimeInputDrafts,
  type StructuredRuntimeInputValues,
} from "./structuredRuntimeInput.ts";

type EvaluationTask = "plan" | "schedule" | "occurrence" | "campaign" | "pair" | "handoff";
type LoadState = "offline" | "loading" | "ready" | "empty" | "failed";
type PendingPlanOperation =
  | { kind: "create" | "revise"; draft: ApplicationEvaluationPlanDraft }
  | { kind: "archive"; draft: null };

const TASKS: Array<{ id: EvaluationTask; number: string; label: string; summary: string; anchor: string }> = [
  { id: "plan", number: "01", label: "Plan", summary: "Exact version", anchor: "application-evaluation-plan" },
  { id: "schedule", number: "02", label: "Schedule", summary: "Daily UTC", anchor: "application-evaluation-schedule" },
  { id: "occurrence", number: "03", label: "Occurrence", summary: "Claim and observe", anchor: "application-evaluation-occurrence" },
  { id: "campaign", number: "04", label: "Campaign", summary: "Exact handoff", anchor: "application-evaluation-campaign" },
];
const TASK_ROUTES: Array<{ id: EvaluationTask; anchor: string }> = [
  ...TASKS,
  { id: "pair", anchor: "application-evaluation-pair" },
  { id: "handoff", anchor: "application-evaluation-handoff" },
];
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
  const [message, setMessage] = useState("");
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

  const selectedPlan = plans.find((plan) => plan.planId === selectedPlanId) ?? null;
  const selectedCampaign = campaigns.find((campaign) => campaign.campaignId === selectedCampaignId) ?? null;
  const terminalCampaigns = campaigns.filter((campaign) => campaign.state === "succeeded");
  const writeBlocked = config.mode === "offline" || !applicationActive || operationPending;

  const loadPlanOwner = useCallback(async (preferredPlanId = "") => {
    const generation = ++requestGeneration.current;
    setLoadState(config.mode === "offline" ? "offline" : "loading");
    setFailureCode(null);
    setMessage("");
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
    } catch (error) {
      if (generation !== requestGeneration.current) return;
      setPlans([]);
      setSelectedPlanId("");
      setLoadState("failed");
      setMessage(strictFailure(error));
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
    setSelectedVersion(null);
    setCampaigns([]);
    setSelectedCampaignId("");
    setPairEnvelope(null);
    setFailureCode(null);
    setMessage("");
    try {
      const [versionEnvelope, campaignEnvelope] = await Promise.all([
        readApplicationEvaluationPlanVersion(config, planId, plan.latestPlanVersion),
        listApplicationEvaluationCampaigns(config, planId),
      ]);
      if (generation !== requestGeneration.current) return;
      if (versionEnvelope.failureCode || !versionEnvelope.version) {
        setFailureCode(versionEnvelope.failureCode);
        setMessage(versionEnvelope.failureSummary || "The exact plan version is unavailable.");
        return;
      }
      if (campaignEnvelope.failureCode) {
        setFailureCode(campaignEnvelope.failureCode);
        setMessage(campaignEnvelope.failureSummary);
        return;
      }
      setSelectedVersion(versionEnvelope.version);
      setCampaigns(campaignEnvelope.campaigns);
      setSelectedCampaignId(campaignEnvelope.campaigns[0]?.campaignId ?? "");
      selectDefaultPair(campaignEnvelope.campaigns);
    } catch (error) {
      if (generation !== requestGeneration.current) return;
      setMessage(strictFailure(error));
    }
  }, [config, plans]);

  useEffect(() => {
    void loadPlanOwner();
    return () => {
      requestGeneration.current += 1;
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
    await loadExactPlan(planId);
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
    setMessage("");
  }

  function loadVersionIntoEditor() {
    if (!selectedVersion) return;
    setPlanName(selectedVersion.name);
    setProfile(selectedVersion.executionProfile);
    setTargetJSON(serializeApplicationEvaluationTarget(selectedVersion.target));
    setItemsJSON(serializeApplicationEvaluationItems(selectedVersion.items));
    setPendingPlanOperation(null);
    setMessage(`Exact plan ${selectedVersion.planId} v${selectedVersion.planVersion} loaded into the editor. No revision was created.`);
  }

  function reviewPlan(kind: "create" | "revise") {
    try {
      const draft = parseApplicationEvaluationPlanDraft(planName, profile, targetJSON, itemsJSON);
      if (kind === "revise" && (!selectedPlan || !selectedVersion)) {
        setMessage("Select and load an exact active plan before reviewing a revision.");
        return;
      }
      setPendingPlanOperation({ kind, draft });
      setMessage("");
    } catch (error) {
      setPendingPlanOperation(null);
      setMessage(error instanceof Error ? error.message : "Plan review failed.");
    }
  }

  async function confirmPlanOperation() {
    if (!pendingPlanOperation || writeBlocked) return;
    setOperationPending(true);
    setMessage("");
    try {
      if (pendingPlanOperation.kind === "archive") {
        if (!selectedPlan) return;
        const envelope = await archiveApplicationEvaluationPlan(config, selectedPlan.planId, selectedPlan.recordVersion);
        if (envelope.failureCode) {
          applyFailure(envelope.failureCode, envelope.failureSummary);
        } else {
          setPendingPlanOperation(null);
          await loadPlanOwner();
          setMessage(`Plan ${selectedPlan.planId} archived. Existing versions and campaigns remain readable.`);
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
      if (envelope.failureCode || !envelope.plan || !envelope.version) {
        applyFailure(envelope.failureCode, envelope.failureSummary);
        return;
      }
      setPendingPlanOperation(null);
      await loadPlanOwner(envelope.plan.planId);
      setMessage(`${pendingPlanOperation.kind === "create" ? "Created" : "Revised"} ${envelope.plan.planId} v${envelope.version.planVersion}.`);
    } catch (error) {
      setMessage(strictFailure(error));
    } finally {
      setOperationPending(false);
    }
  }

  async function confirmCampaign() {
    if (!campaignConfirmation || !selectedPlan || !selectedVersion || writeBlocked) return;
    setOperationPending(true);
    setMessage("");
    try {
      const envelope = await executeApplicationEvaluationCampaign(config, {
        plan: selectedPlan,
        version: selectedVersion,
        clientCampaignKey,
        quotaAPIKeyId,
      });
      if (envelope.failureCode || !envelope.campaign) {
        applyFailure(envelope.failureCode, envelope.failureSummary);
        return;
      }
      setCampaignConfirmation(false);
      setClientCampaignKey(newCampaignKey(applicationId));
      await loadExactPlan(selectedPlan.planId);
      setSelectedCampaignId(envelope.campaign.campaignId);
      setMessage(`Campaign ${envelope.campaign.campaignId} finished as ${envelope.campaign.state}.`);
    } catch (error) {
      setMessage(strictFailure(error));
    } finally {
      setOperationPending(false);
    }
  }

  async function reconcileCampaign() {
    if (!selectedCampaign || writeBlocked) return;
    setOperationPending(true);
    setMessage("");
    try {
      const envelope = await reconcileApplicationEvaluationCampaign(config, selectedCampaign);
      if (envelope.failureCode || !envelope.campaign) {
        applyFailure(envelope.failureCode, envelope.failureSummary);
        return;
      }
      if (selectedPlan) await loadExactPlan(selectedPlan.planId);
      setMessage(`Campaign ${envelope.campaign.campaignId} reconciled as ${envelope.campaign.state}; no provider execution was replayed.`);
    } catch (error) {
      setMessage(strictFailure(error));
    } finally {
      setOperationPending(false);
    }
  }

  async function previewPair() {
    setOperationPending(true);
    setPairEnvelope(null);
    setHandoffConfirmation(false);
    setMessage("");
    try {
      const envelope = await previewApplicationEvaluationPair(config, baselineCampaignId, candidateCampaignId);
      setPairEnvelope(envelope);
      if (envelope.failureCode) applyFailure(envelope.failureCode, envelope.failureSummary);
    } catch (error) {
      setMessage(strictFailure(error));
    } finally {
      setOperationPending(false);
    }
  }

  async function confirmHandoff() {
    const baseline = campaigns.find((campaign) => campaign.campaignId === baselineCampaignId);
    const candidate = campaigns.find((campaign) => campaign.campaignId === candidateCampaignId);
    if (!handoffConfirmation || !baseline || !candidate || writeBlocked) return;
    setOperationPending(true);
    setMessage("");
    try {
      const envelope = await materializeApplicationEvaluationHandoff(config, baseline, candidate);
      setPairEnvelope(envelope);
      setHandoffConfirmation(false);
      if (envelope.failureCode && envelope.failureCode !== "application_evaluation_handoff_partial") {
        applyFailure(envelope.failureCode, envelope.failureSummary);
        return;
      }
      setMessage(envelope.handoff
        ? `Handoff ${envelope.handoff.state}: ${envelope.handoff.caseRefs.length} exact Case refs${envelope.handoff.suiteId ? ` and Suite ${envelope.handoff.suiteId}` : ""}.`
        : envelope.failureSummary);
    } catch (error) {
      setMessage(strictFailure(error));
    } finally {
      setOperationPending(false);
    }
  }

  function applyFailure(code: ApplicationEvaluationFailureCode | null, summary: string) {
    setFailureCode(code);
    setMessage(summary || (code ? failurePresentation(code).summary : "Application evaluation request failed."));
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
          <p className="eyebrow">S10 · Scheduled regression evaluation</p>
          <h3 id="application-evaluation-title">Scheduled regression campaign</h3>
          <p>Run one exact Prompt plan on a guarded UTC cadence, inspect each occurrence, and hand off only an existing exact Campaign.</p>
        </div>
        <div className="application-evaluation-boundaries">
          <span className="status-badge neutral">{config.environment}</span>
          <span className={`status-badge ${applicationActive ? "good" : "neutral"}`}>
            {applicationActive ? "active application" : "archived · read only"}
          </span>
        </div>
      </header>

      <dl className="application-evaluation-scope">
        <div><dt>Tenant</dt><dd>{config.tenantRef}</dd></div>
        <div><dt>Workspace</dt><dd>{config.workspaceId}</dd></div>
        <div><dt>Environment</dt><dd>{config.environment}</dd></div>
        <div><dt>Application</dt><dd>{applicationId}</dd></div>
      </dl>

      <div className="application-evaluation-stage-shell">
        <aside className="application-evaluation-path">
          <span>Schedule owner</span>
          <div className="application-evaluation-task-switcher" role="tablist" aria-label="Scheduled evaluation path">
            {TASKS.map((task) => (
              <button
                key={task.id}
                type="button"
                role="tab"
                aria-selected={activeTask === task.id || task.id === "campaign" && (activeTask === "pair" || activeTask === "handoff")}
                onClick={() => chooseTask(task.id)}
              >
                <b>{task.number}</b><span><strong>{task.label}</strong><small>{task.summary}</small></span>
              </button>
            ))}
          </div>
          <p><strong>DEV / TEST ONLY</strong><span>No cron, replay, catch-up or production worker.</span></p>
        </aside>

        <div className="application-evaluation-stage-owner">
          {activeTask === "plan" ? <div className="application-evaluation-workbench">
            <PlanList
              plans={plans}
              selectedPlanId={selectedPlanId}
              loadState={loadState}
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
          </div> : null}

          {activeTask === "schedule" || activeTask === "occurrence" ? (
            <Suspense fallback={<div className="application-evaluation-pair-empty"><strong>Loading Schedule owner…</strong><span>The owner remains read-only until its strict consumer is ready.</span></div>}>
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

          {activeTask === "campaign" ? <main className="application-evaluation-owner">
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
          </main> : null}

          {activeTask === "pair" || activeTask === "handoff" ? <main className="application-evaluation-owner">
            <Suspense fallback={<div className="application-evaluation-pair-empty"><strong>Loading Campaign evidence…</strong><span>Exact Run references remain unchanged.</span></div>}>
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

      {failureCode || message ? (
        <div className={`application-evaluation-operation ${failureCode ? "is-failure" : ""}`} role="status">
          <strong>{failureCode ? failurePresentation(failureCode).label : "Owner update"}</strong>
          <span>{message}</span>
          {failureCode === "application_evaluation_version_conflict" || failureCode === "application_evaluation_authority_changed" ? (
            <button type="button" onClick={() => void loadPlanOwner(selectedPlanId)}>Reload exact owner</button>
          ) : null}
        </div>
      ) : null}

      <p className="application-evaluation-boundary-note">
        Production, billing, token/cost limits, queues, parallel fan-out, automatic retry, replay, continuation,
        release, deployment and business writeback remain closed.
      </p>
    </section>
  );
}

function PlanList({
  plans,
  selectedPlanId,
  loadState,
  onSelect,
}: {
  plans: ApplicationEvaluationPlan[];
  selectedPlanId: string;
  loadState: LoadState;
  onSelect: (planId: string) => void;
}) {
  return <aside className="application-evaluation-plan-list" aria-label="Immutable evaluation plans">
    <header><span>Immutable plans</span><strong>{plans.length} active</strong></header>
    <div role="listbox" aria-label="Evaluation plans for the selected application">
      {plans.map((plan) => <button
        key={plan.planId}
        type="button"
        role="option"
        aria-selected={selectedPlanId === plan.planId}
        className={selectedPlanId === plan.planId ? "is-selected" : ""}
        onClick={() => onSelect(plan.planId)}
      >
        <i aria-hidden="true" />
        <span><strong>{plan.name}</strong><small>{plan.executionProfile}</small><small>{plan.itemCount} items · v{plan.latestPlanVersion}</small></span>
        <em>{plan.lifecycleState}</em>
      </button>)}
      {loadState === "loading" ? <p className="application-evaluation-empty">Loading the Plan owner…</p> : null}
      {loadState === "empty" ? <p className="application-evaluation-empty">No active evaluation plans exist for this application.</p> : null}
      {loadState === "offline" ? <p className="application-evaluation-empty">HTTP source disabled. No request was sent.</p> : null}
    </div>
    <p className="application-evaluation-selection-note">Only the plan driving this owner receives the ink-blue selection track. Lifecycle and campaign state stay separate.</p>
  </aside>;
}

function PlanOwner({
  profile,
  planName,
  targetJSON,
  itemsJSON,
  selectedPlan,
  selectedVersion,
  pending,
  writeBlocked,
  operationPending,
  onProfileChange,
  onPlanNameChange,
  onTargetJSONChange,
  onItemsJSONChange,
  onLoadVersion,
  onReview,
  onReviewArchive,
  onCancel,
  onConfirm,
}: {
  profile: ApplicationEvaluationProfile;
  planName: string;
  targetJSON: string;
  itemsJSON: string;
  selectedPlan: ApplicationEvaluationPlan | null;
  selectedVersion: ApplicationEvaluationPlanVersion | null;
  pending: PendingPlanOperation | null;
  writeBlocked: boolean;
  operationPending: boolean;
  onProfileChange: (profile: ApplicationEvaluationProfile) => void;
  onPlanNameChange: (value: string) => void;
  onTargetJSONChange: (value: string) => void;
  onItemsJSONChange: (value: string) => void;
  onLoadVersion: () => void;
  onReview: (kind: "create" | "revise") => void;
  onReviewArchive: () => void;
  onCancel: () => void;
  onConfirm: () => void;
}) {
  return (
    <section className="application-evaluation-plan-owner" id="application-evaluation-plan">
      <header>
        <div>
          <p className="eyebrow">Current owner · Plan</p>
          <h4>Immutable evaluation plan</h4>
          <p>Creating or revising freezes one profile, ordered fixtures, expected classifications and a canonical digest.</p>
        </div>
        <span className="status-badge neutral">{selectedVersion ? `v${selectedVersion.planVersion} loaded` : "new draft"}</span>
      </header>
      <div className="application-evaluation-plan-meta">
        <div><span>Selected plan</span><strong>{selectedPlan?.planId ?? "none"}</strong></div>
        <div><span>Record version</span><strong>{selectedPlan?.recordVersion ?? "—"}</strong></div>
        <div><span>Latest immutable</span><strong>{selectedPlan ? `v${selectedPlan.latestPlanVersion}` : "—"}</strong></div>
        <div><span>Digest</span><strong>{shortRef(selectedPlan?.latestPlanDigest ?? "")}</strong></div>
      </div>
      <div className="application-evaluation-plan-editor">
        <label>
          <span>Plan name</span>
          <input value={planName} onChange={(event) => onPlanNameChange(event.target.value)} disabled={operationPending} />
        </label>
        <label>
          <span>Execution profile</span>
          <select value={profile} onChange={(event) => onProfileChange(event.target.value as ApplicationEvaluationProfile)} disabled={operationPending}>
            {APPLICATION_EVALUATION_PROFILES.map((value) => <option key={value} value={value}>{value}</option>)}
          </select>
        </label>
        <label className="is-code">
          <span>Profile target · strict JSON</span>
          <textarea value={targetJSON} onChange={(event) => onTargetJSONChange(event.target.value)} spellCheck={false} disabled={operationPending} />
        </label>
        <label className="is-code is-items">
          <span>{profile === "workflow_definition_executor_v2"
            ? "Ordered fixtures · 1–20 · advanced strict JSON"
            : "Ordered fixtures · 1–20 · strict JSON"}</span>
          <textarea value={itemsJSON} onChange={(event) => onItemsJSONChange(event.target.value)} spellCheck={false} disabled={operationPending} />
        </label>
        {profile === "workflow_definition_executor_v2" ? <StructuredEvaluationFixtureEditor
          targetJSON={targetJSON}
          itemsJSON={itemsJSON}
          disabled={operationPending}
          onItemsJSONChange={onItemsJSONChange}
        /> : null}
      </div>
      {!pending ? (
        <div className="application-evaluation-actions">
          <button type="button" className="secondary-action" onClick={onLoadVersion} disabled={!selectedVersion || operationPending}>Load exact version</button>
          <button type="button" className="secondary-action" onClick={() => onReview("create")} disabled={writeBlocked}>Review create</button>
          <button type="button" onClick={() => onReview("revise")} disabled={writeBlocked || !selectedPlan}>Review revision</button>
          <button type="button" className="attention-action" onClick={onReviewArchive} disabled={writeBlocked || !selectedPlan}>Review archive</button>
        </div>
      ) : (
        <div className="application-evaluation-confirmation">
          <span>CONFIRM {pending.kind.toUpperCase()}</span>
          <strong>{pending.kind === "archive" ? selectedPlan?.name : pending.draft?.name}</strong>
          <p>{pending.kind === "archive"
            ? "No new campaigns may be started. Immutable versions and existing campaigns remain readable."
            : `${pending.draft?.items.length ?? 0} ordered items · ${pending.draft?.executionProfile}`}</p>
          <p>No mutation occurs until this confirmation. Revision uses expected record version {selectedPlan?.recordVersion ?? 0}.</p>
          <div><button type="button" className="secondary-action" onClick={onCancel}>Cancel</button><button type="button" onClick={onConfirm} disabled={writeBlocked}>Confirm {pending.kind}</button></div>
        </div>
      )}
    </section>
  );
}

function StructuredEvaluationFixtureEditor({
  targetJSON,
  itemsJSON,
  disabled,
  onItemsJSONChange,
}: {
  targetJSON: string;
  itemsJSON: string;
  disabled: boolean;
  onItemsJSONChange: (value: string) => void;
}) {
  const source = useMemo(() => {
    try {
      const draft = parseApplicationEvaluationPlanDraft(
        "Structured evaluation fixture editor",
        "workflow_definition_executor_v2",
        targetJSON,
        itemsJSON,
      );
      const contract = draft.target.workflowDefinition?.inputContract ?? null;
      return contract ? { contract, items: draft.items } : null;
    } catch {
      return null;
    }
  }, [itemsJSON, targetJSON]);
  const [selectedItemKey, setSelectedItemKey] = useState("");
  const selectedItem = source?.items.find((item) => item.itemKey === selectedItemKey) ?? source?.items[0] ?? null;
  const sourceInputs = selectedItem?.workflowDefinition?.inputs ?? {};
  const sourceKey = source && selectedItem
    ? `${structuredRuntimeInputAuthorityKey(source.contract)}:${selectedItem.itemKey}:${JSON.stringify(sourceInputs)}`
    : "invalid";
  const [drafts, setDrafts] = useState<StructuredRuntimeInputDrafts>(() => runtimeInputDrafts(sourceInputs));
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [validationSummary, setValidationSummary] = useState("");

  useEffect(() => {
    setSelectedItemKey(selectedItem?.itemKey ?? "");
    setDrafts(runtimeInputDrafts(sourceInputs));
    setFieldErrors({});
    setValidationSummary("");
  }, [sourceKey]);

  if (!source || !selectedItem) {
    return <div className="application-evaluation-structured-fixture invalid">
      <strong>Structured fixture editor unavailable</strong>
      <p>Load an exact v2 Definition target contract and valid ordered fixtures to enable typed editing.</p>
    </div>;
  }
  const validSource = source;
  const currentItem = selectedItem;

  function updateDrafts(next: StructuredRuntimeInputDrafts) {
    setDrafts(next);
    const validation = validateStructuredRuntimeInputDrafts(validSource.contract, next);
    setFieldErrors(validation.fieldErrors);
    setValidationSummary(validation.summary);
    if (!validation.ok) return;
    const nextItems = validSource.items.map((item) => item.itemKey === currentItem.itemKey ? {
      ...item,
      workflowDefinition: item.workflowDefinition ? { ...item.workflowDefinition, inputs: validation.inputs } : null,
    } : item);
    onItemsJSONChange(serializeApplicationEvaluationItems(nextItems));
  }

  return <div className="application-evaluation-structured-fixture">
    <div className="application-evaluation-structured-fixture-toolbar">
      <div>
        <strong>Typed fixture input</strong>
        <p>Values are validated against the immutable Definition contract before they update the fixture JSON.</p>
      </div>
      <label>
        <span>Fixture</span>
        <select value={selectedItem.itemKey} onChange={(event) => setSelectedItemKey(event.target.value)} disabled={disabled}>
          {source.items.map((item) => <option key={item.itemKey} value={item.itemKey}>{item.name} · {item.itemKey}</option>)}
        </select>
      </label>
    </div>
    <StructuredRuntimeInputEditor
      contract={source.contract}
      drafts={drafts}
      fieldErrors={fieldErrors}
      disabled={disabled}
      onChange={updateDrafts}
    />
    {validationSummary ? <p className="application-evaluation-structured-fixture-error" role="alert">{validationSummary}</p> : null}
  </div>;
}

function runtimeInputDrafts(values: StructuredRuntimeInputValues): StructuredRuntimeInputDrafts {
  return Object.fromEntries(Object.entries(values).map(([name, value]) => [name, typeof value === "boolean" ? value : String(value)]));
}

function CampaignOwner({
  plan,
  version,
  campaigns,
  selectedCampaign,
  quotaAPIKeyId,
  clientCampaignKey,
  confirmed,
  writeBlocked,
  operationPending,
  onSelectCampaign,
  onNewCampaign,
  onQuotaAPIKeyIdChange,
  onClientCampaignKeyChange,
  onReview,
  onCancel,
  onConfirm,
  onReconcile,
  onOpenHandoff,
}: {
  plan: ApplicationEvaluationPlan | null;
  version: ApplicationEvaluationPlanVersion | null;
  campaigns: ApplicationEvaluationCampaign[];
  selectedCampaign: ApplicationEvaluationCampaign | null;
  quotaAPIKeyId: string;
  clientCampaignKey: string;
  confirmed: boolean;
  writeBlocked: boolean;
  operationPending: boolean;
  onSelectCampaign: (id: string) => void;
  onNewCampaign: () => void;
  onQuotaAPIKeyIdChange: (value: string) => void;
  onClientCampaignKeyChange: (value: string) => void;
  onReview: () => void;
  onCancel: () => void;
  onConfirm: () => void;
  onReconcile: () => void;
  onOpenHandoff: () => void;
}) {
  const total = selectedCampaign?.items.length ?? version?.items.length ?? 0;
  const completed = (selectedCampaign?.succeededItems ?? 0) + (selectedCampaign?.failedItems ?? 0);
  const terminalCampaigns = campaigns.filter((campaign) => campaign.state === "succeeded");
  return (
    <section className="application-evaluation-campaign-owner">
      <div className="application-evaluation-campaign-context">
        <div>
          <span>Selected campaign</span>
          <label>
            <select
              value={selectedCampaign?.campaignId ?? ""}
              onChange={(event) => onSelectCampaign(event.target.value)}
              aria-label="Selected evaluation campaign"
            >
              <option value="">New campaign draft</option>
              {campaigns.map((campaign) => (
                <option key={campaign.campaignId} value={campaign.campaignId}>
                  {campaign.clientCampaignKey} · v{campaign.recordVersion}
                </option>
              ))}
            </select>
          </label>
        </div>
        <div className="application-evaluation-campaign-context-meta">
          <span>{campaigns.length} campaigns</span>
          <small>{selectedCampaign ? shortRef(selectedCampaign.campaignId) : "No execution selected"}</small>
          <div>
            <span className={`status-badge ${selectedCampaign?.state === "succeeded" ? "good" : "neutral"}`}>
              {selectedCampaign?.state ?? "not started"}
            </span>
            <button type="button" className="secondary-action" aria-label="New campaign" onClick={onNewCampaign} disabled={writeBlocked || !plan || !version}>
              New
            </button>
          </div>
        </div>
      </div>
      <div className="application-evaluation-campaign-surface">
        <div className="application-evaluation-campaign-main">
          <header>
            <div>
              <p className="eyebrow">Current owner · Campaign</p>
              <h4>{selectedCampaign?.clientCampaignKey ?? "Sequential controlled execution"}</h4>
              <p>Each item delegates once to its existing service and checkpoints a deterministic durable Run reference.</p>
            </div>
            <span className={`status-badge ${selectedCampaign?.state === "succeeded" ? "good" : "neutral"}`}>
              {selectedCampaign?.state ?? "draft"}
            </span>
          </header>
          <dl className="application-evaluation-authority">
            <div><dt>Profile</dt><dd>{version?.executionProfile ?? "—"}</dd></div>
            <div><dt>Plan</dt><dd>{version ? `v${version.planVersion} · ${shortRef(version.planDigest)}` : "—"}</dd></div>
            <div><dt>Quota consumer</dt><dd>{selectedCampaign?.quotaAPIKeyId || "required"}</dd></div>
            <div><dt>Authority</dt><dd>{shortRef(selectedCampaign?.authority?.authorityDigest ?? "")}</dd></div>
          </dl>
          {selectedCampaign ? (
            <>
              <div className="application-evaluation-progress">
                <div><span>Sequential progress</span><strong>{completed} / {total}</strong><i style={{ width: `${total ? Math.min(100, completed / total * 100) : 0}%` }} /></div>
                <dl><div><dt>Succeeded</dt><dd>{selectedCampaign.succeededItems}</dd></div><div><dt>Failed</dt><dd>{selectedCampaign.failedItems}</dd></div><div><dt>Retry / replay</dt><dd>0</dd></div></dl>
              </div>
              <CampaignItems campaign={selectedCampaign} />
              {selectedCampaign.state === "interrupted" || selectedCampaign.state === "running" ? (
                <button type="button" className="secondary-action" onClick={onReconcile} disabled={writeBlocked}>Reconcile exact Run · no replay</button>
              ) : null}
            </>
          ) : (
            <div className="application-evaluation-campaign-create">
              <label><span>Active API Key ID · quota consumer only</span><input value={quotaAPIKeyId} onChange={(event) => onQuotaAPIKeyIdChange(event.target.value)} placeholder="key_…" disabled={operationPending} /></label>
              <label><span>Client campaign key · idempotent in this scope</span><input value={clientCampaignKey} onChange={(event) => onClientCampaignKeyChange(event.target.value)} disabled={operationPending} /></label>
              {!confirmed ? <button type="button" onClick={onReview} disabled={writeBlocked || !plan || !version || !quotaAPIKeyId.trim()}>Review execution</button> : (
                <div className="application-evaluation-confirmation">
                  <span>CONFIRM SEQUENTIAL EXECUTION</span>
                  <strong>{version?.name}</strong>
                  <p>{version?.items.length ?? 0} items will run in order under exact plan v{version?.planVersion}, current authority and API Key {quotaAPIKeyId}.</p>
                  <p>Quota is consumed before each provider attempt. Failure stops remaining items; there is no automatic retry, fallback, resume or replay.</p>
                  <div><button type="button" className="secondary-action" onClick={onCancel}>Cancel</button><button type="button" onClick={onConfirm} disabled={writeBlocked}>Confirm campaign</button></div>
                </div>
              )}
            </div>
          )}
        </div>
        <aside className="application-evaluation-campaign-handoff-rail" aria-label="Evaluation evidence handoff readiness">
          <header>
            <span>Evidence handoff</span>
            <strong>{terminalCampaigns.length >= 2 ? "Pair ready" : "Waiting"}</strong>
          </header>
          <strong>{terminalCampaigns.length} / 2</strong>
          <span>succeeded campaigns</span>
          <dl>
            <div><dt>Selected</dt><dd>{selectedCampaign?.state === "succeeded" ? "checkpointed" : "not terminal"}</dd></div>
            <div><dt>Pair review</dt><dd>{terminalCampaigns.length >= 2 ? "available" : "blocked"}</dd></div>
            <div><dt>Case / Suite</dt><dd>explicit only</dd></div>
          </dl>
          <p>Exact Run refs remain in their existing owners. Handoff never approves, releases or deploys a candidate.</p>
          <button type="button" onClick={onOpenHandoff} disabled={terminalCampaigns.length < 2}>
            Review handoff
          </button>
        </aside>
      </div>
    </section>
  );
}

function CampaignItems({ campaign }: { campaign: ApplicationEvaluationCampaign }) {
  return (
    <div className="application-evaluation-items">
      <div className="application-evaluation-item-header"><span>Item</span><span>Durable Run ref</span><span>State</span><span>Boundary</span></div>
      {campaign.items.map((item) => (
        <div key={item.itemKey} className="application-evaluation-item">
          <strong>{item.itemKey}</strong><code>{item.runId || "deterministic · pending"}</code><em className={item.state}>{item.state}</em><small>{item.failureBoundary || (item.state === "succeeded" ? "durable terminal" : "zero side effects")}</small>
        </div>
      ))}
    </div>
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

function shortRef(value: string): string {
  if (!value) return "—";
  return value.length <= 22 ? value : `${value.slice(0, 11)}…${value.slice(-8)}`;
}

function strictFailure(error: unknown): string {
  return error instanceof Error
    ? `${error.message} No response data was accepted and the displayed owner remains unchanged.`
    : "Application evaluation failed strict validation. No response data was accepted.";
}

function failurePresentation(code: ApplicationEvaluationFailureCode): { label: string; summary: string } {
  const values: Record<ApplicationEvaluationFailureCode, { label: string; summary: string }> = {
    application_evaluation_scope_denied: { label: "Permission or scope denied", summary: "The active identity, workspace, application or permission set does not match this owner." },
    application_evaluation_environment_denied: { label: "Environment blocked", summary: "Only the configured development or test environment is accepted." },
    application_evaluation_not_found: { label: "Record missing", summary: "The exact Plan, Campaign or referenced evidence is unavailable in the current scope." },
    application_evaluation_payload_invalid: { label: "Input rejected", summary: "The strict Plan, Campaign or handoff input contract is invalid." },
    application_evaluation_secret_material_forbidden: { label: "Secret material rejected", summary: "Evaluation fixtures cannot contain credentials, tokens, private keys or secret transport material." },
    application_evaluation_profile_ineligible: { label: "Profile ineligible", summary: "The application kind or current runtime authority cannot execute this evaluation profile." },
    application_evaluation_version_conflict: { label: "Version conflict", summary: "The owner changed. Reload its exact record version before reviewing another mutation." },
    application_evaluation_archived: { label: "Archived plan", summary: "Archived Plans remain readable but cannot be revised or used for new Campaigns." },
    application_evaluation_cursor_invalid: { label: "Cursor invalid", summary: "The page cursor does not match the current scope or filter." },
    application_evaluation_campaign_conflict: { label: "Campaign key conflict", summary: "The idempotency key is already bound to a different immutable plan version." },
    application_evaluation_authority_changed: { label: "Authority drift", summary: "Runtime authority changed; remaining items stopped without fallback or automatic continuation." },
    application_evaluation_run_unavailable: { label: "Durable Run unavailable", summary: "The expected terminal Run could not be produced or reconciled. Provider execution is not replayed." },
    application_evaluation_quota_consumer_invalid: { label: "API Key / quota blocked", summary: "The selected API Key is not an active actor-owned quota consumer for this application." },
    application_evaluation_handoff_partial: { label: "Partial handoff", summary: "Completed exact Case refs remain durable; inspect the candidate Campaign before a new explicit action." },
    application_evaluation_store_unavailable: { label: "Store unavailable", summary: "The evaluation repository failed closed. No fallback owner was used." },
    application_evaluation_store_contract_mismatch: { label: "Store contract mismatch", summary: "Persisted evaluation data failed its strict contract and was not rendered." },
    application_evaluation_write_disabled: { label: "Development gate disabled", summary: "Explicit application evaluation HTTP opt-in is required; no request was sent in offline mode." },
  };
  return values[code];
}

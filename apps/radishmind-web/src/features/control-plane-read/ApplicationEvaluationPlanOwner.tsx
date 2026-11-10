import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import "../../i18n/evaluationPlanResources.ts";
import { campaignOperation, campaignState } from "./applicationEvaluationMessages.ts";
import { APPLICATION_EVALUATION_PROFILES, parseApplicationEvaluationPlanDraft, parseStructuredEvaluationFixtureSource, serializeApplicationEvaluationItems, type ApplicationEvaluationConfig, type ApplicationEvaluationDefinitionTarget, type ApplicationEvaluationPlan, type ApplicationEvaluationPlanDraft, type ApplicationEvaluationPlanVersion, type ApplicationEvaluationProfile } from "./applicationEvaluationCampaignConsumer.ts";
import WorkflowEvaluationDefinitionReference from "./WorkflowEvaluationDefinitionReference.tsx";
import StructuredRuntimeInputEditor from "./StructuredRuntimeInputEditor.tsx";
import { validateStructuredRuntimeInputDrafts, type StructuredRuntimeInputDrafts, type StructuredRuntimeInputValues, validateStructuredRuntimeInputValues } from "./structuredRuntimeInput.ts";
import type { PendingPlanOperation, LoadState } from "./applicationEvaluationCampaignPanel.tsx";
export function PlanList({
  plans,
  selectedPlanId,
  loadState,
  disabled,
  onSelect,
}: {
  plans: ApplicationEvaluationPlan[];
  selectedPlanId: string;
  loadState: LoadState;
  disabled: boolean;
  onSelect: (planId: string) => void;
}) {
  const { t } = useTranslation("evaluation");
  return <aside className="application-evaluation-plan-list" aria-label={t($ => $.campaign.immutablePlansLabel)}>
    <header><span>{t($ => $.campaign.immutablePlans)}</span><strong>{t($ => $.campaign.activeCount, { count: plans.length })}</strong></header>
    <div role="listbox" aria-label={t($ => $.campaign.plansLabel)}>
      {plans.map((plan) => <button
        key={plan.planId}
        type="button"
        role="option"
        disabled={disabled}
        aria-selected={selectedPlanId === plan.planId}
        className={selectedPlanId === plan.planId ? "is-selected" : ""}
        onClick={() => onSelect(plan.planId)}
      >
        <i aria-hidden="true" />
        <span><strong>{plan.name}</strong><small>{plan.executionProfile}</small><small>{t($ => $.campaign.itemVersion, { count: plan.itemCount, version: plan.latestPlanVersion })}</small></span>
        <em>{campaignState(t, plan.lifecycleState)}</em>
      </button>)}
      {loadState === "loading" ? <p className="application-evaluation-empty">{t($ => $.campaign.planLoading)}</p> : null}
      {loadState === "empty" ? <p className="application-evaluation-empty">{t($ => $.campaign.plansEmpty)}</p> : null}
      {loadState === "offline" ? <p className="application-evaluation-empty">{t($ => $.campaign.offline)}</p> : null}
    </div>
    <p className="application-evaluation-selection-note">{t($ => $.campaign.selectionNote)}</p>
  </aside>;
}

export function PlanOwner({
  config,
  onAdoptTarget,
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
  config: ApplicationEvaluationConfig;
  onAdoptTarget: (target: ApplicationEvaluationDefinitionTarget) => void;
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
  const { t } = useTranslation("evaluation");
  const validDraft = useMemo(() => {
    try { return parseApplicationEvaluationPlanDraft(planName, profile, targetJSON, itemsJSON); }
    catch { return null; }
  }, [planName, profile, targetJSON, itemsJSON]);
  return (
    <section className="application-evaluation-plan-owner" id="application-evaluation-plan">
      <header>
        <div>
          <p className="eyebrow">{t($ => $.campaign.planOwner)}</p>
          <h4>{t($ => $.campaign.planTitle)}</h4>
          <p>{t($ => $.campaign.planIntro)}</p>
        </div>
        <span className="status-badge neutral">{selectedVersion ? t($ => $.campaign.versionLoaded, { version: selectedVersion.planVersion }) : t($ => $.campaign.newDraft)}</span>
      </header>
      <div className="application-evaluation-plan-meta">
        <div><span>{t($ => $.campaign.selectedPlan)}</span><strong>{selectedPlan?.planId ?? t($ => $.campaign.none)}</strong></div>
        <div><span>{t($ => $.campaign.recordVersion)}</span><strong>{selectedPlan?.recordVersion ?? "—"}</strong></div>
        <div><span>{t($ => $.campaign.latestImmutable)}</span><strong>{selectedPlan ? `v${selectedPlan.latestPlanVersion}` : "—"}</strong></div>
        <div><span>{t($ => $.campaign.digest)}</span><strong>{shortRef(selectedPlan?.latestPlanDigest ?? "")}</strong></div>
      </div>
      {profile.startsWith("workflow_definition_executor_") ? <>
        <WorkflowEvaluationDefinitionReference config={config} disabled={operationPending} onAdopt={onAdoptTarget} />
        <p className="application-evaluation-boundary-note">{t($ => $.campaign.workflowRegressionBoundary)}</p>
      </> : null}
      <div className="application-evaluation-plan-editor">
        <label>
          <span>{t($ => $.campaign.planName)}</span>
          <input value={planName} onChange={(event) => onPlanNameChange(event.target.value)} disabled={operationPending} />
        </label>
        <label>
          <span>{t($ => $.campaign.executionProfile)}</span>
          <select value={profile} onChange={(event) => onProfileChange(event.target.value as ApplicationEvaluationProfile)} disabled={operationPending}>
            {APPLICATION_EVALUATION_PROFILES.map((value) => <option key={value} value={value}>{value}</option>)}
          </select>
        </label>
        <label className="is-code">
          <span>{t($ => $.campaign.targetJSON)}</span>
          <textarea value={targetJSON} onChange={(event) => onTargetJSONChange(event.target.value)} spellCheck={false} disabled={operationPending} />
        </label>
        <label className="is-code is-items">
          <span>{profile === "workflow_definition_executor_v2"
            ? t($ => $.campaign.advancedFixtures)
            : t($ => $.campaign.fixturesJSON)}</span>
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
          <button type="button" className="secondary-action" onClick={onLoadVersion} disabled={!selectedVersion || operationPending}>{t($ => $.campaign.loadVersion)}</button>
          <button type="button" className="secondary-action" onClick={() => onReview("create")} disabled={writeBlocked || !validDraft}>{t($ => $.campaign.reviewCreate)}</button>
          <button type="button" onClick={() => onReview("revise")} disabled={writeBlocked || !selectedPlan || !validDraft}>{t($ => $.campaign.reviewRevision)}</button>
          <button type="button" className="attention-action" onClick={onReviewArchive} disabled={writeBlocked || !selectedPlan}>{t($ => $.campaign.reviewArchive)}</button>
        </div>
      ) : (
        <div className="application-evaluation-confirmation">
          <span>{t($ => $.campaign.confirmOperation, { operation: campaignOperation(t, pending.kind) })}</span>
          <strong>{pending.kind === "archive" ? selectedPlan?.name : pending.draft?.name}</strong>
          <p>{pending.kind === "archive"
            ? t($ => $.campaign.archiveBoundary)
            : t($ => $.campaign.orderedItems, { count: pending.draft?.items.length ?? 0, profile: pending.draft?.executionProfile ?? "" })}</p>
          {pending.draft ? <pre className="application-evaluation-confirmation-values">{JSON.stringify({ target: pending.draft.target, items: pending.draft.items }, null, 2)}</pre> : null}
          <p>{t($ => $.campaign.planConfirmation, { version: selectedPlan?.recordVersion ?? 0 })}</p>
          <div><button type="button" className="secondary-action" onClick={onCancel} disabled={operationPending}>{t($ => $.campaign.cancel)}</button><button type="button" onClick={onConfirm} disabled={writeBlocked}>{t($ => $.campaign.confirmOperation, { operation: campaignOperation(t, pending.kind) })}</button></div>
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
  const { t } = useTranslation("evaluation");
  const source = useMemo(() => {
    try {
      return parseStructuredEvaluationFixtureSource(targetJSON, itemsJSON);
    } catch {
      return null;
    }
  }, [itemsJSON, targetJSON]);
  const [selectedItemIndex, setSelectedItemIndex] = useState(0);
  const selectedIndex = source && selectedItemIndex < source.items.length ? selectedItemIndex : 0;
  const selectedItem = source?.items[selectedIndex] ?? null;
  const sourceInputs = selectedItem?.workflowDefinition?.inputs ?? {};
  const drafts = runtimeInputDrafts(sourceInputs);
  const validation = source ? validateStructuredRuntimeInputValues(source.contract, sourceInputs) : null;
  if (!source || !selectedItem) {
    return <div className="application-evaluation-structured-fixture invalid">
      <strong>{t($ => $.campaign.fixtureUnavailable)}</strong>
      <p>{t($ => $.campaign.fixtureRequirement)}</p>
    </div>;
  }
  const validSource = source;
  const currentItem = selectedItem;

  function updateDrafts(next: StructuredRuntimeInputDrafts) {
    const inputs: StructuredRuntimeInputValues = {};
    for (const [name, value] of Object.entries(next)) {
      if (value === undefined) continue;
      const field = validSource.contract.fields.find(candidate => candidate.name === name);
      const checked = field ? validateStructuredRuntimeInputDrafts({ ...validSource.contract, fields: [field] }, { [name]: value }) : null;
      inputs[name] = checked?.ok ? checked.inputs[name] : value;
    }
    const nextItems = validSource.items.map((item) => item === currentItem ? {
      ...item,
      workflowDefinition: item.workflowDefinition ? { ...item.workflowDefinition, inputs } : null,
    } : item);
    onItemsJSONChange(serializeApplicationEvaluationItems(nextItems));
  }

  return <div className="application-evaluation-structured-fixture">
    <div className="application-evaluation-structured-fixture-toolbar">
      <div>
        <strong>{t($ => $.campaign.typedInput)}</strong>
        <p>{t($ => $.campaign.fixtureValidation)}</p>
      </div>
      <label>
        <span>{t($ => $.campaign.fixture)}</span>
        <select value={selectedIndex} onChange={(event) => setSelectedItemIndex(Number(event.target.value))} disabled={disabled}>
          {source.items.map((item, index) => <option key={index} value={index}>{item.name} · {item.itemKey}</option>)}
        </select>
      </label>
    </div>
    <StructuredRuntimeInputEditor
      contract={source.contract}
      drafts={drafts}
      fieldErrors={validation?.fieldErrors ?? {}}
      boundary={t($ => $.campaign.fixturePersistence)}
      disabled={disabled}
      onChange={updateDrafts}
    />
    {validation && !validation.ok ? <p className="application-evaluation-structured-fixture-error" role="alert">{t($ => $.campaign.fixtureInvalid)}</p> : null}
  </div>;
}

function runtimeInputDrafts(values: StructuredRuntimeInputValues): StructuredRuntimeInputDrafts {
  return Object.fromEntries(Object.entries(values).map(([name, value]) => [name, typeof value === "boolean" ? value : String(value)]));
}


function shortRef(value: string): string {
  return value.length <= 22 ? value || "—" : `${value.slice(0, 11)}…${value.slice(-8)}`;
}

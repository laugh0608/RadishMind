import { useTranslation } from "react-i18next";
import { useLocalePreference } from "../../i18n/LocaleProvider.tsx";
import { formatDisplayNumber } from "../../i18n/formatters.ts";
import { adminDisplayDate } from "./adminManagementFormatters.ts";
import { providerOperationMessage, providerFindingMessage, type ProviderMessage } from "./adminProviderRouteMessages.ts";
import "../../i18n/adminProviderResources.ts";
import { useEffect, useMemo, useState } from "react";

import {
  activateAdminProviderRouteCandidate,
  buildAdminProviderRouteCandidateDiff,
  createAdminProviderRouteCandidate,
  createAdminProviderRouteDraftInput,
  draftInputFromAdminProviderRouteDraft,
  listAdminProviderRouteActivations,
  readAdminProviderRouteCandidate,
  readAdminProviderRouteConfig,
  readAdminProviderRouteDraft,
  readAdminProviderRouteSnapshot,
  reviewAdminProviderRouteCandidate,
  saveAdminProviderRouteDraft,
  validateAdminProviderRouteDraft,
  type AdminModelRouteDefinition,
  type AdminProviderProfileAssignment,
  type AdminProviderRouteActivation,
  type AdminProviderRouteActivationAction,
  type AdminProviderRouteCandidate,
  type AdminProviderRouteDecision,
  type AdminProviderRouteDraftInput,
  type AdminProviderRouteEnvelope,
  type AdminProviderRouteExecutionMode,
  type AdminProviderRouteProtocol,
  type AdminProviderRouteSnapshot,
} from "./adminProviderRouteConsumer.ts";

const defaultConfig = readAdminProviderRouteConfig();
const PROTOCOLS: AdminProviderRouteProtocol[] = ["chat_completions", "responses", "messages"];

type WorkspaceOperation = {
  status: "offline" | "idle" | "loading" | "ready" | "failed";
  message: ProviderMessage;
  failureCode: string;
  requestId: string;
  auditRef: string;
};

export function AdminProviderRouteWorkspacePanel({
  focus = "route",
  applicationId,
}: {
  focus?: "provider" | "profile" | "route";
  applicationId?: string;
}) {
  const { t } = useTranslation("admin");
  const { locale } = useLocalePreference();
  const config = useMemo(() => ({
    ...defaultConfig,
    applicationId: applicationId?.trim() || defaultConfig.applicationId,
  }), [applicationId]);
  const [draftInput, setDraftInput] = useState<AdminProviderRouteDraftInput>(() =>
    createAdminProviderRouteDraftInput(config),
  );
  const [candidateId, setCandidateId] = useState("candidate-one");
  const [candidate, setCandidate] = useState<AdminProviderRouteCandidate | null>(null);
  const [snapshot, setSnapshot] = useState<AdminProviderRouteSnapshot | null>(null);
  const [history, setHistory] = useState<AdminProviderRouteActivation[]>([]);
  const [reviewDecision, setReviewDecision] = useState<AdminProviderRouteDecision>("approve");
  const [reviewReason, setReviewReason] = useState("Reviewed runtime inventory, capabilities, and exact model routes.");
  const [activationAction, setActivationAction] = useState<AdminProviderRouteActivationAction>("activate");
  const [activationReason, setActivationReason] = useState("Enable the independently reviewed route configuration.");
  const [operation, setOperation] = useState<WorkspaceOperation>(() => initialOperation(config));
  const findings = useMemo(() => validateAdminProviderRouteDraft(config, draftInput), [config, draftInput]);
  const candidateDiff = useMemo(
    () => candidate ? buildAdminProviderRouteCandidateDiff(candidate, snapshot) : null,
    [candidate, snapshot],
  );

  useEffect(() => {
    if (config.mode === "dev_admin_provider_route_http") void refreshWorkspace();
  }, [config]);

  async function refreshWorkspace() {
    setOperation(loadingOperation({ key: "loadingWorkspace" }));
    try {
      const [draftResult, snapshotResult, historyResult] = await Promise.all([
        readAdminProviderRouteDraft(config),
        readAdminProviderRouteSnapshot(config),
        listAdminProviderRouteActivations(config),
      ]);
      if (draftResult.draft) setDraftInput(draftInputFromAdminProviderRouteDraft(draftResult.draft));
      if (snapshotResult.snapshot) {
        setSnapshot(snapshotResult.snapshot);
        setCandidateId((current) => current || snapshotResult.snapshot!.candidateId);
      } else if (snapshotResult.failureCode === "admin_provider_route_draft_not_found") {
        setSnapshot(null);
      }
      setHistory(historyResult.activationHistory);
      const fatal = firstUnexpectedFailure([
        draftResult,
        snapshotResult,
        historyResult,
      ], ["admin_provider_route_draft_not_found"]);
      setOperation(fatal ? failedOperation(fatal) : readyOperation(
        { key: "workspaceLoaded" },
        historyResult,
      ));
    } catch {
      setOperation(networkFailure());
    }
  }

  async function saveDraft() {
    if (findings.length) {
      setOperation({
        status: "failed",
        message: { key: "localFindings" },
        failureCode: "admin_provider_route_payload_invalid",
        requestId: "",
        auditRef: "",
      });
      return;
    }
    setOperation(loadingOperation({ key: "savingDraft" }));
    try {
      const result = await saveAdminProviderRouteDraft(config, draftInput);
      if (result.draft) setDraftInput(draftInputFromAdminProviderRouteDraft(result.draft));
      setOperation(operationFromEnvelope(result, { key: "draftSaved" }));
    } catch {
      setOperation(networkFailure());
    }
  }

  async function createCandidate() {
    setOperation(loadingOperation({ key: "creatingCandidate" }));
    try {
      const result = await createAdminProviderRouteCandidate(
        config,
        candidateId,
        draftInput.expectedRevision,
      );
      if (result.candidate) setCandidate(result.candidate);
      setOperation(operationFromEnvelope(result, { key: "candidateCreated" }));
    } catch {
      setOperation(networkFailure());
    }
  }

  async function loadCandidate(id = candidateId) {
    const normalizedId = id.trim();
    if (!normalizedId) return;
    setCandidateId(normalizedId);
    setOperation(loadingOperation({ key: "loadingCandidate", candidateId: normalizedId }));
    try {
      const result = await readAdminProviderRouteCandidate(config, normalizedId);
      if (result.candidate) setCandidate(result.candidate);
      setOperation(operationFromEnvelope(result, { key: "candidateLoaded", candidateId: normalizedId }));
    } catch {
      setOperation(networkFailure());
    }
  }

  async function reviewCandidate() {
    if (!candidate || reviewReason.trim().length < 4) return;
    setOperation(loadingOperation({ key: "recordingReview", decision: reviewDecision }));
    try {
      const result = await reviewAdminProviderRouteCandidate(
        config,
        candidate.candidateId,
        candidate.reviewVersion,
        reviewDecision,
        reviewReason,
      );
      if (result.candidate) setCandidate(result.candidate);
      setOperation(operationFromEnvelope(result, { key: "reviewRecorded" }));
    } catch {
      setOperation(networkFailure());
    }
  }

  async function activateCandidate() {
    if (!candidate || activationReason.trim().length < 4) return;
    const generation = snapshot?.generation ?? 0;
    setOperation(loadingOperation(
      { key: activationAction === "rollback" ? "rollingBack" : "activating", generation },
    ));
    try {
      const result = await activateAdminProviderRouteCandidate(
        config,
        candidate.candidateId,
        generation,
        activationAction,
        activationReason,
      );
      if (result.snapshot) setSnapshot(result.snapshot);
      const historyResult = result.failureCode ? null : await listAdminProviderRouteActivations(config);
      if (historyResult) setHistory(historyResult.activationHistory);
      setOperation(operationFromEnvelope(
        result,
        { key: activationAction === "rollback" ? "rolledBack" : "activated" },
      ));
    } catch {
      setOperation(networkFailure());
    }
  }

  function updateProfile(index: number, patch: Partial<AdminProviderProfileAssignment>) {
    setDraftInput((current) => ({
      ...current,
      providerProfiles: current.providerProfiles.map((profile, profileIndex) =>
        profileIndex === index ? { ...profile, ...patch } : profile),
    }));
  }

  function updateRoute(index: number, nextRoute: AdminModelRouteDefinition) {
    setDraftInput((current) => ({
      ...current,
      modelRoutes: current.modelRoutes.map((route, routeIndex) =>
        routeIndex === index ? nextRoute : route),
    }));
  }

  function addProfile() {
    const suffix = draftInput.providerProfiles.length + 1;
    setDraftInput((current) => ({
      ...current,
      providerProfiles: [...current.providerProfiles, {
        profileId: `profile-${suffix}`,
        displayName: `Runtime profile ${suffix}`,
        providerId: "",
        runtimeProfileRef: "",
        capabilities: ["chat_completions"],
      }],
    }));
  }

  function addRoute() {
    const suffix = draftInput.modelRoutes.length + 1;
    setDraftInput((current) => ({
      ...current,
      modelRoutes: [...current.modelRoutes, {
        contractVersion: "v1",
        routeId: `route-${suffix}`,
        protocol: "chat_completions",
        modelId: "",
        providerProfileId: current.providerProfiles[0]?.profileId ?? "",
      }],
    }));
  }

  const live = config.mode === "dev_admin_provider_route_http";
  const focusLabel = focus === "provider"
    ? t($ => $.provider.focusProvider)
    : focus === "profile"
    ? t($ => $.provider.focusProfile)
    : t($ => $.provider.focusRoute);
  return (
    <div
      className="admin-provider-route-workspace"
      id="admin-provider-route-workspace"
      data-resource-focus={focus}
    >
      <div className="model-gateway-overview-subheading admin-provider-route-workspace-heading">
        <div>
          <p className="eyebrow">{t($ => $.provider.workspaceEyebrow, { focus: t($ => $.tasks[focus].label) })}</p>
          <h4>{t($ => $.provider.workspaceTitle)}</h4>
          <p>{focusLabel}</p>
        </div>
        <span className={`status-badge ${live ? "good" : "neutral"}`}>
          {live ? t($ => $.provider.devTest, { environment: config.environment }) : t($ => $.provider.offline)}
        </span>
      </div>

      {!live ? (
        <article className="model-gateway-overview-trace">
          <p className="eyebrow">{t($ => $.provider.explicitSource)}</p>
          <h5>{t($ => $.provider.offlineTitle)}</h5>
          <p>{t($ => $.provider.offlineHelp)}</p>
        </article>
      ) : (
        <>
          <div className="admin-provider-route-scope">
            <dl className="model-gateway-overview-meta">
              <div><dt>{t($ => $.provider.tenant)}</dt><dd>{config.tenantRef}</dd></div>
              <div><dt>{t($ => $.provider.workspace)}</dt><dd>{config.workspaceId}</dd></div>
              <div><dt>{t($ => $.provider.environment)}</dt><dd>{config.environment}</dd></div>
              <div><dt>{t($ => $.provider.configuration)}</dt><dd>{config.configurationId}</dd></div>
              <div><dt>{t($ => $.provider.applicationHandoff)}</dt><dd>{config.applicationId}</dd></div>
              <div><dt>{t($ => $.provider.currentGeneration)}</dt><dd>{snapshot?.generation ?? 0}</dd></div>
            </dl>
            <button type="button" className="secondary-action" onClick={() => void refreshWorkspace()} disabled={operation.status === "loading"}>
              {t($ => $.provider.refresh)}</button>
          </div>

          <OperationStatus operation={operation} />

          <div className="admin-provider-route-layout">
            <section className="admin-provider-route-stage" aria-labelledby="admin-provider-route-draft-title">
              <StageHeading
                titleId="admin-provider-route-draft-title"
                eyebrow={t($ => $.provider.draftStage)}
                title={t($ => $.provider.draftTitle)}
                status={t($ => $.provider.revision, { version: draftInput.expectedRevision })}
              />
              <label>{t($ => $.provider.displayName)}<input
                  value={draftInput.displayName}
                  maxLength={120}
                  onChange={(event) => setDraftInput((current) => ({ ...current, displayName: event.target.value }))}
                />
              </label>

              <div className="admin-provider-route-resource-heading">
                <h6>{t($ => $.provider.focusProfile)}</h6>
                <button type="button" className="secondary-action" onClick={addProfile}>{t($ => $.provider.addProfile)}</button>
              </div>
              {draftInput.providerProfiles.map((profile, index) => (
                <ProviderProfileEditor
                  key={`${index}-${profile.profileId}`}
                  profile={profile}
                  index={index}
                  environment={config.environment}
                  removable={draftInput.providerProfiles.length > 1}
                  onChange={(patch) => updateProfile(index, patch)}
                  onRemove={() => setDraftInput((current) => ({
                    ...current,
                    providerProfiles: current.providerProfiles.filter((_, profileIndex) => profileIndex !== index),
                  }))}
                />
              ))}

              <div className="admin-provider-route-resource-heading">
                <h6>{t($ => $.provider.modelRoutes)}</h6>
                <button type="button" className="secondary-action" onClick={addRoute}>{t($ => $.provider.addRoute)}</button>
              </div>
              {draftInput.modelRoutes.map((route, index) => (
                <ModelRouteEditor
                  key={`${index}-${route.routeId}`}
                  route={route}
                  profiles={draftInput.providerProfiles}
                  removable={draftInput.modelRoutes.length > 1}
                  onChange={(patch) => updateRoute(index, patch)}
                  onRemove={() => setDraftInput((current) => ({
                    ...current,
                    modelRoutes: current.modelRoutes.filter((_, routeIndex) => routeIndex !== index),
                  }))}
                />
              ))}

              <div className="admin-provider-route-findings" aria-live="polite">
                <p className="eyebrow">{t($ => $.provider.contractPreview)}</p>
                {findings.length ? (
                  <ul>{findings.map((finding, index) => <li key={`${finding.field}-${index}`}><strong>{finding.field}</strong> — {providerFindingMessage(t, finding)}</li>)}</ul>
                ) : <p>{t($ => $.provider.noFindings)}</p>}
              </div>
              <button type="button" onClick={() => void saveDraft()} disabled={findings.length > 0 || operation.status === "loading"}>
                {t($ => $.provider.saveDraft)}</button>
            </section>

            <section className="admin-provider-route-stage" aria-labelledby="admin-provider-route-candidate-title">
              <StageHeading
                titleId="admin-provider-route-candidate-title"
                eyebrow={t($ => $.provider.candidateStage)}
                title={t($ => $.provider.candidateTitle)}
                status={candidate ? t($ => $.provider.states[candidate.candidateState]) : t($ => $.provider.notLoaded)}
              />
              <div className="admin-provider-route-inline-form">
                <label>{t($ => $.provider.candidateId)}<input value={candidateId} maxLength={160} onChange={(event) => setCandidateId(event.target.value)} />
                </label>
                <button type="button" onClick={() => void createCandidate()} disabled={operation.status === "loading" || findings.length > 0 || draftInput.expectedRevision < 1}>
                  {t($ => $.provider.createFrom, { version: draftInput.expectedRevision })}
                </button>
                <button type="button" className="secondary-action" onClick={() => void loadCandidate()} disabled={!candidateId.trim() || operation.status === "loading"}>
                  {t($ => $.provider.loadCandidate)}</button>
              </div>
              {candidate ? (
                <>
                  <CandidateSummary candidate={candidate} />
                  <div className="admin-provider-route-diff">
                    <p className="eyebrow">{t($ => $.provider.diffTitle)}</p>
                    <p>{candidateDiff ? t($ => candidateDiff.changed ? $.provider.diffSummary : $.provider.matches, { countText: (formatDisplayNumber(candidateDiff.items.length, locale) ?? t($ => $.provider.unavailable)), generation: candidateDiff.baselineGeneration }) : null}</p>
                    {candidateDiff?.items.length ? (
                      <ul>{candidateDiff.items.map((item) => (
                        <li key={`${item.kind}-${item.resourceId}`}>
                          <span className={`status-badge ${item.change === "removed" ? "bad" : "neutral"}`}>{t($ => $.provider.changes[item.change])}</span>
                          <strong>{t($ => $.provider.kinds[item.kind])} · {item.resourceId}</strong>
                          <small>{item.before || t($ => $.provider.none)} → {item.after || t($ => $.provider.none)}</small>
                        </li>
                      ))}</ul>
                    ) : <p className="boundary-note">{t($ => $.provider.noDiff)}</p>}
                  </div>
                </>
              ) : <p className="boundary-note">{t($ => $.provider.candidateHelp)}</p>}
            </section>

            <section className="admin-provider-route-stage" aria-labelledby="admin-provider-route-review-title">
              <StageHeading
                titleId="admin-provider-route-review-title"
                eyebrow={t($ => $.provider.reviewStage)}
                title={t($ => $.provider.reviewTitle)}
                status={candidate ? t($ => $.provider.reviewVersion, { version: candidate.reviewVersion }) : t($ => $.provider.candidateRequired)}
              />
              <label>{t($ => $.provider.decision)}<select value={reviewDecision} onChange={(event) => setReviewDecision(event.target.value as AdminProviderRouteDecision)}>
                  <option value="approve">{t($ => $.provider.approve)}</option>
                  <option value="reject">{t($ => $.provider.reject)}</option>
                </select>
              </label>
              <label>{t($ => $.provider.reviewReason)}<textarea value={reviewReason} rows={4} maxLength={500} onChange={(event) => setReviewReason(event.target.value)} />
              </label>
              <button type="button" onClick={() => void reviewCandidate()} disabled={!candidate || candidate.candidateState !== "pending_review" || reviewReason.trim().length < 4 || operation.status === "loading"}>
                {t($ => $.provider.recordReview, { version: candidate?.reviewVersion ?? 0 })}
              </button>
              {candidate?.review ? (
                <dl className="model-gateway-overview-meta">
                  <div><dt>{t($ => $.provider.decision)}</dt><dd>{t($ => $.provider.states[candidate.review!.decision])}</dd></div>
                  <div><dt>{t($ => $.provider.state)}</dt><dd>{t($ => $.provider.states[candidate.review!.resultingState])}</dd></div>
                  <div><dt>{t($ => $.provider.reviewer)}</dt><dd>{candidate.review.reviewerRef}</dd></div>
                  <div><dt>{t($ => $.provider.reviewed)}</dt><dd><time dateTime={candidate.review.reviewedAt} title={candidate.review.reviewedAt}>{adminDisplayDate(candidate.review.reviewedAt, locale)}</time></dd></div>
                  <div><dt>{t($ => $.provider.request)}</dt><dd>{candidate.review.requestId}</dd></div>
                  <div><dt>{t($ => $.provider.audit)}</dt><dd>{candidate.review.auditRef}</dd></div>
                </dl>
              ) : null}
              <p className="boundary-note">{t($ => $.provider.approvalBoundary)}</p>
            </section>

            <section className="admin-provider-route-stage" aria-labelledby="admin-provider-route-activation-title">
              <StageHeading
                titleId="admin-provider-route-activation-title"
                eyebrow={t($ => $.provider.activationStage)}
                title={t($ => $.provider.activationTitle)}
                status={t($ => $.provider.generation, { version: snapshot?.generation ?? 0 })}
              />
              <label>{t($ => $.provider.action)}<select value={activationAction} onChange={(event) => setActivationAction(event.target.value as AdminProviderRouteActivationAction)}>
                  <option value="activate">{t($ => $.provider.activateOption)}</option>
                  <option value="rollback">{t($ => $.provider.rollbackOption)}</option>
                </select>
              </label>
              <label>{t($ => $.provider.activationReason)}<textarea value={activationReason} rows={4} maxLength={500} onChange={(event) => setActivationReason(event.target.value)} />
              </label>
              <button type="button" onClick={() => void activateCandidate()} disabled={!candidate || candidate.candidateState !== "approved" || activationReason.trim().length < 4 || operation.status === "loading"}>
                {t($ => activationAction === "rollback" ? $.provider.commitRollback : $.provider.commitActivate, { version: snapshot?.generation ?? 0 })}
              </button>
              <p className="boundary-note">{t($ => $.provider.activationBoundary)}</p>
            </section>
          </div>

          <div className="admin-provider-route-runtime-grid">
            <ActiveSnapshotPanel snapshot={snapshot} applicationId={config.applicationId} />
            <ActivationHistoryPanel
              history={history}
              currentGeneration={snapshot?.generation ?? 0}
              onLoadRollback={(activation) => {
                setActivationAction("rollback");
                setActivationReason(`Roll back to previously activated candidate ${activation.afterCandidateId}.`);
                void loadCandidate(activation.afterCandidateId);
              }}
            />
          </div>
        </>
      )}
    </div>
  );
}

function ProviderProfileEditor({
  profile,
  index,
  environment,
  removable,
  onChange,
  onRemove,
}: {
  profile: AdminProviderProfileAssignment;
  index: number;
  environment: string;
  removable: boolean;
  onChange: (patch: Partial<AdminProviderProfileAssignment>) => void;
  onRemove: () => void;
}) {
  const { t } = useTranslation("admin");
  const { locale } = useLocalePreference();
  function toggleCapability(capability: AdminProviderRouteProtocol, checked: boolean) {
    const next = checked
      ? [...new Set([...profile.capabilities, capability])]
      : profile.capabilities.filter((item) => item !== capability);
    onChange({ capabilities: next });
  }
  return (
    <fieldset className="admin-provider-route-editor">
      <legend>{t($ => $.provider.profileAssignment, { number: (formatDisplayNumber(index + 1, locale) ?? t($ => $.provider.unavailable)) })}</legend>
      <label>{t($ => $.provider.stableProfileId)}<input value={profile.profileId} maxLength={160} onChange={(event) => onChange({ profileId: event.target.value })} /></label>
      <label>{t($ => $.provider.displayName)}<input value={profile.displayName} maxLength={120} onChange={(event) => onChange({ displayName: event.target.value })} /></label>
      <label>{t($ => $.provider.providerId)}<input value={profile.providerId} maxLength={160} onChange={(event) => onChange({ providerId: event.target.value })} /></label>
      <label>{t($ => $.provider.runtimeProfileRef)}<input
          value={profile.runtimeProfileRef}
          maxLength={240}
          placeholder={`ref:radishmind/${environment}/provider-profiles/<profile>`}
          onChange={(event) => onChange({ runtimeProfileRef: event.target.value })}
        />
      </label>
      <div className="admin-provider-route-capabilities">
        <span>{t($ => $.provider.capabilities)}</span>
        {PROTOCOLS.map((capability) => (
          <label key={capability}>
            <input
              type="checkbox"
              checked={profile.capabilities.includes(capability)}
              onChange={(event) => toggleCapability(capability, event.target.checked)}
            />
            {capability}
          </label>
        ))}
      </div>
      {removable ? <button type="button" className="secondary-action" onClick={onRemove}>{t($ => $.provider.removeProfile)}</button> : null}
    </fieldset>
  );
}

function ModelRouteEditor({
  route,
  profiles,
  removable,
  onChange,
  onRemove,
}: {
  route: AdminModelRouteDefinition;
  profiles: AdminProviderProfileAssignment[];
  removable: boolean;
  onChange: (route: AdminModelRouteDefinition) => void;
  onRemove: () => void;
}) {
  const { t } = useTranslation("admin");
  return (
    <fieldset className="admin-provider-route-editor admin-provider-route-model-editor">
      <legend>{route.routeId || t($ => $.provider.modelRoute)}</legend>
      <label>{t($ => $.provider.routeId)}<input value={route.routeId} maxLength={160} onChange={(event) => onChange({ ...route, routeId: event.target.value })} /></label>
      <label>{t($ => $.provider.protocol)}<select value={route.protocol} onChange={(event) => onChange({ ...route, protocol: event.target.value as AdminProviderRouteProtocol })}>
          {PROTOCOLS.map((protocol) => <option value={protocol} key={protocol}>{protocol}</option>)}
        </select>
      </label>
      <label>{t($ => $.provider.requestedModel)}<input value={route.modelId} maxLength={160} onChange={(event) => onChange({ ...route, modelId: event.target.value })} /></label>
      <label>{t($ => $.provider.routeContract)}<select
          value={route.contractVersion}
          onChange={(event) => onChange(routeForContractVersion(route, event.target.value as "v1" | "v2", profiles))}
        >
          <option value="v1">{t($ => $.provider.contractV1)}</option>
          <option value="v2">{t($ => $.provider.contractV2)}</option>
        </select>
      </label>
      {route.contractVersion === "v1" ? (
        <label>{t($ => $.provider.providerProfile)}<select value={route.providerProfileId} onChange={(event) => onChange({ ...route, providerProfileId: event.target.value })}>
            <option value="">{t($ => $.provider.selectProfile)}</option>
            {profiles.map((profile) => <option value={profile.profileId} key={profile.profileId}>{profile.profileId}</option>)}
          </select>
        </label>
      ) : (
        <div className="admin-provider-route-attempt-plan">
          <label>{t($ => $.provider.executionMode)}<select
              value={route.executionMode}
              onChange={(event) => onChange(routeWithExecutionMode(route, event.target.value as AdminProviderRouteExecutionMode, profiles))}
            >
              <option value="single_attempt">{t($ => $.provider.singleAttempt)}</option>
              <option value="sequential_fallback">{t($ => $.provider.sequentialFallback)}</option>
            </select>
          </label>
          <p className="boundary-note">{t($ => $.provider.fallbackBoundary)}</p>
          {route.attemptTargets.map((target, targetIndex) => (
            <label key={target.ordinal}>{target.ordinal === 1 ? t($ => $.provider.primaryTarget) : t($ => $.provider.backupTarget)}
              <select
                value={target.providerProfileId}
                onChange={(event) => onChange({
                  ...route,
                  attemptTargets: route.attemptTargets.map((item, itemIndex) =>
                    itemIndex === targetIndex ? { ...item, providerProfileId: event.target.value } : item),
                })}
              >
                <option value="">{t($ => $.provider.selectProfile)}</option>
                {profiles.map((profile) => <option value={profile.profileId} key={profile.profileId}>{profile.profileId}</option>)}
              </select>
            </label>
          ))}
        </div>
      )}
      {removable ? <button type="button" className="secondary-action" onClick={onRemove}>{t($ => $.provider.removeRoute)}</button> : null}
    </fieldset>
  );
}

function routeForContractVersion(
  route: AdminModelRouteDefinition,
  version: "v1" | "v2",
  profiles: AdminProviderProfileAssignment[],
): AdminModelRouteDefinition {
  const primary = routeTargets(route)[0] || profiles[0]?.profileId || "";
  if (version === "v1") {
    return { contractVersion: "v1", routeId: route.routeId, protocol: route.protocol, modelId: route.modelId, providerProfileId: primary };
  }
  const backup = profiles.find((profile) => profile.profileId !== primary)?.profileId ?? "";
  return {
    contractVersion: "v2", routeId: route.routeId, protocol: route.protocol, modelId: route.modelId,
    executionMode: "sequential_fallback",
    attemptTargets: [{ ordinal: 1, providerProfileId: primary }, { ordinal: 2, providerProfileId: backup }],
  };
}

function routeWithExecutionMode(
  route: Extract<AdminModelRouteDefinition, { contractVersion: "v2" }>,
  executionMode: AdminProviderRouteExecutionMode,
  profiles: AdminProviderProfileAssignment[],
): AdminModelRouteDefinition {
  const primary = route.attemptTargets[0]?.providerProfileId || profiles[0]?.profileId || "";
  if (executionMode === "single_attempt") {
    return { ...route, executionMode, attemptTargets: [{ ordinal: 1, providerProfileId: primary }] };
  }
  const backup = route.attemptTargets[1]?.providerProfileId ||
    profiles.find((profile) => profile.profileId !== primary)?.profileId || "";
  return {
    ...route,
    executionMode,
    attemptTargets: [{ ordinal: 1, providerProfileId: primary }, { ordinal: 2, providerProfileId: backup }],
  };
}

function routeTargets(route: AdminModelRouteDefinition): string[] {
  return route.contractVersion === "v1"
    ? [route.providerProfileId]
    : route.attemptTargets.map((target) => target.providerProfileId);
}

function CandidateSummary({ candidate }: { candidate: AdminProviderRouteCandidate }) {
  const { t } = useTranslation("admin");
  const { locale } = useLocalePreference();
  return (
    <article className="admin-provider-route-candidate-summary">
      <div className="model-gateway-overview-row-main">
        <div><p className="eyebrow">{t($ => $.provider.immutableCandidate)}</p><h6>{candidate.candidateId}</h6></div>
        <span className={`status-badge ${candidate.candidateState === "approved" ? "good" : candidate.candidateState === "rejected" ? "bad" : "neutral"}`}>
          {t($ => $.provider.states[candidate.candidateState])}
        </span>
      </div>
      <dl className="model-gateway-overview-meta">
        <div><dt>{t($ => $.provider.sourceRevision)}</dt><dd>{candidate.sourceDraftRevision}</dd></div>
        <div><dt>{t($ => $.provider.profilesRoutes)}</dt><dd>{candidate.configuration.providerProfiles.length} / {candidate.configuration.modelRoutes.length}</dd></div>
        <div><dt>{t($ => $.provider.candidateDigest)}</dt><dd title={candidate.candidateDigest}>{shortDigest(candidate.candidateDigest)}</dd></div>
        <div><dt>{t($ => $.provider.inventoryBindings)}</dt><dd>{candidate.inventoryBindings.length}</dd></div>
        <div><dt>{t($ => $.provider.createdBy)}</dt><dd>{candidate.createdByActorRef}</dd></div>
        <div><dt>{t($ => $.provider.created)}</dt><dd><time dateTime={candidate.createdAt} title={candidate.createdAt}>{adminDisplayDate(candidate.createdAt, locale)}</time></dd></div>
      </dl>
      <div className="admin-provider-route-plan-review" aria-label={t($ => $.provider.attemptPlans)}>
        {candidate.configuration.modelRoutes.map((route) => (
          <article key={route.routeId}>
            <p><strong>{route.routeId}</strong> · {route.contractVersion === "v1" ? t($ => $.provider.singleAttemptStatus) : t($ => $.provider.states[route.executionMode])}</p>
            <ol>
              {routeTargets(route).map((profileId, index) => (
                <li key={`${route.routeId}-${index}`}>{index === 0 ? t($ => $.provider.primary) : t($ => $.provider.backup)}: {profileId}</li>
              ))}
            </ol>
            {route.contractVersion === "v2" && route.executionMode === "sequential_fallback" ? (
              <small>{t($ => $.provider.fallbackCost)}</small>
            ) : null}
          </article>
        ))}
      </div>
      <div className="admin-provider-route-binding-list">
        {candidate.inventoryBindings.map((binding) => (
          <p key={binding.profileId}>
            <strong>{binding.profileId}</strong> · {binding.providerId} · {binding.capabilities.join(", ")} ·
            <span className={`status-badge ${binding.enabled ? "good" : "bad"}`}>{binding.enabled ? t($ => $.provider.enabled) : t($ => $.provider.disabled)}</span>
            <small title={binding.inventoryDigest}>{shortDigest(binding.inventoryDigest)}</small>
          </p>
        ))}
      </div>
    </article>
  );
}

function ActiveSnapshotPanel({
  snapshot,
  applicationId,
}: {
  snapshot: AdminProviderRouteSnapshot | null;
  applicationId: string;
}) {
  const { t } = useTranslation("admin");
  const { locale } = useLocalePreference();
  return (
    <section className="admin-provider-route-stage admin-provider-route-runtime" aria-labelledby="admin-provider-route-snapshot-title">
      <StageHeading
        titleId="admin-provider-route-snapshot-title"
                eyebrow={t($ => $.provider.snapshotStage)}
        title={t($ => $.provider.snapshotTitle)}
        status={snapshot ? t($ => $.provider.generation, { version: snapshot.generation }) : t($ => $.provider.notActivated)}
      />
      {snapshot ? (
        <>
          <dl className="model-gateway-overview-meta">
            <div><dt>{t($ => $.provider.candidate)}</dt><dd>{snapshot.candidateId}</dd></div>
            <div><dt>{t($ => $.provider.snapshotDigest)}</dt><dd title={snapshot.snapshotDigest}>{shortDigest(snapshot.snapshotDigest)}</dd></div>
            <div><dt>{t($ => $.provider.activatedBy)}</dt><dd>{snapshot.activatedByActorRef}</dd></div>
            <div><dt>{t($ => $.provider.activated)}</dt><dd><time dateTime={snapshot.activatedAt} title={snapshot.activatedAt}>{adminDisplayDate(snapshot.activatedAt, locale)}</time></dd></div>
          </dl>
          <div className="admin-provider-route-snapshot-routes">
            {snapshot.configuration.modelRoutes.map((route) => (
              <article key={route.routeId}>
                <p className="eyebrow">{route.protocol}</p>
                <h6>{route.modelId}</h6>
                <p>{route.routeId} · {route.contractVersion === "v1" ? t($ => $.provider.singleProfile) : t($ => $.provider.states[route.executionMode])}</p>
                <ol className="admin-provider-route-target-list">
                  {routeTargets(route).map((profileId, index) => (
                    <li key={`${route.routeId}-${index}`}>
                      <strong>{index === 0 ? t($ => $.provider.primary) : t($ => $.provider.backup)}</strong> · {profileId || t($ => $.provider.unavailable)}
                    </li>
                  ))}
                </ol>
              </article>
            ))}
          </div>
          <div className="admin-provider-route-handoffs">
            <a href="#model-gateway-playground">{t($ => $.provider.openPlayground)}</a>
            <a href="#model-gateway-request-history">{t($ => $.provider.openHistory)}</a>
          </div>
          <p className="boundary-note">{t($ => $.provider.handoffHelp, { applicationId })}</p>
        </>
      ) : <p className="boundary-note">{t($ => $.provider.noSnapshot)}</p>}
    </section>
  );
}

function ActivationHistoryPanel({
  history,
  currentGeneration,
  onLoadRollback,
}: {
  history: AdminProviderRouteActivation[];
  currentGeneration: number;
  onLoadRollback: (activation: AdminProviderRouteActivation) => void;
}) {
  const { t } = useTranslation("admin");
  const { locale } = useLocalePreference();
  return (
    <section className="admin-provider-route-stage admin-provider-route-runtime" aria-labelledby="admin-provider-route-history-title">
      <StageHeading
        titleId="admin-provider-route-history-title"
                eyebrow={t($ => $.provider.historyStage)}
        title={t($ => $.provider.historyTitle)}
        status={t($ => $.provider.records, { countText: (formatDisplayNumber(history.length, locale) ?? t($ => $.provider.unavailable)) })}
      />
      {history.length ? (
        <div className="admin-provider-route-history-list">
          {[...history].reverse().map((activation) => (
            <article key={activation.activationId}>
              <div className="model-gateway-overview-row-main">
                <div><p className="eyebrow">{t($ => $.provider.states[activation.action])}</p><h6>{t($ => $.provider.generationTransition, { before: activation.beforeGeneration, after: activation.afterGeneration })}</h6></div>
                <span className={`status-badge ${activation.afterGeneration === currentGeneration ? "good" : "neutral"}`}>
                  {activation.afterGeneration === currentGeneration ? t($ => $.provider.current) : t($ => $.provider.historical)}
                </span>
              </div>
              <p>{activation.afterCandidateId}</p>
              <p>{activation.reason}</p>
              <small title={activation.afterSnapshotDigest}>{shortDigest(activation.afterSnapshotDigest)} · <time dateTime={activation.createdAt} title={activation.createdAt}>{adminDisplayDate(activation.createdAt, locale)}</time></small>
              <button type="button" className="secondary-action" onClick={() => onLoadRollback(activation)}>
                {t($ => $.provider.loadRollback)}</button>
            </article>
          ))}
        </div>
      ) : <p className="boundary-note">{t($ => $.provider.noHistory)}</p>}
    </section>
  );
}

function StageHeading({
  titleId,
  eyebrow,
  title,
  status,
}: {
  titleId: string;
  eyebrow: string;
  title: string;
  status: string;
}) {
  return (
    <div className="model-gateway-overview-row-main">
      <div><p className="eyebrow">{eyebrow}</p><h5 id={titleId}>{title}</h5></div>
      <span className="status-badge neutral">{status}</span>
    </div>
  );
}

function OperationStatus({ operation }: { operation: WorkspaceOperation }) {
  const { t } = useTranslation("admin");
  return (
    <div className={`admin-provider-route-operation ${operation.status}`} aria-live="polite">
      <span className={`status-badge ${operation.status === "failed" ? "bad" : operation.status === "ready" ? "good" : "neutral"}`}>
        {t($ => $.provider.states[operation.status])}
      </span>
      <p>{operation.failureCode ? `${operation.failureCode}: ` : ""}{providerOperationMessage(t, operation.message, operation.failureCode)}</p>
      {operation.requestId ? <small>{t($ => $.provider.lineage, { requestId: operation.requestId, auditRef: operation.auditRef })}</small> : null}
    </div>
  );
}

function initialOperation(config: ReturnType<typeof readAdminProviderRouteConfig>): WorkspaceOperation {
  return config.mode === "dev_admin_provider_route_http"
    ? { status: "idle", message: { key: "idle" }, failureCode: "", requestId: "", auditRef: "" }
    : { status: "offline", message: { key: "offline" }, failureCode: "", requestId: "", auditRef: "" };
}

function loadingOperation(message: ProviderMessage): WorkspaceOperation {
  return { status: "loading", message, failureCode: "", requestId: "", auditRef: "" };
}

function readyOperation(message: ProviderMessage, envelope: AdminProviderRouteEnvelope): WorkspaceOperation {
  return {
    status: "ready",
    message,
    failureCode: "",
    requestId: envelope.requestId,
    auditRef: envelope.auditRef,
  };
}

function failedOperation(envelope: AdminProviderRouteEnvelope): WorkspaceOperation {
  return {
    status: "failed",
    message: { key: "failure" },
    failureCode: envelope.failureCode,
    requestId: envelope.requestId,
    auditRef: envelope.auditRef,
  };
}

function operationFromEnvelope(
  envelope: AdminProviderRouteEnvelope,
  successMessage: ProviderMessage,
): WorkspaceOperation {
  return envelope.failureCode ? failedOperation(envelope) : readyOperation(successMessage, envelope);
}

function networkFailure(): WorkspaceOperation {
  return {
    status: "failed",
    message: { key: "unavailable" },
    failureCode: "admin_provider_route_store_unavailable",
    requestId: "",
    auditRef: "",
  };
}

function firstUnexpectedFailure(
  envelopes: AdminProviderRouteEnvelope[],
  expected: string[],
): AdminProviderRouteEnvelope | null {
  return envelopes.find((envelope) => envelope.failureCode && !expected.includes(envelope.failureCode)) ?? null;
}

function shortDigest(value: string): string {
  return value.length > 24 ? `${value.slice(0, 16)}…${value.slice(-8)}` : value;
}

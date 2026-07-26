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
  type AdminProviderRouteProtocol,
  type AdminProviderRouteSnapshot,
} from "./adminProviderRouteConsumer.ts";

const config = readAdminProviderRouteConfig();
const PROTOCOLS: AdminProviderRouteProtocol[] = ["chat_completions", "responses", "messages"];

type WorkspaceOperation = {
  status: "offline" | "idle" | "loading" | "ready" | "failed";
  label: string;
  failureCode: string;
  requestId: string;
  auditRef: string;
};

export function AdminProviderRouteWorkspacePanel() {
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
  const [operation, setOperation] = useState<WorkspaceOperation>(() => initialOperation());
  const findings = useMemo(() => validateAdminProviderRouteDraft(config, draftInput), [draftInput]);
  const candidateDiff = useMemo(
    () => candidate ? buildAdminProviderRouteCandidateDiff(candidate, snapshot) : null,
    [candidate, snapshot],
  );

  useEffect(() => {
    if (config.mode === "dev_admin_provider_route_http") void refreshWorkspace();
  }, []);

  async function refreshWorkspace() {
    setOperation(loadingOperation("Loading draft, active snapshot, and activation history."));
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
        "Workspace synchronized. Missing draft or active snapshot remains an explicit empty state.",
        historyResult,
      ));
    } catch (error) {
      setOperation(networkFailure(error));
    }
  }

  async function saveDraft() {
    if (findings.length) {
      setOperation({
        status: "failed",
        label: "Resolve the local contract findings before saving.",
        failureCode: "admin_provider_route_payload_invalid",
        requestId: "",
        auditRef: "",
      });
      return;
    }
    setOperation(loadingOperation("Saving draft with expected revision."));
    try {
      const result = await saveAdminProviderRouteDraft(config, draftInput);
      if (result.draft) setDraftInput(draftInputFromAdminProviderRouteDraft(result.draft));
      setOperation(operationFromEnvelope(result, "Draft saved without changing the active Gateway snapshot."));
    } catch (error) {
      setOperation(networkFailure(error));
    }
  }

  async function createCandidate() {
    setOperation(loadingOperation("Resolving runtime inventory and creating an immutable candidate."));
    try {
      const result = await createAdminProviderRouteCandidate(
        config,
        candidateId,
        draftInput.expectedRevision,
      );
      if (result.candidate) setCandidate(result.candidate);
      setOperation(operationFromEnvelope(result, "Immutable candidate created from the exact draft revision."));
    } catch (error) {
      setOperation(networkFailure(error));
    }
  }

  async function loadCandidate(id = candidateId) {
    const normalizedId = id.trim();
    if (!normalizedId) return;
    setCandidateId(normalizedId);
    setOperation(loadingOperation(`Loading candidate ${normalizedId}.`));
    try {
      const result = await readAdminProviderRouteCandidate(config, normalizedId);
      if (result.candidate) setCandidate(result.candidate);
      setOperation(operationFromEnvelope(result, `Candidate ${normalizedId} loaded.`));
    } catch (error) {
      setOperation(networkFailure(error));
    }
  }

  async function reviewCandidate() {
    if (!candidate || reviewReason.trim().length < 4) return;
    setOperation(loadingOperation(`Recording independent ${reviewDecision} review.`));
    try {
      const result = await reviewAdminProviderRouteCandidate(
        config,
        candidate.candidateId,
        candidate.reviewVersion,
        reviewDecision,
        reviewReason,
      );
      if (result.candidate) setCandidate(result.candidate);
      setOperation(operationFromEnvelope(result, "Review recorded. Gateway behavior remains unchanged until activation."));
    } catch (error) {
      setOperation(networkFailure(error));
    }
  }

  async function activateCandidate() {
    if (!candidate || activationReason.trim().length < 4) return;
    const generation = snapshot?.generation ?? 0;
    setOperation(loadingOperation(
      `${activationAction === "rollback" ? "Rolling back" : "Activating"} from expected generation ${generation}.`,
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
        `${activationAction === "rollback" ? "Rollback" : "Activation"} committed as a new generation.`,
      ));
    } catch (error) {
      setOperation(networkFailure(error));
    }
  }

  function updateProfile(index: number, patch: Partial<AdminProviderProfileAssignment>) {
    setDraftInput((current) => ({
      ...current,
      providerProfiles: current.providerProfiles.map((profile, profileIndex) =>
        profileIndex === index ? { ...profile, ...patch } : profile),
    }));
  }

  function updateRoute(index: number, patch: Partial<AdminModelRouteDefinition>) {
    setDraftInput((current) => ({
      ...current,
      modelRoutes: current.modelRoutes.map((route, routeIndex) =>
        routeIndex === index ? { ...route, ...patch } : route),
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
        routeId: `route-${suffix}`,
        protocol: "chat_completions",
        modelId: "",
        providerProfileId: current.providerProfiles[0]?.profileId ?? "",
      }],
    }));
  }

  const live = config.mode === "dev_admin_provider_route_http";
  return (
    <div className="admin-provider-route-workspace" id="admin-provider-route-workspace">
      <div className="model-gateway-overview-subheading admin-provider-route-workspace-heading">
        <div>
          <p className="eyebrow">Controlled Configuration Workspace</p>
          <h4>Draft, review, activation, rollback, and Gateway lineage</h4>
        </div>
        <span className={`status-badge ${live ? "good" : "neutral"}`}>
          {live ? `${config.environment} · dev/test` : "offline"}
        </span>
      </div>

      {!live ? (
        <article className="model-gateway-overview-trace">
          <p className="eyebrow">Explicit source required</p>
          <h5>No Provider route management request is sent</h5>
          <p>Enable the Admin Provider route dev/test source to manage drafts and immutable activation snapshots. The existing evidence review remains available above.</p>
        </article>
      ) : (
        <>
          <div className="admin-provider-route-scope">
            <dl className="model-gateway-overview-meta">
              <div><dt>Tenant</dt><dd>{config.tenantRef}</dd></div>
              <div><dt>Workspace</dt><dd>{config.workspaceId}</dd></div>
              <div><dt>Environment</dt><dd>{config.environment}</dd></div>
              <div><dt>Configuration</dt><dd>{config.configurationId}</dd></div>
              <div><dt>Application handoff</dt><dd>{config.applicationId}</dd></div>
              <div><dt>Current generation</dt><dd>{snapshot?.generation ?? 0}</dd></div>
            </dl>
            <button type="button" className="secondary-action" onClick={() => void refreshWorkspace()} disabled={operation.status === "loading"}>
              Refresh workspace
            </button>
          </div>

          <OperationStatus operation={operation} />

          <div className="admin-provider-route-layout">
            <section className="admin-provider-route-stage" aria-labelledby="admin-provider-route-draft-title">
              <StageHeading
                eyebrow="1 · Mutable Draft"
                title="Edit exact runtime assignments and model routes"
                status={`revision ${draftInput.expectedRevision}`}
              />
              <label>Display name
                <input
                  value={draftInput.displayName}
                  maxLength={120}
                  onChange={(event) => setDraftInput((current) => ({ ...current, displayName: event.target.value }))}
                />
              </label>

              <div className="admin-provider-route-resource-heading">
                <h6>Provider Profile assignments</h6>
                <button type="button" className="secondary-action" onClick={addProfile}>Add profile</button>
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
                <h6>Model routes</h6>
                <button type="button" className="secondary-action" onClick={addRoute}>Add route</button>
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
                <p className="eyebrow">Local contract preview</p>
                {findings.length ? (
                  <ul>{findings.map((finding, index) => <li key={`${finding.field}-${index}`}><strong>{finding.field}</strong> — {finding.summary}</li>)}</ul>
                ) : <p>No local contract findings. Server inventory and CAS checks still apply.</p>}
              </div>
              <button type="button" onClick={() => void saveDraft()} disabled={findings.length > 0 || operation.status === "loading"}>
                Save draft revision
              </button>
            </section>

            <section className="admin-provider-route-stage" aria-labelledby="admin-provider-route-candidate-title">
              <StageHeading
                eyebrow="2 · Immutable Candidate"
                title="Resolve inventory and inspect changes"
                status={candidate?.candidateState ?? "not loaded"}
              />
              <div className="admin-provider-route-inline-form">
                <label>Candidate ID
                  <input value={candidateId} maxLength={160} onChange={(event) => setCandidateId(event.target.value)} />
                </label>
                <button type="button" onClick={() => void createCandidate()} disabled={operation.status === "loading" || findings.length > 0 || draftInput.expectedRevision < 1}>
                  Create from revision {draftInput.expectedRevision}
                </button>
                <button type="button" className="secondary-action" onClick={() => void loadCandidate()} disabled={!candidateId.trim() || operation.status === "loading"}>
                  Load candidate
                </button>
              </div>
              {candidate ? (
                <>
                  <CandidateSummary candidate={candidate} />
                  <div className="admin-provider-route-diff">
                    <p className="eyebrow">Candidate vs active snapshot</p>
                    <p>{candidateDiff?.summary}</p>
                    {candidateDiff?.items.length ? (
                      <ul>{candidateDiff.items.map((item) => (
                        <li key={`${item.kind}-${item.resourceId}`}>
                          <span className={`status-badge ${item.change === "removed" ? "bad" : "neutral"}`}>{item.change}</span>
                          <strong>{item.kind} · {item.resourceId}</strong>
                          <small>{item.before || "none"} → {item.after || "none"}</small>
                        </li>
                      ))}</ul>
                    ) : <p className="boundary-note">No configuration difference from the current active snapshot.</p>}
                  </div>
                </>
              ) : <p className="boundary-note">Create or load a candidate to inspect its immutable digest, inventory bindings, and active-snapshot difference.</p>}
            </section>

            <section className="admin-provider-route-stage" aria-labelledby="admin-provider-route-review-title">
              <StageHeading
                eyebrow="3 · Independent Review"
                title="Approve or reject without changing Gateway behavior"
                status={candidate ? `review v${candidate.reviewVersion}` : "candidate required"}
              />
              <label>Decision
                <select value={reviewDecision} onChange={(event) => setReviewDecision(event.target.value as AdminProviderRouteDecision)}>
                  <option value="approve">Approve</option>
                  <option value="reject">Reject</option>
                </select>
              </label>
              <label>Review reason
                <textarea value={reviewReason} rows={4} maxLength={500} onChange={(event) => setReviewReason(event.target.value)} />
              </label>
              <button type="button" onClick={() => void reviewCandidate()} disabled={!candidate || candidate.candidateState !== "pending_review" || reviewReason.trim().length < 4 || operation.status === "loading"}>
                Record review with expected v{candidate?.reviewVersion ?? 0}
              </button>
              {candidate?.review ? (
                <dl className="model-gateway-overview-meta">
                  <div><dt>Decision</dt><dd>{candidate.review.decision}</dd></div>
                  <div><dt>State</dt><dd>{candidate.review.resultingState}</dd></div>
                  <div><dt>Reviewer</dt><dd>{candidate.review.reviewerRef}</dd></div>
                  <div><dt>Reviewed</dt><dd>{formatTimestamp(candidate.review.reviewedAt)}</dd></div>
                  <div><dt>Request</dt><dd>{candidate.review.requestId}</dd></div>
                  <div><dt>Audit</dt><dd>{candidate.review.auditRef}</dd></div>
                </dl>
              ) : null}
              <p className="boundary-note">Approval only changes candidate eligibility. It does not update the active snapshot or affect Gateway requests.</p>
            </section>

            <section className="admin-provider-route-stage" aria-labelledby="admin-provider-route-activation-title">
              <StageHeading
                eyebrow="4 · Explicit Generation Switch"
                title="Activate or roll back with generation CAS"
                status={`generation ${snapshot?.generation ?? 0}`}
              />
              <label>Action
                <select value={activationAction} onChange={(event) => setActivationAction(event.target.value as AdminProviderRouteActivationAction)}>
                  <option value="activate">Activate approved candidate</option>
                  <option value="rollback">Roll back to historical candidate</option>
                </select>
              </label>
              <label>Activation reason
                <textarea value={activationReason} rows={4} maxLength={500} onChange={(event) => setActivationReason(event.target.value)} />
              </label>
              <button type="button" onClick={() => void activateCandidate()} disabled={!candidate || candidate.candidateState !== "approved" || activationReason.trim().length < 4 || operation.status === "loading"}>
                {activationAction === "rollback" ? "Commit rollback" : "Commit activation"} from generation {snapshot?.generation ?? 0}
              </button>
              <p className="boundary-note">Activation revalidates runtime inventory. Missing or drifted bindings fail before the snapshot transaction and before any Provider call.</p>
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
  function toggleCapability(capability: AdminProviderRouteProtocol, checked: boolean) {
    const next = checked
      ? [...new Set([...profile.capabilities, capability])]
      : profile.capabilities.filter((item) => item !== capability);
    onChange({ capabilities: next });
  }
  return (
    <fieldset className="admin-provider-route-editor">
      <legend>Profile assignment {index + 1}</legend>
      <label>Stable profile ID<input value={profile.profileId} maxLength={160} onChange={(event) => onChange({ profileId: event.target.value })} /></label>
      <label>Display name<input value={profile.displayName} maxLength={120} onChange={(event) => onChange({ displayName: event.target.value })} /></label>
      <label>Provider ID<input value={profile.providerId} maxLength={160} onChange={(event) => onChange({ providerId: event.target.value })} /></label>
      <label>Runtime profile ref
        <input
          value={profile.runtimeProfileRef}
          maxLength={240}
          placeholder={`ref:radishmind/${environment}/provider-profiles/<profile>`}
          onChange={(event) => onChange({ runtimeProfileRef: event.target.value })}
        />
      </label>
      <div className="admin-provider-route-capabilities">
        <span>Capabilities</span>
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
      {removable ? <button type="button" className="secondary-action" onClick={onRemove}>Remove profile</button> : null}
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
  onChange: (patch: Partial<AdminModelRouteDefinition>) => void;
  onRemove: () => void;
}) {
  return (
    <fieldset className="admin-provider-route-editor admin-provider-route-model-editor">
      <legend>{route.routeId || "Model route"}</legend>
      <label>Route ID<input value={route.routeId} maxLength={160} onChange={(event) => onChange({ routeId: event.target.value })} /></label>
      <label>Protocol
        <select value={route.protocol} onChange={(event) => onChange({ protocol: event.target.value as AdminProviderRouteProtocol })}>
          {PROTOCOLS.map((protocol) => <option value={protocol} key={protocol}>{protocol}</option>)}
        </select>
      </label>
      <label>Requested model<input value={route.modelId} maxLength={160} onChange={(event) => onChange({ modelId: event.target.value })} /></label>
      <label>Provider Profile
        <select value={route.providerProfileId} onChange={(event) => onChange({ providerProfileId: event.target.value })}>
          <option value="">Select profile</option>
          {profiles.map((profile) => <option value={profile.profileId} key={profile.profileId}>{profile.profileId}</option>)}
        </select>
      </label>
      {removable ? <button type="button" className="secondary-action" onClick={onRemove}>Remove route</button> : null}
    </fieldset>
  );
}

function CandidateSummary({ candidate }: { candidate: AdminProviderRouteCandidate }) {
  return (
    <article className="admin-provider-route-candidate-summary">
      <div className="model-gateway-overview-row-main">
        <div><p className="eyebrow">Immutable candidate</p><h6>{candidate.candidateId}</h6></div>
        <span className={`status-badge ${candidate.candidateState === "approved" ? "good" : candidate.candidateState === "rejected" ? "bad" : "neutral"}`}>
          {candidate.candidateState}
        </span>
      </div>
      <dl className="model-gateway-overview-meta">
        <div><dt>Source revision</dt><dd>{candidate.sourceDraftRevision}</dd></div>
        <div><dt>Profiles / routes</dt><dd>{candidate.configuration.providerProfiles.length} / {candidate.configuration.modelRoutes.length}</dd></div>
        <div><dt>Candidate digest</dt><dd title={candidate.candidateDigest}>{shortDigest(candidate.candidateDigest)}</dd></div>
        <div><dt>Inventory bindings</dt><dd>{candidate.inventoryBindings.length}</dd></div>
        <div><dt>Created by</dt><dd>{candidate.createdByActorRef}</dd></div>
        <div><dt>Created</dt><dd>{formatTimestamp(candidate.createdAt)}</dd></div>
      </dl>
      <div className="admin-provider-route-binding-list">
        {candidate.inventoryBindings.map((binding) => (
          <p key={binding.profileId}>
            <strong>{binding.profileId}</strong> · {binding.providerId} · {binding.capabilities.join(", ")} ·
            <span className={`status-badge ${binding.enabled ? "good" : "bad"}`}>{binding.enabled ? "enabled" : "disabled"}</span>
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
  return (
    <section className="admin-provider-route-stage admin-provider-route-runtime" aria-labelledby="admin-provider-route-snapshot-title">
      <StageHeading
        eyebrow="Current Runtime Snapshot"
        title="Gateway consumes this immutable generation"
        status={snapshot ? `generation ${snapshot.generation}` : "not activated"}
      />
      {snapshot ? (
        <>
          <dl className="model-gateway-overview-meta">
            <div><dt>Candidate</dt><dd>{snapshot.candidateId}</dd></div>
            <div><dt>Snapshot digest</dt><dd title={snapshot.snapshotDigest}>{shortDigest(snapshot.snapshotDigest)}</dd></div>
            <div><dt>Activated by</dt><dd>{snapshot.activatedByActorRef}</dd></div>
            <div><dt>Activated</dt><dd>{formatTimestamp(snapshot.activatedAt)}</dd></div>
          </dl>
          <div className="admin-provider-route-snapshot-routes">
            {snapshot.configuration.modelRoutes.map((route) => (
              <article key={route.routeId}>
                <p className="eyebrow">{route.protocol}</p>
                <h6>{route.modelId}</h6>
                <p>{route.routeId} → {route.providerProfileId}</p>
              </article>
            ))}
          </div>
          <div className="admin-provider-route-handoffs">
            <a href="#model-gateway-playground">Open Gateway Playground</a>
            <a href="#model-gateway-request-history">Review Gateway history lineage</a>
          </div>
          <p className="boundary-note">Use application {applicationId} and an application-scoped API key. Request History must report this configuration, generation, and snapshot digest.</p>
        </>
      ) : <p className="boundary-note">No active snapshot exists. Approval alone intentionally leaves Gateway behavior unchanged.</p>}
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
  return (
    <section className="admin-provider-route-stage admin-provider-route-runtime" aria-labelledby="admin-provider-route-history-title">
      <StageHeading
        eyebrow="Append-only Activation History"
        title="Generation lineage and rollback targets"
        status={`${history.length} records`}
      />
      {history.length ? (
        <div className="admin-provider-route-history-list">
          {[...history].reverse().map((activation) => (
            <article key={activation.activationId}>
              <div className="model-gateway-overview-row-main">
                <div><p className="eyebrow">{activation.action}</p><h6>generation {activation.beforeGeneration} → {activation.afterGeneration}</h6></div>
                <span className={`status-badge ${activation.afterGeneration === currentGeneration ? "good" : "neutral"}`}>
                  {activation.afterGeneration === currentGeneration ? "current" : "historical"}
                </span>
              </div>
              <p>{activation.afterCandidateId}</p>
              <p>{activation.reason}</p>
              <small title={activation.afterSnapshotDigest}>{shortDigest(activation.afterSnapshotDigest)} · {formatTimestamp(activation.createdAt)}</small>
              <button type="button" className="secondary-action" onClick={() => onLoadRollback(activation)}>
                Load as rollback target
              </button>
            </article>
          ))}
        </div>
      ) : <p className="boundary-note">No activation record exists. Candidate review does not append activation history.</p>}
    </section>
  );
}

function StageHeading({
  eyebrow,
  title,
  status,
}: {
  eyebrow: string;
  title: string;
  status: string;
}) {
  return (
    <div className="model-gateway-overview-row-main">
      <div><p className="eyebrow">{eyebrow}</p><h5>{title}</h5></div>
      <span className="status-badge neutral">{status}</span>
    </div>
  );
}

function OperationStatus({ operation }: { operation: WorkspaceOperation }) {
  return (
    <div className={`admin-provider-route-operation ${operation.status}`} aria-live="polite">
      <span className={`status-badge ${operation.status === "failed" ? "bad" : operation.status === "ready" ? "good" : "neutral"}`}>
        {operation.status}
      </span>
      <p>{operation.failureCode ? `${operation.failureCode}: ` : ""}{operation.label}</p>
      {operation.requestId ? <small>request {operation.requestId} · audit {operation.auditRef}</small> : null}
    </div>
  );
}

function initialOperation(): WorkspaceOperation {
  return config.mode === "dev_admin_provider_route_http"
    ? { status: "idle", label: "Ready to load the controlled configuration workspace.", failureCode: "", requestId: "", auditRef: "" }
    : { status: "offline", label: "Offline evidence mode sends no management request.", failureCode: "", requestId: "", auditRef: "" };
}

function loadingOperation(label: string): WorkspaceOperation {
  return { status: "loading", label, failureCode: "", requestId: "", auditRef: "" };
}

function readyOperation(label: string, envelope: AdminProviderRouteEnvelope): WorkspaceOperation {
  return {
    status: "ready",
    label,
    failureCode: "",
    requestId: envelope.requestId,
    auditRef: envelope.auditRef,
  };
}

function failedOperation(envelope: AdminProviderRouteEnvelope): WorkspaceOperation {
  return {
    status: "failed",
    label: failureSummary(envelope.failureCode),
    failureCode: envelope.failureCode,
    requestId: envelope.requestId,
    auditRef: envelope.auditRef,
  };
}

function operationFromEnvelope(
  envelope: AdminProviderRouteEnvelope,
  successLabel: string,
): WorkspaceOperation {
  return envelope.failureCode ? failedOperation(envelope) : readyOperation(successLabel, envelope);
}

function networkFailure(error: unknown): WorkspaceOperation {
  return {
    status: "failed",
    label: error instanceof Error ? error.message : "Admin Provider route workspace is unavailable.",
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

function failureSummary(failureCode: string): string {
  const summaries: Record<string, string> = {
    admin_provider_route_draft_revision_conflict: "The draft changed. Refresh before applying edits to the current revision.",
    admin_provider_route_review_version_conflict: "The candidate review changed. Reload the candidate before deciding again.",
    admin_provider_route_generation_conflict: "The active generation changed. Refresh the workspace before activation.",
    admin_provider_route_inventory_not_found: "A referenced runtime profile is absent from the current inventory.",
    admin_provider_route_inventory_mismatch: "Runtime inventory drifted after candidate creation. Create and review a new candidate.",
    admin_provider_route_inventory_unavailable: "Runtime inventory cannot be read. No candidate or snapshot was written.",
    admin_provider_route_candidate_not_approved: "Only an independently approved candidate can be activated.",
    admin_provider_route_rollback_target_invalid: "Rollback requires a previously activated approved candidate.",
  };
  return summaries[failureCode] ?? "The operation failed without changing the controlled configuration state.";
}

function shortDigest(value: string): string {
  return value.length > 24 ? `${value.slice(0, 16)}…${value.slice(-8)}` : value;
}

function formatTimestamp(value: string): string {
  const timestamp = new Date(value);
  return Number.isNaN(timestamp.valueOf()) ? value : timestamp.toLocaleString();
}

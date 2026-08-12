import { useEffect, useMemo, useRef, useState } from "react";

import {
  isValidAdminGatewayModelPricingRate,
  isValidAdminGatewayModelPricingReason,
  isValidAdminGatewayModelPricingScope,
  putAdminGatewayModelPricing,
  readAdminGatewayModelPricing,
  readAdminGatewayModelPricingConfig,
  type AdminGatewayModelPricingEnvelope,
  type AdminGatewayModelPricingScope,
} from "./adminGatewayModelPricingConsumer.ts";

type LoadState = "idle" | "loading" | "ready" | "missing" | "failed";

type PricingUpdateReview = {
  expectedVersion: number;
  inputRate: number;
  outputRate: number;
  reason: string;
};

export function AdminGatewayModelPricingPanel({
  tenantRef,
  workspaceId,
}: {
  tenantRef: string;
  workspaceId: string;
}) {
  const config = useMemo(
    () => readAdminGatewayModelPricingConfig({ tenantRef, workspaceId }),
    [tenantRef, workspaceId],
  );
  const [scopeDraft, setScopeDraft] = useState<AdminGatewayModelPricingScope>(config.initialScope);
  const [loadedScope, setLoadedScope] = useState<AdminGatewayModelPricingScope | null>(null);
  const [envelope, setEnvelope] = useState<AdminGatewayModelPricingEnvelope | null>(null);
  const [loadState, setLoadState] = useState<LoadState>("idle");
  const [inputRate, setInputRate] = useState("0");
  const [outputRate, setOutputRate] = useState("0");
  const [reason, setReason] = useState("");
  const [review, setReview] = useState<PricingUpdateReview | null>(null);
  const [operationMessage, setOperationMessage] = useState("");
  const [operationPending, setOperationPending] = useState(false);
  const [reloadRequired, setReloadRequired] = useState(false);
  const generation = useRef(0);

  useEffect(() => {
    generation.current += 1;
    setScopeDraft(config.initialScope);
    setLoadedScope(null);
    setEnvelope(null);
    setLoadState("idle");
    setInputRate("0");
    setOutputRate("0");
    setReason("");
    setReview(null);
    setOperationMessage("");
    setOperationPending(false);
    setReloadRequired(false);
  }, [config]);

  async function loadPricingOwner() {
    const scope = normalizeScope(scopeDraft);
    if (!isValidAdminGatewayModelPricingScope(scope)) {
      setLoadState("failed");
      setOperationMessage("Provider, profile, and model must form one exact valid scope.");
      return;
    }
    const requestGeneration = generation.current + 1;
    generation.current = requestGeneration;
    setLoadedScope(scope);
    setEnvelope(null);
    setLoadState("loading");
    setReview(null);
    setReloadRequired(false);
    setOperationMessage("");
    try {
      const next = await readAdminGatewayModelPricing(config, scope);
      if (generation.current !== requestGeneration) return;
      setEnvelope(next);
      if (next.failureCode === "gateway_pricing_policy_not_found") {
        setLoadState("missing");
        setInputRate("0");
        setOutputRate("0");
        setReason("");
        return;
      }
      if (next.failureCode) {
        setLoadState("failed");
        setOperationMessage(pricingFailureSummary(next.failureCode));
        return;
      }
      setLoadState("ready");
      setInputRate(String(next.policy?.inputPriceMicrosPerTokenUnit ?? 0));
      setOutputRate(String(next.policy?.outputPriceMicrosPerTokenUnit ?? 0));
      setReason(next.policy?.reason ?? "");
    } catch (error) {
      if (generation.current !== requestGeneration) return;
      setLoadState("failed");
      setOperationMessage(error instanceof Error ? error.message : "Pricing owner is unavailable.");
    }
  }

  function reviewUpdate() {
    const parsedInputRate = parseNonNegativeSafeInteger(inputRate);
    const parsedOutputRate = parseNonNegativeSafeInteger(outputRate);
    if (!loadedScope || parsedInputRate === null || parsedOutputRate === null ||
      !isValidAdminGatewayModelPricingReason(reason)) {
      setOperationMessage("Enter non-negative integer micro-USD rates and a sanitized update reason.");
      setReview(null);
      return;
    }
    setOperationMessage("");
    setReview({
      expectedVersion: envelope?.policy?.recordVersion ?? 0,
      inputRate: parsedInputRate,
      outputRate: parsedOutputRate,
      reason: reason.trim(),
    });
  }

  async function confirmUpdate() {
    if (!loadedScope || !review || reloadRequired) return;
    const requestGeneration = generation.current;
    setOperationPending(true);
    setOperationMessage("");
    try {
      const next = await putAdminGatewayModelPricing(config, loadedScope, {
        expectedVersion: review.expectedVersion,
        inputPriceMicrosPerTokenUnit: review.inputRate,
        outputPriceMicrosPerTokenUnit: review.outputRate,
        reason: review.reason,
      });
      if (generation.current !== requestGeneration) return;
      setEnvelope(next);
      setReview(null);
      if (next.failureCode === "gateway_pricing_policy_version_conflict") {
        setLoadState("failed");
        setReloadRequired(true);
        setOperationMessage(`Version conflict: current version is ${next.currentVersion}. Reload before reviewing another update.`);
        return;
      }
      if (next.failureCode || !next.policy) {
        setLoadState("failed");
        setOperationMessage(pricingFailureSummary(next.failureCode ?? "gateway_pricing_store_unavailable"));
        return;
      }
      setLoadState("ready");
      setInputRate(String(next.policy.inputPriceMicrosPerTokenUnit));
      setOutputRate(String(next.policy.outputPriceMicrosPerTokenUnit));
      setReason(next.policy.reason);
      setOperationMessage(`Revision v${next.policy.recordVersion} is current for future request snapshots.`);
    } catch (error) {
      if (generation.current !== requestGeneration) return;
      setLoadState("failed");
      setOperationMessage(error instanceof Error ? error.message : "Pricing update failed closed.");
    } finally {
      if (generation.current === requestGeneration) setOperationPending(false);
    }
  }

  const scopeDirty = loadedScope !== null && !sameScope(scopeDraft, loadedScope);
  const editorEnabled = (loadState === "ready" || loadState === "missing") && !scopeDirty && !reloadRequired;

  return (
    <div className="admin-gateway-pricing-workspace">
      <section className="admin-gateway-pricing-scope" aria-labelledby="admin-gateway-pricing-scope-title">
        <header>
          <div><p className="eyebrow">Exact pricing scope</p><h5 id="admin-gateway-pricing-scope-title">Provider / Profile / Model</h5></div>
          <span className={`admin-control-status is-${config.mode === "dev_admin_gateway_model_pricing_http" ? "ready" : "neutral"}`}>
            {config.mode === "dev_admin_gateway_model_pricing_http" ? config.environment : "offline"}
          </span>
        </header>
        <div className="admin-gateway-pricing-scope-fields">
          <label>Provider ID<input value={scopeDraft.providerId} onChange={(event) => setScopeDraft({ ...scopeDraft, providerId: event.target.value })} /></label>
          <label>Profile ID<input value={scopeDraft.profileId} onChange={(event) => setScopeDraft({ ...scopeDraft, profileId: event.target.value })} /></label>
          <label>Model ID<input value={scopeDraft.modelId} onChange={(event) => setScopeDraft({ ...scopeDraft, modelId: event.target.value })} /></label>
        </div>
        <button type="button" className="secondary-action" onClick={() => void loadPricingOwner()} disabled={loadState === "loading" || operationPending}>
          {loadState === "loading" ? "Loading exact owner…" : "Load exact pricing owner"}
        </button>
        <dl className="admin-control-owner-meta">
          <div><dt>Tenant</dt><dd>{config.tenantRef}</dd></div>
          <div><dt>Workspace</dt><dd>{config.workspaceId}</dd></div>
          <div><dt>Environment</dt><dd>{config.environment}</dd></div>
          <div><dt>Currency / unit</dt><dd>USD / 1M tokens</dd></div>
        </dl>
        <p className="admin-control-boundary-notice"><span aria-hidden="true">!</span>Exact scope only. No alias, wildcard, route fallback, credential, endpoint, invoice, or production price is read.</p>
      </section>

      <section className="admin-gateway-pricing-owner" aria-live="polite">
        <header>
          <div>
            <p className="eyebrow">Immutable revision owner</p>
            <h5>{loadedScope ? `${loadedScope.providerId} / ${loadedScope.profileId} / ${loadedScope.modelId}` : "Load one exact scope"}</h5>
          </div>
          <span className={`admin-control-status is-${loadState === "ready" ? "ready" : loadState === "failed" ? "blocked" : "neutral"}`}>
            {loadState.replaceAll("_", " ")}
          </span>
        </header>

        {config.mode === "offline" ? (
          <PricingNotice title="Pricing management is offline">Enable the explicit development/test pricing source. No route fixture or current public price is used as a fallback.</PricingNotice>
        ) : loadState === "idle" || loadState === "loading" ? (
          <PricingNotice title={loadState === "loading" ? "Reading current immutable revision…" : "Choose the exact selection lineage"}>The owner is not read until Provider, Profile, and Model are explicitly loaded.</PricingNotice>
        ) : null}

        {envelope?.policy && loadState === "ready" ? <PricingPolicyEvidence policy={envelope.policy} /> : null}
        {loadState === "missing" ? (
          <PricingNotice title="No policy exists for this exact scope">Create revision v1 with expected_version = 0. Missing price does not block Provider execution and is not replaced by another model price.</PricingNotice>
        ) : null}
        {loadState === "failed" ? (
          <div className="admin-gateway-pricing-failure" role="alert">
            <strong>{envelope?.failureCode ?? "gateway_pricing_store_unavailable"}</strong>
            <p>{operationMessage || "Pricing evidence failed closed."}</p>
            {reloadRequired ? <button type="button" className="secondary-action" onClick={() => void loadPricingOwner()}>Reload current revision</button> : null}
          </div>
        ) : null}

        {editorEnabled ? (
          <PricingUpdateEditor
            currentVersion={envelope?.policy?.recordVersion ?? 0}
            inputRate={inputRate}
            outputRate={outputRate}
            reason={reason}
            review={review}
            pending={operationPending}
            message={operationMessage}
            onInputRate={setInputRate}
            onOutputRate={setOutputRate}
            onReason={setReason}
            onReview={reviewUpdate}
            onCancel={() => setReview(null)}
            onConfirm={() => void confirmUpdate()}
          />
        ) : null}

        {scopeDirty ? <p className="admin-gateway-pricing-failure" role="alert">Scope changed. Load the exact owner before editing; the previous revision is no longer actionable.</p> : null}
        <p className="admin-gateway-pricing-boundary"><span aria-hidden="true">!</span>Updates affect future request-local snapshots only. Historical estimates are immutable and this surface does not create billing, quota, budget, settlement, or routing decisions.</p>
      </section>
    </div>
  );
}

function PricingPolicyEvidence({ policy }: { policy: NonNullable<AdminGatewayModelPricingEnvelope["policy"]> }) {
  return (
    <div className="admin-gateway-pricing-evidence">
      <div className="admin-gateway-pricing-rates">
        <article><span>Input</span><strong>{formatRate(policy.inputPriceMicrosPerTokenUnit)}</strong><small>USD / 1M tokens</small></article>
        <article><span>Output</span><strong>{formatRate(policy.outputPriceMicrosPerTokenUnit)}</strong><small>USD / 1M tokens</small></article>
      </div>
      <dl className="admin-control-owner-meta">
        <div><dt>Policy / revision</dt><dd>{policy.policyId} · v{policy.recordVersion}</dd></div>
        <div><dt>Digest</dt><dd>{shortDigest(policy.policyDigest)}</dd></div>
        <div><dt>Updated</dt><dd>{formatTimestamp(policy.updatedAt)}</dd></div>
        <div><dt>Updated by</dt><dd>{policy.updatedByActorRef}</dd></div>
        <div><dt>Reason</dt><dd>{policy.reason}</dd></div>
        <div><dt>Request / audit</dt><dd>{policy.requestId} / {policy.auditRef}</dd></div>
      </dl>
    </div>
  );
}

function PricingUpdateEditor({
  currentVersion,
  inputRate,
  outputRate,
  reason,
  review,
  pending,
  message,
  onInputRate,
  onOutputRate,
  onReason,
  onReview,
  onCancel,
  onConfirm,
}: {
  currentVersion: number;
  inputRate: string;
  outputRate: string;
  reason: string;
  review: PricingUpdateReview | null;
  pending: boolean;
  message: string;
  onInputRate: (value: string) => void;
  onOutputRate: (value: string) => void;
  onReason: (value: string) => void;
  onReview: () => void;
  onCancel: () => void;
  onConfirm: () => void;
}) {
  return (
    <div className="admin-gateway-pricing-update">
      <div className="admin-gateway-pricing-editor">
        <div className="card-title-row"><div><p className="eyebrow">Proposed revision</p><h5>Expected version {currentVersion}</h5></div><span>CAS</span></div>
        <div className="admin-gateway-pricing-rate-fields">
          <label>Input micro-USD / 1M<input inputMode="numeric" value={inputRate} onChange={(event) => onInputRate(event.target.value)} /></label>
          <label>Output micro-USD / 1M<input inputMode="numeric" value={outputRate} onChange={(event) => onOutputRate(event.target.value)} /></label>
        </div>
        <label>Sanitized reason<textarea rows={3} value={reason} onChange={(event) => onReason(event.target.value)} /></label>
        <button type="button" className="secondary-action" onClick={onReview} disabled={pending}>Review revision</button>
        {message ? <p className="admin-gateway-pricing-message" role="status">{message}</p> : null}
      </div>
      {review ? (
        <aside className="admin-gateway-pricing-confirm" aria-label="Pricing revision confirmation">
          <p className="eyebrow">Explicit confirmation</p>
          <h5>v{review.expectedVersion} → v{review.expectedVersion + 1}</h5>
          <dl>
            <div><dt>Input</dt><dd>{formatRate(review.inputRate)}</dd></div>
            <div><dt>Output</dt><dd>{formatRate(review.outputRate)}</dd></div>
            <div><dt>Unit</dt><dd>USD / 1M tokens</dd></div>
            <div><dt>History</dt><dd>not recalculated</dd></div>
          </dl>
          <p>{review.reason}</p>
          <div><button type="button" className="secondary-action" onClick={onCancel} disabled={pending}>Cancel</button><button type="button" onClick={onConfirm} disabled={pending}>{pending ? "Writing…" : "Confirm future pricing"}</button></div>
        </aside>
      ) : null}
    </div>
  );
}

function PricingNotice({ title, children }: { title: string; children: string }) {
  return <div className="admin-gateway-pricing-notice"><span aria-hidden="true">∅</span><div><strong>{title}</strong><p>{children}</p></div></div>;
}

function normalizeScope(scope: AdminGatewayModelPricingScope): AdminGatewayModelPricingScope {
  return { providerId: scope.providerId.trim(), profileId: scope.profileId.trim(), modelId: scope.modelId.trim() };
}

function sameScope(left: AdminGatewayModelPricingScope, right: AdminGatewayModelPricingScope): boolean {
  const normalized = normalizeScope(left);
  return normalized.providerId === right.providerId && normalized.profileId === right.profileId && normalized.modelId === right.modelId;
}

function parseNonNegativeSafeInteger(value: string): number | null {
  if (!/^\d+$/u.test(value.trim())) return null;
  const parsed = Number(value.trim());
  return isValidAdminGatewayModelPricingRate(parsed) ? parsed : null;
}

function pricingFailureSummary(failureCode: string): string {
  const summaries: Record<string, string> = {
    gateway_pricing_disabled: "The development/test pricing management gate is disabled.",
    gateway_pricing_scope_denied: "Verified membership does not grant the pricing permission for this workspace.",
    gateway_pricing_environment_forbidden: "The requested environment does not match the enabled development/test owner.",
    gateway_pricing_payload_invalid: "The pricing request failed strict contract validation.",
    gateway_pricing_policy_not_found: "No pricing revision exists for the exact selection scope.",
    gateway_pricing_policy_version_conflict: "Another writer changed the current revision.",
    gateway_pricing_policy_scope_conflict: "The returned owner does not match the exact selection scope.",
    gateway_pricing_store_unavailable: "The pricing repository is unavailable and no fallback was used.",
  };
  return summaries[failureCode] ?? "Pricing evidence failed closed.";
}

function formatRate(value: number): string {
  return `$${(value / 1_000_000).toFixed(6)}`;
}

function shortDigest(value: string): string {
  return value.length > 24 ? `${value.slice(0, 16)}…${value.slice(-8)}` : value;
}

function formatTimestamp(value: string): string {
  const timestamp = new Date(value);
  return Number.isNaN(timestamp.valueOf()) ? value : timestamp.toLocaleString();
}

import { useTranslation } from "react-i18next";
import { useLocalePreference } from "../../i18n/LocaleProvider.tsx";
import { formatDisplayNumber } from "../../i18n/formatters.ts";
import { adminDisplayDate, adminDisplayRate } from "./adminManagementFormatters.ts";
import { pricingNoticeMessage, type PricingNoticeMessage } from "./adminGatewayManagementMessages.ts";
import "../../i18n/adminPricingResources.ts";
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
  const { t } = useTranslation("admin");
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
  const [operationNotice, setOperationNotice] = useState<PricingNoticeMessage>(null);
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
    setOperationNotice(null);
    setOperationPending(false);
    setReloadRequired(false);
  }, [config]);

  async function loadPricingOwner() {
    const scope = normalizeScope(scopeDraft);
    if (!isValidAdminGatewayModelPricingScope(scope)) {
      setLoadState("failed");
      setOperationNotice({ key: "invalidScope" });
      return;
    }
    const requestGeneration = generation.current + 1;
    generation.current = requestGeneration;
    setLoadedScope(scope);
    setEnvelope(null);
    setLoadState("loading");
    setReview(null);
    setReloadRequired(false);
    setOperationNotice(null);
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
        setOperationNotice({ key: "failure", code: next.failureCode });
        return;
      }
      setLoadState("ready");
      setInputRate(String(next.policy?.inputPriceMicrosPerTokenUnit ?? 0));
      setOutputRate(String(next.policy?.outputPriceMicrosPerTokenUnit ?? 0));
      setReason(next.policy?.reason ?? "");
    } catch {
      if (generation.current !== requestGeneration) return;
      setLoadState("failed");
      setOperationNotice({ key: "unavailable" });
    }
  }

  function reviewUpdate() {
    const parsedInputRate = parseNonNegativeSafeInteger(inputRate);
    const parsedOutputRate = parseNonNegativeSafeInteger(outputRate);
    if (!loadedScope || parsedInputRate === null || parsedOutputRate === null ||
      !isValidAdminGatewayModelPricingReason(reason)) {
      setOperationNotice({ key: "invalidUpdate" });
      setReview(null);
      return;
    }
    setOperationNotice(null);
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
    setOperationNotice(null);
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
        setOperationNotice({ key: "conflict", version: next.currentVersion });
        return;
      }
      if (next.failureCode || !next.policy) {
        setLoadState("failed");
        setOperationNotice({ key: "failure", code: next.failureCode ?? "gateway_pricing_store_unavailable" });
        return;
      }
      setLoadState("ready");
      setInputRate(String(next.policy.inputPriceMicrosPerTokenUnit));
      setOutputRate(String(next.policy.outputPriceMicrosPerTokenUnit));
      setReason(next.policy.reason);
      setOperationNotice({ key: "updated", version: next.policy.recordVersion });
    } catch {
      if (generation.current !== requestGeneration) return;
      setLoadState("failed");
      setOperationNotice({ key: "unavailable" });
    } finally {
      if (generation.current === requestGeneration) setOperationPending(false);
    }
  }

  const operationMessage = pricingNoticeMessage(t, operationNotice);
  const scopeDirty = loadedScope !== null && !sameScope(scopeDraft, loadedScope);
  const editorEnabled = (loadState === "ready" || loadState === "missing") && !scopeDirty && !reloadRequired;

  return (
    <div className="admin-gateway-pricing-workspace">
      <section className="admin-gateway-pricing-scope" aria-labelledby="admin-gateway-pricing-scope-title">
        <header>
          <div><p className="eyebrow">{t($ => $.pricing.scopeTitle)}</p><h5 id="admin-gateway-pricing-scope-title">{t($ => $.pricing.scopeHeading)}</h5></div>
          <span className={`admin-control-status is-${config.mode === "dev_admin_gateway_model_pricing_http" ? "ready" : "neutral"}`}>
            {config.mode === "dev_admin_gateway_model_pricing_http" ? config.environment : t($ => $.pricing.offline)}
          </span>
        </header>
        <div className="admin-gateway-pricing-scope-fields">
          <label>{t($ => $.pricing.providerId)}<input value={scopeDraft.providerId} onChange={(event) => setScopeDraft({ ...scopeDraft, providerId: event.target.value })} /></label>
          <label>{t($ => $.pricing.profileId)}<input value={scopeDraft.profileId} onChange={(event) => setScopeDraft({ ...scopeDraft, profileId: event.target.value })} /></label>
          <label>{t($ => $.pricing.modelId)}<input value={scopeDraft.modelId} onChange={(event) => setScopeDraft({ ...scopeDraft, modelId: event.target.value })} /></label>
        </div>
        <button type="button" className="secondary-action" onClick={() => void loadPricingOwner()} disabled={loadState === "loading" || operationPending}>
          {loadState === "loading" ? t($ => $.pricing.loadingOwner) : t($ => $.pricing.loadOwner)}
        </button>
        <dl className="admin-control-owner-meta">
          <div><dt>{t($ => $.pricing.tenant)}</dt><dd>{config.tenantRef}</dd></div>
          <div><dt>{t($ => $.pricing.workspace)}</dt><dd>{config.workspaceId}</dd></div>
          <div><dt>{t($ => $.pricing.environment)}</dt><dd>{config.environment}</dd></div>
          <div><dt>{t($ => $.pricing.currencyUnit)}</dt><dd>{t($ => $.pricing.unitValue)}</dd></div>
        </dl>
        <p className="admin-control-boundary-notice"><span aria-hidden="true">!</span>{t($ => $.pricing.scopeBoundary)}</p>
      </section>

      <section className="admin-gateway-pricing-owner" aria-live="polite">
        <header>
          <div>
            <p className="eyebrow">{t($ => $.pricing.ownerTitle)}</p>
            <h5>{loadedScope ? `${loadedScope.providerId} / ${loadedScope.profileId} / ${loadedScope.modelId}` : t($ => $.pricing.loadScope)}</h5>
          </div>
          <span className={`admin-control-status is-${loadState === "ready" ? "ready" : loadState === "failed" ? "blocked" : "neutral"}`}>
            {t($ => $.pricing.states[loadState])}
          </span>
        </header>

        {config.mode === "offline" ? (
          <PricingNotice title={t($ => $.pricing.offlineTitle)}>{t($ => $.pricing.offlineHelp)}</PricingNotice>
        ) : loadState === "idle" || loadState === "loading" ? (
          <PricingNotice title={loadState === "loading" ? t($ => $.pricing.readingRevision) : t($ => $.pricing.chooseScope)}>{t($ => $.pricing.loadHelp)}</PricingNotice>
        ) : null}

        {envelope?.policy && loadState === "ready" ? <PricingPolicyEvidence policy={envelope.policy} /> : null}
        {loadState === "missing" ? (
          <PricingNotice title={t($ => $.pricing.missingTitle)}>{t($ => $.pricing.missingHelp)}</PricingNotice>
        ) : null}
        {loadState === "failed" ? (
          <div className="admin-gateway-pricing-failure" role="alert">
            <strong>{envelope?.failureCode ?? "gateway_pricing_store_unavailable"}</strong>
            <p>{operationMessage || t($ => $.pricing.failedClosed)}</p>
            {reloadRequired ? <button type="button" className="secondary-action" onClick={() => void loadPricingOwner()}>{t($ => $.pricing.reload)}</button> : null}
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

        {scopeDirty ? <p className="admin-gateway-pricing-failure" role="alert">{t($ => $.pricing.scopeChanged)}</p> : null}
        <p className="admin-gateway-pricing-boundary"><span aria-hidden="true">!</span>{t($ => $.pricing.boundary)}</p>
      </section>
    </div>
  );
}

function PricingPolicyEvidence({ policy }: { policy: NonNullable<AdminGatewayModelPricingEnvelope["policy"]> }) {
  const { t } = useTranslation("admin");
  const { locale } = useLocalePreference();
  return (
    <div className="admin-gateway-pricing-evidence">
      <div className="admin-gateway-pricing-rates">
        <article><span>{t($ => $.pricing.input)}</span><strong>{adminDisplayRate(policy.inputPriceMicrosPerTokenUnit, locale)}</strong><small>{t($ => $.pricing.unitValue)}</small></article>
        <article><span>{t($ => $.pricing.output)}</span><strong>{adminDisplayRate(policy.outputPriceMicrosPerTokenUnit, locale)}</strong><small>{t($ => $.pricing.unitValue)}</small></article>
      </div>
      <dl className="admin-control-owner-meta">
        <div><dt>{t($ => $.pricing.policyRevision)}</dt><dd>{policy.policyId} · v{policy.recordVersion}</dd></div>
        <div><dt>{t($ => $.pricing.digest)}</dt><dd>{shortDigest(policy.policyDigest)}</dd></div>
        <div><dt>{t($ => $.pricing.updated)}</dt><dd><time dateTime={policy.updatedAt} title={policy.updatedAt}>{adminDisplayDate(policy.updatedAt, locale)}</time></dd></div>
        <div><dt>{t($ => $.pricing.updatedBy)}</dt><dd>{policy.updatedByActorRef}</dd></div>
        <div><dt>{t($ => $.pricing.reason)}</dt><dd>{policy.reason}</dd></div>
        <div><dt>{t($ => $.pricing.requestAudit)}</dt><dd>{policy.requestId} / {policy.auditRef}</dd></div>
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
  const { t } = useTranslation("admin");
  const { locale } = useLocalePreference();
  return (
    <div className="admin-gateway-pricing-update">
      <div className="admin-gateway-pricing-editor">
        <div className="card-title-row"><div><p className="eyebrow">{t($ => $.pricing.proposedRevision)}</p><h5>{t($ => $.pricing.expectedVersion, { version: currentVersion })}</h5></div><span>CAS</span></div>
        <div className="admin-gateway-pricing-rate-fields">
          <label>{t($ => $.pricing.inputRate)}<input inputMode="numeric" value={inputRate} onChange={(event) => onInputRate(event.target.value)} /></label>
          <label>{t($ => $.pricing.outputRate)}<input inputMode="numeric" value={outputRate} onChange={(event) => onOutputRate(event.target.value)} /></label>
        </div>
        <label>{t($ => $.pricing.sanitizedReason)}<textarea rows={3} value={reason} onChange={(event) => onReason(event.target.value)} /></label>
        <button type="button" className="secondary-action" onClick={onReview} disabled={pending}>{t($ => $.pricing.reviewRevision)}</button>
        {message ? <p className="admin-gateway-pricing-message" role="status">{message}</p> : null}
      </div>
      {review ? (
        <aside className="admin-gateway-pricing-confirm" aria-label={t($ => $.pricing.confirmationLabel)}>
          <p className="eyebrow">{t($ => $.pricing.explicitConfirmation)}</p>
          <h5>v{review.expectedVersion} → v{review.expectedVersion + 1}</h5>
          <dl>
            <div><dt>{t($ => $.pricing.input)}</dt><dd>{adminDisplayRate(review.inputRate, locale)}</dd></div>
            <div><dt>{t($ => $.pricing.output)}</dt><dd>{adminDisplayRate(review.outputRate, locale)}</dd></div>
            <div><dt>{t($ => $.pricing.unit)}</dt><dd>{t($ => $.pricing.unitValue)}</dd></div>
            <div><dt>{t($ => $.pricing.history)}</dt><dd>{t($ => $.pricing.notRecalculated)}</dd></div>
          </dl>
          <p>{review.reason}</p>
          <div><button type="button" className="secondary-action" onClick={onCancel} disabled={pending}>{t($ => $.pricing.cancel)}</button><button type="button" onClick={onConfirm} disabled={pending}>{pending ? t($ => $.pricing.writing) : t($ => $.pricing.confirm)}</button></div>
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

function shortDigest(value: string): string {
  return value.length > 24 ? `${value.slice(0, 16)}…${value.slice(-8)}` : value;
}

import { useTranslation } from "react-i18next";
import { useLocalePreference } from "../../i18n/LocaleProvider.tsx";
import { formatDisplayNumber } from "../../i18n/formatters.ts";
import { adminDisplayDate } from "./adminManagementFormatters.ts";
import { quotaNoticeMessage, quotaFailureCopy, type QuotaNotice } from "./adminGatewayManagementMessages.ts";
import "../../i18n/adminQuotaResources.ts";
import { useEffect, useMemo, useRef, useState } from "react";

import {
  isValidAdminGatewayRequestQuotaLimit,
  putAdminGatewayRequestQuota,
  readAdminGatewayRequestQuota,
  readAdminGatewayRequestQuotaConfig,
  type AdminGatewayRequestQuotaConfig,
  type AdminGatewayRequestQuotaEnvelope,
  type AdminGatewayRequestQuotaFailureCode,
} from "./adminGatewayRequestQuotaConsumer.ts";
import type { WorkspaceApplicationRow } from "./workspaceApplications.ts";

type LoadState = "loading" | "ready" | "missing" | "failed";

export function AdminGatewayRequestQuotaPanel({
  tenantRef,
  workspaceId,
  selectedApplicationId,
  selectedApplicationDisplayName,
  applications,
  onSelectApplication,
}: {
  tenantRef: string;
  workspaceId: string;
  selectedApplicationId: string;
  selectedApplicationDisplayName: string;
  applications: WorkspaceApplicationRow[];
  onSelectApplication: (applicationId: string) => void;
}) {
  const { t } = useTranslation("admin");
  const config = useMemo(
    () => readAdminGatewayRequestQuotaConfig({ tenantRef, workspaceId, applicationId: selectedApplicationId }),
    [selectedApplicationId, tenantRef, workspaceId],
  );
  const [loadState, setLoadState] = useState<LoadState>("loading");
  const [envelope, setEnvelope] = useState<AdminGatewayRequestQuotaEnvelope | null>(null);
  const [failureCode, setFailureCode] = useState<AdminGatewayRequestQuotaFailureCode | null>(null);
  const [requestLimitInput, setRequestLimitInput] = useState("");
  const [confirmationLimit, setConfirmationLimit] = useState<number | null>(null);
  const [operationPending, setOperationPending] = useState(false);
  const [operationNotice, setOperationNotice] = useState<QuotaNotice>(null);
  const [reloadRequired, setReloadRequired] = useState(false);
  const requestGenerationRef = useRef(0);

  useEffect(() => {
    const generation = ++requestGenerationRef.current;
    setEnvelope(null);
    setFailureCode(null);
    setRequestLimitInput("");
    setConfirmationLimit(null);
    setOperationPending(false);
    setOperationNotice(null);
    setReloadRequired(false);
    void loadQuotaOwner(config, generation);
    return () => {
      if (requestGenerationRef.current === generation) requestGenerationRef.current += 1;
    };
  }, [config]);

  const applicationRows = useMemo(() => {
    if (applications.some((application) => application.applicationRef === selectedApplicationId)) {
      return applications;
    }
    if (!selectedApplicationId) return applications;
    return [{
      applicationRef: selectedApplicationId,
      displayName: selectedApplicationDisplayName || selectedApplicationId,
      applicationKind: "unavailable",
      ownerSubjectRef: "",
      latestWorkflowDefinitionRef: "",
      lastRunStatus: "not_available",
      updatedAt: "",
      workspaceId,
    }, ...applications];
  }, [applications, selectedApplicationDisplayName, selectedApplicationId, workspaceId]);

  async function loadQuotaOwner(nextConfig = config, generation = ++requestGenerationRef.current) {
    setLoadState("loading");
    setFailureCode(null);
    setOperationNotice(null);
    try {
      const nextEnvelope = await readAdminGatewayRequestQuota(nextConfig);
      if (requestGenerationRef.current !== generation) return;
      setEnvelope(nextEnvelope);
      setFailureCode(nextEnvelope.failureCode);
      setConfirmationLimit(null);
      setReloadRequired(false);
      if (nextEnvelope.policy && nextEnvelope.usage) {
        setLoadState("ready");
        setRequestLimitInput(String(nextEnvelope.policy.requestLimit));
      } else if (nextEnvelope.failureCode === "gateway_quota_policy_not_found") {
        setLoadState("missing");
        setRequestLimitInput("");
      } else {
        setLoadState("failed");
      }
    } catch {
      if (requestGenerationRef.current !== generation) return;
      setEnvelope(null);
      setFailureCode(null);
      setLoadState("failed");
      setOperationNotice({ key: "invalidRead" });
    }
  }

  function reviewUpdate() {
    const requestLimit = Number(requestLimitInput);
    if (!isValidAdminGatewayRequestQuotaLimit(requestLimit)) {
      setOperationNotice({ key: "invalidLimit" });
      return;
    }
    if (envelope?.policy && requestLimit === envelope.policy.requestLimit) {
      setOperationNotice({ key: "unchangedLimit" });
      return;
    }
    setOperationNotice(null);
    setConfirmationLimit(requestLimit);
  }

  async function confirmUpdate() {
    if (confirmationLimit === null || !isValidAdminGatewayRequestQuotaLimit(confirmationLimit)) return;
    const expectedVersion = envelope?.policy?.recordVersion ?? 0;
    setOperationPending(true);
    setOperationNotice(null);
    try {
      const nextEnvelope = await putAdminGatewayRequestQuota(config, expectedVersion, confirmationLimit);
      if (nextEnvelope.failureCode) {
        setFailureCode(nextEnvelope.failureCode);
        setConfirmationLimit(null);
        if (nextEnvelope.failureCode === "gateway_quota_policy_version_conflict") {
          setReloadRequired(true);
          setOperationNotice({ key: "conflict" });
        } else {
          setOperationNotice({ key: "failure", code: nextEnvelope.failureCode });
        }
        return;
      }
      setEnvelope(nextEnvelope);
      setFailureCode(null);
      setLoadState("ready");
      setRequestLimitInput(String(nextEnvelope.policy?.requestLimit ?? confirmationLimit));
      setConfirmationLimit(null);
      setReloadRequired(false);
      setOperationNotice({ key: "updated", version: nextEnvelope.policy?.recordVersion ?? null });
    } catch {
      setOperationNotice({ key: "invalidUpdate" });
    } finally {
      setOperationPending(false);
    }
  }

  const operationMessage = quotaNoticeMessage(t, operationNotice);
  const failure = failureCode ? quotaFailureCopy(t, failureCode) : null;
  const policy = envelope?.policy ?? null;
  const usage = envelope?.usage ?? null;
  const selectedStatus = loadState === "ready" && usage?.remainingRequestCount === 0
    ? t($ => $.quota.states.limitReached)
    : loadState === "ready"
    ? t($ => $.quota.states.ready)
    : loadState === "missing"
    ? t($ => $.quota.states.missing)
    : loadState === "loading"
    ? t($ => $.quota.states.loading)
    : failure?.shortLabel ?? t($ => $.quota.states.failed);
  const selectedApplication = applicationRows.find(
    (application) => application.applicationRef === selectedApplicationId,
  );
  const selectedApplicationName = selectedApplication?.displayName || selectedApplicationDisplayName ||
    selectedApplicationId || t($ => $.quota.selectApplication);

  return (
    <div className="admin-gateway-quota-workspace" data-load-state={loadState}>
      <aside className="admin-gateway-quota-applications" aria-label={t($ => $.quota.policiesLabel)}>
        <header>
          <span>{t($ => $.quota.policies)}</span>
          <strong>{selectedApplicationId ? t($ => $.quota.oneDetail) : t($ => $.quota.selectionRequired)}</strong>
        </header>
        <div role="listbox" aria-label={t($ => $.quota.applicationList)}>
          {applicationRows.length ? applicationRows.map((application) => {
            const selected = application.applicationRef === selectedApplicationId;
            return (
              <button
                key={application.applicationRef}
                type="button"
                role="option"
                aria-selected={selected}
                className={`admin-gateway-quota-application ${selected ? "is-selected" : ""}`}
                onClick={() => onSelectApplication(application.applicationRef)}
              >
                <i aria-hidden="true" />
                <span>
                  <strong>{application.displayName}</strong>
                  <small>{application.applicationRef}</small>
                  <small>{application.lifecycleState ? t($ => $.quota.states[application.lifecycleState!]) : t($ => $.quota.devTest)}</small>
                </span>
                <em className={selected && usage?.remainingRequestCount === 0 ? "attention" : "neutral"}>
                  {selected ? selectedStatus : t($ => $.quota.notLoaded)}
                </em>
              </button>
            );
          }) : (
            <p className="admin-gateway-quota-empty">{t($ => $.quota.noApplications)}</p>
          )}
        </div>
        <p className="admin-gateway-quota-selection-note">
          {t($ => $.quota.selectionNote)}</p>
      </aside>

      <section className="admin-gateway-quota-detail" aria-labelledby="admin-gateway-quota-detail-title">
        <header className="admin-gateway-quota-context">
          <div>
            <span>{t($ => $.quota.selectedApplication)}</span>
            <strong>{selectedApplicationName}</strong>
          </div>
          <div>
            <code>{selectedApplicationId || "selection_required"}</code>
            <small>{t($ => $.quota.authority)}</small>
            <QuotaStatus
              tone={usage?.remainingRequestCount === 0 ? "attention" : loadState === "ready" ? "ready" : "blocked"}
            >
              {selectedStatus}
            </QuotaStatus>
          </div>
        </header>

        <section className="admin-gateway-quota-owner" aria-live="polite">
          <header>
            <div>
              <p className="eyebrow">{t($ => $.quota.eyebrow)}</p>
              <h5 id="admin-gateway-quota-detail-title">{t($ => $.quota.title)}</h5>
              <p>{t($ => $.quota.usageSource)}</p>
            </div>
            <span className="admin-gateway-quota-version">
              {policy ? t($ => $.quota.version, { version: policy.recordVersion }) : t($ => $.quota.states[loadState])}
            </span>
          </header>

          <dl className="admin-gateway-quota-scope">
            <div><dt>{t($ => $.quota.tenant)}</dt><dd>{config.tenantRef}</dd></div>
            <div><dt>{t($ => $.quota.workspace)}</dt><dd>{config.workspaceId}</dd></div>
            <div><dt>{t($ => $.quota.environment)}</dt><dd>{config.environment}</dd></div>
            <div><dt>{t($ => $.quota.application)}</dt><dd>{config.applicationId || t($ => $.quota.selectRequired)}</dd></div>
          </dl>

          {loadState === "loading" ? <QuotaLoading /> : null}
          {loadState === "ready" && policy && usage ? (
            <QuotaReady
              policy={policy}
              usage={usage}
              requestLimitInput={requestLimitInput}
              confirmationLimit={confirmationLimit}
              operationPending={operationPending}
              operationMessage={operationMessage}
              reloadRequired={reloadRequired}
              onRequestLimitInput={setRequestLimitInput}
              onReviewUpdate={reviewUpdate}
              onCancelConfirmation={() => setConfirmationLimit(null)}
              onConfirmUpdate={() => void confirmUpdate()}
              onReload={() => void loadQuotaOwner()}
            />
          ) : null}
          {loadState === "missing" ? (
            <QuotaMissingPolicy
              requestLimitInput={requestLimitInput}
              confirmationLimit={confirmationLimit}
              operationPending={operationPending}
              operationMessage={operationMessage}
              onRequestLimitInput={setRequestLimitInput}
              onReviewUpdate={reviewUpdate}
              onCancelConfirmation={() => setConfirmationLimit(null)}
              onConfirmUpdate={() => void confirmUpdate()}
            />
          ) : null}
          {loadState === "failed" ? (
            <QuotaFailure
              config={config}
              presentation={failure}
              message={operationMessage}
              onRetry={() => void loadQuotaOwner()}
            />
          ) : null}
        </section>

        <p className="admin-gateway-quota-boundary">
          <span aria-hidden="true">!</span>
          {t($ => $.quota.boundary)}</p>
      </section>
    </div>
  );
}

function QuotaReady({
  policy,
  usage,
  requestLimitInput,
  confirmationLimit,
  operationPending,
  operationMessage,
  reloadRequired,
  onRequestLimitInput,
  onReviewUpdate,
  onCancelConfirmation,
  onConfirmUpdate,
  onReload,
}: {
  policy: NonNullable<AdminGatewayRequestQuotaEnvelope["policy"]>;
  usage: NonNullable<AdminGatewayRequestQuotaEnvelope["usage"]>;
  requestLimitInput: string;
  confirmationLimit: number | null;
  operationPending: boolean;
  operationMessage: string;
  reloadRequired: boolean;
  onRequestLimitInput: (value: string) => void;
  onReviewUpdate: () => void;
  onCancelConfirmation: () => void;
  onConfirmUpdate: () => void;
  onReload: () => void;
}) {
  const { t } = useTranslation("admin");
  const { locale } = useLocalePreference();
  const fraction = Math.min(100, Math.max(0, (usage.admittedRequestCount / policy.requestLimit) * 100));
  return (
    <div className="admin-gateway-quota-owner-state">
      <div className={`admin-gateway-quota-usage ${usage.remainingRequestCount === 0 ? "is-exceeded" : ""}`}>
        <div>
          <span title={usage.periodStart}>{t($ => $.quota.window, { periodStart: adminDisplayDate(usage.periodStart, locale) })}</span>
          <strong>{t($ => $.quota.usage, { admitted: formatDisplayNumber(usage.admittedRequestCount, locale) ?? t($ => $.quota.states.unknown), limit: formatDisplayNumber(policy.requestLimit, locale) ?? t($ => $.quota.states.unknown) })}</strong>
          <small>{t($ => $.quota.admittedAttempts)}</small>
          <div className="admin-gateway-quota-progress" aria-label={t($ => $.quota.progress, { percent: fraction.toFixed(0) })}>
            <i style={{ width: `${fraction}%` }} />
          </div>
        </div>
        <aside>
          <strong>{formatDisplayNumber(usage.remainingRequestCount, locale)}</strong>
          <span>{t($ => $.quota.remaining)}</span>
          {usage.remainingRequestCount === 0 ? <small>gateway_quota_exceeded</small> : null}
        </aside>
      </div>
      <dl className="admin-gateway-quota-policy-meta">
        <div><dt>{t($ => $.quota.policy)}</dt><dd>{policy.policyId}</dd></div>
        <div><dt>{t($ => $.quota.period)}</dt><dd>{policy.period}</dd></div>
        <div><dt>{t($ => $.quota.recordVersion)}</dt><dd>{policy.recordVersion}</dd></div>
        <div><dt>{t($ => $.quota.updatedBy)}</dt><dd>{policy.updatedBy}</dd></div>
      </dl>
      <QuotaUpdateEditor
        currentLimit={policy.requestLimit}
        expectedVersion={policy.recordVersion}
        requestLimitInput={requestLimitInput}
        confirmationLimit={confirmationLimit}
        operationPending={operationPending}
        operationMessage={operationMessage}
        reloadRequired={reloadRequired}
        onRequestLimitInput={onRequestLimitInput}
        onReviewUpdate={onReviewUpdate}
        onCancelConfirmation={onCancelConfirmation}
        onConfirmUpdate={onConfirmUpdate}
        onReload={onReload}
      />
    </div>
  );
}

function QuotaMissingPolicy({
  requestLimitInput,
  confirmationLimit,
  operationPending,
  operationMessage,
  onRequestLimitInput,
  onReviewUpdate,
  onCancelConfirmation,
  onConfirmUpdate,
}: {
  requestLimitInput: string;
  confirmationLimit: number | null;
  operationPending: boolean;
  operationMessage: string;
  onRequestLimitInput: (value: string) => void;
  onReviewUpdate: () => void;
  onCancelConfirmation: () => void;
  onConfirmUpdate: () => void;
}) {
  const { t } = useTranslation("admin");
  return (
    <div className="admin-gateway-quota-owner-state">
      <div className="admin-gateway-quota-missing">
        <span aria-hidden="true">∅</span>
        <div>
          <strong>{t($ => $.quota.missingTitle)}</strong>
          <p>
            {t($ => $.quota.missingHelp)}
          </p>
        </div>
      </div>
      <QuotaUpdateEditor
        currentLimit={null}
        expectedVersion={0}
        requestLimitInput={requestLimitInput}
        confirmationLimit={confirmationLimit}
        operationPending={operationPending}
        operationMessage={operationMessage}
        reloadRequired={false}
        onRequestLimitInput={onRequestLimitInput}
        onReviewUpdate={onReviewUpdate}
        onCancelConfirmation={onCancelConfirmation}
        onConfirmUpdate={onConfirmUpdate}
        onReload={() => undefined}
      />
    </div>
  );
}

function QuotaUpdateEditor({
  currentLimit,
  expectedVersion,
  requestLimitInput,
  confirmationLimit,
  operationPending,
  operationMessage,
  reloadRequired,
  onRequestLimitInput,
  onReviewUpdate,
  onCancelConfirmation,
  onConfirmUpdate,
  onReload,
}: {
  currentLimit: number | null;
  expectedVersion: number;
  requestLimitInput: string;
  confirmationLimit: number | null;
  operationPending: boolean;
  operationMessage: string;
  reloadRequired: boolean;
  onRequestLimitInput: (value: string) => void;
  onReviewUpdate: () => void;
  onCancelConfirmation: () => void;
  onConfirmUpdate: () => void;
  onReload: () => void;
}) {
  const { t } = useTranslation("admin");
  const { locale } = useLocalePreference();
  return (
    <div className="admin-gateway-quota-update-layout">
      <form
        className="admin-gateway-quota-editor"
        onSubmit={(event) => {
          event.preventDefault();
          onReviewUpdate();
        }}
      >
        <span>{t($ => $.quota.policyUpdate)}</span>
        <strong>{currentLimit === null ? t($ => $.quota.createLimit) : t($ => $.quota.changeLimit)}</strong>
        <small>{t($ => $.quota.limitHelp)}</small>
        <label>
          {t($ => $.quota.requestLimit)}<input
            type="number"
            min="1"
            max="1000000"
            step="1"
            inputMode="numeric"
            value={requestLimitInput}
            disabled={operationPending || reloadRequired}
            onChange={(event) => onRequestLimitInput(event.target.value)}
          />
        </label>
        {reloadRequired ? (
          <button type="button" className="secondary-action" onClick={onReload}>{t($ => $.quota.reload)}</button>
        ) : (
          <button type="submit" className="primary-action" disabled={operationPending}>{t($ => $.quota.reviewUpdate)}</button>
        )}
      </form>
      {confirmationLimit !== null ? (
        <section className="admin-gateway-quota-confirmation" aria-label={t($ => $.quota.confirmationLabel)}>
          <span>{t($ => $.quota.confirmationVersion, { version: expectedVersion })}</span>
          <strong>{currentLimit === null ? t($ => $.quota.createRequests, { limit: formatDisplayNumber(confirmationLimit, locale) ?? t($ => $.quota.states.unknown) }) : t($ => $.quota.updateRequests, { current: formatDisplayNumber(currentLimit, locale) ?? t($ => $.quota.states.unknown), limit: formatDisplayNumber(confirmationLimit, locale) ?? t($ => $.quota.states.unknown) })}</strong>
          <p>
            {t($ => $.quota.confirmationHelp)}</p>
          <small>{t($ => $.quota.staleHelp)}</small>
          <div>
            <button type="button" className="secondary-action" disabled={operationPending} onClick={onCancelConfirmation}>
              {t($ => $.quota.cancel)}</button>
            <button type="button" className="primary-action" disabled={operationPending} onClick={onConfirmUpdate}>
              {operationPending ? t($ => $.quota.updating) : t($ => $.quota.confirmUpdate)}
            </button>
          </div>
        </section>
      ) : (
        <section className="admin-gateway-quota-cas-summary" aria-label={t($ => $.quota.casLabel)}>
          <span>{t($ => $.quota.casGuard)}</span>
          <strong>{t($ => $.quota.expectedVersion, { version: expectedVersion })}</strong>
          <p>{t($ => $.quota.reviewHelp)}</p>
        </section>
      )}
      {operationMessage ? <p className="admin-gateway-quota-operation" role="status">{operationMessage}</p> : null}
    </div>
  );
}

function QuotaLoading() {
  const { t } = useTranslation("admin");
  return (
    <div className="admin-gateway-quota-loading" aria-live="polite">
      <span>{t($ => $.quota.loading)}</span>
      <i /><i /><i />
      <small>{t($ => $.quota.loadingHelp)}</small>
    </div>
  );
}

function QuotaFailure({
  config,
  presentation,
  message,
  onRetry,
}: {
  config: AdminGatewayRequestQuotaConfig;
  presentation: ReturnType<typeof quotaFailureCopy> | null;
  message: string;
  onRetry: () => void;
}) {
  const { t } = useTranslation("admin");
  const title = presentation?.title ?? t($ => $.quota.unavailableTitle);
  const summary = message || presentation?.summary || t($ => $.quota.noAcceptedPolicy);
  return (
    <div className="admin-gateway-quota-failure" role="alert">
      <span aria-hidden="true">!</span>
      <div>
        <strong>{title}</strong>
        <p>{summary}</p>
        <small>{presentation?.code ?? "strict_response_validation_failed"}</small>
        {config.mode === "dev_admin_gateway_request_quota_http" ? (
          <button type="button" className="secondary-action" onClick={onRetry}>{t($ => $.quota.retry)}</button>
        ) : null}
      </div>
    </div>
  );
}

function QuotaStatus({
  tone,
  children,
}: {
  tone: "ready" | "blocked" | "attention";
  children: string;
}) {
  return <span className={`admin-gateway-quota-status is-${tone}`}>{children}</span>;
}

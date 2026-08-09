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
  const [operationMessage, setOperationMessage] = useState("");
  const [reloadRequired, setReloadRequired] = useState(false);
  const requestGenerationRef = useRef(0);

  useEffect(() => {
    const generation = ++requestGenerationRef.current;
    setEnvelope(null);
    setFailureCode(null);
    setRequestLimitInput("");
    setConfirmationLimit(null);
    setOperationPending(false);
    setOperationMessage("");
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
    setOperationMessage("");
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
      setOperationMessage("The quota response failed strict validation. No policy or usage was accepted.");
    }
  }

  function reviewUpdate() {
    const requestLimit = Number(requestLimitInput);
    if (!isValidAdminGatewayRequestQuotaLimit(requestLimit)) {
      setOperationMessage("Request limit must be a positive integer from 1 to 1,000,000.");
      return;
    }
    if (envelope?.policy && requestLimit === envelope.policy.requestLimit) {
      setOperationMessage("Enter a different request limit before reviewing an update.");
      return;
    }
    setOperationMessage("");
    setConfirmationLimit(requestLimit);
  }

  async function confirmUpdate() {
    if (confirmationLimit === null || !isValidAdminGatewayRequestQuotaLimit(confirmationLimit)) return;
    const expectedVersion = envelope?.policy?.recordVersion ?? 0;
    setOperationPending(true);
    setOperationMessage("");
    try {
      const nextEnvelope = await putAdminGatewayRequestQuota(config, expectedVersion, confirmationLimit);
      if (nextEnvelope.failureCode) {
        setFailureCode(nextEnvelope.failureCode);
        setConfirmationLimit(null);
        if (nextEnvelope.failureCode === "gateway_quota_policy_version_conflict") {
          setReloadRequired(true);
          setOperationMessage("The expected version is stale. Reload the quota owner before reviewing another update.");
        } else {
          setOperationMessage(failurePresentation(nextEnvelope.failureCode).summary);
        }
        return;
      }
      setEnvelope(nextEnvelope);
      setFailureCode(null);
      setLoadState("ready");
      setRequestLimitInput(String(nextEnvelope.policy?.requestLimit ?? confirmationLimit));
      setConfirmationLimit(null);
      setReloadRequired(false);
      setOperationMessage(`Policy version ${nextEnvelope.policy?.recordVersion ?? "unknown"} is now current.`);
    } catch {
      setOperationMessage("The quota update response failed strict validation. The displayed policy is unchanged.");
    } finally {
      setOperationPending(false);
    }
  }

  const failure = failureCode ? failurePresentation(failureCode) : null;
  const policy = envelope?.policy ?? null;
  const usage = envelope?.usage ?? null;
  const selectedStatus = loadState === "ready" && usage?.remainingRequestCount === 0
    ? "limit reached"
    : loadState === "ready"
    ? "policy ready"
    : loadState === "missing"
    ? "policy missing"
    : loadState === "loading"
    ? "loading"
    : failure?.shortLabel ?? "blocked";

  return (
    <div className="admin-gateway-quota-workspace" data-load-state={loadState}>
      <aside className="admin-gateway-quota-applications" aria-label="Application quota policies">
        <header>
          <span>Application policies</span>
          <strong>{selectedApplicationId ? "One current detail" : "Selection required"}</strong>
        </header>
        <div role="listbox" aria-label="Applications in the active workspace">
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
                  <small>{application.lifecycleState ?? "development/test"}</small>
                </span>
                <em className={selected && usage?.remainingRequestCount === 0 ? "attention" : "neutral"}>
                  {selected ? selectedStatus : "not loaded"}
                </em>
              </button>
            );
          }) : (
            <p className="admin-gateway-quota-empty">No application owner exists in the active workspace.</p>
          )}
        </div>
        <p className="admin-gateway-quota-selection-note">
          Only the application driving this detail receives the ink-blue selection track. Policy state remains a
          separate text badge.
        </p>
      </aside>

      <section className="admin-gateway-quota-detail" aria-labelledby="admin-gateway-quota-detail-title">
        <header>
          <div>
            <p className="eyebrow">Quota · admin_gateway_quotas:read / write</p>
            <h5 id="admin-gateway-quota-detail-title">UTC daily provider-attempt policy</h5>
            <p>Usage comes from the quota owner, never from Request History or the legacy QuotaSummary.</p>
          </div>
          <QuotaStatus
            tone={usage?.remainingRequestCount === 0 ? "attention" : loadState === "ready" ? "ready" : "blocked"}
          >
            {selectedStatus}
          </QuotaStatus>
        </header>

        <dl className="admin-gateway-quota-scope">
          <div><dt>Tenant</dt><dd>{config.tenantRef}</dd></div>
          <div><dt>Workspace</dt><dd>{config.workspaceId}</dd></div>
          <div><dt>Environment</dt><dd>{config.environment}</dd></div>
          <div><dt>Application</dt><dd>{config.applicationId || "selection required"}</dd></div>
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

        <p className="admin-gateway-quota-boundary">
          <span aria-hidden="true">!</span>
          Production quota, token/cost limits, billing, delete/disable, automatic increase, automatic routing,
          formal membership and OIDC remain closed.
        </p>
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
  const fraction = Math.min(100, Math.max(0, (usage.admittedRequestCount / policy.requestLimit) * 100));
  return (
    <div className="admin-gateway-quota-owner-state">
      <div className={`admin-gateway-quota-usage ${usage.remainingRequestCount === 0 ? "is-exceeded" : ""}`}>
        <div>
          <span>UTC window · {usage.periodStart}</span>
          <strong>{usage.admittedRequestCount} / {policy.requestLimit}</strong>
          <small>admitted provider attempts</small>
          <div className="admin-gateway-quota-progress" aria-label={`${fraction.toFixed(0)}% of the request limit admitted`}>
            <i style={{ width: `${fraction}%` }} />
          </div>
        </div>
        <aside>
          <strong>{usage.remainingRequestCount}</strong>
          <span>remaining today</span>
          {usage.remainingRequestCount === 0 ? <small>gateway_quota_exceeded</small> : null}
        </aside>
      </div>
      <dl className="admin-gateway-quota-policy-meta">
        <div><dt>Policy</dt><dd>{policy.policyId}</dd></div>
        <div><dt>Period</dt><dd>{policy.period}</dd></div>
        <div><dt>Record version</dt><dd>{policy.recordVersion}</dd></div>
        <div><dt>Updated by</dt><dd>{policy.updatedBy}</dd></div>
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
  return (
    <div className="admin-gateway-quota-owner-state">
      <div className="admin-gateway-quota-missing">
        <span aria-hidden="true">∅</span>
        <div>
          <strong>No quota policy exists for this exact application scope</strong>
          <p>
            Create a positive UTC daily request limit with <code>expected_version = 0</code>. Missing policy remains
            fail-closed and is not replaced by Request History or the legacy QuotaSummary.
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
  return (
    <div className="admin-gateway-quota-update-layout">
      <form
        className="admin-gateway-quota-editor"
        onSubmit={(event) => {
          event.preventDefault();
          onReviewUpdate();
        }}
      >
        <span>Policy update</span>
        <strong>{currentLimit === null ? "Create UTC daily request limit" : "Change UTC daily request limit"}</strong>
        <small>Positive integers only · 1–1,000,000 · no delete or disable</small>
        <label>
          Request limit
          <input
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
          <button type="button" className="secondary-action" onClick={onReload}>Reload current policy</button>
        ) : (
          <button type="submit" className="primary-action" disabled={operationPending}>Review update</button>
        )}
      </form>
      {confirmationLimit !== null ? (
        <section className="admin-gateway-quota-confirmation" aria-label="Quota policy update confirmation">
          <span>Confirmation · expected version {expectedVersion}</span>
          <strong>{currentLimit === null ? "Create" : `Update ${currentLimit} →`} {confirmationLimit} requests</strong>
          <p>
            The exact tenant, workspace, environment and application scope changes. The current admitted count is
            never reset or recalculated.
          </p>
          <small>If the expected version is stale, the write is rejected and this owner must be reloaded.</small>
          <div>
            <button type="button" className="secondary-action" disabled={operationPending} onClick={onCancelConfirmation}>
              Cancel
            </button>
            <button type="button" className="primary-action" disabled={operationPending} onClick={onConfirmUpdate}>
              {operationPending ? "Updating…" : "Confirm update"}
            </button>
          </div>
        </section>
      ) : (
        <section className="admin-gateway-quota-cas-summary" aria-label="Quota policy CAS boundary">
          <span>CAS guard</span>
          <strong>Expected version {expectedVersion}</strong>
          <p>Review opens an explicit confirmation. No update is sent from this editor directly.</p>
        </section>
      )}
      {operationMessage ? <p className="admin-gateway-quota-operation" role="status">{operationMessage}</p> : null}
    </div>
  );
}

function QuotaLoading() {
  return (
    <div className="admin-gateway-quota-loading" aria-live="polite">
      <span>Loading the exact quota owner…</span>
      <i /><i /><i />
      <small>Policy update remains disabled while the owner is unresolved.</small>
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
  presentation: ReturnType<typeof failurePresentation> | null;
  message: string;
  onRetry: () => void;
}) {
  const title = presentation?.title ?? "Quota response unavailable";
  const summary = message || presentation?.summary || "No policy or usage was accepted.";
  return (
    <div className="admin-gateway-quota-failure" role="alert">
      <span aria-hidden="true">!</span>
      <div>
        <strong>{title}</strong>
        <p>{summary}</p>
        <small>{presentation?.code ?? "strict_response_validation_failed"}</small>
        {config.mode === "dev_admin_gateway_request_quota_http" ? (
          <button type="button" className="secondary-action" onClick={onRetry}>Retry exact owner</button>
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

function failurePresentation(code: AdminGatewayRequestQuotaFailureCode) {
  const presentations: Record<AdminGatewayRequestQuotaFailureCode, {
    code: AdminGatewayRequestQuotaFailureCode;
    shortLabel: string;
    title: string;
    summary: string;
  }> = {
    gateway_quota_disabled: {
      code,
      shortLabel: "HTTP closed",
      title: "Development/test quota management is closed",
      summary: "The Admin read or write gate is disabled. No fallback policy is available.",
    },
    gateway_quota_scope_denied: {
      code,
      shortLabel: "permission blocked",
      title: "Quota permission is denied",
      summary: "The exact admin_gateway_quotas read or write permission is required for this workspace.",
    },
    gateway_quota_environment_forbidden: {
      code,
      shortLabel: "environment blocked",
      title: "Environment is outside the development/test boundary",
      summary: "Only development or test is accepted. Production quota remains closed.",
    },
    gateway_quota_payload_invalid: {
      code,
      shortLabel: "payload rejected",
      title: "Quota update payload was rejected",
      summary: "The owner accepted neither the policy nor the usage response. Review the positive integer limit.",
    },
    gateway_quota_policy_not_found: {
      code,
      shortLabel: "policy missing",
      title: "Quota policy is missing",
      summary: "Create the first policy with expected version 0. Missing policy remains fail-closed.",
    },
    gateway_quota_policy_version_conflict: {
      code,
      shortLabel: "version conflict",
      title: "Quota policy version changed",
      summary: "The stale write was rejected. Reload the exact owner before reviewing another update.",
    },
    gateway_quota_attempt_conflict: {
      code,
      shortLabel: "attempt conflict",
      title: "Quota admission attempt conflicts",
      summary: "The provider attempt did not proceed. Reload the quota owner before another management action.",
    },
    gateway_quota_exceeded: {
      code,
      shortLabel: "limit reached",
      title: "UTC daily request limit is reached",
      summary: "New provider attempts are blocked. The status does not change the selected application.",
    },
    gateway_quota_store_unavailable: {
      code,
      shortLabel: "store unavailable",
      title: "Quota store is unavailable",
      summary: "The exact quota owner cannot be read. The application is not treated as unlimited.",
    },
  };
  return presentations[code];
}

import "../../i18n/apiKeyResources.ts";
import { useTranslation } from "react-i18next";
import { useEffect, useRef, useState, type FormEvent } from "react";

import {
  issueAPIKey,
  listAPIKeyRecords,
  readAPIKeyLifecycleConfig,
  readAPIKeyRecord,
  replaceAPIKeyListRecord,
  revokeAPIKey,
  type APIKeyEffectiveState,
  type APIKeyListResult,
  type APIKeyRecord,
  type APIKeyScope,
} from "./apiKeyLifecycleConsumer.ts";
import {
  APIKeyRotationSessionError,
  beginAPIKeyRotationSession,
  cancelAPIKeyRotationSession,
  completeAPIKeyRotationSession,
  currentAPIKeyRotationSourceVersion,
  readAPIKeyRotationSession,
  recordAPIKeyRotationReplacement,
  refreshAPIKeyRotationVerification,
  synchronizeAPIKeyRotationApplication,
  type APIKeyRotationSession,
} from "./apiKeyRotationSession.ts";
import { requestAPIKeyModelGatewayPlaygroundHandoff } from "./modelGatewayPlaygroundEvents.ts";
import { readModelGatewayPlaygroundConfig } from "./modelGatewayPlaygroundConsumer.ts";
import { requestWorkflowRAGApplicationCredentialHandoff } from "./workflowRAGApplicationRuntimeEvents.ts";
import { requestPromptApplicationCredentialHandoff } from "./promptApplicationInvocationEvents.ts";
import type {
  WorkspaceApiKeyRow,
  WorkspaceApiKeysMetric,
  WorkspaceApiKeysStatePreview,
  WorkspaceApiKeysViewModel,
} from "./workspaceApiKeys.ts";

const config = readAPIKeyLifecycleConfig();
const playgroundConfig = readModelGatewayPlaygroundConfig();
const AVAILABLE_SCOPES: Array<{ scope: APIKeyScope; label: string }> = [
  { scope: "models:read", label: "Read model catalog" },
  { scope: "chat:invoke", label: "Chat Completions" },
  { scope: "responses:invoke", label: "Responses" },
  { scope: "messages:invoke", label: "Messages" },
  { scope: "application_rag:invoke", label: "Application RAG invocation" },
  { scope: "prompt_application:invoke", label: "Prompt Application invocation" },
  { scope: "agent_copilot:invoke", label: "Agent Copilot invocation" },
];

type IssuedCredential = {
  apiKeyId: string;
  token: string;
};

type OperationNotice = {
  tone: "neutral" | "good" | "bad";
  kind: "none" | "failure" | "issued" | "detailLoaded" | "revoked" | "confirmRevocation" |
    "rotationStarted" | "rotationSessionCleared" | "replacementIssued" | "replacementVerified" |
    "replacementNotAuthenticated" | "rotationCompleted";
  failureCode: string;
  apiKeyId?: string;
  timestamp?: string;
};

export function APIKeyLifecyclePanel({
  applicationId,
  applicationName,
  applicationActive,
  workspaceId,
  offlineView,
}: {
  applicationId: string;
  applicationName: string;
  applicationActive: boolean;
  workspaceId: string;
  offlineView: WorkspaceApiKeysViewModel;
}) {
  const { t } = useTranslation("gateway");
  const [list, setList] = useState<APIKeyListResult>(() => initialList());
  const [effectiveState, setEffectiveState] = useState<APIKeyEffectiveState | "">("");
  const [displayName, setDisplayName] = useState(defaultDisplayName(applicationName));
  const [expiresInDays, setExpiresInDays] = useState(30);
  const [scopes, setScopes] = useState<APIKeyScope[]>(AVAILABLE_SCOPES.map(({ scope }) => scope));
  const [selectedRecord, setSelectedRecord] = useState<APIKeyRecord | null>(null);
  const [issuedCredential, setIssuedCredential] = useState<IssuedCredential | null>(null);
  const [pendingRevokeId, setPendingRevokeId] = useState("");
  const [rotation, setRotation] = useState<APIKeyRotationSession | null>(() => (
    applicationActive ? readAPIKeyRotationSession(applicationId) : null
  ));
  const [rotationDisplayName, setRotationDisplayName] = useState("");
  const [rotationExpiresInDays, setRotationExpiresInDays] = useState(30);
  const [showRotationRetireConfirm, setShowRotationRetireConfirm] = useState(false);
  const [busy, setBusy] = useState<
    "" | "list" | "issue" | "read" | "revoke" | "rotation_issue" | "rotation_verify" | "rotation_retire"
  >("");
  const [copyStatus, setCopyStatus] = useState<"" | "copied" | "failed">("");
  const [notice, setNotice] = useState<OperationNotice>({ tone: "neutral", kind: "none", failureCode: "" });
  const operationGeneration = useRef(0);
  const issueDisclosure = useRef<HTMLDetailsElement>(null);
  const oneTimePanel = useRef<HTMLElement>(null);
  const workspaceScopeMatches = config.mode !== "dev_api_key_lifecycle_http" || config.workspaceId === workspaceId;

  useEffect(() => {
    operationGeneration.current += 1;
    setIssuedCredential(null);
    setSelectedRecord(null);
    setPendingRevokeId("");
    const currentRotation = synchronizeAPIKeyRotationApplication(applicationId);
    const visibleRotation = applicationActive ? currentRotation : null;
    setRotation(visibleRotation);
    setRotationDisplayName(visibleRotation ? replacementDisplayName(visibleRotation.sourceDisplayName) : "");
    setRotationExpiresInDays(30);
    setShowRotationRetireConfirm(false);
    setCopyStatus("");
    setNotice({ tone: "neutral", kind: "none", failureCode: "" });
    setDisplayName(defaultDisplayName(applicationName));
    setScopes(AVAILABLE_SCOPES.map(({ scope }) => scope));
    setExpiresInDays(30);
    setEffectiveState("");
    setList(initialList());
    if (config.mode === "dev_api_key_lifecycle_http" && applicationId && workspaceScopeMatches) {
      void loadRecords(false, "", "");
    }
  }, [applicationActive, applicationId, applicationName, workspaceScopeMatches]);

  useEffect(() => {
    function clearCredentialAfterRouteLeave() {
      if (window.location.hash !== "#workspace-api-keys") setIssuedCredential(null);
    }
    window.addEventListener("hashchange", clearCredentialAfterRouteLeave);
    return () => {
      operationGeneration.current += 1;
      window.removeEventListener("hashchange", clearCredentialAfterRouteLeave);
    };
  }, []);

  if (config.mode === "offline") return <OfflineAPIKeySummary view={offlineView} />;
  if (!workspaceScopeMatches) {
    return (
      <section className="surface-band workspace-api-keys api-key-lifecycle" id="workspace-api-keys" aria-labelledby="workspace-api-keys-title">
        <div className="section-heading">
          <div><p className="eyebrow">{t($ => $.apiKey.credentials)}</p><h3 id="workspace-api-keys-title">{t($ => $.apiKey.apiKeyLifecycle)}</h3></div>
          <span className="status-badge neutral">{t($ => $.apiKey.scopeMismatch)}</span>
        </div>
        <article className="api-key-lifecycle-empty" role="alert">
          <h4>{t($ => $.apiKey.requestsBlocked)}</h4>
          <p>{t($ => $.apiKey.activeWorkspacePrefix)}<code>{workspaceId}</code>{t($ => $.apiKey.configuredForPrefix)}<code>{config.workspaceId}</code>.</p>
        </article>
      </section>
    );
  }

  async function loadRecords(
    append: boolean,
    cursor = list.nextCursor,
    stateFilter: APIKeyEffectiveState | "" = effectiveState,
  ) {
    if (!applicationId) {
      setList(initialList());
      return;
    }
    const generation = ++operationGeneration.current;
    setBusy("list");
    const next = await listAPIKeyRecords(config, applicationId, append ? cursor : "", stateFilter);
    if (operationGeneration.current !== generation) return;
    setBusy("");
    setList(append && next.status === "ready" ? { ...next, records: [...list.records, ...next.records] } : next);
    if (next.status === "failed") setNotice({ tone: "bad", kind: "failure", failureCode: next.failureCode });
  }

  async function submitIssue(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!applicationId || !applicationActive) return;
    const generation = ++operationGeneration.current;
    setIssuedCredential(null);
    setCopyStatus("");
    setBusy("issue");
    const result = await issueAPIKey(config, { applicationId, displayName, scopes, expiresInDays });
    if (operationGeneration.current !== generation) return;
    setBusy("");
    if (result.status === "issued" && result.record && result.credentialToken) {
      setIssuedCredential({ apiKeyId: result.record.apiKeyId, token: result.credentialToken });
      setSelectedRecord(result.record);
      setNotice({ tone: "good", kind: "issued", failureCode: "" });
      await loadRecords(false, "");
      issueDisclosure.current?.removeAttribute("open");
      window.requestAnimationFrame(() => oneTimePanel.current?.scrollIntoView({ block: "center" }));
      return;
    }
    setNotice({ tone: "bad", kind: "failure", failureCode: result.failureCode });
  }

  async function loadDetail(apiKeyId: string) {
    const generation = ++operationGeneration.current;
    setBusy("read");
    const result = await readAPIKeyRecord(config, apiKeyId);
    if (operationGeneration.current !== generation) return;
    setBusy("");
    if (result.status === "loaded" && result.record) {
      setSelectedRecord(result.record);
      setNotice({ tone: "neutral", kind: "detailLoaded", failureCode: "" });
      return;
    }
    setNotice({ tone: "bad", kind: "failure", failureCode: result.failureCode });
  }

  async function confirmRevoke(record: APIKeyRecord) {
    if (pendingRevokeId !== record.apiKeyId) {
      setPendingRevokeId(record.apiKeyId);
      setNotice({ tone: "neutral", kind: "confirmRevocation", apiKeyId: record.apiKeyId, failureCode: "" });
      return;
    }
    const generation = ++operationGeneration.current;
    setBusy("revoke");
    const result = await revokeAPIKey(config, record.apiKeyId, record.recordVersion);
    if (operationGeneration.current !== generation) return;
    setBusy("");
    setPendingRevokeId("");
    if (result.status === "revoked" && result.record) {
      setSelectedRecord(result.record);
      setNotice({ tone: "good", kind: "revoked", failureCode: "" });
      await loadRecords(false, "");
      return;
    }
    setNotice({ tone: "bad", kind: "failure", failureCode: result.failureCode });
    if (result.status === "version_conflict") await loadRecords(false, "");
  }

  function startRotation(record: APIKeyRecord) {
    try {
      const next = beginAPIKeyRotationSession(record);
      setRotation(next);
      setRotationDisplayName(replacementDisplayName(record.displayName));
      setRotationExpiresInDays(30);
      setShowRotationRetireConfirm(false);
      setIssuedCredential(null);
      setCopyStatus("");
      setSelectedRecord(record);
      setNotice({
        tone: "neutral",
        kind: "rotationStarted",
        apiKeyId: record.apiKeyId,
        failureCode: "",
      });
    } catch (error) {
      setNotice(rotationFailureNotice(error));
    }
  }

  function cancelRotation() {
    cancelAPIKeyRotationSession(applicationId);
    setRotation(null);
    setRotationDisplayName("");
    setRotationExpiresInDays(30);
    setShowRotationRetireConfirm(false);
    setIssuedCredential(null);
    setCopyStatus("");
    setNotice({
      tone: "neutral",
      kind: "rotationSessionCleared",
      failureCode: "",
    });
  }

  async function issueRotationReplacement() {
    if (!rotation || rotation.phase !== "replacement_pending" || !applicationActive) return;
    const generation = ++operationGeneration.current;
    setIssuedCredential(null);
    setCopyStatus("");
    setBusy("rotation_issue");
    const result = await issueAPIKey(config, {
      applicationId,
      displayName: rotationDisplayName,
      scopes: rotation.scopes,
      expiresInDays: rotationExpiresInDays,
    });
    if (operationGeneration.current !== generation) return;
    setBusy("");
    if (result.status !== "issued" || !result.record || !result.credentialToken) {
      setNotice({ tone: "bad", kind: "failure", failureCode: result.failureCode });
      return;
    }
    try {
      const next = recordAPIKeyRotationReplacement(applicationId, result.record);
      setRotation(next);
      setIssuedCredential({ apiKeyId: result.record.apiKeyId, token: result.credentialToken });
      setSelectedRecord(result.record);
      setNotice({
        tone: "good",
        kind: "replacementIssued",
        failureCode: "",
      });
      await loadRecords(false, "");
    } catch (error) {
      setIssuedCredential(null);
      setSelectedRecord(result.record);
      setNotice(rotationFailureNotice(error));
      await loadRecords(false, "");
    }
  }

  async function checkRotationVerification() {
    if (!rotation?.replacementApiKeyId) return;
    const generation = ++operationGeneration.current;
    setBusy("rotation_verify");
    const result = await readAPIKeyRecord(config, rotation.replacementApiKeyId);
    if (operationGeneration.current !== generation) return;
    setBusy("");
    if (result.status !== "loaded" || !result.record) {
      setNotice({ tone: "bad", kind: "failure", failureCode: result.failureCode });
      return;
    }
    const replacementRecord = result.record;
    try {
      const next = refreshAPIKeyRotationVerification(applicationId, replacementRecord);
      setRotation(next);
      setSelectedRecord(replacementRecord);
      setList((current) => replaceAPIKeyListRecord(current, replacementRecord));
      setShowRotationRetireConfirm(false);
      setNotice(next.phase === "verified" && next.replacementLastUsedAt
        ? {
          tone: "good",
          kind: "replacementVerified",
          timestamp: next.replacementLastUsedAt,
          failureCode: "",
        }
        : next.phase === "verified" ? {
          tone: "bad",
          kind: "failure",
          failureCode: "api_key_rotation_state_invalid",
        }
        : {
          tone: "neutral",
          kind: "replacementNotAuthenticated",
          failureCode: "",
        });
    } catch (error) {
      setNotice(rotationFailureNotice(error));
    }
  }

  async function retireRotationSource() {
    if (!rotation || rotation.phase !== "verified" || !showRotationRetireConfirm) return;
    const generation = ++operationGeneration.current;
    setBusy("rotation_retire");
    const source = await readAPIKeyRecord(config, rotation.sourceApiKeyId);
    if (operationGeneration.current !== generation) return;
    if (source.status !== "loaded" || !source.record) {
      setBusy("");
      setNotice({ tone: "bad", kind: "failure", failureCode: source.failureCode });
      return;
    }
    let expectedVersion: number;
    try {
      expectedVersion = currentAPIKeyRotationSourceVersion(applicationId, source.record);
    } catch (error) {
      setBusy("");
      setNotice(rotationFailureNotice(error));
      return;
    }
    const revoked = await revokeAPIKey(config, source.record.apiKeyId, expectedVersion);
    if (operationGeneration.current !== generation) return;
    setBusy("");
    if (revoked.status !== "revoked" || !revoked.record) {
      setShowRotationRetireConfirm(false);
      setNotice({ tone: "bad", kind: "failure", failureCode: revoked.failureCode });
      if (revoked.status === "version_conflict") await loadRecords(false, "");
      return;
    }
    try {
      completeAPIKeyRotationSession(applicationId, revoked.record);
      setRotation(null);
      setRotationDisplayName("");
      setRotationExpiresInDays(30);
      setShowRotationRetireConfirm(false);
      setIssuedCredential(null);
      setCopyStatus("");
      setSelectedRecord(revoked.record);
      setNotice({
        tone: "good",
        kind: "rotationCompleted",
        failureCode: "",
      });
      await loadRecords(false, "");
    } catch (error) {
      setNotice(rotationFailureNotice(error));
    }
  }

  async function copyCredential() {
    if (!issuedCredential) return;
    try {
      await navigator.clipboard.writeText(issuedCredential.token);
      setCopyStatus("copied");
    } catch {
      setCopyStatus("failed");
    }
  }

  function handoffCredential() {
    if (!issuedCredential || !applicationId) return;
    requestAPIKeyModelGatewayPlaygroundHandoff(
      applicationId,
      issuedCredential.apiKeyId,
      issuedCredential.token,
      playgroundConfig.defaultModel,
    );
    setIssuedCredential(null);
    setCopyStatus("");
    window.location.hash = "model-gateway-playground";
  }

  function handoffRAGCredential() {
    if (!issuedCredential || !applicationId || !selectedRecord?.scopes.includes("application_rag:invoke")) return;
    requestWorkflowRAGApplicationCredentialHandoff(
      applicationId,
      issuedCredential.apiKeyId,
      issuedCredential.token,
    );
    setIssuedCredential(null);
    setCopyStatus("");
    window.location.hash = "application-rag-invocation";
  }

  function handoffPromptCredential() {
    if (!issuedCredential || !applicationId || !selectedRecord?.scopes.includes("prompt_application:invoke")) return;
    requestPromptApplicationCredentialHandoff(
      applicationId,
      issuedCredential.apiKeyId,
      issuedCredential.token,
    );
    setIssuedCredential(null);
    setCopyStatus("");
    window.location.hash = "prompt-application-invocation";
  }

  function toggleScope(scope: APIKeyScope) {
    setScopes((current) => current.includes(scope) ? current.filter((item) => item !== scope) : [...current, scope]);
  }

  const selectedApplicationAvailable = Boolean(applicationId);
  const noticeMessage = notice.kind === "failure" ? t($ => $.apiKey.operationUnavailable, { code: notice.failureCode || t($ => $.apiKey.unknownFailure) })
    : notice.kind === "issued" ? t($ => $.apiKey.issueSucceeded)
    : notice.kind === "detailLoaded" ? t($ => $.apiKey.detailLoaded)
    : notice.kind === "revoked" ? t($ => $.apiKey.revokeSucceeded)
    : notice.kind === "confirmRevocation" ? t($ => $.apiKey.confirmRevocationIrreversible, { apiKeyId: notice.apiKeyId ?? "" })
    : notice.kind === "rotationStarted" ? t($ => $.apiKey.rotationStarted, { apiKeyId: notice.apiKeyId ?? "" })
    : notice.kind === "rotationSessionCleared" ? t($ => $.apiKey.rotationSessionCleared)
    : notice.kind === "replacementIssued" ? t($ => $.apiKey.replacementIssuedNotice)
    : notice.kind === "replacementVerified" ? t($ => $.apiKey.replacementVerifiedNotice, { timestamp: notice.timestamp ?? "" })
    : notice.kind === "replacementNotAuthenticated" ? t($ => $.apiKey.replacementNotYetAuthenticated)
    : notice.kind === "rotationCompleted" ? t($ => $.apiKey.rotationCompleted) : "";
  const listMessage = list.status === "failed" ? t($ => $.apiKey.listUnavailable, { code: list.failureCode || t($ => $.apiKey.unknownFailure) })
    : list.status === "offline" ? t($ => $.apiKey.offlineList)
    : list.requestId ? list.records.length ? t($ => $.apiKey.loadedRecords, { count: list.records.length }) : t($ => $.apiKey.noKeys)
    : applicationId ? t($ => $.apiKey.loadingKeysForApplication, { applicationName }) : t($ => $.apiKey.selectApplicationToLoad);
  const scopeLabel = (scope: APIKeyScope) => {
    switch (scope) {
      case "models:read": return t($ => $.apiKey.scopeReadModels);
      case "chat:invoke": return t($ => $.apiKey.scopeChat);
      case "responses:invoke": return t($ => $.apiKey.scopeResponses);
      case "messages:invoke": return t($ => $.apiKey.scopeMessages);
      case "application_rag:invoke": return t($ => $.apiKey.scopeRagInvocation);
      case "prompt_application:invoke": return t($ => $.apiKey.scopePromptInvocation);
      case "agent_copilot:invoke": return t($ => $.apiKey.scopeAgentInvocation);
    }
  };
  return (
    <section className="surface-band workspace-api-keys api-key-lifecycle" id="workspace-api-keys" aria-labelledby="workspace-api-keys-title">
      <div className="section-heading">
        <div><p className="eyebrow">{t($ => $.apiKey.credentials)}</p><h3 id="workspace-api-keys-title">{t($ => $.apiKey.apiKeyLifecycle)}</h3></div>
        <span className="status-badge good">{t($ => $.apiKey.devTestInteractive)}</span>
      </div>

      <div className="api-key-lifecycle-scope">
        <article><span>{t($ => $.apiKey.application)}</span><strong>{applicationName || t($ => $.apiKey.noApplicationSelected)}</strong><code>{applicationId || t($ => $.apiKey.unbound)}</code></article>
        <article><span>{t($ => $.apiKey.boundary)}</span><strong>{applicationActive ? t($ => $.apiKey.activeApplication) : t($ => $.apiKey.readOnlyApplication)}</strong><p>{t($ => $.apiKey.separateIdentities)}</p></article>
      </div>

      {!selectedApplicationAvailable ? (
        <article className="api-key-lifecycle-empty" role="status"><h4>{t($ => $.apiKey.createApplicationFirst)}</h4><p>{t($ => $.apiKey.authoritativeCatalog)}</p></article>
      ) : applicationActive ? (
        <div className="api-key-lifecycle-layout">
          <details className="api-key-issue-disclosure" ref={issueDisclosure}>
            <summary>
              <span><strong>{t($ => $.apiKey.issueNewCredential)}</strong><small>{t($ => $.apiKey.chooseExpiryScopes)}</small></span>
              <span aria-hidden="true">+</span>
            </summary>
            <form className="api-key-issue-form" onSubmit={submitIssue}>
              <div className="api-key-card-heading"><div><p className="eyebrow">{t($ => $.apiKey.issueSettings)}</p><h4>{t($ => $.apiKey.oneTimeApiKey)}</h4></div><span className="status-badge neutral">{t($ => $.apiKey.memoryOnly)}</span></div>
              <label>{t($ => $.apiKey.displayName)}<input value={displayName} onChange={(event) => setDisplayName(event.target.value)} minLength={2} maxLength={80} disabled={!applicationActive || Boolean(rotation) || busy === "issue"} /></label>
              <label>{t($ => $.apiKey.expiresInDays)}<input type="number" value={expiresInDays} onChange={(event) => setExpiresInDays(Number(event.target.value))} min={1} max={90} disabled={!applicationActive || Boolean(rotation) || busy === "issue"} /></label>
              <fieldset disabled={!applicationActive || Boolean(rotation) || busy === "issue"}>
                <legend>{t($ => $.apiKey.gatewayScopes)}</legend>
                {AVAILABLE_SCOPES.map((item) => <label key={item.scope}><input type="checkbox" checked={scopes.includes(item.scope)} onChange={() => toggleScope(item.scope)} /> <code>{item.scope}</code><span>{scopeLabel(item.scope)}</span></label>)}
              </fieldset>
              <button type="submit" disabled={!applicationActive || Boolean(rotation) || busy !== "" || scopes.length === 0}>{busy === "issue" ? t($ => $.apiKey.issuing) : t($ => $.apiKey.issueApiKey)}</button>
              {rotation ? <p className="boundary-note">{t($ => $.apiKey.finishRotationFirst)}</p> : null}
            </form>
          </details>

          <article className={`api-key-one-time-panel ${issuedCredential ? "has-credential" : ""}`} aria-live="polite" ref={oneTimePanel}>
            <div className="api-key-card-heading"><div><p className="eyebrow">{t($ => $.apiKey.oneTimeHandoff)}</p><h4>{issuedCredential?.apiKeyId ?? t($ => $.apiKey.noPendingCredential)}</h4></div><span className={`status-badge ${issuedCredential ? "good" : "neutral"}`}>{issuedCredential ? t($ => $.apiKey.availableOnce) : t($ => $.apiKey.cleared)}</span></div>
            {issuedCredential ? (
              <>
                <p>{t($ => $.apiKey.tokenCannotReload)}</p>
                <code className="api-key-one-time-token">{issuedCredential.token}</code>
                <div className="api-key-one-time-actions">
                  <button type="button" onClick={() => void copyCredential()}>{t($ => $.apiKey.copyToken)}</button>
                  <button type="button" onClick={handoffCredential}>{t($ => $.apiKey.useInPlayground)}</button>
                  {selectedRecord?.scopes.includes("application_rag:invoke") ? <button type="button" onClick={handoffRAGCredential}>{t($ => $.apiKey.useInRag)}</button> : null}
                  {selectedRecord?.scopes.includes("prompt_application:invoke") ? <button type="button" onClick={handoffPromptCredential}>{t($ => $.apiKey.useInPrompt)}</button> : null}
                  <button type="button" className="secondary-action" onClick={() => { setIssuedCredential(null); setCopyStatus(""); }}>{t($ => $.apiKey.clearNow)}</button>
                </div>
                {copyStatus ? <p className="boundary-note">{copyStatus === "copied" ? t($ => $.apiKey.copiedToken) : t($ => $.apiKey.clipboardFailed)}</p> : null}
              </>
            ) : <p>{t($ => $.apiKey.noRawCredentialRetained)}</p>}
          </article>
        </div>
      ) : null}

      {rotation && applicationActive ? (
        <article className="api-key-rotation-panel" aria-labelledby="api-key-rotation-title">
          <div className="api-key-card-heading">
            <div>
              <p className="eyebrow">{t($ => $.apiKey.guidedRotation)} {rotation.phase === "replacement_pending" ? t($ => $.apiKey.replacementPending) : rotation.phase === "verification_pending" ? t($ => $.apiKey.verificationPending) : t($ => $.apiKey.replacementVerified)}</p>
              <h4 id="api-key-rotation-title">{t($ => $.apiKey.replaceKey, { apiKeyId: rotation.sourceApiKeyId })}</h4>
            </div>
            <span className={`status-badge ${rotation.phase === "verified" ? "good" : "neutral"}`}>
              {rotation.phase === "verified" ? t($ => $.apiKey.replacementVerified) : t($ => $.apiKey.originalStaysActive)}
            </span>
          </div>
          <ol className="api-key-rotation-steps">
            <li className="complete">{t($ => $.apiKey.sourceReviewed)}</li>
            <li className={rotation.phase !== "replacement_pending" ? "complete" : "current"}>{t($ => $.apiKey.replacementIssued)}</li>
            <li className={rotation.phase === "verified" ? "complete" : rotation.phase === "verification_pending" ? "current" : ""}>{t($ => $.apiKey.authenticationObserved)}</li>
            <li className={rotation.phase === "verified" ? "current" : ""}>{t($ => $.apiKey.retireOriginal)}</li>
          </ol>
          <dl className="api-key-rotation-meta">
            <div><dt>{t($ => $.apiKey.application)}</dt><dd>{rotation.applicationId}</dd></div>
            <div><dt>{t($ => $.apiKey.owner)}</dt><dd>{rotation.ownerSubjectRef}</dd></div>
            <div><dt>{t($ => $.apiKey.originalVersion)}</dt><dd>{rotation.sourceRecordVersion}</dd></div>
            <div><dt>{t($ => $.apiKey.replacement)}</dt><dd>{rotation.replacementApiKeyId || t($ => $.apiKey.notIssued)}</dd></div>
          </dl>
          <div className="api-key-rotation-scopes">
            <strong>{t($ => $.apiKey.lockedScopes)}</strong>
            <div>{rotation.scopes.map((scope) => <code key={scope}>{scope}</code>)}</div>
            <p>{t($ => $.apiKey.cannotChangeScopes)}</p>
          </div>
          {rotation.phase === "replacement_pending" ? (
            <div className="api-key-rotation-form">
              <label>{t($ => $.apiKey.replacementDisplayName)}<input value={rotationDisplayName} onChange={(event) => setRotationDisplayName(event.target.value)} minLength={2} maxLength={80} disabled={busy !== "" || !applicationActive} /></label>
              <label>{t($ => $.apiKey.replacementExpiresDays)}<input type="number" value={rotationExpiresInDays} onChange={(event) => setRotationExpiresInDays(Number(event.target.value))} min={1} max={90} disabled={busy !== "" || !applicationActive} /></label>
              <button type="button" onClick={() => void issueRotationReplacement()} disabled={busy !== "" || !applicationActive}>
                {busy === "rotation_issue" ? t($ => $.apiKey.issuingReplacement) : t($ => $.apiKey.issueSameScopeReplacement)}
              </button>
            </div>
          ) : (
            <div className="api-key-rotation-verification">
              <p>
                {t($ => $.apiKey.verifyReplacementBoundary)}</p>
              <button type="button" onClick={() => void checkRotationVerification()} disabled={busy !== ""}>
                {busy === "rotation_verify" ? t($ => $.apiKey.checkingAuthentication) : t($ => $.apiKey.checkReplacementAuthentication)}
              </button>
              {rotation.replacementLastUsedAt ? <p className="api-key-rotation-evidence">{t($ => $.apiKey.authenticationObservedAt)} <code>{rotation.replacementLastUsedAt}</code>.</p> : null}
            </div>
          )}
          {rotation.phase === "verified" ? (
            <div className="api-key-rotation-retirement">
              {!showRotationRetireConfirm ? (
                <button type="button" className="danger-action" onClick={() => setShowRotationRetireConfirm(true)} disabled={busy !== ""}>{t($ => $.apiKey.reviewOriginalRetirement)}</button>
              ) : (
                <div className="api-key-rotation-confirm" role="alert">
                  <strong>{t($ => $.apiKey.revokeOriginalQuestion, { apiKeyId: rotation.sourceApiKeyId })}</strong>
                  <p>{t($ => $.apiKey.retirementIrreversible)}</p>
                  <button type="button" className="danger-action" onClick={() => void retireRotationSource()} disabled={busy !== ""}>
                    {busy === "rotation_retire" ? t($ => $.apiKey.retiringOriginal) : t($ => $.apiKey.confirmRevokeOriginal)}
                  </button>
                  <button type="button" className="secondary-action" onClick={() => setShowRotationRetireConfirm(false)} disabled={busy !== ""}>{t($ => $.apiKey.keepOriginalActive)}</button>
                </div>
              )}
            </div>
          ) : null}
          <button type="button" className="secondary-action" onClick={cancelRotation} disabled={busy !== ""}>{t($ => $.apiKey.cancelGuidedRotation)}</button>
          <p className="boundary-note">{t($ => $.apiKey.rotationMemoryBoundary)}</p>
        </article>
      ) : null}

      <div className="api-key-list-controls">
        <label>{t($ => $.apiKey.effectiveState)}<select value={effectiveState} onChange={(event) => setEffectiveState(event.target.value as APIKeyEffectiveState | "")} disabled={!applicationId || busy !== ""}><option value="">{t($ => $.apiKey.all)}</option><option value="active">{t($ => $.apiKey.active)}</option><option value="expired">{t($ => $.apiKey.expired)}</option><option value="revoked">{t($ => $.apiKey.revoked)}</option></select></label>
        <button type="button" onClick={() => void loadRecords(false, "")} disabled={!applicationId || busy !== ""}>{busy === "list" ? t($ => $.apiKey.loading) : t($ => $.apiKey.refreshKeys)}</button>
      </div>

      {noticeMessage ? <p className={`api-key-operation-notice ${notice.tone}`} role={notice.tone === "bad" ? "alert" : "status"}>{noticeMessage}</p> : null}
      <p className="api-key-list-summary">{listMessage}</p>
      <div className="api-key-lifecycle-list" aria-label={t($ => $.apiKey.applicationApiKeys)}>
        {list.records.map((record) => (
          <APIKeyLifecycleRow
            key={record.apiKeyId}
            record={record}
            selected={selectedRecord?.apiKeyId === record.apiKeyId}
            pendingRevoke={pendingRevokeId === record.apiKeyId}
            rotationRole={rotation?.sourceApiKeyId === record.apiKeyId ? "source" : rotation?.replacementApiKeyId === record.apiKeyId ? "replacement" : ""}
            rotationActive={Boolean(rotation)}
            applicationActive={applicationActive}
            busy={busy}
            onRead={() => void loadDetail(record.apiKeyId)}
            onRevoke={() => void confirmRevoke(record)}
            onRotate={() => startRotation(record)}
          />
        ))}
      </div>
      {list.nextCursor ? <button type="button" onClick={() => void loadRecords(true)} disabled={busy !== ""}>{t($ => $.apiKey.loadMore)}</button> : null}
      {selectedRecord ? <APIKeyDetail record={selectedRecord} /> : null}
      <p className="boundary-note">{t($ => $.apiKey.devTestBoundary)}</p>
    </section>
  );
}

function APIKeyLifecycleRow({
  record,
  selected,
  pendingRevoke,
  rotationRole,
  rotationActive,
  applicationActive,
  busy,
  onRead,
  onRevoke,
  onRotate,
}: {
  record: APIKeyRecord;
  selected: boolean;
  pendingRevoke: boolean;
  rotationRole: "" | "source" | "replacement";
  rotationActive: boolean;
  applicationActive: boolean;
  busy: string;
  onRead: () => void;
  onRevoke: () => void;
  onRotate: () => void;
}) {
  const { t } = useTranslation("gateway");
  const active = record.lifecycleState === "active" && record.effectiveState === "active";
  return (
    <article className={`api-key-lifecycle-row ${selected ? "selected" : ""}`} data-selected={String(selected)}>
      <div className="api-key-row-main"><div><p className="eyebrow">{record.displayName}</p><h4>{record.apiKeyId}</h4></div><span className={`status-badge ${record.effectiveState === "active" ? "good" : "neutral"}`}>{rotationRole === "source" ? t($ => $.apiKey.rotationSourceDisplay) : rotationRole === "replacement" ? t($ => $.apiKey.replacement) : record.effectiveState === "active" ? t($ => $.apiKey.active) : record.effectiveState === "expired" ? t($ => $.apiKey.expired) : t($ => $.apiKey.revoked)}</span></div>
      <div className="api-key-scopes">{record.scopes.map((scope) => <code key={scope}>{scope}</code>)}</div>
      <dl className="api-key-row-meta"><div><dt>{t($ => $.apiKey.version)}</dt><dd>{record.recordVersion}</dd></div><div><dt>{t($ => $.apiKey.expires)}</dt><dd>{record.expiresAt}</dd></div><div><dt>{t($ => $.apiKey.lastUsed)}</dt><dd>{record.lastUsedAt ?? t($ => $.apiKey.notRecorded)}</dd></div></dl>
      <div className="api-key-row-actions">
        <button type="button" onClick={onRead} disabled={busy !== ""} aria-pressed={selected}>{t($ => $.apiKey.viewDetail)}</button>
        {active && applicationActive ? <button type="button" className="secondary-action" onClick={onRotate} disabled={busy !== "" || rotationActive}>{rotationRole === "source" ? t($ => $.apiKey.rotationActive) : t($ => $.apiKey.rotate)}</button> : null}
        <button type="button" className={pendingRevoke ? "danger-action" : "secondary-action"} onClick={onRevoke} disabled={busy !== "" || record.lifecycleState === "revoked" || rotationRole !== ""}>
          {rotationRole ? t($ => $.apiKey.managedByRotation) : pendingRevoke ? t($ => $.apiKey.confirmRevoke) : t($ => $.apiKey.revoke)}
        </button>
      </div>
    </article>
  );
}

function APIKeyDetail({ record }: { record: APIKeyRecord }) {
  const { t } = useTranslation("gateway");
  return (
    <article className="api-key-detail" aria-label={t($ => $.apiKey.selectedKeyDetail)}>
      <div className="api-key-card-heading"><div><p className="eyebrow">{t($ => $.apiKey.sanitizedDetail)}</p><h4>{record.apiKeyId}</h4></div><span className="status-badge neutral">{t($ => $.apiKey.versionValue, { version: record.recordVersion })}</span></div>
      <dl><div><dt>{t($ => $.apiKey.application)}</dt><dd>{record.applicationId}</dd></div><div><dt>{t($ => $.apiKey.owner)}</dt><dd>{record.ownerSubjectRef}</dd></div><div><dt>{t($ => $.apiKey.created)}</dt><dd>{record.createdAt}</dd></div><div><dt>{t($ => $.apiKey.expires)}</dt><dd>{record.expiresAt}</dd></div><div><dt>{t($ => $.apiKey.revoked)}</dt><dd>{record.revokedAt ?? t($ => $.apiKey.notRevoked)}</dd></div><div><dt>{t($ => $.apiKey.audit)}</dt><dd>{record.auditRef || t($ => $.apiKey.listProjection)}</dd></div></dl>
      <p className="boundary-note">{t($ => $.apiKey.noCredentialInProjection)}</p>
    </article>
  );
}

function OfflineAPIKeySummary({ view }: { view: WorkspaceApiKeysViewModel }) {
  const { t } = useTranslation("gateway");
  return (
    <section className="surface-band workspace-api-keys" id="workspace-api-keys" aria-labelledby="workspace-api-keys-title">
      <div className="section-heading"><div><p className="eyebrow">{t($ => $.apiKey.workspaceReadOnlySummary)}</p><h3 id="workspace-api-keys-title">{t($ => $.apiKey.apiKeys)}</h3></div><span className={`status-badge ${view.canRenderApiKeys ? "good" : "bad"}`}>{view.canRenderApiKeys ? t($ => $.apiKey.readOnlyReady) : t($ => $.apiKey.blocked)}</span></div>
      <p className="boundary-note">{t($ => $.apiKey.offlineWorkspaceBoundary)}</p>
      <div className="api-keys-summary">
        <article className="api-keys-route"><div className="card-title-row"><div><p className="eyebrow">{t($ => $.apiKey.summaryListRoute)}</p><h4>{view.routeId}</h4></div><span className="status-badge neutral">{view.requiredScope}</span></div><p className="route-path">{view.routePath}</p><dl className="tenant-meta"><div><dt>{t($ => $.apiKey.model)}</dt><dd>{view.readModel}</dd></div><div><dt>{t($ => $.apiKey.request)}</dt><dd>{view.requestId}</dd></div><div><dt>{t($ => $.apiKey.nextCursor)}</dt><dd>{view.nextCursor ?? t($ => $.apiKey.none)}</dd></div><div><dt>{t($ => $.apiKey.audit)}</dt><dd>{view.auditRef}</dd></div></dl></article>
        <div className="api-keys-metrics">{view.metrics.map((metric) => <OfflineMetric key={metric.label} metric={metric} />)}</div>
      </div>
      <div className="api-key-list">{view.apiKeys.map((apiKey) => <OfflineRow key={apiKey.apiKeyId} apiKey={apiKey} />)}</div>
      <div className="api-key-states">{view.statePreviews.map((state) => <OfflineState key={state.id} state={state} />)}</div>
    </section>
  );
}

function OfflineMetric({ metric }: { metric: WorkspaceApiKeysMetric }) {
  const { t } = useTranslation("gateway");
  const label = metric.label === "API keys" ? t($ => $.apiKey.apiKeys)
    : metric.label === "Active" ? t($ => $.apiKey.active)
    : metric.label === "Rotation" ? t($ => $.apiKey.rotationMetric)
    : metric.label === "Expiring" ? t($ => $.apiKey.expiringMetric) : metric.label;
  const detail = metric.label === "Active" ? t($ => $.apiKey.readOnlyReferences)
    : metric.label === "Rotation" ? t($ => $.apiKey.requiresOperatorReview)
    : metric.label === "Expiring" ? t($ => $.apiKey.explicitExpiry) : metric.detail;
  return <article className="api-key-metric"><span>{label}</span><strong>{metric.value}</strong><p>{detail}</p></article>;
}

function OfflineRow({ apiKey }: { apiKey: WorkspaceApiKeyRow }) {
  const { t } = useTranslation("gateway");
  return <article className="api-key-row"><div className="api-key-row-main"><div><p className="eyebrow">{apiKey.ownerSubjectRef}</p><h4>{apiKey.apiKeyId}</h4></div><span className={`status-badge ${apiKey.state === "active" ? "good" : "neutral"}`}>{apiKey.state === "active" ? t($ => $.apiKey.active) : apiKey.state === "expired" ? t($ => $.apiKey.expired) : apiKey.state === "revoked" ? t($ => $.apiKey.revoked) : apiKey.state === "rotation_required" ? t($ => $.apiKey.rotationRequired) : apiKey.state}</span></div><div className="api-key-scopes">{apiKey.scopes.map((scope) => <code key={scope}>{scope}</code>)}</div><dl className="api-key-row-meta"><div><dt>{t($ => $.apiKey.created)}</dt><dd>{apiKey.createdAt}</dd></div><div><dt>{t($ => $.apiKey.expires)}</dt><dd>{apiKey.expiresAt ?? t($ => $.apiKey.notSet)}</dd></div><div><dt>{t($ => $.apiKey.lastUsed)}</dt><dd>{apiKey.lastUsedAt ?? t($ => $.apiKey.notRecorded)}</dd></div></dl></article>;
}

function OfflineState({ state }: { state: WorkspaceApiKeysStatePreview }) {
  const { t } = useTranslation("gateway");
  const label = state.id === "ready" ? t($ => $.apiKey.stateReady)
    : state.id === "empty" ? t($ => $.apiKey.stateEmpty)
    : state.id === "denied" ? t($ => $.apiKey.stateDenied)
    : state.id === "stale" ? t($ => $.apiKey.stateStale)
    : state.id === "partial_failure" ? t($ => $.apiKey.statePartialFailure) : t($ => $.apiKey.stateForbiddenProjection);
  const status = state.status === "ready" ? t($ => $.apiKey.stateReady)
    : state.status === "empty" ? t($ => $.apiKey.stateEmpty)
    : state.status === "scope_denied" ? t($ => $.apiKey.stateDenied)
    : state.status === "stale" ? t($ => $.apiKey.stateStale)
    : state.status === "partial_failure" ? t($ => $.apiKey.statePartialFailure)
    : state.status === "blocked" ? t($ => $.apiKey.blocked)
    : state.status === "clear" ? t($ => $.apiKey.stateClear) : state.status;
  const summary = state.id === "ready" ? t($ => $.apiKey.stateReadySummary)
    : state.id === "empty" ? t($ => $.apiKey.stateEmptySummary)
    : state.id === "denied" ? t($ => $.apiKey.stateDeniedSummary)
    : state.id === "stale" ? t($ => $.apiKey.stateStaleSummary)
    : state.id === "partial_failure" ? t($ => $.apiKey.statePartialFailureSummary)
    : t($ => $.apiKey.stateForbiddenProjectionSummary);
  return <article className="api-key-state"><div><strong>{label}</strong><span>{status}</span></div><p>{summary}</p><small>{t($ => $.apiKey.offlineStateCounts, { count: state.itemCount, code: state.failureCode })}</small></article>;
}

function initialList(): APIKeyListResult {
  return { status: "empty", records: [], nextCursor: "", failureCode: "", requestId: "", auditRef: "", summary: "" };
}

function defaultDisplayName(applicationName: string): string {
  const prefix = applicationName.trim().slice(0, 54);
  return prefix ? `${prefix} browser key` : "Browser development key";
}

function replacementDisplayName(sourceDisplayName: string): string {
  const prefix = sourceDisplayName.trim().slice(0, 68);
  return prefix ? `${prefix} replacement` : "Rotated development key";
}

function rotationFailureNotice(error: unknown): OperationNotice {
  const failureCode = error instanceof APIKeyRotationSessionError
    ? error.failureCode
    : "api_key_rotation_state_invalid";
  return {
    tone: "bad",
    kind: "failure",
    failureCode,
  };
}

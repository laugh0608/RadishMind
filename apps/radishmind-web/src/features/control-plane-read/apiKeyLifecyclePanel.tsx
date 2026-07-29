import { useEffect, useRef, useState, type FormEvent } from "react";

import {
  issueAPIKey,
  listAPIKeyRecords,
  readAPIKeyLifecycleConfig,
  readAPIKeyRecord,
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
  summary: string;
  failureCode: string;
};

export function APIKeyLifecyclePanel({
  applicationId,
  applicationName,
  applicationActive,
  offlineView,
}: {
  applicationId: string;
  applicationName: string;
  applicationActive: boolean;
  offlineView: WorkspaceApiKeysViewModel;
}) {
  const [list, setList] = useState<APIKeyListResult>(() => initialList());
  const [effectiveState, setEffectiveState] = useState<APIKeyEffectiveState | "">("");
  const [displayName, setDisplayName] = useState(defaultDisplayName(applicationName));
  const [expiresInDays, setExpiresInDays] = useState(30);
  const [scopes, setScopes] = useState<APIKeyScope[]>(AVAILABLE_SCOPES.map(({ scope }) => scope));
  const [selectedRecord, setSelectedRecord] = useState<APIKeyRecord | null>(null);
  const [issuedCredential, setIssuedCredential] = useState<IssuedCredential | null>(null);
  const [pendingRevokeId, setPendingRevokeId] = useState("");
  const [rotation, setRotation] = useState<APIKeyRotationSession | null>(() => readAPIKeyRotationSession(applicationId));
  const [rotationDisplayName, setRotationDisplayName] = useState("");
  const [rotationExpiresInDays, setRotationExpiresInDays] = useState(30);
  const [showRotationRetireConfirm, setShowRotationRetireConfirm] = useState(false);
  const [busy, setBusy] = useState<
    "" | "list" | "issue" | "read" | "revoke" | "rotation_issue" | "rotation_verify" | "rotation_retire"
  >("");
  const [copyStatus, setCopyStatus] = useState("");
  const [notice, setNotice] = useState<OperationNotice>({ tone: "neutral", summary: "", failureCode: "" });
  const operationGeneration = useRef(0);

  useEffect(() => {
    operationGeneration.current += 1;
    setIssuedCredential(null);
    setSelectedRecord(null);
    setPendingRevokeId("");
    const currentRotation = synchronizeAPIKeyRotationApplication(applicationId);
    setRotation(currentRotation);
    setRotationDisplayName(currentRotation ? replacementDisplayName(currentRotation.sourceDisplayName) : "");
    setRotationExpiresInDays(30);
    setShowRotationRetireConfirm(false);
    setCopyStatus("");
    setNotice({ tone: "neutral", summary: "", failureCode: "" });
    setDisplayName(defaultDisplayName(applicationName));
    setScopes(AVAILABLE_SCOPES.map(({ scope }) => scope));
    setExpiresInDays(30);
    setEffectiveState("");
    if (config.mode === "dev_api_key_lifecycle_http" && applicationId) void loadRecords(false, "", "");
  }, [applicationId, applicationName]);

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

  async function loadRecords(
    append: boolean,
    cursor = list.nextCursor,
    stateFilter: APIKeyEffectiveState | "" = effectiveState,
  ) {
    if (!applicationId) {
      setList(initialList("Select an application before loading API keys."));
      return;
    }
    const generation = ++operationGeneration.current;
    setBusy("list");
    const next = await listAPIKeyRecords(config, applicationId, append ? cursor : "", stateFilter);
    if (operationGeneration.current !== generation) return;
    setBusy("");
    setList(append && next.status === "ready" ? { ...next, records: [...list.records, ...next.records] } : next);
    if (next.status === "failed") setNotice({ tone: "bad", summary: next.summary, failureCode: next.failureCode });
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
      setNotice({ tone: "good", summary: result.summary, failureCode: "" });
      await loadRecords(false, "");
      return;
    }
    setNotice({ tone: "bad", summary: result.summary, failureCode: result.failureCode });
  }

  async function loadDetail(apiKeyId: string) {
    const generation = ++operationGeneration.current;
    setBusy("read");
    const result = await readAPIKeyRecord(config, apiKeyId);
    if (operationGeneration.current !== generation) return;
    setBusy("");
    if (result.status === "loaded" && result.record) {
      setSelectedRecord(result.record);
      setNotice({ tone: "neutral", summary: result.summary, failureCode: "" });
      return;
    }
    setNotice({ tone: "bad", summary: result.summary, failureCode: result.failureCode });
  }

  async function confirmRevoke(record: APIKeyRecord) {
    if (pendingRevokeId !== record.apiKeyId) {
      setPendingRevokeId(record.apiKeyId);
      setNotice({ tone: "neutral", summary: `Confirm revocation for ${record.apiKeyId}. This action cannot be reversed.`, failureCode: "" });
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
      setNotice({ tone: "good", summary: result.summary, failureCode: "" });
      await loadRecords(false, "");
      return;
    }
    setNotice({ tone: "bad", summary: result.summary, failureCode: result.failureCode });
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
        summary: `Rotation started for ${record.apiKeyId}. The original key stays active until its replacement is verified.`,
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
      summary: "Rotation session cleared. Stored API key records were not changed.",
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
      setNotice({ tone: "bad", summary: result.summary, failureCode: result.failureCode });
      return;
    }
    try {
      const next = recordAPIKeyRotationReplacement(applicationId, result.record);
      setRotation(next);
      setIssuedCredential({ apiKeyId: result.record.apiKeyId, token: result.credentialToken });
      setSelectedRecord(result.record);
      setNotice({
        tone: "good",
        summary: "Replacement key issued. The original key remains active; validate the replacement before retirement.",
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
      setNotice({ tone: "bad", summary: result.summary, failureCode: result.failureCode });
      return;
    }
    try {
      const next = refreshAPIKeyRotationVerification(applicationId, result.record);
      setRotation(next);
      setSelectedRecord(result.record);
      setShowRotationRetireConfirm(false);
      setNotice(next.phase === "verified"
        ? {
          tone: "good",
          summary: `Replacement authentication verified at ${next.replacementLastUsedAt}. Review the original key before retirement.`,
          failureCode: "",
        }
        : {
          tone: "neutral",
          summary: "The replacement key has no successful Gateway authentication yet. The original key remains active.",
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
      setNotice({ tone: "bad", summary: source.summary, failureCode: source.failureCode });
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
      setNotice({ tone: "bad", summary: revoked.summary, failureCode: revoked.failureCode });
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
        summary: "Rotation completed. The original key is revoked and the verified replacement remains active.",
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
      setCopyStatus("Copied to the clipboard. The Web app did not persist the token.");
    } catch {
      setCopyStatus("Clipboard access failed. Select the one-time token manually before clearing it.");
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
  return (
    <section className="surface-band workspace-api-keys api-key-lifecycle" id="workspace-api-keys" aria-labelledby="workspace-api-keys-title">
      <div className="section-heading">
        <div><p className="eyebrow">User Workspace</p><h3 id="workspace-api-keys-title">API key lifecycle</h3></div>
        <span className="status-badge good">dev/test interactive</span>
      </div>

      <div className="api-key-lifecycle-scope">
        <article><span>Application</span><strong>{applicationName || "No application selected"}</strong><code>{applicationId || "unbound"}</code></article>
        <article><span>Boundary</span><strong>{applicationActive ? "Active application" : "Read-only application"}</strong><p>Management identity and Gateway credential remain separate.</p></article>
      </div>

      {!selectedApplicationAvailable ? (
        <article className="api-key-lifecycle-empty" role="status"><h4>Create or select an application first</h4><p>API keys can only be issued and listed against the authoritative application catalog.</p></article>
      ) : (
        <div className="api-key-lifecycle-layout">
          <form className="api-key-issue-form" onSubmit={submitIssue}>
            <div className="api-key-card-heading"><div><p className="eyebrow">Issue credential</p><h4>One-time API key</h4></div><span className="status-badge neutral">memory only</span></div>
            <label>Display name<input value={displayName} onChange={(event) => setDisplayName(event.target.value)} minLength={2} maxLength={80} disabled={!applicationActive || Boolean(rotation) || busy === "issue"} /></label>
            <label>Expires in days<input type="number" value={expiresInDays} onChange={(event) => setExpiresInDays(Number(event.target.value))} min={1} max={90} disabled={!applicationActive || Boolean(rotation) || busy === "issue"} /></label>
            <fieldset disabled={!applicationActive || Boolean(rotation) || busy === "issue"}>
              <legend>Gateway scopes</legend>
              {AVAILABLE_SCOPES.map((item) => <label key={item.scope}><input type="checkbox" checked={scopes.includes(item.scope)} onChange={() => toggleScope(item.scope)} /> <code>{item.scope}</code><span>{item.label}</span></label>)}
            </fieldset>
            <button type="submit" disabled={!applicationActive || Boolean(rotation) || busy !== "" || scopes.length === 0}>{busy === "issue" ? "Issuing…" : "Issue API key"}</button>
            {!applicationActive ? <p className="boundary-note">Archived applications keep existing metadata readable and revocable but cannot issue new credentials.</p> : null}
            {rotation ? <p className="boundary-note">Finish or cancel the active guided rotation before issuing an unrelated key.</p> : null}
          </form>

          <article className="api-key-one-time-panel" aria-live="polite">
            <div className="api-key-card-heading"><div><p className="eyebrow">One-time handoff</p><h4>{issuedCredential?.apiKeyId ?? "No pending credential"}</h4></div><span className={`status-badge ${issuedCredential ? "good" : "neutral"}`}>{issuedCredential ? "available once" : "cleared"}</span></div>
            {issuedCredential ? (
              <>
                <p>This raw token cannot be loaded again. Copy it or hand it directly to an explicitly scoped invocation surface before clearing.</p>
                <code className="api-key-one-time-token">{issuedCredential.token}</code>
                <div className="api-key-one-time-actions">
                  <button type="button" onClick={() => void copyCredential()}>Copy token</button>
                  <button type="button" onClick={handoffCredential}>Use in Playground</button>
                  {selectedRecord?.scopes.includes("application_rag:invoke") ? <button type="button" onClick={handoffRAGCredential}>Use in Application RAG</button> : null}
                  {selectedRecord?.scopes.includes("prompt_application:invoke") ? <button type="button" onClick={handoffPromptCredential}>Use in Prompt Application</button> : null}
                  <button type="button" className="secondary-action" onClick={() => { setIssuedCredential(null); setCopyStatus(""); }}>Clear now</button>
                </div>
                {copyStatus ? <p className="boundary-note">{copyStatus}</p> : null}
              </>
            ) : <p>No raw credential is retained. Lists, details, errors, history, and reloads only receive sanitized metadata.</p>}
          </article>
        </div>
      )}

      {rotation ? (
        <article className="api-key-rotation-panel" aria-labelledby="api-key-rotation-title">
          <div className="api-key-card-heading">
            <div>
              <p className="eyebrow">Guided rotation · {rotation.phase.replaceAll("_", " ")}</p>
              <h4 id="api-key-rotation-title">Replace {rotation.sourceApiKeyId}</h4>
            </div>
            <span className={`status-badge ${rotation.phase === "verified" ? "good" : "neutral"}`}>
              {rotation.phase === "verified" ? "replacement verified" : "original stays active"}
            </span>
          </div>
          <ol className="api-key-rotation-steps">
            <li className="complete">Source reviewed</li>
            <li className={rotation.phase !== "replacement_pending" ? "complete" : "current"}>Replacement issued</li>
            <li className={rotation.phase === "verified" ? "complete" : rotation.phase === "verification_pending" ? "current" : ""}>Gateway authentication observed</li>
            <li className={rotation.phase === "verified" ? "current" : ""}>Retire original</li>
          </ol>
          <dl className="api-key-rotation-meta">
            <div><dt>Application</dt><dd>{rotation.applicationId}</dd></div>
            <div><dt>Owner</dt><dd>{rotation.ownerSubjectRef}</dd></div>
            <div><dt>Original version</dt><dd>{rotation.sourceRecordVersion}</dd></div>
            <div><dt>Replacement</dt><dd>{rotation.replacementApiKeyId || "not issued"}</dd></div>
          </dl>
          <div className="api-key-rotation-scopes">
            <strong>Locked scopes</strong>
            <div>{rotation.scopes.map((scope) => <code key={scope}>{scope}</code>)}</div>
            <p>Guided rotation cannot add or remove permissions. Use a separate key lifecycle decision for scope changes.</p>
          </div>
          {rotation.phase === "replacement_pending" ? (
            <div className="api-key-rotation-form">
              <label>Replacement display name<input value={rotationDisplayName} onChange={(event) => setRotationDisplayName(event.target.value)} minLength={2} maxLength={80} disabled={busy !== "" || !applicationActive} /></label>
              <label>Replacement expires in days<input type="number" value={rotationExpiresInDays} onChange={(event) => setRotationExpiresInDays(Number(event.target.value))} min={1} max={90} disabled={busy !== "" || !applicationActive} /></label>
              <button type="button" onClick={() => void issueRotationReplacement()} disabled={busy !== "" || !applicationActive}>
                {busy === "rotation_issue" ? "Issuing replacement…" : "Issue same-scope replacement"}
              </button>
            </div>
          ) : (
            <div className="api-key-rotation-verification">
              <p>
                Use the one-time replacement credential through an existing scoped surface or external client, then check its exact record.
                The original key will not be changed during verification.
              </p>
              <button type="button" onClick={() => void checkRotationVerification()} disabled={busy !== ""}>
                {busy === "rotation_verify" ? "Checking authentication…" : "Check replacement authentication"}
              </button>
              {rotation.replacementLastUsedAt ? <p className="api-key-rotation-evidence">Gateway authentication observed at <code>{rotation.replacementLastUsedAt}</code>.</p> : null}
            </div>
          )}
          {rotation.phase === "verified" ? (
            <div className="api-key-rotation-retirement">
              {!showRotationRetireConfirm ? (
                <button type="button" className="danger-action" onClick={() => setShowRotationRetireConfirm(true)} disabled={busy !== ""}>Review original retirement</button>
              ) : (
                <div className="api-key-rotation-confirm" role="alert">
                  <strong>Revoke original key {rotation.sourceApiKeyId}?</strong>
                  <p>This is irreversible. The source will be read again and revoked with its current version; the verified replacement remains active.</p>
                  <button type="button" className="danger-action" onClick={() => void retireRotationSource()} disabled={busy !== ""}>
                    {busy === "rotation_retire" ? "Retiring original…" : "Confirm verified replacement and revoke original"}
                  </button>
                  <button type="button" className="secondary-action" onClick={() => setShowRotationRetireConfirm(false)} disabled={busy !== ""}>Keep original active</button>
                </div>
              )}
            </div>
          ) : null}
          <button type="button" className="secondary-action" onClick={cancelRotation} disabled={busy !== ""}>Cancel guided rotation</button>
          <p className="boundary-note">Only sanitized rotation metadata survives an in-app route handoff. Browser reload clears the session, and no API key is changed automatically.</p>
        </article>
      ) : null}

      <div className="api-key-list-controls">
        <label>Effective state<select value={effectiveState} onChange={(event) => setEffectiveState(event.target.value as APIKeyEffectiveState | "")} disabled={!applicationId || busy !== ""}><option value="">All</option><option value="active">Active</option><option value="expired">Expired</option><option value="revoked">Revoked</option></select></label>
        <button type="button" onClick={() => void loadRecords(false, "")} disabled={!applicationId || busy !== ""}>{busy === "list" ? "Loading…" : "Refresh keys"}</button>
      </div>

      {notice.summary ? <p className={`api-key-operation-notice ${notice.tone}`} role={notice.tone === "bad" ? "alert" : "status"}>{notice.failureCode ? `${notice.failureCode}: ` : ""}{notice.summary}</p> : null}
      <p className="api-key-list-summary">{list.summary}</p>
      <div className="api-key-lifecycle-list" aria-label="Application API keys">
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
      {list.nextCursor ? <button type="button" onClick={() => void loadRecords(true)} disabled={busy !== ""}>Load more</button> : null}
      {selectedRecord ? <APIKeyDetail record={selectedRecord} /> : null}
      <p className="boundary-note">This dev/test surface does not enable production credentials, workspace membership, quota, rate limits, billing, provider secrets, or public production Gateway access.</p>
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
  const active = record.lifecycleState === "active" && record.effectiveState === "active";
  return (
    <article className={`api-key-lifecycle-row ${selected ? "selected" : ""}`}>
      <div className="api-key-row-main"><div><p className="eyebrow">{record.displayName}</p><h4>{record.apiKeyId}</h4></div><span className={`status-badge ${record.effectiveState === "active" ? "good" : "neutral"}`}>{rotationRole || record.effectiveState}</span></div>
      <div className="api-key-scopes">{record.scopes.map((scope) => <code key={scope}>{scope}</code>)}</div>
      <dl className="api-key-row-meta"><div><dt>Version</dt><dd>{record.recordVersion}</dd></div><div><dt>Expires</dt><dd>{record.expiresAt}</dd></div><div><dt>Last used</dt><dd>{record.lastUsedAt ?? "not recorded"}</dd></div></dl>
      <div className="api-key-row-actions">
        <button type="button" onClick={onRead} disabled={busy !== ""}>View detail</button>
        {active && applicationActive ? <button type="button" className="secondary-action" onClick={onRotate} disabled={busy !== "" || rotationActive}>{rotationRole === "source" ? "Rotation active" : "Rotate"}</button> : null}
        <button type="button" className={pendingRevoke ? "danger-action" : "secondary-action"} onClick={onRevoke} disabled={busy !== "" || record.lifecycleState === "revoked" || rotationRole !== ""}>
          {rotationRole ? "Managed by rotation" : pendingRevoke ? "Confirm revoke" : "Revoke"}
        </button>
      </div>
    </article>
  );
}

function APIKeyDetail({ record }: { record: APIKeyRecord }) {
  return (
    <article className="api-key-detail" aria-label="Selected API key detail">
      <div className="api-key-card-heading"><div><p className="eyebrow">Sanitized detail</p><h4>{record.apiKeyId}</h4></div><span className="status-badge neutral">version {record.recordVersion}</span></div>
      <dl><div><dt>Application</dt><dd>{record.applicationId}</dd></div><div><dt>Owner</dt><dd>{record.ownerSubjectRef}</dd></div><div><dt>Created</dt><dd>{record.createdAt}</dd></div><div><dt>Expires</dt><dd>{record.expiresAt}</dd></div><div><dt>Revoked</dt><dd>{record.revokedAt ?? "not revoked"}</dd></div><div><dt>Audit</dt><dd>{record.auditRef || "list projection"}</dd></div></dl>
      <p className="boundary-note">Credential token and digest are structurally absent from this detail projection.</p>
    </article>
  );
}

function OfflineAPIKeySummary({ view }: { view: WorkspaceApiKeysViewModel }) {
  return (
    <section className="surface-band workspace-api-keys" id="workspace-api-keys" aria-labelledby="workspace-api-keys-title">
      <div className="section-heading"><div><p className="eyebrow">User Workspace</p><h3 id="workspace-api-keys-title">API keys</h3></div><span className={`status-badge ${view.canRenderApiKeys ? "good" : "bad"}`}>{view.canRenderApiKeys ? "read-only ready" : "blocked"}</span></div>
      <div className="api-keys-summary">
        <article className="api-keys-route"><div className="card-title-row"><div><p className="eyebrow">API Key Summary List Route</p><h4>{view.routeId}</h4></div><span className="status-badge neutral">{view.requiredScope}</span></div><p className="route-path">{view.routePath}</p><dl className="tenant-meta"><div><dt>Model</dt><dd>{view.readModel}</dd></div><div><dt>Request</dt><dd>{view.requestId}</dd></div><div><dt>Next cursor</dt><dd>{view.nextCursor ?? "none"}</dd></div><div><dt>Audit</dt><dd>{view.auditRef}</dd></div></dl></article>
        <div className="api-keys-metrics">{view.metrics.map((metric) => <OfflineMetric key={metric.label} metric={metric} />)}</div>
      </div>
      <div className="api-key-list">{view.apiKeys.map((apiKey) => <OfflineRow key={apiKey.apiKeyId} apiKey={apiKey} />)}</div>
      <div className="api-key-states">{view.statePreviews.map((state) => <OfflineState key={state.id} state={state} />)}</div>
    </section>
  );
}

function OfflineMetric({ metric }: { metric: WorkspaceApiKeysMetric }) {
  return <article className="api-key-metric"><span>{metric.label}</span><strong>{metric.value}</strong><p>{metric.detail}</p></article>;
}

function OfflineRow({ apiKey }: { apiKey: WorkspaceApiKeyRow }) {
  return <article className="api-key-row"><div className="api-key-row-main"><div><p className="eyebrow">{apiKey.ownerSubjectRef}</p><h4>{apiKey.apiKeyId}</h4></div><span className={`status-badge ${apiKey.state === "active" ? "good" : "neutral"}`}>{apiKey.state}</span></div><div className="api-key-scopes">{apiKey.scopes.map((scope) => <code key={scope}>{scope}</code>)}</div><dl className="api-key-row-meta"><div><dt>Created</dt><dd>{apiKey.createdAt}</dd></div><div><dt>Expires</dt><dd>{apiKey.expiresAt ?? "not set"}</dd></div><div><dt>Last used</dt><dd>{apiKey.lastUsedAt ?? "not recorded"}</dd></div></dl></article>;
}

function OfflineState({ state }: { state: WorkspaceApiKeysStatePreview }) {
  return <article className="api-key-state"><div><strong>{state.label}</strong><span>{state.status}</span></div><p>{state.summary}</p><small>items {state.itemCount} / failure {state.failureCode}</small></article>;
}

function initialList(summary = "Select an application to load API keys."): APIKeyListResult {
  return { status: "empty", records: [], nextCursor: "", failureCode: "", requestId: "", auditRef: "", summary };
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
    summary: "Guided rotation stopped without revoking the original key. Review the current key records before continuing.",
    failureCode,
  };
}

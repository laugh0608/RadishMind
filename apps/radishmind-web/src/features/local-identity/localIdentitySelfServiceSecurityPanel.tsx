import {
  useEffect,
  useMemo,
  useRef,
  useState,
  type FormEvent,
  type ReactNode,
} from "react";

import {
  LocalIdentitySelfServiceSecurityError,
  listLocalIdentitySelfServiceSessions,
  localIdentitySelfServiceSecurityFailureKind,
  revokeLocalIdentitySelfServiceSession,
  revokeOtherLocalIdentitySelfServiceSessions,
  rotateLocalIdentitySelfServiceCredential,
  type LocalIdentitySelfServiceSecurityFailureKind,
  type LocalIdentitySelfServiceSessionSummary,
} from "./localIdentitySelfServiceSecurityConsumer.ts";
import {
  localIdentitySelfServiceSecurityResponseMatchesScope,
  localIdentitySelfServiceSecurityScope,
  mergeLocalIdentitySelfServiceSessions,
  projectLocalIdentitySelfServiceSessions,
} from "./localIdentitySelfServiceSecurityState.ts";
import type {
  LocalIdentityAccountProfile,
  LocalIdentityConsumerConfig,
} from "./localIdentityConsumer.ts";

type DirectoryState =
  | { status: "loading" }
  | { status: "empty"; snapshotAt: string }
  | { status: "ready"; snapshotAt: string }
  | { status: "failed"; failure: SecurityFailure };

type SecurityFailure = {
  kind: LocalIdentitySelfServiceSecurityFailureKind;
  title: string;
  message: string;
  code: string;
};

type OperationState =
  | { status: "idle" }
  | { status: "pending"; action: "exact" | "bulk" | "credential" }
  | { status: "success"; message: string }
  | { status: "failed"; failure: SecurityFailure };

type ConfirmationState =
  | { kind: "none" }
  | { kind: "exact"; sessionId: string }
  | { kind: "bulk" }
  | { kind: "credential" };

export function LocalIdentitySelfServiceSecurityPanel({
  config,
  profile,
  onClose,
  onRefreshProfile,
  onAuthenticationRequired,
  onSessionChanged,
  onLinkOIDC,
  onLogout,
  accountAction,
  accountActionError,
}: {
  config: LocalIdentityConsumerConfig;
  profile: LocalIdentityAccountProfile;
  onClose: () => void;
  onRefreshProfile: () => Promise<void>;
  onAuthenticationRequired: () => void;
  onSessionChanged: () => void;
  onLinkOIDC: () => Promise<void>;
  onLogout: () => Promise<void>;
  accountAction: "" | "link" | "logout" | "revoke";
  accountActionError: string;
}) {
  const mounted = useRef(true);
  const requestGeneration = useRef(0);
  const requestController = useRef<AbortController | null>(null);
  const pendingCredential = useRef<{ currentPassword: string; newPassword: string } | null>(null);
  const [directory, setDirectory] = useState<DirectoryState>({ status: "loading" });
  const [sessions, setSessions] = useState<LocalIdentitySelfServiceSessionSummary[]>([]);
  const [nextCursor, setNextCursor] = useState("");
  const [loadingMore, setLoadingMore] = useState(false);
  const [selectedSessionId, setSelectedSessionId] = useState("");
  const [confirmation, setConfirmation] = useState<ConfirmationState>({ kind: "none" });
  const [operation, setOperation] = useState<OperationState>({ status: "idle" });
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [newPasswordConfirmation, setNewPasswordConfirmation] = useState("");
  const [credentialImpactConfirmed, setCredentialImpactConfirmed] = useState(false);
  const [credentialInputError, setCredentialInputError] = useState("");

  const projection = useMemo(() => {
    try {
      return projectLocalIdentitySelfServiceSessions(sessions, profile.session.sessionId);
    } catch {
      return null;
    }
  }, [profile.session.sessionId, sessions]);
  const selectedSession = sessions.find((session) => session.sessionId === selectedSessionId) ?? null;
  const busy = operation.status === "pending" || accountAction !== "";

  useEffect(() => {
    mounted.current = true;
    void loadSessions();
    return () => {
      mounted.current = false;
      requestGeneration.current += 1;
      requestController.current?.abort();
      pendingCredential.current = null;
    };
  }, []);

  async function loadSessions(cursor = "") {
    const controller = replaceController();
    const generation = ++requestGeneration.current;
    const expectedScope = localIdentitySelfServiceSecurityScope(profile, generation);
    if (cursor === "") {
      setDirectory({ status: "loading" });
      setSessions([]);
      setNextCursor("");
      setSelectedSessionId("");
      setConfirmation({ kind: "none" });
    } else {
      setLoadingMore(true);
    }
    try {
      const page = await listLocalIdentitySelfServiceSessions(config, {
        state: "all",
        limit: 100,
        ...(cursor === "" ? {} : { cursor }),
      }, controller.signal);
      const observedScope = localIdentitySelfServiceSecurityScope(profile, requestGeneration.current);
      if (!mounted.current || controller.signal.aborted ||
        !localIdentitySelfServiceSecurityResponseMatchesScope(expectedScope, observedScope)) return;
      let nextSessions: LocalIdentitySelfServiceSessionSummary[];
      if (cursor === "") {
        nextSessions = page.sessions;
      } else {
        if (directory.status !== "ready" || directory.snapshotAt !== page.snapshotAt) {
          throw invalidDirectory("Session pagination snapshot changed.");
        }
        nextSessions = mergeLocalIdentitySelfServiceSessions(sessions, page.sessions);
      }
      const nextProjection = projectLocalIdentitySelfServiceSessions(nextSessions, profile.session.sessionId);
      if (nextSessions.length > 0 && !page.nextCursor && !nextProjection.currentSession) {
        throw invalidDirectory("The complete session directory omitted the current session.");
      }
      setSessions(nextSessions);
      setNextCursor(page.nextCursor ?? "");
      setDirectory(nextSessions.length === 0
        ? { status: "empty", snapshotAt: page.snapshotAt }
        : { status: "ready", snapshotAt: page.snapshotAt });
    } catch (error) {
      if (isAbort(error) || !mounted.current || controller.signal.aborted ||
        requestGeneration.current !== generation) return;
      if (localIdentitySelfServiceSecurityFailureKind(error) === "authentication_required") {
        onAuthenticationRequired();
        return;
      }
      setDirectory({ status: "failed", failure: securityFailure(error) });
    } finally {
      if (mounted.current && requestGeneration.current === generation) setLoadingMore(false);
    }
  }

  function reviewExactRevocation(session: LocalIdentitySelfServiceSessionSummary) {
    invalidateInteraction();
    setSelectedSessionId(session.sessionId);
    setConfirmation({ kind: "exact", sessionId: session.sessionId });
  }

  function reviewBulkRevocation() {
    invalidateInteraction();
    setConfirmation({ kind: "bulk" });
  }

  function reviewCredentialRotation(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setCredentialInputError("");
    if (!profile.capabilities.hasActiveLocalCredential) {
      setCredentialInputError("This account has no active local credential to rotate.");
      return;
    }
    if (!profile.capabilities.recentAuthentication) {
      setCredentialInputError("Sign out and authenticate again before rotating the local credential.");
      return;
    }
    if (nextCursor !== "") {
      setCredentialInputError("Load the complete canonical directory before reviewing the exact session impact.");
      return;
    }
    if (currentPassword.length < 1 || currentPassword.length > 1024 ||
      newPassword.length < 12 || newPassword.length > 1024) {
      setCredentialInputError("Enter the current password and a replacement of 12–1024 characters.");
      return;
    }
    if (newPassword !== newPasswordConfirmation) {
      setCredentialInputError("The replacement password confirmation does not match.");
      return;
    }
    if (!credentialImpactConfirmed) {
      setCredentialInputError("Confirm the session impact before continuing.");
      return;
    }
    pendingCredential.current = { currentPassword, newPassword };
    clearCredentialInput(false);
    invalidateInteraction(false);
    setConfirmation({ kind: "credential" });
  }

  async function commitExactRevocation() {
    if (confirmation.kind !== "exact") return;
    const target = sessions.find((session) => session.sessionId === confirmation.sessionId);
    if (!target || target.effectiveState !== "active") {
      setOperation({ status: "failed", failure: invalidSelectionFailure() });
      setConfirmation({ kind: "none" });
      return;
    }
    const operationScope = beginOperation("exact");
    try {
      const result = await revokeLocalIdentitySelfServiceSession(config, {
        sessionId: target.sessionId,
        expectedRecordVersion: target.recordVersion,
      }, operationScope.controller.signal);
      if (!acceptOperation(operationScope)) return;
      setConfirmation({ kind: "none" });
      setSelectedSessionId("");
      onSessionChanged();
      if (result.currentSessionRevoked) {
        setOperation({ status: "success", message: "Current session revoked. Sign in again to continue." });
        onAuthenticationRequired();
        return;
      }
      setOperation({ status: "success", message: "The exact selected session was revoked." });
      await onRefreshProfile();
      if (acceptOperation(operationScope)) await loadSessions();
    } catch (error) {
      handleOperationFailure(error, operationScope);
    }
  }

  async function commitBulkRevocation() {
    if (confirmation.kind !== "bulk" || nextCursor !== "") return;
    const operationScope = beginOperation("bulk");
    try {
      const result = await revokeOtherLocalIdentitySelfServiceSessions(config, operationScope.controller.signal);
      if (!acceptOperation(operationScope)) return;
      setConfirmation({ kind: "none" });
      onSessionChanged();
      setOperation({
        status: "success",
        message: `${result.revokedCount} other active session${result.revokedCount === 1 ? "" : "s"} revoked.`,
      });
      await onRefreshProfile();
      if (acceptOperation(operationScope)) await loadSessions();
    } catch (error) {
      handleOperationFailure(error, operationScope);
    }
  }

  async function commitCredentialRotation() {
    if (confirmation.kind !== "credential") return;
    const input = pendingCredential.current;
    pendingCredential.current = null;
    if (!input) {
      setOperation({ status: "failed", failure: invalidCredentialReviewFailure() });
      setConfirmation({ kind: "none" });
      return;
    }
    setConfirmation({ kind: "none" });
    const operationScope = beginOperation("credential");
    try {
      const result = await rotateLocalIdentitySelfServiceCredential(config, input, operationScope.controller.signal);
      input.currentPassword = "";
      input.newPassword = "";
      if (!acceptOperation(operationScope)) return;
      onSessionChanged();
      if (result.currentSessionRevoked) {
        setOperation({
          status: "success",
          message: `Credential rotated under ${result.policyVersion}; the current local session is closed.`,
        });
        onAuthenticationRequired();
        return;
      }
      setOperation({
        status: "success",
        message: `Credential rotated; ${result.revokedSessionCount} source-bound local session${result.revokedSessionCount === 1 ? "" : "s"} revoked.`,
      });
      await onRefreshProfile();
      if (acceptOperation(operationScope)) await loadSessions();
    } catch (error) {
      input.currentPassword = "";
      input.newPassword = "";
      handleOperationFailure(error, operationScope);
    }
  }

  function beginOperation(action: "exact" | "bulk" | "credential") {
    const controller = replaceController();
    const generation = ++requestGeneration.current;
    setOperation({ status: "pending", action });
    return {
      controller,
      scope: localIdentitySelfServiceSecurityScope(profile, generation),
    };
  }

  function acceptOperation(operationScope: {
    controller: AbortController;
    scope: ReturnType<typeof localIdentitySelfServiceSecurityScope>;
  }): boolean {
    return mounted.current && !operationScope.controller.signal.aborted &&
      localIdentitySelfServiceSecurityResponseMatchesScope(
        operationScope.scope,
        localIdentitySelfServiceSecurityScope(profile, requestGeneration.current),
      );
  }

  function handleOperationFailure(
    error: unknown,
    operationScope: { controller: AbortController; scope: ReturnType<typeof localIdentitySelfServiceSecurityScope> },
  ) {
    if (isAbort(error) || !acceptOperation(operationScope)) return;
    setConfirmation({ kind: "none" });
    setSelectedSessionId("");
    clearCredentialInput();
    const failure = securityFailure(error);
    if (failure.kind === "authentication_required") {
      onAuthenticationRequired();
      return;
    }
    setOperation({ status: "failed", failure });
  }

  function replaceController(): AbortController {
    requestController.current?.abort();
    const controller = new AbortController();
    requestController.current = controller;
    return controller;
  }

  function invalidateInteraction(clearCredential = true) {
    setOperation({ status: "idle" });
    setConfirmation({ kind: "none" });
    if (clearCredential) clearCredentialInput();
  }

  function cancelConfirmation() {
    setConfirmation({ kind: "none" });
    setSelectedSessionId("");
    clearCredentialInput();
  }

  function clearCredentialInput(clearPending = true) {
    if (clearPending) pendingCredential.current = null;
    setCurrentPassword("");
    setNewPassword("");
    setNewPasswordConfirmation("");
    setCredentialImpactConfirmed(false);
    setCredentialInputError("");
  }

  return (
    <section className="local-identity-security-surface" aria-labelledby="local-identity-security-title">
      <header className="local-identity-security-heading">
        <div>
          <p className="eyebrow">Account security · development/test</p>
          <h2 id="local-identity-security-title">Sessions &amp; local credential</h2>
          <p>{profile.account.displayName} · <code>{profile.account.userId}</code></p>
        </div>
        <div className="local-identity-security-heading-actions">
          <button type="button" onClick={() => void onLinkOIDC()} disabled={busy || !profile.capabilities.oidcEnabled || !profile.capabilities.recentAuthentication}>
            {accountAction === "link" ? "Opening…" : "Link Radish identity"}
          </button>
          <button type="button" onClick={() => void onLogout()} disabled={busy}>
            {accountAction === "logout" ? "Signing out…" : "Sign out"}
          </button>
          <button type="button" className="local-identity-security-close" onClick={onClose} disabled={busy} aria-label="Close account security">
            ×
          </button>
        </div>
      </header>

      <dl className="local-identity-security-scope">
        <div><dt>Current method</dt><dd>{authenticationMethodLabel(profile.session.authenticationMethod)}</dd></div>
        <div><dt>Session owner</dt><dd><code>{shortReference(profile.session.sessionId)}</code></dd></div>
        <div><dt>Recent authentication</dt><dd>{profile.capabilities.recentAuthentication ? "verified" : "required"}</dd></div>
        <div><dt>Directory snapshot</dt><dd>{directory.status === "ready" || directory.status === "empty" ? formatDate(directory.snapshotAt) : "pending"}</dd></div>
      </dl>

      {accountActionError ? <p className="local-identity-inline-error" role="alert">{accountActionError}</p> : null}

      {operation.status === "success" ? (
        <div className="local-identity-security-operation is-success" role="status">
          <span aria-hidden="true">✓</span><div><strong>Canonical change committed</strong><p>{operation.message}</p></div>
        </div>
      ) : operation.status === "failed" ? (
        <SecurityFailureNotice failure={operation.failure} onRetry={() => void loadSessions()} />
      ) : null}

      <div className="local-identity-security-workbench">
        <main className="local-identity-session-directory">
          <header>
            <div><p className="eyebrow">Single session owner</p><h3>Session directory</h3></div>
            <span>{sessions.length} loaded{nextCursor ? " · more available" : " · complete window"}</span>
          </header>

          {directory.status === "loading" ? (
            <SecurityDirectoryState title="Reading canonical sessions" message="No retained page or fixture is shown while the owner is loading." />
          ) : directory.status === "failed" ? (
            <SecurityFailureNotice failure={directory.failure} onRetry={() => void loadSessions()} />
          ) : directory.status === "empty" ? (
            <SecurityDirectoryState title="No session rows returned" message="The account remains authenticated, but no self-service session projection is available. No local result is invented." />
          ) : projection ? (
            <>
              <SessionGroup title="Current session" count={projection.currentSession ? 1 : 0}>
                {projection.currentSession ? (
                  <SessionRow
                    session={projection.currentSession}
                    selected={selectedSessionId === projection.currentSession.sessionId}
                    onSelect={() => setSelectedSessionId(projection.currentSession?.sessionId ?? "")}
                    onReview={() => reviewExactRevocation(projection.currentSession as LocalIdentitySelfServiceSessionSummary)}
                    disabled={busy}
                  />
                ) : (
                  <p className="local-identity-security-empty">Current session is outside the loaded window. Load the remaining canonical page before a session mutation.</p>
                )}
              </SessionGroup>

              <SessionGroup title="Other active sessions" count={projection.otherActiveSessions.length}>
                {projection.otherActiveSessions.length > 0 ? projection.otherActiveSessions.map((session) => (
                  <SessionRow
                    key={session.sessionId}
                    session={session}
                    selected={selectedSessionId === session.sessionId}
                    onSelect={() => setSelectedSessionId(session.sessionId)}
                    onReview={() => reviewExactRevocation(session)}
                    disabled={busy}
                  />
                )) : <p className="local-identity-security-empty">No other active sessions in this snapshot.</p>}
              </SessionGroup>

              <details className="local-identity-ended-sessions">
                <summary>Ended history <span>{projection.endedSessions.length}</span></summary>
                <div>
                  {projection.endedSessions.length > 0 ? projection.endedSessions.map((session) => (
                    <SessionRow
                      key={session.sessionId}
                      session={session}
                      selected={selectedSessionId === session.sessionId}
                      onSelect={() => setSelectedSessionId(session.sessionId)}
                      disabled
                    />
                  )) : <p className="local-identity-security-empty">No expired or revoked sessions in this snapshot.</p>}
                </div>
              </details>

              {nextCursor ? (
                <button type="button" className="local-identity-security-load-more" onClick={() => void loadSessions(nextCursor)} disabled={busy || loadingMore}>
                  {loadingMore ? "Loading next snapshot page…" : "Load remaining sessions"}
                </button>
              ) : null}
            </>
          ) : (
            <SecurityDirectoryState title="Invalid session projection" message="The page was rejected because its current-session relationship was inconsistent." />
          )}
        </main>

        <aside className="local-identity-security-actions" aria-label="Credential and bulk session actions">
          <details className="local-identity-credential-disclosure" open>
            <summary>
              <span><small>Subordinate security action</small><strong>Rotate local credential</strong></span>
              <em>{profile.capabilities.hasActiveLocalCredential ? "available" : "unavailable"}</em>
            </summary>
            <form onSubmit={reviewCredentialRotation} autoComplete="off">
              <p>
                Replaces the active local credential and atomically revokes every active session created from it.
                OIDC sessions remain active.
              </p>
              <label>
                <span>Current password</span>
                <input type="password" autoComplete="current-password" value={currentPassword} onChange={(event) => setCurrentPassword(event.target.value)} maxLength={1024} disabled={busy || !profile.capabilities.hasActiveLocalCredential} />
              </label>
              <label>
                <span>Replacement password</span>
                <input type="password" autoComplete="new-password" value={newPassword} onChange={(event) => setNewPassword(event.target.value)} minLength={12} maxLength={1024} disabled={busy || !profile.capabilities.hasActiveLocalCredential} />
                <small>12–1024 characters; the server applies the canonical policy.</small>
              </label>
              <label>
                <span>Confirm replacement</span>
                <input type="password" autoComplete="new-password" value={newPasswordConfirmation} onChange={(event) => setNewPasswordConfirmation(event.target.value)} minLength={12} maxLength={1024} disabled={busy || !profile.capabilities.hasActiveLocalCredential} />
              </label>
              <label className="local-identity-security-check">
                <input type="checkbox" checked={credentialImpactConfirmed} onChange={(event) => setCredentialImpactConfirmed(event.target.checked)} disabled={busy || !profile.capabilities.hasActiveLocalCredential} />
                <span>I understand the source-bound session impact.</span>
              </label>
              {credentialInputError ? <p className="local-identity-inline-error" role="alert">{credentialInputError}</p> : null}
              {nextCursor ? <small>Load the complete directory before credential rotation so the danger state can show its exact loaded impact set.</small> : null}
              <button type="submit" className="local-identity-security-danger" disabled={busy || nextCursor !== "" || !profile.capabilities.hasActiveLocalCredential}>
                Review credential rotation
              </button>
            </form>
          </details>

          <section className="local-identity-bulk-revoke">
            <header><small>Keep current session</small><strong>Revoke other active sessions</strong></header>
            <p>The server performs one aggregate mutation. No client-side revoke loop is used.</p>
            <ul>
              {(projection?.otherActiveSessions ?? []).map((session) => <li key={session.sessionId}><code>{session.sessionId}</code></li>)}
            </ul>
            {nextCursor ? <small>Load the complete directory before reviewing the exact target set.</small> : null}
            <button type="button" onClick={reviewBulkRevocation} disabled={busy || nextCursor !== "" || (projection?.otherActiveSessions.length ?? 0) === 0}>
              Review revoke others
            </button>
          </section>

          <p className="local-identity-security-boundary">
            No device fingerprint, IP, raw User-Agent, upstream token, password, or session directory payload is persisted in the browser.
          </p>
        </aside>
      </div>

      {confirmation.kind !== "none" ? (
        <SecurityConfirmation
          confirmation={confirmation}
          selectedSession={selectedSession}
          profile={profile}
          localPasswordTargets={projection?.activeLocalPasswordSessions ?? []}
          bulkTargets={projection?.otherActiveSessions ?? []}
          pending={operation.status === "pending"}
          onCancel={cancelConfirmation}
          onCommitExact={() => void commitExactRevocation()}
          onCommitBulk={() => void commitBulkRevocation()}
          onCommitCredential={() => void commitCredentialRotation()}
        />
      ) : null}
    </section>
  );
}

function SessionGroup({ title, count, children }: { title: string; count: number; children: ReactNode }) {
  return (
    <section className="local-identity-session-group">
      <header><strong>{title}</strong><span>{count}</span></header>
      <div>{children}</div>
    </section>
  );
}

function SessionRow({
  session,
  selected,
  onSelect,
  onReview,
  disabled,
}: {
  session: LocalIdentitySelfServiceSessionSummary;
  selected: boolean;
  onSelect: () => void;
  onReview?: () => void;
  disabled: boolean;
}) {
  return (
    <article className={`local-identity-session-row${selected ? " is-selected" : ""}${session.currentSession ? " is-current" : ""}`}>
      <button type="button" onClick={onSelect} disabled={disabled} aria-pressed={selected}>
        <span className={`local-identity-session-method is-${session.authenticationMethod}`} aria-hidden="true">
          {session.authenticationMethod === "oidc" ? "R" : "L"}
        </span>
        <span>
          <strong>{session.currentSession ? "This browser session" : authenticationMethodLabel(session.authenticationMethod)}</strong>
          <small><code>{session.sessionId}</code></small>
        </span>
        <span><small>Last verified</small><strong>{formatDate(session.lastVerifiedAt)}</strong></span>
        <em className={`is-${session.effectiveState}`}>{session.effectiveState}</em>
      </button>
      {onReview ? <button type="button" className="local-identity-session-review" onClick={onReview} disabled={disabled}>Review revoke</button> : null}
    </article>
  );
}

function SecurityConfirmation({
  confirmation,
  selectedSession,
  profile,
  localPasswordTargets,
  bulkTargets,
  pending,
  onCancel,
  onCommitExact,
  onCommitBulk,
  onCommitCredential,
}: {
  confirmation: Exclude<ConfirmationState, { kind: "none" }>;
  selectedSession: LocalIdentitySelfServiceSessionSummary | null;
  profile: LocalIdentityAccountProfile;
  localPasswordTargets: LocalIdentitySelfServiceSessionSummary[];
  bulkTargets: LocalIdentitySelfServiceSessionSummary[];
  pending: boolean;
  onCancel: () => void;
  onCommitExact: () => void;
  onCommitBulk: () => void;
  onCommitCredential: () => void;
}) {
  const credential = confirmation.kind === "credential";
  const exact = confirmation.kind === "exact";
  const targets = credential ? localPasswordTargets : confirmation.kind === "bulk" ? bulkTargets : selectedSession ? [selectedSession] : [];
  const currentWillClose = credential
    ? profile.session.authenticationMethod === "local_password"
    : exact && Boolean(selectedSession?.currentSession);
  const title = credential ? "Rotate credential and revoke source-bound sessions?"
    : exact ? "Revoke this exact session?" : "Revoke every other active session?";
  const commit = credential ? onCommitCredential : exact ? onCommitExact : onCommitBulk;
  return (
    <div className="local-identity-security-dialog-backdrop">
      <section className="local-identity-security-dialog" role="alertdialog" aria-modal="true" aria-labelledby="local-identity-security-confirmation-title">
        <header>
          <span aria-hidden="true">!</span>
          <div><p className="eyebrow">Explicit security confirmation</p><h3 id="local-identity-security-confirmation-title">{title}</h3></div>
        </header>
        <p>
          {currentWillClose
            ? "The current local session is in the exact revoke set. Success clears the authentication cookie and requires a new sign-in."
            : "The current session remains active. Any OIDC session outside the listed set remains active."}
        </p>
        <div className="local-identity-security-targets">
          <strong>Exact reviewed target set · {targets.length}</strong>
          {targets.length > 0 ? <ul>{targets.map((session) => <li key={session.sessionId}><code>{session.sessionId}</code><span>{authenticationMethodLabel(session.authenticationMethod)}</span></li>)}</ul>
            : <p>No active target is present in the canonical snapshot.</p>}
        </div>
        {credential ? <p className="local-identity-security-atomicity">If credential replacement or any revoke fails, the old credential and every session remain unchanged.</p> : null}
        <footer>
          <button type="button" onClick={onCancel} disabled={pending}>Cancel and clear input</button>
          <button type="button" className="local-identity-security-danger" onClick={commit} disabled={pending || targets.length === 0}>
            {pending ? "Committing atomically…" : credential ? "Rotate and revoke" : exact ? "Revoke exact session" : "Revoke other sessions"}
          </button>
        </footer>
      </section>
    </div>
  );
}

function SecurityFailureNotice({ failure, onRetry }: { failure: SecurityFailure; onRetry: () => void }) {
  return (
    <div className={`local-identity-security-operation is-${failure.kind}`} role="alert">
      <span aria-hidden="true">!</span>
      <div><strong>{failure.title}</strong><p>{failure.message}</p><small>{failure.code}</small></div>
      <button type="button" onClick={onRetry}>Reload canonical directory</button>
    </div>
  );
}

function SecurityDirectoryState({ title, message }: { title: string; message: string }) {
  return <div className="local-identity-security-state"><span aria-hidden="true">○</span><div><strong>{title}</strong><p>{message}</p></div></div>;
}

function securityFailure(error: unknown): SecurityFailure {
  const kind = localIdentitySelfServiceSecurityFailureKind(error);
  const code = error instanceof LocalIdentitySelfServiceSecurityError ? error.code : "local_identity_request_failed";
  const copy: Record<LocalIdentitySelfServiceSecurityFailureKind, { title: string; message: string }> = {
    authentication_required: { title: "Authentication required", message: "The current session no longer authorizes this security owner. Sign in again." },
    denied: { title: "Session scope denied", message: "The actor, session, Origin, or CSRF scope changed. No mutation was applied." },
    recent_authentication: { title: "Recent authentication required", message: "Sign out and authenticate again before retrying this security action." },
    conflict: { title: "Canonical state changed", message: "The session or credential changed before commit. Reload before making another decision." },
    credential_unavailable: { title: "Local credential unavailable", message: "This account has no active local credential. No login method was created implicitly." },
    credential_invalid: { title: "Credential proof rejected", message: "The current password was invalid or the replacement reused it. All prior state remains unchanged." },
    credential_policy: { title: "Credential policy rejected", message: "The replacement did not satisfy the canonical server policy. Re-enter both password fields." },
    unavailable: { title: "Security owner unavailable", message: "The operation failed closed. No memory, fixture, or retained result is substituted." },
    invalid_response: { title: "Invalid security response", message: "The response did not match the canonical schema or privacy boundary and was rejected." },
    failed: { title: "Security request failed", message: "The operation could not be verified. No mutation result is assumed." },
  };
  return { kind, code, ...copy[kind] };
}

function invalidSelectionFailure(): SecurityFailure {
  return {
    kind: "conflict",
    code: "local_identity_session_selection_stale",
    title: "Selected session changed",
    message: "The exact target is no longer active in this snapshot. Reload before making another decision.",
  };
}

function invalidCredentialReviewFailure(): SecurityFailure {
  return {
    kind: "conflict",
    code: "local_identity_credential_review_stale",
    title: "Credential review expired",
    message: "The in-memory credential input was cleared before commit. Re-enter both password fields.",
  };
}

function invalidDirectory(message: string): LocalIdentitySelfServiceSecurityError {
  return new LocalIdentitySelfServiceSecurityError(0, "local_identity_response_invalid", message);
}

function authenticationMethodLabel(method: "local_password" | "oidc"): string {
  return method === "oidc" ? "Radish OIDC" : "Local password";
}

function shortReference(value: string): string {
  return value.length > 24 ? `${value.slice(0, 12)}…${value.slice(-7)}` : value;
}

function formatDate(value: string): string {
  return new Intl.DateTimeFormat("en-GB", {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: "UTC",
  }).format(new Date(value)) + " UTC";
}

function isAbort(error: unknown): boolean {
  return error instanceof DOMException && error.name === "AbortError";
}

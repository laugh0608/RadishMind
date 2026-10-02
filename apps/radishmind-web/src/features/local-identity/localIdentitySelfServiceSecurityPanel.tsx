import { useTranslation } from "react-i18next";
import "../../i18n/identitySecurityResources.ts";
import { LanguageSelector } from "../../i18n/LanguageSelector.tsx";
import { formatDisplayNumber } from "../../i18n/formatters.ts";
import { identityFailureMessage } from "./localIdentityMessages.ts";
import {
  securityFailure, securityFailureCopy, securitySuccessMessage, securityDate,
  type SecurityFailure, type SecuritySuccess, type CredentialInputError,
} from "./localIdentitySecurityMessages.ts";
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

type OperationState =
  | { status: "idle" }
  | { status: "pending"; action: "exact" | "bulk" | "credential" }
  | { status: "success"; feedback: SecuritySuccess }
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
  onClaimInvitation,
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
  onClaimInvitation: () => void;
  onLinkOIDC: () => Promise<void>;
  onLogout: () => Promise<void>;
  accountAction: "" | "link" | "logout" | "revoke";
  accountActionError: string;
}) {
  const { t, i18n } = useTranslation("identity");
  const locale = i18n.language === "en-US" ? "en-US" : "zh-CN";
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
  const [credentialInputError, setCredentialInputError] = useState<CredentialInputError | "">("");

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
      setCredentialInputError("unavailable");
      return;
    }
    if (!profile.capabilities.recentAuthentication) {
      setCredentialInputError("recent");
      return;
    }
    if (nextCursor !== "") {
      setCredentialInputError("incomplete");
      return;
    }
    if (currentPassword.length < 1 || currentPassword.length > 1024 ||
      newPassword.length < 12 || newPassword.length > 1024) {
      setCredentialInputError("range");
      return;
    }
    if (newPassword !== newPasswordConfirmation) {
      setCredentialInputError("mismatch");
      return;
    }
    if (!credentialImpactConfirmed) {
      setCredentialInputError("impact");
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
        setOperation({ status: "success", feedback: { kind: "currentRevoked" } });
        onAuthenticationRequired();
        return;
      }
      setOperation({ status: "success", feedback: { kind: "exactRevoked" } });
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
        feedback: { kind: "othersRevoked", count: result.revokedCount },
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
          feedback: { kind: "credentialClosed", policyVersion: result.policyVersion },
        });
        onAuthenticationRequired();
        return;
      }
      setOperation({
        status: "success",
        feedback: { kind: "credentialRevoked", count: result.revokedSessionCount },
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
          <p className="eyebrow">{t($ => $.security.eyebrow)}</p>
          <h2 id="local-identity-security-title">{t($ => $.security.title)}</h2>
          <p>{profile.account.displayName} · <code>{profile.account.userId}</code></p>
        </div>
        <div className="local-identity-security-heading-actions">
          <LanguageSelector />
          <button type="button" onClick={onClaimInvitation} disabled={busy}>{t($ => $.security.claim)}</button>
          <button type="button" onClick={() => void onLinkOIDC()} disabled={busy || !profile.capabilities.oidcEnabled || !profile.capabilities.recentAuthentication}>
            {accountAction === "link" ? t($ => $.security.opening) : t($ => $.security.link)}
          </button>
          <button type="button" onClick={() => void onLogout()} disabled={busy}>
            {accountAction === "logout" ? t($ => $.security.signingOut) : t($ => $.security.signOut)}
          </button>
          <button type="button" className="local-identity-security-close" onClick={onClose} disabled={busy} aria-label={t($ => $.security.close)}>
            ×
          </button>
        </div>
      </header>

      <dl className="local-identity-security-scope">
        <div><dt>{t($ => $.security.currentMethod)}</dt><dd>{t($ => $.security.methods[profile.session.authenticationMethod])}</dd></div>
        <div><dt>{t($ => $.security.sessionOwner)}</dt><dd><code>{shortReference(profile.session.sessionId)}</code></dd></div>
        <div><dt>{t($ => $.security.recentAuthentication)}</dt><dd>{profile.capabilities.recentAuthentication ? t($ => $.security.verified) : t($ => $.security.required)}</dd></div>
        <div><dt>{t($ => $.security.snapshot)}</dt><dd title={directory.status === "ready" || directory.status === "empty" ? directory.snapshotAt : undefined}>{directory.status === "ready" || directory.status === "empty" ? securityDate(t, directory.snapshotAt, locale) : t($ => $.security.pending)}</dd></div>
      </dl>

      {accountActionError ? <p className="local-identity-inline-error" role="alert">{identityFailureMessage(t, accountActionError)} <code>{accountActionError}</code></p> : null}

      {operation.status === "success" ? (
        <div className="local-identity-security-operation is-success" role="status">
          <span aria-hidden="true">✓</span><div><strong>{t($ => $.security.committed)}</strong><p>{securitySuccessMessage(t, operation.feedback, locale)}</p></div>
        </div>
      ) : operation.status === "pending" ? (
        <p role="status">{t($ => $.security.committing)}</p>
      ) : operation.status === "failed" ? (
        <SecurityFailureNotice failure={operation.failure} onRetry={() => void loadSessions()} />
      ) : null}

      <div className="local-identity-security-workbench">
        <main className="local-identity-session-directory">
          <header>
            <div><p className="eyebrow">{t($ => $.security.owner)}</p><h3>{t($ => $.security.directory)}</h3></div>
            <span>{nextCursor ? t($ => $.security.loadedMore, { countText: formatDisplayNumber(sessions.length, locale) ?? t($ => $.security.unknown) }) : t($ => $.security.loadedComplete, { countText: formatDisplayNumber(sessions.length, locale) ?? t($ => $.security.unknown) })}</span>
          </header>

          {directory.status === "loading" ? (
            <SecurityDirectoryState title={t($ => $.security.loadingTitle)} message={t($ => $.security.loadingDescription)} />
          ) : directory.status === "failed" ? (
            <SecurityFailureNotice failure={directory.failure} onRetry={() => void loadSessions()} />
          ) : directory.status === "empty" ? (
            <SecurityDirectoryState title={t($ => $.security.emptyTitle)} message={t($ => $.security.emptyDescription)} />
          ) : projection ? (
            <>
              <SessionGroup title={t($ => $.security.currentSession)} count={projection.currentSession ? 1 : 0}>
                {projection.currentSession ? (
                  <SessionRow
                    session={projection.currentSession}
                    selected={selectedSessionId === projection.currentSession.sessionId}
                    onSelect={() => setSelectedSessionId(projection.currentSession?.sessionId ?? "")}
                    onReview={() => reviewExactRevocation(projection.currentSession as LocalIdentitySelfServiceSessionSummary)}
                    disabled={busy}
                  />
                ) : (
                  <p className="local-identity-security-empty">{t($ => $.security.outsideWindow)}</p>
                )}
              </SessionGroup>

              <SessionGroup title={t($ => $.security.otherSessions)} count={projection.otherActiveSessions.length}>
                {projection.otherActiveSessions.length > 0 ? projection.otherActiveSessions.map((session) => (
                  <SessionRow
                    key={session.sessionId}
                    session={session}
                    selected={selectedSessionId === session.sessionId}
                    onSelect={() => setSelectedSessionId(session.sessionId)}
                    onReview={() => reviewExactRevocation(session)}
                    disabled={busy}
                  />
                )) : <p className="local-identity-security-empty">{t($ => $.security.noOtherSessions)}</p>}
              </SessionGroup>

              <details className="local-identity-ended-sessions">
                <summary>{t($ => $.security.endedHistory)} <span>{formatDisplayNumber(projection.endedSessions.length, locale)}</span></summary>
                <div>
                  {projection.endedSessions.length > 0 ? projection.endedSessions.map((session) => (
                    <SessionRow
                      key={session.sessionId}
                      session={session}
                      selected={selectedSessionId === session.sessionId}
                      onSelect={() => setSelectedSessionId(session.sessionId)}
                      disabled
                    />
                  )) : <p className="local-identity-security-empty">{t($ => $.security.noEndedSessions)}</p>}
                </div>
              </details>

              {nextCursor ? (
                <button type="button" className="local-identity-security-load-more" onClick={() => void loadSessions(nextCursor)} disabled={busy || loadingMore}>
                  {loadingMore ? t($ => $.security.loadingMore) : t($ => $.security.loadMore)}
                </button>
              ) : null}
            </>
          ) : (
            <SecurityDirectoryState title={t($ => $.security.invalidTitle)} message={t($ => $.security.invalidDescription)} />
          )}
        </main>

        <aside className="local-identity-security-actions" aria-label={t($ => $.security.actions)}>
          <details className="local-identity-credential-disclosure" open>
            <summary>
              <span><small>{t($ => $.security.credentialAction)}</small><strong>{t($ => $.security.rotate)}</strong></span>
              <em>{profile.capabilities.hasActiveLocalCredential ? t($ => $.security.available) : t($ => $.security.unavailable)}</em>
            </summary>
            <form onSubmit={reviewCredentialRotation} autoComplete="off">
              <p>
                {t($ => $.security.rotationDescription)}
              </p>
              <label>
                <span>{t($ => $.security.currentPassword)}</span>
                <input type="password" autoComplete="current-password" value={currentPassword} onChange={(event) => setCurrentPassword(event.target.value)} maxLength={1024} disabled={busy || !profile.capabilities.hasActiveLocalCredential} />
              </label>
              <label>
                <span id="local-identity-new-password-label">{t($ => $.security.newPassword)}</span>
                <input type="password" autoComplete="new-password" aria-labelledby="local-identity-new-password-label" aria-describedby="local-identity-new-password-help" value={newPassword} onChange={(event) => setNewPassword(event.target.value)} minLength={12} maxLength={1024} disabled={busy || !profile.capabilities.hasActiveLocalCredential} />
                <small id="local-identity-new-password-help">{t($ => $.security.passwordHelp)}</small>
              </label>
              <label>
                <span>{t($ => $.security.confirmPassword)}</span>
                <input type="password" autoComplete="new-password" value={newPasswordConfirmation} onChange={(event) => setNewPasswordConfirmation(event.target.value)} minLength={12} maxLength={1024} disabled={busy || !profile.capabilities.hasActiveLocalCredential} />
              </label>
              <label className="local-identity-security-check">
                <input type="checkbox" checked={credentialImpactConfirmed} onChange={(event) => setCredentialImpactConfirmed(event.target.checked)} disabled={busy || !profile.capabilities.hasActiveLocalCredential} />
                <span>{t($ => $.security.impactConsent)}</span>
              </label>
              {credentialInputError ? <p className="local-identity-inline-error" role="alert">{t($ => $.security.inputErrors[credentialInputError])}</p> : null}
              {nextCursor ? <small>{t($ => $.security.loadBeforeRotation)}</small> : null}
              <button type="submit" className="local-identity-security-danger" disabled={busy || nextCursor !== "" || !profile.capabilities.hasActiveLocalCredential}>
                {t($ => $.security.reviewRotation)}
              </button>
            </form>
          </details>

          <section className="local-identity-bulk-revoke">
            <header><small>{t($ => $.security.keepCurrent)}</small><strong>{t($ => $.security.revokeOthers)}</strong></header>
            <p>{t($ => $.security.bulkDescription)}</p>
            <ul>
              {(projection?.otherActiveSessions ?? []).map((session) => <li key={session.sessionId}><code>{session.sessionId}</code></li>)}
            </ul>
            {nextCursor ? <small>{t($ => $.security.loadBeforeBulk)}</small> : null}
            <button type="button" onClick={reviewBulkRevocation} disabled={busy || nextCursor !== "" || (projection?.otherActiveSessions.length ?? 0) === 0}>
              {t($ => $.security.reviewOthers)}
            </button>
          </section>

          <p className="local-identity-security-boundary">
            {t($ => $.security.privacyBoundary)}
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
  const { i18n } = useTranslation("identity");
  return (
    <section className="local-identity-session-group">
      <header><strong>{title}</strong><span>{formatDisplayNumber(count, i18n.language === "en-US" ? "en-US" : "zh-CN")}</span></header>
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
  const { t, i18n } = useTranslation("identity");
  const locale = i18n.language === "en-US" ? "en-US" : "zh-CN";
  return (
    <article className={`local-identity-session-row${selected ? " is-selected" : ""}${session.currentSession ? " is-current" : ""}`}>
      <button type="button" onClick={onSelect} disabled={disabled} aria-pressed={selected}>
        <span className={`local-identity-session-method is-${session.authenticationMethod}`} aria-hidden="true">
          {session.authenticationMethod === "oidc" ? "R" : "L"}
        </span>
        <span>
          <strong>{session.currentSession ? t($ => $.security.thisSession) : t($ => $.security.methods[session.authenticationMethod])}</strong>
          <small><code>{session.sessionId}</code></small>
        </span>
        <span><small>{t($ => $.security.lastVerified)}</small><strong title={session.lastVerifiedAt}>{securityDate(t, session.lastVerifiedAt, locale)}</strong></span>
        <em className={`is-${session.effectiveState}`}>{t($ => $.security.states[session.effectiveState])}</em>
      </button>
      {onReview ? <button type="button" className="local-identity-session-review" onClick={onReview} disabled={disabled}>{t($ => $.security.reviewRevoke)}</button> : null}
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
  const { t, i18n } = useTranslation("identity");
  const locale = i18n.language === "en-US" ? "en-US" : "zh-CN";
  const credential = confirmation.kind === "credential";
  const exact = confirmation.kind === "exact";
  const targets = credential ? localPasswordTargets : confirmation.kind === "bulk" ? bulkTargets : selectedSession ? [selectedSession] : [];
  const currentWillClose = credential
    ? profile.session.authenticationMethod === "local_password"
    : exact && Boolean(selectedSession?.currentSession);
  const title = credential ? t($ => $.security.confirmRotation)
    : exact ? t($ => $.security.confirmExact) : t($ => $.security.confirmBulk);
  const commit = credential ? onCommitCredential : exact ? onCommitExact : onCommitBulk;
  return (
    <div className="local-identity-security-dialog-backdrop">
      <section className="local-identity-security-dialog" role="alertdialog" aria-modal="true" aria-labelledby="local-identity-security-confirmation-title">
        <header>
          <span aria-hidden="true">!</span>
          <div><p className="eyebrow">{t($ => $.security.confirmation)}</p><h3 id="local-identity-security-confirmation-title">{title}</h3><LanguageSelector /></div>
        </header>
        <p>
          {currentWillClose
            ? t($ => $.security.currentWillClose)
            : t($ => $.security.currentWillRemain)}
        </p>
        <div className="local-identity-security-targets">
          <strong>{t($ => $.security.targetSet, { countText: formatDisplayNumber(targets.length, locale) ?? t($ => $.security.unknown) })}</strong>
          {targets.length > 0 ? <ul>{targets.map((session) => <li key={session.sessionId}><code>{session.sessionId}</code><span>{t($ => $.security.methods[session.authenticationMethod])}</span></li>)}</ul>
            : <p>{t($ => $.security.noTargets)}</p>}
        </div>
        {credential ? <p className="local-identity-security-atomicity">{t($ => $.security.atomicity)}</p> : null}
        <footer>
          <button type="button" onClick={onCancel} disabled={pending}>{t($ => $.security.cancel)}</button>
          <button type="button" className="local-identity-security-danger" onClick={commit} disabled={pending || targets.length === 0}>
            {pending ? t($ => $.security.committing) : credential ? t($ => $.security.rotateAndRevoke) : exact ? t($ => $.security.revokeExact) : t($ => $.security.commitBulk)}
          </button>
        </footer>
      </section>
    </div>
  );
}

function SecurityFailureNotice({ failure, onRetry }: { failure: SecurityFailure; onRetry: () => void }) {
  const { t } = useTranslation("identity");
  const copy = securityFailureCopy(t, failure);
  return (
    <div className={`local-identity-security-operation is-${failure.kind}`} role="alert">
      <span aria-hidden="true">!</span>
      <div><strong>{copy.title}</strong><p>{copy.message}</p><small>{failure.code}</small></div>
      <button type="button" onClick={onRetry}>{t($ => $.security.reload)}</button>
    </div>
  );
}

function SecurityDirectoryState({ title, message }: { title: string; message: string }) {
  return <div className="local-identity-security-state"><span aria-hidden="true">○</span><div><strong>{title}</strong><p>{message}</p></div></div>;
}

function invalidSelectionFailure(): SecurityFailure {
  return {
    kind: "conflict",
    code: "local_identity_session_selection_stale",
  };
}

function invalidCredentialReviewFailure(): SecurityFailure {
  return {
    kind: "conflict",
    code: "local_identity_credential_review_stale",
  };
}

function invalidDirectory(message: string): LocalIdentitySelfServiceSecurityError {
  return new LocalIdentitySelfServiceSecurityError(0, "local_identity_response_invalid", message);
}

function shortReference(value: string): string {
  return value.length > 24 ? `${value.slice(0, 12)}…${value.slice(-7)}` : value;
}

function isAbort(error: unknown): boolean {
  return error instanceof DOMException && error.name === "AbortError";
}

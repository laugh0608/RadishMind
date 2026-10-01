import { useTranslation } from "react-i18next";
import { LanguageSelector } from "../../i18n/LanguageSelector.tsx";
import {
  lazy,
  Suspense,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type FormEvent,
  type ReactNode,
} from "react";

import {
  LocalIdentityConsumerError,
  authenticateLocalIdentity,
  localIdentityReturnTarget,
  logoutLocalIdentity,
  probeLocalIdentitySession,
  readLocalIdentityAccountProfile,
  readLocalIdentityConsumerConfig,
  revokeLocalIdentityExternalIdentity,
  startLocalIdentityOIDC,
  type LocalIdentityAccountProfile,
  type LocalIdentityConsumerConfig,
} from "./localIdentityConsumer.ts";
import { LocalIdentitySelfServiceSecurityPanel } from "./localIdentitySelfServiceSecurityPanel.tsx";
import { localIdentitySelfServiceSecurityScopeKey } from "./localIdentitySelfServiceSecurityState.ts";
import { LocalIdentityContext, type LocalIdentityContextValue } from "./localIdentityContext.ts";
export { useLocalIdentity } from "./localIdentityContext.ts";

const WorkspaceInvitationClaimPanel = lazy(() => import("./workspaceInvitationClaimPanel.tsx").then((module) => ({ default: module.WorkspaceInvitationClaimPanel })));

type LocalIdentityGatewayState =
  | { status: "probing" }
  | { status: "unauthenticated" }
  | { status: "ready"; profile: LocalIdentityAccountProfile }
  | { status: "failed"; message: string; code: string };

export function LocalIdentityGateway({ children }: { children: ReactNode }) {
  const { t } = useTranslation("identity");
  const config = useMemo(() => readLocalIdentityConsumerConfig(), []);
  const [state, setState] = useState<LocalIdentityGatewayState>({ status: "probing" });
  const [accountPanelOpen, setAccountPanelOpen] = useState(false);
  const [accountTask, setAccountTask] = useState<"security" | "claim">("security");
  const [accountAction, setAccountAction] = useState<"" | "link" | "logout" | "revoke">("");
  const [accountActionError, setAccountActionError] = useState("");
  const [securityInvalidation, setSecurityInvalidation] = useState(0);
  const authorityRevision = useRef(0);
  const sessionGeneration = useRef(0);
  const sessionController = useRef<AbortController | null>(null);
  const currentActor = useRef("");
  const workspaceScope = useRef("");
  const readAuthorityRevision = useCallback(() => authorityRevision.current, []);
  const invalidateAuthority = useCallback(() => {
    authorityRevision.current += 1;
    setSecurityInvalidation(authorityRevision.current);
  }, []);
  const onWorkspaceScopeChange = useCallback((tenantRef: string, workspaceId: string) => {
    const scope = JSON.stringify([tenantRef, workspaceId]);
    if (workspaceScope.current === scope) return;
    workspaceScope.current = scope;
    invalidateAuthority();
  }, [invalidateAuthority]);

  const refresh = useCallback(async () => {
    if (config.mode !== "local_identity_dev") return;
    const generation = ++sessionGeneration.current;
    sessionController.current?.abort();
    const controller = new AbortController();
    sessionController.current = controller;
    const isCurrent = () => !controller.signal.aborted && sessionGeneration.current === generation;
    try {
      const authentication = await probeLocalIdentitySession(config, controller.signal);
      if (!isCurrent()) return;
      if (!authentication) {
        invalidateAuthority();
        currentActor.current = "";
        setState({ status: "unauthenticated" });
        return;
      }
      const profile = await readLocalIdentityAccountProfile(config, controller.signal);
      if (!isCurrent()) return;
      const actor = JSON.stringify([profile.account.userId, profile.account.lifecycleState, profile.session,
        profile.capabilities.recentAuthentication]);
      if (actor !== currentActor.current) { invalidateAuthority(); currentActor.current = actor; }
      setState({ status: "ready", profile });
    } catch (error) {
      if (!isCurrent()) return;
      invalidateAuthority();
      const failure = identityFailure(error);
      if (failure.code === "LOCAL_IDENTITY_AUTHENTICATION_REQUIRED") {
        setState({ status: "unauthenticated" });
        return;
      }
      setState({ status: "failed", ...failure });
    }
  }, [config, invalidateAuthority]);

  useEffect(() => {
    if (config.mode !== "local_identity_dev") return;
    void refresh();
    return () => { sessionGeneration.current += 1; sessionController.current?.abort(); };
  }, [config, refresh]);

  useEffect(() => {
    if (config.mode !== "local_identity_dev" || typeof BroadcastChannel === "undefined") return;
    const channel = new BroadcastChannel("radishmind-local-identity-v1");
    channel.onmessage = (event: MessageEvent<unknown>) => {
      if (isSessionChangedEvent(event.data)) {
        invalidateAuthority();
        setState({ status: "probing" });
        void refresh();
      }
    };
    return () => channel.close();
  }, [config.mode, refresh, invalidateAuthority]);

  useEffect(() => {
    if (!accountPanelOpen) return;
    const closeForRouteChange = () => {
      invalidateAuthority();
      setAccountPanelOpen(false);
      setAccountActionError("");
    };
    window.addEventListener("hashchange", closeForRouteChange);
    window.addEventListener("popstate", closeForRouteChange);
    return () => {
      window.removeEventListener("hashchange", closeForRouteChange);
      window.removeEventListener("popstate", closeForRouteChange);
    };
  }, [accountPanelOpen, invalidateAuthority]);

  if (config.mode !== "local_identity_dev") return <>{children}</>;
  if (state.status === "probing") return <LocalIdentityLoading />;
  if (state.status === "unauthenticated") {
    return <LocalIdentityAuthenticationSurface config={config} onAuthenticated={refresh} />;
  }
  if (state.status === "failed") {
    return <LocalIdentityFailureSurface failure={state} onRetry={refresh} />;
  }

  async function handleLinkOIDC() {
    setAccountAction("link");
    setAccountActionError("");
    try {
      const authorization = await startLocalIdentityOIDC(config, "link", localIdentityReturnTarget(window.location));
      window.location.assign(authorization.authorizationUrl);
    } catch (error) {
      setAccountActionError(identityFailure(error).message);
      setAccountAction("");
    }
  }

  async function handleRevokeExternalIdentity(bindingId: string, expectedRecordVersion: number) {
    setAccountAction("revoke");
    setAccountActionError("");
    try {
      await revokeLocalIdentityExternalIdentity(config, bindingId, expectedRecordVersion);
      await refresh();
    } catch (error) {
      setAccountActionError(identityFailure(error).message);
      throw error;
    } finally {
      setAccountAction("");
    }
  }

  async function handleLogout() {
    invalidateAuthority();
    sessionGeneration.current += 1;
    sessionController.current?.abort();
    setAccountAction("logout");
    setAccountActionError("");
    try {
      await logoutLocalIdentity(config);
      broadcastSessionChanged();
      currentActor.current = "";
      setAccountAction("");
      setAccountPanelOpen(false);
      setState({ status: "unauthenticated" });
    } catch (error) {
      setAccountActionError(identityFailure(error).message);
      setAccountAction("");
    }
  }

  function handleAuthenticationRequired() {
    invalidateAuthority();
    sessionGeneration.current += 1;
    sessionController.current?.abort();
    setAccountPanelOpen(false);
    setAccountAction("");
    setAccountActionError("");
    setState({ status: "unauthenticated" });
  }

  function handleSessionChanged() {
    invalidateAuthority();
    broadcastSessionChanged();
  }

  const contextValue: LocalIdentityContextValue = {
    config,
    profile: state.profile,
    authorityRevision: securityInvalidation,
    readAuthorityRevision,
    onWorkspaceScopeChange,
    refresh,
    linkOIDC: handleLinkOIDC,
    revokeExternalIdentity: handleRevokeExternalIdentity,
  };

  return (
    <LocalIdentityContext.Provider value={contextValue}>
      <div inert={accountPanelOpen && accountTask === "claim"}>{children}</div>
      <aside className="local-identity-account-control" aria-label="Local identity session">
        <button
          type="button"
          className="local-identity-account-trigger"
          aria-label={`${state.profile.account.displayName} ${state.profile.session.authenticationMethod === "oidc" ? "Radish OIDC" : "Local session"}`}
          aria-expanded={accountPanelOpen}
          onClick={() => {
            setAccountActionError("");
            setAccountTask("security");
            setAccountPanelOpen(true);
          }}
          disabled={accountPanelOpen}
        >
          <span aria-hidden="true">{state.profile.account.displayName.slice(0, 1).toUpperCase()}</span>
          <strong>{state.profile.account.displayName}</strong>
          <small>{state.profile.session.authenticationMethod === "oidc" ? "Radish OIDC" : t($ => $.localSession)}</small>
        </button>
        {accountPanelOpen && accountTask === "claim" ? (
          <Suspense fallback={<div className="local-identity-security-surface" role="status">Loading invitation claim…</div>}>
            <WorkspaceInvitationClaimPanel
              key={localIdentitySelfServiceSecurityScopeKey(state.profile, securityInvalidation)}
              identity={contextValue}
              onClose={() => { invalidateAuthority(); setAccountPanelOpen(false); }}
              onOpenSecurity={() => { invalidateAuthority(); setAccountTask("security"); }}
              onLogout={handleLogout}
            />
          </Suspense>
        ) : accountPanelOpen ? (
          <LocalIdentitySelfServiceSecurityPanel
            key={localIdentitySelfServiceSecurityScopeKey(state.profile, securityInvalidation)}
            config={config}
            profile={state.profile}
            onClose={() => {
              invalidateAuthority();
              setAccountPanelOpen(false);
              setAccountActionError("");
            }}
            onRefreshProfile={refresh}
            onAuthenticationRequired={handleAuthenticationRequired}
            onSessionChanged={handleSessionChanged}
            onClaimInvitation={() => { invalidateAuthority(); setAccountTask("claim"); }}
            onLinkOIDC={handleLinkOIDC}
            onLogout={handleLogout}
            accountAction={accountAction}
            accountActionError={accountActionError}
          />
        ) : null}
      </aside>
    </LocalIdentityContext.Provider>
  );
}

function LocalIdentityAuthenticationSurface({
  config,
  onAuthenticated,
}: {
  config: LocalIdentityConsumerConfig;
  onAuthenticated: () => Promise<void>;
}) {
  const { t } = useTranslation("identity");
  const [intent, setIntent] = useState<"login" | "register">("login");
  const [loginIdentifier, setLoginIdentifier] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState<"" | "password" | "oidc">("");
  const [error, setError] = useState("");

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy("password");
    setError("");
    try {
      await authenticateLocalIdentity(config, {
        intent,
        loginIdentifier,
        displayName,
        password,
        returnTo: localIdentityReturnTarget(window.location),
      });
      setPassword("");
      broadcastSessionChanged();
      await onAuthenticated();
    } catch (submitError) {
      setPassword("");
      setError(identityFailure(submitError).code);
    } finally {
      setBusy("");
    }
  }

  async function handleOIDCLogin() {
    setBusy("oidc");
    setError("");
    try {
      const authorization = await startLocalIdentityOIDC(config, "login", localIdentityReturnTarget(window.location));
      window.location.assign(authorization.authorizationUrl);
    } catch (oidcError) {
      setError(identityFailure(oidcError).code);
      setBusy("");
    }
  }

  return (
    <main className="local-identity-gateway" data-rd-profile="workbench">
      <header className="local-identity-gateway-header">
        <a href="/" className="local-identity-wordmark" aria-label={t($ => $.home)}>
          <span aria-hidden="true">R</span><strong>RadishMind</strong>
        </a>
        <em>{t($ => $.development)}</em>
        <LanguageSelector />
      </header>
      <div className="local-identity-gateway-layout">
        <aside className="local-identity-trust-rail" aria-label={t($ => $.trustBoundary)}>
          <p className="eyebrow">{t($ => $.trustBoundary)}</p>
          <h1>{t($ => $.oneAccount)}<br />{t($ => $.explicitTrust)}</h1>
          <p>
            {t($ => $.ownership)}</p>
          <ol>
            <li><span>01</span><div><strong>{t($ => $.localSession)}</strong><small>{t($ => $.cookieBoundary)}</small></div></li>
            <li><span>02</span><div><strong>{t($ => $.binding)}</strong><small>{t($ => $.bindingDescription)}</small></div></li>
            <li><span>03</span><div><strong>{t($ => $.grants)}</strong><small>{t($ => $.grantBoundary)}</small></div></li>
          </ol>
          <div className="local-identity-stop-line">
            <strong>{t($ => $.stopLine)}</strong>
            <p>{t($ => $.developmentBoundary)}</p>
          </div>
        </aside>

        <section className="local-identity-form-zone" aria-labelledby="local-identity-form-title">
          <form className="local-identity-card" onSubmit={(event) => void handleSubmit(event)}>
            <header>
              <p className="eyebrow">{t($ => $.secureAccess)}</p>
              <h2 id="local-identity-form-title">{intent === "login" ? t($ => $.welcome) : t($ => $.createAccountTitle)}</h2>
              <p>{intent === "login" ? t($ => $.loginDescription) : t($ => $.registerDescription)}</p>
            </header>
            <div className="local-identity-tabs" role="tablist" aria-label={t($ => $.authenticationMode)}>
              <button type="button" role="tab" aria-selected={intent === "login"} onClick={() => { setIntent("login"); setError(""); }}>{t($ => $.signIn)}</button>
              <button type="button" role="tab" aria-selected={intent === "register"} onClick={() => { setIntent("register"); setError(""); }}>{t($ => $.register)}</button>
            </div>
            {intent === "register" ? (
              <label>
                <span>{t($ => $.displayName)}</span>
                <input autoComplete="name" value={displayName} onChange={(event) => setDisplayName(event.target.value)} required maxLength={120} />
              </label>
            ) : null}
            <label>
              <span>{t($ => $.localId)}</span>
              <input autoComplete="username" value={loginIdentifier} onChange={(event) => setLoginIdentifier(event.target.value)} required maxLength={254} />
            </label>
            <label>
              <span>{t($ => $.password)}</span>
              <input
                type="password"
                autoComplete={intent === "login" ? "current-password" : "new-password"}
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                required
                minLength={12}
                maxLength={1024}
              />
              <small>{t($ => $.passwordHelp)}</small>
            </label>
            {error ? <p className="local-identity-form-error" role="alert"><IdentityFailureMessage code={error} /></p> : null}
            <button type="submit" className="local-identity-primary-action" disabled={busy !== ""}>
              {busy === "password" ? t($ => $.verifying) : intent === "login" ? t($ => $.signInSecurely) : t($ => $.createAccount)}
            </button>
            <div className="local-identity-divider"><span>{t($ => $.or)}</span></div>
            <button type="button" className="local-identity-oidc-action" onClick={() => void handleOIDCLogin()} disabled={busy !== ""}>
              <span aria-hidden="true">R</span>{busy === "oidc" ? t($ => $.openingRadish) : t($ => $.continueRadish)}
            </button>
            <p className="local-identity-recovery-note">
              {t($ => $.recoveryHelp)}</p>
          </form>
          <p className="local-identity-narrow-boundary">
            {t($ => $.sessionBoundary)}</p>
        </section>
      </div>
    </main>
  );
}

function LocalIdentityLoading() {
  const { t } = useTranslation("identity");
  return (
    <main className="local-identity-status-surface" data-rd-profile="workbench">
      <LanguageSelector />
      <div><span className="local-identity-status-mark" aria-hidden="true">R</span><p>{t($ => $.restoring)}</p></div>
    </main>
  );
}

function LocalIdentityFailureSurface({
  failure,
  onRetry,
}: {
  failure: { message: string; code: string };
  onRetry: () => Promise<void>;
}) {
  const { t } = useTranslation("identity");
  return (
    <main className="local-identity-status-surface" data-rd-profile="workbench">
      <LanguageSelector />
      <div className="local-identity-failure-card">
        <p className="eyebrow">{t($ => $.unavailable)}</p>
        <h1>{t($ => $.accessClosed)}</h1>
        <p><IdentityFailureMessage code={failure.code} /></p>
        <small>{failure.code}</small>
        <button type="button" onClick={() => void onRetry()}>{t($ => $.retrySession)}</button>
        <p className="local-identity-account-boundary">{t($ => $.noFallback)}</p>
      </div>
    </main>
  );
}

function identityFailure(error: unknown): { message: string; code: string } {
  if (error instanceof LocalIdentityConsumerError) {
    const guidance = error.code === "LOCAL_IDENTITY_AUTHENTICATION_FAILED"
      ? "The local ID or password is invalid. Disabled accounts receive the same response."
      : error.code === "LOCAL_IDENTITY_ACCOUNT_CHANGE_REQUIRES_RECENT_AUTHENTICATION"
      ? "Sign out and authenticate again before changing a login method."
      : error.code === "LOCAL_IDENTITY_LAST_LOGIN_METHOD_REMOVAL_DENIED"
      ? "Keep at least one active local credential or external identity."
      : error.message;
    return { message: guidance, code: error.code };
  }
  return { message: "The local identity service could not be verified.", code: "local_identity_unavailable" };
}

function broadcastSessionChanged(): void {
  if (typeof BroadcastChannel === "undefined") return;
  const channel = new BroadcastChannel("radishmind-local-identity-v1");
  channel.postMessage({ kind: "session_changed", version: 1 });
  channel.close();
}

function isSessionChangedEvent(value: unknown): boolean {
  if (!value || typeof value !== "object" || Array.isArray(value)) return false;
  const record = value as Record<string, unknown>;
  return Object.keys(record).length === 2 && record.kind === "session_changed" && record.version === 1;
}

function IdentityFailureMessage({ code }: { code: string }) {
  const { t } = useTranslation("identity");
  const message = code === "LOCAL_IDENTITY_AUTHENTICATION_FAILED" ? t($ => $.authenticationFailed)
    : code === "LOCAL_IDENTITY_ACCOUNT_CHANGE_REQUIRES_RECENT_AUTHENTICATION" ? t($ => $.recentAuthentication)
    : code === "LOCAL_IDENTITY_LAST_LOGIN_METHOD_REMOVAL_DENIED" ? t($ => $.keepLoginMethod)
    : t($ => $.serviceFailure);
  return <>{message} <code>{code}</code></>;
}

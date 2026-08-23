import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
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

type LocalIdentityGatewayState =
  | { status: "probing" }
  | { status: "unauthenticated" }
  | { status: "ready"; profile: LocalIdentityAccountProfile }
  | { status: "failed"; message: string; code: string };

type LocalIdentityContextValue = {
  config: LocalIdentityConsumerConfig;
  profile: LocalIdentityAccountProfile;
  refresh: () => Promise<void>;
  linkOIDC: () => Promise<void>;
  revokeExternalIdentity: (bindingId: string, expectedRecordVersion: number) => Promise<void>;
};

const LocalIdentityContext = createContext<LocalIdentityContextValue | null>(null);

export function useLocalIdentity(): LocalIdentityContextValue | null {
  return useContext(LocalIdentityContext);
}

export function LocalIdentityGateway({ children }: { children: ReactNode }) {
  const config = useMemo(() => readLocalIdentityConsumerConfig(), []);
  const [state, setState] = useState<LocalIdentityGatewayState>({ status: "probing" });
  const [accountPanelOpen, setAccountPanelOpen] = useState(false);
  const [accountAction, setAccountAction] = useState<"" | "link" | "logout" | "revoke">("");
  const [accountActionError, setAccountActionError] = useState("");

  const refresh = useCallback(async () => {
    if (config.mode !== "local_identity_dev") return;
    try {
      const authentication = await probeLocalIdentitySession(config);
      if (!authentication) {
        setState({ status: "unauthenticated" });
        return;
      }
      const profile = await readLocalIdentityAccountProfile(config);
      setState({ status: "ready", profile });
    } catch (error) {
      const failure = identityFailure(error);
      if (failure.code === "LOCAL_IDENTITY_AUTHENTICATION_REQUIRED") {
        setState({ status: "unauthenticated" });
        return;
      }
      setState({ status: "failed", ...failure });
    }
  }, [config]);

  useEffect(() => {
    if (config.mode !== "local_identity_dev") return;
    const controller = new AbortController();
    void (async () => {
      try {
        const authentication = await probeLocalIdentitySession(config, controller.signal);
        if (!authentication) {
          setState({ status: "unauthenticated" });
          return;
        }
        const profile = await readLocalIdentityAccountProfile(config, controller.signal);
        setState({ status: "ready", profile });
      } catch (error) {
        if (error instanceof DOMException && error.name === "AbortError") return;
        const failure = identityFailure(error);
        setState({ status: "failed", ...failure });
      }
    })();
    return () => controller.abort();
  }, [config]);

  useEffect(() => {
    if (config.mode !== "local_identity_dev" || typeof BroadcastChannel === "undefined") return;
    const channel = new BroadcastChannel("radishmind-local-identity-v1");
    channel.onmessage = (event: MessageEvent<unknown>) => {
      if (isSessionChangedEvent(event.data)) void refresh();
    };
    return () => channel.close();
  }, [config.mode, refresh]);

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
    setAccountAction("logout");
    setAccountActionError("");
    try {
      await logoutLocalIdentity(config);
      broadcastSessionChanged();
      setAccountPanelOpen(false);
      setState({ status: "unauthenticated" });
    } catch (error) {
      setAccountActionError(identityFailure(error).message);
      setAccountAction("");
    }
  }

  const contextValue: LocalIdentityContextValue = {
    config,
    profile: state.profile,
    refresh,
    linkOIDC: handleLinkOIDC,
    revokeExternalIdentity: handleRevokeExternalIdentity,
  };

  return (
    <LocalIdentityContext.Provider value={contextValue}>
      {children}
      <aside className="local-identity-account-control" aria-label="Local identity session">
        <button
          type="button"
          className="local-identity-account-trigger"
          aria-expanded={accountPanelOpen}
          onClick={() => setAccountPanelOpen((open) => !open)}
        >
          <span aria-hidden="true">{state.profile.account.displayName.slice(0, 1).toUpperCase()}</span>
          <strong>{state.profile.account.displayName}</strong>
          <small>{state.profile.session.authenticationMethod === "oidc" ? "Radish OIDC" : "Local session"}</small>
        </button>
        {accountPanelOpen ? (
          <div className="local-identity-account-panel">
            <header>
              <p className="eyebrow">Account security · development/test</p>
              <strong>{state.profile.account.displayName}</strong>
              <small>{state.profile.account.userId}</small>
            </header>
            <dl>
              <div><dt>Session</dt><dd>{state.profile.session.authenticationMethod}</dd></div>
              <div><dt>External identities</dt><dd>{state.profile.externalIdentities.filter((item) => item.lifecycleState === "active").length}</dd></div>
              <div><dt>Local grants</dt><dd>{state.profile.roleAssignments.filter((item) => item.lifecycleState === "active").length}</dd></div>
            </dl>
            {accountActionError ? <p className="local-identity-inline-error" role="alert">{accountActionError}</p> : null}
            <div className="local-identity-account-actions">
              <a href="#admin-user-directory" onClick={() => setAccountPanelOpen(false)}>Open identity owner</a>
              <button
                type="button"
                onClick={() => void handleLinkOIDC()}
                disabled={!state.profile.capabilities.oidcEnabled || !state.profile.capabilities.recentAuthentication || accountAction !== ""}
              >
                {accountAction === "link" ? "Opening…" : "Link Radish identity"}
              </button>
              <button type="button" onClick={() => void handleLogout()} disabled={accountAction !== ""}>
                {accountAction === "logout" ? "Signing out…" : "Sign out"}
              </button>
            </div>
            <p className="local-identity-account-boundary">
              Session credentials stay in HttpOnly cookies. Upstream claims never become local grants.
            </p>
          </div>
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
      setError(identityFailure(submitError).message);
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
      setError(identityFailure(oidcError).message);
      setBusy("");
    }
  }

  return (
    <main className="local-identity-gateway" data-rd-profile="workbench">
      <header className="local-identity-gateway-header">
        <a href="/" className="local-identity-wordmark" aria-label="RadishMind home">
          <span aria-hidden="true">R</span><strong>RadishMind</strong>
        </a>
        <em>Development / Test authentication</em>
      </header>
      <div className="local-identity-gateway-layout">
        <aside className="local-identity-trust-rail" aria-label="Identity trust boundary">
          <p className="eyebrow">Identity trust boundary</p>
          <h1>One local account.<br />Explicit trust.</h1>
          <p>
            RadishMind owns its local account, Web session, membership and grant records. Radish OIDC is an optional
            login method bound by reviewed issuer and subject evidence.
          </p>
          <ol>
            <li><span>01</span><div><strong>Local session</strong><small>HttpOnly cookie; no browser token persistence.</small></div></li>
            <li><span>02</span><div><strong>Explicit binding</strong><small>Issuer + subject resolve one local user ID.</small></div></li>
            <li><span>03</span><div><strong>Local grants</strong><small>Never inferred from upstream claims or role names.</small></div></li>
          </ol>
          <div className="local-identity-stop-line">
            <strong>Current stop line</strong>
            <p>Development/test only. Production authentication, MFA, automated recovery and real Radish evidence remain closed.</p>
          </div>
        </aside>

        <section className="local-identity-form-zone" aria-labelledby="local-identity-form-title">
          <form className="local-identity-card" onSubmit={(event) => void handleSubmit(event)}>
            <header>
              <p className="eyebrow">Secure workspace access</p>
              <h2 id="local-identity-form-title">{intent === "login" ? "Welcome back" : "Create a local account"}</h2>
              <p>{intent === "login" ? "Use a reviewed local login method to continue." : "Register a development/test account with an explicit local credential."}</p>
            </header>
            <div className="local-identity-tabs" role="tablist" aria-label="Authentication mode">
              <button type="button" role="tab" aria-selected={intent === "login"} onClick={() => { setIntent("login"); setError(""); }}>Sign in</button>
              <button type="button" role="tab" aria-selected={intent === "register"} onClick={() => { setIntent("register"); setError(""); }}>Register</button>
            </div>
            {intent === "register" ? (
              <label>
                <span>Display name</span>
                <input autoComplete="name" value={displayName} onChange={(event) => setDisplayName(event.target.value)} required maxLength={120} />
              </label>
            ) : null}
            <label>
              <span>Local ID</span>
              <input autoComplete="username" value={loginIdentifier} onChange={(event) => setLoginIdentifier(event.target.value)} required maxLength={254} />
            </label>
            <label>
              <span>Password</span>
              <input
                type="password"
                autoComplete={intent === "login" ? "current-password" : "new-password"}
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                required
                minLength={12}
                maxLength={1024}
              />
              <small>Minimum 12 characters. The password never returns from the server.</small>
            </label>
            {error ? <p className="local-identity-form-error" role="alert">{error}</p> : null}
            <button type="submit" className="local-identity-primary-action" disabled={busy !== ""}>
              {busy === "password" ? "Verifying…" : intent === "login" ? "Sign in securely" : "Create local account"}
            </button>
            <div className="local-identity-divider"><span>or</span></div>
            <button type="button" className="local-identity-oidc-action" onClick={() => void handleOIDCLogin()} disabled={busy !== ""}>
              <span aria-hidden="true">R</span>{busy === "oidc" ? "Opening Radish…" : "Continue with Radish"}
            </button>
            <p className="local-identity-recovery-note">
              Account disabled or login method unavailable? Contact the local development administrator. No automated recovery path is enabled.
            </p>
          </form>
          <p className="local-identity-narrow-boundary">
            Session credentials stay in HttpOnly cookies; claims do not grant local permissions.
          </p>
        </section>
      </div>
    </main>
  );
}

function LocalIdentityLoading() {
  return (
    <main className="local-identity-status-surface" data-rd-profile="workbench">
      <div><span className="local-identity-status-mark" aria-hidden="true">R</span><p>Restoring the reviewed local session…</p></div>
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
  return (
    <main className="local-identity-status-surface" data-rd-profile="workbench">
      <div className="local-identity-failure-card">
        <p className="eyebrow">Local identity unavailable</p>
        <h1>Access remains closed</h1>
        <p>{failure.message}</p>
        <small>{failure.code}</small>
        <button type="button" onClick={() => void onRetry()}>Retry session check</button>
        <p className="local-identity-account-boundary">No fallback identity or development header will be substituted.</p>
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

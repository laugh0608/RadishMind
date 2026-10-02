import type { TFunction } from "i18next";

export function identityFailureMessage(t: TFunction<"identity">, code: string): string {
  switch (code) {
    case "LOCAL_IDENTITY_AUTHENTICATION_FAILED": return t($ => $.authenticationFailed);
    case "LOCAL_IDENTITY_ACCOUNT_CHANGE_REQUIRES_RECENT_AUTHENTICATION":
    case "LOCAL_IDENTITY_ACCOUNT_LINK_REQUIRES_RECENT_AUTHENTICATION": return t($ => $.recentAuthentication);
    case "LOCAL_IDENTITY_LAST_LOGIN_METHOD_REMOVAL_DENIED": return t($ => $.keepLoginMethod);
    case "LOCAL_IDENTITY_AUTHENTICATION_REQUIRED": return t($ => $.authenticationRequired);
    case "LOCAL_IDENTITY_CSRF_INVALID":
    case "LOCAL_IDENTITY_ORIGIN_FORBIDDEN": return t($ => $.scopeDenied);
    case "LOCAL_IDENTITY_PAYLOAD_INVALID": return t($ => $.invalidInput);
    case "LOCAL_IDENTITY_RETURN_TARGET_INVALID": return t($ => $.invalidReturnTarget);
    case "LOCAL_IDENTITY_ACCOUNT_CONFLICT": return t($ => $.accountConflict);
    case "LOCAL_IDENTITY_ALREADY_AUTHENTICATED": return t($ => $.alreadyAuthenticated);
    case "LOCAL_IDENTITY_OIDC_DISABLED": return t($ => $.oidcDisabled);
    case "LOCAL_IDENTITY_EXTERNAL_IDENTITY_UNBOUND": return t($ => $.identityUnbound);
    case "LOCAL_IDENTITY_EXTERNAL_IDENTITY_CONFLICT":
    case "LOCAL_IDENTITY_EXTERNAL_IDENTITY_VERSION_CONFLICT": return t($ => $.identityConflict);
    case "LOCAL_IDENTITY_EXTERNAL_IDENTITY_OWNERSHIP_DENIED":
    case "LOCAL_IDENTITY_OIDC_ADMISSION_DENIED": return t($ => $.identityDenied);
    case "LOCAL_IDENTITY_OIDC_STATE_INVALID":
    case "LOCAL_IDENTITY_OIDC_CALLBACK_MISMATCH":
    case "LOCAL_IDENTITY_OIDC_TOKEN_EXCHANGE_FAILED":
    case "LOCAL_IDENTITY_OIDC_IDENTITY_MISMATCH":
    case "LOCAL_IDENTITY_OIDC_PROVIDER_UNAVAILABLE": return t($ => $.oidcFailed);
    default: return t($ => $.serviceFailure);
  }
}

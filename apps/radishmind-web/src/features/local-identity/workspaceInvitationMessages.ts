import type { TFunction } from "i18next";
import type { UiLocale } from "../../i18n/localePreference.ts";
import { formatDisplayDate } from "../../i18n/formatters.ts";
import type { identityInvitations } from "../../i18n/locales/en-US/identityInvitations.ts";

const invitationFailureKeys = new Set<keyof typeof identityInvitations.failures>([
  "workspace_invitation_cursor_invalid", "workspace_invitation_role_ineligible",
  "workspace_invitation_role_catalog_mismatch", "workspace_invitation_version_conflict",
  "workspace_invitation_transition_invalid", "workspace_invitation_invalid",
  "workspace_invitation_not_claimable", "workspace_invitation_account_ineligible",
  "workspace_invitation_membership_conflict", "workspace_invitation_admin_unavailable",
  "workspace_invitation_store_unavailable", "local_identity_recent_authentication_required",
  "local_identity_admin_scope_mismatch", "local_identity_admin_unavailable",
  "workspace_membership_denied", "workspace_permission_denied",
  "LOCAL_IDENTITY_AUTHENTICATION_REQUIRED", "LOCAL_IDENTITY_CSRF_INVALID",
  "LOCAL_IDENTITY_HTTP_DISABLED", "LOCAL_IDENTITY_ORIGIN_FORBIDDEN",
  "LOCAL_IDENTITY_PAYLOAD_INVALID", "LOCAL_IDENTITY_SERVICE_UNAVAILABLE",
  "workspace_invitation_response_invalid", "workspace_invitation_input_invalid",
]);

export function invitationFailureMessage(t: TFunction<"identity">, code: string): string {
  const key = invitationFailureKeys.has(code as keyof typeof identityInvitations.failures)
    ? code as keyof typeof identityInvitations.failures : "unknown";
  return t($ => $.invitations.failures[key]);
}

export function invitationDate(t: TFunction<"identity">, value: string, locale: UiLocale): string {
  const formatted = formatDisplayDate(value, locale, { timeZone: "UTC" });
  return formatted === null ? t($ => $.invitations.unknown) : `${formatted} UTC`;
}

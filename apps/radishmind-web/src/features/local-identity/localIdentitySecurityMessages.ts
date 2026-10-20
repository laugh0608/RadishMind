import type { TFunction } from "i18next";
import type { UiLocale } from "../../i18n/localePreference.ts";
import { formatDisplayDate, formatDisplayNumber } from "../../i18n/formatters.ts";
import type { identitySecurity } from "../../i18n/locales/en-US/identitySecurity.ts";
import {
  LocalIdentitySelfServiceSecurityError,
  localIdentitySelfServiceSecurityFailureKind,
  type LocalIdentitySelfServiceSecurityFailureKind,
} from "./localIdentitySelfServiceSecurityConsumer.ts";

export type CredentialInputError = keyof typeof identitySecurity.inputErrors;
export type SecurityFailure = { kind: LocalIdentitySelfServiceSecurityFailureKind; code: string };
export type SecuritySuccess =
  | { kind: "currentRevoked" | "exactRevoked" }
  | { kind: "othersRevoked" | "credentialRevoked"; count: number }
  | { kind: "credentialClosed"; policyVersion: string };

export function securityFailure(error: unknown): SecurityFailure {
  return {
    kind: localIdentitySelfServiceSecurityFailureKind(error),
    code: error instanceof LocalIdentitySelfServiceSecurityError ? error.code : "local_identity_request_failed",
  };
}

export function securityFailureCopy(t: TFunction<"identity">, failure: SecurityFailure) {
  const key = failure.code === "local_identity_session_selection_stale" ? "selectionStale"
    : failure.code === "local_identity_credential_review_stale" ? "reviewStale" : failure.kind;
  return { title: t($ => $.security.failures[key].title), message: t($ => $.security.failures[key].message) };
}

export function securitySuccessMessage(t: TFunction<"identity">, feedback: SecuritySuccess, locale: UiLocale): string {
  switch (feedback.kind) {
    case "currentRevoked": return t($ => $.security.success.currentRevoked);
    case "exactRevoked": return t($ => $.security.success.exactRevoked);
    case "credentialClosed": return t($ => $.security.success.credentialClosed, { policyVersion: feedback.policyVersion });
    case "othersRevoked": return t($ => $.security.success.othersRevoked, { count: feedback.count, countText: formatDisplayNumber(feedback.count, locale) ?? t($ => $.security.unknown) });
    case "credentialRevoked": return t($ => $.security.success.credentialRevoked, { count: feedback.count, countText: formatDisplayNumber(feedback.count, locale) ?? t($ => $.security.unknown) });
  }
}

export function securityDate(t: TFunction<"identity">, value: string, locale: UiLocale): string {
  const formatted = formatDisplayDate(value, locale, { timeZone: "UTC" });
  return formatted === null ? t($ => $.security.unknown) : `${formatted} UTC`;
}

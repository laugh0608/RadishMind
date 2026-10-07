import type { TFunction } from "i18next";
import type { UiLocale } from "../../i18n/localePreference.ts";
import { formatDisplayDate, formatDisplayNumber } from "../../i18n/formatters.ts";
import {
  toLocalIdentityAdministrationError, localIdentityAdministrationFailureKind,
  type LocalIdentityAdministrationFailureKind,
} from "./localIdentityAdministrationConsumer.ts";

export type MemberFailure = { kind: LocalIdentityAdministrationFailureKind; code: string };
export type MemberSuccess =
  | { kind: "membershipCreated"; membershipId: string }
  | { kind: "membershipRevoked"; count: number }
  | { kind: "roleAssigned"; roleKey: string }
  | { kind: "roleRevoked"; roleKey: string; version: number };

export function memberFailure(error: unknown): MemberFailure {
  const failure = toLocalIdentityAdministrationError(error);
  return { kind: localIdentityAdministrationFailureKind(failure), code: failure.code };
}

export function memberFailureCopy(t: TFunction<"identity">, failure: MemberFailure) {
  return { title: t($ => $.members.failures[failure.kind].title), message: t($ => $.members.failures[failure.kind].message) };
}

export function memberSuccessMessage(t: TFunction<"identity">, success: MemberSuccess, locale: UiLocale): string {
  switch (success.kind) {
    case "membershipCreated": return t($ => $.members.success.membershipCreated, { membershipId: success.membershipId });
    case "membershipRevoked": return t($ => $.members.success.membershipRevoked, { countText: (formatDisplayNumber(success.count, locale) ?? t($ => $.members.unknown)) });
    case "roleAssigned": return t($ => $.members.success.roleAssigned, { roleKey: success.roleKey });
    case "roleRevoked": return t($ => $.members.success.roleRevoked, { roleKey: success.roleKey, version: success.version });
  }
}

export function memberDate(t: TFunction<"identity">, value: string, locale: UiLocale): string {
  const formatted = formatDisplayDate(value, locale, { timeZone: "UTC" });
  return formatted === null ? t($ => $.members.unknown) : `${formatted} UTC`;
}

import type { TFunction } from "i18next";
import type { adminQuota } from "../../i18n/locales/en-US/adminQuota.ts";
import type { adminPricing } from "../../i18n/locales/en-US/adminPricing.ts";
import type { AdminGatewayRequestQuotaFailureCode } from "./adminGatewayRequestQuotaConsumer.ts";

export type QuotaNotice = null
  | { key: Exclude<keyof typeof adminQuota.operations, "updated"> }
  | { key: "updated"; version: number | null }
  | { key: "failure"; code: AdminGatewayRequestQuotaFailureCode };

export function quotaFailureCopy(t: TFunction<"admin">, code: AdminGatewayRequestQuotaFailureCode) {
  return { code, shortLabel: t($ => $.quota.failures[code].shortLabel), title: t($ => $.quota.failures[code].title), summary: t($ => $.quota.failures[code].summary) };
}

export function quotaNoticeMessage(t: TFunction<"admin">, notice: QuotaNotice): string {
  if (!notice) return "";
  if (notice.key === "failure") return quotaFailureCopy(t, notice.code).summary;
  if (notice.key === "updated") return t($ => $.quota.operations.updated, { version: notice.version ?? t($ => $.quota.states.unknown) });
  const key = notice.key;
  return t($ => $.quota.operations[key]);
}

export type PricingNoticeMessage = null
  | { key: Exclude<keyof typeof adminPricing.operations, "conflict" | "updated"> }
  | { key: "conflict" | "updated"; version: number }
  | { key: "failure"; code: string };

export function pricingNoticeMessage(t: TFunction<"admin">, notice: PricingNoticeMessage): string {
  if (!notice) return "";
  if (notice.key === "failure") return pricingFailureMessage(t, notice.code);
  if (notice.key === "conflict" || notice.key === "updated") {
    const key = notice.key;
    return t($ => $.pricing.operations[key], { version: notice.version });
  }
  const key = notice.key;
  return t($ => $.pricing.operations[key]);
}

export function pricingFailureMessage(t: TFunction<"admin">, code: string): string {
  switch (code) {
    case "gateway_pricing_disabled":
    case "gateway_pricing_scope_denied":
    case "gateway_pricing_environment_forbidden":
    case "gateway_pricing_payload_invalid":
    case "gateway_pricing_policy_not_found":
    case "gateway_pricing_policy_version_conflict":
    case "gateway_pricing_policy_scope_conflict":
    case "gateway_pricing_store_unavailable":
      return t($ => $.pricing.failures[code]);
    default: return t($ => $.pricing.failures.unknown);
  }
}

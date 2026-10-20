import { uiI18n } from "../../i18n/instance.ts";
import { resolveUiLocale } from "../../i18n/localePreference.ts";
import { formatDisplayNumber } from "../../i18n/formatters.ts";
import type { TFunction } from "i18next";
import { auxiliaryEvidence } from "../../i18n/locales/en-US/auxiliaryEvidence.ts";
import type { gatewayOverviewEvidence } from "../../i18n/locales/en-US/gatewayOverviewEvidence.ts";
import type { gatewayRouteEvidence } from "../../i18n/locales/en-US/gatewayRouteEvidence.ts";
import type { gatewayUsageEvidence } from "../../i18n/locales/en-US/gatewayUsageEvidence.ts";
import type { gatewayReviewEvidence } from "../../i18n/locales/en-US/gatewayReviewEvidence.ts";
import type { adminOperationsEvidence } from "../../i18n/locales/en-US/adminOperationsEvidence.ts";
import type { adminDeploymentEvidence } from "../../i18n/locales/en-US/adminDeploymentEvidence.ts";

type EvidenceResources = typeof gatewayOverviewEvidence & typeof gatewayRouteEvidence & typeof gatewayUsageEvidence & typeof gatewayReviewEvidence & typeof adminOperationsEvidence & typeof adminDeploymentEvidence;
type Parameters<Text extends string> = Text extends `${string}{{${infer Name}}}${infer Rest}` ? Name | Parameters<Rest> : never;
export type EvidenceProjectionMessage = {
  [Key in keyof EvidenceResources]: { key: Key } & (Parameters<EvidenceResources[Key]> extends never
    ? { values?: never } : { values: Record<Parameters<EvidenceResources[Key]>, string | number | { status: string } | { number: number; style: "money" | "percent" }> });
}[keyof EvidenceResources];
export type EvidenceProjectionField = "label" | "summary" | "value" | "limit" | "used" | "usage" | "missingPrerequisite" | "reviewQuestion" | "operatorQuestion" | "enforcedBy" | "gatewayNarrative" | "routeNarrative" | "evidenceNarrative" | "reviewNarrative" | "resolvedModel" | "protocol";
// Only authored presentation metadata is translated; raw refs and record prose are preserved.
export type EvidenceProjectionCopy = { [Field in `${EvidenceProjectionField}Message`]?: EvidenceProjectionMessage };
export function evidenceProjectionStatus(t: TFunction<"evaluation">, status: string): string {
  return Object.hasOwn(auxiliaryEvidence.states, status)
    ? t($ => $.auxiliary.states[status as keyof typeof auxiliaryEvidence.states]) : status;
}
export function evidenceProjectionMessage(t: TFunction<"evaluation">, message: EvidenceProjectionMessage): string {
  const values = Object.fromEntries(Object.entries(message.values ?? {}).map(([key, value]) =>
    [key, typeof value === "number" ? formatDisplayNumber(value, resolveUiLocale(uiI18n.resolvedLanguage, []))
      : typeof value === "object" && "status" in value ? evidenceProjectionStatus(t, value.status)
        : typeof value === "object" ? formatDisplayNumber(value.number, resolveUiLocale(uiI18n.resolvedLanguage, []), value.style === "percent"
          ? { style: "percent", minimumFractionDigits: 1, maximumFractionDigits: 1 }
          : { minimumFractionDigits: 2, maximumFractionDigits: 2 }) : value]));
  return String(t($ => $.auxiliary[message.key] as string, values));
}
export function evidenceProjectionText<Field extends EvidenceProjectionField>(t: TFunction<"evaluation">, source: EvidenceProjectionCopy & Record<Field, string>, field: Field): string {
  const message = source[`${field}Message`];
  return message ? evidenceProjectionMessage(t, message) : source[field];
}

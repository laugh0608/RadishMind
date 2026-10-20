import type { TFunction } from "i18next";
import type { adminProvider } from "../../i18n/locales/en-US/adminProvider.ts";
import type { AdminProviderRouteValidationFinding, AdminProviderRouteDecision } from "./adminProviderRouteConsumer.ts";

type StaticProviderMessageKey = Exclude<keyof typeof adminProvider.operations, "loadingCandidate" | "candidateLoaded" | "recordingReview" | "activating" | "rollingBack">;
export type ProviderMessage =
  | { key: StaticProviderMessageKey | "failure" }
  | { key: "loadingCandidate" | "candidateLoaded"; candidateId: string }
  | { key: "recordingReview"; decision: AdminProviderRouteDecision }
  | { key: "activating" | "rollingBack"; generation: number };

export function providerOperationMessage(t: TFunction<"admin">, message: ProviderMessage, failureCode: string): string {
  if (message.key === "failure") return providerFailureMessage(t, failureCode);
  if (message.key === "loadingCandidate" || message.key === "candidateLoaded") {
    return t($ => $.provider.operations[message.key], { candidateId: message.candidateId });
  }
  if (message.key === "recordingReview") {
    return t($ => $.provider.operations.recordingReview, { decision: t($ => $.provider.states[message.decision]) });
  }
  if (message.key === "activating" || message.key === "rollingBack") {
    return t($ => $.provider.operations[message.key], { generation: message.generation });
  }
  const key = message.key;
  return t($ => $.provider.operations[key]);
}

export function providerFailureMessage(t: TFunction<"admin">, code: string): string {
  switch (code) {
    case "admin_provider_route_draft_revision_conflict":
    case "admin_provider_route_review_version_conflict":
    case "admin_provider_route_generation_conflict":
    case "admin_provider_route_inventory_not_found":
    case "admin_provider_route_inventory_mismatch":
    case "admin_provider_route_inventory_unavailable":
    case "admin_provider_route_candidate_not_approved":
    case "admin_provider_route_rollback_target_invalid":
      return t($ => $.provider.failures[code]);
    default: return t($ => $.provider.failures.unknown);
  }
}

export function providerFindingMessage(t: TFunction<"admin">, finding: AdminProviderRouteValidationFinding): string {
  switch (finding.kind) {
    case "duplicateProfile": return t($ => $.provider.findings.duplicateProfile, { profileId: finding.profileId });
    case "profileRef": return t($ => $.provider.findings.profileRef, { prefix: finding.prefix });
    case "duplicateRoute": return t($ => $.provider.findings.duplicateRoute, { routeId: finding.routeId });
    case "duplicateBinding": return t($ => $.provider.findings.duplicateBinding, { protocol: finding.protocol, modelId: finding.modelId });
    default: {
      const key = finding.kind;
      return t($ => $.provider.findings[key]);
    }
  }
}

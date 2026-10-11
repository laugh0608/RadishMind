import type { TFunction } from "i18next";
import { runEvidence } from "../../i18n/locales/en-US/runEvidence.ts";

export function runEvidenceState(t: TFunction<"evaluation">, value: string): string {
  if (!value) return t($ => $.evidence.absent);
  if (!Object.hasOwn(runEvidence.states, value)) return value;
  return t($ => $.evidence.states[value as keyof typeof runEvidence.states]);
}

export function runEvidenceBoundary(t: TFunction<"evaluation">, value: string): string {
  if (!value || value === "none") return t($ => $.evidence.none);
  if (!Object.hasOwn(runEvidence.boundaries, value)) return value;
  return t($ => $.evidence.boundaries[value as keyof typeof runEvidence.boundaries]);
}

export function runEvidenceAction(t: TFunction<"evaluation">, value: string): string {
  if (!value) return t($ => $.evidence.none);
  if (!Object.hasOwn(runEvidence.actions, value)) return t($ => $.evidence.unknownAction);
  return t($ => $.evidence.actions[value as keyof typeof runEvidence.actions]);
}

export function runEvidenceFinding(t: TFunction<"evaluation">, value: string): string {
  if (!Object.hasOwn(runEvidence.findings, value)) return t($ => $.evidence.unknownFinding);
  return t($ => $.evidence.findings[value as keyof typeof runEvidence.findings]);
}

export function runEvidenceNodeType(t: TFunction<"evaluation">, value: string): string {
  if (!Object.hasOwn(runEvidence.nodeTypes, value)) return value;
  return t($ => $.evidence.nodeTypes[value as keyof typeof runEvidence.nodeTypes]);
}

export function runEvidenceSourceType(t: TFunction<"evaluation">, value: string): string {
  if (!Object.hasOwn(runEvidence.sourceTypes, value)) return value;
  return t($ => $.evidence.sourceTypes[value as keyof typeof runEvidence.sourceTypes]);
}

export function runComparisonFailure(t: TFunction<"evaluation">, code: string): string {
  const key = Object.hasOwn(runEvidence.failures, code) ? code as keyof typeof runEvidence.failures : "unknown";
  return t($ => $.evidence.failures[key]);
}

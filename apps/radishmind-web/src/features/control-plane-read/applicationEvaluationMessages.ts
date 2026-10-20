import type { TFunction } from "i18next";
import { evaluationCampaign } from "../../i18n/locales/en-US/evaluationCampaign.ts";
import type { ApplicationEvaluationFailureCode } from "./applicationEvaluationCampaignConsumer.ts";
import type { ApplicationEvaluationScheduleFailureCode } from "./applicationEvaluationScheduleConsumer.ts";

type Operation = keyof typeof evaluationCampaign.operations;
export type CampaignMessage =
  | { kind: "strictFailure" | "invalidPlan" | "selectPlan" | "newScheduleMessage" | "chooseScheduleInputs" | "selectSchedule" | "invalidCadence" }
  | { kind: "planLoadedMessage"; id: string; version: number }
  | { kind: "planArchivedMessage"; id: string }
  | { kind: "planSavedMessage"; operation: Operation; id: string; version: number }
  | { kind: "campaignFinishedMessage" | "campaignReconciledMessage"; id: string; state: string }
  | { kind: "handoffMessage"; state: string; count: number; suite: string }
  | { kind: "scheduleSavedMessage"; operation: Operation; version: number }
  | { kind: "occurrenceLoadedMessage"; time: string };

export function campaignState(t: TFunction<"evaluation">, value: string): string {
  return Object.hasOwn(evaluationCampaign.states, value)
    ? t($ => $.campaign.states[value as keyof typeof evaluationCampaign.states]) : value;
}

export function campaignOperation(t: TFunction<"evaluation">, value: Operation): string {
  return t($ => $.campaign.operations[value]);
}

export function campaignFailure(t: TFunction<"evaluation">, code: string | null): string {
  return code && Object.hasOwn(failureKeys, code)
    ? t($ => $.campaign.failures[code as ApplicationEvaluationScheduleFailureCode])
    : t($ => $.campaign.strictFailure);
}

// Exhaustiveness is checked against both bounded protocol failure sets.
const failureKeys: Record<ApplicationEvaluationFailureCode | ApplicationEvaluationScheduleFailureCode, string> = evaluationCampaign.failures;

export function campaignMessage(t: TFunction<"evaluation">, message: CampaignMessage | null): string {
  if (!message) return "";
  switch (message.kind) {
    case "planLoadedMessage": return t($ => $.campaign.planLoadedMessage, message);
    case "planArchivedMessage": return t($ => $.campaign.planArchivedMessage, message);
    case "planSavedMessage": return t($ => $.campaign.planSavedMessage, { ...message, operation: campaignOperation(t, message.operation) });
    case "campaignFinishedMessage": return t($ => $.campaign.campaignFinishedMessage, { ...message, state: campaignState(t, message.state) });
    case "campaignReconciledMessage": return t($ => $.campaign.campaignReconciledMessage, { ...message, state: campaignState(t, message.state) });
    case "handoffMessage": return t($ => $.campaign.handoffMessage, { ...message, state: campaignState(t, message.state), suite: message.suite || t($ => $.campaign.pendingSuiteValue) });
    case "scheduleSavedMessage": return t($ => $.campaign.scheduleSavedMessage, { ...message, operation: campaignOperation(t, message.operation) });
    case "occurrenceLoadedMessage": return t($ => $.campaign.occurrenceLoadedMessage, message);
    default: return t($ => $.campaign[message.kind]);
  }
}

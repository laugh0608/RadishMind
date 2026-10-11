import type { TFunction } from "i18next";
import { evaluationCases } from "../../i18n/locales/en-US/evaluationCases.ts";
import { EvaluationRequestFailure } from "./evaluationFailure.ts";

export function evaluationState(t: TFunction<"evaluation">, value: string): string {
  if (!Object.hasOwn(evaluationCases.states, value)) return value;
  return t($ => $.cases.states[value as keyof typeof evaluationCases.states]);
}

export function evaluationFailureCode(error: unknown): string {
  return error instanceof EvaluationRequestFailure ? error.code : "unknownFailure";
}

const failureKeys = {
  workflow_evaluation_invalid: "invalid",
  workflow_evaluation_suite_invalid: "invalid",
  workflow_evaluation_not_found: "notFound",
  workflow_evaluation_suite_not_found: "notFound",
  workflow_evaluation_run_not_eligible: "notEligible",
  workflow_evaluation_suite_case_not_eligible: "notEligible",
  workflow_evaluation_store_unavailable: "storeUnavailable",
  workflow_evaluation_suite_store_unavailable: "storeUnavailable",
  workflow_evaluation_store_contract_mismatch: "contract",
  workflow_evaluation_suite_store_contract_mismatch: "contract",
  workflow_evaluation_suite_approval_blocked: "approvalBlocked",
  workflow_evaluation_suite_review_changed: "reviewChanged",
  offline: "offline", contract: "contract",
} as const;

export function evaluationFailure(t: TFunction<"evaluation">, code: string): string {
  const key = Object.hasOwn(failureKeys, code) ? failureKeys[code as keyof typeof failureKeys] : "unknownFailure";
  return t($ => $.cases[key]);
}

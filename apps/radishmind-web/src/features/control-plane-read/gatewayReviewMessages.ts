import type { TFunction } from "i18next";
import { gatewayReview } from "../../i18n/locales/en-US/gatewayReview.ts";

export function gatewayReviewState(t: TFunction<"gateway">, value: string): string {
  if (!Object.hasOwn(gatewayReview.states, value)) return value;
  const key = value as keyof typeof gatewayReview.states;
  return t($ => $.review.states[key]);
}

export function gatewayHistoryFailure(t: TFunction<"gateway">, code: string): string {
  const key = Object.hasOwn(gatewayReview.failures, code) ? code as keyof typeof gatewayReview.failures : "unknown";
  return t($ => $.review.failures[key]);
}

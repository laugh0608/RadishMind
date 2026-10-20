import type { ApplicationEvaluationPlan, ApplicationEvaluationPlanDraft, ApplicationEvaluationPlanVersion } from "./applicationEvaluationCampaignConsumer.ts";
import { loadPromptPlanEditor, promptPlanFingerprint, validatePromptPlanEditor, type PromptPlanBase, type PromptPlanEditor } from "./promptEvaluationPlanDraft.ts";
import type { PromptPlanReference } from "./promptEvaluationPlanReference.ts";

export type PromptPlanReview = {
  draft: ApplicationEvaluationPlanDraft;
  base: PromptPlanBase | null;
  fingerprint: string;
  referenceKey: string | null;
  matched: boolean;
};

export function reviewPromptPlan(state: PromptPlanEditor, reference: PromptPlanReference | null): PromptPlanReview | null {
  const validation = validatePromptPlanEditor(state, reference?.source ?? null);
  if (!validation.draft || state.content.mode === "form" && !validation.matched) return null;
  return { draft: validation.draft, base: state.base, fingerprint: promptPlanFingerprint(state), referenceKey: reference?.key ?? null, matched: validation.matched };
}

export function promptPlanReviewIsCurrent(state: PromptPlanEditor, review: PromptPlanReview, reference: PromptPlanReference | null): boolean {
  return review.fingerprint === promptPlanFingerprint(state) && review.referenceKey === (reference?.key ?? null) &&
    JSON.stringify(review.base) === JSON.stringify(state.base);
}

export function acceptPromptPlanSave(state: PromptPlanEditor, review: PromptPlanReview, plan: ApplicationEvaluationPlan, version: ApplicationEvaluationPlanVersion): PromptPlanEditor {
  if (review.fingerprint !== promptPlanFingerprint(state) || JSON.stringify(review.base) !== JSON.stringify(state.base)) {
    throw new Error("Saved plan no longer belongs to this edit.");
  }
  if (review.base && (plan.planId !== review.base.planId || plan.recordVersion !== review.base.recordVersion + 1 ||
    version.planVersion !== review.base.planVersion + 1 || version.previousPlanVersion !== review.base.planVersion)) {
    throw new Error("Saved plan does not continue the reviewed CAS lineage.");
  }
  if (!review.base && (plan.recordVersion !== 1 || version.planVersion !== 1 || version.previousPlanVersion !== 0)) {
    throw new Error("Created plan has an unexpected initial version.");
  }
  return loadPromptPlanEditor(plan, version, state);
}

import assert from "node:assert/strict";
import test from "node:test";
import { acceptPromptPlanSave, promptPlanReviewIsCurrent, reviewPromptPlan } from "../src/features/control-plane-read/promptEvaluationPlanReview.ts";
import { addPromptPlanSample, newPromptPlanEditor, promptPlanIsDirty, switchPromptPlanMode, updatePromptPlanSample } from "../src/features/control-plane-read/promptEvaluationPlanDraft.ts";
import type { PromptPlanReference } from "../src/features/control-plane-read/promptEvaluationPlanReference.ts";
import type { ApplicationEvaluationPlan, ApplicationEvaluationPlanVersion } from "../src/features/control-plane-read/applicationEvaluationCampaignConsumer.ts";

const reference: PromptPlanReference = {
  key: "assignment:1", assignmentVersion: 1,
  source: {
    templateId: "ptpl_aaaaaaaaaaaaaaaa", templateVersion: 1, sourceDraftVersion: 1, templateDigest: `sha256:${"a".repeat(64)}`,
    tenantRef: "tenant_demo", ownerSubjectRef: "subject_demo_user", workspaceId: "workspace_demo", applicationId: "app_aaaaaaaaaaaaaaaa",
    templateName: "Reference", description: "", messages: [{ role: "user", content: "{{question}}" }],
    variables: [{ name: "question", type: "string", required: true, description: "" }], outputContract: { kind: "text", allowEmpty: false, maxBytes: 1024 },
    createdAt: "2026-10-15T10:00:00Z", createdByActorRef: "subject_demo_user", requestId: "read-source", auditRef: "read-source",
  },
};

function draft() {
  const empty = addPromptPlanSample(newPromptPlanEditor());
  return updatePromptPlanSample({ ...empty, name: "New plan" }, empty.selectedKey, { name: "One sample", variables: { question: { kind: "value", value: "Check exact version" } } });
}

test("only a matching form or explicitly reviewed strict JSON can produce a save snapshot", () => {
  assert.equal(reviewPromptPlan(newPromptPlanEditor(), reference), null);
  const state = draft();
  assert.equal(reviewPromptPlan(state, null), null);
  const json = switchPromptPlanMode(state, "json").state;
  const jsonReview = reviewPromptPlan(json, null)!;
  assert.equal(jsonReview.matched, false);
  assert.equal(jsonReview.referenceKey, null);
  assert.equal(jsonReview.base, null);
  assert.equal(promptPlanReviewIsCurrent(json, jsonReview, null), true);
  const formReview = reviewPromptPlan(state, reference)!;
  assert.equal(formReview.matched, true);
  assert.equal(formReview.referenceKey, reference.key);
  assert.equal(promptPlanReviewIsCurrent({ ...state, name: "Changed" }, formReview, reference), false);
  assert.equal(promptPlanReviewIsCurrent(state, formReview, { ...reference, key: "assignment:2" }), false);
  assert.equal(promptPlanReviewIsCurrent(state, formReview, null), false);
  assert.equal(promptPlanReviewIsCurrent({ ...state, base: { planId: "another", planVersion: 1, recordVersion: 1, planDigest: "different" } }, formReview, reference), false);
  assert.equal(JSON.stringify(formReview.draft).includes("assignment"), false);
  assert.equal(JSON.stringify(formReview.draft).includes("templateId"), false);
});

test("save acceptance retains the selected sample and mode, binds returned CAS and rejects stale or unrelated responses", () => {
  const state = switchPromptPlanMode(draft(), "json").state;
  const review = reviewPromptPlan(state, null)!;
  const plan = { planId: "plan_first", recordVersion: 1, latestPlanVersion: 1, latestPlanDigest: reference.source.templateDigest } as ApplicationEvaluationPlan;
  const version = { ...review.draft, planId: plan.planId, planVersion: 1, previousPlanVersion: 0, planDigest: plan.latestPlanDigest } as ApplicationEvaluationPlanVersion;
  const saved = acceptPromptPlanSave(state, review, plan, version);
  assert.equal(saved.content.mode, "json");
  assert.equal(saved.selectedKey, state.selectedKey);
  assert.equal(saved.base?.planId, plan.planId);
  assert.equal(promptPlanIsDirty(saved), false);
  assert.throws(() => acceptPromptPlanSave({ ...state, name: "Typed after request" }, review, plan, version));
  assert.throws(() => acceptPromptPlanSave(state, review, { ...plan, recordVersion: 2 }, version));
  const revision = reviewPromptPlan({ ...saved, name: "Revised" }, null)!;
  const revisedPlan = { ...plan, recordVersion: 2, latestPlanVersion: 2 };
  const revisedVersion = { ...version, name: "Revised", planVersion: 2, previousPlanVersion: 1 };
  const next = acceptPromptPlanSave({ ...saved, name: "Revised" }, revision, revisedPlan, revisedVersion);
  assert.equal(next.base?.recordVersion, 2);
  assert.equal(next.selectedKey, state.selectedKey);
  for (const invalid of [{ planId: "another" }, { recordVersion: 3 }, { latestPlanVersion: 3 }]) {
    assert.throws(() => acceptPromptPlanSave({ ...saved, name: "Revised" }, revision, { ...revisedPlan, ...invalid }, revisedVersion));
  }
  assert.throws(() => acceptPromptPlanSave({ ...saved, name: "Revised" }, revision, revisedPlan, { ...revisedVersion, previousPlanVersion: 0 }));
});

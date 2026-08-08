import assert from "node:assert/strict";
import test from "node:test";

import {
  applicationApiAccessPrimaryHref,
  workflowDesignerPrimaryHref,
  workflowReviewPrimaryHref,
} from "../src/app/productNavigationRoute.ts";

test("maps only the explicit S4 application API access anchors to API & keys", () => {
  for (const anchor of [
    "#application-api-integration",
    "#workspace-api-keys",
  ]) {
    assert.equal(applicationApiAccessPrimaryHref(anchor), "#workspace-api-keys");
  }
});

test("does not claim adjacent application, API key, or prefix-like anchors", () => {
  for (const anchor of [
    "#application-api-integration-preview",
    "#workspace-api-keys-archive",
    "#application-interaction-session",
    "#model-gateway-playground",
    "application-api-integration",
    "#workspace-applications",
  ]) {
    assert.equal(applicationApiAccessPrimaryHref(anchor), null);
  }
});

test("maps only the explicit S3 workflow designer anchors to Workflows", () => {
  for (const anchor of [
    "#workflow-user-workspace-home",
    "#workflow-draft-designer",
    "#workflow-draft-validation-inspector",
    "#workflow-execution-plan-preview",
    "#workflow-runtime-readiness-inspector",
    "#workflow-review-handoff",
  ]) {
    assert.equal(workflowDesignerPrimaryHref(anchor), "#workspace-workflow-definitions");
  }
});

test("does not claim adjacent workflow or prefix-like anchors", () => {
  for (const anchor of [
    "#workflow-draft-designer-preview",
    "#workflow-review-handoff-export",
    "#workflow-http-tool-action-review",
    "#workflow-executor-v0",
    "#workflow-scenario-inspector",
    "#workflow-workspace-review",
    "workflow-draft-designer",
    "#workspace-applications",
  ]) {
    assert.equal(workflowDesignerPrimaryHref(anchor), null);
  }
});

test("maps only the explicit S6 workflow review anchors to Run history", () => {
  for (const anchor of [
    "#workspace-run-history",
    "#workflow-run-comparison",
    "#workflow-evaluation-cases",
    "#workflow-evaluation-release-review",
  ]) {
    assert.equal(workflowReviewPrimaryHref(anchor), "#workspace-run-history");
  }
});

test("does not claim adjacent workflow review anchors", () => {
  for (const anchor of [
    "#workflow-run-comparison-preview",
    "#workflow-evaluation-cases-archive",
    "#workflow-evaluation-release-review-export",
    "workflow-run-comparison",
    "#application-operations",
  ]) {
    assert.equal(workflowReviewPrimaryHref(anchor), null);
  }
});

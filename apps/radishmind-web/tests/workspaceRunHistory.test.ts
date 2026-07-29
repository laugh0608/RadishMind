import assert from "node:assert/strict";
import test from "node:test";

import { toControlPlaneReadCollectionViewModel } from "../../../contracts/typescript/control-plane-read-api.ts";
import { buildWorkspaceRunHistoryViewModel } from "../src/features/control-plane-read/workspaceRunHistory.ts";

test("durable run history renders an explicitly unavailable cost without inventing zero", () => {
  const collection = toControlPlaneReadCollectionViewModel("run-record-summary-list-route", {
    request_id: "request_run_cost_unavailable",
    tenant_ref: "tenant_demo",
    items: [{
      run_id: "run_cost_unavailable",
      tenant_ref: "tenant_demo",
      workflow_definition_id: "workflow_definition_demo",
      application_ref: "app_flow_copilot",
      status: "succeeded",
      failure_code: null,
      cost_summary: {},
      trace_id: "trace_run_cost_unavailable",
      started_at: "2026-07-29T12:00:00Z",
      completed_at: "2026-07-29T12:00:01Z",
    }],
    next_cursor: null,
    failure_code: null,
    audit_ref: "audit_run_cost_unavailable",
  }, { source: "dev_live_http" });

  const view = buildWorkspaceRunHistoryViewModel(collection);

  assert.equal(view.runs[0]?.estimatedCost, "unavailable");
  assert.equal(view.canRenderRuns, true);
});

test("partial or malformed run cost metadata remains visibly invalid", () => {
  const collection = toControlPlaneReadCollectionViewModel("run-record-summary-list-route", {
    request_id: "request_run_cost_invalid",
    tenant_ref: "tenant_demo",
    items: [{
      run_id: "run_cost_invalid",
      tenant_ref: "tenant_demo",
      workflow_definition_id: "workflow_definition_demo",
      application_ref: "app_flow_copilot",
      status: "failed",
      failure_code: "provider_request_failed",
      cost_summary: { estimated_cost: 0.2 },
      trace_id: "trace_run_cost_invalid",
      started_at: "2026-07-29T12:00:00Z",
      completed_at: "2026-07-29T12:00:01Z",
    }],
    next_cursor: null,
    failure_code: null,
    audit_ref: "audit_run_cost_invalid",
  }, { source: "dev_live_http" });

  const view = buildWorkspaceRunHistoryViewModel(collection);

  assert.equal(view.runs[0]?.estimatedCost, "invalid");
});

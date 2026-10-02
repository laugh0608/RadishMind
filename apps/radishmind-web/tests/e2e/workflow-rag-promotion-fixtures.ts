import { randomUUID } from "node:crypto";
import type { Page } from "@playwright/test";
import { expect, type Application } from "./workflow-fixtures";

export const promotionPath = "/v1/user-workspace/workflow-rag-knowledge-promotion-candidates";
const base = "http://127.0.0.1:17000";

// Synthetic authority records use public APIs; every run owns a temporary SQLite database.
export async function promotionRequest(page: Page, applicationId: string, path: string, scopes: string[], data?: object) {
  const requestId = `rag-promotion-e2e-${randomUUID()}`;
  const headers = {
    "X-Request-Id": requestId, "X-RadishMind-Dev-Read-Audit": `audit-${requestId}`,
    "X-RadishMind-Dev-Read-Identity": "radishmind-web-workflow-rag-dev",
    "X-RadishMind-Dev-Read-Tenant": "tenant_demo", "X-RadishMind-Dev-Read-Subject": "subject_demo_user",
    "X-RadishMind-Dev-Read-Scopes": scopes.join(","), "X-RadishMind-Dev-Workflow-Workspace": "workspace_demo",
    "X-RadishMind-Dev-Workflow-Application": applicationId, "X-RadishMind-Active-Workspace": "workspace_demo",
    ...(path.startsWith("/v1/user-workspace/application-drafts") ? {
      "X-RadishMind-Dev-Application-Draft-Workspace": "workspace_demo", "X-RadishMind-Dev-Application-Draft-Application": applicationId,
    } : {}),
    "X-RadishMind-Dev-Read-Membership-Workspace": "workspace_demo", "X-RadishMind-Dev-Read-Membership-Permissions": scopes.join(","),
  };
  const response = data ? await page.request.post(`${base}${path}`, { headers, data }) : await page.request.get(`${base}${path}`, { headers });
  expect(response.ok(), await response.text()).toBeTruthy();
  const body = await response.json();
  expect(body.failure_code, JSON.stringify(body)).toBeNull();
  return body;
}

export async function preparePromotionEvidence(page: Page, application: Application) {
  const scope = { workspace_id: "workspace_demo", application_id: application.id };
  const query = new URLSearchParams({ workspace_id: scope.workspace_id });
  const app = await promotionRequest(page, application.id, `/v1/user-workspace/applications/${application.id}?${query}`, ["applications:read"]);
  const draftId = `promotion-${randomUUID()}`;
  const draft = await promotionRequest(page, application.id, "/v1/user-workspace/application-drafts", ["application_drafts:write"], {
    expected_draft_version: 0, draft: {
      ...scope, draft_id: draftId, base_application_updated_at: app.record.updated_at,
      schema_version: "application_configuration_draft.v1", display_name: application.name,
      description: "Synthetic promotion evidence", application_kind: application.kind,
      default_protocol: "chat_completions", default_model: "profile:prompt-e2e", allowed_protocols: ["chat_completions"],
    },
  });
  expect(draft.draft.validation_summary.state).toBe("valid");
  const key = randomUUID().replaceAll("-", "").slice(0, 12);
  async function snapshot(suffix: string) {
    const result = await promotionRequest(page, application.id, "/v1/user-workspace/workflow-retrieval-snapshots", ["workflow_rag_snapshots:write"], {
      ...scope, snapshot_key: `promotion_${key}_${suffix}`, display_name: `Promotion ${suffix}`, content_classification: "public",
      fragments: [{ fragment_ref: "official_promotion", source_type: "manual", source_ref: "manual.promotion", page_slug: "promotion/evidence", title: "Promotion evidence", is_official: true, content: "Promotion authority evidence for reviewed immutable knowledge." }],
    });
    const record = result.record;
    return { tenant_ref: record.tenant_ref, ...scope, snapshot_id: record.snapshot_id, snapshot_version: record.snapshot_version, snapshot_digest: record.snapshot_digest, rag_ref: record.rag_ref };
  }
  const baseline = await snapshot("baseline"), candidate = await snapshot("candidate");
  const dataset = await promotionRequest(page, application.id, "/v1/user-workspace/workflow-rag-evaluation-datasets", ["workflow_rag_evaluation_datasets:write", "workflow_rag_snapshots:read"], {
    ...scope, dataset_key: `promotion_${key}`, display_name: "Promotion fixture", content_classification: "synthetic_public", baseline_snapshot: baseline,
    thresholds: { hit_at_k: 1, expected_recall_at_k: 1, required_official_recall_at_k: 1, mean_reciprocal_rank: 1, no_evidence_accuracy: 1, sample_pass_rate: 1 },
    review_summary: "Synthetic public review", samples: [{ sample_id: "promotion_evidence", query_text: "Promotion authority evidence", expectation: "evidence_required", expected_citation_refs: ["official_promotion"], required_official_refs: ["official_promotion"], top_k: 1, review_note: "Require official evidence" }],
  });
  const datasetId = dataset.resource.dataset_id;
  const digest = dataset.version.dataset.dataset_digest;
  const review = await promotionRequest(page, application.id, `/v1/user-workspace/workflow-rag-evaluation-datasets/${datasetId}/candidate-reviews`, ["workflow_rag_evaluation_datasets:review", "workflow_rag_evaluation_datasets:read", "workflow_rag_snapshots:read"], {
    ...scope, dataset_version: 1, dataset_digest: digest, candidate_snapshot: candidate,
  });
  expect(review.review.candidate.status).toBe("passed");
  expect(review.review.conclusion).toBe("unchanged");
  return { scope, draftId, datasetId, digest, reviewId: review.review.review_id as string };
}

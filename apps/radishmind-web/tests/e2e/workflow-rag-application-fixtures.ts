import { randomUUID } from "node:crypto";
import type { Page } from "@playwright/test";
import { expect, type Application } from "./workflow-fixtures";
import { preparePromotionEvidence, promotionPath, promotionRequest } from "./workflow-rag-promotion-fixtures";

const publishPath = "/v1/user-workspace/application-publish-candidates";

export async function prepareRAGApplication(page: Page, application: Application) {
  const evidence = await preparePromotionEvidence(page, application, "mock");
  const created = await promotionRequest(page, application.id, promotionPath, ["workflow_rag_promotions:write", "workflow_rag_evaluation_datasets:read", "workflow_rag_snapshots:read", "application_drafts:read"], {
    ...evidence.scope, dataset_id: evidence.datasetId, dataset_version: 1, dataset_digest: evidence.digest,
    candidate_review_id: evidence.reviewId, draft_id: evidence.draftId, expected_draft_version: 1,
  });
  const approved = await promotionRequest(page, application.id, `${promotionPath}/${created.candidate.candidate_id}/decisions`, ["workflow_rag_promotions:review"], {
    ...evidence.scope, expected_record_version: 1, decision: "approve", reason: "Reviewed synthetic public knowledge for controlled invocation.",
  });
  expect(approved.eligibility.eligible).toBe(true);
  const query = new URLSearchParams(evidence.scope);
  const original = await promotionRequest(page, application.id, `/v1/user-workspace/application-drafts/${evidence.draftId}?${query}`, ["application_drafts:read"]);
  const draft = original.draft;
  const binding = { binding_id: approved.binding.binding_id, binding_version: 1, binding_digest: approved.binding.binding_digest };
  const saved = await promotionRequest(page, application.id, "/v1/user-workspace/application-drafts", ["application_drafts:read", "application_drafts:write", "workflow_rag_promotions:bind"], {
    expected_draft_version: 1, draft: {
      ...evidence.scope, draft_id: draft.draft_id, schema_version: "application_configuration_draft.v2",
      base_application_updated_at: draft.base_application_updated_at, display_name: draft.display_name,
      description: draft.description, application_kind: draft.application_kind, default_protocol: draft.default_protocol,
      default_model: draft.default_model, allowed_protocols: draft.allowed_protocols, workflow_rag_binding_ref: binding,
    },
  });
  expect(saved.draft.draft_version).toBe(2); expect(saved.draft.validation_summary.state).toBe("valid");
  async function publish() {
    const candidateId = `rag-application-${randomUUID()}`;
    const candidate = await promotionRequest(page, application.id, publishPath, ["application_publish_candidates:write", "workflow_rag_promotions:read"], {
      candidate_id: candidateId, draft_id: evidence.draftId, expected_draft_version: 2, evidence_request_ids: [],
    });
    const reviewed = await promotionRequest(page, application.id, `${publishPath}/${candidateId}/reviews`, ["application_publish_candidates:review", "workflow_rag_promotions:read"], {
      expected_review_version: candidate.candidate.review_version, decision: "approve", reason: "Reviewed exact immutable RAG binding and configuration.",
    });
    expect(reviewed.candidate.candidate_state).toBe("approved");
    expect(reviewed.candidate.promotion_eligibility.eligible).toBe(false);
    return candidateId;
  }
  return { ...evidence, binding, candidateId: await publish(), publish };
}

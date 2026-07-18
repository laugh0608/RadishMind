package httpapi

import (
	"encoding/json"
	"time"
)

func encodeWorkflowRAGEvaluationResource(resource WorkflowRAGEvaluationDatasetResource, ctx WorkflowRAGSnapshotContext) ([]byte, error) {
	if validateStoredWorkflowRAGEvaluationResource(resource, ctx) != nil {
		return nil, errWorkflowRAGEvaluationContract
	}
	payload, err := json.Marshal(resource)
	if err != nil {
		return nil, errWorkflowRAGEvaluationContract
	}
	return payload, nil
}

func decodeWorkflowRAGEvaluationResource(payload []byte, ctx WorkflowRAGSnapshotContext) (WorkflowRAGEvaluationDatasetResource, error) {
	var resource WorkflowRAGEvaluationDatasetResource
	if decodeWorkflowRAGStrictJSON(payload, &resource) != nil || validateStoredWorkflowRAGEvaluationResource(resource, ctx) != nil {
		return WorkflowRAGEvaluationDatasetResource{}, errWorkflowRAGEvaluationContract
	}
	return resource, nil
}

func encodeWorkflowRAGEvaluationVersion(version WorkflowRAGEvaluationDatasetVersion, ctx WorkflowRAGSnapshotContext) ([]byte, error) {
	if validateStoredWorkflowRAGEvaluationVersion(version, ctx) != nil {
		return nil, errWorkflowRAGEvaluationContract
	}
	payload, err := json.Marshal(version)
	if err != nil {
		return nil, errWorkflowRAGEvaluationContract
	}
	return payload, nil
}

func decodeWorkflowRAGEvaluationVersion(payload []byte, ctx WorkflowRAGSnapshotContext) (WorkflowRAGEvaluationDatasetVersion, error) {
	var version WorkflowRAGEvaluationDatasetVersion
	if decodeWorkflowRAGStrictJSON(payload, &version) != nil || validateStoredWorkflowRAGEvaluationVersion(version, ctx) != nil {
		return WorkflowRAGEvaluationDatasetVersion{}, errWorkflowRAGEvaluationContract
	}
	return version, nil
}

func encodeWorkflowRAGCandidateReview(review WorkflowRAGCandidateSnapshotReview, ctx WorkflowRAGSnapshotContext) ([]byte, error) {
	if validateStoredWorkflowRAGCandidateReview(review, ctx) != nil {
		return nil, errWorkflowRAGEvaluationContract
	}
	payload, err := json.Marshal(review)
	if err != nil {
		return nil, errWorkflowRAGEvaluationContract
	}
	return payload, nil
}

func decodeWorkflowRAGCandidateReview(payload []byte, ctx WorkflowRAGSnapshotContext) (WorkflowRAGCandidateSnapshotReview, error) {
	var review WorkflowRAGCandidateSnapshotReview
	if decodeWorkflowRAGStrictJSON(payload, &review) != nil || review.TenantRef != ctx.TenantRef || review.WorkspaceID != ctx.WorkspaceID || review.ApplicationID != ctx.ApplicationID {
		return WorkflowRAGCandidateSnapshotReview{}, errWorkflowRAGEvaluationContract
	}
	storedContext := ctx
	storedContext.ActorRef = review.CreatedByActorRef
	storedContext.RequestID = review.RequestID
	storedContext.AuditRef = review.AuditRef
	if validateStoredWorkflowRAGCandidateReview(review, storedContext) != nil {
		return WorkflowRAGCandidateSnapshotReview{}, errWorkflowRAGEvaluationContract
	}
	return review, nil
}

func encodeWorkflowRAGEvaluationAudit(audit WorkflowRAGEvaluationAudit, ctx WorkflowRAGSnapshotContext) ([]byte, error) {
	if validateStoredWorkflowRAGEvaluationAudit(audit, ctx) != nil {
		return nil, errWorkflowRAGEvaluationContract
	}
	payload, err := json.Marshal(audit)
	if err != nil {
		return nil, errWorkflowRAGEvaluationContract
	}
	return payload, nil
}

func encodeWorkflowRAGEvaluationCreatePayloads(
	ctx WorkflowRAGSnapshotContext,
	resource WorkflowRAGEvaluationDatasetResource,
	version WorkflowRAGEvaluationDatasetVersion,
	audit WorkflowRAGEvaluationAudit,
) ([]byte, []byte, []byte, error) {
	if !workflowRAGEvaluationResourceVersionMatch(resource, version) || !workflowRAGEvaluationAuditVersionMatch(audit, version) {
		return nil, nil, nil, errWorkflowRAGEvaluationContract
	}
	resourcePayload, err := encodeWorkflowRAGEvaluationResource(resource, ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	versionPayload, err := encodeWorkflowRAGEvaluationVersion(version, ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	auditPayload, err := encodeWorkflowRAGEvaluationAudit(audit, ctx)
	return resourcePayload, versionPayload, auditPayload, err
}

func workflowRAGEvaluationResourceVersionMatch(resource WorkflowRAGEvaluationDatasetResource, version WorkflowRAGEvaluationDatasetVersion) bool {
	return resource.DatasetID == version.Dataset.DatasetID && resource.DatasetKey == version.DatasetKey &&
		resource.DisplayName == version.DisplayName && resource.LifecycleState == version.LifecycleState &&
		resource.ContentClassification == version.Dataset.ContentClassification && resource.LatestVersion == version.Dataset.DatasetVersion &&
		resource.LatestDigest == version.Dataset.DatasetDigest && resource.BaselineSnapshot == version.Dataset.Snapshot &&
		resource.SampleCount == len(version.Dataset.Samples)
}

func workflowRAGEvaluationAuditVersionMatch(audit WorkflowRAGEvaluationAudit, version WorkflowRAGEvaluationDatasetVersion) bool {
	return audit.DatasetID == version.Dataset.DatasetID && audit.DatasetVersion == version.Dataset.DatasetVersion &&
		audit.DatasetDigest == version.Dataset.DatasetDigest && audit.ContentClassification == version.Dataset.ContentClassification &&
		audit.BaselineSnapshot == version.Dataset.Snapshot && audit.SampleCount == len(version.Dataset.Samples)
}

func workflowRAGEvaluationAuditResourceMatch(audit WorkflowRAGEvaluationAudit, resource WorkflowRAGEvaluationDatasetResource) bool {
	return audit.DatasetID == resource.DatasetID && audit.DatasetVersion == resource.LatestVersion &&
		audit.DatasetDigest == resource.LatestDigest && audit.ContentClassification == resource.ContentClassification &&
		audit.BaselineSnapshot == resource.BaselineSnapshot && audit.SampleCount == resource.SampleCount
}

func workflowRAGEvaluationAuditReviewMatch(audit WorkflowRAGEvaluationAudit, review WorkflowRAGCandidateSnapshotReview) bool {
	return audit.EventKind == "candidate_review_created" && audit.DatasetID == review.Dataset.DatasetID &&
		audit.DatasetVersion == review.Dataset.DatasetVersion && audit.DatasetDigest == review.Dataset.DatasetDigest &&
		audit.BaselineSnapshot == review.BaselineSnapshot && audit.SampleCount == review.Baseline.SampleCount
}

func workflowRAGEvaluationUnixNano(value string) (int64, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return 0, errWorkflowRAGEvaluationContract
	}
	return parsed.UnixNano(), nil
}

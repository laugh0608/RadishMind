package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
)

func newWorkflowDefinitionReleaseRepositoryForRunStore(store workflowRunStore) (workflowDefinitionReleaseRepository, error) {
	switch typed := store.(type) {
	case *memoryWorkflowRunStore:
		return newWorkflowDefinitionReleaseStore(), nil
	case *sqliteWorkflowRunStore:
		if typed.database == nil {
			return nil, errors.New("workflow definition release requires the shared SQLite database")
		}
		return newSQLiteWorkflowDefinitionReleaseRepository(typed.database), nil
	case *postgresWorkflowRunStore:
		if typed.pool == nil {
			return nil, errors.New("workflow definition release requires the workflow PostgreSQL pool")
		}
		return newPostgresWorkflowDefinitionReleaseRepository(typed.pool), nil
	default:
		return nil, errors.New("workflow definition release requires a supported workflow runtime backend")
	}
}

func workflowDefinitionRequestContext(value WorkflowDefinitionReleaseContext) context.Context {
	if value.RequestContext != nil {
		return value.RequestContext
	}
	return context.Background()
}

func encodeWorkflowDefinitionRecord(value any) ([]byte, error) {
	payload, err := json.Marshal(value)
	if err != nil || applicationDraftStringContainsSecret(string(payload)) {
		return nil, errWorkflowDefinitionStore
	}
	return payload, nil
}

func decodeWorkflowDefinitionRecord(payload []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return errWorkflowDefinitionStore
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errWorkflowDefinitionStore
	}
	return nil
}

func workflowDefinitionUnixNano(value string) (int64, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return 0, err
	}
	return parsed.UnixNano(), nil
}

func validateStoredWorkflowDefinitionCandidate(value WorkflowDefinitionReleaseCandidate) error {
	created, createdErr := time.Parse(time.RFC3339Nano, value.CreatedAt)
	updated, updatedErr := time.Parse(time.RFC3339Nano, value.UpdatedAt)
	if value.SchemaVersion != workflowDefinitionCandidateSchemaVersion || !applicationDraftIdentifierPattern.MatchString(value.CandidateID) ||
		!applicationDraftIdentifierPattern.MatchString(value.DefinitionID) || !applicationDraftIdentifierPattern.MatchString(value.SourceDraftID) ||
		value.SourceDraftVersion < 1 || !workflowRAGDigestPattern.MatchString(value.SourceDraftDigest) || value.DefinitionDigest != value.SourceDraftDigest ||
		(value.State != workflowDefinitionStatePending && value.State != workflowDefinitionStateApproved && value.State != workflowDefinitionStateRejected) ||
		value.ReviewVersion != len(value.Reviews) || createdErr != nil || updatedErr != nil || updated.Before(created) ||
		strings.TrimSpace(value.CreatedByActorRef) == "" || strings.TrimSpace(value.UpdatedByActorRef) == "" || strings.TrimSpace(value.RequestID) == "" || strings.TrimSpace(value.AuditRef) == "" ||
		workflowDefinitionSnapshotContainsForbiddenMaterial(value.Snapshot) {
		return errWorkflowDefinitionStore
	}
	for index, review := range value.Reviews {
		if review.SchemaVersion != workflowDefinitionDecisionSchemaVersion || review.ReviewVersion != index+1 ||
			(review.Decision != "approve" && review.Decision != "reject") || !validWorkflowDefinitionReason(review.Reason) ||
			strings.TrimSpace(review.ReviewerRef) == "" || strings.TrimSpace(review.RequestID) == "" || strings.TrimSpace(review.AuditRef) == "" {
			return errWorkflowDefinitionStore
		}
		if _, err := time.Parse(time.RFC3339Nano, review.ReviewedAt); err != nil {
			return errWorkflowDefinitionStore
		}
	}
	if value.State == workflowDefinitionStatePending && value.ReviewVersion != 0 || value.State != workflowDefinitionStatePending && value.ReviewVersion != 1 {
		return errWorkflowDefinitionStore
	}
	return nil
}

func validateStoredWorkflowDefinitionVersion(value WorkflowDefinitionVersion) error {
	if value.SchemaVersion != workflowDefinitionVersionSchemaVersion || !applicationDraftIdentifierPattern.MatchString(value.DefinitionID) || value.Version < 1 ||
		!workflowRAGDigestPattern.MatchString(value.DefinitionDigest) || !applicationDraftIdentifierPattern.MatchString(value.CandidateID) || value.CandidateReviewVersion < 1 ||
		!applicationDraftIdentifierPattern.MatchString(value.SourceDraftID) || value.SourceDraftVersion < 1 || !workflowRAGDigestPattern.MatchString(value.SourceDraftDigest) ||
		value.DefinitionDigest != value.SourceDraftDigest || strings.TrimSpace(value.CreatedByActorRef) == "" || strings.TrimSpace(value.RequestID) == "" || strings.TrimSpace(value.AuditRef) == "" ||
		workflowDefinitionSnapshotContainsForbiddenMaterial(value.Snapshot) {
		return errWorkflowDefinitionStore
	}
	if _, err := time.Parse(time.RFC3339Nano, value.CreatedAt); err != nil {
		return errWorkflowDefinitionStore
	}
	return nil
}

func validateStoredWorkflowDefinitionActivation(value WorkflowDefinitionActivation) error {
	if value.SchemaVersion != workflowDefinitionActivationSchemaVersion || !applicationDraftIdentifierPattern.MatchString(value.DefinitionID) || value.PointerVersion != len(value.Events) || value.PointerVersion < 1 ||
		(value.State != workflowDefinitionActivationActive && value.State != workflowDefinitionActivationInactive) ||
		(value.State == workflowDefinitionActivationActive && (value.ActiveVersion < 1 || !workflowRAGDigestPattern.MatchString(value.ActiveDefinitionDigest))) ||
		(value.State == workflowDefinitionActivationInactive && (value.ActiveVersion != 0 || value.ActiveDefinitionDigest != "")) ||
		strings.TrimSpace(value.UpdatedByActorRef) == "" || strings.TrimSpace(value.RequestID) == "" || strings.TrimSpace(value.AuditRef) == "" {
		return errWorkflowDefinitionStore
	}
	if _, err := time.Parse(time.RFC3339Nano, value.UpdatedAt); err != nil {
		return errWorkflowDefinitionStore
	}
	for index, event := range value.Events {
		if event.SchemaVersion != workflowDefinitionActivationEventSchemaVersion || event.DefinitionID != value.DefinitionID || event.AfterPointerVersion != index+1 || event.BeforePointerVersion != index ||
			(event.Decision != "activate" && event.Decision != "replace" && event.Decision != "deactivate") || !validWorkflowDefinitionReason(event.Reason) ||
			strings.TrimSpace(event.EventID) == "" || strings.TrimSpace(event.ActorRef) == "" || strings.TrimSpace(event.RequestID) == "" || strings.TrimSpace(event.AuditRef) == "" {
			return errWorkflowDefinitionStore
		}
		if _, err := time.Parse(time.RFC3339Nano, event.CreatedAt); err != nil {
			return errWorkflowDefinitionStore
		}
	}
	return nil
}

func validateStoredWorkflowDefinitionAudit(value WorkflowDefinitionReleaseAudit) error {
	if value.SchemaVersion != workflowDefinitionAuditSchemaVersion || strings.TrimSpace(value.AuditID) == "" ||
		(value.ResourceKind != "candidate" && value.ResourceKind != "version" && value.ResourceKind != "activation") || strings.TrimSpace(value.ResourceID) == "" || strings.TrimSpace(value.Action) == "" ||
		strings.TrimSpace(value.ActorRef) == "" || strings.TrimSpace(value.RequestID) == "" || strings.TrimSpace(value.AuditRef) == "" {
		return errWorkflowDefinitionStore
	}
	if _, err := time.Parse(time.RFC3339Nano, value.CreatedAt); err != nil {
		return errWorkflowDefinitionStore
	}
	return nil
}

func loadWorkflowDefinitionCandidateIntoStore(store *workflowDefinitionReleaseStore, ctx WorkflowDefinitionReleaseContext, payload []byte) (WorkflowDefinitionReleaseCandidate, error) {
	var value WorkflowDefinitionReleaseCandidate
	if decodeWorkflowDefinitionRecord(payload, &value) != nil || validateStoredWorkflowDefinitionCandidate(value) != nil {
		return WorkflowDefinitionReleaseCandidate{}, errWorkflowDefinitionStore
	}
	store.candidates[workflowDefinitionScopeKey(ctx, value.CandidateID)] = value
	return value, nil
}

func loadWorkflowDefinitionVersionIntoStore(store *workflowDefinitionReleaseStore, ctx WorkflowDefinitionReleaseContext, payload []byte) (WorkflowDefinitionVersion, error) {
	var value WorkflowDefinitionVersion
	if decodeWorkflowDefinitionRecord(payload, &value) != nil || validateStoredWorkflowDefinitionVersion(value) != nil {
		return WorkflowDefinitionVersion{}, errWorkflowDefinitionStore
	}
	key := workflowDefinitionScopeKey(ctx, value.DefinitionID)
	if value.Version != len(store.versions[key])+1 {
		return WorkflowDefinitionVersion{}, errWorkflowDefinitionStore
	}
	store.versions[key] = append(store.versions[key], value)
	return value, nil
}

func loadWorkflowDefinitionActivationIntoStore(store *workflowDefinitionReleaseStore, ctx WorkflowDefinitionReleaseContext, payload []byte) (WorkflowDefinitionActivation, error) {
	var value WorkflowDefinitionActivation
	if decodeWorkflowDefinitionRecord(payload, &value) != nil || validateStoredWorkflowDefinitionActivation(value) != nil {
		return WorkflowDefinitionActivation{}, errWorkflowDefinitionStore
	}
	store.activations[workflowDefinitionScopeKey(ctx, value.DefinitionID)] = value
	return value, nil
}

func loadWorkflowDefinitionAuditIntoStore(store *workflowDefinitionReleaseStore, ctx WorkflowDefinitionReleaseContext, payload []byte) (WorkflowDefinitionReleaseAudit, error) {
	var value WorkflowDefinitionReleaseAudit
	if decodeWorkflowDefinitionRecord(payload, &value) != nil || validateStoredWorkflowDefinitionAudit(value) != nil {
		return WorkflowDefinitionReleaseAudit{}, errWorkflowDefinitionStore
	}
	key := workflowDefinitionScopeKey(ctx, value.ResourceKind)
	store.audits[key] = append(store.audits[key], value)
	return value, nil
}

package httpapi

import (
	"errors"
	"strings"
	"time"
)

const (
	workflowDefinitionFailureScopeDenied        = "workflow_definition_scope_denied"
	workflowDefinitionFailurePayloadInvalid     = "workflow_definition_payload_invalid"
	workflowDefinitionFailureSourceNotFound     = "workflow_definition_source_draft_not_found"
	workflowDefinitionFailureSourceVersionDrift = "workflow_definition_source_draft_version_drift"
	workflowDefinitionFailureSourceDigestDrift  = "workflow_definition_source_draft_digest_drift"
	workflowDefinitionFailureSourceIneligible   = "workflow_definition_source_draft_ineligible"
	workflowDefinitionFailureNotFound           = "workflow_definition_not_found"
	workflowDefinitionFailureVersionConflict    = "workflow_definition_version_conflict"
	workflowDefinitionFailureTransitionInvalid  = "workflow_definition_transition_invalid"
	workflowDefinitionFailureStoreUnavailable   = "workflow_definition_store_unavailable"
)

type WorkflowDefinitionCandidateCreateInput struct {
	CandidateID          string
	DefinitionID         string
	DraftID              string
	ExpectedDraftVersion int
}

type WorkflowDefinitionReviewInput struct {
	ExpectedReviewVersion int
	Decision              string
	Reason                string
}

type WorkflowDefinitionActivationInput struct {
	ExpectedPointerVersion int
	Decision               string
	Version                int
	Reason                 string
}

type WorkflowDefinitionReleaseResult struct {
	Candidate             *WorkflowDefinitionReleaseCandidate
	Version               *WorkflowDefinitionVersion
	Activation            *WorkflowDefinitionActivation
	FailureCode           string
	CurrentReviewVersion  int
	CurrentPointerVersion int
}

type workflowDefinitionReleaseService struct {
	drafts savedWorkflowDraftService
	store  workflowDefinitionReleaseRepository
	now    func() time.Time
}

func newWorkflowDefinitionReleaseService(draftStore savedWorkflowDraftStore, store workflowDefinitionReleaseRepository) workflowDefinitionReleaseService {
	return workflowDefinitionReleaseService{
		drafts: newSavedWorkflowDraftService(draftStore),
		store:  store,
		now:    func() time.Time { return time.Now().UTC() },
	}
}

func (service workflowDefinitionReleaseService) Create(ctx WorkflowDefinitionReleaseContext, input WorkflowDefinitionCandidateCreateInput) WorkflowDefinitionReleaseResult {
	input.CandidateID = strings.TrimSpace(input.CandidateID)
	input.DefinitionID = strings.TrimSpace(input.DefinitionID)
	input.DraftID = strings.TrimSpace(input.DraftID)
	if !validWorkflowDefinitionContext(ctx) || !applicationDraftIdentifierPattern.MatchString(input.CandidateID) ||
		!applicationDraftIdentifierPattern.MatchString(input.DefinitionID) || !applicationDraftIdentifierPattern.MatchString(input.DraftID) || input.ExpectedDraftVersion < 1 {
		return WorkflowDefinitionReleaseResult{FailureCode: workflowDefinitionFailurePayloadInvalid}
	}
	draft, failureCode := service.readExactDraft(ctx, input.DraftID, input.ExpectedDraftVersion)
	if failureCode != "" {
		return WorkflowDefinitionReleaseResult{FailureCode: failureCode}
	}
	candidate, err := service.store.CreateCandidate(ctx, input.CandidateID, input.DefinitionID, draft, service.now())
	if err != nil {
		return workflowDefinitionResultFromError(err)
	}
	return WorkflowDefinitionReleaseResult{Candidate: &candidate, CurrentReviewVersion: candidate.ReviewVersion}
}

func (service workflowDefinitionReleaseService) Review(ctx WorkflowDefinitionReleaseContext, candidateID string, input WorkflowDefinitionReviewInput) WorkflowDefinitionReleaseResult {
	candidateID = strings.TrimSpace(candidateID)
	decision := strings.TrimSpace(input.Decision)
	reason := strings.TrimSpace(input.Reason)
	if !applicationDraftIdentifierPattern.MatchString(candidateID) || input.ExpectedReviewVersion < 0 ||
		(decision != "approve" && decision != "reject") || !validWorkflowDefinitionReason(reason) {
		return WorkflowDefinitionReleaseResult{FailureCode: workflowDefinitionFailurePayloadInvalid}
	}
	candidate, err := service.store.ReadCandidate(ctx, candidateID)
	if err != nil {
		return workflowDefinitionResultFromError(err)
	}
	draft, failureCode := service.readExactDraft(ctx, candidate.SourceDraftID, candidate.SourceDraftVersion)
	if failureCode != "" {
		if failureCode == workflowDefinitionFailureSourceVersionDrift || failureCode == workflowDefinitionFailureSourceNotFound {
			failureCode = workflowDefinitionFailureSourceDigestDrift
		}
		return WorkflowDefinitionReleaseResult{Candidate: &candidate, FailureCode: failureCode, CurrentReviewVersion: candidate.ReviewVersion}
	}
	_, sourceDigest, digestErr := workflowDefinitionSnapshotFromDraft(draft)
	if digestErr != nil {
		return WorkflowDefinitionReleaseResult{Candidate: &candidate, FailureCode: workflowDefinitionFailureStoreUnavailable, CurrentReviewVersion: candidate.ReviewVersion}
	}
	updated, version, err := service.store.Review(ctx, candidateID, input.ExpectedReviewVersion, decision, reason, sourceDigest, service.now())
	if err != nil {
		result := workflowDefinitionResultFromError(err)
		result.CurrentReviewVersion = candidate.ReviewVersion
		return result
	}
	return WorkflowDefinitionReleaseResult{Candidate: &updated, Version: version, CurrentReviewVersion: updated.ReviewVersion}
}

func (service workflowDefinitionReleaseService) ReadCandidate(ctx WorkflowDefinitionReleaseContext, candidateID string) WorkflowDefinitionReleaseResult {
	if !applicationDraftIdentifierPattern.MatchString(strings.TrimSpace(candidateID)) {
		return WorkflowDefinitionReleaseResult{FailureCode: workflowDefinitionFailurePayloadInvalid}
	}
	candidate, err := service.store.ReadCandidate(ctx, strings.TrimSpace(candidateID))
	if err != nil {
		return workflowDefinitionResultFromError(err)
	}
	return WorkflowDefinitionReleaseResult{Candidate: &candidate, CurrentReviewVersion: candidate.ReviewVersion}
}

func (service workflowDefinitionReleaseService) ListCandidates(ctx WorkflowDefinitionReleaseContext) ([]WorkflowDefinitionReleaseCandidate, string) {
	values, err := service.store.ListCandidates(ctx)
	if err != nil {
		return nil, workflowDefinitionResultFromError(err).FailureCode
	}
	return values, ""
}

func (service workflowDefinitionReleaseService) ListVersions(ctx WorkflowDefinitionReleaseContext, definitionID string) ([]WorkflowDefinitionVersion, string) {
	if !applicationDraftIdentifierPattern.MatchString(strings.TrimSpace(definitionID)) {
		return nil, workflowDefinitionFailurePayloadInvalid
	}
	values, err := service.store.ListVersions(ctx, strings.TrimSpace(definitionID))
	if err != nil {
		return nil, workflowDefinitionResultFromError(err).FailureCode
	}
	return values, ""
}

func (service workflowDefinitionReleaseService) ReadVersion(ctx WorkflowDefinitionReleaseContext, definitionID string, version int) WorkflowDefinitionReleaseResult {
	if !applicationDraftIdentifierPattern.MatchString(strings.TrimSpace(definitionID)) || version < 1 {
		return WorkflowDefinitionReleaseResult{FailureCode: workflowDefinitionFailurePayloadInvalid}
	}
	value, err := service.store.ReadVersion(ctx, strings.TrimSpace(definitionID), version)
	if err != nil {
		return workflowDefinitionResultFromError(err)
	}
	return WorkflowDefinitionReleaseResult{Version: &value}
}

func (service workflowDefinitionReleaseService) ReadActivation(ctx WorkflowDefinitionReleaseContext, definitionID string) WorkflowDefinitionReleaseResult {
	if !applicationDraftIdentifierPattern.MatchString(strings.TrimSpace(definitionID)) {
		return WorkflowDefinitionReleaseResult{FailureCode: workflowDefinitionFailurePayloadInvalid}
	}
	value, err := service.store.ReadActivation(ctx, strings.TrimSpace(definitionID))
	if err != nil {
		return workflowDefinitionResultFromError(err)
	}
	return WorkflowDefinitionReleaseResult{Activation: &value, CurrentPointerVersion: value.PointerVersion}
}

func (service workflowDefinitionReleaseService) DecideActivation(ctx WorkflowDefinitionReleaseContext, definitionID string, input WorkflowDefinitionActivationInput) WorkflowDefinitionReleaseResult {
	value, err := service.store.DecideActivation(ctx, strings.TrimSpace(definitionID), input.ExpectedPointerVersion, strings.TrimSpace(input.Decision), input.Version, strings.TrimSpace(input.Reason), service.now())
	if err != nil {
		result := workflowDefinitionResultFromError(err)
		current, readErr := service.store.ReadActivation(ctx, strings.TrimSpace(definitionID))
		if readErr == nil {
			result.CurrentPointerVersion = current.PointerVersion
		}
		return result
	}
	return WorkflowDefinitionReleaseResult{Activation: &value, CurrentPointerVersion: value.PointerVersion}
}

func (service workflowDefinitionReleaseService) readExactDraft(ctx WorkflowDefinitionReleaseContext, draftID string, expectedVersion int) (SavedWorkflowDraft, string) {
	draftContext := SavedWorkflowDraftContext{
		RequestID:       ctx.RequestID,
		TenantRef:       ctx.TenantRef,
		WorkspaceID:     ctx.WorkspaceID,
		ApplicationID:   ctx.ApplicationID,
		ActorRef:        ctx.ActorRef,
		OwnerSubjectRef: ctx.OwnerSubjectRef,
		AuditRef:        ctx.AuditRef,
	}
	result := service.drafts.ReadDraft(draftContext, ReadWorkflowDraftRequest{DraftID: draftID})
	if result.FailureCode == SavedWorkflowDraftFailureNotFound {
		return SavedWorkflowDraft{}, workflowDefinitionFailureSourceNotFound
	}
	if result.FailureCode != "" || result.Draft == nil {
		if result.FailureCode == SavedWorkflowDraftFailureBlockedCapability || result.FailureCode == SavedWorkflowDraftFailureGraphInvalid ||
			result.FailureCode == SavedWorkflowDraftFailureContractInvalid || result.FailureCode == SavedWorkflowDraftFailurePayloadInvalid {
			return SavedWorkflowDraft{}, workflowDefinitionFailureSourceIneligible
		}
		return SavedWorkflowDraft{}, workflowDefinitionFailureStoreUnavailable
	}
	if result.Draft.DraftVersion != expectedVersion {
		return SavedWorkflowDraft{}, workflowDefinitionFailureSourceVersionDrift
	}
	if result.Draft.DraftStatus != SavedWorkflowDraftStatusValidForReview || !result.Draft.ValidationSummary.ValidForReview || len(result.Draft.BlockedCapabilitySummary) > 0 {
		return SavedWorkflowDraft{}, workflowDefinitionFailureSourceIneligible
	}
	return *result.Draft, ""
}

func workflowDefinitionResultFromError(err error) WorkflowDefinitionReleaseResult {
	switch {
	case errors.Is(err, errWorkflowDefinitionNotFound):
		return WorkflowDefinitionReleaseResult{FailureCode: workflowDefinitionFailureNotFound}
	case errors.Is(err, errWorkflowDefinitionConflict):
		return WorkflowDefinitionReleaseResult{FailureCode: workflowDefinitionFailureVersionConflict}
	case errors.Is(err, errWorkflowDefinitionInvalidState):
		return WorkflowDefinitionReleaseResult{FailureCode: workflowDefinitionFailureTransitionInvalid}
	case errors.Is(err, errWorkflowDefinitionPayloadInvalid):
		return WorkflowDefinitionReleaseResult{FailureCode: workflowDefinitionFailurePayloadInvalid}
	case errors.Is(err, errWorkflowDefinitionSourceDrift):
		return WorkflowDefinitionReleaseResult{FailureCode: workflowDefinitionFailureSourceDigestDrift}
	default:
		return WorkflowDefinitionReleaseResult{FailureCode: workflowDefinitionFailureStoreUnavailable}
	}
}

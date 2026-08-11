package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
)

const (
	applicationInteractionAuthoritySchemaVersion           = "application_runtime_authority.v1"
	applicationInteractionStructuredAuthoritySchemaVersion = "application_runtime_authority.v4"
	applicationInteractionProfileWorkflow                  = workflowDefinitionExecutorProfile
	applicationInteractionProfileWorkflowStructured        = workflowDefinitionStructuredExecutorProfile
	applicationInteractionProfileRAG                       = "application_rag_invocation_v1"
	applicationInteractionProfilePrompt                    = promptApplicationInvocationProfile
	applicationInteractionProfileAgentCopilot              = agentCopilotSuggestionProfile
)

const (
	ApplicationInteractionFailureScopeDenied         = "application_session_scope_denied"
	ApplicationInteractionFailurePayloadInvalid      = "application_session_payload_invalid"
	ApplicationInteractionFailureApplicationMissing  = "application_session_application_not_found"
	ApplicationInteractionFailureApplicationArchived = "application_session_application_archived"
	ApplicationInteractionFailureAuthorityMissing    = "application_session_authority_not_found"
	ApplicationInteractionFailureAuthorityChanged    = "application_session_authority_changed"
	ApplicationInteractionFailureProfileIneligible   = "application_session_profile_ineligible"
	ApplicationInteractionFailureStoreUnavailable    = "application_session_store_unavailable"
)

type ApplicationInteractionProfileBinding struct {
	ExecutionProfile string `json:"execution_profile"`
	DefinitionID     string `json:"definition_id,omitempty"`
}

type ApplicationInteractionWorkflowAuthority struct {
	DefinitionID             string                      `json:"definition_id"`
	DefinitionVersion        int                         `json:"definition_version"`
	DefinitionDigest         string                      `json:"definition_digest"`
	ActivationPointerVersion int                         `json:"activation_pointer_version"`
	CandidateID              string                      `json:"candidate_id"`
	CandidateReviewVersion   int                         `json:"candidate_review_version"`
	InputContract            *WorkflowDefinitionContract `json:"input_contract,omitempty"`
}

type ApplicationInteractionRAGAuthority struct {
	AssignmentID            string                           `json:"assignment_id"`
	AssignmentVersion       int                              `json:"assignment_version"`
	AssignmentDigest        string                           `json:"assignment_digest"`
	PublishCandidateID      string                           `json:"publish_candidate_id"`
	PublishReviewVersion    int                              `json:"publish_review_version"`
	DraftID                 string                           `json:"draft_id"`
	DraftVersion            int                              `json:"draft_version"`
	DraftDigest             string                           `json:"draft_digest"`
	BindingRef              WorkflowRAGApplicationBindingRef `json:"binding_ref"`
	SnapshotID              string                           `json:"snapshot_id"`
	SnapshotVersion         int                              `json:"snapshot_version"`
	SnapshotDigest          string                           `json:"snapshot_digest"`
	RetrievalProfileID      string                           `json:"retrieval_profile_id"`
	RetrievalProfileVersion int                              `json:"retrieval_profile_version"`
	RetrievalProfileDigest  string                           `json:"retrieval_profile_digest"`
}

type ApplicationInteractionAuthoritySnapshot struct {
	SchemaVersion            string                                   `json:"schema_version"`
	ExecutionProfile         string                                   `json:"execution_profile"`
	ApplicationID            string                                   `json:"application_id"`
	ApplicationRecordVersion int                                      `json:"application_record_version"`
	ApplicationLifecycle     string                                   `json:"application_lifecycle"`
	WorkflowDefinition       *ApplicationInteractionWorkflowAuthority `json:"workflow_definition"`
	ApplicationRAG           *ApplicationInteractionRAGAuthority      `json:"application_rag"`
	PromptApplication        *PromptApplicationAuthorityV2            `json:"-"`
	AgentCopilot             *AgentCopilotAuthorityV3                 `json:"-"`
	AuthorityDigest          string                                   `json:"authority_digest"`
}

type applicationInteractionAuthorityResolver interface {
	Resolve(ApplicationInteractionContext, ApplicationInteractionProfileBinding) (ApplicationInteractionAuthoritySnapshot, string)
}

type exactApplicationInteractionAuthorityResolver struct {
	applications         applicationCatalogRepository
	definitions          workflowDefinitionReleaseRepository
	ragRuntime           workflowRAGApplicationRuntimeRepository
	ragAuthorityResolver workflowRAGApplicationAuthorityResolver
	resolvePrompt        func(PromptApplicationRuntimeContext) (PromptApplicationRuntimeAuthorityV2, string)
	resolveAgentCopilot  func(AgentCopilotRuntimeContext) (AgentCopilotRuntimeAuthorityV3, string)
}

func newExactApplicationInteractionAuthorityResolver(applications applicationCatalogRepository, definitions workflowDefinitionReleaseRepository, ragRuntime workflowRAGApplicationRuntimeRepository, ragResolver workflowRAGApplicationAuthorityResolver) exactApplicationInteractionAuthorityResolver {
	return exactApplicationInteractionAuthorityResolver{applications: applications, definitions: definitions, ragRuntime: ragRuntime, ragAuthorityResolver: ragResolver}
}

func (resolver exactApplicationInteractionAuthorityResolver) Resolve(ctx ApplicationInteractionContext, binding ApplicationInteractionProfileBinding) (ApplicationInteractionAuthoritySnapshot, string) {
	if validateApplicationInteractionContext(ctx) != nil || validateApplicationInteractionProfileBinding(binding) != nil {
		return ApplicationInteractionAuthoritySnapshot{}, ApplicationInteractionFailureScopeDenied
	}
	if resolver.applications == nil {
		return ApplicationInteractionAuthoritySnapshot{}, ApplicationInteractionFailureStoreUnavailable
	}
	applicationContext := ApplicationCatalogContext{RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ActorRef: ctx.ActorRef, OwnerSubjectRef: ctx.OwnerSubjectRef, AuditRef: ctx.AuditRef}
	application, err := resolver.applications.RequireActive(applicationContext, ctx.ApplicationID)
	if errors.Is(err, errApplicationCatalogNotFound) {
		return ApplicationInteractionAuthoritySnapshot{}, ApplicationInteractionFailureApplicationMissing
	}
	if errors.Is(err, errApplicationCatalogArchived) {
		return ApplicationInteractionAuthoritySnapshot{}, ApplicationInteractionFailureApplicationArchived
	}
	if err != nil {
		return ApplicationInteractionAuthoritySnapshot{}, ApplicationInteractionFailureStoreUnavailable
	}
	authoritySchemaVersion := applicationInteractionAuthoritySchemaVersion
	if binding.ExecutionProfile == applicationInteractionProfileWorkflowStructured {
		authoritySchemaVersion = applicationInteractionStructuredAuthoritySchemaVersion
	}
	snapshot := ApplicationInteractionAuthoritySnapshot{SchemaVersion: authoritySchemaVersion, ExecutionProfile: binding.ExecutionProfile, ApplicationID: application.ApplicationID, ApplicationRecordVersion: application.RecordVersion, ApplicationLifecycle: application.LifecycleState}
	var failure string
	switch binding.ExecutionProfile {
	case applicationInteractionProfileWorkflow, applicationInteractionProfileWorkflowStructured:
		snapshot.WorkflowDefinition, failure = resolver.resolveWorkflowDefinition(ctx, binding.DefinitionID, binding.ExecutionProfile)
	case applicationInteractionProfileRAG:
		snapshot.ApplicationRAG, failure = resolver.resolveApplicationRAG(ctx)
	case applicationInteractionProfilePrompt:
		var prompt PromptApplicationRuntimeAuthorityV2
		prompt, failure = resolver.resolvePromptAuthority(ctx)
		if failure == "" {
			snapshot = applicationInteractionAuthorityFromPrompt(prompt)
		}
	case applicationInteractionProfileAgentCopilot:
		var authority AgentCopilotRuntimeAuthorityV3
		authority, failure = resolver.resolveAgentCopilotAuthority(ctx)
		if failure == "" {
			snapshot = applicationInteractionAuthorityFromAgentCopilot(authority)
		}
	default:
		failure = ApplicationInteractionFailureProfileIneligible
	}
	if failure != "" {
		return ApplicationInteractionAuthoritySnapshot{}, failure
	}
	digest, err := applicationInteractionAuthorityDigest(snapshot)
	if err != nil {
		return ApplicationInteractionAuthoritySnapshot{}, ApplicationInteractionFailureStoreUnavailable
	}
	snapshot.AuthorityDigest = digest
	if validateApplicationInteractionAuthority(snapshot) != nil {
		return ApplicationInteractionAuthoritySnapshot{}, ApplicationInteractionFailureStoreUnavailable
	}
	return snapshot, ""
}

func (resolver exactApplicationInteractionAuthorityResolver) resolveAgentCopilotAuthority(ctx ApplicationInteractionContext) (AgentCopilotRuntimeAuthorityV3, string) {
	if resolver.resolveAgentCopilot == nil {
		return AgentCopilotRuntimeAuthorityV3{}, ApplicationInteractionFailureAuthorityMissing
	}
	authority, failure := resolver.resolveAgentCopilot(AgentCopilotRuntimeContext{
		RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef,
		WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef,
		OwnerSubjectRef: ctx.OwnerSubjectRef, AuditRef: ctx.AuditRef,
	})
	if failure == "" {
		return authority, ""
	}
	switch failure {
	case AgentCopilotRuntimeFailureNotFound:
		return AgentCopilotRuntimeAuthorityV3{}, ApplicationInteractionFailureAuthorityMissing
	case AgentCopilotRuntimeFailureStoreUnavailable:
		return AgentCopilotRuntimeAuthorityV3{}, ApplicationInteractionFailureStoreUnavailable
	default:
		return AgentCopilotRuntimeAuthorityV3{}, ApplicationInteractionFailureAuthorityChanged
	}
}

func (resolver exactApplicationInteractionAuthorityResolver) resolvePromptAuthority(ctx ApplicationInteractionContext) (PromptApplicationRuntimeAuthorityV2, string) {
	if resolver.resolvePrompt == nil {
		return PromptApplicationRuntimeAuthorityV2{}, ApplicationInteractionFailureAuthorityMissing
	}
	runtimeContext := PromptApplicationRuntimeContext{
		RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef,
		WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef,
		OwnerSubjectRef: ctx.OwnerSubjectRef, AuditRef: ctx.AuditRef,
	}
	authority, failure := resolver.resolvePrompt(runtimeContext)
	if failure == "" {
		return authority, ""
	}
	switch failure {
	case PromptApplicationRuntimeFailureNotFound:
		return PromptApplicationRuntimeAuthorityV2{}, ApplicationInteractionFailureAuthorityMissing
	case PromptApplicationRuntimeFailureStoreUnavailable:
		return PromptApplicationRuntimeAuthorityV2{}, ApplicationInteractionFailureStoreUnavailable
	default:
		return PromptApplicationRuntimeAuthorityV2{}, ApplicationInteractionFailureAuthorityChanged
	}
}

func (resolver exactApplicationInteractionAuthorityResolver) resolveWorkflowDefinition(ctx ApplicationInteractionContext, definitionID, executionProfile string) (*ApplicationInteractionWorkflowAuthority, string) {
	if resolver.definitions == nil {
		return nil, ApplicationInteractionFailureAuthorityMissing
	}
	releaseContext := WorkflowDefinitionReleaseContext{RequestContext: ctx.RequestContext, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, OwnerSubjectRef: ctx.OwnerSubjectRef, ActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef}
	activation, err := resolver.definitions.ReadActivation(releaseContext, definitionID)
	if err != nil || activation.State != workflowDefinitionActivationActive || activation.PointerVersion < 1 || activation.ActiveVersion < 1 || !workflowRAGDigestPattern.MatchString(activation.ActiveDefinitionDigest) {
		return nil, ApplicationInteractionFailureAuthorityMissing
	}
	version, err := resolver.definitions.ReadVersion(releaseContext, definitionID, activation.ActiveVersion)
	if err != nil || validateStoredWorkflowDefinitionVersion(version) != nil {
		return nil, ApplicationInteractionFailureAuthorityMissing
	}
	digest, err := workflowDefinitionSnapshotDigest(version.Snapshot)
	if err != nil || digest != version.DefinitionDigest || digest != activation.ActiveDefinitionDigest {
		return nil, ApplicationInteractionFailureAuthorityChanged
	}
	if !version.ActivationEligible || len(version.EligibilityBlockers) != 0 || version.Snapshot.ExecutionProfile != executionProfile {
		return nil, ApplicationInteractionFailureProfileIneligible
	}
	authority := &ApplicationInteractionWorkflowAuthority{DefinitionID: version.DefinitionID, DefinitionVersion: version.Version, DefinitionDigest: version.DefinitionDigest, ActivationPointerVersion: activation.PointerVersion, CandidateID: version.CandidateID, CandidateReviewVersion: version.CandidateReviewVersion}
	if executionProfile == applicationInteractionProfileWorkflowStructured {
		normalized, code, _ := normalizeWorkflowStructuredInputContract(savedWorkflowDraftContractFromDefinition(version.Snapshot.InputContract))
		if code != "" || normalized.ContractDigest != version.Snapshot.InputContract.ContractDigest {
			return nil, ApplicationInteractionFailureAuthorityChanged
		}
		contract := WorkflowDefinitionContract{ContractID: normalized.ContractID, Fields: cloneWorkflowStructuredInputFields(normalized.Fields), Summary: normalized.Summary, ContractDigest: normalized.ContractDigest}
		authority.InputContract = &contract
	}
	return authority, ""
}

func (resolver exactApplicationInteractionAuthorityResolver) resolveApplicationRAG(ctx ApplicationInteractionContext) (*ApplicationInteractionRAGAuthority, string) {
	if resolver.ragRuntime == nil {
		return nil, ApplicationInteractionFailureAuthorityMissing
	}
	runtimeContext := WorkflowRAGApplicationRuntimeContext{RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef, OwnerSubjectRef: ctx.OwnerSubjectRef, AuditRef: ctx.AuditRef}
	assignment, _, _, err := resolver.ragRuntime.Read(runtimeContext)
	if errors.Is(err, errWorkflowRAGApplicationAssignmentNotFound) {
		return nil, ApplicationInteractionFailureAuthorityMissing
	}
	if err != nil {
		return nil, ApplicationInteractionFailureStoreUnavailable
	}
	if assignment.State != workflowRAGApplicationRuntimeStateActive {
		return nil, ApplicationInteractionFailureAuthorityMissing
	}
	authority, failure := resolver.ragAuthorityResolver.Resolve(runtimeContext, assignment.PublishCandidateID, &assignment)
	if failure != "" {
		switch failure {
		case WorkflowRAGApplicationFailureApplicationArchived:
			return nil, ApplicationInteractionFailureApplicationArchived
		case WorkflowRAGApplicationFailureStoreUnavailable:
			return nil, ApplicationInteractionFailureStoreUnavailable
		case WorkflowRAGApplicationFailureConfigurationChanged, WorkflowRAGApplicationFailureCandidateSuperseded, WorkflowRAGApplicationFailureVersionConflict:
			return nil, ApplicationInteractionFailureAuthorityChanged
		default:
			return nil, ApplicationInteractionFailureProfileIneligible
		}
	}
	return &ApplicationInteractionRAGAuthority{AssignmentID: assignment.AssignmentID, AssignmentVersion: assignment.RecordVersion, AssignmentDigest: assignment.AssignmentDigest, PublishCandidateID: authority.Candidate.CandidateID, PublishReviewVersion: authority.Candidate.ReviewVersion, DraftID: authority.Draft.DraftID, DraftVersion: authority.Draft.DraftVersion, DraftDigest: assignment.DraftDigest, BindingRef: authority.Binding.WorkflowRAGApplicationBindingRef, SnapshotID: authority.Snapshot.SnapshotID, SnapshotVersion: authority.Snapshot.SnapshotVersion, SnapshotDigest: authority.Snapshot.SnapshotDigest, RetrievalProfileID: authority.Profile.ProfileID, RetrievalProfileVersion: authority.Profile.ProfileVersion, RetrievalProfileDigest: authority.Profile.ProfileDigest}, ""
}

func validateApplicationInteractionProfileBinding(binding ApplicationInteractionProfileBinding) error {
	binding.ExecutionProfile = strings.TrimSpace(binding.ExecutionProfile)
	binding.DefinitionID = strings.TrimSpace(binding.DefinitionID)
	switch binding.ExecutionProfile {
	case applicationInteractionProfileWorkflow, applicationInteractionProfileWorkflowStructured:
		if !applicationDraftIdentifierPattern.MatchString(binding.DefinitionID) {
			return errWorkflowRunStoreContract
		}
	case applicationInteractionProfileRAG:
		if binding.DefinitionID != "" {
			return errWorkflowRunStoreContract
		}
	case applicationInteractionProfilePrompt, applicationInteractionProfileAgentCopilot:
		if binding.DefinitionID != "" {
			return errWorkflowRunStoreContract
		}
	default:
		return errWorkflowRunStoreContract
	}
	return nil
}

func applicationInteractionAuthorityDigest(snapshot ApplicationInteractionAuthoritySnapshot) (string, error) {
	if snapshot.ExecutionProfile == applicationInteractionProfilePrompt {
		authority, err := promptAuthorityFromApplicationInteraction(snapshot)
		if err != nil {
			return "", err
		}
		return promptApplicationRuntimeAuthorityV2Digest(authority)
	}
	if snapshot.ExecutionProfile == applicationInteractionProfileAgentCopilot {
		authority, err := agentCopilotAuthorityFromApplicationInteraction(snapshot)
		if err != nil {
			return "", err
		}
		return agentCopilotAuthorityDigest(authority)
	}
	snapshot.AuthorityDigest = ""
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func validateApplicationInteractionAuthority(snapshot ApplicationInteractionAuthoritySnapshot) error {
	if snapshot.ExecutionProfile == applicationInteractionProfilePrompt {
		authority, err := promptAuthorityFromApplicationInteraction(snapshot)
		if err != nil || validatePromptApplicationRuntimeAuthorityV2(authority) != nil {
			return errWorkflowRunStoreContract
		}
		return nil
	}
	if snapshot.ExecutionProfile == applicationInteractionProfileAgentCopilot {
		authority, err := agentCopilotAuthorityFromApplicationInteraction(snapshot)
		if err != nil || validateAgentCopilotRuntimeAuthority(authority) != nil {
			return errWorkflowRunStoreContract
		}
		return nil
	}
	if (snapshot.SchemaVersion != applicationInteractionAuthoritySchemaVersion && snapshot.SchemaVersion != applicationInteractionStructuredAuthoritySchemaVersion) || !applicationDraftIdentifierPattern.MatchString(snapshot.ApplicationID) || snapshot.ApplicationRecordVersion < 1 || snapshot.ApplicationLifecycle != applicationCatalogLifecycleActive || !workflowRAGDigestPattern.MatchString(snapshot.AuthorityDigest) {
		return errWorkflowRunStoreContract
	}
	calculated, err := applicationInteractionAuthorityDigest(snapshot)
	if err != nil || calculated != snapshot.AuthorityDigest {
		return errWorkflowRunStoreContract
	}
	switch snapshot.ExecutionProfile {
	case applicationInteractionProfileWorkflow, applicationInteractionProfileWorkflowStructured:
		value := snapshot.WorkflowDefinition
		if value == nil || snapshot.ApplicationRAG != nil || !applicationDraftIdentifierPattern.MatchString(value.DefinitionID) || value.DefinitionVersion < 1 || value.ActivationPointerVersion < 1 || !workflowRAGDigestPattern.MatchString(value.DefinitionDigest) || !applicationDraftIdentifierPattern.MatchString(value.CandidateID) || value.CandidateReviewVersion < 1 {
			return errWorkflowRunStoreContract
		}
		if snapshot.ExecutionProfile == applicationInteractionProfileWorkflow {
			if snapshot.SchemaVersion != applicationInteractionAuthoritySchemaVersion || value.InputContract != nil {
				return errWorkflowRunStoreContract
			}
		} else {
			if snapshot.SchemaVersion != applicationInteractionStructuredAuthoritySchemaVersion || value.InputContract == nil {
				return errWorkflowRunStoreContract
			}
			normalized, code, _ := normalizeWorkflowStructuredInputContract(savedWorkflowDraftContractFromDefinition(*value.InputContract))
			if code != "" || normalized.ContractDigest != value.InputContract.ContractDigest {
				return errWorkflowRunStoreContract
			}
		}
	case applicationInteractionProfileRAG:
		value := snapshot.ApplicationRAG
		expectedProfile := workflowRAGLexicalProfile()
		if value == nil || snapshot.WorkflowDefinition != nil || !workflowRAGApplicationAssignmentIDPattern.MatchString(value.AssignmentID) || value.AssignmentVersion < 1 || !workflowRAGDigestPattern.MatchString(value.AssignmentDigest) || !applicationDraftIdentifierPattern.MatchString(value.PublishCandidateID) || value.PublishReviewVersion < 1 || !applicationDraftIdentifierPattern.MatchString(value.DraftID) || value.DraftVersion < 1 || !workflowRAGDigestPattern.MatchString(value.DraftDigest) || !validWorkflowRAGApplicationBindingRef(value.BindingRef) || !workflowRAGSnapshotIDPattern.MatchString(value.SnapshotID) || value.SnapshotVersion < 1 || !workflowRAGDigestPattern.MatchString(value.SnapshotDigest) || value.RetrievalProfileID != expectedProfile.ProfileID || value.RetrievalProfileVersion != expectedProfile.ProfileVersion || value.RetrievalProfileDigest != expectedProfile.ProfileDigest {
			return errWorkflowRunStoreContract
		}
	default:
		return errWorkflowRunStoreContract
	}
	return nil
}

func applicationInteractionAuthorityFromPrompt(authority PromptApplicationRuntimeAuthorityV2) ApplicationInteractionAuthoritySnapshot {
	prompt := authority.PromptApplication
	return ApplicationInteractionAuthoritySnapshot{
		SchemaVersion: authority.SchemaVersion, ExecutionProfile: authority.ExecutionProfile,
		ApplicationID: authority.ApplicationID, ApplicationRecordVersion: authority.ApplicationRecordVersion,
		ApplicationLifecycle: authority.ApplicationLifecycle, PromptApplication: &prompt,
		AuthorityDigest: authority.AuthorityDigest,
	}
}

func promptAuthorityFromApplicationInteraction(snapshot ApplicationInteractionAuthoritySnapshot) (PromptApplicationRuntimeAuthorityV2, error) {
	if snapshot.SchemaVersion != promptApplicationRuntimeAuthorityV2Schema ||
		snapshot.ExecutionProfile != promptApplicationInvocationProfile || snapshot.PromptApplication == nil ||
		snapshot.WorkflowDefinition != nil || snapshot.ApplicationRAG != nil {
		return PromptApplicationRuntimeAuthorityV2{}, errWorkflowRunStoreContract
	}
	return PromptApplicationRuntimeAuthorityV2{
		SchemaVersion: snapshot.SchemaVersion, ExecutionProfile: snapshot.ExecutionProfile,
		ApplicationID: snapshot.ApplicationID, ApplicationRecordVersion: snapshot.ApplicationRecordVersion,
		ApplicationLifecycle: snapshot.ApplicationLifecycle, PromptApplication: *snapshot.PromptApplication,
		AuthorityDigest: snapshot.AuthorityDigest,
	}, nil
}

func applicationInteractionAuthorityFromAgentCopilot(authority AgentCopilotRuntimeAuthorityV3) ApplicationInteractionAuthoritySnapshot {
	agent := authority.AgentCopilot
	return ApplicationInteractionAuthoritySnapshot{
		SchemaVersion: authority.SchemaVersion, ExecutionProfile: authority.ExecutionProfile,
		ApplicationID: authority.ApplicationID, ApplicationRecordVersion: authority.ApplicationRecordVersion,
		ApplicationLifecycle: authority.ApplicationLifecycle, AgentCopilot: &agent,
		AuthorityDigest: authority.AuthorityDigest,
	}
}

func agentCopilotAuthorityFromApplicationInteraction(snapshot ApplicationInteractionAuthoritySnapshot) (AgentCopilotRuntimeAuthorityV3, error) {
	if snapshot.SchemaVersion != agentCopilotRuntimeAuthorityV3Schema ||
		snapshot.ExecutionProfile != agentCopilotSuggestionProfile || snapshot.AgentCopilot == nil ||
		snapshot.WorkflowDefinition != nil || snapshot.ApplicationRAG != nil || snapshot.PromptApplication != nil {
		return AgentCopilotRuntimeAuthorityV3{}, errWorkflowRunStoreContract
	}
	return AgentCopilotRuntimeAuthorityV3{
		SchemaVersion: snapshot.SchemaVersion, ExecutionProfile: snapshot.ExecutionProfile,
		ApplicationID: snapshot.ApplicationID, ApplicationRecordVersion: snapshot.ApplicationRecordVersion,
		ApplicationLifecycle: snapshot.ApplicationLifecycle, AgentCopilot: *snapshot.AgentCopilot,
		AuthorityDigest: snapshot.AuthorityDigest,
	}, nil
}

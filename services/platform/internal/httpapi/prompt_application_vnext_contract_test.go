package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
)

const promptApplicationContractTestDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestPromptApplicationVNextContractsDecodeStrict(t *testing.T) {
	contracts := promptApplicationVNextContractFixtures(t)
	for schemaVersion, contract := range contracts {
		t.Run(schemaVersion, func(t *testing.T) {
			payload, err := json.Marshal(contract)
			if err != nil {
				t.Fatalf("marshal fixture: %v", err)
			}
			decoded, err := decodePromptApplicationVNextContract(schemaVersion, payload)
			if err != nil || decoded == nil {
				t.Fatalf("decode strict contract: decoded=%#v err=%v payload=%s", decoded, err, payload)
			}
			withUnknown := append(payload[:len(payload)-1], []byte(`,"provider_api_key":"forbidden"}`)...)
			if _, err := decodePromptApplicationVNextContract(schemaVersion, withUnknown); err == nil {
				t.Fatal("unknown sensitive field must be rejected")
			}
			if _, err := decodePromptApplicationVNextContract(schemaVersion, append(payload, []byte("\n{}")...)); err == nil {
				t.Fatal("trailing JSON value must be rejected")
			}
		})
	}
}

func TestPromptApplicationVNextContractsRejectVersionAndProfileDrift(t *testing.T) {
	fixtures := promptApplicationVNextContractFixtures(t)
	tests := []struct {
		name          string
		schemaVersion string
		contract      any
		mutate        func(map[string]any)
	}{
		{name: "configuration v2", schemaVersion: promptApplicationConfigurationDraftV3Schema, contract: fixtures[promptApplicationConfigurationDraftV3Schema], mutate: func(value map[string]any) { value["schema_version"] = applicationConfigurationDraftSchemaVersionV2 }},
		{name: "publish v2", schemaVersion: promptApplicationPublishCandidateV3Schema, contract: fixtures[promptApplicationPublishCandidateV3Schema], mutate: func(value map[string]any) { value["schema_version"] = applicationPublishCandidateSchemaVersionV2 }},
		{name: "authority v1", schemaVersion: promptApplicationRuntimeAuthorityV2Schema, contract: fixtures[promptApplicationRuntimeAuthorityV2Schema], mutate: func(value map[string]any) { value["schema_version"] = applicationInteractionAuthoritySchemaVersion }},
		{name: "workflow profile", schemaVersion: promptApplicationRuntimeAuthorityV2Schema, contract: fixtures[promptApplicationRuntimeAuthorityV2Schema], mutate: func(value map[string]any) { value["execution_profile"] = applicationInteractionProfileWorkflow }},
		{name: "session v1", schemaVersion: promptApplicationSessionV2Schema, contract: fixtures[promptApplicationSessionV2Schema], mutate: func(value map[string]any) { value["schema_version"] = applicationSessionSchemaVersion }},
		{name: "session definition binding", schemaVersion: promptApplicationSessionV2Schema, contract: fixtures[promptApplicationSessionV2Schema], mutate: func(value map[string]any) {
			value["profile_binding"].(map[string]any)["definition_id"] = "definition_forbidden"
		}},
		{name: "turn v1", schemaVersion: promptApplicationSessionTurnV2Schema, contract: fixtures[promptApplicationSessionTurnV2Schema], mutate: func(value map[string]any) { value["schema_version"] = applicationSessionTurnSchemaVersion }},
		{name: "turn run v5", schemaVersion: promptApplicationSessionTurnV2Schema, contract: fixtures[promptApplicationSessionTurnV2Schema], mutate: func(value map[string]any) {
			value["run_ref"].(map[string]any)["schema_version"] = workflowRunRecordDefinitionSchemaVersion
		}},
		{name: "run v5", schemaVersion: promptApplicationRunV6Schema, contract: fixtures[promptApplicationRunV6Schema], mutate: func(value map[string]any) { value["schema_version"] = workflowRunRecordDefinitionSchemaVersion }},
		{name: "run output", schemaVersion: promptApplicationRunV6Schema, contract: fixtures[promptApplicationRunV6Schema], mutate: func(value map[string]any) { value["output"] = "must not persist" }},
		{name: "run raw variables", schemaVersion: promptApplicationRunV6Schema, contract: fixtures[promptApplicationRunV6Schema], mutate: func(value map[string]any) { value["variables"] = map[string]any{"question": "secret"} }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload, err := json.Marshal(test.contract)
			if err != nil {
				t.Fatalf("marshal fixture: %v", err)
			}
			var value map[string]any
			if err := json.Unmarshal(payload, &value); err != nil {
				t.Fatalf("decode mutable fixture: %v", err)
			}
			test.mutate(value)
			payload, err = json.Marshal(value)
			if err != nil {
				t.Fatalf("marshal mutated fixture: %v", err)
			}
			if _, err := decodePromptApplicationVNextContract(test.schemaVersion, payload); err == nil {
				t.Fatalf("drifted contract must be rejected: %s", payload)
			}
		})
	}
}

func TestPromptApplicationVNextContractsRemainMetadataOnly(t *testing.T) {
	for schemaVersion, contract := range promptApplicationVNextContractFixtures(t) {
		payload, err := json.Marshal(contract)
		if err != nil {
			t.Fatalf("marshal %s: %v", schemaVersion, err)
		}
		serialized := string(payload)
		for _, forbidden := range []string{`"template_source"`, `"source"`, `"variables"`, `"messages"`, `"raw_response"`, `"provider_api_key"`} {
			if strings.Contains(serialized, forbidden) {
				t.Fatalf("%s copied forbidden runtime or source field %s", schemaVersion, forbidden)
			}
		}
	}
}

func TestPromptApplicationBatchCActivatesConfigurationAndPublishOwnersOnly(t *testing.T) {
	for _, schemaVersion := range []string{applicationConfigurationDraftSchemaVersionV1, applicationConfigurationDraftSchemaVersionV2} {
		if !applicationConfigurationDraftSchemaSupported(schemaVersion) {
			t.Fatalf("existing configuration schema %s lost compatibility", schemaVersion)
		}
	}
	if !applicationConfigurationDraftSchemaSupported(promptApplicationConfigurationDraftV3Schema) {
		t.Fatal("Batch C must activate v3 in the configuration owner")
	}
	if !applicationPublishCandidateSchemaSupported(promptApplicationPublishCandidateV3Schema) {
		t.Fatal("Batch C must activate v3 in the existing publish candidate owner")
	}
	for _, schemaVersion := range []string{workflowRunRecordLegacySchemaVersion, workflowRunRecordSchemaVersion, workflowRunRecordToolSchemaVersion, workflowRunRecordRAGSchemaVersion, workflowRunRecordAppRAGSchemaVersion, workflowRunRecordDefinitionSchemaVersion} {
		if !validWorkflowRunRecordSchema(schemaVersion) {
			t.Fatalf("existing run schema %s lost compatibility", schemaVersion)
		}
	}
	if validWorkflowRunRecordSchema(promptApplicationRunV6Schema) {
		t.Fatal("Batch C must not activate v6 in the run owner before controlled invocation")
	}
}

func promptApplicationVNextContractFixtures(t *testing.T) map[string]any {
	t.Helper()
	templateRef := PromptApplicationTemplateRef{TemplateID: "ptpl_aaaaaaaaaaaaaaaa", TemplateVersion: 2, TemplateDigest: promptApplicationContractTestDigest}
	configuration := PromptApplicationConfigurationDraftV3{
		SchemaVersion: promptApplicationConfigurationDraftV3Schema, DraftID: "draft_prompt_1", WorkspaceID: "workspace_1", ApplicationID: "app_aaaaaaaaaaaaaaaa", BaseApplicationUpdatedAt: "2026-07-21T08:00:00Z",
		DisplayName: "Support summary", Description: "Summarize a support request.", ApplicationKind: "prompt_application", DefaultProtocol: "responses", DefaultModel: "model:gpt_test", AllowedProtocols: []string{"responses"}, PromptTemplateRef: templateRef,
		DraftVersion: 1, DraftDigest: promptApplicationContractTestDigest, ValidationSummary: ApplicationConfigurationDraftValidation{State: applicationDraftValidationValid, IsValid: true, Findings: []ApplicationConfigurationDraftValidationFinding{}},
		CreatedAt: "2026-07-21T08:00:00Z", UpdatedAt: "2026-07-21T08:01:00Z", CreatedByActorRef: "actor:owner", UpdatedByActorRef: "actor:owner", RequestID: "request:1", AuditRef: "audit:1",
	}
	publish := PromptApplicationPublishCandidateV3{
		SchemaVersion: promptApplicationPublishCandidateV3Schema, CandidateID: "candidate_prompt_1", WorkspaceID: configuration.WorkspaceID, ApplicationID: configuration.ApplicationID, DraftID: configuration.DraftID, DraftVersion: configuration.DraftVersion, DraftDigest: configuration.DraftDigest, BaseApplicationUpdatedAt: configuration.BaseApplicationUpdatedAt,
		Configuration:      PromptApplicationPublishConfigurationV3{DisplayName: configuration.DisplayName, Description: configuration.Description, ApplicationKind: configuration.ApplicationKind, DefaultProtocol: configuration.DefaultProtocol, DefaultModel: configuration.DefaultModel, AllowedProtocols: configuration.AllowedProtocols, PromptTemplateRef: templateRef},
		EvidenceRequestIDs: []string{"request:evidence"}, CandidateState: applicationPublishStatePending, ReviewVersion: 0, Reviews: []ApplicationPublishReviewRecord{}, PromotionEligibility: blockedApplicationPromotionEligibility([]ApplicationPromotionBlocker{{Code: "publish_review_required", Summary: "Review is required."}}),
		CreatedAt: configuration.CreatedAt, UpdatedAt: configuration.UpdatedAt, CreatedByActorRef: configuration.CreatedByActorRef, UpdatedByActorRef: configuration.UpdatedByActorRef, RequestID: configuration.RequestID, AuditRef: configuration.AuditRef,
	}
	assignment := PromptApplicationRuntimeAssignment{
		SchemaVersion: promptApplicationRuntimeAssignmentSchema, AssignmentID: "ptra_aaaaaaaaaaaaaaaa", TenantRef: "tenant:1", WorkspaceID: configuration.WorkspaceID, ApplicationID: configuration.ApplicationID, OwnerSubjectRef: "subject:owner", AssignmentVersion: 1, State: "active",
		CandidateID: publish.CandidateID, CandidateReviewVersion: 1, DraftID: configuration.DraftID, DraftVersion: configuration.DraftVersion, DraftDigest: configuration.DraftDigest, PromptTemplateRef: templateRef, AssignmentDigest: promptApplicationContractTestDigest,
		ActivatedAt: "2026-07-21T08:02:00Z", UpdatedAt: "2026-07-21T08:02:00Z", ActivatedByActorRef: "actor:owner", UpdatedByActorRef: "actor:owner", RequestID: "request:2", AuditRef: "audit:2",
	}
	event := PromptApplicationRuntimeAssignmentEvent{
		SchemaVersion: promptApplicationRuntimeAssignmentEventSchema, EventID: "ptrae_aaaaaaaaaaaaaaaa", AssignmentID: assignment.AssignmentID, TenantRef: assignment.TenantRef, WorkspaceID: assignment.WorkspaceID, ApplicationID: assignment.ApplicationID, OwnerSubjectRef: assignment.OwnerSubjectRef,
		EventSequence: 1, Action: "activate", ExpectedAssignmentVersion: 0, ResultingAssignmentVersion: 1, CandidateID: assignment.CandidateID, CandidateReviewVersion: assignment.CandidateReviewVersion, DraftID: assignment.DraftID, DraftVersion: assignment.DraftVersion, DraftDigest: assignment.DraftDigest, PromptTemplateRef: templateRef, AssignmentDigest: assignment.AssignmentDigest,
		OccurredAt: assignment.ActivatedAt, ActorRef: assignment.ActivatedByActorRef, RequestID: assignment.RequestID, AuditRef: assignment.AuditRef,
	}
	authority := PromptApplicationRuntimeAuthorityV2{
		SchemaVersion: promptApplicationRuntimeAuthorityV2Schema, ExecutionProfile: promptApplicationInvocationProfile, ApplicationID: assignment.ApplicationID, ApplicationRecordVersion: 3, ApplicationLifecycle: applicationCatalogLifecycleActive,
		PromptApplication: PromptApplicationAuthorityV2{AssignmentID: assignment.AssignmentID, AssignmentVersion: assignment.AssignmentVersion, AssignmentDigest: assignment.AssignmentDigest, PublishCandidateID: assignment.CandidateID, PublishReviewVersion: assignment.CandidateReviewVersion, DraftID: assignment.DraftID, DraftVersion: assignment.DraftVersion, DraftDigest: assignment.DraftDigest, PromptTemplateRef: templateRef, DefaultProtocol: "responses", DefaultModel: "model:gpt_test", ProtocolPolicyDigest: promptApplicationContractTestDigest, ModelEligibilityDigest: promptApplicationContractTestDigest},
	}
	var err error
	authority.AuthorityDigest, err = promptApplicationRuntimeAuthorityV2Digest(authority)
	if err != nil {
		t.Fatalf("calculate authority digest: %v", err)
	}
	lastTurnID := "appturn_aaaaaaaaaaaaaaaa"
	closedAt := "2026-07-21T08:05:00Z"
	session := PromptApplicationSessionV2{
		SchemaVersion: promptApplicationSessionV2Schema, SessionID: "appsess_aaaaaaaaaaaaaaaa", TenantRef: assignment.TenantRef, WorkspaceID: assignment.WorkspaceID, ApplicationID: assignment.ApplicationID, OwnerSubjectRef: assignment.OwnerSubjectRef,
		State: applicationSessionStateClosed, RecordVersion: 2, ProfileBinding: PromptApplicationSessionProfileBindingV2{ExecutionProfile: promptApplicationInvocationProfile}, Authority: authority, ContentRetention: applicationSessionRetentionPolicy, TurnCount: 1, LastTurnID: &lastTurnID,
		CreatedAt: "2026-07-21T08:03:00Z", UpdatedAt: "2026-07-21T08:04:00Z", ClosedAt: &closedAt, CreatedByActorRef: "actor:owner", UpdatedByActorRef: "actor:owner", RequestID: "request:3", AuditRef: "audit:3",
	}
	completedAt := "2026-07-21T08:04:00Z"
	turn := PromptApplicationSessionTurnV2{
		SchemaVersion: promptApplicationSessionTurnV2Schema, TurnID: lastTurnID, SessionID: session.SessionID, Sequence: 1, ClientTurnKey: "client_turn_1", TenantRef: session.TenantRef, WorkspaceID: session.WorkspaceID, ApplicationID: session.ApplicationID, OwnerSubjectRef: session.OwnerSubjectRef,
		ExecutionProfile: promptApplicationInvocationProfile, Authority: authority, Status: string(WorkflowRunStatusSucceeded), InputDigest: promptApplicationContractTestDigest, InputBytes: 32, RunRef: &PromptApplicationRunRefV6{RunID: "run_aaaaaaaaaaaaaaaa", SchemaVersion: promptApplicationRunV6Schema}, StartedAt: "2026-07-21T08:03:30Z", CompletedAt: &completedAt,
		ActorRef: "actor:owner", RequestID: "request:4", AuditRef: "audit:4",
	}
	variableNames := []string{"question", "tone"}
	run := PromptApplicationRunRecordV6{
		SchemaVersion: promptApplicationRunV6Schema, RecordVersion: 1, RunID: turn.RunRef.RunID, TenantRef: turn.TenantRef, WorkspaceID: turn.WorkspaceID, ApplicationID: turn.ApplicationID, ExecutionKind: "prompt_application_invocation", ExecutionSourceKind: "prompt_application_template", ExecutionSourceID: templateRef.TemplateID, ExecutionSourceVersion: templateRef.TemplateVersion, ExecutionProfile: promptApplicationInvocationProfile, Authority: authority,
		InputDigest: turn.InputDigest, InputBytes: turn.InputBytes, VariableNames: variableNames, VariableNamesDigest: promptApplicationVariableNamesDigest(variableNames), RequestedProtocol: "responses", SelectedProtocol: "responses", RequestedModel: "model:gpt_test", SelectedProvider: "provider:test", SelectedProfile: "profile:test", SelectedModel: "model:gpt_test", UpstreamModel: "upstream:test", SelectionSource: "gateway:policy",
		Status: string(WorkflowRunStatusSucceeded), StartedAt: turn.StartedAt, CompletedAt: completedAt, Output: "", Usage: PromptApplicationRunUsageV6{State: "provider_reported", InputTokens: 8, OutputTokens: 4, TotalTokens: 12}, SideEffects: WorkflowRunSideEffects{ProviderCalls: 1}, Diagnostic: PromptApplicationRunDiagnosticV6{TerminalWriteState: "stored", GatewayFailureCategory: "none", ObservedAt: completedAt},
		RequestID: turn.RequestID, AuditRef: turn.AuditRef, ActorRef: turn.ActorRef,
	}
	return map[string]any{
		promptApplicationConfigurationDraftV3Schema:   configuration,
		promptApplicationPublishCandidateV3Schema:     publish,
		promptApplicationRuntimeAssignmentSchema:      assignment,
		promptApplicationRuntimeAssignmentEventSchema: event,
		promptApplicationRuntimeAuthorityV2Schema:     authority,
		promptApplicationSessionV2Schema:              session,
		promptApplicationSessionTurnV2Schema:          turn,
		promptApplicationRunV6Schema:                  run,
	}
}

func promptApplicationVariableNamesDigest(names []string) string {
	payload, _ := json.Marshal(names)
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:])
}

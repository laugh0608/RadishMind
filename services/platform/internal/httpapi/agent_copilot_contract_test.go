package httpapi

import (
	"encoding/json"
	"strings"
	"testing"
)

const agentCopilotContractTestDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestAgentCopilotContractsDecodeStrict(t *testing.T) {
	for schemaVersion, contract := range agentCopilotContractFixtures(t) {
		t.Run(schemaVersion, func(t *testing.T) {
			payload, err := json.Marshal(contract)
			if err != nil {
				t.Fatalf("marshal fixture: %v", err)
			}
			decoded, err := decodeAgentCopilotContract(schemaVersion, payload)
			if err != nil || decoded == nil {
				t.Fatalf("decode strict contract: decoded=%#v err=%v payload=%s", decoded, err, payload)
			}
			withUnknown := append(payload[:len(payload)-1], []byte(`,"provider_api_key":"forbidden"}`)...)
			if _, err := decodeAgentCopilotContract(schemaVersion, withUnknown); err == nil {
				t.Fatal("unknown sensitive field must be rejected")
			}
			if _, err := decodeAgentCopilotContract(schemaVersion, append(payload, []byte("\n{}")...)); err == nil {
				t.Fatal("trailing JSON value must be rejected")
			}
		})
	}
}

func TestAgentCopilotContractsRejectPolicyAndVersionDrift(t *testing.T) {
	fixtures := agentCopilotContractFixtures(t)
	tests := []struct {
		name          string
		schemaVersion string
		mutate        func(map[string]any)
	}{
		{name: "cross project task", schemaVersion: agentCopilotProfileDraftSchema, mutate: func(value map[string]any) {
			value["allowed_tasks"] = []any{"answer_docs_question"}
		}},
		{name: "default locale outside allowlist", schemaVersion: agentCopilotProfileDraftSchema, mutate: func(value map[string]any) {
			value["default_locale"] = "ja-JP"
		}},
		{name: "tool execution enabled", schemaVersion: agentCopilotProfileDraftSchema, mutate: func(value map[string]any) {
			value["tool_hints_policy"].(map[string]any)["allow_tool_calls"] = true
		}},
		{name: "confirmation disabled", schemaVersion: agentCopilotProfileVersionSchema, mutate: func(value map[string]any) {
			value["risk_policy"].(map[string]any)["requires_confirmation_for_actions"] = false
		}},
		{name: "configuration v3", schemaVersion: agentCopilotConfigurationDraftV4Schema, mutate: func(value map[string]any) {
			value["schema_version"] = promptApplicationConfigurationDraftV3Schema
		}},
		{name: "publish candidate v3", schemaVersion: agentCopilotPublishCandidateV4Schema, mutate: func(value map[string]any) {
			value["schema_version"] = promptApplicationPublishCandidateV3Schema
		}},
		{name: "authority v2", schemaVersion: agentCopilotRuntimeAuthorityV3Schema, mutate: func(value map[string]any) {
			value["schema_version"] = promptApplicationRuntimeAuthorityV2Schema
		}},
		{name: "session v2", schemaVersion: agentCopilotSessionV3Schema, mutate: func(value map[string]any) {
			value["schema_version"] = promptApplicationSessionV2Schema
		}},
		{name: "turn run v6", schemaVersion: agentCopilotSessionTurnV3Schema, mutate: func(value map[string]any) {
			value["run_ref"].(map[string]any)["schema_version"] = promptApplicationRunV6Schema
		}},
		{name: "run v6", schemaVersion: agentCopilotRunV7Schema, mutate: func(value map[string]any) {
			value["schema_version"] = promptApplicationRunV6Schema
		}},
		{name: "run output persisted", schemaVersion: agentCopilotRunV7Schema, mutate: func(value map[string]any) {
			value["output"] = "forbidden response body"
		}},
		{name: "run action without confirmation", schemaVersion: agentCopilotRunV7Schema, mutate: func(value map[string]any) {
			value["requires_confirmation"] = false
		}},
		{name: "run tool call", schemaVersion: agentCopilotRunV7Schema, mutate: func(value map[string]any) {
			value["side_effects"].(map[string]any)["tool_calls"] = float64(1)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload, err := json.Marshal(fixtures[test.schemaVersion])
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
			if _, err := decodeAgentCopilotContract(test.schemaVersion, payload); err == nil {
				t.Fatalf("drifted contract must be rejected: %s", payload)
			}
		})
	}
}

func TestAgentCopilotContractsRemainMetadataOnly(t *testing.T) {
	for schemaVersion, contract := range agentCopilotContractFixtures(t) {
		payload, err := json.Marshal(contract)
		if err != nil {
			t.Fatalf("marshal %s: %v", schemaVersion, err)
		}
		serialized := string(payload)
		for _, forbidden := range []string{`"system_prompt"`, `"provider_api_key"`, `"credential"`, `"messages"`, `"artifacts"`, `"artifact_content"`, `"raw_response"`, `"response_body"`} {
			if strings.Contains(serialized, forbidden) {
				t.Fatalf("%s copied forbidden source, context, credential, or response field %s", schemaVersion, forbidden)
			}
		}
	}
}

func TestAgentCopilotBatchA1DoesNotRegisterRuntimeVersions(t *testing.T) {
	for _, schemaVersion := range []string{applicationConfigurationDraftSchemaVersionV1, applicationConfigurationDraftSchemaVersionV2, promptApplicationConfigurationDraftV3Schema} {
		if !applicationConfigurationDraftSchemaSupported(schemaVersion) {
			t.Fatalf("existing configuration schema %s lost compatibility", schemaVersion)
		}
	}
	if applicationConfigurationDraftSchemaSupported(agentCopilotConfigurationDraftV4Schema) {
		t.Fatal("Batch A1 must not register application configuration v4")
	}
	for _, schemaVersion := range []string{applicationPublishCandidateSchemaVersionV1, applicationPublishCandidateSchemaVersionV2, promptApplicationPublishCandidateV3Schema} {
		if !applicationPublishCandidateSchemaSupported(schemaVersion) {
			t.Fatalf("existing publish candidate schema %s lost compatibility", schemaVersion)
		}
	}
	if applicationPublishCandidateSchemaSupported(agentCopilotPublishCandidateV4Schema) {
		t.Fatal("Batch A1 must not register application publish candidate v4")
	}
	for _, schemaVersion := range []string{
		workflowRunRecordLegacySchemaVersion,
		workflowRunRecordSchemaVersion,
		workflowRunRecordToolSchemaVersion,
		workflowRunRecordRAGSchemaVersion,
		workflowRunRecordAppRAGSchemaVersion,
		workflowRunRecordDefinitionSchemaVersion,
		promptApplicationRunV6Schema,
	} {
		if !validWorkflowRunRecordSchema(schemaVersion) {
			t.Fatalf("existing run schema %s lost compatibility", schemaVersion)
		}
	}
	if validWorkflowRunRecordSchema(agentCopilotRunV7Schema) {
		t.Fatal("Batch A1 must not register workflow run v7")
	}
	if _, exists := apiKeyAllowedScopes["agent_copilot:invoke"]; exists {
		t.Fatal("Batch A1 must not expose agent_copilot:invoke")
	}
}

func agentCopilotContractFixtures(t *testing.T) map[string]any {
	t.Helper()
	source := AgentCopilotProfileSource{
		ProfileName:    "RadishFlow diagnostics advisor",
		Description:    "Explain diagnostics and produce reviewable candidate actions.",
		Project:        "radishflow",
		AllowedTasks:   []string{"explain_diagnostics", "suggest_flowsheet_edits"},
		DefaultLocale:  "zh-CN",
		AllowedLocales: []string{"zh-CN", "en-US"},
		ContextPolicy:  AgentCopilotContextPolicy{AllowedFields: []string{"diagnostics", "selected_unit_ids"}, MaxBytes: 65536, RequireTaskContext: true},
		ArtifactPolicy: AgentCopilotArtifactPolicy{
			AllowedKinds: []string{"json", "text"},
			AllowedRoles: []string{"primary", "supporting"},
			MaxCount:     8, MaxItemBytes: 65536, MaxTotalBytes: 131072,
		},
		ResponsePolicy: AgentCopilotResponsePolicy{
			AllowedActionKinds: []string{"candidate_edit", "read_only_check"},
			MaxAnswers:         8, MaxIssues: 16, MaxActions: 8, MaxCitations: 16, MaxVisibleTextBytes: 8192,
		},
		RiskPolicy: AgentCopilotRiskPolicy{
			Mode: "advisory", RequiresConfirmationForActions: true,
			ConfirmationActionKinds: []string{"candidate_edit", "candidate_operation", "ghost_completion"},
		},
		ToolHintsPolicy: AgentCopilotToolHintsPolicy{},
	}
	profileRef := AgentCopilotProfileRef{
		ProfileID: "acpf_aaaaaaaaaaaaaaaa", ProfileVersion: 2,
		ProfileDigest: agentCopilotContractTestDigest, PolicyDigest: agentCopilotContractTestDigest,
	}
	draft := AgentCopilotProfileDraftV1{
		SchemaVersion: agentCopilotProfileDraftSchema, ProfileID: profileRef.ProfileID, TenantRef: "tenant:1", WorkspaceID: "workspace_1",
		ApplicationID: "app_aaaaaaaaaaaaaaaa", OwnerSubjectRef: "subject:owner", AgentCopilotProfileSource: source,
		DraftVersion: 2, ProfileDigest: profileRef.ProfileDigest, PolicyDigest: profileRef.PolicyDigest,
		ValidationSummary: ApplicationConfigurationDraftValidation{State: applicationDraftValidationValid, IsValid: true, Findings: []ApplicationConfigurationDraftValidationFinding{}},
		CreatedAt:         "2026-07-25T08:00:00Z", UpdatedAt: "2026-07-25T08:01:00Z",
		CreatedByActorRef: "actor:owner", UpdatedByActorRef: "actor:owner", RequestID: "request:1", AuditRef: "audit:1",
	}
	version := AgentCopilotProfileVersionV1{
		SchemaVersion: agentCopilotProfileVersionSchema, ProfileID: profileRef.ProfileID, ProfileVersion: profileRef.ProfileVersion, SourceDraftVersion: draft.DraftVersion,
		TenantRef: draft.TenantRef, WorkspaceID: draft.WorkspaceID, ApplicationID: draft.ApplicationID, OwnerSubjectRef: draft.OwnerSubjectRef,
		AgentCopilotProfileSource: source, ProfileDigest: profileRef.ProfileDigest, PolicyDigest: profileRef.PolicyDigest,
		PublishedAt: "2026-07-25T08:02:00Z", PublishedByActorRef: "actor:owner", RequestID: "request:2", AuditRef: "audit:2",
	}
	configuration := AgentCopilotConfigurationDraftV4{
		SchemaVersion: agentCopilotConfigurationDraftV4Schema, DraftID: "draft_agent_1", WorkspaceID: draft.WorkspaceID, ApplicationID: draft.ApplicationID,
		BaseApplicationUpdatedAt: "2026-07-25T08:00:00Z", DisplayName: "Diagnostics advisor", Description: "Controlled RadishFlow suggestions.",
		ApplicationKind: "agent", DefaultProtocol: "responses", DefaultModel: "model:gpt_test", AllowedProtocols: []string{"responses"},
		AgentCopilotProfileRef: profileRef, DraftVersion: 1, DraftDigest: agentCopilotContractTestDigest,
		ValidationSummary: ApplicationConfigurationDraftValidation{State: applicationDraftValidationValid, IsValid: true, Findings: []ApplicationConfigurationDraftValidationFinding{}},
		CreatedAt:         "2026-07-25T08:03:00Z", UpdatedAt: "2026-07-25T08:04:00Z",
		CreatedByActorRef: "actor:owner", UpdatedByActorRef: "actor:owner", RequestID: "request:3", AuditRef: "audit:3",
	}
	publish := AgentCopilotPublishCandidateV4{
		SchemaVersion: agentCopilotPublishCandidateV4Schema, CandidateID: "candidate_agent_1", WorkspaceID: configuration.WorkspaceID,
		ApplicationID: configuration.ApplicationID, DraftID: configuration.DraftID, DraftVersion: configuration.DraftVersion, DraftDigest: configuration.DraftDigest,
		BaseApplicationUpdatedAt: configuration.BaseApplicationUpdatedAt,
		Configuration: AgentCopilotPublishConfigurationV4{
			DisplayName: configuration.DisplayName, Description: configuration.Description, ApplicationKind: configuration.ApplicationKind,
			DefaultProtocol: configuration.DefaultProtocol, DefaultModel: configuration.DefaultModel, AllowedProtocols: configuration.AllowedProtocols,
			AgentCopilotProfileRef: profileRef,
		},
		EvidenceRequestIDs: []string{"request:evidence"}, CandidateState: applicationPublishStatePending, ReviewVersion: 0,
		Reviews:              []ApplicationPublishReviewRecord{},
		PromotionEligibility: blockedApplicationPromotionEligibility([]ApplicationPromotionBlocker{{Code: "publish_review_required", Summary: "Review is required."}}),
		CreatedAt:            configuration.CreatedAt, UpdatedAt: configuration.UpdatedAt, CreatedByActorRef: configuration.CreatedByActorRef,
		UpdatedByActorRef: configuration.UpdatedByActorRef, RequestID: configuration.RequestID, AuditRef: configuration.AuditRef,
	}
	assignment := AgentCopilotRuntimeAssignmentV1{
		SchemaVersion: agentCopilotRuntimeAssignmentSchema, AssignmentID: "acra_aaaaaaaaaaaaaaaa", TenantRef: draft.TenantRef,
		WorkspaceID: draft.WorkspaceID, ApplicationID: draft.ApplicationID, OwnerSubjectRef: draft.OwnerSubjectRef,
		AssignmentVersion: 1, State: "active", CandidateID: publish.CandidateID, CandidateReviewVersion: 1,
		DraftID: configuration.DraftID, DraftVersion: configuration.DraftVersion, DraftDigest: configuration.DraftDigest,
		AgentCopilotProfileRef: profileRef, AssignmentDigest: agentCopilotContractTestDigest,
		ActivatedAt: "2026-07-25T08:05:00Z", UpdatedAt: "2026-07-25T08:05:00Z",
		ActivatedByActorRef: "actor:owner", UpdatedByActorRef: "actor:owner", RequestID: "request:4", AuditRef: "audit:4",
	}
	event := AgentCopilotRuntimeAssignmentEventV1{
		SchemaVersion: agentCopilotRuntimeAssignmentEventSchema, EventID: "acrae_aaaaaaaaaaaaaaaa", AssignmentID: assignment.AssignmentID,
		TenantRef: assignment.TenantRef, WorkspaceID: assignment.WorkspaceID, ApplicationID: assignment.ApplicationID, OwnerSubjectRef: assignment.OwnerSubjectRef,
		EventSequence: 1, Action: "activate", ExpectedAssignmentVersion: 0, ResultingAssignmentVersion: 1,
		CandidateID: assignment.CandidateID, CandidateReviewVersion: assignment.CandidateReviewVersion,
		DraftID: assignment.DraftID, DraftVersion: assignment.DraftVersion, DraftDigest: assignment.DraftDigest,
		AgentCopilotProfileRef: profileRef, AssignmentDigest: assignment.AssignmentDigest,
		OccurredAt: assignment.ActivatedAt, ActorRef: assignment.ActivatedByActorRef, RequestID: assignment.RequestID, AuditRef: assignment.AuditRef,
	}
	authority := AgentCopilotRuntimeAuthorityV3{
		SchemaVersion: agentCopilotRuntimeAuthorityV3Schema, ExecutionProfile: agentCopilotSuggestionProfile,
		ApplicationID: assignment.ApplicationID, ApplicationRecordVersion: 3, ApplicationLifecycle: applicationCatalogLifecycleActive,
		AgentCopilot: AgentCopilotAuthorityV3{
			AssignmentID: assignment.AssignmentID, AssignmentVersion: assignment.AssignmentVersion, AssignmentDigest: assignment.AssignmentDigest,
			PublishCandidateID: assignment.CandidateID, PublishReviewVersion: assignment.CandidateReviewVersion,
			DraftID: assignment.DraftID, DraftVersion: assignment.DraftVersion, DraftDigest: assignment.DraftDigest,
			AgentCopilotProfileRef: profileRef, Project: source.Project, AllowedTasksDigest: agentCopilotContractTestDigest,
			DefaultProtocol: "responses", DefaultModel: "model:gpt_test",
			ProtocolPolicyDigest: agentCopilotContractTestDigest, ModelEligibilityDigest: agentCopilotContractTestDigest,
		},
	}
	var err error
	authority.AuthorityDigest, err = agentCopilotAuthorityDigest(authority)
	if err != nil {
		t.Fatalf("calculate authority digest: %v", err)
	}
	lastTurnID := "appturn_aaaaaaaaaaaaaaaa"
	closedAt := "2026-07-25T08:09:00Z"
	session := AgentCopilotSessionV3{
		SchemaVersion: agentCopilotSessionV3Schema, SessionID: "appsess_aaaaaaaaaaaaaaaa", TenantRef: assignment.TenantRef,
		WorkspaceID: assignment.WorkspaceID, ApplicationID: assignment.ApplicationID, OwnerSubjectRef: assignment.OwnerSubjectRef,
		State: applicationSessionStateClosed, RecordVersion: 2,
		ProfileBinding: PromptApplicationSessionProfileBindingV2{ExecutionProfile: agentCopilotSuggestionProfile},
		Authority:      authority, ContentRetention: applicationSessionRetentionPolicy, TurnCount: 1, LastTurnID: &lastTurnID,
		CreatedAt: "2026-07-25T08:06:00Z", UpdatedAt: "2026-07-25T08:08:00Z", ClosedAt: &closedAt,
		CreatedByActorRef: "actor:owner", UpdatedByActorRef: "actor:owner", RequestID: "request:5", AuditRef: "audit:5",
	}
	completedAt := "2026-07-25T08:08:00Z"
	turn := AgentCopilotSessionTurnV3{
		SchemaVersion: agentCopilotSessionTurnV3Schema, TurnID: lastTurnID, SessionID: session.SessionID, Sequence: 1,
		ClientTurnKey: "client_turn_1", TenantRef: session.TenantRef, WorkspaceID: session.WorkspaceID,
		ApplicationID: session.ApplicationID, OwnerSubjectRef: session.OwnerSubjectRef, ExecutionProfile: agentCopilotSuggestionProfile,
		Authority: authority, Status: string(WorkflowRunStatusSucceeded), InputDigest: agentCopilotContractTestDigest, InputBytes: 2048,
		RunRef:    &AgentCopilotRunRefV7{RunID: "run_aaaaaaaaaaaaaaaa", SchemaVersion: agentCopilotRunV7Schema},
		StartedAt: "2026-07-25T08:07:00Z", CompletedAt: &completedAt,
		ActorRef: "actor:owner", RequestID: "request:6", AuditRef: "audit:6",
	}
	run := AgentCopilotRunRecordV7{
		SchemaVersion: agentCopilotRunV7Schema, RecordVersion: 1, RunID: turn.RunRef.RunID, TenantRef: turn.TenantRef,
		WorkspaceID: turn.WorkspaceID, ApplicationID: turn.ApplicationID, ExecutionKind: "agent_copilot_suggestion",
		ExecutionSourceKind: "agent_copilot_profile", ExecutionSourceID: profileRef.ProfileID,
		ExecutionSourceVersion: profileRef.ProfileVersion, ExecutionProfile: agentCopilotSuggestionProfile, Authority: authority,
		Project: source.Project, Task: "suggest_flowsheet_edits", Locale: source.DefaultLocale,
		InputDigest: turn.InputDigest, InputBytes: turn.InputBytes, ContextBytes: 1024, ArtifactCount: 1, ArtifactBytes: 512,
		RequestedProtocol: "responses", SelectedProtocol: "responses", RequestedModel: "model:gpt_test",
		SelectedProvider: "provider:test", SelectedProfile: "profile:test", SelectedModel: "model:gpt_test",
		UpstreamModel: "upstream:test", SelectionSource: "gateway:policy",
		ResponseStatus: "ok", ResponseDigest: agentCopilotContractTestDigest, AnswerCount: 1, IssueCount: 1,
		ActionCount: 1, CitationCount: 1, RiskLevel: "medium", RequiresConfirmation: true,
		Status: string(WorkflowRunStatusSucceeded), StartedAt: turn.StartedAt, CompletedAt: completedAt, Output: "",
		Usage:       PromptApplicationRunUsageV6{State: "provider_reported", InputTokens: 128, OutputTokens: 64, TotalTokens: 192},
		SideEffects: AgentCopilotRunSideEffectsV7{ProviderCalls: 1},
		Diagnostic:  PromptApplicationRunDiagnosticV6{TerminalWriteState: "stored", GatewayFailureCategory: "none", ObservedAt: completedAt},
		RequestID:   turn.RequestID, AuditRef: turn.AuditRef, ActorRef: turn.ActorRef,
	}
	return map[string]any{
		agentCopilotProfileDraftSchema:           draft,
		agentCopilotProfileVersionSchema:         version,
		agentCopilotConfigurationDraftV4Schema:   configuration,
		agentCopilotPublishCandidateV4Schema:     publish,
		agentCopilotRuntimeAssignmentSchema:      assignment,
		agentCopilotRuntimeAssignmentEventSchema: event,
		agentCopilotRuntimeAuthorityV3Schema:     authority,
		agentCopilotSessionV3Schema:              session,
		agentCopilotSessionTurnV3Schema:          turn,
		agentCopilotRunV7Schema:                  run,
	}
}

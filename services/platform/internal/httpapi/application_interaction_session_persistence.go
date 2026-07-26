package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
)

func applicationInteractionRequestContext(ctx ApplicationInteractionContext) context.Context {
	if ctx.RequestContext != nil {
		return ctx.RequestContext
	}
	return context.Background()
}

func encodeApplicationInteractionSession(session ApplicationInteractionSession) ([]byte, error) {
	if err := validateStoredApplicationInteractionSession(ApplicationInteractionContext{
		TenantRef: session.TenantRef, WorkspaceID: session.WorkspaceID, ApplicationID: session.ApplicationID, OwnerSubjectRef: session.OwnerSubjectRef,
	}, session); err != nil {
		return nil, errApplicationSessionContract
	}
	var payload []byte
	var err error
	if session.SchemaVersion == promptApplicationSessionV2Schema {
		contract, convertErr := promptApplicationSessionContractFromInteraction(session)
		if convertErr != nil {
			return nil, errApplicationSessionContract
		}
		payload, err = json.Marshal(contract)
	} else if session.SchemaVersion == agentCopilotSessionV3Schema {
		contract, convertErr := agentCopilotSessionContractFromInteraction(session)
		if convertErr != nil {
			return nil, errApplicationSessionContract
		}
		payload, err = json.Marshal(contract)
	} else {
		payload, err = json.Marshal(session)
	}
	if err != nil {
		return nil, errApplicationSessionContract
	}
	return payload, nil
}

func decodeApplicationInteractionSession(ctx ApplicationInteractionContext, payload []byte) (ApplicationInteractionSession, error) {
	var envelope struct {
		SchemaVersion string `json:"schema_version"`
	}
	if json.Unmarshal(payload, &envelope) != nil {
		return ApplicationInteractionSession{}, errApplicationSessionContract
	}
	if envelope.SchemaVersion == promptApplicationSessionV2Schema {
		value, err := decodePromptApplicationVNextContract(promptApplicationSessionV2Schema, payload)
		if err != nil {
			return ApplicationInteractionSession{}, errApplicationSessionContract
		}
		contract, ok := value.(*PromptApplicationSessionV2)
		if !ok {
			return ApplicationInteractionSession{}, errApplicationSessionContract
		}
		session := applicationInteractionSessionFromPromptContract(*contract)
		if validateStoredApplicationInteractionSession(ctx, session) != nil {
			return ApplicationInteractionSession{}, errApplicationSessionContract
		}
		return session, nil
	}
	if envelope.SchemaVersion == agentCopilotSessionV3Schema {
		value, err := decodeAgentCopilotContract(agentCopilotSessionV3Schema, payload)
		if err != nil {
			return ApplicationInteractionSession{}, errApplicationSessionContract
		}
		contract, ok := value.(*AgentCopilotSessionV3)
		if !ok {
			return ApplicationInteractionSession{}, errApplicationSessionContract
		}
		session := applicationInteractionSessionFromAgentCopilotContract(*contract)
		if validateStoredApplicationInteractionSession(ctx, session) != nil {
			return ApplicationInteractionSession{}, errApplicationSessionContract
		}
		return session, nil
	}
	var session ApplicationInteractionSession
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&session); err != nil {
		return ApplicationInteractionSession{}, errApplicationSessionContract
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return ApplicationInteractionSession{}, errApplicationSessionContract
	}
	if err := validateStoredApplicationInteractionSession(ctx, session); err != nil {
		return ApplicationInteractionSession{}, errApplicationSessionContract
	}
	return session, nil
}

func encodeApplicationInteractionTurn(turn ApplicationInteractionTurn) ([]byte, error) {
	if err := validateStoredApplicationInteractionTurn(ApplicationInteractionContext{
		TenantRef: turn.TenantRef, WorkspaceID: turn.WorkspaceID, ApplicationID: turn.ApplicationID, OwnerSubjectRef: turn.OwnerSubjectRef,
	}, turn); err != nil {
		return nil, errApplicationSessionContract
	}
	var payload []byte
	var err error
	if turn.SchemaVersion == promptApplicationSessionTurnV2Schema {
		contract, convertErr := promptApplicationTurnContractFromInteraction(turn)
		if convertErr != nil {
			return nil, errApplicationSessionContract
		}
		payload, err = json.Marshal(contract)
	} else if turn.SchemaVersion == agentCopilotSessionTurnV3Schema {
		contract, convertErr := agentCopilotTurnContractFromInteraction(turn)
		if convertErr != nil {
			return nil, errApplicationSessionContract
		}
		payload, err = json.Marshal(contract)
	} else {
		payload, err = json.Marshal(turn)
	}
	if err != nil {
		return nil, errApplicationSessionContract
	}
	return payload, nil
}

func decodeApplicationInteractionTurn(ctx ApplicationInteractionContext, payload []byte) (ApplicationInteractionTurn, error) {
	var envelope struct {
		SchemaVersion string `json:"schema_version"`
	}
	if json.Unmarshal(payload, &envelope) != nil {
		return ApplicationInteractionTurn{}, errApplicationSessionContract
	}
	if envelope.SchemaVersion == promptApplicationSessionTurnV2Schema {
		value, err := decodePromptApplicationVNextContract(promptApplicationSessionTurnV2Schema, payload)
		if err != nil {
			return ApplicationInteractionTurn{}, errApplicationSessionContract
		}
		contract, ok := value.(*PromptApplicationSessionTurnV2)
		if !ok {
			return ApplicationInteractionTurn{}, errApplicationSessionContract
		}
		turn := applicationInteractionTurnFromPromptContract(*contract)
		if validateStoredApplicationInteractionTurn(ctx, turn) != nil {
			return ApplicationInteractionTurn{}, errApplicationSessionContract
		}
		return turn, nil
	}
	if envelope.SchemaVersion == agentCopilotSessionTurnV3Schema {
		value, err := decodeAgentCopilotContract(agentCopilotSessionTurnV3Schema, payload)
		if err != nil {
			return ApplicationInteractionTurn{}, errApplicationSessionContract
		}
		contract, ok := value.(*AgentCopilotSessionTurnV3)
		if !ok {
			return ApplicationInteractionTurn{}, errApplicationSessionContract
		}
		turn := applicationInteractionTurnFromAgentCopilotContract(*contract)
		if validateStoredApplicationInteractionTurn(ctx, turn) != nil {
			return ApplicationInteractionTurn{}, errApplicationSessionContract
		}
		return turn, nil
	}
	var turn ApplicationInteractionTurn
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&turn); err != nil {
		return ApplicationInteractionTurn{}, errApplicationSessionContract
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return ApplicationInteractionTurn{}, errApplicationSessionContract
	}
	if err := validateStoredApplicationInteractionTurn(ctx, turn); err != nil {
		return ApplicationInteractionTurn{}, errApplicationSessionContract
	}
	return turn, nil
}

func promptApplicationSessionContractFromInteraction(session ApplicationInteractionSession) (PromptApplicationSessionV2, error) {
	authority, err := promptAuthorityFromApplicationInteraction(session.Authority)
	if err != nil {
		return PromptApplicationSessionV2{}, err
	}
	return PromptApplicationSessionV2{
		SchemaVersion: session.SchemaVersion, SessionID: session.SessionID, TenantRef: session.TenantRef,
		WorkspaceID: session.WorkspaceID, ApplicationID: session.ApplicationID, OwnerSubjectRef: session.OwnerSubjectRef,
		State: session.State, RecordVersion: session.RecordVersion,
		ProfileBinding: PromptApplicationSessionProfileBindingV2{ExecutionProfile: session.ProfileBinding.ExecutionProfile},
		Authority:      authority, ContentRetention: session.ContentRetention, TurnCount: session.TurnCount,
		LastTurnID: session.LastTurnID, CreatedAt: session.CreatedAt, UpdatedAt: session.UpdatedAt,
		ClosedAt: session.ClosedAt, CreatedByActorRef: session.CreatedByActorRef,
		UpdatedByActorRef: session.UpdatedByActorRef, RequestID: session.RequestID, AuditRef: session.AuditRef,
	}, nil
}

func applicationInteractionSessionFromPromptContract(contract PromptApplicationSessionV2) ApplicationInteractionSession {
	return ApplicationInteractionSession{
		SchemaVersion: contract.SchemaVersion, SessionID: contract.SessionID, TenantRef: contract.TenantRef,
		WorkspaceID: contract.WorkspaceID, ApplicationID: contract.ApplicationID, OwnerSubjectRef: contract.OwnerSubjectRef,
		State: contract.State, RecordVersion: contract.RecordVersion,
		ProfileBinding:   ApplicationInteractionProfileBinding{ExecutionProfile: contract.ProfileBinding.ExecutionProfile},
		Authority:        applicationInteractionAuthorityFromPrompt(contract.Authority),
		ContentRetention: contract.ContentRetention, TurnCount: contract.TurnCount,
		LastTurnID: contract.LastTurnID, CreatedAt: contract.CreatedAt, UpdatedAt: contract.UpdatedAt,
		ClosedAt: contract.ClosedAt, CreatedByActorRef: contract.CreatedByActorRef,
		UpdatedByActorRef: contract.UpdatedByActorRef, RequestID: contract.RequestID, AuditRef: contract.AuditRef,
	}
}

func promptApplicationTurnContractFromInteraction(turn ApplicationInteractionTurn) (PromptApplicationSessionTurnV2, error) {
	authority, err := promptAuthorityFromApplicationInteraction(turn.Authority)
	if err != nil {
		return PromptApplicationSessionTurnV2{}, err
	}
	var runRef *PromptApplicationRunRefV6
	if turn.RunRef != nil {
		runRef = &PromptApplicationRunRefV6{RunID: turn.RunRef.RunID, SchemaVersion: turn.RunRef.SchemaVersion}
	}
	return PromptApplicationSessionTurnV2{
		SchemaVersion: turn.SchemaVersion, TurnID: turn.TurnID, SessionID: turn.SessionID, Sequence: turn.Sequence,
		ClientTurnKey: turn.ClientTurnKey, TenantRef: turn.TenantRef, WorkspaceID: turn.WorkspaceID,
		ApplicationID: turn.ApplicationID, OwnerSubjectRef: turn.OwnerSubjectRef,
		ExecutionProfile: turn.ExecutionProfile, Authority: authority, Status: turn.Status,
		InputDigest: turn.InputDigest, InputBytes: turn.InputBytes, RunRef: runRef,
		FailureCode: turn.FailureCode, FailureSummary: turn.FailureSummary,
		StartedAt: turn.StartedAt, CompletedAt: turn.CompletedAt, ActorRef: turn.ActorRef,
		RequestID: turn.RequestID, AuditRef: turn.AuditRef,
	}, nil
}

func applicationInteractionTurnFromPromptContract(contract PromptApplicationSessionTurnV2) ApplicationInteractionTurn {
	var runRef *ApplicationInteractionRunRef
	if contract.RunRef != nil {
		runRef = &ApplicationInteractionRunRef{RunID: contract.RunRef.RunID, SchemaVersion: contract.RunRef.SchemaVersion}
	}
	return ApplicationInteractionTurn{
		SchemaVersion: contract.SchemaVersion, TurnID: contract.TurnID, SessionID: contract.SessionID,
		Sequence: contract.Sequence, ClientTurnKey: contract.ClientTurnKey, TenantRef: contract.TenantRef,
		WorkspaceID: contract.WorkspaceID, ApplicationID: contract.ApplicationID,
		OwnerSubjectRef: contract.OwnerSubjectRef, ExecutionProfile: contract.ExecutionProfile,
		Authority: applicationInteractionAuthorityFromPrompt(contract.Authority), Status: contract.Status,
		InputDigest: contract.InputDigest, InputBytes: contract.InputBytes, RunRef: runRef,
		FailureCode: contract.FailureCode, FailureSummary: contract.FailureSummary,
		StartedAt: contract.StartedAt, CompletedAt: contract.CompletedAt, ActorRef: contract.ActorRef,
		RequestID: contract.RequestID, AuditRef: contract.AuditRef,
	}
}

func agentCopilotSessionContractFromInteraction(session ApplicationInteractionSession) (AgentCopilotSessionV3, error) {
	authority, err := agentCopilotAuthorityFromApplicationInteraction(session.Authority)
	if err != nil {
		return AgentCopilotSessionV3{}, err
	}
	return AgentCopilotSessionV3{
		SchemaVersion: session.SchemaVersion, SessionID: session.SessionID, TenantRef: session.TenantRef,
		WorkspaceID: session.WorkspaceID, ApplicationID: session.ApplicationID, OwnerSubjectRef: session.OwnerSubjectRef,
		State: session.State, RecordVersion: session.RecordVersion,
		ProfileBinding: PromptApplicationSessionProfileBindingV2{ExecutionProfile: session.ProfileBinding.ExecutionProfile},
		Authority:      authority, ContentRetention: session.ContentRetention, TurnCount: session.TurnCount,
		LastTurnID: session.LastTurnID, CreatedAt: session.CreatedAt, UpdatedAt: session.UpdatedAt,
		ClosedAt: session.ClosedAt, CreatedByActorRef: session.CreatedByActorRef,
		UpdatedByActorRef: session.UpdatedByActorRef, RequestID: session.RequestID, AuditRef: session.AuditRef,
	}, nil
}

func applicationInteractionSessionFromAgentCopilotContract(contract AgentCopilotSessionV3) ApplicationInteractionSession {
	return ApplicationInteractionSession{
		SchemaVersion: contract.SchemaVersion, SessionID: contract.SessionID, TenantRef: contract.TenantRef,
		WorkspaceID: contract.WorkspaceID, ApplicationID: contract.ApplicationID, OwnerSubjectRef: contract.OwnerSubjectRef,
		State: contract.State, RecordVersion: contract.RecordVersion,
		ProfileBinding:   ApplicationInteractionProfileBinding{ExecutionProfile: contract.ProfileBinding.ExecutionProfile},
		Authority:        applicationInteractionAuthorityFromAgentCopilot(contract.Authority),
		ContentRetention: contract.ContentRetention, TurnCount: contract.TurnCount,
		LastTurnID: contract.LastTurnID, CreatedAt: contract.CreatedAt, UpdatedAt: contract.UpdatedAt,
		ClosedAt: contract.ClosedAt, CreatedByActorRef: contract.CreatedByActorRef,
		UpdatedByActorRef: contract.UpdatedByActorRef, RequestID: contract.RequestID, AuditRef: contract.AuditRef,
	}
}

func agentCopilotTurnContractFromInteraction(turn ApplicationInteractionTurn) (AgentCopilotSessionTurnV3, error) {
	authority, err := agentCopilotAuthorityFromApplicationInteraction(turn.Authority)
	if err != nil {
		return AgentCopilotSessionTurnV3{}, err
	}
	var runRef *AgentCopilotRunRefV7
	if turn.RunRef != nil {
		runRef = &AgentCopilotRunRefV7{RunID: turn.RunRef.RunID, SchemaVersion: turn.RunRef.SchemaVersion}
	}
	return AgentCopilotSessionTurnV3{
		SchemaVersion: turn.SchemaVersion, TurnID: turn.TurnID, SessionID: turn.SessionID, Sequence: turn.Sequence,
		ClientTurnKey: turn.ClientTurnKey, TenantRef: turn.TenantRef, WorkspaceID: turn.WorkspaceID,
		ApplicationID: turn.ApplicationID, OwnerSubjectRef: turn.OwnerSubjectRef,
		ExecutionProfile: turn.ExecutionProfile, Authority: authority, Status: turn.Status,
		InputDigest: turn.InputDigest, InputBytes: turn.InputBytes, RunRef: runRef,
		FailureCode: turn.FailureCode, FailureSummary: turn.FailureSummary,
		StartedAt: turn.StartedAt, CompletedAt: turn.CompletedAt, ActorRef: turn.ActorRef,
		RequestID: turn.RequestID, AuditRef: turn.AuditRef,
	}, nil
}

func applicationInteractionTurnFromAgentCopilotContract(contract AgentCopilotSessionTurnV3) ApplicationInteractionTurn {
	var runRef *ApplicationInteractionRunRef
	if contract.RunRef != nil {
		runRef = &ApplicationInteractionRunRef{RunID: contract.RunRef.RunID, SchemaVersion: contract.RunRef.SchemaVersion}
	}
	return ApplicationInteractionTurn{
		SchemaVersion: contract.SchemaVersion, TurnID: contract.TurnID, SessionID: contract.SessionID,
		Sequence: contract.Sequence, ClientTurnKey: contract.ClientTurnKey, TenantRef: contract.TenantRef,
		WorkspaceID: contract.WorkspaceID, ApplicationID: contract.ApplicationID,
		OwnerSubjectRef: contract.OwnerSubjectRef, ExecutionProfile: contract.ExecutionProfile,
		Authority: applicationInteractionAuthorityFromAgentCopilot(contract.Authority), Status: contract.Status,
		InputDigest: contract.InputDigest, InputBytes: contract.InputBytes, RunRef: runRef,
		FailureCode: contract.FailureCode, FailureSummary: contract.FailureSummary,
		StartedAt: contract.StartedAt, CompletedAt: contract.CompletedAt, ActorRef: contract.ActorRef,
		RequestID: contract.RequestID, AuditRef: contract.AuditRef,
	}
}

func applicationInteractionTimestamp(value string) (time.Time, error) {
	parsed := parseApplicationInteractionTimestamp(value)
	if parsed == nil {
		return time.Time{}, errApplicationSessionContract
	}
	return parsed.UTC(), nil
}

func applicationInteractionCompletedTimestamp(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := applicationInteractionTimestamp(*value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

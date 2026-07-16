package httpapi

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	workflowHTTPToolID                 = "workflow.http.reviewed-json-read.v1"
	workflowHTTPToolVersion            = 1
	workflowHTTPToolDefinitionSchema   = "workflow_http_tool_definition.v1"
	workflowHTTPToolProfileSchema      = "workflow_http_tool_execution_profile.v1"
	workflowHTTPToolPlanSchema         = "workflow_http_tool_action_plan.v1"
	workflowHTTPToolDecisionSchema     = "workflow_http_tool_confirmation_decision.v1"
	workflowHTTPToolAuditSchema        = "workflow_http_tool_execution_audit.v1"
	workflowHTTPToolProfileID          = "workflow_http_profile_reviewed_json_read_dev_v1"
	workflowHTTPToolProfileVersion     = 1
	workflowHTTPToolTargetPolicyKey    = "reviewed_json_read_dev"
	workflowHTTPToolDefaultExpiry      = 15 * time.Minute
	workflowHTTPToolTimeoutMS          = 5000
	workflowHTTPToolMaxResponseBytes   = 64 * 1024
	workflowHTTPToolMaxOutputBytes     = 16 * 1024
	workflowHTTPToolMaxResourceKeySize = 160
	workflowHTTPToolMaxLocaleSize      = 35
)

var workflowHTTPToolResourceKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._:/-]*$`)
var workflowHTTPToolLocalePattern = regexp.MustCompile(`^[A-Za-z]{2,8}(-[A-Za-z0-9]{1,8})*$`)
var workflowHTTPToolIDPattern = regexp.MustCompile(`^workflow\.http\.[a-z0-9][a-z0-9._-]*\.v[0-9]+$`)
var workflowHTTPToolProfileIDPattern = regexp.MustCompile(`^workflow_http_profile_[a-z0-9_]{1,80}$`)
var workflowHTTPToolTargetPolicyKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_.-]{2,79}$`)
var workflowHTTPToolPlanIDPattern = regexp.MustCompile(`^wtap_[a-z0-9]{16,64}$`)
var workflowHTTPToolConfirmationIDPattern = regexp.MustCompile(`^wtcd_[a-z0-9]{16,64}$`)
var workflowHTTPToolAuditIDPattern = regexp.MustCompile(`^wtae_[a-z0-9]{16,64}$`)
var workflowHTTPToolDigestPattern = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)
var workflowHTTPToolReferencePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:/-]{2,159}$`)
var workflowHTTPToolScopedIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{2,119}$`)

type WorkflowHTTPToolActionFailureCode string

const (
	WorkflowHTTPToolActionFailureScopeDenied        WorkflowHTTPToolActionFailureCode = "workflow_tool_action_scope_denied"
	WorkflowHTTPToolActionFailureInputInvalid       WorkflowHTTPToolActionFailureCode = "workflow_tool_arguments_invalid"
	WorkflowHTTPToolActionFailureDraftNotFound      WorkflowHTTPToolActionFailureCode = "workflow_tool_action_draft_not_found"
	WorkflowHTTPToolActionFailureDraftVersion       WorkflowHTTPToolActionFailureCode = "workflow_tool_confirmation_stale"
	WorkflowHTTPToolActionFailureDraftIneligible    WorkflowHTTPToolActionFailureCode = "workflow_tool_action_draft_ineligible"
	WorkflowHTTPToolActionFailureDefinitionMissing  WorkflowHTTPToolActionFailureCode = "workflow_tool_not_registered"
	WorkflowHTTPToolActionFailureProfileUnavailable WorkflowHTTPToolActionFailureCode = "workflow_tool_profile_disabled"
	WorkflowHTTPToolActionFailureNotFound           WorkflowHTTPToolActionFailureCode = "workflow_tool_action_plan_not_found"
	WorkflowHTTPToolActionFailureConflict           WorkflowHTTPToolActionFailureCode = "workflow_tool_confirmation_stale"
	WorkflowHTTPToolActionFailureRejected           WorkflowHTTPToolActionFailureCode = "workflow_tool_confirmation_rejected"
	WorkflowHTTPToolActionFailureExpired            WorkflowHTTPToolActionFailureCode = "workflow_tool_confirmation_expired"
	WorkflowHTTPToolActionFailureInvalidated        WorkflowHTTPToolActionFailureCode = "workflow_tool_confirmation_invalidated"
	WorkflowHTTPToolActionFailureCanceled           WorkflowHTTPToolActionFailureCode = "workflow_tool_action_canceled"
	WorkflowHTTPToolActionFailureConsumed           WorkflowHTTPToolActionFailureCode = "workflow_tool_action_consumed"
	WorkflowHTTPToolActionFailureStoreUnavailable   WorkflowHTTPToolActionFailureCode = "workflow_tool_store_unavailable"
	WorkflowHTTPToolActionFailureStoreContract      WorkflowHTTPToolActionFailureCode = "workflow_tool_store_contract_mismatch"
)

type WorkflowHTTPToolActionStatus string

const (
	WorkflowHTTPToolActionStatusPending     WorkflowHTTPToolActionStatus = "pending"
	WorkflowHTTPToolActionStatusDeferred    WorkflowHTTPToolActionStatus = "deferred"
	WorkflowHTTPToolActionStatusApproved    WorkflowHTTPToolActionStatus = "approved"
	WorkflowHTTPToolActionStatusRejected    WorkflowHTTPToolActionStatus = "rejected"
	WorkflowHTTPToolActionStatusCanceled    WorkflowHTTPToolActionStatus = "canceled"
	WorkflowHTTPToolActionStatusExpired     WorkflowHTTPToolActionStatus = "expired"
	WorkflowHTTPToolActionStatusInvalidated WorkflowHTTPToolActionStatus = "invalidated"
	WorkflowHTTPToolActionStatusConsumed    WorkflowHTTPToolActionStatus = "consumed"
)

type WorkflowHTTPToolConfirmationOutcome string

const (
	WorkflowHTTPToolConfirmationApprove    WorkflowHTTPToolConfirmationOutcome = "approve"
	WorkflowHTTPToolConfirmationReject     WorkflowHTTPToolConfirmationOutcome = "reject"
	WorkflowHTTPToolConfirmationDefer      WorkflowHTTPToolConfirmationOutcome = "defer"
	WorkflowHTTPToolConfirmationCancel     WorkflowHTTPToolConfirmationOutcome = "cancel"
	workflowHTTPToolConfirmationExpire     WorkflowHTTPToolConfirmationOutcome = "expire"
	workflowHTTPToolConfirmationInvalidate WorkflowHTTPToolConfirmationOutcome = "invalidate"
)

type WorkflowHTTPToolActionContext struct {
	RequestContext context.Context
	RequestID      string
	TenantRef      string
	WorkspaceID    string
	ApplicationID  string
	ActorRef       string
	ScopeGrants    []string
	AuditRef       string
}

type WorkflowHTTPToolDefinition struct {
	SchemaVersion         string                                `json:"schema_version"`
	ToolID                string                                `json:"tool_id"`
	ToolVersion           int                                   `json:"tool_version"`
	DisplayName           string                                `json:"display_name"`
	Purpose               string                                `json:"purpose"`
	OperationKind         string                                `json:"operation_kind"`
	PublicArgumentsSchema WorkflowHTTPToolPublicArgumentsSchema `json:"public_arguments_schema"`
	OutputSchema          WorkflowHTTPToolOutputSchema          `json:"output_schema"`
	OutputFields          []string                              `json:"output_fields"`
	RequiresConfirmation  bool                                  `json:"requires_confirmation"`
	RiskLevel             string                                `json:"risk_level"`
	ReadOnly              bool                                  `json:"read_only"`
	WritesBusinessTruth   bool                                  `json:"writes_business_truth"`
	AuditRef              string                                `json:"audit_ref"`
}

type WorkflowHTTPToolStringArgumentSchema struct {
	Type      string `json:"type"`
	MinLength int    `json:"min_length"`
	MaxLength int    `json:"max_length"`
	Pattern   string `json:"pattern,omitempty"`
}

type WorkflowHTTPToolPublicArgumentsSchema struct {
	Type                 string                                          `json:"type"`
	AdditionalProperties bool                                            `json:"additional_properties"`
	MaxProperties        int                                             `json:"max_properties"`
	MaxSerializedBytes   int                                             `json:"max_serialized_bytes"`
	Required             []string                                        `json:"required"`
	Properties           map[string]WorkflowHTTPToolStringArgumentSchema `json:"properties"`
}

type WorkflowHTTPToolOutputFieldSchema struct {
	Type      string `json:"type"`
	MinLength int    `json:"minLength,omitempty"`
	MaxLength int    `json:"maxLength,omitempty"`
	Format    string `json:"format,omitempty"`
}

type WorkflowHTTPToolOutputSchema struct {
	Type                 string                                       `json:"type"`
	AdditionalProperties bool                                         `json:"additionalProperties"`
	Required             []string                                     `json:"required"`
	Properties           map[string]WorkflowHTTPToolOutputFieldSchema `json:"properties"`
}

type WorkflowHTTPToolExecutionProfile struct {
	SchemaVersion    string                        `json:"schema_version"`
	ProfileID        string                        `json:"profile_id"`
	ProfileVersion   int                           `json:"profile_version"`
	ToolID           string                        `json:"tool_id"`
	DefinitionDigest string                        `json:"definition_digest"`
	Environment      string                        `json:"environment"`
	Enabled          bool                          `json:"enabled"`
	Method           string                        `json:"method"`
	TargetPolicyKey  string                        `json:"target_policy_key"`
	TargetPolicy     WorkflowHTTPToolTargetPolicy  `json:"target_policy"`
	CredentialPolicy string                        `json:"credential_policy"`
	TimeoutMS        int                           `json:"timeout_ms"`
	MaxResponseBytes int                           `json:"max_response_bytes"`
	MaxOutputBytes   int                           `json:"max_output_bytes"`
	MaxAttempts      int                           `json:"max_attempts"`
	NetworkPolicy    WorkflowHTTPToolNetworkPolicy `json:"network_policy"`
	TestOnlyLoopback bool                          `json:"test_only_loopback"`
	AuditRef         string                        `json:"audit_ref"`
}

type WorkflowHTTPToolQueryMapping struct {
	ArgumentKey string `json:"argument_key"`
	QueryKey    string `json:"query_key"`
}

type WorkflowHTTPToolTargetPolicy struct {
	Scheme        string                         `json:"scheme"`
	Host          string                         `json:"host"`
	Port          int                            `json:"port"`
	Path          string                         `json:"path"`
	QueryMappings []WorkflowHTTPToolQueryMapping `json:"query_mappings"`
}

type WorkflowHTTPToolNetworkPolicy struct {
	FollowRedirects                   bool `json:"follow_redirects"`
	UseAmbientProxy                   bool `json:"use_ambient_proxy"`
	AutomaticRetry                    bool `json:"automatic_retry"`
	FallbackEnabled                   bool `json:"fallback_enabled"`
	CrossProfileConnectionReuse       bool `json:"cross_profile_connection_reuse"`
	RequireAllResolvedAddressesPublic bool `json:"require_all_resolved_addresses_public"`
	PinValidatedAddress               bool `json:"pin_validated_address"`
	PreserveTLSServerName             bool `json:"preserve_tls_server_name"`
}

type WorkflowHTTPToolPublicArguments struct {
	ResourceKey string `json:"resource_key"`
	Locale      string `json:"locale,omitempty"`
}

type WorkflowHTTPToolActionPlan struct {
	SchemaVersion          string                          `json:"schema_version"`
	PlanID                 string                          `json:"plan_id"`
	RecordVersion          int                             `json:"record_version"`
	TenantRef              string                          `json:"tenant_ref"`
	WorkspaceID            string                          `json:"workspace_id"`
	ApplicationID          string                          `json:"application_id"`
	DraftID                string                          `json:"draft_id"`
	DraftVersion           int                             `json:"draft_version"`
	NodeID                 string                          `json:"node_id"`
	ToolID                 string                          `json:"tool_id"`
	ToolVersion            int                             `json:"tool_version"`
	DefinitionDigest       string                          `json:"definition_digest"`
	ProfileID              string                          `json:"profile_id"`
	ProfileVersion         int                             `json:"profile_version"`
	ProfileDigest          string                          `json:"profile_digest"`
	Method                 string                          `json:"method"`
	TargetPolicyKey        string                          `json:"target_policy_key"`
	PublicArguments        WorkflowHTTPToolPublicArguments `json:"public_arguments"`
	OutputFields           []string                        `json:"output_fields"`
	OutputSchemaDigest     string                          `json:"output_schema_digest"`
	CredentialPolicy       string                          `json:"credential_policy"`
	TimeoutMS              int                             `json:"timeout_ms"`
	MaxResponseBytes       int                             `json:"max_response_bytes"`
	MaxOutputBytes         int                             `json:"max_output_bytes"`
	PlannedByActorRef      string                          `json:"planned_by_actor_ref"`
	CreatedAt              string                          `json:"created_at"`
	ExpiresAt              string                          `json:"expires_at"`
	ToolPlanDigest         string                          `json:"tool_plan_digest"`
	Status                 WorkflowHTTPToolActionStatus    `json:"status"`
	LastDecisionByActorRef *string                         `json:"last_decision_by_actor_ref"`
	LastDecisionAt         *string                         `json:"last_decision_at"`
	AuditRef               string                          `json:"audit_ref"`
}

type WorkflowHTTPToolConfirmationDecision struct {
	SchemaVersion          string                              `json:"schema_version"`
	ConfirmationID         string                              `json:"confirmation_id"`
	PlanID                 string                              `json:"plan_id"`
	TenantRef              string                              `json:"tenant_ref"`
	WorkspaceID            string                              `json:"workspace_id"`
	ApplicationID          string                              `json:"application_id"`
	DraftID                string                              `json:"draft_id"`
	DraftVersion           int                                 `json:"draft_version"`
	NodeID                 string                              `json:"node_id"`
	ToolID                 string                              `json:"tool_id"`
	ToolVersion            int                                 `json:"tool_version"`
	ToolPlanDigest         string                              `json:"tool_plan_digest"`
	Outcome                WorkflowHTTPToolConfirmationOutcome `json:"outcome"`
	DecidedByActorRef      string                              `json:"decided_by_actor_ref"`
	ActorSource            string                              `json:"actor_source"`
	DecidedAt              string                              `json:"decided_at"`
	ReasonCode             string                              `json:"reason_code"`
	ExpectedRecordVersion  int                                 `json:"expected_record_version"`
	ResultingRecordVersion int                                 `json:"resulting_record_version"`
	AuditRef               string                              `json:"audit_ref"`
}

type WorkflowHTTPToolExecutionAudit struct {
	SchemaVersion      string                           `json:"schema_version"`
	EventID            string                           `json:"event_id"`
	EventKind          string                           `json:"event_kind"`
	OccurredAt         string                           `json:"occurred_at"`
	TenantRef          string                           `json:"tenant_ref"`
	WorkspaceID        string                           `json:"workspace_id"`
	ApplicationID      string                           `json:"application_id"`
	DraftID            string                           `json:"draft_id"`
	DraftVersion       int                              `json:"draft_version"`
	NodeID             string                           `json:"node_id"`
	PlanID             string                           `json:"plan_id"`
	ConfirmationID     *string                          `json:"confirmation_id"`
	ExecutionAttemptID *string                          `json:"execution_attempt_id"`
	RunID              *string                          `json:"run_id"`
	ToolID             string                           `json:"tool_id"`
	ToolVersion        int                              `json:"tool_version"`
	DefinitionDigest   string                           `json:"definition_digest"`
	ProfileID          string                           `json:"profile_id"`
	ProfileDigest      string                           `json:"profile_digest"`
	ToolPlanDigest     string                           `json:"tool_plan_digest"`
	ActorRef           string                           `json:"actor_ref"`
	ActorSource        string                           `json:"actor_source"`
	RequestID          string                           `json:"request_id"`
	AuditRef           string                           `json:"audit_ref"`
	FailureCode        *string                          `json:"failure_code"`
	FailureBoundary    *string                          `json:"failure_boundary"`
	AttemptStatus      string                           `json:"attempt_status"`
	HTTPStatusClass    *string                          `json:"http_status_class"`
	ResponseBytes      int                              `json:"response_bytes"`
	DurationMS         int                              `json:"duration_ms"`
	SideEffects        WorkflowHTTPToolAuditSideEffects `json:"side_effects"`
}

type WorkflowHTTPToolAuditSideEffects struct {
	NetworkAttempts int `json:"network_attempts"`
	ToolCalls       int `json:"tool_calls"`
	ProviderCalls   int `json:"provider_calls"`
	BusinessWrites  int `json:"business_writes"`
	ReplayWrites    int `json:"replay_writes"`
}

type WorkflowHTTPToolCreatePlanRequest struct {
	DraftID         string
	DraftVersion    int
	NodeID          string
	PublicArguments map[string]any
}

type WorkflowHTTPToolDecisionRequest struct {
	PlanID                string
	ExpectedRecordVersion int
	Decision              WorkflowHTTPToolConfirmationOutcome
}

type WorkflowHTTPToolActionResult struct {
	ActionPlan           *WorkflowHTTPToolActionPlan
	ConfirmationDecision *WorkflowHTTPToolConfirmationDecision
	FailureCode          WorkflowHTTPToolActionFailureCode
	FailureSummary       string
}

var (
	errWorkflowHTTPToolActionConflict    = errors.New("workflow HTTP tool action version conflict")
	errWorkflowHTTPToolActionContract    = errors.New("workflow HTTP tool action store contract mismatch")
	errWorkflowHTTPToolActionUnavailable = errors.New("workflow HTTP tool action store unavailable")
)

type workflowHTTPToolActionStore interface {
	CreatePlan(WorkflowHTTPToolActionContext, *WorkflowHTTPToolActionPlan, WorkflowHTTPToolExecutionAudit) error
	ReadPlan(WorkflowHTTPToolActionContext, string) (WorkflowHTTPToolActionPlan, bool, error)
	DecidePlan(WorkflowHTTPToolActionContext, *WorkflowHTTPToolActionPlan, WorkflowHTTPToolConfirmationDecision, WorkflowHTTPToolExecutionAudit) error
}

type workflowHTTPToolRegistry struct {
	definition         WorkflowHTTPToolDefinition
	profile            WorkflowHTTPToolExecutionProfile
	definitionDigest   string
	profileDigest      string
	outputSchemaDigest string
}

func newWorkflowHTTPToolRegistry() (workflowHTTPToolRegistry, error) {
	definition := WorkflowHTTPToolDefinition{
		SchemaVersion: workflowHTTPToolDefinitionSchema, ToolID: workflowHTTPToolID, ToolVersion: workflowHTTPToolVersion,
		DisplayName:   "Reviewed JSON Read",
		Purpose:       "Read one allowlisted JSON resource after an independent human confirmation.",
		OperationKind: "http_read",
		PublicArgumentsSchema: WorkflowHTTPToolPublicArgumentsSchema{
			Type: "object", AdditionalProperties: false, MaxProperties: 2, MaxSerializedBytes: 8192,
			Required: []string{"resource_key"},
			Properties: map[string]WorkflowHTTPToolStringArgumentSchema{
				"resource_key": {Type: "string", MinLength: 1, MaxLength: 160, Pattern: `^(?![A-Za-z][A-Za-z0-9+.-]*://)[A-Za-z0-9][A-Za-z0-9._:/-]*$`},
				"locale":       {Type: "string", MinLength: 2, MaxLength: 35, Pattern: `^[A-Za-z]{2,8}(-[A-Za-z0-9]{1,8})*$`},
			},
		},
		OutputSchema:         workflowHTTPToolOutputSchemaV1(),
		OutputFields:         workflowHTTPToolOutputFieldsV1(),
		RequiresConfirmation: true, RiskLevel: "medium", ReadOnly: true, WritesBusinessTruth: false,
		AuditRef: "audit_workflow_http_tool_definition_v1",
	}
	definitionDigest, err := canonicalSHA256(definition)
	if err != nil {
		return workflowHTTPToolRegistry{}, err
	}
	outputSchemaDigest, err := canonicalSHA256(definition.OutputSchema)
	if err != nil {
		return workflowHTTPToolRegistry{}, err
	}
	profile := WorkflowHTTPToolExecutionProfile{
		SchemaVersion: workflowHTTPToolProfileSchema, ProfileID: workflowHTTPToolProfileID,
		ProfileVersion: workflowHTTPToolProfileVersion, ToolID: workflowHTTPToolID,
		DefinitionDigest: definitionDigest, Environment: "development", Enabled: true, Method: "GET",
		TargetPolicyKey: workflowHTTPToolTargetPolicyKey, CredentialPolicy: "none",
		TargetPolicy: WorkflowHTTPToolTargetPolicy{
			Scheme: "https", Host: "api.dev.example.invalid", Port: 443, Path: "/v1/reviewed-resources",
			QueryMappings: []WorkflowHTTPToolQueryMapping{
				{ArgumentKey: "resource_key", QueryKey: "resource_key"},
				{ArgumentKey: "locale", QueryKey: "locale"},
			},
		},
		TimeoutMS: workflowHTTPToolTimeoutMS, MaxResponseBytes: workflowHTTPToolMaxResponseBytes,
		MaxOutputBytes: workflowHTTPToolMaxOutputBytes, MaxAttempts: 1,
		NetworkPolicy: WorkflowHTTPToolNetworkPolicy{
			RequireAllResolvedAddressesPublic: true, PinValidatedAddress: true, PreserveTLSServerName: true,
		},
		TestOnlyLoopback: false, AuditRef: "audit_workflow_http_profile_v1",
	}
	profileDigest, err := canonicalSHA256(profile)
	if err != nil {
		return workflowHTTPToolRegistry{}, err
	}
	return workflowHTTPToolRegistry{
		definition: definition, profile: profile, definitionDigest: definitionDigest,
		profileDigest: profileDigest, outputSchemaDigest: outputSchemaDigest,
	}, nil
}

func workflowHTTPToolOutputFieldsV1() []string {
	return []string{"resource_key", "title", "summary", "updated_at"}
}

func workflowHTTPToolOutputSchemaV1() WorkflowHTTPToolOutputSchema {
	return WorkflowHTTPToolOutputSchema{
		Type: "object", AdditionalProperties: false, Required: workflowHTTPToolOutputFieldsV1(),
		Properties: map[string]WorkflowHTTPToolOutputFieldSchema{
			"resource_key": {Type: "string", MinLength: 1, MaxLength: 160},
			"title":        {Type: "string", MinLength: 1, MaxLength: 240},
			"summary":      {Type: "string", MinLength: 1, MaxLength: 4096},
			"updated_at":   {Type: "string", Format: "date-time"},
		},
	}
}

type workflowHTTPToolActionService struct {
	readDraft func(SavedWorkflowDraftContext, ReadWorkflowDraftRequest) SavedWorkflowDraftResult
	store     workflowHTTPToolActionStore
	registry  workflowHTTPToolRegistry
	now       func() time.Time
	newID     func(string) (string, error)
}

func newWorkflowHTTPToolActionService(
	readDraft func(SavedWorkflowDraftContext, ReadWorkflowDraftRequest) SavedWorkflowDraftResult,
	store workflowHTTPToolActionStore,
) (workflowHTTPToolActionService, error) {
	registry, err := newWorkflowHTTPToolRegistry()
	if err != nil {
		return workflowHTTPToolActionService{}, err
	}
	return workflowHTTPToolActionService{readDraft: readDraft, store: store, registry: registry, now: func() time.Time { return time.Now().UTC() }, newID: newWorkflowHTTPToolActionID}, nil
}

func (service workflowHTTPToolActionService) CreatePlan(ctx WorkflowHTTPToolActionContext, request WorkflowHTTPToolCreatePlanRequest) WorkflowHTTPToolActionResult {
	if failure := validateWorkflowHTTPToolActionContext(ctx); failure != "" {
		return workflowHTTPToolActionFailure(failure, "Workflow tool action scope is denied.")
	}
	if service.store == nil || service.readDraft == nil {
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureStoreUnavailable, "Workflow HTTP tool action service is unavailable.")
	}
	if service.registry.definition.ToolID != workflowHTTPToolID || service.registry.definition.ToolVersion != workflowHTTPToolVersion || service.registry.definitionDigest == "" {
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureDefinitionMissing, "Workflow HTTP tool definition is not registered.")
	}
	if !service.registry.profile.Enabled || service.registry.profile.ToolID != workflowHTTPToolID ||
		service.registry.profile.DefinitionDigest != service.registry.definitionDigest || service.registry.profileDigest == "" ||
		service.registry.outputSchemaDigest == "" {
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureProfileUnavailable, "Workflow HTTP tool profile is unavailable.")
	}
	arguments, failure := normalizeWorkflowHTTPToolPublicArguments(request.PublicArguments)
	if failure != "" || strings.TrimSpace(request.DraftID) == "" || request.DraftVersion < 1 || strings.TrimSpace(request.NodeID) == "" {
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureInputInvalid, "Workflow HTTP tool plan input is invalid.")
	}
	draft, result := service.readExactEligibleDraft(ctx, ctx.ActorRef, request.DraftID, request.DraftVersion, request.NodeID)
	if result.FailureCode != "" {
		return result
	}
	now := service.now().UTC()
	planID, err := service.newID("wtap_")
	if err != nil {
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureStoreUnavailable, "Workflow HTTP tool plan id could not be created.")
	}
	plan := WorkflowHTTPToolActionPlan{
		SchemaVersion: workflowHTTPToolPlanSchema, PlanID: planID, RecordVersion: 1,
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		DraftID: draft.DraftID, DraftVersion: draft.DraftVersion, NodeID: strings.TrimSpace(request.NodeID),
		ToolID: service.registry.definition.ToolID, ToolVersion: service.registry.definition.ToolVersion,
		DefinitionDigest: service.registry.definitionDigest,
		ProfileID:        service.registry.profile.ProfileID, ProfileVersion: service.registry.profile.ProfileVersion,
		ProfileDigest: service.registry.profileDigest, Method: service.registry.profile.Method,
		TargetPolicyKey: service.registry.profile.TargetPolicyKey, PublicArguments: arguments,
		OutputFields:       append([]string(nil), service.registry.definition.OutputFields...),
		OutputSchemaDigest: service.registry.outputSchemaDigest,
		CredentialPolicy:   service.registry.profile.CredentialPolicy, TimeoutMS: service.registry.profile.TimeoutMS,
		MaxResponseBytes: service.registry.profile.MaxResponseBytes, MaxOutputBytes: service.registry.profile.MaxOutputBytes,
		PlannedByActorRef: ctx.ActorRef, CreatedAt: workflowRunTimestamp(now),
		ExpiresAt: workflowRunTimestamp(now.Add(workflowHTTPToolDefaultExpiry)), Status: WorkflowHTTPToolActionStatusPending,
		AuditRef: ctx.AuditRef,
	}
	plan.ToolPlanDigest, err = workflowHTTPToolPlanDigest(plan)
	if err != nil {
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureStoreContract, "Workflow HTTP tool plan could not be canonicalized.")
	}
	audit, err := service.newAudit(ctx, plan, "confirmation_requested", nil, "human", now)
	if err != nil {
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureStoreUnavailable, "Workflow HTTP tool plan audit id could not be created.")
	}
	if err := service.store.CreatePlan(ctx, &plan, audit); err != nil {
		return workflowHTTPToolActionStoreFailure(err)
	}
	return WorkflowHTTPToolActionResult{ActionPlan: workflowHTTPToolActionPlanPointer(plan)}
}

func (service workflowHTTPToolActionService) ReadPlan(ctx WorkflowHTTPToolActionContext, planID string) WorkflowHTTPToolActionResult {
	if failure := validateWorkflowHTTPToolActionContext(ctx); failure != "" {
		return workflowHTTPToolActionFailure(failure, "Workflow tool action scope is denied.")
	}
	plan, found, err := service.store.ReadPlan(ctx, strings.TrimSpace(planID))
	if err != nil {
		return workflowHTTPToolActionStoreFailure(err)
	}
	if !found {
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureNotFound, "Workflow HTTP tool action plan was not found.")
	}
	plan, transition := service.applySystemTransitionIfRequired(ctx, plan, service.now().UTC())
	if transition.FailureCode != "" {
		return transition
	}
	return WorkflowHTTPToolActionResult{ActionPlan: workflowHTTPToolActionPlanPointer(plan), ConfirmationDecision: transition.ConfirmationDecision}
}

func (service workflowHTTPToolActionService) DecidePlan(ctx WorkflowHTTPToolActionContext, request WorkflowHTTPToolDecisionRequest) WorkflowHTTPToolActionResult {
	if failure := validateWorkflowHTTPToolActionContext(ctx); failure != "" {
		return workflowHTTPToolActionFailure(failure, "Workflow tool action scope is denied.")
	}
	if !workflowHTTPToolHumanDecisionAllowed(request.Decision) || request.ExpectedRecordVersion < 1 {
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureInputInvalid, "Workflow HTTP tool confirmation decision is invalid.")
	}
	plan, found, err := service.store.ReadPlan(ctx, strings.TrimSpace(request.PlanID))
	if err != nil {
		return workflowHTTPToolActionStoreFailure(err)
	}
	if !found {
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureNotFound, "Workflow HTTP tool action plan was not found.")
	}
	if plan.RecordVersion != request.ExpectedRecordVersion {
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureConflict, "Workflow HTTP tool action plan changed before the decision was applied.")
	}
	decidedAt := service.now().UTC()
	if refreshed, transition := service.applySystemTransitionIfRequired(ctx, plan, decidedAt); transition.FailureCode != "" || refreshed.RecordVersion != plan.RecordVersion {
		if transition.FailureCode != "" {
			return transition
		}
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureConflict, "Workflow HTTP tool action plan expired or became invalid before the decision.")
	}
	if !workflowHTTPToolDecisionAllowedFrom(plan.Status, request.Decision) {
		return workflowHTTPToolActionTerminalFailure(plan.Status)
	}
	return service.transition(ctx, plan, request.Decision, decidedAt)
}

func (service workflowHTTPToolActionService) applySystemTransitionIfRequired(
	ctx WorkflowHTTPToolActionContext,
	plan WorkflowHTTPToolActionPlan,
	now time.Time,
) (WorkflowHTTPToolActionPlan, WorkflowHTTPToolActionResult) {
	if workflowHTTPToolTerminalStatus(plan.Status) {
		return plan, WorkflowHTTPToolActionResult{}
	}
	expiresAt, err := time.Parse(time.RFC3339Nano, plan.ExpiresAt)
	if err != nil {
		return plan, workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureStoreContract, "Workflow HTTP tool plan expiry is invalid.")
	}
	now = now.UTC()
	if !now.Before(expiresAt) {
		result := service.transition(ctx, plan, workflowHTTPToolConfirmationExpire, now)
		return workflowHTTPToolActionPlanValue(result.ActionPlan, plan), result
	}
	sourcesMatch, sourceFailure := service.planSourcesMatch(ctx, plan)
	if sourceFailure.FailureCode != "" {
		return plan, sourceFailure
	}
	if !sourcesMatch {
		result := service.transition(ctx, plan, workflowHTTPToolConfirmationInvalidate, now)
		return workflowHTTPToolActionPlanValue(result.ActionPlan, plan), result
	}
	return plan, WorkflowHTTPToolActionResult{}
}

func (service workflowHTTPToolActionService) transition(ctx WorkflowHTTPToolActionContext, plan WorkflowHTTPToolActionPlan, outcome WorkflowHTTPToolConfirmationOutcome, decidedAt time.Time) WorkflowHTTPToolActionResult {
	confirmationID, err := service.newID("wtcd_")
	if err != nil {
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureStoreUnavailable, "Workflow HTTP tool confirmation id could not be created.")
	}
	previousVersion := plan.RecordVersion
	actorRef, actorSource := ctx.ActorRef, "human"
	if outcome == workflowHTTPToolConfirmationExpire || outcome == workflowHTTPToolConfirmationInvalidate {
		actorRef, actorSource = "system:workflow_tool_action_policy", "system"
	}
	plan.Status = workflowHTTPToolStatusForOutcome(outcome)
	plan.RecordVersion++
	decisionTime := workflowRunTimestamp(decidedAt)
	plan.LastDecisionByActorRef = &actorRef
	plan.LastDecisionAt = &decisionTime
	decision := WorkflowHTTPToolConfirmationDecision{
		SchemaVersion: workflowHTTPToolDecisionSchema, ConfirmationID: confirmationID, PlanID: plan.PlanID,
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		DraftID: plan.DraftID, DraftVersion: plan.DraftVersion, NodeID: plan.NodeID,
		ToolID: plan.ToolID, ToolVersion: plan.ToolVersion,
		ToolPlanDigest: plan.ToolPlanDigest, Outcome: outcome, DecidedByActorRef: actorRef, ActorSource: actorSource,
		DecidedAt: decisionTime, ReasonCode: workflowHTTPToolConfirmationReason(outcome), ExpectedRecordVersion: previousVersion,
		ResultingRecordVersion: plan.RecordVersion, AuditRef: ctx.AuditRef,
	}
	audit, err := service.newAudit(ctx, plan, workflowHTTPToolAuditEventForOutcome(outcome), &confirmationID, actorSource, decidedAt)
	if err != nil {
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureStoreUnavailable, "Workflow HTTP tool audit id could not be created.")
	}
	if err := service.store.DecidePlan(ctx, &plan, decision, audit); err != nil {
		return workflowHTTPToolActionStoreFailure(err)
	}
	return WorkflowHTTPToolActionResult{ActionPlan: workflowHTTPToolActionPlanPointer(plan), ConfirmationDecision: &decision}
}

func (service workflowHTTPToolActionService) readExactEligibleDraft(
	ctx WorkflowHTTPToolActionContext,
	draftOwnerRef string,
	draftID string,
	draftVersion int,
	nodeID string,
) (SavedWorkflowDraft, WorkflowHTTPToolActionResult) {
	draftReadScopes := cloneStringSlice(ctx.ScopeGrants)
	if !controlPlaneReadHasScope(draftReadScopes, "workflow_drafts:read") {
		draftReadScopes = append(draftReadScopes, "workflow_drafts:read")
	}
	result := service.readDraft(SavedWorkflowDraftContext{
		RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef,
		WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef,
		OwnerSubjectRef: strings.TrimSpace(draftOwnerRef), ScopeGrants: draftReadScopes, AuditRef: ctx.AuditRef,
	}, ReadWorkflowDraftRequest{DraftID: strings.TrimSpace(draftID)})
	if result.FailureCode != "" || result.Draft == nil {
		return SavedWorkflowDraft{}, workflowHTTPToolSavedDraftReadFailure(result.FailureCode)
	}
	if result.Draft.DraftVersion != draftVersion {
		return SavedWorkflowDraft{}, workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureDraftVersion, "Saved workflow draft version does not match the requested version.")
	}
	if err := validateWorkflowHTTPToolDraft(*result.Draft, strings.TrimSpace(nodeID), service.registry.definition); err != nil {
		return SavedWorkflowDraft{}, workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureDraftIneligible, err.Error())
	}
	return *result.Draft, WorkflowHTTPToolActionResult{}
}

func (service workflowHTTPToolActionService) planSourcesMatch(
	ctx WorkflowHTTPToolActionContext,
	plan WorkflowHTTPToolActionPlan,
) (bool, WorkflowHTTPToolActionResult) {
	if service.registry.definition.ToolID != plan.ToolID || service.registry.definition.ToolVersion != plan.ToolVersion ||
		service.registry.definitionDigest != plan.DefinitionDigest || service.registry.profileDigest != plan.ProfileDigest ||
		service.registry.outputSchemaDigest != plan.OutputSchemaDigest || !service.registry.profile.Enabled {
		return false, WorkflowHTTPToolActionResult{}
	}
	_, result := service.readExactEligibleDraft(ctx, plan.PlannedByActorRef, plan.DraftID, plan.DraftVersion, plan.NodeID)
	if result.FailureCode == "" {
		return true, WorkflowHTTPToolActionResult{}
	}
	switch result.FailureCode {
	case WorkflowHTTPToolActionFailureDraftNotFound, WorkflowHTTPToolActionFailureDraftVersion, WorkflowHTTPToolActionFailureDraftIneligible:
		return false, WorkflowHTTPToolActionResult{}
	default:
		return false, result
	}
}

func workflowHTTPToolSavedDraftReadFailure(code SavedWorkflowDraftFailureCode) WorkflowHTTPToolActionResult {
	switch code {
	case "":
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureStoreContract, "Saved workflow draft source returned an empty contract.")
	case SavedWorkflowDraftFailureNotFound:
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureDraftNotFound, "Saved workflow draft was not found.")
	case SavedWorkflowDraftFailureScopeDenied, SavedWorkflowDraftFailureAuthContextMismatch,
		SavedWorkflowDraftFailureIdentityContextMissing, SavedWorkflowDraftFailureTenantBindingMissing,
		SavedWorkflowDraftFailureWorkspaceMembershipDenied, SavedWorkflowDraftFailureApplicationScopeDenied,
		SavedWorkflowDraftFailureOwnerScopeDenied, SavedWorkflowDraftFailureScopeGrantMissing,
		SavedWorkflowDraftFailureAuditContextMissing:
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureScopeDenied, "Saved workflow draft source scope is unavailable.")
	case SavedWorkflowDraftFailureStoreUnavailable, SavedWorkflowDraftFailureStoreMigrationUnavailable,
		SavedWorkflowDraftFailureRepositoryStoreDisabled, SavedWorkflowDraftFailureWriteDisabled:
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureStoreUnavailable, "Saved workflow draft source store is unavailable.")
	case SavedWorkflowDraftFailureStoreContractMismatch, SavedWorkflowDraftFailureStoreSchemaVersionMismatch,
		SavedWorkflowDraftFailureSchemaMigrationNotApplied, SavedWorkflowDraftFailureInvalidStoreMode:
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureStoreContract, "Saved workflow draft source contract is incompatible.")
	case SavedWorkflowDraftFailureSchemaVersionUnsupported, SavedWorkflowDraftFailurePayloadInvalid,
		SavedWorkflowDraftFailureGraphInvalid, SavedWorkflowDraftFailureContractInvalid,
		SavedWorkflowDraftFailureBlockedCapability, SavedWorkflowDraftFailurePayloadTooLarge,
		SavedWorkflowDraftFailureVersionConflict:
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureDraftIneligible, "Saved workflow draft is not eligible for a tool action plan.")
	default:
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureStoreContract, "Saved workflow draft source returned an unknown failure contract.")
	}
}

func (service workflowHTTPToolActionService) newAudit(ctx WorkflowHTTPToolActionContext, plan WorkflowHTTPToolActionPlan, event string, confirmationID *string, actorSource string, occurredAt time.Time) (WorkflowHTTPToolExecutionAudit, error) {
	eventID, err := service.newID("wtae_")
	if err != nil {
		return WorkflowHTTPToolExecutionAudit{}, err
	}
	actorRef := ctx.ActorRef
	if actorSource == "system" {
		actorRef = "system:workflow_tool_action_policy"
	}
	return WorkflowHTTPToolExecutionAudit{
		SchemaVersion: workflowHTTPToolAuditSchema, EventID: eventID, EventKind: event,
		OccurredAt: workflowRunTimestamp(occurredAt), TenantRef: ctx.TenantRef,
		WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, DraftID: plan.DraftID,
		DraftVersion: plan.DraftVersion, NodeID: plan.NodeID, PlanID: plan.PlanID,
		ConfirmationID: confirmationID, ToolID: plan.ToolID, ToolVersion: plan.ToolVersion,
		DefinitionDigest: plan.DefinitionDigest,
		ProfileID:        plan.ProfileID, ProfileDigest: plan.ProfileDigest, ToolPlanDigest: plan.ToolPlanDigest,
		ActorRef: actorRef, ActorSource: actorSource, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
		AttemptStatus: "not_claimed", SideEffects: WorkflowHTTPToolAuditSideEffects{},
	}, nil
}

func validateWorkflowHTTPToolDraft(draft SavedWorkflowDraft, targetNodeID string, definition WorkflowHTTPToolDefinition) error {
	if draft.SchemaVersion != savedWorkflowDraftSchemaVersion || !workflowHTTPToolDraftFindingsAreEligible(draft.ValidationSummary) {
		return errors.New("Saved workflow draft has blocking findings outside the tool action planning boundary.")
	}
	if len(draft.Nodes) < 4 || len(draft.Edges) != len(draft.Nodes)-1 || !workflowHTTPToolReviewOnlyCapabilities(draft.RequestedCapabilities) {
		return errors.New("Workflow HTTP tool v1 requires one linear prompt-to-output graph.")
	}
	nodes := make(map[string]SavedWorkflowDraftNode, len(draft.Nodes))
	incoming, outgoing := make(map[string]int, len(draft.Nodes)), make(map[string]int, len(draft.Nodes))
	toolCount, promptCount, outputCount, llmCount := 0, 0, 0, 0
	for _, node := range draft.Nodes {
		node.NodeID, node.NodeType = strings.TrimSpace(node.NodeID), strings.ToLower(strings.TrimSpace(node.NodeType))
		if node.NodeID == "" || nodes[node.NodeID].NodeID != "" {
			return errors.New("Workflow HTTP tool graph contains an invalid node id.")
		}
		switch node.NodeType {
		case "prompt":
			promptCount++
		case "llm":
			llmCount++
		case "http_tool":
			toolCount++
			if node.NodeID != targetNodeID || strings.TrimSpace(node.ToolRef) != definition.ToolID || strings.ToLower(strings.TrimSpace(node.RiskLevel)) != definition.RiskLevel || !node.RequiresConfirmation {
				return errors.New("Workflow HTTP tool node does not match the registered definition and confirmation policy.")
			}
		default:
			if node.NodeType != "output" {
				return errors.New("Workflow HTTP tool v1 does not allow condition, RAG, or other node types.")
			}
			outputCount++
		}
		if node.NodeType != "http_tool" && (strings.TrimSpace(node.ToolRef) != "" || strings.TrimSpace(node.RAGRef) != "" || node.RequiresConfirmation) {
			return errors.New("Only the registered HTTP tool node may declare a tool or confirmation boundary.")
		}
		nodes[node.NodeID] = node
	}
	if promptCount != 1 || outputCount != 1 || toolCount != 1 || llmCount < 1 || len(draft.ToolRefs) != 1 || strings.TrimSpace(draft.ToolRefs[0]) != definition.ToolID {
		return errors.New("Workflow HTTP tool v1 requires one prompt, one registered tool, one or more LLM nodes, and one output.")
	}
	for _, edge := range draft.Edges {
		from, to := strings.TrimSpace(edge.FromNodeID), strings.TrimSpace(edge.ToNodeID)
		if nodes[from].NodeID == "" || nodes[to].NodeID == "" || from == to {
			return errors.New("Workflow HTTP tool graph contains an invalid edge.")
		}
		outgoing[from]++
		incoming[to]++
	}
	root, terminal := "", ""
	for id, node := range nodes {
		if incoming[id] == 0 {
			if root != "" || node.NodeType != "prompt" {
				return errors.New("Workflow HTTP tool graph must have one prompt root.")
			}
			root = id
		} else if incoming[id] != 1 {
			return errors.New("Workflow HTTP tool graph must be linear.")
		}
		if outgoing[id] == 0 {
			if terminal != "" || node.NodeType != "output" {
				return errors.New("Workflow HTTP tool graph must have one output terminal.")
			}
			terminal = id
		} else if outgoing[id] != 1 {
			return errors.New("Workflow HTTP tool graph must be linear.")
		}
	}
	seen, current, toolSeen := map[string]bool{}, root, false
	for current != "" {
		if seen[current] {
			return errors.New("Workflow HTTP tool graph must be acyclic.")
		}
		seen[current] = true
		node := nodes[current]
		if node.NodeType == "http_tool" {
			toolSeen = true
		}
		if node.NodeType == "llm" && !toolSeen {
			return errors.New("Every LLM node must follow the HTTP tool node.")
		}
		next := ""
		for _, edge := range draft.Edges {
			if strings.TrimSpace(edge.FromNodeID) == current {
				next = strings.TrimSpace(edge.ToNodeID)
				break
			}
		}
		current = next
	}
	if len(seen) != len(nodes) || terminal == "" {
		return errors.New("Every node must be on the prompt-to-output path.")
	}
	return nil
}

func workflowHTTPToolDraftFindingsAreEligible(summary SavedWorkflowDraftValidationSummary) bool {
	for _, finding := range summary.Findings {
		if finding.Severity != SavedWorkflowDraftValidationBlocking {
			continue
		}
		if finding.Code != SavedWorkflowDraftFailureBlockedCapability || finding.Field != "requested_capabilities" ||
			!workflowHTTPToolReviewOnlyCapabilities([]string{finding.EvidenceID}) {
			return false
		}
	}
	return true
}

func workflowHTTPToolReviewOnlyCapabilities(capabilities []string) bool {
	allowed := map[string]bool{
		"publish": true, "run": true, "confirmation_decision": true, "writeback": true, "replay": true,
	}
	for _, capability := range capabilities {
		if !allowed[strings.TrimSpace(capability)] {
			return false
		}
	}
	return true
}

func normalizeWorkflowHTTPToolPublicArguments(raw map[string]any) (WorkflowHTTPToolPublicArguments, WorkflowHTTPToolActionFailureCode) {
	if len(raw) < 1 || len(raw) > 2 {
		return WorkflowHTTPToolPublicArguments{}, WorkflowHTTPToolActionFailureInputInvalid
	}
	for key := range raw {
		if key != "resource_key" && key != "locale" {
			return WorkflowHTTPToolPublicArguments{}, WorkflowHTTPToolActionFailureInputInvalid
		}
	}
	resourceKey, ok := raw["resource_key"].(string)
	resourceKey = strings.TrimSpace(resourceKey)
	if !ok || len(resourceKey) < 1 || len(resourceKey) > workflowHTTPToolMaxResourceKeySize || strings.Contains(resourceKey, "://") || !workflowHTTPToolResourceKeyPattern.MatchString(resourceKey) {
		return WorkflowHTTPToolPublicArguments{}, WorkflowHTTPToolActionFailureInputInvalid
	}
	locale := ""
	if value, found := raw["locale"]; found {
		var localeOK bool
		locale, localeOK = value.(string)
		locale = strings.TrimSpace(locale)
		if !localeOK || len(locale) < 2 || len(locale) > workflowHTTPToolMaxLocaleSize || !workflowHTTPToolLocalePattern.MatchString(locale) {
			return WorkflowHTTPToolPublicArguments{}, WorkflowHTTPToolActionFailureInputInvalid
		}
	}
	return WorkflowHTTPToolPublicArguments{ResourceKey: resourceKey, Locale: locale}, ""
}

func workflowHTTPToolPlanDigest(plan WorkflowHTTPToolActionPlan) (string, error) {
	payload := struct {
		SchemaVersion      string                          `json:"schema_version"`
		TenantRef          string                          `json:"tenant_ref"`
		WorkspaceID        string                          `json:"workspace_id"`
		ApplicationID      string                          `json:"application_id"`
		DraftID            string                          `json:"draft_id"`
		DraftVersion       int                             `json:"draft_version"`
		NodeID             string                          `json:"node_id"`
		ToolID             string                          `json:"tool_id"`
		ToolVersion        int                             `json:"tool_version"`
		DefinitionDigest   string                          `json:"definition_digest"`
		ProfileID          string                          `json:"profile_id"`
		ProfileVersion     int                             `json:"profile_version"`
		ProfileDigest      string                          `json:"profile_digest"`
		Method             string                          `json:"method"`
		TargetPolicyKey    string                          `json:"target_policy_key"`
		PublicArguments    WorkflowHTTPToolPublicArguments `json:"public_arguments"`
		OutputFields       []string                        `json:"output_fields"`
		OutputSchemaDigest string                          `json:"output_schema_digest"`
		CredentialPolicy   string                          `json:"credential_policy"`
		TimeoutMS          int                             `json:"timeout_ms"`
		MaxResponseBytes   int                             `json:"max_response_bytes"`
		MaxOutputBytes     int                             `json:"max_output_bytes"`
		PlannedByActorRef  string                          `json:"planned_by_actor_ref"`
		CreatedAt          string                          `json:"created_at"`
		ExpiresAt          string                          `json:"expires_at"`
	}{plan.SchemaVersion, plan.TenantRef, plan.WorkspaceID, plan.ApplicationID, plan.DraftID, plan.DraftVersion, plan.NodeID, plan.ToolID, plan.ToolVersion, plan.DefinitionDigest, plan.ProfileID, plan.ProfileVersion, plan.ProfileDigest, plan.Method, plan.TargetPolicyKey, plan.PublicArguments, append([]string(nil), plan.OutputFields...), plan.OutputSchemaDigest, plan.CredentialPolicy, plan.TimeoutMS, plan.MaxResponseBytes, plan.MaxOutputBytes, plan.PlannedByActorRef, plan.CreatedAt, plan.ExpiresAt}
	return canonicalSHA256(payload)
}

func canonicalSHA256(value any) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	var normalized any
	if err := decoder.Decode(&normalized); err != nil {
		return "", err
	}
	var canonical bytes.Buffer
	if err := writeRestrictedJCS(&canonical, normalized); err != nil {
		return "", err
	}
	digest := sha256.Sum256(canonical.Bytes())
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

// writeRestrictedJCS implements RFC 8785 ordering and scalar serialization for
// this contract's deliberately restricted JSON domain: ASCII strings, booleans,
// null, integral numbers, arrays, and objects. Floating point inputs are rejected.
func writeRestrictedJCS(output *bytes.Buffer, value any) error {
	switch typed := value.(type) {
	case nil:
		output.WriteString("null")
	case bool:
		if typed {
			output.WriteString("true")
		} else {
			output.WriteString("false")
		}
	case string:
		for _, character := range typed {
			if character < 0x20 || character > 0x7e || character == '<' || character == '>' || character == '&' {
				return errors.New("restricted JCS input contains a string outside the canonical ASCII domain")
			}
		}
		encoded, err := json.Marshal(typed)
		if err != nil {
			return err
		}
		output.Write(encoded)
	case json.Number:
		integer, err := strconv.ParseInt(string(typed), 10, 64)
		if err != nil {
			return errors.New("restricted JCS input contains a non-integral number")
		}
		output.WriteString(strconv.FormatInt(integer, 10))
	case []any:
		output.WriteByte('[')
		for index, item := range typed {
			if index > 0 {
				output.WriteByte(',')
			}
			if err := writeRestrictedJCS(output, item); err != nil {
				return err
			}
		}
		output.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		output.WriteByte('{')
		for index, key := range keys {
			if index > 0 {
				output.WriteByte(',')
			}
			if err := writeRestrictedJCS(output, key); err != nil {
				return err
			}
			output.WriteByte(':')
			if err := writeRestrictedJCS(output, typed[key]); err != nil {
				return err
			}
		}
		output.WriteByte('}')
	default:
		return errors.New("restricted JCS input contains an unsupported JSON value")
	}
	return nil
}

func newWorkflowHTTPToolActionID(prefix string) (string, error) {
	randomBytes := make([]byte, 12)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(randomBytes), nil
}

func validateWorkflowHTTPToolActionContext(ctx WorkflowHTTPToolActionContext) WorkflowHTTPToolActionFailureCode {
	if ctx.RequestContext == nil || !workflowHTTPToolReferencePattern.MatchString(strings.TrimSpace(ctx.TenantRef)) ||
		!workflowHTTPToolScopedIDPattern.MatchString(strings.TrimSpace(ctx.WorkspaceID)) ||
		!workflowHTTPToolScopedIDPattern.MatchString(strings.TrimSpace(ctx.ApplicationID)) ||
		!workflowHTTPToolReferencePattern.MatchString(strings.TrimSpace(ctx.ActorRef)) ||
		!workflowHTTPToolReferencePattern.MatchString(strings.TrimSpace(ctx.RequestID)) ||
		!workflowHTTPToolReferencePattern.MatchString(strings.TrimSpace(ctx.AuditRef)) {
		return WorkflowHTTPToolActionFailureScopeDenied
	}
	return ""
}

func workflowHTTPToolHumanDecisionAllowed(outcome WorkflowHTTPToolConfirmationOutcome) bool {
	switch outcome {
	case WorkflowHTTPToolConfirmationApprove, WorkflowHTTPToolConfirmationReject, WorkflowHTTPToolConfirmationDefer, WorkflowHTTPToolConfirmationCancel:
		return true
	default:
		return false
	}
}

func workflowHTTPToolDecisionAllowedFrom(status WorkflowHTTPToolActionStatus, outcome WorkflowHTTPToolConfirmationOutcome) bool {
	if outcome == workflowHTTPToolConfirmationExpire || outcome == workflowHTTPToolConfirmationInvalidate {
		return status == WorkflowHTTPToolActionStatusPending || status == WorkflowHTTPToolActionStatusDeferred ||
			status == WorkflowHTTPToolActionStatusApproved
	}
	if status == WorkflowHTTPToolActionStatusApproved {
		return outcome == WorkflowHTTPToolConfirmationCancel
	}
	if status == WorkflowHTTPToolActionStatusDeferred {
		return outcome == WorkflowHTTPToolConfirmationApprove || outcome == WorkflowHTTPToolConfirmationReject ||
			outcome == WorkflowHTTPToolConfirmationCancel
	}
	return status == WorkflowHTTPToolActionStatusPending && workflowHTTPToolHumanDecisionAllowed(outcome)
}

func workflowHTTPToolStatusForOutcome(outcome WorkflowHTTPToolConfirmationOutcome) WorkflowHTTPToolActionStatus {
	switch outcome {
	case WorkflowHTTPToolConfirmationApprove:
		return WorkflowHTTPToolActionStatusApproved
	case WorkflowHTTPToolConfirmationReject:
		return WorkflowHTTPToolActionStatusRejected
	case WorkflowHTTPToolConfirmationDefer:
		return WorkflowHTTPToolActionStatusDeferred
	case WorkflowHTTPToolConfirmationCancel:
		return WorkflowHTTPToolActionStatusCanceled
	case workflowHTTPToolConfirmationExpire:
		return WorkflowHTTPToolActionStatusExpired
	case workflowHTTPToolConfirmationInvalidate:
		return WorkflowHTTPToolActionStatusInvalidated
	default:
		return ""
	}
}

func workflowHTTPToolAuditEventForOutcome(outcome WorkflowHTTPToolConfirmationOutcome) string {
	switch outcome {
	case WorkflowHTTPToolConfirmationApprove:
		return "confirmation_recorded"
	case WorkflowHTTPToolConfirmationReject:
		return "confirmation_rejected"
	case WorkflowHTTPToolConfirmationDefer:
		return "confirmation_deferred"
	case WorkflowHTTPToolConfirmationCancel:
		return "confirmation_canceled"
	case workflowHTTPToolConfirmationExpire:
		return "confirmation_expired"
	case workflowHTTPToolConfirmationInvalidate:
		return "confirmation_invalidated"
	default:
		return ""
	}
}

func workflowHTTPToolConfirmationReason(outcome WorkflowHTTPToolConfirmationOutcome) string {
	return "workflow_tool_confirmation_" + string(outcome)
}

func workflowHTTPToolTerminalStatus(status WorkflowHTTPToolActionStatus) bool {
	switch status {
	case WorkflowHTTPToolActionStatusRejected, WorkflowHTTPToolActionStatusCanceled, WorkflowHTTPToolActionStatusExpired, WorkflowHTTPToolActionStatusInvalidated, WorkflowHTTPToolActionStatusConsumed:
		return true
	default:
		return false
	}
}

func workflowHTTPToolActionStoreFailure(err error) WorkflowHTTPToolActionResult {
	switch {
	case errors.Is(err, errWorkflowHTTPToolActionConflict):
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureConflict, "Workflow HTTP tool action plan changed concurrently.")
	case errors.Is(err, errWorkflowHTTPToolActionContract):
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureStoreContract, "Workflow HTTP tool action store contract is incompatible.")
	default:
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureStoreUnavailable, "Workflow HTTP tool action store is unavailable.")
	}
}

func workflowHTTPToolActionTerminalFailure(status WorkflowHTTPToolActionStatus) WorkflowHTTPToolActionResult {
	switch status {
	case WorkflowHTTPToolActionStatusRejected:
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureRejected, "Workflow HTTP tool action plan was rejected.")
	case WorkflowHTTPToolActionStatusCanceled:
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureCanceled, "Workflow HTTP tool action plan was canceled.")
	case WorkflowHTTPToolActionStatusExpired:
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureExpired, "Workflow HTTP tool action plan expired.")
	case WorkflowHTTPToolActionStatusInvalidated:
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureInvalidated, "Workflow HTTP tool action plan was invalidated.")
	case WorkflowHTTPToolActionStatusConsumed:
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureConsumed, "Workflow HTTP tool action plan was already consumed.")
	default:
		return workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureConflict, "Workflow HTTP tool action plan does not accept this decision in its current state.")
	}
}

func workflowHTTPToolActionFailure(code WorkflowHTTPToolActionFailureCode, summary string) WorkflowHTTPToolActionResult {
	return WorkflowHTTPToolActionResult{FailureCode: code, FailureSummary: summary}
}

func workflowHTTPToolActionPlanPointer(plan WorkflowHTTPToolActionPlan) *WorkflowHTTPToolActionPlan {
	cloned := cloneWorkflowHTTPToolActionPlan(plan)
	return &cloned
}
func workflowHTTPToolActionPlanValue(plan *WorkflowHTTPToolActionPlan, fallback WorkflowHTTPToolActionPlan) WorkflowHTTPToolActionPlan {
	if plan == nil {
		return fallback
	}
	return cloneWorkflowHTTPToolActionPlan(*plan)
}
func cloneWorkflowHTTPToolActionPlan(plan WorkflowHTTPToolActionPlan) WorkflowHTTPToolActionPlan {
	plan.OutputFields = append([]string(nil), plan.OutputFields...)
	if plan.LastDecisionByActorRef != nil {
		value := *plan.LastDecisionByActorRef
		plan.LastDecisionByActorRef = &value
	}
	if plan.LastDecisionAt != nil {
		value := *plan.LastDecisionAt
		plan.LastDecisionAt = &value
	}
	return plan
}

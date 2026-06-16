package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
)

const (
	savedWorkflowDraftSaveRoute     = "POST /v1/user-workspace/workflow-drafts"
	savedWorkflowDraftListRoute     = "GET /v1/user-workspace/workflow-drafts"
	savedWorkflowDraftReadRoute     = "GET /v1/user-workspace/workflow-drafts/{draft_id}"
	savedWorkflowDraftValidateRoute = "POST /v1/user-workspace/workflow-drafts/validate"
)

const (
	savedWorkflowDraftDevWorkspaceHeader   = "X-RadishMind-Dev-Workflow-Workspace"
	savedWorkflowDraftDevApplicationHeader = "X-RadishMind-Dev-Workflow-Application"
)

type savedWorkflowDraftSaveHTTPBody struct {
	ExpectedDraftVersion int                               `json:"expected_draft_version"`
	Draft                savedWorkflowDraftPayloadDocument `json:"draft"`
}

type savedWorkflowDraftValidateHTTPBody struct {
	Draft savedWorkflowDraftPayloadDocument `json:"draft"`
}

type savedWorkflowDraftEnvelope struct {
	RequestID           string                               `json:"request_id"`
	WorkspaceID         string                               `json:"workspace_id"`
	ApplicationID       string                               `json:"application_id"`
	Draft               *savedWorkflowDraftDocument          `json:"draft"`
	FailureCode         *string                              `json:"failure_code"`
	CurrentDraftVersion int                                  `json:"current_draft_version"`
	ValidationSummary   savedWorkflowDraftValidationDocument `json:"validation_summary"`
	BlockedCapabilities []savedWorkflowDraftBlockedDocument  `json:"blocked_capabilities"`
	AuditRef            string                               `json:"audit_ref"`
}

type savedWorkflowDraftListEnvelope struct {
	RequestID      string                              `json:"request_id"`
	WorkspaceID    string                              `json:"workspace_id"`
	ApplicationID  string                              `json:"application_id"`
	DraftSummaries []savedWorkflowDraftSummaryDocument `json:"draft_summaries"`
	FailureCode    *string                             `json:"failure_code"`
	AuditRef       string                              `json:"audit_ref"`
}

type savedWorkflowDraftPayloadDocument struct {
	DraftID               string                             `json:"draft_id"`
	WorkspaceID           string                             `json:"workspace_id"`
	ApplicationID         string                             `json:"application_id"`
	SourceDefinitionID    string                             `json:"source_definition_id"`
	BaseDefinitionVersion int                                `json:"base_definition_version"`
	SchemaVersion         string                             `json:"schema_version"`
	DraftStatus           string                             `json:"draft_status,omitempty"`
	Name                  string                             `json:"name"`
	Description           string                             `json:"description"`
	Nodes                 []savedWorkflowDraftNodeDocument   `json:"nodes"`
	Edges                 []savedWorkflowDraftEdgeDocument   `json:"edges"`
	InputContract         savedWorkflowDraftContractDocument `json:"input_contract"`
	OutputContract        savedWorkflowDraftContractDocument `json:"output_contract"`
	ProviderRefs          []string                           `json:"provider_refs"`
	ToolRefs              []string                           `json:"tool_refs"`
	RAGRefs               []string                           `json:"rag_refs"`
	RequestedCapabilities []string                           `json:"requested_capabilities"`
	AdditionalFields      map[string]any                     `json:"additional_fields,omitempty"`
}

type savedWorkflowDraftDocument struct {
	savedWorkflowDraftPayloadDocument
	DraftVersion               int                                  `json:"draft_version"`
	CreatedAt                  string                               `json:"created_at"`
	UpdatedAt                  string                               `json:"updated_at"`
	CreatedByActorRef          string                               `json:"created_by_actor_ref"`
	UpdatedByActorRef          string                               `json:"updated_by_actor_ref"`
	ValidationSummary          savedWorkflowDraftValidationDocument `json:"validation_summary"`
	BlockedCapabilitySummary   []savedWorkflowDraftBlockedDocument  `json:"blocked_capability_summary"`
	RequestAuditMetadata       savedWorkflowDraftAuditDocument      `json:"request_audit_metadata"`
	SampleOrUnsavedDraftStatus string                               `json:"sample_or_unsaved_draft_status"`
}

type savedWorkflowDraftSummaryDocument struct {
	DraftID                    string `json:"draft_id"`
	WorkspaceID                string `json:"workspace_id"`
	ApplicationID              string `json:"application_id"`
	SourceDefinitionID         string `json:"source_definition_id"`
	DraftVersion               int    `json:"draft_version"`
	SchemaVersion              string `json:"schema_version"`
	DraftStatus                string `json:"draft_status"`
	Name                       string `json:"name"`
	Description                string `json:"description"`
	UpdatedAt                  string `json:"updated_at"`
	UpdatedByActorRef          string `json:"updated_by_actor_ref"`
	NodeCount                  int    `json:"node_count"`
	EdgeCount                  int    `json:"edge_count"`
	BlockedCapabilityCount     int    `json:"blocked_capability_count"`
	ValidationState            string `json:"validation_state"`
	ValidForReview             bool   `json:"valid_for_review"`
	SampleOrUnsavedDraftStatus string `json:"sample_or_unsaved_draft_status"`
}

type savedWorkflowDraftNodeDocument struct {
	NodeID               string `json:"node_id"`
	NodeType             string `json:"node_type"`
	Label                string `json:"label"`
	InputContractRef     string `json:"input_contract_ref"`
	OutputContractRef    string `json:"output_contract_ref"`
	ProviderRef          string `json:"provider_ref"`
	ToolRef              string `json:"tool_ref"`
	RAGRef               string `json:"rag_ref"`
	RiskLevel            string `json:"risk_level"`
	RequiresConfirmation bool   `json:"requires_confirmation"`
}

type savedWorkflowDraftEdgeDocument struct {
	EdgeID           string `json:"edge_id"`
	FromNodeID       string `json:"from_node_id"`
	ToNodeID         string `json:"to_node_id"`
	ConditionSummary string `json:"condition_summary"`
}

type savedWorkflowDraftContractDocument struct {
	ContractID     string   `json:"contract_id"`
	RequiredFields []string `json:"required_fields"`
	Summary        string   `json:"summary"`
}

type savedWorkflowDraftValidationDocument struct {
	ValidationState string                              `json:"validation_state"`
	ValidForReview  bool                                `json:"valid_for_review"`
	Findings        []savedWorkflowDraftFindingDocument `json:"findings"`
}

type savedWorkflowDraftFindingDocument struct {
	Code       string `json:"code"`
	Severity   string `json:"severity"`
	Field      string `json:"field"`
	Summary    string `json:"summary"`
	EvidenceID string `json:"evidence_id"`
}

type savedWorkflowDraftBlockedDocument struct {
	CapabilityID        string `json:"capability_id"`
	MissingPrerequisite string `json:"missing_prerequisite"`
	Summary             string `json:"summary"`
}

type savedWorkflowDraftAuditDocument struct {
	RequestID string `json:"request_id"`
	AuditRef  string `json:"audit_ref"`
	ActorRef  string `json:"actor_ref"`
}

func (s *Server) handleSaveWorkflowDraft(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, savedWorkflowDraftSaveRoute)
	if !s.allowSavedWorkflowDraftDevHTTP(writer, request, trace) {
		return
	}
	var body savedWorkflowDraftSaveHTTPBody
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		s.writePlatformError(writer, trace, "INVALID_JSON", err.Error())
		return
	}
	payload := savedWorkflowDraftPayloadFromDocument(body.Draft)
	context, failureCode := savedWorkflowDraftContextFromRequest(
		request,
		trace,
		payload.WorkspaceID,
		payload.ApplicationID,
		"workflow_drafts:write",
		s.config.WorkflowSavedDraftDevWriteEnabled,
		"save",
	)
	if failureCode != "" {
		writeSavedWorkflowDraftResult(writer, trace, context, savedWorkflowDraftFailure(failureCode, savedWorkflowDraftAuditMetadata(context)))
		return
	}
	result := s.savedWorkflowDraftService().SaveDraft(context, SaveWorkflowDraftRequest{
		ExpectedDraftVersion: body.ExpectedDraftVersion,
		Payload:              payload,
	})
	writeSavedWorkflowDraftResult(writer, trace, context, result)
}

func (s *Server) handleReadWorkflowDraft(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, savedWorkflowDraftReadRoute)
	if !s.allowSavedWorkflowDraftDevHTTP(writer, request, trace) {
		return
	}
	workspaceID := strings.TrimSpace(request.URL.Query().Get("workspace_id"))
	applicationID := strings.TrimSpace(request.URL.Query().Get("application_id"))
	context, failureCode := savedWorkflowDraftContextFromRequest(
		request,
		trace,
		workspaceID,
		applicationID,
		"workflow_drafts:read",
		false,
		"read",
	)
	if failureCode != "" {
		writeSavedWorkflowDraftResult(writer, trace, context, savedWorkflowDraftFailure(failureCode, savedWorkflowDraftAuditMetadata(context)))
		return
	}
	result := s.savedWorkflowDraftService().ReadDraft(context, ReadWorkflowDraftRequest{
		DraftID: strings.TrimSpace(request.PathValue("draft_id")),
	})
	writeSavedWorkflowDraftResult(writer, trace, context, result)
}

func (s *Server) handleListWorkflowDrafts(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, savedWorkflowDraftListRoute)
	if !s.allowSavedWorkflowDraftDevHTTP(writer, request, trace) {
		return
	}
	workspaceID := strings.TrimSpace(request.URL.Query().Get("workspace_id"))
	applicationID := strings.TrimSpace(request.URL.Query().Get("application_id"))
	context, failureCode := savedWorkflowDraftContextFromRequest(
		request,
		trace,
		workspaceID,
		applicationID,
		"workflow_drafts:read",
		false,
		"list",
	)
	if failureCode != "" {
		writeSavedWorkflowDraftListResult(
			writer,
			trace,
			context,
			savedWorkflowDraftListFailure(failureCode, savedWorkflowDraftAuditMetadata(context)),
		)
		return
	}
	result := s.savedWorkflowDraftService().ListDrafts(context, ListWorkflowDraftsRequest{})
	writeSavedWorkflowDraftListResult(writer, trace, context, result)
}

func (s *Server) handleValidateWorkflowDraft(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, savedWorkflowDraftValidateRoute)
	if !s.allowSavedWorkflowDraftDevHTTP(writer, request, trace) {
		return
	}
	var body savedWorkflowDraftValidateHTTPBody
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		s.writePlatformError(writer, trace, "INVALID_JSON", err.Error())
		return
	}
	payload := savedWorkflowDraftPayloadFromDocument(body.Draft)
	context, failureCode := savedWorkflowDraftContextFromRequest(
		request,
		trace,
		payload.WorkspaceID,
		payload.ApplicationID,
		"workflow_drafts:write",
		false,
		"validate",
	)
	if failureCode != "" {
		writeSavedWorkflowDraftResult(writer, trace, context, savedWorkflowDraftFailure(failureCode, savedWorkflowDraftAuditMetadata(context)))
		return
	}
	result := s.savedWorkflowDraftService().ValidateDraft(context, ValidateWorkflowDraftRequest{Payload: payload})
	writeSavedWorkflowDraftResult(writer, trace, context, result)
}

func (s *Server) allowSavedWorkflowDraftDevHTTP(writer http.ResponseWriter, request *http.Request, trace requestTrace) bool {
	if s.config.WorkflowSavedDraftDevHTTPEnabled {
		return true
	}
	s.writePlatformError(writer, trace, "SAVED_WORKFLOW_DRAFT_DEV_HTTP_DISABLED", "saved workflow draft dev route requires explicit opt-in")
	return false
}

func (s *Server) savedWorkflowDraftService() savedWorkflowDraftService {
	if s.savedWorkflowDraftStore == nil {
		s.savedWorkflowDraftStore = newMemorySavedWorkflowDraftStore()
	}
	return newSavedWorkflowDraftService(s.savedWorkflowDraftStore)
}

func savedWorkflowDraftContextFromRequest(
	request *http.Request,
	trace requestTrace,
	workspaceID string,
	applicationID string,
	requiredScope string,
	writeEnabled bool,
	auditSuffix string,
) (SavedWorkflowDraftContext, SavedWorkflowDraftFailureCode) {
	auth, ok := request.Context().Value(controlPlaneReadAuthContextKey{}).(controlPlaneReadAuthContext)
	context := SavedWorkflowDraftContext{
		RequestID:     trace.requestID,
		WorkspaceID:   strings.TrimSpace(workspaceID),
		ApplicationID: strings.TrimSpace(applicationID),
		AuditRef:      auditRefForSavedWorkflowDraft(trace, auditSuffix),
		WriteEnabled:  writeEnabled,
	}
	if !ok || strings.TrimSpace(auth.IdentityContext) == "" || strings.TrimSpace(auth.SubjectBinding) == "" {
		return context, SavedWorkflowDraftFailureScopeDenied
	}
	context.ActorRef = strings.TrimSpace(auth.SubjectBinding)
	if !controlPlaneReadHasScope(auth.ScopeGrants, requiredScope) {
		return context, SavedWorkflowDraftFailureScopeDenied
	}
	headerWorkspaceID := strings.TrimSpace(request.Header.Get(savedWorkflowDraftDevWorkspaceHeader))
	headerApplicationID := strings.TrimSpace(request.Header.Get(savedWorkflowDraftDevApplicationHeader))
	if headerWorkspaceID == "" || headerApplicationID == "" {
		return context, SavedWorkflowDraftFailureScopeDenied
	}
	if context.WorkspaceID == "" || context.ApplicationID == "" ||
		headerWorkspaceID != context.WorkspaceID ||
		headerApplicationID != context.ApplicationID {
		return context, SavedWorkflowDraftFailureScopeDenied
	}
	return context, ""
}

func writeSavedWorkflowDraftResult(
	writer http.ResponseWriter,
	trace requestTrace,
	context SavedWorkflowDraftContext,
	result SavedWorkflowDraftResult,
) {
	failureCode := savedWorkflowDraftFailureCodePointer(result.FailureCode)
	writeObservedJSON(writer, http.StatusOK, trace, savedWorkflowDraftEnvelope{
		RequestID:           trace.requestID,
		WorkspaceID:         strings.TrimSpace(context.WorkspaceID),
		ApplicationID:       strings.TrimSpace(context.ApplicationID),
		Draft:               savedWorkflowDraftDocumentPointer(result.Draft),
		FailureCode:         failureCode,
		CurrentDraftVersion: result.CurrentDraftVersion,
		ValidationSummary:   savedWorkflowDraftValidationToDocument(result.ValidationSummary),
		BlockedCapabilities: savedWorkflowDraftBlockedToDocuments(result.BlockedCapabilities),
		AuditRef:            strings.TrimSpace(result.RequestAuditMetadata.AuditRef),
	})
}

func writeSavedWorkflowDraftListResult(
	writer http.ResponseWriter,
	trace requestTrace,
	context SavedWorkflowDraftContext,
	result SavedWorkflowDraftListResult,
) {
	failureCode := savedWorkflowDraftFailureCodePointer(result.FailureCode)
	writeObservedJSON(writer, http.StatusOK, trace, savedWorkflowDraftListEnvelope{
		RequestID:      trace.requestID,
		WorkspaceID:    strings.TrimSpace(context.WorkspaceID),
		ApplicationID:  strings.TrimSpace(context.ApplicationID),
		DraftSummaries: savedWorkflowDraftSummariesToDocuments(result.Summaries),
		FailureCode:    failureCode,
		AuditRef:       strings.TrimSpace(result.RequestAuditMetadata.AuditRef),
	})
}

func savedWorkflowDraftPayloadFromDocument(document savedWorkflowDraftPayloadDocument) SavedWorkflowDraftPayload {
	return SavedWorkflowDraftPayload{
		DraftID:               document.DraftID,
		WorkspaceID:           document.WorkspaceID,
		ApplicationID:         document.ApplicationID,
		SourceDefinitionID:    document.SourceDefinitionID,
		BaseDefinitionVersion: document.BaseDefinitionVersion,
		SchemaVersion:         document.SchemaVersion,
		DraftStatus:           SavedWorkflowDraftStatus(document.DraftStatus),
		Name:                  document.Name,
		Description:           document.Description,
		Nodes:                 savedWorkflowDraftNodesFromDocuments(document.Nodes),
		Edges:                 savedWorkflowDraftEdgesFromDocuments(document.Edges),
		InputContract:         savedWorkflowDraftContractFromDocument(document.InputContract),
		OutputContract:        savedWorkflowDraftContractFromDocument(document.OutputContract),
		ProviderRefs:          cloneStringSlice(document.ProviderRefs),
		ToolRefs:              cloneStringSlice(document.ToolRefs),
		RAGRefs:               cloneStringSlice(document.RAGRefs),
		RequestedCapabilities: cloneStringSlice(document.RequestedCapabilities),
		AdditionalFields:      document.AdditionalFields,
	}
}

func savedWorkflowDraftDocumentPointer(draft *SavedWorkflowDraft) *savedWorkflowDraftDocument {
	if draft == nil {
		return nil
	}
	payload := savedWorkflowDraftPayloadDocumentFromDraftPayload(savedWorkflowDraftPayloadFromDraft(*draft))
	return &savedWorkflowDraftDocument{
		savedWorkflowDraftPayloadDocument: payload,
		DraftVersion:                      draft.DraftVersion,
		CreatedAt:                         draft.CreatedAt,
		UpdatedAt:                         draft.UpdatedAt,
		CreatedByActorRef:                 draft.CreatedByActorRef,
		UpdatedByActorRef:                 draft.UpdatedByActorRef,
		ValidationSummary:                 savedWorkflowDraftValidationToDocument(draft.ValidationSummary),
		BlockedCapabilitySummary:          savedWorkflowDraftBlockedToDocuments(draft.BlockedCapabilitySummary),
		RequestAuditMetadata:              savedWorkflowDraftAuditToDocument(draft.RequestAuditMetadata),
		SampleOrUnsavedDraftStatus:        draft.SampleOrUnsavedDraftStatus,
	}
}

func savedWorkflowDraftSummariesToDocuments(
	summaries []SavedWorkflowDraftSummary,
) []savedWorkflowDraftSummaryDocument {
	documents := make([]savedWorkflowDraftSummaryDocument, 0, len(summaries))
	for _, summary := range summaries {
		documents = append(documents, savedWorkflowDraftSummaryDocument{
			DraftID:                    summary.DraftID,
			WorkspaceID:                summary.WorkspaceID,
			ApplicationID:              summary.ApplicationID,
			SourceDefinitionID:         summary.SourceDefinitionID,
			DraftVersion:               summary.DraftVersion,
			SchemaVersion:              summary.SchemaVersion,
			DraftStatus:                string(summary.DraftStatus),
			Name:                       summary.Name,
			Description:                summary.Description,
			UpdatedAt:                  summary.UpdatedAt,
			UpdatedByActorRef:          summary.UpdatedByActorRef,
			NodeCount:                  summary.NodeCount,
			EdgeCount:                  summary.EdgeCount,
			BlockedCapabilityCount:     summary.BlockedCapabilityCount,
			ValidationState:            string(summary.ValidationState),
			ValidForReview:             summary.ValidForReview,
			SampleOrUnsavedDraftStatus: summary.SampleOrUnsavedDraftStatus,
		})
	}
	return documents
}

func savedWorkflowDraftPayloadDocumentFromDraftPayload(payload SavedWorkflowDraftPayload) savedWorkflowDraftPayloadDocument {
	return savedWorkflowDraftPayloadDocument{
		DraftID:               payload.DraftID,
		WorkspaceID:           payload.WorkspaceID,
		ApplicationID:         payload.ApplicationID,
		SourceDefinitionID:    payload.SourceDefinitionID,
		BaseDefinitionVersion: payload.BaseDefinitionVersion,
		SchemaVersion:         payload.SchemaVersion,
		DraftStatus:           string(payload.DraftStatus),
		Name:                  payload.Name,
		Description:           payload.Description,
		Nodes:                 savedWorkflowDraftNodesToDocuments(payload.Nodes),
		Edges:                 savedWorkflowDraftEdgesToDocuments(payload.Edges),
		InputContract:         savedWorkflowDraftContractToDocument(payload.InputContract),
		OutputContract:        savedWorkflowDraftContractToDocument(payload.OutputContract),
		ProviderRefs:          cloneStringSlice(payload.ProviderRefs),
		ToolRefs:              cloneStringSlice(payload.ToolRefs),
		RAGRefs:               cloneStringSlice(payload.RAGRefs),
		RequestedCapabilities: cloneStringSlice(payload.RequestedCapabilities),
	}
}

func savedWorkflowDraftNodesFromDocuments(documents []savedWorkflowDraftNodeDocument) []SavedWorkflowDraftNode {
	nodes := make([]SavedWorkflowDraftNode, 0, len(documents))
	for _, document := range documents {
		nodes = append(nodes, SavedWorkflowDraftNode{
			NodeID:               document.NodeID,
			NodeType:             document.NodeType,
			Label:                document.Label,
			InputContractRef:     document.InputContractRef,
			OutputContractRef:    document.OutputContractRef,
			ProviderRef:          document.ProviderRef,
			ToolRef:              document.ToolRef,
			RAGRef:               document.RAGRef,
			RiskLevel:            document.RiskLevel,
			RequiresConfirmation: document.RequiresConfirmation,
		})
	}
	return nodes
}

func savedWorkflowDraftNodesToDocuments(nodes []SavedWorkflowDraftNode) []savedWorkflowDraftNodeDocument {
	documents := make([]savedWorkflowDraftNodeDocument, 0, len(nodes))
	for _, node := range nodes {
		documents = append(documents, savedWorkflowDraftNodeDocument{
			NodeID:               node.NodeID,
			NodeType:             node.NodeType,
			Label:                node.Label,
			InputContractRef:     node.InputContractRef,
			OutputContractRef:    node.OutputContractRef,
			ProviderRef:          node.ProviderRef,
			ToolRef:              node.ToolRef,
			RAGRef:               node.RAGRef,
			RiskLevel:            node.RiskLevel,
			RequiresConfirmation: node.RequiresConfirmation,
		})
	}
	return documents
}

func savedWorkflowDraftEdgesFromDocuments(documents []savedWorkflowDraftEdgeDocument) []SavedWorkflowDraftEdge {
	edges := make([]SavedWorkflowDraftEdge, 0, len(documents))
	for _, document := range documents {
		edges = append(edges, SavedWorkflowDraftEdge{
			EdgeID:           document.EdgeID,
			FromNodeID:       document.FromNodeID,
			ToNodeID:         document.ToNodeID,
			ConditionSummary: document.ConditionSummary,
		})
	}
	return edges
}

func savedWorkflowDraftEdgesToDocuments(edges []SavedWorkflowDraftEdge) []savedWorkflowDraftEdgeDocument {
	documents := make([]savedWorkflowDraftEdgeDocument, 0, len(edges))
	for _, edge := range edges {
		documents = append(documents, savedWorkflowDraftEdgeDocument{
			EdgeID:           edge.EdgeID,
			FromNodeID:       edge.FromNodeID,
			ToNodeID:         edge.ToNodeID,
			ConditionSummary: edge.ConditionSummary,
		})
	}
	return documents
}

func savedWorkflowDraftContractFromDocument(document savedWorkflowDraftContractDocument) SavedWorkflowDraftContract {
	return SavedWorkflowDraftContract{
		ContractID:     document.ContractID,
		RequiredFields: cloneStringSlice(document.RequiredFields),
		Summary:        document.Summary,
	}
}

func savedWorkflowDraftContractToDocument(contract SavedWorkflowDraftContract) savedWorkflowDraftContractDocument {
	return savedWorkflowDraftContractDocument{
		ContractID:     contract.ContractID,
		RequiredFields: cloneStringSlice(contract.RequiredFields),
		Summary:        contract.Summary,
	}
}

func savedWorkflowDraftValidationToDocument(summary SavedWorkflowDraftValidationSummary) savedWorkflowDraftValidationDocument {
	return savedWorkflowDraftValidationDocument{
		ValidationState: string(summary.ValidationState),
		ValidForReview:  summary.ValidForReview,
		Findings:        savedWorkflowDraftFindingsToDocuments(summary.Findings),
	}
}

func savedWorkflowDraftFindingsToDocuments(findings []SavedWorkflowDraftValidationFinding) []savedWorkflowDraftFindingDocument {
	documents := make([]savedWorkflowDraftFindingDocument, 0, len(findings))
	for _, finding := range findings {
		documents = append(documents, savedWorkflowDraftFindingDocument{
			Code:       string(finding.Code),
			Severity:   string(finding.Severity),
			Field:      finding.Field,
			Summary:    finding.Summary,
			EvidenceID: finding.EvidenceID,
		})
	}
	return documents
}

func savedWorkflowDraftBlockedToDocuments(blocked []SavedWorkflowDraftBlockedCapability) []savedWorkflowDraftBlockedDocument {
	documents := make([]savedWorkflowDraftBlockedDocument, 0, len(blocked))
	for _, capability := range blocked {
		documents = append(documents, savedWorkflowDraftBlockedDocument{
			CapabilityID:        capability.CapabilityID,
			MissingPrerequisite: capability.MissingPrerequisite,
			Summary:             capability.Summary,
		})
	}
	return documents
}

func savedWorkflowDraftAuditToDocument(audit SavedWorkflowDraftAuditMetadata) savedWorkflowDraftAuditDocument {
	return savedWorkflowDraftAuditDocument{
		RequestID: audit.RequestID,
		AuditRef:  audit.AuditRef,
		ActorRef:  audit.ActorRef,
	}
}

func savedWorkflowDraftFailureCodePointer(failureCode SavedWorkflowDraftFailureCode) *string {
	if failureCode == "" {
		return nil
	}
	normalized := string(failureCode)
	return &normalized
}

func auditRefForSavedWorkflowDraft(trace requestTrace, suffix string) string {
	return strings.TrimSpace("audit_" + trace.requestID + "_workflow-draft-" + suffix)
}

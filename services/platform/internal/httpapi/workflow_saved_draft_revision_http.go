package httpapi

import (
	"net/http"
	"strconv"
	"strings"
)

const (
	savedWorkflowDraftRevisionListRoute    = "GET /v1/user-workspace/workflow-drafts/{draft_id}/revisions"
	savedWorkflowDraftRevisionReadRoute    = "GET /v1/user-workspace/workflow-drafts/{draft_id}/revisions/{draft_version}"
	savedWorkflowDraftRevisionRestoreRoute = "POST /v1/user-workspace/workflow-drafts/{draft_id}/revisions/{draft_version}/restore"
)

type savedWorkflowDraftRevisionRestoreHTTPBody struct {
	ExpectedCurrentDraftVersion int `json:"expected_current_draft_version"`
	ExpectedLifecycleVersion    int `json:"expected_lifecycle_version"`
}

type savedWorkflowDraftRevisionDocument struct {
	SchemaVersion       string                      `json:"schema_version"`
	Draft               *savedWorkflowDraftDocument `json:"draft"`
	RevisionKind        string                      `json:"revision_kind"`
	RestoredFromVersion int                         `json:"restored_from_version"`
}

type savedWorkflowDraftRevisionSummaryDocument struct {
	SchemaVersion       string `json:"schema_version"`
	DraftID             string `json:"draft_id"`
	DraftVersion        int    `json:"draft_version"`
	RevisionKind        string `json:"revision_kind"`
	RestoredFromVersion int    `json:"restored_from_version"`
	DraftStatus         string `json:"draft_status"`
	Name                string `json:"name"`
	UpdatedAt           string `json:"updated_at"`
	UpdatedByActorRef   string `json:"updated_by_actor_ref"`
	NodeCount           int    `json:"node_count"`
	EdgeCount           int    `json:"edge_count"`
	BlockedCount        int    `json:"blocked_count"`
}

type savedWorkflowDraftRevisionEnvelope struct {
	RequestID               string                               `json:"request_id"`
	WorkspaceID             string                               `json:"workspace_id"`
	ApplicationID           string                               `json:"application_id"`
	Revision                *savedWorkflowDraftRevisionDocument  `json:"revision"`
	Draft                   *savedWorkflowDraftDocument          `json:"draft"`
	FailureCode             *string                              `json:"failure_code"`
	CurrentDraftVersion     int                                  `json:"current_draft_version"`
	CurrentLifecycleVersion int                                  `json:"current_lifecycle_version"`
	CurrentLifecycleState   string                               `json:"current_lifecycle_state"`
	ValidationSummary       savedWorkflowDraftValidationDocument `json:"validation_summary"`
	AuditRef                string                               `json:"audit_ref"`
}

type savedWorkflowDraftRevisionListEnvelope struct {
	RequestID     string                                      `json:"request_id"`
	WorkspaceID   string                                      `json:"workspace_id"`
	ApplicationID string                                      `json:"application_id"`
	Revisions     []savedWorkflowDraftRevisionSummaryDocument `json:"revisions"`
	NextCursor    string                                      `json:"next_cursor"`
	HasMore       bool                                        `json:"has_more"`
	FailureCode   *string                                     `json:"failure_code"`
	AuditRef      string                                      `json:"audit_ref"`
}

func (s *Server) handleListWorkflowDraftRevisions(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, savedWorkflowDraftRevisionListRoute)
	if !s.allowSavedWorkflowDraftDevHTTP(writer, request, trace) {
		return
	}
	context, failure := savedWorkflowDraftRevisionReadContext(request, trace, "revision-list")
	if failure != "" {
		writeSavedWorkflowDraftRevisionListResult(
			writer,
			trace,
			context,
			savedWorkflowDraftRevisionListFailure(failure, savedWorkflowDraftAuditMetadata(context)),
		)
		return
	}
	limit := 0
	if rawLimit := strings.TrimSpace(request.URL.Query().Get("limit")); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil {
			writeSavedWorkflowDraftRevisionListResult(
				writer,
				trace,
				context,
				savedWorkflowDraftRevisionListFailure(
					SavedWorkflowDraftFailurePayloadInvalid,
					savedWorkflowDraftAuditMetadata(context),
				),
			)
			return
		}
		limit = parsed
	}
	result := s.savedWorkflowDraftService().ListDraftRevisions(
		context,
		ListSavedWorkflowDraftRevisionsRequest{
			DraftID: strings.TrimSpace(request.PathValue("draft_id")),
			Limit:   limit,
			Cursor:  strings.TrimSpace(request.URL.Query().Get("cursor")),
		},
	)
	writeSavedWorkflowDraftRevisionListResult(writer, trace, context, result)
}

func (s *Server) handleReadWorkflowDraftRevision(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, savedWorkflowDraftRevisionReadRoute)
	if !s.allowSavedWorkflowDraftDevHTTP(writer, request, trace) {
		return
	}
	context, failure := savedWorkflowDraftRevisionReadContext(request, trace, "revision-read")
	if failure != "" {
		writeSavedWorkflowDraftRevisionResult(
			writer,
			http.StatusOK,
			trace,
			context,
			savedWorkflowDraftRevisionFailure(failure, savedWorkflowDraftAuditMetadata(context)),
		)
		return
	}
	version, err := strconv.Atoi(strings.TrimSpace(request.PathValue("draft_version")))
	if err != nil {
		version = 0
	}
	result := s.savedWorkflowDraftService().ReadDraftRevision(
		context,
		ReadSavedWorkflowDraftRevisionRequest{
			DraftID:      strings.TrimSpace(request.PathValue("draft_id")),
			DraftVersion: version,
		},
	)
	writeSavedWorkflowDraftRevisionResult(writer, http.StatusOK, trace, context, result)
}

func (s *Server) handleRestoreWorkflowDraftRevision(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, savedWorkflowDraftRevisionRestoreRoute)
	if !s.allowSavedWorkflowDraftDevHTTP(writer, request, trace) {
		return
	}
	auth, failureCode, status := s.authorizeWorkspaceScopedPermissions(
		request,
		"workflow_drafts:read",
		"workflow_drafts:write",
	)
	applicationID := strings.TrimSpace(request.Header.Get(savedWorkflowDraftDevApplicationHeader))
	context := savedWorkflowDraftMutationContext(
		request,
		trace,
		auth,
		applicationID,
		s.config.WorkflowSavedDraftDevWriteEnabled,
		"revision-restore",
	)
	if failureCode != "" {
		writeSavedWorkflowDraftRevisionResult(
			writer,
			status,
			trace,
			context,
			savedWorkflowDraftRevisionFailure(
				SavedWorkflowDraftFailureCode(failureCode),
				savedWorkflowDraftAuditMetadata(context),
			),
		)
		return
	}
	if strings.TrimSpace(request.Header.Get(savedWorkflowDraftDevWorkspaceHeader)) !=
		auth.ResourceBinding.WorkspaceID || applicationID == "" {
		writeSavedWorkflowDraftRevisionResult(
			writer,
			http.StatusForbidden,
			trace,
			context,
			savedWorkflowDraftRevisionFailure(
				SavedWorkflowDraftFailureScopeDenied,
				savedWorkflowDraftAuditMetadata(context),
			),
		)
		return
	}
	var body savedWorkflowDraftRevisionRestoreHTTPBody
	if !s.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{
		maxBytes:            maxControlJSONRequestBodyBytes,
		rejectUnknownFields: true,
	}) {
		return
	}
	version, err := strconv.Atoi(strings.TrimSpace(request.PathValue("draft_version")))
	if err != nil {
		version = 0
	}
	result := s.savedWorkflowDraftService().RestoreDraftRevision(
		context,
		RestoreSavedWorkflowDraftRevisionRequest{
			DraftID:                     strings.TrimSpace(request.PathValue("draft_id")),
			SourceDraftVersion:          version,
			ExpectedCurrentDraftVersion: body.ExpectedCurrentDraftVersion,
			ExpectedLifecycleVersion:    body.ExpectedLifecycleVersion,
		},
	)
	writeSavedWorkflowDraftRevisionResult(writer, http.StatusOK, trace, context, result)
}

func savedWorkflowDraftRevisionReadContext(
	request *http.Request,
	trace requestTrace,
	auditSuffix string,
) (SavedWorkflowDraftContext, SavedWorkflowDraftFailureCode) {
	return savedWorkflowDraftContextFromRequest(
		request,
		trace,
		strings.TrimSpace(request.URL.Query().Get("workspace_id")),
		strings.TrimSpace(request.URL.Query().Get("application_id")),
		"workflow_drafts:read",
		false,
		auditSuffix,
	)
}

func writeSavedWorkflowDraftRevisionResult(
	writer http.ResponseWriter,
	status int,
	trace requestTrace,
	context SavedWorkflowDraftContext,
	result SavedWorkflowDraftRevisionResult,
) {
	writeObservedJSON(writer, status, trace, savedWorkflowDraftRevisionEnvelope{
		RequestID:               trace.requestID,
		WorkspaceID:             strings.TrimSpace(context.WorkspaceID),
		ApplicationID:           strings.TrimSpace(context.ApplicationID),
		Revision:                savedWorkflowDraftRevisionDocumentPointer(result.Revision),
		Draft:                   savedWorkflowDraftDocumentPointer(result.Draft),
		FailureCode:             savedWorkflowDraftFailureCodePointer(result.FailureCode),
		CurrentDraftVersion:     result.CurrentDraftVersion,
		CurrentLifecycleVersion: result.CurrentLifecycleVersion,
		CurrentLifecycleState:   string(result.CurrentLifecycleState),
		ValidationSummary:       savedWorkflowDraftValidationToDocument(result.ValidationSummary),
		AuditRef:                strings.TrimSpace(result.RequestAuditMetadata.AuditRef),
	})
}

func writeSavedWorkflowDraftRevisionListResult(
	writer http.ResponseWriter,
	trace requestTrace,
	context SavedWorkflowDraftContext,
	result SavedWorkflowDraftRevisionListResult,
) {
	writeObservedJSON(writer, http.StatusOK, trace, savedWorkflowDraftRevisionListEnvelope{
		RequestID:     trace.requestID,
		WorkspaceID:   strings.TrimSpace(context.WorkspaceID),
		ApplicationID: strings.TrimSpace(context.ApplicationID),
		Revisions:     savedWorkflowDraftRevisionSummariesToDocuments(result.Revisions),
		NextCursor:    result.NextCursor,
		HasMore:       result.HasMore,
		FailureCode:   savedWorkflowDraftFailureCodePointer(result.FailureCode),
		AuditRef:      strings.TrimSpace(result.RequestAuditMetadata.AuditRef),
	})
}

func savedWorkflowDraftRevisionDocumentPointer(
	revision *SavedWorkflowDraftRevision,
) *savedWorkflowDraftRevisionDocument {
	if revision == nil {
		return nil
	}
	return &savedWorkflowDraftRevisionDocument{
		SchemaVersion:       revision.SchemaVersion,
		Draft:               savedWorkflowDraftDocumentPointer(&revision.Draft),
		RevisionKind:        string(revision.RevisionKind),
		RestoredFromVersion: revision.RestoredFromVersion,
	}
}

func savedWorkflowDraftRevisionSummariesToDocuments(
	revisions []SavedWorkflowDraftRevisionSummary,
) []savedWorkflowDraftRevisionSummaryDocument {
	documents := make([]savedWorkflowDraftRevisionSummaryDocument, 0, len(revisions))
	for _, revision := range revisions {
		documents = append(documents, savedWorkflowDraftRevisionSummaryDocument{
			SchemaVersion:       revision.SchemaVersion,
			DraftID:             revision.DraftID,
			DraftVersion:        revision.DraftVersion,
			RevisionKind:        string(revision.RevisionKind),
			RestoredFromVersion: revision.RestoredFromVersion,
			DraftStatus:         string(revision.DraftStatus),
			Name:                revision.Name,
			UpdatedAt:           revision.UpdatedAt,
			UpdatedByActorRef:   revision.UpdatedByActorRef,
			NodeCount:           revision.NodeCount,
			EdgeCount:           revision.EdgeCount,
			BlockedCount:        revision.BlockedCount,
		})
	}
	return documents
}

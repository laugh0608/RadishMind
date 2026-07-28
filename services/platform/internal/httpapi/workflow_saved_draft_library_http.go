package httpapi

import (
	"net/http"
	"strconv"
	"strings"
)

type savedWorkflowDraftLifecycleHTTPBody struct {
	WorkspaceID              string `json:"workspace_id"`
	ApplicationID            string `json:"application_id"`
	ExpectedDraftVersion     int    `json:"expected_draft_version"`
	ExpectedLifecycleVersion int    `json:"expected_lifecycle_version"`
}

type savedWorkflowDraftLifecycleDocument struct {
	DraftID                    string  `json:"draft_id"`
	LifecycleState             string  `json:"lifecycle_state"`
	LifecycleVersion           int     `json:"lifecycle_version"`
	ArchivedAt                 *string `json:"archived_at"`
	LibraryUpdatedAt           string  `json:"library_updated_at"`
	LifecycleUpdatedByActorRef string  `json:"lifecycle_updated_by_actor_ref"`
}

type savedWorkflowDraftLifecycleEnvelope struct {
	RequestID               string                               `json:"request_id"`
	WorkspaceID             string                               `json:"workspace_id"`
	ApplicationID           string                               `json:"application_id"`
	Lifecycle               *savedWorkflowDraftLifecycleDocument `json:"lifecycle"`
	FailureCode             *string                              `json:"failure_code"`
	CurrentDraftVersion     int                                  `json:"current_draft_version"`
	CurrentLifecycleVersion int                                  `json:"current_lifecycle_version"`
	CurrentLifecycleState   string                               `json:"current_lifecycle_state"`
	AuditRef                string                               `json:"audit_ref"`
}

func (s *Server) handleArchiveWorkflowDraft(writer http.ResponseWriter, request *http.Request) {
	s.handleWorkflowDraftLifecycleTransition(
		writer,
		request,
		savedWorkflowDraftArchiveRoute,
		"archive",
		SavedWorkflowDraftLifecycleArchived,
	)
}

func (s *Server) handleUnarchiveWorkflowDraft(writer http.ResponseWriter, request *http.Request) {
	s.handleWorkflowDraftLifecycleTransition(
		writer,
		request,
		savedWorkflowDraftUnarchiveRoute,
		"unarchive",
		SavedWorkflowDraftLifecycleActive,
	)
}

func (s *Server) handleWorkflowDraftLifecycleTransition(
	writer http.ResponseWriter,
	request *http.Request,
	route string,
	auditSuffix string,
	target SavedWorkflowDraftLifecycleState,
) {
	trace := newRequestTrace(request, route)
	if !s.allowSavedWorkflowDraftDevHTTP(writer, request, trace) {
		return
	}
	auth, failureCode, status := s.authorizeWorkspaceScopedPermissions(
		request,
		"workflow_drafts:read",
		"workflow_drafts:archive",
	)
	context := savedWorkflowDraftMutationContext(
		request,
		trace,
		auth,
		request.Header.Get(savedWorkflowDraftDevApplicationHeader),
		s.config.WorkflowSavedDraftDevWriteEnabled,
		auditSuffix,
	)
	if failureCode != "" {
		writeSavedWorkflowDraftLifecycleResultWithStatus(
			writer,
			status,
			trace,
			context,
			savedWorkflowDraftLifecycleFailure(
				SavedWorkflowDraftFailureCode(failureCode),
				savedWorkflowDraftAuditMetadata(context),
			),
		)
		return
	}
	var body savedWorkflowDraftLifecycleHTTPBody
	if !s.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{
		maxBytes:            maxControlJSONRequestBodyBytes,
		rejectUnknownFields: true,
	}) {
		return
	}
	body.WorkspaceID = strings.TrimSpace(body.WorkspaceID)
	body.ApplicationID = strings.TrimSpace(body.ApplicationID)
	context = savedWorkflowDraftMutationContext(
		request,
		trace,
		auth,
		body.ApplicationID,
		s.config.WorkflowSavedDraftDevWriteEnabled,
		auditSuffix,
	)
	if len(request.URL.Query()) != 0 ||
		body.WorkspaceID != auth.ResourceBinding.WorkspaceID ||
		body.WorkspaceID != strings.TrimSpace(request.Header.Get(savedWorkflowDraftDevWorkspaceHeader)) ||
		body.ApplicationID == "" ||
		body.ApplicationID != strings.TrimSpace(request.Header.Get(savedWorkflowDraftDevApplicationHeader)) {
		writeSavedWorkflowDraftLifecycleResultWithStatus(
			writer,
			http.StatusForbidden,
			trace,
			context,
			savedWorkflowDraftLifecycleFailure(
				SavedWorkflowDraftFailureScopeDenied,
				savedWorkflowDraftAuditMetadata(context),
			),
		)
		return
	}
	service := s.savedWorkflowDraftService()
	transitionRequest := TransitionSavedWorkflowDraftLifecycleRequest{
		DraftID:                  strings.TrimSpace(request.PathValue("draft_id")),
		ExpectedDraftVersion:     body.ExpectedDraftVersion,
		ExpectedLifecycleVersion: body.ExpectedLifecycleVersion,
	}
	var result SavedWorkflowDraftLifecycleTransitionResult
	if target == SavedWorkflowDraftLifecycleActive {
		result = service.UnarchiveDraft(context, transitionRequest)
	} else {
		result = service.ArchiveDraft(context, transitionRequest)
	}
	writeSavedWorkflowDraftLifecycleResultWithStatus(
		writer,
		http.StatusOK,
		trace,
		context,
		result,
	)
}

func (s *Server) savedWorkflowDraftLibraryContextFromRequest(
	request *http.Request,
	trace requestTrace,
	workspaceID string,
	applicationID string,
	auditSuffix string,
	requiredPermissions ...string,
) (SavedWorkflowDraftContext, string, int) {
	auth, failureCode, status := s.authorizeWorkspaceScopedPermissions(
		request,
		requiredPermissions...,
	)
	context := savedWorkflowDraftMutationContext(
		request,
		trace,
		auth,
		applicationID,
		false,
		auditSuffix,
	)
	workspaceID = strings.TrimSpace(workspaceID)
	applicationID = strings.TrimSpace(applicationID)
	if failureCode != "" {
		return context, failureCode, status
	}
	if workspaceID == "" ||
		applicationID == "" ||
		workspaceID != auth.ResourceBinding.WorkspaceID ||
		workspaceID != strings.TrimSpace(request.Header.Get(savedWorkflowDraftDevWorkspaceHeader)) ||
		applicationID != strings.TrimSpace(request.Header.Get(savedWorkflowDraftDevApplicationHeader)) {
		return context, string(SavedWorkflowDraftFailureScopeDenied), http.StatusForbidden
	}
	context.WorkspaceID = workspaceID
	context.ApplicationID = applicationID
	return context, "", http.StatusOK
}

func savedWorkflowDraftListRequestFromQuery(
	request *http.Request,
) (ListWorkflowDraftsRequest, SavedWorkflowDraftFailureCode) {
	query := request.URL.Query()
	allowed := map[string]struct{}{
		"workspace_id":     {},
		"application_id":   {},
		"lifecycle_state":  {},
		"limit":            {},
		"cursor":           {},
		"name_prefix":      {},
		"validation_state": {},
		"provenance_kind":  {},
	}
	for key, values := range query {
		if _, ok := allowed[key]; !ok || len(values) != 1 {
			return ListWorkflowDraftsRequest{}, SavedWorkflowDraftFailureListFilterInvalid
		}
	}
	result := ListWorkflowDraftsRequest{
		LifecycleState:  SavedWorkflowDraftLifecycleState(query.Get("lifecycle_state")),
		Cursor:          query.Get("cursor"),
		NamePrefix:      query.Get("name_prefix"),
		ValidationState: SavedWorkflowDraftStatus(query.Get("validation_state")),
		ProvenanceKind:  SavedWorkflowDraftProvenanceKind(query.Get("provenance_kind")),
	}
	if rawLimit, found := query["limit"]; found {
		limit, err := strconv.Atoi(rawLimit[0])
		if err != nil || limit < 1 || limit > maxSavedWorkflowDraftListLimit {
			return ListWorkflowDraftsRequest{}, SavedWorkflowDraftFailureListFilterInvalid
		}
		result.Limit = limit
	}
	return result, ""
}

func writeSavedWorkflowDraftLifecycleResultWithStatus(
	writer http.ResponseWriter,
	status int,
	trace requestTrace,
	context SavedWorkflowDraftContext,
	result SavedWorkflowDraftLifecycleTransitionResult,
) {
	var lifecycle *savedWorkflowDraftLifecycleDocument
	if result.Draft != nil {
		lifecycle = &savedWorkflowDraftLifecycleDocument{
			DraftID:                    result.Draft.DraftID,
			LifecycleState:             string(result.Draft.LifecycleState),
			LifecycleVersion:           result.Draft.LifecycleVersion,
			ArchivedAt:                 savedWorkflowDraftOptionalString(result.Draft.ArchivedAt),
			LibraryUpdatedAt:           result.Draft.LibraryUpdatedAt,
			LifecycleUpdatedByActorRef: result.Draft.LifecycleUpdatedByActorRef,
		}
	}
	writeObservedJSON(writer, status, trace, savedWorkflowDraftLifecycleEnvelope{
		RequestID:               trace.requestID,
		WorkspaceID:             strings.TrimSpace(context.WorkspaceID),
		ApplicationID:           strings.TrimSpace(context.ApplicationID),
		Lifecycle:               lifecycle,
		FailureCode:             savedWorkflowDraftFailureCodePointer(result.FailureCode),
		CurrentDraftVersion:     result.CurrentDraftVersion,
		CurrentLifecycleVersion: result.CurrentLifecycleVersion,
		CurrentLifecycleState:   string(result.CurrentLifecycleState),
		AuditRef:                strings.TrimSpace(result.RequestAuditMetadata.AuditRef),
	})
}

func savedWorkflowDraftOptionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func savedWorkflowDraftStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

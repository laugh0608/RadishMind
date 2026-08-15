package httpapi

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	applicationResultArtifactSchemaVersion        = "application_result_artifact.v1"
	applicationResultArtifactSummarySchemaVersion = "application_result_artifact_summary.v1"
	applicationResultArtifactMaxContentBytes      = 64 * 1024
	applicationResultArtifactDefaultListLimit     = 50
	applicationResultArtifactMaxListLimit         = 100

	ApplicationResultArtifactFailurePayloadInvalid    = "application_result_artifact_payload_invalid"
	ApplicationResultArtifactFailureSourceUnavailable = "application_result_artifact_source_unavailable"
	ApplicationResultArtifactFailureContentTooLarge   = "application_result_artifact_content_too_large"
	ApplicationResultArtifactFailureNotFound          = "application_result_artifact_not_found"
	ApplicationResultArtifactFailureConflict          = "application_result_artifact_conflict"
	ApplicationResultArtifactFailureStoreUnavailable  = "application_result_artifact_store_unavailable"
	ApplicationResultArtifactFailureStoreContract     = "application_result_artifact_store_contract_mismatch"
)

var (
	applicationResultArtifactIDPattern = regexp.MustCompile(`^appres_[a-z2-7]{16}$`)

	errApplicationResultArtifactNotFound = errors.New(ApplicationResultArtifactFailureNotFound)
	errApplicationResultArtifactConflict = errors.New(ApplicationResultArtifactFailureConflict)
	errApplicationResultArtifactStore    = errors.New(ApplicationResultArtifactFailureStoreUnavailable)
	errApplicationResultArtifactContract = errors.New(ApplicationResultArtifactFailureStoreContract)
)

type ApplicationResultArtifact struct {
	SchemaVersion     string                       `json:"schema_version"`
	ArtifactID        string                       `json:"artifact_id"`
	RecordVersion     int                          `json:"record_version"`
	TenantRef         string                       `json:"tenant_ref"`
	WorkspaceID       string                       `json:"workspace_id"`
	ApplicationID     string                       `json:"application_id"`
	OwnerSubjectRef   string                       `json:"owner_subject_ref"`
	SessionID         string                       `json:"session_id"`
	TurnID            string                       `json:"turn_id"`
	ClientTurnKey     string                       `json:"client_turn_key"`
	ExecutionProfile  string                       `json:"execution_profile"`
	RunRef            ApplicationInteractionRunRef `json:"run_ref"`
	ContentType       string                       `json:"content_type"`
	Content           string                       `json:"content"`
	ContentBytes      int                          `json:"content_bytes"`
	ContentDigest     string                       `json:"content_digest"`
	CreatedAt         string                       `json:"created_at"`
	CreatedByActorRef string                       `json:"created_by_actor_ref"`
	RequestID         string                       `json:"request_id"`
	AuditRef          string                       `json:"audit_ref"`
}

type ApplicationResultArtifactSummary struct {
	SchemaVersion    string                       `json:"schema_version"`
	ArtifactID       string                       `json:"artifact_id"`
	RecordVersion    int                          `json:"record_version"`
	TenantRef        string                       `json:"tenant_ref"`
	WorkspaceID      string                       `json:"workspace_id"`
	ApplicationID    string                       `json:"application_id"`
	OwnerSubjectRef  string                       `json:"owner_subject_ref"`
	SessionID        string                       `json:"session_id"`
	TurnID           string                       `json:"turn_id"`
	ClientTurnKey    string                       `json:"client_turn_key"`
	ExecutionProfile string                       `json:"execution_profile"`
	RunRef           ApplicationInteractionRunRef `json:"run_ref"`
	ContentType      string                       `json:"content_type"`
	ContentBytes     int                          `json:"content_bytes"`
	ContentDigest    string                       `json:"content_digest"`
	CreatedAt        string                       `json:"created_at"`
}

type ApplicationResultArtifactCaptureInput struct {
	Turn        ApplicationInteractionTurn
	ContentType string
	Content     string
}

type ApplicationResultArtifactResult struct {
	Artifact         *ApplicationResultArtifact
	Summary          *ApplicationResultArtifactSummary
	FailureCode      string
	IdempotentReplay bool
}

type ApplicationResultArtifactListInput struct {
	SessionID string
	Limit     int
	Cursor    string
}

type ApplicationResultArtifactListResult struct {
	Items       []ApplicationResultArtifactSummary
	NextCursor  *string
	FailureCode string
}

type applicationResultArtifactCursor struct {
	Version         int    `json:"version"`
	TenantRef       string `json:"tenant_ref"`
	WorkspaceID     string `json:"workspace_id"`
	ApplicationID   string `json:"application_id"`
	OwnerSubjectRef string `json:"owner_subject_ref"`
	SessionID       string `json:"session_id"`
	Limit           int    `json:"limit"`
	CreatedAt       string `json:"created_at"`
	ArtifactID      string `json:"artifact_id"`
}

type applicationResultArtifactRepository interface {
	Create(ApplicationInteractionContext, ApplicationResultArtifact) (ApplicationResultArtifact, bool, error)
	Read(ApplicationInteractionContext, string) (ApplicationResultArtifact, error)
	ReadByTurn(ApplicationInteractionContext, string, string) (ApplicationResultArtifact, error)
	List(ApplicationInteractionContext, string) ([]ApplicationResultArtifact, error)
}

type memoryApplicationResultArtifactRepository struct {
	mu          sync.RWMutex
	byID        map[string]ApplicationResultArtifact
	byTurn      map[string]string
	unavailable bool
}

type applicationResultArtifactService struct {
	repository applicationResultArtifactRepository
	turnReader func(ApplicationInteractionContext, string, string) (ApplicationInteractionTurn, error)
	now        func() time.Time
	newID      func(string) (string, error)
}

func newMemoryApplicationResultArtifactRepository() *memoryApplicationResultArtifactRepository {
	return &memoryApplicationResultArtifactRepository{
		byID: make(map[string]ApplicationResultArtifact), byTurn: make(map[string]string),
	}
}

func newApplicationResultArtifactService(
	repository applicationResultArtifactRepository,
	sessionRepositories ...applicationInteractionSessionRepository,
) applicationResultArtifactService {
	service := applicationResultArtifactService{
		repository: repository, now: func() time.Time { return time.Now().UTC() }, newID: newApplicationInteractionID,
	}
	if len(sessionRepositories) != 0 && sessionRepositories[0] != nil {
		service.turnReader = sessionRepositories[0].ReadTurn
	}
	return service
}

func (service applicationResultArtifactService) Capture(
	ctx ApplicationInteractionContext,
	input ApplicationResultArtifactCaptureInput,
) ApplicationResultArtifactResult {
	if service.repository == nil || service.turnReader == nil {
		return applicationResultArtifactFailure(ApplicationResultArtifactFailureStoreUnavailable)
	}
	if !ctx.WriteEnabled || validateApplicationInteractionContext(ctx) != nil ||
		validateStoredApplicationInteractionTurn(ctx, input.Turn) != nil ||
		input.Turn.Status != string(WorkflowRunStatusSucceeded) || input.Turn.RunRef == nil {
		return applicationResultArtifactFailure(ApplicationResultArtifactFailureSourceUnavailable)
	}
	storedTurn, readErr := service.turnReader(ctx, input.Turn.SessionID, input.Turn.TurnID)
	if readErr != nil {
		if errors.Is(readErr, errApplicationSessionStore) {
			return applicationResultArtifactFailure(ApplicationResultArtifactFailureStoreUnavailable)
		}
		return applicationResultArtifactFailure(ApplicationResultArtifactFailureSourceUnavailable)
	}
	if storedTurn.TurnID != input.Turn.TurnID || !applicationInteractionTurnsIdempotentlyEqual(storedTurn, input.Turn) ||
		validateStoredApplicationInteractionTurn(ctx, storedTurn) != nil || storedTurn.Status != string(WorkflowRunStatusSucceeded) {
		return applicationResultArtifactFailure(ApplicationResultArtifactFailureSourceUnavailable)
	}
	contentType := strings.TrimSpace(input.ContentType)
	content := input.Content
	if (contentType != "text/markdown" && contentType != "application/json") ||
		strings.TrimSpace(content) == "" || !utf8.ValidString(content) {
		return applicationResultArtifactFailure(ApplicationResultArtifactFailurePayloadInvalid)
	}
	if len([]byte(content)) > applicationResultArtifactMaxContentBytes {
		return applicationResultArtifactFailure(ApplicationResultArtifactFailureContentTooLarge)
	}
	artifactID, err := service.newID("appres")
	if err != nil {
		return applicationResultArtifactFailure(ApplicationResultArtifactFailureStoreUnavailable)
	}
	now := service.now().UTC().Format(time.RFC3339Nano)
	artifact := ApplicationResultArtifact{
		SchemaVersion: applicationResultArtifactSchemaVersion, ArtifactID: artifactID, RecordVersion: 1,
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		OwnerSubjectRef: ctx.OwnerSubjectRef, SessionID: input.Turn.SessionID, TurnID: input.Turn.TurnID,
		ClientTurnKey: input.Turn.ClientTurnKey, ExecutionProfile: input.Turn.ExecutionProfile,
		RunRef: *cloneApplicationInteractionRunRef(input.Turn.RunRef), ContentType: contentType, Content: content,
		ContentBytes: len([]byte(content)), ContentDigest: applicationResultArtifactContentDigest(contentType, content),
		CreatedAt: now, CreatedByActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
	if validateApplicationResultArtifact(ctx, artifact) != nil {
		return applicationResultArtifactFailure(ApplicationResultArtifactFailureStoreContract)
	}
	created, replay, err := service.repository.Create(ctx, artifact)
	if err != nil {
		return applicationResultArtifactRepositoryFailure(err)
	}
	if validateApplicationResultArtifact(ctx, created) != nil {
		return applicationResultArtifactFailure(ApplicationResultArtifactFailureStoreContract)
	}
	summary := applicationResultArtifactSummary(created)
	return ApplicationResultArtifactResult{Artifact: &created, Summary: &summary, IdempotentReplay: replay}
}

func (service applicationResultArtifactService) Read(
	ctx ApplicationInteractionContext,
	artifactID string,
) ApplicationResultArtifactResult {
	if service.repository == nil {
		return applicationResultArtifactFailure(ApplicationResultArtifactFailureStoreUnavailable)
	}
	if validateApplicationInteractionContext(ctx) != nil ||
		!applicationResultArtifactIDPattern.MatchString(strings.TrimSpace(artifactID)) {
		return applicationResultArtifactFailure(ApplicationResultArtifactFailurePayloadInvalid)
	}
	artifact, err := service.repository.Read(ctx, strings.TrimSpace(artifactID))
	if err != nil {
		return applicationResultArtifactRepositoryFailure(err)
	}
	if validateApplicationResultArtifact(ctx, artifact) != nil {
		return applicationResultArtifactFailure(ApplicationResultArtifactFailureStoreContract)
	}
	summary := applicationResultArtifactSummary(artifact)
	return ApplicationResultArtifactResult{Artifact: &artifact, Summary: &summary}
}

func (service applicationResultArtifactService) ReadByTurn(
	ctx ApplicationInteractionContext,
	sessionID string,
	turnID string,
) ApplicationResultArtifactResult {
	if service.repository == nil {
		return applicationResultArtifactFailure(ApplicationResultArtifactFailureStoreUnavailable)
	}
	if validateApplicationInteractionContext(ctx) != nil ||
		!applicationSessionIDPattern.MatchString(strings.TrimSpace(sessionID)) ||
		!applicationTurnIDPattern.MatchString(strings.TrimSpace(turnID)) {
		return applicationResultArtifactFailure(ApplicationResultArtifactFailurePayloadInvalid)
	}
	artifact, err := service.repository.ReadByTurn(ctx, strings.TrimSpace(sessionID), strings.TrimSpace(turnID))
	if err != nil {
		return applicationResultArtifactRepositoryFailure(err)
	}
	if validateApplicationResultArtifact(ctx, artifact) != nil {
		return applicationResultArtifactFailure(ApplicationResultArtifactFailureStoreContract)
	}
	summary := applicationResultArtifactSummary(artifact)
	return ApplicationResultArtifactResult{Artifact: &artifact, Summary: &summary}
}

func (service applicationResultArtifactService) List(
	ctx ApplicationInteractionContext,
	input ApplicationResultArtifactListInput,
) ApplicationResultArtifactListResult {
	result := ApplicationResultArtifactListResult{Items: []ApplicationResultArtifactSummary{}}
	if service.repository == nil {
		result.FailureCode = ApplicationResultArtifactFailureStoreUnavailable
		return result
	}
	if validateApplicationInteractionContext(ctx) != nil ||
		!applicationSessionIDPattern.MatchString(strings.TrimSpace(input.SessionID)) {
		result.FailureCode = ApplicationResultArtifactFailurePayloadInvalid
		return result
	}
	limit := input.Limit
	if limit == 0 {
		limit = applicationResultArtifactDefaultListLimit
	}
	if limit < 1 || limit > applicationResultArtifactMaxListLimit {
		result.FailureCode = ApplicationResultArtifactFailurePayloadInvalid
		return result
	}
	var cursor *applicationResultArtifactCursor
	if strings.TrimSpace(input.Cursor) != "" {
		decoded, err := decodeApplicationResultArtifactCursor(input.Cursor)
		if err != nil || decoded.Version != 1 || decoded.TenantRef != ctx.TenantRef ||
			decoded.WorkspaceID != ctx.WorkspaceID || decoded.ApplicationID != ctx.ApplicationID ||
			decoded.OwnerSubjectRef != ctx.OwnerSubjectRef || decoded.SessionID != strings.TrimSpace(input.SessionID) ||
			decoded.Limit != limit || parseApplicationInteractionTimestamp(decoded.CreatedAt) == nil ||
			!applicationResultArtifactIDPattern.MatchString(decoded.ArtifactID) {
			result.FailureCode = ApplicationResultArtifactFailurePayloadInvalid
			return result
		}
		cursor = &decoded
	}
	artifacts, err := service.repository.List(ctx, strings.TrimSpace(input.SessionID))
	if err != nil {
		result.FailureCode = applicationResultArtifactRepositoryFailure(err).FailureCode
		return result
	}
	for _, artifact := range artifacts {
		if validateApplicationResultArtifact(ctx, artifact) != nil || artifact.SessionID != strings.TrimSpace(input.SessionID) {
			result.FailureCode = ApplicationResultArtifactFailureStoreContract
			result.Items = []ApplicationResultArtifactSummary{}
			return result
		}
	}
	sort.Slice(artifacts, func(left, right int) bool {
		leftTime := parseApplicationInteractionTimestamp(artifacts[left].CreatedAt)
		rightTime := parseApplicationInteractionTimestamp(artifacts[right].CreatedAt)
		if !leftTime.Equal(*rightTime) {
			return leftTime.After(*rightTime)
		}
		return artifacts[left].ArtifactID > artifacts[right].ArtifactID
	})
	filtered := make([]ApplicationResultArtifact, 0, len(artifacts))
	var cursorTime *time.Time
	if cursor != nil {
		cursorTime = parseApplicationInteractionTimestamp(cursor.CreatedAt)
	}
	for _, artifact := range artifacts {
		if cursor != nil {
			artifactTime := parseApplicationInteractionTimestamp(artifact.CreatedAt)
			if artifactTime.After(*cursorTime) || (artifactTime.Equal(*cursorTime) && artifact.ArtifactID >= cursor.ArtifactID) {
				continue
			}
		}
		filtered = append(filtered, artifact)
	}
	pageSize := len(filtered)
	if pageSize > limit {
		pageSize = limit
	}
	for _, artifact := range filtered[:pageSize] {
		result.Items = append(result.Items, applicationResultArtifactSummary(artifact))
	}
	if len(filtered) > limit {
		last := filtered[limit-1]
		next, err := encodeApplicationResultArtifactCursor(applicationResultArtifactCursor{
			Version: 1, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
			OwnerSubjectRef: ctx.OwnerSubjectRef, SessionID: strings.TrimSpace(input.SessionID), Limit: limit,
			CreatedAt: last.CreatedAt, ArtifactID: last.ArtifactID,
		})
		if err != nil {
			result.Items = []ApplicationResultArtifactSummary{}
			result.FailureCode = ApplicationResultArtifactFailureStoreContract
			return result
		}
		result.NextCursor = &next
	}
	return result
}

func (repository *memoryApplicationResultArtifactRepository) Create(
	ctx ApplicationInteractionContext,
	artifact ApplicationResultArtifact,
) (ApplicationResultArtifact, bool, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return ApplicationResultArtifact{}, false, errApplicationResultArtifactStore
	}
	if validateApplicationResultArtifact(ctx, artifact) != nil {
		return ApplicationResultArtifact{}, false, errApplicationResultArtifactContract
	}
	turnKey := applicationResultArtifactTurnKey(ctx, artifact.SessionID, artifact.TurnID)
	if existingID := repository.byTurn[turnKey]; existingID != "" {
		existing := repository.byID[existingID]
		if applicationResultArtifactsEquivalent(existing, artifact) {
			return cloneApplicationResultArtifact(existing), true, nil
		}
		return ApplicationResultArtifact{}, false, errApplicationResultArtifactConflict
	}
	if _, exists := repository.byID[artifact.ArtifactID]; exists {
		return ApplicationResultArtifact{}, false, errApplicationResultArtifactConflict
	}
	repository.byID[artifact.ArtifactID] = cloneApplicationResultArtifact(artifact)
	repository.byTurn[turnKey] = artifact.ArtifactID
	return cloneApplicationResultArtifact(artifact), false, nil
}

func (repository *memoryApplicationResultArtifactRepository) Read(
	ctx ApplicationInteractionContext,
	artifactID string,
) (ApplicationResultArtifact, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return ApplicationResultArtifact{}, errApplicationResultArtifactStore
	}
	artifact, found := repository.byID[strings.TrimSpace(artifactID)]
	if !found || !applicationResultArtifactScopeMatches(ctx, artifact) {
		return ApplicationResultArtifact{}, errApplicationResultArtifactNotFound
	}
	return cloneApplicationResultArtifact(artifact), nil
}

func (repository *memoryApplicationResultArtifactRepository) ReadByTurn(
	ctx ApplicationInteractionContext,
	sessionID string,
	turnID string,
) (ApplicationResultArtifact, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return ApplicationResultArtifact{}, errApplicationResultArtifactStore
	}
	artifactID := repository.byTurn[applicationResultArtifactTurnKey(ctx, sessionID, turnID)]
	artifact, found := repository.byID[artifactID]
	if !found || !applicationResultArtifactScopeMatches(ctx, artifact) {
		return ApplicationResultArtifact{}, errApplicationResultArtifactNotFound
	}
	return cloneApplicationResultArtifact(artifact), nil
}

func (repository *memoryApplicationResultArtifactRepository) List(
	ctx ApplicationInteractionContext,
	sessionID string,
) ([]ApplicationResultArtifact, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return nil, errApplicationResultArtifactStore
	}
	result := []ApplicationResultArtifact{}
	for _, artifact := range repository.byID {
		if artifact.SessionID == strings.TrimSpace(sessionID) && applicationResultArtifactScopeMatches(ctx, artifact) {
			result = append(result, cloneApplicationResultArtifact(artifact))
		}
	}
	return result, nil
}

func validateApplicationResultArtifact(ctx ApplicationInteractionContext, artifact ApplicationResultArtifact) error {
	if artifact.SchemaVersion != applicationResultArtifactSchemaVersion ||
		!applicationResultArtifactIDPattern.MatchString(strings.TrimSpace(artifact.ArtifactID)) || artifact.RecordVersion != 1 ||
		!applicationResultArtifactScopeMatches(ctx, artifact) ||
		!applicationSessionIDPattern.MatchString(artifact.SessionID) || !applicationTurnIDPattern.MatchString(artifact.TurnID) ||
		!applicationDraftIdentifierPattern.MatchString(artifact.ClientTurnKey) ||
		validateApplicationInteractionRunRef(artifact.ExecutionProfile, &artifact.RunRef) != nil ||
		(artifact.ContentType != "text/markdown" && artifact.ContentType != "application/json") ||
		strings.TrimSpace(artifact.Content) == "" || !utf8.ValidString(artifact.Content) ||
		artifact.ContentBytes != len([]byte(artifact.Content)) || artifact.ContentBytes > applicationResultArtifactMaxContentBytes ||
		artifact.ContentDigest != applicationResultArtifactContentDigest(artifact.ContentType, artifact.Content) ||
		parseApplicationInteractionTimestamp(artifact.CreatedAt) == nil || strings.TrimSpace(artifact.CreatedByActorRef) == "" ||
		strings.TrimSpace(artifact.RequestID) == "" || strings.TrimSpace(artifact.AuditRef) == "" {
		return errApplicationResultArtifactContract
	}
	return nil
}

func applicationResultArtifactScopeMatches(ctx ApplicationInteractionContext, artifact ApplicationResultArtifact) bool {
	return artifact.TenantRef == ctx.TenantRef && artifact.WorkspaceID == ctx.WorkspaceID &&
		artifact.ApplicationID == ctx.ApplicationID && artifact.OwnerSubjectRef == ctx.OwnerSubjectRef
}

func applicationResultArtifactContentDigest(contentType string, content string) string {
	digest := sha256.Sum256([]byte(strings.TrimSpace(contentType) + "\x00" + content))
	return "sha256:" + hex.EncodeToString(digest[:])
}

func applicationResultArtifactTurnKey(ctx ApplicationInteractionContext, sessionID string, turnID string) string {
	return strings.Join([]string{ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, strings.TrimSpace(sessionID), strings.TrimSpace(turnID)}, "\x00")
}

func applicationResultArtifactsEquivalent(left ApplicationResultArtifact, right ApplicationResultArtifact) bool {
	return left.TenantRef == right.TenantRef && left.WorkspaceID == right.WorkspaceID && left.ApplicationID == right.ApplicationID &&
		left.OwnerSubjectRef == right.OwnerSubjectRef && left.SessionID == right.SessionID && left.TurnID == right.TurnID &&
		left.ClientTurnKey == right.ClientTurnKey && left.ExecutionProfile == right.ExecutionProfile &&
		left.RunRef == right.RunRef && left.ContentType == right.ContentType && left.ContentDigest == right.ContentDigest &&
		left.ContentBytes == right.ContentBytes && left.Content == right.Content
}

func applicationResultArtifactSummary(artifact ApplicationResultArtifact) ApplicationResultArtifactSummary {
	return ApplicationResultArtifactSummary{
		SchemaVersion: applicationResultArtifactSummarySchemaVersion, ArtifactID: artifact.ArtifactID,
		RecordVersion: artifact.RecordVersion, TenantRef: artifact.TenantRef, WorkspaceID: artifact.WorkspaceID,
		ApplicationID: artifact.ApplicationID, OwnerSubjectRef: artifact.OwnerSubjectRef, SessionID: artifact.SessionID,
		TurnID: artifact.TurnID, ClientTurnKey: artifact.ClientTurnKey, ExecutionProfile: artifact.ExecutionProfile,
		RunRef: artifact.RunRef, ContentType: artifact.ContentType, ContentBytes: artifact.ContentBytes,
		ContentDigest: artifact.ContentDigest, CreatedAt: artifact.CreatedAt,
	}
}

func applicationResultArtifactFailure(code string) ApplicationResultArtifactResult {
	return ApplicationResultArtifactResult{FailureCode: code}
}

func applicationResultArtifactRepositoryFailure(err error) ApplicationResultArtifactResult {
	switch {
	case errors.Is(err, errApplicationResultArtifactNotFound):
		return applicationResultArtifactFailure(ApplicationResultArtifactFailureNotFound)
	case errors.Is(err, errApplicationResultArtifactConflict):
		return applicationResultArtifactFailure(ApplicationResultArtifactFailureConflict)
	case errors.Is(err, errApplicationResultArtifactContract):
		return applicationResultArtifactFailure(ApplicationResultArtifactFailureStoreContract)
	default:
		return applicationResultArtifactFailure(ApplicationResultArtifactFailureStoreUnavailable)
	}
}

func encodeApplicationResultArtifactCursor(cursor applicationResultArtifactCursor) (string, error) {
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeApplicationResultArtifactCursor(value string) (applicationResultArtifactCursor, error) {
	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil {
		return applicationResultArtifactCursor{}, err
	}
	var cursor applicationResultArtifactCursor
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cursor); err != nil {
		return applicationResultArtifactCursor{}, err
	}
	return cursor, nil
}

func cloneApplicationResultArtifact(artifact ApplicationResultArtifact) ApplicationResultArtifact {
	copy := artifact
	copy.RunRef = *cloneApplicationInteractionRunRef(&artifact.RunRef)
	return copy
}

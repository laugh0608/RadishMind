package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	applicationSessionSchemaVersion     = "application_session.v1"
	applicationSessionTurnSchemaVersion = "application_session_turn.v1"
	applicationSessionStateActive       = "active"
	applicationSessionStateClosed       = "closed"
	applicationSessionRetentionPolicy   = "metadata_only"
	applicationSessionDefaultListLimit  = 25
	applicationSessionMaximumListLimit  = 100
)

const (
	ApplicationInteractionFailureSessionNotFound     = "application_session_not_found"
	ApplicationInteractionFailureSessionClosed       = "application_session_closed"
	ApplicationInteractionFailureVersionConflict     = "application_session_version_conflict"
	ApplicationInteractionFailureIdempotencyConflict = "application_session_idempotency_conflict"
	ApplicationInteractionFailureCursorInvalid       = "application_session_cursor_invalid"
	ApplicationInteractionFailureWriteDisabled       = "application_session_write_disabled"
	ApplicationInteractionFailureStoreContract       = "application_session_store_contract_mismatch"
)

var (
	applicationSessionIDPattern          = regexp.MustCompile(`^appsess_[a-z2-7]{16}$`)
	applicationTurnIDPattern             = regexp.MustCompile(`^appturn_[a-z2-7]{16}$`)
	errApplicationSessionNotFound        = errors.New(ApplicationInteractionFailureSessionNotFound)
	errApplicationSessionClosed          = errors.New(ApplicationInteractionFailureSessionClosed)
	errApplicationSessionVersionConflict = errors.New(ApplicationInteractionFailureVersionConflict)
	errApplicationSessionIdempotency     = errors.New(ApplicationInteractionFailureIdempotencyConflict)
	errApplicationSessionStore           = errors.New(ApplicationInteractionFailureStoreUnavailable)
	errApplicationSessionContract        = errors.New(ApplicationInteractionFailureStoreContract)
)

type ApplicationInteractionContext struct {
	RequestContext  context.Context
	RequestID       string
	TenantRef       string
	WorkspaceID     string
	ApplicationID   string
	ActorRef        string
	OwnerSubjectRef string
	AuditRef        string
	WriteEnabled    bool
}

type ApplicationInteractionSession struct {
	SchemaVersion     string                                  `json:"schema_version"`
	SessionID         string                                  `json:"session_id"`
	TenantRef         string                                  `json:"tenant_ref"`
	WorkspaceID       string                                  `json:"workspace_id"`
	ApplicationID     string                                  `json:"application_id"`
	OwnerSubjectRef   string                                  `json:"owner_subject_ref"`
	State             string                                  `json:"state"`
	RecordVersion     int                                     `json:"record_version"`
	ProfileBinding    ApplicationInteractionProfileBinding    `json:"profile_binding"`
	Authority         ApplicationInteractionAuthoritySnapshot `json:"authority"`
	ContentRetention  string                                  `json:"content_retention"`
	TurnCount         int                                     `json:"turn_count"`
	LastTurnID        *string                                 `json:"last_turn_id"`
	CreatedAt         string                                  `json:"created_at"`
	UpdatedAt         string                                  `json:"updated_at"`
	ClosedAt          *string                                 `json:"closed_at"`
	CreatedByActorRef string                                  `json:"created_by_actor_ref"`
	UpdatedByActorRef string                                  `json:"updated_by_actor_ref"`
	RequestID         string                                  `json:"request_id"`
	AuditRef          string                                  `json:"audit_ref"`
}

type ApplicationInteractionRunRef struct {
	RunID         string `json:"run_id"`
	SchemaVersion string `json:"schema_version"`
}

type ApplicationInteractionTurn struct {
	SchemaVersion    string                                  `json:"schema_version"`
	TurnID           string                                  `json:"turn_id"`
	SessionID        string                                  `json:"session_id"`
	Sequence         int                                     `json:"sequence"`
	ClientTurnKey    string                                  `json:"client_turn_key"`
	TenantRef        string                                  `json:"tenant_ref"`
	WorkspaceID      string                                  `json:"workspace_id"`
	ApplicationID    string                                  `json:"application_id"`
	OwnerSubjectRef  string                                  `json:"owner_subject_ref"`
	ExecutionProfile string                                  `json:"execution_profile"`
	Authority        ApplicationInteractionAuthoritySnapshot `json:"authority"`
	Status           string                                  `json:"status"`
	InputDigest      string                                  `json:"input_digest"`
	InputBytes       int                                     `json:"input_bytes"`
	RunRef           *ApplicationInteractionRunRef           `json:"run_ref"`
	FailureCode      string                                  `json:"failure_code"`
	FailureSummary   string                                  `json:"failure_summary"`
	StartedAt        string                                  `json:"started_at"`
	CompletedAt      *string                                 `json:"completed_at"`
	ActorRef         string                                  `json:"actor_ref"`
	RequestID        string                                  `json:"request_id"`
	AuditRef         string                                  `json:"audit_ref"`
}

type ApplicationInteractionSessionCreateInput struct {
	ProfileBinding ApplicationInteractionProfileBinding
}

type ApplicationInteractionSessionListInput struct {
	State            string
	ExecutionProfile string
	Limit            int
	Cursor           string
}

type ApplicationInteractionTerminalTurnInput struct {
	ExpectedSessionVersion int
	ClientTurnKey          string
	Status                 string
	InputDigest            string
	InputBytes             int
	RunRef                 *ApplicationInteractionRunRef
	FailureCode            string
	FailureSummary         string
	StartedAt              time.Time
	CompletedAt            time.Time
}

type ApplicationInteractionTurnReservationInput struct {
	ExpectedSessionVersion int
	ClientTurnKey          string
	InputDigest            string
	InputBytes             int
	StartedAt              time.Time
}

type ApplicationInteractionTurnCompletionInput struct {
	Status         string
	RunRef         *ApplicationInteractionRunRef
	FailureCode    string
	FailureSummary string
	CompletedAt    time.Time
}

type ApplicationInteractionSessionResult struct {
	Session              *ApplicationInteractionSession
	Turn                 *ApplicationInteractionTurn
	FailureCode          string
	CurrentRecordVersion int
	CurrentState         string
	IdempotentReplay     bool
}

type ApplicationInteractionSessionListResult struct {
	Sessions    []ApplicationInteractionSession
	NextCursor  *string
	FailureCode string
}

type applicationInteractionSessionListQuery struct {
	State            string
	ExecutionProfile string
	Limit            int
	AfterUpdatedAt   string
	AfterSessionID   string
}

type applicationInteractionSessionCursor struct {
	TenantRef        string `json:"tenant_ref"`
	WorkspaceID      string `json:"workspace_id"`
	ApplicationID    string `json:"application_id"`
	OwnerSubjectRef  string `json:"owner_subject_ref"`
	State            string `json:"state"`
	ExecutionProfile string `json:"execution_profile"`
	UpdatedAt        string `json:"updated_at"`
	SessionID        string `json:"session_id"`
}

type applicationInteractionSessionRepository interface {
	Create(ApplicationInteractionContext, ApplicationInteractionSession) (ApplicationInteractionSession, error)
	Read(ApplicationInteractionContext, string) (ApplicationInteractionSession, error)
	List(ApplicationInteractionContext, applicationInteractionSessionListQuery) ([]ApplicationInteractionSession, error)
	Close(ApplicationInteractionContext, string, int, ApplicationInteractionSession) (ApplicationInteractionSession, error)
	ReserveTurn(ApplicationInteractionContext, int, ApplicationInteractionSession, ApplicationInteractionTurn) (ApplicationInteractionSession, ApplicationInteractionTurn, bool, error)
	ReadTurn(ApplicationInteractionContext, string, string) (ApplicationInteractionTurn, error)
	ReadTurnByClientKey(ApplicationInteractionContext, string, string) (ApplicationInteractionTurn, error)
	CompleteTurn(ApplicationInteractionContext, ApplicationInteractionTurn) (ApplicationInteractionSession, ApplicationInteractionTurn, bool, error)
	ListTurns(ApplicationInteractionContext, string) ([]ApplicationInteractionTurn, error)
}

type memoryApplicationInteractionSessionEntry struct {
	session ApplicationInteractionSession
	turns   []ApplicationInteractionTurn
}

type memoryApplicationInteractionSessionRepository struct {
	mu          sync.RWMutex
	entries     map[string]memoryApplicationInteractionSessionEntry
	unavailable bool
}

type applicationInteractionSessionService struct {
	repository applicationInteractionSessionRepository
	resolver   applicationInteractionAuthorityResolver
	now        func() time.Time
	newID      func(string) (string, error)
}

func newMemoryApplicationInteractionSessionRepository() *memoryApplicationInteractionSessionRepository {
	return &memoryApplicationInteractionSessionRepository{entries: make(map[string]memoryApplicationInteractionSessionEntry)}
}

func newApplicationInteractionSessionService(repository applicationInteractionSessionRepository, resolver applicationInteractionAuthorityResolver) applicationInteractionSessionService {
	return applicationInteractionSessionService{repository: repository, resolver: resolver, now: func() time.Time { return time.Now().UTC() }, newID: newApplicationInteractionID}
}

func (service applicationInteractionSessionService) Create(ctx ApplicationInteractionContext, input ApplicationInteractionSessionCreateInput) ApplicationInteractionSessionResult {
	if !ctx.WriteEnabled {
		return applicationInteractionSessionFailure(ApplicationInteractionFailureWriteDisabled)
	}
	if validateApplicationInteractionContext(ctx) != nil || validateApplicationInteractionProfileBinding(input.ProfileBinding) != nil {
		return applicationInteractionSessionFailure(ApplicationInteractionFailurePayloadInvalid)
	}
	authority, failure := service.resolver.Resolve(ctx, input.ProfileBinding)
	if failure != "" {
		return applicationInteractionSessionFailure(failure)
	}
	sessionID, err := service.newID("appsess")
	if err != nil || !applicationSessionIDPattern.MatchString(sessionID) {
		return applicationInteractionSessionFailure(ApplicationInteractionFailureStoreUnavailable)
	}
	now := service.now().UTC().Format(time.RFC3339Nano)
	session := ApplicationInteractionSession{SchemaVersion: applicationSessionSchemaVersion, SessionID: sessionID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, OwnerSubjectRef: ctx.OwnerSubjectRef, State: applicationSessionStateActive, RecordVersion: 1, ProfileBinding: normalizeApplicationInteractionProfileBinding(input.ProfileBinding), Authority: authority, ContentRetention: applicationSessionRetentionPolicy, TurnCount: 0, CreatedAt: now, UpdatedAt: now, CreatedByActorRef: ctx.ActorRef, UpdatedByActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef}
	created, err := service.repository.Create(ctx, session)
	if err != nil {
		return applicationInteractionRepositoryFailure(err)
	}
	return ApplicationInteractionSessionResult{Session: &created, CurrentRecordVersion: created.RecordVersion, CurrentState: created.State}
}

func (service applicationInteractionSessionService) Read(ctx ApplicationInteractionContext, sessionID string) ApplicationInteractionSessionResult {
	if validateApplicationInteractionContext(ctx) != nil || !applicationSessionIDPattern.MatchString(strings.TrimSpace(sessionID)) {
		return applicationInteractionSessionFailure(ApplicationInteractionFailurePayloadInvalid)
	}
	session, err := service.repository.Read(ctx, strings.TrimSpace(sessionID))
	if err != nil {
		return applicationInteractionRepositoryFailure(err)
	}
	return ApplicationInteractionSessionResult{Session: &session, CurrentRecordVersion: session.RecordVersion, CurrentState: session.State}
}

func (service applicationInteractionSessionService) List(ctx ApplicationInteractionContext, input ApplicationInteractionSessionListInput) ApplicationInteractionSessionListResult {
	if validateApplicationInteractionContext(ctx) != nil {
		return ApplicationInteractionSessionListResult{Sessions: []ApplicationInteractionSession{}, FailureCode: ApplicationInteractionFailureScopeDenied}
	}
	state := strings.TrimSpace(input.State)
	if state == "" {
		state = applicationSessionStateActive
	}
	profile := strings.TrimSpace(input.ExecutionProfile)
	if (state != applicationSessionStateActive && state != applicationSessionStateClosed) || (profile != "" && profile != applicationInteractionProfileWorkflow && profile != applicationInteractionProfileRAG) {
		return ApplicationInteractionSessionListResult{Sessions: []ApplicationInteractionSession{}, FailureCode: ApplicationInteractionFailurePayloadInvalid}
	}
	limit := input.Limit
	if limit == 0 {
		limit = applicationSessionDefaultListLimit
	}
	if limit < 1 || limit > applicationSessionMaximumListLimit {
		return ApplicationInteractionSessionListResult{Sessions: []ApplicationInteractionSession{}, FailureCode: ApplicationInteractionFailurePayloadInvalid}
	}
	query := applicationInteractionSessionListQuery{State: state, ExecutionProfile: profile, Limit: limit + 1}
	if strings.TrimSpace(input.Cursor) != "" {
		cursor, err := decodeApplicationInteractionSessionCursor(input.Cursor)
		if err != nil || cursor.TenantRef != ctx.TenantRef || cursor.WorkspaceID != ctx.WorkspaceID || cursor.ApplicationID != ctx.ApplicationID || cursor.OwnerSubjectRef != ctx.OwnerSubjectRef || cursor.State != state || cursor.ExecutionProfile != profile || !applicationSessionIDPattern.MatchString(cursor.SessionID) {
			return ApplicationInteractionSessionListResult{Sessions: []ApplicationInteractionSession{}, FailureCode: ApplicationInteractionFailureCursorInvalid}
		}
		query.AfterUpdatedAt, query.AfterSessionID = cursor.UpdatedAt, cursor.SessionID
	}
	sessions, err := service.repository.List(ctx, query)
	if err != nil {
		return ApplicationInteractionSessionListResult{Sessions: []ApplicationInteractionSession{}, FailureCode: applicationInteractionRepositoryFailure(err).FailureCode}
	}
	var next *string
	if len(sessions) > limit {
		sessions = sessions[:limit]
		last := sessions[len(sessions)-1]
		encoded, err := encodeApplicationInteractionSessionCursor(applicationInteractionSessionCursor{TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, OwnerSubjectRef: ctx.OwnerSubjectRef, State: state, ExecutionProfile: profile, UpdatedAt: last.UpdatedAt, SessionID: last.SessionID})
		if err != nil {
			return ApplicationInteractionSessionListResult{Sessions: []ApplicationInteractionSession{}, FailureCode: ApplicationInteractionFailureStoreUnavailable}
		}
		next = &encoded
	}
	return ApplicationInteractionSessionListResult{Sessions: sessions, NextCursor: next}
}

func (service applicationInteractionSessionService) Close(ctx ApplicationInteractionContext, sessionID string, expectedVersion int) ApplicationInteractionSessionResult {
	if !ctx.WriteEnabled {
		return applicationInteractionSessionFailure(ApplicationInteractionFailureWriteDisabled)
	}
	if validateApplicationInteractionContext(ctx) != nil || !applicationSessionIDPattern.MatchString(strings.TrimSpace(sessionID)) || expectedVersion < 1 {
		return applicationInteractionSessionFailure(ApplicationInteractionFailurePayloadInvalid)
	}
	current, err := service.repository.Read(ctx, strings.TrimSpace(sessionID))
	if err != nil {
		return applicationInteractionRepositoryFailure(err)
	}
	if current.RecordVersion != expectedVersion {
		return ApplicationInteractionSessionResult{FailureCode: ApplicationInteractionFailureVersionConflict, CurrentRecordVersion: current.RecordVersion, CurrentState: current.State}
	}
	if current.State == applicationSessionStateClosed {
		return ApplicationInteractionSessionResult{Session: &current, CurrentRecordVersion: current.RecordVersion, CurrentState: current.State, IdempotentReplay: true}
	}
	now := service.now().UTC().Format(time.RFC3339Nano)
	updated := current
	updated.State, updated.RecordVersion, updated.UpdatedAt, updated.ClosedAt = applicationSessionStateClosed, current.RecordVersion+1, now, &now
	updated.UpdatedByActorRef, updated.RequestID, updated.AuditRef = ctx.ActorRef, ctx.RequestID, ctx.AuditRef
	closed, err := service.repository.Close(ctx, current.SessionID, expectedVersion, updated)
	if err != nil {
		return applicationInteractionRepositoryFailure(err)
	}
	return ApplicationInteractionSessionResult{Session: &closed, CurrentRecordVersion: closed.RecordVersion, CurrentState: closed.State}
}

func (service applicationInteractionSessionService) AppendTerminalTurn(ctx ApplicationInteractionContext, sessionID string, input ApplicationInteractionTerminalTurnInput) ApplicationInteractionSessionResult {
	if validateApplicationInteractionTerminalTurnInput(input) != nil {
		return applicationInteractionSessionFailure(ApplicationInteractionFailurePayloadInvalid)
	}
	reserved := service.ReserveTurn(ctx, sessionID, ApplicationInteractionTurnReservationInput{ExpectedSessionVersion: input.ExpectedSessionVersion, ClientTurnKey: input.ClientTurnKey, InputDigest: input.InputDigest, InputBytes: input.InputBytes, StartedAt: input.StartedAt})
	if reserved.FailureCode != "" || reserved.Turn == nil {
		return reserved
	}
	completed := service.CompleteTurn(ctx, sessionID, reserved.Turn.TurnID, ApplicationInteractionTurnCompletionInput{Status: input.Status, RunRef: input.RunRef, FailureCode: input.FailureCode, FailureSummary: input.FailureSummary, CompletedAt: input.CompletedAt})
	completed.IdempotentReplay = completed.IdempotentReplay || reserved.IdempotentReplay
	return completed
}

func (service applicationInteractionSessionService) ReserveTurn(ctx ApplicationInteractionContext, sessionID string, input ApplicationInteractionTurnReservationInput) ApplicationInteractionSessionResult {
	if !ctx.WriteEnabled {
		return applicationInteractionSessionFailure(ApplicationInteractionFailureWriteDisabled)
	}
	if validateApplicationInteractionContext(ctx) != nil || !applicationSessionIDPattern.MatchString(strings.TrimSpace(sessionID)) || validateApplicationInteractionTurnReservationInput(input) != nil {
		return applicationInteractionSessionFailure(ApplicationInteractionFailurePayloadInvalid)
	}
	current, err := service.repository.Read(ctx, strings.TrimSpace(sessionID))
	if err != nil {
		return applicationInteractionRepositoryFailure(err)
	}
	existing, err := service.repository.ReadTurnByClientKey(ctx, current.SessionID, strings.TrimSpace(input.ClientTurnKey))
	if err == nil {
		if existing.InputDigest != strings.TrimSpace(input.InputDigest) || existing.InputBytes != input.InputBytes {
			return ApplicationInteractionSessionResult{FailureCode: ApplicationInteractionFailureIdempotencyConflict, CurrentRecordVersion: current.RecordVersion, CurrentState: current.State}
		}
		return ApplicationInteractionSessionResult{Session: &current, Turn: &existing, CurrentRecordVersion: current.RecordVersion, CurrentState: current.State, IdempotentReplay: true}
	}
	if !errors.Is(err, errApplicationSessionNotFound) {
		return applicationInteractionRepositoryFailure(err)
	}
	if current.State != applicationSessionStateActive {
		return ApplicationInteractionSessionResult{FailureCode: ApplicationInteractionFailureSessionClosed, CurrentRecordVersion: current.RecordVersion, CurrentState: current.State}
	}
	authority, failure := service.resolver.Resolve(ctx, current.ProfileBinding)
	if failure != "" {
		return applicationInteractionSessionFailure(failure)
	}
	if authority.AuthorityDigest != current.Authority.AuthorityDigest {
		return applicationInteractionSessionFailure(ApplicationInteractionFailureAuthorityChanged)
	}
	turnID, err := service.newID("appturn")
	if err != nil || !applicationTurnIDPattern.MatchString(turnID) {
		return applicationInteractionSessionFailure(ApplicationInteractionFailureStoreUnavailable)
	}
	updated := current
	updated.RecordVersion++
	updated.TurnCount++
	updated.LastTurnID = &turnID
	updated.UpdatedAt = input.StartedAt.UTC().Format(time.RFC3339Nano)
	updated.UpdatedByActorRef, updated.RequestID, updated.AuditRef = ctx.ActorRef, ctx.RequestID, ctx.AuditRef
	turn := ApplicationInteractionTurn{SchemaVersion: applicationSessionTurnSchemaVersion, TurnID: turnID, SessionID: current.SessionID, Sequence: current.TurnCount + 1, ClientTurnKey: strings.TrimSpace(input.ClientTurnKey), TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, OwnerSubjectRef: ctx.OwnerSubjectRef, ExecutionProfile: current.ProfileBinding.ExecutionProfile, Authority: authority, Status: string(WorkflowRunStatusRunning), InputDigest: strings.TrimSpace(input.InputDigest), InputBytes: input.InputBytes, StartedAt: input.StartedAt.UTC().Format(time.RFC3339Nano), ActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef}
	storedSession, storedTurn, replay, err := service.repository.ReserveTurn(ctx, input.ExpectedSessionVersion, updated, turn)
	if err != nil {
		return applicationInteractionRepositoryFailure(err)
	}
	return ApplicationInteractionSessionResult{Session: &storedSession, Turn: &storedTurn, CurrentRecordVersion: storedSession.RecordVersion, CurrentState: storedSession.State, IdempotentReplay: replay}
}

func (service applicationInteractionSessionService) CompleteTurn(ctx ApplicationInteractionContext, sessionID, turnID string, input ApplicationInteractionTurnCompletionInput) ApplicationInteractionSessionResult {
	if !ctx.WriteEnabled {
		return applicationInteractionSessionFailure(ApplicationInteractionFailureWriteDisabled)
	}
	if validateApplicationInteractionContext(ctx) != nil || !applicationSessionIDPattern.MatchString(strings.TrimSpace(sessionID)) || !applicationTurnIDPattern.MatchString(strings.TrimSpace(turnID)) || validateApplicationInteractionTurnCompletionInput(input) != nil {
		return applicationInteractionSessionFailure(ApplicationInteractionFailurePayloadInvalid)
	}
	current, err := service.repository.ReadTurn(ctx, strings.TrimSpace(sessionID), strings.TrimSpace(turnID))
	if err != nil {
		return applicationInteractionRepositoryFailure(err)
	}
	startedAt := parseApplicationInteractionTimestamp(current.StartedAt)
	if startedAt == nil || input.CompletedAt.Before(*startedAt) {
		return applicationInteractionSessionFailure(ApplicationInteractionFailurePayloadInvalid)
	}
	completedAt := input.CompletedAt.UTC().Format(time.RFC3339Nano)
	terminal := current
	terminal.Status = strings.TrimSpace(input.Status)
	terminal.RunRef = cloneApplicationInteractionRunRef(input.RunRef)
	terminal.FailureCode = strings.TrimSpace(input.FailureCode)
	terminal.FailureSummary = strings.TrimSpace(input.FailureSummary)
	terminal.CompletedAt = &completedAt
	terminal.ActorRef, terminal.RequestID, terminal.AuditRef = ctx.ActorRef, ctx.RequestID, ctx.AuditRef
	storedSession, storedTurn, replay, err := service.repository.CompleteTurn(ctx, terminal)
	if err != nil {
		return applicationInteractionRepositoryFailure(err)
	}
	return ApplicationInteractionSessionResult{Session: &storedSession, Turn: &storedTurn, CurrentRecordVersion: storedSession.RecordVersion, CurrentState: storedSession.State, IdempotentReplay: replay}
}

func (service applicationInteractionSessionService) ListTurns(ctx ApplicationInteractionContext, sessionID string) ([]ApplicationInteractionTurn, string) {
	if validateApplicationInteractionContext(ctx) != nil || !applicationSessionIDPattern.MatchString(strings.TrimSpace(sessionID)) {
		return []ApplicationInteractionTurn{}, ApplicationInteractionFailurePayloadInvalid
	}
	turns, err := service.repository.ListTurns(ctx, strings.TrimSpace(sessionID))
	if err != nil {
		return []ApplicationInteractionTurn{}, applicationInteractionRepositoryFailure(err).FailureCode
	}
	return turns, ""
}

func (repository *memoryApplicationInteractionSessionRepository) Create(ctx ApplicationInteractionContext, session ApplicationInteractionSession) (ApplicationInteractionSession, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return ApplicationInteractionSession{}, errApplicationSessionStore
	}
	if validateStoredApplicationInteractionSession(ctx, session) != nil {
		return ApplicationInteractionSession{}, errApplicationSessionContract
	}
	key := applicationInteractionSessionRepositoryKey(ctx, session.SessionID)
	if _, exists := repository.entries[key]; exists {
		return ApplicationInteractionSession{}, errApplicationSessionVersionConflict
	}
	repository.entries[key] = memoryApplicationInteractionSessionEntry{session: cloneApplicationInteractionSession(session), turns: []ApplicationInteractionTurn{}}
	return cloneApplicationInteractionSession(session), nil
}

func (repository *memoryApplicationInteractionSessionRepository) Read(ctx ApplicationInteractionContext, sessionID string) (ApplicationInteractionSession, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return ApplicationInteractionSession{}, errApplicationSessionStore
	}
	entry, exists := repository.entries[applicationInteractionSessionRepositoryKey(ctx, sessionID)]
	if !exists {
		return ApplicationInteractionSession{}, errApplicationSessionNotFound
	}
	if validateStoredApplicationInteractionSession(ctx, entry.session) != nil {
		return ApplicationInteractionSession{}, errApplicationSessionContract
	}
	return cloneApplicationInteractionSession(entry.session), nil
}

func (repository *memoryApplicationInteractionSessionRepository) List(ctx ApplicationInteractionContext, query applicationInteractionSessionListQuery) ([]ApplicationInteractionSession, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return nil, errApplicationSessionStore
	}
	prefix := applicationInteractionSessionRepositoryKey(ctx, "")
	items := make([]ApplicationInteractionSession, 0)
	for key, entry := range repository.entries {
		session := entry.session
		if !strings.HasPrefix(key, prefix) || session.State != query.State || (query.ExecutionProfile != "" && session.ProfileBinding.ExecutionProfile != query.ExecutionProfile) {
			continue
		}
		if query.AfterUpdatedAt != "" && (session.UpdatedAt > query.AfterUpdatedAt || (session.UpdatedAt == query.AfterUpdatedAt && session.SessionID >= query.AfterSessionID)) {
			continue
		}
		if validateStoredApplicationInteractionSession(ctx, session) != nil {
			return nil, errApplicationSessionContract
		}
		items = append(items, cloneApplicationInteractionSession(session))
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].UpdatedAt == items[j].UpdatedAt {
			return items[i].SessionID > items[j].SessionID
		}
		return items[i].UpdatedAt > items[j].UpdatedAt
	})
	if len(items) > query.Limit {
		items = items[:query.Limit]
	}
	return items, nil
}

func (repository *memoryApplicationInteractionSessionRepository) Close(ctx ApplicationInteractionContext, sessionID string, expectedVersion int, updated ApplicationInteractionSession) (ApplicationInteractionSession, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	key := applicationInteractionSessionRepositoryKey(ctx, sessionID)
	entry, exists := repository.entries[key]
	if repository.unavailable {
		return ApplicationInteractionSession{}, errApplicationSessionStore
	}
	if !exists {
		return ApplicationInteractionSession{}, errApplicationSessionNotFound
	}
	if entry.session.RecordVersion != expectedVersion {
		return ApplicationInteractionSession{}, errApplicationSessionVersionConflict
	}
	if entry.session.State != applicationSessionStateActive {
		return ApplicationInteractionSession{}, errApplicationSessionClosed
	}
	if validateApplicationInteractionSessionTransition(entry.session, updated) != nil || validateStoredApplicationInteractionSession(ctx, updated) != nil {
		return ApplicationInteractionSession{}, errApplicationSessionContract
	}
	entry.session = cloneApplicationInteractionSession(updated)
	repository.entries[key] = entry
	return cloneApplicationInteractionSession(updated), nil
}

func (repository *memoryApplicationInteractionSessionRepository) ReserveTurn(ctx ApplicationInteractionContext, expectedVersion int, updated ApplicationInteractionSession, turn ApplicationInteractionTurn) (ApplicationInteractionSession, ApplicationInteractionTurn, bool, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	key := applicationInteractionSessionRepositoryKey(ctx, turn.SessionID)
	entry, exists := repository.entries[key]
	if repository.unavailable {
		return ApplicationInteractionSession{}, ApplicationInteractionTurn{}, false, errApplicationSessionStore
	}
	if !exists {
		return ApplicationInteractionSession{}, ApplicationInteractionTurn{}, false, errApplicationSessionNotFound
	}
	for _, existing := range entry.turns {
		if existing.ClientTurnKey != turn.ClientTurnKey {
			continue
		}
		if applicationInteractionTurnReservationsEqual(existing, turn) {
			return cloneApplicationInteractionSession(entry.session), cloneApplicationInteractionTurn(existing), true, nil
		}
		return ApplicationInteractionSession{}, ApplicationInteractionTurn{}, false, errApplicationSessionIdempotency
	}
	if entry.session.State != applicationSessionStateActive {
		return ApplicationInteractionSession{}, ApplicationInteractionTurn{}, false, errApplicationSessionClosed
	}
	if entry.session.RecordVersion != expectedVersion {
		return ApplicationInteractionSession{}, ApplicationInteractionTurn{}, false, errApplicationSessionVersionConflict
	}
	if validateApplicationInteractionSessionReserve(entry.session, updated, turn) != nil || validateStoredApplicationInteractionSession(ctx, updated) != nil || validateStoredApplicationInteractionTurn(ctx, turn) != nil {
		return ApplicationInteractionSession{}, ApplicationInteractionTurn{}, false, errApplicationSessionContract
	}
	entry.session = cloneApplicationInteractionSession(updated)
	entry.turns = append(entry.turns, cloneApplicationInteractionTurn(turn))
	repository.entries[key] = entry
	return cloneApplicationInteractionSession(updated), cloneApplicationInteractionTurn(turn), false, nil
}

func (repository *memoryApplicationInteractionSessionRepository) ReadTurn(ctx ApplicationInteractionContext, sessionID, turnID string) (ApplicationInteractionTurn, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	entry, exists := repository.entries[applicationInteractionSessionRepositoryKey(ctx, sessionID)]
	if repository.unavailable {
		return ApplicationInteractionTurn{}, errApplicationSessionStore
	}
	if !exists {
		return ApplicationInteractionTurn{}, errApplicationSessionNotFound
	}
	for _, turn := range entry.turns {
		if turn.TurnID == turnID {
			if validateStoredApplicationInteractionTurn(ctx, turn) != nil {
				return ApplicationInteractionTurn{}, errApplicationSessionContract
			}
			return cloneApplicationInteractionTurn(turn), nil
		}
	}
	return ApplicationInteractionTurn{}, errApplicationSessionNotFound
}

func (repository *memoryApplicationInteractionSessionRepository) ReadTurnByClientKey(ctx ApplicationInteractionContext, sessionID, clientTurnKey string) (ApplicationInteractionTurn, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	entry, exists := repository.entries[applicationInteractionSessionRepositoryKey(ctx, sessionID)]
	if repository.unavailable {
		return ApplicationInteractionTurn{}, errApplicationSessionStore
	}
	if !exists {
		return ApplicationInteractionTurn{}, errApplicationSessionNotFound
	}
	for _, turn := range entry.turns {
		if turn.ClientTurnKey == clientTurnKey {
			if validateStoredApplicationInteractionTurn(ctx, turn) != nil {
				return ApplicationInteractionTurn{}, errApplicationSessionContract
			}
			return cloneApplicationInteractionTurn(turn), nil
		}
	}
	return ApplicationInteractionTurn{}, errApplicationSessionNotFound
}

func (repository *memoryApplicationInteractionSessionRepository) CompleteTurn(ctx ApplicationInteractionContext, terminal ApplicationInteractionTurn) (ApplicationInteractionSession, ApplicationInteractionTurn, bool, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	key := applicationInteractionSessionRepositoryKey(ctx, terminal.SessionID)
	entry, exists := repository.entries[key]
	if repository.unavailable {
		return ApplicationInteractionSession{}, ApplicationInteractionTurn{}, false, errApplicationSessionStore
	}
	if !exists {
		return ApplicationInteractionSession{}, ApplicationInteractionTurn{}, false, errApplicationSessionNotFound
	}
	for index, current := range entry.turns {
		if current.TurnID != terminal.TurnID {
			continue
		}
		if current.Status != string(WorkflowRunStatusRunning) {
			if applicationInteractionTurnsIdempotentlyEqual(current, terminal) {
				return cloneApplicationInteractionSession(entry.session), cloneApplicationInteractionTurn(current), true, nil
			}
			return ApplicationInteractionSession{}, ApplicationInteractionTurn{}, false, errApplicationSessionIdempotency
		}
		if validateApplicationInteractionTurnTransition(current, terminal) != nil || validateStoredApplicationInteractionTurn(ctx, terminal) != nil {
			return ApplicationInteractionSession{}, ApplicationInteractionTurn{}, false, errApplicationSessionContract
		}
		entry.turns[index] = cloneApplicationInteractionTurn(terminal)
		repository.entries[key] = entry
		return cloneApplicationInteractionSession(entry.session), cloneApplicationInteractionTurn(terminal), false, nil
	}
	return ApplicationInteractionSession{}, ApplicationInteractionTurn{}, false, errApplicationSessionNotFound
}

func (repository *memoryApplicationInteractionSessionRepository) ListTurns(ctx ApplicationInteractionContext, sessionID string) ([]ApplicationInteractionTurn, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	entry, exists := repository.entries[applicationInteractionSessionRepositoryKey(ctx, sessionID)]
	if repository.unavailable {
		return nil, errApplicationSessionStore
	}
	if !exists {
		return nil, errApplicationSessionNotFound
	}
	turns := make([]ApplicationInteractionTurn, 0, len(entry.turns))
	for _, turn := range entry.turns {
		if validateStoredApplicationInteractionTurn(ctx, turn) != nil {
			return nil, errApplicationSessionContract
		}
		turns = append(turns, cloneApplicationInteractionTurn(turn))
	}
	return turns, nil
}

func validateApplicationInteractionContext(ctx ApplicationInteractionContext) error {
	if !applicationDraftIdentifierPattern.MatchString(strings.TrimSpace(ctx.TenantRef)) || !applicationDraftIdentifierPattern.MatchString(strings.TrimSpace(ctx.WorkspaceID)) || !applicationDraftIdentifierPattern.MatchString(strings.TrimSpace(ctx.ApplicationID)) || !applicationDraftIdentifierPattern.MatchString(strings.TrimSpace(ctx.OwnerSubjectRef)) || !applicationDraftIdentifierPattern.MatchString(strings.TrimSpace(ctx.ActorRef)) || strings.TrimSpace(ctx.RequestID) == "" || strings.TrimSpace(ctx.AuditRef) == "" {
		return errApplicationSessionContract
	}
	return nil
}

func validateApplicationInteractionTerminalTurnInput(input ApplicationInteractionTerminalTurnInput) error {
	if validateApplicationInteractionTurnReservationInput(ApplicationInteractionTurnReservationInput{ExpectedSessionVersion: input.ExpectedSessionVersion, ClientTurnKey: input.ClientTurnKey, InputDigest: input.InputDigest, InputBytes: input.InputBytes, StartedAt: input.StartedAt}) != nil || validateApplicationInteractionTurnCompletionInput(ApplicationInteractionTurnCompletionInput{Status: input.Status, RunRef: input.RunRef, FailureCode: input.FailureCode, FailureSummary: input.FailureSummary, CompletedAt: input.CompletedAt}) != nil || input.CompletedAt.Before(input.StartedAt) {
		return errApplicationSessionContract
	}
	return nil
}

func validateApplicationInteractionTurnReservationInput(input ApplicationInteractionTurnReservationInput) error {
	if input.ExpectedSessionVersion < 1 || !applicationDraftIdentifierPattern.MatchString(strings.TrimSpace(input.ClientTurnKey)) || !workflowRAGDigestPattern.MatchString(strings.TrimSpace(input.InputDigest)) || input.InputBytes < 1 || input.InputBytes > workflowExecutorMaxInputBytes || input.StartedAt.IsZero() {
		return errApplicationSessionContract
	}
	return nil
}

func validateApplicationInteractionTurnCompletionInput(input ApplicationInteractionTurnCompletionInput) error {
	status := strings.TrimSpace(input.Status)
	if input.CompletedAt.IsZero() || len(strings.TrimSpace(input.FailureSummary)) > 256 {
		return errApplicationSessionContract
	}
	if status != string(WorkflowRunStatusSucceeded) && status != string(WorkflowRunStatusFailed) && status != string(WorkflowRunStatusCanceled) && status != string(WorkflowRunStatusOutcomeUnknown) {
		return errApplicationSessionContract
	}
	if status == string(WorkflowRunStatusSucceeded) && (input.RunRef == nil || strings.TrimSpace(input.FailureCode) != "" || strings.TrimSpace(input.FailureSummary) != "") {
		return errApplicationSessionContract
	}
	if status != string(WorkflowRunStatusSucceeded) && strings.TrimSpace(input.FailureCode) == "" {
		return errApplicationSessionContract
	}
	return nil
}

func validateStoredApplicationInteractionSession(ctx ApplicationInteractionContext, session ApplicationInteractionSession) error {
	if session.SchemaVersion != applicationSessionSchemaVersion || !applicationSessionIDPattern.MatchString(session.SessionID) || session.TenantRef != ctx.TenantRef || session.WorkspaceID != ctx.WorkspaceID || session.ApplicationID != ctx.ApplicationID || session.OwnerSubjectRef != ctx.OwnerSubjectRef || (session.State != applicationSessionStateActive && session.State != applicationSessionStateClosed) || session.RecordVersion < 1 || session.ContentRetention != applicationSessionRetentionPolicy || session.TurnCount < 0 || validateApplicationInteractionProfileBinding(session.ProfileBinding) != nil || validateApplicationInteractionAuthority(session.Authority) != nil || session.ProfileBinding.ExecutionProfile != session.Authority.ExecutionProfile || session.ProfileBinding.DefinitionID != applicationInteractionAuthorityDefinitionID(session.Authority) || parseApplicationInteractionTimestamp(session.CreatedAt) == nil || parseApplicationInteractionTimestamp(session.UpdatedAt) == nil || strings.TrimSpace(session.CreatedByActorRef) == "" || strings.TrimSpace(session.UpdatedByActorRef) == "" || strings.TrimSpace(session.RequestID) == "" || strings.TrimSpace(session.AuditRef) == "" {
		return errApplicationSessionContract
	}
	if session.State == applicationSessionStateActive && session.ClosedAt != nil || session.State == applicationSessionStateClosed && (session.ClosedAt == nil || parseApplicationInteractionTimestamp(*session.ClosedAt) == nil) || session.TurnCount == 0 && session.LastTurnID != nil || session.TurnCount > 0 && (session.LastTurnID == nil || !applicationTurnIDPattern.MatchString(*session.LastTurnID)) {
		return errApplicationSessionContract
	}
	return nil
}

func validateStoredApplicationInteractionTurn(ctx ApplicationInteractionContext, turn ApplicationInteractionTurn) error {
	if turn.SchemaVersion != applicationSessionTurnSchemaVersion || !applicationTurnIDPattern.MatchString(turn.TurnID) || !applicationSessionIDPattern.MatchString(turn.SessionID) || turn.Sequence < 1 || !applicationDraftIdentifierPattern.MatchString(turn.ClientTurnKey) || turn.TenantRef != ctx.TenantRef || turn.WorkspaceID != ctx.WorkspaceID || turn.ApplicationID != ctx.ApplicationID || turn.OwnerSubjectRef != ctx.OwnerSubjectRef || turn.ExecutionProfile != turn.Authority.ExecutionProfile || validateApplicationInteractionAuthority(turn.Authority) != nil || !workflowRAGDigestPattern.MatchString(turn.InputDigest) || turn.InputBytes < 1 || turn.InputBytes > workflowExecutorMaxInputBytes || parseApplicationInteractionTimestamp(turn.StartedAt) == nil || strings.TrimSpace(turn.ActorRef) == "" || strings.TrimSpace(turn.RequestID) == "" || strings.TrimSpace(turn.AuditRef) == "" || len(turn.FailureSummary) > 256 {
		return errApplicationSessionContract
	}
	if turn.Status == string(WorkflowRunStatusRunning) {
		if turn.CompletedAt != nil || turn.RunRef != nil || turn.FailureCode != "" || turn.FailureSummary != "" {
			return errApplicationSessionContract
		}
		return nil
	}
	if turn.CompletedAt == nil || parseApplicationInteractionTimestamp(*turn.CompletedAt) == nil {
		return errApplicationSessionContract
	}
	if turn.Status == string(WorkflowRunStatusSucceeded) {
		if turn.RunRef == nil || turn.FailureCode != "" || turn.FailureSummary != "" {
			return errApplicationSessionContract
		}
	} else if turn.Status != string(WorkflowRunStatusFailed) && turn.Status != string(WorkflowRunStatusCanceled) && turn.Status != string(WorkflowRunStatusOutcomeUnknown) {
		return errApplicationSessionContract
	} else if strings.TrimSpace(turn.FailureCode) == "" {
		return errApplicationSessionContract
	}
	return validateApplicationInteractionRunRef(turn.ExecutionProfile, turn.RunRef)
}

func validateApplicationInteractionRunRef(profile string, ref *ApplicationInteractionRunRef) error {
	if ref == nil {
		return nil
	}
	if !workflowRAGRunIDPattern.MatchString(strings.TrimSpace(ref.RunID)) {
		return errApplicationSessionContract
	}
	if profile == applicationInteractionProfileWorkflow && ref.SchemaVersion != workflowRunRecordDefinitionSchemaVersion || profile == applicationInteractionProfileRAG && ref.SchemaVersion != workflowRunRecordAppRAGSchemaVersion {
		return errApplicationSessionContract
	}
	return nil
}

func validateApplicationInteractionSessionTransition(before, after ApplicationInteractionSession) error {
	if before.State != applicationSessionStateActive || after.State != applicationSessionStateClosed || after.RecordVersion != before.RecordVersion+1 || before.SessionID != after.SessionID || before.TurnCount != after.TurnCount || !applicationInteractionAuthoritiesEqual(before.Authority, after.Authority) {
		return errApplicationSessionContract
	}
	return nil
}

func validateApplicationInteractionSessionReserve(before, after ApplicationInteractionSession, turn ApplicationInteractionTurn) error {
	if before.State != applicationSessionStateActive || after.State != applicationSessionStateActive || after.RecordVersion != before.RecordVersion+1 || after.TurnCount != before.TurnCount+1 || after.LastTurnID == nil || *after.LastTurnID != turn.TurnID || turn.Sequence != after.TurnCount || turn.SessionID != before.SessionID || !applicationInteractionAuthoritiesEqual(before.Authority, after.Authority) || !applicationInteractionAuthoritiesEqual(after.Authority, turn.Authority) {
		return errApplicationSessionContract
	}
	return nil
}

func validateApplicationInteractionTurnTransition(before, after ApplicationInteractionTurn) error {
	if before.Status != string(WorkflowRunStatusRunning) || after.Status == string(WorkflowRunStatusRunning) || before.TurnID != after.TurnID || before.SessionID != after.SessionID || before.Sequence != after.Sequence || before.ClientTurnKey != after.ClientTurnKey || before.InputDigest != after.InputDigest || before.InputBytes != after.InputBytes || before.StartedAt != after.StartedAt || !applicationInteractionAuthoritiesEqual(before.Authority, after.Authority) {
		return errApplicationSessionContract
	}
	return nil
}

func applicationInteractionSessionRepositoryKey(ctx ApplicationInteractionContext, sessionID string) string {
	return strings.Join([]string{ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, sessionID}, "\x00")
}

func normalizeApplicationInteractionProfileBinding(binding ApplicationInteractionProfileBinding) ApplicationInteractionProfileBinding {
	binding.ExecutionProfile = strings.TrimSpace(binding.ExecutionProfile)
	binding.DefinitionID = strings.TrimSpace(binding.DefinitionID)
	return binding
}

func applicationInteractionAuthorityDefinitionID(authority ApplicationInteractionAuthoritySnapshot) string {
	if authority.WorkflowDefinition == nil {
		return ""
	}
	return authority.WorkflowDefinition.DefinitionID
}

func applicationInteractionAuthoritiesEqual(left, right ApplicationInteractionAuthoritySnapshot) bool {
	return left.AuthorityDigest != "" && left.AuthorityDigest == right.AuthorityDigest
}

func applicationInteractionTurnsIdempotentlyEqual(left, right ApplicationInteractionTurn) bool {
	return left.SessionID == right.SessionID && left.ClientTurnKey == right.ClientTurnKey && left.ExecutionProfile == right.ExecutionProfile && left.Authority.AuthorityDigest == right.Authority.AuthorityDigest && left.Status == right.Status && left.InputDigest == right.InputDigest && left.InputBytes == right.InputBytes && applicationInteractionRunRefsEqual(left.RunRef, right.RunRef) && left.FailureCode == right.FailureCode && left.FailureSummary == right.FailureSummary
}

func applicationInteractionTurnReservationsEqual(left, right ApplicationInteractionTurn) bool {
	return left.SessionID == right.SessionID && left.ClientTurnKey == right.ClientTurnKey && left.ExecutionProfile == right.ExecutionProfile && left.Authority.AuthorityDigest == right.Authority.AuthorityDigest && left.InputDigest == right.InputDigest && left.InputBytes == right.InputBytes
}

func applicationInteractionRunRefsEqual(left, right *ApplicationInteractionRunRef) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func newApplicationInteractionID(prefix string) (string, error) {
	raw := make([]byte, 10)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return prefix + "_" + strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw)), nil
}

func encodeApplicationInteractionSessionCursor(cursor applicationInteractionSessionCursor) (string, error) {
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeApplicationInteractionSessionCursor(value string) (applicationInteractionSessionCursor, error) {
	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil {
		return applicationInteractionSessionCursor{}, err
	}
	var cursor applicationInteractionSessionCursor
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cursor); err != nil {
		return applicationInteractionSessionCursor{}, err
	}
	return cursor, nil
}

func validateApplicationInteractionContractJSON(contract string, payload []byte) error {
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	switch contract {
	case applicationSessionSchemaVersion:
		var session ApplicationInteractionSession
		if err := decoder.Decode(&session); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
			return errApplicationSessionContract
		}
		ctx := ApplicationInteractionContext{TenantRef: session.TenantRef, WorkspaceID: session.WorkspaceID, ApplicationID: session.ApplicationID, OwnerSubjectRef: session.OwnerSubjectRef}
		return validateStoredApplicationInteractionSession(ctx, session)
	case applicationSessionTurnSchemaVersion:
		var turn ApplicationInteractionTurn
		if err := decoder.Decode(&turn); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
			return errApplicationSessionContract
		}
		ctx := ApplicationInteractionContext{TenantRef: turn.TenantRef, WorkspaceID: turn.WorkspaceID, ApplicationID: turn.ApplicationID, OwnerSubjectRef: turn.OwnerSubjectRef}
		return validateStoredApplicationInteractionTurn(ctx, turn)
	default:
		return errApplicationSessionContract
	}
}

func applicationInteractionSessionFailure(code string) ApplicationInteractionSessionResult {
	return ApplicationInteractionSessionResult{FailureCode: code}
}

func applicationInteractionRepositoryFailure(err error) ApplicationInteractionSessionResult {
	result := applicationInteractionSessionFailure(ApplicationInteractionFailureStoreUnavailable)
	switch {
	case errors.Is(err, errApplicationSessionNotFound):
		result.FailureCode = ApplicationInteractionFailureSessionNotFound
	case errors.Is(err, errApplicationSessionClosed):
		result.FailureCode = ApplicationInteractionFailureSessionClosed
	case errors.Is(err, errApplicationSessionVersionConflict):
		result.FailureCode = ApplicationInteractionFailureVersionConflict
	case errors.Is(err, errApplicationSessionIdempotency):
		result.FailureCode = ApplicationInteractionFailureIdempotencyConflict
	case errors.Is(err, errApplicationSessionContract):
		result.FailureCode = ApplicationInteractionFailureStoreContract
	}
	return result
}

func parseApplicationInteractionTimestamp(value string) *time.Time {
	parsed, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(value))
	if err != nil {
		return nil
	}
	return &parsed
}

func cloneApplicationInteractionSession(value ApplicationInteractionSession) ApplicationInteractionSession {
	copy := value
	if value.LastTurnID != nil {
		last := *value.LastTurnID
		copy.LastTurnID = &last
	}
	if value.ClosedAt != nil {
		closed := *value.ClosedAt
		copy.ClosedAt = &closed
	}
	copy.Authority = cloneApplicationInteractionAuthority(value.Authority)
	return copy
}

func cloneApplicationInteractionTurn(value ApplicationInteractionTurn) ApplicationInteractionTurn {
	copy := value
	copy.Authority = cloneApplicationInteractionAuthority(value.Authority)
	copy.RunRef = cloneApplicationInteractionRunRef(value.RunRef)
	if value.CompletedAt != nil {
		completed := *value.CompletedAt
		copy.CompletedAt = &completed
	}
	return copy
}

func cloneApplicationInteractionAuthority(value ApplicationInteractionAuthoritySnapshot) ApplicationInteractionAuthoritySnapshot {
	copy := value
	if value.WorkflowDefinition != nil {
		workflow := *value.WorkflowDefinition
		copy.WorkflowDefinition = &workflow
	}
	if value.ApplicationRAG != nil {
		rag := *value.ApplicationRAG
		copy.ApplicationRAG = &rag
	}
	return copy
}

func cloneApplicationInteractionRunRef(value *ApplicationInteractionRunRef) *ApplicationInteractionRunRef {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

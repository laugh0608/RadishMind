package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"
)

type localIdentitySelfServiceSessionCursor struct {
	SchemaVersion string `json:"schema_version"`
	UserID        string `json:"user_id"`
	State         string `json:"state"`
	Limit         int    `json:"limit"`
	SnapshotAt    string `json:"snapshot_at"`
	CreatedAt     string `json:"created_at"`
	SessionID     string `json:"session_id"`
	BindingDigest string `json:"binding_digest"`
}

func (repository *memoryLocalIdentityRepository) ListSelfServiceSessions(
	_ context.Context,
	userID string,
	currentSessionID string,
	query LocalIdentitySelfServiceSessionListQuery,
) (LocalIdentitySelfServiceSessionPage, error) {
	if repository == nil {
		return LocalIdentitySelfServiceSessionPage{}, errLocalIdentityStoreUnavailable
	}
	filter, cursor, err := normalizeLocalIdentitySelfServiceSessionQuery(userID, query)
	if err != nil {
		return LocalIdentitySelfServiceSessionPage{}, err
	}
	currentSessionID = strings.TrimSpace(currentSessionID)
	if !localSessionIDPattern.MatchString(currentSessionID) {
		return LocalIdentitySelfServiceSessionPage{}, errLocalIdentitySessionScopeDenied
	}
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if err := repository.validateSelfServiceActorLocked(userID, currentSessionID, filter.requestedAt); err != nil {
		return LocalIdentitySelfServiceSessionPage{}, err
	}
	sessions := make([]WebSession, 0)
	for _, session := range repository.sessions {
		if session.UserID != userID || session.CreatedAt.After(filter.snapshotAt) {
			continue
		}
		effectiveState := localIdentitySessionEffectiveStateAt(session, filter.snapshotAt)
		if filter.State != localIdentitySessionStateFilterAll && filter.State != effectiveState {
			continue
		}
		if cursor.SessionID != "" && !localIdentitySessionComesAfterCursor(session, cursor) {
			continue
		}
		sessions = append(sessions, cloneWebSession(session))
	}
	slices.SortFunc(sessions, func(left, right WebSession) int {
		if !left.CreatedAt.Equal(right.CreatedAt) {
			return right.CreatedAt.Compare(left.CreatedAt)
		}
		return strings.Compare(right.SessionID, left.SessionID)
	})
	hasNext := len(sessions) > filter.Limit
	if hasNext {
		sessions = sessions[:filter.Limit]
	}
	page := LocalIdentitySelfServiceSessionPage{
		Sessions:   make([]LocalIdentitySelfServiceSessionSummary, 0, len(sessions)),
		SnapshotAt: filter.snapshotAt,
	}
	for _, session := range sessions {
		if !validWebSession(session) {
			return LocalIdentitySelfServiceSessionPage{}, errLocalIdentityStoreUnavailable
		}
		page.Sessions = append(page.Sessions, localIdentitySelfServiceSessionSummary(
			session,
			currentSessionID,
			filter.snapshotAt,
		))
	}
	if hasNext && len(sessions) > 0 {
		page.NextCursor, err = encodeLocalIdentitySelfServiceSessionCursor(userID, filter, sessions[len(sessions)-1])
		if err != nil {
			return LocalIdentitySelfServiceSessionPage{}, errLocalIdentityStoreUnavailable
		}
	}
	return page, nil
}

func (repository *memoryLocalIdentityRepository) RevokeOwnedWebSession(
	_ context.Context,
	mutation localIdentityOwnedSessionRevocation,
) (LocalIdentitySelfServiceSessionRevocation, error) {
	if repository == nil {
		return LocalIdentitySelfServiceSessionRevocation{}, errLocalIdentityStoreUnavailable
	}
	mutation.UserID = strings.TrimSpace(mutation.UserID)
	mutation.CurrentSessionID = strings.TrimSpace(mutation.CurrentSessionID)
	mutation.TargetSessionID = strings.TrimSpace(mutation.TargetSessionID)
	mutation.AuditRef = strings.TrimSpace(mutation.AuditRef)
	mutation.RevokedAt = mutation.RevokedAt.UTC()
	if !localUserIDPattern.MatchString(mutation.UserID) || !localSessionIDPattern.MatchString(mutation.CurrentSessionID) ||
		!localSessionIDPattern.MatchString(mutation.TargetSessionID) || mutation.ExpectedVersion < 1 ||
		mutation.RevokedAt.IsZero() || !validAuditRef(mutation.AuditRef) {
		return LocalIdentitySelfServiceSessionRevocation{}, errLocalIdentityContractMismatch
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if err := repository.validateSelfServiceActorLocked(mutation.UserID, mutation.CurrentSessionID, mutation.RevokedAt); err != nil {
		return LocalIdentitySelfServiceSessionRevocation{}, err
	}
	target, exists := repository.sessions[mutation.TargetSessionID]
	if !exists || target.UserID != mutation.UserID {
		return LocalIdentitySelfServiceSessionRevocation{}, errLocalIdentitySessionScopeDenied
	}
	if target.RecordVersion != mutation.ExpectedVersion || target.LifecycleState != localIdentityStateActive ||
		!target.ExpiresAt.After(mutation.RevokedAt) {
		return LocalIdentitySelfServiceSessionRevocation{}, errLocalIdentitySessionVersionConflict
	}
	revoked, err := revokeLocalIdentitySessionCandidate(target, mutation.RevokedAt, mutation.AuditRef)
	if err != nil {
		return LocalIdentitySelfServiceSessionRevocation{}, errLocalIdentityStoreUnavailable
	}
	repository.sessions[target.SessionID] = revoked
	return LocalIdentitySelfServiceSessionRevocation{
		SchemaVersion: localIdentitySelfServiceSessionRevocationSchemaVersion,
		Session: localIdentitySelfServiceSessionSummary(
			revoked,
			mutation.CurrentSessionID,
			mutation.RevokedAt,
		),
		CurrentSessionRevoked: revoked.SessionID == mutation.CurrentSessionID,
	}, nil
}

func (repository *memoryLocalIdentityRepository) RevokeOtherWebSessions(
	_ context.Context,
	mutation localIdentityOtherSessionRevocation,
) (LocalIdentitySelfServiceBulkSessionRevocation, error) {
	if repository == nil {
		return LocalIdentitySelfServiceBulkSessionRevocation{}, errLocalIdentityStoreUnavailable
	}
	mutation.UserID = strings.TrimSpace(mutation.UserID)
	mutation.CurrentSessionID = strings.TrimSpace(mutation.CurrentSessionID)
	mutation.AuditRef = strings.TrimSpace(mutation.AuditRef)
	mutation.RevokedAt = mutation.RevokedAt.UTC()
	if !localUserIDPattern.MatchString(mutation.UserID) || !localSessionIDPattern.MatchString(mutation.CurrentSessionID) ||
		mutation.RevokedAt.IsZero() || !validAuditRef(mutation.AuditRef) {
		return LocalIdentitySelfServiceBulkSessionRevocation{}, errLocalIdentityContractMismatch
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if err := repository.validateSelfServiceActorLocked(mutation.UserID, mutation.CurrentSessionID, mutation.RevokedAt); err != nil {
		return LocalIdentitySelfServiceBulkSessionRevocation{}, err
	}
	updates := make(map[string]WebSession)
	for sessionID, session := range repository.sessions {
		if session.UserID != mutation.UserID || sessionID == mutation.CurrentSessionID ||
			session.LifecycleState != localIdentityStateActive || !session.ExpiresAt.After(mutation.RevokedAt) {
			continue
		}
		if !validWebSession(session) {
			return LocalIdentitySelfServiceBulkSessionRevocation{}, errLocalIdentitySessionBulkRevokeConflict
		}
		revoked, err := revokeLocalIdentitySessionCandidate(session, mutation.RevokedAt, mutation.AuditRef)
		if err != nil {
			return LocalIdentitySelfServiceBulkSessionRevocation{}, errLocalIdentitySessionBulkRevokeConflict
		}
		updates[sessionID] = revoked
	}
	for sessionID, session := range updates {
		repository.sessions[sessionID] = session
	}
	return LocalIdentitySelfServiceBulkSessionRevocation{
		SchemaVersion: localIdentitySelfServiceBulkRevocationSchemaVersion,
		RevokedCount:  len(updates),
	}, nil
}

func (repository *memoryLocalIdentityRepository) RotateLocalCredentialAndRevokeSessions(
	_ context.Context,
	mutation localIdentityCredentialRotation,
) (LocalIdentitySelfServiceCredentialRotation, error) {
	if repository == nil {
		return LocalIdentitySelfServiceCredentialRotation{}, errLocalIdentityStoreUnavailable
	}
	mutation.UserID = strings.TrimSpace(mutation.UserID)
	mutation.CurrentSessionID = strings.TrimSpace(mutation.CurrentSessionID)
	mutation.AuditRef = strings.TrimSpace(mutation.AuditRef)
	mutation.RotatedAt = mutation.RotatedAt.UTC()
	if !localUserIDPattern.MatchString(mutation.UserID) || !localSessionIDPattern.MatchString(mutation.CurrentSessionID) ||
		mutation.CurrentPassword == "" || len(mutation.CurrentPassword) > 1024 || !validLocalIdentityPassword(mutation.NewPassword) ||
		mutation.RotatedAt.IsZero() || !validAuditRef(mutation.AuditRef) || !validLocalCredential(mutation.Replacement) ||
		mutation.Replacement.UserID != mutation.UserID || mutation.Replacement.LifecycleState != localIdentityStateActive ||
		!mutation.Replacement.CreatedAt.Equal(mutation.RotatedAt) || mutation.Replacement.AuditRef != mutation.AuditRef ||
		!VerifyLocalPassword(mutation.NewPassword, mutation.Replacement) {
		return LocalIdentitySelfServiceCredentialRotation{}, errLocalIdentityContractMismatch
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if err := repository.validateSelfServiceActorLocked(mutation.UserID, mutation.CurrentSessionID, mutation.RotatedAt); err != nil {
		return LocalIdentitySelfServiceCredentialRotation{}, err
	}
	currentID := repository.activeCredentialByUser[mutation.UserID]
	current, exists := repository.credentials[currentID]
	if !exists || current.LifecycleState != localIdentityStateActive || current.UserID != mutation.UserID {
		return LocalIdentitySelfServiceCredentialRotation{}, errLocalIdentityCredentialUnavailable
	}
	if !VerifyLocalPassword(mutation.CurrentPassword, current) {
		return LocalIdentitySelfServiceCredentialRotation{}, errLocalIdentityCredentialCurrentInvalid
	}
	if VerifyLocalPassword(mutation.NewPassword, current) {
		return LocalIdentitySelfServiceCredentialRotation{}, errLocalIdentityCredentialReuseDenied
	}
	if _, collision := repository.credentials[mutation.Replacement.CredentialID]; collision ||
		!mutation.Replacement.CreatedAt.After(current.CreatedAt) {
		return LocalIdentitySelfServiceCredentialRotation{}, errLocalIdentityCredentialRotationConflict
	}
	superseded := cloneLocalCredential(current)
	superseded.LifecycleState = localIdentityStateSuperseded
	superseded.RecordVersion++
	superseded.UpdatedAt = mutation.RotatedAt
	superseded.AuditRef = mutation.AuditRef
	if !validLocalCredential(superseded) {
		return LocalIdentitySelfServiceCredentialRotation{}, errLocalIdentityCredentialRotationConflict
	}
	sourceRef := "credential:" + current.CredentialID
	updates := make(map[string]WebSession)
	currentSessionRevoked := false
	for sessionID, session := range repository.sessions {
		if session.UserID != mutation.UserID || session.LifecycleState != localIdentityStateActive ||
			session.AuthenticationMethod != localAuthenticationMethodPassword || session.AuthenticationSourceRef != sourceRef {
			continue
		}
		if !validWebSession(session) {
			return LocalIdentitySelfServiceCredentialRotation{}, errLocalIdentityCredentialRotationConflict
		}
		revoked, err := revokeLocalIdentitySessionCandidate(session, mutation.RotatedAt, mutation.AuditRef)
		if err != nil {
			return LocalIdentitySelfServiceCredentialRotation{}, errLocalIdentityCredentialRotationConflict
		}
		updates[sessionID] = revoked
		currentSessionRevoked = currentSessionRevoked || sessionID == mutation.CurrentSessionID
	}
	repository.credentials[currentID] = superseded
	repository.credentials[mutation.Replacement.CredentialID] = cloneLocalCredential(mutation.Replacement)
	repository.activeCredentialByUser[mutation.UserID] = mutation.Replacement.CredentialID
	for sessionID, session := range updates {
		repository.sessions[sessionID] = session
	}
	return LocalIdentitySelfServiceCredentialRotation{
		SchemaVersion:         localIdentitySelfServiceCredentialRotationSchemaVersion,
		PolicyVersion:         mutation.Replacement.PolicyVersion,
		RevokedSessionCount:   len(updates),
		CurrentSessionRevoked: currentSessionRevoked,
	}, nil
}

func (repository *memoryLocalIdentityRepository) validateSelfServiceActorLocked(
	userID string,
	currentSessionID string,
	asOf time.Time,
) error {
	account, accountExists := repository.accounts[userID]
	session, sessionExists := repository.sessions[currentSessionID]
	if !accountExists || !sessionExists || session.UserID != userID || account.LifecycleState != localIdentityStateActive ||
		session.LifecycleState != localIdentityStateActive || !session.ExpiresAt.After(asOf) ||
		!validUserAccount(account) || !validWebSession(session) {
		return errLocalIdentitySessionScopeDenied
	}
	return nil
}

func normalizeLocalIdentitySelfServiceSessionQuery(
	userID string,
	query LocalIdentitySelfServiceSessionListQuery,
) (LocalIdentitySelfServiceSessionListQuery, localIdentitySelfServiceSessionCursor, error) {
	userID = strings.TrimSpace(userID)
	query.State = strings.TrimSpace(query.State)
	query.Cursor = strings.TrimSpace(query.Cursor)
	query.requestedAt = query.requestedAt.UTC()
	query.snapshotAt = query.snapshotAt.UTC()
	if query.State == "" {
		query.State = localIdentityStateActive
	}
	if query.Limit == 0 {
		query.Limit = localIdentitySelfServiceSessionDefaultLimit
	}
	if !localUserIDPattern.MatchString(userID) || !validLocalIdentitySelfServiceSessionState(query.State) ||
		query.Limit < 1 || query.Limit > localIdentitySelfServiceSessionMaximumLimit || query.requestedAt.IsZero() {
		return LocalIdentitySelfServiceSessionListQuery{}, localIdentitySelfServiceSessionCursor{}, errLocalIdentitySessionCursorInvalid
	}
	if query.Cursor == "" {
		if query.snapshotAt.IsZero() || query.snapshotAt.After(query.requestedAt) {
			return LocalIdentitySelfServiceSessionListQuery{}, localIdentitySelfServiceSessionCursor{}, errLocalIdentitySessionCursorInvalid
		}
		return query, localIdentitySelfServiceSessionCursor{}, nil
	}
	cursor, err := decodeLocalIdentitySelfServiceSessionCursor(userID, query)
	if err != nil {
		return LocalIdentitySelfServiceSessionListQuery{}, localIdentitySelfServiceSessionCursor{}, errLocalIdentitySessionCursorInvalid
	}
	snapshotAt, _ := time.Parse(time.RFC3339Nano, cursor.SnapshotAt)
	if snapshotAt.After(query.requestedAt) {
		return LocalIdentitySelfServiceSessionListQuery{}, localIdentitySelfServiceSessionCursor{}, errLocalIdentitySessionCursorInvalid
	}
	query.snapshotAt = snapshotAt
	return query, cursor, nil
}

func encodeLocalIdentitySelfServiceSessionCursor(
	userID string,
	query LocalIdentitySelfServiceSessionListQuery,
	last WebSession,
) (string, error) {
	document := localIdentitySelfServiceSessionCursor{
		SchemaVersion: localIdentitySelfServiceSessionCursorSchemaVersion,
		UserID:        userID,
		State:         query.State,
		Limit:         query.Limit,
		SnapshotAt:    query.snapshotAt.UTC().Format(time.RFC3339Nano),
		CreatedAt:     last.CreatedAt.UTC().Format(time.RFC3339Nano),
		SessionID:     last.SessionID,
	}
	document.BindingDigest = localIdentitySelfServiceSessionCursorDigest(document)
	payload, err := json.Marshal(document)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeLocalIdentitySelfServiceSessionCursor(
	userID string,
	query LocalIdentitySelfServiceSessionListQuery,
) (localIdentitySelfServiceSessionCursor, error) {
	if len(query.Cursor) > 2048 {
		return localIdentitySelfServiceSessionCursor{}, errLocalIdentitySessionCursorInvalid
	}
	payload, err := base64.RawURLEncoding.DecodeString(query.Cursor)
	if err != nil {
		return localIdentitySelfServiceSessionCursor{}, errLocalIdentitySessionCursorInvalid
	}
	var document localIdentitySelfServiceSessionCursor
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&document) != nil || decoder.Decode(&struct{}{}) != io.EOF ||
		document.SchemaVersion != localIdentitySelfServiceSessionCursorSchemaVersion || document.UserID != userID ||
		document.State != query.State || document.Limit != query.Limit || !localSessionIDPattern.MatchString(document.SessionID) ||
		document.BindingDigest != localIdentitySelfServiceSessionCursorDigest(document) {
		return localIdentitySelfServiceSessionCursor{}, errLocalIdentitySessionCursorInvalid
	}
	if !validCanonicalUTCTime(document.SnapshotAt) || !validCanonicalUTCTime(document.CreatedAt) {
		return localIdentitySelfServiceSessionCursor{}, errLocalIdentitySessionCursorInvalid
	}
	createdAt, _ := time.Parse(time.RFC3339Nano, document.CreatedAt)
	snapshotAt, _ := time.Parse(time.RFC3339Nano, document.SnapshotAt)
	if createdAt.After(snapshotAt) {
		return localIdentitySelfServiceSessionCursor{}, errLocalIdentitySessionCursorInvalid
	}
	return document, nil
}

func localIdentitySelfServiceSessionCursorDigest(document localIdentitySelfServiceSessionCursor) string {
	return localIdentityDigest(
		localIdentitySelfServiceSessionCursorSchemaVersion,
		document.UserID,
		document.State,
		strconv.Itoa(document.Limit),
		document.SnapshotAt,
		document.CreatedAt,
		document.SessionID,
	)
}

func localIdentitySessionComesAfterCursor(session WebSession, cursor localIdentitySelfServiceSessionCursor) bool {
	anchor, err := time.Parse(time.RFC3339Nano, cursor.CreatedAt)
	if err != nil {
		return false
	}
	return session.CreatedAt.Before(anchor) || session.CreatedAt.Equal(anchor) && session.SessionID < cursor.SessionID
}

func localIdentitySessionEffectiveStateAt(session WebSession, snapshotAt time.Time) string {
	if session.RevokedAt != nil && !session.RevokedAt.After(snapshotAt) {
		return localIdentityStateRevoked
	}
	if !session.ExpiresAt.After(snapshotAt) {
		return localIdentitySessionEffectiveStateExpired
	}
	return localIdentityStateActive
}

func localIdentitySelfServiceSessionSummary(
	session WebSession,
	currentSessionID string,
	snapshotAt time.Time,
) LocalIdentitySelfServiceSessionSummary {
	effectiveState := localIdentitySessionEffectiveStateAt(session, snapshotAt)
	var revokedAt *time.Time
	if effectiveState == localIdentityStateRevoked {
		revokedAt = cloneTimePointer(session.RevokedAt)
	}
	return LocalIdentitySelfServiceSessionSummary{
		SchemaVersion:        localIdentitySelfServiceSessionSummarySchemaVersion,
		SessionID:            session.SessionID,
		AuthenticationMethod: session.AuthenticationMethod,
		EffectiveState:       effectiveState,
		RecordVersion:        session.RecordVersion,
		CurrentSession:       session.SessionID == currentSessionID,
		CreatedAt:            session.CreatedAt,
		LastVerifiedAt:       session.LastVerifiedAt,
		ExpiresAt:            session.ExpiresAt,
		RevokedAt:            revokedAt,
	}
}

func revokeLocalIdentitySessionCandidate(session WebSession, revokedAt time.Time, auditRef string) (WebSession, error) {
	if !validWebSession(session) || session.LifecycleState != localIdentityStateActive || revokedAt.Before(session.CreatedAt) {
		return WebSession{}, errLocalIdentityContractMismatch
	}
	revoked := cloneWebSession(session)
	revoked.LifecycleState = localIdentityStateRevoked
	revoked.RecordVersion++
	revoked.UpdatedAt = revokedAt.UTC()
	revoked.RevokedAt = timePointer(revokedAt)
	revoked.AuditRef = strings.TrimSpace(auditRef)
	if !validWebSession(revoked) {
		return WebSession{}, errLocalIdentityContractMismatch
	}
	return revoked, nil
}

func validLocalIdentitySelfServiceSessionState(state string) bool {
	return state == localIdentityStateActive || state == localIdentitySessionEffectiveStateExpired ||
		state == localIdentityStateRevoked || state == localIdentitySessionStateFilterAll
}

func validCanonicalUTCTime(value string) bool {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	return err == nil && parsed.Location() == time.UTC && parsed.Format(time.RFC3339Nano) == value
}

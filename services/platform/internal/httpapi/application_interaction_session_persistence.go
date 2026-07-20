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
	payload, err := json.Marshal(session)
	if err != nil {
		return nil, errApplicationSessionContract
	}
	return payload, nil
}

func decodeApplicationInteractionSession(ctx ApplicationInteractionContext, payload []byte) (ApplicationInteractionSession, error) {
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
	payload, err := json.Marshal(turn)
	if err != nil {
		return nil, errApplicationSessionContract
	}
	return payload, nil
}

func decodeApplicationInteractionTurn(ctx ApplicationInteractionContext, payload []byte) (ApplicationInteractionTurn, error) {
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

package httpapi

import (
	"errors"
	"sort"
)

type combinedApplicationInteractionSessionRepository struct {
	legacy applicationInteractionSessionRepository
	prompt applicationInteractionSessionRepository
}

func newCombinedApplicationInteractionSessionRepository(legacy, prompt applicationInteractionSessionRepository) applicationInteractionSessionRepository {
	return &combinedApplicationInteractionSessionRepository{legacy: legacy, prompt: prompt}
}

func newPromptApplicationSessionRepositoryForLegacy(repository applicationInteractionSessionRepository) (applicationInteractionSessionRepository, error) {
	switch typed := repository.(type) {
	case *memoryApplicationInteractionSessionRepository:
		return newMemoryApplicationInteractionSessionRepository(), nil
	case *sqliteApplicationInteractionSessionRepository:
		if typed.database == nil {
			return nil, errors.New("Prompt application sessions require the shared SQLite database")
		}
		return newSQLitePromptApplicationSessionRepository(typed.database), nil
	case *postgresApplicationInteractionSessionRepository:
		if typed.pool == nil {
			return nil, errors.New("Prompt application sessions require the Workflow PostgreSQL pool")
		}
		return newPostgresPromptApplicationSessionRepository(typed.pool), nil
	default:
		return nil, errors.New("Prompt application sessions require a supported Workflow runtime backend")
	}
}

func (repository *combinedApplicationInteractionSessionRepository) Create(ctx ApplicationInteractionContext, session ApplicationInteractionSession) (ApplicationInteractionSession, error) {
	if session.SchemaVersion == promptApplicationSessionV2Schema {
		return repository.prompt.Create(ctx, session)
	}
	return repository.legacy.Create(ctx, session)
}

func (repository *combinedApplicationInteractionSessionRepository) Read(ctx ApplicationInteractionContext, sessionID string) (ApplicationInteractionSession, error) {
	value, err := repository.legacy.Read(ctx, sessionID)
	if !errors.Is(err, errApplicationSessionNotFound) {
		return value, err
	}
	return repository.prompt.Read(ctx, sessionID)
}

func (repository *combinedApplicationInteractionSessionRepository) List(ctx ApplicationInteractionContext, query applicationInteractionSessionListQuery) ([]ApplicationInteractionSession, error) {
	legacy, err := repository.legacy.List(ctx, query)
	if err != nil {
		return nil, err
	}
	prompt, err := repository.prompt.List(ctx, query)
	if err != nil {
		return nil, err
	}
	values := append(legacy, prompt...)
	sort.Slice(values, func(left, right int) bool {
		if values[left].UpdatedAt == values[right].UpdatedAt {
			return values[left].SessionID > values[right].SessionID
		}
		return values[left].UpdatedAt > values[right].UpdatedAt
	})
	if len(values) > query.Limit {
		values = values[:query.Limit]
	}
	return values, nil
}

func (repository *combinedApplicationInteractionSessionRepository) Close(ctx ApplicationInteractionContext, sessionID string, expectedVersion int, updated ApplicationInteractionSession) (ApplicationInteractionSession, error) {
	if updated.SchemaVersion == promptApplicationSessionV2Schema {
		return repository.prompt.Close(ctx, sessionID, expectedVersion, updated)
	}
	return repository.legacy.Close(ctx, sessionID, expectedVersion, updated)
}

func (repository *combinedApplicationInteractionSessionRepository) ReserveTurn(ctx ApplicationInteractionContext, expectedVersion int, updated ApplicationInteractionSession, turn ApplicationInteractionTurn) (ApplicationInteractionSession, ApplicationInteractionTurn, bool, error) {
	if turn.SchemaVersion == promptApplicationSessionTurnV2Schema {
		return repository.prompt.ReserveTurn(ctx, expectedVersion, updated, turn)
	}
	return repository.legacy.ReserveTurn(ctx, expectedVersion, updated, turn)
}

func (repository *combinedApplicationInteractionSessionRepository) ReadTurn(ctx ApplicationInteractionContext, sessionID, turnID string) (ApplicationInteractionTurn, error) {
	value, err := repository.legacy.ReadTurn(ctx, sessionID, turnID)
	if !errors.Is(err, errApplicationSessionNotFound) {
		return value, err
	}
	return repository.prompt.ReadTurn(ctx, sessionID, turnID)
}

func (repository *combinedApplicationInteractionSessionRepository) ReadTurnByClientKey(ctx ApplicationInteractionContext, sessionID, clientKey string) (ApplicationInteractionTurn, error) {
	value, err := repository.legacy.ReadTurnByClientKey(ctx, sessionID, clientKey)
	if !errors.Is(err, errApplicationSessionNotFound) {
		return value, err
	}
	return repository.prompt.ReadTurnByClientKey(ctx, sessionID, clientKey)
}

func (repository *combinedApplicationInteractionSessionRepository) CompleteTurn(ctx ApplicationInteractionContext, turn ApplicationInteractionTurn) (ApplicationInteractionSession, ApplicationInteractionTurn, bool, error) {
	if turn.SchemaVersion == promptApplicationSessionTurnV2Schema {
		return repository.prompt.CompleteTurn(ctx, turn)
	}
	return repository.legacy.CompleteTurn(ctx, turn)
}

func (repository *combinedApplicationInteractionSessionRepository) ListTurns(ctx ApplicationInteractionContext, sessionID string) ([]ApplicationInteractionTurn, error) {
	values, err := repository.legacy.ListTurns(ctx, sessionID)
	if !errors.Is(err, errApplicationSessionNotFound) {
		return values, err
	}
	return repository.prompt.ListTurns(ctx, sessionID)
}

var _ applicationInteractionSessionRepository = (*combinedApplicationInteractionSessionRepository)(nil)

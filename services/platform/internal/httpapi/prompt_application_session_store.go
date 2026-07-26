package httpapi

import (
	"errors"
	"sort"
)

type combinedApplicationInteractionSessionRepository struct {
	legacy applicationInteractionSessionRepository
	prompt applicationInteractionSessionRepository
	agent  applicationInteractionSessionRepository
}

func newCombinedApplicationInteractionSessionRepository(legacy, prompt applicationInteractionSessionRepository) applicationInteractionSessionRepository {
	return &combinedApplicationInteractionSessionRepository{legacy: legacy, prompt: prompt}
}

func newCombinedApplicationInteractionSessionRepositoryWithAgent(legacy, prompt, agent applicationInteractionSessionRepository) applicationInteractionSessionRepository {
	return &combinedApplicationInteractionSessionRepository{legacy: legacy, prompt: prompt, agent: agent}
}

func newAgentCopilotSessionRepositoryForLegacy(repository applicationInteractionSessionRepository) (applicationInteractionSessionRepository, error) {
	switch typed := repository.(type) {
	case *memoryApplicationInteractionSessionRepository:
		return newMemoryApplicationInteractionSessionRepository(), nil
	case *sqliteApplicationInteractionSessionRepository:
		if typed.database == nil {
			return nil, errors.New("Agent Copilot sessions require the shared SQLite database")
		}
		return &sqliteAgentCopilotSessionRepository{
			database: typed.database, sessionTable: "agent_copilot_sessions",
			turnTable: "agent_copilot_session_turns", sessionSchema: agentCopilotSessionV3Schema,
			profile: applicationInteractionProfileAgentCopilot,
		}, nil
	case *postgresApplicationInteractionSessionRepository:
		if typed.pool == nil {
			return nil, errors.New("Agent Copilot sessions require the Workflow PostgreSQL pool")
		}
		return &postgresAgentCopilotSessionRepository{
			pool: typed.pool, sessionTable: "agent_copilot_sessions",
			turnTable: "agent_copilot_session_turns", sessionSchema: agentCopilotSessionV3Schema,
			profile: applicationInteractionProfileAgentCopilot,
		}, nil
	default:
		return nil, errors.New("Agent Copilot sessions require a supported Workflow runtime backend")
	}
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
	if session.SchemaVersion == agentCopilotSessionV3Schema {
		return repository.agent.Create(ctx, session)
	}
	return repository.legacy.Create(ctx, session)
}

func (repository *combinedApplicationInteractionSessionRepository) Read(ctx ApplicationInteractionContext, sessionID string) (ApplicationInteractionSession, error) {
	value, err := repository.legacy.Read(ctx, sessionID)
	if !errors.Is(err, errApplicationSessionNotFound) {
		return value, err
	}
	value, err = repository.prompt.Read(ctx, sessionID)
	if !errors.Is(err, errApplicationSessionNotFound) || repository.agent == nil {
		return value, err
	}
	return repository.agent.Read(ctx, sessionID)
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
	if repository.agent != nil {
		agent, agentErr := repository.agent.List(ctx, query)
		if agentErr != nil {
			return nil, agentErr
		}
		values = append(values, agent...)
	}
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
	if updated.SchemaVersion == agentCopilotSessionV3Schema {
		return repository.agent.Close(ctx, sessionID, expectedVersion, updated)
	}
	return repository.legacy.Close(ctx, sessionID, expectedVersion, updated)
}

func (repository *combinedApplicationInteractionSessionRepository) ReserveTurn(ctx ApplicationInteractionContext, expectedVersion int, updated ApplicationInteractionSession, turn ApplicationInteractionTurn) (ApplicationInteractionSession, ApplicationInteractionTurn, bool, error) {
	if turn.SchemaVersion == promptApplicationSessionTurnV2Schema {
		return repository.prompt.ReserveTurn(ctx, expectedVersion, updated, turn)
	}
	if turn.SchemaVersion == agentCopilotSessionTurnV3Schema {
		return repository.agent.ReserveTurn(ctx, expectedVersion, updated, turn)
	}
	return repository.legacy.ReserveTurn(ctx, expectedVersion, updated, turn)
}

func (repository *combinedApplicationInteractionSessionRepository) ReadTurn(ctx ApplicationInteractionContext, sessionID, turnID string) (ApplicationInteractionTurn, error) {
	value, err := repository.legacy.ReadTurn(ctx, sessionID, turnID)
	if !errors.Is(err, errApplicationSessionNotFound) {
		return value, err
	}
	value, err = repository.prompt.ReadTurn(ctx, sessionID, turnID)
	if !errors.Is(err, errApplicationSessionNotFound) || repository.agent == nil {
		return value, err
	}
	return repository.agent.ReadTurn(ctx, sessionID, turnID)
}

func (repository *combinedApplicationInteractionSessionRepository) ReadTurnByClientKey(ctx ApplicationInteractionContext, sessionID, clientKey string) (ApplicationInteractionTurn, error) {
	value, err := repository.legacy.ReadTurnByClientKey(ctx, sessionID, clientKey)
	if !errors.Is(err, errApplicationSessionNotFound) {
		return value, err
	}
	value, err = repository.prompt.ReadTurnByClientKey(ctx, sessionID, clientKey)
	if !errors.Is(err, errApplicationSessionNotFound) || repository.agent == nil {
		return value, err
	}
	return repository.agent.ReadTurnByClientKey(ctx, sessionID, clientKey)
}

func (repository *combinedApplicationInteractionSessionRepository) CompleteTurn(ctx ApplicationInteractionContext, turn ApplicationInteractionTurn) (ApplicationInteractionSession, ApplicationInteractionTurn, bool, error) {
	if turn.SchemaVersion == promptApplicationSessionTurnV2Schema {
		return repository.prompt.CompleteTurn(ctx, turn)
	}
	if turn.SchemaVersion == agentCopilotSessionTurnV3Schema {
		return repository.agent.CompleteTurn(ctx, turn)
	}
	return repository.legacy.CompleteTurn(ctx, turn)
}

func (repository *combinedApplicationInteractionSessionRepository) ListTurns(ctx ApplicationInteractionContext, sessionID string) ([]ApplicationInteractionTurn, error) {
	values, err := repository.legacy.ListTurns(ctx, sessionID)
	if !errors.Is(err, errApplicationSessionNotFound) {
		return values, err
	}
	values, err = repository.prompt.ListTurns(ctx, sessionID)
	if !errors.Is(err, errApplicationSessionNotFound) || repository.agent == nil {
		return values, err
	}
	return repository.agent.ListTurns(ctx, sessionID)
}

var _ applicationInteractionSessionRepository = (*combinedApplicationInteractionSessionRepository)(nil)

// Concrete Agent adapters use the same repository algorithm with dedicated
// physical tables; their implementations are provided by the Batch D
// projection files.
type sqliteAgentCopilotSessionRepository = sqlitePromptApplicationSessionRepository
type postgresAgentCopilotSessionRepository = postgresPromptApplicationSessionRepository

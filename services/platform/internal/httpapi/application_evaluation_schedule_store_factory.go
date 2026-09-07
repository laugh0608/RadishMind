package httpapi

import "errors"

func newApplicationEvaluationScheduleRepositoryForRunStore(store workflowRunStore) (applicationEvaluationScheduleRepository, error) {
	switch typed := store.(type) {
	case *memoryWorkflowRunStore:
		return newMemoryApplicationEvaluationScheduleRepository(0, 0), nil
	case *sqliteWorkflowRunStore:
		if typed.database == nil {
			return nil, errors.New("application evaluation schedule requires the shared SQLite database")
		}
		return newSQLiteApplicationEvaluationScheduleRepository(typed.database), nil
	case *postgresWorkflowRunStore:
		if typed.pool == nil {
			return nil, errors.New("application evaluation schedule requires the Workflow PostgreSQL pool")
		}
		return newPostgresApplicationEvaluationScheduleRepository(typed.pool), nil
	default:
		return nil, errors.New("application evaluation schedule requires a supported Workflow runtime backend")
	}
}

package httpapi

import "errors"

func newWorkflowTemplateCatalogRepositoryForSavedDraftStore(store savedWorkflowDraftStore) (workflowTemplateCatalogRepository, error) {
	switch typed := store.(type) {
	case *memorySavedWorkflowDraftStore:
		if typed == nil {
			return nil, errors.New("workflow template catalog memory backend is unavailable")
		}
		return newMemoryWorkflowTemplateCatalogRepository(), nil
	case *repositorySavedWorkflowDraftLibraryStore:
		if typed == nil || typed.repositorySavedWorkflowDraftStore == nil {
			return nil, errors.New("workflow template catalog shared repository backend is unavailable")
		}
		return newWorkflowTemplateCatalogRepositoryForSavedDraftRepository(typed.repository)
	case *repositorySavedWorkflowDraftStore:
		if typed == nil {
			return nil, errors.New("workflow template catalog shared repository backend is unavailable")
		}
		return newWorkflowTemplateCatalogRepositoryForSavedDraftRepository(typed.repository)
	default:
		return nil, errors.New("workflow template catalog requires the selected workflow shared backend")
	}
}

func newWorkflowTemplateCatalogRepositoryForSavedDraftRepository(repository SavedWorkflowDraftRepository) (workflowTemplateCatalogRepository, error) {
	adapter, ok := repository.(SavedWorkflowDraftRepositoryAdapter)
	if !ok {
		return nil, errors.New("workflow template catalog requires the workflow shared repository adapter")
	}
	switch executor := adapter.queryExecutor.(type) {
	case *sqliteSavedWorkflowDraftQueryExecutor:
		if executor == nil || executor.database == nil {
			return nil, errors.New("workflow template catalog shared SQLite database is unavailable")
		}
		return newSQLiteWorkflowTemplateCatalogRepository(executor.database), nil
	case *postgresSavedWorkflowDraftQueryExecutor:
		if executor == nil || executor.pool == nil {
			return nil, errors.New("workflow template catalog shared PostgreSQL pool is unavailable")
		}
		return newPostgresWorkflowTemplateCatalogRepository(executor.pool), nil
	default:
		return nil, errors.New("workflow template catalog shared backend type is unsupported")
	}
}

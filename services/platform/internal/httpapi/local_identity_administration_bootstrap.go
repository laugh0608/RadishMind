package httpapi

import (
	"context"
	"fmt"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/sqlitedev"
)

type LocalIdentityDevTestBootstrapOptions struct {
	StoreMode           string
	SQLiteDatabasePath  string
	PostgresDatabaseURL string
	DatabaseTimeout     time.Duration
	TenantRef           string
	WorkspaceID         string
	UserID              string
	AuditRef            string
}

type LocalIdentityDevTestBootstrapResult struct {
	StoreMode            string `json:"store_mode"`
	TenantRef            string `json:"tenant_ref"`
	WorkspaceID          string `json:"workspace_id"`
	UserID               string `json:"user_id"`
	MembershipID         string `json:"membership_id"`
	RoleAssignmentID     string `json:"role_assignment_id"`
	RoleKey              string `json:"role_key"`
	RoleCatalogVersion   string `json:"role_catalog_version"`
	RoleDefinitionDigest string `json:"role_definition_digest"`
	AuditRef             string `json:"audit_ref"`
}

func BootstrapLocalIdentityWorkspaceAdministratorDevTest(
	ctx context.Context,
	options LocalIdentityDevTestBootstrapOptions,
) (LocalIdentityDevTestBootstrapResult, error) {
	mode := strings.TrimSpace(options.StoreMode)
	if mode != localIdentityStoreModeSQLiteDev && mode != localIdentityStoreModePostgresDevTest {
		return LocalIdentityDevTestBootstrapResult{}, fmt.Errorf(
			"local identity administrator bootstrap requires sqlite_dev or postgres_dev_test: %w",
			errLocalIdentityContractMismatch,
		)
	}

	var sqliteRuntime *sqlitedev.Runtime
	if mode == localIdentityStoreModeSQLiteDev {
		var err error
		sqliteRuntime, err = sqlitedev.Open(identityContext(ctx), sqlitedev.Options{
			DatabasePath: strings.TrimSpace(options.SQLiteDatabasePath),
			BusyTimeout:  options.DatabaseTimeout,
			Migrations:   localPersistenceSQLiteMigrations(),
		})
		if err != nil {
			return LocalIdentityDevTestBootstrapResult{}, fmt.Errorf(
				"open local identity SQLite development store: %v: %w",
				err,
				errLocalIdentityStoreUnavailable,
			)
		}
		defer func() { _ = sqliteRuntime.Close() }()
	}

	repository, closeRepository, err := newLocalIdentityRepositoryFromOptions(localIdentityStoreOptions{
		Mode:                mode,
		SQLiteRuntime:       sqliteRuntime,
		PostgresDatabaseURL: strings.TrimSpace(options.PostgresDatabaseURL),
		DatabaseTimeout:     options.DatabaseTimeout,
	})
	if err != nil {
		return LocalIdentityDevTestBootstrapResult{}, fmt.Errorf(
			"open local identity development repository: %v: %w",
			err,
			errLocalIdentityStoreUnavailable,
		)
	}
	defer closeRepository()
	administrationRepository, ok := repository.(localIdentityAdministrationRepository)
	if !ok {
		return LocalIdentityDevTestBootstrapResult{}, errLocalIdentityAdminUnavailable
	}
	created, err := newLocalIdentityAdministrationService(administrationRepository).BootstrapWorkspaceAdministrator(
		identityContext(ctx),
		LocalIdentityBootstrapWorkspaceAdministratorInput{
			TenantRef:   strings.TrimSpace(options.TenantRef),
			WorkspaceID: strings.TrimSpace(options.WorkspaceID),
			UserID:      strings.TrimSpace(options.UserID),
			AuditRef:    strings.TrimSpace(options.AuditRef),
		},
	)
	if err != nil {
		return LocalIdentityDevTestBootstrapResult{}, err
	}
	return LocalIdentityDevTestBootstrapResult{
		StoreMode:            mode,
		TenantRef:            created.Membership.TenantRef,
		WorkspaceID:          created.Membership.WorkspaceID,
		UserID:               created.Membership.UserID,
		MembershipID:         created.Membership.MembershipID,
		RoleAssignmentID:     created.RoleAssignment.AssignmentID,
		RoleKey:              created.RoleAssignment.RoleKey,
		RoleCatalogVersion:   created.RoleAssignment.RoleCatalogVersion,
		RoleDefinitionDigest: created.RoleAssignment.RoleDefinitionDigest,
		AuditRef:             created.Membership.AuditRef,
	}, nil
}

func LocalIdentityFailureCode(err error) string {
	return localIdentityRepositoryError(err)
}

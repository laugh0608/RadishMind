package localidentityrecords

import (
	_ "embed"

	"radishmind.local/services/platform/internal/sqlitedev"
)

const (
	Component                        = "local_identity_records"
	MigrationID                      = "0001_local_identity_records"
	StoreSchemaVersion               = "local_identity_records_store_v1"
	OIDCAuthorizationMigrationID     = "0002_local_identity_oidc_authorization_transactions"
	OIDCAuthorizationSchemaVersion   = "local_identity_records_store_v2"
	AdministrationMigrationID        = "0003_local_identity_administration"
	AdministrationSchemaVersion      = "local_identity_records_store_v3"
	SelfServiceMigrationID           = "0004_local_identity_self_service_sessions"
	SelfServiceSchemaVersion         = "local_identity_records_store_v4"
	WorkspaceInvitationMigrationID   = "0005_workspace_invitations"
	WorkspaceInvitationSchemaVersion = "local_identity_records_store_v5"
)

//go:embed 0001_local_identity_records.up.sql
var upSQL string

//go:embed 0002_local_identity_oidc_authorization_transactions.up.sql
var oidcAuthorizationUpSQL string

//go:embed 0003_local_identity_administration.up.sql
var administrationUpSQL string

//go:embed 0004_local_identity_self_service_sessions.up.sql
var selfServiceUpSQL string

//go:embed 0005_workspace_invitations.up.sql
var workspaceInvitationUpSQL string

func Migrations() []sqlitedev.Migration {
	return []sqlitedev.Migration{{
		Component:          Component,
		ID:                 MigrationID,
		StoreSchemaVersion: StoreSchemaVersion,
		UpSQL:              upSQL,
	}, {
		Component:          Component,
		ID:                 OIDCAuthorizationMigrationID,
		StoreSchemaVersion: OIDCAuthorizationSchemaVersion,
		UpSQL:              oidcAuthorizationUpSQL,
	}, {
		Component:          Component,
		ID:                 AdministrationMigrationID,
		StoreSchemaVersion: AdministrationSchemaVersion,
		UpSQL:              administrationUpSQL,
	}, {
		Component:          Component,
		ID:                 SelfServiceMigrationID,
		StoreSchemaVersion: SelfServiceSchemaVersion,
		UpSQL:              selfServiceUpSQL,
	}, {
		Component:          Component,
		ID:                 WorkspaceInvitationMigrationID,
		StoreSchemaVersion: WorkspaceInvitationSchemaVersion,
		UpSQL:              workspaceInvitationUpSQL,
	}}
}

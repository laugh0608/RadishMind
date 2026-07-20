package httpapi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type sqliteGatewayRequestStore struct {
	database *sql.DB
}

func newSQLiteGatewayRequestStore(database *sql.DB) *sqliteGatewayRequestStore {
	return &sqliteGatewayRequestStore{database: database}
}

func (store *sqliteGatewayRequestStore) CreateRequest(
	requestContext GatewayRequestContext,
	record *GatewayRequestRecord,
) error {
	if store == nil || store.database == nil || record == nil {
		return errGatewayRequestStoreUnavailable
	}
	if record.StoreMode == "" {
		record.StoreMode = gatewayRequestStoreModeSQLiteDev
	}
	if record.StoreMode != gatewayRequestStoreModeSQLiteDev || record.RecordVersion != 0 ||
		record.Status != GatewayRequestStatusStarted {
		return errGatewayRequestStoreContract
	}
	if err := validateGatewayRequestStoreRecord(requestContext, record); err != nil {
		return err
	}
	next := cloneGatewayRequestRecord(*record)
	next.RecordVersion = 1
	payload, startedAt, completedAt, err := sqliteGatewayRequestStorageValues(next)
	if err != nil {
		return errGatewayRequestStoreContract
	}
	stored, err := scanSQLiteGatewayRequestRecord(requestContext, store.database.QueryRowContext(
		requestDatabaseContext(requestContext),
		`INSERT INTO gateway_request_records
        (tenant_ref, workspace_id, consumer_ref, application_id, request_id, record_version, schema_version,
         store_mode, request_route, protocol, request_status, started_at_unix_nano, completed_at_unix_nano,
         selected_provider, selected_profile, selected_model, failure_boundary, usage_availability,
         sanitized_request_record)
        VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
        ON CONFLICT DO NOTHING RETURNING sanitized_request_record`,
		requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ConsumerRef,
		requestContext.ApplicationID, next.RequestID, next.RecordVersion, next.SchemaVersion, next.StoreMode,
		next.Route, next.Protocol, next.Status, startedAt, completedAt, next.SelectedProvider,
		next.SelectedProfile, next.SelectedModel, next.FailureBoundary, next.Usage.Availability, payload,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return errGatewayRequestStoreConflict
	}
	if err != nil {
		return normalizeSQLiteGatewayRequestStoreError(err)
	}
	*record = stored
	return nil
}

func (store *sqliteGatewayRequestStore) UpdateRequest(
	requestContext GatewayRequestContext,
	record *GatewayRequestRecord,
) error {
	if store == nil || store.database == nil || record == nil {
		return errGatewayRequestStoreUnavailable
	}
	if record.StoreMode != gatewayRequestStoreModeSQLiteDev || record.RecordVersion < 1 {
		return errGatewayRequestStoreContract
	}
	if err := validateGatewayRequestStoreRecord(requestContext, record); err != nil {
		return err
	}
	next := cloneGatewayRequestRecord(*record)
	next.RecordVersion++
	payload, startedAt, completedAt, err := sqliteGatewayRequestStorageValues(next)
	if err != nil {
		return errGatewayRequestStoreContract
	}
	stored, err := scanSQLiteGatewayRequestRecord(requestContext, store.database.QueryRowContext(
		requestDatabaseContext(requestContext),
		`UPDATE gateway_request_records SET
            record_version=record_version+1, schema_version=?, store_mode=?, request_route=?, protocol=?,
            request_status=?, completed_at_unix_nano=?, selected_provider=?, selected_profile=?, selected_model=?,
            failure_boundary=?, usage_availability=?, sanitized_request_record=?
        WHERE tenant_ref=? AND workspace_id=? AND consumer_ref=? AND application_id=? AND request_id=?
          AND record_version=? AND request_status='started' AND started_at_unix_nano=?
        RETURNING sanitized_request_record`,
		next.SchemaVersion, next.StoreMode, next.Route, next.Protocol, next.Status, completedAt,
		next.SelectedProvider, next.SelectedProfile, next.SelectedModel, next.FailureBoundary,
		next.Usage.Availability, payload, requestContext.TenantRef, requestContext.WorkspaceID,
		requestContext.ConsumerRef, requestContext.ApplicationID, next.RequestID, record.RecordVersion, startedAt,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return errGatewayRequestStoreConflict
	}
	if err != nil {
		return normalizeSQLiteGatewayRequestStoreError(err)
	}
	*record = stored
	return nil
}

func (store *sqliteGatewayRequestStore) ReadRequest(
	requestContext GatewayRequestContext,
	requestID string,
) (GatewayRequestRecord, bool, error) {
	if store == nil || store.database == nil {
		return GatewayRequestRecord{}, false, errGatewayRequestStoreUnavailable
	}
	record, err := scanSQLiteGatewayRequestRecord(requestContext, store.database.QueryRowContext(
		requestDatabaseContext(requestContext),
		`SELECT sanitized_request_record FROM gateway_request_records
        WHERE tenant_ref=? AND workspace_id=? AND consumer_ref=? AND application_id=? AND request_id=?`,
		requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ConsumerRef,
		requestContext.ApplicationID, requestID,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return GatewayRequestRecord{}, false, nil
	}
	if err != nil {
		return GatewayRequestRecord{}, false, normalizeSQLiteGatewayRequestStoreError(err)
	}
	return record, true, nil
}

func (store *sqliteGatewayRequestStore) ListRequests(
	requestContext GatewayRequestContext,
	filter GatewayRequestListFilter,
) (GatewayRequestListPage, error) {
	if store == nil || store.database == nil {
		return GatewayRequestListPage{}, errGatewayRequestStoreUnavailable
	}
	startedFrom, err := optionalGatewayRequestUnixNano(filter.StartedFrom)
	if err != nil {
		return GatewayRequestListPage{}, errGatewayRequestStoreContract
	}
	startedTo, err := optionalGatewayRequestUnixNano(filter.StartedTo)
	if err != nil {
		return GatewayRequestListPage{}, errGatewayRequestStoreContract
	}
	beforeTime, err := optionalGatewayRequestUnixNano(filter.BeforeTime)
	if err != nil {
		return GatewayRequestListPage{}, errGatewayRequestStoreContract
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = gatewayRequestListDefaultLimit
	}
	rows, err := store.database.QueryContext(requestDatabaseContext(requestContext),
		`SELECT sanitized_request_record FROM gateway_request_records
        WHERE tenant_ref=? AND workspace_id=? AND consumer_ref=? AND application_id=?
          AND (?='' OR request_route=?) AND (?='' OR protocol=?)
          AND (?='' OR selected_provider=?) AND (?='' OR selected_profile=?) AND (?='' OR selected_model=?)
          AND (?='' OR request_status=?) AND (?='' OR failure_boundary=?)
          AND (?='' OR usage_availability=?)
          AND (? IS NULL OR started_at_unix_nano >= ?) AND (? IS NULL OR started_at_unix_nano <= ?)
          AND (? IS NULL OR started_at_unix_nano < ? OR (started_at_unix_nano = ? AND request_id < ?))
        ORDER BY started_at_unix_nano DESC, request_id DESC LIMIT ?`,
		requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ConsumerRef,
		requestContext.ApplicationID, filter.Route, filter.Route, filter.Protocol, filter.Protocol,
		filter.Provider, filter.Provider, filter.Profile, filter.Profile, filter.Model, filter.Model,
		string(filter.Status), string(filter.Status), filter.FailureBoundary, filter.FailureBoundary,
		string(filter.UsageAvailability), string(filter.UsageAvailability),
		startedFrom, startedFrom, startedTo, startedTo, beforeTime, beforeTime, beforeTime,
		filter.BeforeRequestID, limit+1,
	)
	if err != nil {
		return GatewayRequestListPage{}, errGatewayRequestStoreUnavailable
	}
	defer rows.Close()
	records := make([]GatewayRequestRecord, 0, limit+1)
	for rows.Next() {
		record, scanErr := scanSQLiteGatewayRequestRecord(requestContext, rows)
		if scanErr != nil {
			return GatewayRequestListPage{}, normalizeSQLiteGatewayRequestStoreError(scanErr)
		}
		records = append(records, record)
	}
	if rows.Err() != nil {
		return GatewayRequestListPage{}, errGatewayRequestStoreUnavailable
	}
	hasMore := len(records) > limit
	if hasMore {
		records = records[:limit]
	}
	return GatewayRequestListPage{Records: records, HasMore: hasMore}, nil
}

type sqliteGatewayRequestRow interface {
	Scan(dest ...any) error
}

func scanSQLiteGatewayRequestRecord(
	requestContext GatewayRequestContext,
	row sqliteGatewayRequestRow,
) (GatewayRequestRecord, error) {
	var payload []byte
	if err := row.Scan(&payload); err != nil {
		return GatewayRequestRecord{}, err
	}
	return decodeGatewayRequestStoreRecord(requestContext, payload, gatewayRequestStoreModeSQLiteDev)
}

func sqliteGatewayRequestStorageValues(record GatewayRequestRecord) (string, int64, any, error) {
	payload, err := json.Marshal(record)
	if err != nil {
		return "", 0, nil, err
	}
	startedAt, err := gatewayRequestUnixNano(record.StartedAt)
	if err != nil {
		return "", 0, nil, err
	}
	var completedAt any
	if record.CompletedAt != "" {
		value, parseErr := gatewayRequestUnixNano(record.CompletedAt)
		if parseErr != nil || value < startedAt {
			return "", 0, nil, errGatewayRequestStoreContract
		}
		completedAt = value
	}
	return string(payload), startedAt, completedAt, nil
}

func gatewayRequestUnixNano(value string) (int64, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return 0, err
	}
	parsed = parsed.UTC()
	unixNano := parsed.UnixNano()
	if !time.Unix(0, unixNano).UTC().Equal(parsed) {
		return 0, errors.New("Gateway request time is outside SQLite nanosecond range")
	}
	return unixNano, nil
}

func optionalGatewayRequestUnixNano(value *time.Time) (any, error) {
	if value == nil {
		return nil, nil
	}
	formatted := value.UTC().Format(time.RFC3339Nano)
	return gatewayRequestUnixNano(formatted)
}

func normalizeSQLiteGatewayRequestStoreError(err error) error {
	if errors.Is(err, errGatewayRequestStoreContract) {
		return errGatewayRequestStoreContract
	}
	return errGatewayRequestStoreUnavailable
}

var _ gatewayRequestStore = (*sqliteGatewayRequestStore)(nil)

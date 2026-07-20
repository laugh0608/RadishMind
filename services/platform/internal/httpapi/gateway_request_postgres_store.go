package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresGatewayRequestStore struct {
	pool *pgxpool.Pool
}

func newPostgresGatewayRequestStore(pool *pgxpool.Pool) *postgresGatewayRequestStore {
	return &postgresGatewayRequestStore{pool: pool}
}

func (store *postgresGatewayRequestStore) CreateRequest(requestContext GatewayRequestContext, record *GatewayRequestRecord) error {
	if store == nil || store.pool == nil || record == nil {
		return errGatewayRequestStoreContract
	}
	if record.StoreMode == "" {
		record.StoreMode = gatewayRequestStoreModePostgresDevTest
	}
	if record.StoreMode != gatewayRequestStoreModePostgresDevTest || record.RecordVersion != 0 || record.Status != GatewayRequestStatusStarted {
		return errGatewayRequestStoreContract
	}
	if err := validateGatewayRequestStoreRecord(requestContext, record); err != nil {
		return err
	}
	next := cloneGatewayRequestRecord(*record)
	next.RecordVersion = 1
	payload, startedAt, completedAt, err := encodePostgresGatewayRequestRecord(next)
	if err != nil {
		return err
	}
	var storedVersion int
	err = store.pool.QueryRow(requestDatabaseContext(requestContext), `INSERT INTO gateway_request_records
 (tenant_ref,workspace_id,consumer_ref,application_id,request_id,record_version,schema_version,store_mode,request_route,protocol,request_status,started_at,completed_at,selected_provider,selected_profile,selected_model,failure_boundary,usage_availability,sanitized_request_record)
 VALUES ($1,$2,$3,$4,$5,1,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
 ON CONFLICT DO NOTHING RETURNING record_version`,
		requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ConsumerRef, requestContext.ApplicationID,
		next.RequestID, next.SchemaVersion, next.StoreMode, next.Route, next.Protocol, next.Status, startedAt, completedAt,
		next.SelectedProvider, next.SelectedProfile, next.SelectedModel, next.FailureBoundary, next.Usage.Availability, payload,
	).Scan(&storedVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return errGatewayRequestStoreConflict
	}
	if err != nil {
		return errGatewayRequestStoreUnavailable
	}
	record.RecordVersion = storedVersion
	return nil
}

func (store *postgresGatewayRequestStore) UpdateRequest(requestContext GatewayRequestContext, record *GatewayRequestRecord) error {
	if store == nil || store.pool == nil || record == nil || record.StoreMode != gatewayRequestStoreModePostgresDevTest || record.RecordVersion < 1 {
		return errGatewayRequestStoreContract
	}
	if err := validateGatewayRequestStoreRecord(requestContext, record); err != nil {
		return err
	}
	next := cloneGatewayRequestRecord(*record)
	next.RecordVersion++
	payload, _, completedAt, err := encodePostgresGatewayRequestRecord(next)
	if err != nil {
		return err
	}
	var storedVersion int
	err = store.pool.QueryRow(requestDatabaseContext(requestContext), `UPDATE gateway_request_records SET
 record_version=record_version+1,schema_version=$1,store_mode=$2,request_route=$3,protocol=$4,request_status=$5,
 completed_at=$6,selected_provider=$7,selected_profile=$8,selected_model=$9,failure_boundary=$10,
 usage_availability=$11,sanitized_request_record=$12
 WHERE tenant_ref=$13 AND workspace_id=$14 AND consumer_ref=$15 AND application_id=$16 AND request_id=$17
 AND record_version=$18 AND request_status='started' RETURNING record_version`,
		next.SchemaVersion, next.StoreMode, next.Route, next.Protocol, next.Status, completedAt,
		next.SelectedProvider, next.SelectedProfile, next.SelectedModel, next.FailureBoundary,
		next.Usage.Availability, payload, requestContext.TenantRef, requestContext.WorkspaceID,
		requestContext.ConsumerRef, requestContext.ApplicationID, next.RequestID, record.RecordVersion,
	).Scan(&storedVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return errGatewayRequestStoreConflict
	}
	if err != nil {
		return errGatewayRequestStoreUnavailable
	}
	record.RecordVersion = storedVersion
	return nil
}

func (store *postgresGatewayRequestStore) ReadRequest(requestContext GatewayRequestContext, requestID string) (GatewayRequestRecord, bool, error) {
	if store == nil || store.pool == nil {
		return GatewayRequestRecord{}, false, errGatewayRequestStoreContract
	}
	var payload []byte
	err := store.pool.QueryRow(requestDatabaseContext(requestContext), `SELECT sanitized_request_record FROM gateway_request_records
 WHERE tenant_ref=$1 AND workspace_id=$2 AND consumer_ref=$3 AND application_id=$4 AND request_id=$5`,
		requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ConsumerRef, requestContext.ApplicationID, requestID,
	).Scan(&payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return GatewayRequestRecord{}, false, nil
	}
	if err != nil {
		return GatewayRequestRecord{}, false, errGatewayRequestStoreUnavailable
	}
	record, err := decodePostgresGatewayRequestRecord(requestContext, payload)
	if err != nil {
		return GatewayRequestRecord{}, false, err
	}
	return record, true, nil
}

func (store *postgresGatewayRequestStore) ListRequests(requestContext GatewayRequestContext, filter GatewayRequestListFilter) (GatewayRequestListPage, error) {
	if store == nil || store.pool == nil {
		return GatewayRequestListPage{}, errGatewayRequestStoreContract
	}
	rows, err := store.pool.Query(requestDatabaseContext(requestContext), `SELECT sanitized_request_record FROM gateway_request_records
 WHERE tenant_ref=$1 AND workspace_id=$2 AND consumer_ref=$3 AND application_id=$4
 AND ($5='' OR request_route=$5) AND ($6='' OR protocol=$6)
 AND ($7='' OR selected_provider=$7) AND ($8='' OR selected_profile=$8) AND ($9='' OR selected_model=$9)
 AND ($10='' OR request_status=$10) AND ($11='' OR failure_boundary=$11) AND ($12='' OR usage_availability=$12)
 AND ($13::timestamptz IS NULL OR started_at >= $13) AND ($14::timestamptz IS NULL OR started_at <= $14)
 AND ($15::timestamptz IS NULL OR (started_at,request_id) < ($15,$16))
 ORDER BY started_at DESC, request_id DESC LIMIT $17`,
		requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ConsumerRef, requestContext.ApplicationID,
		filter.Route, filter.Protocol, filter.Provider, filter.Profile, filter.Model, string(filter.Status),
		filter.FailureBoundary, string(filter.UsageAvailability), filter.StartedFrom, filter.StartedTo,
		filter.BeforeTime, filter.BeforeRequestID, filter.Limit+1,
	)
	if err != nil {
		return GatewayRequestListPage{}, errGatewayRequestStoreUnavailable
	}
	defer rows.Close()
	records := make([]GatewayRequestRecord, 0, filter.Limit+1)
	for rows.Next() {
		var payload []byte
		if err = rows.Scan(&payload); err != nil {
			return GatewayRequestListPage{}, errGatewayRequestStoreUnavailable
		}
		record, decodeErr := decodePostgresGatewayRequestRecord(requestContext, payload)
		if decodeErr != nil {
			return GatewayRequestListPage{}, decodeErr
		}
		records = append(records, record)
	}
	if rows.Err() != nil {
		return GatewayRequestListPage{}, errGatewayRequestStoreUnavailable
	}
	hasMore := len(records) > filter.Limit
	if hasMore {
		records = records[:filter.Limit]
	}
	return GatewayRequestListPage{Records: records, HasMore: hasMore}, nil
}

func encodePostgresGatewayRequestRecord(record GatewayRequestRecord) ([]byte, time.Time, *time.Time, error) {
	payload, err := json.Marshal(record)
	if err != nil {
		return nil, time.Time{}, nil, errGatewayRequestStoreContract
	}
	startedAt, err := time.Parse(time.RFC3339Nano, record.StartedAt)
	if err != nil {
		return nil, time.Time{}, nil, errGatewayRequestStoreContract
	}
	var completedAt *time.Time
	if record.CompletedAt != "" {
		parsed, parseErr := time.Parse(time.RFC3339Nano, record.CompletedAt)
		if parseErr != nil {
			return nil, time.Time{}, nil, errGatewayRequestStoreContract
		}
		completedAt = &parsed
	}
	return payload, startedAt, completedAt, nil
}

func decodePostgresGatewayRequestRecord(requestContext GatewayRequestContext, payload []byte) (GatewayRequestRecord, error) {
	return decodeGatewayRequestStoreRecord(requestContext, payload, gatewayRequestStoreModePostgresDevTest)
}

func requestDatabaseContext(requestContext GatewayRequestContext) context.Context {
	if requestContext.RequestContext != nil {
		return requestContext.RequestContext
	}
	return context.Background()
}

var _ gatewayRequestStore = (*postgresGatewayRequestStore)(nil)

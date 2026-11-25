package developmentartifact

import (
	"cmp"
	"context"
	"slices"
	"time"

	control "radishmind.local/services/platform/internal/developmentcontrol"
)

type Summary struct {
	ArtifactID        string `json:"artifact_id"`
	TaskID            string `json:"task_id"`
	AttemptID         string `json:"attempt_id"`
	Digest            string `json:"digest"`
	VerificationState string `json:"verification_state"`
	SavedAtMS         int64  `json:"saved_at_ms"`
}
type Page struct {
	Artifacts  []Summary `json:"artifacts"`
	NextCursor string    `json:"next_cursor"`
	HasMore    bool      `json:"has_more"`
}

func (s *Service) List(ctx context.Context, scope Scope, cursor string, limit int) (Page, error) {
	if err := s.authorize(ctx, scope, PermissionRead); err != nil {
		return Page{}, err
	}
	if limit < 1 || limit > 50 {
		return Page{}, ErrInvalid
	}
	binding := control.TaskScope{Identity: control.Identity{TenantRef: scope.TenantRef, WorkspaceID: scope.WorkspaceID, AccountID: scope.AccountID}, ApplicationID: scope.ApplicationID}
	var after control.HistoryPosition
	if cursor != "" {
		var err error
		after, err = control.DecodeHistoryPosition(binding, "artifacts", cursor)
		if err != nil {
			return Page{}, ErrInvalid
		}
	}
	// The immutable store has a hard 64-package owner quota. Scan only keys
	// and dates, then revalidate the packages selected for the response.
	rows, err := s.store.db.QueryContext(ctx, `SELECT artifact_id,json_extract(record,'$.saved_at') FROM development_artifacts
 WHERE tenant_ref=? AND workspace_id=? AND account_id=? AND application_id=? LIMIT ?`, scope.TenantRef, scope.WorkspaceID, scope.AccountID, scope.ApplicationID, MaxSavedCount+1)
	if err != nil {
		return Page{}, ErrUnavailable
	}
	defer rows.Close()
	type key struct {
		id string
		ms int64
	}
	var keys []key
	for rows.Next() {
		var id, date string
		if err := rows.Scan(&id, &date); err != nil {
			return Page{}, ErrCorrupt
		}
		at, err := time.Parse(time.RFC3339Nano, date)
		if err != nil || !idPattern.MatchString(id) || at.UnixMilli() < 1 {
			return Page{}, ErrCorrupt
		}
		keys = append(keys, key{id, at.UnixMilli()})
	}
	if rows.Err() != nil || rows.Close() != nil {
		return Page{}, ErrUnavailable
	}
	if len(keys) > MaxSavedCount {
		return Page{}, ErrCorrupt
	}
	slices.SortFunc(keys, func(a, b key) int {
		if a.ms != b.ms {
			return cmp.Compare(b.ms, a.ms)
		}
		return cmp.Compare(b.id, a.id)
	})
	page := Page{Artifacts: []Summary{}}
	for _, key := range keys {
		if cursor != "" && (key.ms > after.CreatedAtMS || key.ms == after.CreatedAtMS && key.id >= after.ID) {
			continue
		}
		if len(page.Artifacts) == limit {
			last := page.Artifacts[limit-1]
			page.NextCursor, err = control.EncodeHistoryPosition(binding, "artifacts", last.SavedAtMS, last.ArtifactID)
			if err != nil {
				return Page{}, ErrInvalid
			}
			page.HasMore = true
			break
		}
		record, err := s.store.read(ctx, scope, key.id)
		if err != nil {
			return Page{}, err
		}
		b := record.Bundle
		page.Artifacts = append(page.Artifacts, Summary{record.ArtifactID, b.Origin.TaskID, b.Origin.AttemptID, b.Digest, b.VerificationState, key.ms})
	}
	return page, nil
}

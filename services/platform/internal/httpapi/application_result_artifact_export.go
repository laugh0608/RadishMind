package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"
)

type ApplicationResultArtifactExport struct {
	SchemaVersion      string                             `json:"schema_version"`
	Artifact           ApplicationResultArtifact          `json:"artifact"`
	Lifecycle          ApplicationResultArtifactLifecycle `json:"lifecycle"`
	ExportedAt         string                             `json:"exported_at"`
	ExportedByActorRef string                             `json:"exported_by_actor_ref"`
	RequestID          string                             `json:"request_id"`
	AuditRef           string                             `json:"audit_ref"`
	ExportDigest       string                             `json:"export_digest"`
}

type ApplicationResultArtifactExportResult struct {
	Export      *ApplicationResultArtifactExport
	FailureCode string
}

func (service applicationResultArtifactService) Export(
	ctx ApplicationInteractionContext,
	artifactID string,
) ApplicationResultArtifactExportResult {
	if service.repository == nil {
		return applicationResultArtifactExportFailure(ApplicationResultArtifactFailureStoreUnavailable)
	}
	artifactID = strings.TrimSpace(artifactID)
	if validateApplicationInteractionContext(ctx) != nil || !applicationResultArtifactIDPattern.MatchString(artifactID) {
		return applicationResultArtifactExportFailure(ApplicationResultArtifactFailurePayloadInvalid)
	}
	artifact, err := service.repository.Read(ctx, artifactID)
	if err != nil {
		return applicationResultArtifactExportRepositoryFailure(err)
	}
	lifecycle, err := service.repository.ReadLifecycle(ctx, artifactID)
	if err != nil {
		return applicationResultArtifactExportRepositoryFailure(err)
	}
	export := ApplicationResultArtifactExport{
		SchemaVersion: applicationResultArtifactExportSchemaVersion,
		Artifact:      cloneApplicationResultArtifact(artifact), Lifecycle: cloneApplicationResultArtifactLifecycle(lifecycle),
		ExportedAt:         service.now().UTC().Format(time.RFC3339Nano),
		ExportedByActorRef: strings.TrimSpace(ctx.ActorRef), RequestID: strings.TrimSpace(ctx.RequestID),
		AuditRef: strings.TrimSpace(ctx.AuditRef),
	}
	export.ExportDigest = applicationResultArtifactExportDigest(export)
	if validateApplicationResultArtifactExport(ctx, export) != nil {
		return applicationResultArtifactExportFailure(ApplicationResultArtifactFailureStoreContract)
	}
	return ApplicationResultArtifactExportResult{Export: &export}
}

func validateApplicationResultArtifactExport(
	ctx ApplicationInteractionContext,
	export ApplicationResultArtifactExport,
) error {
	if export.SchemaVersion != applicationResultArtifactExportSchemaVersion ||
		validateApplicationResultArtifact(ctx, export.Artifact) != nil ||
		validateApplicationResultArtifactLifecycle(ctx, export.Lifecycle) != nil ||
		export.Artifact.ArtifactID != export.Lifecycle.ArtifactID ||
		parseApplicationInteractionTimestamp(export.ExportedAt) == nil ||
		strings.TrimSpace(export.ExportedByActorRef) == "" || strings.TrimSpace(export.RequestID) == "" ||
		strings.TrimSpace(export.AuditRef) == "" || export.ExportDigest != applicationResultArtifactExportDigest(export) {
		return errApplicationResultArtifactContract
	}
	return nil
}

func applicationResultArtifactExportDigest(export ApplicationResultArtifactExport) string {
	export.ExportDigest = ""
	payload, err := json.Marshal(export)
	if err != nil {
		return ""
	}
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func applicationResultArtifactExportFailure(code string) ApplicationResultArtifactExportResult {
	return ApplicationResultArtifactExportResult{FailureCode: code}
}

func applicationResultArtifactExportRepositoryFailure(err error) ApplicationResultArtifactExportResult {
	return applicationResultArtifactExportFailure(applicationResultArtifactRepositoryFailure(err).FailureCode)
}

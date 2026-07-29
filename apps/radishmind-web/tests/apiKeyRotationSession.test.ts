import assert from "node:assert/strict";
import test from "node:test";

import type { APIKeyRecord } from "../src/features/control-plane-read/apiKeyLifecycleConsumer.ts";
import {
  APIKeyRotationSessionError,
  beginAPIKeyRotationSession,
  cancelAPIKeyRotationSession,
  completeAPIKeyRotationSession,
  currentAPIKeyRotationSourceVersion,
  readAPIKeyRotationSession,
  recordAPIKeyRotationReplacement,
  refreshAPIKeyRotationVerification,
  synchronizeAPIKeyRotationApplication,
} from "../src/features/control-plane-read/apiKeyRotationSession.ts";

test.afterEach(() => {
  cancelAPIKeyRotationSession("app_aaaaaaaaaaaaaaaa");
  cancelAPIKeyRotationSession("app_bbbbbbbbbbbbbbbb");
});

test("guided rotation requires replacement authentication before source retirement", () => {
  const source = apiKeyRecord();
  const started = beginAPIKeyRotationSession(source);
  assert.equal(started.phase, "replacement_pending");
  assert.deepEqual(started.scopes, ["models:read", "responses:invoke"]);
  assert.equal(JSON.stringify(started).includes("rmd_dev_"), false);

  const replacement = apiKeyRecord({
    apiKeyId: "key_bbbbbbbbbbbbbbbb",
    displayName: "Replacement key",
    createdAt: "2026-07-29T02:00:00Z",
    lastUsedAt: null,
  });
  const issued = recordAPIKeyRotationReplacement(source.applicationId, replacement);
  assert.equal(issued.phase, "verification_pending");
  assert.throws(
    () => currentAPIKeyRotationSourceVersion(source.applicationId, source),
    (error: unknown) => rotationFailure(error, "api_key_rotation_replacement_unverified"),
  );

  const verified = refreshAPIKeyRotationVerification(source.applicationId, {
    ...replacement,
    lastUsedAt: "2026-07-29T02:01:00Z",
  });
  assert.equal(verified.phase, "verified");
  assert.equal(currentAPIKeyRotationSourceVersion(source.applicationId, { ...source, recordVersion: 3 }), 3);

  completeAPIKeyRotationSession(source.applicationId, {
    ...source,
    lifecycleState: "revoked",
    effectiveState: "revoked",
    recordVersion: 4,
    revokedAt: "2026-07-29T02:02:00Z",
  });
  assert.equal(readAPIKeyRotationSession(source.applicationId), null);
});

test("guided rotation rejects inactive sources and replacement scope or owner drift", () => {
  const source = apiKeyRecord();
  assert.throws(
    () => beginAPIKeyRotationSession({ ...source, effectiveState: "expired" }),
    (error: unknown) => rotationFailure(error, "api_key_rotation_source_inactive"),
  );

  beginAPIKeyRotationSession(source);
  assert.throws(
    () => recordAPIKeyRotationReplacement(source.applicationId, apiKeyRecord({
      apiKeyId: "key_bbbbbbbbbbbbbbbb",
      scopes: ["models:read"],
    })),
    (error: unknown) => rotationFailure(error, "api_key_rotation_replacement_mismatch"),
  );
  assert.throws(
    () => recordAPIKeyRotationReplacement(source.applicationId, apiKeyRecord({
      apiKeyId: "key_bbbbbbbbbbbbbbbb",
      ownerSubjectRef: "subject_other",
    })),
    (error: unknown) => rotationFailure(error, "api_key_rotation_replacement_mismatch"),
  );
  assert.throws(
    () => recordAPIKeyRotationReplacement(source.applicationId, apiKeyRecord({
      apiKeyId: "key_bbbbbbbbbbbbbbbb",
      lifecycleState: "revoked",
      effectiveState: "revoked",
    })),
    (error: unknown) => rotationFailure(error, "api_key_rotation_replacement_inactive"),
  );
  assert.equal(readAPIKeyRotationSession(source.applicationId)?.phase, "replacement_pending");
});

test("source reread and completion reject drift instead of clearing the session", () => {
  const source = apiKeyRecord();
  const replacement = apiKeyRecord({
    apiKeyId: "key_bbbbbbbbbbbbbbbb",
    lastUsedAt: "2026-07-29T02:01:00Z",
  });
  beginAPIKeyRotationSession(source);
  recordAPIKeyRotationReplacement(source.applicationId, replacement);

  assert.throws(
    () => currentAPIKeyRotationSourceVersion(source.applicationId, { ...source, scopes: ["models:read"] }),
    (error: unknown) => rotationFailure(error, "api_key_rotation_source_mismatch"),
  );
  assert.throws(
    () => currentAPIKeyRotationSourceVersion(source.applicationId, { ...source, effectiveState: "expired" }),
    (error: unknown) => rotationFailure(error, "api_key_rotation_source_inactive"),
  );
  assert.throws(
    () => completeAPIKeyRotationSession(source.applicationId, {
      ...source,
      scopes: ["models:read"],
      lifecycleState: "revoked",
      effectiveState: "revoked",
      revokedAt: "2026-07-29T02:02:00Z",
    }),
    (error: unknown) => rotationFailure(error, "api_key_rotation_completion_invalid"),
  );
  assert.equal(readAPIKeyRotationSession(source.applicationId)?.phase, "verified");
});

test("application switches and explicit cancellation clear only ephemeral rotation metadata", () => {
  const source = apiKeyRecord();
  beginAPIKeyRotationSession(source);
  assert.equal(readAPIKeyRotationSession(source.applicationId)?.sourceApiKeyId, source.apiKeyId);
  assert.throws(
    () => beginAPIKeyRotationSession({ ...source, apiKeyId: "key_bbbbbbbbbbbbbbbb" }),
    (error: unknown) => rotationFailure(error, "api_key_rotation_session_active"),
  );
  assert.equal(synchronizeAPIKeyRotationApplication("app_bbbbbbbbbbbbbbbb"), null);
  assert.equal(readAPIKeyRotationSession(source.applicationId), null);

  beginAPIKeyRotationSession(source);
  cancelAPIKeyRotationSession(source.applicationId);
  assert.equal(readAPIKeyRotationSession(source.applicationId), null);
});

function apiKeyRecord(overrides: Partial<APIKeyRecord> = {}): APIKeyRecord {
  return {
    schemaVersion: "api_key_record.v1",
    apiKeyId: "key_aaaaaaaaaaaaaaaa",
    tenantRef: "tenant_demo",
    workspaceId: "workspace_demo",
    applicationId: "app_aaaaaaaaaaaaaaaa",
    ownerSubjectRef: "subject_demo_user",
    displayName: "Source key",
    scopes: ["responses:invoke", "models:read"],
    lifecycleState: "active",
    effectiveState: "active",
    recordVersion: 1,
    createdAt: "2026-07-29T01:00:00Z",
    expiresAt: "2026-08-28T01:00:00Z",
    lastUsedAt: "2026-07-29T01:01:00Z",
    revokedAt: null,
    createdByActorRef: "subject_demo_user",
    revokedByActorRef: null,
    requestId: "request-api-key-rotation-0001",
    auditRef: "audit-api-key-rotation-0001",
    ...overrides,
  };
}

function rotationFailure(error: unknown, failureCode: string): boolean {
  return error instanceof APIKeyRotationSessionError && error.failureCode === failureCode;
}

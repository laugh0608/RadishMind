CREATE TABLE development_artifacts (
 tenant_ref TEXT NOT NULL,
 workspace_id TEXT NOT NULL,
 account_id TEXT NOT NULL,
 application_id TEXT NOT NULL,
 artifact_id TEXT NOT NULL,
 request_id TEXT NOT NULL,
 bundle_digest TEXT NOT NULL,
 package_bytes INTEGER NOT NULL CHECK(package_bytes > 0 AND package_bytes <= 1048576),
 record TEXT NOT NULL CHECK(json_valid(record)),
 record_digest TEXT NOT NULL,
 PRIMARY KEY(tenant_ref,workspace_id,account_id,application_id,artifact_id),
 UNIQUE(tenant_ref,workspace_id,account_id,application_id,request_id),
 CHECK(json_extract(record,'$.artifact_id')=artifact_id),
 CHECK(json_extract(record,'$.request_id')=request_id),
 CHECK(json_extract(record,'$.bundle.digest')=bundle_digest),
 CHECK(json_extract(record,'$.bundle.scope.tenant_ref')=tenant_ref),
 CHECK(json_extract(record,'$.bundle.scope.workspace_id')=workspace_id),
 CHECK(json_extract(record,'$.bundle.scope.account_id')=account_id),
 CHECK(json_extract(record,'$.bundle.scope.application_id')=application_id)
);
CREATE TRIGGER development_artifacts_quota BEFORE INSERT ON development_artifacts BEGIN
 SELECT CASE WHEN (SELECT COUNT(*) FROM development_artifacts WHERE tenant_ref=NEW.tenant_ref AND workspace_id=NEW.workspace_id AND account_id=NEW.account_id) >= 64
 OR (SELECT COALESCE(SUM(package_bytes),0) FROM development_artifacts WHERE tenant_ref=NEW.tenant_ref AND workspace_id=NEW.workspace_id AND account_id=NEW.account_id) + NEW.package_bytes > 67108864
 THEN RAISE(ABORT,'development_artifact_quota_exceeded') END;
END;
CREATE TRIGGER development_artifacts_no_update BEFORE UPDATE ON development_artifacts BEGIN
 SELECT RAISE(ABORT,'development_artifact_immutable');
END;
CREATE TRIGGER development_artifacts_no_delete BEFORE DELETE ON development_artifacts BEGIN
 SELECT RAISE(ABORT,'development_artifact_retained');
END;

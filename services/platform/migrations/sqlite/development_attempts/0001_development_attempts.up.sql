CREATE TABLE development_attempts (
 tenant_ref TEXT NOT NULL, workspace_id TEXT NOT NULL, attempt_id TEXT NOT NULL,
 account_id TEXT NOT NULL, confirmation_id TEXT NOT NULL,
 owner_instance_id TEXT NOT NULL, record_version INTEGER NOT NULL CHECK(record_version>0),
 credential_digest TEXT NOT NULL UNIQUE, credential_revoked INTEGER NOT NULL CHECK(credential_revoked IN(0,1)),
 state_record TEXT NOT NULL CHECK(json_valid(state_record)), state_digest TEXT NOT NULL,
 binding_record TEXT NOT NULL CHECK(json_valid(binding_record)), binding_digest TEXT NOT NULL,
 context_digest TEXT NOT NULL, context_count INTEGER NOT NULL CHECK(context_count>=0),
 pending_tool_request_id TEXT NOT NULL,
 PRIMARY KEY(tenant_ref,workspace_id,attempt_id),
 UNIQUE(tenant_ref,workspace_id,account_id,confirmation_id)
) STRICT;
CREATE TABLE development_events (
 tenant_ref TEXT NOT NULL, workspace_id TEXT NOT NULL, attempt_id TEXT NOT NULL,
 sequence INTEGER NOT NULL CHECK(sequence>0), event_record TEXT NOT NULL CHECK(json_valid(event_record)),
 event_digest TEXT NOT NULL,
 PRIMARY KEY(tenant_ref,workspace_id,attempt_id,sequence),
 FOREIGN KEY(tenant_ref,workspace_id,attempt_id) REFERENCES development_attempts(tenant_ref,workspace_id,attempt_id)
) STRICT;
CREATE TABLE development_requests (
 tenant_ref TEXT NOT NULL, workspace_id TEXT NOT NULL, attempt_id TEXT NOT NULL,
 request_id TEXT NOT NULL, kind TEXT NOT NULL CHECK(kind IN('model','tool')),
 payload_digest TEXT NOT NULL,
 PRIMARY KEY(tenant_ref,workspace_id,attempt_id,request_id),
 FOREIGN KEY(tenant_ref,workspace_id,attempt_id) REFERENCES development_attempts(tenant_ref,workspace_id,attempt_id)
) STRICT;
CREATE TABLE development_tool_offers (
 tenant_ref TEXT NOT NULL, workspace_id TEXT NOT NULL, attempt_id TEXT NOT NULL,
 request_id TEXT NOT NULL, parent_request_id TEXT NOT NULL, tool_call_id TEXT NOT NULL,
 payload_digest TEXT NOT NULL, arguments_digest TEXT NOT NULL,
 state TEXT NOT NULL CHECK(state IN('pending','queued','completed')),
 observation_digest TEXT NOT NULL,
 PRIMARY KEY(tenant_ref,workspace_id,attempt_id,request_id),
 UNIQUE(tenant_ref,workspace_id,attempt_id,parent_request_id),
 UNIQUE(tenant_ref,workspace_id,attempt_id,tool_call_id),
 FOREIGN KEY(tenant_ref,workspace_id,attempt_id,parent_request_id) REFERENCES development_requests(tenant_ref,workspace_id,attempt_id,request_id)
) STRICT;
CREATE TRIGGER development_events_no_update BEFORE UPDATE ON development_events BEGIN SELECT RAISE(ABORT,'development events are append-only'); END;
CREATE TRIGGER development_events_no_delete BEFORE DELETE ON development_events BEGIN SELECT RAISE(ABORT,'development events are append-only'); END;
CREATE TRIGGER development_requests_no_update BEFORE UPDATE ON development_requests BEGIN SELECT RAISE(ABORT,'development request identities are immutable'); END;
CREATE TRIGGER development_requests_no_delete BEFORE DELETE ON development_requests BEGIN SELECT RAISE(ABORT,'development request identities are immutable'); END;
CREATE TRIGGER development_tool_offer_identity BEFORE UPDATE ON development_tool_offers
WHEN NEW.tenant_ref!=OLD.tenant_ref OR NEW.workspace_id!=OLD.workspace_id OR NEW.attempt_id!=OLD.attempt_id
 OR NEW.request_id!=OLD.request_id OR NEW.parent_request_id!=OLD.parent_request_id OR NEW.tool_call_id!=OLD.tool_call_id
 OR NEW.payload_digest!=OLD.payload_digest OR NEW.arguments_digest!=OLD.arguments_digest
 OR (OLD.state='completed' AND (NEW.state!='completed' OR NEW.observation_digest!=OLD.observation_digest))
 OR (OLD.state='queued' AND NEW.state='pending')
BEGIN SELECT RAISE(ABORT,'development offer identity and consumption are immutable'); END;
CREATE TRIGGER development_tool_offer_no_delete BEFORE DELETE ON development_tool_offers BEGIN SELECT RAISE(ABORT,'development offers cannot be deleted'); END;

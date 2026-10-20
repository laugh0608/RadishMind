import type { Page } from "@playwright/test";

type Plan = { plan_id: string; latest_plan_version: number; latest_plan_digest: string; item_count: number };
// Browser-only strict contract projections. No worker, delegation or scheduling API
// is enabled by this fixture; real Plan/Campaign persistence is tested separately.
export async function installScheduleProjection(page: Page, applicationId: string, plan: Plan) {
  const time = "2026-10-11T02:30:00Z", id = "aesch_aaaaaaaaaaaaaaaa", digest = `sha256:${"a".repeat(64)}`;
  const scope = { tenant_ref: "tenant_demo", workspace_id: "workspace_demo", environment: "test", application_id: applicationId };
  const refs = { request_id: "request_schedule_fixture", audit_ref: "audit_schedule_fixture" };
  const owner = { ...scope, plan_id: plan.plan_id, plan_version: plan.latest_plan_version, plan_digest: plan.latest_plan_digest, execution_profile: "prompt_application_invocation_v1", quota_api_key_id: "key_aaaaaaaaaaaaaaaa" };
  const actors = { system_actor_ref: "system:application-evaluation-scheduler", delegated_by_user_ref: "subject_demo_user" };
  const record = { schema_version: "application_evaluation_schedule.v1", ...owner, ...actors, ...refs, schedule_id: id,
    record_version: 1, latest_schedule_version: 1, latest_schedule_digest: digest, authorization_model: "system_actor_schedule_scoped_delegation_v1",
    lifecycle_state: "draft", next_due_at: null as string | null, created_at: time, updated_at: time, created_by_actor_ref: "subject_demo_user", updated_by_actor_ref: "subject_demo_user" };
  const version = { schema_version: "application_evaluation_schedule_version.v1", ...owner, ...refs, schedule_id: id, schedule_version: 1,
    previous_schedule_version: 0, schedule_digest: digest, schedule: { rule: "daily_utc", hour: 2, minute: 30 }, item_count: plan.item_count, max_provider_attempts: plan.item_count,
    missed_window_policy: "record_only_no_catch_up", overlap_policy: "skip_while_campaign_non_terminal",
    authorization: { model: "system_actor_schedule_scoped_delegation_v1", ...actors, required_permissions: ["application_evaluations:execute", "workflow_runs:execute"], revalidation_policy: "every_occurrence", api_key_ownership_policy: "delegated_user_current_owner", revocation_policy: "fail_closed_immediate" },
    created_at: time, created_by_actor_ref: "subject_demo_user" };
  const occurrence = { schema_version: "application_evaluation_schedule_occurrence.v1", ...scope, ...actors, ...refs,
    schedule_id: id, schedule_version: 1, schedule_digest: digest, scheduled_for_utc: time, record_version: 4,
    state: "skipped", client_campaign_key: "scheduled_campaign_aaaaaaaaaaaaaaaaaaaaaaaa", campaign_id: null,
    claimed_at: time, failure_code: "application_evaluation_schedule_missed_window", created_at: time, updated_at: time, completed_at: time };
  let exists = false, failure = "", held: { arrived: () => void; gate: Promise<void> } | null = null;
  const mutations: object[] = [];
  const base = `/v1/user-workspace/applications/${applicationId}/evaluation-schedules`;
  await page.route(`http://127.0.0.1:17000${base}**`, async route => {
    const request = route.request(), url = new URL(request.url());
    const envelope = { ...scope, ...refs, failure_code: failure || null, failure_summary: failure ? "SERVER PRIVATE PROSE MUST NOT RENDER" : "" };
    if (url.pathname === base && request.method() === "GET") {
      return route.fulfill({ json: { ...envelope, schedules: exists && url.searchParams.get("lifecycle_state") === record.lifecycle_state ? [record] : [], next_cursor: "", has_more: false } });
    }
    if (request.method() === "POST") {
      const body = request.postDataJSON(); mutations.push(body);
      if (!failure) {
        if (url.pathname === base || url.pathname.endsWith("/revisions")) {
          if (exists) { record.record_version++; version.previous_schedule_version = version.schedule_version; version.schedule_version++; record.latest_schedule_version++; }
          exists = true; version.schedule = body.schedule;
          record.quota_api_key_id = version.quota_api_key_id = body.quota_api_key_id;
        } else {
          record.record_version++;
          const action = url.pathname.split("/").at(-1);
          record.lifecycle_state = action === "pause" ? "paused" : action === "archive" ? "archived" : "active";
          record.next_due_at = record.lifecycle_state === "active" ? time : null;
        }
      }
      if (held) { const current = held; held = null; current.arrived(); await current.gate; }
    }
    const isOccurrence = url.pathname.includes("/occurrences/");
    const isVersion = url.pathname.includes("/versions/");
    return route.fulfill({ status: failure ? 409 : 200, json: { ...envelope, schedule: failure || isOccurrence || isVersion ? null : record,
      version: failure || isOccurrence ? null : version, occurrence: failure || !isOccurrence ? null : occurrence,
      current_record_version: isOccurrence ? occurrence.record_version : record.record_version, current_state: isOccurrence ? occurrence.state : record.lifecycle_state } });
  });
  return {
    id, time, mutations,
    fail(code: string) { failure = code; },
    conflict() { record.record_version++; failure = "application_evaluation_schedule_version_conflict"; },
    hold() {
      let release!: () => void, arrive!: () => void;
      const arrived = new Promise<void>(resolve => { arrive = resolve; });
      const gate = new Promise<void>(resolve => { release = resolve; });
      held = { arrived: arrive, gate }; return { arrived, release };
    },
  };
}

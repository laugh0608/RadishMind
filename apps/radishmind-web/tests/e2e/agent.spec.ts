import { randomUUID } from "node:crypto";
import type { Page, TestInfo } from "@playwright/test";
import { test, expect, holdDraftResponse, type Application } from "./workflow-fixtures";
import { uiText } from "./ui-language";
import { promotionRequest } from "./workflow-rag-promotion-fixtures";
import { agentLanguage, agentLayout, agentSubmit, navigateAgent } from "./agent-ui";

test.use({ applicationKind: "agent" });
const profilePath = "/v1/user-workspace/agent-copilot-profiles";
const candidatePath = "/v1/user-workspace/application-publish-candidates";
const sessionPath = "/v1/user-workspace/application-sessions";
const profile = (page: Page) => page.locator("#agent-copilot-profile-workspace");
const runtime = (page: Page) => page.locator("#agent-copilot-runtime-assignment");
const session = (page: Page) => page.locator("#agent-copilot-invocation");
const button = (panel: ReturnType<typeof profile>, name: string) => panel.getByRole("button", { name, exact: true });
const field = (panel: ReturnType<typeof profile>, name: string) => panel.getByRole("textbox", { name, exact: true });

async function prepareAgent(page: Page, application: Application, testInfo: TestInfo) {
  const language = agentLanguage(page, testInfo), copy = language.copy;
  const scope = { workspace_id: "workspace_demo", application_id: application.id };
  const app = await promotionRequest(page, application.id, `/v1/user-workspace/applications/${application.id}?workspace_id=workspace_demo`, ["applications:read"]);
  const draftId = `agent-e2e-${randomUUID()}`;
  // The shared configuration flow is already accepted; seed only this prerequisite
  // through its public API. All Agent owners are exercised through the browser.
  const draft = await promotionRequest(page, application.id, "/v1/user-workspace/application-drafts", ["application_drafts:write"], {
    expected_draft_version: 0, draft: { ...scope, draft_id: draftId, base_application_updated_at: app.record.updated_at,
      schema_version: "application_configuration_draft.v1", display_name: application.name, description: "Synthetic Agent configuration",
      application_kind: "agent", default_protocol: "chat_completions", default_model: "mock", allowed_protocols: ["chat_completions"] },
  });
  expect(draft.draft.validation_summary.state).toBe("valid");
  await navigateAgent(page, "application-configuration-draft");
  await navigateAgent(page, "agent-copilot-profile-workspace");
  await field(profile(page), copy().name).fill("Synthetic Agent 原始名称");
  await field(profile(page), copy().defaultLocale).fill("en-US");
  const saved = await agentSubmit(page, button(profile(page), copy().save), profilePath);
  const profileId = saved.body.draft.profile_id;
  const version = await agentSubmit(page, button(profile(page), copy().version), `${profilePath}/${profileId}/versions`);
  expect(version.body.version.default_locale).toBe("en-US");
  await expect(profile(page).getByRole("combobox", { name: copy().immutableVersion, exact: true }).locator("option[value=\"1\"]")).toHaveCount(1);
  await language.switch();
  await expect(field(profile(page), copy().defaultLocale)).toHaveValue("en-US");
  await expect(field(profile(page), copy().name)).toHaveValue("Synthetic Agent 原始名称");
  await button(profile(page), copy().loadDrafts).click();
  await profile(page).getByRole("combobox", { name: copy().validDraft, exact: true }).selectOption(draftId);
  await profile(page).getByRole("combobox", { name: copy().immutableVersion, exact: true }).selectOption("1");
  await agentLayout(page, profile(page), testInfo, "agent-profile");
  const bound = await agentSubmit(page, button(profile(page), copy().bind), `/v1/user-workspace/application-configuration-drafts/${draftId}/agent-copilot-profile-binding`);
  expect(bound.body.draft.agent_copilot_profile_ref.profile_id).toBe(profileId);
  await navigateAgent(page, "application-publish-review");
  const review = page.locator("#application-publish-review");
  await button(review, uiText(page, "Load saved drafts")).click();
  await review.getByRole("combobox", { name: uiText(page, "Saved valid draft"), exact: true }).selectOption(draftId);
  const created = await agentSubmit(page, button(review, uiText(page, "Create immutable candidate")), candidatePath);
  const candidateId = created.body.candidate.candidate_id;
  await button(review, uiText(page, "Read exact source")).click();
  await expect(review.locator(".prompt-template-source-review")).toContainText("en-US");
  await field(review, uiText(page, "Review reason")).fill("Reviewed exact synthetic Agent profile and configuration.");
  await agentSubmit(page, button(review, uiText(page, "Record review decision")), `${candidatePath}/${candidateId}/reviews`);
  await navigateAgent(page, "agent-copilot-runtime-assignment");
  await button(runtime(page), copy().load).click();
  await expect(runtime(page)).toContainText(candidateId);
  const runtimePath = `/v1/user-workspace/applications/${application.id}/agent-copilot-runtime-assignment`;
  const active = await agentSubmit(page, button(runtime(page), copy().activate), `${runtimePath}/decisions`);
  expect(active.body.assignment.agent_copilot_profile_ref.profile_id).toBe(profileId);
  await language.switch(); await expect(runtime(page)).toContainText(candidateId);
  await agentLayout(page, runtime(page), testInfo, "agent-assignment");
  await navigateAgent(page, "agent-copilot-invocation");
  const createdSession = await agentSubmit(page, button(session(page), copy().create), sessionPath);
  expect(createdSession.body.session.authority.agent_copilot.agent_copilot_profile_ref.profile_id).toBe(profileId);
  return { language, copy, sessionId: createdSession.body.session.session_id as string, runtimePath, active, candidateId };
}

test("Agent profile validation, pending save and real CAS conflict preserve business locales", async ({ page, application }, testInfo) => {
  const language = agentLanguage(page, testInfo), copy = language.copy;
  await navigateAgent(page, "application-configuration-draft");
  await navigateAgent(page, "agent-copilot-profile-workspace");
  await field(profile(page), copy().name).fill("x");
  await expect(profile(page)).toContainText(copy().invalidName);
  await field(profile(page), copy().defaultLocale).fill("en-US");
  await button(profile(page), copy().validate).click();
  await language.switch();
  await expect(profile(page)).toContainText(copy().invalidName);
  await expect(profile(page)).not.toContainText(copy().invalidName.startsWith("Profile names") ? "Profile 名称必须为" : "Profile names must contain");
  await expect(field(profile(page), copy().name)).toHaveValue("x");
  await expect(button(profile(page), copy().save)).toBeDisabled();
  await field(profile(page), copy().name).fill("Untranslated 原始 Profile");
  const held = await holdDraftResponse(page, profilePath, "POST");
  let saved: Awaited<ReturnType<typeof agentSubmit>>;
  try {
    const pending = agentSubmit(page, button(profile(page), copy().save), profilePath);
    await held.arrived; await language.switch();
    await expect(field(profile(page), copy().defaultLocale)).toHaveValue("en-US");
    await held.deliver(); saved = await pending;
  } finally { await held.dispose(); }
  expect(saved!.body.draft.default_locale).toBe("en-US");
  const competing = await page.request.post(saved!.request.url(), {
    headers: await saved!.request.allHeaders(), data: { ...saved!.request.postDataJSON(), expected_draft_version: 1 },
  });
  expect((await competing.json()).draft.draft_version).toBe(2);
  await field(profile(page), copy().name).fill("Keep this unsaved edit 原始输入");
  const conflict = await agentSubmit(page, button(profile(page), copy().save), profilePath, true);
  expect(conflict.body.failure_code).toBe("agent_copilot_profile_version_conflict");
  await language.switch();
  await expect(profile(page)).toContainText(copy().conflict);
  await expect(field(profile(page), copy().name)).toHaveValue("Keep this unsaved edit 原始输入");
  await expect(field(profile(page), copy().defaultLocale)).toHaveValue("en-US");
  await agentLayout(page, profile(page), testInfo, "agent-conflict");
  await button(profile(page), copy().refresh).first().click();
  await profile(page).locator(".prompt-template-summary").filter({ hasText: saved!.body.draft.profile_id }).click();
  await expect(field(profile(page), copy().name)).toHaveValue("Untranslated 原始 Profile");
  expect(application.kind).toBe("agent");
});

test("Agent exact authority, single turn and replay remain stable across interface languages", async ({ page, application }, testInfo) => {
  test.setTimeout(120_000);
  const { language, copy, sessionId } = await prepareAgent(page, application, testInfo);
  await field(session(page), copy().locale).fill("en-US");
  await field(session(page), copy().context).fill("{");
  await button(session(page), copy().execute).click(); await expect(session(page)).toContainText(copy().invalidContext);
  await language.switch(); await expect(session(page)).toContainText(copy().invalidContext);
  const context = '{"selected_unit_ids":["unit-101"],"diagnostics":[{"code":"not_converged"}]}';
  await field(session(page), copy().context).fill(context);
  const key = `agent-e2e-${randomUUID()}`;
  await field(session(page), copy().key).fill(key);
  const artifacts = session(page).locator(".application-result-artifact-owner");
  await artifacts.getByRole("checkbox").check();
  const path = `${sessionPath}/${sessionId}/turns`;
  const held = await holdDraftResponse(page, path, "POST");
  let turn: Awaited<ReturnType<typeof agentSubmit>>;
  try {
    const pending = agentSubmit(page, button(session(page), copy().execute), path);
    await held.arrived; await language.switch();
    await expect(field(session(page), copy().locale)).toHaveValue("en-US");
    await expect(field(session(page), copy().context)).toHaveValue(context);
    await expect(field(session(page), copy().key)).toHaveValue(key);
    await held.deliver(); turn = await pending;
  } finally { await held.dispose(); }
  expect(turn!.request.postDataJSON()).toMatchObject({ locale: "en-US", context: JSON.parse(context), client_turn_key: key, save_result: true });
  expect(turn!.body.turn.status).toBe("succeeded");
  expect(turn!.body.action_safety.status).toBe("recorded");
  expect(turn!.body.result_artifact.artifact_id).toMatch(/^appres_/);
  const summary = turn!.body.agent_response.summary;
  await expect(session(page)).toContainText(summary);
  await language.switch(); await expect(session(page)).toContainText(summary);
  await expect(artifacts.getByRole("checkbox")).not.toBeChecked();
  await agentLayout(page, session(page), testInfo, "agent-session");
  const replay = await agentSubmit(page, button(session(page), copy().execute), path);
  expect(replay.body.idempotent_replay).toBe(true); expect(replay.body.agent_response ?? null).toBeNull();
  await expect(session(page)).toContainText(copy().replay);
  await language.switch(); await expect(session(page)).toContainText(copy().replay);
  await expect(field(session(page), copy().locale)).toHaveValue("en-US");
  await page.reload();
  await page.locator(".application-catalog-list button").filter({ hasText: application.id }).click();
  await navigateAgent(page, "agent-copilot-invocation");
  await expect(field(session(page), copy().key)).not.toHaveValue(key);
  await expect(session(page).locator(".prompt-template-review")).not.toContainText(summary);
});

test("Agent assignment conflict and revoked authority keep exact failure codes across languages", async ({ page, application }, testInfo) => {
  test.setTimeout(120_000);
  const { language, copy, active, runtimePath, candidateId } = await prepareAgent(page, application, testInfo);
  await navigateAgent(page, "agent-copilot-runtime-assignment");
  await button(runtime(page), copy().load).click();
  await expect(button(runtime(page), copy().revoke)).toBeEnabled();
  const competing = await page.request.post(active.request.url(), {
    headers: await active.request.allHeaders(), data: { workspace_id: "workspace_demo", expected_assignment_version: 1, action: "revoke" },
  });
  expect((await competing.json()).assignment.state).toBe("revoked");
  const conflict = await agentSubmit(page, button(runtime(page), copy().revoke), `${runtimePath}/decisions`, true);
  expect(conflict.body.failure_code).toBe("agent_copilot_runtime_assignment_version_conflict");
  await language.switch(); await expect(runtime(page)).toContainText("agent_copilot_runtime_assignment_version_conflict");
  await expect(runtime(page)).toContainText(candidateId);
  await button(runtime(page), copy().load).click();
  await navigateAgent(page, "agent-copilot-invocation");
  const blocked = await agentSubmit(page, button(session(page), copy().create), sessionPath, true);
  expect(blocked.body.failure_code).toMatch(/revoked|authority|ineligible/);
  await language.switch(); await expect(session(page)).toContainText(blocked.body.failure_code);
  await expect(button(session(page), copy().execute)).toBeDisabled();
});
